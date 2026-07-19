package controller

import (
	"dev_tool/internal/app/dtool/component"
	"github.com/gin-gonic/gin"
	"os"
	"path/filepath"
)

// E2ERecorderProxyHTML 返回 e2e-recorder.html 静态页。
// 部署期需先在 web/dist 下构建出 e2e-recorder.html；找不到文件时返回 404 而不是 panic，
// 避免拖垮 dtool 主进程。
func E2ERecorderProxyHTML(c *gin.Context) {
	// component.EnvClient.WebConfig.WebPath 是构建期配置的 web/dist 根目录；
	// 默认由 InitEnv 在启动时赋值（rootPath/web/dist），需要 dtool 已完成配置初始化。
	baseDir := ""
	if component.EnvClient != nil && component.EnvClient.WebConfig != nil {
		baseDir = component.EnvClient.WebConfig.WebPath
	}
	if baseDir == "" {
		wd, _ := os.Getwd()
		baseDir = filepath.Join(wd, "web", "dist")
	}
	filePath := filepath.Join(baseDir, "e2e-recorder.html")
	body, err := os.ReadFile(filePath)
	if err != nil {
		c.String(404, "recorder html not found: "+err.Error())
		return
	}
	c.Header("Content-Type", "text/html; charset=utf-8")
	c.Header("Cross-Origin-Resource-Policy", "cross-origin")
	c.Header("Content-Security-Policy", "frame-ancestors *")
	c.String(200, string(body))
}
