package p_playwright

import (
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
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
