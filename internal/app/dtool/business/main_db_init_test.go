package business

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"errors"
	"testing"
)

type fakeMainDBGitSyncer struct {
	isRepo        bool
	hasChanges    bool
	pullCount     int
	pushCount     int
	commitCount   int
	addCount      int
	isGitRepoErr  error
	pullErr       error
	hasChangesErr error
	addErr        error
	commitErr     error
	pushErr       error
}

func (h *fakeMainDBGitSyncer) IsGitRepo(string) (bool, error) {
	return h.isRepo, h.isGitRepoErr
}

func (h *fakeMainDBGitSyncer) Pull(string) error {
	h.pullCount++
	return h.pullErr
}

func (h *fakeMainDBGitSyncer) HasFileChanges(string, string) (bool, error) {
	return h.hasChanges, h.hasChangesErr
}

func (h *fakeMainDBGitSyncer) AddFile(string, string) error {
	h.addCount++
	return h.addErr
}

func (h *fakeMainDBGitSyncer) Commit(string, string, string) error {
	h.commitCount++
	return h.commitErr
}

func (h *fakeMainDBGitSyncer) Push(string) error {
	h.pushCount++
	return h.pushErr
}

func TestReadMainDBConfig(t *testing.T) {
	oldEnv := component.EnvClient
	t.Cleanup(func() {
		component.EnvClient = oldEnv
	})

	component.EnvClient = &define.Env{
		ConfigBase: &define.Base{
			DbIsGitRepo: true,
		},
		DbConfig: &define.DbConfig{
			DbPath: `D:\repo\main`,
			DbName: `frog.db`,
		},
	}

	got := ReadMainDBConfig()
	if got.Dir != `D:\repo\main` || got.DBName != `frog.db` || got.DBPath != `D:\repo\main\frog.db` || !got.GitRepoEnabled {
		t.Fatalf("ReadMainDBConfig() = %+v", got)
	}
}

func TestPrepareMainDBStorePullsOnlyWhenGitEnabled(t *testing.T) {
	oldEnv := component.EnvClient
	oldPrepared := preparedMainDBStore
	oldFactory := newMainDBGitSyncer
	t.Cleanup(func() {
		component.EnvClient = oldEnv
		preparedMainDBStore = oldPrepared
		newMainDBGitSyncer = oldFactory
	})

	fakeGit := &fakeMainDBGitSyncer{isRepo: true}
	newMainDBGitSyncer = func() mainDBGitSyncer { return fakeGit }
	component.EnvClient = &define.Env{
		ConfigBase: &define.Base{
			DbIsGitRepo: true,
		},
		DbConfig: &define.DbConfig{
			DbPath: t.TempDir(),
			DbName: `frog.db`,
		},
	}

	if err := PrepareMainDBStore(); err != nil {
		t.Fatalf("PrepareMainDBStore() error = %v", err)
	}
	if fakeGit.pullCount != 1 {
		t.Fatalf("pullCount = %d, want 1", fakeGit.pullCount)
	}
	if preparedMainDBStore == nil || !preparedMainDBStore.Config.IsGitRepo {
		t.Fatalf("preparedMainDBStore = %+v, want git repo enabled", preparedMainDBStore)
	}
}

func TestSyncMainDBStoreOnShutdownSkipsWhenNoChanges(t *testing.T) {
	oldPrepared := preparedMainDBStore
	t.Cleanup(func() {
		preparedMainDBStore = oldPrepared
	})

	preparedMainDBStore = &preparedMainDBBootstrap{
		Config: MainDBConfig{
			Dir:       `C:\repo`,
			DBName:    `frog.db`,
			DBPath:    `C:\repo\frog.db`,
			IsGitRepo: true,
		},
		Git: &fakeMainDBGitSyncer{hasChanges: false},
	}

	if err := SyncMainDBStoreOnShutdown(); err != nil {
		t.Fatalf("SyncMainDBStoreOnShutdown() error = %v", err)
	}
}

func TestSyncMainDBStoreOnShutdownPushesChangedFile(t *testing.T) {
	oldPrepared := preparedMainDBStore
	t.Cleanup(func() {
		preparedMainDBStore = oldPrepared
	})

	fakeGit := &fakeMainDBGitSyncer{hasChanges: true}
	preparedMainDBStore = &preparedMainDBBootstrap{
		Config: MainDBConfig{
			Dir:       `C:\repo`,
			DBName:    `frog.db`,
			DBPath:    `C:\repo\frog.db`,
			IsGitRepo: true,
		},
		Git: fakeGit,
	}

	if err := SyncMainDBStoreOnShutdown(); err != nil {
		t.Fatalf("SyncMainDBStoreOnShutdown() error = %v", err)
	}
	if fakeGit.addCount != 1 || fakeGit.commitCount != 1 || fakeGit.pushCount != 1 {
		t.Fatalf("sync counts = add:%d commit:%d push:%d, want all 1", fakeGit.addCount, fakeGit.commitCount, fakeGit.pushCount)
	}
}

func TestSyncMainDBStoreOnShutdownReturnsPushError(t *testing.T) {
	oldPrepared := preparedMainDBStore
	t.Cleanup(func() {
		preparedMainDBStore = oldPrepared
	})

	preparedMainDBStore = &preparedMainDBBootstrap{
		Config: MainDBConfig{
			Dir:       `C:\repo`,
			DBName:    `frog.db`,
			DBPath:    `C:\repo\frog.db`,
			IsGitRepo: true,
		},
		Git: &fakeMainDBGitSyncer{
			hasChanges: true,
			pushErr:    errors.New(`push failed`),
		},
	}

	if err := SyncMainDBStoreOnShutdown(); err == nil {
		t.Fatalf("SyncMainDBStoreOnShutdown() error = nil, want error")
	}
}
