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

export default {
  HomeTaskList,
  HomeTaskSave,
  HomeTaskArchiveToggle,
  HomeTaskStatusQuickUpdate,
  HomeTaskDelete,
  HomeTaskInfo,
  HomeTaskDailyReportGenerate,
  LocalDirList,
  OpenLocalDir,
}
