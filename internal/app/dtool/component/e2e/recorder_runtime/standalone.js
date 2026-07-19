// Dtool 录制 runtime —— 单文件、自包含、无 ES module、无 fetch。
// 由 Go 端通过 page.Evaluate 直接注入到被测 page。
//
// 行为：
//   1. 在被测 page 顶部注入 toolbar DOM（不挂 iframe）。
//   2. 监听 click / input / scroll / change 事件，把动作 push 到 window.__dtoolRecordBuffer。
//   3. toolbar 上的「结束录制」按钮：把 buffer 序列化为 JSON，复制到剪贴板 + 触发下载 + 暴露
//      window.__dtoolRecordResult 给 Playwright 读取。
//   4. 「放弃」按钮：清空 buffer。
//
// 与 dtool 后端的通信完全通过 Playwright（页面侧的 window 状态），
// 不再依赖同源 iframe / proxy.html / fetch ws_token。

(function () {
  if (window.__dtoolRecorderInjected) return
  window.__dtoolRecorderInjected = true

  // ---- 工具：选择器 ----
  function buildSelectorChain(el) {
    if (!el || el.nodeType !== 1) return ''
    var parts = []
    var cur = el
    while (cur && cur.nodeType === 1 && cur !== document.documentElement) {
      var part = cur.tagName ? cur.tagName.toLowerCase() : '*'
      if (cur.id) {
        part += '#' + cur.id
        parts.unshift(part)
        break
      }
      if (cur.dataset && cur.dataset.testid) {
        part += '[data-testid="' + cur.dataset.testid + '"]'
        parts.unshift(part)
        break
      }
      var cls = (cur.getAttribute && cur.getAttribute('class') || '').trim()
      if (cls) {
        var firstTwo = cls.split(/\s+/).slice(0, 2).join('.')
        if (firstTwo) part += '.' + firstTwo
      }
      parts.unshift(part)
      cur = cur.parentElement
    }
    return parts.join(' > ')
  }

  function viewportCoords(ev) {
    return {
      x: ev.clientX,
      y: ev.clientY,
      viewport_width: window.innerWidth,
      viewport_height: window.innerHeight,
    }
  }

  function nowISO() {
    return new Date().toISOString()
  }

  // ---- 状态 ----
  var buffer = []
  var mode = 'click' // click | click_xy | input | scroll | assert

  // ---- toolbar ----
  var TOOLBAR_CSS = [
    '[data-dtool-recorder]{position:fixed;top:80px;right:20px;z-index:2147483647;',
    'background:#fff;border-radius:8px;box-shadow:0 4px 16px rgba(0,0,0,.18);',
    'width:360px;font:12px/1.4 -apple-system,BlinkMacSystemFont,"Segoe UI",sans-serif;color:#303133;',
    'user-select:none;}',
    '[data-dtool-recorder] header{padding:8px 10px;color:#fff;',
    'background:linear-gradient(90deg,#409eff,#66b1ff);border-radius:8px 8px 0 0;',
    'font-weight:600;display:flex;justify-content:space-between;align-items:center;}',
    '[data-dtool-recorder] header .stat{font-weight:400;font-size:11px;opacity:.95;}',
    '[data-dtool-recorder] .modes{padding:8px 10px;display:flex;gap:6px;flex-wrap:wrap;border-bottom:1px solid #eee;}',
    '[data-dtool-recorder] .modes button{background:#ecf5ff;color:#409eff;border:1px solid #b3d8ff;',
    'border-radius:4px;padding:6px 8px;cursor:pointer;font-size:12px;}',
    '[data-dtool-recorder] .modes button.active{background:#409eff;color:#fff;border-color:#409eff;}',
    '[data-dtool-recorder] .buffer{padding:8px 10px;max-height:200px;overflow:auto;font:11px/1.5 ui-monospace,monospace;',
    'background:#fafafa;border-bottom:1px solid #eee;}',
    '[data-dtool-recorder] .buffer .empty{color:#909399;font-style:italic;}',
    '[data-dtool-recorder] .buffer .item{padding:2px 0;border-bottom:1px dashed #eee;}',
    '[data-dtool-recorder] .buffer .item:last-child{border-bottom:0;}',
    '[data-dtool-recorder] .buffer .item .type{color:#409eff;font-weight:600;}',
    '[data-dtool-recorder] .actions{padding:8px 10px;display:flex;gap:6px;}',
    '[data-dtool-recorder] .actions button{flex:1;border:0;border-radius:4px;padding:8px;cursor:pointer;color:#fff;font-size:12px;}',
    '[data-dtool-recorder] .actions .commit{background:#67c23a;}',
    '[data-dtool-recorder] .actions .cancel{background:#f56c6c;}',
    '[data-dtool-recorder] .actions .copy{background:#909399;}',
    '[data-dtool-recorder] .toast{position:fixed;top:50%;left:50%;transform:translate(-50%,-50%);',
    'background:rgba(0,0,0,.85);color:#fff;padding:12px 20px;border-radius:6px;z-index:2147483647;',
    'font:13px/1.5 -apple-system,sans-serif;}',
  ].join('')

  function ensureStyles() {
    if (document.getElementById('__dtool_recorder_style')) return
    var s = document.createElement('style')
    s.id = '__dtool_recorder_style'
    s.textContent = TOOLBAR_CSS
    ;(document.head || document.documentElement).appendChild(s)
  }

  function toast(msg, durationMs) {
    durationMs = durationMs || 1500
    var t = document.createElement('div')
    t.className = 'toast'
    t.textContent = msg
    document.body.appendChild(t)
    setTimeout(function () { t.remove() }, durationMs)
  }

  function renderBuffer(toolbar) {
    var box = toolbar.querySelector('.buffer')
    if (!buffer.length) {
      box.innerHTML = '<div class="empty">暂无步骤</div>'
      return
    }
    box.innerHTML = buffer.map(function (s, i) {
      var desc = ''
      if (s.type === 'click_v1') desc = s.config && s.config.selector || ''
      else if (s.type === 'click_by_position_v1') desc = 'x=' + (s.config && s.config.x) + ',y=' + (s.config && s.config.y)
      else if (s.type === 'input_v1') desc = (s.config && s.config.selector || '') + ' = ' + (s.config && s.config.value)
      else if (s.type === 'scroll_v1') desc = 'scroll ' + (s.config && s.config.delta_y)
      else desc = JSON.stringify(s.config || {})
      return '<div class="item"><span class="type">' + (i + 1) + '. ' + s.type + '</span> ' + escapeHtml(desc) + '</div>'
    }).join('')
    box.scrollTop = box.scrollHeight
  }

  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]
    })
  }

  function ensureToolbar() {
    var existing = document.querySelector('[data-dtool-recorder]')
    if (existing) return existing
    var wrap = document.createElement('div')
    wrap.setAttribute('data-dtool-recorder', '1')
    wrap.innerHTML = [
      '<header><span>Dtool 录制</span><span class="stat" data-stat>0 步 · click</span></header>',
      '<div class="modes">',
      '<button data-mode="click">元素点击</button>',
      '<button data-mode="click_xy">坐标点击</button>',
      '<button data-mode="input">输入</button>',
      '<button data-mode="scroll">滚动</button>',
      '<button data-mode="assert">断言（占位）</button>',
      '</div>',
      '<div class="buffer" data-buffer><div class="empty">暂无步骤</div></div>',
      '<div class="actions">',
      '<button class="copy" data-copy>复制 JSON</button>',
      '<button class="commit" data-commit>结束并下载</button>',
      '<button class="cancel" data-cancel>放弃</button>',
      '</div>',
    ].join('')
    document.body.appendChild(wrap)
    bindToolbar(wrap)
    return wrap
  }

  function updateStat(toolbar) {
    toolbar.querySelector('[data-stat]').textContent = buffer.length + ' 步 · ' + mode
    toolbar.querySelectorAll('.modes button').forEach(function (b) {
      if (b.dataset.mode === mode) b.classList.add('active')
      else b.classList.remove('active')
    })
  }

  function buildResult() {
    return {
      schema: 'dtool.record.v1',
      recorded_at: nowISO(),
      url: location.href,
      title: document.title,
      step_count: buffer.length,
      steps: buffer.slice(),
    }
  }

  function copyToClipboard(text) {
    if (navigator.clipboard && navigator.clipboard.writeText) {
      return navigator.clipboard.writeText(text).catch(function () {
        return fallbackCopy(text)
      })
    }
    return Promise.resolve(fallbackCopy(text))
  }

  function fallbackCopy(text) {
    var ta = document.createElement('textarea')
    ta.value = text
    ta.style.position = 'fixed'
    ta.style.opacity = '0'
    document.body.appendChild(ta)
    ta.select()
    try { document.execCommand('copy') } catch (e) {}
    ta.remove()
  }

  function downloadJson(filename, text) {
    var blob = new Blob([text], { type: 'application/json' })
    var a = document.createElement('a')
    a.href = URL.createObjectURL(blob)
    a.download = filename
    document.body.appendChild(a)
    a.click()
    setTimeout(function () { URL.revokeObjectURL(a.href); a.remove() }, 0)
  }

  function bindToolbar(toolbar) {
    toolbar.querySelectorAll('.modes button').forEach(function (b) {
      b.addEventListener('click', function (ev) {
        ev.stopPropagation()
        mode = b.dataset.mode
        updateStat(toolbar)
      })
    })

    toolbar.querySelector('[data-copy]').addEventListener('click', function (ev) {
      ev.stopPropagation()
      var text = JSON.stringify(buildResult(), null, 2)
      copyToClipboard(text)
      toast('已复制 JSON（' + buffer.length + ' 步）')
    })

    toolbar.querySelector('[data-commit]').addEventListener('click', function (ev) {
      ev.stopPropagation()
      if (!buffer.length) { toast('没有可导出的步骤'); return }
      var text = JSON.stringify(buildResult(), null, 2)
      var filename = 'dtool-record-' + Date.now() + '.json'
      downloadJson(filename, text)
      copyToClipboard(text)
      // 把最终结果挂到 window，Playwright / 前端可读取
      window.__dtoolRecordResult = buildResult()
      window.__dtoolRecordDone = true
      toast('已下载并复制 JSON')
    })

    toolbar.querySelector('[data-cancel]').addEventListener('click', function (ev) {
      ev.stopPropagation()
      buffer = []
      window.__dtoolRecordResult = null
      window.__dtoolRecordDone = false
      updateStat(toolbar)
      renderBuffer(toolbar)
      toast('已清空')
    })
  }

  function pushStep(step) {
    buffer.push(step)
    var toolbar = ensureToolbar()
    updateStat(toolbar)
    renderBuffer(toolbar)
  }

  // ---- 事件监听（挂在 window 上而不是 document，SPA 路由跳转后新 document 替换时不会丢监听） ----
  function isInToolbar(el) {
    if (!el || !el.closest) return false
    return !!el.closest('[data-dtool-recorder]')
  }

  window.addEventListener('click', function (ev) {
    if (isInToolbar(ev.target)) return
    if (mode !== 'click' && mode !== 'click_xy') return
    var cfg = {}
    if (mode === 'click') {
      cfg.selector = buildSelectorChain(ev.target)
      cfg.selector_type = 'css'
    } else {
      var c = viewportCoords(ev)
      cfg.x = c.x; cfg.y = c.y
      cfg.viewport_width = c.viewport_width
      cfg.viewport_height = c.viewport_height
    }
    pushStep({
      type: mode === 'click' ? 'click_v1' : 'click_by_position_v1',
      version: '1.0',
      description: mode === 'click'
        ? 'click ' + cfg.selector
        : 'click_xy ' + cfg.x + ',' + cfg.y,
      config: cfg,
      wait_after_ms: 200,
      recorded_at: nowISO(),
    })
  }, true)

  window.addEventListener('input', function (ev) {
    if (isInToolbar(ev.target)) return
    if (mode !== 'input') return
    var cfg = {
      selector: buildSelectorChain(ev.target),
      selector_type: 'css',
      value: ev.target.value,
      clear_before: true,
    }
    pushStep({
      type: 'input_v1',
      version: '1.0',
      description: 'input ' + cfg.selector + ' = ' + cfg.value,
      config: cfg,
      wait_after_ms: 200,
      recorded_at: nowISO(),
    })
  }, true)

  window.addEventListener('scroll', function () {
    if (mode !== 'scroll') return
    // 去重：上一次 scroll 距今 < 200ms 就跳过
    var last = buffer[buffer.length - 1]
    if (last && last.type === 'scroll_v1' && (Date.now() - last._ts) < 200) return
    pushStep({
      type: 'scroll_v1',
      version: '1.0',
      description: 'scroll',
      config: { delta_y: window.scrollY, scroll_x: window.scrollX },
      wait_after_ms: 0,
      recorded_at: nowISO(),
      _ts: Date.now(),
    })
  }, true)

  // ---- SPA 路由感知 ----
  // pushState / replaceState 不触发 popstate，需要 patch 它们发自定义事件。
  // 路由变化后 SPA 通常替换 body，重新调用 ensureToolbar() 重建 toolbar DOM。
  try {
    var origPushState = history.pushState
    var origReplaceState = history.replaceState
    history.pushState = function () {
      var ret = origPushState.apply(this, arguments)
      window.dispatchEvent(new Event('dtool:locationchange'))
      return ret
    }
    history.replaceState = function () {
      var ret = origReplaceState.apply(this, arguments)
      window.dispatchEvent(new Event('dtool:locationchange'))
      return ret
    }
  } catch (e) {}

  function onLocationChange() {
    // 让浏览器先完成 DOM 替换（pushState 同步，但 vdom diff 是微任务）
    setTimeout(function () {
      try {
        ensureStyles()
        var tb = ensureToolbar()
        // 重建时 sync 当前 mode / buffer 计数
        updateStat(tb)
        renderBuffer(tb)
      } catch (e) {
        // 重建失败不影响录制，只是 toolbar 临时不见
      }
    }, 50)
  }

  window.addEventListener('popstate', onLocationChange)
  window.addEventListener('hashchange', onLocationChange)
  window.addEventListener('dtool:locationchange', onLocationChange)

  // ---- 暴露给 Playwright / 前端 ----
  window.__dtoolRecordBuffer = buffer
  window.__dtoolRecordFlush = function () {
    window.__dtoolRecordResult = buildResult()
    window.__dtoolRecordDone = true
    return window.__dtoolRecordResult
  }

  // 启动：先确保 toolbar 出现
  if (document.readyState === 'loading') {
    document.addEventListener('DOMContentLoaded', function () {
      ensureStyles()
      ensureToolbar()
    })
  } else {
    ensureStyles()
    ensureToolbar()
  }
})()