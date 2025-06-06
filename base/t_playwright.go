package base

import (
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"log"
	"math"
	"net/url"
	"os"
	"strings"
	"sync"
)

type TPlaywright struct {
	//处理下载后自动打开
	DownloadPath string
	EventLock    sync.Mutex
	//全局
	BrowserWebkitChrome  playwright.Browser
	BrowserWebkitSilence playwright.Browser
	//pw
	Pw  *playwright.Playwright
	Log *gstool.GsSlog
	//文件
	LockFileFullPath string
}

func NewTSmartLink() *TPlaywright {
	gsLog := gstool.NewSlog3(Component.Env.LogPath, `playwright`)
	return &TPlaywright{
		Log:          gsLog,
		DownloadPath: Component.Env.WebkitDownloadPath,
	}
}

func (h *TPlaywright) SetWebkitPath() {
	// 设置自定义浏览器安装路径
	_ = os.Setenv("PLAYWRIGHT_BROWSERS_PATH", Component.Env.WebkitDriverPath)
	_ = os.Setenv("PLAYWRIGHT_DRIVER_PATH", Component.Env.NodePath)
	_ = os.Setenv("GOPROXY", "https://goproxy.cn,direct")
}

func (h *TPlaywright) GetContextUnique(runParams *_struct.PlaywrightRunParams) string {
	return fmt.Sprintf(`context_unique_%d`, runParams.Id)
}

func (h *TPlaywright) WaitForLoadState(page *playwright.Page, timeout float64) {
	_ = (*page).WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateDomcontentloaded,
		Timeout: &timeout,
	})
	_ = (*page).WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateNetworkidle,
		Timeout: &timeout,
	})
	_ = (*page).WaitForLoadState(playwright.PageWaitForLoadStateOptions{
		State:   playwright.LoadStateLoad,
		Timeout: &timeout,
	})
}

func (h *TPlaywright) IsSameLink(smartLinkUniqueKeyS, smartLinkUniqueKeyT string) bool {
	return strings.Split(smartLinkUniqueKeyS, `_`)[0] == strings.Split(smartLinkUniqueKeyT, `_`)[0]
}

func (h *TPlaywright) WitchDownload() {
	if err := os.MkdirAll(h.DownloadPath, 0755); err != nil {
		log.Fatalf("创建目录失败: %v", err)
	}
	h.Log.Debugf(`开始监听%s`, h.DownloadPath)
}

// SetTitle 设置title
func (h *TPlaywright) SetTitle(page playwright.Page, title string) {
	_, _ = page.Evaluate(`(function() {
			document.title = "` + title + `";
	})();`)
}

// AddTipMsg 向页面上输出提示
func (h *TPlaywright) AddTipMsg(page *playwright.Page, tip string) {
	if tip == `` {
		return
	}
	content := Component.TJas.Get(`p_js`, `tip.js`)
	content = gstool.SReplaces(content, map[string]string{
		`{tip}`: tip,
	})
	_, _ = (*page).Evaluate(content)
}

func (h *TPlaywright) SmartCheckAndUpdate() {
	pw, _ := playwright.NewDriver()
	if !gstool.FileIsExisted(h.LockFileFullPath) {
		go h.Install(pw.Version)
	} else {
		content, contentErr := gstool.FileGetContent(h.LockFileFullPath)
		if contentErr != nil {
			h.Log.Errof(`获取文件内容失败 %s`, contentErr.Error())
		} else if content != pw.Version {
			go h.Install(pw.Version)
		} else {
			h.Log.Debugf(`浏览器核心最新版本为：%s ，当前安装版本为：%s,不需要进行更新`, pw.Version, content)
			go h.InitPlaywright()
		}
	}
}

func (h *TPlaywright) InitPlaywright() {
	h.Log.Debugf(`启动浏览器核心..`)
	var pwErr error
	h.Pw, pwErr = playwright.Run()
	if pwErr != nil {
		h.Log.Debugf(`启动浏览器核心失败 %s`, pwErr.Error())
		return
	}
	h.BrowserWebkitSilence, _ = h.Pw.Chromium.Launch()
	h.BrowserWebkitChrome, _ = h.Pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		//DownloadsPath: &h.DownloadPath,
		Headless: playwright.Bool(false), //有界面模式
	})
}

func (h *TPlaywright) Install(version string) {
	gstool.FmtPrintlnLogTime(`开始安装浏览器核心(只安装chrome),大约几分钟时间`)
	err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{`chromium`},
	})
	if err != nil {
		gstool.FmtPrintlnLogTime(`安装浏览器核心失败 %s`, err.Error())
		_ = gstool.FileDelete(h.LockFileFullPath)
	} else {
		_ = gstool.FilePutContentCover(h.LockFileFullPath, version)
		gstool.FmtPrintlnLogTime(`安装完成`)
		h.InitPlaywright()
	}
}

func (h *TPlaywright) GetRunParams(id int, label, userName, password string, openNum int, replaceList *[]map[string]string) (*_struct.PlaywrightRunParams, error) {
	runParams := &_struct.PlaywrightRunParams{}
	if id == 0 {
		return runParams, errors.New(`链接ID不能为空`)
	}
	if label == `` {
		return runParams, errors.New(`链接label不能为空`)
	}
	runParams.Id = id
	smartLink, smartLinkErr := Component.TSqlite.Client.QueryBySql(`select * from tbl_smart_link where id = ? `, id).One()
	if smartLinkErr != nil {
		return runParams, errors.New(smartLinkErr.Error())
	}
	if len(smartLink) == 0 {
		return runParams, errors.New(`不存在的链接`)
	}
	linkList := make([]map[string]any, 0)
	runParams.DownloadFinds = strings.Split(cast.ToString(smartLink[`download_finds`]), `,`)
	runParams.AutoCloseSecond = cast.ToInt(smartLink[`auto_close_second`])
	runParams.Channel = cast.ToString(smartLink[`channel`])
	runParams.ContextUnique = h.GetContextUnique(runParams)
	runParams.ShowCookies = make([]_struct.ShowCookie, 0)
	_ = gstool.JsonDecode(cast.ToString(smartLink[`show_cookies`]), &runParams.ShowCookies)
	if runParams.Channel == `` {
		runParams.Channel = `chromium`
	}
	h.Log.Debugf(`使用浏览器核心 ` + runParams.Channel)
	decodeErr := gstool.JsonDecode(cast.ToString(smartLink[`links`]), &linkList)
	if decodeErr != nil {
		return runParams, errors.New(decodeErr.Error())
	}
	for _, link := range linkList {
		if cast.ToString(link[`label`]) == label {
			runParams.Link = cast.ToString(link[`link`])
			runParams.SmartLinkUniqueKey = cast.ToString(runParams.Id) + `_` + label
			runParams.OpenNum = 0
			runParams.Cookie = cast.ToString(link[`cookie`])
			headerMap := make(map[string]string)
			_ = gstool.JsonDecode(cast.ToString(link[`headers`]), &headerMap)
			runParams.Headers = headerMap
			runParams.BrowserAuthUsername = cast.ToString(link[`browser_auth_username`])
			runParams.BrowserAuthPassword = cast.ToString(link[`browser_auth_password`])
			break
		}
	}
	if runParams.Link == `` {
		return runParams, errors.New(`链接不存在，检查是否json格式错误`)
	}
	runParams.Link = gstool.SReplaces(runParams.Link, map[string]string{
		`{rand}`: Component.TBase.GetUnique(`tool_`),
	})
	runParams.CombineType = cast.ToInt(smartLink[`combine_type`])
	runParams.OpenNum = cast.ToInt(math.Max(1, cast.ToFloat64(openNum)))
	runParams.OpenType = define.OpenType(cast.ToInt(smartLink[`open_type`]))
	process := cast.ToString(smartLink[`process`])
	processList := make([]map[string]any, 0)
	if process != `` {
		decodeErr = gstool.JsonDecode(process, &processList)
		if decodeErr != nil {
			return runParams, errors.New(`配置失败` + decodeErr.Error())
		}
	}
	parsedURL, err := url.Parse(runParams.Link)
	if err != nil {
		return runParams, gstool.Error(`解析地址%s失败 %s`, runParams.Link, err.Error())
	}
	runParams.Domain = parsedURL.Host
	runParams.Scheme = parsedURL.Scheme
	runParams.UserName = userName
	if runParams.UserName != `` {
		runParams.LastIndexLabel = runParams.UserName
	} else {
		runParams.LastIndexLabel = label
	}
	gstool.FmtPrintlnLogTime(`打开链接 %s LastIndexLabel ： %s`, runParams.Link, runParams.LastIndexLabel)
	runParams.Password = password
	runParams.ProcessList = processList
	runParams.ReplaceList = *replaceList
	runParams.LocatorTimeout = 1000
	runParams.GetPageTimeout = 3000
	return runParams, nil
}

// SmartLinkPlaywrightVersion 获取浏览器核心版本
func (h *TPlaywright) SmartLinkPlaywrightVersion() (*playwright.PlaywrightDriver, error) {
	return playwright.NewDriver()
}

func (h *TPlaywright) SmartLinkDownloadPath() error {
	return Component.TOs.OpenDirWindows(gstool.DirPathFormatToWindows(h.DownloadPath))
}

// ShowCookieTip 展示cookie中的某个值
func (h *TPlaywright) ShowCookieTip(page *playwright.Page, runParams *_struct.PlaywrightRunParams) {
	if len(runParams.ShowCookies) == 0 {
		return
	}
	replaceList := make([]_struct.ShowCookie, 0)
	for _, config := range runParams.ShowCookies {
		if gstool.SContains(strings.ToLower((*page).URL()), config.DomainList) {
			replaceList = append(replaceList, config)
		}
	}
	if len(replaceList) == 0 {
		return
	}
	config := gstool.JsonEncode(replaceList)
	Component.TPlaywright.Log.Debugf(`配置的js %s`, config)
	content := Component.TJas.Get(`p_js`, `info.js`)
	content = gstool.SReplaces(content, map[string]string{
		`{config}`: config,
	})
	_, _ = (*page).Evaluate(content)
}

func (h *TPlaywright) ValueFormat(value string, runParams *_struct.PlaywrightRunParams) string {
	value = gstool.SReplaces(value, map[string]string{
		`{user_name}`: runParams.UserName,
		`{password}`:  runParams.Password,
		`{rand}`:      Component.TBase.GetUnique(`input_rand_`),
		`{domain}`:    runParams.Domain,
	})
	//针对输入进行替换
	for _, replaceVal := range runParams.ReplaceList {
		value = gstool.SReplaces(value, replaceVal)
	}
	return value
}

func (h *TPlaywright) ValueClean(value string) string {
	gstool.FmtPrintlnLogTime(`--%s--%s`, value, gstool.SReplaces(value, map[string]string{
		"\n": "",
		" ":  "",
	}))
	gstool.FmtPrintlnLogTime(`--%s--%s`, value, strings.TrimSpace(value))
	return gstool.SReplaces(value, map[string]string{
		"\n": "",
		" ":  "",
	})
}
