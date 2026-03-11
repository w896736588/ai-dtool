---
name: dtool-api-import-update
description: Use when operating the dtool 接口开发模块 and the task involves querying collections, folders, APIs, environments, importing, updating, moving, running, or batch-managing API definitions.
---

# dtool 接口开发模块技能

## 核心目标

让 AI 在 dtool 的“接口开发”模块里稳定完成查询、选择、创建、更新、移动、运行、批量导入等操作，并且始终使用正确的接口与交互顺序。

## 强制约束

1. 交互必须使用简体中文。
2. AI 在操作 dtool 接口开发模块的所有接口时，必须使用 UTF-8 编码处理请求与响应，避免中文字段、错误信息、描述信息出现乱码。
3. 使用 PowerShell 或其他终端前，必须先切换 UTF-8 编码：

```powershell
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
```

4. 调用接口、拼装 JSON、读取接口返回、整理错误信息时，都要默认按 UTF-8 处理。
5. 读取接口说明时，必须先看 [接口说明](references/dtool-api-endpoints.md)。
6. 未确认集合、文件夹或风险边界前，不得直接写入接口数据。
7. 需要批量覆盖、删除、移动时，必须先提示影响范围。

## 推荐工作流

### 场景 1：只做层级查询

1. 查集合列表时，优先用 `/api/CollectionListBasic`。
2. 按集合查文件夹时，优先用 `/api/CollectionFoldersBasic`。
3. 按文件夹查接口基础信息时，优先用 `/api/FolderApisBasic`。
4. 需要完整树时，才用 `/api/Collections`。
5. 需要某个文件夹完整详情时，才用 `/api/FolderDetail`。
6. 需要若干接口完整明细时，用 `/api/ApisDetailByIds`。

### 场景 2：导入或按 URI 更新接口

1. 先调 `/api/CollectionListBasic` 或 `/api/Collections` 让用户确认集合。
2. 再调 `/api/CollectionFoldersBasic` 让用户确认文件夹，并提供“新建文件夹”选项。
3. 若用户选择新建文件夹，调 `/api/CreateDir`。
4. 获取目标文件夹下已有接口时，优先调 `/api/FolderApisBasic`；若还需要更多字段，再调 `/api/FolderDetail`。
5. 按 URI 决定创建或更新：
   - 命中已有接口：调 `/api/CreateApi`，带 `id`
   - 未命中：调 `/api/CreateApi`，不带 `id`
6. 汇总创建、更新、失败结果。

### 场景 3：批量导入整个文件夹

1. 先确认用户接受“同名文件夹覆盖式更新”。
2. 明确提示：同名 folder 导入时，会先清空该 folder 下旧接口。
3. 使用 `/api/ApiBatchImport`，请求为 `multipart/form-data`。

### 场景 4：环境、调试与运行

1. 查询集合环境：`/api/CollectionEnvs`
2. 新建或更新环境：`/api/CreateCollectionEnv`
3. 查询环境变量：`/api/CollectionEnvItems`
4. 新建或更新环境变量：`/api/CreateCollectionEnvItem`
5. 运行接口：`/api/ApiRun`
6. 生成代码：`/api/ApiCode`
7. 提取 JSON 路径：`/api/ApiTakeJsonResult`
8. 下移权重：`/api/ApiWeightDown`

### 场景 5：结构调整

1. 删除集合：`/api/DeleteCollection`
2. 删除文件夹：`/api/DeleteDir`
3. 删除接口：`/api/DeleteApi`
4. 移动接口到其他文件夹：`/api/ApiMove`

## URI 匹配规则

1. 默认使用“规范化 URI + method”匹配。
2. 规范化步骤：
   - 去除前后空格
   - 协议与域名大小写忽略
   - 去掉末尾 `/`
3. 如果用户明确要求“仅按 URI 匹配，不区分 method”，按用户要求执行。
4. 如果同一文件夹内命中多个候选，必须暂停并让用户确认接口 ID。

## 何时优先用轻量接口

- 只做列表选择：优先用 `CollectionListBasic`、`CollectionFoldersBasic`、`FolderApisBasic`
- 只查若干接口详情：优先用 `ApisDetailByIds`
- 只有在确实需要完整树、完整目录、完整接口内容时，才用 `Collections`、`FolderDetail`

这样可以减少无关字段，提高交互稳定性。

## 失败反馈要求

接口调用失败时，必须返回：

- 目标对象名称
- 目标集合或文件夹
- 接口名称或 URI
- 失败接口名
- 后端返回的错误信息

## 可选脚本

需要自动执行导入/更新时，可优先查看：

- `scripts/sync_api_by_uri.py`

使用前仍要先确认：

- 目标集合
- 目标文件夹
- 是否允许自动新建文件夹
- 是否允许覆盖更新
