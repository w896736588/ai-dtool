# Smart Link Node Locator Forms Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** 让自定义网页节点按节点类型使用不同的 Locator 配置表单和 JSON，并优先补齐 `text_content` 对当前后端结构化 Locator 语义的支持。

**Architecture:** 前端继续以结构化 Locator JSON 作为唯一保存协议，但把当前统一的简化 Locator 表单拆成节点级模型。`text_content` 和 `bool_result` 先升级为可输出后端现有 `filters`、`chain`、`pick`、`negate` 语义的高级表单，其他基础节点保留轻量体验并完成内部模型解耦。

**Tech Stack:** Vue 3、Element Plus、现有 `smart_link` 前端组件、Node 侧 `.cjs` 工具测试、Go 后端现有 `LocatorSpec` 执行内核

---

### Task 1: 梳理当前 Locator 前端映射点并补设计引用

**Files:**
- Modify: `D:/go/cache_manager_api/.codex/docs/feature/2026-03/0326-smart-link-node-locator-forms-design.md`
- Reference: `D:/go/cache_manager_api/web/src/components/smart_link/ProcessItemEditor.vue`
- Reference: `D:/go/cache_manager_api/web/src/utils/smart_link_process_validation.cjs`
- Reference: `D:/go/cache_manager_api/web/src/utils/smart_link_process_display.cjs`

**Step 1: 记录本次涉及的核心前端入口**

补充设计文档中的实施影响面，列出：

- `ProcessItemEditor.vue`
- `smart_link_process_validation.cjs`
- `smart_link_process_display.cjs`

**Step 2: 记录节点分层范围**

将以下节点写入设计文档：

- 高级节点：`text_content`、`bool_result`
- 基础节点：`click`、`input`、`bool_exist`、`canvas_image`

**Step 3: 文档复查**

Run: `Get-Content -Encoding UTF8 D:\\go\\cache_manager_api\\.codex\\docs\\feature\\2026-03\\0326-smart-link-node-locator-forms-design.md`
Expected: 设计文档明确实施范围、影响文件和非目标

### Task 2: 先为 Locator 双模式映射写失败测试

**Files:**
- Test: `D:/go/cache_manager_api/web/src/utils/link_run_selection.cjs`
- Test: `D:/go/cache_manager_api/web/src/utils/smart_link_process_validation.cjs`
- Test: `D:/go/cache_manager_api/web/src/utils/smart_link_process_display.cjs`

**Step 1: 写 `text_content` 高级 Locator 序列化测试**

补一个 Node 执行的最小断言脚本，验证输入以下高级表单数据时，输出为：

```json
{
  "spec": {
    "method": "locator",
    "value": ".username",
    "filters": [
      {
        "has_not": {
          "method": "locator",
          "value": ".btn.login_as_reg_btn"
        }
      }
    ]
  }
}
```

**Step 2: 写高级回填测试**

验证已有包含 `filters.has_not` 的结构化 JSON 能被识别为高级模式，并正确拆回表单字段。

**Step 3: 写展示层失败测试**

验证复杂 Locator 在展示层会包含：

- 主定位说明
- “且不包含”说明

**Step 4: 运行测试确认失败**

Run: `node D:\\go\\cache_manager_api\\web\\src\\utils\\link_run_selection.cjs`
Expected: FAIL 或输出不满足新增断言

### Task 3: 抽出高级 Locator 表单的序列化与回填工具

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/components/smart_link/ProcessItemEditor.vue`
- Create or Modify: `D:/go/cache_manager_api/web/src/utils/smart_link_process_validation.cjs`
- Create or Modify: `D:/go/cache_manager_api/web/src/utils/smart_link_process_display.cjs`

**Step 1: 在组件或工具函数中定义高级 Locator 表单结构**

至少定义：

```js
{
  method: 'locator',
  value: '',
  options: {
    exact: false,
    name: ''
  },
  filters: [],
  chain: [],
  pick_mode: 'none',
  nth: 0,
  negate: false,
  timeout_mills: 3000
}
```

**Step 2: 写高级表单转结构化 Locator 的最小实现**

实现：

- `buildAdvancedLocatorPayload(form)`
- `stringifyAdvancedLocatorPayload(form)`

只覆盖当前后端已支持字段：

- `method`
- `value`
- `options`
- `filters`
- `chain`
- `pick`
- `negate`
- `timeout_mills`

**Step 3: 写结构化 Locator 回填到高级表单的最小实现**

实现：

- `deserializeAdvancedLocatorForm(payload)`

要求可识别：

- `filters.has_text`
- `filters.has_not_text`
- `filters.has`
- `filters.has_not`
- `visible`
- `chain`
- `pick`

**Step 4: 运行失败测试确认通过**

Run: `node D:\\go\\cache_manager_api\\web\\src\\utils\\link_run_selection.cjs`
Expected: PASS

### Task 4: 改造 `text_content` 节点表单

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/components/smart_link/ProcessItemEditor.vue`

**Step 1: 先写失败场景**

补一个最小交互验证目标：

- `text_content` 进入高级模式后可以新增 `has_not`
- 保存后 `item.locator` 为结构化 JSON 字符串

如果当前仓库没有现成 Vue 单测基础，至少把该验证点写入注释和手工验证清单。

**Step 2: 为 `text_content` 增加节点专用 Locator 编辑区**

新增字段：

- `locator_editor_mode`
- `locator_advanced_form`

并在模板中只对 `text_content` 展示高级配置区。

**Step 3: 保存逻辑切到节点级序列化**

在 `serializeItem()` 中让 `text_content` 优先使用高级 Locator 序列化逻辑，而不是旧的统一 `locator_structured_form`。

**Step 4: 回填逻辑切到节点级反序列化**

在 `syncFromItem()` 或对应回填逻辑里，检测 `text_content` 的 `locator` 是否包含 `filters` / `chain`，并自动回填到高级表单。

**Step 5: 手工验证**

Run: `npm --prefix D:\\go\\cache_manager_api\\web run build`
Expected: PASS

### Task 5: 改造 `bool_result` 规则内部 Locator

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/components/smart_link/ProcessItemEditor.vue`

**Step 1: 写失败验证点**

明确以下场景当前不支持：

- `bool_result` 规则中配置 `has_not`

**Step 2: 将规则项内部 Locator 表单切到高级模型**

每条规则的 `locator_structured_form` 升级为可承载高级 Locator。

**Step 3: 统一序列化**

规则保存时输出：

```json
[
  {
    "locator": {
      "spec": { ... }
    },
    "return": true
  }
]
```

其中 `spec` 可包含 `filters` / `chain`。

**Step 4: 运行构建验证**

Run: `npm --prefix D:\\go\\cache_manager_api\\web run build`
Expected: PASS

### Task 6: 保持基础节点简单但改成节点独立模型

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/components/smart_link/ProcessItemEditor.vue`

**Step 1: 拆出基础节点公用简化 Locator 模型**

不要再让所有节点硬编码依赖同一个 `locator_structured_form`。

**Step 2: 为基础节点保留现有易用表单**

保持：

- `click`
- `input`
- `bool_exist`
- `canvas_image`

的原有简单配置体验。

**Step 3: 保证旧数据不回归**

验证已有简单结构化 JSON 仍可正确回填、编辑和保存。

**Step 4: 构建校验**

Run: `npm --prefix D:\\go\\cache_manager_api\\web run build`
Expected: PASS

### Task 7: 补齐校验逻辑

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/utils/smart_link_process_validation.cjs`

**Step 1: 写失败测试**

至少覆盖：

- 高级模式缺少根 `method`
- 高级模式缺少根 `value`
- `has_not` 子 `spec` 不完整
- `pick` 互斥不合法

**Step 2: 跑测试确认失败**

Run: `node D:\\go\\cache_manager_api\\web\\src\\utils\\smart_link_process_validation.cjs`
Expected: FAIL

**Step 3: 实现最小校验**

按节点类型区分：

- 简化节点沿用当前规则
- 高级节点新增复杂结构校验

**Step 4: 再跑测试确认通过**

Run: `node D:\\go\\cache_manager_api\\web\\src\\utils\\smart_link_process_validation.cjs`
Expected: PASS

### Task 8: 补齐展示逻辑

**Files:**
- Modify: `D:/go/cache_manager_api/web/src/utils/smart_link_process_display.cjs`
- Modify: `D:/go/cache_manager_api/web/src/components/smart_link/link_flow.vue`

**Step 1: 写失败测试**

至少覆盖：

- `filters.has_not` 展示为“且不包含”
- `filters.has` 展示为“且包含”
- `has_text` / `has_not_text` 展示为文本过滤说明
- `chain` 展示为向下查找说明

**Step 2: 跑测试确认失败**

Run: `node D:\\go\\cache_manager_api\\web\\src\\utils\\smart_link_process_display.cjs`
Expected: FAIL

**Step 3: 实现最小展示格式化**

增强复杂 Locator 的文案格式化，但保持简单 Locator 显示风格不变。

**Step 4: 再跑测试确认通过**

Run: `node D:\\go\\cache_manager_api\\web\\src\\utils\\smart_link_process_display.cjs`
Expected: PASS

### Task 9: 做前端整体回归验证

**Files:**
- Modify: `D:/go/cache_manager_api/.codex/docs/feature/2026-03/0326-smart-link-node-locator-forms-design.md`
- Modify: `D:/go/cache_manager_api/.codex/docs/feature/2026-03/0326-smart-link-node-locator-forms-implementation-plan.md`

**Step 1: 运行前端构建**

Run: `npm --prefix D:\\go\\cache_manager_api\\web run build`
Expected: PASS

**Step 2: 运行相关 Node 侧验证脚本**

Run: `node D:\\go\\cache_manager_api\\web\\src\\utils\\smart_link_process_validation.cjs`
Expected: PASS

Run: `node D:\\go\\cache_manager_api\\web\\src\\utils\\smart_link_process_display.cjs`
Expected: PASS

**Step 3: 补充文档中的验证结果**

记录：

- 已支持的复杂 Locator 语义
- 仍保留简单模式的节点
- 已知未覆盖项

### Task 10: 请求代码审查并收尾

**Files:**
- Reference: `D:/go/cache_manager_api/web/src/components/smart_link/ProcessItemEditor.vue`
- Reference: `D:/go/cache_manager_api/web/src/utils/smart_link_process_validation.cjs`
- Reference: `D:/go/cache_manager_api/web/src/utils/smart_link_process_display.cjs`

**Step 1: 自查**

检查以下问题：

- 是否引入了重复的序列化逻辑
- 是否破坏旧节点保存协议
- 是否把高级模式错误暴露给基础节点

**Step 2: 请求代码审查**

使用项目中的代码审查流程对本分支变更做一次审查。

**Step 3: 记录剩余风险**

至少记录：

- 尚未升级高级表单的基础节点范围
- 仅覆盖当前后端已支持语义
