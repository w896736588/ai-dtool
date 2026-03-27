# 自定义网页节点差异化 Locator 表单设计

**日期：** 2026-03-26

## 背景

当前自定义网页节点编辑器对大多数 `locator` 字段统一使用同一套简化结构化表单，能够覆盖基础的单点定位，但无法表达当前后端已支持的更复杂结构化 Locator 语义，例如：

- 主元素存在，同时要求某个子元素不存在
- 使用 `filters.has` / `filters.has_not`
- 使用 `filters.has_text` / `filters.has_not_text`
- 使用 `chain` 向下继续查找
- 使用 `pick.first` / `pick.last` / `pick.nth`

因此像 `.username && !.btn.login_as_reg_btn` 这类原本依赖组合语义的定位条件，前端当前无法配置出来。问题主要在前端建模和表单能力不足，不在后端执行内核。

## 目标

本次只支持当前后端已经实现的结构化 Locator 语义，不新增后端 DSL，不恢复旧字符串组合表达式协议。

本次要达到的结果：

- `text_content` 节点可以配置复杂结构化 Locator
- 不同节点类型可以拥有不同的 Locator 表单和不同的 JSON 输出策略
- 已有基础节点不回归，旧配置仍可回显和保存
- 最终仍统一保存为后端当前可识别的结构化 JSON

## 非目标

本次不做以下事项：

- 不新增新的后端 Locator 字段
- 不恢复 `&&` / `||` / `!selector` 字符串 DSL 作为正式保存协议
- 不一次性把所有节点都做成完整高级编辑器
- 不修改 Playwright Locator 内核语义

## 现状问题

### 1. 表单建模过于统一

`ProcessItemEditor.vue` 目前把绝大多数节点的 `locator` 都收敛为单个 `locator_structured_form`，字段主要围绕：

- `kind`
- `value`
- `target_text`
- `exact`
- `negate`
- `timeout_mills`
- `pick_mode`
- `nth`

这套表单只能生成单个根级 `spec`，无法继续表达：

- `filters`
- 嵌套 `has` / `has_not`
- `chain`
- 节点类型差异化的定位语义

### 2. 节点语义没有前端分层

当前 `text_content`、`click`、`input`、`bool_exist` 等节点虽然业务语义不同，但前端大体共用同一种 Locator 编辑方式，导致：

- 对提取类节点来说能力不够
- 对简单动作类节点来说表单又不够聚焦

### 3. 校验与展示逻辑只覆盖简化结构

`smart_link_process_validation.cjs` 与 `smart_link_process_display.cjs` 目前主要校验和展示单一结构化 `spec`，对于 `filters` / `chain` 的覆盖不足。

## 设计原则

### 1. 后端协议优先

前端配置能力以当前后端 `LocatorSpec` 能力为边界，不自创新语义。

### 2. 节点类型分层

不同节点类型使用不同的 Locator 表单模型，但最终都序列化为后端结构化 JSON。

### 3. 渐进增强

先让复杂节点可用，再保持简单节点易用，不做一次性大改。

### 4. 兼容旧数据

历史 `locator` 结构化 JSON 需要可回填；已有简单配置不能因为切表单而失效。

## 方案

### 一、引入按节点类型区分的 Locator 表单模型

在前端引入“节点类型 -> Locator 编辑模式”的分层策略。

建议分为两类：

- 基础 Locator 节点：`click`、`input`、`bool_exist`、`canvas_image`
- 高级 Locator 节点：`text_content`、`bool_result`

其中：

- 基础节点继续使用简化表单，但内部数据结构改为节点独立的 `locator_form`
- 高级节点新增“高级结构化定位表单”，支持后端已有的 `filters`、`chain`、`pick` 等配置

这满足“每种节点类型可以有不同配置表单和 JSON”的要求，同时避免所有节点都被复杂化。

### 二、`text_content` 节点新增高级结构化 Locator 编辑器

`text_content` 是本次必须优先解决的节点。

建议为它新增一套专用表单，直接映射后端已有字段：

- 根节点
  - `method`
  - `value`
  - `options`
  - `timeout_mills`
  - `negate`
- 过滤条件列表
  - `has_text`
  - `has_not_text`
  - `has`
  - `has_not`
  - `visible`
- 链式子定位列表
  - `chain`
- 结果选择
  - `first`
  - `last`
  - `nth`

示例：表达“存在 `.username`，且其中不存在 `.btn.login_as_reg_btn`”时，前端保存为：

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

这样直接复用后端现有 `filters.has_not` 语义，不需要再拼接旧字符串表达式。

### 三、`bool_result` 保持规则列表，但规则内部升级

`bool_result` 目前已经是规则列表结构，每条规则有一个 `locator`。该节点继续保持这种形态，但每条规则的 `locator_structured_form` 需要升级为可支持高级结构。

这样布尔判断场景也可以使用后端现有 `has` / `has_not` 语义。

### 四、基础节点保持轻量，但底层结构统一

`click`、`input`、`bool_exist`、`canvas_image` 先不引入完整高级表单，仍保留现有简单配置体验。

但它们内部不再和 `text_content` 共用单一 locator 表单约束，而是：

- 允许节点级单独控制可展示字段
- 为后续扩展高级定位能力保留结构

这一步主要是把“统一表单”改成“节点各自持有自己的 locator form”。

## 数据结构设计

### 前端编辑态

建议在 `formMeta` 中区分不同节点的 Locator 表单数据，例如：

```js
{
  locator_editor_mode: "simple" | "advanced",
  locator_simple_form: { ... },
  locator_advanced_form: {
    method: "locator",
    value: "",
    options: {
      exact: false,
      name: ""
    },
    filters: [],
    chain: [],
    pick_mode: "none",
    nth: 0,
    negate: false,
    timeout_mills: 3000
  }
}
```

是否启用高级模式由节点类型和已有数据共同决定：

- `text_content` 默认支持高级模式
- 如果回填数据中检测到 `filters` 或 `chain`，自动进入高级模式

### 保存态

所有保存结果继续统一为：

```json
{
  "spec": { ... }
}
```

不引入新的数据库字段，不调整后端接口。

## 回填策略

### 1. 已有简化结构化配置

若 `locator` 为简单 `{spec:{method,value,...}}`，则：

- 基础节点回填到简化表单
- `text_content` 可回填到高级表单

### 2. 已有复杂结构化配置

若 `locator.spec` 中包含：

- `filters`
- `chain`
- 更复杂的 `pick`

则：

- 自动回填到高级表单
- 简化表单不强行降级

### 3. 旧非结构化字符串

当前后端已不再接受普通字符串作为正式 Locator 协议。前端对这种旧数据只做兜底显示，不主动扩展支持范围。

## 校验设计

前端校验按节点类型区分：

- 简化节点校验当前已有必填逻辑
- `text_content` 高级模式新增：
  - 根 `method` 必填
  - 根 `value` 在需要时必填
  - `filters` 中每一项只允许使用后端已支持字段
  - `has` / `has_not` 必须是完整子 `spec`
  - `pick` 只允许一种模式生效

## 展示设计

流程卡片详情展示需要增强：

- 简单结构继续按“CSS / 文本 / 标签”描述
- 如果有 `filters.has_not`，展示“且不包含: ...”
- 如果有 `filters.has`，展示“且包含: ...”
- 如果有 `has_text` / `has_not_text`，展示文本过滤说明
- 如果有 `chain`，展示“再向下查找: ...”

这样保存后用户能直接看懂规则，不必查看原始 JSON。

## 测试设计

### 前端

至少覆盖：

- `text_content` 高级表单可生成 `filters.has_not`
- 复杂结构化 JSON 可正确回填
- 基础节点旧表单保存结果不回归
- 展示文案可正确描述 `has_not`

### 手工验证

至少验证：

- `text_content` 配置主元素 + `has_not`
- `text_content` 配置 `has_text`
- `bool_result` 规则中配置 `has_not`
- `click` / `input` / `bool_exist` 旧配置保存和执行不回归

## 风险

### 1. 回填与保存双向映射复杂度上升

缓解方式：将 Locator 表单转换逻辑抽到独立工具函数，不在组件内散写。

### 2. `bool_result` 与普通节点映射不一致

缓解方式：规则项内部复用同一套高级 Locator 序列化逻辑。

### 3. 展示与校验遗漏复杂结构

缓解方式：为 `validation` 和 `display` 补充用例，至少覆盖 `filters.has_not` 和 `chain`。

## 最终建议

本次实施采用以下收敛范围：

- 优先解决 `text_content`
- 同步让 `bool_result` 规则内部可承接高级 Locator
- 基础节点先完成“节点独立表单模型”收口，但不强行全部升级高级表单

这样能在最小风险下解决当前核心问题，并为后续其他节点扩展留下明确结构。
