package plw

import "github.com/playwright-community/playwright-go"

// playwrightPageRoot 将真实 Playwright Page 适配为 locatorRoot。
type playwrightPageRoot struct {
	page playwright.Page
}

func newPlaywrightPageRoot(page playwright.Page) locatorRoot {
	return &playwrightPageRoot{page: page}
}

func (h *playwrightPageRoot) Locator(selector string) locatorNode {
	return &playwrightLocatorNode{locator: h.page.Locator(selector)}
}

func (h *playwrightPageRoot) GetByRole(role string, options *LocatorQueryOptions) locatorNode {
	return &playwrightLocatorNode{locator: h.page.GetByRole(playwright.AriaRole(role), toPlaywrightPageRoleOptions(options))}
}

func (h *playwrightPageRoot) GetByText(text string, options *LocatorQueryOptions) locatorNode {
	return &playwrightLocatorNode{locator: h.page.GetByText(text, toPlaywrightPageTextOptions(options))}
}

func (h *playwrightPageRoot) GetByLabel(text string, options *LocatorQueryOptions) locatorNode {
	return &playwrightLocatorNode{locator: h.page.GetByLabel(text, toPlaywrightPageLabelOptions(options))}
}

func (h *playwrightPageRoot) GetByPlaceholder(text string, options *LocatorQueryOptions) locatorNode {
	return &playwrightLocatorNode{locator: h.page.GetByPlaceholder(text, toPlaywrightPagePlaceholderOptions(options))}
}

func (h *playwrightPageRoot) GetByAltText(text string, options *LocatorQueryOptions) locatorNode {
	return &playwrightLocatorNode{locator: h.page.GetByAltText(text, toPlaywrightPageAltTextOptions(options))}
}

func (h *playwrightPageRoot) GetByTitle(text string, options *LocatorQueryOptions) locatorNode {
	return &playwrightLocatorNode{locator: h.page.GetByTitle(text, toPlaywrightPageTitleOptions(options))}
}

func (h *playwrightPageRoot) GetByTestID(testID string) locatorNode {
	return &playwrightLocatorNode{locator: h.page.GetByTestId(testID)}
}

// playwrightLocatorNode 将真实 Playwright Locator 适配为 locatorNode。
type playwrightLocatorNode struct {
	locator playwright.Locator
}

func (h *playwrightLocatorNode) Locator(selector string) locatorNode {
	return &playwrightLocatorNode{locator: h.locator.Locator(selector)}
}

func (h *playwrightLocatorNode) GetByRole(role string, options *LocatorQueryOptions) locatorNode {
	return &playwrightLocatorNode{locator: h.locator.GetByRole(playwright.AriaRole(role), toPlaywrightLocatorRoleOptions(options))}
}

func (h *playwrightLocatorNode) GetByText(text string, options *LocatorQueryOptions) locatorNode {
	return &playwrightLocatorNode{locator: h.locator.GetByText(text, toPlaywrightLocatorTextOptions(options))}
}

func (h *playwrightLocatorNode) GetByLabel(text string, options *LocatorQueryOptions) locatorNode {
	return &playwrightLocatorNode{locator: h.locator.GetByLabel(text, toPlaywrightLocatorLabelOptions(options))}
}

func (h *playwrightLocatorNode) GetByPlaceholder(text string, options *LocatorQueryOptions) locatorNode {
	return &playwrightLocatorNode{locator: h.locator.GetByPlaceholder(text, toPlaywrightLocatorPlaceholderOptions(options))}
}

func (h *playwrightLocatorNode) GetByAltText(text string, options *LocatorQueryOptions) locatorNode {
	return &playwrightLocatorNode{locator: h.locator.GetByAltText(text, toPlaywrightLocatorAltTextOptions(options))}
}

func (h *playwrightLocatorNode) GetByTitle(text string, options *LocatorQueryOptions) locatorNode {
	return &playwrightLocatorNode{locator: h.locator.GetByTitle(text, toPlaywrightLocatorTitleOptions(options))}
}

func (h *playwrightLocatorNode) GetByTestID(testID string) locatorNode {
	return &playwrightLocatorNode{locator: h.locator.GetByTestId(testID)}
}

func (h *playwrightLocatorNode) Filter(options *LocatorFilterOptions) locatorNode {
	return &playwrightLocatorNode{locator: h.locator.Filter(toPlaywrightFilterOptions(options))}
}

func (h *playwrightLocatorNode) First() locatorNode {
	return &playwrightLocatorNode{locator: h.locator.First()}
}

func (h *playwrightLocatorNode) Last() locatorNode {
	return &playwrightLocatorNode{locator: h.locator.Last()}
}

func (h *playwrightLocatorNode) Nth(index int) locatorNode {
	return &playwrightLocatorNode{locator: h.locator.Nth(index)}
}

func (h *playwrightLocatorNode) WaitFor(timeoutMills float64) error {
	return h.locator.WaitFor(playwright.LocatorWaitForOptions{Timeout: playwright.Float(timeoutMills)})
}

func (h *playwrightLocatorNode) Click(options *ElementActionOptions) error {
	if options == nil {
		return h.locator.Click()
	}
	clickOptions := playwright.LocatorClickOptions{}
	if options.TimeoutMills > 0 {
		clickOptions.Timeout = playwright.Float(options.TimeoutMills)
	}
	if options.Force != nil {
		clickOptions.Force = options.Force
	}
	return h.locator.Click(clickOptions)
}

func (h *playwrightLocatorNode) Fill(value string, options *ElementActionOptions) error {
	if options == nil {
		return h.locator.Fill(value)
	}
	fillOptions := playwright.LocatorFillOptions{}
	if options.TimeoutMills > 0 {
		fillOptions.Timeout = playwright.Float(options.TimeoutMills)
	}
	if options.Force != nil {
		fillOptions.Force = options.Force
	}
	return h.locator.Fill(value, fillOptions)
}

func (h *playwrightLocatorNode) TextContent() (string, error) {
	return h.locator.TextContent()
}

func (h *playwrightLocatorNode) InnerText() (string, error) {
	return h.locator.InnerText()
}

func (h *playwrightLocatorNode) Count() (int, error) {
	return h.locator.Count()
}

func (h *playwrightLocatorNode) GetAttribute(name string) (string, error) {
	return h.locator.GetAttribute(name)
}

func (h *playwrightLocatorNode) Hover(options *ElementActionOptions) error {
	if options == nil {
		return h.locator.Hover()
	}
	hoverOptions := playwright.LocatorHoverOptions{}
	if options.TimeoutMills > 0 {
		hoverOptions.Timeout = playwright.Float(options.TimeoutMills)
	}
	if options.Force != nil {
		hoverOptions.Force = options.Force
	}
	return h.locator.Hover(hoverOptions)
}

func (h *playwrightLocatorNode) Check(options *ElementActionOptions) error {
	if options == nil {
		return h.locator.Check()
	}
	checkOptions := playwright.LocatorCheckOptions{}
	if options.TimeoutMills > 0 {
		checkOptions.Timeout = playwright.Float(options.TimeoutMills)
	}
	if options.Force != nil {
		checkOptions.Force = options.Force
	}
	return h.locator.Check(checkOptions)
}

func (h *playwrightLocatorNode) Uncheck(options *ElementActionOptions) error {
	if options == nil {
		return h.locator.Uncheck()
	}
	uncheckOptions := playwright.LocatorUncheckOptions{}
	if options.TimeoutMills > 0 {
		uncheckOptions.Timeout = playwright.Float(options.TimeoutMills)
	}
	if options.Force != nil {
		uncheckOptions.Force = options.Force
	}
	return h.locator.Uncheck(uncheckOptions)
}

func (h *playwrightLocatorNode) SelectOption(values []string, options *ElementActionOptions) ([]string, error) {
	selectValues := playwright.SelectOptionValues{
		ValuesOrLabels: &values,
	}
	if options == nil {
		return h.locator.SelectOption(selectValues)
	}
	selectOptions := playwright.LocatorSelectOptionOptions{}
	if options.TimeoutMills > 0 {
		selectOptions.Timeout = playwright.Float(options.TimeoutMills)
	}
	if options.Force != nil {
		selectOptions.Force = options.Force
	}
	return h.locator.SelectOption(selectValues, selectOptions)
}

func (h *playwrightLocatorNode) Raw() playwright.Locator {
	return h.locator
}

func toPlaywrightPageRoleOptions(options *LocatorQueryOptions) playwright.PageGetByRoleOptions {
	result := playwright.PageGetByRoleOptions{}
	if options == nil {
		return result
	}
	result.Name = options.Name
	result.Exact = options.Exact
	result.Checked = options.Checked
	result.Disabled = options.Disabled
	result.Selected = options.Selected
	result.Expanded = options.Expanded
	result.IncludeHidden = options.IncludeHidden
	result.Level = options.Level
	return result
}

func toPlaywrightPageTextOptions(options *LocatorQueryOptions) playwright.PageGetByTextOptions {
	result := playwright.PageGetByTextOptions{}
	if options != nil {
		result.Exact = options.Exact
	}
	return result
}

func toPlaywrightPageLabelOptions(options *LocatorQueryOptions) playwright.PageGetByLabelOptions {
	result := playwright.PageGetByLabelOptions{}
	if options != nil {
		result.Exact = options.Exact
	}
	return result
}

func toPlaywrightPagePlaceholderOptions(options *LocatorQueryOptions) playwright.PageGetByPlaceholderOptions {
	result := playwright.PageGetByPlaceholderOptions{}
	if options != nil {
		result.Exact = options.Exact
	}
	return result
}

func toPlaywrightPageAltTextOptions(options *LocatorQueryOptions) playwright.PageGetByAltTextOptions {
	result := playwright.PageGetByAltTextOptions{}
	if options != nil {
		result.Exact = options.Exact
	}
	return result
}

func toPlaywrightPageTitleOptions(options *LocatorQueryOptions) playwright.PageGetByTitleOptions {
	result := playwright.PageGetByTitleOptions{}
	if options != nil {
		result.Exact = options.Exact
	}
	return result
}

func toPlaywrightLocatorRoleOptions(options *LocatorQueryOptions) playwright.LocatorGetByRoleOptions {
	result := playwright.LocatorGetByRoleOptions{}
	if options == nil {
		return result
	}
	result.Name = options.Name
	result.Exact = options.Exact
	result.Checked = options.Checked
	result.Disabled = options.Disabled
	result.Selected = options.Selected
	result.Expanded = options.Expanded
	result.IncludeHidden = options.IncludeHidden
	result.Level = options.Level
	return result
}

func toPlaywrightLocatorTextOptions(options *LocatorQueryOptions) playwright.LocatorGetByTextOptions {
	result := playwright.LocatorGetByTextOptions{}
	if options != nil {
		result.Exact = options.Exact
	}
	return result
}

func toPlaywrightLocatorLabelOptions(options *LocatorQueryOptions) playwright.LocatorGetByLabelOptions {
	result := playwright.LocatorGetByLabelOptions{}
	if options != nil {
		result.Exact = options.Exact
	}
	return result
}

func toPlaywrightLocatorPlaceholderOptions(options *LocatorQueryOptions) playwright.LocatorGetByPlaceholderOptions {
	result := playwright.LocatorGetByPlaceholderOptions{}
	if options != nil {
		result.Exact = options.Exact
	}
	return result
}

func toPlaywrightLocatorAltTextOptions(options *LocatorQueryOptions) playwright.LocatorGetByAltTextOptions {
	result := playwright.LocatorGetByAltTextOptions{}
	if options != nil {
		result.Exact = options.Exact
	}
	return result
}

func toPlaywrightLocatorTitleOptions(options *LocatorQueryOptions) playwright.LocatorGetByTitleOptions {
	result := playwright.LocatorGetByTitleOptions{}
	if options != nil {
		result.Exact = options.Exact
	}
	return result
}

func toPlaywrightFilterOptions(options *LocatorFilterOptions) playwright.LocatorFilterOptions {
	if options == nil {
		return playwright.LocatorFilterOptions{}
	}
	result := playwright.LocatorFilterOptions{
		HasText:    options.HasText,
		HasNotText: options.HasNotText,
	}
	if options.Has != nil {
		result.Has = options.Has.Raw()
	}
	if options.HasNot != nil {
		result.HasNot = options.HasNot.Raw()
	}
	return result
}
