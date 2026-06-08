<template>
  <MdEditor
    v-bind="$attrs"
    :model-value="modelValue"
    class="unified-md-editor"
    @update:modelValue="handleUpdate"
  />
</template>

<script>
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

export default {
  name: 'UnifiedMdEditor',
  components: {
    MdEditor,
  },
  inheritAttrs: false,
  props: {
    modelValue: {
      type: String,
      default: '',
    },
  },
  emits: ['update:modelValue'],
  methods: {
    handleUpdate(value) {
      this.$emit('update:modelValue', value)
    },
  },
}
</script>

<style scoped>
.unified-md-editor :deep(.md-editor) {
  --md-font-size: 13px;
  --md-code-font-size: 13px;
}

.unified-md-editor :deep(.md-editor-input),
.unified-md-editor :deep(.md-editor-preview),
.unified-md-editor :deep(.md-editor-preview-wrapper) {
  font-size: 13px;
  line-height: 1.55;
}
</style>
