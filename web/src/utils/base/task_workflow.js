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

// TaskWorkflowNodeStatusUpdate 更新工作流节点状态。
function TaskWorkflowNodeStatusUpdate(workflowId, nodeStatuses, callBack) {
  base.BasePost('/api/task/workflow/node-status/update', {
    workflow_id: workflowId,
    node_statuses: nodeStatuses,
  }, callBack)
}

export default {
  TaskWorkflowCreateOrGet,
  TaskWorkflowInfo,
  TaskWorkflowPromptsSave,
  TaskWorkflowPromptsRestore,
  TaskWorkflowRequirementFetch,
  TaskWorkflowApiDocReset,
  TaskWorkflowNodeStatusUpdate,
}
