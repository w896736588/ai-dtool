package base

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gshttp"
	"gitee.com/Sxiaobai/gs/gstask"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"log"
	"math"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type TPlaywright struct {
	RunLock sync.Mutex
	//处理下载后自动打开
	downloadPath string
	EventLock    sync.Mutex
	//全局
	BrowserWebkitChrome  playwright.Browser
	BrowserWebkitSilence playwright.Browser
	//pw
	Pw *playwright.Playwright
	//浏览器列表
	ContextList []ContextPage
	//page 活跃时间
	pageActiveTime map[string]PageActiveTime
	//是否运行中
	IsRun bool
	//监听链接
	ListenUrlList map[string]*_struct.ListenUrl
	log           *gstool.GsSlog
	//文件
	LockFileFullPath string
}

type PageActiveTime struct {
	ActiveTime time.Time
	RunParams  *_struct.PlaywrightRunParams
	Page       playwright.Page
}

type ContextPage struct {
	Context            playwright.BrowserContext
	SmartLinkUniqueKey string          //选项唯一值
	UserDataIndex      int             //数据目录索引
	UserDataPath       string          //数据目录
	ContextUnique      string          //唯一标记 context
	OpenType           define.OpenType //打开类型
}

func NewTSmartLink() *TPlaywright {
	gsLog := gstool.NewSlogDefault(Component.Env.LogPath, `playwright`)
	gsLog.DeleteLogs(``)
	return &TPlaywright{
		log:            gsLog,
		downloadPath:   Component.Env.PlaywrightDownload,
		pageActiveTime: make(map[string]PageActiveTime),
		ListenUrlList:  make(map[string]*_struct.ListenUrl),
	}
}

func (h *TPlaywright) SetWebkitPath() {
	// 设置自定义浏览器安装路径
	gstool.FmtPrintlnLogTime(`核心位置 %s`, Component.Env.PkgPath+"/p_webkit")
	_ = os.Setenv("PLAYWRIGHT_BROWSERS_PATH", Component.Env.PkgPath+"/p_webkit")
	_ = os.Setenv("PLAYWRIGHT_DRIVER_PATH", Component.Env.PkgPath+"/p_webkit")
	_ = os.Setenv("GOPROXY", "https://goproxy.cn,direct")
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
func (h *TPlaywright) GetPage(runParams *_struct.PlaywrightRunParams) (playwright.Page, error) {
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
			h.log.Debugf(`关闭页面 %#v`, contextPageList[0].URL())
			_ = contextPageList[0].Close()
		}
	}
	//设置cookie
	if runParams.Cookie != `` {
		cookieErr := page.AddInitScript(playwright.Script{
			Content: playwright.String(runParams.Cookie),
		})
		if cookieErr != nil {
			h.log.Errof(`设置cookie失败 %s`, cookieErr.Error())
		} else {
			h.log.Debugf(`设置cookie成功 %s`, runParams.Cookie)
		}
	}
	//设置header
	if len(runParams.Headers) > 0 {
		// 拦截请求并动态设置头
		err := page.Route("**/*", func(route playwright.Route) {
			// 获取请求的 URL 或类型
			request := route.Request()
			requestUrl := request.URL()

			// 判断是否为资源文件（如 CSS、JS、图片等）
			isResourceFile := strings.HasSuffix(requestUrl, ".css") || strings.HasSuffix(requestUrl, ".js") ||
				strings.HasSuffix(requestUrl, ".png") || strings.HasSuffix(requestUrl, ".jpg")

			// 如果不是资源文件，设置请求头
			if !isResourceFile {
				headers := request.Headers()
				for headerKey, headerVal := range runParams.Headers {
					headers[headerKey] = headerVal
				}
				setErr := route.Continue(playwright.RouteContinueOptions{Headers: headers})
				if setErr != nil {
					h.log.Errof(`setExtraHTTPHeaders %s`, setErr.Error())
				} else {
					h.log.Debugf(`给%s设置header%s`, requestUrl, gstool.JsonEncode(runParams.Headers))
				}
			} else {
				continueErr := route.Continue()
				if continueErr != nil {
					h.log.Errof(`continue %s`, continueErr.Error())
				}
			}
		})
		if err != nil {
			h.log.Errof("could not set up route: %v", err)
		}
	}

	//跳转链接
	u, _ := url.Parse(runParams.Link)
	if _, goErr := page.Goto(u.String()); goErr != nil {
		return nil, goErr
	}
	//等待加载完成
	Component.TPlaywright.WaitForLoadState(page, runParams.LocatorTimeout)
	return page, nil
}

// CleanContextPageList 清理context
func (h *TPlaywright) CleanContextPageList(contextUnique string) {
	newContextList := make([]ContextPage, 0)
	for _, v := range h.ContextList {
		if v.ContextUnique != contextUnique {
			newContextList = append(newContextList, v)
		}
	}
	h.ContextList = newContextList
}

// CleanContextPageByDomain 根据域名清理context
func (h *TPlaywright) CleanContextPageByDomain(domain string) {
	for _, v := range h.ContextList {
		contextPageList := v.Context.Pages()
		for _, page := range contextPageList {
			if gstool.UrlGetHost(page.URL()) == domain {
				_ = page.Close()
			}
		}
	}
}

func (h *TPlaywright) GetPlaywrightRunList() []map[string]any {
	runList := make([]map[string]any, 0)
	for uniKey, runInfo := range h.ContextList {
		pageList := runInfo.Context.Pages()

		runList = append(runList, map[string]any{
			`name`:     runInfo.SmartLinkUniqueKey,
			`unikey`:   uniKey,
			`page_num`: len(pageList),
		})
	}
	return runList
}

func (h *TPlaywright) LinkInit(slink string) string {
	link := gstool.SReplaces(slink, map[string]string{
		`{rand}`:                   Component.TBase.GetUnique(`link_rand`),
		gstool.UrlEncode(`{rand}`): cast.ToString(Component.TBase.GetUnique(`link_rand`)),
	})
	h.log.Debugf(`Link %s => %s`, slink, link)
	return link
}

func (h *TPlaywright) WaitForLoadState(page playwright.Page, timeout float64) {
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

func (h *TPlaywright) IsSameLink(smartLinkUniqueKeyS, smartLinkUniqueKeyT string) bool {
	return strings.Split(smartLinkUniqueKeyS, `_`)[0] == strings.Split(smartLinkUniqueKeyT, `_`)[0]
}

func (h *TPlaywright) GetContextNotSaveUserData(runParams *_struct.PlaywrightRunParams, browser playwright.Browser) (ContextPage, error) {
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
	context.OnPage(func(page playwright.Page) {
		go h.PageEvents(runParams, page)
	})
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

func (h *TPlaywright) GetBrowser(openType define.OpenType) (playwright.Browser, error) {
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
			//DownloadsPath: &h.downloadPath,
			Headless: playwright.Bool(false), //有界面模式
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
func (h *TPlaywright) GetContextSaveUserData(runParams *_struct.PlaywrightRunParams) (ContextPage, bool, error) {
	//固定打开数据索引 关闭同域名的
	if runParams.FixDataId == 1 {
		h.log.Debugf(`清理相同域名page`)
		h.CleanContextPageByDomain(runParams.Domain)
	}
	//获取context
	contextPage := h.GetUserDataContext(runParams)
	//创建数据索引目录
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
	//浏览器自带验证
	if runParams.BrowserAuthUsername != `` && runParams.BrowserAuthPassword != `` {
		context, contextErr = h.Pw.Chromium.LaunchPersistentContext(contextPage.UserDataPath, playwright.BrowserTypeLaunchPersistentContextOptions{
			//DownloadsPath:     &h.downloadPath,
			Headless:          &Headless,
			Channel:           playwright.String(runParams.Channel), // 使用完整版 Chrome 而非 Chromium
			NoViewport:        playwright.Bool(true),
			JavaScriptEnabled: playwright.Bool(true),
			AcceptDownloads:   playwright.Bool(true),
			Locale:            playwright.String(`zh-CN`),
			Timeout:           &runParams.GetPageTimeout,
			IgnoreHttpsErrors: playwright.Bool(true),
			HttpCredentials: &playwright.HttpCredentials{
				Username: runParams.BrowserAuthUsername,
				Password: runParams.BrowserAuthPassword,
			},
			IgnoreDefaultArgs: []string{
				`--enable-automation`,
				`--disable-infobars`,                            //禁用“正在使用自动化软件”提示信息栏。
				`--disable-features=IsolateOrigins`,             //禁用隔离来源功能，允许跨域资源共享。
				`--disable-popup-blocking`,                      //禁用弹出窗口阻止功能。
				`--allow-running-insecure-content`,              //允许加载不安全的内容（如 HTTP 资源）。
				`--disable-blink-features=AutomationControlled`, //禁止传递浏览器自动化标识
			},
		})
		if contextErr != nil {
			h.log.Errof(`启动context报错 %s`, contextErr.Error())
			return ContextPage{}, false, contextErr
		}
		contextPage.Context = context
		contextPage.ContextUnique = Component.TBase.GetUnique(`context_unique_`) //这里是动态生成的唯一ID 其实没啥意义
	} else {
		h.log.Debugf(`启动context 超时时间：%f`, runParams.GetPageTimeout)
		context, contextErr = h.Pw.Chromium.LaunchPersistentContext(contextPage.UserDataPath, playwright.BrowserTypeLaunchPersistentContextOptions{
			//DownloadsPath:     &h.downloadPath,
			Headless: &Headless,
			//Channel:           playwright.String(runParams.Channel),//增加这个会导致问题 关闭后不能正常启动下一个
			NoViewport:        playwright.Bool(true),
			JavaScriptEnabled: playwright.Bool(true),
			AcceptDownloads:   playwright.Bool(true),
			IgnoreHttpsErrors: playwright.Bool(true),
			Locale:            playwright.String(`zh-CN`),
			Timeout:           &runParams.GetPageTimeout,
			IgnoreDefaultArgs: []string{
				`--enable-automation`,
				`--disable-infobars`,                            //禁用“正在使用自动化软件”提示信息栏。
				`--disable-features=IsolateOrigins`,             //禁用隔离来源功能，允许跨域资源共享。
				`--disable-popup-blocking`,                      //禁用弹出窗口阻止功能。
				`--allow-running-insecure-content`,              //允许加载不安全的内容（如 HTTP 资源）。
				`--disable-blink-features=AutomationControlled`, //禁止传递浏览器自动化标识
			},
		})
		if contextErr != nil {
			h.log.Errof(`启动context报错 %s`, contextErr.Error())
			return ContextPage{}, false, contextErr
		}
		h.log.Debugf(`启动 over`)
		contextPage.Context = context
		if runParams.FixDataId != 0 { //如果是固定打开数据索引 那么给予一个固定的
			contextPage.ContextUnique = fmt.Sprintf(`context_unique_%d_%d`, runParams.OpenType, runParams.Id)
		} else {
			contextPage.ContextUnique = Component.TBase.GetUnique(`context_unique_`)
		}
	}
	contextPage.SmartLinkUniqueKey = runParams.SmartLinkUniqueKey
	contextPage.OpenType = runParams.OpenType
	context.OnPage(func(page playwright.Page) {
		go h.PageEvents(runParams, page)
	})
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
func (h *TPlaywright) GetUserDataContext(runParams *_struct.PlaywrightRunParams) ContextPage {
	var userIndex int
	if runParams.FixDataId == 0 {
		userIndex = -1
		userIndexMax := -1
		for _, v := range h.ContextList {
			if userIndexMax < v.UserDataIndex {
				userIndexMax = v.UserDataIndex
			}
			//是否允许合并
			if !runParams.IsCombine {
				continue
			}
			//非同一类型打开方式 不管
			if v.OpenType != runParams.OpenType {
				continue
			}
			//非同一类型的链接 不管
			if !h.IsSameLink(v.SmartLinkUniqueKey, runParams.SmartLinkUniqueKey) {
				continue
			}
			//是否有相同域名的page存在
			boolFindSameDomainPage := false
			pageList := v.Context.Pages()
			h.log.Debugf(`打开的page 数量 %d`, len(pageList))
			for _, page := range pageList {
				if gstool.UrlGetHost(page.URL()) == runParams.Domain {
					boolFindSameDomainPage = true
					break
				}
			}
			//没有找到相同域名的page
			if !boolFindSameDomainPage { //需要合并时才处理
				h.log.Debugf(`找到了可以复用的 %#v`, v)
				return v
			}
		}
		//递增一次索引
		userIndex = userIndexMax + 1
	} else {
		userIndex = runParams.Id
		//fmt.Sprintf(`context_unique_%d`, runParams.FixDataId)
		for _, v := range h.ContextList {
			if v.ContextUnique == fmt.Sprintf(`context_unique_%d_%d`, runParams.OpenType, runParams.Id) {
				h.log.Debugf(`找到了已经存在的context`)
				return v
			}
		}
	}

	dataPath := fmt.Sprintf(Component.Env.PlaywrightUserData+`\%d`, userIndex)
	h.log.Debugf(`context使用的数据目录 %s`, dataPath)
	return ContextPage{
		Context:            nil,                          //初始化空
		SmartLinkUniqueKey: runParams.SmartLinkUniqueKey, //选项唯一值
		UserDataIndex:      userIndex,                    //数据目录索引
		UserDataPath:       dataPath,                     //数据目录
	}
}

func (h *TPlaywright) WitchDownload() {
	_ = gstool.DirCreatePath(h.downloadPath)
	if err := os.MkdirAll(h.downloadPath, 0755); err != nil {
		log.Fatalf("创建目录失败: %v", err)
	}
	h.log.Debugf(`开始监听%s`, h.downloadPath)
}

// SetTitle 设置title
func (h *TPlaywright) SetTitle(page playwright.Page, title string) {
	_, _ = page.Evaluate(`(function() {
			document.title = "` + title + `";
	})();`)
}

// AddTipMsg 向页面上输出提示
func (h *TPlaywright) AddTipMsg(page playwright.Page, tip string) {
	if tip == `` {
		return
	}
	content := Component.TJas.Get(`p_js`, `tip.js`)
	content = gstool.SReplaces(content, map[string]string{
		`{tip}`: tip,
	})
	_, _ = page.Evaluate(content)
}

// ShowCookieTip 展示cookie中的某个值
func (h *TPlaywright) ShowCookieTip(page playwright.Page) {
	configList := []_struct.ShowCookie{
		{
			FindType:   `cookie`,
			FindKey:    "xkf_userid",
			Label:      "UserId",
			DomainList: []string{"xiaokefu.com.cn", "applnk.cn", "ishipinhao.com"},
		},
		{
			FindType:     `any`,
			Label:        "Username",
			FormatList:   []string{`url_decode`},
			RegexFindKey: `s:8:"username";s:\d+:"(.+)"`,
			DomainList:   []string{"xiaokefu.com.cn", "applnk.cn", "ishipinhao.com"},
		},
	}
	replaceList := make([]_struct.ShowCookie, 0)
	for _, config := range configList {
		if gstool.SContains(strings.ToLower(page.URL()), config.DomainList) {
			replaceList = append(replaceList, config)
		}
	}
	if len(replaceList) == 0 {
		return
	}
	config := gstool.JsonEncode(replaceList)
	Component.GsLog.Debugf(`配置的js %s`, config)
	content := Component.TJas.Get(`p_js`, `info.js`)
	content = gstool.SReplaces(content, map[string]string{
		`{config}`: config,
	})
	_, _ = page.Evaluate(content)
}

func (h *TPlaywright) SmartCheckAndUpdate() {
	pw, _ := playwright.NewDriver()
	if !gstool.FileIsExisted(h.LockFileFullPath) {
		go h.Install(pw.Version)
	} else {
		content, contentErr := gstool.FileGetContent(h.LockFileFullPath)
		if contentErr != nil {
			h.log.Errof(`获取文件内容失败 %s`, contentErr.Error())
		} else if content != pw.Version {
			go h.Install(pw.Version)
		} else {
			h.log.Debugf(`浏览器核心最新版本为：%s ，当前安装版本为：%s,不需要进行更新`, pw.Version, content)
			go h.InitPlaywright()
		}
	}
}

func (h *TPlaywright) InitPlaywright() {
	h.log.Debugf(`启动浏览器核心..`)
	var pwErr error
	h.Pw, pwErr = playwright.Run()
	if pwErr != nil {
		h.log.Debugf(`启动浏览器核心失败 %s`, pwErr.Error())
		return
	}
	h.BrowserWebkitSilence, _ = h.Pw.Chromium.Launch()
	h.BrowserWebkitChrome, _ = h.Pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		//DownloadsPath: &h.downloadPath,
		Headless: playwright.Bool(false), //有界面模式
	})
}

func (h *TPlaywright) Install(version string) {
	gstool.FmtPrintlnLogTime(`开始安装浏览器核心(只安装chrome),大约几分钟时间`)
	err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{`chromium`},
	})
	if err != nil {
		gstool.FmtPrintlnLogTime(`安装浏览器核心失败 %s`, err.Error())
		_ = gstool.FileDelete(h.LockFileFullPath)
	} else {
		_ = gstool.FilePutContentCover(h.LockFileFullPath, version)
		gstool.FmtPrintlnLogTime(`安装完成`)
		h.InitPlaywright()
	}
}

func (h *TPlaywright) GetRunParams(id int, label, userName, password string, openNum int, replaceList *[]map[string]string) (*_struct.PlaywrightRunParams, error) {
	runParams := &_struct.PlaywrightRunParams{}
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
	runParams.FixDataId = cast.ToInt(smartLink[`fix_data_id`])
	runParams.DownloadFinds = strings.Split(cast.ToString(smartLink[`download_finds`]), `,`)
	runParams.AutoCloseSecond = cast.ToInt(smartLink[`auto_close_second`])
	runParams.Channel = cast.ToString(smartLink[`channel`])
	if runParams.Channel == `` {
		runParams.Channel = `chromium`
	}
	h.log.Debugf(`使用浏览器核心 ` + runParams.Channel)
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
			headerMap := make(map[string]string)
			_ = gstool.JsonDecode(cast.ToString(link[`headers`]), &headerMap)
			runParams.Headers = headerMap
			runParams.BrowserAuthUsername = cast.ToString(link[`browser_auth_username`])
			runParams.BrowserAuthPassword = cast.ToString(link[`browser_auth_password`])
			break
		}
	}
	if runParams.Link == `` {
		return runParams, errors.New(`链接不存在，检查是否json格式错误`)
	}
	//赋值
	runParams.Link = gstool.SReplaces(runParams.Link, map[string]string{
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
	runParams.ReplaceList = *replaceList
	runParams.LocatorTimeout = 1000
	runParams.GetPageTimeout = 3000
	return runParams, nil
}

// OpenBrowserPlaywright 打开浏览器
func (h *TPlaywright) OpenBrowserPlaywright(runParams *_struct.PlaywrightRunParams) error {
	if Component.TPlaywright.Pw == nil {
		return errors.New(`未启动浏览器核心`)
	}
	page, pageErr := Component.TPlaywright.GetPage(runParams)
	if pageErr != nil {
		h.log.Errof(`获取page报错 %s`, pageErr.Error())
		return pageErr
	}
	//提取内容存储
	takeContentMap := make(map[string]string)
	//输出结果存储
	boolResultMap := make(map[string]bool)
	for _, processVal := range runParams.ProcessList {
		//限制域名执行
		domainLimit := cast.ToString(processVal[`domain_limit`])
		if domainLimit != `` && !strings.Contains(runParams.Domain, domainLimit) {
			continue
		}
		//类型
		processType := cast.ToString(processVal[`type`])
		//元素选择
		locator := cast.ToString(processVal[`Locator`])
		//操作描述
		tip := cast.ToString(processVal[`tip`])
		//检查 替换等
		checkKey := cast.ToString(processVal[`check_key`])
		checkKey = gstool.SReplaces(checkKey, map[string]string{
			`{user_name}`: runParams.UserName,
			`{password}`:  runParams.Password,
		})
		checkKey = gstool.SReplaces(checkKey, takeContentMap)
		//输出
		outKey := cast.ToString(processVal[`out_key`])
		//检查是否允许执行 当需要输出的时候不进行判断
		//h.log.Debugf(`outKey：%s checkKey：%s tip：%s boolResultMap：%v takeContentMap：%v`, outKey, checkKey, tip, boolResultMap, takeContentMap)
		// 等待页面加载完成
		Component.TPlaywright.WaitForLoadState(page, runParams.LocatorTimeout)
		waitUrlErr := page.WaitForURL(page.URL())
		if waitUrlErr != nil {
			return waitUrlErr
		}

		cmdType := define.CmdType(processType)
		switch cmdType {
		case define.TextContent: //提取内容
			Component.TPlaywright.AddTipMsg(page, tip)
			elementOp := &_struct.ElementOp{
				Type: define.ElementTextContent,
			}
			_, elementErr := h.DoLocator(&page, locator, elementOp)
			if elementErr != nil {
				h.callRun(runParams, cmdType, elementErr.Error(), tip, locator)
			} else {
				takeContentMap[outKey] = strings.TrimSpace(elementOp.TextContent)
				h.callRun(runParams, cmdType, ``, tip, elementOp.TextContent)
			}
		case define.BoolResult: //bool结果判断
			Component.TPlaywright.AddTipMsg(page, tip)
			if locator != `` {
				elementOp := &_struct.ElementOp{
					Type: define.ElementCount,
				}
				_, elementErr := h.DoLocator(&page, locator, elementOp)
				if elementErr != nil || elementOp.Count == 0 {
					boolResultMap[outKey] = false //不存在
				} else {
					boolResultMap[outKey] = true //存在
				}
				h.log.Debugf(`判断 %s`, gstool.JsonEncode(boolResultMap))
			} else {
				//根据上面的执行来判断
				h.outKeyBoolResult(outKey, checkKey, boolResultMap, takeContentMap, runParams)
			}
		case define.Exit:
			if !h.allowCheckKey(checkKey, boolResultMap) {
				Component.TPlaywright.AddTipMsg(page, tip)
				return errors.New(tip)
			}
		case define.Close:
			if !h.allowCheckKey(checkKey, boolResultMap) {
				continue
			}
			Component.TPlaywright.AddTipMsg(page, tip)
			_ = page.Close()
		case define.Wait:
			if !h.allowCheckKey(checkKey, boolResultMap) {
				continue
			}
			Component.TPlaywright.AddTipMsg(page, tip)
			time.Sleep(time.Duration(cast.ToInt(processVal[`value`])) * time.Second)
		case define.WaitClose:
			go func() {
				if !h.allowCheckKey(checkKey, boolResultMap) {
					return
				}
				Component.TPlaywright.AddTipMsg(page, tip)
				time.Sleep(time.Duration(cast.ToInt(processVal[`value`])) * time.Second)
				_ = page.Close()
			}()
		case define.Click: //点击
			if !h.allowCheckKey(checkKey, boolResultMap) {
				h.log.Debugf(`点击 %s 不允许`, tip)
				continue
			}
			Component.TPlaywright.AddTipMsg(page, tip)
			h.log.Debugf(`点击 %s 允许`, tip)
			elementOp := &_struct.ElementOp{
				Type: define.ElementClick,
			}
			_, elementErr := h.DoLocator(&page, locator, elementOp)
			if elementErr != nil {
				h.callRun(runParams, cmdType, elementErr.Error(), tip, locator)
				return elementErr
			} else {
				h.callRun(runParams, cmdType, ``, tip, locator)
			}
		case define.Input: //输入
			if !h.allowCheckKey(checkKey, boolResultMap) {
				continue
			}
			Component.TPlaywright.AddTipMsg(page, tip)
			inputValue := cast.ToString(processVal[`value`])
			inputValue = gstool.SReplaces(inputValue, map[string]string{
				`{user_name}`: runParams.UserName,
				`{password}`:  runParams.Password,
				`{rand}`:      Component.TBase.GetUnique(`input_rand_`),
			})
			//针对输入进行替换
			for _, replaceVal := range runParams.ReplaceList {
				inputValue = gstool.SReplaces(inputValue, replaceVal)
			}
			elementOp := &_struct.ElementOp{
				Type:      define.ElementInput,
				FillValue: inputValue,
			}
			_, elementErr := h.DoLocator(&page, locator, elementOp)
			if elementErr != nil {
				h.callRun(runParams, cmdType, elementErr.Error(), tip, locator)
				return errors.New(`无法找到元素` + locator)
			}
			h.callRun(runParams, cmdType, ``, tip, inputValue)
		case define.RedirectUri: //跳转 保持当前域名
			//链接
			redirectUri := cast.ToString(processVal[`value`])
			redirectUri = gstool.SReplaces(redirectUri, map[string]string{
				`{domain}`: runParams.Domain,
			})
			if !h.allowCheckKey(checkKey, boolResultMap) {
				continue
			}
			Component.TPlaywright.AddTipMsg(page, tip)
			currentURL := page.URL()
			parsedURL, err := url.Parse(currentURL)
			if err != nil {
				h.callRun(runParams, cmdType, `解析失败，`+err.Error(), tip, currentURL)
				continue
			}
			domain := parsedURL.Scheme + `://` + parsedURL.Host
			targetUrl := domain + redirectUri
			time.Sleep(time.Second)
			if _, goErr := page.Goto(targetUrl); goErr != nil {
				h.callRun(runParams, cmdType, goErr.Error(), tip, targetUrl)
				return goErr
			} else {
				h.callRun(runParams, cmdType, ``, tip, targetUrl)
			}
		}
	}
	return nil
}

func (h *TPlaywright) parseLocator(Locator string) *_struct.Locator {
	sList := strings.Split(Locator, `|`)
	locator := _struct.Locator{
		Locator: sList[0],
		First:   false,
	}
	if gstool.ArrayExistValue(&sList, `first`) {
		locator.First = true
	}
	if strings.HasPrefix(locator.Locator, `!`) {
		locator.ExistSetNot = true
		locator.Locator = strings.TrimLeft(locator.Locator, `!`)
	}
	return &locator
}

func (h *TPlaywright) callRun(runParams *_struct.PlaywrightRunParams, cmdType define.CmdType, errmsg, tip, content string) {
	if runParams.RunCallFunc != nil {
		runParams.RunCallFunc(cmdType, errmsg, tip, content)
	}
}

// 是否允许执行 用于判断  {login_user}!={user_name}  或者 某个
func (h *TPlaywright) allowCheckKey(checkKey string, boolResult map[string]bool) bool {
	if checkKey == `` {
		return true
	}
	h.log.Debugf(`判断开始--- %s`, checkKey)
	checkList := strings.Split(checkKey, `&&`)
	h.log.Debugf(`判断列表 %s`, gstool.JsonEncode(checkList))
	for _, checkKeyVal := range checkList {

		if strings.HasPrefix(checkKeyVal, `!`) { //不等于时 等于了 那么跳过
			if boolResult[checkKeyVal[1:]] == true {
				h.log.Debugf(`判断1 %s 不允许 %s %t`, checkKeyVal, checkKeyVal[1:], boolResult[checkKeyVal[1:]])
				return false
			}
		} else if !boolResult[checkKeyVal] { //等于时  不等于了 那么跳过
			h.log.Debugf(`判断2 %s 不允许`, checkKeyVal)
			return false
		}
	}
	return true
}

// 提取outKey 返回bool
func (h *TPlaywright) outKeyBoolResult(outKey, checkKey string, boolResultMap map[string]bool, takeContent map[string]string, runParams *_struct.PlaywrightRunParams) {
	if outKey == `` {
		return
	}
	h.log.Debugf(`开始提取outKey %s %s`, outKey, checkKey)
	if strings.Contains(checkKey, `!=`) { //不等于
		checkList := strings.Split(checkKey, `!=`)
		if len(checkList) != 2 {
			return
		}
		leftCheck := strings.TrimSpace(checkList[0])
		rightCheck := strings.TrimSpace(checkList[1])
		if leftCheck != rightCheck {
			boolResultMap[outKey] = true
		} else {
			boolResultMap[outKey] = false
		}
		h.log.Debugf(`结果判断 %t`, boolResultMap[outKey])
	} else if strings.Contains(checkKey, `==`) { //等于
		checkList := strings.Split(checkKey, `==`)
		leftCheck := strings.TrimSpace(checkList[0])
		rightCheck := strings.TrimSpace(checkList[1])
		if leftCheck != rightCheck {
			boolResultMap[outKey] = false
		} else {
			boolResultMap[outKey] = true
		}
	}
}

// PageEvents 用来控制一段时间内不使用浏览器后自动关闭
func (h *TPlaywright) PageEvents(runParams *_struct.PlaywrightRunParams, page playwright.Page) {
	page.On("request", func(request playwright.Request) {
		go h.SetPageActive(page, runParams)
		return
	})
	for listenUri, listen := range h.ListenUrlList {
		listen.Callback(`注册 **`+listenUri, nil)
		h.log.Debugf(`新打开页面 注册请求 %s`, listenUri)
		_ = page.Route("**"+listenUri, func(route playwright.Route) {
			listen.Callback(`捕获到请求`+route.Request().URL(), nil)
			go h.ListenUrl(route, listen)
			_ = route.Abort()
		})
	}

	page.On(`load`, func() {
		go h.ShowCookieTip(page)
	})

	//可以监听到 前端下载
	page.On(`download`, func(download playwright.Download) {
		go h.SetPageActive(page, runParams)
		h.log.Debugf(`下载 %#v`, download)
		go h.AddTipMsg(page, `检测到下载`+download.SuggestedFilename()+`,别急，自动打开中..`)
		localPath := h.downloadPath + `/` + Component.TBase.GetUnique(`download`) + `_` + download.SuggestedFilename()
		h.log.Debugf(`localPath %s`, localPath)
		h.log.Debugf(download.String())
		go func() {
			//这个会一直阻塞
			_ = download.SaveAs(localPath)
		}()
		go func() {
			for {
				time.Sleep(time.Millisecond * 100)
				if gstool.FileIsExisted(localPath) {
					time.Sleep(time.Millisecond * 100)
					_ = download.Cancel()
					go h.AddTipMsg(page, `开始打开`+download.SuggestedFilename())
					openErr := Component.TOs.OpenFileWindows(localPath, localPath)
					if openErr != nil {
						h.log.Debugf(`打开文件失败 %s %s`, localPath, openErr.Error())
					} else {
						h.log.Errof(`打开文件成功 %s`, localPath)
					}
					return
				}
			}
		}()
		return
	})
}

func (h *TPlaywright) ListenUrl(route playwright.Route, listen *_struct.ListenUrl) {
	// 获取原始请求信息
	originalRequest := route.Request()
	requestUrl := originalRequest.URL()
	postData, _ := originalRequest.PostData()
	headers := originalRequest.Headers()
	cli := gshttp.PostJson(requestUrl).
		BodyStr(postData).
		Headers(headers)
	var res []byte
	var resErr error
	listen.StartCallBack(requestUrl)
	if listen.IsSse {
		res, resErr = cli.OpenStreamBytesEnd([]byte("\n\n"), func(s string, err error) {
			listen.Callback(s, err)
		}, func(bytes []byte) []byte {
			return bytes
		}).Request(200).Result()
	} else {
		res, resErr = cli.Request(200).Result()
		if resErr == nil {
			listen.Callback(cast.ToString(res), nil)
		}
	}
	if resErr != nil {
		listen.EndCallBack(resErr.Error())
	} else {
		listen.EndCallBack(`请求完成`)
	}
}

func (h *TPlaywright) SetPageActive(page playwright.Page, runParams *_struct.PlaywrightRunParams) {
	h.EventLock.Lock()
	defer h.EventLock.Unlock()
	if runParams.AutoCloseSecond == 0 {
		return
	}
	h.pageActiveTime[page.URL()] = PageActiveTime{
		ActiveTime: time.Now(),
		RunParams:  runParams,
		Page:       page,
	}
	h.log.Debugf(`page active %s %s`, page.URL(), gstool.TimeNowUnixToString(`Y-m-d H:i:s`))
}

func (h *TPlaywright) TimerCheckClosePage() {
	for {
		time.Sleep(time.Second)
		h.EventLock.Lock()
		newMap := make(map[string]PageActiveTime)
		for pageUrl, pageActiveTime := range h.pageActiveTime {
			if pageActiveTime.ActiveTime.Add(time.Second * time.Duration(pageActiveTime.RunParams.AutoCloseSecond)).Before(time.Now()) {
				h.log.Debugf(`自动关闭页面 %s 设置的活跃时间 %d 上次活跃时间 %s 当前时间 %s`,
					pageUrl, pageActiveTime.RunParams.AutoCloseSecond,
					gstool.TimeUnixToString(pageActiveTime.ActiveTime, `Y-md H:i:s`),
					gstool.TimeNowUnixToString(`Y-m-d H:i:s`))
				go func() {
					_ = pageActiveTime.Page.Close()
				}()
			} else {
				newMap[pageUrl] = pageActiveTime
			}
		}
		h.pageActiveTime = newMap
		h.EventLock.Unlock()
	}
}

// SmartLinkPlaywrightVersion 获取浏览器核心版本
func (h *TPlaywright) SmartLinkPlaywrightVersion() (*playwright.PlaywrightDriver, error) {
	return playwright.NewDriver()
}

func (h *TPlaywright) DoLocator(page *playwright.Page, Locators string, elementOp *_struct.ElementOp) (playwright.Locator, error) {
	list := strings.Split(Locators, `&&`) //多个用&&分割
	task := gstask.NewTask()
	waitSecond := playwright.Float(3000)
	for _, Locator := range list {
		h.log.Debugf(`解析 locator %s`, Locator)
		locator := h.parseLocator(Locator)
		//查找
		task.Add(gstask.CallbackFunc{
			Func: func() gstask.Result {
				selectorLoader := (*page).Locator(locator.Locator)
				if locator.First { //首个
					selectorLoader = selectorLoader.First()
				}
				selectorLoaderWaitErr := selectorLoader.WaitFor(playwright.LocatorWaitForOptions{
					Timeout: waitSecond,
				})
				//如果是反找Locator
				if locator.ExistSetNot {
					if selectorLoaderWaitErr != nil {
						h.log.Debugf(`反查找 %s 失败`, locator.Locator)
						return gstask.Result{
							Result: selectorLoader,
							Err:    errors.New(`找到了反找元素，返回失败`),
						}
					} else {
						h.log.Debugf(`反查找 %s 成功`, locator.Locator)
						return gstask.Result{
							Result: selectorLoader,
							Err:    errors.New(`等待反找元素超时，同样返回失败`),
						}
					}
				} else {
					if selectorLoaderWaitErr != nil {
						h.log.Debugf(`查找 %s 失败`, locator.Locator)
						return gstask.Result{
							Result: nil,
							Err:    errors.New(`没有找到元素 ` + locator.Locator),
						}
					} else {
						h.log.Debugf(`查找 %s 成功`, locator.Locator)
						return gstask.Result{
							Result: selectorLoader,
							Err:    nil,
						}
					}
				}

			},
			Timeout: 5 * time.Second,
		})
	}
	result := task.RunOne()
	if result.Err != nil {
		return nil, result.Err
	}
	element := result.Result.(playwright.Locator)
	switch elementOp.Type {
	case define.ElementInput:
		fillErr := element.Fill(elementOp.FillValue)
		return element, fillErr
	case define.ElementExist:
		return element, nil
	case define.ElementClick:
		clickErr := element.Click()
		return element, clickErr
	case define.ElementTextContent:
		content, textContentErr := element.TextContent()
		elementOp.TextContent = content
		return element, textContentErr
	case define.ElementCount:
		count, numErr := element.Count()
		elementOp.Count = count
		return element, numErr
	default:
		return nil, errors.New(`不支持的操作`)
	}
}

func (h *TPlaywright) SmartLinkRecycle() error {
	h.log.Debugf(`准备重置..`)
	if h.IsRun {
		return gstool.Error(`正在打开中`)
	}
	_ = h.Pw.Stop()
	h.ContextList = make([]ContextPage, 0)
	h.pageActiveTime = make(map[string]PageActiveTime)
	h.InitPlaywright()
	return nil
}

func (h *TPlaywright) SmartLinkDownloadPath() error {
	return Component.TOs.OpenDirWindows(gstool.DirPathFormatToWindows(h.downloadPath))
}
