# Tools Common Actions Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在小工具中新增一个可扩展的“常用操作”面板，并实现“按端口查询占用进程并确认结束”的跨平台能力。

**Architecture:** 前端在 [Tools.vue](C:/work/frog/dev_tool_master/web/src/components/Tools.vue) 新增 `常用操作` tab，并以独立组件承载可扩展动作卡片；后端新增两个接口，分别负责按端口查询进程和按 PID 结束进程。跨平台差异通过后端内部的命令构造与输出解析层处理，控制器保持薄层。

**Tech Stack:** Go, Gin, Vue 3, Element Plus

---

## Chunk 1: 后端端口进程能力

### Task 1: 为端口查询与结束能力写失败测试

**Files:**
- Create: `internal/app/dtool/controller/tool_process_test.go`
- Modify: `internal/app/dtool/router.go`
- Test: `internal/app/dtool/controller/tool_process_test.go`

- [ ] **Step 1: 写端口校验失败测试**

覆盖：
- 非数字端口报错
- 端口超出 `1-65535` 报错

- [ ] **Step 2: 写 Windows 查询输出解析失败测试**

覆盖：
- `netstat -ano -p tcp` 输出可解析出 PID、协议、地址
- 没有匹配端口时返回空列表

- [ ] **Step 3: 写 Unix 查询输出解析失败测试**

覆盖：
- `lsof -nP -iTCP:<port> -sTCP:LISTEN` 输出可解析出进程名、PID、协议、地址
- 标题行和空行被忽略

- [ ] **Step 4: 运行测试并确认失败**

Run: `go test ./internal/app/dtool/controller -run TestToolPortProcess`

Expected: FAIL，提示查询/解析能力尚未实现

### Task 2: 实现查询与结束接口的最小后端代码

**Files:**
- Create: `internal/app/dtool/controller/tool_process.go`
- Modify: `internal/app/dtool/router.go`
- Test: `internal/app/dtool/controller/tool_process_test.go`

- [ ] **Step 1: 定义请求和返回结构**

实现：
- 查询请求：`port`
- 结束请求：`pid`
- 查询返回：`port` + `items`

- [ ] **Step 2: 实现端口参数校验**

要求：
- 接受字符串或数字输入后统一转成整数
- 严格限制在 `1-65535`

- [ ] **Step 3: 实现跨平台命令构造与输出解析**

实现：
- Windows 查询命令与 `tasklist` 名称补全
- Linux/macOS 查询命令
- 结果统一映射为结构化列表

- [ ] **Step 4: 实现结束进程能力**

实现：
- Windows 使用 `taskkill /PID <pid> /F`
- Linux/macOS 使用 `kill -9 <pid>`

- [ ] **Step 5: 注册路由**

新增：
- `POST /api/ToolPortProcessList`
- `POST /api/ToolPortProcessKill`

- [ ] **Step 6: 运行测试并确认通过**

Run: `go test ./internal/app/dtool/controller -run TestToolPortProcess`

Expected: PASS

## Chunk 2: 前端常用操作面板

### Task 3: 先写前端交互骨架，再接接口

**Files:**
- Create: `web/src/components/tools/CommonActions.vue`
- Modify: `web/src/components/Tools.vue`
- Create: `web/src/utils/base/tools.js`

- [ ] **Step 1: 新增常用操作 tab**

在 [Tools.vue](C:/work/frog/dev_tool_master/web/src/components/Tools.vue) 中接入 `CommonActions.vue`

- [ ] **Step 2: 搭建动作卡片面板**

要求：
- 至少包含一个“端口进程管理”卡片
- 预留后续继续扩展的布局结构

- [ ] **Step 3: 实现端口输入与查询状态**

要求：
- 输入框、查询按钮、加载态、空结果提示
- 非法端口在前端先拦截

- [ ] **Step 4: 实现查询结果列表与结束按钮**

要求：
- 每条结果展示 PID、进程名、协议、地址
- 每条结果单独触发结束流程

- [ ] **Step 5: 接入确认弹窗与结束接口**

要求：
- 用户确认后再结束进程
- 成功后自动刷新当前端口结果

## Chunk 3: 验证

### Task 4: 回归验证并记录剩余风险

**Files:**
- Modify: `README.md`（仅当需要补充小工具能力说明时）

- [ ] **Step 1: 运行后端测试**

Run: `go test ./internal/app/dtool/controller`

- [ ] **Step 2: 运行前端构建验证**

Run: `npm run build`
Workdir: `web`

- [ ] **Step 3: 手工验证核心流程**

验证：
- 小工具中新增 `常用操作` tab
- 查询端口后能看到结构化结果
- 二次确认后才会结束进程
- 结束成功后会刷新结果

- [ ] **Step 4: 记录未覆盖风险**

若当前机器无法同时验证 Windows、Linux、macOS 三个平台，需要在结果中明确说明剩余跨平台验证缺口
