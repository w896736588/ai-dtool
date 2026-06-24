package worker

import (
	"dev_tool/internal/app/dtool/common"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

// FCIntermediateMessage FC 循环中间消息记录，用于持久化到历史数据库。
// 每次工具调用产生的 assistant 消息（含 tool_calls）和 tool 结果消息都需要保存，
// 以便后续对话轮次能还原完整上下文，避免 AI"失忆"重新查询。
type FCIntermediateMessage struct {
	Role       string // assistant 或 tool
	Content    string // 消息内容文本
	ToolCalls  string // assistant 消息的 tool_calls JSON（tool 消息为空）
	ToolCallId string // tool 消息的调用 ID（assistant 消息为空）
}

// FCLoopResult FC 循环执行结果。
type FCLoopResult struct {
	Content       string                  // 最终 AI 回复文本
	Success       bool                    // 任务是否成功完成
	ToolUsed      []string                // 使用过的工具名称列表
	StepsRun      []string                // 执行过的步骤文件路径（去重）
	StepFilesRead []string                // 计划/检索阶段通过 file_read 读取的步骤 .md 文件路径（去重）
	LLMCalls      int                     // LLM 累计调用次数
	InputTokens   int                     // 累计输入 token 数
	OutputTokens  int                     // 累计输出 token 数
	CacheTokens   int                     // 累计缓存命中 token 数
	FCMessages    []FCIntermediateMessage // FC 循环中间消息（assistant 含 tool_calls + tool 结果），供持久化到历史
}

// RunFCLoop 执行 Function Calling 循环。
// 反复调用 AI，执行工具调用，直到 AI 返回最终文本回复（不再调用工具）或达到 maxLoop 迭代次数。
// maxLoop 为 0 或负数时取默认值 10。
// historyMessages 现为 []map[string]any 格式，支持包含 FC 中间消息（tool_calls/tool_call_id）的还原。
func RunFCLoop(db *common.CSqlite, modelId int, systemPrompt string, historyMessages []map[string]any, userMessage string, maxLoop int) *FCLoopResult {
	if maxLoop <= 0 {
		maxLoop = 10
	}
	// 构建 messages 列表
	messages := buildFCMessages(systemPrompt, historyMessages, userMessage)
	// 获取工具定义
	tools := ToolDefinitions()
	toolsUsed := make([]string, 0)
	stepsRun := make([]string, 0)
	stepFilesRead := make([]string, 0)
	stepsRunSet := make(map[string]bool)
	stepFilesReadSet := make(map[string]bool)
	fcMessages := make([]FCIntermediateMessage, 0)
	var totalCalls, totalInput, totalOutput, totalCache int

	for i := 0; i < maxLoop; i++ {
		// 调用 AI（非流式，需解析完整 tool_calls）
		content, toolCalls, usage, _, err := db.AIChatByModelWithTools(modelId, messages, tools)
		totalCalls++
		if usage != nil {
			totalInput += usage.InputTokens
			totalOutput += usage.OutputTokens
			totalCache += usage.CacheReadInputTokens
		}
		if err != nil {
			gstool.FmtPrintlnLogTime(`[butler-fc] AI 请求失败 %s`, err.Error())
			return &FCLoopResult{Content: fmt.Sprintf(`任务执行失败：%s`, err.Error()), Success: false, ToolUsed: toolsUsed, StepsRun: stepsRun, StepFilesRead: stepFilesRead, LLMCalls: totalCalls, InputTokens: totalInput, OutputTokens: totalOutput, CacheTokens: totalCache, FCMessages: fcMessages}
		}

		// 没有工具调用 → AI 已给出最终回复
		if len(toolCalls) == 0 {
			return &FCLoopResult{Content: content, Success: true, ToolUsed: toolsUsed, StepsRun: stepsRun, StepFilesRead: stepFilesRead, LLMCalls: totalCalls, InputTokens: totalInput, OutputTokens: totalOutput, CacheTokens: totalCache, FCMessages: fcMessages}
		}

		// 将 tool_calls 序列化为 JSON 字符串，用于持久化
		toolCallsJSON, _ := json.Marshal(toolCalls)

		// 记录 assistant 消息（含 tool_calls）
		assistantMsg := map[string]any{
			`role`:       `assistant`,
			`content`:    content,
			`tool_calls`: toolCalls,
		}
		messages = append(messages, assistantMsg)
		// 收集中间消息：assistant 含 tool_calls
		fcMessages = append(fcMessages, FCIntermediateMessage{
			Role:      `assistant`,
			Content:   content,
			ToolCalls: string(toolCallsJSON),
		})

		// 逐个执行工具调用
		for _, tc := range toolCalls {
			tcMap, ok := tc.(map[string]any)
			if !ok {
				continue
			}
			callID := cast.ToString(tcMap[`id`])
			fnMap, _ := tcMap[`function`].(map[string]any)
			fnName := cast.ToString(fnMap[`name`])
			fnArgs := cast.ToString(fnMap[`arguments`])

			// 提取路径用于结果追踪
			stepPath := extractPathFromArgs(fnArgs)
			if fnName == ToolRunScript && stepPath != `` && !stepsRunSet[stepPath] {
				stepsRunSet[stepPath] = true
				stepsRun = append(stepsRun, stepPath)
			}
			// 追踪 file_read 读取的步骤文件（供计划→执行阶段上下文传递）
			if fnName == ToolFileRead && stepPath != `` && strings.Contains(stepPath, `/step/`) && strings.HasSuffix(stepPath, `.md`) && !stepFilesReadSet[stepPath] {
				stepFilesReadSet[stepPath] = true
				stepFilesRead = append(stepFilesRead, stepPath)
			}

			gstool.FmtPrintlnLogTime(`[butler-fc] 执行工具 %s(%s)`, fnName, truncateForLog(fnArgs, 100))
			result := ExecuteTool(fnName, fnArgs)
			toolsUsed = append(toolsUsed, fnName)
			gstool.FmtPrintlnLogTime(`[butler-fc] 工具结果 %s → %s`, fnName, truncateForLog(result, 200))

			// 添加工具结果消息
			messages = append(messages, map[string]any{
				`role`:         `tool`,
				`tool_call_id`: callID,
				`content`:      result,
			})
			// 收集中间消息：tool 结果
			fcMessages = append(fcMessages, FCIntermediateMessage{
				Role:       `tool`,
				Content:    result,
				ToolCallId: callID,
			})
		}
	}

	// 超过最大迭代次数
	gstool.FmtPrintlnLogTime(`[butler-fc] FC 循环超过最大迭代次数 %d`, maxLoop)
	return &FCLoopResult{Content: fmt.Sprintf(`任务执行超时：已达到最大工具调用次数 %d`, maxLoop), Success: false, ToolUsed: toolsUsed, StepsRun: stepsRun, StepFilesRead: stepFilesRead, LLMCalls: totalCalls, InputTokens: totalInput, OutputTokens: totalOutput, CacheTokens: totalCache, FCMessages: fcMessages}
}

// buildFCMessages 构建 FC 循环的初始 messages 列表。
// historyMessages 为 []map[string]any 格式，支持包含 FC 中间消息（tool_calls/tool_call_id）。
func buildFCMessages(systemPrompt string, historyMessages []map[string]any, userMessage string) []map[string]any {
	messages := make([]map[string]any, 0, len(historyMessages)+2)
	// system prompt
	messages = append(messages, map[string]any{
		`role`:    `system`,
		`content`: systemPrompt,
	})
	// 历史消息（直接拷贝，保留 tool_calls/tool_call_id 等字段）
	for _, msg := range historyMessages {
		messages = append(messages, msg)
	}
	// 当前用户消息
	messages = append(messages, map[string]any{
		`role`:    `user`,
		`content`: userMessage,
	})
	return messages
}

// truncateForLog 截断字符串用于日志输出。
func truncateForLog(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + `...`
}

// extractPathFromArgs 从 JSON 格式的工具参数中提取 path 字段。
func extractPathFromArgs(argsJSON string) string {
	var args map[string]any
	if err := json.Unmarshal([]byte(argsJSON), &args); err != nil {
		return ""
	}
	return cast.ToString(args[`path`])
}
