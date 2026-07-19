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
	ID                  int   `json:"id"`
	Name                string `json:"name"`
	WorkflowTaskID      int   `json:"workflow_task_id"`
	NotificationEnabled bool  `json:"notification_enabled"`
	WebhookConfigID     int   `json:"webhook_config_id"`
	CaseCount           int   `json:"case_count"`
	CreatedAt           int64 `json:"created_at"`
	UpdatedAt           int64 `json:"updated_at"`
}

// E2EGroupCreateRequest 分组新建请求。
type E2EGroupCreateRequest struct {
	Name                string `json:"name"`
	WorkflowTaskID      int    `json:"workflow_task_id,omitempty"`
	NotificationEnabled bool   `json:"notification_enabled,omitempty"`
	WebhookConfigID     int    `json:"webhook_config_id,omitempty"`
}

// E2EGroupUpdateRequest 分组更新请求。
type E2EGroupUpdateRequest struct {
	ID                  int   `json:"id"`
	Name                string `json:"name,omitempty"`
	NotificationEnabled bool   `json:"notification_enabled,omitempty"`
	WebhookConfigID     int   `json:"webhook_config_id,omitempty"`
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
	NotificationEnabled bool            `json:"notification_enabled"`
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
	NotificationEnabled bool            `json:"notification_enabled,omitempty"`
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
	Run        json.RawMessage `json:"run"`
	Steps      json.RawMessage `json:"steps"`
	Assertions json.RawMessage `json:"assertions"`
	Requests   json.RawMessage `json:"requests"`
}

// E2ERunListRequest 执行列表请求。
type E2ERunListRequest struct {
	CaseID   int    `json:"case_id,omitempty"`
	GroupID  int    `json:"group_id,omitempty"`
	Status   string `json:"status,omitempty"`
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
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
	ID               string `json:"id"`
	RunID            int64  `json:"run_id"`
	RunStepID        int    `json:"run_step_id"`
	StepID           string `json:"step_id"`
	URL              string `json:"url"`
	Method           string `json:"method"`
	RequestHeaders   string `json:"request_headers"`
	RequestBody      string `json:"request_body,omitempty"`
	ResponseStatus   int    `json:"response_status"`
	ResponseHeaders  string `json:"response_headers"`
	ResponseBody     string `json:"response_body,omitempty"`
	ResponseTimeMs   int    `json:"response_time_ms"`
	Matched          bool   `json:"matched"`
	MatchedBy        string `json:"matched_by,omitempty"`
	CapturedAt       int64  `json:"captured_at"`
}

// E2ERunRequestDetailRequest 单个请求详情请求。
type E2ERunRequestDetailRequest struct {
	RunID     int64  `json:"run_id"`
	RequestID string `json:"request_id"`
}

// E2ERecordStartRequest 开始录制请求。
type E2ERecordStartRequest struct {
	CaseID     int    `json:"case_id,omitempty"`
	EnvURL     string `json:"env_url"`
	EnvBaseURL string `json:"env_base_url,omitempty"`
	Name       string `json:"name,omitempty"`
}

// E2ERecordStartResponse 开始录制响应（返回 session id）。
type E2ERecordStartResponse struct {
	SessionID   string `json:"session_id"`
	RecorderURL string `json:"recorder_url,omitempty"` // 录制浏览器页面的本地标识（用于前端展示）
	EnvURL      string `json:"env_url"`
}

// E2ERecordStopRequest 停止录制请求。
type E2ERecordStopRequest struct {
	SessionID string `json:"session_id"`
}

// E2ERecordSessionResponse 录制会话详情。
type E2ERecordSessionResponse struct {
	SessionID  string          `json:"session_id"`
	CaseID     int             `json:"case_id"`
	EnvURL     string          `json:"env_url"`
	EnvBaseURL string          `json:"env_base_url"`
	Name       string          `json:"name"`
	Steps      json.RawMessage `json:"steps"`
	UpdatedAt  int64           `json:"updated_at"`
}

// E2ERecordSaveRequest 录制转用例保存请求。
type E2ERecordSaveRequest struct {
	SessionID string          `json:"session_id"`
	GroupID   int             `json:"group_id"`
	Name      string          `json:"name"`
	Steps     json.RawMessage `json:"steps,omitempty"` // 若传入，使用传入的步骤覆盖
}

// ===== v5.0 录制扩展接口 =====

// E2ERecordSessionCreateRequest 创建录制会话（细粒度版本，等价于 Start）。
// v6：携带 smart_link 绑定字段，recorder.js 通过 ws_token 鉴权上报步骤。
type E2ERecordSessionCreateRequest struct {
	SessionName string `json:"session_name"`           // 会话名（前端展示）
	SessionID   string `json:"session_id,omitempty"`    // 客户端可传 UUID，否则后端生成
	CaseID      int    `json:"case_id,omitempty"`       // 关联已有用例（0 表示新建录制）
	GroupID     int    `json:"group_id,omitempty"`      // 关联分组，便于按组汇总
	EnvURL      string `json:"env_url"`                 // 录制入口 URL
	EnvBaseURL  string `json:"env_base_url,omitempty"`  // 环境 base URL
	BrowserID   string `json:"browser_id,omitempty"`    // Playwright 实例标识
	SmartLinkID int    `json:"smart_link_id,omitempty"` // 关联 smart_link，便于复用登录链路
	LinkID      int    `json:"link_id,omitempty"`       // smart_link 对应的子链接 ID
	UserName    string `json:"user_name,omitempty"`     // 录制人
	WSToken     string `json:"ws_token,omitempty"`      // 一次性 token，供 recorder.js 调用
	RecorderURL string `json:"recorder_url,omitempty"`  // recorder proxy iframe 路径
}

// E2ERecordSessionCreateResponse 创建录制会话响应。
type E2ERecordSessionCreateResponse struct {
	ID        int64  `json:"id"`         // 自增主键
	SessionID string `json:"session_id"` // 业务 ID
	Status    string `json:"status"`     // recording
	GroupID   int    `json:"group_id"`
	CaseID    int    `json:"case_id"`
}

// E2ERecordSessionGetRequest 查询录制会话请求。
type E2ERecordSessionGetRequest struct {
	ID int64 `json:"id"`
}

// E2ERecordSessionListRequest 录制会话列表请求。
type E2ERecordSessionListRequest struct {
	Page     int    `json:"page,omitempty"`
	PageSize int    `json:"page_size,omitempty"`
	CaseID   int    `json:"case_id,omitempty"`
	GroupID  int    `json:"group_id,omitempty"`
	Status   string `json:"status,omitempty"`
}

// E2ERecordSessionListResponse 录制会话列表响应。
type E2ERecordSessionListResponse struct {
	Items    []E2ERecordListItem `json:"items"`
	Total    int                 `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}

// E2ERecordSessionDeleteRequest 删除录制会话请求。
type E2ERecordSessionDeleteRequest struct {
	ID int64 `json:"id"`
}

// E2ERecordStepAddRequest 追加录制步骤请求。
type E2ERecordStepAddRequest struct {
	SessionID int64           `json:"session_id"`
	Step      E2ERecordedStep `json:"step"`         // 完整步骤 JSON（与 E2EStep 兼容）
	Position  int             `json:"position,omitempty"` // 插入位置，默认追加到末尾
}

// E2ERecordStepAddResponse 追加步骤响应。
type E2ERecordStepAddResponse struct {
	SessionID string `json:"session_id"`
	StepID    string `json:"step_id"`
	StepIndex int    `json:"step_index"`
}

// E2ERecordStepUpdateRequest 更新步骤请求（编辑 wait_after / description / assertions）。
type E2ERecordStepUpdateRequest struct {
	SessionID int64           `json:"session_id"`
	StepID    string          `json:"step_id"`
	Step      E2ERecordedStep `json:"step"`             // 整步替换
}

// E2ERecordStepDeleteRequest 删除步骤请求。
type E2ERecordStepDeleteRequest struct {
	SessionID int64  `json:"session_id"`
	StepID    string `json:"step_id"`
}

// E2ERecordCommitRequest 录制会话提交到用例请求。
type E2ERecordCommitRequest struct {
	SessionID      int64  `json:"session_id"`
	CaseID         int    `json:"case_id,omitempty"`    // 为 0 表示新建
	GroupID        int    `json:"group_id"`             // 新建用例所属分组
	Name           string `json:"name,omitempty"`
	Tags           string `json:"tags,omitempty"`
	TimeoutSeconds int    `json:"timeout_seconds,omitempty"`
}

// E2ERecordCommitResponse 提交响应。
type E2ERecordCommitResponse struct {
	CaseID  int64 `json:"case_id"`
	Steps   int   `json:"steps"`
	GroupID int   `json:"group_id"`
}

// E2ERecordListItem 录制会话列表项。
type E2ERecordListItem struct {
	ID         int64  `json:"id"`
	SessionID  string `json:"session_id"`
	CaseID     int    `json:"case_id"`
	GroupID    int    `json:"group_id"`
	Name       string `json:"name"`
	EnvURL     string `json:"env_url"`
	EnvBaseURL string `json:"env_base_url,omitempty"`
	BrowserID  string `json:"browser_id,omitempty"`
	Status     string `json:"status"`
	StepCount  int    `json:"step_count"`
	UpdatedAt  int64  `json:"updated_at"`
	CreatedAt  int64  `json:"created_at"`
}

// E2ERecordSessionDetail 录制会话详情（含步骤列表）。
type E2ERecordSessionDetail struct {
	ID          int64           `json:"id"`
	SessionID   string          `json:"session_id"`
	CaseID      int             `json:"case_id"`
	GroupID     int             `json:"group_id"`
	Name        string          `json:"name"`
	EnvURL      string          `json:"env_url"`
	EnvBaseURL  string          `json:"env_base_url,omitempty"`
	BrowserID   string          `json:"browser_id,omitempty"`
	SmartLinkID int             `json:"smart_link_id,omitempty"`
	LinkID      int             `json:"link_id,omitempty"`
	UserName    string          `json:"user_name,omitempty"`
	RecorderURL string          `json:"recorder_url,omitempty"`
	Status      string          `json:"status"`
	Steps       []E2ERecordedStep `json:"steps"`
	CreatedAt   int64           `json:"created_at"`
	UpdatedAt   int64           `json:"updated_at"`
}

// E2ERecordedStep 录制时的步骤结构（与 E2EStep 兼容，额外带 wait_after_ms 等录制元信息）。
// 这里用 alias，避免重复定义导致的不一致。
type E2ERecordedStep = RecordedStep

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

// ===== v6.0 基于 smart_link + ws_token 的录制接口 =====

// E2ERecordOpenRequest 开启一次 smart_link 录制会话。
// smart_link_id 决定登录链路与浏览器上下文；ws_token 内部生成。
// Password 必须由前端从 smart_link 的 userList 中读出传入，与 SmartLinkRunPlaywright 走
// 同一份 GetRunParams 逻辑——否则 process list 中的"输入密码"会用空串覆盖密码框。
type E2ERecordOpenRequest struct {
	SmartLinkID int    `json:"smart_link_id"`
	LinkID      int    `json:"link_id,omitempty"`
	UserName    string `json:"user_name"`
	Password    string `json:"password,omitempty"`
	SessionName string `json:"session_name,omitempty"`
	GroupID     int    `json:"group_id,omitempty"`
	CaseID      int    `json:"case_id,omitempty"`
}

// E2ERecordOpenResponse 开启录制会话响应。
type E2ERecordOpenResponse struct {
	OK          bool   `json:"ok"`
	BrowserID   string `json:"browser_id,omitempty"`
	SessionID   int64  `json:"session_id,omitempty"`
	SessionUUID string `json:"session_uuid,omitempty"`
	WSToken     string `json:"ws_token,omitempty"`
	RecorderURL string `json:"recorder_url,omitempty"`
	EnvURL      string `json:"env_url,omitempty"`
	Error       string `json:"error,omitempty"`
}

// E2ERecordStepByTokenRequest recorder.js 通过 ws_token 上报单步。
type E2ERecordStepByTokenRequest struct {
	Step RecordedStep `json:"step"`
}

// E2ERecordCommitByTokenRequest recorder.js 通过 ws_token 提交录制到用例。
type E2ERecordCommitByTokenRequest struct {
	GroupID int    `json:"group_id"`
	Name    string `json:"name,omitempty"`
	Tags    string `json:"tags,omitempty"`
}

// E2ERecordResumeRequest 续录请求（按 row_id 复用会话）。
type E2ERecordResumeRequest struct {
	SessionID int64 `json:"session_id"`
}
