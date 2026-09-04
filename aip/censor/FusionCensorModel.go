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
	"fmt"
	"net/url"
	"strconv"
)

// 审核结论类型（对应响应里的 conclusionType）。
const (
	// ConclusionTypeCompliant 合规
	ConclusionTypeCompliant = 1
	// ConclusionTypeNonCompliant 不合规
	ConclusionTypeNonCompliant = 2
	// ConclusionTypeSuspected 疑似
	ConclusionTypeSuspected = 3
	// ConclusionTypeFailed 审核失败
	ConclusionTypeFailed = 4
)

// 任务状态（对应响应里的 status）。
const (
	// FusionStatusProvisioning 任务已创建，尚未开始审核
	FusionStatusProvisioning = "PROVISIONING"
	// FusionStatusProcessing 审核中。注意：未完成返回该状态而非报错，不要当失败处理
	FusionStatusProcessing = "PROCESSING"
	// FusionStatusFinished 审核完成，此时才有 conclusion / conclusionType / data
	FusionStatusFinished = "FINISHED"
	// FusionStatusError 审核失败，error_code / error_msg 给出原因
	FusionStatusError = "ERROR"
)

// 视频审核类型（SubmitVideoRequest.DetectType）。
const (
	// VideoDetectTypeFrameAndAudio 视频帧 + 音频同时过审
	VideoDetectTypeFrameAndAudio = 0
	// VideoDetectTypeFrameOnly 仅视频帧（服务端默认值）
	VideoDetectTypeFrameOnly = 1
	// VideoDetectTypeAudioOnly 仅音频
	VideoDetectTypeAudioOnly = 2
)

// __defaultStrategyID 未指定策略时使用的默认策略 ID。
//
// 服务端把 strategyId 标为可选，但真正落到小模型时 0 会被拒（返回 282910
// strategy not exist），所以 SDK 侧统一兜底成 1 —— 语义是「按小模型默认策略审核、
// 不参与转审」，与服务端 FusionConstants.DEFAULT_STRATEGY_ID 一致。
const __defaultStrategyID int64 = 1

// SubmitTextRequest 文本融合审核-提交 入参。
type SubmitTextRequest struct {
	// Text 审核文本，必填。≤2000 字符（空 → 282874，超长 → 282909）
	Text string

	// StrategyID 小模型策略 ID。为 0 时 SDK 按 __defaultStrategyID 兜底
	StrategyID int64

	// UserID 客户自定义用户 ID，可选
	UserID string

	// ExtStr 扩展参数，JSON 字符串，可选。key 固定 extStr1~3，各值 ≤128 字符（违规 → 282004）
	ExtStr string

	// CallbackURL 客户回调地址，可选。http/https，≤2048 字符，不允许内网地址。为空表示不回调
	CallbackURL string

	// UserIP 客户自定义参数，可选
	UserIP string

	// PhoneSha256 客户自定义参数，可选
	PhoneSha256 string

	// DeviceID 客户自定义参数，可选
	DeviceID string
}

// SubmitImageRequest 图像融合审核-提交 入参。
type SubmitImageRequest struct {
	// ImgURL 图像 URL，必填。仅支持 URL（传 base64 → 282801）；建议 JPG/PNG/WebP/BMP，≤10MB
	ImgURL string

	// StrategyID 小模型策略 ID。为 0 时 SDK 按 __defaultStrategyID 兜底
	StrategyID int64

	// UserID 客户自定义用户 ID，可选
	UserID string

	// ExtStr 扩展参数，JSON 字符串，可选。key 固定 extStr1~3，各值 ≤128 字符
	ExtStr string

	// CallbackURL 客户回调地址，可选。http/https，≤2048 字符，不允许内网地址
	CallbackURL string

	// UserIP 客户自定义参数，可选
	UserIP string

	// PhoneSha256 客户自定义参数，可选
	PhoneSha256 string

	// DeviceID 客户自定义参数，可选
	DeviceID string
}

// SubmitVideoRequest 长视频融合审核-提交 入参。
type SubmitVideoRequest struct {
	// URL 视频地址，必填。为空 → 282875；下载失败 → 282876；超 2G → 282978。
	// 格式白名单：mp4/avi/flv/mov/wmv/ts/mpeg/3gpp/asf/f4v/mkv/m4a/mp3/mp2/mpg/ogg/mts/wma/webm/m3u8
	URL string

	// DetectType 审核类型，取 VideoDetectType* 常量。
	// nil 表示不传，服务端按仅视频帧（1）处理
	DetectType *int

	// ExtID 客户侧视频唯一标识，可选
	ExtID string

	// StrategyID 小模型策略 ID。为 0 时 SDK 按 __defaultStrategyID 兜底
	StrategyID int64

	// UserID 客户自定义用户 ID，可选
	UserID string

	// ExtStr 扩展参数，JSON 字符串，可选。key 固定 extStr1~3，各值 ≤128 字符
	ExtStr string

	// CallbackURL 客户回调地址，可选。http/https，≤2048 字符，不允许内网地址
	CallbackURL string

	// UserIP 客户自定义参数，可选
	UserIP string

	// PhoneSha256 客户自定义参数，可选
	PhoneSha256 string

	// DeviceID 客户自定义参数，可选
	DeviceID string
}

// FusionResponse 三个模态 submit / result 共用的响应体。
//
// submit 只返回 TaskID；result 未完成时只有 Status 与 CreateTime，
// 到终态（FINISHED / ERROR）才带 Conclusion / ConclusionType / Data。
type FusionResponse struct {
	// TaskID 任务 ID
	TaskID string `json:"taskId"`

	// LogID 日志串联 ID，成功与失败都返回。
	// 报障时必须提供该值 —— 它是服务端反查日志的唯一锚点。result 阶段恒等于 TaskID
	LogID string `json:"log_id"`

	// ErrorCode 错误码，失败才返回。取值见 error_code 对照表（如 6 无权限、282006 任务不存在）
	ErrorCode int64 `json:"error_code"`

	// ErrorMsg 错误信息，失败才返回。
	// 注意：pull 时该字段取自任务库里的原始信息，与 ErrorCode 的官方文案可能不一致，
	// 业务判断请只依赖 ErrorCode
	ErrorMsg string `json:"error_msg"`

	// Status 任务状态，取 FusionStatus* 常量
	Status string `json:"status"`

	// CreateTime 任务创建时间，ISO-8601 字符串，如 2026-08-28T02:02:59.000+00:00
	CreateTime string `json:"createTime"`

	// FinishTime 审核完成时间，终态才返回
	FinishTime string `json:"finishTime"`

	// Conclusion 审核结论：合规 / 不合规 / 疑似 / 审核失败。终态才返回
	Conclusion string `json:"conclusion"`

	// ConclusionType 结论类型，取 ConclusionType* 常量。终态才返回。
	// 注意：非终态可能返回 0（值域外），表示尚无结论
	ConclusionType int `json:"conclusionType"`

	// Data 违规明细。文本/图像是明细数组，长视频是含 taskDuration/frames/audios 的对象；
	// 合规且未命中白名单时不返回。
	//
	// 保持 json.RawMessage 是有意的：服务端明细字段集不封闭（words、probability、
	// location、hits 等按标签动态出现），建模成固定结构会静默丢字段。
	// 需要结构化访问时用 UnmarshalAuditDetails / UnmarshalVideoData。
	Data json.RawMessage `json:"data"`

	// Raw 原始响应报文，排查问题时用
	Raw string `json:"-"`

	// GatewayCode 网关层错误码，如 IamSignatureInvalid（签名无效）、
	// ResourceNotFound（路径不存在）。
	//
	// 请求被 API Gateway 在到达审核服务之前拦下时，返回的是 BCE 通用错误格式
	// {"code":"...","message":"...","requestId":"..."}，与业务错误的
	// {"error_code":6,...} 是两种不同形状 —— 只看 ErrorCode 会把网关拒绝误判成成功。
	GatewayCode string `json:"code"`

	// GatewayMessage 网关层错误描述
	GatewayMessage string `json:"message"`

	// RequestID 网关请求 ID，网关层报障时提供
	RequestID string `json:"requestId"`
}

// IsSuccess 请求是否成功。
//
// 需要同时检查业务错误码与网关错误码：两者报文形状不同，各自独立出现。
func (response *FusionResponse) IsSuccess() bool {
	return response.ErrorCode == 0 && response.GatewayCode == ""
}

// ErrorDescription 返回可读的失败原因，统一业务错误与网关错误两种报文形状。
// 成功时返回空串。
func (response *FusionResponse) ErrorDescription() string {
	if response.GatewayCode != "" {
		return fmt.Sprintf("gateway %s: %s (requestId=%s)",
			response.GatewayCode, response.GatewayMessage, response.RequestID)
	}
	if response.ErrorCode != 0 {
		return fmt.Sprintf("error_code=%d error_msg=%s (log_id=%s)",
			response.ErrorCode, response.ErrorMsg, response.LogID)
	}
	return ""
}

// IsTerminal 任务是否已到终态（不会再变化）。非终态时不应读取结论字段。
func (response *FusionResponse) IsTerminal() bool {
	return response.Status == FusionStatusFinished || response.Status == FusionStatusError
}

// AuditDetail 文本 / 图像审核明细（Data 数组的元素）。
//
// 只建模稳定字段，其余字段（words、probability、location、hits、conclusion 等）
// 原样保留在 Extras 里 —— 服务端该结构字段集不封闭，且大模型标签可能只有
// AgentType / AgentSubType 而没有 Type / SubType。
type AuditDetail struct {
	ErrorCode      int64  `json:"error_code"`
	ErrorMsg       string `json:"error_msg"`
	Type           int    `json:"type"`
	SubType        int    `json:"subType"`
	ConclusionType int    `json:"conclusionType"`
	Msg            string `json:"msg"`

	// AgentType 大模型一级标签
	AgentType string `json:"agentType"`

	// AgentSubType 大模型二级标签，仅部分标签填充
	AgentSubType string `json:"agentSubType"`

	// Extras 未建模字段，原样保留
	Extras map[string]json.RawMessage `json:"-"`
}

// UnmarshalJSON 先按已知字段解析，再把剩余键收进 Extras。
func (detail *AuditDetail) UnmarshalJSON(data []byte) error {
	type plainAuditDetail AuditDetail
	var plain plainAuditDetail
	if err := json.Unmarshal(data, &plain); err != nil {
		return err
	}
	*detail = AuditDetail(plain)

	var all map[string]json.RawMessage
	if err := json.Unmarshal(data, &all); err != nil {
		return err
	}
	known := map[string]bool{
		"error_code": true, "error_msg": true, "type": true, "subType": true,
		"conclusionType": true, "msg": true, "agentType": true, "agentSubType": true,
	}
	detail.Extras = make(map[string]json.RawMessage)
	for key, value := range all {
		if !known[key] {
			detail.Extras[key] = value
		}
	}
	return nil
}

// UnmarshalAuditDetails 把文本 / 图像响应的 Data 解析成明细数组。
// Data 为空（合规且未命中白名单）时返回 nil, nil。
func (response *FusionResponse) UnmarshalAuditDetails() ([]AuditDetail, error) {
	if len(response.Data) == 0 {
		return nil, nil
	}
	var details []AuditDetail
	if err := json.Unmarshal(response.Data, &details); err != nil {
		return nil, err
	}
	return details, nil
}

// VideoData 长视频响应的 Data 结构。
type VideoData struct {
	// TaskDuration 视频时长，秒。上游未给出时该键不返回
	TaskDuration int `json:"taskDuration"`

	// Frames 违规 / 疑似视频帧。键恒存在，可能为空数组
	Frames []VideoFrame `json:"frames"`

	// Audios 违规 / 疑似音频片段。键恒存在，可能为空数组
	Audios []VideoAudio `json:"audios"`
}

// VideoFrame 长视频的一个违规 / 疑似视频帧。
type VideoFrame struct {
	// FrameTimeStamp 该帧在视频中的时间戳，单位秒。
	// 注意与 VideoAudio 的毫秒单位不一致，这是服务端口径，SDK 不做换算
	FrameTimeStamp int64 `json:"frameTimeStamp"`

	// FrameURL 帧图预签名地址，默认有效期 30 分钟。
	// 过期后需重新调 result 接口换取，不要长期缓存
	FrameURL string `json:"frameUrl"`

	// FrameThumbnailURL 缩略图地址，当前与 FrameURL 相同
	FrameThumbnailURL string `json:"frameThumbnailUrl"`

	// Data 该帧的违规明细，键恒存在，可能为空数组
	Data []AuditDetail `json:"data"`
}

// VideoAudio 长视频的一个违规 / 疑似音频片段。
//
// 音频片段的违规明细不在本层，而在 RawText[].Data 里 —— 语音转写文本才是被审核的对象。
type VideoAudio struct {
	// StartTime 片段起始时间，单位毫秒
	StartTime int64 `json:"startTime"`

	// EndTime 片段结束时间，单位毫秒
	EndTime int64 `json:"endTime"`

	// AudioURL 音频片段预签名地址，有效期同 VideoFrame.FrameURL
	AudioURL string `json:"audioUrl"`

	// AudioAuditResult 声纹类音频特征审核结论（如娇喘识别）。无该类命中时为空
	AudioAuditResult []AudioAuditResult `json:"audioAuditResult"`

	// RawText 语音转写（ASR）文本的审核结果。该片段无语音内容时为空。
	// 当前上游只返回拼接后的整段文本，通常只有 1 个元素
	RawText []VideoAudioText `json:"rawText"`
}

// AudioAuditResult 声纹类音频特征审核结论。
type AudioAuditResult struct {
	Type           int    `json:"type"`
	SubType        int    `json:"subType"`
	ConclusionType int    `json:"conclusionType"`
	Conclusion     string `json:"conclusion"`
	Msg            string `json:"msg"`
}

// VideoAudioText 一段语音转写文本及其审核结果。
type VideoAudioText struct {
	// StartTime 该段文本对应的开始时间，单位毫秒
	StartTime int64 `json:"startTime"`

	// EndTime 该段文本对应的结束时间，单位毫秒
	EndTime int64 `json:"endTime"`

	// Text 语音转写出的文本原文
	Text string `json:"text"`

	// ConclusionType 该段文本的审核结论码，取 ConclusionType* 常量
	ConclusionType int `json:"conclusionType"`

	// Conclusion 该段文本的审核结论中文描述
	Conclusion string `json:"conclusion"`

	// Data 该段文本的违规明细，键恒存在，可能为空数组
	Data []AuditDetail `json:"data"`
}

// UnmarshalVideoData 把长视频响应的 Data 解析成 VideoData。
// Data 为空时返回 nil, nil。
func (response *FusionResponse) UnmarshalVideoData() (*VideoData, error) {
	if len(response.Data) == 0 {
		return nil, nil
	}
	var data VideoData
	if err := json.Unmarshal(response.Data, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

// toForm 组装文本审核的 form 参数。
func (request *SubmitTextRequest) toForm() url.Values {
	form := url.Values{}
	form.Set("text", request.Text)
	addStrategyID(form, request.StrategyID)
	addIfNotEmpty(form, "userId", request.UserID)
	addIfNotEmpty(form, "extStr", request.ExtStr)
	addIfNotEmpty(form, "callbackUrl", request.CallbackURL)
	addIfNotEmpty(form, "userIp", request.UserIP)
	addIfNotEmpty(form, "phoneSha256", request.PhoneSha256)
	addIfNotEmpty(form, "deviceId", request.DeviceID)
	return form
}

// toForm 组装图像审核的 form 参数。
func (request *SubmitImageRequest) toForm() url.Values {
	form := url.Values{}
	form.Set("imgUrl", request.ImgURL)
	addStrategyID(form, request.StrategyID)
	addIfNotEmpty(form, "userId", request.UserID)
	addIfNotEmpty(form, "extStr", request.ExtStr)
	addIfNotEmpty(form, "callbackUrl", request.CallbackURL)
	addIfNotEmpty(form, "userIp", request.UserIP)
	addIfNotEmpty(form, "phoneSha256", request.PhoneSha256)
	addIfNotEmpty(form, "deviceId", request.DeviceID)
	return form
}

// toForm 组装长视频审核的 form 参数。
func (request *SubmitVideoRequest) toForm() url.Values {
	form := url.Values{}
	form.Set("url", request.URL)
	if request.DetectType != nil {
		form.Set("detectType", strconv.Itoa(*request.DetectType))
	}
	addIfNotEmpty(form, "extId", request.ExtID)
	addStrategyID(form, request.StrategyID)
	addIfNotEmpty(form, "userId", request.UserID)
	addIfNotEmpty(form, "extStr", request.ExtStr)
	addIfNotEmpty(form, "callbackUrl", request.CallbackURL)
	addIfNotEmpty(form, "userIp", request.UserIP)
	addIfNotEmpty(form, "phoneSha256", request.PhoneSha256)
	addIfNotEmpty(form, "deviceId", request.DeviceID)
	return form
}

func addStrategyID(form url.Values, strategyID int64) {
	if strategyID <= 0 {
		strategyID = __defaultStrategyID
	}
	form.Set("strategyId", strconv.FormatInt(strategyID, 10))
}

func addIfNotEmpty(form url.Values, key string, value string) {
	if value != "" {
		form.Set(key, value)
	}
}
