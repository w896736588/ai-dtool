package controller

import (
	"dev_tool/base"
	"gitee.com/Sxiaobai/gs/v2/gsapi"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
	"net/url"
)

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
		gstool.FmtPrintlnLogTime(`错误了 %s`, err.Error())
		call(err.Error())
	}
}
