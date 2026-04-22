package dtool

import (
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/controller"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/plw"
	"dev_tool/internal/app/dtool/variable"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_db"
	"dev_tool/internal/pkg/p_gin"
	"dev_tool/internal/pkg/p_shell"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"fmt"
	"html/template"
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
	inicodec "github.com/go-viper/encoding/ini"
	"github.com/spf13/viper"
)

const AppName = `dtool`

const (
	// defaultDatabaseDirName 是默认数据库目录名。
	defaultDatabaseDirName = `database`
	// logDatabaseDirName 是 log 库迁移目录名。
	logDatabaseDirName = `database_log`
	// memoryDatabaseDirName 是记忆库迁移目录名。
	memoryDatabaseDirName = `database_memory`
	// logDatabaseNameSuffix 是 log 库文件名追加的后缀。
	logDatabaseNameSuffix = `.log`
	// databaseFileExt 是 sqlite 文件常用后缀。
	databaseFileExt = `.db`
)

var (
	// initComponentFunc 允许测试替换基础初始化流程 / allow tests to replace bootstrap initialization.
	initComponentFunc              = initComponent
	prepareMainDBStoreBeforeDBFunc = business.PrepareMainDBStore
	// prepareMemoryStoreBeforeDBFunc 允许测试校验记忆库预处理时机 / allow tests to verify memory preflight timing.
	prepareMemoryStoreBeforeDBFunc = business.PrepareMemoryStore
	// initSqliteFunc 允许测试替换数据库初始化流程 / allow tests to replace sqlite initialization.
	initSqliteFunc = initSqlite
	// initGinFunc 允许测试替换 gin 初始化流程 / allow tests to replace gin initialization.
	initGinFunc = initGin
	// initOtherFunc 允许测试替换其他组件初始化流程 / allow tests to replace other component initialization.
	initOtherFunc = initOther
	// initPlaywrightFunc 允许测试替换 Playwright 初始化流程 / allow tests to replace Playwright initialization.
	initPlaywrightFunc = initPlaywright
	// stdLogFunc 允许测试替换标准输出重定向流程 / allow tests to replace stdio redirection flow.
	stdLogFunc = stdLog
)

func formatEnvSummary(env *define.Env) string {
	if env == nil {
		return "配置摘要\n  未加载配置"
	}

	var builder strings.Builder
	builder.WriteString("配置摘要\n")

	writeSummarySection(&builder, "基础", [][2]string{
		{"应用", env.AppName},
		{"根目录", env.RootPath},
		{"配置文件", env.ConfigFile},
		{"配置目录", env.ConfigPath},
	})

	dbName := ""
	dbPath := ""
	dbFullPath := ""
	if env.DbConfig != nil {
		dbName = env.DbConfig.DbName
		dbPath = env.DbConfig.DbPath
	}
	if dbName != "" && dbPath != "" {
		dbFullPath = filepath.Join(dbPath, dbName)
	}
	writeSummarySection(&builder, "数据库", [][2]string{
		{"文件名", dbName},
		{"目录", dbPath},
		{"完整路径", dbFullPath},
		{"log库完整路径", formatLogDBFullPath(env)},
	})

	webPath := ""
	if env.WebConfig != nil {
		webPath = env.WebConfig.WebPath
	}
	writeSummarySection(&builder, "Web", [][2]string{
		{"目录", webPath},
	})

	writeSummarySection(&builder, "Playwright", [][2]string{
		{"Node", env.NodePath},
		{"Driver目录", env.WebkitDriverPath},
		{"下载目录", env.WebkitDownloadPath},
		{"数据目录", env.WebkitDataPath},
	})

	writeSummarySection(&builder, "日志", [][2]string{
		{"目录", env.LogPath},
	})

	return strings.TrimRight(builder.String(), "\n")
}

func writeSummarySection(builder *strings.Builder, title string, lines [][2]string) {
	filtered := make([][2]string, 0, len(lines))
	for _, line := range lines {
		if line[1] == "" {
			continue
		}
		filtered = append(filtered, line)
	}
	if len(filtered) == 0 {
		return
	}

	builder.WriteString("  [")
	builder.WriteString(title)
	builder.WriteString("]\n")
	for _, line := range filtered {
		builder.WriteString("  ")
		builder.WriteString(line[0])
		builder.WriteString(": ")
		builder.WriteString(line[1])
		builder.WriteString("\n")
	}
	builder.WriteString("\n")
}

func InitBase(ConfigFile string) {
	initComponentFunc(AppName, ConfigFile)
	if err := prepareMainDBStoreBeforeDBFunc(); err != nil {
		panic(err.Error())
	}
	// 记忆库若需要 git pull，必须先于所有数据库初始化 / memory git pull must happen before any database init.
	if err := prepareMemoryStoreBeforeDBFunc(); err != nil {
		panic(err.Error())
	}
	initSqliteFunc()
	initGinFunc()
	initOtherFunc()
	initPlaywrightFunc()
	stdLogFunc()
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
	component.MemoryRuntime = common.NewMemoryStore()
	component.MemoryRuntime.OnStatusChange = controller.BroadcastAsyncTasksUpdate
	component.MainDBAutoSyncRuntime = common.NewMainDBAutoSync()
	component.MainDBAutoSyncRuntime.OnStatusChange = controller.BroadcastAsyncTasksUpdate
	component.RedisClient = &p_db.TRedis{RedisClientMap: make(map[string]*gsdb.GsRedis)}
	component.RedisClient.PingAll(common.GetCall())
	component.MysqlClient = &p_db.TMysql{MysqlClientMap: make(map[string]*gsdb.GsMysql)}

	component.ConfigViper = newConfigViper()

	wd, _ := os.Getwd()
	var err error
	gstool.FmtPrintlnLogTime(`当前运行目录 %v`, wd)
	component.EnvClient.RootPath, err = gstool.GetRootPath(wd)
	if err != nil {
		panic(err.Error())
	}
	//初始化配置
	InitEnv(appName, ConfigFile, component.ConfigViper)
	component.EnvClient.DatabaseUpPath = filepath.Join(component.EnvClient.RootPath, `internal`, `app`, AppName, defaultDatabaseDirName)
	component.EnvClient.LogDatabaseUpPath = filepath.Join(component.EnvClient.RootPath, `internal`, `app`, AppName, logDatabaseDirName)
	p_common.TBaseClient = &p_common.TBase{
		StartMillUnix: gstool.TimeNowMilliInt64(),
		LogPath:       component.EnvClient.LogPath,
	}
	//初始化shell
	component.ShellClient = p_shell.NewShell(component.EnvClient.LogPath)
	component.ShellOutClient = common.NewTShellOut(component.EnvClient.LogPath)
	//aesGcm
	gcm := gsencrypt.NewAesGcm(component.EnvClient.AppName)
	p_common.AesGcmClient = gcm
	component.GsLog = gstool.NewSlog3(component.EnvClient.LogPath, component.EnvClient.AppName)
	_ = component.GsLog.CleanOldLogs(2)
}

func newConfigViper() *viper.Viper {
	codecRegistry := viper.NewCodecRegistry()
	if err := codecRegistry.RegisterCodec("ini", inicodec.Codec{}); err != nil {
		panic(err)
	}

	return viper.NewWithOptions(viper.WithCodecRegistry(codecRegistry))
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
		DbFileName:                   viper.GetString(`base.dbFileName`),
		DbPath:                       viper.GetString(`base.dbPath`),
		DbIsGitRepo:                  viper.GetBool(`base.dbIsGitRepo`),
		DbAutoPushDelayMinutes:       common.DefaultMainDBAutoPushDelayMinutes,
		LogDbPath:                    viper.GetString(`base.logDbPath`),
		MemoryDBPath:                 viper.GetString(`base.memoryDbPath`),
		MemoryDBIsGitRepo:            viper.GetBool(`base.memoryDbIsGitRepo`),
		MemoryDBAutoPushDelayMinutes: common.DefaultMemoryAutoPushDelayMinutes,
		WebPath:                      viper.GetString(`base.webPath`),
	}
	if viper.IsSet(`base.dbAutoPushDelayMinutes`) {
		component.EnvClient.ConfigBase.DbAutoPushDelayMinutes = viper.GetInt(`base.dbAutoPushDelayMinutes`)
	}
	if viper.IsSet(`base.memoryDbAutoPushDelayMinutes`) {
		component.EnvClient.ConfigBase.MemoryDBAutoPushDelayMinutes = viper.GetInt(`base.memoryDbAutoPushDelayMinutes`)
	}
	//web
	component.EnvClient.WebConfig = &define.WebConfig{
		WebPath: ``,
	}
	// 前端目录：未配置webPath时，默认使用当前项目根目录下的web/dist
	if component.EnvClient.ConfigBase.WebPath == `` {
		component.EnvClient.WebConfig.WebPath = filepath.Join(component.EnvClient.RootPath, `web`, `dist`)
	} else {
		component.EnvClient.WebConfig.WebPath = component.EnvClient.ConfigBase.WebPath
	}
	//数据库配置
	component.EnvClient.DbConfig = &define.DbConfig{
		DbName:      ``,
		DbPath:      common.ResolveDefaultDToolDir(component.EnvClient.ConfigBase.DbPath),
		DbIsGitRepo: component.EnvClient.ConfigBase.DbIsGitRepo,
	}
	//数据库名
	component.EnvClient.DbConfig.DbName = component.EnvClient.AppName + `.db`
	if component.EnvClient.ConfigBase.DbFileName != `` {
		component.EnvClient.DbConfig.DbName = component.EnvClient.ConfigBase.DbFileName
	}
	// log 库默认与主库放在同一目录，但可以通过 logDbPath 单独配置。
	logDbPath := component.EnvClient.ConfigBase.LogDbPath
	if logDbPath == `` {
		logDbPath = component.EnvClient.DbConfig.DbPath
	}
	component.EnvClient.LogDbConfig = &define.DbConfig{
		DbName: buildLogDBName(component.EnvClient.DbConfig.DbName),
		DbPath: logDbPath,
	}
	// Playwright 路径统一默认到 ~/.dtool/server
	component.EnvClient.WebkitDriverPath,
		component.EnvClient.WebkitDataPath,
		component.EnvClient.WebkitDownloadPath = common.ResolvePlaywrightPaths(`server`)
	// 自定义网页配置
	component.EnvClient.SmartLinkConfig = &define.SmartLinkConfig{
		RunMode:       define.SmartLinkRunMode(viper.GetString(`smart_link.run_mode`)),
		ClientVersion: viper.GetString(`smart_link.client_version`),
	}
	// 默认值为 server
	if component.EnvClient.SmartLinkConfig.RunMode == `` {
		component.EnvClient.SmartLinkConfig.RunMode = define.SmartLinkRunModeServer
	}
	// 默认客户端版本
	if component.EnvClient.SmartLinkConfig.ClientVersion == `` {
		component.EnvClient.SmartLinkConfig.ClientVersion = `1.0.0`
	}
	//创建目录
	_ = gstool.DirCreatePath(component.EnvClient.LogPath)
	_ = gstool.DirCreatePath(component.EnvClient.DbConfig.DbPath)
	_ = gstool.DirCreatePath(component.EnvClient.WebkitDataPath)
	_ = gstool.DirCreatePath(component.EnvClient.WebkitDriverPath)
	_ = gstool.DirCreatePath(component.EnvClient.WebkitDownloadPath)
	gstool.FmtPrintlnLogTime(`输出配置：`)
	gstool.FmtPrintlnLogTime(`%s`, formatEnvSummary(component.EnvClient))
}

func initPlaywright() {
	//初始化playwright
	component.PlaywrightClient = component.NewTPlaywright()
	plw.InitPageActiveTime()
	if !component.PlaywrightClient.EnsureNodeRuntime() {
		gstool.FmtPrintlnLogTime(`未检测到 Node.js，跳过 Playwright 初始化，等待用户安装后再使用自定义网页`)
		return
	}
	go component.PlaywrightClient.WitchDownload()
	go component.PlaywrightClient.SmartCheckAndUpdate(&p_sse.SseShell{})
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
	// 新代码统一从 component 取主库实例；common 保留同步赋值，兼容现有调用。
	component.DbMain = &common.CSqlite{Client: component.SqliteClient, Env: component.EnvClient}
	common.DbMain = component.DbMain
	component.DataBaseUp = business.NewTDataBaseUp(component.DbMain, component.EnvClient.DatabaseUpPath)
	component.DataBaseUp.Run()
	initLogSqlite()
	if err = business.LoadMemoryStore(); err != nil {
		panic(err.Error())
	}
	business.StartMainDBAutoSync()
	component.ShellOutClient.InitGroupConfigs()
}

// initLogSqlite 初始化独立 log 库，并执行 log 库迁移。
func initLogSqlite() {
	fmt.Println(fmt.Sprintf(`log库目录 %s`, component.EnvClient.LogDbConfig.DbPath))
	fmt.Println(fmt.Sprintf(`log库路径 %s`, filepath.Join(component.EnvClient.LogDbConfig.DbPath, component.EnvClient.LogDbConfig.DbName)))

	var err error
	component.LogSqliteClient, err = p_db.InitSqlite(component.EnvClient.LogDbConfig.DbPath, component.EnvClient.LogDbConfig.DbName)
	if err != nil {
		panic(fmt.Sprintf(`连接log sqlite失败 %s`, err.Error()))
	}

	// log 库入口也收敛到 component；common.DbLog 继续保留给旧逻辑平滑过渡。
	component.DbLog = &common.CSqlite{Client: component.LogSqliteClient, Env: component.EnvClient}
	common.DbLog = component.DbLog
	business.NewLogDataBaseUp(component.DbLog, component.EnvClient.LogDatabaseUpPath).Run()
}

// buildLogDBName 基于主库文件名派生 log 库文件名。
func buildLogDBName(mainDBName string) string {
	if strings.HasSuffix(mainDBName, databaseFileExt) {
		return strings.TrimSuffix(mainDBName, databaseFileExt) + logDatabaseNameSuffix + databaseFileExt
	}
	return mainDBName + logDatabaseNameSuffix + databaseFileExt
}

// formatLogDBFullPath 返回 log 库完整路径，便于统一输出配置摘要。
func formatLogDBFullPath(env *define.Env) string {
	if env == nil || env.LogDbConfig == nil {
		return ""
	}
	if env.LogDbConfig.DbName == "" || env.LogDbConfig.DbPath == "" {
		return ""
	}
	return filepath.Join(env.LogDbConfig.DbPath, env.LogDbConfig.DbName)
}

func initGin() {
	host := component.ConfigViper.GetString(`run.host`)
	ports := strings.Split(component.ConfigViper.GetString(`run.ports`), `,`)
	component.EnvClient.Ports = ports
	gin.DefaultWriter = io.Discard
	if err := controller.CleanupPortsByPreference(ports, []string{AppName}); err != nil {
		gstool.FmtPrintlnLogTime(`启动前端口清理失败 %s`, err.Error())
	}
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
				cfg := gstool.JsonEncode(map[string]string{
					"port": port,
					"host": host,
				})
				context.HTML(200, `index.html`, gin.H{"serverConfig": template.JS(string(cfg))})
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
	// VariableClient 放到 component 持有，外层不再直接依赖 variable 包级全局变量。
	component.VariableClient = variable.NewVariableClient()
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
	// 启动Shell连接状态广播器，每5秒推送一次
	controller.ShellConnectionsBroadcasterInstance = controller.NewShellConnectionsBroadcaster(5 * time.Second)
}

func Stop() {
	fmt.Println(`停止`)
	// 停止Shell连接状态广播器
	if controller.ShellConnectionsBroadcasterInstance != nil {
		controller.ShellConnectionsBroadcasterInstance.Stop()
	}
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
	if component.MemoryRuntime.HasPendingTask() {
		component.MemoryRuntime.Stop()
		if err := component.MemoryRuntime.SyncPendingTaskNow(); err != nil && !errors.Is(err, common.ErrMemoryNotConfigured) {
			gstool.FmtPrintlnLogTime(`记忆库关闭前同步待处理任务失败 %s`, err.Error())
		}
	} else if err := component.MemoryRuntime.SyncNow(); err != nil && !errors.Is(err, common.ErrMemoryNotConfigured) {
		gstool.FmtPrintlnLogTime(`记忆库关闭前同步失败 %s`, err.Error())
	}
	business.StopMainDBAutoSync()
	if err := business.SyncMainDBStoreOnShutdown(); err != nil {
		gstool.FmtPrintlnLogTime(`主库关闭前同步失败 %s`, err.Error())
	}
	_ = component.PlaywrightClient.Log.Close()
	if component.VariableClient != nil && component.VariableClient.GetLog() != nil {
		_ = component.VariableClient.GetLog().Close()
	}
	_ = component.GsLog.Close()
}
