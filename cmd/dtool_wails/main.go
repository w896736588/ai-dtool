package main

import (
	"dev_tool/internal/app/dtool/wailsapp"
	"embed"
	"flag"
	"io/fs"
	"os"

	"github.com/wailsapp/wails/v3/pkg/application"
	"github.com/wailsapp/wails/v3/pkg/events"
)

var ConfigFile string

const (
	// desktopWindowDefaultHeight 表示无法获取屏幕信息时的默认窗口高度。 // Defines the fallback window height when screen information is unavailable.
	desktopWindowDefaultHeight = 900
	// desktopWindowMinHeight 表示桌面窗口允许的最小高度。 // Defines the minimum allowed height for the desktop window.
	desktopWindowMinHeight = 700
)

//go:embed frontend/dist
var assets embed.FS

func main() {
	defaultConfigFile := os.Getenv("DTOOL_CONFIG_FILE")
	if defaultConfigFile == "" {
		defaultConfigFile = "config"
	}
	flag.StringVar(&ConfigFile, "ConfigFile", defaultConfigFile, "配置文件名 / Config file name")
	flag.Parse()

	// 显式切到 dist 子目录，避免 Wails 3 资源根目录解析错位。
	// Explicitly serve the dist subtree so Wails 3 resolves asset paths from the frontend bundle root.
	distAssets, err := fs.Sub(assets, "frontend/dist")
	if err != nil {
		panic(err)
	}

	// BundledAssetFileServer 在生产模式服务内嵌资源，在开发模式可自动接管 FRONTEND_DEVSERVER_URL。
	// BundledAssetFileServer serves embedded assets in production and automatically proxies FRONTEND_DEVSERVER_URL in development.
	desktopApp := wailsapp.NewDesktopApp(ConfigFile)
	app := application.New(application.Options{
		Name:        "dtool",
		Description: "dtool desktop client",
		Assets: application.AssetOptions{
			Handler: application.BundledAssetFileServer(distAssets),
		},
		OnShutdown: desktopApp.Shutdown,
	})

	primaryScreen := app.Screen.GetPrimary()
	windowHeight := getDesktopWindowHeight(primaryScreen, desktopWindowMinHeight)

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "dtool",
		Width:            1400,
		Height:           windowHeight,
		MinWidth:         1100,
		MinHeight:        desktopWindowMinHeight,
		BackgroundColour: application.NewRGBA(255, 255, 255, 255),
		URL:              "/",
	})
	desktopApp.BindRuntime(app, window)
	window.OnWindowEvent(events.Common.WindowRuntimeReady, func(_ *application.WindowEvent) {
		desktopApp.DomReady()
	})

	if err := app.Run(); err != nil {
		panic(err)
	}
}

// getDesktopWindowHeight 按主屏工作区高度计算桌面窗口高度，目标为 90%。 // Calculates the desktop window height from the primary screen work area, targeting 90%.
func getDesktopWindowHeight(primaryScreen *application.Screen, minHeight int) int {
	if primaryScreen == nil || primaryScreen.WorkArea.Height <= 0 {
		return desktopWindowDefaultHeight
	}
	targetHeight := primaryScreen.WorkArea.Height * 9 / 10
	// 保证窗口高度不会低于最小高度，避免小屏场景下窗口初始化过小。 // Keep the initial window height above the minimum height to avoid undersized startup windows on small screens.
	if targetHeight < minHeight {
		return minHeight
	}
	return targetHeight
}
