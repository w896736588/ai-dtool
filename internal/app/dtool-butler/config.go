package dtool_butler

import (
	"dev_tool/internal/app/dtool-butler/define"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/pkg/p_db"
	"fmt"
	"os"
	"path/filepath"

	inicodec "github.com/go-viper/encoding/ini"
	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/w896736588/go-tool/gstool"
)

// InitEnv 读取 dtool config.ini，解析数据库与记忆库路径，填充管家 Env。
// 管家与 dtool 共用同一份配置文件，确保连接同一个 SQLite。
func InitEnv(configFile string) *define.Env {
	env := &define.Env{}
	if configFile == `` {
		configFile = `config`
	}
	env.ConfigFile = configFile

	wd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf(`获取当前目录失败 %s`, err.Error()))
	}
	env.RootPath, err = gstool.GetRootPath(wd)
	if err != nil {
		panic(fmt.Sprintf(`获取项目根目录失败 %s`, err.Error()))
	}
	// 管家与 dtool 共用 config/dtool/config.ini
	env.ConfigPath = filepath.Join(env.RootPath, `config`, `dtool`)

	v := newConfigViper()
	v.AddConfigPath(env.ConfigPath)
	v.SetConfigName(env.ConfigFile)
	v.SetConfigType(`ini`)
	if readErr := v.ReadInConfig(); readErr != nil {
		panic(fmt.Sprintf(`读取配置失败 %s`, readErr.Error()))
	}

	// 数据库路径解析：优先用配置，否则回落 ~/.dtool
	dbPath := common.ResolveDefaultDToolDir(v.GetString(`base.dbPath`))
	dbFileName := v.GetString(`base.dbFileName`)
	if dbFileName == `` {
		dbFileName = `dtool.db`
	}
	env.DbPath = dbPath
	env.DbName = dbFileName

	// log 库路径：未配置时与主库同目录
	logDbPath := v.GetString(`base.logDbPath`)
	if logDbPath == `` {
		logDbPath = dbPath
	}
	env.LogDbPath = logDbPath

	// 记忆库路径：索引文档存放于此目录下
	env.MemoryDbPath = common.ResolveDefaultDToolDir(v.GetString(`base.memoryDbPath`))

	// 管家 migration 目录
	env.DatabaseUpDir = filepath.Join(env.RootPath, `internal`, `app`, define.AppName, `database`)

	// 日志目录
	env.LogPath = filepath.Join(env.RootPath, `logs`)
	_ = gstool.DirCreatePath(env.LogPath)

	gstool.FmtPrintlnLogTime(`[butler] 配置摘要`)
	gstool.FmtPrintlnLogTime(`[butler] 根目录: %s`, env.RootPath)
	gstool.FmtPrintlnLogTime(`[butler] 配置: %s`, filepath.Join(env.ConfigPath, env.ConfigFile+`.ini`))
	gstool.FmtPrintlnLogTime(`[butler] 主库: %s`, filepath.Join(env.DbPath, env.DbName))
	gstool.FmtPrintlnLogTime(`[butler] 记忆库: %s`, env.MemoryDbPath)
	gstool.FmtPrintlnLogTime(`[butler] migration: %s`, env.DatabaseUpDir)
	return env
}

// InitSqlite 连接与 dtool 相同的 SQLite 主库，返回封装后的 CSqlite。
func InitSqlite(env *define.Env) *common.CSqlite {
	dbClient, err := p_db.InitSqlite(env.DbPath, env.DbName)
	if err != nil {
		panic(fmt.Sprintf(`[butler] 连接 sqlite 失败 %s`, err.Error()))
	}
	return &common.CSqlite{Client: dbClient}
}

// RunMigration 执行管家自己的 migration SQL，记录表为 tbl_butler_database_up，与 dtool 的迁移记录隔离。
func RunMigration(db *common.CSqlite, env *define.Env) {
	tableName := `tbl_butler_database_up`
	// 确保记录表存在
	name, _ := db.Client.QuickQuery(`sqlite_master`, `name`, map[string]any{
		`name`: tableName,
	}).Value(`name`)
	if cast.ToString(name) == `` {
		_, err := db.Client.ExecBySql(fmt.Sprintf(`
			CREATE TABLE "%s" (
			  "id" INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
			  "filename" TEXT NOT NULL DEFAULT ''
			);
		`, tableName)).Exec()
		if err != nil {
			panic(fmt.Sprintf(`[butler] 迁移记录表创建失败 %s`, err.Error()))
		}
	}
	// 已执行的文件名
	rows, err := db.Client.QuickQuery(tableName, `filename`, nil).All()
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler] 迁移记录查询失败 %s`, err.Error())
		return
	}
	upFileNames := make([]string, 0, len(rows))
	for _, row := range rows {
		upFileNames = append(upFileNames, cast.ToString(row[`filename`]))
	}
	// 扫描 migration 目录，按年月组织
	files := make([]string, 0)
	for year := 2026; year <= cast.ToInt(gstool.TimeNowUnixToString(`Y`)); year++ {
		for month := 1; month <= 12; month++ {
			monthStr := fmt.Sprintf(`%02d`, month)
			dir := filepath.Join(env.DatabaseUpDir, cast.ToString(year), monthStr)
			exist, _ := gstool.DirPathExists(dir)
			if !exist {
				continue
			}
			_ = gstool.DirWalk(dir, func(path string, info os.FileInfo, err error) {
				if info.IsDir() {
					return
				}
				if !gstool.ArrayExistValue(&upFileNames, info.Name()) {
					files = append(files, filepath.Join(dir, info.Name()))
				}
			})
		}
	}
	for _, file := range files {
		gstool.FmtPrintlnLogTime(`[butler] 执行迁移文件 %s`, file)
		sql, err := gstool.FileGetContent(file)
		if err != nil {
			gstool.FmtPrintlnLogTime(`[butler] 读取迁移文件失败 %s %s`, file, err.Error())
			continue
		}
		_, err = db.Client.ExecBySql(sql).Exec()
		if err != nil {
			gstool.FmtPrintlnLogTime(`[butler] 迁移文件执行失败 %s err=%s`, file, err.Error())
			continue
		}
		_, err = db.Client.QuickCreate(tableName, map[string]any{
			`filename`: gstool.FileGetNameByPath(file),
		}).Exec()
		if err != nil {
			gstool.FmtPrintlnLogTime(`[butler] 迁移记录插入失败 %s`, err.Error())
		}
	}
}

func newConfigViper() *viper.Viper {
	codecRegistry := viper.NewCodecRegistry()
	if err := codecRegistry.RegisterCodec("ini", inicodec.Codec{}); err != nil {
		panic(err)
	}
	return viper.NewWithOptions(viper.WithCodecRegistry(codecRegistry))
}
