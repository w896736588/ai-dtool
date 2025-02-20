package base

import (
	"dev_tool/base/define"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdefine"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/fsnotify/fsnotify"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"
)

type TPlayWright struct {
	Page               *playwright.Page
	Browser            *playwright.Browser
	OpenType           int
	SmartLinkUniqueKey string
	ContextS           *ContextS
	CreateTimeDesc     string
}

type ContextS struct {
	Context       playwright.BrowserContext
	Pid           int
	DomainList    []Domain
	UserDataIndex int
	Unique        string
}

type Domain struct {
	Domain        string
	PageUniqueKey string
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
	DomainContextList []*ContextS
	//save data
	DomainContextListU []*ContextS
	//pw
	Pw *playwright.Playwright
}

// GetPage 拿到Page runUniqueKey 格式为0_common3 这种 单浏览器模式，每次打开都会打开一个新的浏览器
// isCombine 是否自动合并不同域名到同一个浏览器
func (h *TSmartLink) GetPage(openType, isSaveUserData int, link, pageUniqueKey, smartLinkUniqueKey, browserAuthUsername, browserAuthPassword string, isCombine int) (*TPlayWright, error) {
	h.RunLock.Lock()
	defer h.RunLock.Unlock()
	link = h.LinkInit(link)
	var timeout float64 = 3000
	var context *ContextS
	var contextErr error
	host := gstool.UrlGetHost(link)
	domain := host
	gstool.FmtPrintlnLogTime(`是否保存用户数据 %d 打开模式 %d`, isSaveUserData, openType)
	boolCleanFirstBlank := false
	if isSaveUserData != 1 { //不保存用户数据
		browser, browserErr := h.GetBrowser(openType)
		if browserErr != nil {
			return nil, browserErr
		}
		context, contextErr = h.GetContextNotSaveUserData(domain, pageUniqueKey, browserAuthUsername, browserAuthPassword, browser, isCombine)
	} else {
		context, boolCleanFirstBlank, contextErr = h.GetContextSaveUserData(domain, pageUniqueKey, browserAuthUsername, browserAuthPassword, timeout, openType)
	}
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
	// 关闭一个blank
	if boolCleanFirstBlank {
		pageList := context.Context.Pages()
		if len(pageList) > 0 {
			gstool.FmtPrintlnLogTime(`关闭页面 %#v`, pageList[0].URL())
			_ = pageList[0].Close()
		}
	}

	//监听下载事件进行重命名
	go h.OnDownload(page)
	//跳转链接
	u, _ := url.Parse(link)
	if _, goErr := page.Goto(u.String()); goErr != nil {
		return nil, goErr
	}
	//等待加载完成
	Component.TSmartLink.WaitForLoadState(page, timeout)
	//记录进程列表
	h.PageList[pageUniqueKey] = &TPlayWright{
		Page:               &page,
		OpenType:           openType,
		ContextS:           context,
		SmartLinkUniqueKey: smartLinkUniqueKey,
		CreateTimeDesc:     cast.ToString(gstool.TimeNowMilliInt64()),
	}

	go func(pageUniqueKey string) {
		page.OnClose(func(page playwright.Page) {
			gstool.FmtPrintlnLogTime(`监听到页面关闭 移除运行列表 %s %s`, page.URL(), pageUniqueKey)
			h.RunLock.Lock()
			defer h.RunLock.Unlock()
			delete(h.PageList, pageUniqueKey)
		})
	}(pageUniqueKey)

	return h.PageList[pageUniqueKey], nil
}

func (h *TSmartLink) LinkInit(slink string) string {
	link := gstool.StringReplaces(slink, map[string]string{
		`{rand}`:                   Component.TBase.GetUnique(`link_rand`),
		gstool.UrlEncode(`{rand}`): cast.ToString(Component.TBase.GetUnique(`link_rand`)),
	})
	gstool.FmtPrintlnLogTime(`Link %s => %s`, slink, link)
	return link
}

// CheckContextActive 检查是否有活跃的并移除domainKey
func (h *TSmartLink) CheckContextActive(domainKey string) {
	newList := make([]*ContextS, 0)
	for _, v := range h.DomainContextList {
		if v.Context.Browser().IsConnected() {
			//检查是否存在被关闭的域名
			if domainKey != `` {
				newDomainList := make([]Domain, 0)
				for _, v1 := range v.DomainList {
					if v1.Domain != domainKey {
						newDomainList = append(newDomainList, v1)
					} else {
						gstool.FmtPrintlnLogTime(`移除context中域名 %s`, domainKey)
					}
				}
				v.DomainList = newDomainList
			}

			newList = append(newList, v)
		}
	}
	h.DomainContextList = newList
}

// CheckContextActiveU 检查是否有活跃的并移除domainKey
func (h *TSmartLink) CheckContextActiveU(domainKey string) {
	newList := make([]*ContextS, 0)
	for _, v := range h.DomainContextListU {
		if v.Context != nil && v.Context.Browser() != nil && v.Context.Browser().IsConnected() {
			//检查是否存在被关闭的域名
			if domainKey != `` {
				newDomainList := make([]Domain, 0)
				for _, v1 := range v.DomainList {
					if v1.Domain != domainKey {
						newDomainList = append(newDomainList, v1)
					} else {
						gstool.FmtPrintlnLogTime(`移除context中域名 %s`, domainKey)
					}
				}
				v.DomainList = newDomainList
			}

			newList = append(newList, v)
		}
	}
	h.DomainContextListU = newList
}

func (h *TSmartLink) WaitForLoadState(page playwright.Page, timeout float64) {
	_ = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateDomcontentloaded,
		Timeout: &timeout,
	})
	_ = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: &timeout,
	})
	_ = page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateLoad,
		Timeout: &timeout,
	})
}

func (h *TSmartLink) GetContextNotSaveUserData(domain, pageUniqueKey, browserAuthUsername, browserAuthPassword string, browser playwright.Browser,
	isCombine int) (*ContextS, error) {
	h.CheckContextActive(``)
	for _, v := range h.DomainContextList {
		//找到一个context没有当前域名的
		boolFind := false
		for _, v1 := range v.DomainList {
			if v1.Domain == domain {
				boolFind = true
				break
			}
		}
		if !boolFind && isCombine == 1 {
			if v.Pid != 0 {
				go func() {
					//_ = h.SetForegroundWindowPid(v.Pid)
				}()
			}
			v.DomainList = append(v.DomainList, Domain{
				Domain:        domain,
				PageUniqueKey: pageUniqueKey,
			})
			return v, nil
		}
	}
	gstool.FmtPrintlnLogTime(`重新创建 %s`, domain)
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
			Locale:            playwright.String(`zh-CN`),
		})
	} else {
		context, contextErr = browser.NewContext(playwright.BrowserNewContextOptions{
			NoViewport:        playwright.Bool(true),
			JavaScriptEnabled: playwright.Bool(true),
			AcceptDownloads:   playwright.Bool(true),
			Locale:            playwright.String(`zh-CN`),
		})
	}
	if contextErr != nil {
		return &ContextS{}, contextErr
	}
	contentS := &ContextS{
		Context:    context,
		Pid:        0,
		DomainList: []Domain{{Domain: domain, PageUniqueKey: pageUniqueKey}},
		Unique:     Component.TBase.GetUnique(`context_`),
	}
	h.DomainContextList = append(h.DomainContextList, contentS)
	//监听关闭
	go func() {
		context.OnClose(func(context playwright.BrowserContext) {
			h.RunLock.Lock()
			defer h.RunLock.Unlock()
			newList := make([]*ContextS, 0)
			//移除已经移出去的context
			for _, v := range h.DomainContextList {
				if v.Unique != contentS.Unique {
					newList = append(newList, v)
				}
			}
			h.DomainContextList = newList
		})
	}()
	return contentS, nil
}

func (h *TSmartLink) GetBrowser(openType int) (playwright.Browser, error) {
	if openType == define.OpenTypeWebkitSilence && h.BrowserWebkitSilence != nil {
		return h.BrowserWebkitSilence, nil
	} else if openType == define.OpenTypeWebkitChrome && h.BrowserWebkitChrome != nil {
		return h.BrowserWebkitChrome, nil
	}
	var browserErr error
	if openType == define.OpenTypeWebkitSilence {
		h.BrowserWebkitSilence, browserErr = h.Pw.Chromium.Launch()
		if browserErr != nil {
			h.BrowserWebkitSilence = nil
			return nil, browserErr
		} else {
			return h.BrowserWebkitSilence, nil
		}
	} else {
		h.BrowserWebkitChrome, browserErr = h.Pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
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

func (h *TSmartLink) GetContextSaveUserData(domain, pageUniqueKey, browserAuthUsername, browserAuthPassword string,
	timeout float64, openType int) (*ContextS, bool, error) {
	dataPath, contextS, userDataIndex := h.GetUserDataDirectory(domain, pageUniqueKey)
	if contextS != nil {
		return contextS, false, nil
	}
	_ = gstool.DirCreatePath(dataPath)
	//打开模式
	Headless := false
	if openType == define.OpenTypeWebkitSilence {
		Headless = true
	}
	gstool.FmtPrintlnLogTime(`是否有界面 %t`, Headless)
	var context playwright.BrowserContext
	var contextErr error
	if browserAuthUsername != `` && browserAuthPassword != `` {
		context, contextErr = h.Pw.Chromium.LaunchPersistentContext(dataPath, playwright.BrowserTypeLaunchPersistentContextOptions{
			DownloadsPath:     &h.DownloadPath,
			Headless:          &Headless,
			NoViewport:        playwright.Bool(true),
			JavaScriptEnabled: playwright.Bool(true),
			AcceptDownloads:   playwright.Bool(true),
			Locale:            playwright.String(`zh-CN`),
			Timeout:           &timeout,
			HttpCredentials: &playwright.HttpCredentials{
				Username: browserAuthUsername,
				Password: browserAuthPassword,
			},
			IgnoreDefaultArgs: []string{
				`--enable-automation`,
			},
		})
	} else {
		gstool.FmtPrintlnLogTime(`启动 LaunchPersistentContext`)
		context, contextErr = h.Pw.Chromium.LaunchPersistentContext(dataPath, playwright.BrowserTypeLaunchPersistentContextOptions{
			//DownloadsPath:     &h.DownloadPath,
			Headless:          &Headless,
			NoViewport:        playwright.Bool(true),
			JavaScriptEnabled: playwright.Bool(true),
			AcceptDownloads:   playwright.Bool(true),
			Locale:            playwright.String(`zh-CN`),
			Timeout:           &timeout,
			IgnoreDefaultArgs: []string{
				`--enable-automation`,
			},
		})
		gstool.FmtPrintlnLogTime(`启动 over`)
	}
	if contextErr != nil {
		return &ContextS{}, false, contextErr
	}

	contentS := &ContextS{
		Context:       context,
		Pid:           0,
		DomainList:    []Domain{{Domain: domain, PageUniqueKey: pageUniqueKey}},
		UserDataIndex: userDataIndex,
		Unique:        Component.TBase.GetUnique(`context_`),
	}
	h.DomainContextListU = append(h.DomainContextListU, contentS)
	//监听关闭
	go func() {
		context.OnClose(func(context playwright.BrowserContext) {
			h.RunLock.Lock()
			defer h.RunLock.Unlock()
			newList := make([]*ContextS, 0)
			//移除已经移出去的context
			for _, v := range h.DomainContextListU {
				if v.Unique != contentS.Unique {
					newList = append(newList, v)
				}
			}
			h.DomainContextListU = newList
		})
	}()
	return contentS, true, nil
}

func (h *TSmartLink) GetUserDataDirectory(domain, pageUniqueKey string) (string, *ContextS, int) {
	userIndex := -1
	userIndexMax := -1
	for k, v := range h.DomainContextListU {
		if userIndexMax < v.UserDataIndex {
			userIndexMax = v.UserDataIndex
		}
		boolFind := false
		for _, v1 := range v.DomainList {
			if v1.Domain == domain {
				boolFind = true
				break
			}
		}
		if boolFind {
			continue
		}
		userIndex = v.UserDataIndex
		v.DomainList = append(v.DomainList, Domain{
			Domain:        domain,
			PageUniqueKey: pageUniqueKey,
		})
		h.DomainContextListU[k] = v
		return fmt.Sprintf(Component.Env.PlaywrightUserData+`\%d`, userIndex), v, userIndex
	}
	userIndex = userIndexMax + 1
	dataPath := fmt.Sprintf(Component.Env.PlaywrightUserData+`\%d`, userIndex)
	return dataPath, nil, userIndex
}

func (h *TSmartLink) OnDownload(page playwright.Page) {
	page.On(`download`, func(download playwright.Download) {
		h.DownloadMapLock.Lock()
		defer h.DownloadMapLock.Unlock()
		gstool.FmtPrintlnLogTime(`下载 %s %s`, download.SuggestedFilename(), download.URL())
		time.Sleep(time.Second)
		localPath := h.DownloadPath + `/` + Component.TBase.GetUnique(`download`) + `_` + download.SuggestedFilename()
		ret := download.SaveAs(localPath)
		gstool.FmtPrintlnLogTime(`下载结果 %#v`, ret)
	})
	page.On("response", func(response playwright.Response) {
		//gstool.FmtPrintlnLogTime(`下载%s`, response.Request().URL())
		//判断是否为文件或者图片 并下载到下载目录
		//h.downloadFileWithSuffixCheck(response.Request().URL())
	})
}

// 下载文件并检查后缀
func (h *TSmartLink) downloadFileWithSuffixCheck(rawURL string) error {
	// 解析 URL 并移除查询参数
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return fmt.Errorf("failed to parse URL: %v", err)
	}
	parsedURL.RawQuery = "" // 移除查询参数

	// 获取清理后的 URL
	cleanURL := parsedURL.String()

	// 检查文件后缀
	if !h.isValidFileSuffix(cleanURL) {
		return fmt.Errorf("invalid file suffix for URL: %s", cleanURL)
	}

	// 发送 HTTP 请求
	resp, err := http.Get(cleanURL)
	if err != nil {
		return fmt.Errorf("failed to fetch URL: %v", err)
	}
	defer resp.Body.Close()

	// 检查 HTTP 响应状态码
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	// 确定文件名和保存路径
	filename := path.Base(parsedURL.Path)
	savePath := path.Join(h.DownloadPath, Component.TBase.GetUnique(`download`)+filename)

	// 保存文件
	file, err := os.Create(savePath)
	if err != nil {
		return fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		return fmt.Errorf("failed to write file: %v", err)
	}

	fmt.Printf("File downloaded successfully: %s\n", savePath)
	return nil
}

// 检查是否为常见文件后缀
func (h *TSmartLink) isValidFileSuffix(url string) bool {
	commonSuffixes := map[string]bool{
		".png":  true,
		".jpg":  true,
		".jpeg": true,
		".gif":  true,
		".bmp":  true,
		".webp": true,
		".pdf":  true,
		".doc":  true,
		".docx": true,
		".xls":  true,
		".xlsx": true,
		".txt":  true,
		".zip":  true,
		".mp4":  true,
		".avi":  true,
	}

	ext := strings.ToLower(path.Ext(url))
	return commonSuffixes[ext]
}

// WitchDownload 监听目录新文件下载 自动识别文件类型 并打开
func (h *TSmartLink) WitchDownload() {
	_ = gstool.DirCreatePath(h.DownloadPath)
	gstool.FmtPrintlnLogTime(`开始监听%s`, h.DownloadPath)
	watch := gstool.NewFileWatch(h.DownloadPath, func(event fsnotify.Event) {
		if event.Op == fsnotify.Create {
			gstool.FmtPrintlnLogTime(`监听到文件下载了 %#v`, event)
			ext, extErr := gstool.FileExtType(event.Name)
			gstool.FmtPrintlnLogTime(`文件后缀 %s %v`, ext, extErr)
			cmd := exec.Command("cmd", "/C", "start", event.Name)
			_ = cmd.Start()
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
func (h *TSmartLink) FindPidMaxWindow(contentS *ContextS) {
	for i := 0; i < 5; i++ {
		time.Sleep(time.Second * 1)
		findPidList := make([]map[string]int, 0)
		list := gstool.ProcessList()
		for _, process := range list {
			cmd := cast.ToString(process[`cmd`])
			exe := cast.ToString(process[`exe`])
			createTime := cast.ToInt(process[`create_time`])
			if strings.Contains(exe, `chromium`) && strings.Contains(exe, `playwright`) && strings.Contains(cmd, `no-startup-window`) && !strings.Contains(cmd, `chromium_headless_shell`) {
				gstool.FmtPrintlnLogTime(`第%d次查找进程 create_time：%s pid：%s cmd：%s exe：%s`, i, createTime, process[`pid`], ``, exe)
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
			for _, context := range h.DomainContextList {
				if context.Pid == choosePid {
					choosePid = 0
					break
				}
			}
			if choosePid > 0 {
				contentS.Pid = choosePid
				gstool.FmtPrintlnLogTime(`找到了pid %s`, choosePid)
				//还是有时候不准
				//h.SetWindowMax(contentS.Pid)
				break
			}
		} else {
			continue
		}
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
			go h.InitPlaywright()
		}
	}
}

func (h *TSmartLink) InitPlaywright() {
	gstool.FmtPrintlnLogTime(`启动浏览器核心..`)
	var pwErr error
	h.Pw, pwErr = playwright.Run()
	if pwErr != nil {
		return
	}
	h.BrowserWebkitSilence, _ = h.Pw.Chromium.Launch()
	h.BrowserWebkitChrome, _ = h.Pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		DownloadsPath: &h.DownloadPath,
		Headless:      playwright.Bool(false), //有界面模式
	})
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
		h.InitPlaywright()
	}
}
