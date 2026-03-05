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
            <div v-if="recentHistoryCommands.length > 0" class="history-command-section">
              <div class="fixed-command-title">历史操作命令</div>
              <div class="history-command-list">
                <div
                  v-for="(historyCmd, historyIndex) in recentHistoryCommands"
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
            </div>
          </div>
          <p class="hint">输入 <kbd>/</kbd> 或直接输入命令（如 <kbd>g</kbd>），<kbd>Tab</kbd> 补全，<kbd>Space</kbd> 继续</p>
        </div>
        <div
          v-for="(msg, index) in messages"
          :key="index"
          :class="['message', msg.type]"
        >
          <template v-if="hasCommandLayout(msg)">
            <div class="message-command">
              <span class="message-command-text">{{ msg.commandText }}</span>
              <span
                v-if="msg.commandStatus"
                :class="['command-status', `command-status-${msg.commandStatus}`]"
              >
                <span v-if="msg.commandStatus === 'running'" class="command-status-spinner"></span>
                <span v-else-if="msg.commandStatus === 'success'" class="command-status-icon">✓</span>
                <span v-else-if="msg.commandStatus === 'failed'" class="command-status-icon">✕</span>
              </span>
            </div>
            <div v-if="msg.resultText" class="message-content">{{ msg.resultText }}</div>
            <div v-if="msg.processText" class="process-window">
              <div class="process-title">执行过程 (SSE)</div>
              <div class="process-text markdown-body" v-html="renderProcessMarkdown(msg.processText)"></div>
            </div>
          </template>
          <div v-else class="message-content">{{ msg.content }}</div>
        </div>
      </div>

      <!-- 命令提示下拉框 -->
      <div v-show="showCommands" class="command-dropdown">
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
          :class="['command-item', { active: activeCommandIndex === index }]"
          @click="selectCommand(cmd)"
          @mouseenter="activeCommandIndex = index"
        >
          <span class="command-icon">{{ cmd.icon }}</span>
          <span class="command-name">{{ cmd.name }}</span>
          <span class="command-desc">
            {{ cmd.desc }}<template v-if="getCommandMatchHint(cmd)"> | 匹配: {{ getCommandMatchHint(cmd) }}</template>
          </span>
          <button
            v-if="cmd.insertOnly"
            type="button"
            class="command-item-delete"
            title="删除该历史命令"
            @mousedown.prevent="markKeepDropdownOnBlur"
            @click.stop="removeHistoryCommand(cmd.insertText || cmd.command || cmd.name)"
          >×</button>
          <span v-if="cmd.children || cmd.needTarget" class="command-arrow">→</span>
        </div>
      </div>

      <!-- 输入区域 -->
      <div class="input-container">
        <div class="input-center-box" :style="{ width: inputWrapperWidth }">
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
            <button class="send-btn" :disabled="!canExecuteCommand" @click="executeCommand">
              <span class="send-icon">→</span>
            </button>
          </div>
          <div class="next-step-tip">{{ nextStepHint }}</div>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, nextTick, onMounted, onUnmounted, onActivated } from 'vue'
import { marked } from 'marked'
import DOMPurify from 'dompurify'
import module from '@/utils/module'
import commandConfig from '@/config/commandConfig.js'
import ssh from '@/utils/base/ssh_set'
import git from '@/utils/base/git'
import compose from '@/utils/base/compose'
import supervisor from '@/utils/base/supervisor'
import shell from '@/utils/base/shell'
import shellOut from '@/utils/base/shell_out'
import smartLinkSet from '@/utils/base/smart_link_set'
import variableSet from '@/utils/base/variable_set'
import group from '@/utils/base/group'
import store from '@/utils/base/store'
import sseDistribute from '@/utils/base/sse_distribute'
import { Throttle_string } from '@/utils/base/throttle_string'

export default {
  name: 'DashboardPage',
  setup() {
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
    const keepDropdownOnBlur = ref(false) // 删除历史候选时，避免 input blur 关闭下拉
    
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
    const commandHistoryCacheKey = 'dashboard_command_history_v1'
    const commandUsageCacheKey = 'dashboard_command_usage_v1'
    
    // SSE 相关状态
    const sseDistributeId = ref('') // SSE 分发 ID
    const isExecuting = ref(false) // 是否正在执行命令
    const currentOutputMessage = ref(null) // 当前输出消息的引用
    // variable 会话状态（用于首页快捷命令多步交互）
    const variableSession = ref({
      active: false,
      variableId: 0,
      variableName: '',
      runCmdId: 0,
      replaceList: {},
      isRun: 0,
      isFinish: 0,
      currentForm: null,
    })

    // 开放的模块列表
    const openModules = module.GetOpenModuleList()

    const normalizeCommandPart = (value) => {
      if (value === null || value === undefined) return ''
      return String(value).trim()
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

    // 获取 link run 命令在命令栈中的选择（环境(含配置)/账号）
    const getLinkRunSelection = (stack) => {
      const sourceStack = Array.isArray(stack) ? stack : []
      const actionIndex = sourceStack.findIndex(item => item?.action === 'linkRun')
      if (actionIndex < 0) {
        return {
          configCmd: null,
          envCmd: null,
          accountCmd: null
        }
      }
      const tailStack = sourceStack.slice(actionIndex + 1)
      const envCmd = tailStack.find(item => item?.data?.__linkType === 'env') || null
      return {
        // 兼容旧结构：如存在 configCmd 则使用；否则从 envCmd.data.config 反推出配置对象。
        configCmd: tailStack.find(item => item?.data?.__linkType === 'config') || (envCmd?.data?.config ? { data: envCmd.data.config } : null),
        envCmd,
        accountCmd: tailStack.find(item => item?.data?.__linkType === 'account') || null
      }
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
        return !!(selection.envCmd && selection.accountCmd)
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
        if (!selection.accountCmd) {
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

    const escapeHtml = (value) => {
      return String(value || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
    }

    // buildLinkEnvOptionsFromConfig 基于已选链接配置生成“环境”候选（高亮兜底）
    const buildLinkEnvOptionsFromConfig = (configCmd) => {
      const linkList = Array.isArray(configCmd?.data?.linkList) ? configCmd.data.linkList : []
      return linkList.map((item, index) => {
        const envName = normalizeCommandPart(item?.label) || `环境${index + 1}`
        return {
          command: envName,
          name: envName,
          dynamicChildren: 'linkAccountList',
          data: {
            __linkType: 'env',
            env: item || {},
            config: configCmd?.data || {}
          }
        }
      })
    }

    // buildLinkAccountOptionsFromEnv 基于已选环境生成“账号”候选（高亮兜底）
    const buildLinkAccountOptionsFromEnv = (envCmd) => {
      const userListRaw = Array.isArray(envCmd?.data?.env?.userList) ? envCmd.data.env.userList : []
      const userList = userListRaw.length > 0 ? userListRaw : [{ user_name: '默认账号(空)', password: '' }]
      return userList.map((item, index) => {
        const userName = normalizeCommandPart(item?.user_name) || `账号${index + 1}`
        return {
          command: userName,
          name: userName,
          data: {
            __linkType: 'account',
            account: {
              user_name: normalizeCommandPart(item?.user_name),
              password: normalizeCommandPart(item?.password)
            }
          }
        }
      })
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
      const html = marked.parse(raw)
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

    const refreshCommandDropdownVisibility = () => {
      showCommands.value = isCommandModeByText(inputText.value) &&
        (currentChildren.value.length > 0 || isLoadingDynamic.value) &&
        !isCommandReadyToExecute()
    }

    // 更新当前命令状态（running/success/failed）
    const updateCurrentCommandStatus = (status) => {
      if (!currentOutputMessage.value || !status) return
      currentOutputMessage.value.commandStatus = status
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

    const appendOutputProcess = (text) => {
      if (!currentOutputMessage.value) return
      const current = String(currentOutputMessage.value.processText || '')
      const merged = sanitizeCommandOutput(current + String(text || ''))
      currentOutputMessage.value.processText = merged.length > 50000 ? merged.slice(-50000) : merged
      scrollToBottom()
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

    // recentHistoryCommands 首页展示最近使用的历史命令（倒序）
    const recentHistoryCommands = computed(() => {
      if (!Array.isArray(commandHistory.value) || commandHistory.value.length === 0) {
        return []
      }
      return commandHistory.value.slice(-12).reverse()
    })

    // 命令面包屑导航
    const commandBreadcrumb = computed(() => {
      if (commandStack.value.length === 0) return ''
      return commandStack.value.map(c => c.name).join(' > ')
    })

    // 输入框提示
    const inputPlaceholder = computed(() => {
      if (commandStack.value.length === 0) {
        return '输入 / 或直接输入命令（如 g），Tab 补全，Space 继续...'
      }
      const lastCmd = commandStack.value[commandStack.value.length - 1]
      const actionCmd = commandStack.value.find(item => item.action)
      if (actionCmd && actionCmd.action === 'linkRun') {
        const selection = getLinkRunSelection(commandStack.value)
        if (!selection.envCmd) {
          return '请选择要执行的环境...'
        }
        if (!selection.accountCmd) {
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

    const filteredCommands = computed(() => {
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
      
      const sortedCommands = [...commands].sort(compareCommandByNaturalAsc)

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

    const canExecuteCommand = computed(() => commandAnalysis.value.canExecute)

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

    // 获取 variable 会话的下一步提示语（首页多步命令）
    const getVariableSessionStepHint = () => {
      const session = variableSession.value || {}
      if (!session.active || !session.variableId) {
        return ''
      }
      if (Number(session.isFinish) === 1) {
        return '下一步：输入 variable run <脚本名> 开始新会话'
      }
      if (Number(session.isRun) === 1) {
        return '下一步：输入 variable exec 并回车执行'
      }
      const cmdType = normalizeCommandPart(session.currentForm?.CmdType)
      if (['3', '17'].includes(cmdType)) {
        return '下一步：输入 variable set <值> 并回车'
      }
      if (['9', '12', '14'].includes(cmdType)) {
        return '下一步：输入 variable choose <选项> 并回车'
      }
      return '下一步：继续 variable 会话，或输入 variable cancel 取消'
    }

    // 计算首页命令行的下一步浅色提示文案
    const nextStepHint = computed(() => {
      if (isExecuting.value) {
        return '正在执行命令，请稍候...'
      }

      const variableHint = getVariableSessionStepHint()
      if (variableHint) {
        return variableHint
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
      showCommands.value = (currentChildren.value.length > 0 || isLoadingDynamic.value) && !isCommandReadyToExecute()
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
        type !== 'variableOptionList' &&
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
        case 'variableScriptList':
          loadVariableScriptList()
          break
        case 'variableOptionList':
          loadVariableOptionList()
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
      const sshId = store.getStore('dockerChooseSshId')
      if (!sshId) {
        ssh.SshList((response) => {
          if (response.ErrCode === 0 && response.Data.length > 0) {
            const firstSshId = response.Data[0].id
            fetchDockerComposeList(firstSshId)
          }
        })
      } else {
        fetchDockerComposeList(sshId)
      }
    }

    const fetchDockerComposeList = (sshId) => {
      compose.DockerComposeList({ ssh_id: sshId }, (response) => {
        isLoadingDynamic.value = false
        if (response.ErrCode === 0) {
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
          const list = response.Data.list.map(item => ({
            command: item.name,
            name: item.name,
            desc: item.compose_yml_path || '',
            id: item.id,
            data: item,
            // 保存 default_service_list 用于快速重启/停止
            default_service_list: normalizeDockerDefaultServices(item)
          }))
          dynamicDataCache.value['dockerComposeList'] = list
          currentChildren.value = list
          // Docker 项目列表为异步加载；当用户已输入 `docker quick-restart <项目>` 时，
          // 需要在数据到达后重新解析一次输入，才能自动进入“服务列表”层级。
          parseInput()
          refreshCommandDropdownVisibility()
        }
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
      const supervisorConfig = {
        ...envCmd.data
      }
      // 这里只是加载候选进程列表，不需要把终端回显串到首页命令执行流。
      supervisor.SupervisorConfList(supervisorConfig, (response) => {
        isLoadingDynamic.value = false
        if (!(response && response.ErrCode === 0)) {
          dynamicDataCache.value['supervisorProcessList'] = []
          currentChildren.value = []
          refreshCommandDropdownVisibility()
          return
        }
        const lines = String(response.Data || '')
          .split('\n')
          .map(line => normalizeCommandPart(line))
          .filter(Boolean)
        const list = lines.map((line, index) => {
          const [configNameRaw, supervisorNameRaw] = line.split('---')
          const configName = normalizeCommandPart(configNameRaw)
          let supervisorName = normalizeCommandPart(supervisorNameRaw)
            .replaceAll('[', '')
            .replaceAll(']', '')
            .replaceAll('program:', '')
          supervisorName = normalizeCommandPart(supervisorName)
          const configDir = normalizeCommandPart(envCmd.data.config_dir)
          const configPath = configDir && configName ? `${configDir}/${configName}` : configName
          const displayName = supervisorName || configName || `进程${index + 1}`
          return {
            command: displayName,
            name: displayName,
            aliases: [configName].filter(Boolean),
            desc: configName || '进程配置',
            id: `${envCmd.id || 'env'}_${index}`,
            data: {
              supervisor_name: supervisorName,
              supervisor_config: configPath
            }
          }
        })
        dynamicDataCache.value['supervisorProcessList'] = list
        currentChildren.value = list
        refreshCommandDropdownVisibility()
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
              aliases: [configName, `${configName}/${envName}`].filter(Boolean),
              desc: `${configName} | ${normalizeCommandPart(envItem?.link) || '未配置链接地址'}`,
              id: `${configItem?.id || 'cfg'}_${envIndex}`,
              dynamicChildren: 'linkAccountList',
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
        refreshCommandDropdownVisibility()
      })
    }

    // 加载已选环境下的账号列表
    const loadLinkAccountList = () => {
      const { configCmd, envCmd } = getLinkRunSelection(commandStack.value)
      const userListRaw = Array.isArray(envCmd?.data?.env?.userList) ? envCmd.data.env.userList : []
      const userList = userListRaw.length > 0
        ? userListRaw
        : [{ user_name: '默认账号(空)', password: '' }]
      const list = userList.map((item, index) => {
        const userName = normalizeCommandPart(item?.user_name) || `账号${index + 1}`
        return {
          command: userName,
          name: userName,
          desc: userListRaw.length > 0 ? '账号' : '该环境未配置账号，使用空账号执行',
          id: `${envCmd?.id || 'env'}_${index}`,
          data: {
            __linkType: 'account',
            account: {
              user_name: normalizeCommandPart(item?.user_name),
              password: normalizeCommandPart(item?.password)
            },
            env: envCmd?.data?.env || {},
            config: configCmd?.data || {}
          }
        }
      })
      dynamicDataCache.value['linkAccountList'] = list
      currentChildren.value = list
      isLoadingDynamic.value = false
      refreshCommandDropdownVisibility()
    }

    // 加载 variable 脚本列表
    const loadVariableScriptList = () => {
      variableSet.VariableList((response) => {
        isLoadingDynamic.value = false
        if (!(response && response.ErrCode === 0)) {
          dynamicDataCache.value['variableScriptList'] = []
          currentChildren.value = []
          refreshCommandDropdownVisibility()
          return
        }
        const variableList = Array.isArray(response.Data?.variable_list) ? response.Data.variable_list : []
        const list = variableList.map(item => ({
          command: normalizeCommandPart(item?.name) || `脚本${item?.id || ''}`,
          name: normalizeCommandPart(item?.name) || `脚本${item?.id || ''}`,
          aliases: [String(item?.id || '')].filter(Boolean),
          desc: normalizeCommandPart(item?.desc) || '自定义脚本',
          id: item?.id,
          data: item
        }))
        dynamicDataCache.value['variableScriptList'] = list
        currentChildren.value = list
        refreshCommandDropdownVisibility()
      })
    }

    // 加载 variable 当前步骤可选项
    const loadVariableOptionList = () => {
      const currentForm = variableSession.value.currentForm
      const cmdType = normalizeCommandPart(currentForm?.CmdType)
      const optionList = Array.isArray(currentForm?.Select?.OptionList) ? currentForm.Select.OptionList : []
      if (!['9', '12', '14'].includes(cmdType) || optionList.length === 0) {
        dynamicDataCache.value['variableOptionList'] = []
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
          id: `${variableSession.value.runCmdId || 'cmd'}_${index}`,
          data: {
            optionValue,
            optionLabel: label
          }
        }
      })
      dynamicDataCache.value['variableOptionList'] = list
      currentChildren.value = list
      isLoadingDynamic.value = false
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
      if (isCommandModeByText(inputText.value)) {
        parseInput()
        activeCommandIndex.value = 0
      } else {
        showCommands.value = false
        commandStack.value = []
        currentChildren.value = []
        currentInputValue.value = ''
      }
    }

    // 处理焦点
    const handleFocus = () => {
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
      try {
        const historyRaw = store.getStore(commandHistoryCacheKey)
        const usageRaw = store.getStore(commandUsageCacheKey)
        const parsedHistory = historyRaw ? JSON.parse(historyRaw) : []
        const parsedUsage = usageRaw ? JSON.parse(usageRaw) : {}
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
      } catch (e) {
        historyList = []
        usageMap = {}
      }
      commandHistory.value = historyList
      commandUsageMap.value = usageMap
      commandHistoryIndex.value = historyList.length
    }

    // 在输入框为空时，使用上下方向键切换历史命令
    const browseCommandHistory = (direction) => {
      if (commandHistory.value.length === 0) return
      const maxIndex = commandHistory.value.length
      // 到达边界继续按方向键时，清空输入框并退出历史浏览状态
      if (direction < 0 && commandHistoryIndex.value <= 0) {
        commandHistoryIndex.value = maxIndex
        inputText.value = ''
        showCommands.value = false
        commandStack.value = []
        currentChildren.value = []
        currentInputValue.value = ''
        return
      }
      if (direction > 0 && commandHistoryIndex.value >= maxIndex) {
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
        inputText.value = ''
        showCommands.value = false
        commandStack.value = []
        currentChildren.value = []
        currentInputValue.value = ''
        return
      }
      inputText.value = commandHistory.value[nextIndex]
      parseInput()
      activeCommandIndex.value = 0
    }

    // 固定一级命令面板点击：快速填充并进入下一层候选
    const quickSelectTopCommand = (cmd) => {
      const commandText = normalizeCommandPart(cmd?.command)
      if (!commandText) return
      inputText.value = `/${commandText} `
      parseInput()
      showCommands.value = true
      activeCommandIndex.value = 0
      nextTick(() => {
        inputRef.value?.focus()
      })
    }

    // quickSelectHistoryCommand 点击历史命令后回填到输入框
    const quickSelectHistoryCommand = (historyCommand) => {
      const commandText = normalizeCommandPart(historyCommand)
      if (!commandText) return
      // 历史命令回填后补一个空格，确保 parseInput 将最后一个 token 视为“已确认”。
      inputText.value = /\s$/.test(commandText) ? commandText : `${commandText} `
      parseInput()
      showCommands.value = isCommandModeByText(inputText.value)
      activeCommandIndex.value = 0
      nextTick(() => {
        inputRef.value?.focus()
      })
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
      if (e.key === 'Enter' && !canExecuteCommand.value) {
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
      const commandText = commandStack.value.map(c => c.command).join(' ')
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
      // history 选择项：仅回填输入框，不改变命令栈
      if (cmd && cmd.insertOnly) {
        const historyText = normalizeCommandPart(cmd.insertText || cmd.command || cmd.name)
        // history 通过 Tab 选中时也补空格，避免“还需再手动敲空格才可执行”。
        inputText.value = /\s$/.test(historyText) ? historyText : `${historyText} `
        showCommands.value = false
        parseInput()
        activeCommandIndex.value = 0
        nextTick(() => {
          inputRef.value?.focus()
        })
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
      if (!isSameCommandItem(stackLast, cmd)) {
        commandStack.value.push(cmd)
      }
      
      // 更新输入文本
      const tokenInfo = parseTokens(inputText.value)
      const prefix = tokenInfo.useSlash ? '/' : ''
      inputText.value = prefix + commandStack.value.map(c => c.command || c.name).join(' ') + ' '
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
        activeCommandIndex.value = 0
        return
      }
      
      if (cmd.dynamicChildren) {
        // 需要加载动态数据
        loadDynamicChildren(cmd.dynamicChildren)
        activeCommandIndex.value = 0
        return
      }

      // 支持“当前选择项自身声明了 nextDynamicChildren”的场景（如 git quick-create-branch 选择基线分支后切到分支类型）。
      if (cmd.nextDynamicChildren) {
        loadDynamicChildren(cmd.nextDynamicChildren)
        activeCommandIndex.value = 0
        return
      }
      
      if (cmd.needTarget) {
        // 需要选择目标，保持下拉框打开（等待动态数据加载）
        activeCommandIndex.value = 0
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
        resultText: '',
        processText: ''
      }
      messages.value.push(outputMsg)
      currentOutputMessage.value = outputMsg
      isExecuting.value = true
      
      // 清理输入状态
      inputText.value = ''
      showCommands.value = false
      const currentStack = [...commandStack.value]
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
        case 'shellCreate':
          executeShellAction('create', currentStack, options.inputValue || '')
          break
        case 'shellList':
          executeShellAction('list', currentStack, options.inputValue || '')
          break
        case 'shellRun':
          executeShellAction('run', currentStack, options.inputValue || '')
          break
        case 'linkRun':
          executeLinkAction(currentStack)
          break
        case 'variableRun':
          executeVariableRunAction(currentStack)
          break
        case 'variableSet':
          executeVariableSessionAction('set', currentStack, options.inputValue || '')
          break
        case 'variableChoose':
          executeVariableSessionAction('choose', currentStack, options.inputValue || '')
          break
        case 'variableExec':
          executeVariableSessionAction('exec', currentStack, options.inputValue || '')
          break
        case 'variableReset':
          executeVariableSessionAction('reset', currentStack, options.inputValue || '')
          break
        case 'variableCancel':
          executeVariableSessionAction('cancel', currentStack, options.inputValue || '')
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

    const getDockerSshId = (callback) => {
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

      getDockerSshId((sshId) => {
        if (!sshId) {
          appendOutputResult('错误：未找到可用 SSH 环境，请先在 /Docker 页面选择环境\n')
          finishExecution()
          return
        }

        const newSseDistributeId = sseDistribute.GetSseDistributeId('dashboard_docker_' + Date.now())
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

      const done = (response, successText) => {
        if (!(response && response.ErrCode === 0)) {
          appendOutputResult(`错误: ${normalizeCommandPart(response?.ErrMsg) || '未知错误'}\n`)
          setTimeout(() => {
            finishExecution()
          }, 1500)
        } else {
          if (successText) {
            appendOutputResult(`${successText}\n`)
          }
          // 结果详情统一在“执行过程(SSE)”里查看，上方只保留成功提示
          setTimeout(() => {
            finishExecution()
          }, 1500)
        }
      }

      switch (action) {
        case 'status':
          appendOutputResult(`正在查看环境 [${envCmd.name}] 的进程状态...\n\n`)
          supervisor.SupervisorStatusList({ ...supervisorConfig }, (response) => done(response, '执行成功'))
          break
        case 'restartAll':
          appendOutputResult(`正在重启环境 [${envCmd.name}] 的全部进程...\n\n`)
          supervisor.SupervisorRestartAll({ ...supervisorConfig }, (response) => done(response, '执行成功'))
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
            (response) => done(response, '执行成功'),
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
          supervisor.SupervisorStop({ ...supervisorConfig }, processCmd.data.supervisor_name, (response) => done(response, '执行成功'))
          break
        case 'config':
          if (!(processCmd && processCmd.data && processCmd.data.supervisor_config)) {
            appendOutputResult('错误：请先选择要查看配置的服务\n')
            finishExecution()
            return
          }
          appendOutputResult(`正在查看服务 [${processCmd.name}] 配置...\n\n`)
          supervisor.SupervisorConfigShow({ ...supervisorConfig }, processCmd.data.supervisor_config, (response) => done(response, '执行成功'))
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

      const newSseDistributeId = sseDistribute.GetSseDistributeId('dashboard_git_group_' + Date.now())
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
          finishExecution()
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
      const newSseDistributeId = sseDistribute.GetSseDistributeId('dashboard_git_' + Date.now())
      
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
          appendOutputResult('执行失败\n')
        } else {
          appendOutputResult(`${gitActionDoneLabelMap[action] || '执行成功'}\n`)
        }
        setTimeout(() => {
          // 给 SSE 尾包一点时间，避免过程/结果末尾被截断
          sseDistribute.UnRegisterReceive(newSseDistributeId)
          finishExecution()
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

    // 解析 shell create 输入参数：任务名 | SSH(名称或ID) | 命令
    const parseShellCreateInput = (inputValue) => {
      const text = String(inputValue || '')
      const parts = text.split('|').map(item => String(item || '').trim()).filter(Boolean)
      if (parts.length < 3) {
        return {
          ok: false,
          err: '参数格式错误，请使用: 任务名 | SSH(名称或ID) | 命令'
        }
      }
      const [name, sshKey, ...commandParts] = parts
      const command = commandParts.join(' | ').trim()
      if (!name || !sshKey || !command) {
        return {
          ok: false,
          err: '参数不完整，请使用: 任务名 | SSH(名称或ID) | 命令'
        }
      }
      return {
        ok: true,
        data: { name, sshKey, command }
      }
    }

    // 解析 SSH 标识：支持 SSH id 或 SSH 名称
    const resolveShellSshId = (sshKey, callback) => {
      const target = normalizeCommandPart(sshKey)
      if (!target) {
        callback('')
        return
      }
      ssh.SshList((response) => {
        if (!(response && response.ErrCode === 0 && Array.isArray(response.Data))) {
          callback('')
          return
        }
        const list = response.Data || []
        const exactId = list.find(item => String(item.id) === target)
        if (exactId) {
          callback(String(exactId.id))
          return
        }
        const lower = target.toLowerCase()
        const exactName = list.find(item => String(item.name || '').toLowerCase() === lower)
        if (exactName) {
          callback(String(exactName.id))
          return
        }
        const fuzzyName = list.find(item => String(item.name || '').toLowerCase().includes(lower))
        callback(fuzzyName ? String(fuzzyName.id) : '')
      })
    }

    // 获取 shell_out 分组 ID（与 ShellOut 页面保持一致：groupType=6）
    const resolveShellGroupId = (callback) => {
      group.GroupList({ type: '6' }, (response) => {
        if (!(response && response.ErrCode === 0 && Array.isArray(response.Data) && response.Data.length > 0)) {
          callback('')
          return
        }
        callback(String(response.Data[0].id || ''))
      })
    }

    // 执行终端输出相关动作：create/list/run
    const executeShellAction = (action, stack, inputValue) => {
      if (action === 'list') {
        const actionIndex = stack.findIndex(item => item.action === 'shellList')
        const targetCmd = actionIndex >= 0 ? stack[actionIndex + 1] : null
        if (!targetCmd || !targetCmd.data) {
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
          appendOutputResult('错误：任务ID为空，无法打开新窗口\n')
          finishExecution()
          return
        }
        const url = `${window.location.origin}/#/fullpage?group_id=${groupId}&id=${id}&title=${title}`
        window.open(url, '_blank')
        appendOutputResult('执行成功\n')
        finishExecution()
        return
      }

      if (action === 'run') {
        const actionIndex = stack.findIndex(item => item.action === 'shellRun')
        const targetCmd = actionIndex >= 0 ? stack[actionIndex + 1] : null
        if (!targetCmd || !targetCmd.data) {
          appendOutputResult('错误：请先选择要运行的终端输出任务\n')
          finishExecution()
          return
        }
        const target = targetCmd.data
        const groupId = Number(target.group_id || 0)
        if (!groupId) {
          appendOutputResult('错误：任务缺少 group_id，无法启动\n')
          finishExecution()
          return
        }

        const newSseDistributeId = sseDistribute.GetSseDistributeId('dashboard_shell_' + Date.now())
        const throttleStringFunc = new Throttle_string(50, (text) => {
          if (currentOutputMessage.value) {
            appendOutputProcess(text)
          }
        })
        sseDistribute.RegisterReceive(newSseDistributeId, (msg) => {
          throttleStringFunc.update(msg)
        })

        appendOutputResult(`正在运行任务 [${target.name || target.id}]...\n\n`)
        shell.ShellOutSetSeeId({
          sse_distribute_id: newSseDistributeId,
          shell_client_id: target.shell_client_id,
          ssh_id: target.ssh_id,
          command: target.command,
          id: target.id,
          group_id: target.group_id,
          is_run: 1,
        }, (response) => {
          if (response.ErrCode !== 0) {
            appendOutputResult('执行失败\n')
          } else {
            appendOutputResult('执行成功\n')
          }
          setTimeout(() => {
            sseDistribute.UnRegisterReceive(newSseDistributeId)
            finishExecution()
          }, 1200)
        })
        return
      }

      if (action === 'create') {
        const parsed = parseShellCreateInput(inputValue)
        if (!parsed.ok) {
          appendOutputResult(`${parsed.err}\n`)
          appendOutputResult('示例: shell create 发布日志 | 1 | tail -f /var/log/app.log\n')
          finishExecution()
          return
        }
        const payload = parsed.data
        resolveShellSshId(payload.sshKey, (sshId) => {
          if (!sshId) {
            appendOutputResult(`错误：未找到 SSH "${payload.sshKey}"\n`)
            finishExecution()
            return
          }
          resolveShellGroupId((groupId) => {
            if (!groupId) {
              appendOutputResult('错误：未找到 shell_out 分组，请先在“终端输出”页面创建分组\n')
              finishExecution()
              return
            }
            // 复用 shellOut 页面创建逻辑的接口
            shell.ShellOutStart({
              id: '',
              command: payload.command,
              sse_distribute_id: '',
              shell_client_id: '',
              ssh_id: sshId,
              name: payload.name,
              is_run: 1,
              group_id: groupId,
            }, (response) => {
              if (response.ErrCode !== 0) {
                appendOutputResult('执行失败\n')
              } else {
                appendOutputResult('执行成功\n')
              }
              finishExecution()
            })
          })
        })
        return
      }

      appendOutputResult('该终端输出操作暂未实现\n')
      finishExecution()
    }

    // 执行 link run：根据“环境(含配置) -> 账号”选择启动自定义链接
    const executeLinkAction = (stack) => {
      const selection = getLinkRunSelection(stack)
      if (!selection.envCmd || !selection.accountCmd) {
        appendOutputResult('错误：请完整选择环境和账号\n')
        finishExecution()
        return
      }

      const configData = (selection.configCmd?.data || selection.envCmd?.data?.config || {})
      const envData = selection.envCmd.data?.env || {}
      const accountData = selection.accountCmd.data?.account || {}

      const payload = {
        id: configData.id,
        label: normalizeCommandPart(envData.label),
        user_name: normalizeCommandPart(accountData.user_name),
        password: normalizeCommandPart(accountData.password),
        open_num: normalizeCommandPart(configData.open_num),
        open_type: normalizeCommandPart(configData.open_type),
        sse_distribute_id: sseDistributeId.value
      }

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
        finishExecution()
      })
    }

    // 重置 variable 会话
    const resetVariableSession = () => {
      variableSession.value = {
        active: false,
        variableId: 0,
        variableName: '',
        runCmdId: 0,
        replaceList: {},
        isRun: 0,
        isFinish: 0,
        currentForm: null,
      }
      dynamicDataCache.value['variableOptionList'] = []
    }

    // 处理 variable API 返回，更新会话与下一步提示
    const handleVariableFlowResponse = (response) => {
      if (!(response && response.ErrCode === 0)) {
        appendOutputResult(`执行失败: ${normalizeCommandPart(response?.ErrMsg) || '未知错误'}\n`)
        finishExecution()
        return
      }
      const data = response.Data || {}
      const runStatus = Number(data.RunStatus)
      const currentForm = data.Form || null
      if (data.ReplaceList && typeof data.ReplaceList === 'object') {
        variableSession.value.replaceList = data.ReplaceList
      }
      if (currentForm && currentForm.Id !== undefined && currentForm.Id !== null) {
        variableSession.value.runCmdId = Number(currentForm.Id) || 0
      }
      variableSession.value.currentForm = currentForm
      variableSession.value.active = true
      variableSession.value.isFinish = runStatus === 2 ? 1 : 0
      variableSession.value.isRun = runStatus === 1 ? 1 : 0

      if (runStatus === 0) {
        const cmdType = normalizeCommandPart(currentForm?.CmdType)
        if (['3', '17'].includes(cmdType)) {
          const label = normalizeCommandPart(currentForm?.Input?.Label) || '输入参数'
          appendOutputResult(`当前步骤: ${label}\n下一步: variable set <值>\n`)
        } else if (['9', '12', '14'].includes(cmdType)) {
          const options = Array.isArray(currentForm?.Select?.OptionList) ? currentForm.Select.OptionList : []
          const optionText = options.map(item => normalizeCommandPart(item?.Label) || normalizeCommandPart(item?.Value)).filter(Boolean).join('、')
          appendOutputResult(`当前步骤: ${normalizeCommandPart(currentForm?.Select?.Label) || '选择选项'}\n可选项: ${optionText || '无'}\n下一步: variable choose <选项>\n`)
        } else {
          appendOutputResult('脚本返回了未适配的步骤类型，请到 Variable 页面执行。\n')
        }
        finishExecution()
        return
      }
      if (runStatus === 1) {
        appendOutputResult('当前脚本已就绪，下一步: variable exec\n')
        finishExecution()
        return
      }
      if (runStatus === 2) {
        appendOutputResult('执行完成\n')
        resetVariableSession()
        finishExecution()
        return
      }
      appendOutputResult('收到未知状态，已保留当前会话\n')
      finishExecution()
    }

    // 执行 variable run：选择脚本并启动
    const executeVariableRunAction = (stack) => {
      const actionIndex = stack.findIndex(item => item.action === 'variableRun')
      const targetCmd = actionIndex >= 0 ? stack[actionIndex + 1] : null
      if (!(targetCmd && targetCmd.data && targetCmd.data.id)) {
        appendOutputResult('错误：请先选择要执行的脚本\n')
        finishExecution()
        return
      }
      resetVariableSession()
      variableSession.value.active = true
      variableSession.value.variableId = Number(targetCmd.data.id) || 0
      variableSession.value.variableName = normalizeCommandPart(targetCmd.data.name) || normalizeCommandPart(targetCmd.name)
      appendOutputResult(`已启动脚本会话: ${variableSession.value.variableName || variableSession.value.variableId}\n`)
      variableSet.VariableRun(
        sseDistributeId.value,
        variableSession.value.variableId,
        0,
        0,
        JSON.stringify({}),
        (response) => {
          handleVariableFlowResponse(response)
        }
      )
    }

    // 执行 variable 会话动作：set/choose/exec/reset/cancel
    const executeVariableSessionAction = (action, stack, inputValue) => {
      const session = variableSession.value
      if (action === 'reset') {
        resetVariableSession()
        appendOutputResult('已重置 variable 会话，可重新执行 variable run <脚本名>\n')
        finishExecution()
        return
      }
      if (action === 'cancel') {
        resetVariableSession()
        appendOutputResult('已取消 variable 会话\n')
        finishExecution()
        return
      }
      if (!session.active || !session.variableId) {
        appendOutputResult('当前没有进行中的 variable 会话，请先执行 variable run <脚本名>\n')
        finishExecution()
        return
      }

      const currentForm = session.currentForm || {}
      const cmdType = normalizeCommandPart(currentForm?.CmdType)

      if (action === 'set') {
        if (!['3', '17'].includes(cmdType)) {
          appendOutputResult('当前步骤不是输入步骤，请使用 variable choose <选项>\n')
          finishExecution()
          return
        }
        const editValue = normalizeCommandPart(inputValue)
        if (!editValue) {
          appendOutputResult('请输入参数值，例如: variable set 123\n')
          finishExecution()
          return
        }
        variableSet.VariableSet(
          session.variableId,
          Number(currentForm.Id) || Number(session.runCmdId) || 0,
          JSON.stringify(session.replaceList || {}),
          editValue,
          (response) => {
            handleVariableFlowResponse(response)
          }
        )
        return
      }

      if (action === 'choose') {
        if (!['9', '12', '14'].includes(cmdType)) {
          appendOutputResult('当前步骤不是选项步骤，请使用 variable set <值>\n')
          finishExecution()
          return
        }
        const actionIndex = stack.findIndex(item => item.action === 'variableChoose')
        const targetCmd = actionIndex >= 0 ? stack[actionIndex + 1] : null
        const selectedValue = normalizeCommandPart(targetCmd?.data?.optionValue) || normalizeCommandPart(inputValue)
        const options = Array.isArray(currentForm?.Select?.OptionList) ? currentForm.Select.OptionList : []
        const matched = options.find(item => {
          const label = normalizeCommandPart(item?.Label)
          const value = normalizeCommandPart(item?.Value)
          return selectedValue && (selectedValue === label || selectedValue === value)
        })
        if (!matched) {
          const optionText = options.map(item => normalizeCommandPart(item?.Label) || normalizeCommandPart(item?.Value)).filter(Boolean).join('、')
          appendOutputResult(`选项不存在，可选项: ${optionText || '无'}\n`)
          finishExecution()
          return
        }
        variableSet.VariableSet(
          session.variableId,
          Number(currentForm.Id) || Number(session.runCmdId) || 0,
          JSON.stringify(session.replaceList || {}),
          normalizeCommandPart(matched?.Value),
          (response) => {
            handleVariableFlowResponse(response)
          }
        )
        return
      }

      if (action === 'exec') {
        if (Number(session.isRun) !== 1) {
          appendOutputResult('当前步骤尚未就绪，不能最终执行\n')
          finishExecution()
          return
        }
        variableSet.VariableRun(
          sseDistributeId.value,
          session.variableId,
          Number(session.runCmdId) || Number(currentForm.Id) || 0,
          1,
          JSON.stringify(session.replaceList || {}),
          (response) => {
            handleVariableFlowResponse(response)
          }
        )
        return
      }

      appendOutputResult('未知 variable 操作\n')
      finishExecution()
    }
    
    // 完成执行
    const finishExecution = () => {
      // 若未显式写入成功/失败状态，执行结束后默认标记为成功
      if (currentOutputMessage.value && currentOutputMessage.value.commandStatus === 'running') {
        currentOutputMessage.value.commandStatus = 'success'
      }
      isExecuting.value = false
      currentOutputMessage.value = null
      scrollToBottom()
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
      showCommands,
      isLoadingDynamic,
      filteredCommands,
      activeCommandIndex,
      inputRef,
      messageList,
      commandBreadcrumb,
      inputPlaceholder,
      canExecuteCommand,
      highlightedInputHtml,
      inputWrapperWidth,
      nextStepHint,
      availableCommands,
      recentHistoryCommands,
      handleInput,
      handleKeydown,
      handleFocus,
      handleBlur,
      quickSelectTopCommand,
      quickSelectHistoryCommand,
      removeHistoryCommand,
      markKeepDropdownOnBlur,
      selectCommand,
      executeCommand,
      getCommandKey,
      getCommandMatchHint,
      renderProcessMarkdown,
      hasCommandLayout,
    }
  }
}
</script>

<style scoped>
.dashboard-container {
  height: 100%;
  display: flex;
  justify-content: stretch;
  align-items: stretch;
  padding: 0;
  background: #fafaf7;
  box-sizing: border-box;
}

.chat-container {
  width: 100%;
  height: 100%;
  background: #fff;
  border-radius: 12px;
  display: flex;
  flex-direction: column;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border: 1px solid #e8e8e0;
  position: relative;
}

.message-list {
  flex: 1;
  overflow-y: auto;
  padding: 24px;
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.welcome-message {
  text-align: center;
  padding: 40px 20px;
  color: #8a8a7a;
}

.welcome-message h2 {
  color: #4a4a4a;
  margin-bottom: 16px;
  font-size: 26px;
  font-weight: 600;
}

.welcome-message .hint {
  font-size: 15px;
}

.welcome-message kbd {
  background: #f0f0e8;
  padding: 4px 10px;
  border-radius: 4px;
  border: 1px solid #d8d8c8;
  font-family: monospace;
  color: #5a8a5a;
}

.fixed-command-panel {
  margin: 0 auto 16px;
  max-width: 980px;
  text-align: left;
}

.fixed-command-title {
  margin-bottom: 10px;
  font-size: 13px;
  color: #6e7d6e;
}

.fixed-command-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(170px, 1fr));
  gap: 8px;
}

.history-command-section {
  margin-top: 14px;
}

.history-command-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.history-command-item-wrap {
  display: inline-flex;
  align-items: center;
  max-width: 100%;
  border: 1px solid #d6e3d2;
  background: #f6fbf4;
  color: #3f6f3f;
  border-radius: 999px;
  transition: all 0.18s ease;
}

.history-command-item-wrap:hover {
  border-color: #a9c3a4;
  background: #eaf4e7;
  color: #2f5c2f;
}

.history-command-item {
  border: 0;
  background: transparent;
  color: inherit;
  padding: 6px 8px 6px 12px;
  font-size: 12px;
  line-height: 1.2;
  cursor: pointer;
  max-width: 100%;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.history-command-delete {
  border: 0;
  background: transparent;
  color: #7a8f7a;
  width: 22px;
  height: 22px;
  margin-right: 4px;
  border-radius: 999px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  line-height: 1;
}

.history-command-delete:hover {
  background: rgba(63, 111, 63, 0.12);
  color: #2f5c2f;
}

.fixed-command-item {
  display: flex;
  align-items: center;
  gap: 8px;
  width: 100%;
  min-height: 42px;
  border: 1px solid #e0e7da;
  background: #f8fbf5;
  border-radius: 8px;
  padding: 6px 8px;
  cursor: pointer;
  text-align: left;
  color: #4f5f4f;
}

.fixed-command-item:hover {
  border-color: #bfd1bf;
  background: #eef5ea;
}

.fixed-command-icon {
  flex-shrink: 0;
}

.fixed-command-name {
  font-weight: 600;
  color: #3f533f;
}

.fixed-command-desc {
  margin-left: auto;
  font-size: 12px;
  color: #7f8c7f;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.message {
  max-width: 100%;
  animation: fadeIn 0.3s ease;
}

@keyframes fadeIn {
  from {
    opacity: 0;
    transform: translateY(10px);
  }
  to {
    opacity: 1;
    transform: translateY(0);
  }
}

.message.user {
  align-self: flex-end;
}

.message.system {
  align-self: flex-start;
  /* 执行过程(SSE)卡片在首页尽量占满可用宽度 */
  width: 100%;
}

.message-command {
  font-size: 12px;
  color: #5a8a5a;
  margin-bottom: 8px;
  padding: 0 4px;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.message-command-text {
  min-width: 0;
  word-break: break-word;
}

.command-status {
  flex: 0 0 auto;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 16px;
}

.command-status-icon {
  font-size: 13px;
  font-weight: 700;
  line-height: 1;
}

.command-status-running {
  color: #7e8f73;
}

.command-status-success {
  color: #2fa35f;
}

.command-status-failed {
  color: #d84a4a;
}

.command-status-spinner {
  width: 12px;
  height: 12px;
  border-radius: 50%;
  border: 2px solid #c9d5bf;
  border-top-color: #6d8f5b;
  animation: command-status-spin 0.8s linear infinite;
}

@keyframes command-status-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.message-content {
  padding: 6px 12px;
  border-radius: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

.process-window {
  margin-top: 8px;
  border: 1px solid #d8e3d2;
  border-radius: 10px;
  background: #edf3e9;
  color: #435244;
  overflow: hidden;
}

.process-title {
  font-size: 12px;
  color: #5a6d5a;
  background: #e4ecdf;
  padding: 6px 10px;
  border-bottom: 1px solid #d3dfcd;
}

.process-text {
  margin: 0;
  padding: 10px 12px;
  max-height: 240px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 13px;
  line-height: 1.45;
  font-family: Consolas, Monaco, 'Courier New', monospace;
}

.process-text.markdown-body {
  color: #435244;
  background: transparent;
}

.process-text.markdown-body :deep(*) {
  color: #435244;
}

.process-text.markdown-body :deep(table) {
  width: 100%;
  border-collapse: collapse;
  background: transparent;
}

.process-text.markdown-body :deep(th),
.process-text.markdown-body :deep(td) {
  border: 1px solid #d3dfcd;
  padding: 4px 6px;
}

.process-text.markdown-body :deep(th) {
  background: #e4ecdf;
}

.process-text.markdown-body :deep(td) {
  background: transparent;
}

.process-text.markdown-body :deep(hr) {
  background-color: #d3dfcd;
  border: 0;
  height: 1px;
}

.process-text.markdown-body :deep(code) {
  background: #e4ecdf;
}

.process-text.markdown-body :deep(pre) {
  background: #e4ecdf;
  border: 1px solid #d3dfcd;
}

.process-text.markdown-body :deep(a) {
  color: #4f7d5f;
}

.message.user .message-content {
  background: linear-gradient(135deg, #7cb87c 0%, #8fc88f 100%);
  color: #fff;
}

.message.system .message-content {
  background: #f5f5f0;
  color: #5a5a5a;
  border: 1px solid #e0e0d8;
}

@media (max-width: 768px) {
  .fixed-command-list {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .history-command-list {
    gap: 6px;
  }

  .history-command-item {
    padding: 5px 10px;
    font-size: 11px;
  }

  .history-command-delete {
    width: 20px;
    height: 20px;
    margin-right: 3px;
  }

  .fixed-command-desc {
    display: none;
  }

  .message {
    max-width: 100%;
  }

  .message.system {
    width: 100%;
  }
}

.command-dropdown {
  position: absolute;
  bottom: 80px;
  left: 24px;
  right: 24px;
  background: #fff;
  border: 1px solid #e0e0d8;
  border-radius: 10px;
  max-height: 450px;
  overflow-y: auto;
  z-index: 100;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.command-item {
  display: flex;
  align-items: center;
  padding: 6px 12px;
  cursor: pointer;
  transition: background 0.15s;
  border-bottom: 1px solid #f0f0e8;
  white-space: nowrap;
  overflow: hidden;
}

.command-item:last-child {
  border-bottom: none;
}

.command-item:hover,
.command-item.active {
  background: #f5f8f5;
}

.command-icon {
  font-size: 16px;
  margin-right: 8px;
  width: 20px;
  text-align: center;
}

.command-name {
  font-weight: 500;
  color: #4a4a4a;
  margin-right: 8px;
  min-width: 70px;
  flex-shrink: 0;
}

.command-desc {
  color: #8a8a7a;
  font-size: 12px;
  flex: 1;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.command-arrow {
  color: #c0c0b8;
  font-size: 14px;
  margin-left: 8px;
}

.command-item-delete {
  border: 0;
  background: transparent;
  color: #8da08d;
  width: 22px;
  height: 22px;
  margin-left: 8px;
  border-radius: 999px;
  cursor: pointer;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  font-size: 14px;
  line-height: 1;
  flex-shrink: 0;
}

.command-item-delete:hover {
  background: rgba(63, 111, 63, 0.12);
  color: #2f5c2f;
}

.command-breadcrumb {
  padding: 10px 16px;
  background: #f5f8f5;
  border-bottom: 1px solid #e8e8e0;
  border-radius: 10px 10px 0 0;
}

.command-loading {
  display: flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
  border-bottom: 1px solid #eef2eb;
  background: #f9fbf7;
}

.command-loading-text {
  font-size: 12px;
  color: #7b8b79;
}

.breadcrumb-text {
  font-size: 12px;
  color: #5a8a5a;
  font-weight: 500;
}

.input-container {
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 16px 24px;
  border-top: 1px solid #e8e8e0;
  background: #fff;
  border-radius: 0 0 12px 12px;
}

.input-center-box {
  max-width: 100%;
  transition: width 0.18s ease;
}

.input-wrapper {
  width: 100%;
  display: flex;
  align-items: center;
  background: #fafaf7;
  border: 1px solid #d8d8c8;
  border-radius: 10px;
  padding: 4px;
  transition: border-color 0.2s;
}

.input-wrapper:focus-within {
  border-color: #8fc88f;
}

.next-step-tip {
  width: 100%;
  min-height: 18px;
  margin-top: 8px;
  padding: 0 4px;
  font-size: 12px;
  line-height: 1.4;
  color: #9aa79a;
}

.input-overlay-box {
  position: relative;
  flex: 1;
  overflow: hidden;
}

.input-highlight-layer {
  position: absolute;
  inset: 0;
  pointer-events: none;
  display: block;
  padding: 6px 12px;
  font-size: 15px;
  line-height: normal;
  font-family: inherit;
  font-weight: inherit;
  letter-spacing: inherit;
  white-space: pre;
  overflow: hidden;
  color: #4a4a4a;
}

.chat-input {
  flex: 1;
  background: transparent;
  border: none;
  padding: 6px 12px;
  font-size: 15px;
  line-height: normal;
  font-family: inherit;
  font-weight: inherit;
  letter-spacing: inherit;
  color: #4a4a4a;
  outline: none;
  width: 100%;
  position: relative;
  z-index: 1;
}

.chat-input-overlay {
  color: transparent;
  caret-color: #4a4a4a;
}

.chat-input::placeholder {
  color: #a0a090;
}

:deep(.token-bg) {
  border-radius: 4px;
  padding: 0;
}

:deep(.token-bg-valid) {
  background: rgba(95, 180, 95, 0.25);
  color: #246524;
}

:deep(.token-bg-invalid) {
  background: rgba(220, 80, 80, 0.22);
  color: #922f2f;
}

:deep(.token-bg-arg) {
  background: rgba(95, 180, 95, 0.2);
  color: #2e6b2e;
}

.send-btn {
  background: linear-gradient(135deg, #7cb87c 0%, #8fc88f 100%);
  border: none;
  border-radius: 8px;
  padding: 10px 16px;
  cursor: pointer;
  transition: transform 0.15s, box-shadow 0.15s;
}

.send-btn:hover {
  transform: scale(1.05);
  box-shadow: 0 4px 12px rgba(120, 180, 120, 0.3);
}

.send-btn:disabled {
  cursor: not-allowed;
  opacity: 0.5;
  transform: none;
  box-shadow: none;
}

.send-icon {
  color: #fff;
  font-size: 16px;
  font-weight: bold;
}

/* 滚动条样式 */
.message-list::-webkit-scrollbar,
.command-dropdown::-webkit-scrollbar {
  width: 6px;
}

.message-list::-webkit-scrollbar-track,
.command-dropdown::-webkit-scrollbar-track {
  background: transparent;
}

.message-list::-webkit-scrollbar-thumb,
.command-dropdown::-webkit-scrollbar-thumb {
  background: #d0d0c8;
  border-radius: 3px;
}

.message-list::-webkit-scrollbar-thumb:hover,
.command-dropdown::-webkit-scrollbar-thumb:hover {
  background: #b8b8a8;
}
</style>
