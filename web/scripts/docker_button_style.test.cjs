const assert = require('assert')
const fs = require('fs')
const path = require('path')

const dockerVuePath = path.join(__dirname, '../src/components/Docker.vue')
const source = fs.readFileSync(dockerVuePath, 'utf8')

const run = () => {
  assert.ok(
    source.includes("import { RefreshRight, Search, Setting, View, Document, VideoPlay, VideoPause, Delete } from '@element-plus/icons-vue';"),
    'Docker page should import the expected action button icons'
  )

  assert.ok(
    /<pl-button[^>]*class="operation-btn operation-btn-primary"[^>]*>\s*<el-icon><View \/><\/el-icon>服务列表<\/pl-button>/s.test(source),
    'Service list button should use the primary operation button style'
  )

  assert.ok(
    /<pl-button[^>]*class="operation-btn operation-btn-primary"[^>]*>\s*<el-icon><Document \/><\/el-icon>运行状态<\/pl-button>/s.test(source),
    'Runtime status button should use the primary operation button style'
  )

  assert.ok(
    /<pl-button[^>]*class="operation-btn operation-btn-success"[^>]*>\s*<el-icon><VideoPlay \/><\/el-icon>启动（up -d）<\/pl-button>/s.test(source),
    'Start button should use the success operation button style'
  )

  assert.ok(
    /<pl-button[^>]*class="operation-btn operation-btn-success"[^>]*>\s*<el-icon><RefreshRight \/><\/el-icon>重启（restart）<\/pl-button>/s.test(source),
    'Restart button should use the success operation button style'
  )

  assert.ok(
    /<pl-button[^>]*class="operation-btn operation-btn-danger"[^>]*>\s*<el-icon><VideoPause \/><\/el-icon>停止\(stop\)<\/pl-button>/s.test(source),
    'Stop button should use the danger operation button style'
  )

  assert.ok(
    /<pl-button[^>]*class="operation-btn operation-btn-primary"[^>]*>\s*<el-icon><Document \/><\/el-icon>查看compose\.yml<\/pl-button>/s.test(source),
    'View compose button should use the primary operation button style'
  )

  assert.ok(
    /<pl-button[^>]*class="operation-btn operation-btn-primary"[^>]*>\s*<el-icon><Document \/><\/el-icon>查看env<\/pl-button>/s.test(source),
    'View env button should use the primary operation button style'
  )

  assert.ok(
    source.includes('--pl-button-text-color: #ffffff;') &&
    source.includes('--pl-button-hover-text-color: #ffffff;'),
    'Docker primary operation buttons should stay white text'
  )

  assert.ok(
    source.includes('--pl-button-background-color: linear-gradient(180deg, #71af8d 0%, #5f9c7a 100%);'),
    'Docker primary operation buttons should use the softer green gradient'
  )

  assert.ok(
    source.includes('.operation-btn.operation-btn-primary {') &&
    source.includes('background: linear-gradient(180deg, #71af8d 0%, #5f9c7a 100%) !important;') &&
    source.includes('border-color: #5f9c7a !important;') &&
    source.includes('color: #ffffff !important;'),
    'Primary operation buttons should directly enforce the softer green gradient'
  )

  assert.ok(
    source.includes('--pl-button-background-color: linear-gradient(180deg, #de6f6f 0%, #d65c5c 100%);'),
    'Danger buttons should keep the red gradient'
  )

  console.log('docker_button_style tests passed')
}

run()
