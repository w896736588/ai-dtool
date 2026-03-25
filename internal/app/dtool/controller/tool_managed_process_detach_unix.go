//go:build linux || darwin

package controller

import (
	"os/exec"
	"syscall"
)

func prepareManagedProcessCommand(cmd *exec.Cmd) {
	// 创建独立会话与进程组 / Create a new session and process group.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Setsid: true,
	}
}
