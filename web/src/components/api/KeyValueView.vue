<template>
  <div class="key-value-view">
    <div class="view-header">
      <div class="header-actions">
        <el-button type="primary" link @click="copyToClipboard">
          <el-icon><DocumentCopy /></el-icon>
          复制全部
        </el-button>
      </div>
    </div>

    <div class="data-list">
      <el-table
          :data="tableDataWithEmpty"
          :show-header="true"
          :row-class-name="tableRowClassName"
      >
        <el-table-column prop="key" label="键" width="200">
        </el-table-column>

        <el-table-column prop="value" label="值" min-width="300">
        </el-table-column>
      </el-table>
    </div>
  </div>
</template>

<script>
import { DocumentCopy } from '@element-plus/icons-vue'

export default {
  name: 'KeyValueView',
  components: {
    DocumentCopy
  },
  props: {
    data: {
      type: Object,
      default: () => ({})
    },
    maxPreviewLength: {
      type: Number,
      default: 150
    }
  },
  data() {
    return {
      expandedKeys: {}
    }
  },
  computed: {
    tableData() {
      return Object.keys(this.data || {}).map(key => ({
        key,
        value: this.formatValuePreview(this.data[key])
      }))
    },
    tableDataWithEmpty() {
      if (Object.keys(this.data || {}).length === 0) {
        return [{
          key: '暂无数据',
          value: '暂无数据'
        }]
      }
      return this.tableData
    }
  },
  watch: {
    data: {
      handler() {
        // 重置展开状态
        this.expandedKeys = {}
      },
      deep: true
    }
  },
  methods: {
    formatValuePreview(value) {
      if (value === null || value === undefined) {
        return 'null'
      }

      // 如果是对象或数组，显示类型和长度
      if (typeof value === 'object' && Object.keys(value).length > 0) {
        const type = Array.isArray(value) ? 'Array' : 'Object'
        const length = Array.isArray(value) ? value.length : Object.keys(value).length
        return `${type}(${length})`
      }

      const str = String(value)
      if (str.length > this.maxPreviewLength) {
        return str.substring(0, this.maxPreviewLength) + '...'
      }

      return str
    },

    tableRowClassName({ row, rowIndex }) {
      if (row.key === '暂无数据' && row.value === '暂无数据') {
        return 'empty-row'
      }
      return 'key-value-row'
    },

    copyToClipboard() {
      const text = JSON.stringify(this.data, null, 2)
      this.copyTextToClipboard(text)
      this.$message.success('已复制到剪贴板')
    },

    copyTextToClipboard(text) {
      // 创建临时文本区域
      const textArea = document.createElement('textarea')
      textArea.value = text
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()

      try {
        document.execCommand('copy')
      } catch (err) {
        console.error('复制失败:', err)
      }

      document.body.removeChild(textArea)

      // 使用 Clipboard API（如果可用）
      if (navigator.clipboard && window.isSecureContext) {
        navigator.clipboard.writeText(text).catch(err => {
          console.error('Clipboard API 复制失败:', err)
        })
      }
    }
  }
}
</script>

<style scoped>
.key-value-view {
  border: 1px solid #e4e7ed;
  border-radius: 4px;
  background: #fff;
  font-size: 14px;
}

.view-header {
  padding: 8px 12px;
  border-bottom: 1px solid #e4e7ed;
  background: #f5f7fa;
}

.header-actions {
  display: flex;
  justify-content: flex-end;
}

.data-list {
  max-height: 400px;
  overflow-y: auto;
}

.key-value-table {
  border-collapse: collapse;
  width: 100%;
}

.key-value-table :deep(.el-table__header) {
  background-color: #f5f7fa;
}

.key-value-table :deep(.el-table__header th) {
  background-color: #f5f7fa;
  color: #606266;
  font-weight: 500;
  border-bottom: 1px solid #ebeef5;
}

.key-value-table :deep(.el-table__row) {
  height: 40px;
}

.key-value-table :deep(.el-table__cell) {
  padding: 8px 0;
  border-bottom: 1px solid #ebeef5;
}

.key-value-table :deep(.el-table__body tr:hover > td) {
  background-color: #f5f7fa;
}

.key-value-table :deep(.el-table__row.empty-row) {
  color: #c0c4cc;
  font-style: italic;
}

.key-value-table :deep(.el-table__row.empty-row:hover > td) {
  background-color: inherit;
}

.key-cell {
  display: flex;
  align-items: center;
  padding: 0 12px;
  height: 100%;
}

.key-text {
  color: #606266;
  font-weight: 500;
  font-size: 13px;
}

.value-cell {
  display: flex;
  align-items: center;
  position: relative;
  padding: 0 12px;
  min-height: 40px;
}

.value-preview {
  flex: 1;
  color: #909399;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.4;
  word-break: break-all;
  cursor: default;
}

.empty-state {
  padding: 40px 20px;
  text-align: center;
}

/* 滚动条样式 */
.data-list::-webkit-scrollbar {
  width: 6px;
}

.data-list::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.data-list::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.data-list::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}
</style>



