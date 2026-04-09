const assert = require('assert')
const fs = require('fs')
const path = require('path')
const vm = require('vm')

const MODULE_PATH = path.resolve(__dirname, '../src/utils/base/sse_distribute.js')

const loadSseDistributeModule = () => {
  let code = fs.readFileSync(MODULE_PATH, 'utf8')
  code = code.replace(/import\s+t\s+from\s+["'][^"']+["'];?\s*/g, '')
  code = code.replace(/import\s+base\s+from\s+["'][^"']+["'];?\s*/g, '')
  code = code.replace(/import\s+store\s+from\s+["'][^"']+["'];?\s*/g, '')
  code = code.replace(/export default\s*\{/, 'module.exports = {')

  const context = {
    module: { exports: {} },
    exports: {},
    console,
    EventSource: function EventSource() {},
    base: {
      GenerateId(prefix) {
        return `${prefix}_fixed`
      },
      GetSseApiHost() {
        return 'http://localhost'
      }
    },
    store: {},
    t: {},
    JSON,
  }
  vm.createContext(context)
  vm.runInContext(code, context, { filename: MODULE_PATH })
  return context.module.exports
}

const run = () => {
  const sseDistribute = loadSseDistributeModule()
  const firstId = sseDistribute.GetSseDistributeId('dashboard_git')
  const secondId = sseDistribute.GetSseDistributeId('dashboard_git')

  assert.notStrictEqual(firstId, secondId, '连续两次调用应该生成不同的 SSE 分发 ID')
  assert.ok(firstId.startsWith('dashboard_git'), '生成的 ID 应保留业务前缀')
  assert.ok(secondId.startsWith('dashboard_git'), '生成的 ID 应保留业务前缀')

  console.log('sse_distribute_id tests passed')
}

run()
