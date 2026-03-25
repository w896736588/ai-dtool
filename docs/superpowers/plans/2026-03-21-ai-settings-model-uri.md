# AI Settings Model URI Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将 AI 设置拆分为“服务商基础域名 + 模型 URI”，并新增 `llm/embedding` 模型类型区分，同时保持历史配置可迁移。

**Architecture:** 通过数据库迁移为模型增加 `uri` 与 `model_type` 字段，并将旧服务商完整地址清洗成基础域名。后端接口和 AI 调用链统一使用 `provider.base_url + model.uri`，前端设置页同步展示和编辑新字段。

**Tech Stack:** Go, Gin, SQLite, Vue 3, Element Plus

---

## Chunk 1: 数据与后端

### Task 1: 数据库迁移与模型接口测试

**Files:**
- Create: `internal/app/dtool/database/2026/03/20260321.ai_model_uri_type.sql`
- Modify: `internal/app/dtool/controller/set_ai.go`
- Modify: `internal/app/dtool/common/info_crawl.go`
- Modify: `internal/app/dtool/common/info_crawl_ai.go`
- Test: `internal/app/dtool/controller/set_ai_test.go`

- [ ] **Step 1: 写失败测试**

覆盖：
- 新增模型时缺少 `uri` 报错
- 空 `model_type` 默认写成 `llm`
- `uri` 自动补 `/`

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/app/dtool/controller -run TestSetAiModel`
Expected: FAIL，提示新字段行为尚未实现

- [ ] **Step 3: 实现最小后端改动**

实现：
- 模型接口接收/返回 `uri` 与 `model_type`
- 统一规范化 `base_url` 与 `uri`
- AI 调用使用拼接地址

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/app/dtool/controller -run TestSetAiModel`
Expected: PASS

### Task 2: 调用链兼容测试

**Files:**
- Modify: `internal/app/dtool/variable/r_cmd.go`
- Test: `internal/app/dtool/variable/r_cmd_llm_test.go`

- [ ] **Step 1: 写失败测试**

覆盖：
- 使用基础域名时仍能生成正确 chat completions 地址

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/app/dtool/variable -run Test`
Expected: FAIL，旧逻辑仍把 `base_url` 当完整地址

- [ ] **Step 3: 实现最小代码**

让变量模块兼容基础域名输入，保留旧默认值行为。

- [ ] **Step 4: 运行测试确认通过**

Run: `go test ./internal/app/dtool/variable -run Test`
Expected: PASS

## Chunk 2: 前端设置页

### Task 3: 设置页表单与列表改造

**Files:**
- Modify: `web/src/components/set/ai_provider.vue`

- [ ] **Step 1: 先根据接口结构调整前端状态与渲染**

新增模型类型和 URI 的列表列、表单项、默认值与文案。

- [ ] **Step 2: 本地检查关键交互**

验证：
- 服务商只录基础域名
- 新增模型默认类型为 `llm`
- URI 允许输入 `/v1/chat/completions`

- [ ] **Step 3: 清理兼容映射**

保留 `request_format` 兼容映射，不引入额外重构。

## Chunk 3: 验证

### Task 4: 回归验证

**Files:**
- Modify: `README.md`（仅在配置说明确实需要补充时）

- [ ] **Step 1: 运行后端相关测试**

Run: `go test ./internal/app/dtool/controller ./internal/app/dtool/variable ./internal/app/dtool/common`

- [ ] **Step 2: 前端最小构建验证**

Run: `npm run build`
Workdir: `web`

- [ ] **Step 3: 记录未覆盖风险**

如果未运行完整端到端验证，明确说明剩余风险点。
