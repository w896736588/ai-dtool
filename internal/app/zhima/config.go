package zhima

import (
	"dev_tool/base"
	_default "dev_tool/internal/app/default"
	"fmt"
	"os"
	"time"

	"gitee.com/Sxiaobai/gs/gsencrypt"
	"gitee.com/Sxiaobai/gs/gstool"
)

var AppName = `zhima`

func InitBase(DbPath, ViewPath string) {
	_default.InitBase(AppName, DbPath, ViewPath)
	initComponent()
}

func initComponent() {
	base.Component.AesGcm = gsencrypt.NewAesGcm(AppName)
	base.Component.EncryptDesCbc = &gsencrypt.DesCbc{
		Key: base.Component.ConfigViper.GetString(`encrypt.key`),
		Iv:  base.Component.ConfigViper.GetString(`encrypt.iv`),
	}
	for _, tGin := range base.Component.TGins {
		if tGin.IsRun == true {
			initRouter(tGin)
			tGin.GinRun()
		} else {
			gstool.FmtPrintlnLogTime(`5秒钟后退出`)
			time.Sleep(5 * time.Second)
			os.Exit(0)
		}
	}

}

func Stop() {
	fmt.Println(`停止`)
	for _, tGin := range base.Component.TGins {
		_ = tGin.GinStop(1)
	}
	_ = base.Component.TPlaywright.Log.Close()
	_ = base.Component.TVariable.Log.Close()
	_ = base.Component.GsLog.Close()
}
