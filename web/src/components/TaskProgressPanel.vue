<template>
  <div v-if="store.summary.total > 0" class="task-progress-panel">
    <!-- 折叠态 -->
    <div
      v-if="collapsed"
      class="task-progress-panel__bar"
      @click="toggle"
    >
      <span class="task-progress-panel__summary">
        <span v-if="store.summary.running > 0" class="task-progress-panel__running-dot"></span>
        <span v-else class="task-progress-panel__summary-icon">&#x1F527;</span>
        {{ store.summary.total }} 个任务
        <template v-if="store.summary.running > 0"> &middot; {{ store.summary.running }} 运行中</template>
        <template v-if="store.summary.completed > 0"> &middot; {{ store.summary.completed }} 已完成</template>
        <template v-if="store.summary.failed > 0"> &middot; {{ store.summary.failed }} 失败</template>
      </span>
      <span class="task-progress-panel__toggle">&#x25B2;</span>
    </div>
    <!-- 展开态 -->
    <div v-else class="task-progress-panel__expanded">
      <div class="task-progress-panel__header" @click="toggle">
        <span class="task-progress-panel__summary">
          <span v-if="store.summary.running > 0" class="task-progress-panel__running-dot"></span>
          <span v-else class="task-progress-panel__summary-icon">&#x1F527;</span>
          {{ store.summary.total }} 个任务
          <template v-if="store.summary.running > 0"> &middot; {{ store.summary.running }} 运行中</template>
          <template v-if="store.summary.completed > 0"> &middot; {{ store.summary.completed }} 已完成</template>
          <template v-if="store.summary.failed > 0"> &middot; {{ store.summary.failed }} 失败</template>
        </span>
        <span class="task-progress-panel__toggle">&#x25BC;</span>
      </div>
      <div class="task-progress-panel__list">
        <div
          v-for="t in store.tasks"
          :key="t.taskId"
          class="task-progress-panel__item"
          :class="{ 'task-progress-panel__item--clickable': t._msgIndex !== undefined }"
          @click="handleItemClick(t)"
        >
          <span class="task-progress-panel__item-status">
            <span v-if="t.status === 'running'" class="task-progress-panel__running-dot"></span>
            <template v-else>{{ statusIcon(t.status) }}</template>
          </span>
          <span class="task-progress-panel__item-desc" :title="t.description">{{ t.description || '-' }}</span>
          <span class="task-progress-panel__item-meta">
            <template v-if="t.lastToolName">{{ t.lastToolName }}</template>
            <template v-if="t.usage && t.usage.duration_ms"> &middot; {{ formatDuration(t.usage.duration_ms) }}</template>
            <template v-if="t.usage && t.usage.total_tokens"> &middot; {{ formatTokens(t.usage.total_tokens) }}</template>
          </span>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import taskProgressStore from '@/utils/task_progress_store'

// 显示常量
const STATUS_ICON_MAP = {
  started: '\u23F3',
  running: '\u23F3',
  completed: '\u2705',
  failed: '\u274C',
}

export default {
  name: 'TaskProgressPanel',
  data() {
    return {
      store: taskProgressStore,
      collapsed: true,
    }
  },
  methods: {
    toggle() {
      this.collapsed = !this.collapsed
    },
    statusIcon(status) {
      return STATUS_ICON_MAP[status] || '\u{1F527}'
    },
    formatDuration(ms) {
      if (!ms) return ''
      if (ms < 1000) return ms + 'ms'
      return (ms / 1000).toFixed(1) + 's'
    },
    formatTokens(n) {
      if (!n) return ''
      if (n >= 1000) return (n / 1000).toFixed(1) + 'k'
      return String(n)
    },
    handleItemClick(t) {
      if (t._msgIndex !== undefined) {
        this.$emit('scroll-to-msg', t._msgIndex)
      }
    },
  },
}
</script>

<style lang="scss" scoped>
.task-progress-panel {
  border-top: 1px solid #ebeef5;
  background: #fafafa;
  font-size: 12px;
  color: #606266;
  user-select: none;
  flex-shrink: 0;
}

.task-progress-panel__bar,
.task-progress-panel__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 6px 12px;
  cursor: pointer;
  &:hover {
    background: #f0f2f5;
  }
}

.task-progress-panel__summary {
  font-size: 12px;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 6px;
}

.task-progress-panel__summary-icon {
  flex-shrink: 0;
}

.task-progress-panel__running-dot {
  display: inline-block;
  width: 10px;
  height: 10px;
  border-radius: 50%;
  border: 2px solid #409eff;
  border-top-color: transparent;
  animation: task-dot-spin 0.8s linear infinite;
  flex-shrink: 0;
}

@keyframes task-dot-spin {
  to { transform: rotate(360deg); }
}

.task-progress-panel__toggle {
  font-size: 10px;
  color: #909399;
}

.task-progress-panel__list {
  max-height: 200px;
  overflow-y: auto;
  border-top: 1px solid #ebeef5;
}

.task-progress-panel__item {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 4px 12px;
  &:not(:last-child) {
    border-bottom: 1px solid #f2f3f5;
  }
}

.task-progress-panel__item-status {
  flex-shrink: 0;
  font-size: 12px;
  width: 16px;
  text-align: center;
}

.task-progress-panel__item-desc {
  flex: 1;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  min-width: 0;
}

.task-progress-panel__item-meta {
  flex-shrink: 0;
  color: #909399;
  font-size: 11px;
  white-space: nowrap;
}

.task-progress-panel__item--clickable {
  cursor: pointer;
  &:hover {
    background: #ecf5ff;
  }
}
</style>
