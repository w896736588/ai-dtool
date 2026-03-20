const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

assert.ok(
  dashboardSource.includes('// 加载历史命令列表（按使用次数升序，使用最多的在最下面）'),
  'Dashboard 历史命令列表应改为按使用次数升序展示'
)

assert.ok(
  dashboardSource.includes('return a.count - b.count'),
  'Dashboard 历史命令列表应将使用次数最多的命令排在最后'
)

assert.ok(
  dashboardSource.includes('if (normalizedList.every(item => item && item.insertOnly)) {\n        return Math.max(normalizedList.length - 1, 0)\n      }') ||
  dashboardSource.includes('if (normalizedList.every(item => item && item.insertOnly)) {\r\n        return Math.max(normalizedList.length - 1, 0)\r\n      }'),
  'Dashboard 历史命令候选应默认选中末项，也就是使用次数最多的命令'
)

console.log('dashboard_history_list_order_regression tests passed')
