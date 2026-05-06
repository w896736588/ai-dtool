package mcp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/playwright-community/playwright-go"
)

// AccessibilitySnapshot 保存一次 accessibility tree 快照的结果。
type AccessibilitySnapshot struct {
	// RawTree 是 playwright 返回的原始 accessibility 节点。
	RawTree map[string]any
	// FormattedText 是格式化后的可读文本，带 ref 编号。
	FormattedText string
	// RefNodes 存储 ref → nodeInfo 映射，供后续工具通过 ref 定位元素。
	RefNodes map[string]*RefNode
	// refCounter 递增计数器，格式化时分配 ref 编号。
	refCounter int
}

// RefNode 存储一个可交互元素的信息，用于后续定位。
type RefNode struct {
	Ref      string
	Role     string
	Name     string
	Value    string
	Level    float64
	Children []*RefNode
}

// TakeSnapshot 对当前页面拍摄 accessibility tree 快照。
func TakeSnapshot(page playwright.Page) (*AccessibilitySnapshot, error) {
	jsCode := `() => {
		function walk(node, depth) {
			if (!node) return null;
			const result = {
				role: node.role || '',
				name: node.name || '',
				value: node.value ? String(node.value).substring(0, 200) : '',
			};
			if (node.description) result.description = node.description;
			if (node.level) result.level = node.level;
			if (node.checked !== undefined) result.checked = node.checked;
			if (node.disabled !== undefined) result.disabled = node.disabled;
			if (node.expanded !== undefined) result.expanded = node.expanded;
			if (node.children && node.children.length > 0) {
				result.children = node.children.map(c => walk(c, depth + 1)).filter(Boolean);
			}
			return result;
		}
		const snapshot = document.defaultView.__playwright?.accessibilitySnapshot
			? document.defaultView.__playwright.accessibilitySnapshot()
			: null;
		return JSON.stringify(snapshot);
	}`

	result, err := page.Evaluate(jsCode)
	if err != nil || result == nil {
		return takeSnapshotFallback(page)
	}

	var tree map[string]any
	switch v := result.(type) {
	case string:
		if v == "" || v == "null" {
			return takeSnapshotFallback(page)
		}
		if err := json.Unmarshal([]byte(v), &tree); err != nil {
			return takeSnapshotFallback(page)
		}
	default:
		return takeSnapshotFallback(page)
	}

	snapshot := &AccessibilitySnapshot{
		RawTree:  tree,
		RefNodes: make(map[string]*RefNode),
	}
	snapshot.FormattedText = snapshot.formatTree(tree, 0)
	return snapshot, nil
}

// takeSnapshotFallback 使用 page.Evaluate 提取 DOM 可交互元素作为备选方案。
func takeSnapshotFallback(page playwright.Page) (*AccessibilitySnapshot, error) {
	jsCode := `() => {
		const result = [];
		document.querySelectorAll(
			'button, a, input, select, textarea, [role="button"], [role="link"], [role="tab"], [role="menuitem"], [role="checkbox"], [role="radio"], [role="switch"], [role="textbox"], [role="searchbox"], [role="combobox"], [role="slider"], [role="spinbutton"], [role="heading"], h1, h2, h3, h4, h5, h6'
		).forEach(el => {
			const role = el.getAttribute('role') || el.tagName.toLowerCase();
			result.push({
				role: role,
				name: (el.getAttribute('aria-label') || el.textContent || '').trim().substring(0, 200),
				value: el.value || '',
				tag: el.tagName.toLowerCase(),
				type: el.type || '',
				placeholder: el.placeholder || '',
				href: el.href || '',
				disabled: el.disabled || false,
				visible: el.offsetParent !== null,
			});
		});
		return JSON.stringify(result);
	}`

	raw, err := page.Evaluate(jsCode)
	if err != nil {
		return nil, fmt.Errorf("获取页面元素失败: %w", err)
	}

	var elements []map[string]any
	switch v := raw.(type) {
	case string:
		if err := json.Unmarshal([]byte(v), &elements); err != nil {
			return nil, fmt.Errorf("解析页面元素失败: %w", err)
		}
	default:
		return nil, fmt.Errorf("意外的返回类型")
	}

	snapshot := &AccessibilitySnapshot{
		RawTree:  map[string]any{"role": "page", "children": elements},
		RefNodes: make(map[string]*RefNode),
	}

	var sb strings.Builder
	sb.WriteString("- page\n")
	for _, el := range elements {
		visible, _ := el["visible"].(bool)
		if !visible {
			continue
		}
		snapshot.refCounter++
		ref := fmt.Sprintf("e%d", snapshot.refCounter)
		role, _ := el["role"].(string)
		name, _ := el["name"].(string)
		value, _ := el["value"].(string)
		node := &RefNode{Ref: ref, Role: role, Name: name, Value: value}
		snapshot.RefNodes[ref] = node

		sb.WriteString(fmt.Sprintf("  - %s", role))
		if name != "" {
			sb.WriteString(fmt.Sprintf(" %q", name))
		}
		sb.WriteString(fmt.Sprintf(" [ref=%s", ref))
		if value != "" {
			sb.WriteString(fmt.Sprintf(", value=%q", truncateStr(value, 50)))
		}
		if tag, ok := el["tag"].(string); ok && tag != "" {
			sb.WriteString(fmt.Sprintf(", tag=%s", tag))
		}
		if typ, ok := el["type"].(string); ok && typ != "" {
			sb.WriteString(fmt.Sprintf(", type=%s", typ))
		}
		if placeholder, ok := el["placeholder"].(string); ok && placeholder != "" {
			sb.WriteString(fmt.Sprintf(", placeholder=%q", placeholder))
		}
		if disabled, ok := el["disabled"].(bool); ok && disabled {
			sb.WriteString(", disabled")
		}
		sb.WriteString("]\n")
	}
	snapshot.FormattedText = sb.String()
	return snapshot, nil
}

// formatTree 递归格式化 accessibility 树为可读文本。
func (s *AccessibilitySnapshot) formatTree(node map[string]any, depth int) string {
	if node == nil {
		return ""
	}
	var sb strings.Builder
	indent := strings.Repeat("  ", depth)

	role, _ := node["role"].(string)
	name, _ := node["name"].(string)
	value, _ := node["value"].(string)

	// 可交互角色才分配 ref
	interactive := isInteractiveRole(role)
	var ref string
	if interactive && name != "" {
		s.refCounter++
		ref = fmt.Sprintf("e%d", s.refCounter)
		nodeInfo := &RefNode{
			Ref:   ref,
			Role:  role,
			Name:  name,
			Value: value,
		}
		if level, ok := node["level"].(float64); ok {
			nodeInfo.Level = level
		}
		s.RefNodes[ref] = nodeInfo
	}

	sb.WriteString(indent + "- " + role)
	if name != "" {
		sb.WriteString(fmt.Sprintf(" %q", name))
	}
	if ref != "" {
		sb.WriteString(fmt.Sprintf(" [ref=%s", ref))
	} else {
		sb.WriteString(" [")
	}

	if value != "" {
		sb.WriteString(fmt.Sprintf("value=%q", truncateStr(value, 50)))
	}
	if level, ok := node["level"].(float64); ok {
		sb.WriteString(fmt.Sprintf("level=%.0f", level))
	}
	if checked, ok := node["checked"].(bool); ok {
		sb.WriteString(fmt.Sprintf("checked=%v", checked))
	}
	if disabled, ok := node["disabled"].(bool); ok && disabled {
		sb.WriteString("disabled")
	}
	if expanded, ok := node["expanded"].(bool); ok {
		sb.WriteString(fmt.Sprintf("expanded=%v", expanded))
	}
	sb.WriteString("]\n")

	if children, ok := node["children"].([]any); ok {
		for _, child := range children {
			if childMap, ok := child.(map[string]any); ok {
				sb.WriteString(s.formatTree(childMap, depth+1))
			}
		}
	}
	return sb.String()
}

func isInteractiveRole(role string) bool {
	switch role {
	case "button", "link", "textbox", "searchbox", "combobox",
		"checkbox", "radio", "switch", "slider", "spinbutton",
		"menuitem", "tab", "treeitem", "option",
		"gridcell", "columnheader", "rowheader":
		return true
	}
	return false
}

func truncateStr(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max] + "..."
}
