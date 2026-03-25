package controller

import "testing"

func TestNormalizeAiProviderBaseURL(t *testing.T) {
	got := normalizeAiProviderBaseURL(" https://api.openai.com/v1/chat/completions ")
	want := "https://api.openai.com"
	if got != want {
		t.Fatalf("normalizeAiProviderBaseURL() = %q, want %q", got, want)
	}
}

func TestNormalizeAiModelURI(t *testing.T) {
	got := normalizeAiModelURI("v1/embeddings")
	want := "/v1/embeddings"
	if got != want {
		t.Fatalf("normalizeAiModelURI() = %q, want %q", got, want)
	}
}

func TestNormalizeAiModelType_DefaultLLM(t *testing.T) {
	got := normalizeAiModelType("")
	want := "llm"
	if got != want {
		t.Fatalf("normalizeAiModelType() = %q, want %q", got, want)
	}
}
