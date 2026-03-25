# 首页 script 自定义脚本命令设计

**日期：** 2026-03-19

**目标：**
在首页命令框中新增 `script` 顶级命令，用于执行自定义脚本；移除首页 `variable` 顶级命令；脚本执行过程完全由命令框输入和候选项驱动，不暴露创建、编辑、管理能力。

## 背景

当前首页命令体系中存在 `variable` 顶级命令，用户需要理解 `run / set / choose / exec` 等动作词才能完成脚本执行。需求希望将首页体验统一为“自定义脚本 = script”，并且将脚本执行交互收敛为：

- 入口使用 `script run`
- 启动后如果需要输入，命令框直接等待输入
- 启动后如果需要选项，命令框直接显示可选项
- 脚本已就绪时，命令框直接回车执行

用户不需要继续输入 `set / choose / exec` 这类辅助命令。

## 范围

本次只调整首页命令体系与首页脚本执行交互：

- 新增首页顶级命令 `script`
- 移除首页顶级命令 `variable`
- 保留其他顶级命令不变，如 `git`、`docker`、`link`、`shell`
- 首页只提供脚本执行能力，不提供脚本创建、编辑、管理入口

不在本次范围内：

- Variable 页面功能改造
- 后端脚本数据模型调整
- 其他命令体系重构

## 用户交互

### 入口

用户在首页通过以下方式进入脚本执行：

- 点击首页固定命令区中的 `script`
- 在命令框输入 `script run`

进入后，命令框下拉直接显示脚本列表，用户选择脚本即可启动执行。

### 执行过程

脚本启动后，首页命令框进入脚本会话模式：

- 如果当前步骤是输入型，命令框 placeholder 显示当前步骤提示，用户直接输入值并回车
- 如果当前步骤是选择型，命令框下拉显示当前步骤选项，用户直接选择
- 如果当前步骤已可执行，命令框提示“按回车执行”，用户直接回车

### 文案原则

所有首页用户可见文案统一使用 `script` 或“脚本”语义，不再出现 `variable`。

典型提示文案：

- `请选择要执行的脚本`
- `已启动脚本: xxx`
- `当前步骤: xxx`
- `请在命令框输入内容并回车`
- `请在命令框选择一个选项`
- `脚本已就绪，按回车执行`
- `脚本执行完成`

## 命令模型

首页命令配置调整为：

- 顶级命令：`script`
- 子命令：`run`
- 动态候选：`scriptList`

用户侧保留的脚本命令主路径是：

- `script run`
- `script run <脚本名>`

后续步骤不再要求继续输入额外动作词。

## 状态机设计

首页新增独立 `scriptSession`，用于承载脚本执行状态，不依赖旧的 `variable` 顶级命令语义。

建议状态：

- `idle`：默认态，没有脚本执行会话
- `selecting_script`：正在选择脚本
- `waiting_input`：当前步骤需要文本输入
- `waiting_option`：当前步骤需要从选项中选择
- `ready_execute`：脚本已准备完成，可直接执行
- `executing`：脚本执行中
- `finished`：脚本执行结束，随后回到 `idle`

建议数据结构：

```js
const scriptSession = ref({
  active: false,
  stage: 'idle',
  scriptId: 0,
  scriptName: '',
  currentStep: null,
  pendingInputLabel: '',
  optionList: [],
  canExecute: false,
})
```

## 候选项设计

### 脚本列表候选

在 `script run` 阶段，下拉显示脚本列表。

候选数据建议结构：

```js
{
  command: 'deploy-api',
  name: 'deploy-api',
  desc: '自定义脚本',
  icon: '📝',
  data: {
    id: 123,
    name: 'deploy-api'
  }
}
```

### 步骤选项候选

在 `waiting_option` 阶段，下拉只显示当前步骤的选项。

候选数据建议结构：

```js
{
  command: 'test',
  name: '测试环境',
  desc: 'test',
  icon: '•',
  data: {
    value: 'test',
    label: '测试环境'
  }
}
```

匹配规则：

- 脚本候选同时匹配 `command / name`
- 选项候选同时匹配 `label / value`

## 事件流

建议将首页脚本行为拆成四类动作：

### 1. 进入脚本模式

`enterScriptRunMode`

- 触发：点击 `script` 或输入 `script run`
- 行为：切换到 `selecting_script`，加载 `scriptList`

### 2. 启动脚本

`startScript(script)`

- 触发：选中脚本
- 行为：记录脚本信息，调用启动接口，根据返回进入输入、选项、可执行或完成状态

### 3. 提交当前步骤

`submitScriptStep(value)`

- 触发：
  - `waiting_input` 时输入回车
  - `waiting_option` 时选项确认
- 行为：提交当前步骤并根据返回推进状态

### 4. 执行脚本

`executeScript()`

- 触发：`ready_execute` 时回车
- 行为：进入 `executing`，调用执行接口，完成后进入 `finished`

## 命令框行为

命令框在首页保持“双模式”：

- 普通模式：沿用现有全局命令解析逻辑
- 脚本会话模式：当 `scriptSession.active` 时，回车和候选逻辑优先按当前脚本阶段解释

回车行为约束：

- `idle`：按普通命令执行
- `selecting_script`：确认当前脚本候选
- `waiting_input`：提交输入值
- `waiting_option`：确认当前选项
- `ready_execute`：直接执行脚本
- `executing`：禁止重复提交

## 错误处理

需要覆盖以下错误：

- 脚本列表加载失败：停留在 `selecting_script`
- 启动脚本失败：返回脚本选择态
- 输入提交失败：保留在 `waiting_input`
- 选项提交失败：保留在 `waiting_option`
- 执行失败：保留当前会话并提示错误

## 回归风险

主要风险点：

- `Dashboard.vue` 中候选来源切换逻辑被脚本会话态干扰
- 回车逻辑在脚本会话态与普通命令态之间切换不完整
- placeholder、下拉候选和结果区文案出现旧 `variable` 残留
- 历史命令回填逻辑仍回填 `variable` 顶级命令

## 测试建议

至少覆盖以下场景：

- `script` 出现在首页固定命令区和候选区
- `variable` 不再作为首页顶级命令展示
- `script run` 后直接进入脚本列表
- 输入型步骤时命令框直接接收用户输入
- 选择型步骤时命令框下拉直接展示步骤选项
- 可执行步骤时回车直接执行
- 执行完成后恢复默认态
- 其他命令如 `git`、`docker` 继续按原逻辑工作

## 结论

本方案通过新增独立 `script` 顶级命令并移除首页 `variable` 顶级命令，将“自定义脚本执行”从命令词驱动改为会话驱动。首页用户只需要知道 `script run`，后续的输入、选择与执行都在同一个命令框中自然完成，同时不影响其他命令体系。
