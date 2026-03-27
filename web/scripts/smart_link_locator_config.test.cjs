const assert = require('assert')

const {
  createBaseLocatorMeta,
  buildLocatorConfigByType,
  deserializeLocatorConfigToFormMeta,
  isLocatorConfigPayload,
} = require('../src/utils/smart_link_locator_config.cjs')

const run = () => {
  const boolConfig = buildLocatorConfigByType('bool_result', {
    bool_result_rules: [
      {
        id: 'rule_1',
        on_found: true,
        base_locator: {
          ...createBaseLocatorMeta(),
          locator_structured_form: {
            ...createBaseLocatorMeta().locator_structured_form,
            kind: 'css',
            value: '.user-info',
          },
        },
      },
      {
        id: 'rule_2',
        on_found: false,
        base_locator: {
          ...createBaseLocatorMeta(),
          locator_structured_form: {
            ...createBaseLocatorMeta().locator_structured_form,
            kind: 'button_text',
            value: '登录',
          },
        },
      },
    ],
  })

  assert.strictEqual(boolConfig.mode, 'bool_result', '布尔判断应生成 bool_result 模式')
  assert.strictEqual(boolConfig.strategy, 'first_match_return', '布尔判断应写入 first_match_return 策略')
  assert.strictEqual(boolConfig.locators[0].on_found, true, '布尔判断规则应保留 on_found=true')
  assert.strictEqual(boolConfig.locators[1].on_found, false, '布尔判断规则应保留 on_found=false')

  const textConfig = buildLocatorConfigByType('text_content', {
    text_content_locators: [
      {
        id: 'rule_1',
        on_found: 'extract_text',
        base_locator: {
          ...createBaseLocatorMeta(),
          locator_structured_form: {
            ...createBaseLocatorMeta().locator_structured_form,
            kind: 'css',
            value: '.content',
          },
        },
      },
      {
        id: 'rule_2',
        on_found: 'return_empty',
        base_locator: {
          ...createBaseLocatorMeta(),
          locator_structured_form: {
            ...createBaseLocatorMeta().locator_structured_form,
            kind: 'css',
            value: '.empty-state',
          },
        },
      },
    ],
  })

  assert.strictEqual(textConfig.mode, 'text_content', '文本提取应生成 text_content 模式')
  assert.strictEqual(textConfig.strategy, 'first_match_return', '文本提取应生成 first_match_return 策略')
  assert.strictEqual(textConfig.locators[0].on_found, 'extract_text', '文本提取规则应保留 extract_text')
  assert.strictEqual(textConfig.locators[1].on_found, 'return_empty', '文本提取规则应保留 return_empty')

  const clickConfig = buildLocatorConfigByType('click', {
    action_strategy: 'first_found_do_action',
    action_locators: [
      {
        id: 'click_1',
        base_locator: {
          ...createBaseLocatorMeta(),
          locator_structured_form: {
            ...createBaseLocatorMeta().locator_structured_form,
            kind: 'css',
            value: '.confirm-btn',
          },
        },
      },
    ],
  })

  assert.strictEqual(clickConfig.strategy, 'first_found_do_action', '点击配置应生成 first_found_do_action 策略')
  assert.strictEqual(clickConfig.options.action_type, 'click', '点击配置应记录 action_type=click')
  assert.strictEqual(isLocatorConfigPayload(clickConfig), true, '新 locator 配置应能被识别')

  const roundTripMeta = deserializeLocatorConfigToFormMeta(textConfig)
  assert.strictEqual(roundTripMeta.text_content_locators.length, 2, '文本提取配置应能回显两个基础定位')
  assert.strictEqual(roundTripMeta.text_content_locators[0].on_found, 'extract_text', 'extract_text 应能正确回显')
  assert.strictEqual(
    roundTripMeta.text_content_locators[1].base_locator.locator_structured_form.value,
    '.empty-state',
    'return_empty 定位值应正确回显'
  )

  console.log('smart_link_locator_config tests passed')
}

run()
