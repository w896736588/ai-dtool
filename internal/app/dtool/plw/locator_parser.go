package plw

import (
	"errors"
	"strings"
)

// LocatorParser 负责将原始 Locator 配置归一化成标准 LocatorSpec。
type LocatorParser struct{}

// NewLocatorParser 创建 Locator 解析器。
func NewLocatorParser() *LocatorParser {
	return &LocatorParser{}
}

// Parse 将原始 Locator 输入转换成标准 LocatorSpec。
func (h *LocatorParser) Parse(input *LocatorInput) (*LocatorSpec, error) {
	if input == nil {
		return nil, errors.New(`locator 配置不能为空`)
	}
	if input.Spec != nil {
		return h.normalizeSpec(input.Spec)
	}
	return nil, errors.New(`locator 必须使用结构化配置`)
}

// normalizeSpec 负责收敛业务别名并校验结构化 Locator 配置。
func (h *LocatorParser) normalizeSpec(input *LocatorSpec) (*LocatorSpec, error) {
	if input == nil {
		return nil, errors.New(`locator spec 不能为空`)
	}
	spec := *input
	if spec.Method == `` {
		spec.Method = spec.FindType
	}
	spec.Method = strings.TrimSpace(spec.Method)
	spec.Value = strings.TrimSpace(spec.Value)
	if spec.Method == `` {
		return nil, errors.New(`locator method 不能为空`)
	}
	if spec.Value == `` {
		return nil, errors.New(`locator value 不能为空`)
	}

	spec.Options = mergeLocatorOptions(spec.Options, &LocatorOptions{
		Exact:         spec.Exact,
		Name:          spec.Name,
		Checked:       spec.Checked,
		Disabled:      spec.Disabled,
		Selected:      spec.Selected,
		Expanded:      spec.Expanded,
		IncludeHidden: spec.IncludeHidden,
		Level:         spec.Level,
	})
	spec.Pick = mergeLocatorPick(spec.Pick, &LocatorPick{
		First: spec.First,
		Last:  spec.Last,
		Nth:   spec.Nth,
	})
	if err := validateLocatorPick(spec.Pick); err != nil {
		return nil, err
	}

	if len(spec.Chain) > 0 {
		chainList := make([]LocatorSpec, 0, len(spec.Chain))
		for _, item := range spec.Chain {
			normalized, err := h.normalizeSpec(&item)
			if err != nil {
				return nil, err
			}
			chainList = append(chainList, *normalized)
		}
		spec.Chain = chainList
	}

	if len(spec.Filters) > 0 {
		filterList := make([]LocatorFilter, 0, len(spec.Filters))
		for _, item := range spec.Filters {
			filter := item
			if item.Has != nil {
				hasSpec, err := h.normalizeSpec(item.Has)
				if err != nil {
					return nil, err
				}
				filter.Has = hasSpec
			}
			if item.HasNot != nil {
				hasNotSpec, err := h.normalizeSpec(item.HasNot)
				if err != nil {
					return nil, err
				}
				filter.HasNot = hasNotSpec
			}
			filterList = append(filterList, filter)
		}
		spec.Filters = filterList
	}

	return &spec, nil
}

// mergeLocatorOptions 将根节点别名合并到标准 Options 中。
func mergeLocatorOptions(current *LocatorOptions, alias *LocatorOptions) *LocatorOptions {
	if current == nil && alias == nil {
		return nil
	}
	result := &LocatorOptions{}
	if current != nil {
		*result = *current
	}
	if alias == nil {
		return result
	}
	if result.Exact == nil {
		result.Exact = alias.Exact
	}
	if result.Name == `` {
		result.Name = alias.Name
	}
	if result.Checked == nil {
		result.Checked = alias.Checked
	}
	if result.Disabled == nil {
		result.Disabled = alias.Disabled
	}
	if result.Selected == nil {
		result.Selected = alias.Selected
	}
	if result.Expanded == nil {
		result.Expanded = alias.Expanded
	}
	if result.IncludeHidden == nil {
		result.IncludeHidden = alias.IncludeHidden
	}
	if result.Level == nil {
		result.Level = alias.Level
	}
	return result
}

// mergeLocatorPick 将根节点别名合并到标准 Pick 中。
func mergeLocatorPick(current *LocatorPick, alias *LocatorPick) *LocatorPick {
	if current == nil && alias == nil {
		return nil
	}
	result := &LocatorPick{}
	if current != nil {
		*result = *current
	}
	if alias == nil {
		return result
	}
	if result.First == nil {
		result.First = alias.First
	}
	if result.Last == nil {
		result.Last = alias.Last
	}
	if result.Nth == nil {
		result.Nth = alias.Nth
	}
	return result
}

// validateLocatorPick 仅允许 first、last、nth 三者之一生效，避免选择语义冲突。
func validateLocatorPick(pick *LocatorPick) error {
	if pick == nil {
		return nil
	}
	pickCount := 0
	if pick.First != nil && *pick.First {
		pickCount++
	}
	if pick.Last != nil && *pick.Last {
		pickCount++
	}
	if pick.Nth != nil {
		pickCount++
	}
	if pickCount > 1 {
		return errors.New(`pick 配置冲突，只能设置 first、last、nth 其中一个`)
	}
	return nil
}
