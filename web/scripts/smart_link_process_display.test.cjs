const assert = require('assert')

const {
  buildProcessItemDisplayDetails,
  formatRawLocator,
  normalizeTokenDisplay,
} = require('../src/utils/smart_link_process_display.cjs')

const run = () => {
  const boolResultLocator = JSON.stringify([
    {
      locator: {
        spec: {
          method: 'locator',
          value: '.user-info.ant-dropdown-trigger',
          timeout_mills: 3000,
        },
      },
      return: false,
    },
    {
      locator: {
        spec: {
          method: 'locator',
          value: '.login-btn',
          timeout_mills: 3000,
        },
      },
      return: true,
    },
  ])

  const locatorLines = formatRawLocator(boolResultLocator)
  assert.strictEqual(locatorLines.length, 2, '布尔判断定位规则应拆分成多行展示')
  assert.ok(locatorLines[0].includes('.user-info.ant-dropdown-trigger'), '第一条规则应展示定位值')
  assert.ok(locatorLines[0].includes('false'), '第一条规则应展示返回结果')
  assert.ok(locatorLines[1].includes('.login-btn'), '第二条规则应展示定位值')

  const detailList = buildProcessItemDisplayDetails({
    type: 'bool_result',
    locator: boolResultLocator,
    out_key: 'need_login',
    append_to_replace: '0',
    wait_mills: 3000,
  })
  assert.strictEqual(detailList[0].label, '定位规则', '详情列表应先展示定位规则')
  assert.strictEqual(detailList[0].lines.length, 2, '定位规则详情应保留多行结果')
  assert.strictEqual(detailList[1].lines[0], '{need_login}', '输出变量应统一展示为带花括号形式')
  assert.strictEqual(detailList[2].lines[0], '否', '替换列表标记应展示为中文是/否')
  assert.strictEqual(detailList[3].lines[0], '3000ms', '等待时长应展示为毫秒文案')

  assert.strictEqual(normalizeTokenDisplay('{user_name}'), '{user_name}', '已带花括号的输出变量不应重复包装')
  assert.strictEqual(normalizeTokenDisplay('password'), '{password}', '普通输出变量应补齐花括号')

  const advancedLocatorLines = formatRawLocator(JSON.stringify({
    spec: {
      method: 'locator',
      value: '.username',
      filters: [
        {
          has_not: {
            method: 'locator',
            value: '.btn.login_as_reg_btn',
          },
        },
      ],
      chain: [
        {
          method: 'locator',
          value: '.nickname',
        },
      ],
    },
  }))
  assert.strictEqual(advancedLocatorLines.length, 1, '单个高级结构化定位应展示为一条说明')
  assert.ok(advancedLocatorLines[0].includes('且不包含'), '高级结构化定位应展示 has_not 说明')
  assert.ok(advancedLocatorLines[0].includes('再向下查找'), '高级结构化定位应展示 chain 说明')

  const clickConfigLines = formatRawLocator(JSON.stringify({
    version: 2,
    mode: 'click',
    strategy: 'first_found_do_action',
    locators: [
      {
        id: 'loc_1',
        query: {
          spec: {
            method: 'locator',
            value: '.confirm-btn',
          },
        },
      },
    ],
    options: {
      action_type: 'click',
    },
  }))
  assert.strictEqual(clickConfigLines.length, 2, '新版 click locator 配置应展示策略和基础定位')
  assert.ok(clickConfigLines[0].includes('任意一个元素存在时执行'), '新版 click 配置应展示操作策略')
  assert.ok(clickConfigLines[1].includes('.confirm-btn'), '新版 click 配置应展示基础定位摘要')

  const textConfigLines = formatRawLocator(JSON.stringify({
    version: 2,
    mode: 'text_content',
    strategy: 'first_match_return',
    locators: [
      {
        id: 'rule_1',
        on_found: 'extract_text',
        query: {
          spec: {
            method: 'locator',
            value: '.username',
          },
        },
      },
      {
        id: 'rule_2',
        on_found: 'return_empty',
        query: {
          spec: {
            method: 'locator',
            value: '.btn.login_as_reg_btn',
          },
        },
      },
    ],
  }))
  assert.strictEqual(textConfigLines.length, 2, '新版 text_content locator 配置应展示两条规则')
  assert.ok(textConfigLines[0].includes('命中返回其提取'), 'extract_text 规则应展示为返回提取内容')
  assert.ok(textConfigLines[1].includes('命中返回空值'), 'return_empty 规则应展示为返回空值')

  console.log('smart_link_process_display tests passed')
}

run()
