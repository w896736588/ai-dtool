package butler

import (
	"sync"
	"time"
)

// Session 会话激活态，记录最后活跃时间，用于 30min 无消息自动休眠。
type Session struct {
	ConversationId string
	LastActive     time.Time
	Active         bool
}

// SessionManager 会话激活态管理器，内存维护，并发安全。
type SessionManager struct {
	mu          sync.RWMutex
	sessions    map[string]*Session
	timeout     time.Duration // 激活态超时时长
	lastSleepAt map[string]time.Time
}

// NewSessionManager 创建会话管理器，timeout 为激活态超时（如 30min）。
func NewSessionManager(timeout time.Duration) *SessionManager {
	return &SessionManager{
		sessions:    make(map[string]*Session),
		lastSleepAt: make(map[string]time.Time),
		timeout:     timeout,
	}
}

// Activate 激活会话并刷新最后活跃时间，返回是否为新激活（之前非激活）。
func (m *SessionManager) Activate(conversationId string) bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	s, exist := m.sessions[conversationId]
	justActivated := false
	if !exist || !s.Active {
		justActivated = true
	}
	if !exist {
		s = &Session{ConversationId: conversationId}
		m.sessions[conversationId] = s
	}
	s.Active = true
	s.LastActive = time.Now()
	return justActivated
}

// IsActive 返回会话是否处于激活态。
func (m *SessionManager) IsActive(conversationId string) bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	s, exist := m.sessions[conversationId]
	return exist && s.Active
}

// TouchLastActive 刷新最后活跃时间（收到消息时调用）。
func (m *SessionManager) TouchLastActive(conversationId string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if s, exist := m.sessions[conversationId]; exist {
		s.LastActive = time.Now()
	}
}

// CheckTimeout 巡检超时会话，返回刚超时（需休眠）的会话列表并标记为非激活。
// 由管家定时调用（如每 1min）。
func (m *SessionManager) CheckTimeout() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	now := time.Now()
	timedOut := make([]string, 0)
	for id, s := range m.sessions {
		if !s.Active {
			continue
		}
		if now.Sub(s.LastActive) >= m.timeout {
			s.Active = false
			m.lastSleepAt[id] = now
			timedOut = append(timedOut, id)
		}
	}
	return timedOut
}
