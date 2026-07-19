<template>
  <!-- 可拖动录制工具条 -->
  <div
    v-if="visible"
    class="e2e-record-toolbar"
    :style="{ left: posX + 'px', top: posY + 'px' }"
    @mousedown="startDrag"
  >
    <div class="toolbar-header" @dblclick="toggleMin">
      <span class="title">
        <el-icon><VideoCamera /></el-icon>
        <span class="title-text">录制工具条</span>
        <el-tag v-if="recording" type="success" size="small">录制中</el-tag>
        <el-tag v-else type="info" size="small">已暂停</el-tag>
      </span>
      <span class="actions">
        <el-button text size="small" @click.stop="toggleMin" title="折叠/展开">
          <el-icon><ArrowDown v-if="!minimized" /><ArrowUp v-else /></el-icon>
        </el-button>
        <el-button text size="small" @click.stop="$emit('close')" title="关闭">
          <el-icon><Close /></el-icon>
        </el-button>
      </span>
    </div>

    <div v-show="!minimized" class="toolbar-body">
      <div class="toolbar-row">
        <el-tooltip content="录制点击元素（在页面中点击目标元素，自动捕获 selector）" placement="bottom">
          <el-button :type="mode === 'click' ? 'danger' : 'primary'" size="small" @click="setMode('click')">
            <el-icon><Pointer /></el-icon>
            元素点击
          </el-button>
        </el-tooltip>
        <el-tooltip content="录制坐标点击（点击页面坐标，按 viewport 比例换算）" placement="bottom">
          <el-button :type="mode === 'click_xy' ? 'danger' : 'primary'" size="small" @click="setMode('click_xy')">
            <el-icon><Aim /></el-icon>
            坐标点击
          </el-button>
        </el-tooltip>
        <el-tooltip content="录制输入（监听 input 事件，捕获 value 后清空，回放时重新输入）" placement="bottom">
          <el-button :type="mode === 'input' ? 'danger' : 'primary'" size="small" @click="setMode('input')">
            <el-icon><EditPen /></el-icon>
            输入
          </el-button>
        </el-tooltip>
        <el-tooltip content="录制滚动（按滚轮 / 滑动距离录制）" placement="bottom">
          <el-button :type="mode === 'scroll' ? 'danger' : 'primary'" size="small" @click="setMode('scroll')">
            <el-icon><Bottom /></el-icon>
            滚动
          </el-button>
        </el-tooltip>
        <el-tooltip content="录制右键" placement="bottom">
          <el-button :type="mode === 'right_click' ? 'danger' : 'primary'" size="small" @click="setMode('right_click')">
            <el-icon><Pointer /></el-icon>
            右键
          </el-button>
        </el-tooltip>
      </div>

      <div class="toolbar-row">
        <el-tooltip content="步骤列表（已录制的步骤）" placement="bottom">
          <el-button :type="mode === 'list' ? 'danger' : 'warning'" size="small" @click="setMode('list')">
            <el-icon><List /></el-icon>
            步骤列表 ({{ stepCount }})
          </el-button>
        </el-tooltip>
        <el-button :type="recording ? 'danger' : 'success'" size="small" @click="toggleRecording">
          <el-icon><VideoPlay v-if="!recording" /><VideoPause v-else /></el-icon>
          {{ recording ? '暂停' : '开始' }}
        </el-button>
        <el-button type="info" size="small" @click="$emit('view-session')">
          <el-icon><View /></el-icon>
          会话详情
        </el-button>
      </div>

      <div class="toolbar-row meta">
        <span class="meta-text">会话：{{ sessionId || '未创建' }}</span>
        <span class="meta-text">模式：{{ modeLabel }}</span>
      </div>

      <div v-if="mode === 'list'" class="step-list-preview">
        <el-empty v-if="!previewSteps.length" description="暂无步骤" :image-size="50" />
        <div v-else class="step-list-scroll">
          <div v-for="(s, i) in previewSteps" :key="s.id || i" class="step-row">
            <span class="step-num">{{ i + 1 }}</span>
            <span class="step-type">{{ s.type }}</span>
            <span class="step-desc">{{ s.description || (s.config ? JSON.stringify(s.config).slice(0, 30) : '') }}</span>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import {
  VideoCamera,
  ArrowDown,
  ArrowUp,
  Close,
  Pointer,
  Aim,
  EditPen,
  Bottom,
  List,
  VideoPlay,
  VideoPause,
  View,
} from '@element-plus/icons-vue'

export default {
  name: 'RecordToolbar',
  components: {
    VideoCamera,
    ArrowDown,
    ArrowUp,
    Close,
    Pointer,
    Aim,
    EditPen,
    Bottom,
    List,
    VideoPlay,
    VideoPause,
    View,
  },
  props: {
    visible: { type: Boolean, default: true },
    sessionId: { type: [String, Number], default: '' },
    recording: { type: Boolean, default: true },
    mode: { type: String, default: 'click' },
    stepCount: { type: Number, default: 0 },
    previewSteps: { type: Array, default: () => [] },
  },
  emits: ['close', 'mode-change', 'toggle-recording', 'view-session'],
  data() {
    return {
      posX: 20,
      posY: 80,
      minimized: false,
      dragging: false,
      dragStart: { x: 0, y: 0, ox: 0, oy: 0 },
    }
  },
  computed: {
    modeLabel() {
      const map = {
        click: '元素点击',
        click_xy: '坐标点击',
        input: '输入',
        scroll: '滚动',
        right_click: '右键',
        list: '步骤列表',
      }
      return map[this.mode] || this.mode
    },
  },
  mounted() {
    // 默认放在右上角
    if (typeof window !== 'undefined') {
      this.posX = Math.max(window.innerWidth - 360, 20)
      this.posY = 80
    }
  },
  methods: {
    setMode(m) {
      this.$emit('mode-change', m)
    },
    toggleRecording() {
      this.$emit('toggle-recording')
    },
    toggleMin() {
      this.minimized = !this.minimized
    },
    startDrag(ev) {
      // 只允许 header 区域触发拖动，避免按钮误触
      if (ev.target.closest('.el-button')) return
      this.dragging = true
      this.dragStart = { x: ev.clientX, y: ev.clientY, ox: this.posX, oy: this.posY }
      window.addEventListener('mousemove', this.onDrag)
      window.addEventListener('mouseup', this.endDrag)
    },
    onDrag(ev) {
      if (!this.dragging) return
      const dx = ev.clientX - this.dragStart.x
      const dy = ev.clientY - this.dragStart.y
      this.posX = this.dragStart.ox + dx
      this.posY = this.dragStart.oy + dy
      // 限制边界
      const maxX = (window.innerWidth || 1200) - 200
      const maxY = (window.innerHeight || 800) - 60
      if (this.posX < 0) this.posX = 0
      if (this.posY < 0) this.posY = 0
      if (this.posX > maxX) this.posX = maxX
      if (this.posY > maxY) this.posY = maxY
    },
    endDrag() {
      this.dragging = false
      window.removeEventListener('mousemove', this.onDrag)
      window.removeEventListener('mouseup', this.endDrag)
    },
  },
}
</script>

<style scoped>
.e2e-record-toolbar {
  position: fixed;
  z-index: 9999;
  width: 340px;
  background: #fff;
  border: 1px solid #dcdfe6;
  border-radius: 8px;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.18);
  user-select: none;
  font-size: 12px;
}
.toolbar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  padding: 8px 10px;
  background: linear-gradient(90deg, #409eff, #66b1ff);
  color: #fff;
  border-radius: 8px 8px 0 0;
  cursor: move;
}
.toolbar-header .title {
  display: inline-flex;
  align-items: center;
  gap: 6px;
}
.toolbar-header .title-text {
  font-weight: 600;
}
.toolbar-header .actions {
  display: inline-flex;
  align-items: center;
}
.toolbar-body {
  padding: 8px 10px;
  display: flex;
  flex-direction: column;
  gap: 6px;
}
.toolbar-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}
.toolbar-row.meta {
  background: #f5f7fa;
  border-radius: 4px;
  padding: 4px 8px;
  color: #606266;
}
.meta-text {
  font-size: 11px;
}
.step-list-preview {
  margin-top: 4px;
  border-top: 1px dashed #dcdfe6;
  padding-top: 6px;
  max-height: 220px;
  overflow: auto;
}
.step-list-scroll {
  display: flex;
  flex-direction: column;
  gap: 4px;
}
.step-row {
  display: flex;
  gap: 6px;
  align-items: center;
  padding: 4px 6px;
  background: #f5f7fa;
  border-radius: 4px;
}
.step-num {
  width: 22px;
  text-align: center;
  font-weight: 600;
  color: #409eff;
}
.step-type {
  font-family: monospace;
  background: #ecf5ff;
  color: #409eff;
  padding: 1px 6px;
  border-radius: 4px;
}
.step-desc {
  flex: 1;
  color: #606266;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}
</style>
