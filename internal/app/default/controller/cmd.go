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

func CmdList(c *gin.Context) {
	gitGroupList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeVariable,
	}).All()
	gitList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_variable`, `*`, nil).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`variable_group_list`: gitGroupList,
		`variable_list`:       gitList,
	})
}

func CmdAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	_type := dataMap[`type`]
	if cast.ToInt(_type) == 0 {
		gsgin.GinResponseError(c, `类型不能为空`, nil)
		return
	}
	configStr := cast.ToString(dataMap[`config`])
	switch _type {
	case define.VariableCmdMysql:
		config := _struct.VariableMysql{}
		_ = gstool.JsonDecode(configStr, &config)
		if cast.ToString(config.Sql) == `` || cast.ToInt(config.MysqlId) == 0 {
			gsgin.GinResponseError(c, `mysql类型格式错误`, nil)
			return
		}
	case define.VariableCmdCmd:
		config := _struct.VariableCmd{}
		_ = gstool.JsonDecode(configStr, &config)
		if cast.ToInt(config.CmdId) == 0 {
			gsgin.GinResponseError(c, `cmd类型格式错误`, nil)
			return
		}
	default:

	}

	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `type`, `variable_group_id`, `remark`})
	updateData[`config`] = configStr
	updateData[`key`] = gstool.TimeNowMilliInt64()
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_variable`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_variable`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func CmdDelete(c *gin.Context) {
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
