package controller

import "testing"

// TestEnsureCreateApiOptionalFieldsDefaults 中文：创建接口时缺少可选 JSON 字段也应补齐默认值，避免复制接口时报格式错误。
// English: Creating an API should backfill optional JSON fields so copy flows do not fail on missing payload keys.
func TestEnsureCreateApiOptionalFieldsDefaults(t *testing.T) {
	updateData := map[string]any{
		`collection_id`: 4,
		`folder_id`:     99,
		`name`:          `工作流聊天测试（SSE）-复制`,
		`method`:        `POST`,
		`url`:           `$Url$/chat/callWorkFlowDialog`,
	}

	ensureCreateApiOptionalFieldsDefaults(updateData)

	if got := updateData[`query_params`]; got != `[]` {
		t.Fatalf("query_params default mismatch, got %v", got)
	}
	if got := updateData[`headers`]; got != `{}` {
		t.Fatalf("headers default mismatch, got %v", got)
	}
	if got := updateData[`body_form`]; got != `[]` {
		t.Fatalf("body_form default mismatch, got %v", got)
	}
}

// TestEnsureCreateApiOptionalFieldsDefaultsKeepsExisting 中文：已有字段值不应被默认值覆盖。
// English: Existing optional JSON fields should remain untouched.
func TestEnsureCreateApiOptionalFieldsDefaultsKeepsExisting(t *testing.T) {
	updateData := map[string]any{
		`query_params`: `[{"field":"page","type":"integer","value":"1"}]`,
		`headers`:      `{"Content-Type":"application/json"}`,
		`body_form`:    `[{"field":"id","type":"integer","value":"1"}]`,
	}

	ensureCreateApiOptionalFieldsDefaults(updateData)

	if got := updateData[`query_params`]; got != `[{"field":"page","type":"integer","value":"1"}]` {
		t.Fatalf("query_params should keep existing value, got %v", got)
	}
	if got := updateData[`headers`]; got != `{"Content-Type":"application/json"}` {
		t.Fatalf("headers should keep existing value, got %v", got)
	}
	if got := updateData[`body_form`]; got != `[{"field":"id","type":"integer","value":"1"}]` {
		t.Fatalf("body_form should keep existing value, got %v", got)
	}
}
