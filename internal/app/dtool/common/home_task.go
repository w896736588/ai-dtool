package common

import (
	"dev_tool/internal/app/dtool/define"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

const homeTaskNameMaxLength = 200

// HomeTaskNameMaxLength 返回任务名称最大长度。
func HomeTaskNameMaxLength() int {
	return homeTaskNameMaxLength
}

const (
	// homeTaskListQuerySQL 用于查询首页任务列表。
	homeTaskListQuerySQL = `
select id,name,task_status,memory_fragment_id,is_archived,start_time,last_operated_at,create_time,update_time,fetch_type,tapd_url,zentao_url,git_id,api_dev_enabled,api_collection_id,api_dir_id,git_ids,api_dev_entries,mysql_id,dev_configs,use_workflow
from tbl_home_task
where is_archived = ?
order by id desc`
	// homeTaskListAllQuerySQL 用于查询所有任务（包含已归档和未归档）。
	homeTaskListAllQuerySQL = `
select id,name,task_status,memory_fragment_id,is_archived,start_time,last_operated_at,create_time,update_time,fetch_type,tapd_url,zentao_url,git_id,api_dev_enabled,api_collection_id,api_dir_id,git_ids,api_dev_entries,mysql_id,dev_configs,use_workflow
from tbl_home_task
order by id desc`
	// homeTaskListPaginationQuerySQL 用于分页查询任务列表（LIMIT+OFFSET 由调用方追加）。
	homeTaskListPaginationQuerySQL = `
select id,name,task_status,memory_fragment_id,is_archived,start_time,last_operated_at,create_time,update_time,fetch_type,tapd_url,zentao_url,git_id,api_dev_enabled,api_collection_id,api_dir_id,git_ids,api_dev_entries,mysql_id,dev_configs,use_workflow
from tbl_home_task
where is_archived = ?
order by id desc
limit ? offset ?`
	// homeTaskCountSQL 用于统计各归档状态的任务数量。
	homeTaskCountSQL = `select is_archived, count(1) as cnt from tbl_home_task group by is_archived`
	// homeTaskListTodayUpdatedQuerySQL 用于查询今天变更过的任务，供工作日报使用。
	homeTaskListTodayUpdatedQuerySQL = `
select id,name,task_status,memory_fragment_id,is_archived,start_time,last_operated_at,create_time,update_time,fetch_type,tapd_url,zentao_url,git_id,api_dev_enabled,api_collection_id,api_dir_id,git_ids,api_dev_entries,mysql_id,dev_configs,use_workflow
from tbl_home_task
where update_time >= ? or create_time >= ?
order by id desc`
	// homeTaskDateLayout 用于只展示年月日格式。
	homeTaskDateLayout = `Y-m-d`
	// homeTaskDateTimeLayout 用于展示完整时间。
	homeTaskDateTimeLayout = `Y-m-d H:i:s`
)

// HomeTaskList 查询首页任务列表。
// isArchived: 0=未归档, 1=已归档, -1=全部。
func (h *CSqlite) HomeTaskList(isArchived int) ([]map[string]any, error) {
	if !isValidHomeTaskArchived(isArchived) {
		return nil, errors.New(`归档状态不合法`)
	}
	// 全量查询：不走 where 条件过滤，直接查所有。
	if isArchived == define.HomeTaskArchivedAll {
		return h.HomeTaskListAll()
	}
	list, err := h.Client.QueryBySql(homeTaskListQuerySQL, isArchived).All()
	if err != nil {
		return nil, err
	}
	h.fillHomeTaskTimeDescList(list)
	return list, nil
}

// HomeTaskListAll 查询所有首页任务（含已归档和未归档）。
func (h *CSqlite) HomeTaskListAll() ([]map[string]any, error) {
	list, err := h.Client.QueryBySql(homeTaskListAllQuerySQL).All()
	if err != nil {
		return nil, err
	}
	h.fillHomeTaskTimeDescList(list)
	return list, nil
}

// HomeTaskCount 返回活跃和归档任务的数量（activeCount=未归档, archivedCount=已归档）。
func (h *CSqlite) HomeTaskCount() (activeCount int, archivedCount int, err error) {
	rows, err := h.Client.QueryBySql(homeTaskCountSQL).All()
	if err != nil {
		return 0, 0, err
	}
	for _, row := range rows {
		isArchived := cast.ToInt(row[`is_archived`])
		cnt := cast.ToInt(row[`cnt`])
		if isArchived == define.HomeTaskArchivedNo {
			activeCount = cnt
		} else if isArchived == define.HomeTaskArchivedYes {
			archivedCount = cnt
		}
	}
	return activeCount, archivedCount, nil
}

// HomeTaskListPaginated 分页查询首页任务列表，返回任务列表和总数。
func (h *CSqlite) HomeTaskListPaginated(isArchived, page, pageSize int) (list []map[string]any, total int, err error) {
	if !isValidHomeTaskArchived(isArchived) {
		return nil, 0, errors.New(`归档状态不合法`)
	}
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 20
	}
	// 先查总数
	countRows, err := h.Client.QueryBySql(homeTaskCountSQL).All()
	if err != nil {
		return nil, 0, err
	}
	for _, row := range countRows {
		if cast.ToInt(row[`is_archived`]) == isArchived {
			total = cast.ToInt(row[`cnt`])
			break
		}
	}
	// 再查分页数据
	offset := (page - 1) * pageSize
	list, err = h.Client.QueryBySql(homeTaskListPaginationQuerySQL, isArchived, pageSize, offset).All()
	if err != nil {
		return nil, 0, err
	}
	h.fillHomeTaskTimeDescList(list)
	return list, total, nil
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
func (h *CSqlite) HomeTaskSave(id int, name, taskStatus string, startTime int64, memoryFragmentID string, fetchType string, tapdUrl string, zentaoUrl string, gitID int, apiDevEnabled int, apiCollectionID int, apiDirID int, mysqlID int, gitIDsJSON string, apiDevEntriesJSON string, devConfigsJSON string, useWorkflow int, workflowTemplateID int) (map[string]any, error) {
	now := time.Now().Unix()
	name = strings.TrimSpace(name)
	taskStatus = strings.TrimSpace(taskStatus)
	fetchType = strings.TrimSpace(strings.ToLower(fetchType))
	startTime = normalizeHomeTaskStartTime(startTime)
	if fetchType == `` {
		fetchType = `tapd`
	}

	// 任务名称为空时直接返回错误，避免写入无意义记录。
	if name == `` {
		return nil, errors.New(`任务名称不能为空`)
	}
	if utf8.RuneCountInString(name) > homeTaskNameMaxLength {
		return nil, errors.New(fmt.Sprintf(`任务名称不能超过%d字`, homeTaskNameMaxLength))
	}
	// 任务状态必须是预定义常量，避免前端误传污染数据。
	if !isValidHomeTaskStatus(taskStatus) {
		return nil, errors.New(`任务状态不合法`)
	}

	updateData := map[string]any{
		`name`:                 name,
		`task_status`:          taskStatus,
		`memory_fragment_id`:   memoryFragmentID,
		`start_time`:           startTime,
		`last_operated_at`:     now,
		`update_time`:          now,
		`fetch_type`:           fetchType,
		`tapd_url`:             tapdUrl,
		`zentao_url`:           zentaoUrl,
		`git_id`:               gitID,
		`api_dev_enabled`:      apiDevEnabled,
		`api_collection_id`:    apiCollectionID,
		`api_dir_id`:           apiDirID,
		`mysql_id`:             mysqlID,
		`git_ids`:              gitIDsJSON,
		`api_dev_entries`:      apiDevEntriesJSON,
		`dev_configs`:          devConfigsJSON,
		`use_workflow`:         useWorkflow,
		`workflow_template_id`: workflowTemplateID,
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

// isValidHomeTaskStatus 校验首页任务状态是否合法（从数据库动态读取）。
func isValidHomeTaskStatus(taskStatus string) bool {
	names, err := DbMain.TaskStatusNames()
	if err != nil {
		return false
	}
	for _, item := range names {
		if item == taskStatus {
			return true
		}
	}
	return false
}

// isValidHomeTaskArchived 校验首页任务归档值是否合法（含全量查询模式）。
func isValidHomeTaskArchived(isArchived int) bool {
	return isArchived == define.HomeTaskArchivedNo || isArchived == define.HomeTaskArchivedYes || isArchived == define.HomeTaskArchivedAll
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

	targetIDMap := make(map[string]struct{}, len(fragmentIDs))
	for _, id := range fragmentIDs {
		id = strings.TrimSpace(id)
		if id != `` {
			targetIDMap[id] = struct{}{}
		}
	}

	// 查 tbl_task_workflow_document 中所有文档的片段引用。
	docRows, err := h.Client.QueryBySql(
		`SELECT twd.workflow_id, tw.home_task_id, ht.name, twd.file_id
		 FROM tbl_task_workflow_document twd
		 JOIN tbl_task_workflow tw ON tw.id = twd.workflow_id
		 LEFT JOIN tbl_home_task ht ON ht.id = tw.home_task_id
		 WHERE twd.file_id != ''`,
	).All()
	if err != nil {
		return nil, err
	}
	for _, row := range docRows {
		taskID := cast.ToInt(row[`home_task_id`])
		taskName := cast.ToString(row[`name`])
		fileID := strings.TrimSpace(cast.ToString(row[`file_id`]))
		if taskID <= 0 || taskName == `` || fileID == `` {
			continue
		}
		if _, ok := targetIDMap[fileID]; !ok {
			continue
		}
		// 去重：同一任务对同一片段只记录一次。
		exists := false
		for _, item := range result[fileID] {
			if item.Type == fragmentRefTypeWorkflow && item.ID == taskID {
				exists = true
				break
			}
		}
		if !exists {
			result[fileID] = append(result[fileID], fragmentReference{
				Type: fragmentRefTypeWorkflow,
				ID:   taskID,
				Name: taskName,
			})
		}
	}

	return result, nil
}

// TaskStatusList 查询所有任务状态，按 sort_order 升序排列。
func (h *CSqlite) TaskStatusList() ([]map[string]any, error) {
	return h.Client.QuickQuery(`tbl_task_status`, `*`, nil).Order(`sort_order asc, id asc`).All()
}

// TaskStatusSave 新增或编辑任务状态。
func (h *CSqlite) TaskStatusSave(id int, name string, sortOrder int) ([]map[string]any, error) {
	name = strings.TrimSpace(name)
	if name == `` {
		return nil, errors.New(`状态名称不能为空`)
	}
	now := time.Now().Unix()
	if id > 0 {
		// 编辑
		_, err := h.Client.QuickUpdate(`tbl_task_status`, map[string]any{
			`id`: id,
		}, map[string]any{
			`name`:        name,
			`sort_order`:  sortOrder,
			`update_time`: now,
		}).Exec()
		if err != nil {
			return nil, err
		}
	} else {
		// 新增
		_, err := h.Client.QuickCreate(`tbl_task_status`, map[string]any{
			`name`:        name,
			`sort_order`:  sortOrder,
			`create_time`: now,
			`update_time`: now,
		}).Exec()
		if err != nil {
			return nil, err
		}
	}
	return h.TaskStatusList()
}

// TaskStatusDelete 删除任务状态。
func (h *CSqlite) TaskStatusDelete(id int) error {
	if id <= 0 {
		return errors.New(`状态id不能为空`)
	}
	_, err := h.Client.QuickDelete(`tbl_task_status`, map[string]any{
		`id`: id,
	}).Exec()
	return err
}

// TaskStatusSort 批量更新任务状态排序。
func (h *CSqlite) TaskStatusSort(ids []int) error {
	if len(ids) == 0 {
		return errors.New(`排序列表不能为空`)
	}
	now := time.Now().Unix()
	for i, id := range ids {
		_, err := h.Client.QuickUpdate(`tbl_task_status`, map[string]any{
			`id`: id,
		}, map[string]any{
			`sort_order`:  i,
			`update_time`: now,
		}).Exec()
		if err != nil {
			return err
		}
	}
	return nil
}

// TaskStatusNames 获取所有任务状态名称列表（用于校验）。
func (h *CSqlite) TaskStatusNames() ([]string, error) {
	list, err := h.TaskStatusList()
	if err != nil {
		return nil, err
	}
	names := make([]string, 0, len(list))
	for _, item := range list {
		names = append(names, cast.ToString(item[`name`]))
	}
	return names, nil
}

// HomeTaskContainsFragmentID 检查指定知识片段是否属于指定任务。
// 检查范围包括：tbl_home_task.memory_fragment_id 以及 tbl_task_workflow 中所有 fragment_id 列。
func (h *CSqlite) HomeTaskContainsFragmentID(homeTaskID int, fragmentID string) (bool, error) {
	if homeTaskID <= 0 || strings.TrimSpace(fragmentID) == `` {
		return false, nil
	}
	fragmentID = strings.TrimSpace(fragmentID)

	// 1. 检查 tbl_home_task.memory_fragment_id。
	homeTask, err := h.Client.QuickQuery(`tbl_home_task`, `memory_fragment_id`, map[string]any{
		`id`: homeTaskID,
	}).One()
	if err == nil && homeTask != nil {
		if strings.TrimSpace(cast.ToString(homeTask[`memory_fragment_id`])) == fragmentID {
			return true, nil
		}
	}

	// 2. 检查 tbl_task_workflow_document 中是否有该片段。
	wf, wfErr := h.Client.QuickQuery(`tbl_task_workflow`, `id`, map[string]any{
		`home_task_id`: homeTaskID,
	}).One()
	if wfErr == nil && wf != nil {
		workflowID := cast.ToInt(wf[`id`])
		if workflowID > 0 {
			if contains, containsErr := h.TaskWorkflowContainsFragmentID(workflowID, fragmentID); containsErr == nil && contains {
				return true, nil
			}
		}
	}

	return false, nil
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

// HomeTaskZcodeSessionIdAppend 向任务追加一个 zcode 对话 sessionId（末尾去重）。
func (h *CSqlite) HomeTaskZcodeSessionIdAppend(id int, sessionID string) error {
	row, err := h.HomeTaskRow(id)
	if err != nil {
		return err
	}
	if row == nil {
		return errors.New(`任务不存在`)
	}
	existing := cast.ToString(row[`zcode_session_ids`])
	lines := strings.Split(existing, "\n")
	// 取最后一行，若与新增 sessionId 相同则跳过
	if len(lines) > 0 {
		last := strings.TrimSpace(lines[len(lines)-1])
		if last == strings.TrimSpace(sessionID) {
			return nil
		}
	}
	newVal := sessionID
	if existing != `` {
		newVal = existing + "\n" + sessionID
	}
	_, err = h.Client.QuickUpdate(`tbl_home_task`, map[string]any{
		`id`: id,
	}, map[string]any{
		`zcode_session_ids`: newVal,
	}).Exec()
	return err
}

const homeTaskUnusedDirLimit = 50

// HomeTaskUnusedLocalDirs 查询最近50个历史任务中未被活跃任务占用的本地目录。
func (h *CSqlite) HomeTaskUnusedLocalDirs(excludeTaskID int) ([]string, error) {
	// 1. 取最近50条任务（按 id DESC），收集所有 dev_configs 中的 local_dir。
	list, err := h.Client.QueryBySql(homeTaskListAllQuerySQL).All()
	if err != nil {
		return nil, err
	}
	dirSet := make(map[string]bool)
	taskCount := 0
	for _, task := range list {
		if excludeTaskID > 0 && cast.ToInt(task[`id`]) == excludeTaskID {
			continue
		}
		devConfigsStr := cast.ToString(task[`dev_configs`])
		if devConfigsStr == `` || devConfigsStr == `[]` {
			continue
		}
		var configs []map[string]any
		if err := json.Unmarshal([]byte(devConfigsStr), &configs); err != nil {
			continue
		}
		for _, cfg := range configs {
			dir := strings.TrimSpace(cast.ToString(cfg[`local_dir`]))
			if dir != `` {
				dirSet[dir] = true
			}
		}
		taskCount++
		if taskCount >= homeTaskUnusedDirLimit {
			break
		}
	}

	// 2. 收集所有活跃任务已占用的 local_dir。
	activeList, err := h.HomeTaskList(define.HomeTaskArchivedNo)
	if err != nil {
		return nil, err
	}
	used := make(map[string]bool)
	for _, task := range activeList {
		devConfigsStr := cast.ToString(task[`dev_configs`])
		if devConfigsStr == `` || devConfigsStr == `[]` {
			continue
		}
		var configs []map[string]any
		if err := json.Unmarshal([]byte(devConfigsStr), &configs); err != nil {
			continue
		}
		for _, cfg := range configs {
			dir := strings.TrimSpace(cast.ToString(cfg[`local_dir`]))
			if dir != `` {
				used[dir] = true
			}
		}
	}

	// 3. 过滤掉已被活跃任务占用的目录。
	result := make([]string, 0)
	for dir := range dirSet {
		if !used[dir] {
			result = append(result, dir)
		}
	}
	return result, nil
}
