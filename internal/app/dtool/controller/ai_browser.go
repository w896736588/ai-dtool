package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/mcp"
	"dev_tool/internal/app/dtool/plw"
	"fmt"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/playwright-community/playwright-go"
	"github.com/spf13/cast"
)

type aiBrowserOpenRequest struct {
	SmartLinkID int    `json:"smart_link_id"`
	Label       string `json:"label"`
	Account     string `json:"account"`
	OpenType    int    `json:"open_type"`
	ReuseIfOpen *bool  `json:"reuse_if_open"`
	// EnableMCP 为 true 时，登录完成后保持浏览器存活并创建 MCP Server，
	// AI 通过 MCP SSE 直接调用工具操作浏览器，无需每步重新打开。
	EnableMCP bool `json:"enable_mcp"`
}

type aiBrowserResolvedAccount struct {
	ID         int
	UserName   string
	Password   string
	AccountKey string
}

// AIBrowserSessionOpen 打开自定义网页并准备可复用的 Chromium 用户数据目录。
// 该接口完成网页登录后会关闭当前浏览器，再把 userDataDir 返回给 AI 侧原生 Playwright 使用。
func AIBrowserSessionOpen(c *gin.Context) {
	var req aiBrowserOpenRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	req.Label = strings.TrimSpace(req.Label)
	if req.SmartLinkID == 0 || req.Label == "" {
		gsgin.GinResponseError(c, "smart_link_id和label不能为空", nil)
		return
	}
	if !ensureSmartLinkNodeInstalled(c, nil) {
		return
	}

	accountInfo, err := resolveAIBrowserAccount(req.Account)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	runParams, err := plw.GetRunParams(req.SmartLinkID, req.Label, accountInfo.UserName, accountInfo.Password, req.OpenType, 1, map[string]string{})
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	runParams.StreamFunc = func(string, string) {}

	contextList := plw.NewContextList(component.PlaywrightClient.Log)
	contextPage, boolCleanFirstBlank, err := contextList.GetContextSaveUserData(runParams)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	// 确保在任何返回路径下都关闭浏览器，避免残留进程锁住 userDataDir 导致 AI 侧崩溃
	// MCP 模式下不关闭，由 MCP 会话管理生命周期
	if !req.EnableMCP {
		defer func() {
			if closeErr := closeAIBrowserPreparedContext(contextList, contextPage); closeErr != nil {
				component.PlaywrightClient.Log.Errof("关闭AI浏览器session失败: %v", closeErr)
			}
		}()
	}

	page := findContextPageForDomain(contextPage, runParams.Domain)
	reused := page != nil
	if page == nil {
		page, err = openAIBrowserPage(contextPage, runParams, boolCleanFirstBlank)
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
	}

	// 同步执行自定义网页流程（登录等），确保流程完整执行后才关闭浏览器
	playwrightRunner := plw.NewPlaywright(runParams, component.PlaywrightClient.Log)
	if err := playwrightRunner.RunProcessesSync(page); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	time.Sleep(3 * time.Second)
	// 等待登录后的页面跳转完成（如点击登录按钮后的重定向）
	component.PlaywrightClient.WaitForLoadState(page, runParams.LocatorTimeout)

	response := buildAIBrowserProfileResponse(req, runParams, contextPage, page, accountInfo, reused)

	// MCP 模式：创建 MCP SSE 会话，保持浏览器存活
	if req.EnableMCP {
		baseURL := fmt.Sprintf("http://%s", c.Request.Host)
		browserSession, mcpErr := mcp.CreateSession(contextPage, page, baseURL)
		if mcpErr != nil {
			gsgin.GinResponseError(c, fmt.Sprintf("创建MCP会话失败: %v", mcpErr), nil)
			return
		}
		response["mcp"] = map[string]any{
			"enabled":      true,
			"session_id":   browserSession.ID,
			"sse_endpoint": fmt.Sprintf("%s/mcp/ai-browser/%s/sse", baseURL, browserSession.ID),
			"msg_endpoint": fmt.Sprintf("%s/mcp/ai-browser/%s/message", baseURL, browserSession.ID),
		}
		response["source_browser_closed"] = false
		response["usage_hint"] = "MCP模式：浏览器保持存活，AI通过MCP SSE端点直接调用browser_snapshot/browser_click等工具操作浏览器，无需重新打开浏览器"
	}

	gsgin.GinResponseSuccess(c, "", response)
}

func buildAIBrowserProfileResponse(req aiBrowserOpenRequest, runParams *plw.PlaywrightRunParams, contextPage *plw.ContextPage, page *playwright.Page, accountInfo aiBrowserResolvedAccount, reused bool) map[string]any {
	accountPayload := map[string]any{
		"id":          accountInfo.ID,
		"user_name":   accountInfo.UserName,
		"account_key": accountInfo.AccountKey,
	}
	if accountInfo.UserName == "" && accountInfo.ID == 0 && accountInfo.AccountKey == "" {
		accountPayload = map[string]any{}
	}
	executablePath := component.PlaywrightClient.Pw.Chromium.ExecutablePath()
	return map[string]any{
		"browser_type":          "chromium",
		"source_browser_closed": true,
		"native_playwright": map[string]any{
			"mode":            "launch_persistent_context",
			"user_data_dir":   contextPage.UserDataPath,
			"executable_path": executablePath,
			"channel":         runParams.Channel,
		},
		"user_data_dir":   contextPage.UserDataPath,
		"user_data_index": contextPage.UserDataIndex,
		"smart_link": map[string]any{
			"id":    req.SmartLinkID,
			"label": req.Label,
		},
		"site": map[string]any{
			"domain": runParams.Domain,
			"url":    runParams.Link,
		},
		"account": accountPayload,
		"reused":  reused,
		"current_page": map[string]any{
			"url":   (*page).URL(),
			"title": safePageTitle(page),
		},
		"usage_hint": "AI应使用Playwright的launchPersistentContext(userDataDir, executablePath)重新接管该目录，必须使用返回的executable_path以确保浏览器版本一致",
	}
}

func closeAIBrowserPreparedContext(contextList *plw.ContextPageList, contextPage *plw.ContextPage) error {
	if contextPage == nil || contextPage.Context == nil || *contextPage.Context == nil {
		return nil
	}
	if err := (*contextPage.Context).Close(); err != nil {
		return fmt.Errorf("关闭准备阶段浏览器失败: %w", err)
	}
	contextList.CleanContextList(false)
	return nil
}

func resolveAIBrowserAccount(account string) (aiBrowserResolvedAccount, error) {
	account = strings.TrimSpace(account)
	if account == "" {
		return aiBrowserResolvedAccount{}, nil
	}
	info, err := common.DbMain.Client.QuickQuery("tbl_account", "*", map[string]any{
		"username": account,
	}).One()
	if err != nil {
		return aiBrowserResolvedAccount{}, err
	}
	if len(info) == 0 {
		return aiBrowserResolvedAccount{}, fmt.Errorf("账号不存在")
	}
	return aiBrowserResolvedAccount{
		ID:         cast.ToInt(info["id"]),
		UserName:   cast.ToString(info["username"]),
		Password:   cast.ToString(info["password"]),
		AccountKey: "account_user_" + account,
	}, nil
}

func findContextPageForDomain(contextPage *plw.ContextPage, domain string) *playwright.Page {
	if contextPage == nil {
		return nil
	}
	for _, page := range contextPage.Pages() {
		pageCopy := page
		if pageCopy.IsClosed() {
			continue
		}
		if domain == "" || gstool.UrlGetHost(pageCopy.URL()) == domain {
			return &pageCopy
		}
	}
	return nil
}

func openAIBrowserPage(contextPage *plw.ContextPage, runParams *plw.PlaywrightRunParams, boolCleanFirstBlank bool) (*playwright.Page, error) {
	page, err := (*contextPage.Context).NewPage()
	if err != nil {
		return nil, err
	}
	contextPage.RegisterLinks(page, runParams.ListenCurls, runParams.FilterUris, func(string, string) {})
	playwrightRunner := plw.NewPlaywright(runParams, component.PlaywrightClient.Log)
	playwrightRunner.LastUserDataIndex(runParams, contextPage.UserDataIndex, nil)
	if boolCleanFirstBlank {
		// 持久化 context 初次启动通常会带一个 blank 页，这里沿用现有策略清理掉。
		contextPage.CloseFirstPage()
	}
	if _, err = page.Goto(runParams.Link); err != nil {
		return nil, err
	}
	component.PlaywrightClient.WaitForLoadState(&page, runParams.LocatorTimeout)
	return &page, nil
}

func safePageTitle(page *playwright.Page) string {
	if page == nil || *page == nil {
		return ""
	}
	title, err := (*page).Title()
	if err != nil {
		return ""
	}
	return title
}
