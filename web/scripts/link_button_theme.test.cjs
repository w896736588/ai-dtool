const assert = require('assert')
const fs = require('fs')
const path = require('path')

const linkVuePath = path.join(__dirname, '../src/components/Link.vue')
const source = fs.readFileSync(linkVuePath, 'utf8')

const run = () => {
  assert.ok(
    source.includes('<div class="link-module-shell">'),
    '自定义网页模块应提供统一外层容器，让 scoped deep 样式稳定命中子页面按钮'
  )

  assert.ok(
    source.includes('.link-module-shell :deep(.git-action-button--primary)') &&
    source.includes('.link-module-shell :deep(.pl-button--primary)') &&
    source.includes('background: linear-gradient(180deg, #5a8a5a 0%, #4f804f 100%) !important;') &&
    source.includes('color: #ffffff !important;'),
    '自定义网页主按钮应统一为 Redis 搜索区同系深绿色白字'
  )

  assert.ok(
    source.includes('.link-module-shell :deep(.git-action-button--warning)') &&
    source.includes('.link-module-shell :deep(.git-action-button--info)') &&
    source.includes('.link-module-shell :deep(.pl-button--warning)') &&
    source.includes('.link-module-shell :deep(.pl-button--info)'),
    '自定义网页非删除按钮都应归并到主绿色体系'
  )

  assert.ok(
    source.includes('background: linear-gradient(180deg, #de6f6f 0%, #d65c5c 100%) !important;'),
    '自定义网页删除按钮应使用红色'
  )

  console.log('link_button_theme tests passed')
}

run()
