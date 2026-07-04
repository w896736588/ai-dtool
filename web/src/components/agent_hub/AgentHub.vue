<template>
  <div class="agent-hub">
    <div class="agent-hub__header">
      <h2>Agent Hub</h2>
      <p class="agent-hub__subtitle">选择 AI Agent 开始工作</p>
    </div>

    <div class="agent-hub__grid">
      <div
        v-for="agent in agents"
        :key="agent.id"
        class="agent-card"
        :class="{
          'agent-card--installed': agent.installed,
          'agent-card--not-installed': !agent.installed
        }"
      >
        <div class="agent-card__icon">
          <span v-if="agent.type === 'pi'" class="agent-icon agent-icon--pi">π</span>
          <span v-else-if="agent.type === 'codex'" class="agent-icon agent-icon--codex">C</span>
          <span v-else class="agent-icon agent-icon--claude">A</span>
        </div>
        <div class="agent-card__body">
          <h3>{{ agent.name }}</h3>
          <p class="agent-card__type">{{ typeLabel(agent.type) }}</p>
          <p class="agent-card__sessions">{{ agent.session_count }} 个会话</p>
        </div>
        <div class="agent-card__actions">
          <template v-if="agent.installed">
            <el-button type="primary" size="small" @click="openChat(agent)">开始对话</el-button>
            <el-button size="small" @click="openConfig(agent)">配置</el-button>
          </template>
          <template v-else>
            <el-tooltip :content="agent.install_hint" placement="top">
              <el-button type="info" size="small" disabled>未安装</el-button>
            </el-tooltip>
          </template>
        </div>
        <div class="agent-card__footer">
          <el-button text size="small" type="danger" @click.stop="deleteAgent(agent)">删除</el-button>
        </div>
      </div>

      <div class="agent-card agent-card--add" @click="showAddDialog = true">
        <div class="agent-card__add-icon">+</div>
        <p>添加 Agent</p>
      </div>
    </div>

    <!-- 添加 Agent 对话框 -->
    <el-dialog v-model="showAddDialog" title="添加 Agent" width="420px" :close-on-click-modal="false">
      <el-form :model="addForm" label-width="60px">
        <el-form-item label="名称">
          <el-input v-model="addForm.name" placeholder="例如：我的 Pi" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="addForm.type" placeholder="选择 Agent 类型" style="width:100%">
            <el-option label="Pi" value="pi" />
            <el-option label="Codex" value="codex" :disabled="true" />
            <el-option label="Claude Code" value="claude-code" :disabled="true" />
          </el-select>
        </el-form-item>
        <div class="form-hint">创建后可进入配置页进行详细设置</div>
      </el-form>
      <template #footer>
        <el-button @click="showAddDialog = false">取消</el-button>
        <el-button type="primary" @click="saveAgent" :disabled="!addForm.name.trim()">创建</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import Base from '@/utils/base.js'

export default {
  name: 'AgentHub',
  data() {
    return {
      agents: [],
      showAddDialog: false,
      addForm: { name: '', type: 'pi' }
    }
  },
  mounted() {
    this.loadAgents()
  },
  methods: {
    loadAgents() {
      Base.BasePost('/api/AgentV2List', {}, (res) => {
        if (res.ErrCode === 0 && res.Data && res.Data.list) {
          this.agents = res.Data.list
        }
      })
    },
    saveAgent() {
      Base.BasePost('/api/AgentV2Save', {
        name: this.addForm.name.trim(),
        type: this.addForm.type,
        config: '{}'
      }, () => {
        this.showAddDialog = false
        this.addForm = { name: '', type: 'pi' }
        this.loadAgents()
      })
    },
    openChat(agent) {
      this.$router.push({ path: '/AgentChat', query: { agent_id: agent.id } })
    },
    openConfig(agent) {
      this.$router.push({ path: '/AgentConfig', query: { agent_id: agent.id } })
    },
    deleteAgent(agent) {
      this.$confirm(`确定删除 Agent「${agent.name}」？关联的工作空间、会话、Skills 都会被删除。`, '确认删除', {
        type: 'warning',
        confirmButtonText: '删除',
        cancelButtonText: '取消'
      }).then(() => {
        Base.BasePost('/api/AgentV2Delete', { id: agent.id }, () => {
          this.loadAgents()
        })
      }).catch(() => {})
    },
    typeLabel(type) {
      const map = { pi: 'Pi', codex: 'Codex CLI', 'claude-code': 'Claude Code' }
      return map[type] || type
    }
  }
}
</script>

<style scoped>
.agent-hub { padding: 24px; max-width: 1200px; margin: 0 auto; }
.agent-hub__header { margin-bottom: 32px; }
.agent-hub__header h2 { margin: 0 0 8px; font-size: 24px; color: #303133; }
.agent-hub__subtitle { color: #909399; margin: 0; }
.agent-hub__grid { display: grid; grid-template-columns: repeat(auto-fill, minmax(280px, 1fr)); gap: 20px; }

.agent-card {
  background: #fff; border-radius: 12px; padding: 24px;
  border: 1px solid #e4e7ed; transition: all .2s;
  display: flex; flex-direction: column; align-items: center; text-align: center; gap: 12px;
  position: relative;
}
.agent-card:hover { box-shadow: 0 4px 16px rgba(0,0,0,.08); border-color: #c0c4cc; }
.agent-card--not-installed { opacity: 0.6; }
.agent-card--add {
  border: 2px dashed #dcdfe6; cursor: pointer; justify-content: center; min-height: 200px;
}
.agent-card--add:hover { border-color: #409eff; background: #ecf5ff; }
.agent-card__add-icon { font-size: 40px; color: #c0c4cc; }
.agent-card--add:hover .agent-card__add-icon { color: #409eff; }
.agent-card--add p { color: #909399; margin: 0; }

.agent-icon {
  width: 64px; height: 64px; border-radius: 16px; display: flex; align-items: center; justify-content: center;
  font-size: 28px; font-weight: bold; color: #fff;
}
.agent-icon--pi { background: linear-gradient(135deg, #667eea, #764ba2); }
.agent-icon--codex { background: linear-gradient(135deg, #10a37f, #1a7f64); }
.agent-icon--claude { background: linear-gradient(135deg, #d97706, #b45309); }

.agent-card__body h3 { margin: 0; font-size: 16px; color: #303133; }
.agent-card__type { color: #909399; font-size: 13px; margin: 4px 0; text-transform: uppercase; }
.agent-card__sessions { color: #c0c4cc; font-size: 12px; margin: 0; }
.agent-card__actions { display: flex; gap: 8px; margin-top: 4px; }
.agent-card__footer { position: absolute; top: 8px; right: 12px; }
.form-hint { font-size: 12px; color: #909399; padding-left: 60px; }
</style>
