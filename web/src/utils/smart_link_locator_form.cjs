function normalizeText(value) {
  return String(value || '').trim()
}

function safeParseJson(text, fallback) {
  if (!text) return fallback
  try {
    return JSON.parse(text)
  } catch (error) {
    return fallback
  }
}

function createSimpleLocatorForm() {
  return {
    kind: 'css',
    method: 'locator',
    value: '',
    target_text: '',
    exact: false,
    negate: false,
    pick_mode: 'none',
    nth: 0,
    timeout_mills: 3000,
  }
}

function createAdvancedLocatorForm() {
  return {
    ...createSimpleLocatorForm(),
    has_text: '',
    has_not_text: '',
    has_kind: '',
    has_value: '',
    has_not_kind: '',
    has_not_value: '',
    visible: '',
    chain_kind: '',
    chain_value: '',
    chain_target_text: '',
  }
}

function parseStructuredLocatorPayload(rawValue) {
  if (!rawValue) return null
  if (typeof rawValue === 'object') {
    return rawValue
  }
  if (typeof rawValue !== 'string') return null
  const parsed = safeParseJson(rawValue, null)
  if (!parsed || typeof parsed !== 'object') return null
  if (parsed.spec && typeof parsed.spec === 'object') return parsed
  if (parsed.method || parsed.value !== undefined) {
    return { spec: parsed }
  }
  return null
}

function resolveSimpleMethod(kind) {
  if (kind === 'button_text') return 'role'
  if (kind === 'text') return 'text'
  if (kind === 'label') return 'label'
  if (kind === 'placeholder') return 'placeholder'
  if (kind === 'alt_text') return 'alt_text'
  if (kind === 'title') return 'title'
  if (kind === 'test_id') return 'test_id'
  return 'locator'
}

function resolveKindBySpec(spec = {}) {
  if (spec.method === 'role' && spec.value === 'button') return 'button_text'
  if (spec.method === 'text') return 'text'
  if (spec.method === 'label') return 'label'
  if (spec.method === 'placeholder') return 'placeholder'
  if (spec.method === 'alt_text') return 'alt_text'
  if (spec.method === 'title') return 'title'
  if (spec.method === 'test_id') return 'test_id'
  return 'css'
}

function buildQuerySpec(formValue, includeTimeout = true) {
  const form = formValue || createSimpleLocatorForm()
  const kind = normalizeText(form.kind) || 'css'
  const method = resolveSimpleMethod(kind)
  const isButtonText = kind === 'button_text'
  const targetText = normalizeText(form.target_text || form.value)
  const spec = {
    method,
    value: isButtonText ? 'button' : normalizeText(form.value),
  }

  const options = {}
  if (isButtonText && targetText) {
    options.name = targetText
  }
  if (!isButtonText && normalizeText(form.target_text)) {
    options.name = normalizeText(form.target_text)
  }
  if (form.exact) {
    options.exact = true
  }
  if (Object.keys(options).length > 0) {
    spec.options = options
  }
  if (form.negate) {
    spec.negate = true
  }
  if (includeTimeout && Number(form.timeout_mills) > 0) {
    spec.timeout_mills = Number(form.timeout_mills)
  }
  if (form.pick_mode === 'first') {
    spec.pick = { first: true }
  } else if (form.pick_mode === 'last') {
    spec.pick = { last: true }
  } else if (form.pick_mode === 'nth') {
    spec.pick = { nth: Number(form.nth || 0) }
  }

  return spec
}

function buildSimpleLocatorPayload(formValue) {
  return {
    spec: buildQuerySpec(formValue, true),
  }
}

function buildChildSpec(kind, value, targetText) {
  const normalizedValue = normalizeText(value)
  if (!normalizedValue) return null
  return buildQuerySpec({
    kind: normalizeText(kind) || 'css',
    value: normalizedValue,
    target_text: normalizeText(targetText),
    exact: false,
    negate: false,
    pick_mode: 'none',
    nth: 0,
    timeout_mills: 0,
  }, false)
}

function buildAdvancedLocatorPayload(formValue) {
  const form = {
    ...createAdvancedLocatorForm(),
    ...(formValue || {}),
  }
  const spec = buildQuerySpec(form, true)
  const filters = []
  const hasText = normalizeText(form.has_text)
  const hasNotText = normalizeText(form.has_not_text)
  const hasSpec = buildChildSpec(form.has_kind, form.has_value)
  const hasNotSpec = buildChildSpec(form.has_not_kind, form.has_not_value)
  const visibleValue = normalizeText(form.visible)

  if (hasText || hasNotText || hasSpec || hasNotSpec || visibleValue) {
    const filterItem = {}
    if (hasText) {
      filterItem.has_text = hasText
    }
    if (hasNotText) {
      filterItem.has_not_text = hasNotText
    }
    if (hasSpec) {
      filterItem.has = hasSpec
    }
    if (hasNotSpec) {
      filterItem.has_not = hasNotSpec
    }
    if (visibleValue === 'true') {
      filterItem.visible = true
    } else if (visibleValue === 'false') {
      filterItem.visible = false
    }
    filters.push(filterItem)
  }

  const chainSpec = buildChildSpec(form.chain_kind, form.chain_value, form.chain_target_text)
  if (filters.length > 0) {
    spec.filters = filters
  }
  if (chainSpec) {
    spec.chain = [chainSpec]
  }

  return { spec }
}

function isAdvancedLocatorPayload(payload) {
  const parsed = parseStructuredLocatorPayload(payload)
  const spec = parsed && parsed.spec ? parsed.spec : {}
  return Array.isArray(spec.filters) && spec.filters.length > 0
    || Array.isArray(spec.chain) && spec.chain.length > 0
}

function applyPickToForm(form, spec = {}) {
  const pick = spec.pick && typeof spec.pick === 'object' ? spec.pick : {}
  if (pick.first) {
    form.pick_mode = 'first'
  } else if (pick.last) {
    form.pick_mode = 'last'
  } else if (Number.isInteger(Number(pick.nth))) {
    form.pick_mode = 'nth'
    form.nth = Number(pick.nth)
  }
}

function deserializeSimpleLocatorForm(payload) {
  const form = createSimpleLocatorForm()
  const parsed = parseStructuredLocatorPayload(payload) || {}
  const spec = parsed.spec && typeof parsed.spec === 'object' ? parsed.spec : {}
  const options = spec.options && typeof spec.options === 'object' ? spec.options : {}

  form.kind = resolveKindBySpec(spec)
  form.method = spec.method || 'locator'
  form.value = form.kind === 'button_text' ? normalizeText(options.name || '') : normalizeText(spec.value)
  form.target_text = normalizeText(options.name)
  form.exact = Boolean(options.exact)
  form.negate = Boolean(spec.negate)
  form.timeout_mills = Number(spec.timeout_mills ?? 3000)
  applyPickToForm(form, spec)

  return form
}

function deserializeAdvancedLocatorForm(payload) {
  const form = {
    ...createAdvancedLocatorForm(),
    ...deserializeSimpleLocatorForm(payload),
  }
  const parsed = parseStructuredLocatorPayload(payload) || {}
  const spec = parsed.spec && typeof parsed.spec === 'object' ? parsed.spec : {}
  const filters = Array.isArray(spec.filters) ? spec.filters : []
  const firstFilter = filters[0] && typeof filters[0] === 'object' ? filters[0] : {}
  const chainList = Array.isArray(spec.chain) ? spec.chain : []
  const firstChain = chainList[0] && typeof chainList[0] === 'object' ? chainList[0] : {}

  form.has_text = normalizeText(firstFilter.has_text)
  form.has_not_text = normalizeText(firstFilter.has_not_text)
  if (firstFilter.has) {
    form.has_kind = resolveKindBySpec(firstFilter.has)
    form.has_value = normalizeText(firstFilter.has.value)
  }
  if (firstFilter.has_not) {
    form.has_not_kind = resolveKindBySpec(firstFilter.has_not)
    form.has_not_value = normalizeText(firstFilter.has_not.value)
  }
  if (typeof firstFilter.visible === 'boolean') {
    form.visible = String(firstFilter.visible)
  }
  if (firstChain.method) {
    form.chain_kind = resolveKindBySpec(firstChain)
    form.chain_value = normalizeText(firstChain.value)
    const chainOptions = firstChain.options && typeof firstChain.options === 'object' ? firstChain.options : {}
    form.chain_target_text = normalizeText(chainOptions.name)
  }

  return form
}

function deserializeLocatorEditorState(payload, options = {}) {
  const preferAdvanced = Boolean(options.preferAdvanced)
  const parsed = parseStructuredLocatorPayload(payload)
  const advanced = isAdvancedLocatorPayload(parsed)
  const mode = preferAdvanced || advanced ? 'advanced' : 'simple'

  return {
    mode,
    payload: parsed,
    simpleForm: deserializeSimpleLocatorForm(parsed),
    advancedForm: deserializeAdvancedLocatorForm(parsed),
  }
}

function stringifyLocatorPayload(payload) {
  return JSON.stringify(payload || { spec: { method: 'locator', value: '' } }, null, 2)
}

module.exports = {
  buildAdvancedLocatorPayload,
  buildSimpleLocatorPayload,
  createAdvancedLocatorForm,
  createSimpleLocatorForm,
  deserializeAdvancedLocatorForm,
  deserializeLocatorEditorState,
  deserializeSimpleLocatorForm,
  isAdvancedLocatorPayload,
  parseStructuredLocatorPayload,
  stringifyLocatorPayload,
}
