package step_executor

import (
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// OpenEnvExecutor open_env: open target environment.
type OpenEnvExecutor struct{}

func (e *OpenEnvExecutor) Type() define.E2EStepType { return define.E2EStepOpenEnv }

func (e *OpenEnvExecutor) Validate(step *define.E2EStep) error {
	if len(step.Config) == 0 {
		return errors.New("open_env: config cannot be empty")
	}
	var cfg struct {
		URL string `json:"url"`
	}
	_ = unmarshalJSON(step.Config, &cfg)
	if strings.TrimSpace(cfg.URL) == "" {
		return errors.New("open_env: url cannot be empty")
	}
	return nil
}

func (e *OpenEnvExecutor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	ctx.Output.Writef("[step] open_env starting")
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
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("goto %s failed: %v", url, err)}
	}
	if cfg.WaitLoad {
		ctx.Page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(timeout),
		})
	}
	ctx.Output.Writef("[step] open_env completed url=%s", url)
	return &StepResult{Success: true}
}

// ClickV1Executor click_v1: basic click.
type ClickV1Executor struct{}

func (e *ClickV1Executor) Type() define.E2EStepType { return define.E2EStepClickV1 }

func (e *ClickV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.ClickV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("click_v1: selector cannot be empty")
	}
	return nil
}

func (e *ClickV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	ctx.Output.Writef("[step] click_v1 starting")
	var cfg define.ClickV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)

	loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, cfg.TimeoutMs)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("locator failed: %v", err)}
	}
	clickOpts := playwright.LocatorClickOptions{Force: playwright.Bool(cfg.Force)}
	if cfg.ClickCount > 1 {
		clickOpts.ClickCount = playwright.Int(cfg.ClickCount)
	}
	if err := loc.Click(clickOpts); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("click failed: %v", err)}
	}
	ctx.Output.Writef("[step] click_v1 completed selector=%s", cfg.Selector)
	return &StepResult{Success: true}
}

// InputV1Executor input_v1: fixed input.
type InputV1Executor struct{}

func (e *InputV1Executor) Type() define.E2EStepType { return define.E2EStepInputV1 }

func (e *InputV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.InputV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("input_v1: selector cannot be empty")
	}
	return nil
}

func (e *InputV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	ctx.Output.Writef("[step] input_v1 starting")
	var cfg define.InputV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)
	value := ctx.Resolver.Resolve(cfg.Value)

	loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, cfg.TimeoutMs)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("locator failed: %v", err)}
	}
	if cfg.ClearBefore {
		_ = loc.Clear()
	}
	if err := loc.Fill(value); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("fill failed: %v", err)}
	}
	if cfg.PressEnter {
		_ = loc.Press("Enter")
	}
	ctx.Output.Writef("[step] input_v1 completed selector=%s", cfg.Selector)
	return &StepResult{Success: true}
}

// InputV2Executor input_v2: multiple input sources (fixed/previous output/API).
type InputV2Executor struct{}

func (e *InputV2Executor) Type() define.E2EStepType { return define.E2EStepInputV2 }

func (e *InputV2Executor) Validate(step *define.E2EStep) error {
	var cfg define.InputV2Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("input_v2: selector cannot be empty")
	}
	if cfg.SourceType == "" {
		return errors.New("input_v2: source_type cannot be empty")
	}
	switch cfg.SourceType {
	case "fixed":
		// fixed mode does not require additional validation
	case "previous_output":
		if cfg.PreviousOutput == nil || cfg.PreviousOutput.StepID == "" || cfg.PreviousOutput.Key == "" {
			return errors.New("input_v2: in previous_output mode, step_id and key cannot be empty")
		}
	case "api":
		if cfg.APIInput == nil || cfg.APIInput.URL == "" {
			return errors.New("input_v2: in api mode, url cannot be empty")
		}
	default:
		return fmt.Errorf("input_v2: unknown source_type: %s", cfg.SourceType)
	}
	return nil
}

func (e *InputV2Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	ctx.Output.Writef("[step] input_v2 starting")
	var cfg define.InputV2Config
	_ = unmarshalJSON(step.Config, &cfg)
	cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)

	var value string
	var fetchErr error

	switch cfg.SourceType {
	case "fixed":
		value = ctx.Resolver.Resolve(cfg.FixedValue)
	case "previous_output":
		if cfg.PreviousOutput != nil {
			varName := cfg.PreviousOutput.Key
			if v, ok := ctx.VarContext.Get(varName); ok {
				value = v
			} else {
				return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("input_v2: variable %s not defined", varName)}
			}
		}
	case "api":
		value, fetchErr = e.fetchAndExtract(ctx, cfg.APIInput)
		if fetchErr != nil {
			return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("input_v2: API fetch failed: %v", fetchErr)}
		}
	}

	loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, 5000)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("locator failed: %v", err)}
	}
	if cfg.ClearBefore {
		_ = loc.Clear()
	}
	if err := loc.Fill(value); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("fill failed: %v", err)}
	}
	if cfg.PressEnter {
		_ = loc.Press("Enter")
	}
	ctx.Output.Writef("[step] input_v2 completed selector=%s, source=%s", cfg.Selector, cfg.SourceType)
	return &StepResult{Success: true}
}

// fetchAndExtract sends an API request and extracts the response value.
func (e *InputV2Executor) fetchAndExtract(ctx *ExecuteContext, apiCfg *define.APIInputConfig) (string, error) {
	if apiCfg == nil {
		return "", errors.New("api_cfg cannot be empty")
	}
	if ctx.RequestRepo == nil {
		return "", errors.New("request repo is empty, ensure API calls are triggered before input_v2")
	}
	// input_v2 api mode should be used after extract_api_v1 to extract API response to variables
	return "", errors.New("input_v2 api mode: use extract_api_v1 step to extract API response to variable first")
}

// ExtractAPIV1Executor extract_api_v1: extracts data from captured API response to variable.
type ExtractAPIV1Executor struct{}

func (e *ExtractAPIV1Executor) Type() define.E2EStepType { return define.E2EStepExtractAPIV1 }

func (e *ExtractAPIV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.ExtractAPIV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if cfg.FindByURL == "" && cfg.FindByPattern == "" && cfg.FindByMethod == "" {
		return errors.New("extract_api_v1: at least 1 search condition required (find_by_url/find_by_pattern/find_by_method)")
	}
	if cfg.ResponseJSONPath == "" {
		return errors.New("extract_api_v1: response_json_path cannot be empty")
	}
	if cfg.ExtractTo == "" {
		return errors.New("extract_api_v1: extract_to cannot be empty")
	}
	return nil
}

func (e *ExtractAPIV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	ctx.Output.Writef("[step] extract_api_v1 starting")
	var cfg define.ExtractAPIV1Config
	_ = unmarshalJSON(step.Config, &cfg)

	if ctx.RequestRepo == nil {
		return &StepResult{Success: false, ErrorMsg: "extract_api_v1: request repo is empty"}
	}

	// Build match conditions
	matchCfg := &struct {
		URL      string
		Contains string
		Method   string
	}{
		URL:      cfg.FindByURL,
		Contains: cfg.FindByPattern,
		Method:   cfg.FindByMethod,
	}

	// Get all captured requests
	allReqs := ctx.RequestRepo.GetAll()

	type matchedRequest struct {
		URL   string
		Body  string
		Index int
	}
	var matched []*matchedRequest

	for i, req := range allReqs {
		// URL exact match
		if matchCfg.URL != "" && req.URL != matchCfg.URL {
			continue
		}
		// URL contains match
		if matchCfg.Contains != "" && !strings.Contains(req.URL, matchCfg.Contains) {
			continue
		}
		// HTTP method match
		if matchCfg.Method != "" && !strings.EqualFold(req.Method, matchCfg.Method) {
			continue
		}
		// Requires response
		if req.Response == nil {
			continue
		}
		matched = append(matched, &matchedRequest{req.URL, req.Response.Body, i})
	}

	if len(matched) == 0 {
		return &StepResult{Success: false, ErrorMsg: "extract_api_v1: no matching captured request found"}
	}

	// Select based on MatchIndex
	target := matched[0]
	if cfg.MatchIndex > 0 && cfg.MatchIndex <= len(matched) {
		target = matched[cfg.MatchIndex-1]
	}

	// Parse JSON and extract value
	var data any
	if err := json.Unmarshal([]byte(target.Body), &data); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("extract_api_v1: response JSON parse failed: %v", err)}
	}

	// Use extractJSONPath helper
	extractedValue := extractJSONPath(data, cfg.ResponseJSONPath)
	if extractedValue == "" && cfg.ResponseJSONPath != "" {
		ctx.Output.Writef("[step] extract_api_v1: path %s not found or value is empty", cfg.ResponseJSONPath)
	}

	ctx.VarContext.Set(cfg.ExtractTo, extractedValue)
	ctx.Output.Writef("[step] extract_api_v1 completed: %s=%s (from %s)", cfg.ExtractTo, extractedValue, target.URL)
	return &StepResult{
		Success: true,
		ExtractedVar: map[string]string{
			cfg.ExtractTo: extractedValue,
		},
	}
}

// extractJSONPath extracts value from JSON by path.
func extractJSONPath(data any, path string) string {
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
	return stringifyValue(current)
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

func stringifyValue(v any) string {
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

// HoverV1Executor hover_v1: hover.
type HoverV1Executor struct{}

func (e *HoverV1Executor) Type() define.E2EStepType { return define.E2EStepHoverV1 }

func (e *HoverV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.HoverV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("hover_v1: selector cannot be empty")
	}
	return nil
}

func (e *HoverV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.HoverV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)
	loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, cfg.TimeoutMs)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("locator failed: %v", err)}
	}
	if err := loc.Hover(); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("hover failed: %v", err)}
	}
	return &StepResult{Success: true}
}

// SelectV1Executor select_v1: dropdown select.
type SelectV1Executor struct{}

func (e *SelectV1Executor) Type() define.E2EStepType { return define.E2EStepSelectV1 }

func (e *SelectV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.SelectV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("select_v1: selector cannot be empty")
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
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("locator failed: %v", err)}
	}
	if _, err := loc.SelectOption(playwright.SelectOptionValues{Values: &[]string{value}}); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("select failed: %v", err)}
	}
	return &StepResult{Success: true}
}

// NavigateV1Executor navigate_v1: navigate to URL.
type NavigateV1Executor struct{}

func (e *NavigateV1Executor) Type() define.E2EStepType { return define.E2EStepNavigateV1 }

func (e *NavigateV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.NavigateV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.URL) == "" {
		return errors.New("navigate_v1: url cannot be empty")
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
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("goto failed: %v", err)}
	}
	if cfg.WaitLoad {
		ctx.Page.WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State:   playwright.LoadStateNetworkidle,
			Timeout: playwright.Float(timeout),
		})
	}
	return &StepResult{Success: true}
}

// WaitElementV1Executor wait_element_v1: wait for element to appear.
type WaitElementV1Executor struct{}

func (e *WaitElementV1Executor) Type() define.E2EStepType { return define.E2EStepWaitElementV1 }

func (e *WaitElementV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.WaitElementV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("wait_element_v1: selector cannot be empty")
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
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("wait failed: %v", err)}
	}
	return &StepResult{Success: true}
}

// WaitTimeoutV1Executor wait_timeout_v1: fixed wait.
type WaitTimeoutV1Executor struct{}

func (e *WaitTimeoutV1Executor) Type() define.E2EStepType { return define.E2EStepWaitTimeoutV1 }

func (e *WaitTimeoutV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.WaitTimeoutV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if cfg.DurationMs <= 0 {
		return errors.New("wait_timeout_v1: duration_ms must be greater than 0")
	}
	return nil
}

func (e *WaitTimeoutV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.WaitTimeoutV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	ctx.Page.WaitForTimeout(float64(cfg.DurationMs))
	return &StepResult{Success: true}
}

// ExtractTextV1Executor extract_text_v1: extract text to variable.
type ExtractTextV1Executor struct{}

func (e *ExtractTextV1Executor) Type() define.E2EStepType { return define.E2EStepExtractTextV1 }

func (e *ExtractTextV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.ExtractTextV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("extract_text_v1: selector cannot be empty")
	}
	if strings.TrimSpace(cfg.ExtractTo) == "" {
		return errors.New("extract_text_v1: extract_to cannot be empty")
	}
	return nil
}

func (e *ExtractTextV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.ExtractTextV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)
	loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, cfg.TimeoutMs)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("locator failed: %v", err)}
	}
	text, err := loc.TextContent()
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("get text failed: %v", err)}
	}
	out := text
	if cfg.Regex != "" {
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

// ExtractAttrV1Executor extract_attr_v1: extract attribute to variable.
type ExtractAttrV1Executor struct{}

func (e *ExtractAttrV1Executor) Type() define.E2EStepType { return define.E2EStepExtractAttrV1 }

func (e *ExtractAttrV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.ExtractAttrV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Selector) == "" {
		return errors.New("extract_attr_v1: selector cannot be empty")
	}
	if strings.TrimSpace(cfg.Attribute) == "" {
		return errors.New("extract_attr_v1: attribute cannot be empty")
	}
	return nil
}

func (e *ExtractAttrV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.ExtractAttrV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	cfg.Selector = ctx.Resolver.Resolve(cfg.Selector)
	loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, cfg.TimeoutMs)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("locator failed: %v", err)}
	}
	val, err := loc.GetAttribute(cfg.Attribute)
	if err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("get attribute failed: %v", err)}
	}
	ctx.VarContext.Set(cfg.ExtractTo, val)
	return &StepResult{Success: true, ExtractedVar: map[string]string{cfg.ExtractTo: val}}
}

// ScriptV1Executor script_v1: execute custom JS.
type ScriptV1Executor struct{}

func (e *ScriptV1Executor) Type() define.E2EStepType { return define.E2EStepScriptV1 }

func (e *ScriptV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.ScriptV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Code) == "" {
		return errors.New("script_v1: code cannot be empty")
	}
	return nil
}

func (e *ScriptV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.ScriptV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	if _, err := ctx.Page.Evaluate(cfg.Code); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("script failed: %v", err)}
	}
	return &StepResult{Success: true}
}

// GoBackV1Executor go_back_v1: browser back.
type GoBackV1Executor struct{}

func (e *GoBackV1Executor) Type() define.E2EStepType { return define.E2EStepGoBackV1 }

func (e *GoBackV1Executor) Validate(step *define.E2EStep) error { return nil }

func (e *GoBackV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	if _, err := ctx.Page.GoBack(); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("go_back failed: %v", err)}
	}
	return &StepResult{Success: true}
}

// ReloadV1Executor reload_v1: reload page.
type ReloadV1Executor struct{}

func (e *ReloadV1Executor) Type() define.E2EStepType { return define.E2EStepReloadV1 }

func (e *ReloadV1Executor) Validate(step *define.E2EStep) error { return nil }

func (e *ReloadV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	if _, err := ctx.Page.Reload(); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("reload failed: %v", err)}
	}
	return &StepResult{Success: true}
}

// PressKeyV1Executor press_key_v1: press key.
type PressKeyV1Executor struct{}

func (e *PressKeyV1Executor) Type() define.E2EStepType { return define.E2EStepPressKeyV1 }

func (e *PressKeyV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.PressKeyV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if strings.TrimSpace(cfg.Key) == "" {
		return errors.New("press_key_v1: key cannot be empty")
	}
	return nil
}

func (e *PressKeyV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.PressKeyV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	if strings.TrimSpace(cfg.Selector) != "" {
		loc, err := ApplySelector(ctx.Page, cfg.Selector, cfg.SelectorType, 5000)
		if err != nil {
			return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("locator failed: %v", err)}
		}
		if err := loc.Press(cfg.Key); err != nil {
			return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("press failed: %v", err)}
		}
		return &StepResult{Success: true}
	}
	if err := ctx.Page.Keyboard().Press(cfg.Key); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("press failed: %v", err)}
	}
	return &StepResult{Success: true}
}

// unmarshalJSON helper function.
func unmarshalJSON(raw []byte, target any) error {
	if len(raw) == 0 {
		return nil
	}
	return json.Unmarshal(raw, target)
}

// ==================== 录制专用步骤类型（v5.0 新增） ====================

// ClickByPositionV1Executor click_by_position_v1: 按坐标点击（录制工具条生成）。
// 回放时若 scale_mode=ratio 且视口尺寸变化，按比例换算。
type ClickByPositionV1Executor struct{}

func (e *ClickByPositionV1Executor) Type() define.E2EStepType { return define.E2EStepClickByPositionV1 }

func (e *ClickByPositionV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.ClickByPositionV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if cfg.ViewportWidth <= 0 || cfg.ViewportHeight <= 0 {
		return errors.New("click_by_position_v1: viewport_width/height 不能为空")
	}
	return nil
}

func (e *ClickByPositionV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.ClickByPositionV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	x, y := cfg.X, cfg.Y
	if cfg.ScaleMode == "ratio" {
		// 按比例换算到当前视口
		curW, curH := currentViewport(ctx.Page)
		if curW > 0 && curH > 0 {
			x = cfg.X * curW / cfg.ViewportWidth
			y = cfg.Y * curH / cfg.ViewportHeight
		}
	}
	button := "left"
	if cfg.Button != "" {
		button = cfg.Button
	}
	mb := playwright.MouseButton(button)
	opts := playwright.MouseClickOptions{Button: &mb}
	if cfg.ClickCount > 1 {
		opts.ClickCount = playwright.Int(cfg.ClickCount)
	}
	if err := ctx.Page.Mouse().Click(float64(x), float64(y), opts); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("坐标点击失败: %v", err)}
	}
	return &StepResult{Success: true}
}

// RightClickV1Executor right_click_v1: 右键点击（录制工具条生成）。
type RightClickV1Executor struct{}

func (e *RightClickV1Executor) Type() define.E2EStepType { return define.E2EStepRightClickV1 }

func (e *RightClickV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.RightClickV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if cfg.ViewportWidth <= 0 || cfg.ViewportHeight <= 0 {
		return errors.New("right_click_v1: viewport_width/height 不能为空")
	}
	return nil
}

func (e *RightClickV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.RightClickV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	x, y := cfg.X, cfg.Y
	if err := ctx.Page.Mouse().Click(float64(x), float64(y), playwright.MouseClickOptions{
		Button: playwright.MouseButtonRight,
	}); err != nil {
		return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("右键点击失败: %v", err)}
	}
	return &StepResult{Success: true}
}

// ScrollV1Executor scroll_v1: 滚动页面（录制工具条生成）。
type ScrollV1Executor struct{}

func (e *ScrollV1Executor) Type() define.E2EStepType { return define.E2EStepScrollV1 }

func (e *ScrollV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.ScrollV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if cfg.DeltaX == 0 && cfg.DeltaY == 0 {
		return errors.New("scroll_v1: delta_x / delta_y 至少有一个非零")
	}
	return nil
}

func (e *ScrollV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.ScrollV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	// 通过 evaluate 滚动（兼容性最好，跨设备一致）
	js := fmt.Sprintf("window.scrollBy(%d, %d);", cfg.DeltaX, cfg.DeltaY)
	if _, err := ctx.Page.Evaluate(js); err != nil {
		// 失败时尝试 Mouse.Wheel
		if wheelErr := ctx.Page.Mouse().Wheel(float64(cfg.DeltaX), float64(cfg.DeltaY)); wheelErr != nil {
			return &StepResult{Success: false, ErrorMsg: fmt.Sprintf("滚动失败: evaluate=%v wheel=%v", err, wheelErr)}
		}
	}
	return &StepResult{Success: true}
}

// WaitAfterV1Executor wait_after_v1: 步骤后等待（录制时常自动追加，用于等待异步请求 / 动画）。
type WaitAfterV1Executor struct{}

func (e *WaitAfterV1Executor) Type() define.E2EStepType { return define.E2EStepWaitAfterV1 }

func (e *WaitAfterV1Executor) Validate(step *define.E2EStep) error {
	var cfg define.WaitAfterV1Config
	if err := unmarshalJSON(step.Config, &cfg); err != nil {
		return err
	}
	if cfg.DurationMs < 0 {
		return errors.New("wait_after_v1: duration_ms 不能为负")
	}
	return nil
}

func (e *WaitAfterV1Executor) Execute(ctx *ExecuteContext, step *define.E2EStep) *StepResult {
	var cfg define.WaitAfterV1Config
	_ = unmarshalJSON(step.Config, &cfg)
	dur := cfg.DurationMs
	if dur <= 0 {
		return &StepResult{Success: true}
	}
	ctx.Page.WaitForTimeout(float64(dur))
	return &StepResult{Success: true}
}

// currentViewport 读取当前 page 的视口尺寸（缩放模式时使用）。
func currentViewport(page playwright.Page) (int, int) {
	if page == nil {
		return 0, 0
	}
	res, err := page.Evaluate("() => ({ w: window.innerWidth, h: window.innerHeight })")
	if err != nil || res == nil {
		return 0, 0
	}
	m, ok := res.(map[string]any)
	if !ok {
		return 0, 0
	}
	w, _ := m["w"].(float64)
	h, _ := m["h"].(float64)
	return int(w), int(h)
}
