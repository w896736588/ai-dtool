function normalizePorts(ports) {
  return (ports || [])
    .map(port => String(port || '').trim())
    .filter(Boolean)
}

function pickRandomPort(ports, random = Math.random) {
  const normalizedPorts = normalizePorts(ports)
  if (normalizedPorts.length === 0) {
    return ''
  }
  const index = Math.floor(random() * normalizedPorts.length)
  return normalizedPorts[index]
}

function getApiPort(ports, random = Math.random, ssePort = '17170') {
  const normalizedPorts = normalizePorts(ports)
  const normalizedSsePort = String(ssePort)
  const apiPorts = normalizedPorts.filter(port => port !== normalizedSsePort)
  if (apiPorts.length > 0) {
    return pickRandomPort(apiPorts, random)
  }
  return pickRandomPort(normalizedPorts, random)
}

function getSsePort(ports, random = Math.random, ssePort = '17170') {
  const normalizedPorts = normalizePorts(ports)
  const normalizedSsePort = String(ssePort)
  if (normalizedPorts.includes(normalizedSsePort)) {
    return normalizedSsePort
  }
  return pickRandomPort(normalizedPorts, random)
}

module.exports = {
  getApiPort,
  getSsePort,
  normalizePorts,
  pickRandomPort,
}
