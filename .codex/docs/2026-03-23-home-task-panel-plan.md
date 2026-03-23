# 首页任务清单下沉实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将任务清单直接展示在首页内容区最下方，并缩小侧边栏底部工具栏按钮。

**Architecture:** 保持任务清单数据和交互逻辑继续放在 `Home.vue`，避免侵入 `Dashboard.vue` 的首页命令逻辑。在 `Home.vue` 主内容区的 `router-view` 下方新增任务面板，同时移除旧弹窗入口；按钮缩小通过扩展共享组件 `GitActionButton.vue` 的紧凑尺寸实现。

**Tech Stack:** Vue 3、Element Plus、现有 `GitActionButton` 基础组件

---

### Task 1: 梳理首页任务清单渲染位置

**Files:**
- Modify: `web/src/components/Home.vue`

**Step 1: 确认首页主内容区结构**

查看 `router-view` 所在位置，确认可在其下方追加首页专属任务面板。

**Step 2: 确认旧入口和旧弹窗结构**

定位侧边栏底部“任务清单”按钮和 `el-dialog`，为迁移做准备。

### Task 2: 迁移任务清单到首页底部

**Files:**
- Modify: `web/src/components/Home.vue`

**Step 1: 先保留现有任务逻辑**

不改动 `loadHomeTaskList`、`saveHomeTask`、`quickUpdateHomeTaskStatus` 等方法行为，只调整渲染位置。

**Step 2: 在首页主内容区底部渲染任务面板**

仅在首页路由显示任务清单，保持活跃/归档标签、表单和快捷操作按钮。

**Step 3: 删除旧弹窗入口**

移除“任务清单”侧边栏按钮和对应 `el-dialog`，避免重复入口。

### Task 3: 缩小侧边栏底部按钮

**Files:**
- Modify: `web/src/components/base/GitActionButton.vue`
- Modify: `web/src/components/Home.vue`

**Step 1: 为共享按钮增加更小的紧凑样式**

在基础组件中新增可复用的更小尺寸 class，不在页面里重复定义按钮样式。

**Step 2: 仅应用到首页侧边栏底部工具栏**

将“新页卡”“小工具”“SSH”按钮切换到更小尺寸，避免影响全站其他按钮。

### Task 4: 验证

**Files:**
- Modify: `web/src/components/Home.vue`
- Modify: `web/src/components/base/GitActionButton.vue`

**Step 1: 运行前端构建验证**

Run: `npm run prod`

Expected: 构建成功，无语法错误。

**Step 2: 检查改动文件差异**

Run: `git diff -- web/src/components/Home.vue web/src/components/base/GitActionButton.vue`

Expected: 仅包含任务清单位置调整和按钮尺寸收敛改动。
