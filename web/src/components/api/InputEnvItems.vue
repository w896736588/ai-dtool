<template>
  <div class="variable-input-container">
    <el-input
        ref="inputRef"
        v-model="inputValue"
        :placeholder="placeholder"
        @input="handleInput"
        @focus="handleFocus"
        @blur="handleBlur"
        @keydown="handleKeydown"
        @click="handleClick"
    >
      <template #prefix>
        <el-icon><Search /></el-icon>
      </template>
    </el-input>

    <!-- 下拉建议列表 -->
    <el-popover
        v-model:visible="suggestionsVisible"
        placement="bottom-start"
        :width="300"
        popper-class="variable-suggestions-popper"
        trigger="manual"
    >
      <template #reference>
        <div></div>
      </template>

      <div class="suggestions-list">
        <div
            v-for="(item, index) in filteredSuggestions"
            :key="item.id"
            class="suggestion-item"
            :class="{ 'active': activeIndex === index }"
            @click="selectSuggestion(item)"
            @mouseenter="activeIndex = index"
        >
          <div class="suggestion-label">{{ item.label }}</div>
          <div class="suggestion-value">{{ item.value }}</div>
        </div>
        <div v-if="filteredSuggestions.length === 0" class="no-suggestions">
          暂无匹配项
        </div>
      </div>
    </el-popover>

    <!-- 变量显示区域 -->
    <div class="variables-display" v-if="variables.length > 0">
      <span
          v-for="(variable, index) in variables"
          :key="index"
          class="variable-tag"
          :data-tooltip="variable.value"
      >
        {Url}
        <el-icon class="remove-variable" @click="removeVariable(index)"><Close /></el-icon>
      </span>
    </div>
  </div>
</template>

<script>
import { ref, reactive, toRefs, nextTick } from 'vue'
import { Search, Close } from '@element-plus/icons-vue'

export default {
  name: 'VariableInput',
  components: {
    Search,
    Close
  },
  props: {
    modelValue: {
      type: String,
      default: ''
    },
    placeholder: {
      type: String,
      default: '输入内容，/ 开始选择变量'
    }
  },
  emits: ['update:modelValue'],
  setup(props, { emit }) {
    const inputRef = ref(null)
    const suggestionsVisible = ref(false)
    const allSuggestions = ref([])
    const filteredSuggestions = ref([])
    const activeIndex = ref(-1)
    const lastSlashIndex = ref(-1)
    const inputValue = ref(props.modelValue)
    const variables = ref([])

    // 模拟API数据
    const fetchSuggestions = async () => {
      try {
        // 这里模拟调用 /api/ent/items 接口
        const response = await new Promise(resolve => {
          setTimeout(() => {
            resolve([
              { id: 1, label: '用户ID', value: 'user_id' },
              { id: 2, label: '订单号', value: 'order_no' },
              { id: 3, label: '商品名称', value: 'product_name' },
              { id: 4, label: '用户邮箱', value: 'user_email' },
              { id: 5, label: '订单金额', value: 'order_amount' }
            ])
          }, 300)
        })
        allSuggestions.value = response
        filteredSuggestions.value = response
      } catch (error) {
        console.error('获取建议失败:', error)
        allSuggestions.value = []
        filteredSuggestions.value = []
      }
    }

    // 处理输入事件
    const handleInput = (value) => {
      inputValue.value = value
      emit('update:modelValue', value)

      // 检查是否输入了 /
      const lastSlash = value.lastIndexOf('/')
      if (lastSlash !== -1) {
        lastSlashIndex.value = lastSlash
        suggestionsVisible.value = true
        const query = value.substring(lastSlash + 1)
        filterSuggestions(query)
      } else {
        suggestionsVisible.value = false
      }
    }

    // 过滤建议
    const filterSuggestions = (query) => {
      if (!query) {
        filteredSuggestions.value = allSuggestions.value
      } else {
        filteredSuggestions.value = allSuggestions.value.filter(item =>
            item.label.toLowerCase().includes(query.toLowerCase()) ||
            item.value.toLowerCase().includes(query.toLowerCase())
        )
      }
      activeIndex.value = 0
    }

    // 选择建议项
    const selectSuggestion = (item) => {
      if (lastSlashIndex.value !== -1) {
        // 替换从最后一个 / 开始的部分
        const prefix = inputValue.value.substring(0, lastSlashIndex.value)
        inputValue.value = prefix + `{Url}`
        emit('update:modelValue', inputValue.value)

        // 添加到变量列表
        variables.value.push({
          label: item.label,
          value: item.value
        })
      }
      suggestionsVisible.value = false
      activeIndex.value = -1
    }

    // 处理键盘事件
    const handleKeydown = (e) => {
      if (!suggestionsVisible.value) return

      switch (e.key) {
        case 'ArrowDown':
          e.preventDefault()
          activeIndex.value = Math.min(activeIndex.value + 1, filteredSuggestions.value.length - 1)
          break
        case 'ArrowUp':
          e.preventDefault()
          activeIndex.value = Math.max(activeIndex.value - 1, 0)
          break
        case 'Enter':
          e.preventDefault()
          if (activeIndex.value >= 0 && filteredSuggestions.value[activeIndex.value]) {
            selectSuggestion(filteredSuggestions.value[activeIndex.value])
          }
          break
        case 'Escape':
          suggestionsVisible.value = false
          activeIndex.value = -1
          break
      }
    }

    // 处理焦点事件
    const handleFocus = () => {
      // 如果输入框内容以 / 结尾，显示建议
      if (inputValue.value.endsWith('/')) {
        suggestionsVisible.value = true
        fetchSuggestions()
      }
    }

    // 处理失焦事件
    const handleBlur = () => {
      // 延迟隐藏建议，避免点击建议项时立即消失
      setTimeout(() => {
        suggestionsVisible.value = false
      }, 200)
    }

    // 处理点击事件
    const handleClick = () => {
      if (inputValue.value.endsWith('/')) {
        suggestionsVisible.value = true
        fetchSuggestions()
      }
    }

    // 移除变量
    const removeVariable = (index) => {
      variables.value.splice(index, 1)
    }

    // 初始化建议数据
    fetchSuggestions()

    return {
      inputRef,
      suggestionsVisible,
      filteredSuggestions,
      activeIndex,
      inputValue,
      variables,
      handleInput,
      handleKeydown,
      handleFocus,
      handleBlur,
      handleClick,
      selectSuggestion,
      removeVariable
    }
  }
}
</script>

<style scoped>
.variable-input-container {
  position: relative;
  width: 100%;
}

.variables-display {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  margin-top: 8px;
}

.variable-tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 8px;
  background-color: #f0f0f0; /* 淡灰色背景 */
  color: #67c23a; /* 绿色文字 */
  border-radius: 4px;
  font-size: 12px;
  cursor: default;
  position: relative;
}

.variable-tag[data-tooltip]:hover::after {
  content: attr(data-tooltip);
  position: absolute;
  bottom: 100%;
  left: 50%;
  transform: translateX(-50%);
  background-color: #333;
  color: white;
  padding: 4px 8px;
  border-radius: 4px;
  font-size: 12px;
  white-space: nowrap;
  z-index: 1000;
  margin-bottom: 4px;
}

.remove-variable {
  margin-left: 6px;
  cursor: pointer;
  font-size: 12px;
}

.remove-variable:hover {
  color: #f56c6c;
}

.suggestions-list {
  max-height: 200px;
  overflow-y: auto;
}

.suggestion-item {
  padding: 8px 12px;
  cursor: pointer;
  border-bottom: 1px solid #f0f0f0;
}

.suggestion-item:last-child {
  border-bottom: none;
}

.suggestion-item:hover,
.suggestion-item.active {
  background-color: #f5f7fa;
}

.suggestion-label {
  font-weight: 500;
  color: #303133;
}

.suggestion-value {
  font-size: 12px;
  color: #909399;
  margin-top: 2px;
}

.no-suggestions {
  padding: 8px 12px;
  color: #909399;
  font-size: 14px;
  text-align: center;
}

:deep(.variable-suggestions-popper) {
  padding: 0 !important;
}
</style>