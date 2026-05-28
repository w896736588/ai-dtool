package business

import (
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
	ListChangedFiles(dir, fileName string) ([]string, error)
}

// MainDBConfig 描述主库 sqlite 与 git 同步配置。 // Describes the main sqlite database and its git sync settings.
type MainDBConfig struct {
	Dir       string
	DBName    string
	DBPath    string
	IsGitRepo bool
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
	if config.Dir != `` && config.DBName != `` {
		config.DBPath = filepath.Join(config.Dir, config.DBName)
	}
	if preparedMainDBStore != nil {
		config.IsGitRepo = preparedMainDBStore.Config.IsGitRepo
	}
	return config
}

// PrepareMainDBStore 在主库初始化前检测目录是否为 git 仓库。 // Detect whether the main db directory is a git repository before boot.
func PrepareMainDBStore() error {
	config := ReadMainDBConfig()
	if config.Dir == `` || config.DBName == `` {
		preparedMainDBStore = nil
		gstool.FmtPrintlnLogTime(`主库未配置完整，跳过 git 检测 dir=%s db=%s`, config.Dir, config.DBName)
		return nil
	}
	if err := gstool.DirCreatePath(config.Dir); err != nil {
		return fmt.Errorf(`创建主库目录失败 %w`, err)
	}

	gitSyncer := newMainDBGitSyncer()
	isGitRepo, err := gitSyncer.IsGitRepo(config.Dir)
	if err != nil {
		return fmt.Errorf(`检测主库目录 git 仓库失败 %w`, err)
	}
	config.IsGitRepo = isGitRepo
	// 启动时若主库目录就是 git 仓库，则先拉取最新内容，避免服务基于过期 sqlite 启动。
	// Pull latest remote state before boot when the main-db directory itself is a git repository.
	if config.IsGitRepo {
		gstool.FmtPrintlnLogTime(`main db directory is a git repo, pulling latest state before boot dir=%s`, config.Dir)
		if err = gitSyncer.Pull(config.Dir); err != nil {
			return fmt.Errorf(`拉取主库目录失败 %w`, err)
		}
	}

	preparedMainDBStore = &preparedMainDBBootstrap{
		Config: config,
		Git:    gitSyncer,
	}
	gstool.FmtPrintlnLogTime(`主库 git 检测完成 dir=%s is_git_repo=%v`, config.Dir, config.IsGitRepo)
	return nil
}
