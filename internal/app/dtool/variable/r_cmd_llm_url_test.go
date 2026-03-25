package variable

import "testing"

func TestResolveOpenAIChatCompletionsURL_BaseDomain(t *testing.T) {
	got := resolveOpenAIChatCompletionsURL("https://api.openai.com")
	want := "https://api.openai.com/v1/chat/completions"
	if got != want {
		t.Fatalf("resolveOpenAIChatCompletionsURL() = %q, want %q", got, want)
	}
}
