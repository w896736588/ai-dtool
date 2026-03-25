# 信息抓取模块设计方案

Date: 2026-03-10

## 1. 目标

新增一个独立的“信息抓取”模块，用于：

- 创建多个抓取任务
- 每个任务维护多个抓取网页
- 每个网页支持手动登录
- 手动触发任务执行
- 使用 Playwright 打开页面并抓取内容
- 把多个网页抓取结果交给 AI 按提示词进行汇总
- 保存执行历史，便于回看

本方案基于当前项目已有能力落地：

- 浏览器能力复用现有 `internal/app/dtool/plw`
- AI 配置复用现有 `tbl_ai_provider`、`tbl_ai_model`
- 前端延续当前 `Supervisor / Docker / Api / MemoryFragment` 的卡片化页面风格

## 2. 结论

不建议直接把该功能塞进现有“自定义网页 / Link / SmartLink”模块。

原因：

- `SmartLink` 偏“通用网页自动化编排”
- 你的需求偏“业务任务管理”
- 新需求天然需要任务、网页、登录状态、执行记录、AI 汇总结果等实体
- 如果继续堆在 `SmartLink` 上，前端交互和后端表结构都会变得难维护

建议做成一个新的独立模块：`InfoCrawl`

但内部尽量复用：

- `plw` 的 Playwright 浏览器与上下文能力
- `set_ai.go` 已有 AI 服务商 / 模型配置
- 当前 SSE 输出机制

## 3. 功能范围

### 3.1 本期包含

- 任务列表、创建、编辑、删除
- 任务下网页列表维护
- 网页登录状态管理
- 手动打开网页进行登录
- 手动执行任务
- 按网页抓取文本内容
- 任务级提示词
- 调用 AI 生成汇总结果
- 执行历史与结果查看

### 3.2 本期不做

- 定时调度
- 自动识别登录表单并自动登录
- 可视化流程编排器
- 复杂字段级结构化抽取 DSL
- 多人协作 / 权限
- 自动重试策略中心

## 4. 核心交互

### 4.1 任务维度

一个任务包含：

- 任务名称
- AI 模型
- 汇总提示词
- 多个网页配置

### 4.2 网页维度

每个网页配置包含：

- 页面名称
- 目标 URL
- 是否启用
- 登录方式：手动登录
- 抓取选择器
- 等待规则
- 抓取说明
- 排序权重

### 4.3 登录流程

建议采用“先人工登录并保存会话，执行时复用会话”的模式。

流程：

1. 用户在网页配置中点击“打开登录页”
2. 系统使用 Playwright 打开该网页的持久化上下文
3. 用户在浏览器里手动完成登录
4. 用户回到系统点击“标记登录完成”
5. 系统校验当前页面已登录并保存该网页的会话目录
6. 之后任务执行时复用该网页登录态

这样可以满足“每个网页需要手动登录”，同时避免每次执行都要重新扫码或输入账号密码。

如果登录过期，则该网页在执行时返回“需要重新登录”。

## 5. 执行流程

### 5.1 单次任务执行

1. 用户点击“执行任务”
2. 后端创建执行记录
3. 逐个网页执行抓取
4. 对每个网页：
   - 载入该网页自己的持久化浏览器会话
   - 打开 URL
   - 等待页面稳定
   - 基于选择器提取正文文本
   - 保存抓取结果
5. 将所有网页抓取结果拼装为 AI 输入
6. 调用用户选择的 AI 模型
7. 生成汇总结论
8. 更新执行状态为成功或失败

### 5.2 失败策略

建议采用“网页级容错，任务级收敛”：

- 单个网页失败，不立即中断整个任务
- 记录该网页失败原因
- 其余网页继续执行
- 最终 AI 汇总时带上成功网页内容和失败网页说明

这样实际可用性更高。

## 6. 数据模型

建议新增 4 张主表。

### 6.1 `tbl_info_crawl_task`

任务主表。

建议字段：

- `id`
- `name`
- `prompt`
- `ai_model_id`
- `status`，`1=正常 0=删除`
- `create_time`
- `update_time`

说明：

- `prompt` 为任务级提示词
- `ai_model_id` 关联 `tbl_ai_model.id`

### 6.2 `tbl_info_crawl_task_page`

任务网页表。

建议字段：

- `id`
- `task_id`
- `name`
- `url`
- `content_selector`
- `wait_selector`
- `wait_mills`
- `sort`
- `note`
- `login_status`，`0=未登录 1=已登录 2=失效`
- `login_check_selector`
- `user_data_dir`
- `status`
- `create_time`
- `update_time`

说明：

- `content_selector`：抓取正文的主选择器，默认空表示抓取 `body.innerText`
- `wait_selector`：页面加载完成判定，可选
- `login_check_selector`：用于校验是否登录成功，比如头像、用户菜单、退出按钮
- `user_data_dir`：保存该网页对应的 Playwright 持久化目录

### 6.3 `tbl_info_crawl_run`

任务执行记录表。

建议字段：

- `id`
- `task_id`
- `status`，`running/success/partial_failed/failed`
- `run_message`
- `prompt_snapshot`
- `ai_model_snapshot`
- `summary_content`
- `page_total`
- `page_success_total`
- `page_failed_total`
- `create_time`
- `update_time`

说明：

- 执行时保存提示词和模型快照，避免后续任务被修改后历史不可追溯

### 6.4 `tbl_info_crawl_run_page`

单网页执行明细表。

建议字段：

- `id`
- `run_id`
- `task_page_id`
- `page_name`
- `url`
- `status`，`success/failed/login_required`
- `error_message`
- `raw_text`
- `raw_html`
- `screenshot_path`
- `create_time`
- `update_time`

说明：

- `raw_text` 作为 AI 汇总主输入
- `raw_html` 可选保留，建议本期允许为空
- `screenshot_path` 用于失败排查

## 7. 后端设计

### 7.1 目录建议

建议新增：

- `internal/app/dtool/controller/info_crawl.go`
- `internal/app/dtool/define/info_crawl.go`
- `internal/app/dtool/struct/info_crawl.go`
- `internal/app/dtool/service/info_crawl_service.go`

如果当前项目不想再拆 `service` 目录，也可以先把核心逻辑放在 `controller + common`，但从复杂度看，建议单独抽服务层。

### 7.2 路由建议

在 [internal/app/dtool/router.go](/c:/work/frog/dev_tool_master/internal/app/dtool/router.go) 中新增一组接口，命名风格保持现有项目一致：

- `POST /api/InfoCrawlTaskList`
- `POST /api/InfoCrawlTaskInfo`
- `POST /api/InfoCrawlTaskSave`
- `POST /api/InfoCrawlTaskDelete`
- `POST /api/InfoCrawlTaskPageSave`
- `POST /api/InfoCrawlTaskPageDelete`
- `POST /api/InfoCrawlTaskPageOpenLogin`
- `POST /api/InfoCrawlTaskPageCheckLogin`
- `POST /api/InfoCrawlTaskRun`
- `POST /api/InfoCrawlRunList`
- `POST /api/InfoCrawlRunInfo`

说明：

- `OpenLogin` 负责打开登录页面
- `CheckLogin` 负责用户手工登录后的状态确认
- `TaskRun` 负责执行任务

### 7.3 Playwright 复用策略

不建议直接复用 `SmartLink` 的流程编排数据结构。

建议复用：

- `plw.PlaywrightClient`
- `ContextPageList`
- 持久化用户目录能力

建议新增一个更薄的抓取执行器，例如：

- `internal/app/dtool/plw/info_crawl_runner.go`

职责：

- 根据网页配置打开持久化上下文
- 等待页面
- 抽取文本
- 截图
- 返回结果

这样不会污染 `SmartLink` 那套流程引擎。

### 7.4 AI 调用复用策略

当前项目已经有：

- AI 服务商配置
- AI 模型配置
- `variable/r_cmd.go` 中的 OpenAI 格式调用逻辑

建议把这部分抽成公共方法，例如：

- `internal/app/dtool/service/ai_chat_service.go`

最少提供一个方法：

- `ChatByModel(modelID int, systemPrompt string, userPrompt string) (string, error)`

内部逻辑：

- 查询 `tbl_ai_model`
- 关联查询 `tbl_ai_provider`
- 按现有 OpenAI 兼容格式发请求

这样后续“信息抓取”和“变量脚本 LLM”都能共用。

## 8. 前端设计

### 8.1 新增模块入口

建议新增左侧菜单：`信息抓取`

对应：

- 路由：`/InfoCrawl`
- 页面组件：`web/src/components/InfoCrawl.vue`

同时在：

- [web/src/router/index.js](/c:/work/frog/dev_tool_master/web/src/router/index.js)
- [web/src/components/Home.vue](/c:/work/frog/dev_tool_master/web/src/components/Home.vue)
- [web/src/utils/module.js](/c:/work/frog/dev_tool_master/web/src/utils/module.js)

增加模块注册。

### 8.2 页面布局

建议沿用当前项目双栏布局。

左侧：

- 任务列表
- 新建任务按钮
- 执行历史入口

右侧：

- 任务基础信息
- AI 模型选择
- 提示词编辑区
- 网页列表配置区
- 执行结果区

### 8.3 网页配置卡片

每个网页卡片建议包含：

- 页面名称
- URL
- 正文选择器
- 等待选择器
- 登录校验选择器
- 等待毫秒数
- 备注
- 登录状态 tag
- 按钮：`打开登录页`、`检查登录状态`、`保存`、`删除`

### 8.4 执行结果区

执行后展示：

- 本次执行状态
- AI 汇总结果
- 每个网页的抓取状态
- 失败原因
- 原始抓取文本预览

## 9. 抓取策略

### 9.1 MVP 抽取规则

本期先采用简单稳定的规则：

- 优先读取 `content_selector` 对应元素的 `innerText`
- 若未配置 `content_selector`，则读取 `document.body.innerText`
- 读取前执行：
  - 等待 `wait_selector`
  - 若未配置则等待固定毫秒数
  - 再额外等待一次 `networkidle` 或短暂静置

### 9.2 为什么先不做复杂抽取 DSL

你的目标是“抓取信息后交给 AI 汇总”，不是“精确抽成结构化字段”。

因此本期先把：

- 页面打开
- 登录复用
- 文本抓取
- AI 汇总

这条链路打通，性价比最高。

后续如果发现某些站点抓取噪音太大，再补：

- 多选择器拼接
- 排除选择器
- iframe 抽取
- 分段抓取

## 10. AI 汇总输入格式

建议构造统一 Prompt：

系统提示词：

- 使用任务配置中的汇总规则

用户提示词：

- 附带任务名称
- 附带执行时间
- 附带网页抓取结果列表

示例结构：

```text
任务名称：竞品情报抓取
执行时间：2026-03-10 22:00:00

请基于以下网页内容进行汇总，输出：
1. 关键信息摘要
2. 每个来源的核心结论
3. 需要重点关注的变化
4. 不确定或疑似噪音的信息

网页1：官网公告
URL: https://example.com/a
内容:
...

网页2：后台报表
URL: https://example.com/b
内容:
...
```

建议增加长度保护：

- 单网页文本超过阈值时截断
- 总输入超过阈值时按网页顺序裁剪

否则很容易超模型上下文。

## 11. 状态定义

### 11.1 任务状态

- `1`：正常
- `0`：删除

### 11.2 网页登录状态

- `0`：未登录
- `1`：已登录
- `2`：登录失效

### 11.3 执行状态

- `running`
- `success`
- `partial_failed`
- `failed`

## 12. 数据库迁移建议

建议新增一个 migration 文件，例如：

- `internal/app/dtool/database/2026/03/20260310.信息抓取.sql`

内容包含：

- 创建 `tbl_info_crawl_task`
- 创建 `tbl_info_crawl_task_page`
- 创建 `tbl_info_crawl_run`
- 创建 `tbl_info_crawl_run_page`
- 必要索引

推荐索引：

- `tbl_info_crawl_task(status, update_time)`
- `tbl_info_crawl_task_page(task_id, status, sort)`
- `tbl_info_crawl_run(task_id, create_time)`
- `tbl_info_crawl_run_page(run_id, task_page_id)`

## 13. 推荐实施顺序

### Phase 1

- 建表
- 后端任务 CRUD
- 前端任务管理页

### Phase 2

- 网页配置 CRUD
- 手动登录流程
- 登录状态校验

### Phase 3

- 任务执行
- 单网页抓取
- 执行历史

### Phase 4

- AI 汇总
- 执行结果展示优化
- 截图和失败排查

## 14. 风险点

### 14.1 登录态失效

很多站点登录态会过期。

处理建议：

- 执行前先检查 `login_check_selector`
- 失败则标记 `login_required`
- 页面上明确提示“请重新登录”

### 14.2 页面噪音过多

如果直接抓 `body.innerText`，可能带很多导航和广告文本。

处理建议：

- 网页配置里支持手填 `content_selector`
- 默认抓全页，站点不理想时再精细配置

### 14.3 模型上下文过长

多个网页一起抓，很容易超过上下文限制。

处理建议：

- 每页单独截断
- 总长度兜底截断
- 后续可扩展为“两阶段摘要”

### 14.4 浏览器资源占用

多个网页登录态会产生多个用户目录。

处理建议：

- 每个网页独立目录
- 提供“清理登录缓存”按钮
- 执行完成后及时关闭 page/context

## 15. 与现有模块关系

### 15.1 与 `Link/SmartLink`

- `SmartLink` 继续承担网页自动化编排
- `InfoCrawl` 负责抓取任务管理
- 底层共用 Playwright

### 15.2 与 `AI Provider / AI Model`

- 直接复用
- 不新建重复配置

### 15.3 与 `MemoryFragment`

本期不耦合。

后续可选增强：

- 把抓取历史摘要一键沉淀到知识片段

## 16. 我建议的最终方案

基于当前仓库，最稳的实现方式是：

1. 新增独立模块 `InfoCrawl`
2. 底层复用现有 Playwright 客户端，但不要复用 `SmartLink` 的流程表
3. 登录采用“人工登录 + 保存上下文 + 失效后重新登录”
4. 抓取先做文本级抽取，不先做复杂字段抽取
5. AI 调用抽成公共服务，避免继续散落在 `variable` 逻辑里

## 16.1 已确认边界后的修订

根据本轮确认，设计边界调整为：

1. 登录后保留登录态，后续执行直接复用 Playwright 会话。
2. 抓取内容不再以固定 selector 配置为主，而是由 AI 根据任务提示词自行决定抓取重点。
3. 汇总结果只在系统中展示，但每次执行都必须保留完整历史，可查看详情。

### 修订一：网页配置改为“AI 驱动”

网页配置建议保留这些字段：

- 页面名称
- URL
- 登录校验选择器
- 页面补充说明
- 排序权重

说明：

- `页面补充说明` 用于告诉 AI 该网页应该重点关注什么，例如“只看公告区”“优先关注价格和发布时间”“进入详情页后抓正文”
- 第一阶段不强依赖人工填写 `content_selector`
- 如后续个别站点不稳定，再补充高级选择器配置

### 修订二：执行链路改为“两段 AI”

建议把 AI 使用拆成两段：

1. 抓取规划阶段
2. 汇总输出阶段

抓取规划阶段：

- 输入任务提示词、网页列表、网页说明、可用动作白名单
- 输出结构化动作计划

汇总输出阶段：

- 输入每个网页的抓取结果
- 输出最终摘要

### 修订三：Playwright 只执行白名单动作

虽然是“AI 自行控制抓取内容”，但不建议让 AI 任意生成脚本。

更稳的方案是：

- AI 只输出结构化 JSON
- 后端只执行白名单动作
- 每一步都记录执行日志

白名单动作建议优先复用当前项目已有 `plw` 能力：

- `goto`
- `wait`
- `click`
- `exist_wait`
- `no_exist_wait`
- `text_content`
- `bool_result`

如果后续确实需要，再扩展更多动作，但第一期不要放开无限制控制。

### 修订四：执行记录需要增加“计划快照”

除原来的执行结果外，建议额外记录：

- 本次 AI 规划内容
- 每个网页的动作摘要
- 每步执行日志

这样后续出现抓取不准时，可以直接回看：

- 是提示词问题
- 是 AI 规划问题
- 还是页面执行问题

### 修订五：执行历史展示要求

历史详情页建议至少展示：

- 执行时间
- 执行状态
- 本次提示词快照
- 本次抓取计划
- AI 汇总结果
- 每个网页的抓取状态
- 每个网页的原始文本结果
- 错误信息与截图

### 修订六：数据库字段补充建议

在现有设计基础上，建议补充这些字段：

`tbl_info_crawl_run`

- `planner_content`：记录本次 AI 抓取计划

`tbl_info_crawl_run_page`

- `planner_action`：记录该网页的抓取动作摘要
- `execute_log`：记录该网页执行日志

## 17. 待确认项

本轮确认后，核心边界已经明确，可以据此继续拆实现方案：

1. 登录后保留登录态。
2. AI 根据提示词控制抓取内容。
3. 结果只在系统展示，但必须记录并可查看每次执行结果。
