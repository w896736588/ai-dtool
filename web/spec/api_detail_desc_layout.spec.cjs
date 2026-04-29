const assert = require('assert')
const fs = require('fs')
const path = require('path')

const rootDir = path.resolve(__dirname, '..')
const vueSource = fs.readFileSync(path.join(rootDir, 'src/components/api/ApiDetail.vue'), 'utf8')
const cssSource = fs.readFileSync(path.join(rootDir, 'src/css/components/api/ApiDetail.css'), 'utf8')
const apiPageCssSource = fs.readFileSync(path.join(rootDir, 'src/css/components/Api.css'), 'utf8')

assert.match(
  vueSource,
  /<el-tabs[^>]*class="detail-tabs api-config-tabs"/,
  '接口配置 tabs 需要独立 class，避免影响其它 detail-tabs'
)

assert.match(
  vueSource,
  /<el-tab-pane[^>]*name="desc"[^>]*class="desc-tab-pane"/,
  '备注 tab 需要独立 class 承接满高布局'
)

assert.match(
  vueSource,
  /<MdEditor[^>]*class="desc-editor"/,
  '备注 MdEditor 需要独立 class 让样式只作用于备注编辑器'
)

assert.match(
  cssSource,
  /\.api-detail\s*\{[^}]*height:\s*100%[^}]*min-height:\s*0/s,
  '接口明细需要填充父容器并允许收缩，不能按整屏高度撑开'
)

assert.doesNotMatch(
  cssSource,
  /\.api-detail\s*\{[^}]*height:\s*100vh/s,
  '接口明细嵌在右侧面板内，不能使用 100vh'
)

assert.doesNotMatch(
  cssSource,
  /\.api-detail\s*\{[^}]*min-height:\s*720px/s,
  '接口明细不能设置 720px 最小高度，否则下半屏会跑出屏幕'
)

assert.match(
  cssSource,
  /\.api-config-tabs\s*\{[^}]*flex:\s*1[^}]*min-height:\s*0[^}]*height:\s*100%/s,
  'api-config-tabs 只能作为普通 flex 子项撑满剩余高度'
)

assert.doesNotMatch(
  cssSource,
  /\.api-config-tabs\s*\{[^}]*display:\s*flex/s,
  'api-config-tabs 不能设置 display:flex，否则 Element Plus 的 tab header 会跑到底部'
)

assert.match(
  cssSource,
  /:deep\(\.api-config-tabs\s*>\s*\.el-tabs__content\)\s*\{[^}]*height:\s*calc\(100%\s*-\s*50px\)[^}]*min-height:\s*0/s,
  'el-tabs 内容区需要撑满剩余高度并留出 tab header 余量'
)

assert.match(
  cssSource,
  /:deep\(\.api-config-tabs\s*>\s*\.el-tabs__content\)\s*\{[^}]*box-sizing:\s*border-box[^}]*width:\s*100%/s,
  'el-tabs 内容区需要 border-box 满宽，padding 不能继续撑爆高度或压缩宽度'
)

assert.doesNotMatch(
  cssSource,
  /:deep\(\.api-config-tabs\s*>\s*\.el-tabs__content\)\s*\{[^}]*display:\s*flex/s,
  'el-tabs 内容区不能设置 display:flex，否则 tab pane 会水平布局导致宽度不撑满'
)

assert.doesNotMatch(
  cssSource,
  /\.detail-tabs\s*\{[^}]*display:\s*flex/s,
  '通用 detail-tabs 不能设置 flex，否则 tab 头可能被挤到底部'
)

assert.match(
  cssSource,
  /:deep\(\.api-config-tabs\s*>\s*\.el-tabs__content\s*>\s*\.el-tab-pane\)\s*\{[^}]*width:\s*100%[^}]*height:\s*100%[^}]*min-height:\s*0/s,
  'api 配置 tab pane 需要满宽满高并允许收缩'
)

assert.match(
  cssSource,
  /\.desc-tab-pane\s*\{[^}]*height:\s*100%[^}]*display:\s*flex[^}]*min-height:\s*0[^}]*box-sizing:\s*border-box/s,
  '备注 tab pane 需要把高度传递给编辑器，并使用 border-box 避免撑爆'
)

assert.match(
  cssSource,
  /\.desc-editor\s*\{[^}]*flex:\s*1\s*1\s*auto[^}]*height:\s*100%[^}]*min-height:\s*0[^}]*width:\s*100%[^}]*max-width:\s*100%/s,
  '备注 MdEditor 需要填满 tab pane 且宽度不能收缩'
)

assert.match(
  cssSource,
  /\.desc-editor\s*:deep\(\.md-editor\)\s*\{[^}]*height:\s*100%[^}]*width:\s*100%[^}]*min-height:\s*0/s,
  'MdEditor 内层根节点也需要继承备注编辑器的满高满宽'
)

assert.match(
  apiPageCssSource,
  /\.panel-content\s*\{[^}]*min-height:\s*0/s,
  '接口明细父级内容区需要允许 flex 子项收缩，避免满高布局溢出'
)

assert.match(
  apiPageCssSource,
  /\.api-settings\s*\{[^}]*height:\s*100%[^}]*min-height:\s*0[^}]*width:\s*100%/s,
  'ApiDetail 的直接父级需要提供明确的满高满宽容器'
)
