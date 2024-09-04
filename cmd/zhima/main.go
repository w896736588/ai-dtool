package main

import (
	"dev_tool/internal/app/zhima"
	"gitee.com/Sxiaobai/gs/gstool"
)

var IsBuild string

func main() {
	zhima.InitBase(IsBuild)
	gstool.CpuSetUsePercent(0.6)
	gstool.SignalDefault()
}
