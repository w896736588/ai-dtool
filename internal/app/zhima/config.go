package zhima

import (
	"dev_tool/base"
	_default "dev_tool/internal/app/default"
	"gitee.com/Sxiaobai/gs/gsencrypt"
)

var AppName = `zhima`

func InitBase(IsBuild string) {
	_default.InitBase(IsBuild, AppName)
	initComponent()
}

func initComponent() {
	base.Component.AesGcm = gsencrypt.NewAesGcm(AppName)
	base.Component.EncryptDesCbc = &gsencrypt.DesCbc{
		Key: base.Component.ConfigViper.GetString(`encrypt.key`),
		Iv:  base.Component.ConfigViper.GetString(`encrypt.iv`),
	}
	initRouter()
	base.Component.TGin.GinRun()
}
