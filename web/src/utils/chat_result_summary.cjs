function toFiniteNumber(value) {
  const num = Number(value)
  return Number.isFinite(num) ? num : 0
}

function formatSeconds(durationMs) {
  const ms = toFiniteNumber(durationMs)
  if (ms <= 0) return ''
  return (ms / 1000).toFixed(1) + 's'
}

function buildResultSummaryLines(msg) {
  const lines = []
  const isError = !!(msg && msg.isError)
  lines.push(isError ? '✕ 执行失败' : '✓ 执行完成')

  const durationText = formatSeconds(msg && msg.durationMs)
  const numTurns = toFiniteNumber(msg && msg.numTurns)
  if (durationText && numTurns > 0) {
    lines.push('耗时 ' + durationText + '  ' + numTurns + '轮对话')
  } else if (durationText) {
    lines.push('耗时 ' + durationText)
  } else if (numTurns > 0) {
    lines.push(numTurns + '轮对话')
  }

  return lines
}

module.exports = {
  buildResultSummaryLines,
  formatSeconds,
  toFiniteNumber,
}
