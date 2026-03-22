//go:build windows

package controller

import (
	"os/exec"
	"syscall"
)

const (
	// Windows 未导出 DETACHED_PROCESS，按 Win32 常量定义补齐 / Mirror Win32 DETACHED_PROCESS constant locally.
	managedProcessWindowsDetachedProcess = 0x00000008
	// 创建独立进程组并脱离控制台 / Create a new process group and detach from parent console.
	managedProcessWindowsCreationFlags = syscall.CREATE_NEW_PROCESS_GROUP | managedProcessWindowsDetachedProcess
)

func prepareManagedProcessCommand(cmd *exec.Cmd) {
	// 创建独立进程组并脱离控制台 / Create a new process group and detach from parent console.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: managedProcessWindowsCreationFlags,
		HideWindow:    true,
	}
}
