package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	_struct "dev_tool/internal/app/dtool/struct"
	"dev_tool/internal/pkg/p_define"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
)

const (
	// asyncTaskTypeHomeTaskDailyReport 标识首页工作日报异步任务。 // asyncTaskTypeHomeTaskDailyReport identifies the home-task daily report async task.
	asyncTaskTypeHomeTaskDailyReport = `home_task_daily_report`
	// asyncTaskTypeMemoryFragmentArrange 标识知识片段整理异步任务。 // asyncTaskTypeMemoryFragmentArrange identifies the memory fragment arrange async task.
	asyncTaskTypeMemoryFragmentArrange = `memory_fragment_arrange`
	// asyncTaskTypeMainDBSync 标识主库自动同步任务。 // asyncTaskTypeMainDBSync identifies the main db auto sync task.
	asyncTaskTypeMainDBSync = `main_db_sync`
	// asyncTaskTypeHomeTaskTapdScrape 标识首页 TAPD 网页抓取异步任务。
	asyncTaskTypeHomeTaskTapdScrape = `home_task_tapd_scrape`

	// asyncTaskActionSaveDailyReport 表示保存日报到知识片段。 // asyncTaskActionSaveDailyReport means saving the report as a memory fragment.
	asyncTaskActionSaveDailyReport = `save_daily_report`
	// asyncTaskActionOverwriteMemoryFragment 表示用整理结果覆盖知识片段。 // asyncTaskActionOverwriteMemoryFragment means overwriting the memory fragment with arranged content.
	asyncTaskActionOverwriteMemoryFragment = `overwrite_memory_fragment`
	// asyncTaskActionOverwriteFragmentWithScrape 表示用抓取结果更新知识片段。
	asyncTaskActionOverwriteFragmentWithScrape = `overwrite_fragment_with_scrape`
	// asyncTaskActionDiscard 表示丢弃结果。 // asyncTaskActionDiscard means discarding the async result.
	asyncTaskActionDiscard = `discard`
)

// asyncTaskBackgroundRunner 允许测试把后台执行切换成同步运行。 // asyncTaskBackgroundRunner allows tests to replace the goroutine runner with a synchronous runner.
var asyncTaskBackgroundRunner = func(_ int, run func()) {
	go run()
}

// buildAsyncHomeTaskDailyReportResult 允许测试替换日报结果构建过程。 // buildAsyncHomeTaskDailyReportResult allows tests to replace the daily report result builder.
var buildAsyncHomeTaskDailyReportResult = defaultBuildAsyncHomeTaskDailyReportResult

// buildAsyncMemoryArrangeResult 允许测试替换知识片段整理结果构建过程。 // buildAsyncMemoryArrangeResult allows tests to replace the memory arrange result builder.
var buildAsyncMemoryArrangeResult = defaultBuildAsyncMemoryArrangeResult

// buildAsyncHomeTaskTapdScrapeResultFn 允许测试替换 TAPD 网页抓取结果构建过程。
var buildAsyncHomeTaskTapdScrapeResultFn = buildAsyncHomeTaskTapdScrapeResult

// AsyncTaskList 查询异步任务列表与汇总。 // AsyncTaskList returns async task summary plus recent items.
func AsyncTaskList(c *gin.Context) {
	db, ok := asyncTaskDBOrResponse(c)
	if !ok {
		return
	}
	request := _struct.AsyncTaskListRequest{}
	_ = gsgin.GinPostBody(c, &request)
	summary, err := db.AsyncTaskSummary(request.Limit)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, summary)
}

// AsyncTaskInfo 查询异步任务详情。 // AsyncTaskInfo returns a single async task detail.
func AsyncTaskInfo(c *gin.Context) {
	db, ok := asyncTaskDBOrResponse(c)
	if !ok {
		return
	}
	request := _struct.AsyncTaskInfoRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := db.AsyncTaskInfo(request.ID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

// AsyncTaskDelete 删除异步任务记录。 // AsyncTaskDelete removes the async task record.
func AsyncTaskDelete(c *gin.Context) {
	db, ok := asyncTaskDBOrResponse(c)
	if !ok {
		return
	}
	request := _struct.AsyncTaskDeleteRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if err := db.AsyncTaskDelete(request.ID); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	// 中文注释：任务删除成功，主动推送更新。
	// English comment: Task deleted, broadcast update.
	BroadcastAsyncTasksUpdate()
	gsgin.GinResponseSuccess(c, ``, nil)
}

// AsyncTaskAction 对异步任务执行确认或丢弃操作。 // AsyncTaskAction applies confirm/discard actions to an async task result.
func AsyncTaskAction(c *gin.Context) {
	db, ok := asyncTaskDBOrResponse(c)
	if !ok {
		return
	}
	request := _struct.AsyncTaskActionRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := db.AsyncTaskInfo(request.ID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	action := strings.TrimSpace(request.Action)
	switch action {
	case asyncTaskActionDiscard:
		if err = db.AsyncTaskMarkFinal(request.ID, common.AsyncTaskStatusRejected); err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		updatedInfo, _ := db.AsyncTaskInfo(request.ID)
		// 中文注释：任务状态变化，主动推送更新。
		// English comment: Task status changed, broadcast update.
		BroadcastAsyncTasksUpdate()
		gsgin.GinResponseSuccess(c, ``, updatedInfo)
		return
	case asyncTaskActionSaveDailyReport:
		resultMap, parseErr := asyncTaskDecodePayload(cast.ToString(info[`result_payload`]))
		if parseErr != nil {
			gsgin.GinResponseError(c, parseErr.Error(), nil)
			return
		}
		if strings.TrimSpace(cast.ToString(info[`task_type`])) != asyncTaskTypeHomeTaskDailyReport {
			gsgin.GinResponseError(c, `异步任务类型不支持保存日报`, nil)
			return
		}
		memoryDB, memoryOk := memoryDBOrResponse(c)
		if !memoryOk {
			return
		}
		memoryInfo, saveErr := memoryDB.MemoryFragmentSave(
			0,
			cast.ToString(resultMap[`report_title`]),
			cast.ToString(resultMap[`markdown`]),
			cast.ToStringSlice(resultMap[`suggested_tags`]),
			``,
		)
		if saveErr != nil {
			gsgin.GinResponseError(c, saveErr.Error(), nil)
			return
		}
		component.MemoryRuntime.ScheduleSync()
		broadcastMemoryFragmentUpsert(memoryInfo)
		if err = db.AsyncTaskMarkFinal(request.ID, common.AsyncTaskStatusConfirmed); err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		updatedInfo, _ := db.AsyncTaskInfo(request.ID)
		updatedInfo[`memory_fragment`] = memoryInfo
		// 中文注释：任务状态变化，主动推送更新。
		// English comment: Task status changed, broadcast update.
		BroadcastAsyncTasksUpdate()
		gsgin.GinResponseSuccess(c, ``, updatedInfo)
		return
	case asyncTaskActionOverwriteMemoryFragment:
		resultMap, parseErr := asyncTaskDecodePayload(cast.ToString(info[`result_payload`]))
		if parseErr != nil {
			gsgin.GinResponseError(c, parseErr.Error(), nil)
			return
		}
		if strings.TrimSpace(cast.ToString(info[`task_type`])) != asyncTaskTypeMemoryFragmentArrange {
			gsgin.GinResponseError(c, `异步任务类型不支持覆盖知识片段`, nil)
			return
		}
		memoryDB, memoryOk := memoryDBOrResponse(c)
		if !memoryOk {
			return
		}
		fragmentID := cast.ToString(resultMap[`fragment_id`])
		existingInfo, existingErr := memoryDB.MemoryFragmentInfo(fragmentID)
		if existingErr != nil {
			gsgin.GinResponseError(c, existingErr.Error(), nil)
			return
		}
		memoryInfo, saveErr := memoryDB.MemoryFragmentSave(
			fragmentID,
			cast.ToString(resultMap[`title`]),
			cast.ToString(resultMap[`arranged_content`]),
			cast.ToStringSlice(existingInfo[`tags`]),
			``,
		)
		if saveErr != nil {
			gsgin.GinResponseError(c, saveErr.Error(), nil)
			return
		}
		component.MemoryRuntime.ScheduleSync()
		broadcastMemoryFragmentUpsert(memoryInfo)
		if err = db.AsyncTaskMarkFinal(request.ID, common.AsyncTaskStatusConfirmed); err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		updatedInfo, _ := db.AsyncTaskInfo(request.ID)
		updatedInfo[`memory_fragment`] = memoryInfo
		// 中文注释：任务状态变化，主动推送更新。
		// English comment: Task status changed, broadcast update.
		BroadcastAsyncTasksUpdate()
		gsgin.GinResponseSuccess(c, ``, updatedInfo)
		return
	case asyncTaskActionOverwriteFragmentWithScrape:
		resultMap, parseErr := asyncTaskDecodePayload(cast.ToString(info[`result_payload`]))
		if parseErr != nil {
			gsgin.GinResponseError(c, parseErr.Error(), nil)
			return
		}
		if strings.TrimSpace(cast.ToString(info[`task_type`])) != asyncTaskTypeHomeTaskTapdScrape {
			gsgin.GinResponseError(c, `异步任务类型不支持抓取覆盖`, nil)
			return
		}
		memoryDB, memoryOk := memoryDBOrResponse(c)
		if !memoryOk {
			return
		}
		fragmentID := cast.ToString(resultMap[`fragment_id`])
		existingInfo, existingErr := memoryDB.MemoryFragmentInfo(fragmentID)
		if existingErr != nil {
			gsgin.GinResponseError(c, existingErr.Error(), nil)
			return
		}
		markdown := cast.ToString(resultMap[`markdown`])
		existingTitle := cast.ToString(existingInfo[`title`])
		existingTags := cast.ToStringSlice(existingInfo[`tags`])
		memoryInfo, saveErr := memoryDB.MemoryFragmentSave(
			fragmentID,
			existingTitle,
			markdown,
			existingTags,
			``,
		)
		if saveErr != nil {
			gsgin.GinResponseError(c, saveErr.Error(), nil)
			return
		}
		component.MemoryRuntime.ScheduleSync()
		broadcastMemoryFragmentUpsert(memoryInfo)
		if err = db.AsyncTaskMarkFinal(request.ID, common.AsyncTaskStatusConfirmed); err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		updatedInfo, _ := db.AsyncTaskInfo(request.ID)
		updatedInfo[`memory_fragment`] = memoryInfo
		BroadcastAsyncTasksUpdate()
		gsgin.GinResponseSuccess(c, ``, updatedInfo)
		return
	default:
		gsgin.GinResponseError(c, `异步任务操作不支持`, nil)
		return
	}
}

// createAsyncTask 创建任务并触发后台执行。 // createAsyncTask creates an async task record and starts background execution.
func createAsyncTask(taskType, title, sourceID string, requestPayload map[string]any, execute func(taskID int)) (map[string]any, error) {
	if common.DbLog == nil || common.DbLog.Client == nil {
		return nil, errors.New(`日志库未初始化`)
	}
	payloadText := gstool.JsonEncode(requestPayload)
	taskInfo, err := common.DbLog.AsyncTaskCreate(taskType, title, sourceID, payloadText)
	if err != nil {
		return nil, err
	}
	taskID := cast.ToInt(taskInfo[`id`])
	// 中文注释：新任务创建成功，主动推送更新。
	// English comment: New task created, broadcast update.
	BroadcastAsyncTasksUpdate()
	asyncTaskBackgroundRunner(taskID, func() {
		execute(taskID)
	})
	return taskInfo, nil
}

// runAsyncTaskAndPersistResult 统一处理后台执行状态流转。 // runAsyncTaskAndPersistResult centralizes async background status transitions and result persistence.
func runAsyncTaskAndPersistResult(taskID int, builder func() (map[string]any, error)) {
	if common.DbLog == nil || common.DbLog.Client == nil {
		return
	}
	appendAsyncTaskRunLog(taskID, "调度任务", "准备进入后台执行")
	if err := common.DbLog.AsyncTaskMarkRunning(taskID); err != nil {
		return
	}
	appendAsyncTaskRunLog(taskID, "调度任务", "任务已标记为运行中")
	// 中文注释：任务状态变为 running，主动推送更新。
	// English comment: Task status changed to running, broadcast update.
	BroadcastAsyncTasksUpdate()
	resultMap, err := builder()
	if err != nil {
		appendAsyncTaskRunLog(taskID, "执行结果", "失败: "+err.Error())
		_ = common.DbLog.AsyncTaskMarkFailed(taskID, err.Error())
		// 中文注释：任务状态变为 failed，主动推送更新。
		// English comment: Task status changed to failed, broadcast update.
		BroadcastAsyncTasksUpdate()
		return
	}
	appendAsyncTaskRunLog(taskID, "执行结果", "执行成功，等待用户确认")
	resultPayload := gstool.JsonEncode(resultMap)
	if markErr := common.DbLog.AsyncTaskMarkAwaitConfirm(taskID, resultPayload); markErr != nil {
		return
	}
	// 中文注释：任务状态变为 await_confirm，主动推送更新。
	// English comment: Task status changed to await_confirm, broadcast update.
	BroadcastAsyncTasksUpdate()
}

func appendAsyncTaskRunLog(taskID int, step, message string) {
	if common.DbLog == nil || common.DbLog.Client == nil {
		return
	}
	if err := common.DbLog.AsyncTaskAppendRunLog(taskID, step, message); err != nil {
		gstool.FmtPrintlnLogTime(`[AsyncTaskRunLog] 写入运行日志失败 task_id=%d step=%s err=%s`, taskID, step, err.Error())
	}
}

// asyncTaskDBOrResponse 返回日志库实例。 // asyncTaskDBOrResponse returns the log database instance or writes an error response.
func asyncTaskDBOrResponse(c *gin.Context) (*common.CSqlite, bool) {
	if common.DbLog == nil || common.DbLog.Client == nil {
		gsgin.GinResponseError(c, `日志库未初始化`, nil)
		return nil, false
	}
	return common.DbLog, true
}

// asyncTaskDecodePayload 解析任务 JSON 结果。 // asyncTaskDecodePayload decodes the task JSON payload.
func asyncTaskDecodePayload(payload string) (map[string]any, error) {
	result := make(map[string]any)
	if strings.TrimSpace(payload) == `` {
		return result, nil
	}
	if err := json.Unmarshal([]byte(payload), &result); err != nil {
		return nil, fmt.Errorf(`异步任务结果解析失败 %w`, err)
	}
	return result, nil
}

// buildAsyncTasksPayload 构造异步任务列表与汇总数据。
// buildAsyncTasksPayload builds the async task summary and list payload.
func buildAsyncTasksPayload(limit int) (map[string]any, error) {
	if common.DbLog == nil || common.DbLog.Client == nil {
		return nil, errors.New(`日志库未初始化`)
	}
	summary, err := common.DbLog.AsyncTaskSummary(limit)
	if err != nil {
		return nil, err
	}
	return summary, nil
}

// sendAsyncTasksSnapshot 向指定 SSE 连接发送一次异步任务状态快照。
// sendAsyncTasksSnapshot sends one async-tasks snapshot to the provided SSE client.
func sendAsyncTasksSnapshot(sse *gsgin.Sse) {
	if sse == nil {
		return
	}
	data, err := buildAsyncTasksPayload(20)
	if err != nil {
		gstool.FmtPrintlnLogTime(`AsyncTasks广播错误 %s`, err.Error())
		return
	}
	err = sse.SendToChan(gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseAsyncTasks,
		Data:            data,
		Type:            p_define.SseContentTypeMsg,
	}))
	if err != nil {
		gstool.FmtPrintlnLogTime(`AsyncTasks广播错误 %s`, err.Error())
	}
}

// BindAsyncTasksSSE 为普通 SSE client 绑定异步任务状态推送。
// BindAsyncTasksSSE attaches async-tasks events to a normal SSE client_id stream.
func BindAsyncTasksSSE(sse *gsgin.Sse, stopC chan int, interval time.Duration) {
	if sse == nil {
		return
	}
	if interval <= 0 {
		interval = 5 * time.Second
	}
	// 中文注释：建连后立即推一次，避免前端初次打开时要等下一个周期。
	// English comment: Push once immediately so the UI does not wait for the first ticker tick.
	sendAsyncTasksSnapshot(sse)
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				sendAsyncTasksSnapshot(sse)
			case <-stopC:
				return
			}
		}
	}()
}

// BroadcastAsyncTasksUpdate 主动广播异步任务状态更新。
// BroadcastAsyncTasksUpdate broadcasts async task status update to all connected SSE clients.
func BroadcastAsyncTasksUpdate() {
	sendAsyncTasksSnapshot(gsgin.SseGetByClientId(define.SseAsyncTasks))
}

// defaultBuildAsyncHomeTaskDailyReportResult 构建日报异步任务结果。 // defaultBuildAsyncHomeTaskDailyReportResult builds the daily report async result.
func defaultBuildAsyncHomeTaskDailyReportResult(taskList []map[string]any, reportTime int64) (map[string]any, error) {
	reportAt := time.Unix(reportTime, 0)
	modelID, prompt, err := homeTaskDailyReportConfig()
	if err != nil {
		return nil, err
	}
	userPrompt, err := buildHomeTaskDailyReportUserPrompt(prompt, taskList, reportAt)
	if err != nil {
		return nil, err
	}
	result, modelInfo, err := common.DbMain.AIChatByModel(modelID, homeTaskDailyReportSystemPrompt(), userPrompt)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		`markdown`:       stripMarkdownCodeFence(result),
		`report_title`:   buildHomeTaskDailyReportTitle(reportAt),
		`prompt`:         prompt,
		`model_id`:       modelID,
		`model`:          modelInfo[`model`],
		`suggested_tags`: []string{homeTaskDailyReportMemoryTag},
	}, nil
}

// defaultBuildAsyncMemoryArrangeResult 构建知识片段整理异步任务结果。 // defaultBuildAsyncMemoryArrangeResult builds the memory fragment arrange async result.
func defaultBuildAsyncMemoryArrangeResult(title, content string) (map[string]any, error) {
	modelID, prompt, err := memoryArrangeConfig()
	if err != nil {
		return nil, err
	}
	userPrompt := buildMemoryArrangeUserPrompt(prompt, title, content)
	result, modelInfo, err := common.DbMain.AIChatByModel(modelID, memoryArrangeSystemPrompt(), userPrompt)
	if err != nil {
		return nil, err
	}
	return map[string]any{
		`title`:            title,
		`original_content`: content,
		`arranged_content`: stripMarkdownCodeFence(result),
		`prompt`:           prompt,
		`model_id`:         modelID,
		`model`:            cast.ToString(modelInfo[`model`]),
	}, nil
}

// buildAsyncHomeTaskTapdScrapeResult 构建首页 TAPD 网页抓取异步任务结果。
func buildAsyncHomeTaskTapdScrapeResult(tapdURL, memoryFragmentID string) (map[string]any, error) {
	return buildAsyncHomeTaskTapdScrapeResultWithLog(tapdURL, memoryFragmentID, nil)
}

func buildAsyncHomeTaskTapdScrapeResultWithLog(tapdURL, memoryFragmentID string, logStep func(string, string)) (map[string]any, error) {
	return buildAsyncHomeTaskRequirementScrapeResultWithLog(`tapd`, tapdURL, memoryFragmentID, logStep)
}

func buildAsyncHomeTaskZentaoScrapeResultWithLog(zentaoURL, memoryFragmentID string, logStep func(string, string)) (map[string]any, error) {
	return buildAsyncHomeTaskRequirementScrapeResultWithLog(`zentao`, zentaoURL, memoryFragmentID, logStep)
}

func buildAsyncHomeTaskRequirementScrapeResultWithLog(fetchType string, sourceURL string, memoryFragmentID string, logStep func(string, string)) (map[string]any, error) {
	sourceName := requirementFetchSourceName(fetchType)
	if logStep != nil {
		logStep("读取配置", "开始读取 "+sourceName+" 抓取配置")
	}
	smartLinkIDKey, linkLabelKey, cssSelectorKey, waitSecondsKey := requirementFetchConfigKeys(fetchType)
	smartLinkIDStr, err := homeTaskConfigValue(smartLinkIDKey)
	if err != nil {
		return nil, fmt.Errorf("读取 %s 自定义网页配置失败: %w", sourceName, err)
	}
	smartLinkID := cast.ToInt(smartLinkIDStr)
	if smartLinkID <= 0 {
		return nil, fmt.Errorf("%s 自定义网页未配置", sourceName)
	}

	linkLabel, err := homeTaskConfigValue(linkLabelKey)
	if err != nil {
		return nil, fmt.Errorf("读取 %s 链接标签配置失败: %w", sourceName, err)
	}
	if strings.TrimSpace(linkLabel) == "" {
		return nil, fmt.Errorf("%s 链接标签未配置", sourceName)
	}

	cssSelector, err := homeTaskConfigValue(cssSelectorKey)
	if err != nil {
		return nil, fmt.Errorf("读取 %s CSS 选择器配置失败: %w", sourceName, err)
	}
	if strings.TrimSpace(cssSelector) == "" {
		return nil, fmt.Errorf("%s CSS 选择器未配置", sourceName)
	}

	waitSecondsStr, err := homeTaskConfigValue(waitSecondsKey)
	if err != nil {
		return nil, fmt.Errorf("读取 %s 等待秒数配置失败: %w", sourceName, err)
	}
	waitSeconds := cast.ToInt(waitSecondsStr)
	if waitSeconds <= 0 {
		waitSeconds = defaultSmartLinkScrapeWaitSeconds
	}
	if waitSeconds > maxSmartLinkScrapeWaitSeconds {
		waitSeconds = maxSmartLinkScrapeWaitSeconds
	}

	if logStep != nil {
		logStep("下发任务", fmt.Sprintf("准备下发 Agent 抓取任务 smart_link_id=%d label=%s wait_seconds=%d", smartLinkID, linkLabel, waitSeconds))
	}
	resultFile, dispatchErr := dispatchScrapeTaskAndAwait(smartLinkID, linkLabel, sourceURL, cssSelector, waitSeconds)
	if dispatchErr != nil {
		return nil, dispatchErr
	}
	if logStep != nil {
		logStep("下发任务", "Agent 已返回抓取结果: "+resultFile.FileName)
		logStep("处理结果", "开始解压 ZIP 并转换图片路径")
	}

	zipResult, zipErr := processScrapeZipResult(resultFile.DownloadURL, memoryFragmentID)
	if zipErr != nil {
		return nil, zipErr
	}
	if logStep != nil {
		logStep("处理结果", "ZIP 处理完成")
	}

	return zipResult, nil
}

func requirementFetchConfigKeys(fetchType string) (string, string, string, string) {
	if strings.TrimSpace(strings.ToLower(fetchType)) == `zentao` {
		return define.HomeTaskConfigZentaoSmartLinkID, define.HomeTaskConfigZentaoLinkLabel, define.HomeTaskConfigZentaoCssSelector, define.HomeTaskConfigZentaoWaitSeconds
	}
	return define.HomeTaskConfigTapdSmartLinkID, define.HomeTaskConfigTapdLinkLabel, define.HomeTaskConfigTapdCssSelector, define.HomeTaskConfigTapdWaitSeconds
}

func requirementFetchSourceName(fetchType string) string {
	if strings.TrimSpace(strings.ToLower(fetchType)) == `zentao` {
		return `禅道`
	}
	return `TAPD`
}

// AsyncTaskRetry 重试失败的异步任务，根据任务类型从 request_payload 提取参数并重新调度。 // AsyncTaskRetry retries a failed async task by resetting its status and re-dispatching.
func AsyncTaskRetry(c *gin.Context) {
	db, ok := asyncTaskDBOrResponse(c)
	if !ok {
		return
	}
	request := _struct.AsyncTaskInfoRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := db.AsyncTaskInfo(request.ID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	taskType := strings.TrimSpace(cast.ToString(info[`task_type`]))
	taskID := cast.ToInt(info[`id`])
	requestPayloadMap, parseErr := asyncTaskDecodePayload(cast.ToString(info[`request_payload`]))
	if parseErr != nil {
		gsgin.GinResponseError(c, `解析任务参数失败: `+parseErr.Error(), nil)
		return
	}
	// 重置任务状态为 pending。
	if resetErr := db.AsyncTaskResetForRetry(taskID); resetErr != nil {
		gsgin.GinResponseError(c, resetErr.Error(), nil)
		return
	}
	BroadcastAsyncTasksUpdate()
	// 根据任务类型重新调度后台执行。
	asyncTaskBackgroundRunner(taskID, func() {
		switch taskType {
		case asyncTaskTypeHomeTaskTapdScrape:
			tapdURL := cast.ToString(requestPayloadMap[`tapd_url`])
			memoryFragmentID := cast.ToString(requestPayloadMap[`memory_fragment_id`])
			runAsyncTaskAndPersistResult(taskID, func() (map[string]any, error) {
				return buildAsyncHomeTaskTapdScrapeResultWithLog(tapdURL, memoryFragmentID, func(step, message string) {
					appendAsyncTaskRunLog(taskID, step, message)
					BroadcastAsyncTasksUpdate()
				})
			})
		case asyncTaskTypeMemoryFragmentArrange:
			fragmentID := cast.ToString(requestPayloadMap[`fragment_id`])
			title := cast.ToString(requestPayloadMap[`title`])
			content := cast.ToString(requestPayloadMap[`content`])
			runAsyncTaskAndPersistResult(taskID, func() (map[string]any, error) {
				resultMap, buildErr := buildAsyncMemoryArrangeResult(title, content)
				if buildErr != nil {
					return nil, buildErr
				}
				resultMap[`fragment_id`] = fragmentID
				return resultMap, nil
			})
		case asyncTaskTypeHomeTaskDailyReport:
			taskList, listErr := common.DbMain.HomeTaskListTodayUpdated()
			if listErr != nil {
				_ = db.AsyncTaskMarkFailed(taskID, listErr.Error())
				BroadcastAsyncTasksUpdate()
				return
			}
			reportTime := time.Now().Unix()
			runAsyncTaskAndPersistResult(taskID, func() (map[string]any, error) {
				return buildAsyncHomeTaskDailyReportResult(taskList, reportTime)
			})
		default:
			_ = db.AsyncTaskMarkFailed(taskID, `不支持重试的任务类型: `+taskType)
			BroadcastAsyncTasksUpdate()
		}
	})
	updatedInfo, _ := db.AsyncTaskInfo(taskID)
	gsgin.GinResponseSuccess(c, ``, updatedInfo)
}
