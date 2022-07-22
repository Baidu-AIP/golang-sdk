package main

import (
	"fmt"

	"github.com/Baidu-AIP/golang-sdk/aip/imagesearch"

	"github.com/Baidu-AIP/golang-sdk/util"

	"github.com/Baidu-AIP/golang-sdk/test/resources"
)

var image = util.ReadFileToBase64("test/resources/image/baidu_image.png")
var url = "https://baidu.aip.test.com"
var brief = "{\"name\":\"百度\", \"id\":\"123\"}"
var options = make(map[string]interface{})
var contSign = "123"

// SimilarAddTest 相似图片搜索 - 入库
func SimilarAddTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SimilarAdd(image, brief, options))
}

// SimilarAddUrlTest 相似图片搜索 - 入库
func SimilarAddUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SimilarAddUrl(url, brief, options))
}

// SimilarSearchTest 相似图片搜索 - 检索
func SimilarSearchTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SimilarSearch(image, options))
}

// SimilarSearchUrlTest 相似图片搜索 - 检索
func SimilarSearchUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SimilarSearchUrl(url, options))
}

// SimilarUpdateTest 相似图片搜索 - 更新
func SimilarUpdateTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SimilarUpdate(image, options))
}

// SimilarUpdateUrlTest 相似图片搜索 - 更新
func SimilarUpdateUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SimilarUpdateUrl(url, options))
}

// SimilarUpdateSignTest 相似图片搜索 - 更新
func SimilarUpdateSignTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SimilarUpdateSign(url, options))
}

// SimilarDeleteTest 相似图片搜索 - 删除
func SimilarDeleteTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SimilarDelete(image, options))
}

// SimilarDeleteUrlTest 相似图片搜索 - 删除
func SimilarDeleteUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SimilarDeleteUrl(url, options))
}

// SimilarDeleteSignTest 相似图片搜索 - 删除
func SimilarDeleteSignTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SimilarDeleteSign(contSign, options))
}

// *****************  相同图片搜索  ********************

// SameHqAddTest 相同图片搜索 - 入库
func SameHqAddTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SameHqAdd(image, brief, options))
}

// SameHqAddUrlTest 相同图片搜索 - 入库
func SameHqAddUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SimilarAddUrl(url, brief, options))
}

// SameHqSearchTest 相同图片搜索 - 检索
func SameHqSearchTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SameHqSearch(image, options))
}

// SameHqSearchUrlTest 相同图片搜索 - 检索
func SameHqSearchUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SameHqSearchUrl(url, options))
}

// SameHqUpdateTest 相同图片搜索 - 更新
func SameHqUpdateTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SameHqUpdate(image, options))
}

// SameHqUpdateUrlTest 相同图片搜索 - 更新
func SameHqUpdateUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SameHqUpdateUrl(url, options))
}

// SameHqUpdateSignTest 相同图片搜索 - 更新
func SameHqUpdateSignTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SameHqUpdateSign(url, options))
}

// SameHqDeleteTest 相同图片搜索 - 删除
func SameHqDeleteTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SameHqDelete(image, options))
}

// SameHqDeleteUrlTest 相同图片搜索 - 删除
func SameHqDeleteUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SimilarDeleteUrl(url, options))
}

// SameHqDeleteSignTest 相同图片搜索 - 删除
func SameHqDeleteSignTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.SameHqDeleteSign(contSign, options))
}

// *****************  商品图片搜索  ********************

// ProductAddTest 商品图片搜索 - 入库
func ProductAddTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.ProductAdd(image, brief, options))
}

// ProductAddUrlTest 商品图片搜索 - 入库
func ProductAddUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.ProductAddUrl(url, brief, options))
}

// ProductSearchTest 商品图片搜索 - 检索
func ProductSearchTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.ProductSearch(image, options))
}

// ProductSearchUrlTest 商品图片搜索 - 检索
func ProductSearchUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.ProductSearchUrl(url, options))
}

// ProductUpdateTest 商品图片搜索 - 更新
func ProductUpdateTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.ProductUpdate(image, options))
}

// ProductUpdateUrlTest 商品图片搜索 - 更新
func ProductUpdateUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.ProductUpdateUrl(url, options))
}

// ProductUpdateSignTest 商品图片搜索 - 更新
func ProductUpdateSignTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.ProductUpdateSign(url, options))
}

// ProductDeleteTest 商品图片搜索 - 删除
func ProductDeleteTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.ProductDelete(image, options))
}

// ProductDeleteUrlTest 商品图片搜索 - 删除
func ProductDeleteUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.ProductDeleteUrl(url, options))
}

// ProductDeleteSignTest 商品图片搜索 - 删除
func ProductDeleteSignTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.ProductDeleteSign(contSign, options))
}


// *****************  绘本图片搜索  ********************

// PictureBookAddTest 绘本图片搜索 - 入库
func PictureBookAddTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.PictureBookAdd(image, brief, options))
}

// PictureBookAddUrlTest 绘本图片搜索 - 入库
func PictureBookAddUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.PictureBookAddUrl(url, brief, options))
}

// PictureBookSearchTest 绘本图片搜索 - 检索
func PictureBookSearchTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.PictureBookSearch(image, options))
}

// PictureBookSearchUrlTest 绘本图片搜索 - 检索
func PictureBookSearchUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.PictureBookSearchUrl(url, options))
}

// PictureBookUpdateTest 绘本图片搜索 - 更新
func PictureBookUpdateTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.PictureBookUpdate(image, options))
}

// PictureBookUpdateUrlTest 绘本图片搜索 - 更新
func PictureBookUpdateUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.PictureBookUpdateUrl(url, options))
}

// PictureBookUpdateSignTest 绘本图片搜索 - 更新
func PictureBookUpdateSignTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.PictureBookUpdateSign(contSign, options))
}

// PictureBookDeleteTest 绘本图片搜索 - 删除
func PictureBookDeleteTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.PictureBookDelete(image, options))
}

// PictureBookDeleteUrlTest 绘本图片搜索 - 删除
func PictureBookDeleteUrlTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.PictureBookDeleteUrl(url, options))
}

// PictureBookDeleteSignTest 绘本图片搜索 - 删除
func PictureBookDeleteSignTest(client *imagesearch.ImageSearchClient) {
	fmt.Println(client.PictureBookDeleteSign(contSign, options))
}


func main() {

	var client = imagesearch.NewClient(resources.Ak, resources.SK)

	SimilarAddTest(client)
	SimilarAddUrlTest(client)
	SimilarSearchTest(client)
	SimilarSearchUrlTest(client)
	SimilarUpdateTest(client)
	SimilarUpdateUrlTest(client)
	SimilarUpdateSignTest(client)
	SimilarDeleteTest(client)
	SimilarDeleteUrlTest(client)
	SimilarDeleteSignTest(client)


	SameHqAddTest(client)
	SameHqAddUrlTest(client)
	SameHqSearchTest(client)
	SameHqSearchUrlTest(client)
	SameHqUpdateTest(client)
	SameHqUpdateUrlTest(client)
	SameHqUpdateSignTest(client)
	SameHqDeleteTest(client)
	SameHqDeleteUrlTest(client)
	SameHqDeleteSignTest(client)



	ProductAddTest(client)
	ProductAddUrlTest(client)
	ProductSearchTest(client)
	ProductSearchUrlTest(client)
	ProductUpdateTest(client)
	ProductUpdateUrlTest(client)
	ProductUpdateSignTest(client)
	ProductDeleteTest(client)
	ProductDeleteUrlTest(client)
	ProductDeleteSignTest(client)


	PictureBookAddTest(client)
	PictureBookAddUrlTest(client)
	PictureBookSearchTest(client)
	PictureBookSearchUrlTest(client)
	PictureBookUpdateTest(client)
	PictureBookUpdateUrlTest(client)
	PictureBookUpdateSignTest(client)
	PictureBookDeleteTest(client)
	PictureBookDeleteUrlTest(client)
	PictureBookDeleteSignTest(client)

}

