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
          <span>历史记录</span>
          <el-tag size="small" effect="light">{{ historyList.length }}</el-tag>
        </div>
        <el-table
          v-loading="loading"
          :data="historyList"
          border
          stripe
          max-height="58vh"
          empty-text="暂无历史记录"
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
        <el-empty v-if="!activeHistory" description="选择一条历史记录后查看差异" />
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
      default: false
    },
    fragmentId: {
      type: Number,
      default: 0
    }
  },
  emits: ['update:modelValue'],
  data() {
    return {
      loading: false,
      historyList: [],
      activeHistory: null,
    }
  },
  watch: {
    // modelValue 打开时重新获取历史记录。
    modelValue: {
      immediate: true,
      handler(value) {
        if (value && this.fragmentId > 0) {
          this.loadHistory()
        }
        if (!value) {
          this.activeHistory = null
        }
      }
    },
    // fragmentId 变化后在弹窗打开状态下同步刷新内容。
    fragmentId(value) {
      if (this.modelValue && value > 0) {
        this.loadHistory()
      }
    }
  },
  methods: {
    // loadHistory 加载片段历史记录。
    loadHistory() {
      if (this.fragmentId <= 0) {
        return
      }
      this.loading = true
      MemoryFragmentApi.MemoryFragmentHistoryList(this.fragmentId, (response) => {
        this.loading = false
        this.historyList = Array.isArray(response.Data) ? response.Data : []
        this.activeHistory = this.historyList.length > 0 ? this.historyList[0] : null
      })
    },
    // selectHistory 选中一条历史记录。
    selectHistory(row) {
      this.activeHistory = row
    },
    // handleClose 关闭弹窗。
    handleClose() {
      this.$emit('update:modelValue', false)
    }
  }
}
</script>

<style scoped>
.history-layout {
  display: grid;
  grid-template-columns: 360px minmax(0, 1fr);
  gap: 16px;
}

.history-list-card,
.history-detail-card {
  border: 1px solid #e8e8e0;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 4px 12px rgba(54, 74, 54, 0.05);
  padding: 16px;
}

.history-head {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
  color: #4d5c47;
  font-weight: 600;
}

.history-detail-card {
  min-height: 66vh;
  display: flex;
  flex-direction: column;
}

.history-tags-panel {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 14px;
}

.tags-block {
  border: 1px solid #edf1e8;
  border-radius: 12px;
  background: #fafbf8;
  padding: 12px;
}

.tags-title {
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #5b6b55;
}

.tags-list {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.tags-empty {
  color: #8b9685;
  font-size: 13px;
}

.title-compare {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 12px;
  margin-bottom: 14px;
}

.title-box {
  border: 1px solid #edf1e8;
  border-radius: 12px;
  background: #fafbf8;
  padding: 12px;
}

.title-label {
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 600;
  color: #5b6b55;
}

.title-content {
  color: #34412f;
  line-height: 1.6;
  word-break: break-word;
}

@media (max-width: 1100px) {
  .history-layout {
    grid-template-columns: 1fr;
  }

  .history-tags-panel,
  .title-compare {
    grid-template-columns: 1fr;
  }
}
</style>
