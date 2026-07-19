// Package store 提供 E2E 用例相关数据访问层。
// 复用项目 common.DbMain（基于 gsdb.GsSqlite）。
package store

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component/e2e/interceptor"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/cast"
)

// GroupStore 分组数据访问。
type GroupStore struct{}

func NewGroupStore() *GroupStore { return &GroupStore{} }

// List 列出分组，支持关键字和分页。
func (s *GroupStore) List(keyword string, page, pageSize int) ([]map[string]any, int64, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	where := "1=1"
	args := []any{}
	if keyword != "" {
		where += " AND g.name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	var total int64
	row, err := common.DbMain.Client.QueryBySql(
		"SELECT COUNT(*) AS cnt FROM tbl_e2e_group g WHERE "+where, args...,
	).One()
	if err != nil {
		return nil, 0, err
	}
	if v, ok := row["cnt"]; ok {
		switch x := v.(type) {
		case int:
			total = int64(x)
		case int64:
			total = x
		case float64:
			total = int64(x)
		}
	}
	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := common.DbMain.Client.QueryBySql(`
		SELECT g.*, (SELECT COUNT(*) FROM tbl_e2e_case c WHERE c.group_id = g.id) AS case_count
		FROM tbl_e2e_group g
		WHERE `+where+`
		ORDER BY g.id DESC
		LIMIT ? OFFSET ?`, args...,
	).All()
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// Get 获取单个分组。
func (s *GroupStore) Get(id int) (map[string]any, error) {
	if id <= 0 {
		return nil, nil
	}
	return common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_e2e_group WHERE id = ?`, id,
	).One()
}

// Create 创建分组。
func (s *GroupStore) Create(req *define.E2EGroupCreateRequest) (int64, error) {
	now := time.Now().Unix()
	id, err := common.DbMain.Client.InsertBySql(`
		INSERT INTO tbl_e2e_group (name, workflow_task_id, notification_enabled, webhook_config_id, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?)`,
		req.Name,
		req.WorkflowTaskID,
		boolToInt(req.NotificationEnabled),
		req.WebhookConfigID,
		now, now,
	).Exec()
	return id, err
}

// Update 更新分组。
func (s *GroupStore) Update(req *define.E2EGroupUpdateRequest) error {
	updates := map[string]any{"updated_at": time.Now().Unix()}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	updates["notification_enabled"] = boolToInt(req.NotificationEnabled)
	if req.WebhookConfigID > 0 {
		updates["webhook_config_id"] = req.WebhookConfigID
	}
	if len(updates) == 1 {
		return nil
	}
	_, err := common.DbMain.Client.QuickUpdate(
		"tbl_e2e_group",
		map[string]any{"id": req.ID},
		updates,
	).Exec()
	return err
}

// boolToInt 布尔转整数（数据库存储用）。
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// Delete 删除分组（联动删除用例 - 由外键 CASCADE 处理）。
func (s *GroupStore) Delete(id int) error {
	_, err := common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_e2e_group WHERE id = ?`, id,
	).Exec()
	return err
}

// CaseStore 用例数据访问。
type CaseStore struct{}

func NewCaseStore() *CaseStore { return &CaseStore{} }

// List 列出用例（不含完整 steps/assertions）。
func (s *CaseStore) List(groupID int, keyword string, page, pageSize int) ([]map[string]any, int64, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	where := "1=1"
	args := []any{}
	if groupID > 0 {
		where += " AND c.group_id = ?"
		args = append(args, groupID)
	}
	if keyword != "" {
		where += " AND c.name LIKE ?"
		args = append(args, "%"+keyword+"%")
	}
	var total int64
	row, err := common.DbMain.Client.QueryBySql(
		"SELECT COUNT(*) AS cnt FROM tbl_e2e_case c WHERE "+where, args...,
	).One()
	if err != nil {
		return nil, 0, err
	}
	if v, ok := row["cnt"]; ok {
		switch x := v.(type) {
		case int:
			total = int64(x)
		case int64:
			total = x
		case float64:
			total = int64(x)
		}
	}
	args = append(args, pageSize, (page-1)*pageSize)
	rows, err := common.DbMain.Client.QueryBySql(`
		SELECT c.*,
			(SELECT status FROM tbl_e2e_run r WHERE r.case_id = c.id ORDER BY r.id DESC LIMIT 1) AS last_run_status,
			(SELECT started_at FROM tbl_e2e_run r WHERE r.case_id = c.id ORDER BY r.id DESC LIMIT 1) AS last_run_at
		FROM tbl_e2e_case c
		WHERE `+where+`
		ORDER BY c.id DESC
		LIMIT ? OFFSET ?`, args...,
	).All()
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// Get 获取完整用例。
func (s *CaseStore) Get(id int) (map[string]any, error) {
	if id <= 0 {
		return nil, nil
	}
	return common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_e2e_case WHERE id = ?`, id,
	).One()
}

// Create 新建用例。
func (s *CaseStore) Create(req *define.E2ECaseSaveRequest) (int64, error) {
	now := time.Now().Unix()
	steps := normalizeJSONArray(req.Steps)
	assertions := normalizeJSONArray(req.Assertions)
	variables := normalizeJSONObject(req.Variables)
	timeout := req.TimeoutSeconds
	if timeout <= 0 {
		timeout = 600
	}
	id, err := common.DbMain.Client.InsertBySql(`
		INSERT INTO tbl_e2e_case (group_id, name, env_url, env_base_url, steps, assertions, variables,
			timeout_seconds, tags, notification_enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.GroupID, req.Name, req.EnvURL, req.EnvBaseURL,
		string(steps), string(assertions), string(variables),
		timeout, req.Tags, boolToInt(req.NotificationEnabled),
		now, now,
	).Exec()
	return id, err
}

// Update 更新用例。
func (s *CaseStore) Update(req *define.E2ECaseSaveRequest) error {
	updates := map[string]any{"updated_at": time.Now().Unix()}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	updates["env_url"] = req.EnvURL
	updates["env_base_url"] = req.EnvBaseURL
	if len(req.Steps) > 0 {
		updates["steps"] = string(normalizeJSONArray(req.Steps))
	}
	if len(req.Assertions) > 0 {
		updates["assertions"] = string(normalizeJSONArray(req.Assertions))
	}
	if len(req.Variables) > 0 {
		updates["variables"] = string(normalizeJSONObject(req.Variables))
	}
	if req.TimeoutSeconds > 0 {
		updates["timeout_seconds"] = req.TimeoutSeconds
	}
	updates["tags"] = req.Tags
	updates["notification_enabled"] = boolToInt(req.NotificationEnabled)
	_, err := common.DbMain.Client.QuickUpdate(
		"tbl_e2e_case",
		map[string]any{"id": req.ID},
		updates,
	).Exec()
	return err
}

// Delete 删除用例。
func (s *CaseStore) Delete(id int) error {
	_, err := common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_e2e_case WHERE id = ?`, id,
	).Exec()
	return err
}

// RunStore 执行记录数据访问。
type RunStore struct{}

func NewRunStore() *RunStore { return &RunStore{} }

// Create 创建执行记录。
func (s *RunStore) Create(caseID, groupID int, totalSteps int, triggerType string) (int64, error) {
	now := time.Now().Unix()
	id, err := common.DbMain.Client.InsertBySql(`
		INSERT INTO tbl_e2e_run (case_id, group_id, status, total_steps, started_at, trigger_type, created_at)
		VALUES (?, ?, 'pending', ?, ?, ?, ?)`,
		caseID, groupID, totalSteps, now, triggerType, now,
	).Exec()
	return id, err
}

// UpdateStatus 更新执行状态。
func (s *RunStore) UpdateStatus(runID int64, status string, startedAtMs int64, fields map[string]any) error {
	updates := map[string]any{"status": status}
	for k, v := range fields {
		updates[k] = v
	}
	if status == define.E2ERunStatusPassed || status == define.E2ERunStatusFailed ||
		status == define.E2ERunStatusStopped || status == define.E2ERunStatusError {
		now := time.Now().UnixMilli()
		updates["finished_at"] = now / 1000
		if startedAtMs > 0 {
			updates["duration_ms"] = now - startedAtMs
		}
	}
	_, err := common.DbMain.Client.QuickUpdate(
		"tbl_e2e_run",
		map[string]any{"id": runID},
		updates,
	).Exec()
	return err
}

// IncrementStepStats 累加步骤通过 / 失败计数。
func (s *RunStore) IncrementStepStats(runID int64, passed, failed int) error {
	_, err := common.DbMain.Client.ExecBySql(`
		UPDATE tbl_e2e_run
		SET passed_steps = passed_steps + ?, failed_steps = failed_steps + ?
		WHERE id = ?`, passed, failed, runID,
	).Exec()
	return err
}

// IncrementAssertStats 累加断言通过 / 失败计数。
func (s *RunStore) IncrementAssertStats(runID int64, passed, failed int) error {
	_, err := common.DbMain.Client.ExecBySql(`
		UPDATE tbl_e2e_run
		SET passed_asserts = passed_asserts + ?, failed_asserts = failed_asserts + ?
		WHERE id = ?`, passed, failed, runID,
	).Exec()
	return err
}

// GetRunDetail 获取执行详情（包含 joins）。
func (s *RunStore) GetRunDetail(runID int64) (map[string]any, error) {
	if runID <= 0 {
		return nil, nil
	}
	row, err := common.DbMain.Client.QueryBySql(`
		SELECT r.*, c.name AS case_name, g.name AS group_name
		FROM tbl_e2e_run r
		LEFT JOIN tbl_e2e_case c ON c.id = r.case_id
		LEFT JOIN tbl_e2e_group g ON g.id = r.group_id
		WHERE r.id = ?`, runID,
	).One()
	if err != nil || len(row) == 0 {
		return row, err
	}
	if caseID := cast.ToInt(row["case_id"]); caseID > 0 {
		if c, err := common.DbMain.Client.QueryBySql(
			`SELECT id, name, env_url, env_base_url, steps, assertions, variables FROM tbl_e2e_case WHERE id = ?`, caseID,
		).One(); err == nil {
			row["case_detail"] = c
		}
	}
	return row, err
}

// ListRunsArgs 列表参数。
type ListRunsArgs struct {
	Page     int
	PageSize int
	CaseID   int
	GroupID  int
	Status   string
}

// ListRuns 列出执行记录。
func (s *RunStore) ListRuns(args *ListRunsArgs) ([]map[string]any, int64, error) {
	if args.PageSize <= 0 {
		args.PageSize = 20
	}
	if args.Page <= 0 {
		args.Page = 1
	}
	where := "1=1"
	sqlArgs := []any{}
	if args.CaseID > 0 {
		where += " AND r.case_id = ?"
		sqlArgs = append(sqlArgs, args.CaseID)
	}
	if args.GroupID > 0 {
		where += " AND r.group_id = ?"
		sqlArgs = append(sqlArgs, args.GroupID)
	}
	if args.Status != "" {
		where += " AND r.status = ?"
		sqlArgs = append(sqlArgs, args.Status)
	}
	var total int64
	row, err := common.DbMain.Client.QueryBySql(
		"SELECT COUNT(*) AS cnt FROM tbl_e2e_run r WHERE "+where, sqlArgs...,
	).One()
	if err != nil {
		return nil, 0, err
	}
	if v, ok := row["cnt"]; ok {
		switch x := v.(type) {
		case int:
			total = int64(x)
		case int64:
			total = x
		case float64:
			total = int64(x)
		}
	}
	sqlArgs = append(sqlArgs, args.PageSize, (args.Page-1)*args.PageSize)
	rows, err := common.DbMain.Client.QueryBySql(`
		SELECT r.*, c.name AS case_name, g.name AS group_name
		FROM tbl_e2e_run r
		LEFT JOIN tbl_e2e_case c ON c.id = r.case_id
		LEFT JOIN tbl_e2e_group g ON g.id = r.group_id
		WHERE `+where+`
		ORDER BY r.id DESC
		LIMIT ? OFFSET ?`, sqlArgs...,
	).All()
	if err != nil {
		return nil, 0, err
	}
	return rows, total, nil
}

// StepStore 步骤执行记录数据访问。
type StepStore struct{}

func NewStepStore() *StepStore { return &StepStore{} }

// CreateStep 创建步骤执行记录。
func (s *StepStore) CreateStep(runID int64, stepIndex int, stepID, stepType, version string,
	configJSON, description, status, errorMsg, screenshot string, durationMs int64) (int64, error) {
	now := time.Now().Unix()
	id, err := common.DbMain.Client.InsertBySql(`
		INSERT INTO tbl_e2e_run_step (run_id, step_index, step_id, step_type, step_version,
			step_config, description, status, error_message, screenshot_path, duration_ms, executed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		runID, stepIndex, stepID, stepType, version, configJSON, description,
		status, errorMsg, screenshot, durationMs, time.Now().Unix(), now,
	).Exec()
	return id, err
}

// ListByRun 获取执行的所有步骤。
func (s *StepStore) ListByRun(runID int64) ([]map[string]any, error) {
	return common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_e2e_run_step WHERE run_id = ? ORDER BY step_index ASC`, runID,
	).All()
}

// CreateAssertion 创建断言执行记录。
func (s *StepStore) CreateAssertion(runID, runStepID int64, assertionID, assertionType, version string,
	configJSON, status, expected, actual, errorMsg, matchedReqURL, matchedReqID string) (int64, error) {
	now := time.Now().Unix()
	id, err := common.DbMain.Client.InsertBySql(`
		INSERT INTO tbl_e2e_run_assertion (run_id, run_step_id, assertion_id, assertion_type,
			assertion_version, assertion_config, status, expected, actual, error_message,
			matched_request_url, matched_request_id, executed_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		runID, runStepID, assertionID, assertionType, version, configJSON,
		status, expected, actual, errorMsg, matchedReqURL, matchedReqID,
		time.Now().Unix(), now,
	).Exec()
	return id, err
}

// ListAssertionsByRun 获取执行的所有断言。
func (s *StepStore) ListAssertionsByRun(runID int64) ([]map[string]any, error) {
	return common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_e2e_run_assertion WHERE run_id = ? ORDER BY id ASC`, runID,
	).All()
}

// RequestStore 捕获请求数据访问。
type RequestStore struct{}

func NewRequestStore() *RequestStore { return &RequestStore{} }

// Insert 写入捕获请求。
func (s *RequestStore) Insert(req *interceptor.CapturedRequest, runID int64, runStepID int) error {
	headersJSON, _ := json.Marshal(req.Headers)
	var respStatus int
	var respBody string
	respHeaders := map[string]string{}
	respTimeMs := 0
	if req.Response != nil {
		respStatus = req.Response.Status
		respBody = req.Response.Body
		respHeaders = req.Response.Headers
		respTimeMs = req.Response.TimeMs
	}
	respHeadersJSON, _ := json.Marshal(respHeaders)
	_, err := common.DbMain.Client.ExecBySql(`
		INSERT OR REPLACE INTO tbl_e2e_captured_request
		(id, run_id, run_step_id, url, method, request_headers, request_body,
			response_status, response_headers, response_body, response_time_ms, matched, matched_by, captured_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		req.ID, runID, runStepID, req.URL, req.Method,
		string(headersJSON), req.PostData,
		respStatus, string(respHeadersJSON), respBody,
		respTimeMs,
		boolToInt(req.Matched), req.MatchedBy,
		req.CapturedAt.Unix(),
	).Exec()
	return err
}

// ListByRun 获取某个 run 的所有捕获请求。
func (s *RequestStore) ListByRun(runID int64, stepID string) ([]map[string]any, error) {
	if stepID != "" {
		return common.DbMain.Client.QueryBySql(`
			SELECT cr.*, rs.step_id AS step_id
			FROM tbl_e2e_captured_request cr
			LEFT JOIN tbl_e2e_run_step rs ON rs.id = cr.run_step_id
			WHERE cr.run_id = ? AND rs.step_id = ?
			ORDER BY cr.captured_at ASC`, runID, stepID,
		).All()
	}
	return common.DbMain.Client.QueryBySql(`
		SELECT cr.*, rs.step_id AS step_id
		FROM tbl_e2e_captured_request cr
		LEFT JOIN tbl_e2e_run_step rs ON rs.id = cr.run_step_id
		WHERE cr.run_id = ?
		ORDER BY cr.captured_at ASC`, runID,
	).All()
}

// GetByID 根据 run_id 和 request_id 获取单个请求详情。
func (s *RequestStore) GetByID(runID int64, requestID string) (map[string]any, error) {
	return common.DbMain.Client.QueryBySql(`
		SELECT cr.*, rs.step_id AS step_id
		FROM tbl_e2e_captured_request cr
		LEFT JOIN tbl_e2e_run_step rs ON rs.id = cr.run_step_id
		WHERE cr.run_id = ? AND cr.id = ?
		LIMIT 1`, runID, requestID,
	).One()
}

// RecordSessionStore 录制会话数据访问。
type RecordSessionStore struct{}

func NewRecordSessionStore() *RecordSessionStore { return &RecordSessionStore{} }

// Create 创建录制会话（自增 row_id + 业务 session_id）。
func (s *RecordSessionStore) Create(name, sessionID, envURL, envBaseURL string, caseID, groupID int, browserID string) (int64, error) {
	now := time.Now().Unix()
	id, err := common.DbMain.Client.InsertBySql(`
		INSERT INTO tbl_e2e_record_session (session_id, case_id, group_id, env_url, env_base_url, browser_id, name, steps, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, '[]', 'recording', ?, ?)`,
		sessionID, caseID, groupID, envURL, envBaseURL, browserID, name, now, now,
	).Exec()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// GetByID 按自增 ID 查询。
func (s *RecordSessionStore) GetByID(id int64) (map[string]any, error) {
	if id <= 0 {
		return nil, nil
	}
	return common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_e2e_record_session WHERE id = ?`, id,
	).One()
}

// GetBySessionID 按业务 session_id 查询。
func (s *RecordSessionStore) GetBySessionID(sessionID string) (map[string]any, error) {
	if sessionID == "" {
		return nil, nil
	}
	return common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_e2e_record_session WHERE session_id = ?`, sessionID,
	).One()
}

// List 列出录制会话（可选 case_id 过滤）。
func (s *RecordSessionStore) List(caseID int, status string, page, pageSize int) ([]map[string]any, int, error) {
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	var (
		rows []map[string]any
		err  error
	)
	if caseID > 0 && status != "" {
		rows, err = common.DbMain.Client.QueryBySql(
			`SELECT * FROM tbl_e2e_record_session WHERE case_id = ? AND status = ? ORDER BY id DESC LIMIT ? OFFSET ?`,
			caseID, status, pageSize, (page-1)*pageSize,
		).All()
	} else if caseID > 0 {
		rows, err = common.DbMain.Client.QueryBySql(
			`SELECT * FROM tbl_e2e_record_session WHERE case_id = ? ORDER BY id DESC LIMIT ? OFFSET ?`,
			caseID, pageSize, (page-1)*pageSize,
		).All()
	} else if status != "" {
		rows, err = common.DbMain.Client.QueryBySql(
			`SELECT * FROM tbl_e2e_record_session WHERE status = ? ORDER BY id DESC LIMIT ? OFFSET ?`,
			status, pageSize, (page-1)*pageSize,
		).All()
	} else {
		rows, err = common.DbMain.Client.QueryBySql(
			`SELECT * FROM tbl_e2e_record_session ORDER BY id DESC LIMIT ? OFFSET ?`,
			pageSize, (page-1)*pageSize,
		).All()
	}
	if err != nil {
		return nil, 0, err
	}
	// total
	var total int
	totalRow, _ := common.DbMain.Client.QueryBySql(
		`SELECT COUNT(1) AS c FROM tbl_e2e_record_session`,
	).One()
	total = cast.ToInt(totalRow["c"])
	return rows, total, nil
}

// AppendStep 向录制会话追加步骤（按 session_id 业务主键）。
func (s *RecordSessionStore) AppendStep(sessionID string, stepJSON string) error {
	now := time.Now().Unix()
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT steps FROM tbl_e2e_record_session WHERE session_id = ?`, sessionID,
	).One()
	if err != nil || len(row) == 0 {
		return err
	}
	existing := cast.ToString(row["steps"])
	if existing == "" {
		existing = "[]"
	}
	var arr []any
	_ = json.Unmarshal([]byte(existing), &arr)
	var step any
	_ = json.Unmarshal([]byte(stepJSON), &step)
	arr = append(arr, step)
	out, _ := json.Marshal(arr)
	_, err = common.DbMain.Client.ExecBySql(`
		UPDATE tbl_e2e_record_session SET steps = ?, updated_at = ? WHERE session_id = ?`,
		string(out), now, sessionID,
	).Exec()
	return err
}

// UpdateStep 更新录制会话中的某一步（按 step_id 匹配）。
func (s *RecordSessionStore) UpdateStep(sessionID, stepID, stepJSON string) error {
	now := time.Now().Unix()
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT steps FROM tbl_e2e_record_session WHERE session_id = ?`, sessionID,
	).One()
	if err != nil || len(row) == 0 {
		return err
	}
	existing := cast.ToString(row["steps"])
	if existing == "" {
		existing = "[]"
	}
	var arr []json.RawMessage
	_ = json.Unmarshal([]byte(existing), &arr)
	updated := false
	for i := range arr {
		var probe map[string]any
		if err := json.Unmarshal(arr[i], &probe); err != nil {
			continue
		}
		if cast.ToString(probe["id"]) == stepID {
			arr[i] = json.RawMessage(stepJSON)
			updated = true
			break
		}
	}
	if !updated {
		return fmt.Errorf("未找到 step_id=%s", stepID)
	}
	out, _ := json.Marshal(arr)
	_, err = common.DbMain.Client.ExecBySql(`
		UPDATE tbl_e2e_record_session SET steps = ?, updated_at = ? WHERE session_id = ?`,
		string(out), now, sessionID,
	).Exec()
	return err
}

// DeleteStep 删除录制会话中的某一步。
func (s *RecordSessionStore) DeleteStep(sessionID, stepID string) error {
	now := time.Now().Unix()
	row, err := common.DbMain.Client.QueryBySql(
		`SELECT steps FROM tbl_e2e_record_session WHERE session_id = ?`, sessionID,
	).One()
	if err != nil || len(row) == 0 {
		return err
	}
	existing := cast.ToString(row["steps"])
	if existing == "" {
		existing = "[]"
	}
	var arr []json.RawMessage
	_ = json.Unmarshal([]byte(existing), &arr)
	out := make([]json.RawMessage, 0, len(arr))
	for i := range arr {
		var probe map[string]any
		if err := json.Unmarshal(arr[i], &probe); err != nil {
			continue
		}
		if cast.ToString(probe["id"]) != stepID {
			out = append(out, arr[i])
		}
	}
	data, _ := json.Marshal(out)
	_, err = common.DbMain.Client.ExecBySql(`
		UPDATE tbl_e2e_record_session SET steps = ?, updated_at = ? WHERE session_id = ?`,
		string(data), now, sessionID,
	).Exec()
	return err
}

// Get 获取录制会话（兼容旧接口）。
func (s *RecordSessionStore) Get(sessionID string) (map[string]any, error) {
	if sessionID == "" {
		return nil, nil
	}
	return common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_e2e_record_session WHERE session_id = ?`, sessionID,
	).One()
}

// Delete 删除录制会话（按 session_id）。
func (s *RecordSessionStore) Delete(sessionID string) error {
	_, err := common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_e2e_record_session WHERE session_id = ?`, sessionID,
	).Exec()
	return err
}

// DeleteByID 删除录制会话（按自增 ID）。
func (s *RecordSessionStore) DeleteByID(id int64) error {
	_, err := common.DbMain.Client.ExecBySql(
		`DELETE FROM tbl_e2e_record_session WHERE id = ?`, id,
	).Exec()
	return err
}

// UpdateStatus 修改会话状态。
func (s *RecordSessionStore) UpdateStatus(sessionID, status string) error {
	now := time.Now().Unix()
	_, err := common.DbMain.Client.ExecBySql(
		`UPDATE tbl_e2e_record_session SET status = ?, updated_at = ? WHERE session_id = ?`,
		status, now, sessionID,
	).Exec()
	return err
}

// normalizeJSONArray 兜底为合法 JSON 数组。
func normalizeJSONArray(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "[]"
	}
	var arr []any
	if err := json.Unmarshal(raw, &arr); err != nil {
		return "[]"
	}
	out, _ := json.Marshal(arr)
	return string(out)
}

// normalizeJSONObject 兜底为合法 JSON 对象。
func normalizeJSONObject(raw json.RawMessage) string {
	if len(raw) == 0 {
		return "{}"
	}
	var obj map[string]any
	if err := json.Unmarshal(raw, &obj); err != nil {
		return "{}"
	}
	out, _ := json.Marshal(obj)
	return string(out)
}
