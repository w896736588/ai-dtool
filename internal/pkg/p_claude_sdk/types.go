package p_claude_sdk

import "dev_tool/internal/pkg/p_claude"

// StreamMessage 复用 p_claude.StreamMessage 格式，保持前端兼容。
// 新增 type 值：
//   - "permission_request"：工具权限审批请求
//   - "permission_response"：工具权限审批响应（前端回传）
//   - "hook_event"：Hook 生命周期事件
//   - "mcp_status"：MCP 服务器状态变更
type StreamMessage = p_claude.StreamMessage

// RunConfig Claude Agent SDK 运行配置。
type RunConfig struct {
	Prompt               string        // 用户提示词
	SessionID            string        // 空=新对话，非空=复用 Client 继续
	Model                string        // 模型标识
	BaseURL              string        // ANTHROPIC_BASE_URL
	APIKey               string        // ANTHROPIC_API_KEY / CLAUDE_API_KEY
	OAuthToken           string        // CLAUDE_CODE_OAUTH_TOKEN
	WorkingDir           string        // 工作目录（CWD）
	UserDataDir          string        // --user-data-dir
	SettingsPath         string        // --settings 路径
	PermissionMode       string        // 权限模式：bypassPermissions / acceptEdits / default
	AllowedTools         []string      // 允许的工具列表
	MaxTurns             int           // 最大对话轮次，0=无限制
	EnableHooks          bool          // 是否启用 Hook 事件推送
	ProcessStartCallback func(pid int) // 进程启动回调（SDK 模式下可能无法获取 PID）
	// HasApprovalSink 标记是否注册了权限审批回调（前端 SSE 连接时设为 true）
	HasApprovalSink bool
}

// ApprovalRequest 权限审批请求（推送前端）。
type ApprovalRequest struct {
	RequestID string `json:"request_id"` // 唯一请求 ID
	ToolName  string `json:"tool_name"`  // 工具名称，如 "Bash"、"Write"
	Input     any    `json:"input"`      // 工具输入参数
	SessionID string `json:"session_id"` // 关联的会话 ID
	ChatID    int64  `json:"chat_id"`    // 关联的对话 ID
}

// ApprovalResponse 权限审批响应（前端回传）。
type ApprovalResponse struct {
	RequestID string `json:"request_id"`       // 对应 ApprovalRequest.RequestID
	Approved  bool   `json:"approved"`         // true=允许，false=拒绝
	Reason    string `json:"reason,omitempty"` // 拒绝原因（可选）
}

// HookEvent Hook 生命周期事件（推送前端）。
type HookEvent struct {
	HookType  string `json:"hook_type"`           // "PreToolUse" / "PostToolUse" 等
	ToolName  string `json:"tool_name,omitempty"` // 触发的工具名称
	Input     any    `json:"input,omitempty"`     // 工具输入
	Output    any    `json:"output,omitempty"`    // 工具输出（PostToolUse）
	SessionID string `json:"session_id"`          // 关联的会话 ID
	ChatID    int64  `json:"chat_id"`             // 关联的对话 ID
}

// 权限模式常量
const (
	PermissionModeBypass  = "bypassPermissions" // 全部放行，无需审批
	PermissionModeAccept  = "acceptEdits"       // 自动允许文件编辑，其他需审批
	PermissionModeDefault = "default"           // 所有工具调用需审批
)

// 权限审批超时时间（5 分钟），超时自动拒绝。
const PermissionTimeoutSeconds = 300

// MaxWaitGoroutineExit SDK 模式下等待 goroutine 退出的最大时间。
const MaxWaitGoroutineExit = 10 * 60 // SDK 模式需要更长的超时时间（秒）
