# dtool 接口开发模块 API 说明

## 使用前置要求

1. 服务请求地址必须由用户明确提供，不得假设默认地址。
2. 路由前缀：`/api`
3. 除 `ApiBatchImport` 外，均为 `POST + application/json`
4. AI 在调用这些接口时，必须使用 UTF-8 编码处理请求与响应，尤其是 `name`、`desc`、错误信息、目录名、集合名等中文字段。
5. 使用 PowerShell 或其他终端前，必须先切换 UTF-8 编码：

```powershell
[Console]::OutputEncoding = [System.Text.Encoding]::UTF8
```

6. 如果通过脚本、终端或手工请求这些接口，默认都要按 UTF-8 发送请求体，并按 UTF-8 解读接口返回内容。

## 一、集合相关

### 1. 查询所有集合树：`/api/Collections`

用途：

- 返回集合 -> 文件夹 -> 接口的完整树

请求：

```json
{}
```

关键返回：

- `data.list[]`
- 集合节点：`id`、`name`、`type=collection`、`children`
- 文件夹节点：`id`、`name`、`type=folder`、`children`
- 接口节点：`id`、`name`、`method`、`url`、`type=api`

### 2. 查询所有集合基础信息：`/api/CollectionListBasic`

用途：

- 仅用于集合选择，不返回文件夹和接口

请求：

```json
{}
```

关键返回：

- `data.list[]`
- 每项字段：`id`、`name`、`create_time`、`update_time`、`type`、`uniqueid`

### 3. 新建或更新集合：`/api/CreateCollection`

创建：

```json
{
  "name": "用户中心"
}
```

更新：

```json
{
  "id": 1,
  "name": "用户中心-新版"
}
```

### 4. 删除集合：`/api/DeleteCollection`

请求：

```json
{
  "id": 1
}
```

## 二、文件夹相关

### 1. 按集合查询文件夹基础信息：`/api/CollectionFoldersBasic`

用途：

- 仅用于文件夹选择

请求：

```json
{
  "collection_id": 1
}
```

关键返回：

- `data.list[]`
- 每项字段：`id`、`collection_id`、`name`、`create_time`、`update_time`、`type`、`uniqueid`

### 2. 新建或更新文件夹：`/api/CreateDir`

创建：

```json
{
  "name": "用户中心",
  "collection_id": 1
}
```

更新：

```json
{
  "id": 12,
  "name": "用户中心-新版",
  "collection_id": 1
}
```

### 3. 删除文件夹：`/api/DeleteDir`

请求：

```json
{
  "id": 12
}
```

### 4. 查询文件夹详情：`/api/FolderDetail`

用途：

- 查询目录完整信息及目录下接口详情

请求：

```json
{
  "dir_id": 12
}
```

关键返回：

- `data.dir`
- `data.dir.children[]` 为接口列表

### 5. 通过文件夹 ID 获取接口文档（Markdown）：`/api/FolderApisMarkdown`

用途：

- 通过文件夹 ID 获取该文件夹下所有接口的 Markdown 格式文档，格式与前端"复制所有接口(Markdown)"按钮完全一致

请求：

```json
{
  "folder_id": 12
}
```

关键返回：

- `data.markdown` — 完整的 Markdown 字符串，包含文件夹下所有接口的文档

## 三、接口列表与详情相关

### 1. 按集合和文件夹查询接口：`/api/Apis`

请求：

```json
{
  "collection_id": 1,
  "dir_id": 12
}
```

关键返回：

- `data.list[]`

### 2. 按文件夹查询接口基础信息：`/api/FolderApisBasic`

用途：

- 仅返回基础信息，不返回请求和响应明细字段

请求：

```json
{
  "folder_id": 12
}
```

关键返回：

- `data.list[]`
- 每项字段：`id`、`folder_id`、`collection_id`、`name`、`method`、`url`、`desc`、`env_id`、`weight`、`create_time`、`update_time`、`type`、`uniqueid`

### 3. 按若干接口 ID 查询接口明细：`/api/ApisDetailByIds`

请求支持两种格式：

```json
{
  "ids": [101, 102, 103]
}
```

```json
{
  "ids": "101,102,103"
}
```

关键返回：

- `data.list[]`
- 返回 `tbl_api` 的完整字段，并附带 `type`、`uniqueid`

### 4. 新建或更新接口：`/api/CreateApi`

逻辑：

- `id` 不存在或为 0：创建
- `id` 存在：更新

#### 参数说明

| 字段 | 类型 | 说明                                             |
|---|---|------------------------------------------------|
| `id` | int | 接口 ID，更新时必传                                    |
| `folder_id` | int | 目标文件夹 ID                                       |
| `collection_id` | int | 目标集合 ID                                        |
| `name` | string | 接口名称                                           |
| `method` | string | 请求方法：GET / POST                                |
| `url` | string | 请求 URL，可含环境变量如 `$Url$/v1/login`                |
| `protocol` | string | 协议：http / https                                |
| `desc` | string | 接口描述                                           |
| `headers` | object | 请求头，键值对 `{"Content-Type":"application/json"}`  |
| `query_params` | array | URL 查询参数数组，见下方字段格式                             |
| `content_type` | string | 请求体类型，**必须根据后端控制器实际代码判断**，见下方对照表               |
| `body_form` | array | 表单参数数组（用于 form-urlencoded / multipart），见下方字段格式 |
| `body_json` | string | JSON 请求体字符串（用于 application/json）               |
| `body_raw` | string | 原始请求体（用于 text/plain / raw）                     |
| `env_id` | int | 接口绑定的环境 ID。绑定后，接口中的 `$变量Key$` 会从该环境下的变量项取值 |
| `take_result` | array | 结果字段备注，**必须填写**，用于写入返回字段描述、字段含义、示例等备注结构，见下方格式  |
| `take_result_desc` | string | 不需要处理                                          |

#### 环境变量在接口中的使用方式

接口可以在以下位置引用环境变量：

- `url`，例如：`$Url$/v1/login`
- `headers`，例如：`{"Cookie":"$Cookie$","Token":"$Token$"}`
- `query_params[].value`
- `body_form[].value`
- `body_json`
- `body_raw`

引用格式统一为：

```text
$变量Key$
```

例如环境变量项的 `key` 为 `Cookie`、`Token`、`Url`，则接口里应写成：

```json
{
  "url": "$Url$/api/user/info",
  "headers": {
    "Cookie": "$Cookie$",
    "Token": "$Token$"
  }
}
```

#### env_id 与变量引用的关系

1. 先在集合下创建环境，例如“测试环境”“预发环境”
2. 再在该环境下创建若干环境变量项，例如 `Url`、`Cookie`、`Token`
3. 创建或更新接口时，把接口的 `env_id` 指向目标环境
4. 接口中的 `url`、`headers`、请求参数里再使用 `$变量Key$` 引用实际值

如果接口里写了 `$Cookie$`、`$Token$`，但没有绑定正确的 `env_id`，或目标环境下不存在对应 `key`，运行接口时就无法得到正确的请求值。

#### 登录态相关请求头的推荐写法

对于登录后抓取得到的请求头，尤其是 `Cookie`、`Token`、`Authorization`、`X-Token` 等认证字段，不要把真实值直接固化在接口定义里，而应优先：

1. 在集合环境下创建同名或语义清晰的变量项
2. 将真实值写入变量项的 `value`
3. 在接口 `headers` 中改为 `$变量Key$` 引用

示例：

```json
{
  "env_id": 5,
  "headers": {
    "Cookie": "$Cookie$",
    "Token": "$Token$",
    "Authorization": "$Authorization$"
  }
}
```

#### 返回结果字段写入规则

- 返回字段描述、字段含义、示例等备注内容必须写入 `take_result` 数组结构。
- 生成或更新接口时，返回字段描述只能写入 `take_result`。
- `take_result_desc` 不需要处理。

#### query_params / body_form 中每项的字段格式

```json
{
  "field": "client_type",
  "type": "string",
  "value": "pc",
  "description": "客户端类型：pc=PC端，h5=移动H5，mini=小程序"
}
```

#### 固定值、常量、枚举值备注规则

请求参数如果存在固定值、常量、枚举值或布尔开关，必须在备注中明确每个值和含义。

- `query_params` 和 `body_form`：写入参数项的 `description` 字段。
- 单一固定值：也要说明“固定传某值”以及该值代表什么。
- 枚举/状态/类型/开关：必须列出每个允许值的含义，不能只写“状态”“类型”“是否启用”。
- 取值来自源码常量时，必须以源码为准；无法确认时先询问用户。
- `body_json` 字段没有独立的参数备注结构时，必须在接口 `desc` 中补充字段取值说明。

示例：
```json
[
  {"field": "client_type", "type": "string", "value": "pc", "description": "客户端类型：pc=PC端，h5=移动H5，mini=小程序"},
  {"field": "status", "type": "integer", "value": "1", "description": "状态：0=禁用，1=启用，2=冻结"},
  {"field": "enabled", "type": "boolean", "value": "true", "description": "是否启用：true=启用，false=停用"},
  {"field": "version", "type": "string", "value": "v1", "description": "接口版本，固定传 v1，表示第一版协议"}
]
```

**type 字段只接受以下值（严禁使用其他值）：**

| 规范值 | 含义 | 错误写法（会被后端拒绝） |
|---|---|---|
| `string` | 字符串 | - |
| **`integer`** | 整数 | ~~`int`~~ (后端会报错) |
| `float` | 浮点数 | - |
| **`boolean`** | 布尔值 | 也接受 `bool`，推荐统一用 `boolean` |
| `file` | 文件上传 | - |

> **重要**：type 写成 `int` 会导致后端报错 `"type 仅支持 integer，不支持 int"`，接口无法创建。`bool` 和 `boolean` 均可正常使用，推荐统一用 `boolean`。

#### content_type 判断规则（必须根据后端 Go 控制器代码决定）

**不得默认所有 POST 都是 `application/json`**，必须先阅读后端控制器源码再决定：

| 后端控制器代码写法 | content_type 值 | 请求体字段 |
|---|---|---|
| `gsgin.GinPostBody(c, &dataMap)` 或 `c.BindJSON()` | `application/json` | `body_json` |
| `c.PostForm("key")` 或 `c.DefaultPostForm("key")` | `application/x-www-form-urlencoded` | `body_form` |
| `c.MultipartForm()` 或涉及文件上传 | `multipart/form-data` | `body_form` |
| 纯文本/二进制请求体 | `text/plain` 或 `raw` | `body_raw` |
| GET 请求 | 不设置或留空 | 无 |

#### take_result 格式（必须填写，描述接口返回字段含义）

```json
[
  {
    "key": "code",
    "type": "number",
    "desc": "状态码，0表示成功"
  },
  {
    "key": "data.token",
    "type": "string",
    "desc": "用户认证令牌"
  },
  {
    "key": "data.user_id",
    "type": "number",
    "desc": "用户唯一ID"
  }
]
```

> **禁止留空 take_result**，至少要描述返回结构中的核心字段。

#### 创建示例（application/json 类型）

```json
{
  "folder_id": 12,
  "collection_id": 1,
  "name": "用户登录",
  "method": "POST",
  "url": "$Url$/v1/login",
  "protocol": "https",
  "desc": "用户登录接口，返回认证令牌",
  "headers": {
    "Content-Type": "application/json"
  },
  "query_params": [],
  "content_type": "application/json",
  "body_form": [],
  "body_json": "{\"username\":\"demo\",\"password\":\"123456\"}",
  "take_result": [
    {"key": "code", "type": "number", "desc": "状态码，0表示成功"},
    {"key": "msg", "type": "string", "desc": "提示信息"},
    {"key": "data.token", "type": "string", "desc": "认证令牌"}
  ]
}
```

#### 创建示例（form-urlencoded 类型）

```json
{
  "folder_id": 12,
  "collection_id": 1,
  "name": "提交表单",
  "method": "POST",
  "url": "$Url$/v1/submit",
  "protocol": "https",
  "desc": "提交表单数据",
  "headers": {
    "Content-Type": "application/x-www-form-urlencoded"
  },
  "query_params": [],
  "content_type": "application/x-www-form-urlencoded",
  "body_form": [
    {"field": "title", "type": "string", "value": "测试标题", "description": "标题"},
    {"field": "count", "type": "integer", "value": "10", "description": "数量"},
    {"field": "enabled", "type": "boolean", "value": "true", "description": "是否启用：true=启用，false=停用"}
  ],
  "body_json": "",
  "take_result": [
    {"key": "code", "type": "number", "desc": "状态码"},
    {"key": "data.id", "type": "number", "desc": "提交结果ID"}
  ]
}
```

#### 更新示例

```json
{
  "id": 201,
  "folder_id": 12,
  "collection_id": 1,
  "name": "用户登录",
  "method": "POST",
  "url": "$Url$/v1/login",
  "protocol": "https",
  "desc": "用户登录接口，返回认证令牌",
  "headers": {
    "Content-Type": "application/json"
  },
  "query_params": [],
  "content_type": "application/json",
  "body_form": [],
  "body_json": "{\"username\":\"demo\",\"password\":\"new-password\"}",
  "take_result": [
    {"key": "code", "type": "number", "desc": "状态码，0表示成功"},
    {"key": "msg", "type": "string", "desc": "提示信息"},
    {"key": "data.token", "type": "string", "desc": "认证令牌"}
  ]
}
```

### 5. 删除接口：`/api/DeleteApi`

请求：

```json
{
  "id": 201
}
```

### 6. 移动接口到其他文件夹：`/api/ApiMove`

请求：

```json
{
  "api_id": 201,
  "folder_id": 20
}
```

说明：

- 目标文件夹必须存在
- 目标文件夹必须与接口属于同一个集合

## 四、环境相关

### 1. 查询集合环境列表：`/api/CollectionEnvs`

请求：

```json
{
  "collection_id": 1
}
```

关键返回：

- `data.list[]`
- 每个环境下会附带 `variables`

### 2. 新建或更新环境：`/api/CreateCollectionEnv`

创建：

```json
{
  "name": "测试环境",
  "collection_id": 1,
  "desc": "测试"
}
```

更新：

```json
{
  "id": 5,
  "name": "测试环境-新版",
  "collection_id": 1,
  "desc": "测试"
}
```

### 3. 查询环境变量列表：`/api/CollectionEnvItems`

请求：

```json
{
  "collection_id": 1,
  "env_id": 5
}
```

### 4. 新建或更新环境变量：`/api/CreateCollectionEnvItem`

创建：

```json
{
  "collection_id": 1,
  "env_id": 5,
  "desc": "网关地址",
  "key": "Url",
  "value": "https://example.com"
}
```

更新：

```json
{
  "id": 21,
  "collection_id": 1,
  "env_id": 5,
  "desc": "网关地址",
  "key": "Url",
  "value": "https://example.com"
}
```

### 5. 环境变量如何应用到接口中

典型步骤：

1. 先调用 `/api/CollectionEnvs` 确认集合下要使用的环境，拿到 `env_id`
2. 再调用 `/api/CollectionEnvItems` 查看该环境已有变量项
3. 如果缺少所需变量，调用 `/api/CreateCollectionEnvItem` 创建，例如 `Url`、`Cookie`、`Token`
4. 创建或更新接口时，调用 `/api/CreateApi`，把接口的 `env_id` 设为该环境 ID
5. 将接口中的真实值改成 `$变量Key$` 形式，例如：
   - `url`: `$Url$/api/order/list`
   - `headers.Cookie`: `$Cookie$`
   - `headers.Token`: `$Token$`

接口示例：

```json
{
  "folder_id": 12,
  "collection_id": 1,
  "name": "订单列表",
  "method": "GET",
  "url": "$Url$/api/order/list",
  "protocol": "https",
  "desc": "查询订单列表",
  "headers": {
    "Cookie": "$Cookie$",
    "Token": "$Token$"
  },
  "query_params": [],
  "content_type": "",
  "body_form": [],
  "body_json": "",
  "body_raw": "",
  "env_id": 5,
  "take_result": [
    {"key": "code", "type": "number", "desc": "状态码"},
    {"key": "data.list", "type": "array", "desc": "订单列表"}
  ]
}
```

### 6. 使用登录态请求头更新环境变量的推荐流程

当 AI 通过浏览器或其他方式拿到登录后的请求头时，推荐按以下顺序处理：

1. 确认目标接口所属集合和目标环境
2. 对每个需要复用的请求头，优先查找是否已有对应环境变量项
3. 已存在则调用 `/api/CreateCollectionEnvItem` 并带 `id` 更新 `value`
4. 不存在则新建环境变量项，`key` 建议与请求头名称一致或保持稳定命名
5. 将所有相关接口的 `headers` 改成 `$变量Key$`
6. 确保这些接口都绑定了正确的 `env_id`

常见登录态请求头示例：

- `Cookie` -> `$Cookie$`
- `Token` -> `$Token$`
- `Authorization` -> `$Authorization$`
- `X-Token` -> `$X-Token$`

**重要**：登录态值应写入集合环境变量，不要直接把真实 Cookie 或 Token 固化在接口 `headers` 中。

## 五、运行、调试与辅助能力

### 1. 运行接口：`/api/ApiRun`

请求：

```json
{
  "id": 201
}
```

参数说明：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int | 要运行的接口 ID（必填） |

完整响应示例：

```json
{
  "code": 0,
  "msg": "",
  "data": {
    "url": "https://example.com/v1/login",
    "status_code": 200,
    "errmsg": "",
    "result": "{\"code\":0,\"data\":{\"token\":\"abc123\",\"user_id\":1001}}",
    "status": "success",
    "millisecond": 156,
    "request_headers": {
      "Content-Type": "application/json",
      "Token": "***"
    },
    "response_headers": {
      "Content-Type": "application/json; charset=utf-8",
      "Date": "Tue, 29 Apr 2026 08:00:00 GMT"
    },
    "body_forms": [],
    "body_raw": "{\"username\":\"demo\",\"password\":\"123456\"}",
    "response_take": [
      {
        "description": "状态码，0表示成功",
        "item_key": "code",
        "value": "0",
        "take_value": "0"
      },
      {
        "description": "认证令牌",
        "item_key": "data.token",
        "value": "abc123",
        "take_value": "abc123"
      }
    ],
    "request_time": "2026-04-29 16:00:00"
  }
}
```

响应字段说明：

| 字段 | 类型 | 说明 |
|---|---|---|
| `url` | string | 实际请求的完整 URL（环境变量已替换） |
| `status_code` | int | HTTP 状态码；未发出请求时为 0 |
| `errmsg` | string | 请求错误描述（成功时为空） |
| `result` | string | 接口返回的原始响应体（JSON 字符串） |
| `status` | string | 请求状态（success / error） |
| `millisecond` | int | 请求耗时（毫秒） |
| `request_headers` | object | 实际发送的请求头 |
| `response_headers` | object | 服务端返回的响应头 |
| `body_forms` | array | 提交的 Form 参数（POST form 时） |
| `body_raw` | string | 提交的原始请求体（JSON/raw 时） |
| `response_take` | array | 按 take_result 配置自动提取的字段值 |
| `request_time` | string | 发起请求的时间 |

`response_take` 每项结构：

| 字段 | 类型 | 说明 |
|---|---|---|
| `description` | string | 字段描述（来自 take_result 的 desc） |
| `item_key` | string | 字段路径（如 `data.token`） |
| `value` | string | 实际值 |
| `take_value` | string | 提取后的值 |

> **注意**：
> - 运行接口会实际发送 HTTP 请求，对写接口（POST/PUT/DELETE）请确认不会影响生产数据
> - 接口自身未配置 `env_id` 时，会自动继承所属文件夹的 `env_id`
> - 运行结果会自动保存到接口的 `last_result` 字段

### 2. 生成代码：`/api/ApiCode`

请求：

```json
{
  "id": 201,
  "code_type": "curl bash(chrome)"
}
```

参数说明：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int | 接口 ID（必填） |
| `code_type` | string | 代码类型，如 `"curl bash(chrome)"`（必填） |

完整响应示例：

```json
{
  "code": 0,
  "msg": "",
  "data": {
    "code": "curl 'https://example.com/v1/login' \\\n  -H 'Content-Type: application/json' \\\n  --data-raw '{\"username\":\"demo\",\"password\":\"123456\"}'"
  }
}
```

### 3. 下移接口权重：`/api/ApiWeightDown`

请求：

```json
{
  "id": 201
}
```

### 4. 提取 JSON 响应路径：`/api/ApiTakeJsonResult`

请求：

```json
{
  "id": 201,
  "json": "{\"code\":0,\"data\":{\"token\":\"abc\",\"user_id\":1001}}"
}
```

参数说明：

| 字段 | 类型 | 说明 |
|---|---|---|
| `id` | int | 接口 ID（必填），用于匹配已有的 take_result 描述 |
| `json` | string | JSON 字符串（必填），会自动提取所有叶子节点路径 |

完整响应示例：

```json
{
  "code": 0,
  "msg": "",
  "data": [
    {"key": "code", "type": "number", "desc": "状态码，0表示成功"},
    {"key": "data.token", "type": "string", "desc": "认证令牌"},
    {"key": "data.user_id", "type": "number", "desc": ""}
  ]
}
```

> **用途**：分析接口返回的 JSON 结构，提取所有可用的字段路径，辅助完善 `take_result` 配置。已配置 `take_result` 的字段会自动带上 `desc`。

### 5. 批量导入接口：`/api/ApiBatchImport`

请求方式：

- `POST + multipart/form-data`

表单字段：

- `collection_id`
- `json`

JSON 结构：

```json
{
  "collection_id": 1,
  "items": [
    {
      "type": "folder",
      "name": "用户中心",
      "children": [
        {
          "type": "api",
          "name": "用户登录",
          "method": "POST",
          "url": "$Url$/v1/login",
          "protocol": "https",
          "desc": "登录",
          "headers": {
            "Content-Type": "application/json"
          },
          "query_params": [],
          "content_type": "application/json",
          "body_form": [],
          "body_json": "{\"username\":\"demo\",\"password\":\"123456\"}",
          "take_result": [
            {"key": "code", "type": "number", "desc": "状态码，0表示成功"},
            {"key": "data.token", "type": "string", "desc": "认证令牌"}
          ]
        }
      ]
    }
  ]
}
```

> **批量导入同样需要遵守上述所有约束**：type 只能用 `integer`（不能用 `int`），推荐统一用 `boolean`（`bool` 也可），必须根据后端代码判断 content_type，必须填写 take_result。

注意：

- 根节点只允许 `folder`
- `folder.children` 只允许 `api`
- 同名 folder 导入时，会先清空该 folder 下旧接口再导入

## 六、导入或更新的决策建议

### 1. 按 URI 决定更新还是创建

推荐流程：

1. 调 `/api/CollectionListBasic` 选集合
2. 调 `/api/CollectionFoldersBasic` 选文件夹
3. 调 `/api/FolderApisBasic` 获取已有接口基础信息
4. 规范化 URI 后比对
5. 命中则走 `/api/CreateApi` 更新
6. 未命中则走 `/api/CreateApi` 创建

### 2. 何时用批量导入

当用户给的是“整文件夹 + 多个接口”的结构，且接受覆盖更新风险时，优先考虑 `/api/ApiBatchImport`。

## 七、cURL 示例

### 1. 查询所有集合基础信息

```bash
curl -X POST "http://localhost:17170/api/CollectionListBasic" \
  -H "Content-Type: application/json" \
  -d "{}"
```

### 2. 按集合查询文件夹基础信息

```bash
curl -X POST "http://localhost:17170/api/CollectionFoldersBasic" \
  -H "Content-Type: application/json" \
  -d "{\"collection_id\":1}"
```

### 3. 按文件夹查询接口基础信息

```bash
curl -X POST "http://localhost:17170/api/FolderApisBasic" \
  -H "Content-Type: application/json" \
  -d "{\"folder_id\":12}"
```

### 4. 按多个接口 ID 查询接口详情

```bash
curl -X POST "http://localhost:17170/api/ApisDetailByIds" \
  -H "Content-Type: application/json" \
  -d "{\"ids\":[101,102,103]}"
```

### 5. 新建文件夹

```bash
curl -X POST "http://localhost:17170/api/CreateDir" \
  -H "Content-Type: application/json" \
  -d "{\"name\":\"用户中心\",\"collection_id\":1}"
```

### 6. 创建接口

```bash
curl -X POST "http://localhost:17170/api/CreateApi" \
  -H "Content-Type: application/json" \
  -d "{\"folder_id\":12,\"collection_id\":1,\"name\":\"用户登录\",\"method\":\"POST\",\"url\":\"$Url$/v1/login\",\"protocol\":\"https\",\"query_params\":[],\"content_type\":\"application/json\",\"body_form\":[],\"body_json\":\"{\\\"username\\\":\\\"demo\\\",\\\"password\\\":\\\"123456\\\"}\",\"take_result\":[{\"key\":\"code\",\"type\":\"number\",\"desc\":\"状态码\"},{\"key\":\"data.token\",\"type\":\"string\",\"desc\":\"认证令牌\"}]}"
```

### 7. 批量导入

```bash
curl -X POST "http://localhost:17170/api/ApiBatchImport" \
  -F "collection_id=1" \
  -F "json={\"collection_id\":1,\"items\":[{\"type\":\"folder\",\"name\":\"用户中心\",\"children\":[]}]}"
```

### 8. 运行接口

```bash
curl -X POST "http://localhost:17170/api/ApiRun" \
  -H "Content-Type: application/json" \
  -d "{\"id\":201}"
```

### 9. 生成接口调用代码

```bash
curl -X POST "http://localhost:17170/api/ApiCode" \
  -H "Content-Type: application/json" \
  -d "{\"id\":201,\"code_type\":\"curl bash(chrome)\"}"
```

### 10. 提取 JSON 响应路径

```bash
curl -X POST "http://localhost:17170/api/ApiTakeJsonResult" \
  -H "Content-Type: application/json" \
  -d "{\"id\":201,\"json\":\"{\\\"code\\\":0,\\\"data\\\":{\\\"token\\\":\\\"abc\\\"}}\"}"
```

### 11. 查询集合环境

```bash
curl -X POST "http://localhost:17170/api/CollectionEnvs" \
  -H "Content-Type: application/json" \
  -d "{\"collection_id\":1}"
```

### 12. 查询环境变量

```bash
curl -X POST "http://localhost:17170/api/CollectionEnvItems" \
  -H "Content-Type: application/json" \
  -d "{\"collection_id\":1,\"env_id\":5}"
```
