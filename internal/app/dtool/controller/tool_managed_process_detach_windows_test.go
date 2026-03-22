//go:build windows

package controller

import (
	"os/exec"
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
