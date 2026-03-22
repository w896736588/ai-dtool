package controller

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

type fakeManagedProcessRunner struct {
	findResult  *managedProcessSnapshot
	startResult *managedProcessSnapshot
	findCalls   int
	startCalls  []managedProcessConfig
	killCalls   []int32
}

func (f *fakeManagedProcessRunner) Find(config managedProcessConfig) (*managedProcessSnapshot, error) {
	f.findCalls++
	return f.findResult, nil
}

func (f *fakeManagedProcessRunner) Start(config managedProcessConfig, logFile string) (*managedProcessSnapshot, error) {
	f.startCalls = append(f.startCalls, config)
	if f.startResult != nil {
		result := *f.startResult
		result.LogFile = logFile
		return &result, nil
	}
	return &managedProcessSnapshot{
		PID:     4321,
		LogFile: logFile,
	}, nil
}

func (f *fakeManagedProcessRunner) Kill(pid int32) error {
	f.killCalls = append(f.killCalls, pid)
	return nil
}

func TestToolManagedProcessNormalizeConfig(t *testing.T) {
	now := time.Date(2026, 3, 22, 9, 0, 0, 0, time.Local)

	cfg, err := normalizeManagedProcessConfig(map[string]any{
		`name`:         `CC Connect`,
		`command_line`: `cc-connect --config "C:\Users\94804\.cc-connect\config.toml"`,
		`workdir`:      `C:\work\frog\dev_tool_master`,
	}, now)
	if err != nil {
		t.Fatalf("normalizeManagedProcessConfig error = %v", err)
	}
	if cfg.Key != `cc-connect` {
		t.Fatalf("key = %q, want %q", cfg.Key, `cc-connect`)
	}
	if cfg.Name != `CC Connect` {
		t.Fatalf("name = %q, want %q", cfg.Name, `CC Connect`)
	}
	if cfg.Executable != `cc-connect` {
		t.Fatalf("executable = %q, want %q", cfg.Executable, `cc-connect`)
	}
	if len(cfg.Args) != 2 || cfg.Args[0] != `--config` || cfg.Args[1] != `C:\Users\94804\.cc-connect\config.toml` {
		t.Fatalf("args = %#v", cfg.Args)
	}

	if _, err = normalizeManagedProcessConfig(map[string]any{
		`key`:          `broken`,
		`command_line`: ` `,
	}, now); err == nil {
		t.Fatal("expected empty command_line to return error")
	}
}

func TestToolManagedProcessEnsureRunningDoesNotDuplicateStart(t *testing.T) {
	runner := &fakeManagedProcessRunner{
		findResult: &managedProcessSnapshot{PID: 8899},
	}
	manager := newManagedProcessManager(`C:\work\frog\dev_tool_master\logs`, runner)

	status, err := manager.EnsureRunning(map[string]any{
		`key`:          `cc-connect`,
		`command_line`: `cc-connect --config C:\Users\94804\.cc-connect\config.toml`,
	}, time.Date(2026, 3, 22, 10, 0, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("EnsureRunning error = %v", err)
	}
	if !status.Running {
		t.Fatal("expected running status")
	}
	if status.PID != 8899 {
		t.Fatalf("pid = %d, want 8899", status.PID)
	}
	if status.LogFile != `` {
		t.Fatalf("log file = %q, want empty for external process", status.LogFile)
	}
	if status.IsManaged {
		t.Fatal("expected external process to not be managed")
	}
	if len(runner.startCalls) != 0 {
		t.Fatalf("startCalls = %d, want 0", len(runner.startCalls))
	}
}

func TestToolManagedProcessRestartKillsThenStarts(t *testing.T) {
	runner := &fakeManagedProcessRunner{
		startResult: &managedProcessSnapshot{PID: 9900},
	}
	manager := newManagedProcessManager(`C:\work\frog\dev_tool_master\logs`, runner)
	manager.processMap[`cc-connect`] = &managedProcessEntry{
		Config: managedProcessConfig{
			Key:         `cc-connect`,
			Name:        `cc-connect`,
			CommandLine: `cc-connect --config C:\Users\94804\.cc-connect\config.toml`,
			Executable:  `cc-connect`,
			Args:        []string{`--config`, `C:\Users\94804\.cc-connect\config.toml`},
		},
		Process: &managedProcessSnapshot{
			PID:     8800,
			LogFile: `C:\work\frog\dev_tool_master\logs\cc-connect-2026-03-21.log`,
		},
	}

	status, err := manager.Restart(map[string]any{
		`key`:          `cc-connect`,
		`name`:         `cc-connect`,
		`command_line`: `cc-connect --config C:\Users\94804\.cc-connect\config.toml --verbose`,
	}, time.Date(2026, 3, 22, 11, 30, 0, 0, time.Local))
	if err != nil {
		t.Fatalf("Restart error = %v", err)
	}
	if len(runner.killCalls) != 1 || runner.killCalls[0] != 8800 {
		t.Fatalf("killCalls = %#v", runner.killCalls)
	}
	if len(runner.startCalls) != 1 {
		t.Fatalf("startCalls = %d, want 1", len(runner.startCalls))
	}
	if status.PID != 9900 {
		t.Fatalf("pid = %d, want 9900", status.PID)
	}
	wantLogFile := filepath.Clean(`C:\work\frog\dev_tool_master\logs\cc-connect-2026-03-22.log`)
	if filepath.Clean(status.LogFile) != wantLogFile {
		t.Fatalf("log file = %q, want %q", status.LogFile, wantLogFile)
	}
}

func TestToolManagedProcessReadLogTail(t *testing.T) {
	dir := t.TempDir()
	logFile := filepath.Join(dir, `cc-connect-2026-03-22.log`)
	content := "line-1\nline-2\nline-3\n"
	if err := os.WriteFile(logFile, []byte(content), 0o644); err != nil {
		t.Fatalf("WriteFile error = %v", err)
	}

	tail, err := readManagedProcessLogTail(logFile, 8)
	if err != nil {
		t.Fatalf("readManagedProcessLogTail error = %v", err)
	}
	if tail != "line-3\n" {
		t.Fatalf("tail = %q, want %q", tail, "line-3\n")
	}

	empty, err := readManagedProcessLogTail(filepath.Join(dir, `missing.log`), 32)
	if err != nil {
		t.Fatalf("read missing log error = %v", err)
	}
	if empty != `` {
		t.Fatalf("missing tail = %q, want empty", empty)
	}
}

func TestBuildManagedProcessStatusKeepsExternalProcessLogEmpty(t *testing.T) {
	// 外部进程不受当前托管器控制 / External processes should not expose fabricated log files.
	status := buildManagedProcessStatus(managedProcessConfig{
		Key:         `cc-connect`,
		Name:        `cc-connect`,
		CommandLine: `cc-connect --config C:\Users\94804\.cc-connect\config.toml`,
	}, &managedProcessSnapshot{
		PID:        7788,
		IsManaged:  false,
		LogFile:    filepath.Clean(`C:\work\frog\dev_tool_master\logs\cc-connect-2026-03-22.log`),
		StatusText: `运行中（外部进程）`,
	})

	if status.LogFile != `` {
		t.Fatalf("log file = %q, want empty for external process status", status.LogFile)
	}
}
