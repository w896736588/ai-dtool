const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

assert.ok(
  dashboardSource.includes("const commandLevelUsageMap = ref({})"),
  'Dashboard 应为子命令层级使用次数维护独立状态，不能复用整条历史命令统计'
)

assert.ok(
  dashboardSource.includes("const commandLevelUsageCacheKey = 'dashboard_command_level_usage_v1'"),
  'Dashboard 应为子命令层级使用次数使用独立本地缓存 key'
)

assert.ok(
  dashboardSource.includes('const recordCommandLevelUsage = (stack) => {'),
  'Dashboard 应提供独立的层级使用次数记录入口'
)

assert.ok(
  dashboardSource.includes('recordCommandLevelUsage(currentStack)'),
  'Dashboard 执行首页命令时应按当前命令栈逐级记录子命令使用次数'
)

assert.ok(
  dashboardSource.includes('const getCommandLevelUsageCount = (cmd, commandList = currentChildren.value) => {'),
  'Dashboard 应按当前候选层级读取子命令使用次数'
)

assert.ok(
  dashboardSource.includes('return a.count - b.count') &&
  dashboardSource.includes('usage-enriched'),
  'Dashboard 候选列表应按子命令使用次数升序排序，使使用最多的项位于最下面'
)

assert.ok(
  dashboardSource.includes('const getDefaultActiveCommandIndex = (commandList = filteredCommands.value) => {'),
  'Dashboard 应提供默认高亮项计算逻辑'
)

assert.ok(
  dashboardSource.includes('activeCommandIndex.value = getDefaultActiveCommandIndex()'),
  'Dashboard 弹出候选列表时应默认选中使用次数最多的那一项'
)

console.log('dashboard_command_level_usage_regression tests passed')
