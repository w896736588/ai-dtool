<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">{{ pageTitle }}</h3>
      <p class="set-config-desc">{{ pageDesc }}</p>
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

        <!-- [smart_link] 自定义网页配置 -->
        <el-divider content-position="left">[smart_link] 自定义网页配置</el-divider>
        <el-descriptions class="memory-config-display" :column="1" border>
          <el-descriptions-item label="run_mode">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'run_mode'">
                <div class="config-edit-row">
                  <el-select v-model="editingItem.value" style="width: 200px">
                    <el-option label="server (服务端执行)" value="server" />
                    <el-option label="local_client (本地客户端执行)" value="local_client" />
                  </el-select>
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" @click="saveItem('smart_link', 'run_mode', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">
                    <el-tag :type="form.run_mode === 'local_client' ? 'success' : 'info'" size="small">
                      {{ form.run_mode || 'server' }}
                    </el-tag>
                    <span class="config-value-desc">
                      {{ form.run_mode === 'local_client' ? '本地客户端执行模式' : '服务端执行模式' }}
                    </span>
                  </div>
                  <GitActionButton compact size="small" @click="startEdit('run_mode', form.run_mode || 'server')">编辑</GitActionButton>
                </div>
              </template>
              <div class="config-item-help">自定义网页运行模式：server(服务端执行) 或 local_client(本地客户端执行)</div>
            </div>
          </el-descriptions-item>
        </el-descriptions>

        <!-- [base] 主库配置 -->
        <el-divider content-position="left">[base] 主库配置</el-divider>
        <el-descriptions class="memory-config-display" :column="1" border>
          <el-descriptions-item label="dbPath">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'db_path'">
                <div class="config-edit-row">
                  <el-input v-model="editingItem.value" placeholder="请输入主库目录" style="flex: 1" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('base', 'db_path', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ form.db_dir || '未配置，请在配置文件中设置' }}</div>
                  <GitActionButton compact size="small" @click="startEdit('db_path', form.db_dir)">编辑</GitActionButton>
                </div>
              </template>
              <div class="config-item-help">主库 sqlite 所在目录；未配置时默认使用 config/{AppName}。</div>
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="dbFileName">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'db_file_name'">
                <div class="config-edit-row">
                  <el-input v-model="editingItem.value" placeholder="请输入主库文件名" style="flex: 1" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('base', 'dbFileName', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ form.db_name || '未配置，请在配置文件中设置' }}</div>
                  <GitActionButton compact size="small" @click="startEdit('db_file_name', form.db_name)">编辑</GitActionButton>
                </div>
              </template>
              <div class="config-item-help">主库 sqlite 文件名；未配置时默认使用 {AppName}.db。</div>
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="dbIsGitRepo">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'db_is_git_repo'">
                <div class="config-edit-row">
                  <el-switch v-model="editingItem.value" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('base', 'db_is_git_repo', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ boolText(form.db_is_git_repo) }}</div>
                  <div class="config-item-actions">
                    <GitActionButton compact size="small" @click="startEdit('db_is_git_repo', form.db_is_git_repo)">编辑</GitActionButton>
                    <GitActionButton
                      v-if="form.db_is_git_repo"
                      compact
                      size="small"
                      variant="info"
                      :loading="syncLoading.main"
                      @click="syncDatabase(RUNTIME_DATABASE_SYNC_TARGET_MAIN)"
                    >
                      同步
                    </GitActionButton>
                  </div>
                </div>
              </template>
              <div class="config-item-help">开启后主库在使用前会 git pull，关闭程序时会自动 push。</div>
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="dbAutoPushDelayMinutes">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'db_auto_push_delay_minutes'">
                <div class="config-edit-row">
                  <el-input-number v-model="editingItem.value" :min="0" :step="1" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('base', 'dbAutoPushDelayMinutes', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ form.db_auto_push_delay_minutes }} 分钟</div>
                  <GitActionButton compact size="small" @click="startEdit('db_auto_push_delay_minutes', form.db_auto_push_delay_minutes)">编辑</GitActionButton>
                </div>
              </template>
              <div class="config-item-help">主库数据变更后，延迟多少分钟自动 git commit + push；0 表示关闭自动 push，默认 10 分钟。</div>
            </div>
          </el-descriptions-item>
        </el-descriptions>

        <!-- [base] 日志库配置 -->
        <el-divider content-position="left">[base] 日志库配置</el-divider>
        <el-descriptions class="memory-config-display" :column="1" border>
          <el-descriptions-item label="logDbPath">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'log_db_path'">
                <div class="config-edit-row">
                  <el-input v-model="editingItem.value" placeholder="请输入日志库目录" style="flex: 1" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('base', 'logDbPath', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ form.log_db_path || '未配置（默认与主库相同）' }}</div>
                  <GitActionButton compact size="small" @click="startEdit('log_db_path', form.log_db_path)">编辑</GitActionButton>
                </div>
              </template>
              <div class="config-item-help">日志库 sqlite 所在目录；未配置时默认使用与主库相同目录。</div>
            </div>
          </el-descriptions-item>
        </el-descriptions>

        <!-- [base] 记忆库配置 -->
        <el-divider content-position="left">[base] 记忆库配置</el-divider>
        <el-descriptions class="memory-config-display" :column="1" border>
          <el-descriptions-item label="memoryDbPath">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'memory_db_path'">
                <div class="config-edit-row">
                  <el-input v-model="editingItem.value" placeholder="请输入记忆库目录" style="flex: 1" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('base', 'memoryDbPath', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ form.memory_dir || '未配置，请在配置文件中设置' }}</div>
                  <GitActionButton compact size="small" @click="startEdit('memory_db_path', form.memory_dir)">编辑</GitActionButton>
                </div>
              </template>
              <div class="config-item-help">记忆库 Markdown 根目录；未配置时记忆库不会初始化。</div>
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="memoryDbIsGitRepo">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'memory_db_is_git_repo'">
                <div class="config-edit-row">
                  <el-switch v-model="editingItem.value" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('base', 'memoryDbIsGitRepo', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ boolText(form.memory_db_is_git_repo) }}</div>
                  <div class="config-item-actions">
                    <GitActionButton compact size="small" @click="startEdit('memory_db_is_git_repo', form.memory_db_is_git_repo)">编辑</GitActionButton>
                    <GitActionButton
                      v-if="form.memory_db_is_git_repo"
                      compact
                      size="small"
                      variant="info"
                      :loading="syncLoading.memory"
                      @click="syncDatabase(RUNTIME_DATABASE_SYNC_TARGET_MEMORY)"
                    >
                      同步
                    </GitActionButton>
                  </div>
                </div>
              </template>
              <div class="config-item-help">开启后记忆库启动前会 git pull，自动同步时会 push。</div>
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="memoryDbAutoPushDelayMinutes">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'memory_db_auto_push_delay_minutes'">
                <div class="config-edit-row">
                  <el-input-number v-model="editingItem.value" :min="0" :step="1" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('base', 'memoryDbAutoPushDelayMinutes', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ form.memory_db_auto_push_delay_minutes }} 分钟</div>
                  <GitActionButton compact size="small" @click="startEdit('memory_db_auto_push_delay_minutes', form.memory_db_auto_push_delay_minutes)">编辑</GitActionButton>
                </div>
              </template>
              <div class="config-item-help">知识片段写入本地 Markdown 后，延迟多少分钟自动 git commit + push；0 表示关闭自动 push。</div>
            </div>
          </el-descriptions-item>
        </el-descriptions>

        <!-- [safe] 安全登录配置 -->
        <el-divider content-position="left">[safe] 安全登录配置</el-divider>
        <el-descriptions class="memory-config-display" :column="1" border>
          <el-descriptions-item label="password">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'safe_password'">
                <div class="config-edit-row">
                  <el-input v-model="editingItem.value" placeholder="请输入后台访问密码" show-password style="flex: 1" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('safe', 'password', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ safePasswordDisplay }}</div>
                  <GitActionButton compact size="small" @click="startEdit('safe_password', form.safe_password)">编辑</GitActionButton>
                </div>
              </template>
              <div class="config-item-help">后台访问密码，留空表示不启用密码保护。</div>
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="sessionExpireMinutes">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'safe_session_expire_minutes'">
                <div class="config-edit-row">
                  <el-input-number v-model="editingItem.value" :min="0" :step="10" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('safe', 'sessionExpireMinutes', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ form.safe_session_expire_minutes }} 分钟</div>
                  <GitActionButton compact size="small" @click="startEdit('safe_session_expire_minutes', form.safe_session_expire_minutes)">编辑</GitActionButton>
                </div>
              </template>
              <div class="config-item-help">会话有效期（分钟），默认 120 分钟。设为 0 表示永不过期。每次请求成功会自动续期。</div>
            </div>
          </el-descriptions-item>
        </el-descriptions>
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
        <el-form-item>
          <GitActionButton type="primary" @click="saveAiConfig">保存 AI 配置</GitActionButton>
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

// createSyncLoading 创建主库和记忆库的独立同步 loading 状态。 // Build independent loading flags for main and memory manual sync actions.
function createSyncLoading() {
  return {
    main: false,
    memory: false,
  }
}

// createEditingItem 创建编辑中的配置项状态。 // Build the editing state for a single config item.
function createEditingItem() {
  return {
    key: '',
    value: null,
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
      syncLoading: createSyncLoading(),
      saving: false,
      editingItem: createEditingItem(),
      form: {
        db_dir: '',
        db_name: '',
        db_is_git_repo: false,
        db_auto_push_delay_minutes: 10,
        db_configured: false,
        log_db_path: '',
        memory_dir: '',
        memory_db_is_git_repo: false,
        memory_db_auto_push_delay_minutes: 1,
        memory_db_configured: false,
        memory_config_file: '',
        memory_arrange_model_id: null,
        memory_arrange_prompt: DEFAULT_MEMORY_ARRANGE_PROMPT,
        home_task_daily_report_model_id: null,
        home_task_daily_report_prompt: DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT,
        safe_password: '',
        safe_session_expire_minutes: 120,
        run_mode: 'server',
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
        return '这里可以查看并编辑当前生效的 [base]、[path]、[safe] 和 [smart_link] 配置，每个配置项可独立编辑保存。'
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
    // safePasswordDisplay 展示密码状态（已设置/未设置）
    safePasswordDisplay() {
      if (this.form.safe_password && this.form.safe_password.length > 0) {
        return '已设置（' + '*'.repeat(Math.min(this.form.safe_password.length, 8)) + '）'
      }
      return '未设置（不启用密码保护）'
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
        if (response.__loginRequired) {
          return
        }
        if (response.ErrCode !== 0) {
          return
        }
        this.aiModelList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    loadConfig() {
      set.MemoryConfigGet((response) => {
        if (response.__loginRequired) {
          return
        }
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.form.db_dir = response.Data.db_dir || ''
        this.form.db_name = response.Data.db_name || ''
        this.form.db_is_git_repo = !!response.Data.db_is_git_repo
        this.form.db_auto_push_delay_minutes = Number(response.Data.db_auto_push_delay_minutes ?? 10)
        this.form.db_configured = !!response.Data.db_configured
        this.form.log_db_path = response.Data.log_db_path || ''
        this.form.memory_dir = response.Data.memory_dir || ''
        this.form.memory_db_is_git_repo = !!response.Data.memory_db_is_git_repo
        this.form.memory_db_auto_push_delay_minutes = Number(response.Data.memory_db_auto_push_delay_minutes ?? 1)
        this.form.memory_db_configured = !!response.Data.memory_db_configured
        this.form.memory_config_file = response.Data.memory_config_file || ''
        this.form.memory_arrange_model_id = response.Data.memory_arrange_model_id || null
        this.form.memory_arrange_prompt = response.Data.memory_arrange_prompt || DEFAULT_MEMORY_ARRANGE_PROMPT
        this.form.home_task_daily_report_model_id = response.Data.home_task_daily_report_model_id || null
        this.form.home_task_daily_report_prompt = response.Data.home_task_daily_report_prompt || DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT
        this.form.safe_password = response.Data.safe_password || ''
        this.form.safe_session_expire_minutes = Number(response.Data.safe_session_expire_minutes ?? 120)
        this.form.run_mode = response.Data.run_mode || 'server'
      })
    },
    // startEdit 开始编辑单个配置项。 // Start editing a single config item.
    startEdit(key, value) {
      this.editingItem = {
        key,
        value: value === null || value === undefined ? '' : value,
      }
    },
    // cancelEdit 取消编辑。 // Cancel editing.
    cancelEdit() {
      this.editingItem = createEditingItem()
    },
    // saveItem 保存单个配置项。 // Save a single config item.
    saveItem(section, key, value) {
      this.saving = true
      const payload = {
        section,
        key,
        value,
      }
      set.RuntimeConfigItemSave(payload, (response) => {
        this.saving = false
        if (response.__loginRequired) {
          return
        }
        if (response.ErrCode !== 0) {
          this.$helperNotify.error(response.ErrMsg || '保存失败')
          return
        }
        this.$helperNotify.success('保存成功')
        this.editingItem = createEditingItem()
        this.loadConfig()
        // 如果需要重新登录
        if (response.Data && response.Data.need_relogin) {
          this.$helperNotify.success('密码已修改，请使用新密码重新登录')
          this.$base.ClearSafeToken()
          if (this.$eventBus) {
            this.$eventBus.emit('safe_auth_required', { message: '密码已修改，请使用新密码登录' })
          }
          return
        }
        // 如果需要重启提示
        if (response.Data && response.Data.need_restart) {
          this.$helperNotify.info('配置已保存，建议重启应用让数据库配置完全生效')
        }
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
        if (response.__loginRequired) {
          return
        }
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
        if (response.__loginRequired) {
          return
        }
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

.config-value-desc {
  margin-left: 8px;
  color: #6b7280;
  font-size: 13px;
}

.config-item-help {
  margin-top: 6px;
  color: #6b7280;
  font-size: 12px;
  line-height: 1.6;
}

.config-item-wrapper {
  width: 100%;
}

.config-display-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.config-edit-row {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.config-edit-actions {
  display: flex;
  gap: 8px;
}

.config-item-actions {
  display: flex;
  gap: 8px;
}
</style>
