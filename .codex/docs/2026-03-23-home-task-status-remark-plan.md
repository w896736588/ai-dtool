# 首页任务清单状态与备注调整 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 调整首页任务清单状态流转，新增备注字段，并把列表操作收敛为单个状态变更按钮。

**Architecture:** 保持现有首页任务的 `define + struct + controller + common + web/src/utils/base + Home.vue` 分层不变。后端通过新增备注字段和状态常量扩展原有接口，前端继续在 `Home.vue` 内完成表单、列表和状态变更菜单渲染，避免引入新的全局状态或额外组件。

**Tech Stack:** Go、SQLite、Vue 3、Element Plus

---

### Task 1: 后端任务模型扩展

**Files:**
- Modify: `internal/app/dtool/define/home_task.go`
- Modify: `internal/app/dtool/struct/home_task.go`
- Modify: `internal/app/dtool/controller/home_task.go`
- Modify: `internal/app/dtool/common/home_task.go`
- Test: `internal/app/dtool/common/home_task_test.go`

**Step 1: Write the failing test**

补充首页任务保存与快捷切换测试，断言：
- 保存时会写入 `remark`
- 新状态集合合法
- 切换到 `开发中` 且没有开始时间时自动补齐开始时间

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/dtool/common -run HomeTask -count=1`
Expected: FAIL，提示 `remark` 字段缺失、状态不合法或旧状态常量不匹配。

**Step 3: Write minimal implementation**

更新首页任务常量、请求结构、控制器透传和数据库保存逻辑，让 `remark` 与新状态集生效。

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/dtool/common -run HomeTask -count=1`
Expected: PASS

### Task 2: 数据库迁移补齐

**Files:**
- Modify: `internal/app/dtool/database/2026/03/202603231030_home_task.sql`
- Create: `internal/app/dtool/database/2026/03/202603231130_home_task_remark_status.sql`

**Step 1: Write the migration**

为已有库和新库统一增加增量迁移，补 `remark` 字段并迁移旧状态文案，避免新库重复执行 `ADD COLUMN` 失败。

**Step 2: Verify migration content**

Run: `git diff -- internal/app/dtool/database/2026/03/202603231030_home_task.sql internal/app/dtool/database/2026/03/202603231130_home_task_remark_status.sql`
Expected: 仅包含首页任务备注字段与状态映射相关 SQL。

### Task 3: 首页任务表单与列表交互调整

**Files:**
- Modify: `web/src/components/Home.vue`
- Modify: `web/src/utils/base/home_task.js`

**Step 1: Update form and state constants**

新增备注输入框、替换状态选项、同步保存请求字段。

**Step 2: Replace quick buttons with one status action**

将任务列表上多个状态按钮替换为右侧单个 `状态变更` 下拉按钮，并保留 `编辑`、`归档/取消归档`。

**Step 3: Show remark directly**

在任务卡片正文直接展示备注，空备注时不渲染备注区。

**Step 4: Run frontend build verification**

Run: `npm run prod`
Workdir: `web`
Expected: PASS

### Task 4: 最终回归验证

**Files:**
- Modify: `web/src/components/Home.vue`
- Modify: `internal/app/dtool/common/home_task.go`

**Step 1: Run focused backend tests**

Run: `go test ./internal/app/dtool/common -run HomeTask -count=1`
Expected: PASS

**Step 2: Run frontend build**

Run: `npm run prod`
Workdir: `web`
Expected: PASS
