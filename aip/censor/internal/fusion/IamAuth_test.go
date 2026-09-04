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
	"net/http"
	"strings"
	"testing"
	"time"
)

// __signerRefTimestamp 与 Java 参照实现取同一时刻（2026-01-01T00:00:00Z），
// 用于逐字节比对 Authorization 头。
var __signerRefTimestamp = time.Unix(1767225600, 0).UTC()

// TestSignRequestMatchesBceV1Signer 用 com.baidubce.auth.BceV1Signer 实跑出的
// Authorization 头作为期望值，锁定 Go 侧签名与 Java SDK 完全一致。
// 期望值由 bce-java-sdk 0.10.336 生成，ak/sk 为构造的测试值。
func TestSignRequestMatchesBceV1Signer(t *testing.T) {
	credentials := Credentials{
		AccessKeyID:     "ALTAKtestak123456",
		SecretAccessKey: "testsk0987654321abcdef",
	}

	cases := []struct {
		name     string
		url      string
		expected string
	}{
		{
			// 生产域名（华北），与 SDK 默认 endpoint 一致
			name: "prod host text submit",
			url:  "https://icr.bj.baidubce.com/api/v1/fusion/fusion_censor/v1/text/submit",
			expected: "bce-auth-v1/ALTAKtestak123456/2026-01-01T00:00:00Z/1800/content-type;host;x-bce-date/" +
				"732ee95b93bc60dcdf31fa37ebe7d4c1726bd47538a6dc93fe2aaa7453c6ba36",
		},
		{
			name: "text submit without query",
			url:  "https://icr-hb-qasandbox.baidu-int.com/api/v1/fusion/fusion_censor/v1/text/submit",
			expected: "bce-auth-v1/ALTAKtestak123456/2026-01-01T00:00:00Z/1800/content-type;host;x-bce-date/" +
				"6952a6a7409477432fe90bc8ddb0fd06f11730257ce4dd87b4d1831b27da6b32",
		},
		{
			name: "image result without query",
			url:  "https://icr-hb-qasandbox.baidu-int.com/api/v1/fusion/fusion_censor/v1/image/result",
			expected: "bce-auth-v1/ALTAKtestak123456/2026-01-01T00:00:00Z/1800/content-type;host;x-bce-date/" +
				"9e2241e07d811da599bdafb00104afcfa0a199a0ca38470657c22e173bd1258f",
		},
		{
			// 非默认端口要拼进 Host，query 含中文与空值
			name: "non default port with utf8 query",
			url: "https://icr-hb-qasandbox.baidu-int.com:8443/api/v1/fusion/fusion_censor/v1/video/submit" +
				"?a=1&b=%E4%B8%AD%E6%96%87&c=",
			expected: "bce-auth-v1/ALTAKtestak123456/2026-01-01T00:00:00Z/1800/content-type;host;x-bce-date/" +
				"c3bf7936cb18ae081a17f6b65fb33ced469e9b7bfd05c891f88aef7567025824",
		},
		{
			// query 含空格与 * ~ 等字符，验证 normalize 与 url.QueryEscape 的差异被正确处理
			name: "query needing rfc3986 encoding",
			url: "http://icr-hb-qasandbox.baidu-int.com/api/v1/fusion/fusion_censor/v1/text/result" +
				"?x%20y=a*b~c&empty=",
			expected: "bce-auth-v1/ALTAKtestak123456/2026-01-01T00:00:00Z/1800/content-type;host;x-bce-date/" +
				"ec283ceb7251b9109c23bb5390332bd0617c35c178c4c820d0952f115f98119f",
		},
	}

	for _, testCase := range cases {
		t.Run(testCase.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodPost, testCase.url, strings.NewReader("text=hi"))
			if err != nil {
				t.Fatalf("build request failed: %v", err)
			}
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

			credentials.SignRequest(req, __signerRefTimestamp, __defaultExpirationInSeconds)

			if got := req.Header.Get("Authorization"); got != testCase.expected {
				t.Errorf("Authorization mismatch\n got: %s\nwant: %s", got, testCase.expected)
			}
		})
	}
}

// TestSignRequestSetsRequiredHeaders Host 必须同时落到 req.Host（决定 wire 上的头）
// 与 req.Header，否则实际发出的 Host 与签名不一致，服务端会判签名失败。
func TestSignRequestSetsRequiredHeaders(t *testing.T) {
	credentials := Credentials{AccessKeyID: "ak", SecretAccessKey: "sk"}
	req, err := http.NewRequest(http.MethodPost,
		"https://icr-hb-qasandbox.baidu-int.com/api/v1/fusion/fusion_censor/v1/text/submit", nil)
	if err != nil {
		t.Fatalf("build request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	credentials.SignRequest(req, __signerRefTimestamp, __defaultExpirationInSeconds)

	if req.Host != "icr-hb-qasandbox.baidu-int.com" {
		t.Errorf("req.Host = %q, want icr-hb-qasandbox.baidu-int.com", req.Host)
	}
	if got := req.Header.Get("Host"); got != "icr-hb-qasandbox.baidu-int.com" {
		t.Errorf("Host header = %q, want icr-hb-qasandbox.baidu-int.com", got)
	}
	if got := req.Header.Get("x-bce-date"); got != "2026-01-01T00:00:00Z" {
		t.Errorf("x-bce-date = %q, want 2026-01-01T00:00:00Z", got)
	}
	if req.Header.Get("x-bce-security-token") != "" {
		t.Error("x-bce-security-token should be absent when SessionToken is empty")
	}
}

// TestSignRequestWithSessionToken 使用 STS 临时凭证时，token 头必须存在且参与签名。
func TestSignRequestWithSessionToken(t *testing.T) {
	credentials := Credentials{
		AccessKeyID:     "ak",
		SecretAccessKey: "sk",
		SessionToken:    "sts-token-value",
	}
	req, err := http.NewRequest(http.MethodPost,
		"https://icr-hb-qasandbox.baidu-int.com/api/v1/fusion/fusion_censor/v1/text/submit", nil)
	if err != nil {
		t.Fatalf("build request failed: %v", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	credentials.SignRequest(req, __signerRefTimestamp, __defaultExpirationInSeconds)

	if got := req.Header.Get("x-bce-security-token"); got != "sts-token-value" {
		t.Errorf("x-bce-security-token = %q, want sts-token-value", got)
	}
	authorization := req.Header.Get("Authorization")
	if !strings.Contains(authorization, "content-type;host;x-bce-date;x-bce-security-token") {
		t.Errorf("signedHeaders should include x-bce-security-token, got %s", authorization)
	}
}

// TestNormalize 验证 RFC 3986 编码：空格必须编成 %20 而非 +，* 必须编码，~ 必须保留。
func TestNormalize(t *testing.T) {
	cases := map[string]string{
		"":                                  "",
		"abcXYZ019":                         "abcXYZ019",
		"-._~":                              "-._~",
		"a b":                               "a%20b",
		"a*b":                               "a%2Ab",
		"a/b":                               "a%2Fb",
		"application/x-www-form-urlencoded": "application%2Fx-www-form-urlencoded",
		"中":                                 "%E4%B8%AD",
	}
	for input, want := range cases {
		if got := normalize(input); got != want {
			t.Errorf("normalize(%q) = %q, want %q", input, got, want)
		}
	}
}

// TestHostHeader 默认端口不拼端口，非默认端口必须拼。
func TestHostHeader(t *testing.T) {
	cases := map[string]string{
		"https://example.com/p":      "example.com",
		"https://example.com:443/p":  "example.com",
		"http://example.com:80/p":    "example.com",
		"https://example.com:8443/p": "example.com:8443",
		"http://example.com:8080/p":  "example.com:8080",
	}
	for rawURL, want := range cases {
		req, err := http.NewRequest(http.MethodPost, rawURL, nil)
		if err != nil {
			t.Fatalf("build request for %s failed: %v", rawURL, err)
		}
		if got := hostHeader(req.URL); got != want {
			t.Errorf("hostHeader(%s) = %q, want %q", rawURL, got, want)
		}
	}
}
