//go:build windows

package controller

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

func TestPrepareManagedProcessCommandSetsDetachAttrs(t *testing.T) {
	cmd := exec.Command(`cmd`)

	prepareManagedProcessCommand(cmd)

	if cmd.SysProcAttr == nil {
		t.Fatal("SysProcAttr should not be nil")
	}
	if cmd.SysProcAttr.CreationFlags != managedProcessWindowsCreationFlags {
		t.Fatalf("CreationFlags = %d, want %d", cmd.SysProcAttr.CreationFlags, managedProcessWindowsCreationFlags)
	}
	if !cmd.SysProcAttr.HideWindow {
		t.Fatal("HideWindow should be true on windows")
	}
}

func TestResolveManagedProcessStartConfigPrefersRealExe(t *testing.T) {
	dir := t.TempDir()
	commandName := `managed-test-tool`
	shimPath := filepath.Join(dir, commandName+`.cmd`)
	realExePath := filepath.Join(dir, `node_modules`, commandName, `bin`, commandName+`.exe`)

	if err := os.WriteFile(shimPath, []byte(`shim`), 0o644); err != nil {
		t.Fatalf("WriteFile shim error = %v", err)
	}
	if err := os.MkdirAll(filepath.Dir(realExePath), 0o755); err != nil {
		t.Fatalf("MkdirAll error = %v", err)
	}
	if err := os.WriteFile(realExePath, []byte(`exe`), 0o644); err != nil {
		t.Fatalf("WriteFile exe error = %v", err)
	}

	oldPath := os.Getenv(`PATH`)
	if err := os.Setenv(`PATH`, dir+string(os.PathListSeparator)+oldPath); err != nil {
		t.Fatalf("Setenv PATH error = %v", err)
	}
	defer func() {
		_ = os.Setenv(`PATH`, oldPath)
	}()

	resolved, err := resolveManagedProcessStartConfig(managedProcessConfig{
		Executable: commandName,
		Args:       []string{`--config`, `C:\Users\94804\.cc-connect\config.toml`},
	})
	if err != nil {
		t.Fatalf("resolveManagedProcessStartConfig error = %v", err)
	}
	if filepath.Clean(resolved.Executable) != filepath.Clean(realExePath) {
		t.Fatalf("resolved executable = %q, want %q", resolved.Executable, realExePath)
	}
}
