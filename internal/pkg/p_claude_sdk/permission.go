package p_claude_sdk

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// =============================================================================
// 权限审批桥接：SDK 权限请求 → SSE 推送 → HTTP 响应 → SDK 返回
// =============================================================================
// 流程：
//   1. SDK 发起权限请求 → PermissionCallback
//   2. 生成 ApprovalRequest（含 request_id）→ SSE 推送到前端
//   3. 用户点击 允许/拒绝 → POST /api/agent/chat/approve → HandleApprovalResponse
//   4. PermissionCallback 从 channel 读取结果 → 返回 SDK
// =============================================================================

// pendingApprovals 等待审批的请求，key=requestID，value=response channel
var pendingApprovals sync.Map

// NewPermissionCallback 创建权限审批回调函数。
// 当 SDK 拦截到工具调用时，将请求推送到前端，阻塞等待前端审批响应。
// sendSse 用于将权限请求推送到前端 SSE 连接。
func NewPermissionCallback(chatID int64, sessionID string, sendSse func(msg StreamMessage)) func(ctx context.Context, toolName string, input interface{}) (bool, error) {
	return func(ctx context.Context, toolName string, input interface{}) (bool, error) {
		requestID := uuid.New().String()
		log.Printf("[sdk-permission] 权限审批请求: request_id=%s chat_id=%d tool=%s", requestID, chatID, toolName)

		// 1. 构建权限请求事件并推送到前端
		req := &ApprovalRequest{
			RequestID: requestID,
			ToolName:  toolName,
			Input:     input,
			SessionID: sessionID,
			ChatID:    chatID,
		}

		msg := BuildPermissionRequestEvent(req)
		sendSse(msg)

		// 2. 阻塞等待前端审批响应（带超时）
		responseCh := make(chan *ApprovalResponse, 1)
		pendingApprovals.Store(requestID, responseCh)
		defer func() {
			pendingApprovals.Delete(requestID)
		}()

		timeout := time.After(time.Duration(PermissionTimeoutSeconds) * time.Second)
		select {
		case resp := <-responseCh:
			log.Printf("[sdk-permission] 审批响应: request_id=%s approved=%v reason=%s",
				requestID, resp.Approved, resp.Reason)
			return resp.Approved, nil
		case <-ctx.Done():
			log.Printf("[sdk-permission] 上下文取消: request_id=%s err=%v", requestID, ctx.Err())
			return false, fmt.Errorf("上下文已取消: %w", ctx.Err())
		case <-timeout:
			log.Printf("[sdk-permission] 审批超时自动拒绝: request_id=%s tool=%s (timeout=%ds)",
				requestID, toolName, PermissionTimeoutSeconds)
			// 推送超时通知到前端（关闭弹窗）
			sendSse(StreamMessage{
				Type: "permission_timeout",
				Data: map[string]any{
					"request_id": requestID,
					"tool_name":  toolName,
				},
			})
			return false, fmt.Errorf("权限审批超时（%d秒）: %s", PermissionTimeoutSeconds, toolName)
		}
	}
}

// HandleApprovalResponse 处理前端审批响应。
// 由 Controller 层 HTTP API 调用，将响应写入对应的等待 channel。
func HandleApprovalResponse(resp *ApprovalResponse) error {
	if resp == nil || resp.RequestID == "" {
		return fmt.Errorf("无效的审批响应")
	}

	val, ok := pendingApprovals.LoadAndDelete(resp.RequestID)
	if !ok {
		return fmt.Errorf("无效的审批请求 ID（可能已超时或已处理）: %s", resp.RequestID)
	}

	ch, ok := val.(chan *ApprovalResponse)
	if !ok {
		return fmt.Errorf("内部错误：审批响应通道类型异常")
	}

	// 非阻塞写入，避免 channel 已关闭导致 panic
	select {
	case ch <- resp:
		return nil
	default:
		return fmt.Errorf("审批响应通道已满或已关闭: %s", resp.RequestID)
	}
}

// CleanupPendingApprovals 清理指定 chatID 的所有待审批请求（对话停止时调用）。
func CleanupPendingApprovals(chatID int64) {
	var cleanedCount int
	pendingApprovals.Range(func(key, value interface{}) bool {
		ch, ok := value.(chan *ApprovalResponse)
		if !ok {
			pendingApprovals.Delete(key)
			cleanedCount++
			return true
		}
		// 尝试发送拒绝响应，避免 PermissionCallback 永远阻塞
		select {
		case ch <- &ApprovalResponse{
			RequestID: fmt.Sprintf("%v", key),
			Approved:  false,
			Reason:    "对话已停止",
		}:
		default:
		}
		pendingApprovals.Delete(key)
		cleanedCount++
		return true
	})

	if cleanedCount > 0 {
		log.Printf("[sdk-permission] 清理了 %d 个待审批请求 (chat_id=%d)", cleanedCount, chatID)
	}
}
