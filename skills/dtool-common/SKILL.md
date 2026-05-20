---
name: dtool-common
description: Use when operating the dtool 通用工具模块 and the task involves uploading files to remote servers, Git branch operations (listing branches, pulling code, switching branches), querying database tables (MySQL/Pgsql), querying table structures, executing SQL SELECT queries, or updating knowledge fragments by file path.
---

# dtool 通用工具技能

- 提供远程文件上传、Git 分支查询与代码拉取、数据库表查询（MySQL/Pgsql）、表结构查询、SQL 查询、知识片段更新（按文件路径）、分支变更文件查看等通用接口。
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

知识片段以 Markdown 文件形式存储，按 `{年份}/{月份}/{uuid}.md` 的目录结构组织。通过传入相对于知识片段文件夹（`fragments/`）的路径来定位和更新片段。

### 9. 更新知识片段（按文件路径）

通过传入相对于知识片段文件夹的文件路径更新片段内容。Python 脚本会自动从路径中提取片段 ID 并调用保存接口。
注意：禁止更新文档名称（文件内容开头的title）

- **路径**: `/api/MemoryFragmentSave`
- **Python 函数**: `memory_fragment_update_by_path(relative_path, content)`（不传 title，禁止修改标题）
- **参数**:

| 字段 | 类型 | 必填 | 说明 |
|---|---|---|---|
| `relative_path` | string | 是 | 相对于知识片段文件夹（`fragments/`）的路径，如 `"2026/05/a59db79a-3e4d-4f37-a02d-1bf87cc0c590.md"` |
| `content` | string | 是 | 新的 Markdown 正文内容，支持占位符 `{需求文档纯文本文件相对地址}`（后端自动替换为实际路径） |

- **返回**: 更新后的片段对象，包含 `id`、`title`、`content`、`update_time_desc` 等字段
- **注意**: 函数内部从 `relative_path` 的文件名（去掉 `.md` 后缀）提取片段 ID，然后调用 `/api/MemoryFragmentSave`
- **示例**:
  ```python
  memory_fragment_update_by_path(
      "2026/05/a59db79a-3e4d-4f37-a02d-1bf87cc0c590.md",
      "## 更新后的内容\n\n新的正文...",
  )
  ```

### 10. 登录后抓取首个接口请求头

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

### 场景 8：更新知识片段（按文件路径）

1. 向用户确认 `base_url`、`Token`
2. 确认要更新的知识片段相对路径（如 `2026/05/uuid.md`）
3. 确认新的正文内容 `content`（禁止修改标题）
4. 调用 `memory_fragment_update_by_path(relative_path, content)`
5. 返回更新后的片段信息
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

## 通用代码编辑脚本

当 Edit 工具因中文、特殊字符或空白字符匹配失败时，使用 `code_edit.py` 替代。

原理：通过 JSON 描述文件声明要修改的文件和操作列表，脚本内部处理编码（BOM、CRLF/LF），每次修改前自动备份。

### 用法

```bash
# 预览（不写入）
python skills/dtool-common/scripts/code_edit.py <描述.json> --dry-run

# 执行修改
python skills/dtool-common/scripts/code_edit.py <描述.json>
```

### 描述文件格式

```json
{
  "files": [
    {
      "path": "C:/work/project/src/file.go",
      "ops": [
        {
          "op": "replace",
          "find": "待替换的原始文本（需精确匹配，含空白字符）",
          "replace": "替换后的文本"
        },
        {
          "op": "replace_all",
          "find": "全文多次出现的文本",
          "replace": "替换后的文本"
        },
        {
          "op": "insert_after",
          "after": "在此行文本之后插入（精确匹配该行）",
          "text": "要插入的文本"
        },
        {
          "op": "insert_before",
          "before": "在此行文本之前插入",
          "text": "要插入的文本"
        }
      ]
    }
  ]
}
```

### 支持的操作

| 操作 | 说明 |
|------|------|
| `replace` | 精确替换一处文本，若匹配到多处则报错（需改为 `replace_all`） |
| `replace_all` | 替换所有匹配到的文本 |
| `insert_after` | 在匹配行之后插入文本 |
| `insert_before` | 在匹配行之前插入文本 |

### 特性

- 自动检测并保留原文件 BOM、换行符（CRLF/LF）、末尾换行
- 修改前在同目录创建 `.bak` 备份文件
- `--dry-run` 仅打印将要执行的修改，不写入文件
- `find`/`after`/`before` 区分大小写、区分空白字符，要求精确匹配

### API 接口按 URI 同步脚本

按 URI 在 dtool 接口开发模块中执行"导入或更新"操作：

- `skills/dtool-common/scripts/sync_api_by_uri.py`

## Python 调用脚本

使用前需先向用户获取 `base_url`、`token`、`git_id`、`mysql_id`，然后替换脚本中的占位值。

详细脚本见 `scripts/dtool_common_api.py`。
