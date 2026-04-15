package common

import (
	"errors"
	"strings"
	"time"

	"github.com/spf13/cast"
)

const (
	// AsyncTaskStatusPending 表示任务已创建，等待后台执行。 // AsyncTaskStatusPending means the task has been created and is waiting to run.
	AsyncTaskStatusPending = `pending`
	// AsyncTaskStatusRunning 表示任务后台执行中。 // AsyncTaskStatusRunning means the task is currently running in background.
	AsyncTaskStatusRunning = `running`
	// AsyncTaskStatusAwaitConfirm 表示 AI 已完成，等待用户确认。 // AsyncTaskStatusAwaitConfirm means AI finished and user confirmation is required.
	AsyncTaskStatusAwaitConfirm = `await_confirm`
	// AsyncTaskStatusFailed 表示任务执行失败。 // AsyncTaskStatusFailed means the task failed.
	AsyncTaskStatusFailed = `failed`
	// AsyncTaskStatusConfirmed 表示用户已确认并落地结果。 // AsyncTaskStatusConfirmed means the user confirmed and applied the result.
	AsyncTaskStatusConfirmed = `confirmed`
	// AsyncTaskStatusRejected 表示用户放弃结果。 // AsyncTaskStatusRejected means the user discarded the result.
	AsyncTaskStatusRejected = `rejected`
)

var asyncTaskFinalStatusMap = map[string]struct{}{
	AsyncTaskStatusConfirmed: {},
	AsyncTaskStatusRejected:  {},
}

// AsyncTaskCreate 创建异步任务记录。 // AsyncTaskCreate creates a new async task record.
func (h *CSqlite) AsyncTaskCreate(taskType, title, sourceID string, requestPayload string) (map[string]any, error) {
	now := time.Now().Unix()
	taskType = strings.TrimSpace(taskType)
	title = strings.TrimSpace(title)
	sourceID = strings.TrimSpace(sourceID)
	if taskType == `` {
		return nil, errors.New(`任务类型不能为空`)
	}
	newID, err := h.Client.QuickCreate(`tbl_async_task`, map[string]any{
		`task_type`:       taskType,
		`task_status`:     AsyncTaskStatusPending,
		`title`:           title,
		`source_id`:       sourceID,
		`request_payload`: requestPayload,
		`create_time`:     now,
		`update_time`:     now,
	}).Exec()
	if err != nil {
		return nil, err
	}
	return h.AsyncTaskInfo(cast.ToInt(newID))
}

// AsyncTaskList 查询异步任务列表，默认按 id 倒序返回。 // AsyncTaskList returns async tasks ordered by newest first.
func (h *CSqlite) AsyncTaskList(limit int) ([]map[string]any, error) {
	query := h.Client.QueryBySql(`
select id,task_type,task_status,title,source_id,request_payload,result_payload,error_message,create_time,start_time,finish_time,update_time
from tbl_async_task
order by id desc`)
	if limit > 0 {
		query = query.Limit(limit)
	}
	return query.All()
}

// AsyncTaskInfo 查询单条异步任务。 // AsyncTaskInfo loads a single async task by id.
func (h *CSqlite) AsyncTaskInfo(id int) (map[string]any, error) {
	if id <= 0 {
		return nil, errors.New(`异步任务id不能为空`)
	}
	info, err := h.Client.QuickQuery(`tbl_async_task`, `*`, map[string]any{
		`id`: id,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(info) == 0 {
		return nil, errors.New(`异步任务不存在`)
	}
	return info, nil
}

// AsyncTaskLatestPendingByType 查询指定类型最新的 pending 异步任务。 // Load the newest pending async task for the given type.
func (h *CSqlite) AsyncTaskLatestPendingByType(taskType string) (map[string]any, error) {
	taskType = strings.TrimSpace(taskType)
	if taskType == `` {
		return nil, errors.New(`任务类型不能为空`)
	}
	list, err := h.Client.QueryBySql(`
select id,task_type,task_status,title,source_id,request_payload,result_payload,error_message,create_time,start_time,finish_time,update_time
from tbl_async_task
where task_type = ? and task_status = ?
order by id desc
limit 1`, taskType, AsyncTaskStatusPending).All()
	if err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return nil, nil
	}
	return list[0], nil
}

// AsyncTaskMarkRunning 标记任务进入运行中。 // AsyncTaskMarkRunning marks the task as running.
func (h *CSqlite) AsyncTaskMarkRunning(id int) error {
	return h.asyncTaskUpdateStatus(id, AsyncTaskStatusRunning, map[string]any{
		`start_time`: time.Now().Unix(),
	})
}

// AsyncTaskMarkAwaitConfirm 写入结果并标记等待确认。 // AsyncTaskMarkAwaitConfirm stores the result and marks the task await-confirm.
func (h *CSqlite) AsyncTaskMarkAwaitConfirm(id int, resultPayload string) error {
	return h.asyncTaskUpdateStatus(id, AsyncTaskStatusAwaitConfirm, map[string]any{
		`result_payload`: resultPayload,
		`finish_time`:    time.Now().Unix(),
		`error_message`:  ``,
	})
}

// AsyncTaskMarkFailed 写入错误并标记失败。 // AsyncTaskMarkFailed stores the error and marks the task as failed.
func (h *CSqlite) AsyncTaskMarkFailed(id int, errMsg string) error {
	return h.asyncTaskUpdateStatus(id, AsyncTaskStatusFailed, map[string]any{
		`error_message`: strings.TrimSpace(errMsg),
		`finish_time`:   time.Now().Unix(),
	})
}

// AsyncTaskMarkFinal 标记任务进入最终状态。 // AsyncTaskMarkFinal marks the task as a final user-decided status.
func (h *CSqlite) AsyncTaskMarkFinal(id int, status string) error {
	status = strings.TrimSpace(status)
	if _, ok := asyncTaskFinalStatusMap[status]; !ok {
		return errors.New(`异步任务最终状态不合法`)
	}
	return h.asyncTaskUpdateStatus(id, status, map[string]any{
		`finish_time`: time.Now().Unix(),
	})
}

// AsyncTaskUpdateRequestPayload 更新任务请求参数，供可恢复任务刷新调度信息。 // Update request payload so resumable tasks can refresh schedule metadata.
func (h *CSqlite) AsyncTaskUpdateRequestPayload(id int, requestPayload string) error {
	return h.asyncTaskUpdateStatus(id, ``, map[string]any{
		`request_payload`: requestPayload,
	})
}

// AsyncTaskDelete 删除异步任务记录。 // AsyncTaskDelete removes the async task record.
func (h *CSqlite) AsyncTaskDelete(id int) error {
	if id <= 0 {
		return errors.New(`异步任务id不能为空`)
	}
	// 先查一次，保证删除不存在记录时返回明确错误。 // Check first so deleting a missing record returns a clear error.
	if _, err := h.AsyncTaskInfo(id); err != nil {
		return err
	}
	_, err := h.Client.QuickDelete(`tbl_async_task`, map[string]any{
		`id`: id,
	}).Exec()
	return err
}

// AsyncTaskSummary 统计任务数量并返回最近列表。 // AsyncTaskSummary returns counts plus a recent task list.
func (h *CSqlite) AsyncTaskSummary(limit int) (map[string]any, error) {
	list, err := h.AsyncTaskList(limit)
	if err != nil {
		return nil, err
	}
	pendingCount, err := h.asyncTaskCountByStatus(AsyncTaskStatusPending)
	if err != nil {
		return nil, err
	}
	runningCount, err := h.asyncTaskCountByStatus(AsyncTaskStatusRunning)
	if err != nil {
		return nil, err
	}
	awaitConfirmCount, err := h.asyncTaskCountByStatus(AsyncTaskStatusAwaitConfirm)
	if err != nil {
		return nil, err
	}
	failedCount, err := h.asyncTaskCountByStatus(AsyncTaskStatusFailed)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		`list`:                list,
		`pending_count`:       pendingCount,
		`running_count`:       runningCount,
		`await_confirm_count`: awaitConfirmCount,
		`failed_count`:        failedCount,
		`total`:               len(list),
	}, nil
}

// asyncTaskUpdateStatus 统一更新状态和时间字段。 // asyncTaskUpdateStatus centralizes status and timestamp updates.
func (h *CSqlite) asyncTaskUpdateStatus(id int, status string, extra map[string]any) error {
	if id <= 0 {
		return errors.New(`异步任务id不能为空`)
	}
	// 必须先确认任务存在，避免 update 0 行时前端误判为成功。 // Ensure the task exists first so a zero-row update is not treated as success.
	if _, err := h.AsyncTaskInfo(id); err != nil {
		return err
	}
	updateData := map[string]any{
		`update_time`: time.Now().Unix(),
	}
	if strings.TrimSpace(status) != `` {
		updateData[`task_status`] = status
	}
	for key, value := range extra {
		updateData[key] = value
	}
	_, err := h.Client.QuickUpdate(`tbl_async_task`, map[string]any{
		`id`: id,
	}, updateData).Exec()
	return err
}

// asyncTaskCountByStatus 按状态统计任务数量。 // asyncTaskCountByStatus counts tasks by status.
func (h *CSqlite) asyncTaskCountByStatus(status string) (int, error) {
	countValue, err := h.Client.QueryBySql(`
select count(1) as total
from tbl_async_task
where task_status = ?`, status).Value(`total`)
	if err != nil {
		return 0, err
	}
	return cast.ToInt(countValue), nil
}
