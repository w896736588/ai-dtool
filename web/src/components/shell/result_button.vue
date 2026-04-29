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

<style scoped src="@/css/components/shell/result_button.css"></style>

