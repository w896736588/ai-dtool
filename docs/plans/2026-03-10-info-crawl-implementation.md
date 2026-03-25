# 信息抓取模块实施清单

Date: 2026-03-10

## 1. 实施目标

基于 [2026-03-10-info-crawl-module-design.md](/c:/work/frog/dev_tool_master/docs/plans/2026-03-10-info-crawl-module-design.md)，输出一份可直接开发的实施清单。

本清单聚焦：

- 数据库迁移
- 后端接口与执行链路
- 前端页面拆分
- AI 抓取规划协议
- 执行历史展示

## 2. 实施原则

- 新增独立模块 `InfoCrawl`
- 复用现有 Playwright 底座，不复用 `SmartLink` 的业务表
- 复用现有 AI Provider / AI Model 配置
- 所有后端方法、前端方法增加中文注释
- 不顺手改无关模块

## 3. 建议落地目录

### 3.1 后端

- `internal/app/dtool/controller/info_crawl.go`
- `internal/app/dtool/define/info_crawl.go`
- `internal/app/dtool/struct/info_crawl.go`
- `internal/app/dtool/common/info_crawl.go`
- `internal/app/dtool/plw/info_crawl_runner.go`
- `internal/app/dtool/plw/info_crawl_planner.go`

说明：

- 当前项目大量 CRUD 逻辑放在 `controller + common`
- 为了不引入新的目录风格，优先沿用这个结构
- 如果后续复杂度继续上升，再抽 `service`

### 3.2 前端

- `web/src/components/InfoCrawl.vue`
- `web/src/components/info_crawl/TaskList.vue`
- `web/src/components/info_crawl/TaskEditor.vue`
- `web/src/components/info_crawl/PageEditor.vue`
- `web/src/components/info_crawl/RunHistoryDrawer.vue`
- `web/src/components/info_crawl/RunDetailDialog.vue`
- `web/src/utils/base/info_crawl.js`

## 4. 数据库迁移

建议新增：

- `internal/app/dtool/database/2026/03/20260310.信息抓取.sql`

### 4.1 SQL 草案

```sql
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

CREATE TABLE IF NOT EXISTS "tbl_info_crawl_task_page"
(
    "id"                   INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "task_id"              INTEGER NOT NULL DEFAULT 0,
    "name"                 TEXT    NOT NULL DEFAULT '',
    "url"                  TEXT    NOT NULL DEFAULT '',
    "note"                 TEXT    NOT NULL DEFAULT '',
    "login_check_selector" TEXT    NOT NULL DEFAULT '',
    "login_status"         INTEGER NOT NULL DEFAULT 0,
    "user_data_dir"        TEXT    NOT NULL DEFAULT '',
    "sort"                 INTEGER NOT NULL DEFAULT 0,
    "status"               INTEGER NOT NULL DEFAULT 1,
    "create_time"          INTEGER NOT NULL DEFAULT 0,
    "update_time"          INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "tbl_info_crawl_run"
(
    "id"                 INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "task_id"            INTEGER NOT NULL DEFAULT 0,
    "status"             TEXT    NOT NULL DEFAULT 'running',
    "run_message"        TEXT    NOT NULL DEFAULT '',
    "prompt_snapshot"    TEXT    NOT NULL DEFAULT '',
    "ai_model_snapshot"  TEXT    NOT NULL DEFAULT '',
    "planner_content"    TEXT    NOT NULL DEFAULT '',
    "summary_content"    TEXT    NOT NULL DEFAULT '',
    "page_total"         INTEGER NOT NULL DEFAULT 0,
    "page_success_total" INTEGER NOT NULL DEFAULT 0,
    "page_failed_total"  INTEGER NOT NULL DEFAULT 0,
    "create_time"        INTEGER NOT NULL DEFAULT 0,
    "update_time"        INTEGER NOT NULL DEFAULT 0
);

CREATE TABLE IF NOT EXISTS "tbl_info_crawl_run_page"
(
    "id"              INTEGER NOT NULL PRIMARY KEY AUTOINCREMENT,
    "run_id"          INTEGER NOT NULL DEFAULT 0,
    "task_page_id"    INTEGER NOT NULL DEFAULT 0,
    "page_name"       TEXT    NOT NULL DEFAULT '',
    "url"             TEXT    NOT NULL DEFAULT '',
    "status"          TEXT    NOT NULL DEFAULT '',
    "error_message"   TEXT    NOT NULL DEFAULT '',
    "planner_action"  TEXT    NOT NULL DEFAULT '',
    "execute_log"     TEXT    NOT NULL DEFAULT '',
    "raw_text"        TEXT    NOT NULL DEFAULT '',
    "raw_html"        TEXT    NOT NULL DEFAULT '',
    "screenshot_path" TEXT    NOT NULL DEFAULT '',
    "create_time"     INTEGER NOT NULL DEFAULT 0,
    "update_time"     INTEGER NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS "idx_info_crawl_task_status_update"
    ON "tbl_info_crawl_task" ("status", "update_time");

CREATE INDEX IF NOT EXISTS "idx_info_crawl_task_page_task_status_sort"
    ON "tbl_info_crawl_task_page" ("task_id", "status", "sort");

CREATE INDEX IF NOT EXISTS "idx_info_crawl_run_task_time"
    ON "tbl_info_crawl_run" ("task_id", "create_time");

CREATE INDEX IF NOT EXISTS "idx_info_crawl_run_page_run_task_page"
    ON "tbl_info_crawl_run_page" ("run_id", "task_page_id");
```

### 4.2 状态约定

`tbl_info_crawl_task.status`

- `1` 正常
- `0` 删除

`tbl_info_crawl_task_page.login_status`

- `0` 未登录
- `1` 已登录
- `2` 失效

`tbl_info_crawl_run.status`

- `running`
- `success`
- `partial_failed`
- `failed`

`tbl_info_crawl_run_page.status`

- `success`
- `failed`
- `login_required`

## 5. 后端接口清单

### 5.1 任务接口

#### `POST /api/InfoCrawlTaskList`

用途：

- 查询任务列表

请求：

```json
{}
```

响应：

```json
{
  "task_list": [
    {
      "id": 1,
      "name": "竞品抓取",
      "ai_model_id": 2,
      "update_time": 1741615200
    }
  ]
}
```

#### `POST /api/InfoCrawlTaskInfo`

用途：

- 查询任务详情
- 一次返回任务信息、网页列表、最近执行记录

请求：

```json
{
  "id": 1
}
```

响应：

```json
{
  "task": {},
  "page_list": [],
  "run_list": []
}
```

#### `POST /api/InfoCrawlTaskSave`

用途：

- 新增或更新任务

请求：

```json
{
  "id": 1,
  "name": "竞品抓取",
  "prompt": "请抓取每个页面中与定价、公告、发布时间相关的信息，并给出简明结论",
  "ai_model_id": 2
}
```

#### `POST /api/InfoCrawlTaskDelete`

用途：

- 软删除任务

### 5.2 网页接口

#### `POST /api/InfoCrawlTaskPageSave`

用途：

- 新增或更新任务网页

请求：

```json
{
  "id": 1,
  "task_id": 1,
  "name": "官网公告",
  "url": "https://example.com/news",
  "note": "重点关注最新公告、标题、发布时间和正文摘要",
  "login_check_selector": ".user-avatar",
  "sort": 10
}
```

#### `POST /api/InfoCrawlTaskPageDelete`

用途：

- 软删除网页配置

#### `POST /api/InfoCrawlTaskPageOpenLogin`

用途：

- 用独立持久化上下文打开登录页

请求：

```json
{
  "task_page_id": 1
}
```

执行逻辑：

- 读取网页配置
- 计算 `user_data_dir`
- 启动 Playwright 持久化会话
- 打开该网页 URL

#### `POST /api/InfoCrawlTaskPageCheckLogin`

用途：

- 用户手动登录后，点击按钮进行校验

请求：

```json
{
  "task_page_id": 1
}
```

执行逻辑：

- 如果配置了 `login_check_selector`，则检测元素是否存在
- 成功后把 `login_status` 更新为 `1`
- 失败则返回错误

### 5.3 执行接口

#### `POST /api/InfoCrawlTaskRun`

用途：

- 手动执行任务

请求：

```json
{
  "task_id": 1,
  "sse_distribute_id": "info_crawl_run_1"
}
```

执行逻辑：

1. 查询任务与网页列表
2. 创建 `tbl_info_crawl_run`
3. 生成抓取规划
4. 逐网页执行
5. 写入 `tbl_info_crawl_run_page`
6. 生成汇总
7. 更新执行结果

#### `POST /api/InfoCrawlRunList`

用途：

- 查询任务执行历史

请求：

```json
{
  "task_id": 1,
  "limit": 20
}
```

#### `POST /api/InfoCrawlRunInfo`

用途：

- 查询一次执行的完整详情

请求：

```json
{
  "id": 1001
}
```

响应：

```json
{
  "run_info": {},
  "run_page_list": []
}
```

## 6. 后端方法拆分建议

### 6.1 `controller/info_crawl.go`

建议包含方法：

- `InfoCrawlTaskList`
- `InfoCrawlTaskInfo`
- `InfoCrawlTaskSave`
- `InfoCrawlTaskDelete`
- `InfoCrawlTaskPageSave`
- `InfoCrawlTaskPageDelete`
- `InfoCrawlTaskPageOpenLogin`
- `InfoCrawlTaskPageCheckLogin`
- `InfoCrawlTaskRun`
- `InfoCrawlRunList`
- `InfoCrawlRunInfo`

### 6.2 `common/info_crawl.go`

建议包含数据库方法：

- `InfoCrawlTaskList()`
- `InfoCrawlTaskInfo(id int)`
- `InfoCrawlTaskSave(data map[string]any)`
- `InfoCrawlTaskDelete(id int)`
- `InfoCrawlTaskPageSave(data map[string]any)`
- `InfoCrawlTaskPageDelete(id int)`
- `InfoCrawlRunCreate(data map[string]any)`
- `InfoCrawlRunUpdate(id int, data map[string]any)`
- `InfoCrawlRunPageCreate(data map[string]any)`
- `InfoCrawlRunList(taskID int, limit int)`
- `InfoCrawlRunInfo(id int)`

## 7. AI 抓取规划协议

### 7.1 为什么需要协议

你要求“根据提示词让 AI 自行控制抓取内容”，但如果完全放开，执行会不稳定。

因此需要一个结构化协议：

- AI 负责规划
- 后端负责校验
- Playwright 负责执行

### 7.2 规划输入

输入给 AI 的内容建议包含：

- 任务名称
- 任务提示词
- 网页列表
- 每个网页的：
  - 名称
  - URL
  - 页面说明
- 允许的动作白名单
- 每个动作的 JSON 格式

### 7.3 规划输出

AI 返回 JSON：

```json
{
  "pages": [
    {
      "task_page_id": 1,
      "goal": "抓取最新公告的标题、发布时间和正文摘要",
      "actions": [
        {
          "type": "wait",
          "value": "1500"
        },
        {
          "type": "text_content",
          "locator": "body",
          "out_key": "page_text"
        }
      ]
    }
  ]
}
```

### 7.4 白名单动作

第一期只允许这些动作：

- `wait`
- `click`
- `exist_wait`
- `no_exist_wait`
- `text_content`
- `bool_result`

说明：

- `goto` 不交给 AI 控制，由后端固定执行
- `input` 第一阶段不开放，避免 AI 误输入
- 登录动作完全由人工完成

### 7.5 校验规则

后端在执行前必须校验：

- `task_page_id` 必须存在且属于当前任务
- `actions` 数量不能超过上限，例如 20
- `type` 必须属于白名单
- `locator` 长度不能超限
- `wait` 不能超过最大值，例如 15000ms

如果校验失败：

- 当前网页直接标记 `failed`
- 记录错误到 `planner_action` / `execute_log`

## 8. Playwright 执行链路

### 8.1 `info_crawl_runner.go`

建议实现一个执行器结构：

```go
type InfoCrawlRunner struct {
    RunID int
    TaskID int
    Log *gstool.GsSlog
}
```

### 8.2 单网页执行步骤

1. 根据 `task_page_id` 查询网页配置
2. 校验 `login_status`
3. 打开该网页持久化上下文
4. 固定执行 `page.Goto(url)`
5. 逐步执行 AI 规划动作
6. 收集抽取到的文本
7. 截图
8. 保存明细

### 8.3 文本汇总策略

一个网页可能有多个 `text_content` 动作结果。

建议最终组装为：

```text
[网页名称]
抓取目标：xxx
抓取结果1：
...

抓取结果2：
...
```

### 8.4 超时控制

建议限制：

- 单网页总超时：60 秒
- 单动作超时：15 秒
- 单任务总超时：10 分钟

## 9. AI 汇总阶段

### 9.1 输入

- 任务名称
- 任务提示词快照
- 执行时间
- 各网页抓取结果
- 失败网页及原因

### 9.2 输出要求

建议在系统 Prompt 里固定要求：

- 输出中文
- 分段清晰
- 标注信息来源
- 明确区分“确定信息”和“不确定信息”

### 9.3 内容裁剪

需要做长度保护：

- 单网页文本最大 12000 字符
- 单任务总输入最大 50000 字符

超过后按网页顺序截断，并在 Prompt 里告知“部分长文本已截断”。

## 10. 前端页面拆分

### 10.1 `InfoCrawl.vue`

主页面容器，建议采用与 [MemoryFragment.vue](/c:/work/frog/dev_tool_master/web/src/components/MemoryFragment.vue) 类似的左右结构：

左侧：

- 任务列表
- 新建任务按钮

右侧：

- 任务编辑区域
- 网页配置区域
- 执行历史区域

### 10.2 `TaskList.vue`

职责：

- 展示任务列表
- 选择当前任务
- 新建任务

### 10.3 `TaskEditor.vue`

职责：

- 编辑任务名称
- 选择 AI 模型
- 编辑提示词
- 执行任务
- 查看历史

### 10.4 `PageEditor.vue`

职责：

- 展示网页列表
- 新增网页
- 编辑网页名称、URL、说明、登录校验选择器
- 打开登录页
- 检查登录状态

### 10.5 `RunHistoryDrawer.vue`

职责：

- 展示当前任务最近执行记录
- 点击打开详情

### 10.6 `RunDetailDialog.vue`

职责：

- 展示一次执行详情
- 展示提示词快照
- 展示抓取计划
- 展示 AI 汇总结果
- 展示每个网页的原始抓取结果

## 11. 前端 API 封装

新增：

- `web/src/utils/base/info_crawl.js`

方法建议：

- `InfoCrawlTaskList`
- `InfoCrawlTaskInfo`
- `InfoCrawlTaskSave`
- `InfoCrawlTaskDelete`
- `InfoCrawlTaskPageSave`
- `InfoCrawlTaskPageDelete`
- `InfoCrawlTaskPageOpenLogin`
- `InfoCrawlTaskPageCheckLogin`
- `InfoCrawlTaskRun`
- `InfoCrawlRunList`
- `InfoCrawlRunInfo`

风格对齐参考：

- [memory_fragment.js](/c:/work/frog/dev_tool_master/web/src/utils/base/memory_fragment.js)

## 12. 路由与菜单改动

### 12.1 路由

在 [index.js](/c:/work/frog/dev_tool_master/web/src/router/index.js) 增加：

- `/InfoCrawl`

### 12.2 菜单

在 [Home.vue](/c:/work/frog/dev_tool_master/web/src/components/Home.vue) 左侧菜单增加：

- `信息抓取`

### 12.3 模块开关

在 [module.js](/c:/work/frog/dev_tool_master/web/src/utils/module.js) 中增加：

- `info_crawl`

## 13. 推荐开发顺序

### 第 1 步

- 新增数据库迁移
- 新增后端常量与结构体

### 第 2 步

- 完成任务 CRUD
- 完成网页 CRUD
- 完成前端基础页面

### 第 3 步

- 完成网页登录态打开与校验
- 完成执行历史表读写

### 第 4 步

- 完成 AI 抓取规划
- 完成 Playwright 白名单动作执行器

### 第 5 步

- 完成 AI 汇总
- 完成执行详情弹窗

## 14. 最小可交付版本

如果你要尽快上线，第一版建议只做：

- 任务管理
- 网页管理
- 登录态保留
- 手动执行
- AI 规划后只执行：
  - `wait`
  - `click`
  - `text_content`
- 执行历史查看

先不要做：

- `bool_result`
- 多轮规划
- HTML 保存
- 截图预览优化

## 15. 下一步建议

如果继续往下推进，建议下一份文档直接输出：

1. 详细 SQL 最终版
2. 后端每个接口的请求/响应结构体
3. 前端组件的字段清单和状态流转
4. AI 抓取规划 Prompt 模板

这一步完成后，基本就可以正式开改代码。
