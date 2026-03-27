<template>
  <div class="memory-editor" @keydown.ctrl.s.prevent="handleSave">
    <div class="editor-toolbar">
      <div class="toolbar-header-row">
        <div class="toolbar-main">
          <el-input
            v-model="draftFragment.title"
            class="title-input"
            :placeholder="titlePlaceholderText"
            @input="handleFormChange"
          />
          <div class="toolbar-status-row">
            <div class="toolbar-status">
              <el-tag size="small" :type="dirty ? statusTagWarningType : statusTagSuccessType" effect="light">
                {{ dirty ? unsavedStatusText : savedStatusText }}
              </el-tag>
              <el-tag size="small" effect="plain">
                {{ draftFragment.index_status_desc || defaultIndexStatusText }}
              </el-tag>
            </div>
            <span class="toolbar-time">{{ updateTimePrefixText }}{{ draftFragment.update_time_desc || emptyTimeText }}</span>
          </div>
        </div>
        <div class="toolbar-actions">
          <GitActionButton variant="info" @click="$emit('show-history', draftFragment.id)">
            <template #icon>
              <el-icon><Clock /></el-icon>
            </template>
            {{ historyButtonText }}
          </GitActionButton>
          <el-popconfirm
            :title="deleteConfirmTitleText"
            :confirm-button-text="deleteButtonText"
            :cancel-button-text="cancelButtonText"
            @confirm="handleDelete"
          >
            <template #reference>
              <GitActionButton variant="danger">
                <template #icon>
                  <el-icon><Delete /></el-icon>
                </template>
                {{ deleteButtonText }}
              </GitActionButton>
            </template>
          </el-popconfirm>
          <GitActionButton :loading="saving" @click="handleSave">
            <template #icon>
              <el-icon><Check /></el-icon>
            </template>
            {{ saveButtonText }}
          </GitActionButton>
        </div>
      </div>

      <div class="toolbar-meta-row">
        <div class="tag-meta-layout">
          <div class="tag-panel">
            <div class="tag-panel-head">
              <span class="tag-panel-label">{{ tagLabelText }}</span>
              <pl-button
                v-if="showTagListToggle"
                class="tag-panel-toggle"
                link
                @click="toggleTagListExpanded"
              >
                {{ tagListToggleText }}
              </pl-button>
            </div>
            <div
              class="tag-list"
              :class="{ collapsed: !tagListExpanded }"
              :style="tagListStyle"
            >
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
          </div>
        </div>
        <div class="tag-input-wrap">
          <el-input
            v-model="tagInput"
            class="tag-input"
            :placeholder="tagInputPlaceholderText"
            @focus="handleTagInputFocus"
            @input="handleTagInputChange"
            @keydown.enter.prevent="handleTagEnter"
            @keydown.down.prevent="moveTagSuggestion(1)"
            @keydown.up.prevent="moveTagSuggestion(-1)"
            @keydown.esc.prevent="closeTagSuggestionPanel"
            @keydown="handleTagKeydown"
            @blur="handleTagInputBlur"
          />
          <div v-if="showTagSuggestionPanel" class="tag-suggestion-dropdown">
            <button
              v-for="(tag, index) in filteredAvailableTags"
              :key="tag"
              class="tag-suggestion-option"
              :class="{ active: index === highlightedTagIndex }"
              @mousedown.prevent="selectExistingTag(tag)"
            >
              {{ tag }}
            </button>
            <div v-if="showTagCreateHint" class="tag-suggestion-empty">
              回车创建标签 “{{ normalizedTagInput }}”
            </div>
          </div>
        </div>
      </div>
    </div>

    <div class="editor-body">
      <div class="editor-body-toolbar">
        <div class="editor-body-title">{{ contentTitleText }}</div>
        <div class="editor-body-actions">
          <GitActionButton
            variant="info"
            :class="{ 'mode-button-active': !contentEditMode }"
            @click="setContentEditMode(false)"
          >
            {{ previewButtonText }}
          </GitActionButton>
          <GitActionButton
            :class="{ 'mode-button-active': contentEditMode }"
            @click="setContentEditMode(true)"
          >
            {{ editButtonText }}
          </GitActionButton>
          <GitActionButton
            variant="warning"
            :loading="organizing"
            @click="handleOrganize"
          >
            <template #icon>
              <el-icon><MagicStick /></el-icon>
            </template>
            {{ organizeButtonText }}
          </GitActionButton>
        </div>
      </div>

      <div v-if="contentEditMode" class="editor-body-content">
        <div class="editor-scroll-shell">
          <MdEditor
            v-model="draftFragment.content"
            preview-theme="github"
            :toolbars="toolbars"
            :style="editorContentStyle"
            @onChange="handleFormChange"
            @onBlur="handleFormChange"
          />
        </div>
      </div>
      <div v-else class="preview-body">
        <MdPreview
          :model-value="draftFragment.content"
          preview-theme="github"
        />
      </div>
    </div>

    <el-dialog
      v-model="organizeDialogVisible"
      :title="organizeDialogTitleText"
      width="84%"
      top="5vh"
      class="memory-organize-dialog"
    >
      <div class="organize-dialog-layout">
        <div class="organize-dialog-summary">
          <div class="summary-item">
            <span class="summary-label">{{ organizeModelLabelText }}</span>
            <span class="summary-value">{{ organizeResult.model || emptyTimeText }}</span>
          </div>
          <div class="summary-item">
            <span class="summary-label">{{ organizePromptLabelText }}</span>
            <span class="summary-value">{{ organizeResult.prompt || emptyTimeText }}</span>
          </div>
        </div>
        <diff-markdown
          :old-text="draftFragment.content || ''"
          :new-text="organizeResult.content || ''"
          :title="organizeDiffTitleText"
        />
      </div>
      <template #footer>
        <div class="dialog-footer">
          <GitActionButton variant="info" @click="organizeDialogVisible = false">{{ cancelButtonText }}</GitActionButton>
          <GitActionButton
            :loading="applyingOrganizeResult"
            @click="applyOrganizeResult"
          >
            <template #icon>
              <el-icon><Check /></el-icon>
            </template>
            {{ confirmWriteButtonText }}
          </GitActionButton>
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
import GitActionButton from '@/components/base/GitActionButton.vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'

// STATUS_TAG_WARNING_TYPE 统一未保存状态标签类型，避免模板中散落硬编码。
const STATUS_TAG_WARNING_TYPE = 'warning'
// STATUS_TAG_SUCCESS_TYPE 统一已保存状态标签类型，避免模板中散落硬编码。
const STATUS_TAG_SUCCESS_TYPE = 'success'
// DEFAULT_INDEX_STATUS_TEXT 统一定义默认索引状态文案，避免散落硬编码。
const DEFAULT_INDEX_STATUS_TEXT = '待索引'
// EMPTY_TIME_TEXT 统一定义时间或空数据占位文案。
const EMPTY_TIME_TEXT = '-'
// TITLE_PLACEHOLDER_TEXT 统一定义标题输入框占位文案。
const TITLE_PLACEHOLDER_TEXT = '输入片段标题'
// TAG_INPUT_PLACEHOLDER_TEXT 统一定义标签输入框占位文案。
const TAG_INPUT_PLACEHOLDER_TEXT = '输入标签后回车，可用逗号分隔'
// UNSAVED_STATUS_TEXT 统一定义未保存状态文案。
const UNSAVED_STATUS_TEXT = '未保存'
// SAVED_STATUS_TEXT 统一定义已保存状态文案。
const SAVED_STATUS_TEXT = '已保存'
// UPDATE_TIME_PREFIX_TEXT 统一定义更新时间前缀文案。
const UPDATE_TIME_PREFIX_TEXT = '更新于 '
// HISTORY_BUTTON_TEXT 统一定义历史记录按钮文案。
const HISTORY_BUTTON_TEXT = '历史记录'
// DELETE_BUTTON_TEXT 统一定义删除按钮文案。
const DELETE_BUTTON_TEXT = '删除'
// SAVE_BUTTON_TEXT 统一定义保存按钮文案。
const SAVE_BUTTON_TEXT = '保存'
// CANCEL_BUTTON_TEXT 统一定义取消按钮文案。
const CANCEL_BUTTON_TEXT = '取消'
// TAG_LABEL_TEXT 统一定义标签区域标题文案。
const TAG_LABEL_TEXT = '标签'
// CONTENT_TITLE_TEXT 统一定义正文区标题文案。
const CONTENT_TITLE_TEXT = '内容'
// PREVIEW_BUTTON_TEXT 统一定义预览按钮文案。
const PREVIEW_BUTTON_TEXT = '查看'
// EDIT_BUTTON_TEXT 统一定义编辑按钮文案。
const EDIT_BUTTON_TEXT = '编辑'
// ORGANIZE_BUTTON_TEXT 统一定义 AI 整理按钮文案。
const ORGANIZE_BUTTON_TEXT = 'AI整理'
// ORGANIZE_DIALOG_TITLE_TEXT 统一定义 AI 整理结果弹窗标题。
const ORGANIZE_DIALOG_TITLE_TEXT = 'AI整理结果对比'
// ORGANIZE_MODEL_LABEL_TEXT 统一定义整理模型字段标签。
const ORGANIZE_MODEL_LABEL_TEXT = '整理模型'
// ORGANIZE_PROMPT_LABEL_TEXT 统一定义整理提示词字段标签。
const ORGANIZE_PROMPT_LABEL_TEXT = '整理提示词'
// ORGANIZE_DIFF_TITLE_TEXT 统一定义正文差异标题。
const ORGANIZE_DIFF_TITLE_TEXT = '正文差异'
// CONFIRM_WRITE_BUTTON_TEXT 统一定义确认写入按钮文案。
const CONFIRM_WRITE_BUTTON_TEXT = '确认写入'
// DELETE_CONFIRM_TITLE_TEXT 统一定义删除确认文案。
const DELETE_CONFIRM_TITLE_TEXT = '确定删除这个片段吗？'
// EMPTY_CONTENT_ERROR_TEXT 统一定义空内容提示，避免散落硬编码。
const EMPTY_CONTENT_ERROR_TEXT = '当前片段内容不能为空'
// EMPTY_ORGANIZE_RESULT_ERROR_TEXT 统一定义空整理结果提示。
const EMPTY_ORGANIZE_RESULT_ERROR_TEXT = '整理结果为空，无法写入'
// ORGANIZE_SUCCESS_TEXT 统一定义整理写回成功提示。
const ORGANIZE_SUCCESS_TEXT = 'AI整理结果已写入'
// EDITOR_TAG_LIST_COLLAPSED_MAX_HEIGHT 控制标签区域折叠时的最大高度。
const EDITOR_TAG_LIST_COLLAPSED_MAX_HEIGHT = 38
// EDITOR_TAG_TOGGLE_MIN_COUNT 控制显示展开入口的最小标签数量。
const EDITOR_TAG_TOGGLE_MIN_COUNT = 6
// TAG_SUGGESTION_VISIBLE_LIMIT 控制已有标签候选区最多展示多少个标签。
const TAG_SUGGESTION_VISIBLE_LIMIT = 12
// TAG_SEPARATOR_PATTERN 统一定义标签分隔规则，兼容中英文逗号。
const TAG_SEPARATOR_PATTERN = /[，,]/
// FULL_WIDTH_COMMA_KEY 统一定义全角逗号按键值，避免键盘判断散落硬编码。
const FULL_WIDTH_COMMA_KEY = '，'

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
    GitActionButton,
  },
  props: {
    fragment: {
      type: Object,
      required: true,
    },
    savedFragment: {
      type: Object,
      required: true,
    },
    availableTags: {
      type: Array,
      default: () => [],
    },
  },
  emits: ['change', 'saved', 'deleted', 'show-history'],
  data() {
    return {
      saving: false,
      organizing: false,
      applyingOrganizeResult: false,
      contentEditMode: false,
      organizeDialogVisible: false,
      defaultIndexStatusText: DEFAULT_INDEX_STATUS_TEXT,
      emptyTimeText: EMPTY_TIME_TEXT,
      titlePlaceholderText: TITLE_PLACEHOLDER_TEXT,
      tagInputPlaceholderText: TAG_INPUT_PLACEHOLDER_TEXT,
      unsavedStatusText: UNSAVED_STATUS_TEXT,
      savedStatusText: SAVED_STATUS_TEXT,
      updateTimePrefixText: UPDATE_TIME_PREFIX_TEXT,
      historyButtonText: HISTORY_BUTTON_TEXT,
      deleteButtonText: DELETE_BUTTON_TEXT,
      saveButtonText: SAVE_BUTTON_TEXT,
      cancelButtonText: CANCEL_BUTTON_TEXT,
      tagLabelText: TAG_LABEL_TEXT,
      contentTitleText: CONTENT_TITLE_TEXT,
      previewButtonText: PREVIEW_BUTTON_TEXT,
      editButtonText: EDIT_BUTTON_TEXT,
      organizeButtonText: ORGANIZE_BUTTON_TEXT,
      organizeDialogTitleText: ORGANIZE_DIALOG_TITLE_TEXT,
      organizeModelLabelText: ORGANIZE_MODEL_LABEL_TEXT,
      organizePromptLabelText: ORGANIZE_PROMPT_LABEL_TEXT,
      organizeDiffTitleText: ORGANIZE_DIFF_TITLE_TEXT,
      confirmWriteButtonText: CONFIRM_WRITE_BUTTON_TEXT,
      deleteConfirmTitleText: DELETE_CONFIRM_TITLE_TEXT,
      statusTagWarningType: STATUS_TAG_WARNING_TYPE,
      statusTagSuccessType: STATUS_TAG_SUCCESS_TYPE,
      organizeResult: {
        content: '',
        model: '',
        prompt: '',
      },
      tagInput: '',
      tagInputFocused: false,
      highlightedTagIndex: 0,
      tagListExpanded: false,
      draftFragment: {
        id: 0,
        title: '',
        content: '',
        tags: [],
        index_status_desc: DEFAULT_INDEX_STATUS_TEXT,
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
        'fullscreen',
      ],
    }
  },
  watch: {
    // fragment.id 变化时重置本地草稿，避免旧数据残留。
    'fragment.id': {
      immediate: true,
      handler() {
        this.resetDraft(false)
      },
    },
    // savedFragment 变化后同步最新已保存内容。
    savedFragment: {
      deep: true,
      handler() {
        this.resetDraft(true)
      },
    },
  },
  computed: {
    // dirty 判断当前片段是否存在未保存改动。
    dirty() {
      return JSON.stringify(this.normalizeFragment(this.draftFragment)) !== JSON.stringify(this.normalizeFragment(this.savedFragment))
    },
    // showTagListToggle 判断标签数是否需要显示展开入口。
    showTagListToggle() {
      return (this.draftFragment.tags || []).length >= EDITOR_TAG_TOGGLE_MIN_COUNT
    },
    // tagListStyle 统一生成标签区域在不同状态下的高度样式。
    tagListStyle() {
      if (this.tagListExpanded) {
        return {}
      }
      return {
        maxHeight: `${EDITOR_TAG_LIST_COLLAPSED_MAX_HEIGHT}px`,
      }
    },
    // tagListToggleText 返回标签区展开或收起文案。
    tagListToggleText() {
      const tagCount = (this.draftFragment.tags || []).length
      return this.tagListExpanded ? '收起标签' : `展开标签（${tagCount}）`
    },
    // availableTagCandidates 返回当前可快速选择的已有标签。
    availableTagCandidates() {
      const selectedTagMap = {}
      ;(this.draftFragment.tags || []).forEach((tag) => {
        selectedTagMap[String(tag).toLowerCase()] = true
      })
      return (this.availableTags || [])
        .map((item) => String(item || '').trim())
        .filter((tag) => tag !== '' && !selectedTagMap[tag.toLowerCase()])
    },
    // normalizedTagInput 返回去空格后的标签输入。
    normalizedTagInput() {
      return String(this.tagInput || '').trim()
    },
    // filteredAvailableTags 返回按输入内容做包含过滤后的候选标签。
    filteredAvailableTags() {
      const normalizedInput = this.normalizedTagInput.toLowerCase()
      const candidateList = this.availableTagCandidates
      if (normalizedInput === '') {
        return candidateList.slice(0, TAG_SUGGESTION_VISIBLE_LIMIT)
      }
      return candidateList
        .filter(tag => tag.toLowerCase().includes(normalizedInput))
        .slice(0, TAG_SUGGESTION_VISIBLE_LIMIT)
    },
    // showTagSuggestionPanel 控制标签候选面板显示。
    showTagSuggestionPanel() {
      return this.tagInputFocused && (this.filteredAvailableTags.length > 0 || this.showTagCreateHint)
    },
    // showTagCreateHint 判断当前是否展示创建标签提示。
    showTagCreateHint() {
      return this.normalizedTagInput !== '' && this.filteredAvailableTags.length === 0
    },
    // editorContentStyle 统一让 Markdown 编辑器撑满弹性容器。
    editorContentStyle() {
      return {
        height: '100%',
      }
    },
  },
  methods: {
    // normalizeFragment 统一片段比较结构，避免无关字段导致误判脏数据。
    normalizeFragment(fragment) {
      return {
        title: fragment.title || '',
        content: fragment.content || '',
        tags: Array.isArray(fragment.tags) ? [...fragment.tags].sort() : [],
      }
    },
    // resetDraft 根据当前 props 重置本地草稿。
    resetDraft(preserveEditMode) {
      this.contentEditMode = preserveEditMode ? this.contentEditMode : false
      this.organizeDialogVisible = false
      this.closeTagSuggestionPanel()
      this.tagListExpanded = false
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
        index_status_desc: this.fragment.index_status_desc || DEFAULT_INDEX_STATUS_TEXT,
        update_time_desc: this.fragment.update_time_desc || '',
        create_time_desc: this.fragment.create_time_desc || '',
      }
    },
    // handleFormChange 在编辑后向父组件同步状态。
    handleFormChange() {
      this.$emit('change', JSON.parse(JSON.stringify(this.draftFragment)))
    },
    // setContentEditMode 切换正文查看或编辑模式。
    setContentEditMode(editMode) {
      this.contentEditMode = !!editMode
    },
    // appendTag 将输入框内容转换为标签并去重。
    appendTag() {
      const rawTags = this.tagInput.split(TAG_SEPARATOR_PATTERN).map(item => item.trim()).filter(Boolean)
      if (rawTags.length === 0) {
        this.tagInput = ''
        return
      }
      const tagMap = {}
      const nextTags = []
      ;(this.draftFragment.tags || []).forEach((tag) => {
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
      this.highlightedTagIndex = 0
      this.handleFormChange()
    },
    // handleTagInputFocus 打开标签候选面板。
    handleTagInputFocus() {
      this.tagInputFocused = true
      this.highlightedTagIndex = 0
    },
    // handleTagInputChange 输入时重置候选高亮。
    handleTagInputChange() {
      this.highlightedTagIndex = 0
    },
    // handleTagEnter 优先选择候选标签，否则创建新标签。
    handleTagEnter() {
      if (this.filteredAvailableTags.length > 0) {
        const targetIndex = Math.min(this.highlightedTagIndex, this.filteredAvailableTags.length - 1)
        this.selectExistingTag(this.filteredAvailableTags[targetIndex])
        return
      }
      this.appendTag()
    },
    // handleTagKeydown 在输入逗号时立即提交标签。
    handleTagKeydown(event) {
      if (event.key !== ',' && event.key !== FULL_WIDTH_COMMA_KEY) {
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
    // selectExistingTag 快速选择一个已有标签。
    selectExistingTag(tag) {
      this.draftFragment.tags = [...(this.draftFragment.tags || []), tag]
      this.tagInput = ''
      this.highlightedTagIndex = 0
      this.handleFormChange()
    },
    // moveTagSuggestion 切换标签候选高亮项。
    moveTagSuggestion(step) {
      if (this.filteredAvailableTags.length === 0) {
        return
      }
      const lastIndex = this.filteredAvailableTags.length - 1
      const nextIndex = this.highlightedTagIndex + step
      if (nextIndex < 0) {
        this.highlightedTagIndex = lastIndex
        return
      }
      if (nextIndex > lastIndex) {
        this.highlightedTagIndex = 0
        return
      }
      this.highlightedTagIndex = nextIndex
    },
    // closeTagSuggestionPanel 关闭标签候选面板。
    closeTagSuggestionPanel() {
      this.tagInputFocused = false
      this.highlightedTagIndex = 0
    },
    // handleTagInputBlur 失焦时延后收起面板，避免点击候选被提前打断。
    handleTagInputBlur() {
      window.setTimeout(() => {
        this.closeTagSuggestionPanel()
      }, 120)
    },
    // toggleTagListExpanded 切换标签区域的展开状态。
    toggleTagListExpanded() {
      this.tagListExpanded = !this.tagListExpanded
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
        this.$helperNotify.error(EMPTY_CONTENT_ERROR_TEXT)
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
        this.$helperNotify.error(EMPTY_ORGANIZE_RESULT_ERROR_TEXT)
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
          this.$helperNotify.success(ORGANIZE_SUCCESS_TEXT)
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
    },
  },
}
</script>

<style scoped>
.memory-editor {
  --memory-toolbar-border-color: #e2e8d8;
  --memory-toolbar-panel-background: linear-gradient(180deg, #fdfefb 0%, #f7faf4 100%);
  --memory-toolbar-text-primary: #2f3c2b;
  --memory-toolbar-text-secondary: #6d7b67;
  --memory-toolbar-text-muted: #8a9486;
  --memory-toolbar-shadow: 0 10px 24px rgba(54, 74, 54, 0.06);
  display: flex;
  flex-direction: column;
  gap: 14px;
  height: 100%;
}

.editor-toolbar {
  display: flex;
  flex-direction: column;
  gap: 12px;
  padding: 16px 18px;
  border: 1px solid var(--memory-toolbar-border-color);
  border-radius: 16px;
  background: var(--memory-toolbar-panel-background);
  box-shadow: var(--memory-toolbar-shadow);
}

.toolbar-header-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.toolbar-main {
  flex: 1;
  min-width: 0;
}

.toolbar-status-row {
  display: flex;
  align-items: center;
  justify-content: flex-start;
  gap: 10px;
  margin-top: 10px;
}

.toolbar-meta-row {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
  min-width: 0;
  padding-top: 12px;
  border-top: 1px solid rgba(164, 178, 157, 0.28);
}

.tag-meta-layout {
  display: flex;
  flex: 1;
  min-width: 0;
}

.title-input :deep(.el-input__wrapper) {
  min-height: 34px;
  height: 34px;
  border-radius: 12px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: inset 0 0 0 1px rgba(205, 214, 198, 0.9);
}

.title-input :deep(.el-input__inner) {
  font-size: 18px;
  font-weight: 600;
  color: var(--memory-toolbar-text-primary);
}

.toolbar-status {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
  color: var(--memory-toolbar-text-secondary);
  font-size: 12px;
}

.toolbar-time {
  margin-left: auto;
  font-size: 12px;
  line-height: 1.4;
  color: var(--memory-toolbar-text-muted);
  text-align: right;
  white-space: nowrap;
}

.toolbar-actions {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 8px;
  flex-wrap: wrap;
  flex-shrink: 0;
  min-height: 34px;
}

.tag-panel {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1;
  min-width: 0;
}

.tag-panel-head {
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
  min-height: 34px;
}

.tag-panel-label {
  color: var(--memory-toolbar-text-secondary);
  font-size: 13px;
  font-weight: 600;
}

.tag-panel-toggle {
  padding: 0;
  border: none;
  background: transparent;
  color: #5f7d56;
  font-size: 12px;
  cursor: pointer;
}

.tag-panel-toggle:hover {
  color: #45603e;
  text-decoration: underline;
}

.tag-list {
  display: flex;
  gap: 8px;
  flex: 1;
  flex-wrap: wrap;
  min-height: 34px;
  min-width: 0;
  overflow: hidden;
  align-content: center;
  align-items: center;
  transition: max-height 0.2s ease;
}

.tag-list.collapsed {
  mask-image: linear-gradient(180deg, #000 0%, #000 70%, rgba(0, 0, 0, 0) 100%);
}

.tag-list :deep(.el-tag) {
  display: inline-flex;
  align-items: center;
  height: 28px;
  border-radius: 999px;
  padding-inline: 10px;
}

.tag-input {
  width: 320px;
  flex-shrink: 0;
}

.tag-input-wrap {
  position: relative;
  width: 320px;
  flex-shrink: 0;
}

.tag-input :deep(.el-input__wrapper) {
  min-height: 34px;
  height: 34px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: inset 0 0 0 1px rgba(211, 220, 204, 0.92);
}

.tag-suggestion-dropdown {
  position: absolute;
  top: calc(100% + 6px);
  left: 0;
  right: 0;
  z-index: 20;
  display: flex;
  flex-direction: column;
  gap: 2px;
  max-height: 220px;
  padding: 6px;
  border: 1px solid #dbe7d4;
  border-radius: 10px;
  background: #ffffff;
  box-shadow: 0 10px 24px rgba(54, 74, 54, 0.12);
  overflow: auto;
}

.tag-suggestion-option {
  width: 100%;
  min-height: 30px;
  padding: 6px 10px;
  border: none;
  border-radius: 8px;
  background: transparent;
  color: #3f5140;
  cursor: pointer;
  text-align: left;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.tag-suggestion-option:hover,
.tag-suggestion-option.active {
  background: #edf6e7;
  color: #35512f;
}

.tag-suggestion-empty {
  padding: 8px 10px;
  color: #6d7b67;
  font-size: 12px;
  line-height: 1.5;
}

.mode-button-active {
  position: relative;
}

.mode-button-active::after {
  content: '';
  position: absolute;
  inset: -1px;
  border-radius: 8px;
  box-shadow: inset 0 0 0 1px rgba(79, 128, 79, 0.2);
  pointer-events: none;
}

.editor-body {
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
  border: 1px solid var(--memory-toolbar-border-color);
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
  border-bottom: 1px solid rgba(226, 232, 216, 0.9);
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
  overflow: hidden;
}

.editor-scroll-shell {
  height: 100%;
  min-height: 0;
  overflow: auto;
}

.editor-body-content :deep(.md-editor) {
  height: 100%;
}

.editor-body-content :deep(.md-editor-content) {
  min-height: 0;
}

.editor-body-content :deep(.md-editor-input-wrapper),
.editor-body-content :deep(.md-editor-preview-wrapper) {
  overflow: auto;
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
  border: 1px solid var(--memory-toolbar-border-color);
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

.dialog-footer {
  display: flex;
  justify-content: flex-end;
  gap: 10px;
}

@media (max-width: 1080px) {
  .toolbar-header-row {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-status-row {
    flex-direction: column;
    align-items: stretch;
  }

  .toolbar-meta-row {
    flex-direction: column;
    align-items: stretch;
  }

  .tag-panel {
    flex-direction: column;
  }

  .tag-input-wrap {
    width: 100%;
  }

  .toolbar-actions {
    justify-content: flex-start;
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
