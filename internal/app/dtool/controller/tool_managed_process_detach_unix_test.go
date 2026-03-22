//go:build linux || darwin

package controller

import (
	"os/exec"
	"testing"
)

func TestPrepareManagedProcessCommandSetsDetachAttrs(t *testing.T) {
	cmd := exec.Command(`sh`)

	prepareManagedProcessCommand(cmd)

	if cmd.SysProcAttr == nil {
		t.Fatal("SysProcAttr should not be nil")
	}
	if !cmd.SysProcAttr.Setsid {
		t.Fatal("Setsid should be true on unix")
	}
}
