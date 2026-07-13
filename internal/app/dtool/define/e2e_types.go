package define

import "encoding/json"

// E2E 用例执行状态常量。
const (
	E2ERunStatusPending = "pending"
	E2ERunStatusRunning = "running"
	E2ERunStatusPassed  = "passed"
	E2ERunStatusFailed  = "failed"
	E2ERunStatusStopped = "stopped"
	E2ERunStatusError   = "error"
)

// E2E 用例 / 分组 / 执行详情 API 请求/响应定义。

// E2EGroupItem 分组列表项。
type E2EGroupItem struct {
	ID                 int    `json:"id"`
	Name               string `json:"name"`
	WorkflowTaskID     int    `json:"workflow_task_id"`
	NotificationEnabled int   `json:"notification_enabled"`
	WebhookConfigID    int    `json:"webhook_config_id"`
	CaseCount          int    `json:"case_count"`
	CreatedAt          int64  `json:"created_at"`
	UpdatedAt          int64  `json:"updated_at"`
}

// E2EGroupCreateRequest 分组新建请求。
type E2EGroupCreateRequest struct {
	Name               string `json:"name"`
	WorkflowTaskID     int    `json:"workflow_task_id,omitempty"`
	NotificationEnabled int   `json:"notification_enabled,omitempty"`
	WebhookConfigID    int    `json:"webhook_config_id,omitempty"`
}

// E2EGroupUpdateRequest 分组更新请求。
type E2EGroupUpdateRequest struct {
	ID                 int    `json:"id"`
	Name               string `json:"name,omitempty"`
	NotificationEnabled int   `json:"notification_enabled,omitempty"`
	WebhookConfigID    int    `json:"webhook_config_id,omitempty"`
}

// E2EGroupDeleteRequest 分组删除请求。
type E2EGroupDeleteRequest struct {
	ID int `json:"id"`
}

// E2EGroupListRequest 分组列表请求。
type E2EGroupListRequest struct {
	Page     int `json:"page,omitempty"`
	PageSize int `json:"page_size,omitempty"`
	Keyword  string `json:"keyword,omitempty"`
}

// E2EGroupListResponse 分组列表响应。
type E2EGroupListResponse struct {
	List       []E2EGroupItem `json:"list"`
	Pagination Pagination     `json:"pagination"`
}

// Pagination 通用分页。
type Pagination struct {
	Page      int   `json:"page"`
	PageSize  int   `json:"page_size"`
	Total     int64 `json:"total"`
	TotalPage int   `json:"total_page"`
}

// E2ECaseItem 用例列表项（不含详细配置）。
type E2ECaseItem struct {
	ID                  int             `json:"id"`
	GroupID             int             `json:"group_id"`
	Name                string          `json:"name"`
	EnvURL              string          `json:"env_url"`
	EnvBaseURL          string          `json:"env_base_url"`
	StepCount           int             `json:"step_count"`
	AssertionCount      int             `json:"assertion_count"`
	Tags                string          `json:"tags"`
	TimeoutSeconds      int             `json:"timeout_seconds"`
	NotificationEnabled int             `json:"notification_enabled"`
	LastRunStatus       string          `json:"last_run_status"`
	LastRunAt           int64           `json:"last_run_at"`
	CreatedAt           int64           `json:"created_at"`
	UpdatedAt           int64           `json:"updated_at"`
	StepSummary         json.RawMessage `json:"step_summary,omitempty"`
}

// E2ECaseListRequest 用例列表请求。
type E2ECaseListRequest struct {
	GroupID  int    `json:"group_id,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	Keyword  string `json:"keyword,omitempty"`
}

// E2ECaseListResponse 用例列表响应。
type E2ECaseListResponse struct {
	List       []E2ECaseItem `json:"list"`
	Pagination Pagination    `json:"pagination"`
}

// E2ECaseDetailRequest 用例详情请求。
type E2ECaseDetailRequest struct {
	ID int `json:"id"`
}

// E2ECaseDetailResponse 用例详情响应。
type E2ECaseDetailResponse struct {
	Case json.RawMessage `json:"case"`
}

// E2ECaseSaveRequest 用例新建/更新请求。
type E2ECaseSaveRequest struct {
	ID                  int             `json:"id,omitempty"`
	GroupID             int             `json:"group_id"`
	Name                string          `json:"name"`
	EnvURL              string          `json:"env_url,omitempty"`
	EnvBaseURL          string          `json:"env_base_url,omitempty"`
	Steps               json.RawMessage `json:"steps"`               // JSON 数组
	Assertions          json.RawMessage `json:"assertions"`          // JSON 数组
	Variables           json.RawMessage `json:"variables"`           // JSON 对象
	Tags                string          `json:"tags,omitempty"`
	TimeoutSeconds      int             `json:"timeout_seconds,omitempty"`
	NotificationEnabled int             `json:"notification_enabled,omitempty"`
}

// E2ECaseDeleteRequest 用例删除请求。
type E2ECaseDeleteRequest struct {
	ID int `json:"id"`
}

// E2ERunExecuteRequest 执行用例请求。
type E2ERunExecuteRequest struct {
	CaseID int    `json:"case_id"`
	Mode   string `json:"mode,omitempty"` // sync/async；默认 async
}

// E2ERunExecuteResponse 执行用例响应。
type E2ERunExecuteResponse struct {
	RunID int64 `json:"run_id"`
}

// E2ERunBatchRequest 批量执行请求（按 group）。
type E2ERunBatchRequest struct {
	GroupID int `json:"group_id"`
}

// E2ERunBatchResponse 批量执行响应。
type E2ERunBatchResponse struct {
	RunIDs []int64 `json:"run_ids"`
}

// E2ERunStopRequest 停止执行请求。
type E2ERunStopRequest struct {
	RunID int64 `json:"run_id"`
}

// E2ERunDetailRequest 执行详情请求。
type E2ERunDetailRequest struct {
	RunID int64 `json:"run_id"`
}

// E2ERunDetailResponse 执行详情响应（包含 steps / assertions / requests）。
type E2ERunDetailResponse struct {
	Run       json.RawMessage `json:"run"`
	Steps     json.RawMessage `json:"steps"`
	Assertions json.RawMessage `json:"assertions"`
	Requests  json.RawMessage `json:"requests"`
}

// E2ERunListRequest 执行列表请求。
type E2ERunListRequest struct {
	CaseID  int    `json:"case_id,omitempty"`
	GroupID int    `json:"group_id,omitempty"`
	Status  string `json:"status,omitempty"`
	Page    int    `json:"page,omitempty"`
	PageSize int   `json:"page_size,omitempty"`
}

// E2ERunListResponse 执行列表响应。
type E2ERunListResponse struct {
	List       []E2ERunItem `json:"list"`
	Pagination Pagination   `json:"pagination"`
}

// E2ERunItem 执行列表项。
type E2ERunItem struct {
	ID             int64  `json:"id"`
	CaseID         int    `json:"case_id"`
	CaseName       string `json:"case_name"`
	GroupID        int    `json:"group_id"`
	GroupName      string `json:"group_name"`
	Status         string `json:"status"`
	TotalSteps     int    `json:"total_steps"`
	PassedSteps    int    `json:"passed_steps"`
	FailedSteps    int    `json:"failed_steps"`
	TotalAsserts   int    `json:"total_asserts"`
	PassedAsserts  int    `json:"passed_asserts"`
	FailedAsserts  int    `json:"failed_asserts"`
	StartedAt      int64  `json:"started_at"`
	FinishedAt     int64  `json:"finished_at"`
	DurationMs     int    `json:"duration_ms"`
	TriggerType    string `json:"trigger_type"`
	ErrorMessage   string `json:"error_message"`
	CreatedAt      int64  `json:"created_at"`
}

// E2ERunRequestsRequest 请求追踪查询请求。
type E2ERunRequestsRequest struct {
	RunID  int64  `json:"run_id"`
	StepID string `json:"step_id,omitempty"`
	Method string `json:"method,omitempty"`
	URL    string `json:"url,omitempty"`
}

// E2ERunRequestItem 请求追踪项。
type E2ERunRequestItem struct {
	ID              string `json:"id"`
	RunID           int64  `json:"run_id"`
	RunStepID       int    `json:"run_step_id"`
	StepID          string `json:"step_id"`
	URL             string `json:"url"`
	Method          string `json:"method"`
	RequestHeaders  string `json:"request_headers"`
	RequestBody     string `json:"request_body,omitempty"`
	ResponseStatus  int    `json:"response_status"`
	ResponseHeaders string `json:"response_headers"`
	ResponseBody    string `json:"response_body,omitempty"`
	ResponseTimeMs  int    `json:"response_time_ms"`
	Matched         bool   `json:"matched"`
	MatchedBy       string `json:"matched_by,omitempty"`
	CapturedAt      int64  `json:"captured_at"`
}

// E2ERunRequestDetailRequest 单个请求详情请求。
type E2ERunRequestDetailRequest struct {
	RunID     int64  `json:"run_id"`
	RequestID string `json:"request_id"`
}

// E2ERecordStartRequest 开始录制请求。
type E2ERecordStartRequest struct {
	CaseID      int    `json:"case_id,omitempty"`
	EnvURL      string `json:"env_url"`
	EnvBaseURL  string `json:"env_base_url,omitempty"`
	Name        string `json:"name,omitempty"`
}

// E2ERecordStartResponse 开始录制响应（返回 session id）。
type E2ERecordStartResponse struct {
	SessionID string `json:"session_id"`
}

// E2ERecordStopRequest 停止录制请求。
type E2ERecordStopRequest struct {
	SessionID string `json:"session_id"`
}

// E2ERecordSessionResponse 录制会话详情。
type E2ERecordSessionResponse struct {
	SessionID   string          `json:"session_id"`
	CaseID      int             `json:"case_id"`
	EnvURL      string          `json:"env_url"`
	EnvBaseURL  string          `json:"env_base_url"`
	Name        string          `json:"name"`
	Steps       json.RawMessage `json:"steps"`
	UpdatedAt   int64           `json:"updated_at"`
}

// E2ERecordSaveRequest 录制转用例保存请求。
type E2ERecordSaveRequest struct {
	SessionID string          `json:"session_id"`
	GroupID   int             `json:"group_id"`
	Name      string          `json:"name"`
	Steps     json.RawMessage `json:"steps,omitempty"` // 若传入，使用传入的步骤覆盖
}

// E2EStepTypeListResponse 步骤类型清单响应（用于前端动态渲染）。
type E2EStepTypeListResponse struct {
	Items []E2EStepTypeMeta `json:"items"`
}

// E2EStepTypeMeta 单个步骤类型元信息。
type E2EStepTypeMeta struct {
	Type        string   `json:"type"`
	BaseType    string   `json:"base_type"`
	Version     string   `json:"version"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	ConfigKeys  []string `json:"config_keys"`
	Group       string   `json:"group"` // action/wait/extract/script
	Deprecated  bool     `json:"deprecated"`
}

// E2EAssertionTypeListResponse 断言类型清单。
type E2EAssertionTypeListResponse struct {
	Items []E2EAssertionTypeMeta `json:"items"`
}

// E2EAssertionTypeMeta 单个断言类型元信息。
type E2EAssertionTypeMeta struct {
	Type        string   `json:"type"`
	BaseType    string   `json:"base_type"`
	Version     string   `json:"version"`
	Label       string   `json:"label"`
	Description string   `json:"description"`
	ConfigKeys  []string `json:"config_keys"`
	Group       string   `json:"group"` // page/api/variable
	Deprecated  bool     `json:"deprecated"`
}
