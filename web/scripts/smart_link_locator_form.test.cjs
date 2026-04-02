const assert = require('assert')

const {
  buildAdvancedLocatorPayload,
  deserializeLocatorEditorState,
  stringifyLocatorPayload,
} = require('../src/utils/smart_link_locator_form.cjs')

const run = () => {
  const advancedPayload = buildAdvancedLocatorPayload({
    kind: 'css',
    value: '.username',
    has_not_kind: 'css',
    has_not_value: '.btn.login_as_reg_btn',
    has_text: '',
    has_not_text: '',
    chain_kind: '',
    chain_value: '',
    exact: false,
    negate: false,
    pick_mode: 'none',
    nth: 0,
    timeout_mills: 3000,
    visible: '',
  })

  assert.deepStrictEqual(
    advancedPayload,
    {
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
        timeout_mills: 3000,
      },
    },
    '高级定位表单应序列化为后端已支持的 has_not 结构'
  )

  const editorState = deserializeLocatorEditorState(JSON.stringify({
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
      timeout_mills: 3000,
    },
  }), { preferAdvanced: true })

  assert.strictEqual(editorState.mode, 'advanced', '带 filters 的结构化定位应自动进入高级模式')
  assert.strictEqual(editorState.advancedForm.value, '.username', '高级模式应正确回填主定位值')
  assert.strictEqual(
    editorState.advancedForm.has_not_value,
    '.btn.login_as_reg_btn',
    '高级模式应正确回填 has_not 子定位值'
  )
  assert.strictEqual(editorState.advancedForm.has_kind, '', '未配置 has 时不应默认把包含子元素查找方式标成必填')
  assert.strictEqual(editorState.advancedForm.chain_kind, '', '未配置 chain 时不应默认把向下查找方式标成必填')

  assert.strictEqual(
    stringifyLocatorPayload(advancedPayload),
    JSON.stringify(advancedPayload, null, 2),
    '定位结构 stringify 应输出格式化 JSON'
  )

  console.log('smart_link_locator_form tests passed')
}

run()
