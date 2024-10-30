package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"fmt"
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
	smartLinkList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_smart_link`, `*`, map[string]interface{}{
		`status`: define.SmartLinkStatusNormal,
	}).All()
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
	//if cast.ToInt(dataMap[`smart_link_group_id`]) == 0 {
	//	gsgin.GinResponseError(c, `组id不能为空 `, nil)
	//	return
	//}
	var id any
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `smart_link_group_id`, `links`, `open_num`, `open_type`, `process`})
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

// SmartLinkRun 执行
func SmartLinkRun(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	link := cast.ToString(dataMap[`link`])
	//user_name := cast.ToString(dataMap[`user_name`])
	//password := cast.ToString(dataMap[`password`])
	openNum := cast.ToInt(dataMap[`open_num`])
	openType := cast.ToInt(dataMap[`open_type`])
	process := cast.ToString(dataMap[`process`])
	processList := make([]map[string]string, 0)
	_ = gstool.JsonDecode(process, &processList)
	gstool.FmtPrintlnLogTime(`processList %#v`, processList)
	//初始化
	pw, err := playwright.Run()
	if err != nil {
		gsgin.GinResponseError(c, fmt.Sprintf("could not start playwright: %v", err.Error()), nil)
		return
	}
	var browser playwright.Browser
	var errLaunch error
	if openType == define.OpenTypeWebkitSilence {
		browser, errLaunch = pw.Chromium.Launch()
		gstool.FmtPrintlnLogTime(`静默`)
	} else {
		browser, errLaunch = pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
			Headless: playwright.Bool(false), // 设置为非 Headless 模式
		})
		gstool.FmtPrintlnLogTime(`静默`)
	}
	if errLaunch != nil {
		gsgin.GinResponseError(c, fmt.Sprintf("could not launch browser: %v", errLaunch.Error()), nil)
		return
	}

	for i := 0; i < openNum; i++ {
		page, pageErr := browser.NewPage()
		if pageErr != nil {
			gsgin.GinResponseError(c, fmt.Sprintf("could not create page:  %v", pageErr.Error()), nil)
			return
		}
		if _, err = page.Goto(link); err != nil {
			gsgin.GinResponseError(c, fmt.Sprintf("could not goto:  %v", err.Error()), nil)
			return
		}
		for _, processVal := range processList {
			processType := cast.ToString(processVal[`type`])
			//selector := cast.ToString(processVal[`selector`])
			existSelector := cast.ToString(processVal[`exist_selector`])
			gstool.FmtPrintlnLogTime(`操作 %#v`, processVal)
			switch processType {
			case `click`:
				if existSelector != `` {
					gstool.FmtPrintlnLogTime(`开始查找 %s`, existSelector)
					exist := page.GetByRole(*playwright.AriaRoleLink, playwright.PageGetByRoleOptions{
						Name: `确定`,
					})
					exist.Click()
					gstool.FmtPrintlnLogTime(`%v`, exist)
					//if existElementErr != nil {
					//	continue
					//} else {
					//	_ = page.Click(selector)
					//}
				}
				break
			}
		}
		//找到键盘
		//if err = page.Click(`.switch-input`); err != nil {
		//	log.Fatalf("could not fill password input: %v", err)
		//}
		//if err = browser.Close(); err != nil {
		//	log.Fatalf("could not close browser: %v", err)
		//}
		//if err = pw.Stop(); err != nil {
		//	log.Fatalf("could not stop Playwright: %v", err)
		//}
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}
