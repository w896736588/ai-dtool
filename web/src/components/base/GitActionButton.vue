<template>
  <pl-button
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
  </pl-button>
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
const BUTTON_VARIANT_SUCCESS = 'success'
const BUTTON_VARIANT_LIST = [
  BUTTON_VARIANT_PRIMARY,
  BUTTON_VARIANT_INFO,
  BUTTON_VARIANT_WARNING,
  BUTTON_VARIANT_DANGER,
  BUTTON_VARIANT_SUCCESS,
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

<style scoped src="@/css/components/base/GitActionButton.css"></style>

