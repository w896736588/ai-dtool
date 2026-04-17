<template>
  <div class="memory-page">
    <aside v-if="memoryConfigured && !sidebarCollapsed" class="memory-sidebar">
      <div class="sidebar-header">
        <span v-show="!sidebarCollapsed" class="sidebar-title">知识片段</span>
        <div v-show="!sidebarCollapsed" class="sidebar-header-actions">
          <pl-button type="primary" plain size="small" @click="createFragment">
            <el-icon><Plus /></el-icon>
            新建
          </pl-button>
          <GitActionButton variant="warning" compact @click="openTrashTab">
            <template #icon>
              <el-icon><Delete /></el-icon>
            </template>
            回收站
          </GitActionButton>
          <pl-button plain size="small" @click="openSettingsDialog">
            设置
          </pl-button>
        </div>
      </div>

      <div v-show="!sidebarCollapsed" class="search-card sidebar-search-card">
        <div class="search-row">
          <el-input
            v-model="searchQuery"
            clearable
            :placeholder="searchPlaceholder"
            @keydown.enter.prevent="submitSearch"
            @clear="handleSearchClear"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
        </div>
        <div class="search-mode-row">
          <el-radio-group v-model="searchMode" size="small">
            <el-radio-button label="keyword">全文检索</el-radio-button>
            <el-radio-button label="semantic">智能检索</el-radio-button>
          </el-radio-group>
        </div>
        <div class="search-actions">
          <pl-button type="primary" @click="submitSearch">
            <el-icon><Search /></el-icon>
            搜索
          </pl-button>
          <pl-button plain @click="clearFilter">清空条件</pl-button>
        </div>
      </div>

      <el-scrollbar v-show="!sidebarCollapsed" class="sidebar-scroll">
        <button
          v-for="item in fragmentList"
          :key="sidebarItemKey(item)"
          :class="[
            'sidebar-item',
            fragmentFreshnessClass(item),
            {
              active: activeFragmentId === item.id,
              'sidebar-item--dirty': isFragmentDirty(item.id),
              'save-success': !!saveFeedbackMap[item.id],
            }
          ]"
          @click="openFragment(item.id)"
        >
          <div class="sidebar-item-main">
            <div class="sidebar-item-title-row">
              <div class="sidebar-item-title">{{ item.title }}</div>
              <span v-if="activeFragmentId === item.id" class="sidebar-item-active-badge">当前</span>
            </div>
          </div>
          <div v-if="item.file_path || item.update_time_desc" class="sidebar-item-meta">
            <span
              v-if="item.file_path"
              class="sidebar-item-copy"
              role="button"
              tabindex="0"
              @click.stop="copyFragmentPath(item.file_path)"
              @keydown.enter.stop.prevent="copyFragmentPath(item.file_path)"
              @keydown.space.stop.prevent="copyFragmentPath(item.file_path)"
            >
              复制文件地址
            </span>
            <div class="sidebar-item-time">{{ item.update_time_desc || '-' }}</div>
          </div>
          <div v-if="saveFeedbackMap[item.id]" class="sidebar-item-check" aria-hidden="true">
            <el-icon><Check /></el-icon>
          </div>
        </button>
      </el-scrollbar>

      <div v-if="memoryGitRepoEnabled && !sidebarCollapsed" class="sidebar-footer">
        <div class="sidebar-footer-row">
          <span class="sidebar-footer-label">{{ pushStatusLabel }}</span>
          <span class="sidebar-footer-value">{{ pushStatusDesc }}</span>
        </div>
        <div v-if="lastPushError" class="sidebar-footer-row sidebar-footer-error">
          <span class="sidebar-footer-label">失败原因</span>
          <span class="sidebar-footer-value">{{ lastPushError }}</span>
        </div>
      </div>
    </aside>

    <button v-if="memoryConfigured" class="sidebar-collapse-btn" :title="sidebarCollapsed ? '展开列表' : '收起列表'" @click="toggleSidebar">
      <el-icon :size="12"><component :is="sidebarCollapsed ? 'DArrowRight' : 'DArrowLeft'" /></el-icon>
    </button>

    <section class="memory-main">
      <div class="workspace-card">
        <el-tabs
          v-model="activeTab"
          type="card"
          closable
          class="memory-tabs"
          @tab-remove="closeTab"
          @tab-click="handleTabChange"
        >
          <el-tab-pane
            v-if="searchTabVisible"
            name="search"
          >
            <template #label>
              <span class="tab-label">{{ searchTabLabel }}</span>
            </template>
            <div v-loading="searchLoading" class="search-result-panel">
              <div class="search-result-toolbar">
                <div class="search-result-summary">
                  <div class="search-result-title">搜索结果</div>
                  <div class="search-result-desc">
                    <span v-if="submittedSearchQuery">关键词：{{ submittedSearchQuery }}</span>
                    <span>模式：{{ submittedSearchMode === 'semantic' ? '智能检索' : '全文检索' }}</span>
                    <span>命中：{{ searchResults.length }}</span>
                  </div>
                </div>
              </div>

              <el-empty
                v-if="!searchLoading && searchResults.length === 0"
                description="没有匹配的文档"
              />

              <div v-else class="search-result-list">
                <button
                  v-for="item in searchResults"
                  :key="item.id"
                  class="search-result-item"
                  @click="openFragment(item.id)"
                >
                  <div class="search-result-item-head">
                    <div class="search-result-item-title">{{ item.title || '未命名片段' }}</div>
                    <div class="search-result-item-time">{{ item.update_time_desc || '-' }}</div>
                  </div>
                  <div class="search-result-item-snippet">
                    <div
                      v-for="(snippet, index) in getSearchSnippetList(item)"
                      :key="item.id + '-snippet-' + index"
                      class="search-result-snippet-line"
                      v-html="highlightSearchKeywords(snippet)"
                    ></div>
                    <div
                      v-if="getSearchSnippetMoreCount(item) > 0"
                      class="search-result-snippet-more"
                    >
                      还有 {{ getSearchSnippetMoreCount(item) }} 个匹配片段...
                    </div>
                  </div>
                </button>
              </div>
            </div>
          </el-tab-pane>

          <el-tab-pane
            v-if="trashTabVisible"
            name="trash"
          >
            <template #label>
              <span class="tab-label">{{ trashTabLabel }}</span>
            </template>
            <div v-loading="trashLoading" class="search-result-panel">
              <div class="search-result-toolbar">
                <div class="search-result-summary">
                  <div class="search-result-title">回收站</div>
                  <div class="search-result-desc">
                    <span>已删除片段：{{ trashList.length }}</span>
                    <span>支持恢复和彻底删除</span>
                  </div>
                </div>
              </div>

              <el-empty
                v-if="!trashLoading && trashList.length === 0"
                description="回收站为空"
              />

              <div v-else class="search-result-list">
                <div
                  v-for="item in trashList"
                  :key="item.id"
                  class="trash-result-item"
                >
                  <div class="search-result-item-head">
                    <div class="search-result-item-title">{{ item.title || '未命名片段' }}</div>
                    <div class="search-result-item-time">{{ item.update_time_desc || '-' }}</div>
                  </div>
                  <div class="trash-result-actions">
                    <GitActionButton variant="info" compact @click="handleFragmentRestore(item.id)">
                      恢复
                    </GitActionButton>
                    <el-popconfirm
                      popper-class="memory-fragment-delete-popconfirm"
                      confirm-button-text="彻底删除"
                      cancel-button-text="取消"
                      @confirm="handleFragmentHardDelete(item.id)"
                    >
                      <template #default>
                        <div class="delete-popconfirm-content">
                          <div class="delete-popconfirm-desc">确定彻底删除这个片段吗？</div>
                          <div class="delete-popconfirm-name">{{ item.title || '未命名片段' }}</div>
                        </div>
                      </template>
                      <template #reference>
                        <GitActionButton variant="danger" compact>彻底删除</GitActionButton>
                      </template>
                    </el-popconfirm>
                  </div>
                </div>
              </div>
            </div>
          </el-tab-pane>

          <el-tab-pane
            v-for="tab in fragmentTabs"
            :key="tab.name"
            :name="tab.name"
          >
            <template #label>
              <span class="tab-label">
                {{ tab.fragment.title || '未命名片段' }}<span v-if="tab.dirty"> *</span>
              </span>
            </template>
            <MemoryEditor
              :ref="(el) => setEditorRef(tab.name, el)"
              :fragment="tab.fragment"
              :saved-fragment="tab.savedFragment"
              :available-tags="[]"
              @change="syncTabDirty(tab.name, $event)"
              @saved="handleFragmentSaved(tab.name, $event)"
              @deleted="handleFragmentDeleted"
              @show-history="showHistory"
            />
          </el-tab-pane>
        </el-tabs>
      </div>
    </section>

    <MemoryHistoryDialog
      v-model="historyDialogVisible"
      :fragment-id="historyFragmentId"
      :git-repo-enabled="memoryGitRepoEnabled"
      :is-git-repo="memoryIsGitRepo"
      @open-settings="openSettingsDialog"
    />

    <SettingsDialog
      v-model="settingsDialogVisible"
      title="记忆设置"
      width="76%"
      @closed="refreshMemoryAfterSettingsClose"
    >
      <MemorySettingPage ref="memorySettingPage" @changed="handleMemorySettingsChanged" />
    </SettingsDialog>
  </div>
</template>

<script>
import { Check, DArrowLeft, DArrowRight, Delete, Plus, Search } from '@element-plus/icons-vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'
import MemoryEditor from '@/components/memory/MemoryEditor.vue'
import MemoryHistoryDialog from '@/components/memory/MemoryHistoryDialog.vue'
import MemorySettingPage from '@/components/set/memory.vue'
import GitActionButton from '@/components/base/GitActionButton.vue'
import SettingsDialog from '@/components/base/SettingsDialog.vue'
import sseDistribute from '@/utils/base/sse_distribute'
const {
  isMemoryFragmentTabName,
  activateMemorySaveFeedback,
  clearExpiredMemorySaveFeedback,
} = require('@/utils/memory_fragment_feedback.cjs')

// TAG_FILTER_COLLAPSED_MAX_HEIGHT 控制左侧标签筛选区收起时的最大高度。
const TAG_FILTER_COLLAPSED_MAX_HEIGHT = 76
// TAG_FILTER_TOGGLE_MIN_COUNT 控制多少个标签时展示展开/收起入口。
const TAG_FILTER_TOGGLE_MIN_COUNT = 10
// SEARCH_TAB_NAME 统一定义搜索结果标签页名称，避免散落硬编码。
const SEARCH_TAB_NAME = 'search'
// TRASH_TAB_NAME 统一定义回收站标签页名称，避免散落硬编码。
const TRASH_TAB_NAME = 'trash'
// KEYWORD_SEARCH_MODE 统一定义关键词搜索模式值，避免散落硬编码。
const KEYWORD_SEARCH_MODE = 'keyword'
// SEMANTIC_SEARCH_MODE 统一定义语义搜索模式值，避免散落硬编码。
const SEMANTIC_SEARCH_MODE = 'semantic'
// MEMORY_FRAGMENT_UPDATES_DISTRIBUTE_ID 统一定义知识片段同步推送通道。
const MEMORY_FRAGMENT_UPDATES_DISTRIBUTE_ID = 'memory_fragment_updates'
// MEMORY_FRAGMENT_STATUS_DISTRIBUTE_ID 统一定义记忆库状态推送通道。
const MEMORY_FRAGMENT_STATUS_DISTRIBUTE_ID = 'memory_fragment_status'
// MEMORY_FRAGMENT_SSE_ACTION_UPSERT 表示片段新增或更新。
const MEMORY_FRAGMENT_SSE_ACTION_UPSERT = 'upsert'
// MEMORY_FRAGMENT_SSE_ACTION_DELETE 表示片段删除。
const MEMORY_FRAGMENT_SSE_ACTION_DELETE = 'delete'

export default {
  name: 'MemoryFragment',
  components: {
    Check,
    DArrowLeft,
    DArrowRight,
    Delete,
    Plus,
    Search,
    GitActionButton,
    MemoryEditor,
    MemoryHistoryDialog,
    MemorySettingPage,
    SettingsDialog,
  },
  data() {
    return {
      fragmentList: [],
      trashList: [],
      searchResults: [],
      searchQuery: '',
      searchMode: KEYWORD_SEARCH_MODE,
      searchTabVisible: false,
      trashTabVisible: false,
      submittedSearchQuery: '',
      submittedSearchMode: KEYWORD_SEARCH_MODE,
      activeTab: '',
      fragmentTabs: [],
      searchLoading: false,
      trashLoading: false,
      historyDialogVisible: false,
      historyFragmentId: '',
      memoryConfigured: true,
      memoryGitRepoEnabled: false,
      memoryIsGitRepo: false,
      nextPushTime: 0,
      lastPushTime: 0,
      lastPushTimeDesc: '-',
      lastPushError: '',
      statusNowTick: Math.floor(Date.now() / 1000),
      settingsDialogVisible: false,
      editorRefMap: {},
      saveFeedbackMap: {},
      saveFeedbackTimers: {},
      saveFeedbackDurationMs: 1000,
      globalSaveShortcutBound: false,
      routeFragmentHandled: false,
      routeFragmentHandledPath: '',
      sidebarCollapsed: false,
    }
  },
  computed: {
    // activeFragmentId 返回当前激活的片段 id。
    activeFragmentId() {
      const tab = this.fragmentTabs.find(item => item.name === this.activeTab)
      return tab ? this.normalizeFragmentId(tab.fragment.id) : ''
    },
    // routeFragmentId 返回路由中指定的片段 id。
    routeFragmentId() {
      return this.normalizeFragmentId(this.$route.query.fragment_id)
    },
    // searchTabLabel 返回搜索结果标签名称。
    searchTabLabel() {
      if (this.submittedSearchQuery.trim() !== '') {
        return `搜索结果: ${this.submittedSearchQuery}`
      }
      return '搜索结果'
    },
    // trashTabLabel 返回回收站标签名称。
    trashTabLabel() {
      return `回收站${this.trashList.length > 0 ? ` (${this.trashList.length})` : ''}`
    },
    // pushStatusLabel 返回记忆库 push 状态标签，优先展示下一次 push。
    pushStatusLabel() {
      return this.nextPushTime > 0 ? '下一次 push 记忆库' : '上一次 push 记忆库'
    },
    // pushStatusDesc 返回记忆库 push 状态文案，优先展示下一次 push 倒计时。
    pushStatusDesc() {
      if (this.nextPushTime > 0) {
        return this.formatRelativeTime(this.nextPushTime, 'future')
      }
      if (this.lastPushTime > 0) {
        return this.formatRelativeTime(this.lastPushTime, 'past')
      }
      return this.lastPushTimeDesc || '-'
    },
    // searchPlaceholder 根据搜索模式返回对应的输入框提示文本。
    searchPlaceholder() {
      if (this.searchMode === 'semantic') {
        return '输入想要查询的内容，自然语言描述，回车打开结果页'
      }
      return '输入关键词，多个关键词使用空格分隔，回车打开结果页'
    }
  },
  mounted() {
    this.bindGlobalSaveShortcut()
    this.registerMemoryFragmentUpdatesSse()
    this.registerMemoryFragmentStatusSse()
    this.loadMemoryStatus()
    this.tryOpenRouteFragmentOnEntry()
  },
  activated() {
    this.bindGlobalSaveShortcut()
    this.registerMemoryFragmentUpdatesSse()
    this.registerMemoryFragmentStatusSse()
    this.loadMemoryStatus()
    this.tryOpenRouteFragmentOnEntry()
  },
  deactivated() {
    this.unbindGlobalSaveShortcut()
    this.unregisterMemoryFragmentUpdatesSse()
    this.unregisterMemoryFragmentStatusSse()
  },
  beforeUnmount() {
    this.unbindGlobalSaveShortcut()
    this.unregisterMemoryFragmentUpdatesSse()
    this.unregisterMemoryFragmentStatusSse()
    this.clearSaveFeedbackTimers()
  },
  watch: {
    '$route.fullPath'() {
      this.routeFragmentHandled = false
      this.tryOpenRouteFragmentOnEntry()
    },
  },
  methods: {
    // toggleSidebar 切换左侧列表的折叠/展开状态。
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
    },
    // registerMemoryFragmentUpdatesSse 注册知识片段实时同步推送。
    registerMemoryFragmentUpdatesSse() {
      sseDistribute.RegisterReceive(MEMORY_FRAGMENT_UPDATES_DISTRIBUTE_ID, (data) => {
        this.handleMemoryFragmentSseUpdate(data)
      })
    },
    // unregisterMemoryFragmentUpdatesSse 清理知识片段同步推送监听。
    unregisterMemoryFragmentUpdatesSse() {
      sseDistribute.UnRegisterReceive(MEMORY_FRAGMENT_UPDATES_DISTRIBUTE_ID)
    },
    // registerMemoryFragmentStatusSse 注册记忆库状态 SSE 推送。
    registerMemoryFragmentStatusSse() {
      sseDistribute.RegisterReceive(MEMORY_FRAGMENT_STATUS_DISTRIBUTE_ID, (data) => {
        this.handleMemoryFragmentStatusSseUpdate(data)
      })
    },
    // unregisterMemoryFragmentStatusSse 清理记忆库状态 SSE 推送监听。
    unregisterMemoryFragmentStatusSse() {
      sseDistribute.UnRegisterReceive(MEMORY_FRAGMENT_STATUS_DISTRIBUTE_ID)
    },
    // handleMemoryFragmentStatusSseUpdate 处理记忆库状态 SSE 推送。
    handleMemoryFragmentStatusSseUpdate(data) {
      this.statusNowTick = Math.floor(Date.now() / 1000)
      this.memoryConfigured = !!(data && data.configured)
      this.memoryGitRepoEnabled = !!(data && data.git_repo_enabled)
      this.memoryIsGitRepo = !!(data && data.is_git_repo)
      this.nextPushTime = data && data.next_push_time ? Number(data.next_push_time) : 0
      this.lastPushTime = data && data.last_push_time ? Number(data.last_push_time) : 0
      this.lastPushTimeDesc = data && data.last_push_time_desc ? data.last_push_time_desc : '-'
      this.lastPushError = data && data.last_push_error ? data.last_push_error : ''
      if (!this.memoryConfigured) {
        this.fragmentList = []
        this.trashList = []
        this.searchResults = []
        this.fragmentTabs = []
        this.activeTab = ''
        this.memoryGitRepoEnabled = false
        this.memoryIsGitRepo = false
        this.nextPushTime = 0
        this.lastPushTime = 0
        return
      }
      // 首次加载时需要加载列表
      if (this.fragmentList.length === 0 && this.trashList.length === 0) {
        this.loadFragmentList()
        this.loadTrashList()
      }
      this.tryOpenRouteFragmentOnEntry()
    },
    // handleMemoryFragmentSseUpdate 处理来自其他页面或异步任务的知识片段变更。
    handleMemoryFragmentSseUpdate(payload) {
      const fragmentId = this.normalizeFragmentId(payload?.fragment_id || payload?.fragment?.id || payload?.fragment?.file_id)
      const action = String(payload?.action || '').trim()
      this.loadFragmentList()
      this.loadTrashList()
      this.rerunSubmittedSearch()
      if (!fragmentId) {
        return
      }
      if (action === MEMORY_FRAGMENT_SSE_ACTION_DELETE) {
        this.handleRemoteDeletedFragment(fragmentId)
        return
      }
      if (action !== MEMORY_FRAGMENT_SSE_ACTION_UPSERT) {
        return
      }
      this.handleRemoteUpsertFragment(fragmentId, payload?.fragment || null)
    },
    // handleRemoteDeletedFragment 同步处理远端删除的片段。
    handleRemoteDeletedFragment(fragmentId) {
      this.fragmentTabs = this.fragmentTabs.filter(item => this.normalizeFragmentId(item.fragment.id) !== fragmentId)
      if (this.activeTab === `fragment-${fragmentId}`) {
        this.activeTab = ''
        this.ensureDefaultFragmentTab()
      }
    },
    // handleRemoteUpsertFragment 在安全前提下同步远端更新的片段内容。
    handleRemoteUpsertFragment(fragmentId, fragment) {
      const targetTab = this.fragmentTabs.find(item => this.normalizeFragmentId(item.fragment.id) === fragmentId)
      if (targetTab && targetTab.dirty) {
        // 中文注释：本地有未保存改动时只提醒，不直接覆盖，避免把用户草稿冲掉。
        // English comment: Warn instead of overwriting when the local editor still has unsaved draft changes.
        this.$helperNotify.warning('当前片段已被其他操作更新，请先处理本地未保存内容')
        return
      }
      if (fragment && typeof fragment === 'object' && Object.keys(fragment).length > 0) {
        this.upsertFragmentTab(fragment, false)
        return
      }
      MemoryFragmentApi.MemoryFragmentInfo(fragmentId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          return
        }
        this.upsertFragmentTab(response.Data, false)
      })
    },
    bindGlobalSaveShortcut() {
      if (this.globalSaveShortcutBound) {
        return
      }
      window.addEventListener('keydown', this.handleGlobalSaveKeydown)
      this.globalSaveShortcutBound = true
    },
    unbindGlobalSaveShortcut() {
      if (!this.globalSaveShortcutBound) {
        return
      }
      window.removeEventListener('keydown', this.handleGlobalSaveKeydown)
      this.globalSaveShortcutBound = false
    },
    handleGlobalSaveKeydown(event) {
      if (!event) {
        return
      }
      const key = String(event.key || '').toLowerCase()
      if ((!event.ctrlKey && !event.metaKey) || key !== 's') {
        return
      }
      if (!isMemoryFragmentTabName(this.activeTab)) {
        return
      }
      event.preventDefault()
      this.triggerActiveFragmentSave()
    },
    setEditorRef(tabName, instance) {
      if (!tabName) {
        return
      }
      if (instance) {
        this.editorRefMap[tabName] = instance
        return
      }
      delete this.editorRefMap[tabName]
    },
    triggerActiveFragmentSave() {
      const editor = this.editorRefMap[this.activeTab]
      if (!editor || typeof editor.triggerSave !== 'function') {
        return
      }
      editor.triggerSave()
    },
    clearSaveFeedbackTimers() {
      Object.values(this.saveFeedbackTimers).forEach((timerId) => {
        window.clearTimeout(timerId)
      })
      this.saveFeedbackTimers = {}
    },
    triggerFragmentSaveFeedback(fragmentId) {
      const normalizedId = this.normalizeFragmentId(fragmentId)
      if (!normalizedId) {
        return
      }
      if (this.saveFeedbackTimers[normalizedId]) {
        window.clearTimeout(this.saveFeedbackTimers[normalizedId])
      }
      this.saveFeedbackMap = activateMemorySaveFeedback(
        this.saveFeedbackMap,
        normalizedId,
        Date.now(),
        this.saveFeedbackDurationMs
      )
      this.saveFeedbackTimers[normalizedId] = window.setTimeout(() => {
        this.saveFeedbackMap = clearExpiredMemorySaveFeedback(this.saveFeedbackMap, Date.now())
        delete this.saveFeedbackTimers[normalizedId]
      }, this.saveFeedbackDurationMs)
    },

    // formatRelativeTime 把 unix 秒时间格式化为“xx小时xx分钟前/后”。
    formatRelativeTime(unixTime, direction) {
      const targetTime = Number(unixTime || 0)
      if (targetTime <= 0) {
        return '-'
      }
      const now = this.statusNowTick
      let diffSeconds = direction === 'future' ? targetTime - now : now - targetTime
      if (diffSeconds <= 0) {
        return direction === 'future' ? '1分钟内' : '刚刚'
      }
      diffSeconds = Math.ceil(diffSeconds / 60) * 60
      const totalMinutes = Math.floor(diffSeconds / 60)
      const hours = Math.floor(totalMinutes / 60)
      const minutes = totalMinutes % 60
      const durationText = hours > 0 ? `${hours}小时${minutes}分钟` : `${minutes}分钟`
      return direction === 'future' ? `${durationText}后` : `${durationText}前`
    },
    // copyFragmentPath 复制知识片段所属文件路径。
    async copyFragmentPath(filePath) {
      if (!filePath || !navigator.clipboard) {
        return
      }
      try {
        await navigator.clipboard.writeText(filePath)
        this.$helperNotify.success('复制地址成功')
      } catch (error) {
        this.$helperNotify.error('复制地址失败')
      }
    },
    loadMemoryStatus(needReloadLists = true) {
      MemoryFragmentApi.MemoryFragmentStatus((response) => {
        this.statusNowTick = Math.floor(Date.now() / 1000)
        this.memoryConfigured = !!(response.Data && response.Data.configured)
        this.memoryGitRepoEnabled = !!(response.Data && response.Data.git_repo_enabled)
        this.memoryIsGitRepo = !!(response.Data && response.Data.is_git_repo)
        this.nextPushTime = response.Data && response.Data.next_push_time ? Number(response.Data.next_push_time) : 0
        this.lastPushTime = response.Data && response.Data.last_push_time ? Number(response.Data.last_push_time) : 0
        this.lastPushTimeDesc = response.Data && response.Data.last_push_time_desc ? response.Data.last_push_time_desc : '-'
        this.lastPushError = response.Data && response.Data.last_push_error ? response.Data.last_push_error : ''
        if (!this.memoryConfigured) {
          this.fragmentList = []
          this.trashList = []
          this.searchResults = []
          this.fragmentTabs = []
          this.activeTab = ''
          this.memoryGitRepoEnabled = false
          this.memoryIsGitRepo = false
          this.nextPushTime = 0
          this.lastPushTime = 0
          return
        }
        if (needReloadLists) {
          this.loadFragmentList()
          this.loadTrashList()
        }
        this.tryOpenRouteFragmentOnEntry()
      })
    },
    // loadFragmentList 加载左侧片段列表。
    loadFragmentList() {
      if (!this.memoryConfigured) {
        return
      }
      MemoryFragmentApi.MemoryFragmentList(0, (response) => {
        this.fragmentList = Array.isArray(response.Data) ? response.Data : []
        this.ensureDefaultFragmentTab()
      })
    },
    // ensureDefaultFragmentTab 在没有激活片段时自动打开列表中的第一个知识片段。
    ensureDefaultFragmentTab() {
      if (!this.memoryConfigured) {
        return
      }
      if (this.activeTab === SEARCH_TAB_NAME || this.activeTab === TRASH_TAB_NAME) {
        return
      }
      const hasActiveFragmentTab = this.fragmentTabs.some(item => item.name === this.activeTab)
      if (hasActiveFragmentTab) {
        return
      }
      const firstItem = this.fragmentList[0]
      if (!firstItem) {
        this.activeTab = ''
        return
      }
      const firstFragmentId = this.normalizeFragmentId(firstItem.id || firstItem.file_id)
      if (!firstFragmentId) {
        this.activeTab = ''
        return
      }
      this.openFragment(firstFragmentId)
    },
    // fragmentFreshnessClass 根据更新时间返回左侧片段的新鲜度样式类。
    fragmentFreshnessClass(item) {
      const dayMs = 24 * 60 * 60 * 1000
      const now = Date.now()
      const startOfToday = new Date()
      startOfToday.setHours(0, 0, 0, 0)
      const updateTime = Number(item && item.update_time ? item.update_time : 0)
      const updateAt = updateTime > 0 ? updateTime * 1000 : 0
      if (updateAt >= startOfToday.getTime()) {
        return 'is-updated-today'
      }
      if (updateAt >= now - 3 * dayMs) {
        return 'is-updated-3d'
      }
      if (updateAt >= now - 7 * dayMs) {
        return 'is-updated-7d'
      }
      return 'is-updated-older'
    },
    // sidebarItemKey 为左侧列表项构造稳定且可重启动画的 key。
    // sidebarItemKey builds a stable sidebar key while allowing save-feedback animation to replay on each save.
    sidebarItemKey(item) {
      const normalizedFragmentId = this.normalizeFragmentId(item && (item.id || item.file_id))
      const feedback = normalizedFragmentId ? this.saveFeedbackMap[normalizedFragmentId] : null
      const feedbackStartedAt = feedback && feedback.startedAt ? Number(feedback.startedAt) : 0
      return `${normalizedFragmentId || 'fragment'}-${feedbackStartedAt}`
    },
    // isFragmentDirty 判断左侧片段是否存在未保存改动。
    // isFragmentDirty checks whether the sidebar fragment currently has unsaved edits.
    isFragmentDirty(fragmentId) {
      const normalizedFragmentId = this.normalizeFragmentId(fragmentId)
      if (!normalizedFragmentId) {
        return false
      }
      return this.fragmentTabs.some(item => item.dirty && item.fragment.id === normalizedFragmentId)
    },
    // loadTrashList 加载回收站片段列表。
    loadTrashList() {
      if (!this.memoryConfigured) {
        return
      }
      this.trashLoading = true
      MemoryFragmentApi.MemoryFragmentTrashList(0, (response) => {
        this.trashLoading = false
        this.trashList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    // runSearch 根据指定条件执行搜索。
    runSearch(query, mode) {
      if (!this.memoryConfigured) {
        return
      }
      this.searchLoading = true
      MemoryFragmentApi.MemoryFragmentSearch(
        query,
        mode,
        [],
        50,
        (response) => {
          this.searchLoading = false
          this.searchResults = Array.isArray(response.Data) ? response.Data : []
        }
      )
    },
    // submitSearch 提交当前搜索条件并打开搜索结果 tab。
    submitSearch() {
      if (this.searchQuery.trim() === '') {
        this.clearFilter()
        return
      }
      this.submittedSearchQuery = this.searchQuery.trim()
      this.submittedSearchMode = this.searchMode
      this.searchTabVisible = true
      this.activeTab = SEARCH_TAB_NAME
      this.runSearch(this.submittedSearchQuery, this.submittedSearchMode)
    },
    // escapeHtml 对文本做 HTML 转义，避免高亮时插入原始标签。
    escapeHtml(text) {
      return String(text || '')
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;')
    },
    // rerunSubmittedSearch 重新执行当前搜索结果 tab 的查询。
    rerunSubmittedSearch() {
      if (!this.searchTabVisible) {
        return
      }
      this.runSearch(this.submittedSearchQuery, this.submittedSearchMode)
    },
    // handleSearchClear 处理搜索输入框清空。
    handleSearchClear() {
      this.searchQuery = ''
    },
    // clearFilter 清空左侧搜索条件，并关闭结果 tab。
    clearFilter() {
      this.searchQuery = ''
      this.searchMode = KEYWORD_SEARCH_MODE
      this.submittedSearchQuery = ''
      this.submittedSearchMode = KEYWORD_SEARCH_MODE
      this.searchTabVisible = false
      this.searchResults = []
      if (this.activeTab === SEARCH_TAB_NAME) {
        this.activeTab = ''
        this.ensureDefaultFragmentTab()
      }
    },
    // getSearchSnippet 生成搜索结果中的命中文本片段。
    getSearchSnippet(item) {
      const sourceText = (item.content_text || item.content || '').replace(/\s+/g, ' ').trim()
      if (sourceText === '') {
        return '无正文内容'
      }
      const keywords = this.buildSearchKeywords()
      if (keywords.length === 0) {
        return sourceText.slice(0, 120)
      }
      const lowerSourceText = sourceText.toLowerCase()
      let hitIndex = -1
      let hitKeyword = ''
      keywords.forEach((keyword) => {
        const index = lowerSourceText.indexOf(keyword.toLowerCase())
        if (index >= 0 && (hitIndex === -1 || index < hitIndex)) {
          hitIndex = index
          hitKeyword = keyword
        }
      })
      if (hitIndex === -1) {
        return sourceText.slice(0, 120)
      }
      const start = Math.max(0, hitIndex - 24)
      const end = Math.min(sourceText.length, hitIndex + hitKeyword.length + 72)
      const prefix = start > 0 ? '...' : ''
      const suffix = end < sourceText.length ? '...' : ''
      return prefix + sourceText.slice(start, end) + suffix
    },
    // buildSearchKeywords 汇总本次已提交搜索条件的关键词。
    buildSearchKeywords() {
      const keywordMap = {}
      const keywords = []
      this.submittedSearchQuery.split(/\s+/).forEach((item) => {
        const keyword = item.trim()
        const normalizedKeyword = keyword.toLowerCase()
        if (keyword === '' || keywordMap[normalizedKeyword]) {
          return
        }
        keywordMap[normalizedKeyword] = true
        keywords.push(keyword)
      })
        return keywords
    },
    // getSearchSnippetList 生成最多 3 条搜索命中片段。
    getSearchSnippetList(item) {
      const serverSnippets = Array.isArray(item.search_snippets) ? item.search_snippets.filter(Boolean) : []
      if (serverSnippets.length > 0) {
        return serverSnippets.slice(0, 3)
      }
      const sourceText = (item.content_text || item.content || '').replace(/\s+/g, ' ').trim()
      if (sourceText === '') {
        return ['无正文内容']
      }
      const keywords = this.buildSearchKeywords()
      if (keywords.length === 0) {
        return [sourceText.slice(0, 120)]
      }
      const lowerSourceText = sourceText.toLowerCase()
      const hitPositions = []
      keywords.forEach((keyword) => {
        const lowerKeyword = keyword.toLowerCase()
        let startIndex = 0
        while (startIndex < lowerSourceText.length) {
          const foundIndex = lowerSourceText.indexOf(lowerKeyword, startIndex)
          if (foundIndex === -1) {
            break
          }
          hitPositions.push({
            index: foundIndex,
            keyword: sourceText.slice(foundIndex, foundIndex + keyword.length),
          })
          startIndex = foundIndex + lowerKeyword.length
        }
      })
      if (hitPositions.length === 0) {
        return [sourceText.slice(0, 120)]
      }
      hitPositions.sort((left, right) => left.index - right.index)
      const snippets = []
      let lastEnd = -1
      hitPositions.forEach((hit) => {
        const snippetStart = Math.max(0, hit.index - 24)
        const snippetEnd = Math.min(sourceText.length, hit.index + hit.keyword.length + 72)
        if (snippetStart < lastEnd) {
          return
        }
        const prefix = snippetStart > 0 ? '...' : ''
        const suffix = snippetEnd < sourceText.length ? '...' : ''
        snippets.push(prefix + sourceText.slice(snippetStart, snippetEnd) + suffix)
        lastEnd = snippetEnd
      })
      return snippets.slice(0, 3)
    },
    // getSearchSnippetMoreCount 返回未展示的命中片段数量。
    getSearchSnippetMoreCount(item) {
      if (Array.isArray(item.search_snippets) && item.search_snippets.length > 0) {
        return Math.max(0, item.search_snippets.length - 3)
      }
      const sourceText = (item.content_text || item.content || '').replace(/\s+/g, ' ').trim()
      if (sourceText === '') {
        return 0
      }
      const keywords = this.buildSearchKeywords()
      if (keywords.length === 0) {
        return 0
      }
      const lowerSourceText = sourceText.toLowerCase()
      const snippetCount = []
      keywords.forEach((keyword) => {
        const lowerKeyword = keyword.toLowerCase()
        let startIndex = 0
        while (startIndex < lowerSourceText.length) {
          const foundIndex = lowerSourceText.indexOf(lowerKeyword, startIndex)
          if (foundIndex === -1) {
            break
          }
          snippetCount.push(foundIndex)
          startIndex = foundIndex + lowerKeyword.length
        }
      })
      const uniqueHitCount = snippetCount.sort((left, right) => left - right).filter((itemIndex, index, arr) => {
        if (index === 0) {
          return true
        }
        return itemIndex !== arr[index - 1]
      }).length
      return Math.max(0, uniqueHitCount - 3)
    },
    // highlightSearchKeywords 把片段中的命中关键词标成红色。
    highlightSearchKeywords(text) {
      let html = this.escapeHtml(text)
      const keywords = this.buildSearchKeywords().sort((left, right) => right.length - left.length)
      keywords.forEach((keyword) => {
        const escapedKeyword = this.escapeHtml(keyword)
        if (escapedKeyword === '') {
          return
        }
        const reg = new RegExp(this.escapeRegExp(escapedKeyword), 'gi')
        html = html.replace(reg, '<span class="search-keyword-highlight">$&</span>')
      })
      return html
    },
    // escapeRegExp 转义正则特殊字符。
    escapeRegExp(text) {
      return String(text || '').replace(/[.*+?^${}()|[\]\\]/g, '\\$&')
    },
    // createFragment 创建一个新片段并自动打开。
    createFragment() {
      if (!this.memoryConfigured) {
        return
      }
      MemoryFragmentApi.MemoryFragmentSave(0, '新知识片段', '# 新知识片段\n\n在这里开始记录。', [], (response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.loadFragmentList()
        this.upsertFragmentTab(response.Data, true)
      })
    },
    // openTrashTab 打开回收站 tab 并刷新内容。
    openTrashTab() {
      this.trashTabVisible = true
      this.activeTab = TRASH_TAB_NAME
      this.loadTrashList()
    },
    // openSettingsDialog 打开记忆设置弹窗，在当前业务页内完成 AI 配置维护。
    // Open the memory settings modal so AI configuration can be maintained in-place.
    openSettingsDialog() {
      this.settingsDialogVisible = true
      this.$nextTick(() => {
        if (this.$refs.memorySettingPage && this.$refs.memorySettingPage.loadConfig) {
          this.$refs.memorySettingPage.loadConfig()
        }
      })
    },
    // handleMemorySettingsChanged 设置保存成功后立即刷新记忆状态区展示。
    // Refresh memory status immediately after settings change.
    handleMemorySettingsChanged() {
      this.loadMemoryStatus(false)
    },
    // refreshMemoryAfterSettingsClose 在弹窗关闭时再做一次兜底刷新。
    // Refresh once more when the dialog closes as a fallback for additional setting edits.
    refreshMemoryAfterSettingsClose() {
      this.loadMemoryStatus(false)
    },
    // openFragment 打开指定片段 tab。
    openFragment(fragmentId) {
      if (!this.memoryConfigured) {
        return
      }
      const normalizedFragmentId = this.normalizeFragmentId(fragmentId)
      if (!normalizedFragmentId) {
        return
      }
      const existingTab = this.fragmentTabs.find(item => item.fragment.id === normalizedFragmentId)
      if (existingTab) {
        this.activeTab = existingTab.name
        return
      }
      MemoryFragmentApi.MemoryFragmentInfo(normalizedFragmentId, (response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.upsertFragmentTab(response.Data, true)
      })
    },
    // openRouteFragment 根据路由参数自动打开目标知识片段。
    openRouteFragment() {
      if (!this.memoryConfigured) {
        return
      }
      const fragmentId = this.routeFragmentId
      if (!fragmentId) {
        return
      }
      this.openFragment(fragmentId)
    },
    // tryOpenRouteFragmentOnEntry 仅在当前路由首次进入时消费 fragment_id，避免轮询刷新反复切回指定片段。
    tryOpenRouteFragmentOnEntry() {
      if (!this.memoryConfigured) {
        return
      }
      const currentPath = this.$route.fullPath || ''
      if (this.routeFragmentHandled && this.routeFragmentHandledPath === currentPath) {
        return
      }
      const fragmentId = this.routeFragmentId
      if (!fragmentId) {
        this.routeFragmentHandled = true
        this.routeFragmentHandledPath = currentPath
        return
      }
      this.routeFragmentHandled = true
      this.routeFragmentHandledPath = currentPath
      this.openFragment(fragmentId)
    },
    // upsertFragmentTab 新增或更新片段 tab。
    upsertFragmentTab(fragment, switchTab) {
      const tabName = `fragment-${fragment.id}`
      const normalized = this.normalizeFragment(fragment)
      const existingIndex = this.fragmentTabs.findIndex(item => item.name === tabName)
      const newTab = {
        name: tabName,
        fragment: normalized,
        savedFragment: this.cloneFragment(normalized),
        dirty: false,
      }
      if (existingIndex >= 0) {
        this.fragmentTabs.splice(existingIndex, 1, newTab)
      } else {
        this.fragmentTabs.push(newTab)
      }
      if (switchTab) {
        this.activeTab = tabName
      }
    },
    // normalizeFragment 统一片段对象结构。
    normalizeFragment(fragment) {
      return {
        id: this.normalizeFragmentId(fragment.id || fragment.file_id),
        title: fragment.title || '',
        content: fragment.content || '',
        file_path: fragment.file_path || '',
        update_time_desc: fragment.update_time_desc || '',
        create_time_desc: fragment.create_time_desc || '',
      }
    },
    normalizeFragmentId(id) {
      const text = String(id || '').trim()
      if (!text || text === '0' || text === 'null' || text === 'undefined') {
        return ''
      }
      return text
    },
    // cloneFragment 克隆片段对象。
    cloneFragment(fragment) {
      return JSON.parse(JSON.stringify(fragment))
    },
    // syncTabDirty 根据当前 tab 内容同步未保存状态。
    syncTabDirty(tabName, fragment) {
      const target = this.fragmentTabs.find(item => item.name === tabName)
      if (!target) {
        return
      }
      target.fragment = this.normalizeFragment(fragment)
      target.dirty = JSON.stringify(this.cloneFragment(target.fragment)) !== JSON.stringify(this.cloneFragment(target.savedFragment))
    },
    // handleFragmentSaved 处理片段保存成功后的联动。
    handleFragmentSaved(tabName, fragment) {
      const target = this.fragmentTabs.find(item => item.name === tabName)
      if (!target) {
        return
      }
      target.fragment = this.normalizeFragment(fragment)
      target.savedFragment = this.cloneFragment(target.fragment)
      target.dirty = false
      this.triggerFragmentSaveFeedback(target.fragment.id)
      this.loadFragmentList()
      this.loadTrashList()
      this.rerunSubmittedSearch()
    },
    // handleFragmentDeleted 删除片段后清理 tab 和列表。
    handleFragmentDeleted(fragmentId) {
      this.fragmentTabs = this.fragmentTabs.filter(item => item.fragment.id !== fragmentId)
      this.loadFragmentList()
      this.loadTrashList()
      this.rerunSubmittedSearch()
      if (this.activeTab === `fragment-${fragmentId}`) {
        this.activeTab = ''
        this.ensureDefaultFragmentTab()
      }
    },
    // handleFragmentRestore 从回收站恢复片段并刷新列表。
    handleFragmentRestore(fragmentId) {
      MemoryFragmentApi.MemoryFragmentRestore(fragmentId, (response) => {
        if (response.ErrCode !== 0) {
          return
        }
        this.loadFragmentList()
        this.loadTrashList()
        this.rerunSubmittedSearch()
      })
    },
    // handleFragmentHardDelete 彻底删除回收站中的片段。
    handleFragmentHardDelete(fragmentId) {
      MemoryFragmentApi.MemoryFragmentHardDelete(fragmentId, (response) => {
        if (response.ErrCode !== 0) {
          return
        }
        this.fragmentTabs = this.fragmentTabs.filter(item => item.fragment.id !== fragmentId)
        this.loadFragmentList()
        this.loadTrashList()
        this.rerunSubmittedSearch()
        if (this.activeTab === `fragment-${fragmentId}`) {
          this.activeTab = this.trashTabVisible ? TRASH_TAB_NAME : ''
          this.ensureDefaultFragmentTab()
        }
      })
    },
    // showHistory 打开历史记录弹窗。
    showHistory(fragmentId) {
      this.historyFragmentId = fragmentId
      this.historyDialogVisible = true
    },
    // closeTab 关闭一个编辑 tab 或搜索结果 tab。
    closeTab(tabName) {
      if (tabName === SEARCH_TAB_NAME) {
        this.searchTabVisible = false
        this.searchResults = []
        if (this.activeTab === SEARCH_TAB_NAME) {
          this.activeTab = ''
          this.ensureDefaultFragmentTab()
        }
        return
      }
      if (tabName === TRASH_TAB_NAME) {
        this.trashTabVisible = false
        if (this.activeTab === TRASH_TAB_NAME) {
          this.activeTab = ''
          this.ensureDefaultFragmentTab()
        }
        return
      }
      const targetIndex = this.fragmentTabs.findIndex(item => item.name === tabName)
      if (targetIndex < 0) {
        return
      }
      this.fragmentTabs.splice(targetIndex, 1)
      if (this.activeTab === tabName) {
        this.activeTab = this.fragmentTabs.length > 0 ? this.fragmentTabs[Math.max(targetIndex - 1, 0)].name : ''
        this.ensureDefaultFragmentTab()
      }
    },
    // handleTabChange 切换 tab 时保持页面状态一致。
    handleTabChange(tabPane) {
      this.activeTab = tabPane.paneName
    },
  }
}
</script>

<style scoped>
@property --sidebar-save-border-angle {
  syntax: '<angle>';
  inherits: false;
  initial-value: 0deg;
}

.memory-page {
  display: flex;
  gap: 4px;
  height: calc(100vh - 40px);
  min-height: 680px;
}

.memory-sidebar {
  width: 320px;
  flex-shrink: 0;
  border: 1px solid #e8e8e0;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 4px 12px rgba(54, 74, 54, 0.05);
  overflow: hidden;
  display: flex;
  flex-direction: column;
}

.sidebar-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 8px;
  padding: 12px 14px;
  border-bottom: 1px solid #ecece4;
  background: #f7f7f2;
}

.sidebar-collapse-btn {
  display: flex;
  align-items: center;
  justify-content: center;
  width: 16px;
  height: 32px;
  border: 1px solid #ddd;
  border-radius: 3px;
  background: #fff;
  color: #bbb;
  cursor: pointer;
  flex-shrink: 0;
  align-self: center;
  transition: color 0.15s, border-color 0.15s;
  padding: 0;
}

.sidebar-collapse-btn:hover {
  color: #5f7d56;
  border-color: #5f7d56;
}

.sidebar-header-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
}

.sidebar-title {
  font-size: 15px;
  font-weight: 600;
  color: #4a5a45;
}

.sidebar-scroll {
  flex: 1;
}

.sidebar-footer {
  display: flex;
  flex-direction: column;
  gap: 6px;
  padding: 10px 14px 12px;
  border-top: 1px solid #ecece4;
  background: #f9faf6;
  font-size: 12px;
  color: #6c7767;
}

.sidebar-footer-row {
  display: flex;
  gap: 8px;
  align-items: center;
  line-height: 1.4;
}

.sidebar-footer-label {
  flex-shrink: 0;
}

.sidebar-footer-value {
  color: #4f5f49;
  flex: 1;
}

.sidebar-footer-error .sidebar-footer-value {
  color: #bb5b4a;
  white-space: normal;
  word-break: break-word;
}

.sidebar-footer-error {
  align-items: flex-start;
}

:global(.memory-fragment-delete-popconfirm) {
  max-width: 320px;
}

.delete-popconfirm-content {
  text-align: center;
}

.delete-popconfirm-desc {
  color: #5f6758;
  line-height: 1.6;
}

.delete-popconfirm-name {
  margin-top: 10px;
  color: #c23b32;
  font-size: 15px;
  font-weight: 700;
  line-height: 1.6;
  word-break: break-word;
}

.sidebar-item {
  position: relative;
  isolation: isolate;
  --sidebar-save-border-angle: 0deg;
  width: calc(100% - 16px);
  margin: 8px;
  padding: 12px;
  border: 1px solid #edf1e8;
  border-radius: 12px;
  background: #fbfcf8;
  cursor: pointer;
  text-align: left;
  transition: all 0.2s ease;
}

.sidebar-item.is-updated-today {
  border-color: #d6e6cf;
  background: #f4f9ef;
}

.sidebar-item.is-updated-3d {
  border-color: #e2eadb;
  background: #f8fbf5;
}

.sidebar-item.is-updated-7d {
  border-color: #e7ece2;
  background: #fafcf8;
}

.sidebar-item.is-updated-older {
  border-color: #ecf0e8;
  background: #fcfdfb;
}

.sidebar-item.sidebar-item--dirty {
  border-color: rgba(224, 191, 146, 0.78);
  background: linear-gradient(180deg, #fdf7ef 0%, #f7eee1 100%);
  box-shadow: inset 0 -1px 0 rgba(241, 214, 179, 0.52);
}

.sidebar-item::before {
  content: '';
  position: absolute;
  inset: -1px;
  border-radius: inherit;
  padding: 2px;
  background: conic-gradient(
    from var(--sidebar-save-border-angle),
    rgba(63, 154, 84, 0) 0deg,
    rgba(63, 154, 84, 0) 235deg,
    rgba(63, 154, 84, 0.24) 275deg,
    rgba(63, 154, 84, 0.98) 312deg,
    rgba(151, 220, 167, 0.92) 330deg,
    rgba(63, 154, 84, 0.12) 345deg,
    rgba(63, 154, 84, 0) 360deg
  );
  opacity: 0;
  pointer-events: none;
  z-index: 0;
  -webkit-mask: linear-gradient(#fff 0 0) content-box, linear-gradient(#fff 0 0);
  -webkit-mask-composite: xor;
  mask-composite: exclude;
}

.sidebar-item.save-success::before {
  opacity: 1;
  animation: sidebar-item-border-flow 1s linear 1;
}

.sidebar-item:hover,
.sidebar-item.active {
  border-color: #cfe0c8;
  background: #f2f8ec;
}

.sidebar-item:hover {
  border-color: #d9e7d3;
  background: #f5faef;
}

.sidebar-item.sidebar-item--dirty:hover {
  border-color: rgba(222, 182, 128, 0.88);
  background: linear-gradient(180deg, #fdf6ec 0%, #f6ead9 100%);
}

.sidebar-item.active {
  border-color: #93b88a;
  background: linear-gradient(135deg, #f1f8e8 0%, #e6f3da 100%);
  box-shadow: 0 14px 26px rgba(83, 122, 72, 0.16);
  transform: translateX(4px);
}

.sidebar-item.sidebar-item--dirty.active {
  border-color: rgba(213, 170, 112, 0.92);
  background: linear-gradient(135deg, #fdf7ef 0%, #f6ead7 100%);
  box-shadow: 0 14px 26px rgba(181, 140, 89, 0.16);
}

.sidebar-item.active::after {
  content: '';
  position: absolute;
  left: -1px;
  top: 10px;
  bottom: 10px;
  width: 4px;
  border-radius: 999px;
  background: linear-gradient(180deg, #5f9754 0%, #78b66c 100%);
  box-shadow: 0 0 0 1px rgba(95, 151, 84, 0.08);
  z-index: 1;
}

.sidebar-item.sidebar-item--dirty.active::after {
  background: linear-gradient(180deg, #d79b53 0%, #e6b77a 100%);
  box-shadow: 0 0 0 1px rgba(215, 155, 83, 0.08);
}

.sidebar-item-main {
  position: relative;
  z-index: 1;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 4px;
}

.sidebar-item-title-row {
  display: flex;
  align-items: flex-start;
  gap: 8px;
}

.sidebar-item-title {
  flex: 1;
  display: block;
  width: 100%;
  min-width: 0;
  font-weight:500;
  font-size: 14px;
  color: #32402f;
  line-height: 1.5;
}

.sidebar-item.active .sidebar-item-title {
  color: #1f301d;
  font-weight: 600;
}

.sidebar-item-active-badge {
  flex-shrink: 0;
  display: inline-flex;
  align-items: center;
  min-height: 22px;
  padding: 0 8px;
  border-radius: 999px;
  background: rgba(72, 127, 61, 0.12);
  color: #3f6d37;
  font-size: 11px;
  font-weight: 600;
  line-height: 1;
}

.sidebar-item.is-updated-3d .sidebar-item-title {
  color: #41503d;
}

.sidebar-item.is-updated-7d .sidebar-item-title {
  color: #4f5d4c;
}

.sidebar-item.is-updated-older .sidebar-item-title {
  color: #616d5f;
}

.sidebar-item-meta {
  position: relative;
  z-index: 1;
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-top: 6px;
}

.sidebar-item-copy {
  display: inline-flex;
  align-items: center;
  min-height: 24px;
  padding: 0 8px;
  border-radius: 999px;
  background: #eef5e9;
  color: #56714f;
  font-size: 11px;
  line-height: 1;
  cursor: pointer;
  user-select: none;
}

.sidebar-item.is-updated-3d .sidebar-item-copy {
  background: #f1f5ed;
  color: #5d7158;
}

.sidebar-item.is-updated-7d .sidebar-item-copy {
  background: #f4f7f1;
  color: #687866;
}

.sidebar-item.is-updated-older .sidebar-item-copy {
  background: #f6f8f4;
  color: #748070;
}

.sidebar-item-copy:focus-visible {
  outline: 2px solid rgba(63, 154, 84, 0.35);
  outline-offset: 2px;
}

.sidebar-item-time {
  color: #7b8576;
  font-size: 12px;
  margin-left: auto;
  white-space: nowrap;
}

.sidebar-item.active .sidebar-item-time {
  color: #55704f;
}

.sidebar-item.is-updated-3d .sidebar-item-time {
  color: #879183;
}

.sidebar-item.is-updated-7d .sidebar-item-time {
  color: #91998d;
}

.sidebar-item.is-updated-older .sidebar-item-time {
  color: #9ba298;
}

.sidebar-item-path {
  position: relative;
  z-index: 1;
  margin-top: 6px;
  padding-top: 8px;
  border-top: 1px dashed #e0e6dc;
}

.sidebar-item-path-text {
  display: block;
  font-size: 11px;
  color: #8a9582;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.sidebar-item-tags {
  position: relative;
  z-index: 1;
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: 10px;
}

.sidebar-item-check {
  position: absolute;
  right: 10px;
  bottom: 10px;
  z-index: 1;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  border-radius: 999px;
  background: linear-gradient(135deg, #40b364 0%, #2a954d 100%);
  color: #ffffff;
  box-shadow: 0 10px 18px rgba(42, 149, 77, 0.24);
  animation: sidebar-item-check-pop 1s ease forwards;
}

@keyframes sidebar-item-border-flow {
  from {
    --sidebar-save-border-angle: 0deg;
  }
  to {
    --sidebar-save-border-angle: 360deg;
  }
}

@keyframes sidebar-item-check-pop {
  0% {
    opacity: 0;
    transform: translateY(4px) scale(0.7);
  }
  18% {
    opacity: 1;
    transform: translateY(0) scale(1.05);
  }
  82% {
    opacity: 1;
    transform: translateY(0) scale(1);
  }
  100% {
    opacity: 0;
    transform: translateY(-2px) scale(0.92);
  }
}

.memory-main {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
}

.search-card,
.workspace-card {
  border: 1px solid #e8e8e0;
  border-radius: 14px;
  background: #fff;
  box-shadow: 0 4px 12px rgba(54, 74, 54, 0.05);
}

.search-card {
  padding: 16px;
}

.sidebar-search-card {
  margin: 12px 12px 0 12px;
  border-radius: 12px;
  box-shadow: none;
  background: #fbfcf8;
}

.sidebar-search-card .search-row {
  flex-direction: column;
  align-items: stretch;
}

.sidebar-search-card .tag-filter-row {
  flex-direction: column;
  align-items: stretch;
}

.sidebar-search-card .search-mode-row {
  margin-top: 10px;
  display: flex;
  justify-content: flex-start;
}

/* 搜索模式切换按钮自定义样式 - 绿色主题 */
.search-mode-row :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background-color: #5f7d56;
  border-color: #5f7d56;
  box-shadow: -1px 0 0 0 #5f7d56;
}

.search-mode-row :deep(.el-radio-button__inner:hover) {
  color: #5f7d56;
}

.search-mode-row :deep(.el-radio-button__original-radio:checked + .el-radio-button__inner:hover) {
  color: #fff;
}

.sidebar-search-card .tag-filter-label {
  min-width: 0;
  line-height: 1.2;
}

.search-row {
  display: flex;
  gap: 12px;
  align-items: center;
}

.search-row :deep(.el-radio-group) {
  flex-shrink: 0;
}

.tag-filter-row {
  display: flex;
  flex-direction: column;
  gap: 10px;
  margin-top: 12px;
}

.tag-filter-head {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
}

.tag-filter-label {
  color: #60705a;
  font-size: 13px;
  line-height: 1.4;
}

.tag-filter-toggle {
  padding: 0;
  border: none;
  background: transparent;
  color: #5f7d56;
  font-size: 12px;
  cursor: pointer;
}

.tag-filter-toggle:hover {
  color: #45603e;
  text-decoration: underline;
}

.tag-filter-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  overflow: hidden;
  transition: max-height 0.2s ease;
}

.tag-filter-list.collapsed {
  mask-image: linear-gradient(180deg, #000 0%, #000 78%, rgba(0, 0, 0, 0) 100%);
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 7px 11px;
  border: 1px solid #dbe7d4;
  border-radius: 999px;
  background: #f8fbf5;
  color: #4f6448;
  cursor: pointer;
  transition: all 0.2s ease;
}

.filter-chip.active {
  border-color: #81a478;
  background: #edf6e7;
  color: #35512f;
}

.filter-count {
  min-width: 20px;
  padding: 1px 6px;
  border-radius: 999px;
  background: rgba(86, 123, 76, 0.12);
  font-size: 12px;
}

.search-actions {
  display: flex;
  gap: 10px;
  margin-top: 12px;
}

.workspace-card {
  flex: 1;
  min-height: 0;
  padding: 14px;
}

.search-result-panel {
  display: flex;
  flex-direction: column;
  gap: 16px;
  min-height: 100%;
}

.search-result-toolbar {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  padding: 18px 20px;
  border: 1px solid #e8e8e0;
  border-radius: 14px;
  background: linear-gradient(135deg, #f7fbf2 0%, #ffffff 60%, #eef5e8 100%);
}

.search-result-summary {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.search-result-title {
  font-size: 18px;
  font-weight: 700;
  color: #42563d;
}

.search-result-desc {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
  color: #667660;
  font-size: 13px;
}

.search-result-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  justify-content: flex-end;
}

.search-result-list {
  display: flex;
  flex-direction: column;
  gap: 10px;
}

.search-result-item {
  width: 100%;
  padding: 16px 18px;
  border: 1px solid #e8eee3;
  border-radius: 14px;
  background: #fbfcf8;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s ease;
}

.search-result-item:hover {
  border-color: #cfe0c8;
  background: #f4f9ee;
  transform: translateY(-1px);
}

.trash-result-item {
  width: 100%;
  padding: 16px 18px;
  border: 1px solid #e8eee3;
  border-radius: 14px;
  background: #fbfcf8;
}

.search-result-item-head {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 12px;
}

.search-result-item-title {
  font-size: 15px;
  font-weight: 700;
  color: #30402d;
  line-height: 1.5;
}

.search-result-item-time {
  color: #7b8576;
  font-size: 12px;
  white-space: nowrap;
}

.search-result-item-snippet {
  margin-top: 10px;
  color: #596a54;
  font-size: 14px;
  line-height: 1.7;
}

.search-result-snippet-line + .search-result-snippet-line {
  margin-top: 6px;
}

.search-result-snippet-more {
  margin-top: 8px;
  color: #7d866f;
  font-size: 12px;
}

.search-result-item-tags {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  margin-top: 12px;
}

.trash-result-actions {
  display: flex;
  gap: 10px;
  margin-top: 14px;
}

.search-result-item-snippet :deep(.search-keyword-highlight) {
  color: #c43d2f;
  font-weight: 700;
}

.memory-tabs {
  height: 100%;
}

.memory-tabs :deep(.el-tabs__content) {
  height: calc(100% - 42px);
  overflow: auto;
}

.memory-tabs :deep(.el-tab-pane) {
  height: 100%;
}

.tab-label {
  display: inline-flex;
  align-items: center;
  gap: 4px;
}

@media (max-width: 1180px) {
  .memory-page {
    flex-direction: column;
    height: auto;
  }

  .memory-sidebar {
    width: 100%;
  }

  .sidebar-search-card .search-row,
  .tag-filter-row,
  .search-actions {
    flex-direction: column;
    align-items: stretch;
  }

  .search-result-toolbar {
    flex-direction: column;
  }

  .search-result-tags {
    justify-content: flex-start;
  }

  .search-result-item-head {
    flex-direction: column;
  }
}
</style>

