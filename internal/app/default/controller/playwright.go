package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"gitee.com/Sxiaobai/gs/gs"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"log"
	"net/url"
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

// SmartLinkRunPlaywright 执行 playwright
func SmartLinkRunPlaywright(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	link := cast.ToString(dataMap[`link`])
	openNum := cast.ToInt(dataMap[`open_num`])
	openType := cast.ToInt(dataMap[`open_type`])
	process := cast.ToString(dataMap[`process`])
	processList := make([]map[string]any, 0)
	_ = gstool.JsonDecode(process, &processList)
	gstool.FmtPrintlnLogTime(`processList %#v`, processList)
	for i := 0; i < openNum; i++ {
		go func() {
			openErr := openBrowserPlaywright(openType, link, processList, dataMap)
			if openErr != nil {
				gstool.FmtPrintlnLogTime(`错误 %s`, openErr.Error())
			}
		}()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// 打开浏览器
func openBrowserPlaywright(openType int, link string, processList []map[string]any, dataMap map[string]any) error {
	browserAuthUsername := cast.ToString(dataMap[`browser_auth_username`])
	browserAuthPassword := cast.ToString(dataMap[`browser_auth_password`])
	page, pageErr := base.Component.TSmartLink.GetPage(openType, browserAuthUsername, browserAuthPassword)
	if pageErr != nil {
		return pageErr
	}
	//跳转链接
	u, _ := url.Parse(link)
	gstool.FmtPrintlnLogTime(`打开网页 %s %#v`, u.String(), dataMap)
	if _, goErr := (*page.Page).Goto(u.String()); goErr != nil {
		return goErr
	}

	// 使用 JavaScript 将浏览器窗口最大化
	jsErr, _ := (*page.Page).Evaluate(`() => {
		console.log('初始化js');
    }`)
	if jsErr != nil {
		log.Fatalf("could not maximize browser window: %v", jsErr)
	}
	for _, processVal := range processList {
		time.Sleep(time.Millisecond * 200)
		//类型
		processType := cast.ToString(processVal[`type`])
		//元素选择
		selector := cast.ToString(processVal[`selector`])
		//链接
		redirectUri := cast.ToString(processVal[`uri`])
		//操作描述
		tip := cast.ToString(processVal[`tip`])
		//等待时间
		waitSecond := cast.ToFloat64(processVal[`wait_second`])
		if waitSecond == 0 {
			waitSecond = cast.ToFloat64(2)
		}
		waitSecond = waitSecond * 1000
		//跳转地址
		waitNavigation := cast.ToInt(processVal[`waitNavigation`])
		gstool.FmtPrintlnLogTime(`操作 %#v`, processVal)

		// 等待导航完成
		gstool.FmtPrintlnLogTime(`等待页面加载完成`)
		waitErr := (*page.Page).WaitForLoadState(playwright.PageWaitForLoadStateOptions{
			State: playwright.LoadStateLoad, //三种LoadStateNetworkidle 网络加载最低程度 LoadStateDomcontentloaded DOM加载完成
		})
		if waitErr != nil {
			gstool.FmtPrintlnLogTime("等待页面 DOM 内容加载完成失败: %s", waitErr.Error())
		}
		gstool.FmtPrintlnLogTime(`%s 元素：%s 等待时间：%f 等待导航完成：%d`, tip, selector, waitSecond, waitNavigation)
		switch processType {
		case `wait_navigation`: //等待导航完成
			waitUrlErr := (*page.Page).WaitForURL((*page.Page).URL())
			if waitUrlErr != nil {
				return waitUrlErr
			}
		case `click`: //点击
			selectorLoader := (*page.Page).Locator(selector)
			selectorLoaderWaitErr := selectorLoader.WaitFor(playwright.LocatorWaitForOptions{
				Timeout: &waitSecond,
				State:   playwright.WaitForSelectorStateVisible,
			})
			if selectorLoaderWaitErr == nil {
				clickErr := selectorLoader.Click()
				if clickErr != nil {
					gstool.FmtPrintlnLogTime(`等待元素后 点击失败 %s`, clickErr.Error())
				}
			}
			break
		case `input`: //输入
			inputValue := cast.ToString(processVal[`value`])
			inputValue = gstool.StringReplaces(inputValue, map[string]string{
				`{user_name}`: cast.ToString(dataMap[`user_name`]),
				`{password}`:  cast.ToString(dataMap[`password`]),
			})
			inputSelecter := (*page.Page).Locator(selector)
			selectorLoaderWaitErr := inputSelecter.WaitFor(playwright.LocatorWaitForOptions{
				Timeout: &waitSecond,
			})
			if selectorLoaderWaitErr == nil {
				inputErr := inputSelecter.Fill(inputValue)
				if inputErr != nil {
					gstool.FmtPrintlnLogTime("无法将元素转换为输入框: %v", inputErr.Error())
				}
			}
			break
		case `redirect_uri`: //跳转 保持当前域名
			currentURL := (*page.Page).URL()
			parsedURL, err := url.Parse(currentURL)
			if err != nil {
				gstool.FmtPrintlnLogTime("could not parse URL: %v", err)
			}
			domain := parsedURL.Scheme + `://` + parsedURL.Host
			gstool.FmtPrintlnLogTime(`跳转地址 %s`, domain+redirectUri)
			if _, goErr := (*page.Page).Goto(domain + redirectUri); goErr != nil {
				return goErr
			}
		}
	}
	//if err = page.Click(`.switch-input`); err != nil {
	//	log.Fatalf("could not fill password input: %v", err)
	//}
	//if err = browser.Close(); err != nil {
	//	log.Fatalf("could not close browser: %v", err)
	//}
	//if err = pw.Stop(); err != nil {
	//	log.Fatalf("could not stop Playwright: %v", err)
	//}
	return nil
}
