<template>
  <div class="key-value-editor">
    <div class="header-row">
      <span class="header-key">键</span>
      <span class="header-value">值</span>
      <span class="header-actions">操作</span>
    </div>

    <div
        v-for="(item, index) in localData"
        :key="item.id"
        class="data-row"
    >
      <el-autocomplete
          v-model="item.key"
          :fetch-suggestions="queryKeySuggestions"
          placeholder="键名"
          class="key-input"
          @select="handleKeySelect"
          @blur="handleDataChange"
      />

      <el-input
          v-model="item.value"
          placeholder="值"
          class="value-input"
          @blur="handleDataChange"
      />

      <div class="actions">
        <el-button
            type="danger"
            link
            size="small"
            @click="removeItem(index)"
        >删除
        </el-button>
      </div>
    </div>

    <div class="footer-actions">
      <el-button type="primary" link @click="addItem">
        <el-icon><Plus /></el-icon>
        添加参数
      </el-button>

      <el-button link @click="handleBulkEdit">
        <el-icon><Edit /></el-icon>
        批量编辑
      </el-button>
    </div>

    <!-- 批量编辑对话框 -->
    <el-dialog
        v-model="bulkEditVisible"
        title="批量编辑"
        width="600px"
    >
      <el-input
          v-model="bulkEditText"
          type="textarea"
          :rows="10"
          placeholder="每行一个参数，格式：键=值&#10;例如：&#10;Content-Type=application/json&#10;Authorization=Bearer token"
      />
      <template #footer>
        <el-button @click="bulkEditVisible = false">取消</el-button>
        <el-button type="primary" @click="applyBulkEdit">应用</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { Delete, Plus, Edit } from '@element-plus/icons-vue'
import { nextTick } from 'vue'

export default {
  name: 'KeyValueEditor',
  components: {
    Delete,
    Plus,
    Edit
  },
  props: {
    modelValue: {
      type: Object,
      default: () => ({})
    },
    keys: {
      type: Array,
      default: () => []
    }
  },
  data() {
    return {
      localData: [],
      bulkEditVisible: false,
      bulkEditText: '',
      nextId: 1
    }
  },
  watch: {
    modelValue: {
      handler(newVal) {
        this.updateLocalData(newVal)
      },
      immediate: true,
      deep: true
    }
  },
  methods: {
    updateLocalData(sourceData) {
      if (!sourceData || Object.keys(sourceData).length === 0) {
        this.localData = [{ id: this.nextId++, key: '', value: ''}]
        return
      }

      this.localData = Object.entries(sourceData).map(([key, value]) => ({
        id: this.nextId++,
        key,
        value: typeof value === 'string' ? value : JSON.stringify(value),
        description: ''
      }))

      // 确保至少有一个空行
      if (this.localData.length === 0) {
        this.localData.push({ id: this.nextId++, key: '', value: ''})
      }
    },

    emitUpdate() {
      const result = {}
      this.localData.forEach(item => {
          result[item.key.trim()] = item.value
      })
      this.$emit('update', result)
    },

    handleDataChange() {
      this.emitUpdate()
    },

    addItem() {
      this.localData.push({ id: this.nextId++, key: '', value: ''})
      // 强制更新视图
      this.emitUpdate()
    },

    removeItem(index) {
      this.localData.splice(index, 1)
      // 如果删除了所有行，添加一个空行
      if (this.localData.length === 0) {
        this.addItem()
      }
      this.emitUpdate()
    },

    queryKeySuggestions(queryString, cb) {
      const suggestions = this.keys.map(key => ({
        value: key,
        label: key
      }))

      const results = queryString
          ? suggestions.filter(item =>
              item.value.toLowerCase().includes(queryString.toLowerCase()))
          : suggestions

      cb(results)
    },

    handleKeySelect(selected) {
      this.handleDataChange()
    },

    handleBulkEdit() {
      this.bulkEditText = this.localData
          .filter(item => item.key.trim())
          .map(item => `${item.key}=${item.value}`)
          .join('\n')
      this.bulkEditVisible = true
    },

    applyBulkEdit() {
      const lines = this.bulkEditText.split('\n').filter(line => line.trim())
      const newData = []

      lines.forEach(line => {
        const [key, ...valueParts] = line.split('=')
        if (key && key.trim()) {
          newData.push({
            id: this.nextId++,
            key: key.trim(),
            value: valueParts.join('='), // 处理值中包含等号的情况
            description: ''
          })
        }
      })

      this.localData = newData.length > 0 ? newData : [{ id: this.nextId++, key: '', value: '', description: '' }]
      this.emitUpdate()
      this.bulkEditVisible = false
      this.$message.success(`成功导入 ${newData.length} 个参数`)
    }
  }
}
</script>

<style scoped>
.key-value-editor {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: #fff;
}

.header-row {
  display: grid;
  grid-template-columns: 2fr 3fr 80px;
  align-items: center;
  padding: 8px 12px;
  background: #f5f7fa;
  border-bottom: 1px solid #e4e7ed;
  font-weight: 600;
  font-size: 14px;
  color: #606266;
  gap: 8px;
}

.data-row {
  display: grid;
  grid-template-columns: 2fr 3fr 80px;
  align-items: center;
  padding: 8px 12px;
  border-bottom: 1px solid #f0f0f0;
  gap: 8px;
}

.data-row:last-child {
  border-bottom: none;
}

.data-row:hover {
  background: #f8f9fa;
}

.key-input,
.value-input,
.actions {
  width: 100%;
}

.actions {
  text-align: center;
}

.footer-actions {
  padding: 12px;
  border-top: 1px solid #e4e7ed;
  background: #fafafa;
}

.footer-actions .el-button {
  margin-right: 16px;
}

/* 确保子组件宽度正确 */
.key-value-editor :deep(.el-autocomplete),
.key-value-editor :deep(.el-input) {
  width: 100%;
}
</style>