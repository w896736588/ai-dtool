package business

import "testing"

func TestE2EEngine_OpenRecorder_RequiresSmartLinkID(t *testing.T) {
	e := NewE2EEngine()
	if _, _, err := e.OpenRecorder(0, "alice"); err == nil {
		t.Fatal("expected error when smart_link_id=0")
	}
}