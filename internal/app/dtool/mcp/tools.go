package mcp

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"

	"dev_tool/internal/app/dtool/component"

	"github.com/mark3labs/mcp-go/mcp"
	mcpserver "github.com/mark3labs/mcp-go/server"
	"github.com/playwright-community/playwright-go"
)

// registerTools 在 MCP Server 上注册所有浏览器操作工具。
func registerTools(srv *mcpserver.MCPServer, session *BrowserSession) {
	// browser_snapshot
	srv.AddTool(
		mcp.NewTool("browser_snapshot",
			mcp.WithDescription("获取当前页面的 Accessibility Tree 快照。返回结构化文本，包含页面元素层级和可交互元素的 ref 编号。用于理解页面结构，不需要截图。"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			session.updateActivity()
			page, err := session.getActivePage()
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			snapshot, err := TakeSnapshot(page)
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("获取快照失败: %v", err)), nil
			}
			session.setSnapshot(snapshot)
			return mcp.NewToolResultText(snapshot.FormattedText), nil
		},
	)

	// browser_click
	srv.AddTool(
		mcp.NewTool("browser_click",
			mcp.WithDescription("点击页面上的元素。使用 browser_snapshot 返回的 ref 编号定位元素，或通过 role+name 组合定位。"),
			mcp.WithString("ref", mcp.Description("browser_snapshot 返回的元素 ref 编号，如 e1、e2")),
			mcp.WithString("role", mcp.Description("元素的 ARIA role，如 button、link、textbox")),
			mcp.WithString("name", mcp.Description("元素的 accessible name")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			session.updateActivity()
			page, err := session.getActivePage()
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			ref := req.GetString("ref", "")
			role := req.GetString("role", "")
			name := req.GetString("name", "")

			resolvedRole, resolvedName, err := resolveElementRef(session, ref, role, name)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			locator := page.GetByRole(playwright.AriaRole(resolvedRole), playwright.PageGetByRoleOptions{Name: resolvedName})
			if err := locator.Click(); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("点击失败: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("已点击 %s %q", resolvedRole, resolvedName)), nil
		},
	)

	// browser_type
	srv.AddTool(
		mcp.NewTool("browser_type",
			mcp.WithDescription("在输入框中输入文本（追加模式）。使用 ref 或 role+name 定位目标元素。"),
			mcp.WithString("text", mcp.Description("要输入的文本"), mcp.Required()),
			mcp.WithString("ref", mcp.Description("browser_snapshot 返回的元素 ref 编号")),
			mcp.WithString("role", mcp.Description("元素的 ARIA role")),
			mcp.WithString("name", mcp.Description("元素的 accessible name")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			session.updateActivity()
			page, err := session.getActivePage()
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			text := req.GetString("text", "")
			if text == "" {
				return mcp.NewToolResultError("text 不能为空"), nil
			}
			ref := req.GetString("ref", "")
			role := req.GetString("role", "")
			name := req.GetString("name", "")

			resolvedRole, resolvedName, err := resolveElementRef(session, ref, role, name)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			locator := page.GetByRole(playwright.AriaRole(resolvedRole), playwright.PageGetByRoleOptions{Name: resolvedName})
			if err := locator.Fill(text); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("输入失败: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("已在 %s %q 中输入文本", resolvedRole, resolvedName)), nil
		},
	)

	// browser_fill
	srv.AddTool(
		mcp.NewTool("browser_fill",
			mcp.WithDescription("清空输入框并填入新文本。与 browser_type 不同，会先清空已有内容。"),
			mcp.WithString("text", mcp.Description("要填入的文本"), mcp.Required()),
			mcp.WithString("ref", mcp.Description("browser_snapshot 返回的元素 ref 编号")),
			mcp.WithString("role", mcp.Description("元素的 ARIA role")),
			mcp.WithString("name", mcp.Description("元素的 accessible name")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			session.updateActivity()
			page, err := session.getActivePage()
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			text := req.GetString("text", "")
			ref := req.GetString("ref", "")
			role := req.GetString("role", "")
			name := req.GetString("name", "")

			resolvedRole, resolvedName, err := resolveElementRef(session, ref, role, name)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			locator := page.GetByRole(playwright.AriaRole(resolvedRole), playwright.PageGetByRoleOptions{Name: resolvedName})
			if err := locator.Clear(); err != nil {
				// Clear 可能不支持某些元素，忽略错误继续 Fill
			}
			if err := locator.Fill(text); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("填充失败: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("已在 %s %q 中填充文本", resolvedRole, resolvedName)), nil
		},
	)

	// browser_navigate
	srv.AddTool(
		mcp.NewTool("browser_navigate",
			mcp.WithDescription("导航到指定 URL。"),
			mcp.WithString("url", mcp.Description("目标 URL"), mcp.Required()),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			session.updateActivity()
			page, err := session.getActivePage()
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			url := req.GetString("url", "")
			if url == "" {
				return mcp.NewToolResultError("url 不能为空"), nil
			}
			if _, err := page.Goto(url); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("导航失败: %v", err)), nil
			}
			component.PlaywrightClient.WaitForLoadState(&page, 30000)
			title, _ := page.Title()
			return mcp.NewToolResultText(fmt.Sprintf("已导航到 %s，页面标题: %s", url, title)), nil
		},
	)

	// browser_select_option
	srv.AddTool(
		mcp.NewTool("browser_select_option",
			mcp.WithDescription("在下拉选择框中选择选项。"),
			mcp.WithString("values", mcp.Description("要选择的选项值，多个值用逗号分隔"), mcp.Required()),
			mcp.WithString("ref", mcp.Description("browser_snapshot 返回的元素 ref 编号")),
			mcp.WithString("role", mcp.Description("元素的 ARIA role")),
			mcp.WithString("name", mcp.Description("元素的 accessible name")),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			session.updateActivity()
			page, err := session.getActivePage()
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			valuesStr := req.GetString("values", "")
			ref := req.GetString("ref", "")
			role := req.GetString("role", "")
			name := req.GetString("name", "")

			resolvedRole, resolvedName, err := resolveElementRef(session, ref, role, name)
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}

			values := strings.Split(valuesStr, ",")
			locator := page.GetByRole(playwright.AriaRole(resolvedRole), playwright.PageGetByRoleOptions{Name: resolvedName})
			if _, err := locator.SelectOption(playwright.SelectOptionValues{ValuesOrLabels: &values}); err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("选择失败: %v", err)), nil
			}
			return mcp.NewToolResultText(fmt.Sprintf("已在 %s %q 中选择 %v", resolvedRole, resolvedName, values)), nil
		},
	)

	// browser_screenshot
	srv.AddTool(
		mcp.NewTool("browser_screenshot",
			mcp.WithDescription("截取当前页面的截图。仅在需要视觉验证最终结果时使用，日常操作请使用 browser_snapshot。返回 base64 编码的 PNG 图片。"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			session.updateActivity()
			page, err := session.getActivePage()
			if err != nil {
				return mcp.NewToolResultError(err.Error()), nil
			}
			imgBytes, err := page.Screenshot()
			if err != nil {
				return mcp.NewToolResultError(fmt.Sprintf("截图失败: %v", err)), nil
			}
			imgBase64 := base64.StdEncoding.EncodeToString(imgBytes)
			return mcp.NewToolResultImage(
				fmt.Sprintf("截图成功，大小: %d bytes", len(imgBytes)),
				imgBase64,
				"image/png",
			), nil
		},
	)

	// browser_close
	srv.AddTool(
		mcp.NewTool("browser_close",
			mcp.WithDescription("关闭当前浏览器会话并释放资源。任务完成后调用。"),
		),
		func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
			RemoveSession(session.ID)
			return mcp.NewToolResultText("浏览器会话已关闭"), nil
		},
	)
}

// resolveElementRef 优先用 ref 查找 snapshot 中的元素信息，否则使用传入的 role+name。
func resolveElementRef(session *BrowserSession, ref, role, name string) (string, string, error) {
	if ref != "" {
		snapshot := session.getSnapshot()
		if snapshot == nil {
			return "", "", fmt.Errorf("请先调用 browser_snapshot 获取页面快照后再使用 ref 定位元素")
		}
		node, ok := snapshot.RefNodes[ref]
		if !ok {
			return "", "", fmt.Errorf("找不到 ref=%s 的元素，请重新调用 browser_snapshot", ref)
		}
		return node.Role, node.Name, nil
	}
	if role != "" {
		return role, name, nil
	}
	return "", "", fmt.Errorf("必须提供 ref 或 role 参数来定位元素")
}
