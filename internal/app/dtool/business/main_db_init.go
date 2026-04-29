package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"fmt"
	"path/filepath"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

const mainDBSyncCommitMessage = `chore: sync main db`

// mainDBGitSyncer 定义主库 git 同步所需能力。 // Defines the git sync capabilities required by the main database.
type mainDBGitSyncer interface {
	IsGitRepo(dir string) (bool, error)
	Pull(dir string) error
	HasFileChanges(dir, fileName string) (bool, error)
	AddFile(dir, fileName string) error
	Commit(dir, fileName, message string) error
	Push(dir string) error
}

// MainDBConfig 描述主库 sqlite 与 git 同步配置。 // Describes the main sqlite database and its git sync settings.
type MainDBConfig struct {
	Dir            string
	DBName         string
	DBPath         string
	IsGitRepo      bool
	GitRepoEnabled bool
}

type preparedMainDBBootstrap struct {
	Config MainDBConfig
	Git    mainDBGitSyncer
}

var preparedMainDBStore *preparedMainDBBootstrap

// newMainDBGitSyncer 允许测试替换 git 同步实现。 // Allows tests to replace the git sync implementation.
var newMainDBGitSyncer = func() mainDBGitSyncer {
	return NewMemoryGit()
}

// SyncMainDBFile 手动同步主库 sqlite 文件到 git 仓库。 // Sync the main sqlite file into the git repository on demand.
func SyncMainDBFile(config MainDBConfig, gitSyncer mainDBGitSyncer) (bool, error) {
	if config.Dir == `` || config.DBPath == `` {
		return false, fmt.Errorf(`主库配置不完整，无法执行同步`)
	}
	if !config.IsGitRepo {
		return false, fmt.Errorf(`主库未启用 git 仓库同步`)
	}
	if gitSyncer == nil {
		return false, fmt.Errorf(`主库 git syncer 未设置`)
	}

	fileName := filepath.Base(config.DBPath)
	gstool.FmtPrintlnLogTime(`主库开始检查变更并执行手动同步 dir=%s file=%s`, config.Dir, fileName)
	hasChanges, err := gitSyncer.HasFileChanges(config.Dir, fileName)
	if err != nil {
		gstool.FmtPrintlnLogTime(`主库手动同步前检查变更失败 dir=%s file=%s err=%s`, config.Dir, fileName, err.Error())
		return false, err
	}
	// 没有文件变更时直接返回未同步，前端可展示“无需同步”。 // Return a no-op result when the database file has no pending changes.
	if !hasChanges {
		gstool.FmtPrintlnLogTime(`主库未检测到文件变更，跳过手动同步 dir=%s file=%s`, config.Dir, fileName)
		return false, nil
	}
	if err = gitSyncer.AddFile(config.Dir, fileName); err != nil {
		gstool.FmtPrintlnLogTime(`主库手动同步 add 失败 dir=%s file=%s err=%s`, config.Dir, fileName, err.Error())
		return false, err
	}
	if err = gitSyncer.Commit(config.Dir, fileName, mainDBSyncCommitMessage); err != nil {
		gstool.FmtPrintlnLogTime(`主库手动同步 commit 失败 dir=%s file=%s err=%s`, config.Dir, fileName, err.Error())
		return false, err
	}
	if err = gitSyncer.Push(config.Dir); err != nil {
		gstool.FmtPrintlnLogTime(`主库手动同步 push 失败 dir=%s file=%s err=%s`, config.Dir, fileName, err.Error())
		return false, err
	}
	gstool.FmtPrintlnLogTime(`主库手动同步成功 dir=%s file=%s`, config.Dir, fileName)
	return true, nil
}

// ReadMainDBConfig 读取已经解析过默认值的主库配置。 // Reads the resolved main database config after defaults are applied.
func ReadMainDBConfig() MainDBConfig {
	if component.EnvClient == nil || component.EnvClient.DbConfig == nil {
		return MainDBConfig{}
	}
	config := MainDBConfig{
		Dir:    strings.TrimSpace(component.EnvClient.DbConfig.DbPath),
		DBName: strings.TrimSpace(component.EnvClient.DbConfig.DbName),
	}
	if component.EnvClient.ConfigBase != nil {
		config.GitRepoEnabled = component.EnvClient.ConfigBase.DbIsGitRepo
	}
	if config.Dir != `` && config.DBName != `` {
		config.DBPath = filepath.Join(config.Dir, config.DBName)
	}
	return config
}

// ReadMainDBAutoSyncConfig 读取主库自动同步配置。 // Read main db auto-sync config from ini and runtime state.
func ReadMainDBAutoSyncConfig() common.MainDBAutoSyncConfig {
	mainConfig := ReadMainDBConfig()
	autoSyncMinutes := common.DefaultMainDBAutoPushDelayMinutes
	if component.EnvClient != nil && component.EnvClient.ConfigBase != nil {
		if component.EnvClient.ConfigBase.DbAutoPushDelayMinutes > 0 {
			autoSyncMinutes = component.EnvClient.ConfigBase.DbAutoPushDelayMinutes
		}
	}
	// IsGitRepo 只在 PrepareMainDBStore 中通过 gitSyncer.IsGitRepo() 实际验证后才为 true，
	// ReadMainDBConfig() 不设置该字段（始终为零值），因此需要从 preparedMainDBStore 取已验证的值。
	isGitRepo := mainConfig.IsGitRepo
	if preparedMainDBStore != nil {
		isGitRepo = preparedMainDBStore.Config.IsGitRepo
	}
	return common.MainDBAutoSyncConfig{
		Dir:             mainConfig.Dir,
		DBName:          mainConfig.DBName,
		IsGitRepo:       isGitRepo,
		GitRepoEnabled:  mainConfig.GitRepoEnabled,
		AutoSyncMinutes: autoSyncMinutes,
	}
}

// PrepareMainDBStore 在主库初始化前按配置执行 git pull。 // Performs git pull before opening the main database when enabled.
func PrepareMainDBStore() error {
	config := ReadMainDBConfig()
	if config.Dir == `` || config.DBName == `` {
		preparedMainDBStore = nil
		gstool.FmtPrintlnLogTime(`主库未配置完整，跳过 git 预处理 dir=%s db=%s`, config.Dir, config.DBName)
		return nil
	}
	if err := gstool.DirCreatePath(config.Dir); err != nil {
		return fmt.Errorf(`创建主库目录失败 %w`, err)
	}

	gitSyncer := newMainDBGitSyncer()
	// 只有显式开启 git 模式时才执行仓库检测与拉取。 // Only detect and pull when git mode is explicitly enabled.
	if config.GitRepoEnabled {
		gstool.FmtPrintlnLogTime(`主库 git 模式已开启，准备检查仓库并执行 pull dir=%s file=%s`, config.Dir, config.DBName)
		isGitRepo, err := gitSyncer.IsGitRepo(config.Dir)
		if err != nil {
			return fmt.Errorf(`检测主库目录 git 仓库失败 %w`, err)
		}
		// 开启了 git 模式却不是仓库时直接失败，避免静默运行。 // Fail fast when git mode is enabled but the directory is not a repository.
		if !isGitRepo {
			return fmt.Errorf(`主库目录未检测到 git 仓库，请检查 base.dbIsGitRepo 和 dbPath 配置`)
		}
		if err = gitSyncer.Pull(config.Dir); err != nil {
			return fmt.Errorf(`拉取主库目录失败 %w`, err)
		}
		gstool.FmtPrintlnLogTime(`主库 git pull 完成 dir=%s`, config.Dir)
		config.IsGitRepo = true
	} else {
		gstool.FmtPrintlnLogTime(`主库 git 模式未开启，跳过 pull dir=%s file=%s`, config.Dir, config.DBName)
	}

	preparedMainDBStore = &preparedMainDBBootstrap{
		Config: config,
		Git:    gitSyncer,
	}
	if component.EnvClient != nil && component.EnvClient.DbConfig != nil {
		component.EnvClient.DbConfig.DbIsGitRepo = config.GitRepoEnabled
	}
	return nil
}

// SyncMainDBStoreOnShutdown 在程序关闭时检查主库文件并执行自动 push。 // Checks the main db file and performs auto-push during shutdown.
func SyncMainDBStoreOnShutdown() error {
	if preparedMainDBStore == nil {
		gstool.FmtPrintlnLogTime(`主库关闭前未找到预处理上下文，跳过 push`)
		return nil
	}
	config := preparedMainDBStore.Config
	if !config.IsGitRepo {
		gstool.FmtPrintlnLogTime(`主库未启用 git 仓库同步，关闭前跳过 push dir=%s file=%s`, config.Dir, config.DBName)
		return nil
	}
	_, err := SyncMainDBFile(config, preparedMainDBStore.Git)
	return err
}

// StartMainDBAutoSync 启动主库自动同步定时器。 // Start the main db auto-sync periodic timer.
func StartMainDBAutoSync() {
	if preparedMainDBStore == nil {
		gstool.FmtPrintlnLogTime(`主库未完成预处理，跳过自动同步启动`)
		return
	}
	config := ReadMainDBAutoSyncConfig()
	if !config.IsGitRepo {
		gstool.FmtPrintlnLogTime(`主库未启用 git 仓库同步，跳过自动同步启动`)
		return
	}
	component.MainDBAutoSyncRuntime.Configure(config, preparedMainDBStore.Git)
	component.MainDBAutoSyncRuntime.Start()
}

// StopMainDBAutoSync 停止主库自动同步定时器并执行最后一次同步。 // Stop the main db auto-sync timer and perform a final sync.
func StopMainDBAutoSync() {
	if component.MainDBAutoSyncRuntime == nil {
		return
	}
	hasPendingTask := component.MainDBAutoSyncRuntime.HasPendingTask()
	component.MainDBAutoSyncRuntime.Stop()
	if hasPendingTask {
		_ = component.MainDBAutoSyncRuntime.SyncPendingTaskNow()
		return
	}
	_ = component.MainDBAutoSyncRuntime.SyncNow()
}
