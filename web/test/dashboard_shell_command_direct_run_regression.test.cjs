const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const configPath = path.join(__dirname, '../src/config/commandConfig.js')
const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const configSource = fs.readFileSync(configPath, 'utf8')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

assert.ok(
  configSource.includes("command: 'shell'"),
  'Dashboard 应保留 shell 一级命令'
)

assert.ok(
  configSource.includes("action: 'shell'") &&
  configSource.includes("dynamicChildren: 'shellOutList'"),
  'shell 一级命令下应直接加载任务列表'
)

assert.ok(
  !configSource.includes("action: 'shellCreate'") &&
  !configSource.includes("action: 'shellList'") &&
  !configSource.includes("action: 'shellRun'"),
  'shell 命令配置中不应再保留创建、任务列表、运行任务三层中转命令'
)

assert.ok(
  dashboardSource.includes("case 'shell':"),
  'Dashboard 应支持 shell 一级命令直接执行'
)

assert.ok(
  dashboardSource.includes("const actionIndex = stack.findIndex(item => item.action === 'shell')"),
  'Dashboard 应按 shell 一级命令后的直接任务项处理'
)

assert.ok(
  dashboardSource.includes("appendOutputResult(`正在打开终端输出任务 [${target.name || target.id}]...\\n`)"),
  'Dashboard 在执行 shell 命令后应先写入进行中提示'
)

assert.ok(
  dashboardSource.includes("appendOutputSummary(`已打开终端输出任务 [${target.name || target.id}]`)"),
  'Dashboard 在执行 shell 命令后应将完成状态追加到同一个结果框'
)

assert.ok(
  !dashboardSource.includes("item.action === 'shellList'") &&
  !dashboardSource.includes("item.action === 'shellRun'") &&
  !dashboardSource.includes("item.action === 'shellCreate'"),
  'Dashboard 不应再依赖 shellCreate、shellList 或 shellRun 中转 action'
)

assert.ok(
  dashboardSource.includes("/#/fullpage?group_id=") &&
  dashboardSource.includes("window.open(url, '_blank')"),
  'Dashboard 选择 shell 任务后应复用“新窗口”按钮行为，打开 fullpage 窗口'
)

assert.ok(
  !dashboardSource.includes("shell.ShellOutSetSeeId({"),
  'Dashboard 不应在首页直接调用 ShellOutSetSeeId 运行 shell 任务'
)

console.log('dashboard_shell_command_direct_run_regression tests passed')
