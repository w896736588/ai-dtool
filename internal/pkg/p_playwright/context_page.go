package p_playwright

import (
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/gshttp"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"time"
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
	RunParams       *_struct.PlaywrightRunParams
}

func NewContextPage(context *playwright.BrowserContext, runParams *_struct.PlaywrightRunParams, userDataPath string,
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
		go base.Component.TPlaywright.ShowCookieTip(page, h.RunParams)
	})

	//可以监听到 前端下载
	(*page).On(`download`, func(download playwright.Download) {
		h.SetPageActive(page)
		h.log.Debugf(`下载 %#v`, download)
		AddTipMsg(page, `检测到下载`+download.SuggestedFilename()+`,别急，自动打开中..`)
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
					AddTipMsg(page, `开始打开`+download.SuggestedFilename())
					openErr := base.Component.TOs.OpenFileWindows(localPath, localPath)
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

func (h *ContextPage) RegisterLinks(page playwright.Page, registerLinks map[string]*_struct.ListenUrl) {
	if registerLinks != nil {
		for listenUri, listen := range registerLinks {
			listen.Callback(listenUri, `注册 **`+listenUri, nil)
			h.RunParams.StreamFunc(`context`, `注册监听链接`+listenUri)
			_ = page.Route("**"+listenUri, func(route playwright.Route) {
				listen.Callback(listenUri, `捕获到请求`+route.Request().URL(), nil)
				h.RunParams.StreamFunc(`context`, `捕获到注册的连接`+route.Request().URL())
				go h.ListenUrl(route, listen)
				_ = route.Abort()
			})
		}
	} else {
		h.RunParams.StreamFunc(`context`, `没有注册链接`)
	}
}

func (h *ContextPage) GetDownloadPath(download playwright.Download) string {
	return base.Component.TPlaywright.DownloadPath + `/` + base.Component.TBase.GetUnique(`download`) + `_` + download.SuggestedFilename()
}

func (h *ContextPage) SetPageActive(page *playwright.Page) {
	if h.AutoCloseSecond == 0 {
		return
	}
	h.ActiveTime.Add(page, h.AutoCloseSecond)
}

func (h *ContextPage) ListenUrl(route playwright.Route, listen *_struct.ListenUrl) {
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
		if listen.ParseType == define.RegisterLinkParseTypeJson {
			res, resErr = cli.OpenStreamBytesEnd([]byte("}}}"), func(s string, err error) {
				listen.Callback(requestUrl, s, err)
			}, func(bytes []byte) []byte {
				return bytes
			}).Request(200).Result()
		} else {
			res, resErr = cli.OpenStreamBytesEnd([]byte("\n\n"), func(s string, err error) {
				listen.Callback(requestUrl, s, err)
			}, func(bytes []byte) []byte {
				return bytes
			}).Request(200).Result()
		}

	} else {
		res, resErr = cli.Request(200).Result()
		if resErr == nil {
			listen.Callback(requestUrl, cast.ToString(res), nil)
		}
	}
	if resErr != nil {
		listen.EndCallBack(resErr.Error())
	} else {
		listen.EndCallBack(`请求完成`)
	}
}
