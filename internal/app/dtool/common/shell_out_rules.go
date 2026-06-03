package common

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gstool"
)

const (
	// ShellOutRuleTypeDrop 表示命中后直接过滤该行。 // ShellOutRuleTypeDrop drops a matched log line from the visible stream.
	ShellOutRuleTypeDrop = `drop`
	// ShellOutRuleTypeAlert 表示命中后记录异常事件。 // ShellOutRuleTypeAlert records an alert event when the line matches.
	ShellOutRuleTypeAlert = `alert`

	// ShellOutRuleMatchTypeContains 表示按包含关系匹配。 // ShellOutRuleMatchTypeContains matches by substring containment.
	ShellOutRuleMatchTypeContains = `contains`
	// ShellOutRuleMatchTypeRegex 表示按正则匹配。 // ShellOutRuleMatchTypeRegex matches by regular expression.
	ShellOutRuleMatchTypeRegex = `regex`
)

// ShellOutRuleItem 描述运行时可执行的一条日志规则。 // ShellOutRuleItem describes one executable runtime log rule.
type ShellOutRuleItem struct {
	ID             int
	RuleSetID      int
	Name           string
	RuleType       string
	MatchType      string
	Pattern        string
	ExcludePattern string
	Priority       int
	IsEnabled      bool
	StopOnMatch    bool
	ConfigJSON     string
	Level          string
	Category       string
}

// ShellOutAlertEvent 描述命中的异常事件。 // ShellOutAlertEvent describes one matched alert event.
type ShellOutAlertEvent struct {
	RuleItemID int    `json:"rule_item_id"`
	RuleName   string `json:"rule_name"`
	Level      string `json:"level"`
	Category   string `json:"category"`
	SampleLine string `json:"sample_line"`
	LineNumber int64  `json:"line_number"`
	Time       string `json:"time"`
}

// RuleApplyResult 表示一行日志执行规则后的结果。 // RuleApplyResult describes the outcome after applying rules to one log line.
type RuleApplyResult struct {
	Dropped     bool
	MatchedRule []int
	AlertEvents []ShellOutAlertEvent
}

// LoadRuleItems 按规则集加载启用中的规则项。 // LoadRuleItems loads enabled rule items for a rule set.
func (h *TShellOut) LoadRuleItems(ruleSetID int) ([]ShellOutRuleItem, error) {
	if ruleSetID <= 0 || DbMain == nil || DbMain.Client == nil {
		return []ShellOutRuleItem{}, nil
	}
	rows, err := DbMain.ShellOutRuleItemsEnabled(ruleSetID)
	if err != nil {
		return nil, err
	}
	result := make([]ShellOutRuleItem, 0, len(rows))
	for _, row := range rows {
		result = append(result, buildShellOutRuleItem(row))
	}
	return result, nil
}

// ApplyRuleItems 依次执行当前输出绑定的规则项。 // ApplyRuleItems executes the rule items bound to the current shell output.
func (h *TShellOut) ApplyRuleItems(shellOut *ShellOut, line string, lineNumber int64) RuleApplyResult {
	result := RuleApplyResult{
		Dropped:     false,
		MatchedRule: make([]int, 0),
		AlertEvents: make([]ShellOutAlertEvent, 0),
	}
	if shellOut == nil || len(shellOut.ruleItems) == 0 {
		return result
	}
	for _, item := range shellOut.ruleItems {
		if !item.IsEnabled || !matchShellOutRuleItem(item, line) {
			continue
		}
		result.MatchedRule = append(result.MatchedRule, item.ID)
		switch item.RuleType {
		case ShellOutRuleTypeDrop:
			h.recordRuleHit(shellOut, item)
			result.Dropped = true
		case ShellOutRuleTypeAlert:
			event := h.recordAlertEvent(shellOut, item, line, lineNumber)
			result.AlertEvents = append(result.AlertEvents, event)
		}
		// 命中且要求中止时，立即结束本行后续规则。 // Stop evaluating later rules when the current match is terminal.
		if item.StopOnMatch {
			return result
		}
	}
	return result
}

func buildShellOutRuleItem(row map[string]any) ShellOutRuleItem {
	item := ShellOutRuleItem{
		ID:             cast.ToInt(row[`id`]),
		RuleSetID:      cast.ToInt(row[`rule_set_id`]),
		Name:           strings.TrimSpace(cast.ToString(row[`name`])),
		RuleType:       strings.TrimSpace(cast.ToString(row[`rule_type`])),
		MatchType:      strings.TrimSpace(cast.ToString(row[`match_type`])),
		Pattern:        cast.ToString(row[`pattern`]),
		ExcludePattern: cast.ToString(row[`exclude_pattern`]),
		Priority:       cast.ToInt(row[`priority`]),
		IsEnabled:      cast.ToInt(row[`is_enabled`]) == 1,
		StopOnMatch:    cast.ToInt(row[`stop_on_match`]) == 1,
		ConfigJSON:     strings.TrimSpace(cast.ToString(row[`config_json`])),
		Level:          `warning`,
		Category:       ``,
	}
	if item.Name == `` {
		item.Name = item.Pattern
	}
	item.Level, item.Category = parseShellOutRuleConfig(item.ConfigJSON)
	return item
}

func parseShellOutRuleConfig(configJSON string) (string, string) {
	level := `warning`
	category := ``
	if strings.TrimSpace(configJSON) == `` {
		return level, category
	}
	configMap := map[string]any{}
	if err := json.Unmarshal([]byte(configJSON), &configMap); err != nil {
		return level, category
	}
	if value := strings.TrimSpace(cast.ToString(configMap[`level`])); value != `` {
		level = value
	}
	category = strings.TrimSpace(cast.ToString(configMap[`category`]))
	return level, category
}

func matchShellOutRuleItem(item ShellOutRuleItem, line string) bool {
	pattern := strings.TrimSpace(item.Pattern)
	if pattern == `` {
		return false
	}
	if excludePattern := strings.TrimSpace(item.ExcludePattern); excludePattern != `` && matchShellOutPattern(item.MatchType, excludePattern, line) {
		return false
	}
	return matchShellOutPattern(item.MatchType, pattern, line)
}

func matchShellOutPattern(matchType, pattern, line string) bool {
	switch strings.TrimSpace(matchType) {
	case ShellOutRuleMatchTypeContains:
		return strings.Contains(line, pattern)
	case ShellOutRuleMatchTypeRegex:
		re, err := regexp.Compile(pattern)
		if err != nil {
			return false
		}
		return re.MatchString(line)
	default:
		return strings.Contains(line, pattern)
	}
}

func (h *TShellOut) recordRuleHit(shellOut *ShellOut, item ShellOutRuleItem) {
	if shellOut.regexFiltersTips == nil {
		shellOut.regexFiltersTips = map[string]int{}
	}
	key := strings.TrimSpace(item.Name)
	if key == `` {
		key = strings.TrimSpace(item.Pattern)
	}
	shellOut.regexFiltersTips[key] += 1
	h.SendFilter(shellOut, key)
}

func (h *TShellOut) recordAlertEvent(shellOut *ShellOut, item ShellOutRuleItem, line string, lineNumber int64) ShellOutAlertEvent {
	event := ShellOutAlertEvent{
		RuleItemID: item.ID,
		RuleName:   item.Name,
		Level:      item.Level,
		Category:   item.Category,
		SampleLine: line,
		LineNumber: lineNumber,
		Time:       gstool.TimeNowUnixToString(``),
	}
	shellOut.alertEvents = append(shellOut.alertEvents, event)
	block := ErrorBlock{
		ErrorLine:  line,
		LineNumber: lineNumber,
		Time:       event.Time,
		RuleName:   item.Name,
		Level:      item.Level,
		Category:   item.Category,
	}
	shellOut.errorList = append(shellOut.errorList, block)
	h.SendErr(shellOut, block)
	return event
}

// RefreshRuleItemsByRuleSetId 刷新使用指定规则集的所有运行中客户端的规则项。 // RefreshRuleItemsByRuleSetId reloads rule items for all running ShellOut clients bound to the given rule set.
func (h *TShellOut) RefreshRuleItemsByRuleSetId(ruleSetID int) {
	if ruleSetID <= 0 {
		return
	}
	h.lock.Lock()
	defer h.lock.Unlock()

	ruleItems, err := h.LoadRuleItems(ruleSetID)
	if err != nil {
		return
	}

	for _, shellOut := range h.ShellOutMap {
		if shellOut != nil && shellOut.ruleSetId == ruleSetID {
			shellOut.ruleItems = ruleItems
		}
	}
}
