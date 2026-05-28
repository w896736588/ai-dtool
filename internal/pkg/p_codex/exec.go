package p_codex

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/spf13/cast"
)

// CleanupOrphanedCodexProcesses 清理残留的 codex 子进程。
// 用于 Go 进程崩溃后重启时的兜底清扫。
func CleanupOrphanedCodexProcesses() {
	cleanupOrphanedCodexProcesses()
}

// maxScanTokenSize bufio.Scanner 最大缓冲区大小（10MB）。
// JSONL 输出中单行可能远超默认 64KB（尤其是工具调用结果）。
const maxScanTokenSize = 10 * 1024 * 1024

// ptyResult 进程启动结果，由各平台 startCodex 实现返回。
type ptyResult struct {
	lineCh   <-chan string       // stdout 行数据通道
	stderrCh <-chan string       // stderr 行数据通道
	pid      int                 // 进程 ID
	waitFn   func() (int, error) // 等待进程退出并返回退出码
	closeFn  func()              // 强制终止进程（含子进程清理）
}

// RunCodexStream 执行 codex exec 命令并逐行推送解析后的消息。
// callback 每收到一行 JSONL 时同步调用。
// 返回 sessionID（从 thread.started 事件的 thread_id 提取）和 error。
func RunCodexStream(ctx context.Context, cfg RunConfig, callback func(msg StreamMessage)) (string, error) {
	args := buildArgs(cfg)
	env := buildEnv(cfg)

	log.Printf("[codex-exec] 启动进程, dir=%s model=%s", cfg.WorkingDir, cfg.Model)

	result, err := startCodex(ctx, args, cfg.WorkingDir, env)
	if err != nil {
		log.Printf("[codex-exec] 启动失败: %v", err)
		return ``, fmt.Errorf("codex start failed: %w", err)
	}
	defer result.closeFn()
	log.Printf("[codex-exec] 进程已启动, pid=%d", result.pid)

	// 进程启动回调，通知上层记录 PID
	if cfg.ProcessStartCallback != nil {
		cfg.ProcessStartCallback(result.pid)
	}

	// 后台收集 stderr
	var stderrLines []string
	stderrDone := make(chan struct{})
	go func() {
		defer close(stderrDone)
		for line := range result.stderrCh {
			stderrLines = append(stderrLines, line)
		}
	}()

	sessionID := ``
	sessionExtracted := false
	lineCount := 0
	for {
		select {
		case <-ctx.Done():
			log.Printf("[codex-exec] 上下文已取消，退出读取循环 (lineCount=%d)", lineCount)
			return sessionID, ctx.Err()
		case line, ok := <-result.lineCh:
			if !ok {
				goto doneReading
			}
			line = strings.TrimSpace(line)
			if line == `` {
				continue
			}
			lineCount++
			if lineCount <= 3 {
				log.Printf("[codex-exec] 收到第%d行(len=%d): %.200s", lineCount, len(line), line)
			}
			msg := parseLine(line)
			callback(msg)

			// 从 thread.started 事件提取 thread_id 作为 session_id
			if !sessionExtracted && cfg.SessionID == `` {
				if sid := extractSessionIDFromLine(line); sid != `` {
					sessionID = sid
					log.Printf("[codex-exec] 提取到 session_id(thread_id)=%s", sid)
					sessionExtracted = true
				}
			}
		}
	}
doneReading:

	log.Printf("[codex-exec] 行通道关闭, 总行数=%d", lineCount)
	<-stderrDone

	exitCode, waitErr := result.waitFn()
	stderrSummary := strings.Join(stderrLines, "\n")
	log.Printf("[codex-exec] 进程结束, exitCode=%d waitErr=%v stderr=%s", exitCode, waitErr, stderrSummary)
	if waitErr != nil {
		if stderrSummary != `` {
			return sessionID, fmt.Errorf("codex 退出异常: %s (stderr: %s)", waitErr.Error(), stderrSummary)
		}
		return sessionID, fmt.Errorf("codex 退出异常: %w", waitErr)
	}
	if exitCode != 0 {
		if stderrSummary != `` {
			return sessionID, fmt.Errorf("codex 返回失败 (exit code %d): %s", exitCode, stderrSummary)
		}
		return sessionID, fmt.Errorf("codex 返回失败 (exit code %d)，无更多错误详情", exitCode)
	}
	return sessionID, nil
}

// buildArgs 构建 codex exec 命令行参数。
func buildArgs(cfg RunConfig) []string {
	args := []string{}

	if cfg.SessionID != `` {
		// 续接会话：codex exec resume <session_id> [--json] [prompt]
		// resume 不支持 --sandbox 参数，需使用 --dangerously-bypass-approvals-and-sandbox 达到等效权限
		args = append(args, `exec`, `resume`, cfg.SessionID, `--json`, `--dangerously-bypass-approvals-and-sandbox`)
		if cfg.Prompt != `` {
			args = append(args, sanitizePrompt(cfg.Prompt))
		}
	} else {
		// 新会话：codex exec <prompt> --json --cd <dir> --model <model> --sandbox <mode>
		args = append(args, `exec`, sanitizePrompt(cfg.Prompt), `--json`)
		if cfg.WorkingDir != `` {
			args = append(args, `--cd`, cfg.WorkingDir)
		}
		if cfg.Model != `` {
			args = append(args, `--model`, cfg.Model)
		}
		sandboxMode := cfg.SandboxMode
		if sandboxMode == `` {
			sandboxMode = DefaultSandboxMode
		}
		args = append(args, `--sandbox`, sandboxMode)
	}

	return args
}

// buildEnv 构建环境变量，注入 Codex 所需的 API 配置。
func buildEnv(cfg RunConfig) []string {
	env := os.Environ()
	if cfg.APIKey != `` {
		env = append(env, `OPENAI_API_KEY=`+cfg.APIKey)
	}
	// 自定义 provider 模式由 ~/.codex/config.toml 的 model_providers 段承载 base_url，
	// 避免再注入 OPENAI_BASE_URL 覆盖 provider 路由。 // Custom provider mode stores base_url in model_providers; avoid overriding it via OPENAI_BASE_URL.
	if cfg.BaseURL != `` && cfg.APIKey == `` {
		env = append(env, `OPENAI_BASE_URL=`+cfg.BaseURL)
	}
	return env
}

// parseLine 解析单行 JSONL 为 StreamMessage。
func parseLine(line string) StreamMessage {
	msg := StreamMessage{RawJSON: line}
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		msg.Type = `raw_text`
		msg.Data = map[string]any{`text`: line}
		if errJSON, e := json.Marshal(msg); e == nil {
			msg.RawJSON = string(errJSON)
		}
		return msg
	}
	msg.Type = cast.ToString(raw[`type`])
	msg.Data = raw

	// 提取 item 相关字段（item.started/item.updated/item.completed 事件）
	if item, ok := raw[`item`].(map[string]any); ok {
		msg.ItemType = cast.ToString(item[`type`])
		msg.ItemID = cast.ToString(item[`id`])
	}

	return msg
}

// sanitizePrompt 将 prompt 中的换行符替换为空格，避免多行内容作为命令行参数传递时导致解析异常。
// 如果内容以 "-" 开头，在前面加空格，防止被误解析为选项标志。
func sanitizePrompt(prompt string) string {
	s := strings.ReplaceAll(prompt, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	if strings.HasPrefix(s, "-") {
		s = " " + s
	}
	return s
}

// BuildCommandLine 根据配置构建完整的 codex CLI 命令字符串（用于前端展示）。
func BuildCommandLine(cfg RunConfig) string {
	var sb strings.Builder
	sb.WriteString(`codex`)
	if cfg.SessionID != `` {
		sb.WriteString(` exec resume `)
		sb.WriteString(cfg.SessionID)
		sb.WriteString(` --json --dangerously-bypass-approvals-and-sandbox`)
		if cfg.Prompt != `` {
			sb.WriteString(` "`)
			sb.WriteString(sanitizePrompt(cfg.Prompt))
			sb.WriteString(`"`)
		}
	} else {
		sb.WriteString(` exec "`)
		sb.WriteString(sanitizePrompt(cfg.Prompt))
		sb.WriteString(`" --json`)
		if cfg.WorkingDir != `` {
			sb.WriteString(` --cd `)
			sb.WriteString(cfg.WorkingDir)
		}
		if cfg.Model != `` {
			sb.WriteString(` --model `)
			sb.WriteString(cfg.Model)
		}
		sandboxMode := cfg.SandboxMode
		if sandboxMode == `` {
			sandboxMode = DefaultSandboxMode
		}
		sb.WriteString(` --sandbox `)
		sb.WriteString(sandboxMode)
	}
	return sb.String()
}

// extractSessionIDFromLine 从 JSONL 行提取 session_id。
// Codex 的 thread.started 事件包含 thread_id，即为 resume 使用的 session ID。
func extractSessionIDFromLine(line string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return ``
	}
	if cast.ToString(data[`type`]) != `thread.started` {
		return ``
	}
	return cast.ToString(data[`thread_id`])
}
