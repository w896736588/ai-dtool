# Agent CLI SDK 对比分析报告

> 对比对象：`claude-agent-sdk-go`、Codex Go SDK 生态 与 当前项目 (ai-dtool) 的 Agent CLI 实现

---

## ⚠️ 重要说明：`coze-dev/coze-go/packages/codexsdk` 路径不存在

`github.com/coze-dev/coze-go` 仓库的 `main` 分支中**不存在** `packages/codexsdk` 目录。Coze Go SDK 是纯 API 客户端库，不包含 Codex CLI 相关的 SDK。

实际的 Codex CLI Go SDK 有两个主流项目：

| SDK | 通信协议 | 仓库 |
|-----|---------|------|
| **hishamkaram/codex-agent-sdk-go** | JSON-RPC 2.0 over stdio (via `codex app-server`) | 功能最全 |
| **picatz/openai/codex** | `codex exec --json` 子进程 JSON 事件流 | 轻量封装 |

下文将基于这两个 Codex SDK 进行对比。

---

## 一、三方架构总览

### 1.1 当前项目 (ai-dtool) — "子进程 + 逐行 JSON 解析"模式

```
┌──────────────┐     启动子进程      ┌─────────────────────┐
│  Controller  │ ──────────────────→ │ claude / codex CLI   │
│  (Gin HTTP)  │                     │ (独立进程)            │
│              │ ←── lineCh ─────── │ stdout → JSONL 逐行  │
│  SSE 推送    │                     │ stderr → 收集       │
└──────────────┘                     └─────────────────────┘
         │
         ↓ callback(StreamMessage)
  前端 SSE 实时推送
```

**核心特征：**
- **单向流**：启动 CLI 子进程 → 读 stdout → 解析每行 JSON → callback → SSE 推送前端
- **无双向控制**：无法在运行时向 CLI 发送指令（如权限审批、中断 turn）
- **回调模型**：`RunClaudeStream(ctx, cfg, callback)` / `RunCodexStream(ctx, cfg, callback)`
- **Claude 使用**：`claude -p --output-format stream-json --permission-mode bypassPermissions`
- **Codex 使用**：`codex exec - --json --sandbox danger-full-access`
- **会话续接**：通过 `--resume` (Claude) / `exec resume` (Codex) 重新启动进程

### 1.2 claude-agent-sdk-go — "双向控制协议"模式

```
┌──────────────┐   Connect()    ┌─────────────────────┐
│   Client     │ ─────────────→ │ Claude Code CLI     │
│   (SDK)      │                │ (Node.js 子进程)     │
│              │ ←── Messages── │ stdin/stdout         │
│  Query()     │ ── Prompt ──→  │ 双向控制协议         │
│  Receive()   │ ←── Stream ──  │ stream-json         │
│              │                │                     │
│  Permission  │ ← 请求 ─────── │ 工具权限审批         │
│  Callback    │ ── 允许/拒绝 → │                     │
│  Hooks       │ ← 事件 ─────── │ PreToolUse 等       │
│  MCP         │ ← 请求 ─────── │ MCP 服务器          │
└──────────────┘                └─────────────────────┘
```

**核心特征：**
- **双向控制协议**：可运行时发送指令、权限审批、Hook 拦截
- **Channel 模型**：`Query()` 返回 `<-chan Message`，`Client.ReceiveResponse()` 返回 `<-chan Message`
- **交互式 Client**：`NewClient()` → `Connect()` → 多次 `Query()` → `Close()`
- **权限回调**：拦截工具调用请求，程序化决定允许/拒绝
- **Hook 系统**：`PreToolUse`、`PostToolUse` 等生命周期事件
- **MCP 支持**：内建 MCP 服务器创建工厂

### 1.3 hishamkaram/codex-agent-sdk-go — "JSON-RPC 2.0 app-server"模式

```
┌──────────────┐   NewClient()  ┌─────────────────────┐
│   Client     │ ─────────────→ │ codex app-server    │
│   (SDK)      │                │ (Rust 子进程)        │
│              │ ←─ JSON-RPC ── │ stdin/stdout        │
│  Thread      │ ── Request ──→ │ 双向 JSON-RPC 2.0   │
│  RunStreamed │ ←─ Events ─── │ 多路复用通知         │
│              │                │                     │
│  Approval    │ ← 请求 ─────── │ 审批回调            │
│  Callback    │ ── 接受/拒绝 → │                     │
│  Hooks       │ ← 事件 ─────── │ HookStarted 等      │
│  MCP Config  │ ── 配置 ──────→ │ MCP 服务器配置      │
└──────────────┘                └─────────────────────┘
```

**核心特征：**
- **JSON-RPC 2.0 协议**：通过 `codex app-server` 子进程，使用 stdio 传输 JSON-RPC 2.0
- **多路复用解调**：三种消息形态（响应、通知、服务端发起请求）统一处理
- **Thread 抽象**：`StartThread` / `Resume` / `Fork` / `Archive` / `List`
- **双执行模式**：`Run()`（缓冲）/ `RunStreamed()`（`<-chan Event` 通道流式）
- **审批回调**：服务端发起审批请求 → 用户回调决定接受/拒绝
- **Hook 系统**：观察者模式 + 编程式回调（自动写入 hooks.json）
- **MCP 配置**：stdio + streamable HTTP 两种 MCP 传输
- **结构化输出**：JSON-schema 约束
- **Slash 命令等价方法**：Compact, SetModel, GitDiff 等

### 1.4 picatz/openai/codex — "CLI exec 封装"模式

```
┌──────────────┐   NewExec()    ┌─────────────────────┐
│   Thread     │ ─────────────→ │ codex exec --json   │
│   (SDK)      │                │ (子进程)             │
│              │ ←─ Events ──── │ stdout JSON 事件流   │
│  Run()       │                │                     │
│  RunStreamed │ ←─ chan ────── │ EventStream         │
└──────────────┘                └─────────────────────┘
```

**核心特征：**
- **CLI 子进程封装**：底层调用 `codex exec` 命令，解析 JSON 事件流
- **Thread 抽象**：支持多轮对话和线程续接
- **Go 1.23 迭代器**：`EventStream()` 返回 `iter.Seq2[*ThreadEvent, error]`
- **类型化事件**：AgentMessage, Reasoning, CommandExecution, FileChange, McpToolCall 等
- **沙箱/审批模式**：与 CLI 参数一一对应

---

## 二、功能对比矩阵

| 功能维度 | ai-dtool (当前项目) | claude-agent-sdk-go | hishamkaram/codex-agent-sdk-go | picatz/openai/codex |
|---------|--------------------|--------------------|------------------------------|--------------------|
| **通信协议** | CLI stdout JSONL | 双向控制协议 (stdin/stdout) | JSON-RPC 2.0 (stdin/stdout) | CLI stdout JSONL |
| **流式输出** | ✅ callback 逐行 | ✅ `<-chan Message` | ✅ `<-chan Event` | ✅ `iter.Seq2` / `<-chan ThreadEvent` |
| **交互式多轮** | ⚠️ 重启进程续接 | ✅ Client 持久会话 | ✅ Thread 持久会话 | ✅ Thread 持久会话 |
| **权限审批** | ❌ bypassPermissions | ✅ PermissionCallback | ✅ ApprovalCallback | ❌ 静态参数 |
| **Hook 系统** | ❌ 无 | ✅ PreToolUse/PostToolUse | ✅ HookStarted/Completed + 编程式 | ❌ 无 |
| **MCP 集成** | ⚠️ 仅配置文件 | ✅ 内建 MCP Server 工厂 | ✅ stdio + HTTP MCP 配置 | ❌ 无 |
| **工具调用拦截** | ❌ 无 | ✅ 运行时拦截 | ✅ 运行时拦截 | ❌ 无 |
| **结构化输出** | ❌ 无 | ❌ 无 | ✅ JSON-schema | ✅ OutputSchema |
| **会话管理** | ⚠️ sessionID 提取 | ✅ Client 连接管理 | ✅ Thread CRUD + Resume/Fork | ✅ ThreadID 续接 |
| **Turn 中断** | ⚠️ context 取消进程 | ✅ 协议级中断 | ✅ 协议级中断 | ⚠️ context 取消进程 |
| **类型化事件** | ⚠️ map[string]any | ✅ 类型化 Message | ✅ 类型化 Event | ✅ 类型化 ThreadEvent |
| **SSE 推送** | ✅ 内建 (Web) | ❌ 需自行实现 | ❌ 需自行实现 | ❌ 需自行实现 |
| **Windows 支持** | ✅ Job Object 管理 | ⚠️ 有限 | ⚠️ 未明确 | ⚠️ 未明确 |
| **零依赖** | ❌ 依赖 spf13/cast | ✅ 核心零依赖 | ❌ 依赖 zap 等 | ❌ 依赖较少 |

---

## 三、能否完美交互式流式支持对应 CLI？

### 3.1 claude-agent-sdk-go → Claude Code CLI

| 评估项 | 结果 | 说明 |
|--------|------|------|
| 一次性查询流式 | ✅ 完美 | `Query()` → `<-chan Message`，逐消息流式消费 |
| 交互式多轮流式 | ✅ 完美 | `Client` → `Query()` / `ReceiveResponse()`，双向协议 |
| 权限审批交互 | ✅ 完美 | `PermissionCallback` 实时拦截工具调用 |
| Hook 拦截 | ✅ 完美 | `PreToolUse` / `PostToolUse` 生命周期拦截 |
| MCP 交互 | ✅ 完美 | 内建 MCP Server 工厂 |
| 流式部分输出 | ✅ 支持 | `--include-partial-messages` 等价 |

**结论：✅ 能完美支持 Claude Code CLI 的交互式流式场景。**

SDK 是 Python 官方 SDK 的完整 Go 移植，功能对齐度 100%。唯一注意点是 Windows 支持有限。

### 3.2 hishamkaram/codex-agent-sdk-go → Codex CLI

| 评估项 | 结果 | 说明 |
|--------|------|------|
| 一次性查询流式 | ✅ 完美 | `codex.Query()` → `<-chan Event` |
| 交互式多轮流式 | ✅ 完美 | `Thread.RunStreamed()` → `<-chan Event`，持久会话 |
| 权限审批交互 | ✅ 完美 | `ApprovalCallback` 实时审批工具调用 |
| Hook 拦截 | ✅ 完美 | 观察者 + 编程式回调双重模式 |
| MCP 配置 | ✅ 完美 | stdio + HTTP 两种传输 |
| 结构化输出 | ✅ 完美 | JSON-schema 约束 |
| Thread 生命周期 | ✅ 完美 | Start/Resume/Fork/Archive/List |
| Slash 命令 | ✅ 完美 | Compact/SetModel/GitDiff 等价 |

**结论：✅ 能完美支持 Codex CLI 的交互式流式场景。**

这是功能最全的 Codex Go SDK，通过 `app-server` 传输层实现了完整的双向 JSON-RPC 2.0 通信。

### 3.3 picatz/openai/codex → Codex CLI

| 评估项 | 结果 | 说明 |
|--------|------|------|
| 一次性查询流式 | ✅ 支持 | `Run()` / `RunStreamed()` |
| 交互式多轮流式 | ⚠️ 有限 | Thread 抽象支持但底层仍为 `codex exec` 子进程 |
| 权限审批交互 | ❌ 不支持 | 只能通过静态 CLI 参数配置 |
| Hook 拦截 | ❌ 不支持 | 无 Hook 系统 |
| MCP 配置 | ❌ 不支持 | 无 MCP 集成 |
| 结构化输出 | ✅ 支持 | OutputSchema |
| 事件类型化 | ✅ 完美 | 完整的 ThreadItem 类型体系 |

**结论：⚠️ 能支持基本的流式场景，但不支持交互式审批/Hook 等高级交互。**

---

## 四、与当前项目 (ai-dtool) 的核心差异

### 4.1 当前项目的局限

| 问题 | 影响 |
|------|------|
| **无双向控制** | 无法运行时审批工具调用，只能 bypassPermissions 全放行 |
| **无 Hook 拦截** | 无法在工具执行前后注入逻辑 |
| **进程级续接** | 每轮对话需重启 CLI 进程，开销大且可能丢失上下文 |
| **弱类型** | `StreamMessage.Data` 为 `map[string]any`，需手动解析 |
| **无结构化输出** | 无法约束 AI 响应格式 |
| **无 MCP 动态管理** | MCP 服务器只能通过配置文件预定义 |

### 4.2 引入 SDK 的收益

| 维度 | 收益 |
|------|------|
| **交互式审批** | 可在前端 UI 实现工具调用审批弹窗，用户点击允许/拒绝 |
| **持久会话** | 一个 Client/Thread 连接可执行多轮对话，无需重启进程 |
| **Hook 系统** | 可注入日志、限流、安全策略等中间件 |
| **类型安全** | 编译期类型检查，减少运行时解析错误 |
| **Turn 中断** | 协议级中断，而非暴力 kill 进程 |
| **结构化输出** | 可要求 AI 返回特定 JSON 格式 |

### 4.3 引入 SDK 的代价

| 代价 | 说明 |
|------|------|
| **架构变更** | 从 callback 模型迁移到 channel/迭代器模型 |
| **SSE 推送适配** | SDK 不提供 SSE，需自行桥接 channel → SSE |
| **Windows 兼容** | 两个 SDK 的 Windows 支持都有限，当前项目已有 Job Object |
| **依赖引入** | claude-agent-sdk-go 零依赖但 codex-agent-sdk-go 有 zap 等 |
| **学习成本** | 双向控制协议比单向 JSON 解析复杂 |
| **调试难度** | JSON-RPC 2.0 多路复用比 JSONL 逐行更难调试 |

---

## 五、最终结论

### 能否完美交互式流式支持对应 CLI？

| SDK | 对应 CLI | 交互式流式支持 | 完美度 |
|-----|---------|--------------|--------|
| **claude-agent-sdk-go** | Claude Code CLI | ✅ 完美 | ★★★★★ |
| **hishamkaram/codex-agent-sdk-go** | Codex CLI (app-server) | ✅ 完美 | ★★★★★ |
| **picatz/openai/codex** | Codex CLI (exec) | ⚠️ 基本流式 | ★★★☆☆ |

### 对当前项目的建议

1. **短期**：当前项目的"子进程 + 逐行 JSON"方案已满足基本流式需求，适合 MVP 阶段
2. **中期**：如需交互式审批、Hook、MCP 动态管理，应考虑引入 SDK：
   - Claude 方向：`claude-agent-sdk-go` 功能最完整
   - Codex 方向：`hishamkaram/codex-agent-sdk-go` 功能最完整（推荐）或 `picatz/openai/codex`（轻量替代）
3. **注意**：两个 SDK 都不是官方产品（claude-agent-sdk-go 是社区移植，codex-agent-sdk-go 也是第三方），需关注维护稳定性和版本兼容性
4. **迁移路径**：可渐进式替换 —— 保留当前 `p_claude`/`p_codex` 作为 fallback，新增 SDK 传输层，通过配置切换
