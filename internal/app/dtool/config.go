package dtool

import (
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/plw"
	"dev_tool/internal/app/dtool/variable"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_db"
	"dev_tool/internal/pkg/p_gin"
	"dev_tool/internal/pkg/p_shell"
	"dev_tool/internal/pkg/p_sse"
	"fmt"
	"gitee.com/Sxiaobai/gs/v2/gstask"
	"github.com/spf13/cast"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"gitee.com/Sxiaobai/gs/v2/gsencrypt"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
)

const AppName = `dtool`

func InitBase(ConfigFile string) {
	initComponent(AppName, ConfigFile)
	initSqlite()
	initGin()
	initOther()
	initPlaywright()
	stdLog()
}

// 如果是编译后运行 那么将所有标准输出和报错重定向到 日志文件
func stdLog() {
	//outFile, outFileErr := os.OpenFile(common.EnvClient.RootPath+`/out.log`, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//if outFileErr != nil {
	//	gstool.FmtPrintlnLogTime("error opening file: %v", outFileErr)
	//}
	//gstool.FmtPrintlnLogTime(`标准输出文件 %s`, common.EnvClient.RootPath+`/out.log`)
	//gstool.FmtPrintlnLogTime(`错误输出文件 %s`, common.EnvClient.RootPath+`/err.log`)
	//errFile, errErr := os.OpenFile(common.EnvClient.RootPath+`/err.log`, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	//if errErr != nil {
	//	gstool.FmtPrintlnLogTime("error opening file: %v", errErr)
	//}
	//os.Stdout = outFile
	//os.Stderr = errFile
}

func initComponent(appName, ConfigFile string) {
	common.EnvClient = &common.Env{}
	p_gin.TGins = make([]*p_gin.Gin, 0)
	p_db.RedisClient = &p_db.TRedis{RedisClientMap: make(map[string]*gsdb.GsRedis)}
	p_db.RedisClient.PingAll(common.GetCall())
	p_db.MysqlClient = &p_db.TMysql{MysqlClientMap: make(map[string]*gsdb.GsMysql)}

	common.ConfigViper = viper.New()

	wd, _ := os.Getwd()
	var err error
	gstool.FmtPrintlnLogTime(`当前运行目录 %v`, wd)
	common.EnvClient.RootPath, err = gstool.GetRootPath(wd)
	if err != nil {
		panic(err.Error())
	}
	//初始化配置
	common.EnvClient.Init(appName, ConfigFile)
	common.EnvClient.DatabaseUpPath = filepath.Join(common.EnvClient.RootPath, `internal`, `app`, AppName, `database`)
	p_common.TBaseClient = &p_common.TBase{
		StartMillUnix: gstool.TimeNowMilliInt64(),
		LogPath:       common.EnvClient.LogPath,
	}
	//初始化shell
	p_shell.ShellClient = p_shell.NewShell(common.EnvClient.LogPath)
	common.ShellOutClient = common.NewTShellOut()
	//aesGcm
	gcm := gsencrypt.NewAesGcm(common.EnvClient.AppName)
	p_common.AesGcmClient = gcm
	common.GsLog = gstool.NewSlog3(common.EnvClient.LogPath, common.EnvClient.AppName)
	_ = common.GsLog.CleanOldLogs(2)
}

func initPlaywright() {
	//初始化playwright
	plw.PlaywrightClient = plw.NewTPlaywright()
	plw.PlaywrightClient.SetWebkitPath()
	plw.PlaywrightClient.LockFileFullPath = filepath.Join(common.EnvClient.RootPath, `playwright.RunLock`)
	plw.InitPageActiveTime()
	go plw.PlaywrightClient.WitchDownload()
	go plw.PlaywrightClient.SmartCheckAndUpdate(&p_sse.SseShell{})
}

func initSqlite() {
	fmt.Println(fmt.Sprintf(`配置库目录 %s`, common.EnvClient.DbConfig.DbPath))
	fmt.Println(fmt.Sprintf(`配置库路径 %s`, filepath.Join(common.EnvClient.DbConfig.DbPath, common.EnvClient.DbConfig.DbName)))
	p_db.InitSqlite(common.EnvClient.DbConfig.DbPath, common.EnvClient.DbConfig.DbName)
	common.DbMain = &common.CSqlite{Client: p_db.SqliteClient, Env: common.EnvClient}
	business.DataBaseUp = business.NewTDataBaseUp()
	business.DataBaseUp.Run()
	common.ShellOutClient.InitGroupConfigs()
}

func initGin() {
	host := common.ConfigViper.GetString(`run.host`)
	ports := strings.Split(common.ConfigViper.GetString(`run.ports`), `,`)
	common.EnvClient.Ports = ports
	gin.DefaultWriter = io.Discard
	for key, port := range ports {
		if !gstool.NetIsPortAvailable(host + `:` + port) {
			gstool.FmtPrintlnLogTime(`端口已被占用 %s`, host+`:`+port)
			return
		}
		tGin := &p_gin.Gin{}
		tGin.SetMode(gin.DebugMode)
		tGin.GinInit(host, port)
		tGin.GinSetAllowCrossDomain()
		//第一个加载前端
		if key == 0 {
			tGin.GinStatic(`/js`, common.EnvClient.WebConfig.WebPath+`/js`)
			tGin.GinStaticFile(`/favicon.ico`, common.EnvClient.WebConfig.WebPath+`/favicon.ico`)
			tGin.GinStatic(`/css`, common.EnvClient.WebConfig.WebPath+`/css`)
			tGin.GinLoadHTMLFiles(common.EnvClient.WebConfig.WebPath + `/index.html`)
			tGin.GinGet(`/`, func(context *gin.Context) {
				context.HTML(200, `index.html`, nil)
			})
		}
		tGin.IsRun = true
		p_gin.TGins = append(p_gin.TGins, tGin)
	}
}

func initOther() {
	p_common.TOsClient = gstool.NewGsOs()
	p_common.TMarkDownClient = &p_common.TMarkDown{}
	p_common.TJasClient = &p_common.TJas{
		Regis: map[string]string{
			`p_js`: common.EnvClient.PkgPath + "/p_js",
		},
		JsData: map[string]string{},
	}
	p_common.TJasClient.Load()
	variable.VariableClient = variable.NewVariableClient()
}

func InitComponent() {
	p_common.AesGcmClient = gsencrypt.NewAesGcm(AppName)
	for _, tGin := range p_gin.TGins {
		if tGin.IsRun == true {
			InitRouter(tGin)
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
	for key, tGin := range p_gin.TGins {
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
	_ = plw.PlaywrightClient.Log.Close()
	_ = variable.VariableClient.Log.Close()
	_ = common.GsLog.Close()
}
