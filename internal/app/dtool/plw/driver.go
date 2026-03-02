package plw

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_sse"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/playwright-community/playwright-go"
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

var PlaywrightClient *TPlaywright
var lookPathFunc = exec.LookPath
var statFunc = os.Stat

func NewTPlaywright() *TPlaywright {
	gsLog := gstool.NewSlog2(component.EnvClient.LogPath, `playwright`)
	_ = gsLog.CleanOldLogs(2)
	return &TPlaywright{
		Log:          gsLog,
		DownloadPath: component.EnvClient.WebkitDownloadPath,
	}
}

func (h *TPlaywright) SetWebkitPath() {
	// 设置自定义浏览器安装路径
	_ = os.Setenv("PLAYWRIGHT_BROWSERS_PATH", component.EnvClient.WebkitDriverPath)
	// PLAYWRIGHT_DRIVER_PATH 是驱动目录，不是 node.exe 路径
	_ = os.Unsetenv("PLAYWRIGHT_DRIVER_PATH")
	_ = os.Setenv("PLAYWRIGHT_NODEJS_PATH", component.EnvClient.NodePath)
	_ = os.Setenv("GOPROXY", "https://goproxy.cn,direct")
}

// EnsureNodeRuntime 确保 Node.js 可用并写回最终路径
func (h *TPlaywright) EnsureNodeRuntime() bool {
	nodePath := resolveNodePath(component.EnvClient.NodePath)
	if nodePath == `` {
		return false
	}
	component.EnvClient.NodePath = nodePath
	h.SetWebkitPath()
	return true
}

// resolveNodePath 解析 Node.js 可执行路径
func resolveNodePath(configNodePath string) string {
	return resolveNodePathWithDeps(configNodePath, lookPathFunc, statFunc)
}

// resolveNodePathWithDeps 通过依赖注入解析路径，便于单测
func resolveNodePathWithDeps(configNodePath string, lookPath func(file string) (string, error), stat func(name string) (os.FileInfo, error)) string {
	configNodePath = strings.TrimSpace(configNodePath)
	tryByStat := func(path string) string {
		if path == `` {
			return ``
		}
		info, err := stat(path)
		if err != nil || info == nil {
			return ``
		}
		if info.IsDir() {
			nodeExe := filepath.Join(path, "node.exe")
			nodeInfo, nodeErr := stat(nodeExe)
			if nodeErr == nil && nodeInfo != nil && !nodeInfo.IsDir() {
				return nodeExe
			}
			return ``
		}
		return path
	}
	tryByLookPath := func(binName string) string {
		if binName == `` {
			return ``
		}
		binPath, err := lookPath(binName)
		if err != nil {
			return ``
		}
		return binPath
	}
	// 优先使用配置值（完整路径、目录、可执行名均支持）
	if configNodePath != `` {
		if path := tryByStat(configNodePath); path != `` {
			return path
		}
		if path := tryByLookPath(configNodePath); path != `` {
			return path
		}
	}
	// 回退系统 PATH
	if path := tryByLookPath("node"); path != `` {
		return path
	}
	return ``
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
	content := p_common.TJasClient.Get(`p_js`, `tip.js`)
	content = gstool.SReplaces(content, map[string]string{
		`{tip}`: tip,
	})
	_, _ = (*page).Evaluate(content)
}

func (h *TPlaywright) SmartCheckAndUpdate(sse *p_sse.SseShell) {
	gstool.FmtPrintlnLogTime(`检查并更新核心`)
	pw, _ := playwright.NewDriver()
	if !gstool.FileIsExisted(h.LockFileFullPath) {
		go h.Install(sse, pw.Version)
	} else {
		content, contentErr := gstool.FileGetContent(h.LockFileFullPath)
		if contentErr != nil {
			gstool.FmtPrintlnLogTime(`获取文件内容失败 %s`, contentErr.Error())
		} else if content != pw.Version {
			go h.Install(sse, pw.Version)
		} else {
			gstool.FmtPrintlnLogTime(`浏览器核心最新版本为：%s ，当前安装版本为：%s,不需要进行更新`, pw.Version, content)
			go h.InitPlaywright()
		}
	}
}

func (h *TPlaywright) InitPlaywright() {
	gstool.FmtPrintlnLogTime(`启动浏览器核心..`)
	var pwErr error
	h.Pw, pwErr = playwright.Run()
	if pwErr != nil {
		gstool.FmtPrintlnLogTime(`启动浏览器核心失败 %s`, pwErr.Error())
		return
	}
	h.BrowserWebkitSilence, _ = h.Pw.Chromium.Launch()
	h.BrowserWebkitChrome, _ = h.Pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		//DownloadsPath: &h.DownloadPath,
		Headless: playwright.Bool(false), //有界面模式
	})
	gstool.FmtPrintlnLogTime(`启动成功..`)
}

func (h *TPlaywright) Install(sse *p_sse.SseShell, version string) {
	sse.Send(`开始安装浏览器核心(只安装chrome),大约几分钟时间` + "\n")
	err := playwright.Install(&playwright.RunOptions{
		Browsers: []string{`chromium`},
	})
	if err != nil {
		sse.Send(fmt.Sprintf(`安装浏览器核心失败 %s`, err.Error()) + "\n")
		_ = gstool.FileDelete(h.LockFileFullPath)
	} else {
		_ = gstool.FilePutContentCover(h.LockFileFullPath, version)
		sse.Send(`安装完成` + "\n")
		h.InitPlaywright()
	}
}

// SmartLinkPlaywrightVersion 获取浏览器核心版本
func (h *TPlaywright) SmartLinkPlaywrightVersion() (*playwright.PlaywrightDriver, error) {
	return playwright.NewDriver()
}

func (h *TPlaywright) SmartLinkDownloadPath() error {
	return p_common.TOsClient.OpenDirWindows(gstool.DirPathFormatToWindows(h.DownloadPath))
}

func (h *TPlaywright) ValueClean(value string) string {
	return gstool.SReplaces(value, map[string]string{
		"\n": "",
		" ":  "",
	})
}
