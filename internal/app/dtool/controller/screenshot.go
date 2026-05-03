package controller

import (
	"encoding/base64"
	"fmt"
	"net/url"
	"strings"

	"dev_tool/internal/app/dtool/component"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
)

const (
	ScreenshotDefaultTimeout = 30
	ScreenshotMaxTimeout     = 120
	ScreenshotDefaultWidth   = 1920
	ScreenshotDefaultHeight  = 1080
	ScreenshotMaxWidth       = 3840
	ScreenshotMaxHeight      = 2160
)

// Screenshot 网页截图接口
// POST /api/Screenshot
// 参数: url(必填), full_page(可选), width(可选), height(可选), timeout(可选), selector(可选)
func Screenshot(c *gin.Context) {
	reqMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &reqMap)

	// 校验 URL
	targetURL := strings.TrimSpace(cast.ToString(reqMap["url"]))
	if targetURL == "" {
		gsgin.GinResponseError(c, "url不能为空", nil)
		return
	}
	parsedURL, parseErr := url.ParseRequestURI(targetURL)
	if parseErr != nil {
		gsgin.GinResponseError(c, "url不合法: "+parseErr.Error(), nil)
		return
	}
	scheme := strings.ToLower(parsedURL.Scheme)
	if scheme != "http" && scheme != "https" {
		gsgin.GinResponseError(c, "url仅支持http/https协议", nil)
		return
	}

	// 解析可选参数
	fullPage := cast.ToBool(reqMap["full_page"])
	width := clampInt(cast.ToInt(reqMap["width"]), ScreenshotDefaultWidth, 1, ScreenshotMaxWidth)
	height := clampInt(cast.ToInt(reqMap["height"]), ScreenshotDefaultHeight, 1, ScreenshotMaxHeight)
	timeout := clampInt(cast.ToInt(reqMap["timeout"]), ScreenshotDefaultTimeout, 1, ScreenshotMaxTimeout)
	selector := strings.TrimSpace(cast.ToString(reqMap["selector"]))

	// 确保浏览器已就绪
	if component.PlaywrightClient == nil || component.PlaywrightClient.Pw == nil {
		gsgin.GinResponseError(c, "浏览器核心未启动", nil)
		return
	}
	browser := component.PlaywrightClient.BrowserWebkitSilence
	if browser == nil {
		gsgin.GinResponseError(c, "浏览器核心未初始化", nil)
		return
	}

	// 创建临时上下文和页面
	timeoutMs := float64(timeout * 1000)
	context, ctxErr := browser.NewContext(playwright.BrowserNewContextOptions{
		Viewport: &playwright.Size{
			Width:  width,
			Height: height,
		},
	})
	if ctxErr != nil {
		gsgin.GinResponseError(c, "创建浏览器上下文失败: "+ctxErr.Error(), nil)
		return
	}
	defer context.Close()

	page, pageErr := context.NewPage()
	if pageErr != nil {
		gsgin.GinResponseError(c, "创建页面失败: "+pageErr.Error(), nil)
		return
	}

	// 导航到目标页面
	if _, gotoErr := page.Goto(targetURL, playwright.PageGotoOptions{
		Timeout:   &timeoutMs,
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
	}); gotoErr != nil {
		gsgin.GinResponseError(c, "页面导航失败: "+gotoErr.Error(), nil)
		return
	}

	// 等待页面加载完成
	component.PlaywrightClient.WaitForLoadState(&page, timeoutMs)

	// 截图
	var screenshotBuf []byte
	var ssErr error
	if selector != "" {
		locator := page.Locator(selector)
		count, countErr := locator.Count()
		if countErr != nil || count == 0 {
			gsgin.GinResponseError(c, fmt.Sprintf("未找到选择器对应的元素: %s", selector), nil)
			return
		}
		screenshotBuf, ssErr = locator.Screenshot(playwright.LocatorScreenshotOptions{
			Timeout: &timeoutMs,
		})
	} else {
		screenshotBuf, ssErr = page.Screenshot(playwright.PageScreenshotOptions{
			FullPage: &fullPage,
			Timeout:  &timeoutMs,
			Type:     playwright.ScreenshotTypePng,
		})
	}
	if ssErr != nil {
		gsgin.GinResponseError(c, "截图失败: "+ssErr.Error(), nil)
		return
	}

	// 返回 base64 编码的图片
	gsgin.GinResponseSuccess(c, "", map[string]any{
		"image":      base64.StdEncoding.EncodeToString(screenshotBuf),
		"image_type": "png",
		"url":        targetURL,
		"full_page":  fullPage,
		"width":      width,
		"height":     height,
	})
}

// clampInt 将值限制在 [min, max] 范围内，如果 val <= 0 则返回 def
func clampInt(val, def, min, max int) int {
	if val <= 0 {
		return def
	}
	if val < min {
		return min
	}
	if val > max {
		return max
	}
	return val
}
