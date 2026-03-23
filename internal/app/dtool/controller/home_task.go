package controller

import (
	"dev_tool/internal/app/dtool/common"
	_struct "dev_tool/internal/app/dtool/struct"

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
