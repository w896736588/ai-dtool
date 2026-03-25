<template>
  <div class="execution-history">
    <div class="history-header">
      <div class="filter-controls">
        <el-select v-model="filter.status" placeholder="状态筛选" clearable>
          <el-option label="全部成功" value="success" />
          <el-option label="全部失败" value="error" />
        </el-select>

        <el-date-picker
            v-model="filter.dateRange"
            type="daterange"
            range-separator="至"
            start-placeholder="开始日期"
            end-placeholder="结束日期"
            value-format="YYYY-MM-DD"
        />

        <pl-button type="primary" @click="loadHistory">查询</pl-button>
        <pl-button @click="resetFilter">重置</pl-button>
      </div>

      <div class="header-actions">
        <pl-button @click="clearHistory" type="danger">
          <el-icon><Delete /></el-icon>
          清空历史
        </pl-button>
      </div>
    </div>

    <el-table
        :data="historyList"
        style="width: 100%"
        v-loading="loading"
        empty-text="暂无执行历史"
    >
      <el-table-column prop="apiName" label="接口名称" min-width="200">
        <template #default="{ row }">
          <div class="api-info">
            <el-tag :type="getMethodTagType(row.method)" size="small">
              {{ row.method }}
            </el-tag>
            <span class="api-name">{{ row.apiName }}</span>
          </div>
        </template>
      </el-table-column>

      <el-table-column prop="url" label="请求URL" min-width="300" show-overflow-tooltip>
        <template #default="{ row }">
          <span class="request-url">{{ row.url }}</span>
        </template>
      </el-table-column>

      <el-table-column prop="status" label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 'success' ? 'success' : 'danger'" size="small">
            {{ row.status === 'success' ? '成功' : '失败' }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column prop="statusCode" label="状态码" width="100" align="center">
        <template #default="{ row }">
          <span :class="getStatusCodeClass(row.statusCode)">{{ row.statusCode }}</span>
        </template>
      </el-table-column>

      <el-table-column prop="responseTime" label="响应时间" width="120" align="center">
        <template #default="{ row }">
          <span>{{ row.responseTime }}ms</span>
        </template>
      </el-table-column>

      <el-table-column prop="executedAt" label="执行时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.executedAt) }}
        </template>
      </el-table-column>

      <el-table-column label="操作" width="120" align="center" fixed="right">
        <template #default="{ row }">
          <pl-button type="primary" link @click="viewDetail(row)">详情</pl-button>
          <pl-button type="primary" link @click="rerunTest(row)">重试</pl-button>
        </template>
      </el-table-column>
    </el-table>

    <div class="pagination-container">
      <el-pagination
          v-model:current-page="pagination.currentPage"
          v-model:page-size="pagination.pageSize"
          :total="pagination.total"
          :page-sizes="[10, 20, 50, 100]"
          layout="total, sizes, prev, pager, next, jumper"
          @size-change="loadHistory"
          @current-change="loadHistory"
      />
    </div>

    <!-- 执行详情对话框 -->
    <el-dialog
        v-model="detailDialogVisible"
        :title="`执行详情 - ${selectedHistory?.apiName}`"
        width="90%"
        top="5vh"
    >
      <history-detail :history="selectedHistory" v-if="detailDialogVisible" />
    </el-dialog>
  </div>
</template>

<script>
import { Delete } from '@element-plus/icons-vue'
import HistoryDetail from './HistoryDetail.vue'

export default {
  name: 'ExecutionHistory',
  components: {
    Delete,
    HistoryDetail
  },
  props: {
    folderId: {
      type: [String, Number],
      default: ''
    }
  },
  data() {
    return {
      loading: false,
      historyList: [],
      filter: {
        status: '',
        dateRange: []
      },
      pagination: {
        currentPage: 1,
        pageSize: 10,
        total: 0
      },
      detailDialogVisible: false,
      selectedHistory: null
    }
  },
  mounted() {
    this.loadHistory()
  },
  methods: {
    loadHistory() {
      this.loading = true
      // 模拟加载执行历史
      setTimeout(() => {
        this.historyList = [
          {
            id: 1,
            apiName: '用户登录',
            method: 'POST',
            url: 'https://api.example.com/api/v1/auth/login',
            status: 'success',
            statusCode: 200,
            responseTime: 156,
            executedAt: new Date().toISOString(),
            requestData: { username: 'admin', password: '123456' },
            responseData: { code: 0, message: 'success', data: { token: 'abc123' } }
          },
          {
            id: 2,
            apiName: '获取用户信息',
            method: 'GET',
            url: 'https://api.example.com/api/v1/user/info',
            status: 'success',
            statusCode: 200,
            responseTime: 89,
            executedAt: new Date(Date.now() - 3600000).toISOString(),
            requestData: {},
            responseData: { code: 0, message: 'success', data: { name: '管理员' } }
          },
          {
            id: 3,
            apiName: '更新用户信息',
            method: 'PUT',
            url: 'https://api.example.com/api/v1/user/info',
            status: 'error',
            statusCode: 401,
            responseTime: 203,
            executedAt: new Date(Date.now() - 7200000).toISOString(),
            requestData: { name: '新名称' },
            responseData: { code: 401, message: '未授权' }
          }
        ]
        this.pagination.total = this.historyList.length
        this.loading = false
      }, 800)
    },

    resetFilter() {
      this.filter = {
        status: '',
        dateRange: []
      }
      this.loadHistory()
    },

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

    formatTime(timeString) {
      if (!timeString) return '-'
      return new Date(timeString).toLocaleString('zh-CN')
    },

    viewDetail(history) {
      this.selectedHistory = history
      this.detailDialogVisible = true
    },

    rerunTest(history) {
      this.$message.info(`重新执行接口: ${history.apiName}`)
      // 这里可以触发重新执行接口的逻辑
    },

    clearHistory() {
      this.$confirm('确定要清空所有执行历史吗？此操作不可恢复。', '确认清空', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        this.historyList = []
        this.pagination.total = 0
        this.$message.success('清空成功')
      })
    }
  }
}
</script>

<style scoped>
.execution-history {
  padding: 12px;
  border: 1px solid #e8eee5;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 6px 18px rgba(80, 110, 80, 0.08);
}

.history-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  flex-wrap: wrap;
  gap: 12px;
  padding: 10px 12px;
  border: 1px solid #e6ece0;
  border-radius: 10px;
  background: #f7f9f5;
}

.filter-controls {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.api-info {
  display: flex;
  align-items: center;
  gap: 8px;
}

.api-name {
  font-weight: 500;
}

.request-url {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  color: #606266;
}

.status-success {
  color: #67c23a;
  font-weight: bold;
}

.status-warning {
  color: #e6a23c;
  font-weight: bold;
}

.status-danger {
  color: #f56c6c;
  font-weight: bold;
}

.status-error {
  color: #f56c6c;
  font-weight: bold;
}

.pagination-container {
  display: flex;
  justify-content: flex-end;
  margin-top: 16px;
  padding: 16px 0;
  border-top: 1px solid #e8eee5;
}

:deep(.el-table) {
  border: 1px solid #e6ece0;
  border-radius: 10px;
  overflow: hidden;
}
</style>

