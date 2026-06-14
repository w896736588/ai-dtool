package p_claude_sdk

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// =============================================================================
// SessionManager：管理 claude-agent-sdk-go Client 生命周期
// =============================================================================
// 设计原则：
//   - 每个 chatID 独占一个 Client（避免并发 Query 冲突，设计评审建议）
//   - Client 内部维护会话上下文，无需每轮重启进程
//   - 支持优雅关闭和 context 取消
//   - 包含过期清理机制防止 Client 泄漏
// =============================================================================

const (
	// clientIdleTimeout Client 空闲超时（30 分钟），超时后自动关闭释放资源
	clientIdleTimeout = 30 * time.Minute
	// clientCleanupInterval 定期清理间隔（10 分钟）
	clientCleanupInterval = 10 * time.Minute
)

// sdkClientEntry 单个 SDK Client 的管理条目。
type sdkClientEntry struct {
	mu sync.Mutex // 保护对 Client 的并发访问（Client 非线程安全）

	// Client 由具体的 SDK 实现提供，此处使用 interface{} 避免硬依赖
	// 实际类型为 claude-agent-sdk-go 的 *claude.Client
	Client interface{}

	// Cancel 取消当前 Client 关联的 context
	Cancel   context.CancelFunc
	ChatID   int64     // 关联的对话 ID
	LastUsed time.Time // 最后使用时间（用于空闲清理）

	// ApprovalCh 权限审批响应通道，由 permission.go 管理
	ApprovalCh chan *ApprovalResponse

	// running 标记当前是否正在执行 Query
	running bool
}

// SessionManager 管理活跃的 SDK Client 实例。
// 所有方法线程安全。
type SessionManager struct {
	mu      sync.RWMutex
	clients map[int64]*sdkClientEntry // key: chatID
}

// NewSessionManager 创建 SessionManager 并启动后台清理 goroutine。
func NewSessionManager() *SessionManager {
	sm := &SessionManager{
		clients: make(map[int64]*sdkClientEntry),
	}
	// 启动定期清理 goroutine
	go sm.cleanupLoop()
	return sm
}

// GetClient 获取指定 chatID 的 Client 条目（不创建新的）。
// 返回 nil 表示不存在或已过期。
func (sm *SessionManager) GetClient(chatID int64) *sdkClientEntry {
	sm.mu.RLock()
	entry, ok := sm.clients[chatID]
	sm.mu.RUnlock()
	if !ok {
		return nil
	}
	// 使用 entry 自身的互斥锁保护 LastUsed 写入，与 cleanupExpiredClients 的读取保持一致
	entry.mu.Lock()
	entry.LastUsed = time.Now()
	entry.mu.Unlock()
	return entry
}

// StoreClient 存储一个新的 Client 条目。
func (sm *SessionManager) StoreClient(chatID int64, entry *sdkClientEntry) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	entry.ChatID = chatID
	entry.LastUsed = time.Now()
	sm.clients[chatID] = entry
}

// RemoveClient 移除并关闭指定 chatID 的 Client。
func (sm *SessionManager) RemoveClient(chatID int64) error {
	sm.mu.Lock()
	entry, ok := sm.clients[chatID]
	if ok {
		delete(sm.clients, chatID)
	}
	sm.mu.Unlock()

	if !ok {
		return nil
	}

	return sm.closeClientEntry(entry)
}

// CloseAll 关闭所有 Client（服务关闭时调用）。
func (sm *SessionManager) CloseAll() error {
	sm.mu.Lock()
	entries := make([]*sdkClientEntry, 0, len(sm.clients))
	for chatID, entry := range sm.clients {
		entries = append(entries, entry)
		delete(sm.clients, chatID)
	}
	sm.mu.Unlock()

	var lastErr error
	for _, entry := range entries {
		if err := sm.closeClientEntry(entry); err != nil {
			log.Printf("[sdk-session] Close client chat_id=%d failed: %v", entry.ChatID, err)
			lastErr = err
		}
	}
	return lastErr
}

// closeClientEntry 关闭单个 Client 条目。
func (sm *SessionManager) closeClientEntry(entry *sdkClientEntry) error {
	entry.mu.Lock()
	defer entry.mu.Unlock()

	// 取消 context
	if entry.Cancel != nil {
		entry.Cancel()
		entry.Cancel = nil
	}

	// 关闭 Client
	if entry.Client != nil {
		if closer, ok := entry.Client.(interface{ Close() error }); ok {
			if err := closer.Close(); err != nil {
				return fmt.Errorf("close sdk client failed (chat_id=%d): %w", entry.ChatID, err)
			}
		}
	}

	log.Printf("[sdk-session] Client closed, chat_id=%d", entry.ChatID)
	return nil
}

// cleanupLoop 定期清理过期的 Client 条目。
func (sm *SessionManager) cleanupLoop() {
	ticker := time.NewTicker(clientCleanupInterval)
	defer ticker.Stop()

	for range ticker.C {
		sm.cleanupExpiredClients()
	}
}

// cleanupExpiredClients 清理空闲超时的 Client。
func (sm *SessionManager) cleanupExpiredClients() {
	now := time.Now()
	sm.mu.Lock()

	// 收集过期的条目（chatID + entry 引用），必须在锁内完成以便安全删除
	type expiredEntry struct {
		chatID int64
		entry  *sdkClientEntry
	}
	var expired []expiredEntry

	for chatID, entry := range sm.clients {
		entry.mu.Lock()
		isRunning := entry.running
		lastUsed := entry.LastUsed
		entry.mu.Unlock()

		if isRunning {
			continue
		}
		if now.Sub(lastUsed) > clientIdleTimeout {
			expired = append(expired, expiredEntry{chatID: chatID, entry: entry})
		}
	}

	// 从 map 中删除
	for _, e := range expired {
		delete(sm.clients, e.chatID)
	}
	sm.mu.Unlock()

	// 释放锁后逐个关闭（直接调用 closeClientEntry，避免 RemoveClient 二次查找 map 失败）
	for _, e := range expired {
		if err := sm.closeClientEntry(e.entry); err != nil {
			log.Printf("[sdk-session] Close expired client chat_id=%d failed: %v", e.chatID, err)
		}
	}

	if len(expired) > 0 {
		log.Printf("[sdk-session] Cleanup %d expired clients", len(expired))
	}
}

// MarkRunning 标记 Client 为运行中，防止空闲清理。
func (entry *sdkClientEntry) MarkRunning(running bool) {
	entry.mu.Lock()
	defer entry.mu.Unlock()
	entry.running = running
	entry.LastUsed = time.Now()
}

// IsRunning 检查 Client 是否正在运行。
func (entry *sdkClientEntry) IsRunning() bool {
	entry.mu.Lock()
	defer entry.mu.Unlock()
	return entry.running
}

// 全局 SessionManager 实例
var globalSessionMgr = NewSessionManager()

// GetSessionManager 获取全局 SessionManager。
func GetSessionManager() *SessionManager {
	return globalSessionMgr
}
