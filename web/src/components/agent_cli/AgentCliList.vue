<template>
  <div class="agent-cli-page">
    <div class="agent-cli-page__header">
      <span class="agent-cli-page__title">Agent Cli 管理</span>
      <div class="agent-cli-page__actions">
        <el-button type="primary" size="small" @click="openCreateDialog">新建</el-button>
        <el-button type="primary" size="small" @click="webhookDialogVisible = true; loadWebhookList()">Webhook 配置</el-button>
        <el-button type="primary" size="small" @click="chromeDevtoolsDialogVisible = true">ChromeDevTools</el-button>
      </div>
    </div>

    <div v-loading="loading" class="agent-cli-card-list">
      <div v-for="item in list" :key="item.id" class="agent-cli-card">
        <div class="agent-cli-card__header">
          <span class="agent-cli-card__name">{{ item.name }}</span>
          <el-tag size="small" type="info">{{ item.type }}</el-tag>
        </div>
        <div class="agent-cli-card__body">
          <template v-if="item.type === 'codex-cli'">
            <div class="agent-cli-card__body-item">
              <span class="agent-cli-card__body-label">模型:</span>
              <span>{{ item.current_model || '-' }}</span>
            </div>
          </template>
          <template v-else>
            <div class="agent-cli-card__body-item">
              <span class="agent-cli-card__body-label">路径:</span>
              <span>{{ item.settings_path }}</span>
            </div>
            <div class="agent-cli-card__body-item">
              <span class="agent-cli-card__body-label">配置文件:</span>
              <el-tag :type="item.settings_exists ? 'success' : 'danger'" size="small">{{ item.settings_exists ? '存在' : '不存在' }}</el-tag>
            </div>
            <div class="agent-cli-card__body-item">
              <span class="agent-cli-card__body-label">模型:</span>
              <span>{{ item.current_model || '-' }}</span>
            </div>
            <div class="agent-cli-card__body-item">
              <span class="agent-cli-card__body-label">McpServers:</span>
              <span>{{ item.mcp_server_count || 0 }} 个</span>
            </div>
            <div class="agent-cli-card__body-item agent-cli-card__body-switch">
              <span class="agent-cli-card__body-label">claude-mem:</span>
              <el-switch
                v-model="item.claude_mem_enabled"
                size="small"
                :loading="item._togglingMem"
                @change="toggleClaudeMem(item)"
              />
              <span style="font-size: 12px; color: #909399; margin-left: 6px;">{{ item.claude_mem_enabled ? '已启用' : '已禁用' }}</span>
            </div>
          </template>
          <div class="agent-cli-card__body-item agent-cli-card__body-webhook">
            <span class="agent-cli-card__body-label">通知:</span>
            <el-select
              v-model="item.webhook_config_id"
              size="small"
              placeholder="未配置"
              clearable
              style="width: 160px;"
              @change="updateWebhookConfig(item)"
            >
              <el-option
                v-for="wh in webhookOptions"
                :key="wh.id"
                :label="wh.name"
                :value="String(wh.id)"
              />
            </el-select>
          </div>
        </div>
        <div class="agent-cli-card__footer">
          <el-button v-if="item.type !== 'codex-cli'" size="small" @click="configureMcp(item)">配置DevtoolsMcp</el-button>
          <el-button size="small" @click="editItem(item)">编辑</el-button>
          <el-button size="small" type="danger" @click="deleteItem(item)">删除</el-button>
        </div>
      </div>

      <div v-if="!loading && list.length === 0" style="color: #909399; width: 100%; text-align: center; padding: 60px 0;">
        暂无 Agent Cli 实例，点击"新建"创建
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
    <el-dialog v-model="dialogVisible" :title="editingId > 0 ? '编辑' : '新建 Agent Cli'" width="460px" :close-on-click-modal="false">
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
            <el-input v-model="form.model_name" placeholder="deepseek-v4-pro[1m]" />
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
          <el-form-item label="模型" required>
            <el-input v-model="form.codex_model" placeholder="o3" />
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

export default {
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
        model_name: '',
        api_key: '',
        base_url: '',
        // Codex CLI 专属字段
        codex_api_key: '',
        codex_model: '',
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
    loadList() {
      this.loading = true
      agentCliApi.AgentCliList((response) => {
        this.loading = false
        if (response && response.ErrCode === 0 && response.Data) {
          const items = response.Data.list || []
          items.forEach(item => {
            item.webhook_config_id = item.webhook_config_id ? String(item.webhook_config_id) : ''
          })
          this.list = items
        }
      })
    },
    openCreateDialog() {
      this.editingId = 0
      this.form = { name: '', type: 'claude-code-cli', settings_path: '', webhook_config_id: '', model_name: '', api_key: '', base_url: '', codex_api_key: '', codex_model: '', codex_base_url: '', codex_sandbox_mode: '' }
      this.dialogVisible = true
    },
    onTypeChange() {
      // 切换类型时清空对方字段
    },
    saveItem() {
      const isCodex = this.form.type === 'codex-cli'
      if (isCodex) {
        if (!this.form.codex_api_key.trim()) {
          this.$message.warning('请输入 API Key')
          return
        }
        if (!this.form.codex_model.trim()) {
          this.$message.warning('请输入模型名')
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
        webhook_config_id: parseInt(this.form.webhook_config_id) || 0,
      }
      // Codex 类型：将配置序列化为 config JSON
      if (isCodex) {
        data.config = JSON.stringify({
          api_key: this.form.codex_api_key.trim(),
          model: this.form.codex_model.trim(),
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
    // 打开编辑对话框，预填当前条目数据并读取配置
    editItem(item) {
      this.editingId = item.id
      const isCodex = item.type === 'codex-cli'
      this.form = {
        name: item.name || '',
        type: item.type || 'claude-code-cli',
        settings_path: item.settings_path || '',
        webhook_config_id: item.webhook_config_id || '',
        model_name: '',
        api_key: '',
        base_url: '',
        codex_api_key: '',
        codex_model: '',
        codex_base_url: '',
        codex_sandbox_mode: '',
      }
      // Codex: 从 config JSON 预填
      if (isCodex && item.config) {
        try {
          const cfg = JSON.parse(item.config)
          this.form.codex_api_key = cfg.api_key || ''
          this.form.codex_model = cfg.model || ''
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
              if (config.env) {
                this.form.api_key = config.env.ANTHROPIC_AUTH_TOKEN || ''
                this.form.base_url = config.env.ANTHROPIC_BASE_URL || ''
              }
            } catch(e) {}
          }
        })
      }
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
  },
}
</script>

<style scoped src="@/css/components/agent_cli/AgentCliList.css"></style>
