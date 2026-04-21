package controller

import (
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/pkg/p_define"
	"path/filepath"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
)

// gitPendingStatusChecker 抽象 git 状态查询能力，方便复用和单元测试。
// gitPendingStatusChecker abstracts git status checks for reuse and unit testing.
type gitPendingStatusChecker interface {
	IsGitRepo(dir string) (bool, error)
	HasFileChanges(dir, fileName string) (bool, error)
}

// detectGitPendingStatus 统一计算主库和记忆库红点状态，确保提示口径与实际同步行为一致。
// detectGitPendingStatus computes badge flags and keeps them aligned with real sync behavior.
func detectGitPendingStatus(gitSyncer gitPendingStatusChecker, mainConfig business.MainDBConfig, memoryConfig common.MemoryConfig, hasMainDBEnv bool) (bool, bool) {
	mainDBPending := false
	memoryPending := false

	// 中文注释：主库同步实际只处理 sqlite 主文件，因此红点也只跟随主文件变更。
	// English comment: Main-db sync only targets the primary sqlite file, so the badge follows that file only.
	if hasMainDBEnv && mainConfig.GitRepoEnabled && mainConfig.Dir != `` && mainConfig.DBPath != `` {
		if isGit, err := gitSyncer.IsGitRepo(mainConfig.Dir); err == nil && isGit {
			fileName := filepath.Base(mainConfig.DBPath)
			if hasChanges, checkErr := gitSyncer.HasFileChanges(mainConfig.Dir, fileName); checkErr == nil && hasChanges {
				mainDBPending = true
			}
		}
	}

	// 中文注释：记忆库同步以整个目录为目标，所以这里继续使用目录级变更判断。
	// English comment: Memory sync operates on the whole directory, so the badge remains directory-based.
	if memoryConfig.GitRepoEnabled && memoryConfig.Dir != `` {
		if isGit, err := gitSyncer.IsGitRepo(memoryConfig.Dir); err == nil && isGit {
			if hasChanges, checkErr := gitSyncer.HasFileChanges(memoryConfig.Dir, `.`); checkErr == nil && hasChanges {
				memoryPending = true
			}
		}
	}

	return mainDBPending, memoryPending
}

// GitPendingStatus 检测主库和记忆库是否存在待提交的 git 变更。
// GitPendingStatus reports whether the main db or memory db has pending git changes.
func GitPendingStatus(c *gin.Context) {
	gitSyncer := business.NewMemoryGit()
	mainConfig := business.ReadMainDBConfig()
	memoryConfig := business.ReadMemoryConfigFromINI()
	mainDBPending, memoryPending := detectGitPendingStatus(gitSyncer, mainConfig, memoryConfig, component.EnvClient != nil && component.EnvClient.DbConfig != nil)

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`main_db_pending`: mainDBPending,
		`memory_pending`:  memoryPending,
	})
}

// buildGitPendingStatusPayload 构造 Git 待提交状态及倒计时数据。
func buildGitPendingStatusPayload() map[string]any {
	gitSyncer := business.NewMemoryGit()
	mainConfig := business.ReadMainDBConfig()
	memoryConfig := business.ReadMemoryConfigFromINI()
	mainDBPending, memoryPending := detectGitPendingStatus(gitSyncer, mainConfig, memoryConfig, component.EnvClient != nil && component.EnvClient.DbConfig != nil)

	// 中文注释：从运行时组件读取排期时间，无需再次执行 git status。
	var mainDBNextPush int64
	var mainDBInterval int
	if component.MainDBAutoSyncRuntime != nil {
		mainDBNextPush = component.MainDBAutoSyncRuntime.NextSyncTime()
		mainDBInterval = component.MainDBAutoSyncRuntime.Config().AutoSyncMinutes * 60
	}

	var memoryNextPush int64
	var memoryInterval int
	if component.MemoryRuntime != nil {
		memoryNextPush = component.MemoryRuntime.NextPushTime()
		memoryInterval = component.MemoryRuntime.Config().AutoPushDelayMinutes * 60
	}

	return map[string]any{
		`main_db_pending`:   mainDBPending,
		`memory_pending`:    memoryPending,
		`main_db_next_push`: mainDBNextPush,
		`memory_next_push`:  memoryNextPush,
		`main_db_interval`:  mainDBInterval,
		`memory_interval`:   memoryInterval,
	}
}

// sendGitPendingStatusSnapshot 向指定 SSE 连接发送一次 Git 待提交状态快照。
func sendGitPendingStatusSnapshot(sse *gsgin.Sse) {
	if sse == nil {
		return
	}
	data := buildGitPendingStatusPayload()
	err := sse.SendToChan(gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseGitPendingStatus,
		Data:            data,
		Type:            p_define.SseContentTypeMsg,
	}))
	if err != nil {
		gstool.FmtPrintlnLogTime(`GitPendingStatus广播错误 %s`, err.Error())
	}
}

// BindGitPendingStatusSSE 为普通 SSE client 绑定 Git 待提交状态推送。
func BindGitPendingStatusSSE(sse *gsgin.Sse, stopC chan int, interval time.Duration) {
	if sse == nil {
		return
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	// 建连后立即推一次，避免前端初次打开时要等下一个周期。
	sendGitPendingStatusSnapshot(sse)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sendGitPendingStatusSnapshot(sse)
			case <-stopC:
				return
			}
		}
	}()
}
