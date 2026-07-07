package define

// Agent V2 类型常量
const (
	AgentV2TypePi         = "pi"
	AgentV2TypeCodex      = "codex"
	AgentV2TypeClaudeCode = "claude-code"
)

// DefaultPiSessionDir Pi Agent 会话 JSONL 默认存储目录
const DefaultPiSessionDir = "logs/pi_agent_sessions"

// DefaultBuiltinToolsDir 内置工具目录（相对于程序工作目录，通常为项目根目录）
// 注意：此路径依赖工作目录，若以不同 CWD 启动可能失效，后续考虑改为基于可执行文件路径或配置文件动态解析
const DefaultBuiltinToolsDir = "internal/app/dtool/data"

// AgentV2Item Agent 配置项
type AgentV2Item struct {
	Id        int    `json:"id"`
	Name      string `json:"name"`
	Type      string `json:"type"`
	Config    string `json:"config"`
	Enabled   int    `json:"enabled"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// AgentV2StatusItem 带状态摘要的 Agent 项
type AgentV2StatusItem struct {
	AgentV2Item
	Installed    bool   `json:"installed"`
	InstallHint  string `json:"install_hint"`
	SessionCount int    `json:"session_count"`
}

// AgentV2Config Pi Agent 配置
type AgentV2PiConfig struct {
	Provider   string `json:"provider"`
	Model      string `json:"model"`
	SessionDir string `json:"session_dir"`
	ExtraArgs  string `json:"extra_args"`
}

// AgentV2SaveRequest 保存请求
type AgentV2SaveRequest struct {
	Id     int    `json:"id,omitempty"`
	Name   string `json:"name"`
	Type   string `json:"type"`
	Config string `json:"config"`
}

// AgentV2Workspace 工作空间
type AgentV2Workspace struct {
	Id        int    `json:"id"`
	AgentId   int    `json:"agent_id"`
	Name      string `json:"name"`
	Path      string `json:"path"`
	CreatedAt int64  `json:"created_at"`
}

// AgentV2WorkspaceSaveRequest 工作空间保存请求
type AgentV2WorkspaceSaveRequest struct {
	Id      int    `json:"id,omitempty"`
	AgentId int    `json:"agent_id"`
	Name    string `json:"name"`
	Path    string `json:"path"`
}

// AgentV2Session 会话
type AgentV2Session struct {
	Id          int    `json:"id"`
	AgentId     int    `json:"agent_id"`
	WorkspaceId int    `json:"workspace_id"`
	Name        string `json:"name"`
	SessionDir  string `json:"session_dir"`
	ModelName   string `json:"model_name"`
	Status      string `json:"status"`
	CreatedAt   int64  `json:"created_at"`
	UpdatedAt   int64  `json:"updated_at"`
}

// AgentV2SessionSaveRequest 会话保存请求
type AgentV2SessionSaveRequest struct {
	Id          int    `json:"id,omitempty"`
	AgentId     int    `json:"agent_id"`
	WorkspaceId int    `json:"workspace_id"`
	Name        string `json:"name"`
}

// AgentV2Skill Skill/Tool 配置
type AgentV2Skill struct {
	Id        int    `json:"id"`
	AgentId   int    `json:"agent_id"`
	Name      string `json:"name"`
	SkillType string `json:"skill_type"`
	Config    string `json:"config"`
	Enabled   int    `json:"enabled"`
	CreatedAt int64  `json:"created_at"`
	UpdatedAt int64  `json:"updated_at"`
}

// AgentV2SkillSaveRequest Skill 保存请求
type AgentV2SkillSaveRequest struct {
	Id        int    `json:"id,omitempty"`
	AgentId   int    `json:"agent_id"`
	Name      string `json:"name"`
	SkillType string `json:"skill_type"`
	Config    string `json:"config"`
	Enabled   int    `json:"enabled"`
}

// AgentV2BuiltinTool 内置工具定义
type AgentV2BuiltinTool struct {
	DirName         string                 `json:"dir_name"`
	Name            string                 `json:"name"`
	ToolName        string                 `json:"tool_name"`
	Description     string                 `json:"description"`
	ToolDescription string                 `json:"tool_description"`
	Parameters      []AgentV2ToolParameter `json:"parameters"`
	ScriptContent   string                 `json:"script_content"`
}

// AgentV2ToolParameter 工具参数定义
type AgentV2ToolParameter struct {
	Name        string `json:"name"`
	Type        string `json:"type"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
}

// AgentV2WSMessage WebSocket 通信消息
type AgentV2WSMessage struct {
	Type    string      `json:"type"`              // command / event / state / error
	Id      string      `json:"id,omitempty"`      // 客户端生成的消息ID，用于关联请求和响应
	Command interface{} `json:"command,omitempty"` // 发送给 Agent 的命令
	Event   interface{} `json:"event,omitempty"`   // Agent 返回的事件
	State   interface{} `json:"state,omitempty"`   // 会话状态
	Error   string      `json:"error,omitempty"`   // 错误信息
}
