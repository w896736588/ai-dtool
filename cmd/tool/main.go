package main

import (
	"dev_tool/internal/app/tool"
	"gitee.com/Sxiaobai/gs/gstool"
)

var IsBuild string

func main() {
	tool.InitBase(IsBuild)
	gstool.CpuSetUsePercent(0.6)
	gstool.SignalDefault()
}
