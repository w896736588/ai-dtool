package plw

import (
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_curl"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

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
	eventPages      sync.Map
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
	// 先注册新页面事件，避免页面创建后事件尚未绑定导致丢失。
	(*h.Context).OnPage(func(page playwright.Page) {
		h.InitEvents(&page)
	})
	// 同时补绑当前 context 已存在页面，避免复用 context 时下载事件未注册。
	for _, page := range (*h.Context).Pages() {
		h.InitEvents(&page)
	}
	(*h.Context).OnClose(func(context playwright.BrowserContext) {
		h.CloseEvent()
	})
}

// InitEvents 这里可能注册链接有一些问题 已经存在的context在新的链接上不会重新进行注册
func (h *ContextPage) InitEvents(page *playwright.Page) {
	if !h.tryMarkPageEventInited(page) {
		return
	}
	(*page).On("request", func(request playwright.Request) {
		go h.SetPageActive(page)
		return
	})

	(*page).On(`load`, func() {
		go ShowCookieTip(page, h.RunParams)
	})

	//可以监听到 前端下载
	(*page).On(`download`, func(download playwright.Download) {
		go h.handleDownload(page, download)
		return
	})
}

// tryMarkPageEventInited 标记页面事件是否已初始化，避免重复注册。
func (h *ContextPage) tryMarkPageEventInited(page *playwright.Page) bool {
	key := fmt.Sprintf("%p", page)
	_, loaded := h.eventPages.LoadOrStore(key, struct{}{})
	return !loaded
}

// handleDownload 处理下载落盘并自动打开文件。
func (h *ContextPage) handleDownload(page *playwright.Page, download playwright.Download) {
	h.SetPageActive(page)
	fileName := sanitizeWindowsFilename(download.SuggestedFilename())
	PlaywrightClient.AddTipMsg(page, `检测到下载`+fileName+`,别急，自动打开中..`)
	localPath := h.GetDownloadPathByFilename(fileName)
	h.log.Debugf(`download localPath %s`, localPath)
	if saveErr := download.SaveAs(localPath); saveErr != nil {
		h.log.Debugf(`下载保存失败 %s %s`, localPath, saveErr.Error())
		PlaywrightClient.AddTipMsg(page, `下载保存失败：`+fileName)
		return
	}
	if !gstool.FileIsExisted(localPath) {
		h.log.Debugf(`下载保存后文件不存在 %s`, localPath)
		PlaywrightClient.AddTipMsg(page, `下载失败，未找到文件：`+fileName)
		return
	}
	PlaywrightClient.AddTipMsg(page, `开始打开`+fileName)
	openErr := p_common.TOsClient.OpenFileWindows(localPath, localPath)
	if openErr != nil {
		h.log.Debugf(`打开文件失败 %s %s`, localPath, openErr.Error())
		PlaywrightClient.AddTipMsg(page, `自动打开失败，请到下载目录查看：`+fileName)
		return
	}
	h.log.Debugf(`打开文件成功 %s`, localPath)
}

func (h *ContextPage) RegisterLinks(page playwright.Page, registerLinks map[string]*p_curl.CurlRun, filterUris []string, tipFunc func(string, string)) {
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
	//拦截
	for _, uri := range filterUris {
		if uri == `` {
			continue
		}
		tipFunc(`注册过滤请求,`, uri)
		_ = page.Route("**"+uri+"*", func(route playwright.Route) {
			_ = route.Abort()
		})
	}
}

func (h *ContextPage) GetDownloadPath(download playwright.Download) string {
	return h.GetDownloadPathByFilename(download.SuggestedFilename())
}

// GetDownloadPathByFilename 生成下载文件路径并统一处理文件名。
func (h *ContextPage) GetDownloadPathByFilename(fileName string) string {
	return filepath.Join(PlaywrightClient.DownloadPath, p_common.TBaseClient.GetUnique(`download`)+`_`+sanitizeWindowsFilename(fileName))
}

// sanitizeWindowsFilename 清洗 Windows 不允许的文件名字符，避免下载保存失败。
func sanitizeWindowsFilename(fileName string) string {
	fileName = strings.TrimSpace(fileName)
	if fileName == `` {
		return `download.bin`
	}
	replacer := strings.NewReplacer(
		`\`, `_`,
		`/`, `_`,
		`:`, `_`,
		`*`, `_`,
		`?`, `_`,
		`"`, `_`,
		`<`, `_`,
		`>`, `_`,
		`|`, `_`,
	)
	fileName = replacer.Replace(fileName)
	fileName = strings.TrimRight(fileName, ` .`)
	if fileName == `` {
		return `download.bin`
	}
	return fileName
}

func (h *ContextPage) SetPageActive(page *playwright.Page) {
	if h.AutoCloseSecond == 0 {
		return
	}
	h.ActiveTime.Add(page, h.AutoCloseSecond)
}
