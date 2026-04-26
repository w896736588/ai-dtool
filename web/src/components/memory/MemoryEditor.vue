<template>
  <div class="memory-editor" @keydown.ctrl.s.prevent="handleSave">
    <div class="editor-body">
      <div
        :class="[
          'editor-body-toolbar',
          dirty ? 'editor-body-toolbar--dirty' : 'editor-body-toolbar--saved',
        ]"
      >
        <div class="editor-body-toolbar-main">
          <div class="editor-body-toolbar-top">
            <div class="editor-body-toolbar-left">
              <el-input
                ref="titleInput"
                v-model="draftFragment.title"
                class="title-input editor-toolbar-title-input"
                :class="{ 'title-input--search-hit': titleSearchMatchCount > 0 }"
                :placeholder="titlePlaceholderText"
              />
            </div>
            <div class="editor-body-toolbar-right">
              <div class="editor-body-actions">
                <GitActionButton
                  variant="info"
                  compact
                  :class="{ 'mode-button-active': !contentEditMode }"
                  @click="setContentEditMode(false)"
                >
                  {{ previewButtonText }}
                </GitActionButton>
                <GitActionButton
                  compact
                  :class="{ 'mode-button-active': contentEditMode }"
                  @click="setContentEditMode(true)"
                >
                  {{ editButtonText }}
                </GitActionButton>
                <GitActionButton
                  variant="info"
                  compact
                  @click="handleCopyContent"
                >
                  <template #icon>
                    <el-icon><CopyDocument /></el-icon>
                  </template>
                  {{ copyContentButtonText }}
                </GitActionButton>
                <GitActionButton compact :loading="saving" @click="handleSave">
                  <template #icon>
                    <el-icon><Check /></el-icon>
                  </template>
                  {{ saveButtonText }}
                </GitActionButton>
                <GitActionButton
                  variant="warning"
                  compact
                  :loading="organizing"
                  @click="handleOrganize"
                >
                  <template #icon>
                    <el-icon><MagicStick /></el-icon>
                  </template>
                  {{ organizeButtonText }}
                </GitActionButton>
                <el-dropdown
                  trigger="click"
                  class="editor-action-dropdown"
                  @command="handleToolbarActionCommand"
                >
                  <button class="editor-action-trigger" type="button" :aria-label="moreActionsText">
                    <el-icon><MoreFilled /></el-icon>
                  </button>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item :command="toolbarActionHistoryCommand">
                        {{ historyButtonText }}
                      </el-dropdown-item>
                      <el-dropdown-item divided :command="toolbarActionDeleteCommand">
                        {{ deleteButtonText }}
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </div>
            </div>
          </div>
            <div class="editor-search-row">
            <div class="editor-search-main">
              <el-input
                v-model="searchQuery"
                class="editor-search-input"
                clearable
                :placeholder="detailSearchPlaceholderText"
                @input="handleSearchInput"
                @keydown.enter.prevent="jumpToSearchMatch(1)"
                @keydown.shift.enter.prevent="jumpToSearchMatch(-1)"
                @keydown.esc.prevent="clearDetailSearch"
                @clear="clearDetailSearch"
              />
              <span class="editor-search-summary">
                {{ searchSummaryText }}
              </span>
            </div>
            <div class="editor-search-actions">
              <GitActionButton
                variant="info"
                :disabled="!hasSearchMatches"
                @click="jumpToSearchMatch(-1)"
              >
                上一个
              </GitActionButton>
              <GitActionButton
                variant="info"
                :disabled="!hasSearchMatches"
                @click="jumpToSearchMatch(1)"
              >
                下一个
              </GitActionButton>
              <GitActionButton
                variant="info"
                :disabled="!searchQuery"
                @click="clearDetailSearch"
              >
                清空
              </GitActionButton>
            </div>
          </div>
        </div>
      </div>

      <div v-if="contentEditMode" class="editor-body-content">
        <div class="editor-edit-layout">
          <div ref="editorScrollShell" class="editor-scroll-shell">
            <MdEditor
              ref="editorRef"
              v-model="draftFragment.content"
              preview-theme="github"
              :preview="false"
              :toolbars="toolbars"
              :style="editorContentStyle"
              :onUploadImg="handleUploadImg"
            />
          </div>
          <div class="editor-preview-shell">
            <div
              ref="previewBody"
              class="preview-renderer"
              @scroll="handlePreviewScroll"
              @click="handleFragmentLinkClick"
            >
              <MdPreview
                :model-value="draftFragment.content"
                preview-theme="github"
              />
            </div>
          </div>
        </div>
      </div>
      <div v-else class="preview-body" :class="{ 'preview-body--with-outline': hasOutline }">
        <div
          ref="previewBody"
          class="preview-renderer"
          @scroll="handlePreviewScroll"
          @click="handleFragmentLinkClick"
        >
          <MdPreview
            :model-value="draftFragment.content"
            preview-theme="github"
          />
        </div>
        <aside v-if="hasOutline" class="preview-outline">
          <div class="preview-outline-card">
            <div class="preview-outline-title">目录</div>
            <button
              v-for="item in outlineItems"
              :key="item.slug"
              type="button"
              class="preview-outline-item"
              :class="{
                active: activeOutlineSlug === item.slug,
                'preview-outline-item--child': item.level > 1,
                'preview-outline-item--grandchild': item.level > 2,
              }"
              @click="scrollToOutline(item.slug)"
            >
              {{ item.text }}
            </button>
          </div>
        </aside>
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
            <span class="summary-label">整理状态</span>
            <span class="summary-value">
              {{ organizing ? '整理中，正在实时输出...' : '整理完成' }}
            </span>
          </div>
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
            :disabled="organizing"
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
import { Check, CopyDocument, DocumentCopy, MagicStick, MoreFilled } from '@element-plus/icons-vue'
import DiffMarkdown from '@/components/base/diff_markwodn.vue'
import GitActionButton from '@/components/base/GitActionButton.vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'
import base from '@/utils/base'
import sseDistribute from '@/utils/base/sse_distribute'
const { buildMarkdownOutline } = require('@/utils/markdown_outline.cjs')
const {
  buildMemoryDetailSearchMatches,
  getNextMemoryDetailMatchIndex,
  normalizeActiveMatchIndex,
} = require('@/utils/memory_detail_search.cjs')

// STATUS_TAG_WARNING_TYPE 统一未保存状态标签类型，避免模板中散落硬编码。
const STATUS_TAG_WARNING_TYPE = 'warning'
// STATUS_TAG_SUCCESS_TYPE 统一已保存状态标签类型，避免模板中散落硬编码。
const STATUS_TAG_SUCCESS_TYPE = 'success'
// EMPTY_TIME_TEXT / 空白时间占位文案 / Fallback copy for empty dialog summary values.
const EMPTY_TIME_TEXT = '-'
// TITLE_PLACEHOLDER_TEXT 统一定义标题输入框占位文案。
const TITLE_PLACEHOLDER_TEXT = '输入片段标题'
// TAG_INPUT_PLACEHOLDER_TEXT 统一定义标签输入框占位文案。
const TAG_INPUT_PLACEHOLDER_TEXT = '输入标签后回车，可用逗号分隔'
// UNSAVED_STATUS_TEXT 统一定义未保存状态文案。
const UNSAVED_STATUS_TEXT = '未保存'
// SAVED_STATUS_TEXT 统一定义已保存状态文案。
const SAVED_STATUS_TEXT = '已保存'
// LAST_SAVE_LABEL_TEXT / 最后保存时间前缀 / Prefix for last saved time display.
const LAST_SAVE_LABEL_TEXT = '最后保存：'
// HISTORY_BUTTON_TEXT / 历史记录按钮文案 / History button copy.
const HISTORY_BUTTON_TEXT = '历史记录'
// DELETE_BUTTON_TEXT 统一定义删除按钮文案。
const DELETE_BUTTON_TEXT = '删除'
// SAVE_BUTTON_TEXT 统一定义保存按钮文案。
const SAVE_BUTTON_TEXT = '保存'
// CANCEL_BUTTON_TEXT 统一定义取消按钮文案。
const CANCEL_BUTTON_TEXT = '取消'
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
// DELETE_CONFIRM_NAME_EMPTY_TEXT 统一定义删除确认时的片段名称兜底文案。
const DELETE_CONFIRM_NAME_EMPTY_TEXT = '未命名片段'
// EMPTY_CONTENT_ERROR_TEXT 统一定义空内容提示，避免散落硬编码。
const EMPTY_CONTENT_ERROR_TEXT = '当前片段内容不能为空'
// EMPTY_ORGANIZE_RESULT_ERROR_TEXT 统一定义空整理结果提示。
const EMPTY_ORGANIZE_RESULT_ERROR_TEXT = '整理结果为空，无法写入'
// ORGANIZE_SUCCESS_TEXT 统一定义整理写回成功提示。
const ORGANIZE_SUCCESS_TEXT = 'AI整理结果已写入'
// COPY_PATH_BUTTON_TEXT 统一定义复制文件地址按钮文案。
const COPY_PATH_BUTTON_TEXT = '复制文件地址'
// COPY_CONTENT_BUTTON_TEXT 统一定义复制完整内容按钮文案。
const COPY_CONTENT_BUTTON_TEXT = '复制'
// INLINE_TAG_VISIBLE_LIMIT / 内容区右侧最多展示的标签数量 / Max visible inline tags beside content actions.
const INLINE_TAG_VISIBLE_LIMIT = 5
// TOOLBAR_ACTION_HISTORY_COMMAND / 工具栏下拉历史记录命令 / Dropdown command for history action.
const TOOLBAR_ACTION_HISTORY_COMMAND = 'history'
// TOOLBAR_ACTION_DELETE_COMMAND / 工具栏下拉删除命令 / Dropdown command for delete action.
const TOOLBAR_ACTION_DELETE_COMMAND = 'delete'
// MORE_ACTIONS_TEXT / 更多操作无障碍文案 / Accessible label for the action dropdown trigger.
const MORE_ACTIONS_TEXT = '更多操作'
// TAG_SUGGESTION_VISIBLE_LIMIT 控制已有标签候选区最多展示多少个标签。
const TAG_SUGGESTION_VISIBLE_LIMIT = 12
// TAG_SEPARATOR_PATTERN 统一定义标签分隔规则，兼容中英文逗号。
const TAG_SEPARATOR_PATTERN = /[，,]/
// FULL_WIDTH_COMMA_KEY 统一定义全角逗号按键值，避免键盘判断散落硬编码。
const FULL_WIDTH_COMMA_KEY = '，'
// DETAIL_SEARCH_PLACEHOLDER_TEXT 统一定义详情内搜索输入框占位文案。
const DETAIL_SEARCH_PLACEHOLDER_TEXT = '搜索标题和正文，Enter 下一个，Shift+Enter 上一个'
// SEARCH_EMPTY_SUMMARY_TEXT 无关键字时的搜索提示。
const SEARCH_EMPTY_SUMMARY_TEXT = '搜索范围：标题和正文'
// SEARCH_NO_RESULT_TEXT 无匹配时的提示文案。
const SEARCH_NO_RESULT_TEXT = '0 项匹配'

export default {
  name: 'MemoryEditor',
  components: {
      MdEditor,
      MdPreview,
      Check,
      CopyDocument,
      DocumentCopy,
      MagicStick,
    MoreFilled,
    DiffMarkdown,
    GitActionButton,
  },
  props: {
    fragment: {
      type: Object,
      required: true,
    },
    // draftFragment 统一监听草稿变化，确保父组件与左侧列表同步未保存状态。
    // draftFragment keeps parent state synced so the sidebar dirty color updates on the first edit.
    draftFragment: {
      deep: true,
      handler() {
        this.$nextTick(() => {
          this.handleFormChange()
        })
      },
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
      emptyTimeText: EMPTY_TIME_TEXT,
      titlePlaceholderText: TITLE_PLACEHOLDER_TEXT,
      tagInputPlaceholderText: TAG_INPUT_PLACEHOLDER_TEXT,
      unsavedStatusText: UNSAVED_STATUS_TEXT,
      savedStatusText: SAVED_STATUS_TEXT,
      lastSaveLabelText: LAST_SAVE_LABEL_TEXT,
      copyPathButtonText: COPY_PATH_BUTTON_TEXT,
      copyContentButtonText: COPY_CONTENT_BUTTON_TEXT,
      historyButtonText: HISTORY_BUTTON_TEXT,
      deleteButtonText: DELETE_BUTTON_TEXT,
      saveButtonText: SAVE_BUTTON_TEXT,
      cancelButtonText: CANCEL_BUTTON_TEXT,
      previewButtonText: PREVIEW_BUTTON_TEXT,
      editButtonText: EDIT_BUTTON_TEXT,
      organizeButtonText: ORGANIZE_BUTTON_TEXT,
      organizeDialogTitleText: ORGANIZE_DIALOG_TITLE_TEXT,
      organizeModelLabelText: ORGANIZE_MODEL_LABEL_TEXT,
      organizePromptLabelText: ORGANIZE_PROMPT_LABEL_TEXT,
      organizeDiffTitleText: ORGANIZE_DIFF_TITLE_TEXT,
      confirmWriteButtonText: CONFIRM_WRITE_BUTTON_TEXT,
      deleteConfirmTitleText: DELETE_CONFIRM_TITLE_TEXT,
      deleteConfirmNameEmptyText: DELETE_CONFIRM_NAME_EMPTY_TEXT,
      statusTagWarningType: STATUS_TAG_WARNING_TYPE,
      statusTagSuccessType: STATUS_TAG_SUCCESS_TYPE,
      toolbarActionHistoryCommand: TOOLBAR_ACTION_HISTORY_COMMAND,
      toolbarActionDeleteCommand: TOOLBAR_ACTION_DELETE_COMMAND,
      moreActionsText: MORE_ACTIONS_TEXT,
      detailSearchPlaceholderText: DETAIL_SEARCH_PLACEHOLDER_TEXT,
      searchQuery: '',
      currentSearchMatchIndex: -1,
      organizeResult: {
        content: '',
        model: '',
        prompt: '',
      },
      organizeSseDistributeId: '',
      tagInput: '',
      tagInputFocused: false,
      highlightedTagIndex: 0,
      activeOutlineSlug: '',
      editorScrollElement: null,
      previewScrollSyncRafId: 0,
      fragmentPathCache: {},
      draftFragment: {
        id: 0,
        title: '',
        content: '',
        file_path: '',
        tags: [],
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
  beforeUnmount() {
    this.unregisterOrganizeSse()
    this.detachEditorScrollListener()
    this.cancelPreviewScrollSync()
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
    // outlineItems 变化后刷新预览区标题锚点。
    outlineItems() {
      this.schedulePreviewOutlineRefresh()
    },
    // draftFragment.content 变化后重新扫描知识片段路径链接。
    'draftFragment.content'() {
      this.schedulePreviewOutlineRefresh()
    },
    // contentEditMode 切回查看模式时重新同步目录锚点与高亮。
    contentEditMode() {
      this.schedulePreviewOutlineRefresh()
    },
    // searchMatches 变化后同步当前命中项索引与高亮。
    searchMatches() {
      this.syncSearchMatchState()
    },
  },
  computed: {
    // dirty 判断当前片段是否存在未保存改动。
    dirty() {
      return JSON.stringify(this.normalizeFragment(this.draftFragment)) !== JSON.stringify(this.normalizeFragment(this.savedFragment))
    },
    // visibleEditorTags / 内容区右侧展示的标签列表 / Inline tags shown in the content toolbar.
    visibleEditorTags() {
      return (this.draftFragment.tags || []).slice(0, INLINE_TAG_VISIBLE_LIMIT)
    },
    // hiddenEditorTagCount / 未直接展示的标签数量 / Count of hidden tags beyond the inline limit.
    hiddenEditorTagCount() {
      return Math.max((this.draftFragment.tags || []).length - INLINE_TAG_VISIBLE_LIMIT, 0)
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
    // outlineItems 返回预览区目录项，当前只收录 h1-h3。
    outlineItems() {
      return buildMarkdownOutline(this.draftFragment.content)
    },
    // hasOutline 标记当前预览区是否展示目录。
    hasOutline() {
      return this.outlineItems.length > 0
    },
    // searchMatches 返回标题和正文的全部匹配项。
    searchMatches() {
      return buildMemoryDetailSearchMatches(
        this.draftFragment.title,
        this.draftFragment.content,
        this.searchQuery
      )
    },
    // hasSearchMatches 判断当前是否存在匹配项。
    hasSearchMatches() {
      return this.searchMatches.length > 0
    },
    // activeSearchMatch 返回当前激活的匹配项。
    activeSearchMatch() {
      if (!this.hasSearchMatches) {
        return null
      }
      return this.searchMatches[normalizeActiveMatchIndex(this.searchMatches, this.currentSearchMatchIndex)]
    },
    // titleSearchMatchCount 返回标题命中数量，便于给标题输入框提供视觉提示。
    titleSearchMatchCount() {
      return this.searchMatches.filter(item => item.field === 'title').length
    },
    // searchSummaryText 显示当前命中位置与总数。
    searchSummaryText() {
      if (!String(this.searchQuery || '').trim()) {
        return SEARCH_EMPTY_SUMMARY_TEXT
      }
      if (!this.hasSearchMatches) {
        return SEARCH_NO_RESULT_TEXT
      }
      return `${normalizeActiveMatchIndex(this.searchMatches, this.currentSearchMatchIndex) + 1} / ${this.searchMatches.length}`
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
        file_path: fragment.file_path || '',
        tags: Array.isArray(fragment.tags) ? [...fragment.tags].sort() : [],
      }
    },
    // resetDraft 根据当前 props 重置本地草稿。
    resetDraft(preserveEditMode) {
      this.detachEditorScrollListener()
      this.contentEditMode = preserveEditMode ? this.contentEditMode : false
      this.organizeDialogVisible = false
      this.activeOutlineSlug = ''
      this.closeTagSuggestionPanel()
      this.organizeResult = {
        content: '',
        model: '',
        prompt: '',
      }
      this.searchQuery = ''
      this.currentSearchMatchIndex = -1
      this.draftFragment = {
        id: this.fragment.id,
        title: this.fragment.title || '',
        content: this.fragment.content || '',
        file_path: this.fragment.file_path || '',
        tags: Array.isArray(this.fragment.tags) ? [...this.fragment.tags] : [],
        update_time_desc: this.fragment.update_time_desc || '',
        create_time_desc: this.fragment.create_time_desc || '',
      }
      this.schedulePreviewOutlineRefresh()
    },
    // handleFormChange 在编辑后向父组件同步状态。
    handleFormChange() {
      this.$emit('change', JSON.parse(JSON.stringify(this.draftFragment)))
    },
    // copyFilePath 复制当前片段实际文件路径。
    async copyFilePath() {
      if (!this.draftFragment.file_path || !navigator.clipboard) {
        return
      }
      try {
        await navigator.clipboard.writeText(this.draftFragment.file_path)
        this.$helperNotify.success(this.copyPathButtonText + '成功')
      } catch (error) {
        this.$helperNotify.error(this.copyPathButtonText + '失败')
      }
    },
    // handleCopyContent 复制当前片段的完整内容（标题+正文）。
    async handleCopyContent() {
      if (!navigator.clipboard) {
        return
      }
      const title = this.draftFragment.title || ''
      const content = this.draftFragment.content || ''
      const fullContent = title + '\n\n' + content
      try {
        await navigator.clipboard.writeText(fullContent)
        this.$helperNotify.success(this.copyContentButtonText + '成功')
      } catch (error) {
        this.$helperNotify.error(this.copyContentButtonText + '失败')
      }
    },
    // setContentEditMode 切换正文查看或编辑模式。
    setContentEditMode(editMode) {
      this.contentEditMode = !!editMode
    },
    // schedulePreviewOutlineRefresh 在预览 DOM 更新后重建标题锚点。
    schedulePreviewOutlineRefresh() {
      this.$nextTick(() => {
        window.setTimeout(() => {
          if (this.contentEditMode) {
            this.attachEditorScrollListener()
          }
          this.decoratePreviewHeadings()
          this.decoratePreviewSearchMatches()
          this.decorateFragmentPathLinks()
          if (this.contentEditMode) {
            this.syncPreviewScrollByEditor()
            this.syncActiveOutlineByEditorScroll()
            return
          }
          this.syncActiveOutlineByScroll()
        }, 0)
      })
    },
    // decorateFragmentPathLinks 扫描预览区文本，将知识片段文件路径替换为蓝色可点击的笔记标题。
    decorateFragmentPathLinks() {
      const previewBody = this.$refs.previewBody
      if (!previewBody) return
      const previewEl = previewBody.querySelector('.md-editor-preview')
      if (!previewEl) return

      const pathRegex = /[A-Za-z]:[\\/][^\s（）()]*[\\/]fragments[\\/]\d{4}[\\/]\d{4}-\d{2}[\\/][\w-]+\.md(?:（[^）]+）|\([^)]+\))?/g
      const paths = []
      const walker = document.createTreeWalker(previewEl, NodeFilter.SHOW_TEXT)
      const textNodes = []
      while (walker.nextNode()) {
        const node = walker.currentNode
        if (node.textContent && pathRegex.test(node.textContent)) {
          textNodes.push(node)
          pathRegex.lastIndex = 0
          let m
          while ((m = pathRegex.exec(node.textContent)) !== null) {
            const raw = m[0]
            const pathMatch = raw.match(/([A-Za-z]:[\\/][^\s（）()]*[\\/]fragments[\\/]\d{4}[\\/]\d{4}-\d{2}[\\/][\w-]+\.md)/)
            if (pathMatch && !paths.includes(pathMatch[1])) {
              paths.push(pathMatch[1])
            }
          }
        }
        pathRegex.lastIndex = 0
      }
      if (paths.length === 0) return

      const replaceNodes = (pathMap) => {
        const replaceRegex = /[A-Za-z]:[\\/][^\s（）()]*[\\/]fragments[\\/]\d{4}[\\/]\d{4}-\d{2}[\\/][\w-]+\.md(?:（[^）]+）|\([^)]+\))?/g
        for (const node of textNodes) {
          if (!node.parentNode) continue
          const text = node.textContent
          replaceRegex.lastIndex = 0
          if (!replaceRegex.test(text)) continue
          replaceRegex.lastIndex = 0
          const frag = document.createDocumentFragment()
          let lastIndex = 0
          let match
          while ((match = replaceRegex.exec(text)) !== null) {
            const raw = match[0]
            const pathMatch = raw.match(/([A-Za-z]:[\\/][^\s（）()]*[\\/]fragments[\\/]\d{4}[\\/]\d{4}-\d{2}[\\/][\w-]+\.md)/)
            if (!pathMatch) continue
            const info = pathMap[pathMatch[1]]
            if (match.index > lastIndex) {
              frag.appendChild(document.createTextNode(text.slice(lastIndex, match.index)))
            }
            const span = document.createElement('span')
            span.className = 'fragment-path-link'
            span.setAttribute('data-fragment-id', info ? info.id : '')
            span.setAttribute('data-fragment-path', pathMatch[1])
            span.textContent = info ? info.title : raw
            frag.appendChild(span)
            lastIndex = match.index + raw.length
          }
          if (lastIndex < text.length) {
            frag.appendChild(document.createTextNode(text.slice(lastIndex)))
          }
          node.parentNode.replaceChild(frag, node)
        }
      }

      // 所有路径均已缓存 → 同步替换，避免异步回调时 DOM 已被重新渲染
      const uncachedPaths = paths.filter(p => !this.fragmentPathCache[p])
      if (uncachedPaths.length === 0) {
        replaceNodes(this.fragmentPathCache)
        return
      }

      MemoryFragmentApi.MemoryFragmentBatchInfoByPaths(paths, (res) => {
        if (!res || res.ErrCode !== 0 || !Array.isArray(res.Data)) return
        for (const item of res.Data) {
          this.fragmentPathCache[item.file_path] = item
        }
        replaceNodes(this.fragmentPathCache)
      })
    },
    // handleFragmentLinkClick 处理预览区片段链接点击，新窗口打开对应知识片段。
    handleFragmentLinkClick(event) {
      const link = event.target.closest('.fragment-path-link')
      if (!link) return
      const fragmentId = link.getAttribute('data-fragment-id')
      if (!fragmentId) {
        this.$helperNotify.warning('未找到对应的知识片段')
        return
      }
      const routeInfo = this.$router.resolve({
        path: '/MemoryFragment',
        query: {
          fragment_id: String(fragmentId),
          hide_menu: '1',
        },
      })
      window.open(routeInfo.href, '_blank')
    },
    // handleSearchInput 在输入关键字时切到第一项匹配并刷新高亮。
    handleSearchInput() {
      this.currentSearchMatchIndex = this.hasSearchMatches ? 0 : -1
      this.$nextTick(() => {
        this.applyActiveSearchMatch(true)
      })
    },
    // syncSearchMatchState 在片段内容或关键字变化后矫正当前命中索引。
    syncSearchMatchState() {
      this.currentSearchMatchIndex = normalizeActiveMatchIndex(this.searchMatches, this.currentSearchMatchIndex)
      this.$nextTick(() => {
        this.applyActiveSearchMatch(false)
      })
    },
    // clearDetailSearch 清空详情内搜索状态。
    clearDetailSearch() {
      this.searchQuery = ''
      this.currentSearchMatchIndex = -1
      this.$nextTick(() => {
        this.clearPreviewSearchMarks()
      })
    },
    // jumpToSearchMatch 切换到上一项或下一项搜索结果。
    jumpToSearchMatch(step) {
      if (!this.hasSearchMatches) {
        return
      }
      this.currentSearchMatchIndex = getNextMemoryDetailMatchIndex(
        this.searchMatches,
        this.currentSearchMatchIndex,
        step
      )
      this.$nextTick(() => {
        this.applyActiveSearchMatch(true)
      })
    },
    // applyActiveSearchMatch 把当前命中同步到标题输入框或正文预览区。
    applyActiveSearchMatch(shouldScroll) {
      if (!String(this.searchQuery || '').trim()) {
        this.clearPreviewSearchMarks()
        return
      }
      this.decoratePreviewSearchMatches()
      if (!this.activeSearchMatch) {
        return
      }
      if (this.activeSearchMatch.field === 'title') {
        this.focusTitleSearchMatch()
        return
      }
      this.scrollToActiveSearchMark(shouldScroll)
    },
    // clearPreviewSearchMarks 移除预览区内已有的搜索高亮。
    clearPreviewSearchMarks() {
      const previewBody = this.$refs.previewBody
      if (!previewBody) {
        return
      }
      const markList = previewBody.querySelectorAll('mark.memory-search-mark')
      markList.forEach((mark) => {
        const parent = mark.parentNode
        if (!parent) {
          return
        }
        parent.replaceChild(document.createTextNode(mark.textContent || ''), mark)
        parent.normalize()
      })
    },
    // decoratePreviewSearchMatches 给预览区正文添加关键词高亮。
    decoratePreviewSearchMatches() {
      const previewBody = this.$refs.previewBody
      if (!previewBody) {
        return
      }
      this.clearPreviewSearchMarks()
      const normalizedQuery = String(this.searchQuery || '').trim()
      if (!normalizedQuery) {
        return
      }
      const previewRoot = previewBody.querySelector('.md-editor-preview')
      if (!previewRoot) {
        return
      }
      const lowerQuery = normalizedQuery.toLocaleLowerCase()
      const walker = document.createTreeWalker(previewRoot, NodeFilter.SHOW_TEXT, {
        acceptNode(node) {
          const parentTagName = node.parentElement ? node.parentElement.tagName : ''
          if (!node.nodeValue || !node.nodeValue.trim()) {
            return NodeFilter.FILTER_REJECT
          }
          if (parentTagName === 'SCRIPT' || parentTagName === 'STYLE' || parentTagName === 'MARK') {
            return NodeFilter.FILTER_REJECT
          }
          return NodeFilter.FILTER_ACCEPT
        },
      })
      const textNodeList = []
      while (walker.nextNode()) {
        textNodeList.push(walker.currentNode)
      }
      textNodeList.forEach((textNode) => {
        const originalText = textNode.nodeValue || ''
        const lowerText = originalText.toLocaleLowerCase()
        let fromIndex = 0
        let matchIndex = lowerText.indexOf(lowerQuery, fromIndex)
        if (matchIndex === -1) {
          return
        }
        const fragment = document.createDocumentFragment()
        while (matchIndex !== -1) {
          if (matchIndex > fromIndex) {
            fragment.appendChild(document.createTextNode(originalText.slice(fromIndex, matchIndex)))
          }
          const mark = document.createElement('mark')
          mark.className = 'memory-search-mark'
          mark.textContent = originalText.slice(matchIndex, matchIndex + normalizedQuery.length)
          fragment.appendChild(mark)
          fromIndex = matchIndex + normalizedQuery.length
          matchIndex = lowerText.indexOf(lowerQuery, fromIndex)
        }
        if (fromIndex < originalText.length) {
          fragment.appendChild(document.createTextNode(originalText.slice(fromIndex)))
        }
        if (textNode.parentNode) {
          textNode.parentNode.replaceChild(fragment, textNode)
        }
      })
      this.syncActivePreviewSearchMark()
    },
    // syncActivePreviewSearchMark 高亮当前激活的正文命中项。
    syncActivePreviewSearchMark() {
      const previewBody = this.$refs.previewBody
      if (!previewBody) {
        return
      }
      const markList = Array.from(previewBody.querySelectorAll('mark.memory-search-mark'))
      markList.forEach(mark => mark.classList.remove('memory-search-mark--active'))
      if (!this.activeSearchMatch || this.activeSearchMatch.field !== 'content') {
        return
      }
      const activeContentIndex = this.searchMatches
        .slice(0, normalizeActiveMatchIndex(this.searchMatches, this.currentSearchMatchIndex) + 1)
        .filter(item => item.field === 'content')
        .length - 1
      if (activeContentIndex < 0 || activeContentIndex >= markList.length) {
        return
      }
      markList[activeContentIndex].classList.add('memory-search-mark--active')
    },
    // scrollToActiveSearchMark 把当前正文命中项滚动到可视区域。
    scrollToActiveSearchMark(shouldScroll) {
      const previewBody = this.$refs.previewBody
      if (!previewBody) {
        return
      }
      this.syncActivePreviewSearchMark()
      if (!shouldScroll) {
        return
      }
      const activeMark = previewBody.querySelector('mark.memory-search-mark--active')
      if (!activeMark) {
        return
      }
      previewBody.scrollTo({
        top: Math.max(activeMark.offsetTop - 40, 0),
        behavior: 'smooth',
      })
    },
    // focusTitleSearchMatch 在当前命中位于标题时聚焦并选中标题输入框对应范围。
    focusTitleSearchMatch() {
      const titleInput = this.$refs.titleInput
      const activeMatch = this.activeSearchMatch
      if (!titleInput || !activeMatch || activeMatch.field !== 'title') {
        return
      }
      if (typeof titleInput.focus === 'function') {
        titleInput.focus()
      }
      const inputEl = titleInput.input
      if (!inputEl || typeof inputEl.setSelectionRange !== 'function') {
        return
      }
      inputEl.setSelectionRange(activeMatch.index, activeMatch.end)
    },
    // buildOutlineHeadingDomId 返回渲染后标题使用的 DOM id。
    buildOutlineHeadingDomId(slug) {
      return `memory-fragment-heading-${slug}`
    },
    // getEditorView 返回 md-editor-v3 暴露的 CodeMirror EditorView。
    getEditorView() {
      const editorRef = this.$refs.editorRef
      if (!editorRef || typeof editorRef.getEditorView !== 'function') {
        return null
      }
      return editorRef.getEditorView()
    },
    // attachEditorScrollListener 在编辑模式下监听 CodeMirror 滚动容器，驱动目录高亮。
    attachEditorScrollListener() {
      const editorView = this.getEditorView()
      const scrollElement = editorView && editorView.scrollDOM ? editorView.scrollDOM : null
      if (!scrollElement) {
        return
      }
      if (this.editorScrollElement === scrollElement) {
        return
      }
      this.detachEditorScrollListener()
      scrollElement.addEventListener('scroll', this.handleEditorScroll, { passive: true })
      this.editorScrollElement = scrollElement
    },
    // detachEditorScrollListener 卸载旧的编辑器滚动监听，避免切换片段后重复绑定。
    detachEditorScrollListener() {
      if (!this.editorScrollElement) {
        return
      }
      this.editorScrollElement.removeEventListener('scroll', this.handleEditorScroll)
      this.editorScrollElement = null
    },
    // handleEditorScroll 在编辑模式滚动 textarea 时同步当前目录高亮。
    handleEditorScroll() {
      this.syncPreviewScrollByEditor()
      this.syncActiveOutlineByEditorScroll()
    },
    // syncPreviewScrollByEditor 在编辑模式下按滚动比例同步右侧预览区。
    syncPreviewScrollByEditor() {
      const editorView = this.getEditorView()
      const previewBody = this.$refs.previewBody
      if (!this.contentEditMode || !editorView || !editorView.scrollDOM || !previewBody) {
        return
      }
      const editorScrollElement = editorView.scrollDOM
      const editorScrollableHeight = Math.max(editorScrollElement.scrollHeight - editorScrollElement.clientHeight, 0)
      const previewScrollableHeight = Math.max(previewBody.scrollHeight - previewBody.clientHeight, 0)
      if (previewScrollableHeight <= 0) {
        previewBody.scrollTop = 0
        return
      }
      const scrollRatio = editorScrollableHeight <= 0
        ? 0
        : Math.min(Math.max(editorScrollElement.scrollTop / editorScrollableHeight, 0), 1)
      const nextPreviewScrollTop = previewScrollableHeight * scrollRatio
      this.cancelPreviewScrollSync()
      this.previewScrollSyncRafId = window.requestAnimationFrame(() => {
        previewBody.scrollTop = nextPreviewScrollTop
        this.previewScrollSyncRafId = 0
      })
    },
    // cancelPreviewScrollSync 取消尚未执行的预览滚动同步帧，避免重复排队。
    cancelPreviewScrollSync() {
      if (!this.previewScrollSyncRafId) {
        return
      }
      window.cancelAnimationFrame(this.previewScrollSyncRafId)
      this.previewScrollSyncRafId = 0
    },
    // scrollEditorToOutline 通过 EditorView 直接滚动到对应标题行。
    scrollEditorToOutline(outlineItem) {
      const editorView = this.getEditorView()
      if (!editorView || !outlineItem) {
        return
      }
      const lineNumber = Math.max(Number(outlineItem.lineNumber || 1), 1)
      const targetLine = editorView.state.doc.line(Math.min(lineNumber, editorView.state.doc.lines))
      const targetTop = editorView.lineBlockAt(targetLine.from).top
      editorView.dispatch({
        selection: { anchor: targetLine.from },
        scrollIntoView: true,
      })
      if (editorView.scrollDOM && typeof editorView.scrollDOM.scrollTo === 'function') {
        editorView.scrollDOM.scrollTo({
          top: Math.max(targetTop - 12, 0),
          behavior: 'smooth',
        })
      }
      editorView.focus()
      this.$nextTick(() => {
        this.syncPreviewScrollByEditor()
        this.syncActiveOutlineByEditorScroll()
      })
    },
    // decoratePreviewHeadings 给预览区标题写入稳定锚点，供目录点击跳转。
    decoratePreviewHeadings() {
      const previewBody = this.$refs.previewBody
      if (!previewBody) {
        return
      }
      const headingList = previewBody.querySelectorAll('.md-editor-preview h1, .md-editor-preview h2, .md-editor-preview h3')
      headingList.forEach((heading, index) => {
        const outlineItem = this.outlineItems[index]
        if (!outlineItem) {
          return
        }
        heading.id = this.buildOutlineHeadingDomId(outlineItem.slug)
        heading.dataset.outlineSlug = outlineItem.slug
      })
    },
    // scrollToOutline 点击目录后滚动到对应标题位置。
    scrollToOutline(slug) {
      const outlineItem = this.outlineItems.find(item => item.slug === slug)
      if (this.contentEditMode) {
        this.activeOutlineSlug = slug
        this.scrollEditorToOutline(outlineItem)
        return
      }
      const previewBody = this.$refs.previewBody
      if (!previewBody) {
        return
      }
      const heading = previewBody.querySelector(`#${this.buildOutlineHeadingDomId(slug)}`)
      if (!heading) {
        return
      }
      previewBody.scrollTo({
        top: Math.max(heading.offsetTop - 12, 0),
        behavior: 'smooth',
      })
      this.activeOutlineSlug = slug
    },
    // handlePreviewScroll 在预览滚动时同步当前目录高亮。
    handlePreviewScroll() {
      this.syncActiveOutlineByScroll()
    },
    // syncActiveOutlineByScroll 根据正文滚动位置高亮最接近的目录项。
    syncActiveOutlineByScroll() {
      const previewBody = this.$refs.previewBody
      if (!previewBody || this.outlineItems.length === 0) {
        this.activeOutlineSlug = ''
        return
      }
      const headingList = Array.from(
        previewBody.querySelectorAll('.md-editor-preview h1, .md-editor-preview h2, .md-editor-preview h3')
      )
      if (headingList.length === 0) {
        this.activeOutlineSlug = this.outlineItems[0].slug
        return
      }
      const scrollTop = previewBody.scrollTop
      const matchedHeading = headingList.reduce((currentHeading, heading) => {
        if (heading.offsetTop - 28 <= scrollTop) {
          return heading
        }
        return currentHeading
      }, headingList[0])
      this.activeOutlineSlug = matchedHeading.dataset.outlineSlug || this.outlineItems[0].slug
    },
    // syncActiveOutlineByEditorScroll 根据编辑器滚动位置推导当前所在标题。
    syncActiveOutlineByEditorScroll() {
      const editorView = this.getEditorView()
      if (!editorView || !editorView.scrollDOM || this.outlineItems.length === 0) {
        if (this.contentEditMode) {
          this.activeOutlineSlug = this.outlineItems[0] ? this.outlineItems[0].slug : ''
        }
        return
      }
      const topLineBlock = editorView.lineBlockAtHeight(editorView.scrollDOM.scrollTop)
      const currentTopLine = editorView.state.doc.lineAt(topLineBlock.from).number
      const matchedItem = this.outlineItems.reduce((currentItem, item) => {
        if (Number(item.lineNumber || 0) <= currentTopLine + 1) {
          return item
        }
        return currentItem
      }, this.outlineItems[0])
      this.activeOutlineSlug = matchedItem ? matchedItem.slug : ''
    },
    // handleToolbarActionCommand / 统一处理右侧下拉操作 / Dispatch commands from the toolbar dropdown.
    handleToolbarActionCommand(command) {
      if (command === TOOLBAR_ACTION_HISTORY_COMMAND) {
        this.$emit('show-history', this.draftFragment.id)
        return
      }
      // 删除属于危险操作 / Delete is destructive, so keep an extra confirmation step.
      if (command === TOOLBAR_ACTION_DELETE_COMMAND) {
        this.confirmDeleteFromToolbar()
      }
    },
    // confirmDeleteFromToolbar / 下拉删除前二次确认 / Ask for confirmation before deleting from dropdown.
    confirmDeleteFromToolbar() {
      this.$confirm(this.buildDeleteConfirmMessage(), this.deleteButtonText, {
        confirmButtonText: this.deleteButtonText,
        cancelButtonText: this.cancelButtonText,
        type: 'warning',
        dangerouslyUseHTMLString: true,
        center: true,
      })
        .then(() => {
          this.handleDelete()
        })
        .catch(() => {})
    },
    // buildDeleteConfirmMessage 生成删除确认弹窗 HTML，突出当前要删除的片段标题。
    // Build the delete confirmation HTML and highlight the fragment title in the center.
    buildDeleteConfirmMessage() {
      const fragmentTitle = this.escapeDeleteConfirmText(this.draftFragment.title || this.deleteConfirmNameEmptyText)
      return `
        <div class="memory-delete-confirm">
          <div class="memory-delete-confirm__desc">${this.deleteConfirmTitleText}</div>
          <div class="memory-delete-confirm__name">${fragmentTitle}</div>
        </div>
      `
    },
    // escapeDeleteConfirmText 对删除确认中的动态标题做 HTML 转义，避免特殊字符破坏弹窗结构。
    // Escape dynamic title text in the delete confirmation dialog to keep the HTML safe.
    escapeDeleteConfirmText(text) {
      return String(text || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;')
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
    // handleSave 保存当前片段。
    handleSave() {
      if (this.saving) {
        return
      }
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
    // triggerSave 供父组件通过 ref 统一触发保存，保证快捷键和点击入口一致。
    triggerSave() {
      this.handleSave()
    },
    // handleOrganize 调用 AI 对当前最新内容执行整理。
    handleOrganize() {
      if (this.organizing) {
        return
      }
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
        '',
        (response) => {
          this.organizing = false
          if (response.ErrCode !== 0 || !response.Data) {
            if (response.ErrMsg) {
              this.$helperNotify.error(response.ErrMsg)
            }
            return
          }
          this.$helperNotify.success('AI 整理任务已加入异步任务列表')
        }
      )
    },
    // unregisterOrganizeSse 清理本次 AI 整理绑定的 SSE 回调，避免重复拼接旧流。
    unregisterOrganizeSse() {
      if (!this.organizeSseDistributeId) {
        return
      }
      sseDistribute.UnRegisterReceive(this.organizeSseDistributeId)
      this.organizeSseDistributeId = ''
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
    // handleUploadImg 处理编辑器中的图片上传（粘贴、拖拽、工具栏上传）。
    handleUploadImg(files, callback) {
      const uploadPromises = Array.from(files).map((file) => {
        return new Promise((resolve) => {
          MemoryFragmentApi.MemoryFragmentImageUpload(file, (response) => {
            if (response.ErrCode === 0 && response.Data && response.Data.url) {
              resolve(base.GetApiHost() + response.Data.url)
            } else {
              resolve('')
            }
          })
        })
      })
      Promise.all(uploadPromises).then((urls) => {
        callback(urls.filter(Boolean))
      })
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
  --memory-toolbar-text-primary: #2f3c2b;
  display: flex;
  flex-direction: column;
  height: 100%;
}

:global(.memory-delete-confirm) {
  text-align: center;
}

:global(.memory-delete-confirm__desc) {
  color: #5f6758;
  line-height: 1.6;
}

:global(.memory-delete-confirm__name) {
  margin-top: 10px;
  color: #c23b32;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.6;
  word-break: break-word;
}

.title-input :deep(.el-input__wrapper) {
  min-height: 30px;
  height: 30px;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: inset 0 0 0 1px rgba(205, 214, 198, 0.9);
}

.title-input--search-hit :deep(.el-input__wrapper) {
  box-shadow: inset 0 0 0 1px rgba(230, 171, 46, 0.92);
}

.title-input :deep(.el-input__inner) {
  font-size: 14px;
  font-weight: 400;
  color: var(--memory-toolbar-text-primary);
}

.tag-input {
  width: 220px;
  flex-shrink: 0;
}

.tag-input-wrap {
  position: relative;
  width: 220px;
  flex-shrink: 0;
}

.tag-input :deep(.el-input__wrapper) {
  min-height: 34px;
  height: 34px;
  border-radius: 10px;
  background: rgba(255, 255, 255, 0.9);
  box-shadow: inset 0 0 0 1px rgba(211, 220, 204, 0.92);
}

.tag-input-compact :deep(.el-input__wrapper) {
  min-height: 30px;
  height: 30px;
  border-radius: 8px;
}

.tag-input-compact :deep(.el-input__inner) {
  font-size: 13px;
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
  flex-direction: column;
  gap: 10px;
  padding: 12px 16px;
  border-bottom: 1px solid rgba(226, 232, 216, 0.9);
  background: #f8faf5;
  transition: background-color 0.2s ease, border-color 0.2s ease, box-shadow 0.2s ease;
}

.editor-body-toolbar--dirty {
  background: linear-gradient(180deg, #fdf7ef 0%, #f7eee1 100%);
  border-bottom-color: rgba(224, 191, 146, 0.78);
  box-shadow: inset 0 -1px 0 rgba(241, 214, 179, 0.52);
}

.editor-body-toolbar--saved {
  background: linear-gradient(180deg, #f6faf2 0%, #edf5e7 100%);
  border-bottom-color: rgba(196, 217, 186, 0.9);
  box-shadow: inset 0 -1px 0 rgba(210, 229, 201, 0.54);
}

.editor-body-toolbar-main {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
  width: 100%;
}

.editor-body-toolbar-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
}

.editor-body-toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  flex: 1;
  flex-wrap: wrap;
}

.editor-toolbar-title-input {
  width: 100%;
  flex: 1 1 200px;
}

.editor-body-toolbar-right {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  min-width: 0;
  flex-shrink: 0;
}

.editor-inline-tags :deep(.el-tag) {
  display: inline-flex;
  align-items: center;
  height: 28px;
  border-radius: 999px;
  padding-inline: 10px;
}

.editor-search-row {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  flex-wrap: wrap;
}

.editor-search-main {
  display: flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1 1 420px;
  flex-wrap: wrap;
}

.editor-search-input {
  width: min(360px, 100%);
}

.editor-search-input :deep(.el-input__wrapper) {
  min-height: 32px;
  height: 32px;
  border-radius: 9px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: inset 0 0 0 1px rgba(211, 220, 204, 0.92);
}

.editor-search-summary {
  color: #64715e;
  font-size: 12px;
  line-height: 1.4;
  white-space: nowrap;
}

.editor-search-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  flex: 0 0 auto;
}

.editor-title-inline-meta {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  min-width: 0;
  flex: 1 1 260px;
  flex-wrap: wrap;
}

.editor-save-time {
  color: #7b8875;
  font-size: 12px;
  line-height: 1.4;
  white-space: nowrap;
  flex-shrink: 0;
}

.editor-title-inline-meta :deep(.el-tag) {
  flex-shrink: 0;
}

.editor-body-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  flex-shrink: 0;
}

.editor-action-trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border: 1px solid rgba(210, 219, 203, 0.95);
  border-radius: 10px;
  background: #ffffff;
  color: #5d6e57;
  cursor: pointer;
  transition: border-color 0.2s ease, color 0.2s ease, background-color 0.2s ease;
}

.editor-action-trigger:hover {
  border-color: #b9c9b1;
  color: #3d5237;
  background: #f7fbf2;
}

.editor-action-trigger:focus-visible {
  outline: 2px solid rgba(95, 125, 86, 0.32);
  outline-offset: 2px;
}

.editor-body-content {
  flex: 1;
  min-height: 0;
  overflow: hidden;
}

.editor-edit-layout {
  height: 100%;
  min-height: 0;
  display: flex;
  gap: 18px;
  padding: 18px 22px;
  background: #fff;
}

.editor-scroll-shell {
  flex: 1;
  height: 100%;
  min-height: 0;
  overflow: auto;
}

.editor-preview-shell {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  padding: 16px 18px;
  border: 1px solid var(--memory-toolbar-border-color);
  border-radius: 14px;
}

.editor-body-content :deep(.md-editor) {
  height: 100%;
}

.editor-body-content :deep(.md-editor-content) {
  min-height: 0;
}

.editor-body-content :deep(.md-editor-input-wrapper) {
  overflow: auto;
}

.preview-body {
  flex: 1;
  min-height: 0;
  display: flex;
  gap: 18px;
  padding: 18px 22px;
  background: #fff;
}

.preview-body--with-outline {
  align-items: stretch;
}

.preview-renderer {
  flex: 1;
  height: 100%;
  min-width: 0;
  min-height: 0;
  overflow: auto;
  scrollbar-gutter: stable;
  scrollbar-width: thin;
  scrollbar-color: #9fb39a #edf3e8;
  padding-right: 6px;
}

.preview-body :deep(.md-editor-preview) {
  font-size: 14px;
  color: #33422f;
}

.preview-body :deep(mark.memory-search-mark) {
  padding: 0 2px;
  border-radius: 4px;
  background: #fff1a8;
  color: inherit;
}

.preview-body :deep(mark.memory-search-mark--active) {
  background: #f5bf46;
  box-shadow: 0 0 0 1px rgba(212, 144, 27, 0.28);
}

.preview-body :deep(.fragment-path-link),
.editor-preview-shell :deep(.fragment-path-link) {
  color: #409eff;
  cursor: pointer;
  text-decoration-style: dashed;
  font-weight: bold;
}
.preview-body :deep(.fragment-path-link:hover),
.editor-preview-shell :deep(.fragment-path-link:hover) {
  color: #66b1ff;
  text-decoration-style: solid;
}

.preview-outline {
  width: 220px;
  flex: 0 0 220px;
  min-width: 0;
}

.preview-outline-card {
  position: sticky;
  top: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: 100%;
  overflow: auto;
  scrollbar-gutter: stable;
  scrollbar-width: thin;
  scrollbar-color: #9fb39a #edf3e8;
  padding: 14px 12px;
  border: 1px solid var(--memory-toolbar-border-color);
  border-radius: 12px;
  background: linear-gradient(180deg, #fbfcf8 0%, #f6f8f2 100%);
}

.preview-renderer::-webkit-scrollbar,
.preview-outline-card::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

.preview-renderer::-webkit-scrollbar-track,
.preview-outline-card::-webkit-scrollbar-track {
  background: #edf3e8;
  border-radius: 999px;
}

.preview-renderer::-webkit-scrollbar-thumb,
.preview-outline-card::-webkit-scrollbar-thumb {
  background: #9fb39a;
  border: 2px solid #edf3e8;
  border-radius: 999px;
}

.preview-renderer::-webkit-scrollbar-thumb:hover,
.preview-outline-card::-webkit-scrollbar-thumb:hover {
  background: #869c82;
}

.preview-outline-title {
  margin-bottom: 4px;
  font-size: 13px;
  font-weight: 700;
  color: #50604a;
}

.preview-outline-item {
  width: 100%;
  padding: 8px 10px;
  border: none;
  border-radius: 10px;
  background: transparent;
  color: #556551;
  cursor: pointer;
  font-size: 13px;
  line-height: 1.5;
  text-align: left;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.preview-outline-item:hover {
  background: #eef4e8;
  color: #30412d;
}

.preview-outline-item.active {
  background: #e6f1de;
  color: #264224;
  font-weight: 600;
}

.preview-outline-item--child {
  padding-left: 22px;
}

.preview-outline-item--grandchild {
  padding-left: 34px;
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
  .tag-input-wrap {
    width: 100%;
  }

  .editor-title-inline-meta,
  .editor-search-main,
  .editor-search-actions {
    width: 100%;
  }

  .editor-body-toolbar-main,
  .editor-body-toolbar-top,
  .editor-body-toolbar,
  .editor-search-row {
    flex-direction: column;
    align-items: stretch;
  }

  .editor-body-toolbar-right {
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

  .preview-body {
    flex-direction: column;
  }

  .editor-edit-layout {
    flex-direction: column;
  }

  .preview-outline {
    width: 100%;
    flex-basis: auto;
  }

  .preview-outline-card {
    position: static;
    max-height: 220px;
  }
}
</style>
