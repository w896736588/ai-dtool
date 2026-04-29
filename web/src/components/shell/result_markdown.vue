<template>
  <MarkdownRenderer id="showShellResult" :source="shellShowResult" :style="{ height: divHeight + 'px' }"></MarkdownRenderer>
</template>

<style scoped src="@/css/components/shell/result_markdown.css"></style>

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
