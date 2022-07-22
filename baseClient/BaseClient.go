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

package baseClient

import (
	"bytes"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"sort"
	"strings"
	"time"
)

const authUrl = "https://aip.baidubce.com/oauth/2.0/token?grant_type=client_credentials&client_id=%s&client_secret=%s"

type Auth struct {
	accessToken string
	times       int64
	hasAuth     bool
	client      http.Client
	error       string
	isCloudUser bool
	ak          string
	sk          string
	i           int8
}

func (auth *Auth) InitAuth(ak string, sk string) {
	auth.ak = ak
	auth.sk = sk
	auth.client = http.Client{
		Timeout: time.Second * 20,
	}
	auth.refresh()
}

func (auth *Auth) InitCloudAuth(ak string, sk string) {
	auth.ak = ak
	auth.sk = sk
	auth.client = http.Client{
		Timeout: time.Second * 20,	
	}
	auth.isCloudUser = true
}
func (auth *Auth) refresh() {
	log.Println("get access_token")
	now := time.Now().Unix()
	url := fmt.Sprintf(authUrl, auth.ak, auth.sk)
	resp, err := auth.client.Get(url)
	if err != nil || resp == nil {
		auth.hasAuth = false
		log.Println(err)
		return
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, _ := ioutil.ReadAll(resp.Body)
	var jsonMap map[string]interface{}
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		log.Println("%s is not json", string(body))
		return
	}
	accessToken, ok := jsonMap["access_token"]
	if !ok {
		auth.hasAuth = false
		log.Println(string(body))
		return
	}
	auth.accessToken = accessToken.(string)
	expires := int64(jsonMap["expires_in"].(float64))
	auth.times = expires + now - 3600
	auth.hasAuth = true
	auth.error = ""
}

func hmacSha256(data string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(data))
	return hex.EncodeToString(h.Sum(nil))
}

func (auth *Auth) setHeader(req *http.Request) {
	now := time.Now().UTC().Format("2006-01-02T15:04:05Z")
	req.Header.Set("Host", req.Host)
	req.Header.Set("x-bce-date", now)
	signedHeaders := "host;x-bce-date"
	heades := []string{}
	heades = append(heades, "host:"+url.QueryEscape(req.Host))
	heades = append(heades, "x-bce-date:"+url.QueryEscape(now))
	sort.Strings(heades)
	canonicalHeaders := strings.Join(heades, "\n")
	expire := "1800"
	val := fmt.Sprintf("bce-auth-v1/%s/%s/%s", auth.ak, now, expire)
	signingKey := hmacSha256(val, auth.sk)

	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}
	canonicalURI = url.QueryEscape(canonicalURI)
	canonicalURI = strings.ReplaceAll(canonicalURI, "%2F", "/")
	canonicalQueryString := req.URL.RawQuery
	if canonicalQueryString != "" {
		queryArray := strings.Split(canonicalQueryString, "&")
		stringList := []string{}
		for _, val := range queryArray {
			kvArray := strings.Split(val, "=")
			kv := ""
			key := kvArray[0]
			kv += url.QueryEscape(key) + "="
			if len(kvArray) >= 2 {
				value := kvArray[1]
				kv += url.QueryEscape(value)
			}
			stringList = append(stringList, kv)
		}
		sort.Strings(stringList)
		canonicalQueryString = strings.Join(stringList, "&")
	}
	canonicalRequest := req.Method + "\n" + canonicalURI + "\n" + canonicalQueryString + "\n" + canonicalHeaders
	signature := hmacSha256(canonicalRequest, signingKey)
	authInfo := fmt.Sprintf("bce-auth-v1/%s/%s/%s/%s/%s", auth.ak, now, expire, signedHeaders, signature)
	req.Header.Set("Authorization", authInfo)
}

func (auth *Auth) setParam(url string) string {
	var buffer bytes.Buffer
	buffer.WriteString(url)
	if strings.Contains(url, "?") {
		buffer.WriteString("&aipSdk=golang")
	} else {
		buffer.WriteString("?aipSdk=golang")
	}
	if auth.accessToken != "" {
		buffer.WriteString("&access_token=")
		buffer.WriteString(auth.accessToken)
	}
	return buffer.String()
}

func PostJson(urlString string, data string, auth *Auth) string {
	now := time.Now().Unix()
	if auth.hasAuth && auth.times < now && !auth.isCloudUser {
		auth.refresh()
	}
	urlString = auth.setParam(urlString)
	req, err := http.NewRequest("POST", urlString, bytes.NewBuffer([]byte(data)))
	req.Header.Set("Content-Type", "application/json")
	auth.setHeader(req)
	client := auth.client
	resp, err := client.Do(req)
	if err != nil {
		return err.Error()
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, _ := ioutil.ReadAll(resp.Body)
	return string(body)
}

func PostUrlForm(urlString string, data map[string]string, auth *Auth) string {
	if !auth.isCloudUser {
		now := time.Now().Unix()
		if !auth.hasAuth || (auth.hasAuth && auth.times < now) {
			auth.refresh()
		}
	}
	urlString = auth.setParam(urlString)
	DataUrlVal := url.Values{}
	for key, val := range data {
		DataUrlVal.Add(key, val)
	}
	req, err := http.NewRequest("POST", urlString, strings.NewReader(DataUrlVal.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if auth.isCloudUser {
		auth.setHeader(req)
	}
	client := auth.client
	resp, err := client.Do(req)
	if err != nil {
		return err.Error()
	}
	defer func(Body io.ReadCloser) {
		_ = Body.Close()
	}(resp.Body)
	body, _ := ioutil.ReadAll(resp.Body)
	var jsonMap map[string]interface{}
	err = json.Unmarshal(body, &jsonMap)
	if err != nil {
		log.Println("ger &s response %s is not json", urlString, string(body))
	} else {
		errorCode, ok := jsonMap["error_code"]
		if ok && int(errorCode.(float64)) == 14 {
			auth.isCloudUser = false
		}
	}
	return string(body)
}
