package step_executor

import (
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// OpenEnvExecutor open_env：打开目标环境。
type OpenEnvExecutor struct{}

func (e *OpenEnvExecutor) Type() define.E2EStepType { return define.E2EStepOpenEnv }

func (e *OpenEnvExecutor) Validate(step *define.E2EStep) error {
	if len(step.Config) == 0 {
		return errors.New("open_env: config 不能为空")
	}
	var cfg struct {
		URL string `json:"url"`
	}
	_ = unmarshalJSON(step.Config, &cfg)
	if strings.TrimSpace(cfg.URL) == "" {
		return errors.New("open_env: url 不能为空")
	}
	return nil
}

func (e *OpenEnvExecutor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	ctx.Output.Writef("[step] open_env 开始")
	var cfg struct {
		URL       string `json:"url"`
		URLType   string `json:"url_type"`
		WaitLoad  bool   `json:"wait_load"`
		TimeoutMs int    `json:"timeout_ms"`
	}
	_ = unmarshalJSON(step.Config, &cfg)
	url := ctx.Resolver.Resolve(cfg.URL)
	timeout := float64(cfg.TimeoutMs)
	if timeout <= 0 {
		timeout = 30000
	}
	if _, err := ctx.Page.Goto(url, playwright.PageGotoOptions{Timeout: playwright.Float(timeout)}); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("goto %s 失败: %v", url, err)}
	}
	if cfg.WaitLoad {
		ctx.Page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(timeout),
		})
	}
	ctx.Output.Writef("[step] open_env 完成 url=%s", url)
	return &StepResult{Success: true}
}

// ClickV1Executor click_v1：基础点击。
type ClickV1Executor struct{}

func (e *ClickV1Executor) Type() define.E2EStepType { return define.E2EStepClickV1 }

func (e *ClickV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.ClickV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("click_v1: selector 不能为空")
	}
	return nil
}

func (e *ClickV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	ctx.Output.Writef("[step] click_v1 开始")
	var cfg define.ClickV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)

	loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, cfg.TimeoutMs)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("定位失败: %v", err)}
	}
	clickOpts := playwright.LocatorClickOptions{Force: playwright.Bool(cfg.Force)}
	if cfg.ClickCount > 1 {
		clickOpts.ClickCount = playwright.Int(cfg.ClickCount)
	}
	if err := loc.Click(clickOpts); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("点击失败: %v", err)}
	}
	ctx.Output.Writef("[step] click_v1 完成 selector=%s", cfg.Selector)
	return &StepResult{Success: true}
}

// InputV1Executor input_v1：固定输入。
type InputV1Executor struct{}

func (e *InputV1Executor) Type() define.E2EStepType { return define.E2EStepInputV1 }

func (e *InputV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.InputV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("input_v1: selector 不能为空")
	}
	return nil
}

func (e *InputV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	ctx.Output.Writef("[step] input_v1 开始")
	var cfg define.InputV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)
	value := ctx.Resolver.Resolve(cfg.Value)

	loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, cfg.TimeoutMs)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("定位失败: %v", err)}
	}
	if cfg.ClearBefore {
		_ = loc.Clear()
	}
	if err := loc.Fill(value); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("填充失败: %v", err)}
	}
	if cfg.PressEnter {
		_ = loc.Press("Enter")
	}
	ctx.Output.Writef("[step] input_v1 完成 selector=%s", cfg.Selector)
	return &StepResult{Success: true}
}

// HoverV1Executor hover_v1：悬停。
type HoverV1Executor struct{}

func (e *HoverV1Executor) Type() define.E2EStepType { return define.E2EStepHoverV1 }

func (e *HoverV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.HoverV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("hover_v1: selector 不能为空")
	}
	return nil
}

func (e *HoverV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.HoverV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)
	loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, cfg.TimeoutMs)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("定位失败: %v", err)}
	}
	if err := loc.Hover(); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("hover 失败: %v", err)}
	}
	return &StepResult{Success: true}
}

// SelectV1Executor select_v1：下拉选择。
type SelectV1Executor struct{}

func (e *SelectV1Executor) Type() define.E2EStepType { return define.E2EStepSelectV1 }

func (e *SelectV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.SelectV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("select_v1: selector 不能为空")
	}
	return nil
}

func (e *SelectV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.SelectV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)
	value := ctx.Resolver.Resolve(cfg.Value)
	loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, cfg.TimeoutMs)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("定位失败: %v", err)}
	}
	if _, err := loc.SelectOption(playwright.SelectOptionValues{Values: &[]string{value}}); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("select 失败: %v", err)}
	}
	return &StepResult{Success: true}
}

// NavigateV1Executor navigate_v1：导航到 URL。
type NavigateV1Executor struct{}

func (e *NavigateV1Executor) Type() define.E2EStepType { return define.E2EStepNavigateV1 }

func (e *NavigateV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.NavigateV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.URL) == "" {
		return errors.New("navigate_v1: url 不能为空")
	}
	return nil
}

func (e *NavigateV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.NavigateV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	url := ctx.Resolver.Resolve(cfg.URL)
	if cfg.URLType == "relative" && ctx.EnvBaseURL != "" {
		url = strings.TrimRight(ctx.EnvBaseURL, "/") + "/" + strings.TrimLeft(url, "/")
	}
	timeout := float64(cfg.TimeoutMs)
	if timeout <= 0 {
		timeout = 30000
	}
	if _, err := ctx.Page.Goto(url, playwright.PageGotoOptions{Timeout: playwright.Float(timeout)}); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("goto 失败: %v", err)}
	}
	if cfg.WaitLoad {
		ctx.Page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(timeout),
		})
	}
	return &StepResult{Success: true}
}

// WaitElementV1Executor wait_element_v1：等待元素出现。
type WaitElementV1Executor struct{}

func (e *WaitElementV1Executor) Type() define.E2EStepType { return define.E2EStepWaitElementV1 }

func (e *WaitElementV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.WaitElementV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("wait_element_v1: selector 不能为空")
	}
	return nil
}

func (e *WaitElementV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.WaitElementV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)
	state := playwright.WaitForSelectorStateVisible
	if strings.EqualFold(cfg.State, "hidden") {
		state = playwright.WaitForSelectorStateHidden
	} else if strings.EqualFold(cfg.State, "attached") {
		state = playwright.WaitForSelectorStateAttached
	} else if strings.EqualFold(cfg.State, "detached") {
		state = playwright.WaitForSelectorStateDetached
	}
	timeout := float64(cfg.TimeoutMs)
	if timeout <= 0 {
		timeout = 30000
	}
	loc := ctx.Page.Locator(cfg.Selector)
	if err := loc.First().WaitFor(playwright.LocatorWaitForOptions{State: state, Timeout: playwright.Float(timeout)}); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("等待失败: %v", err)}
	}
	return &StepResult{Success: true}
}

// WaitTimeoutV1Executor wait_timeout_v1：固定等待。
type WaitTimeoutV1Executor struct{}

func (e *WaitTimeoutV1Executor) Type() define.E2EStepType { return define.E2EStepWaitTimeoutV1 }

func (e *WaitTimeoutV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.WaitTimeoutV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if cfg.DurationMs <= 0 {
		return errors.New("wait_timeout_v1: duration_ms 必须大于 0")
	}
	return nil
}

func (e *WaitTimeoutV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.WaitTimeoutV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	ctx.Page.WaitForTimeout(float64(cfg.DurationMs))
	return &StepResult{Success: true}
}

// ExtractTextV1Executor extract_text_v1：提取文本到变量。
type ExtractTextV1Executor struct{}

func (e *ExtractTextV1Executor) Type() define.E2EStepType { return define.E2EStepExtractTextV1 }

func (e *ExtractTextV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.ExtractTextV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("extract_text_v1: selector 不能为空")
	}
	if strings.TrimSpace(cfg.ExtractTo) == "" {
		return errors.New("extract_text_v1: extract_to 不能为空")
	}
	return nil
}

func (e *ExtractTextV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.ExtractTextV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)
	loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, cfg.TimeoutMs)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("定位失败: %v", err)}
	}
	text, err := loc.TextContent()
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("获取文本失败: %v", err)}
	}
	out := text
	if cfg.Regex != "" {
		// 简化处理：直接采用 contains；项目暂不引入正则包
		if !strings.Contains(text, cfg.Regex) {
			out = ""
		}
	}
	ctx.VarContext.Set(cfg.ExtractTo, out)
	return &StepResult{
		Success: true,
		ExtractedVar: map[string]string{
			cfg.ExtractTo: out,
		},
	}
}

// ExtractAttrV1Executor extract_attr_v1：提取属性到变量。
type ExtractAttrV1Executor struct{}

func (e *ExtractAttrV1Executor) Type() define.E2EStepType { return define.E2EStepExtractAttrV1 }

func (e *ExtractAttrV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.ExtractAttrV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("extract_attr_v1: selector 不能为空")
	}
	if strings.TrimSpace(cfg.Attribute) == "" {
		return errors.New("extract_attr_v1: attribute 不能为空")
	}
	return nil
}

func (e *ExtractAttrV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.ExtractAttrV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)
	loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, cfg.TimeoutMs)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("定位失败: %v", err)}
	}
	val, err := loc.GetAttribute(cfg.Attribute)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("获取属性失败: %v", err)}
	}
	ctx.VarContext.Set(cfg.ExtractTo, val)
	return &StepResult{Success: true, ExtractedVar: map[string]string{cfg.ExtractTo: val}}
}

// ScriptV1Executor script_v1：执行自定义 JS。
type ScriptV1Executor struct{}

func (e *ScriptV1Executor) Type() define.E2EStepType { return define.E2EStepScriptV1 }

func (e *ScriptV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.ScriptV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Code) == "" {
		return errors.New("script_v1: code 不能为空")
	}
	return nil
}

func (e *ScriptV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.ScriptV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	if _, err := ctx.Page.Evaluate(cfg.Code); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("script 失败: %v", err)}
	}
	return &StepResult{Success: true}
}

// GoBackV1Executor go_back_v1：浏览器后退。
type GoBackV1Executor struct{}

func (e *GoBackV1Executor) Type() define.E2EStepType { return define.E2EStepGoBackV1 }

func (e *GoBackV1Executor) Validate(step *define.E2EStep) error { return nil }

func (e *GoBackV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	if _, err := ctx.Page.GoBack(); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("go_back 失败: %v", err)}
	}
	return &StepResult{Success: true}
}

// ReloadV1Executor reload_v1：刷新页面。
type ReloadV1Executor struct{}

func (e *ReloadV1Executor) Type() define.E2EStepType { return define.E2EStepReloadV1 }

func (e *ReloadV1Executor) Validate(step *define.E2EStep) error { return nil }

func (e *ReloadV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	if _, err := ctx.Page.Reload(); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("reload 失败: %v", err)}
	}
	return &StepResult{Success: true}
}

// PressKeyV1Executor press_key_v1：按键。
type PressKeyV1Executor struct{}

func (e *PressKeyV1Executor) Type() define.E2EStepType { return define.E2EStepPressKeyV1 }

func (e *PressKeyV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.PressKeyV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Key) == "" {
		return errors.New("press_key_v1: key 不能为空")
	}
	return nil
}

func (e *PressKeyV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.PressKeyV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	if strings.TrimSpace(cfg.Selector) != "" {
		loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, 5000)
		if err != nil {
			return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("定位失败: %v", err)}
		}
		if err := loc.Press(cfg.Key); err != nil {
			return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("press 失败: %v", err)}
		}
		return &StepResult{Success: true}
	}
	if err := ctx.Page.Keyboard().Press(cfg.Key); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("press 失败: %v", err)}
	}
	return &StepResult{Success: true}
}

// unmarshalJSON 辅助函数。
func unmarshalJSON(raw []byte, target any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, target)
}
