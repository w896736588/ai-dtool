package define

// AgentWsMessageType Agent <-> Server WebSocket 消息类型
type AgentWsMessageType string
type AgentTaskType string

const (
	// Server -> Agent
	AgentWsMsgTaskExecute AgentWsMessageType = "task_execute"      // 下发执行任务
	AgentWsMsgTaskCancel  AgentWsMessageType = "task_cancel"       // 取消任务（预留）
	AgentWsMsgReadyCheck  AgentWsMessageType = "agent_ready_check" // 探测 Agent 在线

	// Agent -> Server
	AgentWsMsgHello      AgentWsMessageType = "agent_hello"     // 建连后基础信息
	AgentWsMsgHeartbeat  AgentWsMessageType = "agent_heartbeat" // 心跳
	AgentWsMsgTaskStatus AgentWsMessageType = "task_status"     // 阶段状态
	AgentWsMsgTaskLog    AgentWsMessageType = "task_log"        // 实时日志
	AgentWsMsgTaskResult AgentWsMessageType = "task_result"     // 最终结果
)

const (
	AgentTaskTypePlaywrightRun    AgentTaskType = "playwright_run"
	AgentTaskTypeScrapeToMarkdown AgentTaskType = "scrape_to_markdown"
)

// AgentWsMessage 统一 WebSocket 消息结构
type AgentWsMessage struct {
	Type            AgentWsMessageType `json:"type"`
	ClientID        string             `json:"client_id,omitempty"`
	TaskID          string             `json:"task_id,omitempty"`
	SseDistributeId string             `json:"sse_distribute_id,omitempty"`
	Data            any                `json:"data,omitempty"`
}

// AgentRunParams 可序列化的 PlaywrightRunParams，用于服务端 -> Agent 下发
type AgentRunParams struct {
	Id                  int               `json:"id"`
	Link                string            `json:"link"`
	LinkIdLabel         string            `json:"link_id_label"`
	OpenNum             int               `json:"open_num"`
	Cookie              string            `json:"cookie"`
	Headers             map[string]string `json:"headers"`
	OpenType            int               `json:"open_type"`
	CombineType         int               `json:"combine_type"`
	ProcessList         []map[string]any  `json:"process_list"`
	ReplaceList         map[string]string `json:"replace_list"`
	BrowserAuthUsername string            `json:"browser_auth_username"`
	BrowserAuthPassword string            `json:"browser_auth_password"`
	Domain              string            `json:"domain"`
	Scheme              string            `json:"scheme"`
	LocatorTimeout      float64           `json:"locator_timeout"`
	GetPageTimeout      float64           `json:"get_page_timeout"`
	LastIndexLabel      string            `json:"last_index_label"`
	LinkId              string            `json:"link_id"`
	DownloadFinds       []string          `json:"download_finds"`
	AutoCloseSecond     int               `json:"auto_close_second"`
	Channel             string            `json:"channel"`
	FilterUris          []string          `json:"filter_uris"`
	ShowCookies         any               `json:"show_cookies"` // 原样传递，agent 侧反序列化
}

type AgentTaskScrapeConfig struct {
	JumpURL     string `json:"jump_url"`
	CssSelector string `json:"css_selector"`
	WaitSeconds int    `json:"wait_seconds"`
}

// AgentTaskExecuteData task_execute 消息的 data 结构
type AgentTaskExecuteData struct {
	TaskID          string                `json:"task_id"`
	SseDistributeId string                `json:"sse_distribute_id"`
	ClientID        string                `json:"client_id"`
	TaskType        AgentTaskType         `json:"task_type,omitempty"`
	SafeToken       string                `json:"safe_token,omitempty"`
	RunParams       AgentRunParams        `json:"run_params"`
	ScrapeConfig    AgentTaskScrapeConfig `json:"scrape_config,omitempty"`
}

// AgentTaskLogData task_log 消息的 data 结构
type AgentTaskLogData struct {
	Name    string `json:"name"`
	Message string `json:"message"`
}

// AgentTaskStatusData task_status 消息的 data 结构
type AgentTaskStatusData struct {
	Status string `json:"status"` // preparing_runtime, running, succeeded, failed
}

// AgentTaskResultData task_result 消息的 data 结构
type AgentTaskResultData struct {
	Status       string `json:"status"`
	ErrorMessage string `json:"error_message,omitempty"`
	FinishTime   int64  `json:"finish_time,omitempty"`
	DownloadURL  string `json:"download_url,omitempty"`
	FileName     string `json:"file_name,omitempty"`
}

type AgentTaskResultFileUploadResponse struct {
	DownloadURL string `json:"download_url"`
	FileName    string `json:"file_name"`
}

// AgentHelloData agent_hello 消息的 data 结构
type AgentHelloData struct {
	ClientVersion string `json:"client_version"`
	Hostname      string `json:"hostname"`
	Os            string `json:"os"`
	Arch          string `json:"arch"`
	UserName      string `json:"user_name"`
	RuntimeReady  bool   `json:"runtime_ready"`
}

// AgentHeartbeatData agent_heartbeat 消息的 data 结构
type AgentHeartbeatData struct {
	RuntimeReady  bool   `json:"runtime_ready"`
	CurrentTaskID string `json:"current_task_id,omitempty"`
}
