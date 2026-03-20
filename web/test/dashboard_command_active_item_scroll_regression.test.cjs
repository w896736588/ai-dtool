const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

assert.ok(
  dashboardSource.includes('ref="commandDropdown"'),
  'Dashboard 命令下拉应暴露容器 ref，便于把当前默认选中项滚动到可视区'
)

assert.ok(
  dashboardSource.includes(':ref="(el) => setCommandItemRef(el, index)"'),
  'Dashboard 命令候选应记录每一项的 DOM 引用，便于定位当前选中项'
)

assert.ok(
  dashboardSource.includes('const ensureActiveCommandVisible = () => {') &&
  dashboardSource.includes('activeElement.scrollIntoView({ block: \'nearest\' })'),
  'Dashboard 应在候选较多时自动滚动到当前选中项，避免默认选中项落在可视区外'
)

assert.ok(
  dashboardSource.includes('watch([showCommands, activeCommandIndex, filteredCommands], () => {'),
  'Dashboard 应在下拉打开、选中项变化或候选变化时重新确保当前选中项可见'
)

console.log('dashboard_command_active_item_scroll_regression tests passed')
