package main

import (
	"dev_tool/internal/app/zhima"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

var ViewPath string

func main() {
	zhima.InitBase(ViewPath)
	gstool.CpuSetUsePercent(0.6)
	gstool.SignalDefault()
	zhima.Stop()
}
