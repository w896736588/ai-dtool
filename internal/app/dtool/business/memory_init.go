package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/memory"
	"fmt"
	"strings"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

// preparedMemoryStore 保存启动阶段已完成的记忆库预处理结果 / stores memory preflight result completed during boot.
var preparedMemoryStore *preparedMemoryBootstrap

// preparedMemoryBootstrap 记录记忆库预处理后的配置和 git 状态 / records memory config and git status after preflight.
type preparedMemoryBootstrap struct {
	Config    common.MemoryConfig
	MemoryGit memoryGitSyncer
}

// memoryGitSyncer 定义记忆库 git 同步所需能力。 // Defines the git sync capabilities required by the memory store.
type memoryGitSyncer interface {
	IsGitRepo(dir string) (bool, error)
	Pull(dir string) error
	HasFileChanges(dir, fileName string) (bool, error)
	AddFile(dir, fileName string) error
	Commit(dir, fileName, message string) error
	Push(dir string) error
}

// newMemoryGitFactory 允许测试替换记忆库 git 实现。 // Allows tests to replace the memory git implementation.
var newMemoryGitFactory = func() memoryGitSyncer {
	return NewMemoryGit()
}

// SyncMemoryDBFile 手动同步记忆库 sqlite 文件到 git 仓库。 // Sync the memory sqlite file into the git repository on demand.
func SyncMemoryDBFile(config common.MemoryConfig, memoryGit memoryGitSyncer) (bool, error) {
	if config.Dir == `` {
		return false, fmt.Errorf(`记忆库配置不完整，无法执行同步`)
	}
	if !config.IsGitRepo {
		return false, fmt.Errorf(`记忆库未启用 git 仓库同步`)
	}
	if memoryGit == nil {
		return false, fmt.Errorf(`记忆库 git syncer 未设置`)
	}

	target := `.`
	gstool.FmtPrintlnLogTime(`记忆库开始检查变更并执行手动同步 dir=%s target=%s`, config.Dir, target)
	hasChanges, err := memoryGit.HasFileChanges(config.Dir, target)
	if err != nil {
		gstool.FmtPrintlnLogTime(`记忆库手动同步前检查变更失败 dir=%s target=%s err=%s`, config.Dir, target, err.Error())
		return false, err
	}
	// 没有文件变更时直接返回未同步，供页面提示无需 push。 // Return a no-op result when the database file has no pending changes.
	if !hasChanges {
		gstool.FmtPrintlnLogTime(`记忆库未检测到文件变更，跳过手动同步 dir=%s target=%s`, config.Dir, target)
		return false, nil
	}
	if err = memoryGit.AddFile(config.Dir, target); err != nil {
		gstool.FmtPrintlnLogTime(`记忆库手动同步 add 失败 dir=%s target=%s err=%s`, config.Dir, target, err.Error())
		return false, err
	}
	if err = memoryGit.Commit(config.Dir, target, common.MemorySyncCommitMessage); err != nil {
		gstool.FmtPrintlnLogTime(`记忆库手动同步 commit 失败 dir=%s target=%s err=%s`, config.Dir, target, err.Error())
		return false, err
	}
	if err = memoryGit.Push(config.Dir); err != nil {
		gstool.FmtPrintlnLogTime(`记忆库手动同步 push 失败 dir=%s target=%s err=%s`, config.Dir, target, err.Error())
		return false, err
	}
	gstool.FmtPrintlnLogTime(`记忆库手动同步成功 dir=%s target=%s`, config.Dir, target)
	return true, nil
}

// ReadMemoryConfigFromINI 从 ini 读取记忆库配置 / read memory db config from ini.
func ReadMemoryConfigFromINI() common.MemoryConfig {
	if component.ConfigViper == nil {
		return common.MemoryConfig{}
	}
	memoryDir := strings.TrimSpace(component.ConfigViper.GetString(`base.memoryDbPath`))
	memoryDBIsGitRepo := component.ConfigViper.GetBool(`base.memoryDbIsGitRepo`)
	config := common.MemoryConfig{
		Dir:                  common.ResolveDefaultDToolDir(memoryDir),
		GitRepoEnabled:       memoryDBIsGitRepo,
		AutoPushDelayMinutes: common.DefaultMemoryAutoPushDelayMinutes,
	}
	if component.ConfigViper.IsSet(`base.memoryDbAutoPushDelayMinutes`) {
		config.AutoPushDelayMinutes = component.ConfigViper.GetInt(`base.memoryDbAutoPushDelayMinutes`)
	}
	return config
}

// PrepareMemoryStore 在任何数据库初始化前完成记忆库目录检查和 git pull / preflight memory store before any database initialization.
func PrepareMemoryStore() error {
	config := ReadMemoryConfigFromINI()
	if config.Dir == `` {
		preparedMemoryStore = nil
		gstool.FmtPrintlnLogTime(`记忆库未在配置文件中配置，跳过初始化`)
		return nil
	}
	if err := gstool.DirCreatePath(config.Dir); err != nil {
		return fmt.Errorf(`创建记忆目录失败 %w`, err)
	}

	memoryGit := newMemoryGitFactory()
	// 仅当配置显式开启 git 同步时才检测和拉取 / only detect and pull git when config explicitly enables git sync.
	if config.GitRepoEnabled {
		gstool.FmtPrintlnLogTime(`记忆库 git 模式已开启，准备检查仓库并执行 pull dir=%s`, config.Dir)
		isGitRepo, err := memoryGit.IsGitRepo(config.Dir)
		if err != nil {
			return fmt.Errorf(`检测记忆目录 git 仓库失败 %w`, err)
		}
		// 配置已启用 git，但目录不是仓库时直接报错 / fail fast when git sync is enabled but the directory is not a repository.
		if !isGitRepo {
			return fmt.Errorf(`记忆目录未检测到 git 仓库，请检查 base.memoryDbIsGitRepo 和 memoryDbPath 配置`)
		}
		if err = memoryGit.Pull(config.Dir); err != nil {
			return fmt.Errorf(`拉取记忆目录失败 %w`, err)
		}
		gstool.FmtPrintlnLogTime(`记忆库 git pull 完成 dir=%s`, config.Dir)
		config.IsGitRepo = true
	} else {
		gstool.FmtPrintlnLogTime(`记忆库 git 模式未开启，跳过 pull dir=%s`, config.Dir)
	}
	config.DBPath = config.Dir
	preparedMemoryStore = &preparedMemoryBootstrap{
		Config:    config,
		MemoryGit: memoryGit,
	}
	return nil
}

func LoadMemoryStore() error {
	component.MemoryRuntime.Reset()

	// 若上游尚未预处理，则在这里兜底，兼容旧调用路径 / fallback here when caller did not preflight memory store.
	if preparedMemoryStore == nil {
		if err := PrepareMemoryStore(); err != nil {
			return err
		}
	}
	if preparedMemoryStore == nil {
		return nil
	}

	config := preparedMemoryStore.Config
	memoryGit := preparedMemoryStore.MemoryGit
	// if migrationReport, err := memory.MigrateNumericFragmentIDs(config.Dir); err != nil {
	// 	return fmt.Errorf(`迁移旧数字知识片段ID失败 %w`, err)
	// } else if component.DbMain != nil {
	// 	// 启动时同步修正首页任务里的旧片段引用，避免数字文件改名后关联失效。
	// 	// Repair home-task fragment references during boot so renamed legacy files do not break associations.
	// 	if err = component.DbMain.ReplaceHomeTaskMemoryFragmentIDs(migrationReport.IDMap); err != nil {
	// 		return fmt.Errorf(`同步首页任务知识片段ID失败 %w`, err)
	// 	}
	// }

	memoryDB := memory.NewService(config.Dir)
	component.MemoryRuntime.SetGitSyncer(memoryGit)
	component.MemoryRuntime.Configure(config, memoryDB)
	memoryDB.LoadAsync()
	if err := memoryDB.StartWatching(); err != nil {
		return fmt.Errorf(`启动记忆目录监听失败 %w`, err)
	}
	return nil
}
