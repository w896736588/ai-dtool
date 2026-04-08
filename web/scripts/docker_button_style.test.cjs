const assert = require('assert')
const fs = require('fs')
const path = require('path')

const dockerVuePath = path.join(__dirname, '../src/components/Docker.vue')
const source = fs.readFileSync(dockerVuePath, 'utf8')

const run = () => {
  assert.ok(
    /<pl-button[^>]*size="small"[^>]*type="primary"[^>]*plain[^>]*@click="dialogServices\(scope\.row\)"[^>]*>服务列表<\/pl-button>/s.test(source),
    'Service list button should use the same primary plain button semantics as Supervisor'
  )

  assert.ok(
    /<pl-button[^>]*size="small"[^>]*type="primary"[^>]*plain[^>]*@click="status\(scope\.row\)"[^>]*>运行状态<\/pl-button>/s.test(source),
    'Runtime status button should use the same primary plain button semantics as Supervisor'
  )

  assert.ok(
    /<pl-button[^>]*size="small"[^>]*type="success"[^>]*plain[^>]*@click="start\(scope\.row\)"[^>]*>启动（up -d）<\/pl-button>/s.test(source),
    'Start button should use the success plain button semantics'
  )

  assert.ok(
    /<pl-button[^>]*size="small"[^>]*type="success"[^>]*plain[^>]*@click="restart\(scope\.row\)"[^>]*>重启（restart）<\/pl-button>/s.test(source),
    'Restart button should use the success plain button semantics'
  )

  assert.ok(
    /<pl-button[^>]*size="small"[^>]*type="warning"[^>]*plain[^>]*@click="stop\(scope\.row\)"[^>]*>停止\(stop\)<\/pl-button>/s.test(source),
    'Stop button should use the warning plain button semantics like Supervisor row actions'
  )

  assert.ok(
    /<pl-button[^>]*size="small"[^>]*type="primary"[^>]*plain[^>]*@click="configShow\(scope\.row\)"[^>]*>查看compose\.yml<\/pl-button>/s.test(source),
    'View compose button should use the primary plain button semantics'
  )

  assert.ok(
    /<pl-button[^>]*size="small"[^>]*type="primary"[^>]*plain[^>]*@click="envShow\(scope\.row\)"[^>]*>查看env<\/pl-button>/s.test(source),
    'View env button should use the primary plain button semantics'
  )

  assert.ok(
    /<pl-button[^>]*size="small"[^>]*type="success"[^>]*plain[^>]*@click="restart\(scope\.row , item\)"[^>]*>{{ item }}<\/pl-button>/s.test(source),
    'Quick restart buttons should also reuse the success plain button style'
  )

  assert.ok(
    /<pl-button[^>]*size="small"[^>]*type="warning"[^>]*plain[^>]*@click="stop\(scope\.row , item\)"[^>]*>{{ item }}<\/pl-button>/s.test(source),
    'Quick stop buttons should also reuse the warning plain button style'
  )

  assert.ok(
    source.includes('.operation-buttons .el-button,') &&
    source.includes('.quick-actions .el-button {') &&
    source.includes('border-radius: 8px;'),
    'Docker button wrapper styles should only keep the shared rounded-corner enhancement'
  )

  assert.ok(
    !source.includes('.operation-btn.operation-btn-primary {') &&
    !source.includes('.quick-action-restart {') &&
    !source.includes('.quick-action-stop {'),
    'Docker page should remove the old custom button color overrides'
  )

  console.log('docker_button_style tests passed')
}

run()
