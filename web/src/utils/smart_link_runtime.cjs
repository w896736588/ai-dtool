// DEFAULT_RUNTIME_CONFIG 定义自定义网页运行时配置的默认结构。
// DEFAULT_RUNTIME_CONFIG defines the default smart-link runtime config shape.
const DEFAULT_RUNTIME_CONFIG = {
  run_mode: 'server',
  required_client_version: '1.0.0',
  build_platforms: ['windows', 'macos'],
}

// normalizeRuntimeConfig 统一接口返回，避免缺字段时前端状态判断失真。
// normalizeRuntimeConfig normalizes API payloads so missing fields do not break UI state checks.
function normalizeRuntimeConfig(nextRuntimeConfig) {
  return {
    ...DEFAULT_RUNTIME_CONFIG,
    ...(nextRuntimeConfig || {}),
    build_platforms: Array.isArray(nextRuntimeConfig && nextRuntimeConfig.build_platforms)
      ? nextRuntimeConfig.build_platforms
      : DEFAULT_RUNTIME_CONFIG.build_platforms,
  }
}

// resolveRuntimeRefreshActions 根据最新运行模式决定是否立刻刷新客户端状态。
// resolveRuntimeRefreshActions decides whether the UI should immediately refresh local client status.
function resolveRuntimeRefreshActions(currentRuntimeConfig, nextRuntimeConfig) {
  const runtimeConfig = normalizeRuntimeConfig(nextRuntimeConfig)

  // 本地客户端模式必须马上补拉一次状态，否则从设置页切回时会继续显示旧模式。
  // Local-client mode must trigger an immediate status refresh, otherwise the page can stay on stale server-mode UI.
  const shouldLoadClientStatus = runtimeConfig.run_mode === 'local_client'

  return {
    runtimeConfig,
    shouldLoadClientStatus,
  }
}

// buildRuntimeApiUrl 统一拼接本地客户端模式下的接口地址，避免 dev 模式误打到 8080 前端服务。
// buildRuntimeApiUrl builds runtime API URLs so development mode can target the fixed backend host instead of the Vue dev server.
function buildRuntimeApiUrl(apiHost, uri) {
  const normalizedApiHost = String(apiHost || '').trim().replace(/\/+$/, '')
  const normalizedUri = String(uri || '').trim()
  if (!normalizedApiHost) {
    return normalizedUri
  }
  return `${normalizedApiHost}${normalizedUri}`
}

// buildRuntimeRequestOptions 统一拼装运行时 fetch 参数，确保 token 头在 dev 和 prod 模式都透传。
// buildRuntimeRequestOptions builds fetch options for runtime requests so auth headers are preserved in both dev and prod.
function buildRuntimeRequestOptions(safeToken, requestOptions) {
  const nextOptions = {
    ...(requestOptions || {}),
    headers: {
      ...((requestOptions && requestOptions.headers) || {}),
    },
  }

  // Token 头沿用现有 BasePost 约定，避免 smart-link 新接口绕过登录态。
  // Reuse the existing Token header contract so smart-link runtime requests do not bypass authentication state.
  if (safeToken) {
    nextOptions.headers.Token = safeToken
  }

  return nextOptions
}

module.exports = {
  DEFAULT_RUNTIME_CONFIG,
  buildRuntimeApiUrl,
  buildRuntimeRequestOptions,
  normalizeRuntimeConfig,
  resolveRuntimeRefreshActions,
}
