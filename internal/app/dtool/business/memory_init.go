package business

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_db"
	"fmt"
	"path/filepath"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

func LoadMemoryStore() error {
	common.MemoryRuntime.Reset()

	memoryDir, err := common.DbMain.GlobalValue(define.GlobalMemoryDir)
	if err != nil {
		return fmt.Errorf(`读取记忆目录配置失败 %w`, err)
	}
	memoryDBName, err := common.DbMain.GlobalValue(define.GlobalMemoryDBName)
	if err != nil {
		return fmt.Errorf(`读取记忆数据库名配置失败 %w`, err)
	}
	if memoryDir == `` || memoryDBName == `` {
		gstool.FmtPrintlnLogTime(`记忆库未配置，跳过初始化`)
		return nil
	}
	if err = gstool.DirCreatePath(memoryDir); err != nil {
		return fmt.Errorf(`创建记忆目录失败 %w`, err)
	}

	memoryGit := NewMemoryGit()
	isGitRepo, err := memoryGit.IsGitRepo(memoryDir)
	if err != nil {
		return fmt.Errorf(`检测记忆目录 git 仓库失败 %w`, err)
	}
	if isGitRepo {
		if err = memoryGit.Pull(memoryDir); err != nil {
			return fmt.Errorf(`拉取记忆目录失败 %w`, err)
		}
	}

	memoryClient, err := p_db.InitSqlite(memoryDir, memoryDBName)
	if err != nil {
		return fmt.Errorf(`连接记忆sqlite失败 %w`, err)
	}
	memoryDB := &common.CSqlite{Client: memoryClient, Env: component.EnvClient}
	NewMemoryDataBaseUp(memoryDB, component.EnvClient.MemoryDatabaseUpPath).Run()
	common.MemoryRuntime.SetGitSyncer(memoryGit)
	common.MemoryRuntime.Configure(common.MemoryConfig{
		Dir:       memoryDir,
		DBName:    memoryDBName,
		DBPath:    filepath.Join(memoryDir, memoryDBName),
		IsGitRepo: isGitRepo,
	}, memoryDB)
	return nil
}
