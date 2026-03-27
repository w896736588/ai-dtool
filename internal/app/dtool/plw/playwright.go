package plw

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_common"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
)

type Playwright struct {
	RunParams       *PlaywrightRunParams //运行时参数
	EventLock       sync.Mutex           //事件锁
	TakeContentMap  map[string]string    //提取内容
	BoolResultMap   map[string]bool      //判断结果
	ContextPageList *ContextPageList     //浏览器上下文列表
	log             *gstool.GsSlog
}

func NewPlaywright(runParams *PlaywrightRunParams, log *gstool.GsSlog) *Playwright {
	return &Playwright{
		RunParams:       runParams,
		TakeContentMap:  make(map[string]string),
		BoolResultMap:   make(map[string]bool),
		ContextPageList: NewContextList(log),
		log:             log,
	}
}

func (h *Playwright) Open(call *p_common.Call, stopCall func() bool) error {
	if PlaywrightClient.Pw == nil {
		return errors.New(`未启动浏览器核心`)
	}
	h.RunParams.StreamFunc(`启动playwright`, `获取page`)
	page, pageErr := h.GetPage(call)
	if pageErr != nil {
		return gstool.Error(`获取page失败 %s`, pageErr.Error())
	}
	for _, processVal := range h.RunParams.ProcessList {
		if stopCall != nil && stopCall() {
			_ = (*page).Close()
			return errors.New(`任务已被取消`)
		}
		if cast.ToInt(processVal[`is_async`]) == 1 {
			go func() {
				h.RunParams.StreamFunc(cast.ToString(processVal[`name`]), `异步执行`)
				_, runErr := h.ProcessRun(processVal, page)
				if runErr != nil {
					h.RunParams.StreamFunc(cast.ToString(processVal[`name`]), fmt.Sprintf(`执行失败 %s`, runErr.Error()))
				}
			}()

		} else {
			h.RunParams.StreamFunc(cast.ToString(processVal[`name`]), `按顺序执行`)
			boolContinue, runErr := h.ProcessRun(processVal, page)
			if runErr != nil {
				if cast.ToInt(processVal[`is_error_continue`]) == 1 {
					h.RunParams.StreamFunc(cast.ToString(processVal[`name`]), fmt.Sprintf(`本节点执行失败 %s，继续执行下一个`, runErr.Error()))
				} else {
					h.RunParams.StreamFunc(cast.ToString(processVal[`name`]), fmt.Sprintf(`执行失败 %s`, runErr.Error()))
					return runErr
				}
			}
			if !boolContinue {
				return nil
			}
		}

	}
	return nil
}

func (h *Playwright) ProcessRun(processVal map[string]any, page *playwright.Page) (bool, error) {
	process := NewProcess(processVal, page, h.RunParams, h.BoolResultMap, h.TakeContentMap, h.log)
	sTime := gstool.TimeNowMilliInt64()
	code, _, err := process.Do()
	h.RunParams.StreamFunc(cast.ToString(processVal[`name`]), fmt.Sprintf(`执行时长 %dms`, gstool.TimeNowMilliInt64()-sTime))
	//h.log.Debugf(`执行结果 %s `, gstool.JsonFormat(map[string]any{
	//	`type`:           process.ProcessType,
	//	`reason`:         reason,
	//	`domain`:         h.RunParams.Domain,
	//	`domain_limit`:   process.DomainLimit,
	//	`Locator`:        process.Locator,
	//	`tip`:            process.Tip,
	//	`code`:           code,
	//	`Checks`:         process.Checks,
	//	`TakeContextMap`: h.TakeContentMap,
	//	`BoolResultMap`:  h.BoolResultMap,
	//	`耗时ms`:           gstool.TimeNowMilliInt64() - sTime,
	//}))
	if err != nil {
		return false, err
	}
	//对结果写入到替换列表
	for takeKey, takeValue := range h.TakeContentMap {
		if takeKey == cast.ToString(processVal[`out_key`]) && cast.ToInt(processVal[`append_to_replace`]) == 1 {
			h.RunParams.ReplaceList[takeKey] = takeValue
		}
	}
	if code == define.ProcessBreak {
		return false, nil
	}
	return true, nil
}

// GetContext 获取浏览器实例
func (h *Playwright) GetContext() (*ContextPage, bool, error) {
	if h.RunParams.CombineType == define.CombineTypeNo { //不保存用户数据
		browser, browserErr := h.GetBrowser()
		if browserErr != nil {
			h.RunParams.StreamFunc(`启动playwright`, fmt.Sprintf(`获取browser失败 %s`, browserErr.Error()))
			return nil, false, browserErr
		}
		contextPage, contextErr := h.ContextPageList.GetContextNotSaveUserData(browser, h.RunParams)
		return contextPage, false, contextErr
	} else { //保留用户数据
		return h.ContextPageList.GetContextSaveUserData(h.RunParams)
	}
}

// GetPage 获取page
func (h *Playwright) GetPage(call *p_common.Call) (*playwright.Page, error) {
	//获取浏览器实例
	contextPage, boolCleanFirstBlank, contextErr := h.GetContext()
	h.RunParams.StreamFunc(`启动playwright`, `获取浏览器实例结束`)
	if contextErr != nil {
		return nil, contextErr
	}
	var page playwright.Page
	var pageErr error
	page, pageErr = (*contextPage.Context).NewPage()
	h.RunParams.StreamFunc(`启动playwright`, `创建page`)
	if pageErr != nil {
		h.RunParams.StreamFunc(`启动playwright`, `创建page报错，尝试重建浏览器实例：`+pageErr.Error())
		//重试创建浏览器实例
		h.ContextPageList.RemoveContextPage(contextPage)
		contextPage, boolCleanFirstBlank, contextErr = h.GetContext()
		if contextErr != nil {
			h.RunParams.StreamFunc(`启动playwright`, `再次创建浏览器实例报错，返回：`+contextErr.Error())
			return nil, contextErr
		}
		page, pageErr = (*contextPage.Context).NewPage()
		h.RunParams.StreamFunc(`启动playwright`, `再次创建page`)
		if pageErr != nil {
			h.RunParams.StreamFunc(`启动playwright`, `再次创建page报错，返回：`+pageErr.Error())
			return nil, pageErr
		}
	}
	h.RunParams.StreamFunc(`启动playwright`, `传入过滤url`+gstool.JsonEncode(h.RunParams.FilterUris))
	(*contextPage).RegisterLinks(page, h.RunParams.ListenCurls, h.RunParams.FilterUris, h.RunParams.StreamFunc)
	//记录登录记录
	h.LastUserDataIndex(h.RunParams, contextPage.UserDataIndex, call)
	// 关闭一个blank
	if boolCleanFirstBlank {
		contextPage.CloseFirstPage()
	}
	//跳转链接
	u, _ := url.Parse(h.RunParams.Link)
	h.RunParams.StreamFunc(`启动playwright`, `打开link：`+h.RunParams.Link)
	if _, goErr := page.Goto(u.String()); goErr != nil {
		return nil, goErr
	}
	//等待加载完成
	//h.RunParams.ReplaceList[`{link}`] = u.String()
	PlaywrightClient.WaitForLoadState(&page, h.RunParams.LocatorTimeout)
	return &page, nil
}

func (h *Playwright) LastUserDataIndex(runParams *PlaywrightRunParams, userDataIndex int, call *p_common.Call) {
	if runParams.LastIndexLabel == `` || runParams.Id == 0 || runParams.Domain == `` {
		return
	}
	sql := `select * from tbl_smart_link_last where  smart_link_id = ? and user_name = ? and domain = ?`
	// 最近一次用户目录索引已经迁移到独立 log 库，这里只查询 log 库。
	smartLinkLast, smartLinkErr := common.DbLog.Client.QueryBySql(sql, runParams.Id, runParams.LastIndexLabel, runParams.Domain).One()
	if smartLinkErr != nil {
		runParams.StreamFunc(`记录历史数据目录`, `失败：`+smartLinkErr.Error())
		return
	} else if len(smartLinkLast) > 0 {
		_, err := call.UpdateSmartLastRecord(map[string]any{
			`smart_link_id`: runParams.Id,
			`user_name`:     runParams.LastIndexLabel,
			`domain`:        runParams.Domain,
		}, map[string]any{
			`user_data_index`: userDataIndex,
			`update_time`:     time.Now().Unix(),
		})
		if err != nil {
			runParams.StreamFunc(`记录历史数据目录`, `更新最后使用索引失败：`+err.Error())
		}
	} else {
		_, err := call.CreateSmartLastRecord(map[string]any{
			`smart_link_id`:   runParams.Id,
			`user_name`:       runParams.LastIndexLabel,
			`user_data_index`: userDataIndex,
			`domain`:          runParams.Domain,
			`create_time`:     time.Now().Unix(),
			`update_time`:     time.Now().Unix(),
		})
		if err != nil {
			// 针对并发创建或历史脏数据引发的唯一键冲突，回退为按唯一键更新，避免中断主流程。
			if strings.Contains(err.Error(), `UNIQUE constraint failed: tbl_smart_link_last.domain, tbl_smart_link_last.user_data_index`) {
				_, updateErr := call.UpdateSmartLastRecord(map[string]any{
					`domain`:          runParams.Domain,
					`user_data_index`: userDataIndex,
				}, map[string]any{
					`smart_link_id`: runParams.Id,
					`user_name`:     runParams.LastIndexLabel,
					`update_time`:   time.Now().Unix(),
				})
				if updateErr != nil {
					runParams.StreamFunc(`记录历史数据目录`, `创建冲突后回退更新失败：`+updateErr.Error())
				}
			} else {
				runParams.StreamFunc(`记录历史数据目录`, `创建最后使用索引失败：`+err.Error())
			}
		}
	}
}

func (h *Playwright) GetBrowser() (playwright.Browser, error) {
	if h.RunParams.OpenType == define.OpenTypeWebkitSilence && PlaywrightClient.BrowserWebkitSilence != nil {
		return PlaywrightClient.BrowserWebkitSilence, nil
	} else if h.RunParams.OpenType == define.OpenTypeWebkitChrome && PlaywrightClient.BrowserWebkitChrome != nil {
		return PlaywrightClient.BrowserWebkitChrome, nil
	}
	var browserErr error
	if h.RunParams.OpenType == define.OpenTypeWebkitSilence {
		PlaywrightClient.BrowserWebkitSilence, browserErr = PlaywrightClient.Pw.Chromium.Launch()
		if browserErr != nil {
			PlaywrightClient.BrowserWebkitSilence = nil
			return nil, browserErr
		} else {
			return PlaywrightClient.BrowserWebkitSilence, nil
		}
	} else {
		PlaywrightClient.BrowserWebkitChrome, browserErr = PlaywrightClient.Pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			//DownloadsPath: &h.downloadPath,
			Headless: playwright.Bool(false), //有界面模式
		})
		if browserErr != nil {
			PlaywrightClient.BrowserWebkitChrome = nil
			return nil, browserErr
		} else {
			return PlaywrightClient.BrowserWebkitChrome, nil
		}
	}
}

func (h *Playwright) Recycle() error {
	h.log.Debugf(`开始回收..`)
	_ = PlaywrightClient.Pw.Stop()
	h.ContextPageList.CleanContextList(true)
	PlaywrightClient.InitPlaywright()
	InitPageActiveTime()
	return nil
}
