<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">定时任务</h3>
      <p class="set-config-desc">设置每天固定时间自动触发功能，服务运行期间后台自动执行。</p>
    </div>

    <div class="set-config-table-card">
      <el-form label-width="120px" class="memory-config-form">
        <el-form-item label="启用定时任务">
          <el-switch v-model="form.cron_daily_report_enabled" :active-value="1" :inactive-value="0" />
        </el-form-item>
        <el-form-item label="触发时间">
          <el-time-picker
            v-model="timePickerValue"
            format="HH:mm"
            placeholder="选择每天触发时间"
            :disabled="form.cron_daily_report_enabled !== 1"
            style="width: 200px;"
          />
        </el-form-item>
        <el-form-item label="触发功能">
          <el-select v-model="cronTaskType" disabled style="width: 100%;">
            <el-option label="AI 生成工作日报" value="daily_report" />
          </el-select>
          <div class="cron-field-help">当前仅支持"AI 生成工作日报"，后续将扩展更多功能。</div>
        </el-form-item>
        <el-form-item>
          <pl-button type="primary" @click="saveConfig">保存定时任务配置</pl-button>
        </el-form-item>
      </el-form>
    </div>
  </div>
</template>

<script>
import set from '@/utils/base/git_set'

export default {
  name: 'CronSetting',
  emits: ['changed'],
  data() {
    return {
      cronTaskType: 'daily_report',
      form: {
        cron_daily_report_enabled: 0,
        cron_daily_report_time: '',
      },
    }
  },
  computed: {
    timePickerValue: {
      get() {
        if (!this.form.cron_daily_report_time) {
          return null
        }
        const [h, m] = this.form.cron_daily_report_time.split(':').map(Number)
        const d = new Date()
        d.setHours(h, m, 0, 0)
        return d
      },
      set(val) {
        if (val) {
          const h = String(val.getHours()).padStart(2, '0')
          const m = String(val.getMinutes()).padStart(2, '0')
          this.form.cron_daily_report_time = `${h}:${m}`
        } else {
          this.form.cron_daily_report_time = ''
        }
      },
    },
  },
  mounted() {
    this.loadConfig()
  },
  methods: {
    loadConfig() {
      set.CronConfigGet((response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.form.cron_daily_report_enabled = response.Data.cron_daily_report_enabled || 0
        this.form.cron_daily_report_time = response.Data.cron_daily_report_time || ''
      })
    },
    saveConfig() {
      if (this.form.cron_daily_report_enabled === 1 && !this.form.cron_daily_report_time) {
        this.$helperNotify.error('启用定时任务时触发时间不能为空')
        return
      }
      const payload = {
        cron_daily_report_enabled: this.form.cron_daily_report_enabled,
        cron_daily_report_time: this.form.cron_daily_report_time,
      }
      set.CronConfigSave(payload, (response) => {
        if (response.ErrCode === 0) {
          this.$helperNotify.success('定时任务配置已保存')
          this.$emit('changed')
        } else {
          this.$helperNotify.error(response.ErrMsg || '保存失败')
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

.cron-field-help {
  margin-top: 4px;
  color: #7a7a6a;
  font-size: 12px;
}
</style>
