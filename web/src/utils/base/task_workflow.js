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

export default {
  TaskWorkflowCreateOrGet,
  TaskWorkflowInfo,
  TaskWorkflowPromptsSave,
  TaskWorkflowPromptsRestore,
}
