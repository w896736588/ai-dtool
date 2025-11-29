package test

import (
	"dev_tool/internal/pkg/p_api"
	"fmt"
	"sync"
	"testing"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

var wg sync.WaitGroup

// TestFpm 测试fpm无session的情况
func TestFpmNoSession(t *testing.T) {

	fmt.Println("Curl命令解析器")
	fmt.Println("=================")

	//linux cmd bash
	//example :=
	//	`curl 'http://dev1.zhikefu.com.cn/UploadManager/upload' \
	//-H 'Accept: application/json, text/plain, */*' \
	//-H 'Accept-Language: zh-CN,zh;q=0.9' \
	//-H 'Connection: keep-alive' \
	//-H 'Content-Type: multipart/form-data; boundary=----WebKitFormBoundaryIX9Aobw5yhqGUBs0' \
	//-b 'PHPSID=637f27df6e0fda4107823642d35342b9' \
	//-H 'Origin: http://dev1admin.zhikefu.com.cn' \
	//-H 'Referer: http://dev1admin.zhikefu.com.cn/' \
	//-H 'User-Agent: Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/142.0.0.0 Safari/537.36' \
	//-H 'token: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE3NjQzMzU4ODcsIm5iZiI6MTc2NDMzNTg4NywiZXhwIjoxNzY0OTQwNjg3LCJkYXRhIjp7ImlkIjoxLCJ1c2VybmFtZSI6IjE4NTAwMDAwMDAxIiwicGFzc3dvcmQiOiI5MGJjOTA5MzgwYTRjZjQyYmZkODA5ZWU3NGNiZGIzOSIsInVzZXJfdHlwZSI6MSwiYWRtaW5faWQiOjAsImNyZWF0ZV90aW1lIjoxNzYxMzYzMzAyLCJ1cGRhdGVfdGltZSI6MTc2MTM2MzMwMn19.1eTuiVpt-53gxG4hOjZzjMXlPY3q5Cimup2URKqikdg' \
	//--data-raw $'------WebKitFormBoundaryIX9Aobw5yhqGUBs0\r\nContent-Disposition: form-data; name="file"; filename="u=2990600787,3256164520&fm=253&gp=0.jpg"\r\nContent-Type: image/jpeg\r\n\r\n\r\n------WebKitFormBoundaryIX9Aobw5yhqGUBs0\r\nContent-Disposition: form-data; name="file_type"\r\n\r\nimage\r\n------WebKitFormBoundaryIX9Aobw5yhqGUBs0\r\nContent-Disposition: form-data; name="business"\r\n\r\nuser\r\n------WebKitFormBoundaryIX9Aobw5yhqGUBs0--\r\n' \
	//--insecure`

	//linux curl cmd
	example1 := `curl --location --request POST 'http://dev1.zhikefu.com.cnbaidu.com?ual=xxx' \
--header 'Content-Type: application/json' \
--data-raw '{"aa" : 1}'`
	//	gstool.FmtPrintlnLog(`%s`, example1)
	parse := p_api.NewParseCurl(example1)
	err := parse.ParseCurl()
	if err != nil {
		fmt.Println("解析错误:", err)
		return
	}
	gstool.FmtPrintlnLogTime(`%s`, gstool.JsonFormat(parse.CurlStruct))
}
