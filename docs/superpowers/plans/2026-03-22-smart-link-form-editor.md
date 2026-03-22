# Smart Link Form Editor Implementation Plan

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 将“自定义网页”的链接配置与执行流程配置改造成完整的小表单编辑器，移除用户直接编辑 JSON 或猜测通用字段语义的负担。

**Architecture:** 前端新增两个共享编辑器组件，分别承载链接配置与流程项表单逻辑；页面层仅保留数据加载、弹窗控制与保存回调。流程项编辑界面按 `type` 驱动专属字段表单，但保存时仍映射回现有接口字段，避免本轮重写后端协议。

**Tech Stack:** Vue 3, Element Plus, Options API + Composition API, 现有 SmartLink 接口封装

---

## Chunk 1: 链接配置表单化

### Task 1: 收敛链接配置编辑器

**Files:**
- Create: `web/src/components/smart_link/LinkConfigEditor.vue`
- Modify: `web/src/components/smart_link/link_run.vue`

- [ ] **Step 1: 写出链接配置编辑器骨架**

提供：
- 链接项列表
- 链接项新增/删除
- 账号列表新增/删除
- 信息提取项列表
- 请求拦截项列表

- [ ] **Step 2: 约定表单内部数据结构**

在组件内部维护：
- `linksFormList`
- `cookieExtractList`
- `filterUriList`

- [ ] **Step 3: 写最小映射方法**

实现：
- 接收页面现有 `smartLinkConfig`
- 映射成内部表单结构
- 再映射回 `links` / `show_cookies` / `filter_uris`

- [ ] **Step 4: 在 `link_run.vue` 中替换 `JsonEditCombine`**

移除“链接配置”中的 JSON 编辑入口，接入共享表单组件。

- [ ] **Step 5: 手工检查保存路径**

确认保存时仍调用 `SmartLinkAdd`，但提交内容来自新表单映射结果。

## Chunk 2: 执行流程项表单化

### Task 2: 收敛流程项共享编辑器

**Files:**
- Create: `web/src/components/smart_link/ProcessItemEditor.vue`
- Modify: `web/src/components/smart_link/link_process.vue`
- Modify: `web/src/components/smart_link/link_flow.vue`

- [ ] **Step 1: 提取公共字段**

把：
- `name`
- `type`
- `tip`
- `weight`
- `wait_mills`
- `domain_limit`
- `append_to_replace`
- `is_async`
- `is_error_continue`

统一放到共享组件。

- [ ] **Step 2: 建立步骤类型定义表**

为每种 `type` 定义：
- 标签
- 专属字段
- 提交映射规则
- 回显映射规则

- [ ] **Step 3: 先接入简单步骤**

覆盖：
- `click`
- `input`
- `wait`
- `close`
- `redirect_uri`

- [ ] **Step 4: 接入条件与提取步骤**

覆盖：
- `text_content`
- `wait_url`
- `bool_result`
- `bool_exist`
- `no_exist_wait`
- `canvas_image`

- [ ] **Step 5: 接入复合步骤**

覆盖：
- `login_username_password`
- `delete_element`

- [ ] **Step 6: 在两个页面中替换旧表单**

让 [link_process.vue](C:/work/frog/dev_tool_master/web/src/components/smart_link/link_process.vue) 和 [link_flow.vue](C:/work/frog/dev_tool_master/web/src/components/smart_link/link_flow.vue) 共用同一套编辑器。

## Chunk 3: 页面整合与回显

### Task 3: 统一页面保存与回显逻辑

**Files:**
- Modify: `web/src/components/smart_link/link_run.vue`
- Modify: `web/src/components/smart_link/link_process.vue`
- Modify: `web/src/components/smart_link/link_flow.vue`

- [ ] **Step 1: 统一新建默认值**

保证新建链接和新建流程项时，表单默认值来自共享组件可识别的结构。

- [ ] **Step 2: 统一编辑回显**

保证编辑时：
- 页面数据能进入共享组件
- 共享组件能映射成表单值

- [ ] **Step 3: 统一保存出口**

保证保存时：
- 页面拿到共享组件产出的结构化结果
- 再映射回旧接口字段提交

## Chunk 4: 验证

### Task 4: 运行校验并手工验证

**Files:**
- Modify: `docs/superpowers/specs/2026-03-22-smart-link-form-editor-design.md`

- [ ] **Step 1: 运行 ESLint 定点校验**

Run: `npx eslint src/components/smart_link/link_run.vue src/components/smart_link/link_process.vue src/components/smart_link/link_flow.vue src/components/smart_link/LinkConfigEditor.vue src/components/smart_link/ProcessItemEditor.vue`
Workdir: `web`

- [ ] **Step 2: 启动前端并手工验证**

Run: `npm run serve`
Workdir: `web`

验证：
- 链接配置不再直接编辑 JSON
- 信息提取与请求拦截可通过小表单配置
- 所有流程类型都能在编辑器中找到明确字段
- 列表视图与流程图视图弹窗一致
- 保存后再次打开，显示与刚编辑内容一致

- [ ] **Step 3: 记录已知风险**

若发现历史旧数据无法回显，需要在结果说明中明确标注“本轮不兼容旧数据”。
