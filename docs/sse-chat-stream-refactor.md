# SSE Chat Stream 重构方案

## 产品概述

彻底重构 Codex CLI 和 Claude CLI 执行任务时的 SSE 逻辑，将"一个对话绑一个 SSE 链接"的旧设计改为业务级独立 SSE 连接模式，每个业务（AgentCli、TaskWorkflow）各自维护一条独立 SSE 长连接，消除状态错乱、重复连接、页面刷新异常、历史对话更新不到位等问题，并清理旧代码。

## 核心特性

- 进入 `/AgentCli` 页面时创建到 `/sse/agent_cli` 的独立 SSE 连接，注册 `agent_cli_chat_output` 分发回调，不再创建任何 per-chat EventSource
- 进入 `/TaskWorkflow/:taskId` 页面时创建到 `/sse/task_workflow` 的独立 SSE 连接，注册 `task_workflow_chat_output` 分发回调，不再创建任何 per-chat EventSource
- 两个业务各自拥有独立 SSE 路由连接，互不干扰，连接生命周期与页面绑定
- 所有对话输出通过各自业务的 SSE 连接推送，消息自动携带 `chat_id`，前端按 chat_id 过滤渲染
- 前端收到对话 completed 事件，且当前执行历史弹窗打开并选中该对话时，自动调用 API 置为已读
- 后端 `TaskWorkflowChatSend`/`AgentChatSend`/`TaskWorkflowChatContinue` API 直接启动 CLI 命令 goroutine，不再依赖 SSE 连接触发
- 移除 `/api/task/workflow/chat/stream` 独立 SSE 路由及全部 chat stream 专用连接管理代码
- 移除 `isChatStreamSseClient` 过滤逻辑和 `chatSseConns` 连接池
- ChatReplyPage 同时连接两个业务 SSE 路由，根据来源接收流式输出

## 技术栈

- 后端：Go (现有项目栈，gin + gsgin SSE 框架)
- 前端：Vue 3 + JavaScript (现有项目栈，Element Plus)
- SSE：每业务独立 SSE 路由（`SseRoute`），内部分发使用 `sse_distribute_id` 区分消息类型

---

## 现有架构分析

### 两套 SSE 体系并存

1. **全局分发 SSE (`sse_distribute`)**：前端维护单一条长连接到 `/sse`，后端按 `sse_distribute_id` 路由消息到各业务回调
2. **对话专用 SSE (`chat stream`)**：每次查看/新建对话创建一条独立 `EventSource` 到 `/api/task/workflow/chat/stream?chat_id=xxx`

### Chat Stream SSE 旧连接模式（需重构）

```
前端每个对话创建独立 EventSource
  ↓
后端 chatSseConns (sync.Map) 按 chatID 管理多条 SSE 连接
  ↓
前端维护 _chatEventSource（前台）和 _backgroundChatEventSources（后台监听其他运行中对话）
  ↓
切换对话时：关闭旧前台 SSE → 旧选中 running 对话切到后台 → 新对话建新前台 SSE
```

### 后端旧推送流程

```
sendLine(line) → broadcastToChatSse(chatID, line) → 遍历 chatSseConns[chatID] 所有连接推送
完成事件通过 buildChatCompletedEvent 构建
未读/状态变更通过全局 SSE 推送（跳过 chat stream 连接 via isChatStreamSseClient 过滤）
```

### 关键问题

- 多个 SSE 连接管理复杂，前台/后台切换容易出错
- 连接断开/重连逻辑复杂
- `isChatStreamSseClient` 过滤在多处重复出现
- 状态同步依赖 SSE 连接生命周期，连接断开即丢失

---

## 目标架构

### 核心设计思路

将"每对话独立 SSE 连接"改为"每业务独立 SSE 连接"，AgentCli 和 TaskWorkflow 各自维护一条独立 SSE 长连接（`/sse/agent_cli` 和 `/sse/task_workflow`），两个业务互不干扰。后端所有 chat 输出推送到对应业务的 SSE 通道（按 `agent_cli_chat_output` / `task_workflow_chat_output` 分发 ID 区分来源），前端在各自的业务 SSE 连接上注册回调，根据消息中的 `chat_id` 决定是否渲染到当前选中对话。

### 新 SSE 连接模型

```
旧模型（每对话独立连接）：
  /api/task/workflow/chat/stream?chat_id=1  → EventSource 1
  /api/task/workflow/chat/stream?chat_id=2  → EventSource 2
  /api/task/workflow/chat/stream?chat_id=3  → EventSource 3
  → N 个对话 = N 个 SSE 连接

新模型（每业务独立连接）：
  /sse/agent_cli?client_id=xxx  → 1 条 SSE 连接，所有 AgentCli 对话共用
  /sse/task_workflow?client_id=xxx  → 1 条 SSE 连接，所有 TaskWorkflow 对话共用
  → 2 个业务 = 2 个 SSE 连接（与对话数量无关）
```

### 新 SSE 消息格式

```json
{
  "sse_distribute_id": "agent_cli_chat_output",
  "data": {
    "chat_id": 123,
    "line": "{\"type\":\"assistant\",\"content\":[{\"type\":\"text\",\"text\":\"...\"}]}"
  },
  "type": "chat_output"
}
```

前端收到后解包 `data.line` 即为原有的 raw line，可直接喂入现有 `chatParser.parseChatLinesIncremental`。

---

## 后端改造

### 1. 新增 SSE 分发 ID 常量

**文件**: `internal/app/dtool/define/sse.go`

新增：

```go
SseAgentCliChatOutput    = `agent_cli_chat_output`    // AgentCli 页面聊天输出分发
SseTaskWorkflowChatOutput = `task_workflow_chat_output` // TaskWorkflow 页面聊天输出分发
```

### 2. 新增两条独立业务 SSE 路由

**文件**: `internal/app/dtool/router.go`

在 `IsSsePort` 条件内，与 `/sse` 全局路由并列注册两条业务 SSE 路由：

```go
// 全局 SSE（现有，不动）
tGin.SseRoute(`/sse`, openFunc, closeFunc)
// AgentCli 业务独立 SSE
tGin.SseRoute(`/sse/agent_cli`, controller.AgentCliChatSseOpen, controller.AgentCliChatSseClose)
// TaskWorkflow 业务独立 SSE
tGin.SseRoute(`/sse/task_workflow`, controller.TaskWorkflowChatSseOpen, controller.TaskWorkflowChatSseClose)
```

注意：这两条路由同样只在 SSE 端口（`IsSsePort`）上注册，与全局 SSE 保持一致。

### 3. 新增业务 SSE Open/Close 函数

**文件**: `internal/app/dtool/controller/task_workflow.go`

#### AgentCliChatSseOpen

```go
func AgentCliChatSseOpen(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
    clientID := strings.TrimSpace(urlValues.Get(`client_id`))
    if clientID == `` {
        return nil, fmt.Errorf(`client_id 不能为空`)
    }
    connID := fmt.Sprintf("agent_cli_sse_%s_%d", clientID, time.Now().UnixNano())
    sse := gsgin.SseRegister(connID, stopC, c)
    agentCliSseConns.Store(clientID, sse) // 仅每个 clientID 一条连接
    gstool.FmtPrintlnLogTime("[agent-cli-sse] clientID=%s 已建立 SSE 连接", clientID)
    return sse, nil
}
```

#### AgentCliChatSseClose

```go
func AgentCliChatSseClose(sse *gsgin.Sse) {
    // 从 agentCliSseConns 中移除该连接
    agentCliSseConns.Range(func(key, value any) bool {
        if value == sse {
            agentCliSseConns.Delete(key)
            return false
        }
        return true
    })
}
```

#### TaskWorkflowChatSseOpen

```go
func TaskWorkflowChatSseOpen(urlValues url.Values, stopC chan int, c *gin.Context) (*gsgin.Sse, error) {
    clientID := strings.TrimSpace(urlValues.Get(`client_id`))
    if clientID == `` {
        return nil, fmt.Errorf(`client_id 不能为空`)
    }
    connID := fmt.Sprintf("task_workflow_sse_%s_%d", clientID, time.Now().UnixNano())
    sse := gsgin.SseRegister(connID, stopC, c)
    taskWorkflowSseConns.Store(clientID, sse) // 仅每个 clientID 一条连接
    gstool.FmtPrintlnLogTime("[task-workflow-sse] clientID=%s 已建立 SSE 连接", clientID)
    return sse, nil
}
```

#### TaskWorkflowChatSseClose

```go
func TaskWorkflowChatSseClose(sse *gsgin.Sse) {
    // 从 taskWorkflowSseConns 中移除该连接
    taskWorkflowSseConns.Range(func(key, value any) bool {
        if value == sse {
            taskWorkflowSseConns.Delete(key)
            return false
        }
        return true
    })
}
```

新增全局变量：

```go
var agentCliSseConns sync.Map      // key: clientID, value: *gsgin.Sse
var taskWorkflowSseConns sync.Map  // key: clientID, value: *gsgin.Sse
```

### 4. 替换 broadcastToChatSse 为业务 SSE 推送

**文件**: `internal/app/dtool/controller/task_workflow.go`

新增函数 `broadcastChatLineToBusinessSse(chatID int64, line string)`：

- 查询 chat 的 `from_type` 确定来源（在 goroutine 启动时缓存 `fromType` 到局部变量，避免每次查 DB）
- 将 `line` 包装为 `{ chat_id, line }` 作为 SseData.Data
- `agent_cli` 来源 → 使用 `SseAgentCliChatOutput` 分发 ID，遍历 `agentCliSseConns` 所有连接推送
- `workflow` 来源 → 使用 `SseTaskWorkflowChatOutput` 分发 ID，遍历 `taskWorkflowSseConns` 所有连接推送
- 移除 `isChatStreamSseClient` 过滤（不再有 chat stream 专用连接）

### 5. 修改 sendLine 逻辑

**文件**: `internal/app/dtool/controller/task_workflow.go`

在 `runClaudeCommand` 和 `runCodexCommand` 中：

- `sendLine(line)` → `broadcastChatLineToBusinessSse(chatID, line)` 替代 `broadcastToChatSse(chatID, line)`
- DB 写入通道逻辑不变
- 在 goroutine 启动时缓存 `fromType` 到局部变量，避免每次 sendLine 都查 DB

### 6. 移除 chat stream SSE 路由和连接管理

**涉及文件**:

- `internal/app/dtool/router.go` — 移除 `/api/task/workflow/chat/stream` 的 SseRoute 注册
- `internal/app/dtool/controller/task_workflow.go` — 删除以下全局变量和函数：
  - `chatSseConns` 全局变量
  - `addChatSseConn` / `removeChatSseConn` / `broadcastToChatSse` / `chatHasActiveSse`
  - `TaskWorkflowChatStreamOpen` / `TaskWorkflowChatStreamClose`
  - `taskWorkflowAutoMarkChatReadIfSseConnected`

### 7. 移除 isChatStreamSseClient 过滤逻辑

**涉及文件**:

- `internal/app/dtool/controller/api_sse.go` — 删除 `isChatStreamSseClient` 函数定义，移除广播中的 `isChatStreamSseClient` 过滤
- `internal/app/dtool/controller/memory_fragment.go` — 移除广播中的 `isChatStreamSseClient` 过滤
- `internal/app/dtool/controller/browser_port_pool.go` — 移除广播中的 `isChatStreamSseClient` 过滤
- `internal/app/dtool/controller/workflow_unread_sse.go` — 移除广播中的 `isChatStreamSseClient` 过滤
- `internal/app/dtool/controller/task_workflow.go` — 移除广播中的 `isChatStreamSseClient` 过滤

### 8. 将命令启动逻辑迁移到 Send/Continue API

**文件**: `internal/app/dtool/controller/task_workflow.go`

提取 `TaskWorkflowChatStreamOpen` 中的 goroutine 启动逻辑为独立函数 `startChatCommand(chatID int64)`：

- 读取 chatInfo 获取 localDir、cliType、settingsPath、sessionID 等配置
- 根据 cliType 启动 `runClaudeCommand` 或 `runCodexCommand`
- 在 `TaskWorkflowChatSend` 创建 DB 记录后调用 `startChatCommand`
- 在 `AgentChatSend` 创建 DB 记录后调用 `startChatCommand`
- 在 `TaskWorkflowChatContinue` 标记 running 后调用 `startChatCommand`

### 9. 孤立 running 状态检测

**文件**: `internal/app/dtool/controller/task_workflow.go`

在 `TaskWorkflowChatDetail` API 中增加检测：当 DB 状态为 running 但 goroutine 不存在时，标记为 interrupted 并广播状态变更。替代旧 SSE 连接重连时的检测逻辑。

---

## 前端改造

### 1. 新增业务 SSE 连接管理工具

**文件**: `web/src/utils/base/sse_business.js`（新增）

封装 AgentCli 和 TaskWorkflow 两条独立业务 SSE 连接的管理，复用 `sse_distribute` 的分发模式：

```javascript
// 每个业务独立维护自己的 EventSource 和分发回调
const businessSse = {
  agent_cli: { conn: null, url: '', receiveHandlers: {} },
  task_workflow: { conn: null, url: '', receiveHandlers: {} },
}

// ConnectBusinessSse(businessType, ssePort, clientId)
//   创建到 /sse/agent_cli 或 /sse/task_workflow 的 EventSource
//   businessType: 'agent_cli' | 'task_workflow'

// RegisterBusinessReceive(businessType, receiveId, callFunc)
//   在指定业务的 SSE 连接上注册分发回调（与 sse_distribute.RegisterReceive 模式一致）

// UnRegisterBusinessReceive(businessType, receiveId, callFunc)
//   注销指定业务的分发回调

// CloseBusinessSse(businessType)
//   关闭指定业务的 SSE 连接，清空所有回调
```

**设计要点**：
- 每个业务独立管理 `EventSource` 实例、URL、回调 Map
- 消息格式与全局 SSE 一致，包含 `sse_distribute_id`、`data`、`type` 字段
- 复用现有的端口查询（`fetchAvailableSsePort`）和 token 参数逻辑
- 连接生命周期与页面绑定：mounted 创建，beforeUnmount 关闭

### 2. AgentCliList.vue

**文件**: `web/src/components/agent_cli/AgentCliList.vue`

**删除**:

- `connectChatStream` 方法
- `_chatEventSource` / `_sseChatId` 状态
- `_backgroundChatEventSources` 状态
- `startBackgroundChatStream` / `stopBackgroundChatStream` / `stopAllBackgroundChatStreams` / `updateBackgroundChatListItem` 方法
- `chatDetailSSERegistered` 状态
- `_initialSseRetryCount` 状态

**新增**:

- `mounted` 时调用 `ConnectBusinessSse('agent_cli', ssePort, clientId)` 建立业务 SSE 连接
- `registerChatOutputSse()` / `unregisterChatOutputSse()`：
  - `mounted` 时注册 `RegisterBusinessReceive('agent_cli', 'agent_cli_chat_output', handler)`
  - handler 中根据 `chat_id` 判断：若等于当前选中 chatDetailId → 追加到 `chatDetailSSELines` + 解析渲染；否则仅更新列表状态
  - `beforeUnmount` 时 `UnRegisterBusinessReceive` + `CloseBusinessSse('agent_cli')`

- completed 事件处理：收到 completed 且 chat_id 匹配当前选中对话 → `loadChatDetail()` + 刷新计数；若执行历史弹窗打开且选中该对话 → 调用 `AgentChatMarkRead`

**修改**:

- 发送/继续对话流程：调用 API 后不再调用 `connectChatStream`，改为设置 chatDetailId + 状态即可，SSE 回调自动接收输出
- `stopChat()`: 移除 `_chatEventSource.close()` 逻辑，仅调用后端 stop API

### 3. TaskWorkflow.vue

**文件**: `web/src/components/TaskWorkflow.vue`

同 AgentCliList.vue 模式，但使用 `task_workflow` 业务类型：
- `mounted` 时调用 `ConnectBusinessSse('task_workflow', ssePort, clientId)` 建立业务 SSE 连接
- 注册 `RegisterBusinessReceive('task_workflow', 'task_workflow_chat_output', handler)`
- `beforeUnmount` 时 `UnRegisterBusinessReceive` + `CloseBusinessSse('task_workflow')`
- 删除所有 chat stream 相关方法，修改发送/继续/停止对话流程

### 4. ChatReplyPage.vue

**文件**: `web/src/components/ChatReplyPage.vue`

- 删除 `connectChatStream` / `_chatEventSource` / `closeSSE`
- 根据对话来源 `from_type`，连接对应业务的 SSE：
  - `agent_cli` 来源 → `ConnectBusinessSse('agent_cli', ...)` + `RegisterBusinessReceive('agent_cli', 'agent_cli_chat_output', handler)`
  - `workflow` 来源 → `ConnectBusinessSse('task_workflow', ...)` + `RegisterBusinessReceive('task_workflow', 'task_workflow_chat_output', handler)`
- 根据 `chat_id` 过滤，只处理匹配当前 chatId 的消息

---

## 涉及文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/app/dtool/define/sse.go` | MODIFY | 新增 SseAgentCliChatOutput、SseTaskWorkflowChatOutput 常量 |
| `internal/app/dtool/controller/task_workflow.go` | MODIFY | 核心改造：移除 chatSseConns 及相关函数；移除 StreamOpen/Close；新增 agentCliSseConns/taskWorkflowSseConns 及 AgentCliChatSseOpen/Close、TaskWorkflowChatSseOpen/Close；新增 broadcastChatLineToBusinessSse、startChatCommand；修改 sendLine 调用；修改 Send/Continue 启动命令；移除 isChatStreamSseClient 过滤；移除 taskWorkflowAutoMarkChatReadIfSseConnected |
| `internal/app/dtool/controller/api_sse.go` | MODIFY | 移除 isChatStreamSseClient 函数，移除广播中的过滤 |
| `internal/app/dtool/controller/memory_fragment.go` | MODIFY | 移除广播中的 isChatStreamSseClient 过滤 |
| `internal/app/dtool/controller/browser_port_pool.go` | MODIFY | 移除广播中的 isChatStreamSseClient 过滤 |
| `internal/app/dtool/controller/workflow_unread_sse.go` | MODIFY | 移除广播中的 isChatStreamSseClient 过滤 |
| `internal/app/dtool/router.go` | MODIFY | 移除 /api/task/workflow/chat/stream SSE 路由；新增 /sse/agent_cli 和 /sse/task_workflow 两条业务 SSE 路由 |
| `web/src/utils/base/sse_business.js` | NEW | 新增业务 SSE 连接管理工具，封装 AgentCli 和 TaskWorkflow 独立 SSE 连接的创建、分发、关闭 |
| `web/src/components/agent_cli/AgentCliList.vue` | MODIFY | 核心改造：移除全部 chat SSE 方法；改用 sse_business 连接和回调注册 |
| `web/src/components/TaskWorkflow.vue` | MODIFY | 核心改造：同 AgentCliList 模式，使用 task_workflow 业务 SSE 连接 |
| `web/src/components/ChatReplyPage.vue` | MODIFY | 移除独立 EventSource，根据来源连接对应业务 SSE 路由 |

---

## 实现注意事项

- 后端 `broadcastChatLineToBusinessSse` 需要查询 chat 的 `from_type`，频繁 DB 查询可能成为瓶颈，**建议在 goroutine 启动时缓存 from_type 到局部变量**
- 前端 SSE 回调中 `chat_id` 过滤逻辑必须覆盖所有事件类型（普通输出行、completed、error），避免遗漏
- `chatCancelFuncs` 保留不动，它只管 goroutine 生命周期（停止/等待退出），与 SSE 连接无关
- 旧的 `web/src/utils/base/sse.js` 暂不清理，仍有非 chat 组件使用它
- `SseTaskWorkflowChatPrefix` 常量可保留但不再用于 SSE 连接 ID 前缀
- 两条业务 SSE 路由同样受 SSE 端口限制（`IsSsePort`），与全局 SSE 保持一致
- 业务 SSE 连接按 `clientID` 维度管理，同一个 clientID 只有一条连接，新连接会替换旧连接
- 前端 `sse_business.js` 中的端口查询逻辑复用全局 SSE 的 `fetchAvailableSsePort` 接口
