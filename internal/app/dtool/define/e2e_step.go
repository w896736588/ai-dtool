package define

import "encoding/json"

// E2EStepType 步骤类型，命名规则：{type}_v{version}
// 同一类型可以有多个版本，如 input_v1、input_v2。
type E2EStepType string

const (
	// ===== 页面操作类 =====
	E2EStepOpenEnv      E2EStepType = "open_env"     // 打开环境
	E2EStepClickV1      E2EStepType = "click_v1"     // 点击（基础）
	E2EStepInputV1      E2EStepType = "input_v1"     // 输入（基础）
	E2EStepHoverV1      E2EStepType = "hover_v1"     // 悬停
	E2EStepSelectV1     E2EStepType = "select_v1"    // 下拉选择
	E2EStepNavigateV1   E2EStepType = "navigate_v1"  // 页面导航
	E2EStepGoBackV1     E2EStepType = "go_back_v1"   // 返回上一页
	E2EStepReloadV1     E2EStepType = "reload_v1"    // 刷新
	E2EStepPressKeyV1   E2EStepType = "press_key_v1" // 按键

	// ===== 等待类 =====
	E2EStepWaitElementV1 E2EStepType = "wait_element_v1" // 等待元素
	E2EStepWaitTimeoutV1 E2EStepType = "wait_timeout_v1" // 固定等待

	// ===== 提取类 =====
	E2EStepExtractTextV1  E2EStepType = "extract_text_v1"  // 提取文本
	E2EStepExtractAttrV1  E2EStepType = "extract_attr_v1"  // 提取属性
	E2EStepExtractAPIV1   E2EStepType = "extract_api_v1"   // 从捕获的 API 响应提取

	// ===== 脚本类 =====
	E2EStepScriptV1 E2EStepType = "script_v1" // 执行自定义 JS 脚本
)

// E2EStep 步骤定义，配置以 JSON 存储，支持任意扩展。
type E2EStep struct {
	ID          string          `json:"id"`          // 步骤唯一 ID（UUID）
	Type        E2EStepType     `json:"type"`        // 步骤类型（含版本）
	Version     string          `json:"version"`     // 配置内部版本（如 "1.0"）
	Description string          `json:"description"` // 步骤描述
	Config      json.RawMessage `json:"config"`      // 步骤特有配置
}

// E2EStepWithMeta 是带元信息的步骤（用于执行器内部）。
type E2EStepWithMeta struct {
	E2EStep
	ParsedConfig any // 执行器已解析后的配置对象（类型由执行器决定）
}

// ClickV1Config click_v1 配置：基础点击。
type ClickV1Config struct {
	Selector     string `json:"selector"`                // 选择器
	SelectorType string `json:"selector_type"`           // css/xpath/text/role
	ClickCount   int    `json:"click_count,omitempty"`  // 1=单击 2=双击
	TimeoutMs    int    `json:"timeout_ms,omitempty"`   // 等待元素出现超时
	Force        bool   `json:"force,omitempty"`        // 强制点击
}

// InputV1Config input_v1 配置：固定输入。
type InputV1Config struct {
	Selector     string `json:"selector"`           // 选择器
	SelectorType string `json:"selector_type"`      // css/xpath/text/role
	Value        string `json:"value"`              // 输入值（支持 {{var}} 变量插值）
	ClearBefore  bool   `json:"clear_before"`       // 输入前清空
	PressEnter   bool   `json:"press_enter"`        // 输入后回车
	TimeoutMs    int    `json:"timeout_ms,omitempty"`
}

// InputV2Config input_v2 配置：多输入源。与 v1 字段独立。
// 同一 step.type 字段在不同版本下含义不同：v1 关注 selector/value，v2 关注 source_type。
type InputV2Config struct {
	Selector     string `json:"selector"`
	SelectorType string `json:"selector_type"`
	ClearBefore  bool   `json:"clear_before"`
	PressEnter   bool   `json:"press_enter"`

	SourceType string `json:"source_type"` // fixed/previous_output/api

	FixedValue      string          `json:"fixed_value,omitempty"`
	PreviousOutput  *PreviousOutput `json:"previous_output,omitempty"`
	APIInput        *APIInputConfig `json:"api_input,omitempty"`
}

// PreviousOutput 从上一步提取的输出引用。
type PreviousOutput struct {
	StepID string `json:"step_id"`
	Key    string `json:"key"`
}

// APIInputConfig 从 API 响应中提取输入值。
type APIInputConfig struct {
	URL      string            `json:"url"`
	Method   string            `json:"method"`
	Headers  map[string]string `json:"headers,omitempty"`
	Body     string            `json:"body,omitempty"`
	JSONPath string            `json:"json_path"`
}

// HoverV1Config hover_v1 悬停配置。
type HoverV1Config struct {
	Selector     string `json:"selector"`
	SelectorType string `json:"selector_type"`
	TimeoutMs    int    `json:"timeout_ms,omitempty"`
}

// SelectV1Config 下拉选择配置。
type SelectV1Config struct {
	Selector     string `json:"selector"`
	SelectorType string `json:"selector_type"`
	Value        string `json:"value"`      // 选项值或文本
	By           string `json:"by,omitempty"` // value/text/index
	TimeoutMs    int    `json:"timeout_ms,omitempty"`
}

// NavigateV1Config 导航配置。
type NavigateV1Config struct {
	URL        string `json:"url"`         // 目标 URL
	URLType    string `json:"url_type"`    // env/full/relative
	WaitLoad   bool   `json:"wait_load"`   // 等待页面加载完成
	TimeoutMs  int    `json:"timeout_ms,omitempty"`
}

// WaitElementV1Config 等待元素配置。
type WaitElementV1Config struct {
	Selector     string `json:"selector"`
	SelectorType string `json:"selector_type"`
	State        string `json:"state"`              // visible/hidden/attached/detached
	TimeoutMs    int    `json:"timeout_ms"`
}

// WaitTimeoutV1Config 固定等待配置。
type WaitTimeoutV1Config struct {
	DurationMs int `json:"duration_ms"`
}

// ExtractTextV1Config 提取文本配置。
type ExtractTextV1Config struct {
	Selector     string `json:"selector"`
	SelectorType string `json:"selector_type"`
	ExtractTo    string `json:"extract_to"` // 存入变量名
	Regex        string `json:"regex,omitempty"`
	TimeoutMs    int    `json:"timeout_ms,omitempty"`
}

// ExtractAttrV1Config 提取属性配置。
type ExtractAttrV1Config struct {
	Selector     string `json:"selector"`
	SelectorType string `json:"selector_type"`
	Attribute    string `json:"attribute"` // 属性名
	ExtractTo    string `json:"extract_to"`
	TimeoutMs    int    `json:"timeout_ms,omitempty"`
}

// ExtractAPIV1Config 从捕获的 API 响应提取到变量。
type ExtractAPIV1Config struct {
	FindByURL     string `json:"find_by_url,omitempty"`
	FindByPattern string `json:"find_by_pattern,omitempty"`
	FindByMethod  string `json:"find_by_method,omitempty"`
	ResponseJSONPath string `json:"response_json_path"`
	ExtractTo     string `json:"extract_to"`
	MatchIndex    int    `json:"match_index,omitempty"`
}

// ScriptV1Config 自定义 JS 脚本。
type ScriptV1Config struct {
	Code        string `json:"code"`     // JavaScript 代码
	Description string `json:"description,omitempty"`
}

// PressKeyV1Config 按键配置。
type PressKeyV1Config struct {
	Key      string `json:"key"` // Enter/Tab/Escape 等
	Selector string `json:"selector,omitempty"`
	SelectorType string `json:"selector_type,omitempty"`
}
