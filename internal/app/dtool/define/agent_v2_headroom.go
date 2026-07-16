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
	AutoStart       bool   `json:"auto_start"`        // 程序启动时自动启动 Headroom，默认 true
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

// AgentV2HeadroomActionRequest Headroom 通用操作请求（升级/统计/日志等）
type AgentV2HeadroomActionRequest struct {
	AgentId int    `json:"agent_id"`
	Action  string `json:"action"`             // upgrade / stats / log_list / log_read
	LogFile string `json:"log_file,omitempty"` // 日志文件名（log_read 时使用）
}

// AgentV2HeadroomUpgradeResponse 升级操作响应
type AgentV2HeadroomUpgradeResponse struct {
	Output  string `json:"output"`  // 命令输出
	Success bool   `json:"success"` // 是否成功
}

// AgentV2HeadroomStatsItem 单条统计项
type AgentV2HeadroomStatsItem struct {
	Label   string `json:"label"`    // 中文标签
	Key     string `json:"key"`      // 原始 JSON key
	Value   string `json:"value"`    // 格式化后的值
	RawJSON string `json:"raw_json"` // 嵌套对象的原始 JSON（可选）
}

// AgentV2HeadroomStatsResponse 统计信息响应
type AgentV2HeadroomStatsResponse struct {
	Items   []AgentV2HeadroomStatsItem `json:"items"`    // 结构化统计项列表
	RawJSON string                     `json:"raw_json"` // 原始 JSON 响应（备用）
}

// AgentV2HeadroomLogItem 日志文件项
type AgentV2HeadroomLogItem struct {
	Name    string `json:"name"`     // 文件名
	Size    int64  `json:"size"`     // 文件大小（字节）
	ModTime int64  `json:"mod_time"` // 修改时间戳
}

// AgentV2HeadroomLogContentResponse 日志内容响应
type AgentV2HeadroomLogContentResponse struct {
	Content  string `json:"content"`   // 日志内容（最多 200KB）
	FileName string `json:"file_name"` // 文件名
}

// AgentV2EnvToolUpgradeRequest 环境工具升级请求
type AgentV2EnvToolUpgradeRequest struct {
	AgentId int    `json:"agent_id"`
	Key     string `json:"key"`   // headroom / rtk
	Check   bool   `json:"check"` // 仅检查新版本（--check）
	Pre     bool   `json:"pre"`   // 包含预发布版本（--pre）
}

// AgentV2EnvToolUpgradeResponse 环境工具升级响应
type AgentV2EnvToolUpgradeResponse struct {
	Output  string `json:"output"`
	Success bool   `json:"success"`
	Check   bool   `json:"check"` // 是否是版本检查
}
