---
name: dtool-common
description: Use when operating the dtool 通用工具模块 and the task involves uploading files to remote servers, querying database tables (MySQL/Pgsql), querying table structures, or executing SQL SELECT queries.
---

# dtool 通用工具技能

提供远程文件上传、数据库表查询（MySQL/Pgsql）、表结构查询、SQL 查询四个通用接口。

## 强制约束

1. 调用接口前，必须向用户确认以下信息：
   - **请求地址**（`base_url`）：dtool 服务的完整地址，如 `http://192.168.1.100:17170`
   - **Token**：认证令牌，放在请求头 `Token` 中
   - **git_id**：Git 配置 ID（用于获取 SSH 远程连接信息，上传文件到远程项目时需要）
   - **mysql_id**：数据库配置 ID（支持 MySQL 和 Pgsql，使用数据库相关接口时需要）
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

## Python 调用脚本

使用前需先向用户获取 `base_url`、`token`、`git_id`、`mysql_id`，然后替换脚本中的占位值。

详细脚本见 `scripts/dtool_common_api.py`。
