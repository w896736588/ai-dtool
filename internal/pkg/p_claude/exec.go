package p_claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cast"
)

// maxScanTokenSize bufio.Scanner 最大缓冲区大小（10MB）。
// stream-json 输出中单行可能远超默认 64KB（尤其是工具调用结果）。
const maxScanTokenSize = 10 * 1024 * 1024

// RunClaudeStream 执行 claude 命令并逐行推送解析后的消息。
// callback 每收到一行 stream-json 时同步调用。
// 返回 sessionID（首行 init 中提取）和 error。
func RunClaudeStream(ctx context.Context, cfg RunConfig, callback func(msg StreamMessage)) (string, error) {
	args := buildArgs(cfg)
	env := buildEnv(cfg)

	log.Printf("[claude-exec] 启动进程, dir=%s model=%s", cfg.WorkingDir, cfg.Model)

	cmd := exec.CommandContext(ctx, `claude`, args...)
	cmd.Dir = cfg.WorkingDir
	cmd.Env = env
	cmd.Stderr = os.Stderr

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		log.Printf("[claude-exec] stdout pipe 失败: %v", err)
		return ``, fmt.Errorf("stdout pipe failed: %w", err)
	}

	if err := cmd.Start(); err != nil {
		log.Printf("[claude-exec] 启动失败: %v", err)
		return ``, fmt.Errorf("claude start failed: %w", err)
	}
	log.Printf("[claude-exec] 进程已启动, pid=%d", cmd.Process.Pid)

	sessionID := ``
	sessionExtracted := false
	lineCount := 0
	scanner := bufio.NewScanner(stdout)
	scanner.Buffer(make([]byte, maxScanTokenSize), maxScanTokenSize)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
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

	log.Printf("[claude-exec] scanner 循环结束, 总行数=%d", lineCount)
	if err := scanner.Err(); err != nil {
		log.Printf("[claude-exec] scanner 错误: %v", err)
		return sessionID, fmt.Errorf("scan stdout: %w", err)
	}

	waitErr := cmd.Wait()
	log.Printf("[claude-exec] 进程结束, waitErr=%v", waitErr)
	if waitErr != nil {
		return sessionID, fmt.Errorf("claude exited with error: %w", waitErr)
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
		`-p`, cfg.Prompt,
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
	return args
}

// buildEnv 构建环境变量（在系统环境基础上追加）。
func buildEnv(cfg RunConfig) []string {
	env := os.Environ()
	if cfg.BaseURL != `` {
		env = append(env, `ANTHROPIC_BASE_URL=`+cfg.BaseURL)
	}
	if cfg.APIKey != `` {
		env = append(env, `ANTHROPIC_API_KEY=`+cfg.APIKey)
	}
	return env
}

// parseLine 解析单行 stream-json 为 StreamMessage。
func parseLine(line string) StreamMessage {
	msg := StreamMessage{RawJSON: line}
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		// stream-json 模式下偶有非 JSON 文本行（如帮助信息、中文提示），
		// 不视为错误，直接透传原始内容作为普通文本展示
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
