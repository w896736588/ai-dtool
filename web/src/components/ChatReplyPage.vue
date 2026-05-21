<template>
  <div class="chat-reply-page">
    <div class="chat-reply-header">
      <span class="chat-reply-header__back" @click="$router.back()">&#8592; 返回</span>
      <span class="chat-reply-header__title">{{ taskName || '对话详情' }}</span>
      <div class="chat-reply-header__meta">
        <span v-if="modelName">智能体: {{ modelName }}</span>
        <span v-if="localDir">目录: {{ localDir }}</span>
        <span>对话 #{{ chatId }}</span>
        <el-tag size="small" :type="statusTagType">{{ statusLabel }}</el-tag>
      </div>
    </div>

    <div class="chat-reply-body">
      <div ref="chatContainer" class="chat-reply-container" @scroll="onScroll">
        <div v-if="messages.length === 0 && status === 'running'" style="text-align: center; padding: 40px; color: #909399;">
          <div>等待 claude code 响应...</div>
        </div>
        <div v-for="(msg, idx) in messages" :key="idx" style="margin-bottom: 8px;">
          <div v-if="msg.type === 'system_init'" style="color: #67c23a; font-size: 12px; padding: 4px 0;">
            {{ msg.text }} | model: {{ msg.model }}
          </div>
          <div v-else-if="msg.type === 'system_command'" style="display: flex; justify-content: flex-end; margin: 4px 0;">
            <div style="background: #ecf5ff; border-radius: 8px 8px 0 8px; padding: 8px 12px; max-width: 75%; width: fit-content; border: 1px solid #d9ecff;">
              <div style="display: flex; align-items: center; justify-content: space-between; margin-bottom: 4px;">
                <span style="font-size: 11px; color: #909399;">提示词</span>
                <span v-if="needCollapseBtn(msg.text)" @click="msg.collapsed = !msg.collapsed" style="cursor: pointer; font-size: 11px; color: #409eff; user-select: none;">{{ msg.collapsed ? '展开 ▼' : '收起 ▲' }}</span>
              </div>
              <div :style="{ maxHeight: needCollapseBtn(msg.text) && msg.collapsed ? '16em' : 'none', overflow: 'hidden', whiteSpace: 'pre-wrap', wordBreak: 'break-word', fontSize: '12px', color: '#303133', lineHeight: '1.6' }">{{ msg.text }}</div>
            </div>
          </div>
          <div v-else-if="msg.type === 'system_hook'" style="color: #909399; font-size: 12px;">
            <span @click="msg.collapsed = !msg.collapsed" style="cursor: pointer;">{{ msg.collapsed ? '▶' : '▼' }} {{ msg.text }}</span>
            <div v-if="!msg.collapsed && (msg.stderr || msg.output)" style="margin-top: 4px; padding: 6px 8px; background: #f5f5f5; border-radius: 4px; font-size: 11px; white-space: pre-wrap; word-break: break-all; max-height: 120px; overflow-y: auto;">
              <div v-if="msg.stderr" style="color: #e6a23c;">{{ msg.stderr }}</div>
              <div v-if="msg.output" style="color: #606266;">{{ msg.output }}</div>
            </div>
          </div>
          <div v-else-if="msg.type === 'system'" style="color: #909399; font-size: 11px;">{{ msg.text }}</div>
          <div v-else-if="msg.type === 'system_status'" style="color: #909399; font-size: 12px; padding: 2px 0;">
            <span :style="msg.status === 'requesting' ? 'color: #409eff;' : ''">{{ msg.text }}</span>
          </div>
          <div v-else-if="msg.type === 'system_task'" style="color: #909399; font-size: 12px; padding: 2px 0;">
            <span v-if="(msg.status === 'started' || msg.status === 'running') && status === 'running'" class="cr-status-spinner"></span>
            <span :style="msg.status === 'started' ? 'color: #409eff;' : ''">🔧 {{ msg.description }}</span>
            <span style="margin-left: 8px; font-size: 11px;">{{ msg.status === 'started' ? '启动' : msg.status }}</span>
          </div>
          <div v-else-if="msg.type === 'assistant'">
            <div v-if="msg.thinking" style="margin-bottom: 8px;">
              <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 4px;">
                <span v-if="isCurrentThinking(msg)" class="cr-status-spinner"></span>
                <span v-if="isCurrentThinking(msg)" style="color: #409eff; font-size: 12px;">思考过程 持续{{ thinkingElapsed }}s</span>
                <span v-else style="color: #909399; font-size: 12px;">思考过程{{ msg._thinkingTiming && msg._thinkingTiming.durationMs ? ' (' + (msg._thinkingTiming.durationMs / 1000).toFixed(1) + 's)' : '' }}</span>
                <span @click="msg._thinkingCollapsed = !msg._thinkingCollapsed" style="cursor: pointer; font-weight: bold; font-size: 12px; color: #909399;">{{ msg._thinkingCollapsed ? '▶' : '▼' }}</span>
              </div>
              <div v-if="!msg._thinkingCollapsed" class="cr-thinking-blockquote">{{ msg.thinking }}</div>
            </div>
            <div v-for="(block, bi) in msg.content" :key="bi">
              <div v-if="block.type === 'text'" class="markdown-body cr-markdown-body" v-html="renderMarkdown(block.text)"></div>
              <div v-else-if="block.type === 'tool_use'" style="background: #f0f9eb; border-radius: 4px; padding: 8px; margin: 4px 0;">
                <div style="display: flex; align-items: center; gap: 4px;">
                  <span v-if="!block._result && status === 'running'" class="cr-status-spinner"></span>
                  <span style="color: #67c23a; font-weight: 500;">🔧 {{ block.name }}</span>
                  <span v-if="block.displayInput" style="font-size: 12px; color: #303133; font-family: Consolas, monospace;">{{ block.displayInput }}</span>
                </div>
                <div v-if="block._tasks" style="margin-top: 6px;">
                  <div v-for="(task, ti) in block._tasks" :key="ti" style="display: flex; align-items: center; gap: 6px; padding: 2px 0; font-size: 12px;">
                    <span :style="{ color: task.status === 'completed' ? '#67c23a' : task.status === 'in_progress' ? '#409eff' : '#909399', fontSize: '14px', lineHeight: 1 }">
                      {{ task.status === 'completed' ? '✅' : task.status === 'in_progress' ? '🔄' : '⬜' }}
                    </span>
                    <span :style="task.status === 'completed' ? 'text-decoration: line-through; color: #909399;' : ''">{{ task.content }}</span>
                  </div>
                </div>
                <div v-if="!block.displayInput && !block._tasks" style="font-size: 12px; color: #909399; margin-top: 4px; cursor: pointer;" @click="block._inputExpanded = !block._inputExpanded">
                  {{ block._inputExpanded ? '▼' : '▶' }} 参数
                </div>
                <pre v-if="!block.displayInput && !block._tasks && block._inputExpanded" style="white-space: pre-wrap; font-size: 12px; color: #606266; margin-top: 4px; font-family: Consolas, monospace;">{{ block.input }}</pre>
                <div v-if="block._result" style="color: #909399; font-size: 12px; margin-top: 6px; border-top: 1px dashed #dcdfe6; padding-top: 4px;">
                  <span @click="block._result.collapsed = !block._result.collapsed" style="cursor: pointer;">{{ block._result.collapsed ? '▶' : '▼' }} 工具执行结果</span>
                  <pre v-if="!block._result.collapsed" style="white-space: pre-wrap; font-size: 11px; margin-top: 4px; max-height: 200px; overflow-y: auto; font-family: Consolas, monospace;">{{ block._result.text }}</pre>
                </div>
              </div>
            </div>
            <div v-if="msg.usage" style="color: #909399; font-size: 11px; margin-top: 8px; border-top: 1px solid #ebeef5; padding-top: 4px;">
              input: {{ msg.usage.input_tokens }} | output: {{ msg.usage.output_tokens }}
            </div>
          </div>
          <!-- standalone tool_use -->
          <div v-else-if="msg.type === 'tool_use'" style="background: #f0f9eb; border-radius: 4px; padding: 8px; margin: 4px 0;">
            <div style="display: flex; align-items: center; gap: 4px;">
              <span v-if="!msg._result && status === 'running'" class="cr-status-spinner"></span>
              <span style="color: #67c23a; font-weight: 500;">🔧 {{ msg.name }}</span>
              <span v-if="msg.displayInput" style="font-size: 12px; color: #303133; font-family: Consolas, monospace;">{{ msg.displayInput }}</span>
            </div>
            <div v-if="msg._tasks" style="margin-top: 6px;">
              <div v-for="(task, ti) in msg._tasks" :key="ti" style="display: flex; align-items: center; gap: 6px; padding: 2px 0; font-size: 12px;">
                <span :style="{ color: task.status === 'completed' ? '#67c23a' : task.status === 'in_progress' ? '#409eff' : '#909399', fontSize: '14px', lineHeight: 1 }">
                  {{ task.status === 'completed' ? '✅' : task.status === 'in_progress' ? '🔄' : '⬜' }}
                </span>
                <span :style="task.status === 'completed' ? 'text-decoration: line-through; color: #909399;' : ''">{{ task.content }}</span>
              </div>
            </div>
            <div v-if="!msg.displayInput && !msg._tasks" style="font-size: 12px; color: #909399; margin-top: 4px; cursor: pointer;" @click="msg._inputExpanded = !msg._inputExpanded">
              {{ msg._inputExpanded ? '▼' : '▶' }} 参数
            </div>
            <pre v-if="!msg.displayInput && !msg._tasks && msg._inputExpanded" style="white-space: pre-wrap; font-size: 12px; color: #606266; margin-top: 4px; font-family: Consolas, monospace;">{{ msg.input }}</pre>
            <div v-if="msg._result" style="color: #909399; font-size: 12px; margin-top: 6px; border-top: 1px dashed #dcdfe6; padding-top: 4px;">
              <span @click="msg._result.collapsed = !msg._result.collapsed" style="cursor: pointer;">{{ msg._result.collapsed ? '▶' : '▼' }} 工具执行结果</span>
              <pre v-if="!msg._result.collapsed" style="white-space: pre-wrap; font-size: 11px; margin-top: 4px; max-height: 200px; overflow-y: auto; font-family: Consolas, monospace;">{{ msg._result.text }}</pre>
            </div>
          </div>
          <!-- tool_result fallback -->
          <div v-else-if="msg.type === 'tool_result'" style="color: #909399; font-size: 12px;">
            <span @click="msg.collapsed = !msg.collapsed" style="cursor: pointer;">{{ msg.collapsed ? '▶' : '▼' }} 工具执行结果</span>
            <pre v-if="!msg.collapsed" style="white-space: pre-wrap; font-size: 11px; margin-top: 4px; max-height: 200px; overflow-y: auto; font-family: Consolas, monospace;">{{ msg.text }}</pre>
          </div>
          <div v-else-if="msg.type === 'assistant_text'" class="markdown-body cr-markdown-body" v-html="renderMarkdown(msg.text)"></div>
          <div v-else-if="msg.type === 'assistant_thinking'" style="color: #909399; font-size: 12px;">
            <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 4px;">
              <span>思考过程{{ msg._thinkingTiming && msg._thinkingTiming.durationMs ? ' (' + (msg._thinkingTiming.durationMs / 1000).toFixed(1) + 's)' : '' }}</span>
              <span @click="msg._thinkingCollapsed = !msg._thinkingCollapsed" style="cursor: pointer; font-weight: bold;">{{ msg._thinkingCollapsed ? '▶' : '▼' }}</span>
            </div>
            <div v-if="!msg._thinkingCollapsed" class="cr-thinking-blockquote">{{ msg.text }}</div>
          </div>
          <div v-else-if="msg.type === 'result'" class="cr-result-card">
            <div class="cr-result-header">
              <span :style="{ color: msg.isError ? '#f56c6c' : '#67c23a', fontWeight: 'bold' }">
                {{ msg.isError ? '✕ 执行失败' : '✓ 执行完成' }}
              </span>
              <span class="cr-result-header-item">耗时 {{ (msg.durationMs / 1000).toFixed(1) }}s</span>
              <span v-if="msg.durationApiMs" class="cr-result-header-item">API {{ (msg.durationApiMs / 1000).toFixed(1) }}s</span>
              <span class="cr-result-header-item">{{ msg.numTurns }} 轮对话</span>
              <span v-if="msg.totalCostUsd != null" class="cr-result-header-item" style="color: #e6a23c;">${{ msg.totalCostUsd.toFixed(4) }}</span>
              <span v-if="msg.stopReason" class="cr-result-header-item" style="color: #909399;">{{ stopReasonLabel(msg.stopReason) }}</span>
            </div>
            <div v-if="msg.usage" class="cr-result-section">
              <div class="cr-result-section-title">Token 用量</div>
              <div class="cr-result-tokens">
                <span>输入 {{ formatNum(msg.usage.input_tokens) }}</span>
                <span>输出 {{ formatNum(msg.usage.output_tokens) }}</span>
                <span v-if="msg.usage.cache_read_input_tokens">缓存读取 {{ formatNum(msg.usage.cache_read_input_tokens) }}</span>
                <span v-if="msg.usage.cache_creation_input_tokens">缓存创建 {{ formatNum(msg.usage.cache_creation_input_tokens) }}</span>
              </div>
            </div>
            <div v-if="msg.resultText" class="cr-result-section">
              <div class="cr-result-section-title">结果</div>
              <pre class="cr-result-text">{{ msg.resultText }}</pre>
            </div>
          </div>
          <div v-else-if="msg.type === 'chat_completed' && status === 'completed'" style="color: #67c23a; text-align: center; padding: 16px;">
            {{ msg.text }}
          </div>
          <div v-else-if="msg.type === 'raw_text'" style="white-space: pre-wrap; color: #e6a23c; padding: 4px 0; word-break: break-all; font-family: Consolas, monospace;">{{ msg.text }}</div>
          <div v-else-if="msg.type === 'error'" style="background: #fef0f0; border-left: 3px solid #f56c6c; border-radius: 4px; padding: 8px 12px; margin: 4px 0;">
            <span style="color: #f56c6c;">错误: </span>
            <span style="color: #303133;">{{ msg.text }}</span>
          </div>
        </div>
      </div>
      <div :class="['chat-reply-scroll-btn', { 'chat-reply-scroll-btn--visible': showScrollBtn }]" @click="scrollToBottom(true)">↓</div>
      <div class="chat-reply-input-row">
        <div class="chat-reply-textarea-wrapper">
          <el-input
            v-model="continueInput"
            type="textarea"
            :rows="3"
            placeholder="输入新消息继续对话..."
            :disabled="status === 'running'"
            class="chat-reply-textarea"
            @keydown.enter.exact.prevent="status !== 'running' && continueChat()"
          />
          <div class="chat-reply-actions">
            <div v-if="thinkingIntensity || modelName" class="chat-reply-info-bar">
              <span v-if="thinkingIntensity">思考强度: {{ thinkingIntensity }}</span>
              <span v-if="thinkingIntensity && modelName"> | </span>
              <span v-if="modelName">智能体: {{ modelName }}</span>
            </div>
            <el-button v-if="status === 'running'" type="danger" size="small" @click="stopChat">停止</el-button>
            <el-button v-else type="primary" size="small" :loading="continueLoading" @click="continueChat">发送</el-button>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import taskWorkflowApi from '@/utils/base/task_workflow'
import baseUtils from '@/utils/base'
import chatParser from '@/utils/chat_parser'
import MarkdownIt from 'markdown-it'

const md = new MarkdownIt({ html: true, breaks: true, linkify: true })

export default {
  data() {
    return {
      chatId: 0,
      taskName: '',
      modelName: '',
      localDir: '',
      thinkingIntensity: '',
      status: '',
      messages: [],
      sseLines: [],
      cliType: 'claude',
      continueInput: '',
      continueLoading: false,
      showScrollBtn: false,
      autoScroll: true,
      thinkingElapsed: 0,
    }
  },
  computed: {
    statusLabel() {
      const map = { running: '运行中', completed: '已完成', interrupted: '已中断', failed: '失败' }
      return map[this.status] || this.status || '加载中'
    },
    statusTagType() {
      const map = { running: '', completed: 'success', interrupted: 'warning', failed: 'danger' }
      return map[this.status] || 'info'
    },
  },
  mounted() {
    this.chatId = parseInt(this.$route.params.chatId) || 0
    if (this.chatId > 0) {
      this.loadChatDetail()
    }
  },
  beforeUnmount() {
    this.closeSSE()
    if (this._thinkingTimer) { clearInterval(this._thinkingTimer); this._thinkingTimer = null }
    if (this._sseBatchTimer) { clearTimeout(this._sseBatchTimer); this._sseBatchTimer = null }
  },
  methods: {
    loadChatDetail() {
      taskWorkflowApi.TaskWorkflowChatDetail(this.chatId, (res) => {
        if (res.ErrCode === 0 && res.Data) {
          const data = res.Data
          this.taskName = data.task_name || ''
          this.modelName = data.model_name || ''
          this.localDir = data.local_dir || ''
          this.thinkingIntensity = data.thinking_intensity || ''
          this.cliType = data.cli_type || 'claude'
          this.status = data.status || ''
          const historicalLines = data.lines || []
          const newSseLines = this.sseLines.filter(l => !historicalLines.includes(l))
          this.sseLines = [...historicalLines, ...newSseLines]
          this.messages = chatParser.parseChatLines(this.sseLines, this.cliType)
          this.messages.forEach(msg => {
            if (msg.type === 'assistant' && msg.thinking) msg._thinkingCollapsed = true
            if (msg.type === 'assistant_thinking') msg._thinkingCollapsed = true
          })
          this.$nextTick(() => { this.scrollToBottom(true) })
          if (this.status === 'running' && this._sseChatId !== this.chatId) {
            this.connectChatStream(this.chatId)
          }
        }
      })
    },
    connectChatStream(chatId, continuePrompt) {
      if (this._sseChatId === chatId && this._chatEventSource && this._chatEventSource.readyState !== EventSource.CLOSED) return
      this.closeSSE()
      this._sseChatId = chatId
      this._thinkingStreamStartTime = 0
      this._sseParseState = { currentMessage: null, toolUseMap: new Map(), pendingPatches: [] }
      this._sseLineBuffer = []
      if (this._sseBatchTimer) { clearTimeout(this._sseBatchTimer); this._sseBatchTimer = null }
      if (this._thinkingTimer) { clearInterval(this._thinkingTimer); this._thinkingTimer = null }
      this.thinkingElapsed = 0
      this._thinkingTimer = setInterval(() => {
        if (this._thinkingStreamStartTime > 0) {
          this.thinkingElapsed = Math.floor((Date.now() - this._thinkingStreamStartTime) / 1000)
        } else {
          this.thinkingElapsed = 0
        }
      }, 200)
      const sseHost = baseUtils.GetSseApiHost()
      let url = sseHost + '/api/task/workflow/chat/stream?chat_id=' + chatId + '&token=' + encodeURIComponent(baseUtils.GetSafeToken())
      if (continuePrompt) {
        url += '&continue=1&prompt=' + encodeURIComponent(continuePrompt)
      }
      const es = new EventSource(url)
      this._chatEventSource = es
      es.onmessage = (event) => {
        const line = event.data
        if (!line) return
        try {
          const obj = JSON.parse(line)
          if (obj.type === 'chat' && obj.subtype === 'completed') {
            this._flushSseBatch()
            this.sseLines.push(line)
            this._sseChatId = 0
            es.close()
            this._chatEventSource = null
            this._sseParseState = null
            this.loadChatDetail()
            this.$nextTick(() => { this.scrollToBottom() })
            return
          }
          if (obj.type === 'stream_event') {
            const evt = obj.event || {}
            if (evt.type === 'content_block_delta') {
              const delta = evt.delta || {}
              if (delta.type === 'thinking_delta' && this._thinkingStreamStartTime === 0) {
                this._thinkingStreamStartTime = Date.now()
              }
            } else if (evt.type === 'message_stop' && this._thinkingStreamStartTime > 0) {
              this._pendingThinkingDurationMs = Date.now() - this._thinkingStreamStartTime
              this._thinkingStreamStartTime = 0
            }
          }
        } catch (e) { /* ignore */ }
        this._sseLineBuffer.push(line)
        if (!this._sseBatchTimer) {
          this._sseBatchTimer = setTimeout(() => { this._flushSseBatch() }, 100)
        }
      }
      es.onerror = () => {
        this._flushSseBatch()
        if (this._thinkingTimer) { clearInterval(this._thinkingTimer); this._thinkingTimer = null }
        this.thinkingElapsed = 0
        es.close()
        this._chatEventSource = null
        this._sseParseState = null
        this.loadChatDetail()
      }
    },
    _flushSseBatch() {
      if (this._sseBatchTimer) { clearTimeout(this._sseBatchTimer); this._sseBatchTimer = null }
      const newLines = this._sseLineBuffer.splice(0)
      if (newLines.length === 0) return
      for (const l of newLines) { this.sseLines.push(l) }
      const result = chatParser.parseChatLinesIncremental(newLines, this._sseParseState, this.messages.length, this.cliType)
      this._sseParseState = result.parseState
      if (result.newMessages.length > 0) {
        for (const msg of result.newMessages) { this.messages.push(msg) }
      }
      for (const patch of result.parseState.pendingPatches) {
        for (let i = this.messages.length - 1; i >= 0; i--) {
          const msg = this.messages[i]
          if (patch.type === 'tool_result' && msg.type === 'assistant' && msg.content) {
            const tu = msg.content.find(b => b.type === 'tool_use' && b.id === patch.toolUseId)
            if (tu) { tu._result = patch.result; break }
          }
          if (patch.type === 'tool_result' && msg.type === 'tool_use' && msg.id === patch.toolUseId) {
            msg._result = patch.result; break
          }
        }
      }
      result.parseState.pendingPatches = []
      if (this._pendingThinkingDurationMs) {
        for (let i = this.messages.length - 1; i >= 0; i--) {
          const m = this.messages[i]
          if (m.type === 'assistant' && m.thinking) {
            m._thinkingTiming = { durationMs: this._pendingThinkingDurationMs }
            break
          }
        }
        this._pendingThinkingDurationMs = 0
      }
      this.$nextTick(() => { this.scrollToBottom() })
    },
    closeSSE() {
      if (this._chatEventSource) {
        this._chatEventSource.close()
        this._chatEventSource = null
      }
      this._sseChatId = 0
    },
    continueChat() {
      const input = this.continueInput.trim()
      if (!input) return
      this.continueLoading = true
      taskWorkflowApi.TaskWorkflowChatContinue(this.chatId, input, (res) => {
        this.continueLoading = false
        if (res.ErrCode === 0) {
          this.continueInput = ''
          this.status = 'running'
          this.connectChatStream(this.chatId, input)
          setTimeout(() => { this.loadChatDetail() }, 500)
        } else {
          this.$message.error(res.ErrMsg || '发送失败')
        }
      })
    },
    stopChat() {
      this.closeSSE()
      taskWorkflowApi.TaskWorkflowChatStop(this.chatId, (res) => {
        if (res.ErrCode !== 0) {
          this.$message.error(res.ErrMsg || '停止失败')
        }
      })
      this.status = 'interrupted'
    },
    onScroll() {
      const el = this.$refs.chatContainer
      if (!el) return
      const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 80
      this.showScrollBtn = !atBottom
      this.autoScroll = atBottom
    },
    scrollToBottom(force) {
      if (!force && !this.autoScroll) return
      const el = this.$refs.chatContainer
      if (el) {
        this.autoScroll = true
        el.scrollTop = el.scrollHeight
      }
    },
    needCollapseBtn(text) {
      return (text || '').split('\n').length > 10
    },
    isCurrentThinking(msg) {
      if (!this._thinkingStreamStartTime) return false
      for (let i = this.messages.length - 1; i >= 0; i--) {
        const m = this.messages[i]
        if (m.type === 'assistant' && m.thinking) return m === msg
      }
      return false
    },
    renderMarkdown(text) {
      if (!text) return ''
      return md.render(text)
    },
    formatNum(num) {
      if (num == null) return '0'
      return Number(num).toLocaleString()
    },
    stopReasonLabel(reason) {
      const map = { end_turn: '正常结束', stop_sequence: '停止序列', max_tokens: '达到上限', tool_use: '工具调用' }
      return map[reason] || reason
    },
  },
}
</script>

<style src="@/css/components/ChatReplyPage.css"></style>
