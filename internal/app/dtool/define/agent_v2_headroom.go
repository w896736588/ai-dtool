package define

// Headroom 代理模式配置（仅支持 pi 类型 Agent）
// 配置独立存储于 tbl_agent_v2.headroom_config 列，与其他扩展配置解耦
type AgentV2HeadroomConfig struct {
	Port            int    `json:"port"`              // 代理端口，默认 8787
	AnthropicApiUrl string `json:"anthropic_api_url"` // Anthropic 上游地址
	OpenaiApiUrl    string `json:"openai_api_url"`    // OpenAI 上游地址
	GeminiApiUrl    string `json:"gemini_api_url"`    // Gemini 上游地址
	CloudcodeApiUrl string `json:"cloudcode_api_url"` // Cloud Code Assist 上游地址
	VertexApiUrl    string `json:"vertex_api_url"`    // Vertex AI 上游地址
}

// AgentV2HeadroomStatus Headroom 运行时状态
type AgentV2HeadroomStatus struct {
	AgentV2HeadroomConfig
	Installed bool   `json:"installed"`  // headroom CLI 是否在 PATH 中
	Version   string `json:"version"`    // headroom 版本号
	Running   bool   `json:"running"`    // 代理进程是否运行中
	Pid       int32  `json:"pid"`        // 进程 PID（运行中时有效）
	StartedAt int64  `json:"started_at"` // 进程启动时间戳
}

// AgentV2HeadroomSaveRequest 保存 Headroom 配置请求
type AgentV2HeadroomSaveRequest struct {
	AgentId int                   `json:"agent_id"`
	Config  AgentV2HeadroomConfig `json:"config"`
}

// AgentV2HeadroomProcessRequest Headroom 进程操作请求
type AgentV2HeadroomProcessRequest struct {
	AgentId int    `json:"agent_id"`
	Action  string `json:"action"` // start / stop / restart
}
