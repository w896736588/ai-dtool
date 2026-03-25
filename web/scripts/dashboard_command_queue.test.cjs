const assert = require('assert')

const MODULE_PATH = '../src/utils/dashboard_command_queue.cjs'

const loadQueueModule = () => require(MODULE_PATH)

const run = () => {
  const queueUtils = loadQueueModule()
  const {
    createPendingCommandItem,
    enqueuePendingCommand,
    dequeuePendingCommand,
    removePendingCommandById,
    consumeNextPendingCommand,
  } = queueUtils

  const firstItem = createPendingCommandItem('git status', 1001)
  assert.strictEqual(firstItem.rawCommand, 'git status')
  assert.strictEqual(firstItem.createdAt, 1001)
  assert.ok(firstItem.id, '队列项需要生成唯一 id')

  const secondItem = createPendingCommandItem('docker status', 1002)
  const thirdItem = createPendingCommandItem('git pull', 1003)

  const queuedList = enqueuePendingCommand([], firstItem)
  const queuedList2 = enqueuePendingCommand(queuedList, secondItem)
  const queuedList3 = enqueuePendingCommand(queuedList2, thirdItem)

  assert.deepStrictEqual(
    queuedList3.map(item => item.rawCommand),
    ['git status', 'docker status', 'git pull']
  )

  const removedList = removePendingCommandById(queuedList3, secondItem.id)
  assert.deepStrictEqual(
    removedList.map(item => item.rawCommand),
    ['git status', 'git pull']
  )

  const firstDequeue = dequeuePendingCommand(removedList)
  assert.strictEqual(firstDequeue.item.rawCommand, 'git status')
  assert.deepStrictEqual(
    firstDequeue.queue.map(item => item.rawCommand),
    ['git pull']
  )

  const secondDequeue = dequeuePendingCommand(firstDequeue.queue)
  assert.strictEqual(secondDequeue.item.rawCommand, 'git pull')
  assert.deepStrictEqual(secondDequeue.queue, [])

  const executionCalls = []
  const consumedResult = consumeNextPendingCommand(queuedList3, (rawCommand) => {
    executionCalls.push(rawCommand)
  })
  assert.deepStrictEqual(executionCalls, ['git status'])
  assert.deepStrictEqual(
    consumedResult.queue.map(item => item.rawCommand),
    ['docker status', 'git pull']
  )

  console.log('dashboard_command_queue tests passed')
}

run()
