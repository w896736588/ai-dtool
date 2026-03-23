# 文件夹默认 Header Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 为接口开发模块增加文件夹默认 Header 配置，并在接口执行时先加载文件夹 Header，再由接口自身 Header 覆盖同名项。

**Architecture:** 在 `tbl_api_dir` 增加 `headers` 持久化字段，目录相关读写接口统一支持该字段；运行时由后端在构建执行对象时完成“目录默认值 + 接口覆盖值”合并；前端文件夹基本信息页复用现有 Header 编辑器进行维护。

**Tech Stack:** Go, Gin, SQLite, Vue 3, Element Plus

---

### Task 1: 后端失败测试

**Files:**
- Modify: `D:/go/cache_manager_api/internal/app/dtool/controller/api_basic_query_test.go`
- Test: `D:/go/cache_manager_api/internal/app/dtool/controller/api_basic_query_test.go`

**Step 1: Write the failing test**

新增测试覆盖：

- `buildFolderBasicInfo` 返回 `headers`
- 目录 header 与接口 header 的合并顺序，接口同名键覆盖目录值

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/dtool/controller`

Expected: FAIL，提示缺少 `headers` 透传或缺少 header 合并函数

**Step 3: Write minimal implementation**

在 `controller` 与 `api` 模块补充目录 header 透传与合并函数。

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/dtool/controller`

Expected: PASS

### Task 2: 后端目录存储与运行逻辑

**Files:**
- Create: `D:/go/cache_manager_api/internal/app/dtool/database/2026/03/202603231453_api_dir_headers.sql`
- Modify: `D:/go/cache_manager_api/internal/app/dtool/controller/api.go`
- Modify: `D:/go/cache_manager_api/internal/app/dtool/api/api.go`

**Step 1: Write the failing test**

复用 Task 1 中的测试作为回归约束。

**Step 2: Run test to verify it fails**

Run: `go test ./internal/app/dtool/controller`

Expected: FAIL

**Step 3: Write minimal implementation**

- 为 `tbl_api_dir` 增加 `headers` 字段迁移
- `CreateDir` 支持 `headers` 校验与持久化
- 目录查询接口返回 `headers`
- 运行时合并目录与接口 headers

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/dtool/controller`

Expected: PASS

### Task 3: 前端文件夹默认 Header 编辑

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/components/api/FolderBasicInfo.vue`
- Modify: `D:/go/cache_manager_api/web/src/components/api/FolderDetail.vue`
- Modify: `D:/go/cache_manager_api/web/src/components/Api.vue`

**Step 1: Write the failing test**

本仓库当前未见该区域现成前端单测基础设施，本任务采用受后端接口约束的最小实现与手工验证。

**Step 2: Run test to verify it fails**

Skip: 无独立前端单测入口，记录为手工验证项。

**Step 3: Write minimal implementation**

- 复用 `HeadersValueEditor`
- 保存文件夹时一并提交 `headers`
- 树节点与 tab 同步 `headers`

**Step 4: Run test to verify it passes**

Run: `go test ./internal/app/dtool/controller`

Expected: PASS，前端通过代码审查与手工验证确认字段流转

### Task 4: 整体验证

**Files:**
- Modify: `D:/go/cache_manager_api/docs/api-import-format.md`

**Step 1: Run targeted verification**

Run: `go test ./internal/app/dtool/controller`

Expected: PASS

**Step 2: Review API docs**

更新目录与接口说明，补充文件夹 `headers` 字段含义与优先级规则。

**Step 3: Final verification**

Run: `go test ./internal/app/dtool/controller`

Expected: PASS
