package main

import (
	"dev_tool/internal/app/zhima"
	"gitee.com/Sxiaobai/gs/gstool"
)

var DbPath string
var ViewPath string
var DbName string

func main() {
	zhima.InitBase(DbPath, DbName, ViewPath)
	gstool.CpuSetUsePercent(0.6)
	gstool.SignalDefault()
	zhima.Stop()
}
