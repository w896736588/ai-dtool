package controller

import (
	"bufio"
	"dev_tool/internal/app/dtool/api"
	"os"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	"dev_tool/internal/app/dtool/define"
	_struct "dev_tool/internal/app/dtool/struct"
	"dev_tool/internal/pkg/p_define"
	"encoding/json"
	"fmt"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"gitee.com/Sxiaobai/gs/v2/gsgin"
	"gitee.com/Sxiaobai/gs/v2/gstool"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
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
	fragmentID := strings.TrimSpace(cast.ToString(workflowInfo[`dev_plan_fragment_id`]))
	if fragmentID == `` {
		gsgin.GinResponseError(c, `开发执行文档未初始化`, nil)
		return
	}
	fragmentInfo, err := memoryDB.MemoryFragmentInfo(fragmentID)
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
	)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
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
	if strings.TrimSpace(request.Label) == `` {
		gsgin.GinResponseError(c, `label不能为空`, nil)
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
	resultFile, err := dispatchScrapeTaskAndAwait(request.SmartLinkID, request.Label, request.JumpURL, request.CssSelector, request.WaitSeconds)
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
		`label`:              request.Label,
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
	requirementFragmentID := strings.TrimSpace(cast.ToString(workflowInfo[`requirement_fragment_id`]))
	if requirementFragmentID == `` {
		gsgin.GinResponseError(c, `需求文档未绑定`, nil)
		return nil, nil, nil, false
	}
	requirementFragment, err := memoryDB.MemoryFragmentInfo(requirementFragmentID)
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
	fragmentID := strings.TrimSpace(cast.ToString(workflowInfo[`dev_plan_fragment_id`]))
	if fragmentID != `` {
		return memoryDB.MemoryFragmentInfo(fragmentID)
	}
	fragmentTitle := strings.TrimSpace(cast.ToString(homeTaskInfo[`name`])) + taskWorkflowDevPlanDefaultTitleSuffix
	if strings.TrimSpace(fragmentTitle) == taskWorkflowDevPlanDefaultTitleSuffix {
		fragmentTitle = `开发执行文档`
	}
	fragmentInfo, err := memoryDB.MemoryFragmentSave(0, fragmentTitle, taskWorkflowDevPlanDefaultTemplate, []string{taskWorkflowDevPlanDefaultTag})
	if err != nil {
		return nil, err
	}
	component.MemoryRuntime.ScheduleSync()
	workflowID := cast.ToInt(workflowInfo[`id`])
	fragmentFileID := strings.TrimSpace(cast.ToString(fragmentInfo[`file_id`]))
	if fragmentFileID == `` {
		return nil, gstool.Error(`开发执行片段创建失败`)
	}
	if err = common.DbMain.TaskWorkflowBindDevPlanFragment(workflowID, fragmentFileID); err != nil {
		return nil, err
	}
	workflowInfo[`dev_plan_fragment_id`] = fragmentFileID
	workflowInfo[`status`] = `dev_plan_ready`
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
	homeTaskInfo, err := common.DbMain.HomeTaskRow(cast.ToInt(workflowInfo[`home_task_id`]))
	if err != nil {
		return nil, err
	}
	// 新创建的工作流提示词为空时，从配置模板初始化。
	workflowID := cast.ToInt(workflowInfo[`id`])
	if workflowID > 0 && strings.TrimSpace(cast.ToString(workflowInfo[`prompt_requirement`])) == `` && strings.TrimSpace(cast.ToString(workflowInfo[`prompt_api_dev`])) == `` && strings.TrimSpace(cast.ToString(workflowInfo[`prompt_api_test`])) == `` && strings.TrimSpace(cast.ToString(workflowInfo[`prompt_design`])) == `` && strings.TrimSpace(cast.ToString(workflowInfo[`prompt_plain_text_requirement`])) == `` && strings.TrimSpace(cast.ToString(workflowInfo[`prompt_design_plan_requirement`])) == `` && strings.TrimSpace(cast.ToString(workflowInfo[`prompt_browser_test`])) == `` && strings.TrimSpace(cast.ToString(workflowInfo[`prompt_code_review`])) == `` {
		prompts := resolveTaskWorkflowPrompts(c, homeTaskInfo, workflowInfo)
		_ = common.DbMain.TaskWorkflowUpdatePrompts(workflowID, prompts[`requirement`], prompts[`api_dev`], prompts[`api_test`], prompts[`design`], prompts[`plain_text_requirement`], prompts[`design_plan_requirement`], prompts[`browser_test`], prompts[`code_review`])
		updatedInfo, updateErr := common.DbMain.TaskWorkflowInfo(workflowID)
		if updateErr == nil {
			workflowInfo = updatedInfo
		}
	}
	// 纯文本需求知识片段不存在时自动创建。
	if workflowID > 0 && strings.TrimSpace(cast.ToString(workflowInfo[`plain_text_requirement_fragment_id`])) == `` {
		ensureTaskWorkflowPlainTextReqFragment(workflowInfo, homeTaskInfo)
		updatedInfo, updateErr := common.DbMain.TaskWorkflowInfo(workflowID)
		if updateErr == nil {
			workflowInfo = updatedInfo
		}
	}
	// 需求设计方案知识片段不存在时自动创建。
	if workflowID > 0 && strings.TrimSpace(cast.ToString(workflowInfo[`design_plan_requirement_fragment_id`])) == `` {
		ensureTaskWorkflowDesignPlanReqFragment(workflowInfo, homeTaskInfo)
		updatedInfo, updateErr := common.DbMain.TaskWorkflowInfo(workflowID)
		if updateErr == nil {
			workflowInfo = updatedInfo
		}
	}
	// 接口文档知识片段不存在时自动创建。
	if workflowID > 0 && strings.TrimSpace(cast.ToString(workflowInfo[`api_doc_fragment_id`])) == `` {
		ensureTaskWorkflowApiDocFragment(workflowInfo, homeTaskInfo)
		updatedInfo, updateErr := common.DbMain.TaskWorkflowInfo(workflowID)
		if updateErr == nil {
			workflowInfo = updatedInfo
		}
	}
	return map[string]any{
		`workflow`:                 workflowInfo,
		`home_task`:                homeTaskInfo,
		`requirement_fetch_config`: taskWorkflowRequirementFetchConfig(),
	}, nil
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
	if request.WorkflowID <= 0 {
		gsgin.GinResponseError(c, `workflow_id不能为空`, nil)
		return
	}
	if strings.TrimSpace(request.NodeStatuses) == `` {
		gsgin.GinResponseError(c, `node_statuses不能为空`, nil)
		return
	}
	err := common.DbMain.TaskWorkflowUpdateNodeStatuses(request.WorkflowID, request.NodeStatuses)
	if err != nil {
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
	err := common.DbMain.TaskWorkflowUpdatePrompts(
		request.WorkflowID,
		request.PromptRequirement,
		request.PromptApiDev,
		request.PromptApiTest,
		request.PromptDesign,
		request.PromptPlainTextRequirement,
		request.PromptDesignPlanRequirement,
		request.PromptBrowserTest,
		request.PromptCodeReview,
	)
	if err != nil {
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
	homeTaskInfo, homeTaskErr := common.DbMain.HomeTaskRow(cast.ToInt(workflowInfo[`home_task_id`]))
	if homeTaskErr != nil {
		gsgin.GinResponseError(c, homeTaskErr.Error(), nil)
		return
	}
	prompts := resolveTaskWorkflowPrompts(c, homeTaskInfo, workflowInfo)
	if updateErr := common.DbMain.TaskWorkflowUpdatePrompts(
		request.WorkflowID,
		prompts[`requirement`],
		prompts[`api_dev`],
		prompts[`api_test`],
		prompts[`design`],
		prompts[`plain_text_requirement`],
		prompts[`design_plan_requirement`],
		prompts[`browser_test`],
		prompts[`code_review`],
	); updateErr != nil {
		gsgin.GinResponseError(c, updateErr.Error(), nil)
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

// resolveTaskWorkflowPrompts 从配置模板解析占位符生成工作流提示词。
func resolveTaskWorkflowPrompts(c *gin.Context, homeTaskInfo map[string]any, workflowInfo map[string]any) map[string]string {
	placeholders := buildTaskWorkflowPlaceholderMap(c, homeTaskInfo, workflowInfo)
	promptDev, _ := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigPromptDev)
	promptApiGen, _ := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigPromptApiGen)
	promptApiTest, _ := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigPromptApiTest)
	promptDesign, _ := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigPromptDesign)
	promptPlainTextRequirement, _ := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigPromptPlainTextReq)
	promptDesignPlanRequirement, _ := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigPromptDesignPlanReq)
	promptBrowserTest, _ := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigPromptBrowserTest)
	promptCodeReview, _ := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigPromptCodeReview)
	return map[string]string{
		`requirement`:             taskWorkflowResolvePlaceholders(promptDev, placeholders),
		`api_dev`:                 taskWorkflowResolvePlaceholders(promptApiGen, placeholders),
		`api_test`:                taskWorkflowResolvePlaceholders(promptApiTest, placeholders),
		`design`:                  taskWorkflowResolvePlaceholders(promptDesign, placeholders),
		`plain_text_requirement`:  taskWorkflowResolvePlaceholders(promptPlainTextRequirement, placeholders),
		`design_plan_requirement`: taskWorkflowResolvePlaceholders(promptDesignPlanRequirement, placeholders),
		`browser_test`:            taskWorkflowResolvePlaceholders(promptBrowserTest, placeholders),
		`code_review`:             taskWorkflowResolvePlaceholders(promptCodeReview, placeholders),
	}
}

// buildTaskWorkflowPlaceholderMap 根据任务信息构建占位符替换映射。
func buildTaskWorkflowPlaceholderMap(c *gin.Context, homeTaskInfo map[string]any, workflowInfo map[string]any) map[string]string {
	apiHost := taskWorkflowBuildAPIHost(c)
	devEnvironment, _ := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigDevEnvironment)
	result := map[string]string{
		`{需求文档地址}`:            taskWorkflowBuildShareURL(c, workflowInfo, apiHost),
		`{需求文档纯文本地址}`:         taskWorkflowBuildPlainTextShareURL(c, workflowInfo, apiHost),
		`{需求文档纯文本文件相对地址}`:     taskWorkflowBuildPlainTextFragmentRelativePath(workflowInfo),
		`{需求设计方案文档地址}`:        taskWorkflowBuildDesignPlanShareURL(c, workflowInfo, apiHost),
		`{需求设计方案文件相对地址}`:      taskWorkflowBuildDesignPlanFragmentRelativePath(workflowInfo),
		`{接口开发API地址}`:         apiHost,
		`{接口开发API的token}`:     taskWorkflowBuildAPIToken(c),
		`{开发项目配置}`:            taskWorkflowBuildDevConfigsMarkdown(homeTaskInfo),
		`{开发配置}`:              taskWorkflowBuildDevConfigsMarkdown(homeTaskInfo),
		`{dtool-api地址}`:       filepath.Join(component.EnvClient.RootPath, `skills`, `dtool-api`),
		`{dtool-common地址}`:    filepath.Join(component.EnvClient.RootPath, `skills`, `dtool-common`),
		`{tool-playwright地址}`: filepath.Join(component.EnvClient.RootPath, `skills`, `dtool-playwright`),
		`{自定义网页}`:             taskWorkflowBuildDevConfigsFieldMarkdown(homeTaskInfo, `smart_link`),
		`{网页标签}`:              taskWorkflowBuildDevConfigsFieldMarkdown(homeTaskInfo, `smart_link_label`),
		`{账号}`:                taskWorkflowBuildDevConfigsFieldMarkdown(homeTaskInfo, `smart_link_account`),
	}
	// 先解析开发环境内容中的其他占位符，再将其加入映射。
	for key, value := range result {
		devEnvironment = strings.ReplaceAll(devEnvironment, key, value)
	}
	taskID := cast.ToString(homeTaskInfo[`id`])
	result[`{开发环境}`] = devEnvironment + "\n\n任务ID: " + taskID
	result[`{zcode配置列表}`] = taskWorkflowBuildZcodeMarkdown()
	return result
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
	return token
}

// taskWorkflowBuildShareURL 为需求文档知识片段生成分享链接。
func taskWorkflowBuildShareURL(c *gin.Context, workflowInfo map[string]any, apiHost string) string {
	fragmentID := strings.TrimSpace(cast.ToString(workflowInfo[`requirement_fragment_id`]))
	if fragmentID == `` || component.MemoryRuntime == nil {
		return ``
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``
	}
	// 确认片段存在。
	if _, err := component.MemoryRuntime.DB().MemoryFragmentInfo(fragmentID); err != nil {
		return ``
	}
	shareStore := memoryFragmentShareStoreForRoot(component.MemoryRuntime.Config().Dir)
	share, err := shareStore.Create(fragmentID, time.Now())
	if err != nil {
		return ``
	}
	if apiHost == `` {
		return ``
	}
	shareURL, _ := url.Parse(apiHost)
	shareURL.Path = `/share/` + url.PathEscape(share.Token)
	return shareURL.String()
}

// taskWorkflowBuildPlainTextShareURL 为纯文本需求知识片段生成分享链接。
func taskWorkflowBuildPlainTextShareURL(c *gin.Context, workflowInfo map[string]any, apiHost string) string {
	fragmentID := strings.TrimSpace(cast.ToString(workflowInfo[`plain_text_requirement_fragment_id`]))
	if fragmentID == `` || component.MemoryRuntime == nil {
		return ``
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``
	}
	if _, err := component.MemoryRuntime.DB().MemoryFragmentInfo(fragmentID); err != nil {
		return ``
	}
	shareStore := memoryFragmentShareStoreForRoot(component.MemoryRuntime.Config().Dir)
	share, err := shareStore.Create(fragmentID, time.Now())
	if err != nil {
		return ``
	}
	if apiHost == `` {
		return ``
	}
	shareURL, _ := url.Parse(apiHost)
	shareURL.Path = `/share/` + url.PathEscape(share.Token)
	return shareURL.String()
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
		smartLinkName := taskWorkflowQueryNameByID("tbl_smart_link", cfg.SmartLinkID)
		if cfg.SmartLinkID > 0 {
			sb.WriteString(fmt.Sprintf("- **自定义网页**: %s（ID: %d）\n", smartLinkName, cfg.SmartLinkID))
		}
		if cfg.SmartLinkLabel != "" {
			sb.WriteString(fmt.Sprintf("- **网页标签**: %s\n", cfg.SmartLinkLabel))
		}
		if cfg.SmartLinkAccount != "" {
			sb.WriteString(fmt.Sprintf("- **账号**: %s\n", cfg.SmartLinkAccount))
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
				name := taskWorkflowQueryNameByID("tbl_smart_link", cfg.SmartLinkID)
				val = fmt.Sprintf("%s（ID: %d）", name, cfg.SmartLinkID)
			}
		case `smart_link_label`:
			val = cfg.SmartLinkLabel
		case `smart_link_account`:
			val = cfg.SmartLinkAccount
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

// TaskWorkflowRequirementFetch 执行工作流首节点：抓取 TAPD 需求并直接写入知识片段。
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
	tapdURL := strings.TrimSpace(cast.ToString(homeTaskInfo[`tapd_url`]))
	if tapdURL == `` {
		gsgin.GinResponseError(c, `当前任务未配置TAPD地址`, nil)
		return
	}
	fragmentID := strings.TrimSpace(cast.ToString(workflowInfo[`requirement_fragment_id`]))
	if fragmentID == `` {
		gsgin.GinResponseError(c, `需求知识片段未绑定`, nil)
		return
	}
	existingFragment, err := memoryDB.MemoryFragmentInfo(fragmentID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), `load_config`, `running`, `开始读取 TAPD 抓取配置`)
	if err = common.DbMain.TaskWorkflowMarkRequirementFetchRunning(request.WorkflowID, tapdURL); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	resultMap, err := buildAsyncHomeTaskTapdScrapeResultWithLog(tapdURL, fragmentID, func(step, message string) {
		taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), taskWorkflowNormalizeFetchStep(step), `running`, message)
	})
	if err != nil {
		_ = common.DbMain.TaskWorkflowMarkRequirementFetchFailed(request.WorkflowID, tapdURL, err.Error())
		taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), `error`, `failed`, err.Error())
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	markdown := cast.ToString(resultMap[`markdown`])
	if strings.TrimSpace(markdown) == `` {
		err = fmt.Errorf(`抓取结果为空`)
		_ = common.DbMain.TaskWorkflowMarkRequirementFetchFailed(request.WorkflowID, tapdURL, err.Error())
		taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), `error`, `failed`, err.Error())
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), `save_fragment`, `running`, `开始写入需求知识片段`)
	savedFragment, err := memoryDB.MemoryFragmentSave(
		fragmentID,
		cast.ToString(existingFragment[`title`]),
		markdown,
		cast.ToStringSlice(existingFragment[`tags`]),
	)
	if err != nil {
		_ = common.DbMain.TaskWorkflowMarkRequirementFetchFailed(request.WorkflowID, tapdURL, err.Error())
		taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), `error`, `failed`, err.Error())
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
	broadcastMemoryFragmentUpsert(savedFragment)
	if err = common.DbMain.TaskWorkflowMarkRequirementFetchSuccess(request.WorkflowID, tapdURL); err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	taskWorkflowAutoCompleteNode(workflowInfo, `requirement-fetch`)
	updatedWorkflowInfo, err := common.DbMain.TaskWorkflowInfo(request.WorkflowID)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	taskWorkflowBroadcastStep(cast.ToInt(workflowInfo[`id`]), `done`, `success`, `TAPD 需求抓取完成并已写入知识片段`)
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
	smartLinkIDStr, smartLinkErr := homeTaskConfigValue(define.HomeTaskConfigTapdSmartLinkID)
	label, labelErr := homeTaskConfigValue(define.HomeTaskConfigTapdLinkLabel)
	cssSelector, selectorErr := homeTaskConfigValue(define.HomeTaskConfigTapdCssSelector)
	waitSecondsStr, waitErr := homeTaskConfigValue(define.HomeTaskConfigTapdWaitSeconds)
	config := map[string]any{
		`smart_link_id`: cast.ToInt(smartLinkIDStr),
		`label`:         strings.TrimSpace(label),
		`css_selector`:  strings.TrimSpace(cssSelector),
		`wait_seconds`:  cast.ToInt(waitSecondsStr),
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

// ensureTaskWorkflowPlainTextReqFragment 确保纯文本需求知识片段存在，不存在则自动创建。
func ensureTaskWorkflowPlainTextReqFragment(workflowInfo map[string]any, homeTaskInfo map[string]any) {
	if strings.TrimSpace(cast.ToString(workflowInfo[`plain_text_requirement_fragment_id`])) != `` {
		return
	}
	if component.MemoryRuntime == nil {
		return
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return
	}
	memoryDB := component.MemoryRuntime.DB()
	if memoryDB == nil {
		return
	}
	fragmentTitle := strings.TrimSpace(cast.ToString(homeTaskInfo[`name`])) + `-纯文本需求`
	if strings.TrimSpace(fragmentTitle) == `-纯文本需求` {
		fragmentTitle = `纯文本需求文档`
	}
	fragmentInfo, err := memoryDB.MemoryFragmentSave(0, fragmentTitle, ``, []string{`纯文本需求`})
	if err != nil {
		return
	}
	component.MemoryRuntime.ScheduleSync()
	workflowID := cast.ToInt(workflowInfo[`id`])
	fragmentFileID := strings.TrimSpace(cast.ToString(fragmentInfo[`file_id`]))
	if fragmentFileID == `` {
		return
	}
	if err = common.DbMain.TaskWorkflowBindPlainTextReqFragment(workflowID, fragmentFileID); err != nil {
		return
	}
	workflowInfo[`plain_text_requirement_fragment_id`] = fragmentFileID
}

// taskWorkflowBuildPlainTextFragmentRelativePath 为纯文本需求知识片段构建相对于 fragments/ 目录的相对路径。
func taskWorkflowBuildPlainTextFragmentRelativePath(workflowInfo map[string]any) string {
	fragmentID := strings.TrimSpace(cast.ToString(workflowInfo[`plain_text_requirement_fragment_id`]))
	if fragmentID == `` || component.MemoryRuntime == nil {
		return ``
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``
	}
	info, err := component.MemoryRuntime.DB().MemoryFragmentInfo(fragmentID)
	if err != nil {
		return ``
	}
	filePath := strings.TrimSpace(cast.ToString(info[`file_path`]))
	if filePath == `` {
		return ``
	}
	fragmentsDir := filepath.Join(component.MemoryRuntime.Config().Dir, `fragments`)
	relPath, err := filepath.Rel(fragmentsDir, filePath)
	if err != nil {
		return ``
	}
	relPath = filepath.ToSlash(relPath)
	if relPath == `.` || strings.HasPrefix(relPath, `../`) {
		return ``
	}
	return relPath
}

// ensureTaskWorkflowDesignPlanReqFragment 确保需求设计方案知识片段存在，不存在则自动创建。
func ensureTaskWorkflowDesignPlanReqFragment(workflowInfo map[string]any, homeTaskInfo map[string]any) {
	if strings.TrimSpace(cast.ToString(workflowInfo[`design_plan_requirement_fragment_id`])) != `` {
		return
	}
	if component.MemoryRuntime == nil {
		return
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return
	}
	memoryDB := component.MemoryRuntime.DB()
	if memoryDB == nil {
		return
	}
	fragmentTitle := strings.TrimSpace(cast.ToString(homeTaskInfo[`name`])) + `-需求设计方案`
	if strings.TrimSpace(fragmentTitle) == `-需求设计方案` {
		fragmentTitle = `需求设计方案文档`
	}
	fragmentInfo, err := memoryDB.MemoryFragmentSave(0, fragmentTitle, ``, []string{`需求设计方案`})
	if err != nil {
		return
	}
	component.MemoryRuntime.ScheduleSync()
	workflowID := cast.ToInt(workflowInfo[`id`])
	fragmentFileID := strings.TrimSpace(cast.ToString(fragmentInfo[`file_id`]))
	if fragmentFileID == `` {
		return
	}
	if err = common.DbMain.TaskWorkflowBindDesignPlanReqFragment(workflowID, fragmentFileID); err != nil {
		return
	}
	workflowInfo[`design_plan_requirement_fragment_id`] = fragmentFileID
}

// taskWorkflowBuildDesignPlanShareURL 为需求设计方案知识片段生成分享链接。
func taskWorkflowBuildDesignPlanShareURL(c *gin.Context, workflowInfo map[string]any, apiHost string) string {
	fragmentID := strings.TrimSpace(cast.ToString(workflowInfo[`design_plan_requirement_fragment_id`]))
	if fragmentID == `` || component.MemoryRuntime == nil {
		return ``
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``
	}
	if _, err := component.MemoryRuntime.DB().MemoryFragmentInfo(fragmentID); err != nil {
		return ``
	}
	shareStore := memoryFragmentShareStoreForRoot(component.MemoryRuntime.Config().Dir)
	share, err := shareStore.Create(fragmentID, time.Now())
	if err != nil {
		return ``
	}
	if apiHost == `` {
		return ``
	}
	shareURL, _ := url.Parse(apiHost)
	shareURL.Path = `/share/` + url.PathEscape(share.Token)
	return shareURL.String()
}

// taskWorkflowBuildDesignPlanFragmentRelativePath 为需求设计方案知识片段构建相对于 fragments/ 目录的相对路径。
func taskWorkflowBuildDesignPlanFragmentRelativePath(workflowInfo map[string]any) string {
	fragmentID := strings.TrimSpace(cast.ToString(workflowInfo[`design_plan_requirement_fragment_id`]))
	if fragmentID == `` || component.MemoryRuntime == nil {
		return ``
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return ``
	}
	info, err := component.MemoryRuntime.DB().MemoryFragmentInfo(fragmentID)
	if err != nil {
		return ``
	}
	filePath := strings.TrimSpace(cast.ToString(info[`file_path`]))
	if filePath == `` {
		return ``
	}
	fragmentsDir := filepath.Join(component.MemoryRuntime.Config().Dir, `fragments`)
	relPath, err := filepath.Rel(fragmentsDir, filePath)
	if err != nil {
		return ``
	}
	relPath = filepath.ToSlash(relPath)
	if relPath == `.` || strings.HasPrefix(relPath, `../`) {
		return ``
	}
	return relPath
}

// ensureTaskWorkflowApiDocFragment 确保接口文档知识片段存在，不存在则自动创建。
func ensureTaskWorkflowApiDocFragment(workflowInfo map[string]any, homeTaskInfo map[string]any) {
	if strings.TrimSpace(cast.ToString(workflowInfo[`api_doc_fragment_id`])) != `` {
		return
	}
	if component.MemoryRuntime == nil {
		return
	}
	if err := component.MemoryRuntime.EnsureConfigured(); err != nil {
		return
	}
	memoryDB := component.MemoryRuntime.DB()
	if memoryDB == nil {
		return
	}
	fragmentTitle := strings.TrimSpace(cast.ToString(homeTaskInfo[`name`])) + `-接口文档`
	if strings.TrimSpace(fragmentTitle) == `-接口文档` {
		fragmentTitle = `接口文档`
	}
	fragmentInfo, err := memoryDB.MemoryFragmentSave(0, fragmentTitle, ``, []string{`接口文档`})
	if err != nil {
		return
	}
	component.MemoryRuntime.ScheduleSync()
	workflowID := cast.ToInt(workflowInfo[`id`])
	fragmentFileID := strings.TrimSpace(cast.ToString(fragmentInfo[`file_id`]))
	if fragmentFileID == `` {
		return
	}
	if err = common.DbMain.TaskWorkflowBindApiDocFragment(workflowID, fragmentFileID); err != nil {
		return
	}
	workflowInfo[`api_doc_fragment_id`] = fragmentFileID
}

// TaskWorkflowApiDocReset 重置接口文档，将所有关联文件夹下的接口 Markdown 合并覆盖到知识片段中。
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
	fragmentID := strings.TrimSpace(cast.ToString(workflowInfo[`api_doc_fragment_id`]))
	if fragmentID == `` {
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
	fragmentInfo, err := memoryDB.MemoryFragmentInfo(fragmentID)
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
	_, err = memoryDB.MemoryFragmentSave(fragmentID, fragmentTitle, combinedMD, tags)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	component.MemoryRuntime.ScheduleSync()
	gsgin.GinResponseSuccess(c, `接口文档已重置`, map[string]any{
		`fragment_id`: fragmentID,
	})
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
	nodeStatusesMap, err := common.DbMain.TaskWorkflowBatchNodeStatusesByHomeTaskIDs(request.HomeTaskIDs)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`node_statuses_map`: nodeStatusesMap,
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

// broadcastChatOutput 向所有 SSE 客户端广播对话输出行。
func broadcastChatOutput(chatID int64, line string) {
	distributeID := define.SseTaskWorkflowChatPrefix + cast.ToString(chatID)
	msg := gstool.JsonEncode(p_define.SseData{
		SseDistributeId: distributeID,
		Data: map[string]any{
			`chat_id`: chatID,
			`line`:    line,
			`time`:    time.Now().Unix(),
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

// extractSessionID 从首行 init JSON 中提取 session_id。
func extractSessionID(line string) string {
	var data map[string]any
	if err := json.Unmarshal([]byte(line), &data); err != nil {
		return ``
	}
	return cast.ToString(data[`session_id`])
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
	var localDir string
	for _, cfg := range devConfigs {
		if strings.TrimSpace(cfg.LocalDir) != `` {
			localDir = strings.TrimSpace(cfg.LocalDir)
			break
		}
	}
	if localDir == `` {
		gsgin.GinResponseError(c, `开发配置中没有本地目录，无法执行 claude code`, nil)
		return
	}

	chatID, err := common.DbMain.TaskWorkflowChatCreate(req.WorkflowID, req.Prompt)
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}

	// 查找匹配的 zcode settings 路径
	settingsPath := zcodeLookupSettingsPath(localDir)

	go runClaudeCommand(chatID, localDir, req.Prompt, false, ``, settingsPath)

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`chat_id`: chatID,
	})
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

	chatInfo, err := common.DbMain.TaskWorkflowChatInfo(int64(req.ChatID))
	if err != nil {
		gsgin.GinResponseError(c, err.Error(), nil)
		return
	}
	sessionID := cast.ToString(chatInfo[`session_id`])
	if sessionID == `` {
		gsgin.GinResponseError(c, `对话未找到有效的 session_id`, nil)
		return
	}

	workflowInfo, err := common.DbMain.TaskWorkflowInfo(cast.ToInt(chatInfo[`workflow_id`]))
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
	var localDir string
	for _, cfg := range devConfigs {
		if strings.TrimSpace(cfg.LocalDir) != `` {
			localDir = strings.TrimSpace(cfg.LocalDir)
			break
		}
	}
	if localDir == `` {
		gsgin.GinResponseError(c, `开发配置中没有本地目录`, nil)
		return
	}

	_ = common.DbMain.TaskWorkflowChatMarkRunning(int64(req.ChatID))

	// 查找匹配的 zcode settings 路径
	settingsPath := zcodeLookupSettingsPath(localDir)

	go runClaudeCommand(int64(req.ChatID), localDir, req.Prompt, true, sessionID, settingsPath)

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`chat_id`: req.ChatID,
	})
}

// TaskWorkflowChatList 列出工作流的所有对话。
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
	type chatItem struct {
		ID        int64  `json:"id"`
		SessionID string `json:"session_id"`
		Prompt    string `json:"prompt"`
		Status    string `json:"status"`
		CreatedAt string `json:"created_at"`
	}
	list := make([]chatItem, 0, len(rows))
	for _, row := range rows {
		list = append(list, chatItem{
			ID:        cast.ToInt64(row[`id`]),
			SessionID: cast.ToString(row[`session_id`]),
			Prompt:    cast.ToString(row[`prompt`]),
			Status:    cast.ToString(row[`status`]),
			CreatedAt: cast.ToString(row[`created_at`]),
		})
	}
	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`list`: list,
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

	rawOutput := cast.ToString(info[`raw_output`])
	lines := []string{}
	if rawOutput != `` {
		lines = strings.Split(rawOutput, "\n")
	}

	gsgin.GinResponseSuccess(c, ``, map[string]any{
		`chat_id`:    info[`id`],
		`session_id`: info[`session_id`],
		`prompt`:     info[`prompt`],
		`status`:     info[`status`],
		`created_at`: info[`created_at`],
		`lines`:      lines,
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

// taskWorkflowBuildZcodeMarkdown 构建 zcode 配置列表的 Markdown 文本用于占位符替换。
func taskWorkflowBuildZcodeMarkdown() string {
	rows, err := common.DbMain.ZcodeProjectMappingList()
	if err != nil || len(rows) == 0 {
		return ``
	}
	var sb strings.Builder
	sb.WriteString("| 工作目录 | Settings 配置文件 |\n")
	sb.WriteString("|----------|------------------|\n")
	for _, row := range rows {
		ws := cast.ToString(row[`workspace_path`])
		sp := cast.ToString(row[`settings_path`])
		sb.WriteString(fmt.Sprintf("| %s | %s |\n", ws, sp))
	}
	return sb.String()
}

// runClaudeCommand 后台执行 claude 命令并捕获输出。
func runClaudeCommand(chatID int64, localDir, prompt string, isResume bool, sessionID, settingsPath string) {
	args := []string{}
	if isResume {
		args = append(args, `--resume`, sessionID)
	}
	args = append(args,
		`-p`, prompt,
		`--add-dir`, localDir,
		`--model`, `deepseek-v4-pro[1m]`,
		`--permission-mode`, `bypassPermissions`,
		`--output-format`, `stream-json`,
		`--include-partial-messages`,
		`--verbose`,
	)
	if settingsPath != `` {
		args = append(args, `--settings`, settingsPath)
	}

	// 构建命令展示字符串
	displayParts := []string{`claude`}
	for _, a := range args {
		escaped := strings.ReplaceAll(a, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		displayParts = append(displayParts, `"`+escaped+`"`)
	}
	cmdDisplay := strings.Join(displayParts, ` `)

	cmdLineBytes, _ := json.Marshal(map[string]string{
		`type`:    `system`,
		`subtype`: `command`,
		`text`:    cmdDisplay,
	})
	_ = common.DbMain.TaskWorkflowChatAppendOutput(chatID, string(cmdLineBytes))
	broadcastChatOutput(chatID, string(cmdLineBytes))

	cmd := exec.Command(`claude`, args...)
	cmd.Dir = localDir

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		_ = common.DbMain.TaskWorkflowChatAppendOutput(chatID, fmt.Sprintf(`{"error":"stdout pipe failed: %s"}`, err.Error()))
		_ = common.DbMain.TaskWorkflowChatMarkCompleted(chatID)
		return
	}

	if err := cmd.Start(); err != nil {
		_ = common.DbMain.TaskWorkflowChatAppendOutput(chatID, fmt.Sprintf(`{"error":"start failed: %s"}`, err.Error()))
		_ = common.DbMain.TaskWorkflowChatMarkCompleted(chatID)
		return
	}

	sessionExtracted := false
	scanner := bufio.NewScanner(stdout)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == `` {
			continue
		}

		_ = common.DbMain.TaskWorkflowChatAppendOutput(chatID, line)
		broadcastChatOutput(chatID, line)

		if !sessionExtracted && !isResume {
			if sid := extractSessionID(line); sid != `` {
				_ = common.DbMain.TaskWorkflowChatUpdateSessionID(chatID, sid)
				sessionExtracted = true
			}
		}
	}

	_ = cmd.Wait()
	_ = common.DbMain.TaskWorkflowChatMarkCompleted(chatID)
	broadcastChatOutput(chatID, fmt.Sprintf(`{"type":"chat","subtype":"completed","chat_id":%d}`, chatID))
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
