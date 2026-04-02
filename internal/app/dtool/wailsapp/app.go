package wailsapp

import (
	"dev_tool/internal/app/dtool"
	"fmt"
	"net"
	"sync"
	"time"

	"github.com/wailsapp/wails/v3/pkg/application"
)

// DesktopApp 管理桌面端生命周期与本地后端启动流程。
// DesktopApp manages desktop lifecycle and the local backend bootstrap flow.
type DesktopApp struct {
	app        *application.App
	window     application.Window
	configFile string
	bootOnce   sync.Once
}

// NewDesktopApp 创建桌面应用实例。
// NewDesktopApp creates the desktop application wrapper.
func NewDesktopApp(configFile string) *DesktopApp {
	return &DesktopApp{configFile: configFile}
}

// BindRuntime 绑定 Wails 3 的应用与窗口句柄，供日志和页面跳转复用。
// BindRuntime stores the Wails 3 app and window handles for logging and page redirection.
func (a *DesktopApp) BindRuntime(app *application.App, window application.Window) {
	a.app = app
	a.window = window
}

// DomReady 在前端运行时就绪后异步拉起后端并跳转到本地页面。
// DomReady boots the backend asynchronously after the frontend runtime is ready.
func (a *DesktopApp) DomReady() {
	go a.bootBackendAndOpen()
}

// Shutdown 在桌面应用退出时关闭后端资源。
// Shutdown releases backend resources when the desktop app exits.
func (a *DesktopApp) Shutdown() {
	dtool.Stop()
}

// bootBackendAndOpen 保证后端只初始化一次，并在端口就绪后执行跳转。
// bootBackendAndOpen ensures backend startup runs once and redirects after the port is ready.
func (a *DesktopApp) bootBackendAndOpen() {
	a.bootOnce.Do(func() {
		dtool.InitBase(a.configFile)
		go dtool.InitComponent()

		if !waitPortReady(dtool.GetPrimaryPort(), 30*time.Second) {
			a.logErrorf("后端端口启动超时: %s / Backend port startup timed out: %s", dtool.GetPrimaryPort(), dtool.GetPrimaryPort())
			return
		}

		targetURL := dtool.GetPrimaryURL()
		// 只有在窗口句柄存在时才执行跳转，避免启动时序异常触发空引用。
		// Redirect only when the window handle is available, avoiding nil access during startup timing races.
		if a.window == nil {
			a.logErrorf("桌面窗口未初始化，无法跳转到后端页面: %s / Desktop window is not ready for backend redirect: %s", targetURL, targetURL)
			return
		}
		a.window.ExecJS(fmt.Sprintf("window.location.replace(%q)", targetURL))
	})
}

// logErrorf 统一输出桌面端错误日志，优先复用 Wails 3 logger。
// logErrorf emits desktop-side errors through the Wails 3 logger when available.
func (a *DesktopApp) logErrorf(format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if a.app != nil && a.app.Logger != nil {
		a.app.Logger.Error(message)
		return
	}
	fmt.Println(message)
}

// waitPortReady 轮询检测端口是否可建立连接。
// waitPortReady polls until the target port accepts TCP connections.
func waitPortReady(port string, timeout time.Duration) bool {
	deadline := time.Now().Add(timeout)
	addr := "127.0.0.1:" + port
	for time.Now().Before(deadline) {
		conn, err := net.DialTimeout("tcp", addr, 600*time.Millisecond)
		if err == nil {
			_ = conn.Close()
			return true
		}
		time.Sleep(250 * time.Millisecond)
	}
	return false
}
