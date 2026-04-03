package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func GroupList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	groupType := cast.ToInt(dataMap[`type`])
	allowGroupList := define.GetGroupTypeList()
	if !gstool.ArrayExistValue(&allowGroupList, groupType) {
		gsgin.GinResponseError(c, `分组类型错误`, nil)
		return
	}
	all, allErr := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: groupType,
	}).All()
	if allErr != nil {
		gsgin.GinResponseError(c, allErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, all)
}

func GroupAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `extra_1`, `extra_2`, `extra_3`, `extra_4`, `extra_5`, `extra_6`})
	if cast.ToString(dataMap[`name`]) == `` {
		gsgin.GinResponseError(c, `分组名称不能为空`, nil)
		return
	}
	groupType := cast.ToInt(dataMap[`type`])
	if cast.ToInt(dataMap[`id`]) == 0 {
		allowGroupList := define.GetGroupTypeList()
		if !gstool.ArrayExistValue(&allowGroupList, groupType) {
			gsgin.GinResponseError(c, `分组类型错误`, nil)
			return
		}
		updateData[`type`] = groupType
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, err := common.DbMain.Client.QuickCreate(`tbl_group`, updateData).Exec()
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_group`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	if groupType == define.GroupTypeShellOut {
		component.ShellOutClient.InitGroupConfigs()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func GroupDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = common.DbMain.Client.QuickDelete(`tbl_group`, map[string]any{
			`id`: dataMap[`id`],
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}
