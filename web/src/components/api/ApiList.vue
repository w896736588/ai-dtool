<template>
  <div class="api-list">
    <div class="list-header">
      <el-input
          v-model="searchKeyword"
          placeholder="搜索接口名称、路径、描述..."
          style="width: 300px;"
          clearable
      />
      <div class="header-actions">
        <pl-button type="primary" @click="handleCreateApi">
          <el-icon><Plus /></el-icon>
          新建接口
        </pl-button>
        <pl-button @click="handleBatchRun" :disabled="selectedApis.length === 0">
          <el-icon><VideoPlay /></el-icon>
          批量运行
        </pl-button>
      </div>
    </div>

    <el-table
        :data="filteredApis"
        style="width: 100%"
        v-loading="loading"
        @selection-change="handleSelectionChange"
    >
      <el-table-column type="selection" width="55" />

      <el-table-column prop="name" label="接口名称" min-width="200">
        <template #default="{ row }">
          <div class="api-name">
            <el-tag :type="getMethodTagType(row.method)" size="small" class="method-tag">
              {{ row.method }}
            </el-tag>
            <span class="name-text">{{ row.name }}</span>
          </div>
        </template>
      </el-table-column>

      <el-table-column prop="path" label="路径" min-width="250">
        <template #default="{ row }">
          <span class="api-path">{{ row.path }}</span>
        </template>
      </el-table-column>

      <el-table-column prop="description" label="描述" min-width="200" show-overflow-tooltip>
        <template #default="{ row }">
          {{ row.description || '-' }}
        </template>
      </el-table-column>

      <el-table-column prop="status" label="状态" width="100" align="center">
        <template #default="{ row }">
          <el-tag :type="row.status === 'active' ? 'success' : 'info'" size="small">
            {{ row.status === 'active' ? '启用' : '禁用' }}
          </el-tag>
        </template>
      </el-table-column>

      <el-table-column prop="updateTime" label="更新时间" width="180">
        <template #default="{ row }">
          {{ formatTime(row.updateTime) }}
        </template>
      </el-table-column>

      <el-table-column label="操作" width="200" align="center" fixed="right">
        <template #default="{ row }">
          <pl-button type="primary" link @click="handleEdit(row)">编辑</pl-button>
          <pl-button type="primary" link @click="handleExecute(row)">运行</pl-button>
          <el-dropdown @command="(command) => handleMore(command, row)">
            <pl-button type="primary" link>
              更多<el-icon class="el-icon--right"><arrow-down /></el-icon>
            </pl-button>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item command="copy">复制接口</el-dropdown-item>
                <el-dropdown-item command="export">导出接口</el-dropdown-item>
                <el-dropdown-item command="history" divided>执行历史</el-dropdown-item>
                <el-dropdown-item command="disable" divided>
                  {{ row.status === 'active' ? '禁用接口' : '启用接口' }}
                </el-dropdown-item>
                <el-dropdown-item command="delete" type="danger">删除</el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
        </template>
      </el-table-column>
    </el-table>

    <div class="list-footer">
      <div class="selected-info" v-if="selectedApis.length > 0">
        已选择 {{ selectedApis.length }} 个接口
      </div>
      <el-pagination
          v-model:current-page="currentPage"
          v-model:page-size="pageSize"
          :page-sizes="[10, 20, 50, 100]"
          :total="total"
          layout="total, sizes, prev, pager, next, jumper"
      />
    </div>
  </div>
</template>

<script>
import { Search, Plus, VideoPlay, ArrowDown } from '@element-plus/icons-vue'

export default {
  name: 'ApiList',
  components: {
    Search,
    Plus,
    VideoPlay,
    ArrowDown
  },
  props: {
    folder: {
      type: Object,
      required: true
    }
  },
  data() {
    return {
      loading: false,
      apiList: [],
      searchKeyword: '',
      selectedApis: [],
      currentPage: 1,
      pageSize: 10,
      total: 0
    }
  },
  computed: {
    filteredApis() {
      let filtered = this.apiList

      if (this.searchKeyword) {
        const keyword = this.searchKeyword.toLowerCase()
        filtered = filtered.filter(api =>
            api.name.toLowerCase().includes(keyword) ||
            api.path.toLowerCase().includes(keyword) ||
            (api.description && api.description.toLowerCase().includes(keyword))
        )
      }

      this.total = filtered.length
      const start = (this.currentPage - 1) * this.pageSize
      const end = start + this.pageSize
      return filtered.slice(start, end)
    }
  },
  watch: {
    folder: {
      handler(newVal) {
        this.loadApiList(newVal)
      },
      immediate: true
    }
  },
  methods: {
    loadApiList(folder) {
      this.loading = true
      // 模拟加载接口列表
      setTimeout(() => {
        this.apiList = [
          {
            id: 1,
            name: '用户登录',
            method: 'POST',
            path: '/api/v1/auth/login',
            description: '用户登录接口，支持用户名密码登录',
            status: 'active',
            updateTime: new Date().toISOString(),
            createTime: new Date().toISOString()
          },
          {
            id: 2,
            name: '获取用户信息',
            method: 'GET',
            path: '/api/v1/user/info',
            description: '获取当前登录用户的基本信息',
            status: 'active',
            updateTime: new Date().toISOString(),
            createTime: new Date().toISOString()
          },
          {
            id: 3,
            name: '更新用户信息',
            method: 'PUT',
            path: '/api/v1/user/info',
            description: '更新用户个人信息',
            status: 'active',
            updateTime: new Date().toISOString(),
            createTime: new Date().toISOString()
          }
        ]
        this.total = this.apiList.length
        this.loading = false
      }, 500)
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

    formatTime(timeString) {
      if (!timeString) return '-'
      return new Date(timeString).toLocaleDateString('zh-CN')
    },

    handleSelectionChange(selection) {
      this.selectedApis = selection
    },

    handleCreateApi() {
      this.$emit('createApi', this.folder.id)
    },

    handleBatchRun() {
      this.$confirm(`确定要批量运行选中的 ${this.selectedApis.length} 个接口吗？`, '批量运行', {
        confirmButtonText: '确定',
        cancelButtonText: '取消'
      }).then(() => {
        this.$message.success(`开始批量运行 ${this.selectedApis.length} 个接口`)
        this.selectedApis.forEach(api => {
          this.$emit('execute', api)
        })
      })
    },

    handleEdit(api) {
      this.$emit('edit', api)
    },

    handleExecute(api) {
      this.$emit('execute', api)
    },

    handleMore(command, api) {
      switch (command) {
        case 'copy':
          this.handleCopyApi(api)
          break
        case 'export':
          this.handleExportApi(api)
          break
        case 'history':
          this.handleShowHistory(api)
          break
        case 'disable':
          this.handleToggleStatus(api)
          break
        case 'delete':
          this.handleDeleteApi(api)
          break
      }
    },

    handleCopyApi(api) {
      this.$message.success(`已复制接口 "${api.name}"`)
    },

    handleExportApi(api) {
      this.$message.info(`导出接口 "${api.name}"`)
    },

    handleShowHistory(api) {
      this.$message.info(`查看接口 "${api.name}" 的执行历史`)
    },

    handleToggleStatus(api) {
      const newStatus = api.status === 'active' ? 'inactive' : 'active'
      this.$message.success(`接口 "${api.name}" 已${newStatus === 'active' ? '启用' : '禁用'}`)
    },

    handleDeleteApi(api) {
      this.$confirm(`确定要删除接口 "${api.name}" 吗？`, '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        const index = this.apiList.findIndex(item => item.id === api.id)
        if (index !== -1) {
          this.apiList.splice(index, 1)
          this.$message.success('删除成功')
        }
      })
    }
  }
}
</script>

<style scoped>
.api-list {
  padding: 0;
  border: 1px solid #e8eee5;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 6px 18px rgba(80, 110, 80, 0.08);
  padding: 12px;
}

.list-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 16px;
  padding: 10px 12px;
  border: 1px solid #e6ece0;
  border-radius: 10px;
  background: #f7f9f5;
}

.header-actions {
  display: flex;
  gap: 12px;
}

.api-name {
  display: flex;
  align-items: center;
  gap: 8px;
}

.method-tag {
  font-weight: bold;
  min-width: 60px;
  text-align: center;
}

.name-text {
  font-weight: 500;
}

.api-path {
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  color: #606266;
}

.list-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 16px;
  padding: 16px 0;
}

.selected-info {
  color: #4f7d4f;
  font-size: 14px;
}

:deep(.el-table) {
  border: 1px solid #e6ece0;
  border-radius: 10px;
  overflow: hidden;
}

:deep(.el-table__header-wrapper th) {
  background: #f7f9f5 !important;
  color: #4e594a;
}
</style>

