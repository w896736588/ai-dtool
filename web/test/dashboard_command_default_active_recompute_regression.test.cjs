const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

assert.ok(
  dashboardSource.includes('watch(filteredCommands, (commandList) => {'),
  'Dashboard 候选列表内容变化后应重新计算默认选中项，避免异步列表仍停留在旧索引'
)

assert.ok(
  dashboardSource.includes('activeCommandIndex.value = getDefaultActiveCommandIndex(commandList)'),
  'Dashboard 候选列表重排后应按最新使用次数重新设置默认选中项'
)

console.log('dashboard_command_default_active_recompute_regression tests passed')
