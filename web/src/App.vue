<template>
  <div id="app">
    <router-view/>
    <div v-if="gitPendingTotalCount > 0" class="status-indicator git-pending-indicator" @click="gitDialogVisible = true">
      Git 未提交 {{ gitPendingTotalCount }}
    </div>
    <div
      v-if="sseConnectionCount > 0"
      class="status-indicator sse-connection-indicator"
      :style="{ backgroundColor: sseConnectionColor }"
      :title="'当前 SSE 连接数: ' + sseConnectionCount + '/' + sseConnectionTotal"
    >
      SSE {{ sseConnectionCount }}/{{ sseConnectionTotal }}
    </div>
    <el-dialog v-model="gitDialogVisible" title="Git 未提交文件" width="720px">
      <div v-if="gitRepos.length === 0">暂无未提交文件</div>
      <div v-for="repo in gitRepos" :key="repo.label + repo.dir" class="git-repo-block">
        <div class="git-repo-head">
          <div>
            <div class="git-repo-title">{{ repo.label }} · {{ repo.count }}</div>
            <div class="git-repo-dir">{{ repo.dir }}</div>
          </div>
          <el-button
            type="primary"
            size="small"
            :loading="commitPushLoadingMap[repo.dir] === true"
            @click="commitPushRepo(repo)"
          >
            commit+push
          </el-button>
        </div>
        <el-table :data="repo.files.map(item => ({ path: item }))" size="small" border max-height="240">
          <el-table-column prop="path" label="文件" />
        </el-table>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import base from '@/utils/base'
import git from '@/utils/base/git'
import sseDistribute from '@/utils/base/sse_distribute'
import { ElMessage, ElMessageBox } from 'element-plus'

const SseConnectionCountId = 'sse_connection_count'
const GitPendingStatusId = 'git_pending_status'

export default {
  name: 'App',
  data() {
    return {
      sseConnectionCount: 0,
      sseConnectionTotal: 0,
      gitPendingTotalCount: 0,
      gitRepos: [],
      gitDialogVisible: false,
      commitPushLoadingMap: {},
    }
  },
  computed: {
    sseConnectionColor() {
      const total = this.sseConnectionTotal
      if (!total) return '#67C23A'
      const pct = Math.round((this.sseConnectionCount / total) * 100)
      if (pct >= 100) return '#F56C6C'
      if (pct >= 90) return '#E6A23C'
      return '#67C23A'
    },
  },
  mounted() {
    base.DisableSaveShortcut()
    const favicon = document.querySelector('link[rel="icon"]')
    if (process.env.NODE_ENV === 'production' && favicon) {
      favicon.href = './favicon.ico'
    }
    this.registerSseConnectionCount()
    this.registerGitPendingStatus()
  },
  unmounted() {
    sseDistribute.UnRegisterReceive(SseConnectionCountId)
    sseDistribute.UnRegisterReceive(GitPendingStatusId)
  },
  methods: {
    registerSseConnectionCount() {
      sseDistribute.RegisterReceive(SseConnectionCountId, (data) => {
        if (data && typeof data === 'object') {
          this.sseConnectionCount = data.count || 0
          this.sseConnectionTotal = data.total || 0
        }
      })
    },
    registerGitPendingStatus() {
      sseDistribute.RegisterReceive(GitPendingStatusId, (data) => {
        if (!data || typeof data !== 'object') {
          this.gitPendingTotalCount = 0
          this.gitRepos = []
          return
        }
        this.gitPendingTotalCount = Number(data.total_count || 0)
        this.gitRepos = Array.isArray(data.repos) ? data.repos : []
      })
    },
    async commitPushRepo(repo) {
      const dir = repo && repo.dir ? String(repo.dir).trim() : ''
      if (!dir) return
      try {
        const result = await ElMessageBox.prompt('请输入 commit message', 'commit+push', {
          confirmButtonText: '提交',
          cancelButtonText: '取消',
          inputValue: `chore: sync pending changes ${new Date().toLocaleString()}`,
          inputPattern: /\S+/,
          inputErrorMessage: 'commit message 不能为空',
        })
        const message = (result.value || '').trim()
        if (!message) return
        this.commitPushLoadingMap = {
          ...this.commitPushLoadingMap,
          [dir]: true,
        }
        git.GitPendingCommitPush({ dir, message }, (response) => {
          this.commitPushLoadingMap = {
            ...this.commitPushLoadingMap,
            [dir]: false,
          }
          if (!response || response.ErrCode !== 0) {
            ElMessage.error((response && response.ErrMsg) || 'commit+push 失败')
            return
          }
          ElMessage.success('commit+push 成功')
          this.refreshGitPendingStatus()
        })
      } catch (err) {
        return
      }
    },
    refreshGitPendingStatus() {
      base.BasePost('/api/GitPendingStatus', {}, (response) => {
        if (!response || response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.gitPendingTotalCount = Number(response.Data.total_count || 0)
        this.gitRepos = Array.isArray(response.Data.repos) ? response.Data.repos : []
      })
    },
  },
}
</script>

<style>
html,
body,
#app {
  height: 100%;
}

#app {
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif;
  -moz-osx-font-smoothing: grayscale;
  color: #2c3e50;
}

body {
  margin: 0;
}

.status-indicator {
  position: fixed;
  bottom: 16px;
  min-width: 96px;
  box-sizing: border-box;
  padding: 6px 12px;
  border-radius: 10px;
  color: #fff;
  font-size: 12px;
  line-height: 1.2;
  font-weight: 700;
  text-align: center;
  z-index: 9999;
  user-select: none;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.2);
}

.git-pending-indicator {
  right: 128px;
  background: #e6a23c;
  cursor: pointer;
}

.sse-connection-indicator {
  right: 16px;
  transition: background-color 0.3s;
}

.git-repo-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 8px;
}

.git-repo-block + .git-repo-block {
  margin-top: 16px;
}

.git-repo-title {
  font-weight: 600;
}

.git-repo-dir {
  margin: 4px 0 8px;
  color: #909399;
  font-size: 12px;
  word-break: break-all;
}
</style>
