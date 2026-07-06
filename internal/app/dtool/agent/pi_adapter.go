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
	// 设置自定义 API 地址 → provider 对应的 _BASE_URL 环境变量
	if config.ModelAddr != "" {
		envKey := apiBaseURLEnvName(config.Provider)
		if envKey != "" {
			a.cmd.Env = append(a.cmd.Env, envKey+"="+config.ModelAddr)
		}
	}
	// 设置 API Key 环境变量
	if config.ApiKey != "" {
		envKey := apiKeyEnvName(config.Provider)
		if envKey != "" {
			a.cmd.Env = append(a.cmd.Env, envKey+"="+config.ApiKey)
		}
	}

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

func (a *PiAdapter) IsInstalled() bool {
	_, err := exec.LookPath("pi")
	return err == nil
}

func (a *PiAdapter) InstallHint() string {
	return "Pi 未安装，请执行: npm install -g --ignore-scripts @earendil-works/pi-coding-agent"
}

// apiKeyEnvName 根据 provider 返回对应的 API Key 环境变量名
func apiKeyEnvName(provider string) string {
	switch provider {
	case "anthropic":
		return "ANTHROPIC_API_KEY"
	case "openai":
		return "OPENAI_API_KEY"
	case "google":
		return "GOOGLE_API_KEY"
	case "deepseek":
		return "DEEPSEEK_API_KEY"
	default:
		return ""
	}
}

// apiBaseURLEnvName 根据 provider 返回对应的 Base URL 环境变量名
func apiBaseURLEnvName(provider string) string {
	switch provider {
	case "anthropic":
		return "ANTHROPIC_BASE_URL"
	case "openai":
		return "OPENAI_BASE_URL"
	case "google":
		return "GOOGLE_BASE_URL"
	case "deepseek":
		return "DEEPSEEK_BASE_URL"
	default:
		return ""
	}
}
