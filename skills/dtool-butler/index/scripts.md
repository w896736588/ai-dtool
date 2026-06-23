# 脚本工具索引

共 11 个脚本工具。

## [dtool-api] 接口模块：集合/文件夹/接口的查询、创建、更新、删除、移动；批量导入接口定义；环境与变量管理；接口运行、结果查看、字段提取、代码生成；基于分支 diff 定位接口变更
- skills/dtool-api/scripts/sync_api_by_uri.py — 同步/创建/更新接口定义

## [dtool-common] 公共模块：统一 dtool API 调用封装、任务 SessionId 追加、通用代码编辑（精确文本替换/插入）
- skills/dtool-common/scripts/api_common.py — 统一 dtool API 调用封装与任务 SessionId 追加
- skills/dtool-common/scripts/code_edit.py — 精确文本替换/插入的通用代码编辑

## [dtool-db] 数据库模块：查询数据库配置对应的所有表、查询指定表结构、执行 SELECT/写入操作（MySQL / Pgsql）
- skills/dtool-db/scripts/db_api.py — 数据库表查询、表结构查询、SQL 执行读写

## [dtool-docker] Docker 模块：重启 Docker Compose 指定服务、查询服务日志
- skills/dtool-docker/scripts/docker_api.py — 重启服务、查询服务日志

## [dtool-git] Git 模块：上传文件、查询/切换分支、拉取代码、查看分支改动
- skills/dtool-git/scripts/git_api.py — 上传本地文件到远程、查询当前分支、拉取代码、切换分支
- skills/dtool-git/scripts/show_branch_diff.py — 查看当前分支相对基线的改动文件列表
- skills/dtool-git/scripts/show_file_diff.py — 查看单文件完整 diff
- skills/dtool-git/scripts/show_file_changes.py — 查看指定文件的变更详情
- skills/dtool-git/scripts/show_backend_branch_diff.py — 查看后端常见文件类型的完整改动及 diff
- skills/dtool-git/scripts/show_frontend_branch_diff.py — 查看前端常见文件类型的完整改动及 diff

## [dtool-know] 知识片段模块：按片段ID更新知识片段内容
- skills/dtool-know/scripts/memory_api.py — 按片段ID更新知识片段内容（不修改标题）

## [dtool-notify] 通知模块：发送钉钉群文本通知
- skills/dtool-notify/scripts/send_dingtalk.py — 钉钉群文本通知（支持普通 Webhook 和带签名密钥的安全模式）

## [dtool-playwright] 浏览器模块：smart-link 登录打开页面、MCP/Playwright 接管浏览器、抓取请求头、网页截图
- skills/dtool-playwright/scripts/browser_api.py — smart-link 登录打开目标页面、MCP 模式接管浏览器会话、抓取页面首个接口请求头
- skills/dtool-playwright/scripts/dtool_playwright_api.py — Playwright 持久化目录模式接管浏览器
- skills/dtool-playwright/scripts/screenshot_api.py — 网页截图

## [dtool-workflow] 工作流模块：更新工作流节点状态
- skills/dtool-workflow/scripts/update_workflow_status.py — 更新工作流节点状态（支持自定义步骤 custom_{id}）

## [dtool-butler] 自进化生成：任务状态查询、自测任务列表、知识片段统计
- skills/dtool-butler/scripts/list_tasks_in_status.py — 查询任务清单中处于自测状态的任务（自进化生成）
- skills/dtool-butler/scripts/query_git_branch.py — 查询指定 Git 分组和项目的当前分支（支持命令行参数，使用 api_common）[更新：替换硬编码脚本，接受参数并调用 api_common]
