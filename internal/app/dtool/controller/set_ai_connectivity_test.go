package controller

import "testing"

func TestBuildAiModelConnectivityRequest_LLM(t *testing.T) {
	method, body, err := buildAiModelConnectivityRequest(map[string]any{
		`model_type`: `llm`,
		`model`:      `gpt-4o-mini`,
	})
	if err != nil {
		t.Fatalf("buildAiModelConnectivityRequest() unexpected err = %v", err)
	}
	if method != "POST" {
		t.Fatalf("method = %q, want POST", method)
	}
	messages, ok := body[`messages`].([]map[string]string)
	if !ok || len(messages) != 1 {
		t.Fatalf("messages = %#v, want single message", body[`messages`])
	}
	if body[`model`] != `gpt-4o-mini` {
		t.Fatalf("model = %v, want gpt-4o-mini", body[`model`])
	}
}

func TestBuildAiModelConnectivityRequest_Embedding(t *testing.T) {
	method, body, err := buildAiModelConnectivityRequest(map[string]any{
		`model_type`: `embedding`,
		`model`:      `text-embedding-3-small`,
	})
	if err != nil {
		t.Fatalf("buildAiModelConnectivityRequest() unexpected err = %v", err)
	}
	if method != "POST" {
		t.Fatalf("method = %q, want POST", method)
	}
	if body[`model`] != `text-embedding-3-small` {
		t.Fatalf("model = %v, want text-embedding-3-small", body[`model`])
	}
	if body[`input`] == nil {
		t.Fatalf("input should not be nil")
	}
}

func TestBuildAiModelConnectivityRequest_UnsupportedType(t *testing.T) {
	_, _, err := buildAiModelConnectivityRequest(map[string]any{
		`model_type`: `image`,
		`model`:      `gpt-image-1`,
	})
	if err == nil {
		t.Fatal("expected unsupported model type error")
	}
}
