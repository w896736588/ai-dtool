# 知识片段 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 新增“知识片段”主菜单与独立知识片段页面，支持 Markdown、自由标签、历史记录、假删除、右侧多文档 Tab，以及 `FTS5 + sqlite-vec` 驱动的关键词检索、自然语言检索和混合检索。
**Architecture:** 采用独立业务模块方案。后端新增 memory 相关表、控制器、检索索引和异步任务处理；前端新增独立页面与子组件，通过左侧片段列表 + 右侧搜索区与 Tab 工作区完成交互。保存时同步写主数据与历史，异步更新全文索引和向量索引。
**Tech Stack:** Go、Gin、SQLite、FTS5、sqlite-vec、goqite、Vue 3、Element Plus、md-editor-v3、PowerShell

---

### Task 1: 设计数据库结构与迁移脚本

**Files:**
- Create: `internal/app/dtool/database/2026/03/20260306.知识片段.sql`
- Modify: `internal/app/dtool/common/db.go`

**Step 1: 写出失败前的校验目标**

人工校验目标：
- 存在主表、标签表、历史表；
- 存在 FTS5 虚表；
- 存在向量索引表；
- 支持片段假删除与索引状态字段；
- 不影响现有 `tbl_markdown` 相关逻辑。

**Step 2: 先编写迁移后的数据访问测试骨架**

```go
func TestMemoryFragmentSaveAndList(t *testing.T) {
    t.Fatal("not implemented")
}
```

**Step 3: 运行测试确认失败**

Run: `go test ./internal/app/dtool/...`
Expected: 因测试占位或新方法未实现而失败，确认测试入口有效。

**Step 4: 编写最小迁移与数据访问方法**

实现内容：
- 新增主表、标签表、历史表
- 新增 FTS5 虚表
- 新增向量索引表
- 在 `db.go` 中补充基础方法签名

**Step 5: 再次运行测试**

Run: `go test ./internal/app/dtool/...`
Expected: 结构性错误减少，至少可以进入下一步后端逻辑实现。

**Step 6: Commit**

```bash
git add internal/app/dtool/database/2026/03/20260306.知识片段.sql internal/app/dtool/common/db.go
git commit -m "feat: add memory fragment schema and search tables"
```

### Task 2: 为主数据保存与历史记录写失败用例

**Files:**
- Create: `internal/app/dtool/controller/memory_fragment_test.go`
- Modify: `internal/app/dtool/common/db.go`

**Step 1: 写失败测试**

```go
func TestMemoryFragmentSaveCreatesHistoryWhenChanged(t *testing.T) {}
func TestMemoryFragmentSoftDeleteHidesFromDefaultList(t *testing.T) {}
func TestMemoryFragmentSaveMarksIndexPending(t *testing.T) {}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/app/dtool/controller/...`
Expected: 因缺少 memory 相关方法或行为不符合预期而失败。

**Step 3: 实现最小后端逻辑**

实现内容：
- 保存片段
- 重建标签
- 内容、标题或标签变化时写历史
- 默认列表过滤 `is_deleted = 0`
- 保存后将 `index_status` 置为 `pending`

**Step 4: 运行测试确认通过**

Run: `go test ./internal/app/dtool/controller/...`
Expected: 新增测试通过。

**Step 5: Commit**

```bash
git add internal/app/dtool/controller/memory_fragment_test.go internal/app/dtool/common/db.go
git commit -m "feat: implement memory fragment save and history"
```

### Task 3: 为 FTS5 检索写失败用例并实现

**Files:**
- Modify: `internal/app/dtool/controller/memory_fragment_test.go`
- Modify: `internal/app/dtool/common/db.go`

**Step 1: 写失败测试**

```go
func TestMemoryFragmentFTSSearchAllTermsMustMatch(t *testing.T) {}
func TestMemoryFragmentFTSSearchSupportsTagAndContent(t *testing.T) {}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/app/dtool/controller/...`
Expected: 因 FTS 查询方法未实现或结果不符合预期而失败。

**Step 3: 实现最小检索逻辑**

实现内容：
- 维护 FTS5 索引更新方法
- 支持标题、正文纯文本、标签文本的全文检索
- 通过 `AND` 构造实现空格分词“全部命中”

**Step 4: 运行测试确认通过**

Run: `go test ./internal/app/dtool/controller/...`
Expected: FTS5 相关测试通过。

**Step 5: Commit**

```bash
git add internal/app/dtool/controller/memory_fragment_test.go internal/app/dtool/common/db.go
git commit -m "feat: add memory fragment fts search"
```

### Task 4: 为向量检索写失败用例并实现基础存取

**Files:**
- Modify: `internal/app/dtool/controller/memory_fragment_test.go`
- Modify: `internal/app/dtool/common/db.go`

**Step 1: 写失败测试**

```go
func TestMemoryFragmentVectorSearchReturnsSimilarFragments(t *testing.T) {}
func TestMemoryFragmentVectorSearchIgnoresDeletedFragments(t *testing.T) {}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/app/dtool/controller/...`
Expected: 因向量索引读写未实现而失败。

**Step 3: 实现最小逻辑**

实现内容：
- 写入向量索引
- 根据查询向量取 TopK 结果
- 默认排除已删除片段

**Step 4: 运行测试确认通过**

Run: `go test ./internal/app/dtool/controller/...`
Expected: 向量检索相关测试通过。

**Step 5: Commit**

```bash
git add internal/app/dtool/controller/memory_fragment_test.go internal/app/dtool/common/db.go
git commit -m "feat: add memory fragment vector search"
```

### Task 5: 实现混合检索与排序

**Files:**
- Modify: `internal/app/dtool/controller/memory_fragment_test.go`
- Modify: `internal/app/dtool/common/db.go`

**Step 1: 写失败测试**

```go
func TestMemoryFragmentHybridSearchMergesAndDeduplicates(t *testing.T) {}
func TestMemoryFragmentHybridSearchSupportsTagFilter(t *testing.T) {}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/app/dtool/controller/...`
Expected: 因混合检索未实现而失败。

**Step 3: 实现最小逻辑**

实现内容：
- 同时执行 FTS 与向量检索
- 合并并去重结果
- 增加简单排序规则
- 支持标签过滤叠加

**Step 4: 运行测试确认通过**

Run: `go test ./internal/app/dtool/controller/...`
Expected: 混合检索相关测试通过。

**Step 5: Commit**

```bash
git add internal/app/dtool/controller/memory_fragment_test.go internal/app/dtool/common/db.go
git commit -m "feat: add memory fragment hybrid search"
```

### Task 6: 增加异步索引任务

**Files:**
- Modify: `internal/taskmanager/*`
- Modify: `internal/app/dtool/common/db.go`
- Modify: `internal/app/dtool/controller/memory_fragment_test.go`

**Step 1: 写失败测试或校验目标**

人工校验目标：
- 保存片段后任务被入队
- 任务执行后可更新 FTS 索引
- 任务执行后可更新向量索引
- 任务失败时 `index_status` 为 `failed`

**Step 2: 运行测试确认失败**

Run: `go test ./internal/...`
Expected: 因任务处理器未实现而失败。

**Step 3: 实现最小逻辑**

实现内容：
- 增加知识片段索引任务入队方法
- 增加 Worker 处理逻辑
- 增加失败重试或失败标记

**Step 4: 运行测试确认通过**

Run: `go test ./internal/...`
Expected: 索引任务相关测试通过。

**Step 5: Commit**

```bash
git add internal/taskmanager internal/app/dtool/common/db.go internal/app/dtool/controller/memory_fragment_test.go
git commit -m "feat: add memory fragment indexing tasks"
```

### Task 7: 增加知识片段控制器与路由

**Files:**
- Create: `internal/app/dtool/controller/memory_fragment.go`
- Modify: `internal/app/dtool/router.go`
- Modify: `internal/app/dtool/controller/base.go`

**Step 1: 先写接口层失败测试**

```go
func TestMemoryFragmentListAPI(t *testing.T) {}
func TestMemoryFragmentSaveAPI(t *testing.T) {}
func TestMemoryFragmentDeleteAPI(t *testing.T) {}
func TestMemoryFragmentSearchAPI(t *testing.T) {}
```

**Step 2: 运行测试确认失败**

Run: `go test ./internal/app/dtool/controller/...`
Expected: 路由或控制器未实现导致失败。

**Step 3: 编写最小实现**

接口建议：
- `/api/MemoryFragmentList`
- `/api/MemoryFragmentInfo`
- `/api/MemoryFragmentSave`
- `/api/MemoryFragmentDelete`
- `/api/MemoryFragmentHistoryList`
- `/api/MemoryFragmentTagList`
- `/api/MemoryFragmentSearch`

要求：
- 每个后端方法增加中文注释
- 参数校验与错误返回风格保持现有项目一致

**Step 4: 运行测试**

Run: `go test ./internal/app/dtool/controller/...`
Expected: 接口测试通过。

**Step 5: Commit**

```bash
git add internal/app/dtool/controller/memory_fragment.go internal/app/dtool/router.go internal/app/dtool/controller/base.go
git commit -m "feat: add memory fragment api endpoints"
```

### Task 8: 新增前端 API 封装与路由入口

**Files:**
- Create: `web/src/utils/base/memory_fragment.js`
- Modify: `web/src/router/index.js`
- Modify: `web/src/components/Home.vue`
- Modify: `web/src/utils/module.js`

**Step 1: 先写人工校验点**

人工校验目标：
- 左侧菜单出现“知识片段”
- 点击后进入新页面
- 仅增加与该模块相关的菜单和路由，不改无关菜单行为

**Step 2: 运行前端基础校验命令**

Run: `npm run lint`
Expected: 若存在历史问题，至少确认新文件无语法错误。

**Step 3: 编写最小实现**

实现内容：
- 新增 memory API 方法封装
- 新增 `/MemoryFragment` 路由
- 在 `Home.vue` 中增加菜单项
- 在模块开关列表中开放 `memory_fragment`

要求：
- 每个前端方法增加中文注释
- 关键状态流转增加中文注释

**Step 4: 再次运行校验**

Run: `npm run lint`
Expected: 新增改动不引入新的语法或 lint 错误。

**Step 5: Commit**

```bash
git add web/src/utils/base/memory_fragment.js web/src/router/index.js web/src/components/Home.vue web/src/utils/module.js
git commit -m "feat: add memory fragment menu and route"
```

### Task 9: 搭建知识片段主页面骨架

**Files:**
- Create: `web/src/components/MemoryFragment.vue`
- Create: `web/src/components/memory/MemoryWelcome.vue`
- Create: `web/src/components/memory/MemoryEditor.vue`
- Create: `web/src/components/memory/MemoryHistoryDialog.vue`

**Step 1: 先写失败前的页面校验目标**

人工校验目标：
- 左侧是片段列表
- 右侧顶部是搜索框、检索模式切换与标签区
- 默认显示“首页”Tab
- 可打开多个片段 Tab

**Step 2: 运行前端构建校验**

Run: `npm run lint`
Expected: 当前分支可执行基础校验。

**Step 3: 编写最小页面骨架**

实现内容：
- 主容器左右布局
- 首页 Tab 固定存在
- 维护打开 Tab 数组与当前激活 Tab
- 左侧点击片段时打开或切换 Tab
- 右侧欢迎页和编辑页按当前 Tab 切换展示

**Step 4: 运行校验**

Run: `npm run lint`
Expected: 页面组件可正常编译。

**Step 5: Commit**

```bash
git add web/src/components/MemoryFragment.vue web/src/components/memory/MemoryWelcome.vue web/src/components/memory/MemoryEditor.vue web/src/components/memory/MemoryHistoryDialog.vue
git commit -m "feat: scaffold memory fragment page"
```

### Task 10: 实现片段列表、详情与保存

**Files:**
- Modify: `web/src/components/MemoryFragment.vue`
- Modify: `web/src/components/memory/MemoryEditor.vue`
- Modify: `web/src/utils/base/memory_fragment.js`

**Step 1: 先写页面行为失败用例或人工校验点**

人工校验目标：
- 可以新建片段
- 可以打开片段详情
- 编辑标题、正文、标签后可保存
- 保存后左侧更新时间同步变化
- 保存后索引状态进入 `待索引`

**Step 2: 运行校验命令**

Run: `npm run lint`
Expected: 当前页面基础通过。

**Step 3: 实现最小保存链路**

实现内容：
- 左侧列表加载
- 片段详情加载
- 新建片段
- 编辑器保存
- Tab 未保存状态 `*`
- 保存后刷新列表与当前详情
- 显示索引状态

**Step 4: 运行校验**

Run: `npm run lint`
Expected: 语法与模板通过。

**Step 5: Commit**

```bash
git add web/src/components/MemoryFragment.vue web/src/components/memory/MemoryEditor.vue web/src/utils/base/memory_fragment.js
git commit -m "feat: add memory fragment editor workflow"
```

### Task 11: 实现自由标签与检索模式切换

**Files:**
- Modify: `web/src/components/MemoryFragment.vue`
- Modify: `web/src/components/memory/MemoryWelcome.vue`
- Modify: `web/src/components/memory/MemoryEditor.vue`
- Modify: `web/src/utils/base/memory_fragment.js`

**Step 1: 写失败前校验目标**

人工校验目标：
- 标签可新增、删除、去重
- 点击标签可筛选
- 可切换 `混合 / 关键词 / 自然语言` 检索模式
- 标签筛选与检索条件叠加生效

**Step 2: 运行校验命令**

Run: `npm run lint`
Expected: 当前状态可继续开发。

**Step 3: 实现最小交互**

实现内容：
- 标签输入组件
- 标签区筛选状态
- 检索模式切换
- 列表与欢迎页联动刷新

**Step 4: 运行校验**

Run: `npm run lint`
Expected: 前端校验通过。

**Step 5: Commit**

```bash
git add web/src/components/MemoryFragment.vue web/src/components/memory/MemoryWelcome.vue web/src/components/memory/MemoryEditor.vue web/src/utils/base/memory_fragment.js
git commit -m "feat: add memory fragment tags and search modes"
```

### Task 12: 实现历史记录与假删除联动

**Files:**
- Modify: `web/src/components/memory/MemoryHistoryDialog.vue`
- Modify: `web/src/components/memory/MemoryEditor.vue`
- Modify: `web/src/components/MemoryFragment.vue`
- Modify: `web/src/utils/base/memory_fragment.js`

**Step 1: 先写失败前校验目标**

人工校验目标：
- 历史记录列表可查看
- 正文 diff 可展示
- 删除后片段从左侧消失
- 已打开 Tab 删除后自动关闭并回退到首页

**Step 2: 运行校验命令**

Run: `npm run lint`
Expected: 当前代码可继续迭代。

**Step 3: 实现最小联动**

实现内容：
- 历史弹窗
- 历史对比
- 假删除按钮
- 删除后的 Tab 清理逻辑

**Step 4: 运行校验**

Run: `npm run lint`
Expected: 页面仍可编译通过。

**Step 5: Commit**

```bash
git add web/src/components/memory/MemoryHistoryDialog.vue web/src/components/memory/MemoryEditor.vue web/src/components/MemoryFragment.vue web/src/utils/base/memory_fragment.js
git commit -m "feat: add memory fragment history and soft delete"
```

### Task 13: 完整验证

**Files:**
- Modify: `internal/app/dtool/controller/memory_fragment_test.go`
- Modify: `web/src/components/MemoryFragment.vue`
- Modify: `web/src/components/memory/MemoryEditor.vue`

**Step 1: 补齐遗漏测试或校验清单**

后端测试补充：
- FTS 与向量检索都排除已删除数据
- 混合检索去重正确
- 索引任务失败与重试状态正确

前端人工校验：
- 首页 Tab 不可关闭
- 同一片段不会重复打开 Tab
- 搜索清空后列表恢复
- 标签取消筛选后结果恢复
- 检索模式切换后结果发生对应变化

**Step 2: 运行后端测试**

Run: `go test ./internal/app/dtool/...`
Expected: memory 相关后端测试通过；若存在历史失败项，需要明确记录与本功能无关。

**Step 3: 运行前端校验**

Run: `npm run lint`
Expected: 前端校验通过；若存在历史问题，需要明确记录与本功能无关。

**Step 4: 运行前端构建**

Run: `npm run prod`
Expected: 构建成功，说明新增页面可正常打包。

**Step 5: Commit**

```bash
git add .
git commit -m "feat: finish memory fragment retrieval module"
```
