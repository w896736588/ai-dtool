package common

import (
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/spf13/cast"
)

const (
	// homeTaskListQuerySQL 用于查询首页任务列表。
	homeTaskListQuerySQL = `
select id,name,task_status,memory_fragment_id,is_archived,start_time,last_operated_at,create_time,update_time,tapd_url,git_id,api_dev_enabled,api_collection_id,api_dir_id,git_ids,api_dev_entries,mysql_id,dev_configs,use_workflow
from tbl_home_task
where is_archived = ?
order by id desc`
	// homeTaskListAllQuerySQL 用于查询所有任务（包含已归档和未归档）。
	homeTaskListAllQuerySQL = `
select id,name,task_status,memory_fragment_id,is_archived,start_time,last_operated_at,create_time,update_time,tapd_url,git_id,api_dev_enabled,api_collection_id,api_dir_id,git_ids,api_dev_entries,mysql_id,dev_configs,use_workflow
from tbl_home_task
order by id desc`
	// homeTaskListTodayUpdatedQuerySQL 用于查询今天变更过的任务，供工作日报使用。
	homeTaskListTodayUpdatedQuerySQL = `
select id,name,task_status,memory_fragment_id,is_archived,start_time,last_operated_at,create_time,update_time,tapd_url,git_id,api_dev_enabled,api_collection_id,api_dir_id,git_ids,api_dev_entries,mysql_id,dev_configs,use_workflow
from tbl_home_task
where update_time >= ? or create_time >= ?
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

// HomeTaskListTodayUpdated 查询今天变更过的任务列表（包含已归档和未归档），用于工作日报。
func (h *CSqlite) HomeTaskListTodayUpdated() ([]map[string]any, error) {
	now := time.Now()
	todayStart := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Unix()
	list, err := h.Client.QueryBySql(homeTaskListTodayUpdatedQuerySQL, todayStart, todayStart).All()
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
func (h *CSqlite) HomeTaskSave(id int, name, taskStatus string, startTime int64, memoryFragmentID string, tapdUrl string, gitID int, apiDevEnabled int, apiCollectionID int, apiDirID int, mysqlID int, gitIDsJSON string, apiDevEntriesJSON string, devConfigsJSON string, useWorkflow int) (map[string]any, error) {
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
		`tapd_url`:           tapdUrl,
		`git_id`:             gitID,
		`api_dev_enabled`:    apiDevEnabled,
		`api_collection_id`:  apiCollectionID,
		`api_dir_id`:         apiDirID,
		`mysql_id`:           mysqlID,
		`git_ids`:            gitIDsJSON,
		`api_dev_entries`:    apiDevEntriesJSON,
		`dev_configs`:        devConfigsJSON,
		`use_workflow`:       useWorkflow,
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

// ReplaceHomeTaskMemoryFragmentIDs 批量替换首页任务关联的知识片段 ID。 // Bulk replace memory fragment IDs referenced by home tasks.
func (h *CSqlite) ReplaceHomeTaskMemoryFragmentIDs(idMap map[string]string) error {
	if len(idMap) == 0 {
		return nil
	}
	for oldID, newID := range idMap {
		oldID = strings.TrimSpace(oldID)
		newID = strings.TrimSpace(newID)
		// 空值或无变化映射直接跳过，避免把无效值写回主库。 // Skip empty or unchanged mappings to avoid persisting invalid values.
		if oldID == `` || newID == `` || oldID == newID {
			continue
		}
		if _, err := h.Client.ExecBySql(`
update tbl_home_task
set memory_fragment_id = ?
where memory_fragment_id = ?`, newID, oldID).Exec(); err != nil {
			return err
		}
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

// fragmentRefTypeWorkflow 表示引用来源为工作流程任务。
const fragmentRefTypeWorkflow = `workflow`

// fragmentReference 描述一个片段引用来源。
type fragmentReference struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
	Name string `json:"name"`
}

// HomeTaskFragmentReferences 批量查询知识片段被哪些工作流程任务引用。
func (h *CSqlite) HomeTaskFragmentReferences(fragmentIDs []string) (map[string][]fragmentReference, error) {
	result := make(map[string][]fragmentReference)
	if len(fragmentIDs) == 0 {
		return result, nil
	}
	// 构建 IN 子句的占位符和参数。
	placeholders := make([]string, 0, len(fragmentIDs))
	args := make([]any, 0, len(fragmentIDs))
	for _, id := range fragmentIDs {
		id = strings.TrimSpace(id)
		if id == `` {
			continue
		}
		placeholders = append(placeholders, `?`)
		args = append(args, id)
	}
	if len(placeholders) == 0 {
		return result, nil
	}
	phStr := strings.Join(placeholders, `,`)

	// 查 tbl_home_task.memory_fragment_id。
	homeTaskRows, err := h.Client.QueryBySql(
		`SELECT id, name, memory_fragment_id FROM tbl_home_task WHERE memory_fragment_id IN (`+phStr+`) AND memory_fragment_id != ''`,
		args...,
	).All()
	if err != nil {
		return nil, err
	}
	for _, row := range homeTaskRows {
		fid := strings.TrimSpace(cast.ToString(row[`memory_fragment_id`]))
		if fid == `` {
			continue
		}
		result[fid] = append(result[fid], fragmentReference{
			Type: fragmentRefTypeWorkflow,
			ID:   cast.ToInt(row[`id`]),
			Name: cast.ToString(row[`name`]),
		})
	}

	// 查 tbl_task_workflow 中所有 fragment_id 字段。
	workflowRows, err := h.Client.QueryBySql(
		`SELECT tw.id, tw.home_task_id, ht.name,
			tw.requirement_fragment_id, tw.dev_plan_fragment_id,
			tw.plain_text_requirement_fragment_id, tw.design_plan_requirement_fragment_id,
			tw.api_doc_fragment_id, tw.design_fragment_id
		 FROM tbl_task_workflow tw
		 LEFT JOIN tbl_home_task ht ON ht.id = tw.home_task_id
		 WHERE tw.requirement_fragment_id IN (`+phStr+`)
		    OR tw.dev_plan_fragment_id IN (`+phStr+`)
		    OR tw.plain_text_requirement_fragment_id IN (`+phStr+`)
		    OR tw.design_plan_requirement_fragment_id IN (`+phStr+`)
		    OR tw.api_doc_fragment_id IN (`+phStr+`)
		    OR tw.design_fragment_id IN (`+phStr+`)`,
		func() []any {
			combined := make([]any, 0, len(args)*6)
			for i := 0; i < 6; i++ {
				combined = append(combined, args...)
			}
			return combined
		}()...,
	).All()
	if err != nil {
		return nil, err
	}
	fragColumns := []string{`requirement_fragment_id`, `dev_plan_fragment_id`, `plain_text_requirement_fragment_id`, `design_plan_requirement_fragment_id`, `api_doc_fragment_id`, `design_fragment_id`}
	for _, row := range workflowRows {
		taskID := cast.ToInt(row[`home_task_id`])
		taskName := cast.ToString(row[`name`])
		if taskID <= 0 || taskName == `` {
			continue
		}
		for _, col := range fragColumns {
			fid := strings.TrimSpace(cast.ToString(row[col]))
			if fid == `` {
				continue
			}
			// 去重：同一任务对同一片段只记录一次。
			exists := false
			for _, ref := range result[fid] {
				if ref.Type == fragmentRefTypeWorkflow && ref.ID == taskID {
					exists = true
					break
				}
			}
			if !exists {
				result[fid] = append(result[fid], fragmentReference{
					Type: fragmentRefTypeWorkflow,
					ID:   taskID,
					Name: taskName,
				})
			}
		}
	}

	return result, nil
}

// HomeTaskLastDevConfigByGitId 根据 git_id 查找最近一个包含该 Git 仓库的任务，返回匹配的 dev_config。
func (h *CSqlite) HomeTaskLastDevConfigByGitId(gitID int) (map[string]any, error) {
	if gitID <= 0 {
		return nil, errors.New(`git_id不合法`)
	}
	// 查询所有任务（包含已归档），按 id 倒序，取最近匹配的一条。
	list, err := h.Client.QueryBySql(homeTaskListAllQuerySQL).All()
	if err != nil {
		return nil, err
	}
	for _, task := range list {
		devConfigsStr := cast.ToString(task[`dev_configs`])
		if devConfigsStr == `` || devConfigsStr == `[]` {
			continue
		}
		var configs []map[string]any
		if err := json.Unmarshal([]byte(devConfigsStr), &configs); err != nil {
			continue
		}
		for _, cfg := range configs {
			if cast.ToInt(cfg[`git_id`]) == gitID {
				return cfg, nil
			}
		}
	}
	return map[string]any{}, nil
}
