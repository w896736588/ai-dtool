package common

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"github.com/spf13/cast"
)

func TestJoinAIRequestURL(t *testing.T) {
	got := joinAIRequestURL("https://api.openai.com/", "/v1/chat/completions")
	want := "https://api.openai.com/v1/chat/completions"
	if got != want {
		t.Fatalf("joinAIRequestURL() = %q, want %q", got, want)
	}
}

func TestAIChatStreamByModel(t *testing.T) {
	var capturedBody map[string]any
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Fatalf("request method = %s, want POST", r.Method)
		}
		if err := json.NewDecoder(r.Body).Decode(&capturedBody); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		w.Header().Set(`Content-Type`, `text/event-stream`)
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"hello \"}}]}\n\n"))
		_, _ = w.Write([]byte("data: {\"choices\":[{\"delta\":{\"content\":\"world\"}}]}\n\n"))
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	defer server.Close()

	db := newAIChatTestDB(t)
	providerID, err := db.Client.QuickCreate(`tbl_ai_provider`, map[string]any{
		`name`:          `test-provider`,
		`provider_type`: `openai`,
		`base_url`:      server.URL,
		`api_key`:       `test-key`,
		`status`:        1,
	}).Exec()
	if err != nil {
		t.Fatalf("QuickCreate(provider) error = %v", err)
	}
	modelID, err := db.Client.QuickCreate(`tbl_ai_model`, map[string]any{
		`provider_id`: providerID,
		`name`:        `test-model`,
		`model`:       `gpt-test`,
		`model_type`:  `llm`,
		`uri`:         `/v1/chat/completions`,
		`status`:      1,
	}).Exec()
	if err != nil {
		t.Fatalf("QuickCreate(model) error = %v", err)
	}

	chunkList := make([]string, 0)
	result, modelInfo, err := db.AIChatStreamByModel(cast.ToInt(modelID), `system prompt`, `user prompt`, func(chunk string) {
		chunkList = append(chunkList, chunk)
	})
	if err != nil {
		t.Fatalf("AIChatStreamByModel() error = %v", err)
	}
	if result != `hello world` {
		t.Fatalf("AIChatStreamByModel() result = %q, want %q", result, `hello world`)
	}
	if strings.Join(chunkList, ``) != `hello world` {
		t.Fatalf("stream chunks = %q, want %q", strings.Join(chunkList, ``), `hello world`)
	}
	if cast.ToString(modelInfo[`model`]) != `gpt-test` {
		t.Fatalf("modelInfo.model = %q, want %q", cast.ToString(modelInfo[`model`]), `gpt-test`)
	}
	if cast.ToBool(capturedBody[`stream`]) != true {
		t.Fatalf("request body stream = %v, want true", capturedBody[`stream`])
	}
}

func newAIChatTestDB(t *testing.T) *CSqlite {
	t.Helper()

	sqliteClient, err := gsdb.NewSqlite(`:memory:`, false)
	if err != nil {
		t.Fatalf("NewSqlite() error = %v", err)
	}
	db := &CSqlite{Client: sqliteClient}

	sqlPathList := []string{
		filepath.Join(`..`, `database`, `2026`, `03`, `20260303.10.30_ai_provider_model.sql`),
		filepath.Join(`..`, `database`, `2026`, `03`, `20260321.09.30_ai_model_uri_type.sql`),
	}
	for _, sqlPath := range sqlPathList {
		sqlBytes, readErr := os.ReadFile(sqlPath)
		if readErr != nil {
			t.Fatalf("ReadFile(%s) error = %v", sqlPath, readErr)
		}
		if _, execErr := db.Client.ExecBySql(string(sqlBytes)).Exec(); execErr != nil {
			t.Fatalf("ExecBySql(%s) error = %v", sqlPath, execErr)
		}
	}
	return db
}
