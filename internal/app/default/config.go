package _default

import (
	"dev_tool/base"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsencrypt"
	"gitee.com/Sxiaobai/gs/gssocket"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
)

var AppName = ``

func InitBase(IsBuild, appName, DbPath string) {
	AppName = appName
	initComponent(IsBuild)
	initSqlite(DbPath)
	initGin()
	stdLog(IsBuild)
}

// 如果是编译后运行 那么将所有标准输出和报错重定向到 日志文件
func stdLog(IsBuild string) {
	if IsBuild != `1` {
		return
	}
	//outFile, outFileErr := os.OpenFile(base.Component.Env.RootPath+`/out.log`, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//if outFileErr != nil {
	//	gstool.FmtPrintlnLogTime("error opening file: %v", outFileErr)
	//}
	//gstool.FmtPrintlnLogTime(`标准输出文件 %s`, base.Component.Env.RootPath+`/out.log`)
	//gstool.FmtPrintlnLogTime(`错误输出文件 %s`, base.Component.Env.RootPath+`/err.log`)
	//errFile, errErr := os.OpenFile(base.Component.Env.RootPath+`/err.log`, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//if errErr != nil {
	//	gstool.FmtPrintlnLogTime("error opening file: %v", errErr)
	//}
	//os.Stdout = outFile
	//os.Stderr = errFile
}

func initComponent(IsBuild string) {
	gstool.FmtPrintlnLogTime(`IsBuild %#v`, IsBuild)
	base.Component = base.TComponent{}
	base.Component.Env = &base.Env{}
	base.Component.TGin = &base.Gin{}
	base.Component.TShell = &base.TShell{ShellClientMap: make(map[string]*gsssh.SshConfig)}
	base.Component.TRedis = &base.TRedis{RedisClientMap: make(map[string]*gsdb.GsRedis)}
	base.Component.TRedis.PingAll()
	base.Component.TMysql = &base.TMysql{MysqlClientMap: make(map[string]*gsdb.GsMysql)}
	base.Component.TCode = &base.TCode{}
	base.Component.TBase = &base.TBase{}
	base.Component.TSmartLink = &base.TSmartLink{
		PageList: make(map[string]*base.TPlayWright),
	}
	base.Component.TSocket = &base.TSocket{
		SocketList: make(map[string]*websocket.Conn),
	}
	base.Component.Env.IsBuild = IsBuild == `1`
	base.Component.Env.AppName = AppName
	gcm := gsencrypt.NewAesGcm(AppName)
	base.Component.AesGcm = gcm
	wd := ``
	gstool.FmtPrintlnLogTime(`运行模式 %v`, base.Component.Env.IsBuild)
	if base.Component.Env.IsBuild {
		wd, _ = os.Executable()
	} else {
		_, wd, _, _ = runtime.Caller(0)
	}
	var err error
	gstool.FmtPrintlnLogTime(`%v`, wd)
	base.Component.Env.RootPath, err = gstool.GetRootPath(wd)
	if err != nil {
		panic(err.Error())
	}
	gstool.FmtPrintlnLogTime(`根目录 %s`, base.Component.Env.RootPath)
	gstool.FmtPrintlnLogTime(`加载配置文件 %s`, base.Component.Env.RootPath+`/config/`+AppName)
	base.Component.ConfigViper = viper.New()
	base.Component.ConfigViper.AddConfigPath(base.Component.Env.RootPath + `/config/` + AppName)
	base.Component.ConfigViper.SetConfigName(`config`)
	base.Component.ConfigViper.SetConfigType(`ini`)
	if readErr := base.Component.ConfigViper.ReadInConfig(); readErr != nil {
		panic(readErr.Error())
	}
	base.Component.GsLog = gstool.SlogCreateDefault(base.Component.Env.RootPath+`/logs`, AppName)
}

func initSqlite(DbPath string) {
	dbDir := DbPath
	var dbPath string
	if dbDir != `` {
		dbPath = fmt.Sprintf(dbDir+`%s`, AppName+`.db`)
	} else {
		dbPath = base.Component.Env.RootPath + `/config/.db/` + AppName + `.db`
	}
	gstool.FmtPrintlnLogTime(`打开db %s`, dbPath)
	_ = gstool.DirCreatePath(dbDir)
	sqlite, err := gsdb.NewSqlite(dbPath, true)
	if err != nil {
		panic(fmt.Sprintf(`连接sqlite失败 %s`, err.Error()))
	}
	sqlite.SetGsLog(base.Component.GsLog)
	createErr := sqlite.CreateConn()
	if createErr != nil {
		panic(fmt.Sprintf(`打开sqlite失败 %s`, createErr.Error()))
	}
	base.Component.TSqlite = &base.TSqlite{Client: sqlite, Env: base.Component.Env}
	//检查表结构
	base.Component.TSqlite.InitTable()
}

func Stop() {
	err := base.Component.TGin.GinStop(10)
	if err != nil {
		base.Component.GsLog.Errof(fmt.Sprintf(`关闭gin失败%s`, err.Error()))
	}
}

func initGin() {
	host := base.Component.ConfigViper.GetString(`run.host`)
	port := base.Component.ConfigViper.GetString(`run.port`)
	if !gstool.NetIsPortAvailable(host + `:` + port) {
		gstool.FmtPrintlnLogTime(`端口已被占用 %s`, host+`:`+port)
		return
	}
	base.Component.TGin.SetMode(gin.TestMode)
	base.Component.TGin.GinInit(host, port)
	base.Component.TGin.GinSetAllowCrossDomain()
	gin.DefaultWriter = io.Discard
	viewPath := filepath.Dir(base.Component.Env.RootPath)
	if base.Component.Env.IsBuild {
		base.Component.TGin.GinStatic(`/js`, viewPath+`/devtool/dist/js`)
		base.Component.TGin.GinStaticFile(`/favicon.ico`, viewPath+`/devtool/dist/favicon.ico`)
		base.Component.TGin.GinStatic(`/css`, viewPath+`/devtool/dist/css`)
		base.Component.TGin.GinLoadHTMLFiles(viewPath + `/devtool/dist/index.html`)
	} else {
		base.Component.TGin.GinStatic(`/js`, base.Component.Env.RootPath+`/`+AppName+`/devtool/dist/js`)
		base.Component.TGin.GinStaticFile(`/favicon.ico`, base.Component.Env.RootPath+`/`+AppName+`/devtool/dist/favicon.ico`)
		base.Component.TGin.GinStatic(`/css`, base.Component.Env.RootPath+`/`+AppName+`/devtool/dist/css`)
		base.Component.TGin.GinLoadHTMLFiles(base.Component.Env.RootPath + `/` + AppName + `/devtool/dist/index.html`)
	}
	base.Component.TGin.GinGet(`/`, func(context *gin.Context) {
		context.HTML(200, `index.html`, nil)
	})
	base.Component.TGin.IsRun = true
}

func initSocket() {
	base.Component.WebSocket = &gssocket.Server{
		Host:        fmt.Sprintf(`0.0.0.0:%s`, base.Component.ConfigViper.GetString(`run.wsPort`)),
		Uri:         `/socket`,
		AllowOrigin: true,
	}
	base.Component.WebSocket.GetClientFunc = func(r *http.Request) string {
		gstool.FmtPrintlnLogTime(`%s`, r.FormValue(`uniqueKey`))
		return r.FormValue(`uniqueKey`)
	}
	base.Component.WebSocket.ReceMsgFunc = func(clientId, receiveMsg string) {

	}
	base.Component.WebSocket.ConnectFunc = func(clientId string, conn *websocket.Conn) {
		base.Component.TSocket.BindSsh(clientId, conn)
	}
	base.Component.WebSocket.CloseFunc = func(clientId string) {
		base.Component.TSocket.UnBindSsh(clientId)
	}
	go base.Component.WebSocket.Start()
}
