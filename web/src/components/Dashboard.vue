<template>
  <div class="dashboard-container">
    <div class="chat-container">
      <!-- 消息列表区域 -->
      <div ref="messageList" class="message-list">
        <div class="welcome-message">
          <h2>开发者工具平台</h2>
          <p class="hint">输入 <kbd>/</kbd> 快速访问功能，<kbd>Tab</kbd> 补全，<kbd>Space</kbd> 继续</p>
        </div>
        <div
          v-for="(msg, index) in messages"
          :key="index"
          :class="['message', msg.type]"
        >
          <div class="message-content">{{ msg.content }}</div>
        </div>
      </div>

      <!-- 命令提示下拉框 -->
      <div v-show="showCommands" class="command-dropdown">
        <div class="command-breadcrumb" v-if="commandBreadcrumb">
          <span class="breadcrumb-text">{{ commandBreadcrumb }}</span>
        </div>
        <div
          v-for="(cmd, index) in filteredCommands"
          :key="cmd.command || cmd.path"
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
import { ref, computed, nextTick, onMounted } from 'vue'
import { useRouter } from 'vue-router'
import module from '@/utils/module'
import commandConfig from '@/config/commandConfig.js'
import ssh from '@/utils/base/ssh_set'
import git from '@/utils/base/git'
import compose from '@/utils/base/compose'
import supervisor from '@/utils/base/supervisor'
import shellOut from '@/utils/base/shell_out'
import store from '@/utils/base/store'

export default {
  name: 'Dashboard',
  setup() {
    const router = useRouter()
    const inputText = ref('')
    const messages = ref([])
    const showCommands = ref(false)
    const activeCommandIndex = ref(0)
    const inputRef = ref(null)
    const messageList = ref(null)
    
    // 多级命令状态
    const commandStack = ref([]) // 命令栈，存储已选择的命令
    const currentChildren = ref([]) // 当前可选的子命令
    const selectedTarget = ref(null) // 已选择的目标
    const dynamicDataCache = ref({}) // 动态数据缓存
    const isLoadingDynamic = ref(false) // 是否正在加载动态数据

    // 开放的模块列表
    const openModules = module.GetOpenModuleList()

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
        return '输入 / 快速访问功能，Tab 补全，Space 继续...'
      }
      const lastCmd = commandStack.value[commandStack.value.length - 1]
      if (lastCmd.needInput) {
        return lastCmd.inputPlaceholder || '请输入...'
      }
      if (lastCmd.needTarget && !selectedTarget.value) {
        return '选择目标...'
      }
      return '继续输入或选择...'
    })

    // 过滤后的命令列表
    const filteredCommands = computed(() => {
      let commands = currentChildren.value.length > 0 
        ? currentChildren.value 
        : availableCommands.value
      
      // 获取当前输入的搜索文本
      const parts = inputText.value.split(' ')
      const searchText = parts[parts.length - 1].toLowerCase().replace('/', '')
      
      if (!searchText) {
        return commands
      }
      
      return commands.filter(cmd =>
        cmd.name.toLowerCase().includes(searchText) ||
        cmd.command?.toLowerCase().includes(searchText) ||
        cmd.desc?.toLowerCase().includes(searchText)
      )
    })

    // 解析输入文本，获取当前命令层级
    const parseInput = () => {
      if (!inputText.value.startsWith('/')) {
        commandStack.value = []
        currentChildren.value = []
        selectedTarget.value = null
        return
      }

      const parts = inputText.value.slice(1).split(' ').filter(p => p)
      
      // 重置状态
      commandStack.value = []
      currentChildren.value = []
      selectedTarget.value = null
      
      let currentLevel = availableCommands.value
      
      for (let i = 0; i < parts.length; i++) {
        const part = parts[i].toLowerCase()
        const found = currentLevel.find(cmd => 
          cmd.command?.toLowerCase() === part ||
          cmd.name?.toLowerCase() === part
        )
        
        if (found) {
          commandStack.value.push(found)
          
          // 如果有子命令，继续
          if (found.children && found.children.length > 0) {
            currentLevel = found.children
            currentChildren.value = found.children
          } 
          // 如果需要动态子命令
          else if (found.dynamicChildren) {
            loadDynamicChildren(found.dynamicChildren, found)
            break
          }
          // 如果需要选择目标
          else if (found.needTarget && !selectedTarget.value) {
            // 目标选择模式，等待选择
            break
          }
          // 如果需要输入
          else if (found.needInput) {
            break
          }
          else {
            currentChildren.value = []
            break
          }
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
          }
          break
        }
      }
    }

    // 加载动态子命令
    const loadDynamicChildren = (type, parentCmd) => {
      if (dynamicDataCache.value[type]) {
        currentChildren.value = dynamicDataCache.value[type]
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
          const list = response.Data.git_list.map(item => ({
            command: item.name,
            name: item.name,
            desc: item.path || '',
            id: item.id,
            data: item
          }))
          dynamicDataCache.value['gitProjectList'] = list
          currentChildren.value = list
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
      if (inputText.value.startsWith('/')) {
        showCommands.value = true
        activeCommandIndex.value = 0
        parseInput()
      } else {
        showCommands.value = false
        commandStack.value = []
        currentChildren.value = []
      }
    }

    // 处理焦点
    const handleFocus = () => {
      if (inputText.value.startsWith('/')) {
        showCommands.value = true
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
          // 如果输入为空且有命令栈，退回上一级
          const parts = inputText.value.split(' ')
          if (parts[parts.length - 1] === '' && commandStack.value.length > 0) {
            e.preventDefault()
            goBackCommand()
          }
          break
      }
    }

    // 退回上一级命令
    const goBackCommand = () => {
      if (commandStack.value.length === 0) return
      
      commandStack.value.pop()
      selectedTarget.value = null
      
      // 重新构建输入文本
      const prefix = '/' + commandStack.value.map(c => c.command).join(' ')
      if (prefix.length > 1) {
        inputText.value = prefix + ' '
      } else {
        inputText.value = '/'
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
      inputText.value = '/' + commandStack.value.map(c => c.command || c.name).join(' ') + ' '
      
      // 检查父命令是否有 nextDynamicChildren（用于快速重启/停止等二级选择）
      if (parentCmd && parentCmd.nextDynamicChildren) {
        // 加载下一级动态数据
        loadDynamicChildren(parentCmd.nextDynamicChildren, cmd)
        activeCommandIndex.value = 0
        return
      }
      
      // 检查是否需要继续
      if (cmd.children && cmd.children.length > 0) {
        currentChildren.value = cmd.children
        activeCommandIndex.value = 0
      } else if (cmd.dynamicChildren) {
        loadDynamicChildren(cmd.dynamicChildren, cmd)
        activeCommandIndex.value = 0
      } else if (cmd.needTarget) {
        // 需要选择目标，保持下拉框打开
        activeCommandIndex.value = 0
      } else if (cmd.needInput) {
        // 需要输入，等待用户输入
        showCommands.value = false
      } else if (cmd.path) {
        // 有路径，直接跳转
        inputText.value = ''
        showCommands.value = false
        commandStack.value = []
        currentChildren.value = []
        router.push(cmd.path)
      } else if (cmd.action) {
        // 有动作，执行动作
        executeAction(cmd)
      } else {
        // 默认行为：如果有父命令的路径，跳转
        const grandParentCmd = commandStack.value.length > 2 
          ? commandStack.value[commandStack.value.length - 3] 
          : null
        if (grandParentCmd && grandParentCmd.path) {
          inputText.value = ''
          showCommands.value = false
          commandStack.value = []
          currentChildren.value = []
          router.push(grandParentCmd.path)
        }
      }
    }

    // 执行命令
    const executeCommand = () => {
      if (!inputText.value.trim()) return

      // 如果有命令栈，执行最后一个命令
      if (commandStack.value.length > 0) {
        const lastCmd = commandStack.value[commandStack.value.length - 1]
        if (lastCmd.action) {
          executeAction(lastCmd)
          return
        }
        if (lastCmd.path) {
          router.push(lastCmd.path)
          inputText.value = ''
          showCommands.value = false
          commandStack.value = []
          currentChildren.value = []
          return
        }
      }

      // 普通消息
      messages.value.push({
        type: 'user',
        content: inputText.value
      })

      setTimeout(() => {
        messages.value.push({
          type: 'system',
          content: `已收到命令: ${inputText.value}`
        })
        scrollToBottom()
      }, 300)

      inputText.value = ''
      showCommands.value = false
      commandStack.value = []
      currentChildren.value = []
      scrollToBottom()
    }

    // 执行动作
    const executeAction = (cmd) => {
      messages.value.push({
        type: 'system',
        content: `执行操作: ${cmd.name}`
      })
      
      // 这里可以根据 cmd.action 执行具体操作
      // 目前简化处理，跳转到对应页面
      const parentCmd = commandStack.value.find(c => c.path)
      if (parentCmd) {
        router.push(parentCmd.path)
      }
      
      inputText.value = ''
      showCommands.value = false
      commandStack.value = []
      currentChildren.value = []
      scrollToBottom()
    }

    // 滚动到底部
    const scrollToBottom = () => {
      nextTick(() => {
        if (messageList.value) {
          messageList.value.scrollTop = messageList.value.scrollHeight
        }
      })
    }

    onMounted(() => {
      inputRef.value?.focus()
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

.message-content {
  padding: 12px 16px;
  border-radius: 12px;
  line-height: 1.5;
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
