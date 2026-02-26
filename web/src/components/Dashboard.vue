<template>
  <div class="dashboard-container">
    <div class="chat-container">
      <!-- 消息列表区域 -->
      <div ref="messageList" class="message-list">
        <div class="welcome-message">
          <h2>开发者工具平台</h2>
          <p class="hint">输入 <kbd>/</kbd> 快速访问功能</p>
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
        <div
          v-for="(cmd, index) in filteredCommands"
          :key="cmd.path"
          :class="['command-item', { active: activeCommandIndex === index }]"
          @click="selectCommand(cmd)"
          @mouseenter="activeCommandIndex = index"
        >
          <span class="command-icon">{{ cmd.icon }}</span>
          <span class="command-name">{{ cmd.name }}</span>
          <span class="command-desc">{{ cmd.desc }}</span>
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
            placeholder="输入 / 快速访问功能..."
            @input="handleInput"
            @keydown="handleKeydown"
            @blur="handleBlur"
            @focus="handleFocus"
          />
          <button class="send-btn" @click="sendMessage">
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

    // 所有可用命令
    const allCommands = [
      { path: '/Redis', name: 'Redis', icon: '🗃️', desc: 'Redis管理', module: 'redis' },
      { path: '/Supervisor', name: 'Supervisor', icon: '⚙️', desc: '进程管理', module: 'supervisor' },
      { path: '/Git', name: 'Git', icon: '📚', desc: 'Git管理', module: 'git' },
      { path: '/Link', name: '自定义网页', icon: '🔗', desc: '自定义网页链接', module: 'login' },
      { path: '/Variable', name: '自定义脚本', icon: '📝', desc: '自定义脚本管理', module: 'variable' },
      { path: '/Docker', name: 'Docker', icon: '🐳', desc: 'Docker容器管理', module: 'docker' },
      { path: '/Api', name: '接口开发', icon: '🔌', desc: 'API接口开发', module: 'api' },
      { path: '/shellout', name: '终端输出', icon: '💻', desc: '终端输出查看', module: 'shellout' },
      { path: '/Set', name: '配置', icon: '🔧', desc: '系统配置', module: null },
    ]

    // 开放的模块列表
    const openModules = module.GetOpenModuleList()

    // 根据模块配置过滤可用命令
    const availableCommands = computed(() => {
      return allCommands.filter(cmd => {
        if (cmd.module === null) return true
        return openModules.includes(cmd.module)
      })
    })

    // 过滤后的命令列表
    const filteredCommands = computed(() => {
      if (!inputText.value.startsWith('/')) {
        return availableCommands.value
      }
      const searchText = inputText.value.slice(1).toLowerCase()
      if (!searchText) {
        return availableCommands.value
      }
      return availableCommands.value.filter(cmd =>
        cmd.name.toLowerCase().includes(searchText) ||
        cmd.desc.toLowerCase().includes(searchText)
      )
    })

    // 处理输入
    const handleInput = () => {
      if (inputText.value.startsWith('/')) {
        showCommands.value = true
        activeCommandIndex.value = 0
      } else {
        showCommands.value = false
      }
    }

    // 处理焦点
    const handleFocus = () => {
      if (inputText.value.startsWith('/')) {
        showCommands.value = true
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
          sendMessage()
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
        case 'Enter':
          e.preventDefault()
          if (filteredCommands.value[activeCommandIndex.value]) {
            selectCommand(filteredCommands.value[activeCommandIndex.value])
          }
          break
        case 'Escape':
          showCommands.value = false
          break
      }
    }

    // 选择命令
    const selectCommand = (cmd) => {
      inputText.value = ''
      showCommands.value = false
      router.push(cmd.path)
    }

    // 发送消息
    const sendMessage = () => {
      if (!inputText.value.trim()) return

      messages.value.push({
        type: 'user',
        content: inputText.value
      })

      // 模拟响应
      setTimeout(() => {
        messages.value.push({
          type: 'system',
          content: `已收到命令: ${inputText.value}`
        })
        scrollToBottom()
      }, 300)

      inputText.value = ''
      showCommands.value = false
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
      handleInput,
      handleKeydown,
      handleFocus,
      handleBlur,
      selectCommand,
      sendMessage,
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
