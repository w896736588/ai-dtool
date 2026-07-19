import { RecorderTransport } from './transport'
import { buildSelectorChain, viewportRelativeCoords } from './dom-helpers'

const TOOLBAR_HTML = `
<div data-dtool-recorder-toolbar style="position:fixed;top:80px;right:20px;z-index:2147483647;background:#fff;border-radius:8px;box-shadow:0 4px 16px rgba(0,0,0,.18);width:340px;font:12px/1.4 -apple-system,BlinkMacSystemFont,'Segoe UI',sans-serif;color:#303133;">
  <div style="padding:8px 10px;color:#fff;background:linear-gradient(90deg,#409eff,#66b1ff);border-radius:8px 8px 0 0;font-weight:600;">录制工具条 <span data-stat></span></div>
  <div style="padding:8px 10px;display:flex;gap:6px;flex-wrap:wrap;">
    <button data-mode="click" style="background:#409eff;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">元素点击</button>
    <button data-mode="click_xy" style="background:#409eff;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">坐标点击</button>
    <button data-mode="input" style="background:#409eff;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">输入</button>
    <button data-mode="scroll" style="background:#409eff;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">滚动</button>
    <button data-commit style="background:#67c23a;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">结束并提交</button>
    <button data-close style="background:#f56c6c;color:#fff;border:0;border-radius:4px;padding:6px 8px;cursor:pointer;">放弃</button>
  </div>
</div>`

function ensureToolbar() {
  if (document.querySelector('[data-dtool-recorder-toolbar]')) return null
  const root = document.createElement('div')
  root.innerHTML = TOOLBAR_HTML
  document.body.appendChild(root.firstElementChild)
  return document.querySelector('[data-dtool-recorder-toolbar]')
}

async function bootRecorder(opts) {
  const proxyIframe = document.querySelector('iframe[src*="/api/e2e/recorder/proxy.html"]')
  if (!proxyIframe) {
    console.warn('[recorder] proxy iframe 未找到')
    return
  }
  const transport = new RecorderTransport({
    baseUrl: window.location.origin,
    wsToken: opts.wsToken,
    getIframe: () => proxyIframe,
  })
  await new Promise((resolve) => {
    if (proxyIframe.contentDocument && proxyIframe.contentDocument.readyState === 'complete') resolve()
    else proxyIframe.addEventListener('load', () => resolve())
  })

  const toolbar = ensureToolbar()
  if (!toolbar) return
  let mode = 'click'
  let steps = 0
  const stat = toolbar.querySelector('[data-stat]')
  const update = () => { stat.textContent = `${steps} 步 · ${mode}` }
  update()

  toolbar.querySelectorAll('button[data-mode]').forEach((b) => {
    b.addEventListener('click', (ev) => {
      ev.stopPropagation()
      mode = b.dataset.mode
      update()
    })
  })

  document.addEventListener('click', async (ev) => {
    if (ev.target && ev.target.closest('[data-dtool-recorder-toolbar]')) return
    const cfg = {}
    if (mode === 'click') {
      cfg.selector = buildSelectorChain(ev.target)
      cfg.selector_type = 'css'
    } else if (mode === 'click_xy') {
      const c = viewportRelativeCoords(ev)
      cfg.x = c.x; cfg.y = c.y; cfg.viewport_width = c.w; cfg.viewport_height = c.h
    } else {
      return
    }
    try {
      await transport.addStep({
        type: mode === 'click' ? 'click_v1' : 'click_by_position_v1',
        version: '1.0',
        description: `${mode} ${cfg.selector || `${cfg.x},${cfg.y}`}`,
        config: cfg,
        wait_after_ms: 200,
        recorded_at: Date.now(),
      })
      steps += 1
      update()
    } catch (e) {
      console.warn('[recorder] add step failed', e)
    }
  }, true)

  document.addEventListener('input', async (ev) => {
    if (mode !== 'input') return
    if (ev.target && ev.target.closest('[data-dtool-recorder-toolbar]')) return
    const cfg = {
      selector: buildSelectorChain(ev.target),
      selector_type: 'css',
      value: ev.target.value,
      clear_before: true,
    }
    try {
      await transport.addStep({
        type: 'input_v1',
        version: '1.0',
        description: `input ${cfg.selector}`,
        config: cfg,
        wait_after_ms: 200,
        recorded_at: Date.now(),
      })
      steps += 1
      update()
    } catch (e) { console.warn(e) }
  }, true)

  toolbar.querySelector('[data-commit]').addEventListener('click', async (ev) => {
    ev.stopPropagation()
    const gid = Number(prompt('提交到 e2e group_id（数字）') || 0)
    if (gid <= 0) return
    try {
      await transport.commit({
        group_id: gid,
        name: `录制 ${new Date().toLocaleString()}`,
        tags: '',
      })
      alert('已提交')
      toolbar.remove()
    } catch (e) {
      alert('提交失败：' + e.message)
    }
  })

  toolbar.querySelector('[data-close]').addEventListener('click', async (ev) => {
    ev.stopPropagation()
    toolbar.remove()
  })
}

(function () {
  const cfg = window.__dtoolRecorder
  if (!cfg) return
  if (document.readyState === 'complete') bootRecorder(cfg)
  else window.addEventListener('load', () => bootRecorder(cfg))
})()