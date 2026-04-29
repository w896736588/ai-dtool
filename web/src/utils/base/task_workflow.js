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

// TaskWorkflowDevPlanInit 初始化开发执行文档。
function TaskWorkflowDevPlanInit(workflowId, callBack) {
  base.BasePost('/api/task/workflow/dev-plan/init', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowDevPlanInfo 查询开发执行文档详情。
function TaskWorkflowDevPlanInfo(workflowId, callBack) {
  base.BasePost('/api/task/workflow/dev-plan/info', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowDevPlanSave 保存开发执行文档。
function TaskWorkflowDevPlanSave(workflowId, content, callBack) {
  base.BasePost('/api/task/workflow/dev-plan/save', {
    workflow_id: workflowId,
    content: content,
  }, callBack)
}

// TaskWorkflowUIAssistGenerate 生成页面辅助识别结果。
function TaskWorkflowUIAssistGenerate(payload, callBack) {
  base.BasePost('/api/task/workflow/ui-assist/generate', payload, callBack)
}

// TaskWorkflowUIAssistInfo 查询页面辅助识别结果。
function TaskWorkflowUIAssistInfo(workflowId, callBack) {
  base.BasePost('/api/task/workflow/ui-assist/info', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowCoverageInfo 查询覆盖分析结果。
function TaskWorkflowCoverageInfo(workflowId, callBack) {
  base.BasePost('/api/task/workflow/coverage/info', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowCoverageGenerate 生成覆盖分析。
function TaskWorkflowCoverageGenerate(workflowId, callBack) {
  base.BasePost('/api/task/workflow/coverage/generate', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowTestPlanInfo 查询测试计划结果。
function TaskWorkflowTestPlanInfo(workflowId, callBack) {
  base.BasePost('/api/task/workflow/test-plan/info', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowTestPlanGenerate 生成测试计划。
function TaskWorkflowTestPlanGenerate(workflowId, callBack) {
  base.BasePost('/api/task/workflow/test-plan/generate', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowTestRunList 查询测试执行历史。
function TaskWorkflowTestRunList(workflowId, callBack) {
  base.BasePost('/api/task/workflow/test-run/list', {
    workflow_id: workflowId,
  }, callBack)
}

// TaskWorkflowTestRunExecute 执行测试计划。
function TaskWorkflowTestRunExecute(workflowId, regeneratePlan, includeCoverage, callBack) {
  base.BasePost('/api/task/workflow/test-run/execute', {
    workflow_id: workflowId,
    regenerate_plan: regeneratePlan,
    include_coverage: includeCoverage,
  }, callBack)
}

export default {
  TaskWorkflowCreateOrGet,
  TaskWorkflowInfo,
  TaskWorkflowDevPlanInit,
  TaskWorkflowDevPlanInfo,
  TaskWorkflowDevPlanSave,
  TaskWorkflowUIAssistGenerate,
  TaskWorkflowUIAssistInfo,
  TaskWorkflowCoverageInfo,
  TaskWorkflowCoverageGenerate,
  TaskWorkflowTestPlanInfo,
  TaskWorkflowTestPlanGenerate,
  TaskWorkflowTestRunExecute,
  TaskWorkflowTestRunList,
}
