<template>
  <div class="task-workflow-page" v-loading="loading">
    <div class="task-workflow-shell">
      <header class="task-workflow-header">
        <div class="task-workflow-header__main">
          <div class="task-workflow-header__eyebrow">任务工作流程</div>
          <h1 class="task-workflow-header__title">{{ homeTask.name || `任务 #${taskId}` }}</h1>
        </div>
        <div class="task-workflow-header__actions">
          <el-tooltip content="返回首页" placement="bottom">
            <el-button class="task-workflow-home-btn" @click="goHome">
              <el-icon :size="18"><HomeFilled /></el-icon>
            </el-button>
          </el-tooltip>
          <GitActionButton compact variant="info" @click="goBackToTaskList">
            返回任务清单
          </GitActionButton>
          <GitActionButton compact :loading="loading" @click="reloadWorkflowPage">
            刷新
          </GitActionButton>
        </div>
      </header>

      <el-alert
        v-if="errorMessage"
        type="error"
        :closable="false"
        :title="errorMessage"
        class="task-workflow-alert"
      />

      <section class="task-workflow-nodes">
        <button
          v-for="node in workflowNodes"
          :key="node.key"
          type="button"
          class="task-workflow-node"
          :class="{
            'task-workflow-node--active': activeNode === node.key,
            'task-workflow-node--success': node.key === 'requirement-fetch' && activeNode === node.key && requirementFetchStatus === 'success',
            'task-workflow-node--failed': node.key === 'requirement-fetch' && activeNode === node.key && requirementFetchStatus === 'failed',
            'task-workflow-node--running': node.key === 'requirement-fetch' && activeNode === node.key && requirementFetchStatus === 'running',
          }"
          @click="activeNode = node.key"
        >
          <span class="task-workflow-node__label">{{ node.label }}</span>
          <span class="task-workflow-node__desc">{{ node.desc }}</span>
        </button>
      </section>

      <section class="task-workflow-content">
        <div v-if="activeNode === 'task-config'" class="task-workflow-tab">
          <div class="task-workflow-card task-workflow-config-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">任务配置</div>
            </div>
            <div class="task-workflow-config-content">
              <div class="task-workflow-config-section">
                <div class="task-workflow-config-section__title">基本信息</div>
                <el-descriptions :column="2" border size="small">
                  <el-descriptions-item label="任务名称">{{ homeTask.name || '-' }}</el-descriptions-item>
                  <el-descriptions-item label="任务状态">
                    <el-tag size="small" effect="light" :type="getTaskStatusTagType(homeTask.task_status)">{{ homeTask.task_status || '-' }}</el-tag>
                    <el-dropdown trigger="click" @command="handleTaskStatusChange" style="margin-left: 8px;">
                      <el-button size="small" :loading="statusUpdating" text type="primary">
                        切换状态
                      </el-button>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item
                            v-for="status in taskStatusOptions"
                            :key="status"
                            :command="status"
                            :disabled="homeTask.task_status === status"
                          >
                            {{ status }}
                          </el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                    </el-dropdown>
                  </el-descriptions-item>
                  <el-descriptions-item label="开始日期">{{ homeTask.start_time_desc || '-' }}</el-descriptions-item>
                  <el-descriptions-item label="TAPD地址">
                    <a v-if="homeTask.tapd_url" :href="homeTask.tapd_url" target="_blank" class="task-workflow-config-link">{{ homeTask.tapd_url }}</a>
                    <span v-else>-</span>
                  </el-descriptions-item>
                  <el-descriptions-item label="使用工作流程">{{ Number(homeTask.use_workflow || 0) === 1 ? '是' : '否' }}</el-descriptions-item>
                  <el-descriptions-item label="最后操作">{{ homeTask.last_operated_at_desc || '-' }}</el-descriptions-item>
                </el-descriptions>
              </div>
              <div v-if="parsedTaskDevConfigs.length > 0" class="task-workflow-config-section">
                <div class="task-workflow-config-section__title">开发项目配置</div>
                <div v-for="(cfg, idx) in parsedTaskDevConfigs" :key="idx" class="task-workflow-config-dev">
                  <div v-if="parsedTaskDevConfigs.length > 1" class="task-workflow-config-dev__index">配置 #{{ idx + 1 }}</div>
                  <el-descriptions :column="2" border size="small">
                    <el-descriptions-item label="Git仓库">{{ getTaskConfigName('git', cfg.git_id) }}</el-descriptions-item>
                    <el-descriptions-item label="Docker">{{ getTaskConfigName('docker', cfg.docker_id) }}</el-descriptions-item>
                    <el-descriptions-item label="Db">{{ getTaskConfigName('mysql', cfg.mysql_id) }}</el-descriptions-item>
                    <el-descriptions-item label="接口集合">{{ getTaskConfigApiLabel(cfg) }}</el-descriptions-item>
                    <el-descriptions-item label="本地目录">{{ cfg.local_dir || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="父分支">{{ cfg.parent_branch || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="分支名">{{ cfg.branch_name || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="规则入口">{{ cfg.rule_entry_file || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="自定义网页">{{ getTaskConfigName('smart_link', cfg.smart_link_id) }}</el-descriptions-item>
                    <el-descriptions-item label="网页标签">{{ cfg.smart_link_label || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="账号">{{ cfg.smart_link_account || '-' }}</el-descriptions-item>
                  </el-descriptions>
                </div>
              </div>
            </div>
          </div>
        </div>

        <div v-else-if="activeNode === 'requirement-fetch'" class="task-workflow-tab">
          <div class="task-workflow-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">抓取 TAPD 需求</div>
              <div class="task-workflow-card__switch">
                <div class="task-workflow-inner-tabs">
                  <button
                    :class="['task-workflow-inner-tab', { 'task-workflow-inner-tab--active': requirementFetchActiveTab === 'tapd-fetch' }]"
                    @click="requirementFetchActiveTab = 'tapd-fetch'"
                  >抓取 TAPD 需求内容</button>
                  <button
                    :class="['task-workflow-inner-tab', { 'task-workflow-inner-tab--active': requirementFetchActiveTab === 'plain-text-prompt' }]"
                    @click="requirementFetchActiveTab = 'plain-text-prompt'"
                  >纯文本需求提示词</button>
                </div>
              </div>
            </div>

            <div v-show="requirementFetchActiveTab === 'tapd-fetch'" class="task-workflow-tapd-fetch-section">
              <div class="task-workflow-card__switch" style="margin-bottom: 12px;">
                <GitActionButton compact :loading="requirementFetchRunning" @click="triggerRequirementFetch(false)">
                  重新抓取
                </GitActionButton>
                <GitActionButton compact variant="info" @click="openRequirementFragment" :disabled="!requirementFragmentId">
                  打开知识片段
                </GitActionButton>
              </div>
              <div v-if="workflow.requirement_fetch_error" class="task-workflow-card__hint task-workflow-card__hint--error">
                最近错误：{{ workflow.requirement_fetch_error }}
              </div>
              <div v-if="!homeTask.tapd_url" class="task-workflow-card__hint">
                当前任务未配置 TAPD 地址，无法自动抓取。
              </div>
              <div class="task-workflow-fragment-view">
                <iframe
                  v-if="requirementShareUrl"
                  :src="requirementShareUrl"
                  class="task-workflow-fragment-view__iframe"
                  title="需求知识片段预览"
                />
                <div v-else class="task-workflow-fragment-view__empty">
                  知识片段分享链接生成中...
                </div>
              </div>
            </div>

            <div v-show="requirementFetchActiveTab === 'plain-text-prompt'" class="task-workflow-prompt-section">
              <div class="task-workflow-card__switch" style="margin-bottom: 12px;">
                <GitActionButton compact :loading="promptSaving === 'plain_text_requirement'" @click="savePrompts('plain_text_requirement')">
                  保存提示词
                </GitActionButton>
                <GitActionButton compact @click="copyText(workflow.prompt_plain_text_requirement || '', '提示词已复制')">
                  复制提示词
                </GitActionButton>
                <GitActionButton compact variant="warning" :loading="promptRestoring === 'plain_text_requirement'" @click="restorePrompts('plain_text_requirement')">
                  还原为默认提示词
                </GitActionButton>
                <GitActionButton compact variant="info" @click="openPlainTextReqFragment" :disabled="!plainTextReqFragmentId">
                  打开知识片段
                </GitActionButton>
              </div>
              <MdEditor
                v-model="workflow.prompt_plain_text_requirement"
                class="task-workflow-prompt-editor"
                preview-theme="github"
                :preview="true"
                :toolbars="promptEditorToolbars"
                height="100%"
              />
            </div>
          </div>
        </div>

        <div v-else-if="activeNode === 'requirement'" class="task-workflow-tab">
          <div class="task-workflow-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">需求文档提示词</div>
              <div class="task-workflow-card__switch">
                <GitActionButton compact variant="info" @click="openRequirementFragment" :disabled="!requirementFragmentId">
                  打开知识片段
                </GitActionButton>
                <GitActionButton compact :loading="promptSaving === 'requirement'" @click="savePrompts('requirement')">
                  保存提示词
                </GitActionButton>
                <GitActionButton compact @click="copyText(workflow.prompt_requirement || '', '提示词已复制')">
                  复制提示词
                </GitActionButton>
                <GitActionButton compact variant="warning" :loading="promptRestoring === 'requirement'" @click="restorePrompts('requirement')">
                  还原为默认提示词
                </GitActionButton>
              </div>
            </div>
            <div class="task-workflow-card__hint">
              当前片段：{{ requirementFragmentTitle }}
            </div>
            <MdEditor
              v-model="workflow.prompt_requirement"
              class="task-workflow-prompt-editor"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              height="100%"
            />
          </div>
        </div>

        <div v-else-if="activeNode === 'design'" class="task-workflow-tab">
          <div class="task-workflow-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">开发设计提示词</div>
              <div class="task-workflow-card__switch">
                <GitActionButton compact :loading="promptSaving === 'design'" @click="savePrompts('design')">
                  保存提示词
                </GitActionButton>
                <GitActionButton compact @click="copyText(workflow.prompt_design || '', '提示词已复制')">
                  复制提示词
                </GitActionButton>
                <GitActionButton compact variant="warning" :loading="promptRestoring === 'design'" @click="restorePrompts('design')">
                  还原为默认提示词
                </GitActionButton>
              </div>
            </div>
            <MdEditor
              v-model="workflow.prompt_design"
              class="task-workflow-prompt-editor"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              height="100%"
            />
          </div>
        </div>

        <div v-else-if="activeNode === 'api-dev'" class="task-workflow-tab">
          <div class="task-workflow-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">接口开发生成提示词</div>
              <div class="task-workflow-card__switch">
                <GitActionButton compact :loading="promptSaving === 'api_dev'" @click="savePrompts('api_dev')">
                  保存提示词
                </GitActionButton>
                <GitActionButton compact @click="copyText(workflow.prompt_api_dev || '', '提示词已复制')">
                  复制提示词
                </GitActionButton>
                <GitActionButton compact variant="warning" :loading="promptRestoring === 'api_dev'" @click="restorePrompts('api_dev')">
                  还原为默认提示词
                </GitActionButton>
              </div>
            </div>
            <MdEditor
              v-model="workflow.prompt_api_dev"
              class="task-workflow-prompt-editor"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              height="100%"
            />
          </div>
        </div>

        <div v-else class="task-workflow-tab">
          <div class="task-workflow-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">接口自动化测试修复提示词</div>
              <div class="task-workflow-card__switch">
                <GitActionButton compact :loading="promptSaving === 'api_test'" @click="savePrompts('api_test')">
                  保存提示词
                </GitActionButton>
                <GitActionButton compact @click="copyText(workflow.prompt_api_test || '', '提示词已复制')">
                  复制提示词
                </GitActionButton>
                <GitActionButton compact variant="warning" :loading="promptRestoring === 'api_test'" @click="restorePrompts('api_test')">
                  还原为默认提示词
                </GitActionButton>
              </div>
            </div>
            <MdEditor
              v-model="workflow.prompt_api_test"
              class="task-workflow-prompt-editor"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              height="100%"
            />
          </div>
        </div>
      </section>
    </div>
  </div>
</template>

<script>
import { HomeFilled } from '@element-plus/icons-vue'
import GitActionButton from '@/components/base/GitActionButton.vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'
import taskWorkflowApi from '@/utils/base/task_workflow'
import homeTaskApi from '@/utils/base/home_task'
import baseUtils from '@/utils/base'
import sseDistribute from '@/utils/base/sse_distribute'
import gitApi from '@/utils/base/git'
import mysqlSetApi from '@/utils/base/mysql_set'
import apiManagement from '@/utils/base/api'
import dockerApi from '@/utils/base/compose'
import smartLinkSetApi from '@/utils/base/smart_link_set'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

const PROMPT_EDITOR_TOOLBARS = [
  'bold', 'italic', 'strikeThrough', 'title', 'quote',
  'unorderedList', 'orderedList', 'task', 'link', 'code',
  'codeRow', 'table', 'preview', 'fullscreen',
]

const TASK_STATUS_TODO = '待开始'
const TASK_STATUS_DEVELOPING = '开发中'
const TASK_STATUS_SELF_TESTING = '自测中'
const TASK_STATUS_SELF_TESTED = '自测完'
const TASK_STATUS_PENDING_INTEGRATION = '待对接'
const TASK_STATUS_INTEGRATING = '对接中'
const TASK_STATUS_TESTING = '测试中'
const TASK_STATUS_RELEASING = '上线中'
const TASK_STATUS_ONLINE = '已上线'
const TASK_STATUS_OPTIONS = [
  TASK_STATUS_TODO,
  TASK_STATUS_DEVELOPING,
  TASK_STATUS_SELF_TESTING,
  TASK_STATUS_SELF_TESTED,
  TASK_STATUS_PENDING_INTEGRATION,
  TASK_STATUS_INTEGRATING,
  TASK_STATUS_TESTING,
  TASK_STATUS_RELEASING,
  TASK_STATUS_ONLINE,
]

const WORKFLOW_NODES = [
  { key: 'task-config', label: '任务配置', desc: '查看当前任务的所有配置信息' },
  { key: 'requirement-fetch', label: '1.抓取TAPD需求', desc: '自动登录和解析tapd需求到知识片段，转为markdown格式供AI解析' },
  { key: 'requirement', label: '2.需求分析', desc: '编写提示词，AI自动结合数据库和代码分析需求，形成开发文档' },
  { key: 'design', label: '3.开发执行', desc: '编写提示词，AI自动结合数据库，代码和开发文档进行开发' },
  { key: 'api-dev', label: '4.接口生成', desc: '编写提示词，AI自动获取登录态，将所有改动接口写入接口开发中' },
  { key: 'api-test-fix', label: '5.自动化测试+修复', desc: '编写提示词，AI自动根据接口开发中的接口设计测试流程，自动上传代码+自动重启服务+自动修复BUG，支持多项目联调' },
]

export default {
  name: 'TaskWorkflow',
  components: {
    HomeFilled,
    GitActionButton,
    MdEditor,
  },
  data() {
    return {
      workflowNodes: WORKFLOW_NODES,
      activeNode: 'requirement-fetch',
      loading: false,
      errorMessage: '',
      workflowId: 0,
      workflow: {},
      homeTask: {},
      requirementFragment: {},
      requirementShareUrl: '',
      requirementFetchConfig: {},
      requirementFetchLogs: [],
      requirementFetchRunning: false,
      requirementFetchAutoTriggered: false,
      workflowSseDistributeId: '',
      promptSaving: '',
      promptRestoring: '',
      requirementFetchActiveTab: 'tapd-fetch',
      promptEditorToolbars: PROMPT_EDITOR_TOOLBARS,
      taskStatusOptions: TASK_STATUS_OPTIONS,
      statusUpdating: false,
      taskConfigGitRepoList: [],
      taskConfigDockerList: [],
      taskConfigMysqlList: [],
      taskConfigCollectionList: [],
      taskConfigSmartLinkList: [],
      taskConfigApiFolderMap: {},
    }
  },
  computed: {
    taskId() {
      return Number(this.$route.params.taskId || 0)
    },
    requirementFetchStatus() {
      return String(this.workflow.requirement_fetch_status || 'idle').trim() || 'idle'
    },
    requirementFetchStatusText() {
      const map = {
        idle: '待执行',
        running: '执行中',
        success: '已完成',
        failed: '执行失败',
      }
      return map[this.requirementFetchStatus] || this.requirementFetchStatus
    },
    requirementFragmentId() {
      return String(this.workflow.requirement_fragment_id || '').trim()
    },
    plainTextReqFragmentId() {
      return String(this.workflow.plain_text_requirement_fragment_id || '').trim()
    },
    requirementFragmentTitle() {
      return String(this.requirementFragment.title || '').trim() || (this.requirementFragmentId ? `#${this.requirementFragmentId}` : '-')
    },
    parsedTaskDevConfigs() {
      const raw = this.homeTask.dev_configs
      if (!raw) return []
      try {
        const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
        return Array.isArray(parsed) ? parsed : []
      } catch {
        return []
      }
    },
  },
  mounted() {
    this.loadWorkflowPage()
    this.loadTaskConfigLookupData()
    window.addEventListener('keydown', this.handleCtrlS)
  },
  beforeUnmount() {
    window.removeEventListener('keydown', this.handleCtrlS)
    this.unregisterWorkflowSse()
  },
  watch: {
    parsedTaskDevConfigs: {
      handler(configs) {
        for (const cfg of configs) {
          const colId = Number(cfg.collection_id || 0)
          if (colId > 0) {
            this.loadTaskConfigApiFoldersForCollection(colId)
          }
        }
      },
      immediate: true,
    },
    '$route.params.taskId'() {
      this.requirementFetchAutoTriggered = false
      this.requirementFetchLogs = []
      this.activeNode = 'requirement-fetch'
      this.requirementFetchActiveTab = 'tapd-fetch'
      this.unregisterWorkflowSse()
      this.loadWorkflowPage()
    },
  },
  methods: {
    handleCtrlS(e) {
      if (!(e.ctrlKey && e.key === 's')) return
      e.preventDefault()
      const nodeToPrompt = { requirement: 'requirement', design: 'design', 'api-dev': 'api_dev', 'api-test-fix': 'api_test' }
      let promptType = nodeToPrompt[this.activeNode]
      if (this.activeNode === 'requirement-fetch' && this.requirementFetchActiveTab === 'plain-text-prompt') {
        promptType = 'plain_text_requirement'
      }
      if (promptType) {
        this.savePrompts(promptType)
      }
    },
    goBackToTaskList() {
      this.$router.push('/HomeTask')
    },
    goHome() {
      this.$router.push('/Dashboard')
    },
    reloadWorkflowPage() {
      this.requirementFetchAutoTriggered = false
      this.requirementFetchLogs = []
      this.requirementFetchActiveTab = 'tapd-fetch'
      this.loadWorkflowPage()
    },
    loadWorkflowPage() {
      if (this.taskId <= 0) {
        this.errorMessage = '任务 id 不合法'
        return
      }
      this.loading = true
      this.errorMessage = ''
      taskWorkflowApi.TaskWorkflowCreateOrGet(this.taskId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.loading = false
          this.errorMessage = response?.ErrMsg || '工作流加载失败'
          return
        }
        this.applyWorkflowPayload(response.Data)
        this.loadRequirementFragment(() => {
          this.loading = false
          this.ensureWorkflowSse()
          this.maybeAutoFetchRequirement()
        })
      })
    },
    applyWorkflowPayload(data) {
      this.workflow = data.workflow || {}
      this.homeTask = data.home_task || this.homeTask || {}
      this.workflowId = Number(this.workflow.id || 0)
      this.requirementFetchConfig = data.requirement_fetch_config || this.requirementFetchConfig || {}
    },
    loadRequirementFragment(done) {
      const fragmentId = this.requirementFragmentId
      if (!fragmentId) {
        this.requirementFragment = {}
        this.requirementShareUrl = ''
        if (typeof done === 'function') done()
        return
      }
      MemoryFragmentApi.MemoryFragmentInfo(fragmentId, (response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.requirementFragment = response.Data || {}
          this.refreshRequirementShareUrl()
        } else {
          this.errorMessage = response?.ErrMsg || '需求文档加载失败'
        }
        if (typeof done === 'function') done()
      })
    },
    refreshRequirementShareUrl() {
      const fragmentId = this.requirementFragmentId
      if (!fragmentId) {
        this.requirementShareUrl = ''
        return
      }
      MemoryFragmentApi.MemoryFragmentShareCreate(fragmentId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          return
        }
        const token = String(response.Data.token || '').trim()
        if (!token) {
          this.requirementShareUrl = ''
          return
        }
        const apiHost = String(baseUtils.GetApiHost() || window.location.origin).trim()
        this.requirementShareUrl = new URL(`/share/${encodeURIComponent(token)}`, apiHost).toString()
        this.replaceRequirementShareUrlPlaceholder()
      })
    },
    replaceRequirementShareUrlPlaceholder() {
      if (!this.requirementShareUrl) {
        return
      }
      const placeholder = '{需求文档地址}'
      if (this.workflow.prompt_requirement && this.workflow.prompt_requirement.includes(placeholder)) {
        this.workflow.prompt_requirement = this.workflow.prompt_requirement.replaceAll(placeholder, this.requirementShareUrl)
      }
      if (this.workflow.prompt_plain_text_requirement && this.workflow.prompt_plain_text_requirement.includes(placeholder)) {
        this.workflow.prompt_plain_text_requirement = this.workflow.prompt_plain_text_requirement.replaceAll(placeholder, this.requirementShareUrl)
      }
    },
    ensureWorkflowSse() {
      if (this.workflowId <= 0) {
        return
      }
      const nextDistributeId = `task_workflow_${this.workflowId}`
      if (this.workflowSseDistributeId === nextDistributeId) {
        return
      }
      this.unregisterWorkflowSse()
      sseDistribute.InitFromLoginStatus().then((created) => {
        if (!created && !sseDistribute.GetSseClientId()) {
          return
        }
        this.workflowSseDistributeId = nextDistributeId
        sseDistribute.RegisterReceive(nextDistributeId, this.handleWorkflowSseMessage)
      })
    },
    unregisterWorkflowSse() {
      if (!this.workflowSseDistributeId) {
        return
      }
      sseDistribute.UnRegisterReceive(this.workflowSseDistributeId)
      this.workflowSseDistributeId = ''
    },
    handleWorkflowSseMessage(data) {
      if (!data || Number(data.workflow_id || 0) !== this.workflowId) {
        return
      }
      this.requirementFetchLogs.push({
        workflow_id: Number(data.workflow_id || 0),
        step: String(data.step || '').trim(),
        status: String(data.status || '').trim(),
        message: String(data.message || '').trim(),
        time: Number(data.time || 0),
      })
    },
    maybeAutoFetchRequirement() {
      if (this.requirementFetchAutoTriggered) {
        return
      }
      if (!String(this.homeTask.tapd_url || '').trim()) {
        return
      }
      if (this.requirementFetchStatus === 'success') {
        return
      }
      if (this.requirementFetchStatus === 'running') {
        this.requirementFetchRunning = true
        return
      }
      this.requirementFetchAutoTriggered = true
      this.triggerRequirementFetch(true)
    },
    triggerRequirementFetch(isAuto) {
      if (this.workflowId <= 0 || this.requirementFetchRunning) {
        return
      }
      if (!String(this.homeTask.tapd_url || '').trim()) {
        this.$helperNotify.error('当前任务未配置 TAPD 地址')
        return
      }
      if (!isAuto) {
        this.requirementFetchAutoTriggered = true
      }
      this.requirementFetchRunning = true
      this.errorMessage = ''
      if (!isAuto) {
        this.requirementFetchLogs.push({
          step: 'manual',
          status: 'running',
          message: '手动触发重新抓取',
          time: Math.floor(Date.now() / 1000),
        })
      }
      taskWorkflowApi.TaskWorkflowRequirementFetch(this.workflowId, (response) => {
        this.requirementFetchRunning = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.errorMessage = response?.ErrMsg || '抓取 TAPD 需求失败'
          this.$helperNotify.error(this.errorMessage)
          this.loadWorkflowPage()
          return
        }
        this.$helperNotify.success('TAPD 需求已抓取并写入知识片段')
        this.applyWorkflowPayload({
          workflow: response.Data.workflow || {},
          home_task: this.homeTask,
          requirement_fetch_config: response.Data.requirement_fetch_config || {},
        })
        this.loadRequirementFragment(() => {})
      })
    },
    openRequirementFragment() {
      if (!this.requirementFragmentId) {
        this.$helperNotify.error('当前工作流未绑定需求知识片段')
        return
      }
      const routeInfo = this.$router.resolve({
        path: '/MemoryFragment',
        query: {
          fragment_id: this.requirementFragmentId,
          hide_menu: '1',
        },
      })
      window.open(routeInfo.href, '_blank')
    },
    openPlainTextReqFragment() {
      if (!this.plainTextReqFragmentId) {
        this.$helperNotify.error('当前工作流未绑定纯文本需求知识片段')
        return
      }
      const routeInfo = this.$router.resolve({
        path: '/MemoryFragment',
        query: {
          fragment_id: this.plainTextReqFragmentId,
          hide_menu: '1',
        },
      })
      window.open(routeInfo.href, '_blank')
    },
    savePrompts(promptType) {
      if (this.promptSaving || this.workflowId <= 0) {
        return
      }
      this.promptSaving = promptType
      taskWorkflowApi.TaskWorkflowPromptsSave({
        workflow_id: this.workflowId,
        prompt_requirement: this.workflow.prompt_requirement || '',
        prompt_api_dev: this.workflow.prompt_api_dev || '',
        prompt_api_test: this.workflow.prompt_api_test || '',
        prompt_design: this.workflow.prompt_design || '',
        prompt_plain_text_requirement: this.workflow.prompt_plain_text_requirement || '',
      }, (response) => {
        this.promptSaving = ''
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '提示词保存失败')
          return
        }
        this.$helperNotify.success('提示词已保存')
        if (response.Data?.workflow) {
          this.workflow = { ...this.workflow, ...response.Data.workflow }
        }
      })
    },
    restorePrompts(promptType) {
      if (this.promptRestoring || this.workflowId <= 0) {
        return
      }
      this.$confirm('确定要还原为默认提示词吗？当前内容将被覆盖。', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        this.doRestorePrompts(promptType)
      }).catch(() => {})
    },
    doRestorePrompts(promptType) {
      this.promptRestoring = promptType
      taskWorkflowApi.TaskWorkflowPromptsRestore(this.workflowId, (response) => {
        this.promptRestoring = ''
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '还原提示词失败')
          return
        }
        this.$helperNotify.success('提示词已还原为默认值')
        if (response.Data?.workflow) {
          this.workflow = response.Data.workflow
          this.$nextTick(() => {
            this.replaceRequirementShareUrlPlaceholder()
          })
        }
      })
    },
    copyText(text, successMessage) {
      const value = String(text || '').trim()
      if (!value) {
        this.$helperNotify.error('没有可复制的内容')
        return
      }
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(value).then(() => {
          this.$helperNotify.success(successMessage)
        }).catch(() => {
          this.fallbackCopyText(value, successMessage)
        })
        return
      }
      this.fallbackCopyText(value, successMessage)
    },
    fallbackCopyText(text, successMessage) {
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
        this.$helperNotify.success(successMessage)
      } catch (error) {
        this.$helperNotify.error('复制失败')
      }
      document.body.removeChild(textArea)
    },
    formatUnixTime(unixTime) {
      const value = Number(unixTime || 0)
      if (value <= 0) {
        return '-'
      }
      return new Date(value * 1000).toLocaleString()
    },
    loadTaskConfigLookupData() {
      gitApi.GitConfigList({}, (response) => {
        if (response && response.ErrCode === 0) {
          this.taskConfigGitRepoList = Array.isArray(response.Data?.git_list) ? response.Data.git_list : []
        }
      })
      mysqlSetApi.MysqlList((response) => {
        if (response && response.ErrCode === 0) {
          this.taskConfigMysqlList = Array.isArray(response.Data) ? response.Data : []
        }
      })
      apiManagement.CollectionListBasic({}, (response) => {
        if (response && response.ErrCode === 0) {
          this.taskConfigCollectionList = Array.isArray(response.Data?.list) ? response.Data.list : []
        }
      })
      dockerApi.DockerComposeList({}, (response) => {
        if (response && response.ErrCode === 0) {
          this.taskConfigDockerList = Array.isArray(response.Data?.list) ? response.Data.list : []
        }
      })
      smartLinkSetApi.SmartLinkList((response) => {
        if (response && response.ErrCode === 0) {
          this.taskConfigSmartLinkList = Array.isArray(response.Data?.smart_link_list) ? response.Data.smart_link_list : []
        }
      })
    },
    loadTaskConfigApiFoldersForCollection(collectionId) {
      if (!collectionId) return
      if (this.taskConfigApiFolderMap[collectionId]) return
      apiManagement.CollectionFoldersBasic({ collection_id: collectionId }, (response) => {
        if (response && response.ErrCode === 0) {
          const list = Array.isArray(response.Data?.list) ? response.Data.list : []
          this.taskConfigApiFolderMap = { ...this.taskConfigApiFolderMap, [collectionId]: list }
        }
      })
    },
    getTaskConfigName(type, id) {
      const numId = Number(id || 0)
      if (numId <= 0) return '-'
      if (type === 'git') {
        const item = this.taskConfigGitRepoList.find(r => Number(r.id) === numId)
        return item ? item.name : String(id)
      }
      if (type === 'docker') {
        const item = this.taskConfigDockerList.find(d => Number(d.id) === numId)
        return item ? item.name : String(id)
      }
      if (type === 'mysql') {
        const item = this.taskConfigMysqlList.find(m => Number(m.id) === numId)
        return item ? item.name : String(id)
      }
      if (type === 'smart_link') {
        const item = this.taskConfigSmartLinkList.find(s => Number(s.id) === numId)
        return item ? item.name : String(id)
      }
      return String(id)
    },
    getTaskConfigApiLabel(cfg) {
      const colId = Number(cfg.collection_id || 0)
      if (colId <= 0) return '-'
      const col = this.taskConfigCollectionList.find(c => Number(c.id) === colId)
      if (!col) return String(cfg.collection_id)
      let label = col.name
      const dirId = Number(cfg.dir_id || 0)
      if (dirId > 0) {
        const folders = this.taskConfigApiFolderMap[colId] || []
        const dir = folders.find(d => Number(d.id) === dirId)
        if (dir) {
          label += '/' + dir.name
        }
      }
      return label
    },
    handleTaskStatusChange(newStatus) {
      if (this.statusUpdating || this.taskId <= 0) return
      if (!newStatus || this.homeTask.task_status === newStatus) return
      this.statusUpdating = true
      homeTaskApi.HomeTaskStatusQuickUpdate(this.taskId, newStatus, (response) => {
        this.statusUpdating = false
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '状态切换失败')
          return
        }
        this.$helperNotify.success('状态已切换')
        this.homeTask = { ...this.homeTask, task_status: newStatus }
      })
    },
    getTaskStatusTagType(taskStatus) {
      if (taskStatus === TASK_STATUS_DEVELOPING) {
        return 'success'
      }
      if (taskStatus === TASK_STATUS_SELF_TESTING || taskStatus === TASK_STATUS_TESTING) {
        return 'warning'
      }
      if (taskStatus === TASK_STATUS_TODO) {
        return 'info'
      }
      if (taskStatus === TASK_STATUS_ONLINE) {
        return 'info'
      }
      return ''
    },
  },
}
</script>

<style scoped>
.task-workflow-page {
  height: 100vh;
  background: linear-gradient(180deg, #fdfdfb 0%, #f8faf5 100%);
  padding: 16px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.task-workflow-shell {
  width: 100%;
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  gap: 12px;
}

.task-workflow-header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  padding: 20px 24px;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border: 1px solid #e8e8e0;
  flex-shrink: 0;
}

.task-workflow-header__eyebrow {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.task-workflow-header__title {
  margin: 0;
  font-size: 22px;
  line-height: 1.3;
  color: #303133;
}

.task-workflow-header__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 6px;
  color: #909399;
  font-size: 13px;
}

.task-workflow-header__link {
  color: #3a7a3a;
  text-decoration: none;
}

.task-workflow-header__link:hover {
  text-decoration: underline;
}

.task-workflow-header__actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
}

.task-workflow-home-btn {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 1px solid #e0e0d8;
  background: #fff;
  color: #909399;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: border-color 0.2s, color 0.2s;
}

.task-workflow-home-btn:hover {
  border-color: #3a7a3a;
  color: #3a7a3a;
}

.task-workflow-alert {
  margin-bottom: 0;
}

.task-workflow-nodes {
  display: grid;
  grid-template-columns: repeat(6, minmax(0, 1fr));
  gap: 10px;
  flex-shrink: 0;
}

.task-workflow-node {
  border: 1px solid #e8e8e0;
  border-radius: 8px;
  background: #fff;
  min-height: 50px;
  padding: 14px 16px;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.task-workflow-node:hover {
  border-color: #b7c9a8;
  transform: translateY(-1px);
}

.task-workflow-node--active {
  border-color: #3a7a3a;
  background: #f3f8ef;
  box-shadow: 0 6px 18px rgba(58, 122, 58, 0.14);
}

.task-workflow-node--success {
  background: #f3f8ef;
}

.task-workflow-node--failed {
  background: #fff5f4;
}

.task-workflow-node--running {
  background: #fff9ec;
}

.task-workflow-node__index {
  font-size: 12px;
  color: #909399;
}

.task-workflow-node__label {
  font-size: 15px;
  line-height: 1.4;
  color: #303133;
  font-weight: 600;
}

.task-workflow-node__desc {
  font-size: 12px;
  line-height: 1.5;
  color: #909399;
  font-weight: 400;
}

.task-workflow-content {
  background: #fff;
  border-radius: 12px;
  padding: 16px 20px 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border: 1px solid #e8e8e0;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

.task-workflow-tab {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1;
  height: 100%;
  min-height: 0;
}

.task-workflow-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.task-workflow-card {
  border-radius: 12px;
  padding: 16px;
  background: #fafaf7;
  border: 1px solid #e8e8e0;
  flex: 1;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.task-workflow-prompt-editor {
  flex: 1;
  height: 100%;
  min-height: 0;
}

.task-workflow-card :deep(.md-editor) {
  flex: 1;
  height: 100%;
  min-height: 0;
}

.task-workflow-card :deep(.md-editor-content) {
  min-height: 0;
}

.task-workflow-card :deep(.md-editor-input-wrapper),
.task-workflow-card :deep(.md-editor-preview-wrapper) {
  overflow: auto;
}

/* MdEditor 滚动条绿色 */
.task-workflow-card :deep(.md-editor) {
  --md-scrollbar-bg-color: #edf3e8;
  --md-scrollbar-thumb-color: #9fb39a;
  --md-scrollbar-thumb-hover-color: #869c82;
  --md-scrollbar-thumb-active-color: #7a8f76;
}

.task-workflow-card :deep(.md-editor .md-editor-preview ::-webkit-scrollbar) {
  width: 10px !important;
  height: 10px !important;
}

.task-workflow-card :deep(.md-editor .md-editor-preview ::-webkit-scrollbar-track) {
  background: #edf3e8 !important;
  border-radius: 999px !important;
}

.task-workflow-card :deep(.md-editor .md-editor-preview ::-webkit-scrollbar-thumb) {
  background: #9fb39a !important;
  border: 2px solid #edf3e8 !important;
  border-radius: 999px !important;
}

.task-workflow-card :deep(.md-editor .md-editor-preview ::-webkit-scrollbar-thumb:hover) {
  background: #869c82 !important;
}

.task-workflow-card :deep(.md-editor .md-editor-preview ::-webkit-scrollbar-corner) {
  background: #edf3e8 !important;
}

/* fragment-view 原生滚动条绿色 */
.task-workflow-fragment-view {
  scrollbar-width: thin;
  scrollbar-color: #9fb39a #edf3e8;
}

.task-workflow-fragment-view::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

.task-workflow-fragment-view::-webkit-scrollbar-track {
  background: #edf3e8;
  border-radius: 999px;
}

.task-workflow-fragment-view::-webkit-scrollbar-thumb {
  background: #9fb39a;
  border: 2px solid #edf3e8;
  border-radius: 999px;
}

.task-workflow-fragment-view::-webkit-scrollbar-thumb:hover {
  background: #869c82;
}

.task-workflow-card__header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  margin-bottom: 12px;
}

.task-workflow-card__title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
}

.task-workflow-card__hint {
  margin-bottom: 10px;
  font-size: 13px;
  color: #909399;
  word-break: break-all;
}

.task-workflow-card__hint--error {
  color: #c45656;
}

.task-workflow-card__switch {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.task-workflow-summary-list {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.task-workflow-summary-list--compact {
  margin-bottom: 12px;
}

.task-workflow-summary-item {
  min-width: 140px;
  padding: 12px 14px;
  border-radius: 8px;
  background: #fff;
  border: 1px solid #e8e8e0;
  color: #909399;
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.task-workflow-summary-item strong {
  color: #303133;
  max-width: 360px;
  text-align: right;
  word-break: break-all;
}

.task-workflow-fragment-view {
  border-radius: 10px;
  border: 1px solid #e8e8e0;
  background: #fff;
  overflow: auto;
  min-height: 0;
  flex: 1;
}

.task-workflow-inner-tabs {
  display: flex;
  gap: 4px;
}

.task-workflow-inner-tab {
  padding: 4px 12px;
  font-size: 13px;
  border: 1px solid #e8e8e0;
  border-radius: 6px;
  background: #fff;
  color: #606266;
  cursor: pointer;
  transition: all 0.2s ease;
}

.task-workflow-inner-tab:hover {
  border-color: #b7c9a8;
  color: #3a7a3a;
}

.task-workflow-inner-tab--active {
  background: #3a7a3a;
  color: #fff;
  border-color: #3a7a3a;
}

.task-workflow-tapd-fetch-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
}

.task-workflow-prompt-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
}

.task-workflow-fragment-view__iframe {
  width: 100%;
  height: 100%;
  min-height: 520px;
  border: 0;
  display: block;
}

.task-workflow-fragment-view__empty {
  min-height: 520px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
  font-size: 13px;
}

.task-workflow-config-card {
  overflow: auto;
}

.task-workflow-config-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.task-workflow-config-section__title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}

.task-workflow-config-dev {
  margin-bottom: 12px;
}

.task-workflow-config-dev__index {
  font-size: 13px;
  font-weight: 600;
  color: #3a7a3a;
  margin-bottom: 6px;
}

.task-workflow-config-link {
  color: #3a7a3a;
  text-decoration: none;
  word-break: break-all;
}

.task-workflow-config-link:hover {
  text-decoration: underline;
}

@media (max-width: 1100px) {
  .task-workflow-nodes {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .task-workflow-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

}

@media (max-width: 900px) {
  .task-workflow-page {
    padding: 12px;
  }

  .task-workflow-header {
    flex-direction: column;
    padding: 16px;
  }

  .task-workflow-card__header {
    flex-direction: column;
    align-items: flex-start;
  }
}
</style>
