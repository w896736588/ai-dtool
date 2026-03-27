// PROCESS_ITEM_DETAIL_LABELS 统一维护流程项列表里使用的详情标签文案。
const PROCESS_ITEM_DETAIL_LABELS = {
  locator: '定位规则',
  value: '执行值',
  outKey: '输出变量',
  checkKey: '执行条件',
  appendToReplace: '写入替换列表',
  waitMills: '等待时长',
}

// STRUCTURED_LOCATOR_KIND_LABELS 用于把结构化 Locator method 映射成更容易理解的说明。
const STRUCTURED_LOCATOR_KIND_LABELS = {
  locator: 'CSS',
  role: '角色',
  text: '文本',
  label: '标签',
  placeholder: '占位文本',
  alt_text: '替代文本',
  title: '标题',
  test_id: 'Test ID',
}

// BOOL_FLAG_FALSE 表示不开启的布尔型字符串值。
const BOOL_FLAG_FALSE = '0'
// DEFAULT_WAIT_MILLS 表示默认等待时长，便于统一展示。
const DEFAULT_WAIT_MILLS = 0

function normalizeText(value) {
  return String(value || '').trim()
}

function safeParseJson(text, fallback) {
  if (!normalizeText(text)) {
    return fallback
  }
  try {
    return JSON.parse(text)
  } catch (error) {
    return fallback
  }
}

function formatBooleanLabel(value) {
  return String(value) === BOOL_FLAG_FALSE ? '否' : '是'
}

function normalizeTokenDisplay(value) {
  const normalizedValue = normalizeText(value)
  if (!normalizedValue) {
    return ''
  }
  if (/^\{.*\}$/.test(normalizedValue)) {
    return normalizedValue
  }
  return `{${normalizedValue}}`
}

function formatStructuredLocator(locatorConfig) {
  const spec = locatorConfig && locatorConfig.spec ? locatorConfig.spec : {}
  const method = normalizeText(spec.method) || 'locator'
  const label = STRUCTURED_LOCATOR_KIND_LABELS[method] || method
  const options = spec.options && typeof spec.options === 'object' ? spec.options : {}
  const pick = spec.pick && typeof spec.pick === 'object' ? spec.pick : {}
  const filters = Array.isArray(spec.filters) ? spec.filters : []
  const chainList = Array.isArray(spec.chain) ? spec.chain : []
  const descriptionParts = []

  if (method === 'role' && normalizeText(spec.value) === 'button') {
    descriptionParts.push(`按钮文字: ${normalizeText(options.name || spec.value) || '未填写'}`)
  } else {
    descriptionParts.push(`${label}: ${normalizeText(spec.value) || '未填写'}`)
  }

  if (normalizeText(options.name) && !(method === 'role' && normalizeText(spec.value) === 'button')) {
    descriptionParts.push(`名称: ${normalizeText(options.name)}`)
  }
  if (options.exact) {
    descriptionParts.push('完全匹配')
  }
  if (spec.negate) {
    descriptionParts.push('要求不存在')
  }
  if (pick.first) {
    descriptionParts.push('取第一个')
  } else if (pick.last) {
    descriptionParts.push('取最后一个')
  } else if (Number.isInteger(Number(pick.nth))) {
    descriptionParts.push(`取第 ${Number(pick.nth)} 个`)
  }
  if (Number(spec.timeout_mills) > DEFAULT_WAIT_MILLS) {
    descriptionParts.push(`超时 ${Number(spec.timeout_mills)}ms`)
  }
  filters.forEach((item) => {
    if (!(item && typeof item === 'object')) return
    if (normalizeText(item.has_text)) {
      descriptionParts.push(`且包含文本: ${normalizeText(item.has_text)}`)
    }
    if (normalizeText(item.has_not_text)) {
      descriptionParts.push(`且不包含文本: ${normalizeText(item.has_not_text)}`)
    }
    if (item.has && typeof item.has === 'object') {
      descriptionParts.push(`且包含: ${formatStructuredLocator({ spec: item.has })}`)
    }
    if (item.has_not && typeof item.has_not === 'object') {
      descriptionParts.push(`且不包含: ${formatStructuredLocator({ spec: item.has_not })}`)
    }
    if (typeof item.visible === 'boolean') {
      descriptionParts.push(item.visible ? '要求可见' : '要求不可见')
    }
  })
  chainList.forEach((item) => {
    if (!(item && typeof item === 'object')) return
    descriptionParts.push(`再向下查找: ${formatStructuredLocator({ spec: item })}`)
  })

  return descriptionParts.join('，')
}

function formatRawLocator(locatorValue) {
  const normalizedValue = normalizeText(locatorValue)
  if (!normalizedValue) {
    return []
  }

  const boolResultRuleList = safeParseJson(normalizedValue, null)
  if (Array.isArray(boolResultRuleList)) {
    return boolResultRuleList
      .filter((item) => item && item.locator)
      .map((item, index) => `${index + 1}. ${formatStructuredLocator(item.locator)}，命中返回 ${item.return === false ? 'false' : 'true'}`)
  }

  const structuredLocator = safeParseJson(normalizedValue, null)
  if (structuredLocator && structuredLocator.spec) {
    return [formatStructuredLocator(structuredLocator)]
  }

  return normalizedValue
    .split(/\r?\n/)
    .map((item) => normalizeText(item))
    .filter(Boolean)
}

// buildProcessItemDisplayDetails 用于把流程项明细转换成适合列表卡片展示的标签化结构。
function buildProcessItemDisplayDetails(item) {
  const detailList = []
  const locatorLineList = formatRawLocator(item && item.locator)
  const valueText = normalizeText(item && item.value)
  const outKeyText = normalizeTokenDisplay(item && item.out_key)
  const checkKeyText = normalizeText(item && item.check_key)
  const waitMillsValue = Number(item && item.wait_mills)

  if (locatorLineList.length > 0) {
    detailList.push({
      key: 'locator',
      label: PROCESS_ITEM_DETAIL_LABELS.locator,
      lines: locatorLineList,
      emphasis: locatorLineList.length > 1 ? 'block' : 'normal',
    })
  }
  if (valueText) {
    detailList.push({
      key: 'value',
      label: PROCESS_ITEM_DETAIL_LABELS.value,
      lines: [valueText],
      emphasis: valueText.length > 60 ? 'block' : 'normal',
    })
  }
  if (outKeyText) {
    detailList.push({
      key: 'out_key',
      label: PROCESS_ITEM_DETAIL_LABELS.outKey,
      lines: [outKeyText],
      emphasis: 'accent',
    })
  }
  if (checkKeyText) {
    detailList.push({
      key: 'check_key',
      label: PROCESS_ITEM_DETAIL_LABELS.checkKey,
      lines: [checkKeyText],
      emphasis: 'normal',
    })
  }
  if (outKeyText) {
    detailList.push({
      key: 'append_to_replace',
      label: PROCESS_ITEM_DETAIL_LABELS.appendToReplace,
      lines: [formatBooleanLabel(item && item.append_to_replace)],
      emphasis: 'normal',
    })
  }
  if (waitMillsValue > DEFAULT_WAIT_MILLS) {
    detailList.push({
      key: 'wait_mills',
      label: PROCESS_ITEM_DETAIL_LABELS.waitMills,
      lines: [`${waitMillsValue}ms`],
      emphasis: 'normal',
    })
  }

  return detailList
}

module.exports = {
  buildProcessItemDisplayDetails,
  formatRawLocator,
  formatStructuredLocator,
  normalizeTokenDisplay,
}
