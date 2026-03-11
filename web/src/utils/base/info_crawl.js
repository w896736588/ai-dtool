import base from '../base'

// InfoCrawlCrawl4AIStatus 查询 Crawl4AI 状态。
function InfoCrawlCrawl4AIStatus(callBack) {
  base.BasePost('/api/InfoCrawlCrawl4AIStatus', {}, callBack)
}

// InfoCrawlTaskList 查询任务列表。
function InfoCrawlTaskList(callBack) {
  base.BasePost('/api/InfoCrawlTaskList', {}, callBack)
}

// InfoCrawlTaskInfo 查询任务详情。
function InfoCrawlTaskInfo(id, callBack) {
  base.BasePost('/api/InfoCrawlTaskInfo', { id: id }, callBack)
}

// InfoCrawlTaskSave 保存任务。
function InfoCrawlTaskSave(data, callBack) {
  base.BasePost('/api/InfoCrawlTaskSave', data, callBack)
}

// InfoCrawlTaskDelete 删除任务。
function InfoCrawlTaskDelete(id, callBack) {
  base.BasePost('/api/InfoCrawlTaskDelete', { id: id }, callBack)
}

// InfoCrawlTaskRun 执行任务。
function InfoCrawlTaskRun(data, callBack) {
  base.BasePost('/api/InfoCrawlTaskRun', data, callBack)
}

// InfoCrawlRunList 查询执行历史。
function InfoCrawlRunList(taskId, limit, callBack) {
  base.BasePost('/api/InfoCrawlRunList', { task_id: taskId, limit: limit }, callBack)
}

// InfoCrawlRunInfo 查询执行详情。
function InfoCrawlRunInfo(id, callBack) {
  base.BasePost('/api/InfoCrawlRunInfo', { id: id }, callBack)
}

export default {
  InfoCrawlCrawl4AIStatus,
  InfoCrawlTaskList,
  InfoCrawlTaskInfo,
  InfoCrawlTaskSave,
  InfoCrawlTaskDelete,
  InfoCrawlTaskRun,
  InfoCrawlRunList,
  InfoCrawlRunInfo,
}
