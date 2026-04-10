const assert = require('assert')
const fs = require('fs')
const path = require('path')

const homeSource = fs.readFileSync(path.join(__dirname, '../src/components/Home.vue'), 'utf8')

assert.match(
  homeSource,
  /require\(['"]@\/utils\/home_task_fragment_options\.cjs['"]\)/,
  'Home.vue 应通过 .cjs 工具模块加载任务知识片段选项工具，避免运行时模块互操作异常'
)

console.log('home_task_fragment_options require path tests passed')
