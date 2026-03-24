<template>
  <!-- 按钮模式 -->
  <div class="shellContainer">
    <pl-button
        v-loading="isRunning"
        class="shellButton"
        round
        type="primary"
        @click="openDrawer"
    >
      {{ btnText }}
    </pl-button>
  </div>

  <!-- 抽屉 -->
  <el-dialog
      v-model="state.showDrawer"
      :append-to-body="false"
      :before-close="drawerClose"
      :header-style="{ padding: '10px' }"
      :lock-scroll="false"
      :modal="true"
      :show-close="false"
      :with-header="true"
      direction="btt"
      modal-class="shellModalDrawer"
      width="70%"
  >
    <template #header>
      <div class="drawer-header">
        <span>{{btnName}}</span>
        <pl-button circle size="small" type="danger" @click="drawerClose">
          <el-icon><Close /></el-icon>
        </pl-button>
      </div>
    </template>

    <template #default>
      <!-- 关键：给 el-scrollbar 加 ref -->
      <el-scrollbar
          id="showShellResult"
          ref="scrollRef"
          style="height: 500px"
      >
        <div
            class="sticky-textarea-div"
            v-html="shellShowResult"
            style="min-height: 500px"
        ></div>
      </el-scrollbar>
    </template>
  </el-dialog>
</template>

<script>
import {
  defineComponent,
  reactive,
  ref,
  computed,
  watch,
  nextTick
} from 'vue'
import { Close } from '@element-plus/icons-vue'

export default defineComponent({
  name: 'ShellOutput',
  components: { Close },
  props: {
    shellShowResult: { type: String, default: '' },
    showModel: { type: String, default: 'button' },
    isRunning: { type: Boolean, default: false },
    divHeight: { type: Number, default: 400 },
    btnName : {type : String , default : '输出'}
  },
  setup(props) {
    const state = reactive({ showDrawer: false })

    /* ---------- 1. 拿到滚动容器 ---------- */
    const scrollRef = ref(null)

    /* ---------- 2. 滚动到底 ---------- */
    function scrollToBottom() {
      nextTick(() => {
        const wrap = scrollRef.value?.wrapRef   // el-scrollbar 的真实滚动层
        if (wrap) wrap.scrollTop = wrap.scrollHeight
      })
    }

    /* ---------- 3. 内容变化自动滚 ---------- */
    watch(
        () => props.shellShowResult,
        () => scrollToBottom(),
        { flush: 'post' }
    )

    /* ---------- 4. 按钮文字 ---------- */
    const showOk = ref(false)
    const btnText = computed(() =>
        showOk.value ? ' run success ! ' : props.btnName + `（${props.shellShowResult.length}）`
    )
    watch(
        () => props.isRunning,
        val => {
          if (!val) {
            showOk.value = true
            setTimeout(() => (showOk.value = false), 1500)
          }
        }
    )

    /* ---------- 5. 开关抽屉 ---------- */
    function openDrawer() {
      if (state.showDrawer) return
      state.showDrawer = true
      scrollToBottom()        // 首次打开也滚到底
    }
    function drawerClose() {
      state.showDrawer = false
    }

    return {
      state,
      scrollRef,   // 模板里需要
      btnText,
      openDrawer,
      drawerClose
    }
  }
})
</script>

<style scoped>
.shellContainer {
  position: fixed;
  width: 98%;
  bottom: 5px;
  text-align: center;
  z-index: 998;
}
.sticky-textarea-div {
  background: #eef3ea;
  color: #435244;
  white-space: pre-wrap;
  word-break: break-all;
  padding: 16px;
  border-radius: 8px;
  border-left: 3px solid #8fae92;
  font-family: 'JetBrains Mono', 'Fira Code', 'SF Mono', 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.7;
  text-rendering: optimizeLegibility;
  -webkit-font-smoothing: antialiased;
  -moz-osx-font-smoothing: grayscale;
  font-weight: 400;
  box-shadow: 0 1px 6px rgba(62, 86, 62, 0.08);
  overflow-y: auto;
  overflow-x: hidden;
  display: block;
  height: 100%;
  transition: all 0.2s ease;
}
#showShellResult {
  height: 100%;
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
.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 0 10px;
}
</style>

