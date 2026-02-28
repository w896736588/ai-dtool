<template>
  <div class="dashboard-container">
    <div class="chat-container">
      <!-- 消息列表区域 -->
      <div ref="messageList" class="message-list">
        <div class="welcome-message">
          <h2>开发者工具平台</h2>
          <p class="hint">输入 <kbd>/</kbd> 或直接输入命令（如 <kbd>g</kbd>），<kbd>Tab</kbd> 补全，<kbd>Space</kbd> 继续</p>
        </div>
        <div
          v-for="(msg, index) in messages"
          :key="index"
          :class="['message', msg.type]"
        >
          <template v-if="hasCommandLayout(msg)">
            <div class="message-command">{{ msg.commandText }}</div>
            <div v-if="msg.resultText" class="message-content">{{ msg.resultText }}</div>
            <div v-if="msg.processText" class="process-window">
              <div class="process-title">执行过程 (SSE)</div>
              <pre class="process-text">{{ msg.processText }}</pre>
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
        <div
          v-for="(cmd, index) in filteredCommands"
          :key="getCommandKey(cmd, index)"
          :class="['command-item', { active: activeCommandIndex === index }]"
          @click="selectCommand(cmd)"
          @mouseenter="activeCommandIndex = index"
        >
          <span class="command-icon">{{ cmd.icon }}</span>
          <span class="command-name">{{ cmd.name }}</span>
          <span class="command-desc">{{ cmd.desc }}</span>
          <span v-if="cmd.children || cmd.needTarget" class="command-arrow">→</span>
        </div>
      </div>

      <!-- 输入区域 -->
      <div class="input-container">
        <div class="input-wrapper">
          <input
            ref="inputRef"
            v-model="inputText"
            type="text"
            class="chat-input"
            :placeholder="inputPlaceholder"
            @input="handleInput"
            @keydown="handleKeydown"
            @blur="handleBlur"
            @focus="handleFocus"
          />
          <button class="send-btn" @click="executeCommand">
            <span class="send-icon">→</span>
          </button>
        </div>
      </div>
    </div>
  </div>
</template>

<script>
import { ref, computed, nextTick, onMounted, onUnmounted } from 'vue'
import module from '@/utils/module'
import commandConfig from '@/config/commandConfig.js'
import ssh from '@/utils/base/ssh_set'
import git from '@/utils/base/git'
import compose from '@/utils/base/compose'
import supervisor from '@/utils/base/supervisor'
import shellOut from '@/utils/base/shell_out'
import store from '@/utils/base/store'
import sseDistribute from '@/utils/base/sse_distribute'
import { Throttle_string } from '@/utils/base/throttle_string'

export default {
  name: 'DashboardPage',
  setup() {
    const inputText = ref('')
    const messages = ref([])
    const showCommands = ref(false)
    const activeCommandIndex = ref(0)
    const inputRef = ref(null)
    const messageList = ref(null)
    
    // 多级命令状态
    const commandStack = ref([]) // 命令栈，存储已选择的命令
    const currentChildren = ref([]) // 当前可选的子命令
    const dynamicDataCache = ref({}) // 动态数据缓存
    const isLoadingDynamic = ref(false) // 是否正在加载动态数据
    const currentInputValue = ref('')
    
    // SSE 相关状态
    const sseDistributeId = ref('') // SSE 分发 ID
    const isExecuting = ref(false) // 是否正在执行命令
    const currentOutputMessage = ref(null) // 当前输出消息的引用

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

    const refreshCommandDropdownVisibility = () => {
      showCommands.value = isCommandModeByText(inputText.value) && currentChildren.value.length > 0
    }

    const appendOutputResult = (text) => {
      if (!currentOutputMessage.value) return
      const current = String(currentOutputMessage.value.resultText || '')
      const merged = current + String(text || '')
      currentOutputMessage.value.resultText = merged.length > 50000 ? merged.slice(-50000) : merged
      scrollToBottom()
    }

    const appendOutputProcess = (text) => {
      if (!currentOutputMessage.value) return
      const current = String(currentOutputMessage.value.processText || '')
      const merged = current + String(text || '')
      currentOutputMessage.value.processText = merged.length > 50000 ? merged.slice(-50000) : merged
      scrollToBottom()
    }

    // 根据模块配置过滤可用命令
    const availableCommands = computed(() => {
      return commandConfig.filter(cmd => {
        if (cmd.module === null) return true
        return openModules.includes(cmd.module)
      })
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
    const filteredCommands = computed(() => {
      let commands = currentChildren.value.length > 0 
        ? currentChildren.value
        : (commandStack.value.length === 0 ? availableCommands.value : [])
      
      // 获取当前输入的搜索文本
      const tokenInfo = parseTokens(inputText.value)
      const parts = tokenInfo.parts
      const rawSearchText = parts.length > 0
        ? normalizeCommandPart(parts[parts.length - 1]).toLowerCase().replace('/', '')
        : ''
      let searchText = rawSearchText

      // 场景：已完整输入动作词（如 git checkout），当前候选已切到“目标列表”
      // 这时不应再用动作词过滤目标，否则会把项目列表全部过滤为空。
      if (commandStack.value.length > 0 && commands.length > 0) {
        const lastCmd = commandStack.value[commandStack.value.length - 1]
        const hasTrailingSpace = /\s$/.test(String(inputText.value || ''))
        if (!hasTrailingSpace && lastCmd?.needTarget) {
          const lastCmdKeywords = getCommandKeywords(lastCmd)
          if (lastCmdKeywords.some(keyword => keyword === rawSearchText)) {
            searchText = ''
          }
        }
      }
      
      if (!searchText) {
        return commands
      }
      
      return commands.filter(cmd => {
        const keywords = getCommandKeywords(cmd)
        return keywords.some(keyword => keyword.includes(searchText))
      })
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
      
      // 重置状态
      commandStack.value = []
      currentChildren.value = []
      currentInputValue.value = ''
      
      let currentLevel = availableCommands.value
      
      for (let i = 0; i < parts.length; i++) {
        const part = parts[i].toLowerCase()
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
            const targetToken = parts[i + 1]
            if (targetToken) {
              const targetFound = findCommandByToken(dynamicList, targetToken)
              if (targetFound) {
                commandStack.value.push(targetFound)
                i += 1
                currentChildren.value = []
                if (found.needInput) {
                  currentInputValue.value = parts.slice(i + 1).join(' ')
                }
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
      showCommands.value = currentChildren.value.length > 0
    }

    // 加载动态子命令
    const loadDynamicChildren = (type) => {
      if (type !== 'gitProjectList' && dynamicDataCache.value[type]) {
        currentChildren.value = dynamicDataCache.value[type]
        refreshCommandDropdownVisibility()
        return
      }
      
      isLoadingDynamic.value = true
      
      switch (type) {
        case 'dockerComposeList':
          loadDockerComposeList()
          break
        case 'gitProjectList':
          loadGitProjectList()
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
          const list = response.Data.list.map(item => ({
            command: item.name,
            name: item.name,
            desc: item.compose_yml_path || '',
            id: item.id,
            data: item,
            // 保存 default_service_list 用于快速重启/停止
            default_service_list: item.default_service_list || []
          }))
          dynamicDataCache.value['dockerComposeList'] = list
          currentChildren.value = list
        }
      })
    }

    // 加载 Docker 服务列表（用于快速重启/停止）
    const loadDockerServiceList = () => {
      // 从命令栈中找到已选择的项目
      const projectCmd = commandStack.value.find(cmd => cmd.data && cmd.data.default_service_list)
      
      if (projectCmd && projectCmd.data.default_service_list) {
        const services = projectCmd.data.default_service_list
        const list = services.map(service => ({
          command: service,
          name: service,
          desc: '服务',
          data: { service, projectId: projectCmd.id }
        }))
        dynamicDataCache.value['dockerServiceList'] = list
        currentChildren.value = list
        isLoadingDynamic.value = false
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
          currentChildren.value = list
          refreshCommandDropdownVisibility()
        }
      })
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
      const supervisorId = store.getStore('chooseSupervisorId')
      if (!supervisorId) {
        loadSupervisorEnvList()
        return
      }
      // 这里需要根据环境获取进程列表，简化处理
      loadSupervisorEnvList()
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

    // 加载 Redis 环境列表
    const loadRedisEnvList = () => {
      // 简化处理，后续可以扩展
      dynamicDataCache.value['redisEnvList'] = []
      currentChildren.value = []
      isLoadingDynamic.value = false
    }

    // 处理输入
    const handleInput = () => {
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

    // 处理失焦
    const handleBlur = () => {
      setTimeout(() => {
        showCommands.value = false
      }, 200)
    }

    // 处理键盘事件
    const handleKeydown = (e) => {
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
        case 'Enter':
          e.preventDefault()
          parseInput()
          {
            const actionCmd = commandStack.value.find(item => item.action)
            if (actionCmd) {
              const actionIndex = commandStack.value.findIndex(item => item.action)
              const targetCmd = actionCmd.needTarget ? commandStack.value[actionIndex + 1] : null
              const targetReady = !actionCmd.needTarget || !!(targetCmd && targetCmd.data)
              const inputReady = !actionCmd.needInput || !!normalizeCommandPart(currentInputValue.value)
              if (targetReady && inputReady) {
                executeCommand()
                break
              }
            }
          }
          if (filteredCommands.value[activeCommandIndex.value]) {
            selectCommand(filteredCommands.value[activeCommandIndex.value])
          } else {
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
      // 构建新的输入文本
      const parts = inputText.value.split(' ')
      parts[parts.length - 1] = cmd.command || cmd.name
      
      // 获取父命令（在选择前）
      const parentCmd = commandStack.value.length > 0 
        ? commandStack.value[commandStack.value.length - 1] 
        : null

      // 添加到命令栈
      commandStack.value.push(cmd)
      
      // 更新输入文本
      const tokenInfo = parseTokens(inputText.value)
      const prefix = tokenInfo.useSlash ? '/' : ''
      inputText.value = prefix + commandStack.value.map(c => c.command || c.name).join(' ') + ' '
      
      // 检查父命令是否有 nextDynamicChildren（用于快速重启/停止等二级选择）
      if (parentCmd && parentCmd.nextDynamicChildren) {
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
        // 有动作，执行动作
        executeAction(cmd)
        return
      }
      
      // 选择的是目标（项目/环境等），检查父命令是否有 action
      if (cmd.data && parentCmd && parentCmd.action) {
        if (parentCmd.needInput) {
          showCommands.value = false
          return
        }
        executeAction(parentCmd)
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
          executeAction(nearestAction)
          return
        }
      }
      
      // 没有可执行的操作，提示用户
      messages.value.push({
        type: 'system',
        content: `命令 "${cmd.name}" 暂不支持快捷操作\n`
      })
      inputText.value = ''
      showCommands.value = false
      commandStack.value = []
      currentChildren.value = []
      currentInputValue.value = ''
      scrollToBottom()
    }

    // 执行命令
    const executeCommand = () => {
      if (!inputText.value.trim()) return

      if (isCommandModeByText(inputText.value)) {
        parseInput()
      }

      // 如果有命令栈，执行最后一个命令
      if (commandStack.value.length > 0) {
        const actionCmd = commandStack.value.find(item => item.action)
        if (actionCmd) {
          const actionIndex = commandStack.value.findIndex(item => item.action)
          const targetCmd = actionCmd.needTarget ? commandStack.value[actionIndex + 1] : null
          if (actionCmd.needTarget && !(targetCmd && targetCmd.data)) {
            messages.value.push({
              type: 'system',
              content: '命令未完成：请先选择项目/环境\n'
            })
            scrollToBottom()
            return
          }
          if (actionCmd.needInput) {
            const branchName = normalizeCommandPart(currentInputValue.value)
            if (!branchName) {
              messages.value.push({
                type: 'system',
                content: `命令未完成：${actionCmd.inputPlaceholder || '请输入参数'}\n`
              })
              scrollToBottom()
              return
            }
          }
          executeAction(actionCmd, { inputValue: currentInputValue.value })
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
      const outputMsg = {
        type: 'system',
        commandText: `执行操作: ${cmd.name}`,
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
        case 'gitPull':
          executeGitAction('pull', currentStack, options.inputValue || '')
          break
        case 'gitStatus':
          executeGitAction('status', currentStack, options.inputValue || '')
          break
        case 'gitBranch':
          executeGitAction('branch', currentStack, options.inputValue || '')
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
        case 'gitSaveCredentials':
          executeGitAction('saveCredentials', currentStack, options.inputValue || '')
          break
        case 'gitSetSafe':
          executeGitAction('setSafe', currentStack, options.inputValue || '')
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
      
      // 处理 HTTP 响应的回调
      const callback = (response) => {
        if (response.ErrCode !== 0) {
          appendOutputResult(`错误: ${response.ErrMsg || '未知错误'}\n`)
        } else if (response.Data) {
          // 显示返回的数据
          if (typeof response.Data === 'string') {
            appendOutputResult(response.Data)
          } else {
            appendOutputResult(`${JSON.stringify(response.Data, null, 2)}\n`)
          }
        } else {
          appendOutputResult('执行成功\n')
        }
        setTimeout(() => {
          // 给 SSE 尾包一点时间，避免过程/结果末尾被截断
          sseDistribute.UnRegisterReceive(newSseDistributeId)
          finishExecution()
        }, 1200)
      }
      
      switch (action) {
        case 'pull':
          appendOutputResult('正在拉取代码...\n\n')
          git.GitPullBranchOrigin(gitConfig, callback)
          break
        case 'status':
          appendOutputResult('正在查询状态...\n\n')
          git.GitQueryStatus(gitConfig, callback)
          break
        case 'branch':
          appendOutputResult('正在查询分支...\n\n')
          git.GitCurrentBranch(gitConfig, callback)
          break
        case 'log':
          appendOutputResult('正在查询日志...\n\n')
          git.GitCommitLog(gitConfig, callback)
          break
        case 'checkout':
          {
            // 需要分支名
            const branchName = normalizeCommandPart(inputValue)
            if (!branchName) {
              appendOutputResult('错误：请输入分支名\n')
              finishExecution()
              return
            }
            appendOutputResult(`正在切换到分支 ${branchName}...\n\n`)
            git.GitChangeBranch(gitConfig, branchName, callback)
          }
          break
        case 'checkoutRemote':
          {
            const branchNameRemote = normalizeCommandPart(inputValue)
            if (!branchNameRemote) {
              appendOutputResult('错误：请输入远程分支名\n')
              finishExecution()
              return
            }
            appendOutputResult(`正在关联并切换远程分支 ${branchNameRemote}...\n\n`)
            git.GitChangeBranchRemote(gitConfig, branchNameRemote, callback)
          }
          break
        case 'saveCredentials':
          appendOutputResult('正在保存账号密码配置...\n\n')
          git.GitSaveCredentials(gitConfig, callback)
          break
        case 'setSafe':
          appendOutputResult('正在设置目录安全...\n\n')
          git.SetSafe(gitConfig, callback)
          break
        default:
          sseDistribute.UnRegisterReceive(newSseDistributeId)
          finishExecution()
      }
    }
    
    // 完成执行
    const finishExecution = () => {
      isExecuting.value = false
      if (currentOutputMessage.value) {
        appendOutputResult('\n[完成]\n')
      }
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

    onMounted(() => {
      inputRef.value?.focus()
      initSseConnection()
    })
    
    onUnmounted(() => {
      // 只取消注册回调，不关闭 SSE 连接（其他页面可能还在使用）
      sseDistribute.UnRegisterReceive(sseDistributeId.value)
    })

    return {
      inputText,
      messages,
      showCommands,
      filteredCommands,
      activeCommandIndex,
      inputRef,
      messageList,
      commandBreadcrumb,
      inputPlaceholder,
      handleInput,
      handleKeydown,
      handleFocus,
      handleBlur,
      selectCommand,
      executeCommand,
      getCommandKey,
      hasCommandLayout,
    }
  }
}
</script>

<style scoped>
.dashboard-container {
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  padding: 20px;
  background: #fafaf7;
}

.chat-container {
  width: 100%;
  max-width: 800px;
  height: 70vh;
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

.message {
  max-width: 80%;
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
}

.message-command {
  font-size: 13px;
  color: #5a8a5a;
  margin-bottom: 8px;
  padding: 0 4px;
}

.message-content {
  padding: 12px 16px;
  border-radius: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-word;
}

.process-window {
  margin-top: 8px;
  border: 1px solid #d9d9cf;
  border-radius: 10px;
  background: #1f1f1b;
  color: #d8e0d2;
  overflow: hidden;
}

.process-title {
  font-size: 12px;
  color: #cdd5c8;
  background: #2a2a25;
  padding: 6px 10px;
  border-bottom: 1px solid #3a3a34;
}

.process-text {
  margin: 0;
  padding: 10px 12px;
  max-height: 240px;
  overflow-y: auto;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 12px;
  line-height: 1.45;
  font-family: Consolas, Monaco, 'Courier New', monospace;
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

.command-dropdown {
  position: absolute;
  bottom: 80px;
  left: 24px;
  right: 24px;
  background: #fff;
  border: 1px solid #e0e0d8;
  border-radius: 10px;
  max-height: 300px;
  overflow-y: auto;
  z-index: 100;
  box-shadow: 0 4px 16px rgba(0, 0, 0, 0.1);
}

.command-item {
  display: flex;
  align-items: center;
  padding: 12px 16px;
  cursor: pointer;
  transition: background 0.15s;
  border-bottom: 1px solid #f0f0e8;
}

.command-item:last-child {
  border-bottom: none;
}

.command-item:hover,
.command-item.active {
  background: #f5f8f5;
}

.command-icon {
  font-size: 18px;
  margin-right: 12px;
  width: 24px;
  text-align: center;
}

.command-name {
  font-weight: 500;
  color: #4a4a4a;
  margin-right: 12px;
  min-width: 80px;
}

.command-desc {
  color: #8a8a7a;
  font-size: 13px;
  flex: 1;
}

.command-arrow {
  color: #c0c0b8;
  font-size: 14px;
  margin-left: 8px;
}

.command-breadcrumb {
  padding: 10px 16px;
  background: #f5f8f5;
  border-bottom: 1px solid #e8e8e0;
  border-radius: 10px 10px 0 0;
}

.breadcrumb-text {
  font-size: 13px;
  color: #5a8a5a;
  font-weight: 500;
}

.input-container {
  padding: 16px 24px;
  border-top: 1px solid #e8e8e0;
  background: #fff;
  border-radius: 0 0 12px 12px;
}

.input-wrapper {
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

.chat-input {
  flex: 1;
  background: transparent;
  border: none;
  padding: 12px 16px;
  font-size: 15px;
  color: #4a4a4a;
  outline: none;
}

.chat-input::placeholder {
  color: #a0a090;
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
