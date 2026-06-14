package p_claude_sdk

import (
	"log"
)

// =============================================================================
// Hook 事件桥接：SDK Hook → SSE 事件推送前端
// =============================================================================
// 将 SDK 的 Hook 生命周期事件转换为前端可识别的 hook_event 类型 SSE 消息。
// 前端 chat_parser.js 已将 hook_started/hook_response subtype 映射为 system_hook 类型，
// 此处新增 hook_event 类型作为补充，可展示更详细的 Hook 信息。
// =============================================================================

// NewPreToolUseHook 创建 PreToolUse Hook 回调。
// 在工具执行前触发，返回 {"continue": true} 允许继续执行。
func NewPreToolUseHook(chatID int64, sessionID string, sendSse func(msg StreamMessage)) func(toolName string, input map[string]any) map[string]any {
	return func(toolName string, input map[string]any) map[string]any {
		event := &HookEvent{
			HookType:  "PreToolUse",
			ToolName:  toolName,
			Input:     input,
			SessionID: sessionID,
			ChatID:    chatID,
		}

		msg := BuildHookEventMessage(event)
		sendSse(msg)

		log.Printf("[sdk-hook] PreToolUse: chat_id=%d tool=%s", chatID, toolName)
		return map[string]interface{}{"continue": true}
	}
}

// NewPostToolUseHook 创建 PostToolUse Hook 回调。
// 在工具执行后触发，可用于记录工具调用结果。
func NewPostToolUseHook(chatID int64, sessionID string, sendSse func(msg StreamMessage)) func(toolName string, input, output map[string]any) map[string]any {
	return func(toolName string, input, output map[string]any) map[string]any {
		event := &HookEvent{
			HookType:  "PostToolUse",
			ToolName:  toolName,
			Input:     input,
			Output:    output,
			SessionID: sessionID,
			ChatID:    chatID,
		}

		msg := BuildHookEventMessage(event)
		sendSse(msg)

		log.Printf("[sdk-hook] PostToolUse: chat_id=%d tool=%s", chatID, toolName)
		return map[string]interface{}{"continue": true}
	}
}

// NewSystemNotificationHook 创建通用系统通知 Hook。
// 用于推送 MCP 状态变更、模型切换等系统事件。
func NewSystemNotificationHook(chatID int64, sessionID string, sendSse func(msg StreamMessage), hookType string) func(data map[string]any) {
	return func(data map[string]any) {
		event := &HookEvent{
			HookType:  hookType,
			SessionID: sessionID,
			ChatID:    chatID,
			Input:     data,
		}

		msg := BuildHookEventMessage(event)
		sendSse(msg)

		log.Printf("[sdk-hook] %s: chat_id=%d", hookType, chatID)
	}
}
