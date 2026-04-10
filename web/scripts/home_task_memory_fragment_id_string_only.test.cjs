const assert = require('assert')
const fs = require('fs')
const path = require('path')

const source = fs.readFileSync(path.join(__dirname, '../src/components/Home.vue'), 'utf8')

assert.doesNotMatch(
  source,
  /Number\(task\.memory_fragment_id\s*\|\|\s*0\)/,
  'Home.vue 不应继续将 memory_fragment_id 按数字解析（task 维度）'
)

assert.doesNotMatch(
  source,
  /Number\(this\.homeTaskForm\.memory_fragment_id\s*\|\|\s*0\)/,
  'Home.vue 不应继续将 memory_fragment_id 按数字解析（form 维度）'
)

console.log('home task memory fragment id string-only tests passed')
