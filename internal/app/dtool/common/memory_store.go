package common

import (
	"errors"
	"sync"
	"time"
)

var ErrMemoryNotConfigured = errors.New(`请先在配置文件中配置记忆库目录`)

type stoppableTimer interface {
	Stop() bool
}

type timeTimer struct {
	timer *time.Timer
}

func (h *timeTimer) Stop() bool {
	return h.timer.Stop()
}

type MemoryConfig struct {
	Dir       string `json:"memory_dir"`
	DBName    string `json:"memory_db_name"`
	DBPath    string `json:"memory_db_path"`
	IsGitRepo bool   `json:"is_git_repo"`
}

// MemoryFragmentStore 定义知识片段运行时存储接口。
type MemoryFragmentStore interface {
	MemoryFragmentList(limit, offset int, folderName string) ([]map[string]any, error)
	MemoryFragmentTrashList(limit int) ([]map[string]any, error)
	MemoryFragmentInfo(id any) (map[string]any, error)
	MemoryFragmentSave(id any, title, content string, tags []string, folderName string) (map[string]any, error)
	MemoryFragmentSoftDelete(id any) (int64, error)
	MemoryFragmentRestore(id any) (int64, error)
	MemoryFragmentHardDelete(id any) error
	MemoryFragmentHistoryList(id any) ([]map[string]any, error)
	MemoryFragmentTagList() ([]map[string]any, error)
	MemoryFragmentSearch(mode, query string, selectedTags []string, folderName string, limit int) ([]map[string]any, error)
	MemoryFragmentFolderList() ([]map[string]any, error)
	MemoryFragmentFolderCreate(name, folderName string) (map[string]any, error)
	MemoryFragmentFolderUpdate(folderName, name string) (map[string]any, error)
	MemoryFragmentChangeFolder(id any, folderName string) (map[string]any, error)
	MemoryFragmentBatchInfoByPaths(paths []string) []map[string]any
	SearchFragmentsOr(keywords []string, limit int) ([]map[string]any, error)
	ReadFragmentContent(filePath string) (string, error)
}

type MemoryStore struct {
	mu     sync.RWMutex
	config MemoryConfig
	db     MemoryFragmentStore
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{}
}

func (h *MemoryStore) Configure(config MemoryConfig, db MemoryFragmentStore) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.config = config
	h.db = db
}

func (h *MemoryStore) Reset() {
	h.Configure(MemoryConfig{}, nil)
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
	return 0
}

func (h *MemoryStore) NextPushTime() int64 {
	return 0
}

func (h *MemoryStore) LastPushError() string {
	return ``
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

func (h *MemoryStore) ScheduleSync() {}

func (h *MemoryStore) SyncNow() error {
	return nil
}

func (h *MemoryStore) SyncPendingTaskNow() error {
	return nil
}

func (h *MemoryStore) HasPendingTask() bool {
	return false
}
