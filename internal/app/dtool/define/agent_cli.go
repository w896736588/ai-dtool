package define

// 支持的 Agent CLI 类型常量
const (
	AgentCliTypeClaudeCodeCli = "claude-code-cli"
	AgentCliTypeCodexCli      = "codex-cli"
)

// Codex CLI 默认 sandbox 模式
const CodexCliDefaultSandboxMode = "danger-full-access"

// CodexCliConfig Codex CLI 实例配置（存储在 tbl_agent_cli.config JSON 字段中）
type CodexCliConfig struct {
	ApiKey      string `json:"api_key"`
	Model       string `json:"model"`
	BaseURL     string `json:"base_url,omitempty"`     // 自定义 API 端点（可选）
	SandboxMode string `json:"sandbox_mode,omitempty"` // 默认 "danger-full-access"
}

// AgentCliItem 列表项
type AgentCliItem struct {
	Id                int    `json:"id"`
	Name              string `json:"name"`
	Type              string `json:"type"`
	SettingsPath      string `json:"settings_path"`
	Config            string `json:"config"` // 提供商专属配置 JSON（Codex 用）
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
	Id        int    `json:"id"`
	ModelName string `json:"model_name"`
	ApiKey    string `json:"api_key"`
	BaseUrl   string `json:"base_url"`
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
	SettingsExists    bool   `json:"settings_exists"`
	CurrentModel      string `json:"current_model"`
	McpServerCount    int    `json:"mcp_server_count"`
	ClaudeMemEnabled  bool   `json:"claude_mem_enabled"`
	WebhookConfigName string `json:"webhook_config_name"`
}

// AgentCliToggleClaudeMemRequest 切换 claude-mem 启停请求
type AgentCliToggleClaudeMemRequest struct {
	Id     int  `json:"id"`
	Enable bool `json:"enable"`
}
