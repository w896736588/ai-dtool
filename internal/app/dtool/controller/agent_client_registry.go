package controller

import (
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"sync"
	"time"
)

// AgentClientInfo 内存中的客户端注册信息（替代 tbl_smart_link_client）
type AgentClientInfo struct {
	ClientID      string
	ClientName    string
	ClientVersion string
	HostName      string
	Os            string
	Arch          string
	UserName      string
	Status        define.SmartLinkClientStatus
	LastSeenTime  int64
	RegisterTime  int64
}

// AgentClientRegistry 全局客户端注册表
type AgentClientRegistry struct {
	mu      sync.RWMutex
	clients map[string]*AgentClientInfo // key = client_id
}

// GlobalClientRegistry 全局单例
var GlobalClientRegistry = &AgentClientRegistry{
	clients: make(map[string]*AgentClientInfo),
}

// Register 注册或更新客户端信息
func (r *AgentClientRegistry) Register(info *AgentClientInfo) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().Unix()
	info.LastSeenTime = now

	// 如果已有记录，保留 RegisterTime
	if existing, ok := r.clients[info.ClientID]; ok {
		info.RegisterTime = existing.RegisterTime
	} else {
		info.RegisterTime = now
	}

	r.clients[info.ClientID] = info
}

// Get 通过 clientID 查找已注册的客户端
func (r *AgentClientRegistry) Get(clientID string) (*AgentClientInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	info, ok := r.clients[clientID]
	return info, ok
}

// UpdateHeartbeat 更新心跳时间和状态
func (r *AgentClientRegistry) UpdateHeartbeat(clientID string, status define.SmartLinkClientStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, ok := r.clients[clientID]
	if !ok {
		return
	}
	info.LastSeenTime = time.Now().Unix()
	info.Status = status
}

// UpdateHelloInfo 通过 agent_hello 消息更新客户端系统信息，若不存在则自动注册
func (r *AgentClientRegistry) UpdateHelloInfo(clientID string, data define.AgentHelloData) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now().Unix()

	info, ok := r.clients[clientID]
	if !ok {
		// 自动注册：hello 消息携带的信息足够完成注册
		info = &AgentClientInfo{
			ClientID:     clientID,
			RegisterTime: now,
		}
		r.clients[clientID] = info
	}
	info.ClientVersion = data.ClientVersion
	info.HostName = data.Hostname
	info.ClientName = data.Hostname
	info.Os = data.Os
	info.Arch = data.Arch
	info.UserName = data.UserName
	info.LastSeenTime = now
}

// SetStatus 仅更新状态
func (r *AgentClientRegistry) SetStatus(clientID string, status define.SmartLinkClientStatus) {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, ok := r.clients[clientID]
	if !ok {
		return
	}
	info.LastSeenTime = time.Now().Unix()
	info.Status = status
}

// GetLatest 获取最近活跃的客户端（替代 ORDER BY last_seen_time DESC LIMIT 1）
func (r *AgentClientRegistry) GetLatest() *AgentClientInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var latest *AgentClientInfo
	for _, info := range r.clients {
		if latest == nil || info.LastSeenTime > latest.LastSeenTime {
			latest = info
		}
	}
	return latest
}

// SetOffline 标记客户端为离线
func (r *AgentClientRegistry) SetOffline(clientID string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	info, ok := r.clients[clientID]
	if !ok {
		return
	}
	info.Status = define.SmartLinkClientStatusOffline
}

// parseAgentHelloData 从 msg.Data 解析 AgentHelloData
func parseAgentHelloData(raw any) define.AgentHelloData {
	dataBytes, _ := json.Marshal(raw)
	var result define.AgentHelloData
	json.Unmarshal(dataBytes, &result)
	return result
}
