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
// 运行：
//
//	go run test/censor/fusion/FusionCensorClientTest.go \
//	  -ak <your-ak> -sk <your-sk> -mode text -content "待审核文本"
package main

import (
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

	var submitResponse *censor.FusionResponse
	var pull func(taskID string) (*censor.FusionResponse, error)

	switch *mode {
	case "text":
		submitResponse, err = client.SubmitText(&censor.SubmitTextRequest{
			Text:       *content,
			StrategyID: *strategyID,
			UserID:     "sdk_demo",
		})
		pull = client.PullTextResult
	case "image":
		submitResponse, err = client.SubmitImage(&censor.SubmitImageRequest{
			ImgURL:     *content,
			StrategyID: *strategyID,
			UserID:     "sdk_demo",
		})
		pull = client.PullImageResult
	case "video":
		detectType := censor.VideoDetectTypeFrameAndAudio
		submitResponse, err = client.SubmitVideo(&censor.SubmitVideoRequest{
			URL:        *content,
			DetectType: &detectType,
			StrategyID: *strategyID,
			UserID:     "sdk_demo",
		})
		pull = client.PullVideoResult
	default:
		log.Fatalf("未知模态 %q，可选 text / image / video", *mode)
	}

	if err != nil {
		log.Fatalf("提交失败: %v", err)
	}
	if !submitResponse.IsSuccess() {
		log.Fatalf("提交被拒绝: %s", submitResponse.ErrorDescription())
	}
	fmt.Printf("提交成功，taskId=%s\n", submitResponse.TaskID)

	// submit 是异步的，只返回 taskId；结论要轮询 result 或配置 callbackUrl 等回调。
	result := pollResult(pull, submitResponse.TaskID)
	if result == nil {
		return
	}
	printResult(*mode, result)
}

// pollResult 轮询到终态。未完成时服务端返回 PROCESSING 而非报错，不能当失败处理。
func pollResult(pull func(taskID string) (*censor.FusionResponse, error),
	taskID string) *censor.FusionResponse {
	for i := 0; i < __maxPullTimes; i++ {
		time.Sleep(__pullInterval)

		response, err := pull(taskID)
		if err != nil {
			log.Printf("第 %d 次拉取失败，继续重试: %v", i+1, err)
			continue
		}
		// 业务错误（无权限 6、任务不存在 282006 等）不会重试成功，直接退出。
		if !response.IsSuccess() && !response.IsTerminal() {
			log.Printf("拉取被拒绝: %s", response.ErrorDescription())
			return nil
		}
		if !response.IsTerminal() {
			fmt.Printf("第 %d 次拉取：status=%s，继续等待\n", i+1, response.Status)
			continue
		}
		return response
	}
	log.Printf("轮询 %d 次仍未完成，taskId=%s", __maxPullTimes, taskID)
	return nil
}

func printResult(mode string, response *censor.FusionResponse) {
	fmt.Printf("\n审核完成：status=%s conclusion=%s conclusionType=%d\n",
		response.Status, response.Conclusion, response.ConclusionType)
	if !response.IsSuccess() {
		fmt.Printf("错误信息：%s\n", response.ErrorDescription())
	}

	if mode == "video" {
		printVideoData(response)
		return
	}

	details, err := response.UnmarshalAuditDetails()
	if err != nil {
		log.Printf("解析明细失败: %v，原始报文: %s", err, response.Raw)
		return
	}
	if len(details) == 0 {
		fmt.Println("无违规明细（合规且未命中白名单）")
		return
	}
	for i, detail := range details {
		fmt.Printf("明细 %d: type=%d subType=%d conclusionType=%d msg=%s agentType=%s agentSubType=%s\n",
			i+1, detail.Type, detail.SubType, detail.ConclusionType,
			detail.Msg, detail.AgentType, detail.AgentSubType)
		// 未建模字段（probability、location、hits、words 等）原样保留在 Extras
		for key, value := range detail.Extras {
			fmt.Printf("    %s = %s\n", key, string(value))
		}
	}
}

func printVideoData(response *censor.FusionResponse) {
	data, err := response.UnmarshalVideoData()
	if err != nil {
		log.Printf("解析长视频结果失败: %v，原始报文: %s", err, response.Raw)
		return
	}
	if data == nil {
		fmt.Println("无违规明细（合规且未命中白名单）")
		return
	}
	fmt.Printf("视频时长 %d 秒，违规帧 %d 个，违规音频片段 %d 个\n",
		data.TaskDuration, len(data.Frames), len(data.Audios))
	for _, frame := range data.Frames {
		fmt.Printf("帧 %dms: %s，明细 %d 条\n",
			frame.FrameTimeStamp, frame.FrameURL, len(frame.Data))
	}
	for _, audio := range data.Audios {
		fmt.Printf("音频 %dms-%dms: %s，识别文本=%v，明细 %d 条\n",
			audio.StartTime, audio.EndTime, audio.AudioURL, audio.RawText, len(audio.Data))
	}
}
