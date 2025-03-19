package _default

import (
	"dev_tool/base"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsencrypt"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gssocket"
	"gitee.com/Sxiaobai/gs/gsssh"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
)

var AppName = ``

func InitBase(IsBuild, appName, DbPath, ViewPath, WebData string) {
	AppName = appName
	initComponent(IsBuild, WebData)
	initSqlite(DbPath)
	initGin(ViewPath)
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

func initComponent(IsBuild, WebData string) {
	gstool.FmtPrintlnLogTime(`IsBuild %#v`, IsBuild)
	base.Component = base.TComponent{}
	base.Component.Env = &base.Env{}
	base.Component.TGin = &base.Gin{}
	base.Component.TShell = &base.TShell{ShellClientMap: make(map[string]*gsssh.SshConfig)}
	base.Component.TRedis = &base.TRedis{RedisClientMap: make(map[string]*gsdb.GsRedis)}
	base.Component.TRedis.PingAll()
	base.Component.TMysql = &base.TMysql{MysqlClientMap: make(map[string]*gsdb.GsMysql)}
	base.Component.TCode = &base.TCode{}
	base.Component.TBase = &base.TBase{
		StartMillUnix: gstool.TimeNowMilliInt64(),
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
	base.Component.Env.ConfigPath = base.Component.Env.RootPath + `/config/` + base.Component.Env.AppName
	base.Component.Env.LogPath = base.Component.Env.RootPath + `/logs`
	if WebData != `` {
		base.Component.Env.PlaywrightUserData = WebData + `/playwright_userdata`
		base.Component.Env.PlaywrightDownload = WebData + `/playwright_download`
	} else {
		base.Component.Env.PlaywrightUserData = base.Component.Env.RootPath + `/playwright_userdata`
		base.Component.Env.PlaywrightDownload = base.Component.Env.RootPath + `/playwright_download`

	}
	gstool.FmtPrintlnLogTime(`根目录 %s`, base.Component.Env.RootPath)
	gstool.FmtPrintlnLogTime(`加载配置文件 %s`, base.Component.Env.ConfigPath)
	//初始化playwright
	base.Component.TSmartLink = &base.TSmartLink{
		DownloadPath: base.Component.Env.PlaywrightDownload,
	}
	base.Component.TSmartLink.WitchDownload()
	base.Component.TSmartLink.SmartCheckAndUpdate()
	//配置初始化
	base.Component.ConfigViper = viper.New()
	base.Component.ConfigViper.AddConfigPath(base.Component.Env.ConfigPath)
	base.Component.ConfigViper.SetConfigName(`config`)
	base.Component.ConfigViper.SetConfigType(`ini`)
	if readErr := base.Component.ConfigViper.ReadInConfig(); readErr != nil {
		panic(readErr.Error())
	}
	base.Component.GsLog = gstool.SlogCreateDefault(base.Component.Env.LogPath, AppName)
}

func initSqlite(DbPath string) {
	dbDir := DbPath
	var dbPath string
	if dbDir != `` {
		dbPath = fmt.Sprintf(dbDir+`%s`, AppName+`.db`)
	} else {
		dbPath = base.Component.Env.RootPath + `/config/` + AppName + `/` + AppName + `.db`
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

func initGin(ViewPath string) {
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
	if ViewPath == `` {
		ViewPath = filepath.Dir(base.Component.Env.RootPath) + `/devtool/dist`
	}
	gstool.FmtPrintlnLogTime(`前端目录 %s`, ViewPath)
	base.Component.TGin.GinStatic(`/js`, ViewPath+`/js`)
	base.Component.TGin.GinStaticFile(`/favicon.ico`, ViewPath+`/favicon.ico`)
	base.Component.TGin.GinStatic(`/css`, ViewPath+`/css`)
	base.Component.TGin.GinLoadHTMLFiles(ViewPath + `/index.html`)
	base.Component.TGin.GinGet(`/`, func(context *gin.Context) {
		context.HTML(200, `index.html`, nil)
	})
	base.Component.TSse = &gsgin.TSse{
		SseList: make(map[string]*gsgin.Sse),
	}
	base.Component.TGin.SseRoute(`/sse`, func(urlValues url.Values, stopC chan int, c *gin.Context) *gsgin.Sse {
		clientId := urlValues.Get(`client_id`)
		return base.Component.TSse.Register(clientId, stopC, c)
	}, func(sse *gsgin.Sse) {
		base.Component.TSse.Pause(sse)
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
