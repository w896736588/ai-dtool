package redis_manager

import (
	"fmt"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/go-vgo/robotgo"
	"syscall"
	"testing"
	"time"
	"unsafe"
)

const (
	_WM_SYSCOMMAND = 0x0112
	_SC_RESTORE    = 0xF120
	_SC_ACTIVATE   = 0xF100
)

func postRestoreAndActivate(hwnd syscall.Handle) error {
	_, _, err := procPostMessage.Call(
		uintptr(hwnd),
		uintptr(_WM_SYSCOMMAND),
		uintptr(_SC_RESTORE|_SC_ACTIVATE),
		0,
	)
	if err != nil {
		return err
	}
	return nil
}

// 加载必要的Windows API函数
var (
	moduser32                    = syscall.NewLazyDLL("user32.dll")
	procEnumWindows              = moduser32.NewProc("EnumWindows")
	procGetWindowThreadProcessId = moduser32.NewProc("GetWindowThreadProcessId")
	procGetWindowTextW           = moduser32.NewProc("GetWindowTextW")
	procShowWindow               = moduser32.NewProc("ShowWindow")
	setForegroundWindow          = moduser32.NewProc("SetForegroundWindow")
	procPostMessage              = moduser32.NewProc("PostMessageW")
)

const (
	SW_RESTORE = 9
)

var foundHwnd syscall.Handle

type ENUMWINDOWSPROC func(hwnd syscall.Handle, lParam uintptr) bool

// EnumWindows callback function type
type callbackFuncType uintptr

// 回调函数的C签名
func callback(hwnd syscall.Handle, lParam uintptr) uintptr {
	var pid uint32
	_, _, _ = procGetWindowThreadProcessId.Call(uintptr(hwnd), uintptr(unsafe.Pointer(&pid)))
	if pid == uint32(lParam) {
		// 找到了匹配的PID，返回0以停止枚举
		gstool.FmtPrintlnLogTime(`找到了 %v %v`, lParam, hwnd)
		foundHwnd = hwnd
		return 0
	}
	// 继续枚举
	return 1
}

// 根据PID获取HWND
func getHWNDByPID(pid int32) (syscall.Handle, error) {
	cb := callbackFuncType(syscall.NewCallback(callback))
	ret, _, err := procEnumWindows.Call(uintptr(cb), uintptr(pid))
	if ret == 0 {
		return 0, err
	}
	// 如果回调没有返回0（即没有找到匹配的PID），则假设没有找到窗口
	// 注意：这里应该有一个更好的错误处理机制，比如区分"没有找到窗口"和"API调用失败"的情况
	return 0, fmt.Errorf("window not found for PID %d", pid)
}

func showWindow(hwnd syscall.Handle, nCmdShow int32) error {
	r, _, err := procShowWindow.Call(uintptr(hwnd), uintptr(nCmdShow))
	if r == 0 {
		return err
	}
	return nil
}

func TestTest(t *testing.T) {
	_, err := getHWNDByPID(9892)
	if err != nil {
		gstool.FmtPrintlnLogTime(`der %s`, err.Error())
	}
	gstool.FmtPrintlnLogTime(`句柄 %v`, foundHwnd)
	robotgo.MaxWindow(9892)
	time.Sleep(time.Second)
	err = showWindow(foundHwnd, SW_RESTORE)
	if err != nil {
		gstool.FmtPrintlnLogTime(`sss %s`, err.Error())
	}
	postRestoreAndActivate(foundHwnd)
	time.Sleep(time.Second)
	ret, _, errS := setForegroundWindow.Call(uintptr(foundHwnd))
	if errS != nil {
		gstool.FmtPrintlnLogTime(`s %s`, errS.Error())
	}
	gstool.FmtPrintlnLogTime(`ret %v`, ret)
	//robotgo.ActivePid(14052)
	return
}
