const assert = require('assert')
const fs = require('fs')
const path = require('path')

const filePath = path.join(__dirname, '../src/components/MemoryFragment.vue')
const source = fs.readFileSync(filePath, 'utf8')
const mainBlock = source.match(/\.sidebar-item-main\s*\{([\s\S]*?)\n\}/)
const titleBlock = source.match(/\.sidebar-item-title\s*\{([\s\S]*?)\n\}/)
const metaBlock = source.match(/\.sidebar-item-meta\s*\{([\s\S]*?)\n\}/)
const timeBlock = source.match(/\.sidebar-item-time\s*\{([\s\S]*?)\n\}/)
const copyBlock = source.match(/\.sidebar-item-copy\s*\{([\s\S]*?)\n\}/)

assert.ok(mainBlock, '应存在 sidebar-item-main 样式块')
assert.ok(titleBlock, '应存在 sidebar-item-title 样式块')
assert.ok(metaBlock, '应存在 sidebar-item-meta 样式块')
assert.ok(timeBlock, '应存在 sidebar-item-time 样式块')
assert.ok(copyBlock, '应存在 sidebar-item-copy 样式块')

assert.match(
  mainBlock[1],
  /flex-direction:\s*column;/,
  '知识片段标题区应改为纵向排布，避免时间占用标题第一行宽度'
)

assert.match(
  titleBlock[1],
  /display:\s*block;/,
  '知识片段标题应作为块级区域占满剩余宽度'
)

assert.match(
  titleBlock[1],
  /width:\s*100%;/,
  '知识片段标题应明确使用完整可用宽度后再换行'
)

assert.match(
  metaBlock[1],
  /justify-content:\s*space-between;/,
  '所属位置元信息行应左右分布复制入口和更新时间'
)

assert.match(
  source,
  /@click\.stop="copyFragmentPath\(item\.file_path\)"/,
  '所属位置左侧应提供复制地址入口'
)

assert.match(
  source,
  /async\s+copyFragmentPath\(filePath\)\s*\{/,
  'MemoryFragment 应提供复制所属位置的方法'
)

assert.match(
  timeBlock[1],
  /margin-left:\s*auto;/,
  '更新时间应固定在所属位置行右侧'
)

assert.match(
  copyBlock[1],
  /cursor:\s*pointer;/,
  '复制地址入口应具备可点击样式'
)

console.log('memory_fragment_sidebar_title_layout tests passed')
