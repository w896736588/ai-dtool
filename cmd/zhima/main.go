package main

import (
	"dev_tool/internal/app/zhima"
	"gitee.com/Sxiaobai/gs/gstool"
)

var IsBuild string
var DbPath string
var ViewPath string

func main() {
	zhima.InitBase(IsBuild, DbPath, ViewPath)
	gstool.CpuSetUsePercent(0.6)
	gstool.SignalDefault()
	zhima.Stop()
}
