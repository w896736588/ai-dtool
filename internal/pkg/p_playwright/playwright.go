package p_playwright

import (
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
	"net/url"
	"strings"
	"sync"
)

type Playwright struct {
	RunParams       *_struct.PlaywrightRunParams //运行时参数
	EventLock       sync.Mutex                   //事件锁
	TakeContentMap  map[string]string            //提取内容
	BoolResultMap   map[string]bool              //判断结果
	ContextPageList *ContextPageList             //浏览器上下文列表
}

func NewPlaywright(runParams *_struct.PlaywrightRunParams, log *gstool.GsSlog) *Playwright {
	return &Playwright{
		RunParams:       runParams,
		TakeContentMap:  make(map[string]string),
		BoolResultMap:   make(map[string]bool),
		ContextPageList: NewContextList(log),
	}
}

func (h *Playwright) Open() error {
	if base.Component.TPlaywright.Pw == nil {
		return errors.New(`未启动浏览器核心`)
	}
	base.Component.TPlaywright.Log.Debugf(`开始获取page`)
	page, pageErr := h.GetPage()
	if pageErr != nil {
		return gstool.Error(`获取page失败 %s`, pageErr.Error())
	}
	//输出结果存储
	base.Component.TPlaywright.Log.Debugf(`开始处理process list`)
	for _, processVal := range h.RunParams.ProcessList {
		h.RunParams.ReplaceList = append(h.RunParams.ReplaceList, h.TakeContentMap)
		process := NewProcess(processVal, page, h.RunParams, h.BoolResultMap, h.TakeContentMap)
		code, reason, err := process.Do()
		base.Component.TPlaywright.Log.Debugf(`执行结果 %s `, gstool.JsonFormat(map[string]any{
			`type`:           process.ProcessType,
			`reason`:         reason,
			`domain`:         h.RunParams.Domain,
			`domain_limit`:   process.DomainLimit,
			`Locator`:        process.Locator,
			`tip`:            process.Tip,
			`code`:           code,
			`Checks`:         process.Checks,
			`TakeContextMap`: h.TakeContentMap,
			`BoolResultMap`:  h.BoolResultMap,
		}))
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Playwright) GetPage() (*playwright.Page, error) {
	var contextErr error
	var contextPage *ContextPage
	boolCleanFirstBlank := false
	if !h.RunParams.IsSaveUserData { //不保存用户数据
		browser, browserErr := h.GetBrowser()
		if browserErr != nil {
			return nil, browserErr
		}
		contextPage, contextErr = h.GetContextNotSaveUserData(browser)
	} else { //保留用户数据
		contextPage, boolCleanFirstBlank, contextErr = h.GetContextSaveUserData()
	}
	base.Component.TPlaywright.Log.Debugf(`获取context结束 %v`, contextErr)
	if contextErr != nil {
		return nil, contextErr
	}

	var page playwright.Page
	var pageErr error
	page, pageErr = (*contextPage.Context).NewPage()
	base.Component.TPlaywright.Log.Debugf(`创建page结束 %v`, pageErr)
	if pageErr != nil {
		return nil, pageErr
	}
	// 关闭一个blank
	if boolCleanFirstBlank {
		contextPage.CloseFirstPage()
	}
	//跳转链接
	u, _ := url.Parse(h.RunParams.Link)
	if _, goErr := page.Goto(u.String()); goErr != nil {
		return nil, goErr
	}
	//等待加载完成
	base.Component.TPlaywright.WaitForLoadState(&page, h.RunParams.LocatorTimeout)
	return &page, nil
}

func (h *Playwright) GetContextNotSaveUserData(browser playwright.Browser) (*ContextPage, error) {
	//查找可用的context
	rContext := h.ContextPageList.FindContextList(func(context *ContextPage) *ContextPage {
		//非同种类型的context跳过
		if !base.Component.TPlaywright.IsSameLink(context.SmartLinkUniqueKey, h.RunParams.SmartLinkUniqueKey) {
			return nil
		}
		//找到一个context没有当前域名的
		boolFind := false
		pageList := (*context.Context).Pages()
		for _, v1 := range pageList {
			if gstool.UrlGetHost(v1.URL()) == h.RunParams.Domain {
				boolFind = true
				break
			}
		}
		if !boolFind && h.RunParams.IsCombine {
			return context
		}
		return nil
	})
	if rContext != nil {
		return rContext, nil
	}
	var context playwright.BrowserContext
	var contextErr error
	if h.RunParams.BrowserAuthUsername != `` && h.RunParams.BrowserAuthPassword != `` {
		context, contextErr = browser.NewContext(playwright.BrowserNewContextOptions{
			HttpCredentials: &playwright.HttpCredentials{
				Username: h.RunParams.BrowserAuthUsername,
				Password: h.RunParams.BrowserAuthPassword,
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
		return nil, contextErr
	}
	closeEvent := func() {
		h.ContextPageList.CleanContextList(false)
	}
	contentPage := NewContextPage(&context, h.RunParams, ``, 0, base.Component.TPlaywright.Log, closeEvent)
	h.ContextPageList.AddContextList(contentPage)
	return contentPage, nil
}

func (h *Playwright) GetBrowser() (playwright.Browser, error) {
	if h.RunParams.OpenType == define.OpenTypeWebkitSilence && base.Component.TPlaywright.BrowserWebkitSilence != nil {
		return base.Component.TPlaywright.BrowserWebkitSilence, nil
	} else if h.RunParams.OpenType == define.OpenTypeWebkitChrome && base.Component.TPlaywright.BrowserWebkitChrome != nil {
		return base.Component.TPlaywright.BrowserWebkitChrome, nil
	}
	var browserErr error
	if h.RunParams.OpenType == define.OpenTypeWebkitSilence {
		base.Component.TPlaywright.BrowserWebkitSilence, browserErr = base.Component.TPlaywright.Pw.Chromium.Launch()
		if browserErr != nil {
			base.Component.TPlaywright.BrowserWebkitSilence = nil
			return nil, browserErr
		} else {
			return base.Component.TPlaywright.BrowserWebkitSilence, nil
		}
	} else {
		base.Component.TPlaywright.BrowserWebkitChrome, browserErr = base.Component.TPlaywright.Pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			//DownloadsPath: &h.downloadPath,
			Headless: playwright.Bool(false), //有界面模式
		})
		if browserErr != nil {
			base.Component.TPlaywright.BrowserWebkitChrome = nil
			return nil, browserErr
		} else {
			return base.Component.TPlaywright.BrowserWebkitChrome, nil
		}
	}
}

// GetContextSaveUserData 获取context 需要保存用户数据
func (h *Playwright) GetContextSaveUserData() (*ContextPage, bool, error) {
	//固定打开数据索引 关闭此context下面的所有页面
	h.CleanContextPagesFixDataId()
	//获取数据索引目录
	userDataIndex := h.GetUserDataIndex()
	//通过索引目录拿到已存在的context
	existContextPage := h.GetContextByIndex(userDataIndex)
	if existContextPage != nil {
		return existContextPage, false, nil
	}
	userDataPath := fmt.Sprintf(base.Component.Env.WebkitDataPath+`/%d`, userDataIndex)
	base.Component.TPlaywright.Log.Debugf(`未找到context，context使用的数据目录 %s`, userDataPath)
	//创建数据索引目录
	_ = gstool.DirCreatePath(userDataPath)
	//打开模式
	Headless := false
	if h.RunParams.OpenType == define.OpenTypeWebkitSilence {
		Headless = true
	}
	var context playwright.BrowserContext
	var contextErr error
	//浏览器自带验证
	if h.RunParams.BrowserAuthUsername != `` && h.RunParams.BrowserAuthPassword != `` {
		context, contextErr = base.Component.TPlaywright.Pw.Chromium.LaunchPersistentContext(userDataPath, playwright.BrowserTypeLaunchPersistentContextOptions{
			//DownloadsPath:     &h.downloadPath,
			Headless:          &Headless,
			Channel:           playwright.String(h.RunParams.Channel), // 使用完整版 Chrome 而非 Chromium
			NoViewport:        playwright.Bool(true),
			JavaScriptEnabled: playwright.Bool(true),
			AcceptDownloads:   playwright.Bool(true),
			Locale:            playwright.String(`zh-CN`),
			Timeout:           &h.RunParams.GetPageTimeout,
			IgnoreHttpsErrors: playwright.Bool(true),
			HttpCredentials: &playwright.HttpCredentials{
				Username: h.RunParams.BrowserAuthUsername,
				Password: h.RunParams.BrowserAuthPassword,
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
			base.Component.TPlaywright.Log.Errof(`启动context报错 %s`, contextErr.Error())
			return nil, false, contextErr
		}
	} else {
		base.Component.TPlaywright.Log.Debugf(`启动context 超时时间：%f`, h.RunParams.GetPageTimeout)
		context, contextErr = base.Component.TPlaywright.Pw.Chromium.LaunchPersistentContext(userDataPath, playwright.BrowserTypeLaunchPersistentContextOptions{
			//DownloadsPath:     &h.downloadPath,
			Headless: &Headless,
			//Channel:           playwright.String(runParams.Channel),//增加这个会导致问题 关闭后不能正常启动下一个
			NoViewport:        playwright.Bool(true),
			JavaScriptEnabled: playwright.Bool(true),
			AcceptDownloads:   playwright.Bool(true),
			IgnoreHttpsErrors: playwright.Bool(true),
			Locale:            playwright.String(`zh-CN`),
			Timeout:           &h.RunParams.GetPageTimeout,
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
			base.Component.TPlaywright.Log.Errof(`启动context报错 %s`, contextErr.Error())
			return nil, false, contextErr
		}
		base.Component.TPlaywright.Log.Debugf(`启动 over`)
	}
	closeEvent := func() {
		base.Component.TPlaywright.Log.Debugf(`context关闭`)
		h.ContextPageList.CleanContextList(false)
	}
	contextPage := NewContextPage(&context, h.RunParams, userDataPath, userDataIndex, base.Component.TPlaywright.Log, closeEvent)
	h.ContextPageList.AddContextList(contextPage)
	return contextPage, true, nil
}

func (h *Playwright) GetContextByIndex(dataIndex int) *ContextPage {
	return h.ContextPageList.FindContextList(func(context *ContextPage) *ContextPage {
		if context.UserDataIndex == dataIndex {
			return context
		}
		return nil
	})
}

func (h *Playwright) GetUserDataIndex() int {
	maxIndex := 500
	//固定索引目录
	if h.RunParams.FixDataId == 1 {
		return h.RunParams.Id
	}
	//不需要合并 找到一个没有用到的就行
	if !h.RunParams.IsCombine {
		for i := 0; i < maxIndex; i++ {
			boolExist := false
			h.ContextPageList.EachContextList(func(context *ContextPage) bool {
				if context.UserDataIndex == i {
					boolExist = true
					return true
				}
				return false
			})
			if !boolExist {
				return i
			}
		}
	}
	//需要合并 找一下可以重复利用的index
	ignoreIndexList := make([]int, 0)
	rContext := h.ContextPageList.FindContextList(func(context *ContextPage) *ContextPage {
		//非同一类型打开方式 不管
		if context.OpenType != h.RunParams.OpenType {
			ignoreIndexList = append(ignoreIndexList, context.UserDataIndex)
			return nil
		}
		//非同一类型的链接 不管
		if !h.IsSameLink(context.SmartLinkUniqueKey, h.RunParams.SmartLinkUniqueKey) {
			ignoreIndexList = append(ignoreIndexList, context.UserDataIndex)
			return nil
		}
		//是否有相同域名的page存在
		boolFindSameDomainPage := false
		pageList := (*context.Context).Pages()
		for _, page := range pageList {
			if gstool.UrlGetHost(page.URL()) == h.RunParams.Domain {
				boolFindSameDomainPage = true
				break
			}
		}
		//没有找到相同域名的page
		if !boolFindSameDomainPage { //需要合并时才处理
			base.Component.TPlaywright.Log.Debugf(`递增目录，找到了已经存在的context %s`, context.ContextUnique)
			return context
		} else {
			ignoreIndexList = append(ignoreIndexList, context.UserDataIndex)
		}
		return nil
	})
	if rContext != nil {
		return rContext.UserDataIndex
	}
	//没有能够复用的数据索引 那么
	for i := 0; i < maxIndex; i++ {
		if !gstool.ArrayExistValue(&ignoreIndexList, i) {
			return i
		}
	}
	return -1 //错误
}

func (h *Playwright) IsSameLink(smartLinkUniqueKeyS, smartLinkUniqueKeyT string) bool {
	return strings.Split(smartLinkUniqueKeyS, `_`)[0] == strings.Split(smartLinkUniqueKeyT, `_`)[0]
}

// CleanContextPagesFixDataId 根据域名清理context
// 注意；这里直接关闭context 防止context假死 导致登录态不能登录
func (h *Playwright) CleanContextPagesFixDataId() {
	if h.RunParams.FixDataId == 0 {
		return
	}
	h.ContextPageList.EachContextList(func(context *ContextPage) bool {
		if context.ContextUnique == h.RunParams.ContextUnique {
			context.CloseContextPages()
		}
		return false
	})
}

func (h *Playwright) Recycle() error {
	base.Component.TPlaywright.Log.Debugf(`开始回收..`)
	_ = base.Component.TPlaywright.Pw.Stop()
	h.ContextPageList.CleanContextList(true)
	base.Component.TPlaywright.InitPlaywright()
	InitPageActiveTime()
	return nil
}
