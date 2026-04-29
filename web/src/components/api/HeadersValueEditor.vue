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
        <pl-button
            type="danger"
            plain
            size="small"
            class="delete-rule-btn"
            @click="removeItem(index)"
        >删除
        </pl-button>
      </div>
    </div>

    <div class="footer-actions">
      <pl-button type="primary" plain size="small" class="add-rule-btn" @click="addItem">
        添加参数
      </pl-button>

      <pl-button plain size="small" class="bulk-edit-btn" @click="handleBulkEdit">
        批量编辑
      </pl-button>
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
        <pl-button @click="bulkEditVisible = false">取消</pl-button>
        <pl-button type="primary" @click="applyBulkEdit">应用</pl-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { nextTick } from 'vue'

export default {
  name: 'KeyValueEditor',
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

<style scoped src="@/css/components/api/HeadersValueEditor.css"></style>

