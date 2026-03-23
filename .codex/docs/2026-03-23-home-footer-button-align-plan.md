# 首页底部按钮对齐优化实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 让首页侧边栏底部“新页卡”“小工具”“SSH”三个按钮的文字与按钮盒模型视觉对齐。

**Architecture:** 仅修改共享按钮组件 `GitActionButton.vue` 的内部布局样式，统一 flex 对齐、行高和文字容器布局，不改首页页面结构。这样既能解决当前按钮不齐的问题，也能避免在单页组件里重复写覆盖样式。

**Tech Stack:** Vue 3、Element Plus、Scoped CSS

---

### Task 1: 调整共享按钮内部布局

**Files:**
- Modify: `web/src/components/base/GitActionButton.vue`

**Step 1: 统一按钮盒模型**

为 `.git-action-button` 增加 `display: inline-flex`、`align-items: center`、`justify-content: center` 和稳定的 `line-height`。

**Step 2: 收紧小尺寸按钮文本布局**

为 `.git-action-button--compact-small` 补充一致的文字高度与最小高度，避免不同文案造成的视觉中心偏移。

### Task 2: 验证

**Files:**
- Modify: `web/src/components/base/GitActionButton.vue`

**Step 1: 运行前端构建**

Run: `npm run prod`

Expected: 构建成功，无新增样式语法错误。

**Step 2: 检查差异**

Run: `git diff -- web/src/components/base/GitActionButton.vue`

Expected: 仅包含按钮内部布局和对齐相关样式改动。
