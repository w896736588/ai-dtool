import base from '../base'

// HomeTaskList 查询首页任务列表。
function HomeTaskList(isArchived, callBack) {
  base.BasePost('/api/HomeTaskList', { is_archived: isArchived }, callBack)
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

export default {
  HomeTaskList,
  HomeTaskSave,
  HomeTaskArchiveToggle,
  HomeTaskStatusQuickUpdate,
}
