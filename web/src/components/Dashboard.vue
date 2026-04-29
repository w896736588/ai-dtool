<template>
  <div class="dashboard-container">
    <div class="chat-container">
      <!-- 消息列表区域 -->
      <div ref="messageList" class="message-list">
        <div class="welcome-message">
          <h2>命令快捷操作</h2>
          <div class="fixed-command-panel">
            <div class="fixed-command-title">支持的一级命令</div>
            <div class="fixed-command-list">
              <button
                v-for="(cmd, index) in availableCommands"
                :key="getCommandKey(cmd, index)"
                type="button"
                class="fixed-command-item"
                @click="quickSelectTopCommand(cmd)"
              >
                <span class="fixed-command-icon">{{ cmd.icon }}</span>
                <span class="fixed-command-name">{{ cmd.command }}</span>
                <span class="fixed-command-desc">{{ cmd.desc }}</span>
              </button>
            </div>
            <!-- <div v-if="topHistoryCommands.length > 0" class="history-command-section">
              <div class="fixed-command-title">高频历史命令 TOP 10</div>
              <div class="history-command-list">
                <div
                  v-for="(historyCmd, historyIndex) in topHistoryCommands"
                  :key="`history_${historyIndex}_${historyCmd}`"
                >
                  <div class="history-command-item-wrap">
                    <button
                      type="button"
                      class="history-command-item"
                      @click="quickSelectHistoryCommand(historyCmd)"
                    >
                      {{ historyCmd }}
                    </button>
                    <button
                      type="button"
                      class="history-command-delete"
                      title="删除该历史命令"
                      @click.stop="removeHistoryCommand(historyCmd)"
                    >×</button>
                  </div>
                </div>
              </div>
            </div> -->
          </div>
          <p class="hint">输入 <kbd>/</kbd> 或直接输入命令（如 <kbd>git</kbd>），<kbd>Tab</kbd> 补全，<kbd>Space</kbd> 继续</p>
        </div>
        <div
          v-for="(msg, index) in messages"
          :key="index"
          :class="['message', msg.type]"
        >
          <template v-if="hasCommandLayout(msg)">
            <div v-if="msg.resultText" class="message-content command-result-content">
              <div class="message-result-command">{{ msg.commandText }}</div>
              <div class="message-result-body">
                <div
                  v-for="(line, lineIndex) in getResultLines(msg.resultText)"
                  :key="`${index}_result_${lineIndex}`"
                  :class="['result-line', `result-line-${getResultLineState(line, lineIndex, getResultLines(msg.resultText))}`]"
                >
                  <span class="result-line-text">{{ line }}</span>
                  <span
                    v-if="getResultLineState(line, lineIndex, getResultLines(msg.resultText)) === 'running'"
                    class="result-line-dots"
                    aria-hidden="true"
                  >
                    <span></span><span></span><span></span>
                  </span>
                  <span
                    v-else-if="getResultLineState(line, lineIndex, getResultLines(msg.resultText)) === 'success'"
                    class="result-line-check"
                    aria-hidden="true"
                  >✓</span>
                  <span
                    v-else-if="getResultLineState(line, lineIndex, getResultLines(msg.resultText)) === 'failed'"
                    class="result-line-failed"
                    aria-hidden="true"
                  >✕</span>
                </div>
              </div>
            </div>
            <div v-if="msg.processText" class="process-window">
              <div class="process-title">执行过程 (SSE)</div>
              <div class="process-text markdown-body" v-html="renderProcessMarkdown(msg.processText)"></div>
            </div>
          </template>
          <div v-else class="message-content">{{ msg.content }}</div>
        </div>
      </div>

      <!-- 命令提示下拉框 -->
      <div ref="commandDropdown" v-show="showCommands" class="command-dropdown">
        <div class="command-breadcrumb" v-if="commandBreadcrumb">
          <span class="breadcrumb-text">{{ commandBreadcrumb }}</span>
        </div>
        <div v-if="isLoadingDynamic" class="command-loading">
          <span class="command-status command-status-running">
            <span class="command-status-spinner"></span>
          </span>
          <span class="command-loading-text">列表加载中...</span>
        </div>
        <div
          v-for="(cmd, index) in filteredCommands"
          :key="getCommandKey(cmd, index)"
          :ref="(el) => setCommandItemRef(el, index)"
          :class="['command-item', { active: activeCommandIndex === index }]"
        >
          <span class="command-icon">{{ cmd.icon }}</span>
          <span class="command-name">{{ cmd.name }}</span>
          <span
            v-if="!cmd.insertOnly && getCommandLevelUsageCount(cmd, filteredCommands) > 0"
            class="command-usage-count"
          >{{ getCommandLevelUsageCount(cmd, filteredCommands) }} 次</span>
          <span class="command-desc">
            {{ cmd.desc }}<template v-if="getCommandMatchHint(cmd)"> | 匹配: {{ getCommandMatchHint(cmd) }}</template>
          </span>
          <pl-button
            v-if="cmd.insertOnly"
            class="command-item-delete"
            link
            title="删除该历史命令"
            @mousedown.prevent="markKeepDropdownOnBlur"
            @click.stop="removeHistoryCommand(cmd.insertText || cmd.command || cmd.name)"
          >×</pl-button>
          <span v-if="cmd.children || cmd.needTarget" class="command-arrow">→</span>
        </div>
      </div>

      <!-- 输入区域 -->
      <div class="input-container">
        <div class="input-center-box" :style="{ width: inputWrapperWidth }">
          <div class="input-main-row">
            <div class="input-main-panel">
              <div class="input-wrapper">
            <div class="input-overlay-box">
              <div
                v-if="inputText"
                class="input-highlight-layer"
                v-html="highlightedInputHtml"
              ></div>
              <input
                ref="inputRef"
                v-model="inputText"
                type="text"
                :class="['chat-input', { 'chat-input-overlay': !!inputText }]"
                :placeholder="inputPlaceholder"
                @input="handleInput"
                @keydown="handleKeydown"
                @blur="handleBlur"
                @focus="handleFocus"
              />
            </div>
            <pl-button class="send-btn" :class="{ 'send-btn--executing': isExecuting }" type="primary" :disabled="!canSubmitCommand" @click="executeCommand">
              <span v-if="isExecuting" class="send-icon send-icon--spinning">↻</span>
              <span v-else class="send-icon">→</span>
            </pl-button>
              </div>
              <div class="next-step-tip">{{ nextStepHint }}</div>
            </div>
              <div v-if="hasPendingCommandQueue" class="pending-command-panel">
                <div class="pending-command-header">
                  <span class="pending-command-title">{{ pendingCommandTitle }}</span>
                  <span class="pending-command-count">{{ pendingCommandQueue.length }}</span>
                </div>
              <div class="pending-command-list pending-command-list--horizontal">
                  <div
                    v-for="item in pendingCommandQueue"
                    :key="item.id"
                    class="pending-command-item"
                >
                  <span class="pending-command-text" :title="item.rawCommand">{{ item.rawCommand }}</span>
                  <pl-button
                    class="pending-command-delete"
                    link
                    @click="removePendingCommand(item.id)"
                  >移除</pl-button>
                </div>
              </div>
            </div>
          </div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, nextTick, onMounted, onUnmounted, onActivated, watch } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import { ElMessageBox } from 'element-plus'
import module from '@/utils/module'
import commandConfig from '@/config/commandConfig.js'
import ssh from '@/utils/base/ssh_set'
import git from '@/utils/base/git'
import compose from '@/utils/base/compose'
import supervisor from '@/utils/base/supervisor'
import shellOut from '@/utils/base/shell_out'
import smartLinkSet from '@/utils/base/smart_link_set'
import variableSet from '@/utils/base/variable_set'
import store from '@/utils/base/store'
import sseDistribute from '@/utils/base/sse_distribute'
import { Throttle_string } from '@/utils/base/throttle_string'
import * as linkRunSelection from '@/utils/link_run_selection.cjs'
import pendingCommandQueueUtils from '@/utils/dashboard_command_queue.cjs'
import dashboardHistoryRankUtils from '@/utils/dashboard_history_rank.cjs'

export default {
  name: 'DashboardPage',
  setup() {
    const {
      buildLinkAccountOptionsFromEnv,
      buildLinkEnvOptionsFromConfig,
      buildLinkRunPayload,
      getLinkRunSelection,
      hasConfiguredLinkAccounts,
      isLinkRunSelectionComplete,
    } = linkRunSelection
    const {
      createPendingCommandItem,
      enqueuePendingCommand,
      dequeuePendingCommand,
      removePendingCommandById,
      consumeNextPendingCommand,
    } = pendingCommandQueueUtils
    const {
      buildTopHistoryCommands,
      normalizeHistoryCommandText,
    } = dashboardHistoryRankUtils

    marked.setOptions({
      gfm: true,
      breaks: true
    })

    const inputText = ref('')
    const messages = ref([])
    const showCommands = ref(false)
    const activeCommandIndex = ref(0)
    const inputRef = ref(null)
    const messageList = ref(null)
    const commandDropdown = ref(null)
    const commandItemRefs = ref([])
    const keepDropdownOnBlur = ref(false) // 删除历史候选时，避免 input blur 关闭下拉
    const suppressDropdownOnNextFocus = ref(false) // 历史命令回填后回焦输入框时，避免重新弹出候选
    
    // 多级命令状态
    const commandStack = ref([]) // 命令栈，存储已选择的命令
    const currentChildren = ref([]) // 当前可选的子命令
    const dynamicDataCache = ref({}) // 动态数据缓存
    const activeDynamicType = ref('') // 当前激活的动态列表类型（避免异步回调串台覆盖）
    const isLoadingDynamic = ref(false) // 是否正在加载动态数据
    const currentInputValue = ref('')
    const commandHistory = ref([]) // 命令历史记录
    const commandHistoryIndex = ref(0) // 命令历史游标（指向“下一条”位置）
    const commandUsageMap = ref({}) // 命令使用次数统计（key=命令文本，value=次数）
    const commandLevelUsageMap = ref({})
    const commandHistoryCacheKey = 'dashboard_command_history_v1'
    const commandUsageCacheKey = 'dashboard_command_usage_v1'
    const commandLevelUsageCacheKey = 'dashboard_command_level_usage_v1'
    const supervisorProcessCacheKeyPrefix = 'dashboard_supervisor_process_cache_v1'
    const supervisorProcessCacheExpireMs = 60 * 60 * 1000
    // 历史命令自动执行状态：选中历史项后，等待动态列表补齐并自动触发执行
    const pendingHistoryExecution = ref({
      active: false,
      commandText: '',
      reparsedTypeMap: {},
      autoExecute: false
    })
    const pendingCommandQueue = ref([])
    
    // SSE 相关状态
    const sseDistributeId = ref('') // SSE 分发 ID
    const isExecuting = ref(false) // 是否正在执行命令
    const currentOutputMessage = ref(null) // 当前输出消息的引用
    // script 会话状态（用于首页脚本执行多步交互）
    const scriptSession = ref({
      active: false,
      stage: 'idle',
      scriptId: 0,
      scriptName: '',
      runCmdId: 0,
      replaceList: {},
      currentForm: null,
      pendingInputLabel: '',
      optionList: [],
      canExecute: false,
    })
    const browserNotificationPermissionRequested = ref(false)

    // 首页命令待执行队列展示文案。
    // 执行中的新命令采用入队提示，避免误以为已立即执行。
    // 当前命令既不可执行也不可入队时，沿用执行中的阻塞提示。

    // 开放的模块列表
    const openModules = module.GetOpenModuleList()

    // 首页命令待执行队列展示文案。
    // 执行中的新命令采用入队提示，避免误以为已立即执行。
    // 当前命令既不可执行也不可入队时，沿用执行中的阻塞提示。

    // Queue labels for pending home commands.
    const QUEUE_PANEL_TITLE_TEXT = '待执行'
    const QUEUE_ENQUEUED_MESSAGE_TEXT = '命令已加入待执行列表，当前任务完成后自动执行\n'
    const QUEUE_RUNNING_MESSAGE_TEXT = '正在执行其他命令，请稍候...\n'

    const normalizeCommandPart = (value) => {
      if (value === null || value === undefined) return ''
      return String(value).trim()
    }

    // 获取当前待入队的完整命令文本，后续按该原始文本重新解析执行。
    const getPendingCommandText = () => {
      return String(inputText.value || '').trim()
    }

    // 入队后清理当前输入态，避免用户误以为该命令仍在编辑中。
    const clearCommandInputState = () => {
      inputText.value = ''
      showCommands.value = false
      commandStack.value = []
      currentChildren.value = []
      currentInputValue.value = ''
    }

    const getCommandKeywords = (cmd) => {
      const aliases = Array.isArray(cmd?.aliases) ? cmd.aliases : []
      return [
        normalizeCommandPart(cmd?.command).toLowerCase(),
        normalizeCommandPart(cmd?.name).toLowerCase(),
        normalizeCommandPart(cmd?.desc).toLowerCase(),
        ...aliases.map(alias => normalizeCommandPart(alias).toLowerCase())
      ].filter(Boolean)
    }

    // getSupervisorProcessCacheKey 按环境生成首页 Supervisor 服务列表缓存 key，避免不同环境串缓存。
    const getSupervisorProcessCacheKey = (envData) => {
      const id = normalizeCommandPart(envData?.id)
      const sshId = normalizeCommandPart(envData?.ssh_id)
      const dockerName = normalizeCommandPart(envData?.docker_name)
      const configDir = normalizeCommandPart(envData?.config_dir)
      const name = normalizeCommandPart(envData?.name)
      const suffix = [id, sshId, dockerName, configDir, name].filter(Boolean).join('__')
      return `${supervisorProcessCacheKeyPrefix}_${suffix || 'default'}`
    }

    // readSupervisorProcessCache 读取 Supervisor 服务列表缓存，仅返回 1 小时内的有效数据。
    const readSupervisorProcessCache = (envData) => {
      const cacheRaw = store.getStore(getSupervisorProcessCacheKey(envData))
      if (!cacheRaw) {
        return []
      }
      try {
        const cacheData = JSON.parse(cacheRaw)
        const cachedAt = Number(cacheData?.cachedAt || 0)
        const list = Array.isArray(cacheData?.list) ? cacheData.list : []
        if (!cachedAt || (Date.now() - cachedAt) > supervisorProcessCacheExpireMs) {
          store.removeStore(getSupervisorProcessCacheKey(envData))
          return []
        }
        return list
      } catch (error) {
        store.removeStore(getSupervisorProcessCacheKey(envData))
        return []
      }
    }

    // writeSupervisorProcessCache 写入 Supervisor 服务列表缓存，供首页重启/停止/查看配置命令复用。
    const writeSupervisorProcessCache = (envData, list) => {
      store.setStore(getSupervisorProcessCacheKey(envData), JSON.stringify({
        cachedAt: Date.now(),
        list: Array.isArray(list) ? list : []
      }))
    }

    // clearSupervisorProcessCache 清理指定环境的 Supervisor 服务列表缓存。
    const clearSupervisorProcessCache = (envData) => {
      store.removeStore(getSupervisorProcessCacheKey(envData))
    }

    // buildSupervisorProcessCommandList 把 SupervisorConfList 返回内容转换为首页可选命令列表。
    const buildSupervisorProcessCommandList = (envCmd, responseData) => {
      const lines = String(responseData || '')
        .split('\n')
        .map(line => normalizeCommandPart(line))
        .filter(Boolean)
      return lines.map((line, index) => {
        const [configNameRaw, supervisorNameRaw] = line.split('---')
        const configName = normalizeCommandPart(configNameRaw)
        let supervisorName = normalizeCommandPart(supervisorNameRaw)
          .replaceAll('[', '')
          .replaceAll(']', '')
          .replaceAll('program:', '')
        supervisorName = normalizeCommandPart(supervisorName)
        const configDir = normalizeCommandPart(envCmd?.data?.config_dir)
        const configPath = configDir && configName ? `${configDir}/${configName}` : configName
        const displayName = supervisorName || configName || `进程${index + 1}`
        return {
          command: displayName,
          name: displayName,
          aliases: [configName].filter(Boolean),
          desc: configName || '进程配置',
          id: `${envCmd?.id || envCmd?.data?.id || 'env'}_${index}`,
          data: {
            supervisor_name: supervisorName,
            supervisor_config: configPath
          }
        }
      })
    }

    // applySupervisorProcessListResult 把 Supervisor 服务列表同步到当前候选区与动态缓存。
    const applySupervisorProcessListResult = (list) => {
      dynamicDataCache.value['supervisorProcessList'] = list
      currentChildren.value = list
      reparseForPendingHistoryExecution('supervisorProcessList')
      refreshCommandDropdownVisibility()
    }

    // fetchSupervisorProcessList 拉取指定环境的 Supervisor 服务列表，支持强制刷新缓存。
    const fetchSupervisorProcessList = (envCmd, options = {}) => {
      const { forceRefresh = false, onSuccess, onError } = options
      if (!(envCmd && envCmd.data && envCmd.data.ssh_id)) {
        if (typeof onError === 'function') {
          onError()
        }
        return
      }
      if (!forceRefresh) {
        const cachedList = readSupervisorProcessCache(envCmd.data)
        if (cachedList.length > 0) {
          if (typeof onSuccess === 'function') {
            onSuccess(cachedList, true)
          }
          return
        }
      }
      const supervisorConfig = {
        ...envCmd.data
      }
      supervisor.SupervisorConfList(supervisorConfig, (response) => {
        if (!(response && response.ErrCode === 0)) {
          if (typeof onError === 'function') {
            onError(response)
          }
          return
        }
        const list = buildSupervisorProcessCommandList(envCmd, response.Data)
        writeSupervisorProcessCache(envCmd.data, list)
        if (typeof onSuccess === 'function') {
          onSuccess(list, false)
        }
      })
    }

    const getCommandMatchHint = (cmd) => {
      const aliases = Array.isArray(cmd?.aliases) ? cmd.aliases : []
      const tokens = [cmd?.command, ...aliases]
        .map(v => normalizeCommandPart(v).toLowerCase())
        .filter(v => /^[a-z][a-z0-9-]*$/.test(v))
      if (tokens.length === 0) return ''
      return [...new Set(tokens)].join(', ')
    }

    const findCommandByToken = (commands, token) => {
      const normalizedToken = normalizeCommandPart(token).toLowerCase()
      if (!normalizedToken) return null
      return commands.find(cmd => {
        const keywords = getCommandKeywords(cmd)
        return keywords.some(keyword => keyword === normalizedToken)
      }) || null
    }

    const getCommandKey = (cmd, index) => {
      if (cmd && cmd.id !== undefined && cmd.id !== null && String(cmd.id) !== '') {
        return `id:${cmd.id}`
      }
      if (cmd && cmd.command && cmd.path) {
        return `cp:${cmd.command}:${cmd.path}`
      }
      if (cmd && cmd.command) {
        return `c:${cmd.command}:${index}`
      }
      if (cmd && cmd.path) {
        return `p:${cmd.path}:${index}`
      }
      return `idx:${index}`
    }

    // getGitQuickCreateSelection 获取 git quick-create-branch 选择项（仓库/基线分支/分支类型）
    const getGitQuickCreateSelection = (stack) => {
      const sourceStack = Array.isArray(stack) ? stack : []
      const actionIndex = sourceStack.findIndex(item => item?.action === 'gitQuickCreateBranch')
      if (actionIndex < 0) {
        return {
          projectCmd: null,
          baseBranchCmd: null,
          branchTypeCmd: null
        }
      }
      return {
        projectCmd: sourceStack[actionIndex + 1] || null,
        baseBranchCmd: sourceStack[actionIndex + 2] || null,
        branchTypeCmd: sourceStack[actionIndex + 3] || null
      }
    }

    // 判断动作命令是否已满足执行条件（含多级目标校验）
    const isActionReady = (actionCmd, stack, inputValue) => {
      if (!actionCmd) return false
      const sourceStack = Array.isArray(stack) ? stack : []
      const actionIndex = sourceStack.findIndex(item => item?.action === actionCmd.action)
      const targetCmd = actionCmd.needTarget ? sourceStack[actionIndex + 1] : null
      const targetReady = !actionCmd.needTarget || !!(targetCmd && targetCmd.data)
      const inputReady = !actionCmd.needInput || !!normalizeCommandPart(inputValue)
      if (!targetReady || !inputReady) {
        return false
      }
      if (actionCmd.action === 'linkRun') {
        const selection = getLinkRunSelection(sourceStack)
        return isLinkRunSelectionComplete(selection)
      }
      // docker quick-restart/quick-stop 需要先选项目，再选服务
      if (actionCmd.action === 'dockerQuickRestart' || actionCmd.action === 'dockerQuickStop') {
        const actionIndex = sourceStack.findIndex(item => item?.action === actionCmd.action)
        const serviceCmd = actionIndex >= 0 ? sourceStack[actionIndex + 2] : null
        return !!(serviceCmd && serviceCmd.data)
      }
      if (actionCmd.action === 'supervisorRestart' || actionCmd.action === 'supervisorStop' || actionCmd.action === 'supervisorConfig') {
        const actionIndex = sourceStack.findIndex(item => item?.action === actionCmd.action)
        const processCmd = actionIndex >= 0 ? sourceStack[actionIndex + 2] : null
        return !!(processCmd && processCmd.data)
      }
      if (actionCmd.action === 'gitQuickCreateBranch') {
        const selection = getGitQuickCreateSelection(sourceStack)
        const businessEN = normalizeCommandPart(inputValue)
        const businessOk = /^[A-Za-z0-9_]+$/.test(businessEN)
        return !!(selection.projectCmd?.data && selection.baseBranchCmd?.data && selection.branchTypeCmd?.data && businessOk)
      }
      return true
    }

    // 获取动作命令未完成时的提示语
    const getActionIncompleteMessage = (actionCmd, stack, inputValue) => {
      if (!actionCmd) {
        return '命令未完成'
      }
      const sourceStack = Array.isArray(stack) ? stack : []
      const actionIndex = sourceStack.findIndex(item => item?.action === actionCmd.action)
      if (actionCmd.action === 'linkRun') {
        const selection = getLinkRunSelection(sourceStack)
        if (!selection.envCmd) {
          return '命令未完成：请选择要执行的环境'
        }
        if (hasConfiguredLinkAccounts(selection.envCmd) && !selection.accountCmd) {
          return '命令未完成：请选择账号'
        }
      }
      const targetCmd = actionCmd.needTarget ? sourceStack[actionIndex + 1] : null
      if (actionCmd.needTarget && !(targetCmd && targetCmd.data)) {
        return '命令未完成：请先选择项目/环境'
      }
      if (actionCmd.action === 'gitQuickCreateBranch') {
        const selection = getGitQuickCreateSelection(sourceStack)
        if (!(selection.projectCmd && selection.projectCmd.data)) {
          return '命令未完成：请先选择仓库'
        }
        if (!(selection.baseBranchCmd && selection.baseBranchCmd.data)) {
          return '命令未完成：请选择基于哪个分支创建'
        }
        if (!(selection.branchTypeCmd && selection.branchTypeCmd.data)) {
          return '命令未完成：请选择分支类型'
        }
        if (!/^[A-Za-z0-9_]+$/.test(normalizeCommandPart(inputValue))) {
          return '命令未完成：业务英文仅支持英文、数字、下划线'
        }
      }
      if (actionCmd.needInput && !normalizeCommandPart(inputValue)) {
        return `命令未完成：${actionCmd.inputPlaceholder || '请输入参数'}`
      }
      if (actionCmd.action === 'dockerQuickRestart' || actionCmd.action === 'dockerQuickStop') {
        const serviceCmd = actionIndex >= 0 ? sourceStack[actionIndex + 2] : null
        if (!(serviceCmd && serviceCmd.data)) {
          return '命令未完成：请选择服务'
        }
      }
      if (actionCmd.action === 'supervisorRestart' || actionCmd.action === 'supervisorStop' || actionCmd.action === 'supervisorConfig') {
        const processCmd = actionIndex >= 0 ? sourceStack[actionIndex + 2] : null
        if (!(processCmd && processCmd.data)) {
          return '命令未完成：请选择服务'
        }
      }
      return '命令未完成'
    }

    const isSameCommandItem = (a, b) => {
      if (!a || !b) return false
      const aId = normalizeCommandPart(a.id)
      const bId = normalizeCommandPart(b.id)
      if (aId && bId) {
        return aId === bId
      }
      const aCmd = normalizeCommandPart(a.command || a.name).toLowerCase()
      const bCmd = normalizeCommandPart(b.command || b.name).toLowerCase()
      const aPath = normalizeCommandPart(a.path).toLowerCase()
      const bPath = normalizeCommandPart(b.path).toLowerCase()
      return aCmd !== '' && aCmd === bCmd && aPath === bPath
    }

    const getCommandInputToken = (cmd) => {
      if (!cmd) return ''
      return normalizeCommandPart(cmd.__selectedInputToken || cmd.insertText || cmd.command || cmd.name)
    }

    const setCommandItemRef = (el, index) => {
      if (!Array.isArray(commandItemRefs.value)) {
        commandItemRefs.value = []
      }
      if (el) {
        commandItemRefs.value[index] = el
        return
      }
      if (index >= 0 && index < commandItemRefs.value.length) {
        commandItemRefs.value[index] = null
      }
    }

    const ensureActiveCommandVisible = () => {
      if (!showCommands.value) return
      const dropdownElement = commandDropdown.value
      if (!dropdownElement) return
      const activeElement = Array.isArray(commandItemRefs.value)
        ? commandItemRefs.value[activeCommandIndex.value]
        : null
      if (!activeElement || typeof activeElement.scrollIntoView !== 'function') return
      activeElement.scrollIntoView({ block: 'nearest' })
    }

    const parseTokens = (rawText) => {
      const text = String(rawText || '')
      const leftTrimmed = text.trimStart()
      const useSlash = leftTrimmed.startsWith('/')
      const withoutSlash = useSlash ? leftTrimmed.slice(1) : leftTrimmed
      const parts = withoutSlash.trim().length > 0
        ? withoutSlash.trim().split(/\s+/)
        : []
      return { useSlash, parts }
    }

    const getCommandLevelUsageKey = (scopeTokens, token) => {
      const normalizedScope = Array.isArray(scopeTokens)
        ? scopeTokens.map(item => normalizeCommandPart(item)).filter(Boolean)
        : []
      const normalizedToken = normalizeCommandPart(token)
      if (!normalizedToken) return ''
      return [...normalizedScope, normalizedToken].join(' > ').toLowerCase()
    }

    const getCommandLevelScopeTokens = () => {
      return commandStack.value
        .map(item => getCommandInputToken(item))
        .filter(Boolean)
    }

    const persistCommandLevelUsageCache = () => {
      store.setStore(commandLevelUsageCacheKey, JSON.stringify(commandLevelUsageMap.value || {}))
    }

    const recordCommandLevelUsage = (stack) => {
      const normalizedStack = Array.isArray(stack) ? stack : []
      const scopeTokens = []
      normalizedStack.forEach((item) => {
        if (!item || item.insertOnly) return
        const token = getCommandInputToken(item)
        if (!token) return
        const usageKey = getCommandLevelUsageKey(scopeTokens, token)
        if (!usageKey) return
        commandLevelUsageMap.value[usageKey] = (Number(commandLevelUsageMap.value[usageKey]) || 0) + 1
        scopeTokens.push(token)
      })
      persistCommandLevelUsageCache()
    }

    const getCommandLevelUsageCount = (cmd, commandList = currentChildren.value) => {
      if (!cmd || cmd.insertOnly) return 0
      const token = getCommandInputToken(cmd)
      if (!token) return 0
      const scopeTokens = Array.isArray(commandList) ? getCommandLevelScopeTokens() : []
      const usageKey = getCommandLevelUsageKey(scopeTokens, token)
      if (!usageKey) return 0
      return Number(commandLevelUsageMap.value[usageKey]) || 0
    }

    // 清理历史命令自动执行状态。
    const clearPendingHistoryExecution = () => {
      pendingHistoryExecution.value = {
        active: false,
        commandText: '',
        reparsedTypeMap: {},
        autoExecute: false
      }
    }

    // 历史命令选中后进入“待自动执行”状态。
    const markPendingHistoryExecution = (rawCommandText, options = {}) => {
      pendingHistoryExecution.value = {
        active: true,
        commandText: normalizeHistoryCommandText(rawCommandText),
        reparsedTypeMap: {},
        autoExecute: options.autoExecute === true
      }
    }

    // tryAutoExecutePendingHistory 当历史命令已经满足执行条件时自动执行。
    const tryAutoExecutePendingHistory = () => {
      if (!pendingHistoryExecution.value.active || isExecuting.value) {
        return false
      }
      const currentText = normalizeHistoryCommandText(String(inputText.value || '').trim())
      if (currentText !== pendingHistoryExecution.value.commandText) {
        clearPendingHistoryExecution()
        return false
      }
      if (!canExecuteCommand.value) {
        return false
      }
      const shouldAutoExecute = pendingHistoryExecution.value.autoExecute
      clearPendingHistoryExecution()
      if (shouldAutoExecute) {
        executeCommand()
      } else {
        showCommands.value = false
      }
      return true
    }

    // reparseForPendingHistoryExecution 仅在历史自动执行场景重解析并尝试自动执行。
    const reparseForPendingHistoryExecution = (dynamicType = '') => {
      if (!pendingHistoryExecution.value.active) {
        return
      }
      // 同一动态类型仅重解析一次，避免“未命中候选”时持续触发循环请求。
      const typeKey = normalizeCommandPart(dynamicType)
      if (typeKey) {
        const reparsedMap = pendingHistoryExecution.value.reparsedTypeMap || {}
        if (reparsedMap[typeKey]) {
          return
        }
        reparsedMap[typeKey] = 1
        pendingHistoryExecution.value.reparsedTypeMap = reparsedMap
      }
      parseInput()
      tryAutoExecutePendingHistory()
    }

    // loadHistoryCommandForExecution 统一加载历史命令，可选择仅进入可执行态或自动执行。
    const loadHistoryCommandForExecution = (historyCommand, options = {}) => {
      const commandText = normalizeHistoryCommandText(historyCommand)
      if (!commandText) return
      const autoExecute = options.autoExecute === true
      markPendingHistoryExecution(commandText, { autoExecute })
      // 补一个空格，确保 parseInput 将最后一个 token 视为已确认输入。
      inputText.value = /\s$/.test(commandText) ? commandText : `${commandText} `
      parseInput()
      showCommands.value = false
      activeCommandIndex.value = getDefaultActiveCommandIndex()
      suppressDropdownOnNextFocus.value = true
      nextTick(() => {
        inputRef.value?.focus()
      })
      tryAutoExecutePendingHistory()
    }

    // startHistoryCommandExecution 统一处理“选中历史命令后直接执行”流程。
    const startHistoryCommandExecution = (historyCommand) => {
      loadHistoryCommandForExecution(historyCommand, { autoExecute: true })
    }

    const escapeHtml = (value) => {
      return String(value || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
    }

    // resolveNextTargetOptions 解析“已选目标”后的下一层候选，支持多级目标命令高亮
    const resolveNextTargetOptions = (parentCmd, targetCmd) => {
      if (!targetCmd) return []
      if (targetCmd.dynamicChildren) {
        const directOptions = dynamicDataCache.value[targetCmd.dynamicChildren] || []
        if (Array.isArray(directOptions) && directOptions.length > 0) {
          return directOptions
        }
        if (targetCmd.dynamicChildren === 'linkAccountList') {
          return buildLinkAccountOptionsFromEnv(targetCmd)
        }
      }
      if (parentCmd?.nextDynamicChildren) {
        const nextOptions = dynamicDataCache.value[parentCmd.nextDynamicChildren] || []
        if (Array.isArray(nextOptions) && nextOptions.length > 0) {
          return nextOptions
        }
        if (parentCmd.nextDynamicChildren === 'linkEnvList') {
          return buildLinkEnvOptionsFromConfig(targetCmd)
        }
      }
      return []
    }

    const renderProcessMarkdown = (text) => {
      const raw = String(text || '')
      // 终端输出里常见 "~/path" 提示符，Markdown 会把 ~...~ 识别为删除线；先转义 ~ 防止误渲染。
      const escapedRaw = raw.replace(/~/g, '\\~')
      const html = marked.parse(escapedRaw)
      return DOMPurify.sanitize(html)
    }

    const hasCommandLayout = (msg) => {
      return !!(msg && (msg.commandText !== undefined || msg.resultText !== undefined || msg.processText !== undefined))
    }

    const isCommandModeByText = (rawText) => {
      const tokenInfo = parseTokens(rawText)
      if (tokenInfo.useSlash) return true
      if (tokenInfo.parts.length === 0) return false
      const first = normalizeCommandPart(tokenInfo.parts[0]).toLowerCase()
      return availableCommands.value.some(cmd => {
        const keywords = getCommandKeywords(cmd)
        return keywords.some(keyword => keyword.includes(first))
      })
    }

    // isHistoryPrefixSearchMode 判断当前是否处于“history 前缀搜索历史命令”模式。
    // 该模式下空格仅用于继续筛选，不应自动确认候选。
    const isHistoryPrefixSearchMode = () => {
      if (!Array.isArray(commandStack.value) || commandStack.value.length === 0) {
        return false
      }
      const firstCmd = commandStack.value[0]
      return !!(firstCmd && firstCmd.dynamicChildren === 'historyList')
    }

    const isScriptSessionMode = () => {
      return !!(scriptSession.value?.active && scriptSession.value?.stage && scriptSession.value.stage !== 'idle')
    }

    // 判断当前命令是否允许进入待执行队列：仅普通命令链路支持排队。
    const canQueueCurrentCommand = () => {
      if (!isExecuting.value || isScriptSessionMode()) {
        return false
      }
      const actionCmd = commandStack.value.find(item => item.action)
      if (!actionCmd) {
        return false
      }
      return isActionReady(actionCmd, commandStack.value, currentInputValue.value)
    }

    // 当前已有命令执行中时，将新命令排到待执行队列末尾。
    const queuePendingCommand = () => {
      const rawCommand = getPendingCommandText()
      if (!rawCommand) {
        return false
      }
      const pendingItem = createPendingCommandItem(rawCommand)
      pendingCommandQueue.value = enqueuePendingCommand(pendingCommandQueue.value, pendingItem)
      messages.value.push({
        type: 'system',
        content: QUEUE_ENQUEUED_MESSAGE_TEXT
      })
      clearCommandInputState()
      scrollToBottom()
      return true
    }

    // 删除指定待执行命令，保留其余排队顺序不变。
    const removePendingCommand = async (pendingCommandId) => {
      try {
        await ElMessageBox.confirm('确认移除这条待执行命令吗？', '提示', {
          confirmButtonText: '移除',
          cancelButtonText: '取消',
          type: 'warning',
        })
      } catch (error) {
        return
      }
      pendingCommandQueue.value = removePendingCommandById(pendingCommandQueue.value, pendingCommandId)
    }

    // 当前命令结束后，自动从队列取下一条并走原有解析执行链路。
    const executePendingCommandText = (rawCommand) => {
      const pendingRawCommand = normalizeCommandPart(rawCommand)
      if (!pendingRawCommand) {
        return false
      }
      inputText.value = pendingRawCommand
      if (isCommandModeByText(pendingRawCommand)) {
        parseInput()
      }
      if (!canExecuteCommand.value) {
        messages.value.push({
          type: 'system',
          content: `待执行命令自动执行失败：${pendingRawCommand}\n`
        })
        clearCommandInputState()
        scrollToBottom()
        return false
      }
      executeCommand()
      return true
    }

    const triggerNextPendingCommand = () => {
      if (isExecuting.value) {
        return false
      }
      const dequeueResult = consumeNextPendingCommand(
        pendingCommandQueue.value,
        (rawCommand) => {
          loadHistoryCommandForExecution(rawCommand, { autoExecute: true })
        }
      )
      if (!dequeueResult.item) {
        return false
      }
      pendingCommandQueue.value = dequeueResult.queue
      return true
    }

    const getDefaultActiveCommandIndex = (commandList = filteredCommands.value) => {
      const normalizedList = Array.isArray(commandList) ? commandList : []
      if (normalizedList.length === 0) return 0
      if (normalizedList.every(item => item && item.insertOnly)) {
        return Math.max(normalizedList.length - 1, 0)
      }
      let maxCount = -1
      let selectedIndex = 0
      normalizedList.forEach((item, index) => {
        const count = getCommandLevelUsageCount(item, normalizedList)
        if (count > maxCount) {
          maxCount = count
          selectedIndex = index
        }
      })
      return maxCount > 0 ? selectedIndex : 0
    }

    const refreshCommandDropdownVisibility = () => {
      if (isScriptSessionMode()) {
        const stage = normalizeCommandPart(scriptSession.value.stage)
        showCommands.value = (
          stage === 'selecting_script' ||
          stage === 'waiting_option'
        ) && (currentChildren.value.length > 0 || isLoadingDynamic.value)
        if (showCommands.value) {
          activeCommandIndex.value = getDefaultActiveCommandIndex()
        }
        return
      }
      showCommands.value = isCommandModeByText(inputText.value) &&
        (currentChildren.value.length > 0 || isLoadingDynamic.value) &&
        !isCommandReadyToExecute()
      if (showCommands.value) {
        activeCommandIndex.value = getDefaultActiveCommandIndex()
      }
    }

    // 更新当前命令状态（running/success/failed）
    const updateCurrentCommandStatus = (status) => {
      if (!currentOutputMessage.value || !status) return
      currentOutputMessage.value.commandStatus = status
    }

    const applyResponseCommandStatus = (response) => {
      if (!response || response.ErrCode === undefined || response.ErrCode === null) return
      if (Number(response?.ErrCode) === 1) {
        updateCurrentCommandStatus('failed')
        return
      }
      if (Number(response?.ErrCode) === 0) {
        updateCurrentCommandStatus('success')
      }
    }

    // 根据输出文本推断状态，并移除独立“执行成功/执行失败”行，避免重复展示
    const parseResultTextAndStatus = (rawText) => {
      let text = String(rawText || '')
      let status = ''
      const statusFromLine = []
      text = text.replace(/(^|\n)\s*执行成功\s*(?=\n|$)/g, (match, prefix) => {
        statusFromLine.push('success')
        return prefix
      })
      text = text.replace(/(^|\n)\s*执行失败\s*(?=\n|$)/g, (match, prefix) => {
        statusFromLine.push('failed')
        return prefix
      })
      if (statusFromLine.length > 0) {
        status = statusFromLine[statusFromLine.length - 1]
      } else if (/执行失败|错误[:：]/.test(rawText)) {
        status = 'failed'
      }
      return { text, status }
    }

    const appendOutputResult = (text) => {
      if (!currentOutputMessage.value) return
      const parsed = parseResultTextAndStatus(String(text || ''))
      if (parsed.status) {
        updateCurrentCommandStatus(parsed.status)
      }
      const current = String(currentOutputMessage.value.resultText || '')
      const merged = sanitizeCommandOutput(current + parsed.text)
      currentOutputMessage.value.resultText = merged.length > 50000 ? merged.slice(-50000) : merged
      scrollToBottom()
    }

    const appendOutputSummary = (text) => {
      if (!currentOutputMessage.value) return
      const parsed = parseResultTextAndStatus(String(text || ''))
      if (parsed.status) {
        updateCurrentCommandStatus(parsed.status)
      }
      const summary = sanitizeCommandOutput(parsed.text).trim()
      if (!summary) return
      const current = String(currentOutputMessage.value.resultText || '')
      const nextText = current && !current.endsWith('\n') ? `${current}\n${summary}` : `${current}${summary}`
      const merged = sanitizeCommandOutput(nextText)
      currentOutputMessage.value.resultText = merged.length > 50000 ? merged.slice(-50000) : merged
      scrollToBottom()
    }

    const appendOutputProcess = (text) => {
      if (!currentOutputMessage.value) return
      const current = String(currentOutputMessage.value.processText || '')
      const merged = sanitizeCommandOutput(current + String(text || ''))
      currentOutputMessage.value.processText = merged.length > 50000 ? merged.slice(-50000) : merged
      scrollToBottom()
    }

    const getResultLines = (text) => {
      return String(text || '')
        .split('\n')
        .map(line => String(line))
        .filter(line => line.trim() !== '')
    }

    const getResultLineState = (line, lineIndex = 0, lines = []) => {
      const text = normalizeCommandPart(line)
      if (!text) return 'default'
      const sourceLines = Array.isArray(lines) ? lines : []
      const hasTerminalLineAfter = sourceLines.slice(lineIndex + 1).some(item => {
        const nextText = normalizeCommandPart(item)
        return /完成$/.test(nextText) || /成功$/.test(nextText) || /失败$/.test(nextText) || /^执行失败[:：]/.test(nextText) || /^错误[:：]/.test(nextText) || /已打开/.test(nextText)
      })
      if (/^正在/.test(text)) {
        if (hasTerminalLineAfter) {
          return 'default'
        }
        return 'running'
      }
      if (/^执行失败[:：]/.test(text) || /^错误[:：]/.test(text) || /失败$/.test(text)) {
        return 'failed'
      }
      if (/完成$/.test(text) || /成功$/.test(text) || /已打开/.test(text)) {
        return 'success'
      }
      return 'default'
    }

    const sanitizeCommandOutput = (rawText) => {
      return String(rawText || '')
    }

    // 根据模块配置过滤可用命令
    const availableCommands = computed(() => {
      return commandConfig.filter(cmd => {
        if (cmd.module === null) return true
        return openModules.includes(cmd.module)
      })
    })

    // topHistoryCommands 首页仅展示使用率最高的 10 个历史命令。
    // topHistoryCommands shows only the top 10 most frequently used history commands on the home panel.
    const topHistoryCommands = computed(() => {
      return buildTopHistoryCommands({
        historyList: commandHistory.value,
        usageMap: commandUsageMap.value,
        limit: 10,
      })
    })

    // 命令面包屑导航
    const commandBreadcrumb = computed(() => {
      if (isScriptSessionMode()) {
        const session = scriptSession.value || {}
        return session.scriptName
          ? `script > ${session.scriptName}`
          : 'script'
      }
      if (commandStack.value.length === 0) return ''
      return commandStack.value.map(c => c.name).join(' > ')
    })

    // 输入框提示
    const inputPlaceholder = computed(() => {
      if (isScriptSessionMode()) {
        const session = scriptSession.value || {}
        if (session.stage === 'selecting_script') {
          return '请选择要执行的脚本'
        }
        if (session.stage === 'waiting_input') {
          return normalizeCommandPart(session.pendingInputLabel) || '请在命令框输入内容并回车'
        }
        if (session.stage === 'waiting_option') {
          return normalizeCommandPart(session.pendingInputLabel) || '请在命令框选择一个选项'
        }
        if (session.stage === 'ready_execute') {
          return '脚本已就绪，按回车执行'
        }
        if (session.stage === 'executing') {
          return '脚本执行中，请稍候...'
        }
      }
      if (commandStack.value.length === 0) {
        return '输入 / 或直接输入命令（如 git），Tab 补全，Space 继续...'
      }
      const lastCmd = commandStack.value[commandStack.value.length - 1]
      const actionCmd = commandStack.value.find(item => item.action)
      if (actionCmd && actionCmd.action === 'linkRun') {
        const selection = getLinkRunSelection(commandStack.value)
        if (!selection.envCmd) {
          return '请选择要执行的环境...'
        }
        if (hasConfiguredLinkAccounts(selection.envCmd) && !selection.accountCmd) {
          return '请选择账号...'
        }
        return '按 Enter 执行命令'
      }
      if (actionCmd && actionCmd.needInput) {
        const actionIndex = commandStack.value.findIndex(item => item.action)
        const targetReady = !actionCmd.needTarget || !!(commandStack.value[actionIndex + 1] && commandStack.value[actionIndex + 1].data)
        if (targetReady && !currentInputValue.value) {
          return actionCmd.inputPlaceholder || '请输入参数...'
        }
      }
      if (lastCmd.needInput) {
        return lastCmd.inputPlaceholder || '请输入...'
      }
      if (lastCmd.needTarget) {
        return '选择目标...'
      }
      if (currentInputValue.value && lastCmd.action) {
        return '按 Enter 执行命令'
      }
      return '继续输入或选择...'
    })

    // 过滤后的命令列表
    // compareCommandByNaturalAsc 命令自然升序：前缀短词优先（如 restart 在 restart-all 前）。
    const compareCommandByNaturalAsc = (a, b) => {
      const aText = normalizeCommandPart(a?.command || a?.name).toLowerCase()
      const bText = normalizeCommandPart(b?.command || b?.name).toLowerCase()
      if (aText === bText) return 0
      if (!aText) return 1
      if (!bText) return -1
      if (aText.startsWith(bText)) return 1
      if (bText.startsWith(aText)) return -1
      return aText.localeCompare(bText, 'zh-Hans-CN', { numeric: true, sensitivity: 'base' })
    }

    const sortCommandsByLevelUsage = (commandList) => {
      const normalizedList = Array.isArray(commandList) ? commandList : []
      if (normalizedList.every(item => item && item.insertOnly)) {
        return [...normalizedList]
      }
      // usage-enriched
      return normalizedList
        .map((cmd) => ({
          cmd,
          count: getCommandLevelUsageCount(cmd, commandList)
        }))
        .sort((a, b) => {
          if (a.count !== b.count) {
            return a.count - b.count
          }
          return compareCommandByNaturalAsc(a.cmd, b.cmd)
        })
        .map(item => item.cmd)
    }

    const filteredCommands = computed(() => {
      if (isScriptSessionMode()) {
        const sourceList = Array.isArray(currentChildren.value) ? currentChildren.value : []
        const searchText = normalizeCommandPart(inputText.value).toLowerCase()
        const sortedList = sortCommandsByLevelUsage(sourceList)
        if (!searchText) {
          return sortedList
        }
        return sortedList.filter(cmd => {
          const keywords = getCommandKeywords(cmd)
          return keywords.some(keyword => keyword.includes(searchText))
        })
      }
      let commands = currentChildren.value.length > 0 
        ? currentChildren.value
        : (commandStack.value.length === 0 ? availableCommands.value : [])
      
      // 获取当前输入的搜索文本
      const tokenInfo = parseTokens(inputText.value)
      const parts = tokenInfo.parts
      const hasTrailingSpace = /\s$/.test(String(inputText.value || ''))
      const rawSearchText = parts.length > 0
        ? normalizeCommandPart(parts[parts.length - 1]).toLowerCase().replace('/', '')
        : ''
      let searchText = hasTrailingSpace ? '' : rawSearchText

      // 场景：已完整输入动作词（如 git checkout），当前候选已切到“目标列表”
      // 这时不应再用动作词过滤目标，否则会把项目列表全部过滤为空。
      if (commandStack.value.length > 0 && commands.length > 0) {
        const lastCmd = commandStack.value[commandStack.value.length - 1]
        if (!hasTrailingSpace && lastCmd?.needTarget) {
          const lastCmdKeywords = getCommandKeywords(lastCmd)
          if (lastCmdKeywords.some(keyword => keyword === rawSearchText)) {
            searchText = ''
          }
        }
      }
      
      const sortedCommands = sortCommandsByLevelUsage(commands)

      // history 前缀模式：使用“history 后整段输入”做历史命令搜索，支持空格短语筛选。
      if (isHistoryPrefixSearchMode()) {
        const historyQuery = normalizeCommandPart(parts.slice(1).join(' ')).toLowerCase()
        if (!historyQuery) {
          return sortedCommands
        }
        const queryTokens = historyQuery.split(/\s+/).filter(Boolean)
        return sortedCommands.filter(cmd => {
          const candidate = normalizeCommandPart(cmd?.command || cmd?.name).toLowerCase()
          if (!candidate) return false
          if (candidate.includes(historyQuery)) return true
          return queryTokens.every(token => candidate.includes(token))
        })
      }

      if (!searchText) {
        return sortedCommands
      }
      
      return sortedCommands.filter(cmd => {
        const keywords = getCommandKeywords(cmd)
        return keywords.some(keyword => keyword.includes(searchText))
      })
    })

    const commandAnalysis = computed(() => {
      const rawText = String(inputText.value || '')
      const hasText = !!rawText.trim()
      const chunks = rawText.match(/\S+|\s+/g) || []

      if (!hasText) {
        return {
          canExecute: false,
          highlightedTokens: []
        }
      }

      const inCommandMode = isCommandModeByText(rawText)
      if (!inCommandMode) {
        return {
          canExecute: false,
          highlightedTokens: chunks.map(chunk => ({ text: chunk, type: 'plain' }))
        }
      }

      const highlightedTokens = []
      let currentLevel = availableCommands.value
      let waitingTargetOptions = null
      let waitingTargetParent = null
      let waitingForInput = false

      for (let tokenIndex = 0; tokenIndex < chunks.length; tokenIndex += 1) {
        const chunk = chunks[tokenIndex]
        if (/^\s+$/.test(chunk)) {
          highlightedTokens.push({ text: chunk, type: 'plain' })
          continue
        }

        const tokenRaw = chunk
        let normalized = normalizeCommandPart(tokenRaw).toLowerCase()
        if (highlightedTokens.filter(item => !/^\s+$/.test(item.text)).length === 0 && normalized.startsWith('/')) {
          normalized = normalizeCommandPart(normalized.slice(1))
        }

        if (!normalized) {
          highlightedTokens.push({ text: tokenRaw, type: 'plain' })
          continue
        }

        if (waitingForInput) {
          highlightedTokens.push({ text: tokenRaw, type: 'argument' })
          continue
        }

        if (waitingTargetOptions) {
          const targetFound = findCommandByToken(waitingTargetOptions, normalized)
          if (targetFound) {
            highlightedTokens.push({ text: tokenRaw, type: 'matched' })
            const nextOptions = resolveNextTargetOptions(waitingTargetParent, targetFound)
            if (Array.isArray(nextOptions) && nextOptions.length > 0) {
              waitingTargetOptions = nextOptions
              waitingTargetParent = targetFound
            } else {
              waitingTargetOptions = null
              waitingTargetParent = null
            }
          } else {
            highlightedTokens.push({ text: tokenRaw, type: 'invalid' })
          }
          continue
        }

        const found = findCommandByToken(currentLevel, normalized)
        if (!found) {
          highlightedTokens.push({ text: tokenRaw, type: 'invalid' })
          continue
        }

        highlightedTokens.push({ text: tokenRaw, type: 'matched' })

        if (found.needInput) {
          waitingForInput = true
        }

        if (found.needTarget) {
          waitingTargetParent = found
          waitingTargetOptions = found.dynamicChildren
            ? (dynamicDataCache.value[found.dynamicChildren] || [])
            : (found.children || [])
        }

        if (found.children && found.children.length > 0) {
          currentLevel = found.children
        } else if (found.dynamicChildren) {
          currentLevel = dynamicDataCache.value[found.dynamicChildren] || []
        } else {
          currentLevel = []
        }
      }

      const actionCmd = commandStack.value.find(item => item.action)
      if (!actionCmd || isExecuting.value) {
        return { canExecute: false, highlightedTokens }
      }

      const actionIndex = commandStack.value.findIndex(item => item.action)
      const targetCmd = actionCmd.needTarget ? commandStack.value[actionIndex + 1] : null
      const targetReady = !actionCmd.needTarget || !!(targetCmd && targetCmd.data)
      const inputReady = !actionCmd.needInput || !!normalizeCommandPart(currentInputValue.value)

      return {
        canExecute: targetReady && inputReady && isActionReady(actionCmd, commandStack.value, currentInputValue.value),
        highlightedTokens
      }
    })

    const canExecuteCommand = computed(() => {
      if (isScriptSessionMode()) {
        const stage = normalizeCommandPart(scriptSession.value.stage)
        if (stage === 'selecting_script' || stage === 'waiting_option') {
          return !!filteredCommands.value[activeCommandIndex.value]
        }
        if (stage === 'waiting_input') {
          return !!normalizeCommandPart(inputText.value)
        }
        if (stage === 'ready_execute') {
          return !!scriptSession.value.canExecute && !isExecuting.value
        }
        return false
      }
      return commandAnalysis.value.canExecute
    })

    const canSubmitCommand = computed(() => {
      if (canExecuteCommand.value) {
        return true
      }
      return canQueueCurrentCommand()
    })

    const hasPendingCommandQueue = computed(() => {
      return pendingCommandQueue.value.length > 0
    })

    const pendingCommandTitle = computed(() => {
      return QUEUE_PANEL_TITLE_TEXT
    })

    const highlightedInputHtml = computed(() => {
      return commandAnalysis.value.highlightedTokens.map(item => {
        const safe = escapeHtml(item.text)
        if (item.type === 'matched') {
          return `<span class="token-bg token-bg-valid">${safe}</span>`
        }
        if (item.type === 'invalid') {
          return `<span class="token-bg token-bg-invalid">${safe}</span>`
        }
        if (item.type === 'argument') {
          return `<span class="token-bg token-bg-arg">${safe}</span>`
        }
        return `<span>${safe}</span>`
      }).join('')
    })

    // 判断当前命令是否已满足执行条件（满足后不应再展示候选）
    const isCommandReadyToExecute = () => {
      const actionCmd = commandStack.value.find(item => item.action)
      if (!actionCmd) return false
      return isActionReady(actionCmd, commandStack.value, currentInputValue.value)
    }

    // 获取 script 会话的下一步提示语（首页多步命令）
    const getScriptSessionStepHint = () => {
      const session = scriptSession.value || {}
      if (!session.active) {
        return ''
      }
      if (session.stage === 'selecting_script') {
        return '下一步：请选择要执行的脚本'
      }
      if (session.stage === 'waiting_input') {
        return `下一步：${normalizeCommandPart(session.pendingInputLabel) || '请在命令框输入内容并回车'}`
      }
      if (session.stage === 'waiting_option') {
        return `下一步：${normalizeCommandPart(session.pendingInputLabel) || '请在命令框选择一个选项'}`
      }
      if (session.stage === 'ready_execute') {
        return '下一步：脚本已就绪，按回车执行'
      }
      if (session.stage === 'executing') {
        return '下一步：脚本执行中，请稍候...'
      }
      return ''
    }

    // 计算首页命令行的下一步浅色提示文案
    const nextStepHint = computed(() => {
      if (isExecuting.value) {
        return '正在执行命令，请稍候...'
      }

      const scriptHint = getScriptSessionStepHint()
      if (scriptHint) {
        return scriptHint
      }

      if (commandStack.value.length === 0) {
        const currentText = normalizeCommandPart(inputText.value)
        if (!currentText) {
          return '下一步：输入 / 查看可用命令'
        }
        if (isCommandModeByText(inputText.value)) {
          return '下一步：从下拉列表选择命令，或继续输入补全'
        }
        return '下一步：输入 / 进入命令模式'
      }

      if (isHistoryPrefixSearchMode()) {
        return '下一步：继续输入筛选历史命令，按 Tab 选中后再执行'
      }

      const actionCmd = commandStack.value.find(item => item.action)
      if (actionCmd) {
        if (actionCmd.action === 'gitQuickCreateBranch') {
          if (isActionReady(actionCmd, commandStack.value, currentInputValue.value)) {
            return '当前步骤：参数已完整，按 Enter 执行命令'
          }
          const quickCreateMessage = normalizeCommandPart(getActionIncompleteMessage(actionCmd, commandStack.value, currentInputValue.value))
          if (quickCreateMessage) {
            return `当前步骤：${quickCreateMessage.replace(/^命令未完成[:：]?/, '').trim()}`
          }
          return '当前步骤：继续补全快捷建分支命令'
        }
        if (isActionReady(actionCmd, commandStack.value, currentInputValue.value)) {
          return '下一步：按 Enter 执行命令'
        }
        const incompleteMessage = normalizeCommandPart(getActionIncompleteMessage(actionCmd, commandStack.value, currentInputValue.value))
        if (incompleteMessage) {
          return `下一步：${incompleteMessage.replace(/^命令未完成[:：]?/, '').trim()}`
        }
        return '下一步：继续补全当前命令'
      }

      const lastCmd = commandStack.value[commandStack.value.length - 1]
      if (lastCmd?.needTarget) {
        return '下一步：请选择目标'
      }
      if (lastCmd?.children?.length > 0 || (lastCmd?.dynamicChildren && currentChildren.value.length > 0)) {
        return '下一步：请选择下一级命令'
      }
      if (lastCmd?.needInput) {
        return `下一步：${lastCmd.inputPlaceholder || '请输入参数'}`
      }
      return '下一步：继续输入，完成后按 Enter 执行'
    })

    // 根据输入内容计算输入区宽度：最小 520，内容变长时扩展，最大不超过 1100 且不超过容器 92%
    const inputWrapperWidth = computed(() => {
      const minWidth = 520
      const maxWidth = 1100
      const contentLength = String(inputText.value || '').length
      const placeholderLength = String(inputPlaceholder.value || '').length
      const effectiveLength = Math.max(contentLength, Math.min(placeholderLength, 40))
      const estimatedWidth = 220 + (effectiveLength * 9)
      const targetWidth = Math.max(minWidth, Math.min(maxWidth, estimatedWidth))
      return `min(92%, ${targetWidth}px)`
    })

    // 解析输入文本，获取当前命令层级
    const parseInput = () => {
      if (isScriptSessionMode()) {
        activeCommandIndex.value = getDefaultActiveCommandIndex()
        refreshCommandDropdownVisibility()
        return
      }
      if (!isCommandModeByText(inputText.value)) {
        commandStack.value = []
        currentChildren.value = []
        currentInputValue.value = ''
        showCommands.value = false
        return
      }

      const tokenInfo = parseTokens(inputText.value)
      const parts = tokenInfo.parts
      const hasTrailingSpace = /\s$/.test(String(inputText.value || ''))
      
      // 重置状态
      commandStack.value = []
      currentChildren.value = []
      currentInputValue.value = ''
      
      let currentLevel = availableCommands.value
      
      for (let i = 0; i < parts.length; i++) {
        const part = parts[i].toLowerCase()
        const isLastPart = i === parts.length - 1

        // 最后一个 token 且没有尾随空格：视为“正在输入中”，仅用于筛选，不进入下一层。
        if (isLastPart && !hasTrailingSpace) {
          if (commandStack.value.length > 0) {
            const lastCmd = commandStack.value[commandStack.value.length - 1]
            if (lastCmd.needInput) {
              currentInputValue.value = parts.slice(i).join(' ')
              currentChildren.value = []
              break
            }
            if (lastCmd.needTarget) {
              if (lastCmd.dynamicChildren) {
                loadDynamicChildren(lastCmd.dynamicChildren)
                currentChildren.value = dynamicDataCache.value[lastCmd.dynamicChildren] || []
              } else if (lastCmd.children && lastCmd.children.length > 0) {
                currentChildren.value = lastCmd.children
              } else {
                currentChildren.value = currentLevel
              }
              break
            }
          }
          currentChildren.value = currentLevel
          break
        }

        const found = findCommandByToken(currentLevel, part)
        
        if (found) {
          commandStack.value.push(found)
          
          // 如果有子命令，继续
          if (found.children && found.children.length > 0) {
            currentLevel = found.children
            currentChildren.value = found.children
            continue
          }
          // 如果需要动态子命令
          if (found.dynamicChildren) {
            loadDynamicChildren(found.dynamicChildren)
            const dynamicList = dynamicDataCache.value[found.dynamicChildren] || []
            currentChildren.value = dynamicList
            // history 命令只做筛选，不做“空格自动确认目标”；必须通过 Tab 才选中历史项。
            if (found.dynamicChildren === 'historyList') {
              break
            }
            const targetToken = parts[i + 1]
            if (targetToken) {
              const targetFound = findCommandByToken(dynamicList, targetToken)
              // 仅在“已确认目标”的场景进入下一层：
              // 1) 有尾随空格；2) 目标后还有其他 token（例如继续输入参数）
              // 注意：仅精确命中但未按空格/Tab，不视为已确认，避免 common 被提前选中。
              const targetIsConfirmed = hasTrailingSpace || (parts.length > i + 2)
              if (targetFound && targetIsConfirmed) {
                commandStack.value.push(targetFound)
                i += 1
                if (found.nextDynamicChildren) {
                  // 例如 docker quick-restart/quick-stop：选完项目后继续选择服务
                  loadDynamicChildren(found.nextDynamicChildren)
                  const nextDynamicList = dynamicDataCache.value[found.nextDynamicChildren] || []
                  currentChildren.value = nextDynamicList
                  showCommands.value = true
                  // 支持一次性输入完整命令：docker quick-restart <项目> <服务>
                  const serviceToken = normalizeCommandPart(parts[i + 1]).toLowerCase()
                  if (serviceToken) {
                    const serviceFound = findCommandByToken(nextDynamicList, serviceToken)
                    const serviceIsConfirmed = hasTrailingSpace || (parts.length > i + 2)
                    if (serviceFound && serviceIsConfirmed) {
                      commandStack.value.push(serviceFound)
                      i += 1
                      // 支持继续向下解析第三级目标（如 link run <配置> <环境> <账号>）
                      // 同时兼容“二级目标选中后，通过 nextDynamicChildren 进入下一级”的场景（如 git quick-create-branch）。
                      const thirdDynamicKey = serviceFound.nextDynamicChildren || serviceFound.dynamicChildren
                      if (thirdDynamicKey) {
                        loadDynamicChildren(thirdDynamicKey)
                        const thirdDynamicList = dynamicDataCache.value[thirdDynamicKey] || []
                        currentChildren.value = thirdDynamicList
                        showCommands.value = true
                        const thirdToken = normalizeCommandPart(parts[i + 1]).toLowerCase()
                        if (thirdToken) {
                          const thirdFound = findCommandByToken(thirdDynamicList, thirdToken)
                          const thirdIsConfirmed = hasTrailingSpace || (parts.length > i + 2)
                          if (thirdFound && thirdIsConfirmed) {
                            commandStack.value.push(thirdFound)
                            i += 1
                            currentChildren.value = []
                            showCommands.value = false
                          }
                        }
                      } else {
                        currentChildren.value = []
                        showCommands.value = false
                      }
                    }
                  }
                } else {
                  // 兼容 link 新流程：选择环境后继续解析账号（env.dynamicChildren = linkAccountList）
                  if (targetFound.dynamicChildren) {
                    loadDynamicChildren(targetFound.dynamicChildren)
                    const childDynamicList = dynamicDataCache.value[targetFound.dynamicChildren] || []
                    currentChildren.value = childDynamicList
                    showCommands.value = true
                    const childToken = normalizeCommandPart(parts[i + 1]).toLowerCase()
                    if (childToken) {
                      const childFound = findCommandByToken(childDynamicList, childToken)
                      // 终端目标（如账号）在“最后一个 token 精确命中”时也视为确认，无需再手动输入空格。
                      const childIsTerminal = !!(childFound && !childFound.dynamicChildren && !childFound.nextDynamicChildren)
                      const childIsConfirmed = hasTrailingSpace || (parts.length > i + 2) || (parts.length === i + 2 && childIsTerminal)
                      if (childFound && childIsConfirmed) {
                        commandStack.value.push(childFound)
                        i += 1
                        currentChildren.value = []
                        showCommands.value = false
                      }
                    }
                  } else {
                    currentChildren.value = []
                  }
                }
                if (found.needInput) {
                  currentInputValue.value = parts.slice(i + 1).join(' ')
                }
              } else if (targetFound && !targetIsConfirmed) {
                // 未确认选择时，保持在目标候选列表，允许继续匹配（如 chatwiki_dev -> chatwiki_dev12）
                currentChildren.value = dynamicList
              } else if (found.needInput && parts.length > i + 1) {
                currentInputValue.value = parts.slice(i + 1).join(' ')
              }
            }
            break
          }
          // 如果需要选择目标
          if (found.needTarget) {
            break
          }
          // 如果需要输入
          if (found.needInput) {
            currentInputValue.value = parts.slice(i + 1).join(' ')
            break
          }
          currentChildren.value = []
          break
        } else {
          // 没找到，可能是目标选择或输入
          if (commandStack.value.length > 0) {
            const lastCmd = commandStack.value[commandStack.value.length - 1]
            if (lastCmd.needTarget) {
              // 在动态数据中查找
              const dynamicKey = lastCmd.dynamicChildren
              if (dynamicKey && dynamicDataCache.value[dynamicKey]) {
                currentChildren.value = dynamicDataCache.value[dynamicKey]
              }
            }
            if (lastCmd.needInput) {
              currentInputValue.value = parts.slice(i).join(' ')
            }
          }
          break
        }
      }

      if (parts.length === 0) {
        currentChildren.value = availableCommands.value
      }
      if (commandStack.value.length === 0 && parts.length > 0) {
        currentChildren.value = availableCommands.value
      }
      // 动态列表加载中时也展示下拉，避免慢查询时列表框闪退。
      refreshCommandDropdownVisibility()
    }

    // 加载动态子命令
    const loadDynamicChildren = (type) => {
      activeDynamicType.value = String(type || '')
      if (
        type !== 'gitProjectList' &&
        type !== 'gitGroupList' &&
        type !== 'gitRemoteBranchList' &&
        type !== 'supervisorProcessList' &&
        type !== 'linkEnvList' &&
        type !== 'linkAccountList' &&
        type !== 'scriptOptionList' &&
        type !== 'historyList' &&
        dynamicDataCache.value[type]
      ) {
        isLoadingDynamic.value = false
        currentChildren.value = dynamicDataCache.value[type]
        refreshCommandDropdownVisibility()
        return
      }
      
      // 开始异步加载前先清空旧候选，避免展示上一次命令的残留选项。
      currentChildren.value = []
      activeCommandIndex.value = 0
      isLoadingDynamic.value = true
      refreshCommandDropdownVisibility()
      
      switch (type) {
        case 'dockerComposeList':
          loadDockerComposeList()
          break
        case 'gitProjectList':
          loadGitProjectList()
          break
        case 'gitGroupList':
          loadGitGroupList()
          break
        case 'gitRemoteBranchList':
          loadGitRemoteBranchList()
          break
        case 'gitQuickBranchTypeList':
          loadGitQuickBranchTypeList()
          break
        case 'supervisorEnvList':
          loadSupervisorEnvList()
          break
        case 'supervisorProcessList':
          loadSupervisorProcessList()
          break
        case 'shellOutList':
          loadShellOutList()
          break
        case 'redisEnvList':
          loadRedisEnvList()
          break
        case 'dockerServiceList':
          loadDockerServiceList()
          break
        case 'linkConfigList':
          loadLinkConfigList()
          break
        case 'linkEnvList':
          loadLinkEnvList()
          break
        case 'linkAccountList':
          loadLinkAccountList()
          break
        case 'scriptList':
          loadScriptList()
          break
        case 'scriptOptionList':
          loadScriptOptionList()
          break
        case 'historyList':
          loadHistoryList()
          break
        default:
          isLoadingDynamic.value = false
      }
    }

    // 加载 Docker Compose 列表
    const loadDockerComposeList = () => {
      const normalizeDockerDefaultServices = (item) => {
        if (Array.isArray(item?.default_service_list) && item.default_service_list.length > 0) {
          return item.default_service_list
            .map(s => normalizeCommandPart(s))
            .filter(Boolean)
        }
        const raw = normalizeCommandPart(item?.default_service)
        if (!raw) return []
        return raw
          .split(',')
          .map(s => normalizeCommandPart(s))
          .filter(Boolean)
      }

      compose.DockerComposeList({}, (response) => {
        const composeItemList = Array.isArray(response?.Data?.list)
          ? response.Data.list
          : (Array.isArray(response?.Data) ? response.Data : [])

        if (response?.ErrCode !== 0 || composeItemList.length === 0) {
          isLoadingDynamic.value = false
          dynamicDataCache.value['dockerComposeList'] = []
          currentChildren.value = []
          refreshCommandDropdownVisibility()
          return
        }

        const duplicateNameCountMap = {}
        composeItemList.forEach(item => {
          const nameKey = normalizeCommandPart(item?.name).toLowerCase()
          if (!nameKey) {
            return
          }
          duplicateNameCountMap[nameKey] = (duplicateNameCountMap[nameKey] || 0) + 1
        })
        const list = composeItemList
          .map(item => {
            const sshId = normalizeCommandPart(item?.ssh_id)
            const sshName = normalizeCommandPart(item?.ssh_name) || `SSH ${sshId || '-'}`
            const normalizedName = normalizeCommandPart(item?.name)
            const hasDuplicateName = !!normalizedName && duplicateNameCountMap[normalizedName.toLowerCase()] > 1
            const displayName = hasDuplicateName ? `${normalizedName} (${sshName})` : normalizedName
            return {
              command: displayName,
              name: displayName,
              insertText: displayName,
              aliases: [
                normalizedName,
                `${normalizedName}@${sshName}`,
                `${normalizedName}(${sshName})`,
                String(item.id || '')
              ].filter(Boolean),
              desc: [sshName, item.compose_yml_path || ''].filter(Boolean).join(' | '),
              id: sshId ? `${sshId}_${item.id}` : String(item.id || normalizedName),
              data: item,
              default_service_list: normalizeDockerDefaultServices(item)
            }
          })
          .sort((a, b) => {
            const aName = normalizeCommandPart(a.name).toLowerCase()
            const bName = normalizeCommandPart(b.name).toLowerCase()
            if (aName !== bName) {
              return aName.localeCompare(bName)
            }
            return normalizeCommandPart(a.desc).toLowerCase().localeCompare(normalizeCommandPart(b.desc).toLowerCase())
          })
        dynamicDataCache.value['dockerComposeList'] = list
        currentChildren.value = list
        isLoadingDynamic.value = false
        parseInput()
        tryAutoExecutePendingHistory()
        refreshCommandDropdownVisibility()
      })
    }
    // 加载 Docker 服务列表（用于快速重启/停止）
    const loadDockerServiceList = () => {
      // 从命令栈中找到已选择的项目
      const projectCmd = commandStack.value.find(cmd =>
        Array.isArray(cmd?.default_service_list) || Array.isArray(cmd?.data?.default_service_list)
      )
      const services = Array.isArray(projectCmd?.default_service_list)
        ? projectCmd.default_service_list
        : (Array.isArray(projectCmd?.data?.default_service_list) ? projectCmd.data.default_service_list : [])

      if (projectCmd && services.length > 0) {
        const list = services.map(service => ({
          command: service,
          name: service,
          desc: '服务',
          data: { service, projectId: projectCmd.id }
        }))
        dynamicDataCache.value['dockerServiceList'] = list
        currentChildren.value = list
        isLoadingDynamic.value = false
        reparseForPendingHistoryExecution('dockerServiceList')
        refreshCommandDropdownVisibility()
      } else {
        // 如果没有找到项目信息，尝试从缓存的 dockerComposeList 中查找
        const cachedList = dynamicDataCache.value['dockerComposeList']
        if (cachedList && cachedList.length > 0) {
          // 找到命令栈中选择的项目名称
          const projectName = commandStack.value.find(cmd => 
            cachedList.some(item => item.name === cmd.name || item.command === cmd.command)
          )?.name || cachedList[0].name
          
          const project = cachedList.find(item => item.name === projectName)
          if (project && project.default_service_list) {
            const list = project.default_service_list.map(service => ({
              command: service,
              name: service,
              desc: '服务',
              data: { service, projectId: project.id }
            }))
            dynamicDataCache.value['dockerServiceList'] = list
            currentChildren.value = list
          }
        }
        isLoadingDynamic.value = false
        reparseForPendingHistoryExecution('dockerServiceList')
        refreshCommandDropdownVisibility()
      }
    }

    // 加载 Git 项目列表
    const loadGitProjectList = () => {
      git.GitConfigList({}, (response) => {
        isLoadingDynamic.value = false
        if (response.ErrCode === 0) {
          const groupMap = {}
          if (Array.isArray(response.Data.git_group_list)) {
            response.Data.git_group_list.forEach(group => {
              groupMap[group.id] = group.name
            })
          }
          const seen = new Set()
          const list = []
          const gitList = Array.isArray(response.Data.git_list) ? response.Data.git_list : []
          gitList.forEach(item => {
            const itemId = normalizeCommandPart(item.id)
            const dedupeKey = itemId || [
              normalizeCommandPart(item.name),
              normalizeCommandPart(item.path || item.code_path),
              normalizeCommandPart(item.ssh_id)
            ].join('::')
            if (seen.has(dedupeKey)) {
              return
            }
            seen.add(dedupeKey)
            list.push({
              command: item.name,
              name: item.name,
              aliases: [item.path || '', item.code_path || ''].filter(Boolean),
              desc: `${groupMap[item.git_group_id] || '未分组'} ${item.path || item.code_path || ''}`.trim(),
              id: item.id,
              data: item
            })
          })
          dynamicDataCache.value['gitProjectList'] = list
          // 仅当当前仍在该动态列表上下文时，才刷新候选，避免覆盖到“下一步”列表。
          if (activeDynamicType.value === 'gitProjectList') {
            currentChildren.value = list
            reparseForPendingHistoryExecution('gitProjectList')
            refreshCommandDropdownVisibility()
          }
        }
      })
    }

    // 加载 Git 分组列表
    const loadGitGroupList = () => {
      git.GitConfigList({}, (response) => {
        isLoadingDynamic.value = false
        if (response.ErrCode === 0) {
          const gitGroupList = Array.isArray(response.Data.git_group_list) ? response.Data.git_group_list : []
          const list = gitGroupList.map(item => ({
            command: item.name,
            name: item.name,
            aliases: [String(item.id || '')].filter(Boolean),
            desc: `分组ID: ${item.id}`,
            id: item.id,
            data: item
          }))
          dynamicDataCache.value['gitGroupList'] = list
          currentChildren.value = list
          reparseForPendingHistoryExecution('gitGroupList')
          refreshCommandDropdownVisibility()
        }
      })
    }

    // loadGitRemoteBranchList 加载 quick-create-branch 的“基于分支”候选
    const loadGitRemoteBranchList = () => {
      const actionCmd = [...commandStack.value].reverse().find(item => item?.action === 'gitQuickCreateBranch')
      if (!actionCmd) {
        dynamicDataCache.value['gitRemoteBranchList'] = []
        currentChildren.value = []
        isLoadingDynamic.value = false
        refreshCommandDropdownVisibility()
        return
      }
      const actionIndex = commandStack.value.findIndex(item => item === actionCmd)
      const projectCmd = actionIndex >= 0 ? commandStack.value[actionIndex + 1] : null
      if (!(projectCmd && projectCmd.data)) {
        dynamicDataCache.value['gitRemoteBranchList'] = []
        currentChildren.value = []
        isLoadingDynamic.value = false
        refreshCommandDropdownVisibility()
        return
      }
      git.GitRemoteBranchList({ ...projectCmd.data }, (response) => {
        isLoadingDynamic.value = false
        if (!(response && response.ErrCode === 0)) {
          dynamicDataCache.value['gitRemoteBranchList'] = []
          if (activeDynamicType.value === 'gitRemoteBranchList') {
            currentChildren.value = []
            refreshCommandDropdownVisibility()
          }
          return
        }
        const branchList = Array.isArray(response.Data?.list) ? response.Data.list : []
        const list = branchList.map(branchName => ({
          command: branchName,
          name: branchName,
          desc: '远程分支',
          data: { base_branch: branchName },
          // 选择基于分支后，继续选择分支类型
          nextDynamicChildren: 'gitQuickBranchTypeList'
        }))
        dynamicDataCache.value['gitRemoteBranchList'] = list
        // 仅当当前仍在该动态列表上下文时，才刷新候选，避免异步请求把列表切回上一步。
        if (activeDynamicType.value === 'gitRemoteBranchList') {
          currentChildren.value = list
          reparseForPendingHistoryExecution('gitRemoteBranchList')
          refreshCommandDropdownVisibility()
        }
      })
    }

    // loadGitQuickBranchTypeList 加载 quick-create-branch 的分支类型候选
    const loadGitQuickBranchTypeList = () => {
      const list = [
        {
          command: 'feature',
          name: 'feature',
          desc: '功能分支',
          data: { branch_type: 'feature' }
        },
        {
          command: 'hotfix',
          name: 'hotfix',
          desc: '紧急修复分支',
          data: { branch_type: 'hotfix' }
        }
      ]
      dynamicDataCache.value['gitQuickBranchTypeList'] = list
      currentChildren.value = list
      isLoadingDynamic.value = false
      reparseForPendingHistoryExecution('gitQuickBranchTypeList')
      refreshCommandDropdownVisibility()
    }

    // 加载 Supervisor 环境列表
    const loadSupervisorEnvList = () => {
      supervisor.SupervisorConfigList({}, (response) => {
        isLoadingDynamic.value = false
        if (response.ErrCode === 0) {
          const list = response.Data.supervisor_list.map(item => ({
            command: item.name,
            name: item.name,
            desc: item.host || '',
            id: item.id,
            data: item
          }))
          dynamicDataCache.value['supervisorEnvList'] = list
          currentChildren.value = list
          reparseForPendingHistoryExecution('supervisorEnvList')
          refreshCommandDropdownVisibility()
        }
      })
    }

    // 加载 Supervisor 进程列表
    const loadSupervisorProcessList = () => {
      const actionCmd = [...commandStack.value].reverse().find(item => {
        const actionName = String(item?.action || '')
        return actionName === 'supervisorRestart' || actionName === 'supervisorStop' || actionName === 'supervisorConfig'
      })
      if (!actionCmd) {
        dynamicDataCache.value['supervisorProcessList'] = []
        currentChildren.value = []
        isLoadingDynamic.value = false
        refreshCommandDropdownVisibility()
        return
      }
      const actionIndex = commandStack.value.findIndex(item => item === actionCmd)
      const envCmd = actionIndex >= 0 ? commandStack.value[actionIndex + 1] : null
      if (!(envCmd && envCmd.data && envCmd.data.ssh_id)) {
        loadSupervisorEnvList()
        return
      }
      // 这里只是加载候选进程列表，不需要把终端回显串到首页命令执行流。
      fetchSupervisorProcessList(envCmd, {
        forceRefresh: false,
        onSuccess: (list) => {
          isLoadingDynamic.value = false
          applySupervisorProcessListResult(list)
        },
        onError: (response) => {
          isLoadingDynamic.value = false
          if (!(response && response.ErrCode === 0)) {
            dynamicDataCache.value['supervisorProcessList'] = []
            currentChildren.value = []
            refreshCommandDropdownVisibility()
          }
        }
      })
    }

    // 加载终端输出列表
    const loadShellOutList = () => {
      shellOut.ShellOuts({}, (response) => {
        isLoadingDynamic.value = false
        if (response.ErrCode === 0) {
          const list = response.Data.map(item => ({
            command: item.name,
            name: item.name,
            desc: item.command || '',
            id: item.id,
            data: item
          }))
          dynamicDataCache.value['shellOutList'] = list
          currentChildren.value = list
        }
      })
    }

    // 加载自定义链接配置列表
    const loadLinkConfigList = () => {
      smartLinkSet.SmartLinkList((response) => {
        isLoadingDynamic.value = false
        if (!(response && response.ErrCode === 0)) {
          dynamicDataCache.value['linkConfigList'] = []
          currentChildren.value = []
          refreshCommandDropdownVisibility()
          return
        }
        const smartLinkList = Array.isArray(response.Data?.smart_link_list) ? response.Data.smart_link_list : []
        const list = smartLinkList.map(item => {
          let linkList = []
          try {
            linkList = Array.isArray(item.links) ? item.links : JSON.parse(item.links || '[]')
          } catch (err) {
            linkList = []
          }
          return {
            command: item.name,
            name: item.name,
            aliases: [String(item.id || '')].filter(Boolean),
            desc: `ID:${item.id || '-'} | 链接数:${linkList.length}`,
            id: item.id,
            data: {
              ...item,
              __linkType: 'config',
              linkList
            }
          }
        })
        dynamicDataCache.value['linkConfigList'] = list
        currentChildren.value = list
        refreshCommandDropdownVisibility()
      })
    }

    // 加载“可执行环境”列表（将链接配置分组与环境合并为一层）
    const loadLinkEnvList = () => {
      smartLinkSet.SmartLinkList((response) => {
        isLoadingDynamic.value = false
        if (!(response && response.ErrCode === 0)) {
          dynamicDataCache.value['linkEnvList'] = []
          currentChildren.value = []
          refreshCommandDropdownVisibility()
          return
        }
        const smartLinkList = Array.isArray(response.Data?.smart_link_list) ? response.Data.smart_link_list : []
        const list = []
        smartLinkList.forEach((configItem) => {
          let linkList = []
          try {
            linkList = Array.isArray(configItem.links) ? configItem.links : JSON.parse(configItem.links || '[]')
          } catch (err) {
            linkList = []
          }
          linkList.forEach((envItem, envIndex) => {
            const envName = normalizeCommandPart(envItem?.label) || `环境${envIndex + 1}`
            const configName = normalizeCommandPart(configItem?.name) || `配置${configItem?.id || ''}`
            list.push({
              command: envName,
              name: envName,
              insertText: `${configName}/${envName}`,
              aliases: [configName, `${configName}/${envName}`].filter(Boolean),
              desc: `${configName} | ${normalizeCommandPart(envItem?.link) || '未配置链接地址'}`,
              id: `${configItem?.id || 'cfg'}_${envIndex}`,
              dynamicChildren: hasConfiguredLinkAccounts(envItem) ? 'linkAccountList' : undefined,
              data: {
                __linkType: 'env',
                env: envItem || {},
                config: {
                  ...configItem,
                  linkList
                }
              }
            })
          })
        })
        dynamicDataCache.value['linkEnvList'] = list
        currentChildren.value = list
        reparseForPendingHistoryExecution('linkEnvList')
        refreshCommandDropdownVisibility()
      })
    }

    // 加载已选环境下的账号列表
    const loadLinkAccountList = () => {
      const { envCmd } = getLinkRunSelection(commandStack.value)
      const list = buildLinkAccountOptionsFromEnv(envCmd).map((item, index) => ({
        ...item,
        desc: '账号',
        id: `${envCmd?.id || 'env'}_${index}`,
        data: {
          ...(item.data || {}),
          env: envCmd?.data?.env || {},
          config: envCmd?.data?.config || {}
        }
      }))
      dynamicDataCache.value['linkAccountList'] = list
      currentChildren.value = list
      isLoadingDynamic.value = false
      reparseForPendingHistoryExecution('linkAccountList')
      refreshCommandDropdownVisibility()
    }

    // 加载 script 脚本列表
    const loadScriptList = () => {
      variableSet.VariableList((response) => {
        isLoadingDynamic.value = false
        if (!(response && response.ErrCode === 0)) {
          dynamicDataCache.value['scriptList'] = []
          currentChildren.value = []
          refreshCommandDropdownVisibility()
          return
        }
        const scriptList = Array.isArray(response.Data?.variable_list) ? response.Data.variable_list : []
        const list = scriptList.map(item => ({
          command: normalizeCommandPart(item?.name) || `脚本${item?.id || ''}`,
          name: normalizeCommandPart(item?.name) || `脚本${item?.id || ''}`,
          aliases: [String(item?.id || '')].filter(Boolean),
          desc: normalizeCommandPart(item?.desc) || '自定义脚本',
          id: item?.id,
          data: item
        }))
        dynamicDataCache.value['scriptList'] = list
        currentChildren.value = list
        reparseForPendingHistoryExecution('scriptList')
        refreshCommandDropdownVisibility()
      })
    }

    // 加载 script 当前步骤可选项
    const loadScriptOptionList = () => {
      const currentForm = scriptSession.value.currentForm
      const cmdType = normalizeCommandPart(currentForm?.CmdType)
      const optionList = Array.isArray(scriptSession.value.optionList) && scriptSession.value.optionList.length > 0
        ? scriptSession.value.optionList
        : (Array.isArray(currentForm?.Select?.OptionList) ? currentForm.Select.OptionList : [])
      if (!['9', '12', '14'].includes(cmdType) || optionList.length === 0) {
        dynamicDataCache.value['scriptOptionList'] = []
        currentChildren.value = []
        isLoadingDynamic.value = false
        refreshCommandDropdownVisibility()
        return
      }
      const list = optionList.map((item, index) => {
        const label = normalizeCommandPart(item?.Label) || `选项${index + 1}`
        const optionValue = normalizeCommandPart(item?.Value)
        return {
          command: label,
          name: label,
          aliases: [optionValue].filter(Boolean),
          desc: optionValue ? `值: ${optionValue}` : '选项',
          id: `${scriptSession.value.runCmdId || 'cmd'}_${index}`,
          data: {
            optionValue,
            optionLabel: label
          }
        }
      })
      dynamicDataCache.value['scriptOptionList'] = list
      currentChildren.value = list
      isLoadingDynamic.value = false
      reparseForPendingHistoryExecution('scriptOptionList')
      refreshCommandDropdownVisibility()
    }

    // 加载 Redis 环境列表
    const loadRedisEnvList = () => {
      // 简化处理，后续可以扩展
      dynamicDataCache.value['redisEnvList'] = []
      currentChildren.value = []
      isLoadingDynamic.value = false
    }

    // 加载历史命令列表（按使用次数升序，使用最多的在最下面）
    const loadHistoryList = () => {
      const usageMap = commandUsageMap.value || {}
      const list = Object.keys(usageMap)
        .map((commandText, index) => {
          const count = Number(usageMap[commandText]) || 0
          return {
            command: commandText,
            name: commandText,
            desc: `使用 ${count} 次`,
            id: `history_${index}`,
            insertOnly: true,
            insertText: commandText,
            count
          }
        })
        .sort((a, b) => {
          if (a.count !== b.count) {
            return a.count - b.count
          }
          return String(a.command).localeCompare(String(b.command), 'zh-Hans-CN')
        })
      dynamicDataCache.value['historyList'] = list
      currentChildren.value = list
      isLoadingDynamic.value = false
      refreshCommandDropdownVisibility()
    }

    // 处理输入
    const handleInput = () => {
      // 输入变化后重置历史游标，保证再次按上键从最新历史开始
      commandHistoryIndex.value = commandHistory.value.length
      // 手工输入与待执行历史命令不一致时，取消自动执行态，避免误触发。
      if (pendingHistoryExecution.value.active) {
        const currentText = normalizeHistoryCommandText(inputText.value)
        if (currentText !== pendingHistoryExecution.value.commandText) {
          clearPendingHistoryExecution()
        }
      }
      if (isScriptSessionMode()) {
        refreshCommandDropdownVisibility()
        return
      }
      if (isCommandModeByText(inputText.value)) {
        parseInput()
      } else {
        showCommands.value = false
        commandStack.value = []
        currentChildren.value = []
        currentInputValue.value = ''
      }
    }

    // 处理焦点
    const handleFocus = () => {
      if (suppressDropdownOnNextFocus.value) {
        suppressDropdownOnNextFocus.value = false
        return
      }
      if (isScriptSessionMode()) {
        refreshCommandDropdownVisibility()
        return
      }
      if (isCommandModeByText(inputText.value)) {
        parseInput()
      }
    }

    // 记录历史命令（去重连续重复项，最多保留 100 条）
    const pushCommandHistory = (rawCommand) => {
      const command = normalizeCommandPart(rawCommand)
      if (!command) return
      const last = commandHistory.value.length > 0
        ? commandHistory.value[commandHistory.value.length - 1]
        : ''
      if (last !== command) {
        commandHistory.value.push(command)
      }
      commandUsageMap.value[command] = (Number(commandUsageMap.value[command]) || 0) + 1
      if (commandHistory.value.length > 100) {
        commandHistory.value.splice(0, commandHistory.value.length - 100)
      }
      commandHistoryIndex.value = commandHistory.value.length
      persistCommandHistoryCache()
    }

    // persistCommandHistoryCache 持久化首页命令历史与使用次数到本地缓存
    const persistCommandHistoryCache = () => {
      store.setStore(commandHistoryCacheKey, JSON.stringify(commandHistory.value || []))
      store.setStore(commandUsageCacheKey, JSON.stringify(commandUsageMap.value || {}))
    }

    // loadCommandHistoryCache 从本地缓存恢复首页命令历史与使用次数
    const loadCommandHistoryCache = () => {
      let historyList = []
      let usageMap = {}
      let levelUsageMap = {}
      try {
        const historyRaw = store.getStore(commandHistoryCacheKey)
        const usageRaw = store.getStore(commandUsageCacheKey)
        const levelUsageRaw = store.getStore(commandLevelUsageCacheKey)
        const parsedHistory = historyRaw ? JSON.parse(historyRaw) : []
        const parsedUsage = usageRaw ? JSON.parse(usageRaw) : {}
        const parsedLevelUsage = levelUsageRaw ? JSON.parse(levelUsageRaw) : {}
        if (Array.isArray(parsedHistory)) {
          historyList = parsedHistory
            .map(item => normalizeCommandPart(item))
            .filter(Boolean)
            .slice(-100)
        }
        if (parsedUsage && typeof parsedUsage === 'object' && !Array.isArray(parsedUsage)) {
          Object.keys(parsedUsage).forEach((key) => {
            const normalizedKey = normalizeCommandPart(key)
            const count = Number(parsedUsage[key]) || 0
            if (normalizedKey && count > 0) {
              usageMap[normalizedKey] = count
            }
          })
        }
        if (parsedLevelUsage && typeof parsedLevelUsage === 'object' && !Array.isArray(parsedLevelUsage)) {
          Object.keys(parsedLevelUsage).forEach((key) => {
            const normalizedKey = normalizeCommandPart(key).toLowerCase()
            const count = Number(parsedLevelUsage[key]) || 0
            if (normalizedKey && count > 0) {
              levelUsageMap[normalizedKey] = count
            }
          })
        }
      } catch (e) {
        historyList = []
        usageMap = {}
        levelUsageMap = {}
      }
      commandHistory.value = historyList
      commandUsageMap.value = usageMap
      commandLevelUsageMap.value = levelUsageMap
      commandHistoryIndex.value = historyList.length
    }

    // 在输入框为空时，使用上下方向键切换历史命令
    const browseCommandHistory = (direction) => {
      if (commandHistory.value.length === 0) return
      const maxIndex = commandHistory.value.length
      // 到达边界继续按方向键时，清空输入框并退出历史浏览状态
      if (direction < 0 && commandHistoryIndex.value <= 0) {
        commandHistoryIndex.value = maxIndex
        clearPendingHistoryExecution()
        inputText.value = ''
        showCommands.value = false
        commandStack.value = []
        currentChildren.value = []
        currentInputValue.value = ''
        return
      }
      if (direction > 0 && commandHistoryIndex.value >= maxIndex) {
        clearPendingHistoryExecution()
        inputText.value = ''
        showCommands.value = false
        commandStack.value = []
        currentChildren.value = []
        currentInputValue.value = ''
        return
      }
      const nextIndex = Math.max(0, Math.min(maxIndex, commandHistoryIndex.value + direction))
      commandHistoryIndex.value = nextIndex
      if (nextIndex >= maxIndex) {
        clearPendingHistoryExecution()
        inputText.value = ''
        showCommands.value = false
        commandStack.value = []
        currentChildren.value = []
        currentInputValue.value = ''
        return
      }
      loadHistoryCommandForExecution(commandHistory.value[nextIndex])
    }

    // 固定一级命令面板点击：快速填充并进入下一层候选
    const quickSelectTopCommand = (cmd) => {
      const commandText = normalizeCommandPart(cmd?.command)
      if (!commandText) return
      if (commandText === 'script') {
        inputText.value = '/script '
        parseInput()
        showCommands.value = true
        activeCommandIndex.value = 0
        nextTick(() => {
          inputRef.value?.focus()
        })
        return
      }
      inputText.value = `/${commandText} `
      parseInput()
      showCommands.value = true
      activeCommandIndex.value = getDefaultActiveCommandIndex()
      nextTick(() => {
        inputRef.value?.focus()
      })
    }

    // quickSelectHistoryCommand 点击历史命令后回填到输入框
    const quickSelectHistoryCommand = (historyCommand) => {
      startHistoryCommandExecution(historyCommand)
    }

    // 处理失焦
    const handleBlur = () => {
      setTimeout(() => {
        // 点击历史候选删除按钮触发的 blur：保留下拉并回焦输入框
        if (keepDropdownOnBlur.value) {
          keepDropdownOnBlur.value = false
          refreshCommandDropdownVisibility()
          nextTick(() => {
            inputRef.value?.focus()
          })
          return
        }
        showCommands.value = false
      }, 200)
    }

    // markKeepDropdownOnBlur 标记本次 blur 不关闭下拉（用于历史候选删除按钮）
    const markKeepDropdownOnBlur = () => {
      keepDropdownOnBlur.value = true
    }

    // 处理键盘事件
    const handleKeydown = (e) => {
      if (e.key === 'Enter' && !canExecuteCommand.value && !showCommands.value) {
        if (canQueueCurrentCommand()) {
          e.preventDefault()
          executeCommand()
          return
        }
        e.preventDefault()
        return
      }

      // 历史浏览：输入框为空可进入；进入后可继续用上下键切换，按下到末尾会清空输入框
      if (e.key === 'ArrowUp' || e.key === 'ArrowDown') {
        const normalizedInput = normalizeCommandPart(inputText.value)
        const canBrowseFromEmpty = normalizedInput === ''
        const isBrowsingHistory = commandHistoryIndex.value < commandHistory.value.length
        if (canBrowseFromEmpty || isBrowsingHistory) {
          e.preventDefault()
          browseCommandHistory(e.key === 'ArrowUp' ? -1 : 1)
          return
        }
      }

      if (!showCommands.value) {
        if (e.key === 'Enter') {
          executeCommand()
        }
        return
      }

      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault()
          activeCommandIndex.value = Math.min(
            activeCommandIndex.value + 1,
            filteredCommands.value.length - 1
          )
          break
        case 'ArrowUp':
          e.preventDefault()
          activeCommandIndex.value = Math.max(activeCommandIndex.value - 1, 0)
          break
        case 'Tab':
          e.preventDefault()
          if (filteredCommands.value[activeCommandIndex.value]) {
            selectCommand(filteredCommands.value[activeCommandIndex.value])
          }
          break
        case ' ':
          // history 前缀模式下，空格仅作为筛选关键字的一部分，不自动选择候选。
          if (isHistoryPrefixSearchMode()) {
            return
          }
          // 已可执行时，空格不应触发候选选择，避免把候选误拼到命令后
          if (isCommandReadyToExecute()) {
            showCommands.value = false
            return
          }
          if (filteredCommands.value[activeCommandIndex.value]) {
            e.preventDefault()
            selectCommand(filteredCommands.value[activeCommandIndex.value])
          }
          break
        case 'Enter':
          e.preventDefault()
          if (showCommands.value && filteredCommands.value[activeCommandIndex.value]) {
            if (isScriptSessionMode()) {
              selectCommand(filteredCommands.value[activeCommandIndex.value])
              return
            }
          }
          if (canExecuteCommand.value) {
            executeCommand()
          }
          break
        case 'Escape':
          // 退回上一级
          if (commandStack.value.length > 0) {
            goBackCommand()
          } else {
            showCommands.value = false
          }
          break
        case 'Backspace':
          {
            // 如果输入为空且有命令栈，退回上一级
            const parts = inputText.value.split(' ')
            if (parts[parts.length - 1] === '' && commandStack.value.length > 0) {
              e.preventDefault()
              goBackCommand()
            }
          }
          break
      }
    }

    // 退回上一级命令
    const goBackCommand = () => {
      if (commandStack.value.length === 0) return
      
      commandStack.value.pop()
      currentInputValue.value = ''
      
      // 重新构建输入文本
      const tokenInfo = parseTokens(inputText.value)
      const prefix = tokenInfo.useSlash ? '/' : ''
      const commandText = commandStack.value.map(getCommandInputToken).join(' ')
      if (commandText.length > 0) {
        inputText.value = prefix + commandText + ' '
      } else {
        inputText.value = prefix
      }
      
      // 重新解析
      parseInput()
    }

    // 选择命令
    const selectCommand = (cmd) => {
      if (isScriptSessionMode()) {
        if (!cmd) return
        const selectedToken = normalizeCommandPart(cmd.insertText || cmd.command || cmd.name)
        inputText.value = selectedToken
        if (scriptSession.value.stage === 'selecting_script') {
          executeAction(
            { action: 'scriptRun', name: '运行脚本' },
            {
              rawCommand: `script ${selectedToken}`,
              stackOverride: [
                { action: 'scriptRun' },
                { ...cmd, data: cmd.data || {} }
              ]
            }
          )
          return
        }
        if (scriptSession.value.stage === 'waiting_option') {
          executeScriptSessionAction(normalizeCommandPart(cmd?.data?.optionValue) || normalizeCommandPart(cmd.command || cmd.name))
          return
        }
      }
      // history 选择项：仅回填输入框，不改变命令栈
      if (cmd && cmd.insertOnly) {
        const historyText = normalizeCommandPart(cmd.insertText || cmd.command || cmd.name)
        startHistoryCommandExecution(historyText)
        return
      }

      // 构建新的输入文本
      const parts = inputText.value.split(' ')
      parts[parts.length - 1] = cmd.command || cmd.name
      
      // 获取父命令（在选择前）
      const parentCmd = commandStack.value.length > 0 
        ? commandStack.value[commandStack.value.length - 1] 
        : null

      // 添加到命令栈（避免重复追加同一个候选，防止出现 "dev8 dev8"）
      const stackLast = commandStack.value.length > 0
        ? commandStack.value[commandStack.value.length - 1]
        : null
      const selectedCmd = {
        ...cmd,
        __selectedInputToken: normalizeCommandPart(cmd.insertText || cmd.command || cmd.name)
      }
      if (!isSameCommandItem(stackLast, cmd)) {
        commandStack.value.push(selectedCmd)
      }
      
      // 更新输入文本
      const tokenInfo = parseTokens(inputText.value)
      const prefix = tokenInfo.useSlash ? '/' : ''
      inputText.value = prefix + commandStack.value.map(getCommandInputToken).join(' ') + ' '
      // 清空上一步残留输入，避免在“选择目标步骤”被误判为已输入业务参数。
      currentInputValue.value = ''
      
      // 检查父命令是否有 nextDynamicChildren（用于快速重启/停止等二级选择）
      if (parentCmd && parentCmd.nextDynamicChildren) {
        // quick-create-branch 选中分支类型（feature/hotfix）后，下一步应进入业务英文输入，不应继续弹分支类型列表。
        if (cmd?.data?.branch_type) {
          showCommands.value = false
          currentChildren.value = []
          activeCommandIndex.value = 0
          return
        }
        // 加载下一级动态数据
        loadDynamicChildren(parentCmd.nextDynamicChildren)
        activeCommandIndex.value = 0
        return
      }
      
      // 检查是否需要继续
      if (cmd.children && cmd.children.length > 0) {
        // 有子命令，显示子命令列表
        currentChildren.value = cmd.children
        activeCommandIndex.value = getDefaultActiveCommandIndex()
        return
      }
      
      if (cmd.dynamicChildren) {
        // 需要加载动态数据
        loadDynamicChildren(cmd.dynamicChildren)
        activeCommandIndex.value = getDefaultActiveCommandIndex()
        return
      }

      // 支持“当前选择项自身声明了 nextDynamicChildren”的场景（如 git quick-create-branch 选择基线分支后切到分支类型）。
      if (cmd.nextDynamicChildren) {
        loadDynamicChildren(cmd.nextDynamicChildren)
        activeCommandIndex.value = getDefaultActiveCommandIndex()
        return
      }
      
      if (cmd.needTarget) {
        // 需要选择目标，保持下拉框打开（等待动态数据加载）
        activeCommandIndex.value = getDefaultActiveCommandIndex()
        return
      }
      
      if (cmd.needInput) {
        // 需要输入，等待用户输入
        showCommands.value = false
        return
      }
      
      if (cmd.action) {
        // 动作命令仅进入待执行状态，实际执行由 Enter / 发送按钮触发
        showCommands.value = false
        currentChildren.value = []
        return
      }

      if (parentCmd && parentCmd.action === 'scriptRun' && cmd.data) {
        executeAction(
          { action: 'scriptRun', name: '运行脚本' },
          {
            rawCommand: `script ${normalizeCommandPart(cmd.insertText || cmd.command || cmd.name)}`,
            stackOverride: [
              { action: 'scriptRun' },
              { ...cmd, data: cmd.data || {} }
            ]
          }
        )
        return
      }
      
      // 选择的是目标（项目/环境等），检查父命令是否有 action
      if (cmd.data && parentCmd && parentCmd.action) {
        if (parentCmd.needInput) {
          showCommands.value = false
          return
        }
        // 目标选择完成后进入待执行状态
        showCommands.value = false
        currentChildren.value = []
        return
      }

      // 兼容命令栈层级异常时的目标执行：回溯最近 action 命令
      if (cmd.data) {
        const nearestAction = [...commandStack.value].reverse().find(item => item.action)
        if (nearestAction) {
          if (nearestAction.needInput) {
            showCommands.value = false
            return
          }
          // 仅完成选择，不自动执行
          showCommands.value = false
          currentChildren.value = []
          return
        }
      }

      // 其余情况默认关闭下拉，等待用户继续输入或执行
      showCommands.value = false
    }

    // 执行命令
    const executeCommand = () => {
      if (isExecuting.value) {
        if (canQueueCurrentCommand()) {
          queuePendingCommand()
        } else {
          messages.value.push({
            type: 'system',
            content: QUEUE_RUNNING_MESSAGE_TEXT
          })
          scrollToBottom()
        }
        return
      }
      if (isScriptSessionMode()) {
        if (scriptSession.value.stage === 'selecting_script' || scriptSession.value.stage === 'waiting_option') {
          const selectedCmd = filteredCommands.value[activeCommandIndex.value]
          if (selectedCmd) {
            selectCommand(selectedCmd)
          }
          return
        }
        if (scriptSession.value.stage === 'waiting_input') {
          executeScriptSessionAction(inputText.value || '')
          return
        }
        executeAction(
          { action: 'scriptSession', name: '执行脚本' },
          {
            inputValue: inputText.value || '',
            rawCommand: normalizeCommandPart(scriptSession.value.scriptName)
              ? `script ${scriptSession.value.scriptName}`
              : 'script'
          }
        )
        return
      }
      if (!canExecuteCommand.value) return

      if (isCommandModeByText(inputText.value)) {
        parseInput()
      }

      // 如果有命令栈，执行最后一个命令
      if (commandStack.value.length > 0) {
        const actionCmd = commandStack.value.find(item => item.action)
        if (actionCmd) {
          if (!isActionReady(actionCmd, commandStack.value, currentInputValue.value)) {
            messages.value.push({
              type: 'system',
              content: `${getActionIncompleteMessage(actionCmd, commandStack.value, currentInputValue.value)}\n`
            })
            scrollToBottom()
            return
          }
          executeAction(actionCmd, {
            inputValue: currentInputValue.value,
            // 记录本次执行时输入框中的完整命令文本，用于结果区展示
            rawCommand: String(inputText.value || '').trim()
          })
          return
        }

        const lastCmd = commandStack.value[commandStack.value.length - 1]
        // 没有可执行的动作
        messages.value.push({
          type: 'system',
          content: `命令 "${lastCmd.name}" 暂不支持快捷操作\n`
        })
        inputText.value = ''
        showCommands.value = false
        commandStack.value = []
        currentChildren.value = []
        currentInputValue.value = ''
        scrollToBottom()
        return
      }

      // 普通消息
      messages.value.push({
        type: 'user',
        content: inputText.value
      })

      setTimeout(() => {
        messages.value.push({
          type: 'system',
          content: `未知命令，请使用 / 或直接输入命令关键字访问快捷操作`
        })
        scrollToBottom()
      }, 300)

      inputText.value = ''
      showCommands.value = false
      commandStack.value = []
      currentChildren.value = []
      currentInputValue.value = ''
      scrollToBottom()
    }

    // 执行动作
    const executeAction = (cmd, options = {}) => {
      clearPendingHistoryExecution()
      if (isExecuting.value) {
        messages.value.push({
          type: 'system',
          content: '正在执行其他命令，请稍候...'
        })
        return
      }
      
      // 创建输出消息
      const displayCommandText = normalizeCommandPart(options.rawCommand) || cmd.name
      pushCommandHistory(displayCommandText)
      const outputMsg = {
        type: 'system',
        commandText: `执行命令: ${displayCommandText}`,
        commandStatus: 'running',
        summaryText: '',
        resultText: '',
        processText: ''
      }
      messages.value.push(outputMsg)
      currentOutputMessage.value = outputMsg
      isExecuting.value = true
      
      // 清理输入状态
      inputText.value = ''
      showCommands.value = false
      const currentStack = Array.isArray(options.stackOverride)
        ? [...options.stackOverride]
        : [...commandStack.value]
      recordCommandLevelUsage(currentStack)
      commandStack.value = []
      currentChildren.value = []
      currentInputValue.value = ''
      scrollToBottom()
      
      // 根据 action 执行具体操作
      switch (cmd.action) {
        case 'dockerServices':
          executeDockerAction('services', currentStack)
          break
        case 'dockerStatus':
          executeDockerAction('status', currentStack)
          break
        case 'dockerUp':
          executeDockerAction('up', currentStack)
          break
        case 'dockerRestart':
          executeDockerAction('restart', currentStack)
          break
        case 'dockerStop':
          executeDockerAction('stop', currentStack)
          break
        case 'dockerConfig':
          executeDockerAction('config', currentStack)
          break
        case 'dockerEnv':
          executeDockerAction('env', currentStack)
          break
        case 'dockerQuickRestart':
          executeDockerAction('quickRestart', currentStack)
          break
        case 'dockerQuickStop':
          executeDockerAction('quickStop', currentStack)
          break
        case 'supervisorStatus':
          executeSupervisorAction('status', currentStack)
          break
        case 'supervisorRestartAll':
          executeSupervisorAction('restartAll', currentStack)
          break
        case 'supervisorRestart':
          executeSupervisorAction('restart', currentStack)
          break
        case 'supervisorStop':
          executeSupervisorAction('stop', currentStack)
          break
        case 'supervisorConfig':
          executeSupervisorAction('config', currentStack)
          break
        case 'supervisorRefreshProcessCache':
          executeSupervisorAction('refreshCache', currentStack)
          break
        case 'gitPull':
          executeGitAction('pull', currentStack, options.inputValue || '')
          break
        case 'gitStatus':
          executeGitAction('status', currentStack, options.inputValue || '')
          break
        case 'gitBranch':
          executeGitAction('branch', currentStack, options.inputValue || '')
          break
        case 'gitGroupBranches':
          executeGitGroupBranchAction(currentStack)
          break
        case 'gitLog':
          executeGitAction('log', currentStack, options.inputValue || '')
          break
        case 'gitCheckout':
          executeGitAction('checkout', currentStack, options.inputValue || '')
          break
        case 'gitCheckoutRemote':
          executeGitAction('checkoutRemote', currentStack, options.inputValue || '')
          break
        case 'gitQuickCreateBranch':
          executeGitAction('quickCreateBranch', currentStack, options.inputValue || '')
          break
        case 'gitSaveCredentials':
          executeGitAction('saveCredentials', currentStack, options.inputValue || '')
          break
        case 'gitSetSafe':
          executeGitAction('setSafe', currentStack, options.inputValue || '')
          break
        case 'shell':
          executeShellAction('run', currentStack, options.inputValue || '')
          break
        case 'linkRun':
          executeLinkAction(currentStack)
          break
        case 'scriptRun':
          executeScriptRunAction(currentStack)
          break
        case 'scriptSession':
          executeScriptSessionAction(currentInputValue.value || options.inputValue || '')
          break
        case 'gitViewConfig':
          appendOutputResult('已禁用页面跳转，请仅使用命令快捷操作。\n')
          finishExecution()
          break
        case 'gitHelp':
          appendOutputResult('已禁用页面跳转，请仅使用命令快捷操作。\n')
          finishExecution()
          break
        default:
          // 未实现的操作
          appendOutputResult('该操作暂未实现\n')
          finishExecution()
      }
    }

    const getDockerSshId = (composeData, callback) => {
      const composeSshId = normalizeCommandPart(composeData?.ssh_id)
      if (composeSshId) {
        callback(composeSshId)
        return
      }
      const cachedSshId = normalizeCommandPart(store.getStore('dockerChooseSshId'))
      if (cachedSshId) {
        callback(cachedSshId)
        return
      }
      ssh.SshList((response) => {
        if (response.ErrCode === 0 && Array.isArray(response.Data) && response.Data.length > 0) {
          const firstSshId = normalizeCommandPart(response.Data[0].id)
          if (firstSshId) {
            store.setStore('dockerChooseSshId', firstSshId)
            callback(firstSshId)
            return
          }
        }
        callback('')
      })
    }


    const toMarkdownTable = (headers, rows) => {
      const safeCell = (v) => String(v === undefined || v === null ? '' : v).replace(/\|/g, '\\|').replace(/\n/g, ' ')
      const headerLine = `| ${headers.join(' | ')} |`
      const splitLine = `| ${headers.map(() => '---').join(' | ')} |`
      const bodyLines = rows.map(row => `| ${row.map(safeCell).join(' | ')} |`)
      return [headerLine, splitLine, ...bodyLines].join('\n')
    }

    const executeDockerAction = (action, stack) => {
      const actionCmd = stack.find(item => item.action && String(item.action).startsWith('docker'))
      const actionIndex = stack.findIndex(item => item.action && String(item.action).startsWith('docker'))
      const composeCmd = actionIndex >= 0 ? stack[actionIndex + 1] : null
      const serviceCmd = actionIndex >= 0 ? stack[actionIndex + 2] : null

      if (!composeCmd || !composeCmd.data) {
        appendOutputResult('错误：未找到 Docker 项目配置\n')
        finishExecution()
        return
      }

      getDockerSshId(composeCmd.data || {}, (sshId) => {
        if (!sshId) {
          appendOutputResult('错误：未找到可用 SSH 环境，请先在 /Docker 页面选择环境\n')
          finishExecution()
          return
        }

        const newSseDistributeId = sseDistribute.GetSseDistributeId('dashboard_docker')
        const throttleStringFunc = new Throttle_string(50, (text) => {
          if (currentOutputMessage.value) {
            appendOutputProcess(text)
          }
        })
        sseDistribute.RegisterReceive(newSseDistributeId, (msg) => {
          throttleStringFunc.update(msg)
        })

        const composeData = composeCmd.data || {}
        const basePayload = {
          ssh_id: sshId,
          id: composeData.id,
          sse_distribute_id: newSseDistributeId
        }

        const done = (response, renderer) => {
          if (response.ErrCode !== 0) {
            appendOutputResult(`错误: ${response.ErrMsg || '未知错误'}\n`)
          } else {
            renderer(response)
          }
          setTimeout(() => {
            sseDistribute.UnRegisterReceive(newSseDistributeId)
            finishExecution()
          }, 1200)
        }

        switch (action) {
          case 'up':
            appendOutputResult(`正在启动项目 ${composeCmd.name}...\n\n`)
            compose.DockerComposeStart(basePayload, (response) => done(response, () => appendOutputResult('启动完成\n')))
            break
          case 'restart':
            appendOutputResult(`正在重启项目 ${composeCmd.name}...\n\n`)
            compose.DockerComposeRestart(basePayload, (response) => done(response, () => appendOutputResult('重启完成\n')))
            break
          case 'stop':
            appendOutputResult(`正在停止项目 ${composeCmd.name}...\n\n`)
            compose.DockerComposeStop(basePayload, (response) => done(response, () => appendOutputResult('停止完成\n')))
            break
          case 'config':
            appendOutputResult(`正在查看 ${composeCmd.name} 的 compose 配置...\n\n`)
            compose.DockerComposeConfigShow({
              ssh_id: sshId,
              config_path: composeData.compose_yml_path,
              sse_distribute_id: newSseDistributeId
            }, (response) => done(response, (res) => appendOutputResult(`${res.Data || ''}\n`)))
            break
          case 'env':
            {
              const envFile = normalizeCommandPart(composeData.env_file) || String(composeData.compose_yml_path || '').replace(/\/[^/]+\.yml$/, '/.env')
              if (!envFile) {
                appendOutputResult('错误：未找到 .env 路径\n')
                sseDistribute.UnRegisterReceive(newSseDistributeId)
                finishExecution()
                return
              }
              appendOutputResult(`正在查看 ${composeCmd.name} 的 env 配置...\n\n`)
              compose.DockerComposeConfigShow({
                ssh_id: sshId,
                config_path: envFile,
                sse_distribute_id: newSseDistributeId
              }, (response) => done(response, (res) => appendOutputResult(`${res.Data || ''}\n`)))
            }
            break
          case 'services':
            appendOutputResult(`正在查询 ${composeCmd.name} 的服务列表...\n\n`)
            compose.DockerComposeServices(basePayload, (response) => done(response, (res) => {
              const services = (res.Data && Array.isArray(res.Data.services)) ? res.Data.services : []
              if (services.length === 0) {
                appendOutputResult('暂无服务\n')
                return
              }
              const table = toMarkdownTable(['服务名'], services.map(s => [s]))
              appendOutputResult(`${table}\n`)
            }))
            break
          case 'status':
            appendOutputResult(`正在查询 ${composeCmd.name} 的运行状态...\n\n`)
            compose.DockerComposeStatus(basePayload, (response) => done(response, (res) => {
              const statusList = (res.Data && Array.isArray(res.Data.status)) ? res.Data.status : []
              if (statusList.length === 0) {
                appendOutputResult('暂无状态数据\n')
                return
              }
              const table = toMarkdownTable(
                ['名称', 'CPU', '内存', '内存%', '网络IO', '磁盘IO'],
                statusList.map(item => [
                  item.NAME || '',
                  item['CPU %'] || '',
                  item['MEM USAGE / LIMIT'] || '',
                  item['MEM %'] || '',
                  item['NET I/O'] || '',
                  item['BLOCK I/O'] || ''
                ])
              )
              appendOutputResult(`${table}\n`)
            }))
            break
          case 'quickRestart':
            {
              const service = serviceCmd && serviceCmd.data ? serviceCmd.data.service : ''
              if (!service) {
                appendOutputResult('错误：请先选择要重启的服务\n')
                sseDistribute.UnRegisterReceive(newSseDistributeId)
                finishExecution()
                return
              }
              appendOutputResult(`正在快速重启服务 ${service}...\n\n`)
              compose.DockerComposeRestart({ ...basePayload, service }, (response) => done(response, () => appendOutputResult('快速重启完成\n')))
            }
            break
          case 'quickStop':
            {
              const serviceStop = serviceCmd && serviceCmd.data ? serviceCmd.data.service : ''
              if (!serviceStop) {
                appendOutputResult('错误：请先选择要停止的服务\n')
                sseDistribute.UnRegisterReceive(newSseDistributeId)
                finishExecution()
                return
              }
              appendOutputResult(`正在快速停止服务 ${serviceStop}...\n\n`)
              compose.DockerComposeStop({ ...basePayload, service: serviceStop }, (response) => done(response, () => appendOutputResult('快速停止完成\n')))
            }
            break
          default:
            sseDistribute.UnRegisterReceive(newSseDistributeId)
            appendOutputResult('该 Docker 操作暂未实现\n')
            finishExecution()
        }
      })
    }

    // 执行 Supervisor 相关操作
    const executeSupervisorAction = (action, stack) => {
      const actionCmd = stack.find(item => item.action && String(item.action).startsWith('supervisor'))
      const actionIndex = stack.findIndex(item => item.action && String(item.action).startsWith('supervisor'))
      const envCmd = actionIndex >= 0 ? stack[actionIndex + 1] : null
      const processCmd = actionIndex >= 0 ? stack[actionIndex + 2] : null

      if (!(envCmd && envCmd.data)) {
        appendOutputResult('错误：请先选择 Supervisor 环境\n')
        finishExecution()
        return
      }

      const supervisorConfig = {
        ...envCmd.data,
        sse_distribute_id: sseDistributeId.value
      }

      // renderSupervisorResult 仅写入操作完成提示，不展示接口返回详情。
      const renderSupervisorResult = (response, successText) => {
        if (successText) {
          appendOutputResult(`${successText}\n`)
        }
      }

      // done 统一处理 Supervisor 操作完成后的状态与收尾逻辑。
      const done = (response, renderer) => {
        if (!(response && response.ErrCode === 0)) {
          appendOutputResult(`错误: ${normalizeCommandPart(response?.ErrMsg) || '未知错误'}\n`)
          setTimeout(() => {
            finishExecution(response)
          }, 1500)
        } else {
          if (typeof renderer === 'function') {
            renderer(response)
          } else if (renderer) {
            appendOutputResult(`${renderer}\n`)
          }
          // 结果详情统一在“执行过程(SSE)”里查看，上方只保留成功提示
          setTimeout(() => {
            finishExecution(response)
          }, 1500)
        }
      }

      switch (action) {
        case 'status':
          appendOutputResult(`正在查看环境 [${envCmd.name}] 的进程状态...\n\n`)
          supervisor.SupervisorStatusList({ ...supervisorConfig }, (response) => done(response, (res) => renderSupervisorResult(res, '查询完成')))
          break
        case 'restartAll':
          appendOutputResult(`正在重启环境 [${envCmd.name}] 的全部进程...\n\n`)
          supervisor.SupervisorRestartAll({ ...supervisorConfig }, (response) => done(response, (res) => renderSupervisorResult(res, '重启完成')))
          break
        case 'restart':
          if (!(processCmd && processCmd.data && processCmd.data.supervisor_name)) {
            appendOutputResult('错误：请先选择要重启的服务\n')
            finishExecution()
            return
          }
          appendOutputResult(`正在重启服务 [${processCmd.name}]...\n\n`)
          supervisor.SupervisorRestart(
            { ...supervisorConfig },
            processCmd.data.supervisor_name,
            (response) => done(response, (res) => renderSupervisorResult(res, '重启完成')),
            { only_current_status: true }
          )
          break
        case 'stop':
          if (!(processCmd && processCmd.data && processCmd.data.supervisor_name)) {
            appendOutputResult('错误：请先选择要停止的服务\n')
            finishExecution()
            return
          }
          appendOutputResult(`正在停止服务 [${processCmd.name}]...\n\n`)
          supervisor.SupervisorStop(
            { ...supervisorConfig },
            processCmd.data.supervisor_name,
            (response) => done(response, (res) => renderSupervisorResult(res, '停止完成'))
          )
          break
        case 'config':
          if (!(processCmd && processCmd.data && processCmd.data.supervisor_config)) {
            appendOutputResult('错误：请先选择要查看配置的服务\n')
            finishExecution()
            return
          }
          appendOutputResult(`正在查看服务 [${processCmd.name}] 配置...\n\n`)
          supervisor.SupervisorConfigShow(
            { ...supervisorConfig },
            processCmd.data.supervisor_config,
            (response) => done(response, (res) => renderSupervisorResult(res, '查看完成'))
          )
          break
        case 'refreshCache':
          appendOutputResult(`正在刷新环境 [${envCmd.name}] 的服务列表缓存...\n\n`)
          clearSupervisorProcessCache(envCmd.data)
          fetchSupervisorProcessList(envCmd, {
            forceRefresh: true,
            onSuccess: (list) => {
              dynamicDataCache.value['supervisorProcessList'] = list
              done({ ErrCode: 0 }, (res) => renderSupervisorResult(res, `刷新完成，共 ${list.length} 项`))
            },
            onError: (response) => done(response)
          })
          break
        default:
          appendOutputResult('该 Supervisor 操作暂未实现\n')
          finishExecution()
      }
    }

    const executeGitGroupBranchAction = (stack) => {
      const groupCmd = stack.find(c => c.data && c.data.id !== undefined && c.data.id !== null)
      if (!groupCmd || !groupCmd.data) {
        appendOutputResult('错误：未找到 Git 分组配置\n')
        finishExecution()
        return
      }

      const newSseDistributeId = sseDistribute.GetSseDistributeId('dashboard_git_group')
      const throttleStringFunc = new Throttle_string(50, (text) => {
        if (currentOutputMessage.value) {
          appendOutputProcess(text)
        }
      })

      sseDistribute.RegisterReceive(newSseDistributeId, (msg) => {
        throttleStringFunc.update(msg)
      })

      const callback = (response) => {
        if (response.ErrCode !== 0) {
          appendOutputResult('执行失败\n')
        } else {
          appendOutputResult('执行成功\n')
        }
        setTimeout(() => {
          sseDistribute.UnRegisterReceive(newSseDistributeId)
          finishExecution(response)
        }, 1200)
      }

      git.GitGroupBranchList({
        git_group_id: groupCmd.data.id,
        sse_distribute_id: newSseDistributeId
      }, callback)
    }
    
    // 执行 Git 相关操作
    const executeGitAction = (action, stack, inputValue) => {
      // 获取选中的 git 项目配置
      const projectCmd = stack.find(c => c.data && c.data.id)
      if (!projectCmd || !projectCmd.data) {
        appendOutputResult('错误：未找到 Git 项目配置\n')
        finishExecution()
        return
      }
      
      // 每次操作生成新的 SSE 分发 ID，确保使用新的连接
      const newSseDistributeId = sseDistribute.GetSseDistributeId('dashboard_git')
      
      // 注册当前操作的 SSE 回调
      const throttleStringFunc = new Throttle_string(50, (text) => {
        if (currentOutputMessage.value) {
          appendOutputProcess(text)
        }
      })
      
      sseDistribute.RegisterReceive(newSseDistributeId, (msg, msgType, sseDistributeId) => {
        throttleStringFunc.update(msg)
      })
      
      const gitConfig = {
        ...projectCmd.data,
        sse_distribute_id: newSseDistributeId
      }

      const gitActionLabelMap = {
        pull: '拉取代码',
        status: '查询仓库状态',
        branch: '查询当前分支',
        log: '查询提交日志',
        checkout: '切换分支',
        checkoutRemote: '关联远程分支切换',
        quickCreateBranch: '快捷创建分支',
        saveCredentials: '保存账号密码配置',
        setSafe: '设置目录安全'
      }

      const gitActionDoneLabelMap = {
        pull: '拉取完成',
        status: '状态查询完成',
        branch: '分支查询完成',
        log: '日志查询完成',
        checkout: '分支切换完成',
        checkoutRemote: '远程分支切换完成',
        quickCreateBranch: '快捷建分支完成',
        saveCredentials: '保存账号密码配置完成',
        setSafe: '设置目录安全完成'
      }

      appendOutputResult(`正在${gitActionLabelMap[action] || '执行 Git 操作'}...\n\n`)
      
      // 处理 HTTP 响应的回调
      const callback = (response) => {
        if (response.ErrCode !== 0) {
          appendOutputResult(`执行失败: ${normalizeCommandPart(response?.ErrMsg) || '未知错误'}\n`)
        } else {
          appendOutputSummary(gitActionDoneLabelMap[action] || '执行成功')
        }
        setTimeout(() => {
          // 给 SSE 尾包一点时间，避免过程/结果末尾被截断
          sseDistribute.UnRegisterReceive(newSseDistributeId)
          finishExecution(response)
        }, 1200)
      }
      
      switch (action) {
        case 'pull':
          git.GitPullBranchOrigin(gitConfig, callback)
          break
        case 'status':
          git.GitQueryStatus(gitConfig, callback)
          break
        case 'branch':
          git.GitCurrentBranch(gitConfig, callback)
          break
        case 'log':
          git.GitCommitLog(gitConfig, callback)
          break
        case 'checkout':
          {
            // 需要分支名
            const branchName = normalizeCommandPart(inputValue)
            if (!branchName) {
              appendOutputResult('执行失败\n')
              finishExecution()
              return
            }
            git.GitChangeBranch(gitConfig, branchName, callback)
          }
          break
        case 'checkoutRemote':
          {
            const branchNameRemote = normalizeCommandPart(inputValue)
            if (!branchNameRemote) {
              appendOutputResult('执行失败\n')
              finishExecution()
              return
            }
            git.GitChangeBranchRemote(gitConfig, branchNameRemote, callback)
          }
          break
        case 'quickCreateBranch':
          {
            const selection = getGitQuickCreateSelection(stack)
            const baseBranch = normalizeCommandPart(selection.baseBranchCmd?.data?.base_branch || selection.baseBranchCmd?.command)
            const branchType = normalizeCommandPart(selection.branchTypeCmd?.data?.branch_type || selection.branchTypeCmd?.command).toLowerCase()
            const businessEN = normalizeCommandPart(inputValue)
            if (!baseBranch || !branchType || !/^[A-Za-z0-9_]+$/.test(businessEN)) {
              appendOutputResult('执行失败\n')
              finishExecution()
              return
            }
            git.GitQuickCreateBranch({
              ...gitConfig,
              base_branch: baseBranch,
              branch_type: branchType,
              business_en: businessEN
            }, callback)
          }
          break
        case 'saveCredentials':
          git.GitSaveCredentials(gitConfig, callback)
          break
        case 'setSafe':
          git.SetSafe(gitConfig, callback)
          break
        default:
          sseDistribute.UnRegisterReceive(newSseDistributeId)
          finishExecution()
      }
    }

    // 执行终端输出相关动作：复用“新窗口”按钮行为
    const executeShellAction = (action, stack, inputValue) => {
      if (action === 'run') {
        const actionIndex = stack.findIndex(item => item.action === 'shell')
        const targetCmd = actionIndex >= 0 ? stack[actionIndex + 1] : null
        if (!targetCmd || !targetCmd.data) {
          appendOutputSummary('打开失败')
          appendOutputResult('错误：请先选择要打开的终端输出任务\n')
          finishExecution()
          return
        }
        const target = targetCmd.data || {}
        const chooseGroupId = normalizeCommandPart(store.getStore('shell_out_choose_group_id'))
        const groupId = chooseGroupId || normalizeCommandPart(target.group_id) || '0'
        const id = normalizeCommandPart(target.id)
        const title = encodeURIComponent(String(target.name || id || 'shellout'))
        if (!id) {
          appendOutputSummary('打开失败')
          appendOutputResult('错误：任务ID为空，无法打开新窗口\n')
          finishExecution()
          return
        }
        appendOutputResult(`正在打开终端输出任务 [${target.name || target.id}]...\n`)
        const url = `${window.location.origin}/#/fullpage?group_id=${groupId}&id=${id}&title=${title}`
        window.open(url, '_blank')
        appendOutputSummary(`已打开终端输出任务 [${target.name || target.id}]`)
        finishExecution()
        return
      }

      appendOutputSummary('执行失败')
      appendOutputResult('该终端输出操作暂未实现\n')
      finishExecution()
    }

    // 执行 link run：根据“环境(含配置) -> 账号”选择启动自定义链接
    const executeLinkAction = (stack) => {
      const selection = getLinkRunSelection(stack)
      if (!isLinkRunSelectionComplete(selection)) {
        appendOutputResult(`${getActionIncompleteMessage({ action: 'linkRun' }, stack, '')}\n`)
        finishExecution()
        return
      }
      const payload = buildLinkRunPayload(selection, sseDistributeId.value, normalizeCommandPart)

      if (!payload.id || !payload.label) {
        appendOutputResult('错误：链接配置不完整，无法执行\n')
        finishExecution()
        return
      }

      appendOutputResult(`正在执行环境 [${payload.label}] 的自定义网页任务...\n\n`)
      smartLinkSet.SmartLinkRun(payload, (response) => {
        if (response && response.ErrCode === 0) {
          appendOutputResult('执行完成\n')
        } else {
          appendOutputResult(`执行失败: ${normalizeCommandPart(response?.ErrMsg) || '未知错误'}\n`)
        }
        finishExecution(response)
      })
    }

    // 重置 script 会话
    const resetScriptSession = () => {
      scriptSession.value = {
        active: false,
        stage: 'idle',
        scriptId: 0,
        scriptName: '',
        runCmdId: 0,
        replaceList: {},
        currentForm: null,
        pendingInputLabel: '',
        optionList: [],
        canExecute: false,
      }
      dynamicDataCache.value['scriptOptionList'] = []
      currentChildren.value = []
      inputText.value = ''
      refreshCommandDropdownVisibility()
    }

    const enterScriptSelectionStage = () => {
      scriptSession.value.active = true
      scriptSession.value.stage = 'selecting_script'
      scriptSession.value.scriptId = 0
      scriptSession.value.scriptName = ''
      scriptSession.value.runCmdId = 0
      scriptSession.value.replaceList = {}
      scriptSession.value.currentForm = null
      scriptSession.value.pendingInputLabel = ''
      scriptSession.value.optionList = []
      scriptSession.value.canExecute = false
      inputText.value = ''
      loadScriptList()
    }

    // 处理 script API 返回，更新会话与下一步提示
    const handleScriptFlowResponse = (response, options = {}) => {
      const { fallbackToSelectingScript = false } = options
      if (!(response && response.ErrCode === 0)) {
        appendOutputResult(`执行失败: ${normalizeCommandPart(response?.ErrMsg) || '未知错误'}\n`)
        if (fallbackToSelectingScript) {
          enterScriptSelectionStage()
        }
        finishExecution(response)
        return
      }
      const data = response.Data || {}
      const runStatus = Number(data.RunStatus)
      const currentForm = data.Form || null
      const optionList = Array.isArray(currentForm?.Select?.OptionList) ? currentForm.Select.OptionList : []
      if (data.ReplaceList && typeof data.ReplaceList === 'object') {
        scriptSession.value.replaceList = data.ReplaceList
      }
      if (currentForm && currentForm.Id !== undefined && currentForm.Id !== null) {
        scriptSession.value.runCmdId = Number(currentForm.Id) || 0
      }
      scriptSession.value.currentForm = currentForm
      scriptSession.value.active = true

      if (runStatus === 0) {
        const cmdType = normalizeCommandPart(currentForm?.CmdType)
        if (['3', '17'].includes(cmdType)) {
          const label = normalizeCommandPart(currentForm?.Input?.Label) || '请在命令框输入内容并回车'
          scriptSession.value.stage = 'waiting_input'
          scriptSession.value.pendingInputLabel = label
          scriptSession.value.optionList = []
          scriptSession.value.canExecute = false
          currentChildren.value = []
          showCommands.value = false
          inputText.value = ''
          appendOutputResult(`当前步骤: ${label}\n请在命令框输入内容并回车\n`)
        } else if (['9', '12', '14'].includes(cmdType)) {
          scriptSession.value.stage = 'waiting_option'
          scriptSession.value.pendingInputLabel = normalizeCommandPart(currentForm?.Select?.Label) || '请在命令框选择一个选项'
          scriptSession.value.optionList = optionList
          scriptSession.value.canExecute = false
          inputText.value = ''
          loadScriptOptionList()
          appendOutputResult(`当前步骤: ${scriptSession.value.pendingInputLabel}\n请在命令框选择一个选项\n`)
        } else {
          appendOutputResult('脚本返回了未适配的步骤类型，请到脚本页面执行。\n')
          enterScriptSelectionStage()
        }
        finishExecution()
        return
      }
      if (runStatus === 1) {
        scriptSession.value.stage = 'ready_execute'
        scriptSession.value.pendingInputLabel = ''
        scriptSession.value.optionList = []
        scriptSession.value.canExecute = true
        currentChildren.value = []
        showCommands.value = false
        inputText.value = ''
        appendOutputResult('脚本已就绪，按回车执行\n')
        finishExecution()
        return
      }
      if (runStatus === 2) {
        appendOutputResult('脚本执行完成\n')
        resetScriptSession()
        finishExecution()
        return
      }
      appendOutputResult('收到未知状态，已保留当前脚本会话\n')
      finishExecution()
    }

    // 执行 script run：选择脚本并启动
    const executeScriptRunAction = (stack) => {
      const actionIndex = stack.findIndex(item => item.action === 'scriptRun')
      const targetCmd = actionIndex >= 0 ? stack[actionIndex + 1] : null
      if (!(targetCmd && targetCmd.data && targetCmd.data.id)) {
        appendOutputResult('错误：请先选择要执行的脚本\n')
        enterScriptSelectionStage()
        finishExecution()
        return
      }
      resetScriptSession()
      scriptSession.value.active = true
      scriptSession.value.stage = 'executing'
      scriptSession.value.scriptId = Number(targetCmd.data.id) || 0
      scriptSession.value.scriptName = normalizeCommandPart(targetCmd.data.name) || normalizeCommandPart(targetCmd.name)
      appendOutputResult(`已启动脚本: ${scriptSession.value.scriptName || scriptSession.value.scriptId}\n`)
      variableSet.VariableRun(
        sseDistributeId.value,
        scriptSession.value.scriptId,
        0,
        0,
        JSON.stringify({}),
        (response) => {
          handleScriptFlowResponse(response, { fallbackToSelectingScript: true })
        }
      )
    }

    // 执行 script 会话动作：按当前阶段自动解释输入/选项/执行
    const executeScriptSessionAction = (inputValue) => {
      const session = scriptSession.value
      if (!session.active || !session.scriptId) {
        enterScriptSelectionStage()
        return
      }

      const currentForm = session.currentForm || {}
      const cmdType = normalizeCommandPart(currentForm?.CmdType)

      if (session.stage === 'waiting_input') {
        if (!['3', '17'].includes(cmdType)) {
          appendOutputResult('当前步骤不是输入步骤，请在命令框选择一个选项\n')
          finishExecution()
          return
        }
        const editValue = normalizeCommandPart(inputValue)
        if (!editValue) {
          appendOutputResult('请在命令框输入内容并回车\n')
          finishExecution()
          return
        }
        session.stage = 'executing'
        variableSet.VariableSet(
          session.scriptId,
          Number(currentForm.Id) || Number(session.runCmdId) || 0,
          JSON.stringify(session.replaceList || {}),
          editValue,
          (response) => {
            handleScriptFlowResponse(response)
          }
        )
        return
      }

      if (session.stage === 'waiting_option') {
        if (!['9', '12', '14'].includes(cmdType)) {
          appendOutputResult('当前步骤不是选项步骤，请在命令框输入内容并回车\n')
          finishExecution()
          return
        }
        const selectedValue = normalizeCommandPart(inputValue)
        const options = Array.isArray(currentForm?.Select?.OptionList) ? currentForm.Select.OptionList : []
        const matched = options.find(item => {
          const label = normalizeCommandPart(item?.Label)
          const value = normalizeCommandPart(item?.Value)
          return selectedValue && (selectedValue === label || selectedValue === value)
        })
        if (!matched) {
          appendOutputResult('请在命令框选择一个选项\n')
          finishExecution()
          return
        }
        session.stage = 'executing'
        variableSet.VariableSet(
          session.scriptId,
          Number(currentForm.Id) || Number(session.runCmdId) || 0,
          JSON.stringify(session.replaceList || {}),
          normalizeCommandPart(matched?.Value),
          (response) => {
            handleScriptFlowResponse(response)
          }
        )
        return
      }

      if (session.stage === 'ready_execute') {
        session.stage = 'executing'
        session.canExecute = false
        variableSet.VariableRun(
          sseDistributeId.value,
          session.scriptId,
          Number(session.runCmdId) || Number(currentForm.Id) || 0,
          1,
          JSON.stringify(session.replaceList || {}),
          (response) => {
            handleScriptFlowResponse(response)
          }
        )
        return
      }

      if (session.stage === 'selecting_script') {
        enterScriptSelectionStage()
        return
      }

      appendOutputResult('脚本执行中，请稍候\n')
      finishExecution()
    }

    // 生成首页命令执行完成后的浏览器通知内容。
    const buildBrowserNotificationPayload = (message) => {
      const status = normalizeCommandPart(message?.commandStatus) || 'success'
      const commandText = normalizeCommandPart(message?.commandText) || '首页命令'
      const resultText = normalizeCommandPart(message?.resultText)
      const title = status === 'failed' ? '命令执行失败' : '命令执行完成'
      const body = resultText
        ? `${commandText}\n${resultText}`.slice(0, 180)
        : `${commandText}\n请返回页面查看详情`
      return {
        title,
        body,
      }
    }

    // 按需申请浏览器通知权限，避免首页初次加载时直接弹权限框。
    const ensureBrowserNotificationPermission = async () => {
      if (typeof window === 'undefined' || typeof Notification === 'undefined') {
        return 'unsupported'
      }
      if (Notification.permission === 'granted') {
        return 'granted'
      }
      if (Notification.permission === 'denied') {
        return 'denied'
      }
      if (browserNotificationPermissionRequested.value) {
        return Notification.permission
      }
      browserNotificationPermissionRequested.value = true
      try {
        return await Notification.requestPermission()
      } catch (error) {
        return 'default'
      }
    }

    // 首页命令执行结束后发送浏览器通知，便于切出页面时感知结果。
    const notifyCommandFinished = async (message) => {
      if (!message) {
        return
      }
      const permission = await ensureBrowserNotificationPermission()
      if (permission !== 'granted') {
        return
      }
      const payload = buildBrowserNotificationPayload(message)
      const notification = new Notification(payload.title, {
        body: payload.body,
        tag: 'dashboard-command-finished',
      })
      notification.onclick = () => {
        window.focus()
        notification.close()
      }
      setTimeout(() => {
        notification.close()
      }, 5000)
    }
    
    // 完成执行
    const finishExecution = (response = null) => {
      applyResponseCommandStatus(response)
      const finishedMessage = currentOutputMessage.value
      // 若未显式写入成功/失败状态，执行结束后默认标记为成功
      if (finishedMessage && finishedMessage.commandStatus === 'running') {
        finishedMessage.commandStatus = 'success'
      }
      isExecuting.value = false
      currentOutputMessage.value = null
      notifyCommandFinished(finishedMessage)
      scrollToBottom()
      triggerNextPendingCommand()
    }

    // 滚动到底部
    const scrollToBottom = () => {
      nextTick(() => {
        if (messageList.value) {
          requestAnimationFrame(() => {
            messageList.value.scrollTop = messageList.value.scrollHeight
            const processTextList = messageList.value.querySelectorAll('.process-text')
            if (processTextList && processTextList.length > 0) {
              const latestProcessText = processTextList[processTextList.length - 1]
              latestProcessText.scrollTop = latestProcessText.scrollHeight
            }
          })
        }
      })
    }

    // 初始化 SSE 连接
    const initSseConnection = () => {
      sseDistributeId.value = sseDistribute.GetSseDistributeId('dashboard')
      
      // 检查是否已存在 SSE 连接，如果不存在则创建
      const existingClientId = sseDistribute.GetSseClientId()
      if (!existingClientId) {
        // 创建 SSE 连接
        sseDistribute.Create()
        sseDistribute.ReceiveMessage()
        
        sseDistribute.OpenFunc(() => {
          console.log('SSE 连接已建立')
        })
        
        sseDistribute.ErrorFunc((err) => {
          console.log('SSE 连接错误', err)
        })
      }
      
      // 注册消息回调（用于通用的 dashboard 消息）
      const throttleStringFunc = new Throttle_string(50, (text) => {
        if (currentOutputMessage.value) {
          appendOutputProcess(text)
        }
      })
      
      sseDistribute.RegisterReceive(sseDistributeId.value, (msg, msgType, sseDistributeId) => {
        throttleStringFunc.update(msg)
      })
    }

    // focusInputOnHome 首页输入框聚焦（首次进入与切回首页时复用）
    const focusInputOnHome = () => {
      nextTick(() => {
        inputRef.value?.focus()
      })
    }

    // removeHistoryCommand 删除指定历史命令（同步清理历史列表与使用次数缓存）
    const removeHistoryCommand = (historyCommand) => {
      const commandText = normalizeCommandPart(historyCommand)
      if (!commandText) return

      // 删除历史列表中该命令的所有记录，避免旧记录残留
      commandHistory.value = commandHistory.value.filter(item => normalizeCommandPart(item) !== commandText)
      // 删除使用次数统计，确保 history 动态列表也同步移除
      if (Object.prototype.hasOwnProperty.call(commandUsageMap.value, commandText)) {
        delete commandUsageMap.value[commandText]
      }
      // 同步刷新缓存中的 history 候选数据
      if (Array.isArray(dynamicDataCache.value['historyList'])) {
        dynamicDataCache.value['historyList'] = dynamicDataCache.value['historyList']
          .filter(item => normalizeCommandPart(item?.command || item?.name) !== commandText)
      }

      commandHistoryIndex.value = commandHistory.value.length
      persistCommandHistoryCache()

      // 若当前输入就是被删除命令，重置输入状态，避免误执行已删除项
      if (normalizeCommandPart(inputText.value) === commandText) {
        inputText.value = ''
        showCommands.value = false
        commandStack.value = []
        currentChildren.value = []
        currentInputValue.value = ''
      } else if (isHistoryPrefixSearchMode()) {
        // history 前缀模式下删除后立即重算候选，保持界面与输入一致
        parseInput()
      }

      nextTick(() => {
        inputRef.value?.focus()
      })
    }

    // handleHomeAppear 首页出现时统一处理：输入框聚焦 + 输出区滚动到底部
    const handleHomeAppear = () => {
      focusInputOnHome()
      scrollToBottom()
    }

    watch(filteredCommands, (commandList) => {
      if (!showCommands.value) {
        return
      }
      activeCommandIndex.value = getDefaultActiveCommandIndex(commandList)
    })

    watch([showCommands, activeCommandIndex, filteredCommands], () => {
      commandItemRefs.value = []
      nextTick(() => {
        ensureActiveCommandVisible()
      })
    })

    onMounted(() => {
      loadCommandHistoryCache()
      handleHomeAppear()
      initSseConnection()
    })

    // keep-alive 组件重新激活时，自动让首页输入框获得焦点
    onActivated(() => {
      handleHomeAppear()
    })
    
    onUnmounted(() => {
      // 只取消注册回调，不关闭 SSE 连接（其他页面可能还在使用）
      sseDistribute.UnRegisterReceive(sseDistributeId.value)
    })

    return {
      inputText,
      messages,
      isExecuting,
      showCommands,
      isLoadingDynamic,
      filteredCommands,
      activeCommandIndex,
      commandDropdown,
      inputRef,
      messageList,
      commandBreadcrumb,
      inputPlaceholder,
      canExecuteCommand,
      canSubmitCommand,
      highlightedInputHtml,
      inputWrapperWidth,
      nextStepHint,
      pendingCommandQueue,
      pendingCommandTitle,
      hasPendingCommandQueue,
      availableCommands,
      topHistoryCommands,
      handleInput,
      handleKeydown,
      handleFocus,
      handleBlur,
      quickSelectTopCommand,
      quickSelectHistoryCommand,
      removeHistoryCommand,
      removePendingCommand,
      markKeepDropdownOnBlur,
      setCommandItemRef,
      selectCommand,
      executeCommand,
      getCommandLevelUsageCount,
      getCommandKey,
      getCommandMatchHint,
      getResultLineState,
      getResultLines,
      renderProcessMarkdown,
      hasCommandLayout,
    }
  }
}
</script>

<style scoped src="@/css/components/Dashboard.css"></style>
