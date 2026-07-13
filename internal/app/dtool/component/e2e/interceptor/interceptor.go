// Package interceptor 提供 E2E 用例执行期间的请求拦截能力。
// 核心原则：通过真实捕获用户操作触发的 XHR/Fetch 请求来支持 API 断言。
package interceptor

import (
	"encoding/json"
	"strconv"
	"strings"
	"sync"
	"time"
)

// RequestRepository 请求仓库，线程安全存储所有捕获到的请求和响应。
type RequestRepository struct {
	mu       sync.RWMutex
	requests map[string]*CapturedRequest
	byURL    map[string][]string
}

func NewRequestRepository() *RequestRepository {
	return &RequestRepository{
		requests: make(map[string]*CapturedRequest),
		byURL:    make(map[string][]string),
	}
}

// Add 添加捕获请求（已含响应）。
func (r *RequestRepository) Add(req *CapturedRequest) {
	if req == nil || req.ID == "" {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.requests[req.ID] = req
	r.byURL[req.URL] = append(r.byURL[req.URL], req.ID)
}

// Get 按 ID 获取。
func (r *RequestRepository) Get(id string) *CapturedRequest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.requests[id]
}

// GetByURL 按精确 URL 获取。
func (r *RequestRepository) GetByURL(url string) []*CapturedRequest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := r.byURL[url]
	out := make([]*CapturedRequest, 0, len(ids))
	for _, id := range ids {
		if req := r.requests[id]; req != nil {
			out = append(out, req)
		}
	}
	return out
}

// GetAll 返回所有捕获请求。
func (r *RequestRepository) GetAll() []*CapturedRequest {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*CapturedRequest, 0, len(r.requests))
	for _, req := range r.requests {
		out = append(out, req)
	}
	return out
}

// MarkMatched 标记某个请求已被指定断言匹配。
func (r *RequestRepository) MarkMatched(id, by string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if req, ok := r.requests[id]; ok {
		req.Matched = true
		req.MatchedBy = by
	}
}

// Count 返回已捕获请求数。
func (r *RequestRepository) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.requests)
}

// CapturedRequest 捕获到的请求 + 响应。
type CapturedRequest struct {
	ID         string            `json:"id"`
	RunStepID  int               `json:"run_step_id"`
	URL        string            `json:"url"`
	Method     string            `json:"method"`
	Headers    map[string]string `json:"headers"`
	PostData   string            `json:"post_data,omitempty"`
	Response   *CapturedResponse `json:"response,omitempty"`
	Matched    bool              `json:"matched"`
	MatchedBy  string            `json:"matched_by,omitempty"`
	CapturedAt time.Time         `json:"captured_at"`
}

// CapturedResponse 响应体。
type CapturedResponse struct {
	Status     int               `json:"status"`
	StatusText string            `json:"status_text"`
	Headers    map[string]string `json:"headers"`
	Body       string            `json:"body"`
	BodySize   int               `json:"body_size"`
	TimeMs     int               `json:"time_ms"`
}

// MatchConfig 通用匹配条件组合。
type MatchConfig struct {
	URL              string // 精确 URL
	Contains         string // URL 包含
	Regex            string // URL 包含通配 * 时转换为 contains 序列
	Suffix           string // URL 后缀
	Method           string // HTTP 方法
	ResponseContains string // 响应体包含
	Status           int    // 响应状态码
}

// Matcher 请求匹配器。
type Matcher struct{}

func NewMatcher() *Matcher { return &Matcher{} }

// Match 按条件匹配所有捕获请求。
func (m *Matcher) Match(repo *RequestRepository, cfg *MatchConfig) []*CapturedRequest {
	if cfg == nil {
		return repo.GetAll()
	}
	all := repo.GetAll()
	out := make([]*CapturedRequest, 0, len(all))
	for _, req := range all {
		if m.matches(req, cfg) {
			out = append(out, req)
		}
	}
	return out
}

func (m *Matcher) matches(req *CapturedRequest, cfg *MatchConfig) bool {
	if cfg.URL != "" && req.URL != cfg.URL {
		return false
	}
	if cfg.Contains != "" && !strings.Contains(req.URL, cfg.Contains) {
		return false
	}
	if cfg.Suffix != "" && !strings.HasSuffix(req.URL, cfg.Suffix) {
		return false
	}
	if cfg.Method != "" && !strings.EqualFold(req.Method, cfg.Method) {
		return false
	}
	if cfg.Status != 0 {
		if req.Response == nil || req.Response.Status != cfg.Status {
			return false
		}
	}
	if cfg.ResponseContains != "" {
		if req.Response == nil || !strings.Contains(req.Response.Body, cfg.ResponseContains) {
			return false
		}
	}
	if cfg.Regex != "" {
		if !matchWildcard(req.URL, cfg.Regex) {
			return false
		}
	}
	return true
}

// matchWildcard 简易通配匹配（仅支持 * 通配符，如 /api/user/*/info）。
func matchWildcard(s, pattern string) bool {
	if !strings.Contains(pattern, "*") {
		return strings.Contains(s, pattern)
	}
	parts := strings.Split(pattern, "*")
	idx := 0
	for i, part := range parts {
		if part == "" {
			continue
		}
		found := strings.Index(s[idx:], part)
		if found < 0 {
			return false
		}
		if i == 0 && found != 0 {
			return false
		}
		idx += found + len(part)
	}
	return true
}

// MatchWildcardGlobal 公开的通配匹配（供外部断言器复用）。
func MatchWildcardGlobal(s, pattern string) bool { return matchWildcard(s, pattern) }

// ExtractJSONPath 从 JSON 值中按 $.a.b[0] 路径提取字符串表示。
// 支持 map[string]any / []any / 基本类型。
func ExtractJSONPath(data any, path string) string {
	trimmed := strings.TrimPrefix(path, "$")
	trimmed = strings.TrimPrefix(trimmed, ".")
	parts := splitJSONPath(trimmed)
	current := data
	for _, part := range parts {
		if current == nil {
			return ""
		}
		switch v := current.(type) {
		case map[string]any:
			current = v[part]
		case []any:
			idx, err := strconv.Atoi(part)
			if err != nil || idx < 0 || idx >= len(v) {
				return ""
			}
			current = v[idx]
		default:
			return ""
		}
	}
	if current == nil {
		return ""
	}
	if s, ok := current.(string); ok {
		return s
	}
	return stringify(current)
}

func splitJSONPath(p string) []string {
	parts := make([]string, 0)
	buf := strings.Builder{}
	for i := 0; i < len(p); i++ {
		c := p[i]
		switch c {
		case '.':
			if buf.Len() > 0 {
				parts = append(parts, buf.String())
				buf.Reset()
			}
		case '[':
			if buf.Len() > 0 {
				parts = append(parts, buf.String())
				buf.Reset()
			}
			end := strings.IndexByte(p[i+1:], ']')
			if end < 0 {
				return parts
			}
			parts = append(parts, p[i+1:i+1+end])
			i += end + 1
		default:
			buf.WriteByte(c)
		}
	}
	if buf.Len() > 0 {
		parts = append(parts, buf.String())
	}
	return parts
}

func stringify(v any) string {
	switch x := v.(type) {
	case string:
		return x
	case float64:
		if x == float64(int64(x)) {
			return strconv.FormatInt(int64(x), 10)
		}
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(x)
	case nil:
		return ""
	default:
		b, err := json.Marshal(v)
		if err != nil {
			return ""
		}
		return string(b)
	}
}
