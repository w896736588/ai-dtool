function isMemoryFragmentTabName(tabName) {
  return /^fragment-\d+$/.test(String(tabName || ''))
}

function activateMemorySaveFeedback(currentState, fragmentId, now, durationMs) {
  const normalizedId = String(Number(fragmentId || 0))
  if (normalizedId === '0' || normalizedId === 'NaN') {
    return { ...(currentState || {}) }
  }
  const nextState = { ...(currentState || {}) }
  nextState[normalizedId] = {
    visible: true,
    expiresAt: Number(now || 0) + Number(durationMs || 0),
  }
  return nextState
}

function clearExpiredMemorySaveFeedback(currentState, now) {
  const nextState = {}
  Object.entries(currentState || {}).forEach(([fragmentId, feedback]) => {
    if (!feedback || Number(feedback.expiresAt || 0) <= Number(now || 0)) {
      return
    }
    nextState[fragmentId] = feedback
  })
  return nextState
}

module.exports = {
  isMemoryFragmentTabName,
  activateMemorySaveFeedback,
  clearExpiredMemorySaveFeedback,
}
