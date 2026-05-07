---
name: dtool-common
description: Use when operating the dtool 通用工具模块 and the task involves uploading files to remote servers, Git branch operations (listing branches, pulling code, switching branches), querying database tables (MySQL/Pgsql), querying table structures, executing SQL SELECT queries, or managing knowledge fragments (creating, editing, searching).
---

# dtool 通用工具技能

- 提供远程文件上传、Git 分支查询与代码拉取、数据库表查询（MySQL/Pgsql）、表结构查询、SQL 查询、知识片段管理、分支变更文件查看等通用接口。
- 新增浏览器登录后抓取接口请求头能力，可在登录完成后刷新页面并返回首个接口请求的 headers。
- dtool-common 不在 Skill 列表中，使用时直接内联 Python 调用其 API，Windows 路径用 r'...' 原始字符串。
## 强制约束

1. 调用接口前，必须向用户确认以下信息：
   - **请求地址**（`base_url`）：dtool 服务的完整地址，如 `http://192.168.1.100:17170`
   - **Token**：认证令牌，放在请求头 `Token` 中
   - **git_id**：Git 配置 ID（用于获取 SSH 远程连接信息，上传文件到远程项目时需要）
   - **mysql_id**：数据库配置 ID（支持 MySQL 和 Pgsql，使用数据库相关接口时需要）
   - **docker_id**：Docker Compose 配置 ID（使用 Docker 日志查询接口时需要）
2. 所有请求使用 `POST`，`Content-Type: application/json; charset=utf-8`。
3. 统一使用 Python 脚本发送请求，避免 bash 编码问题。

## 接口说明

### 1. 上传文件到远程项目

通过 git_id 获取 SSH 远程连接配置（主机、端口、认证信息）和项目路径，使用 SFTP/SCP 协议将文件传输到远程服务器的指定目录。支持一次调用上传多个文件。

- **路径**: `/api/GitUploadFile`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `git_id` | string | 是 | Git 配置 ID，用于获取 SSH 远程连接信息和项目路径 |
| `local_file_paths` | string 或 string[] | 是 | 要上传文件的绝对路径，支持单个字符串或列表（兼容旧参数 `local_file_path`） |
| `upload_dir` | string | 是 | 相对于远程项目根目录的上传目录，如 `src/config`、`public/uploads`（不允许 `..` 或以 `/` 开头） |

- **返回**:

`list` 数组，每项包含：

| 字段 | 说明 |
|---|---|
| `remote_path` | 远程服务器上的完整文件路径 |
| `file_name` | 文件名 |
| `file_size` | 文件大小（字节） |
| `git_id` | Git 配置 ID |

- **限制**: 每个文件最大 10MB


### 2. 查询数据库所有表

- **路径**: `/api/MysqlTables`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `mysql_id` | string | 是 | 数据库配置 ID（支持 MySQL 和 Pgsql） |

- **返回**: `list` 数组，每项包含 `table_name` 和 `table_comment`

### 3. 查询数据库表结构

- **路径**: `/api/MysqlTableStructure`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `mysql_id` | string | 是 | 数据库配置 ID（支持 MySQL 和 Pgsql） |
| `table_name` | string | 是 | 表名（仅允许字母、数字、下划线、点） |

- **返回**: `list` 数组，每项包含 Field, Type, Null, Key, Default, Extra, Comment 等字段

### 4. 执行数据库查询

- **路径**: `/api/MysqlQuery`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `mysql_id` | string | 是 | 数据库配置 ID（支持 MySQL 和 Pgsql） |
| `sql` | string | 是 | SQL 语句（仅允许 SELECT） |

- **返回**: `list` 数组，每项为查询结果的一行（字段名为 key）
- **安全限制**: 仅允许 `SELECT` 开头的 SQL，禁止 INSERT/UPDATE/DELETE/DROP 等

### 4.1 执行数据库写入

- **路径**: `/api/MysqlExec`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `mysql_id` | string | 是 | 数据库配置 ID（支持 MySQL 和 Pgsql） |
| `sql` | string | 是 | SQL 语句（允许 INSERT、UPDATE，禁止 DROP 等危险操作） |

- **返回**: 执行结果，包含影响行数等信息
- **安全限制**: 允许 INSERT、UPDATE，禁止 DROP、TRUNCATE、ALTER 等危险语句

### 5. 查询 Docker Compose 服务日志

- **路径**: `/api/DockerServiceLogs`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `docker_id` | int | 是 | Docker Compose 配置 ID（关联 tbl_docker_compose 表） |
| `command` | string | 是 | 日志查询命令，必须以 `docker compose logs` 开头 |

- **返回**: `logs` 字段，包含日志文本内容
- **安全限制**: command 必须以 `docker compose logs` 开头
- **超时**: 40 秒

示例 command：
- `docker compose logs nginx` — 查看 nginx 服务日志
- `docker compose logs --tail 100 nginx` — 查看 nginx 最近 100 行日志
- `docker compose logs --since 30m nginx php-fpm` — 查看 nginx 和 php-fpm 最近 30 分钟日志

### 6. 查询当前分支

通过 git_id 自动解析 SSH 连接和项目路径，查询指定 Git 仓库的当前分支和远程跟踪分支。

- **路径**: `/api/GitCurrentBranch`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `git_id` | string | 是 | Git 配置 ID（关联 tbl_git 表，自动获取 SSH 连接和 code_path） |

- **返回**: 文本内容，包含当前分支和远程分支信息

### 7. 拉取当前分支最新代码

通过 git_id 自动解析 SSH 连接和项目路径，拉取当前分支的最新代码（执行 git checkout . + clean + fetch + pull）。

- **路径**: `/api/GitPull`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `git_id` | string | 是 | Git 配置 ID（关联 tbl_git 表，自动获取 SSH 连接和 code_path） |

- **返回**: 文本内容，包含当前分支和远程分支信息

### 8. 切换分支

通过 git_id 自动解析 SSH 连接和项目路径，切换到指定分支（执行 git checkout . + clean + fetch + pull + checkout branch + pull origin branch）。

- **路径**: `/api/GitChangeBranchById`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `git_id` | string | 是 | Git 配置 ID（关联 tbl_git 表，自动获取 SSH 连接和 code_path） |
| `branch_name` | string | 是 | 要切换到的分支名（如 `master`、`dev`、`feature_xxx`） |

- **返回**: 文本内容，包含当前分支和远程分支信息

## 知识片段接口

**知识片段** 是 dtool 中用于持久化存储项目知识的载体。每个片段以 Markdown 文件形式存储在 memory 目录中，包含标题、正文内容和标签。典型用途包括：记录开发规范与约定、保存技术决策及其背景、沉淀问题排查经验、存储会议纪要等。片段支持 Git 版本管理，可按关键词搜索、分类标签筛选。

### 9. 创建知识片段

创建一个新的知识片段。不传 `id` 即为新建，接口会自动生成唯一 ID 并持久化为 Markdown 文件。

- **路径**: `/api/MemoryFragmentSave`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `title` | string | 是 | 片段标题，简明扼要描述内容主题，如 `"数据库迁移规范"` |
| `content` | string | 是 | 片段正文，支持 Markdown 格式（代码块、列表、表格等） |
| `tags` | string[] | 否 | 分类标签，用于后续按标签筛选，如 `["规范", "数据库"]` |

- **返回**: 新建的片段对象，包含 `id`、`title`、`content`、`tags`、`create_time_desc`、`update_time_desc` 等字段
- **示例**: 创建一个开发规范片段
  ```json
  {
    "title": "API开发规范",
    "content": "## 接口规范\n\n1. 所有接口使用 POST 方法\n2. 统一返回 {code, msg, data} 结构",
    "tags": ["规范", "后端"]
  }
  ```

### 10. 编辑知识片段

编辑已有的知识片段。传入 `id` 加上需要修改的字段，未传入的字段保持原值不变。

- **路径**: `/api/MemoryFragmentSave`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | string | 是 | 要编辑的片段 ID |
| `title` | string | 否 | 新标题（不传则不修改） |
| `content` | string | 否 | 新正文内容（不传则不修改） |
| `tags` | string[] | 否 | 新标签列表（不传则不修改） |

- **返回**: 更新后的片段对象
- **注意**: Python 脚本中的 `memory_fragment_edit` 会先调用查询接口获取当前值，自动填充未传入的字段，确保只更新指定字段

### 11. 查询知识片段明细

根据片段 ID 查询完整的知识片段内容，包括标题、正文、标签、创建和更新时间。

- **路径**: `/api/MemoryFragmentInfo`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `id` | string | 是 | 片段 ID |

- **返回**: 片段对象，包含 `id`、`title`、`content`、`tags`、`create_time_desc`、`update_time_desc` 等字段

### 12. 搜索知识片段（多关键词 AND）

按关键词搜索知识片段的标题和内容。支持多个关键词，**用空格分隔，之间为 AND 关系**（即所有关键词必须同时匹配才会返回结果）。结果按相关度排序：标题命中权重最高，标签次之，内容命中最低。

- **路径**: `/api/MemoryFragmentSearch`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `query` | string | 是 | 搜索关键词，多个关键词用空格分隔表示 AND 查询。如 `"数据库 迁移"` 表示同时包含"数据库"和"迁移"的片段 |
| `limit` | number | 否 | 返回结果数量上限（默认 20） |

- **返回**: `list` 数组，每项包含 `id`、`title`、`content`、`tags`、`update_time_desc`、`score`（匹配得分）等字段
- **示例**:
  - 单关键词: `{"query": "迁移"}` — 搜索包含"迁移"的片段
  - 多关键词 AND: `{"query": "数据库 迁移"}` — 搜索同时包含"数据库"和"迁移"的片段
- 三关键词 AND: `{"query": "API 规范 前端"}` — 搜索同时包含这三个词的片段

### 13. 登录后抓取首个接口请求头

使用与 `browser_profile_open` 一致的参数，服务端完成网页登录后会自动刷新当前页，抓取首个 `xhr/fetch` 接口请求的 request headers，返回后自动关闭浏览器。

- **路径**: `/api/ai/browser/session/capture-headers`
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `smart_link_id` | int | 是 | 自定义网页配置 ID |
| `label` | string | 是 | 要打开的链接标签名 |
| `account` | string | 否 | 账号用户名，留空表示不代入账号 |
| `open_type` | int | 否 | 打开类型，默认 `0` |
| `reuse_if_open` | bool | 否 | 已打开时是否复用已有浏览器 |
| `enable_mcp` | bool | 否 | 兼容保留参数，此接口会在抓取后关闭浏览器 |

- **返回**: `headers` 对象，直接为实际请求头键值对。

## 推荐工作流

### 场景 1：上传文件到远程项目

1. 向用户确认 `base_url`、`Token`、`git_id`（git_id 用于确定远程 SSH 连接和项目路径）
2. 确认 `local_file_path`（当前项目中要上传文件的绝对路径）
3. 确认 `upload_dir`（远程项目中的目标目录）
4. 调用 `/api/GitUploadFile`
5. 返回 `remote_path` 表示上传成功

### 场景 2：浏览数据库

1. 向用户确认 `base_url`、`Token`、`mysql_id`
2. 调用 `/api/MysqlTables` 获取所有表列表
3. 用户选择表后，调用 `/api/MysqlTableStructure` 查看表结构
4. 根据表结构，调用 `/api/MysqlQuery` 执行 SELECT 查询

### 场景 3：快速查询表数据

1. 已知表名时，直接调用 `/api/MysqlQuery` 执行 `SELECT * FROM table_name LIMIT 10`
2. 需要了解字段含义时，先调 `/api/MysqlTableStructure`

### 场景 3.1：写入数据库

1. 向用户确认 `base_url`、`Token`、`mysql_id`
2. 确认要执行的 INSERT 或 UPDATE 语句
3. 调用 `/api/MysqlExec`
4. 返回执行结果

### 场景 4：查询 Docker 服务日志

1. 向用户确认 `base_url`、`Token`、`docker_id`
2. 确认要查询的服务名和日志条件（如 `--tail 100`）
3. 拼接 command（必须以 `docker compose logs` 开头）
4. 调用 `/api/DockerServiceLogs`
5. 返回 `logs` 字段中的日志内容

### 场景 5：查询 Git 当前分支

1. 向用户确认 `base_url`、`Token`、`git_id`
2. 调用 `/api/GitCurrentBranch`
3. 返回当前分支和远程分支信息

### 场景 6：拉取 Git 最新代码

1. 向用户确认 `base_url`、`Token`、`git_id`
2. 调用 `/api/GitPull`
3. 返回当前分支和远程分支信息

### 场景 7：切换 Git 分支

1. 向用户确认 `base_url`、`Token`、`git_id`
2. 确认要切换的目标分支名 `branch_name`
3. 调用 `/api/GitChangeBranchById`
4. 返回切换后的当前分支和远程分支信息

### 场景 8：创建知识片段

1. 向用户确认 `base_url`、`Token`
2. 确认片段标题 `title` 和内容 `content`（支持 Markdown）
3. 确认是否需要标签 `tags`（可选，用于分类）
4. 调用 `/api/MemoryFragmentSave`（不传 `id`）
5. 返回新创建的片段信息（含自动生成的 `id`）

### 场景 9：编辑知识片段

1. 向用户确认 `base_url`、`Token`
2. 确认要编辑的片段 ID（可通过搜索接口获取）
3. 确认需要修改的字段（`title`、`content`、`tags`，未传的字段保持不变）
4. 调用 `/api/MemoryFragmentSave`（传入 `id` + 要修改的字段）
5. 返回更新后的片段信息

### 场景 10：查询知识片段明细

1. 向用户确认 `base_url`、`Token`
2. 确认要查询的片段 ID
3. 调用 `/api/MemoryFragmentInfo`
4. 返回片段完整内容（标题、正文、标签、时间等）

### 场景 11：搜索知识片段

1. 向用户确认 `base_url`、`Token`
2. 确认搜索关键词（多个关键词用空格分隔，AND 逻辑）
3. 调用 `/api/MemoryFragmentSearch`
4. 返回匹配的知识片段列表（按相关度排序）
## Git 分支变更查看脚本

用于查看当前分支相对基分支的改动文件列表和单文件 diff（类似 GitLab MR 文件列表），跨平台通用。

### 查看分支改动文件列表

```bash
python skills/dtool-common/scripts/show_branch_diff.py <基分支>
```

### 查看单文件 diff

```bash
python skills/dtool-common/scripts/show_file_diff.py <基分支> <文件路径>
```

### API 接口按 URI 同步脚本

按 URI 在 dtool 接口开发模块中执行"导入或更新"操作：

- `skills/dtool-common/scripts/sync_api_by_uri.py`

## Python 调用脚本

使用前需先向用户获取 `base_url`、`token`、`git_id`、`mysql_id`，然后替换脚本中的占位值。

详细脚本见 `scripts/dtool_common_api.py`。
