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
          <el-descriptions-item label="空间分析">
            <div class="config-item-wrapper">
              <div class="config-display-row">
                <div class="config-value">{{ mainDbStorageSummaryText }}</div>
                <GitActionButton compact size="small" :class="{ 'config-alert-button': mainDbStorageAlertVisible }" @click="openMainDBStorageDialog">
                  空间分析
                  <span v-if="mainDbStorageAlertVisible" class="config-inline-alert-dot"></span>
                </GitActionButton>
              </div>
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
        <el-divider content-position="left">分享</el-divider>
        <el-form-item label="分享地址">
          <el-input
            v-model="form.memory_share_base_url"
            clearable
            placeholder="如：https://example.com，留空按当前访问地址生成"
          />
        </el-form-item>
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
          <GitActionButton type="primary" @click="saveAiConfig">保存配置</GitActionButton>
        </el-form-item>
      </el-form>
    </div>

    <el-dialog v-model="mainDbStorageDialogVisible" title="主库空间分析" width="760px">
      <div v-loading="mainDbStorageLoading">
        <el-alert
          v-if="mainDbStorageError"
          :closable="false"
          type="error"
          :title="mainDbStorageError"
        />
        <template v-else-if="mainDbStorage">
          <el-descriptions :column="1" border class="memory-config-display">
            <el-descriptions-item label="主库文件">{{ mainDbStorage.db_path || '-' }}</el-descriptions-item>
            <el-descriptions-item label="文件大小">
              {{ mainDbStorage.file_size_text || '-' }}
              <span v-if="mainDbStorage.exceeds_limit" class="config-danger-text">，已超过 100 MB</span>
            </el-descriptions-item>
            <el-descriptions-item label="清理空闲空间">
              <GitActionButton
                type="warning"
                :loading="mainDbStorageVacuuming"
                @click="runMainDBVacuum"
              >
                清理空闲空间
              </GitActionButton>
            </el-descriptions-item>
          </el-descriptions>
          <el-table :data="mainDbStorage.tables || []" stripe style="width: 100%; margin-top: 12px;">
            <el-table-column prop="name" label="表名" min-width="260" />
            <el-table-column prop="total_size_text" label="空间占用" width="160" />
            <el-table-column prop="page_count" label="页数" width="120" />
          </el-table>
        </template>
      </div>
    </el-dialog>
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
      mainDbStorageDialogVisible: false,
      mainDbStorageLoading: false,
      mainDbStorageVacuuming: false,
      mainDbStorageError: '',
      mainDbStorage: null,
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
        memory_share_base_url: '',
        safe_password: '',
      },
    }
  },
  computed: {
    pageTitle() {
      return this.showRuntimeConfig ? '配置文件' : '知识片段设置'
    },
    pageDesc() {
      return this.showRuntimeConfig ? '这里可以查看并编辑当前运行配置。' : '这里维护知识片段相关配置。'
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
    mainDbStorageAlertVisible() {
      return !!this.mainDbStorage?.exceeds_limit
    },
    mainDbStorageSummaryText() {
      if (!this.mainDbStorage) return '点击查看主库与各表空间占用'
      return `${this.mainDbStorage.file_size_text || '-'} / 阈值 100.00 MB`
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
    normalizeShareBaseUrl(value) {
      return String(value || '').trim().replace(/\/+$/, '')
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
        this.form.memory_share_base_url = response.Data.memory_share_base_url || ''
        this.form.safe_password = response.Data.safe_password || ''
        this.mainDbStorage = response.Data.main_db_storage || null
        this.broadcastMainDbStorageAlert()
      })
    },
    broadcastMainDbStorageAlert() {
      if (!this.$eventBus) return
      this.$eventBus.emit('main_db_storage_alert_changed', {
        exceeds_limit: !!this.mainDbStorage?.exceeds_limit,
        data: this.mainDbStorage,
      })
    },
    openMainDBStorageDialog() {
      this.mainDbStorageDialogVisible = true
      this.loadMainDBStorageAnalysis()
    },
    loadMainDBStorageAnalysis() {
      this.mainDbStorageLoading = true
      this.mainDbStorageError = ''
      set.MainDBStorageAnalysis((response) => {
        this.mainDbStorageLoading = false
        if (response.__loginRequired) return
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.mainDbStorageError = response?.ErrMsg || '空间分析加载失败'
          return
        }
        this.mainDbStorage = response.Data
        this.broadcastMainDbStorageAlert()
      })
    },
    runMainDBVacuum() {
      if (this.mainDbStorageVacuuming) return
      this.$confirm('将执行 VACUUM; 回收主库空闲页。执行期间可能持续一段时间，是否继续？', '清理空闲空间', {
        confirmButtonText: '执行',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        this.mainDbStorageVacuuming = true
        set.MainDBStorageVacuum((response) => {
          this.mainDbStorageVacuuming = false
          if (response.__loginRequired) return
          if (!(response && response.ErrCode === 0 && response.Data)) {
            this.$message.error(response?.ErrMsg || '清理失败')
            return
          }
          this.mainDbStorage = response.Data
          this.mainDbStorageError = ''
          this.broadcastMainDbStorageAlert()
          this.$message.success(response?.ErrMsg || '清理完成')
        })
      }).catch(() => {})
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
        memory_share_base_url: this.normalizeShareBaseUrl(this.form.memory_share_base_url),
      }
      set.MemoryConfigSave(payload, (response) => {
        if (response.__loginRequired) return
        if (response.ErrCode === 0) {
          this.form.memory_share_base_url = payload.memory_share_base_url
          this.$helperNotify.success('配置已保存')
          this.$emit('changed')
          return
        }
        this.$helperNotify.error(response.ErrMsg || '配置保存失败')
      })
    },
  },
}
</script>

<style scoped src="@/css/components/set/memory.css"></style>
<style scoped>
.config-inline-alert-dot {
  display: inline-block;
  width: 8px;
  height: 8px;
  border-radius: 999px;
  background: #f04438;
  margin-left: 6px;
  vertical-align: middle;
}

.config-alert-button {
  position: relative;
}

.config-danger-text {
  color: #f04438;
}
</style>
