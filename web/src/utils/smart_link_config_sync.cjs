function safeParseJsonArray(value) {
  if (!value) return []
  if (Array.isArray(value)) return value
  try {
    const parsed = JSON.parse(value)
    return Array.isArray(parsed) ? parsed : []
  } catch (error) {
    return []
  }
}

function normalizeSmartLinkRecord(record) {
  const normalized = {
    ...(record || {}),
  }
  normalized.open_type = Number(normalized.open_type || 0)
  normalized.open_type_new = normalized.open_type
  normalized.open_num = Number(normalized.open_num || 0)
  normalized.open_num_new = normalized.open_num
  normalized.combine_type = String(normalized.combine_type || '')
  normalized.channel = normalized.channel || ''
  normalized.linkList = safeParseJsonArray(normalized.links)
  return normalized
}

function mergeSavedSmartLinkIntoList(currentList, savedRecord) {
  const sourceList = Array.isArray(currentList) ? currentList : []
  const normalizedSavedRecord = normalizeSmartLinkRecord(savedRecord)
  let hasMatched = false
  const nextList = sourceList.map(item => {
    if (Number(item && item.id) !== Number(normalizedSavedRecord.id)) {
      return item
    }
    hasMatched = true
    return normalizedSavedRecord
  })
  if (!hasMatched) {
    nextList.push(normalizedSavedRecord)
  }
  return nextList
}

module.exports = {
  mergeSavedSmartLinkIntoList,
  normalizeSmartLinkRecord,
}
