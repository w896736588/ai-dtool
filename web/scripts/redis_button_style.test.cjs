const assert = require('assert')
const fs = require('fs')
const path = require('path')

const redisVuePath = path.join(__dirname, '../src/components/Redis.vue')
const source = fs.readFileSync(redisVuePath, 'utf8')

const run = () => {
  assert.ok(
    source.includes('.settings-btn {') &&
    source.includes('--pl-button-background-color: #f3e6d3;') &&
    source.includes('--pl-button-border-color: #dcc5a6;') &&
    source.includes('--pl-button-text-color: #7a5a33;'),
    'Redis 页面设置按钮应使用更贴整体风格的茶米色主态'
  )

  assert.ok(
    source.includes('.settings-btn:hover {') &&
    source.includes('--pl-button-background-color: #ebdac2;') &&
    source.includes('--pl-button-border-color: #d1b896;') &&
    source.includes('--pl-button-text-color: #6d502d;'),
    'Redis 页面设置按钮 hover 应保持柔和的茶米色反馈'
  )

  assert.ok(
    source.includes('.settings-btn:active {') &&
    source.includes('--pl-button-background-color: #e1ccb0;') &&
    source.includes('--pl-button-border-color: #c6ab87;') &&
    source.includes('--pl-button-text-color: #624726;'),
    'Redis 页面设置按钮 active 应使用略深的茶米色'
  )

  console.log('redis_button_style tests passed')
}

run()
