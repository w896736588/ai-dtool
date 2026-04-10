const assert = require('assert')
const fs = require('fs')
const path = require('path')

const homeSource = fs.readFileSync(path.join(__dirname, '../src/components/Home.vue'), 'utf8')

assert.match(
  homeSource,
  /已关联知识片段/,
  'Home.vue 应在任务状态区域展示“已关联知识片段”标签文案'
)

assert.match(
  homeSource,
  /class="home-task-memory-link-tag"[\s\S]*@click(?:\.stop)?="openHomeTaskMemoryFragment\(task\)"/,
  '关联知识片段标签应支持点击并复用 openHomeTaskMemoryFragment(task)'
)

assert.doesNotMatch(
  homeSource,
  /编辑知识片段/,
  '任务卡片右侧不应再显示“编辑知识片段”按钮'
)

console.log('home task memory fragment tag tests passed')
