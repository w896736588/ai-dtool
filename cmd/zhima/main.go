package main

import (
	"dev_tool/internal/app/zhima"
	"gitee.com/Sxiaobai/gs/gstool"
)

var IsBuild string
var DbPath string

func main() {
	zhima.InitBase(IsBuild, DbPath)
	gstool.CpuSetUsePercent(0.6)
	gstool.SignalDefault()
}
