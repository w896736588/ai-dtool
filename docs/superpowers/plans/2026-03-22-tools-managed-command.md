# Tools Managed Command Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在小工具的“常用操作”中新增一个通用命令托管卡片，默认托管 `cc-connect`，支持进入页面自动确保运行、启动、关闭、重启、配置变更自动重启，以及按天落日志并在页面实时查看。

**Architecture:** 后端新增一组“托管命令”接口，统一处理命令配置校验、进程探测、启动/停止/重启、日志文件路径生成与日志尾部读取；前端在 `CommonActions.vue` 中新增独立卡片，默认填充 `cc-connect` 配置，页面加载时自动调用 ensure 接口，并在配置变更后自动触发重启与日志刷新。日志按 key + 日期写入 `logs` 目录，实时查看采用轮询最新日志尾部的方式，不引入新的 SSE 通道。

**Tech Stack:** Go, Gin, Vue 3, Element Plus

---

## Chunk 1: 后端托管命令能力

### Task 1: 先写托管命令核心失败测试

**Files:**
- Create: `internal/app/dtool/controller/tool_managed_process_test.go`
- Modify: `internal/app/dtool/router.go`
- Test: `internal/app/dtool/controller/tool_managed_process_test.go`

- [ ] **Step 1: 写配置归一化失败测试**

覆盖：
- 空命令报错
- 未传 key 时自动生成稳定 key
- 带引号的命令行可正确拆分

- [ ] **Step 2: 写 ensure 不重复启动测试**

覆盖：
- 已存在同命令进程时，`EnsureRunning` 返回运行中且不再次启动

- [ ] **Step 3: 写重启行为测试**

覆盖：
- `Restart` 会先结束当前 pid，再按新配置启动
- 新日志文件路径按 `logs/<key>-YYYY-MM-DD.log` 生成

- [ ] **Step 4: 写日志尾部读取测试**

覆盖：
- 日志文件不存在时返回空内容
- 文件超长时只返回尾部内容

- [ ] **Step 5: 运行测试并确认失败**

Run: `go test ./internal/app/dtool/controller -run TestToolManagedProcess`

Expected: FAIL，提示托管命令能力尚未实现

### Task 2: 实现最小后端能力与接口

**Files:**
- Create: `internal/app/dtool/controller/tool_managed_process.go`
- Modify: `internal/app/dtool/router.go`
- Modify: `internal/app/dtool/controller/tool_process.go`
- Test: `internal/app/dtool/controller/tool_managed_process_test.go`

- [ ] **Step 1: 实现配置归一化与命令拆分**

要求：
- 接收 `key`、`name`、`command_line`、`workdir`
- 保证 key 可安全用于日志文件名
- 命令行支持简单引号

- [ ] **Step 2: 实现进程探测与避免重复启动**

要求：
- 优先复用同 key 已托管的进程状态
- 若当前会话未托管，但系统已有同命令进程，也视为已运行

- [ ] **Step 3: 实现启动 / 停止 / 重启**

要求：
- 启动成功后记录 pid、时间、日志路径
- 停止使用 pid 精确结束
- 重启先停后起

- [ ] **Step 4: 实现日志按天落盘与尾部读取**

要求：
- 文件路径落到项目 `logs` 目录
- 页面可轮询读取最新尾部内容

- [ ] **Step 5: 注册接口**

新增：
- `POST /api/ToolManagedProcessStatus`
- `POST /api/ToolManagedProcessEnsureRunning`
- `POST /api/ToolManagedProcessStart`
- `POST /api/ToolManagedProcessStop`
- `POST /api/ToolManagedProcessRestart`
- `POST /api/ToolManagedProcessLogTail`

- [ ] **Step 6: 运行测试并确认通过**

Run: `go test ./internal/app/dtool/controller -run TestToolManagedProcess`

Expected: PASS

## Chunk 2: 前端通用命令托管卡片

### Task 3: 接入卡片、默认 cc-connect 配置与自动重启

**Files:**
- Modify: `web/src/components/tools/CommonActions.vue`
- Modify: `web/src/utils/base/tools.js`

- [ ] **Step 1: 新增命令托管卡片**

要求：
- 默认展示 `cc-connect`
- 可编辑名称、key、启动命令、工作目录

- [ ] **Step 2: 页面进入时自动 ensure**

要求：
- 若未运行则启动
- 若已运行则只展示状态，不重复启动

- [ ] **Step 3: 接入启动 / 关闭 / 重启按钮**

要求：
- 按钮状态与请求 loading 清晰
- 重启完成后自动刷新状态和日志

- [ ] **Step 4: 配置变更自动重启**

要求：
- 使用输入项的 `change` 事件，避免每次按键都重启
- 配置更新后持久化到本地存储

- [ ] **Step 5: 增加实时日志区**

要求：
- 定时轮询日志尾部
- 展示当前日志文件名
- 日志自动滚到底部

## Chunk 3: 验证

### Task 4: 运行验证并记录风险

**Files:**
- Modify: `docs\superpowers\specs\2026-03-21-tools-common-actions-design.md`

- [ ] **Step 1: 运行后端测试**

Run: `go test ./internal/app/dtool/controller`

- [ ] **Step 2: 运行前端构建验证**

Run: `npm run prod`
Workdir: `web`

- [ ] **Step 3: 手工验证核心流程**

验证：
- 进入“常用操作”页面会自动确保 `cc-connect` 运行
- 已运行时不会重复启动
- 修改命令配置后会自动重启
- 启动 / 关闭 / 重启按钮工作正常
- 页面能实时看到当天日志

- [ ] **Step 4: 记录剩余风险**

若当前仅在 Windows 验证，需要在结果中明确说明其他平台的进程探测与命令匹配仍有回归缺口
