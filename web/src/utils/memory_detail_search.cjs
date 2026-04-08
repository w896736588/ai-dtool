function normalizeSearchText(value) {
  return String(value || '')
}

function normalizeSearchQuery(query) {
  return normalizeSearchText(query).trim()
}

function buildFieldMatches(field, text, query) {
  const sourceText = normalizeSearchText(text)
  const normalizedQuery = normalizeSearchQuery(query)
  if (normalizedQuery === '') {
    return []
  }

  const lowerText = sourceText.toLocaleLowerCase()
  const lowerQuery = normalizedQuery.toLocaleLowerCase()
  const matches = []
  let fromIndex = 0

  while (fromIndex <= lowerText.length - lowerQuery.length) {
    const index = lowerText.indexOf(lowerQuery, fromIndex)
    if (index === -1) {
      break
    }
    matches.push({
      field,
      index,
      end: index + normalizedQuery.length,
      text: sourceText.slice(index, index + normalizedQuery.length),
    })
    fromIndex = index + lowerQuery.length
  }

  return matches
}

function buildMemoryDetailSearchMatches(title, content, query) {
  return [
    ...buildFieldMatches('title', title, query),
    ...buildFieldMatches('content', content, query),
  ]
}

function normalizeActiveMatchIndex(matches, currentIndex) {
  if (!Array.isArray(matches) || matches.length === 0) {
    return -1
  }
  const normalizedIndex = Number(currentIndex)
  if (Number.isInteger(normalizedIndex) && normalizedIndex >= 0 && normalizedIndex < matches.length) {
    return normalizedIndex
  }
  return 0
}

function getNextMemoryDetailMatchIndex(matches, currentIndex, step) {
  if (!Array.isArray(matches) || matches.length === 0) {
    return -1
  }
  const normalizedStep = Number(step) < 0 ? -1 : 1
  const normalizedIndex = normalizeActiveMatchIndex(matches, currentIndex)
  return (normalizedIndex + normalizedStep + matches.length) % matches.length
}

module.exports = {
  buildMemoryDetailSearchMatches,
  getNextMemoryDetailMatchIndex,
  normalizeActiveMatchIndex,
}
