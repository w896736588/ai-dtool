# 问题修改提示词模板 - 实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新增全局"问题修改提示词模板"并在 TaskWorkflow 页面提供快捷按钮弹窗，组合用户改动要求与已解析模板。

**Architecture:** 模板以 key-value 存入 `tbl_home_task_config`，在 TaskWorkflow 页通过专用 API 实时解析后与用户输入拼接展示，不走 per-workflow 存储。

**Tech Stack:** Go (Gin) 后端, Vue 2 + Element Plus 前端

---

### Task 1: 后端 - 新增常量与读取保存

**Files:**
- Modify: `internal/app/dtool/define/home_task_config.go:22`
- Modify: `internal/app/dtool/controller/set.go:1563-1567, 1718-1723` 附近

- [ ] **Step 1: 新增常量 `HomeTaskConfigPromptIssueFix`**

在 `home_task_config.go` 文件的常量组末尾新增一行：

```go
HomeTaskConfigPromptIssueFix = `home_task_prompt_issue_fix`
```

位置：在 `HomeTaskConfigPromptCodeReview` 常量之后。

- [ ] **Step 2: 在 `SetHomeTaskConfigGet` 中读取新配置**

在 `controller/set.go` 的 `SetHomeTaskConfigGet` 函数中，在 `promptCodeReview` 读取之后新增：

```go
promptIssueFix, err := homeTaskConfigValue(define.HomeTaskConfigPromptIssueFix)
if err != nil {
    gsgin.GinResponseError(c, err.Error(), nil)
    return
}
```

在返回的 `map[string]any{}` 末尾新增字段：

```go
`home_task_prompt_issue_fix`: promptIssueFix,
```

- [ ] **Step 3: 在 `SetHomeTaskConfigSave` 中保存新配置**

在 `controller/set.go` 的 `SetHomeTaskConfigSave` 函数中，在 `homeTaskPromptCodeReview` 保存逻辑之后新增：

```go
homeTaskPromptIssueFix := strings.TrimSpace(cast.ToString(dataMap[`home_task_prompt_issue_fix`]))
saveHomeTaskPromptWithLog(define.HomeTaskConfigPromptIssueFix, `问题修改提示词`, homeTaskPromptIssueFix, `工作流-问题修改提示词模板`)
if err := common.DbMain.HomeTaskConfigSave(`问题修改提示词`, define.HomeTaskConfigPromptIssueFix, homeTaskPromptIssueFix, `工作流-问题修改提示词模板`); err != nil {
    gsgin.GinResponseError(c, err.Error(), nil)
    return
}
```

- [ ] **Step 4: 在 `promptConfigKeys` 中注册新 key**

在 `controller/set.go` 的 `promptConfigKeys` map 末尾新增一行：

```go
define.HomeTaskConfigPromptIssueFix: `问题修改提示词`,
```

- [ ] **Step 5: Commit**

```bash
git add internal/app/dtool/define/home_task_config.go internal/app/dtool/controller/set.go
git commit -m "feat: 后端新增问题修改提示词模板常量和读写逻辑"
```

---

### Task 2: 后端 - 新增模板解析 API

**Files:**
- Modify: `internal/app/dtool/controller/task_workflow.go:1376-1425` 附近
- Modify: `internal/app/dtool/router.go:353` 附近

- [ ] **Step 1: 新增 `TaskWorkflowIssueFixResolve` controller 方法**

在 `controller/task_workflow.go` 文件底部（所有现有函数之后）新增方法：

```go
// TaskWorkflowIssueFixResolve 解析问题修改提示词模板。
func TaskWorkflowIssueFixResolve(c *gin.Context) {
    var req _struct.TaskWorkflowInfoRequest
    if err := gsgin.GinPostBody(c, &req); err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    if req.WorkflowID <= 0 {
        gsgin.GinResponseError(c, `工作流id不能为空`, nil)
        return
    }
    workflowInfo, err := common.DbMain.TaskWorkflowInfo(req.WorkflowID)
    if err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    homeTaskInfo, err := common.DbMain.HomeTaskRow(cast.ToInt(workflowInfo[`home_task_id`]))
    if err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    template, err := common.DbMain.HomeTaskConfigValue(define.HomeTaskConfigPromptIssueFix)
    if err != nil {
        gsgin.GinResponseError(c, err.Error(), nil)
        return
    }
    placeholders := buildTaskWorkflowPlaceholderMap(c, homeTaskInfo, workflowInfo)
    resolved := taskWorkflowResolvePlaceholders(template, placeholders)
    gsgin.GinResponseSuccess(c, ``, map[string]string{
        `prompt`: resolved,
    })
}
```

- [ ] **Step 2: 注册路由**

在 `router.go` 的 `taskWorkflow` 函数中，在现有路由列表末尾新增：

```go
tGin.GinPost(`/api/task/workflow/issue-fix/resolve`, controller.TaskWorkflowIssueFixResolve)
```

- [ ] **Step 3: Commit**

```bash
git add internal/app/dtool/controller/task_workflow.go internal/app/dtool/router.go
git commit -m "feat: 新增问题修改提示词实时解析 API"
```

---

### Task 3: 前端 - 设置页新增模板编辑

**Files:**
- Modify: `web/src/components/set/home_task_report.vue:115-133, 386-387, 471-472, 538-539, 530-550`

- [ ] **Step 1: 在模板区域新增子 tab**

在 `home_task_report.vue` 的 `<el-tabs v-model="activePromptTab">` 内部，在"代码检查提示词" tab 之后新增：

```html
<el-tab-pane label="问题修改提示词" name="issue_fix">
  <MdEditor
    v-model="form.home_task_prompt_issue_fix"
    preview-theme="github"
    :preview="true"
    :toolbars="promptEditorToolbars"
    class="prompt-template-editor"
  />
</el-tab-pane>
```

- [ ] **Step 2: 在 data() 的 form 中新增字段**

在 form 对象中，`home_task_prompt_design_plan_requirement` 之后新增：

```javascript
home_task_prompt_issue_fix: '',
```

- [ ] **Step 3: 在 loadConfig 方法中读取新字段**

在 `loadConfig` 方法中新增：

```javascript
this.form.home_task_prompt_issue_fix = response.Data.home_task_prompt_issue_fix || ''
```

- [ ] **Step 4: 在 buildFullPayload 方法中新增字段**

在 `buildFullPayload` 返回对象中新增：

```javascript
home_task_prompt_issue_fix: this.form.home_task_prompt_issue_fix,
```

- [ ] **Step 5: Commit**

```bash
git add web/src/components/set/home_task_report.vue
git commit -m "feat: 设置页新增问题修改提示词模板编辑"
```

---

### Task 4: 前端 - API 工具函数

**Files:**
- Modify: `web/src/utils/base/task_workflow.js:64-66`

- [ ] **Step 1: 新增 API 函数**

在 `task_workflow.js` 的 export default 对象中，`TaskWorkflowNodeStatusUpdate` 之前新增：

```javascript
// TaskWorkflowIssueFixResolve 解析问题修改提示词模板。
function TaskWorkflowIssueFixResolve(workflowId, callBack) {
  base.BasePost('/api/task/workflow/issue-fix/resolve', {
    workflow_id: workflowId,
  }, callBack)
}
```

并在 export default 对象中导出：

```javascript
TaskWorkflowIssueFixResolve,
```

- [ ] **Step 2: Commit**

```bash
git add web/src/utils/base/task_workflow.js
git commit -m "feat: 新增问题修改提示词解析 API 调用"
```

---

### Task 5: 前端 - TaskWorkflow 页新增按钮与弹窗

**Files:**
- Modify: `web/src/components/TaskWorkflow.vue:9-21, 605-648, 720, 1120` 附近

- [ ] **Step 1: 在 header actions 区域新增按钮**

在 `task-workflow-header__actions` div 中，在"刷新"按钮之前新增：

```html
<GitActionButton compact variant="warning" @click="openIssueFixDialog">
  问题修改提示词
</GitActionButton>
```

- [ ] **Step 2: 在 template 末尾（`</el-dialog>` 之后，`</div>` 之前）新增弹窗**

在现有 fragment dialog 之后新增：

```html
<el-dialog
  v-model="issueFixDialogVisible"
  title="问题修改提示词"
  width="900px"
  :close-on-click-modal="false"
  destroy-on-close
>
  <div class="task-workflow-issue-fix">
    <div class="task-workflow-issue-fix__input">
      <div class="task-workflow-issue-fix__label">改动要求</div>
      <el-input
        v-model="issueFixInput"
        type="textarea"
        :rows="4"
        placeholder="请描述需要修改的问题"
      />
    </div>
    <div class="task-workflow-issue-fix__output">
      <div class="task-workflow-issue-fix__label">完整提示词</div>
      <MdEditor
        v-model="issueFixCombinedText"
        preview-theme="github"
        :preview="true"
        :toolbars="['preview', 'fullscreen']"
        class="task-workflow-issue-fix__editor"
      />
    </div>
  </div>
  <template #footer>
    <el-button @click="issueFixDialogVisible = false">关闭</el-button>
    <el-button type="primary" @click="copyIssueFixText">复制到剪贴板</el-button>
  </template>
</el-dialog>
```

- [ ] **Step 3: 在 data() 中新增字段**

在 data() return 对象中新增：

```javascript
issueFixDialogVisible: false,
issueFixInput: '',
issueFixResolvedTemplate: '',
```

- [ ] **Step 4: 在 computed 中新增组合文本**

在 computed 中新增：

```javascript
issueFixCombinedText() {
  const input = (this.issueFixInput || '').trim()
  const template = (this.issueFixResolvedTemplate || '').trim()
  if (!input && !template) return ''
  if (!input) return template
  if (!template) return input
  return input + '\n\n' + template
},
```

- [ ] **Step 5: 在 methods 中新增弹窗和复制方法**

在 methods 中（`copyText` 方法之后）新增：

```javascript
openIssueFixDialog() {
  this.issueFixDialogVisible = true
  this.issueFixInput = ''
  this.issueFixResolvedTemplate = ''
  if (this.workflowId <= 0) return
  taskWorkflowApi.TaskWorkflowIssueFixResolve(this.workflowId, (response) => {
    if (response && response.ErrCode === 0 && response.Data) {
      this.issueFixResolvedTemplate = response.Data.prompt || ''
    }
  })
},
copyIssueFixText() {
  this.copyText(this.issueFixCombinedText, '已复制到剪贴板')
},
```

- [ ] **Step 6: 新增 CSS 样式**

在文件末尾的 `<style>` 中新增：

```css
.task-workflow-issue-fix__input {
  margin-bottom: 16px;
}
.task-workflow-issue-fix__label {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 8px;
}
.task-workflow-issue-fix__output {
  margin-top: 16px;
}
.task-workflow-issue-fix__editor {
  min-height: 300px;
}
```

- [ ] **Step 7: Commit**

```bash
git add web/src/components/TaskWorkflow.vue
git commit -m "feat: TaskWorkflow 页新增问题修改提示词按钮与弹窗"
```

---

### Task 6: 验证

- [ ] **Step 1: 重启后端服务并验证**

重启 Go 后端服务，确认无编译错误。

- [ ] **Step 2: 浏览器验证 - 设置页**

打开 `http://localhost:8080/#/HomeTaskSetting`，切换到"工作流程提示词模板" tab，确认出现"问题修改提示词"子 tab，可编辑并保存。

- [ ] **Step 3: 浏览器验证 - 工作流页**

打开 `http://localhost:8080/#/TaskWorkflow/54`，确认顶部出现"问题修改提示词"按钮，点击弹窗，输入改动要求后下方显示组合文本，点击"复制到剪贴板"可复制。
