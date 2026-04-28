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

#### Tab 2：开发执行 MD

用途：

- 任务创建时自动生成一个新的知识片段
- 供 AI 或人工写入开发方案、接口补充说明、实施记录
- 作为后续生成覆盖分析与测试计划的重要输入

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

不建议第一期承担：

- 主接口测试执行
- 主覆盖判断
- 主测试结果判定

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

### 5.4 为什么必须保留 Snapshot

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

#### 准备前置数据

规则：

- 优先走业务接口
- 不允许直接写数据库
- 无法造数则记录阻塞项

#### 执行接口测试

逐条调用 `/api/ApiRun`

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
- 登录态复用
- 仅失败用例重跑
- 历史记录对比
- 覆盖趋势分析
- 外部 AI 一键读取失败上下文进行修复

---

## 十七、结论

该方案第一期完全可行，且与当前项目基础能力高度匹配。

核心价值不在于立即做自动修复，而在于先把下面三件事做稳定：

1. 需求是否实现
2. 接口是否符合需求
3. 每次测试是否可留痕、可重跑、可复盘

建议从 `API-only + 工作流程页 + 覆盖检查 + 接口测试历史` 这条主链路启动，后续再逐步扩展到更完整的 AI 测试与修复闭环。
