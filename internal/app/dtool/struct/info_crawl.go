package _struct

// InfoCrawlTask 信息抓取任务。
type InfoCrawlTask struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Prompt     string `json:"prompt"`
	AiModelID  int    `json:"ai_model_id"`
	Status     int    `json:"status"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

// InfoCrawlRun 信息抓取执行记录。
type InfoCrawlRun struct {
	ID              int    `json:"id"`
	TaskID          int    `json:"task_id"`
	Status          string `json:"status"`
	RunMessage      string `json:"run_message"`
	PromptSnapshot  string `json:"prompt_snapshot"`
	AiModelSnapshot string `json:"ai_model_snapshot"`
	OutputContent   string `json:"output_content"`
	ErrorMessage    string `json:"error_message"`
	CreateTime      int64  `json:"create_time"`
	UpdateTime      int64  `json:"update_time"`
}

// InfoCrawlTaskSaveRequest 保存任务请求。
type InfoCrawlTaskSaveRequest struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Prompt    string `json:"prompt"`
	AiModelID int    `json:"ai_model_id"`
}

// InfoCrawlTaskRunRequest 执行任务请求。
type InfoCrawlTaskRunRequest struct {
	TaskID          int    `json:"task_id"`
	SseDistributeID string `json:"sse_distribute_id"`
}
