<template>
  <el-scrollbar id="showShellResult" :style="scrollbarStyle">
    <div
      ref="contentDivRef"
      class="sticky-textarea-div"
      v-html="shellShowResult"
      :style="contentStyle"
      @mousedown="onContentMouseDown"
    ></div>
  </el-scrollbar>
</template>

<script setup>
/* global defineProps */
import { computed, nextTick, watch, onMounted, onBeforeUnmount, ref } from 'vue'

const props = defineProps({
  shellShowResult: { type: String, default: '' },
  divHeight: { type: Number, default: 200 },
  useContainerHeight: { type: Boolean, default: false }
})

const emit = defineEmits(['selection-change'])

const contentDivRef = ref(null)

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

// 用户选区状态追踪：当用户在内容区有活动文本选区时，通知父组件暂停内容更新
let _selectionCheckTimer = null
let _wasSelecting = false

/**
 * hasActiveSelection - 检测当前是否有文本选区位于本组件内容区域内
 * 父组件（ShellOut.vue）在SSE消息到达时调用此方法，若有选区则缓存消息延迟更新
 * @returns {boolean}
 */
function hasActiveSelection() {
  try {
    const sel = window.getSelection()
    if (!sel || sel.isCollapsed || sel.toString().trim().length === 0) {
      return false
    }
    const contentEl = contentDivRef.value
    if (!contentEl) return false
    // 检查选区是否在内容区域内
    const range = sel.getRangeAt(0)
    return contentEl.contains(range.commonAncestorContainer)
  } catch (_) {
    return false
  }
}

/**
 * _checkSelectionChange - 定时检测选区状态变化，通过emit通知父组件
 */
function _checkSelectionChange() {
  const nowSelecting = hasActiveSelection()
  if (nowSelecting !== _wasSelecting) {
    _wasSelecting = nowSelecting
    emit('selection-change', nowSelecting)
  }
  _selectionCheckTimer = setTimeout(_checkSelectionChange, 200)
}

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

// 用户在内容区按下鼠标时，立即标记停止自动滚动，方便用户选择文本
function onContentMouseDown() {
  autoScroll = false
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
  // 启动选区变化检测
  _selectionCheckTimer = setTimeout(_checkSelectionChange, 300)
})

onBeforeUnmount(() => {
  if (wrapEl) wrapEl.removeEventListener('scroll', onScroll)
  if (_selectionCheckTimer) {
    clearTimeout(_selectionCheckTimer)
    _selectionCheckTimer = null
  }
})

defineExpose({ hasActiveSelection })
</script>

<style scoped src="@/css/components/shell/result_div.css"></style>
