function normalizeHomeTaskFragmentOption(fragment) {
  const fragmentId = String(
    (fragment && (fragment.file_id || fragment.id)) || ''
  ).trim()
  if (!fragmentId || fragmentId === '0') {
    return null
  }
  return {
    id: fragmentId,
    title: String((fragment && fragment.title) || '').trim() || `#${fragmentId}`,
    tags: Array.isArray(fragment && fragment.tags) ? fragment.tags : [],
  }
}

function mergeHomeTaskFragmentOptions(fragments, selectedFragment) {
  const normalizedList = Array.isArray(fragments)
    ? fragments.map((item) => normalizeHomeTaskFragmentOption(item)).filter(Boolean)
    : []
  const currentFragment = normalizeHomeTaskFragmentOption(selectedFragment)
  if (!currentFragment) {
    return normalizedList
  }
  const exists = normalizedList.some((item) => item.id === currentFragment.id)
  if (exists) {
    return normalizedList
  }
  return [currentFragment, ...normalizedList]
}

module.exports = {
  normalizeHomeTaskFragmentOption,
  mergeHomeTaskFragmentOptions,
}
