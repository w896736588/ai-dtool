<template>
  <div class="set-config-page">
    <!-- 工作日报 AI 设置 -->
    <div v-show="activeTab === 'daily-report' || !activeTab">
      <div class="set-config-header">
        <h3 class="set-config-title">工作日报 AI 设置</h3>
        <p class="set-config-desc">这里维护任务清单右侧"AI 生成工作日报"按钮使用的模型和提示词。</p>
      </div>

      <div class="set-config-table-card">
        <el-form label-width="120px" class="memory-config-form">
          <el-form-item label="日报模型">
            <el-select
              v-model="form.home_task_daily_report_model_id"
              clearable
              filterable
              style="width: 100%;"
              placeholder="请选择用于生成工作日报的 LLM 模型"
            >
              <el-option
                v-for="item in aiModelList"
                :key="item.id"
                :label="buildModelLabel(item)"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="日报提示词">
            <MdEditor
              v-model="form.home_task_daily_report_prompt"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              class="daily-report-editor"
            />
          </el-form-item>
          <el-form-item>
            <pl-button type="primary" @click="saveConfig">保存工作日报配置</pl-button>
            <pl-button @click="showChangeLog">改动记录</pl-button>
          </el-form-item>
        </el-form>
      </div>
    </div>

    <!-- 工作流提示词模板 -->
    <div v-show="activeTab === 'prompt-template'" class="prompt-template-section">
      <div class="set-config-header">
        <h3 class="set-config-title">工作流提示词模板</h3>
        <p class="set-config-desc">
          编辑工作流中使用的提示词模板，可点击下方占位符复制后粘贴到模板中。
        </p>
      </div>

      <div class="prompt-placeholder-bar">
        <span class="prompt-placeholder-bar__label">内置占位符：</span>
        <span
          v-for="ph in promptPlaceholders"
          :key="ph.value"
          class="prompt-placeholder-tag"
          @click="copyPlaceholder(ph)"
        >
          {{ ph.label }}
          <el-icon class="prompt-placeholder-tag__icon"><CopyDocument /></el-icon>
          <el-tooltip v-if="ph.tip" :content="ph.tip" placement="top">
            <el-icon class="prompt-placeholder-tag__help"><QuestionFilled /></el-icon>
          </el-tooltip>
        </span>
      </div>

      <div class="set-config-table-card prompt-template-card">
        <el-tabs v-model="activePromptTab" class="prompt-template-tabs">
          <el-tab-pane label="纯文本TAPD需求提示词" name="plain_text_requirement">
            <MdEditor
              v-model="form.home_task_prompt_plain_text_requirement"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              class="prompt-template-editor"
            />
          </el-tab-pane>

          <el-tab-pane label="需求分析设计提示词" name="dev">
            <MdEditor
              v-model="form.home_task_prompt_dev"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              class="prompt-template-editor"
            />
          </el-tab-pane>
          <el-tab-pane label="开发提示词" name="design">
            <MdEditor
              v-model="form.home_task_prompt_design"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              class="prompt-template-editor"
            />
          </el-tab-pane>
          <el-tab-pane label="接口生成提示词" name="api_gen">
            <MdEditor
              v-model="form.home_task_prompt_api_gen"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              class="prompt-template-editor"
            />
          </el-tab-pane>
          <el-tab-pane label="接口自动化测试提示词" name="api_test">
            <MdEditor
              v-model="form.home_task_prompt_api_test"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              class="prompt-template-editor"
            />
          </el-tab-pane>
          <el-tab-pane label="需求核对浏览器测试提示词" name="browser_test">
            <MdEditor
              v-model="form.home_task_prompt_browser_test"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              class="prompt-template-editor"
            />
          </el-tab-pane>
          <el-tab-pane label="代码检查提示词" name="code_review">
            <MdEditor
              v-model="form.home_task_prompt_code_review"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              class="prompt-template-editor"
            />
          </el-tab-pane>

          <el-tab-pane label="问题修改提示词" name="issue_fix">
            <MdEditor
              v-model="form.home_task_prompt_issue_fix"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              class="prompt-template-editor"
            />
          </el-tab-pane>
        </el-tabs>
        <div class="prompt-template-footer">
          <pl-button type="primary" @click="savePromptConfig">保存提示词模板配置</pl-button>
          <pl-button @click="showChangeLog">改动记录</pl-button>
        </div>
      </div>
    </div>

    <!-- 开发环境 -->
    <div v-show="activeTab === 'dev-environment'" class="dev-environment-section">
      <div class="set-config-header">
        <h3 class="set-config-title">开发环境</h3>
        <p class="set-config-desc">
          描述当前项目的开发环境信息，支持 Markdown 语法。保存后可在工作流提示词中使用
          <code style="color:#3d6b3d;background:#eef4ec;padding:1px 6px;border-radius:3px;">{开发环境}</code>
          占位符引用此内容。
        </p>
      </div>

      <div class="prompt-placeholder-bar">
        <span class="prompt-placeholder-bar__label">内置占位符：</span>
        <span
          v-for="ph in devEnvironmentPlaceholders"
          :key="ph.value"
          class="prompt-placeholder-tag"
          @click="copyPlaceholder(ph)"
        >
          {{ ph.label }}
          <el-icon class="prompt-placeholder-tag__icon"><CopyDocument /></el-icon>
          <el-tooltip v-if="ph.tip" :content="ph.tip" placement="top">
            <el-icon class="prompt-placeholder-tag__help"><QuestionFilled /></el-icon>
          </el-tooltip>
        </span>
      </div>

      <div class="set-config-table-card dev-environment-card">
        <MdEditor
          v-model="form.home_task_dev_environment"
          preview-theme="github"
          :preview="true"
          :toolbars="promptEditorToolbars"
          class="dev-environment-editor"
        />
        <div class="dev-environment-footer">
          <pl-button type="primary" @click="saveDevEnvironmentConfig">保存开发环境配置</pl-button>
          <pl-button @click="showChangeLog">改动记录</pl-button>
        </div>
      </div>
    </div>

    <!-- 需求抓取配置 -->
    <div v-show="activeTab === 'requirement-fetch'">
      <div class="set-config-header">
        <h3 class="set-config-title">需求抓取配置</h3>
        <p class="set-config-desc">
          维护需求抓取入口和页面解析规则，新建任务后会根据抓取类型选择对应配置执行。
        </p>
      </div>

      <div class="set-config-table-card">
        <el-tabs v-model="activeRequirementFetchTab">
          <el-tab-pane label="TAPD 抓取配置" name="tapd">
            <el-form label-width="120px" class="memory-config-form">
              <el-form-item label="自定义网页">
                <el-select
                  v-model="form.home_task_tapd_smart_link_id"
                  clearable
                  filterable
                  style="width: 100%;"
                  placeholder="请选择自定义网页"
                  @change="onRequirementSmartLinkChange('tapd')"
                >
                  <el-option
                    v-for="item in smartLinkList"
                    :key="item.id"
                    :label="item.name"
                    :value="item.id"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="网页链接">
                <el-select
                  v-model="form.home_task_tapd_link_label"
                  clearable
                  filterable
                  style="width: 100%;"
                  placeholder="请选择具体链接"
                >
                  <el-option
                    v-for="(link, idx) in currentTapdLinkOptions"
                    :key="idx"
                    :label="link.label"
                    :value="link.label"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="CSS选择器">
                <el-input
                  v-model="form.home_task_tapd_css_selector"
                  placeholder="如 .content-wrapper 或 #main"
                />
              </el-form-item>
              <el-form-item label="抓取前等待秒数">
                <el-input-number
                  v-model="form.home_task_tapd_wait_seconds"
                  :min="1"
                  :max="30"
                />
              </el-form-item>
              <el-form-item>
                <pl-button type="primary" @click="saveRequirementFetchConfig">保存 TAPD 抓取配置</pl-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>

          <el-tab-pane label="禅道 抓取配置" name="zentao">
            <el-form label-width="120px" class="memory-config-form">
              <el-form-item label="自定义网页">
                <el-select
                  v-model="form.home_task_zentao_smart_link_id"
                  clearable
                  filterable
                  style="width: 100%;"
                  placeholder="请选择自定义网页"
                  @change="onRequirementSmartLinkChange('zentao')"
                >
                  <el-option
                    v-for="item in smartLinkList"
                    :key="item.id"
                    :label="item.name"
                    :value="item.id"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="网页链接">
                <el-select
                  v-model="form.home_task_zentao_link_label"
                  clearable
                  filterable
                  style="width: 100%;"
                  placeholder="请选择具体链接"
                >
                  <el-option
                    v-for="(link, idx) in currentZentaoLinkOptions"
                    :key="idx"
                    :label="link.label"
                    :value="link.label"
                  />
                </el-select>
              </el-form-item>
              <el-form-item label="CSS选择器">
                <el-input
                  v-model="form.home_task_zentao_css_selector"
                  placeholder="如 .content-wrapper 或 #main"
                />
              </el-form-item>
              <el-form-item label="抓取前等待秒数">
                <el-input-number
                  v-model="form.home_task_zentao_wait_seconds"
                  :min="1"
                  :max="30"
                />
              </el-form-item>
              <el-form-item>
                <pl-button type="primary" @click="saveRequirementFetchConfig">保存禅道抓取配置</pl-button>
              </el-form-item>
            </el-form>
          </el-tab-pane>
        </el-tabs>
      </div>
    </div>

    <!-- 分支名生成提示词 -->
    <div v-show="activeTab === 'branch-name'">
      <div class="set-config-header">
        <h3 class="set-config-title">分支名生成提示词</h3>
        <p class="set-config-desc">
          用于生成分支名称的提示词模板，支持 Markdown 语法。
        </p>
      </div>

      <div class="prompt-placeholder-bar">
        <span class="prompt-placeholder-bar__label">内置占位符：</span>
        <span
          v-for="ph in branchNamePlaceholders"
          :key="ph.value"
          class="prompt-placeholder-tag"
          @click="copyPlaceholder(ph)"
        >
          {{ ph.label }}
          <el-icon class="prompt-placeholder-tag__icon"><CopyDocument /></el-icon>
          <el-tooltip v-if="ph.tip" :content="ph.tip" placement="top">
            <el-icon class="prompt-placeholder-tag__help"><QuestionFilled /></el-icon>
          </el-tooltip>
        </span>
      </div>

      <div class="set-config-table-card">
        <el-form label-width="120px" class="memory-config-form">
          <el-form-item label="生成模型">
            <el-select
              v-model="form.home_task_branch_name_model_id"
              clearable
              filterable
              style="width: 100%;"
              placeholder="请选择用于生成分支名的 LLM 模型"
            >
              <el-option
                v-for="item in aiModelList"
                :key="item.id"
                :label="buildModelLabel(item)"
                :value="item.id"
              />
            </el-select>
          </el-form-item>
          <el-form-item label="提示词">
            <MdEditor
              v-model="form.home_task_branch_name_prompt"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              class="branch-name-editor"
              style="height: 360px;"
            />
          </el-form-item>
          <el-form-item>
            <pl-button type="primary" @click="saveBranchNameConfig">保存分支名生成提示词</pl-button>
            <pl-button @click="showChangeLog">改动记录</pl-button>
          </el-form-item>
        </el-form>
      </div>
    </div>

    <!-- 提示词改动记录弹窗 -->
    <el-dialog v-model="changeLogVisible" title="提示词改动记录" width="720px" >
      <el-table :data="changeLogList" stripe max-height="480">
        <el-table-column prop="config_name" label="配置项" width="160" />
        <el-table-column prop="create_time_desc" label="改动时间" width="170" />
        <el-table-column label="操作" width="80" fixed="right">
          <template #default="{ row }">
            <el-button type="primary" link size="small" @click="showChangeDetail(row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
      <div v-if="changeLogList.length === 0" style="text-align:center;color:#999;padding:24px 0;">暂无改动记录</div>
    </el-dialog>

    <!-- 改动详情弹窗 -->
    <el-dialog v-model="changeDetailVisible" :title="changeDetailTitle" width="720px" >
      <div style="display:flex;gap:16px;">
        <div style="flex:1;min-width:0;">
          <div style="font-weight:bold;margin-bottom:8px;color:#e6a23c;">修改前：</div>
          <div style="background:#fdf6ec;padding:12px;border-radius:6px;max-height:360px;overflow-y:auto;white-space:pre-wrap;word-break:break-all;font-size:13px;">{{ changeDetailOld }}</div>
        </div>
        <div style="flex:1;min-width:0;">
          <div style="font-weight:bold;margin-bottom:8px;color:#67c23a;">修改后：</div>
          <div style="background:#f0f9eb;padding:12px;border-radius:6px;max-height:360px;overflow-y:auto;white-space:pre-wrap;word-break:break-all;font-size:13px;">{{ changeDetailNew }}</div>
        </div>
      </div>
    </el-dialog>
  </div>
</template>

<script>
import set from '@/utils/base/git_set'
import AiSetApi from '@/utils/base/ai_set'
import SmartLinkSet from '@/utils/base/smart_link_set'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { CopyDocument, QuestionFilled } from '@element-plus/icons-vue'

const DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT = '请基于当前活跃任务生成中文工作日报，按已完成、进行中、风险与阻塞三个部分总结，输出 Markdown，禁止编造未提供的信息。'

const PROMPT_PLACEHOLDERS = [
  { label: '任务名称', value: '{任务名称}', tip: '替换为当前任务的名称' },
  { label: '需求文档地址', value: '{需求文档地址}', tip: '替换为需求知识片段的分享链接' },
  { label: '需求文档纯文本地址', value: '{需求文档纯文本地址}', tip: '替换为纯文本需求片段的分享链接' },
  { label: '需求文档纯文本文件相对地址', value: '{需求文档纯文本文件相对地址}', tip: '替换为纯文本需求片段文件的相对路径' },
  { label: '需求设计方案文档地址', value: '{需求设计方案文档地址}', tip: '替换为设计方案片段的分享链接' },
  { label: '需求设计方案文件相对地址', value: '{需求设计方案文件相对地址}', tip: '替换为设计方案片段文件的相对路径' },
  { label: '接口开发API地址', value: '{接口开发API地址}', tip: '替换为当前服务的 API 基地址（scheme://host）' },
  { label: '接口开发API的token', value: '{接口开发API的token}', tip: '替换为请求的 Authorization token' },
  { label: '开发项目配置', value: '{开发项目配置}', tip: '替换为开发项目配置的 Markdown 列表' },
  { label: '自定义网页', value: '{自定义网页}', tip: '替换为智能链接（smart_link）的名称和 ID' },
  { label: '网页标签', value: '{网页标签}', tip: '替换为智能链接的标签（smart_link_label）' },
  { label: '账号', value: '{账号}', tip: '替换为智能链接的账号（smart_link_account）' },
  { label: 'dtool-api地址', value: '{dtool-api地址}', tip: '替换为 skills/dtool-api 目录的本地路径' },
  { label: 'dtool-common地址', value: '{dtool-common地址}', tip: '替换为 skills/dtool-common 目录的本地路径' },
  { label: 'dtool-workflow地址', value: '{dtool-workflow地址}', tip: '替换为 skills/dtool-workflow 目录的本地路径' },
  { label: 'dtool-playwright地址', value: '{dtool-playwright地址}', tip: '替换为 skills/dtool-playwright 目录的本地路径' },
  { label: 'dtool-notify地址', value: '{dtool-notify地址}', tip: '替换为 skills/dtool-notify 目录的本地路径' },
  { label: '工作流程ID', value: '{工作流程ID}', tip: '替换为当前工作流程的 ID' },
  { label: '任务ID', value: '{任务ID}', tip: '替换为当前任务的 ID' },
  { label: '开发环境', value: '{开发环境}', tip: '替换为开发环境配置（已递归解析内部占位符）' },
]

const PROMPT_EDITOR_TOOLBARS = [
  'bold', 'italic', 'strikeThrough', 'title', 'quote',
  'unorderedList', 'orderedList', 'task', 'link', 'code',
  'codeRow', 'table', 'preview', 'fullscreen',
]

export default {
  name: 'HomeTaskReportSetting',
  emits: ['changed'],
  props: {
    activeTab: {
      type: String,
      default: '',
    },
  },
  data() {
    return {
      aiModelList: [],
      smartLinkList: [],
      activePromptTab: 'dev',
      activeRequirementFetchTab: 'tapd',
      form: {
        home_task_daily_report_model_id: null,
        home_task_daily_report_prompt: DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT,
        home_task_fragment_prompt: '',
        home_task_prompt_dev: '',
        home_task_prompt_api_gen: '',
        home_task_prompt_api_test: '',
        home_task_prompt_browser_test: '',
        home_task_prompt_code_review: '',
        home_task_prompt_design: '',
        home_task_tapd_smart_link_id: null,
        home_task_tapd_link_label: '',
        home_task_tapd_css_selector: '',
        home_task_tapd_wait_seconds: 5,
        home_task_zentao_smart_link_id: null,
        home_task_zentao_link_label: '',
        home_task_zentao_css_selector: '',
        home_task_zentao_wait_seconds: 5,
        home_task_dev_environment: '',
        home_task_branch_name_prompt: '',
        home_task_branch_name_model_id: null,
        home_task_prompt_plain_text_requirement: '',
        home_task_prompt_design_plan_requirement: '',
        home_task_prompt_issue_fix: '',
      },
      promptPlaceholders: PROMPT_PLACEHOLDERS,
      promptEditorToolbars: PROMPT_EDITOR_TOOLBARS,
      changeLogVisible: false,
      changeLogList: [],
      changeDetailVisible: false,
      changeDetailTitle: '',
      changeDetailOld: '',
      changeDetailNew: '',
    }
  },
  computed: {
    branchNamePlaceholders() {
      return [
        { label: '需求名', value: '{需求名}', tip: '替换为任务名称' },
        { label: '父分支', value: '{父分支}', tip: '替换为父分支名称' },
        { label: '任务创建日期', value: '{任务创建日期}', tip: '替换为任务创建日期' },
      ]
    },
    devEnvironmentPlaceholders() {
      return PROMPT_PLACEHOLDERS.filter(ph => ph.value !== '{开发环境}')
    },
    currentTapdLinkOptions() {
      return this.getSmartLinkOptions(this.form.home_task_tapd_smart_link_id)
    },
    currentZentaoLinkOptions() {
      return this.getSmartLinkOptions(this.form.home_task_zentao_smart_link_id)
    },
  },
  mounted() {
    this.loadAiModelList()
    this.loadSmartLinkList()
    this.loadConfig()
  },
  methods: {
    buildModelLabel(item) {
      const provider = item.provider_name || '未命名服务商'
      const model = item.name || item.model || `模型#${item.id}`
      return `${provider} / ${model}`
    },
    getSmartLinkOptions(smartLinkId) {
      if (!smartLinkId) return []
      const smartLink = this.smartLinkList.find(item => item.id === smartLinkId)
      if (!smartLink || !smartLink.links) return []
      try {
        return JSON.parse(smartLink.links)
      } catch {
        return []
      }
    },
    onRequirementSmartLinkChange(type) {
      if (type === 'zentao') {
        this.form.home_task_zentao_link_label = ''
        return
      }
      this.form.home_task_tapd_link_label = ''
    },
    loadAiModelList() {
      AiSetApi.AiModelList({ model_type: 'llm' }, (response) => {
        if (response.ErrCode !== 0) {
          return
        }
        this.aiModelList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    loadSmartLinkList() {
      SmartLinkSet.SmartLinkList((response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.smartLinkList = Array.isArray(response.Data.smart_link_list) ? response.Data.smart_link_list : []
      })
    },
    loadConfig() {
      set.HomeTaskConfigGet((response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.form.home_task_daily_report_model_id = response.Data.home_task_daily_report_model_id || null
        this.form.home_task_daily_report_prompt = response.Data.home_task_daily_report_prompt || DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT
        this.form.home_task_fragment_prompt = response.Data.home_task_fragment_prompt || ''
        this.form.home_task_prompt_dev = response.Data.home_task_prompt_dev || ''
        this.form.home_task_prompt_api_gen = response.Data.home_task_prompt_api_gen || ''
        this.form.home_task_prompt_api_test = response.Data.home_task_prompt_api_test || ''
        this.form.home_task_prompt_browser_test = response.Data.home_task_prompt_browser_test || ''
        this.form.home_task_prompt_code_review = response.Data.home_task_prompt_code_review || ''
        this.form.home_task_prompt_design = response.Data.home_task_prompt_design || ''
        this.form.home_task_tapd_smart_link_id = response.Data.home_task_tapd_smart_link_id || null
        this.form.home_task_tapd_link_label = response.Data.home_task_tapd_link_label || ''
        this.form.home_task_tapd_css_selector = response.Data.home_task_tapd_css_selector || ''
        this.form.home_task_tapd_wait_seconds = response.Data.home_task_tapd_wait_seconds || 5
        this.form.home_task_zentao_smart_link_id = response.Data.home_task_zentao_smart_link_id || null
        this.form.home_task_zentao_link_label = response.Data.home_task_zentao_link_label || ''
        this.form.home_task_zentao_css_selector = response.Data.home_task_zentao_css_selector || ''
        this.form.home_task_zentao_wait_seconds = response.Data.home_task_zentao_wait_seconds || 5
        this.form.home_task_dev_environment = response.Data.home_task_dev_environment || ''
        this.form.home_task_branch_name_prompt = response.Data.home_task_branch_name_prompt || ''
        this.form.home_task_branch_name_model_id = response.Data.home_task_branch_name_model_id || null
        this.form.home_task_prompt_plain_text_requirement = response.Data.home_task_prompt_plain_text_requirement || ''
        this.form.home_task_prompt_design_plan_requirement = response.Data.home_task_prompt_design_plan_requirement || ''
        this.form.home_task_prompt_issue_fix = response.Data.home_task_prompt_issue_fix || ''
      })
    },
    saveConfig() {
      const payload = this.buildFullPayload()
      set.HomeTaskConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('配置已保存')
          this.$emit('changed')
        }
      })
    },
    saveRequirementFetchConfig() {
      const payload = this.buildFullPayload()
      set.HomeTaskConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('需求抓取配置已保存')
          this.$emit('changed')
        }
      })
    },
    savePromptConfig() {
      const payload = this.buildFullPayload()
      set.HomeTaskConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('提示词模板配置已保存')
          this.$emit('changed')
        }
      })
    },
    saveDevEnvironmentConfig() {
      const payload = this.buildFullPayload()
      set.HomeTaskConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('开发环境配置已保存')
          this.$emit('changed')
        }
      })
    },
    saveBranchNameConfig() {
      const payload = this.buildFullPayload()
      set.HomeTaskConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('分支名生成提示词已保存')
          this.$emit('changed')
        }
      })
    },
    buildFullPayload() {
      return {
        home_task_daily_report_model_id: this.form.home_task_daily_report_model_id,
        home_task_daily_report_prompt: this.form.home_task_daily_report_prompt,
        home_task_fragment_prompt: this.form.home_task_fragment_prompt,
        home_task_prompt_dev: this.form.home_task_prompt_dev,
        home_task_prompt_api_gen: this.form.home_task_prompt_api_gen,
        home_task_prompt_api_test: this.form.home_task_prompt_api_test,
        home_task_prompt_browser_test: this.form.home_task_prompt_browser_test,
        home_task_prompt_code_review: this.form.home_task_prompt_code_review,
        home_task_prompt_design: this.form.home_task_prompt_design,
        home_task_tapd_smart_link_id: this.form.home_task_tapd_smart_link_id,
        home_task_tapd_link_label: this.form.home_task_tapd_link_label,
        home_task_tapd_css_selector: this.form.home_task_tapd_css_selector,
        home_task_tapd_wait_seconds: this.form.home_task_tapd_wait_seconds,
        home_task_zentao_smart_link_id: this.form.home_task_zentao_smart_link_id,
        home_task_zentao_link_label: this.form.home_task_zentao_link_label,
        home_task_zentao_css_selector: this.form.home_task_zentao_css_selector,
        home_task_zentao_wait_seconds: this.form.home_task_zentao_wait_seconds,
        home_task_dev_environment: this.form.home_task_dev_environment,
        home_task_branch_name_prompt: this.form.home_task_branch_name_prompt,
        home_task_branch_name_model_id: this.form.home_task_branch_name_model_id,
        home_task_prompt_plain_text_requirement: this.form.home_task_prompt_plain_text_requirement,
        home_task_prompt_design_plan_requirement: this.form.home_task_prompt_design_plan_requirement,
        home_task_prompt_issue_fix: this.form.home_task_prompt_issue_fix,
      }
    },
    copyPlaceholder(placeholder) {
      const text = placeholder.value
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(text).then(() => {
          this.$helperNotify.success(`已复制：${text}`)
        }).catch(() => {
          this.fallbackCopy(text)
        })
      } else {
        this.fallbackCopy(text)
      }
    },
    showChangeLog() {
      set.PromptChangeLogList((response) => {
        if (response.ErrCode === 0) {
          this.changeLogList = Array.isArray(response.Data) ? response.Data : []
          this.changeLogVisible = true
        }
      })
    },
    showChangeDetail(row) {
      this.changeDetailTitle = row.config_name + ' - ' + row.create_time_desc
      this.changeDetailOld = row.old_value || ''
      this.changeDetailNew = row.new_value || ''
      this.changeDetailVisible = true
    },
    fallbackCopy(text) {
      const textarea = document.createElement('textarea')
      textarea.value = text
      textarea.style.position = 'fixed'
      textarea.style.opacity = '0'
      document.body.appendChild(textarea)
      textarea.select()
      document.execCommand('copy')
      document.body.removeChild(textarea)
      this.$helperNotify.success(`已复制：${text}`)
    },
  },
  components: {
    MdEditor,
    CopyDocument,
  },
}
</script>

<style scoped src="@/css/components/set/home_task_report.css"></style>

<style>
.set-config-page {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.prompt-template-section {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  height: 100%;
  min-height: 0;
}

.prompt-template-card {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 0;
}

.prompt-template-card .el-tabs {
  display: flex;
  flex-direction: column;
  flex: 1 1 auto;
  min-height: 0;
}

.prompt-template-card .el-tabs__header {
  /* 二级 Tab 头部必须固定在内容上方，避免被 flex 顺序挤到底部。 */
  /* Keep nested tab headers above the content instead of being pushed to the bottom. */
  order: -1;
  flex-shrink: 0;
}

.prompt-template-card .el-tabs__content {
  flex: 1 1 auto;
  min-height: 0;
  overflow: hidden;
}

.prompt-template-card .el-tab-pane {
  height: 100%;
}

.daily-report-editor,
.prompt-template-editor,
.branch-name-editor {
  height: calc(100vh - 460px);
  min-height: 200px;
  max-height: 800px;
}

.daily-report-editor :deep(.md-editor),
.prompt-template-editor :deep(.md-editor),
.branch-name-editor :deep(.md-editor) {
  --md-font-size: 13px;
  --md-code-font-size: 13px;
  --md-font-family: Consolas, Avenir, Helvetica, Arial, sans-serif;
}

.prompt-template-editor {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.prompt-template-editor :deep(.md-editor) {
  width: 100%;
  height: 100%;
  min-height: 0;
}

.daily-report-editor :deep(.md-editor-input),
.daily-report-editor :deep(.md-editor-preview-wrapper),
.daily-report-editor :deep(.cm-content),
.daily-report-editor :deep(.md-editor-preview),
.prompt-template-editor :deep(.md-editor-input),
.prompt-template-editor :deep(.md-editor-preview-wrapper),
.prompt-template-editor :deep(.cm-content),
.prompt-template-editor :deep(.md-editor-preview),
.branch-name-editor :deep(.md-editor-input),
.branch-name-editor :deep(.md-editor-preview-wrapper),
.branch-name-editor :deep(.cm-content),
.branch-name-editor :deep(.md-editor-preview) {
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif !important;
  font-size: 13px !important;
  line-height: 1.6;
}

.daily-report-editor :deep(.md-editor-preview p),
.daily-report-editor :deep(.md-editor-preview li),
.daily-report-editor :deep(.md-editor-preview blockquote),
.daily-report-editor :deep(.md-editor-preview table),
.daily-report-editor :deep(.md-editor-preview td),
.daily-report-editor :deep(.md-editor-preview th),
.daily-report-editor :deep(.md-editor-preview code),
.prompt-template-editor :deep(.md-editor-preview p),
.prompt-template-editor :deep(.md-editor-preview li),
.prompt-template-editor :deep(.md-editor-preview blockquote),
.prompt-template-editor :deep(.md-editor-preview table),
.prompt-template-editor :deep(.md-editor-preview td),
.prompt-template-editor :deep(.md-editor-preview th),
.prompt-template-editor :deep(.md-editor-preview code),
.branch-name-editor :deep(.md-editor-preview p),
.branch-name-editor :deep(.md-editor-preview li),
.branch-name-editor :deep(.md-editor-preview blockquote),
.branch-name-editor :deep(.md-editor-preview table),
.branch-name-editor :deep(.md-editor-preview td),
.branch-name-editor :deep(.md-editor-preview th),
.branch-name-editor :deep(.md-editor-preview code) {
  font-size: 13px !important;
}

.prompt-template-footer {
  padding-top: 12px;
  flex-shrink: 0;
}

.dev-environment-section {
  display: flex;
  flex-direction: column;
  height: 100%;
  min-height: 0;
}

.dev-environment-card {
  display: flex;
  flex: 1 1 auto;
  flex-direction: column;
  min-height: 0;
}

.dev-environment-editor {
  flex: 1 1 auto;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.dev-environment-editor :deep(.md-editor) {
  --md-font-size: 13px;
  --md-code-font-size: 13px;
  --md-font-family: Consolas, Avenir, Helvetica, Arial, sans-serif;
  width: 100%;
  height: 100%;
  min-height: 0;
}

.dev-environment-editor :deep(.md-editor-input),
.dev-environment-editor :deep(.md-editor-preview-wrapper),
.dev-environment-editor :deep(.cm-content),
.dev-environment-editor :deep(.md-editor-preview) {
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif !important;
  font-size: 13px !important;
  line-height: 1.6;
}

.dev-environment-editor :deep(.md-editor-preview p),
.dev-environment-editor :deep(.md-editor-preview li),
.dev-environment-editor :deep(.md-editor-preview blockquote),
.dev-environment-editor :deep(.md-editor-preview table),
.dev-environment-editor :deep(.md-editor-preview td),
.dev-environment-editor :deep(.md-editor-preview th),
.dev-environment-editor :deep(.md-editor-preview code) {
  font-size: 13px !important;
}

.dev-environment-footer {
  padding-top: 12px;
  flex-shrink: 0;
}

/* MdEditor 预览区域绿色滚动条 —— 覆盖 md-editor-v3 的 CSS 变量 */
.set-config-page .md-editor {
  --md-scrollbar-bg-color: #edf3e8;
  --md-scrollbar-thumb-color: #9fb39a;
  --md-scrollbar-thumb-hover-color: #869c82;
  --md-scrollbar-thumb-active-color: #7a8f76;
}

/* 同时覆盖原生滚动条样式 */
.set-config-page .md-editor .md-editor-preview ::-webkit-scrollbar {
  width: 10px !important;
  height: 10px !important;
}

.set-config-page .md-editor .md-editor-preview ::-webkit-scrollbar-track {
  background: #edf3e8 !important;
  border-radius: 999px !important;
}

.set-config-page .md-editor .md-editor-preview ::-webkit-scrollbar-thumb {
  background: #9fb39a !important;
  border: 2px solid #edf3e8 !important;
  border-radius: 999px !important;
}

.set-config-page .md-editor .md-editor-preview ::-webkit-scrollbar-thumb:hover {
  background: #869c82 !important;
}

.set-config-page .md-editor .md-editor-preview ::-webkit-scrollbar-corner {
  background: #edf3e8 !important;
}
</style>

<style>
.set-config-page .daily-report-editor .md-editor-preview,
.set-config-page .daily-report-editor .md-editor-preview-wrapper,
.set-config-page .daily-report-editor .md-editor-html,
.set-config-page .prompt-template-editor .md-editor-preview,
.set-config-page .prompt-template-editor .md-editor-preview-wrapper,
.set-config-page .prompt-template-editor .md-editor-html,
.set-config-page .branch-name-editor .md-editor-preview,
.set-config-page .branch-name-editor .md-editor-preview-wrapper,
.set-config-page .branch-name-editor .md-editor-html,
.set-config-page .dev-environment-editor .md-editor-preview,
.set-config-page .dev-environment-editor .md-editor-preview-wrapper,
.set-config-page .dev-environment-editor .md-editor-html {
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif !important;
  font-size: 13px !important;
  line-height: 1.6 !important;
}

.set-config-page .daily-report-editor .md-editor-preview *,
.set-config-page .daily-report-editor .md-editor-html *,
.set-config-page .prompt-template-editor .md-editor-preview *,
.set-config-page .prompt-template-editor .md-editor-html *,
.set-config-page .branch-name-editor .md-editor-preview *,
.set-config-page .branch-name-editor .md-editor-html *,
.set-config-page .dev-environment-editor .md-editor-preview *,
.set-config-page .dev-environment-editor .md-editor-html * {
  font-size: inherit !important;
}

.set-config-page .daily-report-editor .cm-content,
.set-config-page .daily-report-editor .md-editor-input,
.set-config-page .prompt-template-editor .cm-content,
.set-config-page .prompt-template-editor .md-editor-input,
.set-config-page .branch-name-editor .cm-content,
.set-config-page .branch-name-editor .md-editor-input,
.set-config-page .dev-environment-editor .cm-content,
.set-config-page .dev-environment-editor .md-editor-input {
  font-family: Consolas, Avenir, Helvetica, Arial, sans-serif !important;
  font-size: 13px !important;
  line-height: 1.6 !important;
}
</style>
