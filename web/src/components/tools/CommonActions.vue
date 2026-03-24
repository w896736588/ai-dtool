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
              <div class="action-card__subtitle">支持托管多个长期运行命令；进入页面后会自动逐个确保运行。</div>
            </div>
          </template>

          <div class="managed-layout">
            <div class="managed-sidebar">
              <div class="managed-sidebar__header">
                <div class="action-card__title action-card__title--small">命令列表</div>
                <pl-button type="primary" plain size="small" @click="addManagedProcessItem">新增命令</pl-button>
              </div>

              <div class="managed-sidebar__list">
                <div
                  v-for="item in managedProcessItems"
                  :key="item.id"
                  :class="['managed-sidebar__item', { 'is-active': item.id === activeManagedProcessId }]"
                  @click="selectManagedProcessItem(item.id)"
                >
                  <div class="managed-sidebar__item-main">
                    <div class="managed-sidebar__item-name">{{ item.name || item.key || '未命名命令' }}</div>
                    <div class="managed-sidebar__item-command">{{ item.command_line || '未配置启动命令' }}</div>
                  </div>
                  <div class="managed-sidebar__item-extra">
                    <el-tag size="small" :type="getManagedState(item.id).running ? 'success' : 'info'">
                      {{ getManagedState(item.id).status_text || (getManagedState(item.id).running ? '运行中' : '未运行') }}
                    </el-tag>
                    <pl-button
                      class="managed-sidebar__delete"
                      text
                      type="danger"
                      size="small"
                      @click.stop="removeManagedProcessItem(item.id)"
                    >
                      删除
                    </pl-button>
                  </div>
                </div>
              </div>
            </div>

            <div class="managed-panel">
              <template v-if="activeManagedProcessItem">
                <el-form label-position="top" @submit.prevent>
                  <el-row :gutter="12">
                    <el-col :xs="24" :md="12">
                      <el-form-item label="显示名称">
                        <el-input
                          v-model.trim="activeManagedProcessItem.name"
                          placeholder="例如 cc-connect"
                          @change="handleManagedConfigChange"
                        />
                      </el-form-item>
                    </el-col>
                    <el-col :xs="24" :md="12">
                      <el-form-item label="唯一标识">
                        <el-input
                          v-model.trim="activeManagedProcessItem.key"
                          placeholder="例如 cc-connect"
                          @change="handleManagedConfigChange"
                        />
                      </el-form-item>
                    </el-col>
                  </el-row>

                  <el-form-item label="启动命令">
                    <el-input
                      v-model.trim="activeManagedProcessItem.command_line"
                      placeholder="例如 cc-connect --config C:\\Users\\94804\\.cc-connect\\config.toml"
                      @change="handleManagedConfigChange"
                    />
                  </el-form-item>

                  <el-form-item label="工作目录">
                    <el-input
                      v-model.trim="activeManagedProcessItem.workdir"
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
                    <el-tag :type="activeManagedProcessState.running ? 'success' : 'info'">
                      {{ activeManagedProcessState.status_text || (activeManagedProcessState.running ? '运行中' : '未运行') }}
                    </el-tag>
                  </div>
                  <div class="managed-meta__item">
                    <span class="managed-meta__label">PID</span>
                    <span>{{ activeManagedProcessState.pid || '-' }}</span>
                  </div>
                  <div class="managed-meta__item">
                    <span class="managed-meta__label">日志文件</span>
                    <span class="managed-meta__value">{{ activeManagedProcessState.log_file || '-' }}</span>
                  </div>
                </div>

                <div class="action-card__buttons action-card__buttons--wrap">
                  <pl-button
                    type="primary"
                    :loading="getManagedLoading(activeManagedProcessId).ensure || getManagedLoading(activeManagedProcessId).start"
                    @click="startManagedProcess"
                  >
                    启动
                  </pl-button>
                  <pl-button :loading="getManagedLoading(activeManagedProcessId).stop" @click="stopManagedProcess">
                    关闭
                  </pl-button>
                  <pl-button :loading="getManagedLoading(activeManagedProcessId).restart" @click="restartManagedProcess()">
                    重启
                  </pl-button>
                  <pl-button :loading="getManagedLoading(activeManagedProcessId).status" @click="refreshManagedState()">
                    刷新状态
                  </pl-button>
                </div>

                <div class="managed-log">
                  <div class="managed-log__header">
                    <div class="action-card__title action-card__title--small">实时日志</div>
                    <div class="managed-log__hint">轮询最新 {{ managedLogMaxBytes / 1024 }}KB</div>
                  </div>
                  <el-alert
                    v-if="activeManagedLogNotice"
                    :title="activeManagedLogNotice"
                    type="warning"
                    :closable="false"
                    class="managed-log__notice"
                  />
                  <div ref="managedLogContent" class="managed-log__content">
                    {{ activeManagedLogContent || '暂无日志输出' }}
                  </div>
                </div>
              </template>
            </div>
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
                <pl-button type="primary" :loading="queryLoading" @click="queryProcesses">查询占用进程</pl-button>
                <pl-button :disabled="!lastQueryPort || queryLoading" @click="refreshProcesses">刷新</pl-button>
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
                <pl-button
                  type="danger"
                  plain
                  size="small"
                  :loading="killingPid === scope.row.pid"
                  @click="confirmKill(scope.row)"
                >
                  结束进程
                </pl-button>
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

const legacyManagedProcessStorageKey = 'tools.common_actions.managed_process'
const managedProcessStorageKey = 'tools.common_actions.managed_process.v2'
const defaultActionTab = 'managed'
const defaultManagedProcessForm = Object.freeze({
  name: 'cc-connect',
  key: 'cc-connect',
  command_line: 'cc-connect --config C:\\Users\\94804\\.cc-connect\\config.toml',
  workdir: '',
})

// 默认状态 / Default status for each managed command card.
function createDefaultManagedProcessState() {
  return {
    running: false,
    pid: 0,
    log_file: '',
    status_text: '未运行',
    is_managed: false,
  }
}

// 默认加载态 / Default loading flags for each managed command card.
function createDefaultManagedLoading() {
  return {
    ensure: false,
    status: false,
    start: false,
    stop: false,
    restart: false,
    log: false,
  }
}

// 默认命令项 / Default managed command entry.
function createDefaultManagedProcessItem(index = 1) {
  if (index === 1) {
    return {
      id: `managed-${Date.now()}`,
      ...defaultManagedProcessForm,
    }
  }

  return {
    id: `managed-${Date.now()}-${index}`,
    name: `命令 ${index}`,
    key: `command-${index}`,
    command_line: '',
    workdir: '',
  }
}

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
      managedProcessItems: [],
      activeManagedProcessId: '',
      managedProcessStateMap: {},
      managedLogContentMap: {},
      managedLogNoticeMap: {},
      managedLoadingMap: {},
      // 上次成功应用的运行配置 / Last applied runtime config for safe restart on key changes.
      managedRuntimeConfigMap: {},
      managedLogMaxBytes: 32 * 1024,
      managedStatusPollingTimer: null,
      managedLogPollingTimer: null,
      suppressManagedConfigChange: true,
    }
  },
  computed: {
    activeManagedProcessItem() {
      return this.managedProcessItems.find((item) => item.id === this.activeManagedProcessId) || null
    },
    activeManagedProcessState() {
      return this.getManagedState(this.activeManagedProcessId)
    },
    activeManagedLogContent() {
      return this.managedLogContentMap[this.activeManagedProcessId] || ''
    },
    activeManagedLogNotice() {
      return this.managedLogNoticeMap[this.activeManagedProcessId] || ''
    },
    managedStatusBanner() {
      if (!this.activeManagedProcessItem) {
        return '请选择一个命令项。'
      }
      const commandLine = this.activeManagedProcessItem.command_line || '未配置命令'
      const workdirText = this.activeManagedProcessItem.workdir || '后端当前目录'
      return `当前命令：${commandLine}；工作目录：${workdirText}。配置变更后会自动重启。`
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
      this.loadManagedProcessItems()
      this.$nextTick(() => {
        this.suppressManagedConfigChange = false
      })
      this.ensureAllManagedProcessesRunning()
      this.startManagedPolling()
    },
    loadManagedProcessItems() {
      const stored = window.localStorage.getItem(managedProcessStorageKey)
      const legacyStored = window.localStorage.getItem(legacyManagedProcessStorageKey)
      let items = []
      let activeId = ''

      if (stored) {
        try {
          const parsed = JSON.parse(stored)
          items = Array.isArray(parsed?.items) ? parsed.items : []
          activeId = parsed?.activeId || ''
        } catch (error) {
          items = []
          activeId = ''
        }
      } else if (legacyStored) {
        try {
          const parsed = JSON.parse(legacyStored)
          items = [{
            id: createDefaultManagedProcessItem().id,
            name: parsed?.name || defaultManagedProcessForm.name,
            key: parsed?.key || defaultManagedProcessForm.key,
            command_line: parsed?.command_line || defaultManagedProcessForm.command_line,
            workdir: parsed?.workdir || '',
          }]
        } catch (error) {
          items = []
        }
      }

      if (items.length === 0) {
        items = [createDefaultManagedProcessItem()]
      }

      this.managedProcessItems = items.map((item, index) => this.normalizeManagedProcessItem(item, index))
      this.activeManagedProcessId = this.managedProcessItems.some((item) => item.id === activeId)
        ? activeId
        : this.managedProcessItems[0].id
      this.saveManagedProcessItems()
    },
    normalizeManagedProcessItem(item, index) {
      const fallback = createDefaultManagedProcessItem(index + 1)
      return {
        id: item?.id || fallback.id,
        name: item?.name || fallback.name,
        key: item?.key || fallback.key,
        command_line: item?.command_line || fallback.command_line,
        workdir: item?.workdir || '',
      }
    },
    saveManagedProcessItems() {
      window.localStorage.setItem(managedProcessStorageKey, JSON.stringify({
        items: this.managedProcessItems,
        activeId: this.activeManagedProcessId,
      }))
    },
    getManagedState(id) {
      if (!id) {
        return createDefaultManagedProcessState()
      }
      if (!this.managedProcessStateMap[id]) {
        this.managedProcessStateMap[id] = createDefaultManagedProcessState()
      }
      return this.managedProcessStateMap[id]
    },
    getManagedLoading(id) {
      if (!id) {
        return createDefaultManagedLoading()
      }
      if (!this.managedLoadingMap[id]) {
        this.managedLoadingMap[id] = createDefaultManagedLoading()
      }
      return this.managedLoadingMap[id]
    },
    setManagedLoading(id, field, value) {
      if (!id) {
        return
      }
      const loading = this.getManagedLoading(id)
      loading[field] = value
    },
    buildManagedPayload(item) {
      return {
        key: (item?.key || '').trim(),
        name: (item?.name || '').trim(),
        command_line: (item?.command_line || '').trim(),
        workdir: (item?.workdir || '').trim(),
      }
    },
    getActiveManagedPayload() {
      return this.buildManagedPayload(this.activeManagedProcessItem)
    },
    updateManagedStateByResponse(id, response) {
      if (!(response && response.ErrCode === 0 && response.Data)) {
        return false
      }

      this.managedProcessStateMap[id] = {
        running: !!response.Data.running,
        pid: Number(response.Data.pid || 0),
        log_file: response.Data.log_file || '',
        status_text: response.Data.status_text || (response.Data.running ? '运行中' : '未运行'),
        is_managed: !!response.Data.is_managed,
      }

      // 外部进程没有可信日志文件 / External processes do not expose trusted log files here.
      if (!response.Data.is_managed) {
        this.managedProcessStateMap[id].log_file = ''
      }
      return true
    },
    selectManagedProcessItem(id) {
      this.activeManagedProcessId = id
      this.saveManagedProcessItems()
      this.refreshManagedState()
      this.refreshManagedLog()
    },
    addManagedProcessItem() {
      const nextIndex = this.managedProcessItems.length + 1
      const newItem = createDefaultManagedProcessItem(nextIndex)
      this.managedProcessItems.push(newItem)
      this.activeManagedProcessId = newItem.id
      this.saveManagedProcessItems()
      this.$nextTick(() => {
        this.refreshManagedState()
      })
    },
    async removeManagedProcessItem(id) {
      const item = this.managedProcessItems.find((one) => one.id === id)
      if (!item) {
        return
      }

      try {
        await ElMessageBox.confirm(
          `确认删除命令配置“${item.name || item.key || '未命名命令'}”吗？删除配置不会自动关闭已运行进程。`,
          '删除命令配置',
          {
            type: 'warning',
            confirmButtonText: '确认删除',
            cancelButtonText: '取消',
          }
        )
      } catch (error) {
        return
      }

      this.managedProcessItems = this.managedProcessItems.filter((one) => one.id !== id)
      delete this.managedProcessStateMap[id]
      delete this.managedLogContentMap[id]
      delete this.managedLogNoticeMap[id]
      delete this.managedLoadingMap[id]
      delete this.managedRuntimeConfigMap[id]

      if (this.managedProcessItems.length === 0) {
        const fallback = createDefaultManagedProcessItem()
        this.managedProcessItems = [fallback]
      }

      if (!this.managedProcessItems.some((one) => one.id === this.activeManagedProcessId)) {
        this.activeManagedProcessId = this.managedProcessItems[0].id
      }

      this.saveManagedProcessItems()
      this.$helperNotify.success('命令配置已删除')
    },
    ensureAllManagedProcessesRunning() {
      // 顺序执行，避免同屏初始化时并发请求过多 / Run sequentially to avoid burst requests on mount.
      this.managedProcessItems.forEach((item, index) => {
        window.setTimeout(() => {
          this.ensureManagedProcessRunning(item.id, false)
        }, index * 120)
      })
    },
    ensureManagedProcessRunning(id = this.activeManagedProcessId, showSuccess = false) {
      const item = this.managedProcessItems.find((one) => one.id === id)
      if (!item) {
        return
      }

      const payload = this.buildManagedPayload(item)
      if (!payload.command_line) {
        return
      }

      this.setManagedLoading(id, 'ensure', true)
      toolsApi.ToolManagedProcessEnsureRunning(payload, (response) => {
        this.setManagedLoading(id, 'ensure', false)
        if (!this.updateManagedStateByResponse(id, response)) {
          return
        }
        this.managedRuntimeConfigMap[id] = { ...payload }
        if (showSuccess) {
          this.$helperNotify.success(this.getManagedState(id).running ? '命令已启动' : '命令未运行')
        }
        if (id === this.activeManagedProcessId) {
          this.refreshManagedLog()
        }
      })
    },
    refreshManagedState(id = this.activeManagedProcessId) {
      const item = this.managedProcessItems.find((one) => one.id === id)
      if (!item) {
        return
      }

      const payload = this.buildManagedPayload(item)
      if (!payload.command_line) {
        this.managedProcessStateMap[id] = createDefaultManagedProcessState()
        return
      }

      this.setManagedLoading(id, 'status', true)
      toolsApi.ToolManagedProcessStatus(payload, (response) => {
        this.setManagedLoading(id, 'status', false)
        if (!this.updateManagedStateByResponse(id, response)) {
          return
        }
        if (id === this.activeManagedProcessId) {
          this.refreshManagedLog()
        }
      })
    },
    refreshAllManagedStatuses() {
      this.managedProcessItems.forEach((item) => {
        this.refreshManagedState(item.id)
      })
    },
    startManagedProcess() {
      const id = this.activeManagedProcessId
      const payload = this.getActiveManagedPayload()
      if (!payload.command_line) {
        this.$helperNotify.error('请先填写启动命令')
        return
      }

      this.setManagedLoading(id, 'start', true)
      toolsApi.ToolManagedProcessStart(payload, (response) => {
        this.setManagedLoading(id, 'start', false)
        if (!this.updateManagedStateByResponse(id, response)) {
          return
        }
        this.managedRuntimeConfigMap[id] = { ...payload }
        this.$helperNotify.success('命令已启动')
        this.refreshManagedLog()
      })
    },
    stopManagedProcess(options = {}) {
      const id = options.id || this.activeManagedProcessId
      const item = this.managedProcessItems.find((one) => one.id === id)
      if (!item) {
        return
      }

      const payload = options.payload || this.buildManagedPayload(item)
      if (!payload.command_line) {
        this.managedProcessStateMap[id] = createDefaultManagedProcessState()
        return
      }

      this.setManagedLoading(id, 'stop', true)
      toolsApi.ToolManagedProcessStop(payload, (response) => {
        this.setManagedLoading(id, 'stop', false)
        if (!this.updateManagedStateByResponse(id, response)) {
          return
        }
        this.managedLogContentMap[id] = ''
        this.managedLogNoticeMap[id] = ''
        if (options.clearRuntimeConfig !== false) {
          delete this.managedRuntimeConfigMap[id]
        }
        if (options.showSuccess !== false) {
          this.$helperNotify.success('命令已关闭')
        }
        if (typeof options.onSuccess === 'function') {
          options.onSuccess()
        }
      })
    },
    restartManagedProcess(showSuccess = true) {
      const id = this.activeManagedProcessId
      const payload = this.getActiveManagedPayload()
      if (!payload.command_line) {
        this.$helperNotify.error('请先填写启动命令')
        return
      }

      this.setManagedLoading(id, 'restart', true)
      toolsApi.ToolManagedProcessRestart(payload, (response) => {
        this.setManagedLoading(id, 'restart', false)
        if (!this.updateManagedStateByResponse(id, response)) {
          return
        }
        this.managedRuntimeConfigMap[id] = { ...payload }
        if (showSuccess) {
          this.$helperNotify.success('命令已重启')
        }
        this.refreshManagedLog()
      })
    },
    handleManagedConfigChange() {
      if (this.suppressManagedConfigChange || !this.activeManagedProcessItem) {
        return
      }

      const id = this.activeManagedProcessId
      const currentPayload = this.getActiveManagedPayload()
      const previousPayload = this.managedRuntimeConfigMap[id] || null

      this.saveManagedProcessItems()

      // 只有运行参数变更才自动重启 / Auto restart only when runtime fields changed.
      const runtimeChanged = !previousPayload ||
        previousPayload.key !== currentPayload.key ||
        previousPayload.command_line !== currentPayload.command_line ||
        previousPayload.workdir !== currentPayload.workdir

      if (!runtimeChanged) {
        return
      }

      if (!currentPayload.command_line) {
        this.managedProcessStateMap[id] = createDefaultManagedProcessState()
        this.managedLogContentMap[id] = ''
        this.managedLogNoticeMap[id] = ''
        delete this.managedRuntimeConfigMap[id]
        return
      }

      // key 变更时先按旧配置关闭，再按新配置启动 / Stop old config first when key changed, then start new config.
      if (previousPayload && previousPayload.key && previousPayload.key !== currentPayload.key) {
        this.stopManagedProcess({
          id,
          payload: previousPayload,
          showSuccess: false,
          clearRuntimeConfig: false,
          onSuccess: () => {
            this.startManagedProcess()
          },
        })
        return
      }

      if (!previousPayload || !previousPayload.command_line) {
        this.startManagedProcess()
        return
      }

      this.restartManagedProcess(false)
    },
    refreshManagedLog(id = this.activeManagedProcessId) {
      const item = this.managedProcessItems.find((one) => one.id === id)
      if (!item) {
        return
      }

      const payload = this.buildManagedPayload(item)
      if (!payload.command_line) {
        this.managedLogContentMap[id] = ''
        this.managedLogNoticeMap[id] = ''
        return
      }

      if (this.getManagedLoading(id).log) {
        return
      }

      this.setManagedLoading(id, 'log', true)
      toolsApi.ToolManagedProcessLogTail({
        ...payload,
        max_bytes: this.managedLogMaxBytes,
      }, (response) => {
        this.setManagedLoading(id, 'log', false)
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.managedLogNoticeMap[id] = ''
          return
        }
        this.managedLogNoticeMap[id] = response.Data.message || ''
        this.managedLogContentMap[id] = response.Data.content || ''
        if (response.Data.log_file && this.managedProcessStateMap[id]) {
          this.managedProcessStateMap[id].log_file = response.Data.log_file
        }
        if (id === this.activeManagedProcessId) {
          this.scrollManagedLogToBottom()
        }
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
      // 状态轮询覆盖全部命令，日志仅轮询当前选中项 / Poll all statuses, but only tail logs for active entry.
      this.clearManagedPolling()
      this.managedStatusPollingTimer = window.setInterval(() => {
        this.refreshAllManagedStatuses()
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
  height: 100%;
  padding: 0;
}

.common-actions__header {
  display: none;
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
  height: 100%;
  min-height: 640px;
}

.common-actions-tabs :deep(.el-tabs__content) {
  height: 100%;
  min-width: 0;
  padding-right: 184px;
}

.common-actions-tabs :deep(.el-tabs__header.is-right) {
  margin-left: 0;
  width: 168px;
}

.common-actions-tabs :deep(.el-tabs__nav-scroll) {
  width: 100%;
}

.common-actions-tabs :deep(.el-tabs__nav-wrap.is-right) {
  width: 168px;
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
  width: 100%;
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

.managed-layout {
  display: flex;
  align-items: flex-start;
  width: 100%;
  gap: 20px;
}

.managed-sidebar {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 0 0 240px;
  min-width: 0;
}

.managed-sidebar__header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.managed-sidebar__list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.managed-sidebar__item {
  display: flex;
  flex-direction: column;
  gap: 10px;
  width: 100%;
  padding: 12px;
  border: 1px solid #d8e8d8;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.72);
  cursor: pointer;
  text-align: left;
  transition: border-color 0.2s ease, transform 0.2s ease, box-shadow 0.2s ease;
}

.managed-sidebar__item:hover {
  border-color: #93b397;
  transform: translateY(-1px);
}

.managed-sidebar__item.is-active {
  border-color: #5a875f;
  box-shadow: 0 10px 24px rgba(62, 95, 67, 0.12);
  background: linear-gradient(180deg, #ffffff 0%, #eef7ea 100%);
}

.managed-sidebar__item-main {
  min-width: 0;
}

.managed-sidebar__item-name {
  font-size: 14px;
  font-weight: 600;
  color: #35553a;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.managed-sidebar__item-command {
  margin-top: 4px;
  font-size: 12px;
  color: #6a7a6c;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.managed-sidebar__item-extra {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
}

.managed-sidebar__delete {
  padding: 0;
}

.managed-panel {
  flex: 1 1 auto;
  min-width: 0;
  overflow: hidden;
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

@media (max-width: 960px) {
  .common-actions-tabs :deep(.el-tabs__content) {
    padding-right: 16px;
  }

  .common-actions-tabs :deep(.el-tabs__header.is-right) {
    width: auto;
  }

  .common-actions-tabs :deep(.el-tabs__nav-wrap.is-right) {
    width: auto;
  }

  .managed-layout {
    flex-direction: column;
  }

  .managed-sidebar {
    flex-basis: auto;
    width: 100%;
  }
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

