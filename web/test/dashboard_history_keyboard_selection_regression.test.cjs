const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

assert.ok(
  dashboardSource.includes('const loadHistoryCommandForExecution = (historyCommand, options = {}) => {'),
  'Dashboard 应提供统一的历史命令加载入口，区分“加载为可执行状态”和“立即自动执行”'
)

assert.ok(
  dashboardSource.includes('const autoExecute = options.autoExecute === true'),
  'Dashboard 历史命令加载入口应显式控制是否自动执行，避免方向键回填时沿用自动执行逻辑'
)

assert.ok(
  dashboardSource.includes('markPendingHistoryExecution(commandText, { autoExecute })'),
  'Dashboard 历史命令回填后应保留待解析状态，等待异步动态候选加载完成后再收敛到最终可执行态'
)

assert.ok(
  dashboardSource.includes("reparseForPendingHistoryExecution('gitGroupList')"),
  'Dashboard Git 分组列表异步返回后应重新解析历史命令，避免完整命令仍被误判为待选择'
)

assert.ok(
  dashboardSource.includes('loadHistoryCommandForExecution(commandHistory.value[nextIndex])'),
  'Dashboard 上下方向键浏览历史命令时，应复用历史命令加载入口而不是直接 parseInput 导致候选重新弹出'
)

assert.ok(
  dashboardSource.includes('const shouldAutoExecute = pendingHistoryExecution.value.autoExecute') &&
  dashboardSource.includes('if (shouldAutoExecute) {'),
  'Dashboard 历史命令补齐后应区分“自动执行”和“仅进入可执行态”两种收敛路径'
)

assert.ok(
  dashboardSource.includes('const suppressDropdownOnNextFocus = ref(false)'),
  'Dashboard 应在历史命令回填后记录一次性 focus 抑制标记，避免重新聚焦时再次打开候选'
)

assert.ok(
  dashboardSource.includes('suppressDropdownOnNextFocus.value = true'),
  'Dashboard 历史命令回填后应抑制下一次 focus 引发的二次解析'
)

assert.ok(
  dashboardSource.includes('if (suppressDropdownOnNextFocus.value) {'),
  'Dashboard handleFocus 应识别历史命令回填后的 focus 抑制状态'
)

console.log('dashboard_history_keyboard_selection_regression tests passed')
