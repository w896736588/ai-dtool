<template>
  <el-dialog
    :model-value="modelValue"
    width="82%"
    top="5vh"
    class="memory-history-dialog"
    title="变更历史"
    @close="handleClose"
  >
    <div class="history-layout">
      <div class="history-list-card">
        <div class="history-head">
          <span>{{ historyTitleText }}</span>
          <el-tag size="small" effect="light">{{ historyList.length }}</el-tag>
        </div>
        <el-alert
          v-if="showGitSettingHint"
          type="info"
          :closable="false"
          class="history-setting-alert"
        >
          <template #title>
            <div class="history-setting-alert-title">
              <span>{{ settingHint }}</span>
              <pl-button type="primary" link @click="emitOpenSettings">打开设置</pl-button>
            </div>
          </template>
        </el-alert>
        <el-table
          v-loading="loading"
          :data="historyList"
          border
          stripe
          max-height="58vh"
          :empty-text="historyEmptyText"
          @row-click="selectHistory"
        >
          <el-table-column prop="create_time_desc" label="变更时间" width="180" />
          <el-table-column prop="change_desc" label="变更摘要" min-width="240" />
        </el-table>
      </div>

      <div class="history-detail-card">
        <div class="history-head">
          <span>内容对比</span>
          <el-tag v-if="activeHistory" size="small" effect="light">{{ activeHistory.create_time_desc }}</el-tag>
        </div>
        <el-empty v-if="!activeHistory" :description="detailEmptyText" />
        <template v-else>
          <div class="history-tags-panel">
            <div class="tags-block">
              <div class="tags-title">旧标签</div>
              <div class="tags-list">
                <el-tag
                  v-for="tag in activeHistory.tags_old_list || []"
                  :key="'old-' + tag"
                  size="small"
                  effect="plain"
                >
                  {{ tag }}
                </el-tag>
                <span v-if="!(activeHistory.tags_old_list || []).length" class="tags-empty">无</span>
              </div>
            </div>
            <div class="tags-block">
              <div class="tags-title">新标签</div>
              <div class="tags-list">
                <el-tag
                  v-for="tag in activeHistory.tags_new_list || []"
                  :key="'new-' + tag"
                  size="small"
                  effect="plain"
                  type="success"
                >
                  {{ tag }}
                </el-tag>
                <span v-if="!(activeHistory.tags_new_list || []).length" class="tags-empty">无</span>
              </div>
            </div>
          </div>
          <div class="title-compare">
            <div class="title-box">
              <div class="title-label">旧标题</div>
              <div class="title-content">{{ activeHistory.title_old || '无标题' }}</div>
            </div>
            <div class="title-box">
              <div class="title-label">新标题</div>
              <div class="title-content">{{ activeHistory.title_new || '无标题' }}</div>
            </div>
          </div>
          <diff-markdown
            :old-text="activeHistory.content_old || ''"
            :new-text="activeHistory.content_new || ''"
            title="正文差异"
          />
        </template>
      </div>
    </div>
  </el-dialog>
</template>

<script>
import DiffMarkdown from '@/components/base/diff_markwodn.vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'

export default {
  name: 'MemoryHistoryDialog',
  components: {
    DiffMarkdown,
  },
  props: {
    modelValue: {
      type: Boolean,
      default: false,
    },
    fragmentId: {
      type: [Number, String],
      default: 0,
    },
    gitRepoEnabled: {
      type: Boolean,
      default: false,
    },
    isGitRepo: {
      type: Boolean,
      default: false,
    },
  },
  emits: ['update:modelValue', 'open-settings'],
  data() {
    return {
      loading: false,
      historyList: [],
      activeHistory: null,
      historySource: 'none',
      settingHint: '当前知识片段目录不是 Git 仓库，无法查看历史记录。',
    }
  },
  computed: {
    hasValidFragmentId() {
      const text = String(this.fragmentId || '').trim()
      return text !== '' && text !== '0' && text !== 'null' && text !== 'undefined'
    },
    historyTitleText() {
      return this.historySource === 'git' ? 'Git 历史记录' : '历史记录'
    },
    showGitSettingHint() {
      return !this.isGitRepo
    },
    historyEmptyText() {
      return this.showGitSettingHint ? '当前不是 Git 仓库' : '暂无历史记录'
    },
    detailEmptyText() {
      return this.showGitSettingHint ? this.settingHint : '选择一条历史记录后查看差异'
    },
  },
  watch: {
    modelValue: {
      immediate: true,
      handler(value) {
        if (value && this.hasValidFragmentId) {
          this.loadHistory()
        }
        if (!value) {
          this.activeHistory = null
        }
      },
    },
    fragmentId(value) {
      const text = String(value || '').trim()
      if (this.modelValue && text !== '' && text !== '0' && text !== 'null' && text !== 'undefined') {
        this.loadHistory()
      }
    },
  },
  methods: {
    loadHistory() {
      if (!this.hasValidFragmentId) return
      this.loading = true
      MemoryFragmentApi.MemoryFragmentHistoryList(this.fragmentId, (response) => {
        this.loading = false
        const responseData = response && response.Data ? response.Data : {}
        this.historyList = Array.isArray(responseData.list) ? responseData.list : []
        this.activeHistory = this.historyList.length > 0 ? this.historyList[0] : null
        this.historySource = responseData.history_source || 'none'
        if (responseData.setting_hint) {
          this.settingHint = responseData.setting_hint
        }
      })
    },
    selectHistory(row) {
      this.activeHistory = row
    },
    handleClose() {
      this.$emit('update:modelValue', false)
    },
    emitOpenSettings() {
      this.$emit('open-settings')
    },
  },
}
</script>

<style scoped src="@/css/components/memory/MemoryHistoryDialog.css"></style>
