<template>
  <div class="info-crawl-page">
    <aside class="info-crawl-sidebar">
      <div class="sidebar-header">
        <div>
          <div class="sidebar-title">信息抓取</div>
          <div class="sidebar-desc">模型、提示词与实时输出</div>
        </div>
        <pl-button type="primary" plain @click="createTask">
          <el-icon><Plus /></el-icon>
          新建任务
        </pl-button>
      </div>

      <el-scrollbar class="sidebar-scroll">
        <button
          v-for="item in taskList"
          :key="item.id"
          class="task-item"
          :class="{ active: currentTaskId === item.id }"
          @click="selectTask(item.id)"
        >
          <div class="task-item-title">{{ item.name || '未命名任务' }}</div>
          <div class="task-item-time">{{ item.update_time_desc || '-' }}</div>
        </button>
      </el-scrollbar>
    </aside>

    <section class="info-crawl-main" v-loading="taskLoading">
      <div v-if="!hasTaskEditor" class="empty-wrap">
        <el-empty description="请先创建一个信息抓取任务" />
      </div>

      <template v-else>
        <div class="page-toolbar">
          <div class="toolbar-title-wrap">
            <div class="page-title">任务配置</div>
            <div class="page-desc">配置 AI 模型与提示词，并在提示词中附带待采集的网址</div>
            <div
              v-if="crawl4aiStatus.status !== 'ready'"
              class="crawl4ai-status-banner"
              :class="{
                installing: crawl4aiStatus.status === 'installing',
                failed: crawl4aiStatus.status === 'failed',
              }"
            >
              {{ crawl4aiBannerText }}
            </div>
          </div>
          <div class="toolbar-actions">
            <pl-button plain @click="openCrawl4AIInstallDialog(false)">
              <el-icon><Tools /></el-icon>
              安装指引
            </pl-button>
            <pl-button :disabled="!currentTaskId" @click="openHistoryDrawer">
              <el-icon><Clock /></el-icon>
              执行历史
            </pl-button>
            <pl-button type="danger" plain :disabled="!currentTaskId" @click="deleteTask">
              <el-icon><Delete /></el-icon>
              删除任务
            </pl-button>
            <pl-button type="primary" :loading="taskSaving" @click="saveTask">
              <el-icon><Check /></el-icon>
              保存任务
            </pl-button>
            <pl-button
              type="success"
              :loading="runSubmitting || runWatching"
              :disabled="!currentTaskId || !isCrawl4AIReady"
              @click="runTask"
            >
              <el-icon><VideoPlay /></el-icon>
              执行任务
            </pl-button>
          </div>
        </div>

        <div class="editor-grid">
          <div class="editor-card">
            <div class="card-title">基础信息</div>
            <el-form label-position="top">
              <el-form-item label="任务名称">
                <el-input v-model="taskForm.name" placeholder="例如：竞品动态抓取" />
              </el-form-item>
              <el-form-item label="AI 模型">
                <el-select v-model="taskForm.ai_model_id" placeholder="请选择 AI 模型" style="width: 100%">
                  <el-option
                    v-for="item in aiModelList"
                    :key="item.id"
                    :label="(item.name || item.model) + ' / ' + (item.provider_name || '-')"
                    :value="item.id"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="任务提示词">
                <el-input
                  v-model="taskForm.prompt"
                  type="textarea"
                  :rows="12"
                  placeholder="请输入采集要求，并在提示词中附带 http/https 网址，例如：https://example.com"
                />
              </el-form-item>
            </el-form>
          </div>

          <div class="editor-card run-live-card">
            <div class="card-head">
              <div class="card-title">实时输出</div>
              <pl-button plain @click="clearRunLiveLog">
                <el-icon><Refresh /></el-icon>
                清空输出
              </pl-button>
            </div>
            <div class="run-live-status">{{ runLiveStatus }}</div>
            <pre class="detail-pre live-pre">{{ runLiveLog || '等待执行输出...' }}</pre>
          </div>
        </div>
      </template>
    </section>

    <el-drawer v-model="historyDrawerVisible" title="执行历史" size="42%">
      <div class="drawer-toolbar">
        <pl-button type="primary" plain @click="refreshRunList">刷新</pl-button>
      </div>
      <el-empty v-if="runList.length === 0" description="暂无执行历史" />
      <div v-else class="history-list">
        <button
          v-for="item in runList"
          :key="item.id"
          class="history-item"
          @click="openRunDetail(item.id)"
        >
          <div class="history-item-head">
            <span>{{ item.create_time_desc || '-' }}</span>
            <el-tag
              size="small"
              effect="light"
              :type="item.status === 'success' ? 'success' : item.status === 'running' ? 'info' : 'danger'"
            >
              {{ item.status }}
            </el-tag>
          </div>
          <div class="history-item-desc">{{ item.run_message || '无摘要' }}</div>
        </button>
      </div>
    </el-drawer>

    <el-dialog v-model="runDetailVisible" title="执行详情" width="72%" top="4vh" class="run-dialog">
      <div v-if="runDetail.run_info.id" class="run-detail">
        <div class="detail-section">
          <div class="detail-title">执行状态</div>
          <div class="detail-meta">
            <el-tag size="small" effect="light">{{ runDetail.run_info.status }}</el-tag>
            <span>{{ runDetail.run_info.create_time_desc || '-' }}</span>
          </div>
          <pre class="detail-pre">{{ runDetail.run_info.run_message || '-' }}</pre>
        </div>

        <div class="detail-section">
          <div class="detail-title">任务提示词快照</div>
          <pre class="detail-pre">{{ runDetail.run_info.prompt_snapshot || '-' }}</pre>
        </div>

        <div class="detail-section">
          <div class="detail-title">最终输出</div>
          <pre class="detail-pre">{{ runDetail.run_info.output_content || '暂无输出内容' }}</pre>
        </div>

        <div v-if="runDetail.run_info.error_message" class="detail-section">
          <div class="detail-title">错误信息</div>
          <pre class="detail-pre">{{ runDetail.run_info.error_message }}</pre>
        </div>
      </div>
    </el-dialog>

    <el-dialog
      v-model="crawl4aiInstallDialogVisible"
      :title="crawl4aiInstallGuide.title || 'Crawl4AI Docker 安装指引'"
      width="820px"
      class="crawl4ai-install-dialog"
    >
      <div class="crawl4ai-install-dialog-body">
        <div class="crawl4ai-install-tip">
          {{ crawl4aiInstallGuide.tip || '建议通过 Docker 启动 Crawl4AI 服务，Windows 建议使用 WSL。' }}
        </div>

        <div class="crawl4ai-install-tip highlight">
          {{ crawl4aiInstallGuide.use_wsl_tip || 'Windows 环境建议使用 WSL 运行 Docker 命令。' }}
        </div>

        <div class="crawl4ai-install-section">
          <div class="detail-title">安装步骤</div>
          <div class="crawl4ai-install-step">1. 先确保本机已安装 Docker，Windows 建议在 WSL 中执行下面命令。</div>
          <div class="crawl4ai-install-step">2. 先执行镜像拉取命令。</div>
          <div class="crawl4ai-install-step">3. 再执行容器启动命令，服务会监听 `11235` 端口并设置开机自启。</div>
          <div class="crawl4ai-install-step">4. 启动后访问下方地址确认服务正常，再刷新当前页面。</div>
        </div>

        <div class="crawl4ai-install-section">
          <div class="crawl4ai-install-command-head">
            <div class="detail-title">1. 拉取镜像</div>
            <pl-button size="small" type="primary" plain @click="copyText(crawl4aiInstallGuide.pull_command, '拉取命令已复制')">
              复制命令
            </pl-button>
          </div>
          <pre class="detail-pre crawl4ai-install-command">{{ crawl4aiInstallGuide.pull_command }}</pre>
        </div>

        <div class="crawl4ai-install-section">
          <div class="crawl4ai-install-command-head">
            <div class="detail-title">2. 启动并设置开机自启</div>
            <pl-button size="small" type="primary" plain @click="copyText(crawl4aiInstallGuide.run_command, '启动命令已复制')">
              复制命令
            </pl-button>
          </div>
          <pre class="detail-pre crawl4ai-install-command">{{ crawl4aiInstallGuide.run_command }}</pre>
        </div>

        <div class="crawl4ai-install-section">
          <div class="crawl4ai-install-command-head">
            <div class="detail-title">3. 访问服务</div>
            <pl-button size="small" type="success" plain @click="copyText(crawl4aiInstallGuide.docs_url, '访问地址已复制')">
              复制地址
            </pl-button>
          </div>
          <pre class="detail-pre crawl4ai-install-command">{{ crawl4aiInstallGuide.docs_url }}</pre>
          <el-link :href="crawl4aiInstallGuide.docs_url" target="_blank" type="primary">
            打开 Playground
          </el-link>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { Check, Clock, Delete, Plus, Refresh, Tools, VideoPlay } from '@element-plus/icons-vue'
import AiSetApi from '@/utils/base/ai_set'
import InfoCrawlApi from '@/utils/base/info_crawl'
import sseDistribute from '@/utils/base/sse_distribute'

// defaultInstallGuide 返回默认 Docker 安装指引。
function defaultInstallGuide() {
  return {
    title: 'Crawl4AI Docker 安装指引',
    tip: '建议通过 Docker 启动 Crawl4AI 服务，Windows 建议使用 WSL。',
    use_wsl_tip: 'Windows 环境建议使用 WSL 运行 Docker 命令。',
    pull_command: 'docker pull unclecode/crawl4ai:latest',
    run_command: 'docker run -d --name crawl4ai -p 11235:11235 --shm-size=2g --restart always unclecode/crawl4ai:latest',
    docs_url: 'http://localhost:11235/playground/',
  }
}

export default {
  name: 'InfoCrawl',
  components: {
    Check,
    Clock,
    Delete,
    Plus,
    Refresh,
    Tools,
    VideoPlay,
  },
  data() {
    return {
      taskList: [],
      currentTaskId: 0,
      isCreatingTask: false,
      taskLoading: false,
      taskSaving: false,
      taskForm: {
        id: 0,
        name: '',
        prompt: '',
        ai_model_id: 0,
      },
      runList: [],
      aiModelList: [],
      historyDrawerVisible: false,
      runDetailVisible: false,
      runDetail: {
        run_info: {},
      },
      runSubmitting: false,
      runWatching: false,
      runningRunId: 0,
      runStatusTimer: null,
      crawl4aiStatusTimer: null,
      crawl4aiStatus: {
        status: 'idle',
        status_text: '等待连接 Crawl4AI 服务',
        error_message: '',
        is_ready: false,
        is_installing: false,
        need_install: false,
        install_guide: {},
      },
      crawl4aiInstallDialogVisible: false,
      crawl4aiInstallDialogAutoOpened: false,
      crawl4aiInstallGuide: defaultInstallGuide(),
      runLiveLog: '',
      runLiveStatus: '未开始执行',
      runSseDistributeId: '',
    }
  },
  computed: {
    // hasTaskEditor 判断是否显示编辑区。
    hasTaskEditor() {
      return this.isCreatingTask || this.currentTaskId > 0 || !!this.taskForm.id
    },
    // isCrawl4AIReady 判断 Crawl4AI 是否已就绪。
    isCrawl4AIReady() {
      return !!this.crawl4aiStatus.is_ready
    },
    // crawl4aiNeedInstall 判断是否需要提示用户先部署服务。
    crawl4aiNeedInstall() {
      return !!this.crawl4aiStatus.need_install
    },
    // crawl4aiBannerText 返回顶部提示文案。
    crawl4aiBannerText() {
      if (this.crawl4aiNeedInstall) {
        return this.crawl4aiStatus.error_message || '未检测到 Crawl4AI 服务，请先按 Docker 指引完成启动后重试'
      }
      if (this.crawl4aiStatus.status === 'failed') {
        return this.crawl4aiStatus.error_message || 'Crawl4AI 服务连接失败'
      }
      return this.crawl4aiStatus.status_text || '正在检查 Crawl4AI 服务'
    },
  },
  mounted() {
    this.fetchCrawl4AIStatus()
    this.startCrawl4AIStatusWatch()
    this.loadAiModelList()
    this.loadTaskList()
  },
  beforeUnmount() {
    this.stopRunWatch()
    this.stopCrawl4AIStatusWatch()
    this.unregisterRunSse()
  },
  methods: {
    // normalizeCrawl4AIInstallGuide 规范化后端返回的安装指引。
    normalizeCrawl4AIInstallGuide(installGuide) {
      return {
        ...defaultInstallGuide(),
        ...(installGuide || {}),
      }
    },
    // applyCrawl4AIStatus 应用 Crawl4AI 状态并处理弹窗。
    applyCrawl4AIStatus(statusData) {
      this.crawl4aiInstallGuide = this.normalizeCrawl4AIInstallGuide(statusData?.install_guide)
      this.crawl4aiStatus = {
        status: statusData?.status || 'idle',
        status_text: statusData?.status_text || '等待连接 Crawl4AI 服务',
        error_message: statusData?.error_message || '',
        is_ready: !!statusData?.is_ready,
        is_installing: !!statusData?.is_installing,
        need_install: !!statusData?.need_install,
        install_guide: statusData?.install_guide || {},
      }
      if (!this.runSubmitting && !this.runWatching && this.crawl4aiStatus.is_installing) {
        this.runLiveStatus = this.crawl4aiStatus.status_text || '正在检查 Crawl4AI 服务'
        this.runLiveLog = this.crawl4aiStatus.error_message || '正在检查 Crawl4AI 服务，请稍候...'
      }
      if (this.crawl4aiNeedInstall && !this.crawl4aiInstallDialogAutoOpened) {
        this.openCrawl4AIInstallDialog(true)
        this.crawl4aiInstallDialogAutoOpened = true
      }
      if (!this.crawl4aiNeedInstall) {
        this.crawl4aiInstallDialogAutoOpened = false
      }
    },
    // fetchCrawl4AIStatus 查询 Crawl4AI 当前状态。
    fetchCrawl4AIStatus() {
      InfoCrawlApi.InfoCrawlCrawl4AIStatus((response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          return
        }
        this.applyCrawl4AIStatus(response.Data)
      })
    },
    // openCrawl4AIInstallDialog 打开安装指引弹窗。
    openCrawl4AIInstallDialog(isAutoOpen) {
      this.crawl4aiInstallDialogVisible = true
      if (!isAutoOpen) {
        this.crawl4aiInstallDialogAutoOpened = true
      }
    },
    // startCrawl4AIStatusWatch 轮询 Crawl4AI 状态。
    startCrawl4AIStatusWatch() {
      this.stopCrawl4AIStatusWatch()
      this.crawl4aiStatusTimer = window.setInterval(() => {
        this.fetchCrawl4AIStatus()
      }, 3000)
    },
    // stopCrawl4AIStatusWatch 停止轮询 Crawl4AI 状态。
    stopCrawl4AIStatusWatch() {
      if (this.crawl4aiStatusTimer) {
        window.clearInterval(this.crawl4aiStatusTimer)
        this.crawl4aiStatusTimer = null
      }
    },
    // loadAiModelList 加载 AI 模型列表。
    loadAiModelList() {
      AiSetApi.AiModelList({ model_type: 'llm' }, (response) => {
        this.aiModelList = response && response.ErrCode === 0 && Array.isArray(response.Data)
          ? response.Data
          : []
      })
    },
    // loadTaskList 加载任务列表。
    loadTaskList() {
      this.taskLoading = true
      InfoCrawlApi.InfoCrawlTaskList((response) => {
        this.taskLoading = false
        this.taskList = response && response.ErrCode === 0 && Array.isArray(response.Data?.task_list)
          ? response.Data.task_list
          : []
        if (this.currentTaskId) {
          const exists = this.taskList.some((item) => item.id === this.currentTaskId)
          if (exists) {
            return
          }
        }
        if (!this.isCreatingTask && this.taskList.length > 0) {
          this.selectTask(this.taskList[0].id)
        }
      })
    },
    // loadTaskDetail 加载任务详情。
    loadTaskDetail(taskId) {
      if (!taskId) {
        return
      }
      this.taskLoading = true
      InfoCrawlApi.InfoCrawlTaskInfo(taskId, (response) => {
        this.taskLoading = false
        if (!(response && response.ErrCode === 0 && response.Data?.task)) {
          return
        }
        this.isCreatingTask = false
        this.currentTaskId = taskId
        this.taskForm = {
          id: response.Data.task.id || 0,
          name: response.Data.task.name || '',
          prompt: response.Data.task.prompt || '',
          ai_model_id: response.Data.task.ai_model_id || 0,
        }
        this.runList = Array.isArray(response.Data.run_list) ? response.Data.run_list : []
      })
    },
    // selectTask 选择任务。
    selectTask(taskId) {
      this.clearRunLiveLog()
      this.loadTaskDetail(taskId)
    },
    // createTask 创建新任务表单。
    createTask() {
      this.stopRunWatch()
      this.unregisterRunSse()
      this.isCreatingTask = true
      this.currentTaskId = 0
      this.runList = []
      this.runDetail = { run_info: {} }
      this.taskForm = {
        id: 0,
        name: '',
        prompt: '',
        ai_model_id: 0,
      }
      this.runLiveLog = ''
      this.runLiveStatus = '请填写任务后保存'
    },
    // saveTask 保存任务。
    saveTask() {
      this.taskSaving = true
      InfoCrawlApi.InfoCrawlTaskSave(this.taskForm, (response) => {
        this.taskSaving = false
        if (!(response && response.ErrCode === 0 && response.Data?.id)) {
          return
        }
        this.$helperNotify.success('保存成功')
        this.currentTaskId = response.Data.id
        this.isCreatingTask = false
        this.taskForm = {
          id: response.Data.id || 0,
          name: response.Data.name || '',
          prompt: response.Data.prompt || '',
          ai_model_id: response.Data.ai_model_id || 0,
        }
        this.loadTaskList()
        this.loadTaskDetail(this.currentTaskId)
      })
    },
    // deleteTask 删除当前任务。
    deleteTask() {
      if (!this.currentTaskId) {
        return
      }
      this.$confirm('确认删除当前任务吗？', '提示', {
        type: 'warning',
      }).then(() => {
        InfoCrawlApi.InfoCrawlTaskDelete(this.currentTaskId, (response) => {
          if (!(response && response.ErrCode === 0)) {
            return
          }
          this.$helperNotify.success('删除成功')
          this.stopRunWatch()
          this.unregisterRunSse()
          this.currentTaskId = 0
          this.isCreatingTask = false
          this.taskForm = {
            id: 0,
            name: '',
            prompt: '',
            ai_model_id: 0,
          }
          this.runList = []
          this.runDetail = { run_info: {} }
          this.runLiveLog = ''
          this.runLiveStatus = '未开始执行'
          this.loadTaskList()
        })
      }).catch(() => {})
    },
    // openHistoryDrawer 打开执行历史抽屉。
    openHistoryDrawer() {
      if (!this.currentTaskId) {
        return
      }
      this.historyDrawerVisible = true
      this.refreshRunList()
    },
    // refreshRunList 刷新执行历史。
    refreshRunList() {
      if (!this.currentTaskId) {
        this.runList = []
        return
      }
      InfoCrawlApi.InfoCrawlRunList(this.currentTaskId, 20, (response) => {
        this.runList = response && response.ErrCode === 0 && Array.isArray(response.Data?.run_list)
          ? response.Data.run_list
          : []
      })
    },
    // openRunDetail 打开执行详情。
    openRunDetail(runId) {
      if (!runId) {
        return
      }
      InfoCrawlApi.InfoCrawlRunInfo(runId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data?.run_info)) {
          return
        }
        this.runDetail = response.Data
        this.runDetailVisible = true
      })
    },
    // runTask 执行当前任务。
    runTask() {
      if (!this.currentTaskId) {
        this.$helperNotify.error('请先保存任务')
        return
      }
      if (!this.isCrawl4AIReady) {
        this.fetchCrawl4AIStatus()
        this.runLiveStatus = this.crawl4aiStatus.status_text || '正在检查 Crawl4AI 服务'
        this.runLiveLog = this.crawl4aiStatus.error_message || 'Crawl4AI 尚未就绪，请稍后重试'
        if (this.crawl4aiNeedInstall) {
          this.openCrawl4AIInstallDialog(false)
        }
        this.$helperNotify.error(this.crawl4aiBannerText)
        return
      }
      if (this.runWatching) {
        this.$helperNotify.error('当前已有任务在后台执行，请等待完成后再发起')
        return
      }
      const sseDistributeId = sseDistribute.GetSseDistributeId(`info_crawl_${this.currentTaskId}`)
      this.unregisterRunSse()
      this.runSseDistributeId = sseDistributeId
      this.runLiveLog = ''
      this.runLiveStatus = '正在提交任务'
      this.registerRunSse(sseDistributeId)
      this.runSubmitting = true
      InfoCrawlApi.InfoCrawlTaskRun({
        task_id: this.currentTaskId,
        sse_distribute_id: sseDistributeId,
      }, (response) => {
        this.runSubmitting = false
        if (!(response && response.ErrCode === 0 && response.Data?.run_id)) {
          this.fetchCrawl4AIStatus()
          if (response?.Data?.status_text) {
            this.runLiveLog = response.Data.error_message || this.runLiveLog
            this.runLiveStatus = response.Data.status_text
          } else {
            this.runLiveStatus = '任务提交失败'
          }
          if (response?.Data?.need_install) {
            this.openCrawl4AIInstallDialog(false)
          }
          this.unregisterRunSse()
          return
        }
        this.runningRunId = response.Data.run_id || 0
        this.runWatching = true
        this.runLiveStatus = response.Data.run_message || '任务已提交'
        this.refreshRunList()
        this.startRunWatch()
        this.$helperNotify.success('任务已提交，正在后台执行')
      })
    },
    // startRunWatch 开始轮询执行状态。
    startRunWatch() {
      this.stopRunWatch()
      if (!this.runningRunId) {
        return
      }
      this.runWatching = true
      this.fetchRunStatus()
      this.runStatusTimer = window.setInterval(() => {
        this.fetchRunStatus()
      }, 3000)
    },
    // stopRunWatch 停止轮询执行状态。
    stopRunWatch() {
      if (this.runStatusTimer) {
        window.clearInterval(this.runStatusTimer)
        this.runStatusTimer = null
      }
      this.runWatching = false
      this.runningRunId = 0
    },
    // fetchRunStatus 查询执行状态。
    fetchRunStatus() {
      if (!this.runningRunId) {
        return
      }
      InfoCrawlApi.InfoCrawlRunInfo(this.runningRunId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data?.run_info)) {
          return
        }
        const runInfo = response.Data.run_info
        if (this.runDetailVisible && this.runDetail.run_info?.id === this.runningRunId) {
          this.runDetail = response.Data
        }
        if (runInfo.status === 'running') {
          return
        }
        this.runLiveStatus = runInfo.run_message || `任务执行完成，状态：${runInfo.status || '-'}`
        this.stopRunWatch()
        this.unregisterRunSse()
        this.refreshRunList()
        this.loadTaskDetail(this.currentTaskId)
      })
    },
    // registerRunSse 注册执行过程 SSE。
    registerRunSse(sseDistributeId) {
      if (!sseDistributeId) {
        return
      }
      sseDistribute.RegisterReceive(sseDistributeId, (msg, msgType) => {
        if (msgType === 'info_crawl_status') {
          this.runLiveStatus = msg || '任务执行中'
          return
        }
        if (msgType === 'info_crawl_chunk') {
          this.runLiveStatus = 'AI 正在输出'
          this.runLiveLog += (msg || '').replace(/\r/g, '')
          return
        }
        if (msgType === 'info_crawl_done') {
          this.runLiveStatus = msg || '执行完成'
          window.setTimeout(() => {
            this.fetchRunStatus()
          }, 400)
          return
        }
        if (msgType === 'error') {
          this.runLiveStatus = '执行失败'
          this.runLiveLog += `\n[错误] ${(msg || '').replace(/\r/g, '')}`
        }
      })
    },
    // unregisterRunSse 注销执行过程 SSE。
    unregisterRunSse() {
      if (!this.runSseDistributeId) {
        return
      }
      sseDistribute.UnRegisterReceive(this.runSseDistributeId)
      this.runSseDistributeId = ''
    },
    // clearRunLiveLog 清空实时输出。
    clearRunLiveLog() {
      this.runLiveLog = ''
      if (this.crawl4aiStatus.is_installing) {
        this.runLiveStatus = this.crawl4aiStatus.status_text || '正在检查 Crawl4AI 服务'
        this.runLiveLog = this.crawl4aiStatus.error_message || '正在检查 Crawl4AI 服务，请稍候...'
        return
      }
      if (!this.runSubmitting && !this.runWatching) {
        this.runLiveStatus = this.currentTaskId ? '未开始执行' : '请先创建或选择任务'
      }
    },
    // copyText 复制文本到剪贴板。
    copyText(text, successMsg) {
      const val = (text || '').trim()
      if (!val) {
        this.$message.error('无可复制内容')
        return
      }
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(val).then(() => {
          this.$message.success(successMsg || '复制成功')
        }).catch(() => {
          this.$message.error('复制失败，请手动复制')
        })
        return
      }
      const input = document.createElement('textarea')
      input.value = val
      document.body.appendChild(input)
      input.select()
      try {
        document.execCommand('copy')
        this.$message.success(successMsg || '复制成功')
      } catch (e) {
        this.$message.error('复制失败，请手动复制')
      } finally {
        document.body.removeChild(input)
      }
    },
  },
}
</script>

<style scoped>
.info-crawl-page {
  display: flex;
  min-height: calc(100vh - 96px);
  gap: 16px;
  padding: 16px;
  background: #f5f7fa;
}

.info-crawl-sidebar {
  width: 280px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.06);
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 18px 18px 12px;
  border-bottom: 1px solid #eef2f7;
}

.sidebar-title,
.page-title,
.card-title,
.detail-title {
  font-size: 16px;
  font-weight: 600;
  color: #111827;
}

.sidebar-desc,
.page-desc,
.task-item-time,
.run-live-status {
  font-size: 12px;
  color: #6b7280;
}

.crawl4ai-status-banner {
  margin-top: 8px;
  padding: 8px 12px;
  border-radius: 10px;
  font-size: 12px;
  color: #92400e;
  background: #fef3c7;
}

.crawl4ai-status-banner.installing {
  color: #92400e;
  background: #fef3c7;
}

.crawl4ai-status-banner.failed {
  color: #991b1b;
  background: #fee2e2;
}

.sidebar-scroll {
  flex: 1;
  padding: 12px;
}

.task-item,
.history-item {
  width: 100%;
  border: 1px solid #e5e7eb;
  border-radius: 14px;
  background: #fff;
  padding: 12px 14px;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s ease;
}

.task-item + .task-item,
.history-item + .history-item {
  margin-top: 10px;
}

.task-item:hover,
.history-item:hover,
.task-item.active {
  border-color: #67c23a;
  box-shadow: 0 8px 20px rgba(103, 194, 58, 0.12);
}

.task-item-title {
  font-size: 14px;
  font-weight: 600;
  color: #111827;
  margin-bottom: 6px;
}

.info-crawl-main {
  flex: 1;
  min-width: 0;
}

.empty-wrap,
.editor-card,
.history-list,
.detail-section {
  background: #fff;
  border: 1px solid #e5e7eb;
  border-radius: 16px;
  box-shadow: 0 8px 24px rgba(15, 23, 42, 0.06);
}

.empty-wrap {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
}

.page-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 16px;
}

.toolbar-actions {
  display: flex;
  flex-wrap: wrap;
  justify-content: flex-end;
  gap: 12px;
}

.editor-grid {
  display: grid;
  grid-template-columns: minmax(360px, 480px) minmax(0, 1fr);
  gap: 16px;
}

.editor-card,
.detail-section {
  padding: 18px;
}

.card-head,
.history-item-head,
.detail-meta,
.drawer-toolbar,
.crawl4ai-install-command-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.run-live-card {
  min-height: 520px;
}

.live-pre {
  min-height: 440px;
}

.detail-pre {
  margin: 0;
  padding: 14px;
  background: #0f172a;
  color: #e2e8f0;
  border-radius: 12px;
  white-space: pre-wrap;
  word-break: break-word;
  font-size: 13px;
  line-height: 1.6;
}

.history-list {
  padding: 14px;
}

.history-item-desc {
  margin-top: 8px;
  font-size: 13px;
  color: #4b5563;
}

.run-detail,
.crawl4ai-install-dialog-body,
.crawl4ai-install-section {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.crawl4ai-install-tip,
.crawl4ai-install-step {
  font-size: 13px;
  line-height: 1.7;
  color: #4b5563;
}

.crawl4ai-install-tip.highlight {
  color: #92400e;
}

.crawl4ai-install-command {
  min-height: 88px;
}

@media (max-width: 1200px) {
  .editor-grid {
    grid-template-columns: 1fr;
  }
}

@media (max-width: 900px) {
  .info-crawl-page {
    flex-direction: column;
  }

  .info-crawl-sidebar {
    width: 100%;
  }

  .page-toolbar {
    flex-direction: column;
  }

  .toolbar-actions {
    justify-content: flex-start;
  }
}
</style>

