<template>
  <div class="common-actions">
    <div class="common-actions__header">
      <div class="common-actions__title">常用操作</div>
      <div class="common-actions__desc">当前操作集中到右侧菜单，左侧展示对应面板。</div>
    </div>

    <el-tabs
      v-model="activeActionTab"
      tab-position="right"
      class="common-actions-tabs"
      stretch
    >
      <el-tab-pane label="命令托管" name="managed">
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
                placeholder="例如 cc-connect --config C:\\Users\\94804\\.cc-connect\\config.toml"
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
            <el-alert
              v-if="managedLogNotice"
              :title="managedLogNotice"
              type="warning"
              :closable="false"
              class="managed-log__notice"
            />
            <div ref="managedLogContent" class="managed-log__content">{{ managedLogContent || '暂无日志输出' }}</div>
          </div>
        </el-card>
      </el-tab-pane>

      <el-tab-pane label="端口进程管理" name="port-process">
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
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import { nextTick } from 'vue'
import { ElMessageBox } from 'element-plus'
import toolsApi from '@/utils/base/tools'

const managedProcessStorageKey = 'tools.common_actions.managed_process'
const defaultActionTab = 'managed'
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
      activeActionTab: defaultActionTab,
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
      managedLogNotice: '',
      managedLogMaxBytes: 32 * 1024,
      managedLoading: {
        ensure: false,
        status: false,
        start: false,
        stop: false,
        restart: false,
        log: false,
      },
      managedStatusPollingTimer: null,
      managedLogPollingTimer: null,
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
      // 外部进程没有可信日志文件，避免继续展示旧日志 / External processes do not have a trustworthy log file, so clear stale log state.
      if (!response.Data.is_managed) {
        this.managedProcess.log_file = ''
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
        this.managedLogNotice = ''
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
      // 没有启动命令时不请求日志 / Skip log requests when the command is empty.
      if (!this.getManagedPayload().command_line) {
        this.managedLogContent = ''
        this.managedLogNotice = ''
        return
      }
      // 避免日志轮询重叠请求 / Avoid overlapping polling requests for log tail.
      if (this.managedLoading.log) {
        return
      }
      this.managedLoading.log = true
      toolsApi.ToolManagedProcessLogTail({
        ...this.getManagedPayload(),
        max_bytes: this.managedLogMaxBytes,
      }, (response) => {
        this.managedLoading.log = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.managedLogNotice = ''
          return
        }
        this.managedLogNotice = response.Data.message || ''
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
      // 状态轮询和日志轮询分开，避免日志刷新完全依赖状态接口 / Separate status polling from log polling so logs do not depend on status refresh.
      this.managedStatusPollingTimer = window.setInterval(() => {
        this.refreshManagedState()
      }, 3000)
      this.managedLogPollingTimer = window.setInterval(() => {
        this.refreshManagedLog()
      }, 1500)
    },
    clearManagedPolling() {
      if (this.managedStatusPollingTimer) {
        window.clearInterval(this.managedStatusPollingTimer)
        this.managedStatusPollingTimer = null
      }
      if (this.managedLogPollingTimer) {
        window.clearInterval(this.managedLogPollingTimer)
        this.managedLogPollingTimer = null
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

.common-actions-tabs {
  min-height: 640px;
}

.common-actions-tabs :deep(.el-tabs__content) {
  padding-right: 16px;
}

.common-actions-tabs :deep(.el-tabs__header.is-right) {
  margin-left: 0;
}

.common-actions-tabs :deep(.el-tabs__nav-wrap.is-right) {
  padding: 8px 0;
}

.common-actions-tabs :deep(.el-tabs__item.is-right) {
  justify-content: flex-start;
  min-width: 148px;
  padding: 14px 18px;
  color: #5d6f61;
  border-radius: 12px;
  margin-bottom: 8px;
}

.common-actions-tabs :deep(.el-tabs__item.is-right.is-active) {
  color: #264a2e;
  background: linear-gradient(180deg, #eef7ea 0%, #e1efdb 100%);
}

.common-actions-tabs :deep(.el-tabs__active-bar) {
  background-color: #5a875f;
}

.action-card {
  border-radius: 12px;
}

.action-card--primary {
  border: 1px solid #d8e8d8;
  background: linear-gradient(180deg, #f8fcf6 0%, #f2f8ef 100%);
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

.managed-log__notice {
  margin: 12px 12px 0;
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

@media (max-width: 768px) {
  .common-actions-tabs {
    min-height: auto;
  }

  .common-actions-tabs :deep(.el-tabs__header.is-right) {
    margin: 0 0 12px;
  }

  .common-actions-tabs :deep(.el-tabs__content) {
    padding-right: 0;
  }

  .action-card__buttons {
    justify-content: flex-start;
    margin-bottom: 8px;
    flex-wrap: wrap;
  }
}
</style>
