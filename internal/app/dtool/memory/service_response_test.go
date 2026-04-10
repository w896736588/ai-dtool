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
