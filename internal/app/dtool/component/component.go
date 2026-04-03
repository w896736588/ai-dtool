package component

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_db"
	"dev_tool/internal/pkg/p_gin"
	"dev_tool/internal/pkg/p_shell"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/viper"
)

var ShellClient *p_shell.Shell
var TGins []*p_gin.Gin
var MysqlClient *p_db.TMysql
var RedisClient *p_db.TRedis
var SqliteClient *gsdb.GsSqlite
var LogSqliteClient *gsdb.GsSqlite
var ShellOutClient *common.TShellOut
var MemoryRuntime *common.MemoryStore
var EnvClient *define.Env
var ConfigViper *viper.Viper
var GsLog *gstool.GsSlog
var PlaywrightClient *TPlaywright
