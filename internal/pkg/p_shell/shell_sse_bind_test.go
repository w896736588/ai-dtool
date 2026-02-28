package p_shell

import (
	"testing"
)

type fakeReceiveBinder struct {
	handler func(string) string
}

func (f *fakeReceiveBinder) SetFuncReceiveMsg(handler func(string) string) {
	f.handler = handler
}

func TestBindReceiveHandler_RebindsOnEveryCall(t *testing.T) {
	binder := &fakeReceiveBinder{}

	bindReceiveHandler(binder, nil, nil)
	if binder.handler == nil {
		t.Fatalf("expected handler to be set")
	}
	_ = binder.handler("first")

	callCount := 0
	bindReceiveHandler(binder, nil, func(msg string) []string {
		callCount++
		return []string{msg}
	})
	if binder.handler == nil {
		t.Fatalf("expected handler to be re-set")
	}
	_ = binder.handler("second")
	if callCount != 1 {
		t.Fatalf("expected rebound handler to use new formatter, callCount=%d", callCount)
	}
}
