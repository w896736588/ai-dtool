package p_common

import "testing"

func TestReplaceSupportsRegexMetaKeys(t *testing.T) {
	got := Replace(`demo*name?`, map[string]string{
		`*`: `_`,
		`?`: `_`,
	})
	if got != `demo_name_` {
		t.Fatalf("unexpected replace result: %q", got)
	}
}

func TestReplaceModuloWithRegexMetaKey(t *testing.T) {
	got := Replace(`*%3`, map[string]string{
		`*`: `10`,
	})
	if got != `1` {
		t.Fatalf("unexpected modulo replace result: %q", got)
	}
}
