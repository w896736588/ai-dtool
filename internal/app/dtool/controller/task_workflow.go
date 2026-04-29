package controller

import (
	"dev_tool/internal/app/dtool/api"
	"dev_tool/internal/app/dtool/common"
	"dev_tool/internal/app/dtool/component"
	_struct "dev_tool/internal/app/dtool/struct"
	"encoding/json"
	"regexp"
	"sort"
	"strings"

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
	responseData, err := buildTaskWorkflowResponse(info)
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
	responseData, err := buildTaskWorkflowResponse(info)
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

func buildTaskWorkflowResponse(workflowInfo map[string]any) (map[string]any, error) {
	homeTaskInfo, err := common.DbMain.HomeTaskRow(cast.ToInt(workflowInfo[`home_task_id`]))
	if err != nil {
		return nil, err
	}
	return map[string]any{
		`workflow`:  workflowInfo,
		`home_task`: homeTaskInfo,
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
