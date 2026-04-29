<template>
  <div class="key-value-view">
    <div class="view-header">
      <div class="header-actions">
        <pl-button type="primary" link @click="copyToClipboard">
          <el-icon><DocumentCopy /></el-icon>
          复制全部
        </pl-button>
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

<style scoped src="@/css/components/api/KeyValueView.css"></style>




