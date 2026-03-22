<template>
  <div class="memory-editor" @keydown.ctrl.s.prevent="handleSave">
    <div class="editor-toolbar">
      <div class="toolbar-main">
        <el-input
          v-model="draftFragment.title"
          class="title-input"
          placeholder="输入片段标题"
          @input="handleFormChange"
        />
        <div class="toolbar-status">
          <el-tag size="small" :type="dirty ? 'warning' : 'success'" effect="light">
            {{ dirty ? '未保存' : '已保存' }}
          </el-tag>
          <el-tag size="small" effect="plain">
            {{ draftFragment.index_status_desc || '待索引' }}
          </el-tag>
          <span class="toolbar-time">更新于 {{ draftFragment.update_time_desc || '-' }}</span>
        </div>
      </div>
      <div class="toolbar-actions">
        <el-button plain @click="$emit('show-history', draftFragment.id)">
          <el-icon><Clock /></el-icon>
          历史记录
        </el-button>
        <el-popconfirm
          title="确定删除这个片段吗？"
          confirm-button-text="删除"
          cancel-button-text="取消"
          @confirm="handleDelete"
        >
          <template #reference>
            <el-button type="danger" plain>
              <el-icon><Delete /></el-icon>
              删除
            </el-button>
          </template>
        </el-popconfirm>
        <el-button type="primary" :loading="saving" @click="handleSave">
          <el-icon><Check /></el-icon>
          保存
        </el-button>
      </div>
    </div>

    <div class="tag-panel">
        <div class="tag-list">
          <el-tag
          v-for="tag in draftFragment.tags"
          :key="tag"
          size="small"
          closable
          @close="removeTag(tag)"
        >
          {{ tag }}
        </el-tag>
      </div>
      <el-input
        v-model="tagInput"
        class="tag-input"
        placeholder="输入标签后回车，可用逗号分隔"
        @keydown.enter.prevent="appendTag"
        @keydown="handleTagKeydown"
        @blur="appendTag"
      />
    </div>

    <div class="editor-body">
      <div class="editor-body-toolbar">
        <div class="editor-body-title">内容</div>
        <div class="editor-body-actions">
          <el-button
            :type="!contentEditMode ? 'primary' : 'default'"
            plain
            @click="setContentEditMode(false)"
          >
            查看
          </el-button>
          <el-button
            :type="contentEditMode ? 'primary' : 'default'"
            plain
            @click="setContentEditMode(true)"
          >
            编辑
          </el-button>
          <el-button
            type="primary"
            plain
            :loading="organizing"
            @click="handleOrganize"
          >
            <el-icon><MagicStick /></el-icon>
            AI整理
          </el-button>
        </div>
      </div>

      <div v-if="contentEditMode" class="editor-body-content">
        <MdEditor
          v-model="draftFragment.content"
          :preview-theme="'github'"
          :toolbars="toolbars"
          @onChange="handleFormChange"
          @onBlur="handleFormChange"
        />
      </div>
      <div v-else class="preview-body">
        <MdPreview
          :model-value="draftFragment.content"
          :preview-theme="'github'"
        />
      </div>
    </div>

    <el-dialog
      v-model="organizeDialogVisible"
      title="AI整理结果对比"
      width="84%"
      top="5vh"
      class="memory-organize-dialog"
    >
      <div class="organize-dialog-layout">
        <div class="organize-dialog-summary">
          <div class="summary-item">
            <span class="summary-label">整理模型</span>
            <span class="summary-value">{{ organizeResult.model || '-' }}</span>
          </div>
          <div class="summary-item">
            <span class="summary-label">整理提示词</span>
            <span class="summary-value">{{ organizeResult.prompt || '-' }}</span>
          </div>
        </div>
        <diff-markdown
          :old-text="draftFragment.content || ''"
          :new-text="organizeResult.content || ''"
          title="正文差异"
        />
      </div>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="organizeDialogVisible = false">取消</el-button>
          <el-button
            type="primary"
            :loading="applyingOrganizeResult"
            @click="applyOrganizeResult"
          >
            确认写入
          </el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { MdEditor, MdPreview } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import { Check, Clock, Delete, MagicStick } from '@element-plus/icons-vue'
import DiffMarkdown from '@/components/base/diff_markwodn.vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'

export default {
  name: 'MemoryEditor',
  components: {
    MdEditor,
    MdPreview,
    Check,
    Clock,
    Delete,
    MagicStick,
    DiffMarkdown,
  },
  props: {
    fragment: {
      type: Object,
      required: true
    },
    savedFragment: {
      type: Object,
      required: true
    }
  },
  emits: ['change', 'saved', 'deleted', 'show-history'],
  data() {
    return {
      saving: false,
      organizing: false,
      applyingOrganizeResult: false,
      contentEditMode: false,
      organizeDialogVisible: false,
      organizeResult: {
        content: '',
        model: '',
        prompt: '',
      },
      tagInput: '',
      draftFragment: {
        id: 0,
        title: '',
        content: '',
        tags: [],
        index_status_desc: '待索引',
        update_time_desc: '',
      },
      toolbars: [
        'bold',
        'italic',
        'strikeThrough',
        'title',
        'quote',
        'unorderedList',
        'orderedList',
        'task',
        'link',
        'image',
        'code',
        'codeRow',
        'table',
        'preview',
        'fullscreen'
      ],
    }
  },
  watch: {
    // fragment.id 变化时重置本地草稿。
    'fragment.id': {
      immediate: true,
      handler() {
        this.resetDraft()
      }
    },
    // savedFragment 变化后同步最新已保存内容。
    savedFragment: {
      deep: true,
      handler() {
        this.resetDraft()
      }
    }
  },
  computed: {
    // dirty 判断当前片段是否存在未保存改动。
    dirty() {
      return JSON.stringify(this.normalizeFragment(this.draftFragment)) !== JSON.stringify(this.normalizeFragment(this.savedFragment))
    }
  },
  methods: {
    // normalizeFragment 统一片段比较结构。
    normalizeFragment(fragment) {
      return {
        title: fragment.title || '',
        content: fragment.content || '',
        tags: Array.isArray(fragment.tags) ? [...fragment.tags].sort() : []
      }
    },
    // resetDraft 根据当前 props 重置本地草稿。
    resetDraft() {
      this.contentEditMode = false
      this.organizeDialogVisible = false
      this.organizeResult = {
        content: '',
        model: '',
        prompt: '',
      }
      this.draftFragment = {
        id: this.fragment.id,
        title: this.fragment.title || '',
        content: this.fragment.content || '',
        tags: Array.isArray(this.fragment.tags) ? [...this.fragment.tags] : [],
        index_status: this.fragment.index_status || 'pending',
        index_status_desc: this.fragment.index_status_desc || '待索引',
        update_time_desc: this.fragment.update_time_desc || '',
        create_time_desc: this.fragment.create_time_desc || '',
      }
    },
    // handleFormChange 在编辑后向父组件同步状态。
    handleFormChange() {
      this.$emit('change', JSON.parse(JSON.stringify(this.draftFragment)))
    },
    // setContentEditMode 切换正文查看/编辑模式。
    setContentEditMode(editMode) {
      this.contentEditMode = !!editMode
    },
    // appendTag 将输入框内容转换为标签。
    appendTag() {
      const rawTags = this.tagInput.split(/[，,]/).map(item => item.trim()).filter(Boolean)
      if (rawTags.length === 0) {
        this.tagInput = ''
        return
      }
      const tagMap = {}
      const nextTags = []
      ;(this.fragment.tags || []).forEach((tag) => {
        tagMap[tag.toLowerCase()] = true
        nextTags.push(tag)
      })
      rawTags.forEach((tag) => {
        const lowerTag = tag.toLowerCase()
        if (!tagMap[lowerTag]) {
          tagMap[lowerTag] = true
          nextTags.push(tag)
        }
      })
      this.draftFragment.tags = nextTags
      this.tagInput = ''
      this.handleFormChange()
    },
    // handleTagKeydown 在输入逗号时立即提交标签。
    handleTagKeydown(event) {
      if (event.key !== ',' && event.key !== '，') {
        return
      }
      event.preventDefault()
      this.appendTag()
    },
    // removeTag 删除一个已有标签。
    removeTag(tag) {
      this.draftFragment.tags = (this.draftFragment.tags || []).filter(item => item !== tag)
      this.handleFormChange()
    },
    // handleSave 保存当前片段。
    handleSave() {
      this.appendTag()
      this.saving = true
      MemoryFragmentApi.MemoryFragmentSave(
        this.draftFragment.id,
        this.draftFragment.title,
        this.draftFragment.content,
        this.draftFragment.tags || [],
        (response) => {
          this.saving = false
          if (response.ErrCode !== 0) {
            return
          }
          this.$emit('saved', response.Data)
        }
      )
    },
    // handleOrganize 调用 AI 对当前最新内容执行整理。
    handleOrganize() {
      this.appendTag()
      if (!this.draftFragment.content || this.draftFragment.content.trim() === '') {
        this.$helperNotify.error('当前片段内容不能为空')
        return
      }
      this.organizing = true
      MemoryFragmentApi.MemoryFragmentOrganize(
        this.draftFragment.id,
        this.draftFragment.title,
        this.draftFragment.content,
        this.draftFragment.tags || [],
        (response) => {
          this.organizing = false
          if (response.ErrCode !== 0 || !response.Data) {
            if (response.ErrMsg) {
              this.$helperNotify.error(response.ErrMsg)
            }
            return
          }
          this.organizeResult = {
            content: response.Data.content || '',
            model: response.Data.model || '',
            prompt: response.Data.prompt || '',
          }
          this.organizeDialogVisible = true
        }
      )
    },
    // applyOrganizeResult 确认后把整理结果写回当前片段并持久化保存。
    applyOrganizeResult() {
      if (!this.organizeResult.content || this.organizeResult.content.trim() === '') {
        this.$helperNotify.error('整理结果为空，无法写入')
        return
      }
      this.applyingOrganizeResult = true
      MemoryFragmentApi.MemoryFragmentSave(
        this.draftFragment.id,
        this.draftFragment.title,
        this.organizeResult.content,
        this.draftFragment.tags || [],
        (response) => {
          this.applyingOrganizeResult = false
          if (response.ErrCode !== 0 || !response.Data) {
            if (response.ErrMsg) {
              this.$helperNotify.error(response.ErrMsg)
            }
            return
          }
          this.organizeDialogVisible = false
          this.$emit('saved', response.Data)
          this.$helperNotify.success('AI整理结果已写入')
        }
      )
    },
    // handleDelete 删除当前片段。
    handleDelete() {
      MemoryFragmentApi.MemoryFragmentDelete(this.draftFragment.id, (response) => {
        if (response.ErrCode !== 0) {
          return
        }
        this.$emit('deleted', this.draftFragment.id)
      })
    }
  }
}
</script>

<style scoped>
.memory-editor {
  display: flex;
  flex-direction: column;
  gap: 14px;
  height: 100%;
}

.editor-toolbar {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  gap: 14px;
  padding: 16px;
  border: 1px solid #e8e8e0;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 4px 12px rgba(54, 74, 54, 0.05);
}

.toolbar-main {
  flex: 1;
  min-width: 0;
}

.title-input :deep(.el-input__wrapper) {
  min-height: 42px;
  border-radius: 10px;
}

.title-input :deep(.el-input__inner) {
  font-size: 18px;
  font-weight: 600;
  color: #33422f;
}

.toolbar-status {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 12px;
  color: #73806d;
  font-size: 12px;
}

.toolbar-time {
  font-size: 12px;
}

.toolbar-actions {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.tag-panel {
  display: flex;
  gap: 12px;
  align-items: center;
  padding: 14px 16px;
  border: 1px solid #e8e8e0;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 4px 12px rgba(54, 74, 54, 0.05);
}

.tag-list {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  min-height: 32px;
  flex: 1;
}

.tag-input {
  width: 260px;
  flex-shrink: 0;
}

.editor-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  border: 1px solid #e8e8e0;
  border-radius: 14px;
  overflow: hidden;
  background: #fff;
  box-shadow: 0 4px 12px rgba(54, 74, 54, 0.05);
}

.editor-body-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  padding: 12px 16px;
  border-bottom: 1px solid #ecece4;
  background: #f8faf5;
}

.editor-body-title {
  font-size: 14px;
  font-weight: 600;
  color: #455640;
}

.editor-body-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.editor-body-content {
  flex: 1;
  min-height: 0;
}

.editor-body-content :deep(.md-editor) {
  height: calc(100vh - 295px);
}

.preview-body {
  flex: 1;
  min-height: 0;
  overflow: auto;
  padding: 18px 22px;
  background: #fff;
}

 .preview-body :deep(.md-editor-preview) {
  font-size: 14px;
  color: #33422f;
}

.organize-dialog-layout {
  display: flex;
  flex-direction: column;
  gap: 14px;
  min-height: 70vh;
}

.organize-dialog-summary {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 14px 16px;
  border: 1px solid #e8e8e0;
  border-radius: 12px;
  background: #fafbf7;
}

.summary-item {
  display: flex;
  gap: 10px;
  align-items: flex-start;
}

.summary-label {
  width: 72px;
  flex-shrink: 0;
  color: #677560;
  font-size: 13px;
}

.summary-value {
  color: #34412f;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

@media (max-width: 1080px) {
  .editor-toolbar,
  .tag-panel {
    flex-direction: column;
    align-items: stretch;
  }

  .tag-input {
    width: 100%;
  }

  .editor-body-toolbar {
    flex-direction: column;
    align-items: stretch;
  }

  .editor-body-actions {
    justify-content: flex-end;
  }

  .summary-item {
    flex-direction: column;
    gap: 4px;
  }

  .editor-body-content :deep(.md-editor) {
    height: 60vh;
  }
}
</style>
