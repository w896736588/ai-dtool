<template>
  <div class="agent-cli-page">
    <div class="agent-cli-page__header">
      <span class="agent-cli-page__title">Agent Cli 管理</span>
      <el-button type="primary" size="small" @click="openCreateDialog">新建</el-button>
    </div>

    <div v-loading="loading" class="agent-cli-card-list">
      <div v-for="item in list" :key="item.id" class="agent-cli-card">
        <div class="agent-cli-card__header">
          <span class="agent-cli-card__name">{{ item.name }}</span>
          <el-tag size="small" type="info">{{ item.type }}</el-tag>
        </div>
        <div class="agent-cli-card__body">
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
            <span class="agent-cli-card__body-label">思考折叠:</span>
            <el-switch v-model="item.thinking_collapsed" :active-value="1" :inactive-value="0" active-text="开" inactive-text="关" size="small" @change="toggleThinkingCollapsed(item)" />
          </div>
        </div>
        <div class="agent-cli-card__footer">
          <el-button size="small" @click="configureMcp(item)">配置DevtoolsMcp</el-button>
          <el-button size="small" type="primary" @click="openDeepSeekDialog(item)">配置DeepSeek</el-button>
          <el-button size="small" type="danger" @click="deleteItem(item)">删除</el-button>
        </div>
      </div>

      <div v-if="!loading && list.length === 0" style="color: #909399; width: 100%; text-align: center; padding: 60px 0;">
        暂无 Agent Cli 实例，点击"新建"创建
      </div>
    </div>

    <!-- 新建/编辑对话框 -->
    <el-dialog v-model="dialogVisible" :title="editingId > 0 ? '编辑' : '新建 Agent Cli'" width="460px" :close-on-click-modal="false">
      <el-form :model="form" label-width="140px">
        <el-form-item label="名称">
          <el-input v-model="form.name" placeholder="默认 Claude Code CLI" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" style="width: 100%">
            <el-option label="Claude Code CLI" value="claude-code-cli" />
          </el-select>
        </el-form-item>
        <el-form-item label="settings.json 路径">
          <el-input v-model="form.settings_path" placeholder="请输入 settings.json 的绝对路径" />
          <div class="agent-cli-form-tip">例如: C:\Users\xxx\.claude\settings.json</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveItem">保存</el-button>
      </template>
    </el-dialog>

    <!-- 配置 DeepSeek 对话框 -->
    <el-dialog v-model="deepSeekDialogVisible" title="配置 DeepSeek" width="460px" :close-on-click-modal="false">
      <el-form :model="deepSeekForm" label-width="120px">
        <el-form-item label="模型名">
          <el-input v-model="deepSeekForm.model_name" placeholder="deepseek-v4-pro[1m]" />
        </el-form-item>
        <el-form-item label="API Key">
          <el-input v-model="deepSeekForm.api_key" type="password" show-password placeholder="请输入 DeepSeek API Key" />
        </el-form-item>
        <el-form-item label="Base URL">
          <el-input v-model="deepSeekForm.base_url" placeholder="https://api.deepseek.com/anthropic" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="deepSeekDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="deepSeekSaving" @click="saveDeepSeek">保存到 settings.json</el-button>
      </template>
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
      editingId: 0,
      saving: false,
      form: {
        name: '',
        type: 'claude-code-cli',
        settings_path: '',
      },
      // DeepSeek
      deepSeekDialogVisible: false,
      deepSeekSaving: false,
      deepSeekTargetId: 0,
      deepSeekForm: {
        model_name: 'deepseek-v4-pro[1m]',
        api_key: '',
        base_url: 'https://api.deepseek.com/anthropic',
      },
    }
  },
  mounted() {
    this.loadList()
  },
  methods: {
    loadList() {
      this.loading = true
      agentCliApi.AgentCliList((response) => {
        this.loading = false
        if (response && response.ErrCode === 0 && response.Data) {
          this.list = response.Data.list || []
        }
      })
    },
    openCreateDialog() {
      this.editingId = 0
      this.form = { name: '', type: 'claude-code-cli', settings_path: '' }
      this.dialogVisible = true
    },
    saveItem() {
      if (!this.form.settings_path.trim()) {
        this.$message.warning('请输入 settings.json 路径')
        return
      }
      this.saving = true
      const data = {
        id: this.editingId,
        name: this.form.name,
        type: this.form.type,
        settings_path: this.form.settings_path.trim(),
      }
      agentCliApi.AgentCliSave(data, (response) => {
        this.saving = false
        if (response && response.ErrCode === 0) {
          this.dialogVisible = false
          this.$message.success('保存成功')
          this.loadList()
        } else {
          this.$message.error(response?.ErrMsg || '保存失败')
        }
      })
    },
    deleteItem(item) {
      this.$confirm(`确定要删除 "${item.name}" 吗？此操作不删除 settings.json 文件。`, '确认删除', {
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
    openDeepSeekDialog(item) {
      this.deepSeekTargetId = item.id
      this.deepSeekForm = {
        model_name: 'deepseek-v4-pro[1m]',
        api_key: '',
        base_url: 'https://api.deepseek.com/anthropic',
      }
      this.deepSeekDialogVisible = true
    },
    toggleThinkingCollapsed(item) {
      const data = {
        id: item.id,
        name: item.name,
        type: item.type,
        settings_path: item.settings_path,
        thinking_collapsed: item.thinking_collapsed,
      }
      agentCliApi.AgentCliSave(data, () => {
        // 静默保存，无需提示
      })
    },
    saveDeepSeek() {
      if (!this.deepSeekForm.api_key.trim()) {
        this.$message.warning('请输入 API Key')
        return
      }
      if (!this.deepSeekForm.model_name.trim()) {
        this.$message.warning('请输入模型名')
        return
      }
      this.deepSeekSaving = true
      const data = {
        id: this.deepSeekTargetId,
        model_name: this.deepSeekForm.model_name.trim(),
        api_key: this.deepSeekForm.api_key.trim(),
        base_url: this.deepSeekForm.base_url.trim(),
      }
      agentCliApi.AgentCliWriteDeepSeek(data, (response) => {
        this.deepSeekSaving = false
        if (response && response.ErrCode === 0) {
          this.deepSeekDialogVisible = false
          this.$message.success('DeepSeek 配置已写入')
          this.loadList()
        } else {
          this.$message.error(response?.ErrMsg || '配置失败')
        }
      })
    },
  },
}
</script>

<style scoped src="@/css/components/agent_cli/AgentCliList.css"></style>
