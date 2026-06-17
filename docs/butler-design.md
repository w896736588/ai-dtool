# AI 管家（dtool-butler）设计与实施计划

> 基于流式机器人启动、能自进化的 dtool 智能管家。
> 创建时间：2026-06-17

---

## 一、已确认决策

| # | 决策点 | 结论 |
|---|--------|------|
| 1 | 子管家执行层 | 可选，关联 dtool 现有 `tbl_agent_cli` 配置；简单文件操作走 Function Calling（直接调脚本），复杂开发走 Agent CLI |
| 2 | 机器人平台 | 先仅支持钉钉，预留飞书/企微扩展位 |
| 3 | 消息收发 | **钉钉 Stream 模式（WebSocket 长连接）接收**，无需公网 IP；发送用 webhook 主动 POST |
| 4 | 索引文档 | 用固定 md 文件，**放记忆库目录**（`{memoryDbPath}/butler/index/`，便于 Git 自动同步） |
| 5 | 工程结构 | 在 `cmd/dtool-butler/` 下新起独立项目；配置走 dtool 项目；共用一个 SQLite |
| 6 | migration | **管家表 SQL 放管家项目自己的 database 目录，管家启动时执行**（与 dtool migration 隔离，避免两进程冲突） |
| 7 | 索引文档位置 | `{memoryDbPath}/butler/index/` 下 3 个 md（capabilities.md / scripts.md / apis.md） |

---

## 二、技术现实说明（钉钉 Stream）

- 钉钉机器人接收消息有两种官方模式：**HTTP 模式**（需公网回调地址，即 webhook）和 **Stream 模式**（WebSocket 长连接，无需公网）。
- "长轮询、不用 webhook 回调"的诉求与 Stream 模式意图一致：主动维持长连接、无需公网 IP、被动接收消息。
- Go 官方 SDK：`github.com/open-dingtalk/dingtalk-stream-sdk-go` v0.9.1，已加入 go.mod。
- 关键 API：
  - `client.NewStreamClient(client.WithAppCredential(client.NewAppCredentialConfig(appKey, appSecret)))`
  - `cli.RegisterChatBotCallbackRouter(handler)`，handler 签名 `func(ctx, *chatbot.BotCallbackDataModel) ([]byte, error)`
  - `cli.Start(ctx)` 非阻塞，内部起 goroutine 维持长连接，自动重连
- 消息回复：机器人消息携带 `SessionWebhook`（临时有效），用 `chatbot.NewChatbotReplier().SimpleReplyText(ctx, sessionWebhook, content)` 直接回复，无需单独配置发送地址。
- 主动推送（打招呼/休眠通知等无 incoming 场景）：通过群机器人 webhook_url 主动 POST（加签），复用 dtool `webhook_notify.go` 思路。

---

## 三、整体架构

```
┌──────────────────────────────────────────────────────────────┐
│  钉钉（用户在群里/单聊发消息）                                │
└──────────────────────────────────────────────────────────────┘
              ↑ Stream 长连接接收            ↓ webhook 主动 POST 发送
┌──────────────────────────────────────────────────────────────┐
│  cmd/dtool-butler  （独立进程，共用 dtool.db）                 │
│                                                                │
│  ┌─ bot/        钉钉 Stream 网关（收发 + 预留多平台）          │
│  ├─ butler/     管家核心（角色/激活态/命令/意图/历史）          │
│  ├─ worker/     子管家调度（FC 工具循环 / Agent CLI / 验收）   │
│  ├─ index/      索引文档（固定 md，生成/检索/自进化回写）       │
│  └─ define/     类型与常量                                     │
│                                                                │
│  复用 dev_tool/internal/app/dtool/{common,define,component}    │
│  复用 dev_tool/internal/pkg/{p_db,p_claude,p_codex,p_common}   │
└──────────────────────────────────────────────────────────────┘
              ↑ 读配置(共用库)                 ↑ 必要时 HTTP 互调
┌──────────────────────────────────────────────────────────────┐
│  cmd/dtool  （原有进程）                                       │
│  - 配置管理：tbl_butler_bot_config / tbl_butler_role /         │
│    tbl_butler_config / tbl_ai_model / tbl_agent_cli (Web CRUD) │
└──────────────────────────────────────────────────────────────┘
```

### 复用与边界
- **新项目复用 dtool 代码包**（同 go module `dev_tool`，直接 import），不重复造轮子。
- **配置管理在 dtool**：dtool 新增 `controller/set_butler.go` + 前端配置页，CRUD 管家配置表。管家项目只读这些表。
- **migration 隔离**：管家表 SQL 放管家项目 `internal/app/dtool-butler/database/`，管家启动时执行，记录表为 `tbl_butler_database_up`，与 dtool 的 `tbl_database_up` 完全隔离，避免两进程重复 migration 冲突。
- **管家与 dtool 互调**：默认走共用库读取配置；运行时若管家需触发 dtool 动作（如执行某接口），走 dtool HTTP API。

---

## 四、项目目录结构

```
cmd/dtool-butler/
  main.go                        # 入口：InitEnv → InitSqlite → RunMigration → 加载配置 → 启动网关 → 启动管家 → 阻塞等待
internal/app/dtool-butler/
  config.go                      # 读 dtool config.ini，连同一 SQLite，执行管家 migration
  component.go                   # 管家全局实例（Env、DbMain、BotGateway、ButlerCore）
  define/
    butler.go                    # 类型/常量/状态枚举
  database/                      # 管家自己的 migration SQL（按年月组织）
    2026/06/20260617100000_butler_init.sql
  bot/
    gateway.go                   # 统一网关接口（Gateway），预留多平台
    dingtalk_gateway.go          # 钉钉 Stream 接收 + webhook 发送
  butler/
    core.go                      # 管家主循环：打招呼/消费消息/固定回复/休眠巡检
    session.go                   # 会话/激活态管理 + 30min 休眠回收
    history.go                   # 历史对话存储/查询/清理
    role.go                      # 角色系统：加载 persona/tone/system_prompt（Phase 2）
    command.go                   # 内置命令：clean/init/status/help（Phase 2）
    intent.go                    # 意图分析 + 自动追问（Phase 3）
  worker/
    dispatcher.go                # 任务拆解 → 路由 FC 或 Agent CLI（Phase 4/5）
    tools.go                     # 基础文件工具注册（read/write/modify/delete）（Phase 4）
    fc_loop.go                   # Function Calling 工具循环（Phase 4）
    agent_cli.go                 # 复用 p_claude/p_codex 执行复杂任务（Phase 5）
    verify.go                    # 监督验收（Phase 4）
  index/
    doc.go                       # 索引 md 文件读写（Phase 6）
    init.go                      # init 命令：扫描 skills/ 生成索引（Phase 6）
    retrieve.go                  # 检索匹配（Phase 6）
    evolve.go                    # 自进化：新脚本回写索引（Phase 6）
```

---

## 五、数据库表设计（建在共用 dtool.db，migration 由管家执行）

### 5.1 配置类（dtool Web 管理，管家只读）
```sql
-- 钉钉机器人配置（Stream 模式所需）
CREATE TABLE IF NOT EXISTS "tbl_butler_bot_config" (
  "id"           INTEGER PRIMARY KEY AUTOINCREMENT,
  "platform"     TEXT NOT NULL DEFAULT 'dingtalk',   -- 预留 feishu/wecom
  "name"         TEXT NOT NULL DEFAULT '',
  "app_key"      TEXT NOT NULL DEFAULT '',            -- 钉钉应用 AppKey
  "app_secret"   TEXT NOT NULL DEFAULT '',            -- 钉钉应用 AppSecret
  "robot_code"   TEXT NOT NULL DEFAULT '',            -- 机器人 robotCode
  "webhook_url"  TEXT NOT NULL DEFAULT '',            -- 发送用 webhook（主动推送）
  "secret"       TEXT NOT NULL DEFAULT '',            -- 发送加签
  "status"       INTEGER NOT NULL DEFAULT 1,
  "created_at"   INTEGER NOT NULL DEFAULT 0,
  "updated_at"   INTEGER NOT NULL DEFAULT 0
);

-- 管家角色
CREATE TABLE IF NOT EXISTS "tbl_butler_role" (
  "id"             INTEGER PRIMARY KEY AUTOINCREMENT,
  "name"           TEXT NOT NULL DEFAULT '',
  "persona"        TEXT NOT NULL DEFAULT '',   -- 定位（如"严谨的技术管家"）
  "tone"           TEXT NOT NULL DEFAULT '',   -- 语气（如"简洁专业"）
  "system_prompt"  TEXT NOT NULL DEFAULT '',   -- 完整 system prompt
  "init_greeting"  TEXT NOT NULL DEFAULT '',   -- 启动打招呼语
  "status"         INTEGER NOT NULL DEFAULT 1,
  "created_at"     INTEGER NOT NULL DEFAULT 0,
  "updated_at"     INTEGER NOT NULL DEFAULT 0
);

-- 管家运行参数
CREATE TABLE IF NOT EXISTS "tbl_butler_config" (
  "id"                       INTEGER PRIMARY KEY AUTOINCREMENT,
  "name"                     TEXT NOT NULL DEFAULT '',
  "role_id"                  INTEGER NOT NULL DEFAULT 0,   -- 关联角色
  "model_id"                 INTEGER NOT NULL DEFAULT 0,   -- 管家主模型（tbl_ai_model）
  "fc_model_id"              INTEGER NOT NULL DEFAULT 0,   -- Function Calling 用模型
  "agent_cli_id"             INTEGER NOT NULL DEFAULT 0,   -- 可选：复杂任务用的 AgentCli
  "bot_config_id"            INTEGER NOT NULL DEFAULT 0,   -- 关联机器人配置
  "active_timeout_minutes"   INTEGER NOT NULL DEFAULT 30,  -- 激活态超时
  "max_history"              INTEGER NOT NULL DEFAULT 100, -- 历史上限
  "auto_clean_on_new_topic"  INTEGER NOT NULL DEFAULT 1,   -- 新问题自动清历史
  "index_doc_path"           TEXT NOT NULL DEFAULT '',     -- 索引 md 目录（留空用默认）
  "auto_init_on_start"       INTEGER NOT NULL DEFAULT 1,   -- 启动自动 init 索引
  "status"                   INTEGER NOT NULL DEFAULT 1,
  "created_at"               INTEGER NOT NULL DEFAULT 0,
  "updated_at"               INTEGER NOT NULL DEFAULT 0
);
```

### 5.2 运行时类（管家读写）
```sql
-- 会话历史
CREATE TABLE IF NOT EXISTS "tbl_butler_message" (
  "id"          INTEGER PRIMARY KEY AUTOINCREMENT,
  "session_id"  TEXT NOT NULL DEFAULT '',        -- 会话标识（如钉钉 conversationId）
  "role"        TEXT NOT NULL DEFAULT '',        -- user/assistant/system
  "content"     TEXT NOT NULL DEFAULT '',
  "token_count" INTEGER NOT NULL DEFAULT 0,
  "topic"       TEXT NOT NULL DEFAULT '',        -- 当前主题（用于新问题判定）
  "created_at"  INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS "idx_butler_msg_session" ON "tbl_butler_message"("session_id", "id");

-- 管家任务记录（监督/执行/验收）
CREATE TABLE IF NOT EXISTS "tbl_butler_task" (
  "id"          INTEGER PRIMARY KEY AUTOINCREMENT,
  "session_id"  TEXT NOT NULL DEFAULT '',
  "title"       TEXT NOT NULL DEFAULT '',
  "status"      TEXT NOT NULL DEFAULT 'pending', -- pending/executing/verifying/done/failed
  "plan"        TEXT NOT NULL DEFAULT '',        -- 任务拆解
  "result"      TEXT NOT NULL DEFAULT '',        -- 验收结果
  "executor"    TEXT NOT NULL DEFAULT '',        -- fc/agent_cli
  "created_at"  INTEGER NOT NULL DEFAULT 0,
  "updated_at"  INTEGER NOT NULL DEFAULT 0
);
```

---

## 六、索引文档设计（固定 md，放记忆库目录）

位置：`{config.ini: base.memoryDbPath}/butler/index/`（如 `C:\work\self\dev_tool_db\memory\butler\index\`）。
管家启动时从 config.ini 读 `base.memoryDbPath` 定位，复用记忆库的 Git 自动同步能力做版本管理。

```
{memoryDbPath}/butler/index/
  capabilities.md     # 总能力清单（管家能做什么）
  scripts.md          # 脚本工具索引（skills/ 下脚本 + 自进化新脚本）
  apis.md             # dtool 可用 HTTP 接口索引
```

### scripts.md 结构
```markdown
# 脚本工具索引

## [db_api] 数据库查询
- 路径: skills/dtool-db/scripts/db_api.py
- 用途: 查询/验证数据库字段、索引
- 入参: table_name, conditions
- 出参: JSON 行数据
- 示例: python db_api.py --table tbl_user

## [git_diff] 获取分支改动
- 路径: skills/dtool-common/scripts/xxx.py
- 用途: ...
```

### 自进化流程
1. 任务来 → `retrieve.go` 把"任务 + scripts.md 摘要"喂 AI，判断是否有现成脚本
2. 命中 → 直接调用
3. 未命中 → 子管家编写新脚本 → 执行 → `evolve.go` 追加条目到 scripts.md

---

## 七、分阶段实施计划

### Phase 1：钉钉双向通信 + 管家骨架（最小闭环） ✅ 代码已完成
**目标**：启动管家 → 自动发打招呼 → 用户发消息 → 管家回复（先不接 AI，固定回复）→ 30min 无消息休眠
**检查点**：钉钉发消息管家能收到并回复；启动自动发打招呼；30min 无消息自动休眠并通知

| 子任务 | 状态 | 说明 |
|--------|------|------|
| P1-1 工程骨架 + 共用库连接 | ✅ | `main.go` + `config.go` + migration，已验证建表成功 |
| P1-2 钉钉 Stream 接收 | ✅ | `dingtalk_gateway.go` 接入 SDK，代码已完成待凭证联调 |
| P1-3 钉钉发送 | ✅ | webhook POST 发送，代码已完成待凭证联调 |
| P1-4 管家主循环 + 激活态 | ✅ | `core.go` + `session.go`，打招呼/消费/休眠巡检 |
| P1-5 历史消息存储 | ✅ | `history.go`，消息存 `tbl_butler_message` |

**Phase 1 关键发现（踩坑记录）**：
- `gsdb` 的 `QueryBySql(...).One()` 会自动追加 `LIMIT 1`，SQL 里不能再自带 `LIMIT`，否则产生 `... LIMIT 1 LIMIT 1` 语法错误。使用 `.One()` 时 SQL 不要写 `LIMIT`；需要自定义 limit 时用 `.All()` + SQL 自带 `LIMIT`。
- 钉钉 SDK 回复机制：消息携带 `SessionWebhook`（临时有效），直接用它回复，无需单独发送配置；仅主动推送（打招呼/休眠通知）需群机器人 webhook_url。

**Phase 1 联调前置条件**：
1. 在 dtool 配置 `tbl_butler_bot_config`（钉钉 app_key/app_secret/webhook_url/secret）
2. 在 dtool 配置 `tbl_butler_role`（角色/打招呼语）
3. 在 dtool 配置 `tbl_butler_config`（关联 role_id/bot_config_id，设置 active_timeout_minutes 等）
4. 钉钉开放平台创建企业内部应用并开通 Stream 模式机器人

### Phase 2：角色系统 + 内置命令 + 历史管理
**目标**：管家有"人格"，支持 clean/init/status 命令，历史存库
- [ ] `role.go` 加载角色 persona/tone，拼装 system_prompt
- [ ] `command.go` 解析 clean/init/status/help
- [ ] `history.go` clean 清当前 session
- [ ] 管家回复接 `AIChatStreamByModel`（流式，简洁约束）
- [ ] 验证：clean 清历史，角色语气生效

### Phase 3：意图分析 + 自动追问
**目标**：模糊问题返回 2-3 个澄清提问，明确则进入任务
- [ ] `intent.go` 轻量 AI Chat 判断意图清晰度 + 新问题判定
- [ ] 新问题 + `auto_clean_on_new_topic` → 自动清历史
- [ ] 历史超 `max_history` → 提示是否清理
- [ ] 验证：模糊问题自动追问，新主题自动清历史

### Phase 4：子管家执行层（基础文件工具 + Function Calling）
**目标**：用户发文件操作任务 → FC 工具循环执行 → 汇报
- [ ] `worker/tools.go` 注册 file_read/write/modify/delete（调 dtool-common skill 脚本）
- [ ] `worker/fc_loop.go` AI Chat + tools 循环
- [ ] `worker/verify.go` 验收产出
- [ ] 验证：用户说"在 xx 创建文件 yy"，管家完成并汇报

### Phase 5：Agent CLI 复杂任务执行
**目标**：复杂开发任务走 `tbl_butler_config.agent_cli_id` 指定的 AgentCli
- [ ] `worker/agent_cli.go` 复用 `p_claude.RunClaudeStream`/`p_codex.RunCodexStream`
- [ ] `worker/dispatcher.go` 路由：简单→FC，复杂→Agent CLI
- [ ] 验证：用户发"开发一个 xx 接口"，管家起 AgentCli 执行并汇报

### Phase 6：索引文档 + 自进化
**目标**：init 生成索引，任务前检索，新脚本回写
- [ ] `index/init.go` 扫描 `skills/` 生成 scripts.md
- [ ] `index/retrieve.go` 任务前检索匹配
- [ ] `index/evolve.go` 新脚本回写 scripts.md
- [ ] 验证：init 生成索引，重复任务命中已有脚本，新任务完成后索引增长

每个 Phase 完成后验证再推进下一阶段。

---

## 八、数据流（典型：用户发任务）

```
用户在钉钉发消息
  → DingTalkGateway Stream 接收 → 投递 IncomingMessage 到 msgChan
  → Butler Core 消费 → sessions.Activate 激活态 + 重置 30min 定时器
  → history.Append 存用户消息到 tbl_butler_message
  → (Phase 2+) 加载角色 system_prompt + 历史对话
  → (Phase 3+) intent.go 意图分析：
      ├─ 命中内置命令(clean/init) → 直接执行
      ├─ 意图不明确 → 返回澄清提问
      └─ 意图明确 → 任务拆解
  → (Phase 4/5) 任务分发到 Sub-Butler：
      ├─ 检索脚本索引 → 命中已有脚本 → 直接调用
      └─ 未命中 → 子管家编写新脚本 → 执行 → 回写索引
  → (Phase 4) verify.go 验收子管家产出
  → 结果汇总 → replier.SimpleReplyText 回复到钉钉
  → history.Append 存管家回复
  → (Phase 3) 历史超阈值 → 触发清理提示/自动清理
```
