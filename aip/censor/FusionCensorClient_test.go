/*
Copyright 2021 baidu

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package censor

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// capturedRequest 记录 stub 服务端收到的请求，供断言使用。
type capturedRequest struct {
	method      string
	path        string
	contentType string
	authIsBce   bool
	bceDate     string
	form        url.Values
}

// newStubServer 起一个回放固定报文的 stub 服务端，并把收到的请求写进 captured。
func newStubServer(t *testing.T, status int, body string, captured *capturedRequest) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		raw, err := io.ReadAll(req.Body)
		if err != nil {
			t.Errorf("read request body failed: %v", err)
		}
		form, err := url.ParseQuery(string(raw))
		if err != nil {
			t.Errorf("parse request form failed: %v", err)
		}
		*captured = capturedRequest{
			method:      req.Method,
			path:        req.URL.Path,
			contentType: req.Header.Get("Content-Type"),
			authIsBce:   strings.HasPrefix(req.Header.Get("Authorization"), "bce-auth-v1/"),
			bceDate:     req.Header.Get("x-bce-date"),
			form:        form,
		}
		writer.Header().Set("Content-Type", "application/json")
		writer.WriteHeader(status)
		_, _ = writer.Write([]byte(body))
	}))
}

func newStubClient(t *testing.T, server *httptest.Server) *FusionCensorClient {
	t.Helper()
	client, err := NewFusionClientWithEndpoint("ak", "sk", server.URL)
	if err != nil {
		t.Fatalf("create client failed: %v", err)
	}
	return client
}

// TestNewFusionClientUsesNorthChina 默认接入华北地域；ak/sk 缺失必须报错。
func TestNewFusionClientUsesNorthChina(t *testing.T) {
	client, err := NewFusionClient("ak", "sk")
	if err != nil {
		t.Fatalf("unexpected error %v", err)
	}
	if client.Endpoint() != __fusionNorthChinaEndpoint {
		t.Errorf("endpoint = %q, want %q", client.Endpoint(), __fusionNorthChinaEndpoint)
	}

	if _, err := NewFusionClient("", "sk"); err == nil {
		t.Error("empty ak should fail")
	}
	if _, err := NewFusionClient("ak", ""); err == nil {
		t.Error("empty sk should fail")
	}
	if _, err := NewFusionClientWithEndpoint("ak", "sk", ""); err == nil {
		t.Error("empty endpoint should fail")
	}
}

// TestSubmitTextRequestWire 校验实际发出的报文：路径、方法、Content-Type、
// 签名头、以及 form 字段名与服务端 SubmitTextRequest 对齐。
func TestSubmitTextRequestWire(t *testing.T) {
	var captured capturedRequest
	server := newStubServer(t, http.StatusOK, `{"taskId":"tid-text"}`, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	response, err := client.SubmitText(&SubmitTextRequest{
		Text:        "加微信免费领取红包",
		StrategyID:  2017191,
		UserID:      "qa_demo",
		ExtStr:      `{"extStr1":"a"}`,
		CallbackURL: "https://example.com/cb",
		UserIP:      "10.0.0.1",
		PhoneSha256: "abc",
		DeviceID:    "dev-1",
	})
	if err != nil {
		t.Fatalf("SubmitText failed: %v", err)
	}

	if response.TaskID != "tid-text" {
		t.Errorf("taskId = %q, want tid-text", response.TaskID)
	}
	if !response.IsSuccess() {
		t.Errorf("expect success, got error_code=%d", response.ErrorCode)
	}
	if captured.method != http.MethodPost {
		t.Errorf("method = %s, want POST", captured.method)
	}
	if captured.path != "/api/v1/fusion/fusion_censor/v1/text/submit" {
		t.Errorf("path = %s, want /api/v1/fusion/fusion_censor/v1/text/submit", captured.path)
	}
	if captured.contentType != "application/x-www-form-urlencoded" {
		t.Errorf("Content-Type = %q, want application/x-www-form-urlencoded", captured.contentType)
	}
	if !captured.authIsBce {
		t.Error("Authorization header should start with bce-auth-v1/")
	}
	if captured.bceDate == "" {
		t.Error("x-bce-date header should be set")
	}

	want := map[string]string{
		"text":        "加微信免费领取红包",
		"strategyId":  "2017191",
		"userId":      "qa_demo",
		"extStr":      `{"extStr1":"a"}`,
		"callbackUrl": "https://example.com/cb",
		"userIp":      "10.0.0.1",
		"phoneSha256": "abc",
		"deviceId":    "dev-1",
	}
	for key, wantValue := range want {
		if got := captured.form.Get(key); got != wantValue {
			t.Errorf("form[%s] = %q, want %q", key, got, wantValue)
		}
	}
}

// TestSubmitOmitsEmptyOptionalFields 可选字段为空时不应出现在 form 里，
// 避免服务端把空串当成有效值（如 callbackUrl="" 触发格式校验）。
func TestSubmitOmitsEmptyOptionalFields(t *testing.T) {
	var captured capturedRequest
	server := newStubServer(t, http.StatusOK, `{"taskId":"tid"}`, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	if _, err := client.SubmitText(&SubmitTextRequest{Text: "hi"}); err != nil {
		t.Fatalf("SubmitText failed: %v", err)
	}

	for _, key := range []string{"userId", "extStr", "callbackUrl", "userIp", "phoneSha256", "deviceId"} {
		if _, ok := captured.form[key]; ok {
			t.Errorf("form should not contain empty %s", key)
		}
	}
}

// TestSubmitDefaultsStrategyID 不传 strategyId 时必须兜底成 1：
// 落到服务端会变成 0，小模型直接返回 strategy not exist（282910）。
func TestSubmitDefaultsStrategyID(t *testing.T) {
	var captured capturedRequest
	server := newStubServer(t, http.StatusOK, `{"taskId":"tid"}`, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	if _, err := client.SubmitText(&SubmitTextRequest{Text: "hi"}); err != nil {
		t.Fatalf("SubmitText failed: %v", err)
	}
	if got := captured.form.Get("strategyId"); got != "1" {
		t.Errorf("strategyId = %q, want 1", got)
	}
}

// TestSubmitImageWire 图像提交用 imgUrl 字段，路径为 image/submit。
func TestSubmitImageWire(t *testing.T) {
	var captured capturedRequest
	server := newStubServer(t, http.StatusOK, `{"taskId":"tid-img"}`, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	if _, err := client.SubmitImage(&SubmitImageRequest{
		ImgURL:     "https://example.com/a.jpg",
		StrategyID: 2017191,
	}); err != nil {
		t.Fatalf("SubmitImage failed: %v", err)
	}

	if captured.path != "/api/v1/fusion/fusion_censor/v1/image/submit" {
		t.Errorf("path = %s", captured.path)
	}
	if got := captured.form.Get("imgUrl"); got != "https://example.com/a.jpg" {
		t.Errorf("imgUrl = %q", got)
	}
}

// TestSubmitVideoWire 视频提交用 url 字段；DetectType 为 nil 时不发该参数，
// 由服务端按默认值 1（仅视频帧）处理。
func TestSubmitVideoWire(t *testing.T) {
	var captured capturedRequest
	server := newStubServer(t, http.StatusOK, `{"taskId":"tid-video"}`, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	if _, err := client.SubmitVideo(&SubmitVideoRequest{
		URL:   "https://example.com/a.mp4",
		ExtID: "ext-1",
	}); err != nil {
		t.Fatalf("SubmitVideo failed: %v", err)
	}
	if captured.path != "/api/v1/fusion/fusion_censor/v1/video/submit" {
		t.Errorf("path = %s", captured.path)
	}
	if got := captured.form.Get("url"); got != "https://example.com/a.mp4" {
		t.Errorf("url = %q", got)
	}
	if got := captured.form.Get("extId"); got != "ext-1" {
		t.Errorf("extId = %q", got)
	}
	if _, ok := captured.form["detectType"]; ok {
		t.Error("detectType should be omitted when DetectType is nil")
	}

	detectType := VideoDetectTypeFrameAndAudio
	if _, err := client.SubmitVideo(&SubmitVideoRequest{
		URL:        "https://example.com/a.mp4",
		DetectType: &detectType,
	}); err != nil {
		t.Fatalf("SubmitVideo failed: %v", err)
	}
	if got := captured.form.Get("detectType"); got != "0" {
		t.Errorf("detectType = %q, want 0 (frame+audio)", got)
	}
}

// TestSubmitRejectsMissingRequiredField 必填字段缺失应在本地拦下，不发请求。
func TestSubmitRejectsMissingRequiredField(t *testing.T) {
	client, err := NewFusionClient("ak", "sk")
	if err != nil {
		t.Fatalf("create client failed: %v", err)
	}
	if _, err := client.SubmitText(&SubmitTextRequest{}); err == nil {
		t.Error("empty text should fail locally")
	}
	if _, err := client.SubmitImage(&SubmitImageRequest{}); err == nil {
		t.Error("empty imgUrl should fail locally")
	}
	if _, err := client.SubmitVideo(&SubmitVideoRequest{}); err == nil {
		t.Error("empty video url should fail locally")
	}
	if _, err := client.PullTextResult(""); err == nil {
		t.Error("empty taskId should fail locally")
	}
}

// TestPullResultPaths 三个模态各自独立的 result 路径，taskId 走 form。
func TestPullResultPaths(t *testing.T) {
	cases := []struct {
		name string
		pull func(client *FusionCensorClient, taskID string) (*FusionResponse, error)
		path string
	}{
		{"text", (*FusionCensorClient).PullTextResult, "/api/v1/fusion/fusion_censor/v1/text/result"},
		{"image", (*FusionCensorClient).PullImageResult, "/api/v1/fusion/fusion_censor/v1/image/result"},
		{"video", (*FusionCensorClient).PullVideoResult, "/api/v1/fusion/fusion_censor/v1/video/result"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var captured capturedRequest
			server := newStubServer(t, http.StatusOK,
				`{"taskId":"tid","status":"PROCESSING","createTime":"2026-08-28T02:02:59.000+00:00"}`, &captured)
			defer server.Close()

			client := newStubClient(t, server)
			response, err := testCase.pull(client, "tid")
			if err != nil {
				t.Fatalf("pull failed: %v", err)
			}
			if captured.path != testCase.path {
				t.Errorf("path = %s, want %s", captured.path, testCase.path)
			}
			if got := captured.form.Get("taskId"); got != "tid" {
				t.Errorf("taskId = %q, want tid", got)
			}
			if response.Status != FusionStatusProcessing {
				t.Errorf("status = %q, want PROCESSING", response.Status)
			}
			// PROCESSING 不是失败，且不是终态
			if !response.IsSuccess() {
				t.Error("PROCESSING should not be treated as failure")
			}
			if response.IsTerminal() {
				t.Error("PROCESSING should not be terminal")
			}
		})
	}
}

// TestGatewayErrorSurfaces API Gateway 在到达审核服务前拦下请求时，返回的是
// BCE 通用错误格式（code / message / requestId），与业务错误的 error_code 形状不同。
// 只看 error_code 会把网关拒绝误判成成功 —— 实测中签名无效正是这个形状。
func TestGatewayErrorSurfaces(t *testing.T) {
	var captured capturedRequest
	server := newStubServer(t, http.StatusForbidden,
		`{"message":"IamSignatureInvalid, cause: Could not find credential.",`+
			`"code":"IamSignatureInvalid","requestId":"req-123"}`, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	response, err := client.SubmitText(&SubmitTextRequest{Text: "hi"})
	if err != nil {
		t.Fatalf("SubmitText should not return transport error: %v", err)
	}
	if response.IsSuccess() {
		t.Error("gateway rejection must not be reported as success")
	}
	if response.GatewayCode != "IamSignatureInvalid" {
		t.Errorf("GatewayCode = %q, want IamSignatureInvalid", response.GatewayCode)
	}
	if response.RequestID != "req-123" {
		t.Errorf("RequestID = %q, want req-123", response.RequestID)
	}
	if !strings.Contains(response.ErrorDescription(), "IamSignatureInvalid") {
		t.Errorf("ErrorDescription = %q, should mention the gateway code", response.ErrorDescription())
	}
	// 业务错误码为 0，若只看它就会误判成成功
	if response.ErrorCode != 0 {
		t.Errorf("ErrorCode = %d, want 0 for gateway-shaped error", response.ErrorCode)
	}
}

// TestLogIDParsed log_id 是报障时反查服务端日志的唯一锚点，必须解析出来。
func TestLogIDParsed(t *testing.T) {
	var captured capturedRequest
	server := newStubServer(t, http.StatusOK,
		`{"taskId":"tid-abc","log_id":"log-xyz"}`, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	response, err := client.SubmitText(&SubmitTextRequest{Text: "hi"})
	if err != nil {
		t.Fatalf("SubmitText failed: %v", err)
	}
	if response.LogID != "log-xyz" {
		t.Errorf("LogID = %q, want log-xyz", response.LogID)
	}
}

// TestErrorDescriptionCarriesLogID 业务错误的可读描述里要带上 log_id。
func TestErrorDescriptionCarriesLogID(t *testing.T) {
	var captured capturedRequest
	server := newStubServer(t, http.StatusBadRequest,
		`{"log_id":"lg-1","error_code":282909,"error_msg":"text length 2100 exceeds limit 2000"}`,
		&captured)
	defer server.Close()

	client := newStubClient(t, server)
	response, err := client.SubmitText(&SubmitTextRequest{Text: "hi"})
	if err != nil {
		t.Fatalf("SubmitText failed: %v", err)
	}
	description := response.ErrorDescription()
	for _, want := range []string{"282909", "exceeds limit", "lg-1"} {
		if !strings.Contains(description, want) {
			t.Errorf("ErrorDescription = %q, should contain %q", description, want)
		}
	}
}

// TestErrorResponseOnNon2xx 服务端在 4xx/5xx 上仍返回完整错误报文，
// 客户端必须解析出 error_code 而不是把 HTTP 状态当成传输失败。
func TestErrorResponseOnNon2xx(t *testing.T) {
	var captured capturedRequest
	server := newStubServer(t, http.StatusForbidden,
		`{"error_code":6,"error_msg":"No permission to access data"}`, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	response, err := client.SubmitText(&SubmitTextRequest{Text: "hi"})
	if err != nil {
		t.Fatalf("SubmitText should not return transport error: %v", err)
	}
	if response.IsSuccess() {
		t.Error("403 with error_code=6 should not be success")
	}
	if response.ErrorCode != 6 {
		t.Errorf("error_code = %d, want 6", response.ErrorCode)
	}
	if response.ErrorMsg != "No permission to access data" {
		t.Errorf("error_msg = %q", response.ErrorMsg)
	}
}

// TestNonJSONResponseReturnsError 报文不是 JSON（如网关返回 HTML 跳转页）时必须报错，
// 并把原始内容带进错误信息，否则排查时只看到一个空结构体。
func TestNonJSONResponseReturnsError(t *testing.T) {
	var captured capturedRequest
	server := newStubServer(t, http.StatusOK, `<html>302 Found</html>`, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	_, err := client.SubmitText(&SubmitTextRequest{Text: "hi"})
	if err == nil {
		t.Fatal("non-json response should return error")
	}
	if !strings.Contains(err.Error(), "302 Found") {
		t.Errorf("error should carry raw body, got %v", err)
	}
}

// TestUnmarshalAuditDetails 文本 / 图像明细：已建模字段正常解析，
// 未建模字段（probability、location、hits 等）必须保留在 Extras 里不丢失。
func TestUnmarshalAuditDetails(t *testing.T) {
	body := `{
		"taskId":"tid","status":"FINISHED",
		"createTime":"2026-08-28T05:47:48.000+00:00",
		"finishTime":"2026-08-28T05:48:24.000+00:00",
		"conclusion":"不合规","conclusionType":2,
		"data":[
			{"type":4,"subType":0,"conclusionType":2,"msg":"存在水印不合规",
			 "conclusion":"不合规","probability":0.9914713,
			 "location":[{"score":0.99147128,"top":515,"left":822,"width":136.0,"height":24.0}]},
			{"type":47,"subType":470101,"conclusionType":2,"msg":"存在诱导点击按钮不合规",
			 "agentType":"广告法","agentSubType":"虚假诱导点击按钮"}
		]}`
	var captured capturedRequest
	server := newStubServer(t, http.StatusOK, body, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	response, err := client.PullImageResult("tid")
	if err != nil {
		t.Fatalf("PullImageResult failed: %v", err)
	}
	if !response.IsTerminal() {
		t.Error("FINISHED should be terminal")
	}
	if response.ConclusionType != ConclusionTypeNonCompliant {
		t.Errorf("conclusionType = %d, want 2", response.ConclusionType)
	}

	details, err := response.UnmarshalAuditDetails()
	if err != nil {
		t.Fatalf("UnmarshalAuditDetails failed: %v", err)
	}
	if len(details) != 2 {
		t.Fatalf("details len = %d, want 2", len(details))
	}
	if details[0].Type != 4 || details[0].Msg != "存在水印不合规" {
		t.Errorf("details[0] = %+v", details[0])
	}
	if _, ok := details[0].Extras["probability"]; !ok {
		t.Error("probability should be kept in Extras")
	}
	if _, ok := details[0].Extras["location"]; !ok {
		t.Error("location should be kept in Extras")
	}
	if _, ok := details[0].Extras["msg"]; ok {
		t.Error("modeled field msg should not appear in Extras")
	}
	if details[1].AgentType != "广告法" || details[1].AgentSubType != "虚假诱导点击按钮" {
		t.Errorf("details[1] agent labels = %q / %q", details[1].AgentType, details[1].AgentSubType)
	}
}

// TestUnmarshalVideoData 长视频 data 是对象而非数组，frames / audios 键恒存在。
// 帧的时间戳单位是秒、音频是毫秒（服务端口径不一致，SDK 不换算）；
// 音频的违规明细挂在 rawText[].data 而非 audios[] 本层。
func TestUnmarshalVideoData(t *testing.T) {
	body := `{
		"taskId":"tid","status":"FINISHED","conclusion":"不合规","conclusionType":2,
		"data":{"taskDuration":42,
			"frames":[{"frameTimeStamp":10,"frameUrl":"https://bos/f.jpg",
				"frameThumbnailUrl":"https://bos/f.jpg",
				"data":[{"type":4,"subType":0,"conclusionType":2,"msg":"水印"}]}],
			"audios":[{"startTime":0,"endTime":5000,"audioUrl":"https://bos/a.pcm",
				"audioAuditResult":[{"type":33,"subType":330105,"conclusionType":2,
					"conclusion":"不合规","msg":"存在娇喘不合规"}],
				"rawText":[{"startTime":0,"endTime":5000,"text":"加微信领红包",
					"conclusionType":2,"conclusion":"不合规",
					"data":[{"type":12,"subType":4,"msg":"存在广告不合规"},
						{"agentType":"广告法"}]}]}]}}`
	var captured capturedRequest
	server := newStubServer(t, http.StatusOK, body, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	response, err := client.PullVideoResult("tid")
	if err != nil {
		t.Fatalf("PullVideoResult failed: %v", err)
	}

	data, err := response.UnmarshalVideoData()
	if err != nil {
		t.Fatalf("UnmarshalVideoData failed: %v", err)
	}
	if data.TaskDuration != 42 {
		t.Errorf("taskDuration = %d, want 42", data.TaskDuration)
	}
	if len(data.Frames) != 1 || data.Frames[0].FrameTimeStamp != 10 {
		t.Fatalf("frames = %+v", data.Frames)
	}
	if len(data.Frames[0].Data) != 1 || data.Frames[0].Data[0].Type != 4 {
		t.Errorf("frame details = %+v", data.Frames[0].Data)
	}
	if len(data.Audios) != 1 {
		t.Fatalf("audios = %+v", data.Audios)
	}
	audio := data.Audios[0]
	if audio.EndTime != 5000 || audio.AudioURL != "https://bos/a.pcm" {
		t.Errorf("audio = %+v", audio)
	}
	// audioAuditResult 是数组，且带 conclusionType / conclusion / msg
	if len(audio.AudioAuditResult) != 1 {
		t.Fatalf("audioAuditResult = %+v", audio.AudioAuditResult)
	}
	feature := audio.AudioAuditResult[0]
	if feature.SubType != 330105 || feature.Msg != "存在娇喘不合规" ||
		feature.ConclusionType != ConclusionTypeNonCompliant {
		t.Errorf("audioAuditResult[0] = %+v", feature)
	}
	// rawText 是对象数组，违规明细挂在它下面
	if len(audio.RawText) != 1 {
		t.Fatalf("rawText = %+v", audio.RawText)
	}
	rawText := audio.RawText[0]
	if rawText.Text != "加微信领红包" || rawText.ConclusionType != ConclusionTypeNonCompliant {
		t.Errorf("rawText[0] = %+v", rawText)
	}
	if len(rawText.Data) != 2 || rawText.Data[0].Type != 12 || rawText.Data[1].AgentType != "广告法" {
		t.Errorf("rawText[0].data = %+v", rawText.Data)
	}
}

// TestUnmarshalDataAbsent 合规且未命中白名单时不返回 data，解析应返回空值而非报错。
func TestUnmarshalDataAbsent(t *testing.T) {
	body := `{"taskId":"tid","status":"FINISHED","conclusion":"合规","conclusionType":1}`
	var captured capturedRequest
	server := newStubServer(t, http.StatusOK, body, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	response, err := client.PullTextResult("tid")
	if err != nil {
		t.Fatalf("PullTextResult failed: %v", err)
	}
	if response.ConclusionType != ConclusionTypeCompliant {
		t.Errorf("conclusionType = %d, want 1", response.ConclusionType)
	}

	details, err := response.UnmarshalAuditDetails()
	if err != nil {
		t.Fatalf("UnmarshalAuditDetails failed: %v", err)
	}
	if details != nil {
		t.Errorf("details should be nil, got %+v", details)
	}
	videoData, err := response.UnmarshalVideoData()
	if err != nil {
		t.Fatalf("UnmarshalVideoData failed: %v", err)
	}
	if videoData != nil {
		t.Errorf("video data should be nil, got %+v", videoData)
	}
}

// TestRawKeepsOriginalBody Raw 必须留下原始报文，便于排查未建模字段。
func TestRawKeepsOriginalBody(t *testing.T) {
	body := `{"taskId":"tid","status":"PROVISIONING","conclusionType":0}`
	var captured capturedRequest
	server := newStubServer(t, http.StatusOK, body, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	response, err := client.PullTextResult("tid")
	if err != nil {
		t.Fatalf("PullTextResult failed: %v", err)
	}
	if response.Raw != body {
		t.Errorf("Raw = %q, want %q", response.Raw, body)
	}
	// PROVISIONING 阶段服务端会带值域外的 conclusionType=0，表示尚无结论
	if response.ConclusionType != 0 {
		t.Errorf("conclusionType = %d, want 0", response.ConclusionType)
	}
	if response.IsTerminal() {
		t.Error("PROVISIONING should not be terminal")
	}
}

// TestSessionTokenAddedToRequest 设置 STS token 后请求头必须带上。
func TestSessionTokenAddedToRequest(t *testing.T) {
	var gotToken string
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		gotToken = req.Header.Get("x-bce-security-token")
		_, _ = writer.Write([]byte(`{"taskId":"tid"}`))
	}))
	defer server.Close()

	client, err := NewFusionClientWithEndpoint("ak", "sk", server.URL)
	if err != nil {
		t.Fatalf("create client failed: %v", err)
	}
	client.SetSessionToken("sts-token")
	if _, err := client.SubmitText(&SubmitTextRequest{Text: "hi"}); err != nil {
		t.Fatalf("SubmitText failed: %v", err)
	}
	if gotToken != "sts-token" {
		t.Errorf("x-bce-security-token = %q, want sts-token", gotToken)
	}
}

// TestForbiddenIdentityParamsNotSent SDK 不得发送身份类参数 ——
// 身份只由 IAM 签名承载，服务端收到这些参数会忽略并记安全日志。
func TestForbiddenIdentityParamsNotSent(t *testing.T) {
	var captured capturedRequest
	server := newStubServer(t, http.StatusOK, `{"taskId":"tid"}`, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	if _, err := client.SubmitText(&SubmitTextRequest{Text: "hi", UserID: "u"}); err != nil {
		t.Fatalf("SubmitText failed: %v", err)
	}
	forbidden := []string{
		"accountId", "account_id", "appid", "appId", "app_id",
		"cloudId", "cloud_id", "access_token", "accessToken",
	}
	for _, key := range forbidden {
		if _, ok := captured.form[key]; ok {
			t.Errorf("form must not contain identity param %s", key)
		}
	}
}

// TestRedirectNotFollowed 签名与 Host、Path 绑定，跟随 3xx 换域名后签名必然失效。
// 客户端必须把 3xx 当成响应返回（进而报「不是 JSON」），而不是静默跟到登录页。
func TestRedirectNotFollowed(t *testing.T) {
	var redirectTargetHit bool
	target := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		redirectTargetHit = true
		_, _ = writer.Write([]byte(`{"taskId":"should-not-reach"}`))
	}))
	defer target.Close()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, req *http.Request) {
		http.Redirect(writer, req, target.URL, http.StatusFound)
	}))
	defer server.Close()

	client := newStubClient(t, server)
	_, err := client.SubmitText(&SubmitTextRequest{Text: "hi"})
	if err == nil {
		t.Fatal("redirect response should surface as error, not be followed")
	}
	if redirectTargetHit {
		t.Error("client must not follow redirect: signature would be invalid for the new host")
	}
}

// TestAuditDetailUnmarshalTypePreserved 明细里的未建模字段应保持原始 JSON，
// 数字不能因中转丢精度。
func TestAuditDetailUnmarshalTypePreserved(t *testing.T) {
	var detail AuditDetail
	raw := `{"type":47,"probability":0.9999166666666667,"words":["a","b"]}`
	if err := json.Unmarshal([]byte(raw), &detail); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	if detail.Type != 47 {
		t.Errorf("type = %d, want 47", detail.Type)
	}
	if got := string(detail.Extras["probability"]); got != "0.9999166666666667" {
		t.Errorf("probability = %s, want 0.9999166666666667", got)
	}
	if got := string(detail.Extras["words"]); got != `["a","b"]` {
		t.Errorf("words = %s", got)
	}
}
