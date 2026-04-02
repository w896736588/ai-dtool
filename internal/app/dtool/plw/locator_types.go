package plw

import (
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"
)

const (
	// LocatorMethodLocator 表示通过原始选择器定位元素。
	LocatorMethodLocator = `locator`

	// LocatorMethodRole 表示通过 aria role 定位元素。
	LocatorMethodRole = `role`

	// LocatorMethodText 表示通过文本内容定位元素。
	LocatorMethodText = `text`

	// LocatorMethodLabel 表示通过 label 文本定位元素。
	LocatorMethodLabel = `label`

	// LocatorMethodPlaceholder 表示通过 placeholder 定位元素。
	LocatorMethodPlaceholder = `placeholder`

	// LocatorMethodAltText 表示通过 alt 文本定位元素。
	LocatorMethodAltText = `alt_text`

	// LocatorMethodTitle 表示通过 title 文本定位元素。
	LocatorMethodTitle = `title`

	// LocatorMethodTestID 表示通过 test id 定位元素。
	LocatorMethodTestID = `test_id`
)

// LocatorInput 表示流程层接收的结构化定位配置。
type LocatorInput struct {
	// Spec 表示结构化 Locator 配置。
	Spec *LocatorSpec `json:"spec,omitempty"`
}

// LocatorSpec 表示标准化后的元素定位配置。
// 该结构统一承载 Playwright 原生查询语义，避免业务层继续依赖零散字段。
type LocatorSpec struct {
	// Method 表示标准化后的定位方式，例如 locator、role、text。
	Method string `json:"method,omitempty"`

	// FindType 表示业务配置中可能传入的定位方式别名。
	FindType string `json:"find_type,omitempty"`

	// Value 表示查询主值，例如选择器、文本或 role 名称。
	Value string `json:"value,omitempty"`

	// Options 表示查询附加参数，例如 exact、name。
	Options *LocatorOptions `json:"options,omitempty"`

	// Filters 表示当前 Locator 上的过滤条件。
	Filters []LocatorFilter `json:"filters,omitempty"`

	// Chain 表示链式子 Locator，会基于当前 Locator 继续向下查找。
	Chain []LocatorSpec `json:"chain,omitempty"`

	// Pick 表示从匹配结果中选择 first、last 或 nth。
	Pick *LocatorPick `json:"pick,omitempty"`

	// Negate 表示反向存在判断，仅用于存在性校验场景。
	Negate bool `json:"negate,omitempty"`

	// TimeoutMills 表示当前 Locator 的超时时间，单位毫秒。
	TimeoutMills float64 `json:"timeout_mills,omitempty"`

	// Name 表示业务配置里直接写在根节点上的 role name 别名。
	Name string `json:"name,omitempty"`

	// Exact 表示业务配置里直接写在根节点上的 exact 别名。
	Exact *bool `json:"exact,omitempty"`

	// Checked 表示业务配置里直接写在根节点上的 checked 别名。
	Checked *bool `json:"checked,omitempty"`

	// Disabled 表示业务配置里直接写在根节点上的 disabled 别名。
	Disabled *bool `json:"disabled,omitempty"`

	// Selected 表示业务配置里直接写在根节点上的 selected 别名。
	Selected *bool `json:"selected,omitempty"`

	// Expanded 表示业务配置里直接写在根节点上的 expanded 别名。
	Expanded *bool `json:"expanded,omitempty"`

	// IncludeHidden 表示业务配置里直接写在根节点上的 include_hidden 别名。
	IncludeHidden *bool `json:"include_hidden,omitempty"`

	// Level 表示业务配置里直接写在根节点上的 level 别名。
	Level *int `json:"level,omitempty"`

	// First 表示业务配置里直接写在根节点上的 first 别名。
	First *bool `json:"first,omitempty"`

	// Last 表示业务配置里直接写在根节点上的 last 别名。
	Last *bool `json:"last,omitempty"`

	// Nth 表示业务配置里直接写在根节点上的 nth 别名。
	Nth *int `json:"nth,omitempty"`
}

// LocatorOptions 表示 Locator 查询附加参数。
type LocatorOptions struct {
	Exact         *bool  `json:"exact,omitempty"`
	Name          string `json:"name,omitempty"`
	Checked       *bool  `json:"checked,omitempty"`
	Disabled      *bool  `json:"disabled,omitempty"`
	Selected      *bool  `json:"selected,omitempty"`
	Expanded      *bool  `json:"expanded,omitempty"`
	IncludeHidden *bool  `json:"include_hidden,omitempty"`
	Level         *int   `json:"level,omitempty"`
}

// LocatorFilter 表示对当前 Locator 结果做进一步过滤。
type LocatorFilter struct {
	HasText    string       `json:"has_text,omitempty"`
	HasNotText string       `json:"has_not_text,omitempty"`
	Has        *LocatorSpec `json:"has,omitempty"`
	HasNot     *LocatorSpec `json:"has_not,omitempty"`
	Visible    *bool        `json:"visible,omitempty"`
}

// LocatorPick 表示从匹配结果中选择特定元素。
type LocatorPick struct {
	First *bool `json:"first,omitempty"`
	Last  *bool `json:"last,omitempty"`
	Nth   *int  `json:"nth,omitempty"`
}

// ElementAction 表示元素动作定义。
type ElementAction struct {
	Type    string                `json:"type"`
	Value   string                `json:"value,omitempty"`
	Options *ElementActionOptions `json:"options,omitempty"`
}

// ElementActionOptions 表示动作附加参数。
type ElementActionOptions struct {
	TimeoutMills  float64 `json:"timeout_mills,omitempty"`
	Force         *bool   `json:"force,omitempty"`
	AttributeName string  `json:"attribute_name,omitempty"`
}

// ElementResult 表示动作执行结果。
type ElementResult struct {
	Locator        playwright.Locator
	Exists         bool
	Count          int
	TextContent    string
	InnerText      string
	AttributeValue string
}

// LocatorQueryOptions 表示解析后的查询附加参数。
type LocatorQueryOptions struct {
	Exact         *bool
	Name          string
	Checked       *bool
	Disabled      *bool
	Selected      *bool
	Expanded      *bool
	IncludeHidden *bool
	Level         *int
}

// LocatorFilterOptions 表示解析后的过滤条件。
type LocatorFilterOptions struct {
	HasText    string
	HasNotText string
	Has        locatorNode
	HasNot     locatorNode
	Visible    *bool
}

// DebugString 返回过滤配置的调试串，便于测试和日志输出。
func (h *LocatorFilterOptions) DebugString() string {
	if h == nil {
		return ``
	}
	debugList := make([]string, 0, 3)
	if h.HasText != `` {
		debugList = append(debugList, `hasText=`+h.HasText)
	}
	if h.HasNotText != `` {
		debugList = append(debugList, `hasNotText=`+h.HasNotText)
	}
	if h.Visible != nil {
		debugList = append(debugList, fmt.Sprintf(`visible=%t`, *h.Visible))
	}
	return strings.Join(debugList, `,`)
}
