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
    HOME_DASHBOARD_PAGE_SWITCH_HOT_ZONE_WIDTH,
    findBlockingScrollableAncestor,
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

  const dashboardMessageList = createScrollableElement({
    clientHeight: 720,
    scrollHeight: 1320,
    scrollTop: 200,
    parentElement: rootElement,
  })
  const processPanel = createScrollableElement({
    clientHeight: 240,
    scrollHeight: 840,
    scrollTop: 180,
    parentElement: dashboardMessageList,
  })
  const processMarkdownLine = {
    parentElement: processPanel,
  }
  const ordinaryMessageLine = {
    parentElement: dashboardMessageList,
  }

  assert.strictEqual(
    findBlockingScrollableAncestor(processMarkdownLine, 48, rootElement),
    processPanel,
    '执行过程输出框存在更内层滚动时，应优先识别该滚动容器'
  )

  assert.strictEqual(
    findBlockingScrollableAncestor(ordinaryMessageLine, 48, rootElement),
    dashboardMessageList,
    '普通消息区域没有更内层滚动时，应回退到首页消息列表滚动容器'
  )

  assert.strictEqual(
    isHomeDashboardPageSwitchHotZone(980, { left: 0, right: 1000 }),
    true,
    '鼠标位于首页最右侧 200px 热区时，应允许直接触发翻页'
  )

  assert.strictEqual(
    HOME_DASHBOARD_PAGE_SWITCH_HOT_ZONE_WIDTH,
    300,
    '首页强制翻页热区应保持在最右侧 300px，命中后可直接切屏'
  )

  assert.strictEqual(
    isHomeDashboardPageSwitchHotZone(760, { left: 0, right: 1000 }),
    true,
    '鼠标位于首页最右侧 300px 热区内时，应命中强制翻页热区'
  )

  console.log('home_dashboard_wheel tests passed')
}

run()
