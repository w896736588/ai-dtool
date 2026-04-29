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

<style scoped src="@/css/components/shell/result_div.css"></style>
