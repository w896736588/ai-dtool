# AI 设置页模型 URI 设计

## 目标

将 AI 配置拆成更清晰的两层：

- 服务商只保存基础域名，例如 `https://api.openai.com`
- 模型保存具体请求 `uri`，例如 `/v1/chat/completions`
- 模型增加类型区分，支持 `llm` 与 `embedding`

最终调用地址统一为：`provider.base_url + model.uri`。

## 现状

- `tbl_ai_provider.base_url` 目前被当作完整请求地址使用。
- `tbl_ai_model` 目前只有 `provider_id`、`name`、`model`。
- 设置页模型配置只有展示名和模型标识。
- 信息抓取与变量模块都存在“如果地址不含 `/chat/completions` 就自动补齐”的旧逻辑。

## 设计

### 数据层

- 保留 `tbl_ai_provider.base_url` 字段名，但语义调整为“基础域名”。
- 为 `tbl_ai_model` 新增：
  - `uri`：模型请求路径，必须是相对路径
  - `model_type`：模型类型，枚举值 `llm` / `embedding`

### 兼容迁移

- 历史模型记录补充 `model_type='llm'`
- 历史模型记录补充 `uri='/v1/chat/completions'`
- 历史服务商 `base_url` 如果保存的是完整 `chat/completions` 地址，则迁移为基础域名

### 后端接口

- `SetAiModelList` 返回 `uri`、`model_type`
- `SetAiModelAdd` 校验并规范化：
  - `provider_id` 必填
  - `model` 必填
  - `uri` 必填，自动补前导 `/`
  - `model_type` 默认 `llm`
- AI 调用地址改为 `base_url + uri`

### 前端设置页

- 服务商配置页：
  - `Base URL` 改成“基础域名”
  - 示例改为 `https://api.openai.com`
- 模型配置页：
  - 表格新增“模型类型”“URI”
  - 编辑弹窗新增“模型类型”“URI”
  - 模型类型提供 `LLM`、`嵌入模型`

## 风险点

- 旧数据可能混有不同格式的 `base_url`，迁移需要尽量保守，只剥离明确的路径部分。
- 变量模块和信息抓取模块都依赖 AI 配置，调用链必须一起改，不能只改设置页。

## 验证

- 后端测试覆盖模型新增默认值与 `uri` 规范化
- 后端测试覆盖旧完整地址迁移后的请求地址拼接
- 前端手工验证服务商与模型的新增、编辑、列表展示
