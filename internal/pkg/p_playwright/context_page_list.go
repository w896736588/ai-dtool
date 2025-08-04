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
		h.log.Debugf(`context关闭 %s %d %s`, contextP.ContextUnique, contextP.UserDataIndex, contextP.SmartLinkUniqueKey)
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
			return nil
		}
		//非同种类型的context跳过
		if !base.Component.TPlaywright.IsSameLink(context.SmartLinkUniqueKey, runParams.SmartLinkUniqueKey) {
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
			return context
		}
		return nil
	})
}

func (h *ContextPageList) CleanContextPagesFixDataId(runParams *_struct.PlaywrightRunParams) {
	if runParams.CombineType != define.CombineTypeFix {
		return
	}
	h.EachContextList(func(context *ContextPage) bool {
		if context.ContextUnique == runParams.ContextUnique {
			context.CloseContextPages()
		}
		return false
	})
}

func (h *ContextPageList) GetUserDataIndex(runParams *_struct.PlaywrightRunParams) int {
	h.log.Debugf(`当前合并类型为 %d`, runParams.CombineType)
	//固定索引目录
	if runParams.CombineType == define.CombineTypeFix {
		return runParams.Id
	}
	//不需要合并 找到一个没有用到的就行
	if runParams.CombineType == define.CombineTypeNo {
		noUserDataIndex := h.GetNoUserDataIndex()
		if noUserDataIndex != 0 {
			return noUserDataIndex
		} else {
			return 99 //找不到都给到99吧
		}
	}
	//自动找到上一次登录的目录索引
	if runParams.CombineType == define.CombineTypeLast {
		lastUserDataIndex := h.GetLastUserDataIndex(runParams)
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
	ignoreIndexList := make([]int, 0)
	rContext := h.FindContextList(func(context *ContextPage) *ContextPage {
		//非同一类型打开方式 不管
		if context.OpenType != runParams.OpenType {
			ignoreIndexList = append(ignoreIndexList, context.UserDataIndex)
			return nil
		}
		//非同一类型的链接 不管
		if !h.IsSameLink(context.SmartLinkUniqueKey, runParams.SmartLinkUniqueKey) {
			ignoreIndexList = append(ignoreIndexList, context.UserDataIndex)
			return nil
		}
		//是否有相同域名的page存在
		boolFindSameDomainPage := false
		pageList := (*context.Context).Pages()
		for _, page := range pageList {
			if gstool.UrlGetHost(page.URL()) == runParams.Domain {
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
		if gstool.ArrayExistValue(&ignoreIndexList, i) {
			continue
		}
		//是否已存在相同域名在使用
		if h.ExistDomainUserDataIndex(i, runParams) {
			continue
		}
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
		return existContextPage, existContextPage.UserDataIndex, existContextPage.UserDataPath
	}
	userDataPath := fmt.Sprintf(base.Component.Env.WebkitDataPath+`/%d`, userDataIndex)
	h.log.Debugf(`未找到context，context使用的数据目录 %s`, userDataPath)
	//创建数据索引目录
	_ = gstool.DirCreatePath(userDataPath)
	return nil, userDataIndex, userDataPath
}

// GetContextSaveUserData 获取context 需要保存用户数据
func (h *ContextPageList) GetContextSaveUserData(runParams *_struct.PlaywrightRunParams) (*ContextPage, bool, error) {
	h.log.Debugf(`需要保存用户数据`)
	existContextPage, userDataIndex, userDataPath := h.GetContextParam(runParams)
	if existContextPage != nil {
		return existContextPage, false, nil
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
		h.log.Debugf(`打开contxt，使用用户名和密码%s %s`, runParams.BrowserAuthUsername, runParams.BrowserAuthPassword)
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
			h.log.Errof(`启动context报错 %s`, contextErr.Error())
			return nil, false, contextErr
		}
	} else {
		h.log.Debugf(`启动context 超时时间：%f`, runParams.GetPageTimeout)
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
			h.log.Errof(`启动context报错 %s`, contextErr.Error())
			return nil, false, contextErr
		}
		h.log.Debugf(`启动 over`)
	}
	closeEvent := func() {
		h.log.Debugf(`context关闭`)
		h.CleanContextList(false)
	}
	contextPage := NewContextPage(&context, runParams, userDataPath, userDataIndex, h.log, closeEvent)
	h.AddContextList(contextPage)
	return contextPage, true, nil
}

func (h *ContextPageList) GetContextNotSaveUserData(browser playwright.Browser, runParams *_struct.PlaywrightRunParams) (*ContextPage, error) {
	//查找可用的context
	h.log.Debugf(`不保留用户数据`)
	rContext := h.FindNotSaveUserDataContext(runParams)
	if rContext != nil {
		return rContext, nil
	}
	h.log.Debugf(`准备创建新的context %s %s`, runParams.BrowserAuthUsername, runParams.BrowserAuthPassword)
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
		return nil, contextErr
	}
	closeEvent := func() {
		h.CleanContextList(false)
	}
	contentPage := NewContextPage(&context, runParams, ``, 0, h.log, closeEvent)
	h.AddContextList(contentPage)
	return contentPage, nil
}
