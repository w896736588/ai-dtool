const assert = require('assert').strict
const fs = require('fs')
const path = require('path')

const dashboardPath = path.join(__dirname, '../src/components/Dashboard.vue')
const dashboardSource = fs.readFileSync(dashboardPath, 'utf8')

assert.ok(
  dashboardSource.includes("from '@/utils/link_run_selection.cjs'"),
  'Dashboard 应复用公共的 link_run_selection 工具，避免首页和工具逻辑漂移'
)

assert.ok(
  !dashboardSource.includes('默认账号(空)'),
  'Dashboard 不应再为未配置账号的环境伪造默认空账号'
)

assert.ok(
  !dashboardSource.includes('return !!(selection.envCmd && selection.accountCmd)'),
  'Dashboard 不应在未配置账号的环境里强制要求选择账号'
)

assert.ok(
  dashboardSource.includes('insertText: `${configName}/${envName}`'),
  'Dashboard 选择 link 环境时应写回唯一的“配置/环境”标识，避免同名环境回车时串选'
)

assert.ok(
  dashboardSource.includes('const getCommandInputToken = (cmd) => {'),
  'Dashboard 应统一使用选中时的输入 token 回填命令，避免 Tab 选中后 Enter 重解析漂移'
)

assert.ok(
  dashboardSource.includes("if (hasConfiguredLinkAccounts(selection.envCmd) && !selection.accountCmd)"),
  'Dashboard 只有在环境实际配置账号时才应继续提示选择账号'
)

console.log('dashboard_link_run_selection_regression tests passed')
