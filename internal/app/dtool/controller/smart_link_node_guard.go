package controller

import (
	"dev_tool/internal/app/dtool/plw"
	"dev_tool/internal/pkg/p_sse"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
)

const nodeInstallURL = "https://nodejs.org/zh-cn/download"

// ensureSmartLinkNodeInstalled 校验自定义网页运行所需 Node.js 环境
func ensureSmartLinkNodeInstalled(c *gin.Context, sse *p_sse.SseShell) bool {
	if plw.PlaywrightClient != nil && plw.PlaywrightClient.EnsureNodeRuntime() {
		return true
	}
	if sse != nil {
		sse.Send("检测到未安装 Node.js，请先安装后再使用自定义网页\n")
	}
	gsgin.GinResponseError(c, "未检测到 Node.js，请先安装后再使用自定义网页", map[string]any{
		"need_install_node": 1,
		"install_url":       nodeInstallURL,
		"install_tip":       "请安装 Node.js（建议 LTS 版本），安装完成后重新进入自定义网页页面。",
	})
	return false
}
