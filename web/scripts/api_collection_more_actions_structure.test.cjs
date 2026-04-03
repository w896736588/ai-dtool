const assert = require('assert')
const fs = require('fs')
const path = require('path')

const filePath = path.join(__dirname, '../src/components/Api.vue')
const source = fs.readFileSync(filePath, 'utf8')

const collectionActionsMatch = source.match(
  /<span v-if="data\.type === 'collection'" class="node-actions">([\s\S]*?)<\/span>/
)

assert.ok(collectionActionsMatch, '应存在集合节点的更多操作区域')

const collectionActionsBlock = collectionActionsMatch[1]

assert.match(
  collectionActionsBlock,
  /<el-dropdown>/,
  '集合节点的更多操作应继续使用 el-dropdown'
)

assert.match(
  collectionActionsBlock,
  /class="node-action-trigger"/,
  '集合节点的更多操作按钮应使用专门的小图标触发器样式'
)

assert.doesNotMatch(
  collectionActionsBlock,
  /toggleCollection\(data\)/,
  '集合节点的更多操作区不应再嵌套切换展开按钮，避免 hover 出现异常白块'
)

assert.doesNotMatch(
  collectionActionsBlock,
  /<pl-button[^>]*toggleCollection\(data\)[\s\S]*<el-dropdown>/,
  '集合节点更多操作不应出现按钮嵌套 dropdown 的结构'
)

console.log('api_collection_more_actions_structure tests passed')
