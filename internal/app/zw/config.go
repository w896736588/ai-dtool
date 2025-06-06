package zw

import (
	"dev_tool/base"
	_default "dev_tool/internal/app/default"
	"gitee.com/Sxiaobai/gs/gsencrypt"
	"gitee.com/Sxiaobai/gs/gstool"
	"os"
	"time"
)

var AppName = `zw`

func InitBase(IsBuild, DbPath, ViewPath string) {
	_default.InitBase(IsBuild, AppName, DbPath, ViewPath)
	initComponent()
}

func initComponent() {
	base.Component.AesGcm = gsencrypt.NewAesGcm(AppName)
	base.Component.EncryptDesCbc = &gsencrypt.DesCbc{
		Key: base.Component.ConfigViper.GetString(`encrypt.key`),
		Iv:  base.Component.ConfigViper.GetString(`encrypt.iv`),
	}
	if base.Component.TGin.IsRun == true {
		initRouter()
		base.Component.TGin.GinRun()
	} else {
		gstool.FmtPrintlnLogTime(`5秒钟后退出`)
		time.Sleep(5 * time.Second)
		os.Exit(0)
	}
}

func Stop() {
	base.Component.GsLog.Debugf(`停止`)
	_ = base.Component.TGin.GinStop(5)
	_ = base.Component.TPlaywright.Log.Close()
	_ = base.Component.TVariable.Log.Close()
	_ = base.Component.GsLog.Close()
}
