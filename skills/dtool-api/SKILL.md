---
name: dtool-api
description: Use when operating the dtool 接口开发模块 and the task involves querying collections, folders, APIs, environments, importing, updating, moving, running, or batch-managing API definitions.
---

# dtool 接口开发模块技能

## 核心目标

让 AI 在 dtool 的“接口开发”模块里稳定完成查询、选择、创建、更新、移动、运行、批量导入等操作，并且始终使用正确的接口与交互顺序。

## 强制约束

1. 交互必须使用简体中文。
2. AI 在操作 dtool 接口开发模块的所有接口时，必须使用 UTF-8 编码处理请求与响应，避免中文字段、错误信息、描述信息出现乱码。
3. 使用 PowerShell 或其他终端前，必须先切换 UTF-8 编码.
4. 改动或创建文件夹或接口时，统一使用 Python 脚本发送请求，避免 bash 环境的编码问题
5. 调用 dtool 接口前，必须向用户确认请求地址，不得假设默认地址。
6. 调用 dtool 接口前，必须向用户确认 Header 头 `Token` 的具体值；所有请求都必须携带 `Token`。

```powershell
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
```

7. 调用接口、拼装 JSON、读取接口返回、整理错误信息时，都要默认按 UTF-8 处理。
8. 读取接口说明时，必须先看 [接口说明](references/dtool-api-endpoints.md)。
9. 未确认集合、文件夹或风险边界前，不得直接写入接口数据。
10. 需要批量覆盖、删除、移动时，必须先提示影响范围。

## 推荐工作流

### 场景 0：先检查当前分支改动的接口，再补接口文档

1. 先在仓库根目录运行变更文件脚本，筛出当前分支相对基分支的改动文件：
   - Windows / PowerShell：`powershell -File skills/dtool-api/scripts/show-branch-diff.ps1 [base-branch]`
   - Linux / macOS / bash：`bash skills/dtool-api/scripts/show-branch-diff.sh [base-branch]`
2. 重点关注接口相关文件，按“接口定义与入口优先”原则筛选，不要绑定当前仓库路径：
   - `.php` 文件，尤其是控制器、路由、请求类、资源类、服务入口等
   - `.go` 文件，尤其是 controller、router、handler、service entry、request binding、response struct 等
   - 用户明确指定的接口实现文件、路由文件、控制器文件
   - 只有当前端文件里直接定义了接口地址、请求参数或返回字段映射时，才把它们作为辅助参考，而不是主依据
3. 对每个疑似有接口变更的文件，再查看单文件 diff：
   - Windows / PowerShell：`powershell -File skills/dtool-api/scripts/show-file-diff.ps1 <file-path> [base-branch]`
   - Linux / macOS / bash：`bash skills/dtool-api/scripts/show-file-diff.sh <file-path> [base-branch]`
4. 从 diff 中确认以下信息后，再开始写接口文档：
   - 是否新增接口、修改 URI、修改 method、修改请求参数、修改响应字段
   - 后端控制器实际如何接收参数，以决定 `content_type`、`query_params`、`body_form`、`body_json`、`body_raw`
   - 返回结构里哪些字段需要写入 `take_result`
5. 如果变更里同时包含非接口文件，优先聚焦接口语义变更，不要把纯样式、dist、构建产物误当成接口改动
6. 在没有确认 diff 之前，不得直接假设接口文档需要新增或修改

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

## 接口创建 / 导入的强制约束

创建或更新接口（`/api/CreateApi`、`/api/ApiBatchImport`）时，必须严格遵守以下规则：

### 1. 必须完整定义请求参数

- **禁止只定义接口而不设置请求参数**。每个接口都必须根据后端控制器的实际代码，完整填写 `query_params`、`body_form` 或 `body_json`。
- 如果接口有 URL 查询参数，必须在 `query_params` 中逐一定义。
- 如果是 POST 请求，必须根据后端控制器实际接收方式，填写 `body_form` 或 `body_json`。
- 即便接口暂时没有参数，也必须传空数组 `[]` 或空字符串 `""`，不得省略字段。

### 2. 固定值、常量、枚举值必须写清含义

请求参数如果存在常量、固定值、枚举值或布尔开关，必须在参数备注中明确列出每个值和含义。

- `query_params` 和 `body_form` 参数必须写入每项的 `description` 字段。
- 如果只有一个固定值，也要说明“固定传某值”以及该值代表什么。
- 如果参数来自代码常量、状态码、类型字段、开关字段，必须回到源码确认取值，不得只写“类型”“状态”“是否启用”这类笼统描述。
- `body_json` 中的字段如果有固定值或枚举值，必须在接口 `desc` 中补充字段取值说明；能拆成表单参数的场景，优先写入对应参数的 `description`。

示例：
```json
[
  {"field": "client_type", "type": "string", "value": "pc", "description": "客户端类型：pc=PC端，h5=移动H5，mini=小程序"},
  {"field": "status", "type": "integer", "value": "1", "description": "状态：0=禁用，1=启用，2=冻结"},
  {"field": "version", "type": "string", "value": "v1", "description": "接口版本，固定传 v1，表示第一版协议"}
]
```

### 3. 参数类型必须使用规范名称

`query_params` 和 `body_form` 中每项的 `type` 字段只接受以下值：

| 规范值 | 含义 | 禁止使用的旧值 |
|---|---|---|
| `string` | 字符串 | - |
| `integer` | 整数 | **禁止使用 `int`**，后端会直接拒绝 |
| `float` | 浮点数 | - |
| `boolean` | 布尔值 | 也接受 `bool`，推荐统一用 `boolean` |
| `file` | 文件上传 | - |

**如果 type 写成 `int`，后端会报错并拒绝写入。`bool` 和 `boolean` 均可正常使用，推荐统一用 `boolean`。**

### 4. POST 请求的 content_type 必须根据后端控制器代码判断

不得默认将所有 POST 请求的 `content_type` 设为 `application/json`。必须根据后端 Go 控制器代码实际接收参数的方式来决定：

| 后端控制器写法 | 对应的 content_type | 请求数据字段 |
|---|---|---|
| `gsgin.GinPostBody(c, &dataMap)` 或 `c.BindJSON()` | `application/json` | `body_json` |
| `c.PostForm("key")` 或 `c.DefaultPostForm()` | `application/x-www-form-urlencoded` | `body_form` |
| `c.MultipartForm()` 或文件上传场景 | `multipart/form-data` | `body_form` |
| 纯文本/二进制请求体 | `text/plain` 或 `raw` | `body_raw` |

**判断步骤**：
1. 先阅读后端控制器的源码，确认它如何读取请求参数
2. 根据实际代码确定 `content_type` 和对应的请求数据字段
3. 如果无法确定，优先询问用户

### 5. 必须设置结果字段备注（take_result）

每个接口都必须设置 `take_result`（结果字段备注），用于描述接口返回结果中各字段的含义。

返回结果字段描述、字段含义、示例等备注内容必须写入 `take_result` 数组结构。

`take_result` 格式为 JSON 数组，每项包含：
- `key`：返回字段路径（必填，如 `code`、`data.token`）
- `type`：字段类型（必填，如 `string`、`number`、`boolean`、`object`、`array`）
- `desc`：字段含义描述（必填）

示例：
```json
[
  {"key": "code", "type": "number", "desc": "状态码，0表示成功"},
  {"key": "data.token", "type": "string", "desc": "用户令牌"},
  {"key": "data.user_id", "type": "number", "desc": "用户ID"}
]
```

**禁止在生成接口时留空 take_result**，至少要定义返回结构中的核心字段及其中文描述。

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
- `scripts/show-branch-diff.ps1`
- `scripts/show-file-diff.ps1`
- `scripts/show-branch-diff.sh`
- `scripts/show-file-diff.sh`
- `scripts/test_show_branch_diff.ps1`

使用前仍要先确认：

- 目标集合
- 目标文件夹
- 是否允许自动新建文件夹
- 是否允许覆盖更新
