package controller

import (
	"dev_tool/internal/app/dtool/common"
	_struct "dev_tool/internal/app/dtool/struct"

	"github.com/gin-gonic/gin"
	"github.com/w896736588/go-tool/gsgin"
)

// TaskStatusList 查询所有任务状态。
func TaskStatusList(c *gin.Context) {
	list, err := common.DbMain.TaskStatusList()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

// TaskStatusSave 新增或编辑任务状态。
func TaskStatusSave(c *gin.Context) {
	request := _struct.TaskStatusSaveRequest{}
	_ = gsgin.GinPostBody(c, &request)
	list, err := common.DbMain.TaskStatusSave(request.ID, request.Name, request.SortOrder)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

// TaskStatusDelete 删除任务状态。
func TaskStatusDelete(c *gin.Context) {
	request := _struct.TaskStatusDeleteRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if err := common.DbMain.TaskStatusDelete(request.ID); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	list, _ := common.DbMain.TaskStatusList()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

// TaskStatusSort 更新任务状态排序。
func TaskStatusSort(c *gin.Context) {
	request := _struct.TaskStatusSortRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if err := common.DbMain.TaskStatusSort(request.IDs); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	list, _ := common.DbMain.TaskStatusList()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}
