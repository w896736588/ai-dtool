# AI Model Connectivity Test Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 在 AI 模型列表中增加“测试”按钮，验证当前模型配置的真实连通性。

**Architecture:** 后端新增一个模型测试接口，按 `model_type` 生成最小请求体并向 `provider.base_url + model.uri` 发起请求。前端在模型列表中增加测试按钮，调用该接口并展示成功或失败结果。

**Tech Stack:** Go, Gin, Vue 3, Element Plus

---

## Chunk 1: 后端

### Task 1: 请求构造测试与接口实现

**Files:**
- Modify: `internal/app/dtool/controller/set_ai.go`
- Test: `internal/app/dtool/controller/set_ai_connectivity_test.go`
- Modify: `internal/app/dtool/router.go`

- [ ] **Step 1: Write the failing test**
- [ ] **Step 2: Run test to verify it fails**
Run: `go test -vet=off ./internal/app/dtool/controller -run TestBuildAiModelConnectivityRequest`
Expected: FAIL because helper does not exist yet
- [ ] **Step 3: Write minimal implementation**
- [ ] **Step 4: Run test to verify it passes**
Run: `go test -vet=off ./internal/app/dtool/controller -run TestBuildAiModelConnectivityRequest`
Expected: PASS

## Chunk 2: 前端

### Task 2: 模型列表测试按钮

**Files:**
- Modify: `web/src/utils/base/ai_set.js`
- Modify: `web/src/components/set/ai_provider.vue`

- [ ] **Step 1: Add button and loading state**
- [ ] **Step 2: Call backend and show result**
- [ ] **Step 3: Verify no regression in list operations**

## Chunk 3: 验证

### Task 3: 回归

**Files:**
- Test: `internal/app/dtool/controller/set_ai_connectivity_test.go`

- [ ] **Step 1: Run backend tests**
Run: `go test -vet=off ./internal/app/dtool/controller`
- [ ] **Step 2: Run frontend build**
Run: `npm run prod`
Workdir: `web`
