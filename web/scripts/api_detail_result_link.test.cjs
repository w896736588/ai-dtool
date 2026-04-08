const assert = require('assert')
const fs = require('fs')
const path = require('path')

const apiDetailPath = path.join(__dirname, '../src/components/api/ApiDetail.vue')
const source = fs.readFileSync(apiDetailPath, 'utf8')

const run = () => {
  assert.ok(
    /<div class="request-url-main">\s*<div class="request-url-text">{{ apiForm\.method }} {{ apiForm\.last_result_data\.url }}<\/div>\s*<pl-button[^>]*class="request-url-copy-btn"/s.test(source),
    'Result request address should render as plain text with the copy icon immediately next to it'
  )

  assert.ok(
    /<pl-button[^>]*class="request-url-copy-btn"[^>]*link[^>]*@click="copyUrl\(apiForm\.last_result_data\.url\)"[^>]*>/s.test(source),
    'Result request address should have a dedicated copy icon button'
  )

  assert.ok(
    /<pl-button[^>]*class="request-run-btn"[^>]*type="success"[^>]*@click="handleExecute"[^>]*>\s*<el-icon><VideoPlay \/><\/el-icon>\s*执行\s*<\/pl-button>/s.test(source),
    'Execute action should move next to the request address and render with icon styling'
  )

  assert.ok(
    !/:href="apiForm\.last_result_data\.url"/.test(source),
    'Request address should no longer be rendered as a clickable link'
  )

  console.log('api_detail_result_link tests passed')
}

run()
