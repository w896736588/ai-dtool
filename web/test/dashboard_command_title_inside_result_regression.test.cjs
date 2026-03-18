const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

assert.ok(
  dashboardSource.includes('class="message-content command-result-content"'),
  'Dashboard 命令结果应使用统一结果框容器'
)

assert.ok(
  dashboardSource.includes('class="message-result-command">{{ msg.commandText }}</div>'),
  'Dashboard 应将“执行命令: xxx”渲染到结果框顶部'
)

assert.ok(
  dashboardSource.includes('class="message-result-body"'),
  'Dashboard 应将进行中/完成文本渲染在命令标题下方'
)

assert.ok(
  !dashboardSource.includes('v-if="msg.processText" class="message-command"'),
  'Dashboard 不应在执行过程存在时额外渲染一层重复的命令标题'
)

assert.ok(
  dashboardSource.includes('.message-result-command {'),
  'Dashboard 应为结果框内的命令标题提供独立样式'
)

console.log('dashboard_command_title_inside_result_regression tests passed')
