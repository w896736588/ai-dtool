package agent

import (
	"context"
	"encoding/json"
	"io"
)

// AgentEvent Agent 输出的事件
type AgentEvent struct {
	Raw json.RawMessage `json:"raw"`
}

// AgentAdapter Agent 适配器接口，所有 Agent 类型（Pi/Codex/Claude）需实现此接口
type AgentAdapter interface {
	// Start 启动 Agent 子进程
	Start(ctx context.Context, config AgentStartConfig) error
	// Stop 停止 Agent 子进程
	Stop() error
	// SendCommand 向 Agent 发送命令（JSON 格式，写入子进程 stdin）
	SendCommand(raw json.RawMessage) error
	// Events 返回事件通道（从子进程 stdout 读取的 JSONL 行）
	Events() <-chan AgentEvent
	// IsRunning 检查是否正在运行
	IsRunning() bool
	// IsInstalled 检查 Agent CLI 是否已安装
	IsInstalled() bool
	// InstallHint 返回安装提示
	InstallHint() string
}

// AgentStartConfig 启动配置
type AgentStartConfig struct {
	WorkDir    string            // 工作目录
	SessionDir string            // 会话持久化目录
	SessionID  string            // Agent 原生会话 ID（用于 resume）
	Provider   string            // LLM Provider（如 anthropic, openai）
	Model      string            // 模型 ID
	ModelAddr  string            // 自定义模型 API 地址（如 https://api.example.com/v1）
	ApiKey     string            // API Key
	ExtraArgs  []string          // 额外 CLI 参数
	Env        map[string]string // 环境变量
}

// ReadJSONLLines 从 reader 逐行读取 JSONL，发送到通道
func ReadJSONLLines(r io.Reader, ch chan<- AgentEvent, done <-chan struct{}) {
	defer close(ch)
	buf := make([]byte, 0, 65536)
	tmp := make([]byte, 4096)
	for {
		select {
		case <-done:
			return
		default:
		}
		n, err := r.Read(tmp)
		if n > 0 {
			buf = append(buf, tmp[:n]...)
			for {
				idx := -1
				for i, b := range buf {
					if b == '\n' {
						idx = i
						break
					}
				}
				if idx < 0 {
					break
				}
				line := buf[:idx]
				buf = buf[idx+1:]
				if len(line) == 0 {
					continue
				}
				// 去掉末尾 \r
				if line[len(line)-1] == '\r' {
					line = line[:len(line)-1]
				}
				if len(line) == 0 {
					continue
				}
				select {
				case ch <- AgentEvent{Raw: line}:
				case <-done:
					return
				}
			}
		}
		if err != nil {
			return
		}
	}
}
