# Structured Locator Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 将 Playwright 元素定位链路升级为结构化 `LocatorSpec` 封装，解耦查询解析、Locator 构建和动作执行，同时兼容现有旧字符串配置。

**Architecture:** 以 `LocatorInput -> LocatorSpec -> LocatorResolver -> ElementActionExecutor -> LocatorService` 为主链路重构现有 Locator 内核。第一阶段通过兼容转换保留旧字符串配置入口，逐步将 `Process` 从 `ElementOp` 共享状态模式迁移到显式动作执行模式。

**Tech Stack:** Go、playwright-go、项目现有 `plw` 流程编排、项目现有日志与工具库

---

### Task 1: 补齐现状摸底与影响点清单

**Files:**
- Modify: `D:/go/cache_manager_api/.codex/docs/feature/2026-03/0325-structured-locator-design.md`
- Reference: `D:/go/cache_manager_api/internal/app/dtool/plw/locator.go`
- Reference: `D:/go/cache_manager_api/internal/app/dtool/plw/process.go`
- Reference: `D:/go/cache_manager_api/internal/app/dtool/plw/define.go`

**Step 1: 复核所有 Locator 入口**

Run: `rg -n "NewLocator\\(|\\.Locator\\.Do\\(|FindLocator\\(|ElementOp" D:\\go\\cache_manager_api\\internal\\app\\dtool\\plw`
Expected: 能列出 `Process` 层全部旧 Locator 使用点

**Step 2: 记录迁移第一阶段必须覆盖的流程**

将以下流程写入设计文档中的迁移清单：

- `PClick`
- `PInput`
- `PTextContent`
- `PBoolExist`
- `ExistWait`
- `NoExistWait`
- `DoBoolResult`

**Step 3: 明确暂不处理的流程**

在设计文档里追加“非第一阶段改造项”，避免实施时扩大范围：

- 删除元素逻辑
- 页面跳转逻辑
- 与 Locator 无关的 URL 等待逻辑

**Step 4: 提交文档检查**

Run: `Get-Content -Raw -Encoding UTF8 D:\\go\\cache_manager_api\\.codex\\docs\\feature\\2026-03\\0325-structured-locator-design.md`
Expected: 设计文档中已明确第一阶段范围和非目标

### Task 2: 新增结构化 Locator 类型定义

**Files:**
- Create: `D:/go/cache_manager_api/internal/app/dtool/plw/locator_types.go`
- Modify: `D:/go/cache_manager_api/internal/app/dtool/plw/define.go`

**Step 1: 写结构定义测试占位说明**

在计划执行时先补一个轻量测试文件草稿，覆盖：

- `LocatorSpec`
- `LocatorOptions`
- `LocatorFilter`
- `LocatorPick`
- `LocatorInput`
- `ElementAction`
- `ElementResult`

建议测试文件：

- `D:/go/cache_manager_api/internal/app/dtool/plw/locator_parser_test.go`

**Step 2: 新增类型定义文件**

在 `locator_types.go` 中写入最小结构定义，并为结构体和关键字段补齐中文注释。

**Step 3: 收敛旧状态对象的职责**

在 `define.go` 中保留旧 `ElementOp` 仅用于迁移期兼容，并增加中文注释标注“待迁移移除”。

**Step 4: 运行编译校验**

Run: `go test ./internal/app/dtool/plw/...`
Expected: 至少通过编译，不出现类型重复定义或循环引用

**Step 5: Commit**

```bash
git add internal/app/dtool/plw/locator_types.go internal/app/dtool/plw/define.go
git commit -m "feat: add structured locator types"
```

### Task 3: 实现原始输入到 LocatorSpec 的解析器

**Files:**
- Create: `D:/go/cache_manager_api/internal/app/dtool/plw/locator_parser.go`
- Create: `D:/go/cache_manager_api/internal/app/dtool/plw/locator_parser_test.go`

**Step 1: 写失败测试，覆盖结构化配置解析**

测试至少覆盖：

- `spec` 直接透传
- 业务别名字段归一化
- `pick.first` / `pick.last` / `pick.nth` 互斥校验
- `method` 为空报错

**Step 2: 跑单测确认失败**

Run: `go test ./internal/app/dtool/plw -run TestLocatorParser -v`
Expected: FAIL，提示解析器未实现或断言不通过

**Step 3: 实现最小解析器**

实现：

- `Parse(input *LocatorInput) (*LocatorSpec, error)`
- 配置空值校验
- 业务别名转标准字段
- `pick` 互斥校验

**Step 4: 再跑单测**

Run: `go test ./internal/app/dtool/plw -run TestLocatorParser -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/dtool/plw/locator_parser.go internal/app/dtool/plw/locator_parser_test.go
git commit -m "feat: add locator spec parser"
```

### Task 4: 实现旧字符串 Locator 兼容转换

**Files:**
- Create: `D:/go/cache_manager_api/internal/app/dtool/plw/locator_legacy.go`
- Modify: `D:/go/cache_manager_api/internal/app/dtool/plw/locator_parser_test.go`

**Step 1: 写失败测试，覆盖旧字符串兼容**

测试至少覆盖：

- `".submit-btn"` 转换为 `method=locator`
- `".submit-btn|first"` 转换为 `pick.first=true`
- `"!.error-tip"` 转换为 `negate=true`

**Step 2: 跑单测确认失败**

Run: `go test ./internal/app/dtool/plw -run TestLegacyLocator -v`
Expected: FAIL

**Step 3: 实现旧字符串到 LocatorSpec 的转换**

实现最小兼容逻辑：

- 兼容普通 selector
- 兼容 `|first`
- 兼容前缀 `!`

注意：

- 不再扩展新的字符串 DSL
- 兼容层仅负责迁移过渡

**Step 4: 再跑单测**

Run: `go test ./internal/app/dtool/plw -run "TestLocatorParser|TestLegacyLocator" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/dtool/plw/locator_legacy.go internal/app/dtool/plw/locator_parser_test.go
git commit -m "feat: add legacy locator compatibility"
```

### Task 5: 实现 LocatorResolver

**Files:**
- Create: `D:/go/cache_manager_api/internal/app/dtool/plw/locator_resolver.go`
- Create: `D:/go/cache_manager_api/internal/app/dtool/plw/locator_resolver_test.go`

**Step 1: 写失败测试，覆盖标准查询解析**

测试至少覆盖：

- `method=locator`
- `method=role`
- `method=text`
- `chain`
- `pick.first`
- `pick.last`
- `pick.nth`

如果直接对真实 Playwright 对象做单测成本较高，可以先抽象内部构建步骤，优先验证：

- 输入校验
- 构建分支选择
- 过滤与选取顺序

**Step 2: 跑测试确认失败**

Run: `go test ./internal/app/dtool/plw -run TestLocatorResolver -v`
Expected: FAIL

**Step 3: 实现最小 Resolver**

实现：

- 基础 `method` 映射
- `chain` 递归处理
- `pick` 处理
- `timeout_mills` 获取

**Step 4: 再跑测试**

Run: `go test ./internal/app/dtool/plw -run TestLocatorResolver -v`
Expected: PASS 或至少覆盖核心逻辑分支

**Step 5: Commit**

```bash
git add internal/app/dtool/plw/locator_resolver.go internal/app/dtool/plw/locator_resolver_test.go
git commit -m "feat: add locator resolver"
```

### Task 6: 实现动作执行器与统一服务层

**Files:**
- Create: `D:/go/cache_manager_api/internal/app/dtool/plw/locator_action.go`
- Create: `D:/go/cache_manager_api/internal/app/dtool/plw/locator_service.go`
- Create: `D:/go/cache_manager_api/internal/app/dtool/plw/locator_action_test.go`

**Step 1: 写失败测试，覆盖动作执行**

测试至少覆盖：

- `click`
- `input`
- `exist`
- `count`
- `text_content`

需要验证：

- 返回值是否落在 `ElementResult`
- 查询失败与动作失败是否区分

**Step 2: 跑测试确认失败**

Run: `go test ./internal/app/dtool/plw -run "TestElementAction|TestLocatorService" -v`
Expected: FAIL

**Step 3: 实现最小动作执行器**

实现：

- `Execute(locator playwright.Locator, action *ElementAction) (*ElementResult, error)`
- `FindAndExecute(page, input, action, waitMills)` 统一入口

**Step 4: 跑测试确认通过**

Run: `go test ./internal/app/dtool/plw -run "TestElementAction|TestLocatorService" -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/dtool/plw/locator_action.go internal/app/dtool/plw/locator_service.go internal/app/dtool/plw/locator_action_test.go
git commit -m "feat: add locator service and action executor"
```

### Task 7: 将 Process 的高频流程切到新 LocatorService

**Files:**
- Modify: `D:/go/cache_manager_api/internal/app/dtool/plw/process.go`
- Modify: `D:/go/cache_manager_api/internal/app/dtool/plw/locator.go`

**Step 1: 写失败测试或补充集成验证点**

优先覆盖这些流程：

- `PClick`
- `PInput`
- `PTextContent`
- `PBoolExist`
- `ExistWait`
- `NoExistWait`

如果现有没有方便的单测基础，至少补充最小化的流程层测试或明确手工验证脚本。

**Step 2: 跑测试确认失败**

Run: `go test ./internal/app/dtool/plw -v`
Expected: FAIL 或出现旧行为与新接口不一致的问题

**Step 3: 改造 Process 层调用方式**

调整思路：

- 不再通过 `ElementOp.Type` 驱动动作
- 改为显式创建 `ElementAction`
- 从 `ElementResult` 中读取文本、数量、存在性结果

同时为关键分支补中文注释，特别是：

- 迁移期兼容分支
- `exist` / `negate` 判定分支
- `TakeContentMap` 与 `BoolResultMap` 写回分支

**Step 4: 再跑测试**

Run: `go test ./internal/app/dtool/plw -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/app/dtool/plw/process.go internal/app/dtool/plw/locator.go
git commit -m "refactor: migrate process to locator service"
```

### Task 8: 保留旧入口并清理可收敛代码

**Files:**
- Modify: `D:/go/cache_manager_api/internal/app/dtool/plw/locator.go`
- Modify: `D:/go/cache_manager_api/internal/app/dtool/plw/define.go`

**Step 1: 明确旧入口职责**

保留 `locator.go` 仅作为迁移期适配层时，需要在代码注释中标注：

- 该文件只做兼容
- 新逻辑统一走 `LocatorService`
- 后续可删除的旧结构有哪些

**Step 2: 删除已失效逻辑**

删除或下沉以下内容：

- `parseLocator()` 中已迁出的逻辑
- `Do()` 中与新服务重复的动作执行逻辑

前提是不会影响旧调用方。

**Step 3: 跑回归测试**

Run: `go test ./internal/app/dtool/plw -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/app/dtool/plw/locator.go internal/app/dtool/plw/define.go
git commit -m "refactor: narrow legacy locator responsibilities"
```

### Task 9: 补充回归验证与文档说明

**Files:**
- Modify: `D:/go/cache_manager_api/.codex/docs/feature/2026-03/0325-structured-locator-design.md`
- Modify: `D:/go/cache_manager_api/.codex/docs/feature/2026-03/0325-structured-locator-implementation-plan.md`

**Step 1: 跑完整测试**

Run: `go test ./internal/app/dtool/plw/...`
Expected: PASS

**Step 2: 记录验证结果**

在设计文档和计划文档追加：

- 已完成能力
- 保留兼容能力
- 已知未覆盖项

**Step 3: 记录手工验证建议**

至少记录以下手工验证场景：

- 点击按钮
- 输入用户名密码
- 提取文本
- 判断元素存在
- 判断元素不存在

**Step 4: Commit**

```bash
git add .codex/docs/feature/2026-03/0325-structured-locator-design.md .codex/docs/feature/2026-03/0325-structured-locator-implementation-plan.md
git commit -m "docs: finalize structured locator plan"
```

### Task 10: 第二阶段扩展项排期

**Files:**
- Modify: `D:/go/cache_manager_api/.codex/docs/feature/2026-03/0325-structured-locator-design.md`

**Step 1: 列出第二阶段范围**

写入以下扩展能力：

- `filters.has`
- `filters.has_not`
- `filters.has_not_text`
- `frame`
- `and`
- `or`
- 动作级 options

**Step 2: 标注依赖前提**

明确第二阶段开始前需要满足：

- 第一阶段旧流程已切新内核
- 前端可输出 `locator.spec`
- 有最小稳定测试覆盖

**Step 3: 文档复查**

Run: `Get-Content -Raw -Encoding UTF8 D:\\go\\cache_manager_api\\.codex\\docs\\feature\\2026-03\\0325-structured-locator-implementation-plan.md`
Expected: 计划文档结构完整、任务粒度清晰

**Step 4: Commit**

```bash
git add .codex/docs/feature/2026-03/0325-structured-locator-design.md .codex/docs/feature/2026-03/0325-structured-locator-implementation-plan.md
git commit -m "docs: add phase two locator roadmap"
```
