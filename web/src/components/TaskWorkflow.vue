<template>
  <div class="task-workflow-page" v-loading="loading">
    <div class="task-workflow-shell">
      <header class="task-workflow-header">
        <div class="task-workflow-header__main">
          <div class="task-workflow-header__eyebrow">任务工作流程</div>
          <h1 class="task-workflow-header__title">{{ homeTask.name || `任务 #${taskId}` }}</h1>
          <div class="task-workflow-header__meta">
            <span>状态：{{ workflow.status || '-' }}</span>
            <span>阶段：{{ workflow.current_stage || '-' }}</span>
            <a
              v-if="homeTask.tapd_url"
              :href="homeTask.tapd_url"
              target="_blank"
              class="task-workflow-header__link"
            >
              打开 TAPD
            </a>
          </div>
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
          <GitActionButton compact :loading="loading" @click="loadWorkflowPage">
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

      <el-tabs v-model="activeTab" class="task-workflow-tabs">
        <el-tab-pane label="需求文档 MD" name="requirement">
          <div class="task-workflow-tab">
            <div class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">AI 提示词</div>
                <div class="task-workflow-card__switch">
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
              <MdEditor
                v-model="workflow.prompt_requirement"
                preview-theme="github"
                :preview="true"
                :toolbars="promptEditorToolbars"
                height="100%"
              />
            </div>

          </div>
        </el-tab-pane>

        <el-tab-pane label="开发设计" name="design">
          <div class="task-workflow-tab">
            <div class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">AI 提示词</div>
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
                preview-theme="github"
                :preview="true"
                :toolbars="promptEditorToolbars"
                height="100%"
              />
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="接口开发生成" name="api-dev">
          <div class="task-workflow-tab">
            <div class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">AI 提示词</div>
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
                preview-theme="github"
                :preview="true"
                :toolbars="promptEditorToolbars"
                height="100%"
              />
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="接口自动化测试修复" name="api-test-fix">
          <div class="task-workflow-tab">
            <div class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">AI 提示词</div>
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
                preview-theme="github"
                :preview="true"
                :toolbars="promptEditorToolbars"
                height="100%"
              />
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script>
import { HomeFilled } from '@element-plus/icons-vue'
import GitActionButton from '@/components/base/GitActionButton.vue'
import MarkdownRenderer from '@/components/base/markdown.vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'
import taskWorkflowApi from '@/utils/base/task_workflow'
import baseUtils from '@/utils/base'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

const PROMPT_EDITOR_TOOLBARS = [
  'bold', 'italic', 'strikeThrough', 'title', 'quote',
  'unorderedList', 'orderedList', 'task', 'link', 'code',
  'codeRow', 'table', 'preview', 'fullscreen',
]

export default {
  name: 'TaskWorkflow',
  components: {
    HomeFilled,
    GitActionButton,
    MarkdownRenderer,
    MdEditor,
  },
  data() {
    return {
      activeTab: 'requirement',
      loading: false,
      requirementShareLoading: false,
      errorMessage: '',
      workflowId: 0,
      workflow: {},
      homeTask: {},
      requirementFragment: {},
      requirementShareUrl: '',
      requirementViewMode: 'preview',
      promptSaving: '',
      promptRestoring: '',
      promptEditorToolbars: PROMPT_EDITOR_TOOLBARS,
    }
  },
  computed: {
    taskId() {
      return Number(this.$route.params.taskId || 0)
    },
  },
  mounted() {
    this.loadWorkflowPage()
    window.addEventListener('keydown', this.handleCtrlS)
  },
  beforeUnmount() {
    window.removeEventListener('keydown', this.handleCtrlS)
  },
  watch: {
    '$route.params.taskId'() {
      this.loadWorkflowPage()
    },
  },
  methods: {
    handleCtrlS(e) {
      if (!(e.ctrlKey && e.key === 's')) return
      e.preventDefault()
      const TAB_TO_PROMPT = { requirement: 'requirement', design: 'design', 'api-dev': 'api_dev', 'api-test-fix': 'api_test' }
      const promptType = TAB_TO_PROMPT[this.activeTab]
      if (promptType) this.savePrompts(promptType)
    },
    goBackToTaskList() {
      this.$router.push('/HomeTask')
    },
    goHome() {
      this.$router.push('/Dashboard')
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
        this.loadRequirementFragment()
      })
    },
    applyWorkflowPayload(data) {
      this.workflow = data.workflow || {}
      this.homeTask = data.home_task || {}
      this.workflowId = Number(this.workflow.id || 0)
    },
    loadRequirementFragment() {
      const fragmentId = String(this.workflow.requirement_fragment_id || '').trim()
      if (!fragmentId) {
        this.requirementFragment = {}
        this.requirementShareUrl = ''
        this.loading = false
        return
      }
      MemoryFragmentApi.MemoryFragmentInfo(fragmentId, (response) => {
        this.loading = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.errorMessage = response?.ErrMsg || '需求文档加载失败'
          return
        }
        this.requirementFragment = response.Data || {}
        this.refreshRequirementShareUrl()
      })
    },
    refreshRequirementShareUrl() {
      const fragmentId = String(this.workflow.requirement_fragment_id || '').trim()
      if (!fragmentId) {
        this.requirementShareUrl = ''
        return
      }
      this.requirementShareLoading = true
      MemoryFragmentApi.MemoryFragmentShareCreate(fragmentId, (response) => {
        this.requirementShareLoading = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.$helperNotify.error(response?.ErrMsg || '分享链接生成失败')
          return
        }
        const token = String(response.Data.token || '').trim()
        if (!token) {
          this.requirementShareUrl = ''
          return
        }
        const apiHost = String(baseUtils.GetApiHost() || window.location.origin).trim()
        this.requirementShareUrl = new URL(`/share/${encodeURIComponent(token)}`, apiHost).toString()
        // 获取到分享链接后自动替换需求提示词中的占位符。
        this.replaceRequirementShareUrlPlaceholder()
      })
    },
    replaceRequirementShareUrlPlaceholder() {
      if (!this.requirementShareUrl || !this.workflow.prompt_requirement) {
        return
      }
      const placeholder = '{需求文档地址}'
      if (this.workflow.prompt_requirement.includes(placeholder)) {
        this.workflow.prompt_requirement = this.workflow.prompt_requirement.replaceAll(placeholder, this.requirementShareUrl)
      }
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
          // 还原后如果有分享链接，自动替换需求文档地址占位符。
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
  },
}
</script>

<style scoped>
.task-workflow-page {
  height: 100vh;
  background: #fafaf7;
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

.task-workflow-tabs {
  background: #fff;
  border-radius: 12px;
  padding: 16px 20px 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border: 1px solid #e8e8e0;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.task-workflow-tabs :deep(.el-tabs__header) {
  order: -1;
  flex-shrink: 0;
}

.task-workflow-tabs :deep(.el-tabs__content) {
  flex: 1;
  min-height: 0;
  overflow: auto;
}

.task-workflow-tabs :deep(.el-tab-pane) {
  height: 100%;
}

.task-workflow-tab {
  display: flex;
  flex-direction: column;
  gap: 12px;
  height: 100%;
}

.task-workflow-toolbar {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.task-workflow-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.task-workflow-stat-card {
  border-radius: 12px;
  padding: 16px;
  background: #fafaf7;
  border: 1px solid #e8e8e0;
}

.task-workflow-stat-card__label {
  font-size: 13px;
  color: #909399;
  margin-bottom: 6px;
}

.task-workflow-stat-card__value {
  font-size: 28px;
  font-weight: 700;
  color: #303133;
}

.task-workflow-card {
  border-radius: 12px;
  padding: 16px;
  background: #fafaf7;
  border: 1px solid #e8e8e0;
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.task-workflow-card :deep(.md-editor) {
  flex: 1;
  min-height: 0;
}

.task-workflow-card__label {
  font-size: 13px;
  color: #909399;
  margin-bottom: 8px;
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
}

.task-workflow-card__switch {
  display: flex;
  gap: 8px;
}

.task-workflow-mode-button--active {
  box-shadow: inset 0 0 0 1px rgba(58, 122, 58, 0.24);
}

.task-workflow-inline {
  display: flex;
  gap: 8px;
}

.task-workflow-markdown {
  max-height: 720px;
  overflow: auto;
  background: #fff;
  border-radius: 8px;
  padding: 12px;
  border: 1px solid #e8e8e0;
}

.task-workflow-markdown--compact {
  max-height: 200px;
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
  background: #fafaf7;
  border: 1px solid #e8e8e0;
  color: #909399;
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.task-workflow-summary-item strong {
  color: #303133;
}

.task-workflow-history-list {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.task-workflow-history-item {
  display: grid;
  grid-template-columns: 1.2fr 1fr 1fr;
  gap: 12px;
  padding: 12px 14px;
  border-radius: 8px;
  background: #fafaf7;
  border: 1px solid #e8e8e0;
  color: #606266;
  cursor: pointer;
  transition: background-color 0.2s ease;
}

.task-workflow-history-item:hover {
  background: #f0f0e8;
}

.task-workflow-history-item--active {
  background: #e8f0e3;
  border-color: #3a7a3a;
}

.task-workflow-detail-tabs {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 12px;
}

.task-workflow-history-report {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.task-workflow-ui-assist {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.task-workflow-ui-assist__item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: #606266;
  font-size: 13px;
}

.task-workflow-ui-assist__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
}

.task-workflow-ui-assist__list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.task-workflow-ui-assist__list-item {
  padding: 6px 10px;
  border-radius: 6px;
  background: #fafaf7;
  border: 1px solid #e8e8e0;
  color: #606266;
  font-size: 12px;
}

.task-workflow-ui-assist__tag {
  display: inline-flex;
  align-items: center;
  padding: 4px 10px;
  border-radius: 999px;
  background: #eef3e7;
  color: #3a7a3a;
  font-size: 12px;
}

.task-workflow-dialog-form :deep(.el-form-item) {
  margin-bottom: 16px;
}

.task-workflow-report-table {
  border: 1px solid #e8e8e0;
  border-radius: 8px;
  overflow: hidden;
  background: #fff;
}

.task-workflow-report-table__head,
.task-workflow-report-table__row {
  display: grid;
  grid-template-columns: 1.6fr 1.4fr 0.8fr 0.8fr 0.8fr;
  gap: 12px;
  padding: 10px 14px;
  align-items: center;
}

.task-workflow-report-table__head {
  background: #fafaf7;
  color: #909399;
  font-size: 13px;
  font-weight: 600;
}

.task-workflow-report-table__row {
  border-top: 1px solid #e8e8e0;
  color: #606266;
  font-size: 13px;
}

.task-workflow-report-table__cell-main {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.task-workflow-report-table__sub {
  color: #909399;
  font-size: 12px;
  word-break: break-all;
}

.task-workflow-result--pass {
  color: #2d7a48;
  font-weight: 600;
}

.task-workflow-result--fail {
  color: #b14f46;
  font-weight: 600;
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

  .task-workflow-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

  .task-workflow-report-table__head,
  .task-workflow-report-table__row {
    grid-template-columns: 1.2fr 1fr 0.8fr 0.8fr 0.8fr;
    font-size: 12px;
  }
}
</style>
