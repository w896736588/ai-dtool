package common

import (
	"dev_tool/internal/app/dtool/define"
	"errors"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

const (
	// homeTaskListQuerySQL 用于查询首页任务列表。
	homeTaskListQuerySQL = `
select id,name,task_status,memory_fragment_id,is_archived,start_time,last_operated_at,create_time,update_time
from tbl_home_task
where is_archived = ?
order by id desc`
	// homeTaskDateLayout 用于只展示年月日格式。
	homeTaskDateLayout = `Y-m-d`
	// homeTaskDateTimeLayout 用于展示完整时间。
	homeTaskDateTimeLayout = `Y-m-d H:i:s`
)

// HomeTaskList 查询首页任务列表。
func (h *CSqlite) HomeTaskList(isArchived int) ([]map[string]any, error) {
	if !isValidHomeTaskArchived(isArchived) {
		return nil, errors.New(`归档状态不合法`)
	}
	list, err := h.Client.QueryBySql(homeTaskListQuerySQL, isArchived).All()
	if err != nil {
		return nil, err
	}
	h.fillHomeTaskTimeDescList(list)
	return list, nil
}

// HomeTaskRow 查询单条首页任务。
func (h *CSqlite) HomeTaskRow(id int) (map[string]any, error) {
	if id <= 0 {
		return nil, errors.New(`任务id不能为空`)
	}
	task, err := h.Client.QuickQuery(`tbl_home_task`, `*`, map[string]any{
		`id`: id,
	}).One()
	if err != nil {
		return nil, err
	}
	if len(task) == 0 {
		return nil, errors.New(`任务不存在`)
	}
	h.fillHomeTaskTimeDesc(task)
	return task, nil
}

// HomeTaskSave 保存首页任务。
func (h *CSqlite) HomeTaskSave(id int, name, taskStatus string, startTime int64, memoryFragmentID string) (map[string]any, error) {
	now := time.Now().Unix()
	name = strings.TrimSpace(name)
	taskStatus = strings.TrimSpace(taskStatus)
	startTime = normalizeHomeTaskStartTime(startTime)

	// 任务名称为空时直接返回错误，避免写入无意义记录。
	if name == `` {
		return nil, errors.New(`任务名称不能为空`)
	}
	// 任务状态必须是预定义常量，避免前端误传污染数据。
	if !isValidHomeTaskStatus(taskStatus) {
		return nil, errors.New(`任务状态不合法`)
	}

	updateData := map[string]any{
		`name`:               name,
		`task_status`:        taskStatus,
		`memory_fragment_id`: memoryFragmentID,
		`start_time`:         startTime,
		`last_operated_at`:   now,
		`update_time`:        now,
	}
	if id <= 0 {
		updateData[`is_archived`] = define.HomeTaskArchivedNo
		updateData[`create_time`] = now
		newID, err := h.Client.QuickCreate(`tbl_home_task`, updateData).Exec()
		if err != nil {
			return nil, err
		}
		id = cast.ToInt(newID)
	} else {
		_, err := h.Client.QuickUpdate(`tbl_home_task`, map[string]any{
			`id`: id,
		}, updateData).Exec()
		if err != nil {
			return nil, err
		}
	}
	return h.HomeTaskRow(id)
}

// HomeTaskArchiveToggle 切换首页任务归档状态。
func (h *CSqlite) HomeTaskArchiveToggle(id, isArchived int) (map[string]any, error) {
	if id <= 0 {
		return nil, errors.New(`任务id不能为空`)
	}
	if !isValidHomeTaskArchived(isArchived) {
		return nil, errors.New(`归档状态不合法`)
	}
	now := time.Now().Unix()
	_, err := h.Client.QuickUpdate(`tbl_home_task`, map[string]any{
		`id`: id,
	}, map[string]any{
		`is_archived`:      isArchived,
		`last_operated_at`: now,
	}).Exec()
	if err != nil {
		return nil, err
	}
	return h.HomeTaskRow(id)
}

// HomeTaskStatusQuickUpdate 快捷切换首页任务状态。
func (h *CSqlite) HomeTaskStatusQuickUpdate(id int, taskStatus string) (map[string]any, error) {
	taskStatus = strings.TrimSpace(taskStatus)
	if id <= 0 {
		return nil, errors.New(`任务id不能为空`)
	}
	if !isValidHomeTaskStatus(taskStatus) {
		return nil, errors.New(`任务状态不合法`)
	}
	taskInfo, err := h.HomeTaskRow(id)
	if err != nil {
		return nil, err
	}

	now := time.Now().Unix()
	updateData := map[string]any{
		`task_status`:      taskStatus,
		`last_operated_at`: now,
		`update_time`:      now,
	}
	// 当任务首次切到开发中且尚未记录开始时间时，自动补齐开始时间。
	if taskStatus == define.HomeTaskStatusDeveloping && cast.ToInt64(taskInfo[`start_time`]) <= 0 {
		updateData[`start_time`] = now
	}
	_, err = h.Client.QuickUpdate(`tbl_home_task`, map[string]any{
		`id`: id,
	}, updateData).Exec()
	if err != nil {
		return nil, err
	}
	return h.HomeTaskRow(id)
}

// HomeTaskDelete 真正删除首页任务，删除后不可恢复。
func (h *CSqlite) HomeTaskDelete(id int) error {
	if id <= 0 {
		return errors.New(`任务id不能为空`)
	}
	// 先确认任务存在，避免删除不存在记录时前端误判为成功。
	if _, err := h.HomeTaskRow(id); err != nil {
		return err
	}
	_, err := h.Client.ExecBySql(`delete from tbl_home_task where id = ?`, id).Exec()
	if err != nil {
		return err
	}
	return nil
}

// fillHomeTaskTimeDescList 批量填充首页任务时间描述字段。
func (h *CSqlite) fillHomeTaskTimeDescList(list []map[string]any) {
	for _, item := range list {
		h.fillHomeTaskTimeDesc(item)
	}
}

// fillHomeTaskTimeDesc 填充首页任务时间描述字段。
func (h *CSqlite) fillHomeTaskTimeDesc(row map[string]any) {
	row[`start_time_desc`] = formatHomeTaskTime(cast.ToInt64(row[`start_time`]), homeTaskDateLayout)
	row[`last_operated_at_desc`] = formatHomeTaskTime(cast.ToInt64(row[`last_operated_at`]), homeTaskDateTimeLayout)
	row[`create_time_desc`] = formatHomeTaskTime(cast.ToInt64(row[`create_time`]), homeTaskDateTimeLayout)
	row[`update_time_desc`] = formatHomeTaskTime(cast.ToInt64(row[`update_time`]), homeTaskDateTimeLayout)
}

// formatHomeTaskTime 格式化首页任务时间。
func formatHomeTaskTime(unixTime int64, layout string) string {
	if unixTime <= 0 {
		return ``
	}
	return gstool.TimeUnixToString(time.Unix(unixTime, 0), layout)
}

// normalizeHomeTaskStartTime 统一把空开始日期补到当天零点，避免前端漏传时留下空值。
func normalizeHomeTaskStartTime(startTime int64) int64 {
	if startTime > 0 {
		return startTime
	}
	now := time.Now()
	return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
}

// isValidHomeTaskStatus 校验首页任务状态是否合法。
func isValidHomeTaskStatus(taskStatus string) bool {
	for _, item := range define.HomeTaskStatusList {
		if item == taskStatus {
			return true
		}
	}
	return false
}

// isValidHomeTaskArchived 校验首页任务归档值是否合法。
func isValidHomeTaskArchived(isArchived int) bool {
	return isArchived == define.HomeTaskArchivedNo || isArchived == define.HomeTaskArchivedYes
}
