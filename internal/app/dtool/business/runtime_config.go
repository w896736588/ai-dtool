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

// ReloadEditableRuntimeConfig 重新把当前 viper 中的可编辑配置同步到运行时环境。
func ReloadEditableRuntimeConfig() {
	if component.ConfigViper == nil || component.EnvClient == nil {
		return
	}
	if component.EnvClient.ConfigBase == nil {
		component.EnvClient.ConfigBase = &define.Base{}
	}

	component.EnvClient.ConfigBase.DbFileName = component.ConfigViper.GetString(`base.dbFileName`)
	component.EnvClient.ConfigBase.DbPath = component.ConfigViper.GetString(`base.dbPath`)
	component.EnvClient.ConfigBase.LogDbPath = component.ConfigViper.GetString(`base.logDbPath`)
	component.EnvClient.ConfigBase.MemoryDBPath = component.ConfigViper.GetString(`base.memoryDbPath`)
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
	component.EnvClient.DbConfig.DbName = component.EnvClient.AppName + `.db`
	if component.EnvClient.ConfigBase.DbFileName != `` {
		component.EnvClient.DbConfig.DbName = component.EnvClient.ConfigBase.DbFileName
	}

	if component.EnvClient.LogDbConfig == nil {
		component.EnvClient.LogDbConfig = &define.DbConfig{}
	}
	component.EnvClient.LogDbConfig.DbName = buildRuntimeLogDBName(component.EnvClient.DbConfig.DbName)
	if component.EnvClient.ConfigBase.LogDbPath != `` {
		component.EnvClient.LogDbConfig.DbPath = common.ResolveDefaultDToolDir(component.EnvClient.ConfigBase.LogDbPath)
	} else {
		component.EnvClient.LogDbConfig.DbPath = component.EnvClient.DbConfig.DbPath
	}

	component.EnvClient.WebkitDriverPath,
		component.EnvClient.WebkitDataPath,
		component.EnvClient.WebkitDownloadPath = common.ResolvePlaywrightPaths(`server`)
}

func buildRuntimeLogDBName(mainDBName string) string {
	if strings.HasSuffix(mainDBName, runtimeConfigDatabaseFileExt) {
		return strings.TrimSuffix(mainDBName, runtimeConfigDatabaseFileExt) + runtimeConfigLogDatabaseSuffix + runtimeConfigDatabaseFileExt
	}
	return mainDBName + runtimeConfigLogDatabaseSuffix + runtimeConfigDatabaseFileExt
}
