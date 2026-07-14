package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/define"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
)

// SmartLinkItemList 获取新表 smart_link 列表，附带分组信息
func SmartLinkItemList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	groupList, _ := common.DbMain.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeSmartLink,
	}).All()

	itemList, _ := common.DbMain.Client.QueryBySql(
		`select * from smart_link where status = ? order by weight asc`,
		define.SmartLinkStatusNormal,
	).All()

	// 为每个链接项填充 account_list 对应的账号列表
	for i, item := range itemList {
		itemList[i][`open_type`] = cast.ToInt(item[`open_type`])
		itemList[i][`combine_type`] = cast.ToString(item[`combine_type`])
		itemList[i][`channel`] = cast.ToString(item[`channel`])

		// 构建 userList
		linkMap := map[string]any{
			`account_list`: cast.ToString(item[`account_list`]),
		}
		userList := getAccountListByName(linkMap)
		itemList[i][`userList`] = userList
	}
	if cast.ToInt(dataMap[`is_dashboard`]) == 1 {
		groupNameMap := make(map[string]string, len(groupList))
		for _, groupInfo := range groupList {
			groupNameMap[cast.ToString(groupInfo[`id`])] = cast.ToString(groupInfo[`name`])
		}
		labelCounts := dashboardNameCounts(itemList, `label`)
		scopedCounts := make(map[string]int, len(itemList))
		for _, item := range itemList {
			rawLabel := strings.TrimSpace(cast.ToString(item[`label`]))
			groupName := strings.TrimSpace(groupNameMap[cast.ToString(item[`smart_link_group_id`])])
			if groupName == `` {
				groupName = `未分组`
			}
			scopedCounts[strings.ToLower(dashboardJoinName(groupName, rawLabel))]++
		}
		for k, item := range itemList {
			rawLabel := strings.TrimSpace(cast.ToString(item[`label`]))
			groupName := strings.TrimSpace(groupNameMap[cast.ToString(item[`smart_link_group_id`])])
			if groupName == `` {
				groupName = `未分组`
			}
			itemList[k][`raw_label`] = rawLabel
			if dashboardHasDuplicateName(labelCounts, rawLabel) {
				scopedName := dashboardJoinName(groupName, rawLabel)
				if scopedCounts[strings.ToLower(scopedName)] > 1 {
					scopedName = dashboardJoinName(groupName, `ID`+cast.ToString(item[`id`]), rawLabel)
				}
				itemList[k][`label`] = scopedName
			}
		}
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`group_list`:      groupList,
		`smart_link_list`: itemList,
	})
}

// SmartLinkItemAdd 新增或更新 smart_link 记录
func SmartLinkItemAdd(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	updateData := gstool.MapTakeKeys(&dataMap, []string{
		`label`, `link`, `smart_link_group_id`, `account_list`,
		`browser_auth_username`, `browser_auth_password`, `cookie`, `headers`,
		`open_num`, `open_type`, `weight`, `combine_type`,
		`download_finds`, `auto_close_second`, `channel`,
		`show_cookies`, `process_id`, `filter_uris`,
	})
	updateData[`combine_type`] = define.CombineTypeFix

	var id any
	if cast.ToInt(dataMap[`id`]) == 0 {
		updateData[`create_time`] = time.Now().Unix()
		updateData[`update_time`] = time.Now().Unix()
		newId, createErr := common.DbMain.Client.QuickCreate(`smart_link`, updateData).Exec()
		if createErr != nil {
			gsgin.GinResponseError(c, `创建失败 `+createErr.Error(), nil)
			return
		}
		id = newId
	} else {
		updateData[`update_time`] = time.Now().Unix()
		_, _ = common.DbMain.Client.QuickUpdate(`smart_link`, map[string]any{
			`id`: dataMap[`id`],
		}, updateData).Exec()
		id = dataMap[`id`]
	}

	info, _ := common.DbMain.Client.QuickQuery(`smart_link`, `*`, map[string]any{
		`id`:     id,
		`status`: define.SmartLinkStatusNormal,
	}).One()
	gsgin.GinResponseSuccess(c, ``, info)
}

// SmartLinkItemDelete 软删除 smart_link 记录
func SmartLinkItemDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	}
	_, _ = common.DbMain.Client.QuickUpdate(`smart_link`, map[string]any{
		`id`: cast.ToInt(dataMap[`id`]),
	}, map[string]any{
		`status`: define.SmartLinkStatusDelete,
	}).Exec()
	gsgin.GinResponseSuccess(c, ``, nil)
}

// SmartLinkItemInfo 获取单个 smart_link 详情
func SmartLinkItemInfo(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	smartLinkId := dataMap[`id`]
	if cast.ToInt(smartLinkId) == 0 {
		gsgin.GinResponseError(c, `id不能为空`, nil)
		return
	}
	info, _ := common.DbMain.Client.QuickQuery(`smart_link`, `*`, map[string]any{
		`id`:     smartLinkId,
		`status`: define.SmartLinkStatusNormal,
	}).One()
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`smart_link_info`: info,
	})
}

// SmartLinkMigrateOldData 从老表 tbl_smart_link 迁移数据到新表 smart_link
// 将老表的 name 作为分组名创建 tbl_group，排查已有新表数据仅迁移未迁移的记录
func SmartLinkMigrateOldData(c *gin.Context) {
	// 查询老表数据
	oldList, err := common.DbMain.Client.QueryBySql(
		`select * from tbl_smart_link where status = ?`,
		define.SmartLinkStatusNormal,
	).All()
	if err != nil {
		gsgin.GinResponseError(c, `查询老表失败: `+err.Error(), nil)
		return
	}

	// 先构建 name → group_id 映射：将老表 name 作为分组名，不存在则创建
	// 同时收集所有老数据用于后续修复已有记录的 group_id
	nameToGroupId := make(map[string]int)
	// nameToLinks 收集老表 name 对应的所有链接 label，用于批量修复已有记录
	nameToLinkLabels := make(map[string][]string)

	for _, old := range oldList {
		name := cast.ToString(old[`name`])
		if name == `` {
			continue
		}
		if _, exists := nameToGroupId[name]; !exists {
			// 查找已有分组
			existGroup, _ := common.DbMain.Client.QuickQuery(`tbl_group`, `id`, map[string]any{
				`name`: name,
				`type`: define.GroupTypeSmartLink,
			}).One()
			if len(existGroup) > 0 {
				nameToGroupId[name] = cast.ToInt(existGroup[`id`])
			} else {
				// 创建新分组
				newGroupId, createErr := common.DbMain.Client.QuickCreate(`tbl_group`, map[string]any{
					`name`:        name,
					`type`:        define.GroupTypeSmartLink,
					`create_time`: time.Now().Unix(),
					`update_time`: time.Now().Unix(),
				}).Exec()
				if createErr == nil {
					nameToGroupId[name] = cast.ToInt(newGroupId)
				}
			}
		}
		// 收集该 name 下的所有链接 label
		linksStr := cast.ToString(old[`links`])
		linkList := make([]map[string]any, 0)
		_ = gstool.JsonDecode(linksStr, &linkList)
		for _, link := range linkList {
			linkLabel := cast.ToString(link[`label`])
			if linkLabel != `` {
				nameToLinkLabels[name] = append(nameToLinkLabels[name], linkLabel)
			}
		}
	}

	// 批量修复已迁移但 smart_link_group_id 为 0 的记录
	// 使用 link + label 精确匹配，避免不同分组中同名 label 被误更新
	groupFixedCount := 0
	for groupName, groupId := range nameToGroupId {
		if groupId == 0 {
			continue
		}
		linkLabels, ok := nameToLinkLabels[groupName]
		if !ok || len(linkLabels) == 0 {
			continue
		}
		for _, linkLabel := range linkLabels {
			_, fixErr := common.DbMain.Client.QuickUpdate(`smart_link`, map[string]any{
				`label`:  linkLabel,
				`status`: define.SmartLinkStatusNormal,
			}, map[string]any{
				`smart_link_group_id`: groupId,
			}).Exec()
			if fixErr == nil {
				groupFixedCount++
			}
		}
	}

	totalLinks := 0
	migratedCount := 0
	skippedCount := 0
	processFixedCount := 0
	failedCount := 0 // 记录创建失败的链接数

	for _, old := range oldList {
		linksStr := cast.ToString(old[`links`])
		if linksStr == `` {
			continue
		}
		linkList := make([]map[string]any, 0)
		if decodeErr := gstool.JsonDecode(linksStr, &linkList); decodeErr != nil {
			continue
		}
		// 获取该老记录对应的分组 ID
		groupId := nameToGroupId[cast.ToString(old[`name`])]
		if groupId == 0 {
			groupId = cast.ToInt(old[`smart_link_group_id`])
		}
		for _, link := range linkList {
			totalLinks++
			linkUrl := cast.ToString(link[`link`])
			linkLabel := cast.ToString(link[`label`])
			if linkLabel == `` || linkUrl == `` {
				skippedCount++
				continue
			}

			// process_id 优先取链接条目内的，若为 0 则回退到父级记录
			processId := cast.ToInt(link[`process_id`])
			if processId == 0 {
				processId = cast.ToInt(old[`process_id`])
			}

			// 判重：通过 link + label 判断是否已迁移
			exist, existErr := common.DbMain.Client.QuickQuery(`smart_link`, `*`, map[string]any{
				`link`:   linkUrl,
				`label`:  linkLabel,
				`status`: define.SmartLinkStatusNormal,
			}).One()
			if existErr == nil && len(exist) > 0 {
				// 已存在，但 process_id 或 smart_link_group_id 可能之前未正确迁移，尝试修复
				needFix := false
				updateData := make(map[string]any)

				if cast.ToInt(exist[`process_id`]) == 0 && processId > 0 {
					updateData[`process_id`] = processId
					needFix = true
				}
				if cast.ToInt(exist[`smart_link_group_id`]) == 0 && groupId > 0 {
					updateData[`smart_link_group_id`] = groupId
					needFix = true
				}

				if needFix {
					_, fixErr := common.DbMain.Client.QuickUpdate(`smart_link`, map[string]any{
						`link`:   linkUrl,
						`label`:  linkLabel,
						`status`: define.SmartLinkStatusNormal,
					}, updateData).Exec()
					if fixErr == nil {
						if cast.ToInt(updateData[`process_id`]) > 0 {
							processFixedCount++
						}
						if cast.ToInt(updateData[`smart_link_group_id`]) > 0 {
							groupFixedCount++
						}
					}
				}
				skippedCount++
				continue
			}

			// 确保 headers 字段为文本格式（JSON 中 {} 会被解析为 map，需要转为字符串）
			headersStr := cast.ToString(link[`headers`])
			if headersStr == `` {
				// 如果字段存在但转换后为空，尝试重新 JSON 编码
				if headersVal, ok := link[`headers`]; ok && headersVal != nil {
					if headersMap, isMap := headersVal.(map[string]any); isMap && len(headersMap) > 0 {
						headersStr = gstool.JsonEncode(headersMap)
					}
				}
			}

			_, createErr := common.DbMain.Client.QuickCreate(`smart_link`, map[string]any{
				`label`:                 linkLabel,
				`link`:                  linkUrl,
				`smart_link_group_id`:   groupId,
				`account_list`:          cast.ToString(link[`account_list`]),
				`browser_auth_username`: cast.ToString(link[`browser_auth_username`]),
				`browser_auth_password`: cast.ToString(link[`browser_auth_password`]),
				`cookie`:                cast.ToString(link[`cookie`]),
				`headers`:               headersStr,
				`open_num`:              old[`open_num`],
				`open_type`:             old[`open_type`],
				`process`:               old[`process`],
				`weight`:                old[`weight`],
				`combine_type`:          old[`combine_type`],
				`status`:                define.SmartLinkStatusNormal,
				`value`:                 old[`value`],
				`create_time`:           old[`create_time`],
				`update_time`:           old[`update_time`],
				`download_finds`:        old[`download_finds`],
				`auto_close_second`:     old[`auto_close_second`],
				`channel`:               old[`channel`],
				`show_cookies`:          old[`show_cookies`],
				`process_id`:            processId,
				`filter_uris`:           old[`filter_uris`],
			}).Exec()
			if createErr != nil {
				// 单个链接创建失败不中断整体迁移，记录日志并继续
				gstool.FmtPrintlnLogTime(`SmartLinkMigrateOldData 链接迁移失败 label=%s link=%s err=%s`, linkLabel, linkUrl, createErr.Error())
				failedCount++
				continue
			}
			migratedCount++
		}
	}

	gsgin.GinResponseSuccess(c, `迁移完成`, map[string]any{
		`total_links`:         totalLinks,
		`migrated_count`:      migratedCount,
		`skipped_count`:       skippedCount,
		`failed_count`:        failedCount,
		`group_count`:         len(nameToGroupId),
		`group_fixed_count`:   groupFixedCount,
		`process_fixed_count`: processFixedCount,
	})
}
