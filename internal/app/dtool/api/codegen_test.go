package api

import (
	"strings"
	"testing"

	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_curl"
)

// newCodegenTestAPI 中文：构造一个最小可用的代码生成测试对象。 English: Build a minimal API instance for snippet tests.
func newCodegenTestAPI() *Api {
	return &Api{
		BaseInfo: &BaseInfo{
			EnvItems: map[string]string{
				"$host$": "https://example.com",
			},
		},
		CurlStruct: p_curl.CurlStruct{
			Method: httpMethodPost,
			Url:    "$host$/v1/orders",
			Headers: map[string]string{
				"Authorization": "Bearer $token$",
				"Content-Type":  define.ContentTypeJson,
			},
			ContentType: define.ContentTypeJson,
			BodyJson:    `{"product_id":123,"quantity":1}`,
		},
	}
}

const (
	// httpMethodPost 中文：测试里复用的 POST 常量。 English: Shared POST constant for tests.
	httpMethodPost = "POST"
)

func TestGenerateCodeSupportsLegacyTypes(t *testing.T) {
	apiItem := newCodegenTestAPI()
	apiItem.BaseInfo.EnvItems["$token$"] = "test-token"

	chromeCode := apiItem.GenerateCode(CodeTypeCurlChrome)
	if !strings.Contains(chromeCode, "curl --request POST 'https://example.com/v1/orders'") {
		t.Fatalf("chrome code missing request line: %s", chromeCode)
	}
	if !strings.Contains(chromeCode, "--data-raw '{\"product_id\":123,\"quantity\":1}'") {
		t.Fatalf("chrome code missing body: %s", chromeCode)
	}

	apifoxCode := apiItem.GenerateCode(CodeTypeCurlApifox)
	if !strings.Contains(apifoxCode, "curl --location --request POST") {
		t.Fatalf("apifox code missing --location: %s", apifoxCode)
	}
}

func TestGenerateCodeSupportsMoreCommonTypes(t *testing.T) {
	apiItem := newCodegenTestAPI()
	apiItem.BaseInfo.EnvItems["$token$"] = "test-token"

	fetchCode := apiItem.GenerateCode(CodeTypeFetch)
	if !strings.Contains(fetchCode, "fetch(url, {") || !strings.Contains(fetchCode, "const body = `") {
		t.Fatalf("fetch code not generated correctly: %s", fetchCode)
	}

	axiosCode := apiItem.GenerateCode(CodeTypeAxios)
	if !strings.Contains(axiosCode, "import axios from 'axios';") || !strings.Contains(axiosCode, "axios.request(config)") {
		t.Fatalf("axios code not generated correctly: %s", axiosCode)
	}

	pythonCode := apiItem.GenerateCode(CodeTypePython)
	if !strings.Contains(pythonCode, "import requests") || !strings.Contains(pythonCode, "requests.post(url") {
		t.Fatalf("python code not generated correctly: %s", pythonCode)
	}

	phpCode := apiItem.GenerateCode(CodeTypePHP)
	if !strings.Contains(phpCode, "curl_setopt_array($curl, [") || !strings.Contains(phpCode, "CURLOPT_POSTFIELDS") {
		t.Fatalf("php code not generated correctly: %s", phpCode)
	}

	goCode := apiItem.GenerateCode(CodeTypeGolang)
	if !strings.Contains(goCode, `http.NewRequest("POST", "https://example.com/v1/orders", bodyReader)`) || !strings.Contains(goCode, `req.Header.Set("Authorization", "Bearer test-token")`) {
		t.Fatalf("golang code not generated correctly: %s", goCode)
	}

	postmanCode := apiItem.GenerateCode(CodeTypePostman)
	if !strings.Contains(postmanCode, `"schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"`) || !strings.Contains(postmanCode, `"method": "POST"`) {
		t.Fatalf("postman code not generated correctly: %s", postmanCode)
	}
}

func TestGenerateCodeSupportsFormAndMultipart(t *testing.T) {
	formAPI := &Api{
		BaseInfo: &BaseInfo{},
		CurlStruct: p_curl.CurlStruct{
			Method:      httpMethodPost,
			Url:         "https://example.com/v1/login",
			ContentType: define.ContentTypeForm,
			BodyForm: []p_curl.KeyValue{
				{Field: "username", Type: "string", Value: "frog"},
				{Field: "remember", Type: "boolean", Value: "true"},
			},
		},
	}
	formCode := formAPI.GenerateCode(CodeTypeCurlChrome)
	if !strings.Contains(formCode, "--data-urlencode 'username=frog'") || !strings.Contains(formCode, "--data-urlencode 'remember=true'") {
		t.Fatalf("form curl code not generated correctly: %s", formCode)
	}

	multipartAPI := &Api{
		BaseInfo: &BaseInfo{},
		CurlStruct: p_curl.CurlStruct{
			Method:      httpMethodPost,
			Url:         "https://example.com/v1/upload",
			ContentType: define.ContentTypeMultiForm,
			BodyForm: []p_curl.KeyValue{
				{Field: "file", Type: "file", Value: "/tmp/demo.txt"},
				{Field: "scene", Type: "string", Value: "avatar"},
			},
		},
	}
	multipartCode := multipartAPI.GenerateCode(CodeTypeFetch)
	if !strings.Contains(multipartCode, "new FormData()") || !strings.Contains(multipartCode, "new Blob()") {
		t.Fatalf("multipart fetch code not generated correctly: %s", multipartCode)
	}
}
