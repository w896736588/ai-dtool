package p_claude_sdk

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

// =============================================================================
// RunClaudeSdkStream：使用 claude-agent-sdk-go 执行对话的核心入口
// =============================================================================
// 接口签名与 p_claude.RunClaudeStream 完全一致，上层调用无需修改。
// 返回 sessionID 和 error。
// =============================================================================

// ClaudeCLIExecutable Claude CLI 可执行文件路径。
const ClaudeCLIExecutable = "claude"

// RunClaudeSdkStream 使用 claude-agent-sdk-go 执行对话并逐条推送消息。
//
// 当前实现采用"子进程 + stream-json"模式作为 Phase 1 基础框架。
// 该模式复用现有的进程管理机制（Windows Job Object / Unix 进程组），
// 同时通过独立的 permission.go / hook.go 提供权限审批和 Hook 事件桥接能力。
//
// 待 claude-agent-sdk-go 依赖正式引入且 Windows 兼容性验证通过后，
// 可切换为真正的 SDK 双向控制协议模式。
//
// callback 每收到一条消息时同步调用。
func RunClaudeSdkStream(ctx context.Context, cfg RunConfig, callback func(msg StreamMessage)) (string, error) {
	// 构建 claude CLI 参数（与 p_claude.buildArgs 对齐）
	args := buildSdkArgs(cfg)
	env := buildSdkEnv(cfg)

	log.Printf("[sdk-exec] 启动 SDK 模式 claude 进程, dir=%s model=%s permission_mode=%s",
		cfg.WorkingDir, cfg.Model, cfg.PermissionMode)
	log.Printf("[sdk-exec] 完整参数: %v", args)
	// 记录鉴权相关环境变量是否存在（不打印值）
	if cfg.BaseURL != "" {
		log.Printf("[sdk-exec] ANTHROPIC_BASE_URL 已配置 (len=%d)", len(cfg.BaseURL))
	}
	if cfg.APIKey != "" {
		log.Printf("[sdk-exec] API key 已配置，将通过 ANTHROPIC_AUTH_TOKEN 传递 (len=%d)", len(cfg.APIKey))
	}
	if cfg.OAuthToken != "" {
		log.Printf("[sdk-exec] OAuth token 已配置 (len=%d)", len(cfg.OAuthToken))
	}

	if cfg.SessionID != "" {
		log.Printf("[sdk-exec] 尝试恢复 session_id=%s", cfg.SessionID)
	}
	if cfg.SettingsPath != "" {
		log.Printf("[sdk-exec] settings 路径=%s", cfg.SettingsPath)
	}

	// 创建 prompt 临时文件
	stdinFile, cleanupPromptFile, err := prepareSdkPromptStdinFile(cfg.Prompt)
	if err != nil {
		return "", fmt.Errorf("prepare prompt file failed: %w", err)
	}
	defer cleanupPromptFile()

	// 启动 claude 子进程（复用现有进程管理机制）
	result, err := startSdkClaude(ctx, args, cfg.WorkingDir, env, stdinFile)
	if err != nil {
		log.Printf("[sdk-exec] 启动失败: %v", err)
		return "", fmt.Errorf("claude sdk start failed: %w", err)
	}
	defer result.closeFn()
	log.Printf("[sdk-exec] 进程已启动, pid=%d", result.pid)

	// 进程启动回调
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

	sessionID := cfg.SessionID
	sessionExtracted := sessionID != ""
	lineCount := 0

	// 环形缓冲区：保留最近 N 行原始内容，用于异常退出时辅助诊断
	recentLines := make([]string, 0, sdkRecentLinesKeep)
	// 最后收到的 result 消息的解析数据，用于提取 API 错误信息
	var lastResultData map[string]any

	for {
		select {
		case <-ctx.Done():
			log.Printf("[sdk-exec] 上下文已取消，退出读取循环 (lineCount=%d)", lineCount)
			return sessionID, ctx.Err()
		case line, ok := <-result.lineCh:
			if !ok {
				goto doneReading
			}
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			lineCount++

			// 记录最近 N 行原始内容用于错误诊断
			if len(recentLines) >= sdkRecentLinesKeep {
				recentLines = recentLines[1:]
			}
			recentLines = append(recentLines, line)

			// 调试日志：输出每条 stdout 行（截断过长内容）
			debugLine := line
			if len(debugLine) > sdkDebugLineMaxLen {
				debugLine = debugLine[:sdkDebugLineMaxLen] + "..."
			}
			log.Printf("[sdk-exec] stdout#%d: %s", lineCount, debugLine)

			// 使用 p_claude 相同的 parseLine 逻辑转换
			msg := parseSDKLine(line)
			callback(msg)

			// 记录最后一条 result 消息，用于异常退出时提取 API 错误
			if msg.Type == "result" {
				lastResultData = msg.Data
			}

			if !sessionExtracted {
				if sid := extractSessionIDFromSDKLine(line); sid != "" {
					sessionID = sid
					log.Printf("[sdk-exec] 提取到 session_id=%s", sid)
					sessionExtracted = true
				}
			}
		}
	}
doneReading:

	log.Printf("[sdk-exec] 行通道关闭, 总行数=%d", lineCount)
	<-stderrDone

	exitCode, waitErr := result.waitFn()
	stderrSummary := strings.Join(stderrLines, "\n")
	log.Printf("[sdk-exec] 进程结束, exitCode=%d waitErr=%v lineCount=%d stderrLineCount=%d",
		exitCode, waitErr, lineCount, len(stderrLines))
	if stderrSummary != "" {
		log.Printf("[sdk-exec] stderr内容: %s", stderrSummary)
	}

	if waitErr != nil {
		if stderrSummary != "" {
			return sessionID, fmt.Errorf("claude 退出异常: %s (stderr: %s)", waitErr.Error(), stderrSummary)
		}
		return sessionID, fmt.Errorf("claude 退出异常: %w", waitErr)
	}
	if exitCode != 0 {
		if stderrSummary != "" {
			return sessionID, fmt.Errorf("claude 返回失败 (exit code %d): %s", exitCode, stderrSummary)
		}
		// 优先从 result 消息提取 API 错误（比通用错误信息更精确）
		if apiErr := extractResultError(lastResultData); apiErr != "" {
			log.Printf("[sdk-exec] 从 result 消息提取到 API 错误: %s", apiErr)
			return sessionID, fmt.Errorf("API 错误: %s", apiErr)
		}
		errMsg := fmt.Sprintf("claude 返回失败 (exit code %d)，无 stderr 输出。", exitCode)
		if lineCount == 0 {
			errMsg += " 未收到任何 stdout 输出，可能是 Claude CLI 启动失败或配置错误。"
		} else if len(recentLines) > 0 {
			// stderr 为空时，将最近几行 stdout 附到错误信息中以辅助诊断
			errMsg += fmt.Sprintf(" 最近 %d 行 stdout: %s", len(recentLines), strings.Join(recentLines, " | "))
		}
		return sessionID, errors.New(errMsg)
	}
	return sessionID, nil
}

// buildSdkArgs 构建 claude SDK 模式的命令行参数。
func buildSdkArgs(cfg RunConfig) []string {
	args := []string{}
	if cfg.SessionID != "" {
		args = append(args, "--resume", cfg.SessionID)
	}
	args = append(args,
		"-p",
		"--add-dir", cfg.WorkingDir,
		"--output-format", "stream-json",
		"--include-partial-messages",
		"--verbose",
	)

	// 权限模式：SDK 模式支持可配置的权限模式
	permissionMode := cfg.PermissionMode
	if permissionMode == "" {
		// 如果有前端 SSE 审批能力，使用默认模式（需审批）；否则 bypass
		if cfg.HasApprovalSink {
			permissionMode = PermissionModeDefault
		} else {
			permissionMode = PermissionModeBypass
		}
	}
	args = append(args, "--permission-mode", permissionMode)

	if cfg.Model != "" {
		args = append(args, "--model", cfg.Model)
	}
	if cfg.UserDataDir != "" {
		args = append(args, "--user-data-dir", cfg.UserDataDir)
	}
	if cfg.SettingsPath != "" {
		args = append(args, "--settings", cfg.SettingsPath)
	}
	if cfg.MaxTurns > 0 {
		args = append(args, "--max-turns", strconv.Itoa(cfg.MaxTurns))
	}
	return args
}

// sdkEnvFilterPrefixes Anthropic 相关的环境变量前缀，在构建子进程环境时需要清理。
var sdkEnvFilterPrefixes = []string{
	"ANTHROPIC_",
	"CLAUDE_CODE_",
}

// buildSdkEnv 构建 SDK 模式的环境变量。
// 先清理父进程中所有 Anthropic 相关环境变量（避免残留值干扰），再设置 SDK 配置值。
// 与传统 settings.json 行为保持一致：API key 通过 ANTHROPIC_AUTH_TOKEN 传递，
// 并显式清空 ANTHROPIC_API_KEY，避免 Claude CLI 误用 x-api-key 头格式。
func buildSdkEnv(cfg RunConfig) []string {
	// 从父进程环境继承非 Anthropic 相关的变量
	parentEnv := os.Environ()
	env := make([]string, 0, len(parentEnv)+8)
	for _, e := range parentEnv {
		skip := false
		for _, prefix := range sdkEnvFilterPrefixes {
			if strings.HasPrefix(strings.ToUpper(e), prefix) {
				skip = true
				break
			}
		}
		if !skip {
			env = append(env, e)
		}
	}

	// 设置 SDK 配置的环境变量
	if cfg.BaseURL != "" {
		env = append(env, "ANTHROPIC_BASE_URL="+cfg.BaseURL)
	}
	if cfg.APIKey != "" {
		env = append(env, "ANTHROPIC_AUTH_TOKEN="+cfg.APIKey)
		env = append(env, "ANTHROPIC_API_KEY=") // 显式置空
	}
	if cfg.OAuthToken != "" {
		env = append(env, "CLAUDE_CODE_OAUTH_TOKEN="+cfg.OAuthToken)
	}
	return env
}

// sdkProcessResult 进程启动结果，由各平台 startSdkClaude 实现返回。
type sdkProcessResult struct {
	lineCh   <-chan string       // stdout 行数据通道
	stderrCh <-chan string       // stderr 行数据通道
	pid      int                 // 进程 ID
	waitFn   func() (int, error) // 等待进程退出并返回退出码
	closeFn  func()              // 强制终止进程（含子进程清理）
}

// maxScanTokenSize bufio.Scanner 最大缓冲区大小（10MB）。
const maxScanTokenSize = 10 * 1024 * 1024

// sdkRecentLinesKeep 环形缓冲区保留的最近 stdout 行数，用于异常退出时的错误诊断。
const sdkRecentLinesKeep = 5

// sdkDebugLineMaxLen 单行调试日志的最大长度（截断超出部分）。
const sdkDebugLineMaxLen = 500

// parseSDKLine 解析单行 stream-json 为 StreamMessage。
// 与 p_claude.parseLine 逻辑对齐。
func parseSDKLine(line string) StreamMessage {
	msg := StreamMessage{RawJSON: line}
	var raw map[string]any
	if err := json.Unmarshal([]byte(line), &raw); err != nil {
		msg.Type = "raw_text"
		msg.Data = map[string]any{"text": line}
		if errJSON, e := json.Marshal(msg); e == nil {
			msg.RawJSON = string(errJSON)
		}
		return msg
	}
	msg.Type = cast.ToString(raw["type"])
	msg.Subtype = cast.ToString(raw["subtype"])
	msg.Data = raw
	return msg
}

// extractSessionIDFromSDKLine 从 stream-json 行提取 session_id。
func extractSessionIDFromSDKLine(line string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return ""
	}
	if cast.ToString(data["type"]) != "system" || cast.ToString(data["subtype"]) != "init" {
		return ""
	}
	return cast.ToString(data["session_id"])
}

// prepareSdkPromptStdinFile 创建 prompt 临时文件并返回 file handle 和清理函数。
func prepareSdkPromptStdinFile(prompt string) (*os.File, func(), error) {
	file, err := os.CreateTemp("", "dtool-sdk-prompt-*.txt")
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

// extractResultError 从 result 消息的解析数据中提取 API 错误信息。
// result 消息格式示例：
//
//	{"type":"result","subtype":"success","is_error":true,"api_error_status":400,
//	 "result":"API Error: 400 provider is not authorized for auth key\n",...}
func extractResultError(data map[string]any) string {
	if data == nil {
		return ""
	}
	isErr, _ := data["is_error"].(bool)
	if !isErr {
		return ""
	}
	// 优先使用 result 字段（Claude CLI 的 result 消息格式）
	if resultText, ok := data["result"].(string); ok && strings.TrimSpace(resultText) != "" {
		return strings.TrimSpace(resultText)
	}
	// 降级：检查 error 字段（assistant 消息格式）
	if errText, ok := data["error"].(string); ok && strings.TrimSpace(errText) != "" {
		return strings.TrimSpace(errText)
	}
	return ""
}

// BuildSdkCommandLine 根据配置构建完整的 claude SDK CLI 命令字符串（用于前端展示）。
func BuildSdkCommandLine(cfg RunConfig) string {
	var sb strings.Builder
	sb.WriteString("claude")
	if cfg.SessionID != "" {
		sb.WriteString(" --resume ")
		sb.WriteString(cfg.SessionID)
	}
	sb.WriteString(" -p < prompt-file")
	sb.WriteString(" --add-dir \"")
	sb.WriteString(cfg.WorkingDir)
	sb.WriteString("\"")
	sb.WriteString(" --output-format stream-json --include-partial-messages --verbose")

	permissionMode := cfg.PermissionMode
	if permissionMode == "" {
		if cfg.HasApprovalSink {
			permissionMode = PermissionModeDefault
		} else {
			permissionMode = PermissionModeBypass
		}
	}
	sb.WriteString(" --permission-mode ")
	sb.WriteString(permissionMode)

	if cfg.Model != "" {
		sb.WriteString(" --model ")
		sb.WriteString(cfg.Model)
	}
	if cfg.UserDataDir != "" {
		sb.WriteString(" --user-data-dir \"")
		sb.WriteString(cfg.UserDataDir)
		sb.WriteString("\"")
	}
	if cfg.SettingsPath != "" {
		sb.WriteString(" --settings \"")
		sb.WriteString(cfg.SettingsPath)
		sb.WriteString("\"")
	}
	if cfg.MaxTurns > 0 {
		sb.WriteString(" --max-turns ")
		sb.WriteString(strconv.Itoa(cfg.MaxTurns))
	}
	return sb.String()
}
