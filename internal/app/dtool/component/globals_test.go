package component

import "testing"

func TestComponentExposesRuntimeGlobals(t *testing.T) {
	if ShellOutClient != nil {
		t.Fatal("expected ShellOutClient to default to nil in package test")
	}
	if MemoryRuntime != nil {
		t.Fatal("expected MemoryRuntime to default to nil in package test")
	}
}
