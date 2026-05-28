package component

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	_struct "dev_tool/internal/app/dtool/struct"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_db"
	"dev_tool/internal/pkg/p_gin"
	"dev_tool/internal/pkg/p_shell"
	"dev_tool/internal/pkg/p_sse"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/viper"
)

// DataBaseUpRunner 抽象数据库迁移执行器，避免 component 直接依赖 business 产生包循环。
type DataBaseUpRunner interface {
	Run()
}

// VariableRuntime 抽象变量运行时入口，避免 component 反向依赖 variable 包实现细节。
type VariableRuntime interface {
	CreateTask(taskId string)
	IsStop(taskId string) bool
	ExistReplaceParam(data string) bool
	ExistReplaceParamFull(data string) bool
	ParseIdContent(str string) (int, string, error)
	AddReplace(replaceList map[string]string, key, value string)
	RegisterAllGlobal(replaceList map[string]string, sse *p_sse.SseShell, call *p_common.Call)
	ChecksCanDo(cmd map[string]any) bool
	PreConnSsh(sshId int, sshUniqueKey, sftpUniqueKey string, sse *p_sse.SseShell, call *p_common.Call) error
	SelectChooseReplace(variableForm *_struct.VForm, replaceList map[string]string, chooseValue string)
	ParseConfig(config string, call *p_common.Call) (string, error)
	GetLog() *gstool.GsSlog
	SetLoginCredentials(username, password string)
	ClearLoginCredentials()
	GetLoginUsername() string
	GetLoginPassword() string
}

var ShellClient *p_shell.Shell
var TGins []*p_gin.Gin
var MysqlClient *p_db.TMysql
var PgsqlClient *p_db.TPgsql
var RedisClient *p_db.TRedis
var SqliteClient *gsdb.GsSqlite
var LogSqliteClient *gsdb.GsSqlite

// 这几个实例迁移到 component 作为统一入口，方便初始化和跨模块访问保持一致。
var DbMain *common.CSqlite
var DbLog *common.CSqlite
var DataBaseUp DataBaseUpRunner
var VariableClient VariableRuntime
var ShellOutClient *common.TShellOut
var MemoryRuntime *common.MemoryStore
var CronSchedulers map[string]*common.CronScheduler

// CronTaskFuncRegistry 存储定时任务类型到执行函数的映射，由 controller 在初始化时注册。
var CronTaskFuncRegistry = make(map[string]func())
var EnvClient *define.Env
var ConfigViper *viper.Viper
var GsLog *gstool.GsSlog
var PlaywrightClient *TPlaywright
