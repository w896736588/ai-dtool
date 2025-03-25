package controller

import (
	"dev_tool/base"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsapi"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"net/url"
)

func GitLabLogs(c *gin.Context) {
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("Transfer-Encoding", "chunked")
	// 开启流式输出
	reqData := gsgin.GinGetParams(c)
	flusher, ok := c.Writer.(gin.ResponseWriter)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "Streaming not supported"})
		return
	}
	GitLogs(reqData, func(s string) {
		data := make(map[string]any)
		data[`data`] = s + "\n"
		_, err := fmt.Fprintf(c.Writer, "data: "+gstool.JsonEncode(data)+"\n\n")
		if err != nil {
			gstool.FmtPrintlnLogTime(`错误 %s`, err.Error())
			return
		}
		gstool.FmtPrintlnLogTime(`输出 %s`, "data: "+gstool.JsonEncode(data)+"\n\n")
		flusher.Flush()
	})
}

func GitLogs(reqData url.Values, call func(string)) {
	accessToken := cast.ToString(reqData.Get(`access_token`))
	baseUrl := cast.ToString(reqData.Get(`base_url`))
	author := cast.ToString(reqData.Get(`author`))
	if accessToken == `` || baseUrl == `` || author == `` {
		call(`参数错误`)
		return
	}
	gitlab := base.TGitlab{
		GitLab: gsapi.GsGitLab{
			BaseUrl:     baseUrl,
			AccessToken: accessToken,
		},
		LogFunc: call,
		Author:  author,
	}
	_, err := gitlab.GetTodayLogs()
	if err != nil {
		call(err.Error())
	}
}
