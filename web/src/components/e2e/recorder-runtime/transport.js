export class RecorderTransport {
  constructor({ baseUrl, wsToken, getIframe }) {
    this.baseUrl = baseUrl
    this.wsToken = wsToken
    this.getIframe = getIframe
  }

  async call(path, body) {
    const iframe = this.getIframe()
    const f = iframe && iframe.contentWindow
    if (!f) throw new Error('iframe proxy 尚未挂载')
    const res = await f.fetch(`${this.baseUrl}${path}?ws_token=${encodeURIComponent(this.wsToken)}`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
    })
    if (!res.ok) throw new Error(`record api ${path} failed: ${res.status}`)
    return res.json()
  }

  addStep(step) { return this.call('/api/e2e/record/by_token/step/add', { step }) }
  commit(req) { return this.call('/api/e2e/record/by_token/commit', req) }
}