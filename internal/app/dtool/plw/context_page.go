package plw

import (
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_curl"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/playwright-community/playwright-go"
)

type ContextPage struct {
	Context         *playwright.BrowserContext
	LinkIdLabel     string          //选项唯一值  链接配置ID_label  记录是哪个类型的context 用于计数
	UserDataIndex   int             //数据目录索引
	UserDataPath    string          //数据目录
	LinkId          string          //唯一标记 context 记录是哪个目录的context
	OpenType        define.OpenType //打开类型
	CombineType     int             //查找context类型
	AutoCloseSecond int             //非活跃自动关闭 1开启 0关闭
	CloseEvent      func()          //关闭事件
	log             *gstool.GsSlog
	ActiveTime      *PageActiveTime
	RunParams       *PlaywrightRunParams
}

func NewContextPage(context *playwright.BrowserContext, runParams *PlaywrightRunParams, userDataPath string,
	userDataIndex int, log *gstool.GsSlog, closeEvent func()) *ContextPage {
	c := &ContextPage{
		Context:         context,
		LinkIdLabel:     runParams.LinkIdLabel,
		UserDataIndex:   userDataIndex,
		UserDataPath:    userDataPath,
		LinkId:          runParams.LinkId,
		OpenType:        runParams.OpenType,
		AutoCloseSecond: runParams.AutoCloseSecond,
		CombineType:     runParams.CombineType,
		CloseEvent:      closeEvent,
		log:             log,
		ActiveTime:      NewPageActiveTime(),
		RunParams:       runParams,
	}
	c.Init()
	return c
}

func (h *ContextPage) CloseContextPages() {
	for _, page := range h.Pages() {
		_ = page.Close()
	}
}

func (h *ContextPage) Pages() []playwright.Page {
	return (*h.Context).Pages()
}

func (h *ContextPage) CloseFirstPage() {
	contextPageList := h.Pages()
	if len(contextPageList) > 0 {
		h.log.Debugf(`关闭页面 %#v`, contextPageList[0].URL())
		_ = contextPageList[0].Close()
	}
}

func (h *ContextPage) Init() {
	go func() {
		(*h.Context).OnPage(func(page playwright.Page) {
			go h.InitEvents(&page)
		})
		(*h.Context).OnClose(func(context playwright.BrowserContext) {
			h.CloseEvent()
		})
	}()
}

// InitEvents 这里可能注册链接有一些问题 已经存在的context在新的链接上不会重新进行注册
func (h *ContextPage) InitEvents(page *playwright.Page) {
	(*page).On("request", func(request playwright.Request) {
		go h.SetPageActive(page)
		return
	})

	(*page).On(`load`, func() {
		go ShowCookieTip(page, h.RunParams)
	})

	//可以监听到 前端下载
	(*page).On(`download`, func(download playwright.Download) {
		h.SetPageActive(page)
		PlaywrightClient.AddTipMsg(page, `检测到下载`+download.SuggestedFilename()+`,别急，自动打开中..`)
		localPath := h.GetDownloadPath(download)
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
					PlaywrightClient.AddTipMsg(page, `开始打开`+download.SuggestedFilename())
					openErr := p_common.TOsClient.OpenFileWindows(localPath, localPath)
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

func (h *ContextPage) RegisterLinks(page playwright.Page, registerLinks map[string]*p_curl.CurlRun) {
	if registerLinks != nil {
		for listenUri, cur := range registerLinks {
			cur.CurlEvents.NoticeCall(`注册 **` + listenUri)
			_ = page.Route("**"+listenUri+"*", func(route playwright.Route) {
				go func() {
					originalRequest := route.Request()
					cur.ParseConfig.Headers = originalRequest.Headers()
					cur.ParseConfig.Body, _ = originalRequest.PostData()
					cur.ParseConfig.Url = originalRequest.URL()
					_, _ = cur.Run()
				}()
				_ = route.Abort()
			})
		}
	}
}

func (h *ContextPage) GetDownloadPath(download playwright.Download) string {
	return PlaywrightClient.DownloadPath + `/` + p_common.TBaseClient.GetUnique(`download`) + `_` + download.SuggestedFilename()
}

func (h *ContextPage) SetPageActive(page *playwright.Page) {
	if h.AutoCloseSecond == 0 {
		return
	}
	h.ActiveTime.Add(page, h.AutoCloseSecond)
}
