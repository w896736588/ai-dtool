package controller

import (
	"dev_tool/internal/pkg/p_gitlab"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsapi"
	"github.com/w896736588/go-tool/gstool"
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
	gitlab := p_gitlab.TGitlab{
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
