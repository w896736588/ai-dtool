package base

import (
	"dev_tool/base/define"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"net/url"
	"sync"
)

type TPlayWright struct {
	Page           *playwright.Page
	Browser        *playwright.Browser
	OpenType       int
	Value          string
	Context        *playwright.BrowserContext
	BrowserPid     int
	CreateTimeDesc string
}

type TSmartLink struct {
	PageList  map[string]*TPlayWright
	lock      sync.Mutex
	IsInstall bool
}

// GetPage 拿到Page
func (h *TSmartLink) GetPage(openType int, link, value, browserAuthUsername, browserAuthPassword string) (*TPlayWright, error) {
	h.lock.Lock()
	defer h.lock.Unlock()
	noViewPort := true
	javascript := true
	pw, pwErr := playwright.Run()
	if pwErr != nil {
		return nil, pwErr
	}
	//startFindTime := time.Now().UnixMilli()
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
	var context playwright.BrowserContext
	var contextErr error
	if browserAuthUsername != `` && browserAuthPassword != `` {
		context, contextErr = browser.NewContext(playwright.BrowserNewContextOptions{
			HttpCredentials: &playwright.HttpCredentials{
				Username: browserAuthUsername,
				Password: browserAuthPassword,
			},
			NoViewport:        &noViewPort,
			JavaScriptEnabled: &javascript,
		})
		if contextErr != nil {
			gstool.FmtPrintlnLogTime("Failed to create context: %v", contextErr)
		}
	} else {
		context, contextErr = browser.NewContext(playwright.BrowserNewContextOptions{
			NoViewport:        &noViewPort,
			JavaScriptEnabled: &javascript,
		})
		if contextErr != nil {
			gstool.FmtPrintlnLogTime("Failed to create context: %v", contextErr)
		}
	}
	page, pageErr = context.NewPage()
	if pageErr != nil {
		return nil, pageErr
	}
	//跳转链接
	u, _ := url.Parse(link)
	if _, goErr := page.Goto(u.String()); goErr != nil {
		return nil, goErr
	}
	//等待加载完成
	waitErr := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateLoad, //三种LoadStateNetworkidle 网络加载最低程度 LoadStateDomcontentloaded DOM加载完成
	})
	if waitErr != nil {
		gstool.FmtPrintlnLogTime("等待页面 DOM 内容加载完成失败: %s", waitErr.Error())
	}
	h.AddTipMsg(page, `寻找窗口中...`)
	createTimeDesc := cast.ToString(gstool.TimeNowMilliInt64())
	go h.FindPidMaxWindow(createTimeDesc)
	h.PageList[createTimeDesc] = &TPlayWright{
		Page:           &page,
		Browser:        &browser,
		OpenType:       openType,
		Context:        &context,
		Value:          value,
		CreateTimeDesc: createTimeDesc,
	}
	go func(createTimeDesc string) {
		page.OnClose(func(page playwright.Page) {
			h.lock.Lock()
			defer h.lock.Unlock()
			delete(h.PageList, createTimeDesc)
		})
	}(createTimeDesc)
	return h.PageList[createTimeDesc], nil
}

// FindPidMaxWindow 找到弹出的浏览器
func (h *TSmartLink) FindPidMaxWindow(createTimeDesc string) {
	//注意：当开启一个新页卡后或最小化后 进程无法唤醒
	nodePid := gstool.ProcessFindNewPidByName(`node.exe`)
	gstool.FmtPrintlnLogTime(`node pid %d`, nodePid)
	browserPid := gstool.ProcessFindNewPidByPPid(nodePid)
	h.SetWindowMax(cast.ToInt(browserPid))
	if pageInfo, ok := h.PageList[createTimeDesc]; ok {
		pageInfo.BrowserPid = cast.ToInt(browserPid)
	}
	gstool.FmtPrintlnLogTime(`创建page browserPid：%d`, browserPid)
}

// SetTitle 设置title
func (h *TSmartLink) SetTitle(page playwright.Page, title string) {
	_, _ = page.Evaluate(`(function() {
			document.title = "` + title + `";
	})();`)
}

// AddTipMsg 向页面上输出提示
func (h *TSmartLink) AddTipMsg(page playwright.Page, tip string) {
	_, _ = page.Evaluate(`(function() {
			setTimeout(function() {
				var existTip = document.getElementById('playwrightTipId');
				if (existTip) {
					existTip.remove();
				}
				var messageBox = document.createElement('div');
				messageBox.id = 'playwrightTipId';
				messageBox.textContent = '` + tip + `';
				messageBox.style.position = 'fixed';
				messageBox.style.top = '50%';
				messageBox.style.left = '50%';
				messageBox.style.transform = 'translate(-50%, -50%)';
				messageBox.style.color = 'white';
				messageBox.style.backgroundColor = 'black';
				messageBox.style.padding = '15px';
				messageBox.style.borderRadius = '10px';
				messageBox.style.boxShadow = '0 0 10px rgba(0, 0, 0, 0.5)';
				messageBox.style.zIndex = 2000;
				messageBox.style.display = 'block'; // 初始状态隐藏
				document.body.appendChild(messageBox);
				setTimeout(function() {
					var existTip = document.getElementById('playwrightTipId');
					if (existTip) {
						existTip.remove();
					}
				}, 2000); 
			}, 100); 
		})();`)
}
