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

	desktopApp := wailsapp.NewDesktopApp(ConfigFile)
	app := application.New(application.Options{
		Name:        "dtool",
		Description: "dtool desktop client",
		Assets: application.AssetOptions{
			Handler: application.AssetFileServerFS(distAssets),
		},
		OnShutdown: desktopApp.Shutdown,
	})

	window := app.Window.NewWithOptions(application.WebviewWindowOptions{
		Title:            "dtool",
		Width:            1400,
		Height:           900,
		MinWidth:         1100,
		MinHeight:        700,
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
