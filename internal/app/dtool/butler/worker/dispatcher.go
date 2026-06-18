package worker

import (
	"dev_tool/internal/app/dtool/common"
	"fmt"
	"strings"

	"github.com/w896736588/go-tool/gstool"
)

// TaskType 任务路由类型。
const (
	TaskTypeFC       = `fc`        // 简单任务 → FC 循环（文件操作等）
	TaskTypeAgentCli = `agent_cli` // 复杂任务 → Agent CLI（开发、重构等）
)

// DispatchResult 任务路由结果。
type DispatchResult struct {
	TaskType string // fc 或 agent_cli
}

// Dispatch 根据用户消息判断任务路由：简单→FC，复杂→Agent CLI。
// 使用 fc_model_id 的 AI 来判断任务复杂度。
// 当 agentCliId 为 0 时，始终返回 FC（无 Agent CLI 可用）。
func Dispatch(db *common.CSqlite, modelId int, userMessage string, agentCliId int) *DispatchResult {
	// 无 Agent CLI 配置 → 始终走 FC
	if agentCliId <= 0 {
		return &DispatchResult{TaskType: TaskTypeFC}
	}
	if modelId <= 0 {
		gstool.FmtPrintlnLogTime(`[butler-dispatch] 模型未配置，默认走 FC`)
		return &DispatchResult{TaskType: TaskTypeFC}
	}

	// 使用 AI 判断任务复杂度
	prompt := buildDispatchPrompt(userMessage)
	messages := []map[string]any{
		{`role`: `system`, `content`: dispatchSystemPrompt},
		{`role`: `user`, `content`: prompt},
	}
	content, _, _, err := db.AIChatByModelWithTools(modelId, messages, nil)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[butler-dispatch] AI 分类失败 %s，默认走 FC`, err.Error())
		return &DispatchResult{TaskType: TaskTypeFC}
	}

	taskType := parseDispatchResult(content)
	gstool.FmtPrintlnLogTime(`[butler-dispatch] 路由结果 taskType=%s userMessage=%s`, taskType, truncateForLog(userMessage, 80))
	return &DispatchResult{TaskType: taskType}
}

// dispatchSystemPrompt 任务路由的系统提示词。
const dispatchSystemPrompt = `你是一个任务分类器。根据用户的任务描述，判断应该使用哪种执行方式。

分类标准：
- fc：简单查询（查询 Git 分支/状态/配置等通过 HTTP API 可完成的操作）、文件操作（读取、创建、修改、删除文件）、日常对话。这些任务可通过 http_call 或文件工具完成。
- agent_cli：需要修改代码或执行本地命令的任务（切换分支、提交代码、开发新接口、重构代码、多文件改动、调试复杂问题、编写测试、实现新功能等）。这些任务需要 shell 命令执行能力或多步骤推理。

请只输出一个词：fc 或 agent_cli，不要输出任何其他内容。`

// buildDispatchPrompt 构建任务路由的用户提示词。
func buildDispatchPrompt(userMessage string) string {
	return fmt.Sprintf(`用户任务：%s

请判断此任务应该使用 fc 还是 agent_cli 执行。`, userMessage)
}

// parseDispatchResult 解析 AI 返回的路由结果，容错提取。
func parseDispatchResult(content string) string {
	text := strings.TrimSpace(strings.ToLower(content))
	// 尝试精确匹配
	if text == `agent_cli` || text == `agent-cli` || text == `agentcli` {
		return TaskTypeAgentCli
	}
	// 包含关键词
	if strings.Contains(text, `agent`) {
		return TaskTypeAgentCli
	}
	// 默认走 FC
	return TaskTypeFC
}
