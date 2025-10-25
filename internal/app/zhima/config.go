package zhima

import (
	"dev_tool/base"
	_default "dev_tool/internal/app/default"
	"fmt"
	"os"
	"time"

	"gitee.com/Sxiaobai/gs/gsencrypt"
	"gitee.com/Sxiaobai/gs/gstask"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/cast"
)

var AppName = `zhima`

func InitBase(DbPath, DbName, ViewPath string) {
	_default.InitBase(AppName, DbPath, DbName, ViewPath)
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
	task := gstask.NewTask()
	for key, tGin := range base.Component.TGins {
		task.Add(gstask.CallbackFunc{
			Id: cast.ToString(key),
			Func: func() *gstask.Result {
				_ = tGin.GinStop(1)
				return &gstask.Result{
					Result: nil,
					Err:    nil,
				}
			},
			Timeout: time.Second * 1,
		})
	}
	task.RunAll()
	_ = base.Component.TPlaywright.Log.Close()
	_ = base.Component.TVariable.Log.Close()
	_ = base.Component.GsLog.Close()
}
