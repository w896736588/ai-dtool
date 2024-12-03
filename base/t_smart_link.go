package base

import (
	"dev_tool/base/define"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/go-vgo/robotgo"
	"github.com/playwright-community/playwright-go"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cast"
	"net/url"
	"strings"
	"sync"
	"time"
)

type TPlayWright struct {
	Page       *playwright.Page
	Browser    *playwright.Browser
	OpenType   int
	Value      string
	Context    *playwright.BrowserContext
	BrowserPid int
}

type TSmartLink struct {
	PageList map[string]*TPlayWright
	lock     sync.Mutex
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

	waitErr := page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded, //三种LoadStateNetworkidle 网络加载最低程度 LoadStateDomcontentloaded DOM加载完成
	})
	if waitErr != nil {
		gstool.FmtPrintlnLogTime("等待页面 DOM 内容加载完成失败: %s", waitErr.Error())
	}
	//唯一值
	createTimeDesc := cast.ToString(gstool.TimeNowMilliInt64())
	sourceTitle, _ := page.Title()
	h.SetTitle(page, createTimeDesc)
	//停一下
	time.Sleep(time.Millisecond * 200)
	browserPid := h.FindRecentChildPid(createTimeDesc, `chrome`)
	//还原title
	h.SetTitle(page, sourceTitle)
	h.PageList[createTimeDesc] = &TPlayWright{
		Page:       &page,
		Browser:    &browser,
		OpenType:   openType,
		Context:    &context,
		Value:      value,
		BrowserPid: browserPid,
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

// SetTitle 设置title
func (h *TSmartLink) SetTitle(page playwright.Page, title string) {
	_, _ = page.Evaluate(`(function() {
			document.title = "` + title + `";
	})();`)
}

func (h *TSmartLink) FindChildPIDs(parentPID int32) ([]int32, error) {
	processList, err := process.Processes()
	if err != nil {
		return nil, err
	}

	var childPIDs []int32
	for _, proc := range processList {
		name, _ := proc.Name()
		gstool.FmtPrintlnLogTime(proc.String() + ` ` + name + ` ` + cast.ToString(proc.Pid))
		ppid, ppidErr := proc.Ppid()
		if ppidErr != nil {
			continue
		}
		if ppid == parentPID {
			childPIDs = append(childPIDs, proc.Pid)
		}
	}
	return childPIDs, nil
}

// FindRecentChildPid 获取title和名字获取进程
func (h *TSmartLink) FindRecentChildPid(title, name string) int {
	processList, err := process.Processes()
	if err != nil {
		return 0
	}
	for _, proc := range processList {
		pName, _ := proc.Name()
		ppid, _ := proc.Ppid()
		pTitle := robotgo.GetTitle(cast.ToInt(proc.Pid))
		if pTitle == `` {
			continue
		}
		if name != `` && !strings.Contains(pName, name) {
			continue
		}
		gstool.FmtPrintlnLogTime(`name:%v ,  title:%v ,  pid:%v , ppid:%v`, pName, pTitle, proc.Pid, ppid)
		if strings.Contains(pTitle, title) {
			return cast.ToInt(proc.Pid)
		}
	}
	return 0
}
