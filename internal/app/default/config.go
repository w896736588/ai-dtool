package _default

import (
	"dev_tool/base"
	"dev_tool/internal/pkg/p_playwright"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"gitee.com/Sxiaobai/gs/gsdb"
	"gitee.com/Sxiaobai/gs/gsencrypt"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gssocket"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/spf13/viper"
)

func InitBase(appName, ViewPath string) {
	initComponent(appName, ViewPath)
	initSqlite()
	initGin()
	initOther()
	initPlaywright()
	stdLog()
}

// 如果是编译后运行 那么将所有标准输出和报错重定向到 日志文件
func stdLog() {
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

func initComponent(appName, ViewPath string) {
	base.Component = base.TComponent{}
	base.Component.Env = &base.Env{}
	base.Component.TGins = make([]*base.Gin, 0)
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
	base.Component.ConfigViper = viper.New()

	wd, _ := os.Getwd()
	var err error
	gstool.FmtPrintlnLogTime(`当前运行目录 %v`, wd)
	base.Component.Env.RootPath, err = gstool.GetRootPath(wd)
	if err != nil {
		panic(err.Error())
	}
	//初始化配置
	base.Component.Env.Init(appName)
	base.Component.Env.DatabaseUpPath = filepath.Join(base.Component.Env.RootPath, `internal`, `app`, `default`, `database`)
	//初始化shell
	base.Component.TShell = base.NewTShell()
	base.Component.TShellOut = base.NewTShellOut()
	//aesGcm
	gcm := gsencrypt.NewAesGcm(base.Component.Env.AppName)
	base.Component.AesGcm = gcm
	base.Component.GsLog = gstool.NewSlog3(base.Component.Env.LogPath, base.Component.Env.AppName)
	_ = base.Component.GsLog.CleanOldLogs(2)
}

func initPlaywright() {
	//初始化playwright
	base.Component.TPlaywright = base.NewTSmartLink()
	base.Component.TPlaywright.SetWebkitPath()
	base.Component.TPlaywright.LockFileFullPath = filepath.Join(base.Component.Env.RootPath, `playwright.RunLock`)
	p_playwright.InitPageActiveTime()
	go base.Component.TPlaywright.WitchDownload()
	go base.Component.TPlaywright.SmartCheckAndUpdate()
}

func initSqlite() {
	fmt.Println(fmt.Sprintf(`配置库目录 %s`, base.Component.Env.DbConfig.DbPath))
	fmt.Println(fmt.Sprintf(`配置库路径 %s`, filepath.Join(base.Component.Env.DbConfig.DbPath, base.Component.Env.DbConfig.DbName)))
	sqlite, err := gsdb.NewSqlite(filepath.Join(base.Component.Env.DbConfig.DbPath, base.Component.Env.DbConfig.DbName), true)
	if err != nil {
		panic(fmt.Sprintf(`连接sqlite失败 %s`, err.Error()))
	}
	sqlite.SetGsLog(base.Component.GsLog)
	//sqlite.OpenDebug()
	createErr := sqlite.CreateConn()
	if createErr != nil {
		panic(fmt.Sprintf(`打开sqlite失败 %s`, createErr.Error()))
	}
	base.Component.TSqlite = &base.TSqlite{Client: sqlite, Env: base.Component.Env}
	base.Component.TDataBaseUp.Run()
	base.Component.TShellOut.InitGroupConfigs()
}

func Stop() {
	for _, tGin := range base.Component.TGins {
		err := tGin.GinStop(10)
		if err != nil {
			base.Component.GsLog.Errof(fmt.Sprintf(`关闭gin失败%s`, err.Error()))
		}
	}

}

func initGin() {
	host := base.Component.ConfigViper.GetString(`run.host`)
	ports := strings.Split(base.Component.ConfigViper.GetString(`run.ports`), `,`)
	base.Component.Env.Ports = ports
	gin.DefaultWriter = io.Discard
	for key, port := range ports {
		if !gstool.NetIsPortAvailable(host + `:` + port) {
			gstool.FmtPrintlnLogTime(`端口已被占用 %s`, host+`:`+port)
			return
		}
		tGin := &base.Gin{}
		tGin.SetMode(gin.DebugMode)
		tGin.GinInit(host, port)
		tGin.GinSetAllowCrossDomain()
		//第一个加载前端
		if key == 0 {
			tGin.GinStatic(`/js`, base.Component.Env.WebConfig.WebPath+`/js`)
			tGin.GinStaticFile(`/favicon.ico`, base.Component.Env.WebConfig.WebPath+`/favicon.ico`)
			tGin.GinStatic(`/css`, base.Component.Env.WebConfig.WebPath+`/css`)
			tGin.GinLoadHTMLFiles(base.Component.Env.WebConfig.WebPath + `/index.html`)
			tGin.GinGet(`/`, func(context *gin.Context) {
				context.HTML(200, `index.html`, nil)
			})
		}
		tGin.IsRun = true
		base.Component.TGins = append(base.Component.TGins, tGin)
	}
}

func initOther() {
	base.Component.TSse = &base.TSse{
		Sse: &gsgin.TSse{SseList: make(map[string]*gsgin.Sse)},
	}
	base.Component.TOs = gstool.NewGsOs()
	base.Component.TMarkDown = &base.TMarkDown{}
	base.Component.TAi = &base.TAi{}
	base.Component.TAi.Init()
	base.Component.TJas = &base.TJas{
		Regis: map[string]string{
			`p_js`: base.Component.Env.PkgPath + "/p_js",
		},
		JsData: map[string]string{},
	}
	base.Component.TJas.Load()
	base.Component.TVariable = base.NewVariable()
	base.Component.TDataBaseUp = base.NewTDataBaseUp()
}

func initSocket() {
	base.Component.WebSocket = &gssocket.Server{
		Host:        fmt.Sprintf(`0.0.0.0:%s`, base.Component.ConfigViper.GetString(`run.wsPort`)),
		Uri:         `/socket`,
		AllowOrigin: true,
	}
	base.Component.WebSocket.GetClientFunc = func(r *http.Request) string {
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
