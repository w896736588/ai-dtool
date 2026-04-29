package common

import (
	"errors"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

const (
	taskWorkflowStatusInit = `init`
	taskWorkflowStageIdle  = `idle`
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
		`home_task_id`:            homeTaskID,
		`status`:                  taskWorkflowStatusInit,
		`current_stage`:           taskWorkflowStageIdle,
		`requirement_fragment_id`: strings.TrimSpace(cast.ToString(homeTaskInfo[`memory_fragment_id`])),
		`dev_plan_fragment_id`:    ``,
		`latest_plan_run_id`:      0,
		`latest_test_run_id`:      0,
		`base_branch`:             ``,
		`feature_branch`:          ``,
		`last_error`:              ``,
		`create_time`:             now,
		`update_time`:             now,
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
