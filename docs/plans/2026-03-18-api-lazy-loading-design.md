# 接口开发树按需加载设计

## 背景

当前接口开发页在 [Api.vue](/D:/go/cache_manager_api/web/src/components/Api.vue) 的 `loadCollectionData()` 中直接调用 `Api.Collections` 拉取整棵树，集合、文件夹、接口摘要和接口详情一次性进入前端。集合较多、文件夹较多或接口详情较大时，首屏反序列化和树渲染成本过高，导致页面卡顿。

后端已经提供了分层接口：

- `CollectionListBasic`：集合基础信息
- `CollectionFoldersBasic`：集合下文件夹基础信息
- `FolderApisBasic`：文件夹下接口基础信息
- `ApisDetailByIds`：接口明细

本次改造目标是把左侧树和右侧详情改为分层按需请求，避免全量拉取。

## 目标

- 首屏只加载集合基础信息。
- 展开集合时再加载该集合的文件夹。
- 展开文件夹时再加载该文件夹的接口摘要。
- 点击接口时再加载接口明细。
- 创建、删除、复制、下移后只刷新受影响节点，不再全量重拉整树。
- 保留当前的树展开状态缓存能力。

## 非目标

- 不改后端接口契约。
- 不引入虚拟滚动或新的状态管理库。
- 不在本次改造中重构所有树操作逻辑，只收敛到按需加载所需的最小改动。

## 方案对比

### 方案 A：完全分层按需加载

页面只拉集合，展开集合拉文件夹，展开文件夹拉接口摘要，点击接口拉详情。

优点：

- 与现有后端接口完全匹配。
- 首屏和树操作性能改善最明显。
- 数据粒度清晰，后续局部刷新容易实现。

缺点：

- 需要补齐节点 `loaded/loading` 管理。
- 展开状态恢复需要串行加载，逻辑比当前复杂。

### 方案 B：首屏拉集合和文件夹，接口按需

优点：

- 前端改动稍小。

缺点：

- 文件夹很多时首屏仍然偏重。
- 不符合“完全按需加载”的目标。

### 方案 C：保留全量接口，仅做前端渲染优化

优点：

- 最少改动。

缺点：

- 根因未消除，网络和 JSON 处理成本仍在。

结论：采用方案 A。

## 数据模型调整

在 [Api.vue](/D:/go/cache_manager_api/web/src/components/Api.vue) 中统一给树节点补充以下前端状态字段：

- `type`：`collection` / `folder` / `api`
- `children`：子节点数组；未加载时为空数组
- `loaded`：当前层数据是否已加载
- `loading`：当前节点是否正在加载

集合节点只保留集合级展示字段；文件夹节点只保留文件夹级展示字段；接口节点仅保留树展示所需字段，如 `id/name/method/url/folder_id/collection_id/uniqueid`。

## 数据流设计

### 1. 首屏加载

`loadCollectionData()` 由调用 `Api.Collections` 改为调用 `Api.CollectionListBasic`。

返回结果经前端标准化后写入 `treeData`：

- `children = []`
- `loaded = false`
- `loading = false`
- `type = 'collection'`

### 2. 展开集合

在 `handleNodeExpand(data)` 中，如果节点为集合且 `loaded = false`，调用 `Api.CollectionFoldersBasic({ collection_id: data.id })`。

成功后将返回文件夹映射为 `folder` 节点并填入集合 `children`，然后标记该集合 `loaded = true`。

### 3. 展开文件夹

在 `handleNodeExpand(data)` 中，如果节点为文件夹且 `loaded = false`，调用 `Api.FolderApisBasic({ dir_id: data.id })`。

成功后将返回接口摘要映射为 `api` 节点并填入文件夹 `children`，然后标记该文件夹 `loaded = true`。

### 4. 点击接口

`handleNodeClick(data)` 对 `api` 节点不再直接把树节点传给 [ApiDetail.vue](/D:/go/cache_manager_api/web/src/components/api/ApiDetail.vue)，而是调用 `Api.ApisDetailByIds({ ids: [data.id] })` 获取完整详情。

详情返回后：

- 更新 `selectedItem`
- 调用 `refApiDetail.InitApiDetail(detail)`

右侧接口详情区域需要有加载态，避免点击新接口时短暂保留旧接口内容。

## 展开状态恢复

当前缓存键 `collection:<id>` / `folder:<id>` 可以保留，但恢复逻辑必须改为串行：

1. 读取缓存中需要展开的集合键。
2. 逐个加载并展开集合。
3. 在集合文件夹加载完成后，逐个加载并展开对应文件夹。

不能继续沿用当前“树渲染后直接 `expand()`”的方式，因为懒加载场景下文件夹节点在集合加载前并不存在。

## 局部刷新策略

### 创建集合

保持现有逻辑，可以直接在 `treeData` 追加一个空集合节点。

### 创建文件夹 / 删除文件夹

只刷新对应集合节点的文件夹列表，不调用全量 `loadCollectionData()`。

### 创建接口 / 复制接口 / 删除接口 / 下移接口

只刷新对应文件夹节点的接口摘要列表。

### 更新接口详情

保存详情成功后，不刷新整树；仅同步当前接口节点的摘要字段，例如：

- `name`
- `method`
- `url`

如果接口移动到其他文件夹，当前实现没有覆盖这一类操作，本次不额外扩展。

### JSON 导入

导入可能影响多个文件夹，继续允许走全量集合刷新，作为少数操作的兜底路径。

## 组件影响

### [Api.vue](/D:/go/cache_manager_api/web/src/components/Api.vue)

主要承接全部懒加载、节点标准化、局部刷新和展开缓存恢复逻辑。

### [ApiDetail.vue](/D:/go/cache_manager_api/web/src/components/api/ApiDetail.vue)

保留 `InitApiDetail` 作为详情写入入口；父组件负责先拿到完整详情再调用。必要时增加一个详情加载中的占位态。

### [FolderDetail.vue](/D:/go/cache_manager_api/web/src/components/api/FolderDetail.vue)

依赖 `folder.children` 渲染接口文档；当文件夹节点尚未加载接口摘要时，父组件应保证在选择文件夹或打开文档前已触发该文件夹接口摘要加载。

## 错误处理

- 节点加载失败时只提示当前节点失败，不清空整棵树。
- `loading = true` 时禁止同节点重复发请求。
- 空集合 / 空文件夹视为成功加载，`children = []`，`loaded = true`。
- 接口详情加载失败时，不覆盖当前树结构；右侧提示错误并清空或保持安全占位态。

## 验证要点

- 首屏只请求集合基础信息。
- 首次展开集合才请求文件夹，重复展开不重复请求。
- 首次展开文件夹才请求接口摘要，重复展开不重复请求。
- 点击接口才请求详情。
- 创建、删除、复制、下移接口后，只刷新目标文件夹。
- 页面刷新后可按缓存恢复已展开的集合和文件夹。

## 风险

- 展开恢复改为异步串行后，若实现不稳，会出现刷新后展开状态丢失。
- 现有很多逻辑直接遍历 `treeData[i].children[j].children`，需要统一经过“确保节点已加载”的入口，否则会出现空数组与未加载状态混淆。
- 若 `FolderDetail` 中的接口文档依赖完整接口字段，而 `FolderApisBasic` 仅返回摘要，则需要确认文档页仅基于摘要即可工作；否则需要在文档页单独拉详情。
