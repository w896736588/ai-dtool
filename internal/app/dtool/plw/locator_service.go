package plw

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// locatorRoot 抽象了 Locator 解析的根节点，便于测试与替换。
type locatorRoot interface {
	Locator(selector string) locatorNode
	GetByRole(role string, options *LocatorQueryOptions) locatorNode
	GetByText(text string, options *LocatorQueryOptions) locatorNode
	GetByLabel(text string, options *LocatorQueryOptions) locatorNode
	GetByPlaceholder(text string, options *LocatorQueryOptions) locatorNode
	GetByAltText(text string, options *LocatorQueryOptions) locatorNode
	GetByTitle(text string, options *LocatorQueryOptions) locatorNode
	GetByTestID(testID string) locatorNode
}

// locatorNode 抽象了运行期 Locator 能力，避免纯逻辑层强耦合完整 Playwright 接口。
type locatorNode interface {
	Locator(selector string) locatorNode
	GetByRole(role string, options *LocatorQueryOptions) locatorNode
	GetByText(text string, options *LocatorQueryOptions) locatorNode
	GetByLabel(text string, options *LocatorQueryOptions) locatorNode
	GetByPlaceholder(text string, options *LocatorQueryOptions) locatorNode
	GetByAltText(text string, options *LocatorQueryOptions) locatorNode
	GetByTitle(text string, options *LocatorQueryOptions) locatorNode
	GetByTestID(testID string) locatorNode
	Filter(options *LocatorFilterOptions) locatorNode
	First() locatorNode
	Last() locatorNode
	Nth(index int) locatorNode
	WaitFor(timeoutMills float64) error
	Click(options *ElementActionOptions) error
	Fill(value string, options *ElementActionOptions) error
	TextContent() (string, error)
	InnerText() (string, error)
	Count() (int, error)
	GetAttribute(name string) (string, error)
	Hover(options *ElementActionOptions) error
	Check(options *ElementActionOptions) error
	Uncheck(options *ElementActionOptions) error
	SelectOption(values []string, options *ElementActionOptions) ([]string, error)
	Raw() playwright.Locator
}

// LocatorService 负责串联解析、定位和动作执行。
type LocatorService struct {
	parser   *LocatorParser
	resolver *LocatorResolver
	executor *ElementActionExecutor
}

// NewLocatorService 创建 Locator 服务。
func NewLocatorService() *LocatorService {
	return &LocatorService{
		parser:   NewLocatorParser(),
		resolver: NewLocatorResolver(),
		executor: NewElementActionExecutor(),
	}
}

// ParseInputValue 根据流程原始配置构建 LocatorInput，并返回日志展示串。
func (h *LocatorService) ParseInputValue(raw any) (*LocatorInput, string, error) {
	if raw == nil {
		return &LocatorInput{}, ``, nil
	}
	if text, ok := raw.(string); ok {
		trimmedText := strings.TrimSpace(text)
		if trimmedText == `` {
			return &LocatorInput{}, ``, nil
		}
		if !strings.HasPrefix(trimmedText, `{`) {
			return nil, ``, errors.New(`locator 必须是结构化 JSON 对象`)
		}
		input, err := h.decodeStructuredInput([]byte(trimmedText))
		if err != nil {
			return nil, ``, err
		}
		return input, trimmedText, nil
	}
	displayBytes, err := json.Marshal(raw)
	if err != nil {
		return nil, ``, err
	}
	display := string(displayBytes)
	input, err := h.decodeStructuredInput(displayBytes)
	if err != nil {
		return nil, ``, err
	}
	return input, display, nil
}

// decodeStructuredInput 将 JSON 内容解码成结构化 LocatorInput。
func (h *LocatorService) decodeStructuredInput(payload []byte) (*LocatorInput, error) {
	input := &LocatorInput{}
	if err := json.Unmarshal(payload, input); err == nil && input.Spec != nil {
		return input, nil
	}

	spec := &LocatorSpec{}
	if err := json.Unmarshal(payload, spec); err != nil {
		return nil, err
	}
	if spec.Method == `` && spec.FindType == `` && spec.Value == `` && len(spec.Chain) == 0 && len(spec.Filters) == 0 && spec.Pick == nil {
		return nil, errors.New(`locator 必须是结构化 JSON 对象`)
	}
	return &LocatorInput{Spec: spec}, nil
}

// FindAndExecute 根据页面和输入配置执行 Locator 解析与动作。
func (h *LocatorService) FindAndExecute(page *playwright.Page, input *LocatorInput, action *ElementAction, waitMills float64) (*ElementResult, error) {
	if page == nil || *page == nil {
		return nil, errors.New(`page 不能为空`)
	}
	return h.FindAndExecuteWithRoot(newPlaywrightPageRoot(*page), input, action, waitMills)
}

// FindAndExecuteWithRoot 允许在测试中直接传入假 root，避免依赖真实 Playwright 运行时。
func (h *LocatorService) FindAndExecuteWithRoot(root locatorRoot, input *LocatorInput, action *ElementAction, waitMills float64) (*ElementResult, error) {
	spec, err := h.parser.Parse(input)
	if err != nil {
		return nil, err
	}
	node, err := h.resolver.Resolve(root, spec)
	if err != nil {
		return nil, err
	}
	timeoutMills := waitMills
	if spec.TimeoutMills > 0 {
		timeoutMills = spec.TimeoutMills
	}
	if timeoutMills > 0 {
		waitErr := node.WaitFor(timeoutMills)
		if spec.Negate {
			if waitErr != nil {
				return &ElementResult{
					Locator: node.Raw(),
					Exists:  true,
				}, nil
			}
			return nil, errors.New(`找到了反找元素，返回失败`)
		}
		if waitErr != nil {
			return nil, waitErr
		}
	}
	if action == nil {
		return &ElementResult{Locator: node.Raw()}, nil
	}
	result, err := h.executor.Execute(node, action)
	if err != nil {
		return nil, err
	}
	result.Locator = node.Raw()
	if action.Type == `exist` {
		result.Exists = true
	}
	return result, nil
}

// LocatorResolver 负责将 LocatorSpec 转换成运行时 Locator。
type LocatorResolver struct{}

// NewLocatorResolver 创建 Locator 解析执行器。
func NewLocatorResolver() *LocatorResolver {
	return &LocatorResolver{}
}

// Resolve 将标准 LocatorSpec 转换为运行时 Locator 节点。
func (h *LocatorResolver) Resolve(root locatorRoot, spec *LocatorSpec) (locatorNode, error) {
	if root == nil {
		return nil, errors.New(`locator root 不能为空`)
	}
	if spec == nil {
		return nil, errors.New(`locator spec 不能为空`)
	}
	current, err := h.resolveRootMethod(root, spec)
	if err != nil {
		return nil, err
	}
	return h.resolveNode(current, spec)
}

func (h *LocatorResolver) resolveRootMethod(root locatorRoot, spec *LocatorSpec) (locatorNode, error) {
	options := toLocatorQueryOptions(spec.Options)
	switch spec.Method {
	case LocatorMethodLocator:
		return root.Locator(spec.Value), nil
	case LocatorMethodRole:
		return root.GetByRole(spec.Value, options), nil
	case LocatorMethodText:
		return root.GetByText(spec.Value, options), nil
	case LocatorMethodLabel:
		return root.GetByLabel(spec.Value, options), nil
	case LocatorMethodPlaceholder:
		return root.GetByPlaceholder(spec.Value, options), nil
	case LocatorMethodAltText:
		return root.GetByAltText(spec.Value, options), nil
	case LocatorMethodTitle:
		return root.GetByTitle(spec.Value, options), nil
	case LocatorMethodTestID:
		return root.GetByTestID(spec.Value), nil
	default:
		return nil, errors.New(`不支持的 locator method`)
	}
}

func (h *LocatorResolver) resolveNode(current locatorNode, spec *LocatorSpec) (locatorNode, error) {
	filtered, err := h.applyFilters(current, spec.Filters)
	if err != nil {
		return nil, err
	}
	for _, item := range spec.Chain {
		next, chainErr := h.resolveChild(filtered, &item)
		if chainErr != nil {
			return nil, chainErr
		}
		filtered = next
	}
	return applyPick(filtered, spec.Pick), nil
}

func (h *LocatorResolver) resolveChild(parent locatorNode, spec *LocatorSpec) (locatorNode, error) {
	options := toLocatorQueryOptions(spec.Options)
	var current locatorNode
	switch spec.Method {
	case LocatorMethodLocator:
		current = parent.Locator(spec.Value)
	case LocatorMethodRole:
		current = parent.GetByRole(spec.Value, options)
	case LocatorMethodText:
		current = parent.GetByText(spec.Value, options)
	case LocatorMethodLabel:
		current = parent.GetByLabel(spec.Value, options)
	case LocatorMethodPlaceholder:
		current = parent.GetByPlaceholder(spec.Value, options)
	case LocatorMethodAltText:
		current = parent.GetByAltText(spec.Value, options)
	case LocatorMethodTitle:
		current = parent.GetByTitle(spec.Value, options)
	case LocatorMethodTestID:
		current = parent.GetByTestID(spec.Value)
	default:
		return nil, errors.New(`不支持的子 locator method`)
	}
	return h.resolveNode(current, spec)
}

func (h *LocatorResolver) applyFilters(current locatorNode, filters []LocatorFilter) (locatorNode, error) {
	filtered := current
	for _, item := range filters {
		options := &LocatorFilterOptions{
			HasText:    item.HasText,
			HasNotText: item.HasNotText,
			Visible:    item.Visible,
		}
		if item.Has != nil {
			child, err := h.Resolve(locatorNodeAsRoot{node: filtered}, item.Has)
			if err != nil {
				return nil, err
			}
			options.Has = child
		}
		if item.HasNot != nil {
			child, err := h.Resolve(locatorNodeAsRoot{node: filtered}, item.HasNot)
			if err != nil {
				return nil, err
			}
			options.HasNot = child
		}
		filtered = filtered.Filter(options)
	}
	return filtered, nil
}

// locatorNodeAsRoot 将已有 Locator 节点提升为子查询根节点。
type locatorNodeAsRoot struct {
	node locatorNode
}

func (h locatorNodeAsRoot) Locator(selector string) locatorNode {
	return h.node.Locator(selector)
}

func (h locatorNodeAsRoot) GetByRole(role string, options *LocatorQueryOptions) locatorNode {
	return h.node.GetByRole(role, options)
}

func (h locatorNodeAsRoot) GetByText(text string, options *LocatorQueryOptions) locatorNode {
	return h.node.GetByText(text, options)
}

func (h locatorNodeAsRoot) GetByLabel(text string, options *LocatorQueryOptions) locatorNode {
	return h.node.GetByLabel(text, options)
}

func (h locatorNodeAsRoot) GetByPlaceholder(text string, options *LocatorQueryOptions) locatorNode {
	return h.node.GetByPlaceholder(text, options)
}

func (h locatorNodeAsRoot) GetByAltText(text string, options *LocatorQueryOptions) locatorNode {
	return h.node.GetByAltText(text, options)
}

func (h locatorNodeAsRoot) GetByTitle(text string, options *LocatorQueryOptions) locatorNode {
	return h.node.GetByTitle(text, options)
}

func (h locatorNodeAsRoot) GetByTestID(testID string) locatorNode {
	return h.node.GetByTestID(testID)
}

func applyPick(current locatorNode, pick *LocatorPick) locatorNode {
	if pick == nil {
		return current
	}
	if pick.First != nil && *pick.First {
		return current.First()
	}
	if pick.Last != nil && *pick.Last {
		return current.Last()
	}
	if pick.Nth != nil {
		return current.Nth(*pick.Nth)
	}
	return current
}

func toLocatorQueryOptions(options *LocatorOptions) *LocatorQueryOptions {
	if options == nil {
		return nil
	}
	return &LocatorQueryOptions{
		Exact:         options.Exact,
		Name:          options.Name,
		Checked:       options.Checked,
		Disabled:      options.Disabled,
		Selected:      options.Selected,
		Expanded:      options.Expanded,
		IncludeHidden: options.IncludeHidden,
		Level:         options.Level,
	}
}

// ElementActionExecutor 负责执行元素动作。
type ElementActionExecutor struct{}

// NewElementActionExecutor 创建动作执行器。
func NewElementActionExecutor() *ElementActionExecutor {
	return &ElementActionExecutor{}
}

// Execute 根据动作类型执行具体元素操作。
func (h *ElementActionExecutor) Execute(node locatorNode, action *ElementAction) (*ElementResult, error) {
	if node == nil {
		return nil, errors.New(`locator 不能为空`)
	}
	if action == nil {
		return &ElementResult{Locator: node.Raw()}, nil
	}
	result := &ElementResult{Locator: node.Raw()}
	switch action.Type {
	case `click`:
		return result, node.Click(action.Options)
	case `input`:
		return result, node.Fill(action.Value, action.Options)
	case `exist`:
		result.Exists = true
		return result, nil
	case `text_content`:
		content, err := node.TextContent()
		result.TextContent = strings.TrimSpace(content)
		return result, err
	case `inner_text`:
		content, err := node.InnerText()
		result.InnerText = strings.TrimSpace(content)
		return result, err
	case `count`:
		count, err := node.Count()
		result.Count = count
		return result, err
	case `get_attribute`:
		if action.Options == nil || strings.TrimSpace(action.Options.AttributeName) == `` {
			return nil, errors.New(`attribute_name 不能为空`)
		}
		value, err := node.GetAttribute(action.Options.AttributeName)
		result.AttributeValue = value
		return result, err
	case `hover`:
		return result, node.Hover(action.Options)
	case `check`:
		return result, node.Check(action.Options)
	case `uncheck`:
		return result, node.Uncheck(action.Options)
	case `select_option`:
		valueList := splitActionValues(action.Value)
		_, err := node.SelectOption(valueList, action.Options)
		return result, err
	default:
		return nil, errors.New(`不支持的操作`)
	}
}

func splitActionValues(raw string) []string {
	if strings.TrimSpace(raw) == `` {
		return nil
	}
	partList := strings.Split(raw, `,`)
	valueList := make([]string, 0, len(partList))
	for _, item := range partList {
		trimmed := strings.TrimSpace(item)
		if trimmed == `` {
			continue
		}
		valueList = append(valueList, trimmed)
	}
	return valueList
}
