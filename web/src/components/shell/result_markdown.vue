<template>
  <MarkdownRenderer id="showShellResult" :source="shellShowResult" :style="{ height: divHeight + 'px' }"></MarkdownRenderer>
</template>

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
  /* 强制显示滚动条 */
  overflow-y: auto; /* 改为 scroll 强制显示 */
  overflow-x: hidden;
  display: block;
  height:100%;
}

#showShellResult{
  height : 100%;
}

/* 更具体的滚动条样式 - 修改为深绿色并去除透明度 */
:deep(.el-scrollbar__thumb) {
  background-color: #2e7d32 !important;
  border-radius: 4px !important;
  opacity: 1 !important; /* 关键：去除透明度 */
}

:deep(.el-scrollbar__thumb:hover) {
  background-color: #388e3c !important;
  opacity: 1 !important; /* 悬停时也保持不透明 */
}

:deep(.el-scrollbar__bar) {
  background-color: #cccccc !important;
}

@keyframes gentle-blink {
  0%, 100% {
    opacity: 0.7;
  }
  50% {
    opacity: 0.3;
  }
}


</style>

<script>
import {
  defineExpose,
  defineComponent,
  inject,
  defineEmits,
  getCurrentInstance,
  reactive,
  computed,
  ref,
  watch
} from 'vue';
import shell from '@/utils/base/shell'
import MarkdownRenderer from "@/components/base/markdown.vue";
import {Close} from '@element-plus/icons-vue'

export default defineComponent({
  components: {MarkdownRenderer, Close},
  props: {
    shellShowResult: {
      type: String
    },
    showModel: {
      type: String
    },
    isRunning: {
      type: Boolean
    },
    divHeight: {
      type: Number,
    }
  },
  setup(props) {
    const proxy = getCurrentInstance().proxy
    const showOk = ref(false)
    /* 1. 计算属性：运行中显示计数，刚结束显示 ok! */
    const btnText = computed(() =>
        showOk.value ? ' run success ! ' : `shell 输出（${props.shellShowResult.length}）`
    )
    /* 2. 监听 isRunning，变 false 时切到 ok!，1.5s 后恢复 */
    watch(() => props.isRunning, val => {
      if (!val) {
        showOk.value = true
        setTimeout(() => showOk.value = false, 1500)
      }
    })
    return {
      btnText,
    }
  }
})
</script>