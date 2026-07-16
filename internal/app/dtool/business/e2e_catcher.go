package business

import (
	"dev_tool/internal/app/dtool/component/e2e/interceptor"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"
)

// E2ERequestCatcher 在 Playwright Page 上接管请求 / 响应事件，自动写入 repository。
// 实现要点：
// - 通过 page.OnRequest / OnResponse 监听请求/响应
// - 通过 page.OnRequestFinished 补齐未收到 response 的请求
// - 使用请求 ID + 序列号确保同一 URL+Method 的多个请求能正确区分
type E2ERequestCatcher struct {
	page      playwright.Page
	repo      *interceptor.RequestRepository
	pendings  map[string]*interceptor.CapturedRequest
	seq       int64 // 序列号，确保 key 唯一
	mu        sync.Mutex
}

func NewE2ERequestCatcher(page playwright.Page, repo *interceptor.RequestRepository) *E2ERequestCatcher {
	c := &E2ERequestCatcher{
		page:     page,
		repo:     repo,
		pendings: make(map[string]*interceptor.CapturedRequest),
	}
	c.register()
	return c
}

func (c *E2ERequestCatcher) register() {
	c.page.OnRequest(func(req playwright.Request) {
		if !c.shouldCapture(req.Method(), req.URL()) {
			return
		}
		postData := ""
		if pd, err := req.PostData(); err == nil {
			postData = pd
		}

		c.mu.Lock()
		c.seq++
		seq := c.seq
		c.mu.Unlock()

		// 使用请求 ID + 序列号作为 key，避免同一 URL+Method 的请求覆盖
		key := req.URL() + "::" + req.Method() + "::" + string(rune(seq))

		captured := &interceptor.CapturedRequest{
			ID:         uuid.New().String(),
			URL:        req.URL(),
			Method:     req.Method(),
			Headers:    req.Headers(),
			PostData:   postData,
			CapturedAt: time.Now(),
		}
		c.mu.Lock()
		c.pendings[key] = captured
		c.mu.Unlock()
	})

	c.page.OnResponse(func(resp playwright.Response) {
		key := resp.URL() + "::" + resp.Request().Method()

		// 查找匹配的 key（可能需要遍历）
		c.mu.Lock()
		defer c.mu.Unlock()

		var matchedKey string
		var cap *interceptor.CapturedRequest
		for k, v := range c.pendings {
			if strings.HasPrefix(k, key+"::") && v.URL == resp.URL() && v.Method == resp.Request().Method() {
				matchedKey = k
				cap = v
				break
			}
		}
		if matchedKey == "" {
			return
		}

		body := ""
		if b, err := resp.Body(); err == nil {
			body = string(b)
		}

		startTime := cap.CapturedAt
		timeMs := int(time.Since(startTime).Milliseconds())

		cap.Response = &interceptor.CapturedResponse{
			Status:     resp.Status(),
			StatusText: resp.StatusText(),
			Headers:    resp.Headers(),
			Body:       body,
			BodySize:   len(body),
			TimeMs:     timeMs,
		}
		c.repo.Add(cap)
		delete(c.pendings, matchedKey)
	})

	c.page.OnRequestFinished(func(req playwright.Request) {
		if !c.shouldCapture(req.Method(), req.URL()) {
			return
		}

		c.mu.Lock()
		defer c.mu.Unlock()

		// 查找匹配的 key
		key := req.URL() + "::" + req.Method()
		for k, v := range c.pendings {
			if strings.HasPrefix(k, key+"::") && v.URL == req.URL() && v.Method == req.Method() {
				if v.Response == nil {
					// 响应未收到，添加无响应的请求记录
					c.repo.Add(v)
				}
				delete(c.pendings, k)
				return
			}
		}
	})
}

// shouldCapture 过滤：仅保留可能包含业务 API 的请求，剔除静态资源。
func (c *E2ERequestCatcher) shouldCapture(method, url string) bool {
	if url == "" {
		return false
	}
	if strings.HasPrefix(url, "data:") || strings.HasPrefix(url, "blob:") {
		return false
	}
	lower := strings.ToLower(url)
	staticExts := []string{".js", ".css", ".png", ".jpg", ".jpeg", ".svg", ".ico", ".woff", ".woff2", ".map", ".gif", ".webp"}
	for _, ext := range staticExts {
		if strings.HasSuffix(lower, ext) {
			return false
		}
	}
	if strings.Contains(lower, "/api/") || strings.Contains(lower, "/v1/") ||
		strings.Contains(lower, "/v2/") || strings.Contains(lower, ".json") ||
		strings.Contains(lower, "/rpc") || strings.Contains(lower, "/graphql") {
		return true
	}
	if method == "POST" || method == "PUT" || method == "DELETE" || method == "PATCH" {
		return true
	}
	log.Printf("[e2e catcher] skip url=%s method=%s", url, method)
	return false
}
