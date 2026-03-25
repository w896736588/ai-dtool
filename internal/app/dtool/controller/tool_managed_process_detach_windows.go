//go:build windows

package controller

import (
	"os/exec"
	"syscall"
)

const (
	// Windows 未导出 DETACHED_PROCESS，按 Win32 常量定义补齐 / Mirror Win32 DETACHED_PROCESS constant locally.
	managedProcessWindowsDetachedProcess = 0x00000008
	// 创建无控制台窗口进程 / Create the process without attaching a visible console window.
	managedProcessWindowsCreateNoWindow = 0x08000000
	// 创建独立进程组并脱离控制台 / Create a new process group and detach from parent console.
	managedProcessWindowsCreationFlags = syscall.CREATE_NEW_PROCESS_GROUP | managedProcessWindowsDetachedProcess | managedProcessWindowsCreateNoWindow
)

func prepareManagedProcessCommand(cmd *exec.Cmd) {
	// 创建独立进程组并脱离控制台 / Create a new process group and detach from parent console.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: managedProcessWindowsCreationFlags,
		HideWindow:    true,
	}
}
