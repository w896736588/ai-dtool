const assert = require('assert')
const fs = require('fs')
const path = require('path')

const homePath = path.join(__dirname, '../src/components/Home.vue')
const source = fs.readFileSync(homePath, 'utf8')

const run = () => {
  assert.ok(
    /isDashboardCommandRunning\(\)\s*\{[\s\S]*querySelector\((HOME_DASHBOARD_RUNNING_SELECTOR|['"][^'"]*(command-status-running|result-line-running)[^'"]*['"])\)/s.test(source),
    'Home dashboard should expose running-state detection for command panel'
  )

  assert.ok(
    /if \(this\.homeDashboardPageIndex === HOME_DASHBOARD_PAGE_COMMAND\) \{/.test(source) &&
      /if \(!isRightHotZone && this\.isDashboardCommandRunning\(\)\) \{\s*return\s*\}/.test(source) &&
      /this\.switchHomeDashboardPage\(HOME_DASHBOARD_PAGE_TASK\)/.test(source),
    'Wheel switching from command page should be blocked while command execution is running'
  )

  assert.ok(
    /const blockingScrollableAncestor = resolveHomeDashboardPageSwitchBlocker\(event\.target, deltaY, currentTarget\)/.test(source) &&
      /if \(!isRightHotZone && blockingScrollableAncestor\) \{\s*return\s*\}/.test(source),
    'Hot-zone page switching should keep absolute priority while non-hot-zone wheel events remain blocked by inner scroll containers'
  )

  console.log('home_dashboard_execution_wheel_guard tests passed')
}

run()
