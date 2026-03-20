const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

assert.ok(
  !dashboardSource.includes('@click="selectCommand(cmd)"'),
  'Dashboard 命令候选不应再支持鼠标点击直接选择，避免鼠标位置干扰当前键盘选中项'
)

assert.ok(
  !dashboardSource.includes('@mouseenter="activeCommandIndex = index"'),
  'Dashboard 命令候选不应再在鼠标移入时改写 activeCommandIndex'
)

assert.ok(
  !dashboardSource.includes('.command-item:hover {'),
  'Dashboard 命令候选不应再保留 hover 高亮样式'
)

console.log('dashboard_command_mouse_selection_removed_regression tests passed')
