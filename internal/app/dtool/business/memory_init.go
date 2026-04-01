package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/pkg/p_db"
	"fmt"
	"path/filepath"
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

// ReadMemoryConfigFromINI 从 ini 读取记忆库配置 / read memory db config from ini.
func ReadMemoryConfigFromINI() common.MemoryConfig {
	if component.ConfigViper == nil {
		return common.MemoryConfig{}
	}
	memoryDir := strings.TrimSpace(component.ConfigViper.GetString(`base.memoryDbPath`))
	memoryDBName := strings.TrimSpace(component.ConfigViper.GetString(`base.memoryDbFileName`))
	memoryDBIsGitRepo := component.ConfigViper.GetBool(`base.memoryDbIsGitRepo`)
	if memoryDBName == `` {
		// 未显式配置文件名时使用 memory.db，和注释说明保持一致。 // Use memory.db when no file name is explicitly configured, matching the config comment.
		memoryDBName = `memory.db`
	}
	config := common.MemoryConfig{
		Dir:            common.ResolveDefaultDToolDir(memoryDir),
		DBName:         memoryDBName,
		GitRepoEnabled: memoryDBIsGitRepo,
	}
	config.DBPath = filepath.Join(config.Dir, config.DBName)
	return config
}

// PrepareMemoryStore 在任何数据库初始化前完成记忆库目录检查和 git pull / preflight memory store before any database initialization.
func PrepareMemoryStore() error {
	config := ReadMemoryConfigFromINI()
	if config.Dir == `` || config.DBName == `` {
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
		gstool.FmtPrintlnLogTime(`记忆库 git 模式已开启，准备检查仓库并执行 pull dir=%s file=%s`, config.Dir, config.DBName)
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
		gstool.FmtPrintlnLogTime(`记忆库 git 模式未开启，跳过 pull dir=%s file=%s`, config.Dir, config.DBName)
	}
	config.DBPath = filepath.Join(config.Dir, config.DBName)
	preparedMemoryStore = &preparedMemoryBootstrap{
		Config:    config,
		MemoryGit: memoryGit,
	}
	return nil
}

func LoadMemoryStore() error {
	common.MemoryRuntime.Reset()

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

	memoryClient, err := p_db.InitSqlite(config.Dir, config.DBName)
	if err != nil {
		return fmt.Errorf(`连接记忆 sqlite 失败 %w`, err)
	}
	memoryDB := &common.CSqlite{Client: memoryClient, Env: component.EnvClient}
	NewMemoryDataBaseUp(memoryDB, component.EnvClient.MemoryDatabaseUpPath).Run()
	common.MemoryRuntime.SetGitSyncer(memoryGit)
	common.MemoryRuntime.Configure(config, memoryDB)
	return nil
}
