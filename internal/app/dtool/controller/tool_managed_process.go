package controller

import (
	"bytes"
	"dev_tool/internal/app/dtool/component"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/shirou/gopsutil/v3/process"
	"github.com/spf13/cast"
)

const defaultManagedProcessTailBytes = 32 * 1024

type managedProcessConfig struct {
	Key         string   `json:"key"`
	Name        string   `json:"name"`
	CommandLine string   `json:"command_line"`
	Workdir     string   `json:"workdir"`
	Executable  string   `json:"-"`
	Args        []string `json:"-"`
}

type managedProcessSnapshot struct {
	PID        int32  `json:"pid"`
	LogFile    string `json:"log_file"`
	StartedAt  int64  `json:"started_at"`
	IsManaged  bool   `json:"is_managed"`
	StatusText string `json:"status_text"`
}

type managedProcessEntry struct {
	Config  managedProcessConfig
	Process *managedProcessSnapshot
}

type managedProcessStatus struct {
	Key         string `json:"key"`
	Name        string `json:"name"`
	CommandLine string `json:"command_line"`
	Workdir     string `json:"workdir"`
	Running     bool   `json:"running"`
	PID         int32  `json:"pid"`
	LogFile     string `json:"log_file"`
	StartedAt   int64  `json:"started_at"`
	IsManaged   bool   `json:"is_managed"`
	StatusText  string `json:"status_text"`
}

type managedProcessRunner interface {
	Find(config managedProcessConfig) (*managedProcessSnapshot, error)
	Start(config managedProcessConfig, logFile string) (*managedProcessSnapshot, error)
	Kill(pid int32) error
}

type managedProcessManager struct {
	mu         sync.Mutex
	logDir     string
	runner     managedProcessRunner
	processMap map[string]*managedProcessEntry
}

type systemManagedProcessRunner struct{}

var toolManagedProcessClient = newManagedProcessManager(getManagedProcessLogDir(), &systemManagedProcessRunner{})

func ToolManagedProcessStatus(c *gin.Context) {
	toolManagedProcessClient.syncLogDir(getManagedProcessLogDir())
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	status, err := toolManagedProcessClient.Status(dataMap, time.Now())
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, status)
}

func ToolManagedProcessEnsureRunning(c *gin.Context) {
	toolManagedProcessClient.syncLogDir(getManagedProcessLogDir())
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	status, err := toolManagedProcessClient.EnsureRunning(dataMap, time.Now())
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, status)
}

func ToolManagedProcessStart(c *gin.Context) {
	toolManagedProcessClient.syncLogDir(getManagedProcessLogDir())
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	status, err := toolManagedProcessClient.Start(dataMap, time.Now())
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, status)
}

func ToolManagedProcessStop(c *gin.Context) {
	toolManagedProcessClient.syncLogDir(getManagedProcessLogDir())
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	status, err := toolManagedProcessClient.Stop(dataMap, time.Now())
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, status)
}

func ToolManagedProcessRestart(c *gin.Context) {
	toolManagedProcessClient.syncLogDir(getManagedProcessLogDir())
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	status, err := toolManagedProcessClient.Restart(dataMap, time.Now())
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, status)
}

func ToolManagedProcessLogTail(c *gin.Context) {
	toolManagedProcessClient.syncLogDir(getManagedProcessLogDir())
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	config, err := normalizeManagedProcessConfig(dataMap, time.Now())
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	maxBytes := cast.ToInt(dataMap[`max_bytes`])
	if maxBytes <= 0 {
		maxBytes = defaultManagedProcessTailBytes
	}

	logFile := buildManagedProcessLogFile(toolManagedProcessClient.logDir, config.Key, time.Now())
	content, err := readManagedProcessLogTail(logFile, maxBytes)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`key`:      config.Key,
		`log_file`: logFile,
		`content`:  content,
	})
}

func newManagedProcessManager(logDir string, runner managedProcessRunner) *managedProcessManager {
	if logDir == `` {
		logDir = `logs`
	}
	return &managedProcessManager{
		logDir:     logDir,
		runner:     runner,
		processMap: make(map[string]*managedProcessEntry),
	}
}

func (m *managedProcessManager) syncLogDir(logDir string) {
	if logDir == `` {
		return
	}
	m.mu.Lock()
	m.logDir = logDir
	m.mu.Unlock()
}

func (m *managedProcessManager) Status(raw map[string]any, now time.Time) (*managedProcessStatus, error) {
	config, err := normalizeManagedProcessConfig(raw, now)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	entry := m.processMap[config.Key]
	if entry != nil {
		entry.Config = config
		if entry.Process != nil && entry.Process.PID > 0 {
			status := buildManagedProcessStatus(entry.Config, entry.Process)
			m.mu.Unlock()
			return status, nil
		}
	}
	m.mu.Unlock()

	found, err := m.runner.Find(config)
	if err != nil {
		return nil, err
	}
	if found == nil {
		return &managedProcessStatus{
			Key:         config.Key,
			Name:        config.Name,
			CommandLine: config.CommandLine,
			Workdir:     config.Workdir,
			StatusText:  `未运行`,
		}, nil
	}

	found.LogFile = buildManagedProcessLogFile(m.logDir, config.Key, now)
	found.StatusText = `运行中（外部进程）`
	m.setEntry(config, found)
	return buildManagedProcessStatus(config, found), nil
}

func (m *managedProcessManager) EnsureRunning(raw map[string]any, now time.Time) (*managedProcessStatus, error) {
	config, err := normalizeManagedProcessConfig(raw, now)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	entry := m.processMap[config.Key]
	if entry != nil {
		entry.Config = config
		if entry.Process != nil && entry.Process.PID > 0 {
			status := buildManagedProcessStatus(entry.Config, entry.Process)
			m.mu.Unlock()
			return status, nil
		}
	}
	m.mu.Unlock()

	found, err := m.runner.Find(config)
	if err != nil {
		return nil, err
	}
	if found != nil && found.PID > 0 {
		found.LogFile = buildManagedProcessLogFile(m.logDir, config.Key, now)
		found.StatusText = `运行中（外部进程）`
		m.setEntry(config, found)
		return buildManagedProcessStatus(config, found), nil
	}

	return m.startNormalized(config, now)
}

func (m *managedProcessManager) Start(raw map[string]any, now time.Time) (*managedProcessStatus, error) {
	return m.EnsureRunning(raw, now)
}

func (m *managedProcessManager) Stop(raw map[string]any, now time.Time) (*managedProcessStatus, error) {
	config, err := normalizeManagedProcessConfig(raw, now)
	if err != nil {
		return nil, err
	}

	var pid int32
	m.mu.Lock()
	entry := m.processMap[config.Key]
	if entry != nil {
		entry.Config = config
		if entry.Process != nil {
			pid = entry.Process.PID
		}
	}
	m.mu.Unlock()

	if pid == 0 {
		found, findErr := m.runner.Find(config)
		if findErr != nil {
			return nil, findErr
		}
		if found != nil {
			pid = found.PID
		}
	}

	if pid > 0 {
		if err = m.runner.Kill(pid); err != nil {
			return nil, err
		}
	}

	m.mu.Lock()
	if entry != nil {
		entry.Process = nil
	}
	m.mu.Unlock()

	return &managedProcessStatus{
		Key:         config.Key,
		Name:        config.Name,
		CommandLine: config.CommandLine,
		Workdir:     config.Workdir,
		LogFile:     buildManagedProcessLogFile(m.logDir, config.Key, now),
		StatusText:  `已停止`,
	}, nil
}

func (m *managedProcessManager) Restart(raw map[string]any, now time.Time) (*managedProcessStatus, error) {
	config, err := normalizeManagedProcessConfig(raw, now)
	if err != nil {
		return nil, err
	}

	m.mu.Lock()
	entry := m.processMap[config.Key]
	var pid int32
	if entry != nil && entry.Process != nil {
		pid = entry.Process.PID
	}
	m.mu.Unlock()

	if pid > 0 {
		if err = m.runner.Kill(pid); err != nil {
			return nil, err
		}
	} else {
		found, findErr := m.runner.Find(config)
		if findErr != nil {
			return nil, findErr
		}
		if found != nil && found.PID > 0 {
			if err = m.runner.Kill(found.PID); err != nil {
				return nil, err
			}
		}
	}

	return m.startNormalized(config, now)
}

func (m *managedProcessManager) startNormalized(config managedProcessConfig, now time.Time) (*managedProcessStatus, error) {
	logFile := buildManagedProcessLogFile(m.logDir, config.Key, now)
	snapshot, err := m.runner.Start(config, logFile)
	if err != nil {
		return nil, err
	}
	snapshot.LogFile = logFile
	snapshot.IsManaged = true
	if snapshot.StartedAt == 0 {
		snapshot.StartedAt = now.Unix()
	}
	snapshot.StatusText = `运行中`
	m.setEntry(config, snapshot)
	return buildManagedProcessStatus(config, snapshot), nil
}

func (m *managedProcessManager) setEntry(config managedProcessConfig, snapshot *managedProcessSnapshot) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.processMap[config.Key] = &managedProcessEntry{
		Config:  config,
		Process: snapshot,
	}
}

func buildManagedProcessStatus(config managedProcessConfig, snapshot *managedProcessSnapshot) *managedProcessStatus {
	return &managedProcessStatus{
		Key:         config.Key,
		Name:        config.Name,
		CommandLine: config.CommandLine,
		Workdir:     config.Workdir,
		Running:     snapshot != nil && snapshot.PID > 0,
		PID:         snapshot.PID,
		LogFile:     snapshot.LogFile,
		StartedAt:   snapshot.StartedAt,
		IsManaged:   snapshot.IsManaged,
		StatusText:  snapshot.StatusText,
	}
}

func normalizeManagedProcessConfig(raw map[string]any, now time.Time) (managedProcessConfig, error) {
	commandLine := strings.TrimSpace(cast.ToString(raw[`command_line`]))
	if commandLine == `` {
		return managedProcessConfig{}, errors.New(`command_line不能为空`)
	}

	parts, err := splitManagedCommandLine(commandLine)
	if err != nil {
		return managedProcessConfig{}, err
	}
	if len(parts) == 0 {
		return managedProcessConfig{}, errors.New(`command_line不能为空`)
	}

	name := strings.TrimSpace(cast.ToString(raw[`name`]))
	key := strings.TrimSpace(cast.ToString(raw[`key`]))
	if key == `` {
		if name != `` {
			key = sanitizeManagedProcessKey(name)
		} else {
			key = sanitizeManagedProcessKey(parts[0])
		}
	}
	if key == `` {
		key = fmt.Sprintf(`managed-%s`, now.Format(`20060102`))
	}
	if name == `` {
		name = key
	}

	return managedProcessConfig{
		Key:         key,
		Name:        name,
		CommandLine: commandLine,
		Workdir:     strings.TrimSpace(cast.ToString(raw[`workdir`])),
		Executable:  parts[0],
		Args:        parts[1:],
	}, nil
}

func splitManagedCommandLine(commandLine string) ([]string, error) {
	var parts []string
	var current strings.Builder
	var quote rune

	for _, ch := range strings.TrimSpace(commandLine) {
		switch {
		case quote != 0:
			if ch == quote {
				quote = 0
				continue
			}
			current.WriteRune(ch)
		case ch == '\'' || ch == '"':
			quote = ch
		case ch == ' ' || ch == '\t' || ch == '\n' || ch == '\r':
			if current.Len() == 0 {
				continue
			}
			parts = append(parts, current.String())
			current.Reset()
		default:
			current.WriteRune(ch)
		}
	}

	if quote != 0 {
		return nil, errors.New(`command_line引号未闭合`)
	}
	if current.Len() > 0 {
		parts = append(parts, current.String())
	}
	return parts, nil
}

func sanitizeManagedProcessKey(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	var builder strings.Builder
	lastDash := false
	for _, ch := range value {
		switch {
		case (ch >= 'a' && ch <= 'z') || (ch >= '0' && ch <= '9'):
			builder.WriteRune(ch)
			lastDash = false
		default:
			if builder.Len() == 0 || lastDash {
				continue
			}
			builder.WriteByte('-')
			lastDash = true
		}
	}
	return strings.Trim(builder.String(), `-`)
}

func buildManagedProcessLogFile(logDir, key string, now time.Time) string {
	return filepath.Join(logDir, fmt.Sprintf(`%s-%s.log`, sanitizeManagedProcessKey(key), now.Format(`2006-01-02`)))
}

func readManagedProcessLogTail(logFile string, maxBytes int) (string, error) {
	if maxBytes <= 0 {
		maxBytes = defaultManagedProcessTailBytes
	}

	file, err := os.Open(logFile)
	if errors.Is(err, os.ErrNotExist) {
		return ``, nil
	}
	if err != nil {
		return ``, err
	}
	defer func() { _ = file.Close() }()

	info, err := file.Stat()
	if err != nil {
		return ``, err
	}
	if info.Size() == 0 {
		return ``, nil
	}

	readSize := int64(maxBytes)
	if info.Size() < readSize {
		readSize = info.Size()
	}
	if _, err = file.Seek(-readSize, io.SeekEnd); err != nil {
		return ``, err
	}

	data := make([]byte, readSize)
	n, readErr := io.ReadFull(file, data)
	if readErr != nil && !errors.Is(readErr, io.ErrUnexpectedEOF) {
		return ``, readErr
	}
	content := string(data[:n])
	if readSize < info.Size() {
		if index := strings.Index(content, "\n"); index >= 0 && index < len(content)-1 {
			content = content[index+1:]
		}
	}
	return content, nil
}

func getManagedProcessLogDir() string {
	if component.EnvClient != nil && component.EnvClient.LogPath != `` {
		return component.EnvClient.LogPath
	}
	return `logs`
}

func (r *systemManagedProcessRunner) Find(config managedProcessConfig) (*managedProcessSnapshot, error) {
	processes, err := process.Processes()
	if err != nil {
		return nil, err
	}

	for _, item := range processes {
		cmdline, cmdErr := item.Cmdline()
		if cmdErr != nil {
			continue
		}
		if !isManagedProcessMatch(config, cmdline) {
			continue
		}
		return &managedProcessSnapshot{
			PID:        item.Pid,
			IsManaged:  false,
			StatusText: `运行中（外部进程）`,
		}, nil
	}
	return nil, nil
}

func (r *systemManagedProcessRunner) Start(config managedProcessConfig, logFile string) (*managedProcessSnapshot, error) {
	if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
		return nil, err
	}

	writer, err := newManagedProcessLogWriter(logFile)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(config.Executable, config.Args...)
	if config.Workdir != `` {
		cmd.Dir = config.Workdir
	}
	cmd.Stdout = writer
	cmd.Stderr = writer
	// 脱离父进程生命周期 / Detach from parent process lifecycle.
	prepareManagedProcessCommand(cmd)

	if err = cmd.Start(); err != nil {
		_ = writer.Close()
		return nil, err
	}

	go func() {
		_ = cmd.Wait()
		_ = writer.Close()
	}()

	return &managedProcessSnapshot{
		PID:       int32(cmd.Process.Pid),
		LogFile:   logFile,
		StartedAt: time.Now().Unix(),
		IsManaged: true,
	}, nil
}

func (r *systemManagedProcessRunner) Kill(pid int32) error {
	proc, err := os.FindProcess(int(pid))
	if err != nil {
		return err
	}
	return proc.Kill()
}

func isManagedProcessMatch(config managedProcessConfig, cmdline string) bool {
	parts, err := splitManagedCommandLine(cmdline)
	if err != nil || len(parts) == 0 {
		return strings.EqualFold(strings.TrimSpace(cmdline), strings.TrimSpace(config.CommandLine))
	}

	if normalizeManagedExecutable(parts[0]) != normalizeManagedExecutable(config.Executable) {
		return false
	}
	if len(parts)-1 != len(config.Args) {
		return false
	}
	for index, arg := range config.Args {
		if !strings.EqualFold(strings.TrimSpace(parts[index+1]), strings.TrimSpace(arg)) {
			return false
		}
	}
	return true
}

func normalizeManagedExecutable(value string) string {
	value = strings.ToLower(strings.TrimSpace(value))
	value = filepath.Base(value)
	return strings.TrimSuffix(value, `.exe`)
}

type managedProcessLogWriter struct {
	mu       sync.Mutex
	baseFile string
	file     *os.File
}

func newManagedProcessLogWriter(logFile string) (*managedProcessLogWriter, error) {
	writer := &managedProcessLogWriter{baseFile: logFile}
	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, err
	}
	writer.file = file
	return writer, nil
}

func (w *managedProcessLogWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		file, err := os.OpenFile(w.baseFile, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
		if err != nil {
			return 0, err
		}
		w.file = file
	}

	p = bytes.ReplaceAll(p, []byte("\x00"), []byte{})
	return w.file.Write(p)
}

func (w *managedProcessLogWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	return err
}
