export function buildSelectorChain(el) {
  const parts = []
  let cur = el
  while (cur && cur !== document.documentElement) {
    let part = cur.tagName.toLowerCase()
    if (cur.id) {
      part += `#${cur.id}`
      parts.unshift(part)
      break
    }
    if (cur.dataset && cur.dataset.testid) {
      part += `[data-testid="${cur.dataset.testid}"]`
      parts.unshift(part)
      break
    }
    const cls = (cur.getAttribute('class') || '').trim().split(/\s+/).slice(0, 2).join('.')
    if (cls) part += '.' + cls
    parts.unshift(part)
    cur = cur.parentElement
  }
  return parts.join(' > ')
}

export function viewportRelativeCoords(ev) {
  return {
    x: ev.clientX,
    y: ev.clientY,
    w: window.innerWidth,
    h: window.innerHeight,
  }
}