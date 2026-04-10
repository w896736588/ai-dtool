// WHEEL_SCROLL_DIRECTION_* 用于标识滚轮方向，避免在判断中直接使用魔法值。
const WHEEL_SCROLL_DIRECTION_UP = -1
const WHEEL_SCROLL_DIRECTION_DOWN = 1
// SCROLL_POSITION_MIN 表示滚动容器顶部位置。
const SCROLL_POSITION_MIN = 0
// HOME_DASHBOARD_PAGE_SWITCH_HOT_ZONE_WIDTH 表示首页最右侧允许强制翻页的热区宽度。
const HOME_DASHBOARD_PAGE_SWITCH_HOT_ZONE_WIDTH = 300

// getParentElement 兼容 Element 与普通节点，统一向上查找父元素。
function getParentElement(node) {
  if (!node || typeof node !== 'object') {
    return null
  }
  if (node.parentElement) {
    return node.parentElement
  }
  if (node.parentNode && typeof node.parentNode === 'object') {
    return node.parentNode
  }
  return null
}

// normalizeWheelDirection 将滚轮位移转换成统一方向，便于后续滚动判断复用。
function normalizeWheelDirection(deltaY) {
  const numericDeltaY = Number(deltaY || 0)
  if (numericDeltaY > SCROLL_POSITION_MIN) {
    return WHEEL_SCROLL_DIRECTION_DOWN
  }
  if (numericDeltaY < SCROLL_POSITION_MIN) {
    return WHEEL_SCROLL_DIRECTION_UP
  }
  return SCROLL_POSITION_MIN
}

// isScrollableElement 仅识别真正有滚动空间的容器，避免普通布局节点误拦截翻页。
function isScrollableElement(element) {
  if (!element || typeof element !== 'object') {
    return false
  }
  const clientHeight = Number(element.clientHeight || 0)
  const scrollHeight = Number(element.scrollHeight || 0)
  return clientHeight > SCROLL_POSITION_MIN && scrollHeight > clientHeight
}

// findScrollableAncestor 用于定位事件源对应的最近可滚动祖先容器。
function findScrollableAncestor(target, stopElement = null) {
  let currentElement = target
  while (currentElement && currentElement !== stopElement) {
    if (isScrollableElement(currentElement)) {
      return currentElement
    }
    currentElement = getParentElement(currentElement)
  }
  return null
}

// canElementScrollInDirection 判断容器在当前滚轮方向上是否还能继续内部滚动。
function canElementScrollInDirection(element, direction) {
  if (!isScrollableElement(element)) {
    return false
  }
  const scrollTop = Number(element.scrollTop || 0)
  const maxScrollTop = Math.max(Number(element.scrollHeight || 0) - Number(element.clientHeight || 0), SCROLL_POSITION_MIN)
  if (direction === WHEEL_SCROLL_DIRECTION_DOWN) {
    return scrollTop < maxScrollTop
  }
  if (direction === WHEEL_SCROLL_DIRECTION_UP) {
    return scrollTop > SCROLL_POSITION_MIN
  }
  return false
}

// findBlockingScrollableAncestor 返回当前滚轮方向上真正会拦截整屏切换的最近滚动容器。
function findBlockingScrollableAncestor(target, deltaY, stopElement = null) {
  const direction = normalizeWheelDirection(deltaY)
  if (direction === SCROLL_POSITION_MIN) {
    return null
  }
  let currentElement = target
  while (currentElement && currentElement !== stopElement) {
    if (canElementScrollInDirection(currentElement, direction)) {
      return currentElement
    }
    currentElement = getParentElement(currentElement)
  }
  return null
}

// resolveHomeDashboardPageSwitchBlocker 兼容首页命名，返回当前滚轮命中的实际拦截容器。
function resolveHomeDashboardPageSwitchBlocker(target, deltaY, stopElement = null) {
  return findBlockingScrollableAncestor(target, deltaY, stopElement)
}

// isHomeDashboardPageSwitchHotZone 判断鼠标是否位于首页最右侧热区内。
function isHomeDashboardPageSwitchHotZone(clientX, containerRect, hotZoneWidth = HOME_DASHBOARD_PAGE_SWITCH_HOT_ZONE_WIDTH) {
  const numericClientX = Number(clientX || 0)
  if (!containerRect || typeof containerRect !== 'object') {
    return false
  }
  const containerLeft = Number(containerRect.left || 0)
  const containerRight = Number(containerRect.right || 0)
  const numericHotZoneWidth = Math.max(Number(hotZoneWidth || 0), SCROLL_POSITION_MIN)
  if (containerRight <= containerLeft || numericHotZoneWidth <= SCROLL_POSITION_MIN) {
    return false
  }
  return numericClientX >= containerRight - numericHotZoneWidth && numericClientX <= containerRight
}

// shouldBlockHomeDashboardPageSwitch 只在命中可继续滚动的内部容器时阻止首页整屏切换。
function shouldBlockHomeDashboardPageSwitch(target, deltaY, stopElement = null) {
  return !!resolveHomeDashboardPageSwitchBlocker(target, deltaY, stopElement)
}

module.exports = {
  HOME_DASHBOARD_PAGE_SWITCH_HOT_ZONE_WIDTH,
  WHEEL_SCROLL_DIRECTION_UP,
  WHEEL_SCROLL_DIRECTION_DOWN,
  normalizeWheelDirection,
  isScrollableElement,
  findScrollableAncestor,
  canElementScrollInDirection,
  findBlockingScrollableAncestor,
  resolveHomeDashboardPageSwitchBlocker,
  isHomeDashboardPageSwitchHotZone,
  shouldBlockHomeDashboardPageSwitch,
}
