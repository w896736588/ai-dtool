package base

import (
	"dev_tool/base/define"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"sync"
)

type TPlayWright struct {
	Page     *playwright.Page
	Browser  *playwright.Browser
	OpenType int
}

type TSmartLink struct {
	PageList map[string]*TPlayWright
	lock     sync.Mutex
}

func (h *TSmartLink) GetPage(openType int, browserAuthUsername, browserAuthPassword string) (*TPlayWright, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	noViewPort := true
	javascript := true
	pw, pwErr := playwright.Run()
	if pwErr != nil {
		return nil, pwErr
	}
	var browser playwright.Browser
	var browserErr error
	if openType == define.OpenTypeWebkitSilence {
		browser, browserErr = pw.Chromium.Launch()
	} else {
		browser, browserErr = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(false), //有界面模式
		})
	}
	if browserErr != nil {
		return nil, browserErr
	}
	// 创建带有认证信息的浏览器上下文
	var page playwright.Page
	var pageErr error
	if browserAuthUsername != `` && browserAuthPassword != `` {
		context, contextErr := browser.NewContext(playwright.BrowserNewContextOptions{
			HttpCredentials: &playwright.HttpCredentials{
				Username: "admin",
				Password: "123456",
			},
			NoViewport:        &noViewPort,
			JavaScriptEnabled: &javascript,
		})
		if contextErr != nil {
			gstool.FmtPrintlnLogTime("Failed to create context: %v", contextErr)
		}
		page, pageErr = context.NewPage()
		if pageErr != nil {
			return nil, pageErr
		}
	} else {
		page, pageErr = browser.NewPage(playwright.BrowserNewPageOptions{NoViewport: &noViewPort, JavaScriptEnabled: &javascript})
		if pageErr != nil {
			return nil, pageErr
		}
	}

	createTimeDesc := cast.ToString(gstool.TimeNowMilliInt64())
	h.PageList[createTimeDesc] = &TPlayWright{
		Page:     &page,
		Browser:  &browser,
		OpenType: openType,
	}
	return h.PageList[createTimeDesc], nil
}
