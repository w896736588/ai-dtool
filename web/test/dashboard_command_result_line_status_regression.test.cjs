const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

assert.ok(
  dashboardSource.includes('v-for="(line, lineIndex) in getResultLines(msg.resultText)"'),
  'Dashboard 应逐行渲染结果文本，便于给进行中/完成态单独加标记'
)

assert.ok(
  dashboardSource.includes("getResultLineState(line, lineIndex, getResultLines(msg.resultText)) === 'running'") &&
  dashboardSource.includes('result-line-dots'),
  'Dashboard 应为“正在...”结果行渲染执行中动画'
)

assert.ok(
  dashboardSource.includes("getResultLineState(line, lineIndex, getResultLines(msg.resultText)) === 'success'") &&
  dashboardSource.includes('result-line-check'),
  'Dashboard 应为完成结果行渲染对勾标记'
)

assert.ok(
  dashboardSource.includes('const hasTerminalLineAfter = sourceLines.slice(lineIndex + 1).some(item => {'),
  'Dashboard 应在后续已有完成/失败行时移除前面“正在...”行的执行动画'
)

console.log('dashboard_command_result_line_status_regression tests passed')
