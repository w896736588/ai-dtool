const assert = require('assert')
const fs = require('fs')
const path = require('path')

const filePath = path.join(__dirname, '../src/components/memory/MemoryEditor.vue')
const source = fs.readFileSync(filePath, 'utf8')

assert.match(
  source,
  /\.preview-renderer\s*\{[\s\S]*overflow:\s*auto;[\s\S]*scrollbar-width:\s*thin;/,
  '预览正文区域应声明可见滚动条样式'
)

assert.match(
  source,
  /\.preview-outline-card\s*\{[\s\S]*overflow:\s*auto;[\s\S]*scrollbar-width:\s*thin;/,
  '目录区域应声明可见滚动条样式'
)

assert.match(
  source,
  /\.preview-renderer::-webkit-scrollbar[\s\S]*\.preview-outline-card::-webkit-scrollbar/,
  '预览正文和目录区域都应提供 webkit 滚动条样式'
)

console.log('memory_editor_scrollbar tests passed')
