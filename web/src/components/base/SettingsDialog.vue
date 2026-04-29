<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    :width="width"
    destroy-on-close
    append-to-body
    class="settings-dialog"
    @close="handleClose"
    @update:model-value="handleModelValueChange"
  >
    <div class="settings-dialog__body">
      <slot />
    </div>
  </el-dialog>
</template>

<script>
export default {
  name: 'SettingsDialog',
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: '设置',
    },
    width: {
      type: String,
      default: '76%',
    },
  },
  emits: ['update:modelValue', 'closed'],
  methods: {
    // handleClose 统一转发关闭事件，便于业务页在弹窗收起后刷新页面。
    // Forward close event so host pages can refresh after the settings modal closes.
    handleClose() {
      this.$emit('update:modelValue', false)
      this.$emit('closed')
    },
    // handleModelValueChange 统一同步 v-model，避免弹窗显隐状态漂移。
    // Sync v-model updates from Element Plus to keep dialog visibility consistent.
    handleModelValueChange(value) {
      this.$emit('update:modelValue', value)
    },
  },
}
</script>

<style scoped src="@/css/components/base/SettingsDialog.css"></style>
