package controller

import (
	"dev_tool/base"
	"dev_tool/base/define"
	"dev_tool/internal/pkg/p_playwright"
	"errors"
	"fmt"
	"gitee.com/Sxiaobai/gs/gsgin"
	"gitee.com/Sxiaobai/gs/gstool"
	"github.com/gin-gonic/gin"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
	"strings"
	"time"
)

// SmartLinkUpWebkit 更新核心
func SmartLinkUpWebkit(c *gin.Context) {
	pw, _ := playwright.NewDriver()
	go base.Component.TPlaywright.Install(pw.Version)
	gsgin.GinResponseSuccess(c, `更新浏览器核心中`, ``)
	return
}

func SmartLinkRecycle(c *gin.Context) {
	base.Component.TPlaywright.SseMsg(`开始释放实例`, true)
	p := p_playwright.NewPlaywright(nil, base.Component.TPlaywright.Log)
	err := p.Recycle()
	if err != nil {
		base.Component.TPlaywright.SseMsg(fmt.Sprintf(`释放失败 `+err.Error()), true)
		gsgin.GinResponseError(c, fmt.Sprintf(`释放失败 %s`, err.Error()), nil)
		return
	}
	base.Component.TPlaywright.SseMsg(fmt.Sprintf(`释放成功 `), true)
	gsgin.GinResponseSuccess(c, `释放成功`, ``)
	return
}

func SmartLinkDownloadPath(c *gin.Context) {
	err := base.Component.TPlaywright.SmartLinkDownloadPath()
	if err != nil {
		gsgin.GinResponseError(c, fmt.Sprintf(`释放失败 %s`, err.Error()), nil)
		return
	}
	gsgin.GinResponseSuccess(c, `释放成功`, ``)
	return
}

// SmartLinkList 获取列表
func SmartLinkList(c *gin.Context) {
	variableGroupList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeSmartLink,
	}).All()
	smartLinkList, _ := base.Component.TSqlite.Client.QueryBySql(`select * from tbl_smart_link where status = ? order by weight asc`, define.SmartLinkStatusNormal).All()
	//查找配置的账号组
	for smartLinkKey, smartLink := range smartLinkList {
		links := cast.ToString(smartLink[`links`])
		if links != `` {
			linkList := make([]map[string]any, 0)
			_ = gstool.JsonDecode(links, &linkList)
			//循环每个链接及其配置
			for linkKey, link := range linkList {
				userList := getAccountListByName(link)
				linkList[linkKey][`userList`] = userList
			}
			smartLinkList[smartLinkKey][`links`] = gstool.JsonEncode(linkList)
			smartLinkList[smartLinkKey][`open_type`] = cast.ToInt(smartLink[`open_type`])
			smartLinkList[smartLinkKey][`combine_type`] = cast.ToString(smartLink[`combine_type`])
			smartLinkList[smartLinkKey][`channel`] = cast.ToString(smartLink[`channel`])
		}
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`group_list`:      variableGroupList,
		`smart_link_list`: smartLinkList,
	})
}

func getAccountListByName(link map[string]any) []map[string]string {
	userList := make([]map[string]string, 0)

	accountListConfig := cast.ToString(link[`account_list`])
	accountListConfig = gstool.SReplaces(accountListConfig, map[string]string{
		`{`: ``,
		`}`: ``,
	})
	accountConfigGroup := strings.Split(accountListConfig, `:`)
	if len(accountConfigGroup) != 3 {
		return userList
	}
	groupName := accountConfigGroup[2]

	groupInfo, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`name`: groupName,
		`type`: define.GroupTypeAccount,
	}).One()
	if len(groupInfo) == 0 {
		return userList
	}
	groupId := cast.ToInt(groupInfo[`id`])
	accountList, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_account`, `*`, map[string]any{
		`account_group_id`: groupId,
	}).All()
	if len(accountList) == 0 {
		return userList
	}

	for _, account := range accountList {
		userList = append(userList, map[string]string{
			`user_name`: cast.ToString(account[`username`]),
			`password`:  cast.ToString(account[`password`]),
		})
	}
	return userList
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
	var id any
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `smart_link_group_id`, `links`, `open_num`, `open_type`, `weight`, `combine_type`, `download_finds`, `auto_close_second`, `channel`, `show_cookies`, `process_id`})
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

func validateProcess(processVal map[string]any) error {
	//类型
	processType := cast.ToString(processVal[`type`])
	if processType == `` {
		return errors.New(`type不能为空`)
	}
	//元素选择
	Locator := cast.ToString(processVal[`locator`])
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
		if cast.ToString(processVal[`value`]) == `` {
			return errors.New(`type为redirect_uri时，value不能为空`)
		}
	}
	return nil
}

// SmartLinkDelete 删除
func SmartLinkDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link`, map[string]any{
			`id`: cast.ToInt(dataMap[`id`]),
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
	userName := cast.ToString(dataMap[`user_name`])
	password := cast.ToString(dataMap[`password`])
	openNum := max(1, cast.ToInt(dataMap[`open_num`]))
	replaceList := make(map[string]string)
	gstool.FmtPrintlnLogTime(`开始运行？`)
	base.Component.TPlaywright.SseMsg(base.Component.TMarkDown.BlockQuote(`运行,开始----------------我是分隔君`), true)
	for i := 0; i < openNum; i++ {
		go func() {
			streamFunc := func(name, msg string) {
				//输出到前端
				base.Component.TPlaywright.SseMsg(base.Component.TMarkDown.Bold(name)+`,`+msg, true)
			}
			runParams, runParamsErr := base.Component.TPlaywright.GetRunParams(id, label, userName, password, openNum, replaceList)
			if runParamsErr != nil {
				gstool.FmtPrintlnLogTime(`打开错误 %s`, runParamsErr.Error())
				return
			}
			runParams.StreamFunc = streamFunc
			p := p_playwright.NewPlaywright(runParams, base.Component.TPlaywright.Log)
			openErr := p.Open()
			if openErr != nil {
				gstool.FmtPrintlnLogTime(`错误 %s`, openErr.Error())
			}
		}()
		time.Sleep(time.Second * 2)
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SmartLinkRunPlaywrightList 获取运行的列表
func SmartLinkRunPlaywrightList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	contextPageList := p_playwright.NewContextList(base.Component.TPlaywright.Log)
	runList := contextPageList.GetPlaywrightRunList()
	gsgin.GinResponseSuccess(c, ``, runList)
}

func SmartLinkPlaywrightVersion(c *gin.Context) {
	base.Component.TPlaywright.SseMsg(`获取核心版本`, true)
	pw, pwErr := base.Component.TPlaywright.SmartLinkPlaywrightVersion()
	if pwErr != nil {
		base.Component.TPlaywright.SseMsg(`获取核心版本失败`+pwErr.Error(), true)
		gsgin.GinResponseError(c, `查询失败`+pwErr.Error(), nil)
		return
	}
	//是否在安装中
	isInstall := 0
	if !gstool.FileIsExisted(base.Component.TPlaywright.LockFileFullPath) {
		base.Component.TPlaywright.SseMsg(`核心正在安装中`, true)
		isInstall = 1
	} else {
		content, _ := gstool.FileGetContent(base.Component.TPlaywright.LockFileFullPath)
		if content == `` {
			base.Component.TPlaywright.SseMsg(`核心正在安装中`, true)
			isInstall = 1
		} else {
			base.Component.TPlaywright.SseMsg(`当前核心版本`+content, true)
		}
	}
	gsgin.GinResponseSuccess(c, pw.Version, map[string]any{
		`is_install`: isInstall,
		`version`:    pw.Version,
	})
}

// SmartProcessList 获取列表
func SmartProcessList(c *gin.Context) {
	list, _ := base.Component.TSqlite.Client.QueryBySql(`select * from tbl_smart_link_process where status = 1  order by id desc`).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

// SmartProcessAdd 新增
func SmartProcessAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	var id any
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_smart_link_process`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link_process`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_smart_link_process`, `*`, map[string]any{
		`id`:     id,
		`status`: 1,
	}).One()
	gsgin.GinResponseSuccess(c, ``, info)
}

// SmartProcessDelete 删除
func SmartProcessDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link_process`, map[string]any{
			`id`: cast.ToInt(dataMap[`id`]),
		}, map[string]interface{}{
			`status`: define.SmartLinkStatusDelete,
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SmartProcessItemList 获取列表
func SmartProcessItemList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	smartLinkProcessId := cast.ToInt(dataMap[`smart_link_process_id`])
	if smartLinkProcessId == 0 {
		gsgin.GinResponseError(c, `smart_link_process_id不能为空`, nil)
		return
	}
	list, _ := base.Component.TSqlite.Client.QueryBySql(`
		select * from tbl_smart_link_process_item where smart_link_process_id = ? and status = ? order by weight asc`, smartLinkProcessId, 1).All()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
	})
}

// SmartProcessItemAdd 新增
func SmartProcessItemAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	validateErr := validateProcess(dataMap)
	if validateErr != nil {
		gsgin.GinResponseError(c, validateErr.Error(), nil)
		return
	}
	smartLinkProcessId := cast.ToInt(dataMap[`smart_link_process_id`])
	if smartLinkProcessId == 0 {
		gsgin.GinResponseError(c, `smart_link_process_id不能为空`, nil)
		return
	}
	var id any
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `wait_mills`, `is_async`, `append_to_replace`, `smart_link_process_id`, `type`, `locator`, `tip`, `value`, `out_key`, `check_key`, `weight`, `domain_limit`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := base.Component.TSqlite.Client.QuickCreate(`tbl_smart_link_process_item`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link_process_item`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := base.Component.TSqlite.Client.QuickQuery(`tbl_smart_link_process_item`, `*`, map[string]any{
		`id`:     id,
		`status`: 1,
	}).One()
	gsgin.GinResponseSuccess(c, ``, info)
}

// SmartProcessItemDelete 删除
func SmartProcessItemDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	} else {
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link_process_item`, map[string]any{
			`id`: cast.ToInt(dataMap[`id`]),
		}, map[string]interface{}{
			`status`: define.SmartLinkStatusDelete,
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SmartProcessItemSort 排序
func SmartProcessItemSort(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	smartLinkProcessId := cast.ToInt(dataMap[`smart_link_process_id`])
	if smartLinkProcessId == 0 {
		gsgin.GinResponseError(c, `smart_link_process_id不能为空`, nil)
		return
	}
	smartLinkProcessItemIds := cast.ToString(dataMap[`smart_link_process_item_ids`])
	if smartLinkProcessItemIds == `` {
		gsgin.GinResponseError(c, `smart_link_process_item_ids不能为空`, nil)
		return
	}
	smartLinkProcessItemIdsArr := strings.Split(smartLinkProcessItemIds, `,`)
	for index, item := range smartLinkProcessItemIdsArr {
		_, _ = base.Component.TSqlite.Client.QuickUpdate(`tbl_smart_link_process_item`, map[string]any{
			`id`: cast.ToInt(item),
		}, map[string]interface{}{
			`weight`: index + 1,
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}
