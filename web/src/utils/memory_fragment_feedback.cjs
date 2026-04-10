function isMemoryFragmentTabName(tabName) {
  return /^fragment-\d+$/.test(String(tabName || ''))
}

function activateMemorySaveFeedback(currentState, fragmentId, now, durationMs) {
  const rawId = String(fragmentId || '').trim()
  const startedAt = Number(now || 0)
  if (!rawId || rawId === '0' || rawId === 'null' || rawId === 'undefined') {
    return { ...(currentState || {}) }
  }
  const normalizedId = rawId
  const nextState = { ...(currentState || {}) }
  nextState[normalizedId] = {
    visible: true,
    // startedAt 保存本次反馈的启动时间，用来给列表项生成新的 key，确保 CSS 动画能重新播放。
    // startedAt keeps the current feedback cycle timestamp so the sidebar item can remount and replay CSS animation.
    startedAt,
    expiresAt: startedAt + Number(durationMs || 0),
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
