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

<style scoped>
.pl-button {
  --pl-button-text-color: #4f5f47;
  --pl-button-border-color: #d8ded2;
  --pl-button-background-color: #f6f8f3;
  --pl-button-hover-text-color: #3f6f3f;
  --pl-button-hover-border-color: #bfd1bf;
  --pl-button-hover-background-color: #eef4ea;
  --pl-button-link-color: #4d7a4d;
  --pl-button-link-hover-color: #345f34;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  border-radius: 8px;
  border-color: var(--pl-button-border-color) !important;
  background: var(--pl-button-background-color) !important;
  color: var(--pl-button-text-color) !important;
  box-shadow: none !important;
  transition: background-color 0.2s ease, border-color 0.2s ease, color 0.2s ease, opacity 0.2s ease;
}

.pl-button:hover,
.pl-button:focus-visible {
  border-color: var(--pl-button-hover-border-color) !important;
  background: var(--pl-button-hover-background-color) !important;
  color: var(--pl-button-hover-text-color) !important;
  box-shadow: none !important;
}

.pl-button :deep(span) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  gap: 4px;
}

.pl-button--size-large {
  min-height: 40px;
  padding: 10px 18px;
  font-size: 14px;
}

.pl-button--size-default {
  min-height: 34px;
  padding: 8px 14px;
  font-size: 13px;
}

.pl-button--size-small {
  min-height: 28px;
  padding: 6px 10px;
  font-size: 12px;
}

.pl-button--default {
  --pl-button-text-color: #4f5f47;
  --pl-button-border-color: #d8ded2;
  --pl-button-background-color: #f6f8f3;
  --pl-button-hover-text-color: #3f6f3f;
  --pl-button-hover-border-color: #bfd1bf;
  --pl-button-hover-background-color: #eef4ea;
  --pl-button-link-color: #567256;
  --pl-button-link-hover-color: #355535;
}

.pl-button--primary {
  --pl-button-text-color: #ffffff;
  --pl-button-border-color: #5d895d;
  --pl-button-background-color: linear-gradient(180deg, #6d9a6d 0%, #5d895d 100%);
  --pl-button-hover-text-color: #ffffff;
  --pl-button-hover-border-color: #4f7a4f;
  --pl-button-hover-background-color: linear-gradient(180deg, #5f8b5f 0%, #4f7a4f 100%);
  --pl-button-link-color: #4a7a4a;
  --pl-button-link-hover-color: #315b31;
}

.pl-button--success {
  --pl-button-text-color: #ffffff;
  --pl-button-border-color: #5f9c7a;
  --pl-button-background-color: linear-gradient(180deg, #71af8d 0%, #5f9c7a 100%);
  --pl-button-hover-text-color: #ffffff;
  --pl-button-hover-border-color: #4d8567;
  --pl-button-hover-background-color: linear-gradient(180deg, #629e7d 0%, #4d8567 100%);
  --pl-button-link-color: #4e8a69;
  --pl-button-link-hover-color: #31654a;
}

.pl-button--info {
  --pl-button-text-color: #455b72;
  --pl-button-border-color: #d6dee7;
  --pl-button-background-color: #f4f7fa;
  --pl-button-hover-text-color: #35465a;
  --pl-button-hover-border-color: #bdc9d8;
  --pl-button-hover-background-color: #e9eef4;
  --pl-button-link-color: #4f6680;
  --pl-button-link-hover-color: #34495f;
}

.pl-button--warning {
  --pl-button-text-color: #7b5524;
  --pl-button-border-color: #ead8bb;
  --pl-button-background-color: #fbf5ea;
  --pl-button-hover-text-color: #664419;
  --pl-button-hover-border-color: #ddc49e;
  --pl-button-hover-background-color: #f4ead7;
  --pl-button-link-color: #8a5b22;
  --pl-button-link-hover-color: #6b4314;
}

.pl-button--danger {
  --pl-button-text-color: #ffffff;
  --pl-button-border-color: #d65c5c;
  --pl-button-background-color: linear-gradient(180deg, #de6f6f 0%, #d65c5c 100%);
  --pl-button-hover-text-color: #ffffff;
  --pl-button-hover-border-color: #bb4747;
  --pl-button-hover-background-color: linear-gradient(180deg, #c95757 0%, #bb4747 100%);
  --pl-button-link-color: #cf4b4b;
  --pl-button-link-hover-color: #a83838;
}

.pl-button--plain.pl-button--primary,
.pl-button--plain.pl-button--success,
.pl-button--plain.pl-button--danger {
  --pl-button-border-color: color-mix(in srgb, currentColor 28%, #d8ded2);
}

.pl-button--link {
  border-color: transparent !important;
  background: transparent !important;
  color: var(--pl-button-link-color) !important;
  padding-left: 4px;
  padding-right: 4px;
  min-height: auto;
}

.pl-button--link:hover,
.pl-button--link:focus-visible {
  border-color: transparent !important;
  background: transparent !important;
  color: var(--pl-button-link-hover-color) !important;
}

.pl-button.is-disabled,
.pl-button.is-disabled:hover,
.pl-button[disabled],
.pl-button[disabled]:hover {
  opacity: 0.6;
  box-shadow: none !important;
}
</style>
