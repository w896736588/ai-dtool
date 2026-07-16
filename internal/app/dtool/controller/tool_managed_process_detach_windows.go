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
	// 允许子进程脱离父进程所在的 Job Object / Allow the child process to break away from the parent's job object.
	// 确保父进程退出时子进程不会被连带终止 / Ensures child is not terminated when parent exits.
	managedProcessWindowsBreakawayFromJob = 0x01000000
	// 创建独立进程组并脱离控制台和父进程 Job / Create isolated process group, detach from console and parent job.
	managedProcessWindowsCreationFlags = syscall.CREATE_NEW_PROCESS_GROUP | managedProcessWindowsDetachedProcess | managedProcessWindowsCreateNoWindow | managedProcessWindowsBreakawayFromJob
)

func prepareManagedProcessCommand(cmd *exec.Cmd) {
	// 创建独立进程组并脱离控制台和父进程 Job / Create isolated process group, detach from console and parent job.
	cmd.SysProcAttr = &syscall.SysProcAttr{
		CreationFlags: managedProcessWindowsCreationFlags,
		HideWindow:    true,
	}
}
