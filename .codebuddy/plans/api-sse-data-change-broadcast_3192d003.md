---
name: api-sse-data-change-broadcast
overview: 为 API 管理模块（集合/文件夹/接口）增加 SSE 实时推送，当 CRUD 操作完成后广播变更通知给所有已连接的 SSE 客户端，前端注册回调后自动刷新数据。
todos:
  - id: add-sse-constant
    content: 在 define/sse.go 中新增 SseApiDataChange 常量
    status: completed
  - id: create-api-sse
    content: 新建 controller/api_sse.go，实现 BroadcastApiChange 广播函数
    status: completed
    dependencies:
      - add-sse-constant
  - id: modify-api-controller
    content: 在 api.go 各 CRUD 方法中添加 goroutine 广播调用
    status: completed
    dependencies:
      - create-api-sse
  - id: modify-api-vue
    content: 在 Api.vue 中注册/注销 SSE 回调并实现 handleApiChangeSSE 刷新逻辑
    status: completed
    dependencies:
      - modify-api-controller
---

## 产品概述

为 API 管理模块（集合/文件夹/接口）增加 SSE 实时推送能力，当后端数据发生 CRUD 变更时，主动通知所有已连接的前端客户端刷新数据。

## 核心功能

- 后端在 API 集合、文件夹、接口的增删改操作完成后，通过 SSE 广播变更通知给所有已连接的前端客户端
- 前端 Api.vue 组件注册 SSE 回调，收到变更通知后自动调用对应的刷新方法
- 推送通知中包含变更类型（collection/folder/api 的 created/updated/deleted/moved 等）和受影响的 ID，前端据此精确刷新
- 前端通过 source_client_id 跳过自身触发的变更，避免重复刷新
- 保留现有编辑后直接使用返回数据的模式不变

## 技术栈

- 后端: Go + Gin + 已有 SSE 基础设施（`gsgin.SseStatus()` + `gsgin.SseGetByClientId()` + `p_define.SseData`）
- 前端: Vue 3 + 已有 SSE 分发管理（`sse_distribute.js` 的 `RegisterReceive` / `UnRegisterReceive`）

## 实现方案

### 架构设计

复用已有的 SSE 广播模式（与 `broadcastMemoryFragmentEvent`、`BroadcastSmartLinkClientStatusUpdate` 一致）：

1. 后端新增 `api_sse.go` 模块，封装广播函数 `BroadcastApiChange`
2. 广播函数通过 `gsgin.SseStatus()` 获取所有已连接客户端，遍历并通过 `gsgin.SseGetByClientId()` 发送消息
3. 各 CRUD controller 方法在操作成功后，以 goroutine 方式调用广播函数
4. 前端在 Api.vue 的 mounted/beforeUnmount 中注册/注销 `api_data_change` 通道回调

### 关键设计决策

- **无需自建客户端注册表**：直接复用 `gsgin.SseStatus()` 获取所有已连接 SSE 客户端
- **无需修改 router.go**：广播复用已有 SSE 连接池，不需要在 openFunc/closeFunc 做额外注册
- **广播使用 goroutine**：避免阻塞 HTTP 响应
- **前端跳过自身**：通过 `source_client_id === sse.GetSseClientId()` 判断

### 推送数据格式

```
{
  "source_client_id": "sse_client_id_xxx",
  "change_type": "api_created",
  "collection_id": 1,
  "folder_id": 2,
  "api_id": 3,
  "old_folder_id": null
}
```

### 目录结构

```
internal/app/dtool/
├── define/
│   └── sse.go                    # [MODIFY] 新增 SseApiDataChange 常量
├── controller/
│   ├── api_sse.go                # [NEW] API变更SSE广播辅助模块
│   └── api.go                    # [MODIFY] 各CRUD方法中添加广播调用

web/src/
├── components/
│   └── Api.vue                   # [MODIFY] 注册SSE回调，实现自动刷新逻辑
```

### 关键文件说明

**`internal/app/dtool/define/sse.go` [MODIFY]**

- 新增 `SseApiDataChange = "api_data_change"` 常量

**`internal/app/dtool/controller/api_sse.go` [NEW]**

- `BroadcastApiChange(sourceClientId, changeType string, ids map[string]any)` — 遍历 `gsgin.SseStatus()` 获取所有客户端，通过 `gsgin.SseGetByClientId()` 获取 SSE 连接并发送 `p_define.SseData` 消息
- 发送失败静默忽略（连接可能已断开）
- 参照 `memory_fragment.go` 的 `broadcastMemoryFragmentEvent` 实现模式

**`internal/app/dtool/controller/api.go` [MODIFY]**

- 每个成功 CRUD 操作后，从 `c.GetHeader("SseClientId")` 获取来源客户端 ID
- 以 goroutine 方式调用 `BroadcastApiChange`，不阻塞 HTTP 响应
- 涉及方法：`ApiCreateCollection`、`ApiDeleteCollection`、`ApiCreateDir`、`ApiDeleteDir`、`ApiCreateApi`、`ApiDeleteApi`、`ApiWeightDown`、`ApiBatchImport`、`ApiMoveApi`

**`web/src/components/Api.vue` [MODIFY]**

- 导入 `sse_distribute`
- mounted() 中注册 `sse.RegisterReceive('api_data_change', this.handleApiChangeSSE)`
- beforeUnmount() 中注销 `sse.UnRegisterReceive('api_data_change')`
- handleApiChangeSSE 方法：比较 source_client_id 跳过自身，根据 change_type 调用 loadCollectionData / refreshCollectionFolders / refreshFolderApis

### 实现要点

- 广播模式严格参照 `memory_fragment.go` 的 `broadcastMemoryFragmentEvent`：`gsgin.SseStatus()` 遍历 + `gsgin.SseGetByClientId()` 发送
- 前端 `main.js` 已配置 axios interceptor 自动在每个请求 header 中携带 `SseClientId`
- ApiMoveApi 广播 `api_moved` 时需同时携带 `old_folder_id` 和新 `folder_id`，刷新源文件夹和目标文件夹
- ApiBatchImport 广播 `batch_imported`，仅携带 `collection_id`，触发整个集合刷新
- 保留所有现有的手动刷新逻辑不变

## Agent Extensions

### SubAgent

- **code-explorer**: 在实现过程中用于搜索跨文件的依赖关系和确认代码模式，确保广播逻辑与现有 SSE 基础设施一致