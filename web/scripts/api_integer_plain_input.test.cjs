const assert = require('assert')
const fs = require('fs')
const path = require('path')

const filePath = path.join(__dirname, '../src/components/api/KeyValueEditor.vue')
const source = fs.readFileSync(filePath, 'utf8')

assert.match(
  source,
  /<div v-else-if="item\.type === 'integer'">[\s\S]*?<el-input[\s\S]*?<\/div>/,
  'integer 类型应使用普通 el-input 输入框'
)

assert.doesNotMatch(
  source,
  /<div v-else-if="item\.type === 'integer'">[\s\S]*?<el-input-number[\s\S]*?<\/div>/,
  'integer 类型不应继续使用 el-input-number'
)

console.log('api_integer_plain_input tests passed')
