<template>
  <div class="agent-cli-page">
    <div class="agent-cli-header-card">
      <div class="agent-cli-header-title">
        <svg class="agent-cli-header-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect x="3" y="4" width="18" height="16" rx="2.5" stroke="currentColor" stroke-width="2" />
          <path d="M7 9H17" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
          <path d="M7 13H13" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
          <path d="M15.5 16.5L17 18L19.5 15.5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
        <span>Agent Cli 管理</span>
      </div>
      <div class="agent-cli-header-actions">
        <GitActionButton compact @click="openCreateDialog">新建</GitActionButton>
        <GitActionButton compact variant="info" @click="openWebhookDialog">Webhook 配置</GitActionButton>
        <GitActionButton compact variant="warning" @click="chromeDevtoolsDialogVisible = true">ChromeDevTools</GitActionButton>
      </div>
    </div>

    <div v-loading="loading" class="agent-cli-content-card">
      <div class="agent-cli-list">
        <div v-if="list.length === 0" class="agent-cli-empty">
          暂无 Agent Cli 实例，点击“新建”创建
        </div>
        <div
          v-for="row in list"
          :key="row.id"
          class="agent-cli-card"
          :class="{ 'agent-cli-card--inactive': !row.displayed_enabled }"
        >
          <div class="agent-cli-card__header">
            <div class="agent-cli-card__main">
              <div class="agent-cli-card__title-row">
                <div class="agent-cli-card__title">{{ row.name || '-' }}</div>
                <el-tag size="small" type="info">{{ formatTypeLabel(row.type) }}</el-tag>
                <span class="agent-cli-card__status-dot" :class="row.displayed_enabled ? 'agent-cli-card__status-dot--active' : 'agent-cli-card__status-dot--inactive'"></span>
                <span class="agent-cli-card__status-text">{{ row.displayed_enabled ? '已启用' : '已停止' }}</span>
              </div>
              <div class="agent-cli-card__meta">
                <span>ID：{{ row.id }}</span>
                <span v-if="row.type !== 'codex-cli'">配置文件：{{ row.settings_exists ? '存在' : '不存在' }}</span>
                <span>可选模型：{{ formatModelOptions(row.model_options) }}</span>
                <span v-if="row.type !== 'codex-cli'">McpServers：{{ row.mcp_server_count || 0 }} 个</span>
              </div>
              <div class="agent-cli-card__summary-grid">
                <div class="agent-cli-info-block">
                  <div class="agent-cli-info-block__label">启停状态</div>
                  <div class="agent-cli-switch-line">
                    <el-switch
                      :model-value="row.displayed_enabled"
                      size="small"
                      :loading="row._togglingEnabled"
                      @change="toggleEnabled(row, $event)"
                    />
                    <span class="agent-cli-switch-line__text">{{ row.displayed_enabled ? '运行中' : '已停止' }}</span>
                  </div>
                </div>

                <div class="agent-cli-info-block">
                  <div class="agent-cli-info-block__label">通知配置</div>
                  <el-select
                    v-model="row.webhook_config_id"
                    size="small"
                    placeholder="未配置"
                    clearable
                    class="agent-cli-webhook-select"
                    @change="updateWebhookConfig(row)"
                  >
                    <el-option
                      v-for="wh in webhookOptions"
                      :key="wh.id"
                      :label="wh.name"
                      :value="String(wh.id)"
                    />
                  </el-select>
                </div>

                <div v-if="row.type !== 'codex-cli'" class="agent-cli-info-block">
                  <div class="agent-cli-info-block__label">claude-mem</div>
                  <div class="agent-cli-switch-line">
                    <el-switch
                      v-model="row.claude_mem_enabled"
                      size="small"
                      :loading="row._togglingMem"
                      @change="toggleClaudeMem(row)"
                    />
                    <span class="agent-cli-switch-line__text">{{ row.claude_mem_enabled ? '已启用' : '已禁用' }}</span>
                  </div>
                </div>
              </div>

              <div class="agent-cli-config-table-wrap">
                <table class="agent-cli-config-table">
                  <tbody>
                    <tr>
                      <th>类型</th>
                      <td>{{ formatTypeLabel(row.type) }}</td>
                      <th>模型列表</th>
                      <td colspan="3" class="agent-cli-config-table__value agent-cli-config-table__value--break">{{ formatModelOptions(row.model_options) }}</td>
                    </tr>
                    <tr>
                      <th>请求地址</th>
                      <td class="agent-cli-config-table__value agent-cli-config-table__value--break">{{ row.request_url || '-' }}</td>
                      <th>Webhook</th>
                      <td>{{ row.webhook_config_name || '-' }}</td>
                    </tr>
                    <tr v-if="row.type !== 'codex-cli'">
                      <th>路径</th>
                      <td class="agent-cli-config-table__value agent-cli-config-table__value--break">{{ row.settings_path || '-' }}</td>
                      <th>配置文件</th>
                      <td>
                        <el-tag :type="row.settings_exists ? 'success' : 'danger'" size="small">
                          {{ row.settings_exists ? '存在' : '不存在' }}
                        </el-tag>
                      </td>
                    </tr>
                    <tr v-if="row.type !== 'codex-cli'">
                      <th>McpServers</th>
                      <td>{{ row.mcp_server_count || 0 }} 个</td>
                      <th>claude-mem</th>
                      <td>{{ row.claude_mem_enabled ? '已启用' : '已禁用' }}</td>
                    </tr>
                  </tbody>
                </table>
              </div>
            </div>

            <div class="agent-cli-card__actions">
              <GitActionButton
                v-if="row.type !== 'codex-cli'"
                compact
                variant="primary"
                @click="configureMcp(row)"
              >
                配置DevtoolsMcp
              </GitActionButton>
              <GitActionButton compact variant="info" @click="editItem(row)">编辑</GitActionButton>
              <GitActionButton compact variant="danger" @click="deleteItem(row)">删除</GitActionButton>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ChromeDevTools 弹窗 -->
    <el-dialog
      v-model="chromeDevtoolsDialogVisible"
      title="Chrome DevTools 管理"
      width="1000px"
      top="5vh"
      :destroy-on-close="true"
    >
      <iframe
        src="/#/Mcp/chrome-devtools?hide_menu=1&embed=1"
        style="width: 100%; height: 78vh; border: none;"
      />
    </el-dialog>

    <!-- 新建/编辑对话框 -->
    <!-- 新建/编辑弹窗：加宽并允许点击蒙层关闭 / Wider dialog and close on backdrop click. -->
    <el-dialog v-model="dialogVisible" :title="editingId > 0 ? '编辑' : '新建 Agent Cli'" width="720px">
      <el-form :model="form" label-width="140px">
        <el-form-item label="名称">
          <el-input v-model="form.name" :placeholder="form.type === 'codex-cli' ? '默认 Codex CLI' : '默认 Claude Code CLI'" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" style="width: 100%" @change="onTypeChange">
            <el-option label="Claude Code CLI" value="claude-code-cli" />
            <el-option label="Codex CLI" value="codex-cli" />
          </el-select>
        </el-form-item>
        <!-- Claude Code CLI 配置 -->
        <template v-if="form.type !== 'codex-cli'">
          <el-form-item label="settings.json 路径">
            <el-input v-model="form.settings_path" placeholder="请输入 settings.json 的绝对路径" />
            <div class="agent-cli-form-tip">例如: C:\Users\xxx\.claude\settings.json</div>
          </el-form-item>
          <el-form-item label="模型名">
            <el-input v-model="form.model_name" placeholder="默认模型，例如 deepseek-v4-pro[1m]" />
          </el-form-item>
          <el-form-item label="模型列表">
            <el-input
              v-model="form.model_list_text"
              type="textarea"
              :rows="4"
              placeholder="每行一个模型；留空则仅使用上方默认模型"
            />
            <div class="agent-cli-form-tip">首个模型会作为默认模型；执行任务时可再选择具体模型。</div>
          </el-form-item>
          <el-form-item label="API Key">
            <el-input v-model="form.api_key" type="password" show-password placeholder="请输入 DeepSeek API Key" />
          </el-form-item>
          <el-form-item label="Base URL">
            <el-input v-model="form.base_url" placeholder="https://api.deepseek.com/anthropic" />
          </el-form-item>
        </template>
        <!-- Codex CLI 配置 -->
        <template v-else>
          <el-form-item label="API Key" required>
            <el-input v-model="form.codex_api_key" type="password" show-password placeholder="请输入 OpenAI API Key" />
          </el-form-item>
          <el-form-item label="模型列表">
            <el-input
              v-model="form.codex_model_list_text"
              type="textarea"
              :rows="4"
              placeholder="每行一个模型；留空则仅使用上方默认模型"
            />
            <div class="agent-cli-form-tip">首个模型会作为默认模型；执行任务时可再选择具体模型。</div>
          </el-form-item>
          <el-form-item label="Base URL">
            <el-input v-model="form.codex_base_url" placeholder="自定义 API 端点（可选）" />
          </el-form-item>
          <el-form-item label="Sandbox Mode">
            <el-input v-model="form.codex_sandbox_mode" placeholder="danger-full-access" />
            <div class="agent-cli-form-tip">默认: danger-full-access</div>
          </el-form-item>
        </template>
        <el-form-item label="Webhook 通知">
          <el-select v-model="form.webhook_config_id" placeholder="不通知" clearable style="width: 100%">
            <el-option
              v-for="wh in webhookOptions"
              :key="wh.id"
              :label="wh.name"
              :value="String(wh.id)"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveItem">保存</el-button>
      </template>
    </el-dialog>

    <!-- Webhook 配置管理弹窗 -->
    <el-dialog v-model="webhookDialogVisible" title="Webhook 通知配置" width="640px">
      <div style="margin-bottom: 12px; text-align: right;">
        <el-button type="primary" size="small" @click="openWebhookForm(null)">新增</el-button>
      </div>
      <el-table :data="webhookList" v-loading="webhookLoading" size="small" border>
        <el-table-column prop="name" label="名称" min-width="100" />
        <el-table-column prop="type" label="类型" width="90">
          <template #default="{ row }">
            <el-tag size="small">{{ webhookTypeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="webhook_url" label="Webhook 地址" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button size="small" link type="primary" @click="openWebhookForm(row)">编辑</el-button>
            <el-button size="small" link type="danger" @click="deleteWebhook(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 内嵌新增/编辑表单 -->
      <div v-if="webhookFormVisible" class="webhook-form-section">
        <div class="webhook-form-section__title">{{ webhookForm.id > 0 ? '编辑配置' : '新增配置' }}</div>
        <el-form :model="webhookForm" label-width="100px" size="small">
          <el-form-item label="名称">
            <el-input v-model="webhookForm.name" placeholder="如: 前端组钉钉群" />
          </el-form-item>
          <el-form-item label="类型">
            <el-select v-model="webhookForm.type" style="width: 100%">
              <el-option label="钉钉" value="dingtalk" />
              <el-option label="飞书" value="feishu" />
              <el-option label="企业微信" value="wecom" />
            </el-select>
          </el-form-item>
          <el-form-item label="Webhook 地址">
            <el-input v-model="webhookForm.webhook_url" placeholder="https://oapi.dingtalk.com/robot/send?access_token=xxx" />
          </el-form-item>
          <el-form-item label="签名密钥">
            <el-input v-model="webhookForm.secret" placeholder="SEC... (可选)" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="webhookSaving" @click="saveWebhook">保存</el-button>
            <el-button @click="webhookFormVisible = false">取消</el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-dialog>

  </div>
</template>

<script>
import agentCliApi from '../../utils/base/agent_cli'
import GitActionButton from '@/components/base/GitActionButton.vue'

// AGENT_CLI_ENABLED_SORT_TRUE 启用状态排序值，启用项排在前面。 // Sort weight for enabled rows so active items stay at the top.
const AGENT_CLI_ENABLED_SORT_TRUE = 1
// AGENT_CLI_ENABLED_SORT_FALSE 禁用状态排序值，禁用项排在后面。 // Sort weight for disabled rows so inactive items move below active ones.
const AGENT_CLI_ENABLED_SORT_FALSE = 0

export default {
  components: {
    GitActionButton,
  },
  data() {
    return {
      loading: false,
      list: [],
      // 新建/编辑
      dialogVisible: false,
      chromeDevtoolsDialogVisible: false,
      editingId: 0,
      saving: false,
      form: {
        name: '',
        type: 'claude-code-cli',
        settings_path: '',
        webhook_config_id: '',
        enabled: 1,
        model_name: '',
        model_list_text: '',
        api_key: '',
        base_url: '',
        // Codex CLI 专属字段
        codex_api_key: '',
        codex_model_list_text: '',
        codex_base_url: '',
        codex_sandbox_mode: '',
      },
      // webhook 配置
      webhookDialogVisible: false,
      webhookLoading: false,
      webhookList: [],
      webhookOptions: [],
      webhookFormVisible: false,
      webhookSaving: false,
      webhookForm: {
        id: 0,
        name: '',
        type: 'dingtalk',
        webhook_url: '',
        secret: '',
      },
    }
  },
  mounted() {
    this.loadList()
    this.loadWebhookOptions()
  },
  methods: {
    // openWebhookDialog 打开 webhook 配置弹窗并同步刷新列表。 // openWebhookDialog opens the webhook dialog and refreshes its list before display.
    openWebhookDialog() {
      this.webhookDialogVisible = true
      this.loadWebhookList()
    },
    // formatTypeLabel 统一格式化实例类型文案，避免页面直接暴露内部值。 // formatTypeLabel normalizes instance type labels so the UI does not expose raw internal values.
    formatTypeLabel(type) {
      if (type === 'codex-cli') {
        return 'Codex CLI'
      }
      if (type === 'claude-code-cli') {
        return 'Claude Code CLI'
      }
      return type || '-'
    },
    loadList() {
      this.loading = true
      agentCliApi.AgentCliList((response) => {
        this.loading = false
        if (response && response.ErrCode === 0 && response.Data) {
          const items = response.Data.list || []
          items.forEach(item => {
            item.webhook_config_id = item.webhook_config_id ? String(item.webhook_config_id) : ''
            item.displayed_enabled = !!item.displayed_enabled
          })
          this.list = this.sortAgentCliList(items)
        }
      })
    },
    openCreateDialog() {
      this.editingId = 0
      this.form = {
        name: '',
        type: 'claude-code-cli',
        settings_path: '',
        webhook_config_id: '',
        enabled: 1,
        model_name: '',
        model_list_text: '',
        api_key: '',
        base_url: '',
        codex_api_key: '',
        codex_model_list_text: '',
        codex_base_url: '',
        codex_sandbox_mode: '',
      }
      this.dialogVisible = true
    },
    onTypeChange() {
      // 类型切换时同步默认启停状态 / Sync default enabled status when type changes.
      this.form.enabled = this.form.type === 'codex-cli' ? 0 : 1
    },
    saveItem() {
      const isCodex = this.form.type === 'codex-cli'
      const claudeModels = this.parseModelList(this.form.model_list_text, this.form.model_name)
      const codexModels = this.parseModelList(this.form.codex_model_list_text, '')
      if (isCodex) {
        // Codex 现在只保留模型列表字段，首项作为默认模型 / Codex now only uses the model list and the first item becomes default.
        if (!this.form.codex_api_key.trim()) {
          this.$message.warning('请输入 API Key')
          return
        }
        if (codexModels.length === 0) {
          this.$message.warning('请输入模型列表')
          return
        }
      } else {
        if (!this.form.settings_path.trim()) {
          this.$message.warning('请输入 settings.json 路径')
          return
        }
      }
      this.saving = true
      const data = {
        id: this.editingId,
        name: this.form.name,
        type: this.form.type,
        settings_path: isCodex ? '' : this.form.settings_path.trim(),
        enabled: this.form.enabled,
        webhook_config_id: parseInt(this.form.webhook_config_id) || 0,
      }
      // Codex 类型：将配置序列化为 config JSON
      if (isCodex) {
        data.config = JSON.stringify({
          api_key: this.form.codex_api_key.trim(),
          model: codexModels[0] || '',
          models: codexModels,
          base_url: this.form.codex_base_url.trim() || undefined,
          sandbox_mode: this.form.codex_sandbox_mode.trim() || undefined,
        })
      }
      agentCliApi.AgentCliSave(data, (response) => {
        if (response && response.ErrCode === 0) {
          // 新建时从返回值取 ID，后续 DeepSeek 写入依赖此 ID
          if (!this.editingId && response.Data && response.Data.id) {
            this.editingId = response.Data.id
          }
          // Claude 类型：密钥字段非空时，一并写入 DeepSeek 配置
          if (!isCodex && this.form.model_name.trim() && this.form.api_key.trim()) {
            const dsData = {
              id: this.editingId,
              model_name: this.form.model_name.trim(),
              model_list: claudeModels,
              api_key: this.form.api_key.trim(),
              base_url: this.form.base_url.trim(),
            }
            agentCliApi.AgentCliWriteDeepSeek(dsData, (dsResponse) => {
              this.saving = false
              if (dsResponse && dsResponse.ErrCode === 0) {
                this.dialogVisible = false
                this.$message.success('保存成功')
                this.loadList()
              } else {
                this.$message.error(dsResponse?.ErrMsg || '密钥保存失败')
              }
            })
            return
          }
          this.saving = false
          this.dialogVisible = false
          this.$message.success('保存成功')
          this.loadList()
        } else {
          this.saving = false
          this.$message.error(response?.ErrMsg || '保存失败')
        }
      })
    },
    deleteItem(item) {
      this.$confirm(`确定要删除 "${item.name}" 吗？` + (item.type !== 'codex-cli' ? '此操作不删除 settings.json 文件。' : ''), '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        agentCliApi.AgentCliDelete(item.id, (response) => {
          if (response && response.ErrCode === 0) {
            this.$message.success('删除成功')
            this.loadList()
          } else {
            this.$message.error(response?.ErrMsg || '删除失败')
          }
        })
      }).catch(() => {})
    },
    configureMcp(item) {
      const loading = this.$loading({ text: '正在写入 mcpServers 配置...' })
      agentCliApi.AgentCliWriteMcpServers(item.id, (response) => {
        loading.close()
        if (response && response.ErrCode === 0) {
          this.$message.success('ChromeDevtoolsMcp 配置已写入')
          this.loadList()
        } else {
          this.$message.error(response?.ErrMsg || '配置失败')
        }
      })
    },
    toggleClaudeMem(item) {
      item._togglingMem = true
      agentCliApi.AgentCliToggleClaudeMem({ id: item.id, enable: item.claude_mem_enabled }, (response) => {
        item._togglingMem = false
        if (response && response.ErrCode === 0) {
          this.$message.success(`claude-mem 已${item.claude_mem_enabled ? '启用' : '禁用'}`)
        } else {
          this.$message.error(response?.ErrMsg || '操作失败')
          item.claude_mem_enabled = !item.claude_mem_enabled
        }
      })
    },
    toggleEnabled(item, enabled) {
      const previousDisplayedEnabled = !!item.displayed_enabled
      item._togglingEnabled = true
      item.displayed_enabled = !!enabled
      // 启停切换后先本地重排，保证启用实例即时显示在顶部。 // Re-sort immediately after toggling so enabled rows move to the top without waiting for reload.
      this.list = this.sortAgentCliList(this.list)
      agentCliApi.AgentCliToggleEnabled({ id: item.id, enable: !!enabled }, (response) => {
        item._togglingEnabled = false
        if (response && response.ErrCode === 0) {
          this.$message.success(`Agent CLI 已${enabled ? '启用' : '停止'}`)
          this.loadList()
        } else {
          this.$message.error(response?.ErrMsg || '操作失败')
          item.displayed_enabled = previousDisplayedEnabled
          this.list = this.sortAgentCliList(this.list)
        }
      })
    },
    // 打开编辑对话框，预填当前条目数据并读取配置
    editItem(item) {
      this.editingId = item.id
      const isCodex = item.type === 'codex-cli'
      this.form = {
        name: item.name || '',
        type: item.type || 'claude-code-cli',
        settings_path: item.settings_path || '',
        webhook_config_id: item.webhook_config_id || '',
        enabled: item.enabled || 0,
        model_name: '',
        model_list_text: '',
        api_key: '',
        base_url: '',
        codex_api_key: '',
        codex_model_list_text: '',
        codex_base_url: '',
        codex_sandbox_mode: '',
      }
      // Codex: 从 config JSON 预填
      if (isCodex && item.config) {
        try {
          const cfg = JSON.parse(item.config)
          this.form.codex_api_key = cfg.api_key || ''
          this.form.codex_model_list_text = Array.isArray(cfg.models) ? cfg.models.join('\n') : (cfg.model || '')
          this.form.codex_base_url = cfg.base_url || ''
          this.form.codex_sandbox_mode = cfg.sandbox_mode || ''
        } catch (e) {}
      }
      this.dialogVisible = true
      // Claude: 读取 settings.json 以预填密钥字段
      if (!isCodex) {
        agentCliApi.AgentCliReadSettings(item.id, (response) => {
          if (response && response.ErrCode === 0 && response.Data && response.Data.content) {
            try {
              const config = JSON.parse(response.Data.content)
              this.form.model_name = config.model || ''
              this.form.model_list_text = Array.isArray(config.dtool_models) ? config.dtool_models.join('\n') : (config.model || '')
              if (config.env) {
                this.form.api_key = config.env.ANTHROPIC_AUTH_TOKEN || ''
                this.form.base_url = config.env.ANTHROPIC_BASE_URL || ''
              }
            } catch(e) {}
          }
        })
      }
    },
    // parseModelList 解析文本模型列表，并确保默认模型排在首位。
    // parseModelList parses textarea models and keeps the default model at the front.
    parseModelList(modelListText, defaultModel) {
      const list = String(modelListText || '')
        .split(/\r?\n/)
        .map(item => item.trim())
        .filter(Boolean)
      const merged = []
      const seen = new Set()
      const normalizedDefaultModel = String(defaultModel || '').trim()
      if (normalizedDefaultModel) {
        merged.push(normalizedDefaultModel)
        seen.add(normalizedDefaultModel)
      }
      list.forEach(modelName => {
        if (seen.has(modelName)) {
          return
        }
        merged.push(modelName)
        seen.add(modelName)
      })
      return merged
    },
    // formatModelOptions 统一格式化模型列表文案，避免空值时出现脏展示。
    // formatModelOptions normalizes the model list text and keeps the card display compact.
    formatModelOptions(modelOptions) {
      if (!Array.isArray(modelOptions) || modelOptions.length === 0) {
        return '-'
      }
      return modelOptions.join(' / ')
    },
    // ---- Webhook 配置相关 ----
    loadWebhookOptions() {
      agentCliApi.WebhookConfigList((response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.webhookOptions = response.Data.list || []
        }
      })
    },
    loadWebhookList() {
      this.webhookLoading = true
      agentCliApi.WebhookConfigList((response) => {
        this.webhookLoading = false
        if (response && response.ErrCode === 0 && response.Data) {
          this.webhookList = response.Data.list || []
          this.webhookOptions = this.webhookList
        }
      })
    },
    webhookTypeLabel(type) {
      const map = { dingtalk: '钉钉', feishu: '飞书', wecom: '企微' }
      return map[type] || type
    },
    openWebhookForm(row) {
      if (row) {
        this.webhookForm = {
          id: row.id,
          name: row.name,
          type: row.type,
          webhook_url: row.webhook_url,
          secret: row.secret,
        }
      } else {
        this.webhookForm = { id: 0, name: '', type: 'dingtalk', webhook_url: '', secret: '' }
      }
      this.webhookFormVisible = true
    },
    saveWebhook() {
      if (!this.webhookForm.name.trim()) {
        this.$message.warning('请输入配置名称')
        return
      }
      if (!this.webhookForm.webhook_url.trim()) {
        this.$message.warning('请输入 Webhook 地址')
        return
      }
      this.webhookSaving = true
      agentCliApi.WebhookConfigSave(this.webhookForm, (response) => {
        this.webhookSaving = false
        if (response && response.ErrCode === 0) {
          this.$message.success('保存成功')
          this.webhookFormVisible = false
          this.loadWebhookList()
        } else {
          this.$message.error(response?.ErrMsg || '保存失败')
        }
      })
    },
    deleteWebhook(row) {
      this.$confirm(`确定要删除 "${row.name}" 吗？`, '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        agentCliApi.WebhookConfigDelete(row.id, (response) => {
          if (response && response.ErrCode === 0) {
            this.$message.success('删除成功')
            this.loadWebhookList()
          } else {
            this.$message.error(response?.ErrMsg || '删除失败')
          }
        })
      }).catch(() => {})
    },
    updateWebhookConfig(item) {
      const data = {
        id: item.id,
        name: item.name,
        type: item.type,
        settings_path: item.settings_path || '',
        webhook_config_id: parseInt(item.webhook_config_id) || 0,
      }
      if (item.config) data.config = item.config
      agentCliApi.AgentCliSave(data, (response) => {
        if (response && response.ErrCode === 0) {
          this.$message.success('通知配置已更新')
        } else {
          this.$message.error(response?.ErrMsg || '更新失败')
        }
      })
    },
    // sortAgentCliList 将启用实例排在前面，同状态下按 ID 倒序，减少列表跳动并保留最近项优先。 // sortAgentCliList keeps enabled items first and orders same-state rows by descending ID.
    sortAgentCliList(items) {
      const sortedList = Array.isArray(items) ? [...items] : []
      sortedList.sort((firstItem, secondItem) => {
        const firstEnabledWeight = firstItem?.displayed_enabled ? AGENT_CLI_ENABLED_SORT_TRUE : AGENT_CLI_ENABLED_SORT_FALSE
        const secondEnabledWeight = secondItem?.displayed_enabled ? AGENT_CLI_ENABLED_SORT_TRUE : AGENT_CLI_ENABLED_SORT_FALSE
        if (firstEnabledWeight !== secondEnabledWeight) {
          return secondEnabledWeight - firstEnabledWeight
        }
        return (secondItem?.id || 0) - (firstItem?.id || 0)
      })
      return sortedList
    },
  },
}
</script>

<style scoped src="@/css/components/agent_cli/AgentCliList.css"></style>
