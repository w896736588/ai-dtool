package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"path/filepath"
	"strings"
)

const (
	runtimeConfigDatabaseFileExt   = `.db`
	runtimeConfigLogDatabaseSuffix = `.log`
)

// ReloadEditableRuntimeConfig 重新把当前 viper 中的可编辑配置同步到运行时环境。 // Reload editable config values from viper into the runtime environment.
func ReloadEditableRuntimeConfig() {
	if component.ConfigViper == nil || component.EnvClient == nil {
		return
	}
	if component.EnvClient.ConfigBase == nil {
		component.EnvClient.ConfigBase = &define.Base{}
	}

	component.EnvClient.ConfigBase.DbFileName = component.ConfigViper.GetString(`base.dbFileName`)
	component.EnvClient.ConfigBase.DbPath = component.ConfigViper.GetString(`base.dbPath`)
	component.EnvClient.ConfigBase.DbIsGitRepo = component.ConfigViper.GetBool(`base.dbIsGitRepo`)
	component.EnvClient.ConfigBase.DbAutoPushDelayMinutes = common.DefaultMainDBAutoPushDelayMinutes
	if component.ConfigViper.IsSet(`base.dbAutoPushDelayMinutes`) {
		component.EnvClient.ConfigBase.DbAutoPushDelayMinutes = component.ConfigViper.GetInt(`base.dbAutoPushDelayMinutes`)
	}
	component.EnvClient.ConfigBase.LogDbPath = component.ConfigViper.GetString(`base.logDbPath`)
	component.EnvClient.ConfigBase.MemoryDBPath = component.ConfigViper.GetString(`base.memoryDbPath`)
	component.EnvClient.ConfigBase.MemoryDBIsGitRepo = component.ConfigViper.GetBool(`base.memoryDbIsGitRepo`)
	component.EnvClient.ConfigBase.MemoryDBAutoPushDelayMinutes = common.DefaultMemoryAutoPushDelayMinutes
	if component.ConfigViper.IsSet(`base.memoryDbAutoPushDelayMinutes`) {
		component.EnvClient.ConfigBase.MemoryDBAutoPushDelayMinutes = component.ConfigViper.GetInt(`base.memoryDbAutoPushDelayMinutes`)
	}
	component.EnvClient.ConfigBase.WebPath = component.ConfigViper.GetString(`base.webPath`)

	if component.EnvClient.WebConfig == nil {
		component.EnvClient.WebConfig = &define.WebConfig{}
	}
	if component.EnvClient.ConfigBase.WebPath == `` {
		component.EnvClient.WebConfig.WebPath = filepath.Join(component.EnvClient.RootPath, `web`, `dist`)
	} else {
		component.EnvClient.WebConfig.WebPath = component.EnvClient.ConfigBase.WebPath
	}

	if component.EnvClient.DbConfig == nil {
		component.EnvClient.DbConfig = &define.DbConfig{}
	}
	component.EnvClient.DbConfig.DbPath = common.ResolveDefaultDToolDir(component.EnvClient.ConfigBase.DbPath)
	component.EnvClient.DbConfig.DbIsGitRepo = component.EnvClient.ConfigBase.DbIsGitRepo
	component.EnvClient.DbConfig.DbName = component.EnvClient.AppName + `.db`
	if component.EnvClient.ConfigBase.DbFileName != `` {
		component.EnvClient.DbConfig.DbName = component.EnvClient.ConfigBase.DbFileName
	}

	if component.EnvClient.LogDbConfig == nil {
		component.EnvClient.LogDbConfig = &define.DbConfig{}
	}
	component.EnvClient.LogDbConfig.DbName = buildRuntimeLogDBName(component.EnvClient.DbConfig.DbName)
	// 日志库路径：优先使用独立的 logDbPath 配置，否则沿用主库路径。
	if component.EnvClient.ConfigBase.LogDbPath != `` {
		component.EnvClient.LogDbConfig.DbPath = common.ResolveDefaultDToolDir(component.EnvClient.ConfigBase.LogDbPath)
	} else {
		component.EnvClient.LogDbConfig.DbPath = component.EnvClient.DbConfig.DbPath
	}

	// Playwright 路径统一默认到 ~/.dtool/server，不再从配置文件读取
	component.EnvClient.WebkitDriverPath,
		component.EnvClient.WebkitDataPath,
		component.EnvClient.WebkitDownloadPath = common.ResolvePlaywrightPaths(`server`)
}

// buildRuntimeLogDBName 基于主库文件名生成 log 库文件名。 // Build log db file name from the main database file name.
func buildRuntimeLogDBName(mainDBName string) string {
	if strings.HasSuffix(mainDBName, runtimeConfigDatabaseFileExt) {
		return strings.TrimSuffix(mainDBName, runtimeConfigDatabaseFileExt) + runtimeConfigLogDatabaseSuffix + runtimeConfigDatabaseFileExt
	}
	return mainDBName + runtimeConfigLogDatabaseSuffix + runtimeConfigDatabaseFileExt
}
