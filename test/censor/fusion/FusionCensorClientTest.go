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

// 融合审核（文本 / 图像 / 长视频）SDK 使用示例。
//
// SDK 返回的是响应报文原文（JSON 字符串），本示例演示如何按需解析：
// 只取轮询需要的 status / error_code，其余字段原样打印。
//
// 运行：
//
//	go run test/censor/fusion/FusionCensorClientTest.go \
//	  -ak <your-ak> -sk <your-sk> -mode text -content "待审核文本"
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/Baidu-AIP/golang-sdk/aip/censor"
)

var (
	ak         = flag.String("ak", "", "百度智能云 access key")
	sk         = flag.String("sk", "", "百度智能云 secret key")
	mode       = flag.String("mode", "text", "审核模态：text / image / video")
	strategyID = flag.Int64("strategyId", 0, "小模型策略 ID，不传则用默认策略 1")
	content    = flag.String("content", "", "文本内容 / 图片 URL / 视频 URL")
)

// __pullInterval 轮询间隔。长视频耗时明显长于文本与图像，可按需放大。
const __pullInterval = 3 * time.Second

// __maxPullTimes 最多轮询次数，防止无限等待。
const __maxPullTimes = 40

// censorResponse 只建模轮询流程必需的字段。
//
// 审核明细（data）字段集不封闭，且各模态结构不同（文本 / 图像是数组，
// 长视频是对象），所以这里不建模，由调用方按自己的业务需要解析。
type censorResponse struct {
	TaskID    string `json:"taskId"`
	LogID     string `json:"log_id"`
	Status    string `json:"status"`
	ErrorCode int64  `json:"error_code"`
	ErrorMsg  string `json:"error_msg"`

	// 网关（API Gateway）在请求到达审核服务前拒绝时，返回的是 BCE 通用错误格式，
	// 与业务错误的 error_code 形状不同，两者都要检查
	GatewayCode    string `json:"code"`
	GatewayMessage string `json:"message"`
	RequestID      string `json:"requestId"`
}

// isSuccess 业务错误码与网关错误码都为空才算成功。
func (response *censorResponse) isSuccess() bool {
	return response.ErrorCode == 0 && response.GatewayCode == ""
}

// isTerminal 任务是否已到终态。非终态时报文里没有结论字段。
func (response *censorResponse) isTerminal() bool {
	return response.Status == censor.FusionStatusFinished ||
		response.Status == censor.FusionStatusError
}

// errorDescription 统一业务错误与网关错误两种报文形状。
func (response *censorResponse) errorDescription() string {
	if response.GatewayCode != "" {
		return fmt.Sprintf("gateway %s: %s (requestId=%s)",
			response.GatewayCode, response.GatewayMessage, response.RequestID)
	}
	return fmt.Sprintf("error_code=%d error_msg=%s (log_id=%s)",
		response.ErrorCode, response.ErrorMsg, response.LogID)
}

func main() {
	flag.Parse()
	if *ak == "" || *sk == "" || *content == "" {
		flag.Usage()
		log.Fatal("ak / sk / content 均为必填")
	}

	// 用 ak/sk 初始化客户端：IAM 签名在每次请求内部现算，不需要换取 access_token。
	// 当前服务只部署了华北地域，无需选择地域。
	client, err := censor.NewFusionClient(*ak, *sk)
	if err != nil {
		log.Fatalf("初始化客户端失败: %v", err)
	}
	fmt.Printf("服务域名 %s\n", client.Endpoint())

	var submitRaw string
	var pull func(taskID string) (string, error)

	switch *mode {
	case "text":
		submitRaw, err = client.SubmitText(&censor.SubmitTextRequest{
			Text:       *content,
			StrategyID: *strategyID,
			UserID:     "sdk_demo",
		})
		pull = client.PullTextResult
	case "image":
		submitRaw, err = client.SubmitImage(&censor.SubmitImageRequest{
			ImgURL:     *content,
			StrategyID: *strategyID,
			UserID:     "sdk_demo",
		})
		pull = client.PullImageResult
	case "video":
		detectType := censor.VideoDetectTypeFrameAndAudio
		submitRaw, err = client.SubmitVideo(&censor.SubmitVideoRequest{
			URL:        *content,
			DetectType: &detectType,
			StrategyID: *strategyID,
			UserID:     "sdk_demo",
		})
		pull = client.PullVideoResult
	default:
		log.Fatalf("未知模态 %q，可选 text / image / video", *mode)
	}

	// err 只表示传输失败或本地参数校验失败；服务端拒绝（含 4xx/5xx）会正常返回报文
	if err != nil {
		log.Fatalf("提交失败: %v", err)
	}
	fmt.Printf("提交响应: %s\n", submitRaw)

	var submitResponse censorResponse
	if err := json.Unmarshal([]byte(submitRaw), &submitResponse); err != nil {
		log.Fatalf("提交响应不是 JSON: %v", err)
	}
	if !submitResponse.isSuccess() {
		log.Fatalf("提交被拒绝: %s", submitResponse.errorDescription())
	}
	fmt.Printf("提交成功，taskId=%s\n", submitResponse.TaskID)

	// submit 是异步的，只返回 taskId；结论要轮询 result 或配置 callbackUrl 走回调。
	resultRaw := pollResult(pull, submitResponse.TaskID)
	if resultRaw == "" {
		return
	}
	fmt.Printf("\n审核完成，完整报文：\n%s\n", prettyJSON(resultRaw))
}

// pollResult 轮询到终态，返回终态报文原文。未完成时服务端返回 PROCESSING
// 而非报错，不能当失败处理。
func pollResult(pull func(taskID string) (string, error), taskID string) string {
	for i := 0; i < __maxPullTimes; i++ {
		time.Sleep(__pullInterval)

		raw, err := pull(taskID)
		if err != nil {
			log.Printf("第 %d 次拉取失败，继续重试: %v", i+1, err)
			continue
		}

		var response censorResponse
		if err := json.Unmarshal([]byte(raw), &response); err != nil {
			log.Printf("第 %d 次拉取的响应不是 JSON: %v，原始报文: %s", i+1, err, raw)
			continue
		}
		// 业务错误（无权限 6、任务不存在 282006 等）不会重试成功，直接退出。
		if !response.isSuccess() {
			log.Printf("拉取被拒绝: %s", response.errorDescription())
			return ""
		}
		if !response.isTerminal() {
			fmt.Printf("第 %d 次拉取：status=%s，继续等待\n", i+1, response.Status)
			continue
		}
		return raw
	}
	log.Printf("轮询 %d 次仍未完成，taskId=%s", __maxPullTimes, taskID)
	return ""
}

// prettyJSON 格式化报文便于阅读，解析失败则原样返回。
func prettyJSON(raw string) string {
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(raw), "", "    "); err != nil {
		return raw
	}
	return buf.String()
}
