package common

import (
	"errors"
	"strings"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

const MemorySyncCommitMessage = `chore: sync memory db`
const DefaultMemoryAutoPushDelayMinutes = 1
const memoryAutoSyncTaskType = `memory_db_sync`
const memoryAutoSyncTaskKind = `db_git_sync`
const memoryAutoSyncTaskTarget = `memory_db`
const memoryAutoSyncResumeStrategy = `resume_or_run_now`

var ErrMemoryNotConfigured = errors.New(`请先在配置文件中配置记忆库目录`)

type MemoryConfig struct {
	Dir                  string `json:"memory_dir"`
	DBName               string `json:"memory_db_name"`
	DBPath               string `json:"memory_db_path"`
	IsGitRepo            bool   `json:"is_git_repo"`
	GitRepoEnabled       bool   `json:"git_repo_enabled"`
	AutoPushDelayMinutes int    `json:"auto_push_delay_minutes"`
}

type stoppableTimer interface {
	Stop() bool
}

type memoryGitSyncer interface {
	HasFileChanges(dir, fileName string) (bool, error)
	AddFile(dir, fileName string) error
	Commit(dir, fileName, message string) error
	Push(dir string) error
}

type timeTimer struct {
	timer *time.Timer
}

func (h *timeTimer) Stop() bool {
	return h.timer.Stop()
}

type MemoryStore struct {
	mu           sync.RWMutex
	config       MemoryConfig
	db           MemoryFragmentStore
	timer        stoppableTimer
	dirty        bool
	nextPushTime int64
	lastPushTime int64
	lastPushErr  string
	afterFunc    func(time.Duration, func()) stoppableTimer
	gitSyncer    memoryGitSyncer
	// syncMu 防止并发 git 操作（避免 index.lock 冲突）。 // Prevent concurrent git operations to avoid index.lock conflicts.
	syncMu sync.Mutex
	// pendingTaskID 记录当前防抖窗口对应的异步任务 id，避免重复创建任务。 // Track the async task id for the current debounce window and avoid duplicates.
	pendingTaskID int
	// pendingTaskRecovered 标记当前待处理任务是否来自启动恢复。 // Mark whether the current pending task was restored during startup.
	pendingTaskRecovered bool
	// pendingTaskInitMu 串行化 pending 任务的加载与创建，避免并发慢速路径产生孤立任务。 // Serialize pending-task load/create so concurrent slow paths cannot create orphan tasks.
	pendingTaskInitMu sync.Mutex
	// createAsyncTask 允许测试替换异步任务创建流程。 // Allow tests to replace async task creation.
	createAsyncTask func(config MemoryConfig) (int, error)
	// markAsyncTaskRunning 允许测试替换任务 running 状态更新。 // Allow tests to replace task running updates.
	markAsyncTaskRunning func(taskID int) error
	// markAsyncTaskFailed 允许测试替换任务 failed 状态更新。 // Allow tests to replace task failed updates.
	markAsyncTaskFailed func(taskID int, errMsg string) error
	// markAsyncTaskConfirmed 允许测试替换任务 confirmed 状态更新。 // Allow tests to replace task confirmed updates.
	markAsyncTaskConfirmed func(taskID int) error
	// updateAsyncTaskRequestPayload 允许测试替换任务请求参数更新。 // Allow tests to replace task request payload updates.
	updateAsyncTaskRequestPayload func(taskID int, payload string) error
	// loadPendingTask 允许测试替换 pending 任务加载。 // Allow tests to replace pending task loading during startup recovery.
	loadPendingTask func(config MemoryConfig) (int, string, error)
	// OnStatusChange 在异步任务状态变更时调用，用于 SSE 广播。
	OnStatusChange func()
}

// MemoryFragmentStore 定义知识片段运行时存储接口。
type MemoryFragmentStore interface {
	MemoryFragmentList(limit int) ([]map[string]any, error)
	MemoryFragmentTrashList(limit int) ([]map[string]any, error)
	MemoryFragmentInfo(id any) (map[string]any, error)
	MemoryFragmentSave(id any, title, content string, tags []string) (map[string]any, error)
	MemoryFragmentSoftDelete(id any) (int64, error)
	MemoryFragmentRestore(id any) (int64, error)
	MemoryFragmentHardDelete(id any) error
	MemoryFragmentHistoryList(id any) ([]map[string]any, error)
	MemoryFragmentTagList() ([]map[string]any, error)
	MemoryFragmentSearch(mode, query string, selectedTags []string, limit int) ([]map[string]any, error)
	MemoryFragmentBatchInfoByPaths(paths []string) []map[string]any
}

func NewMemoryStore() *MemoryStore {
	s := &MemoryStore{
		afterFunc: func(duration time.Duration, callback func()) stoppableTimer {
			return &timeTimer{timer: time.AfterFunc(duration, callback)}
		},
	}
	s.createAsyncTask = s.defaultCreateAsyncTask
	s.markAsyncTaskRunning = s.defaultMarkAsyncTaskRunning
	s.markAsyncTaskFailed = s.defaultMarkAsyncTaskFailed
	s.markAsyncTaskConfirmed = s.defaultMarkAsyncTaskConfirmed
	s.updateAsyncTaskRequestPayload = s.defaultUpdateAsyncTaskRequestPayload
	s.loadPendingTask = s.defaultLoadPendingTask
	return s
}

func (h *MemoryStore) SetGitSyncer(syncer memoryGitSyncer) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.gitSyncer = syncer
}

func (h *MemoryStore) Configure(config MemoryConfig, db MemoryFragmentStore) {
	h.mu.Lock()
	h.config = config
	h.db = db
	h.dirty = false
	h.nextPushTime = 0
	h.lastPushTime = 0
	h.lastPushErr = ``
	h.pendingTaskID = 0
	h.pendingTaskRecovered = false
	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}
	h.mu.Unlock()

	// 启动恢复：从数据库中恢复未完成的 pending 任务。 // Restore pending tasks from the database during startup.
	h.restorePendingTask()
}

func (h *MemoryStore) Reset() {
	h.Configure(MemoryConfig{}, nil)
}

// UpdateConfigPreserveState 仅更新配置和防抖计时器，保留 db/dirty/lastPushTime/lastPushErr 等运行状态。
// 适用于仅修改自动同步间隔等配置参数时使用，避免重置已有的待同步数据。
func (h *MemoryStore) UpdateConfigPreserveState(config MemoryConfig) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config = config
	// 如果当前有未触发的定时器，用新间隔重建
	if h.timer != nil && h.dirty {
		h.timer.Stop()
		if config.AutoPushDelayMinutes > 0 {
			h.nextPushTime = time.Now().Add(time.Duration(config.AutoPushDelayMinutes) * time.Minute).Unix()
			h.timer = h.afterFunc(time.Duration(config.AutoPushDelayMinutes)*time.Minute, func() {
				_ = h.syncWithAsyncTask()
			})
		} else {
			h.timer = nil
			h.nextPushTime = 0
		}
	}
}

func (h *MemoryStore) Config() MemoryConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config
}

func (h *MemoryStore) DB() MemoryFragmentStore {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.db
}

func (h *MemoryStore) LastPushTime() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastPushTime
}

func (h *MemoryStore) NextPushTime() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.nextPushTime
}

func (h *MemoryStore) LastPushError() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastPushErr
}

func (h *MemoryStore) IsConfigured() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config.Dir != `` && h.db != nil
}

func (h *MemoryStore) EnsureConfigured() error {
	if h.IsConfigured() {
		return nil
	}
	return ErrMemoryNotConfigured
}

// ScheduleSync 重置防抖计时器，到期后自动执行同步（带异步任务记录）。
func (h *MemoryStore) ScheduleSync() {
	h.mu.Lock()
	if h.config.Dir == `` {
		h.mu.Unlock()
		return
	}
	h.dirty = true
	if h.config.AutoPushDelayMinutes <= 0 {
		if h.timer != nil {
			h.timer.Stop()
			h.timer = nil
		}
		h.nextPushTime = 0
		h.mu.Unlock()
		return
	}
	config := h.config
	pendingTaskID := h.pendingTaskID
	h.mu.Unlock()

	if pendingTaskID == 0 && config.IsGitRepo {
		taskID, err := h.ensurePendingTask(config)
		if err != nil {
			gstool.FmtPrintlnLogTime(`记忆库自动同步创建待处理任务失败 dir=%s err=%s`, config.Dir, err.Error())
		} else {
			pendingTaskID = taskID
		}
	}

	scheduledAt := time.Now().Add(time.Duration(config.AutoPushDelayMinutes) * time.Minute).Unix()
	if pendingTaskID > 0 {
		if err := h.persistPendingTaskSchedule(config, pendingTaskID, scheduledAt); err != nil {
			gstool.FmtPrintlnLogTime(`记忆库自动同步更新待处理任务排期失败 task_id=%d err=%s`, pendingTaskID, err.Error())
		}
	}

	h.mu.Lock()
	if h.timer != nil {
		h.timer.Stop()
	}
	delay := time.Until(time.Unix(scheduledAt, 0))
	if delay < 0 {
		delay = 0
	}
	h.nextPushTime = scheduledAt
	h.timer = h.afterFunc(delay, func() {
		_ = h.syncWithAsyncTask()
	})
	h.mu.Unlock()

	gstool.FmtPrintlnLogTime(`记忆库自动同步已排期，%d 分钟后执行 scheduled_at=%d`, config.AutoPushDelayMinutes, scheduledAt)
}

// SyncNow 立即执行一次同步（供外部手动调用和关闭前同步，不创建异步任务）。
func (h *MemoryStore) SyncNow() error {
	if !h.tryLockSync() {
		gstool.FmtPrintlnLogTime(`记忆库手动同步跳过，已有同步任务正在执行`)
		return nil
	}
	defer h.unlockSync()
	// 清理防抖计时器，防止同步后旧 timer 仍然触发。
	h.mu.Lock()
	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}
	h.nextPushTime = 0
	h.mu.Unlock()
	return h.syncNowInternal()
}

// SyncPendingTaskNow 立即执行当前待处理任务，并沿用异步任务状态流转。 // Flush the current pending task immediately using async-task status transitions.
func (h *MemoryStore) SyncPendingTaskNow() error {
	if !h.tryLockSync() {
		gstool.FmtPrintlnLogTime(`记忆库待处理任务同步跳过，已有同步任务正在执行`)
		return nil
	}
	defer h.unlockSync()
	return h.syncWithAsyncTaskInternal()
}

// HasPendingTask 返回当前是否仍存在待处理任务。 // Report whether a pending async task is currently tracked in memory.
func (h *MemoryStore) HasPendingTask() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.pendingTaskID > 0
}

// Stop 停止防抖计时器。
func (h *MemoryStore) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}
}

// --- 异步任务集成 ---

// syncWithAsyncTask 执行同步并创建异步任务记录，管理完整状态流转。 // Execute sync with full async-task status transitions.
func (h *MemoryStore) syncWithAsyncTask() error {
	gstool.FmtPrintlnLogTime(`记忆库自动同步 syncWithAsyncTask 入口`)
	if !h.tryLockSync() {
		h.mu.RLock()
		dir := h.config.Dir
		h.mu.RUnlock()
		gstool.FmtPrintlnLogTime(`记忆库自动同步跳过，已有同步任务正在执行 dir=%s`, dir)
		return nil
	}
	defer h.unlockSync()
	return h.syncWithAsyncTaskInternal()
}

// syncWithAsyncTaskInternal 执行同步的核心逻辑（调用方需已持有 syncMutex）。 // Core sync logic; caller must hold syncMutex.
func (h *MemoryStore) syncWithAsyncTaskInternal() error {
	h.mu.RLock()
	config := h.config
	h.mu.RUnlock()

	if config.Dir == `` {
		return nil
	}

	// 停止防抖计时器并清理 nextPushTime。
	h.mu.Lock()
	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}
	h.nextPushTime = 0
	h.mu.Unlock()

	taskID, err := h.ensurePendingTask(config)
	if err != nil {
		gstool.FmtPrintlnLogTime(`记忆库自动同步准备待处理任务失败 dir=%s err=%s`, config.Dir, err.Error())
		syncErr := h.syncNowInternal()
		h.clearPendingTaskIDIfMatch(taskID)
		if syncErr != nil {
			return syncErr
		}
		return err
	}

	// 标记为 running。
	if taskID > 0 {
		if markErr := h.markAsyncTaskRunning(taskID); markErr != nil {
			gstool.FmtPrintlnLogTime(`记忆库自动同步标记任务 running 失败 task_id=%d err=%s`, taskID, markErr.Error())
		}
		h.notifyStatusChange()
	}

	// 执行同步。
	syncErr := h.syncNowInternal()

	// 标记终态。
	if taskID > 0 {
		if syncErr != nil {
			if markErr := h.markAsyncTaskFailed(taskID, syncErr.Error()); markErr != nil {
				gstool.FmtPrintlnLogTime(`记忆库自动同步标记任务 failed 失败 task_id=%d err=%s`, taskID, markErr.Error())
			}
		} else {
			if markErr := h.markAsyncTaskConfirmed(taskID); markErr != nil {
				gstool.FmtPrintlnLogTime(`记忆库自动同步标记任务 confirmed 失败 task_id=%d err=%s`, taskID, markErr.Error())
			}
		}
		h.notifyStatusChange()
	}
	h.clearPendingTaskIDIfMatch(taskID)
	return syncErr
}

// syncNowInternal 执行实际的 git 同步操作（不带异步任务记录）。 // Execute the actual git sync without async task tracking.
func (h *MemoryStore) syncNowInternal() error {
	h.mu.RLock()
	config := h.config
	syncer := h.gitSyncer
	dirty := h.dirty
	h.mu.RUnlock()

	if config.Dir == `` {
		h.setLastPushError(ErrMemoryNotConfigured.Error())
		gstool.FmtPrintlnLogTime(`记忆库同步失败：配置不完整 dir=%s`, config.Dir)
		return ErrMemoryNotConfigured
	}
	if !config.IsGitRepo {
		// 未开启 git 模式时直接清理脏标记。 // Clear dirty state directly when git sync is disabled.
		h.mu.Lock()
		h.dirty = false
		h.lastPushErr = ``
		h.mu.Unlock()
		gstool.FmtPrintlnLogTime(`记忆库未启用 git 仓库同步，跳过 push dir=%s`, config.Dir)
		return nil
	}
	if !dirty {
		// 没有脏数据时不需要触发 git push。 // Skip git push when there are no pending changes.
		gstool.FmtPrintlnLogTime(`记忆库当前没有待同步变更，跳过 push dir=%s`, config.Dir)
		return nil
	}
	if syncer == nil {
		err := errors.New(`memory git syncer not set`)
		h.setLastPushError(err.Error())
		gstool.FmtPrintlnLogTime(`记忆库同步失败：git syncer 未设置 dir=%s`, config.Dir)
		return err
	}

	target := `.`
	if config.DBPath != `` {
		target = config.DBPath
	}
	gstool.FmtPrintlnLogTime(`记忆库开始检查变更并执行 push dir=%s target=%s`, config.Dir, target)
	hasChanges, err := syncer.HasFileChanges(config.Dir, target)
	if err != nil {
		h.setLastPushError(err.Error())
		gstool.FmtPrintlnLogTime(`记忆库检查变更失败 dir=%s target=%s err=%s`, config.Dir, target, err.Error())
		return err
	}
	if !hasChanges {
		h.mu.Lock()
		h.dirty = false
		h.lastPushErr = ``
		h.mu.Unlock()
		gstool.FmtPrintlnLogTime(`记忆库未检测到文件变更，跳过 push dir=%s target=%s`, config.Dir, target)
		return nil
	}
	if err = syncer.AddFile(config.Dir, target); err != nil {
		h.setLastPushError(err.Error())
		gstool.FmtPrintlnLogTime(`记忆库 add 失败 dir=%s target=%s err=%s`, config.Dir, target, err.Error())
		return err
	}
	if err = syncer.Commit(config.Dir, target, MemorySyncCommitMessage); err != nil {
		h.setLastPushError(err.Error())
		gstool.FmtPrintlnLogTime(`记忆库 commit 失败 dir=%s target=%s err=%s`, config.Dir, target, err.Error())
		return err
	}
	if err = syncer.Push(config.Dir); err != nil {
		h.setLastPushError(err.Error())
		gstool.FmtPrintlnLogTime(`记忆库 push 失败 dir=%s target=%s err=%s`, config.Dir, target, err.Error())
		return err
	}

	h.mu.Lock()
	h.dirty = false
	h.lastPushTime = time.Now().Unix()
	h.lastPushErr = ``
	h.mu.Unlock()
	gstool.FmtPrintlnLogTime(`记忆库 push 成功 dir=%s target=%s`, config.Dir, target)
	return nil
}

// --- 并发控制 ---

// tryLockSync 尝试获取同步互斥锁，如果已有同步在执行则返回 false。 // Try to acquire the sync mutex; return false if a sync is already running.
func (h *MemoryStore) tryLockSync() bool {
	ok := h.syncMu.TryLock()
	gstool.FmtPrintlnLogTime(`记忆库自动同步 tryLockSync result=%v`, ok)
	return ok
}

// unlockSync 释放同步互斥锁。 // Release the sync mutex.
func (h *MemoryStore) unlockSync() {
	h.syncMu.Unlock()
}

func (h *MemoryStore) setLastPushError(message string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastPushErr = message
}

// notifyStatusChange 安全调用 OnStatusChange 回调。
func (h *MemoryStore) notifyStatusChange() {
	h.mu.RLock()
	cb := h.OnStatusChange
	h.mu.RUnlock()
	if cb != nil {
		cb()
	}
}

// --- 异步任务持久化 ---

// ensurePendingTask 确保当前防抖窗口存在一条 pending 异步任务。 // Ensure the current debounce window has one pending async task.
// 使用双重检查锁定防止并发创建重复任务。 // Uses double-checked locking to prevent concurrent duplicate task creation.
func (h *MemoryStore) ensurePendingTask(config MemoryConfig) (int, error) {
	// 快速路径：读锁检查是否已有待处理任务。 // Fast path: check under read lock.
	h.mu.RLock()
	if h.pendingTaskID > 0 {
		taskID := h.pendingTaskID
		h.mu.RUnlock()
		return taskID, nil
	}
	h.mu.RUnlock()

	// 慢速路径：使用独立初始化锁串行化 load/create，避免两个 goroutine 同时创建孤立任务。
	// English comment: Serialize load/create with a dedicated init lock so concurrent goroutines cannot create orphan pending tasks.
	h.pendingTaskInitMu.Lock()
	defer h.pendingTaskInitMu.Unlock()

	// 双重检查：进入慢速路径后再次确认，避免等待初始化锁期间其他 goroutine 已完成创建。
	// English comment: Re-check after entering the serialized slow path in case another goroutine already created the task.
	h.mu.Lock()
	if h.pendingTaskID > 0 {
		taskID := h.pendingTaskID
		h.mu.Unlock()
		return taskID, nil
	}

	loadPendingTask := h.loadPendingTask
	createTask := h.createAsyncTask
	h.mu.Unlock()

	// 中文注释：load/create 都可能触发查库或状态广播，必须放到 h.mu 外执行，避免回调再次读取同一把锁导致自锁。
	// English comment: load/create may trigger DB reads or status callbacks, so they must run outside h.mu to avoid self-deadlock on re-entrant lock access.
	if loadPendingTask != nil {
		taskID, _, err := loadPendingTask(config)
		if err != nil {
			return 0, err
		}
		if taskID > 0 {
			h.mu.Lock()
			if h.pendingTaskID == 0 {
				h.pendingTaskID = taskID
				h.pendingTaskRecovered = false
			}
			resultTaskID := h.pendingTaskID
			h.mu.Unlock()
			h.notifyStatusChange()
			return resultTaskID, nil
		}
	}
	if createTask == nil {
		return 0, nil
	}
	// 中文注释：createTask 在 pendingTaskInitMu 保护下执行，保证同一时刻只会创建一条 pending 记录。
	// English comment: createTask runs under pendingTaskInitMu so only one pending record can be created at a time.
	taskID, err := createTask(config)
	if err != nil {
		return 0, err
	}
	if taskID <= 0 {
		return 0, nil
	}
	h.mu.Lock()
	if h.pendingTaskID == 0 {
		h.pendingTaskID = taskID
		h.pendingTaskRecovered = false
	}
	resultTaskID := h.pendingTaskID
	h.mu.Unlock()
	h.notifyStatusChange()
	return resultTaskID, nil
}

// clearPendingTaskIDIfMatch 仅当内存中的 pendingTaskID 与 taskID 一致时才清零，避免误清新创建的任务 ID。 // Clear pendingTaskID only when it matches taskID to prevent overwriting a newly created task.
func (h *MemoryStore) clearPendingTaskIDIfMatch(taskID int) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.pendingTaskID == taskID {
		h.pendingTaskID = 0
		h.pendingTaskRecovered = false
	}
}

// defaultCreateAsyncTask 创建记忆库自动同步待处理任务。 // Create the pending async task for memory auto sync.
func (h *MemoryStore) defaultCreateAsyncTask(config MemoryConfig) (int, error) {
	taskTitle := `记忆库自动同步 ` + time.Now().Format(`2006-01-02 15:04:05`)
	if DbLog == nil || DbLog.Client == nil {
		gstool.FmtPrintlnLogTime(`记忆库自动同步日志库未初始化，跳过任务记录`)
		return 0, nil
	}
	affected, delErr := DbLog.AsyncTaskDeleteByType(memoryAutoSyncTaskType)
	if delErr != nil {
		gstool.FmtPrintlnLogTime(`记忆库自动同步清理旧任务失败 err=%s`, delErr.Error())
	} else if affected > 0 {
		gstool.FmtPrintlnLogTime(`记忆库自动同步已清理旧任务 count=%d`, affected)
		h.notifyStatusChange()
	}
	scheduledAt := time.Now().Add(time.Duration(config.AutoPushDelayMinutes) * time.Minute).Unix()
	taskInfo, err := DbLog.AsyncTaskCreate(memoryAutoSyncTaskType, taskTitle, ``, h.buildTaskPayload(config, scheduledAt))
	if err != nil {
		return 0, err
	}
	if id, ok := taskInfo[`id`]; ok {
		return castToInt(id), nil
	}
	return 0, nil
}

// defaultMarkAsyncTaskRunning 将记忆库自动同步任务标记为 running。 // Mark the memory auto sync task as running.
func (h *MemoryStore) defaultMarkAsyncTaskRunning(taskID int) error {
	if DbLog == nil || DbLog.Client == nil || taskID <= 0 {
		return nil
	}
	return DbLog.AsyncTaskMarkRunning(taskID)
}

// defaultMarkAsyncTaskFailed 将记忆库自动同步任务标记为 failed。 // Mark the memory auto sync task as failed.
func (h *MemoryStore) defaultMarkAsyncTaskFailed(taskID int, errMsg string) error {
	if DbLog == nil || DbLog.Client == nil || taskID <= 0 {
		return nil
	}
	return DbLog.AsyncTaskMarkFailed(taskID, errMsg)
}

// defaultMarkAsyncTaskConfirmed 将记忆库自动同步任务标记为 confirmed。 // Mark the memory auto sync task as confirmed.
func (h *MemoryStore) defaultMarkAsyncTaskConfirmed(taskID int) error {
	if DbLog == nil || DbLog.Client == nil || taskID <= 0 {
		return nil
	}
	return DbLog.AsyncTaskMarkFinal(taskID, AsyncTaskStatusConfirmed)
}

// defaultUpdateAsyncTaskRequestPayload 更新待处理任务的请求参数。 // Update request payload for an existing pending task.
func (h *MemoryStore) defaultUpdateAsyncTaskRequestPayload(taskID int, payload string) error {
	if DbLog == nil || DbLog.Client == nil || taskID <= 0 {
		return nil
	}
	return DbLog.AsyncTaskUpdateRequestPayload(taskID, payload)
}

// defaultLoadPendingTask 读取最新的记忆库待处理同步任务用于启动恢复。 // Load the latest pending memory sync task for startup recovery.
func (h *MemoryStore) defaultLoadPendingTask(config MemoryConfig) (int, string, error) {
	if DbLog == nil || DbLog.Client == nil {
		return 0, ``, nil
	}
	taskInfo, err := DbLog.AsyncTaskLatestPendingByType(memoryAutoSyncTaskType)
	if err != nil || taskInfo == nil {
		return 0, ``, err
	}
	payload := castToString(taskInfo[`request_payload`])
	payloadMap, parseErr := h.parseTaskPayload(payload)
	if parseErr != nil {
		return 0, ``, parseErr
	}
	if !h.isPayloadMatchCurrentConfig(config, payloadMap) {
		return 0, ``, nil
	}
	return castToInt(taskInfo[`id`]), payload, nil
}

// --- 恢复机制 ---

// restorePendingTask 从持久化任务中恢复待处理同步计时器。 // Restore pending sync timer from persisted task metadata.
func (h *MemoryStore) restorePendingTask() {
	h.mu.RLock()
	config := h.config
	loadPendingTask := h.loadPendingTask
	h.mu.RUnlock()

	if loadPendingTask == nil || config.Dir == `` {
		return
	}
	taskID, payload, err := loadPendingTask(config)
	if err != nil {
		gstool.FmtPrintlnLogTime(`记忆库自动同步恢复待处理任务失败 dir=%s dbPath=%s dbName=%s err=%s`, config.Dir, config.DBPath, config.DBName, err.Error())
		return
	}
	if taskID <= 0 || strings.TrimSpace(payload) == `` {
		return
	}
	scheduledAt, parseErr := h.extractScheduledAt(payload)
	if parseErr != nil {
		gstool.FmtPrintlnLogTime(`记忆库自动同步解析恢复任务排期失败 task_id=%d err=%s`, taskID, parseErr.Error())
		return
	}
	h.mu.Lock()
	h.pendingTaskID = taskID
	h.pendingTaskRecovered = true
	h.dirty = true
	h.mu.Unlock()
	if err := h.persistPendingTaskSchedule(config, taskID, scheduledAt); err != nil {
		gstool.FmtPrintlnLogTime(`记忆库自动同步回写恢复任务描述失败 task_id=%d err=%s`, taskID, err.Error())
	}
	gstool.FmtPrintlnLogTime(`记忆库自动同步恢复待处理任务 task_id=%d scheduled_at=%d`, taskID, scheduledAt)
	h.notifyStatusChange()
	h.scheduleRecoveredTask(config, scheduledAt)
}

// scheduleRecoveredTask 根据恢复的计划时间重新挂载 timer，过期任务立即执行。 // Re-schedule recovered pending task and run immediately when overdue.
func (h *MemoryStore) scheduleRecoveredTask(config MemoryConfig, scheduledAt int64) {
	delay := time.Until(time.Unix(scheduledAt, 0))
	if delay < 0 {
		delay = 0
	}
	h.mu.Lock()
	if h.timer != nil {
		h.timer.Stop()
	}
	h.nextPushTime = scheduledAt
	h.timer = h.afterFunc(delay, func() {
		_ = h.syncWithAsyncTask()
	})
	taskID := h.pendingTaskID
	h.mu.Unlock()
	gstool.FmtPrintlnLogTime(`记忆库自动同步恢复排期 task_id=%d delay_ms=%d`, taskID, delay.Milliseconds())
}

// persistPendingTaskSchedule 回写待处理任务的调度参数，保证重启后可恢复。 // Persist pending task schedule metadata for restart recovery.
func (h *MemoryStore) persistPendingTaskSchedule(config MemoryConfig, taskID int, scheduledAt int64) error {
	h.mu.RLock()
	restored := h.pendingTaskRecovered
	h.mu.RUnlock()
	updatePayload := h.buildTaskPayloadWithRecovery(config, scheduledAt, restored)
	updateTaskPayload := h.updateAsyncTaskRequestPayload
	if updateTaskPayload == nil {
		return nil
	}
	return updateTaskPayload(taskID, updatePayload)
}

// --- payload 构建 ---

// buildTaskPayload 构造通用可恢复调度元数据。 // Build a generic resumable payload for memory auto sync tasks.
func (h *MemoryStore) buildTaskPayload(config MemoryConfig, scheduledAt int64) string {
	return h.buildTaskPayloadWithRecovery(config, scheduledAt, false)
}

// buildTaskPayloadWithRecovery 构造带恢复标记的记忆库同步 payload。 // Build memory sync payload with optional recovery flag.
func (h *MemoryStore) buildTaskPayloadWithRecovery(config MemoryConfig, scheduledAt int64, restored bool) string {
	scheduledAtDesc := formatMemoryAutoSyncScheduleTime(scheduledAt)
	taskDescription := `已检测到记忆库变更，预计在 ` + scheduledAtDesc + ` 自动同步`
	if restored {
		taskDescription = `已从未完成任务恢复，预计在 ` + scheduledAtDesc + ` 自动同步`
	}
	payload := map[string]any{
		`task_kind`:        memoryAutoSyncTaskKind,
		`task_target`:      memoryAutoSyncTaskTarget,
		`task_description`: taskDescription,
		`task_params`: map[string]any{
			`dir`:         config.Dir,
			`target_name`: config.DBPath,
			`target_type`: `markdown_files`,
		},
		`schedule`: map[string]any{
			`trigger`:              `controller`,
			`scheduled_at`:         scheduledAt,
			`scheduled_at_desc`:    scheduledAtDesc,
			`delay_minutes`:        config.AutoPushDelayMinutes,
			`created_from_runtime`: `memory_auto_sync`,
		},
		`resume`: map[string]any{
			`resume_key`:      h.buildResumeKey(config),
			`resume_strategy`: memoryAutoSyncResumeStrategy,
			`restored`:        restored,
		},
	}
	return gstool.JsonEncode(payload)
}

// buildResumeKey 构造当前记忆库同步任务的恢复键。 // Build the resume key for the current memory sync task.
func (h *MemoryStore) buildResumeKey(config MemoryConfig) string {
	return memoryAutoSyncTaskKind + `:` + memoryAutoSyncTaskTarget + `:` + config.Dir
}

// parseTaskPayload 解析记忆库同步任务 payload。 // Decode the memory sync payload.
func (h *MemoryStore) parseTaskPayload(payload string) (map[string]any, error) {
	payloadMap := make(map[string]any)
	if err := gstool.JsonDecode(payload, &payloadMap); err != nil {
		return nil, err
	}
	return payloadMap, nil
}

// extractScheduledAt 从 payload 中读取计划执行时间。 // Read scheduled_at from the persisted task payload.
func (h *MemoryStore) extractScheduledAt(payload string) (int64, error) {
	payloadMap, err := h.parseTaskPayload(payload)
	if err != nil {
		return 0, err
	}
	scheduleMap, _ := payloadMap[`schedule`].(map[string]any)
	if scheduleMap == nil {
		return 0, errors.New(`任务缺少 schedule 配置`)
	}
	return castToInt64(scheduleMap[`scheduled_at`]), nil
}

// isPayloadMatchCurrentConfig 校验恢复任务是否仍然匹配当前记忆库配置。 // Check whether a persisted pending task still matches current memory config.
func (h *MemoryStore) isPayloadMatchCurrentConfig(config MemoryConfig, payloadMap map[string]any) bool {
	if payloadMap[`task_kind`] != memoryAutoSyncTaskKind || payloadMap[`task_target`] != memoryAutoSyncTaskTarget {
		return false
	}
	resumeMap, _ := payloadMap[`resume`].(map[string]any)
	if resumeMap == nil {
		return false
	}
	return castToString(resumeMap[`resume_key`]) == h.buildResumeKey(config)
}

func formatMemoryAutoSyncScheduleTime(ts int64) string {
	if ts <= 0 {
		return `-`
	}
	return time.Unix(ts, 0).Format(`2006-01-02 15:04:05`)
}
