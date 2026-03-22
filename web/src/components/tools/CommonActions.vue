<template>
  <div class="common-actions">
    <div class="common-actions__header">
      <div class="common-actions__title">常用操作</div>
      <div class="common-actions__desc">可扩展动作面板，当前提供通用命令托管与端口占用查询。</div>
    </div>

    <el-row :gutter="16">
      <el-col :xs="24" :lg="16">
        <el-card shadow="hover" class="action-card action-card--primary">
          <template #header>
            <div class="action-card__header">
              <div class="action-card__title">命令托管</div>
              <div class="action-card__subtitle">默认托管 cc-connect，可改成任意长期运行命令；进入页面会自动确保运行。</div>
            </div>
          </template>

          <el-form label-position="top" @submit.prevent>
            <el-row :gutter="12">
              <el-col :xs="24" :md="12">
                <el-form-item label="显示名称">
                  <el-input
                    v-model.trim="managedForm.name"
                    placeholder="例如 cc-connect"
                    @change="handleManagedConfigChange"
                  />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :md="12">
                <el-form-item label="唯一标识">
                  <el-input
                    v-model.trim="managedForm.key"
                    placeholder="例如 cc-connect"
                    @change="handleManagedConfigChange"
                  />
                </el-form-item>
              </el-col>
            </el-row>

            <el-form-item label="启动命令">
              <el-input
                v-model.trim="managedForm.command_line"
                placeholder='例如 cc-connect --config C:\Users\94804\.cc-connect\config.toml'
                @change="handleManagedConfigChange"
              />
            </el-form-item>

            <el-form-item label="工作目录">
              <el-input
                v-model.trim="managedForm.workdir"
                clearable
                placeholder="可选，不填则使用后端当前目录"
                @change="handleManagedConfigChange"
              />
            </el-form-item>
          </el-form>

          <el-alert
            :title="managedStatusBanner"
            type="info"
            :closable="false"
            show-icon
            class="action-card__alert"
          />

          <div class="managed-meta">
            <div class="managed-meta__item">
              <span class="managed-meta__label">状态</span>
              <el-tag :type="managedProcess.running ? 'success' : 'info'">
                {{ managedProcess.status_text || (managedProcess.running ? '运行中' : '未运行') }}
              </el-tag>
            </div>
            <div class="managed-meta__item">
              <span class="managed-meta__label">PID</span>
              <span>{{ managedProcess.pid || '-' }}</span>
            </div>
            <div class="managed-meta__item">
              <span class="managed-meta__label">日志文件</span>
              <span class="managed-meta__value">{{ managedProcess.log_file || '-' }}</span>
            </div>
          </div>

          <div class="action-card__buttons action-card__buttons--wrap">
            <el-button type="primary" :loading="managedLoading.ensure || managedLoading.start" @click="startManagedProcess">
              启动
            </el-button>
            <el-button :loading="managedLoading.stop" @click="stopManagedProcess">关闭</el-button>
            <el-button :loading="managedLoading.restart" @click="restartManagedProcess">重启</el-button>
            <el-button :loading="managedLoading.status" @click="refreshManagedState">刷新状态</el-button>
          </div>

          <div class="managed-log">
            <div class="managed-log__header">
              <div class="action-card__title action-card__title--small">实时日志</div>
              <div class="managed-log__hint">轮询最新 {{ managedLogMaxBytes / 1024 }}KB</div>
            </div>
            <div ref="managedLogContent" class="managed-log__content">{{ managedLogContent || '暂无日志输出' }}</div>
          </div>
        </el-card>

        <el-card shadow="hover" class="action-card">
          <template #header>
            <div class="action-card__header">
              <div class="action-card__title">端口进程管理</div>
              <div class="action-card__subtitle">输入端口，先查询占用进程，再确认结束。</div>
            </div>
          </template>

          <el-form @submit.prevent>
            <el-row :gutter="12">
              <el-col :xs="24" :sm="14" :md="12">
                <el-form-item label="端口">
                  <el-input
                    v-model.trim="portInput"
                    clearable
                    placeholder="例如 8080"
                    @keyup.enter="queryProcesses"
                  />
                </el-form-item>
              </el-col>
              <el-col :xs="24" :sm="10" :md="12" class="action-card__buttons">
                <el-button type="primary" :loading="queryLoading" @click="queryProcesses">查询占用进程</el-button>
                <el-button :disabled="!lastQueryPort || queryLoading" @click="refreshProcesses">刷新</el-button>
              </el-col>
            </el-row>
          </el-form>

          <el-alert
            title="结束操作会强制终止目标进程，请先确认 PID 和进程名。"
            type="warning"
            :closable="false"
            show-icon
            class="action-card__alert"
          />

          <el-empty v-if="hasSearched && processList.length === 0 && !queryLoading" description="当前端口没有监听进程" />

          <el-table
            v-if="processList.length > 0"
            :data="processList"
            border
            stripe
            class="action-card__table"
          >
            <el-table-column prop="pid" label="PID" width="120" />
            <el-table-column prop="command" label="进程名" min-width="180">
              <template #default="scope">
                {{ scope.row.command || '-' }}
              </template>
            </el-table-column>
            <el-table-column prop="protocol" label="协议" width="120" />
            <el-table-column prop="address" label="监听地址" min-width="220" />
            <el-table-column label="操作" width="160" fixed="right">
              <template #default="scope">
                <el-button
                  type="danger"
                  plain
                  size="small"
                  :loading="killingPid === scope.row.pid"
                  @click="confirmKill(scope.row)"
                >
                  结束进程
                </el-button>
              </template>
            </el-table-column>
          </el-table>
        </el-card>
      </el-col>

      <el-col :xs="24" :lg="8">
        <el-card shadow="never" class="action-card action-card--muted">
          <template #header>
            <div class="action-card__title">后续可扩展</div>
          </template>
          <div class="placeholder-list">
            <div class="placeholder-item">通用命令托管</div>
            <div class="placeholder-item">按端口查占用</div>
            <div class="placeholder-item">结束指定 PID</div>
            <div class="placeholder-item">打开常用目录</div>
            <div class="placeholder-item">清理临时文件</div>
          </div>
        </el-card>
      </el-col>
    </el-row>
  </div>
</template>

<script>
import { nextTick } from 'vue'
import { ElMessageBox } from 'element-plus'
import toolsApi from '@/utils/base/tools'

const managedProcessStorageKey = 'tools.common_actions.managed_process'
const defaultManagedProcessForm = Object.freeze({
  name: 'cc-connect',
  key: 'cc-connect',
  command_line: 'cc-connect --config C:\\Users\\94804\\.cc-connect\\config.toml',
  workdir: '',
})

export default {
  name: 'CommonActions',
  data() {
    return {
      portInput: '',
      processList: [],
      queryLoading: false,
      killingPid: 0,
      hasSearched: false,
      lastQueryPort: 0,
      managedForm: {
        ...defaultManagedProcessForm,
      },
      managedProcess: {
        running: false,
        pid: 0,
        log_file: '',
        status_text: '未运行',
      },
      managedLogContent: '',
      managedLogMaxBytes: 32 * 1024,
      managedLoading: {
        ensure: false,
        status: false,
        start: false,
        stop: false,
        restart: false,
        log: false,
      },
      managedPollingTimer: null,
      managedInitialized: false,
      suppressManagedConfigChange: true,
    }
  },
  computed: {
    managedStatusBanner() {
      const commandLine = this.managedForm.command_line || '未配置命令'
      const workdirText = this.managedForm.workdir || '后端当前目录'
      return `当前命令：${commandLine}；工作目录：${workdirText}。配置项变更后会自动重启。`
    },
  },
  mounted() {
    this.initManagedProcessCard()
  },
  beforeUnmount() {
    this.clearManagedPolling()
  },
  methods: {
    initManagedProcessCard() {
      this.loadManagedForm()
      this.$nextTick(() => {
        this.suppressManagedConfigChange = false
      })
      this.ensureManagedProcessRunning()
      this.startManagedPolling()
    },
    loadManagedForm() {
      const stored = window.localStorage.getItem(managedProcessStorageKey)
      if (!stored) {
        return
      }
      try {
        const parsed = JSON.parse(stored)
        this.managedForm = {
          ...defaultManagedProcessForm,
          ...parsed,
        }
      } catch (error) {
        this.managedForm = {
          ...defaultManagedProcessForm,
        }
      }
    },
    saveManagedForm() {
      window.localStorage.setItem(managedProcessStorageKey, JSON.stringify(this.managedForm))
    },
    getManagedPayload(extra = {}) {
      return {
        key: (this.managedForm.key || '').trim(),
        name: (this.managedForm.name || '').trim(),
        command_line: (this.managedForm.command_line || '').trim(),
        workdir: (this.managedForm.workdir || '').trim(),
        ...extra,
      }
    },
    updateManagedState(response) {
      if (!(response && response.ErrCode === 0 && response.Data)) {
        return false
      }
      this.managedProcess = {
        running: !!response.Data.running,
        pid: Number(response.Data.pid || 0),
        log_file: response.Data.log_file || '',
        status_text: response.Data.status_text || (response.Data.running ? '运行中' : '未运行'),
      }
      this.managedInitialized = true
      return true
    },
    ensureManagedProcessRunning(showSuccess) {
      if (!this.getManagedPayload().command_line) {
        this.$helperNotify.error('请先填写启动命令')
        return
      }
      this.managedLoading.ensure = true
      toolsApi.ToolManagedProcessEnsureRunning(this.getManagedPayload(), (response) => {
        this.managedLoading.ensure = false
        if (!this.updateManagedState(response)) {
          return
        }
        if (showSuccess) {
          this.$helperNotify.success(this.managedProcess.running ? '命令已启动' : '命令未运行')
        }
        this.refreshManagedLog()
      })
    },
    refreshManagedState() {
      this.managedLoading.status = true
      toolsApi.ToolManagedProcessStatus(this.getManagedPayload(), (response) => {
        this.managedLoading.status = false
        if (!this.updateManagedState(response)) {
          return
        }
        this.refreshManagedLog()
      })
    },
    startManagedProcess() {
      this.managedLoading.start = true
      toolsApi.ToolManagedProcessStart(this.getManagedPayload(), (response) => {
        this.managedLoading.start = false
        if (!this.updateManagedState(response)) {
          return
        }
        this.$helperNotify.success('命令已启动')
        this.refreshManagedLog()
      })
    },
    stopManagedProcess() {
      this.managedLoading.stop = true
      toolsApi.ToolManagedProcessStop(this.getManagedPayload(), (response) => {
        this.managedLoading.stop = false
        if (!this.updateManagedState(response)) {
          return
        }
        this.managedLogContent = ''
        this.$helperNotify.success('命令已关闭')
      })
    },
    restartManagedProcess(showSuccess = true) {
      this.managedLoading.restart = true
      toolsApi.ToolManagedProcessRestart(this.getManagedPayload(), (response) => {
        this.managedLoading.restart = false
        if (!this.updateManagedState(response)) {
          return
        }
        if (showSuccess) {
          this.$helperNotify.success('命令已重启')
        }
        this.refreshManagedLog()
      })
    },
    handleManagedConfigChange() {
      if (this.suppressManagedConfigChange) {
        return
      }
      this.saveManagedForm()
      if (!this.getManagedPayload().command_line) {
        return
      }
      this.restartManagedProcess(false)
    },
    refreshManagedLog() {
      this.managedLoading.log = true
      toolsApi.ToolManagedProcessLogTail({
        ...this.getManagedPayload(),
        max_bytes: this.managedLogMaxBytes,
      }, (response) => {
        this.managedLoading.log = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          return
        }
        this.managedLogContent = response.Data.content || ''
        if (response.Data.log_file) {
          this.managedProcess.log_file = response.Data.log_file
        }
        this.scrollManagedLogToBottom()
      })
    },
    scrollManagedLogToBottom() {
      nextTick(() => {
        const el = this.$refs.managedLogContent
        if (!el) {
          return
        }
        el.scrollTop = el.scrollHeight
      })
    },
    startManagedPolling() {
      this.clearManagedPolling()
      this.managedPollingTimer = window.setInterval(() => {
        this.refreshManagedState()
      }, 3000)
    },
    clearManagedPolling() {
      if (this.managedPollingTimer) {
        window.clearInterval(this.managedPollingTimer)
        this.managedPollingTimer = null
      }
    },
    parsePortValue() {
      const port = Number(this.portInput)
      if (!Number.isInteger(port) || port < 1 || port > 65535) {
        this.$helperNotify.error('请输入 1-65535 之间的端口')
        return 0
      }
      return port
    },
    queryProcesses() {
      const port = this.parsePortValue()
      if (!port) {
        return
      }
      this.queryLoading = true
      toolsApi.ToolPortProcessList({ port }, (response) => {
        this.queryLoading = false
        this.hasSearched = true
        if (response.ErrCode !== 0) {
          return
        }
        this.lastQueryPort = port
        this.processList = Array.isArray(response.Data?.items) ? response.Data.items : []
        if (this.processList.length === 0) {
          this.$helperNotify.warning('当前端口没有监听进程')
          return
        }
        this.$helperNotify.success(`已查询到 ${this.processList.length} 个进程`)
      })
    },
    refreshProcesses() {
      if (!this.lastQueryPort) {
        return
      }
      this.portInput = String(this.lastQueryPort)
      this.queryProcesses()
    },
    async confirmKill(row) {
      try {
        await ElMessageBox.confirm(
          `确认结束 PID ${row.pid}${row.command ? `（${row.command}）` : ''} 吗？`,
          '结束进程确认',
          {
            type: 'warning',
            confirmButtonText: '确认结束',
            cancelButtonText: '取消',
          }
        )
      } catch (error) {
        return
      }

      this.killingPid = row.pid
      toolsApi.ToolPortProcessKill({ pid: row.pid }, (response) => {
        this.killingPid = 0
        if (response.ErrCode !== 0) {
          return
        }
        this.$helperNotify.success(`PID ${row.pid} 已结束`)
        this.refreshProcesses()
      })
    },
  },
}
</script>

<style scoped>
.common-actions {
  padding: 4px 6px 18px;
}

.common-actions__header {
  margin-bottom: 14px;
}

.common-actions__title {
  font-size: 18px;
  font-weight: 600;
  color: #324a34;
}

.common-actions__desc {
  margin-top: 4px;
  color: #66756a;
  font-size: 13px;
}

.action-card {
  border-radius: 12px;
  margin-bottom: 16px;
}

.action-card--primary {
  border: 1px solid #d8e8d8;
  background: linear-gradient(180deg, #f8fcf6 0%, #f2f8ef 100%);
}

.action-card--muted {
  background: linear-gradient(180deg, #fafcf8 0%, #f4f8f1 100%);
}

.action-card__header {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.action-card__title {
  font-size: 16px;
  font-weight: 600;
  color: #35553a;
}

.action-card__subtitle {
  color: #708171;
  font-size: 12px;
}

.action-card__title--small {
  font-size: 14px;
}

.action-card__buttons {
  display: flex;
  align-items: center;
  gap: 8px;
}

.action-card__buttons--wrap {
  flex-wrap: wrap;
  margin-bottom: 16px;
}

.action-card__alert {
  margin-bottom: 16px;
}

.action-card__table {
  margin-top: 8px;
}

.managed-meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-bottom: 16px;
}

.managed-meta__item {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  color: #48604c;
  font-size: 13px;
}

.managed-meta__label {
  min-width: 64px;
  color: #6a7a6c;
}

.managed-meta__value {
  word-break: break-all;
}

.managed-log {
  border: 1px solid #dbe7d9;
  border-radius: 10px;
  background: #1d281f;
  overflow: hidden;
}

.managed-log__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 12px;
  background: rgba(255, 255, 255, 0.06);
  color: #dbe7d9;
}

.managed-log__hint {
  font-size: 12px;
  color: #b7cab8;
}

.managed-log__content {
  max-height: 320px;
  overflow: auto;
  padding: 12px;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: Consolas, Monaco, monospace;
  font-size: 12px;
  line-height: 1.55;
  color: #e7f4e8;
}

.placeholder-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.placeholder-item {
  padding: 12px 14px;
  border: 1px dashed #c8d7c5;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.72);
  color: #55715a;
}

@media (max-width: 768px) {
  .action-card__buttons {
    justify-content: flex-start;
    margin-bottom: 8px;
  }
}
</style>
