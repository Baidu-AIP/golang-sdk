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
	server := newStubServer(t, http.StatusOK, `{"taskId":"tid-text","log_id":"lg-1"}`, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	result, err := client.SubmitText(&SubmitTextRequest{
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

	// 响应原样返回，不做解析
	if result != `{"taskId":"tid-text","log_id":"lg-1"}` {
		t.Errorf("result = %q, want raw response body", result)
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
	result, err := client.SubmitImage(&SubmitImageRequest{
		ImgURL:     "https://example.com/a.jpg",
		StrategyID: 2017191,
	})
	if err != nil {
		t.Fatalf("SubmitImage failed: %v", err)
	}
	if result != `{"taskId":"tid-img"}` {
		t.Errorf("result = %q", result)
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

// TestPullResultPaths 三个模态各自独立的 result 路径，taskId 走 form，
// 响应报文原样返回。
func TestPullResultPaths(t *testing.T) {
	body := `{"taskId":"tid","status":"PROCESSING","createTime":"2026-09-04T11:42:51"}`
	cases := []struct {
		name string
		pull func(client *FusionCensorClient, taskID string) (string, error)
		path string
	}{
		{"text", (*FusionCensorClient).PullTextResult, "/api/v1/fusion/fusion_censor/v1/text/result"},
		{"image", (*FusionCensorClient).PullImageResult, "/api/v1/fusion/fusion_censor/v1/image/result"},
		{"video", (*FusionCensorClient).PullVideoResult, "/api/v1/fusion/fusion_censor/v1/video/result"},
	}
	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			var captured capturedRequest
			server := newStubServer(t, http.StatusOK, body, &captured)
			defer server.Close()

			client := newStubClient(t, server)
			result, err := testCase.pull(client, "tid")
			if err != nil {
				t.Fatalf("pull failed: %v", err)
			}
			if captured.path != testCase.path {
				t.Errorf("path = %s, want %s", captured.path, testCase.path)
			}
			if got := captured.form.Get("taskId"); got != "tid" {
				t.Errorf("taskId = %q, want tid", got)
			}
			if result != body {
				t.Errorf("result = %q, want raw body", result)
			}
		})
	}
}

// TestNon2xxStillReturnsBody 服务端在 4xx/5xx 上仍返回完整错误报文，
// 客户端必须把它当正常返回值交给调用方，而不是当传输失败。
func TestNon2xxStillReturnsBody(t *testing.T) {
	body := `{"log_id":"lg-1","error_code":6,"error_msg":"No permission to access data"}`
	var captured capturedRequest
	server := newStubServer(t, http.StatusForbidden, body, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	result, err := client.SubmitText(&SubmitTextRequest{Text: "hi"})
	if err != nil {
		t.Fatalf("SubmitText should not return transport error: %v", err)
	}
	if result != body {
		t.Errorf("result = %q, want raw error body", result)
	}
}

// TestGatewayErrorBodyReturned API Gateway 拦下请求时返回的是 BCE 通用错误格式
// （code / message / requestId），与业务错误的 error_code 形状不同。
// SDK 不解析报文，两种形状都原样交给调用方。
func TestGatewayErrorBodyReturned(t *testing.T) {
	body := `{"message":"IamSignatureInvalid, cause: Could not find credential.",` +
		`"code":"IamSignatureInvalid","requestId":"req-123"}`
	var captured capturedRequest
	server := newStubServer(t, http.StatusForbidden, body, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	result, err := client.SubmitText(&SubmitTextRequest{Text: "hi"})
	if err != nil {
		t.Fatalf("SubmitText should not return transport error: %v", err)
	}
	if result != body {
		t.Errorf("result = %q, want raw gateway error body", result)
	}
}

// TestNonJSONResponseReturnedAsIs 报文不是 JSON（如网关返回 HTML 跳转页）时，
// SDK 不做解析也就不该报错，原样返回让调用方自己判断。
func TestNonJSONResponseReturnedAsIs(t *testing.T) {
	body := `<html>302 Found</html>`
	var captured capturedRequest
	server := newStubServer(t, http.StatusOK, body, &captured)
	defer server.Close()

	client := newStubClient(t, server)
	result, err := client.SubmitText(&SubmitTextRequest{Text: "hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result != body {
		t.Errorf("result = %q, want %q", result, body)
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
// 客户端必须把 3xx 当成响应返回，而不是静默跟到登录页。
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
	result, err := client.SubmitText(&SubmitTextRequest{Text: "hi"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if redirectTargetHit {
		t.Error("client must not follow redirect: signature would be invalid for the new host")
	}
	if strings.Contains(result, "should-not-reach") {
		t.Errorf("redirect was followed, result = %q", result)
	}
}

// TestTransportErrorReturnsError 服务端不可达时必须返回 error，
// 而不是把空串当成正常响应。
func TestTransportErrorReturnsError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	endpoint := server.URL
	server.Close() // 关掉服务端，制造连接失败

	client, err := NewFusionClientWithEndpoint("ak", "sk", endpoint)
	if err != nil {
		t.Fatalf("create client failed: %v", err)
	}
	result, err := client.SubmitText(&SubmitTextRequest{Text: "hi"})
	if err == nil {
		t.Fatal("unreachable server should return error")
	}
	if result != "" {
		t.Errorf("result should be empty on transport error, got %q", result)
	}
}
