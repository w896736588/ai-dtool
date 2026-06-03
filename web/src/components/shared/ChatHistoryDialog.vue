<template>
  <el-dialog
    :model-value="modelValue"
    width="80vw"
    top="3vh"
    destroy-on-close
    @close="handleClose"
    @closed="$emit('closed')"
  >
    <template #header>
      <div class="chat-history-dialog__header">
        <span class="chat-history-dialog__title">{{ title }}</span>
        <span
          v-if="workspaceSummaryText"
          class="chat-history-dialog__workspace-summary"
          :title="workspaceSummaryText"
        >
          {{ workspaceSummaryText }}
        </span>
      </div>
    </template>
    <div class="chat-combined-body" v-loading="loading">
      <div class="chat-combined-list">
        <template v-for="group in groupedItems" :key="group.dir">
          <div class="chat-list-group-header" @click="toggleGroup(group.dir)">
            <span class="chat-list-group-header__arrow">{{ isGroupCollapsed(group.dir) ? '▶' : '▼' }}</span>
            <span class="chat-list-group-header__dir" :title="group.dir">{{ group.dirLabel }}</span>
            <span v-if="group.unreadCount > 0" class="chat-list-group-header__unread">{{ group.unreadCount }}</span>
            <span class="chat-list-group-header__count">{{ group.items.length }}</span>
          </div>
          <template v-if="!isGroupCollapsed(group.dir)">
            <div
              v-for="item in group.items"
              :key="item.id"
              :class="[
                'chat-list-item',
                {
                  'chat-list-item--active': selectedId === item.id,
                  'chat-list-item--stale': isItemStale(item),
                },
              ]"
              @click="$emit('select', item)"
            >
              <div class="chat-list-item__name">
                <div class="chat-list-item__tags">
                  <span v-if="item.is_read === false" class="chat-list-item__unread-dot"></span>
                  <span class="chat-list-item__id">{{ item.id }}</span>
                  <span v-if="getItemAgentName(item)" class="chat-list-item__agent-name">智能体: {{ getItemAgentName(item) }}</span>
                  <span v-if="item._killed_pid" class="chat-list-item__killed-pid">杀进程:{{ item._killed_pid }}</span>
                </div>
                <div class="chat-list-item__prompt" :title="item.prompt || '未命名'">
                  {{ (item.prompt || '未命名').substring(0, 30) }}{{ (item.prompt || '').length > 30 ? '...' : '' }}
                </div>
                <div v-if="getItemModelName(item)" class="chat-list-item__meta">
                  <span v-if="getItemModelName(item)" class="chat-list-item__meta-tag chat-list-item__meta-tag--model">
                    模型: {{ getItemModelName(item) }}
                  </span>
                </div>
                <div v-if="getItemTerminalReasonText(item)" class="chat-list-item__terminal-reason" :title="getItemTerminalReasonText(item)">
                  终止原因: {{ getItemTerminalReasonText(item) }}
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
          </template>
        </template>
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
                  <pre v-if="msg.cmdLine" class="chat-message-command-block"><code>{{ msg.collapsed ? truncateCmdPrompt(msg.cmdLine, 15) : msg.cmdLine }}</code></pre>
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
                    </div>
                    <pre v-if="block.displayInput" class="chat-tool-card__command"><code>{{ block.displayInput }}</code></pre>
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
                </div>
                <pre v-if="msg.displayInput" class="chat-tool-card__command"><code>{{ msg.displayInput }}</code></pre>
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
                  <div
                    v-for="(line, lineIndex) in buildResultSummaryLines(msg)"
                    :key="lineIndex"
                    :class="[
                      'chat-result-header-item',
                      lineIndex === 0 ? (msg.isError ? 'chat-result-header__state chat-result-header__state--error' : 'chat-result-header__state chat-result-header__state--success') : ''
                    ]"
                  >
                    {{ line }}
                  </div>
                </div>
                <div v-if="getPrimaryUsageRow(msg)" class="chat-result-section">
                  <div class="chat-result-section-title">Token 用量</div>
                  <div class="chat-result-model-row">
                    <span class="chat-result-model-name">{{ getPrimaryUsageRow(msg).name }}</span>
                    <span>输入 {{ formatNum(getPrimaryUsageRow(msg).inputTokens) }}</span>
                    <span>输出 {{ formatNum(getPrimaryUsageRow(msg).outputTokens) }}</span>
                    <span v-if="getPrimaryUsageRow(msg).cacheReadInputTokens">缓存读取 {{ formatNum(getPrimaryUsageRow(msg).cacheReadInputTokens) }}</span>
                  </div>
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
                <div v-if="showDetailInfoBar" class="chat-detail-info-bar">
                  <div class="chat-detail-info-bar__left">
                    <span v-if="thinkingIntensity">思考强度: {{ thinkingIntensity }}</span>
                    <span v-if="thinkingIntensity && agentName"> | </span>
                    <span v-if="agentName">智能体: {{ agentName }}</span>
                  </div>
                  <div v-if="lastUsageSummary" class="chat-detail-info-bar__right">
                    <span>最后一次输入: {{ formatCompactToken(lastUsageSummary.inputTokens) }}</span>
                    <span>命中缓存: {{ formatCompactToken(lastUsageSummary.cacheReadInputTokens) }}</span>
                  </div>
                </div>
                <el-button v-if="detailStatus === 'running'" type="danger" size="small" @click="$emit('stop')">停止</el-button>
                <template v-else>
                  <el-button
                    v-if="showNewChatButton"
                    size="small"
                    :disabled="continueDisabled"
                    :loading="continueLoading"
                    @click="$emit('new-chat')"
                  >
                    新对话
                  </el-button>
                  <el-button
                    type="primary"
                    size="small"
                    :disabled="continueDisabled"
                    :loading="continueLoading"
                    @click="$emit('continue')"
                  >
                    继续对话
                  </el-button>
                </template>
              </div>
            </div>
          </div>
        </template>
      </div>
    </div>
  </el-dialog>
</template>

<script>
import resultSummaryUtils from '@/utils/chat_result_summary.cjs'

// STORAGE_KEY_GROUP_COLLAPSED localStorage 键名，用于持久化分组展开/关闭状态。
const STORAGE_KEY_GROUP_COLLAPSED = 'chat_history_group_collapsed'
// DEFAULT_GROUP_LABEL 当 item 没有 local_dir 时使用的默认分组标签。
const DEFAULT_GROUP_LABEL = '未指定目录'

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
    agentName: {
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
    detailCliType: {
      type: String,
      default: 'claude',
    },
    detailMessages: {
      type: Array,
      default: () => [],
    },
    lastUsageSummaryData: {
      type: Object,
      default: null,
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
    continueDisabled: {
      type: Boolean,
      default: false,
    },
    showNewChatButton: {
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
  emits: ['update:modelValue', 'select', 'update:continueInput', 'continue', 'new-chat', 'stop', 'scroll', 'scroll-to-bottom', 'closed'],
  data() {
    return {
      // collapsedGroups 记录当前各分组的展开/关闭状态，key 为 local_dir 值。
      collapsedGroups: {},
      // hasAppliedInitialSelectedGroupExpand 确保仅在弹窗打开后的初始化阶段自动展开一次所属工作空间。
      hasAppliedInitialSelectedGroupExpand: false,
    }
  },
  created() {
    this.loadGroupCollapsedState()
  },
  watch: {
    modelValue(visible) {
      if (visible) {
        this.hasAppliedInitialSelectedGroupExpand = false
        this.$nextTick(() => {
          this.ensureInitialSelectedGroupExpanded()
        })
      }
    },
    selectedId() {
      this.$nextTick(() => {
        this.ensureInitialSelectedGroupExpanded()
      })
    },
    items: {
      handler() {
        this.$nextTick(() => {
          this.ensureInitialSelectedGroupExpanded()
        })
      },
      deep: true,
    },
  },
  computed: {
    // groupedItems 将 items 按 local_dir 分组，保持每组内原始顺序。 // Groups items by local_dir while preserving original order within each group.
    groupedItems() {
      const groups = []
      const dirMap = new Map()
      const itemList = Array.isArray(this.items) ? this.items : []
      itemList.forEach((item) => {
        const dir = this.getItemWorkspacePath(item)
        const dirKey = dir || ''
        if (!dirMap.has(dirKey)) {
          const group = { dir: dirKey, dirLabel: this.getCompactDirLabel(dir), items: [], unreadCount: 0 }
          dirMap.set(dirKey, group)
          groups.push(group)
        }
        const group = dirMap.get(dirKey)
        group.items.push(item)
        if (this.isItemUnread(item)) {
          group.unreadCount += 1
        }
      })
      return groups
    },
    // workspaceSummaryText 聚合顶部展示的工作空间简称列表，按首次出现顺序去重。
    workspaceSummaryText() {
      const labelSet = new Set()
      const labels = []
      const itemList = Array.isArray(this.items) ? this.items : []
      itemList.forEach((item) => {
        const label = this.getItemWorkspaceLabel(item)
        if (!label || labelSet.has(label)) return
        labelSet.add(label)
        labels.push(label)
      })
      return labels.join(', ')
    },
    // showDetailInfoBar 控制底部信息栏展示。 // Controls rendering of the footer info bar.
    showDetailInfoBar() {
      return Boolean(this.thinkingIntensity || this.agentName || this.lastUsageSummary)
    },
    // lastUsageSummary 提取当前会话最后一次 token 统计。 // Extracts the latest token usage stats from the current chat detail.
    lastUsageSummary() {
      const messages = Array.isArray(this.detailMessages) ? this.detailMessages : []
      for (let index = messages.length - 1; index >= 0; index -= 1) {
        const usageRow = this.getPrimaryUsageRow(messages[index])
        if (!usageRow) continue
        const inputTokens = Number(usageRow.inputTokens || 0)
        const cacheReadInputTokens = Number(usageRow.cacheReadInputTokens || 0)
        if (inputTokens > 0 || cacheReadInputTokens > 0) {
          return { inputTokens, cacheReadInputTokens }
        }
      }
      const fallbackInputTokens = Number(this.lastUsageSummaryData?.inputTokens || 0)
      const fallbackCacheReadInputTokens = Number(this.lastUsageSummaryData?.cacheReadInputTokens || 0)
      if (fallbackInputTokens > 0 || fallbackCacheReadInputTokens > 0) {
        return {
          inputTokens: fallbackInputTokens,
          cacheReadInputTokens: fallbackCacheReadInputTokens,
        }
      }
      return null
    },
  },
  methods: {
    handleClose() {
      this.$emit('update:modelValue', false)
    },
    // loadGroupCollapsedState 从 localStorage 加载分组展开/关闭状态。 // Loads group collapsed state from localStorage on mount.
    loadGroupCollapsedState() {
      try {
        const raw = localStorage.getItem(STORAGE_KEY_GROUP_COLLAPSED)
        if (raw) {
          const parsed = JSON.parse(raw)
          if (parsed && typeof parsed === 'object') {
            this.collapsedGroups = parsed
          }
        }
      } catch (_e) {
        // localStorage 不可用或数据损坏时忽略
      }
    },
    // saveGroupCollapsedState 将分组展开/关闭状态持久化到 localStorage。 // Persists group collapsed state to localStorage.
    saveGroupCollapsedState() {
      try {
        localStorage.setItem(STORAGE_KEY_GROUP_COLLAPSED, JSON.stringify(this.collapsedGroups))
      } catch (_e) {
        // localStorage 不可用时忽略
      }
    },
    // isGroupCollapsed 判断指定分组是否处于折叠状态。 // Returns whether the given group directory is collapsed.
    isGroupCollapsed(dir) {
      return Boolean(this.collapsedGroups[dir])
    },
    // toggleGroup 切换指定分组的展开/关闭状态并持久化。 // Toggles collapsed state for a group directory and persists it.
    toggleGroup(dir) {
      if (this.collapsedGroups[dir]) {
        delete this.collapsedGroups[dir]
      } else {
        this.collapsedGroups[dir] = true
      }
      // 触发响应式更新
      this.collapsedGroups = { ...this.collapsedGroups }
      this.saveGroupCollapsedState()
    },
    // ensureInitialSelectedGroupExpanded 仅在弹窗打开后的初始化阶段自动展开当前选中记录所属工作空间。 // Expands the selected chat's workspace group only during dialog-open initialization.
    ensureInitialSelectedGroupExpanded() {
      if (!this.modelValue) return
      if (this.hasAppliedInitialSelectedGroupExpand) return
      const itemList = Array.isArray(this.items) ? this.items : []
      const activeId = Number(this.selectedId || 0)
      if (activeId <= 0 || itemList.length === 0) return
      const selectedItem = itemList.find(item => Number(item?.id || 0) === activeId)
      if (!selectedItem) return
      this.hasAppliedInitialSelectedGroupExpand = true
      const dir = this.getItemWorkspacePath(selectedItem)
      const dirKey = dir || ''
      if (!this.collapsedGroups[dirKey]) return
      delete this.collapsedGroups[dirKey]
      this.collapsedGroups = { ...this.collapsedGroups }
      this.saveGroupCollapsedState()
    },
    // getItemAgentName 统一提取左侧列表项的智能体名称，兼容不同页面返回字段。 // Extracts the agent name for list rows across different history sources.
    getItemAgentName(item) {
      const agentName = String(item?.agent_cli_name || item?.agent_name || item?.agentName || '').trim()
      return agentName
    },
    // getItemModelName 统一提取左侧列表项的模型名称，优先使用执行时快照模型。 // Extracts the model name for list rows, preferring the execution snapshot model.
    getItemModelName(item) {
      const modelName = String(item?.model_name || item?.modelName || item?.current_model || '').trim()
      return modelName
    },
    // getItemWorkspacePath 统一提取工作空间路径，兼容不同历史来源字段。 // Extracts the workspace path for a history row across sources.
    getItemWorkspacePath(item) {
      return String(item?.workspace_path || item?.local_dir || item?.workspacePath || '').trim()
    },
    isItemUnread(item) {
      return item?.is_read === false && item?.status !== 'running'
    },
    // getItemWorkspaceLabel 将工作空间路径压缩为最后一级目录名。 // Reduces a workspace path to its last segment for compact display.
    getItemWorkspaceLabel(item) {
      const workspacePath = this.getItemWorkspacePath(item)
      return this.getCompactDirLabel(workspacePath, '')
    },
    // getCompactDirLabel 将目录路径压缩为最后一级目录名，空值时回退到指定文案。 // Reduces a directory path to its last segment with an optional fallback label.
    getCompactDirLabel(dirPath, fallback = DEFAULT_GROUP_LABEL) {
      const workspacePath = String(dirPath || '').trim()
      if (!workspacePath) return fallback
      const normalized = workspacePath.replace(/[\\/]+$/, '')
      if (!normalized) return workspacePath
      const segments = normalized.split(/[\\/]/).filter(Boolean)
      return segments.length > 0 ? segments[segments.length - 1] : normalized
    },
    // getItemTerminalReasonText 统一提取左侧列表项的终止原因展示文案。 // Extracts a readable terminal reason label for history rows.
    getItemTerminalReasonText(item) {
      if (!item) return ''
      const status = String(item?.status || '').trim()
      if (!status || status === 'running' || status === 'completed') return ''
      const rawText = String(item?.stop_reason_text || '').trim()
      if (rawText) return rawText
      const rawReason = String(item?.stop_reason || '').trim()
      if (!rawReason) return ''
      return this.stopReasonLabel(rawReason)
    },
    // parseHistoryTime 统一解析历史列表中的时间字符串，兼容 YYYY-MM-DD HH:mm:ss 与 ISO 格式。
    parseHistoryTime(value) {
      const text = String(value || '').trim()
      if (!text) return 0
      const parsed = new Date(text.replace(/-/g, '/')).getTime()
      return Number.isNaN(parsed) ? 0 : parsed
    },
    // getItemLastUpdateTime 返回历史项最后更新时间，优先使用 updated_at。
    getItemLastUpdateTime(item) {
      if (!item) return 0
      return this.parseHistoryTime(item.updated_at || item.end_time || item.finished_at || item.created_at)
    },
    // isItemStale 标记 1 小时内没有更新过的任务，便于识别刚完成的记录。
    isItemStale(item) {
      if (!item || item.status === 'running') return false
      const lastUpdateTime = this.getItemLastUpdateTime(item)
      if (lastUpdateTime <= 0) return false
      return (Date.now() - lastUpdateTime) >= 60 * 60 * 1000
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
    // formatCompactToken 将 token 数格式化为 k/m。 // Formats token counts into compact k/m labels.
    formatCompactToken(num) {
      const value = Number(num || 0)
      if (!Number.isFinite(value) || value <= 0) return '0'
      if (value >= 1000000) {
        return (value / 1000000).toFixed(value >= 10000000 ? 0 : 1).replace(/\.0$/, '') + 'm'
      }
      if (value >= 1000) {
        return (value / 1000).toFixed(value >= 10000 ? 0 : 1).replace(/\.0$/, '') + 'k'
      }
      return String(value)
    },
    buildResultSummaryLines(msg) {
      return resultSummaryUtils.buildResultSummaryLines(msg)
    },
    getPrimaryUsageRow(msg) {
      if (!msg) return null
      if (Array.isArray(msg.modelUsage) && msg.modelUsage.length > 0) {
        const currentModelName = String(this.modelName || '').trim()
        const exactMatch = currentModelName
          ? msg.modelUsage.find(item => String(item.name || '').trim() === currentModelName)
          : null
        const picked = exactMatch || msg.modelUsage[0]
        if (picked) {
          return {
            name: picked.name || currentModelName || '模型',
            inputTokens: picked.inputTokens || 0,
            outputTokens: picked.outputTokens || 0,
            cacheReadInputTokens: picked.cacheReadInputTokens || 0,
          }
        }
      }
      if (msg.usage) {
        return {
          name: this.modelName || '模型',
          inputTokens: msg.usage.input_tokens || 0,
          outputTokens: msg.usage.output_tokens || 0,
          cacheReadInputTokens: msg.usage.cache_read_input_tokens || 0,
        }
      }
      return null
    },
  },
}
</script>

<style scoped src="@/css/components/shared/ChatHistory.css"></style>
