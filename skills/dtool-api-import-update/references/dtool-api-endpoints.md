# dtool 接口开发模块 API 说明

## 使用前置要求

1. 默认服务地址：`http://localhost:17170`
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

常用参数：

- `id`
- `folder_id`
- `collection_id`
- `name`
- `method`
- `url`
- `protocol`
- `desc`
- `headers`
- `query_params`
- `content_type`
- `body_form`
- `body_json`
- `env_id`
- `response_take`
- `take_result`
- `take_result_desc`

创建示例：

```json
{
  "folder_id": 12,
  "collection_id": 1,
  "name": "用户登录",
  "method": "POST",
  "url": "$Url$/v1/login",
  "protocol": "https",
  "desc": "登录接口",
  "headers": {
    "Content-Type": "application/json"
  },
  "query_params": [],
  "content_type": "application/json",
  "body_form": [],
  "body_json": "{\"username\":\"demo\",\"password\":\"123456\"}"
}
```

更新示例：

```json
{
  "id": 201,
  "folder_id": 12,
  "collection_id": 1,
  "name": "用户登录",
  "method": "POST",
  "url": "$Url$/v1/login",
  "protocol": "https",
  "desc": "登录接口-更新",
  "headers": {
    "Content-Type": "application/json"
  },
  "query_params": [],
  "content_type": "application/json",
  "body_form": [],
  "body_json": "{\"username\":\"demo\",\"password\":\"new-password\"}"
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
  "name": "域名",
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
  "name": "域名",
  "collection_id": 1,
  "env_id": 5,
  "desc": "网关地址",
  "key": "Url",
  "value": "https://example.com"
}
```

## 五、运行、调试与辅助能力

### 1. 运行接口：`/api/ApiRun`

请求：

```json
{
  "id": 201
}
```

### 2. 生成代码：`/api/ApiCode`

请求：

```json
{
  "id": 201,
  "code_type": "curl bash(chrome)"
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
  "json": "{\"code\":0,\"data\":{\"token\":\"abc\"}}"
}
```

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
          "body_json": "{\"username\":\"demo\",\"password\":\"123456\"}"
        }
      ]
    }
  ]
}
```

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
  -d "{\"folder_id\":12,\"collection_id\":1,\"name\":\"用户登录\",\"method\":\"POST\",\"url\":\"$Url$/v1/login\",\"protocol\":\"https\"}"
```

### 7. 批量导入

```bash
curl -X POST "http://localhost:17170/api/ApiBatchImport" \
  -F "collection_id=1" \
  -F "json={\"collection_id\":1,\"items\":[{\"type\":\"folder\",\"name\":\"用户中心\",\"children\":[]}]}"
```
