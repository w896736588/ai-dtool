package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/plw"
	"errors"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

const (
	// defaultSmartLinkScrapeWaitSeconds 抓取任务默认等待秒数。
	// 注意：TAPD 等 SPA 页面需求详情走 AJAX 异步加载，等待时间过短会在节点未渲染前就调 locator.Count() 导致"未找到抓取节点"。
	defaultSmartLinkScrapeWaitSeconds = 5
	// maxSmartLinkScrapeWaitSeconds 抓取任务允许的最大等待秒数。
	maxSmartLinkScrapeWaitSeconds = 60
	// smartLinkScrapeTaskTimeout 抓取任务接口同步等待 Agent 完成的超时时间。
	smartLinkScrapeTaskTimeout = 5 * time.Minute
)

type SmartLinkScrapeRequest struct {
	SmartLinkID     int    `json:"smart_link_id"`
	Label           string `json:"label"`
	UserName        string `json:"user_name"`
	Password        string `json:"password"`
	OpenNum         int    `json:"open_num"`
	OpenType        int    `json:"open_type"`
	RunParams       string `json:"run_params"`
	JumpURL         string `json:"jump_url"`
	CssSelector     string `json:"css_selector"`
	WaitSeconds     int    `json:"wait_seconds"`
	SseDistributeId string `json:"sse_distribute_id"`
}

// summarizeSmartLinkScrapeRequest 生成抓取请求的日志摘要，避免输出敏感信息。
func summarizeSmartLinkScrapeRequest(req SmartLinkScrapeRequest) string {
	return fmt.Sprintf(
		"smart_link_id=%d label=%s user_name=%s password_set=%t open_num=%d open_type=%d run_params_len=%d jump_url=%s css_selector=%s wait_seconds=%d sse_distribute_id=%s",
		req.SmartLinkID,
		req.Label,
		req.UserName,
		req.Password != "",
		req.OpenNum,
		req.OpenType,
		len(req.RunParams),
		req.JumpURL,
		req.CssSelector,
		req.WaitSeconds,
		req.SseDistributeId,
	)
}

func extractSafeTokenFromRequest(c *gin.Context) string {
	if c == nil {
		return ""
	}
	token := strings.TrimSpace(c.GetHeader("Token"))
	if token != "" {
		return token
	}
	token, _ = c.Cookie("safe_token")
	token = strings.TrimSpace(token)
	if token != "" {
		return token
	}
	return strings.TrimSpace(c.Query("token"))
}

// parseSmartLinkScrapeRequest 统一校验抓取任务请求，避免控制器里散落重复判断。
func parseSmartLinkScrapeRequest(req map[string]any) (SmartLinkScrapeRequest, error) {
	result := SmartLinkScrapeRequest{
		SmartLinkID:     cast.ToInt(req["smart_link_id"]),
		Label:           strings.TrimSpace(cast.ToString(req["label"])),
		UserName:        strings.TrimSpace(cast.ToString(req["user_name"])),
		Password:        cast.ToString(req["password"]),
		OpenNum:         cast.ToInt(req["open_num"]),
		OpenType:        int(define.OpenTypeWebkitChrome),
		RunParams:       cast.ToString(req["run_params"]),
		JumpURL:         strings.TrimSpace(cast.ToString(req["jump_url"])),
		CssSelector:     strings.TrimSpace(cast.ToString(req["css_selector"])),
		WaitSeconds:     cast.ToInt(req["wait_seconds"]),
		SseDistributeId: strings.TrimSpace(cast.ToString(req["sse_distribute_id"])),
	}
	if result.SmartLinkID <= 0 {
		return result, errors.New("smart_link_id不能为空")
	}
	if result.Label == "" {
		return result, errors.New("label不能为空")
	}
	if result.JumpURL == "" {
		return result, errors.New("jump_url不能为空")
	}
	if _, err := url.ParseRequestURI(result.JumpURL); err != nil {
		return result, errors.New("jump_url不合法")
	}
	if result.CssSelector == "" {
		return result, errors.New("css_selector不能为空")
	}
	if result.WaitSeconds <= 0 {
		result.WaitSeconds = defaultSmartLinkScrapeWaitSeconds
	}
	if result.WaitSeconds > maxSmartLinkScrapeWaitSeconds {
		result.WaitSeconds = maxSmartLinkScrapeWaitSeconds
	}
	return result, nil
}

// buildScrapeTaskRequestPayload 统一保存抓取任务下发快照，便于排查服务端与 Agent 的协商参数。
func buildScrapeTaskRequestPayload(taskData define.AgentTaskExecuteData) string {
	return gstool.JsonEncode(taskData)
}

// saveSmartLinkScrapeResultFile 将 Agent 回传的 ZIP 保存到 web/download 并返回下载地址。
func saveSmartLinkScrapeResultFile(taskID, originalName string, content []byte) (define.AgentTaskResultFileUploadResponse, error) {
	if taskID == "" {
		return define.AgentTaskResultFileUploadResponse{}, errors.New("task_id不能为空")
	}
	fileName := strings.TrimSpace(originalName)
	if fileName == "" {
		fileName = taskID + ".zip"
	}
	fileName = filepath.Base(fileName)
	if !strings.HasSuffix(strings.ToLower(fileName), ".zip") {
		fileName = fileName + ".zip"
	}
	finalName := fmt.Sprintf("%s_%d_%s", taskID, time.Now().Unix(), fileName)
	targetPath := buildWebDownloadFilePath(finalName)
	if err := gstool.DirCreatePath(filepath.Dir(targetPath)); err != nil {
		return define.AgentTaskResultFileUploadResponse{}, err
	}
	if err := os.WriteFile(targetPath, content, 0o644); err != nil {
		return define.AgentTaskResultFileUploadResponse{}, err
	}
	return define.AgentTaskResultFileUploadResponse{
		DownloadURL: "/api/download/" + finalName,
		FileName:    finalName,
	}, nil
}

// buildAbsoluteDownloadURL 将下载相对路径补全为当前请求可直接访问的绝对地址，并在可用时附带 token。
func buildAbsoluteDownloadURL(c *gin.Context, downloadPath string) string {
	normalizedDownloadPath := strings.TrimSpace(downloadPath)
	if normalizedDownloadPath == "" {
		return ""
	}
	if c == nil || c.Request == nil {
		return normalizedDownloadPath
	}

	host := strings.TrimSpace(c.Request.Host)
	if host == "" {
		return normalizedDownloadPath
	}

	downloadURL, err := url.Parse(normalizedDownloadPath)
	if err != nil {
		return normalizedDownloadPath
	}
	if downloadURL.IsAbs() {
		return normalizedDownloadPath
	}

	// 优先使用代理透传协议，避免 https 入口被拼成 http 下载地址。
	scheme := strings.TrimSpace(c.GetHeader("X-Forwarded-Proto"))
	if scheme == "" {
		if c.Request.TLS != nil {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}

	token := extractSafeTokenFromRequest(c)
	if token != "" {
		queryValues := downloadURL.Query()
		queryValues.Set("token", token)
		downloadURL.RawQuery = queryValues.Encode()
	}

	return scheme + "://" + host + downloadURL.RequestURI()
}

// getFirstAccountFromSmartLink 根据 smartLinkID 和 label 找到对应 link 的 account_list 配置，
// 取关联账号组中的第一个账号返回。
func getFirstAccountFromSmartLink(smartLinkID int, label string) (string, string) {
	smartLink, err := common.DbMain.Client.QueryBySql(`select links from tbl_smart_link where id = ?`, smartLinkID).One()
	if err != nil || len(smartLink) == 0 {
		return "", ""
	}
	linkList := make([]map[string]any, 0)
	if decodeErr := gstool.JsonDecode(cast.ToString(smartLink["links"]), &linkList); decodeErr != nil {
		return "", ""
	}
	for _, link := range linkList {
		if cast.ToString(link["label"]) == label {
			accountList := getAccountListByName(link)
			if len(accountList) > 0 {
				return accountList[0]["user_name"], accountList[0]["password"]
			}
			break
		}
	}
	return "", ""
}

// dispatchScrapeTaskAndAwait 派发抓取任务并同步等待结果，不依赖 gin.Context。
// server 模式下直接在本机执行抓取；local_client 模式下通过 WebSocket 下发 Agent。
func dispatchScrapeTaskAndAwait(smartLinkID int, label, jumpURL, cssSelector string, waitSeconds int) (define.AgentTaskResultFileUploadResponse, error) {
	// 从 smart_link 的 links 中找到对应 label，再通过 account_list 关联的账号组取第一个账号。
	accountUserName, accountPassword := getFirstAccountFromSmartLink(smartLinkID, label)

	runParams, runParamsErr := plw.GetRunParams(smartLinkID, label, accountUserName, accountPassword, 0, 1, make(map[string]string))
	if runParamsErr != nil {
		return define.AgentTaskResultFileUploadResponse{}, fmt.Errorf("构建运行参数失败: %w", runParamsErr)
	}

	cfg := getSmartLinkConfig()
	if cfg.RunMode != define.SmartLinkRunModeLocalClient {
		return dispatchScrapeTaskLocal(runParams, jumpURL, cssSelector, waitSeconds)
	}

	info := GlobalClientRegistry.GetLatest()
	if info == nil || GlobalAgentWsManager.GetConnection(info.ClientID) == nil {
		return define.AgentTaskResultFileUploadResponse{}, errors.New("SMART_LINK_CLIENT_OFFLINE")
	}
	if info.ClientVersion != cfg.ClientVersion {
		return define.AgentTaskResultFileUploadResponse{}, errors.New("SMART_LINK_CLIENT_VERSION_MISMATCH")
	}
	if info.Status == define.SmartLinkClientStatusPreparingRuntime {
		return define.AgentTaskResultFileUploadResponse{}, errors.New("SMART_LINK_CLIENT_PREPARING_RUNTIME")
	}

	now := time.Now().Unix()
	taskID := "scrape_task_" + cast.ToString(now) + "_" + cast.ToString(smartLinkID)
	sseDistributeID := "smart_link_scrape_" + cast.ToString(now)

	tokenManager := getSafeTokenManager()
	safeToken, _, tokenErr := tokenManager.GenerateToken()
	if tokenErr != nil {
		return define.AgentTaskResultFileUploadResponse{}, fmt.Errorf("生成token失败: %w", tokenErr)
	}

	taskData := define.AgentTaskExecuteData{
		TaskID:          taskID,
		SseDistributeId: sseDistributeID,
		ClientID:        info.ClientID,
		TaskType:        define.AgentTaskTypeScrapeToMarkdown,
		SafeToken:       safeToken,
		RunParams:       BuildAgentRunParams(runParams),
		ScrapeConfig: define.AgentTaskScrapeConfig{
			JumpURL:     jumpURL,
			CssSelector: cssSelector,
			WaitSeconds: waitSeconds,
		},
	}

	_, createErr := common.DbMain.Client.QuickCreate("tbl_smart_link_task", map[string]any{
		"task_id":         taskID,
		"client_id":       info.ClientID,
		"smart_link_id":   smartLinkID,
		"label":           label,
		"status":          define.SmartLinkTaskStatusPending,
		"run_mode":        define.SmartLinkRunModeLocalClient,
		"request_payload": buildScrapeTaskRequestPayload(taskData),
		"create_time":     now,
		"update_time":     now,
	}).Exec()
	if createErr != nil {
		return define.AgentTaskResultFileUploadResponse{}, fmt.Errorf("创建任务失败: %w", createErr)
	}

	wsMsg := define.AgentWsMessage{
		Type:            define.AgentWsMsgTaskExecute,
		ClientID:        info.ClientID,
		TaskID:          taskID,
		SseDistributeId: sseDistributeID,
		Data:            taskData,
	}
	if sendErr := GlobalAgentWsManager.Send(info.ClientID, wsMsg); sendErr != nil {
		return define.AgentTaskResultFileUploadResponse{}, fmt.Errorf("下发任务到Agent失败: %w", sendErr)
	}

	resultFile, waitErr := waitForSmartLinkTaskResult(taskID, smartLinkScrapeTaskTimeout, querySmartLinkTaskByID)
	if waitErr != nil {
		return define.AgentTaskResultFileUploadResponse{}, waitErr
	}
	return resultFile, nil
}

// SmartLinkScrapeToMarkdown 创建抓取 Markdown 任务并下发给本地 Agent。
func SmartLinkScrapeToMarkdown(c *gin.Context) {
	reqMap := make(map[string]any)
	if err := gsgin.GinPostBody(c, &reqMap); err != nil {
		gstool.FmtPrintlnLogTime(`[SmartLinkScrapeToMarkdown][01] 解析请求体失败 err=%s`, err.Error())
		gsgin.GinResponseError(c, "请求参数错误", nil)
		return
	}
	req, err := parseSmartLinkScrapeRequest(reqMap)
	if err != nil {
		gstool.FmtPrintlnLogTime(`[SmartLinkScrapeToMarkdown][02] 请求校验失败 req=%s err=%s`, gstool.JsonEncode(reqMap), err.Error())
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gstool.FmtPrintlnLogTime(`[SmartLinkScrapeToMarkdown][03] 收到抓取请求 %s`, summarizeSmartLinkScrapeRequest(req))

	resultFile, dispatchErr := dispatchScrapeTaskAndAwait(req.SmartLinkID, req.Label, req.JumpURL, req.CssSelector, req.WaitSeconds)
	if dispatchErr != nil {
		errMsg := dispatchErr.Error()
		gstool.FmtPrintlnLogTime(`[SmartLinkScrapeToMarkdown][04] 派发抓取任务失败 err=%s`, errMsg)
		gsgin.GinResponseError(c, errMsg, nil)
		return
	}
	resultFile.DownloadURL = buildAbsoluteDownloadURL(c, resultFile.DownloadURL)
	gstool.FmtPrintlnLogTime(`[SmartLinkScrapeToMarkdown][15] 等待任务完成成功 download_url=%s file_name=%s`, resultFile.DownloadURL, resultFile.FileName)

	gsgin.GinResponseSuccess(c, "", map[string]any{
		"download_url": resultFile.DownloadURL,
		"file_name":    resultFile.FileName,
	})
}

// SmartLinkTaskResultFileUpload 接收 Agent 上传的抓取 ZIP 结果文件。
func SmartLinkTaskResultFileUpload(c *gin.Context) {
	taskID := strings.TrimSpace(c.PostForm("task_id"))
	if taskID == "" {
		gstool.FmtPrintlnLogTime(`[SmartLinkTaskResultFileUpload][01] 缺少task_id`)
		gsgin.GinResponseError(c, "task_id不能为空", nil)
		return
	}
	fileHeader, err := c.FormFile("file")
	if err != nil {
		gstool.FmtPrintlnLogTime(`[SmartLinkTaskResultFileUpload][02] 缺少上传文件 task_id=%s err=%s`, taskID, err.Error())
		gsgin.GinResponseError(c, "file不能为空", nil)
		return
	}
	gstool.FmtPrintlnLogTime(`[SmartLinkTaskResultFileUpload][03] 收到结果文件 task_id=%s file_name=%s file_size=%d`, taskID, fileHeader.Filename, fileHeader.Size)
	file, openErr := fileHeader.Open()
	if openErr != nil {
		gstool.FmtPrintlnLogTime(`[SmartLinkTaskResultFileUpload][04] 打开上传文件失败 task_id=%s err=%s`, taskID, openErr.Error())
		gsgin.GinResponseError(c, "打开上传文件失败: "+openErr.Error(), nil)
		return
	}
	defer file.Close()

	content, readErr := io.ReadAll(file)
	if readErr != nil {
		gstool.FmtPrintlnLogTime(`[SmartLinkTaskResultFileUpload][05] 读取上传文件失败 task_id=%s err=%s`, taskID, readErr.Error())
		gsgin.GinResponseError(c, "读取上传文件失败: "+readErr.Error(), nil)
		return
	}
	resp, saveErr := saveSmartLinkScrapeResultFile(taskID, fileHeader.Filename, content)
	if saveErr != nil {
		gstool.FmtPrintlnLogTime(`[SmartLinkTaskResultFileUpload][06] 保存上传文件失败 task_id=%s err=%s`, taskID, saveErr.Error())
		gsgin.GinResponseError(c, "保存上传文件失败: "+saveErr.Error(), nil)
		return
	}
	gstool.FmtPrintlnLogTime(`[SmartLinkTaskResultFileUpload][07] 保存上传文件成功 task_id=%s saved_file_name=%s download_url=%s size=%d`, taskID, resp.FileName, resp.DownloadURL, len(content))

	gsgin.GinResponseSuccess(c, "", resp)
}

// dispatchScrapeTaskLocal 在 server 模式下直接在本机通过 Playwright 执行抓取，无需 Agent。
func dispatchScrapeTaskLocal(runParams *plw.PlaywrightRunParams, jumpURL, cssSelector string, waitSeconds int) (define.AgentTaskResultFileUploadResponse, error) {
	scrapeConfig := define.AgentTaskScrapeConfig{
		JumpURL:     jumpURL,
		CssSelector: cssSelector,
		WaitSeconds: waitSeconds,
	}
	result, err := plw.RunScrapeToMarkdown(runParams, scrapeConfig, component.PlaywrightClient.Log)
	if err != nil {
		return define.AgentTaskResultFileUploadResponse{}, fmt.Errorf("本地抓取执行失败: %w", err)
	}
	now := time.Now().Unix()
	taskID := "scrape_task_" + cast.ToString(now)
	resp, saveErr := saveSmartLinkScrapeResultFile(taskID, result.FileName, result.ZipBytes)
	if saveErr != nil {
		return define.AgentTaskResultFileUploadResponse{}, fmt.Errorf("保存抓取结果失败: %w", saveErr)
	}
	return resp, nil
}
