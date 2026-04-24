<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">工作日报 AI 设置</h3>
      <p class="set-config-desc">这里维护任务清单右侧"AI 生成工作日报"按钮使用的模型和提示词。</p>
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
        <el-form-item>
          <pl-button type="primary" @click="saveConfig">保存工作日报配置</pl-button>
        </el-form-item>
      </el-form>
    </div>

    <div class="set-config-header" style="margin-top: 24px;">
      <h3 class="set-config-title">TAPD 登录页配置</h3>
      <p class="set-config-desc">
        从自定义网页中选择一个链接，用于在任务中快速跳转到 TAPD 登录页。
      </p>
    </div>

    <div class="set-config-table-card">
      <el-form label-width="120px" class="memory-config-form">
        <el-form-item label="自定义网页">
          <el-select
            v-model="form.home_task_tapd_smart_link_id"
            clearable
            filterable
            style="width: 100%;"
            placeholder="请选择自定义网页"
            @change="onSmartLinkChange"
          >
            <el-option
              v-for="item in smartLinkList"
              :key="item.id"
              :label="item.name"
              :value="item.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="网页链接">
          <el-select
            v-model="form.home_task_tapd_link_label"
            clearable
            filterable
            style="width: 100%;"
            placeholder="请选择具体链接"
          >
            <el-option
              v-for="(link, idx) in currentLinkOptions"
              :key="idx"
              :label="link.label"
              :value="link.label"
            />
          </el-select>
        </el-form-item>
        <el-form-item>
          <pl-button type="primary" @click="saveTapdConfig">保存 TAPD 登录页配置</pl-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script>
import set from '@/utils/base/git_set'
import AiSetApi from '@/utils/base/ai_set'
import SmartLinkSet from '@/utils/base/smart_link_set'

const DEFAULT_MEMORY_ARRANGE_PROMPT = '帮我把当前 markdown 进行整理格式，让它看起来更顺畅清晰，注意禁止修改内容'
const DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT = '请基于当前活跃任务生成中文工作日报，按已完成、进行中、风险与阻塞三个部分总结，输出 Markdown，禁止编造未提供的信息。'

export default {
  name: 'HomeTaskReportSetting',
  emits: ['changed'],
  data() {
    return {
      aiModelList: [],
      smartLinkList: [],
      form: {
        memory_arrange_model_id: null,
        memory_arrange_prompt: DEFAULT_MEMORY_ARRANGE_PROMPT,
        home_task_daily_report_model_id: null,
        home_task_daily_report_prompt: DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT,
        home_task_fragment_prompt: '',
        home_task_tapd_smart_link_id: null,
        home_task_tapd_link_label: '',
      },
    }
  },
  computed: {
    currentLinkOptions() {
      if (!this.form.home_task_tapd_smart_link_id) return []
      const smartLink = this.smartLinkList.find(item => item.id === this.form.home_task_tapd_smart_link_id)
      if (!smartLink || !smartLink.links) return []
      try {
        return JSON.parse(smartLink.links)
      } catch {
        return []
      }
    },
  },
  mounted() {
    this.loadAiModelList()
    this.loadSmartLinkList()
    this.loadConfig()
  },
  methods: {
    buildModelLabel(item) {
      const provider = item.provider_name || '未命名服务商'
      const model = item.name || item.model || `模型#${item.id}`
      return `${provider} / ${model}`
    },
    onSmartLinkChange() {
      this.form.home_task_tapd_link_label = ''
    },
    loadAiModelList() {
      AiSetApi.AiModelList({ model_type: 'llm' }, (response) => {
        if (response.ErrCode !== 0) {
          return
        }
        this.aiModelList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    loadSmartLinkList() {
      SmartLinkSet.SmartLinkList((response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.smartLinkList = Array.isArray(response.Data.smart_link_list) ? response.Data.smart_link_list : []
      })
    },
    loadConfig() {
      set.MemoryConfigGet((response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.form.memory_arrange_model_id = response.Data.memory_arrange_model_id || null
        this.form.memory_arrange_prompt = response.Data.memory_arrange_prompt || DEFAULT_MEMORY_ARRANGE_PROMPT
        this.form.home_task_daily_report_model_id = response.Data.home_task_daily_report_model_id || null
        this.form.home_task_daily_report_prompt = response.Data.home_task_daily_report_prompt || DEFAULT_HOME_TASK_DAILY_REPORT_PROMPT
        this.form.home_task_fragment_prompt = response.Data.home_task_fragment_prompt || ''
        this.form.home_task_tapd_smart_link_id = response.Data.home_task_tapd_smart_link_id || null
        this.form.home_task_tapd_link_label = response.Data.home_task_tapd_link_label || ''
      })
    },
    saveConfig() {
      const payload = {
        memory_arrange_model_id: this.form.memory_arrange_model_id,
        memory_arrange_prompt: this.form.memory_arrange_prompt,
        home_task_daily_report_model_id: this.form.home_task_daily_report_model_id,
        home_task_daily_report_prompt: this.form.home_task_daily_report_prompt,
        home_task_fragment_prompt: this.form.home_task_fragment_prompt,
        home_task_tapd_smart_link_id: this.form.home_task_tapd_smart_link_id,
        home_task_tapd_link_label: this.form.home_task_tapd_link_label,
      }
      set.MemoryConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('配置已保存')
          this.$emit('changed')
        }
      })
    },
    saveTapdConfig() {
      const payload = {
        memory_arrange_model_id: this.form.memory_arrange_model_id,
        memory_arrange_prompt: this.form.memory_arrange_prompt,
        home_task_daily_report_model_id: this.form.home_task_daily_report_model_id,
        home_task_daily_report_prompt: this.form.home_task_daily_report_prompt,
        home_task_fragment_prompt: this.form.home_task_fragment_prompt,
        home_task_tapd_smart_link_id: this.form.home_task_tapd_smart_link_id,
        home_task_tapd_link_label: this.form.home_task_tapd_link_label,
      }
      set.MemoryConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('TAPD 登录页配置已保存')
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
