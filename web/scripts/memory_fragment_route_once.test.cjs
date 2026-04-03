const assert = require('assert')
const fs = require('fs')
const path = require('path')

const filePath = path.join(__dirname, '../src/components/MemoryFragment.vue')
const source = fs.readFileSync(filePath, 'utf8')

assert.match(
  source,
  /routeFragmentHandled:\s*false/,
  'MemoryFragment 应记录路由指定片段是否已经消费过'
)

assert.match(
  source,
  /if\s*\(\s*needReloadLists\s*\)\s*\{[\s\S]*?\}[\s\S]*?this\.tryOpenRouteFragmentOnEntry\(\)/,
  '初始化状态加载后应通过一次性入口方法处理路由片段'
)

assert.doesNotMatch(
  source,
  /if\s*\(\s*this\.routeFragmentId\s*>\s*0\s*\)\s*\{[\s\S]*?this\.openRouteFragment\(\)[\s\S]*?\}/,
  '轮询状态刷新不应继续无条件强制打开路由指定片段'
)

assert.match(
  source,
  /tryOpenRouteFragmentOnEntry\(\)\s*\{[\s\S]*?this\.routeFragmentHandled\s*=\s*true[\s\S]*?this\.openFragment\(fragmentId\)/,
  '一次性入口方法应在首次消费后标记 handled，并打开目标片段'
)

console.log('memory_fragment_route_once tests passed')
