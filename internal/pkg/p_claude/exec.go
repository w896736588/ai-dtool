package p_claude

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/spf13/cast"
)

// CleanupOrphanedMcpProcesses 清理残留的 chrome-devtools-mcp 子进程。
// 用于 Go 进程崩溃后重启时的兜底清扫（Windows Job Object / Unix 进程组已覆盖正常退出场景）。
func CleanupOrphanedMcpProcesses() {
	cleanupOrphanedMcpProcesses()
}

// maxScanTokenSize bufio.Scanner 最大缓冲区大小（10MB）。
// stream-json 输出中单行可能远超默认 64KB（尤其是工具调用结果）。
const maxScanTokenSize = 10 * 1024 * 1024

// ptyResult 进程启动结果，由各平台 startClaude 实现返回。
type ptyResult struct {
	lineCh   <-chan string       // stdout 行数据通道
	stderrCh <-chan string       // stderr 行数据通道
	pid      int                 // 进程 ID
	waitFn   func() (int, error) // 等待进程退出并返回退出码
	closeFn  func()              // 强制终止进程（含子进程清理：Windows Job Object / Unix 进程组）
}

// RunClaudeStream 执行 claude 命令并逐行推送解析后的消息。
// callback 每收到一行 stream-json 时同步调用。
// 返回 sessionID（首行 init 中提取）和 error。
func RunClaudeStream(ctx context.Context, cfg RunConfig, callback func(msg StreamMessage)) (string, error) {
	args := buildArgs(cfg)
	env := buildEnv(cfg)
	stdinFile, cleanupPromptFile, err := preparePromptStdinFile(cfg.Prompt)
	if err != nil {
		return ``, fmt.Errorf("prepare claude prompt file failed: %w", err)
	}
	defer cleanupPromptFile()

	log.Printf("[claude-exec] 启动进程, dir=%s model=%s", cfg.WorkingDir, cfg.Model)
	log.Printf("[claude-exec] 完整参数: %v", args)
	if cfg.SessionID != "" {
		log.Printf("[claude-exec] 尝试恢复 session_id=%s", cfg.SessionID)
	}
	if cfg.SettingsPath != "" {
		log.Printf("[claude-exec] settings 路径=%s", cfg.SettingsPath)
	}

	result, err := startClaude(ctx, args, cfg.WorkingDir, env, stdinFile)
	if err != nil {
		log.Printf("[claude-exec] 启动失败: %v", err)
		return ``, fmt.Errorf("claude start failed: %w", err)
	}
	defer result.closeFn()
	log.Printf("[claude-exec] 进程已启动, pid=%d", result.pid)

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
			log.Printf("[claude-exec] 上下文已取消，退出读取循环 (lineCount=%d)", lineCount)
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
	}
doneReading:

	log.Printf("[claude-exec] 行通道关闭, 总行数=%d", lineCount)
	<-stderrDone

	exitCode, waitErr := result.waitFn()
	stderrSummary := strings.Join(stderrLines, "\n")
	log.Printf("[claude-exec] 进程结束, exitCode=%d waitErr=%v lineCount=%d stderrLineCount=%d", exitCode, waitErr, lineCount, len(stderrLines))
	if stderrSummary != "" {
		log.Printf("[claude-exec] stderr内容: %s", stderrSummary)
	}
	if waitErr != nil {
		if stderrSummary != `` {
			return sessionID, fmt.Errorf("claude 退出异常: %s (stderr: %s)", waitErr.Error(), stderrSummary)
		}
		return sessionID, fmt.Errorf("claude 退出异常: %w", waitErr)
	}
	if exitCode != 0 {
		// 增强错误信息：当 stderr 为空时，提供更多上下文帮助排查
		if stderrSummary != `` {
			return sessionID, fmt.Errorf("claude 返回失败 (exit code %d): %s", exitCode, stderrSummary)
		}
		// stderr 为空时，输出更多诊断信息
		errMsg := fmt.Sprintf("claude 返回失败 (exit code %d)，无 stderr 输出。", exitCode)
		if cfg.SessionID != "" {
			errMsg += fmt.Sprintf(" session_id=%s", cfg.SessionID)
		}
		if lineCount == 0 {
			errMsg += " 未收到任何 stdout 输出，可能是 Claude CLI 启动失败或配置错误。"
		} else {
			errMsg += fmt.Sprintf(" 已收到 %d 行输出。", lineCount)
		}
		if cfg.SettingsPath != "" {
			errMsg += fmt.Sprintf(" settings=%s", cfg.SettingsPath)
		}
		return sessionID, errors.New(errMsg)
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
		`-p`,
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

func preparePromptStdinFile(prompt string) (*os.File, func(), error) {
	file, err := os.CreateTemp("", "dtool-claude-prompt-*.txt")
	if err != nil {
		return nil, func() {}, err
	}

	cleanup := func() {
		_ = file.Close()
		_ = os.Remove(file.Name())
	}

	if _, err := io.WriteString(file, prompt); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	if _, err := file.Seek(0, 0); err != nil {
		cleanup()
		return nil, func() {}, err
	}
	return file, cleanup, nil
}

// BuildCommandLine 根据配置构建完整的 claude CLI 命令字符串（用于前端展示）。
func BuildCommandLine(cfg RunConfig) string {
	var sb strings.Builder
	sb.WriteString(`claude`)
	if cfg.SessionID != `` {
		sb.WriteString(` --resume `)
		sb.WriteString(cfg.SessionID)
	}
	sb.WriteString(` -p < prompt-file`)
	sb.WriteString(` --add-dir "`)
	sb.WriteString(cfg.WorkingDir)
	sb.WriteString(`"`)
	sb.WriteString(` --output-format stream-json --include-partial-messages --verbose --permission-mode bypassPermissions`)
	if cfg.Model != `` {
		sb.WriteString(` --model `)
		sb.WriteString(cfg.Model)
	}
	if cfg.UserDataDir != `` {
		sb.WriteString(` --user-data-dir "`)
		sb.WriteString(cfg.UserDataDir)
		sb.WriteString(`"`)
	}
	if cfg.SettingsPath != `` {
		sb.WriteString(` --settings "`)
		sb.WriteString(cfg.SettingsPath)
		sb.WriteString(`"`)
	}
	if cfg.Effort != `` {
		sb.WriteString(` --effort `)
		sb.WriteString(cfg.Effort)
	}
	return sb.String()
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
