<template>
  <el-dialog
    v-model="visible"
    title="工具权限审批"
    width="520px"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="false"
  >
    <div class="permission-dialog-body">
      <div class="permission-info">
        <span class="permission-label">工具</span>
        <el-tag type="warning" size="default">{{ request.tool_name }}</el-tag>
      </div>
      <div class="permission-info" v-if="request.input">
        <span class="permission-label">参数</span>
        <pre class="permission-input">{{ formatInput(request.input) }}</pre>
      </div>
      <!-- 超时倒计时 -->
      <div class="permission-timeout" v-if="timeLeft > 0">
        <span class="timeout-text">
          {{ formatTime(timeLeft) }}
        </span>
      </div>
    </div>

    <template #footer>
      <div class="permission-footer">
        <el-button type="info" @click="handleReject" :loading="loading">拒绝</el-button>
        <el-button type="primary" @click="handleApprove" :loading="loading">允许</el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script>
export default {
  name: 'PermissionDialog',
  emits: ['approve', 'reject'],
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    request: {
      type: Object,
      default: () => ({
        request_id: '',
        tool_name: '',
        input: null,
        session_id: '',
        chat_id: 0,
      }),
    },
    // 总超时时间（秒）
    totalTimeout: {
      type: Number,
      default: 300,
    },
  },
  data() {
    return {
      loading: false,
      timeLeft: this.totalTimeout,
      _timer: null,
    }
  },
  computed: {
    visible: {
      get() {
        return this.modelValue
      },
      set(val) {
        this.$emit('update:modelValue', val)
      },
    },
  },
  watch: {
    modelValue(val) {
      if (val) {
        this.timeLeft = this.totalTimeout
        this.startTimer()
      } else {
        this.stopTimer()
      }
    },
  },
  methods: {
    startTimer() {
      this.stopTimer()
      this._timer = setInterval(() => {
        if (this.timeLeft > 0) {
          this.timeLeft--
        }
        if (this.timeLeft <= 0) {
          this.stopTimer()
        }
      }, 1000)
    },
    stopTimer() {
      if (this._timer) {
        clearInterval(this._timer)
        this._timer = null
      }
    },
    formatInput(input) {
      if (typeof input === 'string') return input
      try {
        return JSON.stringify(input, null, 2)
      } catch {
        return String(input)
      }
    },
    formatTime(seconds) {
      const m = Math.floor(seconds / 60)
      const s = seconds % 60
      if (m > 0) {
        return `超时等待 ${m}:${String(s).padStart(2, '0')}`
      }
      return `超时等待 ${s}秒`
    },
    handleApprove() {
      this.loading = true
      this.$emit('approve', this.request)
    },
    handleReject() {
      this.loading = true
      this.$emit('reject', this.request)
    },
    resetLoading() {
      this.loading = false
    },
  },
  beforeUnmount() {
    this.stopTimer()
  },
}
</script>

<style scoped>
.permission-dialog-body {
  padding: 10px 0;
}
.permission-info {
  margin-bottom: 16px;
}
.permission-label {
  font-weight: 600;
  margin-right: 8px;
  color: #606266;
}
.permission-input {
  margin-top: 8px;
  margin-bottom: 0;
  padding: 12px;
  background: #f5f7fa;
  border-radius: 4px;
  font-size: 13px;
  line-height: 1.5;
  max-height: 220px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-all;
}
.permission-timeout {
  display: flex;
  align-items: center;
  justify-content: center;
  margin-top: 12px;
}
.timeout-text {
  font-size: 13px;
  color: #909399;
}
.permission-footer {
  display: flex;
  justify-content: flex-end;
  gap: 12px;
}
</style>
