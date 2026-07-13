package define

import "encoding/json"

// E2EAssertionType 断言类型，命名规则：{type}_v{version}。
type E2EAssertionType string

const (
	// ===== 页面断言 =====
	E2EAssertTextV1       E2EAssertionType = "assert_text_v1"       // 文本断言
	E2EAssertTextV2       E2EAssertionType = "assert_text_v2"       // 文本断言（增强：忽略大小写、空格规范化）
	E2EAssertElementV1    E2EAssertionType = "assert_element_v1"    // 元素断言
	E2EAssertURLV1        E2EAssertionType = "assert_url_v1"        // URL 断言
	E2EAssertTitleV1      E2EAssertionType = "assert_title_v1"      // 标题断言

	// ===== API 断言（基于捕获的请求） =====
	E2EAssertAPIResponseV1  E2EAssertionType = "assert_api_response_v1"  // API 响应断言
	E2EAssertAPIRequestV1   E2EAssertionType = "assert_api_request_v1"   // API 请求断言

	// ===== 变量断言 =====
	E2EAssertVariableV1 E2EAssertionType = "assert_variable_v1" // 变量值断言
)

// E2EAssertion 断言定义。
type E2EAssertion struct {
	ID          string             `json:"id"`           // 断言唯一 ID
	Type        E2EAssertionType   `json:"type"`         // 断言类型（含版本）
	Version     string             `json:"version"`      // 配置内部版本
	Description string             `json:"description"`  // 描述
	StepID      string             `json:"step_id"`      // 关联到哪个步骤
	Config      json.RawMessage    `json:"config"`       // 断言特有配置
}

// TextAssertionV1Config assert_text_v1 配置。
type TextAssertionV1Config struct {
	Selector     string `json:"selector,omitempty"`
	SelectorType string `json:"selector_type,omitempty"`
	Operator     string `json:"operator"` // exact/contains/regex/not_contains
	Value        string `json:"value"`
}

// TextAssertionV2Config assert_text_v2 配置（与 v1 完全独立）。
type TextAssertionV2Config struct {
	Selector     string `json:"selector,omitempty"`
	SelectorType string `json:"selector_type,omitempty"`
	Operator     string `json:"operator"`
	Value        string `json:"value"`
	IgnoreCase   bool   `json:"ignore_case"`
	NormalizeWS  bool   `json:"normalize_ws"`
}

// ElementAssertionV1Config assert_element_v1 配置。
type ElementAssertionV1Config struct {
	Selector     string `json:"selector"`
	SelectorType string `json:"selector_type"`
	State        string `json:"state"`          // present/absent/visible/hidden/enabled
	Count        *int   `json:"count,omitempty"` // 期望数量，nil=至少 1
}

// URLAssertionV1Config assert_url_v1 配置。
type URLAssertionV1Config struct {
	Type        string `json:"type"`         // exact/contains/regex
	Value       string `json:"value"`
	IgnoreQuery bool   `json:"ignore_query"`
}

// TitleAssertionV1Config assert_title_v1 配置。
type TitleAssertionV1Config struct {
	Operator string `json:"operator"` // exact/contains/regex
	Value    string `json:"value"`
}

// APIResponseAssertionV1Config assert_api_response_v1 配置。
type APIResponseAssertionV1Config struct {
	FindByURL             string `json:"find_by_url,omitempty"`
	FindByPattern         string `json:"find_by_pattern,omitempty"`
	FindByMethod          string `json:"find_by_method,omitempty"`
	FindByResponseContains string `json:"find_by_response_contains,omitempty"`

	MatchIndex int `json:"match_index,omitempty"` // 1-based; 0 = 第一个
	MatchAll   bool `json:"match_all,omitempty"`

	// 断言内容（至少填一个）
	ResponseStatus    int    `json:"response_status,omitempty"`
	ResponseContains  string `json:"response_contains,omitempty"`
	ResponseJSONPath  string `json:"response_json_path,omitempty"`
	ExpectedValue     any    `json:"expected_value,omitempty"`
	IgnoreCase        bool   `json:"ignore_case,omitempty"`
}

// APIRequestAssertionV1Config assert_api_request_v1 配置。
type APIRequestAssertionV1Config struct {
	FindByURL     string `json:"find_by_url,omitempty"`
	FindByPattern string `json:"find_by_pattern,omitempty"`
	FindByMethod  string `json:"find_by_method,omitempty"`

	MatchIndex int  `json:"match_index,omitempty"`
	MatchAll   bool `json:"match_all,omitempty"`

	RequestHeaderName  string `json:"request_header_name,omitempty"`
	RequestHeaderValue string `json:"request_header_value,omitempty"`
	RequestBodyJSONPath string `json:"request_body_json_path,omitempty"`
	RequestBodyExpected any    `json:"request_body_expected,omitempty"`
}

// VariableAssertionV1Config 变量断言。
type VariableAssertionV1Config struct {
	VarName  string `json:"var_name"`
	Operator string `json:"operator"` // eq/contains/regex/gt/lt
	Expected string `json:"expected"`
}
