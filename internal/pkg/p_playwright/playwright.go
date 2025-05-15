package p_playwright

import (
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
	"net/url"
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
	h.log.Debugf(`###############################################开始获取page`)
	page, pageErr := h.GetPage()
	if pageErr != nil {
		return gstool.Error(`获取page失败 %s`, pageErr.Error())
	}
	//输出结果存储
	h.log.Debugf(`开始处理process list`)
	for _, processVal := range h.RunParams.ProcessList {
		h.RunParams.ReplaceList = append(h.RunParams.ReplaceList, h.TakeContentMap)
		process := NewProcess(processVal, page, h.RunParams, h.BoolResultMap, h.TakeContentMap, h.log)
		sTime := gstool.TimeNowMilliInt64()
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
			`耗时ms`:           gstool.TimeNowMilliInt64() - sTime,
		}))
		if err != nil {
			return err
		}
		if code == define.ProcessBreak {
			return nil
		}
	}
	return nil
}

func (h *Playwright) GetPage() (*playwright.Page, error) {
	var contextErr error
	var contextPage *ContextPage
	boolCleanFirstBlank := false
	if h.RunParams.CombineType == define.CombineTypeNo { //不保存用户数据
		browser, browserErr := h.GetBrowser()
		if browserErr != nil {
			return nil, browserErr
		}
		contextPage, contextErr = h.ContextPageList.GetContextNotSaveUserData(browser, h.RunParams)
	} else { //保留用户数据
		contextPage, boolCleanFirstBlank, contextErr = h.ContextPageList.GetContextSaveUserData(h.RunParams)
	}
	h.log.Debugf(`获取context结束 %v`, contextErr)
	if contextErr != nil {
		return nil, contextErr
	}
	//注册新的监听
	h.log.Debugf(`NOTICE：注册新的监听，旧的监听失效 %#v`, h.RunParams.ListenUrlList)
	(*contextPage).ListenUrlList = h.RunParams.ListenUrlList
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
	if runParams.UserName == `` || runParams.Id == 0 || runParams.Domain == `` {
		return
	}
	sql := `select * from tbl_smart_link_last where  smart_link_id = ? and user_name = ? and domain = ?`
	smartLinkLast, smartLinkErr := base.Component.TSqlite.Client.QueryBySql(sql, runParams.Id, runParams.UserName, runParams.Domain).One()
	if smartLinkErr != nil {
		h.log.Debugf(`获取最后使用索引失败 %s %s`, sql, smartLinkErr.Error())
		gstool.FmtPrintlnLogTime(`查询失败 %s`, smartLinkErr.Error())
		return
	} else if len(smartLinkLast) > 0 {
		_, err := base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link_last`, map[string]any{
			`smart_link_id`: runParams.Id,
			`user_name`:     runParams.UserName,
			`domain`:        runParams.Domain,
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
			`domain`:          runParams.Domain,
			`create_time`:     time.Now().Unix(),
			`update_time`:     time.Now().Unix(),
		}).Exec()
		if err != nil {
			gstool.FmtPrintlnLogTime(`创建最后使用索引失败 %s`, err.Error())
		}
	}
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

func (h *Playwright) Recycle() error {
	h.log.Debugf(`开始回收..`)
	_ = base.Component.TPlaywright.Pw.Stop()
	h.ContextPageList.CleanContextList(true)
	base.Component.TPlaywright.InitPlaywright()
	InitPageActiveTime()
	return nil
}
