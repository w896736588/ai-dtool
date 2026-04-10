<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">{{ pageTitle }}</h3>
      <p class="set-config-desc">{{ pageDesc }}</p>
      <div class="set-config-actions">
        <template v-if="showRuntimeConfig">
          <pl-button v-if="!runtimeEditMode" type="primary" @click="startRuntimeEdit">编辑配置</pl-button>
          <template v-else>
            <pl-button @click="cancelRuntimeEdit">取消</pl-button>
            <pl-button type="primary" @click="saveRuntimeConfig">保存并重新加载</pl-button>
          </template>
        </template>
        <pl-button v-else type="primary" @click="saveAiConfig">保存 AI 配置</pl-button>
      </div>
    </div>

    <div class="set-config-table-card">
      <el-alert
        v-if="showRuntimeConfig"
        :closable="false"
        :type="runtimeConfigAlertType"
        :title="memoryConfigAlertTitle"
      />

      <template v-if="showRuntimeConfig">
        <el-divider content-position="left">配置文件</el-divider>
        <el-descriptions class="memory-config-display" :column="1" border>
          <el-descriptions-item label="当前文件">
            <div class="config-value">{{ form.memory_config_file || '-' }}</div>
            <div class="config-item-help">当前页面展示和编辑的配置都来自这个 ini 文件。</div>
          </el-descriptions-item>
        </el-descriptions>

        <div v-if="runtimeEditMode" class="memory-config-form">
          <el-divider content-position="left">[base] 主库配置</el-divider>
          <el-form label-width="150px">
            <el-form-item label="dbPath">
              <el-input v-model="runtimeEditForm.db_path" placeholder="请输入主库目录" />
              <div class="config-item-help">主库 sqlite 所在目录；未配置时默认使用 config/{AppName}。</div>
            </el-form-item>
            <el-form-item label="dbFileName">
              <el-input v-model="runtimeEditForm.db_file_name" placeholder="请输入主库文件名" />
              <div class="config-item-help">主库 sqlite 文件名；未配置时默认使用 {AppName}.db。</div>
            </el-form-item>
            <el-form-item label="dbIsGitRepo">
              <el-switch v-model="runtimeEditForm.db_is_git_repo" />
              <div class="config-item-help">开启后主库在使用前会 git pull，关闭程序时会自动 push。</div>
            </el-form-item>
          </el-form>

          <el-divider content-position="left">[base] 日志库配置</el-divider>
          <el-form label-width="150px">
            <el-form-item label="logDbPath">
              <el-input v-model="runtimeEditForm.log_db_path" placeholder="请输入日志库目录" />
              <div class="config-item-help">日志库 sqlite 所在目录；未配置时默认使用与主库相同目录。</div>
            </el-form-item>
          </el-form>

          <el-divider content-position="left">[base] 记忆库配置</el-divider>
          <el-form label-width="150px">
            <el-form-item label="memoryDbPath">
              <el-input v-model="runtimeEditForm.memory_db_path" placeholder="请输入记忆库目录" />
              <div class="config-item-help">记忆库 Markdown 根目录；未配置时记忆库不会初始化。</div>
            </el-form-item>
            <el-form-item label="memoryDbIsGitRepo">
              <el-switch v-model="runtimeEditForm.memory_db_is_git_repo" />
              <div class="config-item-help">开启后记忆库启动前会 git pull，自动同步时会 push。</div>
            </el-form-item>
            <el-form-item label="memoryDbAutoPushDelayMinutes">
              <el-input-number v-model="runtimeEditForm.memory_db_auto_push_delay_minutes" :min="0" :step="1" />
              <div class="config-item-help">知识片段写入本地 Markdown 后，延迟多少分钟自动 git commit + push；0 表示关闭自动 push。</div>
            </el-form-item>
          </el-form>

          <el-divider content-position="left">[path] 路径配置</el-divider>
          <el-form label-width="150px">
            <el-form-item label="webkit_driver_path">
              <el-input v-model="runtimeEditForm.webkit_driver_path" placeholder="请输入 webkit driver 目录" />
              <div class="config-item-help">Playwright WebKit 驱动目录，支持使用 {DRIVE} 占位符。</div>
            </el-form-item>
            <el-form-item label="webkit_data_path">
              <el-input v-model="runtimeEditForm.webkit_data_path" placeholder="请输入 webkit 数据目录" />
              <div class="config-item-help">浏览器运行数据目录，支持使用 {DRIVE} 占位符。</div>
            </el-form-item>
            <el-form-item label="webkit_download_path">
              <el-input v-model="runtimeEditForm.webkit_download_path" placeholder="请输入下载目录" />
              <div class="config-item-help">Playwright 文件下载目录，支持使用 {DRIVE} 占位符。</div>
            </el-form-item>
          </el-form>
        </div>

        <template v-else>
          <el-divider content-position="left">[base] 主库配置</el-divider>
          <el-descriptions class="memory-config-display" :column="1" border>
            <el-descriptions-item label="dbPath">
              <div class="config-value">{{ form.db_dir || '未配置，请在配置文件中设置' }}</div>
              <div class="config-item-help">主库 sqlite 所在目录；未配置时默认使用 config/{AppName}。</div>
            </el-descriptions-item>
            <el-descriptions-item label="dbFileName">
              <div class="config-value">{{ form.db_name || '未配置，请在配置文件中设置' }}</div>
              <div class="config-item-help">主库 sqlite 文件名；未配置时默认使用 {AppName}.db。</div>
            </el-descriptions-item>
            <el-descriptions-item label="dbIsGitRepo">
              <div class="config-value">{{ boolText(form.db_is_git_repo) }}</div>
              <div class="config-item-help">开启后主库在使用前会 git pull，关闭程序时会自动 push。</div>
              <div v-if="showMainDbSyncButton" class="config-item-actions">
                <GitActionButton
                  compact
                  variant="info"
                  :loading="syncLoading.main"
                  @click="syncDatabase(RUNTIME_DATABASE_SYNC_TARGET_MAIN)"
                >
                  同步
                </GitActionButton>
              </div>
            </el-descriptions-item>
          </el-descriptions>

          <el-divider content-position="left">[base] 日志库配置</el-divider>
          <el-descriptions class="memory-config-display" :column="1" border>
            <el-descriptions-item label="logDbPath">
              <div class="config-value">{{ form.log_db_path || '未配置（默认与主库相同）' }}</div>
              <div class="config-item-help">日志库 sqlite 所在目录；未配置时默认使用与主库相同目录。</div>
            </el-descriptions-item>
          </el-descriptions>

          <el-divider content-position="left">[base] 记忆库配置</el-divider>
          <el-descriptions class="memory-config-display" :column="1" border>
            <el-descriptions-item label="memoryDbPath">
              <div class="config-value">{{ form.memory_dir || '未配置，请在配置文件中设置' }}</div>
              <div class="config-item-help">记忆库 Markdown 根目录；未配置时记忆库不会初始化。</div>
            </el-descriptions-item>
            <el-descriptions-item label="memoryDbIsGitRepo">
              <div class="config-value">{{ boolText(form.memory_db_is_git_repo) }}</div>
              <div class="config-item-help">开启后记忆库启动前会 git pull，自动同步时会 push。</div>
              <div v-if="showMemoryDbSyncButton" class="config-item-actions">
                <GitActionButton
                  compact
                  variant="info"
                  :loading="syncLoading.memory"
                  @click="syncDatabase(RUNTIME_DATABASE_SYNC_TARGET_MEMORY)"
                >
                  同步
                </GitActionButton>
              </div>
            </el-descriptions-item>
            <el-descriptions-item label="memoryDbAutoPushDelayMinutes">
              <div class="config-value">{{ form.memory_db_auto_push_delay_minutes }}</div>
              <div class="config-item-help">知识片段写入本地 Markdown 后，延迟多少分钟自动 git commit + push；0 表示关闭自动 push。</div>
            </el-descriptions-item>
          </el-descriptions>

          <el-divider content-position="left">[path] 路径配置</el-divider>
          <el-descriptions class="memory-config-display" :column="1" border>
            <el-descriptions-item label="webkit_driver_path">
              <div class="config-value">{{ form.webkit_driver_path || '未配置，请在配置文件中设置' }}</div>
              <div class="config-item-help">Playwright WebKit 驱动目录，支持使用 {DRIVE} 占位符。</div>
            </el-descriptions-item>
            <el-descriptions-item label="webkit_data_path">
              <div class="config-value">{{ form.webkit_data_path || '未配置，请在配置文件中设置' }}</div>
              <div class="config-item-help">浏览器运行数据目录，支持使用 {DRIVE} 占位符。</div>
            </el-descriptions-item>
            <el-descriptions-item label="webkit_download_path">
              <div class="config-value">{{ form.webkit_download_path || '未配置，请在配置文件中设置' }}</div>
              <div class="config-item-help">Playwright 文件下载目录，支持使用 {DRIVE} 占位符。</div>
            </el-descriptions-item>
          </el-descriptions>
        </template>

        <el-alert
          :closable="false"
          type="info"
          title="保存后会写回当前 ini 文件并重新读取配置；如果修改了数据库路径或 git 仓库开关，建议重启应用让数据库连接和启动流程完全生效。"
        />
      </template>

      <el-form v-else label-width="120px" class="memory-config-form">
        <el-divider content-position="left">AI 整理</el-divider>
        <el-form-item label="整理模型">
          <el-select
            v-model="form.memory_arrange_model_id"
            clearable
            filterable
            style="width: 100%;"
            placeholder="请选择用于整理知识片段的 LLM 模型"
          >
            <el-option
              v-for="item in aiModelList"
              :key="item.id"
              :label="buildModelLabel(item)"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="整理提示词">
          <el-input
            v-model="form.memory_arrange_prompt"
            type="textarea"
            :rows="4"
            placeholder="请输入 AI 整理提示词"
          />
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script>
import set from '@/utils/base/git_set'
import AiSetApi from '@/utils/base/ai_set'
import GitActionButton from '@/components/base/GitActionButton.vue'

// DEFAULT_MEMORY_ARRANGE_PROMPT 定义记忆整理默认提示词。 // Default prompt used when arranging memory fragments with AI.
const DEFAULT_MEMORY_ARRANGE_PROMPT = '帮我把当前 markdown 进行整理格式，让它看起来更顺畅清晰，注意禁止修改内容'
// DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT 作为透传字段保留工作日报配置。 // Keep daily report fields in state so saving memory settings does not overwrite them.
const DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT = '请基于当前活跃任务生成中文工作日报，按已完成、进行中、风险与阻塞三个部分总结，输出 Markdown，禁止编造未提供的信息。'
// RUNTIME_DATABASE_SYNC_TARGET_* 约束手动同步接口的库类型。 // Enumerates the supported manual database sync targets.
const RUNTIME_DATABASE_SYNC_TARGET_MAIN = 'main'
const RUNTIME_DATABASE_SYNC_TARGET_MEMORY = 'memory'

// createRuntimeEditForm 创建可编辑运行时配置表单默认值。 // Build the default state for editable runtime config fields.
function createRuntimeEditForm() {
  return {
    db_path: '',
    db_file_name: '',
    db_is_git_repo: false,
    log_db_path: '',
    memory_db_path: '',
    memory_db_is_git_repo: false,
    memory_db_auto_push_delay_minutes: 1,
    webkit_driver_path: '',
    webkit_data_path: '',
    webkit_download_path: '',
  }
}

// createSyncLoading 创建主库和记忆库的独立同步 loading 状态。 // Build independent loading flags for main and memory manual sync actions.
function createSyncLoading() {
  return {
    main: false,
    memory: false,
  }
}

export default {
  name: 'MemorySet',
  components: {
    GitActionButton,
  },
  props: {
    showRuntimeConfig: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['changed'],
  data() {
    return {
      aiModelList: [],
      runtimeEditMode: false,
      runtimeEditForm: createRuntimeEditForm(),
      syncLoading: createSyncLoading(),
      form: {
        db_dir: '',
        db_name: '',
        db_is_git_repo: false,
        db_configured: false,
        log_db_path: '',
        webkit_driver_path: '',
        webkit_data_path: '',
        webkit_download_path: '',
        memory_dir: '',
        memory_db_is_git_repo: false,
        memory_db_auto_push_delay_minutes: 1,
        memory_db_configured: false,
        memory_config_file: '',
        memory_arrange_model_id: null,
        memory_arrange_prompt: DEFAULT_MEMORY_ARRANGE_PROMPT,
        home_task_daily_report_model_id: null,
        home_task_daily_report_prompt: DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT,
      },
      RUNTIME_DATABASE_SYNC_TARGET_MAIN,
      RUNTIME_DATABASE_SYNC_TARGET_MEMORY,
    }
  },
  computed: {
    // pageTitle 根据当前展示模式返回页面标题。 // Returns the page title for the current display mode.
    pageTitle() {
      return this.showRuntimeConfig ? '配置文件' : '知识片段 AI 设置'
    },
    // pageDesc 根据当前展示模式返回页面说明。 // Returns the page description for the current display mode.
    pageDesc() {
      if (this.showRuntimeConfig) {
        return '这里可以查看并编辑当前生效的 [base] 与 [path] 配置。'
      }
      return '这里维护知识片段整理相关的 AI 参数。'
    },
    // runtimeConfigAlertType 根据配置完整度返回提示类型。 // Return the alert style based on whether runtime db config is complete.
    runtimeConfigAlertType() {
      return this.form.db_configured && this.form.memory_db_configured ? 'info' : 'warning'
    },
    // memoryConfigAlertTitle 统一生成主库和记忆库配置提示。 // Build a consistent hint message for the current database configuration.
    memoryConfigAlertTitle() {
      const configFile = this.form.memory_config_file || '配置文件'
      if (!this.form.db_configured) {
        return `未检测到主库配置，请在 ${configFile} 的 [base] 节点中配置 dbPath 和 dbFileName。`
      }
      if (!this.form.memory_db_configured) {
        return `未检测到记忆库配置，请在 ${configFile} 的 [base] 节点中配置 memoryDbPath。`
      }
      return `当前主库、记忆库和路径配置均来自 ${configFile} 的 [base] 与 [path] 节点。`
    },
    // showMainDbSyncButton 仅在运行时配置展示态且主库启用 Git 时展示同步按钮。 // Show the main-db sync button only in display mode when Git sync is enabled.
    showMainDbSyncButton() {
      return this.showRuntimeConfig && !this.runtimeEditMode && this.form.db_is_git_repo
    },
    // showMemoryDbSyncButton 仅在运行时配置展示态且记忆库启用 Git 时展示同步按钮。 // Show the memory-db sync button only in display mode when Git sync is enabled.
    showMemoryDbSyncButton() {
      return this.showRuntimeConfig && !this.runtimeEditMode && this.form.memory_db_is_git_repo
    },
  },
  mounted() {
    this.loadAiModelList()
    this.loadConfig()
  },
  methods: {
    // buildModelLabel 生成模型下拉展示文案，统一显示服务商和模型名。 // Build a readable select label with provider and model information.
    buildModelLabel(item) {
      const provider = item.provider_name || '未命名服务商'
      const model = item.name || item.model || `模型#${item.id}`
      return `${provider} / ${model}`
    },
    // boolText 把布尔值转换为字符串，方便配置展示。 // Convert boolean config values to readable true or false text.
    boolText(value) {
      return value ? 'true' : 'false'
    },
    loadAiModelList() {
      if (this.showRuntimeConfig) {
        return
      }
      AiSetApi.AiModelList({ model_type: 'llm' }, (response) => {
        if (response.ErrCode !== 0) {
          return
        }
        this.aiModelList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    // syncRuntimeEditForm 把当前展示数据同步到编辑表单。 // Sync the current display values into the runtime edit form.
    syncRuntimeEditForm() {
      this.runtimeEditForm = {
        db_path: this.form.db_dir || '',
        db_file_name: this.form.db_name || '',
        db_is_git_repo: !!this.form.db_is_git_repo,
        log_db_path: this.form.log_db_path || '',
        memory_db_path: this.form.memory_dir || '',
        memory_db_is_git_repo: !!this.form.memory_db_is_git_repo,
        memory_db_auto_push_delay_minutes: Number(this.form.memory_db_auto_push_delay_minutes ?? 1),
        webkit_driver_path: this.form.webkit_driver_path || '',
        webkit_data_path: this.form.webkit_data_path || '',
        webkit_download_path: this.form.webkit_download_path || '',
      }
    },
    loadConfig() {
      set.MemoryConfigGet((response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.form.db_dir = response.Data.db_dir || ''
        this.form.db_name = response.Data.db_name || ''
        this.form.db_is_git_repo = !!response.Data.db_is_git_repo
        this.form.db_configured = !!response.Data.db_configured
        this.form.log_db_path = response.Data.log_db_path || ''
        this.form.webkit_driver_path = response.Data.webkit_driver_path || ''
        this.form.webkit_data_path = response.Data.webkit_data_path || ''
        this.form.webkit_download_path = response.Data.webkit_download_path || ''
        this.form.memory_dir = response.Data.memory_dir || ''
        this.form.memory_db_is_git_repo = !!response.Data.memory_db_is_git_repo
        this.form.memory_db_auto_push_delay_minutes = Number(response.Data.memory_db_auto_push_delay_minutes ?? 1)
        this.form.memory_db_configured = !!response.Data.memory_db_configured
        this.form.memory_config_file = response.Data.memory_config_file || ''
        this.form.memory_arrange_model_id = response.Data.memory_arrange_model_id || null
        this.form.memory_arrange_prompt = response.Data.memory_arrange_prompt || DEFAULT_MEMORY_ARRANGE_PROMPT
        this.form.home_task_daily_report_model_id = response.Data.home_task_daily_report_model_id || null
        this.form.home_task_daily_report_prompt = response.Data.home_task_daily_report_prompt || DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT
        this.syncRuntimeEditForm()
      })
    },
    // startRuntimeEdit 进入运行时配置编辑模式。 // Enter runtime config edit mode.
    startRuntimeEdit() {
      this.syncRuntimeEditForm()
      this.runtimeEditMode = true
    },
    // cancelRuntimeEdit 退出运行时配置编辑模式并恢复当前展示值。 // Exit runtime config edit mode and restore the current values.
    cancelRuntimeEdit() {
      this.runtimeEditMode = false
      this.syncRuntimeEditForm()
    },
    // saveRuntimeConfig 保存 ini 中可编辑的运行时配置，并重新加载页面数据。 // Save editable ini config values and reload current config state.
    saveRuntimeConfig() {
      const payload = {
        db_path: this.runtimeEditForm.db_path,
        db_file_name: this.runtimeEditForm.db_file_name,
        db_is_git_repo: this.runtimeEditForm.db_is_git_repo,
        log_db_path: this.runtimeEditForm.log_db_path,
        memory_db_path: this.runtimeEditForm.memory_db_path,
        memory_db_is_git_repo: this.runtimeEditForm.memory_db_is_git_repo,
        memory_db_auto_push_delay_minutes: Number(this.runtimeEditForm.memory_db_auto_push_delay_minutes || 0),
        webkit_driver_path: this.runtimeEditForm.webkit_driver_path,
        webkit_data_path: this.runtimeEditForm.webkit_data_path,
        webkit_download_path: this.runtimeEditForm.webkit_download_path,
      }
      set.RuntimeConfigSave(payload, (response) => {
        if (response.ErrCode !== 0) {
          this.$helperNotify.error(response.ErrMsg || '配置保存失败')
          return
        }
        this.runtimeEditMode = false
        this.loadConfig()
        this.$helperNotify.success('配置已写入文件并重新加载')
        this.$emit('changed')
      })
    },
    // syncDatabase 手动触发当前库的 git commit push，并保留后端错误原文。 // Trigger git commit and push for the selected database and preserve backend error details.
    syncDatabase(target) {
      const loadingKey = target === RUNTIME_DATABASE_SYNC_TARGET_MAIN ? 'main' : 'memory'
      const successText = target === RUNTIME_DATABASE_SYNC_TARGET_MAIN ? '主库同步完成' : '记忆库同步完成'
      const idleText = target === RUNTIME_DATABASE_SYNC_TARGET_MAIN ? '主库未检测到变更，无需同步' : '记忆库未检测到变更，无需同步'
      this.syncLoading[loadingKey] = true
      set.RuntimeDatabaseGitSync({ target }, (response) => {
        this.syncLoading[loadingKey] = false
        if (response.ErrCode !== 0) {
          this.$helperNotify.error(response.ErrMsg || '同步失败')
          return
        }
        if (response.Data && response.Data.changed) {
          this.$helperNotify.success(successText)
          return
        }
        this.$helperNotify.info(idleText)
      })
    },
    // saveAiConfig 保存知识片段 AI 配置，并透传日报配置防止被覆盖。 // Save memory AI config and pass through daily-report config to avoid overwriting it.
    saveAiConfig() {
      const payload = {
        memory_arrange_model_id: this.form.memory_arrange_model_id,
        memory_arrange_prompt: this.form.memory_arrange_prompt,
        home_task_daily_report_model_id: this.form.home_task_daily_report_model_id,
        home_task_daily_report_prompt: this.form.home_task_daily_report_prompt,
      }
      set.MemoryConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('AI 配置已保存')
          this.$emit('changed')
          return
        }
        this.$helperNotify.error(response.ErrMsg || 'AI 配置保存失败')
      })
    },
  },
}
</script>

<style scoped>
@import "@/css/set_module_unified.css";

.memory-config-display {
  margin: 16px 0;
}

.memory-config-form {
  margin-top: 16px;
}

.config-value {
  color: #24312f;
  line-height: 1.6;
  word-break: break-all;
}

.config-item-help {
  margin-top: 6px;
  color: #6b7280;
  font-size: 12px;
  line-height: 1.6;
}

.config-item-actions {
  margin-top: 10px;
}
</style>
