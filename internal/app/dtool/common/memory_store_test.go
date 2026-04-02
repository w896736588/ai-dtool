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
		Dir:                  `C:/memory`,
		DBName:               `memory.db`,
		DBPath:               `C:/memory/memory.db`,
		IsGitRepo:            true,
		AutoPushDelayMinutes: 3,
	}

	firstTimer := &fakeTimer{}
	secondTimer := &fakeTimer{}
	created := make([]*fakeTimer, 0, 2)
	var durations []time.Duration
	store.afterFunc = func(duration time.Duration, _ func()) stoppableTimer {
		durations = append(durations, duration)
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
	if len(durations) != 2 {
		t.Fatalf("scheduled durations = %d, want 2", len(durations))
	}
	for _, duration := range durations {
		if duration != 3*time.Minute {
			t.Fatalf("scheduled duration = %v, want %v", duration, 3*time.Minute)
		}
	}
	if firstTimer.stopCount != 1 {
		t.Fatalf("first timer stop count = %d, want 1", firstTimer.stopCount)
	}
	if secondTimer.stopCount != 0 {
		t.Fatalf("second timer stop count = %d, want 0", secondTimer.stopCount)
	}
	if store.NextPushTime() <= time.Now().Unix() {
		t.Fatalf("NextPushTime() = %d, want a future unix timestamp", store.NextPushTime())
	}
}

func TestMemoryStoreScheduleSyncDisabledWhenDelayNonPositive(t *testing.T) {
	t.Parallel()

	store := NewMemoryStore()
	store.config = MemoryConfig{
		Dir:                  `C:/memory`,
		DBName:               `memory.db`,
		DBPath:               `C:/memory/memory.db`,
		IsGitRepo:            true,
		AutoPushDelayMinutes: 0,
	}
	scheduled := false
	store.afterFunc = func(_ time.Duration, _ func()) stoppableTimer {
		scheduled = true
		return &fakeTimer{}
	}

	store.ScheduleSync()

	if scheduled {
		t.Fatalf("ScheduleSync() scheduled timer when auto push delay disabled")
	}
	if !store.dirty {
		t.Fatalf("dirty = false, want true")
	}
	if store.NextPushTime() != 0 {
		t.Fatalf("NextPushTime() = %d, want 0 when auto push disabled", store.NextPushTime())
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
	if store.NextPushTime() != 0 {
		t.Fatalf("NextPushTime() = %d, want 0 after SyncNow()", store.NextPushTime())
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
