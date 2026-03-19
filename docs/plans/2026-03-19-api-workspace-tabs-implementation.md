# API Workspace Tabs Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将接口开发页右侧从单内容区改为可同时保留多个集合、文件夹、接口内容的 tab 工作区，并支持重复点击切换到已有 tab 后重新加载数据。

**Architecture:** 在 `Api.vue` 中引入工作区 tab 状态层，由 `openTabs + activeTabKey` 驱动右侧内容渲染；保留 `selectedItem` 作为当前激活 tab 的兼容镜像，渐进替换现有单实例逻辑。集合、文件夹、接口详情组件继续复用，只改父组件状态编排和事件联动。

**Tech Stack:** Vue 3 Options API, Element Plus `el-tabs`, 现有 Api 请求封装

---

### Task 1: 为工作区 tab 状态写失败测试说明并完成状态骨架

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/components/Api.vue`

**Step 1: 写失败测试说明**

由于仓库当前未配置前端单元测试框架，本任务改为先写最小可验证状态骨架，并在实现阶段通过单文件 ESLint + 手工回归验证覆盖。不要引入新的测试基础设施。

**Step 2: 添加工作区状态字段**

在 `data()` 中新增：

- `openTabs: []`
- `activeTabKey: ''`
- `activeCollectionInnerTabMap: {}`
- `activeFolderInnerTabMap: {}`

保留 `selectedItem`，但注释说明其改为当前激活 tab 的镜像。

**Step 3: 添加 tab 标识与查找方法**

实现最小方法：

- `buildWorkspaceTabKey(node)`
- `getWorkspaceTabByKey(tabKey)`
- `getActiveWorkspaceTab()`

**Step 4: 运行静态校验**

Run: `npx eslint src/components/Api.vue`

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Api.vue
git commit -m "refactor: add api workspace tab state"
```

### Task 2: 将右侧单内容区改为 tab 工作区容器

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/components/Api.vue`

**Step 1: 写失败验证目标**

定义失败场景：

- 当前右侧只能显示一个内容。
- 打开第二个节点时，第一个内容被覆盖。

**Step 2: 添加工作区 tab 条**

在右侧面板中新增 `el-tabs` 作为工作区容器，使用：

- `v-model="activeTabKey"`
- `closable`
- `@tab-remove="closeWorkspaceTab"`
- `@tab-change="handleWorkspaceTabChange"`

tab label 显示节点名称，必要时追加类型标识。

**Step 3: 将内容渲染改为按 active tab 输出**

保留现有集合、文件夹、接口的三个内容分支，但它们的输入改为“当前激活 tab 的数据”，而不是直接取 `selectedItem` 的历史写法。

**Step 4: 空状态改为无 tab 时展示**

当 `openTabs.length === 0` 时，显示原有空状态。

**Step 5: 运行静态校验**

Run: `npx eslint src/components/Api.vue`

Expected: PASS

**Step 6: Commit**

```bash
git add web/src/components/Api.vue
git commit -m "feat: render api workspace tabs"
```

### Task 3: 改造节点点击逻辑为“打开或激活 tab”

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/components/Api.vue`

**Step 1: 写失败验证目标**

定义失败场景：

- 点击树节点时没有创建 tab。
- 重复点击同一节点时会重复创建 tab。

**Step 2: 实现 tab 打开与激活方法**

实现：

- `createWorkspaceTab(node)`
- `openWorkspaceTab(node, { reload })`
- `activateWorkspaceTab(tabKey, { reload })`
- `syncSelectedItemFromActiveTab()`

规则：

- 首次点击新建 tab。
- 已有 tab 切换过去，不重复创建。
- 切换后同步左侧树高亮和 `selectedItem`。

**Step 3: 将 `handleNodeClick` 接入工作区**

把当前直接改 `selectedItem` 的逻辑替换为：

- 集合：打开或激活集合 tab，并刷新集合数据。
- 文件夹：打开或激活文件夹 tab，并刷新文件夹数据。
- 接口：打开或激活接口 tab，并刷新接口详情。

**Step 4: 运行静态校验**

Run: `npx eslint src/components/Api.vue`

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Api.vue
git commit -m "feat: open api nodes in workspace tabs"
```

### Task 4: 实现重复点击已有 tab 时重新加载数据

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/components/Api.vue`

**Step 1: 写失败验证目标**

定义失败场景：

- 重复点击已有集合/文件夹/接口 tab 时只切换，不刷新数据。

**Step 2: 实现统一刷新入口**

新增：

- `reloadWorkspaceTab(tabKey)`
- `reloadCollectionTab(tab)`
- `reloadFolderTab(tab)`
- `reloadApiTab(tab)`

要求：

- 集合：刷新集合基础信息。
- 文件夹：刷新文件夹信息和接口摘要。
- 接口：刷新接口详情。

**Step 3: 接入 `handleNodeClick` 与 `handleWorkspaceTabChange`**

确保重复点击树节点时执行 reload；仅切换右侧 tab 时不额外重复请求。

**Step 4: 运行静态校验**

Run: `npx eslint src/components/Api.vue`

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Api.vue
git commit -m "feat: reload existing api workspace tabs on reopen"
```

### Task 5: 将创建、更新、删除动作接入工作区 tab

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/components/Api.vue`

**Step 1: 写失败验证目标**

定义失败场景：

- 创建接口后未自动切到新接口 tab。
- 删除文件夹后右侧保留已失效的文件夹或接口 tab。
- 更新名称后 tab 标题不变。

**Step 2: 实现 tab 同步辅助方法**

新增：

- `upsertWorkspaceTabData(nodeLike)`
- `closeWorkspaceTab(tabKey)`
- `closeWorkspaceTabsByFolder(folderId)`
- `closeWorkspaceTabsByCollection(collectionId)`

**Step 3: 接入现有 CRUD 逻辑**

在这些流程里更新 tab 状态：

- `createNewCollection`
- `createNewDir`
- `handleFolderCreateApi`
- `copyApi`
- `handleCollectionUpdate/Delete`
- `handleFolderUpdate/Delete`
- `handleApiUpdate`
- `handleApiAction('delete_api')`

**Step 4: 运行静态校验**

Run: `npx eslint src/components/Api.vue`

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Api.vue
git commit -m "refactor: sync api workspace tabs with crud actions"
```

### Task 6: 保证树高亮、懒加载恢复与 tab 切换一致

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/components/Api.vue`

**Step 1: 写失败验证目标**

定义失败场景：

- 切换到接口 tab 时左侧树未高亮对应节点。
- 所属集合或文件夹未加载时无法正确定位当前接口。

**Step 2: 实现树定位方法**

新增：

- `ensureNodeVisibleInTree(tab)`
- `highlightWorkspaceTreeNode(tab)`

要求：

- 集合 tab：直接高亮集合节点。
- 文件夹 tab：先确保集合文件夹已加载，再高亮文件夹。
- 接口 tab：先确保集合和文件夹都已加载，再高亮接口。

**Step 3: 在 tab 激活时接入树同步**

在 `activateWorkspaceTab` 或 `handleWorkspaceTabChange` 中调用上述方法。

**Step 4: 运行静态校验**

Run: `npx eslint src/components/Api.vue`

Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Api.vue
git commit -m "feat: sync api workspace tabs with tree selection"
```

### Task 7: 手工回归验证并补充文档说明

**Files:**
- Modify: `D:/go/cache_manager_api/docs/plans/2026-03-19-api-workspace-tabs-design.md`
- Modify: `D:/go/cache_manager_api/docs/plans/2026-03-19-api-workspace-tabs-implementation.md`

**Step 1: 手工验证**

验证以下路径：

- 打开多个集合、文件夹、接口，右侧可保留多个 tab。
- 重复点击同一节点时切换到已有 tab，并刷新数据。
- 新建接口后自动打开新接口 tab。
- 删除文件夹后，相关文件夹和接口 tab 被关闭。
- 切换 tab 时左侧树高亮同步。

**Step 2: 记录实际验证结果**

在设计文档或实现计划中补充“实际验证结果 / 未覆盖项”。

**Step 3: 运行静态校验**

Run: `npx eslint src/components/Api.vue`

Expected: PASS

**Step 4: Commit**

```bash
git add docs/plans/2026-03-19-api-workspace-tabs-design.md docs/plans/2026-03-19-api-workspace-tabs-implementation.md web/src/components/Api.vue
git commit -m "docs: finalize api workspace tabs rollout notes"
```
