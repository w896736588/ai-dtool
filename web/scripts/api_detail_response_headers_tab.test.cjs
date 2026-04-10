const assert = require('assert')
const fs = require('fs')
const path = require('path')

const filePath = path.join(__dirname, '../src/components/api/ApiDetail.vue')
const source = fs.readFileSync(filePath, 'utf8')

const run = () => {
  assert.ok(
    /<el-tab-pane label="请求头" name="headers">[\s\S]*<key-value-view :data="requestHeadersData" \/>/s.test(source),
    'Request headers tab should bind to requestHeadersData'
  )

  assert.ok(
    /<el-tab-pane label="返回头" name="responseHeaders">[\s\S]*<key-value-view :data="responseHeadersData" \/>/s.test(source),
    'Response headers tab should exist next to request headers and bind to responseHeadersData'
  )

  assert.ok(
    /resolvedResponseViewMode\(\)\s*\{[\s\S]*this\.responseHeadersData/s.test(source),
    'Auto view mode should infer content type from response headers'
  )

  console.log('api_detail_response_headers_tab tests passed')
}

run()