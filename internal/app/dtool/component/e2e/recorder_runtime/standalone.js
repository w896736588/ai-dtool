// Dtool 录制 runtime —— 单一脚本，被 dtool server 通过 page.AddInitScript 注册。
//
// 设计要点：
//   1. AddInitScript 在「每次 navigation / 新 document」浏览器执行（在 page scripts 之前），
//      所以任何 `<a href>` 点击 / `location.href = xxx` 都会触发本脚本重新跑一次。
//   2. 选举：用 window.__dtoolRecorderHost 判断，true 表示本 window 已选举为 host。
//      只有 host 才挂 toolbar / 维护 buffer；slave 监听本 window 事件，通过 BroadcastChannel 报给 host。
//   3. buffer 持久化到 localStorage 'dtool-record-buffer'，跨 navigation 不丢；
//      slave 上报的 step 被 host 接收后写入 buffer + 同步刷新 localStorage。
//   4. 同 origin 同 context 的所有 page 都加入同一个 BroadcastChannel 'dtool-recorder'，
//      host 切换或 buffer 状态变化通过 channel 广播。
//   5. SPA 路由（pushState/replaceState）属于同 document，不触发 AddInitScript。
//      这类路由 host 已挂在 document.body 上，不需要重建。

;(function () {
  var HOST_KEY = '__dtoolRecorderHost'
  var BUF_KEY = 'dtool-record-buffer'
  var MODE_KEY = 'dtool-record-mode'
  var CHANNEL_NAME = 'dtool-recorder'

  // 工具：safe ls 读写（可能 SSR / 隐私模式禁用）
  function lsRead(key, fallback) {
    try { var v = localStorage.getItem(key); return v == null ? fallback : v } catch (e) { return fallback }
  }
  function lsWrite(key, v) { try { localStorage.setItem(key, v) } catch (e) {} }

  function nowISO() { return new Date().toISOString() }
  function escapeHtml(s) {
    return String(s).replace(/[&<>"']/g, function (c) {
      return { '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]
    })
  }
  function randomStepId() {
    return 'stp_' + Date.now() + '_' + Math.floor(Math.random() * 100000)
  }

  // ---- 共享业务逻辑：selector / viewport ----
  function buildSelectorChain(el) {
    if (!el || el.nodeType !== 1) return ''
    var parts = []
    var cur = el
    while (cur && cur.nodeType === 1 && cur !== document.documentElement) {
      var part = cur.tagName ? cur.tagName.toLowerCase() : '*'
      if (cur.id) { part += '#' + cur.id; parts.unshift(part); break }
      if (cur.className && typeof cur.className === 'string') {
        var cls = cur.className.trim().split(/\s+/).slice(0, 2).join('.')
        if (cls) part += '.' + cls
      }
      var name = cur.getAttribute && cur.getAttribute('name')
      if (name) part += '[name="' + name + '"]'
      parts.unshift(part)
      cur = cur.parentNode
    }
    return parts.join(' > ')
  }
  function viewportCoords(ev) {
    return {
      x: ev.clientX, y: ev.clientY,
      viewport_width: window.innerWidth, viewport_height: window.innerHeight,
    }
  }
  function isInToolbar(el) {
    if (!el || !el.closest) return false
    return !!el.closest('[data-dtool-recorder]')
  }

  // ---- buffer 持久化 ----
  // 返回 steps 数组（始终是 fresh copy + 补上 id）
  function loadBuffer() {
    var raw = lsRead(BUF_KEY, '[]')
    try {
      var arr = JSON.parse(raw)
      return Array.isArray(arr) ? arr.filter(function (s) { return s && typeof s === 'object' && s.type }) : []
    } catch (e) { return [] }
  }
  function saveBuffer(arr) {
    lsWrite(BUF_KEY, JSON.stringify(arr.slice(-500)))
  }
  function loadMode() {
    return lsRead(MODE_KEY, 'click') || 'click'
  }
  function saveMode(m) { lsWrite(MODE_KEY, m || 'click') }

  // ============== SLAVE 分支 ==============
  // 非 host 的 window（也即 navigation 后产生的新 document / 同 context 其它 page）：
  // - 不挂 toolbar
  // - 监听本 window 的 click / input，按当前 mode 记录 step 后通过 BroadcastChannel 上报给 host
  function runSlave() {
    var mode = loadMode()
    var bc
    try { bc = new BroadcastChannel(CHANNEL_NAME) } catch (e) { bc = null }

    function report(step) {
      step.id = randomStepId()
      step.recorded_at = nowISO()
      if (bc) { try { bc.postMessage({ type: 'dtool:step', step: step }) } catch (e) {} }
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
      report({
        type: mode === 'click' ? 'click_v1' : 'click_by_position_v1',
        version: '1.0',
        description: mode === 'click'
          ? 'click ' + cfg.selector
          : 'click_xy ' + cfg.x + ',' + cfg.y,
        config: cfg,
        wait_after_ms: 200,
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
      report({
        type: 'input_v1',
        version: '1.0',
        description: 'input ' + cfg.selector + ' = ' + cfg.value,
        config: cfg,
        wait_after_ms: 200,
      })
    }, true)

    window.addEventListener('scroll', function () {
      if (mode !== 'scroll') return
      var last = loadBuffer().slice(-1)[0]
      if (last && last.type === 'scroll_v1' && (Date.now() - (last._ts || 0)) < 200) return
      report({
        type: 'scroll_v1',
        version: '1.0',
        description: 'scroll',
        config: { delta_y: window.scrollY, scroll_x: window.scrollX },
        wait_after_ms: 0,
        _ts: Date.now(),
      })
    }, true)

    // 接收 host 的 mode 切换广播
    if (bc) {
      bc.onmessage = function (ev) {
        if (ev && ev.data && ev.data.type === 'dtool:mode') {
          mode = ev.data.mode || 'click'
          saveMode(mode)
        }
      }
    }
  }

  // ============== HOST 分支 ==============
  function runHost() {
    window[HOST_KEY] = true

    // 广播 channel
    var bc
    try { bc = new BroadcastChannel(CHANNEL_NAME) } catch (e) { bc = null }

    // 共享 buffer + mode：直接用 localStorage 持久化（host 自己读写）
    var buffer = loadBuffer()
    var mode = loadMode()

    function persist() { saveBuffer(buffer) }
    function setMode(m) {
      mode = m
      saveMode(m)
      if (bc) { try { bc.postMessage({ type: 'dtool:mode', mode: m }) } catch (e) {} }
    }

    // ---- style ----
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
      if (!box) return
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
      var stat = toolbar.querySelector('[data-stat]')
      if (stat) stat.textContent = buffer.length + ' 步 · ' + mode
      toolbar.querySelectorAll('.modes button').forEach(function (b) {
        if (b.dataset.mode === mode) b.classList.add('active')
        else b.classList.remove('active')
      })
    }

    function bindToolbar(toolbar) {
      toolbar.querySelectorAll('.modes button').forEach(function (b) {
        b.addEventListener('click', function (ev) {
          ev.stopPropagation()
          setMode(b.dataset.mode)
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
        window.__dtoolRecordResult = buildResult()
        window.__dtoolRecordDone = true
        toast('已下载并复制 JSON')
      })

      toolbar.querySelector('[data-cancel]').addEventListener('click', function (ev) {
        ev.stopPropagation()
        buffer = []
        persist()
        window.__dtoolRecordResult = null
        window.__dtoolRecordDone = false
        updateStat(toolbar)
        renderBuffer(toolbar)
        toast('已清空')
      })
    }

    function pushStep(step) {
      if (!step || !step.type) return
      buffer.push(step)
      persist()
      var tb = ensureToolbar()
      updateStat(tb)
      renderBuffer(tb)
    }

    function buildResult() {
      return {
        schema: 'dtool.record.v1',
        recorded_at: nowISO(),
        url: window.__dtoolRecorderOriginURL || location.href,
        title: window.__dtoolRecorderOriginTitle || document.title,
        step_count: buffer.length,
        steps: buffer.slice(),
      }
    }

    function copyToClipboard(text) {
      if (navigator.clipboard && navigator.clipboard.writeText) {
        return navigator.clipboard.writeText(text).catch(function () { return fallbackCopy(text) })
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

    // ---- host 自身的事件监听（同 document 触发的动作） ----
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
        id: randomStepId(),
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
      pushStep({
        id: randomStepId(),
        type: 'input_v1',
        version: '1.0',
        description: 'input ' + buildSelectorChain(ev.target) + ' = ' + ev.target.value,
        config: {
          selector: buildSelectorChain(ev.target),
          selector_type: 'css',
          value: ev.target.value,
          clear_before: true,
        },
        wait_after_ms: 200,
        recorded_at: nowISO(),
      })
    }, true)

    window.addEventListener('scroll', function () {
      if (mode !== 'scroll') return
      var last = buffer[buffer.length - 1]
      if (last && last.type === 'scroll_v1' && (Date.now() - (last._ts || 0)) < 200) return
      pushStep({
        id: randomStepId(),
        type: 'scroll_v1',
        version: '1.0',
        description: 'scroll',
        config: { delta_y: window.scrollY, scroll_x: window.scrollX },
        wait_after_ms: 0,
        recorded_at: nowISO(),
        _ts: Date.now(),
      })
    }, true)

    // 接收 slave 上报的 step
    if (bc) {
      bc.onmessage = function (ev) {
        if (!ev || !ev.data) return
        if (ev.data.type === 'dtool:step') {
          pushStep(ev.data.step)
        }
      }
    }

    // 暴露给 Playwright / 前端
    window.__dtoolRecordBuffer = buffer
    window.__dtoolRecordFlush = function () {
      window.__dtoolRecordResult = buildResult()
      window.__dtoolRecordDone = true
      return window.__dtoolRecordResult
    }

    function boot() {
      ensureStyles()
      var tb = ensureToolbar()
      updateStat(tb)
      renderBuffer(tb)
    }

    if (document.readyState === 'loading') {
      document.addEventListener('DOMContentLoaded', boot)
    } else {
      boot()
    }
  }

  // ---- 入口：根据 HOST_KEY 选举 ----
  if (window[HOST_KEY]) {
    runSlave()
  } else {
    // 注意：同一浏览器同 context 中的所有 page 会同时执行该 IIFE。
    // 这里不能依赖"宿主是第一个跑的 document"——只有当同 context 多 tab 同时打开时，
    // 第一个完成的 window 才是 host；BroadcastChannel 用来在切换时转交 host 角色。
    // 简化起见：因为 E2E 录制只会开一个 page，这里"先标 host 后跑"足够。
    try {
      var w = window.top || window
      if (w[HOST_KEY]) {
        runSlave()
      } else {
        runHost()
        // 同时把 top 上也标记一遍，避免 iframe 套娃里混乱
        try { w[HOST_KEY] = true } catch (e) {}
      }
    } catch (e) {
      runHost()
    }
  }
})()
