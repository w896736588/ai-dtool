package memory

import (
	"testing"
	"time"
)

func TestFragmentToMapOmitsIndexStatusFields(t *testing.T) {
	fragment := Fragment{
		ID:        "12",
		FilePath:  "memory/fragments/demo.md",
		Title:     "Demo",
		Content:   "# Demo\n\ncontent",
		CreatedAt: time.Date(2026, 4, 10, 8, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC),
	}

	result := fragmentToMap(fragment)

	if _, ok := result["index_status"]; ok {
		t.Fatalf("fragmentToMap should not expose index_status")
	}
	if _, ok := result["index_status_desc"]; ok {
		t.Fatalf("fragmentToMap should not expose index_status_desc")
	}
}

func TestFragmentToMapKeepsLargeNumericIDAsString(t *testing.T) {
	fragment := Fragment{
		ID:        "1775811447303500600",
		FilePath:  "memory/fragments/2026/2026-04/1775811447303500600.md",
		Title:     "Large ID",
		Content:   "# Large ID\n\ncontent",
		CreatedAt: time.Date(2026, 4, 10, 8, 0, 0, 0, time.UTC),
		UpdatedAt: time.Date(2026, 4, 10, 9, 0, 0, 0, time.UTC),
	}

	result := fragmentToMap(fragment)

	// 保持字符串返回，避免前端把超出 JS 安全整数范围的 ID 解析成不精确 number。
	// Keep large numeric IDs as strings to avoid JS safe-integer precision loss in the frontend.
	if got := result["id"]; got != fragment.ID {
		t.Fatalf("fragmentToMap large id = %#v, want %q", got, fragment.ID)
	}
	if got := result["file_id"]; got != fragment.ID {
		t.Fatalf("fragmentToMap file_id = %#v, want %q", got, fragment.ID)
	}
}
