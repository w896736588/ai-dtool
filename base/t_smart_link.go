package base

import (
	"dev_tool/base/define"
	"gitee.com/Sxiaobai/gs/gsdefine"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/fsnotify/fsnotify"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"net/url"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type TPlayWright struct {
	Page           *playwright.Page
	Browser        *playwright.Browser
	OpenType       int
	RunUniqueKey   string
	ContextS       *ContextS
	CreateTimeDesc string
}

type ContextS struct {
	Context playwright.BrowserContext
	Pid     int
}

type TSmartLink struct {
	PageList map[string]*TPlayWright
	RunLock  sync.Mutex
	//处理下载后自动打开
	DownloadPath    string
	DownloadMapLock sync.Mutex
	//全局
	BrowserWebkitChrome  playwright.Browser
	BrowserWebkitSilence playwright.Browser
	//domain context
	DomainContextMap map[string]ContextS
}

// GetPageSingle 拿到Page runUniqueKey 格式为0_common3 这种 单浏览器模式，每次打开都会打开一个新的浏览器
// isCombine 是否自动合并不同域名到同一个浏览器
func (h *TSmartLink) GetPageSingle(openType int, link, runUniqueKey, browserAuthUsername, browserAuthPassword string, isCombine int) (*TPlayWright, error) {
	h.RunLock.Lock()
	defer h.RunLock.Unlock()
	//browser
	browser, browserErr := h.GetBrowser(openType)
	if browserErr != nil {
		return nil, browserErr
	}
	host := gstool.UrlGetHost(link)
	domainKey := host + `:` + browserAuthUsername + `:` + browserAuthPassword
	//context
	context, contextErr := h.GetContext(domainKey, browserAuthUsername, browserAuthPassword, browser, isCombine)
	if contextErr != nil {
		return nil, contextErr
	}
	//page
	var page playwright.Page
	var pageErr error
	page, pageErr = context.Context.NewPage()
	if pageErr != nil {
		return nil, pageErr
	}
	//监听下载事件进行重命名
	go h.OnDownload(page)
	//跳转链接
	u, _ := url.Parse(link)
	if _, goErr := page.Goto(u.String()); goErr != nil {
		return nil, goErr
	}
	//等待加载完成
	Component.TSmartLink.WaitForLoadState(page)
	//记录进程列表
	createTimeDesc := cast.ToString(gstool.TimeNowMilliInt64())
	h.PageList[createTimeDesc] = &TPlayWright{
		Page:           &page,
		Browser:        &browser,
		OpenType:       openType,
		ContextS:       &context,
		RunUniqueKey:   runUniqueKey,
		CreateTimeDesc: createTimeDesc,
	}
	go func(createTimeDesc string) {
		page.OnClose(func(page playwright.Page) {
			gstool.FmtPrintlnLogTime(`监听到页面关闭 %s`, createTimeDesc)
			h.RunLock.Lock()
			defer h.RunLock.Unlock()
			delete(h.PageList, createTimeDesc)
			h.CheckContextActive()
		})
	}(createTimeDesc)

	return h.PageList[createTimeDesc], nil
}

func (h *TSmartLink) CheckContextActive() {
	for k, v := range h.DomainContextMap {
		if !v.Context.Browser().IsConnected() {
			delete(h.DomainContextMap, k)
		}
	}
}

func (h *TSmartLink) WaitForLoadState(page playwright.Page) {
	_ = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateDomcontentloaded,
	})
	_ = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateNetworkidle,
	})
	_ = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State: playwright.LoadStateLoad,
	})
}

func (h *TSmartLink) GetContext(domainKey, browserAuthUsername, browserAuthPassword string, browser playwright.Browser, isCombine int) (ContextS, error) {
	h.CheckContextActive()
	for k, v := range h.DomainContextMap {
		if v.Context.Browser().IsConnected() && k != domainKey && isCombine == 1 {
			if v.Pid != 0 {
				go func() {
					_ = h.SetForegroundWindowPid(v.Pid)
				}()
			}
			return v, nil
		}
	}
	var context playwright.BrowserContext
	var contextErr error
	if browserAuthUsername != `` && browserAuthPassword != `` {
		context, contextErr = browser.NewContext(playwright.BrowserNewContextOptions{
			HttpCredentials: &playwright.HttpCredentials{
				Username: browserAuthUsername,
				Password: browserAuthPassword,
			},
			NoViewport:        playwright.Bool(true),
			JavaScriptEnabled: playwright.Bool(true),
			AcceptDownloads:   playwright.Bool(true),
		})
		if contextErr != nil {
			return ContextS{}, contextErr
		}
	} else {
		context, contextErr = browser.NewContext(playwright.BrowserNewContextOptions{
			NoViewport:        playwright.Bool(true),
			JavaScriptEnabled: playwright.Bool(true),
			AcceptDownloads:   playwright.Bool(true),
		})
		if contextErr != nil {
			return ContextS{}, contextErr
		}
	}

	h.DomainContextMap[domainKey] = ContextS{
		Context: context,
		Pid:     0,
	}
	go h.FindPidMaxWindow(domainKey)
	return h.DomainContextMap[domainKey], nil
}

func (h *TSmartLink) GetBrowser(openType int) (playwright.Browser, error) {
	if openType == define.OpenTypeWebkitSilence && h.BrowserWebkitSilence != nil {
		return h.BrowserWebkitSilence, nil
	} else if openType == define.OpenTypeWebkitChrome && h.BrowserWebkitChrome != nil {
		return h.BrowserWebkitChrome, nil
	}
	pw, pwErr := playwright.Run()
	if pwErr != nil {
		return nil, pwErr
	}
	var browserErr error
	if openType == define.OpenTypeWebkitSilence {
		h.BrowserWebkitSilence, browserErr = pw.Chromium.Launch()
		if browserErr != nil {
			h.BrowserWebkitSilence = nil
			return nil, browserErr
		} else {
			return h.BrowserWebkitSilence, nil
		}
	} else {
		h.BrowserWebkitChrome, browserErr = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			DownloadsPath: &h.DownloadPath,
			Headless:      playwright.Bool(false), //有界面模式
		})
		if browserErr != nil {
			h.BrowserWebkitChrome = nil
			return nil, browserErr
		} else {
			return h.BrowserWebkitChrome, nil
		}
	}

}

func (h *TSmartLink) OnDownload(page playwright.Page) {
	page.On(`download`, func(download playwright.Download) {
		h.DownloadMapLock.Lock()
		defer h.DownloadMapLock.Unlock()
		//content, _ := gstool.FileGetContent(download.URL())
		//其实没用 应为内容都不一样。。
		//h.DownloadMap[gstool.Md5(content)] = filepath.Base(download.String())
		//gstool.FmtPrintlnLogTime(`下载 %s => %s`, gstool.Md5(content), filepath.Base(download.String()))
		//gstool.FmtPrintlnLogTime(`下载文件 %s %s`, download.URL(), download.String())
		//downloadErr := download.SaveAs(h.DownloadPath + `\` + download.String())
		//if downloadErr != nil {
		//	gstool.FmtPrintlnLogTime(`下载`)
		//	return
		//} else {
		//	gstool.FmtPrintlnLogTime(`另存下载文件%s`, h.DownloadPath+`\`+download.String())
		//}
	})
	page.On("response", func(response playwright.Response) {
		//gstool.FmtPrintlnLogTime(`下载%s`, response.Request().URL())
	})
}

// WitchDownload 监听目录新文件下载 自动识别文件类型 并打开
func (h *TSmartLink) WitchDownload() {
	_ = os.RemoveAll(h.DownloadPath)
	_ = gstool.DirCreatePath(h.DownloadPath)
	gstool.FmtPrintlnLogTime(`开始监听%s`, h.DownloadPath)
	watch := gstool.NewFileWatch(h.DownloadPath, func(event fsnotify.Event) {
		if event.Op == fsnotify.Create {
			if strings.HasSuffix(event.Name, `.crdownload`) || strings.Contains(event.Name, `~$`) {
				return
			}
			targetName := event.Name
			gstool.FmtPrintlnLogTime(`targetName %s => %s`, event.Name, targetName)
			isXlsx := gstool.FileIsXlsx(event.Name)
			if isXlsx {
				h.OpenFile(event.Name, targetName, `xlsx`)
			} else {
				ext, extErr := gstool.FileExtType(event.Name)
				if extErr == nil {
					h.OpenFile(event.Name, targetName, ext.Extension)
				}
			}

		}
	})
	go func() {
		err := watch.Start()
		if err != nil {
			gstool.FmtPrintlnLogTime(`监听失败 ^%s`, err.Error())
		}
	}()
}

func (h *TSmartLink) OpenFile(filePath, targetFilePath, ext string) {
	allowTypeList := []string{
		`xlsx`, `xls`, `csv`, `doc`, `docx`, `ppt`, `pptx`, `pdf`,
		`txt`, `md`, `html`, `htm`, `jpg`, `jpeg`, `png`, `gif`,
		`bmp`, `ico`, `svg`, `mp4`, `mp3`, `wav`,
	}
	compareName := strings.ToLower(filePath)
	boolStart := false
	for _, allowType := range allowTypeList {
		if strings.Contains(compareName, allowType) {
			boolStart = true
			break
		}
	}
	if boolStart {
		gstool.FmtPrintlnLogTime(`直接打开 %s`, filePath)
		cmd := exec.Command("cmd", "/C", "start", filePath)
		_ = cmd.Start()
	} else {
		renameErr := os.Rename(filePath, targetFilePath+`.`+ext)
		gstool.FmtPrintlnLogTime(`移动后打开 %s => %s`, filePath, targetFilePath+`.`+ext)
		if renameErr != nil {
			gstool.FmtPrintlnLogTime(`重命名错误 %s`, renameErr.Error())
		}
	}

}

// FindPidMaxWindow 找到弹出的浏览器 还是时常不准，还会影响加载速度，先不管了
func (h *TSmartLink) FindPidMaxWindow(domainKey string) {
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second * 1)
		findPidList := make([]map[string]int, 0)
		list := gstool.ProcessList()
		for _, process := range list {
			cmd := cast.ToString(process[`cmd`])
			exe := cast.ToString(process[`exe`])
			createTime := cast.ToInt(process[`create_time`])
			if strings.Contains(exe, `chromium`) && strings.Contains(exe, `playwright`) && strings.Contains(cmd, `no-startup-window`) {
				findPidList = append(findPidList, map[string]int{
					`create_time`: createTime,
					`pid`:         cast.ToInt(process[`pid`]),
				})
			}
		}
		gstool.ArrayMapSort(&findPidList, `create_time`, gsdefine.SortDesc)
		choosePid := 0
		if len(findPidList) > 0 {
			choosePid = cast.ToInt(findPidList[0][`pid`])
			//判断是否已经存在了
			for _, context := range h.DomainContextMap {
				if context.Pid == choosePid {
					choosePid = 0
					break
				}
			}
			if choosePid > 0 {
				h.SetPid(domainKey, choosePid)
				break
			}
		} else {
			continue
		}
	}
}

func (h *TSmartLink) SetPid(domainKey string, pid int) {
	h.RunLock.Lock()
	defer h.RunLock.Unlock()
	if domainVal, ok := h.DomainContextMap[domainKey]; ok {
		domainVal.Pid = pid
		h.DomainContextMap[domainKey] = domainVal
		h.SetWindowMax(domainVal.Pid)
	}
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

func (h *TSmartLink) SmartCheckAndUpdate() {
	pw, _ := playwright.NewDriver()
	lockFileName := `playwright.RunLock`
	lockFileFullPath := Component.Env.RootPath + `/` + lockFileName
	if !gstool.FileIsExisted(lockFileFullPath) {
		go h.install(pw.Version, lockFileFullPath)
	} else {
		content, contentErr := gstool.FileGetContent(lockFileFullPath)
		if contentErr != nil {
			gstool.FmtPrintlnLogTime(`获取文件内容失败 %s`, contentErr.Error())
		} else if content != pw.Version {
			go h.install(pw.Version, lockFileFullPath)
		} else {
			gstool.FmtPrintlnLogTime(`浏览器核心最新版本为：%s ，当前安装版本为：%s,不需要进行更新`, pw.Version, content)
		}
	}
}

func (h *TSmartLink) install(version, lockFileFullPath string) {
	_ = gstool.FilePutContentCover(lockFileFullPath, version)
	gstool.FmtPrintlnLogTime(`开始安装浏览器核心(只安装chrome),大约几分钟时间`)
	err := playwright.Install(&playwright.RunOptions{Browsers: []string{`chromium`}})
	if err != nil {
		gstool.FmtPrintlnLogTime(`安装浏览器核心失败 %s`, err.Error())
		_ = gstool.FileDelete(lockFileFullPath)
	} else {
		gstool.FmtPrintlnLogTime(`安装完成`)
	}
}
