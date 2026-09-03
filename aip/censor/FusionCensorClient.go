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
	"errors"
	"net/http"
	"net/url"
	"time"

	"github.com/Baidu-AIP/golang-sdk/aip/censor/internal/fusion"
)

// __fusionNorthChinaEndpoint 华北地域服务域名。
//
// 当前只部署了华北，所以客户端不暴露地域选择。华东（苏州）、华南（广州）
// 上线后再新增带地域参数的构造函数，本函数保持为华北的快捷入口。
const __fusionNorthChinaEndpoint = "https://icr.bj.baidubce.com"

// __fusionPathPrefix 对客接口的路径前缀。
//
// 走内网直连，不经 OpenAPI 网关。服务端同时挂了 /rest/2.0/fusion（网关改写后的
// 对外路径）与本前缀两个等价入口，这里固定用后者。
const __fusionPathPrefix = "/api/v1/fusion"

// 六个对客接口的相对路径（拼在路径前缀之后）。
const (
	__fusionTextSubmitPath  = "/fusion_censor/v1/text/submit"
	__fusionTextResultPath  = "/fusion_censor/v1/text/result"
	__fusionImageSubmitPath = "/fusion_censor/v1/image/submit"
	__fusionImageResultPath = "/fusion_censor/v1/image/result"
	__fusionVideoSubmitPath = "/fusion_censor/v1/video/submit"
	__fusionVideoResultPath = "/fusion_censor/v1/video/result"
)

// FusionCensorClient 融合审核（文本 / 图像 / 长视频）客户端。
//
// 用 ak/sk 初始化，每次请求内部按 bce-auth-v1 协议现算 IAM 签名，
// 无需也不会去换取 access_token。签名与传输实现在 internal/fusion 包。
// 并发安全，建议全局复用一个实例。
type FusionCensorClient struct {
	transport *fusion.Transport
}

// NewFusionClient 创建融合审核客户端，接入华北地域。
//
// 当前服务只部署了华北，因此无需选择地域。
func NewFusionClient(ak string, sk string) (*FusionCensorClient, error) {
	return NewFusionClientWithEndpoint(ak, sk, __fusionNorthChinaEndpoint)
}

// NewFusionClientWithEndpoint 用指定域名创建客户端，供联调或私有化部署使用。
// endpoint 形如 https://host[:port]，不含路径。
func NewFusionClientWithEndpoint(ak string, sk string, endpoint string) (*FusionCensorClient, error) {
	transport, err := fusion.NewTransport(ak, sk, endpoint, __fusionPathPrefix)
	if err != nil {
		return nil, err
	}
	return &FusionCensorClient{transport: transport}, nil
}

// Endpoint 返回当前客户端使用的服务域名。
func (client *FusionCensorClient) Endpoint() string {
	return client.transport.Endpoint()
}

// SetSessionToken 使用 STS 临时凭证时设置 session token；传空串表示清除。
func (client *FusionCensorClient) SetSessionToken(sessionToken string) {
	client.transport.SetSessionToken(sessionToken)
}

// SetTimeout 设置单次请求超时。
func (client *FusionCensorClient) SetTimeout(timeout time.Duration) {
	client.transport.SetTimeout(timeout)
}

// SetHTTPClient 替换底层 http.Client，用于自定义连接池、代理等。
func (client *FusionCensorClient) SetHTTPClient(httpClient *http.Client) {
	client.transport.SetHTTPClient(httpClient)
}

// SubmitText 提交文本审核任务，同步返回 taskId，结论用 PullTextResult 轮询或走回调。
func (client *FusionCensorClient) SubmitText(request *SubmitTextRequest) (*FusionResponse, error) {
	if request == nil || request.Text == "" {
		return nil, errors.New("censor: text must not be empty")
	}
	return client.post(__fusionTextSubmitPath, request.toForm())
}

// SubmitImage 提交图像审核任务。仅支持 URL 传入，base64 会被服务端拒绝（282801）。
func (client *FusionCensorClient) SubmitImage(request *SubmitImageRequest) (*FusionResponse, error) {
	if request == nil || request.ImgURL == "" {
		return nil, errors.New("censor: imgUrl must not be empty")
	}
	return client.post(__fusionImageSubmitPath, request.toForm())
}

// SubmitVideo 提交长视频审核任务。视频转存与抽帧由服务端异步推进，耗时明显长于文本 / 图像。
func (client *FusionCensorClient) SubmitVideo(request *SubmitVideoRequest) (*FusionResponse, error) {
	if request == nil || request.URL == "" {
		return nil, errors.New("censor: video url must not be empty")
	}
	return client.post(__fusionVideoSubmitPath, request.toForm())
}

// PullTextResult 拉取文本审核结果。
//
// 未完成时返回 Status=PROCESSING 而不是报错，不要当失败处理。
// taskId 必须用提交时对应模态的接口拉取：拉错模态会返回 282006（任务不存在）。
func (client *FusionCensorClient) PullTextResult(taskID string) (*FusionResponse, error) {
	return client.pullResult(__fusionTextResultPath, taskID)
}

// PullImageResult 拉取图像审核结果。语义同 PullTextResult。
func (client *FusionCensorClient) PullImageResult(taskID string) (*FusionResponse, error) {
	return client.pullResult(__fusionImageResultPath, taskID)
}

// PullVideoResult 拉取长视频审核结果。Data 是含 frames / audios 的对象，
// 用 FusionResponse.UnmarshalVideoData 解析。
func (client *FusionCensorClient) PullVideoResult(taskID string) (*FusionResponse, error) {
	return client.pullResult(__fusionVideoResultPath, taskID)
}

func (client *FusionCensorClient) pullResult(path string, taskID string) (*FusionResponse, error) {
	if taskID == "" {
		return nil, errors.New("censor: taskId must not be empty")
	}
	form := url.Values{}
	form.Set("taskId", taskID)
	return client.post(path, form)
}

// post 发一次请求并组装响应。
//
// 返回的 error 只表示传输层或报文解析失败；业务错误（无权限、参数非法等）
// 通过 FusionResponse.ErrorCode 返回，因为服务端在 4xx/5xx 上仍会带完整错误报文。
func (client *FusionCensorClient) post(path string, form url.Values) (*FusionResponse, error) {
	var response FusionResponse
	raw, err := client.transport.PostForm(path, form, &response)
	if err != nil {
		return nil, err
	}
	response.Raw = raw
	return &response, nil
}
