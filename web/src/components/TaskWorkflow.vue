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
            <div class="task-workflow-toolbar">
              <GitActionButton compact variant="info" :loading="requirementShareLoading" @click="refreshRequirementShareUrl">
                刷新分享链接
              </GitActionButton>
              <GitActionButton compact @click="copyRequirementPrompt">
                复制 AI 提示词
              </GitActionButton>
            </div>

            <div class="task-workflow-card">
              <div class="task-workflow-card__label">知识片段分享地址</div>
              <div class="task-workflow-inline">
                <el-input :model-value="requirementShareUrl" readonly />
                <GitActionButton compact @click="copyText(requirementShareUrl, '分享地址已复制')">
                  复制
                </GitActionButton>
              </div>
            </div>

            <div class="task-workflow-card">
              <div class="task-workflow-card__label">给 AI 的提示词</div>
              <el-input
                :model-value="requirementPromptText"
                type="textarea"
                :rows="3"
                readonly
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

        <el-tab-pane label="开发执行 MD" name="dev-plan">
          <div class="task-workflow-tab">
            <div class="task-workflow-toolbar">
              <GitActionButton v-if="!devPlanContent" compact variant="info" :loading="devPlanInitializing" @click="initDevPlanIfNeeded">
                初始化开发执行文档
              </GitActionButton>
              <GitActionButton compact :loading="devPlanSaving" @click="saveDevPlan">
                保存
              </GitActionButton>
            </div>

            <div v-if="devPlanContent" class="task-workflow-card">
              <div class="task-workflow-card__label">知识片段分享地址</div>
              <div class="task-workflow-inline">
                <el-input :model-value="devPlanShareUrl" readonly />
                <GitActionButton compact @click="copyText(devPlanShareUrl, '分享地址已复制')">
                  复制
                </GitActionButton>
                <GitActionButton compact variant="info" :loading="devPlanShareLoading" @click="refreshDevPlanShareUrl">
                  刷新
                </GitActionButton>
              </div>
            </div>

            <div class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">给 AI 的提示词</div>
                <GitActionButton compact @click="copyDevPlanPrompt">
                  复制 AI 提示词
                </GitActionButton>
              </div>
              <MarkdownRenderer
                :source="devPlanPromptText"
                class="task-workflow-markdown task-workflow-markdown--compact"
              />
            </div>

            <div class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">开发执行内容</div>
                <div class="task-workflow-card__switch">
                  <GitActionButton
                    compact
                    :class="{ 'task-workflow-mode-button--active': devPlanViewMode === 'edit' }"
                    @click="devPlanViewMode = 'edit'"
                  >
                    编辑
                  </GitActionButton>
                  <GitActionButton
                    compact
                    variant="info"
                    :class="{ 'task-workflow-mode-button--active': devPlanViewMode === 'preview' }"
                    @click="devPlanViewMode = 'preview'"
                  >
                    预览
                  </GitActionButton>
                </div>
              </div>
              <el-input
                v-if="devPlanViewMode === 'edit'"
                v-model="devPlanContent"
                type="textarea"
                :rows="20"
                placeholder="初始化后可在这里编辑开发执行文档"
              />
              <MarkdownRenderer
                v-else
                :source="devPlanContent || ''"
                class="task-workflow-markdown"
              />
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="测试接口计划" name="test-plan">
          <div class="task-workflow-tab">
            <div class="task-workflow-toolbar">
              <GitActionButton compact variant="info" :loading="uiAssistGenerating" :disabled="!canGenerateArtifacts" @click="openUIAssistDialog">
                页面辅助识别
              </GitActionButton>
              <GitActionButton compact :loading="coverageGenerating" :disabled="!canGenerateArtifacts" @click="generateCoverageReport">
                生成覆盖分析
              </GitActionButton>
              <GitActionButton compact :loading="testPlanGenerating" :disabled="!canGenerateArtifacts" @click="generateTestPlan">
                生成测试计划
              </GitActionButton>
            </div>

            <div class="task-workflow-grid">
              <div class="task-workflow-stat-card">
                <div class="task-workflow-stat-card__label">需求点</div>
                <div class="task-workflow-stat-card__value">{{ coverageSummary.requirement_points }}</div>
              </div>
              <div class="task-workflow-stat-card">
                <div class="task-workflow-stat-card__label">覆盖接口</div>
                <div class="task-workflow-stat-card__value">{{ testPlanApiCaseCount }}</div>
              </div>
              <div class="task-workflow-stat-card">
                <div class="task-workflow-stat-card__label">前置条件</div>
                <div class="task-workflow-stat-card__value">{{ testPlanPreconditionCount }}</div>
              </div>
              <div class="task-workflow-stat-card">
                <div class="task-workflow-stat-card__label">阻塞项</div>
                <div class="task-workflow-stat-card__value">{{ testPlanBlockedCount }}</div>
              </div>
            </div>

            <div class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">页面辅助识别</div>
              </div>
              <el-empty
                v-if="!uiAssistInfo.download_url"
                description="还没有页面辅助识别结果"
              />
              <div v-else class="task-workflow-ui-assist">
                <div class="task-workflow-card__hint">
                  最近识别：{{ uiAssistRunInfo.run_no || '-' }} / {{ uiAssistRunInfo.status || '-' }}
                </div>
                <div class="task-workflow-ui-assist__item">
                  <span>抓取结果下载地址</span>
                  <a :href="uiAssistInfo.download_url" target="_blank" class="task-workflow-header__link">{{ uiAssistInfo.download_url }}</a>
                </div>
                <div class="task-workflow-ui-assist__item">
                  <span>推荐提示词</span>
                </div>
                <el-input
                  :model-value="uiAssistInfo.prompt_text || ''"
                  type="textarea"
                  :rows="3"
                  readonly
                />
                <div class="task-workflow-ui-assist__item">
                  <span>候选接口</span>
                  <div v-if="uiAssistApiCandidates.length" class="task-workflow-ui-assist__tags">
                    <span
                      v-for="item in uiAssistApiCandidates"
                      :key="item"
                      class="task-workflow-ui-assist__tag"
                    >
                      {{ item }}
                    </span>
                  </div>
                  <span v-else>暂未从抓取内容中识别到接口路径</span>
                </div>
                <div class="task-workflow-ui-assist__item">
                  <span>步骤标题</span>
                  <div v-if="uiAssistStepTitles.length" class="task-workflow-ui-assist__list">
                    <div
                      v-for="item in uiAssistStepTitles"
                      :key="item"
                      class="task-workflow-ui-assist__list-item"
                    >
                      {{ item }}
                    </div>
                  </div>
                  <span v-else>暂未提取到明确步骤标题</span>
                </div>
                <div class="task-workflow-ui-assist__item">
                  <span>参数线索</span>
                  <div v-if="uiAssistParameterHints.length" class="task-workflow-ui-assist__tags">
                    <span
                      v-for="item in uiAssistParameterHints"
                      :key="item"
                      class="task-workflow-ui-assist__tag"
                    >
                      {{ item }}
                    </span>
                  </div>
                  <span v-else>暂未提取到参数线索</span>
                </div>
                <div class="task-workflow-ui-assist__item">
                  <span>抓取内容 Markdown</span>
                </div>
                <el-input
                  :model-value="uiAssistInfo.markdown || ''"
                  type="textarea"
                  :rows="10"
                  readonly
                />
              </div>
            </div>

            <div class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">覆盖分析摘要</div>
              </div>
              <div v-if="coverageRunInfo.run_no" class="task-workflow-card__hint">
                最近生成：{{ coverageRunInfo.run_no }} / {{ coverageRunInfo.status || '-' }}
              </div>
              <div class="task-workflow-summary-list">
                <div class="task-workflow-summary-item">
                  <span>已覆盖</span>
                  <strong>{{ coverageSummary.covered }}</strong>
                </div>
                <div class="task-workflow-summary-item">
                  <span>部分覆盖</span>
                  <strong>{{ coverageSummary.partial }}</strong>
                </div>
                <div class="task-workflow-summary-item">
                  <span>缺失</span>
                  <strong>{{ coverageSummary.missing }}</strong>
                </div>
                <div class="task-workflow-summary-item">
                  <span>疑问项</span>
                  <strong>{{ coverageSummary.questions }}</strong>
                </div>
              </div>
            </div>

            <div class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">测试计划 JSON</div>
              </div>
              <div v-if="testPlanRunInfo.run_no" class="task-workflow-card__hint">
                最近生成：{{ testPlanRunInfo.run_no }} / {{ testPlanRunInfo.status || '-' }}
              </div>
              <div v-if="testPlanUsesUIAssist" class="task-workflow-card__hint">
                本次计划已合并页面辅助识别结果，共使用 {{ testPlanUIAssistCandidates.length }} 个候选接口
              </div>
              <el-input
                :model-value="formatJson(testPlanInfo)"
                type="textarea"
                :rows="18"
                readonly
              />
            </div>
          </div>
        </el-tab-pane>

        <el-tab-pane label="接口测试与覆盖检查" name="test-run">
          <div class="task-workflow-tab">
            <div class="task-workflow-toolbar">
              <GitActionButton compact variant="info" :loading="coverageGenerating" :disabled="!canGenerateArtifacts" @click="generateCoverageReport">
                执行覆盖检查
              </GitActionButton>
              <GitActionButton compact :loading="testRunExecuting" :disabled="!canGenerateArtifacts" @click="executeTestPlan(false)">
                执行接口测试
              </GitActionButton>
              <GitActionButton compact :loading="testRunExecuting" :disabled="!canGenerateArtifacts" @click="executeTestPlan(true)">
                执行覆盖检查+接口测试
              </GitActionButton>
            </div>

            <div class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">当前状态</div>
              </div>
              <div class="task-workflow-summary-list">
                <div class="task-workflow-summary-item">
                  <span>工作流状态</span>
                  <strong>{{ workflow.status || '-' }}</strong>
                </div>
                <div class="task-workflow-summary-item">
                  <span>当前阶段</span>
                  <strong>{{ workflow.current_stage || '-' }}</strong>
                </div>
                <div class="task-workflow-summary-item">
                  <span>历史记录数</span>
                  <strong>{{ testRunHistory.length }}</strong>
                </div>
                <div class="task-workflow-summary-item">
                  <span>最近通过数</span>
                  <strong>{{ latestTestReportPassed }}</strong>
                </div>
                <div class="task-workflow-summary-item">
                  <span>最近失败数</span>
                  <strong>{{ latestTestReportFailed }}</strong>
                </div>
              </div>
            </div>

            <div class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">最近测试报告</div>
              </div>
              <div v-if="latestTestCaseResults.length" class="task-workflow-report-table">
                <div class="task-workflow-report-table__head">
                  <div>用例</div>
                  <div>接口</div>
                  <div>状态码</div>
                  <div>耗时</div>
                  <div>结果</div>
                </div>
                <div
                  v-for="item in latestTestCaseResults"
                  :key="item.case_id || item.api_id || item.api_uri"
                  class="task-workflow-report-table__row"
                >
                  <div>{{ item.name || item.case_id || '-' }}</div>
                  <div>{{ item.api_uri || '-' }}</div>
                  <div>{{ item.status_code || '-' }}</div>
                  <div>{{ item.response_time_ms || 0 }} ms</div>
                  <div :class="item.passed ? 'task-workflow-result--pass' : 'task-workflow-result--fail'">
                    {{ item.passed ? '通过' : '失败' }}
                  </div>
                </div>
              </div>
              <el-input
                v-else
                :model-value="formatJson(latestTestReport)"
                type="textarea"
                :rows="12"
                readonly
              />
            </div>

            <div class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">历史记录</div>
              </div>
              <el-empty
                v-if="testRunHistory.length === 0"
                description="当前还没有测试执行记录"
              />
              <div v-else class="task-workflow-history-list">
                <div
                  v-for="item in testRunHistory"
                  :key="item.id || item.run_no"
                  class="task-workflow-history-item"
                  :class="{ 'task-workflow-history-item--active': activeHistoryRunId === Number(item.id || 0) }"
                  @click="selectHistoryItem(item)"
                >
                  <div>{{ item.run_no || `#${item.id}` }}</div>
                  <div>{{ item.run_type || '-' }}</div>
                  <div>{{ item.status || '-' }}</div>
                </div>
              </div>
            </div>

            <div v-if="activeHistoryItem.id" class="task-workflow-card">
              <div class="task-workflow-card__header">
                <div class="task-workflow-card__title">历史详情</div>
                <div class="task-workflow-card__hint">
                  {{ activeHistoryItem.run_no || `#${activeHistoryItem.id}` }} / {{ activeHistoryItem.run_type || '-' }}
                </div>
              </div>
              <div class="task-workflow-detail-tabs">
                <GitActionButton
                  compact
                  :class="{ 'task-workflow-mode-button--active': historyDetailTab === 'report' }"
                  @click="historyDetailTab = 'report'"
                >
                  测试报告
                </GitActionButton>
                <GitActionButton
                  compact
                  variant="info"
                  :class="{ 'task-workflow-mode-button--active': historyDetailTab === 'plan' }"
                  @click="historyDetailTab = 'plan'"
                >
                  测试计划
                </GitActionButton>
                <GitActionButton
                  compact
                  variant="info"
                  :class="{ 'task-workflow-mode-button--active': historyDetailTab === 'coverage' }"
                  @click="historyDetailTab = 'coverage'"
                >
                  覆盖分析
                </GitActionButton>
                <GitActionButton
                  compact
                  variant="info"
                  :class="{ 'task-workflow-mode-button--active': historyDetailTab === 'snapshot' }"
                  @click="historyDetailTab = 'snapshot'"
                >
                  文档快照
                </GitActionButton>
              </div>
              <el-input
                v-if="historyDetailTab === 'snapshot'"
                :model-value="historySnapshotText"
                type="textarea"
                :rows="16"
                readonly
              />
              <div v-else-if="historyDetailTab === 'report' && activeHistoryCaseResults.length" class="task-workflow-history-report">
                <div class="task-workflow-summary-list task-workflow-summary-list--compact">
                  <div class="task-workflow-summary-item">
                    <span>总数</span>
                    <strong>{{ activeHistoryReportSummary.total || 0 }}</strong>
                  </div>
                  <div class="task-workflow-summary-item">
                    <span>通过</span>
                    <strong>{{ activeHistoryReportSummary.passed || 0 }}</strong>
                  </div>
                  <div class="task-workflow-summary-item">
                    <span>失败</span>
                    <strong>{{ activeHistoryReportSummary.failed || 0 }}</strong>
                  </div>
                </div>
                <div class="task-workflow-report-table">
                  <div class="task-workflow-report-table__head">
                    <div>用例</div>
                    <div>接口</div>
                    <div>状态码</div>
                    <div>耗时</div>
                    <div>结果</div>
                  </div>
                  <div
                    v-for="item in activeHistoryCaseResults"
                    :key="item.case_id || item.api_id || item.api_uri"
                    class="task-workflow-report-table__row"
                  >
                    <div class="task-workflow-report-table__cell-main">
                      <div>{{ item.name || item.case_id || '-' }}</div>
                      <div v-if="item.error || item.errmsg" class="task-workflow-report-table__sub">
                        {{ item.error || item.errmsg }}
                      </div>
                    </div>
                    <div class="task-workflow-report-table__cell-main">
                      <div>{{ item.api_uri || '-' }}</div>
                      <div v-if="item.method" class="task-workflow-report-table__sub">{{ item.method }}</div>
                    </div>
                    <div>{{ item.status_code || '-' }}</div>
                    <div>{{ item.response_time_ms || 0 }} ms</div>
                    <div :class="item.passed ? 'task-workflow-result--pass' : 'task-workflow-result--fail'">
                      {{ item.passed ? '通过' : '失败' }}
                    </div>
                  </div>
                </div>
              </div>
              <el-input
                v-else
                :model-value="formatJson(activeHistoryPayload)"
                type="textarea"
                :rows="16"
                readonly
              />
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>
    <el-dialog v-model="uiAssistDialogVisible" title="页面辅助识别" width="720px">
      <el-form label-width="120px" class="task-workflow-dialog-form">
        <el-form-item label="SmartLink">
          <el-select v-model="uiAssistForm.smart_link_id" placeholder="选择 SmartLink" style="width: 100%" @change="onUIAssistSmartLinkChange">
            <el-option
              v-for="item in smartLinkOptions"
              :key="item.id"
              :label="item.name || `#${item.id}`"
              :value="Number(item.id)"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="链接标签">
          <el-select v-model="uiAssistForm.label" placeholder="选择 link label" style="width: 100%">
            <el-option
              v-for="item in selectedSmartLinkLabels"
              :key="item.label"
              :label="item.label"
              :value="item.label"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="页面地址">
          <el-input v-model="uiAssistForm.jump_url" placeholder="https://example.com/page" />
        </el-form-item>
        <el-form-item label="CSS Selector">
          <el-input v-model="uiAssistForm.css_selector" placeholder=".page-root 或 [data-testid='content']" />
        </el-form-item>
        <el-form-item label="等待秒数">
          <el-input-number v-model="uiAssistForm.wait_seconds" :min="1" :max="30" />
        </el-form-item>
      </el-form>
      <template #footer>
        <GitActionButton @click="uiAssistDialogVisible = false">取消</GitActionButton>
        <GitActionButton :loading="uiAssistGenerating" @click="submitUIAssist">开始识别</GitActionButton>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import GitActionButton from '@/components/base/GitActionButton.vue'
import MarkdownRenderer from '@/components/base/markdown.vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'
import smartLinkSet from '@/utils/base/smart_link_set'
import taskWorkflowApi from '@/utils/base/task_workflow'
import baseUtils from '@/utils/base'

export default {
  name: 'TaskWorkflow',
  components: {
    GitActionButton,
    MarkdownRenderer,
  },
  data() {
    return {
      activeTab: 'requirement',
      loading: false,
      devPlanInitializing: false,
      devPlanSaving: false,
      requirementShareLoading: false,
      uiAssistGenerating: false,
      coverageGenerating: false,
      testPlanGenerating: false,
      testRunExecuting: false,
      errorMessage: '',
      workflowId: 0,
      workflow: {},
      homeTask: {},
      requirementFragment: {},
      requirementShareUrl: '',
      requirementViewMode: 'preview',
      devPlanViewMode: 'edit',
      devPlanContent: '',
      devPlanFragment: {},
      devPlanShareUrl: '',
      devPlanShareLoading: false,
      coverageInfo: {},
      coverageRunInfo: {},
      uiAssistInfo: {},
      uiAssistRunInfo: {},
      testPlanInfo: {},
      testPlanRunInfo: {},
      testRunHistory: [],
      activeHistoryRunId: 0,
      historyDetailTab: 'report',
      uiAssistDialogVisible: false,
      smartLinkOptions: [],
      uiAssistForm: {
        smart_link_id: 0,
        label: '',
        jump_url: '',
        css_selector: '',
        wait_seconds: 5,
      },
    }
  },
  computed: {
    taskId() {
      return Number(this.$route.params.taskId || 0)
    },
    requirementPromptText() {
      const shareUrl = this.requirementShareUrl || 'xxxxx'
      return `读取 ${shareUrl}（TAPD 抓取后生成的知识片段分享地址），分析并设计方案`
    },
    devPlanPromptText() {
      const workflowId = Number(this.workflow.id || this.workflowId || 0)
      const apiHost = baseUtils.GetApiHost() || window.location.origin
      const token = baseUtils.GetSafeToken()
      return `将开发方案通过以下接口更新，content 为完整 Markdown 内容：

\`\`\`python
import requests
requests.post('${apiHost}/api/task/workflow/dev-plan/save', headers={'Content-Type': 'application/json', 'Token': '${token}'}, json={'workflow_id': ${workflowId}, 'content': '# 开发执行说明\\n...'})
\`\`\``
    },
    coverageSummary() {
      return this.coverageInfo.summary || {
        requirement_points: 0,
        covered: 0,
        partial: 0,
        missing: 0,
        questions: 0,
        blocked: 0,
      }
    },
    testPlanApiCaseCount() {
      return Array.isArray(this.testPlanInfo.api_cases) ? this.testPlanInfo.api_cases.length : 0
    },
    testPlanPreconditionCount() {
      return Array.isArray(this.testPlanInfo.preconditions) ? this.testPlanInfo.preconditions.length : 0
    },
    testPlanBlockedCount() {
      return Array.isArray(this.testPlanInfo.blocked_items) ? this.testPlanInfo.blocked_items.length : 0
    },
    testPlanUsesUIAssist() {
      return !!this.testPlanInfo.ui_assist_used
    },
    testPlanUIAssistCandidates() {
      return Array.isArray(this.testPlanInfo.ui_assist_candidates) ? this.testPlanInfo.ui_assist_candidates : []
    },
    canGenerateArtifacts() {
      return this.workflowId > 0 && String(this.workflow.requirement_fragment_id || '').trim() !== ''
    },
    latestTestRun() {
      return this.testRunHistory.find(item => item.run_type === 'api_test_execute') || {}
    },
    latestTestReport() {
      return this.latestTestRun.test_report || {}
    },
    latestTestCaseResults() {
      return Array.isArray(this.latestTestReport?.case_results) ? this.latestTestReport.case_results : []
    },
    latestTestReportPassed() {
      return Number(this.latestTestReport?.summary?.passed || 0)
    },
    latestTestReportFailed() {
      return Number(this.latestTestReport?.summary?.failed || 0)
    },
    activeHistoryItem() {
      return this.testRunHistory.find(item => Number(item.id || 0) === this.activeHistoryRunId) || {}
    },
    activeHistoryPayload() {
      if (this.historyDetailTab === 'plan') {
        return this.activeHistoryItem.test_plan || {}
      }
      if (this.historyDetailTab === 'coverage') {
        return this.activeHistoryItem.coverage_report || {}
      }
      return this.activeHistoryItem.test_report || {}
    },
    historySnapshotText() {
      if (!this.activeHistoryItem.id) {
        return ''
      }
      return [
        '# 需求文档快照',
        String(this.activeHistoryItem.requirement_snapshot_md || ''),
        '',
        '# 开发执行快照',
        String(this.activeHistoryItem.dev_plan_snapshot_md || ''),
      ].join('\n')
    },
    activeHistoryReportSummary() {
      return this.activeHistoryItem?.test_report?.summary || {}
    },
    activeHistoryCaseResults() {
      return Array.isArray(this.activeHistoryItem?.test_report?.case_results) ? this.activeHistoryItem.test_report.case_results : []
    },
    selectedSmartLink() {
      return this.smartLinkOptions.find(item => Number(item.id || 0) === Number(this.uiAssistForm.smart_link_id || 0)) || {}
    },
    selectedSmartLinkLabels() {
      return Array.isArray(this.selectedSmartLink.linkList) ? this.selectedSmartLink.linkList : []
    },
    uiAssistApiCandidates() {
      return Array.isArray(this.uiAssistInfo.api_candidates) ? this.uiAssistInfo.api_candidates : []
    },
    uiAssistStructuredSummary() {
      return this.uiAssistInfo.structured_summary || {}
    },
    uiAssistStepTitles() {
      return Array.isArray(this.uiAssistStructuredSummary.step_titles) ? this.uiAssistStructuredSummary.step_titles : []
    },
    uiAssistParameterHints() {
      return Array.isArray(this.uiAssistStructuredSummary.parameter_hints) ? this.uiAssistStructuredSummary.parameter_hints : []
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
        this.loadDevPlanFragment()
        this.loadUIAssistInfo()
        this.loadCoverageInfo()
        this.loadTestPlanInfo()
        this.loadTestRunHistory()
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
      })
    },
    loadDevPlanFragment() {
      if (this.workflowId <= 0) {
        return
      }
      taskWorkflowApi.TaskWorkflowDevPlanInfo(this.workflowId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.devPlanContent = ''
          this.devPlanFragment = {}
          return
        }
        this.devPlanFragment = response.Data.fragment || {}
        this.devPlanContent = String(this.devPlanFragment.content || '')
        this.refreshDevPlanShareUrl()
      })
    },
    refreshDevPlanShareUrl() {
      const fragmentId = String(this.devPlanFragment.id || '').trim()
      if (!fragmentId) {
        this.devPlanShareUrl = ''
        return
      }
      this.devPlanShareLoading = true
      MemoryFragmentApi.MemoryFragmentShareCreate(fragmentId, (response) => {
        this.devPlanShareLoading = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.$helperNotify.error(response?.ErrMsg || '分享链接生成失败')
          return
        }
        const token = String(response.Data.token || '').trim()
        if (!token) {
          this.devPlanShareUrl = ''
          return
        }
        const apiHost = String(baseUtils.GetApiHost() || window.location.origin).trim()
        this.devPlanShareUrl = new URL(`/share/${encodeURIComponent(token)}`, apiHost).toString()
      })
    },
    loadCoverageInfo() {
      if (this.workflowId <= 0) {
        this.coverageInfo = {}
        return
      }
      taskWorkflowApi.TaskWorkflowCoverageInfo(this.workflowId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.coverageInfo = {}
          this.coverageRunInfo = {}
          return
        }
        this.coverageInfo = response.Data.coverage_report || {}
        this.coverageRunInfo = response.Data.test_run || {}
      })
    },
    loadUIAssistInfo() {
      if (this.workflowId <= 0) {
        this.uiAssistInfo = {}
        this.uiAssistRunInfo = {}
        return
      }
      taskWorkflowApi.TaskWorkflowUIAssistInfo(this.workflowId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.uiAssistInfo = {}
          this.uiAssistRunInfo = {}
          return
        }
        this.uiAssistInfo = response.Data.ui_assist || {}
        this.uiAssistRunInfo = response.Data.test_run || {}
      })
    },
    loadTestPlanInfo() {
      if (this.workflowId <= 0) {
        this.testPlanInfo = {}
        return
      }
      taskWorkflowApi.TaskWorkflowTestPlanInfo(this.workflowId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.testPlanInfo = {}
          this.testPlanRunInfo = {}
          return
        }
        this.testPlanInfo = response.Data.test_plan || {}
        this.testPlanRunInfo = response.Data.test_run || {}
      })
    },
    loadTestRunHistory() {
      if (this.workflowId <= 0) {
        this.testRunHistory = []
        return
      }
      taskWorkflowApi.TaskWorkflowTestRunList(this.workflowId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.testRunHistory = []
          this.activeHistoryRunId = 0
          return
        }
        this.testRunHistory = Array.isArray(response.Data.list) ? response.Data.list : []
        if (!this.testRunHistory.length) {
          this.activeHistoryRunId = 0
          return
        }
        const activeExists = this.testRunHistory.some(item => Number(item.id || 0) === this.activeHistoryRunId)
        if (!activeExists) {
          this.activeHistoryRunId = Number(this.testRunHistory[0].id || 0)
        }
      })
    },
    initDevPlanIfNeeded() {
      if (this.workflowId <= 0) {
        return
      }
      this.devPlanInitializing = true
      taskWorkflowApi.TaskWorkflowDevPlanInit(this.workflowId, (response) => {
        this.devPlanInitializing = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.$helperNotify.error(response?.ErrMsg || '开发执行文档初始化失败')
          return
        }
        this.workflow = response.Data.workflow || this.workflow
        this.devPlanFragment = response.Data.fragment || {}
        this.devPlanContent = String(this.devPlanFragment.content || '')
        this.$helperNotify.success('开发执行文档已初始化')
      })
    },
    saveDevPlan() {
      if (this.workflowId <= 0) {
        return
      }
      if (!String(this.devPlanContent || '').trim()) {
        this.$helperNotify.error('开发执行内容不能为空')
        return
      }
      this.devPlanSaving = true
      taskWorkflowApi.TaskWorkflowDevPlanSave(this.workflowId, this.devPlanContent, (response) => {
        this.devPlanSaving = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.$helperNotify.error(response?.ErrMsg || '开发执行文档保存失败')
          return
        }
        this.workflow = response.Data.workflow || this.workflow
        this.devPlanFragment = response.Data.fragment || {}
        this.devPlanContent = String(this.devPlanFragment.content || this.devPlanContent)
        this.$helperNotify.success('开发执行文档已保存')
      })
    },
    generateCoverageReport() {
      if (!this.canGenerateArtifacts) {
        return
      }
      this.coverageGenerating = true
      taskWorkflowApi.TaskWorkflowCoverageGenerate(this.workflowId, (response) => {
        this.coverageGenerating = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.$helperNotify.error(response?.ErrMsg || '覆盖分析生成失败')
          return
        }
        this.workflow = response.Data.workflow || this.workflow
        this.coverageInfo = response.Data.coverage_report || {}
        this.coverageRunInfo = response.Data.test_run || {}
        this.loadTestRunHistory()
        this.$helperNotify.success('覆盖分析已生成')
      })
    },
    generateTestPlan() {
      if (!this.canGenerateArtifacts) {
        return
      }
      this.testPlanGenerating = true
      taskWorkflowApi.TaskWorkflowTestPlanGenerate(this.workflowId, (response) => {
        this.testPlanGenerating = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.$helperNotify.error(response?.ErrMsg || '测试计划生成失败')
          return
        }
        this.workflow = response.Data.workflow || this.workflow
        this.testPlanInfo = response.Data.test_plan || {}
        this.testPlanRunInfo = response.Data.test_run || {}
        this.loadTestRunHistory()
        this.$helperNotify.success('测试计划已生成')
      })
    },
    executeTestPlan(regeneratePlan) {
      if (!this.canGenerateArtifacts) {
        return
      }
      this.testRunExecuting = true
      taskWorkflowApi.TaskWorkflowTestRunExecute(this.workflowId, !!regeneratePlan, !!regeneratePlan, (response) => {
        this.testRunExecuting = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.$helperNotify.error(response?.ErrMsg || '接口测试执行失败')
          return
        }
        this.workflow = response.Data.workflow || this.workflow
        this.testPlanInfo = response.Data.test_plan || this.testPlanInfo
        this.testPlanRunInfo = response.Data.test_run || this.testPlanRunInfo
        this.loadCoverageInfo()
        this.loadTestPlanInfo()
        this.loadTestRunHistory()
        this.$helperNotify.success(regeneratePlan ? '覆盖检查和接口测试已执行' : '接口测试已执行')
      })
    },
    openUIAssistDialog() {
      this.uiAssistDialogVisible = true
      if (this.smartLinkOptions.length > 0) {
        return
      }
      smartLinkSet.SmartLinkList((response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.$helperNotify.error(response?.ErrMsg || 'SmartLink 列表加载失败')
          return
        }
        const rawList = Array.isArray(response.Data.smart_link_list) ? response.Data.smart_link_list : []
        this.smartLinkOptions = rawList.map(item => {
          let linkList = []
          try {
            linkList = JSON.parse(item.links || '[]')
          } catch (error) {
            linkList = []
          }
          return {
            ...item,
            linkList,
          }
        })
      })
    },
    onUIAssistSmartLinkChange() {
      const firstLabel = this.selectedSmartLinkLabels[0]?.label || ''
      this.uiAssistForm.label = firstLabel
    },
    submitUIAssist() {
      if (this.workflowId <= 0) {
        return
      }
      if (!this.uiAssistForm.smart_link_id || !this.uiAssistForm.label || !this.uiAssistForm.jump_url || !this.uiAssistForm.css_selector) {
        this.$helperNotify.error('请先补全页面辅助识别参数')
        return
      }
      this.uiAssistGenerating = true
      taskWorkflowApi.TaskWorkflowUIAssistGenerate({
        workflow_id: this.workflowId,
        smart_link_id: Number(this.uiAssistForm.smart_link_id || 0),
        label: this.uiAssistForm.label,
        jump_url: this.uiAssistForm.jump_url,
        css_selector: this.uiAssistForm.css_selector,
        wait_seconds: Number(this.uiAssistForm.wait_seconds || 5),
      }, (response) => {
        this.uiAssistGenerating = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.$helperNotify.error(response?.ErrMsg || '页面辅助识别失败')
          return
        }
        this.workflow = response.Data.workflow || this.workflow
        this.uiAssistInfo = response.Data.ui_assist || {}
        this.uiAssistRunInfo = response.Data.test_run || {}
        this.uiAssistDialogVisible = false
        this.loadTestRunHistory()
        this.$helperNotify.success('页面辅助识别已完成')
      })
    },
    selectHistoryItem(item) {
      this.activeHistoryRunId = Number(item.id || 0)
    },
    copyRequirementPrompt() {
      this.copyText(this.requirementPromptText, '需求文档提示词已复制')
    },
    copyDevPlanPrompt() {
      this.copyText(this.devPlanPromptText, '开发执行提示词已复制')
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
    formatJson(data) {
      return JSON.stringify(data || {}, null, 2)
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
