package controller

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"dev_tool/internal/app/dtool/common"

	"gitee.com/Sxiaobai/gs/v2/gsdb"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func TestSetAiRequestLogListIncludesDetailFields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logDB := newAIRequestLogTestDB(t)
	previousDbLog := common.DbLog
	common.DbLog = logDB
	defer func() {
		common.DbLog = previousDbLog
	}()

	_, err := common.DbLog.Client.QuickCreate(`tbl_ai_request_log`, map[string]any{
		`provider_id`:           1,
		`provider_name`:         `provider-a`,
		`model_id`:              2,
		`model_name`:            `model-a`,
		`model`:                 `gpt-test`,
		`model_type`:            `llm`,
		`request_format`:        `openai`,
		`base_url`:              `https://api.example.com`,
		`request_url`:           `https://api.example.com/v1/chat/completions`,
		`request_method`:        `POST`,
		`request_params`:        `{"messages":[{"role":"user","content":"hello"}]}`,
		`request_headers`:       `{"Authorization":"Bearer ******"}`,
		`response_status_code`:  200,
		`response_body`:         `{"choices":[{"message":{"content":"hi"}}]}`,
		`input_tokens`:          10,
		`output_tokens`:         5,
		`cost_time_ms`:          1234,
		`success`:               1,
		`error_message`:         ``,
		`create_time`:           1710000000,
	}).Exec()
	if err != nil {
		t.Fatalf("QuickCreate(tbl_ai_request_log) error = %v", err)
	}

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, `/api/Set/AiRequestLogList`, bytes.NewBufferString(`{"limit":10}`))
	ctx.Request.Header.Set(`Content-Type`, `application/json`)

	SetAiRequestLogList(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("response code = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response struct {
		ErrCode int              `json:"ErrCode"`
		Data    []map[string]any `json:"Data"`
	}
	if err = json.Unmarshal(recorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("Unmarshal() error = %v", err)
	}
	if response.ErrCode != 0 {
		t.Fatalf("ErrCode = %d, want 0, body=%s", response.ErrCode, recorder.Body.String())
	}
	if len(response.Data) != 1 {
		t.Fatalf("Data len = %d, want 1", len(response.Data))
	}
	if cast.ToString(response.Data[0][`request_params`]) == `` {
		t.Fatalf("request_params = empty, want populated")
	}
	if cast.ToString(response.Data[0][`response_body`]) == `` {
		t.Fatalf("response_body = empty, want populated")
	}
	if cast.ToString(response.Data[0][`request_headers`]) == `` {
		t.Fatalf("request_headers = empty, want populated")
	}
}

func newAIRequestLogTestDB(t *testing.T) *common.CSqlite {
	t.Helper()

	sqliteClient, err := gsdb.NewSqlite(`:memory:`, false)
	if err != nil {
		t.Fatalf("NewSqlite() error = %v", err)
	}
	db := &common.CSqlite{Client: sqliteClient}

	sqlPath := filepath.Join(`..`, `database_log`, `2026`, `04`, `20260409.093000_ai_request_log.sql`)
	sqlBytes, err := os.ReadFile(sqlPath)
	if err != nil {
		t.Fatalf("ReadFile() error = %v", err)
	}
	if _, err = db.Client.ExecBySql(string(sqlBytes)).Exec(); err != nil {
		t.Fatalf("ExecBySql() error = %v", err)
	}
	return db
}
