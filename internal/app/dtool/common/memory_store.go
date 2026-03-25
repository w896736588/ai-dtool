package common

import (
	"errors"
	"path/filepath"
	"sync"
	"time"
)

const MemorySyncCommitMessage = `chore: sync memory db`

var ErrMemoryNotConfigured = errors.New(`请先到设置页面配置记忆目录和数据库名`)

type MemoryConfig struct {
	Dir       string `json:"memory_dir"`
	DBName    string `json:"memory_db_name"`
	DBPath    string `json:"memory_db_path"`
	IsGitRepo bool   `json:"is_git_repo"`
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
	lastPushTime int64
	lastPushErr  string
	afterFunc    func(time.Duration, func()) stoppableTimer
	gitSyncer    memoryGitSyncer
}

var MemoryRuntime = NewMemoryStore()

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
	if h.timer != nil {
		h.timer.Stop()
	}
	h.timer = h.afterFunc(time.Minute, func() {
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
	h.mu.Unlock()

	if config.Dir == `` || config.DBName == `` {
		h.setLastPushError(ErrMemoryNotConfigured.Error())
		return ErrMemoryNotConfigured
	}
	if !config.IsGitRepo {
		h.mu.Lock()
		h.dirty = false
		h.lastPushErr = ``
		h.mu.Unlock()
		return nil
	}
	if !dirty {
		return nil
	}
	if syncer == nil {
		err := errors.New(`memory git syncer not set`)
		h.setLastPushError(err.Error())
		return err
	}

	hasChanges, err := syncer.HasFileChanges(config.Dir, filepath.Base(config.DBPath))
	if err != nil {
		h.setLastPushError(err.Error())
		return err
	}
	if !hasChanges {
		h.mu.Lock()
		h.dirty = false
		h.lastPushErr = ``
		h.mu.Unlock()
		return nil
	}
	if err = syncer.AddFile(config.Dir, filepath.Base(config.DBPath)); err != nil {
		h.setLastPushError(err.Error())
		return err
	}
	if err = syncer.Commit(config.Dir, filepath.Base(config.DBPath), MemorySyncCommitMessage); err != nil {
		h.setLastPushError(err.Error())
		return err
	}
	if err = syncer.Push(config.Dir); err != nil {
		h.setLastPushError(err.Error())
		return err
	}

	h.mu.Lock()
	h.dirty = false
	h.lastPushTime = time.Now().Unix()
	h.lastPushErr = ``
	h.mu.Unlock()
	return nil
}

func (h *MemoryStore) setLastPushError(message string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.lastPushErr = message
}
