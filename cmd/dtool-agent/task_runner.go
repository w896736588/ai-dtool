package main

import (
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	"dev_tool/internal/app/dtool/plw"
	"dev_tool/internal/pkg/p_common"
	"encoding/json"
	"fmt"
	"sync"

	"gitee.com/Sxiaobai/gs/v2/gstool"
)

// TaskRunner 管理任务执行
type TaskRunner struct {
	wsClient    *WsClient
	currentTask string
	mu          sync.Mutex
}

// NewTaskRunner 创建任务执行器
func NewTaskRunner(wsClient *WsClient) *TaskRunner {
	return &TaskRunner{wsClient: wsClient}
}

// HandleTask 处理从 WebSocket 收到的任务
func (t *TaskRunner) HandleTask(msg define.AgentWsMessage) {
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

	// 防止并发执行
	t.mu.Lock()
	if t.currentTask != "" {
		t.mu.Unlock()
		t.wsClient.SendTaskResult(taskData.TaskID, taskData.SseDistributeId, "failed", "Agent正在执行其他任务")
		return
	}
	t.currentTask = taskData.TaskID
	t.mu.Unlock()

	defer func() {
		t.mu.Lock()
		t.currentTask = ""
		t.mu.Unlock()
	}()

	// 异步执行任务
	go t.executeTask(taskData)
}

// executeTask 执行任务
func (t *TaskRunner) executeTask(taskData define.AgentTaskExecuteData) {
	taskID := taskData.TaskID
	sseDistributeId := taskData.SseDistributeId

	gstool.FmtPrintlnLogTime(`开始执行任务 task_id=%s`, taskID)

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
		LastIndexLabel:      "", // Agent 模式下置空，避免 DB 依赖
		LinkId:              taskData.RunParams.LinkId,
		DownloadFinds:       taskData.RunParams.DownloadFinds,
		AutoCloseSecond:     taskData.RunParams.AutoCloseSecond,
		Channel:             taskData.RunParams.Channel,
		StreamFunc:          streamFunc,
		RunCallFunc:         nil,
		ListenCurls:         nil, // Agent 不需要拦截请求功能，nil map 对 range 安全
		FilterUris:          taskData.RunParams.FilterUris,
		ShowCookies:         showCookies,
	}

	// Agent 模式约束：使用 CombineTypeNo 避免访问 DB
	// 如果服务端传了需要 DB 的 CombineType，这里强制改为 No
	if runParams.CombineType != define.CombineTypeNo {
		streamFunc("运行约束", fmt.Sprintf("Agent模式不支持数据目录合并(原类型=%d)，已改为每次新建", runParams.CombineType))
		runParams.CombineType = define.CombineTypeNo
	}

	// Agent 模式强制使用有头浏览器（headful）
	if runParams.OpenType != define.OpenTypeWebkitChrome {
		streamFunc("运行约束", "Agent模式强制使用有头浏览器模式")
		runParams.OpenType = define.OpenTypeWebkitChrome
	}

	streamFunc("构建run_params", "成功，准备打开的链接："+runParams.Link)

	// 执行 Playwright 任务
	p := plw.NewPlaywright(runParams, component.PlaywrightClient.Log)

	// Agent 模式下不传 Call（避免 DB 操作）
	openErr := p.Open(&p_common.Call{}, nil)
	if openErr != nil {
		streamFunc("执行结果", "失败："+openErr.Error())
		t.wsClient.SendTaskResult(taskID, sseDistributeId, "failed", openErr.Error())
		return
	}

	streamFunc("浏览器实例执行", "结束")
	t.wsClient.SendTaskResult(taskID, sseDistributeId, "succeeded", "")
}

// GetCurrentTaskID 获取当前执行中的任务 ID
func (t *TaskRunner) GetCurrentTaskID() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.currentTask
}
