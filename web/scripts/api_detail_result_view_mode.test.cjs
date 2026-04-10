const assert = require('assert')
const fs = require('fs')
const path = require('path')

const apiDetailPath = path.join(__dirname, '../src/components/api/ApiDetail.vue')
const source = fs.readFileSync(apiDetailPath, 'utf8')

const run = () => {
  assert.ok(
    /<el-radio-button\s+value="auto">自动<\/el-radio-button>/.test(source) &&
      /<el-radio-button\s+value="json">JSON<\/el-radio-button>/.test(source) &&
      /<el-radio-button\s+value="html">HTML预览<\/el-radio-button>/.test(source) &&
      /<el-radio-button\s+value="raw">原始<\/el-radio-button>/.test(source),
    'Result view mode buttons (auto/json/html/raw) should exist'
  )

  assert.ok(
    /<iframe\s+v-else-if="resolvedResponseViewMode === 'html'"[\s\S]*class="response-html-preview"[\s\S]*:srcdoc="sanitizedResponseHtml"[\s\S]*sandbox=""/s.test(source),
    'HTML preview iframe should render with srcdoc and strict sandbox'
  )

  assert.ok(
    /responseViewMode:\s*'auto'/.test(source) &&
      /resolvedResponseViewMode\(\)/.test(source) &&
      /sanitizedResponseHtml\(\)/.test(source),
    'Response view mode state and computed fields should be defined'
  )

  assert.ok(
    /const bodyIsJson = this\.isJsonResponse\(body\)/.test(source) &&
      /const bodyIsHtml = this\.isHtmlResponse\(body\)/.test(source) &&
      /contentType\.includes\('application\/json'\)[\s\S]*&&[\s\S]*bodyIsJson/.test(source) &&
      /contentType\.includes\('text\/html'\)[\s\S]*&&[\s\S]*bodyIsHtml/.test(source),
    'Auto mode should not trust Content-Type alone; it must verify body shape before deciding'
  )

  assert.ok(
    /sanitizeHtmlForPreview\(html\)\s*\{/.test(source) &&
      /forbiddenSelectors\s*=\s*\[/.test(source) &&
      /attrName\.startsWith\('on'\)/.test(source),
    'HTML preview should sanitize forbidden tags and inline event handlers'
  )

  console.log('api_detail_result_view_mode tests passed')
}

run()
