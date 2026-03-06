<template>
  <MarkdownRenderer id="showShellResult" :source="shellShowResult" :style="{ height: divHeight + 'px' }"></MarkdownRenderer>
</template>

<style scoped>
#showShellResult{
  display: block;
  width: 100%;
  min-width: 0;
  height: 100%;
  background: #eef3ea;
  color: #435244;
  border-radius: 8px;
  border-left: 3px solid #8fae92;
  box-shadow: 0 1px 6px rgba(62, 86, 62, 0.08);
  box-sizing: border-box;
}

::deep(.el-scrollbar__thumb) {
  background: #a4b7a3 !important;
  border-radius: 4px !important;
  opacity: 0.7 !important;
  transition: opacity 0.2s ease;
}

::deep(.el-scrollbar__thumb:hover) {
  background: #8fa48f !important;
  opacity: 1 !important;
}

::deep(.el-scrollbar__bar) {
  background: #dfe8da !important;
  border-radius: 4px;
}

::deep(#showShellResult pre),
::deep(#showShellResult code) {
  background: #e4ecdf !important;
  color: #3f4f40 !important;
}

::deep(#showShellResult a) {
  color: #4f7d5f !important;
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
