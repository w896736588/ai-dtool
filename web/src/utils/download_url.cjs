// buildDownloadUrlWithToken 为下载链接附加地址栏 token，便于浏览器直链下载时通过后端鉴权。
function buildDownloadUrlWithToken(url, token) {
  const normalizedUrl = String(url || '').trim()
  const normalizedToken = String(token || '').trim()
  if (!normalizedUrl) {
    return ''
  }
  if (!normalizedToken) {
    return normalizedUrl
  }

  const parsedUrl = new URL(normalizedUrl, 'http://localhost')
  parsedUrl.searchParams.set('token', normalizedToken)

  if (/^https?:\/\//i.test(normalizedUrl)) {
    return parsedUrl.toString()
  }
  return parsedUrl.pathname + parsedUrl.search + parsedUrl.hash
}

module.exports = {
  buildDownloadUrlWithToken,
}
