const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

assert.ok(
  !dashboardSource.includes('.command-item:hover,\r\n.command-item.active {') &&
  !dashboardSource.includes('.command-item:hover,\n.command-item.active {'),
  'Dashboard 命令候选的 hover 与 active 不应共用同一高亮样式，否则鼠标停在首项时会掩盖键盘上下切换'
)

console.log('dashboard_command_keyboard_hover_regression tests passed')
