const assert = require('assert')

const {
  LOCATOR_AUTO_EXTRACT_SYSTEM_PROMPT,
} = require('../src/utils/smart_link_locator_ai.cjs')

const run = () => {
  assert.ok(LOCATOR_AUTO_EXTRACT_SYSTEM_PROMPT.includes('Playwright'), '提示词应明确说明 Playwright 场景')
  assert.ok(LOCATOR_AUTO_EXTRACT_SYSTEM_PROMPT.includes('first'), '提示词应包含 first 提取说明')
  assert.ok(LOCATOR_AUTO_EXTRACT_SYSTEM_PROMPT.includes('last'), '提示词应包含 last 提取说明')
  assert.ok(LOCATOR_AUTO_EXTRACT_SYSTEM_PROMPT.includes('nth'), '提示词应包含 nth 提取说明')
  console.log('smart_link_locator_ai tests passed')
}

run()
