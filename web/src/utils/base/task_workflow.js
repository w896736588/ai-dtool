import base from '../base'

// TaskWorkflowCreateOrGet 查询或创建任务工作流。
function TaskWorkflowCreateOrGet(homeTaskId, callBack) {
  base.BasePost('/api/task/workflow/create_or_get', {
    home_task_id: homeTaskId,
  }, callBack)
}

// TaskWorkflowInfo 查询任务工作流详情。
function TaskWorkflowInfo(workflowId, callBack) {
  base.BasePost('/api/task/workflow/info', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowPromptsSave 保存工作流提示词。
function TaskWorkflowPromptsSave(data, callBack) {
  base.BasePost('/api/task/workflow/prompts/save', data, callBack)
}

// TaskWorkflowPromptsRestore 还原工作流提示词为默认值。
function TaskWorkflowPromptsRestore(workflowId, callBack) {
  base.BasePost('/api/task/workflow/prompts/restore', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowRequirementFetch 执行工作流首节点 TAPD 抓取。
function TaskWorkflowRequirementFetch(workflowId, callBack) {
  base.BasePost('/api/task/workflow/requirement/fetch', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowApiDocReset 重置接口文档，将所有关联文件夹下的接口 Markdown 合并覆盖到知识片段中。
function TaskWorkflowApiDocReset(workflowId, callBack) {
  base.BasePost('/api/task/workflow/api-doc/reset', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowBatchNodeStatus 批量查询工作流节点状态。
function TaskWorkflowBatchNodeStatus(homeTaskIds, callBack) {
  base.BasePost('/api/task/workflow/batch-node-status', {
    home_task_ids: homeTaskIds,
  }, callBack)
}

// TaskWorkflowNodeStatusUpdate 更新工作流节点状态。
function TaskWorkflowNodeStatusUpdate(workflowId, nodeStatuses, callBack) {
  base.BasePost('/api/task/workflow/node-status/update', {
    workflow_id: workflowId,
    node_statuses: nodeStatuses,
  }, callBack)
}

// TaskWorkflowIssueFixResolve 解析问题修改提示词模板。
function TaskWorkflowIssueFixResolve(workflowId, callBack) {
  base.BasePost('/api/task/workflow/issue-fix/resolve', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowChatSend 发送对话到 claude code。
function TaskWorkflowChatSend(workflowId, prompt, promptType, localDir, cliType, agentCliId, modelName, thinkingIntensity, callBack) {
  base.BasePost('/api/task/workflow/chat/send', {
    workflow_id: workflowId,
    prompt: prompt,
    prompt_type: promptType || '',
    local_dir: localDir,
    cli_type: cliType || 'claude',
    agent_cli_id: agentCliId || 0,
    model_name: modelName || '',
    thinking_intensity: thinkingIntensity || '高',
  }, callBack)
}

// TaskWorkflowChatStop 停止运行中的对话。
function TaskWorkflowChatStop(chatId, callBack) {
  base.BasePost('/api/task/workflow/chat/stop', {
    chat_id: chatId,
  }, callBack)
}

// TaskWorkflowChatContinue 继续已有对话。
function TaskWorkflowChatContinue(chatId, prompt, callBack) {
  base.BasePost('/api/task/workflow/chat/continue', {
    chat_id: chatId,
    prompt: prompt,
  }, callBack)
}

// TaskWorkflowChatList 列出对话列表。
function TaskWorkflowChatList(workflowId, callBack) {
  base.BasePost('/api/task/workflow/chat/list', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowChatDetail 获取对话详情。
function TaskWorkflowChatDetail(chatId, callBack) {
  base.BasePost('/api/task/workflow/chat/detail', {
    chat_id: chatId,
  }, callBack)
}

// TaskWorkflowChatDirs 获取可用的工作目录列表。
function TaskWorkflowChatDirs(workflowId, callBack) {
  base.BasePost('/api/task/workflow/chat/dirs', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowZcodeSave 保存 zcode 工作目录配置。
function TaskWorkflowZcodeSave(zcodeDir, callBack) {
  base.BasePost('/api/task/workflow/zcode/save', {
    zcode_dir: zcodeDir,
  }, callBack)
}

// TaskWorkflowZcodeGet 获取当前 zcode 配置及项目映射。
function TaskWorkflowZcodeGet(callBack) {
  base.BasePost('/api/task/workflow/zcode/get', {}, callBack)
}

// TaskWorkflowZcodeDelete 删除 zcode 配置。
function TaskWorkflowZcodeDelete(callBack) {
  base.BasePost('/api/task/workflow/zcode/delete', {}, callBack)
}

// TaskWorkflowChatListByPromptType 按提示词类型查询对话列表。
function TaskWorkflowChatListByPromptType(workflowId, promptType, callBack) {
  base.BasePost('/api/task/workflow/chat/list-by-prompt-type', {
    workflow_id: workflowId,
    prompt_type: promptType,
  }, callBack)
}

// TaskWorkflowChatListByAgentCli 按 Agent CLI 查询对话列表。
function TaskWorkflowChatListByAgentCli(agentCliId, callBack) {
  base.BasePost('/api/task/workflow/chat/list-by-agent-cli', {
    agent_cli_id: agentCliId,
  }, callBack)
}

// TaskWorkflowFileChangesSummary 获取文件变更汇总。
function TaskWorkflowFileChangesSummary(localDirs, callBack) {
  base.BasePost('/api/task/workflow/file-changes/summary', {
    local_dirs: localDirs,
  }, callBack)
}

// TaskWorkflowFileChangesDetail 获取文件变更详情（含 diff）。
function TaskWorkflowFileChangesDetail(localDir, parentBranch, callBack) {
  base.BasePost('/api/task/workflow/file-changes/detail', {
    local_dir: localDir,
    parent_branch: parentBranch || '',
  }, callBack)
}

export default {
  TaskWorkflowBatchNodeStatus,
  TaskWorkflowCreateOrGet,
  TaskWorkflowInfo,
  TaskWorkflowPromptsSave,
  TaskWorkflowPromptsRestore,
  TaskWorkflowRequirementFetch,
  TaskWorkflowApiDocReset,
  TaskWorkflowNodeStatusUpdate,
  TaskWorkflowIssueFixResolve,
  TaskWorkflowChatSend,
  TaskWorkflowChatStop,
  TaskWorkflowChatContinue,
  TaskWorkflowChatList,
  TaskWorkflowChatDetail,
  TaskWorkflowChatDirs,
  TaskWorkflowZcodeSave,
  TaskWorkflowZcodeGet,
  TaskWorkflowZcodeDelete,
  TaskWorkflowChatListByPromptType,
  TaskWorkflowChatListByAgentCli,
  TaskWorkflowFileChangesSummary,
  TaskWorkflowFileChangesDetail,
}
