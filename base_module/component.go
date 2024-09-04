package base_module

import (
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsencrypt"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
)

type Component struct {
	ShellClient       *gsssh.SshConfig
	RedisClient       *gsdb.GsRedis
	XkfMysqlClient    *gsdb.GsMysql
	AppUrlMysqlClient *gsdb.GsMysql
	Global            *Global
	ReqMap            map[string]any
	Logger            *gstool.GsSlog
	Encrypt           *gsencrypt.DesCbc
}
