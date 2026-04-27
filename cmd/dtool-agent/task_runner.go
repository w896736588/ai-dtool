package main

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/plw"
	"dev_tool/internal/pkg/p_common"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

// TaskRunner 管理任务执行
type TaskRunner struct {
	wsClient    *WsClient
	currentTask string
	mu          sync.Mutex
}

// resolveAgentTaskCombineType 统一 Agent 任务的会话目录策略。
func resolveAgentTaskCombineType(taskType define.AgentTaskType, combineType int) (int, bool) {
	return combineType, false
}

// summarizeAgentTaskData 输出 Agent 任务摘要，避免记录敏感明文。
func summarizeAgentTaskData(taskData define.AgentTaskExecuteData) string {
	return fmt.Sprintf(
		"task_id=%s client_id=%s sse_distribute_id=%s task_type=%s smart_link_id=%d link=%s link_id_label=%s open_type=%d combine_type=%d process_count=%d filter_uri_count=%d browser_auth_username=%s browser_auth_password_set=%t safe_token_set=%t jump_url=%s css_selector=%s wait_seconds=%d",
		taskData.TaskID,
		taskData.ClientID,
		taskData.SseDistributeId,
		taskData.TaskType,
		taskData.RunParams.Id,
		taskData.RunParams.Link,
		taskData.RunParams.LinkIdLabel,
		taskData.RunParams.OpenType,
		taskData.RunParams.CombineType,
		len(taskData.RunParams.ProcessList),
		len(taskData.RunParams.FilterUris),
		taskData.RunParams.BrowserAuthUsername,
		taskData.RunParams.BrowserAuthPassword != "",
		taskData.SafeToken != "",
		taskData.ScrapeConfig.JumpURL,
		taskData.ScrapeConfig.CssSelector,
		taskData.ScrapeConfig.WaitSeconds,
	)
}

// NewTaskRunner 创建任务执行器
func NewTaskRunner(wsClient *WsClient) *TaskRunner {
	return &TaskRunner{wsClient: wsClient}
}

// HandleTask 处理从 WebSocket 收到的任务
func (t *TaskRunner) HandleTask(msg define.AgentWsMessage) {
	gstool.FmtPrintlnLogTime(`收到任务消息 type=%s task_id=%s sse_distribute_id=%s`, msg.Type, msg.TaskID, msg.SseDistributeId)
	dataBytes, err := json.Marshal(msg.Data)
	if err != nil {
		gstool.FmtPrintlnLogTime(`序列化任务数据失败 %s`, err.Error())
		return
	}

	var taskData define.AgentTaskExecuteData
	if err := json.Unmarshal(dataBytes, &taskData); err != nil {
		gstool.FmtPrintlnLogTime(`解析任务数据失败 %s`, err.Error())
		return
	}
	gstool.FmtPrintlnLogTime(`任务数据解析完成 %s`, summarizeAgentTaskData(taskData))

	// 防止并发执行
	if !t.beginTask(taskData.TaskID) {
		gstool.FmtPrintlnLogTime(`任务被拒绝，当前已有执行中任务 current_task=%s reject_task=%s`, t.GetCurrentTaskID(), taskData.TaskID)
		t.wsClient.SendTaskResult(taskData.TaskID, taskData.SseDistributeId, "failed", "Agent正在执行其他任务")
		return
	}

	// 异步执行任务
	go func() {
		defer t.finishTask()
		t.executeTask(taskData)
	}()
}

// executeTask 执行任务
func (t *TaskRunner) executeTask(taskData define.AgentTaskExecuteData) {
	taskID := taskData.TaskID
	sseDistributeId := taskData.SseDistributeId

	gstool.FmtPrintlnLogTime(`开始执行任务 %s`, summarizeAgentTaskData(taskData))

	// 上报 running 状态
	t.wsClient.SendTaskStatus(taskID, sseDistributeId, "running")

	// 检查运行环境是否就绪
	if component.PlaywrightClient.Pw == nil {
		t.wsClient.SendTaskLog(taskID, sseDistributeId, "环境检测", "Playwright 浏览器核心未就绪")
		t.wsClient.SendTaskResult(taskID, sseDistributeId, "failed", "Playwright 浏览器核心未就绪")
		return
	}

	// 构造 StreamFunc：将日志实时回传
	streamFunc := func(name, message string) {
		gstool.FmtPrintlnLogTime(`任务日志 task_id=%s step=%s message=%s`, taskID, name, message)
		t.wsClient.SendTaskLog(taskID, sseDistributeId, name, message)
	}

	// 反序列化 ShowCookies
	showCookies := make([]plw.ShowCookie, 0)
	if taskData.RunParams.ShowCookies != nil {
		cookiesBytes, _ := json.Marshal(taskData.RunParams.ShowCookies)
		_ = json.Unmarshal(cookiesBytes, &showCookies)
	}

	// 构造 PlaywrightRunParams
	runParams := &plw.PlaywrightRunParams{
		Id:                  taskData.RunParams.Id,
		Link:                taskData.RunParams.Link,
		LinkIdLabel:         taskData.RunParams.LinkIdLabel,
		OpenNum:             taskData.RunParams.OpenNum,
		Cookie:              taskData.RunParams.Cookie,
		Headers:             taskData.RunParams.Headers,
		OpenType:            define.OpenType(taskData.RunParams.OpenType),
		CombineType:         taskData.RunParams.CombineType,
		ProcessList:         taskData.RunParams.ProcessList,
		ReplaceList:         taskData.RunParams.ReplaceList,
		BrowserAuthUsername: taskData.RunParams.BrowserAuthUsername,
		BrowserAuthPassword: taskData.RunParams.BrowserAuthPassword,
		Domain:              taskData.RunParams.Domain,
		Scheme:              taskData.RunParams.Scheme,
		LocatorTimeout:      taskData.RunParams.LocatorTimeout,
		GetPageTimeout:      taskData.RunParams.GetPageTimeout,
		LastIndexLabel:      taskData.RunParams.LastIndexLabel,
		LinkId:              taskData.RunParams.LinkId,
		DownloadFinds:       taskData.RunParams.DownloadFinds,
		AutoCloseSecond:     taskData.RunParams.AutoCloseSecond,
		Channel:             taskData.RunParams.Channel,
		StreamFunc:          streamFunc,
		RunCallFunc:         nil,
		ListenCurls:         nil, // Agent 不需要拦截请求功能，nil map 对 range 安全
		FilterUris:          taskData.RunParams.FilterUris,
		ShowCookies:         showCookies,
		// Agent 不连接配置库；历史目录索引通过服务端接口查询和写入。
		SmartLinkLastStore: newAgentSmartLinkLastStore(t.wsClient.config.ServerURL, taskData.SafeToken),
	}

	// 普通任务与抓取任务统一走相同的数据目录策略，确保登录态来源一致。
	if resolvedCombineType, changed := resolveAgentTaskCombineType(taskData.TaskType, runParams.CombineType); changed {
		streamFunc("运行约束", fmt.Sprintf("Agent模式调整数据目录合并类型(原类型=%d，新类型=%d)", runParams.CombineType, resolvedCombineType))
		runParams.CombineType = resolvedCombineType
	}

	// Agent 模式强制使用有头浏览器（headful）
	if runParams.OpenType != define.OpenTypeWebkitChrome {
		streamFunc("运行约束", "Agent模式强制使用有头浏览器模式")
		runParams.OpenType = define.OpenTypeWebkitChrome
	}

	streamFunc("构建run_params", "成功，准备打开的链接："+runParams.Link)

	if taskData.TaskType == define.AgentTaskTypeScrapeToMarkdown {
		gstool.FmtPrintlnLogTime(`进入抓取任务分支 task_id=%s jump_url=%s css_selector=%s`, taskID, taskData.ScrapeConfig.JumpURL, taskData.ScrapeConfig.CssSelector)
		result, err := plw.RunScrapeToMarkdown(runParams, taskData.ScrapeConfig, component.PlaywrightClient.Log)
		if err != nil {
			gstool.FmtPrintlnLogTime(`抓取任务执行失败 task_id=%s err=%s`, taskID, err.Error())
			streamFunc("执行结果", "失败："+err.Error())
			t.wsClient.SendTaskResult(taskID, sseDistributeId, "failed", err.Error())
			return
		}
		gstool.FmtPrintlnLogTime(`抓取任务执行完成 task_id=%s markdown_bytes=%d zip_bytes=%d file_name=%s`, taskID, len(result.Markdown), len(result.ZipBytes), result.FileName)
		uploadResult, err := t.wsClient.UploadScrapeResultFile(taskID, result.FileName, result.ZipBytes, taskData.SafeToken)
		if err != nil {
			gstool.FmtPrintlnLogTime(`抓取结果上传失败 task_id=%s file_name=%s err=%s`, taskID, result.FileName, err.Error())
			streamFunc("上传结果", "失败："+err.Error())
			t.wsClient.SendTaskResult(taskID, sseDistributeId, "failed", err.Error())
			return
		}
		gstool.FmtPrintlnLogTime(`抓取结果上传成功 task_id=%s saved_file_name=%s download_url=%s`, taskID, uploadResult.FileName, uploadResult.DownloadURL)
		streamFunc("上传结果", "成功："+uploadResult.FileName)
		t.wsClient.SendTaskResultData(taskID, sseDistributeId, define.AgentTaskResultData{
			Status:      "succeeded",
			FinishTime:  time.Now().Unix(),
			DownloadURL: uploadResult.DownloadURL,
			FileName:    uploadResult.FileName,
		})
		return
	}

	// 执行普通 Playwright 任务
	p := plw.NewPlaywright(runParams, component.PlaywrightClient.Log)
	openErr := p.Open(&p_common.Call{}, nil)
	if openErr != nil {
		gstool.FmtPrintlnLogTime(`普通任务执行失败 task_id=%s err=%s`, taskID, openErr.Error())
		streamFunc("执行结果", "失败："+openErr.Error())
		t.wsClient.SendTaskResult(taskID, sseDistributeId, "failed", openErr.Error())
		return
	}
	gstool.FmtPrintlnLogTime(`普通任务执行成功 task_id=%s`, taskID)
	streamFunc("浏览器实例执行", "结束")
	t.wsClient.SendTaskResult(taskID, sseDistributeId, "succeeded", "")
}

func (t *TaskRunner) beginTask(taskID string) bool {
	t.mu.Lock()
	defer t.mu.Unlock()
	if t.currentTask != "" {
		return false
	}
	t.currentTask = taskID
	return true
}

func (t *TaskRunner) finishTask() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.currentTask = ""
}

// GetCurrentTaskID 获取当前执行中的任务 ID
func (t *TaskRunner) GetCurrentTaskID() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.currentTask
}
