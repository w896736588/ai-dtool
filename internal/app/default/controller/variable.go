package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"dev_tool/internal/pkg/p_variable"
	"gitee.com/Sxiaobai/gs/gs"
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
	for _, variable := range variableList {
		variableCmdList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
			`variable_id`: cast.ToString(variable[`id`]),
			`status`:      define.VariableStatusNormal,
		}).Order(`weight asc`).All()
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
	dataGMap := gs.NewTransMap(&dataMap)
	if dataGMap.G(`id`).IsZero() {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_variable`, map[string]any{
			`id`: dataGMap.G(`id`).ToStr(),
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
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `type`, `variable_id`, `is_pre`, `result_key`, `options`, `remark`, `sql`, `cmd`, `bash`, `weight`, `default`, `smart_link_id`, `smart_link_label`})
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
	dataGMap := gs.NewTransMap(&dataMap)
	if dataGMap.G(`id`).IsZero() {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_variable_cmd`, map[string]any{
			`id`: dataGMap.G(`id`).ToStr(),
		}, map[string]interface{}{
			`status`: define.VariableStatusDelete,
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// VariableCmdRunPre 执行前第一步
func VariableCmdRunPre(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	variableId := cast.ToInt(dataMap[`variable_id`])
	if variableId == 0 {
		gsgin.GinResponseError(c, `变量id不能为空`, nil)
		return
	}
	variable := p_variable.NewVariableRun()
	variable.VariableId = cast.ToString(variableId)
	formList, replaceList, isCanRun, err := variable.RunPre(variableId)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`form_list`:     formList,
		`replace_list`:  replaceList,
		`is_can_run`:    isCanRun,
		`run_unique_id`: variable.RunUniqueId,
	})
}

// VariableCmdRunProcess 执行第二步
func VariableCmdRunProcess(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	variableFormListStr := cast.ToString(dataMap[`variable_form_list`])
	variableFormList := make([]_struct.VariableForm, 0)
	decodeErr := gstool.JsonDecode(variableFormListStr, &variableFormList)
	if decodeErr != nil {
		gsgin.GinResponseError(c, decodeErr.Error(), nil)
		return
	}
	replaceListStr := cast.ToString(dataMap[`replace_list`])
	replaceList := make([]map[string]string, 0)
	decodeErr = gstool.JsonDecode(replaceListStr, &replaceList)
	if decodeErr != nil {
		gsgin.GinResponseError(c, decodeErr.Error(), nil)
		return
	}
	variable := p_variable.NewVariableRun()
	variable.VariableId = cast.ToString(dataMap[`variable_id`])
	variable.CmdId = cast.ToString(dataMap[`cmd_id`])
	formList, replaceList, isCanRun, err := variable.RunProcess(variableFormList, replaceList)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`form_list`:     formList,
		`replace_list`:  replaceList,
		`is_can_run`:    isCanRun,
		`run_unique_id`: variable.RunUniqueId,
	})
}

// VariableCmdRunDone 最终执行
func VariableCmdRunDone(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	variableId := cast.ToInt(dataMap[`variable_id`])
	replaceLists := cast.ToString(dataMap[`replace_list`])
	variableFormListStr := cast.ToString(dataMap[`variable_form_list`])
	variableFormList := make([]_struct.VariableForm, 0)
	decodeErr := gstool.JsonDecode(variableFormListStr, &variableFormList)
	if decodeErr != nil {
		gsgin.GinResponseError(c, decodeErr.Error(), nil)
		return
	}
	replaceList := make([]map[string]string, 0)
	err := gstool.JsonDecode(replaceLists, &replaceList)
	if err != nil {
		gsgin.GinResponseError(c, `解析replace_list失败`, nil)
		return
	}
	if variableId == 0 {
		gsgin.GinResponseError(c, `变量id不能为空`, nil)
		return
	}
	variable := p_variable.NewVariableRun()
	variable.VariableId = cast.ToString(dataMap[`variable_id`])
	err = variable.RunDone(variableId, replaceList, variableFormList)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]string{
		`run_unique_id`: variable.RunUniqueId,
	})
}
