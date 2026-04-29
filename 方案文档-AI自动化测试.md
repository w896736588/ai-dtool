# AI 自动化测试方案

> 目标：从任务清单中的 `tapd_url` 出发，自动抓取需求文档 MD，结合开发执行文档、当前分支代码变更、接口定义与只读数据库信息，生成可执行测试计划，执行接口测试，并保留每次测试历史记录与覆盖检查结果。

---

## 一、背景与目标

### 1.1 当前基础能力

当前项目已经具备较完整的自动化底座：

- 任务清单支持 `tapd_url`，可自动抓取 TAPD 页面并转为 Markdown
- 已有知识片段能力，可承载需求文档、开发执行文档等 MD 内容
- 已有异步任务体系 `async_task`，适合承接长流程 AI 编排任务
- 已有接口开发模块，支持接口定义、环境管理、接口执行 ` /api/ApiRun`
- 已有分支变更检测脚本 `show-branch-diff`
- 已集成 Smart Link / Playwright / dtool-agent，可作为后续 UI 辅助能力
- 项目已具备 MySQL 配置，可为测试编排 Agent 提供只读数据库查询能力

### 1.2 第一阶段目标

第一阶段聚焦 `API-only` 主链路，不做自动修复，先把“需求是否实现”和“接口是否可用”两件事做扎实。

目标闭环如下：

```text
tapd_url
-> 抓取需求 MD
-> 生成开发执行 MD
-> 生成覆盖检查结果
-> 生成可执行测试计划
-> 执行接口测试
-> 输出测试报告
-> 保留历史记录，供人工或外部 AI 修复后再次回归
```

### 1.3 验证维度

系统需要同时回答 3 个问题：

1. 需求 MD 中描述的功能是否已经在当前实现中落地
2. 已落地的接口是否符合需求预期
3. 当前测试过程与测试结果是否可回溯、可重跑、可复盘

---

## 二、最终方案定位

### 2.1 采用路线

采用 `方案 1：轻编排方案`，但按长期可扩展的边界设计：

- 前端新增任务工作流程页，挂到任务清单
- 后端复用现有 `async_task + /api/ApiRun + 知识片段 + diff 脚本`
- 新增轻量工作流数据表与测试历史表
- 第四个 Tab 只做“接口测试与覆盖检查”，不做自动修复
- 修复动作由人工或外部 AI 单独完成，再回到系统重新执行测试

### 2.2 第一期不做的事情

为了控制复杂度，第一期明确不做：

- 自动修复代码
- 直接写数据库造测试数据
- Playwright 主导的 E2E 自动测试
- 复杂审批流
- 大规模回归平台化能力

### 2.3 数据库工具边界

测试编排 Agent 可以调用数据库工具，但严格限定为只读。

允许：

- 查询表结构
- 查询字段类型、主键、索引
- 查询少量样本数据
- 校验接口执行前后的数据库结果

禁止：

- 直接 `insert / update / delete`
- 绕过业务接口直接补业务数据
- 为了测试方便修改生产语义数据

前置造数原则：

1. 优先调用当前代码中已经存在的业务接口准备数据
2. 如果找不到合适接口，则记录为 `阻塞项`
3. 阻塞项需要在页面中明确展示，而不是悄悄跳过

---

## 三、产品形态：任务工作流程页

### 3.1 入口

在任务清单中为每个任务新增 `工作流程` 按钮，点击进入：

```text
/task-workflow/:taskId
```

### 3.2 页面顶部信息

页面头部建议显示：

- 任务名称
- 任务状态
- TAPD 链接
- 需求文档更新时间
- 开发执行更新时间
- 最近测试时间
- 最近测试结果
- 当前分支
- 基线分支

### 3.3 四个 Tab 的定义

#### Tab 1：需求文档 MD

用途：

- 展示 TAPD 抓取后的需求片段
- 支持 `预览 / 源码` 切换
- 支持查看最近抓取时间

给 AI 的提示词：

```text
读取 xxxxx（TAPD 抓取后生成的知识片段分享地址），分析并设计方案
```

补充说明：

- `xxxxx` 由当前任务关联的需求知识片段分享地址动态填入
- AI 输出建议为结构化 Markdown，至少包含：需求摘要、功能点拆分、涉及接口、风险点、待确认问题
- 该输出默认写入 Tab 2 作为开发执行 MD 的初稿来源

#### Tab 2：开发执行 MD

用途：

- 任务创建时自动生成一个新的知识片段
- 供 AI 或人工写入开发方案、接口补充说明、实施记录
- 作为后续生成覆盖分析与测试计划的重要输入

给 AI 的提示词：

```text
将开发方案通过 /api/task/workflow/dev-plan/save 接口更新进去。请求方式：POST。入参示例：{"workflow_id":10,"content":"# 开发执行说明\n..."}。其中 workflow_id 为当前工作流 id，content 为 AI 生成或补充后的完整 Markdown 内容。
```

补充说明：

- `workflow_id` 来自 `/api/task/workflow/create_or_get` 或 `/api/task/workflow/info`
- `content` 建议传完整文档内容，而不是仅传增量 patch，避免版本合并复杂化
- 后端保存后应返回最新更新时间、版本号或摘要信息，供页面刷新展示

建议默认模板：

```md
# 开发执行说明

## 需求摘要

## 开发方案

## 涉及接口

## 数据影响

## 风险与限制

## 实施记录
```

#### Tab 3：测试接口计划

用途：

- 展示 AI 生成的可执行测试计划
- 底层主数据为 `test_plan.json`
- 页面负责把结构化计划渲染为可读摘要

推荐按钮：

- `生成覆盖分析`
- `生成测试计划`
- `查看计划 JSON`

推荐展示内容：

- 覆盖需求点列表
- 关联接口列表
- 前置条件
- 测试用例列表
- 疑问项
- 阻塞项

#### Tab 4：接口测试与覆盖检查

用途：

- 执行覆盖检查
- 执行接口测试
- 查看测试历史
- 支持按历史记录重跑
- 明确列出需求未实现项、疑问项、阻塞项

推荐按钮：

- `执行覆盖检查`
- `执行接口测试`
- `执行覆盖检查+接口测试`
- `按历史记录重跑`
- `查看历史记录`

推荐展示内容：

- 当前执行阶段与进度
- 实时日志
- 覆盖检查结果
- 本次测试结果
- 历史执行记录

---

## 四、整体架构

### 4.1 核心思路

第一阶段采用“当前项目内置测试编排能力 + 外部大模型推理”的组合方式：

- `dtool` 负责：任务、页面、状态、异步任务、接口执行、知识片段、历史记录
- `测试编排 Agent` 负责：解析上下文、生成覆盖分析、生成测试计划、归纳失败结果
- `大模型` 负责：推理与结构化输出

### 4.2 推荐流程

```text
任务清单
-> 工作流程页
-> 需求文档 MD
-> 开发执行 MD
-> 覆盖检查
-> 测试计划生成
-> 接口测试执行
-> 数据库只读校验
-> 测试报告
-> 历史记录
```

### 4.3 Playwright 的定位

第一期中，Playwright 不作为主测试执行器，只作为辅助信息采集器。

适用场景：

- 自动登录系统
- 辅助识别页面功能对应的接口
- 抓取页面触发的接口请求样例
- 帮助 AI 理解某些功能入口
- 当需求主要以页面操作描述、接口映射不清时，补充真实请求证据

不建议第一期承担：

- 主接口测试执行
- 主覆盖判断
- 主测试结果判定

### 4.4 Playwright 与现有能力的结合方式

核心原则：

- `Playwright` 负责“页面辅助识别”
- `/api/ApiRun` 负责“正式接口测试执行”
- 数据库工具负责“只读校验”

推荐链路：

```text
需求 MD / 开发执行 MD
-> AI 初步识别需求点
-> 如接口不明确，则触发 Playwright 页面辅助识别
-> 采集页面真实请求样例 ui_api_trace.json
-> AI 结合代码、接口定义、数据库 schema 生成 test_plan.json
-> 执行器逐条调用 /api/ApiRun
-> 数据库只读校验
-> 输出测试报告
```

Playwright 产物建议单独沉淀为：

- `ui_api_trace.json`
- 页面动作日志
- 关键请求响应样例

`ui_api_trace.json` 建议结构：

```json
{
  "workflow_id": 10,
  "env_id": 3,
  "entry_url": "/order/list",
  "action_template": "create_order",
  "trace_summary": {
    "action_count": 3,
    "request_count": 6,
    "matched_api_count": 2
  },
  "actions": [
    {
      "step": 1,
      "action": "click",
      "target": "[data-testid='create-btn']",
      "label": "点击创建按钮"
    }
  ],
  "requests": [
    {
      "request_id": "req-1",
      "step": 3,
      "type": "xhr",
      "method": "POST",
      "url": "/api/order/create",
      "query": {},
      "request_headers": {
        "content-type": "application/json"
      },
      "request_body": {
        "product_id": 101,
        "count": 2
      },
      "response_status": 200,
      "response_body_sample": {
        "code": 0,
        "data": {
          "id": 123
        }
      },
      "matched_api_id": 1001,
      "role": "target_api"
    }
  ],
  "open_questions": [],
  "blocked_items": []
}
```

字段设计重点：

- `actions` 用于回看页面操作路径
- `requests` 用于沉淀真实请求证据
- `matched_api_id` 用于与接口定义关联
- `role` 用于标记该请求是 `prepare_api`、`target_api` 还是 `noise`

这样可以保证：

- 浏览器采集与接口执行解耦
- 测试计划生成有真实页面证据，而不只依赖代码猜测
- 后续即使不再重复跑页面采集，也可复用已有 trace 生成或修正测试计划

### 4.5 页面动作模板建议

第一期建议仅支持“预设动作模板”，不做开放式浏览器录制。

推荐模板字段：

```json
{
  "template_code": "create_order",
  "name": "创建订单",
  "entry_url": "/order/list",
  "steps": [
    {
      "type": "click",
      "selector": "[data-testid='create-btn']",
      "label": "打开创建弹窗"
    },
    {
      "type": "fill",
      "selector": "[name='productName']",
      "value_from": "context.product_name",
      "label": "填写商品名"
    },
    {
      "type": "click",
      "selector": "[data-testid='submit-btn']",
      "label": "提交表单"
    }
  ]
}
```

建议支持的步骤类型：

- `goto`
- `click`
- `fill`
- `select`
- `check`
- `wait_response`
- `wait_visible`

模板的主要价值：

- 降低第一期的交互复杂度
- 避免自由录制导致脚本不稳定
- 便于沉淀业务功能与接口映射关系

---

## 五、后端数据设计

### 5.1 工作流主表 `tbl_task_workflow`

一条任务对应一条工作流主记录。

建议字段：

- `id`
- `home_task_id`
- `status`
- `current_stage`
- `requirement_fragment_id`
- `dev_plan_fragment_id`
- `latest_plan_run_id`
- `latest_test_run_id`
- `base_branch`
- `feature_branch`
- `last_error`
- `create_time`
- `update_time`

### 5.2 测试运行表 `tbl_task_test_run`

一条记录表示一次覆盖分析或一次完整测试执行，必须保留历史快照。

建议字段：

- `id`
- `workflow_id`
- `run_no`
- `run_type`
- `status`
- `trigger_source`
- `requirement_snapshot_md`
- `dev_plan_snapshot_md`
- `diff_snapshot_text`
- `coverage_report_json`
- `test_plan_json`
- `test_report_json`
- `summary_md`
- `started_at`
- `finished_at`
- `create_time`

推荐 `run_type`：

- `coverage_only`
- `plan_generate`
- `test_execute`
- `plan_and_test`

### 5.3 测试用例结果表 `tbl_task_test_case_result`

如果需要在页面细粒度展示每条用例结果，建议新增此表。

建议字段：

- `id`
- `test_run_id`
- `case_id`
- `case_name`
- `requirement_id`
- `api_id`
- `api_uri`
- `status`
- `duration_ms`
- `request_snapshot_json`
- `response_snapshot_json`
- `assertions_json`
- `db_checks_json`
- `failure_reason`
- `create_time`

### 5.4 页面动作模板表 `tbl_task_ui_action_template`

如果第一期要稳定落地 `Playwright` 辅助识别，建议增加页面动作模板表，避免把模板硬编码在程序里。

建议字段：

- `id`
- `template_code`
- `template_name`
- `biz_key`
- `entry_url`
- `env_scope`
- `steps_json`
- `status`
- `remark`
- `create_time`
- `update_time`

字段说明：

- `template_code`：如 `create_order`
- `biz_key`：用于按业务域区分，如 `order`、`product`
- `env_scope`：标记模板适用环境，避免测试环境与预发环境页面结构不一致
- `steps_json`：保存预设步骤定义
- `status`：建议取值 `enabled`、`disabled`

### 5.5 为什么必须保留 Snapshot

需求文档、开发执行文档和代码分支后续都可能变化，因此每次执行必须保留当时的上下文快照。

这样才能保证：

- 历史记录可回放
- 失败结果可复盘
- 修复前后可对比

---

## 六、状态流转设计

### 6.1 工作流宏观状态 `status`

建议值：

- `init`
- `dev_plan_ready`
- `coverage_ready`
- `test_plan_ready`
- `testing`
- `await_review`
- `failed`

### 6.2 当前阶段 `current_stage`

建议值：

- `idle`
- `loading_context`
- `checking_coverage`
- `generating_plan`
- `preparing_data`
- `running_cases`
- `checking_db`
- `writing_report`

### 6.3 流转建议

1. 创建任务后：`init`
2. 自动创建开发执行 MD 后：`dev_plan_ready`
3. 覆盖分析成功后：`coverage_ready`
4. 测试计划生成成功后：`test_plan_ready`
5. 测试执行中：`testing`
6. 测试完成待人工查看：`await_review`
7. 执行异常：`failed`

---

## 七、覆盖检查设计

### 7.1 目标

第四个 Tab 不仅要回答“接口能不能跑”，还要回答：

```text
需求 MD 中写的功能，现在到底有没有在接口中实现出来？
```

### 7.2 输出结构

建议单独产出 `coverage_report.json`：

```json
{
  "summary": {
    "requirement_points": 6,
    "covered": 4,
    "partial": 1,
    "missing": 1,
    "questions": 2,
    "blocked": 1
  },
  "items": [
    {
      "requirement_id": "req-1",
      "title": "创建订单",
      "status": "covered",
      "evidence": [
        {"type": "api", "value": "/api/order/create"},
        {"type": "code", "value": "internal/app/order/controller.go"}
      ]
    },
    {
      "requirement_id": "req-2",
      "title": "撤销订单",
      "status": "missing",
      "evidence": [],
      "question": "需求描述中存在撤销能力，但当前 diff 与接口定义中未发现对应接口"
    }
  ],
  "questions": [
    "需求中提到批量操作，但当前仅发现单条操作接口，是否遗漏批量接口？"
  ],
  "blocked": [
    "需要构造某类业务数据，但当前未发现可用于造数的现有接口"
  ]
}
```

### 7.3 覆盖判断证据来源

按优先级建议如下：

1. 开发执行 MD
2. 当前分支相对基线分支的 diff
3. 接口定义
4. 路由与控制器代码
5. 数据库 schema 辅助判断

规则：

- 每个结论必须带证据
- 无法确认的内容进入 `questions`
- 缺少前置能力的内容进入 `blocked`

---

## 八、测试计划设计

### 8.1 核心产物

第三个 Tab 的主产物是机器可执行的 `test_plan.json`，而不是纯 Markdown。

推荐生成输入来源：

1. 需求文档 MD
2. 开发执行 MD
3. 覆盖分析结果 `coverage_report.json`
4. 当前分支相对基线分支的 diff
5. 接口定义详情
6. 路由 / 控制器代码摘要
7. 数据库 schema 摘要与少量样本数据
8. 可选的 `ui_api_trace.json`

生成原则：

- 默认优先依据代码、接口定义和数据库结构直接生成接口测试方案
- 当需求描述偏页面行为、接口不明确时，再引入 `Playwright` 采集结果辅助识别
- `Playwright` 只提供真实请求样例和页面触发顺序，不直接承担最终测试执行

### 8.2 结构示例

```json
{
  "plan_name": "任务123-接口测试计划",
  "workflow_id": 123,
  "source": {
    "task_id": 123,
    "requirement_fragment_id": "req_frag_xxx",
    "dev_plan_fragment_id": "dev_frag_xxx",
    "base_branch": "main",
    "feature_branch": "feature/order"
  },
  "coverage_links": [
    {
      "requirement_id": "req-1",
      "title": "创建订单",
      "apis": ["/api/order/create"]
    }
  ],
  "preconditions": [
    {
      "id": "pre-1",
      "type": "api_prepare",
      "purpose": "创建可用商品",
      "api_uri": "/api/product/create"
    }
  ],
  "api_cases": [
    {
      "case_id": "case-001",
      "name": "创建订单-正常流程",
      "requirement_id": "req-1",
      "api_id": 1001,
      "api_uri": "/api/order/create",
      "method": "POST",
      "request_data": {
        "product_id": "{{pre-1.data.id}}",
        "count": 2
      },
      "assertions": [
        {"type": "status_code", "expected": 200},
        {"type": "json_path", "path": "code", "expected": 0},
        {"type": "json_not_null", "path": "data.id"}
      ],
      "db_checks": [
        {
          "type": "table_exists",
          "table": "orders",
          "condition": "id={{response.data.id}}"
        }
      ]
    }
  ],
  "open_questions": [
    "撤销订单能力未发现对应接口"
  ],
  "blocked_items": [
    "缺少用于创建测试客户的现有接口"
  ]
}
```

### 8.3 设计约束

- 每条用例必须绑定 `requirement_id`
- 前置造数优先调用已有业务接口
- 数据库仅做只读校验
- 阻塞项必须显式输出
- 疑问项必须显式输出

### 8.3.1 `test_plan.json` Schema 建议

建议对 `test_plan.json` 做固定 Schema 校验，至少约束以下字段：

```json
{
  "type": "object",
  "required": [
    "plan_name",
    "workflow_id",
    "source",
    "coverage_links",
    "preconditions",
    "api_cases",
    "open_questions",
    "blocked_items"
  ],
  "properties": {
    "plan_name": {"type": "string", "minLength": 1},
    "workflow_id": {"type": "integer"},
    "source": {
      "type": "object",
      "required": ["task_id", "base_branch", "feature_branch"],
      "properties": {
        "task_id": {"type": "integer"},
        "requirement_fragment_id": {"type": "string"},
        "dev_plan_fragment_id": {"type": "string"},
        "base_branch": {"type": "string"},
        "feature_branch": {"type": "string"},
        "ui_trace_id": {"type": "integer"}
      }
    },
    "coverage_links": {
      "type": "array"
    },
    "preconditions": {
      "type": "array"
    },
    "api_cases": {
      "type": "array",
      "minItems": 1
    },
    "open_questions": {
      "type": "array"
    },
    "blocked_items": {
      "type": "array"
    }
  }
}
```

建议额外增加业务校验：

- `api_cases[*].case_id` 不能重复
- `api_cases[*].requirement_id` 必须在 `coverage_links` 中可找到
- `preconditions[*].id` 不能重复
- `request_data` 中引用的变量必须能在前置步骤或环境变量中解析
- `blocked_items` 非空时，不应把对应阻塞用例标记为可直接执行

### 8.3.2 `api_cases` 字段约束建议

每条 `api_case` 建议至少包含：

- `case_id`
- `name`
- `requirement_id`
- `api_id` 或 `api_uri`
- `method`
- `request_data`
- `assertions`

建议可选字段：

- `tags`
- `priority`
- `timeout_ms`
- `depends_on`
- `skip_reason`

优先级建议：

- `P0`：主链路必过
- `P1`：重要分支
- `P2`：边界或增强验证

### 8.4 页面辅助识别产物如何转成测试计划

当存在 `ui_api_trace.json` 时，AI 生成测试计划需要额外完成以下映射：

1. 将页面动作映射为需求点
2. 将真实请求 URL 映射为接口定义中的 `api_id`
3. 从请求体中抽取参数样例，整理为 `request_data`
4. 从响应体中抽取稳定断言字段
5. 识别哪些请求属于前置造数，哪些请求属于主验证接口

例如：

- 页面点击“创建订单”抓到 `/api/order/create`
- 请求体包含 `product_id`、`count`
- 响应体包含 `code` 和 `data.id`

则 AI 应把它转成：

- 一个 `api_prepare` 类型前置条件，或
- 一个正式 `api_case`

而不是直接保存为浏览器脚本步骤

转换规则建议：

1. 命中 `/api/` 且在接口定义中可识别的请求，优先作为候选业务接口
2. 列表查询、字典查询、埋点、心跳类请求标记为 `noise`
3. 页面打开即自动触发、但与当前需求无强关联的请求不进入正式用例
4. 提交动作后触发且返回关键业务 id 的请求，优先识别为 `target_api`
5. 在主请求前用于创建前置资源的请求，识别为 `prepare_api`

这样生成的 `test_plan.json` 更接近“标准接口测试计划”，而不是“页面录制回放结果”

### 8.5 何时需要 Playwright

建议仅在以下场景触发：

- 需求主要描述页面行为，未明确后端接口
- 接口定义缺失或字段说明不足
- 需要从页面真实操作中抓取请求参数样例
- 需要辅助识别同一功能涉及的多个串联接口

以下场景不建议触发：

- 接口定义已经完整
- 开发执行 MD 已明确写出涉及接口
- 只做常规接口回归
- 仅需基于已有用例重跑

---

## 九、测试报告设计

### 9.1 目标

测试报告既要给前端渲染，也要给后续 AI 或人工复盘使用。

### 9.2 结构示例

```json
{
  "summary": {
    "total": 12,
    "passed": 9,
    "failed": 2,
    "skipped": 1,
    "duration_ms": 18230
  },
  "cases": [
    {
      "case_id": "case-001",
      "name": "创建订单-正常流程",
      "status": "passed",
      "duration_ms": 320,
      "request_snapshot": {},
      "response_snapshot": {},
      "assertions": [
        {
          "type": "status_code",
          "expected": 200,
          "actual": 200,
          "passed": true
        }
      ],
      "db_checks": [
        {
          "table": "orders",
          "passed": true,
          "actual_count": 1
        }
      ]
    }
  ],
  "failures": [
    {
      "case_id": "case-003",
      "reason": "返回字段 code 与预期不一致",
      "suspected_area": "/api/order/cancel"
    }
  ],
  "questions": [],
  "blocked": []
}
```

### 9.3 报告与计划的关联要求

建议 `test_report.json` 中每条结果都保留以下关联字段：

- `case_id`
- `requirement_id`
- `api_id`
- `api_uri`
- `source_plan_version`

这样前端可以从测试报告直接回跳：

- 对应需求点
- 对应接口
- 对应计划版本
- 对应历史执行记录

---

## 十、核心接口设计

建议统一走 `task/workflow` 前缀。

### 10.1 基础信息

- `/api/task/workflow/create_or_get`
- `/api/task/workflow/info`

### 10.2 开发执行 MD

- `/api/task/workflow/dev-plan/init`
- `/api/task/workflow/dev-plan/info`
- `/api/task/workflow/dev-plan/save`

### 10.3 覆盖分析与测试计划

- `/api/task/workflow/coverage/generate`
- `/api/task/workflow/coverage/info`
- `/api/task/workflow/test-plan/generate`
- `/api/task/workflow/test-plan/info`

### 10.4 测试执行

- `/api/task/workflow/test-run/start`
- `/api/task/workflow/test-run/info`
- `/api/task/workflow/test-run/list`
- `/api/task/workflow/test-run/cases`
- `/api/task/workflow/test-run/retry`

### 10.5 只读数据库工具

- `/api/task/workflow/db/schema`
- `/api/task/workflow/db/sample`
- `/api/task/workflow/db/check`

### 10.6 页面辅助识别相关接口补充约定

建议新增：

- `/api/task/workflow/ui-template/list`
- `/api/task/workflow/ui-template/detail`

用途：

- 给前端提供可选动作模板列表
- 允许页面按业务域筛选模板

推荐返回字段：

- `template_code`
- `template_name`
- `biz_key`
- `entry_url`
- `status`

### 10.7 统一错误码建议

建议工作流相关接口统一返回业务错误码，便于前端做精细提示。

可先约定以下错误：

- `TASK_WORKFLOW_NOT_FOUND`
- `TASK_WORKFLOW_DEV_PLAN_EMPTY`
- `TASK_WORKFLOW_REQUIREMENT_EMPTY`
- `TASK_WORKFLOW_UI_TEMPLATE_NOT_FOUND`
- `TASK_WORKFLOW_UI_TRACE_FAILED`
- `TASK_WORKFLOW_TEST_PLAN_INVALID`
- `TASK_WORKFLOW_API_MAPPING_FAILED`
- `TASK_WORKFLOW_DB_CHECK_FAILED`

错误返回建议结构：

```json
{
  "code": "TASK_WORKFLOW_UI_TEMPLATE_NOT_FOUND",
  "message": "未找到可用的页面动作模板",
  "detail": {
    "template_code": "create_order"
  }
}
```

---

## 十一、前端交互补充建议

### 11.1 Tab 按钮可用条件

建议按状态控制按钮可用性：

- `Tab 1`
  - `复制 AI 提示词`：有需求知识片段分享地址即可用
- `Tab 2`
  - `保存`：文档有改动即可用
  - `复制 AI 提示词`：有 `workflow_id` 即可用
- `Tab 3`
  - `页面辅助识别`：已选择环境且存在可用动作模板时可用
  - `生成覆盖分析`：需求 MD 存在时可用
  - `生成测试计划`：开发执行 MD 不为空时可用
  - `基于页面痕迹生成测试计划`：存在成功的 `ui_trace` 时可用
- `Tab 4`
  - `执行接口测试`：存在成功的 `test_plan` 时可用
  - `执行覆盖检查+接口测试`：需求 MD 与开发执行 MD 都存在时可用
  - `按历史记录重跑`：存在历史成功或失败记录时可用

### 11.2 页面辅助识别交互流程

建议流程：

1. 用户点击 `页面辅助识别`
2. 弹出侧边栏或弹窗
3. 选择环境
4. 选择动作模板
5. 填写模板上下文参数
6. 提交后创建异步任务
7. 在页面展示实时日志与结果摘要
8. 成功后允许一键生成测试计划

弹窗建议字段：

- `env_id`
- `template_code`
- `entry_url`
- `template_context`

### 11.3 测试计划区展示建议

建议默认展示“摘要视图”，按需展开 JSON：

- 需求点数
- 接口数
- 前置条件数
- 用例数
- 阻塞项数
- 疑问项数

每条用例建议支持展开查看：

- 请求参数
- 断言
- 数据库校验
- 来源依据

### 11.4 测试执行区展示建议

当前执行区建议固定展示：

- 执行编号
- 执行类型
- 当前阶段
- 进度百分比
- 当前用例名
- 成功 / 失败 / 跳过计数

日志区建议支持：

- 按阶段筛选
- 仅看失败
- 复制日志
- 自动滚动开关

### 11.5 历史记录区交互建议

每条历史记录建议支持：

- 查看执行摘要
- 查看覆盖分析
- 查看测试计划快照
- 查看失败用例详情
- 重跑本次计划
- 基于本次上下文重新生成计划

### 10.6 推荐实现方式

以下动作建议都通过异步任务执行：

- 生成覆盖分析
- 生成测试计划
- 执行接口测试

页面实时状态可继续复用现有 SSE / async_task 广播模式。

---

## 十一、异步任务阶段设计

### 11.1 推荐新增任务类型

- `task_workflow_coverage_generate`
- `task_workflow_test_plan_generate`
- `task_workflow_test_execute`

### 11.2 `test_execute` 阶段建议

1. `加载上下文`
2. `执行覆盖检查`
3. `生成测试计划`
4. `准备前置数据`
5. `执行接口测试`
6. `执行数据库校验`
7. `汇总测试结果`
8. `写入执行记录`

### 11.3 阶段说明

#### 加载上下文

读取：

- 需求 MD
- 开发执行 MD
- branch diff
- 相关接口定义
- 测试环境信息

#### 执行覆盖检查

输出：

- 已覆盖功能点
- 未覆盖功能点
- 疑问项
- 阻塞项

#### 生成测试计划

输出：

- `test_plan.json`
- 测试计划摘要 Markdown
- 疑问项和阻塞项清单

建议生成步骤：

1. 解析覆盖分析结果，抽取已覆盖与部分覆盖的需求点
2. 匹配每个需求点对应的接口定义
3. 判断是否需要前置造数接口
4. 若存在 `ui_api_trace.json`，则用其补充真实请求参数样例
5. 为每个接口生成断言与数据库只读校验项
6. 生成 `open_questions` 与 `blocked_items`

#### 准备前置数据

规则：

- 优先走业务接口
- 不允许直接写数据库
- 无法造数则记录阻塞项

#### 执行接口测试

逐条调用 `/api/ApiRun`

执行细则建议：

- 每条 `api_case` 先解析变量占位符，例如 `{{pre-1.data.id}}`
- 组装环境、请求方法、路径、请求体后调用 `/api/ApiRun`
- 保存请求快照与响应快照
- 单条失败继续后续用例，不中断整个批次

`/api/ApiRun` 建议透传的关键上下文：

- `env_id`
- `api_id` 或 `api_uri`
- `method`
- `request_data`
- 变量解析结果
- 调用来源 `task_workflow_test_execute`

这样便于后续：

- 统一审计调用来源
- 与既有接口执行模块复用日志与回放能力
- 在失败时快速回查原始请求参数

#### 执行数据库校验

只做只读检查：

- 数据是否写入
- 状态是否变化
- 关联记录是否存在

#### 汇总测试结果

输出：

- `test_report.json`
- `summary_md`

#### 写入执行记录

把结构化结果、日志、快照统一落表，并更新工作流状态。

---

## 十二、AI 编排 Agent 设计

### 12.1 Agent 负责什么

- 解析需求 MD
- 解析开发执行 MD
- 结合 diff 判断实现范围
- 从接口定义中寻找候选接口
- 必要时结合 `Playwright` 页面采集结果识别真实请求链路
- 必要时调用只读数据库工具辅助理解数据结构
- 生成覆盖分析
- 生成测试计划
- 在测试完成后生成失败总结

### 12.2 Agent 不负责什么

- 不直接修改数据库
- 不直接修复代码
- 不替代后端做任务调度
- 不持有长期状态

### 12.3 设计原则

- `dtool` 做状态机和执行器
- `测试编排 Agent` 做分析器和生成器
- `大模型` 只负责推理输出

---

## 十三、日志与历史记录

### 13.1 日志格式建议

建议每次运行都记录阶段化日志：

- 阶段
- 动作
- 结果
- 补充信息

示例：

- `加载上下文 | 读取需求文档 | 成功 | fragment_id=req_xxx`
- `覆盖检查 | 匹配接口 | 成功 | 命中 5 个接口`
- `测试计划 | 生成用例 | 成功 | 共 12 条`
- `接口测试 | 执行 case-003 | 失败 | code 断言不匹配`
- `数据库校验 | 校验 orders 记录 | 成功 | 命中 1 条`

### 13.2 历史记录要求

每次执行都必须生成新的 `test_run` 记录，不能覆盖历史。

建议支持：

- 完整重跑
- 基于最近计划重跑
- 基于某次历史记录重跑

---

## 十四、风险与控制策略

| 风险 | 影响 | 控制策略 |
|---|---|---|
| 需求 MD 描述不够清晰 | 覆盖分析和测试计划不准确 | 输出疑问项并要求人工确认 |
| 缺少可用于造数的业务接口 | 测试无法落地 | 标记阻塞项，禁止直接写库绕过 |
| AI 输出结构不稳定 | 前端渲染或执行失败 | 所有 JSON 产物先做 schema 校验再入库 |
| 分支 diff 不完整 | 覆盖判断偏差 | 同时结合接口定义与控制器证据 |
| 测试环境数据状态不稳定 | 用例结果波动 | 优先使用独立测试环境，并记录前置接口造数路径 |

---

## 十五、第一期最小闭环

推荐先做如下能力：

1. 任务清单新增工作流程入口
2. 自动初始化开发执行 MD
3. 生成覆盖分析
4. 生成可执行测试计划
5. 执行 `/api/ApiRun`
6. 执行数据库只读校验
7. 保存测试报告与历史记录
8. 支持查看执行详情与重跑

这套最小闭环已经能够解决：

- 需求是否实现
- 实现的接口是否可用
- 修复后能否快速回归验证

---

## 十六、后续演进方向

在第一期稳定后，可继续扩展：

- Playwright 辅助页面功能识别
- 页面操作模板沉淀与复用
- `ui_api_trace.json` 与测试计划的自动映射优化
- 登录态复用
- 仅失败用例重跑
- 历史记录对比
- 覆盖趋势分析
- 外部 AI 一键读取失败上下文进行修复

---

## 十七、实施里程碑建议

### 17.1 里程碑 M1：基础工作流闭环

目标：

- 页面能进入工作流程页
- 自动初始化开发执行 MD
- 可生成覆盖分析
- 可生成测试计划

建议范围：

- `tbl_task_workflow`
- `tbl_task_test_run`
- `/api/task/workflow/create_or_get`
- `/api/task/workflow/info`
- `/api/task/workflow/dev-plan/*`
- `/api/task/workflow/coverage/*`
- `/api/task/workflow/test-plan/*`

完成标志：

- 用户可以从任务清单进入工作流程页
- Tab 1 和 Tab 2 可完整查看与保存
- Tab 3 可看到覆盖分析和测试计划结果

### 17.2 里程碑 M2：接口测试执行闭环

目标：

- 测试计划可正式执行
- 执行结果可落历史
- 失败详情可回看

建议范围：

- `/api/task/workflow/test-run/*`
- `/api/task/workflow/db/*`
- `/api/ApiRun` 对接
- `tbl_task_test_case_result`

完成标志：

- 可执行 `plan_and_test`
- 可查看单次执行详情
- 可查看每条用例请求、响应、断言、数据库校验结果

### 17.3 里程碑 M3：Playwright 辅助识别闭环

目标：

- 页面动作模板可管理
- 可采集页面真实请求样例
- 可基于页面痕迹生成测试计划

建议范围：

- `tbl_task_ui_action_template`
- `/api/task/workflow/ui-template/*`
- `/api/task/workflow/ui-trace/*`
- `task_workflow_ui_trace_generate`

完成标志：

- 用户可选择模板触发页面辅助识别
- 成功生成 `ui_api_trace.json`
- 可基于 `ui_trace` 辅助生成测试计划

### 17.4 推荐排期策略

建议按 `M1 -> M2 -> M3` 顺序推进。

原因：

- 先做 `API-only` 主链路最稳
- 没有 `M1` 和 `M2`，`Playwright` 采集结果没有稳定落点
- `M3` 是增强项，应建立在已有测试执行闭环之上

---

## 十八、角色分工建议

### 18.1 前端

负责：

- 工作流程页路由与四个 Tab
- 文档展示、提示词复制、计划摘要展示
- 当前执行区、日志区、历史记录区
- 页面辅助识别弹窗与模板选择
- SSE 或轮询刷新

交付物：

- `/task-workflow/:taskId` 页面
- 覆盖分析、测试计划、测试报告的可视化展示
- 历史记录与重跑交互

### 18.2 后端

负责：

- 工作流主表与运行表
- 文档初始化和保存接口
- 覆盖分析、测试计划、测试执行异步任务
- `/api/ApiRun` 与数据库只读校验对接
- 历史快照、日志、详情查询

交付物：

- 任务工作流接口
- 异步执行器
- 测试结果落库与查询接口

### 18.3 AI 编排

负责：

- 覆盖分析 Prompt
- 测试计划 Prompt
- 页面辅助识别 Prompt
- 失败总结 Prompt
- JSON Schema 校验与失败兜底

交付物：

- 结构稳定的 `coverage_report.json`
- 结构稳定的 `test_plan.json`
- 结构稳定的失败摘要

### 18.4 Playwright 能力

负责：

- 登录和页面访问
- 执行动作模板
- 抓取真实请求样例
- 输出 `ui_api_trace.json`

交付物：

- 模板执行器
- 请求过滤器
- trace 结果结构化产物

### 18.5 联调与验收

负责：

- 验证接口入参与页面展示一致
- 验证异步任务状态、日志、进度推送
- 验证 `test_plan.json` 和 `test_report.json` 可正确渲染
- 验证历史重跑和失败回放链路

---

## 十九、结论

该方案第一期完全可行，且与当前项目基础能力高度匹配。

核心价值不在于立即做自动修复，而在于先把下面三件事做稳定：

1. 需求是否实现
2. 接口是否符合需求
3. 每次测试是否可留痕、可重跑、可复盘

建议从 `API-only + 工作流程页 + 覆盖检查 + 接口测试历史` 这条主链路启动，后续再逐步扩展到更完整的 AI 测试与修复闭环。
