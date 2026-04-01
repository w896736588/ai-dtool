package wailsapp

import (
	"testing"

	"github.com/wailsapp/wails/v3/pkg/application"
)

type fakeRedirectWindow struct {
	lastURL string
}

func (h *fakeRedirectWindow) SetURL(url string) application.Window {
	h.lastURL = url
	return nil
}

func TestOpenBackendURLRedirectsWindowToBackend(t *testing.T) {
	window := &fakeRedirectWindow{}
	app := &DesktopApp{
		window: window,
	}

	app.openBackendURL(`http://localhost:17170/`)

	if window.lastURL != `http://localhost:17170/` {
		t.Fatalf("lastURL = %q, want %q", window.lastURL, `http://localhost:17170/`)
	}
}
