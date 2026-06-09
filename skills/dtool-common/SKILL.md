---
name: dtool-common
description: Use when working with the dtool common module for remote file upload, Git operations, database queries, Docker service restart and log lookup, screenshot or browser capture helpers, knowledge fragment updates, task helpers, or shared helper scripts.
---

# dtool-common

## 这个 skill 可以做什么

- 上传本地文件到远程项目目录
- 查询 Git 当前分支、拉取代码、切换分支
- 查询数据库表列表、表结构、执行 SQL 查询或有限写入
- 重启 Docker Compose 服务
- 查询 Docker Compose 服务日志
- 网页截图
- 按文件路径更新知识片段
- 登录后抓取页面首个接口请求头
- 向任务追加 zcode sessionId
- 查看当前分支相对基线分支的改动文件和单文件 diff
- 提供通用代码编辑脚本和共享 API 调用脚本

## 必要约束

- 调用 dtool 前，先向用户确认所需参数：`base_url`、`Token`，以及任务相关的 `git_id`、`mysql_id`、`docker_id`、`smart_link_id`
- 需要调用 dtool 接口时，优先使用 `Python` 脚本，不直接拼 bash 请求
- 数据库查询优先使用只读方式；涉及写入时必须确认影响范围
- 使用 Git 相关能力时，不假设默认分支名，由用户明确指定
- 重启 Docker 服务前，先让用户明确 `docker_id` 和 `service`；查看日志时禁止使用 `-f` / `--follow`
- 需要具体参数、接口路径或脚本用法时，再去看 `scripts/` 下文件

## 细节位置

- 通用 API 基础封装：`scripts/api_common.py`
- Git 相关接口：`scripts/git_api.py`
- 数据库相关接口：`scripts/db_api.py`
- Docker 重启与日志接口：`scripts/docker_api.py`
- 网页截图接口：`scripts/screenshot_api.py`
- 浏览器请求头抓取接口：`scripts/browser_api.py`
- 知识片段更新接口：`scripts/memory_api.py`
- 任务辅助接口：`scripts/task_api.py`
- 查看分支改动文件：`scripts/show_branch_diff.py`
- 查看单文件 diff：`scripts/show_file_diff.py`
- 查看前端常见文件类型的全部改动及 diff（默认排除 `dist`）：`scripts/show_frontend_branch_diff.py`
- 查看后端常见文件类型与配置文件的全部改动及 diff，可排除指定目录（默认排除 `dist`）：`scripts/show_backend_branch_diff.py`
- 文本替换型代码编辑：`scripts/code_edit.py`
