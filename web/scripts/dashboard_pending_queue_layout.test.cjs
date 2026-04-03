const assert = require('assert')
const fs = require('fs')
const path = require('path')

const dashboardVuePath = path.join(__dirname, '../src/components/Dashboard.vue')
const source = fs.readFileSync(dashboardVuePath, 'utf8')

const run = () => {
  assert.ok(
    source.includes('class="pending-command-list pending-command-list--horizontal"'),
    'Pending command list should opt into the horizontal queue layout class'
  )

  assert.ok(
    /\.pending-command-panel\s*\{[\s\S]*flex:\s*0 0 clamp\(240px, 32vw, 360px\);[\s\S]*max-width:\s*clamp\(240px, 32vw, 360px\);/m.test(source),
    'Pending command panel should reserve a bounded width instead of squeezing the main input area'
  )

  assert.ok(
    /\.pending-command-list--horizontal\s*\{[\s\S]*display:\s*flex;[\s\S]*flex-direction:\s*row;[\s\S]*flex-wrap:\s*nowrap;[\s\S]*overflow:\s*hidden;/m.test(source),
    'Horizontal pending command list should stay on one row and clip overflow'
  )

  assert.ok(
    /\.pending-command-item\s*\{[\s\S]*flex:\s*0 0 168px;[\s\S]*min-width:\s*0;/m.test(source),
    'Pending command items should keep a fixed horizontal slot and allow internal text truncation'
  )

  console.log('dashboard_pending_queue_layout tests passed')
}

run()
