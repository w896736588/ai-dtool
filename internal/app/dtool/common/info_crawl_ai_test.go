package common

import "testing"

func TestJoinAIRequestURL(t *testing.T) {
	got := joinAIRequestURL("https://api.openai.com/", "/v1/chat/completions")
	want := "https://api.openai.com/v1/chat/completions"
	if got != want {
		t.Fatalf("joinAIRequestURL() = %q, want %q", got, want)
	}
}
