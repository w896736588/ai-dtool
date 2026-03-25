function normalizeDockerDefaultServices(rawValue) {
  if (Array.isArray(rawValue)) {
    return rawValue
      .map(item => String(item || '').trim())
      .filter(Boolean)
      .filter((item, index, list) => list.indexOf(item) === index)
  }

  return String(rawValue || '')
    .split(',')
    .map(item => item.trim())
    .filter(Boolean)
    .filter((item, index, list) => list.indexOf(item) === index)
}

function stringifyDockerDefaultServices(services) {
  return normalizeDockerDefaultServices(services).join(',')
}

function isDockerDefaultServiceEnabled(defaultService, serviceName) {
  if (!serviceName) {
    return false
  }
  return normalizeDockerDefaultServices(defaultService).includes(String(serviceName).trim())
}

function toggleDockerDefaultService(defaultService, serviceName, enabled) {
  const normalizedServiceName = String(serviceName || '').trim()
  const services = normalizeDockerDefaultServices(defaultService)
  if (!normalizedServiceName) {
    return stringifyDockerDefaultServices(services)
  }

  const exists = services.includes(normalizedServiceName)
  if (enabled && !exists) {
    services.push(normalizedServiceName)
  }
  if (!enabled && exists) {
    return stringifyDockerDefaultServices(services.filter(item => item !== normalizedServiceName))
  }

  return stringifyDockerDefaultServices(services)
}

module.exports = {
  isDockerDefaultServiceEnabled,
  normalizeDockerDefaultServices,
  stringifyDockerDefaultServices,
  toggleDockerDefaultService,
}
