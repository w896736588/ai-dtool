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
      <div class="chat-sidebar__section chat-sidebar__section--workspaces">
        <div class="chat-sidebar__section-title">
          <span>工作空间</span>
          <el-button text size="small" @click="showWorkspaceDialog = true">+</el-button>
        </div>
        <div class="workspace-list">
          <el-tooltip
            v-for="ws in workspaces"
            :key="ws.id"
            :content="ws.path"
            placement="right"
            effect="dark"
          >
            <div
              class="workspace-item"
              :class="{ 'workspace-item--active': ws.id === currentWorkspaceId }"
              @click="selectWorkspace(ws)"
            >
              <el-icon><Folder /></el-icon>
              <span class="workspace-item__name">{{ ws.name }}</span>
            </div>
          </el-tooltip>
          <div v-if="workspaces.length === 0" class="empty-hint">暂无工作空间，点击 + 添加</div>
        </div>
      </div>

      <!-- 会话列表 -->
      <div class="chat-sidebar__section chat-sidebar__section--grow">
        <div class="chat-sidebar__section-title">
          <span>对话列表</span>
          <el-button text size="small" @click="showWorkspaceDialog = true">工作空间 +</el-button>
        </div>
        <div class="session-list">
          <div
            v-for="session in groupedSessions"
            :key="session._key || session.id"
            class="session-item"
            :class="{
              'session-item--active': !session._isWorkspaceHeader && !session._isShowMore && session.id === currentSessionId,
              'workspace-group-header': session._isWorkspaceHeader,
              'workspace-group-header--active': session._isWorkspaceHeader && Number(session.workspace_id) === Number(currentWorkspaceId),
              'workspace-group-header--collapsed': session._isWorkspaceHeader && session._isCollapsed,
              'workspace-group-header--dragover': session._isWorkspaceHeader && Number(session.workspace_id) === Number(this.dragOverWorkspaceId),
              'session-item--show-more': session._isShowMore
            }"
            :draggable="session._isWorkspaceHeader"
            @dragstart="session._isWorkspaceHeader ? onWorkspaceDragStart(session, $event) : null"
            @dragover.prevent="session._isWorkspaceHeader ? onWorkspaceDragOver(session, $event) : null"
            @drop="session._isWorkspaceHeader ? onWorkspaceDrop(session, $event) : null"
            @dragend="onWorkspaceDragEnd()"
            @click="session._isShowMore ? showMoreSessions(session.workspace_id) : (session._isWorkspaceHeader ? toggleWorkspaceGroup(session) : selectSession(session))"
          >
            <template v-if="session._isWorkspaceHeader">
              <el-icon class="workspace-group-header__arrow">
                <ArrowRight v-if="session._isCollapsed" />
                <ArrowDown v-else />
              </el-icon>
              <el-icon class="workspace-group-header__folder"><Folder /></el-icon>
            </template>
            <template v-else-if="session._isShowMore">
              <el-icon class="session-item__more-icon"><ArrowDown /></el-icon>
              <span class="session-item__more-text">展示更多（还剩 {{ session._remaining }} 个）</span>
            </template>
            <span v-else-if="sessionRunningMap[session.id]" class="agent-status-spinner session-item__spinner"></span>
            <el-icon v-else><ChatDotRound /></el-icon>
            <div class="session-item__info">
              <el-tooltip v-if="session._isWorkspaceHeader" :content="session.path" placement="right" effect="dark">
                <span class="session-item__name">{{ session.name }}</span>
              </el-tooltip>
              <span v-else class="session-item__name">{{ session.name }}</span>
              <span v-if="!session._isWorkspaceHeader && !session._isShowMore && session.exec_duration_ms" class="session-item__exec">{{ fmtExecDuration(session.exec_duration_ms) }}</span>
            </div>
            <span v-if="session._isWorkspaceHeader" class="workspace-group-header__count">{{ session.count }} 个对话</span>
            <el-button v-if="!session._isWorkspaceHeader && !session._isShowMore" text size="small" class="session-item__del" @click.stop="deleteSession(session)">
              <el-icon><Close /></el-icon>
            </el-button>
            <el-button v-else-if="session._isWorkspaceHeader" text size="small" class="session-item__del session-item__add" @click.stop="createSession(session.workspace)">+</el-button>
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
          <span v-if="runningSessionCount > 0" class="agent-info__running-count">
            {{ runningSessionCount }} 运行中
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
          <span v-if="currentSession" class="chat-header__created">创建于 {{ formatTime(currentSession.created_at) }}</span>
        </div>
        <div class="chat-header__actions">
          <!-- Pi 内置计划模式已启用：显示计划图标 -->
          <el-tooltip v-if="planExtensionInstalled" content="当前处于计划-确认-执行模式" placement="bottom" effect="dark">
            <span class="plan-mode-badge" @click="scrollToPlan">📋</span>
          </el-tooltip>
          <!-- 会话执行耗时（后端计时，WS exec_progress 推送） -->
          <div class="exec-time" :class="{ 'exec-time--running': executionRunning }">
            <el-icon class="exec-time__icon"><Timer /></el-icon>
            <span class="exec-time__label">执行</span>
            <span class="exec-time__value">{{ fmtExecDuration(executionElapsedMs) }}</span>
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
      <div class="chat-messages-wrap">
        <div
          class="chat-messages"
          ref="messagesContainer"
          @scroll="handleMessagesScroll"
          @wheel.passive="handleMessagesWheel"
        >
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
                  <span v-if="msg._live && thinkingStartAt" class="agent-status-spinner"></span>
                  <span v-else class="agent-status-check">✓</span>
                  <span>思考过程<template v-if="msg._live && thinkingStartAt">（{{ getStreamingThinkingDurationText() }}）</template></span>
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
                  :class="['tool-call--' + tc.status, { 'tool-call--expanded': !isToolCallCollapsed(tc) }]"
                >
                  <!-- read/bash/edit 紧凑一行 -->
                  <template v-if="isReadOrBashTool(tc)">
                    <div class="tool-call__compact-row" @click="toggleToolCallCollapse(tc)">
                      <span v-if="isToolRunning(tc)" class="agent-status-spinner"></span>
                      <span v-else-if="isToolDone(tc)" class="agent-status-check">✓</span>
                      <span v-if="toolMeta(tc.name).found" class="tool-call__icon" :style="{ background: toolMeta(tc.name).bg, color: toolMeta(tc.name).color }">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" v-html="toolMeta(tc.name).svg"></svg>
                      </span>
                      <span v-if="!toolMeta(tc.name).found" class="tool-call__compact-label">{{ tc.name }}</span>
                      <span class="tool-call__compact-text" :title="getCompactText(tc)">{{ getCompactText(tc) }}</span>
                      <span v-if="!isToolDone(tc)" class="tool-call__compact-status">{{ statusLabel(tc.status) }}<template v-if="!isToolDone(tc)">（{{ getToolDurationText(tc) }}）</template></span>
                      <el-icon class="tool-call__compact-arrow" :class="{ 'tool-call__compact-arrow--open': !isToolCallCollapsed(tc) }">
                        <ArrowRight />
                      </el-icon>
                    </div>
                    <div v-if="!isToolCallCollapsed(tc)" class="tool-call__details">
                      <pre class="tool-call__input" v-if="tc.input">{{ formatJSON(tc.input) }}</pre>
                      <pre class="tool-call__output" v-if="tc.output">{{ formatToolOutput(tc.output) }}</pre>
                    </div>
                  </template>
                  <!-- 其他工具完整展示 -->
                  <template v-else>
                    <div class="tool-call__header">
                      <span v-if="isToolRunning(tc)" class="agent-status-spinner"></span>
                      <span v-else-if="isToolDone(tc)" class="agent-status-check">✓</span>
                      <span v-if="toolMeta(tc.name).found" class="tool-call__icon" :style="{ background: toolMeta(tc.name).bg, color: toolMeta(tc.name).color }">
                        <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" v-html="toolMeta(tc.name).svg"></svg>
                      </span>
                      <span v-if="!toolMeta(tc.name).found" class="tool-call__name">{{ tc.name }}</span>
                      <el-tag v-if="!isToolDone(tc)" :type="tc.status === 'done' ? 'success' : tc.status === 'running' ? 'warning' : 'info'" size="small">
                        {{ statusLabel(tc.status) }}<template v-if="!isToolDone(tc)">（{{ getToolDurationText(tc) }}）</template>
                      </el-tag>
                    </div>
                    <pre class="tool-call__input" v-if="tc.input">{{ formatJSON(tc.input) }}</pre>
                    <pre class="tool-call__output" v-if="tc.output">{{ formatToolOutput(tc.output) }}</pre>
                  </template>
                </div>
              </div>
            </div>
          </div>

          <!-- 压缩通知 -->
          <div v-if="compacting" class="compaction-notice">
            <el-icon><Loading /></el-icon> 正在压缩上下文...
          </div>
        </div>

        <el-button
          v-show="showScrollToBottom"
          class="chat-scroll-bottom"
          circle
          title="滚动到底部"
          aria-label="滚动到底部"
          @click="scrollToBottomAndResume"
        >
          <el-icon><ArrowDown /></el-icon>
        </el-button>
      </div>

      <!-- 输入区域 -->
      <footer class="chat-input" v-if="currentSession || pendingSession">
        <!-- 计划-确认-执行 面板：紧贴输入框，方便确认或继续补充问题 -->
        <transition name="el-fade-in">
          <div v-if="planState.visible" ref="planPanel" class="plan-panel" :class="{ 'plan-panel--collapsed': planCollapsed }">
            <div class="plan-panel__header" @click="planCollapsed = !planCollapsed">
              <span class="plan-panel__icon">📋</span>
              <span class="plan-panel__title">计划与进度</span>
              <el-tag size="small" :type="planState.phase === 'plan' ? 'warning' : 'success'">
                {{ planState.phase === 'plan' ? (pendingPlanChoice ? '待确认' : '计划中') : (planState.phase === 'done' ? '已完成' : '执行中') }}
              </el-tag>
              <span class="plan-panel__progress">{{ planDoneCount }}/{{ planState.items.length }}</span>
              <span class="plan-panel__toggle">{{ planCollapsed ? '展开' : '收起' }}</span>
            </div>
            <div v-show="!planCollapsed" class="plan-panel__body">
              <ul class="plan-list">
                <li v-for="(it, i) in planState.items" :key="i" :class="{ 'plan-list__done': it.done }">
                  <span class="plan-list__check">{{ it.done ? '☑' : '☐' }}</span>
                  <span class="plan-list__text">{{ it.text }}</span>
                </li>
              </ul>
              <div v-if="planState.phase === 'plan' && planState.items.length && pendingPlanChoice" class="plan-panel__actions">
                <el-button size="small" @click="deferPlanExecution">暂不执行，继续提问</el-button>
                <el-button size="small" type="primary" @click="approvePlan">确认执行</el-button>
              </div>
            </div>
          </div>
        </transition>

        <div class="chat-input__wrapper">
          <el-input
            ref="messageInput"
            v-model="inputText"
            type="textarea"
            :rows="2"
            placeholder="输入消息，Enter 发送，Shift+Enter 换行..."
            :disabled="isStreaming || this.sending || (sessionRunningMap[currentSessionId] && !wsConnected)"
            @keydown.enter.exact.prevent="sendMessage"
            resize="none"
          />
          <div class="chat-input__toolbar">
            <div class="chat-input__toolbar-left">
              <!-- Pi 计划模式扩展：模式切换固定放在输入框最左侧 -->
              <el-dropdown
                v-if="planExtensionInstalled"
                class="chat-input__mode-select"
                :disabled="(isStreaming && !pendingPlanChoice) || switchingInteractionMode"
                trigger="click"
                placement="top-start"
                popper-class="chat-mode-menu"
                @command="setInteractionMode"
              >
                <el-button
                  size="small"
                  class="chat-input__mode-button"
                  :class="'chat-input__mode-button--' + interactionMode"
                >
                  <el-icon class="chat-input__mode-icon">
                    <Memo v-if="interactionMode === 'plan'" />
                    <VideoPlay v-else />
                  </el-icon>
                  <span>{{ interactionMode === 'plan' ? '计划' : '执行' }}</span>
                  <el-icon class="chat-input__mode-arrow"><ArrowUp /></el-icon>
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item command="execute" :class="{ 'is-active': interactionMode === 'execute' }">
                      <el-icon><VideoPlay /></el-icon>
                      <span>执行模式</span>
                    </el-dropdown-item>
                    <el-dropdown-item command="plan" :class="{ 'is-active': interactionMode === 'plan' }">
                      <el-icon><Memo /></el-icon>
                      <span>计划模式</span>
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>

              <!-- 模型选择（按 Provider 分组） -->
              <el-select
                v-if="providerModels.length > 0"
                v-model="selectedModel"
                size="small"
                class="chat-input__model-select"
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
                :loading="sending"
                :disabled="!inputText.trim() || (sessionRunningMap[currentSessionId] && !wsConnected)"
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


  </div>
</template>

<script>
import Base from '@/utils/base.js'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { getToolMeta } from '@/utils/toolIcons.js'
import {
  ArrowLeft,
  ArrowDown,
  ArrowUp,
  ArrowRight,
  Folder,
  ChatDotRound,
  Close,
  Tools,
  Loading,
  Timer,
  Memo,
  VideoPlay
} from '@element-plus/icons-vue'

export default {
  name: 'AgentChat',
  components: {
    ArrowLeft,
    ArrowDown,
    ArrowUp,
    ArrowRight,
    Folder,
    ChatDotRound,
    Close,
    Tools,
    Loading,
    Timer,
    Memo,
    VideoPlay
  },
  data() {
    return {
      agentId: 0,
      agentName: '',
      agentConfig: null,

      // 计划模式扩展
      planExtensionInstalled: false,
      defaultInteractionMode: 'execute',
      interactionMode: 'execute',
      switchingInteractionMode: false,
      pendingPlanChoice: null,
      planState: { visible: false, phase: '', items: [] },
      planCollapsed: false,


      pendingSession: false, // 新建对话标记：为 true 时不创建 DB 记录，等用户发消息后再创建

      workspaces: [],
      currentWorkspaceId: 0,
      collapsedWorkspaceMap: {},
      // 每个工作空间默认展示的会话数量（展示更多时累加 5），按 workspaceId 记录
      workspaceSessionLimit: {},
      // 工作空间拖动排序状态
      dragWorkspaceId: 0,
      dragOverWorkspaceId: 0,
      showWorkspaceDialog: false,
      workspaceForm: { name: '', path: '' },

      sessions: [],
      currentSessionId: 0,
      currentSession: null,


      messages: [],
      inputText: '',
      isStreaming: false,
      sending: false, // 发送中：等待后端确认收到用户问题（agent_start 时清空输入框并停止转圈）
      streamingText: '',
      streamingThinking: '',
      autoScrollEnabled: true,
      showScrollToBottom: false,
      programmaticScrollActive: false,

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

      // 会话执行耗时（后端计时，通过 WS exec_progress 推送）
      // executionServerMs：后端最近一次推送的累计耗时；executionSyncAt：收到推送时的本地时间（用于插值平滑）
      executionRunning: false,
      executionServerMs: 0,
      executionSyncAt: 0,
      executionElapsedMs: 0,
      _execTicker: null,

      // 加载状态
      historyLoading: false,

      // Skills 数据
      skills: [],
      selectedSkillIds: [],

      // 多会话并发：存储后台会话的状态与未回放的 WS 事件
      sessionStates: {}
    }
  },
  computed: {
    hasRunningTools() {
      return Object.values(this.pendingToolCalls).some(tc => tc.status !== 'done')
    },
    planDoneCount() {
      if (!this.planState.items) return 0
      return this.planState.items.filter((it) => it.done).length
    },
    inputHint() {
      if (!this.wsConnected) return '未连接'
      if (this.isStreaming) return 'Pi 正在思考...'
      if (this.sending) return '发送中...'
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
    },
    // 各会话的运行状态（用于侧边栏指示器）
    sessionRunningMap() {
      const map = {}
      this.sessions.forEach(session => {
        if (session && session.status === 'running') map[session.id] = true
      })
      // 当前前台会话
      if (this.currentSessionId && this.isStreaming) {
        map[this.currentSessionId] = true
      }
      // 后台会话
      Object.entries(this.sessionStates).forEach(([sid, ss]) => {
        if (ss._isRunning) map[Number(sid)] = true
      })
      return map
    },
    runningSessionCount() {
      return Object.keys(this.sessionRunningMap).length
    },
    groupedSessions() {
      const rows = []
      const byWorkspace = new Map()
      this.sessions.forEach(session => {
        const workspaceId = Number(session.workspace_id || 0)
        if (!byWorkspace.has(workspaceId)) byWorkspace.set(workspaceId, [])
        byWorkspace.get(workspaceId).push(session)
      })
      this.workspaces.forEach(ws => {
        const workspaceId = Number(ws.id)
        const collapsed = !!this.collapsedWorkspaceMap[workspaceId]
        const workspaceSessions = byWorkspace.get(workspaceId) || []
        rows.push({
          _isWorkspaceHeader: true,
          _key: 'workspace-' + ws.id,
          id: 0,
          workspace: ws,
          workspace_id: ws.id,
          name: ws.name,
          path: ws.path,
          count: workspaceSessions.length,
          _isCollapsed: collapsed
        })
        if (!collapsed) this.appendWorkspaceSessions(rows, workspaceId, workspaceSessions)
      })
      byWorkspace.forEach((items, workspaceId) => {
        if (this.workspaces.some(ws => Number(ws.id) === Number(workspaceId))) return
        const collapsed = !!this.collapsedWorkspaceMap[workspaceId]
        rows.push({
          _isWorkspaceHeader: true,
          _key: 'workspace-' + workspaceId,
          id: 0,
          workspace: { id: workspaceId, name: '未归属工作空间', path: '' },
          workspace_id: workspaceId,
          name: '未归属工作空间',
          path: '',
          count: items.length,
          _isCollapsed: collapsed
        })
        if (!collapsed) this.appendWorkspaceSessions(rows, workspaceId, items)
      })
      return rows
    },
  },
  mounted() {
      this.agentId = parseInt(this.$route.query.agent_id) || 0
    if (!this.agentId) {
      this.$router.push('/AgentHub')
      return
    }
    this.loadCollapsedState()
    this.loadAgentInfo()
    this.loadWorkspaces()
    this.loadSkills()
    // 执行耗时由后端推送，前端仅做插值平滑展示
    this.ensureExecTicker()
  },
  beforeUnmount() {
    this.disconnectAllWS()
    this.stopRuntimeTicker()
    this.stopStatsPolling()
    if (this._execTicker) {
      clearInterval(this._execTicker)
      this._execTicker = null
    }
    this.resetExecutionTimer()
  },
  methods: {
    goBack() {
      this.disconnectAllWS()
      this.$router.push('/AgentHub')
    },

    // ========== 计划模式（Pi 内置） ==========
    // 通过 Agent 启动参数判断是否启用 Pi 内置计划模式（--plan + --extension <plan-mode>），决定是否显示计划图标
    checkPlanExtension() {
      const cfg = this.agentConfig || {}
      const extra = (cfg.extra_args || '').trim()
      if (!extra) {
        this.planExtensionInstalled = false
        this.defaultInteractionMode = 'execute'
        this.interactionMode = 'execute'
        return
      }
      const hasPlan = /(?:^|\s)--plan(?:\s|$)/.test(extra)
      const hasPlanExt = /(?:^|\s)--extension\s+(?:"[^"]*plan-mode[^"]*"|'[^']*plan-mode[^']*'|\S*plan-mode\S*)/.test(extra)
      this.planExtensionInstalled = hasPlanExt
      this.defaultInteractionMode = hasPlan && hasPlanExt ? 'plan' : 'execute'
      const remembered = this.loadRememberedInteractionMode(this.currentSessionId)
      if (!this.currentSessionId || !remembered) this.interactionMode = this.defaultInteractionMode
    },
    interactionModeStorageKey() {
      return 'agentchat_interaction_modes_' + this.agentId
    },
    loadRememberedInteractionMode(sessionId) {
      if (!sessionId) return ''
      try {
        const modes = JSON.parse(localStorage.getItem(this.interactionModeStorageKey()) || '{}')
        return modes[String(sessionId)] === 'plan' ? 'plan' : (modes[String(sessionId)] === 'execute' ? 'execute' : '')
      } catch (e) {
        return ''
      }
    },
    rememberInteractionMode() {
      if (!this.currentSessionId) return
      try {
        const key = this.interactionModeStorageKey()
        const modes = JSON.parse(localStorage.getItem(key) || '{}')
        modes[String(this.currentSessionId)] = this.interactionMode
        localStorage.setItem(key, JSON.stringify(modes))
      } catch (e) {
        // localStorage 不可用时仅保留当前页面内的选择。
      }
    },
    // 用户确认执行计划：直接回复 Pi 扩展的 select 请求，避免用自然语言模拟协议。
    approvePlan() {
      if (!this.planState.items.length || !this.pendingPlanChoice) return
      this.changeInteractionMode('execute')
    },
    // 暂不执行：正式选择 Stay in plan mode，解除扩展等待并允许继续提问/改进计划。
    deferPlanExecution() {
      if (!this.pendingPlanChoice) return
      this.changeInteractionMode('plan')
    },
    releasePlanInputWait() {
      this.isStreaming = false
      this.sending = false
      this.switchingInteractionMode = false
      this.stopThinkingTimer()
      this.stopStatsPolling()
      this.stopExecutionTimer()
      this.stopRuntimeTickerIfIdle()
      this.$nextTick(() => {
        if (this.$refs.messageInput && typeof this.$refs.messageInput.focus === 'function') {
          this.$refs.messageInput.focus()
        }
      })
    },
    setInteractionMode(mode) {
      this.changeInteractionMode(mode)
    },
    changeInteractionMode(mode) {
      const targetMode = mode === 'plan' ? 'plan' : 'execute'
      const pending = this.pendingPlanChoice

      // 计划生成后扩展正在等待选择，直接回复原始 UI 请求。
      if (pending) {
        const option = targetMode === 'execute'
          ? pending.options.find((item) => String(item).startsWith('Execute'))
          : pending.options.find((item) => String(item).startsWith('Stay'))
        if (option) {
          this.sendWS({
            type: 'command',
            command: { type: 'extension_ui_response', id: pending.id, value: option }
          })
          this.pendingPlanChoice = null
          this.interactionMode = targetMode
          this.rememberInteractionMode()
          this.planState = {
            ...this.planState,
            visible: this.planState.items.length > 0,
            phase: targetMode === 'execute' ? 'execute' : 'plan'
          }
          if (targetMode === 'plan') {
            this.releasePlanInputWait()
          }
          return
        }
      }

      if (targetMode === this.interactionMode) return
      this.interactionMode = targetMode
      this.rememberInteractionMode()
      if (!this.wsConnected) {
        if (targetMode === 'plan') {
          this.planState = { visible: false, phase: 'plan', items: [] }
        } else if (this.planState.phase === 'plan') {
          this.planState = { ...this.planState, visible: false, phase: '' }
        }
        return
      }
      // /plan 是扩展提供的正式切换命令；后端会保持它位于消息首部。
      this.switchingInteractionMode = true
      this.sendWS({
        type: 'command',
        command: { type: 'prompt', message: '/plan' }
      })
      if (targetMode === 'plan') {
        this.planState = { visible: false, phase: 'plan', items: [] }
      } else if (this.planState.phase === 'plan') {
        this.planState = { ...this.planState, visible: false, phase: '' }
      }
      window.setTimeout(() => {
        this.switchingInteractionMode = false
      }, 500)
    },
    extractPlanItemsFromMessages() {
      const assistant = [...this.messages].reverse().find((msg) => msg && msg.role === 'assistant' && msg.content)
      if (!assistant) return []
      const lines = String(assistant.content).split(/\r?\n/)
      const items = []
      for (const line of lines) {
        const match = line.match(/^\s*\d+[.)]\s+(.+?)\s*$/)
        if (match) items.push({ text: match[1], done: false })
      }
      return items
    },
    scrollToPlan() {
      this.planCollapsed = false
      this.$nextTick(() => {
        if (this.$refs.planPanel && this.$refs.planPanel.scrollIntoView) {
          this.$refs.planPanel.scrollIntoView({ behavior: 'smooth', block: 'nearest' })
        }
      })
    },

    openConfig() {
      this.$router.push({ path: '/AgentConfig', query: { agent_id: this.agentId } })
    },

    // ========== 工作空间 ==========
    async loadWorkspaces() {
      Base.BasePost('/api/AgentV2WorkspaceList', { agent_id: this.agentId }, (res) => {
        const list = (res.ErrCode === 0 && res.Data && res.Data.list) ? res.Data.list : []
        this.workspaces = this.applyWorkspaceOrder(list)
        if (!this.currentWorkspaceId && this.workspaces.length > 0) {
          this.currentWorkspaceId = this.workspaces[0].id
        }
        this.loadSessions()
      })
    },
    selectWorkspace(ws) {
      this.currentWorkspaceId = ws.id
    },
    toggleWorkspaceGroup(session) {
      if (!session || !session.workspace) return
      this.currentWorkspaceId = session.workspace.id
      const workspaceId = Number(session.workspace_id || session.workspace.id || 0)
      this.collapsedWorkspaceMap = {
        ...this.collapsedWorkspaceMap,
        [workspaceId]: !this.collapsedWorkspaceMap[workspaceId]
      }
      this.saveCollapsedState()
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
    // 展开的工作空间会话追加到列表（按更新时间倒序，默认 5 个，超出显示展示更多）
    appendWorkspaceSessions(rows, workspaceId, sessions) {
      const wsId = Number(workspaceId)
      const sorted = sessions.slice().sort((a, b) => (Number(b.updated_at) || 0) - (Number(a.updated_at) || 0))
      const limit = this.workspaceSessionLimit[wsId] || 5
      rows.push(...sorted.slice(0, limit))
      if (sorted.length > limit) {
        rows.push({
          _isShowMore: true,
          _key: 'showmore-' + wsId,
          id: 0,
          workspace_id: wsId,
          _remaining: sorted.length - limit
        })
      }
    },
    showMoreSessions(workspaceId) {
      const wsId = Number(workspaceId)
      const current = this.workspaceSessionLimit[wsId] || 5
      this.workspaceSessionLimit = {
        ...this.workspaceSessionLimit,
        [wsId]: current + 5
      }
    },
    collapsedStorageKey() {
      return 'agentchat_collapsed_ws_' + this.agentId
    },
    loadCollapsedState() {
      try {
        const raw = localStorage.getItem(this.collapsedStorageKey())
        this.collapsedWorkspaceMap = raw ? JSON.parse(raw) : {}
      } catch (e) {
        this.collapsedWorkspaceMap = {}
      }
    },
    saveCollapsedState() {
      try {
        localStorage.setItem(this.collapsedStorageKey(), JSON.stringify(this.collapsedWorkspaceMap))
      } catch (e) {
        // localStorage 不可用时静默忽略
      }
    },

    // ========== 工作空间拖动排序 ==========
    workspaceOrderKey() {
      return 'agentchat_ws_order_' + this.agentId
    },
    loadWorkspaceOrder() {
      try {
        const raw = localStorage.getItem(this.workspaceOrderKey())
        return raw ? JSON.parse(raw) : []
      } catch (e) {
        return []
      }
    },
    saveWorkspaceOrder() {
      try {
        localStorage.setItem(this.workspaceOrderKey(), JSON.stringify(this.workspaces.map(w => Number(w.id))))
      } catch (e) {
        // localStorage 不可用时静默忽略
      }
    },
    applyWorkspaceOrder(list) {
      const order = this.loadWorkspaceOrder()
      if (!order || !order.length) return list
      const rank = (id) => {
        const i = order.indexOf(Number(id))
        return i < 0 ? order.length : i
      }
      return list.slice().sort((a, b) => rank(a.id) - rank(b.id))
    },
    onWorkspaceDragStart(session, event) {
      this.dragWorkspaceId = Number(session.workspace_id)
      if (event.dataTransfer) {
        event.dataTransfer.effectAllowed = 'move'
        event.dataTransfer.setData('text/plain', String(this.dragWorkspaceId))
      }
    },
    onWorkspaceDragOver(session) {
      this.dragOverWorkspaceId = Number(session.workspace_id)
    },
    onWorkspaceDrop(session) {
      const fromId = this.dragWorkspaceId
      const toId = Number(session.workspace_id)
      this.dragOverWorkspaceId = 0
      if (!fromId || fromId === toId) return
      this.reorderWorkspace(fromId, toId)
    },
    onWorkspaceDragEnd() {
      this.dragWorkspaceId = 0
      this.dragOverWorkspaceId = 0
    },
    reorderWorkspace(fromId, toId) {
      const list = this.workspaces.slice()
      const fromIdx = list.findIndex(w => Number(w.id) === Number(fromId))
      const toIdx = list.findIndex(w => Number(w.id) === Number(toId))
      if (fromIdx < 0 || toIdx < 0) return
      const [moved] = list.splice(fromIdx, 1)
      list.splice(toIdx, 0, moved)
      this.workspaces = list
      this.saveWorkspaceOrder()
      // 同步后端排序数据（sort_order）
      Base.BasePost('/api/AgentV2WorkspaceReorder', {
        agent_id: this.agentId,
        ordered_ids: list.map(w => Number(w.id))
      })
    },
    // 对话执行时把当前工作空间置顶展示，并同步到后端排序数据
    bumpWorkspaceToFront(workspaceId) {
      const id = Number(workspaceId)
      if (!id || !this.workspaces.length) return
      // 已在首位则无需调整
      if (Number(this.workspaces[0].id) === id) return
      const idx = this.workspaces.findIndex(w => Number(w.id) === id)
      if (idx < 0) return
      const next = this.workspaces.slice()
      const [moved] = next.splice(idx, 1)
      next.unshift(moved)
      this.workspaces = next
      this.saveWorkspaceOrder()
      Base.BasePost('/api/AgentV2WorkspaceReorder', {
        agent_id: this.agentId,
        ordered_ids: next.map(w => Number(w.id))
      })
    },

    // ========== 会话管理 ==========
    async loadSessions() {
      Base.BasePost('/api/AgentV2SessionList', { agent_id: this.agentId }, (res) => {
        this.sessions = (res.ErrCode === 0 && res.Data) ? (res.Data.list || []) : []
        this.attachRunningSessions()
      })
    },
    createSession(workspace) {
      if (workspace && workspace.id) this.currentWorkspaceId = workspace.id
      if (!this.currentWorkspaceId) return
      // 仅打开空白聊天区，不创建 DB 记录、不连 WebSocket
      // 等用户输入第一条消息时才真正创建会话
      // 保存当前会话状态（不关闭后台 WS，保持并发执行）
      this.saveCurrentSession()
      this.currentSessionId = 0
      this.currentSession = null
      this.pendingSession = true
      this.interactionMode = this.defaultInteractionMode
      this.pendingPlanChoice = null
      this.planState = { visible: false, phase: '', items: [] }
      this.planCollapsed = false
      this.switchingInteractionMode = false
      this.messages = []
      this._historyLoaded = false
      this.streamingText = ''
      this.streamingThinking = ''
      this.pendingToolCalls = {}
      this.tokenStats = null
      this.compacting = false
      this._assistantPushedInTurn = false
      this.isStreaming = false
      this.sending = false
      this.resetMessageAutoScroll()
    },
    selectSession(session) {
      if (this.currentSessionId === session.id) return

      // 保存当前会话的前台状态（不关闭 WS，保持并发执行）
      this.saveCurrentSession()

      this.currentSessionId = session.id
      this.currentSession = session
      this.interactionMode = this.loadRememberedInteractionMode(session.id) || this.defaultInteractionMode
      this.pendingPlanChoice = null
      this.planState = { visible: false, phase: '', items: [] }
      this.planCollapsed = false
      this.switchingInteractionMode = false
      // 选中会话时同步当前工作空间，确保「在旧对话继续提问」也能把所属工作空间置顶
      const selWsId = Number(session.workspace_id || 0)
      if (selWsId) this.currentWorkspaceId = selWsId
      this.pendingSession = false
      this.messages = []
      this.streamingText = ''
      this.streamingThinking = ''
      this.pendingToolCalls = {}
      this.tokenStats = null
      this.compacting = false
      this._assistantPushedInTurn = false
      this.isStreaming = false
      this.sending = false
      this.resetMessageAutoScroll()
      this.stopThinkingTimer()
      this.stopStatsPolling()
      this.resetExecutionTimer()
      this._historyLoaded = false // 标记：HTTP API 是否已加载了历史消息

      // 恢复该会话最后使用的模型
      if (session.model_name && this.providerModels.length > 0) {
        this.restoreSessionModel(session.model_name)
      }

      // 尝试从后台状态恢复（WS 仍活跃则恢复连接）
      const restored = this.restoreSessionState(session.id)
      if (restored) {
        this.historyLoading = false
        if (this.messages.length === 0) this.loadSessionMessages()
      } else {
        // 需要新建 WS 连接
        this.historyLoading = true
        this.loadSessionMessages()
        if (session.status === 'running') this.connectWS(true)
      }
      // 执行耗时由后端计时：已结束会话直接用库里的 exec_duration_ms 展示；运行中的会话由 WS exec_progress 实时推送
      if (session.status !== 'running' && session.exec_duration_ms) {
        this.executionServerMs = Number(session.exec_duration_ms || 0)
        this.executionRunning = false
        this.executionSyncAt = Date.now()
        this.executionElapsedMs = this.executionServerMs
      } else {
        this.resetExecutionTimer()
      }
    },
    deleteSession(session) {
      this.$confirm('确定删除此对话？', '提示', { type: 'warning' }).then(() => {
        // 先断开该会话的 WS（无论前台还是后台）
        this.disconnectSessionWS(session.id)
        if (this.currentSessionId === session.id) {
          // 删除的是当前前台会话
          this.ws = null
          this.wsConnected = false
          this.currentSessionId = 0
          this.currentSession = null
          this.messages = []
          this.isStreaming = false
          this.streamingText = ''
          this.streamingThinking = ''
          this.pendingToolCalls = {}
          this.tokenStats = null
          this.compacting = false
          this.pendingPlanChoice = null
          this.planState = { visible: false, phase: '', items: [] }
          this.planCollapsed = false
          this.switchingInteractionMode = false
          this.stopThinkingTimer()
          this.stopStatsPolling()
          this.resetExecutionTimer()
          this.resetMessageAutoScroll()
        }
        // 调用后端删除
        Base.BasePost('/api/AgentV2SessionDelete', { id: session.id }, () => {
          this.loadSessions()
        })
      }).catch(() => {})
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
          const historyPlan = res.Data.plan_state
          if (historyPlan && Array.isArray(historyPlan.items) && historyPlan.items.length > 0 &&
              (!this.planState.items || this.planState.items.length === 0)) {
            this.planState = {
              visible: historyPlan.visible !== false,
              phase: historyPlan.phase || 'plan',
              items: historyPlan.items.map((item) => ({
                text: item && item.text != null ? String(item.text) : '',
                done: !!(item && item.done)
              }))
            }
            if (this.currentSession && this.currentSession.status === 'running' &&
                historyPlan.pending_plan_choice && !this.pendingPlanChoice) {
              this.pendingPlanChoice = historyPlan.pending_plan_choice
            }
          }
          this._historyLoaded = true
          this.scrollToBottom({ force: true })
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
            } catch (e) {
              // agent.config 可能为空或非 JSON，保持默认配置即可
            }
          }
          // Agent 配置就绪后，依据启动参数判断计划模式是否启用
          this.checkPlanExtension()
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
    cloneSessionData(value, fallback) {
      if (value === undefined || value === null) return fallback
      try {
        return JSON.parse(JSON.stringify(value))
      } catch (e) {
        // 会话快照仅用于 UI 恢复，克隆失败时回退到安全默认值
        return fallback
      }
    },
    captureSessionRuntimeState(target) {
      target.messages = this.cloneSessionData(this.messages, [])
      target.streamingText = this.streamingText
      target.streamingThinking = this.streamingThinking
      target.pendingToolCalls = this.cloneSessionData(this.pendingToolCalls, {})
      target.tokenStats = this.cloneSessionData(this.tokenStats, null)
      target.compacting = this.compacting
      target._assistantPushedInTurn = this._assistantPushedInTurn
      target._historyLoaded = this._historyLoaded
      target._lastUserMessage = this._lastUserMessage || ''
      target.thinkingStartAt = this.thinkingStartAt
      target.thinkingElapsedSeconds = this.thinkingElapsedSeconds
      target.runtimeNow = this.runtimeNow
      target.pendingPlanChoice = this.cloneSessionData(this.pendingPlanChoice, null)
      target.planState = this.cloneSessionData(this.planState, { visible: false, phase: '', items: [] })
      target.planCollapsed = this.planCollapsed
      return target
    },
    applySessionRuntimeState(snapshot) {
      this.messages = this.cloneSessionData(snapshot.messages, [])
      this.streamingText = snapshot.streamingText || ''
      this.streamingThinking = snapshot.streamingThinking || ''
      this.pendingToolCalls = this.cloneSessionData(snapshot.pendingToolCalls, {})
      this.tokenStats = this.cloneSessionData(snapshot.tokenStats, null)
      this.compacting = Boolean(snapshot.compacting)
      this._assistantPushedInTurn = Boolean(snapshot._assistantPushedInTurn)
      this._historyLoaded = Boolean(snapshot._historyLoaded)
      this._lastUserMessage = snapshot._lastUserMessage || ''
      this.thinkingStartAt = Number(snapshot.thinkingStartAt || 0)
      this.pendingPlanChoice = this.cloneSessionData(snapshot.pendingPlanChoice, null)
      this.planState = this.cloneSessionData(snapshot.planState, { visible: false, phase: '', items: [] })
      this.planCollapsed = Boolean(snapshot.planCollapsed)
      this.switchingInteractionMode = false
      this.runtimeNow = Date.now()
      this.thinkingElapsedSeconds = this.thinkingStartAt
        ? Math.max(0, Math.floor((this.runtimeNow - this.thinkingStartAt) / 1000))
        : 0

      // 执行耗时由后端驱动，不在前端快照中保存

      this.syncLiveAssistantMessage()

      const hasRunningTool = Object.values(this.pendingToolCalls).some(tc => this.isToolRunning(tc))
      if (this.thinkingStartAt || hasRunningTool || this.isStreaming) {
        this.startRuntimeTicker()
      }
      if (this.isStreaming) {
        this.startStatsPolling()
      }
      this.scrollToBottom({ force: true })
    },
    replayBufferedSessionMessages(bufferedMessages) {
      for (const data of bufferedMessages || []) {
        this.handleWSMessage(data)
      }
    },
    attachRunningSessions() {
      this.sessions
        .filter(session => session && session.status === 'running')
        .forEach(session => {
          if (this.currentSessionId === session.id && this.ws && this.ws.readyState !== WebSocket.CLOSED) return
          const existing = this.sessionStates[session.id]
          if (existing && existing.ws && (existing.ws.readyState === WebSocket.OPEN || existing.ws.readyState === WebSocket.CONNECTING)) return
          this.connectBackgroundSessionWS(session)
        })
    },
    connectBackgroundSessionWS(session) {
      if (!session || !session.id) return
      const apiHost = Base.GetAbsoluteApiHost()
      const protocol = apiHost.startsWith('https') ? 'wss:' : 'ws:'
      const host = apiHost.replace(/^https?:\/\//, '')
      const token = Base.GetSafeToken() || ''
      const url = `${protocol}//${host}/api/AgentV2WS?agent_id=${this.agentId}&session_id=${session.id}&token=${token}&attach_only=1`
      const ws = new WebSocket(url)
      ws._sessionId = session.id
      const state = this.sessionStates[session.id] || {}
      state.ws = ws
      state.wsConnected = false
      state.selectedModel = session.model_name || this.selectedModel
      state._isRunning = true
      state._bufferedMessages = state._bufferedMessages || []
      this.sessionStates[session.id] = state

      ws.onopen = () => {
        state.wsConnected = true
        state._isRunning = true
      }
      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          state._bufferedMessages.push(data)
          if (data.type === 'state' && data.state && data.state.running === false) {
            state._isRunning = false
            this.markSessionStatus(session.id, 'active')
            try { ws.close() } catch (e) { /* ignore */ }
            delete this.sessionStates[session.id]
            return
          }
          if (data.type === 'event' && data.event) {
            const evtType = data.event.type
            if (evtType === 'agent_start') state._isRunning = true
            if (evtType === 'agent_end') {
              state._isRunning = false
              this.markSessionStatus(session.id, 'active')
              try { ws.close() } catch (e) { /* ignore */ }
              delete this.sessionStates[session.id]
              this.loadSessions()
            }
            if (evtType === 'extension_ui_request') {
              this.promoteBackgroundSession(session.id, '后台会话需要交互，已切回前台')
            }
          }
        } catch (e) { /* ignore parse errors in background */ }
      }
      ws.onclose = () => {
        state.wsConnected = false
      }
      ws.onerror = () => {
        state.wsConnected = false
      }
    },
    markSessionStatus(sessionId, status) {
      const session = this.sessions.find(item => item.id === sessionId)
      if (session) session.status = status
    },
    getWorkspaceName(workspaceId) {
      const workspace = this.workspaces.find(item => Number(item.id) === Number(workspaceId))
      return workspace ? workspace.name : ''
    },
    getWorkspacePath(workspaceId) {
      const workspace = this.workspaces.find(item => Number(item.id) === Number(workspaceId))
      return workspace ? workspace.path : ''
    },

    // ========== 多会话并发状态管理 ==========
    // 保存当前前台会话状态，并将 WS 切换为后台监听模式
    saveCurrentSession() {
      const sid = this.currentSessionId
      if (!sid || !this.ws) return

      const existing = this.sessionStates[sid] || {}
      existing.ws = this.ws
      existing.wsConnected = this.ws.readyState === WebSocket.OPEN
      existing.selectedModel = this.selectedModel
      existing.interactionMode = this.interactionMode
      existing._isRunning = this.isStreaming // 保持当前运行状态，避免切换后转圈消失
      existing._bufferedMessages = existing._bufferedMessages || []
      this.captureSessionRuntimeState(existing)

      // 保存原始 WS 回调
      existing._onmessage = this.ws.onmessage
      existing._onclose = this.ws.onclose
      existing._onerror = this.ws.onerror

      // 替换为后台处理器（仅追踪运行状态，不修改前台属性）
      this.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          existing._bufferedMessages.push(data)
          if (data.type === 'event' && data.event) {
            const evtType = data.event.type
            if (evtType === 'agent_start') existing._isRunning = true
            if (evtType === 'agent_end') existing._isRunning = false

            if (evtType === 'extension_ui_request') {
              this.promoteBackgroundSession(sid, '后台会话需要交互，已切回前台')
              return
            }

            if (evtType === 'extension_error') {
              this.$message.warning(this.getBackgroundSessionLabel(sid) + '扩展错误: ' + (data.event.error || data.event.message || '未知错误'))
            }
          } else if (data.type === 'error') {
            this.$message.error(this.getBackgroundSessionLabel(sid) + (data.error || '后台会话发生错误'))
          }
        } catch (e) { /* ignore parse errors in background */ }
      }
      this.ws.onclose = () => {
        existing.wsConnected = false
      }
      this.ws.onerror = () => {
        existing.wsConnected = false
      }

      this.sessionStates[sid] = existing

      // 分离前台引用（不关闭 WS）
      this.ws = null
      this.wsConnected = false
      if (this._runtimeTicker) { clearInterval(this._runtimeTicker); this._runtimeTicker = null }
      if (this._statsPollTimer) { clearInterval(this._statsPollTimer); this._statsPollTimer = null }
    },

    // 恢复指定会话到前台
    // 返回 true 表示从已保存状态恢复，false 表示需要新建 WS 连接
    restoreSessionState(sessionId) {
      const ss = this.sessionStates[sessionId]
      if (!ss || !ss.ws) return false

      // 若会话已停止（Pi 进程已退出），清理旧连接并返回 false，触发 connectWS() 新建连接
      if (!ss._isRunning || ss.ws.readyState === WebSocket.CLOSED || ss.ws.readyState === WebSocket.CLOSING) {
        this.disconnectSessionWS(sessionId)
        return false
      }

      const bufferedMessages = Array.isArray(ss._bufferedMessages) ? ss._bufferedMessages.slice() : []
      ss._bufferedMessages = []

      // 恢复 WS 原始回调
      ss.ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data)
          this.handleWSMessage(data)
        } catch (e) {
          console.error('WS message parse error:', e)
        }
      }
      ss.ws.onopen = () => {
        this.wsConnected = true
        this.requestTokenStats()
      }
      ss.ws.onclose = () => {
        this.wsConnected = false
        this.isStreaming = false
      }
      ss.ws.onerror = (e) => {
        console.error('WS error:', e)
        this.wsConnected = false
      }

      this.ws = ss.ws
      this.wsConnected = ss.ws.readyState === WebSocket.OPEN
      if (ss.selectedModel) this.selectedModel = ss.selectedModel
      if (ss.interactionMode) this.interactionMode = ss.interactionMode

      // 若后台会话仍在执行中，恢复 isStreaming 状态
      this.isStreaming = Boolean(ss._isRunning)
      this.applySessionRuntimeState(ss)

      // 清除已恢复的 sessionStates，避免重复引用
      delete this.sessionStates[sessionId]
      this.replayBufferedSessionMessages(bufferedMessages)

      return true
    },
    getBackgroundSessionLabel(sessionId) {
      const session = this.sessions.find(item => item.id === sessionId)
      return session ? ('[' + session.name + '] ') : ''
    },
    promoteBackgroundSession(sessionId, tip) {
      if (this.currentSessionId === sessionId) return
      const session = this.sessions.find(item => item.id === sessionId)
      if (!session) {
        this.$message.warning(tip)
        return
      }
      this.$message.warning(tip)
      this.selectSession(session)
    },

    // 断开指定会话的 WS（不触发 onclose 中的前台状态修改）
    disconnectSessionWS(sessionId) {
      if (this.ws && this.currentSessionId === sessionId) {
        this.disconnectWS()
        return
      }
      const ss = this.sessionStates[sessionId]
      if (ss && ss.ws) {
        ss.ws.onclose = null
        ss.ws.onerror = null
        ss.ws.onmessage = null
        try { ss.ws.close() } catch (e) { /* ignore */ }
      }
      delete this.sessionStates[sessionId]
    },

    // 断开所有会话的 WS（组件销毁时调用）
    disconnectAllWS() {
      // 断开前台 WS
      if (this.ws) {
        this.ws.onclose = null
        this.ws.onerror = null
        try { this.ws.close() } catch (e) { /* ignore */ }
        this.ws = null
      }
      // 断开所有后台 WS
      Object.keys(this.sessionStates).forEach(sid => {
        this.disconnectSessionWS(Number(sid))
      })
      this.wsConnected = false
      this.isStreaming = false
      this.stopThinkingTimer()
      this.stopStatsPolling()
      this.stopExecutionTimer()
    },

    // ========== WebSocket ==========
    connectWS(attachOnly = false) {
      if (!this.currentSessionId) return
      const existing = this.sessionStates[this.currentSessionId]
      if (existing && existing.ws && (existing.ws.readyState === WebSocket.OPEN || existing.ws.readyState === WebSocket.CONNECTING)) {
        this.restoreSessionState(this.currentSessionId)
        return
      }
      const apiHost = Base.GetAbsoluteApiHost() // dev: http://localhost:17170, prod: current origin
      const protocol = apiHost.startsWith('https') ? 'wss:' : 'ws:'
      const host = apiHost.replace(/^https?:\/\//, '')
      const token = Base.GetSafeToken() || ''
      const modelParam = this.selectedModel ? `&model=${encodeURIComponent(this.selectedModel)}` : ''
      const modeParam = this.planExtensionInstalled ? `&interaction_mode=${encodeURIComponent(this.interactionMode)}` : ''
      const attachParam = attachOnly ? '&attach_only=1' : ''
      const url = `${protocol}//${host}/api/AgentV2WS?agent_id=${this.agentId}&session_id=${this.currentSessionId}&token=${token}${modelParam}${modeParam}${attachParam}`

      const sessionId = this.currentSessionId // 闭包捕获
      this.ws = new WebSocket(url)
      this.ws._sessionId = sessionId
      this.ws.onopen = () => {
        this.wsConnected = true
        // 连接成功立即请求会话统计（上下文使用率、Token 等）
        this.requestTokenStats()
        // 懒创建模式：发送暂存的首条消息
        if (this._pendingFirstMessage) {
          const msg = this._pendingFirstMessage
          this._pendingFirstMessage = ''
          this.markSessionStatus(sessionId, 'running')
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
            if (attachOnly && data.type === 'state' && data.state && data.state.running === false) {
              this.markSessionStatus(sessionId, 'active')
              this.isStreaming = false
              this.stopThinkingTimer()
              this.stopStatsPolling()
              this.stopExecutionTimer()
              this.stopRuntimeTickerIfIdle()
            }
          } catch (e) {
            console.error('WS message parse error:', e)
        }
      }
      this.ws.onclose = () => {
        this.wsConnected = false
      }
      this.ws.onerror = (e) => {
        console.error('WS error:', e)
        this.wsConnected = false
      }
    },
    disconnectWS() {
      if (this.ws) {
        this.ws.onclose = null
        this.ws.onerror = null
        this.ws.close()
        this.ws = null
      }
      this.wsConnected = false
      this.isStreaming = false
      this.stopThinkingTimer()
      this.stopStatsPolling()
      this.stopExecutionTimer()
    },
    handleWSMessage(data) {
      // 防御性检查：忽略来自非当前会话的消息
      if (this.ws && this.ws._sessionId && this.ws._sessionId !== this.currentSessionId) return
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
        if (data.state && data.state.running === false) {
          const sid = Number(data.state.session_id || this.currentSessionId)
          this.markSessionStatus(sid, 'active')
          if (sid === this.currentSessionId) {
            this.isStreaming = false
            this.stopThinkingTimer()
            this.stopStatsPolling()
            this.stopExecutionTimer()
            this.stopRuntimeTickerIfIdle()
          }
        }
      } else if (data.type === 'history' && data.messages) {
        // 如果 HTTP API 已加载历史消息，不覆盖（避免重复造成闪烁）
        if (!this._historyLoaded || this.messages.length === 0) {
          this.messages = data.messages
          this.scrollToBottom({ force: true })
        }
      } else if (data.type === 'error') {
        this.$message.error(data.error)
        this.sending = false
      }
    },
    handlePiEvent(event) {
      const evtType = event.type

      switch (evtType) {
        // ===== 后端推送的执行耗时（agent_start / 工具·思考完成 / agent_end 等触发） =====
        case 'exec_progress': {
          this.handleExecProgress(event.total_ms, event.running)
          break
        }

        // ===== 消息流式更新 =====
        case 'message_update': {
          const msgEvt = event.assistantMessageEvent || {}
          const deltaType = msgEvt.type

          if (deltaType === 'text_delta') {
            this.streamingText += (msgEvt.delta || '')
            this.syncLiveAssistantMessage()
            this.scrollToBottom()
          } else if (deltaType === 'thinking_delta') {
            if (!this.thinkingStartAt) this.startThinkingTimer()
            this.streamingThinking += (msgEvt.delta || '')
            this.syncLiveAssistantMessage()
            this.scrollToBottom()
          } else if (deltaType === 'text_start' || deltaType === 'text_end' ||
                     deltaType === 'thinking_start' || deltaType === 'thinking_end') {
            if (deltaType === 'thinking_start') {
              if (!this.thinkingStartAt) this.startThinkingTimer()
              // 占位零宽空格，确保 thinking 块在 toolcall_start 之前立即可见，避免顺序错位
              if (!this.streamingThinking) this.streamingThinking = '\u200B'
            }
            if (deltaType === 'thinking_end') this.stopThinkingTimer()
            this.syncLiveAssistantMessage()
            this.scrollToBottom()
          } else if (deltaType === 'toolcall_start' || deltaType === 'toolcall_delta' || deltaType === 'toolcall_end') {
            // 支持 Anthropic (msgEvt.toolCall) 和 DeepSeek/OpenAI (partial.content) 两种格式
            this.handleToolCallInMessageUpdate(msgEvt)
            this.scrollToBottom()
          }
          break
        }

        // ===== 消息生命周期 =====
        case 'message_start': {
          const msg = event.message
          if (msg && msg.role === 'user') {
            this.scrollToBottom({ force: true })
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
            // 清理零宽空格占位符，避免空的 thinking 块被当作有效内容
            const thinkingContent = this.streamingThinking.replace(/\u200B/g, '')
            // 仅在有实际内容时才 push（与后端 reconstructMessagesFromPiEvents 一致）
            if (text || errorMsg || thinkingContent || Object.keys(this.pendingToolCalls).length > 0) {
              this.finalizeLiveAssistantMessage(text || (errorMsg ? '**Error:** ' + errorMsg : ''), thinkingContent)
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
          const target = this.ensureToolExecutionTarget(tcId, event)
          if (target) {
            target.tc.status = 'running'
            if (!target.tc.startedAt) {
              target.tc.startedAt = Date.now()
            }
            this.startRuntimeTicker()
            if (!target.inMessage) this.syncLiveAssistantMessage()
            this.scrollToBottom()
          }
          break
        }
        case 'tool_execution_update': {
          const tcId = event.toolCallId || event.id
          const target = this.ensureToolExecutionTarget(tcId, event)
          if (target) {
            this.applyToolExecutionOutput(target.tc, event, false)
            if (!target.inMessage) this.syncLiveAssistantMessage()
            this.scrollToBottom()
          }
          break
        }
        case 'tool_execution_end': {
          const tcId = event.toolCallId || event.id
          const target = this.ensureToolExecutionTarget(tcId, event)
          if (target) {
            target.tc.status = 'done'
            target.tc.durationMs = target.tc.startedAt ? (Date.now() - target.tc.startedAt) : 0
            this.applyToolExecutionOutput(target.tc, event, true)
            if (!target.inMessage) this.syncLiveAssistantMessage()
            this.stopRuntimeTickerIfIdle()
            this.scrollToBottom()
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
          // 后端已确认收到用户问题：清空输入框并停止发送转圈
          this.inputText = ''
          this.sending = false
          this.ensureLiveAssistantMessage()
          // 用户主动追问时，无论之前是否滚动到顶部，都强制滚动到底部
          this.scrollToBottom({ force: true })
          // 执行耗时由后端在 agent_start 时推送 exec_progress，前端无需自行计时
          break
        }
        case 'agent_end': {
          this.markSessionStatus(this.currentSessionId, 'active')
          this.isStreaming = false
          this.stopThinkingTimer()
          this.stopExecutionTimer()
          this.stopStatsPolling()
          // 仅在 message_end 未推送时才兜底推送（与后端 needPushAssistant 逻辑一致）
          const thinkingContent = this.streamingThinking.replace(/\u200B/g, '')
          if (!this._assistantPushedInTurn && (this.streamingText || thinkingContent || Object.values(this.pendingToolCalls).length > 0)) {
            this.finalizeLiveAssistantMessage(this.streamingText, thinkingContent)
          } else {
            this.removeLiveAssistantMessage()
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

        // ===== 计划模式扩展：计划/进度更新 =====
        case 'plan_update': {
          const items = Array.isArray(event.items)
            ? event.items.map((it) => ({
                text: it && it.text != null ? String(it.text) : '',
                done: !!(it && it.done)
              }))
            : []
          this.planState = {
            visible: true,
            phase: event.phase || 'plan',
            items
          }
          this.scrollToBottom()
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

      if (method === 'setStatus' && event.statusKey === 'plan-mode') {
        const status = String(event.statusText || '')
        this.interactionMode = status.includes('plan') ? 'plan' : 'execute'
        this.switchingInteractionMode = false
        if (status && !status.includes('plan')) {
          this.planState = { ...this.planState, visible: this.planState.items.length > 0, phase: 'execute' }
        }
      } else if (method === 'setWidget' && event.widgetKey === 'plan-todos') {
        const lines = Array.isArray(event.widgetLines) ? event.widgetLines : []
        if (lines.length > 0) {
          this.planState = {
            visible: true,
            phase: 'execute',
            items: lines.map((line) => {
              const text = String(line)
              return {
                done: text.trim().startsWith('☑'),
                text: text.replace(/^\s*[☑☐]\s*/, '').replace(/^~+|~+$/g, '')
              }
            })
          }
        } else if (this.planState.phase === 'execute') {
          this.planState = { visible: false, phase: '', items: [] }
        }
      } else if (method === 'confirm') {
        this.$confirm(event.message || event.title || '确认操作?', event.title || '提示', {
          confirmButtonText: '确认',
          cancelButtonText: '取消'
        }).then(() => {
          this.sendWS({ type: 'command', command: { type: 'extension_ui_response', id: reqId, confirmed: true } })
        }).catch(() => {
          this.sendWS({ type: 'command', command: { type: 'extension_ui_response', id: reqId, cancelled: true } })
        })
      } else if (method === 'select') {
        const options = event.options || []
        if (options.length === 0) {
          this.sendWS({ type: 'command', command: { type: 'extension_ui_response', id: reqId, cancelled: true } })
          return
        }

        // Pi 的 plan-mode 示例扩展会在计划生成后发出英文三选一请求。
        // 在网页端把它收进模式切换器和计划面板，不再弹英文系统框。
        const isPlanModeChoice = options.some((item) => String(item).startsWith('Execute')) &&
          options.some((item) => String(item).startsWith('Stay'))
        if (isPlanModeChoice) {
          const items = this.extractPlanItemsFromMessages()
          this.pendingPlanChoice = { id: reqId, options }
          this.interactionMode = 'plan'
          this.planState = {
            visible: items.length > 0,
            phase: 'plan',
            items
          }
          this.scrollToBottom()
          return
        }

        // 其他扩展的通用选择请求仍使用弹窗。
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
          const partialArgs = block.partialArgs || block.partialJson
          if (partialArgs && (!args || (typeof args === 'object' && Object.keys(args).length === 0))) {
            try { this.pendingToolCalls[block.id].input = JSON.parse(partialArgs) } catch(e) {
              this.pendingToolCalls[block.id].input = partialArgs
            }
          }
          // toolcall_end 时标记参数收集完毕
          if (msgEvt.type === 'toolcall_end' && this.pendingToolCalls[block.id]) {
            this.pendingToolCalls[block.id]._inputFinalized = true
          }
        }
      }
      this.syncLiveAssistantMessage()
    },

    findToolCallInMessages(tcId) {
      if (!tcId) return null
      for (let i = this.messages.length - 1; i >= 0; i--) {
        const msg = this.messages[i]
        if (msg.role === 'assistant' && Array.isArray(msg.toolCalls)) {
          const msgTc = msg.toolCalls.find(t => t.id === tcId)
          if (msgTc) return msgTc
        }
      }
      return null
    },
    ensureToolExecutionTarget(tcId, event = {}) {
      if (!tcId) return null
      if (this.pendingToolCalls[tcId]) {
        const tc = this.pendingToolCalls[tcId]
        if (!tc.name || tc.name === 'unknown') tc.name = event.toolName || event.name || 'unknown'
        if (!tc.input && (event.args || event.input)) tc.input = event.args || event.input
        return { tc, inMessage: false }
      }

      const messageTc = this.findToolCallInMessages(tcId)
      if (messageTc) {
        if (!messageTc.name || messageTc.name === 'unknown') messageTc.name = event.toolName || event.name || 'unknown'
        if (!messageTc.input && (event.args || event.input)) messageTc.input = event.args || event.input
        return { tc: messageTc, inMessage: true }
      }

      if (!this.pendingToolCalls[tcId]) {
        this.pendingToolCalls[tcId] = {
          id: tcId,
          name: event.toolName || event.name || 'unknown',
          status: 'running',
          input: event.args || event.input || '',
          output: '',
          _collapsed: true,
          startedAt: Date.now(),
          durationMs: 0
        }
      }
      return { tc: this.pendingToolCalls[tcId], inMessage: false }
    },
    applyToolExecutionOutput(tc, event, final) {
      if (!tc) return
      const hasOwn = Object.prototype.hasOwnProperty
      if (hasOwn.call(event, 'output')) {
        const output = this.formatToolOutput(event.output)
        tc.output = final ? (output || tc.output || '') : ((tc.output || '') + output)
        return
      }
      if (hasOwn.call(event, 'result')) {
        const output = this.formatToolOutput(event.result)
        tc.output = output || tc.output || ''
        return
      }
      if (hasOwn.call(event, 'partialResult')) {
        const output = this.formatToolOutput(event.partialResult)
        tc.output = output || tc.output || ''
      }
    },
    syncToolCallToMessages(tcId) {
      const tc = this.pendingToolCalls[tcId]
      const msgTc = this.findToolCallInMessages(tcId)
      if (!tc || !msgTc) return
      msgTc.status = tc.status
      msgTc.output = tc.output
    },
    ensureLiveAssistantMessage() {
      const lastMsg = this.messages[this.messages.length - 1]
      if (lastMsg && lastMsg.role === 'assistant' && lastMsg._live) return lastMsg
      const liveMsg = {
        role: 'assistant',
        content: '',
        thinking: '',
        thinkingDurationMs: 0,
        toolCalls: undefined,
        _live: true,
        _thinkingCollapsed: false
      }
      this.messages.push(liveMsg)
      return liveMsg
    },
    removeLiveAssistantMessage() {
      const lastIdx = this.messages.length - 1
      if (lastIdx < 0) return
      const lastMsg = this.messages[lastIdx]
      if (lastMsg && lastMsg.role === 'assistant' && lastMsg._live) {
        this.messages.splice(lastIdx, 1)
      }
    },
    syncLiveAssistantMessage() {
      if (!this.isStreaming && !this.streamingText && !this.streamingThinking && Object.keys(this.pendingToolCalls).length === 0) return
      const liveMsg = this.ensureLiveAssistantMessage()
      const thinkingContent = this.streamingThinking.replace(/\u200B/g, '')
      liveMsg.content = this.streamingText
      liveMsg.thinking = thinkingContent
      liveMsg.thinkingDurationMs = this.getCurrentThinkingDurationMs(false)
      const toolCalls = Object.values(this.pendingToolCalls).map(item => ({ ...item }))
      liveMsg.toolCalls = toolCalls.length > 0 ? toolCalls : undefined
    },
    finalizeLiveAssistantMessage(content, thinkingContent) {
      const liveMsg = this.ensureLiveAssistantMessage()
      liveMsg.content = content
      liveMsg.thinking = thinkingContent
      liveMsg.thinkingDurationMs = this.getCurrentThinkingDurationMs(true)
      const toolCalls = Object.values(this.pendingToolCalls).map(item => ({ ...item }))
      liveMsg.toolCalls = toolCalls.length > 0 ? toolCalls : undefined
      delete liveMsg._live
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
    // ========== 执行耗时（后端 WS exec_progress 推送，前端仅插值平滑） ==========
    // 后端在 agent_start / 工具·思考完成 / agent_end 等事件时推送最新累计耗时，
    // 并在每轮结束时落库 tbl_agent_v2_session.exec_duration_ms
    handleExecProgress(totalMs, running) {
      this.executionServerMs = Number(totalMs || 0)
      this.executionRunning = !!running
      this.executionSyncAt = Date.now()
      this.executionElapsedMs = this.executionServerMs
      // 非运行态（已结束）时把最终耗时同步到当前会话，便于列表/头部持久展示
      if (!running && this.currentSession) {
        this.currentSession.exec_duration_ms = this.executionServerMs
      }
    },
    // 全局插值计时器：运行中时按本地时钟平滑累加，避免 2s 推送间隔导致跳变
    ensureExecTicker() {
      if (this._execTicker) return
      this._execTicker = setInterval(() => {
        if (this.executionRunning) {
          this.executionElapsedMs = this.executionServerMs + (Date.now() - this.executionSyncAt)
        }
      }, 250)
    },
    fmtExecDuration(ms) {
      if (!ms || ms < 0) return '0s'
      const totalSec = Math.floor(ms / 1000)
      const h = Math.floor(totalSec / 3600)
      const m = Math.floor((totalSec % 3600) / 60)
      const s = totalSec % 60
      if (h > 0) return `${h}h${m}m${s}s`
      if (m > 0) return `${m}m${s}s`
      return `${s}s`
    },
    // WS 断开时冻结展示值（保留最近一次后端推送的累计耗时），待重连后由 exec_progress 恢复
    stopExecutionTimer() {
      this.executionRunning = false
    },
    resetExecutionTimer() {
      this.executionRunning = false
      this.executionServerMs = 0
      this.executionSyncAt = 0
      this.executionElapsedMs = 0
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
    sendMessage(overrideText) {
      // 注意：通过 @click / @keydown 绑定调用时会传入 DOM 事件对象，
      // 只有当传入的是字符串时才作为覆盖文本，否则使用输入框内容
      const text = (typeof overrideText === 'string' ? overrideText : this.inputText).trim()
      if (!text || this.isStreaming || this.sending) return

      // 保存最后发送的消息文本（agent_start 时用于展示用户消息）
      this._lastUserMessage = text
      // 发送中：按钮转圈，先不清空输入框，待后端 agent_start 确认收到后再清空
      this.sending = true

      // 对话执行时把所在工作空间置顶（同时写后端排序数据）
      if (this.currentWorkspaceId) {
        this.bumpWorkspaceToFront(this.currentWorkspaceId)
      }

      // 立即用当前问题更新会话标题（与后端 prompt 重命名保持一致，
      // 避免新建会话发送首条消息后顶部仍显示创建时间的占位名）
      if (this.currentSession && this.currentSessionId) {
        const title = this.truncateTitle(text)
        this.currentSession.name = title
        const inList = this.sessions.find(s => s.id === this.currentSessionId)
        if (inList) inList.name = title
      }

      // 懒创建模式：先暂存消息，等会话创建+WS 连接成功后再发送
      if (this.pendingSession && !this.currentSessionId) {
        this._pendingFirstMessage = text
        this.createRealSessionAndSend(text)
        return
      }

      if (!this.wsConnected) {
        this._pendingFirstMessage = text
        this.connectWS()
        return
      }

      this.markSessionStatus(this.currentSessionId, 'running')
      this.sendWS({
        type: 'command',
        command: { type: 'prompt', message: text }
      })
    },
    createRealSessionAndSend(firstText) {
      const title = this.truncateTitle(firstText)
      Base.BasePost('/api/AgentV2SessionCreate', {
        agent_id: this.agentId,
        workspace_id: this.currentWorkspaceId,
        name: title
      }, (res) => {
        const newId = (res.ErrCode === 0 && res.Data) ? res.Data.id : null
        if (!newId) {
          this.$message.error('创建会话失败')
          this.pendingSession = false
          this.sending = false
          return
        }
        // 添加到会话列表
        const now = Math.floor(Date.now() / 1000)
        const newSession = {
          id: newId,
          agent_id: this.agentId,
          workspace_id: this.currentWorkspaceId,
          workspace_name: this.getWorkspaceName(this.currentWorkspaceId),
          workspace_path: this.getWorkspacePath(this.currentWorkspaceId),
          name: title,
          status: 'running',
          created_at: now,
          updated_at: now
        }
        this.sessions.unshift(newSession)
        this.currentSessionId = newId
        this.currentSession = newSession
        this.rememberInteractionMode()
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
      // Pi 的计划选择框属于扩展 UI 等待，单独发送 abort 无法解除；先取消该请求。
      if (this.pendingPlanChoice) {
        this.sendWS({
          type: 'command',
          command: {
            type: 'extension_ui_response',
            id: this.pendingPlanChoice.id,
            cancelled: true
          }
        })
        this.pendingPlanChoice = null
        this.planState = { ...this.planState, phase: 'plan' }
        this.releasePlanInputWait()
      }
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
    formatToolOutput(output) {
      if (output === undefined || output === null) return ''
      if (typeof output === 'string') return output
      const content = output.content
      if (Array.isArray(content)) {
        const text = content.map(block => {
          if (!block) return ''
          if (typeof block === 'string') return block
          if (block.type === 'text') return block.text || ''
          if (block.text) return block.text
          return ''
        }).join('')
        return text
      }
      try {
        return JSON.stringify(output, null, 2)
      } catch (e) {
        return String(output)
      }
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
    // 复用共享工具图标映射（src/utils/toolIcons.js）
    toolMeta(name) {
      return getToolMeta(name)
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
    isMessagesNearBottom(el) {
      if (!el) return true
      return el.scrollHeight - el.scrollTop - el.clientHeight <= 24
    },
    resetMessageAutoScroll() {
      this.autoScrollEnabled = true
      this.showScrollToBottom = false
      this.programmaticScrollActive = false
    },
    updateMessageAutoScrollState(forceFollow = false) {
      const el = this.$refs.messagesContainer
      if (!el) return
      const nearBottom = this.isMessagesNearBottom(el)
      if (forceFollow || nearBottom) {
        this.autoScrollEnabled = true
        this.showScrollToBottom = false
      } else {
        this.autoScrollEnabled = false
        this.showScrollToBottom = true
      }
    },
    handleMessagesWheel(event) {
      if (event.deltaY < 0) {
        this.autoScrollEnabled = false
        this.showScrollToBottom = true
      }
    },
    handleMessagesScroll() {
      if (this.programmaticScrollActive) {
        this.updateMessageAutoScrollState(true)
        return
      }
      this.updateMessageAutoScrollState()
    },
    scrollToBottom(options = {}) {
      const force = Boolean(options.force)
      if (!force && !this.autoScrollEnabled) {
        this.showScrollToBottom = true
        return
      }
      this.$nextTick(() => {
        const el = this.$refs.messagesContainer
        if (!el) return
        this.programmaticScrollActive = true
        el.scrollTop = el.scrollHeight
        this.autoScrollEnabled = true
        this.showScrollToBottom = false
        window.setTimeout(() => {
          this.programmaticScrollActive = false
          this.updateMessageAutoScrollState(true)
        }, 0)
      })
    },
    scrollToBottomAndResume() {
      this.autoScrollEnabled = true
      this.scrollToBottom({ force: true })
    },
    formatTime(ts) {
      if (!ts) return ''
      const d = new Date(ts * 1000)
      return d.toLocaleDateString() + ' ' + d.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })
    },
    // 按字符（rune）截断标题，避免按字节切 UTF-8 多字节中文产生乱码，与后端 prompt 重命名逻辑一致
    truncateTitle(text) {
      if (!text) return ''
      const runes = Array.from(text)
      if (runes.length > 50) return runes.slice(0, 50).join('') + '...'
      return text
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
.chat-sidebar__section--workspaces { display: none; }
.chat-sidebar__section--grow { flex: 1; overflow-y: auto; }
.chat-sidebar__section-title {
  display: flex; justify-content: space-between; align-items: center;
  gap: 8px;
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


.session-list { padding: 0 8px; }
.session-item {
  display: flex; align-items: center; gap: 8px; padding: 10px 12px;
  border-radius: 8px; cursor: pointer; font-size: 13px;
}
.session-item:hover { background: #f5f7fa; }
.session-item--active { background: #ecf5ff; }
.workspace-group-header {
  margin: 10px 0 4px;
  padding: 9px 10px;
  background: #f3f7ff;
  border: 1px solid #d9e8ff;
  border-left: 4px solid #409eff;
  border-radius: 6px;
  color: #1f2d3d;
  font-weight: 700;
}
.workspace-group-header:hover { background: #eaf3ff; border-color: #bcd9ff; }
.workspace-group-header--active { background: #e8f3ff; color: #1f5f99; }
.workspace-group-header--collapsed { margin-bottom: 8px; }
.workspace-group-header--dragover { border-top: 2px solid #409eff; box-shadow: 0 -2px 0 #409eff; }
.workspace-group-header__arrow,
.workspace-group-header__folder { flex-shrink: 0; color: #409eff; }
.workspace-group-header__count {
  flex-shrink: 0;
  font-size: 11px;
  color: #5f7897;
  background: #fff;
  border: 1px solid #d9e8ff;
  border-radius: 999px;
  padding: 2px 7px;
}
.session-item__add { opacity: 1; font-weight: 700; }
.session-item__info { flex: 1; min-width: 0; display: flex; align-items: center; gap: 6px; }
.session-item__name { flex: 1 1 auto; min-width: 0; overflow: hidden; text-overflow: ellipsis; white-space: nowrap; }
.workspace-group-header .session-item__name { flex: 0 0 auto; display: inline-block; max-width: 100%; vertical-align: middle; }
.session-item__exec { flex: 0 0 auto; margin-left: auto; font-size: 11px; color: #c0c4cc; }
.chat-header__created { display: block; margin-top: 3px; font-size: 12px; font-weight: 400; color: #909399; }
.session-item__del { opacity: 0; }
.session-item:hover .session-item__del { opacity: 1; }
.session-item__spinner {
  width: 14px; height: 14px; flex-shrink: 0;
}
.session-item--show-more {
  justify-content: center;
  color: #409eff;
  background: #f5f9ff;
  font-size: 12px;
  margin-top: 2px;
}
.session-item--show-more:hover { background: #eaf3ff; }
.session-item__more-icon { font-size: 13px; color: #409eff; }
.session-item__more-text { font-weight: 500; }

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
.agent-info__running-count {
  font-size: 11px; color: #e6a23c; padding: 2px 6px; border-radius: 4px; background: #fef7e0;
}

.chat-main { flex: 1; display: flex; flex-direction: column; min-width: 0; }
.chat-header {
  padding: 12px 24px; background: #fff; border-bottom: 1px solid #e4e7ed;
  display: flex; justify-content: space-between; align-items: center;
}
.chat-header__title { font-weight: 500; font-size: 15px; }
.chat-header__actions { display: flex; gap: 8px; align-items: center; }

.plan-mode-badge {
  font-size: 18px; cursor: pointer; line-height: 1;
  filter: grayscale(0.1); transition: transform .15s;
}
.plan-mode-badge:hover { transform: scale(1.15); }

/* 计划-确认-执行 面板 */
.plan-panel {
  width: min(680px, 100%);
  box-sizing: border-box;
  margin: 0 auto 10px;
  background: #fafbfc;
  border: 1px solid #dfe4ea;
  border-radius: 7px;
  box-shadow: 0 1px 2px rgba(0,0,0,.03);
  overflow: hidden;
}
.plan-panel__header {
  display: flex; align-items: center; gap: 6px;
  padding: 6px 10px; cursor: pointer;
  background: #f6f7f9; border-bottom: 1px solid #e6e9ed;
}
.plan-panel--collapsed .plan-panel__header { border-bottom: none; }
.plan-panel__icon { font-size: 14px; }
.plan-panel__title { font-weight: 600; font-size: 13px; color: #303133; }
.plan-panel__progress {
  font-size: 12px; color: #909399;
  font-variant-numeric: tabular-nums;
}
.plan-panel__toggle { margin-left: auto; font-size: 12px; color: #409eff; }
.plan-panel__body { padding: 7px 10px 9px; }
.plan-list {
  list-style: none;
  max-height: 160px;
  overflow-y: auto;
  margin: 0;
  padding: 0;
}
.plan-list li {
  display: flex; gap: 8px; align-items: flex-start;
  padding: 3px 0; font-size: 12px; color: #303133; line-height: 1.45;
}
.plan-list__check { color: #c0c4cc; flex-shrink: 0; }
.plan-list__done .plan-list__check { color: #67c23a; }
.plan-list__done .plan-list__text { color: #909399; text-decoration: line-through; }
.plan-panel__actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px solid #ebeef5;
}

.chat-input__mode-select {
  flex: 0 0 auto;
}
.chat-input__mode-button {
  display: inline-flex;
  align-items: center;
  gap: 5px;
  height: 28px;
  padding: 0 9px;
  border: 0;
  border-radius: 7px;
  font-weight: 600;
  box-shadow: none;
}
.chat-input__mode-button--execute {
  color: #176b3a;
  background: transparent;
}
.chat-input__mode-button--execute:hover,
.chat-input__mode-button--execute:focus {
  color: #0f5d30;
  background: transparent;
}
.chat-input__mode-button--plan {
  color: #9a5b00;
  background: transparent;
}
.chat-input__mode-button--plan:hover,
.chat-input__mode-button--plan:focus {
  color: #804b00;
  background: transparent;
}
.chat-input__mode-icon {
  font-size: 15px;
}
.chat-input__mode-arrow {
  margin-left: 1px;
  font-size: 11px;
  opacity: .65;
}
:global(.chat-mode-menu .el-dropdown-menu__item) {
  min-width: 126px;
  gap: 7px;
}
:global(.chat-mode-menu .el-dropdown-menu__item.is-active) {
  color: #409eff;
  background: #ecf5ff;
  font-weight: 600;
}


.exec-time {
  display: flex; align-items: center; gap: 4px;
  padding: 3px 10px; border-radius: 999px;
  background: #f4f6fa; border: 1px solid #e4e7ed;
  font-size: 12px; color: #606266; white-space: nowrap; user-select: none;
}
.exec-time__icon { font-size: 13px; color: #909399; }
.exec-time__label { color: #909399; }
.exec-time__value { font-variant-numeric: tabular-nums; font-weight: 600; color: #303133; }
.exec-time--running {
  background: #ecf5ff; border-color: #b3d8ff;
}
.exec-time--running .exec-time__icon { color: #409eff; animation: exec-pulse 1s ease-in-out infinite; }
@keyframes exec-pulse { 0%,100% { opacity: 1 } 50% { opacity: .35 } }

.chat-messages-wrap {
  position: relative;
  flex: 1;
  min-height: 0;
  margin-bottom: 8px;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}
.chat-messages {
  flex: 1;
  min-height: 0;
  overflow-y: auto;
  padding: 16px;
}
.chat-scroll-bottom {
  position: absolute;
  right: 24px;
  bottom: 20px;
  width: 36px;
  height: 36px;
  color: #606266;
  background: #fff;
  border: 1px solid #dcdfe6;
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.12);
  z-index: 5;
}
.chat-scroll-bottom:hover {
  color: #409eff;
  border-color: #409eff;
  background: #f5f9ff;
}
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

.chat-message { display: flex; gap: 8px; margin-bottom: 16px; max-width: 90%; }
.chat-message--assistant { max-width: 70%; }
.chat-message--user { margin-left: auto; flex-direction: row-reverse; max-width: 70%; }

.avatar {
  width: 28px; height: 28px; border-radius: 6px; flex-shrink: 0;
  display: flex; align-items: center; justify-content: center;
  font-weight: bold; font-size: 12px; color: #fff;
}
.avatar--user { background: #409eff; }
.avatar--assistant { background: linear-gradient(135deg, #667eea, #764ba2); }
.avatar--tool { background: #e6a23c; font-size: 16px; }

.chat-message__body { min-width: 0; }
.chat-message__content {
  background: #fff; border-radius: 10px; padding: 8px 12px;
  border: 1px solid #ebeef5; line-height: 1.5; font-size: 13px;
}
.chat-message--user .chat-message__content {
  background: #eaf2ff; color: #303133; border-color: #cfe0f7;
  max-height: 300px; overflow-y: auto; word-break: break-word;
}

.thinking-block { margin-bottom: 4px; }
.thinking-block__header {
  display: flex; align-items: center; gap: 6px; padding: 4px 10px;
  background: #fef7e0; border-radius: 6px; cursor: pointer; font-size: 13px; color: #b88230;
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
  background: #fffef8; border: 1px solid #faecd8; border-radius: 6px;
  padding: 8px; margin-top: 2px; font-size: 13px; color: #8c6d3a; white-space: pre-wrap; max-height: 200px; overflow-y: auto;
}

.tool-calls { margin-top: 4px; }
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
  background: #f8f9fc; border: 1px solid transparent; border-radius: 8px; padding: 6px 10px; margin-bottom: 2px;
  transition: border-color .2s ease;
}
.tool-call--expanded.tool-call--running { border-color: #e6a23c; background: #fef7e0; }
.tool-call--expanded.tool-call--done { border-color: #67c23a; }
.tool-call__header {
  display: flex; align-items: center; gap: 6px; margin-bottom: 4px; font-size: 13px;
}
.tool-call__name { font-weight: 500; }
.tool-call__icon {
  display: inline-flex; align-items: center; justify-content: center;
  width: 22px; height: 22px; border-radius: 6px; flex-shrink: 0;
  box-shadow: 0 1px 2px rgba(0,0,0,.04);
}
.tool-call__icon svg { width: 14px; height: 14px; display: block; }
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

.chat-input { padding: 12px 24px 14px; background: #fff; border-top: 1px solid #e4e7ed; }
.chat-input__wrapper {
  width: 100%; overflow: hidden; background: #fff;
  border: 1px solid #dcdfe6; border-radius: 10px;
  transition: border-color .2s ease, box-shadow .2s ease;
}
.chat-input__wrapper:hover { border-color: #c0c4cc; }
.chat-input__wrapper:focus-within {
  border-color: #409eff; box-shadow: 0 0 0 2px rgba(64, 158, 255, .1);
}
.chat-input__wrapper :deep(.el-textarea__inner) {
  min-height: 72px !important; padding: 14px 16px 8px;
  border: 0; border-radius: 0; box-shadow: none; background: transparent;
}
.chat-input__wrapper :deep(.el-textarea__inner:focus) { box-shadow: none; }
.chat-input__hint { font-size: 12px; color: #c0c4cc; margin-right: 8px; }

/* 输入框内底部工具栏 */
.chat-input__toolbar {
  display: flex; justify-content: space-between; align-items: center;
  min-height: 40px; padding: 4px 10px 8px; gap: 8px;
}
.chat-input__toolbar-left {
  display: flex; align-items: center; gap: 8px; flex-wrap: wrap; min-width: 0;
}
.chat-input__toolbar-right {
  display: flex; align-items: center; flex-shrink: 0;
}
.chat-input__model-select { width: 200px; }
.chat-input__model-select :deep(.el-select__wrapper),
.chat-input__model-select :deep(.el-input__wrapper) {
  padding: 0 6px; background: transparent; box-shadow: none !important;
}
.chat-input__model-select :deep(.el-select__wrapper:hover),
.chat-input__model-select :deep(.el-select__wrapper.is-hovering),
.chat-input__model-select :deep(.el-select__wrapper.is-focused),
.chat-input__model-select :deep(.el-input__wrapper:hover),
.chat-input__model-select :deep(.el-input__wrapper.is-focus) {
  background: #f5f7fa; box-shadow: none !important;
}
.chat-input__toolbar-btn {
  font-size: 12px; color: #606266; padding: 4px 6px;
  border: 0; background: transparent;
}
.chat-input__toolbar-btn:hover,
.chat-input__toolbar-btn:focus { color: #409eff; border-color: transparent; background: #f5f7fa; }
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
  background: rgba(0,0,0,.045); color: #303133;
}
</style>
