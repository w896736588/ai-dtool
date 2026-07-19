// E2E 业务层入口（属于 business 包）。
package business

import (
	"dev_tool/internal/app/dtool/component/e2e/store"
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"strings"
	"sync"
	"time"

	"github.com/spf13/cast"
)

var (
	e2eOnce      sync.Once
	e2eEngine    *E2EEngine
)

// GetE2EEngine 全局执行引擎单例。
func GetE2EEngine() *E2EEngine {
	e2eOnce.Do(func() {
		e2eEngine = NewE2EEngine()
	})
	return e2eEngine
}

// ---- 业务方法：分组 ----

// E2EGroupList 列出分组。
func E2EGroupList(req *define.E2EGroupListRequest) (*define.E2EGroupListResponse, error) {
	gs := store.NewGroupStore()
	rows, total, err := gs.List(req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	items := make([]define.E2EGroupItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, mapE2EGroupRow(r))
	}
	page, pageSize := req.Page, req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	totalPage := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &define.E2EGroupListResponse{
		List: items,
		Pagination: define.Pagination{
			Page: page, PageSize: pageSize, Total: total, TotalPage: totalPage,
		},
	}, nil
}

// E2EGroupCreate 创建分组。
func E2EGroupCreate(req *define.E2EGroupCreateRequest) (int64, error) {
	gs := store.NewGroupStore()
	return gs.Create(req)
}

// E2EGroupUpdate 更新分组。
func E2EGroupUpdate(req *define.E2EGroupUpdateRequest) error {
	gs := store.NewGroupStore()
	return gs.Update(req)
}

// E2EGroupDelete 删除分组。
func E2EGroupDelete(id int) error {
	gs := store.NewGroupStore()
	return gs.Delete(id)
}

// ---- 业务方法：用例 ----

// E2ECaseList 用例列表。
func E2ECaseList(req *define.E2ECaseListRequest) (*define.E2ECaseListResponse, error) {
	cs := store.NewCaseStore()
	rows, total, err := cs.List(req.GroupID, req.Keyword, req.Page, req.PageSize)
	if err != nil {
		return nil, err
	}
	items := make([]define.E2ECaseItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, mapE2ECaseRow(r))
	}
	page, pageSize := req.Page, req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	totalPage := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &define.E2ECaseListResponse{
		List: items,
		Pagination: define.Pagination{
			Page: page, PageSize: pageSize, Total: total, TotalPage: totalPage,
		},
	}, nil
}

// E2ECaseDetail 用例详情。
func E2ECaseDetail(id int) (map[string]any, error) {
	cs := store.NewCaseStore()
	return cs.Get(id)
}

// E2ECaseSave 创建或更新用例。
func E2ECaseSave(req *define.E2ECaseSaveRequest) (int64, error) {
	cs := store.NewCaseStore()
	if req.ID > 0 {
		if err := cs.Update(req); err != nil {
			return int64(req.ID), err
		}
		return int64(req.ID), nil
	}
	return cs.Create(req)
}

// E2ECaseDelete 删除用例。
func E2ECaseDelete(id int) error {
	cs := store.NewCaseStore()
	return cs.Delete(id)
}

// ---- 业务方法：执行 ----

// E2ERunExecute 触发执行（异步）。
func E2ERunExecute(caseID int) (int64, error) {
	runID, err := GetE2EEngine().RunCase(caseID, "manual")
	if err != nil {
		return runID, err
	}
	go E2ENotifyRunCompleted(runID)
	return runID, nil
}

// E2ERunExecuteBatch 批量执行（按 group）。
func E2ERunExecuteBatch(groupID int) ([]int64, error) {
	cs := store.NewCaseStore()
	rows, _, err := cs.List(groupID, "", 1, 1000)
	if err != nil {
		return nil, err
	}
	ids := make([]int64, 0, len(rows))
	for _, r := range rows {
		caseID := e2eToInt(r["id"])
		if caseID <= 0 {
			continue
		}
		runID, err := GetE2EEngine().RunCase(caseID, "batch")
		if err != nil {
			continue
		}
		ids = append(ids, runID)
		go E2ENotifyRunCompleted(runID)
	}
	return ids, nil
}

// E2ERunStop 停止执行。
func E2ERunStop(runID int64) error {
	return GetE2EEngine().StopRun(runID)
}

// E2ERunList 执行列表。
func E2ERunList(req *define.E2ERunListRequest) (*define.E2ERunListResponse, error) {
	rs := store.NewRunStore()
	args := &store.ListRunsArgs{
		Page: req.Page, PageSize: req.PageSize,
		CaseID: req.CaseID, GroupID: req.GroupID, Status: req.Status,
	}
	rows, total, err := rs.ListRuns(args)
	if err != nil {
		return nil, err
	}
	items := make([]define.E2ERunItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, define.E2ERunItem{
			ID:            e2eToInt64(r["id"]),
			CaseID:        e2eToInt(r["case_id"]),
			CaseName:      e2eToStr(r["case_name"]),
			GroupID:       e2eToInt(r["group_id"]),
			GroupName:     e2eToStr(r["group_name"]),
			Status:        e2eToStr(r["status"]),
			TotalSteps:    e2eToInt(r["total_steps"]),
			PassedSteps:   e2eToInt(r["passed_steps"]),
			FailedSteps:   e2eToInt(r["failed_steps"]),
			TotalAsserts:  e2eToInt(r["total_asserts"]),
			PassedAsserts: e2eToInt(r["passed_asserts"]),
			FailedAsserts: e2eToInt(r["failed_asserts"]),
			StartedAt:     e2eToInt64(r["started_at"]),
			FinishedAt:    e2eToInt64(r["finished_at"]),
			DurationMs:    e2eToInt(r["duration_ms"]),
			TriggerType:   e2eToStr(r["trigger_type"]),
			ErrorMessage:  e2eToStr(r["error_message"]),
			CreatedAt:     e2eToInt64(r["created_at"]),
		})
	}
	page, pageSize := req.Page, req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	if page <= 0 {
		page = 1
	}
	totalPage := int((total + int64(pageSize) - 1) / int64(pageSize))
	return &define.E2ERunListResponse{
		List: items,
		Pagination: define.Pagination{
			Page: page, PageSize: pageSize, Total: total, TotalPage: totalPage,
		},
	}, nil
}

// E2ERunDetail 详情（run + steps + assertions + requests）。
func E2ERunDetail(runID int64) (map[string]any, error) {
	rs := store.NewRunStore()
	ss := store.NewStepStore()
	rqs := store.NewRequestStore()
	run, err := rs.GetRunDetail(runID)
	if err != nil {
		return nil, err
	}
	steps, _ := ss.ListByRun(runID)
	assertions, _ := ss.ListAssertionsByRun(runID)
	requests, _ := rqs.ListByRun(runID, "")
	return map[string]any{
		"run":        run,
		"steps":      steps,
		"assertions": assertions,
		"requests":   requests,
	}, nil
}

// E2ERunRequests 某个 run 的请求追踪列表。
func E2ERunRequests(req *define.E2ERunRequestsRequest) ([]map[string]any, error) {
	rqs := store.NewRequestStore()
	return rqs.ListByRun(req.RunID, req.StepID)
}

// E2ERunRequestDetail 单个请求详情。
func E2ERunRequestDetail(runID int64, requestID string) (map[string]any, error) {
	rqs := store.NewRequestStore()
	return rqs.GetByID(runID, requestID)
}

// ---- 业务方法：类型清单 ----

// E2EStepTypeList 返回前端动态渲染用的步骤类型列表。
func E2EStepTypeList() *define.E2EStepTypeListResponse {
	items := []define.E2EStepTypeMeta{
		{Type: "open_env", BaseType: "open_env", Version: "1.0", Label: "打开环境", Group: "action", ConfigKeys: []string{"url", "wait_load"}},
		{Type: "click_v1", BaseType: "click", Version: "1.0", Label: "点击元素", Group: "action", ConfigKeys: []string{"selector", "selector_type"}},
		{Type: "click_by_position_v1", BaseType: "click_by_position", Version: "1.0", Label: "坐标点击", Group: "action", ConfigKeys: []string{"x", "y", "viewport_width", "viewport_height"}},
		{Type: "right_click_v1", BaseType: "right_click", Version: "1.0", Label: "右键点击", Group: "action", ConfigKeys: []string{"x", "y", "viewport_width", "viewport_height"}},
		{Type: "input_v1", BaseType: "input", Version: "1.0", Label: "输入（固定）", Group: "action", ConfigKeys: []string{"selector", "value", "clear_before"}},
		{Type: "input_v2", BaseType: "input", Version: "2.0", Label: "输入（多输入源）", Group: "action", ConfigKeys: []string{"selector", "source_type"}},
		{Type: "hover_v1", BaseType: "hover", Version: "1.0", Label: "悬停", Group: "action", ConfigKeys: []string{"selector"}},
		{Type: "select_v1", BaseType: "select", Version: "1.0", Label: "下拉选择", Group: "action", ConfigKeys: []string{"selector", "value"}},
		{Type: "navigate_v1", BaseType: "navigate", Version: "1.0", Label: "页面导航", Group: "action", ConfigKeys: []string{"url"}},
		{Type: "go_back_v1", BaseType: "go_back", Version: "1.0", Label: "返回上一页", Group: "action"},
		{Type: "reload_v1", BaseType: "reload", Version: "1.0", Label: "刷新", Group: "action"},
		{Type: "press_key_v1", BaseType: "press_key", Version: "1.0", Label: "按键", Group: "action", ConfigKeys: []string{"key"}},
		{Type: "scroll_v1", BaseType: "scroll", Version: "1.0", Label: "滚动页面", Group: "action", ConfigKeys: []string{"delta_x", "delta_y"}},
		{Type: "wait_element_v1", BaseType: "wait_element", Version: "1.0", Label: "等待元素", Group: "wait", ConfigKeys: []string{"selector", "timeout_ms"}},
		{Type: "wait_timeout_v1", BaseType: "wait_timeout", Version: "1.0", Label: "固定等待", Group: "wait", ConfigKeys: []string{"duration_ms"}},
		{Type: "wait_after_v1", BaseType: "wait_after", Version: "1.0", Label: "步骤后等待", Group: "wait", ConfigKeys: []string{"duration_ms"}},
		{Type: "extract_text_v1", BaseType: "extract_text", Version: "1.0", Label: "提取文本", Group: "extract", ConfigKeys: []string{"selector", "extract_to"}},
		{Type: "extract_attr_v1", BaseType: "extract_attr", Version: "1.0", Label: "提取属性", Group: "extract", ConfigKeys: []string{"selector", "attribute", "extract_to"}},
		{Type: "extract_api_v1", BaseType: "extract_api", Version: "1.0", Label: "提取API响应", Group: "extract", ConfigKeys: []string{"find_by_pattern", "response_json_path", "extract_to"}},
		{Type: "script_v1", BaseType: "script", Version: "1.0", Label: "执行 JS", Group: "script", ConfigKeys: []string{"code"}},
	}
	return &define.E2EStepTypeListResponse{Items: items}
}

// E2EAssertionTypeList 返回断言类型清单。
func E2EAssertionTypeList() *define.E2EAssertionTypeListResponse {
	items := []define.E2EAssertionTypeMeta{
		{Type: "assert_text_v1", BaseType: "assert_text", Version: "1.0", Label: "文本断言", Group: "page"},
		{Type: "assert_text_v2", BaseType: "assert_text", Version: "2.0", Label: "文本断言（增强）", Group: "page"},
		{Type: "assert_element_v1", BaseType: "assert_element", Version: "1.0", Label: "元素断言", Group: "page"},
		{Type: "assert_url_v1", BaseType: "assert_url", Version: "1.0", Label: "URL 断言", Group: "page"},
		{Type: "assert_title_v1", BaseType: "assert_title", Version: "1.0", Label: "标题断言", Group: "page"},
		{Type: "assert_api_response_v1", BaseType: "assert_api_response", Version: "1.0", Label: "API 响应断言（基于捕获）", Group: "api"},
		{Type: "assert_api_request_v1", BaseType: "assert_api_request", Version: "1.0", Label: "API 请求断言（基于捕获）", Group: "api"},
		{Type: "assert_variable_v1", BaseType: "assert_variable", Version: "1.0", Label: "变量断言", Group: "variable"},
	}
	return &define.E2EAssertionTypeListResponse{Items: items}
}

// ---- 辅助 ----

func e2eToInt(v any) int {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case int:
		return x
	case int64:
		return int(x)
	case float64:
		return int(x)
	case string:
		var n int
		_, _ = jsonScan(x, &n)
		return n
	}
	return 0
}

func e2eToBool(v any) bool {
	if v == nil {
		return false
	}
	switch x := v.(type) {
	case bool:
		return x
	case int:
		return x != 0
	case int64:
		return x != 0
	case float64:
		return x != 0
	}
	return false
}

func e2eToInt64(v any) int64 {
	if v == nil {
		return 0
	}
	switch x := v.(type) {
	case int:
		return int64(x)
	case int64:
		return x
	case float64:
		return int64(x)
	case string:
		var n int64
		_, _ = jsonScan(x, &n)
		return n
	}
	return 0
}

func e2eToStr(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

func mapE2EGroupRow(r map[string]any) define.E2EGroupItem {
	return define.E2EGroupItem{
		ID:                  e2eToInt(r["id"]),
		Name:                e2eToStr(r["name"]),
		WorkflowTaskID:      e2eToInt(r["workflow_task_id"]),
		NotificationEnabled: e2eToBool(r["notification_enabled"]),
		WebhookConfigID:     e2eToInt(r["webhook_config_id"]),
		CaseCount:           e2eToInt(r["case_count"]),
		CreatedAt:           e2eToInt64(r["created_at"]),
		UpdatedAt:           e2eToInt64(r["updated_at"]),
	}
}

func mapE2ECaseRow(r map[string]any) define.E2ECaseItem {
	stepsRaw := e2eToStr(r["steps"])
	stepCount := 0
	if stepsRaw != "" {
		var arr []any
		_ = json.Unmarshal([]byte(stepsRaw), &arr)
		stepCount = len(arr)
	}
	assertsRaw := e2eToStr(r["assertions"])
	assertsCount := 0
	if assertsRaw != "" {
		var arr []any
		_ = json.Unmarshal([]byte(assertsRaw), &arr)
		assertsCount = len(arr)
	}
	return define.E2ECaseItem{
		ID:                  e2eToInt(r["id"]),
		GroupID:             e2eToInt(r["group_id"]),
		Name:                e2eToStr(r["name"]),
		EnvURL:              e2eToStr(r["env_url"]),
		EnvBaseURL:          e2eToStr(r["env_base_url"]),
		StepCount:           stepCount,
		AssertionCount:      assertsCount,
		Tags:                e2eToStr(r["tags"]),
		TimeoutSeconds:      e2eToInt(r["timeout_seconds"]),
		NotificationEnabled: e2eToBool(r["notification_enabled"]),
		LastRunStatus:       e2eToStr(r["last_run_status"]),
		LastRunAt:           e2eToInt64(r["last_run_at"]),
		CreatedAt:           e2eToInt64(r["created_at"]),
		UpdatedAt:           e2eToInt64(r["updated_at"]),
	}
}

// jsonScan 简易整数解析（不引入 strconv 减少 imports）。
func jsonScan(s string, target any) (int, error) {
	var n int64
	var sign int64 = 1
	i := 0
	for i < len(s) && s[i] == ' ' {
		i++
	}
	if i < len(s) && (s[i] == '+' || s[i] == '-') {
		if s[i] == '-' {
			sign = -1
		}
		i++
	}
	for i < len(s) && s[i] >= '0' && s[i] <= '9' {
		n = n*10 + int64(s[i]-'0')
		i++
	}
	n *= sign
	switch t := target.(type) {
	case *int:
		*t = int(n)
	case *int64:
		*t = n
	}
	return i, nil
}

// =============== 录制功能（v5.0）业务实现 ===============

// E2ERecordSessionCreate 创建录制会话。
// v6：当请求体携带 smart_link 绑定字段时，附加调用 UpdateSmartLink 写入。
func E2ERecordSessionCreate(req *define.E2ERecordSessionCreateRequest) (*define.E2ERecordSessionCreateResponse, error) {
	if req == nil {
		return nil, errors.New("请求不能为空")
	}
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = "rec_" + time.Now().Format("20060102150405") + "_" + cast.ToString(time.Now().UnixNano())
	}
	id, err := store.NewRecordSessionStore().Create(
		strings.TrimSpace(req.SessionName), sessionID,
		strings.TrimSpace(req.EnvURL), strings.TrimSpace(req.EnvBaseURL),
		req.CaseID, req.GroupID, strings.TrimSpace(req.BrowserID),
	)
	if err != nil {
		return nil, err
	}
	if req.SmartLinkID > 0 || req.UserName != "" || req.LinkID > 0 || req.WSToken != "" || req.RecorderURL != "" {
		if err := store.NewRecordSessionStore().UpdateSmartLink(
			id, req.SmartLinkID, req.UserName,
			req.WSToken, req.RecorderURL, req.LinkID,
		); err != nil {
			return nil, err
		}
	}
	return &define.E2ERecordSessionCreateResponse{
		ID:        id,
		SessionID: sessionID,
		Status:    "recording",
	}, nil
}

// E2ERecordSessionGet 获取录制会话详情。
func E2ERecordSessionGet(id int64) (*define.E2ERecordSessionDetail, error) {
	row, err := store.NewRecordSessionStore().GetByID(id)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, nil
	}
	return mapRecordSessionRow(row), nil
}

// E2ERecordSessionList 列出录制会话。
func E2ERecordSessionList(req *define.E2ERecordSessionListRequest) (*define.E2ERecordSessionListResponse, error) {
	page := req.Page
	if page <= 0 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize <= 0 {
		pageSize = 20
	}
	rows, total, err := store.NewRecordSessionStore().List(req.CaseID, req.Status, page, pageSize)
	if err != nil {
		return nil, err
	}
	items := make([]define.E2ERecordListItem, 0, len(rows))
	for _, r := range rows {
		items = append(items, define.E2ERecordListItem{
			ID:            int64(e2eToInt(r["row_id"])),
			SessionID:     e2eToStr(r["session_id"]),
			CaseID:        e2eToInt(r["case_id"]),
			GroupID:       e2eToInt(r["group_id"]),
			Name:          e2eToStr(r["name"]),
			EnvURL:        e2eToStr(r["env_url"]),
			EnvBaseURL:    e2eToStr(r["env_base_url"]),
			BrowserID:     e2eToStr(r["browser_id"]),
			Status:        e2eToStr(r["status"]),
			StepCount:     countJSONArray(r["steps"]),
			CreatedAt:     e2eToInt64(r["created_at"]),
			UpdatedAt:     e2eToInt64(r["updated_at"]),
		})
	}
	return &define.E2ERecordSessionListResponse{Items: items, Total: total, Page: page, PageSize: pageSize}, nil
}

// E2ERecordSessionDelete 删除录制会话。
func E2ERecordSessionDelete(id int64) error {
	row, err := store.NewRecordSessionStore().GetByID(id)
	if err != nil {
		return err
	}
	if row == nil {
		return nil
	}
	// 标记为 discarded 而非物理删除，便于追溯
	if e2eToStr(row["status"]) == "recording" {
		_ = store.NewRecordSessionStore().UpdateStatus(e2eToStr(row["session_id"]), "discarded")
		return nil
	}
	return store.NewRecordSessionStore().DeleteByID(id)
}

// E2ERecordStepAdd 追加一步。
func E2ERecordStepAdd(req *define.E2ERecordStepAddRequest) (*define.E2ERecordStepAddResponse, error) {
	row, err := store.NewRecordSessionStore().GetByID(req.SessionID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errors.New("录制会话不存在")
	}
	sessionID := e2eToStr(row["session_id"])
	step := req.Step
	// 缺省 ID
	if step.ID == "" {
		step.ID = "stp_" + cast.ToString(time.Now().UnixNano())
	}
	// 缺省 version
	if step.Version == "" {
		step.Version = "1.0"
	}
	// 自动追加 wait_after_ms：若前端没传，默认 200ms
	if step.WaitAfterMs <= 0 {
		step.WaitAfterMs = 200
	}
	step.RecordedAt = time.Now().UnixMilli()
	stepJSON, _ := json.Marshal(step)
	if err := store.NewRecordSessionStore().AppendStep(sessionID, string(stepJSON)); err != nil {
		return nil, err
	}
	return &define.E2ERecordStepAddResponse{StepID: step.ID, SessionID: sessionID, StepIndex: countJSONArray(row["steps"])}, nil
}

// E2ERecordStepUpdate 更新一步。
func E2ERecordStepUpdate(req *define.E2ERecordStepUpdateRequest) error {
	row, err := store.NewRecordSessionStore().GetByID(req.SessionID)
	if err != nil {
		return err
	}
	if row == nil {
		return errors.New("录制会话不存在")
	}
	sessionID := e2eToStr(row["session_id"])
	if req.Step.ID == "" {
		req.Step.ID = req.StepID
	}
	data, _ := json.Marshal(req.Step)
	return store.NewRecordSessionStore().UpdateStep(sessionID, req.StepID, string(data))
}

// E2ERecordStepDelete 删除一步。
func E2ERecordStepDelete(req *define.E2ERecordStepDeleteRequest) error {
	row, err := store.NewRecordSessionStore().GetByID(req.SessionID)
	if err != nil {
		return err
	}
	if row == nil {
		return errors.New("录制会话不存在")
	}
	sessionID := e2eToStr(row["session_id"])
	return store.NewRecordSessionStore().DeleteStep(sessionID, req.StepID)
}

// E2ERecordCommit 将录制会话落库为用例。
func E2ERecordCommit(req *define.E2ERecordCommitRequest) (*define.E2ERecordCommitResponse, error) {
	row, err := store.NewRecordSessionStore().GetByID(req.SessionID)
	if err != nil {
		return nil, err
	}
	if row == nil {
		return nil, errors.New("录制会话不存在")
	}
	steps := parseRecordedSteps(row["steps"])
	// 转换为 E2EStep（含断言）
	e2eSteps := make([]define.E2EStep, 0, len(steps))
	for _, s := range steps {
		e2eSteps = append(e2eSteps, define.E2EStep{
			ID:          s.ID,
			Type:        s.Type,
			Version:     s.Version,
			Description: s.Description,
			WaitAfterMs: s.WaitAfterMs,
			Config:      s.Config,
		})
	}
	stepsJSON, _ := json.Marshal(e2eSteps)
	// 收集所有断言
	var allAsserts []define.E2EAssertion
	for _, s := range steps {
		if len(s.Assertions) > 0 {
			var arr []define.E2EAssertion
			if err := json.Unmarshal(s.Assertions, &arr); err == nil {
				allAsserts = append(allAsserts, arr...)
			}
		}
	}
	assertsJSON, _ := json.Marshal(allAsserts)
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = e2eToStr(row["name"])
	}
	if name == "" {
		name = "录制用例 " + time.Now().Format("20060102 150405")
	}
	envURL := e2eToStr(row["env_url"])
	envBaseURL := e2eToStr(row["env_base_url"])

	// 调用 case store 的 Create / Update
	caseStore := store.NewCaseStore()
	var caseID int64
	if req.CaseID > 0 {
		// 更新现有用例
		caseID = int64(req.CaseID)
		updateReq := &define.E2ECaseSaveRequest{
			Name:           name,
			GroupID:        req.GroupID,
			EnvURL:         envURL,
			EnvBaseURL:     envBaseURL,
			Steps:          stepsJSON,
			Assertions:     assertsJSON,
			Tags:           strings.TrimSpace(req.Tags),
			TimeoutSeconds: req.TimeoutSeconds,
		}
		if err := caseStore.Update(updateReq); err != nil {
			return nil, err
		}
	} else {
		createReq := &define.E2ECaseSaveRequest{
			Name:           name,
			GroupID:        req.GroupID,
			EnvURL:         envURL,
			EnvBaseURL:     envBaseURL,
			Steps:          stepsJSON,
			Assertions:     assertsJSON,
			Tags:           strings.TrimSpace(req.Tags),
			TimeoutSeconds: req.TimeoutSeconds,
		}
		caseID, err = caseStore.Create(createReq)
		if err != nil {
			return nil, err
		}
	}

	// 标记会话状态
	_ = store.NewRecordSessionStore().UpdateStatus(e2eToStr(row["session_id"]), "committed")

	return &define.E2ERecordCommitResponse{
		CaseID:  caseID,
		Steps:   len(e2eSteps),
		GroupID: req.GroupID,
	}, nil
}

// ---- 录制辅助 ----

func mapRecordSessionRow(r map[string]any) *define.E2ERecordSessionDetail {
	if r == nil {
		return nil
	}
	steps := parseRecordedSteps(r["steps"])
	return &define.E2ERecordSessionDetail{
		ID:          int64(e2eToInt(r["row_id"])),
		SessionID:   e2eToStr(r["session_id"]),
		CaseID:      e2eToInt(r["case_id"]),
		GroupID:     e2eToInt(r["group_id"]),
		Name:        e2eToStr(r["name"]),
		EnvURL:      e2eToStr(r["env_url"]),
		EnvBaseURL:  e2eToStr(r["env_base_url"]),
		BrowserID:   e2eToStr(r["browser_id"]),
		SmartLinkID: e2eToInt(r["smart_link_id"]),
		LinkID:      e2eToInt(r["link_id"]),
		UserName:    e2eToStr(r["user_name"]),
		RecorderURL: e2eToStr(r["recorder_url"]),
		Status:      e2eToStr(r["status"]),
		Steps:       steps,
		CreatedAt:   e2eToInt64(r["created_at"]),
		UpdatedAt:   e2eToInt64(r["updated_at"]),
	}
}

func parseRecordedSteps(raw any) []define.RecordedStep {
	str := e2eToStr(raw)
	if str == "" {
		return nil
	}
	var steps []define.RecordedStep
	if err := json.Unmarshal([]byte(str), &steps); err != nil {
		return nil
	}
	return steps
}

func countJSONArray(raw any) int {
	str := e2eToStr(raw)
	if str == "" {
		return 0
	}
	var arr []any
	_ = json.Unmarshal([]byte(str), &arr)
	return len(arr)
}
