package common

import (
	"errors"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

const (
	taskWorkflowStatusInit              = `init`
	taskWorkflowStageIdle               = `idle`
	taskWorkflowStageRequirementFetch   = `requirement_fetch`
	taskWorkflowRequirementFetchIdle    = `idle`
	taskWorkflowRequirementFetchRunning = `running`
	taskWorkflowRequirementFetchSuccess = `success`
	taskWorkflowRequirementFetchFailed  = `failed`
)

const (
	taskWorkflowChatStatusRunning     = `running`
	taskWorkflowChatStatusCompleted   = `completed`
	taskWorkflowChatStatusError       = `error`
	taskWorkflowChatStatusInterrupted = `interrupted`
)

// TaskWorkflowCreateOrGetByHomeTaskID 查询或创建任务工作流主记录。
func (h *CSqlite) TaskWorkflowCreateOrGetByHomeTaskID(homeTaskID int) (map[string]any, error) {
	if homeTaskID <= 0 {
		return nil, errors.New(`任务id不能为空`)
	}
	existing, err := h.Client.QuickQuery(`tbl_task_workflow`, `*`, map[string]any{
		`home_task_id`: homeTaskID,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(existing) > 0 {
		return existing, nil
	}
	homeTaskInfo, err := h.HomeTaskRow(homeTaskID)
	if err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	newID, err := h.Client.QuickCreate(`tbl_task_workflow`, map[string]any{
		`home_task_id`:                  homeTaskID,
		`status`:                        taskWorkflowStatusInit,
		`current_stage`:                 taskWorkflowStageIdle,
		`requirement_fragment_id`:       strings.TrimSpace(cast.ToString(homeTaskInfo[`memory_fragment_id`])),
		`requirement_fetch_status`:      taskWorkflowRequirementFetchIdle,
		`requirement_fetch_started_at`:  0,
		`requirement_fetch_finished_at`: 0,
		`requirement_fetch_error`:       ``,
		`requirement_source_url`:        strings.TrimSpace(cast.ToString(homeTaskInfo[`tapd_url`])),
		`dev_plan_fragment_id`:          ``,
		`latest_plan_run_id`:            0,
		`latest_test_run_id`:            0,
		`base_branch`:                   ``,
		`feature_branch`:                ``,
		`last_error`:                    ``,
		`create_time`:                   now,
		`update_time`:                   now,
	}).Exec()
	if err != nil {
		return nil, err
	}
	return h.TaskWorkflowInfo(cast.ToInt(newID))
}

// TaskWorkflowInfo 查询单条任务工作流主记录。
func (h *CSqlite) TaskWorkflowInfo(id int) (map[string]any, error) {
	if id <= 0 {
		return nil, errors.New(`工作流id不能为空`)
	}
	info, err := h.Client.QuickQuery(`tbl_task_workflow`, `*`, map[string]any{
		`id`: id,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, errors.New(`工作流不存在`)
	}
	return info, nil
}

// TaskWorkflowBindDevPlanFragment 绑定开发执行片段 id。
func (h *CSqlite) TaskWorkflowBindDevPlanFragment(workflowID int, fragmentID string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	fragmentID = strings.TrimSpace(fragmentID)
	if fragmentID == `` {
		return errors.New(`开发执行片段id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`dev_plan_fragment_id`: fragmentID,
		`status`:               `dev_plan_ready`,
		`update_time`:          time.Now().Unix(),
	}).Exec()
	return err
}

// TaskWorkflowBindPlainTextReqFragment 绑定纯文本需求片段 id。
func (h *CSqlite) TaskWorkflowBindPlainTextReqFragment(workflowID int, fragmentID string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	fragmentID = strings.TrimSpace(fragmentID)
	if fragmentID == `` {
		return errors.New(`纯文本需求片段id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`plain_text_requirement_fragment_id`: fragmentID,
		`update_time`:                        time.Now().Unix(),
	}).Exec()
	return err
}

// TaskWorkflowBindDesignPlanReqFragment 绑定需求设计方案片段 id。
func (h *CSqlite) TaskWorkflowBindDesignPlanReqFragment(workflowID int, fragmentID string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	fragmentID = strings.TrimSpace(fragmentID)
	if fragmentID == `` {
		return errors.New(`需求设计方案片段id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`design_plan_requirement_fragment_id`: fragmentID,
		`update_time`:                         time.Now().Unix(),
	}).Exec()
	return err
}

// TaskWorkflowCreateRun 创建一条工作流测试/计划运行记录。
func (h *CSqlite) TaskWorkflowCreateRun(workflowID int, runType, triggerSource, requirementSnapshotMD, devPlanSnapshotMD string, coverageReport, testPlan, testReport map[string]any, summaryMD string) (map[string]any, error) {
	if workflowID <= 0 {
		return nil, errors.New(`工作流id不能为空`)
	}
	runType = strings.TrimSpace(runType)
	if runType == `` {
		return nil, errors.New(`运行类型不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return nil, err
	}
	now := time.Now().Unix()
	runNo := `tw-` + cast.ToString(workflowID) + `-` + cast.ToString(now)
	newID, err := h.Client.QuickCreate(`tbl_task_test_run`, map[string]any{
		`workflow_id`:             workflowID,
		`run_no`:                  runNo,
		`run_type`:                runType,
		`status`:                  `success`,
		`trigger_source`:          strings.TrimSpace(triggerSource),
		`requirement_snapshot_md`: requirementSnapshotMD,
		`dev_plan_snapshot_md`:    devPlanSnapshotMD,
		`diff_snapshot_text`:      ``,
		`coverage_report_json`:    gstool.JsonEncode(coverageReport),
		`test_plan_json`:          gstool.JsonEncode(testPlan),
		`test_report_json`:        gstool.JsonEncode(testReport),
		`summary_md`:              summaryMD,
		`started_at`:              now,
		`finished_at`:             now,
		`create_time`:             now,
	}).Exec()
	if err != nil {
		return nil, err
	}
	updateData := map[string]any{
		`status`:        `test_plan_ready`,
		`current_stage`: runType,
		`update_time`:   now,
	}
	if runType == `test_plan_generate` {
		updateData[`latest_plan_run_id`] = cast.ToInt(newID)
	}
	if runType == `api_test_execute` {
		updateData[`status`] = `test_run_ready`
	}
	updateData[`latest_test_run_id`] = cast.ToInt(newID)
	_, err = h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, updateData).Exec()
	if err != nil {
		return nil, err
	}
	return h.TaskWorkflowRunInfo(cast.ToInt(newID))
}

// TaskWorkflowRunInfo 查询单条运行记录。
func (h *CSqlite) TaskWorkflowRunInfo(id int) (map[string]any, error) {
	if id <= 0 {
		return nil, errors.New(`运行记录id不能为空`)
	}
	info, err := h.Client.QuickQuery(`tbl_task_test_run`, `*`, map[string]any{
		`id`: id,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, errors.New(`运行记录不存在`)
	}
	return info, nil
}

// TaskWorkflowLatestRunByType 查询某类最新运行记录。
func (h *CSqlite) TaskWorkflowLatestRunByType(workflowID int, runType string) (map[string]any, error) {
	if workflowID <= 0 {
		return nil, errors.New(`工作流id不能为空`)
	}
	runType = strings.TrimSpace(runType)
	if runType == `` {
		return nil, errors.New(`运行类型不能为空`)
	}
	list, err := h.Client.QueryBySql(`select * from tbl_task_test_run where workflow_id = ? and run_type = ? order by id desc limit 1`, workflowID, runType).All()
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return map[string]any{}, nil
	}
	return list[0], nil
}

// TaskWorkflowRunList 查询工作流运行历史。
func (h *CSqlite) TaskWorkflowRunList(workflowID int) ([]map[string]any, error) {
	if workflowID <= 0 {
		return nil, errors.New(`工作流id不能为空`)
	}
	return h.Client.QueryBySql(`select * from tbl_task_test_run where workflow_id = ? order by id desc`, workflowID).All()
}

// TaskWorkflowUpdatePrompts 更新工作流的提示词。
func (h *CSqlite) TaskWorkflowUpdatePrompts(workflowID int, promptRequirement, promptApiDev, promptApiTest, promptDesign, promptPlainTextRequirement, promptDesignPlanRequirement, promptBrowserTest, promptCodeReview string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`prompt_requirement`:             promptRequirement,
		`prompt_api_dev`:                 promptApiDev,
		`prompt_api_test`:                promptApiTest,
		`prompt_design`:                  promptDesign,
		`prompt_plain_text_requirement`:  promptPlainTextRequirement,
		`prompt_design_plan_requirement`: promptDesignPlanRequirement,
		`prompt_browser_test`:            promptBrowserTest,
		`prompt_code_review`:             promptCodeReview,
		`update_time`:                    time.Now().Unix(),
	}).Exec()
	return err
}

// TaskWorkflowMarkRequirementFetchRunning 标记需求抓取节点开始执行。
func (h *CSqlite) TaskWorkflowMarkRequirementFetchRunning(workflowID int, sourceURL string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	now := time.Now().Unix()
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`status`:                        taskWorkflowStageRequirementFetch,
		`current_stage`:                 taskWorkflowStageRequirementFetch,
		`requirement_fetch_status`:      taskWorkflowRequirementFetchRunning,
		`requirement_fetch_started_at`:  now,
		`requirement_fetch_finished_at`: 0,
		`requirement_fetch_error`:       ``,
		`requirement_source_url`:        strings.TrimSpace(sourceURL),
		`last_error`:                    ``,
		`update_time`:                   now,
	}).Exec()
	return err
}

// TaskWorkflowMarkRequirementFetchSuccess 标记需求抓取节点成功完成。
func (h *CSqlite) TaskWorkflowMarkRequirementFetchSuccess(workflowID int, sourceURL string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	now := time.Now().Unix()
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`status`:                        taskWorkflowStatusInit,
		`current_stage`:                 taskWorkflowStageRequirementFetch,
		`requirement_fetch_status`:      taskWorkflowRequirementFetchSuccess,
		`requirement_fetch_finished_at`: now,
		`requirement_fetch_error`:       ``,
		`requirement_source_url`:        strings.TrimSpace(sourceURL),
		`last_error`:                    ``,
		`update_time`:                   now,
	}).Exec()
	return err
}

// TaskWorkflowMarkRequirementFetchFailed 标记需求抓取节点执行失败。
func (h *CSqlite) TaskWorkflowMarkRequirementFetchFailed(workflowID int, sourceURL, errMsg string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	now := time.Now().Unix()
	errMsg = strings.TrimSpace(errMsg)
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`status`:                        taskWorkflowStatusInit,
		`current_stage`:                 taskWorkflowStageRequirementFetch,
		`requirement_fetch_status`:      taskWorkflowRequirementFetchFailed,
		`requirement_fetch_finished_at`: now,
		`requirement_fetch_error`:       errMsg,
		`requirement_source_url`:        strings.TrimSpace(sourceURL),
		`last_error`:                    errMsg,
		`update_time`:                   now,
	}).Exec()
	return err
}

// TaskWorkflowUpdateNodeStatuses 更新工作流节点状态。
func (h *CSqlite) TaskWorkflowUpdateNodeStatuses(workflowID int, nodeStatuses string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`node_statuses`: strings.TrimSpace(nodeStatuses),
		`update_time`:   time.Now().Unix(),
	}).Exec()
	return err
}

// TaskWorkflowBatchNodeStatusesByHomeTaskIDs 根据 home_task_id 列表批量查询工作流 node_statuses。
func (h *CSqlite) TaskWorkflowBatchNodeStatusesByHomeTaskIDs(homeTaskIDs []int) (map[int]string, error) {
	result := map[int]string{}
	if len(homeTaskIDs) == 0 {
		return result, nil
	}
	placeholders := make([]string, 0, len(homeTaskIDs))
	args := make([]any, 0, len(homeTaskIDs))
	for _, id := range homeTaskIDs {
		placeholders = append(placeholders, `?`)
		args = append(args, id)
	}
	list, err := h.Client.QueryBySql(`select home_task_id, node_statuses from tbl_task_workflow where home_task_id in (`+strings.Join(placeholders, `,`)+`)`, args...).All()
	if err != nil {
		return nil, err
	}
	for _, row := range list {
		result[cast.ToInt(row[`home_task_id`])] = strings.TrimSpace(cast.ToString(row[`node_statuses`]))
	}
	return result, nil
}

// TaskWorkflowBindApiDocFragment 绑定接口文档片段 id。
func (h *CSqlite) TaskWorkflowBindApiDocFragment(workflowID int, fragmentID string) error {
	if workflowID <= 0 {
		return errors.New(`工作流id不能为空`)
	}
	fragmentID = strings.TrimSpace(fragmentID)
	if fragmentID == `` {
		return errors.New(`接口文档片段id不能为空`)
	}
	if _, err := h.TaskWorkflowInfo(workflowID); err != nil {
		return err
	}
	_, err := h.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{
		`id`: workflowID,
	}, map[string]any{
		`api_doc_fragment_id`: fragmentID,
		`update_time`:         time.Now().Unix(),
	}).Exec()
	return err
}

// TaskWorkflowChatCreate 创建对话记录。
func (h *CSqlite) TaskWorkflowChatCreate(workflowID int, prompt, promptType, cliType string, agentCliID int, localDir, settingsPath, modelName string, thinkingCollapsed int, thinkingIntensity string) (int64, error) {
	if strings.TrimSpace(cliType) == `` {
		cliType = `claude`
	}
	now := time.Now().Format(`2006-01-02 15:04:05`)
	id, err := h.Client.QuickCreate(`tbl_task_workflow_chat`, map[string]any{
		`workflow_id`:        workflowID,
		`prompt`:             prompt,
		`prompt_type`:        promptType,
		`cli_type`:           cliType,
		`agent_cli_id`:       agentCliID,
		`local_dir`:          localDir,
		`settings_path`:      settingsPath,
		`model_name`:         modelName,
		`thinking_collapsed`: thinkingCollapsed,
		`thinking_intensity`: thinkingIntensity,
		`status`:             taskWorkflowChatStatusRunning,
		`raw_output`:         ``,
		`created_at`:         now,
		`updated_at`:         now,
	}).Exec()
	if err != nil {
		return 0, err
	}
	return id, nil
}

// TaskWorkflowChatUpdateSessionID 更新 session_id。
func (h *CSqlite) TaskWorkflowChatUpdateSessionID(chatID int64, sessionID string) error {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	_, err := h.Client.QuickUpdate(`tbl_task_workflow_chat`, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`session_id`: sessionID,
		`updated_at`: now,
	}).Exec()
	return err
}

const (
	// ChatOutputFlushBatchSize SSE 对话输出批量写 DB 的行数阈值
	ChatOutputFlushBatchSize = 200
)

// TaskWorkflowChatAppendOutput 追加一行 raw_output。
func (h *CSqlite) TaskWorkflowChatAppendOutput(chatID int64, line string) error {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	info, err := h.TaskWorkflowChatInfo(chatID)
	if err != nil {
		return err
	}
	current := cast.ToString(info[`raw_output`])
	newOutput := current
	if current != `` {
		newOutput += "\n"
	}
	newOutput += line
	_, err = h.Client.QuickUpdate(`tbl_task_workflow_chat`, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`raw_output`: newOutput,
		`updated_at`: now,
	}).Exec()
	return err
}

// TaskWorkflowChatAppendOutputBatch 批量追加多行 raw_output，一次 DB 读写完成。
func (h *CSqlite) TaskWorkflowChatAppendOutputBatch(chatID int64, lines []string) error {
	if len(lines) == 0 {
		return nil
	}
	now := time.Now().Format(`2006-01-02 15:04:05`)
	info, err := h.TaskWorkflowChatInfo(chatID)
	if err != nil {
		return err
	}
	current := cast.ToString(info[`raw_output`])
	newOutput := current
	if current != `` {
		newOutput += "\n"
	}
	newOutput += strings.Join(lines, "\n")
	_, err = h.Client.QuickUpdate(`tbl_task_workflow_chat`, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`raw_output`: newOutput,
		`updated_at`: now,
	}).Exec()
	return err
}

// TaskWorkflowChatMarkCompleted 标记对话完成。
func (h *CSqlite) TaskWorkflowChatMarkCompleted(chatID int64) error {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	_, err := h.Client.QuickUpdate(`tbl_task_workflow_chat`, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`status`:     taskWorkflowChatStatusCompleted,
		`updated_at`: now,
	}).Exec()
	return err
}

// TaskWorkflowChatMarkError 标记对话异常终止。
func (h *CSqlite) TaskWorkflowChatMarkError(chatID int64) error {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	_, err := h.Client.QuickUpdate(`tbl_task_workflow_chat`, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`status`:     taskWorkflowChatStatusError,
		`updated_at`: now,
	}).Exec()
	return err
}

// TaskWorkflowChatRecoverInterrupted 启动时将所有 running 状态的记录标记为 interrupted（进程已随上次进程退出而终止）。
func (h *CSqlite) TaskWorkflowChatRecoverInterrupted() {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	upNumber, err := h.Client.ExecBySql(
		`update tbl_task_workflow_chat set status = ?, updated_at = ? where status = ?`,
		taskWorkflowChatStatusInterrupted, now, taskWorkflowChatStatusRunning,
	).Exec()
	if err != nil {
		gstool.FmtPrintlnLogTime(`TaskWorkflowChatRecoverInterrupted 失败: %v`, err)
	} else {
		gstool.FmtPrintlnLogTime(`[agent cli] 更新状态进行中的为异常中断，数量%d`, upNumber)
	}
}

// TaskWorkflowChatInfo 获取单条对话记录。
func (h *CSqlite) TaskWorkflowChatInfo(chatID int64) (map[string]any, error) {
	return h.Client.QuickQuery(`tbl_task_workflow_chat`, `*`, map[string]any{
		`id`: chatID,
	}).One()
}

// TaskWorkflowChatList 获取 workflow 下所有对话记录。
func (h *CSqlite) TaskWorkflowChatList(workflowID int) ([]map[string]any, error) {
	rows, err := h.Client.QuickQuery(`tbl_task_workflow_chat`, `*`, map[string]any{
		`workflow_id`: workflowID,
	}).Order(`id DESC`).All()
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// TaskWorkflowChatMarkRunning 标记对话为运行中（用于继续对话）。
func (h *CSqlite) TaskWorkflowChatMarkRunning(chatID int64) error {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	_, err := h.Client.QuickUpdate(`tbl_task_workflow_chat`, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`status`:     taskWorkflowChatStatusRunning,
		`updated_at`: now,
	}).Exec()
	return err
}

// TaskWorkflowChatMarkInterrupted 标记对话为用户主动中断。
func (h *CSqlite) TaskWorkflowChatMarkInterrupted(chatID int64) error {
	now := time.Now().Format(`2006-01-02 15:04:05`)
	_, err := h.Client.QuickUpdate(`tbl_task_workflow_chat`, map[string]any{
		`id`: chatID,
	}, map[string]any{
		`status`:     taskWorkflowChatStatusInterrupted,
		`updated_at`: now,
	}).Exec()
	return err
}

// TaskWorkflowChatListByPromptType 按提示词类型查询对话历史。
func (h *CSqlite) TaskWorkflowChatListByPromptType(workflowID int, promptType string) ([]map[string]any, error) {
	if workflowID <= 0 {
		return nil, errors.New(`工作流id不能为空`)
	}
	promptType = strings.TrimSpace(promptType)
	if promptType == `` {
		return nil, errors.New(`提示词类型不能为空`)
	}
	rows, err := h.Client.QuickQuery(`tbl_task_workflow_chat`, `*`, map[string]any{
		`workflow_id`: workflowID,
		`prompt_type`: promptType,
	}).Order(`id DESC`).All()
	if err != nil {
		return nil, err
	}
	return rows, nil
}

// TaskWorkflowClearChatSessionIDs 删除指定 prompt_type 的所有对话记录。
func (h *CSqlite) TaskWorkflowClearChatSessionIDs(workflowID int, promptType string) error {
	promptType = strings.TrimSpace(promptType)
	if promptType == `` {
		return nil
	}
	_, err := h.Client.QuickDelete(`tbl_task_workflow_chat`, map[string]any{
		`workflow_id`: workflowID,
		`prompt_type`: promptType,
	}).Exec()
	return err
}
