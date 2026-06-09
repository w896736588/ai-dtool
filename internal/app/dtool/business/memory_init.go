package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/memory"
	"fmt"
	"strings"

	"github.com/w896736588/go-tool/gstool"
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
	ListChangedFiles(dir, fileName string) ([]string, error)
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
	if err = memoryGit.Commit(config.Dir, target, `chore: sync memory db`); err != nil {
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
	config := common.MemoryConfig{
		Dir: common.ResolveDefaultDToolDir(memoryDir),
	}
	return config
}

// PrepareMemoryStore 在任何数据库初始化前完成记忆库目录检查和 git 仓库识别 / preflight memory store before any database initialization.
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
	isGitRepo, err := memoryGit.IsGitRepo(config.Dir)
	if err != nil {
		return fmt.Errorf(`检测记忆目录 git 仓库失败 %w`, err)
	}
	config.IsGitRepo = isGitRepo
	config.DBPath = config.Dir
	// 启动时若知识片段目录本身就是 git 仓库，则先拉取最新内容，避免基于旧文件索引启动。
	// Pull latest remote state before boot when the fragment directory itself is a git repository.
	if config.IsGitRepo {
		gstool.FmtPrintlnLogTime(`memory directory is a git repo, pulling latest state before boot dir=%s`, config.Dir)
		if err = memoryGit.Pull(config.Dir); err != nil {
			return fmt.Errorf(`拉取记忆目录失败 %w`, err)
		}
	}
	preparedMemoryStore = &preparedMemoryBootstrap{
		Config:    config,
		MemoryGit: memoryGit,
	}
	gstool.FmtPrintlnLogTime(`记忆库 git 检测完成 dir=%s is_git_repo=%v`, config.Dir, config.IsGitRepo)
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
	if repairReport, err := memory.RepairInvalidFragmentIDs(config.Dir); err != nil {
		return fmt.Errorf(`修复非法知识片段ID失败 %w`, err)
	} else if len(repairReport.Renamed) > 0 {
		gstool.FmtPrintlnLogTime(`修复非法知识片段ID完成 count=%d`, len(repairReport.Renamed))
	}
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
	component.MemoryRuntime.Configure(config, memoryDB)
	memoryDB.LoadAsync()
	if err := memoryDB.StartWatching(); err != nil {
		return fmt.Errorf(`启动记忆目录监听失败 %w`, err)
	}
	return nil
}
