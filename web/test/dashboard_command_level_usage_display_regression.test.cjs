const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

assert.ok(
  dashboardSource.includes('v-if="!cmd.insertOnly && getCommandLevelUsageCount(cmd, filteredCommands) > 0"'),
  'Dashboard 普通命令候选应显示当前层级的本地操作次数'
)

assert.ok(
  dashboardSource.includes('class="command-usage-count"'),
  'Dashboard 命令候选应提供独立的次数展示节点'
)

assert.ok(
  dashboardSource.includes('.command-usage-count {'),
  'Dashboard 应为命令操作次数提供独立样式'
)

console.log('dashboard_command_level_usage_display_regression tests passed')
