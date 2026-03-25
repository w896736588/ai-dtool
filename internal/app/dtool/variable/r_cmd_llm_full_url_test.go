package variable

import "testing"

func TestResolveOpenAIChatCompletionsURL_KeepExistingPath(t *testing.T) {
	got := resolveOpenAIChatCompletionsURL("https://proxy.example.com/openai/v1/chat/completions")
	want := "https://proxy.example.com/openai/v1/chat/completions"
	if got != want {
		t.Fatalf("resolveOpenAIChatCompletionsURL() = %q, want %q", got, want)
	}
}
