package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"dev_tool/internal/pkg/p_variable"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"strings"
	"time"
)

func VariableList(c *gin.Context) {
	variableGroupList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeVariable,
	}).All()
	for _, variableGroup := range variableGroupList {
		variableGroup[`variable_list`] = make([]map[string]any, 0)
	}
	variableList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_variable`, `*`, map[string]interface{}{
		`status`: define.VariableStatusNormal,
	}).All()
	for keyVariable, variable := range variableList {
		variableList[keyVariable][`id`] = cast.ToString(variable[`id`])
		variableCmdList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
			`variable_id`: cast.ToString(variable[`id`]),
			`status`:      define.VariableStatusNormal,
		}).Order(`weight asc`).All()
		//转换类型
		for cmdKey, variableCmd := range variableCmdList {
			variableCmdList[cmdKey][`type`] = cast.ToString(variableCmd[`type`])
			variableCmdList[cmdKey][`id`] = cast.ToString(variableCmd[`id`])
		}
		variable[`variable_cmd_list`] = variableCmdList
		//归到分组中
		for _, variableGroup := range variableGroupList {
			if cast.ToString(variableGroup[`id`]) == cast.ToString(variable[`variable_group_id`]) {
				variableGroup[`variable_list`] = append(variableGroup[`variable_list`].([]map[string]any), variable)
			}
		}
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`variable_group_list`: variableGroupList,
		`variable_list`:       variableList,
	})
}

func VariableInfo(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	variableId := dataMap[`id`]
	if cast.ToInt(variableId) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	}
	variableInfo, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_variable`, `*`, map[string]any{
		`id`:     variableId,
		`status`: define.VariableStatusNormal,
	}).One()
	variableCmdList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
		`variable_id`: variableId,
		`status`:      define.VariableStatusNormal,
	}).Order(`weight asc`).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`variable_info`:     variableInfo,
		`variable_cmd_list`: variableCmdList,
	})
}

func VariableSetLogin(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	base.Component.TVariable.LoginUsername = cast.ToString(dataMap[`username`])
	base.Component.TVariable.LoginPassword = cast.ToString(dataMap[`password`])
	gsgin.GinResponseSuccess(c, ``, nil)
}

func VariableAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	var id any
	dataMap[`type`] = 1 //固定为脚本
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `variable_group_id`, `remark`, `type`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_variable`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_variable`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	variable, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_variable`, `*`, map[string]any{
		`id`:     id,
		`status`: define.VariableStatusNormal,
	}).One()
	gsgin.GinResponseSuccess(c, ``, variable)
}

func VariableDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_variable`, map[string]any{
			`id`: cast.ToInt(dataMap[`id`]),
		}, map[string]interface{}{
			`status`: define.VariableStatusDelete,
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func VariableCmdAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	_type := dataMap[`type`]
	if cast.ToInt(_type) == 0 {
		gsgin.GinResponseError(c, `类型不能为空`, nil)
		return
	}
	switch _type {
	case define.VariableCmdMysql:
		if cast.ToString(dataMap[`sql`]) == `` {
			gsgin.GinResponseError(c, `mysql缺少语句`, nil)
			return
		}
	case define.VariableCmdCmd:
		if cast.ToString(dataMap[`cmd`]) == `` {
			gsgin.GinResponseError(c, `cmd缺少内容`, nil)
			return
		}
	default:

	}
	resultKey := cast.ToString(dataMap[`result_key`])
	if resultKey != `` && (!strings.HasPrefix(resultKey, `{`) || !strings.HasSuffix(resultKey, `}`)) {
		gsgin.GinResponseError(c, `输出的key必须以'{'开头，以'}'结尾`, nil)
		return
	}
	if cast.ToInt(dataMap[`weight`]) == 0 {
		gsgin.GinResponseError(c, `weight不能为0`, nil)
		return
	}
	runTypeList := []string{define.RunTypeForm, define.RunTypeRun, define.RunTypeMiddle}
	if !gstool.ArrayExistValue(&runTypeList, cast.ToString(dataMap[`run_type`])) {
		gsgin.GinResponseError(c, `执行类型错误，只支持`+gstool.JsonEncode(runTypeList), nil)
		return
	}
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `type`, `variable_id`, `is_pre`, `result_key`, `options`, `remark`, `sql`, `cmd`, `bash`, `weight`, `default`, `smart_link_id`, `smart_link_label`, `checks`, `run_type`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`key`] = gstool.TimeNowMilliInt64()
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_variable_cmd`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_variable_cmd`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func VariableCmdDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_variable_cmd`, map[string]any{
			`id`: cast.ToInt(dataMap[`id`]),
		}, map[string]interface{}{
			`status`: define.VariableStatusDelete,
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func VariableCmdSet(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	variableId := cast.ToInt(dataMap[`variable_id`])
	runCmdId := cast.ToInt(dataMap[`run_cmd_id`])
	editValue := cast.ToString(dataMap[`edit_value`])
	runUniqueId := cast.ToString(dataMap[`run_unique_id`])
	if runUniqueId == `` {
		gsgin.GinResponseError(c, `缺少本次执行唯一ID`, nil)
		return
	}
	replaceLists := cast.ToString(dataMap[`replace_list`])
	replaceList := make(map[string]string, 0)
	err := gstool.JsonDecode(replaceLists, &replaceList)
	if err != nil {
		gsgin.GinResponseError(c, `解析replace_list失败`, nil)
		return
	}
	set := p_variable.NewVariableSet(variableId, runCmdId, editValue, runUniqueId, replaceList)
	result, setErr := set.Set()
	if setErr != nil {
		result.RunStatus = 2
		set.StreamMsg(fmt.Sprintf(`error：%s`, setErr.Error()), true)
	}
	gsgin.GinResponseSuccess(c, ``, result)
}

// VariableCmdRun 执行
func VariableCmdRun(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	variableId := cast.ToInt(dataMap[`variable_id`])
	runCmdId := cast.ToInt(dataMap[`run_cmd_id`])
	runUniqueId := cast.ToString(dataMap[`run_unique_id`])
	if runCmdId != 0 && runUniqueId == `` { //初始
		gsgin.GinResponseError(c, `缺少本次执行唯一ID`, nil)
		return
	}
	isRun := cast.ToInt(dataMap[`is_run`])
	replaceLists := cast.ToString(dataMap[`replace_list`])
	replaceList := make(map[string]string, 0)
	if replaceLists != `` {
		err := gstool.JsonDecode(replaceLists, &replaceList)
		if err != nil {
			gsgin.GinResponseError(c, `解析replace_list失败`, nil)
			return
		}
	}

	variable := p_variable.NewVariable(variableId, runCmdId, isRun, replaceList, runUniqueId)
	result, resultErr := variable.Run()
	if resultErr != nil {
		result.RunStatus = 2
		variable.StreamMsg(fmt.Sprintf(`执行失败%s`, resultErr.Error()), true)
	}
	gsgin.GinResponseSuccess(c, ``, result)
}
