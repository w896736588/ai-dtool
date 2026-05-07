package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	_struct "dev_tool/internal/app/dtool/struct"
	"encoding/json"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

const (
	homeTaskApiDevDisabled = 0
	homeTaskApiDevEnabled  = 1
)

// HomeTaskList 查询首页任务列表。
func HomeTaskList(c *gin.Context) {
	request := _struct.HomeTaskListRequest{}
	_ = gsgin.GinPostBody(c, &request)
	list, err := common.DbMain.HomeTaskList(request.IsArchived)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	enrichHomeTaskListWithMemoryFragment(list)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`task_list`: list,
	})
}

// HomeTaskSave 保存首页任务。
func HomeTaskSave(c *gin.Context) {
	request := _struct.HomeTaskSaveRequest{}
	_ = gsgin.GinPostBody(c, &request)
	memoryFragmentID, err := ensureHomeTaskMemoryFragment(
		request.ID, request.Name, normalizeHomeTaskMemoryFragmentID(request.MemoryFragmentID),
		request.TapdUrl, request.ApiHost, request.ApiToken,
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	// 解析 dev_configs JSON，从中派生旧字段实现向后兼容。
	devConfigsJSON := resolveHomeTaskDevConfigsJSON(&request)
	var devConfigs []_struct.DevConfig
	_ = json.Unmarshal([]byte(devConfigsJSON), &devConfigs)

	// 从 dev_configs 派生 git_ids。
	gitIDsJSON := resolveHomeTaskGitIDsJSON(&request)
	if len(devConfigs) > 0 {
		var gitIDs []int
		for _, cfg := range devConfigs {
			if cfg.GitID > 0 {
				gitIDs = append(gitIDs, cfg.GitID)
			}
		}
		if len(gitIDs) > 0 {
			gitIDsBytes, _ := json.Marshal(gitIDs)
			gitIDsJSON = string(gitIDsBytes)
		}
	}

	// 从 dev_configs 派生 api_dev_entries。
	apiDevEntriesJSON := `[]`
	var apiEntries []_struct.ApiDevEntry
	for _, cfg := range devConfigs {
		if cfg.CollectionID > 0 {
			apiEntries = append(apiEntries, _struct.ApiDevEntry{CollectionID: cfg.CollectionID, DirID: cfg.DirID})
		}
	}
	if len(apiEntries) > 0 {
		entriesBytes, _ := json.Marshal(apiEntries)
		apiDevEntriesJSON = string(entriesBytes)
	} else {
		apiDevEntriesJSON = resolveHomeTaskApiDevEntriesJSON(&request)
	}

	// 自动创建文件夹：遍历 dev_configs 中每个 dir_id=0 且 collection_id>0 的条目。
	devConfigsJSON = autoCreateHomeTaskDevConfigDirs(request.Name, devConfigsJSON)
	_ = json.Unmarshal([]byte(devConfigsJSON), &devConfigs)

	apiDevEnabled := request.ApiDevEnabled
	apiCollectionID := request.ApiCollectionID
	apiDirID := request.ApiDirID
	if len(apiEntries) > 0 {
		apiDevEnabled = homeTaskApiDevEnabled
		apiCollectionID = apiEntries[0].CollectionID
		apiDirID = apiEntries[0].DirID
	}
	// 重新从 dev_configs 读取（dir_id 可能已被自动创建更新）。
	if len(devConfigs) > 0 {
		for i := range apiEntries {
			for _, cfg := range devConfigs {
				if cfg.CollectionID == apiEntries[i].CollectionID {
					apiEntries[i].DirID = cfg.DirID
				}
			}
		}
		entriesBytes, _ := json.Marshal(apiEntries)
		apiDevEntriesJSON = string(entriesBytes)
		apiDirID = devConfigs[0].DirID
	}

	// 兼容旧字段：从 git_ids 回填 git_id。
	var gitIDs []int
	if json.Unmarshal([]byte(gitIDsJSON), &gitIDs) == nil && len(gitIDs) > 0 {
		request.GitID = gitIDs[0]
	}

	// 兼容旧字段：从 dev_configs 回填 mysql_id。
	if len(devConfigs) > 0 && devConfigs[0].MysqlID > 0 {
		request.MysqlID = devConfigs[0].MysqlID
	}

	info, err := common.DbMain.HomeTaskSave(request.ID, request.Name, request.TaskStatus, request.StartTime, memoryFragmentID, request.TapdUrl, request.GitID, apiDevEnabled, apiCollectionID, apiDirID, request.MysqlID, gitIDsJSON, apiDevEntriesJSON, devConfigsJSON)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if request.ID <= 0 {
		_, _ = common.DbMain.TaskWorkflowCreateOrGetByHomeTaskID(cast.ToInt(info[`id`]))
	}
	enrichHomeTaskListWithMemoryFragment([]map[string]any{info})

	gsgin.GinResponseSuccess(c, ``, info)
}

// HomeTaskInfo 查询单条首页任务详情。
func HomeTaskInfo(c *gin.Context) {
	request := _struct.HomeTaskDeleteRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := common.DbMain.HomeTaskRow(request.ID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	enrichHomeTaskListWithMemoryFragment([]map[string]any{info})
	gsgin.GinResponseSuccess(c, ``, info)
}

// HomeTaskArchiveToggle 切换首页任务归档状态。
func HomeTaskArchiveToggle(c *gin.Context) {
	request := _struct.HomeTaskArchiveToggleRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := common.DbMain.HomeTaskArchiveToggle(request.ID, request.IsArchived)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	enrichHomeTaskListWithMemoryFragment([]map[string]any{info})
	gsgin.GinResponseSuccess(c, ``, info)
}

// HomeTaskStatusQuickUpdate 快捷切换首页任务状态。
func HomeTaskStatusQuickUpdate(c *gin.Context) {
	request := _struct.HomeTaskStatusQuickUpdateRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := common.DbMain.HomeTaskStatusQuickUpdate(request.ID, request.TaskStatus)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	enrichHomeTaskListWithMemoryFragment([]map[string]any{info})
	gsgin.GinResponseSuccess(c, ``, info)
}

// HomeTaskDelete 删除首页任务。
func HomeTaskDelete(c *gin.Context) {
	request := _struct.HomeTaskDeleteRequest{}
	_ = gsgin.GinPostBody(c, &request)
	err := common.DbMain.HomeTaskDelete(request.ID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// HomeTaskDailyReportGenerate 创建首页工作日报异步任务。 // HomeTaskDailyReportGenerate creates an async home-task daily report task.

// HomeTaskLastDevConfigByGitId 根据 Git 仓库 ID 查找最近一个使用该仓库的任务的开发配置。
func HomeTaskLastDevConfigByGitId(c *gin.Context) {
	request := _struct.HomeTaskLastDevConfigByGitIdRequest{}
	_ = gsgin.GinPostBody(c, &request)
	cfg, err := common.DbMain.HomeTaskLastDevConfigByGitId(request.GitID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, cfg)
}
func HomeTaskDailyReportGenerate(c *gin.Context) {
	if _, ok := memoryDBOrResponse(c); !ok {
		return
	}
	taskList, err := common.DbMain.HomeTaskListTodayUpdated()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	reportTime := time.Now().Unix()
	if _, err = buildHomeTaskDailyReportTasksSnapshot(taskList); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	taskInfo, err := createAsyncTask(
		asyncTaskTypeHomeTaskDailyReport,
		buildHomeTaskDailyReportTitle(time.Unix(reportTime, 0)),
		``,
		map[string]any{
			`report_time`: reportTime,
			`task_count`:  len(taskList),
		},
		func(taskID int) {
			runAsyncTaskAndPersistResult(taskID, func() (map[string]any, error) {
				return buildAsyncHomeTaskDailyReportResult(taskList, reportTime)
			})
		},
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`task_id`:     taskInfo[`id`],
		`task_status`: taskInfo[`task_status`],
		`task_type`:   taskInfo[`task_type`],
	})
}

func ensureHomeTaskMemoryFragment(taskID int, taskName string, memoryFragmentID string, tapdUrl string, apiHost string, apiToken string) (string, error) {
	taskName = strings.TrimSpace(taskName)
	if taskName == `` {
		return ``, gstool.Error(`任务名称不能为空`)
	}
	if component.MemoryRuntime == nil {
		return ``, common.ErrMemoryNotConfigured
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``, err
	}
	memoryDB := component.MemoryRuntime.DB()
	if memoryFragmentID != `` {
		if _, infoErr := memoryDB.MemoryFragmentInfo(memoryFragmentID); infoErr != nil {
			return ``, infoErr
		}
		return memoryFragmentID, nil
	}
	if !shouldAutoCreateHomeTaskMemoryFragment(taskID, memoryFragmentID) {
		return ``, nil
	}
	fragmentContent := buildHomeTaskFragmentContent(taskName, tapdUrl, apiHost, apiToken)
	fragmentInfo, saveErr := memoryDB.MemoryFragmentSave(0, taskName, fragmentContent, []string{`需求`})
	if saveErr != nil {
		return ``, saveErr
	}
	component.MemoryRuntime.ScheduleSync()
	fragmentID := strings.TrimSpace(cast.ToString(fragmentInfo[`file_id`]))
	if fragmentID == `` || fragmentID == `0` {
		return ``, gstool.Error(`自动创建知识片段失败`)
	}
	return fragmentID, nil
}

func shouldAutoCreateHomeTaskMemoryFragment(taskID int, memoryFragmentID string) bool {
	return taskID <= 0 && strings.TrimSpace(memoryFragmentID) == ``
}

func buildHomeTaskFragmentContent(taskName string, tapdUrl string, apiHost string, apiToken string) string {
	return "# " + taskName + "\n"
}

func normalizeHomeTaskMemoryFragmentID(raw any) string {
	idText := strings.TrimSpace(cast.ToString(raw))
	if idText == `` || idText == `0` {
		return ``
	}
	return idText
}

func enrichHomeTaskListWithMemoryFragment(list []map[string]any) {
	if component.MemoryRuntime == nil || component.MemoryRuntime.EnsureConfigured() != nil {
		return
	}
	memoryDB := component.MemoryRuntime.DB()
	for _, item := range list {
		memoryFragmentID := normalizeHomeTaskMemoryFragmentID(item[`memory_fragment_id`])
		item[`memory_fragment_id`] = memoryFragmentID
		if memoryFragmentID == `` {
			item[`memory_fragment`] = map[string]any{}
			continue
		}
		info, infoErr := memoryDB.MemoryFragmentInfo(memoryFragmentID)
		if infoErr != nil {
			item[`memory_fragment`] = map[string]any{
				`id`:      memoryFragmentID,
				`file_id`: memoryFragmentID,
				`title`:   `关联片段不存在`,
				`tags`:    []string{},
				`content`: ``,
				`missing`: true,
			}
			continue
		}
		item[`memory_fragment`] = map[string]any{
			`id`:      info[`id`],
			`file_id`: cast.ToString(info[`file_id`]),
			`title`:   cast.ToString(info[`title`]),
			`tags`:    cast.ToStringSlice(info[`tags`]),
			`content`: cast.ToString(info[`content`]),
			`missing`: false,
		}
	}
}

// autoCreateHomeTaskApiDir 自动在指定集合下创建接口文件夹，返回文件夹 ID。
func autoCreateHomeTaskApiDir(taskName string, collectionID int) (int, error) {
	taskName = strings.TrimSpace(taskName)
	if taskName == `` {
		taskName = `默认文件夹`
	}
	now := time.Now().Unix()
	dirData := map[string]any{
		`name`:          taskName,
		`collection_id`: collectionID,
		`headers`:       `{}`,
		`create_time`:   now,
		`update_time`:   now,
	}
	newID, err := common.DbMain.Client.QuickCreate(`tbl_api_dir`, dirData).Exec()
	if err != nil {
		return 0, err
	}
	return cast.ToInt(newID), nil
}

// resolveHomeTaskGitIDsJSON 解析 git_ids JSON，若前端未传则从旧字段回退。
func resolveHomeTaskGitIDsJSON(req *_struct.HomeTaskSaveRequest) string {
	raw := strings.TrimSpace(req.GitIds)
	if raw != `` && raw != `[]` {
		var ids []int
		if err := json.Unmarshal([]byte(raw), &ids); err == nil && len(ids) > 0 {
			return raw
		}
	}
	if req.GitID > 0 {
		bytes, _ := json.Marshal([]int{req.GitID})
		return string(bytes)
	}
	return `[]`
}

// resolveHomeTaskApiDevEntriesJSON 解析 api_dev_entries JSON，若前端未传则从旧字段回退。
func resolveHomeTaskApiDevEntriesJSON(req *_struct.HomeTaskSaveRequest) string {
	raw := strings.TrimSpace(req.ApiDevEntries)
	if raw != `` && raw != `[]` {
		var entries []_struct.ApiDevEntry
		if err := json.Unmarshal([]byte(raw), &entries); err == nil && len(entries) > 0 {
			return raw
		}
	}
	if req.ApiCollectionID > 0 {
		entry := _struct.ApiDevEntry{CollectionID: req.ApiCollectionID, DirID: req.ApiDirID}
		bytes, _ := json.Marshal([]_struct.ApiDevEntry{entry})
		return string(bytes)
	}
	return `[]`
}

// autoCreateHomeTaskApiDirs 为 api_dev_entries 中每个没有 dir_id 的条目自动创建文件夹，返回更新后的 JSON。
func autoCreateHomeTaskApiDirs(taskName string, entriesJSON string) string {
	var entries []_struct.ApiDevEntry
	if err := json.Unmarshal([]byte(entriesJSON), &entries); err != nil {
		return entriesJSON
	}
	changed := false
	for i, entry := range entries {
		if entry.CollectionID > 0 && entry.DirID <= 0 {
			dirID, err := autoCreateHomeTaskApiDir(taskName, entry.CollectionID)
			if err != nil {
				continue
			}
			entries[i].DirID = dirID
			changed = true
		}
	}
	if !changed {
		return entriesJSON
	}
	bytes, _ := json.Marshal(entries)
	return string(bytes)
}

// resolveHomeTaskDevConfigsJSON 解析 dev_configs JSON，若前端未传则从旧字段回退构建。
func resolveHomeTaskDevConfigsJSON(req *_struct.HomeTaskSaveRequest) string {
	raw := strings.TrimSpace(req.DevConfigs)
	if raw != `` && raw != `[]` {
		var configs []_struct.DevConfig
		if err := json.Unmarshal([]byte(raw), &configs); err == nil && len(configs) > 0 {
			return raw
		}
	}
	// 从旧字段回退构建一个 dev_config 条目。
	cfg := _struct.DevConfig{
		GitID:        req.GitID,
		CollectionID: req.ApiCollectionID,
		DirID:        req.ApiDirID,
	}
	if cfg.GitID > 0 || cfg.CollectionID > 0 {
		bytes, _ := json.Marshal([]_struct.DevConfig{cfg})
		return string(bytes)
	}
	return `[]`
}

// autoCreateHomeTaskDevConfigDirs 为 dev_configs 中每个 dir_id=0 且 collection_id>0 的条目自动创建文件夹。
func autoCreateHomeTaskDevConfigDirs(taskName string, configsJSON string) string {
	var configs []_struct.DevConfig
	if err := json.Unmarshal([]byte(configsJSON), &configs); err != nil {
		return configsJSON
	}
	changed := false
	for i, cfg := range configs {
		if cfg.CollectionID > 0 && cfg.DirID <= 0 {
			dirID, err := autoCreateHomeTaskApiDir(taskName, cfg.CollectionID)
			if err != nil {
				continue
			}
			configs[i].DirID = dirID
			changed = true
		}
	}
	if !changed {
		return configsJSON
	}
	bytes, _ := json.Marshal(configs)
	return string(bytes)
}
