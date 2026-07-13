package assertion

import (
	"dev_tool/internal/app/dtool/component/e2e/interceptor"
	"dev_tool/internal/app/dtool/component/e2e/step_executor"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// TextV1Asserter assert_text_v1：基础文本断言。
type TextV1Asserter struct{}

func (a *TextV1Asserter) Type() define.E2EAssertionType { return define.E2EAssertTextV1 }

func (a *TextV1Asserter) Validate(assertion *define.E2EAssertion) error {
	var cfg define.TextAssertionV1Config
	if err := json.Unmarshal(assertion.Config, &cfg); err != nil {
		return err
	}
	if cfg.Value == "" {
		return errors.New("assert_text_v1: value 不能为空")
	}
	return nil
}

func (a *TextV1Asserter) Assert(ctx *AssertionContext, assertion *define.E2EAssertion) *AssertionResult {
	var cfg define.TextAssertionV1Config
	_ = json.Unmarshal(assertion.Config, &cfg)
	if ctx.Resolver != nil {
		cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)
		cfg.Value = ctx.Resolver.Resolve(cfg.Value)
	}
	var actual string
	if cfg.Selector != "" {
		loc, err := step_executor.ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, 5000)
		if err != nil {
			return &AssertionResult{Success: false, Expected: cfg.Value, ErrorMsg: fmt.Sprintf("定位失败: %v", err)}
		}
		text, terr := loc.TextContent()
		if terr != nil {
			return &AssertionResult{Success: false, Expected: cfg.Value, ErrorMsg: fmt.Sprintf("读取文本失败: %v", terr)}
		}
		actual = text
	} else {
		// 选择器为空：尝试通过 JS 读取整个 body
		bodyText, _ := ctx.Page.Evaluate("() => document.body && document.body.innerText || ''")
		if s, ok := bodyText.(string); ok {
			actual = s
		}
	}
	if !compareText(actual, cfg.Value, cfg.Operator, false, false) {
		return &AssertionResult{
			Success:  false,
			Expected: fmt.Sprintf("operator=%s value=%s", cfg.Operator, cfg.Value),
			Actual:   truncate(actual, 1000),
			ErrorMsg: "文本断言失败",
		}
	}
	return &AssertionResult{Success: true}
}

// TextV2Asserter assert_text_v2：增强文本断言（新增 ignore_case / normalize_ws）。
type TextV2Asserter struct{}

func (a *TextV2Asserter) Type() define.E2EAssertionType { return define.E2EAssertTextV2 }

func (a *TextV2Asserter) Validate(assertion *define.E2EAssertion) error {
	var cfg define.TextAssertionV2Config
	if err := json.Unmarshal(assertion.Config, &cfg); err != nil {
		return err
	}
	if cfg.Value == "" {
		return errors.New("assert_text_v2: value 不能为空")
	}
	return nil
}

func (a *TextV2Asserter) Assert(ctx *AssertionContext, assertion *define.E2EAssertion) *AssertionResult {
	var cfg define.TextAssertionV2Config
	_ = json.Unmarshal(assertion.Config, &cfg)
	if ctx.Resolver != nil {
		cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)
		cfg.Value = ctx.Resolver.Resolve(cfg.Value)
	}
	var actual string
	if cfg.Selector != "" {
		loc, err := step_executor.ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, 5000)
		if err != nil {
			return &AssertionResult{Success: false, Expected: cfg.Value, ErrorMsg: err.Error()}
		}
		text, terr := loc.TextContent()
		if terr != nil {
			return &AssertionResult{Success: false, ErrorMsg: terr.Error()}
		}
		actual = text
	} else {
		bodyText, _ := ctx.Page.Evaluate("() => document.body && document.body.innerText || ''")
		if s, ok := bodyText.(string); ok {
			actual = s
		}
	}
	if !compareText(actual, cfg.Value, cfg.Operator, cfg.IgnoreCase, cfg.NormalizeWS) {
		return &AssertionResult{
			Success:  false,
			Expected: fmt.Sprintf("operator=%s value=%s", cfg.Operator, cfg.Value),
			Actual:   truncate(actual, 1000),
			ErrorMsg: "文本断言失败",
		}
	}
	return &AssertionResult{Success: true}
}

// ElementV1Asserter assert_element_v1：元素状态断言。
type ElementV1Asserter struct{}

func (a *ElementV1Asserter) Type() define.E2EAssertionType { return define.E2EAssertElementV1 }

func (a *ElementV1Asserter) Validate(assertion *define.E2EAssertion) error {
	var cfg define.ElementAssertionV1Config
	if err := json.Unmarshal(assertion.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("assert_element_v1: selector 不能为空")
	}
	return nil
}

func (a *ElementV1Asserter) Assert(ctx *AssertionContext, assertion *define.E2EAssertion) *AssertionResult {
	var cfg define.ElementAssertionV1Config
	_ = json.Unmarshal(assertion.Config, &cfg)
	if ctx.Resolver != nil {
		cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)
	}
	loc := ctx.Page.Locator(cfg.Selector)
	count, err := loc.Count()
	if err != nil {
		return &AssertionResult{Success: false, ErrorMsg: err.Error()}
	}
	state := strings.ToLower(strings.TrimSpace(cfg.State))
	if state == "" {
		state = "present"
	}

	expectCount := 1
	if cfg.Count != nil {
		expectCount = *cfg.Count
	}
	switch state {
	case "present":
		if count < expectCount {
			return &AssertionResult{
				Success:  false,
				Expected: fmt.Sprintf("present (count>=%d)", expectCount),
				Actual:   strconv.Itoa(count),
				ErrorMsg: "元素数量不足",
			}
		}
	case "absent", "hidden", "detached":
		if count > 0 {
			return &AssertionResult{
				Success:  false,
				Expected: "absent",
				Actual:   strconv.Itoa(count),
				ErrorMsg: "元素存在",
			}
		}
	}
	return &AssertionResult{Success: true}
}

// URLV1Asserter assert_url_v1：URL 断言。
type URLV1Asserter struct{}

func (a *URLV1Asserter) Type() define.E2EAssertionType { return define.E2EAssertURLV1 }

func (a *URLV1Asserter) Validate(assertion *define.E2EAssertion) error {
	var cfg define.URLAssertionV1Config
	if err := json.Unmarshal(assertion.Config, &cfg); err != nil {
		return err
	}
	if cfg.Value == "" {
		return errors.New("assert_url_v1: value 不能为空")
	}
	return nil
}

func (a *URLV1Asserter) Assert(ctx *AssertionContext, assertion *define.E2EAssertion) *AssertionResult {
	var cfg define.URLAssertionV1Config
	_ = json.Unmarshal(assertion.Config, &cfg)
	if ctx.Resolver != nil {
		cfg.Value = ctx.Resolver.Resolve(cfg.Value)
	}
	url := ctx.Page.URL()
	if cfg.IgnoreQuery {
		if idx := strings.Index(url, "?"); idx >= 0 {
			url = url[:idx]
		}
	}
	ok := false
	switch cfg.Type {
	case "", "exact":
		ok = url == cfg.Value
	case "contains":
		ok = strings.Contains(url, cfg.Value)
	case "regex":
		ok = interceptor.MatchWildcardGlobal(url, cfg.Value)
	default:
		ok = url == cfg.Value
	}
	if !ok {
		return &AssertionResult{
			Success:  false,
			Expected: fmt.Sprintf("type=%s value=%s", cfg.Type, cfg.Value),
			Actual:   url,
			ErrorMsg: "URL 不匹配",
		}
	}
	return &AssertionResult{Success: true}
}

// TitleV1Asserter assert_title_v1：标题断言。
type TitleV1Asserter struct{}

func (a *TitleV1Asserter) Type() define.E2EAssertionType { return define.E2EAssertTitleV1 }

func (a *TitleV1Asserter) Validate(assertion *define.E2EAssertion) error {
	var cfg define.TitleAssertionV1Config
	if err := json.Unmarshal(assertion.Config, &cfg); err != nil {
		return err
	}
	if cfg.Value == "" {
		return errors.New("assert_title_v1: value 不能为空")
	}
	return nil
}

func (a *TitleV1Asserter) Assert(ctx *AssertionContext, assertion *define.E2EAssertion) *AssertionResult {
	var cfg define.TitleAssertionV1Config
	_ = json.Unmarshal(assertion.Config, &cfg)
	title, _ := ctx.Page.Title()
	if !compareText(title, cfg.Value, cfg.Operator, false, false) {
		return &AssertionResult{
			Success:  false,
			Expected: cfg.Value,
			Actual:   title,
			ErrorMsg: "标题断言失败",
		}
	}
	return &AssertionResult{Success: true}
}

// VariableV1Asserter assert_variable_v1：变量值断言。
type VariableV1Asserter struct{}

func (a *VariableV1Asserter) Type() define.E2EAssertionType { return define.E2EAssertVariableV1 }

func (a *VariableV1Asserter) Validate(assertion *define.E2EAssertion) error {
	var cfg define.VariableAssertionV1Config
	if err := json.Unmarshal(assertion.Config, &cfg); err != nil {
		return err
	}
	if cfg.VarName == "" {
		return errors.New("assert_variable_v1: var_name 不能为空")
	}
	return nil
}

func (a *VariableV1Asserter) Assert(ctx *AssertionContext, assertion *define.E2EAssertion) *AssertionResult {
	var cfg define.VariableAssertionV1Config
	_ = json.Unmarshal(assertion.Config, &cfg)
	actual, ok := ctx.VarContext.Get(cfg.VarName)
	if !ok {
		return &AssertionResult{
			Success:  false,
			Expected: cfg.Expected,
			Actual:   "(未定义变量)",
			ErrorMsg: fmt.Sprintf("变量 %s 未定义", cfg.VarName),
		}
	}
	switch cfg.Operator {
	case "", "eq":
		if actual != cfg.Expected {
			return &AssertionResult{Success: false, Expected: cfg.Expected, Actual: actual, ErrorMsg: "不相等"}
		}
	case "contains":
		if !strings.Contains(actual, cfg.Expected) {
			return &AssertionResult{Success: false, Expected: cfg.Expected, Actual: actual, ErrorMsg: "未包含"}
		}
	case "regex":
		if !interceptor.MatchWildcardGlobal(actual, cfg.Expected) {
			return &AssertionResult{Success: false, Expected: cfg.Expected, Actual: actual, ErrorMsg: "正则不匹配"}
		}
	default:
		// 数值比较 gt/lt
		af, _ := strconv.ParseFloat(actual, 64)
		ef, _ := strconv.ParseFloat(cfg.Expected, 64)
		switch cfg.Operator {
		case "gt":
			if !(af > ef) {
				return &AssertionResult{Success: false, Expected: cfg.Expected, Actual: actual, ErrorMsg: "不大于"}
			}
		case "lt":
			if !(af < ef) {
				return &AssertionResult{Success: false, Expected: cfg.Expected, Actual: actual, ErrorMsg: "不小于"}
			}
		}
	}
	return &AssertionResult{Success: true, Actual: actual}
}

func compareText(actual, expected, op string, ignoreCase, normalizeWS bool) bool {
	if normalizeWS {
		actual = normalizeWhitespace(actual)
		expected = normalizeWhitespace(expected)
	}
	if ignoreCase {
		actual = strings.ToLower(actual)
		expected = strings.ToLower(expected)
	}
	switch op {
	case "", "exact", "eq":
		return actual == expected
	case "contains":
		return strings.Contains(actual, expected)
	case "not_contains":
		return !strings.Contains(actual, expected)
	case "regex":
		return interceptor.MatchWildcardGlobal(actual, expected)
	default:
		return actual == expected
	}
}

func normalizeWhitespace(s string) string {
	var b strings.Builder
	inSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !inSpace {
				b.WriteByte(' ')
				inSpace = true
			}
			continue
		}
		inSpace = false
		b.WriteRune(r)
	}
	return strings.TrimSpace(b.String())
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "...(截断)"
}

// 引入 playwright 类型（包级）以避免 import 顺序不一致
var _ playwright.WebError = nil
