<template>
  <el-scrollbar id="showShellResult" :style="{ height: divHeight - 17 + 'px' }">
    <div
        class="sticky-textarea-div"
        v-html="shellShowResult"
        :style="{ minHeight: divHeight - 25 + 'px' }"
    ></div>
  </el-scrollbar>
</template>

<script setup>
/* ---------- 依赖 ---------- */
import { nextTick, watch, onMounted, onBeforeUnmount } from 'vue'

/* ----------  props  ---------- */
const props = defineProps({
  shellShowResult: { type: String, default: '' },
  divHeight: { type: Number, default: 200 }
})

/* ---------- 自动滚动逻辑 ---------- */
const scrollThreshold = 10          // 离底部 ≤ 10 px 认为“已到底”
let autoScroll = true               // 默认自动滚
let wrapEl = null                   // 真正的滚动容器 .el-scrollbar__wrap
let rafLock = false                 // 防止滚动事件高频触发

/* 获取真实滚动容器 */
function getWrap () {
  const sb = document.getElementById('showShellResult')
  return sb?.parentNode
}

/* 滚到底 */
function scrollToBottom () {
  if (!autoScroll || !wrapEl) return
  wrapEl.scrollTop = wrapEl.scrollHeight
}

/* 滚动事件：判断是否需要切换自动滚动状态 */
function onScroll () {
  if (rafLock) return
  rafLock = true
  window.requestAnimationFrame(() => {
    const distance = wrapEl.scrollHeight - wrapEl.scrollTop - wrapEl.clientHeight
    // 到底就恢复自动滚，否则暂停
    autoScroll = distance <= scrollThreshold
    rafLock = false
  })
}

/* 监听内容变化 -> 自动滚到底 */
watch(
    () => props.shellShowResult,
    () => nextTick(scrollToBottom),
    { flush: 'post' }
)

/* 生命周期：挂载时绑定，卸载时清理 */
onMounted(() => {
  nextTick(() => {
    wrapEl = getWrap()
    if (!wrapEl) return
    scrollToBottom()                // 初始滚到底
    wrapEl.addEventListener('scroll', onScroll, { passive: true })
  })
})

onBeforeUnmount(() => {
  if (wrapEl) wrapEl.removeEventListener('scroll', onScroll)
})
</script>

<style scoped>
.sticky-textarea-div {
  background-color: #545c64;
  color: #fff;
  white-space: pre-wrap;
  word-break: break-all;
  padding: 3px;
  border-radius: 6px;
  border-left: 3px solid #6a8d73;
  font-family: 'SF Mono', 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.6;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  font-weight: 300;
  box-shadow: 0 1px 6px rgba(0, 0, 0, 0.08);
  overflow-y: auto;
  overflow-x: hidden;
  display: block;
  height: 100%;
}

#showShellResult {
  height: 100%;
}

/* 深绿色滚动条，不透明 */
:deep(.el-scrollbar__thumb) {
  background-color: #2e7d32 !important;
  border-radius: 4px !important;
  opacity: 1 !important;
}
:deep(.el-scrollbar__thumb:hover) {
  background-color: #388e3c !important;
  opacity: 1 !important;
}
:deep(.el-scrollbar__bar) {
  background-color: #cccccc !important;
}
</style>