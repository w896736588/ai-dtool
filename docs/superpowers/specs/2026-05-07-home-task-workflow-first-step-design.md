# 任务工作流首节点改造设计

## 背景

当前首页任务创建流程存在两段割裂体验：

1. 创建任务后仍停留在任务清单，用户需要手动点击 `工作流程` 进入工作流页。
2. 若配置了 `tapd_url`，后端会额外创建一条 `home_task_tapd_scrape` 异步任务，抓取结果需要再确认后才能写回知识片段。

与此同时，[TaskWorkflow.vue](/C:/work/self/cache_manager_api/web/src/components/TaskWorkflow.vue) 当前只提供四个提示词 `tab`，没有真正承载“执行步骤”的视觉结构，也没有把 TAPD 抓取作为工作流的首个节点展示。

## 目标

将任务清单与工作流程改造成统一链路：

1. 首页创建任务成功后，直接新开工作流程页。
2. 工作流程页顶部使用横向“执行节点”替代现有 `tab` 视觉。
3. 工作流新增首个节点 `抓取TAPD需求内容`。
4. 工作流程页首次打开时自动触发首节点执行，并实时展示执行日志。
5. TAPD 抓取完成后，直接将抓取结果写入工作流关联的需求知识片段，不再走异步任务确认流程。
6. 删除“创建任务后自动生成 TAPD 抓取异步任务”的旧流程。

## 非目标

1. 不调整其他工作流节点的业务能力，仅改其导航表现。
2. 不改造全局异步任务中心 UI。
3. 不引入新的 AI 生成逻辑，只复用现有 TAPD 抓取、ZIP 处理、知识片段保存能力。

## 方案概览

### 一、任务创建后的跳转

修改 [HomeTask.vue](/C:/work/self/cache_manager_api/web/src/components/HomeTask.vue) 的保存成功逻辑：

1. 新建任务成功后关闭弹窗。
2. 继续沿用当前“新标签页打开”的交互方式。
3. 直接使用返回的任务 `id` 打开 `/TaskWorkflow/:taskId`。

编辑任务时保持现状，不自动跳转。

### 二、工作流程页改成执行节点

修改 [TaskWorkflow.vue](/C:/work/self/cache_manager_api/web/src/components/TaskWorkflow.vue)：

1. 去掉 `el-tabs` 作为主容器。
2. 新增横向步骤条，节点顺序固定为：
   - `requirement-fetch` 抓取TAPD需求
   - `requirement` 需求文档 MD
   - `design` 开发设计
   - `api-dev` 接口开发生成
   - `api-test-fix` 接口自动化测试修复
3. 点击节点只切换内容区，不改变路由。
4. 页面默认进入第一个节点。

现有四个提示词编辑区保留，只是挂到新节点内容区下。

### 三、首节点业务改造

新增“工作流首节点抓取”接口，由工作流程页直接触发，不再借助首页异步任务。

建议新增接口：

- `POST /api/task/workflow/requirement/fetch`

接口职责：

1. 根据 `workflow_id` 加载工作流和首页任务。
2. 校验 `tapd_url`、需求知识片段、TAPD 抓取配置。
3. 更新工作流首节点状态为运行中。
4. 通过现有 `dispatchScrapeTaskAndAwait` 执行抓取。
5. 通过现有 `processScrapeZipResult` 处理 ZIP，拿到 markdown。
6. 直接将 markdown 覆盖写入 `requirement_fragment_id` 对应知识片段。
7. 更新工作流首节点状态为成功或失败。
8. 通过 SSE 推送步骤日志。

### 四、去除旧异步任务确认链路

修改 [home_task.go](/C:/work/self/cache_manager_api/internal/app/dtool/controller/home_task.go)：

1. 保留创建知识片段逻辑。
2. 删除新建任务后自动调用 `createAsyncTask(asyncTaskTypeHomeTaskTapdScrape, ...)` 的逻辑。
3. 新建任务成功后，额外确保已存在 `tbl_task_workflow` 记录，避免工作流程页首次打开再补建时状态不一致。

旧的 `home_task_tapd_scrape` 异步任务类型保留兼容，不再由任务创建入口触发。

### 五、工作流状态字段

当前 `tbl_task_workflow` 只有通用字段：

- `status`
- `current_stage`
- `last_error`

它们不足以稳定支撑首节点单独展示与自动恢复，因此新增以下字段：

- `requirement_fetch_status`
- `requirement_fetch_started_at`
- `requirement_fetch_finished_at`
- `requirement_fetch_error`
- `requirement_source_url`

状态建议值：

- `idle`
- `running`
- `success`
- `failed`

这些字段只服务“抓取TAPD需求内容”节点，不扩散到其他节点。

### 六、SSE 推送设计

不复用 `async_tasks` 分发语义，新增工作流专用 SSE 分发通道：

- `task_workflow_<workflowId>`

首节点执行过程中推送结构化消息，至少包含：

- `workflow_id`
- `step`
- `status`
- `message`
- `time`

推荐步骤：

1. `load_config`
2. `dispatch_scrape`
3. `wait_result`
4. `process_zip`
5. `save_fragment`
6. `done`
7. `error`

前端进入工作流程页时注册该分发 ID；离开页面时注销监听。

### 七、首节点 UI 展示

首节点内容区至少展示：

1. 当前抓取状态。
2. TAPD 地址。
3. 关联知识片段标题/ID。
4. 抓取配置摘要：
   - `smart_link_id`
   - `label`
   - `css_selector`
   - `wait_seconds`
5. 实时日志流。
6. 最近一次抓取完成时间。
7. 最近一次错误信息。

自动执行规则：

1. 页面加载完成后，如果 `tapd_url` 为空，则仅提示未配置 TAPD 地址，不自动执行。
2. 若 `requirement_fetch_status` 为 `success`，默认只展示结果，不重复自动抓取。
3. 若状态为 `idle` 或 `failed`，页面首次进入时自动执行一次。
4. 若状态已是 `running`，页面只接管 SSE 展示，不再重复下发。

### 八、工作流详情返回字段

`/api/task/workflow/create_or_get` 与 `/api/task/workflow/info` 返回体保持现有结构，同时确保 `workflow` 中带回新增首节点字段，供前端直接渲染。

必要时可在返回体中补充一个只读对象：

- `requirement_fetch_config`

包含当前 TAPD 抓取配置快照，避免前端单独查设置接口。

## 数据流

### 新建任务

1. 用户在任务清单提交表单。
2. 后端创建或复用需求知识片段。
3. 后端保存首页任务。
4. 后端确保创建工作流记录。
5. 前端提示成功，并新开 `/TaskWorkflow/:taskId`。

### 首节点执行

1. 工作流程页 `create_or_get` 成功。
2. 前端默认激活 `requirement-fetch`。
3. 前端注册 `task_workflow_<workflowId>` SSE 回调。
4. 若状态允许自动执行，则调用 `/api/task/workflow/requirement/fetch`。
5. 后端推送步骤日志。
6. 抓取完成后，后端直接覆盖写入需求知识片段。
7. 前端刷新工作流详情，并在后续节点中使用最新知识片段内容。

## 错误处理

### TAPD 地址缺失

首节点不执行，展示“当前任务未配置 TAPD 地址”。

### 抓取配置缺失

接口返回错误并写入 `requirement_fetch_error`，前端展示错误。

### 抓取失败

状态置为 `failed`，保留错误文案和开始/结束时间；用户可手动点击“重新抓取”再次触发。

### 知识片段不存在

接口返回错误，不尝试隐式重建，避免覆盖到错误片段。

## 涉及文件

### 后端

- [internal/app/dtool/controller/home_task.go](/C:/work/self/cache_manager_api/internal/app/dtool/controller/home_task.go)
- [internal/app/dtool/controller/task_workflow.go](/C:/work/self/cache_manager_api/internal/app/dtool/controller/task_workflow.go)
- [internal/app/dtool/common/task_workflow.go](/C:/work/self/cache_manager_api/internal/app/dtool/common/task_workflow.go)
- [internal/app/dtool/struct/task_workflow.go](/C:/work/self/cache_manager_api/internal/app/dtool/struct/task_workflow.go)
- [internal/app/dtool/router.go](/C:/work/self/cache_manager_api/internal/app/dtool/router.go)
- [internal/app/dtool/database/2026/05/20260507170000-task_workflow_requirement_fetch.sql](/C:/work/self/cache_manager_api/internal/app/dtool/database/2026/05/20260507170000-task_workflow_requirement_fetch.sql)

### 前端

- [web/src/components/HomeTask.vue](/C:/work/self/cache_manager_api/web/src/components/HomeTask.vue)
- [web/src/components/TaskWorkflow.vue](/C:/work/self/cache_manager_api/web/src/components/TaskWorkflow.vue)
- [web/src/utils/base/task_workflow.js](/C:/work/self/cache_manager_api/web/src/utils/base/task_workflow.js)
- [web/src/utils/base/sse_distribute.js](/C:/work/self/cache_manager_api/web/src/utils/base/sse_distribute.js)（仅在需要封装辅助方法时修改）

## 验收标准

1. 新建任务成功后，会自动新开工作流程页。
2. 工作流程页顶部显示横向执行节点，而不是旧的 tab 视觉。
3. 第一个节点为“抓取TAPD需求内容”。
4. 页面首次打开时会自动触发抓取，并实时看到日志。
5. 抓取完成后，需求知识片段内容已被直接覆盖更新。
6. 首页创建任务后不再出现 `home_task_tapd_scrape` 异步确认任务。
7. 抓取失败时，页面能展示错误并支持手动重试。
