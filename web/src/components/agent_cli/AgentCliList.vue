<template>
  <div class="agent-cli-page">
    <div class="agent-cli-page__header">
      <span class="agent-cli-page__title">Agent Cli 管理</span>
      <div class="agent-cli-page__actions">
        <el-button type="primary" size="small" @click="openCreateDialog">新建</el-button>
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
        </div>
        <div class="agent-cli-card__footer">
          <el-button size="small" @click="configureMcp(item)">配置DevtoolsMcp</el-button>
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
        <el-form-item label="模型名">
          <el-input v-model="form.model_name" placeholder="deepseek-v4-pro[1m]" />
        </el-form-item>
        <el-form-item label="API Key">
          <el-input v-model="form.api_key" type="password" show-password placeholder="请输入 DeepSeek API Key" />
        </el-form-item>
        <el-form-item label="Base URL">
          <el-input v-model="form.base_url" placeholder="https://api.deepseek.com/anthropic" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveItem">保存</el-button>
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
      chromeDevtoolsDialogVisible: false,
      editingId: 0,
      saving: false,
      form: {
        name: '',
        type: 'claude-code-cli',
        settings_path: '',
        model_name: '',
        api_key: '',
        base_url: '',
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
      this.form = { name: '', type: 'claude-code-cli', settings_path: '', model_name: '', api_key: '', base_url: '' }
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
        if (response && response.ErrCode === 0) {
          // 新建时从返回值取 ID，后续 DeepSeek 写入依赖此 ID
          if (!this.editingId && response.Data && response.Data.id) {
            this.editingId = response.Data.id
          }
          // 密钥字段非空时，一并写入 DeepSeek 配置
          if (this.form.model_name.trim() && this.form.api_key.trim()) {
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
    // 打开编辑对话框，预填当前条目数据并读取 settings.json 中的密钥
    editItem(item) {
      this.editingId = item.id
      this.form = {
        name: item.name || '',
        type: item.type || 'claude-code-cli',
        settings_path: item.settings_path || '',
        model_name: '',
        api_key: '',
        base_url: '',
      }
      this.dialogVisible = true
      // 读取 settings.json 以预填密钥字段
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
    },
  },
}
</script>

<style scoped src="@/css/components/agent_cli/AgentCliList.css"></style>
