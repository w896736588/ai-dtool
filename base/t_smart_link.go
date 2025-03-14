package base

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gstask"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/fsnotify/fsnotify"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"io"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path"
	"strings"
	"sync"
	"time"
)

type TSmartLink struct {
	RunLock sync.Mutex
	//处理下载后自动打开
	DownloadPath    string
	DownloadMapLock sync.Mutex
	//全局
	BrowserWebkitChrome  playwright.Browser
	BrowserWebkitSilence playwright.Browser
	//pw
	Pw *playwright.Playwright
	//浏览器列表
	ContextList []ContextPage
}

type ContextPage struct {
	Context            playwright.BrowserContext
	SmartLinkUniqueKey string //选项唯一值
	UserDataIndex      int    //数据目录索引
	UserDataPath       string //数据目录
	ContextUnique      string //唯一标记 context
}

// GetPage 拿到Page runUniqueKey 格式为0_common3 这种 单浏览器模式，每次打开都会打开一个新的浏览器
// openType 2 playwright静默打开 3 playwright有界面打开
// isSaveUserData 是否保存用户数据 1保存 0不保存
// link 打开的链接
// pageUniqueKey 本次开启新page的唯一值
// smartLinkUniqueKey 所属于链接唯一值
// browserUsername 浏览器自带验证用户名
// browserPassword 浏览器自带验证密码
// isCombine 是否自动合并不同域名到同一个浏览器 1合并 0不合并
// cookie 页面打开时设置的cookie
func (h *TSmartLink) GetPage(runParams _struct.SmartLinkRunParams) (playwright.Page, error) {
	h.RunLock.Lock()
	defer h.RunLock.Unlock()
	var contextErr error
	var contextPage ContextPage
	boolCleanFirstBlank := false
	if !runParams.IsSaveUserData { //不保存用户数据
		browser, browserErr := h.GetBrowser(runParams.OpenType)
		if browserErr != nil {
			return nil, browserErr
		}
		contextPage, contextErr = h.GetContextNotSaveUserData(runParams, browser)
	} else { //保留用户数据
		contextPage, boolCleanFirstBlank, contextErr = h.GetContextSaveUserData(runParams)
	}
	if contextErr != nil {
		return nil, contextErr
	}
	var page playwright.Page
	var pageErr error
	page, pageErr = contextPage.Context.NewPage()
	if pageErr != nil {
		return nil, pageErr
	}
	// 关闭一个blank
	if boolCleanFirstBlank {
		contextPageList := contextPage.Context.Pages()
		if len(contextPageList) > 0 {
			gstool.FmtPrintlnLogTime(`关闭页面 %#v`, contextPageList[0].URL())
			_ = contextPageList[0].Close()
		}
	}

	//监听下载事件进行重命名
	go h.OnDownload(page)
	//设置cookie
	if runParams.Cookie != `` {
		cookieErr := page.AddInitScript(playwright.Script{
			Content: playwright.String(runParams.Cookie),
		})
		if cookieErr != nil {
			gstool.FmtPrintlnLogTime(`设置cookie失败 %s`, cookieErr.Error())
		} else {
			gstool.FmtPrintlnLogTime(`设置cookie成功 %s`, runParams.Cookie)
		}
	}

	//跳转链接
	u, _ := url.Parse(runParams.Link)
	if _, goErr := page.Goto(u.String()); goErr != nil {
		return nil, goErr
	}
	//等待加载完成
	Component.TSmartLink.WaitForLoadState(page, runParams.Timeout)
	return page, nil
}

// CleanContextPageList 清理context
func (h *TSmartLink) CleanContextPageList(contextUnique string) {
	newContextList := make([]ContextPage, 0)
	for _, v := range h.ContextList {
		if v.ContextUnique != contextUnique {
			newContextList = append(newContextList, v)
		}
	}
	h.ContextList = newContextList
}

func (h *TSmartLink) GetPlaywrightRunList() []map[string]any {
	runList := make([]map[string]any, 0)
	for uniKey, runInfo := range h.ContextList {
		runList = append(runList, map[string]any{
			`name`:   runInfo.SmartLinkUniqueKey,
			`unikey`: uniKey,
		})
	}
	return runList
}

func (h *TSmartLink) LinkInit(slink string) string {
	link := gstool.StringReplaces(slink, map[string]string{
		`{rand}`:                   Component.TBase.GetUnique(`link_rand`),
		gstool.UrlEncode(`{rand}`): cast.ToString(Component.TBase.GetUnique(`link_rand`)),
	})
	gstool.FmtPrintlnLogTime(`Link %s => %s`, slink, link)
	return link
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

func (h *TSmartLink) IsSameLink(smartLinkUniqueKeyS, smartLinkUniqueKeyT string) bool {
	return strings.Split(smartLinkUniqueKeyS, `_`)[0] == strings.Split(smartLinkUniqueKeyT, `_`)[0]
}

func (h *TSmartLink) GetContextNotSaveUserData(runParams _struct.SmartLinkRunParams, browser playwright.Browser) (ContextPage, error) {
	for _, v := range h.ContextList {
		//非同种类型的context跳过
		if !h.IsSameLink(v.SmartLinkUniqueKey, runParams.SmartLinkUniqueKey) {
			continue
		}
		//找到一个context没有当前域名的
		boolFind := false
		pageList := v.Context.Pages()
		for _, v1 := range pageList {
			if gstool.UrlGetHost(v1.URL()) == runParams.Domain {
				boolFind = true
				break
			}
		}
		if !boolFind && runParams.IsCombine {
			return v, nil
		}
	}
	var context playwright.BrowserContext
	var contextErr error
	if runParams.BrowserAuthUsername != `` && runParams.BrowserAuthPassword != `` {
		context, contextErr = browser.NewContext(playwright.BrowserNewContextOptions{
			HttpCredentials: &playwright.HttpCredentials{
				Username: runParams.BrowserAuthUsername,
				Password: runParams.BrowserAuthPassword,
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
		return ContextPage{}, contextErr
	}
	contentPage := ContextPage{
		Context:            context,
		SmartLinkUniqueKey: runParams.SmartLinkUniqueKey,
		ContextUnique:      Component.TBase.GetUnique(`context_unique_`),
	}
	h.ContextList = append(h.ContextList, contentPage)
	//监听关闭
	go func() {
		context.OnClose(func(context playwright.BrowserContext) {
			h.CleanContextPageList(contentPage.ContextUnique)
		})
	}()
	return contentPage, nil
}

func (h *TSmartLink) GetBrowser(openType define.OpenType) (playwright.Browser, error) {
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

// GetContextSaveUserData 获取context 需要保存用户数据
func (h *TSmartLink) GetContextSaveUserData(runParams _struct.SmartLinkRunParams) (ContextPage, bool, error) {
	contextPage := h.GetUserDataContext(runParams)
	_ = gstool.DirCreatePath(contextPage.UserDataPath)
	if contextPage.Context != nil {
		return contextPage, false, nil
	}
	//打开模式
	Headless := false
	if runParams.OpenType == define.OpenTypeWebkitSilence {
		Headless = true
	}
	var context playwright.BrowserContext
	var contextErr error
	if runParams.BrowserAuthUsername != `` && runParams.BrowserAuthPassword != `` {
		context, contextErr = h.Pw.Chromium.LaunchPersistentContext(contextPage.UserDataPath, playwright.BrowserTypeLaunchPersistentContextOptions{
			DownloadsPath:     &h.DownloadPath,
			Headless:          &Headless,
			NoViewport:        playwright.Bool(true),
			JavaScriptEnabled: playwright.Bool(true),
			AcceptDownloads:   playwright.Bool(true),
			Locale:            playwright.String(`zh-CN`),
			Timeout:           &runParams.Timeout,
			IgnoreHttpsErrors: playwright.Bool(true),
			HttpCredentials: &playwright.HttpCredentials{
				Username: runParams.BrowserAuthUsername,
				Password: runParams.BrowserAuthPassword,
			},
			IgnoreDefaultArgs: []string{
				`--enable-automation`,
				`--disable-infobars`,                //禁用“正在使用自动化软件”提示信息栏。
				`--disable-features=IsolateOrigins`, //禁用隔离来源功能，允许跨域资源共享。
				`--disable-popup-blocking`,          //禁用弹出窗口阻止功能。
				`--allow-running-insecure-content`,  //允许加载不安全的内容（如 HTTP 资源）。
			},
		})
	} else {
		context, contextErr = h.Pw.Chromium.LaunchPersistentContext(contextPage.UserDataPath, playwright.BrowserTypeLaunchPersistentContextOptions{
			//DownloadsPath:     &h.DownloadPath,
			Headless:          &Headless,
			NoViewport:        playwright.Bool(true),
			JavaScriptEnabled: playwright.Bool(true),
			AcceptDownloads:   playwright.Bool(true),
			IgnoreHttpsErrors: playwright.Bool(true),
			Locale:            playwright.String(`zh-CN`),
			Timeout:           &runParams.Timeout,
			IgnoreDefaultArgs: []string{
				`--enable-automation`,
				`--disable-infobars`,                //禁用“正在使用自动化软件”提示信息栏。
				`--disable-features=IsolateOrigins`, //禁用隔离来源功能，允许跨域资源共享。
				`--disable-popup-blocking`,          //禁用弹出窗口阻止功能。
				`--allow-running-insecure-content`,  //允许加载不安全的内容（如 HTTP 资源）。
			},
		})
		gstool.FmtPrintlnLogTime(`启动 over`)
	}
	if contextErr != nil {
		return ContextPage{}, false, contextErr
	}
	contextPage.Context = context
	contextPage.ContextUnique = Component.TBase.GetUnique(`context_unique_`)
	h.ContextList = append(h.ContextList, contextPage)
	//监听关闭
	go func() {
		context.OnClose(func(context playwright.BrowserContext) {
			h.CleanContextPageList(contextPage.ContextUnique)
		})
	}()
	return contextPage, true, nil
}

// GetUserDataContext 拿到数据保存目录
func (h *TSmartLink) GetUserDataContext(runParams _struct.SmartLinkRunParams) ContextPage {
	userIndex := -1
	userIndexMax := -1
	for _, v := range h.ContextList {
		if userIndexMax < v.UserDataIndex {
			userIndexMax = v.UserDataIndex
		}
		//非同一类型的链接 不管
		if !h.IsSameLink(v.SmartLinkUniqueKey, runParams.SmartLinkUniqueKey) {
			continue
		}
		boolFind := false
		pageList := v.Context.Pages()
		gstool.FmtPrintlnLogTime(`打开的page 数量 %d`, len(pageList))
		for _, page := range pageList {
			if gstool.UrlGetHost(page.URL()) == runParams.Domain {
				boolFind = true
				break
			}
		}
		if !boolFind && runParams.IsCombine { //需要合并时才处理
			gstool.FmtPrintlnLogTime(`找到了可以复用的 %#v`, v)
			return v
		}
	}
	userIndex = userIndexMax + 1
	dataPath := fmt.Sprintf(Component.Env.PlaywrightUserData+`\%d`, userIndex)
	gstool.FmtPrintlnLogTime(`准备重新创建context %s`, dataPath)
	return ContextPage{
		Context:            nil,                          //初始化空
		SmartLinkUniqueKey: runParams.SmartLinkUniqueKey, //选项唯一值
		UserDataIndex:      userIndex,                    //数据目录索引
		UserDataPath:       dataPath,                     //数据目录
	}
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
		//gstool.FmtPrintlnLogTime(`下载%s`, response.Api().URL())
		//判断是否为文件或者图片 并下载到下载目录
		//h.downloadFileWithSuffixCheck(response.Api().URL())
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

func (h *TSmartLink) GetRunParams(id int, label, userName, password string, openNum int, replaceList []map[string]string) (_struct.SmartLinkRunParams, error) {
	runParams := _struct.SmartLinkRunParams{}
	if id == 0 {
		return runParams, errors.New(`链接ID不能为空`)
	}
	if label == `` {
		return runParams, errors.New(`链接label不能为空`)
	}
	runParams.Id = id
	smartLink, smartLinkErr := Component.TSqlite.Client.QueryBySql(`select * from tbl_smart_link where id = ? `, id).One()
	if smartLinkErr != nil {
		return runParams, errors.New(smartLinkErr.Error())
	}
	if len(smartLink) == 0 {
		return runParams, errors.New(`不存在的链接`)
	}
	linkList := make([]map[string]any, 0)
	decodeErr := gstool.JsonDecode(cast.ToString(smartLink[`links`]), &linkList)
	if decodeErr != nil {
		return runParams, errors.New(decodeErr.Error())
	}
	for _, link := range linkList {
		if cast.ToString(link[`label`]) == label {
			runParams.Link = cast.ToString(link[`link`])
			runParams.SmartLinkUniqueKey = cast.ToString(runParams.Id) + `_` + label
			runParams.OpenNum = 0
			runParams.Cookie = cast.ToString(link[`cookie`])
			runParams.BrowserAuthUsername = cast.ToString(link[`browser_auth_username`])
			runParams.BrowserAuthPassword = cast.ToString(link[`browser_auth_password`])
			break
		}
	}
	if runParams.Link == `` {
		return runParams, errors.New(`链接不存在，检查是否json格式错误`)
	}
	//赋值
	runParams.Link = gstool.StringReplaces(runParams.Link, map[string]string{
		`{rand}`: Component.TBase.GetCombineKey(),
	})
	runParams.IsSaveUserData = cast.ToInt(smartLink[`is_save_user_data`]) == 1
	runParams.IsCombine = cast.ToInt(smartLink[`is_combine`]) == 1
	runParams.OpenNum = cast.ToInt(math.Max(1, cast.ToFloat64(openNum)))
	runParams.OpenType = define.OpenType(cast.ToInt(smartLink[`open_type`]))
	process := cast.ToString(smartLink[`process`])
	processList := make([]map[string]any, 0)
	if process != `` {
		decodeErr = gstool.JsonDecode(process, &processList)
		if decodeErr != nil {
			return runParams, errors.New(`配置失败` + decodeErr.Error())
		}
	}
	runParams.Domain = gstool.UrlGetHost(runParams.Link)
	runParams.UserName = userName
	runParams.Password = password
	runParams.ProcessList = processList
	runParams.ReplaceList = replaceList
	runParams.Timeout = 3000
	return runParams, nil
}

// OpenBrowserPlaywright 打开浏览器
func (h *TSmartLink) OpenBrowserPlaywright(runParams _struct.SmartLinkRunParams) error {
	if Component.TSmartLink.Pw == nil {
		return errors.New(`未启动浏览器核心`)
	}
	page, pageErr := Component.TSmartLink.GetPage(runParams)

	if pageErr != nil {
		gstool.FmtPrintlnLogTime(`获取page报错 %s`, pageErr.Error())
		return pageErr
	}
	for _, processVal := range runParams.ProcessList {
		//类型
		processType := cast.ToString(processVal[`type`])
		//如果不存在
		notExistLocator := cast.ToString(processVal[`not_exist_Locator`])
		//元素选择
		Locator := cast.ToString(processVal[`Locator`])
		//链接
		redirectUri := cast.ToString(processVal[`uri`])
		//操作描述
		tip := cast.ToString(processVal[`tip`])
		// 等待页面加载完成
		Component.TSmartLink.WaitForLoadState(page, runParams.Timeout)
		waitUrlErr := page.WaitForURL(page.URL())
		if waitUrlErr != nil {
			return waitUrlErr
		}
		Component.TSmartLink.AddTipMsg(page, tip)
		switch processType {
		case `wait`:
			time.Sleep(time.Duration(cast.ToInt(processVal[`value`])) * time.Second)
		case `click`: //点击
			clickErr := h.click(Locator, notExistLocator, runParams.Timeout, page)
			if clickErr != nil {
				return clickErr
			}
		case `input`: //输入
			inputValue := cast.ToString(processVal[`value`])
			inputValue = gstool.StringReplaces(inputValue, map[string]string{
				`{user_name}`: runParams.UserName,
				`{password}`:  runParams.Password,
				`{rand}`:      Component.TBase.GetUnique(`input_rand_`),
			})
			//针对输入进行替换
			for _, replaceVal := range runParams.ReplaceList {
				inputValue = gstool.StringReplaces(inputValue, replaceVal)
			}
			inputSelecter := page.Locator(Locator)
			selectorLoaderWaitErr := inputSelecter.WaitFor(playwright.LocatorWaitForOptions{
				Timeout: &runParams.Timeout,
			})
			if selectorLoaderWaitErr == nil {
				inputErr := inputSelecter.Fill(inputValue)
				if inputErr != nil {
					gstool.FmtPrintlnLogTime("无法将元素转换为输入框: %v", inputErr.Error())
				}
			} else {
				Component.TSmartLink.AddTipMsg(page, `无法找到元素`+Locator+`,结束`)
				return errors.New(`无法找到元素` + Locator)
			}
		case `redirect_uri`: //跳转 保持当前域名
			currentURL := page.URL()
			parsedURL, err := url.Parse(currentURL)
			if err != nil {
				gstool.FmtPrintlnLogTime("could not parse URL: %v", err)
			}
			domain := parsedURL.Scheme + `://` + parsedURL.Host
			targetUrl := domain + redirectUri
			Component.TSmartLink.AddTipMsg(page, `准备跳转`)
			time.Sleep(time.Second)
			if _, goErr := page.Goto(targetUrl); goErr != nil {
				gstool.FmtPrintlnLogTime(`跳转地址出错 %s %s`, targetUrl, goErr.Error())
				return goErr
			}
		}
	}
	//无界面的5秒钟后自动关闭
	if runParams.OpenType == define.OpenTypeWebkitSilence {
		go func() {
			time.Sleep(time.Second * 5)
			closeErr := page.Close()
			if closeErr != nil {
				gstool.FmtPrintlnLogTime(`page close error：%s`, closeErr.Error())
			}
		}()
	}
	return nil
}

// SmartLinkPlaywrightVersion 获取浏览器核心版本
func (h *TSmartLink) SmartLinkPlaywrightVersion() (*playwright.PlaywrightDriver, error) {
	return playwright.NewDriver()
}

// 点击
func (h *TSmartLink) click(Locator, notExistLocator string, waitSecond float64, page playwright.Page) error {
	task := gstask.NewTask()
	waitSecond = 3 * 1000
	task.Add(gstask.CallbackFunc{
		Func: func() gstask.Result {
			selectorLoader := page.Locator(Locator)
			selectorLoaderWaitErr := selectorLoader.WaitFor(playwright.LocatorWaitForOptions{
				Timeout: &waitSecond,
				State:   playwright.WaitForSelectorStateVisible,
			})
			if selectorLoaderWaitErr != nil {
				return gstask.Result{
					Result: nil,
					State:  1,
					Err:    selectorLoaderWaitErr,
				}
			} else {
				return gstask.Result{
					Result: selectorLoader,
					State:  1,
					Err:    nil,
				}
			}
		},
		Timeout: 5 * time.Second,
	})
	if notExistLocator != `` {
		task.Add(gstask.CallbackFunc{
			Func: func() gstask.Result {
				existLoader := page.Locator(notExistLocator)
				existLoaderErr := existLoader.WaitFor(playwright.LocatorWaitForOptions{
					Timeout: &waitSecond,
					State:   playwright.WaitForSelectorStateVisible,
				})
				if existLoaderErr != nil {
					return gstask.Result{
						Result: nil,
						State:  2,
						Err:    existLoaderErr,
					}
				} else {
					return gstask.Result{
						Result: existLoader,
						State:  2,
						Err:    nil,
					}
				}
			},
			Timeout: 5 * time.Second,
		})
	}
	result := task.RunOne()
	if result.Err != nil {
		return result.Err
	}
	if result.State == 1 {
		element := result.Result.(playwright.Locator)
		_ = element.Click()
	}
	return nil
}
