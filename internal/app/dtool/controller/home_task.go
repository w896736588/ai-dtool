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
	memoryFragmentID, err := ensureHomeTaskMemoryFragment(request)
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

func ensureHomeTaskMemoryFragment(request _struct.HomeTaskSaveRequest) (int, error) {
	taskName := strings.TrimSpace(request.Name)
	if taskName == `` {
		return 0, gstool.Error(`任务名称不能为空`)
	}
	if component.MemoryRuntime == nil {
		return 0, common.ErrMemoryNotConfigured
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return 0, err
	}
	memoryDB := component.MemoryRuntime.DB()
	memoryFragmentID := request.MemoryFragmentID
	if memoryFragmentID > 0 {
		if _, infoErr := memoryDB.MemoryFragmentInfo(memoryFragmentID); infoErr != nil {
			return 0, infoErr
		}
		return memoryFragmentID, nil
	}
	fragmentInfo, saveErr := memoryDB.MemoryFragmentSave(0, taskName, "# "+taskName+"\n\n", []string{`需求`})
	if saveErr != nil {
		return 0, saveErr
	}
	component.MemoryRuntime.ScheduleSync()
	return cast.ToInt(fragmentInfo[`id`]), nil
}

func enrichHomeTaskListWithMemoryFragment(list []map[string]any) {
	if component.MemoryRuntime == nil || component.MemoryRuntime.EnsureConfigured() != nil {
		return
	}
	memoryDB := component.MemoryRuntime.DB()
	for _, item := range list {
		memoryFragmentID := cast.ToInt(item[`memory_fragment_id`])
		if memoryFragmentID <= 0 {
			item[`memory_fragment`] = map[string]any{}
			continue
		}
		info, infoErr := memoryDB.MemoryFragmentInfo(memoryFragmentID)
		if infoErr != nil {
			item[`memory_fragment`] = map[string]any{
				`id`:      memoryFragmentID,
				`title`:   `关联片段不存在`,
				`tags`:    []string{},
				`missing`: true,
			}
			continue
		}
		item[`memory_fragment`] = map[string]any{
			`id`:      cast.ToInt(info[`id`]),
			`title`:   cast.ToString(info[`title`]),
			`tags`:    cast.ToStringSlice(info[`tags`]),
			`missing`: false,
		}
	}
}
