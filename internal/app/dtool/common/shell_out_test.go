package common

import "testing"

func TestNewTShellOutUsesProvidedLogPath(t *testing.T) {
	client := NewTShellOut(t.TempDir())
	if client == nil {
		t.Fatal("NewTShellOut() returned nil")
	}
	if client.log == nil {
		t.Fatal("NewTShellOut() returned client without logger")
	}
	t.Cleanup(func() {
		_ = client.log.Close()
	})
}
