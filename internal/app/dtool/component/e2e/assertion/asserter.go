// Package assertion 提供 E2E 用例的断言器框架。
// 核心原则：每个断言类型独立实现，可组合使用。
package assertion

import (
	"dev_tool/internal/app/dtool/component/e2e/interceptor"
	"dev_tool/internal/app/dtool/component/e2e/variable"
	"dev_tool/internal/app/dtool/define"
	"sync"

	"github.com/playwright-community/playwright-go"
)

// AssertionContext 断言执行上下文。
type AssertionContext struct {
	Page        playwright.Page
	RequestRepo *interceptor.RequestRepository
	VarContext  *variable.Context
	Resolver    *variable.Resolver
}

// AssertionResult 断言结果。
type AssertionResult struct {
	Success      bool   `json:"success"`
	Expected     string `json:"expected,omitempty"`
	Actual       string `json:"actual,omitempty"`
	ErrorMsg     string `json:"error_msg,omitempty"`
	MatchedURL   string `json:"matched_url,omitempty"`
	MatchedReqID string `json:"matched_request_id,omitempty"`
}

// Asserter 断言器接口。
type Asserter interface {
	Type() define.E2EAssertionType
	Assert(ctx *AssertionContext, assertion *define.E2EAssertion) *AssertionResult
	Validate(assertion *define.E2EAssertion) error
}

// Registry 断言版本注册表（key: "assert_text_v1"）。
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]Asserter
}

func NewRegistry() *Registry {
	return &Registry{handlers: make(map[string]Asserter)}
}

func (r *Registry) Register(a Asserter) {
	if a == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[string(a.Type())] = a
}

func (r *Registry) Get(t define.E2EAssertionType) (Asserter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[string(t)]
	return h, ok
}
