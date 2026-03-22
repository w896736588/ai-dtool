<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">记忆配置</h3>
      <p class="set-config-desc">配置记忆专属仓库目录与 sqlite 文件名。保存后重启应用生效。</p>
      <div class="set-config-actions">
        <el-button type="primary" @click="saveConfig">保存</el-button>
      </div>
    </div>

    <div class="set-config-table-card">
      <el-form label-width="120px" class="memory-config-form">
        <el-form-item label="memory_dir">
          <el-input v-model="form.memory_dir" placeholder="例如 D:\\repo\\memory-data 或 /data/memory" />
        </el-form-item>
        <el-form-item label="memory_db_name">
          <el-input v-model="form.memory_db_name" placeholder="例如 memory.db" />
        </el-form-item>
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
      <el-alert
        :closable="false"
        type="info"
        title="启动时会先判断目录是否为 git 仓库；是则先 git pull，之后加载该 sqlite。AI 整理配置保存后可立即生效。"
      />
    </div>
  </div>
</template>

<script>
import set from '@/utils/base/git_set'
import AiSetApi from '@/utils/base/ai_set'

const DEFAULT_MEMORY_ARRANGE_PROMPT = '帮我把当前markdown进行整理格式，让它看起来更顺畅清晰，注意禁止修改内容'

export default {
  name: 'MemorySet',
  data() {
    return {
      aiModelList: [],
      form: {
        memory_dir: '',
        memory_db_name: '',
        memory_arrange_model_id: null,
        memory_arrange_prompt: DEFAULT_MEMORY_ARRANGE_PROMPT,
      }
    }
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
      AiSetApi.AiModelList({ model_type: 'llm' }, (response) => {
        if (response.ErrCode !== 0) {
          return
        }
        this.aiModelList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    loadConfig() {
      set.MemoryConfigGet((response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.form.memory_dir = response.Data.memory_dir || ''
        this.form.memory_db_name = response.Data.memory_db_name || ''
        this.form.memory_arrange_model_id = response.Data.memory_arrange_model_id || null
        this.form.memory_arrange_prompt = response.Data.memory_arrange_prompt || DEFAULT_MEMORY_ARRANGE_PROMPT
      })
    },
    saveConfig() {
      set.MemoryConfigSave(this.form, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('记忆配置已保存')
        }
      })
    }
  }
}
</script>

<style scoped>
@import "@/css/set_module_unified.css";

.memory-config-form {
  margin-bottom: 16px;
}
</style>
