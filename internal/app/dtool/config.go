package dtool

import (
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/plw"
	"dev_tool/internal/app/dtool/variable"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_db"
	"dev_tool/internal/pkg/p_gin"
	"dev_tool/internal/pkg/p_shell"
	"dev_tool/internal/pkg/p_sse"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstask"
	"github.com/spf13/cast"

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
	component.EnvClient = &define.Env{}
	component.TGins = make([]*p_gin.Gin, 0)
	component.RedisClient = &p_db.TRedis{RedisClientMap: make(map[string]*gsdb.GsRedis)}
	component.RedisClient.PingAll(common.GetCall())
	component.MysqlClient = &p_db.TMysql{MysqlClientMap: make(map[string]*gsdb.GsMysql)}

	component.ConfigViper = viper.New()

	wd, _ := os.Getwd()
	var err error
	gstool.FmtPrintlnLogTime(`当前运行目录 %v`, wd)
	component.EnvClient.RootPath, err = gstool.GetRootPath(wd)
	if err != nil {
		panic(err.Error())
	}
	//初始化配置
	InitEnv(appName, ConfigFile, component.ConfigViper)
	component.EnvClient.DatabaseUpPath = filepath.Join(component.EnvClient.RootPath, `internal`, `app`, AppName, `database`)
	p_common.TBaseClient = &p_common.TBase{
		StartMillUnix: gstool.TimeNowMilliInt64(),
		LogPath:       component.EnvClient.LogPath,
	}
	//初始化shell
	component.ShellClient = p_shell.NewShell(component.EnvClient.LogPath)
	common.ShellOutClient = common.NewTShellOut()
	//aesGcm
	gcm := gsencrypt.NewAesGcm(component.EnvClient.AppName)
	p_common.AesGcmClient = gcm
	component.GsLog = gstool.NewSlog3(component.EnvClient.LogPath, component.EnvClient.AppName)
	_ = component.GsLog.CleanOldLogs(2)
}

func InitEnv(appName, ConfigFile string, viper *viper.Viper) {
	if component.EnvClient.RootPath == `` {
		panic(`root_path不能为空`)
	}
	component.EnvClient.AppName = appName
	if ConfigFile == `` {
		ConfigFile = `config`
	}
	component.EnvClient.ConfigFile = ConfigFile

	//基础
	component.EnvClient.ConfigPath = filepath.Join(component.EnvClient.RootPath, `config`, component.EnvClient.AppName)
	//配置初始化
	viper.AddConfigPath(component.EnvClient.ConfigPath)
	viper.SetConfigName(component.EnvClient.ConfigFile)
	viper.SetConfigType(`ini`)
	if readErr := viper.ReadInConfig(); readErr != nil {
		panic(readErr.Error())
	}
	component.EnvClient.PkgPath = filepath.Join(component.EnvClient.RootPath, `internal`, `pkg`)
	component.EnvClient.LogPath = filepath.Join(component.EnvClient.RootPath, `logs`)
	//webkit
	component.EnvClient.NodePath = `node`
	//base配置初始化
	component.EnvClient.ConfigBase = &define.Base{
		DbFileName: viper.GetString(`base.dbFileName`),
		DbPath:     viper.GetString(`base.dbPath`),
		WebPath:    viper.GetString(`base.webPath`),
	}
	//web
	component.EnvClient.WebConfig = &define.WebConfig{
		WebPath: ``,
	}
	//前端目录
	if component.EnvClient.ConfigBase.WebPath == `` {
		component.EnvClient.WebConfig.WebPath = filepath.Join(filepath.Dir(component.EnvClient.RootPath), `devtool`, `dist`)
	} else {
		component.EnvClient.WebConfig.WebPath = component.EnvClient.ConfigBase.WebPath
	}
	// webPath不存在时，优先兜底到当前项目的web/dist，避免桌面端无法加载页面
	if _, webPathErr := os.Stat(component.EnvClient.WebConfig.WebPath); webPathErr != nil {
		fallbackWebPath := filepath.Join(component.EnvClient.RootPath, `web`, `dist`)
		if _, fallbackErr := os.Stat(fallbackWebPath); fallbackErr == nil {
			component.EnvClient.WebConfig.WebPath = fallbackWebPath
		}
	}
	//数据库配置
	component.EnvClient.DbConfig = &define.DbConfig{
		DbName: ``,
		DbPath: component.EnvClient.ConfigBase.DbPath,
	}
	//数据库名
	component.EnvClient.DbConfig.DbName = component.EnvClient.AppName + `.db`
	if component.EnvClient.ConfigBase.DbFileName != `` {
		component.EnvClient.DbConfig.DbName = component.EnvClient.ConfigBase.DbFileName
	}
	//配置文件目录
	if component.EnvClient.DbConfig.DbPath == `` {
		component.EnvClient.DbConfig.DbPath = filepath.Join(component.EnvClient.RootPath, `config`, component.EnvClient.AppName)
	}
	//判断是否存在D盘如果没有那么就改为C盘
	drive := ``
	drivePath := string(`D`) + ":\\"
	_, err := os.Stat(drivePath)
	if err == nil {
		drive = `D`
	} else {
		drive = `C`
	}
	component.EnvClient.WebkitDriverPath = viper.GetString(`path.webkit_driver_path`)
	component.EnvClient.WebkitDataPath = viper.GetString(`path.webkit_data_path`)
	component.EnvClient.WebkitDownloadPath = viper.GetString(`path.webkit_download_path`)
	component.EnvClient.WebkitDataPath = gstool.SReplaces(component.EnvClient.WebkitDataPath, map[string]string{
		`{DRIVE}`: drive,
	})
	component.EnvClient.WebkitDownloadPath = gstool.SReplaces(component.EnvClient.WebkitDownloadPath, map[string]string{
		`{DRIVE}`: drive,
	})
	component.EnvClient.WebkitDriverPath = gstool.SReplaces(component.EnvClient.WebkitDriverPath, map[string]string{
		`{DRIVE}`: drive,
	})
	//创建目录
	_ = gstool.DirCreatePath(component.EnvClient.LogPath)
	_ = gstool.DirCreatePath(component.EnvClient.DbConfig.DbPath)
	_ = gstool.DirCreatePath(component.EnvClient.WebkitDataPath)
	_ = gstool.DirCreatePath(component.EnvClient.WebkitDriverPath)
	_ = gstool.DirCreatePath(component.EnvClient.WebkitDownloadPath)
	gstool.FmtPrintlnLogTime(`输出配置：`)
	gstool.FmtPrintlnLogTime(gstool.JsonFormat(component.EnvClient))
}

func initPlaywright() {
	//初始化playwright
	plw.PlaywrightClient = plw.NewTPlaywright()
	plw.PlaywrightClient.LockFileFullPath = filepath.Join(component.EnvClient.RootPath, `playwright.RunLock`)
	plw.InitPageActiveTime()
	if !plw.PlaywrightClient.EnsureNodeRuntime() {
		gstool.FmtPrintlnLogTime(`未检测到 Node.js，跳过 Playwright 初始化，等待用户安装后再使用自定义网页`)
		return
	}
	go plw.PlaywrightClient.WitchDownload()
	go plw.PlaywrightClient.SmartCheckAndUpdate(&p_sse.SseShell{})
}

func initSqlite() {
	fmt.Println(fmt.Sprintf(`配置库目录 %s`, component.EnvClient.DbConfig.DbPath))
	fmt.Println(fmt.Sprintf(`配置库路径 %s`, filepath.Join(component.EnvClient.DbConfig.DbPath, component.EnvClient.DbConfig.DbName)))
	var err error
	component.SqliteClient, err = p_db.InitSqlite(component.EnvClient.DbConfig.DbPath, component.EnvClient.DbConfig.DbName)
	if err != nil {
		panic(fmt.Sprintf(`连接sqlite失败 %s`, err.Error()))
	}
	p_db.InitMysql()
	common.DbMain = &common.CSqlite{Client: component.SqliteClient, Env: component.EnvClient}
	business.DataBaseUp = business.NewTDataBaseUp()
	business.DataBaseUp.Run()
	common.ShellOutClient.InitGroupConfigs()
}

func initGin() {
	host := component.ConfigViper.GetString(`run.host`)
	ports := strings.Split(component.ConfigViper.GetString(`run.ports`), `,`)
	component.EnvClient.Ports = ports
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
			tGin.GinStatic(`/js`, component.EnvClient.WebConfig.WebPath+`/js`)
			tGin.GinStaticFile(`/favicon.ico`, component.EnvClient.WebConfig.WebPath+`/favicon.ico`)
			tGin.GinStatic(`/css`, component.EnvClient.WebConfig.WebPath+`/css`)
			tGin.GinLoadHTMLFiles(component.EnvClient.WebConfig.WebPath + `/index.html`)
			tGin.GinGet(`/`, func(context *gin.Context) {
				context.HTML(200, `index.html`, nil)
			})
		}
		tGin.IsRun = true
		component.TGins = append(component.TGins, tGin)
	}
}

func initOther() {
	p_common.TOsClient = gstool.NewGsOs()
	p_common.TMarkDownClient = &p_common.TMarkDown{}
	p_common.TJasClient = &p_common.TJas{
		Regis: map[string]string{
			`p_js`: component.EnvClient.PkgPath + "/p_js",
		},
		JsData: map[string]string{},
	}
	p_common.TJasClient.Load()
	variable.VariableClient = variable.NewVariableClient()
}

func InitComponent() {
	p_common.AesGcmClient = gsencrypt.NewAesGcm(AppName)
	for _, tGin := range component.TGins {
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
	for key, tGin := range component.TGins {
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
	_ = component.GsLog.Close()
}
