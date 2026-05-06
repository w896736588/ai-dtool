package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/plw"
	"dev_tool/internal/pkg/p_common"
	"dev_tool/internal/pkg/p_sse"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/playwright-community/playwright-go"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// SmartLinkUpWebkit 更新核心
func SmartLinkUpWebkit(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	sseDistributeId := cast.ToString(dataMap[`sse_distribute_id`])
	pw, _ := playwright.NewDriver()
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: sseDistributeId,
	}
	if !ensureSmartLinkNodeInstalled(c, sse) {
		return
	}
	go component.PlaywrightClient.Install(sse, pw.Version)
	gsgin.GinResponseSuccess(c, `更新浏览器核心中`, ``)
	return
}

func SmartLinkRecycle(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	sseDistributeId := cast.ToString(dataMap[`sse_distribute_id`])
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: sseDistributeId,
	}
	if !ensureSmartLinkNodeInstalled(c, sse) {
		return
	}
	sse.Send(`开始释放实例` + "\n")
	p := plw.NewPlaywright(nil, component.PlaywrightClient.Log)
	err := p.Recycle()
	if err != nil {
		sse.Send(`释放失败 ` + err.Error() + "\n")
		gsgin.GinResponseError(c, fmt.Sprintf(`释放失败 %s`, err.Error()), nil)
		return
	}
	sse.Send(`释放成功 ` + "\n")
	gsgin.GinResponseSuccess(c, `释放成功`, ``)
	return
}

func SmartLinkDownloadPath(c *gin.Context) {
	if !ensureSmartLinkNodeInstalled(c, nil) {
		return
	}
	err := component.PlaywrightClient.SmartLinkDownloadPath()
	if err != nil {
		gsgin.GinResponseError(c, fmt.Sprintf(`打开失败 %s`, err.Error()), nil)
		return
	}
	gsgin.GinResponseSuccess(c, `打开成功`, ``)
	return
}

func SmartLinkOpenDataDir(c *gin.Context) {
	err := p_common.TOsClient.OpenDirWindows(gstool.DirPathFormatToWindows(component.EnvClient.WebkitDataPath))
	if err != nil {
		gsgin.GinResponseError(c, fmt.Sprintf(`打开失败 %s`, err.Error()), nil)
		return
	}
	gsgin.GinResponseSuccess(c, `打开成功`, ``)
}

// SmartLinkList 获取列表
func SmartLinkList(c *gin.Context) {
	variableGroupList, _ := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeSmartLink,
	}).All()
	smartLinkList, _ := common.DbMain.Client.QueryBySql(`select * from tbl_smart_link where status = ? order by weight asc`, define.SmartLinkStatusNormal).All()
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

	groupInfo, _ := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`name`: groupName,
		`type`: define.GroupTypeAccount,
	}).One()
	if len(groupInfo) == 0 {
		return userList
	}
	groupId := cast.ToInt(groupInfo[`id`])
	accountList, _ := common.DbMain.Client.QuickQuery(`tbl_account`, `*`, map[string]any{
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
	smartLinkInfo, _ := common.DbMain.Client.QuickQuery(`tbl_smart_link`, `*`, map[string]any{
		`id`:     smartLinkId,
		`status`: define.SmartLinkStatusNormal,
	}).One()
	smartLinkProcessList, _ := common.DbMain.Client.QuickQuery(`tbl_smart_link_process`, `*`, map[string]any{
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
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `filter_uris`, `smart_link_group_id`, `links`, `is_error_continue`, `open_num`, `open_type`, `weight`, `combine_type`, `download_finds`, `auto_close_second`, `channel`, `show_cookies`, `process_id`})
	updateData[`combine_type`] = define.CombineTypeFix
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := common.DbMain.Client.QuickCreate(`tbl_smart_link`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_smart_link`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	variable, _ := common.DbMain.Client.QuickQuery(`tbl_smart_link`, `*`, map[string]any{
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
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_smart_link`, map[string]any{
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
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: cast.ToString(dataMap[`sse_distribute_id`]),
	}

	// 客户端模式下，走任务下发流程
	cfg := getSmartLinkConfig()
	if cfg.RunMode == define.SmartLinkRunModeLocalClient {
		smartLinkRunPlaywrightViaClient(c, dataMap, id, label, sse)
		return
	}

	if !ensureSmartLinkNodeInstalled(c, sse) {
		return
	}
	userName := cast.ToString(dataMap[`user_name`])
	password := cast.ToString(dataMap[`password`])
	openNum := max(1, cast.ToInt(dataMap[`open_num`]))
	openType := cast.ToInt(dataMap[`open_type`])
	replaceList := make(map[string]string)
	sse.Send(p_common.TMarkDownClient.BlockQuote(`运行,开始----------------我是分隔君`) + "\n")
	for i := 0; i < openNum; i++ {
		go func() {
			//生成一个唯一ID
			runUniqueId := p_common.TBaseClient.GetUnique(`playwright_run_`)
			streamFunc := func(name, msg string) {
				//输出到前端
				sse.Send(p_common.TMarkDownClient.Bold(label+`,`+runUniqueId) + ` ` + name + ` ` + msg + "\n")
			}
			streamFunc(`构建run_params`, `开始`)
			runParams, runParamsErr := plw.GetRunParams(id, label, userName, password, openType, openNum, replaceList)
			if runParamsErr != nil {
				streamFunc(`构建run_params`, `失败:`+runParamsErr.Error())
				return
			}
			runParams.StreamFunc = streamFunc
			streamFunc(`构建run_params`, `成功，准备打开的链接：`+runParams.Link+`,链接类型：`+runParams.LinkIdLabel)
			streamFunc(`打开浏览器实例`, `开始`)
			p := plw.NewPlaywright(runParams, component.PlaywrightClient.Log)
			openErr := p.Open(common.GetCall(), nil)
			if openErr != nil {
				streamFunc(`打开浏览器实例`, `失败：`+openErr.Error())
				return
			}
			streamFunc(`浏览器实例执行`, `结束`)
		}()
		time.Sleep(time.Second * 2)
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// smartLinkRunPlaywrightViaClient 客户端模式下通过 WebSocket 下发任务给 Agent
func smartLinkRunPlaywrightViaClient(c *gin.Context, dataMap map[string]any, id int, label string, sse *p_sse.SseShell) {
	// 从内存检查客户端状态
	info := GlobalClientRegistry.GetLatest()
	if info == nil {
		sse.Send(`客户端未连接，无法执行任务` + "\n")
		gsgin.GinResponseError(c, "SMART_LINK_CLIENT_OFFLINE", nil)
		return
	}

	isConnected := GlobalAgentWsManager.GetConnection(info.ClientID) != nil
	if !isConnected {
		sse.Send(`客户端未连接，无法执行任务` + "\n")
		gsgin.GinResponseError(c, "SMART_LINK_CLIENT_OFFLINE", nil)
		return
	}

	cfg := getSmartLinkConfig()
	if info.ClientVersion != cfg.ClientVersion {
		sse.Send(`客户端版本不匹配，请更新客户端` + "\n")
		gsgin.GinResponseError(c, "SMART_LINK_CLIENT_VERSION_MISMATCH", nil)
		return
	}

	if info.Status == define.SmartLinkClientStatusPreparingRuntime {
		sse.Send(`客户端正在准备运行环境，请稍后重试` + "\n")
		gsgin.GinResponseError(c, "SMART_LINK_CLIENT_PREPARING_RUNTIME", nil)
		return
	}

	clientID := info.ClientID
	userName := cast.ToString(dataMap[`user_name`])
	password := cast.ToString(dataMap[`password`])
	openType := cast.ToInt(dataMap[`open_type`])
	openNum := cast.ToInt(dataMap[`open_num`])
	replaceList := make(map[string]string)

	// 构建运行参数
	sse.Send(p_common.TMarkDownClient.BlockQuote(`运行,开始----------------我是分隔君`) + "\n")
	sse.Send(p_common.TMarkDownClient.Bold(label) + ` 构建run_params 开始` + "\n")

	runParams, runParamsErr := plw.GetRunParams(id, label, userName, password, openType, openNum, replaceList)
	if runParamsErr != nil {
		sse.Send(p_common.TMarkDownClient.Bold(label) + ` 构建run_params 失败:` + runParamsErr.Error() + "\n")
		gsgin.GinResponseError(c, `构建运行参数失败: `+runParamsErr.Error(), nil)
		return
	}

	sse.Send(p_common.TMarkDownClient.Bold(label) + ` 构建run_params 成功，准备下发给客户端执行` + "\n")

	// 生成任务 ID 和 SSE 分发 ID
	now := time.Now().Unix()
	taskID := "task_" + cast.ToString(now) + "_" + cast.ToString(id)
	sseDistributeId := cast.ToString(dataMap[`sse_distribute_id`])
	if sseDistributeId == `` {
		sseDistributeId = "smart_link_run_" + cast.ToString(now)
	}

	// 创建任务记录到数据库（用于状态追踪）
	_, createErr := common.DbMain.Client.QuickCreate("tbl_smart_link_task", map[string]any{
		"task_id":       taskID,
		"client_id":     clientID,
		"smart_link_id": id,
		"label":         label,
		"status":        define.SmartLinkTaskStatusPending,
		"run_mode":      define.SmartLinkRunModeLocalClient,
		"create_time":   now,
		"update_time":   now,
	}).Exec()
	if createErr != nil {
		sse.Send(`创建任务记录失败: ` + createErr.Error() + "\n")
		gsgin.GinResponseError(c, `创建任务失败: `+createErr.Error(), nil)
		return
	}

	// 通过 WebSocket 下发任务给 Agent
	agentRunParams := BuildAgentRunParams(runParams)
	wsMsg := define.AgentWsMessage{
		Type:            define.AgentWsMsgTaskExecute,
		ClientID:        clientID,
		TaskID:          taskID,
		SseDistributeId: sseDistributeId,
		Data: define.AgentTaskExecuteData{
			TaskID:          taskID,
			SseDistributeId: sseDistributeId,
			ClientID:        clientID,
			RunParams:       agentRunParams,
		},
	}

	if sendErr := GlobalAgentWsManager.Send(clientID, wsMsg); sendErr != nil {
		sse.Send(`下发任务到客户端失败: ` + sendErr.Error() + "\n")
		gsgin.GinResponseError(c, `下发任务到Agent失败: `+sendErr.Error(), nil)
		return
	}

	sse.Send(p_common.TMarkDownClient.Bold(label) + ` 任务已下发给客户端，等待执行` + "\n")
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		"task_id":           taskID,
		"client_id":         clientID,
		"status":            define.SmartLinkTaskStatusPending,
		"sse_distribute_id": sseDistributeId,
	})
}

// SmartLinkRunPlaywrightList 获取运行的列表
func SmartLinkRunPlaywrightList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if !ensureSmartLinkNodeInstalled(c, nil) {
		return
	}
	contextPageList := plw.NewContextList(component.PlaywrightClient.Log)
	runList := contextPageList.GetPlaywrightRunList()
	gsgin.GinResponseSuccess(c, ``, runList)
}

func SmartLinkPlaywrightVersion(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: cast.ToString(dataMap[`sse_distribute_id`]),
	}
	if !ensureSmartLinkNodeInstalled(c, sse) {
		return
	}
	sse.Send(`获取核心版本` + "\n")
	pw, pwErr := component.PlaywrightClient.SmartLinkPlaywrightVersion()
	if pwErr != nil {
		sse.Send(`获取核心版本失败` + pwErr.Error() + "\n")
		gsgin.GinResponseError(c, `查询失败`+pwErr.Error(), nil)
		return
	}
	//是否在安装中
	isInstall := 0
	if component.PlaywrightClient.IsInstalling() {
		sse.Send(`核心正在安装中` + "\n")
		isInstall = 1
	} else {
		sse.Send(`当前未处于安装中` + "\n")
	}
	gsgin.GinResponseSuccess(c, pw.Version, map[string]any{
		`is_install`: isInstall,
		`version`:    pw.Version,
	})
}

// SmartProcessList 获取列表
func SmartProcessList(c *gin.Context) {
	list, _ := common.DbMain.Client.QueryBySql(`select * from tbl_smart_link_process where status = 1  order by id desc`).All()
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
		newId, createErr := common.DbMain.Client.QuickCreate(`tbl_smart_link_process`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_smart_link_process`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := common.DbMain.Client.QuickQuery(`tbl_smart_link_process`, `*`, map[string]any{
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
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_smart_link_process`, map[string]any{
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
	list, _ := common.DbMain.Client.QueryBySql(`
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
	updateData := gstool.MapTakeKeys(&dataMap, []string{`name`, `wait_mills`, `is_async`, `is_error_continue`, `append_to_replace`, `smart_link_process_id`, `type`, `locator`, `tip`, `value`, `out_key`, `check_key`, `weight`, `domain_limit`, `x`, `y`, `next_ids`})
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := common.DbMain.Client.QuickCreate(`tbl_smart_link_process_item`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_smart_link_process_item`,
			map[string]any{
				`id`: dataMap[`id`],
			}, updateData).Exec()
		id = dataMap[`id`]
	}
	info, _ := common.DbMain.Client.QuickQuery(`tbl_smart_link_process_item`, `*`, map[string]any{
		`id`:     id,
		`status`: 1,
	}).One()
	gsgin.GinResponseSuccess(c, ``, info)
}

// SmartProcessCancelRelation 移除连线
func SmartProcessCancelRelation(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	var prevId = cast.ToInt(dataMap[`prev_id`])
	var nextId = cast.ToString(dataMap[`next_id`])
	if prevId == 0 || nextId == `` {
		gsgin.GinResponseError(c, `节点不能为空 `, nil)
		return
	}
	info, _ := common.DbMain.Client.QuickQuery(`tbl_smart_link_process_item`, `*`, map[string]any{
		`id`:     prevId,
		`status`: 1,
	}).One()
	if len(info) == 0 {
		gsgin.GinResponseError(c, `节点不存在`, nil)
		return
	}
	nextIds := cast.ToString(info[`next_ids`])
	nextIdList := strings.Split(nextIds, `,`)
	gstool.ArrayDeleteValue(&nextIdList, nextId)
	updateData := make(map[string]any)
	updateData[`next_ids`] = strings.Join(nextIdList, `,`)
	_, _ = common.DbMain.Client.QuickUpdate(`tbl_smart_link_process_item`,
		map[string]any{
			`id`: prevId,
		}, updateData).Exec()

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
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_smart_link_process_item`, map[string]any{
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
		_, _ = common.DbMain.Client.QuickUpdate(`tbl_smart_link_process_item`, map[string]any{
			`id`: cast.ToInt(item),
		}, map[string]interface{}{
			`weight`: index + 1,
		}).Exec()
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SmartProcessSetPosition(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	smartLinkProcessId := cast.ToInt(dataMap[`id`])
	if smartLinkProcessId == 0 {
		gsgin.GinResponseError(c, `smart_link_process_id不能为空`, nil)
		return
	}
	x := cast.ToInt(dataMap[`x`])
	y := cast.ToInt(dataMap[`y`])
	_, _ = common.DbMain.Client.QuickUpdate(`tbl_smart_link_process_item`, map[string]any{
		`id`: cast.ToInt(smartLinkProcessId),
	}, map[string]interface{}{
		`x`: x,
		`y`: y,
	}).Exec()
	gsgin.GinResponseSuccess(c, ``, nil)
}

func SmartProcessSetRelation(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	prevId := cast.ToInt(dataMap[`prev_id`])
	nextId := cast.ToInt(dataMap[`next_id`])
	if prevId == 0 || nextId == 0 {
		gsgin.GinResponseError(c, `prev_id或next_id不能为空`, nil)
		return
	}
	info, err := common.DbMain.Client.QuickQuery(`tbl_smart_link_process_item`, `*`, map[string]any{
		`id`:     prevId,
		`status`: 1,
	}).One()
	if err != nil {
		gsgin.GinResponseError(c, `prev_id不存在`, nil)
		return
	}
	nextIds := cast.ToString(info[`next_ids`])
	nextIdList := strings.Split(nextIds, `,`)
	for _, item := range nextIdList {
		if item == cast.ToString(nextId) {
			gsgin.GinResponseError(c, `next_id已存在`, nil)
			return
		}
	}
	nextIdList = append(nextIdList, cast.ToString(nextId))
	newNextIdList := make([]string, 0)
	for _, item := range nextIdList {
		if cast.ToInt(item) == 0 {
			continue
		}
		newNextIdList = append(newNextIdList, item)
	}
	_, _ = common.DbMain.Client.QuickUpdate(`tbl_smart_link_process_item`, map[string]any{
		`id`: cast.ToInt(prevId),
	}, map[string]interface{}{
		`next_ids`: strings.Join(newNextIdList, `,`),
	}).Exec()
	gsgin.GinResponseSuccess(c, ``, nil)
}
