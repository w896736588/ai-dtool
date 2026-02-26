<template>
  <!-- 按钮模式 -->
  <div class="shellContainer">
    <el-button
        v-loading="isRunning"
        class="shellButton"
        round
        type="primary"
        @click="openDrawer"
    >
      {{ btnText }}
    </el-button>
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
        <el-button circle size="small" type="danger" @click="drawerClose">
          <el-icon><Close /></el-icon>
        </el-button>
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
/* 保持你原来的样式，无需改动 */
.shellContainer {
  position: fixed;
  width: 98%;
  bottom: 5px;
  text-align: center;
  z-index: 998;
}
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
  background-color: #ccc !important;
}
.drawer-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  width: 100%;
  padding: 0 10px;
}
</style>