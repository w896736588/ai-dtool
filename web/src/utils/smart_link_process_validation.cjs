// PROCESS_ITEM_FIELD_GUIDES 用于集中维护流程项字段说明。
// PROCESS_ITEM_FIELD_GUIDES centralizes field guidance for process item forms.
const {
  parseStructuredLocatorPayload,
} = require('./smart_link_locator_form.cjs')
const {
  createBaseLocatorMeta,
  isLocatorConfigPayload,
  normalizeBaseLocatorMeta,
} = require('./smart_link_locator_config.cjs')

const PROCESS_ITEM_FIELD_GUIDES = {
  locator: '统一使用结构化 Locator 配置，按页面可见内容填写后系统会自动生成后端需要的 {"spec": {...}} 结构。',
  value: '跳转步骤填写目标地址；输入步骤可使用 {user_name}、{password}、{rand} 等占位符。',
  out_key: '仅支持英文字母开头，后续可包含字母、数字、下划线，例如 {login_state}。',
  check_key: '仅支持英文字母开头，后续可包含字母、数字、下划线，例如 login_state。',
  response_url: '支持 http(s):// 完整地址、/path 相对路径，以及 {scheme}://{domain}/path 这类占位写法。',
  register_response_urls: '每条等待地址都需要合法地址格式，并填写大于 0 的等待秒数。',
  bool_result_rules: '每条规则都使用结构化定位，命中定位时按右侧 true / false 返回结果。',
  domain_limit: '可选，仅填写域名本身，例如 example.com 或 sub.example.com，不要包含协议和路径。',
}

// PROCESS_TYPE_FIELDS 定义每种类型需要展示和校验的字段。
// PROCESS_TYPE_FIELDS describes which logical fields each process type uses.
const PROCESS_TYPE_FIELDS = {
  text_content: ['locator', 'out_key', 'check_key'],
  redirect_uri: ['value', 'register_response_urls', 'check_key'],
  wait_url: ['response_url', 'wait_second', 'check_key'],
  wait: ['check_key'],
  bool_result: ['bool_result_rules', 'out_key', 'check_key'],
  bool_exist: ['locator', 'out_key', 'check_key'],
  click: ['locator', 'check_key'],
  input: ['locator', 'value', 'out_key', 'check_key'],
  close: [],
  no_exist_wait: ['locator', 'wait_second', 'wait_count', 'out_key', 'check_key'],
  canvas_image: ['locator', 'out_key', 'check_key'],
  login_username_password: ['check_key'],
  delete_element: ['locator'],
}

const KEY_PATTERN = /^[A-Za-z][A-Za-z0-9_]*$/
const DOMAIN_PATTERN = /^(?=.{1,253}$)(?!-)(?:[A-Za-z0-9-]{1,63}\.)+[A-Za-z]{2,63}$/

function isValidTokenKey(value) {
  const normalizedValue = normalizeText(value)
  if (!normalizedValue) return false
  if (KEY_PATTERN.test(normalizedValue)) return true
  const braceMatch = normalizedValue.match(/^\{([A-Za-z][A-Za-z0-9_]*)\}$/)
  return Boolean(braceMatch)
}

function showTypeField(type, fieldName) {
  return (PROCESS_TYPE_FIELDS[type] || []).includes(fieldName)
}

function shouldShowAppendToReplace(item = {}) {
  const type = normalizeText(item.type)
  if (type === 'click' || type === 'delete_element') return false
  if (type === 'input') return false
  if (!showTypeField(type, 'out_key')) return false
  return normalizeText(item.out_key) !== ''
}

function normalizeText(value) {
  return String(value || '').trim()
}

function isPositiveInteger(value) {
  return Number.isInteger(Number(value)) && Number(value) > 0
}

function isUrlLikeOrPathLike(value) {
  const normalizedValue = normalizeText(value)
  if (!normalizedValue || /\s/.test(normalizedValue)) return false
  if (normalizedValue.startsWith('/')) return true
  if (/^\{[A-Za-z0-9_]+\}:\/\/\{[A-Za-z0-9_]+\}(\/.*)?$/.test(normalizedValue)) return true
  try {
    const parsedUrl = new URL(normalizedValue)
    return parsedUrl.protocol === 'http:' || parsedUrl.protocol === 'https:'
  } catch (error) {
    return false
  }
}

function safeParseJson(text, fallback) {
  if (!text) return fallback
  try {
    return JSON.parse(text)
  } catch (error) {
    return fallback
  }
}

function normalizeRegisterResponseUrlList(list) {
  const sourceList = Array.isArray(list) ? list : []
  return sourceList.map((item) => ({
    uid: item.uid || '',
    url: item.Url || item.url || '',
    wait_second: Number(item.WaitSecond || item.wait_second || 10),
  }))
}

// parseCheckKeyExpression 用于把后端的 && / ! 条件表达式转成前端可编辑列表。
// parseCheckKeyExpression parses backend check_key expressions into editable rules.
function parseCheckKeyExpression(rawValue) {
  const normalizedValue = normalizeText(rawValue)
  if (!normalizedValue) return []
  return normalizedValue
    .split('&&')
    .map((item) => normalizeText(item))
    .filter(Boolean)
    .map((item) => ({
      key: item.startsWith('!') ? item.slice(1) : item,
      expect: item.startsWith('!') ? 'false' : 'true',
    }))
    .filter((item) => normalizeText(item.key))
}

// serializeCheckKeyExpression 用于把判断条件列表转回后端 check_key 表达式。
// serializeCheckKeyExpression serializes check key rules back to backend syntax.
function serializeCheckKeyExpression(ruleList) {
  return (Array.isArray(ruleList) ? ruleList : [])
    .map((item) => {
      const key = normalizeText(item && item.key)
      if (!key) return ''
      return item.expect === 'false' ? `!${key}` : key
    })
    .filter(Boolean)
    .join('&&')
}

// parseCheckConfig 用于区分布尔判断和字符串比较两类 check_key 配置。
// parseCheckConfig distinguishes boolean checks from compare checks.
function parseCheckConfig(rawValue) {
  const normalizedValue = normalizeText(rawValue)
  if (!normalizedValue) {
    return {
      mode: 'none',
      bool_rules: [],
      compare_rule: {
        left: '',
        operator: '==',
        right: '',
      },
    }
  }

  if (normalizedValue.includes('!=')) {
    const checkList = normalizedValue.split('!=')
    return {
      mode: 'compare',
      bool_rules: [],
      compare_rule: {
        left: normalizeText(checkList[0]),
        operator: '!=',
        right: normalizeText(checkList.slice(1).join('!=')),
      },
    }
  }

  if (normalizedValue.includes('==')) {
    const checkList = normalizedValue.split('==')
    return {
      mode: 'compare',
      bool_rules: [],
      compare_rule: {
        left: normalizeText(checkList[0]),
        operator: '==',
        right: normalizeText(checkList.slice(1).join('==')),
      },
    }
  }

  return {
    mode: 'bool',
    bool_rules: parseCheckKeyExpression(normalizedValue),
    compare_rule: {
      left: '',
      operator: '==',
      right: '',
    },
  }
}

// serializeCheckConfig 用于把判断条件表单统一写回后端 check_key 字符串。
// serializeCheckConfig serializes the check form back to backend check_key syntax.
function serializeCheckConfig(config = {}) {
  if (config.mode === 'compare') {
    const left = normalizeText(config.compare_rule && config.compare_rule.left)
    const operator = config.compare_rule && config.compare_rule.operator === '!=' ? '!=' : '=='
    const right = normalizeText(config.compare_rule && config.compare_rule.right)
    if (!left || !right) return ''
    return `${left}${operator}${right}`
  }
  if (config.mode === 'bool') {
    return serializeCheckKeyExpression(config.bool_rules)
  }
  return ''
}

// parseWaitUrlValue 用于兼容旧小写字段和新后端字段的等待接口配置回显。
// parseWaitUrlValue parses wait_url payloads from both lowercase and backend-style keys.
function parseWaitUrlValue(rawValue) {
  const parsed = safeParseJson(rawValue, {})
  return {
    response_url: parsed.ResponseUrl || parsed.response_url || '',
    wait_second: Number(parsed.WaitSecond || parsed.wait_second || 10),
  }
}

// serializeWaitUrlValue 用于把等待接口表单序列化回后端结构。
// serializeWaitUrlValue serializes wait_url form data back to backend payload.
function serializeWaitUrlValue(formMeta) {
  return JSON.stringify({
    ResponseUrl: normalizeText(formMeta.response_url),
    WaitSecond: Number(formMeta.wait_second || 10),
  })
}

// parseRedirectUriValue 用于兼容旧小写字段和新后端字段的跳转配置回显。
// parseRedirectUriValue parses redirect_uri payloads from lowercase and backend-style keys.
function parseRedirectUriValue(rawValue) {
  const parsed = safeParseJson(rawValue, null)
  if (parsed && typeof parsed === 'object') {
    return {
      value: parsed.Url || parsed.url || '',
      register_response_urls: normalizeRegisterResponseUrlList(
        parsed.RegisterResponseUrl || parsed.register_response_url || []
      ),
    }
  }
  return {
    value: normalizeText(rawValue),
    register_response_urls: [],
  }
}

// serializeRedirectUriValue 用于把跳转表单序列化回后端结构。
// serializeRedirectUriValue serializes redirect_uri form data back to backend payload.
function serializeRedirectUriValue(formMeta) {
  const registerResponseUrlList = normalizeRegisterResponseUrlList(formMeta.register_response_urls)
    .filter((item) => normalizeText(item.url))
    .map((item) => ({
      Url: normalizeText(item.url),
      WaitSecond: Number(item.wait_second || 10),
    }))

  if (registerResponseUrlList.length === 0) {
    return normalizeText(formMeta.value)
  }

  return JSON.stringify({
    Url: normalizeText(formMeta.value),
    RegisterResponseUrl: registerResponseUrlList,
  })
}

function isLocatorSegmentValid(segment) {
  const normalizedSegment = normalizeText(segment)
  if (!normalizedSegment) return false

  const partList = normalizedSegment.split('|').map(item => item.trim()).filter(Boolean)
  if (partList.length === 0) return false

  const locatorValue = partList[0].startsWith('!') ? partList[0].slice(1) : partList[0]
  if (!locatorValue) return false

  return partList.slice(1).every(flag => flag === 'first')
}

function validateRawLocator(rawLocator) {
  const normalizedLocator = normalizeText(rawLocator)
  if (!normalizedLocator) {
    return '主元素定位不能为空，请至少填写一个定位表达式。'
  }

  const invalidEdgePattern = /^(&&|\|\|)|(&&|\|\|)$/
  if (invalidEdgePattern.test(normalizedLocator)) {
    return '主元素定位不能以 && 或 || 开头或结尾。'
  }

  const hasAnd = normalizedLocator.includes('&&')
  const hasOr = normalizedLocator.includes('||')
  const segmentList = hasAnd && !hasOr
    ? normalizedLocator.split('&&')
    : hasOr && !hasAnd
      ? normalizedLocator.split('||')
      : hasAnd && hasOr
        ? normalizedLocator.split(/&&|\|\|/)
        : [normalizedLocator]

  if (segmentList.some(segment => !isLocatorSegmentValid(segment))) {
    return '主元素定位格式不正确，请检查 !selector、|first、&&、|| 的组合是否完整。'
  }

  return ''
}

function validateLocatorField(formMeta) {
  if (normalizeText(formMeta.locator_editor_mode) === 'advanced') {
    return validateAdvancedLocatorField(formMeta.locator_advanced_form)
  }
  const structuredForm = formMeta.locator_structured_form || {}
  if (!normalizeText(structuredForm.kind)) {
    return '请先选择查找方式。'
  }
  if (normalizeText(structuredForm.kind) === 'button_text') {
    if (!normalizeText(structuredForm.value) && !normalizeText(structuredForm.target_text)) {
      return '按按钮文字查找时，至少要填写按钮上显示的文字。'
    }
    return ''
  }
  if (!normalizeText(structuredForm.value)) {
    return '当前查找方式还没有填写内容。'
  }
  return ''
}

function validateBaseLocatorMeta(baseLocator) {
  const meta = normalizeBaseLocatorMeta({
    locator_editor_mode: 'simple',
    locator_structured_form: createBaseLocatorMeta().locator_structured_form,
    locator_advanced_form: createBaseLocatorMeta().locator_advanced_form,
    ...(baseLocator || {}),
  })
  if (normalizeText(meta.locator_editor_mode) !== 'simple') {
    return '基础定位只支持 CSS / XPath 这类可直接用于 Locator 的定位表达式。'
  }
  if (normalizeText(meta.locator_structured_form && meta.locator_structured_form.kind) !== 'css') {
    return '主元素定位方式固定为 CSS / XPath。'
  }
  return validateLocatorField(meta)
}

function validateAdvancedLocatorField(locatorAdvancedForm) {
  const advancedForm = locatorAdvancedForm || {}
  if (!normalizeText(advancedForm.kind)) {
    return '请先选择主元素查找方式。'
  }
  if (!normalizeText(advancedForm.value)) {
    return '请先填写主元素定位内容。'
  }
  if (normalizeText(advancedForm.has_kind) && !normalizeText(advancedForm.has_value)) {
    return '包含子元素时，需要填写完整的子元素定位内容。'
  }
  if (normalizeText(advancedForm.has_value) && !normalizeText(advancedForm.has_kind)) {
    return '包含子元素时，需要先选择子元素查找方式。'
  }
  if (normalizeText(advancedForm.has_not_kind) && !normalizeText(advancedForm.has_not_value)) {
    return '不包含子元素时，需要填写完整的子元素定位内容。'
  }
  if (normalizeText(advancedForm.has_not_value) && !normalizeText(advancedForm.has_not_kind)) {
    return '不包含子元素时，需要先选择子元素查找方式。'
  }
  if (normalizeText(advancedForm.chain_value) && !normalizeText(advancedForm.chain_kind)) {
    return '向下继续查找时，需要先选择子节点查找方式。'
  }
  if (normalizeText(advancedForm.pick_mode) === 'nth' && !Number.isInteger(Number(advancedForm.nth))) {
    return '取第 N 个结果时，必须填写有效的索引。'
  }
  return ''
}

function setFieldError(fieldErrors, fieldName, message) {
  if (!message || fieldErrors[fieldName]) return
  fieldErrors[fieldName] = message
}

// validateProcessItemForm 校验流程项表单语义与格式。
// validateProcessItemForm validates process item form semantics before save.
function validateProcessItemForm({ item = {}, formMeta = {} }) {
  const fieldErrors = {}
  const type = normalizeText(item.type)
  const useLocatorConfig = isLocatorConfigPayload(item.locator)

  if (!normalizeText(item.name)) {
    setFieldError(fieldErrors, 'name', '名称不能为空。')
  }
  if (!type) {
    setFieldError(fieldErrors, 'type', '请选择类型。')
  }
  if (Number(item.wait_mills) < 0) {
    setFieldError(fieldErrors, 'wait_mills', '等待时长不能小于 0。')
  }

  const domainLimit = normalizeText(item.domain_limit)
  if (domainLimit && !DOMAIN_PATTERN.test(domainLimit)) {
    setFieldError(fieldErrors, 'domain_limit', '域名限制格式不正确，请填写 example.com 这类纯域名。')
  }

  if (showTypeField(type, 'locator')) {
    if (!(useLocatorConfig && (type === 'text_content' || type === 'click' || type === 'input'))) {
      setFieldError(fieldErrors, 'locator', validateLocatorField(formMeta))
    }
  }
  if (showTypeField(type, 'value')) {
    const valueText = normalizeText(formMeta.value)
    if (!valueText) {
      setFieldError(fieldErrors, 'value', '当前类型要求填写值。')
    } else if (type === 'redirect_uri' && !isUrlLikeOrPathLike(valueText)) {
      setFieldError(fieldErrors, 'value', '跳转地址格式不正确，请填写 http(s):// 地址或 /path。')
    }
  }

  const normalizedOutKey = normalizeText(formMeta.out_key || item.out_key)
  if (showTypeField(type, 'out_key') && normalizedOutKey && !isValidTokenKey(normalizedOutKey)) {
    setFieldError(fieldErrors, 'out_key', '输出键格式不正确，请使用英文字母开头，后续仅包含字母、数字、下划线。')
  }
  if (showTypeField(type, 'check_key')) {
    const checkMode = normalizeText(formMeta.check_mode)
    if (checkMode === 'bool') {
      const checkRuleList = Array.isArray(formMeta.check_rule_list) ? formMeta.check_rule_list : []
      const hasInvalidCheckRule = checkRuleList.some((rule) => !normalizeText(rule && rule.key))
      const normalizedKeyList = checkRuleList
        .map((rule) => normalizeText(rule && rule.key))
        .filter(Boolean)
      const hasDuplicateCheckKey = new Set(normalizedKeyList).size !== normalizedKeyList.length
      if (hasInvalidCheckRule) {
        setFieldError(fieldErrors, 'check_key', '判断条件中有未选择的输出。')
      } else if (hasDuplicateCheckKey) {
        setFieldError(fieldErrors, 'check_key', '同一个前序输出不需要重复添加，请保留一条并设置期望结果。')
      }
    } else if (checkMode === 'compare') {
      const left = normalizeText(formMeta.compare_rule && formMeta.compare_rule.left)
      const right = normalizeText(formMeta.compare_rule && formMeta.compare_rule.right)
      if (!left || !right) {
        setFieldError(fieldErrors, 'check_key', '比较条件需要完整选择左值、比较类型和右值。')
      }
    }
  }

  if (showTypeField(type, 'wait_second') && !isPositiveInteger(formMeta.wait_second)) {
    setFieldError(fieldErrors, 'wait_second', '等待秒数必须是大于 0 的整数。')
  }
  if (showTypeField(type, 'wait_count') && !isPositiveInteger(formMeta.wait_count)) {
    setFieldError(fieldErrors, 'wait_count', '轮询次数必须是大于 0 的整数。')
  }

  if (showTypeField(type, 'response_url') && !isUrlLikeOrPathLike(formMeta.response_url)) {
    setFieldError(fieldErrors, 'response_url', '等待地址格式不正确，请填写完整 URL、/path 或 {scheme}://{domain}/path。')
  }

  if (showTypeField(type, 'register_response_urls')) {
    const responseUrlList = Array.isArray(formMeta.register_response_urls) ? formMeta.register_response_urls : []
    const hasInvalidResponseUrl = responseUrlList.some(item => {
      return !isUrlLikeOrPathLike(item && item.url) || !isPositiveInteger(item && item.wait_second)
    })
    if (hasInvalidResponseUrl) {
      setFieldError(fieldErrors, 'register_response_urls', '跳转后的等待地址配置不完整，每条都需要合法地址和大于 0 的等待秒数。')
    }
  }

  if (showTypeField(type, 'bool_result_rules')) {
    const boolResultRules = Array.isArray(formMeta.bool_result_rules) ? formMeta.bool_result_rules : []
    const hasInvalidBoolResultRule = boolResultRules.some((rule) => {
      if (!rule) return true
      if (rule.base_locator) {
        return validateBaseLocatorMeta(rule.base_locator) !== ''
      }
      if (rule.locator_advanced_form) {
        return validateAdvancedLocatorField(rule.locator_advanced_form) !== ''
      }
      const structuredLocator = parseStructuredLocatorPayload(rule.locator)
      if (structuredLocator && structuredLocator.spec) {
        const spec = structuredLocator.spec
        return !normalizeText(spec.method) || !normalizeText(spec.value)
      }
      const structuredForm = rule.locator_structured_form || {}
      if (!normalizeText(structuredForm.kind)) return true
      if (normalizeText(structuredForm.kind) === 'button_text') {
        return !normalizeText(structuredForm.value) && !normalizeText(structuredForm.target_text)
      }
      return !normalizeText(structuredForm.value)
    })
    if (hasInvalidBoolResultRule) {
      setFieldError(fieldErrors, 'bool_result_rules', '布尔判断规则里有未填写完整的定位，请检查每一条规则。')
    }
  }

  if (type === 'text_content' && isLocatorConfigPayload(item.locator)) {
    const locatorList = Array.isArray(formMeta.text_content_locators) ? formMeta.text_content_locators : []
    const hasInvalidLocator = locatorList.length === 0 || locatorList.some((item) => {
      const normalizedAction = normalizeText(item && item.on_found)
      return validateBaseLocatorMeta(item && item.base_locator) !== ''
        || (normalizedAction !== 'extract_text' && normalizedAction !== 'return_empty')
    })
    if (hasInvalidLocator) {
      setFieldError(fieldErrors, 'locator', '文本提取至少需要 1 条完整规则，且每条规则都要选择“返回其提取”或“返回空值”。')
    }
  }

  if ((type === 'click' || type === 'input') && isLocatorConfigPayload(item.locator)) {
    const locatorList = Array.isArray(formMeta.action_locators) ? formMeta.action_locators : []
    const hasInvalidLocator = locatorList.length === 0 || locatorList.some((item) => validateBaseLocatorMeta(item && item.base_locator) !== '')
    if (hasInvalidLocator) {
      setFieldError(fieldErrors, 'locator', '当前操作至少需要配置 1 个完整的基础定位。')
    }
  }

  return {
    valid: Object.keys(fieldErrors).length === 0,
    fieldErrors,
  }
}

module.exports = {
  PROCESS_ITEM_FIELD_GUIDES,
  PROCESS_TYPE_FIELDS,
  parseCheckConfig,
  parseCheckKeyExpression,
  parseRedirectUriValue,
  parseWaitUrlValue,
  serializeCheckConfig,
  serializeCheckKeyExpression,
  serializeRedirectUriValue,
  serializeWaitUrlValue,
  shouldShowAppendToReplace,
  validateProcessItemForm,
}
