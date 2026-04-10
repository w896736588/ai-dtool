package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	_struct "dev_tool/internal/app/dtool/struct"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
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
	memoryFragmentID, err := ensureHomeTaskMemoryFragment(request.ID, request.Name, normalizeHomeTaskMemoryFragmentID(request.MemoryFragmentID))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	info, err := common.DbMain.HomeTaskSave(request.ID, request.Name, request.TaskStatus, request.StartTime, memoryFragmentID)
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

// HomeTaskDailyReportGenerate 生成首页工作日报并写入记忆库。
func HomeTaskDailyReportGenerate(c *gin.Context) {
	memoryDB, ok := memoryDBOrResponse(c)
	if !ok {
		return
	}
	activeTaskList, err := common.DbMain.HomeTaskList(define.HomeTaskArchivedNo)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	archivedTaskList, err := common.DbMain.HomeTaskList(define.HomeTaskArchivedYes)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	taskList := mergeHomeTaskDailyReportTaskList(activeTaskList, archivedTaskList)
	modelID, prompt, err := homeTaskDailyReportConfig()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	reportTime := time.Now()
	userPrompt, err := buildHomeTaskDailyReportUserPrompt(prompt, taskList, reportTime)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	result, modelInfo, err := common.DbMain.AIChatByModel(modelID, homeTaskDailyReportSystemPrompt(), userPrompt)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	memoryInfo, err := memoryDB.MemoryFragmentSave(
		0,
		buildHomeTaskDailyReportTitle(reportTime),
		stripMarkdownCodeFence(result),
		[]string{homeTaskDailyReportMemoryTag},
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`memory_fragment`: memoryInfo,
		`model_id`:        modelID,
		`model`:           modelInfo[`model`],
		`prompt`:          prompt,
	})
}

func ensureHomeTaskMemoryFragment(taskID int, taskName string, memoryFragmentID string) (string, error) {
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
	fragmentInfo, saveErr := memoryDB.MemoryFragmentSave(0, taskName, "# "+taskName+"\n\n", []string{`需求`})
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
