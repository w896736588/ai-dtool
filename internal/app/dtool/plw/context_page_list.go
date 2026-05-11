package plw

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
)

// list 所有浏览器列表
var list []*ContextPage
var ContextLock sync.RWMutex

type ContextPageList struct {
	log                *gstool.GsSlog
	smartLinkDirStore  SmartLinkDirectoryStore
	smartLinkLastStore SmartLinkLastStore
}

func getList() *[]*ContextPage {
	if list == nil {
		list = make([]*ContextPage, 0)
	}
	return &list
}

func NewContextList(log *gstool.GsSlog) *ContextPageList {
	return &ContextPageList{
		log:                log,
		smartLinkDirStore:  NewDBSmartLinkDirectoryStore(),
		smartLinkLastStore: NewDBSmartLinkLastStore(),
	}
}

// SetSmartLinkDirectoryStore 允许 agent 注入远程固定目录映射存储实现，避免在 agent 侧访问 sqlite。
func (h *ContextPageList) SetSmartLinkDirectoryStore(store SmartLinkDirectoryStore) {
	if store != nil {
		h.smartLinkDirStore = store
	}
}

// getSmartLinkDirectoryStore 返回固定目录映射存储；未注入时使用服务端默认 DB 实现。
func (h *ContextPageList) getSmartLinkDirectoryStore() SmartLinkDirectoryStore {
	if h.smartLinkDirStore == nil {
		h.smartLinkDirStore = NewDBSmartLinkDirectoryStore()
	}
	return h.smartLinkDirStore
}

// SetSmartLinkLastStore 允许 agent 注入远程存储实现，避免在 agent 侧访问 sqlite。
func (h *ContextPageList) SetSmartLinkLastStore(store SmartLinkLastStore) {
	if store != nil {
		h.smartLinkLastStore = store
	}
}

// getSmartLinkLastStore 返回历史目录存储；未注入时使用服务端默认 DB 实现。
func (h *ContextPageList) getSmartLinkLastStore() SmartLinkLastStore {
	if h.smartLinkLastStore == nil {
		h.smartLinkLastStore = NewDBSmartLinkLastStore()
	}
	return h.smartLinkLastStore
}

func (h *ContextPageList) EventContextClose(contextP *ContextPage) {
	go (*contextP.Context).OnClose(func(context playwright.BrowserContext) {
		contextP.RunParams.StreamFunc(`浏览器实例关闭`, fmt.Sprintf(`%s,%s`, contextP.LinkId, contextP.LinkIdLabel))
		h.CleanContextList(false)
	})
}

func (h *ContextPageList) AddContextList(contextP *ContextPage) {
	ContextLock.Lock()
	defer ContextLock.Unlock()
	*getList() = append(*getList(), contextP)
	h.EventContextClose(contextP)
}

func (h *ContextPageList) EachContextList(f func(context *ContextPage) bool) {
	ContextLock.Lock()
	defer ContextLock.Unlock()
	for _, context := range *getList() {
		if f(context) {
			break
		}
	}
}

func (h *ContextPageList) FindContextList(f func(context *ContextPage) *ContextPage) *ContextPage {
	ContextLock.Lock()
	defer ContextLock.Unlock()
	for _, context := range *getList() {
		rContext := f(context)
		if rContext != nil {
			return rContext
		}
	}
	return nil
}

func (h *ContextPageList) CleanContextList(cleanAll bool) {
	ContextLock.Lock()
	defer ContextLock.Unlock()
	if cleanAll {
		for _, context := range *getList() {
			h.CloseContextPages(context.Context)
		}
		*getList() = make([]*ContextPage, 0)
	} else {
		newContextList := make([]*ContextPage, 0)
		for _, context := range *getList() {
			if context.Context != nil && len((*context.Context).Pages()) > 0 {
				newContextList = append(newContextList, context)
			}
		}
		*getList() = newContextList
	}
}

// CloseContextPages 关闭所有页面
func (h *ContextPageList) CloseContextPages(context *playwright.BrowserContext) {
	pageList := (*context).Pages()
	for _, page := range pageList {
		_ = page.Close()
	}
}

// RemoveContextPage 移除context_page
func (h *ContextPageList) RemoveContextPage(rmContextPage *ContextPage) {
	ContextLock.Lock()
	defer ContextLock.Unlock()
	defer func() {
		if err := recover(); err != nil {
			h.log.Errof(`移除浏览器实例失败 %v`, err)
		}
	}()
	newList := make([]*ContextPage, 0)
	for _, context := range *getList() {
		if context.LinkId == rmContextPage.LinkId {
			_ = (*context.Context).Close()
		} else {
			newList = append(newList, context)
		}
	}
	*getList() = newList
}

func (h *ContextPageList) GetPlaywrightRunList() []map[string]any {
	runList := make([]map[string]any, 0)
	h.EachContextList(func(context *ContextPage) bool {
		pageList := (*context.Context).Pages()
		runList = append(runList, map[string]any{
			`name`:     context.LinkIdLabel,
			`page_num`: len(pageList),
		})
		return false
	})
	return runList
}

func (h *ContextPageList) FindAIContextByRunParams(runParams *PlaywrightRunParams) *ContextPage {
	return h.FindContextList(func(context *ContextPage) *ContextPage {
		if context.OpenType != runParams.OpenType {
			return nil
		}
		// 这里沿用 smart-link 的“同一链接类型”判断，避免 AI 会话复用到别的配置实例上。
		if !h.IsSameLink(context.LinkIdLabel, runParams.LinkIdLabel) {
			return nil
		}
		return context
	})
}

func (h *ContextPageList) FindContextPageByPageKey(pageKey string) (*ContextPage, *playwright.Page) {
	return h.findContextPageByPageKey(pageKey)
}

func (h *ContextPageList) findContextPageByPageKey(pageKey string) (*ContextPage, *playwright.Page) {
	var resultContext *ContextPage
	var resultPage *playwright.Page
	h.EachContextList(func(context *ContextPage) bool {
		page := context.FindPageByKey(pageKey)
		if page != nil {
			resultContext = context
			resultPage = page
			return true
		}
		return false
	})
	return resultContext, resultPage
}

// CleanContextPagesFixDataId 固定目录的，先关掉其他页面
func (h *ContextPageList) CleanContextPagesFixDataId(runParams *PlaywrightRunParams) {
	runParams.StreamFunc(`获取数据目录`, `当前为固定目录，开始清理旧页面`)
	h.EachContextList(func(context *ContextPage) bool {
		//打开方式
		if context.OpenType != runParams.OpenType {
			runParams.StreamFunc(`获取数据目录`, context.LinkIdLabel+`,`+context.LinkId+` 打开方式不一致，不进行清理`)
			return false
		}
		if context.LinkId == runParams.LinkId {
			runParams.StreamFunc(`获取数据目录`, context.LinkIdLabel+`,`+context.LinkId+` 查找到当前实例，开始清理旧页面`)
			context.CloseContextPages()
		}

		return false
	})
	runParams.StreamFunc(`获取数据目录`, `当前为固定目录，结束清理旧页面`)
	time.Sleep(time.Second * 1)
}

func (h *ContextPageList) getUserDataIndex(runParams *PlaywrightRunParams) (int, error) {
	userDataIndex, err := h.GetFixUserDataIndex(runParams)
	if err != nil {
		runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`固定目录获取失败，mappingKey=%s, err=%s`, runParams.DirectoryMappingKey, err.Error()))
		return 0, err
	}
	runParams.StreamFunc(`获取数据目录`, `固定目录，最终使用目录索引 `+cast.ToString(userDataIndex))
	return userDataIndex, nil
}

func (h *ContextPageList) getMappedDirectoryIndex(runParams *PlaywrightRunParams, logPrefix string) (int, error) {
	userDataIndex, err := h.getSmartLinkDirectoryStore().GetByMappingKey(runParams.DirectoryMappingKey)
	if err != nil {
		runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`%s查询固定目录映射失败，mappingKey=%s, err=%s`, logPrefix, runParams.DirectoryMappingKey, err.Error()))
		return 0, err
	}
	if userDataIndex > 0 {
		runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`%s固定目录映射命中，mappingKey=%s, 目录索引=%d`, logPrefix, runParams.DirectoryMappingKey, userDataIndex))
	}
	return userDataIndex, nil
}

func (h *ContextPageList) GetFixUserDataIndex(runParams *PlaywrightRunParams) (int, error) {
	if runParams.DirectoryMappingKey == `` {
		err := errors.New(`固定目录映射键为空，无法按 mapping table 分配目录`)
		runParams.StreamFunc(`获取数据目录`, err.Error())
		return 0, err
	}
	directoryStore := h.getSmartLinkDirectoryStore()
	userDataIndex, err := h.getMappedDirectoryIndex(runParams, ``)
	if err != nil {
		return 0, err
	}
	if userDataIndex > 0 {
		return userDataIndex, nil
	}

	runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`固定目录映射未命中，开始分配目录索引，mappingKey=%s`, runParams.DirectoryMappingKey))
	for i := 1; i < define.MaxUserDataIndex; i++ {
		occupied, occupiedErr := directoryStore.ExistsUserDataIndex(i)
		if occupiedErr != nil {
			runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`检查固定目录索引占用失败，mappingKey=%s, index=%d, err=%s`, runParams.DirectoryMappingKey, i, occupiedErr.Error()))
			return 0, occupiedErr
		}
		if occupied {
			runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`固定目录索引 %d 已被映射表占用，跳过`, i))
			continue
		}
		if h.GetContextByIndexIgnoreOpenType(i) != nil {
			runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`固定目录索引 %d 当前有活跃实例，跳过`, i))
			continue
		}
		upsertErr := directoryStore.UpsertMapping(
			runParams.DirectoryMappingKey,
			runParams.Id,
			runParams.Label,
			runParams.AccountKey,
			i,
		)
		if upsertErr == nil {
			runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`固定目录映射分配成功，mappingKey=%s, 目录索引=%d`, runParams.DirectoryMappingKey, i))
			return i, nil
		}
		if errors.Is(upsertErr, ErrSmartLinkDirectoryIndexOccupied) {
			runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`固定目录索引 %d 并发占用，先重查 mappingKey=%s`, i, runParams.DirectoryMappingKey))
			reloadedIndex, reloadErr := h.getMappedDirectoryIndex(runParams, `重查后`)
			if reloadErr != nil {
				return 0, reloadErr
			}
			if reloadedIndex > 0 {
				runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`固定目录重查命中，mappingKey=%s, 目录索引=%d`, runParams.DirectoryMappingKey, reloadedIndex))
				return reloadedIndex, nil
			}
			runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`固定目录重查仍未命中，mappingKey=%s，继续尝试下一个索引`, runParams.DirectoryMappingKey))
			continue
		}
		runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`写入固定目录映射失败，mappingKey=%s, index=%d, err=%s`, runParams.DirectoryMappingKey, i, upsertErr.Error()))
		return 0, upsertErr
	}
	err = fmt.Errorf(`固定目录映射分配失败，mappingKey=%s，没有找到可用目录索引`, runParams.DirectoryMappingKey)
	runParams.StreamFunc(`获取数据目录`, err.Error())
	return 0, err
}

func (h *ContextPageList) GetLastUserDataIndex(runParams *PlaywrightRunParams) int {
	if runParams.LastIndexLabel == `` {
		return 0
	}
	lastUserDataIndex, smartLinkErr := h.getSmartLinkLastStore().GetLastUserDataIndex(runParams.LastIndexLabel, runParams.Domain)
	if smartLinkErr != nil {
		runParams.StreamFunc(`查询历史数据目录`, fmt.Sprintf(`获取上次使用索引失败 %s`, smartLinkErr.Error()))
		return 0
	} else {
		runParams.StreamFunc(`查询历史数据目录`, `获取上次使用索引成功 `+cast.ToString(lastUserDataIndex))
		return lastUserDataIndex
	}
}

func (h *ContextPageList) IsSameLink(smartLinkUniqueKeyS, smartLinkUniqueKeyT string) bool {
	return strings.Split(smartLinkUniqueKeyS, `_`)[0] == strings.Split(smartLinkUniqueKeyT, `_`)[0]
}

func (h *ContextPageList) GetContextByIndex(dataIndex int, openType define.OpenType) *ContextPage {
	return h.FindContextList(func(context *ContextPage) *ContextPage {
		// 同一个 userDataIndex 可能残留不同打开模式的实例，复用时必须同时校验 openType，
		// 否则首次无头打开后，后续切换为有头模式仍会错误复用旧的无头 context。
		if context.UserDataIndex == dataIndex && context.OpenType == openType {
			return context
		}
		return nil
	})
}

func (h *ContextPageList) GetContextByIndexIgnoreOpenType(dataIndex int) *ContextPage {
	return h.FindContextList(func(context *ContextPage) *ContextPage {
		if context.UserDataIndex == dataIndex {
			return context
		}
		return nil
	})
}

func (h *ContextPageList) GetContextParam(runParams *PlaywrightRunParams) (*ContextPage, int, string, error) {
	runParams.StreamFunc(`获取数据目录`, `开始`)
	//固定打开数据索引 关闭此context下面的所有页面
	h.CleanContextPagesFixDataId(runParams)
	//获取数据索引目录
	userDataIndex, err := h.getUserDataIndex(runParams)
	if err != nil {
		return nil, 0, ``, err
	}
	//通过索引目录拿到已存在的context
	existContextPage := h.GetContextByIndex(userDataIndex, runParams.OpenType)
	if existContextPage != nil {
		runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`已存在浏览器实例 %s ,直接使用`, existContextPage.LinkId))
		return existContextPage, existContextPage.UserDataIndex, existContextPage.UserDataPath, nil
	}
	mismatchContextPage := h.GetContextByIndexIgnoreOpenType(userDataIndex)
	if mismatchContextPage != nil {
		// 同目录下若已有另一种打开模式的实例，先关闭旧实例再重建；
		// 持久化目录可以复用，但浏览器是否有头是启动期选项，不能沿用旧 context 强行切换。
		runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`同目录已存在 %d 类型的旧实例，当前需要 %d，先关闭旧实例再重建`, mismatchContextPage.OpenType, runParams.OpenType))
		h.RemoveContextPage(mismatchContextPage)
	}
	userDataPath := fmt.Sprintf(component.EnvClient.WebkitDataPath+`/%d`, userDataIndex)
	runParams.StreamFunc(`获取数据目录`, fmt.Sprintf(`未找到已存在的浏览器实例，使用的数据目录 %s,开始创建实例`, userDataPath))
	//创建数据索引目录
	_ = gstool.DirCreatePath(userDataPath)
	return nil, userDataIndex, userDataPath, nil
}

// GetContextSaveUserData 获取context 需要保存用户数据
func (h *ContextPageList) GetContextSaveUserData(runParams *PlaywrightRunParams) (*ContextPage, bool, error) {
	runParams.StreamFunc(`获取浏览器实例`, `需要保存用户数据 `+runParams.LinkIdLabel+`,`+runParams.LinkId)
	existContextPage, userDataIndex, userDataPath, err := h.GetContextParam(runParams)
	if err != nil {
		runParams.StreamFunc(`获取浏览器实例`, fmt.Sprintf(`获取数据目录失败 %s`, err.Error()))
		return nil, false, err
	}
	if existContextPage != nil {
		runParams.StreamFunc(`获取浏览器实例`, fmt.Sprintf(`已存在实例：%s ,直接使用 数据保存目录：%s`, runParams.LinkIdLabel+`,`+runParams.LinkId, userDataPath))
		return existContextPage, false, nil
	}
	//打开模式
	Headless := false
	if runParams.OpenType == define.OpenTypeWebkitSilence {
		runParams.StreamFunc(`获取浏览器实例`, `使用无头模式打开`)
		Headless = true
	} else {
		runParams.StreamFunc(`获取浏览器实例`, `使用有头模式打开`)
	}
	var context playwright.BrowserContext
	var contextErr error
	//浏览器自带验证
	if runParams.BrowserAuthUsername != `` && runParams.BrowserAuthPassword != `` {
		runParams.StreamFunc(`获取浏览器实例`, fmt.Sprintf(`打开contxt，使用浏览器自带验证 用户名%s,超时时间 %f`, runParams.BrowserAuthUsername, runParams.GetPageTimeout))
		context, contextErr = component.PlaywrightClient.Pw.Chromium.LaunchPersistentContext(userDataPath, playwright.BrowserTypeLaunchPersistentContextOptions{
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
			Args: append([]string{
				`--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36`, //隐藏无头
			}, runParams.ExtraBrowserArgs...),
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
			runParams.StreamFunc(`获取浏览器实例`, fmt.Sprintf(`启动报错 %s`, contextErr.Error()))
			return nil, false, contextErr
		}
	} else {
		runParams.StreamFunc(`获取浏览器实例`, fmt.Sprintf(`启动超时时间：%f`, runParams.GetPageTimeout))
		context, contextErr = component.PlaywrightClient.Pw.Chromium.LaunchPersistentContext(userDataPath, playwright.BrowserTypeLaunchPersistentContextOptions{
			//DownloadsPath:     &h.downloadPath,
			Headless: &Headless,
			//Channel:           playwright.String(runParams.Channel),//增加这个会导致问题 关闭后不能正常启动下一个
			NoViewport:        playwright.Bool(true),
			JavaScriptEnabled: playwright.Bool(true),
			AcceptDownloads:   playwright.Bool(true),
			IgnoreHttpsErrors: playwright.Bool(true),
			Locale:            playwright.String(`zh-CN`),
			Timeout:           &runParams.GetPageTimeout,
			Args: append([]string{
				`--user-agent=Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/114.0.0.0 Safari/537.36`, //隐藏无头
			}, runParams.ExtraBrowserArgs...),
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
			runParams.StreamFunc(`获取浏览器实例`, fmt.Sprintf(`启动报错 %s`, contextErr.Error()))
			return nil, false, contextErr
		}
		runParams.StreamFunc(`获取浏览器实例`, `启动完成`)
	}
	//context关闭回调
	closeEvent := func() {
		//runParams.StreamFunc(`浏览器实例关闭事件回调`, `关闭 `+runParams.LinkIdLabel+`,`+runParams.LinkId)
	}
	runParams.StreamFunc(`获取浏览器实例`, `成功，创建实例对象`)
	contextPage := NewContextPage(&context, runParams, userDataPath, userDataIndex, h.log, closeEvent)
	h.AddContextList(contextPage)
	runParams.StreamFunc(`获取浏览器实例`, `创建实例对象成功，类型值：`+contextPage.LinkIdLabel+`,唯一值：`+contextPage.LinkId)
	return contextPage, true, nil
}
