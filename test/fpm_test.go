package test

import (
	"dev_tool/internal/pkg/p_api"
	"fmt"
	"sync"
	"testing"
)

var wg sync.WaitGroup

// TestFpm 测试fpm无session的情况
func TestFpmNoSession(t *testing.T) {

	fmt.Println("Curl命令解析器")
	fmt.Println("=================")

	// 示例curl命令
	example :=
		`curl --location --request POST 'http://dev1.zhikefu.com.cn/UploadManager/upload' \
--header 'TOKEN: eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJpYXQiOjE3NDg3NjAzMDksIm5iZiI6MTc0ODc2MDMwOSwiZXhwIjoxNzQ5MzY1MTA5LCJkYXRhIjp7ImlkIjoxLCJ1c2VybmFtZSI6InprX2FkbWluIiwicGFzc3dvcmQiOiJiNmNjMTU5MjY1ZTlkMjQ3ZWJhYWEzYTU4NWQ5MTkzNCIsInVzZXJfdHlwZSI6MSwiYWRtaW5fdXNlcl9pZCI6MCwiY3JlYXRlX3RpbWUiOjE3NDgxNDk2MDAsInVwZGF0ZV90aW1lIjoxNzQ4MTQ5NjAwfX0.HR3oMzgE0tBEHK1s3SEWmzwTpmSLix2Ew9ApgySaKP4' \
--form 'file=@"C:\\Users\\94804\\Desktop\\资源\\u=2990600787,3256164520&fm=253&gp=0.jpg"' \
--form 'business="app"'`

	fmt.Printf("原始命令: %s\n", example)
	parsed, err := p_api.ParseCurlCommand(example)
	if err != nil {
		fmt.Printf("解析错误: %v\n", err)
		return
	}

	fmt.Printf("解析结果:\n%s", parsed.String())
	fmt.Printf("等效命令: %s\n", parsed.ToCurlCommand())

	fmt.Println("程序结束")
}
