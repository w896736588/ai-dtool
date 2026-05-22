package main

import (
	"sync"
	"time"

	"github.com/google/uuid"
)

// StreamChunk 流式响应中的单个数据块
type StreamChunk struct {
	Timestamp time.Time `json:"timestamp"`
	Data      string    `json:"data"`
}

// Record 单次请求的完整记录
type Record struct {
	ID                   string            `json:"id"`
	Time                 time.Time         `json:"time"`
	CompleteTime         time.Time         `json:"complete_time"`
	Method               string            `json:"method"`
	Path                 string            `json:"path"`
	Headers              map[string]string `json:"headers"`
	QueryParams          map[string]string `json:"query_params"`
	Body                 string            `json:"body"`
	StatusCode           int               `json:"status_code"`
	ResponseBody         string            `json:"response_body"`
	Duration             int64             `json:"duration_ms"`
	InputTokens          int               `json:"input_tokens"`
	OutputTokens         int               `json:"output_tokens"`
	CacheReadInputTokens int               `json:"cache_read_input_tokens"`
	Model                string            `json:"model"`
	RuleID               string            `json:"rule_id"`
	IsStream             bool              `json:"is_stream"`
	Completed            bool              `json:"completed"`
	StreamChunks         []StreamChunk     `json:"stream_chunks,omitempty"`
}

// sseEvent SSE 推送事件
type sseEvent struct {
	Type   string    `json:"type"` // "init" 或 "update"
	Record *Record   `json:"record,omitempty"`
	All    []*Record `json:"all,omitempty"`
}

// Store 线程安全的内存数据存储，支持发布/订阅
type Store struct {
	mu          sync.RWMutex
	records     []*Record
	subscribers map[string]chan *sseEvent
}

// NewStore 创建新的存储实例
func NewStore() *Store {
	return &Store{
		records:     make([]*Record, 0),
		subscribers: make(map[string]chan *sseEvent),
	}
}

// Subscribe 注册订阅者，返回唯一 ID 和事件 channel
func (s *Store) Subscribe() (string, <-chan *sseEvent) {
	s.mu.Lock()
	defer s.mu.Unlock()
	id := uuid.New().String()
	// 使用带缓冲的 channel，避免慢客户端阻塞写入
	ch := make(chan *sseEvent, 1000)
	s.subscribers[id] = ch

	// 发送初始全量数据
	all := s.listLocked()
	go func() {
		ch <- &sseEvent{Type: "init", All: all}
	}()

	return id, ch
}

// Unsubscribe 取消订阅
func (s *Store) Unsubscribe(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ch, ok := s.subscribers[id]; ok {
		close(ch)
		delete(s.subscribers, id)
	}
}

// Add 添加一条记录，并通知所有订阅者
func (s *Store) Add(r *Record) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = append(s.records, r)
	s.broadcastLocked(&sseEvent{Type: "update", Record: r})
}

// UpdateRecord 更新已存在的记录，并通知所有订阅者
func (s *Store) UpdateRecord(r *Record) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i, rec := range s.records {
		if rec.ID == r.ID {
			s.records[i] = r
			break
		}
	}
	s.broadcastLocked(&sseEvent{Type: "update", Record: r})
}

// List 返回所有记录（按时间倒序）
func (s *Store) List() []*Record {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.listLocked()
}

// listLocked 内部使用，调用者需持有锁
func (s *Store) listLocked() []*Record {
	result := make([]*Record, len(s.records))
	for i, r := range s.records {
		result[len(s.records)-1-i] = r
	}
	return result
}

// Clear 清空所有记录，通知订阅者
func (s *Store) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.records = make([]*Record, 0)
	s.broadcastLocked(&sseEvent{Type: "init", All: make([]*Record, 0)})
}

// Count 返回记录总数
func (s *Store) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.records)
}

// broadcastLocked 广播给所有订阅者，调用者需持有锁
func (s *Store) broadcastLocked(evt *sseEvent) {
	for _, ch := range s.subscribers {
		select {
		case ch <- evt:
		default:
			// channel 满了，丢弃该次推送（慢客户端跳过）
		}
	}
}
