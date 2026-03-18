# API Lazy Loading Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将接口开发页从整树全量加载改为集合、文件夹、接口摘要、接口详情四级按需请求。

**Architecture:** 在 [Api.vue](/D:/go/cache_manager_api/web/src/components/Api.vue) 中引入节点标准化与懒加载入口，首屏仅拉集合基础信息；集合和文件夹在展开时加载其子节点；接口详情在点击节点时单独拉取并交给 [ApiDetail.vue](/D:/go/cache_manager_api/web/src/components/api/ApiDetail.vue)。保留本地展开状态缓存，但恢复逻辑改为按层串行恢复。

**Tech Stack:** Vue 3 Options API, Element Plus, existing `@/utils/base/api` request helpers, ESLint via Vue CLI

---

### Task 1: 梳理并固定树节点数据结构

**Files:**
- Modify: `web/src/components/Api.vue`
- Modify: `web/src/utils/base/api.js`

**Step 1: Write the failing test**

由于当前仓库没有现成的组件测试基建，先写一个最小的设计约束清单到实现注释中，明确节点结构必须统一包含 `type/children/loaded/loading`，并在实现时以此为验收条件。

**Step 2: Run test to verify it fails**

Run: `npx eslint src/components/Api.vue`
Expected: PASS today, but there is no behavior coverage for lazy-loading node state. This confirms the gap before implementation.

**Step 3: Write minimal implementation**

在 [Api.vue](/D:/go/cache_manager_api/web/src/components/Api.vue) 新增标准化方法：

- `normalizeCollectionNode(collection)`
- `normalizeFolderNode(folder, collectionId)`
- `normalizeApiNode(api, folderId, collectionId)`

要求每个方法显式返回：

```js
{
  ...raw,
  type: 'collection' | 'folder' | 'api',
  children: [],
  loaded: false,
  loading: false,
}
```

对 `api.js` 不做接口契约修改，仅确认继续导出：

- `CollectionListBasic`
- `CollectionFoldersBasic`
- `FolderApisBasic`
- `ApisDetailByIds`

**Step 4: Run test to verify it passes**

Run: `npx eslint src/components/Api.vue src/utils/base/api.js`
Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Api.vue web/src/utils/base/api.js
git commit -m "refactor: normalize api tree node state"
```

### Task 2: 将首屏全量加载替换为集合基础信息加载

**Files:**
- Modify: `web/src/components/Api.vue`

**Step 1: Write the failing test**

记录预期行为：`loadCollectionData()` 不再调用 `Api.Collections`，而是调用 `Api.CollectionListBasic`，并把结果映射为空子节点的集合节点。

**Step 2: Run test to verify it fails**

Run: `Select-String -Path 'D:/go/cache_manager_api/web/src/components/Api.vue' -Pattern 'Api.Collections\\(' -Encoding UTF8`
Expected: 找到 `loadCollectionData()` 中仍在使用 `Api.Collections`

**Step 3: Write minimal implementation**

修改 [Api.vue](/D:/go/cache_manager_api/web/src/components/Api.vue) 的 `loadCollectionData()`：

- 将 `Api.Collections({}, ...)` 替换为 `Api.CollectionListBasic({}, ...)`
- 将 `res.Data.list` 映射为 `normalizeCollectionNode`
- 保留 `applyTreeSortCache()` 和 `initTreeExpansion()`
- 删除任何依赖全量 `children` 的首屏假设

**Step 4: Run test to verify it passes**

Run: `npx eslint src/components/Api.vue`
Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Api.vue
git commit -m "refactor: load api collections lazily"
```

### Task 3: 实现集合和文件夹的懒加载入口

**Files:**
- Modify: `web/src/components/Api.vue`

**Step 1: Write the failing test**

先确认当前失败现状：

- 展开集合只写缓存，不拉文件夹
- 展开文件夹依赖 `fillCollectionApis()` 手动加载接口，并且接口使用的是全量 `Api.Apis`

**Step 2: Run test to verify it fails**

Run: `Select-String -Path 'D:/go/cache_manager_api/web/src/components/Api.vue' -Pattern 'handleNodeExpand|fillCollectionApis|Api.Apis\\(' -Encoding UTF8`
Expected: 看到集合展开未触发加载，文件夹接口列表仍走 `Api.Apis`

**Step 3: Write minimal implementation**

在 [Api.vue](/D:/go/cache_manager_api/web/src/components/Api.vue) 新增：

- `loadCollectionFolders(collectionNode)`
- `loadFolderApis(folderNode)`
- `ensureCollectionFoldersLoaded(collectionNode)`
- `ensureFolderApisLoaded(folderNode)`

实现要求：

- 加载前判断 `loading` / `loaded`
- 集合加载走 `Api.CollectionFoldersBasic({ collection_id: collectionNode.id })`
- 文件夹加载走 `Api.FolderApisBasic({ dir_id: folderNode.id })`
- 成功后把返回数据标准化写回 `children`
- 空数组也要标记 `loaded = true`

然后改造：

- `handleNodeExpand(data)`：在更新缓存前或后调用对应 `ensure*Loaded`
- `handleNodeDoubleClick(data)`：对文件夹展开不再直接调用 `fillCollectionApis`
- `fillCollectionApis()`：删除或替换为 `refreshFolderApis(folderNode)`，统一走 `FolderApisBasic`

**Step 4: Run test to verify it passes**

Run: `npx eslint src/components/Api.vue`
Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Api.vue
git commit -m "feat: lazy load api tree children"
```

### Task 4: 将接口点击改为按需拉取详情

**Files:**
- Modify: `web/src/components/Api.vue`
- Modify: `web/src/components/api/ApiDetail.vue`

**Step 1: Write the failing test**

定义目标行为：点击 `api` 节点时，树节点摘要不能直接进入详情页，必须先通过 `ApisDetailByIds` 拉完整详情，再调用 `InitApiDetail`。

**Step 2: Run test to verify it fails**

Run: `Select-String -Path 'D:/go/cache_manager_api/web/src/components/Api.vue' -Pattern 'InitApiDetail\\(data\\)' -Encoding UTF8`
Expected: 看到 `handleNodeClick()` 直接把树节点 `data` 传给详情组件

**Step 3: Write minimal implementation**

在 [Api.vue](/D:/go/cache_manager_api/web/src/components/Api.vue) 中新增：

- `loadingApiDetail: false`
- `loadApiDetail(apiNode)`

实现：

- 调用 `Api.ApisDetailByIds({ ids: [apiNode.id] })`
- 从返回列表取第一项
- 成功后同步更新当前树节点的摘要字段（如名称、方法、URL）
- 调用 `this.$refs.refApiDetail.InitApiDetail(detail)`
- 失败时提示消息并清空或重置右侧接口区域

改造 `handleNodeClick(data)`：

- 对集合/文件夹直接设置 `selectedItem = data`
- 对接口先设置加载态，再调用 `loadApiDetail(data)`

在 [ApiDetail.vue](/D:/go/cache_manager_api/web/src/components/api/ApiDetail.vue) 中补一个可选加载占位入口，至少保证父组件在详情未返回前不会继续展示上一条接口的旧内容。

**Step 4: Run test to verify it passes**

Run: `npx eslint src/components/Api.vue src/components/api/ApiDetail.vue`
Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Api.vue web/src/components/api/ApiDetail.vue
git commit -m "feat: load api detail on demand"
```

### Task 5: 重写展开状态恢复逻辑以适配懒加载

**Files:**
- Modify: `web/src/components/Api.vue`

**Step 1: Write the failing test**

定义目标行为：刷新页面后，缓存中记录的集合和文件夹应恢复展开；恢复过程中允许异步加载，但最终状态必须一致。

**Step 2: Run test to verify it fails**

Run: `Select-String -Path 'D:/go/cache_manager_api/web/src/components/Api.vue' -Pattern 'applyExpandedStateFromCache|expandAllNodes' -Encoding UTF8`
Expected: 当前实现直接遍历现有节点 `expand()`，不能处理“节点尚未加载”的懒加载场景

**Step 3: Write minimal implementation**

在 [Api.vue](/D:/go/cache_manager_api/web/src/components/Api.vue) 中：

- 用 `restoreExpandedNodes()` 替换 `applyExpandedStateFromCache()`
- 第一轮遍历缓存中的集合键，依次 `await ensureCollectionFoldersLoaded(collection)` 后再 `expand()`
- 第二轮遍历缓存中的文件夹键，找到已加载出来的文件夹节点，`await ensureFolderApisLoaded(folder)` 后再 `expand()`
- `initTreeExpansion()` 在懒加载模式下不再使用“首次全部展开”的默认行为，改为：
  - 有缓存则恢复缓存
  - 无缓存则保持仅显示集合层

**Step 4: Run test to verify it passes**

Run: `npx eslint src/components/Api.vue`
Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Api.vue
git commit -m "feat: restore lazy api tree expansion state"
```

### Task 6: 将增删改复制下移改为局部刷新

**Files:**
- Modify: `web/src/components/Api.vue`

**Step 1: Write the failing test**

定义目标行为：

- 创建文件夹 / 删除文件夹后仅刷新所属集合
- 创建接口 / 复制接口 / 删除接口 / 下移接口后仅刷新所属文件夹
- JSON 导入允许保留全量刷新

**Step 2: Run test to verify it fails**

Run: `Select-String -Path 'D:/go/cache_manager_api/web/src/components/Api.vue' -Pattern 'pushUniqueByKey|fillCollectionApis\\(|loadCollectionData\\(' -Encoding UTF8`
Expected: 当前实现仍以本地数组直接插入或全量刷新为主

**Step 3: Write minimal implementation**

在 [Api.vue](/D:/go/cache_manager_api/web/src/components/Api.vue) 中新增：

- `findCollectionNode(collectionId)`
- `findFolderNode(collectionId, folderId)`
- `refreshCollectionFolders(collectionId, { force: true })`
- `refreshFolderApis(collectionId, folderId, { force: true })`

改造以下入口：

- `createNewDir()` 成功后调用 `refreshCollectionFolders`
- `handleFolderDelete()` 成功后调用 `refreshCollectionFolders`
- `handleFolderCreateApi()` 成功后调用 `refreshFolderApis`
- `copyApi()` 成功后调用 `refreshFolderApis`
- `handleApiAction('delete_api')` 成功后调用 `refreshFolderApis`
- `handleApiAction('down_api')` 成功后调用 `refreshFolderApis`
- `handleApiUpdate(api)` 成功后同步树上当前节点摘要
- `apiImportJson()` 允许保留 `loadCollectionData()`

**Step 4: Run test to verify it passes**

Run: `npx eslint src/components/Api.vue`
Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Api.vue
git commit -m "refactor: refresh api tree nodes incrementally"
```

### Task 7: 做手工验证并记录结果

**Files:**
- Modify: `docs/plans/2026-03-18-api-lazy-loading-implementation.md`

**Step 1: Write the failing test**

列出必须验证的行为：

- 首屏只请求 `CollectionListBasic`
- 展开集合时请求 `CollectionFoldersBasic`
- 展开文件夹时请求 `FolderApisBasic`
- 点击接口时请求 `ApisDetailByIds`
- 创建、复制、删除、下移接口后仅刷新目标文件夹
- 刷新页面后展开状态恢复

**Step 2: Run test to verify it fails**

Run: 手工打开开发者工具 Network 面板并操作接口页
Expected: 在实现前会看到 `Collections` 全量请求或错误的全量数据依赖

**Step 3: Write minimal implementation**

将验证结果补到计划末尾的执行记录中，注明：

- 实测触发的接口
- 是否存在重复请求
- 是否还有卡顿点

**Step 4: Run test to verify it passes**

Run: `npx eslint src/components/Api.vue src/components/api/ApiDetail.vue`
Expected: PASS

**Step 5: Commit**

```bash
git add web/src/components/Api.vue web/src/components/api/ApiDetail.vue docs/plans/2026-03-18-api-lazy-loading-implementation.md
git commit -m "docs: verify api lazy loading migration"
```
