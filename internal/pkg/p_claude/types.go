package p_claude

// StreamMessage claude stream-json 单行解析结果。
type StreamMessage struct {
	Type    string         `json:"type"`
	Subtype string         `json:"subtype,omitempty"`
	Event   string         `json:"event,omitempty"`
	Data    map[string]any `json:"data,omitempty"`
	RawJSON string         `json:"-"`
}

// RunConfig claude 命令运行配置。
type RunConfig struct {
	Prompt         string // 用户提示词
	SessionID      string // 空=新对话，非空=--resume 续接
	Model          string // 模型标识，如 deepseek-v4-pro[1m]
	BaseURL        string // 服务商 API 地址（ANTHROPIC_BASE_URL）
	APIKey         string // 服务商 API Key（ANTHROPIC_API_KEY）
	WorkingDir     string // --add-dir 项目代码目录
	UserDataDir    string // --user-data-dir Claude Code 配置目录
	SettingsPath   string // --settings Agent CLI 配置路径
	Effort         string // --effort 思考强度 (low/medium/high/max)
	ThinkingBudget int    // THINKING_BUDGET 环境变量，精细控制
}

// DefaultUserDataDir 默认 Claude Code 用户数据目录。
// 空字符串表示使用 claude 默认目录（~/.claude），该目录已配置好 provider 信息。
const DefaultUserDataDir = ``
