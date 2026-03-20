const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const configPath = path.join(__dirname, '../src/config/commandConfig.js')
const configSource = fs.readFileSync(configPath, 'utf8')

assert.ok(
  !configSource.includes("command: 'script'"),
  '首页命令配置中不应再暴露 script 顶级命令'
)

assert.ok(
  !configSource.includes("action: 'scriptRun'") &&
  !configSource.includes("dynamicChildren: 'scriptList'"),
  '首页命令配置中不应再保留 script 脚本命令入口配置'
)

console.log('dashboard_script_command_removed_regression tests passed')
