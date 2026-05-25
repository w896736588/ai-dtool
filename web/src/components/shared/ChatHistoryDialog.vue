<template>
  <el-dialog
    :model-value="modelValue"
    :title="title"
    width="80vw"
    top="3vh"
    destroy-on-close
    @close="handleClose"
    @closed="$emit('closed')"
  >
    <div class="chat-combined-body" v-loading="loading">
      <div class="chat-combined-list">
        <div
          v-for="item in items"
          :key="item.id"
          :class="['chat-list-item', { 'chat-list-item--active': selectedId === item.id }]"
          @click="$emit('select', item)"
        >
          <div class="chat-list-item__name">
            <div class="chat-list-item__tags">
              <span class="chat-list-item__id">{{ item.id }}</span>
              <span v-if="item.agent_cli_name" class="chat-list-item__agent-name">{{ item.agent_cli_name }}</span>
            </div>
            <div class="chat-list-item__prompt" :title="item.prompt || '未命名'">
              {{ (item.prompt || '未命名').substring(0, 30) }}{{ (item.prompt || '').length > 30 ? '...' : '' }}
            </div>
          </div>
          <div class="chat-list-item__time">
            <span v-if="item.status === 'running' && runtimeDurationTextFn(item)" class="chat-list-item__time-running">
              {{ runtimeDurationTextFn(item) }}
            </span>
            <span v-else-if="item.duration_ms > 0">{{ formatDurationDisplay(item.duration_ms) }}</span>
            <span v-else>{{ formatCreatedAt(item.created_at) }}</span>
            <span v-if="itemMsgCountFn(item) > 0" class="chat-list-item__msg-count">{{ itemMsgCountFn(item) }}条</span>
          </div>
          <span :class="['chat-list-item__status', 'chat-list-item__status--' + (item.status || '')]">
            <span v-if="item.status === 'running'" class="chat-list-item__running-dot"></span>
            <span v-else-if="item.status === 'error'" class="chat-list-item__error-icon">!</span>
            {{ statusTextFn(item.status) }} {{ formatCreatedAtFn(item.created_at) }}
          </span>
        </div>
        <div v-if="items.length === 0 && !loading" class="chat-combined-list__empty">{{ listEmptyText }}</div>
      </div>

      <div class="chat-combined-detail">
        <div v-if="!selectedId" class="chat-combined-detail__placeholder">{{ detailPlaceholderText }}</div>
        <template v-else>
          <div class="chat-detail-task-name">{{ detailTitle }}</div>
          <div v-if="modelName || localDir || thinkingIntensity" class="chat-detail-meta">
            <span v-if="modelName">模型: {{ modelName }}</span>
            <span v-if="modelName && localDir"> | </span>
            <span v-if="localDir">目录: {{ localDir }}</span>
            <span v-if="thinkingIntensity && (modelName || localDir)"> | </span>
            <span v-if="thinkingIntensity">思考强度: {{ thinkingIntensity }}</span>
          </div>

          <div ref="detailContainer" class="chat-detail-container" @scroll="$emit('scroll')">
            <div v-if="detailMessages.length === 0 && detailStatus === 'running'" class="chat-detail-empty-running">
              <div>{{ runningText }}</div>
            </div>
            <div v-for="(msg, idx) in detailMessages" :key="idx" class="chat-message-item">
              <div v-if="msg.type === 'system_init'" class="chat-message-system-init">
                {{ msg.text }} | model: {{ msg.model || modelName || '-' }}
              </div>
              <div v-else-if="msg.type === 'system_command'" class="chat-message-system-command">
                <div class="chat-message-command-card">
                  <div class="chat-message-command-header">
                    <span class="chat-message-command-type">{{ formatCliType(msg.cliType) }}</span>
                    <span v-if="isLongText(msg.cmdLine || msg.text, 20)" class="chat-message-toggle" @click="msg.collapsed = !msg.collapsed">
                      {{ msg.collapsed ? '展开 ▼' : '收起 ▲' }}
                    </span>
                  </div>
                  <div v-if="msg.cmdLine" class="markdown-body chat-markdown-body" v-html="renderMarkdown('> ' + (msg.collapsed ? truncateCmdPrompt(msg.cmdLine, 15) : msg.cmdLine))"></div>
                  <div v-else class="chat-message-plain" :style="{ maxHeight: msg.collapsed ? '20em' : 'none', overflow: msg.collapsed ? 'hidden' : 'visible' }">{{ msg.text }}</div>
                  <div v-if="msg.cmdLine" class="chat-message-command-text" :style="{ maxHeight: msg.collapsed ? '15em' : 'none', overflow: msg.collapsed ? 'hidden' : 'visible' }">{{ msg.text }}</div>
                </div>
              </div>
              <div v-else-if="msg.type === 'system_hook'" class="chat-message-system-hook">
                <span class="chat-message-toggle" @click="msg.collapsed = !msg.collapsed">{{ msg.collapsed ? '▶' : '▼' }} {{ msg.text }}</span>
                <div v-if="!msg.collapsed && (msg.stderr || msg.output)" class="chat-message-hook-output">
                  <div v-if="msg.stderr" class="chat-message-hook-stderr">{{ msg.stderr }}</div>
                  <div v-if="msg.output" class="chat-message-hook-output-text">{{ msg.output }}</div>
                </div>
              </div>
              <div v-else-if="msg.type === 'system'" class="chat-message-system">{{ msg.text }}</div>
              <div v-else-if="msg.type === 'system_status'" class="chat-message-system-status">
                <span :class="{ 'chat-message-system-status--active': msg.status === 'requesting' }">{{ msg.text }}</span>
              </div>
              <div v-else-if="msg.type === 'system_task'" class="chat-message-system-task">
                <span v-if="(msg.status === 'started' || msg.status === 'running') && detailStatus === 'running'" class="chat-detail-status-spinner"></span>
                <span :class="{ 'chat-message-system-task--active': msg.status === 'started' }">🔧 {{ msg.description }}</span>
                <span class="chat-message-system-task__state">{{ msg.status === 'started' ? '启动' : msg.status }}</span>
              </div>
              <div v-else-if="msg.type === 'assistant'">
                <div v-if="msg.thinking" class="chat-message-thinking-wrap">
                  <div class="chat-message-thinking-head">
                    <span v-if="isCurrentThinkingFn(msg)" class="chat-detail-status-spinner"></span>
                    <span v-if="isCurrentThinkingFn(msg)" class="chat-message-thinking-running">思考过程 持续{{ thinkingStreamElapsed }}s</span>
                    <span v-else class="chat-message-thinking-text">思考过程{{ msg._thinkingTiming && msg._thinkingTiming.durationMs ? ' (' + (msg._thinkingTiming.durationMs / 1000).toFixed(1) + 's)' : '' }}</span>
                    <span class="chat-message-toggle" @click="toggleThinking(msg)">{{ msg._thinkingCollapsed ? '▶' : '▼' }}</span>
                  </div>
                  <div v-if="!msg._thinkingCollapsed" class="thinking-blockquote">{{ msg.thinking }}</div>
                </div>
                <div v-for="(block, bi) in msg.content" :key="bi">
                  <div v-if="block.type === 'text'" class="markdown-body chat-markdown-body" v-html="renderMarkdown(block.text)"></div>
                  <div v-else-if="block.type === 'tool_use'" class="chat-tool-card">
                    <div class="chat-tool-card__head">
                      <span v-if="!block._result && detailStatus === 'running'" class="chat-detail-status-spinner"></span>
                      <span class="chat-tool-card__name">🔧 {{ block.name }}</span>
                      <span v-if="block.displayInput" class="chat-tool-card__input">{{ block.displayInput }}</span>
                    </div>
                    <div v-if="block._tasks" class="chat-task-list">
                      <div v-for="(task, ti) in block._tasks" :key="ti" class="chat-task-list__item">
                        <span :class="['chat-task-list__icon', task.status]">
                          {{ task.status === 'completed' ? '✅' : task.status === 'in_progress' ? '🔄' : '⬜' }}
                        </span>
                        <span :class="{ 'chat-task-list__done': task.status === 'completed' }">{{ task.content }}</span>
                        <span v-if="task.activeForm && task.status === 'in_progress'" class="chat-task-list__form">{{ task.activeForm }}</span>
                      </div>
                    </div>
                    <div v-if="block._askQuestions" class="chat-ask-questions">
                      <div v-for="(q, qi) in block._askQuestions" :key="qi" class="chat-ask-question-item">
                        <div class="chat-ask-question-title">{{ q.question }}</div>
                        <div class="chat-ask-question-meta">类型: {{ q.header || '选择' }}{{ q.multiSelect ? ' (多选)' : '' }}</div>
                        <div v-for="(opt, oi) in q.options" :key="oi" class="chat-ask-option">
                          <span class="chat-ask-option__mark">{{ q.multiSelect ? '☐' : '○' }}</span>
                          <div>
                            <div>{{ opt.label }}</div>
                            <div v-if="opt.description" class="chat-ask-option__desc">{{ opt.description }}</div>
                          </div>
                        </div>
                      </div>
                    </div>
                    <div v-if="!block.displayInput && !block._tasks && !block._askQuestions" class="chat-tool-card__toggle" @click="block._inputExpanded = !block._inputExpanded">
                      {{ block._inputExpanded ? '▼' : '▶' }} 参数
                    </div>
                    <pre v-if="!block.displayInput && !block._tasks && !block._askQuestions && block._inputExpanded" class="chat-tool-card__pre">{{ block.input }}</pre>
                    <div v-if="block._result" class="chat-tool-card__result">
                      <span class="chat-message-toggle" @click="block._result.collapsed = !block._result.collapsed">{{ block._result.collapsed ? '▶' : '▼' }} 工具执行结果</span>
                      <div v-if="!block._result.collapsed && block._result._tasks" class="chat-task-list chat-task-list--result">
                        <div v-for="(task, ti) in block._result._tasks" :key="ti" class="chat-task-list__item">
                          <span :class="['chat-task-list__icon', task.status]">
                            {{ task.status === 'completed' ? '✅' : task.status === 'in_progress' ? '🔄' : '⬜' }}
                          </span>
                          <span :class="{ 'chat-task-list__done': task.status === 'completed' }">{{ task.content }}</span>
                          <span v-if="task.activeForm && task.status === 'in_progress'" class="chat-task-list__form">{{ task.activeForm }}</span>
                        </div>
                      </div>
                      <pre v-if="!block._result.collapsed" class="chat-tool-card__result-pre">{{ block._result.text }}</pre>
                    </div>
                  </div>
                </div>
                <div v-if="msg.usage" class="chat-message-usage">
                  input: {{ msg.usage.input_tokens }} | output: {{ msg.usage.output_tokens }}
                </div>
              </div>
              <div v-else-if="msg.type === 'tool_use'" class="chat-tool-card">
                <div class="chat-tool-card__head">
                  <span v-if="!msg._result && detailStatus === 'running'" class="chat-detail-status-spinner"></span>
                  <span class="chat-tool-card__name">🔧 {{ msg.name }}</span>
                  <span v-if="msg.displayInput" class="chat-tool-card__input">{{ msg.displayInput }}</span>
                </div>
                <div v-if="msg._tasks" class="chat-task-list">
                  <div v-for="(task, ti) in msg._tasks" :key="ti" class="chat-task-list__item">
                    <span :class="['chat-task-list__icon', task.status]">
                      {{ task.status === 'completed' ? '✅' : task.status === 'in_progress' ? '🔄' : '⬜' }}
                    </span>
                    <span :class="{ 'chat-task-list__done': task.status === 'completed' }">{{ task.content }}</span>
                    <span v-if="task.activeForm && task.status === 'in_progress'" class="chat-task-list__form">{{ task.activeForm }}</span>
                  </div>
                </div>
                <div v-if="msg._askQuestions" class="chat-ask-questions">
                  <div v-for="(q, qi) in msg._askQuestions" :key="qi" class="chat-ask-question-item">
                    <div class="chat-ask-question-title">{{ q.question }}</div>
                    <div class="chat-ask-question-meta">类型: {{ q.header || '选择' }}{{ q.multiSelect ? ' (多选)' : '' }}</div>
                    <div v-for="(opt, oi) in q.options" :key="oi" class="chat-ask-option">
                      <span class="chat-ask-option__mark">{{ q.multiSelect ? '☐' : '○' }}</span>
                      <div>
                        <div>{{ opt.label }}</div>
                        <div v-if="opt.description" class="chat-ask-option__desc">{{ opt.description }}</div>
                      </div>
                    </div>
                  </div>
                </div>
                <div v-if="!msg.displayInput && !msg._tasks && !msg._askQuestions" class="chat-tool-card__toggle" @click="msg._inputExpanded = !msg._inputExpanded">
                  {{ msg._inputExpanded ? '▼' : '▶' }} 参数
                </div>
                <pre v-if="!msg.displayInput && !msg._tasks && !msg._askQuestions && msg._inputExpanded" class="chat-tool-card__pre">{{ msg.input }}</pre>
                <div v-if="msg._result" class="chat-tool-card__result">
                  <span class="chat-message-toggle" @click="msg._result.collapsed = !msg._result.collapsed">{{ msg._result.collapsed ? '▶' : '▼' }} 工具执行结果</span>
                  <div v-if="!msg._result.collapsed && msg._result._tasks" class="chat-task-list chat-task-list--result">
                    <div v-for="(task, ti) in msg._result._tasks" :key="ti" class="chat-task-list__item">
                      <span :class="['chat-task-list__icon', task.status]">
                        {{ task.status === 'completed' ? '✅' : task.status === 'in_progress' ? '🔄' : '⬜' }}
                      </span>
                      <span :class="{ 'chat-task-list__done': task.status === 'completed' }">{{ task.content }}</span>
                      <span v-if="task.activeForm && task.status === 'in_progress'" class="chat-task-list__form">{{ task.activeForm }}</span>
                    </div>
                  </div>
                  <pre v-if="!msg._result.collapsed" class="chat-tool-card__result-pre">{{ msg._result.text }}</pre>
                </div>
              </div>
              <div v-else-if="msg.type === 'tool_result'" class="chat-message-tool-result">
                <span class="chat-message-toggle" @click="msg.collapsed = !msg.collapsed">{{ msg.collapsed ? '▶' : '▼' }} 工具执行结果</span>
                <div v-if="!msg.collapsed && msg._tasks" class="chat-task-list chat-task-list--result">
                  <div v-for="(task, ti) in msg._tasks" :key="ti" class="chat-task-list__item">
                    <span :class="['chat-task-list__icon', task.status]">
                      {{ task.status === 'completed' ? '✅' : task.status === 'in_progress' ? '🔄' : '⬜' }}
                    </span>
                    <span :class="{ 'chat-task-list__done': task.status === 'completed' }">{{ task.content }}</span>
                    <span v-if="task.activeForm && task.status === 'in_progress'" class="chat-task-list__form">{{ task.activeForm }}</span>
                  </div>
                </div>
                <pre v-if="!msg.collapsed" class="chat-tool-card__result-pre">{{ msg.text }}</pre>
              </div>
              <div v-else-if="msg.type === 'assistant_text'" class="markdown-body chat-markdown-body" v-html="renderMarkdown(msg.text)"></div>
              <div v-else-if="msg.type === 'assistant_thinking'" class="chat-message-thinking">
                <div class="chat-message-thinking-head">
                  <span>思考过程{{ msg._thinkingTiming && msg._thinkingTiming.durationMs ? ' (' + (msg._thinkingTiming.durationMs / 1000).toFixed(1) + 's)' : '' }}</span>
                  <span class="chat-message-toggle" @click="toggleThinking(msg)">{{ msg._thinkingCollapsed ? '▶' : '▼' }}</span>
                </div>
                <div v-if="!msg._thinkingCollapsed" class="thinking-blockquote">{{ msg.text }}</div>
              </div>
              <div v-else-if="msg.type === 'result'" class="chat-result-card">
                <div class="chat-result-header">
                  <span :class="['chat-result-header__state', msg.isError ? 'chat-result-header__state--error' : 'chat-result-header__state--success']">
                    {{ msg.isError ? '✕ 执行失败' : '✓ 执行完成' }}
                  </span>
                  <span class="chat-result-header-item">耗时 {{ (msg.durationMs / 1000).toFixed(1) }}s</span>
                  <span v-if="msg.durationApiMs" class="chat-result-header-item">API {{ (msg.durationApiMs / 1000).toFixed(1) }}s</span>
                  <span class="chat-result-header-item">{{ msg.numTurns }} 轮对话</span>
                  <span v-if="msg.totalCostUsd != null" class="chat-result-header-item chat-result-header-item--cost">${{ msg.totalCostUsd.toFixed(4) }}</span>
                  <span v-if="msg.stopReason" class="chat-result-header-item">{{ stopReasonLabel(msg.stopReason) }}</span>
                </div>
                <div v-if="msg.usage" class="chat-result-section">
                  <div class="chat-result-section-title">Token 用量</div>
                  <div class="chat-result-tokens">
                    <span>输入 {{ formatNum(msg.usage.input_tokens) }}</span>
                    <span>输出 {{ formatNum(msg.usage.output_tokens) }}</span>
                    <span v-if="msg.usage.cache_read_input_tokens">缓存读取 {{ formatNum(msg.usage.cache_read_input_tokens) }}</span>
                    <span v-if="msg.usage.cache_creation_input_tokens">缓存创建 {{ formatNum(msg.usage.cache_creation_input_tokens) }}</span>
                  </div>
                </div>
                <div v-if="msg.modelUsage && msg.modelUsage.length" class="chat-result-section">
                  <div class="chat-result-section-title">模型用量</div>
                  <div v-for="mu in msg.modelUsage" :key="mu.name" class="chat-result-model-row">
                    <span class="chat-result-model-name">{{ mu.name }}</span>
                    <span>输入 {{ formatNum(mu.inputTokens) }}</span>
                    <span>输出 {{ formatNum(mu.outputTokens) }}</span>
                    <span v-if="mu.costUSD">${{ mu.costUSD.toFixed(4) }}</span>
                  </div>
                </div>
                <div v-if="msg.permissionDenials && msg.permissionDenials.length" class="chat-result-section">
                  <div class="chat-result-section-title chat-result-section-title--warn">权限询问 ({{ msg.permissionDenials.length }})</div>
                  <div v-for="(pd, pdi) in msg.permissionDenials" :key="pdi" class="chat-result-permission-item">
                    <span>{{ pd.tool_name }}</span>
                    <span v-if="pd.tool_use_id" class="chat-result-permission-item__id">{{ pd.tool_use_id.slice(0, 8) }}...</span>
                  </div>
                </div>
                <div v-if="msg.resultText" class="chat-result-section">
                  <div class="chat-result-section-title">结果</div>
                  <pre class="chat-result-text">{{ msg.resultText }}</pre>
                </div>
              </div>
              <div v-else-if="msg.type === 'chat_completed' && detailStatus === 'completed'" class="chat-message-completed">
                {{ msg.text }}
              </div>
              <div v-else-if="msg.type === 'raw_text'" class="chat-message-raw">{{ msg.text }}</div>
              <div v-else-if="msg.type === 'parse_error'" class="chat-message-parse-error">
                <div class="chat-message-parse-error__title">解析错误</div>
                <div v-if="msg.error" class="chat-message-parse-error__msg">{{ msg.error }}</div>
                <pre class="chat-message-parse-error__text">{{ msg.text }}</pre>
              </div>
              <div v-else-if="msg.type === 'error'" class="chat-message-error">
                <span class="chat-message-error__label">错误:</span>
                <span>{{ msg.text }}</span>
              </div>
            </div>
          </div>
          <div class="chat-detail-scroll-btn" :class="{ 'chat-detail-scroll-btn--visible': scrollButtonVisible }" @click="$emit('scroll-to-bottom')">↓</div>
          <slot name="before-input" />
          <div class="chat-detail-input-row">
            <div class="chat-detail-textarea-wrapper">
              <el-input
                :model-value="continueInput"
                type="textarea"
                :rows="3"
                :placeholder="continuePlaceholder"
                :disabled="detailStatus === 'running'"
                class="chat-detail-textarea"
                @update:model-value="$emit('update:continueInput', $event)"
                @keydown.enter.exact.prevent="detailStatus !== 'running' && $emit('continue')"
              />
              <div class="chat-detail-actions">
                <div v-if="thinkingIntensity || modelName" class="chat-detail-info-bar">
                  <span v-if="thinkingIntensity">思考强度: {{ thinkingIntensity }}</span>
                  <span v-if="thinkingIntensity && modelName"> | </span>
                  <span v-if="modelName">智能体: {{ modelName }}</span>
                </div>
                <el-button v-if="detailStatus === 'running'" type="danger" size="small" @click="$emit('stop')">停止</el-button>
                <el-button v-else type="primary" size="small" :loading="continueLoading" @click="$emit('continue')">发送</el-button>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>
  </el-dialog>
</template>

<script>
export default {
  name: 'ChatHistoryDialog',
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    title: {
      type: String,
      default: '执行历史',
    },
    loading: {
      type: Boolean,
      default: false,
    },
    items: {
      type: Array,
      default: () => [],
    },
    selectedId: {
      type: [Number, String],
      default: 0,
    },
    listEmptyText: {
      type: String,
      default: '暂无执行记录',
    },
    detailPlaceholderText: {
      type: String,
      default: '请选择一条执行记录',
    },
    detailTitle: {
      type: String,
      default: '',
    },
    modelName: {
      type: String,
      default: '',
    },
    localDir: {
      type: String,
      default: '',
    },
    thinkingIntensity: {
      type: String,
      default: '',
    },
    detailStatus: {
      type: String,
      default: '',
    },
    detailMessages: {
      type: Array,
      default: () => [],
    },
    continueInput: {
      type: String,
      default: '',
    },
    continuePlaceholder: {
      type: String,
      default: '输入新消息继续对话...',
    },
    continueLoading: {
      type: Boolean,
      default: false,
    },
    scrollButtonVisible: {
      type: Boolean,
      default: false,
    },
    runningText: {
      type: String,
      default: '等待执行响应...',
    },
    thinkingStreamElapsed: {
      type: Number,
      default: 0,
    },
    itemMsgCountFn: {
      type: Function,
      default: (item) => item.line_count || 0,
    },
    statusTextFn: {
      type: Function,
      default: (status) => {
        const map = { running: '执行中', completed: '已完成', error: '异常终止', interrupted: '中断' }
        return map[status] || status || '-'
      },
    },
    runtimeDurationTextFn: {
      type: Function,
      default: () => '',
    },
    formatDurationDisplayFn: {
      type: Function,
      default: (durationMs) => {
        const ms = Number(durationMs || 0)
        if (ms <= 0) return ''
        const totalSeconds = Math.floor(ms / 1000)
        const minutes = Math.floor(totalSeconds / 60)
        const seconds = totalSeconds % 60
        if (minutes > 0) return minutes + 'm' + seconds + 's'
        return seconds + 's'
      },
    },
    formatCreatedAtFn: {
      type: Function,
      default: (createdAt) => {
        if (!createdAt) return '-'
        const d = new Date(String(createdAt).replace(/-/g, '/'))
        if (Number.isNaN(d.getTime())) return ''
        const pad = (n) => String(n).padStart(2, '0')
        return d.getFullYear() + '/' + pad(d.getMonth() + 1) + '/' + pad(d.getDate()) + ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds())
      },
    },
    renderMarkdownFn: {
      type: Function,
      default: (text) => text || '',
    },
    isCurrentThinkingFn: {
      type: Function,
      default: () => false,
    },
    formatCliTypeFn: {
      type: Function,
      default: (cliType) => {
        if (!cliType) return '提示词'
        return cliType.charAt(0).toUpperCase() + cliType.slice(1)
      },
    },
    isLongTextFn: {
      type: Function,
      default: () => false,
    },
    truncateCmdPromptFn: {
      type: Function,
      default: (cmdLine) => cmdLine || '',
    },
    stopReasonLabelFn: {
      type: Function,
      default: (reason) => reason || '',
    },
    formatNumFn: {
      type: Function,
      default: (num) => (num == null ? '0' : Number(num).toLocaleString()),
    },
  },
  emits: ['update:modelValue', 'select', 'update:continueInput', 'continue', 'stop', 'scroll', 'scroll-to-bottom', 'closed'],
  methods: {
    handleClose() {
      this.$emit('update:modelValue', false)
    },
    // 获取详情滚动容器 / Get the detail scroll container.
    getDetailContainer() {
      return this.$refs.detailContainer || null
    },
    // 判断详情容器是否接近底部 / Check whether the detail container is near the bottom.
    isDetailNearBottom(threshold = 30) {
      const el = this.getDetailContainer()
      if (!el) return true
      return el.scrollHeight - el.scrollTop - el.clientHeight < threshold
    },
    // 滚动详情到底部 / Scroll the detail view to the bottom.
    scrollDetailToBottom(behavior = 'auto') {
      const el = this.getDetailContainer()
      if (el) {
        el.scrollTo({ top: el.scrollHeight, behavior })
      }
    },
    toggleThinking(msg) {
      msg._thinkingCollapsed = !msg._thinkingCollapsed
      msg._thinkingManuallyToggled = true
    },
    formatDurationDisplay(durationMs) {
      return this.formatDurationDisplayFn(durationMs)
    },
    formatCreatedAt(createdAt) {
      return this.formatCreatedAtFn(createdAt)
    },
    renderMarkdown(text) {
      return this.renderMarkdownFn(text)
    },
    isCurrentThinking(msg) {
      return this.isCurrentThinkingFn(msg)
    },
    formatCliType(cliType) {
      return this.formatCliTypeFn(cliType)
    },
    isLongText(text, maxBytes) {
      return this.isLongTextFn(text, maxBytes)
    },
    truncateCmdPrompt(cmdLine, maxLen) {
      return this.truncateCmdPromptFn(cmdLine, maxLen)
    },
    stopReasonLabel(reason) {
      return this.stopReasonLabelFn(reason)
    },
    formatNum(num) {
      return this.formatNumFn(num)
    },
  },
}
</script>

<style scoped src="@/css/components/shared/ChatHistory.css"></style>
