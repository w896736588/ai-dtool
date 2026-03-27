# 知识片段问题修复与回收站 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 修复知识片段编辑区滚动与保存交互问题，并新增回收站及已有标签选择能力。

**Architecture:** 基于现有知识片段模块做增量修改。后端复用当前软删除模型，补充回收站列表、恢复和彻底删除接口；前端保持现有 Tab 工作区结构，在编辑器内部修正弹性布局与编辑状态同步，并扩展标签输入区为“自由输入 + 已有标签候选”。

**Tech Stack:** Go, Gin, SQLite, Vue 3, Element Plus, md-editor-v3

---

### Task 1: 回收站后端能力

**Files:**
- Modify: `internal/app/dtool/common/db.go`
- Modify: `internal/app/dtool/controller/memory_fragment.go`
- Modify: `web/src/utils/base/memory_fragment.js`

**Step 1: 写数据库层失败用例**

- 覆盖软删除后可出现在回收站列表。
- 覆盖恢复后重新出现在正常列表且不在回收站。
- 覆盖彻底删除后详情与回收站都不可见。

**Step 2: 运行目标测试确认失败**

Run: `go test ./internal/app/dtool/common -run MemoryFragment -count=1`

**Step 3: 实现最小后端能力**

- 增加回收站列表查询方法。
- 增加恢复片段方法。
- 增加彻底删除片段及关联标签、历史、索引清理方法。
- 增加 Gin 控制器与前端 API 封装。

**Step 4: 重跑目标测试确认通过**

Run: `go test ./internal/app/dtool/common -run MemoryFragment -count=1`

### Task 2: 编辑器滚动与保存态修复

**Files:**
- Modify: `web/src/components/memory/MemoryEditor.vue`

**Step 1: 先补行为用例或最小验证点**

- 明确 `savedFragment` 更新后不应强制退出编辑模式。
- 明确编辑器容器使用 flex + `min-height: 0`，由内容区独立滚动。

**Step 2: 实现最小前端改动**

- 调整编辑器区域容器结构和样式，消除固定 `100vh` 高度造成的滚动截断。
- 调整草稿重置逻辑，区分“切换片段”和“保存当前片段”，保存后保留编辑态。

**Step 3: 本地构建验证**

Run: `npm --prefix web run build`

### Task 3: 回收站页面与标签候选交互

**Files:**
- Modify: `web/src/components/MemoryFragment.vue`
- Modify: `web/src/components/memory/MemoryEditor.vue`

**Step 1: 接入回收站 Tab**

- 增加左侧入口。
- 增加已删除片段列表展示。
- 支持恢复与彻底删除，并刷新主列表/标签/搜索结果。

**Step 2: 接入已有标签候选区**

- 在编辑器标签区域展示可选已有标签。
- 过滤当前已选标签。
- 点击候选标签立即加入并同步脏状态。

**Step 3: 重跑前端验证**

Run: `npm --prefix web run build`

### Task 4: 整体验证

**Files:**
- Verify only

**Step 1: 运行后端测试**

Run: `go test ./internal/app/dtool/common -run MemoryFragment -count=1`

**Step 2: 运行前端构建**

Run: `npm --prefix web run build`

**Step 3: 手工核对需求**

- 编辑区底部内容可滚动可见。
- `Ctrl+S` 保存后保持编辑模式。
- 删除后可在回收站看到。
- 回收站支持恢复和彻底删除。
- 标签区既可输入新标签，也可点选已有标签。
