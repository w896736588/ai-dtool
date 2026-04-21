package component

import (
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
	// 处理下载后自动打开
	DownloadPath string
	EventLock    sync.Mutex
	// 全局
	BrowserWebkitChrome  playwright.Browser
	BrowserWebkitSilence playwright.Browser
	// pw
	Pw  *playwright.Playwright
	Log *gstool.GsSlog
	// 文件
	LockFileFullPath string
}

var lookPathFunc = exec.LookPath
var statFunc = os.Stat
var playwrightInstallFunc = func(h *TPlaywright, version string) {
	go h.Install(nil, version)
}
var newDriverVersionFunc = func() (string, error) {
	pw, err := playwright.NewDriver()
	if err != nil {
		return ``, err
	}
	return pw.Version, nil
}
var playwrightInstallRunOptionsFunc = func() error {
	return playwright.Install(&playwright.RunOptions{
		Browsers: []string{`chromium`},
	})
}
var initPlaywrightBrowserFunc = func(h *TPlaywright) {
	h.InitPlaywright()
}

func NewTPlaywright() *TPlaywright {
	gsLog := gstool.NewSlog2(EnvClient.LogPath, `playwright`)
	_ = gsLog.CleanOldLogs(2)
	return &TPlaywright{
		Log:          gsLog,
		DownloadPath: EnvClient.WebkitDownloadPath,
	}
}

func (h *TPlaywright) SetWebkitPath() {
	// 设置自定义浏览器安装路径
	_ = os.Setenv("PLAYWRIGHT_BROWSERS_PATH", EnvClient.WebkitDriverPath)
	// PLAYWRIGHT_DRIVER_PATH 是驱动目录，不是 node.exe 路径
	_ = os.Unsetenv("PLAYWRIGHT_DRIVER_PATH")
	_ = os.Setenv("PLAYWRIGHT_NODEJS_PATH", EnvClient.NodePath)
	_ = os.Setenv("GOPROXY", "https://goproxy.cn,direct")
}

// EnsureNodeRuntime 确保 Node.js 可用并写回最终路径
func (h *TPlaywright) EnsureNodeRuntime() bool {
	nodePath := resolveNodePath(EnvClient.NodePath)
	if nodePath == `` {
		return false
	}
	EnvClient.NodePath = nodePath
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

// SetTitle 设置 title
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
	if gstool.FileIsExisted(h.LockFileFullPath) {
		lockFile, lockErr := h.TryAcquireInstallLock()
		if lockErr != nil {
			if IsPlaywrightLockBusyError(lockErr) {
				gstool.FmtPrintlnLogTime(`浏览器核心正在安装中，跳过本次重复安装`)
				return
			}
			gstool.FmtPrintlnLogTime(`检查浏览器核心锁失败 %s`, lockErr.Error())
			return
		}
		_ = ReleaseLockedFile(lockFile)
	}
	version, versionErr := newDriverVersionFunc()
	if versionErr != nil {
		gstool.FmtPrintlnLogTime(`获取浏览器核心版本失败 %s`, versionErr.Error())
		return
	}
	playwrightInstallFunc(h, version)
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
		Headless: playwright.Bool(false),
	})
	gstool.FmtPrintlnLogTime(`启动成功..`)
}

func (h *TPlaywright) Install(sse *p_sse.SseShell, version string) {
	_ = version
	lockFile, lockErr := h.TryAcquireInstallLock()
	if lockErr != nil {
		if IsPlaywrightLockBusyError(lockErr) {
			sse.Send(`浏览器核心正在安装中，跳过本次重复安装` + "\n")
			return
		}
		sse.Send(fmt.Sprintf(`获取浏览器核心安装锁失败 %s`, lockErr.Error()) + "\n")
		return
	}
	defer func() {
		_ = ReleaseLockedFile(lockFile)
		_ = gstool.FileDelete(h.LockFileFullPath)
	}()
	sse.Send(`开始安装浏览器核心(只安装chrome),大约几分钟时间` + "\n")
	// 安装锁用于避免并发重复安装，安装结束后删除，方便下次启动重新安装更新。
	_ = h.WriteInstallLockContent(lockFile, `installing`)
	err := playwrightInstallRunOptionsFunc()
	if err != nil {
		sse.Send(fmt.Sprintf(`安装浏览器核心失败 %s`, err.Error()) + "\n")
	} else {
		sse.Send(`安装完成` + "\n")
		initPlaywrightBrowserFunc(h)
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

// TryAcquireInstallLock 尝试获取浏览器核心安装锁，成功时由调用方负责释放。
func (h *TPlaywright) TryAcquireInstallLock() (*os.File, error) {
	return TryLockFileNonBlocking(h.LockFileFullPath)
}

// WriteInstallLockContent 将安装状态写入已加锁的锁文件，避免重新打开文件破坏锁语义。
func (h *TPlaywright) WriteInstallLockContent(lockFile *os.File, content string) error {
	if lockFile == nil {
		return fmt.Errorf(`lock file is nil`)
	}
	if truncateErr := lockFile.Truncate(0); truncateErr != nil {
		return truncateErr
	}
	if _, seekErr := lockFile.Seek(0, 0); seekErr != nil {
		return seekErr
	}
	if _, writeErr := lockFile.WriteString(content); writeErr != nil {
		return writeErr
	}
	return lockFile.Sync()
}
