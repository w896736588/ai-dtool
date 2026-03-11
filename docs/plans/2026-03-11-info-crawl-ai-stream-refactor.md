# 信息抓取改为 AI 流式输出改造方案

Date: 2026-03-11

## 1. 本次调整目标

将当前“信息抓取”模块从“任务 + 网页配置 + Playwright 抓取 + AI 规划/汇总”调整为：

- 仅配置：`模型`、`提示词`
- 移除：`网页配置`、`登录检测`、`网页级执行记录`
- 运行时：后端直接调用 AI 的流式接口
- 输出方式：后端通过项目现有 SSE 分发给前端
- 前端：实时查看 AI 输出内容
- 数据处理：**不兼容老数据**，直接通过 SQL 重建本模块表结构

## 2. 基于现有代码的现状

当前实现仍然是“网页抓取型”设计：

- 后端任务接口：`internal/app/dtool/controller/info_crawl.go`
- 后端数据层：`internal/app/dtool/common/info_crawl.go`
- AI 调用：`internal/app/dtool/common/info_crawl_ai.go`
- Playwright 规划器：`internal/app/dtool/plw/info_crawl_planner.go`
- Playwright 执行器：`internal/app/dtool/plw/info_crawl_runner.go`
- 数据结构：`internal/app/dtool/struct/info_crawl.go`
- 常量：`internal/app/dtool/define/info_crawl.go`
- 路由：`internal/app/dtool/router.go:248`
- 前端页面：`web/src/components/InfoCrawl.vue:1`
- 前端 API：`web/src/utils/base/info_crawl.js:1`

当前数据库存在以下表：

- `tbl_info_crawl_task`
- `tbl_info_crawl_task_page`
- `tbl_info_crawl_run`
- `tbl_info_crawl_run_page`

这套结构明显围绕“网页列表 + 抓取计划 + 网页执行明细”展开，不适合本次“仅模型 + 提示词 + AI 流式输出”的目标。

## 3. 本次改造后的目标结构

### 3.1 任务模型

保留任务概念，但任务仅包含：

- `name`：任务名称
- `prompt`：提示词
- `ai_model_id`：模型配置

不再包含：

- 网页列表
- 登录状态
- 登录目录
- 页面排序
- 页面说明

### 3.2 执行模型

一次执行只记录：

- 任务快照
- 模型快照
- 运行状态
- AI 实时输出的最终结果
- 错误信息

不再记录：

- AI 抓取规划内容
- 网页明细
- 网页原始文本
- 网页截图
- 网页动作日志

### 3.3 运行方式

运行时链路改为：

1. 前端点击“执行任务”
2. 后端创建 `run` 记录
3. 后端调用 AI 流式接口（OpenAI 兼容 SSE）
4. 后端边接收 chunk，边通过项目 SSE 推送给前端
5. 前端实时追加输出内容
6. 流结束后后端落库最终结果，前端刷新执行详情

## 4. 重要前提

本方案按你的描述默认采用以下前提：

- **不再使用 Playwright 浏览器抓取网页**
- **不再维护网页配置**
- **AI 模型本身具备联网/搜索/抓取能力，或者提示词中的信息来源由模型自行处理**

如果你的真实需求是“仍保留浏览器访问网页，只是把 AI 汇总改为流式输出”，那将是另一套改法，本方案不覆盖。

## 5. 数据库改造方案

由于你明确说明：

- 不需要处理老数据
- 直接改数据库生成 SQL

因此建议 **直接删除旧表并重建为最小可用结构**。

### 5.1 新 SQL 文件

建议新增文件：

- `internal/app/dtool/database/2026/03/20260311.信息抓取-AI流式改造.sql`

### 5.2 SQL 内容

```sql
DROP INDEX IF EXISTS "idx_info_crawl_run_page_run_task_page";
DROP INDEX IF EXISTS "idx_info_crawl_run_task_time";
DROP INDEX IF EXISTS "idx_info_crawl_task_page_task_status_sort";
DROP INDEX IF EXISTS "idx_info_crawl_task_status_update";

DROP TABLE IF EXISTS "tbl_info_crawl_run_page";
DROP TABLE IF EXISTS "tbl_info_crawl_task_page";
DROP TABLE IF EXISTS "tbl_info_crawl_run";
DROP TABLE IF EXISTS "tbl_info_crawl_task";

CREATE TABLE IF NOT EXISTS "tbl_info_crawl_task"
(
    "id"           INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "name"         TEXT    NOT NULL DEFAULT '',
    "prompt"       TEXT    NOT NULL DEFAULT '',
    "ai_model_id"  INTEGER NOT NULL DEFAULT 0,
    "status"       INTEGER NOT NULL DEFAULT 1,
    "create_time"  INTEGER NOT NULL DEFAULT 0,
    "update_time"  INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "tbl_info_crawl_run"
(
    "id"                 INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "task_id"            INTEGER NOT NULL DEFAULT 0,
    "status"             TEXT    NOT NULL DEFAULT 'running',
    "run_message"        TEXT    NOT NULL DEFAULT '',
    "prompt_snapshot"    TEXT    NOT NULL DEFAULT '',
    "ai_model_snapshot"  TEXT    NOT NULL DEFAULT '',
    "output_content"     TEXT    NOT NULL DEFAULT '',
    "error_message"      TEXT    NOT NULL DEFAULT '',
    "create_time"        INTEGER NOT NULL DEFAULT 0,
    "update_time"        INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS "idx_info_crawl_task_status_update"
    ON "tbl_info_crawl_task" ("status", "update_time");

CREATE INDEX IF NOT EXISTS "idx_info_crawl_run_task_time"
    ON "tbl_info_crawl_run" ("task_id", "create_time");
```

### 5.3 说明

- 旧的 `tbl_info_crawl_task_page`、`tbl_info_crawl_run_page` 全部删除
- 旧的 `planner_content`、`summary_content` 字段全部删除
- 新结果统一落到 `tbl_info_crawl_run.output_content`
- 失败原因统一落到 `tbl_info_crawl_run.error_message`

## 6. 后端改造清单

## 6.1 `internal/app/dtool/define/info_crawl.go`

保留：

- `InfoCrawlTaskStatusDelete`
- `InfoCrawlTaskStatusNormal`
- `InfoCrawlRunStatusRunning`
- `InfoCrawlRunStatusSuccess`
- `InfoCrawlRunStatusFailed`

删除：

- 所有 `Page` 相关常量
- 所有 `PlannerAction` 相关常量
- `InfoCrawlRunStatusPartialFailed`
- `InfoCrawlPageTextMaxLength`
- `InfoCrawlSummaryInputMaxLength`

建议新增：

```go
const InfoCrawlSseTypeStatus = `info_crawl_status`
const InfoCrawlSseTypeChunk = `info_crawl_chunk`
const InfoCrawlSseTypeDone = `info_crawl_done`
const InfoCrawlSseTypeError = `error`
```

## 6.2 `internal/app/dtool/struct/info_crawl.go`

保留：

- `InfoCrawlTask`
- `InfoCrawlRun`
- `InfoCrawlTaskSaveRequest`
- `InfoCrawlTaskRunRequest`

删除：

- `InfoCrawlTaskPage`
- `InfoCrawlRunPage`
- `InfoCrawlTaskPageSaveRequest`
- `InfoCrawlPlannerAction`
- `InfoCrawlPlannerPage`
- `InfoCrawlPlannerResult`

建议把 `InfoCrawlRun` 改为：

```go
// InfoCrawlRun 信息抓取执行记录。
type InfoCrawlRun struct {
    ID              int    `json:"id"`
    TaskID          int    `json:"task_id"`
    Status          string `json:"status"`
    RunMessage      string `json:"run_message"`
    PromptSnapshot  string `json:"prompt_snapshot"`
    AiModelSnapshot string `json:"ai_model_snapshot"`
    OutputContent   string `json:"output_content"`
    ErrorMessage    string `json:"error_message"`
    CreateTime      int64  `json:"create_time"`
    UpdateTime      int64  `json:"update_time"`
}
```

## 6.3 `internal/app/dtool/common/info_crawl.go`

### 需要保留的方法

- `InfoCrawlTaskList`
- `InfoCrawlTaskRow`
- `InfoCrawlTaskSave`
- `InfoCrawlTaskDelete`
- `InfoCrawlRunCreate`
- `InfoCrawlRunUpdate`
- `InfoCrawlRunList`
- `InfoCrawlRunInfo`
- `InfoCrawlAiModelInfo`

### 需要删除的方法

- `InfoCrawlTaskPageList`
- `InfoCrawlTaskPageRow`
- `InfoCrawlTaskPageSave`
- `InfoCrawlTaskPageDelete`
- `InfoCrawlTaskPageSetLoginStatus`
- `InfoCrawlBuildUserDataDir`
- `InfoCrawlValidatePlanner`
- `InfoCrawlNormalizePlannerMap`
- `InfoCrawlSortPages`
- `infoCrawlFillPageStatusDesc`

### 需要调整的方法

#### `InfoCrawlTaskInfo`

返回值从：

```json
{
  "task": {},
  "page_list": [],
  "run_list": []
}
```

调整为：

```json
{
  "task": {},
  "run_list": []
}
```

#### `InfoCrawlRunCreate`

插入字段改为：

- `task_id`
- `status`
- `run_message`
- `prompt_snapshot`
- `ai_model_snapshot`
- `output_content`
- `error_message`
- `create_time`
- `update_time`

#### `InfoCrawlRunInfo`

返回值从：

```json
{
  "run_info": {},
  "run_page_list": []
}
```

改为：

```json
{
  "run_info": {}
}
```

## 6.4 `internal/app/dtool/common/info_crawl_ai.go`

当前仅有非流式方法：

- `InfoCrawlChatByModel`

本次建议保留该方法作为兼容/回退，同时新增流式方法：

```go
// InfoCrawlChatStreamByModel 使用模型发起流式 AI 请求。
func (h *CSqlite) InfoCrawlChatStreamByModel(
    modelID int,
    systemPrompt string,
    userPrompt string,
    onChunk func(string),
) (string, map[string]any, error)
```

### 实现要求

请求体增加：

```json
{
  "model": "xxx",
  "stream": true,
  "messages": [
    {"role": "system", "content": "..."},
    {"role": "user", "content": "..."}
  ]
}
```

解析逻辑：

- 按行读取响应体
- 只处理 `data:` 开头内容
- 遇到 `data: [DONE]` 结束
- 从 JSON 中提取 `choices[0].delta.content`
- 每拿到一段内容立即回调 `onChunk`
- 同时在后端把所有 chunk 拼成最终字符串返回

### 系统提示词建议

由于现在不再有网页配置，建议固定系统提示词，不再拆“规划”和“汇总”两阶段：

```text
你是一个信息抓取与整理助手。
请严格根据用户提示词完成信息收集与结果整理。
如果模型具备联网/搜索能力，可自行检索并整理结果。
输出使用中文，内容尽量结构化，避免无依据编造。
```

## 6.5 `internal/app/dtool/controller/info_crawl.go`

### 保留接口

- `InfoCrawlTaskList`
- `InfoCrawlTaskInfo`
- `InfoCrawlTaskSave`
- `InfoCrawlTaskDelete`
- `InfoCrawlTaskRun`
- `InfoCrawlRunList`
- `InfoCrawlRunInfo`

### 删除接口

- `InfoCrawlTaskPageSave`
- `InfoCrawlTaskPageDelete`
- `InfoCrawlTaskPageOpenLogin`
- `InfoCrawlTaskPageCheckLogin`

### `InfoCrawlTaskRun` 新逻辑

改造后流程应为：

1. 校验 `task_id`
2. 读取任务
3. 读取 AI 模型配置
4. 创建 `run` 记录
5. 立即返回 `run_id`
6. 开 goroutine 调用 `runInfoCrawlTaskAsync`

### `runInfoCrawlTaskAsync` 新逻辑

旧逻辑包含：

- 生成抓取计划
- 顺序抓取网页
- 汇总多网页结果

这些逻辑全部删除，改为：

1. SSE 推送“任务开始”
2. 调用 `InfoCrawlChatStreamByModel`
3. 每收到 chunk：
   - 通过 SSE 推送 `info_crawl_chunk`
   - 追加到本地 `outputBuilder`
4. 完成后更新 `tbl_info_crawl_run.output_content`
5. SSE 推送 `info_crawl_done`
6. 若失败则更新 `status=failed` 和 `error_message`

### 建议伪代码

```go
func runInfoCrawlTaskAsync(runID int, taskInfo, modelInfo map[string]any, sse *p_sse.SseShell) {
    sse.Send("开始执行", define.InfoCrawlSseTypeStatus)

    systemPrompt := "你是一个信息抓取与整理助手..."
    userPrompt := cast.ToString(taskInfo[`prompt`])

    content, _, err := common.DbMain.InfoCrawlChatStreamByModel(
        cast.ToInt(taskInfo[`ai_model_id`]),
        systemPrompt,
        userPrompt,
        func(chunk string) {
            if strings.TrimSpace(chunk) == `` {
                return
            }
            sse.Send(chunk, define.InfoCrawlSseTypeChunk)
        },
    )
    if err != nil {
        _ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
            `status`:        define.InfoCrawlRunStatusFailed,
            `run_message`:   `执行失败`,
            `error_message`: err.Error(),
        })
        sse.Send(err.Error(), define.InfoCrawlSseTypeError)
        return
    }

    _ = common.DbMain.InfoCrawlRunUpdate(runID, map[string]any{
        `status`:         define.InfoCrawlRunStatusSuccess,
        `run_message`:    `执行完成`,
        `output_content`: content,
        `error_message`:  ``,
    })
    sse.Send(`执行完成`, define.InfoCrawlSseTypeDone)
}
```

## 6.6 删除不再需要的 Playwright 依赖

以下文件将不再被信息抓取模块使用：

- `internal/app/dtool/plw/info_crawl_planner.go`
- `internal/app/dtool/plw/info_crawl_runner.go`

这两个文件可以：

- 保留但不再引用；或
- 直接删除

建议本次先 **保留文件、移除调用**，这样对项目整体影响更小。

## 6.7 `internal/app/dtool/router.go`

需要删除以下路由：

- `POST /api/InfoCrawlTaskPageSave`
- `POST /api/InfoCrawlTaskPageDelete`
- `POST /api/InfoCrawlTaskPageOpenLogin`
- `POST /api/InfoCrawlTaskPageCheckLogin`

保留：

- `POST /api/InfoCrawlTaskList`
- `POST /api/InfoCrawlTaskInfo`
- `POST /api/InfoCrawlTaskSave`
- `POST /api/InfoCrawlTaskDelete`
- `POST /api/InfoCrawlTaskRun`
- `POST /api/InfoCrawlRunList`
- `POST /api/InfoCrawlRunInfo`

## 7. 前端改造清单

## 7.1 `web/src/utils/base/info_crawl.js`

保留：

- `InfoCrawlTaskList`
- `InfoCrawlTaskInfo`
- `InfoCrawlTaskSave`
- `InfoCrawlTaskDelete`
- `InfoCrawlTaskRun`
- `InfoCrawlRunList`
- `InfoCrawlRunInfo`

删除：

- `InfoCrawlTaskPageSave`
- `InfoCrawlTaskPageDelete`
- `InfoCrawlTaskPageOpenLogin`
- `InfoCrawlTaskPageCheckLogin`

## 7.2 `web/src/components/InfoCrawl.vue`

### 页面结构调整

当前页面包含两个编辑卡片：

- 基础信息
- 网页配置

改造后只保留：

- 基础信息（任务名称、AI 模型、提示词）
- 实时输出区域
- 执行历史

### 需要删除的界面能力

- 网页配置列表
- 新增网页按钮
- 删除网页按钮
- 打开登录页
- 检查登录状态

### 需要修改的数据结构

删除：

- `pageList`
- `tempPageIndex`

保留：

- `taskList`
- `taskForm`
- `runList`
- `runDetail`
- `runLiveLog`
- `runSseDistributeId`

### `runTask` 方法调整

删除校验：

```js
if (this.pageList.length === 0) {
  this.$helperNotify.error('至少需要一个网页配置')
  return
}
```

提交后前端仍通过现有 `sse_distribute` 收消息，但要按新的事件类型处理：

- `info_crawl_status`：更新状态文本
- `info_crawl_chunk`：追加内容到 `runLiveLog`
- `info_crawl_done`：结束执行并刷新详情
- `error`：显示错误信息

### SSE 处理建议

当前 `registerRunSse` 是把所有文本直接拼到日志区。

改造后建议：

- 状态类消息显示在状态条
- chunk 类消息直接拼接正文区域
- done 后调用一次 `InfoCrawlRunInfo`

示例处理逻辑：

```js
sseDistribute.RegisterReceive(sseDistributeId, (msg, msgType) => {
  if (msgType === 'info_crawl_status') {
    this.runLiveStatus = msg || '执行中'
    return
  }
  if (msgType === 'info_crawl_chunk') {
    this.runLiveLog += (msg || '').replace(/\r/g, '')
    this.runLiveStatus = 'AI 正在输出'
    return
  }
  if (msgType === 'info_crawl_done') {
    this.runLiveStatus = msg || '执行完成'
    this.fetchRunStatus()
    return
  }
  if (msgType === 'error') {
    this.runLiveStatus = '执行失败'
    this.runLiveLog += `\n[错误] ${(msg || '').replace(/\r/g, '')}`
  }
})
```

### 执行详情弹窗调整

当前详情包含：

- 执行摘要
- 任务提示词快照
- 抓取计划
- 网页结果

改造后应调整为：

- 执行状态
- 任务提示词快照
- 最终输出结果
- 错误信息（如果有）

即：

- 删除 `planner_content` 展示
- 删除 `run_page_list` 展示
- 新增 `output_content` 展示
- 新增 `error_message` 展示

## 8. SSE 事件约定

建议统一使用以下事件类型：

### 8.1 状态事件

类型：`info_crawl_status`

示例消息：

- `任务已提交`
- `正在连接 AI`
- `AI 正在输出`
- `正在写入执行结果`

### 8.2 内容事件

类型：`info_crawl_chunk`

内容：AI 返回的增量文本片段

示例：

- `根据公开信息，`
- `该产品近期主要变化如下：`

### 8.3 完成事件

类型：`info_crawl_done`

示例消息：

- `执行完成`

### 8.4 错误事件

类型：`error`

示例消息：

- `AI 请求失败: xxx`
- `读取流式响应失败: xxx`

## 9. 推荐落地顺序

1. 先执行数据库 SQL，重建信息抓取表
2. 精简 `define` 与 `struct`
3. 精简 `common/info_crawl.go`
4. 在 `common/info_crawl_ai.go` 增加流式调用方法
5. 重写 `controller/info_crawl.go` 中的运行链路
6. 删除路由中的网页配置接口
7. 精简前端 `info_crawl.js`
8. 重构 `InfoCrawl.vue` 页面
9. 联调 SSE 实时输出

## 10. 风险点

### 10.1 模型是否真的具备联网能力

这是本次改造的最大前提。

如果模型本身不具备联网/搜索能力，那么“信息抓取”将退化为“普通提示词生成”，不能替代原先 Playwright 抓网页。

### 10.2 OpenAI 兼容接口的流式格式差异

虽然大多数兼容服务都走 SSE，但字段结构可能略有差异，建议优先按：

- `choices[0].delta.content`

解析；如果后续接入特定厂商，再做兼容扩展。

### 10.3 前端不要再依赖轮询作为主链路

本次核心要求是“运行时通过 SSE 调用 AI，前端实时查看输出”，所以：

- **SSE 是主链路**
- `InfoCrawlRunInfo` 查询只作为完成后的结果刷新

## 11. 结论

这次改造本质上不是“小修”，而是把当前模块从：

- **网页抓取任务系统**

改为：

- **AI 流式执行任务系统**

因此最合适的做法是：

- 直接砍掉网页配置相关表、接口、页面和执行逻辑
- 保留任务与执行历史两个核心对象
- 新增 AI 流式调用方法
- 用现有 SSE 分发机制实时推送输出

这样改动路径最直，且不会为了兼容旧设计增加额外复杂度。
