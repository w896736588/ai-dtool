package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/memory"
	"dev_tool/internal/pkg/p_db"
	"fmt"
	"path/filepath"
	"strings"
	"sync"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsdb"
	"github.com/w896736588/go-tool/gstool"
)

// hotReloadMu 保护所有热重载入口，同一时刻只允许一个配置项执行热重载。
var hotReloadMu sync.Mutex

// HotReloadLogDB 热重载日志库（仅 logDbPath 变更时使用）。
func HotReloadLogDB() error {
	hotReloadMu.Lock()
	defer hotReloadMu.Unlock()

	// 保存旧引用
	oldLogSqliteClient := component.LogSqliteClient
	oldDbLog := component.DbLog

	// 基于 ReloadEditableRuntimeConfig 已刷新的 viper 配置重建 log 库路径
	ReloadEditableRuntimeConfig()
	logDbPath := component.EnvClient.LogDbConfig.DbPath
	logDbName := component.EnvClient.LogDbConfig.DbName
	if logDbPath == `` || logDbName == `` {
		return fmt.Errorf(`日志库配置不完整，无法热重载`)
	}
	if err := gstool.DirCreatePath(logDbPath); err != nil {
		return fmt.Errorf(`创建日志库目录失败 %w`, err)
	}

	newLogClient, err := p_db.InitSqlite(logDbPath, logDbName)
	if err != nil {
		return fmt.Errorf(`连接新日志库失败 %w`, err)
	}
	newDbLog := &common.CSqlite{Client: newLogClient, Env: component.EnvClient}
	NewLogDataBaseUp(newDbLog, component.EnvClient.LogDatabaseUpPath).Run()

	// 切换全局引用
	component.LogSqliteClient = newLogClient
	component.DbLog = newDbLog
	common.DbLog = newDbLog

	gstool.FmtPrintlnLogTime(`日志库热重载成功 path=%s name=%s`, logDbPath, logDbName)
	_ = oldLogSqliteClient // 旧连接保持存活，由进程生命周期管理
	_ = oldDbLog
	return nil
}

// HotReloadMainDB 热重载主库（db_path 或 dbFileName 变更时使用）。
func HotReloadMainDB(changedKey string) error {
	hotReloadMu.Lock()
	defer hotReloadMu.Unlock()

	// 保存旧引用
	oldSqliteClient := component.SqliteClient
	oldDbMain := component.DbMain
	oldPreparedMainDBStore := preparedMainDBStore

	// 基于 viper 已刷新的配置重新计算主库路径
	ReloadEditableRuntimeConfig()
	newDbPath := component.EnvClient.DbConfig.DbPath
	newDbName := component.EnvClient.DbConfig.DbName
	if newDbPath == `` || newDbName == `` {
		return fmt.Errorf(`主库配置不完整，无法热重载`)
	}

	// 1. PrepareMainDBStore 预处理（目录创建 + git pull）
	if err := PrepareMainDBStore(); err != nil {
		return fmt.Errorf(`主库预处理失败 %w`, err)
	}

	// 2. 创建新主库连接
	newSqliteClient, err := p_db.InitSqlite(newDbPath, newDbName)
	if err != nil {
		restorePreparedMainDBStore(oldPreparedMainDBStore)
		return fmt.Errorf(`连接新主库失败 %w`, err)
	}

	newDbMain := &common.CSqlite{Client: newSqliteClient, Env: component.EnvClient}

	// 3. 执行新主库迁移
	NewTDataBaseUp(newDbMain, component.EnvClient.DatabaseUpPath).Run()
	// SQL 迁移后将旧 step_key 统一迁移到 custom_xx 格式
	newDbMain.WorkflowMigrateLegacyStepKeys()

	// 4. 判断 log 库是否受影响
	logDBAffected := false
	if changedKey == `dbFileName` {
		// dbFileName 变化 -> log 库文件名跟着变化
		logDBAffected = true
	} else if changedKey == `db_path` && component.EnvClient.ConfigBase.LogDbPath == `` {
		// db_path 变化且 logDbPath 未单独配置 -> log 库目录跟着变化
		logDBAffected = true
	}

	var newLogSqliteClient *gsdb.GsSqlite
	var newDbLog *common.CSqlite
	if logDBAffected {
		logDbPath := component.EnvClient.LogDbConfig.DbPath
		logDbName := component.EnvClient.LogDbConfig.DbName
		if logDbPath != `` && logDbName != `` {
			if err := gstool.DirCreatePath(logDbPath); err != nil {
				restorePreparedMainDBStore(oldPreparedMainDBStore)
				return fmt.Errorf(`创建日志库目录失败 %w`, err)
			}
			newLogClient, logErr := p_db.InitSqlite(logDbPath, logDbName)
			if logErr != nil {
				restorePreparedMainDBStore(oldPreparedMainDBStore)
				return fmt.Errorf(`连接新日志库失败 %w`, logErr)
			}
			newDbLog = &common.CSqlite{Client: newLogClient, Env: component.EnvClient}
			NewLogDataBaseUp(newDbLog, component.EnvClient.LogDatabaseUpPath).Run()
			newLogSqliteClient = newLogClient
		}
	}

	// 5. 切换窗口：替换全局引用
	component.SqliteClient = newSqliteClient
	component.DbMain = newDbMain
	common.DbMain = newDbMain
	component.DataBaseUp = NewTDataBaseUp(newDbMain, component.EnvClient.DatabaseUpPath)
	if logDBAffected && newLogSqliteClient != nil {
		component.LogSqliteClient = newLogSqliteClient
		component.DbLog = newDbLog
		common.DbLog = newDbLog
	}

	// 6. 刷新 shell 规则
	if component.ShellOutClient != nil {
		component.ShellOutClient.InitGroupConfigs()
	}

	gstool.FmtPrintlnLogTime(`主库热重载成功 path=%s name=%s logAffected=%v`, newDbPath, newDbName, logDBAffected)
	_ = oldSqliteClient
	_ = oldDbMain
	return nil
}

// HotReloadMemoryDB 热重载记忆库（memoryDbPath 变更时使用）。
func HotReloadMemoryDB() error {
	hotReloadMu.Lock()
	defer hotReloadMu.Unlock()

	// 保存旧引用
	oldDB := component.MemoryRuntime.DB()
	var oldService *memory.Service
	if svc, ok := oldDB.(*memory.Service); ok {
		oldService = svc
	}

	// 停止旧 watcher
	if oldService != nil {
		_ = oldService.StopWatching()
	}

	// 基于 viper 已刷新的配置预处理新记忆库目录
	ReloadEditableRuntimeConfig()
	if err := PrepareMemoryStore(); err != nil {
		// 预处理失败，尝试恢复旧 watcher
		if oldService != nil {
			_ = oldService.StartWatching()
		}
		return fmt.Errorf(`记忆库预处理失败 %w`, err)
	}

	// LoadMemoryStore 会 Reset 旧 runtime、重建 memory service 和 watcher
	if err := LoadMemoryStore(); err != nil {
		// 加载失败，尝试恢复旧 watcher
		if oldService != nil {
			_ = oldService.StartWatching()
		}
		return fmt.Errorf(`记忆库加载失败 %w`, err)
	}

	gstool.FmtPrintlnLogTime(`记忆库热重载成功`)
	return nil
}

// HotReloadCronScheduler 热重载指定类型的定时任务调度器。
func HotReloadCronScheduler(taskType string) error {
	hotReloadMu.Lock()
	defer hotReloadMu.Unlock()

	one, err := common.DbMain.CronTaskByType(taskType)
	if err != nil && !common.DbRowMissing(err) {
		return fmt.Errorf(`读取定时任务配置失败 %w`, err)
	}
	enabled := cast.ToInt(one[`enabled`]) == 1
	triggerTime := strings.TrimSpace(cast.ToString(one[`trigger_time`]))

	startCronSchedulerByType(taskType, enabled, triggerTime)
	gstool.FmtPrintlnLogTime(`定时任务热重载成功 type=%s enabled=%v time=%s`, taskType, enabled, triggerTime)
	return nil
}

// NeedsLogDBReload 判断配置项是否会导致 log 库需要联动切换。
func NeedsLogDBReload(changedKey string) bool {
	if changedKey == `dbFileName` {
		return true
	}
	if changedKey == `db_path` {
		logDbPath := strings.TrimSpace(component.ConfigViper.GetString(`base.logDbPath`))
		return logDbPath == ``
	}
	return false
}

// restorePreparedMainDBStore 恢复旧的主库预处理状态。
func restorePreparedMainDBStore(old *preparedMainDBBootstrap) {
	if old != nil {
		preparedMainDBStore = old
	}
}

// needsMainDBHotReload 判断 key 是否需要主库热重载。
func needsMainDBHotReload(key string) bool {
	return key == `db_path` || key == `dbFileName`
}

// needsMemoryDBHotReload 判断 key 是否需要记忆库热重载。
func needsMemoryDBHotReload(key string) bool {
	return key == `memoryDbPath`
}

// buildCurrentLogDBPath 获取当前生效的 log 库完整路径。
func buildCurrentLogDBPath() string {
	if component.EnvClient == nil || component.EnvClient.LogDbConfig == nil {
		return ``
	}
	return filepath.Join(component.EnvClient.LogDbConfig.DbPath, component.EnvClient.LogDbConfig.DbName)
}
