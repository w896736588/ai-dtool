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
                style="height: 560px;"
              />
            </div>

            <div class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">需求文档内容</div>
                <div class="task-workflow-card__switch">
                  <GitActionButton
                    compact
                    :class="{ 'task-workflow-mode-button--active': requirementViewMode === 'preview' }"
                    @click="requirementViewMode = 'preview'"
                  >
                    预览
                  </GitActionButton>
                  <GitActionButton
                    compact
                    variant="info"
                    :class="{ 'task-workflow-mode-button--active': requirementViewMode === 'source' }"
                    @click="requirementViewMode = 'source'"
                  >
                    源码
                  </GitActionButton>
                </div>
              </div>
              <MarkdownRenderer
                v-if="requirementViewMode === 'preview'"
                :source="requirementFragment.content || ''"
                class="task-workflow-markdown"
              />
              <el-input
                v-else
                :model-value="requirementFragment.content || ''"
                type="textarea"
                :rows="20"
                readonly
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
                style="height: 560px;"
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
                style="height: 560px;"
              />
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
  </div>
</template>

<script>
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
  },
  watch: {
    '$route.params.taskId'() {
      this.loadWorkflowPage()
    },
  },
  methods: {
    goBackToTaskList() {
      this.$router.push('/HomeTask')
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
  min-height: 100vh;
  background:
    radial-gradient(circle at top left, rgba(150, 190, 160, 0.18), transparent 32%),
    linear-gradient(180deg, #f4f0e8 0%, #f7f5ef 48%, #eef4ec 100%);
  padding: 28px;
  box-sizing: border-box;
}

.task-workflow-shell {
  max-width: 1240px;
  margin: 0 auto;
}

.task-workflow-header {
  display: flex;
  justify-content: space-between;
  gap: 20px;
  align-items: flex-start;
  padding: 28px;
  border-radius: 24px;
  background: rgba(255, 252, 246, 0.92);
  box-shadow: 0 18px 50px rgba(88, 94, 72, 0.08);
  border: 1px solid rgba(114, 129, 101, 0.12);
  margin-bottom: 20px;
}

.task-workflow-header__eyebrow {
  font-size: 12px;
  letter-spacing: 0.18em;
  text-transform: uppercase;
  color: #7b8167;
  margin-bottom: 8px;
}

.task-workflow-header__title {
  margin: 0;
  font-size: 30px;
  line-height: 1.2;
  color: #2f3a2e;
}

.task-workflow-header__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 14px;
  margin-top: 10px;
  color: #5e6553;
  font-size: 14px;
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
  gap: 10px;
  flex-wrap: wrap;
}

.task-workflow-alert {
  margin-bottom: 16px;
}

.task-workflow-tabs {
  background: rgba(255, 255, 255, 0.78);
  border-radius: 24px;
  padding: 18px 20px 24px;
  box-shadow: 0 16px 42px rgba(68, 86, 63, 0.08);
}

.task-workflow-tab {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.task-workflow-toolbar {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.task-workflow-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.task-workflow-stat-card {
  border-radius: 18px;
  padding: 18px;
  background: linear-gradient(180deg, rgba(253, 252, 247, 0.96), rgba(242, 248, 240, 0.96));
  border: 1px solid rgba(120, 136, 108, 0.14);
}

.task-workflow-stat-card__label {
  font-size: 13px;
  color: #70765f;
  margin-bottom: 8px;
}

.task-workflow-stat-card__value {
  font-size: 30px;
  font-weight: 700;
  color: #2f3b2f;
}

.task-workflow-card {
  border-radius: 20px;
  padding: 18px;
  background: rgba(255, 255, 255, 0.86);
  border: 1px solid rgba(122, 136, 114, 0.12);
}

.task-workflow-card__label {
  font-size: 13px;
  color: #6b725d;
  margin-bottom: 10px;
}

.task-workflow-card__header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  margin-bottom: 14px;
}

.task-workflow-card__title {
  font-size: 18px;
  font-weight: 600;
  color: #2d352e;
}

.task-workflow-card__hint {
  margin-bottom: 12px;
  font-size: 13px;
  color: #6d735f;
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
  gap: 10px;
}

.task-workflow-markdown {
  max-height: 720px;
  overflow: auto;
  background: #fcfbf7;
  border-radius: 14px;
  padding: 12px;
}

.task-workflow-markdown--compact {
  max-height: 200px;
}

.task-workflow-summary-list {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
}

.task-workflow-summary-list--compact {
  margin-bottom: 14px;
}

.task-workflow-summary-item {
  min-width: 140px;
  padding: 14px 16px;
  border-radius: 14px;
  background: #f9f7f1;
  color: #626958;
  display: flex;
  justify-content: space-between;
  gap: 14px;
}

.task-workflow-summary-item strong {
  color: #2f3a2e;
}

.task-workflow-history-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.task-workflow-history-item {
  display: grid;
  grid-template-columns: 1.2fr 1fr 1fr;
  gap: 12px;
  padding: 14px 16px;
  border-radius: 14px;
  background: #f8f5ee;
  color: #455241;
  cursor: pointer;
  transition: background-color 0.2s ease, transform 0.2s ease;
}

.task-workflow-history-item:hover {
  background: #f1ecdf;
  transform: translateY(-1px);
}

.task-workflow-history-item--active {
  background: #e8f0e3;
  box-shadow: inset 0 0 0 1px rgba(77, 120, 68, 0.18);
}

.task-workflow-detail-tabs {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-bottom: 14px;
}

.task-workflow-history-report {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-workflow-ui-assist {
  display: flex;
  flex-direction: column;
  gap: 12px;
}

.task-workflow-ui-assist__item {
  display: flex;
  flex-direction: column;
  gap: 6px;
  color: #55614f;
  font-size: 13px;
}

.task-workflow-ui-assist__tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.task-workflow-ui-assist__list {
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.task-workflow-ui-assist__list-item {
  padding: 8px 10px;
  border-radius: 10px;
  background: #f6f3ec;
  color: #445340;
  font-size: 12px;
}

.task-workflow-ui-assist__tag {
  display: inline-flex;
  align-items: center;
  padding: 6px 10px;
  border-radius: 999px;
  background: #eef3e7;
  color: #40533d;
  font-size: 12px;
}

.task-workflow-dialog-form :deep(.el-form-item) {
  margin-bottom: 16px;
}

.task-workflow-report-table {
  border: 1px solid rgba(122, 136, 114, 0.14);
  border-radius: 16px;
  overflow: hidden;
  background: #fcfbf7;
}

.task-workflow-report-table__head,
.task-workflow-report-table__row {
  display: grid;
  grid-template-columns: 1.6fr 1.4fr 0.8fr 0.8fr 0.8fr;
  gap: 12px;
  padding: 12px 14px;
  align-items: center;
}

.task-workflow-report-table__head {
  background: #eef3e7;
  color: #55614f;
  font-size: 13px;
  font-weight: 600;
}

.task-workflow-report-table__row {
  border-top: 1px solid rgba(122, 136, 114, 0.1);
  color: #344034;
  font-size: 13px;
}

.task-workflow-report-table__cell-main {
  display: flex;
  flex-direction: column;
  gap: 4px;
  min-width: 0;
}

.task-workflow-report-table__sub {
  color: #7a725f;
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
    padding: 16px;
  }

  .task-workflow-header {
    flex-direction: column;
    padding: 20px;
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
