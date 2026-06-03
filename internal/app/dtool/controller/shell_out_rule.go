package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
)

// ShellOutRuleSetList 返回全部规则集。 // ShellOutRuleSetList returns all shell-out rule sets.
func ShellOutRuleSetList(c *gin.Context) {
	list, err := common.DbMain.ShellOutRuleSetList()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, list)
}

// ShellOutRuleSetInfo 返回规则集详情与规则项。 // ShellOutRuleSetInfo returns one rule set and its nested rule items.
func ShellOutRuleSetInfo(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	info, items, err := common.DbMain.ShellOutRuleSetInfo(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`rule_set`:   info,
		`rule_items`: items,
	})
}

// ShellOutRuleSetSave 保存规则集与其规则项快照，并刷新运行中客户端的规则。 // ShellOutRuleSetSave saves one rule set plus its child rule items snapshot, then refreshes running clients.
func ShellOutRuleSetSave(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)

	ruleSetID, err := common.DbMain.ShellOutRuleSetSave(map[string]any{
		`id`:          cast.ToInt(dataMap[`id`]),
		`name`:        cast.ToString(dataMap[`name`]),
		`description`: cast.ToString(dataMap[`description`]),
		`is_enabled`:  cast.ToInt(dataMap[`is_enabled`]),
		`match_mode`:  cast.ToString(dataMap[`match_mode`]),
		`update_time`: time.Now().Unix(),
	}, normalizeShellOutRuleItemsPayload(dataMap[`rule_items`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	// 刷新使用该规则集的所有运行中客户端的规则项
	component.ShellOutClient.RefreshRuleItemsByRuleSetId(ruleSetID)

	info, items, err := common.DbMain.ShellOutRuleSetInfo(ruleSetID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`id`:         ruleSetID,
		`rule_set`:   info,
		`rule_items`: items,
	})
}

// ShellOutRuleSetDelete 删除规则集。 // ShellOutRuleSetDelete removes a shell-out rule set and its child items.
func ShellOutRuleSetDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if err := common.DbMain.ShellOutRuleSetDelete(cast.ToInt(dataMap[`id`])); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// ShellOutRuleImportLegacy 把旧分组上的正则配置迁移成新规则集。 // ShellOutRuleImportLegacy migrates legacy group regex fields into the new rule-center tables.
func ShellOutRuleImportLegacy(c *gin.Context) {
	result, err := common.DbMain.ImportLegacyShellOutGroupRules()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, result)
}

func normalizeShellOutRuleItemsPayload(raw any) []map[string]any {
	if raw == nil {
		return []map[string]any{}
	}
	rawList, ok := raw.([]any)
	if !ok {
		typedList, ok := raw.([]map[string]any)
		if ok {
			return typedList
		}
		return []map[string]any{}
	}
	result := make([]map[string]any, 0, len(rawList))
	for _, item := range rawList {
		rowMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		result = append(result, rowMap)
	}
	return result
}
