package controller

import (
	"bytes"
	"context"
	"dev_tool/internal/app/dtool/api"
	"dev_tool/internal/app/dtool/business"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	_struct "dev_tool/internal/app/dtool/struct"
	"dev_tool/internal/pkg/p_claude"
	"dev_tool/internal/pkg/p_codex"
	"dev_tool/internal/pkg/p_define"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"github.com/w896736588/go-tool/gsgin"
	"github.com/w896736588/go-tool/gstool"
)

const (
	taskWorkflowDevPlanDefaultTitleSuffix = `-开发执行`
	taskWorkflowDevPlanDefaultTemplate    = "# 开发执行说明\n\n## 需求摘要\n\n## 开发方案\n\n## 涉及接口\n\n## 数据影响\n\n## 风险与限制\n\n## 实施记录\n"
	taskWorkflowDevPlanDefaultTag         = `开发执行`
	taskWorkflowRunTypeCoverageGenerate   = `coverage_generate`
	taskWorkflowRunTypeTestPlanGenerate   = `test_plan_generate`
	taskWorkflowRunTypeAPIExecute         = `api_test_execute`
	taskWorkflowRunTypeUIAssistGenerate   = `ui_assist_generate`
)

var taskWorkflowAPIPathReg = regexp.MustCompile(`/api/[A-Za-z0-9_./:-]+`)
var taskWorkflowJSONFenceReg = regexp.MustCompile("(?s)```json\\s*(.*?)\\s*```")
var taskWorkflowURLWithQueryReg = regexp.MustCompile(`(?:https?://[^\s)]+|/api/[^\s)]+)\?[^\s)]+`)

// cachedPythonCmd 缓存找到的 Python 可执行文件路径。
var cachedPythonCmd string

// chatCancelFuncs 存储运行中对话的 cancel 函数，key 为 chatID。
var chatCancelFuncs sync.Map

// agentCliSseConns 存储 AgentCli 业务 SSE 连接，key 为 clientID，value 为 *gsgin.Sse。
// 每个 clientID 只有一条连接，新连接会替换旧连接。
var agentCliSseConns sync.Map

// taskWorkflowSseConns 存储 TaskWorkflow 业务 SSE 连接，key 为 clientID，value 为 *gsgin.Sse。
// 每个 clientID 只有一条连接，新连接会替换旧连接。
var taskWorkflowSseConns sync.Map

// AgentCliChatSseOpen 是 /sse/agent_cli 的 SSE 连接建立函数。
func AgentCliChatSseOpen(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
	clientID := strings.TrimSpace(urlValues.Get(`client_id`))
	if clientID == `` {
		return nil, fmt.Errorf(`client_id 不能为空`)
	}
	// 先关闭同 clientID 的旧连接，确保每 clientID 只有一条活跃连接
	if old, ok := agentCliSseConns.LoadAndDelete(clientID); ok {
		if oldSse, ok2 := old.(*gsgin.Sse); ok2 && oldSse != nil {
			gstool.FmtPrintlnLogTime(`[agent-cli-sse] 替换旧连接 clientID=%s old_connID=%s`, clientID, oldSse.ClientId)
			oldSse.UnRegister()
		}
	}
	connID := fmt.Sprintf("agent_cli_sse_%s_%d", clientID, time.Now().UnixNano())
	sse := gsgin.SseRegister(connID, stopC, c)
	agentCliSseConns.Store(clientID, sse)
	gstool.FmtPrintlnLogTime("[agent-cli-sse] clientID=%s 已建立 SSE 连接", clientID)
	return sse, nil
}

// AgentCliChatSseClose 是 /sse/agent_cli 的 SSE 连接关闭函数。
func AgentCliChatSseClose(sse *gsgin.Sse) {
	if sse == nil {
		return
	}
	agentCliSseConns.Range(func(key, value any) bool {
		if value == sse {
			agentCliSseConns.Delete(key)
			gstool.FmtPrintlnLogTime(`[agent-cli-sse] 连接关闭 connID=%s clientID=%s`, sse.ClientId, key)
			return false
		}
		return true
	})
	sse.UnRegister()
}

// TaskWorkflowChatSseOpen 是 /sse/task_workflow 的 SSE 连接建立函数。
func TaskWorkflowChatSseOpen(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
	clientID := strings.TrimSpace(urlValues.Get(`client_id`))
	if clientID == `` {
		return nil, fmt.Errorf(`client_id 不能为空`)
	}
	// 先关闭同 clientID 的旧连接，确保每 clientID 只有一条活跃连接
	if old, ok := taskWorkflowSseConns.LoadAndDelete(clientID); ok {
		if oldSse, ok2 := old.(*gsgin.Sse); ok2 && oldSse != nil {
			gstool.FmtPrintlnLogTime(`[task-workflow-sse] 替换旧连接 clientID=%s old_connID=%s`, clientID, oldSse.ClientId)
			oldSse.UnRegister()
		}
	}
	connID := fmt.Sprintf("task_workflow_sse_%s_%d", clientID, time.Now().UnixNano())
	sse := gsgin.SseRegister(connID, stopC, c)
	taskWorkflowSseConns.Store(clientID, sse)
	gstool.FmtPrintlnLogTime("[task-workflow-sse] clientID=%s 已建立 SSE 连接", clientID)
	return sse, nil
}

// TaskWorkflowChatSseClose 是 /sse/task_workflow 的 SSE 连接关闭函数。
func TaskWorkflowChatSseClose(sse *gsgin.Sse) {
	if sse == nil {
		return
	}
	taskWorkflowSseConns.Range(func(key, value any) bool {
		if value == sse {
			taskWorkflowSseConns.Delete(key)
			gstool.FmtPrintlnLogTime(`[task-workflow-sse] 连接关闭 connID=%s clientID=%s`, sse.ClientId, key)
			return false
		}
		return true
	})
	sse.UnRegister()
}

// broadcastChatLineToBusinessSse 将对话输出行广播到对应业务的 SSE 连接。
// 根据 chat 的 from_type 确定推送到 agent_cli 还是 task_workflow 的业务 SSE 通道。
func broadcastChatLineToBusinessSse(chatID int64, fromType string, line string) {
	var distributeID string
	var conns *sync.Map

	switch fromType {
	case common.AgentChatSourceTypeAgentCli:
		distributeID = define.SseAgentCliChatOutput
		conns = &agentCliSseConns
	default:
		distributeID = define.SseTaskWorkflowChatOutput
		conns = &taskWorkflowSseConns
	}

	msg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: distributeID,
		Data: map[string]any{
			`chat_id`: chatID,
			`line`:    line,
		},
		Type: p_define.SseContentTypeMsg,
	})

	conns.Range(func(key, value any) bool {
		if sse, ok := value.(*gsgin.Sse); ok {
			_ = sse.SendToChan(msg)
		}
		return true
	})
}

// TaskWorkflowCreateOrGet 查询或创建任务工作流。
func TaskWorkflowCreateOrGet(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.TaskWorkflowCreateOrGetRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := common.DbMain.TaskWorkflowCreateOrGetByHomeTaskID(request.HomeTaskID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	responseData, err := buildTaskWorkflowResponse(c, info)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, responseData)
}

// TaskWorkflowInfo 查询任务工作流详情。
func TaskWorkflowInfo(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.TaskWorkflowInfoRequest{}
	_ = gsgin.GinPostBody(c, &request)
	info, err := common.DbMain.TaskWorkflowInfo(request.WorkflowID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	responseData, err := buildTaskWorkflowResponse(c, info)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, responseData)
}

// TaskWorkflowDevPlanInit 初始化开发执行文档片段。
func TaskWorkflowDevPlanInit(c *gin.Context) {
	workflowInfo, memoryDB, homeTaskInfo, ok := taskWorkflowLoadContextForDevPlan(c)
	if !ok {
		return
	}
	fragmentInfo, err := ensureTaskWorkflowDevPlanFragment(workflowInfo, homeTaskInfo, memoryDB)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`: workflowInfo,
		`fragment`: fragmentInfo,
	})
}

// TaskWorkflowDevPlanInfo 查询开发执行文档片段详情。
func TaskWorkflowDevPlanInfo(c *gin.Context) {
	workflowInfo, memoryDB, _, ok := taskWorkflowLoadContextForDevPlan(c)
	if !ok {
		return
	}
	workflowID := cast.ToInt(workflowInfo[`id`])
	fileID := taskWorkflowGetDocFragmentFileID(workflowID, common.TaskWorkflowDocTypeDevPlan)
	if fileID == `` {
		gsgin.GinResponseError(c, `开发执行文档未初始化`, nil)
		return
	}
	fragmentInfo, err := memoryDB.MemoryFragmentInfo(fileID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`: workflowInfo,
		`fragment`: fragmentInfo,
	})
}

// TaskWorkflowDevPlanSave 保存开发执行文档内容。
func TaskWorkflowDevPlanSave(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.TaskWorkflowDevPlanSaveRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if strings.TrimSpace(request.Content) == `` {
		gsgin.GinResponseError(c, `开发执行内容不能为空`, nil)
		return
	}
	workflowInfo, memoryDB, homeTaskInfo, ok := taskWorkflowLoadContextForDevPlanByID(c, request.WorkflowID)
	if !ok {
		return
	}
	fragmentInfo, err := ensureTaskWorkflowDevPlanFragment(workflowInfo, homeTaskInfo, memoryDB)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	savedInfo, err := memoryDB.MemoryFragmentSave(
		cast.ToString(fragmentInfo[`file_id`]),
		cast.ToString(fragmentInfo[`title`]),
		request.Content,
		cast.ToStringSlice(fragmentInfo[`tags`]),
		cast.ToString(fragmentInfo[`folder_name`]),
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
	// 广播片段变更 SSE 事件，使前端文档 Tab 实时更新
	broadcastMemoryFragmentUpsert(savedInfo)
	updatedWorkflowInfo, err := common.DbMain.TaskWorkflowInfo(cast.ToInt(workflowInfo[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`: updatedWorkflowInfo,
		`fragment`: savedInfo,
	})
}

// TaskWorkflowCoverageGenerate 生成覆盖分析并落历史记录。
func TaskWorkflowCoverageGenerate(c *gin.Context) {
	workflowInfo, requirementFragment, devPlanFragment, ok := taskWorkflowLoadGenerationContext(c)
	if !ok {
		return
	}
	requirementContent := cast.ToString(requirementFragment[`content`])
	devPlanContent := cast.ToString(devPlanFragment[`content`])
	uiAssistInfo := taskWorkflowLatestUIAssistReport(cast.ToInt(workflowInfo[`id`]))
	coverageReport, testPlan := taskWorkflowBuildCoverageAndPlan(cast.ToInt(workflowInfo[`id`]), requirementContent, devPlanContent, uiAssistInfo)
	runInfo, err := common.DbMain.TaskWorkflowCreateRun(
		cast.ToInt(workflowInfo[`id`]),
		taskWorkflowRunTypeCoverageGenerate,
		`manual`,
		requirementContent,
		devPlanContent,
		coverageReport,
		testPlan,
		map[string]any{},
		taskWorkflowBuildSummaryMarkdown(coverageReport, testPlan),
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	updatedWorkflowInfo, err := common.DbMain.TaskWorkflowInfo(cast.ToInt(workflowInfo[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`:        updatedWorkflowInfo,
		`coverage_report`: coverageReport,
		`test_run`:        taskWorkflowNormalizeRunInfo(runInfo),
	})
}

// TaskWorkflowUIAssistGenerate 触发页面辅助识别抓取并记录结果。
func TaskWorkflowUIAssistGenerate(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.TaskWorkflowUIAssistGenerateRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if request.WorkflowID <= 0 {
		gsgin.GinResponseError(c, `workflow_id不能为空`, nil)
		return
	}
	if request.SmartLinkID <= 0 {
		gsgin.GinResponseError(c, `smart_link_id不能为空`, nil)
		return
	}
	if strings.TrimSpace(request.JumpURL) == `` {
		gsgin.GinResponseError(c, `jump_url不能为空`, nil)
		return
	}
	if strings.TrimSpace(request.CssSelector) == `` {
		gsgin.GinResponseError(c, `css_selector不能为空`, nil)
		return
	}
	workflowInfo, requirementFragment, devPlanFragment, ok := taskWorkflowLoadGenerationContextByWorkflowID(c, request.WorkflowID)
	if !ok {
		return
	}
	// 从新表 smart_link 查询 label
	label, labelErr := querySmartLinkLabel(request.SmartLinkID)
	if labelErr != nil {
		gsgin.GinResponseError(c, labelErr.Error(), nil)
		return
	}
	resultFile, err := dispatchScrapeTaskAndAwait(request.SmartLinkID, label, request.JumpURL, request.CssSelector, request.WaitSeconds)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	resultFile.DownloadURL = buildAbsoluteDownloadURL(c, resultFile.DownloadURL)
	requirementContent := cast.ToString(requirementFragment[`content`])
	devPlanContent := cast.ToString(devPlanFragment[`content`])
	processedMarkdown := ``
	imageCount := 0
	apiCandidates := []string{}
	zipResult, zipErr := processScrapeZipResult(resultFile.DownloadURL, ``)
	if zipErr == nil {
		processedMarkdown = cast.ToString(zipResult[`markdown`])
		imageCount = cast.ToInt(zipResult[`image_count`])
		apiCandidates = taskWorkflowExtractAPIPaths(processedMarkdown)
	}
	structuredSummary := taskWorkflowBuildUIAssistStructuredSummary(processedMarkdown, apiCandidates)
	uiAssistReport := map[string]any{
		`smart_link_id`:      request.SmartLinkID,
		`label`:              label,
		`jump_url`:           request.JumpURL,
		`css_selector`:       request.CssSelector,
		`wait_seconds`:       request.WaitSeconds,
		`download_url`:       resultFile.DownloadURL,
		`file_name`:          resultFile.FileName,
		`markdown`:           processedMarkdown,
		`image_count`:        imageCount,
		`api_candidates`:     apiCandidates,
		`structured_summary`: structuredSummary,
		`prompt_text`:        `下载 ` + resultFile.DownloadURL + ` ，分析页面抓取结果，整理真实请求接口、关键参数样例和页面操作链路`,
	}
	runInfo, err := common.DbMain.TaskWorkflowCreateRun(
		cast.ToInt(workflowInfo[`id`]),
		taskWorkflowRunTypeUIAssistGenerate,
		`manual`,
		requirementContent,
		devPlanContent,
		map[string]any{},
		map[string]any{},
		uiAssistReport,
		taskWorkflowBuildUIAssistSummaryMarkdown(uiAssistReport),
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	updatedWorkflowInfo, err := common.DbMain.TaskWorkflowInfo(cast.ToInt(workflowInfo[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`:  updatedWorkflowInfo,
		`ui_assist`: uiAssistReport,
		`test_run`:  taskWorkflowNormalizeRunInfo(runInfo),
	})
}

// TaskWorkflowUIAssistInfo 查询最近一次页面辅助识别结果。
func TaskWorkflowUIAssistInfo(c *gin.Context) {
	workflowInfo, ok := taskWorkflowLoadInfoOrResponse(c)
	if !ok {
		return
	}
	runInfo, err := common.DbMain.TaskWorkflowLatestRunByType(cast.ToInt(workflowInfo[`id`]), taskWorkflowRunTypeUIAssistGenerate)
	if err == nil && len(runInfo) > 0 {
		normalized := taskWorkflowNormalizeRunInfo(runInfo)
		gsgin.GinResponseSuccess(c, ``, map[string]any{
			`workflow`:  workflowInfo,
			`ui_assist`: normalized[`test_report`],
			`test_run`:  normalized,
		})
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`:  workflowInfo,
		`ui_assist`: map[string]any{},
		`test_run`:  map[string]any{},
	})
}

// TaskWorkflowCoverageInfo 查询覆盖分析结果。
func TaskWorkflowCoverageInfo(c *gin.Context) {
	workflowInfo, ok := taskWorkflowLoadInfoOrResponse(c)
	if !ok {
		return
	}
	runInfo, err := common.DbMain.TaskWorkflowLatestRunByType(cast.ToInt(workflowInfo[`id`]), taskWorkflowRunTypeCoverageGenerate)
	if err == nil && len(runInfo) > 0 {
		gsgin.GinResponseSuccess(c, ``, map[string]any{
			`workflow`:        workflowInfo,
			`coverage_report`: taskWorkflowDecodeJSONMap(runInfo[`coverage_report_json`]),
			`test_run`:        taskWorkflowNormalizeRunInfo(runInfo),
		})
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`:        workflowInfo,
		`coverage_report`: taskWorkflowDefaultCoverageReport(),
		`test_run`:        map[string]any{},
	})
}

// TaskWorkflowTestPlanGenerate 生成测试计划并落历史记录。
func TaskWorkflowTestPlanGenerate(c *gin.Context) {
	workflowInfo, requirementFragment, devPlanFragment, ok := taskWorkflowLoadGenerationContext(c)
	if !ok {
		return
	}
	requirementContent := cast.ToString(requirementFragment[`content`])
	devPlanContent := cast.ToString(devPlanFragment[`content`])
	uiAssistInfo := taskWorkflowLatestUIAssistReport(cast.ToInt(workflowInfo[`id`]))
	coverageReport, testPlan := taskWorkflowBuildCoverageAndPlan(cast.ToInt(workflowInfo[`id`]), requirementContent, devPlanContent, uiAssistInfo)
	runInfo, err := common.DbMain.TaskWorkflowCreateRun(
		cast.ToInt(workflowInfo[`id`]),
		taskWorkflowRunTypeTestPlanGenerate,
		`manual`,
		requirementContent,
		devPlanContent,
		coverageReport,
		testPlan,
		map[string]any{},
		taskWorkflowBuildSummaryMarkdown(coverageReport, testPlan),
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	updatedWorkflowInfo, err := common.DbMain.TaskWorkflowInfo(cast.ToInt(workflowInfo[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`:  updatedWorkflowInfo,
		`test_plan`: testPlan,
		`test_run`:  taskWorkflowNormalizeRunInfo(runInfo),
	})
}

// TaskWorkflowTestPlanInfo 查询测试计划结果。
func TaskWorkflowTestPlanInfo(c *gin.Context) {
	workflowInfo, ok := taskWorkflowLoadInfoOrResponse(c)
	if !ok {
		return
	}
	runInfo, err := common.DbMain.TaskWorkflowLatestRunByType(cast.ToInt(workflowInfo[`id`]), taskWorkflowRunTypeTestPlanGenerate)
	if err == nil && len(runInfo) > 0 {
		gsgin.GinResponseSuccess(c, ``, map[string]any{
			`workflow`:  workflowInfo,
			`test_plan`: taskWorkflowDecodeJSONMap(runInfo[`test_plan_json`]),
			`test_run`:  taskWorkflowNormalizeRunInfo(runInfo),
		})
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`:  workflowInfo,
		`test_plan`: taskWorkflowDefaultTestPlan(cast.ToInt(workflowInfo[`id`])),
		`test_run`:  map[string]any{},
	})
}

// TaskWorkflowTestRunList 查询测试执行历史列表。
func TaskWorkflowTestRunList(c *gin.Context) {
	workflowInfo, ok := taskWorkflowLoadInfoOrResponse(c)
	if !ok {
		return
	}
	runList, err := common.DbMain.TaskWorkflowRunList(cast.ToInt(workflowInfo[`id`]))
	if err != nil {
		runList = []map[string]any{}
	}
	normalizedList := make([]map[string]any, 0, len(runList))
	for _, item := range runList {
		normalizedList = append(normalizedList, taskWorkflowNormalizeRunInfo(item))
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`: workflowInfo,
		`list`:     normalizedList,
	})
}

// TaskWorkflowTestRunExecute 执行测试计划并落测试报告。
func TaskWorkflowTestRunExecute(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.TaskWorkflowExecuteRequest{}
	_ = gsgin.GinPostBody(c, &request)
	workflowInfo, requirementFragment, devPlanFragment, ok := taskWorkflowLoadGenerationContextByWorkflowID(c, request.WorkflowID)
	if !ok {
		return
	}
	requirementContent := cast.ToString(requirementFragment[`content`])
	devPlanContent := cast.ToString(devPlanFragment[`content`])
	uiAssistInfo := taskWorkflowLatestUIAssistReport(cast.ToInt(workflowInfo[`id`]))
	coverageReport, testPlan := taskWorkflowBuildCoverageAndPlan(cast.ToInt(workflowInfo[`id`]), requirementContent, devPlanContent, uiAssistInfo)
	if !request.RegeneratePlan {
		latestPlanRun, err := common.DbMain.TaskWorkflowLatestRunByType(cast.ToInt(workflowInfo[`id`]), taskWorkflowRunTypeTestPlanGenerate)
		if err == nil && len(latestPlanRun) > 0 {
			if decodedCoverage := taskWorkflowDecodeJSONMap(latestPlanRun[`coverage_report_json`]); len(decodedCoverage) > 0 {
				coverageReport = decodedCoverage
			}
			if decodedPlan := taskWorkflowDecodeJSONMap(latestPlanRun[`test_plan_json`]); len(decodedPlan) > 0 {
				testPlan = decodedPlan
			}
		}
	}
	testReport := taskWorkflowExecuteTestPlan(testPlan)
	runInfo, err := common.DbMain.TaskWorkflowCreateRun(
		cast.ToInt(workflowInfo[`id`]),
		taskWorkflowRunTypeAPIExecute,
		`manual`,
		requirementContent,
		devPlanContent,
		coverageReport,
		testPlan,
		testReport,
		taskWorkflowBuildExecutionSummaryMarkdown(testReport),
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	updatedWorkflowInfo, err := common.DbMain.TaskWorkflowInfo(cast.ToInt(workflowInfo[`id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`:    updatedWorkflowInfo,
		`test_plan`:   testPlan,
		`test_report`: testReport,
		`test_run`:    taskWorkflowNormalizeRunInfo(runInfo),
	})
}

func taskWorkflowLoadContextForDevPlan(c *gin.Context) (map[string]any, common.MemoryFragmentStore, map[string]any, bool) {
	request := _struct.TaskWorkflowInfoRequest{}
	_ = gsgin.GinPostBody(c, &request)
	return taskWorkflowLoadContextForDevPlanByID(c, request.WorkflowID)
}

func taskWorkflowLoadInfoOrResponse(c *gin.Context) (map[string]any, bool) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return nil, false
	}
	request := _struct.TaskWorkflowInfoRequest{}
	_ = gsgin.GinPostBody(c, &request)
	workflowInfo, err := common.DbMain.TaskWorkflowInfo(request.WorkflowID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return nil, false
	}
	return workflowInfo, true
}

func taskWorkflowLoadContextForDevPlanByID(c *gin.Context, workflowID int) (map[string]any, common.MemoryFragmentStore, map[string]any, bool) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return nil, nil, nil, false
	}
	workflowInfo, err := common.DbMain.TaskWorkflowInfo(workflowID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return nil, nil, nil, false
	}
	memoryDB, ok := taskWorkflowMemoryDBOrResponse(c)
	if !ok {
		return nil, nil, nil, false
	}
	if err = taskWorkflowNormalizeFragmentRefs(workflowInfo, memoryDB); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return nil, nil, nil, false
	}
	homeTaskInfo, err := common.DbMain.HomeTaskRow(cast.ToInt(workflowInfo[`home_task_id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return nil, nil, nil, false
	}
	return workflowInfo, memoryDB, homeTaskInfo, true
}

func taskWorkflowLoadGenerationContext(c *gin.Context) (map[string]any, map[string]any, map[string]any, bool) {
	request := _struct.TaskWorkflowGenerateRequest{}
	_ = gsgin.GinPostBody(c, &request)
	return taskWorkflowLoadGenerationContextByWorkflowID(c, request.WorkflowID)
}

func taskWorkflowLoadGenerationContextByWorkflowID(c *gin.Context, workflowID int) (map[string]any, map[string]any, map[string]any, bool) {
	workflowInfo, memoryDB, homeTaskInfo, ok := taskWorkflowLoadContextForDevPlanByID(c, workflowID)
	if !ok {
		return nil, nil, nil, false
	}
	requirementRef := common.TaskWorkflowParseFragmentRef(cast.ToString(workflowInfo[`requirement_fragment_id`]), taskWorkflowWorkflowFragmentFolderName(workflowInfo))
	if requirementRef.FileID == `` {
		gsgin.GinResponseError(c, `需求文档未绑定`, nil)
		return nil, nil, nil, false
	}
	requirementFragment, err := memoryDB.MemoryFragmentInfo(requirementRef.FileID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return nil, nil, nil, false
	}
	devPlanFragment, err := ensureTaskWorkflowDevPlanFragment(workflowInfo, homeTaskInfo, memoryDB)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return nil, nil, nil, false
	}
	return workflowInfo, requirementFragment, devPlanFragment, true
}

func ensureTaskWorkflowDevPlanFragment(workflowInfo map[string]any, homeTaskInfo map[string]any, memoryDB common.MemoryFragmentStore) (map[string]any, error) {
	workflowID := cast.ToInt(workflowInfo[`id`])
	existingFileID := taskWorkflowGetDocFragmentFileID(workflowID, common.TaskWorkflowDocTypeDevPlan)
	if existingFileID != `` {
		return memoryDB.MemoryFragmentInfo(existingFileID)
	}
	fragmentTitle := strings.TrimSpace(cast.ToString(homeTaskInfo[`name`])) + taskWorkflowDevPlanDefaultTitleSuffix
	if strings.TrimSpace(fragmentTitle) == taskWorkflowDevPlanDefaultTitleSuffix {
		fragmentTitle = `开发执行文档`
	}
	fragmentInfo, err := memoryDB.MemoryFragmentSave(0, fragmentTitle, taskWorkflowDevPlanDefaultTemplate, []string{taskWorkflowDevPlanDefaultTag}, taskWorkflowWorkflowFragmentFolderName(workflowInfo))
	if err != nil {
		return nil, err
	}
	component.MemoryRuntime.ScheduleSync()
	fragmentFileID := strings.TrimSpace(cast.ToString(fragmentInfo[`file_id`]))
	if fragmentFileID == `` {
		return nil, gstool.Error(`开发执行片段创建失败`)
	}
	fragmentFullRef := common.TaskWorkflowBuildFragmentRef(cast.ToString(fragmentInfo[`folder_name`]), fragmentFileID)
	if err = common.DbMain.TaskWorkflowBindDevPlanFragment(workflowID, fragmentFullRef); err != nil {
		return nil, err
	}
	return fragmentInfo, nil
}

func taskWorkflowMemoryDBOrResponse(c *gin.Context) (common.MemoryFragmentStore, bool) {
	if component.MemoryRuntime == nil {
		gsgin.GinResponseError(c, common.ErrMemoryNotConfigured.Error(), map[string]any{
			`configured`: false,
		})
		return nil, false
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		gsgin.GinResponseError(c, err.Error(), map[string]any{
			`configured`: false,
		})
		return nil, false
	}
	return component.MemoryRuntime.DB(), true
}

func buildTaskWorkflowResponse(c *gin.Context, workflowInfo map[string]any) (map[string]any, error) {
	memoryDB := taskWorkflowMemoryDBIfConfigured()
	if err := taskWorkflowNormalizeFragmentRefs(workflowInfo, memoryDB); err != nil {
		gstool.FmtPrintlnLogTime("[buildTaskWorkflowResponse] taskWorkflowNormalizeFragmentRefs 失败 memoryDB=%v err=%v，步骤文档片段将被跳过", memoryDB != nil, err)
		return nil, err
	}
	homeTaskInfo, err := common.DbMain.HomeTaskRow(cast.ToInt(workflowInfo[`home_task_id`]))
	if err != nil {
		gstool.FmtPrintlnLogTime("[buildTaskWorkflowResponse] HomeTaskRow 失败 homeTaskID=%d err=%v", cast.ToInt(workflowInfo[`home_task_id`]), err)
		return nil, err
	}
	// 获取关联的模板和步骤信息（尽早获取，便于预生成步骤文档片段）。
	homeTaskID := cast.ToInt(workflowInfo[`home_task_id`])
	template, templateSteps, templateErr := common.DbMain.HomeTaskWorkflowTemplateSteps(homeTaskID)
	if templateErr != nil {
		gstool.FmtPrintlnLogTime("[buildTaskWorkflowResponse] HomeTaskWorkflowTemplateSteps 失败 homeTaskID=%d err=%v", homeTaskID, templateErr)
	}

	workflowID := cast.ToInt(workflowInfo[`id`])
	// 预生成模板步骤中配置的知识片段文档，确保所有占位符在提示词替换前可用。
	// 内置文档（plain_text_requirement/design_plan_requirement/api_doc）已废弃，
	// 统一由步骤文档（step_document）通过 ensureTaskWorkflowStepFragments 创建。
	if workflowID > 0 && len(templateSteps) > 0 {
		gstool.FmtPrintlnLogTime("[buildTaskWorkflowResponse] 进入 ensureTaskWorkflowStepFragments workflowID=%d templateSteps=%d", workflowID, len(templateSteps))
		ensureTaskWorkflowStepFragments(c, workflowInfo, homeTaskInfo, templateSteps)
	} else {
		gstool.FmtPrintlnLogTime("[buildTaskWorkflowResponse] 跳过 ensureTaskWorkflowStepFragments workflowID=%d templateSteps=%d", workflowID, len(templateSteps))
	}
	// 知识片段创建完成后，逐个初始化缺失的提示词。
	// 优先从模板步骤的 prompt_content 获取，回退到旧全局配置。
	if workflowID > 0 {
		// 有模板步骤时，使用模板步骤的 prompt_content 初始化 step_prompts JSON
		if len(templateSteps) > 0 {
			placeholders := buildTaskWorkflowPlaceholderMap(c, homeTaskInfo, workflowInfo)
			existingPrompts, _ := common.DbMain.WorkflowStepPromptsRead(workflowID)
			needUpdate := false
			for _, step := range templateSteps {
				stepKey := cast.ToString(step[`step_key`])
				promptContent := strings.TrimSpace(cast.ToString(step[`prompt_content`]))
				if promptContent == `` {
					continue
				}
				// 只填充当前为空的，保留用户已有修改
				if strings.TrimSpace(existingPrompts[stepKey]) != `` {
					continue
				}
				// 注入步骤级占位符：{步骤ID} 替换为当前步骤的 step_key
				placeholders[`{步骤ID}`] = stepKey
				existingPrompts[stepKey] = taskWorkflowResolvePlaceholders(promptContent, placeholders)
				needUpdate = true
			}
			if needUpdate {
				jsonBytes, _ := json.Marshal(existingPrompts)
				now := time.Now().Unix()
				common.DbMain.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{`id`: workflowID}, map[string]any{
					`step_prompts`: string(jsonBytes),
					`update_time`:  now,
				}).Exec()
				// 从数据库刷新，确保返回给前端的数据包含已初始化的提示词
				updatedInfo, updateErr := common.DbMain.TaskWorkflowInfo(workflowID)
				if updateErr == nil {
					workflowInfo = updatedInfo
				}
			}
		}
	}

	// 从新文档表读取文档列表
	workflowDocuments, _ := common.DbMain.TaskWorkflowDocumentList(workflowID)
	hasWorkflowDocuments := common.DbMain.TaskWorkflowDocumentHasRecords(workflowID)

	return map[string]any{
		`workflow`:                 workflowInfo,
		`home_task`:                homeTaskInfo,
		`template`:                 template,
		`template_steps`:           templateSteps,
		`requirement_fetch_config`: taskWorkflowRequirementFetchConfig(),
		`documents`:                workflowDocuments,
		`has_workflow_documents`:   hasWorkflowDocuments,
	}, nil
}

func taskWorkflowWorkflowFragmentFolderName(workflowInfo map[string]any) string {
	return common.TaskWorkflowFragmentFolderName(workflowInfo)
}

// taskWorkflowGetDocFragmentFileID 从文档表查询指定类型的文档片段 file_id。
func taskWorkflowGetDocFragmentFileID(workflowID int, docType string) string {
	if workflowID <= 0 || docType == `` {
		return ``
	}
	docs, err := common.DbMain.TaskWorkflowDocumentList(workflowID)
	if err != nil {
		return ``
	}
	for _, doc := range docs {
		if cast.ToString(doc[`document_type`]) == docType {
			return strings.TrimSpace(cast.ToString(doc[`file_id`]))
		}
	}
	return ``
}

// taskWorkflowGetStepDocFileID 从文档表按 document_id + template_step_id 精确查找步骤文档的 file_id。
// 避免 taskWorkflowGetDocFragmentFileID 按 docType 模糊匹配时错误返回其他步骤的文档片段。
func taskWorkflowGetStepDocFileID(workflowID int, documentID string, templateStepID int) string {
	if workflowID <= 0 || documentID == `` || templateStepID <= 0 {
		return ``
	}
	docs, err := common.DbMain.TaskWorkflowDocumentList(workflowID)
	if err != nil {
		return ``
	}
	for _, doc := range docs {
		if cast.ToString(doc[`document_id`]) == documentID && cast.ToInt(doc[`template_step_id`]) == templateStepID {
			return strings.TrimSpace(cast.ToString(doc[`file_id`]))
		}
	}
	return ``
}

// taskWorkflowGetFragmentFileIDByName 按文档名称从文档表中查找任意类型文档的 file_id。
// 用于替代按 document_type 查找的内置文档逻辑，统一按名称匹配（内置文档和步骤文档均可匹配）。
func taskWorkflowGetFragmentFileIDByName(workflowID int, docName string) string {
	if workflowID <= 0 || docName == `` {
		return ``
	}
	docs, err := common.DbMain.TaskWorkflowDocumentList(workflowID)
	if err != nil {
		return ``
	}
	for _, doc := range docs {
		if cast.ToString(doc[`document_name`]) == docName {
			return strings.TrimSpace(cast.ToString(doc[`file_id`]))
		}
	}
	return ``
}

func taskWorkflowMemoryDBIfConfigured() common.MemoryFragmentStore {
	if component.MemoryRuntime == nil {
		return nil
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return nil
	}
	return component.MemoryRuntime.DB()
}

func taskWorkflowNormalizeFragmentRefs(workflowInfo map[string]any, memoryDB common.MemoryFragmentStore) error {
	if len(workflowInfo) == 0 {
		return nil
	}
	workflowID := cast.ToInt(workflowInfo[`id`])
	folderName := taskWorkflowWorkflowFragmentFolderName(workflowInfo)
	if strings.TrimSpace(cast.ToString(workflowInfo[`fragment_folder_name`])) != folderName {
		_ = common.DbMain.TaskWorkflowUpdateFragmentFolderName(workflowID, folderName)
		workflowInfo[`fragment_folder_name`] = folderName
	}
	// 规范化 requirement_fragment_id 的文件夹路径
	if raw := cast.ToString(workflowInfo[`requirement_fragment_id`]); raw != `` {
		ref := common.TaskWorkflowParseFragmentRef(raw, folderName)
		if ref.FileID != `` {
			nextRef := ref.FullRef
			if memoryDB != nil {
				if info, err := memoryDB.MemoryFragmentInfo(ref.FileID); err == nil {
					nextRef = common.TaskWorkflowBuildFragmentRef(cast.ToString(info[`folder_name`]), ref.FileID)
				}
			}
			if nextRef != `` && nextRef != raw {
				workflowInfo[`requirement_fragment_id`] = nextRef
				_, _ = common.DbMain.Client.QuickUpdate(`tbl_task_workflow`, map[string]any{`id`: workflowID}, map[string]any{
					`requirement_fragment_id`: nextRef,
					`update_time`:             time.Now().Unix(),
				}).Exec()
			}
		}
	}
	return nil
}

func taskWorkflowDefaultCoverageReport() map[string]any {
	return map[string]any{
		`summary`: map[string]any{
			`requirement_points`: 0,
			`covered`:            0,
			`partial`:            0,
			`missing`:            0,
			`questions`:          0,
			`blocked`:            0,
		},
		`items`:     []map[string]any{},
		`questions`: []string{},
		`blocked`:   []string{},
	}
}

func taskWorkflowDefaultTestPlan(workflowID int) map[string]any {
	return map[string]any{
		`plan_name`:      ``,
		`workflow_id`:    workflowID,
		`coverage_links`: []map[string]any{},
		`preconditions`:  []map[string]any{},
		`api_cases`:      []map[string]any{},
		`open_questions`: []string{},
		`blocked_items`:  []string{},
	}
}

func taskWorkflowDecodeJSONMap(raw any) map[string]any {
	text := strings.TrimSpace(cast.ToString(raw))
	if text == `` {
		return map[string]any{}
	}
	result := map[string]any{}
	if err := gstool.JsonDecode(text, &result); err != nil {
		return map[string]any{}
	}
	return result
}

func taskWorkflowNormalizeRunInfo(runInfo map[string]any) map[string]any {
	if len(runInfo) == 0 {
		return map[string]any{}
	}
	result := map[string]any{}
	for key, value := range runInfo {
		result[key] = value
	}
	if strings.TrimSpace(cast.ToString(result[`coverage_report_json`])) != `` {
		result[`coverage_report`] = taskWorkflowDecodeJSONMap(result[`coverage_report_json`])
	}
	if strings.TrimSpace(cast.ToString(result[`test_plan_json`])) != `` {
		result[`test_plan`] = taskWorkflowDecodeJSONMap(result[`test_plan_json`])
	}
	if strings.TrimSpace(cast.ToString(result[`test_report_json`])) != `` {
		result[`test_report`] = taskWorkflowDecodeJSONMap(result[`test_report_json`])
	}
	return result
}

func taskWorkflowBuildCoverageAndPlan(workflowID int, requirementContent, devPlanContent string, uiAssistInfo map[string]any) (map[string]any, map[string]any) {
	requirementPoints := taskWorkflowExtractRequirementPoints(requirementContent)
	apiPathList := taskWorkflowMergeAPIPaths(
		taskWorkflowExtractAPIPaths(devPlanContent),
		cast.ToStringSlice(uiAssistInfo[`api_candidates`]),
	)
	uiAssistSummary := cast.ToStringMap(uiAssistInfo[`structured_summary`])
	uiAssistAPIHintMap := taskWorkflowBuildUIAssistAPIHintMap(uiAssistSummary)
	apiInfoList := taskWorkflowQueryAPIInfo(apiPathList)
	apiInfoByPath := map[string]map[string]any{}
	for _, item := range apiInfoList {
		apiInfoByPath[cast.ToString(item[`url`])] = item
		for _, apiPath := range apiPathList {
			if strings.HasSuffix(cast.ToString(item[`url`]), apiPath) {
				apiInfoByPath[apiPath] = item
			}
		}
	}
	coverageItems := make([]map[string]any, 0, len(requirementPoints))
	coverageLinks := make([]map[string]any, 0, len(requirementPoints))
	for index, point := range requirementPoints {
		status := `missing`
		matchedAPI := map[string]any{}
		if index < len(apiPathList) {
			apiPath := apiPathList[index]
			if info, ok := apiInfoByPath[apiPath]; ok {
				status = `covered`
				matchedAPI = info
			} else {
				status = `partial`
			}
		}
		coverageItems = append(coverageItems, map[string]any{
			`requirement_key`:   `req-` + cast.ToString(index+1),
			`requirement_title`: point,
			`status`:            status,
			`api_paths`:         taskWorkflowSliceOrEmpty(apiPathList, index),
			`matched_api_ids`:   taskWorkflowMatchedAPIIDs(matchedAPI),
		})
		coverageLinks = append(coverageLinks, map[string]any{
			`requirement_key`:   `req-` + cast.ToString(index+1),
			`requirement_title`: point,
			`status`:            status,
		})
	}
	summary := map[string]any{
		`requirement_points`: len(requirementPoints),
		`covered`:            taskWorkflowCountCoverageStatus(coverageItems, `covered`),
		`partial`:            taskWorkflowCountCoverageStatus(coverageItems, `partial`),
		`missing`:            taskWorkflowCountCoverageStatus(coverageItems, `missing`),
		`questions`:          0,
		`blocked`:            0,
	}
	coverageReport := map[string]any{
		`summary`:   summary,
		`items`:     coverageItems,
		`questions`: []string{},
		`blocked`:   []string{},
		`ui_assist`: map[string]any{
			`used`:           len(uiAssistInfo) > 0,
			`api_candidates`: cast.ToStringSlice(uiAssistInfo[`api_candidates`]),
		},
	}
	apiCases := make([]map[string]any, 0, len(apiPathList))
	for index, apiPath := range apiPathList {
		apiInfo := apiInfoByPath[apiPath]
		apiCases = append(apiCases, map[string]any{
			`case_id`:       `wf-` + cast.ToString(workflowID) + `-api-` + cast.ToString(index+1),
			`name`:          taskWorkflowCaseName(apiInfo, apiPath, index),
			`api_uri`:       apiPath,
			`method`:        taskWorkflowCaseMethod(apiInfo),
			`priority`:      `P1`,
			`request_hints`: uiAssistAPIHintMap[apiPath],
			`assertions`: []map[string]any{
				{
					`type`:     `status_code`,
					`expected`: 200,
				},
			},
		})
	}
	testPlan := taskWorkflowDefaultTestPlan(workflowID)
	testPlan[`plan_name`] = `任务工作流-` + cast.ToString(workflowID)
	testPlan[`coverage_links`] = coverageLinks
	testPlan[`preconditions`] = []map[string]any{}
	testPlan[`api_cases`] = apiCases
	testPlan[`ui_assist_used`] = len(uiAssistInfo) > 0
	testPlan[`ui_assist_candidates`] = cast.ToStringSlice(uiAssistInfo[`api_candidates`])
	testPlan[`ui_assist_summary`] = uiAssistSummary
	return coverageReport, testPlan
}

func taskWorkflowExtractRequirementPoints(content string) []string {
	lines := strings.Split(content, "\n")
	result := make([]string, 0)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, `## `) {
			continue
		}
		title := strings.TrimSpace(strings.TrimPrefix(trimmed, `## `))
		if title == `` {
			continue
		}
		result = append(result, title)
	}
	return result
}

func taskWorkflowExtractAPIPaths(content string) []string {
	matchList := taskWorkflowAPIPathReg.FindAllString(content, -1)
	if len(matchList) == 0 {
		return []string{}
	}
	uniqueMap := map[string]struct{}{}
	result := make([]string, 0, len(matchList))
	for _, item := range matchList {
		item = strings.TrimSpace(item)
		if item == `` {
			continue
		}
		if _, exists := uniqueMap[item]; exists {
			continue
		}
		uniqueMap[item] = struct{}{}
		result = append(result, item)
	}
	sort.Strings(result)
	return result
}

func taskWorkflowMergeAPIPaths(pathGroups ...[]string) []string {
	uniqueMap := map[string]struct{}{}
	result := make([]string, 0)
	for _, group := range pathGroups {
		for _, item := range group {
			item = strings.TrimSpace(item)
			if item == `` {
				continue
			}
			if _, exists := uniqueMap[item]; exists {
				continue
			}
			uniqueMap[item] = struct{}{}
			result = append(result, item)
		}
	}
	sort.Strings(result)
	return result
}

func taskWorkflowLatestUIAssistReport(workflowID int) map[string]any {
	if common.DbMain == nil || common.DbMain.Client == nil || workflowID <= 0 {
		return map[string]any{}
	}
	runInfo, err := common.DbMain.TaskWorkflowLatestRunByType(workflowID, taskWorkflowRunTypeUIAssistGenerate)
	if err != nil || len(runInfo) == 0 {
		return map[string]any{}
	}
	normalized := taskWorkflowNormalizeRunInfo(runInfo)
	return cast.ToStringMap(normalized[`test_report`])
}

func taskWorkflowBuildUIAssistStructuredSummary(markdown string, apiCandidates []string) map[string]any {
	stepTitles := taskWorkflowExtractUIStepTitles(markdown)
	queryHints := taskWorkflowExtractURLQueryHints(markdown)
	jsonSamples := taskWorkflowExtractJSONBlockSamples(markdown)
	apiSamples := taskWorkflowBuildUIAssistAPISamples(apiCandidates, queryHints, jsonSamples)
	parameterHints := taskWorkflowCollectUIParameterHints(queryHints, jsonSamples)
	return map[string]any{
		`step_titles`:     stepTitles,
		`parameter_hints`: parameterHints,
		`api_samples`:     apiSamples,
		`json_samples`:    jsonSamples,
	}
}

func taskWorkflowExtractUIStepTitles(markdown string) []string {
	lines := strings.Split(markdown, "\n")
	result := make([]string, 0)
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == `` {
			continue
		}
		if strings.HasPrefix(trimmed, `#`) {
			result = append(result, strings.TrimSpace(strings.TrimLeft(trimmed, `# `)))
			continue
		}
		if taskWorkflowLooksLikeOrderedStep(trimmed) {
			result = append(result, trimmed)
		}
		if len(result) >= 8 {
			break
		}
	}
	return result
}

func taskWorkflowLooksLikeOrderedStep(line string) bool {
	runes := []rune(line)
	if len(runes) < 3 {
		return false
	}
	for i := 0; i < len(runes); i++ {
		if runes[i] < '0' || runes[i] > '9' {
			if i > 0 && i+1 < len(runes) && (runes[i] == '.' || runes[i] == '、') && runes[i+1] == ' ' {
				return true
			}
			return false
		}
	}
	return false
}

func taskWorkflowExtractURLQueryHints(markdown string) []map[string]any {
	matchList := taskWorkflowURLWithQueryReg.FindAllString(markdown, -1)
	result := make([]map[string]any, 0)
	for _, item := range matchList {
		parts := strings.SplitN(item, `?`, 2)
		if len(parts) != 2 {
			continue
		}
		queryKeys := make([]string, 0)
		for _, pair := range strings.Split(parts[1], `&`) {
			key := strings.TrimSpace(strings.SplitN(pair, `=`, 2)[0])
			if key == `` {
				continue
			}
			queryKeys = append(queryKeys, key)
		}
		result = append(result, map[string]any{
			`api_uri`:    parts[0],
			`query_keys`: taskWorkflowUniqueStrings(queryKeys),
			`sample_url`: item,
		})
	}
	return result
}

func taskWorkflowExtractJSONBlockSamples(markdown string) []map[string]any {
	matchList := taskWorkflowJSONFenceReg.FindAllStringSubmatch(markdown, -1)
	result := make([]map[string]any, 0)
	for _, item := range matchList {
		if len(item) < 2 {
			continue
		}
		content := strings.TrimSpace(item[1])
		if content == `` {
			continue
		}
		jsonKeys := taskWorkflowExtractTopLevelJSONKeys(content)
		if len(jsonKeys) == 0 {
			continue
		}
		result = append(result, map[string]any{
			`keys`:        jsonKeys,
			`sample_text`: taskWorkflowTruncateText(content, 240),
		})
	}
	return result
}

func taskWorkflowExtractTopLevelJSONKeys(content string) []string {
	var raw any
	if err := json.Unmarshal([]byte(content), &raw); err != nil {
		return []string{}
	}
	switch val := raw.(type) {
	case map[string]any:
		return taskWorkflowSortedMapKeys(val)
	case []any:
		if len(val) == 0 {
			return []string{}
		}
		first, ok := val[0].(map[string]any)
		if !ok {
			return []string{}
		}
		return taskWorkflowSortedMapKeys(first)
	default:
		return []string{}
	}
}

func taskWorkflowSortedMapKeys(data map[string]any) []string {
	result := make([]string, 0, len(data))
	for key := range data {
		key = strings.TrimSpace(key)
		if key == `` {
			continue
		}
		result = append(result, key)
	}
	sort.Strings(result)
	return result
}

func taskWorkflowBuildUIAssistAPISamples(apiCandidates []string, queryHints []map[string]any, jsonSamples []map[string]any) []map[string]any {
	result := make([]map[string]any, 0, len(apiCandidates))
	defaultBodyKeys := []string{}
	if len(jsonSamples) > 0 {
		defaultBodyKeys = cast.ToStringSlice(jsonSamples[0][`keys`])
	}
	for _, apiURI := range apiCandidates {
		sample := map[string]any{
			`api_uri`:     apiURI,
			`query_keys`:  []string{},
			`body_keys`:   defaultBodyKeys,
			`sample_text`: ``,
		}
		for _, hint := range queryHints {
			hintURI := strings.TrimSpace(cast.ToString(hint[`api_uri`]))
			if hintURI == apiURI || strings.HasSuffix(hintURI, apiURI) {
				sample[`query_keys`] = cast.ToStringSlice(hint[`query_keys`])
				sample[`sample_text`] = cast.ToString(hint[`sample_url`])
				break
			}
		}
		result = append(result, sample)
	}
	return result
}

func taskWorkflowCollectUIParameterHints(queryHints []map[string]any, jsonSamples []map[string]any) []string {
	allKeys := make([]string, 0)
	for _, item := range queryHints {
		allKeys = append(allKeys, cast.ToStringSlice(item[`query_keys`])...)
	}
	for _, item := range jsonSamples {
		allKeys = append(allKeys, cast.ToStringSlice(item[`keys`])...)
	}
	return taskWorkflowUniqueStrings(allKeys)
}

func taskWorkflowUniqueStrings(values []string) []string {
	uniqueMap := map[string]struct{}{}
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == `` {
			continue
		}
		if _, exists := uniqueMap[value]; exists {
			continue
		}
		uniqueMap[value] = struct{}{}
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func taskWorkflowTruncateText(text string, maxLen int) string {
	if maxLen <= 0 || len(text) <= maxLen {
		return text
	}
	return text[:maxLen] + `...`
}

func taskWorkflowBuildUIAssistAPIHintMap(summary map[string]any) map[string]map[string]any {
	result := map[string]map[string]any{}
	apiSamples := cast.ToSlice(summary[`api_samples`])
	for _, item := range apiSamples {
		itemMap, ok := item.(map[string]any)
		if !ok {
			continue
		}
		apiURI := strings.TrimSpace(cast.ToString(itemMap[`api_uri`]))
		if apiURI == `` {
			continue
		}
		result[apiURI] = map[string]any{
			`query_keys`:  cast.ToStringSlice(itemMap[`query_keys`]),
			`body_keys`:   cast.ToStringSlice(itemMap[`body_keys`]),
			`sample_text`: cast.ToString(itemMap[`sample_text`]),
		}
	}
	return result
}

func taskWorkflowQueryAPIInfo(apiPathList []string) []map[string]any {
	if common.DbMain == nil || common.DbMain.Client == nil || len(apiPathList) == 0 {
		return []map[string]any{}
	}
	placeholderList := make([]string, 0, len(apiPathList))
	args := make([]any, 0, len(apiPathList)*2)
	for _, item := range apiPathList {
		placeholderList = append(placeholderList, `url = ? or url like ?`)
		args = append(args, item)
		args = append(args, `%`+item)
	}
	sql := `select id,name,method,url from tbl_api where ` + strings.Join(placeholderList, ` or `) + ` order by id asc`
	list, err := common.DbMain.Client.QueryBySql(sql, args...).All()
	if err != nil {
		return []map[string]any{}
	}
	return list
}

func taskWorkflowCountCoverageStatus(itemList []map[string]any, status string) int {
	total := 0
	for _, item := range itemList {
		if cast.ToString(item[`status`]) == status {
			total++
		}
	}
	return total
}

func taskWorkflowMatchedAPIIDs(apiInfo map[string]any) []int {
	if len(apiInfo) == 0 {
		return []int{}
	}
	return []int{cast.ToInt(apiInfo[`id`])}
}

func taskWorkflowSliceOrEmpty(apiPathList []string, index int) []string {
	if index >= 0 && index < len(apiPathList) {
		return []string{apiPathList[index]}
	}
	return []string{}
}

func taskWorkflowCaseName(apiInfo map[string]any, apiPath string, index int) string {
	if strings.TrimSpace(cast.ToString(apiInfo[`name`])) != `` {
		return cast.ToString(apiInfo[`name`])
	}
	return `接口用例-` + cast.ToString(index+1) + `-` + apiPath
}

func taskWorkflowCaseMethod(apiInfo map[string]any) string {
	method := strings.ToUpper(strings.TrimSpace(cast.ToString(apiInfo[`method`])))
	if method == `` {
		return `POST`
	}
	return method
}

func taskWorkflowBuildSummaryMarkdown(coverageReport, testPlan map[string]any) string {
	summary := cast.ToStringMap(coverageReport[`summary`])
	apiCaseCount := 0
	if apiCases, ok := testPlan[`api_cases`].([]map[string]any); ok {
		apiCaseCount = len(apiCases)
	}
	return "# 生成摘要\n\n" +
		"- 需求点：" + cast.ToString(summary[`requirement_points`]) + "\n" +
		"- 已覆盖：" + cast.ToString(summary[`covered`]) + "\n" +
		"- 部分覆盖：" + cast.ToString(summary[`partial`]) + "\n" +
		"- 缺失：" + cast.ToString(summary[`missing`]) + "\n" +
		"- 接口用例：" + cast.ToString(apiCaseCount) + "\n"
}

func taskWorkflowExecuteTestPlan(testPlan map[string]any) map[string]any {
	apiCasesRaw := taskWorkflowToMapSlice(testPlan[`api_cases`])
	caseResults := make([]map[string]any, 0, len(apiCasesRaw))
	passed := 0
	failed := 0
	for _, caseInfo := range apiCasesRaw {
		result := taskWorkflowExecuteSingleCase(caseInfo)
		caseResults = append(caseResults, result)
		if cast.ToBool(result[`passed`]) {
			passed++
		} else {
			failed++
		}
	}
	return map[string]any{
		`summary`: map[string]any{
			`total`:  len(caseResults),
			`passed`: passed,
			`failed`: failed,
		},
		`case_results`: caseResults,
	}
}

func taskWorkflowToMapSlice(raw any) []map[string]any {
	if raw == nil {
		return []map[string]any{}
	}
	if list, ok := raw.([]map[string]any); ok {
		return list
	}
	result := make([]map[string]any, 0)
	if list, ok := raw.([]any); ok {
		for _, item := range list {
			if itemMap, ok := item.(map[string]any); ok {
				result = append(result, itemMap)
			}
		}
	}
	return result
}

func taskWorkflowExecuteSingleCase(caseInfo map[string]any) map[string]any {
	apiPath := strings.TrimSpace(cast.ToString(caseInfo[`api_uri`]))
	baseResult := map[string]any{
		`case_id`: caseInfo[`case_id`],
		`name`:    caseInfo[`name`],
		`api_uri`: apiPath,
		`method`:  caseInfo[`method`],
		`passed`:  false,
	}
	if apiPath == `` {
		baseResult[`error`] = `api_uri 不能为空`
		return baseResult
	}
	apiInfo, err := common.DbMain.Client.QuickQuery(`tbl_api`, `*`, map[string]any{
		`url`: apiPath,
	}).One()
	if err != nil || len(apiInfo) == 0 {
		list, queryErr := common.DbMain.Client.QueryBySql(`select * from tbl_api where url like ? order by id asc limit 1`, `%`+apiPath).All()
		if queryErr == nil && len(list) > 0 {
			apiInfo = list[0]
		}
	}
	if len(apiInfo) == 0 {
		baseResult[`error`] = `未找到接口定义`
		return baseResult
	}
	folderInfo, _ := common.DbMain.Client.QuickQuery(`tbl_api_dir`, `headers,env_id`, map[string]any{
		`id`: apiInfo[`folder_id`],
	}).One()
	if len(folderInfo) > 0 {
		apiInfo[`folder_headers`] = folderInfo[`headers`]
		if cast.ToInt(apiInfo[`env_id`]) == 0 {
			apiInfo[`env_id`] = folderInfo[`env_id`]
		}
	}
	apiCli := api.NewApi(apiInfo)
	runErr := apiCli.Run()
	if runErr != nil {
		baseResult[`error`] = runErr.Error()
		return baseResult
	}
	apiCli.ResponseTake()
	_, _ = common.DbMain.Client.QuickUpdate(`tbl_api`, map[string]any{
		`id`: apiInfo[`id`],
	}, map[string]any{
		`last_result`: gstool.JsonEncode(apiCli.Result),
	}).Exec()
	baseResult[`api_id`] = cast.ToInt(apiInfo[`id`])
	baseResult[`status_code`] = apiCli.Result.StatusCode
	baseResult[`status`] = apiCli.Result.Status
	baseResult[`errmsg`] = apiCli.Result.Errmsg
	baseResult[`result`] = apiCli.Result.Result
	baseResult[`response_time_ms`] = apiCli.Result.Millisecond
	baseResult[`passed`] = apiCli.Result.Errmsg == `` && apiCli.Result.StatusCode >= 200 && apiCli.Result.StatusCode < 300
	return baseResult
}

func taskWorkflowBuildExecutionSummaryMarkdown(testReport map[string]any) string {
	summary := cast.ToStringMap(testReport[`summary`])
	return "# 执行摘要\n\n" +
		"- 用例总数：" + cast.ToString(summary[`total`]) + "\n" +
		"- 通过：" + cast.ToString(summary[`passed`]) + "\n" +
		"- 失败：" + cast.ToString(summary[`failed`]) + "\n"
}

func taskWorkflowBuildUIAssistSummaryMarkdown(uiAssistReport map[string]any) string {
	return "# 页面辅助识别摘要\n\n" +
		"- smart_link_id：" + cast.ToString(uiAssistReport[`smart_link_id`]) + "\n" +
		"- label：" + cast.ToString(uiAssistReport[`label`]) + "\n" +
		"- jump_url：" + cast.ToString(uiAssistReport[`jump_url`]) + "\n" +
		"- 下载地址：" + cast.ToString(uiAssistReport[`download_url`]) + "\n"
}

// TaskWorkflowNodeStatusUpdate 更新工作流节点状态。
func TaskWorkflowNodeStatusUpdate(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.TaskWorkflowNodeStatusUpdateRequest{}
	_ = gsgin.GinPostBody(c, &request)
	workflowID := request.WorkflowID
	// 传了 home_task_id 则自动查找 workflow_id
	if request.HomeTaskID > 0 {
		info, err := common.DbMain.TaskWorkflowCreateOrGetByHomeTaskID(request.HomeTaskID)
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		workflowID = cast.ToInt(info[`id`])
	}
	if workflowID <= 0 {
		gsgin.GinResponseError(c, `workflow_id或home_task_id不能为空`, nil)
		return
	}
	step := strings.TrimSpace(request.Step)
	nodeStatuses := strings.TrimSpace(request.NodeStatuses)
	// 传了 step 则自动合并，否则直接用 node_statuses（前端手动切换场景）
	if step != `` {
		merged, err := taskWorkflowMergeNodeStatus(workflowID, step, strings.TrimSpace(request.Status))
		if err != nil {
			gsgin.GinResponseError(c, err.Error(), nil)
			return
		}
		nodeStatuses = merged
	}
	if nodeStatuses == `` {
		gsgin.GinResponseError(c, `node_statuses或step不能同时为空`, nil)
		return
	}
	if err := common.DbMain.TaskWorkflowUpdateNodeStatuses(workflowID, nodeStatuses); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	taskWorkflowBroadcastNodeStatus(workflowID, nodeStatuses)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow_id`: workflowID,
	})
}

// taskWorkflowMergeNodeStatus 加载现有 node_statuses 并合并指定 step + status，返回合并后的 JSON 字符串。
func taskWorkflowMergeNodeStatus(workflowID int, step, status string) (string, error) {
	info, err := common.DbMain.TaskWorkflowInfo(workflowID)
	if err != nil {
		return ``, err
	}
	nodeStatuses := map[string]string{}
	if raw := strings.TrimSpace(cast.ToString(info[`node_statuses`])); raw != `` {
		_ = json.Unmarshal([]byte(raw), &nodeStatuses)
	}
	nodeStatuses[step] = status
	data, err := json.Marshal(nodeStatuses)
	if err != nil {
		return ``, err
	}
	return string(data), nil
}

// TaskWorkflowPromptsSave 保存工作流提示词。
func TaskWorkflowPromptsSave(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.TaskWorkflowPromptsSaveRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if request.WorkflowID <= 0 {
		gsgin.GinResponseError(c, `workflow_id不能为空`, nil)
		return
	}

	if err := common.DbMain.WorkflowStepPromptsSave(request.WorkflowID, request.StepKey, request.StepPrompt); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	updatedInfo, err := common.DbMain.TaskWorkflowInfo(request.WorkflowID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`: updatedInfo,
	})
}

// TaskWorkflowPromptsRestore 还原工作流提示词为默认值。
// 优先从模板步骤的 prompt_content 还原；若模板不存在则回退到旧全局配置。
func TaskWorkflowPromptsRestore(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.TaskWorkflowPromptsRestoreRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if request.WorkflowID <= 0 {
		gsgin.GinResponseError(c, `workflow_id不能为空`, nil)
		return
	}
	workflowInfo, err := common.DbMain.TaskWorkflowInfo(request.WorkflowID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	homeTaskID := cast.ToInt(workflowInfo[`home_task_id`])
	// 提前获取任务信息，两个分支都需要
	homeTaskInfo, homeTaskErr := common.DbMain.HomeTaskRow(homeTaskID)
	if homeTaskErr != nil {
		gsgin.GinResponseError(c, homeTaskErr.Error(), nil)
		return
	}
	_, templateSteps, templateErr := common.DbMain.HomeTaskWorkflowTemplateSteps(homeTaskID)

	if templateErr != nil || len(templateSteps) == 0 {
		gsgin.GinResponseError(c, `未关联工作流模板`, nil)
		return
	}
	// 从模板步骤还原，先解析占位符再写入
	placeholders := buildTaskWorkflowPlaceholderMap(c, homeTaskInfo, workflowInfo)
	// 复制模板步骤并解析 prompt_content 中的占位符
	resolvedSteps := make([]map[string]any, len(templateSteps))
	for i, step := range templateSteps {
		stepCopy := make(map[string]any, len(step))
		for k, v := range step {
			stepCopy[k] = v
		}
		promptContent := strings.TrimSpace(cast.ToString(step[`prompt_content`]))
		if promptContent != `` {
			// 注入步骤级占位符：{步骤ID} 替换为当前步骤的 step_key
			placeholders[`{步骤ID}`] = cast.ToString(step[`step_key`])
			stepCopy[`prompt_content`] = taskWorkflowResolvePlaceholders(promptContent, placeholders)
		}
		resolvedSteps[i] = stepCopy
	}
	if restoreErr := common.DbMain.WorkflowStepPromptsRestore(request.WorkflowID, resolvedSteps); restoreErr != nil {
		gsgin.GinResponseError(c, restoreErr.Error(), nil)
		return
	}

	// 清空所有提示词类型对应的执行历史
	allPromptTypes := []string{`requirement`, `design`, `api_dev`, `code_review`, `browser_test`, `api_test`}
	// 使用模板时，附加模板步骤的 step_key（如 custom_3、issue_fix 等）
	if templateErr == nil && len(templateSteps) > 0 {
		seen := make(map[string]bool, len(allPromptTypes)+len(templateSteps))
		for _, k := range allPromptTypes {
			seen[k] = true
		}
		for _, step := range templateSteps {
			stepKey := cast.ToString(step[`step_key`])
			if stepKey != `` && !seen[stepKey] {
				seen[stepKey] = true
				allPromptTypes = append(allPromptTypes, stepKey)
			}
		}
	}
	for _, promptType := range allPromptTypes {
		_ = common.DbMain.TaskWorkflowClearChatSessionIDs(request.WorkflowID, promptType)
	}
	updatedInfo, err := common.DbMain.TaskWorkflowInfo(request.WorkflowID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	workflowDocuments, _ := common.DbMain.TaskWorkflowDocumentList(request.WorkflowID)
	hasWorkflowDocuments := common.DbMain.TaskWorkflowDocumentHasRecords(request.WorkflowID)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`:               updatedInfo,
		`documents`:              workflowDocuments,
		`has_workflow_documents`: hasWorkflowDocuments,
	})
}

// taskWorkflowBuildBasePlaceholderMap 构建基础占位符映射（不含步骤文档占位符）。
func taskWorkflowBuildBasePlaceholderMap(c *gin.Context, homeTaskInfo map[string]any, workflowInfo map[string]any) map[string]string {
	apiHost := taskWorkflowBuildAPIHost(c)
	devEnvironment, _ := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigDevEnvironment)
	result := map[string]string{
		`{需求文档地址}`:        taskWorkflowBuildShareURL(c, workflowInfo, apiHost),
		`{需求文档纯文本地址}`:     taskWorkflowBuildPlainTextShareURL(c, workflowInfo, apiHost),
		`{需求文档纯文本文件相对地址}`: taskWorkflowBuildPlainTextFragmentRelativePath(workflowInfo),
		`{需求设计方案文档地址}`:    taskWorkflowBuildDesignPlanShareURL(c, workflowInfo, apiHost),
		`{需求设计方案文件相对地址}`:  taskWorkflowBuildDesignPlanFragmentRelativePath(workflowInfo),
		`{接口开发API地址}`:     apiHost,
		`{接口开发API的token}`: taskWorkflowBuildAPIToken(c),
		`{开发项目配置}`:        taskWorkflowBuildDevConfigsMarkdown(homeTaskInfo),
		`{开发配置}`:          taskWorkflowBuildDevConfigsMarkdown(homeTaskInfo),
		`{自定义网页}`:         taskWorkflowBuildDevConfigsFieldMarkdown(homeTaskInfo, `smart_link`),
	}
	// 内置文档 ID 占位符：{xxx地址} -> {xxx地址ID}，映射为对应片段的 file_id
	result[`{需求文档地址ID}`] = common.TaskWorkflowParseFragmentRef(cast.ToString(workflowInfo[`requirement_fragment_id`]), taskWorkflowWorkflowFragmentFolderName(workflowInfo)).FileID
	result[`{需求文档纯文本地址ID}`] = taskWorkflowGetFragmentFileIDByName(cast.ToInt(workflowInfo[`id`]), `纯文本需求文档`)
	// 内置占位符 {任务名称}，替换为当前任务的名称
	result[`{任务名称}`] = cast.ToString(homeTaskInfo[`name`])
	// 内置占位符 {工作流程ID}，替换为工作流程任务的 ID
	result[`{工作流程ID}`] = cast.ToString(workflowInfo[`id`])
	// 内置占位符 {任务ID}，替换为该工作流程关联的任务ID（tbl_home_task表的id）
	result[`{任务ID}`] = cast.ToString(homeTaskInfo[`id`])
	// 动态读取 skills 目录，为每个子目录生成占位符 {xxx地址}
	skillsDir := filepath.Join(component.EnvClient.RootPath, `skills`)
	if skillEntries, skillErr := os.ReadDir(skillsDir); skillErr == nil {
		for _, entry := range skillEntries {
			if entry.IsDir() && !strings.HasPrefix(entry.Name(), `.`) {
				result[`{`+entry.Name()+`地址}`] = filepath.Join(skillsDir, entry.Name())
			}
		}
	}
	// 先解析开发环境内容中的其他占位符，再将其加入映射。
	for key, value := range result {
		devEnvironment = strings.ReplaceAll(devEnvironment, key, value)
	}
	result[`{开发环境}`] = devEnvironment
	return result
}

// buildTaskWorkflowPlaceholderMap 根据任务信息构建占位符替换映射，包含步骤文档占位符。
func buildTaskWorkflowPlaceholderMap(c *gin.Context, homeTaskInfo map[string]any, workflowInfo map[string]any) map[string]string {
	result := taskWorkflowBuildBasePlaceholderMap(c, homeTaskInfo, workflowInfo)
	apiHost := taskWorkflowBuildAPIHost(c)
	// 注入步骤文档占位符：{xxx地址}、{xxx文件相对地址} 与 {xxxID}
	workflowID := cast.ToInt(workflowInfo[`id`])
	if workflowID > 0 {
		// 优先从新文档表读取占位符
		documents, docErr := common.DbMain.TaskWorkflowDocumentList(workflowID)
		if docErr == nil && len(documents) > 0 {
			for _, doc := range documents {
				placeholder := strings.TrimSpace(cast.ToString(doc[`placeholder`]))
				fileID := strings.TrimSpace(cast.ToString(doc[`file_id`]))
				if placeholder == `` || fileID == `` {
					continue
				}
				result[placeholder] = taskWorkflowBuildStepFragmentShareURL(fileID, apiHost)
				relativePlaceholder := common.WorkflowTemplateStepDocumentsToRelativePlaceholder(placeholder)
				if relativePlaceholder != `` {
					result[relativePlaceholder] = taskWorkflowBuildStepFragmentRelativePath(map[string]string{
						`file_id`:     fileID,
						`folder_name`: cast.ToString(doc[`folder_name`]),
					})
				}
				idPlaceholder := common.WorkflowTemplateStepDocumentsToIDPlaceholder(placeholder)
				if idPlaceholder != `` {
					result[idPlaceholder] = fileID
				}
			}
		}
	}
	return result
}

// taskWorkflowBuildStepFragmentShareURL 为步骤文档知识片段生成分享链接。
func taskWorkflowBuildStepFragmentShareURL(fileID, apiHost string) string {
	fileID = strings.TrimSpace(fileID)
	if fileID == `` || component.MemoryRuntime == nil {
		return ``
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``
	}
	if _, err := component.MemoryRuntime.DB().MemoryFragmentInfo(fileID); err != nil {
		return ``
	}
	shareStore := memoryFragmentShareStoreForRoot(component.MemoryRuntime.Config().Dir)
	share, err := shareStore.Create(fileID, time.Now())
	if err != nil {
		return ``
	}
	return memoryFragmentBuildShareURL(fileID, share.Token, apiHost)
}

// taskWorkflowBuildStepFragmentRelativePath 为步骤文档知识片段构建相对于 fragments/ 目录的相对路径。
func taskWorkflowBuildStepFragmentRelativePath(ref map[string]string) string {
	fileID := strings.TrimSpace(ref[`file_id`])
	if fileID == `` || component.MemoryRuntime == nil {
		return ``
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``
	}
	info, err := component.MemoryRuntime.DB().MemoryFragmentInfo(fileID)
	if err != nil {
		return ``
	}
	filePath := strings.TrimSpace(cast.ToString(info[`file_path`]))
	if filePath == `` {
		return ``
	}
	fragmentsDir := filepath.Join(component.MemoryRuntime.Config().Dir, `fragments`)
	rel, err := filepath.Rel(fragmentsDir, filePath)
	if err != nil {
		return ``
	}
	return strings.ReplaceAll(rel, `\`, `/`)
}

// taskWorkflowBuildAPIHost 从请求上下文构建 API 基地址。
func taskWorkflowBuildAPIHost(c *gin.Context) string {
	if c == nil || c.Request == nil {
		return ``
	}
	scheme := `http`
	if c.Request.TLS != nil {
		scheme = `https`
	}
	if forwarded := strings.TrimSpace(c.GetHeader(`X-Forwarded-Proto`)); forwarded != `` {
		scheme = forwarded
	}
	host := strings.TrimSpace(c.Request.Host)
	if host == `` {
		return ``
	}
	return scheme + `://` + host
}

// taskWorkflowBuildAPIToken 从请求上下文获取认证 token。
func taskWorkflowBuildAPIToken(c *gin.Context) string {
	if c == nil {
		return ``
	}
	token := strings.TrimSpace(c.GetHeader(`Authorization`))
	if token != `` {
		return token
	}
	token = strings.TrimSpace(c.GetHeader(`token`))
	if token == `` {
		token = define.DtoolAPIDefaultToken
	}
	return token
}

// taskWorkflowBuildFragmentFileID 从工作流信息中提取指定片段引用列的 file_id。
func taskWorkflowBuildFragmentFileID(workflowInfo map[string]any, column string) string {
	fragmentRef := common.TaskWorkflowParseFragmentRef(cast.ToString(workflowInfo[column]), taskWorkflowWorkflowFragmentFolderName(workflowInfo))
	return fragmentRef.FileID
}

// taskWorkflowBuildShareURL 为需求文档知识片段生成分享链接。
func taskWorkflowBuildShareURL(c *gin.Context, workflowInfo map[string]any, apiHost string) string {
	fragmentRef := common.TaskWorkflowParseFragmentRef(cast.ToString(workflowInfo[`requirement_fragment_id`]), taskWorkflowWorkflowFragmentFolderName(workflowInfo))
	if fragmentRef.FileID == `` || component.MemoryRuntime == nil {
		return ``
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``
	}
	// 确认片段存在。
	if _, err := component.MemoryRuntime.DB().MemoryFragmentInfo(fragmentRef.FileID); err != nil {
		return ``
	}
	shareStore := memoryFragmentShareStoreForRoot(component.MemoryRuntime.Config().Dir)
	share, err := shareStore.Create(fragmentRef.FileID, time.Now())
	if err != nil {
		return ``
	}
	return memoryFragmentBuildShareURL(fragmentRef.FileID, share.Token, apiHost)
}

// taskWorkflowBuildPlainTextShareURL 为纯文本需求知识片段生成分享链接。
func taskWorkflowBuildPlainTextShareURL(c *gin.Context, workflowInfo map[string]any, apiHost string) string {
	workflowID := cast.ToInt(workflowInfo[`id`])
	fileID := taskWorkflowGetFragmentFileIDByName(workflowID, `纯文本需求文档`)
	if fileID == `` || component.MemoryRuntime == nil {
		return ``
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``
	}
	if _, err := component.MemoryRuntime.DB().MemoryFragmentInfo(fileID); err != nil {
		return ``
	}
	shareStore := memoryFragmentShareStoreForRoot(component.MemoryRuntime.Config().Dir)
	share, err := shareStore.Create(fileID, time.Now())
	if err != nil {
		return ``
	}
	return memoryFragmentBuildShareURL(fileID, share.Token, apiHost)
}

// taskWorkflowResolvePlaceholders 批量替换模板中的占位符。
func taskWorkflowResolvePlaceholders(template string, placeholders map[string]string) string {
	result := template
	for key, value := range placeholders {
		result = strings.ReplaceAll(result, key, value)
	}
	return result
}

// taskWorkflowQuerySkillPaths 根据 git_ids 查询多个项目根目录，拼接 skill 子目录路径，逗号分隔。
func taskWorkflowQuerySkillPaths(homeTaskInfo map[string]any, subDir string) string {
	gitIDs := homeTaskGitIDs(homeTaskInfo)
	var paths []string
	for _, gitID := range gitIDs {
		info, err := common.DbMain.Client.QuickQuery(`tbl_git`, `code_path`, map[string]any{
			`id`: gitID,
		}).One()
		if err != nil || len(info) == 0 {
			continue
		}
		codePath := strings.TrimSpace(cast.ToString(info[`code_path`]))
		if codePath == `` {
			continue
		}
		paths = append(paths, filepath.Join(codePath, subDir))
	}
	return strings.Join(paths, `,`)
}

// homeTaskGitIDs 从 homeTaskInfo 解析 git_ids JSON，优先从 dev_configs 派生，回退到旧字段。
func homeTaskGitIDs(homeTaskInfo map[string]any) []int {
	// 优先从 dev_configs 派生。
	devConfigs := homeTaskDevConfigs(homeTaskInfo)
	if len(devConfigs) > 0 {
		var ids []int
		for _, cfg := range devConfigs {
			if cfg.GitID > 0 {
				ids = append(ids, cfg.GitID)
			}
		}
		if len(ids) > 0 {
			return ids
		}
	}
	gitIDsJSON := strings.TrimSpace(cast.ToString(homeTaskInfo[`git_ids`]))
	if gitIDsJSON != `` && gitIDsJSON != `[]` {
		var ids []int
		if err := json.Unmarshal([]byte(gitIDsJSON), &ids); err == nil && len(ids) > 0 {
			return ids
		}
	}
	legacyID := cast.ToInt(homeTaskInfo[`git_id`])
	if legacyID > 0 {
		return []int{legacyID}
	}
	return nil
}

// homeTaskApiDevEntries 从 homeTaskInfo 解析 api_dev_entries JSON，优先从 dev_configs 派生，回退到旧字段。
func homeTaskApiDevEntries(homeTaskInfo map[string]any) []_struct.ApiDevEntry {
	// 优先从 dev_configs 派生。
	devConfigs := homeTaskDevConfigs(homeTaskInfo)
	if len(devConfigs) > 0 {
		var entries []_struct.ApiDevEntry
		for _, cfg := range devConfigs {
			if cfg.CollectionID > 0 {
				entries = append(entries, _struct.ApiDevEntry{CollectionID: cfg.CollectionID, DirID: cfg.DirID})
			}
		}
		if len(entries) > 0 {
			return entries
		}
	}
	entriesJSON := strings.TrimSpace(cast.ToString(homeTaskInfo[`api_dev_entries`]))
	if entriesJSON != `` && entriesJSON != `[]` {
		var entries []_struct.ApiDevEntry
		if err := json.Unmarshal([]byte(entriesJSON), &entries); err == nil && len(entries) > 0 {
			return entries
		}
	}
	colID := cast.ToInt(homeTaskInfo[`api_collection_id`])
	if colID > 0 {
		return []_struct.ApiDevEntry{{CollectionID: colID, DirID: cast.ToInt(homeTaskInfo[`api_dir_id`])}}
	}
	return nil
}

// homeTaskDevConfigs 从 homeTaskInfo 解析 dev_configs JSON，回退到旧字段构建。
func homeTaskDevConfigs(homeTaskInfo map[string]any) []_struct.DevConfig {
	configsJSON := strings.TrimSpace(cast.ToString(homeTaskInfo[`dev_configs`]))
	if configsJSON != `` && configsJSON != `[]` {
		var configs []_struct.DevConfig
		if err := json.Unmarshal([]byte(configsJSON), &configs); err == nil && len(configs) > 0 {
			return configs
		}
	}
	return nil
}

// taskWorkflowBuildDevConfigsMarkdown 将 dev_configs 转为 markdown 列表。
func taskWorkflowBuildDevConfigsMarkdown(homeTaskInfo map[string]any) string {
	devConfigs := homeTaskDevConfigs(homeTaskInfo)
	if len(devConfigs) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, cfg := range devConfigs {
		if i > 0 {
			sb.WriteString("\n")
		}
		sb.WriteString(fmt.Sprintf("#### 配置 %d\n", i+1))
		gitName := taskWorkflowQueryNameByID("tbl_git", cfg.GitID)
		gitCodePath := taskWorkflowQueryCodePath(cfg.GitID)
		if cfg.GitID > 0 {
			sb.WriteString(fmt.Sprintf("- **Git配置**: %s（ID: %d）\n", gitName, cfg.GitID))
		}
		if gitCodePath != "" {
			sb.WriteString(fmt.Sprintf("- **Git仓库路径**: %s\n", gitCodePath))
		}
		if cfg.LocalDir != "" {
			sb.WriteString(fmt.Sprintf("- **本地目录**: %s\n", cfg.LocalDir))
		}
		if cfg.ParentBranch != "" {
			sb.WriteString(fmt.Sprintf("- **父分支**: %s（用于提取当前分支改动文件）\n", cfg.ParentBranch))
		}
		if cfg.BranchName != "" {
			sb.WriteString(fmt.Sprintf("- **分支名**: %s\n", cfg.BranchName))
		}
		//if cfg.RuleEntryFile != "" {
		//	sb.WriteString(fmt.Sprintf("- **项目规则文件**: 你必须加载 `%s` 这个规则文件\n", cfg.RuleEntryFile))
		//}
		collectionName := taskWorkflowQueryNameByID("tbl_api_collection", cfg.CollectionID)
		if cfg.CollectionID > 0 {
			dirName := taskWorkflowQueryNameByID("tbl_api_dir", cfg.DirID)
			sb.WriteString(fmt.Sprintf("- **接口集合**: %s（ID: %d）\n", collectionName, cfg.CollectionID))
			if cfg.DirID > 0 {
				sb.WriteString(fmt.Sprintf("- **接口文件夹**: %s（ID: %d）\n", dirName, cfg.DirID))
			}
		}
		dockerName := taskWorkflowQueryNameByID("tbl_docker_compose", cfg.DockerID)
		if cfg.DockerID > 0 {
			sb.WriteString(fmt.Sprintf("- **Docker**: %s（ID: %d）\n", dockerName, cfg.DockerID))
		}
		mysqlName := taskWorkflowQueryNameByID("tbl_mysql", cfg.MysqlID)
		if cfg.MysqlID > 0 {
			sb.WriteString(fmt.Sprintf("- **MySQL**: %s（ID: %d）\n", mysqlName, cfg.MysqlID))
		}
		smartLinkName := taskWorkflowQuerySmartLinkLabel(cfg.SmartLinkID)
		if cfg.SmartLinkID > 0 {
			if cfg.SmartLinkAccount != "" {
				sb.WriteString(fmt.Sprintf("- **自定义网页**: %s（ID: %d, 账号: %s）\n", smartLinkName, cfg.SmartLinkID, cfg.SmartLinkAccount))
			} else {
				sb.WriteString(fmt.Sprintf("- **自定义网页**: %s（ID: %d）\n", smartLinkName, cfg.SmartLinkID))
			}
		}
	}
	return sb.String()
}

// taskWorkflowBuildDevConfigsFieldMarkdown 提取所有 dev_configs 中指定字段的值，构建 markdown 列表。
func taskWorkflowBuildDevConfigsFieldMarkdown(homeTaskInfo map[string]any, field string) string {
	devConfigs := homeTaskDevConfigs(homeTaskInfo)
	if len(devConfigs) == 0 {
		return ""
	}
	var sb strings.Builder
	for i, cfg := range devConfigs {
		var val string
		switch field {
		case `smart_link`:
			if cfg.SmartLinkID > 0 {
				name := taskWorkflowQuerySmartLinkLabel(cfg.SmartLinkID)
				if cfg.SmartLinkAccount != "" {
					val = fmt.Sprintf("%s（ID: %d, 账号: %s）", name, cfg.SmartLinkID, cfg.SmartLinkAccount)
				} else {
					val = fmt.Sprintf("%s（ID: %d）", name, cfg.SmartLinkID)
				}
			}
		}
		if val != "" {
			if sb.Len() > 0 {
				sb.WriteString("\n")
			}
			sb.WriteString(fmt.Sprintf("- 配置%d: %s", i+1, val))
		}
	}
	return sb.String()
}

// taskWorkflowQueryCodePath 根据 git_id 查询 tbl_git 表的 code_path 字段。
func taskWorkflowQueryCodePath(gitID int) string {
	if gitID <= 0 {
		return ""
	}
	info, err := common.DbMain.Client.QuickQuery("tbl_git", "code_path", map[string]any{
		"id": gitID,
	}).One()
	if err != nil || len(info) == 0 {
		return ""
	}
	return cast.ToString(info["code_path"])
}

// taskWorkflowQueryNameByID 根据 ID 查询表中的 name 字段。
func taskWorkflowQueryNameByID(tableName string, id int) string {
	if id <= 0 {
		return ""
	}
	info, err := common.DbMain.Client.QuickQuery(tableName, "name", map[string]any{
		"id": id,
	}).One()
	if err != nil || len(info) == 0 {
		return ""
	}
	return cast.ToString(info["name"])
}

// taskWorkflowQuerySmartLinkLabel 根据 ID 查询 smart_link 新表的 label 字段。
func taskWorkflowQuerySmartLinkLabel(id int) string {
	if id <= 0 {
		return ""
	}
	info, err := common.DbMain.Client.QuickQuery("smart_link", "label", map[string]any{
		"id": id,
	}).One()
	if err != nil || len(info) == 0 {
		return ""
	}
	return cast.ToString(info["label"])
}

// TaskWorkflowRequirementFetch 执行工作流首节点：抓取需求并直接写入知识片段。
func TaskWorkflowRequirementFetch(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.TaskWorkflowRequirementFetchRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if request.WorkflowID <= 0 {
		gsgin.GinResponseError(c, `workflow_id不能为空`, nil)
		return
	}
	workflowInfo, memoryDB, homeTaskInfo, ok := taskWorkflowLoadContextForDevPlanByID(c, request.WorkflowID)
	if !ok {
		return
	}
	fetchType := strings.TrimSpace(strings.ToLower(cast.ToString(homeTaskInfo[`fetch_type`])))
	if fetchType == `` {
		fetchType = `tapd`
	}
	sourceName := requirementFetchSourceName(fetchType)
	sourceURL := strings.TrimSpace(cast.ToString(homeTaskInfo[`tapd_url`]))
	if fetchType == `zentao` {
		sourceURL = strings.TrimSpace(cast.ToString(homeTaskInfo[`zentao_url`]))
	}
	if sourceURL == `` {
		gsgin.GinResponseError(c, `当前任务未配置`+sourceName+`地址`, nil)
		return
	}
	fragmentRef := common.TaskWorkflowParseFragmentRef(cast.ToString(workflowInfo[`requirement_fragment_id`]), taskWorkflowWorkflowFragmentFolderName(workflowInfo))
	if fragmentRef.FileID == `` {
		gsgin.GinResponseError(c, `需求知识片段未绑定`, nil)
		return
	}
	existingFragment, err := memoryDB.MemoryFragmentInfo(fragmentRef.FileID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), `load_config`, `running`, `开始读取 `+sourceName+` 抓取配置`)
	if err = common.DbMain.TaskWorkflowMarkRequirementFetchRunning(request.WorkflowID, sourceURL); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	resultMap, err := buildAsyncHomeTaskRequirementScrapeResultWithLog(fetchType, sourceURL, fragmentRef.FileID, func(step, message string) {
		taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), taskWorkflowNormalizeFetchStep(step), `running`, message)
	})
	if err != nil {
		_ = common.DbMain.TaskWorkflowMarkRequirementFetchFailed(request.WorkflowID, sourceURL, err.Error())
		taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), `error`, `failed`, err.Error())
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	markdown := cast.ToString(resultMap[`markdown`])
	if strings.TrimSpace(markdown) == `` {
		err = fmt.Errorf(`抓取结果为空`)
		_ = common.DbMain.TaskWorkflowMarkRequirementFetchFailed(request.WorkflowID, sourceURL, err.Error())
		taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), `error`, `failed`, err.Error())
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), `save_fragment`, `running`, `开始写入需求知识片段`)
	savedFragment, err := memoryDB.MemoryFragmentSave(
		fragmentRef.FileID,
		cast.ToString(existingFragment[`title`]),
		markdown,
		cast.ToStringSlice(existingFragment[`tags`]),
		cast.ToString(existingFragment[`folder_name`]),
	)
	if err != nil {
		_ = common.DbMain.TaskWorkflowMarkRequirementFetchFailed(request.WorkflowID, sourceURL, err.Error())
		taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), `error`, `failed`, err.Error())
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
	broadcastMemoryFragmentUpsert(savedFragment)
	if err = common.DbMain.TaskWorkflowMarkRequirementFetchSuccess(request.WorkflowID, sourceURL); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	taskWorkflowAutoCompleteNode(workflowInfo, `requirement-fetch`)
	updatedWorkflowInfo, err := common.DbMain.TaskWorkflowInfo(request.WorkflowID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), `done`, `success`, sourceName+` 需求抓取完成并已写入知识片段`)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`workflow`: updatedWorkflowInfo,
		`fragment`: map[string]any{
			`id`:      savedFragment[`id`],
			`file_id`: savedFragment[`file_id`],
			`title`:   savedFragment[`title`],
			`content`: savedFragment[`content`],
			`tags`:    savedFragment[`tags`],
		},
		`requirement_fetch_config`: taskWorkflowRequirementFetchConfig(),
	})
}

// taskWorkflowRequirementFetchConfig 返回当前工作流首节点抓取配置快照。
func taskWorkflowRequirementFetchConfig() map[string]any {
	return taskWorkflowRequirementFetchConfigByType(`tapd`)
}

func taskWorkflowRequirementFetchConfigByType(fetchType string) map[string]any {
	// 优先从自定义配置列表查找
	if cfg, found := findRequirementFetchConfig(fetchType); found {
		config := map[string]any{
			`smart_link_id`: cfg.SmartLinkID,
			`label`:         cfg.LinkLabel,
			`css_selector`:  cfg.CssSelector,
			`wait_seconds`:  cfg.WaitSeconds,
			`fetch_type`:    cfg.Type,
			`source_name`:   cfg.Name,
			`configured`:    true,
		}
		if config[`wait_seconds`].(int) <= 0 {
			config[`wait_seconds`] = defaultSmartLinkScrapeWaitSeconds
		}
		return config
	}
	// 回退旧独立key
	smartLinkIDKey, linkLabelKey, cssSelectorKey, waitSecondsKey := requirementFetchConfigKeys(fetchType)
	smartLinkIDStr, smartLinkErr := homeTaskConfigValue(smartLinkIDKey)
	label, labelErr := homeTaskConfigValue(linkLabelKey)
	cssSelector, selectorErr := homeTaskConfigValue(cssSelectorKey)
	waitSecondsStr, waitErr := homeTaskConfigValue(waitSecondsKey)
	config := map[string]any{
		`smart_link_id`: cast.ToInt(smartLinkIDStr),
		`label`:         strings.TrimSpace(label),
		`css_selector`:  strings.TrimSpace(cssSelector),
		`wait_seconds`:  cast.ToInt(waitSecondsStr),
		`fetch_type`:    strings.TrimSpace(strings.ToLower(fetchType)),
		`source_name`:   requirementFetchSourceName(fetchType),
		`configured`:    true,
	}
	if config[`wait_seconds`].(int) <= 0 {
		config[`wait_seconds`] = defaultSmartLinkScrapeWaitSeconds
	}
	if smartLinkErr != nil || labelErr != nil || selectorErr != nil || waitErr != nil {
		config[`configured`] = false
	}
	return config
}

// taskWorkflowBroadcastStep 向所有在线 SSE 客户端广播当前工作流步骤日志。
func taskWorkflowBroadcastStep(workflowID int, step, status, message string) {
	if workflowID <= 0 {
		return
	}
	distributeID := define.SseTaskWorkflowPrefix + cast.ToString(workflowID)
	msg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: distributeID,
		Data: map[string]any{
			`workflow_id`: workflowID,
			`step`:        strings.TrimSpace(step),
			`status`:      strings.TrimSpace(status),
			`message`:     strings.TrimSpace(message),
			`time`:        time.Now().Unix(),
		},
		Type: p_define.SseContentTypeMsg,
	})
	for _, item := range gsgin.SseStatus() {
		clientID := strings.TrimSpace(strings.TrimPrefix(item, apiDataChangeSseStatusPrefix))
		if clientID == `` || clientID == item {
			continue
		}
		sse := gsgin.SseGetByClientId(clientID)
		if sse == nil {
			continue
		}
		_ = sse.SendToChan(msg)
	}
}

// taskWorkflowBroadcastNodeStatus 通过 SSE 向所有在线客户端广播工作流节点状态变更。
func taskWorkflowBroadcastNodeStatus(workflowID int, nodeStatuses string) {
	if workflowID <= 0 {
		return
	}
	distributeID := define.SseTaskWorkflowPrefix + cast.ToString(workflowID)
	msg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: distributeID,
		Data: map[string]any{
			`type`:          `node_status_change`,
			`workflow_id`:   workflowID,
			`node_statuses`: nodeStatuses,
		},
		Type: p_define.SseContentTypeMsg,
	})
	for _, item := range gsgin.SseStatus() {
		clientID := strings.TrimSpace(strings.TrimPrefix(item, apiDataChangeSseStatusPrefix))
		if clientID == `` || clientID == item {
			continue
		}
		sse := gsgin.SseGetByClientId(clientID)
		if sse == nil {
			continue
		}
		_ = sse.SendToChan(msg)
	}
}

// taskWorkflowNormalizeFetchStep 将旧日志步骤映射为前端统一步骤名。
func taskWorkflowNormalizeFetchStep(step string) string {
	step = strings.TrimSpace(step)
	switch step {
	case `读取配置`:
		return `load_config`
	case `下发任务`:
		return `dispatch_scrape`
	case `处理结果`:
		return `process_zip`
	case `保存片段`:
		return `save_fragment`
	default:
		return `progress`
	}
}

// taskWorkflowBuildPlainTextFragmentRelativePath 为纯文本需求知识片段构建相对于 fragments/ 目录的相对路径。
func taskWorkflowBuildPlainTextFragmentRelativePath(workflowInfo map[string]any) string {
	workflowID := cast.ToInt(workflowInfo[`id`])
	fileID := taskWorkflowGetFragmentFileIDByName(workflowID, `纯文本需求文档`)
	if fileID == `` || component.MemoryRuntime == nil {
		return ``
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``
	}
	info, err := component.MemoryRuntime.DB().MemoryFragmentInfo(fileID)
	if err != nil {
		return ``
	}
	filePath := strings.TrimSpace(cast.ToString(info[`file_path`]))
	if filePath == `` {
		return ``
	}
	memoryDir := component.MemoryRuntime.Config().Dir
	relPath, err := filepath.Rel(memoryDir, filePath)
	if err != nil {
		return ``
	}
	relPath = filepath.ToSlash(relPath)
	if relPath == `.` || strings.HasPrefix(relPath, `../`) {
		return ``
	}
	return relPath
}

// taskWorkflowBuildDesignPlanShareURL 为需求设计方案知识片段生成分享链接。
func taskWorkflowBuildDesignPlanShareURL(c *gin.Context, workflowInfo map[string]any, apiHost string) string {
	workflowID := cast.ToInt(workflowInfo[`id`])
	fileID := taskWorkflowGetFragmentFileIDByName(workflowID, `设计方案需求文档`)
	if fileID == `` || component.MemoryRuntime == nil {
		return ``
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``
	}
	if _, err := component.MemoryRuntime.DB().MemoryFragmentInfo(fileID); err != nil {
		return ``
	}
	shareStore := memoryFragmentShareStoreForRoot(component.MemoryRuntime.Config().Dir)
	share, err := shareStore.Create(fileID, time.Now())
	if err != nil {
		return ``
	}
	return memoryFragmentBuildShareURL(fileID, share.Token, apiHost)
}

// taskWorkflowBuildDesignPlanFragmentRelativePath 为需求设计方案知识片段构建相对路径。
func taskWorkflowBuildDesignPlanFragmentRelativePath(workflowInfo map[string]any) string {
	workflowID := cast.ToInt(workflowInfo[`id`])
	fileID := taskWorkflowGetFragmentFileIDByName(workflowID, `设计方案需求文档`)
	if fileID == `` || component.MemoryRuntime == nil {
		return ``
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``
	}
	info, err := component.MemoryRuntime.DB().MemoryFragmentInfo(fileID)
	if err != nil {
		return ``
	}
	filePath := strings.TrimSpace(cast.ToString(info[`file_path`]))
	if filePath == `` {
		return ``
	}
	memoryDir := component.MemoryRuntime.Config().Dir
	relPath, err := filepath.Rel(memoryDir, filePath)
	if err != nil {
		return ``
	}
	relPath = filepath.ToSlash(relPath)
	if relPath == `.` || strings.HasPrefix(relPath, `../`) {
		return ``
	}
	return relPath
}

// ensureTaskWorkflowStepFragments 根据模板步骤的文档配置，为工作流预生成所有步骤知识片段。
// 注意：此函数需要先创建所有片段并保存引用，再统一替换提示词占位符，避免顺序问题导致替换不到。
func ensureTaskWorkflowStepFragments(c *gin.Context, workflowInfo map[string]any, homeTaskInfo map[string]any, templateSteps []map[string]any) {
	if component.MemoryRuntime == nil {
		gstool.FmtPrintlnLogTime("[ensureTaskWorkflowStepFragments] MemoryRuntime 为 nil，跳过创建步骤文档片段")
		return
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		gstool.FmtPrintlnLogTime("[ensureTaskWorkflowStepFragments] 记忆库未配置: %v，跳过创建步骤文档片段", err)
		return
	}
	memoryDB := component.MemoryRuntime.DB()
	if memoryDB == nil {
		gstool.FmtPrintlnLogTime("[ensureTaskWorkflowStepFragments] memoryDB 为 nil，跳过创建步骤文档片段")
		return
	}
	workflowID := cast.ToInt(workflowInfo[`id`])
	if workflowID <= 0 {
		gstool.FmtPrintlnLogTime("[ensureTaskWorkflowStepFragments] workflowID=%d 无效，跳过", workflowID)
		return
	}
	folderName := taskWorkflowWorkflowFragmentFolderName(workflowInfo)
	basePlaceholders := taskWorkflowBuildBasePlaceholderMap(c, homeTaskInfo, workflowInfo)
	// 仅保留抓取需求文档的绑定引用（requirement_fragment_id）
	reqFileID := common.TaskWorkflowParseFragmentRef(cast.ToString(workflowInfo[`requirement_fragment_id`]), folderName).FileID

	templateID, _ := common.DbMain.HomeTaskWorkflowTemplateID(cast.ToInt(workflowInfo[`home_task_id`]))
	for _, step := range templateSteps {
		stepKey := cast.ToString(step[`step_key`])
		stepID := cast.ToInt(step[`id`])
		// 注入步骤级占位符：{步骤ID} 替换为当前步骤的 step_key
		basePlaceholders[`{步骤ID}`] = stepKey
		docs := common.WorkflowTemplateStepDocumentsParse(cast.ToString(step[`step_documents`]))
		if len(docs) == 0 {
			continue
		}
		for _, doc := range docs {
			placeholder := strings.TrimSpace(doc.Placeholder)
			// 仅对 {需求文档地址} 占位符绑定到已有的抓取需求片段
			if placeholder == `{需求文档地址}` && reqFileID != `` {
				_ = common.DbMain.TaskWorkflowDocumentUpsert(
					workflowID, doc.ID, doc.Name, common.TaskWorkflowDocTypeStepDocument,
					templateID, stepID, reqFileID, folderName, placeholder,
				)
				continue
			}
			// 按 document_id + template_step_id 精确查找已有步骤文档记录，避免误匹配其他步骤的文档片段
			existingFileID := taskWorkflowGetStepDocFileID(workflowID, doc.ID, stepID)
			if existingFileID != `` {
				if _, err := memoryDB.MemoryFragmentInfo(existingFileID); err == nil {
					continue
				}
			}
			title := taskWorkflowResolvePlaceholders(doc.Title, basePlaceholders)
			if strings.TrimSpace(title) == `` {
				taskName := strings.TrimSpace(cast.ToString(homeTaskInfo[`name`]))
				if taskName != `` {
					title = taskName + `-` + doc.Name
				} else {
					title = doc.Name
				}
			}
			content := taskWorkflowResolvePlaceholders(doc.Content, basePlaceholders)
			tags := []string{doc.Name}
			fragmentInfo, err := memoryDB.MemoryFragmentSave(0, title, content, tags, folderName)
			if err != nil {
				gstool.FmtPrintlnLogTime("[ensureTaskWorkflowStepFragments] workflowID=%d 创建片段失败 docID=%s docName=%s err=%v", workflowID, doc.ID, doc.Name, err)
				continue
			}
			component.MemoryRuntime.ScheduleSync()
			fragmentFileID := strings.TrimSpace(cast.ToString(fragmentInfo[`file_id`]))
			if fragmentFileID == `` {
				gstool.FmtPrintlnLogTime("[ensureTaskWorkflowStepFragments] workflowID=%d MemoryFragmentSave 返回空 file_id docID=%s docName=%s", workflowID, doc.ID, doc.Name)
				continue
			}
			_ = common.DbMain.TaskWorkflowDocumentUpsert(
				workflowID, doc.ID, doc.Name, common.TaskWorkflowDocTypeStepDocument,
				templateID, stepID, fragmentFileID, cast.ToString(fragmentInfo[`folder_name`]), placeholder,
			)
		}
	}
}

// TaskWorkflowApiDocReset 重置接口文档，将所有关联文件夹下的接口 Markdown 合并覆盖到知识片段中。
// 支持 step_key 参数：指定步骤 key 时，从文档表中查找该步骤的 is_api_doc 文档片段并更新。
func TaskWorkflowApiDocReset(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.TaskWorkflowInfoRequest{}
	_ = gsgin.GinPostBody(c, &request)
	workflowInfo, err := common.DbMain.TaskWorkflowInfo(request.WorkflowID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	// 确定要写入的目标片段 ID
	var fragmentFileID string
	stepKey := strings.TrimSpace(request.StepKey)
	if stepKey != `` {
		fragmentFileID = taskWorkflowFindApiDocFragmentID(workflowInfo, stepKey)
	} else {
		workflowID := cast.ToInt(workflowInfo[`id`])
		fragmentFileID = taskWorkflowGetFragmentFileIDByName(workflowID, `接口文档`)
	}
	if fragmentFileID == `` {
		gsgin.GinResponseError(c, `接口文档片段未创建`, nil)
		return
	}

	memoryDB, ok := taskWorkflowMemoryDBOrResponse(c)
	if !ok {
		return
	}
	homeTaskInfo, err := common.DbMain.HomeTaskRow(cast.ToInt(workflowInfo[`home_task_id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	// 解析 dev_configs 获取所有关联文件夹
	devConfigsStr := strings.TrimSpace(cast.ToString(homeTaskInfo[`dev_configs`]))
	if devConfigsStr == `` || devConfigsStr == `[]` {
		gsgin.GinResponseError(c, `未配置开发环境`, nil)
		return
	}
	var devConfigs []_struct.DevConfig
	if err := json.Unmarshal([]byte(devConfigsStr), &devConfigs); err != nil {
		gsgin.GinResponseError(c, `开发配置解析失败`, nil)
		return
	}
	// 收集所有有 dir_id 的配置，生成 Markdown
	var sb strings.Builder
	sb.WriteString(`# 接口文档`)
	sb.WriteString(`

`)
	for _, cfg := range devConfigs {
		if cfg.DirID <= 0 {
			continue
		}
		folderMD := buildFolderApisMarkdown(cfg.DirID)
		if folderMD == `` {
			continue
		}
		// 用一级大标题分隔不同项目
		dir, _ := common.DbMain.Client.QuickQuery(`tbl_api_dir`, `*`, map[string]any{`id`: cfg.DirID}).One()
		folderName := `未命名项目`
		if len(dir) > 0 {
			folderName = cast.ToString(dir[`name`])
		}
		sb.WriteString(fmt.Sprintf(`# %s`, folderName))
		sb.WriteString(`

`)
		sb.WriteString(folderMD)
		sb.WriteString(`

`)
	}
	combinedMD := sb.String()
	if strings.TrimSpace(combinedMD) == `# 接口文档` || strings.TrimSpace(combinedMD) == `` {
		gsgin.GinResponseError(c, `未找到关联的接口文件夹`, nil)
		return
	}
	// 获取现有片段信息以保留标题和标签
	fragmentInfo, err := memoryDB.MemoryFragmentInfo(fragmentFileID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	fragmentTitle := cast.ToString(fragmentInfo[`title`])
	tags := []string{`接口文档`}
	if rawTags, ok := fragmentInfo[`tags`]; ok {
		if tagStr := cast.ToString(rawTags); tagStr != `` {
			var parsedTags []string
			if json.Unmarshal([]byte(tagStr), &parsedTags) == nil {
				tags = parsedTags
			}
		}
	}
	// 覆盖写入知识片段
	info, err := memoryDB.MemoryFragmentSave(fragmentFileID, fragmentTitle, combinedMD, tags, cast.ToString(fragmentInfo[`folder_name`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
	// 广播片段变更 SSE 事件，使前端文档 Tab 实时更新
	broadcastMemoryFragmentUpsert(info)
	gsgin.GinResponseSuccess(c, `接口文档已重置`, map[string]any{
		`fragment_id`: fragmentFileID,
	})
}

// taskWorkflowFindApiDocFragmentID 从文档表中查找指定步骤的 is_api_doc 文档的片段 ID。
func taskWorkflowFindApiDocFragmentID(workflowInfo map[string]any, stepKey string) string {
	workflowID := cast.ToInt(workflowInfo[`id`])
	if workflowID <= 0 || stepKey == `` {
		return ``
	}
	// 获取模板步骤，找到 is_api_doc 文档
	homeTaskID := cast.ToInt(workflowInfo[`home_task_id`])
	_, templateSteps, _ := common.DbMain.HomeTaskWorkflowTemplateSteps(homeTaskID)
	var targetDocID string
	for _, step := range templateSteps {
		if cast.ToString(step[`step_key`]) != stepKey {
			continue
		}
		docs := common.WorkflowTemplateStepDocumentsParse(cast.ToString(step[`step_documents`]))
		for _, doc := range docs {
			if doc.IsApiDoc {
				targetDocID = doc.ID
				break
			}
		}
		break
	}
	if targetDocID == `` {
		return ``
	}
	// 从文档表中查找匹配的文档记录
	documents, err := common.DbMain.TaskWorkflowDocumentList(workflowID)
	if err != nil {
		return ``
	}
	for _, doc := range documents {
		if cast.ToString(doc[`document_id`]) == targetDocID {
			return strings.TrimSpace(cast.ToString(doc[`file_id`]))
		}
	}
	return ``
}

// TaskWorkflowBatchNodeStatus 批量查询工作流节点状态。
func TaskWorkflowBatchNodeStatus(c *gin.Context) {
	if common.DbMain == nil || common.DbMain.Client == nil {
		gsgin.GinResponseError(c, `主库未初始化`, nil)
		return
	}
	request := _struct.TaskWorkflowBatchNodeStatusRequest{}
	_ = gsgin.GinPostBody(c, &request)
	if len(request.HomeTaskIDs) == 0 {
		gsgin.GinResponseSuccess(c, ``, map[string]any{})
		return
	}
	nodeStatusesMap, unreadCountMap, err := common.DbMain.TaskWorkflowBatchWorkflowSummaryByHomeTaskIDs(request.HomeTaskIDs)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`node_statuses_map`: nodeStatusesMap,
		`unread_count_map`:  unreadCountMap,
	})
}

// TaskWorkflowIssueFixResolve 解析问题修改提示词模板。
func TaskWorkflowIssueFixResolve(c *gin.Context) {
	var req _struct.TaskWorkflowInfoRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if req.WorkflowID <= 0 {
		gsgin.GinResponseError(c, `工作流id不能为空`, nil)
		return
	}
	workflowInfo, err := common.DbMain.TaskWorkflowInfo(req.WorkflowID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	homeTaskInfo, err := common.DbMain.HomeTaskRow(cast.ToInt(workflowInfo[`home_task_id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	template, err := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigPromptIssueFix)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	placeholders := buildTaskWorkflowPlaceholderMap(c, homeTaskInfo, workflowInfo)
	resolved := taskWorkflowResolvePlaceholders(template, placeholders)
	gsgin.GinResponseSuccess(c, ``, map[string]string{
		`prompt`: resolved,
	})
}

// extractSessionID 从首行 init JSON 中提取 session_id。
func extractSessionID(line string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return ``
	}
	return cast.ToString(data[`session_id`])
}

// startChatCommand 根据 chatID 启动 CLI 命令 goroutine。
// 从 DB 加载 chat 配置后根据 cli_type 分发到不同的 CLI 执行器。
// 当 sessionID 非空时，视为继续对话（isResume=true），CLI 使用 --resume 恢复会话。
func startChatCommand(chatID int64) {
	chatInfo, err := common.DbMain.TaskWorkflowChatInfo(chatID)
	if err != nil || len(chatInfo) == 0 {
		gstool.FmtPrintlnLogTime("[start-chat] chat_id=%d 查询对话信息失败: %v", chatID, err)
		return
	}
	localDir := cast.ToString(chatInfo[`local_dir`])
	prompt := cast.ToString(chatInfo[`prompt`])
	sessionID := cast.ToString(chatInfo[`session_id`])
	fromType := cast.ToString(chatInfo[`from_type`])
	// sessionID 非空表示继续对话（resume），用于正确标记 system_init 的 is_resume 字段
	isResume := sessionID != ``

	cliType := cast.ToString(chatInfo[`cli_type`])
	switch cliType {
	case `codex`:
		agentCliId := cast.ToInt(chatInfo[`agent_cli_id`])
		configJson := ``
		if agentCliId > 0 {
			cliRow, _ := common.DbMain.Client.QueryBySql(
				`SELECT config FROM tbl_agent_cli WHERE id = ?`, agentCliId,
			).One()
			if len(cliRow) > 0 {
				configJson = cast.ToString(cliRow[`config`])
			}
		}
		selectedModelName := strings.TrimSpace(cast.ToString(chatInfo[`model_name`]))
		go runCodexCommand(chatID, fromType, localDir, prompt, isResume, sessionID, configJson, selectedModelName)
	default:
		settingsPath := cast.ToString(chatInfo[`settings_path`])
		modelName, baseURL, apiKey := ``, ``, ``
		if settingsPath != `` {
			_, content, _ := business.ReadAgentCliSettings(settingsPath)
			if content != `` {
				modelName, baseURL, apiKey = business.GetAgentCliModelConfig(content)
			}
		}
		selectedModelName := strings.TrimSpace(cast.ToString(chatInfo[`model_name`]))
		if selectedModelName != `` {
			modelName = selectedModelName
		}
		thinkingIntensity := cast.ToString(chatInfo[`thinking_intensity`])
		thinkingBudget := define.ThinkingIntensityBudgetMap[thinkingIntensity]
		thinkingEffort := define.ThinkingIntensityEffortMap[thinkingIntensity]
		go runClaudeCommand(chatID, fromType, localDir, prompt, isResume, sessionID, baseURL, apiKey, modelName, settingsPath, thinkingEffort, thinkingBudget)
	}
}

// loadAgentCliChatRuntimeConfig 加载 Agent CLI 运行配置。
// loadAgentCliChatRuntimeConfig resolves settings/model metadata shared by workflow chats and standalone AgentCli chats.
func loadAgentCliChatRuntimeConfig(agentCliID int, requestedModelName string, c *gin.Context) (settingsPath string, modelName string, thinkingCollapsed int, ok bool) {
	settingsPath = ``
	modelName = strings.TrimSpace(requestedModelName)
	thinkingCollapsed = 0
	ok = true
	if agentCliID <= 0 {
		return
	}
	cliRow, err := common.DbMain.Client.QueryBySql(
		`SELECT * FROM tbl_agent_cli WHERE id = ?`, agentCliID,
	).One()
	if err != nil || len(cliRow) == 0 {
		gsgin.GinResponseError(c, `Agent Cli 实例不存在`, nil)
		ok = false
		return
	}
	if cast.ToInt(cliRow["enabled"]) != define.AgentCliEnabled {
		gsgin.GinResponseError(c, `Agent Cli 实例未启用`, nil)
		ok = false
		return
	}
	settingsPath = cast.ToString(cliRow["settings_path"])
	thinkingCollapsed = cast.ToInt(cliRow["thinking_collapsed"])
	return
}

// TaskWorkflowChatSend 启动新的 claude code 对话。
func TaskWorkflowChatSend(c *gin.Context) {
	var req _struct.TaskWorkflowChatSendRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if req.WorkflowID <= 0 {
		gsgin.GinResponseError(c, `工作流id不能为空`, nil)
		return
	}
	if strings.TrimSpace(req.Prompt) == `` {
		gsgin.GinResponseError(c, `提示词不能为空`, nil)
		return
	}
	if strings.TrimSpace(req.LocalDir) == `` {
		gsgin.GinResponseError(c, `请选择工作目录`, nil)
		return
	}
	if strings.TrimSpace(req.CliType) == `` {
		req.CliType = `claude`
	}

	settingsPath, modelName, thinkingCollapsed, ok := loadAgentCliChatRuntimeConfig(req.AgentCliId, req.ModelName, c)
	if !ok {
		return
	}

	// prompt_type 为可选，非空时用于按类型查询对话历史
	promptType := strings.TrimSpace(req.PromptType)

	chatID, err := common.DbMain.TaskWorkflowChatCreate(req.WorkflowID, req.Prompt, promptType, req.CliType, req.AgentCliId, req.LocalDir, settingsPath, modelName, thinkingCollapsed, req.ThinkingIntensity)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	go startChatCommand(chatID)

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`chat_id`: chatID,
	})
}

// AgentChatSend 启动新的 AgentCli 独立对话。
// AgentChatSend starts a standalone AgentCli execution without requiring any workflow id.
func AgentChatSend(c *gin.Context) {
	var req _struct.AgentChatSendRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if req.AgentCliId <= 0 {
		gsgin.GinResponseError(c, `agent_cli_id不能为空`, nil)
		return
	}
	if strings.TrimSpace(req.Prompt) == `` {
		gsgin.GinResponseError(c, `提示词不能为空`, nil)
		return
	}
	if strings.TrimSpace(req.LocalDir) == `` {
		gsgin.GinResponseError(c, `请选择工作目录`, nil)
		return
	}
	if strings.TrimSpace(req.CliType) == `` {
		req.CliType = `claude`
	}
	localDir := strings.TrimSpace(req.LocalDir)
	if err := validateAgentCliLocalDir(localDir); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	settingsPath, modelName, thinkingCollapsed, ok := loadAgentCliChatRuntimeConfig(req.AgentCliId, req.ModelName, c)
	if !ok {
		return
	}
	if err := validateAgentCliRuntimeConfig(req.AgentCliId, req.CliType); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	chatID, err := common.DbMain.AgentChatCreate(req.AgentCliId, req.Prompt, strings.TrimSpace(req.PromptType), req.CliType, localDir, settingsPath, modelName, thinkingCollapsed, req.ThinkingIntensity)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	go startChatCommand(chatID)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`chat_id`: chatID,
	})
}

// validateAgentCliLocalDir 校验独立 Agent CLI 执行使用的工作目录。
// 独立执行允许手工输入目录，需在入库前做存在性校验，避免把无效路径带入后续 CLI 启动层。
func validateAgentCliLocalDir(localDir string) error {
	localDir = strings.TrimSpace(localDir)
	if localDir == `` {
		return fmt.Errorf(`请选择工作目录`)
	}
	info, err := os.Stat(localDir)
	if err != nil || !info.IsDir() {
		return fmt.Errorf(`工作目录不存在或不是目录: %s`, localDir)
	}
	return nil
}

// TaskWorkflowChatContinue 继续已有对话。
func TaskWorkflowChatContinue(c *gin.Context) {
	var req _struct.TaskWorkflowChatContinueRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if req.ChatID <= 0 {
		gsgin.GinResponseError(c, `对话id不能为空`, nil)
		return
	}
	if strings.TrimSpace(req.Prompt) == `` {
		gsgin.GinResponseError(c, `提示词不能为空`, nil)
		return
	}
	gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d ========== 开始继续对话流程 ==========", req.ChatID)

	chatInfo, err := common.DbMain.TaskWorkflowChatInfo(int64(req.ChatID))
	if err != nil {
		gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d 查询对话信息失败: %v", req.ChatID, err)
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	currentStatus := cast.ToString(chatInfo[`status`])
	gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d 当前DB状态=%s prompt长度=%d", req.ChatID, currentStatus, len(req.Prompt))
	sessionID := cast.ToString(chatInfo[`session_id`])
	if sessionID == `` {
		gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d session_id为空，拒绝继续", req.ChatID)
		gsgin.GinResponseError(c, `对话未找到有效的 session_id`, nil)
		return
	}
	localDir := cast.ToString(chatInfo[`local_dir`])
	if localDir == `` {
		gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d local_dir为空，拒绝继续", req.ChatID)
		gsgin.GinResponseError(c, `对话未找到工作目录`, nil)
		return
	}

	chatID := int64(req.ChatID)

	// 防御：确保旧 goroutine 已完全退出（TaskWorkflowChatStop 中的 waitGoroutineExit 可能超时未完成）
	gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d 准备等待旧goroutine退出...", chatID)
	waitGoroutineExit(chatID)
	gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d 旧goroutine已退出（或超时）", chatID)

	// 检查 goroutine 是否正在运行（允许多个 SSE 连接共存，但不允许并发 goroutine）
	if _, running := chatCancelFuncs.Load(chatID); running {
		gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d goroutine仍在运行，拒绝继续", chatID)
		gsgin.GinResponseError(c, `对话正在执行中，请等待完成后再继续`, nil)
		return
	}
	gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d 确认无活跃goroutine", chatID)

	continuePrompt := strings.TrimSpace(req.Prompt)
	if err := common.DbMain.TaskWorkflowChatMarkRunning(chatID, continuePrompt); err != nil {
		gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d MarkRunning 失败: %v", chatID, err)
		gsgin.GinResponseError(c, `更新对话状态失败: `+err.Error(), nil)
		return
	}
	gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d 状态已设为 running（旧状态: %s），prompt已更新长度=%d", chatID, currentStatus, len(continuePrompt))

	// 立即验证DB状态是否真正更新为running
	verifyInfo, verifyErr := common.DbMain.TaskWorkflowChatInfo(chatID)
	if verifyErr != nil {
		gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d 验证查询失败: %v", chatID, verifyErr)
	} else {
		verifyStatus := cast.ToString(verifyInfo[`status`])
		gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d 验证DB状态=%s (期望running)", chatID, verifyStatus)
	}

	// 通知工作流页面刷新 chat 状态计数（执行历史按钮动画和状态数量）
	taskWorkflowBroadcastChatStatus(chatID)

	// 启动 CLI 命令 goroutine
	go startChatCommand(chatID)

	gstool.FmtPrintlnLogTime("[chat-continue] chat_id=%d ========== 继续对话流程完成，返回成功 ==========", chatID)
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`chat_id`: req.ChatID,
	})
}

// TaskWorkflowChatStop 停止运行中的对话。
// 先 context cancel 让 goroutine 正常退出，再通过 DB PID 强制杀进程兜底。
// 修复：等待 goroutine 完全退出后再返回，避免旧 goroutine 的清理代码与新的 Continue 操作发生竞态（如旧
// goroutine 的 error/completed 事件泄漏到新 SSE 连接、旧状态覆盖新的 MarkRunning 等）。
func TaskWorkflowChatStop(c *gin.Context) {
	var req _struct.TaskWorkflowChatStopRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if req.ChatID <= 0 {
		gsgin.GinResponseError(c, `对话id不能为空`, nil)
		return
	}
	chatID := int64(req.ChatID)
	// stopEvent 先落库，确保后续历史列表与通知判断都能识别"用户主动终止"。 // Persist the manual-stop marker first so history rows and notifications can distinguish user-initiated termination.
	_ = common.DbMain.TaskWorkflowChatAppendOutputBatch(chatID, []string{
		buildChatStopEvent(chatID, taskWorkflowChatStopReasonUserStop),
	})

	// 1. 先调 cancel 关闭 stopped 通道 + 取消 context，让 goroutine 感知用户主动停止
	//    使用 Load 而非 LoadAndDelete，让 goroutine 的 defer 自行调用 Delete
	cancelVal, ok := chatCancelFuncs.Load(chatID)
	if ok {
		if cancelFn, ok := cancelVal.(func()); ok {
			cancelFn()
		}
	}

	// 2. 兜底：通过 DB PID 强制杀进程（避免 cancel 未生效导致进程残留）
	killedPID := killProcessByChatID(chatID)

	// 3. 等待 goroutine 完全退出后再更新状态并返回
	//    旧 goroutine 的 defer 会在清理完成后调用 chatCancelFuncs.Delete(chatID)
	//    必须等待旧 goroutine 彻底退出，否则其后续的 sendLine（error/completed 事件广播）
	//    和状态更新（MarkInterrupted/MarkError）可能与新的 Continue 操作产生竞态
	waitGoroutineExit(chatID)

	_ = common.DbMain.TaskWorkflowChatMarkInterrupted(chatID)
	gsgin.GinResponseSuccess(c, `对话已停止`, map[string]any{
		`killed_pid`: killedPID,
	})
}

// waitGoroutineExit 等待指定 chatID 的 goroutine 从 chatCancelFuncs 中自行移除。
// goroutine 的 defer 会在 cleanup 完成后调用 chatCancelFuncs.Delete(chatID)。
func waitGoroutineExit(chatID int64) {
	const maxWait = 10 * time.Second
	const pollInterval = 50 * time.Millisecond
	deadline := time.Now().Add(maxWait)
	startTime := time.Now()
	for time.Now().Before(deadline) {
		if _, running := chatCancelFuncs.Load(chatID); !running {
			elapsed := time.Since(startTime)
			gstool.FmtPrintlnLogTime("[goroutine-exit] chat_id=%d goroutine已退出，耗时=%v", chatID, elapsed)
			return
		}
		time.Sleep(pollInterval)
	}
	gstool.FmtPrintlnLogTime("[goroutine-exit] chat_id=%d 等待 goroutine 退出超时(%v)，继续后续流程", chatID, maxWait)
}

// killProcessByChatID 从 agent_chat 表读取 pid 并强制杀进程兜底。
// 无论 Kill 是否成功（可能已被 cancel 正常退出），都返回 DB 中的 pid 供前端展示。
func killProcessByChatID(chatID int64) int {
	info, err := common.DbMain.TaskWorkflowChatInfo(chatID)
	if err != nil || len(info) == 0 {
		return 0
	}
	pid := cast.ToInt(info[`pid`])
	if pid <= 0 {
		return 0
	}
	// 兜底 Kill：cancel 可能已让进程退出，此处不判断返回值
	proc, err := os.FindProcess(pid)
	if err != nil {
		gstool.FmtPrintlnLogTime("[chat-stop] chat_id=%d FindProcess(pid=%d) 进程已不存在: %v", chatID, pid, err)
		return pid
	}
	if err := proc.Kill(); err != nil {
		gstool.FmtPrintlnLogTime("[chat-stop] chat_id=%d Kill(pid=%d) 失败（可能已被 cancel 终止）: %v", chatID, pid, err)
		return pid
	}
	gstool.FmtPrintlnLogTime("[chat-stop] chat_id=%d 已强制终止进程 pid=%d", chatID, pid)
	return pid
}

// buildAgentChatListResponse 构造统一的对话列表响应。
// buildAgentChatListResponse keeps workflow and standalone AgentCli history cards on the same response schema.
func buildAgentChatListResponse(rows []map[string]any) []map[string]any {
	const timeLayout = `2006-01-02 15:04:05`
	cliNameMap := make(map[int]string)
	cliRows, _ := common.DbMain.Client.QueryBySql(`SELECT id, name FROM tbl_agent_cli`).All()
	for _, cr := range cliRows {
		cliNameMap[cast.ToInt(cr[`id`])] = cast.ToString(cr[`name`])
	}
	list := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		status := cast.ToString(row[`status`])
		var durationMs int64
		if status != `running` {
			createdAt, _ := time.Parse(timeLayout, cast.ToString(row[`created_at`]))
			updatedAt, _ := time.Parse(timeLayout, cast.ToString(row[`updated_at`]))
			if !createdAt.IsZero() && !updatedAt.IsZero() {
				durationMs = updatedAt.Sub(createdAt).Milliseconds()
				if durationMs < 0 {
					durationMs = 0
				}
			}
		}
		rawOutput := cast.ToString(row[`raw_output`])
		lineCount := 0
		if rawOutput != `` {
			lineCount = len(strings.Split(rawOutput, "\n"))
		}
		stopReason, stopReasonText := taskWorkflowExtractTerminalReason(status, rawOutput)
		list = append(list, map[string]any{
			`id`:               cast.ToInt64(row[`id`]),
			`session_id`:       cast.ToString(row[`session_id`]),
			`prompt`:           cast.ToString(row[`prompt`]),
			`prompt_type`:      cast.ToString(row[`prompt_type`]),
			`agent_cli_id`:     cast.ToInt(row[`agent_cli_id`]),
			`agent_cli_name`:   cliNameMap[cast.ToInt(row[`agent_cli_id`])],
			`local_dir`:        cast.ToString(row[`local_dir`]),
			`workspace_path`:   cast.ToString(row[`local_dir`]),
			`status`:           status,
			`cli_type`:         cast.ToString(row[`cli_type`]),
			`created_at`:       cast.ToString(row[`created_at`]),
			`updated_at`:       cast.ToString(row[`updated_at`]),
			`duration_ms`:      durationMs,
			`line_count`:       lineCount,
			`is_read`:          cast.ToInt(row[`is_read`]) == 1,
			`stop_reason`:      stopReason,
			`stop_reason_text`: stopReasonText,
		})
	}
	return list
}

// TaskWorkflowChatList 列出工作流的所有对话。
func buildTaskWorkflowChatListSnapshot(workflowID int) ([]map[string]any, error) {
	if workflowID <= 0 {
		return []map[string]any{}, nil
	}
	rows, err := common.DbMain.TaskWorkflowChatList(workflowID)
	if err != nil {
		return nil, err
	}
	return buildAgentChatListResponse(rows), nil
}

func taskWorkflowBroadcastWorkflowDetail(workflowID int, eventType string, chatID int64) {
	if workflowID <= 0 {
		return
	}
	chatList, err := buildTaskWorkflowChatListSnapshot(workflowID)
	if err != nil {
		return
	}
	distributeID := define.SseTaskWorkflowPrefix + cast.ToString(workflowID)
	msg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: distributeID,
		Data: map[string]any{
			`type`:        strings.TrimSpace(eventType),
			`chat_id`:     chatID,
			`workflow_id`: workflowID,
			`chat_list`:   chatList,
		},
		Type: p_define.SseContentTypeMsg,
	})
	for _, item := range gsgin.SseStatus() {
		clientID := strings.TrimSpace(strings.TrimPrefix(item, apiDataChangeSseStatusPrefix))
		if clientID == `` || clientID == item {
			continue
		}
		sse := gsgin.SseGetByClientId(clientID)
		if sse == nil {
			continue
		}
		_ = sse.SendToChan(msg)
	}
}

func TaskWorkflowChatList(c *gin.Context) {
	var req _struct.TaskWorkflowChatListRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if req.WorkflowID <= 0 {
		gsgin.GinResponseError(c, `工作流id不能为空`, nil)
		return
	}
	rows, err := common.DbMain.TaskWorkflowChatList(req.WorkflowID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: buildAgentChatListResponse(rows),
	})
}

// TaskWorkflowChatDetail 获取对话详情（含原始输出行）。
func TaskWorkflowChatDetail(c *gin.Context) {
	var req _struct.TaskWorkflowChatDetailRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if req.ChatID <= 0 {
		gsgin.GinResponseError(c, `对话id不能为空`, nil)
		return
	}
	info, err := common.DbMain.TaskWorkflowChatInfo(int64(req.ChatID))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if len(info) == 0 {
		gsgin.GinResponseError(c, `对话不存在`, nil)
		return
	}

	// 孤立 running 状态检测：当 DB 状态为 running 但 goroutine 不存在时，标记为 interrupted
	chatID := int64(req.ChatID)
	chatStatus := cast.ToString(info[`status`])
	if chatStatus == "running" {
		if _, running := chatCancelFuncs.Load(chatID); !running {
			gstool.FmtPrintlnLogTime("[chat-detail] chat_id=%d DB 状态为 running 但 goroutine 不存在，标记为中断", chatID)
			_ = common.DbMain.TaskWorkflowChatMarkInterrupted(chatID)
			taskWorkflowBroadcastChatStatus(chatID)
			// 重新查询以获取更新后的状态
			info, _ = common.DbMain.TaskWorkflowChatInfo(chatID)
		}
	}

	rawOutput := cast.ToString(info[`raw_output`])
	lines := []string{}
	if rawOutput != `` {
		lines = strings.Split(rawOutput, "\n")
	}
	lastUsageSummary := extractChatLastUsageSummary(lines)

	modelName := ``
	agentCliName := ``
	selectedModelName := strings.TrimSpace(cast.ToString(info[`model_name`]))
	if agentCliId := cast.ToInt(info[`agent_cli_id`]); agentCliId > 0 {
		cliRow, err := common.DbMain.Client.QueryBySql(`SELECT name FROM tbl_agent_cli WHERE id = ?`, agentCliId).One()
		if err == nil && len(cliRow) > 0 {
			agentCliName = cast.ToString(cliRow["name"])
		}
	}
	if selectedModelName != `` {
		modelName = selectedModelName
	}

	taskName := ""
	// 只有工作流来源才反查任务名，避免 AgentCli 独立执行误关联任务。 // Only workflow-origin chats should resolve workflow task names.
	if cast.ToString(info["from_type"]) == common.AgentChatSourceTypeWorkflow {
		if workflowID := cast.ToInt(info["from_id"]); workflowID > 0 {
			wfRow, _ := common.DbMain.Client.QueryBySql(`SELECT home_task_id FROM tbl_task_workflow WHERE id = ?`, workflowID).One()
			if homeTaskId := cast.ToInt(wfRow["home_task_id"]); homeTaskId > 0 {
				htRow, _ := common.DbMain.Client.QueryBySql(`SELECT name FROM tbl_home_task WHERE id = ?`, homeTaskId).One()
				taskName = cast.ToString(htRow["name"])
			}
		}
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`chat_id`:            info[`id`],
		`session_id`:         info[`session_id`],
		`prompt`:             info[`prompt`],
		`agent_cli_id`:       info[`agent_cli_id`],
		`agent_cli_name`:     agentCliName,
		`cli_type`:           info[`cli_type`],
		`model_name`:         modelName,
		`task_name`:          taskName,
		`local_dir`:          info[`local_dir`],
		`status`:             info[`status`],
		`created_at`:         info[`created_at`],
		`thinking_collapsed`: info[`thinking_collapsed`],
		`thinking_intensity`: info[`thinking_intensity`],
		`from_type`:          info[`from_type`],
		`last_usage_summary`: lastUsageSummary,
		`lines`:              lines,
	})
}

// extractChatLastUsageSummary 从原始输出行中提取最近一次 token 统计，兼容 Claude/Codex 历史与运行态回填。
// extractChatLastUsageSummary extracts the latest token usage snapshot from raw output lines for both Claude and Codex chat details.
func extractChatLastUsageSummary(lines []string) map[string]any {
	for index := len(lines) - 1; index >= 0; index-- {
		line := strings.TrimSpace(lines[index])
		if line == `` {
			continue
		}
		var payload map[string]any
		if err := json.Unmarshal([]byte(line), &payload); err != nil {
			continue
		}
		summary := buildUsageSummaryFromPayload(payload)
		if summary != nil {
			return summary
		}
	}
	return nil
}

// buildUsageSummaryFromPayload 统一从事件负载中提取输入 token 和缓存命中 token。
// buildUsageSummaryFromPayload normalizes usage fields from different event payload shapes.
func buildUsageSummaryFromPayload(payload map[string]any) map[string]any {
	if len(payload) == 0 {
		return nil
	}
	if usage, ok := payload[`usage`].(map[string]any); ok {
		inputTokens := cast.ToInt64(usage[`input_tokens`])
		cacheReadInputTokens := cast.ToInt64(usage[`cache_read_input_tokens`])
		// 只有存在有效 token 统计时才返回，避免把空 usage 当成真实快照。
		// Only return when token stats are actually present, so empty usage objects do not override valid history.
		if inputTokens > 0 || cacheReadInputTokens > 0 {
			return map[string]any{
				`inputTokens`:          inputTokens,
				`cacheReadInputTokens`: cacheReadInputTokens,
			}
		}
	}
	if modelUsage, ok := payload[`modelUsage`].(map[string]any); ok {
		for _, row := range modelUsage {
			rowMap, ok := row.(map[string]any)
			if !ok {
				continue
			}
			inputTokens := cast.ToInt64(rowMap[`inputTokens`])
			cacheReadInputTokens := cast.ToInt64(rowMap[`cacheReadInputTokens`])
			if inputTokens > 0 || cacheReadInputTokens > 0 {
				return map[string]any{
					`inputTokens`:          inputTokens,
					`cacheReadInputTokens`: cacheReadInputTokens,
				}
			}
		}
	}
	return nil
}

// TaskWorkflowChatDirs 获取当前任务可选的工作目录列表。
func TaskWorkflowChatDirs(c *gin.Context) {
	var req _struct.TaskWorkflowChatDirsRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if req.WorkflowID <= 0 {
		gsgin.GinResponseError(c, `工作流id不能为空`, nil)
		return
	}

	workflowInfo, err := common.DbMain.TaskWorkflowInfo(req.WorkflowID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	homeTaskInfo, err := common.DbMain.HomeTaskRow(cast.ToInt(workflowInfo[`home_task_id`]))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	devConfigs := homeTaskDevConfigs(homeTaskInfo)
	seen := map[string]bool{}
	dirs := make([]string, 0)
	for _, cfg := range devConfigs {
		dir := strings.TrimSpace(cfg.LocalDir)
		if dir == `` || seen[dir] {
			continue
		}
		if info, statErr := os.Stat(dir); statErr != nil || !info.IsDir() {
			continue
		}
		seen[dir] = true
		dirs = append(dirs, dir)
	}
	if len(dirs) == 0 {
		gsgin.GinResponseError(c, `开发配置中没有可用的本地目录`, nil)
		return
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`dirs`: dirs,
	})
}

// TaskWorkflowChatListByPromptType 按提示词类型列出对话。
func TaskWorkflowChatListByPromptType(c *gin.Context) {
	var req _struct.TaskWorkflowChatListByPromptTypeRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if req.WorkflowID <= 0 {
		gsgin.GinResponseError(c, `工作流id不能为空`, nil)
		return
	}
	if strings.TrimSpace(req.PromptType) == `` {
		gsgin.GinResponseError(c, `提示词类型不能为空`, nil)
		return
	}
	rows, err := common.DbMain.TaskWorkflowChatListByPromptType(req.WorkflowID, req.PromptType)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: buildAgentChatListResponse(rows),
	})
}

// TaskWorkflowChatListByAgentCli 按 Agent CLI 列出对话。
// TaskWorkflowChatListByAgentCli returns standalone AgentCli execution history for one card.
func TaskWorkflowChatListByAgentCli(c *gin.Context) {
	var req _struct.TaskWorkflowChatListByAgentCliRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	agentChatListByAgentCli(c, req.AgentCliID)
}

// AgentChatListByAgentCli 按 Agent CLI 列出独立执行对话。
// AgentChatListByAgentCli exposes the dedicated AgentCli history endpoint while reusing the same query logic.
func AgentChatListByAgentCli(c *gin.Context) {
	var req _struct.AgentChatListByAgentCliRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	agentChatListByAgentCli(c, req.AgentCliID)
}

// AgentChatMarkRead 将历史对话标记为已读。
func AgentChatMarkRead(c *gin.Context) {
	var req _struct.AgentChatMarkReadRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if req.ChatID <= 0 {
		gsgin.GinResponseError(c, `chat_id不能为空`, nil)
		return
	}
	info, err := common.DbMain.TaskWorkflowChatInfo(int64(req.ChatID))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if len(info) == 0 {
		gsgin.GinResponseError(c, `对话不存在`, nil)
		return
	}
	if err := common.DbMain.AgentChatMarkRead(int64(req.ChatID)); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if cast.ToString(info[`from_type`]) == common.AgentChatSourceTypeWorkflow {
		taskWorkflowBroadcastWorkflowDetail(cast.ToInt(info[`from_id`]), `chat_read_change`, int64(req.ChatID))
	}
	taskWorkflowBroadcastUnreadChanged(int64(req.ChatID))
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`chat_id`:   req.ChatID,
		`is_read`:   true,
		`from_id`:   cast.ToInt(info[`from_id`]),
		`from_type`: cast.ToString(info[`from_type`]),
	})
}

// agentChatListByAgentCli 返回一个 Agent CLI 的独立执行历史。
// agentChatListByAgentCli is shared by the new Agent endpoint and the temporary workflow-compatible endpoint.
func agentChatListByAgentCli(c *gin.Context, agentCliID int) {
	if agentCliID <= 0 {
		gsgin.GinResponseError(c, `agent_cli_id不能为空`, nil)
		return
	}
	rows, err := common.DbMain.AgentChatListByAgentCli(agentCliID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: buildAgentChatListResponse(rows),
	})
}

// TaskWorkflowZcodeSave 保存 zcode 工作目录配置，自动扫描子文件夹解析项目映射。
func TaskWorkflowZcodeSave(c *gin.Context) {
	var req _struct.TaskWorkflowZcodeSaveRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	zcodeDir := strings.TrimSpace(req.ZcodeDir)
	if zcodeDir == `` {
		gsgin.GinResponseError(c, `zcode 工作目录地址不能为空`, nil)
		return
	}
	info, err := os.Stat(zcodeDir)
	if err != nil || !info.IsDir() {
		gsgin.GinResponseError(c, `zcode 工作目录不存在或不是目录: `+zcodeDir, nil)
		return
	}
	configID, err := common.DbMain.ZcodeConfigSave(zcodeDir)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	// 扫描子文件夹，解析 workspace-path.txt
	projects, scanErr := scanZcodeProjects(zcodeDir)
	if scanErr != nil {
		gsgin.GinResponseError(c, scanErr.Error(), nil)
		return
	}
	// 批量替换映射
	items := make([]common.ZcodeProjectMappingItem, 0, len(projects))
	for _, p := range projects {
		items = append(items, common.ZcodeProjectMappingItem{
			ProjectKey:    p.ProjectKey,
			WorkspacePath: p.WorkspacePath,
			SettingsPath:  p.SettingsPath,
		})
	}
	if err := common.DbMain.ZcodeProjectMappingReplace(configID, items); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`zcode_dir`: zcodeDir,
		`projects`:  projects,
	})
}

// TaskWorkflowZcodeGet 获取当前 zcode 配置及所有项目映射。
func TaskWorkflowZcodeGet(c *gin.Context) {
	config, err := common.DbMain.ZcodeConfigGet()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	zcodeDir := ``
	if config != nil {
		zcodeDir = cast.ToString(config[`zcode_dir`])
	}
	rows, err := common.DbMain.ZcodeProjectMappingList()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	projects := make([]_struct.TaskWorkflowZcodeProjectItem, 0, len(rows))
	for _, row := range rows {
		projects = append(projects, _struct.TaskWorkflowZcodeProjectItem{
			ProjectKey:    cast.ToString(row[`project_key`]),
			WorkspacePath: cast.ToString(row[`workspace_path`]),
			SettingsPath:  cast.ToString(row[`settings_path`]),
		})
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`zcode_dir`: zcodeDir,
		`projects`:  projects,
	})
}

// TaskWorkflowZcodeDelete 删除 zcode 配置及所有关联的项目映射。
func TaskWorkflowZcodeDelete(c *gin.Context) {
	config, err := common.DbMain.ZcodeConfigGet()
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	if config == nil {
		gsgin.GinResponseSuccess(c, `已删除`, nil)
		return
	}
	configID := cast.ToInt64(config[`id`])
	_ = common.DbMain.ZcodeProjectMappingReplace(configID, nil)
	if err := common.DbMain.ZcodeConfigDelete(); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, `已删除`, nil)
}

// zcodeLookupSettingsPath 根据本地目录精确匹配 zcode 项目映射中的 settings 路径。
func zcodeLookupSettingsPath(localDir string) string {
	localDir = strings.TrimSpace(localDir)
	if localDir == `` {
		return ``
	}
	row, err := common.DbMain.ZcodeProjectMappingGetByWorkspacePath(localDir)
	if err != nil || row == nil {
		return ``
	}
	return cast.ToString(row[`settings_path`])
}

// scanZcodeProjects 遍历 zcode 目录下的子文件夹，读取 workspace-path.txt 并构建映射。
func scanZcodeProjects(zcodeDir string) ([]_struct.TaskWorkflowZcodeProjectItem, error) {
	entries, err := os.ReadDir(zcodeDir)
	if err != nil {
		return nil, err
	}
	var projects []_struct.TaskWorkflowZcodeProjectItem
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		subDir := filepath.Join(zcodeDir, entry.Name())
		wsPathFile := filepath.Join(subDir, `workspace-path.txt`)
		data, err := os.ReadFile(wsPathFile)
		if err != nil {
			if !os.IsNotExist(err) {
				return nil, err
			}
			continue
		}
		workspacePath := strings.TrimSpace(string(data))
		if workspacePath == `` {
			continue
		}
		settingsPath := filepath.Join(subDir, `settings.json`)
		projects = append(projects, _struct.TaskWorkflowZcodeProjectItem{
			ProjectKey:    entry.Name(),
			WorkspacePath: workspacePath,
			SettingsPath:  filepath.ToSlash(settingsPath),
		})
	}
	return projects, nil
}

// runClaudeCommand 后台执行 claude 命令并通过业务 SSE 连接向所有活跃连接广播输出。
func runClaudeCommand(chatID int64, fromType string, localDir, prompt string, isResume bool, sessionID string, baseURL, apiKey, modelName, settingsPath, thinkingEffort string, thinkingBudget int) {

	defer func() {
		if r := recover(); r != nil {
			gstool.FmtPrintlnLogTime("[chat-run] chat_id=%d panic: %v", chatID, r)
			_ = common.DbMain.TaskWorkflowChatMarkError(chatID)
			broadcastChatLineToBusinessSse(chatID, fromType, buildChatCompletedEvent(chatID, `error`))
			taskWorkflowBroadcastChatStatus(chatID)
		}
	}()

	// DB 写入通道：将批量写 DB 与 SSE 推送解耦，避免 DB 写入阻塞 SSE 实时推送
	dbWriteCh := make(chan string, 4096)
	dbWriteDone := make(chan struct{})

	go func() {
		defer close(dbWriteDone)
		dbBuf := make([]string, 0, common.ChatOutputFlushBatchSize)
		flushTimer := time.NewTicker(2 * time.Second)
		defer flushTimer.Stop()

		flushDB := func() {
			if len(dbBuf) == 0 {
				return
			}
			lines := make([]string, len(dbBuf))
			copy(lines, dbBuf)
			dbBuf = dbBuf[:0]
			_ = common.DbMain.TaskWorkflowChatAppendOutputBatch(chatID, lines)
		}

		for {
			select {
			case line, ok := <-dbWriteCh:
				if !ok {
					flushDB()
					return
				}
				dbBuf = append(dbBuf, line)
				if len(dbBuf) >= common.ChatOutputFlushBatchSize {
					flushDB()
				}
			case <-flushTimer.C:
				flushDB()
			}
		}
	}()

	defer func() {
		close(dbWriteCh)
		<-dbWriteDone
	}()

	// sendLine SSE 实时推送，DB 通过独立 channel 攒批写入（不阻塞 SSE）
	sendLine := func(line string) {
		select {
		case dbWriteCh <- line:
		default:
		}
		broadcastChatLineToBusinessSse(chatID, fromType, line)
	}
	cfg := p_claude.RunConfig{
		Prompt:         prompt,
		SessionID:      sessionID,
		Model:          modelName,
		BaseURL:        baseURL,
		APIKey:         apiKey,
		WorkingDir:     localDir,
		UserDataDir:    p_claude.DefaultUserDataDir,
		SettingsPath:   settingsPath,
		Effort:         thinkingEffort,
		ThinkingBudget: thinkingBudget,
		ProcessStartCallback: func(pid int) {
			_ = common.DbMain.TaskWorkflowChatUpdatePID(chatID, pid)
		},
	}

	// 推送提示词到前端展示
	if prompt != "" {
		promptJSON, _ := json.Marshal(map[string]string{
			`type`:     `system`,
			`subtype`:  `command`,
			`cli_type`: `claude`,
			`cmd_line`: p_claude.BuildCommandLine(cfg),
			`text`:     prompt,
		})
		sendLine(string(promptJSON))
	}

	stopped := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	chatCancelFuncs.Store(chatID, func() {
		close(stopped)
		cancel()
	})
	defer func() {
		cancel()
		chatCancelFuncs.Delete(chatID)
	}()

	sessionExtracted := false
	callbackCount := 0
	var lastAssistantText string
	// lastResultText 来自 Claude stream-json 的 result 类型消息，是本轮 turn 的最终汇总文本，
	// 比逐条 assistant 流式累积更权威；webhook 通知优先使用该值，缺失时再回退到 lastAssistantText。
	var lastResultText string

	gstool.FmtPrintlnLogTime("[chat-run] chat_id=%d 开始RunClaudeStream dir=%s model=%s", chatID, localDir, modelName)

	_, err := p_claude.RunClaudeStream(ctx, cfg, func(msg p_claude.StreamMessage) {
		callbackCount++
		if callbackCount <= 3 {
			gstool.FmtPrintlnLogTime("[chat-run] callback:%d type=%s subtype=%s len=%d", callbackCount, msg.Type, msg.Subtype, len(msg.RawJSON))
		}
		rawJSON := msg.RawJSON
		if msg.Type == `system` && msg.Subtype == `init` {
			var initData map[string]any
			if err := json.Unmarshal([]byte(rawJSON), &initData); err == nil {
				initData[`is_resume`] = isResume
				// 为每次继续对话添加唯一时间戳，避免前端去重逻辑因同内容 system_init 行
				// 过滤掉后续继续轮次的"继续对话"分隔气泡。
				if isResume {
					initData[`continue_at`] = time.Now().UnixMilli()
				}
				if modified, e := json.Marshal(initData); e == nil {
					rawJSON = string(modified)
				}
			}
		}
		sendLine(rawJSON)

		// 跟踪最后一条 assistant 消息的文本内容（webhook 通知兜底用）
		if msg.Type == `assistant` {
			if text := extractAssistantText(msg.Data); text != "" {
				lastAssistantText = text
			}
		}
		// 跟踪 result 类型消息的 result 字段，作为 webhook 通知的首选内容
		if msg.Type == `result` {
			if text := strings.TrimSpace(cast.ToString(msg.Data[`result`])); text != "" {
				lastResultText = text
			}
		}

		if !sessionExtracted && !isResume && msg.Type == `system` && msg.Subtype == `init` {
			if sid, ok := msg.Data[`session_id`].(string); ok && sid != `` {
				_ = common.DbMain.TaskWorkflowChatUpdateSessionID(chatID, sid)
				sessionExtracted = true
			}
		}
	})

	gstool.FmtPrintlnLogTime("[chat-run] chat_id=%d RunClaudeStream结束 callbackCount=%d err=%v", chatID, callbackCount, err)

	if err != nil {
		errJSON, _ := json.Marshal(map[string]string{
			`type`: `error`,
			`text`: err.Error(),
		})
		sendLine(string(errJSON))
		// 区分用户主动停止与系统异常
		select {
		case <-stopped:
			gstool.FmtPrintlnLogTime("[chat-run] chat_id=%d 检测到stopped信号，设置状态为interrupted", chatID)
			_ = common.DbMain.TaskWorkflowChatMarkInterrupted(chatID)
		default:
			gstool.FmtPrintlnLogTime("[chat-run] chat_id=%d 异常退出，错误=%v，设置状态为error", chatID, err)
			_ = common.DbMain.TaskWorkflowChatMarkError(chatID)
		}
	} else {
		gstool.FmtPrintlnLogTime("[chat-run] chat_id=%d 正常完成，设置状态为completed", chatID)
		_ = common.DbMain.TaskWorkflowChatMarkCompleted(chatID)
	}
	finalStatus := common.TaskWorkflowChatStatusCompleted
	if err != nil {
		select {
		case <-stopped:
			finalStatus = common.TaskWorkflowChatStatusInterrupted
		default:
			finalStatus = common.TaskWorkflowChatStatusError
		}
	}
	gstool.FmtPrintlnLogTime("[chat-run] chat_id=%d 最终状态=%s，发送completed事件", chatID, finalStatus)
	sendLine(buildChatCompletedEvent(chatID, finalStatus))
	// 通知工作流页面刷新 chat 状态计数（执行历史按钮动画和状态数量）
	taskWorkflowBroadcastChatStatus(chatID)
	_ = lastAssistantText
	_ = lastResultText
}

// runCodexCommand 执行 Codex CLI 命令并通过业务 SSE 连接向所有活跃连接广播输出。
// 与 runClaudeCommand 平级，通过 cli_type 分发调用。
func runCodexCommand(chatID int64, fromType string, localDir, prompt string, isResume bool, sessionID string, configJson string, selectedModelName string) {
	startTime := time.Now()

	defer func() {
		if r := recover(); r != nil {
			gstool.FmtPrintlnLogTime("[codex-run] chat_id=%d panic: %v", chatID, r)
			_ = common.DbMain.TaskWorkflowChatMarkError(chatID)
			broadcastChatLineToBusinessSse(chatID, fromType, buildChatCompletedEvent(chatID, `error`))
			taskWorkflowBroadcastChatStatus(chatID)
		}
	}()

	// DB 写入通道：批量写 DB 与 SSE 推送解耦
	dbWriteCh := make(chan string, 4096)
	dbWriteDone := make(chan struct{})

	go func() {
		defer close(dbWriteDone)
		dbBuf := make([]string, 0, common.ChatOutputFlushBatchSize)
		flushTimer := time.NewTicker(2 * time.Second)
		defer flushTimer.Stop()

		flushDB := func() {
			if len(dbBuf) == 0 {
				return
			}
			lines := make([]string, len(dbBuf))
			copy(lines, dbBuf)
			dbBuf = dbBuf[:0]
			_ = common.DbMain.TaskWorkflowChatAppendOutputBatch(chatID, lines)
		}

		for {
			select {
			case line, ok := <-dbWriteCh:
				if !ok {
					flushDB()
					return
				}
				dbBuf = append(dbBuf, line)
				if len(dbBuf) >= common.ChatOutputFlushBatchSize {
					flushDB()
				}
			case <-flushTimer.C:
				flushDB()
			}
		}
	}()

	defer func() {
		close(dbWriteCh)
		<-dbWriteDone
	}()

	sendLine := func(line string) {
		select {
		case dbWriteCh <- line:
		default:
		}
		broadcastChatLineToBusinessSse(chatID, fromType, line)
	}
	codexCfg, cfgErr := business.GetCodexCliConfig(configJson)
	if cfgErr != nil {
		errJSON, _ := json.Marshal(map[string]string{
			`type`: `error`,
			`text`: `Codex CLI 配置解析失败: ` + cfgErr.Error(),
		})
		sendLine(string(errJSON))
		_ = common.DbMain.TaskWorkflowChatMarkError(chatID)
		sendLine(buildChatCompletedEvent(chatID, `error`))
		taskWorkflowBroadcastChatStatus(chatID)
		return
	}
	if strings.TrimSpace(selectedModelName) != `` {
		codexCfg.Model = strings.TrimSpace(selectedModelName)
	}

	cfg := p_codex.RunConfig{
		Prompt:      prompt,
		SessionID:   sessionID,
		Model:       codexCfg.Model,
		APIKey:      codexCfg.ApiKey,
		BaseURL:     codexCfg.BaseURL,
		WorkingDir:  localDir,
		SandboxMode: codexCfg.SandboxMode,
		ProcessStartCallback: func(pid int) {
			_ = common.DbMain.TaskWorkflowChatUpdatePID(chatID, pid)
		},
	}

	// 推送提示词到前端展示（复用 Claude 的 system/command 格式，前端可统一识别）
	if prompt != "" {
		promptJSON, _ := json.Marshal(map[string]string{
			`type`:     `system`,
			`subtype`:  `command`,
			`cli_type`: `codex`,
			`cmd_line`: p_codex.BuildCommandLine(cfg),
			`text`:     prompt,
		})
		sendLine(string(promptJSON))
	}

	// 继续对话时注入 system_init 行，前端解析为"继续对话"分隔气泡
	// Codex resume 不输出 thread.started 事件，需要手动补充
	if isResume {
		initJSON, _ := json.Marshal(map[string]any{
			`type`:        `system`,
			`subtype`:     `init`,
			`is_resume`:   true,
			`continue_at`: time.Now().UnixMilli(),
			`model`:       selectedModelName,
			`session_id`:  sessionID,
		})
		sendLine(string(initJSON))
	}

	stopped := make(chan struct{})
	ctx, cancel := context.WithCancel(context.Background())
	chatCancelFuncs.Store(chatID, func() {
		close(stopped)
		cancel()
	})
	defer func() {
		cancel()
		chatCancelFuncs.Delete(chatID)
	}()

	sessionExtracted := false
	callbackCount := 0
	var lastAgentMessageText string
	turnCount := 0
	var lastTurnUsage map[string]any

	gstool.FmtPrintlnLogTime("[codex-run] chat_id=%d 开始RunCodexStream dir=%s model=%s", chatID, localDir, codexCfg.Model)

	_, err := p_codex.RunCodexStream(ctx, cfg, func(msg p_codex.StreamMessage) {
		callbackCount++
		if callbackCount <= 3 {
			gstool.FmtPrintlnLogTime("[codex-run] callback:%d type=%s item_type=%s len=%d", callbackCount, msg.Type, msg.ItemType, len(msg.RawJSON))
		}
		sendLine(msg.RawJSON)

		// 跟踪最后一条 agent_message 的文本内容（webhook 通知用）
		if msg.ItemType == `agent_message` {
			if item, ok := msg.Data[`item`].(map[string]any); ok {
				if text := cast.ToString(item[`text`]); text != "" {
					lastAgentMessageText = text
				}
			}
		}
		if msg.Type == `turn.completed` {
			turnCount++
			if usage, ok := msg.Data[`usage`].(map[string]any); ok {
				lastTurnUsage = usage
			}
		}

		// 从 thread.started 提取 thread_id 作为 session_id
		if !sessionExtracted && !isResume && msg.Type == `thread.started` {
			if tid := cast.ToString(msg.Data[`thread_id`]); tid != `` {
				_ = common.DbMain.TaskWorkflowChatUpdateSessionID(chatID, tid)
				sessionExtracted = true
			}
		}
	})

	gstool.FmtPrintlnLogTime("[codex-run] chat_id=%d RunCodexStream结束 callbackCount=%d err=%v", chatID, callbackCount, err)

	if err != nil {
		errJSON, _ := json.Marshal(map[string]string{
			`type`: `error`,
			`text`: err.Error(),
		})
		sendLine(string(errJSON))
		select {
		case <-stopped:
			gstool.FmtPrintlnLogTime("[codex-run] chat_id=%d 检测到stopped信号，设置状态为interrupted", chatID)
			_ = common.DbMain.TaskWorkflowChatMarkInterrupted(chatID)
		default:
			gstool.FmtPrintlnLogTime("[codex-run] chat_id=%d 异常退出，错误=%v，设置状态为error", chatID, err)
			_ = common.DbMain.TaskWorkflowChatMarkError(chatID)
		}
	} else {
		gstool.FmtPrintlnLogTime("[codex-run] chat_id=%d 正常完成，设置状态为completed", chatID)
		_ = common.DbMain.TaskWorkflowChatMarkCompleted(chatID)
	}

	durationMs := time.Since(startTime).Milliseconds()
	if durationMs < 0 {
		durationMs = 0
	}
	if turnCount <= 0 {
		turnCount = 1
	}
	modelUsage := map[string]any{}
	if codexCfg.Model != `` {
		modelUsage[codexCfg.Model] = map[string]any{
			`inputTokens`:          cast.ToInt64(lastTurnUsage[`input_tokens`]),
			`outputTokens`:         cast.ToInt64(lastTurnUsage[`output_tokens`]),
			`cacheReadInputTokens`: cast.ToInt64(lastTurnUsage[`cache_read_input_tokens`]),
		}
	}
	resultJSON, _ := json.Marshal(map[string]any{
		`type`:        `result`,
		`subtype`:     `completed`,
		`duration_ms`: durationMs,
		`num_turns`:   turnCount,
		`usage`:       lastTurnUsage,
		`modelUsage`:  modelUsage,
		`is_error`:    err != nil,
		`result`:      lastAgentMessageText,
	})
	sendLine(string(resultJSON))
	finalStatus := common.TaskWorkflowChatStatusCompleted
	if err != nil {
		select {
		case <-stopped:
			finalStatus = common.TaskWorkflowChatStatusInterrupted
		default:
			finalStatus = common.TaskWorkflowChatStatusError
		}
	}
	gstool.FmtPrintlnLogTime("[codex-run] chat_id=%d 最终状态=%s，发送completed事件", chatID, finalStatus)
	sendLine(buildChatCompletedEvent(chatID, finalStatus))
	taskWorkflowBroadcastChatStatus(chatID)
	_ = lastAgentMessageText
}

// buildChatCompletedEvent 构造对话终态 SSE 事件，携带最终状态供前端背景列表即时更新。
// buildChatCompletedEvent builds the terminal SSE payload with final status so background history rows can update immediately.
func buildChatCompletedEvent(chatID int64, status string) string {
	chatInfo, _ := common.DbMain.TaskWorkflowChatInfo(chatID)
	payload, _ := json.Marshal(map[string]any{
		`type`:    `chat`,
		`subtype`: `completed`,
		`chat_id`: chatID,
		`status`:  strings.TrimSpace(status),
		`is_read`: cast.ToInt(chatInfo[`is_read`]) == 1,
	})
	return string(payload)
}

const (
	// taskWorkflowChatStopReasonUserStop 标识用户主动点击"停止"结束对话。 // Marks that the chat was terminated explicitly by the user.
	taskWorkflowChatStopReasonUserStop = `user_stop`
)

// buildChatStopEvent 构造终止原因事件，写入 raw_output 供历史列表和通知逻辑复用。 // Builds a lightweight terminal-reason event persisted in raw_output for history rows and notification rules.
func buildChatStopEvent(chatID int64, reason string) string {
	payload, _ := json.Marshal(map[string]any{
		`type`:        `chat`,
		`subtype`:     `stopped`,
		`chat_id`:     chatID,
		`stop_reason`: strings.TrimSpace(reason),
	})
	return string(payload)
}

// taskWorkflowExtractTerminalReason 从原始输出中提取终止原因代码和文案。 // Extracts the terminal reason code and readable label from persisted raw output.
func taskWorkflowExtractTerminalReason(status, rawOutput string) (string, string) {
	status = strings.TrimSpace(status)
	rawOutput = strings.TrimSpace(rawOutput)
	if rawOutput != `` {
		lines := strings.Split(rawOutput, "\n")
		for i := len(lines) - 1; i >= 0; i-- {
			line := strings.TrimSpace(lines[i])
			if line == `` || !strings.HasPrefix(line, `{`) {
				continue
			}
			var item map[string]any
			if err := json.Unmarshal([]byte(line), &item); err != nil {
				continue
			}
			itemType := strings.TrimSpace(cast.ToString(item[`type`]))
			switch itemType {
			case `result`:
				reason := strings.TrimSpace(cast.ToString(item[`stop_reason`]))
				if reason != `` {
					return reason, taskWorkflowStopReasonLabel(reason)
				}
			case `chat`:
				if strings.TrimSpace(cast.ToString(item[`subtype`])) == `stopped` {
					reason := strings.TrimSpace(cast.ToString(item[`stop_reason`]))
					if reason != `` {
						return reason, taskWorkflowStopReasonLabel(reason)
					}
				}
			case `error`:
				errText := strings.TrimSpace(cast.ToString(item[`text`]))
				if errText != `` {
					return `error`, errText
				}
			}
		}
	}
	if status == `interrupted` {
		return taskWorkflowChatStopReasonUserStop, taskWorkflowStopReasonLabel(taskWorkflowChatStopReasonUserStop)
	}
	return ``, ``
}

// taskWorkflowStopReasonLabel 将终止原因代码转换为用户可读文案。 // Maps terminal reason codes into readable labels for the UI and notifications.
func taskWorkflowStopReasonLabel(reason string) string {
	switch strings.TrimSpace(reason) {
	case `end_turn`:
		return `正常结束`
	case `stop_sequence`:
		return `停止序列`
	case `max_tokens`:
		return `达到上限`
	case `tool_use`:
		return `工具调用`
	case taskWorkflowChatStopReasonUserStop:
		return `用户主动终止`
	case `error`:
		return `异常终止`
	default:
		return strings.TrimSpace(reason)
	}
}

// taskWorkflowAutoCompleteNode 自动将指定节点标记为已完成。
func taskWorkflowAutoCompleteNode(workflowInfo map[string]any, nodeKey string) {
	workflowID := cast.ToInt(workflowInfo[`id`])
	if workflowID <= 0 {
		return
	}
	nodeStatuses := make(map[string]string)
	if raw := strings.TrimSpace(cast.ToString(workflowInfo[`node_statuses`])); raw != `` {
		_ = json.Unmarshal([]byte(raw), &nodeStatuses)
	}
	nodeStatuses[nodeKey] = `completed`
	if data, err := json.Marshal(nodeStatuses); err == nil {
		_ = common.DbMain.TaskWorkflowUpdateNodeStatuses(workflowID, string(data))
	}
}

// taskWorkflowBroadcastChatStatus 在 chat 状态变更后通知对应工作流页面刷新计数。
func taskWorkflowBroadcastChatStatus(chatID int64) {
	chatInfo, err := common.DbMain.TaskWorkflowChatInfo(chatID)
	if err != nil || len(chatInfo) == 0 {
		return
	}
	taskWorkflowBroadcastUnreadChanged(chatID)
	// 仅工作流来源的 chat 需要广播到工作流页面。 // Only workflow-origin chats should refresh workflow-side counters.
	if cast.ToString(chatInfo[`from_type`]) != common.AgentChatSourceTypeWorkflow {
		return
	}
	workflowID := cast.ToInt(chatInfo[`from_id`])
	if workflowID <= 0 {
		return
	}
	taskWorkflowBroadcastWorkflowDetail(workflowID, `chat_status_change`, chatID)
}

// taskWorkflowBroadcastUnreadChanged 在普通 SSE 通道广播未读数量变化，供工作流页与 Agent CLI 页实时更新红点。
func taskWorkflowBroadcastUnreadChanged(chatID int64) {
	chatInfo, err := common.DbMain.TaskWorkflowChatInfo(chatID)
	if err != nil || len(chatInfo) == 0 {
		return
	}
	fromType := cast.ToString(chatInfo[`from_type`])
	fromID := cast.ToInt(chatInfo[`from_id`])
	agentCliID := cast.ToInt(chatInfo[`agent_cli_id`])
	agentCliUnread := 0
	if fromType == common.AgentChatSourceTypeAgentCli && agentCliID > 0 {
		agentCliUnread, _ = common.DbMain.AgentChatUnreadCountByAgentCli(agentCliID)
	}
	agentCliHomeMsg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseAgentCliUnreadHome,
		Data: map[string]any{
			`type`:             `agent_cli_unread_home`,
			`chat_id`:          chatID,
			`from_type`:        fromType,
			`from_id`:          fromID,
			`agent_cli_id`:     agentCliID,
			`agent_cli_unread`: agentCliUnread,
		},
		Type: p_define.SseContentTypeMsg,
	})
	agentCliGlobalMsg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: define.SseAgentCliUnreadGlobal,
		Data: map[string]any{
			`type`:             `agent_cli_unread_global`,
			`chat_id`:          chatID,
			`from_type`:        fromType,
			`from_id`:          fromID,
			`agent_cli_id`:     agentCliID,
			`agent_cli_unread`: agentCliUnread,
		},
		Type: p_define.SseContentTypeMsg,
	})
	workflowUnreadData, workflowUnreadErr := buildWorkflowUnreadSnapshotData()
	for _, item := range gsgin.SseStatus() {
		clientID := strings.TrimSpace(strings.TrimPrefix(item, apiDataChangeSseStatusPrefix))
		if clientID == `` || clientID == item {
			continue
		}
		sse := gsgin.SseGetByClientId(clientID)
		if sse == nil {
			continue
		}
		_ = sse.SendToChan(agentCliHomeMsg)
		_ = sse.SendToChan(agentCliGlobalMsg)
		if workflowUnreadErr != nil {
			continue
		}
		for _, distributeID := range []string{
			define.SseWorkflowUnreadHomeMenu,
			define.SseWorkflowUnreadHomeTask,
			define.SseWorkflowUnreadDetail,
		} {
			_ = sse.SendToChan(gstool.JsonEncode(p_define.SseData{
				SseDistributeId: distributeID,
				Data:            workflowUnreadData,
				Type:            p_define.SseContentTypeMsg,
			}))
		}
	}
}

// extractAssistantText 从 assistant 消息中提取文本内容。
// Claude Code stream-json 的 assistant 消息结构为：{"type":"assistant","message":{"content":[{"type":"text","text":"..."}]}}
// 所以 content 数组要从 data["message"]["content"] 取，而不是 data["content"]。
func extractAssistantText(data map[string]any) string {
	msgMap, ok := data["message"].(map[string]any)
	if !ok {
		return ""
	}
	contentRaw, ok := msgMap["content"]
	if !ok {
		return ""
	}
	contentArr, ok := contentRaw.([]any)
	if !ok {
		return ""
	}
	var texts []string
	for _, block := range contentArr {
		blockMap, ok := block.(map[string]any)
		if !ok {
			continue
		}
		if cast.ToString(blockMap["type"]) == "text" {
			if t := cast.ToString(blockMap["text"]); t != "" {
				texts = append(texts, t)
			}
		}
	}
	return strings.Join(texts, "\n")
}

// taskWorkflowFileChangesSummary 获取指定本地目录的文件变更汇总。
// 当 parentBranch 非空时，调用 show_branch_diff.py 获取文件列表（含 Committed/Staged/Modified/Untracked 状态）；
// 否则降级为 git status --short 并映射为旧分类。
func taskWorkflowFileChangesSummary(localDir, parentBranch string) map[string]any {
	result := map[string]any{
		`local_dir`:   localDir,
		`error`:       ``,
		`summary`:     map[string]int{`committed`: 0, `staged`: 0, `modified`: 0, `untracked`: 0, `total`: 0, `additions`: 0, `deletions`: 0},
		`files`:       []map[string]any{},
		`has_changes`: false,
	}

	info, statErr := os.Stat(localDir)
	if statErr != nil || !info.IsDir() {
		result[`error`] = `目录不存在`
		return result
	}

	if parentBranch != `` {
		return taskWorkflowFileChangesFromBranchDiff(localDir, parentBranch, result)
	}

	// 无 parentBranch 时降级使用 git status --short
	cmd := exec.Command(`git`, `-C`, localDir, `status`, `--short`)
	output, runErr := cmd.CombinedOutput()
	if runErr != nil {
		msg := strings.TrimSpace(string(output))
		if msg == `` {
			msg = runErr.Error()
		}
		result[`error`] = msg
		return result
	}

	trimmed := strings.TrimSpace(string(output))
	if trimmed == `` {
		return result
	}

	summary := map[string]int{`committed`: 0, `staged`: 0, `modified`: 0, `untracked`: 0, `total`: 0, `additions`: 0, `deletions`: 0}
	files := make([]map[string]any, 0)

	lines := strings.Split(trimmed, "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == `` {
			continue
		}
		statusCode, filePath, cat := categorizeGitStatusLineV2(line)
		// DEBUG: 打印原始行和解析结果，排查路径异常
		fmt.Printf("[DEBUG git-status] raw=%q hex=%x => code=%q path=%q cat=%q\n", line, []byte(line), statusCode, filePath, cat)
		summary[cat]++
		summary[`total`]++
		files = append(files, map[string]any{
			`path`:        filePath,
			`type`:        cat,
			`status_code`: statusCode,
			`additions`:   0,
			`deletions`:   0,
		})
	}

	// 获取工作区增删行数统计（git diff --numstat HEAD）
	gitNumstat := map[string][2]int{}
	for _, cmdArgs := range [][]string{
		{"diff", "--numstat", "HEAD", "--", "."},
		{"diff", "--numstat", "--cached", "--", "."},
	} {
		numstatCmd := exec.Command(`git`, append([]string{`-C`, localDir}, cmdArgs...)...)
		numstatOut, numstatErr := numstatCmd.CombinedOutput()
		if numstatErr == nil {
			for _, numLine := range strings.Split(strings.TrimSpace(string(numstatOut)), "\n") {
				if strings.TrimSpace(numLine) == `` {
					continue
				}
				parts := strings.SplitN(numLine, "\t", 3)
				if len(parts) >= 3 {
					numPath := parts[2]
					addStr, delStr := parts[0], parts[1]
					if addStr == "-" || delStr == "-" {
						// 二进制文件
						gitNumstat[numPath] = [2]int{1, 1}
					} else {
						add, _ := strconv.Atoi(addStr)
						del, _ := strconv.Atoi(delStr)
						if prev, ok := gitNumstat[numPath]; ok {
							gitNumstat[numPath] = [2]int{prev[0] + add, prev[1] + del}
						} else {
							gitNumstat[numPath] = [2]int{add, del}
						}
					}
				}
			}
		}
	}

	// 将 numstat 合并到 files 中
	for i, f := range files {
		fp := cast.ToString(f[`path`])
		if stats, ok := gitNumstat[fp]; ok {
			f[`additions`] = stats[0]
			f[`deletions`] = stats[1]
			summary[`additions`] += stats[0]
			summary[`deletions`] += stats[1]
		} else if cat := cast.ToString(f[`type`]); cat == `untracked` {
			// 未跟踪文件：统计行数
			fullPath := filepath.Join(localDir, fp)
			if content, readErr := os.ReadFile(fullPath); readErr == nil {
				lineCount := len(bytes.Split(content, []byte{'\n'}))
				if lineCount > 0 && len(bytes.TrimSpace(content)) == 0 {
					lineCount = 0
				}
				f[`additions`] = lineCount
				f[`deletions`] = 0
				summary[`additions`] += lineCount
			} else {
				// 二进制等无法读取的文件
				f[`additions`] = 1
				f[`deletions`] = 1
				summary[`additions`] += 1
				summary[`deletions`] += 1
			}
		} else {
			// 其他无法获取 numstat 的文件（可能已删除等）
			f[`additions`] = 1
			f[`deletions`] = 1
			summary[`additions`] += 1
			summary[`deletions`] += 1
		}
		files[i] = f
	}

	result[`summary`] = summary
	result[`files`] = files
	result[`has_changes`] = summary[`total`] > 0
	return result
}

// categorizeGitStatusLineV2 解析 git status --short 的单行输出，返回状态码、文件路径和分类。
// git status --short 输出格式: "XY path" (X=索引状态, Y=工作区状态, 空格, 文件路径)
// 分类映射：?? → untracked, A → staged, M(索引) → staged, M(工作区) → modified, 其他 → modified
func categorizeGitStatusLineV2(line string) (statusCode string, filePath string, category string) {
	if len(line) < 3 {
		return `M`, strings.TrimSpace(line), `modified`
	}

	// 保留原始的2字符状态码，不做 TrimSpace，因为状态码可能包含前导/后置空格
	// 如 " M" 表示索引无变化、工作区已修改；"M " 表示索引已暂存、工作区无变化
	code := line[:2]

	// 提取路径: 跳过状态码(2字符)+分隔空格(1字符)
	// 标准格式索引2为空格，但防御性地处理空格缺失的异常情况
	rest := ``
	if len(line) >= 4 && line[2] == ' ' {
		rest = strings.TrimSpace(line[3:])
	} else {
		// 边缘情况：状态码和路径之间缺少空格，搜索第一个空格/制表符
		for i := 2; i < len(line); i++ {
			if line[i] == ' ' || line[i] == '\t' {
				rest = strings.TrimSpace(line[i+1:])
				break
			}
		}
		if rest == `` {
			rest = strings.TrimSpace(line[2:])
		}
	}

	// 处理重命名 "R  old -> new"
	if idx := strings.Index(rest, ` -> `); idx >= 0 {
		rest = strings.TrimSpace(rest[idx+4:])
	}

	filePath = rest

	switch {
	case code == `??`:
		return `U`, filePath, `untracked`
	case code == `A `, code == `AM`, code == `A?`:
		return `S`, filePath, `staged`
	case code == `M `:
		return `S`, filePath, `staged`
	case code == ` M`, code == `MM`:
		return `M`, filePath, `modified`
	case code == `D `, code == ` D`, code == `DM`, code == `MD`:
		return `M`, filePath, `modified`
	case code == `R `, code == ` R`:
		return `S`, filePath, `staged`
	default:
		return `M`, filePath, `modified`
	}
}

// getPythonCommand 返回可用的 Python 命令路径。
// 依次尝试 python3、python，以及 Windows 常见安装路径。结果会缓存。
func getPythonCommand() string {
	if cachedPythonCmd != `` {
		return cachedPythonCmd
	}
	candidates := []string{`python3`, `python`}
	// Windows 常见安装路径
	if runtime.GOOS == `windows` {
		localAppData := os.Getenv(`LOCALAPPDATA`)
		if localAppData != `` {
			candidates = append(candidates,
				filepath.Join(localAppData, `Programs`, `Python`, `Python313`, `python.exe`),
				filepath.Join(localAppData, `Programs`, `Python`, `Python312`, `python.exe`),
				filepath.Join(localAppData, `Programs`, `Python`, `Python311`, `python.exe`),
				filepath.Join(localAppData, `Programs`, `Python`, `Python310`, `python.exe`),
				filepath.Join(localAppData, `Programs`, `Python`, `Python39`, `python.exe`),
			)
		}
		programFiles := os.Getenv(`ProgramFiles`)
		if programFiles != `` {
			candidates = append(candidates,
				filepath.Join(programFiles, `Python313`, `python.exe`),
				filepath.Join(programFiles, `Python312`, `python.exe`),
			)
		}
	}
	for _, cmd := range candidates {
		// 跳过 Windows Store 占位符（0 字节）
		if fi, err := os.Stat(cmd); err == nil && fi.Size() == 0 {
			continue
		}
		if p, err := exec.LookPath(cmd); err == nil {
			// 再次检查文件大小（LookPath 返回的可能是占位符）
			if fi, err := os.Stat(p); err == nil && fi.Size() > 0 {
				cachedPythonCmd = p
				return p
			}
			continue
		}
		// 对于绝对路径，直接检查文件是否存在且非空
		if filepath.IsAbs(cmd) {
			if fi, err := os.Stat(cmd); err == nil && fi.Size() > 0 {
				cachedPythonCmd = cmd
				return cmd
			}
		}
	}
	cachedPythonCmd = `python`
	return `python`
}

// taskWorkflowFileChangesFromBranchDiff 调用 show_branch_diff.py 获取文件变更列表。
// 脚本输出格式：file_path\t[Status1,Status2]
// 状态映射：Committed→C, Staged→S, Modified→M, Untracked→U
func taskWorkflowFileChangesFromBranchDiff(localDir, parentBranch string, result map[string]any) map[string]any {
	workspaceRoot := os.Getenv(`WORKSPACE`)
	if workspaceRoot == `` {
		workspaceRoot = getDefaultWorkspaceRoot()
	}
	scriptPath := filepath.Join(workspaceRoot, `skills`, `dtool-git`, `scripts`, `show_branch_diff.py`)

	cmd := exec.Command(getPythonCommand(), scriptPath, parentBranch)
	cmd.Dir = localDir

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == `` {
			errMsg = err.Error()
		}
		result[`error`] = errMsg
		return result
	}

	summary := map[string]int{`committed`: 0, `staged`: 0, `modified`: 0, `untracked`: 0, `total`: 0, `additions`: 0, `deletions`: 0}
	files := make([]map[string]any, 0)

	output := strings.TrimSpace(stdout.String())
	if output == `` {
		result[`summary`] = summary
		result[`files`] = files
		return result
	}

	// 状态全称 → 首字母缩写
	statusAbbr := map[string]string{
		`Committed`: `C`,
		`Staged`:    `S`,
		`Modified`:  `M`,
		`Untracked`: `U`,
	}
	// 状态全称 → 分类
	statusCategory := map[string]string{
		`Committed`: `committed`,
		`Staged`:    `staged`,
		`Modified`:  `modified`,
		`Untracked`: `untracked`,
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) == `` {
			continue
		}
		// 格式: file_path\t[Status1,Status2]\tadditions\tdeletions
		tabIdx := strings.Index(line, "\t")
		if tabIdx < 0 {
			continue
		}
		filePath := strings.TrimSpace(line[:tabIdx])
		rest := strings.TrimSpace(line[tabIdx+1:])

		// 解析 [Status1,Status2]
		statusPart := rest
		var additions, deletions int

		// 尝试从 rest 中提取 additions 和 deletions（格式: [Status]\tadditions\tdeletions）
		if tabIdx2 := strings.Index(rest, "\t"); tabIdx2 >= 0 {
			statusPart = strings.TrimSpace(rest[:tabIdx2])
			numPart := strings.TrimSpace(rest[tabIdx2+1:])
			// numPart 格式: "additions\tdeletions"
			numFields := strings.SplitN(numPart, "\t", 2)
			if len(numFields) >= 2 {
				additions, _ = strconv.Atoi(strings.TrimSpace(numFields[0]))
				deletions, _ = strconv.Atoi(strings.TrimSpace(numFields[1]))
			} else if len(numFields) == 1 {
				additions, _ = strconv.Atoi(strings.TrimSpace(numFields[0]))
			}
		}

		var abbrs []string
		var primaryCat string
		statusPart = strings.TrimPrefix(statusPart, `[`)
		statusPart = strings.TrimSuffix(statusPart, `]`)
		for _, s := range strings.Split(statusPart, `,`) {
			s = strings.TrimSpace(s)
			if abbr, ok := statusAbbr[s]; ok {
				abbrs = append(abbrs, abbr)
			}
			if cat, ok := statusCategory[s]; ok && primaryCat == `` {
				primaryCat = cat
			}
		}

		if primaryCat == `` {
			primaryCat = `modified`
		}
		statusCode := strings.Join(abbrs, `,`)
		if statusCode == `` {
			statusCode = `M`
		}

		summary[primaryCat]++
		summary[`total`]++
		summary[`additions`] += additions
		summary[`deletions`] += deletions
		files = append(files, map[string]any{
			`path`:        filePath,
			`type`:        primaryCat,
			`status_code`: statusCode,
			`additions`:   additions,
			`deletions`:   deletions,
		})
	}

	result[`summary`] = summary
	result[`files`] = files
	result[`has_changes`] = summary[`total`] > 0
	return result
}

// TaskWorkflowFileChangesSummary 获取文件变更汇总（按本地目录批量）。
func TaskWorkflowFileChangesSummary(c *gin.Context) {
	var req _struct.TaskWorkflowFileChangesSummaryRequest
	_ = gsgin.GinPostBody(c, &req)

	if len(req.Items) == 0 {
		gsgin.GinResponseError(c, `items 不能为空`, nil)
		return
	}

	result := make(map[string]map[string]any, len(req.Items))
	for _, item := range req.Items {
		dir := strings.TrimSpace(item.LocalDir)
		if dir == `` {
			continue
		}
		branch := strings.TrimSpace(item.ParentBranch)
		result[dir] = taskWorkflowFileChangesSummary(dir, branch)
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`dirs`: result,
	})
}

// TaskWorkflowFileChangesDetail 获取文件变更详情（调用 show_branch_diff.py 获取文件列表）。
func TaskWorkflowFileChangesDetail(c *gin.Context) {
	var req _struct.TaskWorkflowFileChangesDetailRequest
	_ = gsgin.GinPostBody(c, &req)

	req.LocalDir = strings.TrimSpace(req.LocalDir)
	if req.LocalDir == `` {
		gsgin.GinResponseError(c, `local_dir 不能为空`, nil)
		return
	}

	req.ParentBranch = strings.TrimSpace(req.ParentBranch)

	result := taskWorkflowFileChangesSummary(req.LocalDir, req.ParentBranch)
	if result[`error`] != nil && result[`error`] != `` {
		gsgin.GinResponseError(c, cast.ToString(result[`error`]), result)
		return
	}

	gsgin.GinResponseSuccess(c, ``, result)
}

// TaskWorkflowFileChangesFileDiff 获取单个文件的 diff（调用 show_file_diff.py）。
func TaskWorkflowFileChangesFileDiff(c *gin.Context) {
	var req _struct.TaskWorkflowFileChangesFileDiffRequest
	_ = gsgin.GinPostBody(c, &req)

	req.LocalDir = strings.TrimSpace(req.LocalDir)
	req.ParentBranch = strings.TrimSpace(req.ParentBranch)
	req.FilePath = strings.TrimSpace(req.FilePath)

	if req.LocalDir == `` {
		gsgin.GinResponseError(c, `local_dir 不能为空`, nil)
		return
	}
	if req.FilePath == `` {
		gsgin.GinResponseError(c, `file_path 不能为空`, nil)
		return
	}

	// parentBranch 为空时使用工作区模式（对比 HEAD）
	branchArg := req.ParentBranch
	if branchArg == `` {
		branchArg = `_workspace_`
	}

	workspaceRoot := os.Getenv(`WORKSPACE`)
	if workspaceRoot == `` {
		workspaceRoot = getDefaultWorkspaceRoot()
	}
	scriptPath := filepath.Join(workspaceRoot, `skills`, `dtool-git`, `scripts`, `show_file_diff.py`)

	cmd := exec.Command(getPythonCommand(), scriptPath, branchArg, req.FilePath)
	cmd.Dir = req.LocalDir

	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		errMsg := strings.TrimSpace(stderr.String())
		if errMsg == `` {
			errMsg = err.Error()
		}
		gsgin.GinResponseError(c, `获取文件 diff 失败: `+errMsg, nil)
		return
	}

	// 解析 Python 脚本输出的 JSON（可包含 diff/old_content/new_content 或 is_binary/is_image）
	type fileDiffResult struct {
		Diff       string `json:"diff"`
		OldContent string `json:"old_content"`
		NewContent string `json:"new_content"`
		IsBinary   bool   `json:"is_binary"`
		FileType   string `json:"file_type"`
		OldSize    int64  `json:"old_size"`
		NewSize    int64  `json:"new_size"`
		IsImage    bool   `json:"is_image"`
		ImageType  string `json:"image_type"`
		OldImage   string `json:"old_image"`
		NewImage   string `json:"new_image"`
	}
	var diffResult fileDiffResult
	if jsonErr := json.Unmarshal([]byte(stdout.String()), &diffResult); jsonErr == nil {
		respData := map[string]any{
			`file_path`: req.FilePath,
		}
		// 二进制文件响应
		if diffResult.IsBinary {
			respData[`is_binary`] = true
			respData[`file_type`] = diffResult.FileType
			respData[`old_size`] = diffResult.OldSize
			respData[`new_size`] = diffResult.NewSize
		} else if diffResult.IsImage {
			// 图片文件响应
			respData[`is_image`] = true
			respData[`image_type`] = diffResult.ImageType
			respData[`old_image`] = diffResult.OldImage
			respData[`new_image`] = diffResult.NewImage
		} else {
			// 文本文件响应
			respData[`diff`] = diffResult.Diff
			respData[`old_content`] = diffResult.OldContent
			respData[`new_content`] = diffResult.NewContent
		}
		gsgin.GinResponseSuccess(c, ``, respData)
		return
	}

	// 降级：若非 JSON 则当作纯 diff 文本
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`diff`:        stdout.String(),
		`old_content`: ``,
		`new_content`: ``,
		`file_path`:   req.FilePath,
	})
}

// TaskWorkflowOpenInEditor 在指定的 IDE 中打开工作目录。
func TaskWorkflowOpenInEditor(c *gin.Context) {
	var req _struct.TaskWorkflowOpenInEditorRequest
	if err := gsgin.GinPostBody(c, &req); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	localDir := strings.TrimSpace(req.LocalDir)
	if localDir == `` {
		gsgin.GinResponseError(c, `工作目录不能为空`, nil)
		return
	}

	// 校验目录存在性
	if info, statErr := os.Stat(localDir); statErr != nil || !info.IsDir() {
		gsgin.GinResponseError(c, fmt.Sprintf(`目录不存在: %s`, localDir), nil)
		return
	}

	// 根据 IDE 类型获取命令
	command := getEditorLaunchCommand(req.EditorType)
	if command == `` {
		gsgin.GinResponseError(c, fmt.Sprintf(`不支持的编辑器类型: %s`, req.EditorType), nil)
		return
	}

	cmd := exec.Command(command, localDir)
	// 不等待完成，直接启动进程
	if err := cmd.Start(); err != nil {
		gsgin.GinResponseError(c, fmt.Sprintf(`启动 %s 失败: %s`, req.EditorType, err.Error()), nil)
		return
	}

	// 释放资源，不等待进程结束
	go cmd.Process.Release()

	gsgin.GinResponseSuccess(c, fmt.Sprintf(`已在 %s 中打开目录: %s`, req.EditorType, localDir), nil)
}

// getEditorLaunchCommand 返回不同 IDE 的启动命令。
func getEditorLaunchCommand(editorType string) string {
	switch editorType {
	case `vscode`:
		return `code`
	case `cursor`:
		return `cursor`
	case `goland`:
		if runtime.GOOS == `windows` {
			return `goland64.exe`
		}
		return `goland`
	case `phpstorm`:
		if runtime.GOOS == `windows` {
			return `phpstorm64.exe`
		}
		return `phpstorm`
	default:
		return ``
	}
}

// getDefaultWorkspaceRoot 获取默认的工作空间根目录。
func getDefaultWorkspaceRoot() string {
	if wd, err := os.Getwd(); err == nil {
		return wd
	}
	return `.`
}
