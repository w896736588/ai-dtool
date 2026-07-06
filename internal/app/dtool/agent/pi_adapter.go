package agent

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os/exec"
	"sync"
)

// PiAdapter Pi Agent 的 RPC 模式适配器
type PiAdapter struct {
	mu      sync.Mutex
	cmd     *exec.Cmd
	stdin   io.WriteCloser
	running bool
	events  chan AgentEvent
	done    chan struct{}
	exitErr error
}

func NewPiAdapter() *PiAdapter {
	return &PiAdapter{
		events: make(chan AgentEvent, 256),
	}
}

func (a *PiAdapter) Start(ctx context.Context, config AgentStartConfig) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if a.running {
		return fmt.Errorf("pi adapter already running")
	}

	args := []string{"--mode", "rpc"}
	if config.Provider != "" {
		args = append(args, "--provider", config.Provider)
	}
	if config.Model != "" {
		args = append(args, "--model", config.Model)
	}
	if config.SessionDir != "" {
		args = append(args, "--session-dir", config.SessionDir)
	}
	args = append(args, config.ExtraArgs...)

	a.cmd = exec.CommandContext(ctx, "pi", args...)
	if config.WorkDir != "" {
		a.cmd.Dir = config.WorkDir
	}
	a.cmd.Env = a.cmd.Environ()
	for k, v := range config.Env {
		a.cmd.Env = append(a.cmd.Env, k+"="+v)
	}
	// models.json 包含 baseUrl/apiKey/api 全部信息，不再通过环境变量覆盖
	// 环境变量会覆盖 models.json 导致冲突（如 baseUrl 双拼接）

	var err error
	a.stdin, err = a.cmd.StdinPipe()
	if err != nil {
		return fmt.Errorf("create stdin pipe: %w", err)
	}

	stdout, err := a.cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("create stdout pipe: %w", err)
	}

	// 捕获 stderr（必须在 Start 前调用）
	stderr, err := a.cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("create stderr pipe: %w", err)
	}

	if err := a.cmd.Start(); err != nil {
		return fmt.Errorf("start pi process: %w", err)
	}

	a.running = true
	a.done = make(chan struct{})
	a.events = make(chan AgentEvent, 256)

	// stderr → 日志
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			log.Printf("[agent-v2/pi] stderr: %s", scanner.Text())
		}
	}()

	// 监控进程退出
	go func() {
		err := a.cmd.Wait()
		a.mu.Lock()
		a.running = false
		a.exitErr = err
		a.mu.Unlock()
		if err != nil {
			log.Printf("[agent-v2/pi] pi process exited with error: %v", err)
		} else {
			log.Printf("[agent-v2/pi] pi process exited normally")
		}
	}()

	go ReadJSONLLines(stdout, a.events, a.done)

	return nil
}

func (a *PiAdapter) Stop() error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running {
		return nil
	}

	// 发送 abort 命令优雅退出
	abortCmd, _ := json.Marshal(map[string]string{"type": "abort"})
	abortCmd = append(abortCmd, '\n')
	if a.stdin != nil {
		a.stdin.Write(abortCmd)
	}

	if a.done != nil {
		close(a.done)
	}
	if a.cmd != nil && a.cmd.Process != nil {
		a.cmd.Process.Kill()
	}
	a.running = false
	return nil
}

func (a *PiAdapter) SendCommand(raw json.RawMessage) error {
	a.mu.Lock()
	defer a.mu.Unlock()

	if !a.running || a.stdin == nil {
		return fmt.Errorf("adapter not running")
	}

	// 确保以换行符结尾
	data := make([]byte, len(raw))
	copy(data, raw)
	if len(data) == 0 || data[len(data)-1] != '\n' {
		data = append(data, '\n')
	}

	_, err := a.stdin.Write(data)
	return err
}

func (a *PiAdapter) Events() <-chan AgentEvent {
	return a.events
}

func (a *PiAdapter) IsRunning() bool {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.running
}

// ExitError 返回进程退出错误（仅在进程已退出后有效）
func (a *PiAdapter) ExitError() error {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.exitErr
}

func (a *PiAdapter) IsInstalled() bool {
	_, err := exec.LookPath("pi")
	return err == nil
}

func (a *PiAdapter) InstallHint() string {
	return "Pi 未安装，请执行: npm install -g --ignore-scripts @earendil-works/pi-coding-agent"
}
