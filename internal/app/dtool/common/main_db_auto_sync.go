package common

import (
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/fsnotify/fsnotify"
)

const DefaultMainDBAutoPushDelayMinutes = 10
const MainDBAutoSyncCommitMessage = `chore: auto sync main db`
const mainDBAutoSyncTaskType = `main_db_sync`
const mainDBAutoSyncTaskKind = `db_git_sync`
const mainDBAutoSyncTaskTarget = `main_db`
const mainDBAutoSyncResumeStrategy = `resume_or_run_now`

var ErrMainDBNotConfigured = errors.New(`主库未配置，无法执行自动同步`)

// MainDBAutoSyncConfig 描述主库自动同步配置。
type MainDBAutoSyncConfig struct {
	Dir             string
	DBName          string
	IsGitRepo       bool
	GitRepoEnabled  bool
	AutoSyncMinutes int
}

// MainDBGitSyncer 定义主库 git 同步所需能力。
type MainDBGitSyncer interface {
	HasFileChanges(dir, fileName string) (bool, error)
	AddFile(dir, fileName string) error
	Commit(dir, fileName, message string) error
	Push(dir string) error
}

// MainDBAutoSync 管理主库的自动同步（git add + commit + push）。
// 通过 fsnotify 监听 db 文件变更，防抖 AutoSyncMinutes 分钟后触发同步。
type MainDBAutoSync struct {
	mu            sync.RWMutex
	config        MainDBAutoSyncConfig
	syncer        MainDBGitSyncer
	watcher       *fsnotify.Watcher
	debounceTimer stoppableTimer
	lastSyncTime  int64
	lastSyncErr   string
	// syncMu 防止并发 git 操作（避免 index.lock 冲突）。 // Prevent concurrent git operations to avoid index.lock conflicts.
	syncMu sync.Mutex
	// pendingTaskID 记录当前防抖窗口对应的异步任务 id，避免重复创建任务。 // Track the async task id for the current debounce window and avoid duplicates.
	pendingTaskID int
	// pendingTaskRecovered 标记当前待处理任务是否来自启动恢复。 // Mark whether the current pending task was restored during startup.
	pendingTaskRecovered bool
	// afterFunc 允许测试替换防抖计时器实现。 // Allow tests to replace the debounce timer implementation.
	afterFunc func(time.Duration, func()) stoppableTimer
	// createAsyncTask 允许测试替换异步任务创建流程。 // Allow tests to replace async task creation.
	createAsyncTask func(config MainDBAutoSyncConfig) (int, error)
	// markAsyncTaskRunning 允许测试替换任务 running 状态更新。 // Allow tests to replace task running updates.
	markAsyncTaskRunning func(taskID int) error
	// markAsyncTaskFailed 允许测试替换任务 failed 状态更新。 // Allow tests to replace task failed updates.
	markAsyncTaskFailed func(taskID int, errMsg string) error
	// markAsyncTaskConfirmed 允许测试替换任务 confirmed 状态更新。 // Allow tests to replace task confirmed updates.
	markAsyncTaskConfirmed func(taskID int) error
	// updateAsyncTaskRequestPayload 允许测试替换任务请求参数更新。 // Allow tests to replace task request payload updates.
	updateAsyncTaskRequestPayload func(taskID int, payload string) error
	// loadPendingTask 允许测试替换 pending 任务加载。 // Allow tests to replace pending task loading during startup recovery.
	loadPendingTask func(config MainDBAutoSyncConfig) (int, string, error)
	// OnStatusChange 在异步任务状态变更时调用，用于 SSE 广播。
	OnStatusChange func()
}

// NewMainDBAutoSync 创建主库自动同步实例。
func NewMainDBAutoSync() *MainDBAutoSync {
	runtime := &MainDBAutoSync{}
	runtime.afterFunc = func(duration time.Duration, callback func()) stoppableTimer {
		return &timeTimer{timer: time.AfterFunc(duration, callback)}
	}
	runtime.createAsyncTask = runtime.defaultCreateAsyncTask
	runtime.markAsyncTaskRunning = runtime.defaultMarkAsyncTaskRunning
	runtime.markAsyncTaskFailed = runtime.defaultMarkAsyncTaskFailed
	runtime.markAsyncTaskConfirmed = runtime.defaultMarkAsyncTaskConfirmed
	runtime.updateAsyncTaskRequestPayload = runtime.defaultUpdateAsyncTaskRequestPayload
	runtime.loadPendingTask = runtime.defaultLoadPendingTask
	return runtime
}

// Configure 更新自动同步配置和 git syncer，若已在运行会先停止再重启。
func (h *MainDBAutoSync) Configure(config MainDBAutoSyncConfig, syncer MainDBGitSyncer) {
	h.mu.Lock()
	wasRunning := h.watcher != nil
	h.config = config
	h.syncer = syncer
	h.mu.Unlock()

	if wasRunning {
		h.Stop()
		h.Start()
	}
}

// Start 启动 fsnotify 文件监听，监听主库 db 文件变更。
func (h *MainDBAutoSync) Start() {
	h.mu.Lock()
	config := h.config
	h.mu.Unlock()

	if config.Dir == `` || config.DBName == `` {
		gstool.FmtPrintlnLogTime(`主库自动同步未配置完整，跳过启动 dir=%s db=%s`, config.Dir, config.DBName)
		return
	}
	if !config.IsGitRepo {
		gstool.FmtPrintlnLogTime(`主库未启用 git 仓库同步，跳过自动同步 dir=%s`, config.Dir)
		return
	}
	if config.AutoSyncMinutes <= 0 {
		gstool.FmtPrintlnLogTime(`主库自动同步间隔为 0，已关闭 dir=%s`, config.Dir)
		return
	}

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		gstool.FmtPrintlnLogTime(`主库自动同步 fsnotify 初始化失败 err=%s`, err.Error())
		return
	}
	// 监听目录而非单个文件，因为 fsnotify 对单文件监听在部分 OS 上不可靠。
	if err = watcher.Add(config.Dir); err != nil {
		gstool.FmtPrintlnLogTime(`主库自动同步监听目录失败 dir=%s err=%s`, config.Dir, err.Error())
		watcher.Close()
		return
	}

	h.mu.Lock()
	h.watcher = watcher
	h.mu.Unlock()

	mainDBPath := filepath.Join(config.Dir, config.DBName)
	walPath := mainDBPath + `-wal`
	shmPath := mainDBPath + `-shm`
	gstool.FmtPrintlnLogTime(`主库自动同步已启动，监听目录=%s 主文件=%s wal=%s shm=%s 防抖=%d分钟`, config.Dir, mainDBPath, walPath, shmPath, config.AutoSyncMinutes)
	h.restorePendingTask()
	go h.watchEvents()
}

// Stop 停止文件监听和防抖计时器。
func (h *MainDBAutoSync) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.debounceTimer != nil {
		h.debounceTimer.Stop()
		h.debounceTimer = nil
	}
	if h.watcher != nil {
		h.watcher.Close()
		h.watcher = nil
	}
}

// SyncNow 立即执行一次同步（供外部手动调用和关闭前同步，不创建异步任务）。
func (h *MainDBAutoSync) SyncNow() error {
	if !h.tryLockSync() {
		gstool.FmtPrintlnLogTime(`主库手动同步跳过，已有同步任务正在执行`)
		return nil
	}
	defer h.unlockSync()
	return h.syncNow()
}

// SyncPendingTaskNow 立即执行当前待处理任务，并沿用异步任务状态流转。 // Flush the current pending task immediately using async-task status transitions.
func (h *MainDBAutoSync) SyncPendingTaskNow() error {
	if !h.tryLockSync() {
		gstool.FmtPrintlnLogTime(`主库待处理任务同步跳过，已有同步任务正在执行`)
		return nil
	}
	defer h.unlockSync()
	return h.syncWithAsyncTaskInternal()
}

// HasPendingTask 返回当前是否仍存在待处理任务。 // Report whether a pending async task is currently tracked in memory.
func (h *MainDBAutoSync) HasPendingTask() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.pendingTaskID > 0
}

// ScheduleSync 重置防抖计时器，到期后自动执行同步（带异步任务记录）。
// 类似 MemoryStore.ScheduleSync() 的防抖模式。
func (h *MainDBAutoSync) ScheduleSync() {
	h.mu.Lock()
	config := h.config
	pendingTaskID := h.pendingTaskID
	h.mu.Unlock()

	if config.Dir == `` || config.DBName == `` || !config.IsGitRepo {
		return
	}
	if config.AutoSyncMinutes <= 0 {
		return
	}

	scheduledAt := time.Now().Add(time.Duration(config.AutoSyncMinutes) * time.Minute).Unix()
	if pendingTaskID > 0 {
		if err := h.persistPendingTaskSchedule(config, pendingTaskID, scheduledAt); err != nil {
			gstool.FmtPrintlnLogTime(`主库自动同步更新待处理任务排期失败 task_id=%d err=%s`, pendingTaskID, err.Error())
		}
	}

	h.mu.Lock()
	if h.debounceTimer != nil {
		h.debounceTimer.Stop()
	}
	delay := time.Until(time.Unix(scheduledAt, 0))
	if delay < 0 {
		delay = 0
	}
	h.debounceTimer = h.afterFunc(delay, func() {
		_ = h.syncWithAsyncTask()
	})
	h.mu.Unlock()

	gstool.FmtPrintlnLogTime(`主库自动同步已排期，%d 分钟后执行 scheduled_at=%d`, config.AutoSyncMinutes, scheduledAt)
}

// Config 返回当前配置快照。
func (h *MainDBAutoSync) Config() MainDBAutoSyncConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config
}

// LastSyncTime 返回上次成功同步的 unix 时间戳。
func (h *MainDBAutoSync) LastSyncTime() int64 {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastSyncTime
}

// LastSyncError 返回上次同步错误信息。
func (h *MainDBAutoSync) LastSyncError() string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.lastSyncErr
}

// watchEvents 消费 fsnotify 事件，过滤出主库 db 文件的变更并触发防抖。
func (h *MainDBAutoSync) watchEvents() {
	for {
		h.mu.RLock()
		w := h.watcher
		config := h.config
		h.mu.RUnlock()

		if w == nil {
			return
		}

		select {
		case event, ok := <-w.Events:
			if !ok {
				return
			}
			h.handleFileEvent(event, config)
		case _, ok := <-w.Errors:
			if !ok {
				return
			}
		}
	}
}

// handleFileEvent 过滤 fsnotify 事件，处理主库 db 文件及其 WAL 附属文件的 Write/Create。
// SQLite WAL 模式下，数据变更写入 .db-wal 文件而非 .db 主文件，
// 因此必须同时监听 .db、.db-wal、.db-shm 三种文件的变更。
func (h *MainDBAutoSync) handleFileEvent(event fsnotify.Event, config MainDBAutoSyncConfig) {
	// 只处理 Write 和 Create 事件。
	if event.Op&fsnotify.Write == 0 && event.Op&fsnotify.Create == 0 && event.Op&fsnotify.Rename == 0 {
		return
	}
	// 过滤：只处理 DBName 主文件及其 WAL 附属文件（.db-wal、.db-shm）。
	eventBase := filepath.Base(event.Name)
	if eventBase != config.DBName && eventBase != config.DBName+`-wal` && eventBase != config.DBName+`-shm` {
		return
	}

	gstool.FmtPrintlnLogTime(`主库自动同步检测到目标文件变更 file=%s op=%s target_db=%s`, event.Name, formatFSNotifyOp(event.Op), filepath.Join(config.Dir, config.DBName))
	if _, err := h.ensurePendingTask(config); err != nil {
		gstool.FmtPrintlnLogTime(`主库自动同步创建待处理任务失败 dir=%s file=%s err=%s`, config.Dir, config.DBName, err.Error())
	}
	h.ScheduleSync()
}

// syncWithAsyncTask 执行同步并创建异步任务记录，管理完整状态流转。
// 使用互斥锁保证同一时间只有一个同步操作在执行，避免并发 git 操作产生 index.lock。
func (h *MainDBAutoSync) syncWithAsyncTask() error {
	gstool.FmtPrintlnLogTime(`主库自动同步 syncWithAsyncTask 入口`)
	if !h.tryLockSync() {
		h.mu.RLock()
		dir := h.config.Dir
		dbName := h.config.DBName
		h.mu.RUnlock()
		gstool.FmtPrintlnLogTime(`主库自动同步跳过，已有同步任务正在执行 dir=%s file=%s`, dir, dbName)
		return nil
	}
	defer h.unlockSync()
	return h.syncWithAsyncTaskInternal()
}

// syncWithAsyncTaskInternal 执行同步的核心逻辑（调用方需已持有 syncMutex）。 // Core sync logic; caller must hold syncMutex.
func (h *MainDBAutoSync) syncWithAsyncTaskInternal() error {
	h.mu.RLock()
	config := h.config
	h.mu.RUnlock()

	if config.Dir == `` || config.DBName == `` || !config.IsGitRepo {
		return nil
	}

	taskID, err := h.ensurePendingTask(config)
	if err != nil {
		gstool.FmtPrintlnLogTime(`主库自动同步准备待处理任务失败 dir=%s file=%s err=%s`, config.Dir, config.DBName, err.Error())
		syncErr := h.syncNow()
		h.clearPendingTaskID()
		if syncErr != nil {
			return syncErr
		}
		return err
	}

	// 标记为 running。
	if taskID > 0 {
		if markErr := h.markAsyncTaskRunning(taskID); markErr != nil {
			gstool.FmtPrintlnLogTime(`主库自动同步标记任务 running 失败 task_id=%d err=%s`, taskID, markErr.Error())
		}
		h.notifyStatusChange()
	}

	// 执行同步。
	syncErr := h.syncNow()

	// 标记终态。
	if taskID > 0 {
		if syncErr != nil {
			if markErr := h.markAsyncTaskFailed(taskID, syncErr.Error()); markErr != nil {
				gstool.FmtPrintlnLogTime(`主库自动同步标记任务 failed 失败 task_id=%d err=%s`, taskID, markErr.Error())
			}
		} else {
			if markErr := h.markAsyncTaskConfirmed(taskID); markErr != nil {
				gstool.FmtPrintlnLogTime(`主库自动同步标记任务 confirmed 失败 task_id=%d err=%s`, taskID, markErr.Error())
			}
		}
		h.notifyStatusChange()
	}
	h.clearPendingTaskID()
	return syncErr
}

// syncNow 执行实际的 git 同步操作（不带异步任务记录）。
func (h *MainDBAutoSync) syncNow() error {
	h.mu.RLock()
	config := h.config
	syncer := h.syncer
	h.mu.RUnlock()

	if config.Dir == `` || config.DBName == `` {
		h.setLastSyncError(ErrMainDBNotConfigured.Error())
		return ErrMainDBNotConfigured
	}
	if !config.IsGitRepo || syncer == nil {
		return nil
	}

	fileName := filepath.Base(filepath.Join(config.Dir, config.DBName))
	gstool.FmtPrintlnLogTime(`主库自动同步开始检查变更 dir=%s file=%s`, config.Dir, fileName)

	hasChanges, err := syncer.HasFileChanges(config.Dir, fileName)
	if err != nil {
		h.setLastSyncError(err.Error())
		gstool.FmtPrintlnLogTime(`主库自动同步检查变更失败 dir=%s file=%s err=%s`, config.Dir, fileName, err.Error())
		return err
	}
	if !hasChanges {
		h.setLastSyncError(``)
		gstool.FmtPrintlnLogTime(`主库自动同步未检测到文件变更，跳过 dir=%s file=%s`, config.Dir, fileName)
		return nil
	}
	if err = syncer.AddFile(config.Dir, fileName); err != nil {
		h.setLastSyncError(err.Error())
		gstool.FmtPrintlnLogTime(`主库自动同步 add 失败 dir=%s file=%s err=%s`, config.Dir, fileName, err.Error())
		return err
	}
	if err = syncer.Commit(config.Dir, fileName, MainDBAutoSyncCommitMessage); err != nil {
		h.setLastSyncError(err.Error())
		gstool.FmtPrintlnLogTime(`主库自动同步 commit 失败 dir=%s file=%s err=%s`, config.Dir, fileName, err.Error())
		return err
	}
	if err = syncer.Push(config.Dir); err != nil {
		h.setLastSyncError(err.Error())
		gstool.FmtPrintlnLogTime(`主库自动同步 push 失败 dir=%s file=%s err=%s`, config.Dir, fileName, err.Error())
		return err
	}

	h.mu.Lock()
	h.lastSyncTime = time.Now().Unix()
	h.lastSyncErr = ``
	h.mu.Unlock()

	gstool.FmtPrintlnLogTime(`主库自动同步成功 dir=%s file=%s`, config.Dir, fileName)
	return nil
}

// tryLockSync 尝试获取同步互斥锁，如果已有同步在执行则返回 false。 // Try to acquire the sync mutex; return false if a sync is already running.
func (h *MainDBAutoSync) tryLockSync() bool {
	ok := h.syncMu.TryLock()
	gstool.FmtPrintlnLogTime(`主库自动同步 tryLockSync result=%v`, ok)
	return ok
}

// unlockSync 释放同步互斥锁。 // Release the sync mutex.
func (h *MainDBAutoSync) unlockSync() {
	h.syncMu.Unlock()
}

func (h *MainDBAutoSync) setLastSyncError(message string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastSyncErr = message
}

// notifyStatusChange 安全调用 OnStatusChange 回调。
func (h *MainDBAutoSync) notifyStatusChange() {
	h.mu.RLock()
	cb := h.OnStatusChange
	h.mu.RUnlock()
	if cb != nil {
		cb()
	}
}

// castToInt 将 any 转为 int，用于从 map[string]any 中提取 task id。
func castToInt(v any) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	default:
		return 0
	}
}

// formatFSNotifyOp 格式化 fsnotify 事件类型，便于定位主库监听触发来源。 // Format fsnotify operations for main db watcher logs.
func formatFSNotifyOp(op fsnotify.Op) string {
	opList := make([]string, 0, 5)
	if op&fsnotify.Create != 0 {
		opList = append(opList, `CREATE`)
	}
	if op&fsnotify.Write != 0 {
		opList = append(opList, `WRITE`)
	}
	if op&fsnotify.Remove != 0 {
		opList = append(opList, `REMOVE`)
	}
	if op&fsnotify.Rename != 0 {
		opList = append(opList, `RENAME`)
	}
	if op&fsnotify.Chmod != 0 {
		opList = append(opList, `CHMOD`)
	}
	if len(opList) == 0 {
		return op.String()
	}
	return strings.Join(opList, `|`)
}

// ensurePendingTask 确保当前防抖窗口存在一条 pending 异步任务。 // Ensure the current debounce window has one pending async task.
// 使用双重检查锁定防止 watcher goroutine 和 timer goroutine 并发创建重复任务。 // Uses double-checked locking to prevent concurrent duplicate task creation.
func (h *MainDBAutoSync) ensurePendingTask(config MainDBAutoSyncConfig) (int, error) {
	// 快速路径：读锁检查是否已有待处理任务。 // Fast path: check under read lock.
	h.mu.RLock()
	if h.pendingTaskID > 0 {
		taskID := h.pendingTaskID
		h.mu.RUnlock()
		return taskID, nil
	}
	h.mu.RUnlock()

	// 慢速路径：升级为写锁，防止并发创建重复任务。 // Slow path: acquire write lock to prevent concurrent duplicate creation.
	h.mu.Lock()
	// 双重检查：获取写锁后再次确认，避免另一个 goroutine 刚创建完任务。 // Double-check after acquiring write lock.
	if h.pendingTaskID > 0 {
		taskID := h.pendingTaskID
		h.mu.Unlock()
		return taskID, nil
	}

	createTask := h.createAsyncTask
	loadPendingTask := h.loadPendingTask

	if loadPendingTask != nil {
		// 在写锁内执行 DB 查询，确保只有一个 goroutine 能进入此分支。 // DB query inside write lock ensures only one goroutine enters.
		taskID, _, err := loadPendingTask(config)
		if err != nil {
			h.mu.Unlock()
			return 0, err
		}
		if taskID > 0 {
			h.pendingTaskID = taskID
			h.mu.Unlock()
			h.notifyStatusChange()
			return taskID, nil
		}
	}

	if createTask == nil {
		h.mu.Unlock()
		return 0, nil
	}
	// 在写锁内创建任务，保证同一时刻只创建一条 pending 记录。 // Create task inside write lock to guarantee only one pending record.
	taskID, err := createTask(config)
	if err != nil {
		h.mu.Unlock()
		return 0, err
	}
	if taskID <= 0 {
		h.mu.Unlock()
		return 0, nil
	}

	h.pendingTaskID = taskID
	h.pendingTaskRecovered = false
	h.mu.Unlock()
	h.notifyStatusChange()
	return taskID, nil
}

// clearPendingTaskID 清理当前防抖窗口对应的待处理任务标记。 // Clear the pending async task id for the current debounce window.
func (h *MainDBAutoSync) clearPendingTaskID() {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.pendingTaskID = 0
	h.pendingTaskRecovered = false
}

// defaultCreateAsyncTask 创建主库自动同步待处理任务。 // Create the pending async task for main db auto sync.
func (h *MainDBAutoSync) defaultCreateAsyncTask(config MainDBAutoSyncConfig) (int, error) {
	taskTitle := `主库自动同步 ` + time.Now().Format(`2006-01-02 15:04:05`)
	if DbLog == nil || DbLog.Client == nil {
		gstool.FmtPrintlnLogTime(`主库自动同步日志库未初始化，跳过任务记录`)
		return 0, nil
	}
	scheduledAt := time.Now().Add(time.Duration(config.AutoSyncMinutes) * time.Minute).Unix()
	taskInfo, err := DbLog.AsyncTaskCreate(mainDBAutoSyncTaskType, taskTitle, ``, h.buildTaskPayload(config, scheduledAt))
	if err != nil {
		return 0, err
	}
	if id, ok := taskInfo[`id`]; ok {
		return castToInt(id), nil
	}
	return 0, nil
}

// defaultMarkAsyncTaskRunning 将主库自动同步任务标记为 running。 // Mark the main db auto sync task as running.
func (h *MainDBAutoSync) defaultMarkAsyncTaskRunning(taskID int) error {
	if DbLog == nil || DbLog.Client == nil || taskID <= 0 {
		return nil
	}
	return DbLog.AsyncTaskMarkRunning(taskID)
}

// defaultMarkAsyncTaskFailed 将主库自动同步任务标记为 failed。 // Mark the main db auto sync task as failed.
func (h *MainDBAutoSync) defaultMarkAsyncTaskFailed(taskID int, errMsg string) error {
	if DbLog == nil || DbLog.Client == nil || taskID <= 0 {
		return nil
	}
	return DbLog.AsyncTaskMarkFailed(taskID, errMsg)
}

// defaultMarkAsyncTaskConfirmed 将主库自动同步任务标记为 confirmed。 // Mark the main db auto sync task as confirmed.
func (h *MainDBAutoSync) defaultMarkAsyncTaskConfirmed(taskID int) error {
	if DbLog == nil || DbLog.Client == nil || taskID <= 0 {
		return nil
	}
	return DbLog.AsyncTaskMarkFinal(taskID, AsyncTaskStatusConfirmed)
}

// defaultUpdateAsyncTaskRequestPayload 更新待处理任务的请求参数。 // Update request payload for an existing pending task.
func (h *MainDBAutoSync) defaultUpdateAsyncTaskRequestPayload(taskID int, payload string) error {
	if DbLog == nil || DbLog.Client == nil || taskID <= 0 {
		return nil
	}
	return DbLog.AsyncTaskUpdateRequestPayload(taskID, payload)
}

// defaultLoadPendingTask 读取最新的主库待处理同步任务用于启动恢复。 // Load the latest pending main db sync task for startup recovery.
func (h *MainDBAutoSync) defaultLoadPendingTask(config MainDBAutoSyncConfig) (int, string, error) {
	if DbLog == nil || DbLog.Client == nil {
		return 0, ``, nil
	}
	taskInfo, err := DbLog.AsyncTaskLatestPendingByType(mainDBAutoSyncTaskType)
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

// restorePendingTask 从持久化任务中恢复待处理同步计时器。 // Restore pending sync timer from persisted task metadata.
func (h *MainDBAutoSync) restorePendingTask() {
	h.mu.RLock()
	config := h.config
	loadPendingTask := h.loadPendingTask
	h.mu.RUnlock()

	if loadPendingTask == nil || config.Dir == `` || config.DBName == `` || !config.IsGitRepo {
		return
	}
	taskID, payload, err := loadPendingTask(config)
	if err != nil {
		gstool.FmtPrintlnLogTime(`主库自动同步恢复待处理任务失败 dir=%s file=%s err=%s`, config.Dir, config.DBName, err.Error())
		return
	}
	if taskID <= 0 || strings.TrimSpace(payload) == `` {
		return
	}
	scheduledAt, parseErr := h.extractScheduledAt(payload)
	if parseErr != nil {
		gstool.FmtPrintlnLogTime(`主库自动同步解析恢复任务排期失败 task_id=%d err=%s`, taskID, parseErr.Error())
		return
	}
	h.mu.Lock()
	h.pendingTaskID = taskID
	h.pendingTaskRecovered = true
	h.mu.Unlock()
	if err := h.persistPendingTaskSchedule(config, taskID, scheduledAt); err != nil {
		gstool.FmtPrintlnLogTime(`主库自动同步回写恢复任务描述失败 task_id=%d err=%s`, taskID, err.Error())
	}
	gstool.FmtPrintlnLogTime(`主库自动同步恢复待处理任务 task_id=%d scheduled_at=%d`, taskID, scheduledAt)
	h.notifyStatusChange()
	h.scheduleRecoveredTask(config, scheduledAt)
}

// scheduleRecoveredTask 根据恢复的计划时间重新挂载 timer，过期任务立即执行。 // Re-schedule recovered pending task and run immediately when overdue.
func (h *MainDBAutoSync) scheduleRecoveredTask(config MainDBAutoSyncConfig, scheduledAt int64) {
	delay := time.Until(time.Unix(scheduledAt, 0))
	if delay < 0 {
		delay = 0
	}
	h.mu.Lock()
	if h.debounceTimer != nil {
		h.debounceTimer.Stop()
	}
	h.debounceTimer = h.afterFunc(delay, func() {
		_ = h.syncWithAsyncTask()
	})
	h.mu.Unlock()
	gstool.FmtPrintlnLogTime(`主库自动同步恢复排期 task_id=%d delay_ms=%d`, h.pendingTaskID, delay.Milliseconds())
}

// persistPendingTaskSchedule 回写待处理任务的调度参数，保证重启后可恢复。 // Persist pending task schedule metadata for restart recovery.
func (h *MainDBAutoSync) persistPendingTaskSchedule(config MainDBAutoSyncConfig, taskID int, scheduledAt int64) error {
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

// buildTaskPayload 构造通用可恢复调度元数据。 // Build a generic resumable payload for main db auto sync tasks.
func (h *MainDBAutoSync) buildTaskPayload(config MainDBAutoSyncConfig, scheduledAt int64) string {
	return h.buildTaskPayloadWithRecovery(config, scheduledAt, false)
}

// buildTaskPayloadWithRecovery 构造带恢复标记的主库同步 payload。 // Build main db sync payload with optional recovery flag.
func (h *MainDBAutoSync) buildTaskPayloadWithRecovery(config MainDBAutoSyncConfig, scheduledAt int64, restored bool) string {
	mainDBPath := filepath.Join(config.Dir, config.DBName)
	scheduledAtDesc := formatMainDBAutoSyncScheduleTime(scheduledAt)
	taskDescription := `已检测到主库变更，预计在 ` + scheduledAtDesc + ` 自动同步`
	if restored {
		taskDescription = `已从未完成任务恢复，预计在 ` + scheduledAtDesc + ` 自动同步`
	}
	payload := map[string]any{
		`task_kind`:        mainDBAutoSyncTaskKind,
		`task_target`:      mainDBAutoSyncTaskTarget,
		`task_description`: taskDescription,
		`task_params`: map[string]any{
			`dir`:         config.Dir,
			`target_name`: config.DBName,
			`target_type`: `sqlite_file`,
			`watched_files`: []string{
				mainDBPath,
				mainDBPath + `-wal`,
				mainDBPath + `-shm`,
			},
		},
		`schedule`: map[string]any{
			`trigger`:              `fsnotify`,
			`scheduled_at`:         scheduledAt,
			`scheduled_at_desc`:    scheduledAtDesc,
			`delay_minutes`:        config.AutoSyncMinutes,
			`created_from_runtime`: `main_db_auto_sync`,
		},
		`resume`: map[string]any{
			`resume_key`:      h.buildResumeKey(config),
			`resume_strategy`: mainDBAutoSyncResumeStrategy,
			`restored`:        restored,
		},
	}
	return gstool.JsonEncode(payload)
}

// buildResumeKey 构造当前主库同步任务的恢复键。 // Build the resume key for the current main db sync task.
func (h *MainDBAutoSync) buildResumeKey(config MainDBAutoSyncConfig) string {
	return mainDBAutoSyncTaskKind + `:` + mainDBAutoSyncTaskTarget + `:` + filepath.Join(config.Dir, config.DBName)
}

// parseTaskPayload 解析主库同步任务 payload。 // Decode the main db sync payload.
func (h *MainDBAutoSync) parseTaskPayload(payload string) (map[string]any, error) {
	payloadMap := make(map[string]any)
	if err := gstool.JsonDecode(payload, &payloadMap); err != nil {
		return nil, err
	}
	return payloadMap, nil
}

// extractScheduledAt 从 payload 中读取计划执行时间。 // Read scheduled_at from the persisted task payload.
func (h *MainDBAutoSync) extractScheduledAt(payload string) (int64, error) {
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

// isPayloadMatchCurrentConfig 校验恢复任务是否仍然匹配当前主库配置。 // Check whether a persisted pending task still matches current main db config.
func (h *MainDBAutoSync) isPayloadMatchCurrentConfig(config MainDBAutoSyncConfig, payloadMap map[string]any) bool {
	if payloadMap[`task_kind`] != mainDBAutoSyncTaskKind || payloadMap[`task_target`] != mainDBAutoSyncTaskTarget {
		return false
	}
	resumeMap, _ := payloadMap[`resume`].(map[string]any)
	if resumeMap == nil {
		return false
	}
	return castToString(resumeMap[`resume_key`]) == h.buildResumeKey(config)
}

func castToString(v any) string {
	switch val := v.(type) {
	case string:
		return val
	default:
		return ``
	}
}

func castToInt64(v any) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	default:
		return 0
	}
}

func formatMainDBAutoSyncScheduleTime(ts int64) string {
	if ts <= 0 {
		return `-`
	}
	return time.Unix(ts, 0).Format(`2006-01-02 15:04:05`)
}
