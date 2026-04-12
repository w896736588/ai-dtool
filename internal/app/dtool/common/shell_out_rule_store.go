package common

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"dev_tool/internal/app/dtool/define"

	"github.com/spf13/cast"
)

const (
	// ShellOutRuleMatchModeLine 表示第一版规则按单行日志执行。 // ShellOutRuleMatchModeLine means rule evaluation runs line-by-line in v1.
	ShellOutRuleMatchModeLine = `line`
	// ShellOutLegacyImportRuleSetNamePrefix 标记旧分组导入后的规则集名称前缀。 // ShellOutLegacyImportRuleSetNamePrefix marks rule sets imported from legacy group regex fields.
	ShellOutLegacyImportRuleSetNamePrefix = `旧分组迁移 - `
)

// ShellOutRuleSetList 返回规则集列表，按最新优先排序。 // ShellOutRuleSetList returns rule sets ordered by newest first.
func (h *CSqlite) ShellOutRuleSetList() ([]map[string]any, error) {
	return h.Client.QueryBySql(`
select rs.*,
       (
         select count(1)
         from tbl_shell_out_rule_item ri
         where ri.rule_set_id = rs.id
       ) as rule_item_count
from tbl_shell_out_rule_set rs
order by rs.id desc`).All()
}

// ShellOutRuleSetInfo 返回规则集详情和其下全部规则项。 // ShellOutRuleSetInfo returns one rule set and all of its child items.
func (h *CSqlite) ShellOutRuleSetInfo(id int) (map[string]any, []map[string]any, error) {
	if id <= 0 {
		return nil, nil, errors.New(`规则集id不能为空`)
	}
	info, err := h.Client.QuickQuery(`tbl_shell_out_rule_set`, `*`, map[string]any{
		`id`: id,
	}).One()
	if err != nil {
		return nil, nil, err
	}
	if len(info) == 0 {
		return nil, nil, errors.New(`规则集不存在`)
	}
	items, err := h.Client.QueryBySql(`
select *
from tbl_shell_out_rule_item
where rule_set_id = ?
order by priority asc, id asc`, id).All()
	if err != nil {
		return nil, nil, err
	}
	return info, items, nil
}

// ShellOutRuleSetSave 保存规则集并整体替换其子规则。 // ShellOutRuleSetSave saves one rule set and replaces all nested rule items.
func (h *CSqlite) ShellOutRuleSetSave(ruleSet map[string]any, ruleItems []map[string]any) (int, error) {
	now := time.Now().Unix()
	name := strings.TrimSpace(cast.ToString(ruleSet[`name`]))
	if name == `` {
		return 0, errors.New(`规则集名称不能为空`)
	}

	ruleSetID := cast.ToInt(ruleSet[`id`])
	matchMode := strings.TrimSpace(cast.ToString(ruleSet[`match_mode`]))
	if matchMode == `` {
		matchMode = ShellOutRuleMatchModeLine
	}

	ruleSetData := map[string]any{
		`name`:        name,
		`description`: strings.TrimSpace(cast.ToString(ruleSet[`description`])),
		`is_enabled`:  cast.ToInt(ruleSet[`is_enabled`]),
		`match_mode`:  matchMode,
		`update_time`: now,
	}

	if ruleSetID <= 0 {
		ruleSetData[`create_time`] = now
		newID, err := h.Client.QuickCreate(`tbl_shell_out_rule_set`, ruleSetData).Exec()
		if err != nil {
			return 0, err
		}
		ruleSetID = cast.ToInt(newID)
	} else {
		if _, _, err := h.ShellOutRuleSetInfo(ruleSetID); err != nil {
			return 0, err
		}
		if _, err := h.Client.QuickUpdate(`tbl_shell_out_rule_set`, map[string]any{
			`id`: ruleSetID,
		}, ruleSetData).Exec(); err != nil {
			return 0, err
		}
	}

	// 子规则改为整套替换，避免前端局部删改时残留旧记录。 // Replace child rules as a full snapshot so removed items do not linger.
	if _, err := h.Client.QuickDelete(`tbl_shell_out_rule_item`, map[string]any{
		`rule_set_id`: ruleSetID,
	}).Exec(); err != nil {
		return 0, err
	}

	for _, item := range ruleItems {
		itemName := strings.TrimSpace(cast.ToString(item[`name`]))
		if itemName == `` {
			return 0, errors.New(`规则项名称不能为空`)
		}
		if _, err := h.Client.QuickCreate(`tbl_shell_out_rule_item`, map[string]any{
			`rule_set_id`:     ruleSetID,
			`name`:            itemName,
			`rule_type`:       strings.TrimSpace(cast.ToString(item[`rule_type`])),
			`match_type`:      strings.TrimSpace(cast.ToString(item[`match_type`])),
			`pattern`:         cast.ToString(item[`pattern`]),
			`exclude_pattern`: cast.ToString(item[`exclude_pattern`]),
			`priority`:        cast.ToInt(item[`priority`]),
			`is_enabled`:      cast.ToInt(item[`is_enabled`]),
			`stop_on_match`:   cast.ToInt(item[`stop_on_match`]),
			`config_json`:     strings.TrimSpace(cast.ToString(item[`config_json`])),
			`create_time`:     now,
			`update_time`:     now,
		}).Exec(); err != nil {
			return 0, err
		}
	}

	return ruleSetID, nil
}

// ShellOutRuleSetDelete 删除规则集及其子规则。 // ShellOutRuleSetDelete removes one rule set and all nested rule items.
func (h *CSqlite) ShellOutRuleSetDelete(id int) error {
	if id <= 0 {
		return errors.New(`规则集id不能为空`)
	}
	if _, _, err := h.ShellOutRuleSetInfo(id); err != nil {
		return err
	}
	if _, err := h.Client.QuickDelete(`tbl_shell_out_rule_item`, map[string]any{
		`rule_set_id`: id,
	}).Exec(); err != nil {
		return err
	}
	_, err := h.Client.QuickDelete(`tbl_shell_out_rule_set`, map[string]any{
		`id`: id,
	}).Exec()
	return err
}

// ShellOutRuleItemsEnabled 返回启用中的规则项，按优先级排序。 // ShellOutRuleItemsEnabled returns enabled child rules ordered by priority.
func (h *CSqlite) ShellOutRuleItemsEnabled(ruleSetID int) ([]map[string]any, error) {
	if ruleSetID <= 0 {
		return []map[string]any{}, nil
	}
	return h.Client.QueryBySql(`
select *
from tbl_shell_out_rule_item
where rule_set_id = ?
  and is_enabled = 1
order by priority asc, id asc`, ruleSetID).All()
}

// ImportLegacyShellOutGroupRules 把旧分组中的正则配置迁移为新规则集并绑定终端输出。 // ImportLegacyShellOutGroupRules migrates legacy group regex fields into new rule sets and binds shell-out records.
func (h *CSqlite) ImportLegacyShellOutGroupRules() (map[string]any, error) {
	groupList, err := h.Client.QuickQuery(`tbl_group`, `*`, map[string]any{
		`type`: define.GroupTypeShellOut,
	}).Order(`id asc`).All()
	if err != nil {
		return nil, err
	}
	result := map[string]any{
		`group_count`:              len(groupList),
		`imported_rule_set_count`:  0,
		`imported_rule_item_count`: 0,
		`bound_shell_out_count`:    0,
		`skipped_group_count`:      0,
	}
	for _, groupItem := range groupList {
		groupID := cast.ToInt(groupItem[`id`])
		groupName := strings.TrimSpace(cast.ToString(groupItem[`name`]))
		filterRules := parseLegacyRuleLines(cast.ToString(groupItem[`extra_1`]))
		alertRules := parseLegacyRuleLines(cast.ToString(groupItem[`extra_2`]))
		noErrorLines := parseLegacyPlainLines(cast.ToString(groupItem[`extra_3`]))
		if len(filterRules) == 0 && len(alertRules) == 0 {
			result[`skipped_group_count`] = cast.ToInt(result[`skipped_group_count`]) + 1
			continue
		}

		ruleSetName := fmt.Sprintf(`%s%s`, ShellOutLegacyImportRuleSetNamePrefix, groupName)
		ruleSetID, findErr := h.findLegacyImportedRuleSetID(ruleSetName)
		if findErr != nil {
			return nil, findErr
		}

		ruleItems := buildLegacyRuleItems(filterRules, alertRules, noErrorLines)
		savedRuleSetID, saveErr := h.ShellOutRuleSetSave(map[string]any{
			`id`:          ruleSetID,
			`name`:        ruleSetName,
			`description`: fmt.Sprintf(`由终端输出分组「%s」中的旧正则配置自动导入`, groupName),
			`is_enabled`:  1,
			`match_mode`:  ShellOutRuleMatchModeLine,
		}, ruleItems)
		if saveErr != nil {
			return nil, saveErr
		}

		boundShellOutList, queryBindErr := h.Client.QuickQuery(`tbl_shell_out`, `id`, map[string]any{
			`group_id`: groupID,
		}).All()
		if queryBindErr != nil {
			return nil, queryBindErr
		}
		if _, updateErr := h.Client.QuickUpdate(`tbl_shell_out`, map[string]any{
			`group_id`: groupID,
		}, map[string]any{
			`rule_set_id`: savedRuleSetID,
			`update_time`: time.Now().Unix(),
		}).Exec(); updateErr != nil {
			return nil, updateErr
		}
		result[`imported_rule_set_count`] = cast.ToInt(result[`imported_rule_set_count`]) + 1
		result[`imported_rule_item_count`] = cast.ToInt(result[`imported_rule_item_count`]) + len(ruleItems)
		result[`bound_shell_out_count`] = cast.ToInt(result[`bound_shell_out_count`]) + len(boundShellOutList)
	}
	return result, nil
}

type legacyRuleLine struct {
	Name    string
	Pattern string
}

func parseLegacyRuleLines(raw string) []legacyRuleLine {
	lines := strings.Split(raw, "\n")
	result := make([]legacyRuleLine, 0, len(lines))
	for _, line := range lines {
		currentLine := strings.TrimSpace(line)
		if currentLine == `` {
			continue
		}
		name := ``
		pattern := currentLine
		params := strings.SplitN(currentLine, `#`, 2)
		if len(params) == 2 {
			name = strings.TrimSpace(params[0])
			pattern = strings.TrimSpace(params[1])
		}
		if pattern == `` {
			continue
		}
		result = append(result, legacyRuleLine{
			Name:    name,
			Pattern: pattern,
		})
	}
	return result
}

func parseLegacyPlainLines(raw string) []string {
	lines := strings.Split(raw, "\n")
	result := make([]string, 0, len(lines))
	for _, line := range lines {
		currentLine := strings.TrimSpace(line)
		if currentLine == `` {
			continue
		}
		result = append(result, currentLine)
	}
	return result
}

func buildLegacyRuleItems(filterRules, alertRules []legacyRuleLine, noErrorLines []string) []map[string]any {
	result := make([]map[string]any, 0, len(filterRules)+len(alertRules))
	priority := 0
	for _, ruleItem := range filterRules {
		ruleName := strings.TrimSpace(ruleItem.Name)
		if ruleName == `` {
			ruleName = ruleItem.Pattern
		}
		result = append(result, map[string]any{
			`name`:            ruleName,
			`rule_type`:       ShellOutRuleTypeDrop,
			`match_type`:      ShellOutRuleMatchTypeRegex,
			`pattern`:         ruleItem.Pattern,
			`exclude_pattern`: ``,
			`priority`:        priority,
			`is_enabled`:      1,
			`stop_on_match`:   1,
			`config_json`:     `{}`,
		})
		priority += 10
	}

	excludePattern := buildLegacyExcludePattern(noErrorLines)
	alertPriority := 1000
	for _, ruleItem := range alertRules {
		ruleName := strings.TrimSpace(ruleItem.Name)
		if ruleName == `` {
			ruleName = ruleItem.Pattern
		}
		result = append(result, map[string]any{
			`name`:            ruleName,
			`rule_type`:       ShellOutRuleTypeAlert,
			`match_type`:      ShellOutRuleMatchTypeRegex,
			`pattern`:         ruleItem.Pattern,
			`exclude_pattern`: excludePattern,
			`priority`:        alertPriority,
			`is_enabled`:      1,
			`stop_on_match`:   0,
			`config_json`:     `{"level":"warning"}`,
		})
		alertPriority += 10
	}
	return result
}

func buildLegacyExcludePattern(lines []string) string {
	if len(lines) == 0 {
		return ``
	}
	parts := make([]string, 0, len(lines))
	for _, line := range lines {
		if strings.TrimSpace(line) == `` {
			continue
		}
		parts = append(parts, regexp.QuoteMeta(strings.TrimSpace(line)))
	}
	if len(parts) == 0 {
		return ``
	}
	return strings.Join(parts, `|`)
}

func (h *CSqlite) findLegacyImportedRuleSetID(ruleSetName string) (int, error) {
	ruleSetInfo, err := h.Client.QuickQuery(`tbl_shell_out_rule_set`, `*`, map[string]any{
		`name`: ruleSetName,
	}).One()
	if err != nil {
		return 0, err
	}
	return cast.ToInt(ruleSetInfo[`id`]), nil
}
