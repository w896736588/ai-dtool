const assert = require('assert')

const {
  validateProcessItemForm,
  parseRedirectUriValue,
  parseWaitUrlValue,
  serializeRedirectUriValue,
  serializeWaitUrlValue,
  parseCheckKeyExpression,
  parseCheckConfig,
  serializeCheckKeyExpression,
  serializeCheckConfig,
} = require('../src/utils/smart_link_process_validation.cjs')

const createBaseItem = () => ({
  id: 0,
  name: '测试节点',
  type: 'click',
  locator: '.submit-btn',
  value: '',
  out_key: '',
  check_key: '',
  wait_mills: 3000,
  weight: 0,
  domain_limit: '',
  append_to_replace: '0',
  is_async: '0',
  is_error_continue: '0',
  next_ids: '',
})

const run = () => {
  const invalidLocatorResult = validateProcessItemForm({
    item: createBaseItem(),
    formMeta: {
      locator_structured_form: {
        kind: '',
      },
      locator_joiner: 'raw',
      locator_raw: '.dialog&&',
      locator_list: [],
      bool_result_rules: [],
      register_response_urls: [],
      next_id_list: [],
    },
  })
  assert.strictEqual(invalidLocatorResult.valid, false, '非法主元素定位表达式应被拦截')
  assert.ok(
    invalidLocatorResult.fieldErrors.locator.includes('查找方式'),
    '主元素定位应返回对应字段错误说明'
  )

  const invalidKeyResult = validateProcessItemForm({
    item: {
      ...createBaseItem(),
      type: 'bool_exist',
      out_key: 'login-state',
    },
    formMeta: {
      locator_structured_form: {
        kind: 'css',
        value: '.user-info',
      },
      locator_joiner: 'single',
      locator_raw: '',
      locator_list: [{ value: '.user-info', exist_mode: 'exist', match_mode: 'all' }],
      bool_result_rules: [],
      register_response_urls: [],
      next_id_list: [],
    },
  })
  assert.strictEqual(invalidKeyResult.valid, false, '非法输出键格式应被拦截')
  assert.ok(
    invalidKeyResult.fieldErrors.out_key.includes('字母'),
    '输出键错误提示应说明允许格式'
  )

  const invalidRedirectResult = validateProcessItemForm({
    item: {
      ...createBaseItem(),
      type: 'redirect_uri',
    },
    formMeta: {
      value: 'not a url',
      locator_joiner: 'single',
      locator_raw: '',
      locator_list: [],
      bool_result_rules: [],
      register_response_urls: [{ url: 'bad url', wait_second: 0 }],
      next_id_list: [],
    },
  })
  assert.strictEqual(invalidRedirectResult.valid, false, '非法跳转地址和等待地址应被拦截')
  assert.ok(invalidRedirectResult.fieldErrors.value, '跳转地址应有字段错误')
  assert.ok(
    invalidRedirectResult.fieldErrors.register_response_urls,
    '跳转后的等待地址应有字段错误'
  )

  const validBoolResult = validateProcessItemForm({
    item: {
      ...createBaseItem(),
      type: 'bool_result',
      out_key: 'login_state',
    },
    formMeta: {
      locator_joiner: 'single',
      locator_raw: '',
      locator_list: [],
      bool_result_rules: [
        {
          locator_advanced_form: {
            kind: 'css',
            value: '.user-info',
            has_not_kind: 'css',
            has_not_value: '.btn.login_as_reg_btn',
          },
          return: true,
        },
        {
          locator_advanced_form: {
            kind: 'css',
            value: '.login-btn',
          },
          return: false,
        },
      ],
      register_response_urls: [],
      next_id_list: [],
    },
  })
  assert.strictEqual(validBoolResult.valid, true, '合法配置应通过校验')

  const invalidAdvancedLocatorResult = validateProcessItemForm({
    item: {
      ...createBaseItem(),
      type: 'text_content',
    },
    formMeta: {
      locator_editor_mode: 'advanced',
      locator_advanced_form: {
        kind: 'css',
        value: '.username',
        has_not_kind: 'css',
        has_not_value: '',
      },
      bool_result_rules: [],
      register_response_urls: [],
      next_id_list: [],
    },
  })
  assert.strictEqual(invalidAdvancedLocatorResult.valid, false, '高级定位缺少 has_not 子定位值时应被拦截')
  assert.ok(
    invalidAdvancedLocatorResult.fieldErrors.locator.includes('子元素'),
    '高级定位错误提示应指出子元素定位不完整'
  )

  const waitUrlMeta = parseWaitUrlValue('{"response_url":"{scheme}://{domain}/kefuLogin/getLoginQrcode","wait_second":5}')
  assert.strictEqual(
    waitUrlMeta.response_url,
    '{scheme}://{domain}/kefuLogin/getLoginQrcode',
    '等待接口配置应兼容 response_url 小写字段回显'
  )
  assert.strictEqual(waitUrlMeta.wait_second, 5, '等待接口配置应兼容 wait_second 小写字段回显')

  const redirectMeta = parseRedirectUriValue('{"url":"/login","register_response_url":[{"url":"/home","wait_second":8}]}')
  assert.strictEqual(redirectMeta.value, '/login', '跳转登录应兼容 url 小写字段回显')
  assert.strictEqual(redirectMeta.register_response_urls.length, 1, '跳转后等待地址应正确回显为表单列表')
  assert.strictEqual(
    redirectMeta.register_response_urls[0].url,
    '/home',
    '跳转后等待地址应兼容小写 url 字段'
  )

  assert.strictEqual(
    serializeWaitUrlValue({
      response_url: '/api/qrcode',
      wait_second: 6,
    }),
    '{"ResponseUrl":"/api/qrcode","WaitSecond":6}',
    '等待接口配置应序列化回后端结构'
  )

  assert.strictEqual(
    serializeRedirectUriValue({
      value: '/login',
      register_response_urls: [{ url: '/home', wait_second: 8 }],
    }),
    '{"Url":"/login","RegisterResponseUrl":[{"Url":"/home","WaitSecond":8}]}',
    '跳转登录表单应序列化回后端结构'
  )

  const checkRules = parseCheckKeyExpression('login_state&&!has_error')
  assert.strictEqual(checkRules.length, 2, '判断键表达式应拆成多条条件')
  assert.strictEqual(checkRules[0].key, 'login_state', '判断键应正确解析普通条件')
  assert.strictEqual(checkRules[0].expect, 'true', '普通条件应解析为必须为真')
  assert.strictEqual(checkRules[1].key, 'has_error', '判断键应正确解析取反条件')
  assert.strictEqual(checkRules[1].expect, 'false', '取反条件应解析为必须为假')

  assert.strictEqual(
    serializeCheckKeyExpression([
      { key: 'login_state', expect: 'true' },
      { key: 'has_error', expect: 'false' },
    ]),
    'login_state&&!has_error',
    '判断键表单应序列化回后端 &&/! 表达式'
  )

  const compareConfig = parseCheckConfig('{login_user}!={user_name}')
  assert.strictEqual(compareConfig.mode, 'compare', '带 == / != 的判断条件应识别为比较模式')
  assert.strictEqual(compareConfig.compare_rule.left, '{login_user}', '比较条件左值应正确解析')
  assert.strictEqual(compareConfig.compare_rule.operator, '!=', '比较条件运算符应正确解析')
  assert.strictEqual(compareConfig.compare_rule.right, '{user_name}', '比较条件右值应正确解析')

  assert.strictEqual(
    serializeCheckConfig({
      mode: 'compare',
      compare_rule: {
        left: '{login_user}',
        operator: '!=',
        right: '{user_name}',
      },
    }),
    '{login_user}!={user_name}',
    '比较条件表单应序列化回后端比较表达式'
  )

  console.log('smart_link_process_validation tests passed')
}

run()
