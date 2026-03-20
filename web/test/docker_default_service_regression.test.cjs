const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const dockerPath = path.join(__dirname, '../src/components/Docker.vue')
const dockerSource = fs.readFileSync(dockerPath, 'utf8')

assert.ok(
  dockerSource.includes('加入默认服务'),
  'Docker 服务列表弹窗应提供“加入默认服务”按钮'
)

assert.ok(
  dockerSource.includes('移除默认服务'),
  'Docker 服务列表弹窗应在已加入时切换为“移除默认服务”按钮'
)

assert.ok(
  dockerSource.includes('toggleDockerDefaultService'),
  'Docker 页面应复用默认服务切换工具，避免字符串处理分散'
)

console.log('docker_default_service_regression tests passed')
