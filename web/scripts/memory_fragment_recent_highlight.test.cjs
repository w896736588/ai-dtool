const assert = require('assert')
const fs = require('fs')
const path = require('path')

const filePath = path.join(__dirname, '../src/components/MemoryFragment.vue')
const source = fs.readFileSync(filePath, 'utf8')

assert.doesNotMatch(
  source,
  /<el-tab-pane\s+name="home"/,
  '知识片段页不应继续保留首页标签页'
)

assert.match(
  source,
  /activeTab:\s*''/,
  '知识片段页默认激活标签应为空，等待自动打开第一个片段'
)

assert.match(
  source,
  /ensureDefaultFragmentTab\(\)\s*\{/,
  'MemoryFragment 应提供默认打开首个知识片段的方法'
)

assert.match(
  source,
  /this\.ensureDefaultFragmentTab\(\)/,
  '列表加载或回退场景应调用默认打开首个知识片段逻辑'
)

assert.match(
  source,
  /fragmentFreshnessClass\(item\)\s*\{/,
  'MemoryFragment 应根据更新时间返回侧边栏项的新鲜度样式类'
)

assert.match(
  source,
  /is-updated-today|is-updated-3d|is-updated-7d|is-updated-older/,
  '左侧知识片段应定义按更新时间分层的样式类'
)

console.log('memory_fragment_recent_highlight tests passed')
