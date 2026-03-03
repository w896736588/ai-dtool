package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/variable"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_sse"
	"fmt"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

func VariableList(c *gin.Context) {
	variableGroupList, _ := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeVariable,
	}).All()
	for _, variableGroup := range variableGroupList {
		variableGroup[`variable_list`] = make([]map[string]any, 0)
	}
	variableList, _ := common.DbMain.Client.QuickQuery(`tbl_variable`, `*`, map[string]interface{}{
		`status`: define.VariableStatusNormal,
	}).All()
	for keyVariable, variable := range variableList {
		variableList[keyVariable][`id`] = cast.ToString(variable[`id`])
		variableCmdList, _ := common.DbMain.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
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
	variableInfo, _ := common.DbMain.Client.QuickQuery(`tbl_variable`, `*`, map[string]any{
		`id`:     variableId,
		`status`: define.VariableStatusNormal,
	}).One()
	variableCmdList, _ := common.DbMain.Client.QuickQuery(`tbl_variable_cmd`, `*`, map[string]any{
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
	variable.VariableClient.LoginUsername = cast.ToString(dataMap[`username`])
	variable.VariableClient.LoginPassword = cast.ToString(dataMap[`password`])
	gsgin.GinResponseSuccess(c, ``, nil)
}

func VariableAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	var id any
	dataMap[`type`] = 1 //固定为脚本
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `desc`, `variable_group_id`, `remark`, `type`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := common.DbMain.Client.QuickCreate(`tbl_variable`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_variable`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	variable, _ := common.DbMain.Client.QuickQuery(`tbl_variable`, `*`, map[string]any{
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
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_variable`, map[string]any{
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
	case define.VariableCmdLlm:
		// 大模型命令校验：必须包含提示词与模型
		if strings.TrimSpace(cast.ToString(dataMap[`bash`])) == `` {
			gsgin.GinResponseError(c, `大模型提示词不能为空`, nil)
			return
		}
		optionsMap := make(map[string]any)
		options := cast.ToString(dataMap[`options`])
		if options == `` {
			gsgin.GinResponseError(c, `大模型配置不能为空`, nil)
			return
		}
		if err := gstool.JsonDecode(options, &optionsMap); err != nil {
			gsgin.GinResponseError(c, `大模型配置格式错误`, nil)
			return
		}
		if strings.TrimSpace(cast.ToString(optionsMap[`model`])) == `` {
			gsgin.GinResponseError(c, `大模型model不能为空`, nil)
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
		_, createErr := common.DbMain.Client.QuickCreate(`tbl_variable_cmd`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_variable_cmd`,
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
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_variable_cmd`, map[string]any{
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
	replaceLists := cast.ToString(dataMap[`replace_list`])
	replaceList := make(map[string]string, 0)
	err := gstool.JsonDecode(replaceLists, &replaceList)
	if err != nil {
		gsgin.GinResponseError(c, `解析replace_list失败`, nil)
		return
	}
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: cast.ToString(dataMap[`sse_distribute_id`]),
	}
	set := variable.NewVariableSet(sse, variableId, runCmdId, editValue, replaceList, common.GetCall())
	result, setErr := set.Set()
	if setErr != nil {
		result.RunStatus = 2
		sse.Send(fmt.Sprintf(`error：%s`, setErr.Error()) + "\n")
	}
	gsgin.GinResponseSuccess(c, ``, result)
}

// VariableCmdRun 执行
func VariableCmdRun(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	variableId := cast.ToInt(dataMap[`variable_id`])
	runCmdId := cast.ToInt(dataMap[`run_cmd_id`]) //本次执行的某一步操作
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
	//注入常量
	p_common.TBaseClient.FillConst(replaceList)
	//如果是预执行 那么重置任务ID为0 让前一个退出
	if runCmdId == 0 {
		variable.VariableClient.CreateTask(``)
	}
	//登录任务执行中
	taskId := ``
	if isRun == 1 {
		taskId = p_common.TBaseClient.GetUnique(`variable_run_`)
		variable.VariableClient.CreateTask(taskId)
	}
	//sse
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: cast.ToString(dataMap[`sse_distribute_id`]),
	}

	variable := variable.NewVariable(sse, variableId, runCmdId, taskId, replaceList, common.GetCall())
	result, resultErr := variable.Run()
	if resultErr != nil {
		result.RunStatus = 2
		sse.Send(fmt.Sprintf(`执行失败%s`, resultErr.Error()) + "\n")
		gsgin.GinResponseError(c, fmt.Sprintf(`执行失败%s`, resultErr.Error()), nil)
	} else {
		gsgin.GinResponseSuccess(c, ``, result)
	}
}
