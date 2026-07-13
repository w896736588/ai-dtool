// Package step_executor 提供 E2E 用例的步骤执行器框架。
// 核心原则：每个步骤类型独立实现，支持版本控制（input_v1 / input_v2 互不影响）。
package step_executor

import (
	"dev_tool/internal/app/dtool/component/e2e/interceptor"
	"dev_tool/internal/app/dtool/component/e2e/variable"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"strings"
	"sync"

	"github.com/playwright-community/playwright-go"
)

// ExecuteContext 步骤执行上下文（执行单个步骤时由引擎装配）。
type ExecuteContext struct {
	RunID         int64
	RunStepDBID   int
	CaseID        int
	EnvURL        string
	EnvBaseURL    string
	Page          playwright.Page
	Browser       playwright.Browser
	VarContext    *variable.Context
	Resolver      *variable.Resolver
	RequestRepo   *interceptor.RequestRepository
	Output        *OutputBuffer
	ScreenshotDir string
}

// OutputBuffer 引擎侧供前端 SSE 拉取的输出缓冲区。
type OutputBuffer struct {
	mu      sync.Mutex
	Lines   []string
	OnWrite func(line string)
}

func (o *OutputBuffer) Writef(format string, args ...any) {
	if o == nil {
		return
	}
	line := formatString(format, args...)
	o.mu.Lock()
	o.Lines = append(o.Lines, line)
	cb := o.OnWrite
	o.mu.Unlock()
	if cb != nil {
		cb(line)
	}
}

func formatString(format string, args ...any) string {
	if len(args) == 0 {
		return format
	}
	out := format
	for _, a := range args {
		out = strings.Replace(out, "{}", toString(a), 1)
	}
	return out
}

func toString(v any) string {
	if v == nil {
		return ""
	}
	switch s := v.(type) {
	case string:
		return s
	default:
		b, _ := json.Marshal(v)
		return string(b)
	}
}

// StepResult 单个步骤的执行结果。
type StepResult struct {
	Success      bool              `json:"success"`
	ErrorMsg     string            `json:"error_msg,omitempty"`
	Screenshot   string            `json:"screenshot,omitempty"`
	ExtractedVar map[string]string `json:"extracted_var,omitempty"`
	DurationMs   int64             `json:"duration_ms,omitempty"`
}

// ResolveVariablesInConfig 深拷贝 Config 并对字符串字段进行变量解析。
// 实现仅对 string 类型的字段递归处理；遇到 map/slice 不递归，避免破坏结构。
func ResolveVariablesInConfig(raw json.RawMessage, resolver *variable.Resolver) json.RawMessage {
	if len(raw) == 0 {
		return raw
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return raw
	}
	resolved := resolveValue(v, resolver)
	out, err := json.Marshal(resolved)
	if err != nil {
		return raw
	}
	return out
}

func resolveValue(v any, r *variable.Resolver) any {
	switch x := v.(type) {
	case string:
		return r.Resolve(x)
	case map[string]any:
		out := make(map[string]any, len(x))
		for k, val := range x {
			out[k] = resolveValue(val, r)
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, val := range x {
			out[i] = resolveValue(val, r)
		}
		return out
	default:
		return v
	}
}

// ParseVersionedConfig 通用版本感知解析：按 type 中 _v 后的部分确定版本。
func ParseVersionedConfig(raw json.RawMessage, target any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, target)
}

// ApplySelector 把 selector + selector_type 转成 playwright 定位函数。
func ApplySelector(page playwright.Page, selector, selectorType string, timeoutMs int) (playwright.Locator, error) {
	if timeoutMs <= 0 {
		timeoutMs = 5000
	}
	st := strings.ToLower(strings.TrimSpace(selectorType))
	if st == "" {
		st = "css"
	}
	var loc playwright.Locator
	switch st {
	case "css":
		loc = page.Locator(selector)
	case "xpath":
		loc = page.Locator("xpath=" + selector)
	case "text":
		loc = page.Locator("text=" + selector)
	case "role":
		loc = page.GetByRole(playwright.AriaRole(selector))
	case "id":
		loc = page.Locator("#" + selector)
	default:
		loc = page.Locator(selector)
	}
	if err := loc.First().WaitFor(playwright.LocatorWaitForOptions{
		State:   playwright.WaitForSelectorStateVisible,
		Timeout: playwright.Float(float64(timeoutMs)),
	}); err != nil {
		return nil, err
	}
	return loc, nil
}

// StepExecutor 步骤执行器统一接口。
type StepExecutor interface {
	Type() define.E2EStepType
	Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult
	Validate(step *define.E2EStep) error
}

// Registry 步骤版本注册表（key: "input_v1"）。
type Registry struct {
	mu       sync.RWMutex
	handlers map[string]StepExecutor
	deprecated map[string]string
}

func NewRegistry() *Registry {
	return &Registry{
		handlers:   make(map[string]StepExecutor),
		deprecated: make(map[string]string),
	}
}

func (r *Registry) Register(exec StepExecutor) {
	if exec == nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[string(exec.Type())] = exec
}

func (r *Registry) MarkDeprecated(t define.E2EStepType, reason string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.deprecated[string(t)] = reason
}

// IsDeprecated 报告某类型是否已废弃。
func (r *Registry) IsDeprecated(t define.E2EStepType) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.deprecated[string(t)]
	return ok
}

func (r *Registry) Get(t define.E2EStepType) (StepExecutor, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	h, ok := r.handlers[string(t)]
	return h, ok
}
