package p_claude

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/cast"
)

// RunClaudeStream 执行 claude 命令并逐行推送解析后的消息。
// callback 每收到一行 stream-json 时同步调用。
// 返回 sessionID（首行 init 中提取）和 error。
func RunClaudeStream(ctx context.Context, cfg RunConfig, callback func(msg StreamMessage)) (string, error) {
	args := buildArgs(cfg)
	env := buildEnv(cfg)

	cmd := exec.CommandContext(ctx, `claude`, args...)
	cmd.Dir = cfg.WorkingDir
	cmd.Env = env

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return ``, fmt.Errorf("stdout pipe failed: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return ``, fmt.Errorf("claude start failed: %w", err)
	}

	sessionID := ``
	sessionExtracted := false
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == `` {
			continue
		}
		msg := parseLine(line)
		callback(msg)

		if !sessionExtracted && cfg.SessionID == `` {
			if sid := extractSessionIDFromLine(line); sid != `` {
				sessionID = sid
				sessionExtracted = true
			}
		}
	}

	_ = cmd.Wait()
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
		msg.Type = `parse_error`
		msg.Data = map[string]any{`error`: err.Error(), `line`: line}
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
