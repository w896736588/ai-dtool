# 引入 claude-agent-sdk-go 方案设计

> 目标：新增 Agent CLI 类型 `claude-agent-cli`，基于 `claude-agent-sdk-go` 实现双向控制协议，
> 同时保持与现有前端 SSE 推送接口完全兼容。

---

## 一、整体架构

```
┌─────────────────────────────────────────────────────────────────────────┐
│                        前端 (Vue SPA)                                   │
│   AgentCliList.vue — 新增卡片类型 "claude-agent-cli"                     │
│   ChatPanel.vue — 复用现有 SSE 消息渲染逻辑                              │
│   + 新增：权限审批弹窗（PermissionDialog）                                │
│   + 新增：Hook 事件展示                                                  │
└──────────┬────────────────────────────────────┬─────────────────────────┘
           │ SSE 推送（兼容现有格式）              │ HTTP POST（审批回调）
           ▼                                     ▼
┌──────────────────────────────────────────────────────────────────────────┐
│                    Gin Router (router.go)                                │
│  /sse/agent_cli              — SSE 连接（不变）                          │
│  /api/agent/chat/send        — 发送对话（不变）                           │
│  /api/agent/chat/continue    — 继续对话（不变）                           │
│  /api/agent/chat/stop        — 停止对话（增强：支持协议级 Interrupt）       │
│  + /api/agent/chat/approve   — 新增：权限审批响应                         │
│  + /api/agent/chat/hooks     — 新增：查询 Hook 事件日志                   │
│  + /api/agent/chat/mcp/status — 新增：查询 MCP 服务器状态                 │
└──────────┬───────────────────────────────────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│               Controller (task_workflow.go)                               │
│  startChatCommand() — 新增 "claude-agent" 分支                           │
│  AgentChatSend()     — 新增 "claude-agent" 分支                          │
└──────────┬───────────────────────────────────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│          新增包：internal/pkg/p_claude_sdk/                               │
│                                                                          │
│  ┌──────────────────────────────────────────────────────────────┐        │
│  │  SessionManager                                              │        │
│  │  - 管理 SDK Client 生命周期（Connect/Close/复用）             │        │
│  │  - 同一 AgentCli 实例 + 同一 localDir → 复用 Client           │        │
│  │  - 会话级 context 取消                                        │        │
│  └──────────────────────────────────────────────────────────────┘        │
│           │                                                              │
│           ▼                                                              │
│  ┌──────────────────────────────────────────────────────────────┐        │
│  │  RunClaudeSdkStream(ctx, cfg, callback)                      │        │
│  │  - 构建 ClaudeAgentOptions                                    │        │
│  │  - 注册 PermissionCallback → 转为 SSE 权限请求事件             │        │
│  │  - 注册 Hook → 转为 SSE Hook 事件                             │        │
│  │  - Query/ReceiveResponse → 逐 Message 转换为 StreamMessage   │        │
│  │  - 兼容 p_claude.StreamMessage 格式                           │        │
│  └──────────────────────────────────────────────────────────────┘        │
│           │                                                              │
│           ▼                                                              │
│  ┌──────────────────────────────────────────────────────────────┐        │
│  │  MessageConverter                                             │        │
│  │  SDK Message → p_claude.StreamMessage (JSON 格式兼容)         │        │
│  │  - AssistantMessage → type:"assistant"                       │        │
│  │  - ResultMessage → type:"result"                              │        │
│  │  - SystemMessage → type:"system"                             │        │
│  │  - ToolUseBlock → type:"assistant", content.tool_use          │        │
│  │  - ToolResultBlock → type:"assistant", content.tool_result   │        │
│  │  - 权限请求 → type:"permission_request" (新增)                │        │
│  │  - Hook 事件 → type:"hook_event" (新增)                      │        │
│  └──────────────────────────────────────────────────────────────┘        │
└──────────────────────────────────────────────────────────────────────────┘
           │
           ▼
┌──────────────────────────────────────────────────────────────────────────┐
│          claude-agent-sdk-go (第三方依赖)                                 │
│  Client.Connect() → Client.Query() → Client.ReceiveResponse()            │
│  PermissionCallback, Hook, MCP                                          │
│       ↕ 双向控制协议 (stdin/stdout)                                       │
│  Claude Code CLI (Node.js 子进程)                                         │
└──────────────────────────────────────────────────────────────────────────┘
```

---

## 二、类型系统变更

### 2.1 新增 CLI 类型常量

**文件**: `internal/app/dtool/define/agent_cli.go`

```go
const (
    AgentCliTypeClaudeCodeCli  = "claude-code-cli"   // 现有：子进程 + stream-json
    AgentCliTypeCodexCli       = "codex-cli"          // 现有：codex exec --json
    AgentCliTypeClaudeAgentSdk = "claude-agent-cli"   // 新增：claude-agent-sdk-go 双向协议
)
```

### 2.2 新增运行时短值映射

**文件**: `internal/app/dtool/controller/agent_cli_runtime_validation.go`

```go
func normalizeAgentCliRuntimeType(cliType string) string {
    switch strings.TrimSpace(cliType) {
    case `codex`, define.AgentCliTypeCodexCli:
        return define.AgentCliTypeCodexCli
    case `claude`, ``, define.AgentCliTypeClaudeCodeCli:
        return define.AgentCliTypeClaudeCodeCli
    case `claude-agent`, define.AgentCliTypeClaudeAgentSdk:  // 新增
        return define.AgentCliTypeClaudeAgentSdk
    default:
        return strings.TrimSpace(cliType)
    }
}
```

### 2.3 新增配置结构

**文件**: `internal/app/dtool/define/agent_cli.go`

```go
// ClaudeAgentSdkConfig Claude Agent SDK 实例配置（存储在 tbl_agent_cli.config JSON 字段中）
type ClaudeAgentSdkConfig struct {
    ApiKey           string   `json:"api_key"`                         // ANTHROPIC_API_KEY 或 CLAUDE_API_KEY
    OAuthToken       string   `json:"oauth_token,omitempty"`           // CLAUDE_CODE_OAUTH_TOKEN（Max 订阅）
    Model            string   `json:"model"`                           // 默认模型，如 claude-sonnet-4-20250514
    Models           []string `json:"models,omitempty"`                // 可选模型列表
    BaseURL          string   `json:"base_url,omitempty"`              // 自定义 API 端点（ANTHROPIC_BASE_URL）
    UserDataDir      string   `json:"user_data_dir,omitempty"`        // --user-data-dir
    SettingsPath     string   `json:"settings_path,omitempty"`        // --settings 路径
    PermissionMode   string   `json:"permission_mode,omitempty"`       // 默认权限模式：bypassPermissions / acceptEdits / default
    AllowedTools     []string `json:"allowed_tools,omitempty"`         // 允许的工具列表
    MaxTurns         int      `json:"max_turns,omitempty"`             // 最大对话轮次
    EnableHooks      bool     `json:"enable_hooks,omitempty"`          // 是否启用 Hook 事件推送
    EnableMcpStatus  bool     `json:"enable_mcp_status,omitempty"`    // 是否启用 MCP 状态查询
}
```

### 2.4 数据库变更

`tbl_agent_cli` 表的 `type` 字段新增值 `"claude-agent-cli"`，`config` 字段存储 `ClaudeAgentSdkConfig` JSON。**无需 DDL 变更**，type 字段为 varchar。

`tbl_agent_chat` 表的 `cli_type` 字段新增值 `"claude-agent"`。**无需 DDL 变更**。

---

## 三、新增包：internal/pkg/p_claude_sdk/

### 3.1 目录结构

```
internal/pkg/p_claude_sdk/
├── types.go              # 配置结构、SSE 事件扩展类型
├── converter.go           # SDK Message → StreamMessage 格式转换
├── session.go             # SessionManager：Client 生命周期管理
├── exec.go                # RunClaudeSdkStream：核心执行入口
├── exec_windows.go        # Windows 平台特定处理（进程管理）
├── permission.go          # 权限审批桥接：SSE 请求 ↔ HTTP 响应
└── hook.go                # Hook 事件桥接：SDK Hook → SSE 事件
```

### 3.2 types.go — 核心类型定义

```go
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
    Prompt             string                       // 用户提示词
    SessionID          string                       // 空=新对话，非空=复用 Client 继续
    Model              string                       // 模型标识
    BaseURL            string                       // ANTHROPIC_BASE_URL
    APIKey             string                       // ANTHROPIC_API_KEY / CLAUDE_API_KEY
    OAuthToken         string                       // CLAUDE_CODE_OAUTH_TOKEN
    WorkingDir         string                       // 工作目录（CWD）
    UserDataDir        string                       // --user-data-dir
    SettingsPath       string                       // --settings 路径
    PermissionMode     string                       // 权限模式
    AllowedTools       []string                     // 允许的工具列表
    MaxTurns           int                          // 最大对话轮次
    EnableHooks        bool                         // 是否启用 Hook 事件
    ProcessStartCallback func(pid int)              // 进程启动回调
    // ApprovalChan 用于接收前端的审批响应，由 permission.go 管理
    ApprovalChan       chan<- *ApprovalResponse
}

// ApprovalRequest 权限审批请求（推送前端）。
type ApprovalRequest struct {
    RequestID string `json:"request_id"`           // 唯一请求 ID
    ToolName  string `json:"tool_name"`             // 工具名称，如 "Bash"、"Write"
    Input     any    `json:"input"`                  // 工具输入参数
    SessionID string `json:"session_id"`            // 关联的会话 ID
    ChatID    int64  `json:"chat_id"`                // 关联的对话 ID
}

// ApprovalResponse 权限审批响应（前端回传）。
type ApprovalResponse struct {
    RequestID string `json:"request_id"`            // 对应 ApprovalRequest.RequestID
    Approved  bool   `json:"approved"`               // true=允许，false=拒绝
    Reason    string `json:"reason,omitempty"`       // 拒绝原因（可选）
}

// HookEvent Hook 生命周期事件（推送前端）。
type HookEvent struct {
    HookType  string `json:"hook_type"`              // "PreToolUse" / "PostToolUse" 等
    ToolName  string `json:"tool_name,omitempty"`    // 触发的工具名称
    Input     any    `json:"input,omitempty"`        // 工具输入
    Output    any    `json:"output,omitempty"`        // 工具输出（PostToolUse）
    SessionID string `json:"session_id"`             // 关联的会话 ID
    ChatID    int64  `json:"chat_id"`                // 关联的对话 ID
}
```

### 3.3 converter.go — SDK 消息格式转换

核心职责：将 `claude-agent-sdk-go` 的 `types.Message` 转换为前端已知的 `p_claude.StreamMessage` JSON 格式。

**映射规则**：

| SDK Message 类型 | 转换后的 StreamMessage JSON | 说明 |
|-----------------|---------------------------|------|
| `*types.AssistantMessage` | `{"type":"assistant","message":{...},"session_id":"..."}` | 与现有 claude stream-json 格式对齐 |
| `*types.UserMessage` | `{"type":"user","message":{...}}` | 用户消息回显 |
| `*types.SystemMessage` | `{"type":"system","subtype":"init",...}` | 初始化/系统消息 |
| `*types.ResultMessage` | `{"type":"result","result":"...","session_id":"...","cost_usd":...}` | 最终结果 |
| 权限请求 | `{"type":"permission_request","request_id":"...","tool_name":"Bash","input":{...}}` | **新增类型** |
| Hook 事件 | `{"type":"hook_event","hook_type":"PreToolUse","tool_name":"Bash",...}` | **新增类型** |

**关键**：AssistantMessage 的 ContentBlock 需要特别处理：

| SDK ContentBlock | 前端已有对应字段 | 说明 |
|-----------------|---------------|------|
| `*types.TextBlock` | `content[].type="text"` | 文本内容 |
| `*types.ToolUseBlock` | `content[].type="tool_use"` | 工具调用 |
| `*types.ToolResultBlock` | `content[].type="tool_result"` | 工具结果 |

```go
// ConvertSDKMessage 将 SDK Message 转换为前端兼容的 StreamMessage。
// 保持与 p_claude stream-json 输出格式完全一致，前端无需修改。
func ConvertSDKMessage(msg types.Message, sessionID string) StreamMessage {
    switch m := msg.(type) {
    case *types.AssistantMessage:
        return convertAssistantMessage(m, sessionID)
    case *types.ResultMessage:
        return convertResultMessage(m, sessionID)
    case *types.SystemMessage:
        return convertSystemMessage(m, sessionID)
    case *types.UserMessage:
        return convertUserMessage(m, sessionID)
    default:
        return StreamMessage{Type: "unknown", Data: map[string]any{"raw": fmt.Sprintf("%v", msg)}}
    }
}
```

### 3.4 session.go — Client 生命周期管理

```go
// SessionManager 管理活跃的 SDK Client 实例。
// 设计原则：
//   - 同一 (agentCliID + localDir) 组合复用同一个 Client
//   - Client 内部维护会话上下文，无需每轮重启进程
//   - 支持优雅关闭和 context 取消
type SessionManager struct {
    mu      sync.RWMutex
    clients map[string]*sdkClientEntry // key: "agentCliID:localDir"
}

type sdkClientEntry struct {
    Client    *claude.Client
    Cancel    context.CancelFunc
    SessionID string               // SDK 会话 ID
    RefCount  int                   // 引用计数
    LastUsed  time.Time
    ApprovalCh chan *ApprovalResponse  // 权限审批响应通道
}

// GetOrCreateClient 获取或创建 SDK Client。
func (sm *SessionManager) GetOrCreateClient(ctx context.Context, cfg RunConfig) (*claude.Client, string, error)

// CloseClient 关闭指定 Client。
func (sm *SessionManager) CloseClient(agentCliID int, localDir string) error

// CloseAll 关闭所有 Client（服务关闭时调用）。
func (sm *SessionManager) CloseAll() error
```

### 3.5 exec.go — 核心执行入口

```go
// RunClaudeSdkStream 使用 claude-agent-sdk-go 执行对话并逐条推送消息。
// 接口签名与 p_claude.RunClaudeStream 完全一致，上层调用无需修改。
func RunClaudeSdkStream(ctx context.Context, cfg RunConfig, callback func(msg StreamMessage)) (string, error) {
    // 1. 从 SessionManager 获取或创建 Client
    client, sessionID, err := sessionMgr.GetOrCreateClient(ctx, cfg)
    
    // 2. 如果是新 Client，需要 Connect
    if !client.IsConnected() {
        if err := client.Connect(ctx); err != nil {
            return "", fmt.Errorf("claude-agent-sdk connect failed: %w", err)
        }
    }
    
    // 3. 发送 Query
    if err := client.Query(ctx, cfg.Prompt); err != nil {
        return "", fmt.Errorf("claude-agent-sdk query failed: %w", err)
    }
    
    // 4. 消费 ReceiveResponse channel
    for msg := range client.ReceiveResponse(ctx) {
        streamMsg := ConvertSDKMessage(msg, sessionID)
        callback(streamMsg)
        
        // 提取 session_id
        if !sessionExtracted {
            if sid := extractSessionID(msg); sid != "" {
                sessionID = sid
                sessionExtracted = true
            }
        }
    }
    
    return sessionID, nil
}
```

### 3.6 permission.go — 权限审批桥接

这是引入 SDK 的**核心新增能力**。流程如下：

```
SDK 发起权限请求 → PermissionCallback
    ↓ 生成 ApprovalRequest（含 request_id）
    ↓ 推送 SSE 事件 type:"permission_request" → 前端显示审批弹窗
    ↓ 用户点击 允许/拒绝 → POST /api/agent/chat/approve
    ↓ 写入 approvalMap[request_id] = response
    ↓ PermissionCallback 从 approvalMap 读取结果 → 返回 SDK
```

```go
// pendingApprovals 等待审批的请求，key=requestID
var pendingApprovals sync.Map

// NewPermissionCallback 创建权限审批回调。
// 当 SDK 拦截到工具调用时，将请求推送到前端，然后阻塞等待前端响应。
func NewPermissionCallback(chatID int64, sendSse func(msg StreamMessage)) func(ctx context.Context, toolName string, input interface{}) (bool, error) {
    return func(ctx context.Context, toolName string, input interface{}) (bool, error) {
        requestID := uuid.New().String()
        
        // 1. 构建权限请求事件
        req := ApprovalRequest{
            RequestID: requestID,
            ToolName:  toolName,
            Input:     input,
            ChatID:    chatID,
        }
        
        // 2. 通过 SSE 推送到前端
        reqJSON, _ := json.Marshal(req)
        sendSse(StreamMessage{
            Type:    "permission_request",
            Data:    map[string]any{"raw": string(reqJSON)},
            RawJSON: string(reqJSON),
        })
        
        // 3. 阻塞等待前端审批响应（带超时）
        responseCh := make(chan *ApprovalResponse, 1)
        pendingApprovals.Store(requestID, responseCh)
        defer pendingApprovals.Delete(requestID)
        
        select {
        case resp := <-responseCh:
            return resp.Approved, nil
        case <-ctx.Done():
            return false, ctx.Err()
        case <-time.After(5 * time.Minute):
            return false, fmt.Errorf("权限审批超时: %s", toolName)
        }
    }
}

// HandleApprovalResponse 处理前端审批响应。
func HandleApprovalResponse(resp *ApprovalResponse) error {
    val, ok := pendingApprovals.LoadAndDelete(resp.RequestID)
    if !ok {
        return fmt.Errorf("无效的审批请求 ID: %s", resp.RequestID)
    }
    val.(chan *ApprovalResponse) <- resp
    return nil
}
```

### 3.7 hook.go — Hook 事件桥接

```go
// NewHookCallbacks 创建 Hook 回调。
func NewHookCallbacks(chatID int64, sendSse func(msg StreamMessage)) map[types.HookEvent][]types.HookCallbackFunc {
    hooks := make(map[types.HookEvent][]types.HookCallbackFunc)
    
    preToolUse := func(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
        event := HookEvent{
            HookType: "PreToolUse",
            ToolName: extractToolName(input),
            Input:    input,
            ChatID:   chatID,
        }
        eventJSON, _ := json.Marshal(event)
        sendSse(StreamMessage{
            Type:    "hook_event",
            RawJSON: string(eventJSON),
            Data:    map[string]any{"raw": string(eventJSON)},
        })
        return map[string]interface{}{"continue": true}, nil
    }
    
    postToolUse := func(ctx context.Context, input interface{}, toolUseID *string, hookCtx types.HookContext) (interface{}, error) {
        event := HookEvent{
            HookType: "PostToolUse",
            ToolName: extractToolName(input),
            Output:   input,
            ChatID:   chatID,
        }
        eventJSON, _ := json.Marshal(event)
        sendSse(StreamMessage{
            Type:    "hook_event",
            RawJSON: string(eventJSON),
            Data:    map[string]any{"raw": string(eventJSON)},
        })
        return map[string]interface{}{"continue": true}, nil
    }
    
    hooks[types.HookEventPreToolUse] = []types.HookCallbackFunc{preToolUse}
    hooks[types.HookEventPostToolUse] = []types.HookCallbackFunc{postToolUse}
    return hooks
}
```

---

## 四、Controller 层变更

### 4.1 startChatCommand — 新增分支

**文件**: `internal/app/dtool/controller/task_workflow.go`

```go
func startChatCommand(chatID int64) {
    // ... 现有逻辑 ...
    cliType := cast.ToString(chatInfo[`cli_type`])
    switch cliType {
    case `codex`:
        // ... 现有 codex 逻辑 ...
    case `claude-agent`:  // 新增分支
        agentCliId := cast.ToInt(chatInfo[`agent_cli_id`])
        configJson := ``
        if agentCliId > 0 {
            cliRow, _ := common.DbMain.Client.QueryBySql(
                `SELECT config FROM tbl_agent_cli WHERE id = ?`, agentCliId,
            ).One()
            if len(cliRow) > 0 {
                configJson = cast.ToString(cliRow[`config`])
            }
        }
        selectedModelName := strings.TrimSpace(cast.ToString(chatInfo[`model_name`]))
        go runClaudeSdkCommand(chatID, fromType, localDir, prompt, isResume, sessionID, configJson, selectedModelName)
    default:
        // ... 现有 claude 逻辑 ...
    }
}
```

### 4.2 新增 runClaudeSdkCommand

```go
// runClaudeSdkCommand 使用 claude-agent-sdk-go 执行对话。
// 结构与 runClaudeCommand 对齐，复用 SSE 推送和 DB 写入逻辑。
func runClaudeSdkCommand(chatID int64, fromType string, localDir, prompt string, isResume bool, sessionID string, configJson string, selectedModelName string) {
    // 1. 解析 ClaudeAgentSdkConfig
    var cfgData define.ClaudeAgentSdkConfig
    _ = json.Unmarshal([]byte(configJson), &cfgData)
    
    // 2. 构建 p_claude_sdk.RunConfig
    if selectedModelName != "" {
        cfgData.Model = selectedModelName
    }
    cfg := p_claude_sdk.RunConfig{
        Prompt:         prompt,
        SessionID:      sessionID,
        Model:          cfgData.Model,
        BaseURL:        cfgData.BaseURL,
        APIKey:         cfgData.ApiKey,
        OAuthToken:      cfgData.OAuthToken,
        WorkingDir:     localDir,
        UserDataDir:    cfgData.UserDataDir,
        SettingsPath:   cfgData.SettingsPath,
        PermissionMode: cfgData.PermissionMode,
        AllowedTools:   cfgData.AllowedTools,
        MaxTurns:       cfgData.MaxTurns,
        EnableHooks:    cfgData.EnableHooks,
        ProcessStartCallback: func(pid int) {
            _ = common.DbMain.TaskWorkflowChatUpdatePID(chatID, pid)
        },
    }
    
    // 3. 复用 runClaudeCommand 的 SSE/DB 推送逻辑
    // ...（与 runClaudeCommand 相同的 sendLine/dbWriteCh 逻辑）
    
    // 4. 执行 SDK 流式查询
    _, err := p_claude_sdk.RunClaudeSdkStream(ctx, cfg, func(msg p_claude_sdk.StreamMessage) {
        // 与 runClaudeCommand 中的 callback 逻辑一致
        sendLine(msg.RawJSON)
        // ... session_id 提取、assistant text 跟踪等 ...
    })
    
    // 5. 完成后处理（与 runClaudeCommand 一致）
}
```

### 4.3 新增审批 API

**文件**: `internal/app/dtool/controller/task_workflow.go`

```go
// AgentChatApprove 处理前端权限审批响应。
func AgentChatApprove(c *gin.Context) {
    var req struct {
        RequestID string `json:"request_id"`
        Approved  bool   `json:"approved"`
        Reason    string `json:"reason,omitempty"`
    }
    if err := gsgin.GinPostBody(c, &req); err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    
    resp := &p_claude_sdk.ApprovalResponse{
        RequestID: req.RequestID,
        Approved:  req.Approved,
        Reason:    req.Reason,
    }
    
    if err := p_claude_sdk.HandleApprovalResponse(resp); err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    
    gsgin.GinResponseSuccess(c, "", nil)
}
```

### 4.4 新增 MCP 状态查询 API

```go
// AgentChatMcpStatus 查询 MCP 服务器状态。
func AgentChatMcpStatus(c *gin.Context) {
    var req struct {
        AgentCliID int    `json:"agent_cli_id"`
        LocalDir   string `json:"local_dir"`
    }
    if err := gsgin.GinPostBody(c, &req); err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    
    status, err := p_claude_sdk.GetMcpStatus(c.Request.Context(), req.AgentCliID, req.LocalDir)
    if err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    
    gsgin.GinResponseSuccess(c, "", status)
}
```

### 4.5 增强停止对话 — 协议级 Interrupt

```go
// 现有的 TaskWorkflowChatStop / AgentChatStop 逻辑增强：
// 如果是 claude-agent 类型，优先调用 client.Interrupt() 协议级中断，
// 而不是直接 kill 进程。
func stopChatByCliType(chatID int64, cliType string) {
    switch cliType {
    case "claude-agent":
        // 协议级中断
        p_claude_sdk.InterruptSession(chatID)
    default:
        // 现有逻辑：context 取消 + 进程 kill
        if cancelFn, ok := chatCancelFuncs.Load(chatID); ok {
            cancelFn.(func())()
        }
    }
}
```

---

## 五、路由注册

**文件**: `internal/app/dtool/router.go`

```go
// 新增路由
agentChatGroup.POST("/approve", controller.AgentChatApprove)
agentChatGroup.POST("/mcp/status", controller.AgentChatMcpStatus)
```

---

## 六、Business 层变更

**文件**: `internal/app/dtool/business/agent_cli.go`

新增 `ClaudeAgentSdkConfig` 相关的读写函数：

```go
// GetClaudeAgentSdkConfig 从 DB config 字段解析 ClaudeAgentSdkConfig。
func GetClaudeAgentSdkConfig(configJson string) define.ClaudeAgentSdkConfig {
    var cfg define.ClaudeAgentSdkConfig
    _ = json.Unmarshal([]byte(configJson), &cfg)
    return cfg
}

// GetClaudeAgentSdkModelConfig 获取模型和 API 配置。
func GetClaudeAgentSdkModelConfig(configJson string) (model, baseURL, apiKey string) {
    cfg := GetClaudeAgentSdkConfig(configJson)
    return cfg.Model, cfg.BaseURL, cfg.ApiKey
}
```

---

## 七、前端变更（最小化）

### 7.1 原则

- **不影响现有逻辑**：现有 `claude-code-cli` 和 `codex-cli` 卡片和对话流程完全不变
- **样式兼容**：新类型复用现有卡片样式和对话面板
- **渐进增强**：新增的权限审批和 Hook 事件只在 `claude-agent-cli` 类型时展示

### 7.2 AgentCliList.vue 变更

```javascript
// 1. 新增类型选项
const AGENT_CLI_TYPES = [
    { value: 'claude-code-cli', label: 'Claude Code CLI', desc: '子进程 + stream-json 单向流' },
    { value: 'codex-cli', label: 'Codex CLI', desc: 'codex exec --json 子进程' },
    { value: 'claude-agent-cli', label: 'Claude Agent SDK', desc: '双向控制协议，支持权限审批/Hook/MCP' },  // 新增
]

// 2. 新建卡片时，根据类型显示不同配置表单
//    - claude-code-cli: settings.json 路径选择（现有）
//    - codex-cli: API Key / Model / Base URL 配置（现有）
//    - claude-agent-cli: API Key / OAuth Token / Model / 权限模式 / 允许工具列表 / Hook 开关（新增）

// 3. 卡片状态展示
//    - 现有字段：名称、类型、模型、MCP 服务器数量、启用状态
//    - 新增展示：权限模式标签、Hook 状态图标（仅 claude-agent-cli 类型）
```

### 7.3 ChatPanel.vue 变更

```javascript
// 1. SSE 消息渲染 — 完全复用现有逻辑
//    因为 p_claude_sdk.ConvertSDKMessage 输出的 JSON 格式与 p_claude 的 stream-json 格式一致，
//    前端的消息气泡、代码块、工具调用展示等无需修改。

// 2. 新增：权限审批弹窗（仅 claude-agent-cli 类型时显示）
//    监听 SSE 事件 type === 'permission_request'
//    显示弹窗：工具名称 + 输入参数 + 允许/拒绝按钮
//    用户点击后 POST /api/agent/chat/approve

// 3. 新增：Hook 事件展示（可选，低优先级）
//    监听 SSE 事件 type === 'hook_event'
//    在工具调用气泡旁显示 Hook 状态标记（如 PreToolUse → 执行中，PostToolUse → 已完成）
```

### 7.4 新增组件

```
web/src/components/agent_cli/
├── PermissionDialog.vue    # 权限审批弹窗
└── HookEventBadge.vue      # Hook 事件状态标记（可选）
```

**PermissionDialog.vue 设计**：

```
┌─────────────────────────────────────────┐
│  🔐 工具权限请求                         │
│                                          │
│  工具: Bash                              │
│  命令: rm -rf /tmp/old-files             │
│                                          │
│  ┌─────────────┐  ┌─────────────┐       │
│  │   ✅ 允许    │  │   ❌ 拒绝    │       │
│  └─────────────┘  └─────────────┘       │
│                                          │
│  ⏱️ 等待审批（5:00 超时）                  │
└─────────────────────────────────────────┘
```

---

## 八、依赖管理

### 8.1 go.mod 新增依赖

```
go get github.com/schlunsen/claude-agent-sdk-go@v0.9.0
```

该 SDK 核心零依赖，仅使用 Go 标准库。

### 8.2 前端无新增依赖

PermissionDialog 和 HookEventBadge 使用现有 Element Plus / 自定义组件实现。

---

## 九、配置管理

### 9.1 Claude Agent SDK 配置表单

当用户选择 `claude-agent-cli` 类型时，前端显示以下配置表单：

| 字段 | 类型 | 必填 | 说明 |
|------|------|------|------|
| API Key | input(password) | 是* | ANTHROPIC_API_KEY |
| OAuth Token | input(password) | 否 | Max 订阅用户可选 |
| 默认模型 | select | 是 | 从 models 列表选择 |
| 模型列表 | tag-input | 否 | 可选模型，供执行前切换 |
| API 端点 | input | 否 | 自定义 ANTHROPIC_BASE_URL |
| 用户数据目录 | input | 否 | --user-data-dir |
| Settings 路径 | input | 否 | --settings |
| 权限模式 | select | 是 | bypassPermissions / acceptEdits / default |
| 允许工具 | tag-input | 否 | 如 Bash, Write, Read |
| 最大轮次 | number | 否 | 0=无限制 |
| 启用 Hook | switch | 否 | 是否推送 Hook 事件 |
| 启用 MCP 状态 | switch | 否 | 是否启用 MCP 状态查询 |

> *API Key 和 OAuth Token 二选一必填

### 9.2 权限模式说明

| 模式 | 说明 | 适用场景 |
|------|------|---------|
| `bypassPermissions` | 全部放行，无需审批 | 受信环境、自动化 |
| `acceptEdits` | 自动允许文件编辑，其他需审批 | 开发场景 |
| `default` | 所有工具调用需审批 | 安全敏感场景 |

---

## 十、与现有类型对比

| 特性 | claude-code-cli | codex-cli | claude-agent-cli (新增) |
|------|----------------|-----------|------------------------|
| 通信方式 | CLI stdout → JSONL | CLI stdout → JSONL | SDK 双向控制协议 |
| 流式输出 | ✅ callback | ✅ callback | ✅ channel → callback |
| 多轮对话 | ⚠️ 重启进程 --resume | ⚠️ 重启进程 exec resume | ✅ 持久 Client 复用 |
| 权限审批 | ❌ bypassPermissions | ❌ danger-full-access | ✅ 前端弹窗交互审批 |
| Hook 系统 | ❌ | ❌ | ✅ PreToolUse/PostToolUse |
| MCP 动态管理 | ❌ 配置文件预定义 | ❌ 配置文件预定义 | ✅ 运行时查询/重连 |
| 协议级中断 | ❌ kill 进程 | ❌ kill 进程 | ✅ client.Interrupt() |
| 结构化日志 | ❌ | ❌ | ✅ SDK stderr 回调 |
| 会话管理 | ⚠️ session_id 提取 | ⚠️ thread_id 提取 | ✅ SessionManager |
| 设置模型 | ❌ 重启生效 | ❌ 重启生效 | ✅ client.SetModel() |
| 前端兼容 | ✅ | ✅ | ✅（消息格式对齐） |
| Windows 支持 | ✅ Job Object | ✅ Job Object | ⚠️ SDK 声明有限 |

---

## 十一、实施步骤

### Phase 1：基础框架（2-3 天）

1. 新增 `AgentCliTypeClaudeAgentSdk` 类型常量和配置结构
2. 创建 `internal/pkg/p_claude_sdk/` 包骨架
3. 实现 `converter.go`：SDK Message → StreamMessage 格式转换
4. 实现 `exec.go`：`RunClaudeSdkStream` 基础版（无权限/Hook）
5. Controller 层新增 `claude-agent` 分支和 `runClaudeSdkCommand`
6. 前端新增 `claude-agent-cli` 卡片类型选项和基础配置表单

**验收标准**：能通过 `claude-agent-cli` 类型发起对话，前端正常显示消息流。

### Phase 2：权限审批（2 天）

1. 实现 `permission.go`：权限审批桥接
2. Controller 新增 `/api/agent/chat/approve` 接口
3. 前端新增 `PermissionDialog.vue`
4. SSE 新增 `permission_request` 事件处理

**验收标准**：设置为 `acceptEdits`/`default` 模式时，工具调用会在前端弹窗请求审批。

### Phase 3：Hook 和 MCP（2 天）

1. 实现 `hook.go`：Hook 事件桥接
2. 实现 `session.go`：SessionManager 客户端复用
3. Controller 新增 `/api/agent/chat/mcp/status` 接口
4. 前端新增 Hook 事件展示

**验收标准**：Hook 事件实时显示，MCP 状态可查询，Client 多轮复用。

### Phase 4：增强功能（1-2 天）

1. 协议级 Interrupt 替代进程 kill
2. 动态模型切换 `client.SetModel()`
3. 运行时权限模式切换 `client.SetPermissionMode()`
4. SDK stderr 日志集成

**验收标准**：停止对话使用协议级中断，可运行时切换模型和权限模式。

---

## 十二、风险与应对

| 风险 | 影响 | 应对 |
|------|------|------|
| SDK Windows 兼容性 | 进程管理可能异常 | 保留 p_claude 作为 fallback，配置切换 |
| SDK 版本更新 | CLI 版本不兼容 | 锁定 SDK 版本，跟进测试 |
| 权限审批超时 | 对话卡住 | 5 分钟超时自动拒绝 |
| Client 泄漏 | 资源占用 | SessionManager 定期清理过期 Client |
| 消息格式偏差 | 前端渲染异常 | converter.go 充分测试，对照 stream-json 输出 |
| SDK 依赖变更 | 编译问题 | go.sum 锁定 |

---

## 十三、回退方案

如果 `claude-agent-sdk-go` 在生产环境中出现严重问题（如 Windows 兼容性），可以一键回退：

1. 前端隐藏 `claude-agent-cli` 类型选项
2. Controller 层 `claude-agent` 分支 fallback 到现有 `runClaudeCommand` 逻辑
3. 数据库中 `type=claude-agent-cli` 的记录降级为 `claude-code-cli`

不影响现有两种类型的任何功能。
