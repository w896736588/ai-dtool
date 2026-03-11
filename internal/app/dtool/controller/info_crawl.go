package controller

import (
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	_struct "dev_tool/internal/app/dtool/struct"
	"dev_tool/internal/pkg/p_sse"
	"fmt"
	"runtime/debug"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
)

// InfoCrawlCrawl4AIStatus 查询 Crawl4AI 初始化状态。
func InfoCrawlCrawl4AIStatus(c *gin.Context) {
	if component.Crawl4AIClient == nil {
		gsgin.GinResponseError(c, `Crawl4AI 服务未初始化`, nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, component.Crawl4AIClient.Status())
}

// InfoCrawlTaskList 查询信息抓取任务列表。
func InfoCrawlTaskList(c *gin.Context) {
	list, err := common.DbMain.InfoCrawlTaskList()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`task_list`: list,
	})
}

// InfoCrawlTaskInfo 查询信息抓取任务详情。
func InfoCrawlTaskInfo(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if cast.ToInt(dataMap[`id`]) <= 0 {
		gsgin.GinResponseError(c, `任务id不能为空`, nil)
		return
	}
	info, err := common.DbMain.InfoCrawlTaskInfo(cast.ToInt(dataMap[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

// InfoCrawlTaskSave 保存信息抓取任务。
func InfoCrawlTaskSave(c *gin.Context) {
	request := _struct.InfoCrawlTaskSaveRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := common.DbMain.InfoCrawlTaskSave(request.ID, request.Name, request.Prompt, request.AiModelID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, info)
}

// InfoCrawlTaskDelete 删除信息抓取任务。
func InfoCrawlTaskDelete(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	if err := common.DbMain.InfoCrawlTaskDelete(cast.ToInt(dataMap[`id`])); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, nil)
}

// InfoCrawlTaskRun 执行信息抓取任务。
func InfoCrawlTaskRun(c *gin.Context) {
	request := _struct.InfoCrawlTaskRunRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if request.TaskID <= 0 {
		gsgin.GinResponseError(c, `任务id不能为空`, nil)
		return
	}
	if component.Crawl4AIClient == nil {
		gsgin.GinResponseError(c, `Crawl4AI 服务未初始化`, nil)
		return
	}
	crawlStatus := component.Crawl4AIClient.Status()
	if cast.ToBool(crawlStatus[`is_installing`]) {
		gsgin.GinResponseError(c, cast.ToString(crawlStatus[`status_text`]), crawlStatus)
		return
	}
	if !cast.ToBool(crawlStatus[`is_ready`]) {
		if cast.ToString(crawlStatus[`status`]) == define.Crawl4AIStatusFailed {
			gsgin.GinResponseError(c, cast.ToString(crawlStatus[`error_message`]), crawlStatus)
			return
		}
		component.Crawl4AIClient.EnsureReadyAsync()
		gsgin.GinResponseError(c, `Crawl4AI 正在初始化，请稍后重试`, component.Crawl4AIClient.Status())
		return
	}
	taskInfo, err := common.DbMain.InfoCrawlTaskRow(request.TaskID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	modelInfo, err := common.DbMain.InfoCrawlAiModelInfo(cast.ToInt(taskInfo[`ai_model_id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	runID, err := common.DbMain.InfoCrawlRunCreate(request.TaskID, taskInfo, modelInfo)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	sse := &p_sse.SseShell{
		Sse:             gsgin.SseGetByClientId(c.GetHeader(`SseClientId`)),
		SseDistributeId: request.SseDistributeID,
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`run_id`:      runID,
		`status`:      define.InfoCrawlRunStatusRunning,
		`run_message`: `任务已提交，正在后台执行`,
	})
	go runInfoCrawlTaskAsync(runID, taskInfo, sse)
}

// runInfoCrawlTaskAsync 异步执行信息抓取任务。
func runInfoCrawlTaskAsync(runID int, taskInfo map[string]any, sse *p_sse.SseShell) {
	defer func() {
		if recoverErr := recover(); recoverErr != nil {
			stackText := string(debug.Stack())
			errorMessage := fmt.Sprintf(`后台执行异常：%v`, recoverErr)
			_ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
				`status`:        define.InfoCrawlRunStatusFailed,
				`run_message`:   `执行失败`,
				`error_message`: errorMessage,
			})
			sse.Send(errorMessage, define.InfoCrawlSseTypeError)
			gstool.FmtPrintlnLogTime(`info crawl async panic run_id=%d err=%v stack=%s`, runID, recoverErr, stackText)
		}
	}()
	_ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
		`status`:        define.InfoCrawlRunStatusRunning,
		`run_message`:   `任务已提交，正在后台执行`,
		`error_message`: ``,
	})
	sse.Send(`任务已提交`, define.InfoCrawlSseTypeStatus)
	if component.Crawl4AIClient == nil {
		_ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
			`status`:        define.InfoCrawlRunStatusFailed,
			`run_message`:   `执行失败`,
			`error_message`: `Crawl4AI 服务未初始化`,
		})
		sse.Send(`Crawl4AI 服务未初始化`, define.InfoCrawlSseTypeError)
		return
	}
	urlList := component.Crawl4AIClient.ExtractURLs(cast.ToString(taskInfo[`prompt`]))
	if len(urlList) == 0 {
		_ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
			`status`:        define.InfoCrawlRunStatusFailed,
			`run_message`:   `执行失败`,
			`error_message`: `提示词中未找到可采集的网址，请在提示词中附带 http/https 链接`,
		})
		sse.Send(`提示词中未找到可采集的网址，请在提示词中附带 http/https 链接`, define.InfoCrawlSseTypeError)
		return
	}
	sse.Send(`已识别到 `+cast.ToString(len(urlList))+` 个网址，正在使用 Crawl4AI 采集`, define.InfoCrawlSseTypeStatus)
	crawlResultList, err := component.Crawl4AIClient.CrawlURLs(urlList, 10*time.Minute)
	if err != nil {
		_ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
			`status`:        define.InfoCrawlRunStatusFailed,
			`run_message`:   `执行失败`,
			`error_message`: err.Error(),
		})
		sse.Send(err.Error(), define.InfoCrawlSseTypeError)
		return
	}
	successCount := 0
	for _, item := range crawlResultList {
		if item.Success {
			successCount++
			sse.Send(`采集成功：`+item.URL, define.InfoCrawlSseTypeStatus)
		} else {
			sse.Send(`采集失败：`+item.URL+` `+item.Error, define.InfoCrawlSseTypeStatus)
		}
	}
	if successCount == 0 {
		_ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
			`status`:        define.InfoCrawlRunStatusFailed,
			`run_message`:   `执行失败`,
			`error_message`: `Crawl4AI 未成功采集任何网页`,
		})
		sse.Send(`Crawl4AI 未成功采集任何网页`, define.InfoCrawlSseTypeError)
		return
	}
	sse.Send(`正在连接 AI`, define.InfoCrawlSseTypeStatus)
	content, _, err := common.DbMain.InfoCrawlChatStreamByModel(
		cast.ToInt(taskInfo[`ai_model_id`]),
		common.DbMain.InfoCrawlSystemPrompt(),
		common.DbMain.InfoCrawlBuildUserPrompt(taskInfo, crawlResultList),
		func(chunk string) {
			if strings.TrimSpace(chunk) == `` {
				return
			}
			sse.Send(chunk, define.InfoCrawlSseTypeChunk)
		},
	)
	if err != nil {
		_ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
			`status`:        define.InfoCrawlRunStatusFailed,
			`run_message`:   `执行失败`,
			`error_message`: err.Error(),
		})
		sse.Send(err.Error(), define.InfoCrawlSseTypeError)
		return
	}
	sse.Send(`正在写入执行结果`, define.InfoCrawlSseTypeStatus)
	runMessage := `执行完成`
	if strings.TrimSpace(content) == `` {
		runMessage = `执行完成，但模型未返回内容`
	}
	_ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
		`status`:         define.InfoCrawlRunStatusSuccess,
		`run_message`:    runMessage,
		`output_content`: content,
		`error_message`:  ``,
	})
	sse.Send(runMessage, define.InfoCrawlSseTypeDone)
}

// InfoCrawlRunList 查询执行历史。
func InfoCrawlRunList(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	taskID := cast.ToInt(dataMap[`task_id`])
	if taskID <= 0 {
		gsgin.GinResponseError(c, `任务id不能为空`, nil)
		return
	}
	list, err := common.DbMain.InfoCrawlRunList(taskID, cast.ToInt(dataMap[`limit`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`run_list`: list,
	})
}

// InfoCrawlRunInfo 查询执行详情。
func InfoCrawlRunInfo(c *gin.Context) {
	dataMap := make(map[string]any)
	_ = gsgin.GinPostBody(c, &dataMap)
	id := cast.ToInt(dataMap[`id`])
	if id <= 0 {
		gsgin.GinResponseError(c, `执行记录id不能为空`, nil)
		return
	}
	info, err := common.DbMain.InfoCrawlRunInfo(id)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, info)
}
