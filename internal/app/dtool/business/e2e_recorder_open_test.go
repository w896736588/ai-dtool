package business

import (
	"dev_tool/internal/app/dtool/define"
	"testing"
)

func TestE2EEngine_OpenRecorder_RequiresSmartLinkID(t *testing.T) {
	e := NewE2EEngine()
	if _, _, err := e.OpenRecorder(0, "alice"); err == nil {
		t.Fatal("expected error when smart_link_id=0")
	}
}

func TestE2ERecordOpen_RequiresSmartLinkID(t *testing.T) {
	if _, err := E2ERecordOpen(&define.E2ERecordOpenRequest{SmartLinkID: 0, UserName: "alice"}); err == nil {
		t.Fatal("期望错误")
	}
}

func TestE2ERecordResume_InvalidID(t *testing.T) {
	if _, err := E2ERecordResume(0); err == nil {
		t.Fatal("期望错误")
	}
}