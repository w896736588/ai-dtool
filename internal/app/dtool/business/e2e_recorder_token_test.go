package business

import (
	"dev_tool/internal/app/dtool/define"
	"testing"
)

func TestE2ERecordStepAddByToken_EmptyTokenRejected(t *testing.T) {
	if _, err := E2ERecordStepAddByToken("", &define.RecordedStep{}); err == nil {
		t.Fatal("期望空 token 报错")
	}
}

func TestE2ERecordCommitByToken_EmptyTokenRejected(t *testing.T) {
	if _, err := E2ERecordCommitByToken("", nil); err == nil {
		t.Fatal("期望空 token 报错")
	}
}