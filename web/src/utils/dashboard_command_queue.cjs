// 队列项默认前缀，用于区分首页命令待执行项。
const PENDING_COMMAND_ID_PREFIX = 'pending-command'

// 创建首页命令待执行项，保留原始命令文本用于后续重新解析执行。
function createPendingCommandItem(rawCommand, createdAt = Date.now()) {
  const normalizedCommand = String(rawCommand || '').trim()
  return {
    id: `${PENDING_COMMAND_ID_PREFIX}-${createdAt}-${Math.random().toString(36).slice(2, 8)}`,
    rawCommand: normalizedCommand,
    createdAt,
  }
}

// 入队时始终返回新数组，避免直接修改响应式源数据。
function enqueuePendingCommand(queue, item) {
  const queueList = Array.isArray(queue) ? queue : []
  return [...queueList, item]
}

// 出队时返回队首元素和剩余队列，便于页面层统一调度。
function dequeuePendingCommand(queue) {
  const queueList = Array.isArray(queue) ? queue : []
  if (queueList.length === 0) {
    return {
      item: null,
      queue: [],
    }
  }
  return {
    item: queueList[0],
    queue: queueList.slice(1),
  }
}

// 删除待执行项时仅移除目标 id，保留其余顺序不变。
function removePendingCommandById(queue, targetId) {
  const queueList = Array.isArray(queue) ? queue : []
  const normalizedTargetId = String(targetId || '').trim()
  if (!normalizedTargetId) {
    return queueList
  }
  return queueList.filter(item => item && item.id !== normalizedTargetId)
}

// 消费队首待执行命令，并立即交给执行回调处理。
function consumeNextPendingCommand(queue, executor) {
  const dequeueResult = dequeuePendingCommand(queue)
  if (dequeueResult.item && typeof executor === 'function') {
    executor(dequeueResult.item.rawCommand, dequeueResult.item)
  }
  return dequeueResult
}

module.exports = {
  PENDING_COMMAND_ID_PREFIX,
  createPendingCommandItem,
  enqueuePendingCommand,
  dequeuePendingCommand,
  removePendingCommandById,
  consumeNextPendingCommand,
}
