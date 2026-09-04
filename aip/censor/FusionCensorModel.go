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
	"net/url"
	"strconv"
)

// 审核结论类型（响应报文里 conclusionType 的取值）。
//
// SDK 不解析响应，这些常量供调用方自行解析报文后做比对。
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

// 任务状态（响应报文里 status 的取值）。
//
// PROVISIONING / PROCESSING 是非终态，此时报文里没有 conclusion、
// conclusionType、data、finishTime 四个键；轮询到 FINISHED 或 ERROR 才有结论。
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
