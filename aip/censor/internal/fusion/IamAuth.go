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

// Package fusion 是融合审核客户端的内部实现：IAM 签名与 HTTP 传输。
//
// 本包位于 internal 下，只允许 aip/censor 使用，不对外暴露。
// 对客 API 全部在 aip/censor 包。
package fusion

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// IAM（bce-auth-v1）签名相关常量。与 Java 侧 com.baidubce.auth.BceV1Signer 保持一致。
//
// 注意：不要与 baseClient.Auth.setHeader 混用 —— 那份实现用 url.QueryEscape 做转义
// （空格编成 + 、* 不编码）、且只签 host 与 x-bce-date，不符合 BCE 协议，
// 仅为兼容既有接口保留。融合审核必须走本包。
const (
	__bceAuthVersion = "bce-auth-v1"

	// __defaultExpirationInSeconds 签名有效期，与 BCE SDK 默认值一致（半小时）
	__defaultExpirationInSeconds = 1800

	__headerHost             = "Host"
	__headerAuthorization    = "Authorization"
	__headerContentType      = "Content-Type"
	__headerBceDate          = "x-bce-date"
	__headerBceSecurityToken = "x-bce-security-token"
	__bceTimestampLayout     = "2006-01-02T15:04:05Z"
)

// Credentials 云上账号凭证。ak/sk 从百度智能云控制台获取；
// SessionToken 仅在使用 STS 临时凭证时需要填写。
type Credentials struct {
	AccessKeyID     string
	SecretAccessKey string
	SessionToken    string
}

// SignRequest 按 bce-auth-v1 协议为请求生成 Authorization 头。
//
// 签名串构造过程（顺序不可调整）：
//
//	authString       = bce-auth-v1/{ak}/{timestamp}/{expirationInSeconds}
//	signingKey       = HMAC-SHA256-HEX(sk, authString)
//	canonicalRequest = METHOD \n CanonicalURI \n CanonicalQueryString \n CanonicalHeaders
//	signature        = HMAC-SHA256-HEX(signingKey, canonicalRequest)
//	Authorization    = authString/{signedHeaders}/{signature}
//
// 参与签名的头固定为 content-type;host;x-bce-date（用 STS 时追加
// x-bce-security-token），显式写进 signedHeaders，避免依赖服务端的默认头集合。
//
// Content-Type 参与签名，调用方必须在本函数之前设好该头。
// expirationInSeconds ≤ 0 时按默认值 1800 处理。
func (credentials *Credentials) SignRequest(req *http.Request, timestamp time.Time,
	expirationInSeconds int) {
	if expirationInSeconds <= 0 {
		expirationInSeconds = __defaultExpirationInSeconds
	}

	// Host 必须与实际发出的请求一致：Go 用 req.Host 决定 wire 上的 Host 头，
	// 只设 req.Header 不生效，两处都要写。
	host := hostHeader(req.URL)
	req.Host = host
	req.Header.Set(__headerHost, host)
	req.Header.Set(__headerBceDate, timestamp.UTC().Format(__bceTimestampLayout))
	if credentials.SessionToken != "" {
		req.Header.Set(__headerBceSecurityToken, credentials.SessionToken)
	}

	headersToSign := []string{
		strings.ToLower(__headerContentType),
		strings.ToLower(__headerHost),
		__headerBceDate,
	}
	if credentials.SessionToken != "" {
		headersToSign = append(headersToSign, __headerBceSecurityToken)
	}
	sort.Strings(headersToSign)

	authString := __bceAuthVersion + "/" + credentials.AccessKeyID + "/" +
		timestamp.UTC().Format(__bceTimestampLayout) + "/" +
		strconv.Itoa(expirationInSeconds)
	signingKey := hmacSha256Hex(credentials.SecretAccessKey, authString)

	canonicalRequest := req.Method + "\n" +
		canonicalURIPath(req.URL.Path) + "\n" +
		canonicalQueryString(req.URL.Query()) + "\n" +
		canonicalHeaders(req.Header, headersToSign)
	signature := hmacSha256Hex(signingKey, canonicalRequest)

	req.Header.Set(__headerAuthorization,
		authString+"/"+strings.Join(headersToSign, ";")+"/"+signature)
}

// hostHeader 生成 Host 头：非默认端口才拼端口，与 apache http client 行为一致。
func hostHeader(u *url.URL) string {
	host := u.Hostname()
	port := u.Port()
	if port == "" {
		return host
	}
	if (u.Scheme == "http" && port == "80") || (u.Scheme == "https" && port == "443") {
		return host
	}
	return host + ":" + port
}

// canonicalURIPath 规范化请求路径。normalize 后把 %2F 还原成 /，保留路径分隔语义。
func canonicalURIPath(path string) string {
	if path == "" {
		return "/"
	}
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return strings.ReplaceAll(normalize(path), "%2F", "/")
}

// canonicalQueryString 规范化 query：逐项 normalize 后按字典序排序，用 & 连接。
// Authorization 参数不参与签名。
func canonicalQueryString(query url.Values) string {
	if len(query) == 0 {
		return ""
	}
	items := make([]string, 0, len(query))
	for key, values := range query {
		if strings.EqualFold(key, __headerAuthorization) {
			continue
		}
		for _, value := range values {
			items = append(items, normalize(key)+"="+normalize(value))
		}
	}
	sort.Strings(items)
	return strings.Join(items, "&")
}

// canonicalHeaders 规范化待签名头：key 转小写、key 与 value 分别 normalize，
// 按字典序排序后用换行连接。
func canonicalHeaders(header http.Header, headersToSign []string) string {
	items := make([]string, 0, len(headersToSign))
	for _, key := range headersToSign {
		value := strings.TrimSpace(header.Get(key))
		items = append(items, normalize(strings.ToLower(strings.TrimSpace(key)))+":"+
			normalize(value))
	}
	sort.Strings(items)
	return strings.Join(items, "\n")
}

// normalize 按 RFC 3986 做百分号编码：除 A-Za-z0-9-._~ 外全部编码，十六进制用大写。
// 不能用 url.QueryEscape —— 它把空格编成 +、且不编码 * 等字符，与 BCE 协议不符。
func normalize(value string) string {
	var builder strings.Builder
	builder.Grow(len(value))
	for i := 0; i < len(value); i++ {
		char := value[i]
		if isURIUnreserved(char) {
			builder.WriteByte(char)
		} else {
			builder.WriteByte('%')
			builder.WriteByte(__upperHex[char>>4])
			builder.WriteByte(__upperHex[char&0x0F])
		}
	}
	return builder.String()
}

const __upperHex = "0123456789ABCDEF"

func isURIUnreserved(char byte) bool {
	switch {
	case char >= 'a' && char <= 'z':
		return true
	case char >= 'A' && char <= 'Z':
		return true
	case char >= '0' && char <= '9':
		return true
	case char == '-' || char == '.' || char == '_' || char == '~':
		return true
	}
	return false
}

func hmacSha256Hex(key string, data string) string {
	mac := hmac.New(sha256.New, []byte(key))
	mac.Write([]byte(data))
	return hex.EncodeToString(mac.Sum(nil))
}
