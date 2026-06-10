# SSE Chat Stream 重构方案

## 产品概述

彻底重构 Codex CLI 和 Claude CLI 执行任务时的 SSE 逻辑，将"一个对话绑一个 SSE 链接"的旧设计改为页面级 SSE 推送模式，消除状态错乱、重复连接、页面刷新异常、历史对话更新不到位等问题，并清理旧代码。

## 核心特性

- 进入 `/AgentCli` 页面时只复用全局 SSE 连接，注册 `agent_cli_chat_output` 分发回调，不再创建任何独立 EventSource
- 进入 `/TaskWorkflow/:taskId` 页面时只复用全局 SSE 连接，注册 `task_workflow_chat_output` 分发回调，不再创建任何独立 EventSource
- 所有对话输出统一通过全局 SSE 连接推送，消息自动携带 `chat_id`，前端按 chat_id 过滤渲染
- 前端收到对话 completed 事件，且当前执行历史弹窗打开并选中该对话时，自动调用 API 置为已读
- 后端 `TaskWorkflowChatSend`/`AgentChatSend`/`TaskWorkflowChatContinue` API 直接启动 CLI 命令 goroutine，不再依赖 SSE 连接触发
- 移除 `/api/task/workflow/chat/stream` 独立 SSE 路由及全部 chat stream 专用连接管理代码
- 移除 `isChatStreamSseClient` 过滤逻辑和 `chatSseConns` 连接池
- ChatReplyPage 同样通过全局 SSE 注册回调接收流式输出

## 技术栈

- 后端：Go (现有项目栈，gin + gsgin SSE 框架)
- 前端：Vue 3 + JavaScript (现有项目栈，Element Plus)
- SSE：复用现有 `sse_distribute` 全局单连接分发模式

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

将"每对话独立 SSE 连接"改为"页面级 SSE 分发"，后端所有 chat 输出统一推送到全局 SSE 通道（按 `agent_cli_chat_output` / `task_workflow_chat_output` 分发 ID 区分来源），前端通过 `sse_distribute.RegisterReceive` 注册回调，根据消息中的 `chat_id` 决定是否渲染到当前选中对话。

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

### 2. 替换 broadcastToChatSse 为全局 SSE 推送

**文件**: `internal/app/dtool/controller/task_workflow.go`

新增函数 `broadcastChatLineToGlobalSse(chatID int64, line string)`：

- 查询 chat 的 `from_type` 确定来源
- 将 `line` 包装为 `{ chat_id, line }` 作为 SseData.Data
- `agent_cli` 来源 → 使用 `SseAgentCliChatOutput` 分发 ID
- `workflow` 来源 → 使用 `SseTaskWorkflowChatOutput` 分发 ID
- 两种来源都推送到全局 SSE（遍历所有非空 clientID 的 SSE 连接发送）
- 移除 `isChatStreamSseClient` 过滤（不再有 chat stream 专用连接）

### 3. 修改 sendLine 逻辑

**文件**: `internal/app/dtool/controller/task_workflow.go`

在 `runClaudeCommand` 和 `runCodexCommand` 中：

- `sendLine(line)` → `broadcastChatLineToGlobalSse(chatID, line)` 替代 `broadcastToChatSse(chatID, line)`
- DB 写入通道逻辑不变
- 在 goroutine 启动时缓存 `fromType` 到局部变量，避免每次 sendLine 都查 DB

### 4. 移除 chat stream SSE 路由和连接管理

**涉及文件**:

- `internal/app/dtool/router.go` — 移除 `/api/task/workflow/chat/stream` 的 SseRoute 注册
- `internal/app/dtool/controller/task_workflow.go` — 删除以下全局变量和函数：
  - `chatSseConns` 全局变量
  - `addChatSseConn` / `removeChatSseConn` / `broadcastToChatSse` / `chatHasActiveSse`
  - `TaskWorkflowChatStreamOpen` / `TaskWorkflowChatStreamClose`
  - `taskWorkflowAutoMarkChatReadIfSseConnected`

### 5. 移除 isChatStreamSseClient 过滤逻辑

**涉及文件**:

- `internal/app/dtool/controller/api_sse.go` — 删除 `isChatStreamSseClient` 函数定义，移除广播中的 `isChatStreamSseClient` 过滤
- `internal/app/dtool/controller/memory_fragment.go` — 移除广播中的 `isChatStreamSseClient` 过滤
- `internal/app/dtool/controller/browser_port_pool.go` — 移除广播中的 `isChatStreamSseClient` 过滤
- `internal/app/dtool/controller/workflow_unread_sse.go` — 移除广播中的 `isChatStreamSseClient` 过滤
- `internal/app/dtool/controller/task_workflow.go` — 移除广播中的 `isChatStreamSseClient` 过滤

### 6. 将命令启动逻辑迁移到 Send/Continue API

**文件**: `internal/app/dtool/controller/task_workflow.go`

提取 `TaskWorkflowChatStreamOpen` 中的 goroutine 启动逻辑为独立函数 `startChatCommand(chatID int64)`：

- 读取 chatInfo 获取 localDir、cliType、settingsPath、sessionID 等配置
- 根据 cliType 启动 `runClaudeCommand` 或 `runCodexCommand`
- 在 `TaskWorkflowChatSend` 创建 DB 记录后调用 `startChatCommand`
- 在 `AgentChatSend` 创建 DB 记录后调用 `startChatCommand`
- 在 `TaskWorkflowChatContinue` 标记 running 后调用 `startChatCommand`

### 7. 孤立 running 状态检测

**文件**: `internal/app/dtool/controller/task_workflow.go`

在 `TaskWorkflowChatDetail` API 中增加检测：当 DB 状态为 running 但 goroutine 不存在时，标记为 interrupted 并广播状态变更。替代旧 SSE 连接重连时的检测逻辑。

---

## 前端改造

### 1. AgentCliList.vue

**文件**: `web/src/components/agent_cli/AgentCliList.vue`

**删除**:

- `connectChatStream` 方法
- `_chatEventSource` / `_sseChatId` 状态
- `_backgroundChatEventSources` 状态
- `startBackgroundChatStream` / `stopBackgroundChatStream` / `stopAllBackgroundChatStreams` / `updateBackgroundChatListItem` 方法
- `chatDetailSSERegistered` 状态
- `_initialSseRetryCount` 状态

**新增**:

- `registerChatOutputSse()` / `unregisterChatOutputSse()`：
  - `mounted` 时注册 `sse_distribute.RegisterReceive('agent_cli_chat_output', handler)`
  - handler 中根据 `chat_id` 判断：若等于当前选中 chatDetailId → 追加到 `chatDetailSSELines` + 解析渲染；否则仅更新列表状态
  - `beforeUnmount` 时 `UnRegisterReceive`

- completed 事件处理：收到 completed 且 chat_id 匹配当前选中对话 → `loadChatDetail()` + 刷新计数；若执行历史弹窗打开且选中该对话 → 调用 `AgentChatMarkRead`

**修改**:

- 发送/继续对话流程：调用 API 后不再调用 `connectChatStream`，改为设置 chatDetailId + 状态即可，SSE 回调自动接收输出
- `stopChat()`: 移除 `_chatEventSource.close()` 逻辑，仅调用后端 stop API

### 2. TaskWorkflow.vue

**文件**: `web/src/components/TaskWorkflow.vue`

同 AgentCliList.vue 模式，注册 `task_workflow_chat_output` 分发 ID，删除所有 chat stream 相关方法，修改发送/继续/停止对话流程。

### 3. ChatReplyPage.vue

**文件**: `web/src/components/ChatReplyPage.vue`

- 删除 `connectChatStream` / `_chatEventSource` / `closeSSE`
- 同时注册 `agent_cli_chat_output` 和 `task_workflow_chat_output` 两个回调
- 根据 `chat_id` 过滤，只处理匹配当前 chatId 的消息

---

## 涉及文件清单

| 文件 | 操作 | 说明 |
|------|------|------|
| `internal/app/dtool/define/sse.go` | MODIFY | 新增 SseAgentCliChatOutput、SseTaskWorkflowChatOutput 常量 |
| `internal/app/dtool/controller/task_workflow.go` | MODIFY | 核心改造：移除 chatSseConns 及相关函数；移除 StreamOpen/Close；新增 broadcastChatLineToGlobalSse、startChatCommand；修改 sendLine 调用；修改 Send/Continue 启动命令；移除 isChatStreamSseClient 过滤；移除 taskWorkflowAutoMarkChatReadIfSseConnected |
| `internal/app/dtool/controller/api_sse.go` | MODIFY | 移除 isChatStreamSseClient 函数，移除广播中的过滤 |
| `internal/app/dtool/controller/memory_fragment.go` | MODIFY | 移除广播中的 isChatStreamSseClient 过滤 |
| `internal/app/dtool/controller/browser_port_pool.go` | MODIFY | 移除广播中的 isChatStreamSseClient 过滤 |
| `internal/app/dtool/controller/workflow_unread_sse.go` | MODIFY | 移除广播中的 isChatStreamSseClient 过滤 |
| `internal/app/dtool/router.go` | MODIFY | 移除 /api/task/workflow/chat/stream SSE 路由 |
| `web/src/components/agent_cli/AgentCliList.vue` | MODIFY | 核心改造：移除全部 chat SSE 方法；新增 registerChatOutputSse |
| `web/src/components/TaskWorkflow.vue` | MODIFY | 核心改造：同 AgentCliList 模式 |
| `web/src/components/ChatReplyPage.vue` | MODIFY | 移除独立 EventSource，注册双分发 ID 回调 |

---

## 实现注意事项

- 后端 `broadcastChatLineToGlobalSse` 需要查询 chat 的 `from_type`，频繁 DB 查询可能成为瓶颈，**建议在 goroutine 启动时缓存 from_type 到局部变量**
- 前端 SSE 回调中 `chat_id` 过滤逻辑必须覆盖所有事件类型（普通输出行、completed、error），避免遗漏
- `chatCancelFuncs` 保留不动，它只管 goroutine 生命周期（停止/等待退出），与 SSE 连接无关
- 旧的 `web/src/utils/base/sse.js` 暂不清理，仍有非 chat 组件使用它
- `SseTaskWorkflowChatPrefix` 常量可保留但不再用于 SSE 连接 ID 前缀
