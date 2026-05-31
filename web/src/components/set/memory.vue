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
          </el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">[smart_link]</el-divider>
        <el-descriptions class="memory-config-display" :column="1" border>
          <el-descriptions-item label="run_mode">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'run_mode'">
                <div class="config-edit-row">
                  <el-select v-model="editingItem.value" style="width: 200px">
                    <el-option label="server" value="server" />
                  </el-select>
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" @click="saveItem('smart_link', 'run_mode', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ form.run_mode || 'server' }}</div>
                  <GitActionButton compact size="small" @click="startEdit('run_mode', form.run_mode || 'server')">编辑</GitActionButton>
                </div>
              </template>
            </div>
          </el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">[base] 主库</el-divider>
        <el-descriptions class="memory-config-display" :column="1" border>
          <el-descriptions-item label="dbPath">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'db_path'">
                <div class="config-edit-row">
                  <el-input v-model="editingItem.value" style="flex: 1" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('base', 'db_path', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ form.db_dir || '-' }}</div>
                  <GitActionButton compact size="small" @click="startEdit('db_path', form.db_dir)">编辑</GitActionButton>
                </div>
              </template>
            </div>
          </el-descriptions-item>
          <el-descriptions-item label="dbFileName">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'db_file_name'">
                <div class="config-edit-row">
                  <el-input v-model="editingItem.value" style="flex: 1" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('base', 'dbFileName', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ form.db_name || '-' }}</div>
                  <GitActionButton compact size="small" @click="startEdit('db_file_name', form.db_name)">编辑</GitActionButton>
                </div>
              </template>
            </div>
          </el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">[base] 日志库</el-divider>
        <el-descriptions class="memory-config-display" :column="1" border>
          <el-descriptions-item label="logDbPath">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'log_db_path'">
                <div class="config-edit-row">
                  <el-input v-model="editingItem.value" style="flex: 1" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('base', 'logDbPath', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ form.log_db_path || '-' }}</div>
                  <GitActionButton compact size="small" @click="startEdit('log_db_path', form.log_db_path)">编辑</GitActionButton>
                </div>
              </template>
            </div>
          </el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">[base] 知识片段</el-divider>
        <el-descriptions class="memory-config-display" :column="1" border>
          <el-descriptions-item label="memoryDbPath">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'memory_db_path'">
                <div class="config-edit-row">
                  <el-input v-model="editingItem.value" style="flex: 1" />
                  <div class="config-edit-actions">
                    <GitActionButton compact size="small" :loading="saving" @click="saveItem('base', 'memoryDbPath', editingItem.value)">保存</GitActionButton>
                    <GitActionButton compact size="small" @click="cancelEdit">取消</GitActionButton>
                  </div>
                </div>
              </template>
              <template v-else>
                <div class="config-display-row">
                  <div class="config-value">{{ form.memory_dir || '-' }}</div>
                  <GitActionButton compact size="small" @click="startEdit('memory_db_path', form.memory_dir)">编辑</GitActionButton>
                </div>
              </template>
            </div>
          </el-descriptions-item>
        </el-descriptions>

        <el-divider content-position="left">[safe]</el-divider>
        <el-descriptions class="memory-config-display" :column="1" border>
          <el-descriptions-item label="password">
            <div class="config-item-wrapper">
              <template v-if="editingItem.key === 'safe_password'">
                <div class="config-edit-row">
                  <el-input v-model="editingItem.value" show-password style="flex: 1" />
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
            </div>
          </el-descriptions-item>
        </el-descriptions>
      </template>

      <el-form v-else label-width="120px" class="memory-config-form">
        <el-divider content-position="left">AI 整理</el-divider>
        <el-form-item label="整理模型">
          <el-select v-model="form.memory_arrange_model_id" clearable filterable style="width: 100%;">
            <el-option
              v-for="item in aiModelList"
              :key="item.id"
              :label="buildModelLabel(item)"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="整理提示词">
          <el-input v-model="form.memory_arrange_prompt" type="textarea" :rows="4" />
        </el-form-item>
        <el-divider content-position="left">AI 搜索</el-divider>
        <el-form-item label="搜索模型">
          <el-select v-model="form.memory_ai_search_model_id" clearable filterable style="width: 100%;">
            <el-option
              v-for="item in aiModelList"
              :key="item.id"
              :label="buildModelLabel(item)"
              :value="item.id"
            />
          </el-select>
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

const DEFAULT_MEMORY_ARRANGE_PROMPT = '帮我把当前 markdown 进行整理格式，让它看起来更顺畅清晰，注意禁止修改内容'

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
      saving: false,
      editingItem: createEditingItem(),
      form: {
        db_dir: '',
        db_name: '',
        db_configured: false,
        log_db_path: '',
        memory_dir: '',
        memory_db_configured: false,
        memory_config_file: '',
        memory_arrange_model_id: null,
        memory_arrange_prompt: DEFAULT_MEMORY_ARRANGE_PROMPT,
        memory_ai_search_model_id: null,
        safe_password: '',
        run_mode: 'server',
      },
    }
  },
  computed: {
    pageTitle() {
      return this.showRuntimeConfig ? '配置文件' : '知识片段 AI 设置'
    },
    pageDesc() {
      return this.showRuntimeConfig ? '这里可以查看并编辑当前运行配置。' : '这里维护知识片段相关的 AI 参数。'
    },
    runtimeConfigAlertType() {
      return this.form.db_configured && this.form.memory_db_configured ? 'info' : 'warning'
    },
    memoryConfigAlertTitle() {
      const configFile = this.form.memory_config_file || '配置文件'
      if (!this.form.db_configured) return `未检测到主库配置，请检查 ${configFile}`
      if (!this.form.memory_db_configured) return `未检测到知识片段目录配置，请检查 ${configFile}`
      return `当前配置来自 ${configFile}`
    },
    safePasswordDisplay() {
      return this.form.safe_password ? '已设置' : '未设置'
    },
  },
  mounted() {
    this.loadAiModelList()
    this.loadConfig()
  },
  methods: {
    buildModelLabel(item) {
      const provider = item.provider_name || '未命名服务商'
      const model = item.name || item.model || `模型#${item.id}`
      return `${provider} / ${model}`
    },
    loadAiModelList() {
      if (this.showRuntimeConfig) return
      AiSetApi.AiModelList({ model_type: 'llm' }, (response) => {
        if (response.__loginRequired || response.ErrCode !== 0) return
        this.aiModelList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    loadConfig() {
      set.MemoryConfigGet((response) => {
        if (response.__loginRequired || response.ErrCode !== 0 || !response.Data) return
        this.form.db_dir = response.Data.db_dir || ''
        this.form.db_name = response.Data.db_name || ''
        this.form.db_configured = !!response.Data.db_configured
        this.form.log_db_path = response.Data.log_db_path || ''
        this.form.memory_dir = response.Data.memory_dir || ''
        this.form.memory_db_configured = !!response.Data.memory_db_configured
        this.form.memory_config_file = response.Data.memory_config_file || ''
        this.form.memory_arrange_model_id = response.Data.memory_arrange_model_id || null
        this.form.memory_arrange_prompt = response.Data.memory_arrange_prompt || DEFAULT_MEMORY_ARRANGE_PROMPT
        this.form.memory_ai_search_model_id = response.Data.memory_ai_search_model_id || null
        this.form.safe_password = response.Data.safe_password || ''
        this.form.run_mode = response.Data.run_mode || 'server'
      })
    },
    startEdit(key, value) {
      this.editingItem = { key, value: value === null || value === undefined ? '' : value }
    },
    cancelEdit() {
      this.editingItem = createEditingItem()
    },
    saveItem(section, key, value) {
      this.saving = true
      set.RuntimeConfigItemSave({ section, key, value }, (response) => {
        this.saving = false
        if (response.__loginRequired) return
        if (response.ErrCode !== 0) {
          this.$helperNotify.error(response.ErrMsg || '保存失败')
          return
        }
        this.$helperNotify.success('保存成功')
        this.editingItem = createEditingItem()
        this.loadConfig()
        if (response.Data && response.Data.need_relogin) {
          this.$base.ClearSafeToken()
          if (this.$eventBus) {
            this.$eventBus.emit('safe_auth_required', { message: '密码已修改，请重新登录' })
          }
          return
        }
        this.$emit('changed')
      })
    },
    saveAiConfig() {
      const payload = {
        memory_arrange_model_id: this.form.memory_arrange_model_id,
        memory_arrange_prompt: this.form.memory_arrange_prompt,
        memory_ai_search_model_id: this.form.memory_ai_search_model_id,
      }
      set.MemoryConfigSave(payload, (response) => {
        if (response.__loginRequired) return
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

<style scoped src="@/css/components/set/memory.css"></style>
