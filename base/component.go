package base

import (
	"gitee.com/Sxiaobai/gs/v2/gsencrypt"
	"gitee.com/Sxiaobai/gs/v2/gssocket"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/viper"
)

var Component TComponent

type TComponent struct {
	TSqlite       *TSqlite
	GsLog         *gstool.GsSlog
	EncryptDesCbc *gsencrypt.DesCbc
	Env           *Env
	TGins         []*Gin
	WebSocket     *gssocket.Server
	TShell        *TShell
	TShellOut     *TShellOut
	TCode         *TCode
	TBase         *TBase
	AesGcm        *gsencrypt.AesGcm
	ConfigViper   *viper.Viper
	TRedis        *TRedis
	TMysql        *TMysql
	TSocket       *TSocket
	TPlaywright   *TPlaywright
	TSse          *TSse
	TOs           *gstool.GsOs
	TMarkDown     *TMarkDown
	TAi           *TAi
	TVariable     *TVariable
	TJas          *TJas
	TDataBaseUp   *TDataBaseUp
}
