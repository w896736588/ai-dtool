function normalizeWhitespace(value) {
  return String(value || '').replace(/\s+/g, ' ').trim()
}

function normalizeSupervisorProgramNameFromHeader(headerText) {
  const normalizedHeader = String(headerText || '')
    .replace(/\r/g, '')
    .trim()
    .replace(/^\[/, '')
    .replace(/\]$/, '')
  if (!normalizedHeader) {
    return ''
  }
  const colonIndex = normalizedHeader.indexOf(':')
  if (colonIndex >= 0) {
    return normalizedHeader.slice(colonIndex + 1).trim()
  }
  return normalizedHeader
}

function parseSupervisorStatusLine(lineText) {
  const cleanedLine = String(lineText || '').replace(/\r/g, '').trim()
  if (!cleanedLine) {
    return null
  }
  const statusMatch = cleanedLine.match(/^([^\s]+)\s+(RUNNING|FATAL|STOPPED|BACKOFF|STARTING|EXITED|STOPPING)\s+(.+)$/)
  if (!statusMatch) {
    return null
  }
  const processName = String(statusMatch[1] || '').trim()
  if (!processName) {
    return null
  }
  return {
    processName,
    groupName: processName.split(':')[0].trim(),
    statusText: normalizeWhitespace(`${statusMatch[2]} ${statusMatch[3]}`),
  }
}

module.exports = {
  normalizeSupervisorProgramNameFromHeader,
  parseSupervisorStatusLine,
}
