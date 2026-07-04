<template>
  <div class="agent-config">
    <div class="config-header">
      <el-button text @click="$router.back()">
        <el-icon><ArrowLeft /></el-icon>
        返回
      </el-button>
      <h2>{{ agentName }} 配置</h2>
    </div>

    <el-tabs v-model="activeTab" class="config-tabs">
      <!-- 基本配置 -->
      <el-tab-pane label="基本配置" name="basic">
        <el-form :model="configForm" label-width="120px" style="max-width: 600px;">
          <el-form-item label="Agent 名称">
            <el-input v-model="configForm.name" placeholder="给 Agent 起个名字" />
          </el-form-item>
          <el-form-item label="Agent 类型">
            <el-tag>{{ typeLabel(configForm.type) }}</el-tag>
          </el-form-item>
          <el-divider content-position="left">LLM 配置</el-divider>
          <el-form-item label="Provider">
            <el-select v-model="piConfig.provider" style="width:100%">
              <el-option label="Anthropic (Claude)" value="anthropic" />
              <el-option label="OpenAI (GPT-4o)" value="openai" />
              <el-option label="DeepSeek" value="deepseek" />
              <el-option label="Google (Gemini)" value="google" />
            </el-select>
            <div class="field-hint">LLM 服务提供商，选择 DeepSeek 直接使用官方接口无需填写模型 API 地址</div>
          </el-form-item>
          <el-form-item label="默认模型">
            <el-input v-model="piConfig.model" placeholder="claude-sonnet-4-20250514" />
            <div class="field-hint">启动 Pi 时使用的默认模型 ID</div>
          </el-form-item>
          <el-form-item label="模型 API 地址" v-show="piConfig.provider !== 'deepseek'">
            <el-input v-model="piConfig.model_addr" placeholder="留空使用默认地址，例如 https://api.example.com/v1" />
            <div class="field-hint">自定义 LLM API 端点，适用于代理或兼容接口（对应 --model-addr 参数）</div>
          </el-form-item>
          <el-form-item label="API Key">
            <el-input v-model="piConfig.api_key" type="password" show-password placeholder="留空使用系统环境变量" />
            <div class="field-hint">API 认证密钥，根据 Provider 自动映射为对应环境变量（ANTHROPIC_API_KEY / OPENAI_API_KEY / GOOGLE_API_KEY）</div>
          </el-form-item>
          <el-form-item label="可选模型">
            <div class="model-tag-list">
              <el-tag
                v-for="(m, idx) in modelList"
                :key="idx"
                closable
                :disable-transitions="false"
                @close="removeModel(idx)"
                style="margin: 0 6px 6px 0;"
              >
                {{ m }}
              </el-tag>
              <el-input
                v-if="showModelInput"
                ref="modelInputRef"
                v-model="modelInputValue"
                size="small"
                style="width:200px"
                placeholder="输入模型ID，回车确认"
                @keyup.enter="addModel"
                @blur="addModel"
              />
              <el-button v-else size="small" @click="showModelInput = true">+ 添加模型</el-button>
            </div>
            <div class="field-hint">可在对话中切换的模型列表</div>
          </el-form-item>
          <el-divider content-position="left">高级选项</el-divider>
          <el-form-item label="会话存储目录">
            <el-input v-model="piConfig.session_dir" placeholder="留空使用默认目录" />
            <div class="field-hint">Pi 会话 JSONL 文件的存储路径，留空则默认 data/agent_sessions/</div>
          </el-form-item>
          <el-form-item label="额外启动参数">
            <el-input v-model="piConfig.extra_args" placeholder="例如：--no-session" />
            <div class="field-hint">空格分隔的额外命令行参数，如 --no-session 禁用会话持久化</div>
          </el-form-item>

          <el-form-item label="安装状态">
            <el-tag :type="installed ? 'success' : 'danger'">
              {{ installed ? '已安装' : '未安装' }}
            </el-tag>
            <span v-if="!installed" style="margin-left: 8px; color: #909399; font-size: 12px;">
              {{ installHint }}
            </span>
            <el-button size="small" @click="checkInstall" style="margin-left:12px">重新检测</el-button>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" @click="saveConfig">保存配置</el-button>
          </el-form-item>
        </el-form>
      </el-tab-pane>

      <!-- Skills / 自定义工具 -->
      <el-tab-pane label="Skills & Tools" name="skills">
        <div class="skills-toolbar">
          <el-button type="primary" size="small" @click="openSkillAdd">添加 Skill</el-button>
          <span class="skills-hint">Skills 是可复用的按需能力模块，让 Agent 加载特定场景的专业知识或工具</span>
        </div>

        <el-table :data="skills" style="margin-top: 16px;" empty-text="暂无 Skills">
          <el-table-column prop="name" label="名称" min-width="140" />
          <el-table-column prop="skill_type" label="类型" width="100">
            <template #default="{ row }">
              <el-tag size="small" :type="row.skill_type === 'tool' ? 'warning' : ''">
                {{ row.skill_type === 'tool' ? '工具' : '技能' }}
              </el-tag>
            </template>
          </el-table-column>
          <el-table-column label="描述" min-width="180">
            <template #default="{ row }">
              <span style="color:#909399;font-size:13px">{{ skillDesc(row) }}</span>
            </template>
          </el-table-column>
          <el-table-column prop="enabled" label="状态" width="80">
            <template #default="{ row }">
              <el-switch :model-value="row.enabled === 1" @change="toggleSkill(row)" size="small" />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="120">
            <template #default="{ row }">
              <el-button text size="small" @click="openSkillEdit(row)">编辑</el-button>
              <el-button text size="small" type="danger" @click="deleteSkill(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>

      <!-- 工作空间 -->
      <el-tab-pane label="工作空间" name="workspaces">
        <div style="margin-bottom:12px">
          <el-button type="primary" size="small" @click="showWorkspaceDialog = true">添加工作空间</el-button>
        </div>
        <el-table :data="workspaces" empty-text="暂无工作空间，请先添加">
          <el-table-column prop="name" label="名称" width="180" />
          <el-table-column prop="path" label="路径" min-width="300" />
          <el-table-column label="操作" width="80">
            <template #default="{ row }">
              <el-button text size="small" type="danger" @click="deleteWorkspace(row)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- Skill 编辑对话框 -->
    <el-dialog v-model="showSkillDialog" :title="editingSkillId ? '编辑 Skill' : '添加 Skill'" width="580px" :close-on-click-modal="false">
      <el-form :model="skillForm" label-width="100px">
        <el-form-item label="名称" required>
          <el-input v-model="skillForm.name" placeholder="Skill 唯一标识名称" />
        </el-form-item>
        <el-form-item label="类型" required>
          <el-radio-group v-model="skillForm.skill_type">
            <el-radio value="skill">Skill（按需技能）</el-radio>
            <el-radio value="tool">Tool（自定义工具）</el-radio>
          </el-radio-group>
        </el-form-item>
        <el-form-item label="描述">
          <el-input v-model="skillForm.description" type="textarea" :rows="2" placeholder="描述此 Skill/Tool 的功能" />
        </el-form-item>

        <!-- Skill 专属字段 -->
        <template v-if="skillForm.skill_type === 'skill'">
          <el-form-item label="命令名">
            <el-input v-model="skillForm.command" placeholder="斜杠命令名，如 my-review" />
            <div class="field-hint">Agent 通过 /命令名 触发此 Skill</div>
          </el-form-item>
          <el-form-item label="提示词内容">
            <el-input v-model="skillForm.prompt" type="textarea" :rows="6" placeholder="Skill 的详细提示词/指令内容" />
            <div class="field-hint">当 Agent 加载此 Skill 时执行的系统提示词</div>
          </el-form-item>
        </template>

        <!-- Tool 专属字段 -->
        <template v-if="skillForm.skill_type === 'tool'">
          <el-form-item label="工具名">
            <el-input v-model="skillForm.tool_name" placeholder="工具函数名，如 search_code" />
            <div class="field-hint">Agent 调用的函数名</div>
          </el-form-item>
          <el-form-item label="工具描述">
            <el-input v-model="skillForm.tool_description" type="textarea" :rows="2" placeholder="描述工具的功能" />
          </el-form-item>
          <el-form-item label="参数定义">
            <div class="param-list">
              <div v-for="(p, idx) in skillForm.parameters" :key="idx" class="param-row">
                <el-input v-model="p.name" placeholder="参数名" size="small" style="width:120px" />
                <el-select v-model="p.type" size="small" style="width:90px">
                  <el-option label="string" value="string" />
                  <el-option label="number" value="number" />
                  <el-option label="boolean" value="boolean" />
                </el-select>
                <el-input v-model="p.description" placeholder="参数描述" size="small" style="flex:1" />
                <el-checkbox v-model="p.required" size="small" style="margin-left:4px">必填</el-checkbox>
                <el-button text size="small" type="danger" @click="skillForm.parameters.splice(idx, 1)">×</el-button>
              </div>
              <el-button size="small" @click="skillForm.parameters.push({ name:'', type:'string', description:'', required:false })">
                + 添加参数
              </el-button>
            </div>
          </el-form-item>
        </template>

        <el-form-item label="启用">
          <el-switch v-model="skillForm.enabled" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showSkillDialog = false">取消</el-button>
        <el-button type="primary" @click="saveSkill" :disabled="!skillForm.name.trim()">保存</el-button>
      </template>
    </el-dialog>

    <!-- 工作空间对话框 -->
    <el-dialog v-model="showWorkspaceDialog" title="添加工作空间" width="480px" :close-on-click-modal="false">
      <el-form label-width="60px">
        <el-form-item label="名称">
          <el-input v-model="workspaceForm.name" placeholder="例如：my-project" />
        </el-form-item>
        <el-form-item label="路径">
          <el-input v-model="workspaceForm.path" placeholder="例如：C:/work/my-project" />
          <div class="field-hint">本地项目的绝对路径</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="showWorkspaceDialog = false">取消</el-button>
        <el-button type="primary" @click="saveWorkspace">保存</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import Base from '@/utils/base.js'
import { ArrowLeft } from '@element-plus/icons-vue'

export default {
  name: 'AgentConfig',
  components: { ArrowLeft },
  data() {
    return {
      activeTab: 'basic',
      agentId: 0,
      agentName: '',
      installed: false,
      installHint: '',

      configForm: { name: '', type: '' },
      piConfig: { provider: 'anthropic', model: '', model_addr: '', api_key: '', session_dir: '', extra_args: '' },
      modelList: [],
      showModelInput: false,
      modelInputValue: '',

      skills: [],
      showSkillDialog: false,
      editingSkillId: null,
      skillForm: this.emptySkillForm(),

      workspaces: [],
      showWorkspaceDialog: false,
      workspaceForm: { name: '', path: '' }
    }
  },
  watch: {
    'piConfig.provider'(val) {
      if (val === 'deepseek') {
        this.piConfig.model_addr = ''
      }
    }
  },
  mounted() {
    this.agentId = parseInt(this.$route.query.agent_id) || 0
    if (!this.agentId) {
      this.$router.push('/AgentHub')
      return
    }
    this.loadData()
  },
  methods: {
    emptySkillForm() {
      return {
        name: '', skill_type: 'skill', description: '', enabled: true,
        command: '', prompt: '',
        tool_name: '', tool_description: '', parameters: []
      }
    },
    loadData() {
      Base.BasePost('/api/AgentV2List', {}, (res) => {
        const agents = (res.ErrCode === 0 && res.Data && res.Data.list) ? res.Data.list : []
        const agent = agents.find(a => a.id === this.agentId)
        if (agent) {
          this.agentName = agent.name
          this.installed = agent.installed
          this.installHint = agent.install_hint
          this.configForm = { name: agent.name, type: agent.type }
          if (agent.config) {
            try {
              const cfg = JSON.parse(agent.config)
              if (cfg.provider) this.piConfig.provider = cfg.provider
              if (cfg.model) this.piConfig.model = cfg.model
              if (cfg.model_addr) this.piConfig.model_addr = cfg.model_addr
              if (cfg.api_key) this.piConfig.api_key = cfg.api_key
              if (cfg.session_dir) this.piConfig.session_dir = cfg.session_dir
              if (cfg.extra_args) this.piConfig.extra_args = cfg.extra_args
              if (cfg.models && Array.isArray(cfg.models)) {
                this.modelList = cfg.models
              }
            } catch (e) {}
          }
        }
      })
      this.loadSkills()
      this.loadWorkspaces()
    },

    // 模型标签管理
    addModel() {
      const val = this.modelInputValue.trim()
      if (val && !this.modelList.includes(val)) {
        this.modelList.push(val)
      }
      this.modelInputValue = ''
      this.showModelInput = false
    },
    removeModel(idx) {
      this.modelList.splice(idx, 1)
    },

    saveConfig() {
      const configObj = {
        provider: this.piConfig.provider,
        model: this.piConfig.model,
        model_addr: this.piConfig.provider === 'deepseek' ? '' : this.piConfig.model_addr,
        api_key: this.piConfig.api_key,
        models: this.modelList,
        session_dir: this.piConfig.session_dir,
        extra_args: this.piConfig.extra_args
      }
      const config = JSON.stringify(configObj)
      Base.BasePost('/api/AgentV2Save', {
        id: this.agentId,
        name: this.configForm.name,
        type: this.configForm.type,
        config: config
      }, () => {
        this.$message.success('配置已保存')
        this.loadData()
      })
    },
    checkInstall() {
      Base.BasePost('/api/AgentV2CheckInstall', { type: this.configForm.type }, (res) => {
        this.installed = (res.ErrCode === 0 && res.Data) ? res.Data.installed : false
        this.installHint = (res.ErrCode === 0 && res.Data) ? (res.Data.install_hint || '') : ''
      })
    },

    // Skills
    loadSkills() {
      Base.BasePost('/api/AgentV2SkillList', { agent_id: this.agentId }, (res) => {
        this.skills = (res.ErrCode === 0 && res.Data && res.Data.list) ? res.Data.list : []
      })
    },
    skillDesc(row) {
      try {
        const cfg = JSON.parse(row.config || '{}')
        return cfg.description || '-'
      } catch (e) { return '-' }
    },
    openSkillAdd() {
      this.editingSkillId = null
      this.skillForm = this.emptySkillForm()
      this.showSkillDialog = true
    },
    openSkillEdit(row) {
      this.editingSkillId = row.id
      try {
        const cfg = JSON.parse(row.config || '{}')
        this.skillForm = {
          name: row.name,
          skill_type: row.skill_type,
          description: cfg.description || '',
          enabled: row.enabled === 1,
          command: cfg.command || '',
          prompt: cfg.prompt || '',
          tool_name: cfg.tool_name || '',
          tool_description: cfg.tool_description || '',
          parameters: cfg.parameters || []
        }
      } catch (e) {
        this.skillForm = {
          name: row.name,
          skill_type: row.skill_type,
          description: '',
          enabled: row.enabled === 1,
          command: '', prompt: '',
          tool_name: '', tool_description: '', parameters: []
        }
      }
      this.showSkillDialog = true
    },
    saveSkill() {
      const s = this.skillForm
      let configObj = {}
      if (s.skill_type === 'skill') {
        configObj = {
          description: s.description,
          command: s.command,
          prompt: s.prompt
        }
      } else {
        configObj = {
          description: s.description || s.tool_description,
          tool_name: s.tool_name,
          tool_description: s.tool_description,
          parameters: s.parameters.filter(p => p.name.trim()).map(p => ({
            name: p.name, type: p.type, description: p.description, required: !!p.required
          }))
        }
      }

      Base.BasePost('/api/AgentV2SkillSave', {
        id: this.editingSkillId || undefined,
        agent_id: this.agentId,
        name: s.name,
        skill_type: s.skill_type,
        config: JSON.stringify(configObj),
        enabled: s.enabled ? 1 : 0
      }, () => {
        this.showSkillDialog = false
        this.loadSkills()
      })
    },
    toggleSkill(row) {
      Base.BasePost('/api/AgentV2SkillSave', {
        id: row.id, agent_id: this.agentId,
        name: row.name, skill_type: row.skill_type,
        config: row.config,
        enabled: row.enabled === 1 ? 0 : 1
      }, () => { this.loadSkills() })
    },
    deleteSkill(row) {
      this.$confirm('确定删除此 Skill？', '提示', { type: 'warning' }).then(() => {
        Base.BasePost('/api/AgentV2SkillDelete', { id: row.id }, () => { this.loadSkills() })
      }).catch(() => {})
    },

    // 工作空间
    loadWorkspaces() {
      Base.BasePost('/api/AgentV2WorkspaceList', { agent_id: this.agentId }, (res) => {
        this.workspaces = (res.ErrCode === 0 && res.Data && res.Data.list) ? res.Data.list : []
      })
    },
    saveWorkspace() {
      if (!this.workspaceForm.name || !this.workspaceForm.path) {
        this.$message.warning('请填写名称和路径')
        return
      }
      Base.BasePost('/api/AgentV2WorkspaceSave', {
        agent_id: this.agentId,
        name: this.workspaceForm.name,
        path: this.workspaceForm.path
      }, () => {
        this.showWorkspaceDialog = false
        this.workspaceForm = { name: '', path: '' }
        this.loadWorkspaces()
      })
    },
    deleteWorkspace(row) {
      this.$confirm('确定删除此工作空间？', '提示', { type: 'warning' }).then(() => {
        Base.BasePost('/api/AgentV2WorkspaceDelete', { id: row.id }, () => { this.loadWorkspaces() })
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
.agent-config { padding: 24px; max-width: 960px; margin: 0 auto; }
.config-header { display: flex; align-items: center; gap: 12px; margin-bottom: 24px; }
.config-header h2 { margin: 0; font-size: 20px; }
.config-tabs { background: #fff; border-radius: 8px; padding: 20px; border: 1px solid #e4e7ed; }
.skills-toolbar { display: flex; align-items: center; gap: 16px; }
.skills-hint { font-size: 12px; color: #909399; }
.field-hint { font-size: 11px; color: #c0c4cc; margin-top: 4px; line-height: 1.4; }
.model-tag-list { display: flex; flex-wrap: wrap; align-items: center; }
.param-list { width: 100%; }
.param-row { display: flex; gap: 6px; align-items: center; margin-bottom: 6px; }
</style>
