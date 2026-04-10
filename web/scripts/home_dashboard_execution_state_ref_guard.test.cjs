const assert = require('assert')
const fs = require('fs')
const path = require('path')

const homePath = path.join(__dirname, '../src/components/Home.vue')
const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const homeSource = fs.readFileSync(homePath, 'utf8')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

const run = () => {
  assert.ok(
    /const dashboardRef = this\.\$refs\.currentRef/.test(homeSource),
    'Home should read dashboard component ref when checking running state'
  )

  assert.ok(
    /dashboardRef\?\.isExecuting|dashboardRef && .*isExecuting/.test(homeSource),
    'Home running-state detection should include Dashboard.isExecuting'
  )

  assert.ok(
    /return \{[\s\S]*\bisExecuting\b[\s\S]*\}/s.test(dashboardSource),
    'Dashboard should expose isExecuting so Home can detect active execution immediately'
  )

  console.log('home_dashboard_execution_state_ref_guard tests passed')
}

run()
