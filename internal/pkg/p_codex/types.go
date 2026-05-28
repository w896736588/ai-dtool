package p_codex

// StreamMessage codex exec --json 单行 JSONL 解析结果。
type StreamMessage struct {
	Type     string         `json:"type"`      // "thread.started", "turn.started", "item.started", "item.updated", "item.completed", "turn.completed", "turn.failed", "error"
	ItemType string         `json:"item_type"` // "agent_message", "reasoning", "command_execution", "file_change", "mcp_tool_call", "web_search", "todo_list"
	ItemID   string         `json:"item_id"`
	Data     map[string]any `json:"data,omitempty"`
	RawJSON  string         `json:"-"`
}

// RunConfig codex 命令运行配置。
type RunConfig struct {
	Prompt               string        // 用户提示词
	SessionID            string        // 空=新对话，非空=codex exec resume <session_id> 续接
	Model                string        // 模型标识，如 o3, codex-mini-latest
	APIKey               string        // OPENAI_API_KEY
	BaseURL              string        // OPENAI_BASE_URL（可选，自定义 API 端点）
	WorkingDir           string        // --cd 项目代码目录
	SandboxMode          string        // --sandbox 模式，默认 "danger-full-access"
	ProcessStartCallback func(pid int) // 非空时在进程启动后同步回调，用于上层记录 PID
}

// DefaultSandboxMode 默认 sandbox 模式，允许文件编辑和网络访问。
const DefaultSandboxMode = "danger-full-access"
