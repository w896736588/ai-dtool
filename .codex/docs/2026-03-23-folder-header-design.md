# 文件夹默认 Header 设计

## 背景

接口开发模块当前仅支持接口级别的 `headers` 配置，文件夹 `tbl_api_dir` 未提供默认请求头能力。需求要求为文件夹增加 header 配置，并在实际请求时先加载所属文件夹 header，再用接口自身 header 覆盖同名项。

## 目标

1. 文件夹支持保存默认请求头。
2. 文件夹 header 优先级低于接口 header。
3. 不改写已有接口存量数据，仅在运行态进行合并。
4. 前端可在文件夹详情页直接编辑默认 header。

## 方案

### 数据存储

在 `tbl_api_dir` 增加 `headers` 字段，类型为 `text`，默认值为 `'{}'`。

原因：

- 与 `tbl_api.headers` 保持同构，复用现有 JSON 解析和校验逻辑。
- 改动范围小，不需要新增关联表。
- 目录详情、目录列表和运行态都可以直接透传该字段。

### 后端读取与保存

`CreateDir` 增加 `headers` 字段读写与格式校验，空值统一落为 `'{}'`。

`CollectionFoldersBasic`、`Collections`、`FolderDetail` 返回目录详情时透传 `headers`，保证前端刷新和树节点同步后仍能保留目录默认请求头。

### 运行时合并规则

接口执行时新增目录 header 合并逻辑：

1. 读取接口所属目录 `tbl_api_dir.headers`
2. 解析目录 header 为 map
3. 再解析接口自身 `tbl_api.headers`
4. 以目录 header 为基础，接口 header 逐项覆盖

这样可以满足：

- 目录提供默认公共 header
- 接口可覆盖同名默认值
- 接口未声明的 header 继续继承目录值

### 前端交互

在文件夹基本信息区域增加“默认请求头”编辑器，复用现有 `HeadersValueEditor` 组件，避免重复实现键值编辑交互。

保存时调用 `CreateDir`，携带：

- `id`
- `name`
- `collection_id`
- `headers`

前端树节点同步时需要保留 `headers` 字段，避免保存后切换 tab 丢失状态。

## 风险与边界

1. 只支持一层文件夹，不涉及父子目录继承链。
2. 仅影响接口执行态，不调整导出文档逻辑。
3. 导入逻辑本次不要求支持目录级 header 导入，除非现有结构已自然兼容。

## 验证

1. 后端单测验证目录与接口 header 的合并顺序。
2. 后端单测验证目录基础信息构建时包含 `headers`。
3. 手工验证文件夹详情保存后重新打开仍可看到默认 header。
