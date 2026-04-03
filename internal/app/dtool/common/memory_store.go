package common

import (
	"errors"
	"path/filepath"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

const MemorySyncCommitMessage = `chore: sync memory db`
const DefaultMemoryAutoPushDelayMinutes = 1

var ErrMemoryNotConfigured = errors.New(`请先在配置文件中配置记忆库目录和数据库名`)

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
	db           *CSqlite
	timer        stoppableTimer
	dirty        bool
	nextPushTime int64
	lastPushTime int64
	lastPushErr  string
	afterFunc    func(time.Duration, func()) stoppableTimer
	gitSyncer    memoryGitSyncer
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		afterFunc: func(duration time.Duration, callback func()) stoppableTimer {
			return &timeTimer{timer: time.AfterFunc(duration, callback)}
		},
	}
}

func (h *MemoryStore) SetGitSyncer(syncer memoryGitSyncer) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.gitSyncer = syncer
}

func (h *MemoryStore) Configure(config MemoryConfig, db *CSqlite) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config = config
	h.db = db
	h.dirty = false
	h.nextPushTime = 0
	h.lastPushTime = 0
	h.lastPushErr = ``
	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}
}

func (h *MemoryStore) Reset() {
	h.Configure(MemoryConfig{}, nil)
}

func (h *MemoryStore) Config() MemoryConfig {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.config
}

func (h *MemoryStore) DB() *CSqlite {
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
	return h.config.Dir != `` && h.config.DBName != `` && h.db != nil
}

func (h *MemoryStore) EnsureConfigured() error {
	if h.IsConfigured() {
		return nil
	}
	return ErrMemoryNotConfigured
}

func (h *MemoryStore) ScheduleSync() {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.config.Dir == `` || h.config.DBName == `` {
		return
	}
	h.dirty = true
	if h.config.AutoPushDelayMinutes <= 0 {
		if h.timer != nil {
			h.timer.Stop()
			h.timer = nil
		}
		h.nextPushTime = 0
		return
	}
	if h.timer != nil {
		h.timer.Stop()
	}
	h.nextPushTime = time.Now().Add(time.Duration(h.config.AutoPushDelayMinutes) * time.Minute).Unix()
	h.timer = h.afterFunc(time.Duration(h.config.AutoPushDelayMinutes)*time.Minute, func() {
		_ = h.SyncNow()
	})
}

func (h *MemoryStore) SyncNow() error {
	h.mu.Lock()
	config := h.config
	syncer := h.gitSyncer
	dirty := h.dirty
	if h.timer != nil {
		h.timer.Stop()
		h.timer = nil
	}
	h.nextPushTime = 0
	h.mu.Unlock()

	if config.Dir == `` || config.DBName == `` {
		h.setLastPushError(ErrMemoryNotConfigured.Error())
		gstool.FmtPrintlnLogTime(`记忆库同步失败：配置不完整 dir=%s db=%s`, config.Dir, config.DBName)
		return ErrMemoryNotConfigured
	}
	if !config.IsGitRepo {
		// 未开启 git 模式时直接清理脏标记。 // Clear dirty state directly when git sync is disabled.
		h.mu.Lock()
		h.dirty = false
		h.lastPushErr = ``
		h.mu.Unlock()
		gstool.FmtPrintlnLogTime(`记忆库未启用 git 仓库同步，跳过 push dir=%s file=%s`, config.Dir, config.DBName)
		return nil
	}
	if !dirty {
		// 没有脏数据时不需要触发 git push。 // Skip git push when there are no pending changes.
		gstool.FmtPrintlnLogTime(`记忆库当前没有待同步变更，跳过 push dir=%s file=%s`, config.Dir, filepath.Base(config.DBPath))
		return nil
	}
	if syncer == nil {
		err := errors.New(`memory git syncer not set`)
		h.setLastPushError(err.Error())
		gstool.FmtPrintlnLogTime(`记忆库同步失败：git syncer 未设置 dir=%s file=%s`, config.Dir, filepath.Base(config.DBPath))
		return err
	}

	gstool.FmtPrintlnLogTime(`记忆库开始检查变更并执行 push dir=%s file=%s`, config.Dir, filepath.Base(config.DBPath))
	hasChanges, err := syncer.HasFileChanges(config.Dir, filepath.Base(config.DBPath))
	if err != nil {
		h.setLastPushError(err.Error())
		gstool.FmtPrintlnLogTime(`记忆库检查变更失败 dir=%s file=%s err=%s`, config.Dir, filepath.Base(config.DBPath), err.Error())
		return err
	}
	if !hasChanges {
		h.mu.Lock()
		h.dirty = false
		h.lastPushErr = ``
		h.mu.Unlock()
		gstool.FmtPrintlnLogTime(`记忆库未检测到文件变更，跳过 push dir=%s file=%s`, config.Dir, filepath.Base(config.DBPath))
		return nil
	}
	if err = syncer.AddFile(config.Dir, filepath.Base(config.DBPath)); err != nil {
		h.setLastPushError(err.Error())
		gstool.FmtPrintlnLogTime(`记忆库 add 失败 dir=%s file=%s err=%s`, config.Dir, filepath.Base(config.DBPath), err.Error())
		return err
	}
	if err = syncer.Commit(config.Dir, filepath.Base(config.DBPath), MemorySyncCommitMessage); err != nil {
		h.setLastPushError(err.Error())
		gstool.FmtPrintlnLogTime(`记忆库 commit 失败 dir=%s file=%s err=%s`, config.Dir, filepath.Base(config.DBPath), err.Error())
		return err
	}
	if err = syncer.Push(config.Dir); err != nil {
		h.setLastPushError(err.Error())
		gstool.FmtPrintlnLogTime(`记忆库 push 失败 dir=%s file=%s err=%s`, config.Dir, filepath.Base(config.DBPath), err.Error())
		return err
	}

	h.mu.Lock()
	h.dirty = false
	h.lastPushTime = time.Now().Unix()
	h.lastPushErr = ``
	h.mu.Unlock()
	gstool.FmtPrintlnLogTime(`记忆库 push 成功 dir=%s file=%s`, config.Dir, filepath.Base(config.DBPath))
	return nil
}

func (h *MemoryStore) setLastPushError(message string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastPushErr = message
}
