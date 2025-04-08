package base

import (
	"gitee.com/Sxiaobai/gs/gsencrypt"
	"gitee.com/Sxiaobai/gs/gssocket"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/spf13/viper"
)

var Component TComponent

type TComponent struct {
	TSqlite       *TSqlite
	GsLog         *gstool.GsSlog
	EncryptDesCbc *gsencrypt.DesCbc
	Env           *Env
	TGin          *Gin
	WebSocket     *gssocket.Server
	TShell        *TShell
	TCode         *TCode
	TBase         *TBase
	AesGcm        *gsencrypt.AesGcm
	ConfigViper   *viper.Viper
	TRedis        *TRedis
	TMysql        *TMysql
	TSocket       *TSocket
	TSmartLink    *TSmartLink
	TSse          *TSse
	TOs           *gstool.GsOs
	TMarkDown     *TMarkDown
	TAi           *TAi
}
