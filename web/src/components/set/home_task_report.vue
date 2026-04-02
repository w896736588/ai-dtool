<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">工作日报 AI 设置</h3>
      <p class="set-config-desc">这里维护任务清单右侧“AI 生成工作日报”按钮使用的模型和提示词。</p>
      <div class="set-config-actions">
        <pl-button type="primary" @click="saveConfig">保存工作日报配置</pl-button>
      </div>
    </div>

    <div class="set-config-table-card">
      <el-form label-width="120px" class="memory-config-form">
        <el-form-item label="日报模型">
          <el-select
            v-model="form.home_task_daily_report_model_id"
            clearable
            filterable
            style="width: 100%;"
            placeholder="请选择用于生成工作日报的 LLM 模型"
          >
            <el-option
              v-for="item in aiModelList"
              :key="item.id"
              :label="buildModelLabel(item)"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="日报提示词">
          <el-input
            v-model="form.home_task_daily_report_prompt"
            type="textarea"
            :rows="5"
            placeholder="请输入首页任务工作日报提示词"
          />
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script>
import set from '@/utils/base/git_set'
import AiSetApi from '@/utils/base/ai_set'

// DEFAULT_MEMORY_ARRANGE_PROMPT 作为透传字段保留知识片段整理配置。 // Keep memory arrange config in state so daily-report saves do not overwrite it.
const DEFAULT_MEMORY_ARRANGE_PROMPT = '帮我把当前 markdown 进行整理格式，让它看起来更顺畅清晰，注意禁止修改内容'
// DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT 定义首页任务日报默认提示词。 // Default prompt used when generating the home task daily report.
const DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT = '请基于当前活跃任务生成中文工作日报，按已完成、进行中、风险与阻塞三个部分总结，输出 Markdown，禁止编造未提供的信息。'

export default {
  name: 'HomeTaskReportSetting',
  emits: ['changed'],
  data() {
    return {
      aiModelList: [],
      form: {
        memory_arrange_model_id: null,
        memory_arrange_prompt: DEFAULT_MEMORY_ARRANGE_PROMPT,
        home_task_daily_report_model_id: null,
        home_task_daily_report_prompt: DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT,
      },
    }
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
    loadAiModelList() {
      AiSetApi.AiModelList({ model_type: 'llm' }, (response) => {
        if (response.ErrCode !== 0) {
          return
        }
        this.aiModelList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    // loadConfig 读取完整 AI 配置，并保留知识片段整理字段以便保存时透传。 // Load the full AI config and preserve memory fields for save pass-through.
    loadConfig() {
      set.MemoryConfigGet((response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.form.memory_arrange_model_id = response.Data.memory_arrange_model_id || null
        this.form.memory_arrange_prompt = response.Data.memory_arrange_prompt || DEFAULT_MEMORY_ARRANGE_PROMPT
        this.form.home_task_daily_report_model_id = response.Data.home_task_daily_report_model_id || null
        this.form.home_task_daily_report_prompt = response.Data.home_task_daily_report_prompt || DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT
      })
    },
    // saveConfig 保存工作日报 AI 配置，并透传知识片段整理配置防止被覆盖。 // Save daily-report AI config and pass through memory config to avoid overwriting it.
    saveConfig() {
      const payload = {
        memory_arrange_model_id: this.form.memory_arrange_model_id,
        memory_arrange_prompt: this.form.memory_arrange_prompt,
        home_task_daily_report_model_id: this.form.home_task_daily_report_model_id,
        home_task_daily_report_prompt: this.form.home_task_daily_report_prompt,
      }
      set.MemoryConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('工作日报 AI 配置已保存')
          this.$emit('changed')
        }
      })
    },
  },
}
</script>

<style scoped>
@import "@/css/set_module_unified.css";

.memory-config-form {
  margin-bottom: 0;
}
</style>
