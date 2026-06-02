<template>
  <GitActionButton
    compact
    :variant="variant"
    :class="{
      'chat-history-btn--running': running,
      'chat-history-btn--unread': unread,
    }"
    @click="$emit('click')"
  >
    <slot />
    <span v-if="totalCount > 0" class="chat-history-btn__counts">
      {{ runningCount }}/{{ interruptedCount }}/{{ totalCount }}
    </span>
  </GitActionButton>
</template>

<script>
import GitActionButton from '@/components/base/GitActionButton.vue'

export default {
  name: 'ChatHistoryButton',
  components: { GitActionButton },
  props: {
    variant: {
      type: String,
      default: 'info',
    },
    running: {
      type: Boolean,
      default: false,
    },
    runningCount: {
      type: Number,
      default: 0,
    },
    interruptedCount: {
      type: Number,
      default: 0,
    },
    totalCount: {
      type: Number,
      default: 0,
    },
    unread: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['click'],
}
</script>

<style scoped src="@/css/components/shared/ChatHistory.css"></style>
