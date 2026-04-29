<template>
  <div class="history-detail">
    <el-tabs v-model="activeTab" class="detail-tabs">
      <!-- 请求信息标签页 -->
      <el-tab-pane label="请求信息" name="request">
        <div class="tab-content">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="接口名称">{{ history.apiName }}</el-descriptions-item>
            <el-descriptions-item label="请求方法">
              <el-tag :type="getMethodTagType(history.method)">{{ history.method }}</el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="请求URL" :span="2">
              <span class="url-text">{{ history.url }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="执行时间">{{ formatTime(history.executedAt) }}</el-descriptions-item>
            <el-descriptions-item label="响应时间">
              <span :class="getResponseTimeClass(history.responseTime)">
                {{ history.responseTime }}ms
              </span>
            </el-descriptions-item>
          </el-descriptions>

          <el-divider content-position="left">请求头</el-divider>
          <key-value-view
              v-if="history.requestHeaders && Object.keys(history.requestHeaders).length > 0"
              :data="history.requestHeaders"
          />
          <el-empty v-else description="无请求头" :image-size="60" />

          <el-divider content-position="left">请求参数</el-divider>
          <div v-if="history.method === 'GET' || history.method === 'DELETE'">
            <key-value-view
                v-if="history.queryParams && Object.keys(history.queryParams).length > 0"
                :data="history.queryParams"
            />
            <el-empty v-else description="无查询参数" :image-size="60" />
          </div>
          <div v-else>
            <div class="body-type-selector">
              <el-radio-group v-model="requestBodyViewType" size="small">
                <el-radio-button value="raw">Raw</el-radio-button>
                <el-radio-button value="formatted">格式化</el-radio-button>
              </el-radio-group>
            </div>
            <div class="request-body">
              <pre v-if="requestBodyViewType === 'raw'" class="raw-body">{{ formatRawBody(history.requestData) }}</pre>
              <key-value-view
                  v-else-if="history.requestData && Object.keys(history.requestData).length > 0"
                  :data="history.requestData"
              />
              <el-empty v-else description="无请求体" :image-size="60" />
            </div>
          </div>
        </div>
      </el-tab-pane>

      <!-- 响应信息标签页 -->
      <el-tab-pane label="响应信息" name="response">
        <div class="tab-content">
          <el-descriptions :column="2" border>
            <el-descriptions-item label="状态码">
              <span :class="getStatusCodeClass(history.statusCode)">{{ history.statusCode }}</span>
            </el-descriptions-item>
            <el-descriptions-item label="执行状态">
              <el-tag :type="history.status === 'success' ? 'success' : 'danger'">
                {{ history.status === 'success' ? '成功' : '失败' }}
              </el-tag>
            </el-descriptions-item>
            <el-descriptions-item label="响应大小" v-if="history.responseSize">
              {{ formatFileSize(history.responseSize) }}
            </el-descriptions-item>
            <el-descriptions-item label="Content-Type" v-if="history.responseHeaders?.['Content-Type']">
              {{ history.responseHeaders['Content-Type'] }}
            </el-descriptions-item>
          </el-descriptions>

          <el-divider content-position="left">响应头</el-divider>
          <key-value-view
              v-if="history.responseHeaders && Object.keys(history.responseHeaders).length > 0"
              :data="history.responseHeaders"
          />
          <el-empty v-else description="无响应头" :image-size="60" />

          <el-divider content-position="left">响应体</el-divider>
          <div class="response-body-section">
            <div class="body-toolbar">
              <el-radio-group v-model="responseBodyViewType" size="small">
                <el-radio-button value="raw">Raw</el-radio-button>
                <el-radio-button value="formatted">格式化</el-radio-button>
                <el-radio-button value="preview">预览</el-radio-button>
              </el-radio-group>

              <div class="toolbar-actions">
                <pl-button type="primary" link @click="copyResponseBody">
                  <el-icon><DocumentCopy /></el-icon>
                  复制
                </pl-button>
                <pl-button link @click="downloadResponseBody" v-if="isDownloadable">
                  <el-icon><Download /></el-icon>
                  下载
                </pl-button>
              </div>
            </div>

            <div class="response-body-content">
              <div v-if="responseBodyViewType === 'preview' && isJSON" class="json-preview">
                <!-- 这里可以使用 json-viewer 组件，如果没有可以注释掉 -->
                <!-- <json-viewer :value="parsedResponseBody" :expand-depth="3" copyable /> -->
                <pre class="formatted-body">{{ formatJSON(history.responseData) }}</pre>
              </div>
              <pre
                  v-else-if="responseBodyViewType === 'formatted' && isJSON"
                  class="formatted-body"
              >{{ formatJSON(history.responseData) }}</pre>
              <pre v-else class="raw-body">{{ formatRawBody(history.responseData) }}</pre>
            </div>
          </div>
        </div>
      </el-tab-pane>

      <!-- 断言结果标签页 -->
      <el-tab-pane label="断言结果" name="assertions">
        <div class="tab-content">
          <div v-if="history.assertions && history.assertions.length > 0">
            <el-table :data="history.assertions" style="width: 100%">
              <el-table-column prop="name" label="断言名称" min-width="200" />
              <el-table-column prop="expression" label="断言表达式" min-width="300" show-overflow-tooltip />
              <el-table-column prop="expected" label="期望值" min-width="150" show-overflow-tooltip />
              <el-table-column prop="actual" label="实际值" min-width="150" show-overflow-tooltip />
              <el-table-column prop="result" label="结果" width="100" align="center">
                <template #default="{ row }">
                  <el-tag :type="row.success ? 'success' : 'danger'" size="small">
                    {{ row.success ? '通过' : '失败' }}
                  </el-tag>
                </template>
              </el-table-column>
              <el-table-column prop="message" label="详细信息" min-width="200" show-overflow-tooltip />
            </el-table>
          </div>
          <el-empty v-else description="无断言配置" :image-size="80" />
        </div>
      </el-tab-pane>

      <!-- 控制台日志标签页 -->
      <el-tab-pane label="控制台日志" name="console">
        <div class="tab-content">
          <div v-if="history.consoleLogs && history.consoleLogs.length > 0" class="console-log">
            <div
                v-for="(log, index) in history.consoleLogs"
                :key="index"
                class="log-entry"
                :class="getLogLevelClass(log.level)"
            >
              <span class="log-time">[{{ formatLogTime(log.timestamp) }}]</span>
              <span class="log-level">[{{ log.level.toUpperCase() }}]</span>
              <span class="log-message">{{ log.message }}</span>
            </div>
          </div>
          <el-empty v-else description="无控制台日志" :image-size="80" />
        </div>
      </el-tab-pane>
    </el-tabs>

    <div class="detail-actions">
      <pl-button type="primary" @click="rerunTest">
        <el-icon><Refresh /></el-icon>
        重新执行
      </pl-button>
      <pl-button @click="exportResult">
        <el-icon><Download /></el-icon>
        导出结果
      </pl-button>
      <pl-button @click="createTestCase" v-if="history.status === 'success'">
        <el-icon><Plus /></el-icon>
        创建测试用例
      </pl-button>
      <pl-button type="danger" @click="deleteHistory">
        <el-icon><Delete /></el-icon>
        删除记录
      </pl-button>
    </div>
  </div>
</template>

<script>
import { DocumentCopy, Download, Refresh, Plus, Delete } from '@element-plus/icons-vue'
import KeyValueView from './KeyValueView.vue'

export default {
  name: 'HistoryDetail',
  components: {
    DocumentCopy,
    Download,
    Refresh,
    Plus,
    Delete,
    KeyValueView
  },
  props: {
    history: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      activeTab: 'request',
      requestBodyViewType: 'formatted',
      responseBodyViewType: 'formatted'
    }
  },
  computed: {
    isJSON() {
      try {
        const data = this.history.responseData
        if (typeof data === 'string') {
          JSON.parse(data)
        } else if (typeof data === 'object') {
          JSON.stringify(data)
        }
        return true
      } catch {
        return false
      }
    },

    parsedResponseBody() {
      try {
        const data = this.history.responseData
        if (typeof data === 'string') {
          return JSON.parse(data)
        }
        return data
      } catch {
        return this.history.responseData
      }
    },

    isDownloadable() {
      const contentType = this.history.responseHeaders?.['Content-Type']
      return contentType && (
          contentType.includes('application/octet-stream') ||
          contentType.includes('application/pdf') ||
          contentType.includes('image/') ||
          contentType.includes('text/')
      )
    }
  },
  methods: {
    getMethodTagType(method) {
      const types = {
        GET: 'success',
        POST: 'warning',
        PUT: 'primary',
        DELETE: 'danger',
        PATCH: 'info'
      }
      return types[method] || 'info'
    },

    getStatusCodeClass(statusCode) {
      if (statusCode >= 200 && statusCode < 300) return 'status-success'
      if (statusCode >= 300 && statusCode < 400) return 'status-warning'
      if (statusCode >= 400 && statusCode < 500) return 'status-danger'
      if (statusCode >= 500) return 'status-error'
      return ''
    },

    getResponseTimeClass(responseTime) {
      if (responseTime < 100) return 'response-fast'
      if (responseTime < 500) return 'response-normal'
      if (responseTime < 1000) return 'response-slow'
      return 'response-very-slow'
    },

    getLogLevelClass(level) {
      const classes = {
        info: 'log-info',
        warn: 'log-warn',
        error: 'log-error',
        debug: 'log-debug'
      }
      return classes[level] || 'log-info'
    },

    formatTime(timeString) {
      if (!timeString) return '-'
      return new Date(timeString).toLocaleString('zh-CN')
    },

    formatLogTime(timestamp) {
      if (!timestamp) return '-'
      return new Date(timestamp).toLocaleTimeString('zh-CN', {
        hour12: false,
        hour: '2-digit',
        minute: '2-digit',
        second: '2-digit',
        fractionalSecondDigits: 3
      })
    },

    formatFileSize(bytes) {
      if (!bytes) return '0 B'
      const units = ['B', 'KB', 'MB', 'GB']
      let size = bytes
      let unitIndex = 0
      while (size >= 1024 && unitIndex < units.length - 1) {
        size /= 1024
        unitIndex++
      }
      return `${size.toFixed(2)} ${units[unitIndex]}`
    },

    formatRawBody(data) {
      if (typeof data === 'string') {
        return data
      } else if (typeof data === 'object') {
        return JSON.stringify(data, null, 2)
      }
      return String(data)
    },

    formatJSON(data) {
      try {
        if (typeof data === 'string') {
          return JSON.stringify(JSON.parse(data), null, 2)
        } else if (typeof data === 'object') {
          return JSON.stringify(data, null, 2)
        }
        return String(data)
      } catch {
        return String(data)
      }
    },

    copyResponseBody() {
      const text = this.formatRawBody(this.history.responseData)
      this.copyToClipboard(text)
      this.$message.success('响应体已复制到剪贴板')
    },

    copyToClipboard(text) {
      const textArea = document.createElement('textarea')
      textArea.value = text
      document.body.appendChild(textArea)
      textArea.select()
      document.execCommand('copy')
      document.body.removeChild(textArea)
    },

    downloadResponseBody() {
      const blob = new Blob([this.formatRawBody(this.history.responseData)], {
        type: this.history.responseHeaders?.['Content-Type'] || 'text/plain'
      })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `response_${this.history.id}_${new Date().getTime()}.txt`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
      this.$message.success('响应体下载完成')
    },

    rerunTest() {
      this.$emit('rerun', this.history)
      this.$message.info(`开始重新执行: ${this.history.apiName}`)
    },

    exportResult() {
      const data = {
        history: this.history,
        exportTime: new Date().toISOString(),
        version: '1.0'
      }
      const blob = new Blob([JSON.stringify(data, null, 2)], { type: 'application/json' })
      const url = URL.createObjectURL(blob)
      const a = document.createElement('a')
      a.href = url
      a.download = `test_result_${this.history.id}.json`
      document.body.appendChild(a)
      a.click()
      document.body.removeChild(a)
      URL.revokeObjectURL(url)
      this.$message.success('测试结果导出完成')
    },

    createTestCase() {
      this.$emit('createTestCase', this.history)
      this.$message.info('创建测试用例功能开发中')
    },

    deleteHistory() {
      this.$confirm('确定要删除这条执行记录吗？此操作不可恢复。', '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.$emit('delete', this.history.id)
        this.$message.success('删除成功')
      })
    }
  }
}
</script>

<style scoped src="@/css/components/api/HistoryDetail.css"></style>

