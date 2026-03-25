# 自定义网页表单化编辑器设计

## 目标

将“自定义网页”中的链接配置与执行流程配置从当前的半结构化输入方式改为明确的小表单编辑体验，降低配置门槛，避免用户直接理解 JSON 结构或通用字段语义。

## 现状

- [link_run.vue](C:/work/frog/dev_tool_master/web/src/components/smart_link/link_run.vue) 的主编辑弹窗中：
  - `links` 通过 `JsonEditCombine` 编辑，虽然是表单模式，但底层仍要求理解数组结构。
  - `show_cookies`、`filter_uris` 仍是大文本输入。
- [link_process.vue](C:/work/frog/dev_tool_master/web/src/components/smart_link/link_process.vue) 与 [link_flow.vue](C:/work/frog/dev_tool_master/web/src/components/smart_link/link_flow.vue) 的执行流程项编辑器：
  - 仍围绕 `locator`、`value`、`out_key`、`check_key` 等通用字段展开。
  - 不同步骤类型没有真正的专属字段表单，用户需要自行推断每种类型该填什么。
  - 两处表单实现重复，后续维护成本高。
- 用户已明确接受“不兼容旧数据”，因此本轮可以优先清晰的新表单模型，不必为旧配置自动迁移兜底。

## 目标范围

本轮覆盖两块：

1. 自定义网页主配置编辑弹窗
2. 执行流程项编辑器

不包含：

- 后端流程执行语义重写
- 历史旧配置自动迁移
- 流程图节点渲染样式大改

## 方案对比

### 方案 A：仅前端表单化，保存时映射回现有后端字段

为链接配置和每种流程步骤建立专属表单，但提交时仍落到现有接口字段，如 `locator`、`value`、`out_key`、`check_key` 等。

优点：

- 后端接口改动最小
- 可以快速完成“全部表单化”
- 风险集中在前端映射层

缺点：

- 前端需要维护“表单字段 <-> 通用字段”的双向映射
- 后续若继续演进，映射层仍需维护

### 方案 B：前后端一起改为新的结构化协议

前端和后端同时采用新的字段模型，例如流程项直接按类型保存专属字段。

优点：

- 数据语义最清晰
- 长期维护最好

缺点：

- 改动范围过大
- 需要重写接口、存储、执行层

### 方案 C：只改链接配置，不改执行流程

优点：

- 速度最快

缺点：

- 不满足本次需求

## 采用方案

采用方案 A。

原因：

- 能满足“所有都改成表单化”的目标。
- 不需要同时重构后端执行层。
- 可以在前端内部引入清晰的字段配置和共享编辑器，先把体验做好。

## 详细设计

### 一、链接配置表单化

将 `smartLinkConfig.links` 从通用 JSON 编辑改为“链接项列表”编辑器。

每个链接项拆成以下字段：

- `label`：展示名称
- `link`：跳转地址
- `browser_auth_username`：浏览器认证用户名
- `browser_auth_password`：浏览器认证密码
- `userList`：账号列表

账号列表每项字段：

- `user_name`
- `password`

同时把以下文本字段改为可增删行的小表单：

- `show_cookies`
  - 改为“信息提取项列表”
  - 先按“单项一行字符串”建模，避免本轮强行引入未知复杂结构
- `filter_uris`
  - 改为“请求拦截项列表”
  - 每项一条半匹配规则

保存时：

- 链接列表序列化为 `smartLinkConfig.links`
- 信息提取项列表按换行或约定格式回写到 `show_cookies`
- 请求拦截项列表回写到 `filter_uris`

### 二、执行流程项表单化

新增统一的“流程项表单编辑器”，按步骤类型展示不同字段，不再直接向用户暴露通用字段含义。

计划覆盖的步骤类型：

- `text_content`
- `redirect_uri`
- `wait_url`
- `wait`
- `bool_result`
- `bool_exist`
- `click`
- `input`
- `close`
- `no_exist_wait`
- `canvas_image`
- `login_username_password`
- `delete_element`

统一保留的公共字段：

- `name`
- `type`
- `tip`
- `weight`
- `wait_mills`
- `domain_limit`
- `append_to_replace`
- `is_async`
- `is_error_continue`

不同步骤的专属表单含义：

- `text_content`
  - 元素定位
  - 输出键
  - 是否追加到替换列表
- `redirect_uri`
  - 跳转地址
- `wait_url`
  - 等待地址匹配规则
  - 超时等待毫秒
- `wait`
  - 等待毫秒
- `bool_result`
  - 判断键
  - 期望值
- `bool_exist`
  - 元素定位
  - 判断输出键
- `click`
  - 元素定位
- `input`
  - 元素定位
  - 输入内容
  - 输出键
- `close`
  - 无专属字段，仅公共控制字段
- `no_exist_wait`
  - 元素定位
  - 等待毫秒
- `canvas_image`
  - 元素定位
  - 输出键
- `login_username_password`
  - 用户名输入框定位
  - 密码输入框定位
  - 提交按钮定位
  - 用户名输出键
  - 密码输出键
- `delete_element`
  - 删除策略
  - 目标定位或标识

注意：

- 为避免本轮动后端协议，流程项保存时仍映射回 `locator`、`value`、`out_key`、`check_key` 等现有字段。
- 也就是说：界面是专属表单，提交协议仍是当前结构。

### 三、共享组件收敛

当前 [link_process.vue](C:/work/frog/dev_tool_master/web/src/components/smart_link/link_process.vue) 和 [link_flow.vue](C:/work/frog/dev_tool_master/web/src/components/smart_link/link_flow.vue) 各维护了一份流程项编辑表单。

本轮应抽出共享组件，例如：

- `web/src/components/smart_link/ProcessItemEditor.vue`
- `web/src/components/smart_link/LinkConfigEditor.vue`

职责划分：

- `ProcessItemEditor.vue`
  - 负责流程项类型定义
  - 负责步骤专属字段表单
  - 负责表单数据与旧接口字段之间的映射
- `LinkConfigEditor.vue`
  - 负责链接项列表、账号列表、信息提取、请求拦截的可视化表单

原页面仅负责：

- 打开/关闭弹窗
- 拉取数据
- 保存回调

### 四、验证

手工验证至少覆盖：

1. 创建自定义网页时，可以通过小表单新增多个链接与账号。
2. 编辑现有链接时，小表单显示正常。
3. 信息提取、请求拦截可增删改。
4. 13 种执行流程步骤都能通过表单创建。
5. 执行流程列表视图与流程图视图使用同一套编辑器。
6. 保存后再次打开，同一条配置显示一致。

## 风险

- 因为不兼容旧数据，若接口返回历史旧格式配置，可能无法完整回显。
- `login_username_password` 这类复合步骤目前后端仍只有通用字段，前端映射要非常明确，否则容易语义丢失。
- 当前仓库没有现成前端单测基础设施，本轮主要依赖 ESLint 与手工验证。

## 结论

本轮应把“自定义网页”配置体验从“懂 JSON/懂底层字段的人可用”提升到“按步骤和业务字段直接填写”，优先通过前端共享表单组件完成，而不是同时重写后端协议。
