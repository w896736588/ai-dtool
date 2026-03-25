# Tools Common Actions Right Tabs Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将小工具“常用操作”区域从块状布局改为右侧竖向 Tab 菜单，左侧显示对应操作面板。

**Architecture:** 保持 `Tools.vue` 外层导航不动，只在 `CommonActions.vue` 内部引入右侧 `el-tabs` 容器，把现有“命令托管”和“端口进程管理”分别收敛为两个 Tab 面板。原有接口和状态逻辑继续复用，只调整模板结构和样式。

**Tech Stack:** Vue 3, Element Plus

---

## Chunk 1: 常用操作布局改造

### Task 1: 改造常用操作模板结构

**Files:**
- Modify: `web/src/components/tools/CommonActions.vue`

- [ ] **Step 1: 写出右侧 Tab 需要的最小状态**

添加当前激活页签状态，默认选中 `命令托管`。

- [ ] **Step 2: 将块状布局改为 `el-tabs`**

保留原有两个操作面板内容，分别放入两个 `el-tab-pane`。

- [ ] **Step 3: 删除右侧占位卡**

后续扩展改为新增 Tab，不再保留独立占位区。

### Task 2: 调整样式与响应式行为

**Files:**
- Modify: `web/src/components/tools/CommonActions.vue`

- [ ] **Step 1: 为右侧 Tab 容器补充布局样式**

确保桌面端为“左内容、右菜单”。

- [ ] **Step 2: 为移动端补充降级样式**

窄屏时允许内容与菜单纵向排布，但不改变 Tab 交互。

## Chunk 2: 验证

### Task 3: 运行前端校验

**Files:**
- Modify: `docs/superpowers/specs/2026-03-22-tools-common-actions-right-tabs-design.md`

- [ ] **Step 1: 运行 lint**

Run: `npm run lint`
Workdir: `web`

- [ ] **Step 2: 手工验证关键路径**

验证：
- “常用操作”显示为右侧竖向 Tab
- `命令托管` 和 `端口进程管理` 可切换
- 原有按钮与结果区域仍正常显示
