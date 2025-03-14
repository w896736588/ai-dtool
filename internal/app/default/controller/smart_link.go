package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"errors"
	"gitee.com/Sxiaobai/gs/gs"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"time"
)

// SmartLinkUpWebkit 更新核心
func SmartLinkUpWebkit(c *gin.Context) {
	installErr := playwright.Install()
	if installErr != nil {
		gsgin.GinResponseError(c, `安装浏览器核心失败 %s`, installErr.Error())
		return
	}
	gsgin.GinResponseSuccess(c, `更新浏览器核心成功`, ``)
	return
}

// SmartLinkList 获取列表
func SmartLinkList(c *gin.Context) {
	variableGroupList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeSmartLink,
	}).All()
	smartLinkList, _ := base.Component.TSqlite.Client.QueryBySql(`select * from tbl_smart_link where status = ? order by weight asc`, define.SmartLinkStatusNormal).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`group_list`:      variableGroupList,
		`smart_link_list`: smartLinkList,
	})
}

// SmartLinkInfo 获取单个详情
func SmartLinkInfo(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	smartLinkId := dataMap[`id`]
	if cast.ToInt(smartLinkId) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	}
	smartLinkInfo, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_smart_link`, `*`, map[string]any{
		`id`:     smartLinkId,
		`status`: define.SmartLinkStatusNormal,
	}).One()
	smartLinkProcessList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_smart_link_process`, `*`, map[string]any{
		`smart_link_id`: smartLinkId,
		`status`:        define.SmartLinkStatusNormal,
	}).Order(`weight asc`).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`smart_link_info`:         smartLinkInfo,
		`smart_link_process_list`: smartLinkProcessList,
	})
}

// SmartLinkAdd 新增
func SmartLinkAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	validateErr := validateProcess(cast.ToString(dataMap[`process`]))
	if validateErr != nil {
		gsgin.GinResponseError(c, validateErr.Error(), nil)
		return
	}
	var id any
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `smart_link_group_id`, `links`, `open_num`, `open_type`, `process`, `weight`, `is_save_user_data`, `is_combine`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_smart_link`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	variable, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_smart_link`, `*`, map[string]any{
		`id`:     id,
		`status`: define.SmartLinkStatusNormal,
	}).One()
	gsgin.GinResponseSuccess(c, ``, variable)
}

func validateProcess(process string) error {
	if process == `` {
		return nil
	}
	processList := make([]map[string]any, 0)
	decodeErr := gstool.JsonDecode(process, &processList)
	if decodeErr != nil {
		return errors.New(`解析process失败`)
	}
	for _, processVal := range processList {
		//类型
		processType := cast.ToString(processVal[`type`])
		if processType == `` {
			return errors.New(`type不能为空`)
		}
		//元素选择
		Locator := cast.ToString(processVal[`Locator`])
		switch processType {
		case `click`: //点击
			if Locator == `` {
				return errors.New(`type为click时Locator不能为空`)
			}
		case `input`: //输入
			if cast.ToString(processVal[`value`]) == `` {
				return errors.New(`type为input时value不能为空`)
			}
		case `redirect_uri`: //跳转 保持当前域名
			if cast.ToString(processVal[`uri`]) == `` {
				return errors.New(`type为redirect_uri时，uri不能为空`)
			}
		}
	}
	return nil
}

// SmartLinkDelete 删除
func SmartLinkDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	dataGMap := gs.NewTransMap(&dataMap)
	if dataGMap.G(`id`).IsZero() {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link`, map[string]any{
			`id`: dataGMap.G(`id`).ToStr(),
		}, map[string]interface{}{
			`status`: define.SmartLinkStatusDelete,
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SmartLinkProcessAdd 新增子操作
func SmartLinkProcessAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `process_type`, `smart_link_id`, `selecter`, `weight`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		_, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_smart_link_process`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link_process`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SmartLinkProcessDel 删除子操作
func SmartLinkProcessDel(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	dataGMap := gs.NewTransMap(&dataMap)
	if dataGMap.G(`id`).IsZero() {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link_process`, map[string]any{
			`id`: dataGMap.G(`id`).ToStr(),
		}, map[string]interface{}{
			`status`: define.SmartLinkStatusDelete,
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SmartLinkRunPlaywright 执行 playwright
func SmartLinkRunPlaywright(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	id := cast.ToInt(dataMap[`id`])
	label := cast.ToString(dataMap[`label`])
	if id == 0 || label == `` {
		gsgin.GinResponseError(c, `id和label不能为空`, nil)
		return
	}
	browserAuthUsername := cast.ToString(dataMap[`browser_auth_username`])
	browserAuthPassword := cast.ToString(dataMap[`browser_auth_password`])
	userName := cast.ToString(dataMap[`user_name`])
	password := cast.ToString(dataMap[`password`])
	openNum := cast.ToInt(dataMap[`open_num`])
	runParams, runParamsErr := base.Component.TSmartLink.GetRunParams(id, label, browserAuthUsername, browserAuthPassword, userName, password, openNum, make([]map[string]string, 0))
	if runParamsErr != nil {
		gsgin.GinResponseError(c, runParamsErr.Error(), nil)
		return
	}
	gstool.FmtPrintlnLogTime(gstool.JsonEncode(runParams))
	for i := 0; i < runParams.OpenNum; i++ {
		openErr := base.Component.TSmartLink.OpenBrowserPlaywright(runParams)
		if openErr != nil {
			gstool.FmtPrintlnLogTime(`错误 %s`, openErr.Error())
		}
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SmartLinkRunPlaywrightList 获取运行的列表
func SmartLinkRunPlaywrightList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	runList := base.Component.TSmartLink.GetPlaywrightRunList()
	gsgin.GinResponseSuccess(c, ``, runList)
}

func SmartLinkPlaywrightVersion(c *gin.Context) {
	pw, pwErr := base.Component.TSmartLink.SmartLinkPlaywrightVersion()
	if pwErr != nil {
		gsgin.GinResponseError(c, `查询失败`+pwErr.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, pw.Version, nil)
}
