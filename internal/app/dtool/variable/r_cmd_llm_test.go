package variable

import "testing"

// TestParseLlmRunConfig_ValidJson 测试合法配置解析
func TestParseLlmRunConfig_ValidJson(t *testing.T) {
	cmd := map[string]any{
		"options": `{"model":"gpt-4o-mini","temperature":0.3}`,
		"bash":    "请总结今天的变更",
	}
	cfg, err := parseLlmRunConfig(cmd)
	if err != nil {
		t.Fatalf("期望解析成功，实际失败: %v", err)
	}
	if cfg.Model != "gpt-4o-mini" {
		t.Fatalf("model 不正确，got=%s", cfg.Model)
	}
	if cfg.Prompt != "请总结今天的变更" {
		t.Fatalf("prompt 不正确，got=%s", cfg.Prompt)
	}
}

// TestParseLlmRunConfig_EmptyPrompt 测试空提示词校验
func TestParseLlmRunConfig_EmptyPrompt(t *testing.T) {
	cmd := map[string]any{
		"options": `{"model":"gpt-4o-mini"}`,
		"bash":    "",
	}
	_, err := parseLlmRunConfig(cmd)
	if err == nil {
		t.Fatalf("期望提示词为空时报错")
	}
}

// TestParseLlmRunConfig_InvalidJson 测试非法 JSON 配置
func TestParseLlmRunConfig_InvalidJson(t *testing.T) {
	cmd := map[string]any{
		"options": `{"model":`,
		"bash":    "hello",
	}
	_, err := parseLlmRunConfig(cmd)
	if err == nil {
		t.Fatalf("期望 options 非法 JSON 时报错")
	}
}
