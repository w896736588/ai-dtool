package controller

import "testing"

// TestFilterEmptyArrayMapRejectsLegacyIntegerType 中文：验证旧的 int 类型会被直接拒绝，避免继续写入历史格式。 English: Verify the legacy int type is rejected instead of being persisted.
func TestFilterEmptyArrayMapRejectsLegacyIntegerType(t *testing.T) {
	input := `[{"field":"page","type":"int","value":"1"},{"field":"name","type":"string","value":"frog"}]`

	_, err := filterEmptyArrayMap(input, `field`, `请求参数格式错误`, 500)
	if err == nil {
		t.Fatalf("filterEmptyArrayMap expected error for legacy int type")
	}
}
