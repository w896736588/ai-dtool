package p_claude_sdk

import (
	"encoding/json"
	"fmt"
)

// =============================================================================
// 消息格式转换器：将 claude-agent-sdk-go 的 Message 转换为前端兼容的 StreamMessage
// =============================================================================
// 设计目标：
//   1. 保持与 p_claude stream-json 输出格式完全一致，前端无需修改现有渲染逻辑
//   2. 新增 permission_request / hook_event 类型仅需前端增加对应分支处理
// =============================================================================

// ConvertSDKMessage 将 SDK Message 转换为前端兼容的 StreamMessage。
// 将 SDK 内部 + 字节序列化信息一律归入 RawJSON 字段，维持与直接解析 CLI stdout 完全相同的形态。
func ConvertSDKMessage(msgType string, msgData map[string]any, sessionID string) StreamMessage {
	// 将类型注入 data 中，前端 parser 依赖 type 字段识别消息类型
	if msgData == nil {
		msgData = make(map[string]any)
	}
	msgData["type"] = msgType
	if sessionID != "" {
		msgData["session_id"] = sessionID
	}

	rawJSON, err := json.Marshal(msgData)
	if err != nil {
		return StreamMessage{
			Type: "error",
			Data: map[string]any{"text": fmt.Sprintf("消息序列化失败: %v", err)},
		}
	}

	msg := StreamMessage{
		Type:    msgType,
		RawJSON: string(rawJSON),
		Data:    msgData,
	}

	// 提取 subtype（如果存在）
	if subtype, ok := msgData["subtype"]; ok {
		if subtypeStr, ok := subtype.(string); ok {
			msg.Subtype = subtypeStr
		}
	}
	return msg
}

// ConvertAssistantMessage 转换 SDK AssistantMessage 为 StreamMessage。
// SDK 的 AssistantMessage 包含 ContentBlock 列表，需要映射为 stream-json 格式。
func ConvertAssistantMessage(modelName string, contentBlocks []map[string]any, sessionID string) StreamMessage {
	data := map[string]any{
		"type":    "assistant",
		"message": map[string]any{"content": contentBlocks},
	}
	if modelName != "" {
		data["message"].(map[string]any)["model"] = modelName
	}
	if sessionID != "" {
		data["session_id"] = sessionID
	}

	rawJSON, _ := json.Marshal(data)
	return StreamMessage{
		Type:    "assistant",
		RawJSON: string(rawJSON),
		Data:    data,
	}
}

// ConvertTextBlock 将 SDK TextBlock 转换为 stream-json content 格式。
func ConvertTextBlock(text string) map[string]any {
	return map[string]any{
		"type": "text",
		"text": text,
	}
}

// ConvertToolUseBlock 将 SDK ToolUseBlock 转换为 stream-json content 格式。
func ConvertToolUseBlock(toolUseID, toolName string, input map[string]any) map[string]any {
	return map[string]any{
		"type":  "tool_use",
		"id":    toolUseID,
		"name":  toolName,
		"input": input,
	}
}

// ConvertToolResultBlock 将 SDK ToolResultBlock 转换为 stream-json content 格式。
func ConvertToolResultBlock(toolUseID, content string, isError bool) map[string]any {
	result := map[string]any{
		"type":        "tool_result",
		"tool_use_id": toolUseID,
		"content":     content,
	}
	if isError {
		result["is_error"] = true
	}
	return result
}

// ConvertSystemInitMessage 转换系统初始化消息。
func ConvertSystemInitMessage(sessionID, modelName, workingDir string, isResume bool, continueAt int64) StreamMessage {
	data := map[string]any{
		"type":        "system",
		"subtype":     "init",
		"session_id":  sessionID,
		"model":       modelName,
		"cwd":         workingDir,
		"is_resume":   isResume,
		"continue_at": continueAt,
	}

	rawJSON, _ := json.Marshal(data)
	return StreamMessage{
		Type:    "system",
		Subtype: "init",
		RawJSON: string(rawJSON),
		Data:    data,
	}
}

// ConvertResultMessage 转换最终结果消息。
func ConvertResultMessage(resultText string, sessionID string, costUSD float64) StreamMessage {
	data := map[string]any{
		"type":   "result",
		"result": resultText,
	}
	if sessionID != "" {
		data["session_id"] = sessionID
	}
	if costUSD > 0 {
		data["cost_usd"] = costUSD
	}

	rawJSON, _ := json.Marshal(data)
	return StreamMessage{
		Type:    "result",
		RawJSON: string(rawJSON),
		Data:    data,
	}
}

// ConvertUserMessage 转换用户消息回显。
func ConvertUserMessage(prompt string, sessionID string) StreamMessage {
	data := map[string]any{
		"type":    "user",
		"message": map[string]any{"content": prompt},
	}
	if sessionID != "" {
		data["session_id"] = sessionID
	}

	rawJSON, _ := json.Marshal(data)
	return StreamMessage{
		Type:    "user",
		RawJSON: string(rawJSON),
		Data:    data,
	}
}

// BuildCommandEvent 构建命令展示事件（推送前端显示完整命令）。
func BuildCommandEvent(cmdLine, prompt, cliType string) StreamMessage {
	data := map[string]any{
		"type":     "system",
		"subtype":  "command",
		"cli_type": cliType,
		"cmd_line": cmdLine,
		"text":     prompt,
	}

	rawJSON, _ := json.Marshal(data)
	return StreamMessage{
		Type:    "system",
		Subtype: "command",
		RawJSON: string(rawJSON),
		Data:    data,
	}
}

// BuildPermissionRequestEvent 构建权限请求 SSE 事件。
func BuildPermissionRequestEvent(req *ApprovalRequest) StreamMessage {
	reqJSON, _ := json.Marshal(req)
	rawJSON := string(reqJSON)

	return StreamMessage{
		Type:    "permission_request",
		RawJSON: rawJSON,
		Data:    map[string]any{"raw": rawJSON},
	}
}

// BuildHookEventMessage 构建 Hook 事件 SSE 消息。
func BuildHookEventMessage(event *HookEvent) StreamMessage {
	eventJSON, _ := json.Marshal(event)
	rawJSON := string(eventJSON)

	return StreamMessage{
		Type:    "hook_event",
		RawJSON: rawJSON,
		Data:    map[string]any{"raw": rawJSON},
	}
}

// BuildErrorEvent 构建错误 SSE 事件。
func BuildErrorEvent(errText string) StreamMessage {
	data := map[string]any{
		"type": "error",
		"text": errText,
	}

	rawJSON, _ := json.Marshal(data)
	return StreamMessage{
		Type:    "error",
		RawJSON: string(rawJSON),
		Data:    data,
	}
}
