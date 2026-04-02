package plw

import (
	"testing"

	"dev_tool/internal/app/dtool/define"
)

// TestDecodeBoolResultRules 验证 bool_result 规则中的 locator 会保留结构化对象。
func TestDecodeBoolResultRules(t *testing.T) {
	raw := `[{"locator":{"spec":{"method":"locator","value":".submit-btn"}},"return":true},{"locator":{"spec":{"method":"text","value":"登录成功"}},"return":false}]`

	ruleList, err := decodeBoolResultRules(raw)
	if err != nil {
		t.Fatalf("decodeBoolResultRules() error = %v", err)
	}
	if len(ruleList) != 2 {
		t.Fatalf("len(ruleList) = %d, want 2", len(ruleList))
	}
	locatorMap, ok := ruleList[0].Locator.(map[string]any)
	if !ok {
		t.Fatalf("ruleList[0].Locator type = %T, want map[string]any", ruleList[0].Locator)
	}
	spec, ok := locatorMap[`spec`].(map[string]any)
	if !ok || spec[`method`] != `locator` {
		t.Fatalf("locator spec = %#v, want method=locator", locatorMap)
	}
}

// TestNewProcessWithBoolResultLocatorList 验证 bool_result 使用规则数组时不会被当成单个 locator 解析。
func TestNewProcessWithBoolResultLocatorList(t *testing.T) {
	runParams := &PlaywrightRunParams{
		Domain:      `example.com`,
		ReplaceList: map[string]string{},
		RunCallFunc: func(define.ProcessType, string, string, string) {},
		StreamFunc:  func(string, string) {},
	}
	process := NewProcess(map[string]any{
		`name`:    `是否需要登录`,
		`type`:    string(define.BoolResult),
		`locator`: `[{"locator":{"spec":{"method":"locator","value":".user-avatar","timeout_mills":3000}},"return":false},{"locator":{"spec":{"method":"locator","value":"#basic_username","timeout_mills":3000}},"return":true}]`,
	}, nil, runParams, map[string]bool{}, map[string]string{}, nil)

	if process == nil || process.Locator == nil {
		t.Fatal("process or process.Locator = nil, want initialized bool_result process")
	}
	if process.Locator.parseErr != nil {
		t.Fatalf("process.Locator.parseErr = %v, want nil", process.Locator.parseErr)
	}
	if process.LocatorInput != nil {
		t.Fatalf("process.LocatorInput = %#v, want nil for bool_result rule list", process.LocatorInput)
	}
	if process.Locators == `` {
		t.Fatal("process.Locators = empty, want original bool_result locator rules")
	}
}
