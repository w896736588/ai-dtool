# 信息抓取模块接口与状态约定

Date: 2026-03-10

## 1. 用途

本文件补充：

- 后端结构体建议
- AI 抓取规划 Prompt 模板
- 前端页面状态流转
- SSE 事件建议

配套文档：

- [2026-03-10-info-crawl-module-design.md](/c:/work/frog/dev_tool_master/docs/plans/2026-03-10-info-crawl-module-design.md)
- [2026-03-10-info-crawl-implementation.md](/c:/work/frog/dev_tool_master/docs/plans/2026-03-10-info-crawl-implementation.md)

## 2. 后端结构体建议

建议新增文件：

- `internal/app/dtool/struct/info_crawl.go`

### 2.1 任务结构

```go
package _struct

// InfoCrawlTask 信息抓取任务
type InfoCrawlTask struct {
	ID         int    `json:"id"`
	Name       string `json:"name"`
	Prompt     string `json:"prompt"`
	AiModelID  int    `json:"ai_model_id"`
	Status     int    `json:"status"`
	CreateTime int64  `json:"create_time"`
	UpdateTime int64  `json:"update_time"`
}

// InfoCrawlTaskPage 信息抓取任务网页配置
type InfoCrawlTaskPage struct {
	ID                 int    `json:"id"`
	TaskID             int    `json:"task_id"`
	Name               string `json:"name"`
	URL                string `json:"url"`
	Note               string `json:"note"`
	LoginCheckSelector string `json:"login_check_selector"`
	LoginStatus        int    `json:"login_status"`
	UserDataDir        string `json:"user_data_dir"`
	Sort               int    `json:"sort"`
	Status             int    `json:"status"`
	CreateTime         int64  `json:"create_time"`
	UpdateTime         int64  `json:"update_time"`
}
```

### 2.2 执行记录结构

```go
package _struct

// InfoCrawlRun 信息抓取执行记录
type InfoCrawlRun struct {
	ID               int    `json:"id"`
	TaskID           int    `json:"task_id"`
	Status           string `json:"status"`
	RunMessage       string `json:"run_message"`
	PromptSnapshot   string `json:"prompt_snapshot"`
	AiModelSnapshot  string `json:"ai_model_snapshot"`
	PlannerContent   string `json:"planner_content"`
	SummaryContent   string `json:"summary_content"`
	PageTotal        int    `json:"page_total"`
	PageSuccessTotal int    `json:"page_success_total"`
	PageFailedTotal  int    `json:"page_failed_total"`
	CreateTime       int64  `json:"create_time"`
	UpdateTime       int64  `json:"update_time"`
}

// InfoCrawlRunPage 信息抓取网页执行明细
type InfoCrawlRunPage struct {
	ID             int    `json:"id"`
	RunID          int    `json:"run_id"`
	TaskPageID     int    `json:"task_page_id"`
	PageName       string `json:"page_name"`
	URL            string `json:"url"`
	Status         string `json:"status"`
	ErrorMessage   string `json:"error_message"`
	PlannerAction  string `json:"planner_action"`
	ExecuteLog     string `json:"execute_log"`
	RawText        string `json:"raw_text"`
	RawHTML        string `json:"raw_html"`
	ScreenshotPath string `json:"screenshot_path"`
	CreateTime     int64  `json:"create_time"`
	UpdateTime     int64  `json:"update_time"`
}
```

### 2.3 请求结构

```go
package _struct

// InfoCrawlTaskSaveRequest 保存任务请求
type InfoCrawlTaskSaveRequest struct {
	ID        int    `json:"id"`
	Name      string `json:"name"`
	Prompt    string `json:"prompt"`
	AiModelID int    `json:"ai_model_id"`
}

// InfoCrawlTaskPageSaveRequest 保存网页请求
type InfoCrawlTaskPageSaveRequest struct {
	ID                 int    `json:"id"`
	TaskID             int    `json:"task_id"`
	Name               string `json:"name"`
	URL                string `json:"url"`
	Note               string `json:"note"`
	LoginCheckSelector string `json:"login_check_selector"`
	Sort               int    `json:"sort"`
}

// InfoCrawlTaskRunRequest 执行任务请求
type InfoCrawlTaskRunRequest struct {
	TaskID          int    `json:"task_id"`
	SseDistributeID string `json:"sse_distribute_id"`
}
```

### 2.4 规划结构

```go
package _struct

// InfoCrawlPlannerAction AI 规划的单个动作
type InfoCrawlPlannerAction struct {
	Type   string `json:"type"`
	Locator string `json:"locator"`
	Value  string `json:"value"`
	OutKey string `json:"out_key"`
	Tip    string `json:"tip"`
}

// InfoCrawlPlannerPage AI 规划的单网页抓取方案
type InfoCrawlPlannerPage struct {
	TaskPageID int                      `json:"task_page_id"`
	Goal       string                   `json:"goal"`
	Actions    []InfoCrawlPlannerAction `json:"actions"`
}

// InfoCrawlPlannerResult AI 规划结果
type InfoCrawlPlannerResult struct {
	Pages []InfoCrawlPlannerPage `json:"pages"`
}
```

## 3. 常量建议

建议新增文件：

- `internal/app/dtool/define/info_crawl.go`

```go
package define

const InfoCrawlTaskStatusNormal = 1
const InfoCrawlTaskStatusDelete = 0

const InfoCrawlPageLoginStatusNo = 0
const InfoCrawlPageLoginStatusOk = 1
const InfoCrawlPageLoginStatusExpired = 2

const InfoCrawlRunStatusRunning = "running"
const InfoCrawlRunStatusSuccess = "success"
const InfoCrawlRunStatusPartialFailed = "partial_failed"
const InfoCrawlRunStatusFailed = "failed"

const InfoCrawlRunPageStatusSuccess = "success"
const InfoCrawlRunPageStatusFailed = "failed"
const InfoCrawlRunPageStatusLoginRequired = "login_required"

const InfoCrawlPlannerActionWait = "wait"
const InfoCrawlPlannerActionClick = "click"
const InfoCrawlPlannerActionExistWait = "exist_wait"
const InfoCrawlPlannerActionNoExistWait = "no_exist_wait"
const InfoCrawlPlannerActionTextContent = "text_content"
const InfoCrawlPlannerActionBoolResult = "bool_result"
```

## 4. 控制器返回格式建议

延续当前项目风格，返回 `GinResponseSuccess/Error` 即可。

### 4.1 `InfoCrawlTaskInfo`

建议返回：

```json
{
  "task": {
    "id": 1,
    "name": "竞品抓取",
    "prompt": "请重点关注价格和公告",
    "ai_model_id": 2
  },
  "page_list": [
    {
      "id": 11,
      "task_id": 1,
      "name": "官网公告",
      "url": "https://example.com/news",
      "note": "只看最新公告",
      "login_check_selector": ".avatar",
      "login_status": 1
    }
  ],
  "run_list": [
    {
      "id": 101,
      "status": "success",
      "create_time": 1741615200,
      "page_total": 2,
      "page_success_total": 2,
      "page_failed_total": 0
    }
  ]
}
```

### 4.2 `InfoCrawlRunInfo`

建议返回：

```json
{
  "run_info": {
    "id": 101,
    "task_id": 1,
    "status": "success",
    "prompt_snapshot": "请重点关注价格和公告",
    "planner_content": "{...}",
    "summary_content": "### 汇总结果 ..."
  },
  "run_page_list": [
    {
      "id": 1001,
      "task_page_id": 11,
      "page_name": "官网公告",
      "status": "success",
      "planner_action": "抓取最新公告标题与摘要",
      "execute_log": "step1 wait; step2 text_content body",
      "raw_text": "..."
    }
  ]
}
```

## 5. AI 抓取规划 Prompt 模板

### 5.1 系统 Prompt 模板

建议做成固定模板，不让前端直接改：

```text
你是一个网页抓取规划助手，不负责输出结论，只负责为 Playwright 生成结构化抓取计划。

你的任务：
1. 根据任务目标和网页说明，判断每个网页应该抓取哪些信息。
2. 只输出合法 JSON，不要输出 markdown，不要输出解释。
3. 每个网页最多输出 8 个动作。
4. 只能使用以下动作：
   - wait
   - click
   - exist_wait
   - no_exist_wait
   - text_content
   - bool_result
5. 不允许输出 input、goto、evaluate、press、hover 等未授权动作。
6. locator 必须尽量简短稳定。
7. 如果无法确定具体区域，允许直接抓取 body 的 text_content。

输出格式：
{
  "pages": [
    {
      "task_page_id": 1,
      "goal": "一句话说明该页面的抓取目标",
      "actions": [
        {
          "type": "wait",
          "locator": "",
          "value": "1500",
          "out_key": "",
          "tip": "等待页面稳定"
        }
      ]
    }
  ]
}
```

### 5.2 用户 Prompt 模板

建议拼装为：

```text
任务名称：
{{task_name}}

任务目标：
{{task_prompt}}

网页列表：
{{page_descriptions}}

请为每个网页生成抓取计划。
```

其中 `page_descriptions` 建议格式：

```text
网页ID: 11
网页名称: 官网公告
URL: https://example.com/news
网页说明: 重点关注最新公告、发布时间、正文摘要

网页ID: 12
网页名称: 后台价格页
URL: https://example.com/pricing
网页说明: 重点关注当前套餐价格、版本名称、是否有促销信息
```

### 5.3 汇总系统 Prompt 模板

```text
你是一个信息整理助手，请根据多个网页抓取结果输出汇总结论。

要求：
1. 输出中文。
2. 先给整体摘要。
3. 再按网页来源分别总结。
4. 明确标注不确定信息。
5. 不要编造未抓取到的内容。
```

### 5.4 汇总用户 Prompt 模板

```text
任务名称：
{{task_name}}

任务提示词：
{{task_prompt}}

执行时间：
{{run_time}}

网页抓取结果：
{{page_results}}

请输出：
1. 整体摘要
2. 各网页核心信息
3. 需要重点关注的变化
4. 不确定或疑似噪音的信息
```

## 6. 前端状态设计

建议主页面使用一个统一状态对象，风格可参考现有 Vue 组件。

### 6.1 页面状态

```js
const state = reactive({
  taskList: [],
  currentTaskId: 0,
  taskLoading: false,
  taskSaving: false,
  taskForm: {
    id: 0,
    name: '',
    prompt: '',
    ai_model_id: 0,
  },
  pageList: [],
  runList: [],
  aiModelList: [],
  historyDrawerVisible: false,
  runDetailVisible: false,
  currentRunId: 0,
  runDetail: {
    run_info: {},
    run_page_list: [],
  },
  runSubmitting: false,
})
```

### 6.2 页面状态流转

#### 初始化

1. 加载任务列表
2. 默认打开第一个任务
3. 拉取任务详情

#### 任务切换

1. 保存当前编辑内容提示可选
2. 切换 `currentTaskId`
3. 重新加载详情

#### 保存任务

1. 校验 `name`
2. 校验 `ai_model_id`
3. 提交 `InfoCrawlTaskSave`
4. 成功后刷新列表和详情

#### 新增网页

1. 打开行内编辑或弹窗
2. 提交 `InfoCrawlTaskPageSave`
3. 刷新当前任务详情

#### 打开登录页

1. 调用 `InfoCrawlTaskPageOpenLogin`
2. 提示用户去浏览器里完成登录
3. 登录后点击“检查登录状态”

#### 执行任务

1. 校验任务已保存
2. 校验至少有一个网页
3. 调用 `InfoCrawlTaskRun`
4. 打开历史抽屉
5. 刷新执行记录

#### 查看执行详情

1. 调用 `InfoCrawlRunInfo`
2. 打开详情弹窗

## 7. 前端组件通信建议

### 7.1 `InfoCrawl.vue`

负责：

- 统一数据加载
- 与后端 API 通信
- 向子组件分发数据

### 7.2 `TaskList.vue`

Props：

- `task-list`
- `current-task-id`
- `loading`

Events：

- `select-task`
- `create-task`
- `delete-task`

### 7.3 `TaskEditor.vue`

Props：

- `task-form`
- `ai-model-list`
- `saving`
- `running`

Events：

- `save-task`
- `run-task`
- `open-history`

### 7.4 `PageEditor.vue`

Props：

- `page-list`

Events：

- `save-page`
- `delete-page`
- `open-login`
- `check-login`

### 7.5 `RunHistoryDrawer.vue`

Props：

- `visible`
- `run-list`

Events：

- `close`
- `open-detail`
- `refresh`

### 7.6 `RunDetailDialog.vue`

Props：

- `visible`
- `run-detail`

Events：

- `close`

## 8. SSE 事件建议

任务执行较长，建议沿用当前项目 SSE 风格。

### 8.1 事件文本建议

可以直接发文本，不必另建复杂协议。

建议输出示例：

```text
[任务] 开始执行 竞品抓取
[规划] 正在生成抓取计划
[规划] 计划生成完成，共 2 个网页
[网页] 官网公告 开始执行
[网页] 官网公告 步骤1 wait 1500ms
[网页] 官网公告 步骤2 text_content body
[网页] 官网公告 执行完成
[汇总] 正在生成总结
[汇总] 汇总完成
[任务] 执行成功
```

### 8.2 前端处理建议

第一期可以不做复杂流式展示。

更简单的做法：

- 执行时提示“任务已开始”
- 执行结束后刷新历史列表
- 需要时后续再补实时日志面板

## 9. 校验规则建议

### 9.1 任务保存

- `name` 不能为空
- `prompt` 不能为空
- `ai_model_id` 必须大于 0

### 9.2 网页保存

- `task_id` 必须存在
- `name` 不能为空
- `url` 不能为空
- `url` 必须以 `http://` 或 `https://` 开头

### 9.3 执行任务

- 任务必须存在
- 至少存在一个有效网页
- 所有网页不要求都已登录，但未登录网页执行时应记入失败明细

## 10. 建议的最终交付顺序

1. 先按 [2026-03-10-info-crawl-implementation.md](/c:/work/frog/dev_tool_master/docs/plans/2026-03-10-info-crawl-implementation.md) 建表和搭页面骨架。
2. 再按本文件补结构体与接口响应。
3. 然后实现登录态与历史记录。
4. 最后接入 AI 抓取规划和 AI 汇总。

## 11. 到这里为止的文档状态

现在文档已经覆盖：

- 设计边界
- 实施步骤
- SQL 草案
- 接口清单
- 结构体建议
- Prompt 模板
- 前端状态流转

后续如果继续，下一步就不应该再写设计文档了，而应该进入正式编码。
