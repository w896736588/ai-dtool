package p_playwright

import (
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"strings"
	"sync"
	"time"
)

// list 所有浏览器列表
var list []*ContextPage

func InitContextPageList() {
	list = make([]*ContextPage, 0)
}

type ContextPageList struct {
	ContextLock sync.RWMutex
	log         *gstool.GsSlog
}

func NewContextList(log *gstool.GsSlog) *ContextPageList {
	return &ContextPageList{
		log: log,
	}
}

func (h *ContextPageList) EventContextClose(contextP *ContextPage) {
	go (*contextP.Context).OnClose(func(context playwright.BrowserContext) {
		contextP.RunParams.StreamFunc(`context关闭`, fmt.Sprintf(`%s %d %s`, contextP.ContextUnique, contextP.UserDataIndex, contextP.SmartLinkUniqueKey))
		h.CleanContextList(false)
	})
}

func (h *ContextPageList) AddContextList(contextP *ContextPage) {
	h.ContextLock.Lock()
	defer h.ContextLock.Unlock()
	list = append(list, contextP)
	h.EventContextClose(contextP)
}

func (h *ContextPageList) EachContextList(f func(context *ContextPage) bool) {
	h.ContextLock.Lock()
	defer h.ContextLock.Unlock()
	for _, context := range list {
		if f(context) {
			break
		}
	}
}

func (h *ContextPageList) FindContextList(f func(context *ContextPage) *ContextPage) *ContextPage {
	h.ContextLock.Lock()
	defer h.ContextLock.Unlock()
	for _, context := range list {
		rContext := f(context)
		if rContext != nil {
			return rContext
		}
	}
	return nil
}

func (h *ContextPageList) CleanContextList(cleanAll bool) {
	h.ContextLock.Lock()
	defer h.ContextLock.Unlock()
	if cleanAll {
		for _, context := range list {
			h.CloseContextPages(context.Context)
		}
		list = make([]*ContextPage, 0)
	} else {
		newContextList := make([]*ContextPage, 0)
		for _, context := range list {
			if context.Context != nil && len((*context.Context).Pages()) > 0 {
				newContextList = append(newContextList, context)
			}
		}
		list = newContextList
	}
}

func (h *ContextPageList) CloseContextPages(context *playwright.BrowserContext) {
	pageList := (*context).Pages()
	for _, page := range pageList {
		_ = page.Close()
	}
}

func (h *ContextPageList) GetPlaywrightRunList() []map[string]any {
	runList := make([]map[string]any, 0)
	h.EachContextList(func(context *ContextPage) bool {
		pageList := (*context.Context).Pages()
		runList = append(runList, map[string]any{
			`name`:     context.SmartLinkUniqueKey,
			`page_num`: len(pageList),
		})
		return false
	})
	return runList
}

// FindNotSaveUserDataContext 查找可用的 不需要保存数据的context
func (h *ContextPageList) FindNotSaveUserDataContext(runParams *_struct.PlaywrightRunParams) *ContextPage {
	return h.FindContextList(func(context *ContextPage) *ContextPage {
		//不保存数据过滤
		if context.CombineType != define.CombineTypeNo {
			runParams.StreamFunc(`context`, context.ContextUnique+`不是每次打开新的session，不处理`)
			return nil
		}
		//非同种类型的context跳过
		if !base.Component.TPlaywright.IsSameLink(context.SmartLinkUniqueKey, runParams.SmartLinkUniqueKey) {
			runParams.StreamFunc(`context`, context.ContextUnique+`不属于同一类型链接，不处理，`+context.SmartLinkUniqueKey)
			return nil
		}
		//找到一个context没有当前域名的
		existSameDomain := false
		pageList := (*context.Context).Pages()
		for _, v1 := range pageList {
			if gstool.UrlGetHost(v1.URL()) == runParams.Domain {
				existSameDomain = true
				break
			}
		}
		//h.RunParams.CombineType != define.CombineTypeNo
		if !existSameDomain {
			runParams.StreamFunc(`context`, context.ContextUnique+`没有当前域名的网页，可以打开`)
			return context
		}
		return nil
	})
}

func (h *ContextPageList) CleanContextPagesFixDataId(runParams *_struct.PlaywrightRunParams) {
	if runParams.CombineType != define.CombineTypeFix {
		return
	}
	runParams.StreamFunc(`context`, `固定目录，开始清理旧页面`)
	h.EachContextList(func(context *ContextPage) bool {
		if context.ContextUnique == runParams.ContextUnique {
			runParams.StreamFunc(`context`, `固定目录，开始清理context`+context.ContextUnique)
			context.CloseContextPages()
		}
		return false
	})
	time.Sleep(time.Second * 1)
}

func (h *ContextPageList) GetUserDataIndex(runParams *_struct.PlaywrightRunParams) int {
	combineNameMap := map[int]string{
		define.CombineTypeFind: `自动查找`,
		define.CombineTypeLast: `使用上次登录的`,
		define.CombineTypeNo:   `每次打开新的`,
		define.CombineTypeFix:  `固定目录`,
	}
	runParams.StreamFunc(`context`, fmt.Sprintf(`当前合并类型为 %s`, combineNameMap[runParams.CombineType]))
	//固定索引目录
	if runParams.CombineType == define.CombineTypeFix {
		userDataIndex := runParams.Id
		runParams.StreamFunc(`context`, `固定目录，以`+cast.ToString(userDataIndex)+`作为数据目录`)
		return userDataIndex
	}
	//不需要合并 找到一个没有用到的就行
	if runParams.CombineType == define.CombineTypeNo {
		noUserDataIndex := h.GetNoUserDataIndex()
		runParams.StreamFunc(`context`, `不需要合并，开始寻找一个没有用过的目录,找到`+cast.ToString(noUserDataIndex))
		if noUserDataIndex != 0 {
			return noUserDataIndex
		} else {
			return 99 //找不到都给到99吧
		}
	}
	//自动找到上一次登录的目录索引
	if runParams.CombineType == define.CombineTypeLast {
		lastUserDataIndex := h.GetLastUserDataIndex(runParams)
		runParams.StreamFunc(`context`, fmt.Sprintf(`根据上一次打开的来找到目录 %d`, lastUserDataIndex))
		if lastUserDataIndex != 0 {
			return lastUserDataIndex
		}
	}
	//需要合并 找一下可以重复利用的index
	findUserDataIndex := h.GetFindUserDataIndex(runParams)
	if findUserDataIndex != 0 {
		return findUserDataIndex
	}
	return 99 //错误
}

func (h *ContextPageList) GetLastUserDataIndex(runParams *_struct.PlaywrightRunParams) int {
	if runParams.LastIndexLabel == `` {
		return 0
	}
	sql := `select * from tbl_smart_link_last where user_name = ? and domain = ? `
	smartLinkLast, smartLinkErr := base.Component.TSqlite.Client.QueryBySql(sql, runParams.LastIndexLabel, runParams.Domain).One()
	if smartLinkErr != nil {
		return 0
	} else {
		return cast.ToInt(smartLinkLast[`user_data_index`])
	}
}

func (h *ContextPageList) GetFindUserDataIndex(runParams *_struct.PlaywrightRunParams) int {
	runParams.StreamFunc(`context`, `开始寻找可复用context`)
	ignoreIndexList := make([]int, 0)
	rContext := h.FindContextList(func(context *ContextPage) *ContextPage {
		//非同一类型打开方式 不管
		if context.OpenType != runParams.OpenType {
			runParams.StreamFunc(`context`, context.ContextUnique+`非同一类型打开方式，不处理`)
			ignoreIndexList = append(ignoreIndexList, context.UserDataIndex)
			return nil
		}
		//非同一类型的链接 不管
		if !h.IsSameLink(context.SmartLinkUniqueKey, runParams.SmartLinkUniqueKey) {
			runParams.StreamFunc(`context`, context.ContextUnique+`不属于同一类型链接，不处理，`+context.SmartLinkUniqueKey)
			ignoreIndexList = append(ignoreIndexList, context.UserDataIndex)
			return nil
		}
		//是否有相同域名的page存在
		boolFindSameDomainPage := false
		pageList := (*context.Context).Pages()
		for _, page := range pageList {
			if gstool.UrlGetHost(page.URL()) == runParams.Domain {
				runParams.StreamFunc(`context`, context.ContextUnique+`有当前域名的网页，不可以打开`)
				boolFindSameDomainPage = true
				break
			}
		}
		//没有找到相同域名的page
		if !boolFindSameDomainPage { //需要合并时才处理
			runParams.StreamFunc(`context`, fmt.Sprintf(`找到了已经存在的context %s`, context.ContextUnique))
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
		if gstool.ArrayExistValue(&ignoreIndexList, i) {
			runParams.StreamFunc(`context`, fmt.Sprintf(`挨个判断，是否有可以用的目录，%d该目录已处于忽略列表`, i))
			continue
		}
		//是否已存在相同域名在使用
		if h.ExistDomainUserDataIndex(i, runParams) {
			runParams.StreamFunc(`context`, fmt.Sprintf(`挨个判断，是否有可以用的目录，%d该目录下有相同域名`, i))
			continue
		}
		runParams.StreamFunc(`context`, fmt.Sprintf(`挨个判断，是否有可以用的目录，%d该目录可用`, i))
		return i
	}
	return 99
}

func (h *ContextPageList) ExistDomainUserDataIndex(userDataIndex int, runParams *_struct.PlaywrightRunParams) bool {
	sql := `select * from tbl_smart_link_last where domain = ? and user_data_index = ? `
	smartLinkLast, smartLinkErr := base.Component.TSqlite.Client.QueryBySql(sql, runParams.Domain, userDataIndex).One()
	if smartLinkErr != nil {
		return false
	} else if len(smartLinkLast) > 0 {
		return true
	} else {
		return false
	}
}

func (h *ContextPageList) IsSameLink(smartLinkUniqueKeyS, smartLinkUniqueKeyT string) bool {
	return strings.Split(smartLinkUniqueKeyS, `_`)[0] == strings.Split(smartLinkUniqueKeyT, `_`)[0]
}

func (h *ContextPageList) GetNoUserDataIndex() int {
	for i := 1; i < define.MaxUserDataIndex; i++ {
		boolExist := false
		h.EachContextList(func(context *ContextPage) bool {
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

func (h *ContextPageList) GetContextByIndex(dataIndex int) *ContextPage {
	return h.FindContextList(func(context *ContextPage) *ContextPage {
		if context.UserDataIndex == dataIndex {
			return context
		}
		return nil
	})
}

func (h *ContextPageList) GetContextParam(runParams *_struct.PlaywrightRunParams) (*ContextPage, int, string) {
	//固定打开数据索引 关闭此context下面的所有页面
	h.CleanContextPagesFixDataId(runParams)
	//获取数据索引目录
	userDataIndex := h.GetUserDataIndex(runParams)
	//通过索引目录拿到已存在的context
	existContextPage := h.GetContextByIndex(userDataIndex)
	if existContextPage != nil {
		runParams.StreamFunc(`context`, fmt.Sprintf(`已存在context %s ,直接使用`, existContextPage.ContextUnique))
		return existContextPage, existContextPage.UserDataIndex, existContextPage.UserDataPath
	}
	userDataPath := fmt.Sprintf(base.Component.Env.WebkitDataPath+`/%d`, userDataIndex)
	runParams.StreamFunc(`context`, fmt.Sprintf(`未找到已存在的context，context使用的数据目录 %s,开始创建context`, userDataPath))
	//创建数据索引目录
	_ = gstool.DirCreatePath(userDataPath)
	return nil, userDataIndex, userDataPath
}

// GetContextSaveUserData 获取context 需要保存用户数据
func (h *ContextPageList) GetContextSaveUserData(runParams *_struct.PlaywrightRunParams) (*ContextPage, bool, error) {
	runParams.StreamFunc(`处理session`, `需要保存用户数据 `+runParams.ContextUnique+` `+runParams.SmartLinkUniqueKey)
	existContextPage, userDataIndex, userDataPath := h.GetContextParam(runParams)
	if existContextPage != nil {
		runParams.StreamFunc(`处理session`, fmt.Sprintf(`已存在context %s ,直接使用%s`, existContextPage.ContextUnique, userDataPath))
		return existContextPage, false, nil
	}
	//打开模式
	Headless := false
	if runParams.OpenType == define.OpenTypeWebkitSilence {
		runParams.StreamFunc(`context`, `使用无头模式打开`)
		Headless = true
	} else {
		runParams.StreamFunc(`context`, `使用有头模式打开`)
	}
	var context playwright.BrowserContext
	var contextErr error
	//浏览器自带验证
	if runParams.BrowserAuthUsername != `` && runParams.BrowserAuthPassword != `` {
		runParams.StreamFunc(`context`, fmt.Sprintf(`打开contxt，使用浏览器自带验证 用户名%s,超时时间 %f`, runParams.BrowserAuthUsername, runParams.GetPageTimeout))
		context, contextErr = base.Component.TPlaywright.Pw.Chromium.LaunchPersistentContext(userDataPath, playwright.BrowserTypeLaunchPersistentContextOptions{
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
			runParams.StreamFunc(`context`, fmt.Sprintf(`启动context报错 %s`, contextErr.Error()))
			return nil, false, contextErr
		}
	} else {
		runParams.StreamFunc(`context`, fmt.Sprintf(`启动context 超时时间：%f`, runParams.GetPageTimeout))
		context, contextErr = base.Component.TPlaywright.Pw.Chromium.LaunchPersistentContext(userDataPath, playwright.BrowserTypeLaunchPersistentContextOptions{
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
			runParams.StreamFunc(`context`, fmt.Sprintf(`启动context报错 %s`, contextErr.Error()))
			return nil, false, contextErr
		}
		runParams.StreamFunc(`context`, `启动完成`)
	}
	closeEvent := func() {
		runParams.StreamFunc(`context`, `context关闭 `+runParams.ContextUnique+` `+runParams.SmartLinkUniqueKey)
		h.CleanContextList(false)
	}
	contextPage := NewContextPage(&context, runParams, userDataPath, userDataIndex, h.log, closeEvent)
	h.AddContextList(contextPage)
	return contextPage, true, nil
}

func (h *ContextPageList) GetContextNotSaveUserData(browser playwright.Browser, runParams *_struct.PlaywrightRunParams) (*ContextPage, error) {
	//查找可用的context
	rContext := h.FindNotSaveUserDataContext(runParams)
	if rContext != nil {
		return rContext, nil
	}
	runParams.StreamFunc(`处理session`, fmt.Sprintf(`没有找到可用的session`))
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
		runParams.StreamFunc(`处理session`, fmt.Sprintf(`创建并传递浏览器验证 用户名%s`, runParams.BrowserAuthUsername))
	} else {
		context, contextErr = browser.NewContext(playwright.BrowserNewContextOptions{
			NoViewport:        playwright.Bool(true),
			JavaScriptEnabled: playwright.Bool(true),
			AcceptDownloads:   playwright.Bool(true),
			Locale:            playwright.String(`zh-CN`),
		})
		runParams.StreamFunc(`处理session`, fmt.Sprintf(`创建完全全新的session`))
	}
	if contextErr != nil {
		return nil, contextErr
	}
	closeEvent := func() {
		h.CleanContextList(false)
	}
	runParams.StreamFunc(`处理session`, fmt.Sprintf(`userDataPath %s userDataIndex %d`, ``, 0))
	contentPage := NewContextPage(&context, runParams, ``, 0, h.log, closeEvent)
	h.AddContextList(contentPage)
	return contentPage, nil
}
