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
	"strconv"

	"github.com/Baidu-AIP/golang-sdk/baseClient"
)

const __imageCensorUserDefinedUrl = "https://aip.baidubce.com/rest/2.0/solution/v1/img_censor/v2/user_defined"

const __textCensorUserDefinedUrl = "https://aip.baidubce.com/rest/2.0/solution/v1/text_censor/v2/user_defined"

const __voiceCensorUserDefinedUrl = "https://aip.baidubce.com/rest/2.0/solution/v1/voice_censor/v3/user_defined"

const __videoCensorUserDefinedUrl = "https://aip.baidubce.com/rest/2.0/solution/v1/video_censor/v2/user_defined"

type ContentCensorClient struct {
	auth baseClient.Auth
}

func NewClient(ak string, sk string) *ContentCensorClient {
	var client ContentCensorClient
	client.auth = baseClient.Auth{}
	client.auth.InitAuth(ak, sk)
	return &client
}

func NewCloudClient(ak string, sk string) *ContentCensorClient {
	var client ContentCensorClient
	client.auth = baseClient.Auth{}
	client.auth.InitCloudAuth(ak, sk)
	return &client
}

func (client *ContentCensorClient) TextCensor(text string) (result string) {
	data := make(map[string]string)
	data["text"] = text
	return baseClient.PostUrlForm(__textCensorUserDefinedUrl, data, &client.auth)
}

func (client *ContentCensorClient) ImgCensor(image string, options map[string]interface{}) (result string) {
	data := make(map[string]string)
	data["image"] = image
	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}
	return baseClient.PostUrlForm(__imageCensorUserDefinedUrl, data, &client.auth)
}

func (client *ContentCensorClient) ImgCensorUrl(imgUrl string, options map[string]interface{}) (result string) {
	data := make(map[string]string)
	data["imgUrl"] = imgUrl
	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}
	return baseClient.PostUrlForm(__imageCensorUserDefinedUrl, data, &client.auth)
}

func (client *ContentCensorClient) VoiceCensorUrl(url string, rate int, fmt string, options map[string]interface{}) (result string) {
	data := make(map[string]string)
	data["url"] = url
	data["fmt"] = fmt
	data["rate"] = strconv.FormatInt(int64(rate), 10)
	for key, val := range options {
		switch val := val.(type) {
		case bool:
			data[key] = strconv.FormatBool(val)
		}
	}
	return baseClient.PostUrlForm(__voiceCensorUserDefinedUrl, data, &client.auth)
}

func (client *ContentCensorClient) VoiceCensor(base64 string, rate int, fmt string, options map[string]interface{}) (result string) {
	data := make(map[string]string)
	data["base64"] = base64
	data["fmt"] = fmt
	data["rate"] = strconv.FormatInt(int64(rate), 10)
	for key, val := range options {
		switch val := val.(type) {
		case bool:
			data[key] = strconv.FormatBool(val)
		}
	}
	return baseClient.PostUrlForm(__voiceCensorUserDefinedUrl, data, &client.auth)
}

func (client *ContentCensorClient) VideoCensor(name string, videoUrl string, extId string, options map[string]interface{}) (result string) {
	data := make(map[string]string)
	data["name"] = name
	data["videoUrl"] = videoUrl
	data["extId"] = extId
	for key, val := range options {
		switch val := val.(type) {
		case string:
			data[key] = val
		}
	}
	return baseClient.PostUrlForm(__videoCensorUserDefinedUrl, data, &client.auth)
}
