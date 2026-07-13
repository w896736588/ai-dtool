package business

import (
	"dev_tool/internal/app/dtool/component/e2e/interceptor"
	"log"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/playwright-community/playwright-go"
)

// E2ERequestCatcher 在 Playwright Page 上接管请求 / 响应事件，自动写入 repository。
// 实现要点：
// - 通过 page.OnRequest / OnResponse 监听请求/响应
// - 通过 page.OnRequestFinished 补齐未收到 response 的请求
type E2ERequestCatcher struct {
	page     playwright.Page
	repo     *interceptor.RequestRepository
	pendings map[string]*interceptor.CapturedRequest
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
		captured := &interceptor.CapturedRequest{
			ID:         uuid.New().String(),
			URL:        req.URL(),
			Method:     req.Method(),
			Headers:    req.Headers(),
			PostData:   postData,
			CapturedAt: time.Now(),
		}
		key := req.URL() + "::" + req.Method()
		c.pendings[key] = captured
	})

	c.page.OnResponse(func(resp playwright.Response) {
		key := resp.URL() + "::" + resp.Request().Method()
		cap, ok := c.pendings[key]
		if !ok {
			return
		}
		body := ""
		if b, err := resp.Body(); err == nil {
			body = string(b)
		}
		cap.Response = &interceptor.CapturedResponse{
			Status:     resp.Status(),
			StatusText: resp.StatusText(),
			Headers:    resp.Headers(),
			Body:       body,
			BodySize:   len(body),
		}
		c.repo.Add(cap)
		delete(c.pendings, key)
	})

	c.page.OnRequestFinished(func(req playwright.Request) {
		if !c.shouldCapture(req.Method(), req.URL()) {
			return
		}
		key := req.URL() + "::" + req.Method()
		cap, ok := c.pendings[key]
		if !ok {
			return
		}
		if cap.Response == nil {
			c.repo.Add(cap)
			delete(c.pendings, key)
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
