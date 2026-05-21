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
    <div v-show="activeTab === 'dev-environment'">
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
        </span>
      </div>

      <div class="set-config-table-card">
        <MdEditor
          v-model="form.home_task_dev_environment"
          preview-theme="github"
          :preview="true"
          :toolbars="promptEditorToolbars"
          style="height: 360px;"
        />
        <div style="padding-top: 12px;">
          <pl-button type="primary" @click="saveDevEnvironmentConfig">保存开发环境配置</pl-button>
          <pl-button @click="showChangeLog">改动记录</pl-button>
        </div>
      </div>
    </div>

    <!-- TAPD 需求抓取配置 -->
    <div v-show="activeTab === 'tapd'">
      <div class="set-config-header">
        <h3 class="set-config-title">TAPD 需求抓取配置</h3>
        <p class="set-config-desc">
          从自定义网页中选择一个链接，用于在任务中快速跳转到 TAPD 登录页。
        </p>
      </div>

      <div class="set-config-table-card">
        <el-form label-width="120px" class="memory-config-form">
          <el-form-item label="自定义网页">
            <el-select
              v-model="form.home_task_tapd_smart_link_id"
              clearable
              filterable
              style="width: 100%;"
              placeholder="请选择自定义网页"
              @change="onSmartLinkChange"
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
                v-for="(link, idx) in currentLinkOptions"
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
          <el-form-item label="等待秒数">
            <el-input-number
              v-model="form.home_task_tapd_wait_seconds"
              :min="1"
              :max="30"
            />
          </el-form-item>
          <el-form-item>
            <pl-button type="primary" @click="saveTapdConfig">保存 TAPD 需求抓取配置</pl-button>
          </el-form-item>
        </el-form>
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
import { CopyDocument } from '@element-plus/icons-vue'

const DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT = '请基于当前活跃任务生成中文工作日报，按已完成、进行中、风险与阻塞三个部分总结，输出 Markdown，禁止编造未提供的信息。'

const PROMPT_PLACEHOLDERS = [
  { label: '需求文档地址', value: '{需求文档地址}' },
  { label: '需求文档纯文本地址', value: '{需求文档纯文本地址}' },
  { label: '需求文档纯文本文件相对地址', value: '{需求文档纯文本文件相对地址}' },
  { label: '需求设计方案文档地址', value: '{需求设计方案文档地址}' },
  { label: '需求设计方案文件相对地址', value: '{需求设计方案文件相对地址}' },
  { label: '接口开发API地址', value: '{接口开发API地址}' },
  { label: '接口开发API的token', value: '{接口开发API的token}' },
  { label: '开发项目配置', value: '{开发项目配置}' },
  { label: '自定义网页', value: '{自定义网页}' },
  { label: '网页标签', value: '{网页标签}' },
  { label: '账号', value: '{账号}' },
  { label: 'dtool-api地址', value: '{dtool-api地址}' },
  { label: 'dtool-common地址', value: '{dtool-common地址}' },
  { label: 'dtool-workflow地址', value: '{dtool-workflow地址}' },
  { label: 'dtool-playwright地址', value: '{dtool-playwright地址}' },
  { label: 'dtool-notify地址', value: '{dtool-notify地址}' },
  { label: '工作流程ID', value: '{工作流程ID}' },
  { label: '开发环境', value: '{开发环境}' },
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
        home_task_tapd_wait_seconds: 3,
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
        { label: '需求名', value: '{需求名}' },
        { label: '父分支', value: '{父分支}' },
      ]
    },
    devEnvironmentPlaceholders() {
      return PROMPT_PLACEHOLDERS.filter(ph => ph.value !== '{开发环境}')
    },
    currentLinkOptions() {
      if (!this.form.home_task_tapd_smart_link_id) return []
      const smartLink = this.smartLinkList.find(item => item.id === this.form.home_task_tapd_smart_link_id)
      if (!smartLink || !smartLink.links) return []
      try {
        return JSON.parse(smartLink.links)
      } catch {
        return []
      }
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
    onSmartLinkChange() {
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
        this.form.home_task_tapd_wait_seconds = response.Data.home_task_tapd_wait_seconds || 3
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
    saveTapdConfig() {
      const payload = this.buildFullPayload()
      set.HomeTaskConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('TAPD 需求抓取配置已保存')
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
.prompt-template-section {
  display: flex;
  flex-direction: column;
}

.prompt-template-card {
  display: flex;
  flex-direction: column;
}

.prompt-template-card .el-tabs {
  display: flex;
  flex-direction: column;
}

.prompt-template-card .el-tabs__header {
  /* 二级 Tab 头部必须固定在内容上方，避免被 flex 顺序挤到底部。 */
  /* Keep nested tab headers above the content instead of being pushed to the bottom. */
  order: -1;
  flex-shrink: 0;
}

.prompt-template-card .el-tabs__content {
  overflow: visible;
}

.prompt-template-card .el-tab-pane {
  height: auto;
}

.prompt-template-editor {
  height: calc(100vh - 460px);
  min-height: 200px;
  max-height: 800px;
}

.prompt-template-footer {
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
