const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

assert.ok(
  !dashboardSource.includes('v-if="msg.summaryText" class="message-summary"'),
  'Dashboard 命令结果不应再单独渲染摘要区，进行中与完成状态应同属一个结果框'
)

assert.ok(
  dashboardSource.includes('summaryText: \'\','),
  'Dashboard 创建命令输出消息时应初始化 summaryText'
)

assert.ok(
  dashboardSource.includes('const appendOutputSummary = (text) => {'),
  'Dashboard 应提供单独写入状态摘要的方法'
)

assert.ok(
  dashboardSource.includes("const current = String(currentOutputMessage.value.resultText || '')") &&
  dashboardSource.includes("currentOutputMessage.value.resultText = merged.length > 50000 ? merged.slice(-50000) : merged"),
  'Dashboard 应将完成状态追加回同一个结果框'
)

assert.ok(
  dashboardSource.includes('.message-command {') &&
  dashboardSource.includes('font-size: 14px;'),
  'Dashboard 命令标题应使用更明显的字号'
)

console.log('dashboard_command_result_layout_regression tests passed')
