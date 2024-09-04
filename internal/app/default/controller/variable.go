package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	_struct "dev_tool/base/struct"
	"gitee.com/Sxiaobai/gs/gs"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"time"
)

func VariableList(c *gin.Context) {
	gitGroupList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeVariable,
	}).All()
	gitList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_variable`, `*`, nil).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`variable_group_list`: gitGroupList,
		`variable_list`:       gitList,
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
		`id`: variableId,
	}).One()
	variableCmdList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
		`variable_id`: variableId,
	}).Order(`weight asc`).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`variable_info`:     variableInfo,
		`variable_cmd_list`: variableCmdList,
	})
}

func VariableAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`variable_group_id`]) == 0 {
		gsgin.GinResponseError(c, `组id不能为空 `, nil)
		return
	}
	var id any
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `variable_group_id`, `remark`})
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
		`id`: id,
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_variable`, map[string]any{
			`id`: dataGMap.G(`id`).ToStr(),
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
	case define.VariableTypeMysql:
		if cast.ToString(dataMap[`sql`]) == `` || cast.ToInt(dataMap[`mysql_id`]) == 0 {
			gsgin.GinResponseError(c, `mysql类型格式错误`, nil)
			return
		}
	case define.VariableTypeCmd:
		if cast.ToString(dataMap[`cmd`]) == `` {
			gsgin.GinResponseError(c, `cmd格式错误`, nil)
			return
		}
	default:

	}

	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `type`, `variable_id`, `is_pre`, `result_key`, `options`, `remark`, `sql`, `sql_id`, `cmd`, `mysql_id`, `ssh_id`, `bash`, `weight`})
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
		_, _ = base.Component.TSqlite.Client.QuickDelete(`tbl_variable_cmd`, map[string]any{
			`id`: dataGMap.G(`id`).ToStr(),
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
	variable := base.NewVariable()
	formList, replaceList, isCanRun, err := variable.RunPre(variableId)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`form_list`:    formList,
		`replace_list`: replaceList,
		`is_can_run`:   isCanRun,
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
	variable := base.NewVariable()
	formList, replaceList, isCanRun, err := variable.RunProcess(variableFormList, replaceList)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`form_list`:    formList,
		`replace_list`: replaceList,
		`is_can_run`:   isCanRun,
	})
}

// VariableCmdRunDone 最终执行
func VariableCmdRunDone(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	variableId := cast.ToInt(dataMap[`variable_id`])
	replaceLists := cast.ToString(dataMap[`replace_list`])
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
	variable := base.NewVariable()
	err = variable.RunDone(variableId, replaceList)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}
