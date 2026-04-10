const assert = require('assert')
const fs = require('fs')
const path = require('path')

const files = [
  path.join(__dirname, '../src/components/MemoryFragment.vue'),
  path.join(__dirname, '../src/components/memory/MemoryWelcome.vue'),
  path.join(__dirname, '../src/components/memory/MemoryEditor.vue'),
]

for (const filePath of files) {
  const source = fs.readFileSync(filePath, 'utf8')
  assert.doesNotMatch(
    source,
    /index_status_desc|index_status|索引成功|索引失败|待索引/,
    `${path.basename(filePath)} 不应继续依赖知识片段索引状态字段或文案`
  )
}

console.log('memory_fragment_remove_index_status tests passed')
