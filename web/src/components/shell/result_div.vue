<template>
  <el-scrollbar id="showShellResult" :style="scrollbarStyle">
    <div
      class="sticky-textarea-div"
      v-html="shellShowResult"
      :style="contentStyle"
    ></div>
  </el-scrollbar>
</template>

<script setup>
/* global defineProps */
import { computed, nextTick, watch, onMounted, onBeforeUnmount } from 'vue'

const props = defineProps({
  shellShowResult: { type: String, default: '' },
  divHeight: { type: Number, default: 200 },
  useContainerHeight: { type: Boolean, default: false }
})

const scrollbarStyle = computed(() => {
  if (props.useContainerHeight) {
    return { height: '100%' }
  }
  return { height: props.divHeight - 17 + 'px' }
})

const contentStyle = computed(() => {
  if (props.useContainerHeight) {
    return { minHeight: '100%' }
  }
  return { minHeight: props.divHeight - 25 + 'px' }
})

const scrollThreshold = 10
let autoScroll = true
let wrapEl = null
let rafLock = false

function getWrap() {
  const sb = document.getElementById('showShellResult')
  return sb?.parentNode
}

function scrollToBottom() {
  if (!autoScroll || !wrapEl) return
  wrapEl.scrollTop = wrapEl.scrollHeight
}

function onScroll() {
  if (rafLock) return
  rafLock = true
  window.requestAnimationFrame(() => {
    const distance = wrapEl.scrollHeight - wrapEl.scrollTop - wrapEl.clientHeight
    autoScroll = distance <= scrollThreshold
    rafLock = false
  })
}

watch(
  () => props.shellShowResult,
  () => nextTick(scrollToBottom),
  { flush: 'post' }
)

onMounted(() => {
  nextTick(() => {
    wrapEl = getWrap()
    if (!wrapEl) return
    scrollToBottom()
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
