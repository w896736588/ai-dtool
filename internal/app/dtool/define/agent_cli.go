package define

// 支持的 Agent CLI 类型常量
const (
	AgentCliTypeClaudeCodeCli = "claude-code-cli"
	AgentCliTypeCodexCli      = "codex-cli"
)

// Agent CLI 启停状态常量 / Agent CLI enabled flag constants.
const (
	AgentCliDisabled = 0
	AgentCliEnabled  = 1
)

// Codex CLI 默认 sandbox 模式
const CodexCliDefaultSandboxMode = "danger-full-access"

// Codex CLI 默认 wire_api，兼容现有行为。
const CodexCliDefaultWireAPI = "responses"

// CodexCliConfig Codex CLI 实例配置（存储在 tbl_agent_cli.config JSON 字段中）
type CodexCliConfig struct {
	ApiKey             string   `json:"api_key"`
	Model              string   `json:"model"`
	Models             []string `json:"models,omitempty"`              // 可选模型列表，首个模型视为默认模型。 // Optional model list, the first item is treated as default.
	BaseURL            string   `json:"base_url,omitempty"`            // 自定义 API 端点（可选）
	SandboxMode        string   `json:"sandbox_mode,omitempty"`        // 默认 "danger-full-access"
	WireAPI            string   `json:"wire_api,omitempty"`            // 请求格式，支持 responses / chat。
	SupportsWebsockets *bool    `json:"supports_websockets,omitempty"` // 是否允许 Responses API WebSocket 传输。 // Whether Responses API WebSocket transport is enabled.
}

// AgentCliConfig Agent CLI 通用配置（当前仅存储模型列表等 UI 扩展字段）。
// AgentCliConfig stores shared Agent CLI config extensions, currently model options for UI/runtime selection.
type AgentCliConfig struct {
	Models []string `json:"models,omitempty"`
}

// AgentCliItem 列表项
type AgentCliItem struct {
	Id                int    `json:"id"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	SettingsPath      string `json:"settings_path"`
	Config            string `json:"config"` // 提供商专属配置 JSON（Codex 用）
	Enabled           int    `json:"enabled"`
	ThinkingCollapsed int    `json:"thinking_collapsed"`
	WebhookConfigId   int    `json:"webhook_config_id"`
	CreatedAt         int64  `json:"created_at"`
	UpdatedAt         int64  `json:"updated_at"`
}

// AgentCliSaveRequest 新建/编辑请求
type AgentCliSaveRequest struct {
	Id                int    `json:"id,omitempty"`
	Name              string `json:"name,omitempty"`
	Type              string `json:"type,omitempty"`
	SettingsPath      string `json:"settings_path,omitempty"`
	Config            string `json:"config,omitempty"` // 提供商专属配置 JSON（Codex 用）
	Enabled           int    `json:"enabled,omitempty"`
	ThinkingCollapsed int    `json:"thinking_collapsed,omitempty"`
	WebhookConfigId   int    `json:"webhook_config_id,omitempty"`
}

// AgentCliDeleteRequest 删除请求
type AgentCliDeleteRequest struct {
	Id int `json:"id"`
}

// AgentCliReadSettingsRequest 读取 settings.json 请求
type AgentCliReadSettingsRequest struct {
	Id int `json:"id"`
}

// AgentCliReadSettingsResponse 读取 settings.json 响应
type AgentCliReadSettingsResponse struct {
	Exists       bool   `json:"exists"`
	Content      string `json:"content"`
	SettingsPath string `json:"settings_path"`
}

// AgentCliWriteMcpRequest 写入 mcpServers 请求
type AgentCliWriteMcpRequest struct {
	Id int `json:"id"`
}

// AgentCliWriteDeepSeekRequest 写入 DeepSeek 配置请求
type AgentCliWriteDeepSeekRequest struct {
	Id        int      `json:"id"`
	ModelName string   `json:"model_name"`
	ModelList []string `json:"model_list,omitempty"` // 模型列表，首个模型会写回 settings.json 作为默认模型。 // Model list; the first model is written back to settings.json as the default.
	ApiKey    string   `json:"api_key"`
	BaseUrl   string   `json:"base_url"`
}

// 思考强度常量
const (
	ThinkingIntensityLow      = "低"
	ThinkingIntensityMedium   = "中等"
	ThinkingIntensityHigh     = "高"
	ThinkingIntensityVeryHigh = "很高"
	ThinkingIntensityHighest  = "最高"
)

// ThinkingIntensityBudgetMap 思考强度对应的 thinking budget token 数
var ThinkingIntensityBudgetMap = map[string]int{
	ThinkingIntensityLow:      1024,
	ThinkingIntensityMedium:   4096,
	ThinkingIntensityHigh:     8192,
	ThinkingIntensityVeryHigh: 16384,
	ThinkingIntensityHighest:  32000,
}

// ThinkingIntensityEffortMap 思考强度对应的 --effort 值
var ThinkingIntensityEffortMap = map[string]string{
	ThinkingIntensityLow:      "low",
	ThinkingIntensityMedium:   "medium",
	ThinkingIntensityHigh:     "high",
	ThinkingIntensityVeryHigh: "max",
	ThinkingIntensityHighest:  "max",
}

// AgentCliStatusItem 列表带状态
type AgentCliStatusItem struct {
	AgentCliItem
	SettingsExists    bool     `json:"settings_exists"`
	DisplayedEnabled  bool     `json:"displayed_enabled"`
	CurrentModel      string   `json:"current_model"`
	ModelOptions      []string `json:"model_options"` // 卡片可选模型列表，用于前端执行前选择。 // Selectable models for this card, used before task execution.
	RequestURL        string   `json:"request_url"`
	McpServerCount    int      `json:"mcp_server_count"`
	WebhookConfigName string   `json:"webhook_config_name"`
	GroupIds          []int    `json:"group_ids"` // 所属分组 ID 列表（多对多） // Group IDs this Agent CLI belongs to.
}

// AgentCliToggleEnabledRequest 切换 Agent CLI 启停请求 / Toggle Agent CLI enabled status request.
type AgentCliToggleEnabledRequest struct {
	Id     int  `json:"id"`
	Enable bool `json:"enable"`
}
