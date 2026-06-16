import base from '../base'

// ssePost 向 SSE 端口发送 POST 请求（SSE 客户端只在 SSE 端口的 gin 实例上注册）。
function ssePost(uri, params, callBack) {
  const sseHost = base.GetSseApiHost()
  if (!sseHost) {
    callBack({ ErrCode: -1, ErrMsg: 'SSE连接未建立' })
    return
  }
  base.Globals().$axios.post(sseHost + uri, params, {
    headers: {
      'Content-Type': 'application/json',
      'Token': base.GetSafeToken(),
    }
  }).then(response => {
    if (response && response.data) {
      callBack(response.data)
    } else {
      callBack({ ErrCode: -1, ErrMsg: '响应数据为空' })
    }
  }).catch(error => {
    callBack({ ErrCode: -1, ErrMsg: error.message || '请求失败' })
  })
}

// HomeTaskList 查询首页任务列表。支持分页参数 page/page_size。
function HomeTaskList(isArchived, callBack, page = 0, pageSize = 0) {
  const params = { is_archived: isArchived }
  if (page > 0 && pageSize > 0) {
    params.page = page
    params.page_size = pageSize
  }
  base.BasePost('/api/HomeTaskList', params, callBack)
}

// HomeTaskCount 获取活跃和归档任务的数量。
function HomeTaskCount(callBack) {
  base.BasePost('/api/HomeTaskCount', {}, callBack)
}

// HomeTaskSave 保存首页任务。
function HomeTaskSave(data, callBack) {
  base.BasePost('/api/HomeTaskSave', data, callBack)
}

// HomeTaskArchiveToggle 切换首页任务归档状态。
function HomeTaskArchiveToggle(id, isArchived, callBack) {
  base.BasePost('/api/HomeTaskArchiveToggle', { id: id, is_archived: isArchived }, callBack)
}

// HomeTaskStatusQuickUpdate 快捷切换首页任务状态。
function HomeTaskStatusQuickUpdate(id, taskStatus, callBack) {
  base.BasePost('/api/HomeTaskStatusQuickUpdate', { id: id, task_status: taskStatus }, callBack)
}

// HomeTaskDelete 删除首页任务。
function HomeTaskDelete(id, callBack) {
  base.BasePost('/api/HomeTaskDelete', { id: id }, callBack)
}

// HomeTaskInfo 查询单条首页任务详情。
function HomeTaskInfo(id, callBack) {
  base.BasePost('/api/HomeTaskInfo', { id: id }, callBack)
}

// HomeTaskDailyReportGenerate 生成首页工作日报。
function HomeTaskDailyReportGenerate(callBack) {
  base.BasePost('/api/HomeTaskDailyReportGenerate', {}, callBack)
}

// LocalDirList 浏览本地目录，返回子目录列表。
function LocalDirList(dirPath, callBack) {
  base.BasePost('/api/Set/LocalDirList', { path: dirPath || '' }, callBack)
}

// OpenLocalDir 使用系统文件管理器打开指定本地目录。
function OpenLocalDir(dirPath, callBack) {
  base.BasePost('/api/Set/OpenLocalDir', { path: dirPath }, callBack)
}

// LocalDirBatchCheck 批量检查本地目录是否存在。
function LocalDirBatchCheck(paths, callBack) {
  base.BasePost('/api/Set/LocalDirBatchCheck', { paths: paths }, callBack)
}

// HomeTaskLastDevConfigByGitId 根据 Git 仓库 ID 查找最近匹配的开发配置。
function HomeTaskLastDevConfigByGitId(gitId, callBack) {
  base.BasePost('/api/HomeTaskLastDevConfigByGitId', { git_id: gitId }, callBack)
}

// HomeTaskBranchNameGenerate 使用 AI 生成分支名。
function HomeTaskBranchNameGenerate(taskName, parentBranch, createdDate, callBack) {
  base.BasePost('/api/HomeTaskBranchNameGenerate', { task_name: taskName, parent_branch: parentBranch, created_date: createdDate }, callBack)
}

// HomeTaskUnusedLocalDirs 查询历史任务中未被活跃任务占用的本地目录。
function HomeTaskUnusedLocalDirs(excludeTaskId, callBack) {
  base.BasePost('/api/HomeTaskUnusedLocalDirs', { exclude_task_id: excludeTaskId }, callBack)
}

// HomeTaskPageDataLoad 触发后端 SSE 推送页面附加数据（Git列表/集合/Docker/MySQL/SmartLink/记忆库/工作流计数）。
// 注意：必须 POST 到 SSE 端口，因为 SSE 客户端只在 SSE 端口的 gin 实例上注册。
function HomeTaskPageDataLoad(clientId, taskIds, callBack) {
  ssePost('/api/HomeTaskPageDataLoad', { client_id: clientId, task_ids: taskIds }, callBack)
}

// HomeTaskPageDataDirCheck 触发后端 SSE 推送本地目录存在性检查结果。
function HomeTaskPageDataDirCheck(clientId, paths, callBack) {
  ssePost('/api/HomeTaskPageDataDirCheck', { client_id: clientId, paths: paths }, callBack)
}

// HomeTaskPageDataBranchCheck 触发后端 SSE 推送分支匹配状态检查结果。
function HomeTaskPageDataBranchCheck(clientId, items, callBack) {
  ssePost('/api/HomeTaskPageDataBranchCheck', { client_id: clientId, items: items }, callBack)
}

// TaskStatusList 查询所有任务状态。
function TaskStatusList(callBack) {
  base.BasePost('/api/TaskStatusList', {}, callBack)
}

// TaskStatusSave 新增或编辑任务状态。
function TaskStatusSave(data, callBack) {
  base.BasePost('/api/TaskStatusSave', data, callBack)
}

// TaskStatusDelete 删除任务状态。
function TaskStatusDelete(id, callBack) {
  base.BasePost('/api/TaskStatusDelete', { id: id }, callBack)
}

// TaskStatusSort 更新任务状态排序。
function TaskStatusSort(ids, callBack) {
  base.BasePost('/api/TaskStatusSort', { ids: ids }, callBack)
}

// LocalBranchBatchCheck 批量检查本地目录当前 Git 分支是否与期望分支匹配。
function LocalBranchBatchCheck(items, callBack) {
  base.BasePost('/api/Set/LocalBranchBatchCheck', { items: items }, callBack)
}

// LocalBranchMismatchDetail 查询分支匹配详情，并返回未提交/已变更文件。
function LocalBranchMismatchDetail(items, callBack) {
  base.BasePost('/api/Set/LocalBranchMismatchDetail', { items: items }, callBack)
}

// RemoteBranchCheck 批量检查本地目录当前 Git 分支的远程推送状态和同步状态。
function RemoteBranchCheck(items, callBack) {
  base.BasePost('/api/Set/RemoteBranchCheck', { items: items }, callBack)
}

// RemoteBranchPush 推送当前分支并设置上游追踪。
function RemoteBranchPush(data, callBack) {
  base.BasePost('/api/Set/RemoteBranchPush', data, callBack)
}

// RemoteBranchSwitch 通过 git_id 切换远程工作目录的分支。
function RemoteBranchSwitch(gitId, branchName, callBack) {
  base.BasePost('/api/GitChangeBranchById', { git_id: gitId, branch_name: branchName }, callBack)
}

export default {
  HomeTaskList,
  HomeTaskCount,
  HomeTaskSave,
  HomeTaskArchiveToggle,
  HomeTaskStatusQuickUpdate,
  HomeTaskDelete,
  HomeTaskInfo,
  HomeTaskDailyReportGenerate,
  LocalDirList,
  OpenLocalDir,
  LocalDirBatchCheck,
  HomeTaskLastDevConfigByGitId,
  HomeTaskBranchNameGenerate,
  HomeTaskUnusedLocalDirs,
  LocalBranchBatchCheck,
  LocalBranchMismatchDetail,
  RemoteBranchCheck,
  RemoteBranchPush,
  RemoteBranchSwitch,
  HomeTaskPageDataLoad,
  HomeTaskPageDataDirCheck,
  HomeTaskPageDataBranchCheck,
  TaskStatusList,
  TaskStatusSave,
  TaskStatusDelete,
  TaskStatusSort,
}
