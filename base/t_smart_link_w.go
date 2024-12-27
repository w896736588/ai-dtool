package base

import (
	"github.com/go-vgo/robotgo"
)

// SetForegroundWindowPid 设置活跃
func (h *TSmartLink) SetForegroundWindowPid(pid int) error {
	return robotgo.ActivePid(pid)
}

// GetActiveWindowPid 获取当前活跃窗口的PID
func (h *TSmartLink) GetActiveWindowPid() int {
	return robotgo.GetPid()
}

// SetWindowMax 设置窗口最大化
func (h *TSmartLink) SetWindowMax(pid int) {
	robotgo.MaxWindow(pid)
}
