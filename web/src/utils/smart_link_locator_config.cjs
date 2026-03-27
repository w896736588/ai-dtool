const {
  buildAdvancedLocatorPayload,
  buildSimpleLocatorPayload,
  createAdvancedLocatorForm,
  createSimpleLocatorForm,
  deserializeLocatorEditorState,
  parseStructuredLocatorPayload,
} = require('./smart_link_locator_form.cjs')

function normalizeText(value) {
  return String(value || '').trim()
}

function createBaseLocatorMeta() {
  return {
    id: '',
    summary: '',
    locator_editor_mode: 'simple',
    locator_structured_form: createSimpleLocatorForm(),
    locator_advanced_form: createAdvancedLocatorForm(),
  }
}

function createBaseLocatorFromPayload(payload, fallback = {}) {
  const editorState = deserializeLocatorEditorState(payload, {
    preferAdvanced: fallback.locator_editor_mode === 'advanced',
  })
  return {
    ...createBaseLocatorMeta(),
    ...fallback,
    locator_editor_mode: editorState.mode,
    locator_structured_form: editorState.simpleForm,
    locator_advanced_form: editorState.advancedForm,
  }
}

function buildBaseLocatorQuery(baseLocator) {
  const meta = {
    ...createBaseLocatorMeta(),
    ...(baseLocator || {}),
  }
  return meta.locator_editor_mode === 'advanced'
    ? buildAdvancedLocatorPayload(meta.locator_advanced_form)
    : buildSimpleLocatorPayload(meta.locator_structured_form)
}

function buildLocatorConfigByType(type, formMeta = {}) {
  if (type === 'bool_result') {
    return {
      version: 2,
      mode: 'bool_result',
      strategy: 'first_match_return',
      locators: (Array.isArray(formMeta.bool_result_rules) ? formMeta.bool_result_rules : [])
        .map((rule, index) => ({
          id: normalizeText(rule && rule.id) || `rule_${index + 1}`,
          query: buildBaseLocatorQuery(rule && rule.base_locator),
          on_found: rule && rule.on_found !== false,
        })),
      options: {},
    }
  }

  if (type === 'text_content') {
    return {
      version: 2,
      mode: 'text_content',
      strategy: 'first_match_return',
      locators: (Array.isArray(formMeta.text_content_locators) ? formMeta.text_content_locators : [])
        .map((item, index) => ({
          id: normalizeText(item && item.id) || `text_${index + 1}`,
          query: buildBaseLocatorQuery(item && item.base_locator),
          on_found: normalizeText(item && item.on_found) || 'extract_text',
        })),
      options: {
        extract_type: 'text_content',
      },
    }
  }

  if (type === 'click' || type === 'input') {
    return {
      version: 2,
      mode: type,
      strategy: normalizeText(formMeta.action_strategy) || 'first_found_do_action',
      locators: (Array.isArray(formMeta.action_locators) ? formMeta.action_locators : [])
        .map((item, index) => ({
          id: normalizeText(item && item.id) || `action_${index + 1}`,
          query: buildBaseLocatorQuery(item && item.base_locator),
        })),
      options: {
        action_type: type,
      },
    }
  }

  return null
}

function isLocatorConfigPayload(rawValue) {
  const payload = rawValue && typeof rawValue === 'string' ? safeParseJson(rawValue, null) : rawValue
  return Boolean(
    payload
    && typeof payload === 'object'
    && Number(payload.version) === 2
    && normalizeText(payload.mode)
    && Array.isArray(payload.locators)
  )
}

function deserializeLocatorConfigToFormMeta(rawValue) {
  const payload = rawValue && typeof rawValue === 'string' ? safeParseJson(rawValue, null) : rawValue
  if (!isLocatorConfigPayload(payload)) {
    return null
  }

  if (payload.mode === 'bool_result') {
    return {
      locator_config_mode: 'bool_result',
      bool_result_rules: payload.locators.map((item) => ({
        id: item.id || '',
        on_found: item.on_found !== false,
        base_locator: createBaseLocatorFromPayload(item.query),
      })),
    }
  }

  if (payload.mode === 'text_content') {
    return {
      locator_config_mode: 'text_content',
      text_content_locators: payload.locators.map((item) => ({
        id: item.id || '',
        on_found: normalizeText(item.on_found) || 'extract_text',
        base_locator: createBaseLocatorFromPayload(item.query, { locator_editor_mode: 'advanced' }),
      })),
    }
  }

  if (payload.mode === 'click' || payload.mode === 'input') {
    return {
      locator_config_mode: payload.mode,
      action_strategy: payload.strategy || 'first_found_do_action',
      action_locators: payload.locators.map((item) => ({
        id: item.id || '',
        base_locator: createBaseLocatorFromPayload(item.query),
      })),
    }
  }

  return null
}

function safeParseJson(text, fallback) {
  if (!normalizeText(text)) return fallback
  try {
    return JSON.parse(text)
  } catch (error) {
    return fallback
  }
}

module.exports = {
  buildBaseLocatorQuery,
  buildLocatorConfigByType,
  createBaseLocatorFromPayload,
  createBaseLocatorMeta,
  deserializeLocatorConfigToFormMeta,
  isLocatorConfigPayload,
  parseStructuredLocatorPayload,
}
