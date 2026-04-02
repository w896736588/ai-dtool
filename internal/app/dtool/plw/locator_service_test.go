package plw

import (
	"errors"
	"strconv"
	"testing"

	"github.com/playwright-community/playwright-go"
)

// TestLocatorServiceParseInputValueWithJSONString 验证 JSON 字符串形式的结构化 locator 能被正确回填解析。
func TestLocatorServiceParseInputValueWithJSONString(t *testing.T) {
	service := NewLocatorService()

	input, display, err := service.ParseInputValue(`{"spec":{"method":"role","value":"button","options":{"name":"提交"},"pick":{"first":true}}}`)
	if err != nil {
		t.Fatalf("ParseInputValue() error = %v", err)
	}
	if display != `{"spec":{"method":"role","value":"button","options":{"name":"提交"},"pick":{"first":true}}}` {
		t.Fatalf("display = %q, want original json string", display)
	}
	if input == nil || input.Spec == nil {
		t.Fatal("input.Spec = nil, want parsed locator spec")
	}
	if input.Spec.Method != `role` || input.Spec.Value != `button` {
		t.Fatalf("input.Spec = %#v, want role/button", input.Spec)
	}
	if input.Spec.Options == nil || input.Spec.Options.Name != `提交` {
		t.Fatalf("input.Spec.Options = %#v, want Name=提交", input.Spec.Options)
	}
	if input.Spec.Pick == nil || input.Spec.Pick.First == nil || !*input.Spec.Pick.First {
		t.Fatalf("input.Spec.Pick = %#v, want first=true", input.Spec.Pick)
	}
}

// TestLocatorServiceParseInputValueRejectLegacyString 验证旧字符串 selector 不再被当成 locator 配置。
func TestLocatorServiceParseInputValueRejectLegacyString(t *testing.T) {
	service := NewLocatorService()

	_, _, err := service.ParseInputValue(`#basic_username`)
	if err == nil {
		t.Fatal("ParseInputValue() error = nil, want structured locator error")
	}
}

// TestLocatorServiceFindAndExecute 验证结构化 Locator 能完成链式解析与文本提取。
func TestLocatorServiceFindAndExecute(t *testing.T) {
	service := NewLocatorService()
	root := newFakeLocatorRoot()
	root.textContent = "  已提交  "

	result, err := service.FindAndExecuteWithRoot(root, &LocatorInput{
		Spec: &LocatorSpec{
			Method: `role`,
			Value:  `button`,
			Name:   `提交`,
			Chain: []LocatorSpec{
				{
					Method: `locator`,
					Value:  `.label`,
				},
			},
		},
	}, &ElementAction{
		Type: defineElementTextContent,
	}, 1200)
	if err != nil {
		t.Fatalf("FindAndExecuteWithRoot() error = %v", err)
	}
	if result.TextContent != `已提交` {
		t.Fatalf("result.TextContent = %q, want %q", result.TextContent, `已提交`)
	}
	if got := root.lastPath(); got != `root.getByRole(button:提交).locator(.label)` {
		t.Fatalf("root.lastPath() = %q, want %q", got, `root.getByRole(button:提交).locator(.label)`)
	}
	if root.waitTimeout != 1200 {
		t.Fatalf("root.waitTimeout = %v, want 1200", root.waitTimeout)
	}
}

// TestLocatorServiceNegate 验证反向存在判断会把未找到元素视为成功。
func TestLocatorServiceNegate(t *testing.T) {
	service := NewLocatorService()
	root := newFakeLocatorRoot()
	root.waitErr = errors.New("not found")

	result, err := service.FindAndExecuteWithRoot(root, &LocatorInput{
		Spec: &LocatorSpec{
			Method: `text`,
			Value:  `系统异常`,
			Negate: true,
		},
	}, &ElementAction{
		Type: defineElementExist,
	}, 500)
	if err != nil {
		t.Fatalf("FindAndExecuteWithRoot() error = %v, want nil", err)
	}
	if !result.Exists {
		t.Fatal("result.Exists = false, want true")
	}
}

// TestLocatorActionExecutorPickAction 验证动作执行器会调用预期的 Locator 方法。
func TestLocatorActionExecutorPickAction(t *testing.T) {
	executor := NewElementActionExecutor()
	node := &fakeLocatorNode{
		textContent: `  内容  `,
		count:       3,
	}

	textResult, err := executor.Execute(node, &ElementAction{Type: defineElementTextContent})
	if err != nil {
		t.Fatalf("Execute(text_content) error = %v", err)
	}
	if textResult.TextContent != `内容` {
		t.Fatalf("textResult.TextContent = %q, want %q", textResult.TextContent, `内容`)
	}

	countResult, err := executor.Execute(node, &ElementAction{Type: defineElementCount})
	if err != nil {
		t.Fatalf("Execute(count) error = %v", err)
	}
	if countResult.Count != 3 {
		t.Fatalf("countResult.Count = %d, want 3", countResult.Count)
	}

	if _, err = executor.Execute(node, &ElementAction{Type: `unknown`}); err == nil {
		t.Fatal("Execute(unknown) error = nil, want unsupported action error")
	}
}

const (
	// defineElementTextContent 用于测试文本提取动作，避免循环依赖业务 define 包。
	defineElementTextContent = `text_content`

	// defineElementExist 用于测试存在性动作。
	defineElementExist = `exist`

	// defineElementCount 用于测试计数动作。
	defineElementCount = `count`
)

// fakeLocatorRoot 用于测试 Locator 解析与执行链路。
type fakeLocatorRoot struct {
	current     *fakeLocatorNode
	textContent string
	waitErr     error
	waitTimeout float64
}

func newFakeLocatorRoot() *fakeLocatorRoot {
	root := &fakeLocatorRoot{}
	root.current = &fakeLocatorNode{
		path:        `root`,
		textContent: root.textContent,
		root:        root,
	}
	return root
}

func (h *fakeLocatorRoot) Locator(selector string) locatorNode {
	h.current = &fakeLocatorNode{path: h.current.path + `.locator(` + selector + `)`, textContent: h.textContent, root: h}
	return h.current
}

func (h *fakeLocatorRoot) GetByRole(role string, options *LocatorQueryOptions) locatorNode {
	name := ``
	if options != nil {
		name = options.Name
	}
	h.current = &fakeLocatorNode{path: h.current.path + `.getByRole(` + role + `:` + name + `)`, textContent: h.textContent, root: h}
	return h.current
}

func (h *fakeLocatorRoot) GetByText(text string, options *LocatorQueryOptions) locatorNode {
	h.current = &fakeLocatorNode{path: h.current.path + `.getByText(` + text + `)`, textContent: h.textContent, root: h}
	return h.current
}

func (h *fakeLocatorRoot) GetByLabel(text string, options *LocatorQueryOptions) locatorNode {
	h.current = &fakeLocatorNode{path: h.current.path + `.getByLabel(` + text + `)`, textContent: h.textContent, root: h}
	return h.current
}

func (h *fakeLocatorRoot) GetByPlaceholder(text string, options *LocatorQueryOptions) locatorNode {
	h.current = &fakeLocatorNode{path: h.current.path + `.getByPlaceholder(` + text + `)`, textContent: h.textContent, root: h}
	return h.current
}

func (h *fakeLocatorRoot) GetByAltText(text string, options *LocatorQueryOptions) locatorNode {
	h.current = &fakeLocatorNode{path: h.current.path + `.getByAltText(` + text + `)`, textContent: h.textContent, root: h}
	return h.current
}

func (h *fakeLocatorRoot) GetByTitle(text string, options *LocatorQueryOptions) locatorNode {
	h.current = &fakeLocatorNode{path: h.current.path + `.getByTitle(` + text + `)`, textContent: h.textContent, root: h}
	return h.current
}

func (h *fakeLocatorRoot) GetByTestID(testID string) locatorNode {
	h.current = &fakeLocatorNode{path: h.current.path + `.getByTestID(` + testID + `)`, textContent: h.textContent, root: h}
	return h.current
}

func (h *fakeLocatorRoot) lastPath() string {
	return h.current.path
}

// fakeLocatorNode 用于替代真实 Playwright Locator，便于验证执行链路。
type fakeLocatorNode struct {
	path        string
	textContent string
	count       int
	waitErr     error
	root        *fakeLocatorRoot
}

func (h *fakeLocatorNode) Locator(selector string) locatorNode {
	node := &fakeLocatorNode{path: h.path + `.locator(` + selector + `)`, textContent: h.textContent, count: h.count, root: h.root}
	if h.root != nil {
		h.root.current = node
	}
	return node
}

func (h *fakeLocatorNode) GetByRole(role string, options *LocatorQueryOptions) locatorNode {
	name := ``
	if options != nil {
		name = options.Name
	}
	node := &fakeLocatorNode{path: h.path + `.getByRole(` + role + `:` + name + `)`, textContent: h.textContent, count: h.count, root: h.root}
	if h.root != nil {
		h.root.current = node
	}
	return node
}

func (h *fakeLocatorNode) GetByText(text string, options *LocatorQueryOptions) locatorNode {
	node := &fakeLocatorNode{path: h.path + `.getByText(` + text + `)`, textContent: h.textContent, count: h.count, root: h.root}
	if h.root != nil {
		h.root.current = node
	}
	return node
}

func (h *fakeLocatorNode) GetByLabel(text string, options *LocatorQueryOptions) locatorNode {
	node := &fakeLocatorNode{path: h.path + `.getByLabel(` + text + `)`, textContent: h.textContent, count: h.count, root: h.root}
	if h.root != nil {
		h.root.current = node
	}
	return node
}

func (h *fakeLocatorNode) GetByPlaceholder(text string, options *LocatorQueryOptions) locatorNode {
	node := &fakeLocatorNode{path: h.path + `.getByPlaceholder(` + text + `)`, textContent: h.textContent, count: h.count, root: h.root}
	if h.root != nil {
		h.root.current = node
	}
	return node
}

func (h *fakeLocatorNode) GetByAltText(text string, options *LocatorQueryOptions) locatorNode {
	node := &fakeLocatorNode{path: h.path + `.getByAltText(` + text + `)`, textContent: h.textContent, count: h.count, root: h.root}
	if h.root != nil {
		h.root.current = node
	}
	return node
}

func (h *fakeLocatorNode) GetByTitle(text string, options *LocatorQueryOptions) locatorNode {
	node := &fakeLocatorNode{path: h.path + `.getByTitle(` + text + `)`, textContent: h.textContent, count: h.count, root: h.root}
	if h.root != nil {
		h.root.current = node
	}
	return node
}

func (h *fakeLocatorNode) GetByTestID(testID string) locatorNode {
	node := &fakeLocatorNode{path: h.path + `.getByTestID(` + testID + `)`, textContent: h.textContent, count: h.count, root: h.root}
	if h.root != nil {
		h.root.current = node
	}
	return node
}

func (h *fakeLocatorNode) Filter(options *LocatorFilterOptions) locatorNode {
	filterDesc := ``
	if options != nil {
		filterDesc = options.DebugString()
	}
	node := &fakeLocatorNode{path: h.path + `.filter(` + filterDesc + `)`, textContent: h.textContent, count: h.count, root: h.root}
	if h.root != nil {
		h.root.current = node
	}
	return node
}

func (h *fakeLocatorNode) First() locatorNode {
	node := &fakeLocatorNode{path: h.path + `.first()`, textContent: h.textContent, count: h.count, root: h.root}
	if h.root != nil {
		h.root.current = node
	}
	return node
}

func (h *fakeLocatorNode) Last() locatorNode {
	node := &fakeLocatorNode{path: h.path + `.last()`, textContent: h.textContent, count: h.count, root: h.root}
	if h.root != nil {
		h.root.current = node
	}
	return node
}

func (h *fakeLocatorNode) Nth(index int) locatorNode {
	node := &fakeLocatorNode{path: h.path + `.nth(` + strconv.Itoa(index) + `)`, textContent: h.textContent, count: h.count, root: h.root}
	if h.root != nil {
		h.root.current = node
	}
	return node
}

func (h *fakeLocatorNode) WaitFor(timeoutMills float64) error {
	if h.root != nil {
		h.root.waitTimeout = timeoutMills
		if h.root.waitErr != nil {
			return h.root.waitErr
		}
	}
	return h.waitErr
}

func (h *fakeLocatorNode) Click(options *ElementActionOptions) error {
	return nil
}

func (h *fakeLocatorNode) Fill(value string, options *ElementActionOptions) error {
	return nil
}

func (h *fakeLocatorNode) TextContent() (string, error) {
	return h.textContent, nil
}

func (h *fakeLocatorNode) InnerText() (string, error) {
	return h.textContent, nil
}

func (h *fakeLocatorNode) Count() (int, error) {
	return h.count, nil
}

func (h *fakeLocatorNode) GetAttribute(name string) (string, error) {
	return ``, nil
}

func (h *fakeLocatorNode) Hover(options *ElementActionOptions) error {
	return nil
}

func (h *fakeLocatorNode) Check(options *ElementActionOptions) error {
	return nil
}

func (h *fakeLocatorNode) Uncheck(options *ElementActionOptions) error {
	return nil
}

func (h *fakeLocatorNode) SelectOption(values []string, options *ElementActionOptions) ([]string, error) {
	return values, nil
}

func (h *fakeLocatorNode) Raw() playwright.Locator {
	return nil
}
