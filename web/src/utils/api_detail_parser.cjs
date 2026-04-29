// DEFAULT_API_OBJECT 中文注释：接口详情对象类型字段的默认值，兼容空字符串和非法 JSON。
// DEFAULT_API_OBJECT English comment: Fallback object value for API detail fields when payload is empty or invalid.
const DEFAULT_API_OBJECT = {}

// DEFAULT_API_ARRAY 中文注释：接口详情数组类型字段的默认值，兼容空字符串和非法 JSON。
// DEFAULT_API_ARRAY English comment: Fallback array value for API detail fields when payload is empty or invalid.
const DEFAULT_API_ARRAY = []

// cloneObjectFallback 中文注释：对象默认值需要浅拷贝，避免多个字段共享同一引用。
// cloneObjectFallback English comment: Clone object defaults so callers do not accidentally share mutable references.
function cloneObjectFallback(fallbackValue) {
  return fallbackValue && typeof fallbackValue === 'object' ? { ...fallbackValue } : {}
}

// cloneArrayFallback 中文注释：数组默认值需要浅拷贝，避免多个字段共享同一引用。
// cloneArrayFallback English comment: Clone array defaults so callers do not accidentally share mutable references.
function cloneArrayFallback(fallbackValue) {
  return Array.isArray(fallbackValue) ? [...fallbackValue] : []
}

// parseApiObjectField 中文注释：统一解析对象类字段，兼容对象本身、JSON 字符串、空字符串和异常数据。
// parseApiObjectField English comment: Safely parse object-like API fields from raw objects, JSON strings, or empty payloads.
function parseApiObjectField(rawValue, fallbackValue = DEFAULT_API_OBJECT) {
  if (rawValue === null || rawValue === undefined || rawValue === '') {
    return cloneObjectFallback(fallbackValue)
  }
  if (typeof rawValue === 'object' && !Array.isArray(rawValue)) {
    return { ...rawValue }
  }
  if (typeof rawValue !== 'string') {
    return cloneObjectFallback(fallbackValue)
  }
  try {
    const parsedValue = JSON.parse(rawValue)
    return parsedValue && typeof parsedValue === 'object' && !Array.isArray(parsedValue)
      ? parsedValue
      : cloneObjectFallback(fallbackValue)
  } catch (error) {
    return cloneObjectFallback(fallbackValue)
  }
}

// parseApiArrayField 中文注释：统一解析数组类字段，兼容数组本身、JSON 字符串、空字符串和异常数据。
// parseApiArrayField English comment: Safely parse array-like API fields from raw arrays, JSON strings, or empty payloads.
function parseApiArrayField(rawValue, fallbackValue = DEFAULT_API_ARRAY) {
  if (rawValue === null || rawValue === undefined || rawValue === '') {
    return cloneArrayFallback(fallbackValue)
  }
  if (Array.isArray(rawValue)) {
    return [...rawValue]
  }
  if (typeof rawValue !== 'string') {
    return cloneArrayFallback(fallbackValue)
  }
  try {
    const parsedValue = JSON.parse(rawValue)
    return Array.isArray(parsedValue) ? parsedValue : cloneArrayFallback(fallbackValue)
  } catch (error) {
    return cloneArrayFallback(fallbackValue)
  }
}

module.exports = {
  parseApiObjectField,
  parseApiArrayField,
}
