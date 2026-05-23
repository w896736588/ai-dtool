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
	HomeTaskID   int    `json:"home_task_id"`
	NodeStatuses string `json:"node_statuses"`
	Step         string `json:"step"`
	Status       string `json:"status"`
}

// TaskWorkflowBatchNodeStatusRequest 批量查询工作流节点状态请求。
type TaskWorkflowBatchNodeStatusRequest struct {
	HomeTaskIDs []int `json:"home_task_ids"`
}

// TaskWorkflowChatSendRequest 发送对话到 claude code 请求。
type TaskWorkflowChatSendRequest struct {
	WorkflowID        int    `json:"workflow_id"`
	Prompt            string `json:"prompt"`
	PromptType        string `json:"prompt_type"`
	CliType           string `json:"cli_type"`
	LocalDir          string `json:"local_dir"`
	AgentCliId        int    `json:"agent_cli_id"`
	ModelName         string `json:"model_name"`
	ThinkingIntensity string `json:"thinking_intensity"`
}

// TaskWorkflowChatDirsRequest 获取可选工作目录列表请求。
type TaskWorkflowChatDirsRequest struct {
	WorkflowID int `json:"workflow_id"`
}

// TaskWorkflowChatContinueRequest 继续已有对话请求。
type TaskWorkflowChatContinueRequest struct {
	ChatID int    `json:"chat_id"`
	Prompt string `json:"prompt"`
}

// TaskWorkflowChatListRequest 列出对话列表请求。
type TaskWorkflowChatListRequest struct {
	WorkflowID int `json:"workflow_id"`
}

// TaskWorkflowChatDetailRequest 获取对话详情请求。
type TaskWorkflowChatDetailRequest struct {
	ChatID int `json:"chat_id"`
}

// TaskWorkflowZcodeSaveRequest 保存 zcode 配置请求。
type TaskWorkflowZcodeSaveRequest struct {
	ZcodeDir string `json:"zcode_dir"`
}

// TaskWorkflowZcodeProjectItem 项目映射条目（用于响应）。
type TaskWorkflowZcodeProjectItem struct {
	ProjectKey    string `json:"project_key"`
	WorkspacePath string `json:"workspace_path"`
	SettingsPath  string `json:"settings_path"`
}

// TaskWorkflowChatStopRequest 停止运行中的对话请求。
type TaskWorkflowChatStopRequest struct {
	ChatID int `json:"chat_id"`
}

// TaskWorkflowChatListByPromptTypeRequest 按提示词类型查询对话列表请求。
type TaskWorkflowChatListByPromptTypeRequest struct {
	WorkflowID int    `json:"workflow_id"`
	PromptType string `json:"prompt_type"`
}
