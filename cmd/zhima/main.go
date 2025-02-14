package main

import (
	"dev_tool/internal/app/zhima"
	"gitee.com/Sxiaobai/gs/gstool"
)

var IsBuild string
var DbPath string
var ViewPath string //前端代码目录
var WebData string  //浏览器设置存储目录

func main() {
	gstool.FmtPrintlnLogTime(`参数接收 IsBuild %s DbPath %s ViewPath %s WebData %s`, IsBuild, DbPath, ViewPath, WebData)
	zhima.InitBase(IsBuild, DbPath, ViewPath, WebData)
	gstool.CpuSetUsePercent(0.6)
	gstool.SignalDefault()
}
