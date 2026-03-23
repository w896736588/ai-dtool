<template>
  <el-button
    class="git-action-button"
    :class="buttonClassList"
    :type="buttonType"
    :plain="buttonPlain"
    v-bind="$attrs"
  >
    <template v-if="$slots.icon" #icon>
      <slot name="icon" />
    </template>
    <slot />
  </el-button>
</template>

<script>
// BUTTON_DEFAULT_TYPE keeps the Element Plus status aligned with the Git page button behavior.
const BUTTON_DEFAULT_TYPE = 'primary'
// BUTTON_DEFAULT_PLAIN preserves the plain button structure while the shared CSS controls the final look.
const BUTTON_DEFAULT_PLAIN = true
// BUTTON_VARIANT_* defines reusable semantic styles so pages can distinguish actions without redefining CSS.
const BUTTON_VARIANT_PRIMARY = 'primary'
const BUTTON_VARIANT_INFO = 'info'
const BUTTON_VARIANT_WARNING = 'warning'
const BUTTON_VARIANT_DANGER = 'danger'
const BUTTON_VARIANT_LIST = [
  BUTTON_VARIANT_PRIMARY,
  BUTTON_VARIANT_INFO,
  BUTTON_VARIANT_WARNING,
  BUTTON_VARIANT_DANGER,
]
// BUTTON_SIZE_DEFAULT 表示默认按钮尺寸。
const BUTTON_SIZE_DEFAULT = 'default'
// BUTTON_SIZE_COMPACT_SMALL 用于侧边栏等空间更紧凑的场景。
const BUTTON_SIZE_COMPACT_SMALL = 'compact-small'
// BUTTON_SIZE_LIST 控制允许传入的共享尺寸选项。
const BUTTON_SIZE_LIST = [
  BUTTON_SIZE_DEFAULT,
  BUTTON_SIZE_COMPACT_SMALL,
]

export default {
  name: 'GitActionButton',
  inheritAttrs: false,
  props: {
    compact: {
      type: Boolean,
      default: false,
    },
    variant: {
      type: String,
      default: BUTTON_VARIANT_PRIMARY,
      validator(value) {
        return BUTTON_VARIANT_LIST.includes(value)
      },
    },
    sizeMode: {
      type: String,
      default: BUTTON_SIZE_DEFAULT,
      validator(value) {
        return BUTTON_SIZE_LIST.includes(value)
      },
    },
  },
  data() {
    return {
      buttonType: BUTTON_DEFAULT_TYPE,
      buttonPlain: BUTTON_DEFAULT_PLAIN,
    }
  },
  computed: {
    buttonClassList() {
      return {
        'git-action-button--compact': this.compact,
        'git-action-button--compact-small': this.compact && this.sizeMode === BUTTON_SIZE_COMPACT_SMALL,
        [`git-action-button--${this.variant}`]: true,
      }
    },
  },
}
</script>

<style scoped>
.git-action-button {
  /* Shared Git page button tokens for reuse across feature modules. */
  display: inline-flex;
  align-items: center;
  justify-content: center;
  box-sizing: border-box;
  border-radius: 8px;
  box-shadow: none !important;
  height: 34px;
  padding: 7px 14px;
  font-size: 13px;
  line-height: 1;
  text-align: center;
  white-space: nowrap;
  vertical-align: middle;
  transition: background-color 0.2s ease, border-color 0.2s ease, color 0.2s ease;
}

.git-action-button:hover,
.git-action-button:focus-visible {
  box-shadow: none !important;
}

.git-action-button--primary {
  --git-button-text-color: #4f804f;
  --git-button-border-color: #d8ded2;
  --git-button-background-color: #f6f8f3;
  --git-button-hover-text-color: #3f6f3f;
  --git-button-hover-border-color: #bfd1bf;
  --git-button-hover-background-color: #eef4ea;
}

.git-action-button--info {
  --git-button-text-color: #4b627a;
  --git-button-border-color: #d3dbe5;
  --git-button-background-color: #f4f7fa;
  --git-button-hover-text-color: #384d63;
  --git-button-hover-border-color: #bcc8d6;
  --git-button-hover-background-color: #e9eef4;
}

.git-action-button--warning {
  --git-button-text-color: #8a5b22;
  --git-button-border-color: #ead8bb;
  --git-button-background-color: #fbf5ea;
  --git-button-hover-text-color: #724816;
  --git-button-hover-border-color: #ddc49e;
  --git-button-hover-background-color: #f4ead7;
}

.git-action-button--danger {
  --git-button-text-color: #ffffff;
  --git-button-border-color: #d65c5c;
  --git-button-background-color: #d65c5c;
  --git-button-hover-text-color: #ffffff;
  --git-button-hover-border-color: #bb4747;
  --git-button-hover-background-color: #bb4747;
}

.git-action-button {
  border-color: var(--git-button-border-color) !important;
  background: var(--git-button-background-color) !important;
  color: var(--git-button-text-color) !important;
}

.git-action-button:hover,
.git-action-button:focus-visible {
  border-color: var(--git-button-hover-border-color) !important;
  background: var(--git-button-hover-background-color) !important;
  color: var(--git-button-hover-text-color) !important;
}

.git-action-button :deep(span) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  height: 100%;
  width: 100%;
  line-height: 1;
}

.git-action-button--compact {
  height: 30px;
  padding: 5px 10px;
  font-size: 12px;
  line-height: 1;
}

.git-action-button--compact-small {
  height: 26px;
  padding: 0 8px;
  font-size: 11px;
  line-height: 1;
  border-radius: 7px;
}
</style>
