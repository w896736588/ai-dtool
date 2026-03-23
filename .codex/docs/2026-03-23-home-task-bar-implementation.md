# 首页任务清单按钮栏 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 在首页底部增加统一按钮栏，并接入基于后端表的任务清单弹窗与快捷状态切换能力。

**Architecture:** 沿用现有 `common + controller + router + web/src/utils/base + Home.vue` 分层。后端用 SQLite 新表持久化首页任务，前端直接在首页组件内承载弹窗与列表操作，避免引入额外全局状态。

**Tech Stack:** Go, Gin, SQLite, Vue 3 Options API, Element Plus

---

### Task 1: 新增后端任务常量与数据库访问测试

**Files:**
- Modify: `internal/app/dtool/define`
- Create: `internal/app/dtool/common/home_task_test.go`

**Step 1: Write the failing test**

- 为 `HomeTaskSave` 新增测试，断言新增后状态、开始时间、最后操作时间写入正确。
- 为 `HomeTaskStatusQuickUpdate` 新增测试，断言切到 `进行中` 时会补齐开始时间并刷新最后操作时间。

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/dtool/common -run HomeTask -count=1`

Expected: FAIL，提示缺少首页任务相关实现。

**Step 3: Write minimal implementation**

- 增加首页任务状态常量。
- 增加数据库访问方法签名与最小实现。

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/dtool/common -run HomeTask -count=1`

Expected: PASS

### Task 2: 新增首页任务迁移与后端实现

**Files:**
- Create: `internal/app/dtool/database/2026/03/202603231030_home_task.sql`
- Create: `internal/app/dtool/common/home_task.go`
- Modify: `internal/app/dtool/controller`
- Modify: `internal/app/dtool/router.go`
- Modify: `internal/app/dtool/define`

**Step 1: Write the failing test**

- 保持 Task 1 的失败测试作为约束，不新增重复测试。

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/dtool/common -run HomeTask -count=1`

Expected: FAIL

**Step 3: Write minimal implementation**

- 创建表 `tbl_home_task`
- 封装 `HomeTaskList`、`HomeTaskSave`、`HomeTaskArchiveToggle`、`HomeTaskStatusQuickUpdate`
- 增加对应 controller 和 router 注册
- 所有硬编码状态提取为常量

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/dtool/common -run HomeTask -count=1`

Expected: PASS

### Task 3: 新增前端接口封装

**Files:**
- Create: `web/src/utils/base/home_task.js`

**Step 1: Write the failing test**

- 当前仓库缺少前端单测基建，本任务以接口文件静态实现为主，不单独写单测。

**Step 2: Write minimal implementation**

- 封装列表、保存、归档切换、快捷状态切换接口。

### Task 4: 首页底部按钮栏与任务弹窗实现

**Files:**
- Modify: `web/src/components/Home.vue`

**Step 1: Write the failing test**

- 当前仓库缺少首页前端单测基建，本任务改为在实现后执行构建验证。

**Step 2: Write minimal implementation**

- 增加统一按钮栏
- 增加任务清单按钮
- 增加居中弹窗、Tab、任务新增表单、任务列表、快捷状态按钮、归档按钮
- 显示开始时间与最后操作时间

**Step 3: Run verification**

Run: `npm run build`

Workdir: `web`

Expected: PASS

### Task 5: 整体验证

**Files:**
- Modify: `internal/app/dtool/common/home_task_test.go`
- Modify: `web/src/components/Home.vue`

**Step 1: Run backend tests**

Run: `go test ./internal/app/dtool/common -run HomeTask -count=1`

Expected: PASS

**Step 2: Run frontend build**

Run: `npm run build`

Workdir: `web`

Expected: PASS

**Step 3: Manual check**

- 打开首页
- 点击底部 `任务清单`
- 新增一条任务
- 快捷切换到 `进行中`
- 归档该任务并切换到 `归档` Tab 检查显示
