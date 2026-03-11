<template>
  <div class="info-crawl-page">
    <aside class="info-crawl-sidebar">
      <div class="sidebar-header">
        <div>
          <div class="sidebar-title">信息抓取</div>
          <div class="sidebar-desc">任务、网页和执行历史</div>
        </div>
        <el-button type="primary" plain @click="createTask">
          <el-icon><Plus /></el-icon>
          新建任务
        </el-button>
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
      <div v-if="!currentTaskId" class="empty-wrap">
        <el-empty description="请先创建一个信息抓取任务" />
      </div>

      <template v-else>
        <div class="page-toolbar">
          <div class="toolbar-title-wrap">
            <div class="page-title">任务配置</div>
            <div class="page-desc">配置任务目标、AI 模型和网页抓取列表</div>
          </div>
          <div class="toolbar-actions">
            <el-button @click="openHistoryDrawer">
              <el-icon><Clock /></el-icon>
              执行历史
            </el-button>
            <el-button type="danger" plain @click="deleteTask">
              <el-icon><Delete /></el-icon>
              删除任务
            </el-button>
            <el-button type="primary" :loading="taskSaving" @click="saveTask">
              <el-icon><Check /></el-icon>
              保存任务
            </el-button>
            <el-button type="success" :loading="runSubmitting || runWatching" @click="runTask">
              <el-icon><VideoPlay /></el-icon>
              执行任务
            </el-button>
          </div>
        </div>

        <div class="editor-grid">
          <div class="editor-card">
            <div class="card-title">基础信息</div>
            <el-form label-position="top">
              <el-form-item label="任务名称">
                <el-input v-model="taskForm.name" placeholder="例如：竞品公告抓取" />
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
                  :rows="8"
                  placeholder="例如：请重点关注价格、公告、发布时间和正文摘要"
                />
              </el-form-item>
            </el-form>
          </div>

          <div class="editor-card">
            <div class="card-head">
              <div class="card-title">网页配置</div>
              <el-button type="primary" plain @click="appendPage">
                <el-icon><Plus /></el-icon>
                新增网页
              </el-button>
            </div>

            <div v-if="pageList.length === 0" class="empty-inline">
              <el-empty description="还没有网页配置" />
            </div>

            <div v-else class="page-card-list">
              <div v-for="page in pageList" :key="page.id || page._temp_id" class="page-card">
                <div class="page-card-head">
                  <div>
                    <div class="page-card-title">{{ page.name || '未命名网页' }}</div>
                    <div class="page-card-status">
                      <el-tag
                        size="small"
                        effect="light"
                        :type="page.login_status === 1 ? 'success' : page.login_status === 2 ? 'warning' : 'info'"
                      >
                        {{ page.login_status_desc || '未登录' }}
                      </el-tag>
                    </div>
                  </div>
                  <el-button type="danger" plain size="small" @click="deletePage(page)">
                    <el-icon><Delete /></el-icon>
                    删除
                  </el-button>
                </div>

                <el-form label-position="top" class="page-form">
                  <div class="form-row">
                    <el-form-item label="网页名称">
                      <el-input v-model="page.name" placeholder="例如：官网公告" />
                    </el-form-item>
                    <el-form-item label="排序">
                      <el-input-number v-model="page.sort" :min="0" :step="10" />
                    </el-form-item>
                  </div>

                  <el-form-item label="URL">
                    <el-input v-model="page.url" placeholder="https://example.com/news" />
                  </el-form-item>

                  <el-form-item label="网页说明">
                    <el-input
                      v-model="page.note"
                      type="textarea"
                      :rows="3"
                      placeholder="告诉 AI 这个网页要重点关注什么内容"
                    />
                  </el-form-item>

                  <el-form-item label="登录校验选择器">
                    <el-input v-model="page.login_check_selector" placeholder="例如：.avatar 或 .user-menu" />
                  </el-form-item>
                </el-form>

                <div class="page-card-actions">
                  <el-button type="primary" plain size="small" @click="savePage(page)">
                    <el-icon><Check /></el-icon>
                    保存网页
                  </el-button>
                  <el-button size="small" @click="openPageLogin(page)">
                    <el-icon><Position /></el-icon>
                    打开登录页
                  </el-button>
                  <el-button size="small" @click="checkPageLogin(page)">
                    <el-icon><Refresh /></el-icon>
                    检查登录
                  </el-button>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-if="runSubmitting || runWatching || runLiveLog" class="editor-card run-live-card">
          <div class="card-head">
            <div>
              <div class="card-title">执行过程</div>
              <div class="page-desc">{{ runLiveStatus }}</div>
              <div class="run-feedback-meta">{{ runFeedbackPhaseText }}</div>
              <div v-if="runFeedbackHint" class="run-feedback-hint">{{ runFeedbackHint }}</div>
            </div>
            <el-button text @click="clearRunLiveLog">清空</el-button>
          </div>
          <pre class="detail-pre live-pre">{{ runLiveLog || '等待执行日志...' }}</pre>
        </div>

        <div class="editor-card summary-card">
          <div class="card-head">
            <div class="card-title">最近执行结果</div>
            <el-button text @click="refreshRunList">刷新</el-button>
          </div>
          <el-empty v-if="runList.length === 0" description="暂无执行记录" />
          <div v-else class="run-list">
            <button
              v-for="item in runList.slice(0, 5)"
              :key="item.id"
              class="run-item"
              @click="openRunDetail(item.id)"
            >
              <div class="run-item-head">
                <span class="run-item-time">{{ item.create_time_desc || '-' }}</span>
                <el-tag
                  size="small"
                  effect="light"
                  :type="item.status === 'success' ? 'success' : item.status === 'partial_failed' ? 'warning' : item.status === 'running' ? 'info' : 'danger'"
                >
                  {{ item.status }}
                </el-tag>
              </div>
              <div class="run-item-desc">
                成功 {{ item.page_success_total || 0 }} / 失败 {{ item.page_failed_total || 0 }}
              </div>
            </button>
          </div>
        </div>
      </template>
    </section>

    <el-drawer v-model="historyDrawerVisible" title="执行历史" size="42%">
      <div class="drawer-toolbar">
        <el-button type="primary" plain @click="refreshRunList">刷新</el-button>
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
              :type="item.status === 'success' ? 'success' : item.status === 'partial_failed' ? 'warning' : item.status === 'running' ? 'info' : 'danger'"
            >
              {{ item.status }}
            </el-tag>
          </div>
          <div class="history-item-desc">{{ item.run_message || '无摘要' }}</div>
        </button>
      </div>
    </el-drawer>

    <el-dialog v-model="runDetailVisible" title="执行详情" width="76%" top="4vh" class="run-dialog">
      <div v-if="runDetail.run_info.id" class="run-detail">
        <div class="detail-section">
          <div class="detail-title">执行摘要</div>
          <div class="detail-meta">
            <el-tag size="small" effect="light">{{ runDetail.run_info.status }}</el-tag>
            <span>{{ runDetail.run_info.create_time_desc || '-' }}</span>
          </div>
          <pre class="detail-pre">{{ runDetail.run_info.summary_content || runDetail.run_info.run_message || '暂无汇总内容' }}</pre>
        </div>

        <div class="detail-section">
          <div class="detail-title">任务提示词快照</div>
          <pre class="detail-pre">{{ runDetail.run_info.prompt_snapshot || '-' }}</pre>
        </div>

        <div class="detail-section">
          <div class="detail-title">抓取计划</div>
          <pre class="detail-pre">{{ runDetail.run_info.planner_content || '-' }}</pre>
        </div>

        <div class="detail-section">
          <div class="detail-title">网页结果</div>
          <div class="result-page-list">
            <div v-for="item in runDetail.run_page_list" :key="item.id" class="result-page-card">
              <div class="result-page-head">
                <div>{{ item.page_name }}</div>
                <el-tag
                  size="small"
                  effect="light"
                  :type="item.status === 'success' ? 'success' : item.status === 'login_required' ? 'warning' : 'danger'"
                >
                  {{ item.status }}
                </el-tag>
              </div>
              <div class="result-page-sub">{{ item.planner_action || '-' }}</div>
              <pre class="detail-pre small">{{ item.raw_text || item.error_message || '无内容' }}</pre>
            </div>
          </div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import { Check, Clock, Delete, Plus, Position, Refresh, VideoPlay } from '@element-plus/icons-vue'
import AiSetApi from '@/utils/base/ai_set'
import InfoCrawlApi from '@/utils/base/info_crawl'
import sseDistribute from '@/utils/base/sse_distribute'

export default {
  name: 'InfoCrawl',
  components: {
    Check,
    Clock,
    Delete,
    Plus,
    Position,
    Refresh,
    VideoPlay,
  },
  data() {
    return {
      taskList: [],
      currentTaskId: 0,
      taskLoading: false,
      taskSaving: false,
      taskForm: {
        id: 0,
        name: '',
        prompt: '',
        ai_model_id: 0,
      },
      pageList: [],
      runList: [],
      aiModelList: [],
      historyDrawerVisible: false,
      runDetailVisible: false,
      runDetail: {
        run_info: {},
        run_page_list: [],
      },
      runSubmitting: false,
      runWatching: false,
      runningRunId: 0,
      runStatusTimer: null,
      tempPageIndex: 1,
      runLiveLog: '',
      runLiveStatus: '未开始执行',
      runSseDistributeId: '',
      runFeedbackPhase: 'idle',
      runFeedbackPhaseText: '未开始执行',
      runFeedbackHint: '',
      runFeedbackLastLogAt: 0,
      runFeedbackPhaseStartAt: 0,
      runFeedbackTicker: null,
    }
  },
  mounted() {
    this.loadAiModelList()
    this.loadTaskList()
  },
  beforeUnmount() {
    this.stopRunWatch()
    this.stopRunFeedbackTicker()
    this.unregisterRunSse()
  },
  methods: {
    // loadAiModelList 加载 AI 模型列表。
    loadAiModelList() {
      AiSetApi.AiModelList({}, (response) => {
        this.aiModelList = response && response.ErrCode === 0 && Array.isArray(response.Data)
          ? response.Data
          : []
      })
    },
    // loadTaskList 加载任务列表并默认打开第一项。
    loadTaskList() {
      this.taskLoading = true
      InfoCrawlApi.InfoCrawlTaskList((response) => {
        this.taskLoading = false
        const list = response && response.ErrCode === 0 && Array.isArray(response.Data?.task_list)
          ? response.Data.task_list
          : []
        this.taskList = list
        if (this.currentTaskId > 0) {
          const currentExist = list.find(item => item.id === this.currentTaskId)
          if (currentExist) {
            this.loadTaskDetail(this.currentTaskId)
            return
          }
        }
        if (list.length > 0) {
          this.selectTask(list[0].id)
        } else {
          this.resetTaskState()
        }
      })
    },
    // loadTaskDetail 加载任务详情。
    loadTaskDetail(taskId) {
      if (!taskId) {
        this.resetTaskState()
        return
      }
      this.taskLoading = true
      InfoCrawlApi.InfoCrawlTaskInfo(taskId, (response) => {
        this.taskLoading = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          return
        }
        const task = response.Data.task || {}
        this.currentTaskId = task.id || 0
        this.taskForm = {
          id: task.id || 0,
          name: task.name || '',
          prompt: task.prompt || '',
          ai_model_id: task.ai_model_id || 0,
        }
        this.pageList = Array.isArray(response.Data.page_list) ? response.Data.page_list.map(this.normalizePage) : []
        this.runList = Array.isArray(response.Data.run_list) ? response.Data.run_list : []
      })
    },
    // resetTaskState 清空页面状态。
    resetTaskState() {
      this.currentTaskId = 0
      this.taskForm = {
        id: 0,
        name: '',
        prompt: '',
        ai_model_id: 0,
      }
      this.pageList = []
      this.runList = []
      this.runDetail = {
        run_info: {},
        run_page_list: [],
      }
      this.stopRunWatch()
      this.runLiveLog = ''
      this.runLiveStatus = '未开始执行'
      this.resetRunFeedback()
      this.unregisterRunSse()
    },
    // resetRunFeedback 重置执行反馈状态。
    resetRunFeedback() {
      this.runFeedbackPhase = 'idle'
      this.runFeedbackPhaseText = '未开始执行'
      this.runFeedbackHint = ''
      this.runFeedbackLastLogAt = 0
      this.runFeedbackPhaseStartAt = 0
      this.stopRunFeedbackTicker()
    },
    // setRunFeedbackPhase 设置当前执行阶段。
    setRunFeedbackPhase(phase, phaseText) {
      const now = Date.now()
      if (this.runFeedbackPhase !== phase || !this.runFeedbackPhaseStartAt) {
        this.runFeedbackPhaseStartAt = now
      }
      this.runFeedbackPhase = phase
      this.runFeedbackPhaseText = phaseText
      if (!this.runFeedbackLastLogAt) {
        this.runFeedbackLastLogAt = now
      }
      this.refreshRunFeedbackHint()
    },
    // touchRunFeedback 记录最近一次日志时间。
    touchRunFeedback() {
      this.runFeedbackLastLogAt = Date.now()
      if (!this.runFeedbackPhaseStartAt) {
        this.runFeedbackPhaseStartAt = this.runFeedbackLastLogAt
      }
      this.refreshRunFeedbackHint()
    },
    // startRunFeedbackTicker 启动执行反馈刷新定时器。
    startRunFeedbackTicker() {
      this.stopRunFeedbackTicker()
      this.refreshRunFeedbackHint()
      this.runFeedbackTicker = window.setInterval(() => {
        this.refreshRunFeedbackHint()
      }, 1000)
    },
    // stopRunFeedbackTicker 停止执行反馈刷新定时器。
    stopRunFeedbackTicker() {
      if (this.runFeedbackTicker) {
        window.clearInterval(this.runFeedbackTicker)
        this.runFeedbackTicker = null
      }
    },
    // formatRunFeedbackSeconds 将秒数转换成中文时长。
    formatRunFeedbackSeconds(seconds) {
      if (seconds < 60) {
        return `${seconds} 秒`
      }
      const minute = Math.floor(seconds / 60)
      const remainSecond = seconds % 60
      if (remainSecond === 0) {
        return `${minute} 分钟`
      }
      return `${minute} 分 ${remainSecond} 秒`
    },
    // refreshRunFeedbackHint 刷新执行提示文案。
    refreshRunFeedbackHint() {
      if (!this.runFeedbackPhaseStartAt || this.runFeedbackPhase === 'idle') {
        this.runFeedbackHint = ''
        return
      }
      const now = Date.now()
      const phaseSeconds = Math.max(1, Math.floor((now - this.runFeedbackPhaseStartAt) / 1000))
      const lastLogSeconds = this.runFeedbackLastLogAt
        ? Math.max(0, Math.floor((now - this.runFeedbackLastLogAt) / 1000))
        : 0
      const durationText = `本阶段已持续 ${this.formatRunFeedbackSeconds(phaseSeconds)}`
      const lastLogText = this.runFeedbackLastLogAt
        ? `最近更新于 ${this.formatRunFeedbackSeconds(lastLogSeconds)}前`
        : '等待首条执行日志'
      if (this.runFeedbackPhase === 'planning' && lastLogSeconds >= 6) {
        this.runFeedbackHint = `${durationText}，${lastLogText}，规划处理中，暂未返回新日志，任务仍在后台执行`
        return
      }
      if (this.runFeedbackPhase === 'finished' || this.runFeedbackPhase === 'failed') {
        this.runFeedbackHint = durationText
        return
      }
      this.runFeedbackHint = `${durationText}，${lastLogText}`
    },
    // updateRunFeedbackByLog 根据日志前缀切换执行阶段。
    updateRunFeedbackByLog(msg, msgType) {
      if (msgType === 'error' && msg.includes('[任务]')) {
        this.setRunFeedbackPhase('failed', '任务执行失败')
        return
      }
      if (msg.includes('[规划] 正在生成抓取计划')) {
        this.setRunFeedbackPhase('planning', '正在生成抓取计划，请稍候')
        return
      }
      if (msg.includes('[规划] 计划生成完成') || msg.includes('[执行]')) {
        this.setRunFeedbackPhase('crawling', '抓取计划已生成，正在按网页执行')
        return
      }
      if (msg.includes('[汇总] 正在生成总结')) {
        this.setRunFeedbackPhase('summarizing', '网页抓取完成，正在生成总结')
        return
      }
      if (msg.includes('[任务] 执行完成')) {
        this.setRunFeedbackPhase('finished', '任务执行完成')
      }
    },
    // normalizePage 统一网页对象结构。
    normalizePage(page) {
      return {
        id: page.id || 0,
        _temp_id: page._temp_id || '',
        task_id: page.task_id || this.currentTaskId,
        name: page.name || '',
        url: page.url || '',
        note: page.note || '',
        login_check_selector: page.login_check_selector || '',
        login_status: page.login_status || 0,
        login_status_desc: page.login_status_desc || '未登录',
        sort: page.sort || 0,
        user_data_dir: page.user_data_dir || '',
      }
    },
    // selectTask 切换当前任务。
    selectTask(taskId) {
      this.currentTaskId = taskId
      this.loadTaskDetail(taskId)
    },
    // createTask 创建默认任务。
    createTask() {
      if (this.aiModelList.length === 0) {
        this.$helperNotify.error('请先在配置中创建 AI 模型')
        return
      }
      const defaultModelId = this.aiModelList.length > 0 ? this.aiModelList[0].id : 0
      InfoCrawlApi.InfoCrawlTaskSave({
        id: 0,
        name: '新信息抓取任务',
        prompt: '请重点整理关键信息并输出中文摘要。',
        ai_model_id: defaultModelId,
      }, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          return
        }
        this.loadTaskList()
        this.currentTaskId = response.Data.id || 0
        this.loadTaskDetail(this.currentTaskId)
      })
    },
    // saveTask 保存当前任务。
    saveTask() {
      if (this.taskForm.name.trim() === '') {
        this.$helperNotify.error('任务名称不能为空')
        return
      }
      if (this.taskForm.prompt.trim() === '') {
        this.$helperNotify.error('任务提示词不能为空')
        return
      }
      if (!this.taskForm.ai_model_id) {
        this.$helperNotify.error('请选择 AI 模型')
        return
      }
      this.taskSaving = true
      InfoCrawlApi.InfoCrawlTaskSave(this.taskForm, (response) => {
        this.taskSaving = false
        if (!(response && response.ErrCode === 0)) {
          return
        }
        this.$helperNotify.success('保存成功')
        this.loadTaskList()
        if (response.Data?.id) {
          this.loadTaskDetail(response.Data.id)
        }
      })
    },
    // deleteTask 删除当前任务。
    deleteTask() {
      if (!this.currentTaskId) {
        return
      }
      InfoCrawlApi.InfoCrawlTaskDelete(this.currentTaskId, (response) => {
        if (!(response && response.ErrCode === 0)) {
          return
        }
        this.$helperNotify.success('任务已删除')
        this.loadTaskList()
      })
    },
    // appendPage 追加一个本地网页配置。
    appendPage() {
      this.pageList.push(this.normalizePage({
        _temp_id: `temp-${this.tempPageIndex++}`,
        task_id: this.currentTaskId,
        sort: this.pageList.length * 10 + 10,
      }))
    },
    // savePage 保存单个网页配置。
    savePage(page) {
      if (!this.currentTaskId) {
        this.$helperNotify.error('请先保存任务')
        return
      }
      if ((page.name || '').trim() === '') {
        this.$helperNotify.error('网页名称不能为空')
        return
      }
      if ((page.url || '').trim() === '') {
        this.$helperNotify.error('网页URL不能为空')
        return
      }
      const payload = {
        id: page.id || 0,
        task_id: this.currentTaskId,
        name: page.name,
        url: page.url,
        note: page.note,
        login_check_selector: page.login_check_selector,
        sort: page.sort || 0,
      }
      InfoCrawlApi.InfoCrawlTaskPageSave(payload, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          return
        }
        this.$helperNotify.success('网页已保存')
        this.loadTaskDetail(this.currentTaskId)
      })
    },
    // deletePage 删除单个网页配置。
    deletePage(page) {
      if (!page.id) {
        this.pageList = this.pageList.filter(item => item._temp_id !== page._temp_id)
        return
      }
      InfoCrawlApi.InfoCrawlTaskPageDelete(page.id, (response) => {
        if (!(response && response.ErrCode === 0)) {
          return
        }
        this.$helperNotify.success('网页已删除')
        this.loadTaskDetail(this.currentTaskId)
      })
    },
    // openPageLogin 打开网页登录窗口。
    openPageLogin(page) {
      if (!page.id) {
        this.$helperNotify.error('请先保存网页配置')
        return
      }
      InfoCrawlApi.InfoCrawlTaskPageOpenLogin(page.id, (response) => {
        if (response && response.ErrCode === 0) {
          this.$helperNotify.success(response.ErrMsg || '请在浏览器中完成登录')
        }
      })
    },
    // checkPageLogin 检查网页登录状态。
    checkPageLogin(page) {
      if (!page.id) {
        this.$helperNotify.error('请先保存网页配置')
        return
      }
      InfoCrawlApi.InfoCrawlTaskPageCheckLogin(page.id, (response) => {
        if (response && response.ErrCode === 0) {
          this.$helperNotify.success('登录状态正常')
          this.loadTaskDetail(this.currentTaskId)
        }
      })
    },
    // runTask 执行当前任务。
    runTask() {
      if (!this.currentTaskId) {
        return
      }
      if (this.runWatching) {
        this.$helperNotify.error('当前已有任务在后台执行，请等待完成后再发起')
        return
      }
      if (this.pageList.length === 0) {
        this.$helperNotify.error('至少需要一个网页配置')
        return
      }
      const sseDistributeId = sseDistribute.GetSseDistributeId(`info_crawl_${this.currentTaskId}`)
      this.unregisterRunSse()
      this.runSseDistributeId = sseDistributeId
      this.runLiveLog = ''
      this.runLiveStatus = '准备提交任务...'
      this.resetRunFeedback()
      this.setRunFeedbackPhase('submitting', '正在提交任务')
      this.startRunFeedbackTicker()
      this.registerRunSse(sseDistributeId)
      this.runSubmitting = true
      InfoCrawlApi.InfoCrawlTaskRun({
        task_id: this.currentTaskId,
        sse_distribute_id: sseDistributeId,
      }, (response) => {
        this.runSubmitting = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.runLiveStatus = '任务提交失败'
          this.setRunFeedbackPhase('failed', '任务提交失败')
          this.stopRunFeedbackTicker()
          this.unregisterRunSse()
          return
        }
        this.runningRunId = response.Data.run_id || 0
        this.runWatching = true
        this.runLiveStatus = '任务已提交，正在等待抓取计划'
        this.setRunFeedbackPhase('planning', '正在生成抓取计划，请稍候')
        this.refreshRunList()
        this.startRunWatch()
        this.$helperNotify.success('任务已提交，正在后台执行')
      })
    },
    // startRunWatch 开始轮询后台执行状态。
    startRunWatch() {
      this.stopRunWatch()
      if (!this.runningRunId) {
        return
      }
      this.fetchRunStatus()
      this.runStatusTimer = window.setInterval(() => {
        this.fetchRunStatus()
      }, 3000)
    },
    // stopRunWatch 停止轮询后台执行状态。
    stopRunWatch() {
      if (this.runStatusTimer) {
        window.clearInterval(this.runStatusTimer)
        this.runStatusTimer = null
      }
      this.runWatching = false
      this.runningRunId = 0
      this.stopRunFeedbackTicker()
    },
    // fetchRunStatus 查询当前执行任务状态。
    fetchRunStatus() {
      if (!this.runningRunId) {
        this.stopRunWatch()
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
          this.runWatching = true
          this.runLiveStatus = '任务仍在后台执行'
          if (this.runFeedbackPhase === 'submitting' || this.runFeedbackPhase === 'idle') {
            this.setRunFeedbackPhase('planning', '正在生成抓取计划，请稍候')
          }
          return
        }
        this.runLiveStatus = `任务执行完成，状态：${runInfo.status || '-'}`
        if (runInfo.status === 'success' || runInfo.status === 'partial_failed') {
          this.setRunFeedbackPhase('finished', `任务执行完成，状态：${runInfo.status || '-'}`)
        } else {
          this.setRunFeedbackPhase('failed', `任务执行结束，状态：${runInfo.status || '-'}`)
        }
        this.stopRunWatch()
        this.unregisterRunSse()
        this.refreshRunList()
        this.loadTaskDetail(this.currentTaskId)
      })
    },
    // registerRunSse 注册执行过程 SSE 输出。
    registerRunSse(sseDistributeId) {
      if (!sseDistributeId) {
        return
      }
      sseDistribute.RegisterReceive(sseDistributeId, (msg, msgType) => {
        if (typeof msg !== 'string' || msg.trim() === '') {
          return
        }
        this.touchRunFeedback()
        this.updateRunFeedbackByLog(msg, msgType)
        this.runLiveStatus = '任务执行中'
        const prefix = msgType === 'error' ? '[错误] ' : ''
        this.runLiveLog += `${prefix}${msg}`.replace(/\r/g, '')
      })
    },
    // unregisterRunSse 注销执行过程 SSE 输出。
    unregisterRunSse() {
      if (!this.runSseDistributeId) {
        return
      }
      sseDistribute.UnRegisterReceive(this.runSseDistributeId)
      this.runSseDistributeId = ''
    },
    // clearRunLiveLog 清空执行过程输出。
    clearRunLiveLog() {
      this.runLiveLog = ''
      if (!this.runSubmitting && !this.runWatching) {
        this.runLiveStatus = '未开始执行'
        this.resetRunFeedback()
      }
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
    // openHistoryDrawer 打开执行历史抽屉。
    openHistoryDrawer() {
      this.historyDrawerVisible = true
      this.refreshRunList()
    },
    // openRunDetail 打开执行详情。
    openRunDetail(runId) {
      InfoCrawlApi.InfoCrawlRunInfo(runId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          return
        }
        this.runDetail = response.Data
        this.runDetailVisible = true
      })
    },
  },
}
</script>

<style scoped>
.info-crawl-page {
  display: flex;
  gap: 14px;
  min-height: calc(100vh - 40px);
}

.info-crawl-sidebar,
.editor-card {
  border: 1px solid #e8e8e0;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 4px 12px rgba(54, 74, 54, 0.05);
}

.info-crawl-sidebar {
  width: 300px;
  flex-shrink: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.sidebar-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid #ecece4;
  background: #f7f7f2;
}

.sidebar-title,
.page-title,
.card-title {
  font-size: 16px;
  font-weight: 700;
  color: #3e5140;
}

.sidebar-desc,
.page-desc {
  margin-top: 4px;
  color: #748172;
  font-size: 12px;
}

.sidebar-scroll {
  flex: 1;
}

.task-item {
  width: calc(100% - 16px);
  margin: 8px;
  padding: 12px;
  border: 1px solid #edf1e8;
  border-radius: 12px;
  background: #fbfcf8;
  cursor: pointer;
  text-align: left;
  transition: all 0.2s ease;
}

.task-item:hover,
.task-item.active {
  border-color: #cfe0c8;
  background: #f2f8ec;
}

.task-item-title {
  color: #314233;
  font-size: 14px;
  font-weight: 600;
}

.task-item-time {
  margin-top: 6px;
  color: #7c8576;
  font-size: 12px;
}

.info-crawl-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 14px;
}

.empty-wrap {
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
  border: 1px dashed #d9dfd5;
  border-radius: 16px;
  background: linear-gradient(135deg, #f7fbf4 0%, #ffffff 100%);
}

.page-toolbar,
.drawer-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 12px;
  padding: 18px 20px;
  border: 1px solid #e8e8e0;
  border-radius: 14px;
  background: linear-gradient(135deg, #f6fbf4 0%, #ffffff 100%);
}

.toolbar-actions,
.card-head,
.page-card-actions,
.detail-meta {
  display: flex;
  align-items: center;
  gap: 10px;
  flex-wrap: wrap;
}

.editor-grid {
  display: grid;
  grid-template-columns: minmax(320px, 1fr) minmax(420px, 1.3fr);
  gap: 14px;
}

.editor-card {
  padding: 18px;
}

.card-head {
  justify-content: space-between;
  margin-bottom: 14px;
}

.page-card-list,
.run-list,
.history-list,
.result-page-list {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.page-card,
.run-item,
.history-item,
.result-page-card {
  border: 1px solid #e8eee3;
  border-radius: 14px;
  background: #fbfcf8;
}

.page-card {
  padding: 14px;
}

.page-card-head,
.run-item-head,
.history-item-head,
.result-page-head {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 12px;
}

.page-card-title {
  color: #314233;
  font-size: 14px;
  font-weight: 700;
}

.page-card-status {
  margin-top: 6px;
}

.page-form {
  margin-top: 12px;
}

.form-row {
  display: grid;
  grid-template-columns: 1fr 160px;
  gap: 12px;
}

.page-card-actions {
  margin-top: 10px;
}

.summary-card {
  min-height: 180px;
}

.run-live-card {
  min-height: 180px;
}

.run-item,
.history-item {
  width: 100%;
  padding: 14px 16px;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s ease;
}

.run-item:hover,
.history-item:hover {
  border-color: #cfe0c8;
  background: #f4f9ee;
}

.run-item-desc,
.history-item-desc,
.result-page-sub {
  margin-top: 8px;
  color: #60705a;
  font-size: 13px;
  line-height: 1.6;
}

.run-detail {
  display: flex;
  flex-direction: column;
  gap: 16px;
  max-height: 76vh;
  overflow: auto;
}

.detail-section {
  border: 1px solid #e8eee3;
  border-radius: 14px;
  background: #fbfcf8;
  padding: 16px;
}

.detail-title {
  color: #3e5140;
  font-size: 15px;
  font-weight: 700;
  margin-bottom: 10px;
}

.detail-pre {
  margin: 0;
  padding: 14px;
  border-radius: 12px;
  background: #f3f6ef;
  color: #334335;
  font-size: 13px;
  line-height: 1.7;
  white-space: pre-wrap;
  word-break: break-word;
}

.live-pre {
  max-height: 260px;
  overflow: auto;
}

.run-feedback-meta {
  margin-top: 6px;
  color: #4f6250;
  font-size: 13px;
  line-height: 1.6;
}

.run-feedback-hint {
  margin-top: 6px;
  color: #7a866f;
  font-size: 12px;
  line-height: 1.7;
}

.detail-pre.small {
  margin-top: 10px;
  max-height: 260px;
  overflow: auto;
}

.empty-inline {
  padding: 10px 0;
}

@media (max-width: 1180px) {
  .info-crawl-page {
    flex-direction: column;
  }

  .info-crawl-sidebar {
    width: 100%;
  }

  .page-toolbar,
  .toolbar-actions,
  .editor-grid,
  .form-row {
    display: flex;
    flex-direction: column;
  }
}
</style>
