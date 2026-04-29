import base from '../base'

// AsyncTaskList 查询异步任务列表与摘要。 // AsyncTaskList queries async task summary and recent items.
function AsyncTaskList(limit, callBack) {
  base.BasePost('/api/AsyncTaskList', {
    limit: limit,
  }, callBack)
}

// AsyncTaskInfo 查询单个异步任务详情。 // AsyncTaskInfo queries a single async task detail.
function AsyncTaskInfo(id, callBack) {
  base.BasePost('/api/AsyncTaskInfo', {
    id: id,
  }, callBack)
}

// AsyncTaskAction 执行异步任务确认或丢弃操作。 // AsyncTaskAction applies confirm/discard actions to an async task.
function AsyncTaskAction(id, action, callBack) {
  base.BasePost('/api/AsyncTaskAction', {
    id: id,
    action: action,
  }, callBack)
}

// AsyncTaskDelete 删除异步任务记录。 // AsyncTaskDelete deletes an async task record.
function AsyncTaskDelete(id, callBack) {
  base.BasePost('/api/AsyncTaskDelete', {
    id: id,
  }, callBack)
}

export default {
  AsyncTaskList,
  AsyncTaskInfo,
  AsyncTaskAction,
  AsyncTaskDelete,
}
