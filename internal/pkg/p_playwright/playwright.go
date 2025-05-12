package p_playwright

import (
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"net/url"
	"strings"
	"sync"
	"time"
)

type Playwright struct {
	RunParams       *_struct.PlaywrightRunParams //运行时参数
	EventLock       sync.Mutex                   //事件锁
	TakeContentMap  map[string]string            //提取内容
	BoolResultMap   map[string]bool              //判断结果
	ContextPageList *ContextPageList             //浏览器上下文列表
	log             *gstool.GsSlog
}

func NewPlaywright(runParams *_struct.PlaywrightRunParams, log *gstool.GsSlog) *Playwright {
	return &Playwright{
		RunParams:       runParams,
		TakeContentMap:  make(map[string]string),
		BoolResultMap:   make(map[string]bool),
		ContextPageList: NewContextList(log),
		log:             log,
	}
}

func (h *Playwright) Open() error {
	if base.Component.TPlaywright.Pw == nil {
		return errors.New(`未启动浏览器核心`)
	}
	h.log.Debugf(`开始获取page`)
	page, pageErr := h.GetPage()
	if pageErr != nil {
		return gstool.Error(`获取page失败 %s`, pageErr.Error())
	}
	//输出结果存储
	h.log.Debugf(`开始处理process list`)
	for _, processVal := range h.RunParams.ProcessList {
		h.RunParams.ReplaceList = append(h.RunParams.ReplaceList, h.TakeContentMap)
		process := NewProcess(processVal, page, h.RunParams, h.BoolResultMap, h.TakeContentMap, h.log)
		code, reason, err := process.Do()
		h.log.Debugf(`执行结果 %s `, gstool.JsonFormat(map[string]any{
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
	h.log.Debugf(`获取context结束 %v`, contextErr)
	if contextErr != nil {
		return nil, contextErr
	}

	var page playwright.Page
	var pageErr error
	page, pageErr = (*contextPage.Context).NewPage()
	h.log.Debugf(`创建page结束 %v`, pageErr)
	if pageErr != nil {
		return nil, pageErr
	}
	//记录登录记录
	h.LastUserDataIndex(h.RunParams, contextPage.UserDataIndex)
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

func (h *Playwright) LastUserDataIndex(runParams *_struct.PlaywrightRunParams, userDataIndex int) {
	gstool.FmtPrintlnLogTime(`userName %s %d %d`, runParams.UserName, runParams.Id, userDataIndex)
	if runParams.UserName == `` || runParams.Id == 0 {
		return
	}
	sql := `select * from tbl_smart_link_last where  smart_link_id = ? and user_name = ? `
	smartLinkLast, smartLinkErr := base.Component.TSqlite.Client.QueryBySql(sql, runParams.Id, runParams.UserName).One()
	if smartLinkErr != nil {
		gstool.FmtPrintlnLogTime(`查询失败 %s`, smartLinkErr.Error())
		return
	} else if len(smartLinkLast) > 0 {
		_, err := base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link_last`, map[string]any{
			`smart_link_id`: runParams.Id,
			`user_name`:     runParams.UserName,
		}, map[string]any{
			`user_data_index`: userDataIndex,
			`update_time`:     time.Now().Unix(),
		}).Exec()
		if err != nil {
			gstool.FmtPrintlnLogTime(`更新最后使用索引失败 %s`, err.Error())
		}
	} else {
		_, err := base.Component.TSqlite.Client.QuickCreate(`tbl_smart_link_last`, map[string]any{
			`smart_link_id`:   runParams.Id,
			`user_name`:       runParams.UserName,
			`user_data_index`: userDataIndex,
			`create_time`:     time.Now().Unix(),
			`update_time`:     time.Now().Unix(),
		}).Exec()
		if err != nil {
			gstool.FmtPrintlnLogTime(`创建最后使用索引失败 %s`, err.Error())
		}
	}
}

func (h *Playwright) GetContextNotSaveUserData(browser playwright.Browser) (*ContextPage, error) {
	//查找可用的context
	rContext := h.ContextPageList.FindContextList(func(context *ContextPage) *ContextPage {
		//不保存数据过滤
		if context.IsSaveUserData {
			return nil
		}
		//非同种类型的context跳过
		if !base.Component.TPlaywright.IsSameLink(context.SmartLinkUniqueKey, h.RunParams.SmartLinkUniqueKey) {
			return nil
		}
		//找到一个context没有当前域名的
		existSameDomain := false
		pageList := (*context.Context).Pages()
		for _, v1 := range pageList {
			if gstool.UrlGetHost(v1.URL()) == h.RunParams.Domain {
				existSameDomain = true
				break
			}
		}
		//h.RunParams.CombineType != define.CombineTypeNo
		if !existSameDomain {
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
	contentPage := NewContextPage(&context, h.RunParams, ``, 0, h.log, closeEvent)
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
	h.log.Debugf(`未找到context，context使用的数据目录 %s`, userDataPath)
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
			h.log.Errof(`启动context报错 %s`, contextErr.Error())
			return nil, false, contextErr
		}
	} else {
		h.log.Debugf(`启动context 超时时间：%f`, h.RunParams.GetPageTimeout)
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
			h.log.Errof(`启动context报错 %s`, contextErr.Error())
			return nil, false, contextErr
		}
		h.log.Debugf(`启动 over`)
	}
	closeEvent := func() {
		h.log.Debugf(`context关闭`)
		h.ContextPageList.CleanContextList(false)
	}
	contextPage := NewContextPage(&context, h.RunParams, userDataPath, userDataIndex, h.log, closeEvent)
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
	//固定索引目录
	if h.RunParams.CombineType == define.CombineTypeFix {
		return h.RunParams.Id
	}
	//不需要合并 找到一个没有用到的就行
	if h.RunParams.CombineType == define.CombineTypeNo {
		noUserDataIndex := h.GetNoUserDataIndex()
		if noUserDataIndex != 0 {
			return noUserDataIndex
		} else {
			return 99 //找不到都给到99吧
		}
	}
	//自动找到上一次登录的目录索引
	if h.RunParams.CombineType == define.CombineTypeLast {
		lastUserDataIndex := h.GetLastUserDataIndex()
		if lastUserDataIndex != 0 {
			return lastUserDataIndex
		}
	}
	//需要合并 找一下可以重复利用的index
	findUserDataIndex := h.GetFindUserDataIndex()
	if findUserDataIndex != 0 {
		return findUserDataIndex
	}
	return 99 //错误
}

func (h *Playwright) IsSameLink(smartLinkUniqueKeyS, smartLinkUniqueKeyT string) bool {
	return strings.Split(smartLinkUniqueKeyS, `_`)[0] == strings.Split(smartLinkUniqueKeyT, `_`)[0]
}

// CleanContextPagesFixDataId 根据域名清理context
// 注意；这里直接关闭context 防止context假死 导致登录态不能登录
func (h *Playwright) CleanContextPagesFixDataId() {
	if h.RunParams.CombineType != define.CombineTypeFix {
		return
	}
	h.ContextPageList.EachContextList(func(context *ContextPage) bool {
		if context.ContextUnique == h.RunParams.ContextUnique {
			context.CloseContextPages()
		}
		return false
	})
}

func (h *Playwright) GetNoUserDataIndex() int {
	for i := 1; i < define.MaxUserDataIndex; i++ {
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
	return 0
}

func (h *Playwright) GetLastUserDataIndex() int {
	if h.RunParams.UserName == `` {
		return 0
	}
	sql := `select * from tbl_smart_link_last where  smart_link_id = ? and user_name = ? `
	smartLinkLast, smartLinkErr := base.Component.TSqlite.Client.QueryBySql(sql, h.RunParams.Id, h.RunParams.UserName).One()
	if smartLinkErr != nil {
		return 0
	} else {
		return cast.ToInt(smartLinkLast[`user_data_index`])
	}
}

func (h *Playwright) GetFindUserDataIndex() int {
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
			h.log.Debugf(`递增目录，找到了已经存在的context %s`, context.ContextUnique)
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
	for i := 1; i < define.MaxUserDataIndex; i++ {
		if !gstool.ArrayExistValue(&ignoreIndexList, i) {
			return i
		}
	}
	return 99
}

func (h *Playwright) Recycle() error {
	h.log.Debugf(`开始回收..`)
	_ = base.Component.TPlaywright.Pw.Stop()
	h.ContextPageList.CleanContextList(true)
	base.Component.TPlaywright.InitPlaywright()
	InitPageActiveTime()
	return nil
}
