package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

func TestStartupSplashPageLoadsWailsRuntime(t *testing.T) {
	indexPath := filepath.Join(`frontend`, `dist`, `index.html`)
	content, err := os.ReadFile(indexPath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error = %v", indexPath, err)
	}

	if !strings.Contains(string(content), `/wails/runtime.js`) {
		t.Fatalf("startup splash page should load /wails/runtime.js so WindowRuntimeReady can fire")
	}
}

func TestGetDesktopWindowHeightUsesNinetyPercentOfPrimaryWorkArea(t *testing.T) {
	screen := &application.Screen{
		WorkArea: application.Rect{
			Height: 1000,
		},
	}

	got := getDesktopWindowHeight(screen, desktopWindowMinHeight)
	if got != 900 {
		t.Fatalf("getDesktopWindowHeight() = %d, want %d", got, 900)
	}
}

func TestGetDesktopWindowHeightRespectsMinimumHeight(t *testing.T) {
	screen := &application.Screen{
		WorkArea: application.Rect{
			Height: 720,
		},
	}

	got := getDesktopWindowHeight(screen, desktopWindowMinHeight)
	if got != desktopWindowMinHeight {
		t.Fatalf("getDesktopWindowHeight() = %d, want %d", got, desktopWindowMinHeight)
	}
}
