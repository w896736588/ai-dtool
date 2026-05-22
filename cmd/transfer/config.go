package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
)

// 默认配置文件路径
const configFileName = "transfer_config.json"

// ProxyRule 代理转发规则
type ProxyRule struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Port      int    `json:"port"`
	Host      string `json:"host"`
	TargetURL string `json:"target_url"`
	Format    string `json:"format"` // openai 或 anthropic
}

// ConfigStore 规则持久化存储，操作后自动保存到文件
type ConfigStore struct {
	mu       sync.RWMutex
	rules    []*ProxyRule
	filePath string
}

// NewConfigStore 创建配置存储实例，从文件加载已有规则
func NewConfigStore(filePath string) (*ConfigStore, error) {
	cs := &ConfigStore{
		rules:    make([]*ProxyRule, 0),
		filePath: filePath,
	}
	if err := cs.load(); err != nil {
		return nil, err
	}
	return cs, nil
}

// load 从 JSON 文件加载规则
func (cs *ConfigStore) load() error {
	data, err := os.ReadFile(cs.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil // 文件不存在，使用空列表
		}
		return fmt.Errorf("读取配置文件失败: %w", err)
	}
	if len(data) == 0 {
		return nil
	}
	if err := json.Unmarshal(data, &cs.rules); err != nil {
		return fmt.Errorf("解析配置文件失败: %w", err)
	}
	return nil
}

// save 将规则列表保存到 JSON 文件
func (cs *ConfigStore) save() error {
	data, err := json.MarshalIndent(cs.rules, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %w", err)
	}
	if err := os.WriteFile(cs.filePath, data, 0644); err != nil {
		return fmt.Errorf("写入配置文件失败: %w", err)
	}
	return nil
}

// List 返回所有规则
func (cs *ConfigStore) List() []*ProxyRule {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	result := make([]*ProxyRule, len(cs.rules))
	copy(result, cs.rules)
	return result
}

// Get 按 ID 获取规则
func (cs *ConfigStore) Get(id string) *ProxyRule {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	for _, r := range cs.rules {
		if r.ID == id {
			return r
		}
	}
	return nil
}

// Add 新增规则，自动分配 ID，保存到文件
func (cs *ConfigStore) Add(rule *ProxyRule) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	rule.ID = uuid.New().String()
	if rule.Host == "" {
		rule.Host = "localhost"
	}
	cs.rules = append(cs.rules, rule)
	return cs.save()
}

// Update 更新规则，保存到文件
func (cs *ConfigStore) Update(rule *ProxyRule) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	if rule.Host == "" {
		rule.Host = "localhost"
	}

	for i, r := range cs.rules {
		if r.ID == rule.ID {
			cs.rules[i] = rule
			return cs.save()
		}
	}
	return fmt.Errorf("规则不存在: %s", rule.ID)
}

// Delete 按 ID 删除规则，保存到文件
func (cs *ConfigStore) Delete(id string) error {
	cs.mu.Lock()
	defer cs.mu.Unlock()

	for i, r := range cs.rules {
		if r.ID == id {
			cs.rules = append(cs.rules[:i], cs.rules[i+1:]...)
			return cs.save()
		}
	}
	return fmt.Errorf("规则不存在: %s", id)
}

// FindByPort 按端口查找规则
func (cs *ConfigStore) FindByPort(port int) *ProxyRule {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	for _, r := range cs.rules {
		if r.Port == port {
			return r
		}
	}
	return nil
}

// runningServer 运行中的代理服务信息
type runningServer struct {
	rule   *ProxyRule
	server *http.Server
}

// ServerManager 动态管理代理 HTTP 服务的启停
type ServerManager struct {
	mu        sync.Mutex
	servers   map[string]*runningServer // ruleID -> server
	store     *Store
	adminPort int
}

// NewServerManager 创建服务管理器
func NewServerManager(store *Store, adminPort int) *ServerManager {
	return &ServerManager{
		servers:   make(map[string]*runningServer),
		store:     store,
		adminPort: adminPort,
	}
}

// StartRule 根据规则启动代理 HTTP 服务
func (sm *ServerManager) StartRule(rule *ProxyRule) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	// 检查是否已运行
	if _, exists := sm.servers[rule.ID]; exists {
		return fmt.Errorf("规则 %s 的代理服务已在运行", rule.ID)
	}

	// 检查端口是否已被管理端口占用
	if rule.Port == sm.adminPort {
		return fmt.Errorf("端口 %d 已被管理服务占用", rule.Port)
	}

	// 检查端口是否已被其他规则占用
	for _, rs := range sm.servers {
		if rs.rule.Port == rule.Port {
			return fmt.Errorf("端口 %d 已被规则 %s 占用", rule.Port, rs.rule.Name)
		}
	}

	proxy, err := NewProxy(rule.TargetURL, sm.store, rule.Format, rule.ID)
	if err != nil {
		return fmt.Errorf("创建代理失败: %w", err)
	}

	addr := fmt.Sprintf("%s:%d", rule.Host, rule.Port)
	srv := &http.Server{
		Addr:         addr,
		Handler:      proxy,
		ReadTimeout:  0,
		WriteTimeout: 0,
	}

	rs := &runningServer{
		rule:   rule,
		server: srv,
	}
	sm.servers[rule.ID] = rs

	// 在后台 goroutine 中启动服务
	go func() {
		log.Printf("[代理启动] %s (端口 %d) -> %s [%s]", rule.Name, rule.Port, rule.TargetURL, rule.Format)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[代理错误] %s: %s", rule.Name, err.Error())
		}
	}()

	return nil
}

// StopRule 停止指定规则的代理服务
func (sm *ServerManager) StopRule(ruleID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	rs, exists := sm.servers[ruleID]
	if !exists {
		return fmt.Errorf("规则 %s 的代理服务未运行", ruleID)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rs.server.Shutdown(ctx); err != nil {
		log.Printf("[代理停止] %s: 强制关闭 (%s)", rs.rule.Name, err.Error())
	} else {
		log.Printf("[代理停止] %s (端口 %d)", rs.rule.Name, rs.rule.Port)
	}

	delete(sm.servers, ruleID)
	return nil
}

// StopAll 停止所有代理服务
func (sm *ServerManager) StopAll() {
	sm.mu.Lock()
	defer sm.mu.Unlock()

	for id, rs := range sm.servers {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		log.Printf("[代理停止] %s (端口 %d)", rs.rule.Name, rs.rule.Port)
		rs.server.Shutdown(ctx)
		cancel()
		delete(sm.servers, id)
	}
}

// GetRunningPorts 返回所有运行中规则占用的端口列表
func (sm *ServerManager) GetRunningPorts() map[int]bool {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	ports := make(map[int]bool)
	for _, rs := range sm.servers {
		ports[rs.rule.Port] = true
	}
	return ports
}
