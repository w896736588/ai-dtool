package common

import (
	"errors"
	"sync"
	"testing"
	"time"
)

type fakeTimer struct {
	stopCount int
	mu        sync.Mutex
}

func (h *fakeTimer) Stop() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.stopCount++
	return true
}

type fakeGitSyncer struct {
	hasChanges bool
	syncCount  int
	pushErr    error
}

func (h *fakeGitSyncer) HasFileChanges(string, string) (bool, error) {
	return h.hasChanges, nil
}

func (h *fakeGitSyncer) AddFile(string, string) error {
	return nil
}

func (h *fakeGitSyncer) Commit(string, string, string) error {
	return nil
}

func (h *fakeGitSyncer) Push(string) error {
	if h.pushErr != nil {
		return h.pushErr
	}
	h.syncCount++
	return nil
}

func TestMemoryStoreScheduleSyncDebounce(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	store.config = MemoryConfig{
		Dir:       `C:/memory`,
		DBName:    `memory.db`,
		DBPath:    `C:/memory/memory.db`,
		IsGitRepo: true,
	}

	firstTimer := &fakeTimer{}
	secondTimer := &fakeTimer{}
	created := make([]*fakeTimer, 0, 2)
	store.afterFunc = func(_ time.Duration, _ func()) stoppableTimer {
		timer := firstTimer
		if len(created) > 0 {
			timer = secondTimer
		}
		created = append(created, timer)
		return timer
	}

	store.ScheduleSync()
	store.ScheduleSync()

	if len(created) != 2 {
		t.Fatalf("created timers = %d, want 2", len(created))
	}
	if firstTimer.stopCount != 1 {
		t.Fatalf("first timer stop count = %d, want 1", firstTimer.stopCount)
	}
	if secondTimer.stopCount != 0 {
		t.Fatalf("second timer stop count = %d, want 0", secondTimer.stopCount)
	}
}

func TestMemoryStoreSyncNowOnlyPushesChangedFile(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	gitSyncer := &fakeGitSyncer{hasChanges: true}
	store.gitSyncer = gitSyncer
	store.config = MemoryConfig{
		Dir:       `C:/memory`,
		DBName:    `memory.db`,
		DBPath:    `C:/memory/memory.db`,
		IsGitRepo: true,
	}
	store.dirty = true

	if err := store.SyncNow(); err != nil {
		t.Fatalf("SyncNow() error = %v", err)
	}
	if gitSyncer.syncCount != 1 {
		t.Fatalf("push count = %d, want 1", gitSyncer.syncCount)
	}
	if store.dirty {
		t.Fatalf("dirty = true, want false")
	}
	if store.LastPushTime() <= 0 {
		t.Fatalf("LastPushTime() = %d, want > 0", store.LastPushTime())
	}
}

func TestMemoryStoreSyncNowSkipsNonGitRepo(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	store.config = MemoryConfig{
		Dir:       `C:/memory`,
		DBName:    `memory.db`,
		DBPath:    `C:/memory/memory.db`,
		IsGitRepo: false,
	}
	store.dirty = true

	if err := store.SyncNow(); err != nil {
		t.Fatalf("SyncNow() error = %v", err)
	}
	if store.dirty {
		t.Fatalf("dirty = true, want false")
	}
}

func TestMemoryStoreSyncNowStoresLastPushError(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	store.gitSyncer = &fakeGitSyncer{
		hasChanges: true,
		pushErr:    errors.New(`push failed`),
	}
	store.config = MemoryConfig{
		Dir:       `C:/memory`,
		DBName:    `memory.db`,
		DBPath:    `C:/memory/memory.db`,
		IsGitRepo: true,
	}
	store.dirty = true

	err := store.SyncNow()
	if err == nil {
		t.Fatalf("SyncNow() error = nil, want error")
	}
	if store.LastPushError() != `push failed` {
		t.Fatalf("LastPushError() = %q, want %q", store.LastPushError(), `push failed`)
	}
}
