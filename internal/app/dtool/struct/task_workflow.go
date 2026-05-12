package _struct

// TaskWorkflowCreateOrGetRequest 查询或创建任务工作流请求。
type TaskWorkflowCreateOrGetRequest struct {
	HomeTaskID int `json:"home_task_id"`
}

// TaskWorkflowInfoRequest 查询任务工作流详情请求。
type TaskWorkflowInfoRequest struct {
	WorkflowID int `json:"workflow_id"`
}

// TaskWorkflowDevPlanSaveRequest 保存开发执行文档请求。
type TaskWorkflowDevPlanSaveRequest struct {
	WorkflowID int    `json:"workflow_id"`
	Content    string `json:"content"`
}

// TaskWorkflowGenerateRequest 生成覆盖分析/测试计划请求。
type TaskWorkflowGenerateRequest struct {
	WorkflowID int `json:"workflow_id"`
}

// TaskWorkflowExecuteRequest 执行测试计划请求。
type TaskWorkflowExecuteRequest struct {
	WorkflowID      int  `json:"workflow_id"`
	RegeneratePlan  bool `json:"regenerate_plan"`
	IncludeCoverage bool `json:"include_coverage"`
}

// TaskWorkflowUIAssistGenerateRequest 页面辅助识别请求。
type TaskWorkflowUIAssistGenerateRequest struct {
	WorkflowID  int    `json:"workflow_id"`
	SmartLinkID int    `json:"smart_link_id"`
	Label       string `json:"label"`
	JumpURL     string `json:"jump_url"`
	CssSelector string `json:"css_selector"`
	WaitSeconds int    `json:"wait_seconds"`
}

// TaskWorkflowPromptsSaveRequest 保存工作流提示词请求。
type TaskWorkflowPromptsSaveRequest struct {
	WorkflowID                  int    `json:"workflow_id"`
	PromptRequirement           string `json:"prompt_requirement"`
	PromptApiDev                string `json:"prompt_api_dev"`
	PromptApiTest               string `json:"prompt_api_test"`
	PromptDesign                string `json:"prompt_design"`
	PromptPlainTextRequirement  string `json:"prompt_plain_text_requirement"`
	PromptDesignPlanRequirement string `json:"prompt_design_plan_requirement"`
	PromptBrowserTest           string `json:"prompt_browser_test"`
	PromptCodeReview            string `json:"prompt_code_review"`
}

// TaskWorkflowPromptsRestoreRequest 还原工作流提示词请求。
type TaskWorkflowPromptsRestoreRequest struct {
	WorkflowID int `json:"workflow_id"`
}

// TaskWorkflowRequirementFetchRequest 抓取 TAPD 需求文档请求。
type TaskWorkflowRequirementFetchRequest struct {
	WorkflowID int `json:"workflow_id"`
}

// TaskWorkflowNodeStatusUpdateRequest 更新工作流节点状态请求。
type TaskWorkflowNodeStatusUpdateRequest struct {
	WorkflowID   int    `json:"workflow_id"`
	NodeStatuses string `json:"node_statuses"`
}

// TaskWorkflowBatchNodeStatusRequest 批量查询工作流节点状态请求。
type TaskWorkflowBatchNodeStatusRequest struct {
	HomeTaskIDs []int `json:"home_task_ids"`
}
