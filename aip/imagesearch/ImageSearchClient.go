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

package imagesearch

import (

	"strconv"

	"github.com/Baidu-AIP/golang-sdk/baseClient"
)

const __imageSearchSameHqAdd = "https://aip.baidubce.com/rest/2.0/realtime_search/same_hq/add"

const __imageSearchSameHqSearch = "https://aip.baidubce.com/rest/2.0/realtime_search/same_hq/search"

const __imageSearchSameHqUpdate = "https://aip.baidubce.com/rest/2.0/realtime_search/same_hq/update"

const __imageSearchSameHqDelete = "https://aip.baidubce.com/rest/2.0/realtime_search/same_hq/delete"

const __imageSearchSimilarAdd = "https://aip.baidubce.com/rest/2.0/image-classify/v1/realtime_search/similar/add"

const __imageSearchSimilarSearch = "https://aip.baidubce.com/rest/2.0/image-classify/v1/realtime_search/similar/search"

const __imageSearchSimilarUpdate = "https://aip.baidubce.com/rest/2.0/image-classify/v1/realtime_search/similar/update"

const __imageSearchSimilarDelete = "https://aip.baidubce.com/rest/2.0/image-classify/v1/realtime_search/similar/delete"

const __imageSearchProductAdd = "https://aip.baidubce.com/rest/2.0/image-classify/v1/realtime_search/product/add"

const __imageSearchProductSearch = "https://aip.baidubce.com/rest/2.0/image-classify/v1/realtime_search/product/search"

const __imageSearchProductUpdate = "https://aip.baidubce.com/rest/2.0/image-classify/v1/realtime_search/product/update"

const __imageSearchProductDelete = "https://aip.baidubce.com/rest/2.0/image-classify/v1/realtime_search/product/delete"

const __imageSearchPictureBookAdd = "https://aip.baidubce.com/rest/2.0/imagesearch/v1/realtime_search/picturebook/add"

const __imageSearchPictureBookSearch = "https://aip.baidubce.com/rest/2.0/imagesearch/v1/realtime_search/picturebook/search"

const __imageSearchPictureBookDelete = "https://aip.baidubce.com/rest/2.0/imagesearch/v1/realtime_search/picturebook/delete"

const __imageSearchPictureBookUpdate = "https://aip.baidubce.com/rest/2.0/imagesearch/v1/realtime_search/picturebook/update"


type ImageSearchClient struct {
	auth baseClient.Auth
}

func NewClient(ak string, sk string) *ImageSearchClient {
	var client ImageSearchClient
	client.auth = baseClient.Auth{}
	client.auth.InitAuth(ak, sk)
	return &client
}

func NewCloudClient(ak string, sk string) *ImageSearchClient {
	var client ImageSearchClient
	client.auth = baseClient.Auth{}
	client.auth.InitCloudAuth(ak, sk)
	return &client
}

// SimilarAdd ?????????????????? - ??????
func (client *ImageSearchClient) SimilarAdd(image string, brief string,
	options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image
	data["brief"] = brief

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSimilarAdd, data, &client.auth)
}

// SimilarAddUrl ?????????????????? - ??????
func (client *ImageSearchClient) SimilarAddUrl(url string, brief string,
	options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url
	data["brief"] = brief

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSimilarAdd, data, &client.auth)
}

// SimilarSearch ?????????????????? - ??????
func (client *ImageSearchClient) SimilarSearch(image string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSimilarSearch, data, &client.auth)
}

// SimilarSearchUrl ?????????????????? - ??????
func (client *ImageSearchClient) SimilarSearchUrl(url string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSimilarSearch, data, &client.auth)
}

// SimilarDelete ?????????????????? - ??????
func (client *ImageSearchClient) SimilarDelete(image string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSimilarDelete, data, &client.auth)
}

// SimilarDeleteUrl ?????????????????? - ??????
func (client *ImageSearchClient) SimilarDeleteUrl(url string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSimilarDelete, data, &client.auth)
}

// SimilarDeleteSign ?????????????????? - ??????
func (client *ImageSearchClient) SimilarDeleteSign(contSign string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["cont_sign"] = contSign

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSimilarDelete, data, &client.auth)
}

// SimilarUpdate ?????????????????? - ??????
func (client *ImageSearchClient) SimilarUpdate(image string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSimilarUpdate, data, &client.auth)
}

// SimilarUpdateUrl ?????????????????? - ??????
func (client *ImageSearchClient) SimilarUpdateUrl(url string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSimilarUpdate, data, &client.auth)
}

// SimilarUpdateSign ?????????????????? - ??????
func (client *ImageSearchClient) SimilarUpdateSign(contSign string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["cont_sign"] = contSign

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSimilarUpdate, data, &client.auth)
}

// SameHqAdd ?????????????????? - ??????
func (client *ImageSearchClient) SameHqAdd(image string, brief string,
	options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image
	data["brief"] = brief

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSameHqAdd, data, &client.auth)
}

// SameHqAddUrl ?????????????????? - ??????
func (client *ImageSearchClient) SameHqAddUrl(image string, brief string,
	options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image
	data["brief"] = brief

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSameHqAdd, data, &client.auth)
}

// SameHqSearch ?????????????????? - ??????
func (client *ImageSearchClient) SameHqSearch(image string,	options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSameHqSearch, data, &client.auth)
}

// SameHqSearchUrl ?????????????????? - ??????
func (client *ImageSearchClient) SameHqSearchUrl(url string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSameHqSearch, data, &client.auth)
}

// SameHqDelete ?????????????????? - ??????
func (client *ImageSearchClient) SameHqDelete(image string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSameHqDelete, data, &client.auth)
}

// SameHqDeleteUrl ?????????????????? - ??????
func (client *ImageSearchClient) SameHqDeleteUrl(url string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSameHqDelete, data, &client.auth)
}

// SameHqDeleteSign ?????????????????? - ??????
func (client *ImageSearchClient) SameHqDeleteSign(contSign string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["cont_sign"] = contSign

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSameHqDelete, data, &client.auth)
}

// SameHqUpdate ?????????????????? - ??????
func (client *ImageSearchClient) SameHqUpdate(image string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSameHqUpdate, data, &client.auth)
}

// SameHqUpdateUrl ?????????????????? - ??????
func (client *ImageSearchClient) SameHqUpdateUrl(url string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSameHqUpdate, data, &client.auth)
}

// SameHqUpdateSign ?????????????????? - ??????
func (client *ImageSearchClient) SameHqUpdateSign(contSign string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["cont_sign"] = contSign

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchSameHqUpdate, data, &client.auth)
}

// ProductAdd ?????????????????? - ??????
func (client *ImageSearchClient) ProductAdd(image string, brief string,
	options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image
	data["brief"] = brief

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchProductAdd, data, &client.auth)
}

// ProductAddUrl ?????????????????? - ??????
func (client *ImageSearchClient) ProductAddUrl(url string, brief string,
	options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url
	data["brief"] = brief

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchProductAdd, data, &client.auth)
}

// ProductSearch ?????????????????? - ??????
func (client *ImageSearchClient) ProductSearch(image string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchProductSearch, data, &client.auth)
}

// ProductSearchUrl ?????????????????? - ??????
func (client *ImageSearchClient) ProductSearchUrl(url string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchProductSearch, data, &client.auth)
}

// ProductUpdate ?????????????????? - ??????
func (client *ImageSearchClient) ProductUpdate(image string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchProductUpdate, data, &client.auth)
}

// ProductUpdateUrl ?????????????????? - ??????
func (client *ImageSearchClient) ProductUpdateUrl(url string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchProductUpdate, data, &client.auth)
}

// ProductUpdateSign ?????????????????? - ??????
func (client *ImageSearchClient) ProductUpdateSign(contSign string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["cont_sign"] = contSign

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchProductUpdate, data, &client.auth)
}

// ProductDelete ?????????????????? - ??????
func (client *ImageSearchClient) ProductDelete(image string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchProductDelete, data, &client.auth)
}

// ProductDeleteUrl ?????????????????? - ??????
func (client *ImageSearchClient) ProductDeleteUrl(url string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchProductDelete, data, &client.auth)
}

// ProductDeleteSign ?????????????????? - ??????
func (client *ImageSearchClient) ProductDeleteSign(contSign string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["cont_sign"] = contSign

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchProductDelete, data, &client.auth)
}

// PictureBookAdd ?????????????????? - ??????
func (client *ImageSearchClient) PictureBookAdd(image string, brief string,
	options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image
	data["brief"] = brief

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchPictureBookAdd, data, &client.auth)
}

// PictureBookAddUrl ?????????????????? - ??????
func (client *ImageSearchClient) PictureBookAddUrl(url string, brief string,
	options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url
	data["brief"] = brief

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchPictureBookAdd, data, &client.auth)
}

// PictureBookSearch ?????????????????? - ??????
func (client *ImageSearchClient) PictureBookSearch(image string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchPictureBookSearch, data, &client.auth)
}

// PictureBookSearchUrl ?????????????????? - ??????
func (client *ImageSearchClient) PictureBookSearchUrl(url string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchPictureBookSearch, data, &client.auth)
}

// PictureBookDelete ?????????????????? - ??????
func (client *ImageSearchClient) PictureBookDelete(image string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchPictureBookDelete, data, &client.auth)
}

// PictureBookDeleteUrl ?????????????????? - ??????
func (client *ImageSearchClient) PictureBookDeleteUrl(url string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchPictureBookDelete, data, &client.auth)
}

// PictureBookDeleteSign ?????????????????? - ??????
func (client *ImageSearchClient) PictureBookDeleteSign(contSign string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["cont_sign"] = contSign

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchPictureBookDelete, data, &client.auth)
}

// PictureBookUpdate ?????????????????? - ??????
func (client *ImageSearchClient) PictureBookUpdate(image string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["image"] = image

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchPictureBookUpdate, data, &client.auth)
}

// PictureBookUpdateUrl ?????????????????? - ??????
func (client *ImageSearchClient) PictureBookUpdateUrl(url string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["url"] = url

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchPictureBookUpdate, data, &client.auth)
}

// PictureBookUpdateSign ?????????????????? - ??????
func (client *ImageSearchClient) PictureBookUpdateSign(contSign string, options map[string]interface{}) (result string) {
	data := make(map[string]string)

	data["cont_sign"] = contSign

	for key, val := range options {
		switch val := val.(type) {
		case int:
			data[key] = strconv.Itoa(val)
		}
	}

	return baseClient.PostUrlForm(__imageSearchPictureBookUpdate, data, &client.auth)
}



