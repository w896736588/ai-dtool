<template>
  <el-button
    class="pl-button"
    :class="buttonClassList"
    :loading="mergedLoading"
    v-bind="$attrs"
    @click="handleClick"
  >
    <template v-if="$slots.icon" #icon>
      <slot name="icon" />
    </template>
    <slot />
  </el-button>
</template>

<script>
// BUTTON_TYPE_DEFAULT 统一默认按钮语义，避免页面遗漏 type 时回落到浏览器原生风格。
const BUTTON_TYPE_DEFAULT = 'default'
// BUTTON_TYPE_LIST 控制允许映射的状态类型，便于统一视觉 token。
const BUTTON_TYPE_LIST = ['default', 'primary', 'success', 'info', 'warning', 'danger']
// BUTTON_SIZE_LIST 控制共享尺寸选项，兼容 Element Plus 常用尺寸定义。
const BUTTON_SIZE_LIST = ['large', 'default', 'small']

export default {
  name: 'pl-button',
  inheritAttrs: false,
  props: {
    autoLoading: {
      type: Boolean,
      default: false,
    },
    variant: {
      type: String,
      default: '',
    },
    sizeMode: {
      type: String,
      default: '',
    },
  },
  emits: ['click'],
  data() {
    return {
      innerLoading: false,
    }
  },
  computed: {
    resolvedType() {
      const attrType = typeof this.$attrs.type === 'string' ? this.$attrs.type : ''
      const rawType = this.variant || attrType || BUTTON_TYPE_DEFAULT
      if (BUTTON_TYPE_LIST.includes(rawType)) {
        return rawType
      }
      return BUTTON_TYPE_DEFAULT
    },
    resolvedSize() {
      const attrSize = typeof this.$attrs.size === 'string' ? this.$attrs.size : ''
      const rawSize = this.sizeMode || attrSize || 'default'
      if (BUTTON_SIZE_LIST.includes(rawSize)) {
        return rawSize
      }
      return 'default'
    },
    isLinkButton() {
      return this.$attrs.link === '' || this.$attrs.link === true || this.$attrs.text === '' || this.$attrs.text === true
    },
    isPlainButton() {
      return this.$attrs.plain === '' || this.$attrs.plain === true
    },
    mergedLoading() {
      return Boolean(this.$attrs.loading) || this.innerLoading
    },
    buttonClassList() {
      return [
        `pl-button--${this.resolvedType}`,
        `pl-button--size-${this.resolvedSize}`,
        {
          'pl-button--link': this.isLinkButton,
          'pl-button--plain': this.isPlainButton,
          'pl-button--loading': this.mergedLoading,
        },
      ]
    },
  },
  methods: {
    handleClick(event) {
      if (this.autoLoading && !this.$attrs.loading && !this.$attrs.disabled) {
        this.innerLoading = true
      }
      this.$emit('click', event, () => {
        this.innerLoading = false
      })
    },
  },
}
</script>

<style scoped src="@/css/components/base/button.css"></style>
