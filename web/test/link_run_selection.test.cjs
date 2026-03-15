const assert = require('assert').strict
const {
  buildLinkAccountOptionsFromEnv,
  buildLinkEnvOptionsFromConfig,
  getLinkRunSelection,
  isLinkRunSelectionComplete,
} = require('../src/utils/link_run_selection.cjs')

const configCmd = {
  data: {
    id: 11,
    open_num: '1',
    open_type: 'tab',
    linkList: [
      {
        label: '已配账号环境',
        userList: [{ user_name: 'alice', password: 'secret' }],
      },
      {
        label: '未配账号环境',
        userList: [],
      },
    ],
  },
}

const envOptions = buildLinkEnvOptionsFromConfig(configCmd)
assert.equal(envOptions[0].dynamicChildren, 'linkAccountList')
assert.equal(envOptions[1].dynamicChildren, undefined, '未配置账号的环境不应继续要求选择账号')

const accountOptions = buildLinkAccountOptionsFromEnv({
  data: {
    env: configCmd.data.linkList[1],
  },
})
assert.equal(accountOptions.length, 0, '未配置账号时不应伪造默认空账号')

const selection = getLinkRunSelection([
  { action: 'linkRun' },
  envOptions[1],
])
assert.equal(isLinkRunSelectionComplete(selection), true, '未配置账号的环境应允许直接执行')

console.log('link_run_selection tests passed')
