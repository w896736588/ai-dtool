package store

import (
	"dev_tool/internal/app/dtool/common"
	"testing"
)

func newTestRecordSessionStore(t *testing.T) (*RecordSessionStore, bool) {
	t.Helper()
	if common.DbMain == nil || common.DbMain.Client == nil {
		return nil, false
	}
	return &RecordSessionStore{}, true
}

func TestRecordSessionStore_UpdateSmartLink(t *testing.T) {
	s, ok := newTestRecordSessionStore(t)
	if !ok {
		t.Skip("common.DbMain 未注入，跳过")
	}
	id, err := s.Create("demo", "sess-UpdateSmartLink", "https://e", "/api", 0, 0, "")
	if err != nil {
		t.Fatalf("Create: %v", err)
	}
	if err := s.UpdateSmartLink(id, 42, "alice", "wsTok", "/api/e2e/recorder/proxy.html", 7); err != nil {
		t.Fatalf("UpdateSmartLink: %v", err)
	}
	row, err := s.GetByID(id)
	if err != nil || row == nil {
		t.Fatalf("GetByID: %v", err)
	}
	if v, ok := row["smart_link_id"].(int64); !ok || v != 42 {
		t.Fatalf("smart_link_id 期望 42, 实际 %v (%T)", row["smart_link_id"], row["smart_link_id"])
	}
	if name, _ := row["user_name"].(string); name != "alice" {
		t.Fatalf("user_name 不一致: %v", row["user_name"])
	}
}

func TestRecordSessionStore_FindByToken_Empty(t *testing.T) {
	s, ok := newTestRecordSessionStore(t)
	if !ok {
		t.Skip("common.DbMain 未注入")
	}
	row, err := s.FindByToken("__no_such_token__")
	if err != nil {
		t.Fatalf("FindByToken: %v", err)
	}
	if row != nil {
		t.Fatalf("期望 nil，实际 %v", row)
	}
}

func TestRecordSessionStore_FindByToken_BlankReturnsNil(t *testing.T) {
	s, ok := newTestRecordSessionStore(t)
	if !ok {
		t.Skip("common.DbMain 未注入")
	}
	row, err := s.FindByToken("")
	if err != nil {
		t.Fatalf("FindByToken empty: %v", err)
	}
	if row != nil {
		t.Fatalf("期望空 token 返回 nil，实际 %v", row)
	}
}