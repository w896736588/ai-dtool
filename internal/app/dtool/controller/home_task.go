package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	_struct "dev_tool/internal/app/dtool/struct"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
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
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`task_list`: list,
	})
}

// HomeTaskSave 保存首页任务。
func HomeTaskSave(c *gin.Context) {
	request := _struct.HomeTaskSaveRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := common.DbMain.HomeTaskSave(request.ID, request.Name, request.TaskStatus, request.Remark, request.StartTime)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
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
	taskList, err := common.DbMain.HomeTaskList(define.HomeTaskArchivedNo)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
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
