# Playwright 结构化 Locator 重构设计

## 背景

当前 Playwright 元素定位能力主要集中在 `internal/app/dtool/plw/locator.go`，存在以下问题：

1. `Locator` 同时承担了配置解析、`playwright.Locator` 构建、元素动作执行三类职责，边界不清晰。
2. `parseLocator()` 基于字符串拆分和少量自定义约定工作，仅支持 `|first`、`!selector` 等有限语法，扩展成本高。
3. `ElementOp` 同时承载动作入参和动作结果，例如 `FillValue`、`TextContent`、`Count`，状态混杂，不利于后续扩展和调试。
4. `Do()` 内部通过 `switch` 硬编码动作类型，导致每新增一种动作或一种 Locator 语义，都需要继续堆叠判断分支。
5. 当前字符串 Locator 表达式与 Playwright 原生查询语义存在偏差，不利于完整支持 `role`、`text`、`filter`、`nth`、`frame` 等能力。

本次目标不是继续增强旧字符串 DSL，而是升级为结构化配置，并在后端内部统一归一化成接近 Playwright 原生语义的标准结构。

## 目标

1. 将 Locator 查询配置升级为结构化对象，替代当前字符串拼接解析方案。
2. 对外兼容业务化字段别名，对内统一归一化为 Playwright 风格的 `LocatorSpec`。
3. 将 Locator 查询解析、Locator 构建、元素动作执行彻底解耦。
4. 第一阶段在最小影响现有流程的前提下，完成查询内核替换，并保留旧配置兼容能力。
5. 为后续补齐 `filter`、`nth`、`last`、`has`、`frame` 等 Playwright 查询能力预留稳定扩展点。

## 非目标

1. 第一阶段不要求一次性替换所有前端旧配置。
2. 第一阶段不强制实现所有 Playwright API，只优先覆盖当前流程必需能力和高频查询能力。
3. 第一阶段不对 `Process` 的业务流程做大规模重写，仅做查询与动作链路的内核替换。

## 现状问题拆解

### 问题一：Locator 结构职责过多

当前 `Locator` 结构包含：

- 原始 Locator 字符串
- `Page`
- `ElementOp`
- `log`

它既负责：

1. 解析字符串 Locator 表达式
2. 调用 `page.Locator(...)` 构建 Playwright Locator
3. 根据 `ElementOp.Type` 执行动作
4. 把动作结果写回 `ElementOp`

这种设计会导致单个结构在扩展时持续膨胀，后续新增 `hover`、`check`、`get_attribute`、`filter(has)` 等能力时，复杂度会继续累积。

### 问题二：字符串 DSL 表达力不足

当前仅支持以下少量语义：

- `selector`
- `selector|first`
- `!selector`
- `&&`
- `||`

而 Playwright 常用查询能力包括：

- `locator`
- `getByRole`
- `getByText`
- `getByLabel`
- `getByPlaceholder`
- `getByAltText`
- `getByTitle`
- `getByTestId`
- `filter(hasText)`
- `filter(has)`
- `nth`
- `first`
- `last`
- `frameLocator`

继续通过字符串扩展这些能力，会让解析规则越来越零散，也会让前端配置越来越难维护。

### 问题三：动作状态设计耦合

`ElementOp` 当前既表示动作类型，也表示动作输入参数与输出结果：

- `Type`
- `FillValue`
- `TextContent`
- `Count`

这会导致：

1. 同一个结构在不同流程中承担不同语义。
2. 函数返回值不完整，需要额外回写共享状态。
3. 一旦加入重试、并发或嵌套调用，容易出现状态污染。

## 推荐方案

推荐采用“外部兼容业务字段，内部统一归一化为 Playwright 原生语义”的方案：

1. 对外暴露结构化 `locator.spec` 配置。
2. 对内统一转换成标准 `LocatorSpec`。
3. 使用独立的 `LocatorResolver` 负责构建 `playwright.Locator`。
4. 使用独立的 `ElementActionExecutor` 负责执行元素动作。
5. 使用 `LocatorService` 作为 `Process` 层统一入口，屏蔽解析与执行细节。

## 核心数据结构设计

### LocatorInput

`LocatorInput` 表示原始 Locator 输入，用于兼容迁移期旧配置与新配置。

```go
// LocatorInput 表示流程层接收的原始定位配置。
// 迁移期内同时兼容旧版字符串配置和新版结构化配置。
type LocatorInput struct {
	// Raw 表示旧版字符串 Locator，仅迁移期兼容使用。
	Raw string `json:"raw,omitempty"`

	// Spec 表示新版结构化 Locator 配置。
	Spec *LocatorSpec `json:"spec,omitempty"`

	// LegacyFirst 表示旧版 first 兼容字段，仅迁移期保留。
	LegacyFirst *bool `json:"first,omitempty"`
}
```

### LocatorSpec

`LocatorSpec` 表示后端内部统一后的标准查询结构。

```go
// LocatorSpec 表示标准化后的元素定位配置。
// 该结构统一承载 Playwright 原生查询语义，避免业务层直接依赖零散字段。
type LocatorSpec struct {
	// Method 表示元素查询方式，例如 locator、role、text、label。
	Method string `json:"method"`

	// Value 表示查询主值，例如选择器、文本或 role 名称。
	Value string `json:"value"`

	// Options 表示查询附加参数，例如 exact、name 等。
	Options *LocatorOptions `json:"options,omitempty"`

	// Filters 表示当前节点上的过滤条件，例如 has_text、has。
	Filters []LocatorFilter `json:"filters,omitempty"`

	// Chain 表示链式子查询，会基于当前 Locator 继续向下查找。
	Chain []LocatorSpec `json:"chain,omitempty"`

	// Pick 表示最终从匹配结果中选择 first、last 或 nth。
	Pick *LocatorPick `json:"pick,omitempty"`

	// Negate 表示反向存在判断，仅用于存在性校验场景。
	Negate bool `json:"negate,omitempty"`

	// TimeoutMills 表示当前 Locator 的超时时间，单位毫秒。
	TimeoutMills float64 `json:"timeout_mills,omitempty"`
}
```

### LocatorOptions

```go
// LocatorOptions 表示 Locator 查询的附加参数。
type LocatorOptions struct {
	// Exact 表示是否精确匹配。
	Exact *bool `json:"exact,omitempty"`

	// Name 表示 role 查询时的 name 参数。
	Name string `json:"name,omitempty"`

	// Checked 表示 role 查询时的 checked 参数。
	Checked *bool `json:"checked,omitempty"`

	// Disabled 表示 role 查询时的 disabled 参数。
	Disabled *bool `json:"disabled,omitempty"`

	// Selected 表示 role 查询时的 selected 参数。
	Selected *bool `json:"selected,omitempty"`

	// Expanded 表示 role 查询时的 expanded 参数。
	Expanded *bool `json:"expanded,omitempty"`

	// IncludeHidden 表示 role 查询时是否包含隐藏元素。
	IncludeHidden *bool `json:"include_hidden,omitempty"`

	// Level 表示标题类 role 的层级限制。
	Level *int `json:"level,omitempty"`
}
```

### LocatorFilter

```go
// LocatorFilter 表示对当前 Locator 结果做进一步过滤。
type LocatorFilter struct {
	// HasText 表示仅保留包含指定文本的元素。
	HasText string `json:"has_text,omitempty"`

	// HasNotText 表示排除包含指定文本的元素。
	HasNotText string `json:"has_not_text,omitempty"`

	// Has 表示仅保留内部还能匹配到指定子 Locator 的元素。
	Has *LocatorSpec `json:"has,omitempty"`

	// HasNot 表示排除内部还能匹配到指定子 Locator 的元素。
	HasNot *LocatorSpec `json:"has_not,omitempty"`

	// Visible 表示是否仅保留可见元素。
	Visible *bool `json:"visible,omitempty"`
}
```

### LocatorPick

```go
// LocatorPick 表示从匹配结果中选择特定元素。
type LocatorPick struct {
	// First 表示仅取第一个匹配结果。
	First *bool `json:"first,omitempty"`

	// Last 表示仅取最后一个匹配结果。
	Last *bool `json:"last,omitempty"`

	// Nth 表示取第 N 个匹配结果，从 0 开始。
	Nth *int `json:"nth,omitempty"`
}
```

### ElementAction 与 ElementResult

```go
// ElementAction 表示元素动作定义。
type ElementAction struct {
	// Type 表示动作类型，例如 click、input、exist、count。
	Type string `json:"type"`

	// Value 表示动作参数，例如输入框填充值。
	Value string `json:"value,omitempty"`

	// Options 表示动作附加参数，例如 click timeout、force。
	Options *ElementActionOptions `json:"options,omitempty"`
}

// ElementActionOptions 表示动作附加参数。
type ElementActionOptions struct {
	// TimeoutMills 表示动作执行超时时间，单位毫秒。
	TimeoutMills float64 `json:"timeout_mills,omitempty"`

	// Force 表示是否强制执行动作。
	Force *bool `json:"force,omitempty"`

	// AttributeName 表示属性提取动作需要读取的属性名。
	AttributeName string `json:"attribute_name,omitempty"`
}

// ElementResult 表示动作执行结果。
type ElementResult struct {
	// Locator 表示最终解析出的 Playwright Locator。
	Locator playwright.Locator

	// Exists 表示元素是否存在。
	Exists bool

	// Count 表示元素数量。
	Count int

	// TextContent 表示提取到的 textContent。
	TextContent string

	// InnerText 表示提取到的 innerText。
	InnerText string

	// AttributeValue 表示提取到的属性值。
	AttributeValue string
}
```

## 查询能力支持范围

### 第一阶段建议支持

建议第一阶段优先支持以下查询入口和能力：

1. `locator`
2. `role`
3. `text`
4. `label`
5. `placeholder`
6. `alt_text`
7. `title`
8. `test_id`
9. `filters.has_text`
10. `pick.first`
11. `pick.last`
12. `pick.nth`
13. `negate`
14. `chain`

这些能力已经足够覆盖当前大多数页面自动化场景，并且与 Playwright 的常用查询方式保持一致。

### 第二阶段建议补充

第二阶段再按需补齐：

1. `filters.has`
2. `filters.has_not`
3. `filters.has_not_text`
4. `frame`
5. 组合查询 `and`
6. 组合查询 `or`

第二阶段的重点是增强复杂页面和多层嵌套场景，不建议和第一阶段一起落地，以免改动面过大。

## 典型配置示例

### 按 role 点击按钮

```json
{
  "type": "click",
  "locator": {
    "spec": {
      "method": "role",
      "value": "button",
      "options": {
        "name": "提交",
        "exact": true
      },
      "pick": {
        "first": true
      }
    }
  }
}
```

### 按 label 输入内容

```json
{
  "type": "input",
  "value": "{username}",
  "locator": {
    "spec": {
      "method": "label",
      "value": "用户名",
      "options": {
        "exact": true
      }
    }
  }
}
```

### 查容器后定位子按钮

```json
{
  "type": "click",
  "locator": {
    "spec": {
      "method": "locator",
      "value": ".dialog-wrapper",
      "chain": [
        {
          "method": "role",
          "value": "button",
          "options": {
            "name": "确认"
          }
        }
      ]
    }
  }
}
```

### 提取文本

```json
{
  "type": "text_content",
  "out_key": "user_name",
  "locator": {
    "spec": {
      "method": "locator",
      "value": ".user-profile",
      "chain": [
        {
          "method": "locator",
          "value": ".name"
        }
      ]
    }
  }
}
```

### 判断元素不存在

```json
{
  "type": "bool_exist",
  "out_key": "error_disappeared",
  "locator": {
    "spec": {
      "method": "text",
      "value": "系统异常",
      "negate": true,
      "timeout_mills": 1000
    }
  }
}
```

## 模块拆分建议

建议在 `internal/app/dtool/plw/` 下按职责拆分以下文件：

1. `locator_types.go`
2. `locator_parser.go`
3. `locator_resolver.go`
4. `locator_action.go`
5. `locator_service.go`
6. `locator_legacy.go`

各文件建议职责如下：

### locator_types.go

存放：

- `LocatorInput`
- `LocatorSpec`
- `LocatorOptions`
- `LocatorFilter`
- `LocatorPick`
- `ElementAction`
- `ElementActionOptions`
- `ElementResult`

### locator_parser.go

负责：

1. 解析原始输入
2. 处理业务别名字段
3. 归一化成标准 `LocatorSpec`
4. 校验 `pick`、`method`、`value` 等配置合法性

### locator_resolver.go

负责：

1. 将 `LocatorSpec` 转换成 `playwright.Locator`
2. 根据 `Method` 选择具体 Playwright 查询入口
3. 递归处理 `Filters`、`Chain`、`Pick`

### locator_action.go

负责：

1. 执行动作
2. 处理动作返回值
3. 把查询错误和动作错误区分开

### locator_service.go

负责串联：

1. 解析
2. 查询
3. 动作执行
4. 统一错误包装与日志输出

### locator_legacy.go

仅迁移期存在，用于：

1. 将旧字符串配置转换成 `LocatorSpec`
2. 兼容 `!selector`
3. 兼容 `selector|first`

待全部旧配置迁移后，可删除该文件。

## Process 层改造方案

### 当前问题

`Process` 当前通过设置 `ElementOp.Type` 来隐式驱动动作执行，例如：

```go
h.ElementOp.Type = define.ElementClick
_, elementErr := h.Locator.Do(h.WaitMills)
```

这种方式的缺点是：

1. 动作参数依赖共享状态对象传递。
2. 动作结果需要通过共享状态对象回写。
3. `Process` 无法明确知道动作输入和动作输出。

### 推荐改造方式

将 `Process` 改为显式构造动作请求，并调用统一的 `LocatorService`：

```go
// 构造点击动作，避免继续依赖共享状态对象。
action := &ElementAction{
	Type: define.ElementClick,
}

// 执行 Locator 查找与动作。
result, err := h.LocatorService.FindAndExecute(h.Page, h.LocatorInput, action, h.WaitMills)
```

例如 `PInput()` 的目标形态可以是：

```go
func (h *Process) PInput() (define.ProcessCode, string, error) {
	// 先做变量替换，确保输入动作使用最终值。
	value := p_common.Replace(h.Value, h.runParams.ReplaceList)

	// 构造输入动作。
	action := &ElementAction{
		Type:  define.ElementInput,
		Value: value,
	}

	// 执行元素定位与输入动作。
	_, err := h.LocatorService.FindAndExecute(h.Page, h.LocatorInput, action, h.WaitMills)
	if err != nil {
		return define.ProcessBreak, `获取需要输入的元素失败`, err
	}
	return define.ProcessOk, ``, nil
}
```

`PTextContent()` 则通过 `ElementResult` 读取结果：

```go
func (h *Process) PTextContent() (define.ProcessCode, string, error) {
	// 构造文本提取动作。
	action := &ElementAction{
		Type: define.ElementTextContent,
	}

	// 执行元素定位与文本提取。
	result, err := h.LocatorService.FindAndExecute(h.Page, h.LocatorInput, action, h.WaitMills)
	if err != nil {
		h.TakeContentMap[h.OutKey] = ``
		return define.ProcessOk, ``, nil
	}

	h.TakeContentMap[h.OutKey] = result.TextContent
	return define.ProcessOk, ``, nil
}
```

## 迁移策略

### 原则

迁移应采用“兼容旧入口、内部统一替换”的方式，不建议一次性删除旧逻辑。

### 推荐顺序

1. 新增标准结构和服务层，不删除旧 `Locator` 文件。
2. 新增兼容转换逻辑，将旧字符串 Locator 转成标准 `LocatorSpec`。
3. 先改 `PClick`、`PInput`、`PTextContent`、`PBoolExist`、`ExistWait`、`NoExistWait` 这些高频流程。
4. `DoBoolResult` 保留旧入口，但内部改成走新解析和执行链。
5. 待所有流程切换完成后，再删除 `parseLocator()`、`ElementOp`、`Locator.Do()` 中的动作 `switch`。

### 兼容策略

迁移期内解析规则建议为：

1. 若存在 `locator.spec`，优先走结构化配置。
2. 若不存在 `locator.spec`，但存在旧字符串 `locator/raw`，则先转换为标准 `LocatorSpec`。
3. 若两者都不存在，则返回配置错误。

### 建议逐步废弃的旧字段

1. `Locators string`
2. `parseLocator()`
3. `LocatorParse`
4. `ElementOp`
5. `ElementOp.FillValue`
6. `ElementOp.TextContent`
7. `ElementOp.Count`

## 动作能力扩展建议

除了当前已有动作，建议在动作执行器中预留以下动作类型：

1. `click`
2. `input`
3. `exist`
4. `count`
5. `text_content`
6. `inner_text`
7. `get_attribute`
8. `hover`
9. `check`
10. `uncheck`
11. `select_option`

这批动作属于页面自动化的高频能力，提前纳入统一动作层后，后续扩展成本会明显下降。

## 错误处理规范

建议统一区分三类错误：

1. 配置错误
2. 查询失败
3. 动作失败

例如：

- `method` 为空，属于配置错误
- 元素等待超时，属于查询失败
- `Click()` 报错，属于动作失败

这样日志输出和前端提示可以更准确，也更方便排障。

## 测试建议

第一阶段应至少覆盖以下测试场景：

1. 旧字符串 `selector` 能被正确转换为 `LocatorSpec`
2. 旧字符串 `selector|first` 能被正确转换
3. 旧字符串 `!selector` 能正确触发反向存在判断
4. 结构化 `role + name` 查询能正确构建 Locator
5. `chain` 能正确拼接子 Locator
6. `pick.first`、`pick.last`、`pick.nth` 互斥校验正确
7. `click`、`input`、`text_content`、`count`、`exist` 能返回预期结果
8. `negate` 场景下存在与不存在的返回结果正确

## 推荐结论

推荐本次以“结构化 Locator + 内核统一重构”的方式推进：

1. 对外新增 `locator.spec`
2. 对内引入 `LocatorSpec + LocatorResolver + ElementActionExecutor + LocatorService`
3. 保留旧字符串配置兼容，但所有执行链统一走新内核
4. 第一阶段先替换查询与动作内核，第二阶段再补齐更复杂的 Playwright 查询能力

这样可以在不一次性推翻现有流程的情况下，先解决当前最核心的两个问题：

1. 配置解析逻辑混乱
2. Locator 查询和元素动作强耦合

后续再新增查询能力时，只需要扩展标准结构和解析器，不需要继续堆叠新的字符串语法和动作分支。
