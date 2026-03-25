const assert = require('assert')

const MODULE_PATH = '../src/utils/home_dashboard_wheel.cjs'

const loadWheelModule = () => require(MODULE_PATH)

const createScrollableElement = ({
  clientHeight,
  scrollHeight,
  scrollTop,
  parentElement = null,
}) => ({
  clientHeight,
  scrollHeight,
  scrollTop,
  parentElement,
})

const run = () => {
  const {
    shouldBlockHomeDashboardPageSwitch,
    isHomeDashboardPageSwitchHotZone,
  } = loadWheelModule()

  const rootElement = createScrollableElement({
    clientHeight: 480,
    scrollHeight: 480,
    scrollTop: 0,
  })
  const processContainer = createScrollableElement({
    clientHeight: 200,
    scrollHeight: 600,
    scrollTop: 120,
    parentElement: rootElement,
  })
  const processTextLine = {
    parentElement: processContainer,
  }

  assert.strictEqual(
    shouldBlockHomeDashboardPageSwitch(processTextLine, 48),
    true,
    '命令执行过程输出框还能继续向下滚动时，不应触发首页翻页'
  )

  processContainer.scrollTop = 80
  assert.strictEqual(
    shouldBlockHomeDashboardPageSwitch(processTextLine, -48),
    true,
    '命令执行过程输出框还能继续向上滚动时，不应触发首页翻页'
  )

  processContainer.scrollTop = 0
  assert.strictEqual(
    shouldBlockHomeDashboardPageSwitch(processTextLine, -48),
    false,
    '命令执行过程输出框滚到顶部后，应允许继续向上切换首页页面'
  )

  processContainer.scrollTop = 400
  assert.strictEqual(
    shouldBlockHomeDashboardPageSwitch(processTextLine, 48),
    false,
    '命令执行过程输出框滚到底部后，应允许继续向下切换首页页面'
  )

  const staticContainer = createScrollableElement({
    clientHeight: 240,
    scrollHeight: 240,
    scrollTop: 0,
    parentElement: rootElement,
  })
  const staticChild = {
    parentElement: staticContainer,
  }
  assert.strictEqual(
    shouldBlockHomeDashboardPageSwitch(staticChild, 48, rootElement),
    false,
    '非可滚动区域的滚轮事件应该继续交给首页翻页逻辑处理'
  )

  assert.strictEqual(
    isHomeDashboardPageSwitchHotZone(980, { left: 0, right: 1000 }),
    true,
    '鼠标位于首页最右侧 200px 热区时，应允许直接触发翻页'
  )

  assert.strictEqual(
    isHomeDashboardPageSwitchHotZone(760, { left: 0, right: 1000 }),
    false,
    '鼠标离开首页最右侧 200px 热区后，不应命中强制翻页热区'
  )

  console.log('home_dashboard_wheel tests passed')
}

run()
