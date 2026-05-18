package p_claude

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cast"
)

// maxScanTokenSize bufio.Scanner 最大缓冲区大小（10MB）。
// stream-json 输出中单行可能远超默认 64KB（尤其是工具调用结果）。
const maxScanTokenSize = 10 * 1024 * 1024

// ptyResult 进程启动结果，由各平台 startClaude 实现返回。
type ptyResult struct {
	lineCh   <-chan string       // stdout 行数据通道
	stderrCh <-chan string       // stderr 行数据通道
	pid      int                 // 进程 ID
	waitFn   func() (int, error) // 等待进程退出并返回退出码
	closeFn  func()              // 强制终止进程
}

// RunClaudeStream 执行 claude 命令并逐行推送解析后的消息。
// callback 每收到一行 stream-json 时同步调用。
// 返回 sessionID（首行 init 中提取）和 error。
func RunClaudeStream(ctx context.Context, cfg RunConfig, callback func(msg StreamMessage)) (string, error) {
	args := buildArgs(cfg)
	env := buildEnv(cfg)

	log.Printf("[claude-exec] 启动进程, dir=%s model=%s", cfg.WorkingDir, cfg.Model)

	result, err := startClaude(ctx, args, cfg.WorkingDir, env)
	if err != nil {
		log.Printf("[claude-exec] 启动失败: %v", err)
		return ``, fmt.Errorf("claude start failed: %w", err)
	}
	defer result.closeFn()
	log.Printf("[claude-exec] 进程已启动, pid=%d", result.pid)

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
	for line := range result.lineCh {
		line = strings.TrimSpace(line)
		if line == `` {
			continue
		}
		lineCount++
		if lineCount <= 3 {
			log.Printf("[claude-exec] 收到第%d行(len=%d): %.200s", lineCount, len(line), line)
		}
		msg := parseLine(line)
		callback(msg)

		if !sessionExtracted && cfg.SessionID == `` {
			if sid := extractSessionIDFromLine(line); sid != `` {
				sessionID = sid
				log.Printf("[claude-exec] 提取到 session_id=%s", sid)
				sessionExtracted = true
			}
		}
	}

	log.Printf("[claude-exec] 行通道关闭, 总行数=%d", lineCount)
	<-stderrDone

	exitCode, waitErr := result.waitFn()
	stderrSummary := strings.Join(stderrLines, "\n")
	log.Printf("[claude-exec] 进程结束, exitCode=%d waitErr=%v stderr=%s", exitCode, waitErr, stderrSummary)
	if waitErr != nil {
		if stderrSummary != `` {
			return sessionID, fmt.Errorf("claude 退出异常: %s (stderr: %s)", waitErr.Error(), stderrSummary)
		}
		return sessionID, fmt.Errorf("claude 退出异常: %w", waitErr)
	}
	if exitCode != 0 {
		if stderrSummary != `` {
			return sessionID, fmt.Errorf("claude 返回失败 (exit code %d): %s", exitCode, stderrSummary)
		}
		return sessionID, fmt.Errorf("claude 返回失败 (exit code %d)，无更多错误详情", exitCode)
	}
	return sessionID, nil
}

// buildArgs 构建 claude 命令行参数。
func buildArgs(cfg RunConfig) []string {
	args := []string{}
	if cfg.SessionID != `` {
		args = append(args, `--resume`, cfg.SessionID)
	}
	args = append(args,
		`-p`, sanitizePrompt(cfg.Prompt),
		`--add-dir`, cfg.WorkingDir,
		`--output-format`, `stream-json`,
		`--include-partial-messages`,
		`--verbose`,
		`--permission-mode`, `bypassPermissions`,
	)
	if cfg.Model != `` {
		args = append(args, `--model`, cfg.Model)
	}
	if cfg.UserDataDir != `` {
		args = append(args, `--user-data-dir`, cfg.UserDataDir)
	}
	if cfg.SettingsPath != `` {
		args = append(args, `--settings`, cfg.SettingsPath)
	}
	if cfg.Effort != `` {
		args = append(args, `--effort`, cfg.Effort)
	}
	return args
}

// buildEnv 构建环境变量。
func buildEnv(cfg RunConfig) []string {
	env := os.Environ()
	if cfg.BaseURL != `` {
		env = append(env, `ANTHROPIC_BASE_URL=`+cfg.BaseURL)
	}
	if cfg.APIKey != `` {
		env = append(env, `ANTHROPIC_API_KEY=`+cfg.APIKey)
	}
	if cfg.ThinkingBudget > 0 {
		env = append(env, `THINKING_BUDGET=`+strconv.Itoa(cfg.ThinkingBudget))
	}
	return env
}

// parseLine 解析单行 stream-json 为 StreamMessage。
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
	msg.Subtype = cast.ToString(raw[`subtype`])
	if event, ok := raw[`event`].(map[string]any); ok {
		msg.Event = cast.ToString(event[`type`])
	}
	msg.Data = raw
	return msg
}

// sanitizePrompt 将 prompt 中的换行符替换为空格，避免多行内容作为命令行参数传递时导致 claude CLI 解析异常。
// 如果内容以 "-" 开头，在前面加空格，防止 claude CLI 将其误解析为选项标志。
func sanitizePrompt(prompt string) string {
	s := strings.ReplaceAll(prompt, "\r\n", " ")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	if strings.HasPrefix(s, "-") {
		s = " " + s
	}
	return s
}

// extractSessionIDFromLine 从 stream-json 行提取 session_id。
func extractSessionIDFromLine(line string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return ``
	}
	if cast.ToString(data[`type`]) != `system` || cast.ToString(data[`subtype`]) != `init` {
		return ``
	}
	return cast.ToString(data[`session_id`])
}
