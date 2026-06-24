# dtool HTTP 接口索引

共 395 个已注册接口。所有接口均为 POST 方法，需携带 Token 鉴权。
此索引由 `/init` 自动根据当前代码生成，始终与代码同步。

## 其他

| 接口路径 | 说明 |
|----------|------|
| / | func1 |
| /favicon.ico | func1 |

## Git

| 接口路径 | 说明 |
|----------|------|
| /api/GitCleanupAndSwitchBranchById | func1 |
| /api/GitChangeBranch | GitChangeBranch |
| /api/GitChangeBranchRemote | GitChangeBranchRemote |
| /api/GitChangeBranchById | GitChangeBranchById |
| /api/GitCommitLog | GitCommitLog |
| /api/GitConfigList | Git 列表 |
| /api/GitCurrentBranch | GitCurrentBranchById |
| /api/GitPull | Git 拉取 |
| /api/GitPullBranchOrigin | GitPullBranchOrigin |
| /api/GitPendingCommitPush | Git 推送 |
| /api/GitQueryCurrentBranch | GitCurrentBranch |
| /api/GitQueryStatus | Query状态 |
| /api/GitSetSafeLog | GitSetSafeLog |
| /api/GitSaveCredentials | GitSaveCredentials |
| /api/GitRemoteBranchList | Git 列表 |
| /api/GitUploadFile | GitUploadFile |

## GitLab

| 接口路径 | 说明 |
|----------|------|
| /api/GitLab | func1 |

## 记忆库

| 接口路径 | 说明 |
|----------|------|
| /api/MemoryFragmentDownloadZip | MemoryFragmentDownloadZip |
| /api/MemoryFragmentSave | 记忆片段 保存 |
| /api/MemoryFragmentSaveById | MemoryFragmentSaveById |
| /api/MemoryFragmentShareInfo | 记忆片段 详情 |
| /api/MemoryFragmentSearch | 记忆片段 搜索 |
| /api/MemoryFragmentFolderChange | 记忆片段 变更 |
| /api/MemoryFragmentFolderList | 记忆片段 列表 |
| /api/MemoryFragmentHistoryList | 记忆片段 列表 |
| /api/MemoryFragmentTrashList | 记忆片段 列表 |
| /api/MemoryFragmentTagList | 记忆片段 列表 |
| /api/MemoryFragmentInfo | 记忆片段 详情 |
| /api/MemoryFragmentImageUpload | 记忆库 上传 |
| /api/MemoryFragmentRestore | 记忆片段 恢复 |
| /api/MemoryFragmentReferences | MemoryFragmentReferences |
| /api/MemoryFragmentDelete | 记忆片段 删除 |

## 记忆片段

| 接口路径 | 说明 |
|----------|------|
| /api/MemoryFragmentAiSearch | func1 |
| /api/MemoryFragmentShareCreate | 记忆片段 创建 |
| /api/MemoryFragmentStatus | 记忆片段 状态 |
| /api/MemoryFragmentFolderCreate | 记忆片段 创建 |
| /api/MemoryFragmentFolderUpdate | 记忆片段 更新 |
| /api/MemoryFragmentHardDelete | 记忆片段 删除 |
| /api/MemoryFragmentUploadZip | MemoryFragmentUploadZip |
| /api/MemoryFragmentUpdateZip | MemoryFragmentUpdateZip |
| /api/MemoryFragmentList | 记忆片段 列表 |
| /api/MemoryFragmentCreate | 记忆片段 创建 |
| /api/MemoryFragmentOrganize | 记忆片段 整理 |
| /api/MemoryFragmentBatchInfoByPaths | MemoryFragmentBatchInfoByPaths |

## download

| 接口路径 | 说明 |
|----------|------|
| /api/download/:name | DownloadWebFile |

## *filepath

| 接口路径 | 说明 |
|----------|------|
| /js/*filepath | func1 |
| /css/*filepath | func1 |

## Git 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/GitLabTokenCreate | Git 新增 |
| /api/Set/GitLabTokenList | Git 列表 |
| /api/Set/GitList | Git 列表 |
| /api/Set/GitAdd | Git 新增 |
| /api/Set/GitDelete | Git 删除 |

## GitLab 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/GitLabTokenDelete | Git 删除 |

## Git分组 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/GitGroupList | Git分组 列表 |
| /api/Set/GitGroupAdd | Git分组 新增 |
| /api/Set/GitGroupDelete | Git分组 删除 |

## Git快捷 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/GitQuickList | Git快捷 列表 |

## 全局配置 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/GlobalCreate | 全局配置 新增 |
| /api/Set/GlobalDelete | 全局配置 删除 |
| /api/Set/GlobalList | 全局配置 列表 |

## 分组 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/GroupList | 分组 列表 |
| /api/Set/GroupAdd | 分组 新增 |
| /api/Set/GroupDelete | 分组 删除 |

## AI模型 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/AiModelList | AI模型 列表 |
| /api/Set/AiModelAdd | AI模型 新增 |
| /api/Set/AiModelDelete | AI模型 删除 |
| /api/Set/AiModelTest | AI模型 测试 |

## AI服务商 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/AiProviderList | AI服务商 列表 |
| /api/Set/AiProviderAdd | AI服务商 新增 |
| /api/Set/AiProviderDelete | AI服务商 删除 |

## AI日志 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/AiRequestLogList | AI日志 列表 |

## 账号 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/AccountGroupList | 账号分组 列表 |
| /api/Set/AccountGroupAdd | 账号分组 新增 |
| /api/Set/AccountGroupDelete | 账号分组 删除 |
| /api/Set/AccountList | 账号 列表 |
| /api/Set/AccountAdd | 账号 新增 |
| /api/Set/AccountDelete | 账号 删除 |

## 管家机器人 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/ButlerBotConfigList | 管家机器人 列表 |
| /api/Set/ButlerBotConfigAdd | 管家机器人 新增 |
| /api/Set/ButlerBotConfigDelete | 管家机器人 删除 |

## 管家 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/ButlerBotConfigTest | 管家机器人 测试 |
| /api/Set/ButlerRoleList | 管家角色 列表 |
| /api/Set/ButlerTaskList | 管家 列表 |
| /api/Set/ButlerMessageList | 管家消息 列表 |

## 管家角色 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/ButlerRoleAdd | 管家角色 新增 |
| /api/Set/ButlerRoleDelete | 管家角色 删除 |

## 管家参数 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/ButlerConfigList | 管家参数 列表 |
| /api/Set/ButlerConfigAdd | 管家参数 新增 |
| /api/Set/ButlerConfigDelete | 管家参数 删除 |

## 管家API 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/ButlerApiIndex | SetButlerApiIndex |

## SSH 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/SshList | SSH 列表 |
| /api/Set/SshStatus | SSH 状态 |
| /api/Set/SshAdd | SSH 新增 |
| /api/Set/SshDelete | SSH 删除 |

## Supervisor 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/SupervisorList | Supervisor 列表 |
| /api/Set/SupervisorAdd | Supervisor 新增 |
| /api/Set/SupervisorDelete | Supervisor 删除 |

## 智能链接分组 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/SmartLinkGroupList | 智能链接分组 列表 |

## 智能链接 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/SmartLinkGroupAdd | 智能链接分组 新增 |
| /api/Set/SmartLinkGroupDelete | 智能链接分组 删除 |

## MySQL 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/MysqlList | MySQL 列表 |
| /api/Set/MysqlAdd | MySQL 新增 |
| /api/Set/MysqlDelete | MySQL 删除 |

## 记忆库 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/MemoryConfigGet | SetMemoryConfigGet |

## 记忆库配置 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/MemoryConfigSave | 记忆库配置 保存 |

## 主库 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/MainDBStorageAnalysis | 主库 分析 |
| /api/Set/MainDBStorageVacuum | 主库 回收 |

## Redis 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/RedisList | Redis 列表 |
| /api/Set/RedisAdd | Redis 新增 |
| /api/Set/RedisDelete | Redis 删除 |

## 远程分支 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/RemoteBranchCheck | 远程分支 检查 |
| /api/Set/RemoteBranchPush | 远程分支 推送 |

## 运行时配置 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/RuntimeConfigSave | 运行时配置 保存 |
| /api/Set/RuntimeConfigItemSave | 运行时配置 保存 |

## 命令分组 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/CmdGroupList | 命令分组 列表 |
| /api/Set/CmdGroupAdd | 命令分组 新增 |
| /api/Set/CmdGroupDelete | 命令分组 删除 |

## 定时任务 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/CronConfigTypes | SetCronConfigTypes |
| /api/Set/CronConfigGet | SetCronConfigGet |
| /api/Set/CronConfigSave | 定时任务 保存 |

## 本地目录 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/LocalDirList | 本地目录 列表 |
| /api/Set/LocalDirBatchCheck | 本地目录 检查 |

## 本地分支 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/LocalBranchBatchCheck | 本地分支 检查 |
| /api/Set/LocalBranchMismatchDetail | 本地分支 详情 |

## 变量分组 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/VariableGroupList | 变量分组 列表 |
| /api/Set/VariableGroupAdd | 变量分组 新增 |

## 变量 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/VariableGroupDelete | 变量分组 删除 |

## Docker 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/DockerComposeList | Docker 列表 |
| /api/Set/DockerComposeAdd | Docker 新增 |
| /api/Set/DockerComposeDelete | Docker 删除 |

## 首页任务 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/HomeTaskConfigGet | SetHomeTaskConfigGet |

## 首页任务配置 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/HomeTaskConfigSave | 首页任务配置 保存 |

## Prompt变更 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/PromptChangeLogList | Prompt变更 列表 |

## 打开 配置

| 接口路径 | 说明 |
|----------|------|
| /api/Set/OpenLocalDir | SetOpenLocalDir |

## 智能链接项

| 接口路径 | 说明 |
|----------|------|
| /api/SmartLinkItemList | 智能链接项 列表 |

## 智能链接

| 接口路径 | 说明 |
|----------|------|
| /api/SmartLinkItemAdd | 智能链接项 新增 |
| /api/SmartLinkItemDelete | 智能链接项 删除 |
| /api/SmartLinkItemInfo | 智能链接项 详情 |
| /api/SmartLinkInfo | 智能链接 详情 |
| /api/SmartLinkRun | SmartLinkRunPlaywright |
| /api/SmartLinkRunList | 智能链接 列表 |
| /api/SmartLinkRecycle | 智能链接 回收 |
| /api/SmartLinkChromeDownload | SmartLinkUpWebkit |
| /api/SmartLinkDel | 智能链接 删除 |
| /api/SmartLinkList | 智能链接 列表 |
| /api/SmartLinkLocatorAutoExtract | SmartLinkLocatorAutoExtract |
| /api/SmartLinkAdd | 智能链接 新增 |
| /api/SmartLinkMigrateOldData | SmartLinkMigrateOldData |
| /api/SmartLinkOpenDataDir | SmartLinkOpenDataDir |
| /api/smart-link/scrape-to-markdown | SmartLinkScrapeToMarkdown |

## 智能链接Chrome

| 接口路径 | 说明 |
|----------|------|
| /api/SmartLinkChromeVersion | 智能链接 版本 |

## 智能链接下载

| 接口路径 | 说明 |
|----------|------|
| /api/SmartLinkDownloadPath | SmartLinkDownloadPath |

## 智能流程

| 接口路径 | 说明 |
|----------|------|
| /api/SmartProcessItemList | 智能流程项 列表 |
| /api/SmartProcessItemAdd | 智能流程项 新增 |
| /api/SmartProcessItemDelete | 智能流程项 删除 |
| /api/SmartProcessSetPosition | SmartProcessSetPosition |
| /api/SmartProcessSetRelation | SmartProcessSetRelation |
| /api/SmartProcessList | 智能流程 列表 |
| /api/SmartProcessAdd | 智能流程 新增 |
| /api/SmartProcessDelete | 智能流程 删除 |
| /api/SmartProcessCancelRelation | SmartProcessCancelRelation |

## 智能流程项

| 接口路径 | 说明 |
|----------|------|
| /api/SmartProcessItemSort | 智能流程项 排序 |

## Supervisor

| 接口路径 | 说明 |
|----------|------|
| /api/SupervisorStop | Supervisor 停止 |
| /api/SupervisorStopAll | Supervisor 全部 |
| /api/SupervisorStatusList | Supervisor 列表 |
| /api/SupervisorConfigShow | SupervisorConfigShow |
| /api/SupervisorConfigList | Supervisor 列表 |
| /api/SupervisorConfList | Supervisor 列表 |
| /api/SupervisorRestart | Supervisor 重启 |
| /api/SupervisorRestartAll | Supervisor 全部 |

## Shell

| 接口路径 | 说明 |
|----------|------|
| /api/ShellOutRuleSetList | Shell规则集 列表 |
| /api/ShellOutRuleSetInfo | Shell规则集 详情 |
| /api/ShellOutRuleSetDelete | Shell规则集 删除 |
| /api/ShellOutRuleImportLegacy | ShellOutRuleImportLegacy |

## Shell规则集

| 接口路径 | 说明 |
|----------|------|
| /api/ShellOutRuleSetSave | Shell规则集 保存 |

## 收藏

| 接口路径 | 说明 |
|----------|------|
| /api/StarList | 收藏 列表 |
| /api/StarAdd | 收藏 新增 |
| /api/StarDel | 收藏 删除 |

## SSE

| 接口路径 | 说明 |
|----------|------|
| /api/SseAvailablePort | SseAvailablePort |
| /api/SseConnectionDetails | SseConnectionDetails |

## 截图

| 接口路径 | 说明 |
|----------|------|
| /api/Screenshot | Screenshot |

## MCP

| 接口路径 | 说明 |
|----------|------|
| /api/McpBindingList | MCP绑定 列表 |
| /api/McpBindingAdd | MCP绑定 新增 |
| /api/McpBindingRemove | MCP绑定 移除 |
| /api/McpBindingInstruction | MCP绑定 说明 |
| /api/McpChromeDevtoolsConfigSave | MCP工具 保存 |
| /api/McpChromeDevtoolsConfigDelete | MCP工具 删除 |
| /api/McpConfigPreview | MCP配置 预览 |
| /api/McpAgentTargetDelete | MCP目标 删除 |

## MCP工具

| 接口路径 | 说明 |
|----------|------|
| /api/McpChromeDevtoolsConfigList | MCP工具 列表 |

## MCP目标

| 接口路径 | 说明 |
|----------|------|
| /api/McpAgentTargetList | MCP目标 列表 |
| /api/McpAgentTargetSave | MCP目标 保存 |

## MCP类型

| 接口路径 | 说明 |
|----------|------|
| /api/McpTypeList | MCP类型 列表 |

## Markdown

| 接口路径 | 说明 |
|----------|------|
| /api/MarkdownHistoryList | Markdown 列表 |
| /api/MarkdownHistoryDel | Markdown 删除 |
| /api/MarkdownList | Markdown 列表 |
| /api/MarkdownAdd | Markdown 新增 |
| /api/MarkdownDel | Markdown 删除 |
| /api/MarkdownSort | Markdown 排序 |

## MySQL

| 接口路径 | 说明 |
|----------|------|
| /api/MysqlTables | MysqlTables |
| /api/MysqlTableStructure | MysqlTableStructure |
| /api/MysqlQuery | MySQL 查询 |
| /api/MysqlExec | MysqlExec |

## 任务workflow

| 接口路径 | 说明 |
|----------|------|
| /api/task/workflow/chat/list | 任务工作流 列表 |
| /api/task/workflow/chat/list-by-prompt-type | 任务工作流 类型 |
| /api/task/workflow/chat/list-by-agent-cli | TaskWorkflowChatListByAgentCli |
| /api/task/workflow/chat/send | 任务工作流 发送 |
| /api/task/workflow/chat/stop | 任务工作流 停止 |
| /api/task/workflow/chat/detail | 任务工作流 详情 |
| /api/task/workflow/chat/dirs | TaskWorkflowChatDirs |
| /api/task/workflow/chat/continue | 任务工作流 继续 |
| /api/task/workflow/coverage/generate | 任务工作流 生成 |
| /api/task/workflow/coverage/info | 任务工作流 详情 |
| /api/task/workflow/create_or_get | TaskWorkflowCreateOrGet |
| /api/task/workflow/test-plan/generate | 任务工作流 生成 |
| /api/task/workflow/test-plan/info | 任务工作流 详情 |
| /api/task/workflow/test-run/execute | TaskWorkflowTestRunExecute |
| /api/task/workflow/test-run/list | 任务工作流 列表 |
| /api/task/workflow/dev-plan/init | TaskWorkflowDevPlanInit |
| /api/task/workflow/dev-plan/info | 任务工作流 详情 |
| /api/task/workflow/dev-plan/save | 任务工作流 保存 |
| /api/task/workflow/zcode/save | 任务工作流 保存 |
| /api/task/workflow/zcode/get | TaskWorkflowZcodeGet |
| /api/task/workflow/zcode/delete | 任务工作流 删除 |
| /api/task/workflow/file-changes/summary | TaskWorkflowFileChangesSummary |
| /api/task/workflow/file-changes/detail | 任务工作流 详情 |
| /api/task/workflow/file-changes/file-diff | TaskWorkflowFileChangesFileDiff |
| /api/task/workflow/ui-assist/generate | 任务工作流 生成 |
| /api/task/workflow/ui-assist/info | 任务工作流 详情 |
| /api/task/workflow/prompts/save | 任务工作流 保存 |
| /api/task/workflow/prompts/restore | 任务工作流 恢复 |
| /api/task/workflow/info | 任务工作流 详情 |
| /api/task/workflow/issue-fix/resolve | TaskWorkflowIssueFixResolve |
| /api/task/workflow/requirement/fetch | 任务工作流 拉取 |
| /api/task/workflow/api-doc/reset | 任务工作流 重置 |
| /api/task/workflow/node-status/update | 任务工作流 更新 |
| /api/task/workflow/batch-node-status | 任务工作流 状态 |
| /api/task/workflow/open-in-editor | TaskWorkflowOpenInEditor |

## AgentCLI分组

| 接口路径 | 说明 |
|----------|------|
| /api/AgentCliGroupList | AgentCLI分组 列表 |
| /api/AgentCliGroupDelete | AgentCLI分组 删除 |
| /api/AgentCliGroupRelSave | AgentCLI 保存 |

## AgentCLI

| 接口路径 | 说明 |
|----------|------|
| /api/AgentCliGroupSave | AgentCLI分组 保存 |
| /api/AgentCliPromptTemplateList | AgentCLI模板 列表 |
| /api/AgentCliPromptTemplateSave | AgentCLI模板 保存 |
| /api/AgentCliPromptTemplateDelete | AgentCLI模板 删除 |
| /api/AgentCliWriteMcpServers | AgentCliWriteMcpServers |
| /api/AgentCliWriteDeepSeek | AgentCliWriteDeepSeek |
| /api/AgentCliList | AgentCLI 列表 |
| /api/AgentCliSave | AgentCLI 保存 |
| /api/AgentCliDelete | AgentCLI 删除 |
| /api/AgentCliReadSettings | AgentCliReadSettings |
| /api/AgentCliToggleEnabled | AgentCliToggleEnabled |

## API

| 接口路径 | 说明 |
|----------|------|
| /api/Apis | Apis |
| /api/ApisDetailByIds | ApiApisDetailByIds |
| /api/ApiRun | API 执行 |
| /api/ApiCode | ApiCode |
| /api/ApiWeightDown | ApiWeightDown |
| /api/ApiTakeJsonResult | ApiTakeJsonResult |
| /api/ApiBatchImport | API 导入 |
| /api/ApiMove | ApiMoveApi |

## 异步任务

| 接口路径 | 说明 |
|----------|------|
| /api/AsyncTaskList | 异步任务 列表 |
| /api/AsyncTaskInfo | 异步任务 详情 |
| /api/AsyncTaskAction | 异步任务 操作 |
| /api/AsyncTaskDelete | 异步任务 删除 |
| /api/AsyncTaskRetry | 异步任务 重试 |

## 归档

| 接口路径 | 说明 |
|----------|------|
| /api/ArchiveFolderList | API 列表 |

## Git待提交

| 接口路径 | 说明 |
|----------|------|
| /api/GitPendingStatus | Git待提交 状态 |

## Git快捷

| 接口路径 | 说明 |
|----------|------|
| /api/GitQuickCreateBranch | GitQuickCreateBranch |

## Git分组

| 接口路径 | 说明 |
|----------|------|
| /api/GitGroupBranchList | Git分组 列表 |

## GetLocalIP

| 接口路径 | 说明 |
|----------|------|
| /api/GetLocalIP | GetLocalIP |

## Docker

| 接口路径 | 说明 |
|----------|------|
| /api/DockerComposeStatus | Docker 状态 |
| /api/DockerComposeStart | Docker 启动 |
| /api/DockerComposeStop | Docker 停止 |
| /api/DockerComposeServices | DockerComposeServices |
| /api/DockerComposeList | Docker 列表 |
| /api/DockerComposeRestart | Docker 重启 |
| /api/DockerComposeConfigShow | DockerComposeConfigShow |
| /api/DockerContainerStop | Docker容器 停止 |
| /api/DockerContainerRemove | Docker容器 移除 |
| /api/DockerImageList | Docker镜像 列表 |
| /api/DockerImageContainers | DockerImageContainers |
| /api/DockerServiceRestart | Docker服务 重启 |

## Docker容器

| 接口路径 | 说明 |
|----------|------|
| /api/DockerContainerLogTruncate | Docker 清空 |

## Docker镜像

| 接口路径 | 说明 |
|----------|------|
| /api/DockerImageRemove | Docker镜像 移除 |

## Docker服务

| 接口路径 | 说明 |
|----------|------|
| /api/DockerServiceLogs | Docker服务 日志 |

## DeleteCollection

| 接口路径 | 说明 |
|----------|------|
| /api/DeleteCollection | ApiDeleteCollection |

## DeleteApi

| 接口路径 | 说明 |
|----------|------|
| /api/DeleteApi | ApiDeleteApi |

## DeleteDir

| 接口路径 | 说明 |
|----------|------|
| /api/DeleteDir | ApiDeleteDir |

## 首页页面数据

| 接口路径 | 说明 |
|----------|------|
| /api/HomeTaskPageDataLoad | HomeTaskPageDataLoad |

## 首页任务

| 接口路径 | 说明 |
|----------|------|
| /api/HomeTaskPageDataDirCheck | CheckAndPushLocalDirs |
| /api/HomeTaskPageDataBranchCheck | CheckAndPushBranch状态 |
| /api/HomeTaskSave | 首页任务 保存 |
| /api/HomeTaskStatusQuickUpdate | 首页任务 更新 |
| /api/HomeTaskDelete | 首页任务 删除 |
| /api/HomeTaskDailyReportGenerate | 首页日报 生成 |
| /api/HomeTaskList | 首页任务 列表 |
| /api/HomeTaskLastDevConfigByGitId | HomeTaskLastDevConfigByGitId |
| /api/HomeTaskCount | HomeTaskCount |
| /api/HomeTaskInfo | 首页任务 详情 |
| /api/HomeTaskArchiveToggle | 首页任务 切换 |
| /api/HomeTaskBranchNameGenerate | 首页分支名 生成 |
| /api/HomeTaskZcodeSessionIdAppend | HomeTaskZcodeSessionIdAppend |

## 首页清理

| 接口路径 | 说明 |
|----------|------|
| /api/HomeTaskUnusedLocalDirs | HomeTaskUnusedLocalDirs |

## CreateCollection

| 接口路径 | 说明 |
|----------|------|
| /api/CreateCollection | ApiCreateCollection |

## CreateCollectionEnv

| 接口路径 | 说明 |
|----------|------|
| /api/CreateCollectionEnv | ApiCreateCollectionEnv |

## CreateCollectionEnvItem

| 接口路径 | 说明 |
|----------|------|
| /api/CreateCollectionEnvItem | ApiCreateCollectionEnvItem |

## CreateMerge

| 接口路径 | 说明 |
|----------|------|
| /api/CreateMerge | CreateMerge |

## CreateDir

| 接口路径 | 说明 |
|----------|------|
| /api/CreateDir | ApiCreateDir |

## CreateApi

| 接口路径 | 说明 |
|----------|------|
| /api/CreateApi | ApiCreateApi |

## API集合

| 接口路径 | 说明 |
|----------|------|
| /api/CollectionEnvs | ApiCollectionEnvs |
| /api/CollectionEnvItems | ApiCollectionEnvItems |
| /api/Collections | ApiCollections |
| /api/CollectionListBasic | API集合 基础信息 |
| /api/CollectionFoldersBasic | API集合 基础信息 |

## CleanupCandidateFolders

| 接口路径 | 说明 |
|----------|------|
| /api/CleanupCandidateFolders | ApiCleanupCandidateFolders |

## CleanupArchiveFolders

| 接口路径 | 说明 |
|----------|------|
| /api/CleanupArchiveFolders | ApiCleanupArchiveFolders |

## Redis

| 接口路径 | 说明 |
|----------|------|
| /api/RedisKeys | RedisKeys |
| /api/RedisKeysType | Redis 类型 |
| /api/RedisKeyType | Redis 类型 |
| /api/RedisDelKey | RedisDelKey |
| /api/RedisDelSub | Redis 子项 |
| /api/RedisDeleteAll | RedisDelAllKey |
| /api/RedisSearch | Redis 搜索 |
| /api/RedisSaveString | RedisSaveString |
| /api/RedisEditTtl | RedisEditTtl |
| /api/RedisEditSub | Redis 子项 |
| /api/RedisAvailableList | Redis 列表 |
| /api/RedisCreateCache | RedisCreateCache |

## RestoreFolder

| 接口路径 | 说明 |
|----------|------|
| /api/RestoreFolder | ApiRestoreFolder |

## 工具

| 接口路径 | 说明 |
|----------|------|
| /api/ToolManagedProcessStatus | 工具 状态 |
| /api/ToolManagedProcessStart | 托管进程 启动 |
| /api/ToolManagedProcessStop | 工具 停止 |
| /api/ToolManagedProcessEnsureRunning | 托管进程 保活 |
| /api/ToolPortProcessList | 端口进程 列表 |

## 托管进程

| 接口路径 | 说明 |
|----------|------|
| /api/ToolManagedProcessRestart | 托管进程 重启 |
| /api/ToolManagedProcessLogTail | 托管进程 尾行 |

## 端口进程

| 接口路径 | 说明 |
|----------|------|
| /api/ToolPortProcessKill | ToolPortProcessKill |

## 任务状态

| 接口路径 | 说明 |
|----------|------|
| /api/TaskStatusSave | 任务状态 保存 |
| /api/TaskStatusSort | 任务状态 排序 |
| /api/TaskStatusList | 任务状态 列表 |
| /api/TaskStatusDelete | 任务状态 删除 |

## shellOut

| 接口路径 | 说明 |
|----------|------|
| /api/shellOut | ShellOut |

## shellOutSetSeeId

| 接口路径 | 说明 |
|----------|------|
| /api/shellOutSetSeeId | ShellOutSetSeeId |

## shellOutSearchContent

| 接口路径 | 说明 |
|----------|------|
| /api/shellOutSearchContent | ShellOutSearchContent |

## shellOutStop

| 接口路径 | 说明 |
|----------|------|
| /api/shellOutStop | Shell 停止 |

## shellOutEdit

| 接口路径 | 说明 |
|----------|------|
| /api/shellOutEdit | ShellOutEdit |

## shellOutErrorContext

| 接口路径 | 说明 |
|----------|------|
| /api/shellOutErrorContext | ShellOutErrorContext |

## shellOutCleanErrors

| 接口路径 | 说明 |
|----------|------|
| /api/shellOutCleanErrors | ShellOutCleanErrors |

## shellOutCleanLog

| 接口路径 | 说明 |
|----------|------|
| /api/shellOutCleanLog | ShellOutCleanLog |

## shellOuts

| 接口路径 | 说明 |
|----------|------|
| /api/shellOuts | GetShellOuts |

## shellOutDelete

| 接口路径 | 说明 |
|----------|------|
| /api/shellOutDelete | Shell 删除 |

## shellOutReconnect

| 接口路径 | 说明 |
|----------|------|
| /api/shellOutReconnect | ShellOutReconnect |

## 工作流

| 接口路径 | 说明 |
|----------|------|
| /api/workflow/template/step/save | 工作流模板 保存 |
| /api/workflow/template/step/sort | 工作流模板 排序 |
| /api/workflow/template/step/delete | 工作流模板 删除 |
| /api/workflow/template/save | 工作流模板 保存 |
| /api/workflow/template/list | 工作流模板 列表 |
| /api/workflow/template/list-basic | 工作流模板 基础信息 |
| /api/workflow/template/delete | 工作流模板 删除 |
| /api/workflow/template/import | 工作流模板 导入 |
| /api/workflow/skill/list | 工作流技能 列表 |

## 变量

| 接口路径 | 说明 |
|----------|------|
| /api/VariableCmdAdd | 变量命令 新增 |
| /api/VariableSet | 变量命令 设置 |
| /api/VariableSetLogin | 变量 登录 |
| /api/VariableList | 变量 列表 |
| /api/VariableAdd | 变量 新增 |
| /api/VariableDel | 变量 删除 |
| /api/VariableInfo | 变量 详情 |
| /api/VariableRun | 变量命令 执行 |

## 变量命令

| 接口路径 | 说明 |
|----------|------|
| /api/VariableCmdDel | 变量命令 删除 |

## 登录

| 接口路径 | 说明 |
|----------|------|
| /api/BaseLogin | 基础 登录 |
| /api/BaseLoginStatus | 登录 状态 |

## 基础

| 接口路径 | 说明 |
|----------|------|
| /api/BaseRegisterService | BaseRegisterService |
| /api/BaseCheckUnikeyExist | BaseCheckUnikeyExist |
| /api/BaseSshList | SSH 列表 |

## Agent

| 接口路径 | 说明 |
|----------|------|
| /api/agent/chat/send | AgentChat发送 |
| /api/agent/chat/list-by-agent-cli | AgentChatListByAgentCli |
| /api/agent/chat/mark-read | AgentChat标记已读 |

## AI浏览器

| 接口路径 | 说明 |
|----------|------|
| /api/ai/browser/session/open | AIBrowserSessionOpen |
| /api/ai/browser/session/capture-headers | AIBrowserSessionCaptureHeaders |

## Webhook

| 接口路径 | 说明 |
|----------|------|
| /api/WebhookConfigList | Webhook 列表 |
| /api/WebhookConfigSave | Webhook 保存 |
| /api/WebhookConfigDelete | Webhook 删除 |
| /api/WebhookConfigTest | Webhook 测试 |

## PHP

| 接口路径 | 说明 |
|----------|------|
| /api/PhpUnserialize | PhpPhpUnSerialize |
| /api/PhpUnserialize2 | PhpPhpUnSerialize2 |

## PermanentDeleteDir

| 接口路径 | 说明 |
|----------|------|
| /api/PermanentDeleteDir | ApiPermanentDeleteDir |

## FolderApisBasic

| 接口路径 | 说明 |
|----------|------|
| /api/FolderApisBasic | API目录 基础信息 |

## FolderApisMarkdown

| 接口路径 | 说明 |
|----------|------|
| /api/FolderApisMarkdown | ApiFolderApisMarkdown |

## FolderDetail

| 接口路径 | 说明 |
|----------|------|
| /api/FolderDetail | API目录 详情 |

## Ip

| 接口路径 | 说明 |
|----------|------|
| /api/Ip | Ip |

## Upload

| 接口路径 | 说明 |
|----------|------|
| /api/Upload | 上传 |

## multiformdata

| 接口路径 | 说明 |
|----------|------|
| /test/multiformdata | func1 |

