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

package fusion

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// __defaultTimeout 单次请求超时。融合审核的 submit / result 都是同步快返回接口。
const __defaultTimeout = 20 * time.Second

// Transport 融合审核的传输层：拼 URL、签名、发 form-urlencoded 请求、解析报文。
// 由上层业务客户端持有，本身不含任何审核业务语义。
type Transport struct {
	credentials Credentials
	endpoint    string
	pathPrefix  string
	httpClient  *http.Client
}

// NewTransport 创建传输层。endpoint 形如 https://host[:port]，不含路径；
// pathPrefix 是接口路径的公共前缀，会拼在每次请求的相对路径之前。
func NewTransport(ak string, sk string, endpoint string, pathPrefix string) (*Transport, error) {
	if ak == "" || sk == "" {
		return nil, errors.New("fusion: ak and sk must not be empty")
	}
	if endpoint == "" {
		return nil, errors.New("fusion: endpoint must not be empty")
	}
	return &Transport{
		credentials: Credentials{AccessKeyID: ak, SecretAccessKey: sk},
		endpoint:    strings.TrimRight(endpoint, "/"),
		pathPrefix:  strings.TrimRight(pathPrefix, "/"),
		httpClient:  newHTTPClient(),
	}, nil
}

// newHTTPClient 创建默认 http.Client。
//
// 显式关闭重定向跟随：签名与请求的 Host、Path 绑定，跟随 3xx 换域名后
// 原签名必然失效，静默跟随只会得到一个难以定位的鉴权失败。
func newHTTPClient() *http.Client {
	return &http.Client{
		Timeout: __defaultTimeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
}

// Endpoint 返回当前使用的服务域名。
func (transport *Transport) Endpoint() string {
	return transport.endpoint
}

// SetSessionToken 使用 STS 临时凭证时设置 session token；传空串表示清除。
func (transport *Transport) SetSessionToken(sessionToken string) {
	transport.credentials.SessionToken = sessionToken
}

// SetTimeout 设置单次请求超时。
func (transport *Transport) SetTimeout(timeout time.Duration) {
	if timeout > 0 {
		transport.httpClient.Timeout = timeout
	}
}

// SetHTTPClient 替换底层 http.Client，用于自定义连接池、代理等。
func (transport *Transport) SetHTTPClient(httpClient *http.Client) {
	if httpClient != nil {
		transport.httpClient = httpClient
	}
}

// PostForm 发送一次已签名的 form-urlencoded 请求，返回响应体原文。
//
// 不解析报文：审核结果的字段集不封闭（明细里的 probability、location、hits
// 等按命中标签动态出现，且随服务端迭代增加），在 SDK 里建模只会静默丢字段。
// 交给调用方按需自行解析。
//
// 返回的 error 只表示传输层失败。HTTP 4xx/5xx 不算错误 —— 服务端在非 2xx 上
// 仍返回完整的业务错误报文（error_code / error_msg），调用方需要读到它。
func (transport *Transport) PostForm(path string, form url.Values) (string, error) {
	requestURL := transport.endpoint + transport.pathPrefix + path
	req, err := http.NewRequest(http.MethodPost, requestURL, strings.NewReader(form.Encode()))
	if err != nil {
		return "", fmt.Errorf("fusion: build request failed: %w", err)
	}
	// Content-Type 参与签名，必须在签名之前设置。
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	transport.credentials.SignRequest(req, time.Now(), __defaultExpirationInSeconds)

	resp, err := transport.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("fusion: request %s failed: %w", path, err)
	}
	defer func(body io.ReadCloser) {
		_ = body.Close()
	}(resp.Body)

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("fusion: read response of %s failed: %w", path, err)
	}
	return string(raw), nil
}
