<template>
  <el-dialog
    v-model="dialogVisible"
    :title="title"
    width="80%"
    top="5vh"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    @open="onOpen"
    @closed="onClosed"
  >
    <div v-if="step" class="step-confirm">
      <!-- 1. 步骤基本信息 -->
      <el-card shadow="never" class="block">
        <template #header>
          <div class="card-header">
            <span><el-icon><InfoFilled /></el-icon> 步骤信息</span>
            <el-tag size="small">{{ step.type }}</el-tag>
          </div>
        </template>
        <el-descriptions :column="2" border size="small">
          <el-descriptions-item label="步骤 ID">{{ step.id }}</el-descriptions-item>
          <el-descriptions-item label="类型">{{ step.type }} (v{{ step.version || '1.0' }})</el-descriptions-item>
          <el-descriptions-item label="录制时间">{{ formatTime(step.recorded_at) }}</el-descriptions-item>
          <el-descriptions-item label="等待时间">
            <el-input-number v-model="localStep.wait_after_ms" :min="0" :max="60000" :step="100" size="small" />
            <span style="margin-left: 6px; color: #909399; font-size: 12px">毫秒（步骤执行后等待）</span>
          </el-descriptions-item>
          <el-descriptions-item label="步骤描述" :span="2">
            <el-input v-model="localStep.description" placeholder="可选：对该步骤的描述" />
          </el-descriptions-item>
        </el-descriptions>
      </el-card>

      <!-- 2. 步骤配置（按 type 渲染不同字段） -->
      <el-card shadow="never" class="block">
        <template #header>
          <div class="card-header">
            <span><el-icon><Setting /></el-icon> 步骤配置</span>
            <el-button size="small" type="primary" @click="replayStep" :loading="replaying">
              <el-icon><VideoPlay /></el-icon>
              试回放
            </el-button>
          </div>
        </template>
        <pre class="config-preview">{{ formatJson(localStep.config) }}</pre>

        <!-- 坐标点击：可视化展示 -->
        <div v-if="isPositionClick" class="config-detail">
          <el-row :gutter="10">
            <el-col :span="6"><el-form-item label="X">{{ localStep.config.x }}</el-form-item></el-col>
            <el-col :span="6"><el-form-item label="Y">{{ localStep.config.y }}</el-form-item></el-col>
            <el-col :span="6"><el-form-item label="视口宽">{{ localStep.config.viewport_width }}</el-form-item></el-col>
            <el-col :span="6"><el-form-item label="视口高">{{ localStep.config.viewport_height }}</el-form-item></el-col>
          </el-row>
        </div>

        <!-- 输入：展示 selector / value -->
        <div v-if="step.type === 'input_v1'" class="config-detail">
          <el-row :gutter="10">
            <el-col :span="12"><el-form-item label="选择器"><code>{{ localStep.config.selector }}</code></el-form-item></el-col>
            <el-col :span="12"><el-form-item label="输入内容">{{ localStep.config.value }}</el-form-item></el-col>
          </el-row>
        </div>

        <!-- 滚动：展示 delta_x / delta_y -->
        <div v-if="step.type === 'scroll_v1'" class="config-detail">
          <el-row :gutter="10">
            <el-col :span="12"><el-form-item label="水平 delta">{{ localStep.config.delta_x }}</el-form-item></el-col>
            <el-col :span="12"><el-form-item label="垂直 delta">{{ localStep.config.delta_y }}</el-form-item></el-col>
          </el-row>
        </div>

        <div v-if="replayResult" class="replay-result" :class="{ ok: replayResult.success, fail: !replayResult.success }">
          <el-alert :title="replayResult.success ? '回放成功' : '回放失败'" :type="replayResult.success ? 'success' : 'error'" :closable="false">
            <div v-if="replayResult.error">{{ replayResult.error }}</div>
            <pre v-if="replayResult.log" class="replay-log">{{ replayResult.log }}</pre>
          </el-alert>
        </div>
      </el-card>

      <!-- 3. 断言编辑 -->
      <el-card shadow="never" class="block">
        <template #header>
          <div class="card-header">
            <span><el-icon><CircleCheck /></el-icon> 断言（可多个，逻辑关系可设 AND / OR）</span>
            <div>
              <el-select v-model="logicOp" size="small" style="width: 110px" v-if="localStep.assertions.length > 1">
                <el-option label="全部满足 (AND)" value="and" />
                <el-option label="任一满足 (OR)" value="or" />
              </el-select>
              <el-button size="small" type="primary" @click="addAssertion" style="margin-left: 8px">
                <el-icon><Plus /></el-icon>
                添加断言
              </el-button>
            </div>
          </div>
        </template>
        <el-empty v-if="!localStep.assertions.length" description="暂无断言，可点击右上角添加" :image-size="60" />
        <div v-for="(a, idx) in localStep.assertions" :key="idx" class="assertion-row">
          <el-row :gutter="8" align="middle">
            <el-col :span="6">
              <el-select v-model="a.type" size="small" @change="onAssertionTypeChange(a)">
                <el-option v-for="t in assertionTypes" :key="t.type" :value="t.type" :label="t.label" />
              </el-select>
            </el-col>
            <el-col :span="14">
              <!-- 文本断言 -->
              <template v-if="a.type === 'text_v1' || a.type === 'text_v2'">
                <el-input v-model="a.config.text" size="small" placeholder="期望文本" />
                <el-select v-model="a.config.match" size="small" style="width: 80px; margin-left: 4px">
                  <el-option label="包含" value="contains" />
                  <el-option label="等于" value="equals" />
                  <el-option label="正则" value="regex" />
                </el-select>
              </template>
              <!-- URL 断言 -->
              <template v-else-if="a.type === 'url_v1'">
                <el-input v-model="a.config.pattern" size="small" placeholder="URL 正则或包含串" />
              </template>
              <!-- 元素存在 -->
              <template v-else-if="a.type === 'element_v1'">
                <el-input v-model="a.config.selector" size="small" placeholder="CSS 选择器" />
              </template>
              <!-- 变量 -->
              <template v-else-if="a.type === 'variable_v1'">
                <el-input v-model="a.config.var_name" size="small" placeholder="变量名" style="width: 30%" />
                <el-select v-model="a.config.compare" size="small" style="width: 80px; margin-left: 4px">
                  <el-option label="等于" value="equals" />
                  <el-option label="包含" value="contains" />
                  <el-option label="正则" value="regex" />
                </el-select>
                <el-input v-model="a.config.expected" size="small" placeholder="期望值" style="margin-left: 4px" />
              </template>
              <!-- API 响应 -->
              <template v-else-if="a.type === 'api_response_v1'">
                <el-input v-model="a.config.url_pattern" size="small" placeholder="URL 匹配（包含 / 正则）" />
                <el-input v-model="a.config.json_path" size="small" placeholder="JSON Path（$.code）" style="margin-left: 4px" />
                <el-input v-model="a.config.expected" size="small" placeholder="期望值" style="margin-left: 4px" />
              </template>
              <!-- API 请求 -->
              <template v-else-if="a.type === 'api_request_v1'">
                <el-input v-model="a.config.url_pattern" size="small" placeholder="URL 匹配" />
                <el-input v-model="a.config.json_path" size="small" placeholder="JSON Path" style="margin-left: 4px" />
                <el-input v-model="a.config.expected" size="small" placeholder="期望值" style="margin-left: 4px" />
              </template>
              <span v-else style="color: #909399; font-size: 12px">自定义类型，请在 config 中编辑</span>
            </el-col>
            <el-col :span="4" style="text-align: right">
              <el-button size="small" type="danger" text @click="removeAssertion(idx)">
                <el-icon><Delete /></el-icon>
              </el-button>
            </el-col>
          </el-row>
        </div>
      </el-card>
    </div>

    <template #footer>
      <div class="dialog-footer">
        <el-button @click="onCancel">取消</el-button>
        <el-button type="primary" @click="onConfirm" :loading="saving">
          <el-icon><Check /></el-icon>
          确定步骤
        </el-button>
      </div>
    </template>
  </el-dialog>
</template>

<script>
import {
  InfoFilled,
  Setting,
  VideoPlay,
  CircleCheck,
  Plus,
  Delete,
  Check,
} from '@element-plus/icons-vue'
import base from '../../utils/base'

export default {
  name: 'StepConfirmDialog',
  components: {
    InfoFilled,
    Setting,
    VideoPlay,
    CircleCheck,
    Plus,
    Delete,
    Check,
  },
  props: {
    visible: { type: Boolean, default: false },
    step: { type: Object, default: () => null },
    sessionId: { type: [String, Number], default: '' },
    title: { type: String, default: '确认录制步骤' },
  },
  emits: ['update:visible', 'confirmed', 'cancelled'],
  data() {
    return {
      localStep: null,
      logicOp: 'and',
      assertionTypes: [],
      saving: false,
      replaying: false,
      replayResult: null,
    }
  },
  computed: {
    dialogVisible: {
      get() { return this.visible },
      set(v) { this.$emit('update:visible', v) },
    },
    isPositionClick() {
      return this.step && (this.step.type === 'click_by_position_v1' || this.step.type === 'right_click_v1')
    },
  },
  watch: {
    step: {
      immediate: true,
      handler(v) {
        if (v) {
          // 浅拷贝避免修改原对象
          this.localStep = JSON.parse(JSON.stringify({
            ...v,
            config: typeof v.config === 'object' ? v.config : this.safeParse(v.config),
            assertions: Array.isArray(v.assertions) ? v.assertions : this.safeParse(v.assertions) || [],
          }))
          if (!this.localStep.wait_after_ms) this.localStep.wait_after_ms = 200
          if (this.localStep.assertions.length > 1 && this.localStep.assertions[0]?.logic_op) {
            this.logicOp = this.localStep.assertions[0].logic_op
          }
        }
      },
    },
  },
  methods: {
    safeParse(s) {
      if (!s) return null
      try { return typeof s === 'string' ? JSON.parse(s) : s } catch { return null }
    },
    formatTime(ts) {
      if (!ts) return '-'
      const d = new Date(ts)
      if (isNaN(d.getTime())) return '-'
      return d.toLocaleString()
    },
    formatJson(obj) {
      if (!obj) return ''
      try { return JSON.stringify(obj, null, 2) } catch { return String(obj) }
    },
    loadAssertionTypes() {
      base.BasePost('/api/e2e/assertion/type/list', {}, (res) => {
        if (res && res.ErrCode === 0) {
          this.assertionTypes = res.Data?.items || []
        }
      })
    },
    addAssertion() {
      const a = {
        type: 'element_v1',
        version: '1.0',
        config: { selector: '' },
        logic_op: this.logicOp,
      }
      this.localStep.assertions.push(a)
    },
    removeAssertion(idx) {
      this.localStep.assertions.splice(idx, 1)
    },
    onAssertionTypeChange(a) {
      // 默认填充空 config
      a.config = {}
      a.version = '1.0'
    },
    onOpen() {
      this.replayResult = null
      this.loadAssertionTypes()
    },
    onClosed() {
      this.localStep = null
      this.replayResult = null
    },
    onCancel() {
      this.$emit('cancelled')
      this.dialogVisible = false
    },
    onConfirm() {
      // 把 logic_op 写入每个 assertion
      const stepCopy = JSON.parse(JSON.stringify(this.localStep))
      stepCopy.assertions = (stepCopy.assertions || []).map(a => ({
        ...a,
        logic_op: this.logicOp,
      }))
      this.$emit('confirmed', stepCopy)
    },
    replayStep() {
      if (!this.sessionId) {
        this.$message.warning('会话尚未创建，无法回放')
        return
      }
      this.replaying = true
      this.replayResult = null
      base.BasePost('/api/e2e/record/step/replay', {
        session_id: Number(this.sessionId),
        step_id: this.step.id,
      }, (res) => {
        this.replaying = false
        if (res && res.ErrCode === 0) {
          this.replayResult = res.Data
          if (res.Data?.success) {
            this.$message.success('回放成功')
          } else {
            this.$message.error('回放失败：' + (res.Data?.error || ''))
          }
        } else {
          this.$message.error(res?.ErrMsg || '回放请求失败')
        }
      })
    },
  },
}
</script>

<style scoped>
.step-confirm {
  display: flex;
  flex-direction: column;
  gap: 10px;
  max-height: 70vh;
  overflow: auto;
}
.block {
  border: 1px solid #ebeef5;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  font-weight: 600;
}
.config-preview {
  background: #f5f7fa;
  border-radius: 4px;
  padding: 8px;
  max-height: 160px;
  overflow: auto;
  font-size: 12px;
}
.config-detail {
  margin-top: 10px;
  padding-top: 10px;
  border-top: 1px dashed #dcdfe6;
}
.replay-result {
  margin-top: 10px;
}
.replay-log {
  background: #fafafa;
  border-radius: 4px;
  padding: 8px;
  margin-top: 6px;
  font-size: 12px;
  max-height: 200px;
  overflow: auto;
}
.assertion-row {
  padding: 8px 4px;
  border-bottom: 1px dashed #ebeef5;
}
.assertion-row:last-child {
  border-bottom: none;
}
.dialog-footer {
  text-align: right;
}
</style>
