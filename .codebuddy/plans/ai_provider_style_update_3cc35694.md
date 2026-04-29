---
name: ai_provider_style_update
overview: 修改 AI 服务商与模型配置页面的操作栏宽度和按钮风格，使其与 Supervisor 页面保持一致。
todos:
  - id: modify-ai-provider-operations
    content: 修改 ai_provider.vue 操作栏宽度和按钮风格
    status: completed
---

## 用户需求

修改"AI 服务商与模型配置"页面 (ai_provider.vue)：

1. 操作栏宽度加宽
2. 按钮风格调整为 Supervisor 页面风格

## 具体修改内容

### 1. 操作栏宽度调整

- 服务商配置操作列：`width="220"` → `width="280"`
- 模型配置操作列：`width="250"` → `width="280"`

### 2. 按钮风格统一调整

将 ai_provider.vue 中的操作栏按钮从 `type="primary" link` 改为 `type="success" plain`，并添加 `size="small"` 属性（与 Supervisor 页面风格保持一致）

**服务商配置操作栏按钮修改：**

- 编辑/复制新增/管理模型：从 `type="primary" link` → `type="success" plain size="small"`
- 删除：从 `link type="danger"` → `type="danger" plain size="small"`

**模型配置操作栏按钮修改：**

- 编辑/复制新增/测试：从 `type="primary" link` → `type="success" plain size="small"`
- 删除：从 `link type="danger"` → `type="danger" plain size="small"`

### 3. 样式清理

移除 ai_provider.vue 中不再需要的 link 按钮相关 CSS 样式（.el-button--primary.is-link）

## 技术方案

### 修改文件

- `web/src/components/set/ai_provider.vue`

### 实现方式

直接在 Vue 模板中修改 el-table-column 的 width 属性和 pl-button 的 type 属性。

### 样式参考

参照 Supervisor.vue 中的按钮样式：

- 使用 `type="success" plain` 替代 `type="primary" link`
- 使用 `type="danger" plain` 替代 `link type="danger"`
- 添加 `size="small"` 属性使按钮更紧凑
- Supervisor.vue 中使用圆角边框、浅绿背景的朴素按钮风格

# Agent Extensions

无需要使用的 Agent Extensions