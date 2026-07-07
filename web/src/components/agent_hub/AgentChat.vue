<template>
  <div class="agent-chat">
    <!-- 左侧面板：工作空间 + 会话列表 -->
    <aside class="chat-sidebar">
      <div class="chat-sidebar__header">
        <el-button text @click="goBack">
          <el-icon><ArrowLeft /></el-icon>
          Agent Hub
        </el-button>
      </div>

      <!-- 工作空间选择 -->
      <div class="chat-sidebar__section">
        <div class="chat-sidebar__section-title">
          <span>工作空间</span>
          <el-button text size="small" @click="showWorkspaceDialog = true">+</el-button>
        </div>
        <div class="workspace-list">
          <div
            v-for="ws in workspaces"
            :key="ws.id"
            class="workspace-item"
            :class="{ 'workspace-item--active': ws.id === currentWorkspaceId }"
            @click="selectWorkspace(ws)"
          >
            <el-icon><Folder /></el-icon>
            <span class="workspace-item__name">{{ ws.name }}</span>
            <span class="workspace-item__path">{{ ws.path }}</span>
          </div>
          <div v-if="workspaces.length === 0" class="empty-hint">暂无工作空间，点击 + 添加</div>
        </div>
      </div>

      <!-- 会话列表 -->
      <div class="chat-sidebar__section chat-sidebar__section--grow">
        <div class="chat-sidebar__section-title">
          <span>对话列表</span>
          <el-button text size="small" @click="createSession" :disabled="!currentWorkspaceId">+</el-button>
        </div>
        <div class="session-list">
          <div
            v-for="session in sessions"
            :key="session.id"
            class="session-item"
            :class="{ 'session-item--active': session.id === currentSessionId }"
            @click="selectSession(session)"
            @contextmenu.prevent="showSessionMenu($event, session)"
          >
            <el-icon><ChatDotRound /></el-icon>
            <div class="session-item__info">
              <span class="session-item__name">{{ session.name }}</span>
              <span class="session-item__time">{{ formatTime(session.updated_at) }}</span>
            </div>
            <el-button text size="small" class="session-item__del" @click.stop="deleteSession(session)">
              <el-icon><Close /></el-icon>
            </el-button>
          </div>
          <div v-if="sessions.length === 0" class="empty-hint">
            {{ currentWorkspaceId ? '暂无对话，点击 + 创建' : '请先选择工作空间' }}
          </div>
        </div>
      </div>

      <!-- 底部 Agent 信息 -->
      <div class="chat-sidebar__footer">
        <div class="agent-info">
          <span class="agent-info__badge agent-info__badge--pi">π</span>
          <span>{{ agentName }}</span>
          <span class="agent-info__status" :class="{ 'agent-info__status--connected': wsConnected }">
            {{ wsConnected ? '已连接' : '未连接' }}
          </span>
        </div>
        <el-button text size="small" @click="openConfig">配置</el-button>
      </div>
    </aside>

    <!-- 右侧主区域：聊天界面 -->
    <main class="chat-main">
      <!-- 顶部标题栏 -->
      <header class="chat-header">
        <div class="chat-header__title">
          <span v-if="currentSession">{{ currentSession.name }}</span>
          <span v-else>选择一个对话或创建新对话</span>
        </div>
        <div class="chat-header__actions">
          <!-- Token 用量统计（会话级） -->
          <div v-if="tokenStats" class="token-stats">
            <span class="token-stats__item">输入: {{ fmtNum(tokenStats.input_tokens) }}</span>
            <span class="token-stats__item">输出: {{ fmtNum(tokenStats.output_tokens) }}</span>
            <span class="token-stats__item" v-if="tokenStats.total_cost">费用: ${{ fmtCost(tokenStats.total_cost) }}</span>
          </div>
          <el-button
            v-if="isStreaming"
            type="danger"
            size="small"
            plain
            @click="abortAgent"
          >停止</el-button>
        </div>
      </header>

      <!-- 消息列表 -->
      <div class="chat-messages" ref="messagesContainer">
        <div v-if="messages.length === 0 && !isStreaming" class="chat-empty">
          <div v-if="historyLoading" class="chat-empty__loading">
            <span class="agent-status-spinner agent-status-spinner--large"></span>
            <p>加载历史对话...</p>
          </div>
          <template v-else>
          <div class="chat-empty__icon">π</div>
          <p>开始与 Pi Agent 对话</p>
          <p class="chat-empty__hint">Pi 可以读取、编辑和运行代码，帮助你完成开发任务</p>
          </template>
        </div>

        <div
          v-for="(msg, idx) in messages"
          :key="idx"
          class="chat-message"
          :class="'chat-message--' + msg.role"
        >
          <div class="chat-message__avatar">
            <span v-if="msg.role === 'user'" class="avatar avatar--user">U</span>
            <span v-else-if="msg.role === 'tool'" class="avatar avatar--tool">🔧</span>
            <span v-else class="avatar avatar--assistant">π</span>
          </div>
          <div class="chat-message__body">
            <div v-if="msg.role === 'assistant' && msg.thinking" class="thinking-block">
              <div class="thinking-block__header" @click="toggleMessageThinking(msg)">
                <span class="agent-status-check">✓</span>
                <span>思考过程（{{ getThinkingDurationText(msg) }}）</span>
                <el-icon class="thinking-block__arrow" :class="{ 'thinking-block__arrow--open': !isMessageThinkingCollapsed(msg) }">
                  <ArrowDown />
                </el-icon>
              </div>
              <div v-if="!isMessageThinkingCollapsed(msg)" class="thinking-block__content">{{ msg.thinking }}</div>
            </div>

            <div v-if="msg.role === 'tool' || msg.content" class="chat-message__content" v-html="renderContent(msg)"></div>

            <!-- 工具调用展示 -->
            <div v-if="msg.toolCalls && msg.toolCalls.length" class="tool-calls">
              <div
                v-for="tc in msg.toolCalls"
                :key="tc.id"
                class="tool-call"
                :class="'tool-call--' + tc.status"
              >
                <!-- read/bash/edit 紧凑一行 -->
                <template v-if="isReadOrBashTool(tc)">
                  <div class="tool-call__compact-row" @click="toggleToolCallCollapse(tc)">
                    <span v-if="isToolRunning(tc)" class="agent-status-spinner"></span>
                    <span v-else-if="isToolDone(tc)" class="agent-status-check">✓</span>
                    <span class="tool-call__compact-label">{{ tc.name }}</span>
                    <span class="tool-call__compact-text" :title="getCompactText(tc)">{{ getCompactText(tc) }}</span>
                    <span class="tool-call__compact-status">{{ statusLabel(tc.status) }}（{{ getToolDurationText(tc) }}）</span>
                    <el-icon class="tool-call__compact-arrow" :class="{ 'tool-call__compact-arrow--open': !isToolCallCollapsed(tc) }">
                      <ArrowRight />
                    </el-icon>
                  </div>
                  <div v-if="!isToolCallCollapsed(tc)" class="tool-call__details">
                    <pre class="tool-call__input" v-if="tc.input">{{ formatJSON(tc.input) }}</pre>
                    <pre class="tool-call__output" v-if="tc.output">{{ tc.output }}</pre>
                  </div>
                </template>
                <!-- 其他工具完整展示 -->
                <template v-else>
                  <div class="tool-call__header">
                    <span v-if="isToolRunning(tc)" class="agent-status-spinner"></span>
                    <span v-else-if="isToolDone(tc)" class="agent-status-check">✓</span>
                    <el-icon><Tools /></el-icon>
                    <span class="tool-call__name">{{ tc.name }}</span>
                    <el-tag :type="tc.status === 'done' ? 'success' : tc.status === 'running' ? 'warning' : 'info'" size="small">
                      {{ statusLabel(tc.status) }}（{{ getToolDurationText(tc) }}）
                    </el-tag>
                  </div>
                  <pre class="tool-call__input" v-if="tc.input">{{ formatJSON(tc.input) }}</pre>
                  <pre class="tool-call__output" v-if="tc.output">{{ tc.output }}</pre>
                </template>
              </div>
            </div>
          </div>
        </div>

        <!-- 流式输出中的消息 -->
        <div v-if="streamingText || streamingThinking || hasRunningTools" class="chat-message chat-message--assistant">
          <div class="chat-message__avatar">
            <span class="avatar avatar--assistant">π</span>
          </div>
          <div class="chat-message__body">
            <!-- 思考过程 -->
            <div v-if="streamingThinking" class="thinking-block">
              <div class="thinking-block__header" @click="showThinking = !showThinking">
                <span class="agent-status-spinner"></span>
                <span>思考过程（{{ getStreamingThinkingDurationText() }}）</span>
                <el-icon class="thinking-block__arrow" :class="{ 'thinking-block__arrow--open': showThinking }">
                  <ArrowDown />
                </el-icon>
              </div>
              <div v-if="showThinking" class="thinking-block__content">{{ streamingThinking }}</div>
            </div>
            <!-- 流式工具调用 -->
            <div v-if="hasRunningTools" class="tool-calls">
              <div v-for="tc in runningToolCalls" :key="tc.id" class="tool-call" :class="'tool-call--' + tc.status">
                <!-- read/bash/edit 紧凑一行 -->
                <template v-if="isReadOrBashTool(tc)">
                  <div class="tool-call__compact-row" @click="toggleToolCallCollapse(tc)">
                    <span v-if="isToolRunning(tc)" class="agent-status-spinner"></span>
                    <span v-else-if="isToolDone(tc)" class="agent-status-check">✓</span>
                    <span class="tool-call__compact-label">{{ tc.name }}</span>
                    <span class="tool-call__compact-text" :title="getCompactText(tc)">{{ getCompactText(tc) }}</span>
                    <span class="tool-call__compact-status">{{ statusLabel(tc.status) }}（{{ getToolDurationText(tc) }}）</span>
                    <el-icon class="tool-call__compact-arrow" :class="{ 'tool-call__compact-arrow--open': !isToolCallCollapsed(tc) }">
                      <ArrowRight />
                    </el-icon>
                  </div>
                  <div v-if="!isToolCallCollapsed(tc)" class="tool-call__details">
                    <pre class="tool-call__input" v-if="tc.input">{{ formatJSON(tc.input) }}</pre>
                    <pre class="tool-call__output" v-if="tc.output">{{ tc.output }}</pre>
                  </div>
                </template>
                <!-- 其他工具完整展示 -->
                <template v-else>
                  <div class="tool-call__header">
                    <span v-if="isToolRunning(tc)" class="agent-status-spinner"></span>
                    <span v-else-if="isToolDone(tc)" class="agent-status-check">✓</span>
                    <el-icon><Tools /></el-icon>
                    <span class="tool-call__name">{{ tc.name }}</span>
                    <el-tag :type="tc.status === 'done' ? 'success' : 'warning'" size="small">
                      {{ statusLabel(tc.status) }}（{{ getToolDurationText(tc) }}）
                    </el-tag>
                  </div>
                  <pre class="tool-call__input" v-if="tc.input">{{ formatJSON(tc.input) }}</pre>
                  <pre class="tool-call__output" v-if="tc.output">{{ tc.output }}</pre>
                </template>
              </div>
            </div>
            <div v-if="streamingText" class="chat-message__content" v-html="renderMarkdown(streamingText)"></div>
            <span class="cursor-blink">▊</span>
          </div>
        </div>

        <!-- 压缩通知 -->
        <div v-if="compacting" class="compaction-notice">
          <el-icon><Loading /></el-icon> 正在压缩上下文...
        </div>
      </div>

      <!-- 输入区域 -->
      <footer class="chat-input" v-if="currentSession || pendingSession">
        <div class="chat-input__wrapper">
          <el-input
            v-model="inputText"
            type="textarea"
            :rows="2"
            placeholder="输入消息，Enter 发送，Shift+Enter 换行..."
            :disabled="isStreaming || (!!currentSessionId && !wsConnected)"
            @keydown.enter.exact.prevent="sendMessage"
            resize="none"
          />
          <div class="chat-input__toolbar">
            <div class="chat-input__toolbar-left">
              <!-- 模型选择（按 Provider 分组） -->
              <el-select
                v-if="providerModels.length > 0"
                v-model="selectedModel"
                size="small"
                style="width: 200px"
                placeholder="选择模型"
                :disabled="isStreaming"
                @change="setModel"
              >
                <el-option-group
                  v-for="group in providerModels"
                  :key="group.provider_id"
                  :label="group.provider_name"
                >
                  <el-option
                    v-for="m in group.models"
                    :key="m.id"
                    :label="m.name + ' (' + m.model + ')'"
                    :value="group.provider_name + '/' + m.model"
                  />
                </el-option-group>
              </el-select>

              <!-- Skills 选择 -->
              <el-popover
                placement="top-start"
                trigger="click"
                :width="260"
                popper-class="skills-popover"
              >
                <template #reference>
                  <el-button size="small" class="chat-input__toolbar-btn">
                    Skills
                    <el-icon class="el-icon--right"><ArrowDown /></el-icon>
                  </el-button>
                </template>
                <div class="skills-popover__list">
                  <div
                    v-for="sk in skills"
                    :key="sk.id"
                    class="skills-popover__item"
                    :class="{ 'skills-popover__item--active': selectedSkillIds.includes(sk.id) }"
                    @click="toggleSkill(sk)"
                  >
                    <el-checkbox :model-value="selectedSkillIds.includes(sk.id)" size="small" />
                    <span class="skills-popover__item-name">{{ sk.name }}</span>
                    <el-tag v-if="sk.skill_type" size="small" type="info" effect="plain">{{ sk.skill_type }}</el-tag>
                  </div>
                  <div v-if="skills.length === 0" class="skills-popover__empty">暂无 Skills</div>
                </div>
              </el-popover>

              <!-- 上下文使用率 -->
              <span class="chat-input__stat-item">
                上下文: <strong>{{ contextUsage }}</strong>
              </span>

              <!-- Token 统计 -->
              <span class="chat-input__stat-item">
                输入: <strong>{{ tStat('input_tokens') }}</strong>
                <span class="chat-input__stat-divider">/</span>
                缓存: <strong>{{ tStat('cached_input_tokens') }}</strong>
                <span class="chat-input__stat-divider">/</span>
                输出: <strong>{{ tStat('output_tokens') }}</strong>
              </span>
            </div>
            <div class="chat-input__toolbar-right">
              <span class="chat-input__hint">{{ inputHint }}</span>
              <el-button
                v-if="isStreaming"
                type="danger"
                @click="abortAgent"
              >
                终止
              </el-button>
              <el-button
                v-else
                type="primary"
                :disabled="!inputText.trim() || !wsConnected"
                @click="sendMessage"
              >
                发送
              </el-button>
            </div>
          </div>
        </div>
      </footer>
    </main>

    <!-- 工作空间对话框 -->
    <el-dialog v-model="showWorkspaceDialog" title="添加工作空间" width="480px">
      <el-form label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="workspaceForm.name" placeholder="例如：my-project" />
        </el-form-item>
        <el-form-item label="路径">
          <el-input v-model="workspaceForm.path" placeholder="例如：C:/work/my-project" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showWorkspaceDialog = false">取消</el-button>
        <el-button type="primary" @click="saveWorkspace">保存</el-button>
      </template>
    </el-dialog>

    <!-- 会话重命名对话框 -->
    <el-dialog v-model="showRenameDialog" title="重命名会话" width="400px">
      <el-input v-model="renameForm.name" />
      <template #footer>
        <el-button @click="showRenameDialog = false">取消</el-button>
        <el-button type="primary" @click="renameSession">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import Base from '@/utils/base.js'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import {
  ArrowLeft,
  ArrowDown,
  ArrowRight,
  Folder,
  ChatDotRound,
  Close,
  Tools,
  Loading
} from '@element-plus/icons-vue'

export default {
  name: 'AgentChat',
  components: {
    ArrowLeft,
    ArrowDown,
    ArrowRight,
    Folder,
    ChatDotRound,
    Close,
    Tools,
    Loading
  },
  data() {
    return {
      agentId: 0,
      agentName: '',
      agentConfig: null,

      pendingSession: false, // 新建对话标记：为 true 时不创建 DB 记录，等用户发消息后再创建

      workspaces: [],
      currentWorkspaceId: 0,
      showWorkspaceDialog: false,
      workspaceForm: { name: '', path: '' },

      sessions: [],
      currentSessionId: 0,
      currentSession: null,
      showRenameDialog: false,
      renameForm: { id: 0, name: '' },

      messages: [],
      inputText: '',
      isStreaming: false,
      streamingText: '',
      streamingThinking: '',
      showThinking: true,

      selectedModel: '',
      // 按 Provider 分组的模型列表 [{provider_id, provider_name, provider_type, models: [{id, name, model}]}]
      providerModels: [],

      ws: null,
      wsConnected: false,

      pendingToolCalls: {},
      tokenStats: null,
      compacting: false,
      turnCount: 0,

      // 运行时计时器
      thinkingStartAt: 0,
      thinkingElapsedSeconds: 0,
      runtimeNow: Date.now(),
      _runtimeTicker: null,
      _statsPollTimer: null,

      // 加载状态
      historyLoading: false,

      // Skills 数据
      skills: [],
      selectedSkillIds: []
    }
  },
  computed: {
    hasRunningTools() {
      return Object.values(this.pendingToolCalls).some(tc => tc.status !== 'done')
    },
    runningToolCalls() {
      return Object.values(this.pendingToolCalls)
    },
    inputHint() {
      if (!this.wsConnected) return '未连接'
      if (this.isStreaming) return 'Pi 正在思考...'
      if (this.compacting) return '正在压缩上下文...'
      return 'Enter 发送'
    },
    contextUsage() {
      if (!this.tokenStats) return '--'
      const used = this.tokenStats.context_used || this.tokenStats.input_tokens || 0
      const total = this.tokenStats.context_total || this.tokenStats.max_input_tokens || 0
      if (total <= 0) return '--'
      const pct = Math.round((used / total) * 100)
      return pct + '%' + ' (' + this.fmtNum(used) + '/' + this.fmtNum(total) + ')'
    }
  },
  mounted() {
    this.agentId = parseInt(this.$route.query.agent_id) || 0
    if (!this.agentId) {
      this.$router.push('/AgentHub')
      return
    }
    this.loadAgentInfo()
    this.loadWorkspaces()
    this.loadSkills()
  },
  beforeUnmount() {
    this.disconnectWS()
    this.stopRuntimeTicker()
    this.stopStatsPolling()
  },
  methods: {
    goBack() {
      this.disconnectWS()
      this.stopRuntimeTicker()
      this.$router.push('/AgentHub')
    },
    openConfig() {
      this.$router.push({ path: '/AgentConfig', query: { agent_id: this.agentId } })
    },

    // ========== 工作空间 ==========
    async loadWorkspaces() {
      Base.BasePost('/api/AgentV2WorkspaceList', { agent_id: this.agentId }, (res) => {
        this.workspaces = (res.ErrCode === 0 && res.Data && res.Data.list) ? res.Data.list : []
        if (this.workspaces.length === 1) {
          this.selectWorkspace(this.workspaces[0])
        }
      })
    },
    selectWorkspace(ws) {
      this.currentWorkspaceId = ws.id
      this.loadSessions()
    },
    saveWorkspace() {
      if (!this.workspaceForm.name || !this.workspaceForm.path) return
      Base.BasePost('/api/AgentV2WorkspaceSave', {
        agent_id: this.agentId,
        name: this.workspaceForm.name,
        path: this.workspaceForm.path
      }, () => {
        this.showWorkspaceDialog = false
        this.workspaceForm = { name: '', path: '' }
        this.loadWorkspaces()
      })
    },

    // ========== 会话管理 ==========
    async loadSessions() {
      Base.BasePost('/api/AgentV2SessionList', { agent_id: this.agentId, workspace_id: this.currentWorkspaceId }, (res) => {
        this.sessions = (res.ErrCode === 0 && res.Data) ? (res.Data.list || []) : []
      })
    },
    createSession() {
      if (!this.currentWorkspaceId) return
      // 仅打开空白聊天区，不创建 DB 记录、不连 WebSocket
      // 等用户输入第一条消息时才真正创建会话
      this.disconnectWS()
      this.currentSessionId = 0
      this.currentSession = null
      this.pendingSession = true
      this.messages = []
      this._historyLoaded = false
      this.streamingText = ''
      this.streamingThinking = ''
      this.pendingToolCalls = {}
      this.tokenStats = null
      this.compacting = false
      this._assistantPushedInTurn = false
    },
    selectSession(session) {
      if (this.currentSessionId === session.id) return
      this.disconnectWS()
      this.currentSessionId = session.id
      this.currentSession = session
      this.pendingSession = false
      this.messages = []
      this.streamingText = ''
      this.streamingThinking = ''
      this.pendingToolCalls = {}
      this.tokenStats = null
      this.compacting = false
      this._assistantPushedInTurn = false
      this._historyLoaded = false // 标记：HTTP API 是否已加载了历史消息
      this.historyLoading = true
      // 恢复该会话最后使用的模型
      if (session.model_name && this.providerModels.length > 0) {
        this.restoreSessionModel(session.model_name)
      }
      this.loadSessionMessages()
      this.connectWS()
    },
    deleteSession(session) {
      this.$confirm('确定删除此对话？', '提示', { type: 'warning' }).then(() => {
        Base.BasePost('/api/AgentV2SessionDelete', { id: session.id }, () => {
          if (this.currentSessionId === session.id) {
            this.disconnectWS()
            this.currentSessionId = 0
            this.currentSession = null
            this.messages = []
          }
          this.loadSessions()
        })
      }).catch(() => {})
    },
    showSessionMenu(event, session) {
      this.renameForm = { id: session.id, name: session.name }
      this.showRenameDialog = true
    },
    renameSession() {
      Base.BasePost('/api/AgentV2SessionRename', {
        id: this.renameForm.id,
        name: this.renameForm.name
      }, () => {
        this.showRenameDialog = false
        this.loadSessions()
        if (this.currentSession && this.currentSession.id === this.renameForm.id) {
          this.currentSession.name = this.renameForm.name
        }
      })
    },
    loadSessionMessages() {
      const sessionId = this.currentSessionId
      Base.BasePost('/api/AgentV2SessionMessages', { session_id: sessionId }, (res) => {
        // 防止竞态：仅当请求的会话仍是当前选中会话时才设置消息
        if (this.currentSessionId !== sessionId) return
        this.historyLoading = false
        if (res.ErrCode === 0 && res.Data && res.Data.messages) {
          if (res.Data.messages.length > 0) {
            this.messages = res.Data.messages
          }
          this._historyLoaded = true
          this.scrollToBottom()
        }
      })
    },

    // ========== Agent 信息 ==========
    loadAgentInfo() {
      Base.BasePost('/api/AgentV2List', {}, (res) => {
        const agents = (res.ErrCode === 0 && res.Data && res.Data.list) ? res.Data.list : []
        const agent = agents.find(a => a.id === this.agentId)
        if (agent) {
          this.agentName = agent.name
          if (agent.config) {
            try {
              this.agentConfig = JSON.parse(agent.config)
            } catch(e) {}
          }
        }
        // Agent 信息加载完成后加载模型列表（确保 agentConfig 已就绪）
        this.loadProviderModels()
      })
    },
    loadProviderModels() {
      Base.BasePost('/api/AgentV2ProviderModels', {}, (res) => {
        if (res.ErrCode === 0 && res.Data && res.Data.providers) {
          this.providerModels = res.Data.providers
            .filter(p => (p.models || []).length > 0)
            .map(p => ({
              provider_id: p.id,
              provider_name: p.name,
              provider_type: p.provider_type,
              models: p.models || []
            }))
          // 设置默认选中模型
          if (this.providerModels.length > 0 && !this.selectedModel) {
            const cfg = this.agentConfig || {}
            if (cfg.provider_id && cfg.model_id) {
              for (const g of this.providerModels) {
                if (g.provider_id === cfg.provider_id) {
                  const m = g.models.find(m => m.id === cfg.model_id)
                  if (m) {
                    this.selectedModel = g.provider_name + '/' + m.model
                    break
                  }
                }
              }
            }
            if (!this.selectedModel) {
              const first = this.providerModels[0]
              const firstModel = first.models[0]
              this.selectedModel = first.provider_name + '/' + firstModel.model
            }
          }
        }
      })
    },

    // ========== WebSocket ==========
    connectWS() {
      if (!this.currentSessionId) return
      const apiHost = Base.GetAbsoluteApiHost() // dev: http://localhost:17170, prod: current origin
      const protocol = apiHost.startsWith('https') ? 'wss:' : 'ws:'
      const host = apiHost.replace(/^https?:\/\//, '')
      const token = Base.GetSafeToken() || ''
      const modelParam = this.selectedModel ? `&model=${encodeURIComponent(this.selectedModel)}` : ''
      const url = `${protocol}//${host}/api/AgentV2WS?agent_id=${this.agentId}&session_id=${this.currentSessionId}&token=${token}${modelParam}`

      this.ws = new WebSocket(url)
      this.ws.onopen = () => {
        this.wsConnected = true
        // 连接成功立即请求会话统计（上下文使用率、Token 等）
        this.requestTokenStats()
        // 懒创建模式：发送暂存的首条消息
        if (this._pendingFirstMessage) {
          const msg = this._pendingFirstMessage
          this._pendingFirstMessage = ''
          this.sendWS({
            type: 'command',
            command: { type: 'prompt', message: msg }
          })
        }
      }
      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          this.handleWSMessage(data)
        } catch (e) {
          console.error('WS message parse error:', e)
        }
      }
      this.ws.onclose = () => {
        this.wsConnected = false
        this.isStreaming = false
      }
      this.ws.onerror = (e) => {
        console.error('WS error:', e)
        this.wsConnected = false
      }
    },
    disconnectWS() {
      if (this.ws) {
        this.ws.close()
        this.ws = null
      }
      this.wsConnected = false
      this.isStreaming = false
      this.stopThinkingTimer()
      this.stopStatsPolling()
    },
    handleWSMessage(data) {
      // 忽略来自已断开/旧 WebSocket 的消息（连接被关闭后仍可能收到缓冲消息）
      if (!this.ws || this.ws.readyState !== WebSocket.OPEN) return

      if (data.type === 'event' && data.event) {
        this.handlePiEvent(data.event)
      } else if (data.type === 'state') {
        // 更新模型信息（状态中 model 是纯模型 ID，provider 是 provider 类型）
        if (data.state?.model && data.state?.provider) {
          const lookupVal = data.state.provider + '/' + data.state.model
          this.selectedModel = lookupVal
        }
      } else if (data.type === 'history' && data.messages) {
        // 如果 HTTP API 已加载历史消息，不覆盖（避免重复造成闪烁）
        if (!this._historyLoaded) {
          this.messages = data.messages
          this.scrollToBottom()
        }
      } else if (data.type === 'error') {
        this.$message.error(data.error)
      }
    },
    handlePiEvent(event) {
      const evtType = event.type

      switch (evtType) {
        // ===== 消息流式更新 =====
        case 'message_update': {
          const msgEvt = event.assistantMessageEvent || {}
          const deltaType = msgEvt.type

          if (deltaType === 'text_delta') {
            this.streamingText += (msgEvt.delta || '')
            this.scrollToBottom()
          } else if (deltaType === 'thinking_delta') {
            if (!this.thinkingStartAt) this.startThinkingTimer()
            this.streamingThinking += (msgEvt.delta || '')
            this.scrollToBottom()
          } else if (deltaType === 'text_start' || deltaType === 'text_end' ||
                     deltaType === 'thinking_start' || deltaType === 'thinking_end') {
            if (deltaType === 'thinking_start' && !this.thinkingStartAt) this.startThinkingTimer()
            if (deltaType === 'thinking_end') this.stopThinkingTimer()
            this.scrollToBottom()
          } else if (deltaType === 'toolcall_start' || deltaType === 'toolcall_delta' || deltaType === 'toolcall_end') {
            // 支持 Anthropic (msgEvt.toolCall) 和 DeepSeek/OpenAI (partial.content) 两种格式
            this.handleToolCallInMessageUpdate(msgEvt)
          }
          break
        }

        // ===== 消息生命周期 =====
        case 'message_start': {
          const msg = event.message
          if (msg && msg.role === 'user') {
            this.scrollToBottom()
          } else if (msg && msg.role === 'assistant') {
            // 新 assistant 消息开始时清理上一轮 tool_execution 残留
            this.pendingToolCalls = {}
          }
          break
        }
        case 'message_end': {
          const msg = event.message
          if (msg && msg.role === 'assistant') {
            const text = this.extractPiContent(msg.content)
            const errorMsg = msg.errorMessage || ''
            // 仅在有实际内容时才 push（与后端 reconstructMessagesFromPiEvents 一致）
            if (text || errorMsg || this.streamingThinking || Object.keys(this.pendingToolCalls).length > 0) {
              const toolCalls = Object.values(this.pendingToolCalls)
              this.messages.push({
                role: 'assistant',
                content: text || (errorMsg ? '**Error:** ' + errorMsg : ''),
                thinking: this.streamingThinking,
                thinkingDurationMs: this.getCurrentThinkingDurationMs(true),
                toolCalls: toolCalls.length > 0 ? toolCalls.map(item => ({ ...item })) : undefined
              })
              this.streamingThinking = ''
              this.streamingText = ''
              this.pendingToolCalls = {}
              this._assistantPushedInTurn = true
            }
          }
          this.scrollToBottom()
          // 每次消息结束刷新上下文统计
          this.requestTokenStats()
          break
        }

        // ===== Turn 生命周期 =====
        case 'turn_start':
          this.turnCount++
          break
        case 'turn_end':
          break

        // ===== 工具执行 =====
        case 'tool_execution_start': {
          const tcId = event.toolCallId || event.id
          if (tcId && !this.pendingToolCalls[tcId]) {
            this.pendingToolCalls[tcId] = {
              id: tcId, name: event.toolName || event.name || 'unknown',
              status: 'running', input: '', output: '', _collapsed: true,
              startedAt: Date.now(), durationMs: 0
            }
          }
          if (tcId && this.pendingToolCalls[tcId]) {
            this.pendingToolCalls[tcId].status = 'running'
            if (!this.pendingToolCalls[tcId].startedAt) {
              this.pendingToolCalls[tcId].startedAt = Date.now()
            }
            this.startRuntimeTicker()
          }
          break
        }
        case 'tool_execution_update': {
          const tcId = event.toolCallId || event.id
          if (tcId && this.pendingToolCalls[tcId]) {
            if (event.output) {
              this.pendingToolCalls[tcId].output = (this.pendingToolCalls[tcId].output || '') + event.output
            }
          }
          break
        }
        case 'tool_execution_end': {
          const tcId = event.toolCallId || event.id
          if (tcId && this.pendingToolCalls[tcId]) {
            this.pendingToolCalls[tcId].status = 'done'
            this.pendingToolCalls[tcId].durationMs = this.pendingToolCalls[tcId].startedAt ? (Date.now() - this.pendingToolCalls[tcId].startedAt) : 0
            this.pendingToolCalls[tcId].output = event.output || event.result || this.pendingToolCalls[tcId].output || ''
            // 同步更新已推送消息中的 toolCalls（让最终消息也显示执行结果）
            this.syncToolCallToMessages(tcId)
            this.stopRuntimeTickerIfIdle()
          }
          break
        }

        // ===== Agent 生命周期 =====
        case 'agent_start': {
          this.isStreaming = true
          this.streamingText = ''
          this.streamingThinking = ''
          this.pendingToolCalls = {}
          this.compacting = false
          this.stopThinkingTimer()
          this.startRuntimeTicker()
          this.startStatsPolling()
          this._assistantPushedInTurn = false
          // 将最后发送的消息展示为用户消息
          if (this._lastUserMessage) {
            this.messages.push({ role: 'user', content: this._lastUserMessage })
            this._lastUserMessage = ''
          }
          break
        }
        case 'agent_end': {
          this.isStreaming = false
          this.stopThinkingTimer()
          this.stopStatsPolling()
          // 仅在 message_end 未推送时才兜底推送（与后端 needPushAssistant 逻辑一致）
          if (!this._assistantPushedInTurn && (this.streamingText || Object.values(this.pendingToolCalls).length > 0)) {
            const toolCalls = Object.values(this.pendingToolCalls)
            this.messages.push({
              role: 'assistant',
              content: this.streamingText,
              thinking: this.streamingThinking,
              thinkingDurationMs: this.getCurrentThinkingDurationMs(true),
              toolCalls: toolCalls.length > 0 ? toolCalls.map(item => ({ ...item })) : undefined
            })
          }
          this.streamingText = ''
          this.streamingThinking = ''
          this.pendingToolCalls = {}
          this._assistantPushedInTurn = false
          this.stopRuntimeTickerIfIdle()
          this.scrollToBottom()
          // 自动获取 token 统计
          this.requestTokenStats()
          // 刷新会话列表以获取最新标题
          this.loadSessions()
          break
        }

        // ===== 压缩 =====
        case 'compaction_start':
          this.compacting = true
          this.requestTokenStats()
          break
        case 'compaction_end':
          this.compacting = false
          this.requestTokenStats()
          break

        // ===== 队列更新 =====
        case 'queue_update':
          break

        // ===== 扩展错误 =====
        case 'extension_error':
          this.$message.warning('扩展错误: ' + (event.error || event.message || '未知错误'))
          break

        // ===== 扩展 UI =====
        case 'extension_ui_request': {
          this.handleExtensionUI(event)
          break
        }

        // ===== 响应（含 get_state / get_session_stats 等） =====
        case 'response': {
          const cmd = event._command || event.command
          if (cmd === 'get_session_stats' && event.success && event.data) {
            this.tokenStats = event.data
          }
          break
        }

        default:
          console.log('[AgentChat] unhandled pi event type:', evtType, event)
          break
      }
    },

    handleExtensionUI(event) {
      const method = event.method
      const reqId = event.id

      if (method === 'confirm') {
        this.$confirm(event.message || event.title || '确认操作?', event.title || '提示', {
          confirmButtonText: '确认',
          cancelButtonText: '取消'
        }).then(() => {
          this.sendWS({ type: 'command', command: { type: 'extension_ui_response', id: reqId, confirmed: true } })
        }).catch(() => {
          this.sendWS({ type: 'command', command: { type: 'extension_ui_response', id: reqId, cancelled: true } })
        })
      } else if (method === 'select') {
        // 使用 options 列表弹出选择
        const options = event.options || []
        if (options.length === 0) {
          this.sendWS({ type: 'command', command: { type: 'extension_ui_response', id: reqId, cancelled: true } })
          return
        }
        this.$msgbox({
          title: event.title || '选择',
          message: '请选择一项操作',
          showCancelButton: true,
          confirmButtonText: options[0],
          cancelButtonText: options.length > 1 ? options[options.length - 1] : '取消',
          distinguishCancelAndClose: true
        }).then(() => {
          this.sendWS({ type: 'command', command: { type: 'extension_ui_response', id: reqId, value: options[0] } })
        }).catch((action) => {
          if (action === 'cancel' && options.length > 1) {
            this.sendWS({ type: 'command', command: { type: 'extension_ui_response', id: reqId, value: options[options.length - 1] } })
          } else {
            this.sendWS({ type: 'command', command: { type: 'extension_ui_response', id: reqId, cancelled: true } })
          }
        })
      } else if (method === 'input') {
        this.$prompt(event.title || '请输入', '输入', {
          confirmButtonText: '确认',
          cancelButtonText: '取消',
          inputValue: event.prefill || ''
        }).then(({ value }) => {
          this.sendWS({ type: 'command', command: { type: 'extension_ui_response', id: reqId, value: value } })
        }).catch(() => {
          this.sendWS({ type: 'command', command: { type: 'extension_ui_response', id: reqId, cancelled: true } })
        })
      } else {
        // notify / setStatus / setWidget / setTitle / set_editor_text 无需响应
        if (method === 'notify') {
          this.$message.info(event.message || event.title || '')
        }
      }
    },

    // ========== 工具调用辅助方法 ==========
    handleToolCallInMessageUpdate(msgEvt) {
      // 格式1: Anthropic — msgEvt.toolCall 直接携带工具调用信息
      const tcDirect = msgEvt.toolCall
      if (tcDirect && tcDirect.id) {
        if (!this.pendingToolCalls[tcDirect.id]) {
          this.pendingToolCalls[tcDirect.id] = { id: tcDirect.id, name: tcDirect.name || 'unknown', status: 'running', input: '', output: '', _collapsed: true, startedAt: Date.now(), durationMs: 0 }
        }
        if (tcDirect.arguments) {
          try { this.pendingToolCalls[tcDirect.id].input = JSON.parse(tcDirect.arguments) } catch(e) {
            this.pendingToolCalls[tcDirect.id].input = tcDirect.arguments
          }
        }
      }
      // 格式2: DeepSeek/OpenAI — partial.content 数组中的 toolCall 块
      const partialContent = (msgEvt.partial && msgEvt.partial.content) || []
      for (const block of partialContent) {
        if (block.type === 'toolCall' && block.id) {
          if (!this.pendingToolCalls[block.id]) {
            this.pendingToolCalls[block.id] = { id: block.id, name: block.name || 'unknown', status: 'running', input: '', output: '', _collapsed: true, startedAt: Date.now(), durationMs: 0 }
          }
          // arguments（完整参数对象或 JSON 字符串）
          const args = block.arguments
          if (args !== undefined && args !== null) {
            if (typeof args === 'string' && args) {
              try { this.pendingToolCalls[block.id].input = JSON.parse(args) } catch(e) { this.pendingToolCalls[block.id].input = args }
            } else if (typeof args === 'object' && (Array.isArray(args) || Object.keys(args).length > 0)) {
              this.pendingToolCalls[block.id].input = args
            }
          }
          // partialArgs（流式参数字符串，可能是不完整 JSON）
          if (block.partialArgs && (!args || (typeof args === 'object' && Object.keys(args).length === 0))) {
            try { this.pendingToolCalls[block.id].input = JSON.parse(block.partialArgs) } catch(e) {
              this.pendingToolCalls[block.id].input = block.partialArgs
            }
          }
          // toolcall_end 时标记参数收集完毕
          if (msgEvt.type === 'toolcall_end' && this.pendingToolCalls[block.id]) {
            this.pendingToolCalls[block.id]._inputFinalized = true
          }
        }
      }
    },

    syncToolCallToMessages(tcId) {
      // 从后往前找到最近的包含此 toolCall 的助手消息，同步 status/output
      const tc = this.pendingToolCalls[tcId]
      if (!tc) return
      for (let i = this.messages.length - 1; i >= 0; i--) {
        const msg = this.messages[i]
        if (msg.role === 'assistant' && msg.toolCalls) {
          const msgTc = msg.toolCalls.find(t => t.id === tcId)
          if (msgTc) {
            msgTc.status = tc.status
            msgTc.output = tc.output
            break
          }
        }
      }
    },
    toggleMessageThinking(msg) {
      msg._thinkingCollapsed = !this.isMessageThinkingCollapsed(msg)
    },
    startThinkingTimer() {
      if (!this.thinkingStartAt) {
        this.thinkingStartAt = Date.now()
        this.thinkingElapsedSeconds = 0
      }
      this.startRuntimeTicker()
    },
    stopThinkingTimer() {
      this.thinkingStartAt = 0
      this.thinkingElapsedSeconds = 0
      this.stopRuntimeTickerIfIdle()
    },
    startRuntimeTicker() {
      if (this._runtimeTicker) return
      this.runtimeNow = Date.now()
      this._runtimeTicker = setInterval(() => {
        this.runtimeNow = Date.now()
        if (this.thinkingStartAt) {
          this.thinkingElapsedSeconds = Math.max(0, Math.floor((this.runtimeNow - this.thinkingStartAt) / 1000))
        }
      }, 200)
    },
    stopRuntimeTicker() {
      if (this._runtimeTicker) {
        clearInterval(this._runtimeTicker)
        this._runtimeTicker = null
      }
    },
    stopRuntimeTickerIfIdle() {
      const hasRunningTool = Object.values(this.pendingToolCalls).some(tc => this.isToolRunning(tc))
      if (!this.thinkingStartAt && !hasRunningTool && !this.isStreaming) {
        this.stopRuntimeTicker()
      }
    },
    isThinkingRunning(msg) {
      return Boolean(msg && Number(msg.thinkingDurationMs || 0) <= 0)
    },
    getCurrentThinkingDurationMs(finalize = false) {
      if (!this.thinkingStartAt) return 0
      const now = finalize ? Date.now() : this.runtimeNow
      return Math.max(0, now - this.thinkingStartAt)
    },
    getThinkingDurationText(msg) {
      const durationMs = Number(msg?.thinkingDurationMs || 0)
      return this.formatDuration(durationMs)
    },
    getStreamingThinkingDurationText() {
      return this.formatDuration(this.getCurrentThinkingDurationMs(false))
    },
    isToolRunning(tc) {
      return tc && tc.status === 'running'
    },
    isToolDone(tc) {
      return tc && tc.status === 'done'
    },
    getToolDurationText(tc) {
      if (!tc) return '0s'
      const durationMs = Number(tc.durationMs || 0) > 0
        ? Number(tc.durationMs || 0)
        : (tc.startedAt ? Math.max(0, this.runtimeNow - tc.startedAt) : 0)
      return this.formatDuration(durationMs)
    },
    formatDuration(durationMs) {
      const ms = Number(durationMs || 0)
      const totalSeconds = Math.max(0, Math.floor(ms / 1000))
      const minutes = Math.floor(totalSeconds / 60)
      const seconds = totalSeconds % 60
      if (minutes > 0) return `${minutes}m${seconds}s`
      return `${seconds}s`
    },
    // 恢复会话上次使用的模型
    restoreSessionModel(modelName) {
      if (!modelName) return
      for (const group of this.providerModels) {
        const m = group.models.find(m => m.model === modelName)
        if (m) {
          this.selectedModel = group.provider_name + '/' + m.model
          return
        }
      }
    },
    // 定时刷新上下文统计（流式执行中每 5 秒更新）
    startStatsPolling() {
      if (this._statsPollTimer) return
      this._statsPollTimer = setInterval(() => {
        this.requestTokenStats()
      }, 5000)
    },
    stopStatsPolling() {
      if (this._statsPollTimer) {
        clearInterval(this._statsPollTimer)
        this._statsPollTimer = null
      }
    },
    isMessageThinkingCollapsed(msg) {
      return msg._thinkingCollapsed !== false
    },

    // ========== Skills & Token 统计 ==========
    tStat(key) {
      if (!this.tokenStats) return '--'
      return this.fmtNum(this.tokenStats[key] || 0)
    },
    loadSkills() {
      Base.BasePost('/api/AgentV2SkillList', { agent_id: this.agentId }, (res) => {
        if (res.ErrCode === 0 && res.Data && res.Data.list) {
          this.skills = res.Data.list
        }
      })
    },
    toggleSkill(sk) {
      const idx = this.selectedSkillIds.indexOf(sk.id)
      if (idx >= 0) {
        this.selectedSkillIds.splice(idx, 1)
      } else {
        this.selectedSkillIds.push(sk.id)
        // 将 skill 名称追加到输入框
        if (this.inputText && !this.inputText.endsWith(' ')) {
          this.inputText += ' '
        }
        this.inputText += 'use skill "' + sk.name + '"'
      }
    },

    // ========== 发送消息 ==========
    sendMessage() {
      const text = this.inputText.trim()
      if (!text || this.isStreaming) return

      // 保存最后发送的消息文本（agent_start 时用于展示用户消息）
      this._lastUserMessage = text
      this.inputText = ''

      // 懒创建模式：先暂存消息，等会话创建+WS 连接成功后再发送
      if (this.pendingSession && !this.currentSessionId) {
        this._pendingFirstMessage = text
        this.createRealSessionAndSend()
        return
      }

      if (!this.wsConnected) return

      this.sendWS({
        type: 'command',
        command: { type: 'prompt', message: text }
      })
    },
    createRealSessionAndSend() {
      Base.BasePost('/api/AgentV2SessionCreate', {
        agent_id: this.agentId,
        workspace_id: this.currentWorkspaceId,
        name: new Date().toLocaleString()
      }, (res) => {
        const newId = (res.ErrCode === 0 && res.Data) ? res.Data.id : null
        if (!newId) {
          this.$message.error('创建会话失败')
          this.pendingSession = false
          return
        }
        // 添加到会话列表
        const newSession = {
          id: newId,
          agent_id: this.agentId,
          workspace_id: this.currentWorkspaceId,
          name: new Date().toLocaleString(),
          updated_at: Math.floor(Date.now() / 1000)
        }
        this.sessions.unshift(newSession)
        this.currentSessionId = newId
        this.currentSession = newSession
        this.pendingSession = false
        this._historyLoaded = false
        // 建立 WebSocket（onopen 中会发送 _pendingFirstMessage）
        this.connectWS()
      })
    },
    sendWS(data) {
      if (this.ws && this.ws.readyState === WebSocket.OPEN) {
        this.ws.send(JSON.stringify(data))
      }
    },
    abortAgent() {
      this.sendWS({ type: 'command', command: { type: 'abort' } })
    },
    setModel() {
      if (!this.selectedModel) return
      // 运行中不允许切换（下拉框已 disabled，此处兜底）
      if (this.isStreaming) return
      // 切换模型：断开 WS 并重连，后端重启 Pi 进程以新模型启动（100%生效）
      // 仅在已有活跃会话时重连；否则只更新 selectedModel，下次连接时自动使用
      if (this.currentSessionId && this.wsConnected) {
        this.disconnectWS()
        this.streamingText = ''
        this.streamingThinking = ''
        this.pendingToolCalls = {}
        this.compacting = false
        this._assistantPushedInTurn = false
        this._historyLoaded = true // 已加载的历史不覆盖
        this.connectWS()
      }
    },
    requestTokenStats() {
      this.sendWS({ type: 'get_session_stats' })
    },

    // ========== 渲染 ==========
    renderContent(msg) {
      if (msg.role === 'tool') {
        let html = '<div class="tool-result">'
        html += '<strong>🔧 ' + (msg.tool_name || 'Tool') + '</strong>'
        if (msg.tool_output) {
          html += '<pre>' + this.escapeHtml(typeof msg.tool_output === 'string' ? msg.tool_output : JSON.stringify(msg.tool_output, null, 2)) + '</pre>'
        }
        html += '</div>'
        return html
      }
      let content = msg.content || ''

      // 检测模型 API 错误，友好格式化（避免显示原始 JSON）
      const apiError = this.parseApiError(content)
      if (apiError) {
        content = this.renderApiError(apiError)
      } else {
        content = this.renderMarkdown(content)
      }
      return content
    },
    // 解析模型 API 返回的错误（如 403 forbidden 等）
    parseApiError(text) {
      if (!text) return null
      // 匹配 Pi Agent 错误输出格式: "Error: <status> <json>"
      const m = text.match(/^Error:\s*(\d{3})\s*(\{[\s\S]*\})/)
      if (!m) return null
      try {
        const body = JSON.parse(m[2])
        const status = parseInt(m[1])
        const detail = (body.error && body.error.message) || body.message || ''
        return { status, detail }
      } catch (e) {
        // JSON 解析失败，返回基本错误信息
        return null
      }
    },
    // 渲染 API 错误为友好的 HTML
    renderApiError(err) {
      const statusLabel = {
        400: '请求参数错误',
        401: '认证失败',
        403: '请求被拒绝',
        404: '资源不存在',
        429: '请求过于频繁',
        500: '服务器内部错误',
        502: '网关错误',
        503: '服务暂不可用'
      }
      const label = statusLabel[err.status] || 'HTTP ' + err.status + ' 错误'

      let html = '<div class="api-error">'
      html += '<div class="api-error__header"><span class="api-error__icon">⚠</span>'
      html += '<span class="api-error__code">' + label + ' (' + err.status + ')</span></div>'
      html += '<div class="api-error__message">' + this.escapeHtml(err.detail || '') + '</div>'
      html += '<div class="api-error__hint">请检查 API Key、模型配置或网络连接</div>'
      html += '</div>'
      return html
    },
    renderMarkdown(text) {
      if (!text) return ''
      try {
        const raw = marked.parse(text, { breaks: true })
        return DOMPurify.sanitize(raw)
      } catch (e) {
        return this.escapeHtml(text)
      }
    },
    escapeHtml(text) {
      if (!text) return ''
      return text
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;')
    },
    extractPiContent(content) {
      if (!content || !Array.isArray(content)) return ''
      return content
        .filter(block => block.type === 'text')
        .map(block => block.text || '')
        .join('')
    },
    formatJSON(obj) {
      if (!obj) return ''
      return typeof obj === 'string' ? obj : JSON.stringify(obj, null, 2)
    },
    statusLabel(status) {
      const map = { running: '执行中', done: '完成', pending: '等待' }
      return map[status] || status
    },
    // read/bash/edit 紧凑展示辅助方法
    isReadOrBashTool(tc) {
      return tc.name === 'read' || tc.name === 'read_file' || tc.name === 'bash'
        || tc.name === 'edit' || tc.name === 'write' || tc.name === 'write_file'
    },
    getCompactText(tc) {
      if (!tc.input) return ''
      let obj = tc.input
      if (typeof obj === 'string') {
        try { obj = JSON.parse(obj) } catch(e) { return obj }
      }
      if (typeof obj !== 'object' || obj === null) return String(obj)
      if (tc.name === 'read' || tc.name === 'read_file' || tc.name === 'edit' || tc.name === 'write' || tc.name === 'write_file') {
        return obj.path || obj.file_path || JSON.stringify(obj)
      }
      if (tc.name === 'bash') {
        return obj.command || JSON.stringify(obj)
      }
      return JSON.stringify(obj)
    },
    isToolCallCollapsed(tc) {
      return tc._collapsed !== false // 默认收起
    },
    toggleToolCallCollapse(tc) {
      tc._collapsed = this.isToolCallCollapsed(tc) ? false : true
    },
    scrollToBottom() {
      this.$nextTick(() => {
        const el = this.$refs.messagesContainer
        if (el) el.scrollTop = el.scrollHeight
      })
    },
    formatTime(ts) {
      if (!ts) return ''
      const d = new Date(ts * 1000)
      return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    },
    fmtNum(n) {
      if (!n) return '0'
      if (n > 1000000) return (n / 1000000).toFixed(1) + 'M'
      if (n > 1000) return (n / 1000).toFixed(1) + 'K'
      return String(n)
    },
    fmtCost(c) {
      if (c === undefined || c === null) return '0'
      return Number(c).toFixed(4)
    }
  }
}
</script>

<style scoped>
.agent-chat {
  display: flex; height: calc(100vh - 40px); background: #f8f9fa;
}

.chat-sidebar {
  width: 300px; min-width: 300px; background: #fff; border-right: 1px solid #e4e7ed;
  display: flex; flex-direction: column;
}
.chat-sidebar__header {
  padding: 12px 16px; border-bottom: 1px solid #ebeef5;
}
.chat-sidebar__section {
  padding: 12px 0; border-bottom: 1px solid #ebeef5;
}
.chat-sidebar__section--grow { flex: 1; overflow-y: auto; }
.chat-sidebar__section-title {
  display: flex; justify-content: space-between; align-items: center;
  padding: 0 16px 8px; font-size: 12px; color: #909399; text-transform: uppercase; letter-spacing: .5px;
}
.chat-sidebar__footer {
  padding: 12px 16px; border-top: 1px solid #ebeef5;
  display: flex; justify-content: space-between; align-items: center;
}

.workspace-list { padding: 0 8px; }
.workspace-item {
  display: flex; align-items: center; gap: 8px; padding: 8px 12px;
  border-radius: 8px; cursor: pointer; font-size: 13px;
}
.workspace-item:hover { background: #f5f7fa; }
.workspace-item--active { background: #ecf5ff; color: #409eff; }
.workspace-item__name { font-weight: 500; }
.workspace-item__path { font-size: 11px; color: #c0c4cc; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }

.session-list { padding: 0 8px; }
.session-item {
  display: flex; align-items: center; gap: 8px; padding: 10px 12px;
  border-radius: 8px; cursor: pointer; font-size: 13px;
}
.session-item:hover { background: #f5f7fa; }
.session-item--active { background: #ecf5ff; }
.session-item__info { flex: 1; min-width: 0; }
.session-item__name { display: block; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.session-item__time { font-size: 11px; color: #c0c4cc; }
.session-item__del { opacity: 0; }
.session-item:hover .session-item__del { opacity: 1; }

.empty-hint { padding: 16px; text-align: center; color: #c0c4cc; font-size: 13px; }

.agent-info { display: flex; align-items: center; gap: 6px; font-size: 13px; }
.agent-info__badge {
  width: 24px; height: 24px; border-radius: 6px; display: flex; align-items: center; justify-content: center;
  font-weight: bold; font-size: 14px; color: #fff;
}
.agent-info__badge--pi { background: #667eea; }
.agent-info__status {
  font-size: 11px; padding: 2px 6px; border-radius: 4px; background: #f0f0f0; color: #909399;
}
.agent-info__status--connected { background: #e8f5e9; color: #67c23a; }

.chat-main { flex: 1; display: flex; flex-direction: column; min-width: 0; }
.chat-header {
  padding: 12px 24px; background: #fff; border-bottom: 1px solid #e4e7ed;
  display: flex; justify-content: space-between; align-items: center;
}
.chat-header__title { font-weight: 500; font-size: 15px; }
.chat-header__actions { display: flex; gap: 8px; align-items: center; }

.token-stats { display: flex; gap: 12px; font-size: 11px; color: #909399; }
.token-stats__item { white-space: nowrap; }

.chat-messages { flex: 1; overflow-y: auto; padding: 24px; }
.chat-empty {
  display: flex; flex-direction: column; align-items: center; justify-content: center;
  height: 100%; color: #c0c4cc;
}
.chat-empty__icon {
  width: 80px; height: 80px; border-radius: 20px;
  background: linear-gradient(135deg, #667eea, #764ba2);
  color: #fff; font-size: 36px; font-weight: bold;
  display: flex; align-items: center; justify-content: center; margin-bottom: 16px;
}
.chat-empty__hint { font-size: 13px; margin-top: 4px; }

.chat-message { display: flex; gap: 12px; margin-bottom: 24px; max-width: 90%; }
.chat-message--assistant { max-width: 100%; }
.chat-message--user { margin-left: auto; flex-direction: row-reverse; }

.avatar {
  width: 36px; height: 36px; border-radius: 8px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  font-weight: bold; font-size: 14px; color: #fff;
}
.avatar--user { background: #409eff; }
.avatar--assistant { background: linear-gradient(135deg, #667eea, #764ba2); }
.avatar--tool { background: #e6a23c; font-size: 16px; }

.chat-message__body { min-width: 0; }
.chat-message__content {
  background: #fff; border-radius: 12px; padding: 12px 16px;
  border: 1px solid #ebeef5; line-height: 1.6; font-size: 14px;
}
.chat-message--user .chat-message__content {
  background: #409eff; color: #fff; border-color: #409eff;
}

.thinking-block { margin-bottom: 8px; }
.thinking-block__header {
  display: flex; align-items: center; gap: 6px; padding: 6px 12px;
  background: #fef7e0; border-radius: 8px; cursor: pointer; font-size: 13px; color: #b88230;
}
.agent-status-spinner {
  display: inline-block;
  width: 12px;
  height: 12px;
  border: 1.5px solid #409eff;
  border-top-color: transparent;
  border-radius: 50%;
  animation: agent-spin 0.8s linear infinite;
  flex-shrink: 0;
}
.agent-status-spinner--large {
  width: 24px;
  height: 24px;
  border-width: 2.5px;
  margin-bottom: 12px;
}
.agent-status-check {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 14px;
  height: 14px;
  color: #67c23a;
  font-size: 14px;
  font-weight: 700;
  line-height: 1;
  flex-shrink: 0;
}
.thinking-block__arrow { transition: transform .2s; }
.thinking-block__arrow--open { transform: rotate(180deg); }
.thinking-block__content {
  background: #fffef8; border: 1px solid #faecd8; border-radius: 8px;
  padding: 12px; margin-top: 4px; font-size: 13px; color: #8c6d3a; white-space: pre-wrap; max-height: 200px; overflow-y: auto;
}

.tool-calls { margin-top: 8px; }
.api-error {
  background: #fef0f0; border: 1px solid #fde2e2; border-radius: 10px;
  padding: 14px 16px; margin-top: 4px;
}
.api-error__header {
  display: flex; align-items: center; gap: 8px; margin-bottom: 8px;
}
.api-error__icon { font-size: 18px; }
.api-error__code { font-weight: 600; font-size: 14px; color: #f56c6c; }
.api-error__message {
  color: #909399; font-size: 13px; line-height: 1.6; margin-bottom: 8px;
  white-space: pre-wrap; word-break: break-all;
}
.api-error__hint {
  font-size: 12px; color: #c0c4cc; border-top: 1px solid #fde2e2; padding-top: 8px;
}
.tool-call {
  background: #f8f9fc; border: 1px solid #e4e7ed; border-radius: 8px; padding: 10px 12px; margin-bottom: 8px;
}
.tool-call--running { border-color: #e6a23c; background: #fef7e0; }
.tool-call--done { border-color: #67c23a; }
.tool-call__header {
  display: flex; align-items: center; gap: 6px; margin-bottom: 6px; font-size: 13px;
}
.tool-call__name { font-weight: 500; }
.tool-call__input, .tool-call__output {
  background: #fff; border-radius: 4px; padding: 8px; font-size: 12px;
  max-height: 150px; overflow: auto; margin: 0; border: 1px solid #ebeef5;
}
.tool-call__output { color: #67c23a; margin-top: 6px; }

.tool-call__compact-row {
  display: flex; align-items: center; gap: 8px; cursor: pointer;
  padding: 2px 0; font-size: 13px;
}
.tool-call__compact-row:hover { background: rgba(0,0,0,.02); border-radius: 4px; }
.tool-call__compact-label {
  font-weight: 600; flex-shrink: 0; color: #409eff;
  min-width: 32px; font-size: 12px; text-transform: uppercase; letter-spacing: .3px;
}
.tool-call__compact-text {
  flex: 1; overflow: hidden; text-overflow: ellipsis; white-space: nowrap;
  color: #606266; font-family: 'Cascadia Code', 'Fira Code', 'Consolas', monospace;
  font-size: 12px; line-height: 1.5;
}
.tool-call__compact-status {
  flex-shrink: 0; font-size: 11px; color: #909399;
}
.tool-call__compact-arrow {
  flex-shrink: 0; font-size: 14px; color: #c0c4cc;
  transition: transform .2s ease;
}
.tool-call__compact-arrow--open { transform: rotate(90deg); }
.tool-call__details { margin-top: 8px; border-top: 1px dashed #e4e7ed; padding-top: 8px; }

.compaction-notice {
  text-align: center; padding: 8px; color: #e6a23c; font-size: 13px;
}

.cursor-blink { animation: blink 1s infinite; color: #667eea; }
@keyframes blink { 0%,100% { opacity: 1 } 50% { opacity: 0 } }
@keyframes agent-spin { to { transform: rotate(360deg); } }

.chat-input { padding: 12px 24px 10px; background: #fff; border-top: 1px solid #e4e7ed; }
.chat-input__wrapper { width: 100%; }
.chat-input__hint { font-size: 12px; color: #c0c4cc; margin-right: 8px; }

/* 底部工具栏 */
.chat-input__toolbar {
  display: flex; justify-content: space-between; align-items: center; margin-top: 8px; gap: 8px;
}
.chat-input__toolbar-left {
  display: flex; align-items: center; gap: 10px; flex-wrap: wrap; min-width: 0;
}
.chat-input__toolbar-right {
  display: flex; align-items: center; flex-shrink: 0;
}
.chat-input__toolbar-btn {
  font-size: 12px; color: #606266; padding: 4px 10px;
}
.chat-input__toolbar-btn:hover { color: #409eff; }
.chat-input__stat-item {
  font-size: 11px; color: #909399; white-space: nowrap; user-select: none;
}
.chat-input__stat-item strong {
  color: #606266; font-weight: 500;
}
.chat-input__stat-divider {
  margin: 0 1px; color: #dcdfe6;
}

/* Skills 弹窗 */
.skills-popover { padding: 4px 0; }
.skills-popover__list { max-height: 240px; overflow-y: auto; }
.skills-popover__item {
  display: flex; align-items: center; gap: 8px;
  padding: 8px 12px; cursor: pointer; border-radius: 4px;
  font-size: 13px; transition: background .15s;
}
.skills-popover__item:hover { background: #f5f7fa; }
.skills-popover__item--active { background: #ecf5ff; }
.skills-popover__item-name { flex: 1; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.skills-popover__empty { padding: 16px; text-align: center; color: #c0c4cc; font-size: 13px; }

:deep(.tool-result) { font-size: 13px; }
:deep(.tool-result pre) {
  background: #f5f7fa; border-radius: 6px; padding: 8px; margin-top: 4px;
  max-height: 200px; overflow: auto; font-size: 12px;
}
:deep(.chat-message__content pre) {
  background: #f5f7fa; border-radius: 6px; padding: 12px; overflow-x: auto; font-size: 13px;
}
:deep(.chat-message__content code) {
  background: #f5f7fa; padding: 2px 6px; border-radius: 4px; font-size: 13px;
}
:deep(.chat-message__content pre code) { background: none; padding: 0; }
:deep(.chat-message--user .chat-message__content pre),
:deep(.chat-message--user .chat-message__content code) {
  background: rgba(255,255,255,.2); color: #fff;
}
</style>
