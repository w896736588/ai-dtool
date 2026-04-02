package controller

import "testing"

// TestSmartLinkLocatorAutoExtractSystemPrompt 验证内置提示词包含 Playwright 提取能力说明。
// TestSmartLinkLocatorAutoExtractSystemPrompt verifies the prompt includes Playwright pick guidance.
func TestSmartLinkLocatorAutoExtractSystemPrompt(t *testing.T) {
	prompt := smartLinkLocatorAutoExtractSystemPrompt()
	if prompt == "" {
		t.Fatal("smartLinkLocatorAutoExtractSystemPrompt() should not be empty")
	}
	for _, keyword := range []string{"Playwright", "first", "last", "nth", "button_text", "placeholder"} {
		if !containsText(prompt, keyword) {
			t.Fatalf("prompt = %q, want to contain %q", prompt, keyword)
		}
	}
}

// TestParseSmartLinkLocatorAutoExtractResult 验证 AI 返回 JSON 可被清洗并解析。
// TestParseSmartLinkLocatorAutoExtractResult verifies AI JSON output can be stripped and parsed.
func TestParseSmartLinkLocatorAutoExtractResult(t *testing.T) {
	raw := "```json\n{\"locator_editor_mode\":\"simple\",\"locator_structured_form\":{\"kind\":\"text\",\"value\":\"欢迎登录\",\"target_text\":\"\",\"exact\":false,\"negate\":false,\"pick_mode\":\"first\",\"nth\":0,\"timeout_mills\":3000},\"reason\":\"\"}\n```"
	result, err := parseSmartLinkLocatorAutoExtractResult(raw)
	if err != nil {
		t.Fatalf("parseSmartLinkLocatorAutoExtractResult() error = %v", err)
	}
	if result.LocatorEditorMode != `simple` {
		t.Fatalf("LocatorEditorMode = %q, want simple", result.LocatorEditorMode)
	}
	if result.LocatorStructuredForm.Kind != `text` || result.LocatorStructuredForm.PickMode != `first` {
		t.Fatalf("LocatorStructuredForm = %#v, want text/first", result.LocatorStructuredForm)
	}
}

// TestParseSmartLinkLocatorAutoExtractResultRejectsInvalidPick 验证非法 pick_mode 会被拦截。
// TestParseSmartLinkLocatorAutoExtractResultRejectsInvalidPick verifies invalid pick_mode is rejected.
func TestParseSmartLinkLocatorAutoExtractResultRejectsInvalidPick(t *testing.T) {
	raw := "{\"locator_editor_mode\":\"simple\",\"locator_structured_form\":{\"kind\":\"css\",\"value\":\".login-btn\",\"pick_mode\":\"random\",\"nth\":0,\"timeout_mills\":3000}}"
	_, err := parseSmartLinkLocatorAutoExtractResult(raw)
	if err == nil {
		t.Fatal("parseSmartLinkLocatorAutoExtractResult() should reject invalid pick_mode")
	}
}
