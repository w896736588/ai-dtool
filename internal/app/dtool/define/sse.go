package define

const (
	SseAiCode                   = `ai_code`                     //废弃
	SseGitLab                   = `gitlab`                      //固定全局唯一
	SseShellConnections         = `shell_connections`           //Shell连接状态推送
	SseMemoryFragmentUpdates    = `memory_fragment_updates`     //知识片段变更推送
	SseMemoryFragmentStatus     = `memory_fragment_status`      //知识片段状态推送
	SseAsyncTasks               = `async_tasks`                 //异步任务状态推送
	SseSafeAuthRequired         = `safe_auth_required`          //安全认证失效通知
	SseApiDataChange            = `api_data_change`             //API数据变更推送
	SseGitPendingStatus         = `git_pending_status`          //Git待提交状态及倒计时推送
	SseAgentCliUnreadHome       = `agent_cli_unread_home`       //主页左侧 Agent Cli 菜单红点推送
	SseAgentCliUnreadGlobal     = `agent_cli_unread_global`     //Agent Cli 页面未读数推送
	SseWorkflowUnreadSnapshot   = `workflow_unread_snapshot`    //Workflow 未读红点快照推送
	SseWorkflowUnreadHomeMenu   = `workflow_unread_home_menu`   //主页左侧 Workflow 菜单红点推送
	SseWorkflowUnreadHomeTask   = `workflow_unread_home_task`   //任务清单 Workflow 红点推送
	SseWorkflowUnreadDetail     = `workflow_unread_detail`      //工作流详情页红点推送
	SseTaskWorkflowPrefix       = `task_workflow_`              //任务工作流步骤推送前缀
	SseTaskWorkflowChatPrefix   = `task_workflow_chat_`         //任务工作流 claude code 对话推送前缀
	SseConnectionCount          = `sse_connection_count`        //SSE连接数推送
	SseChromeDevtoolsPortStatus = `chrome_devtools_port_status` //Chrome DevTools 端口占用状态变更推送
)

const (
	SseEventClean = `[CLEAN]`                   //清除前端的数据
	SseEventLogin = `[LOGIN_USERNAME_PASSWORD]` //通知前端弹窗输入账号密码
	SseDown       = `[DONE]`                    //前端换个行
	SseConnect    = `[CONNECT]`                 //链接已建立
)

type SseEvent string
