<template>
  <div class="memory-page">
    <aside class="memory-sidebar">
      <div class="sidebar-header">
        <div class="sidebar-title">片段列表</div>
        <el-button type="primary" plain @click="createFragment">
          <el-icon><Plus /></el-icon>
          新建片段
        </el-button>
      </div>

      <div class="search-card sidebar-search-card">
        <div class="search-row">
          <el-input
            v-model="searchQuery"
            clearable
            placeholder="搜索标题、正文或标签，回车打开结果页"
            @keydown.enter.prevent="submitSearch"
            @clear="handleSearchClear"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-radio-group v-model="searchMode">
            <el-radio-button label="hybrid">混合</el-radio-button>
            <el-radio-button label="keyword">关键词</el-radio-button>
            <el-radio-button label="natural">自然语言</el-radio-button>
          </el-radio-group>
        </div>
        <div class="tag-filter-row">
          <span class="tag-filter-label">标签筛选</span>
          <div class="tag-filter-list">
            <button
              v-for="tag in tagList"
              :key="tag.tag_name"
              class="filter-chip"
              :class="{ active: selectedTags.includes(tag.tag_name) }"
              @click="toggleTag(tag.tag_name)"
            >
              <span>{{ tag.tag_name }}</span>
              <span class="filter-count">{{ tag.use_count }}</span>
            </button>
          </div>
        </div>
        <div class="search-actions">
          <el-button type="primary" @click="submitSearch">
            <el-icon><Search /></el-icon>
            搜索
          </el-button>
          <el-button plain @click="clearFilter">清空条件</el-button>
        </div>
      </div>

      <el-scrollbar class="sidebar-scroll">
        <button
          v-for="item in fragmentList"
          :key="item.id"
          class="sidebar-item"
          :class="{ active: activeFragmentId === item.id }"
          @click="openFragment(item.id)"
        >
          <div class="sidebar-item-main">
            <div class="sidebar-item-title">{{ item.title }}</div>
            <div class="sidebar-item-time">{{ item.update_time_desc || '-' }}</div>
          </div>
          <div v-if="item.tags && item.tags.length > 0" class="sidebar-item-tags">
            <el-tag
              v-for="tag in item.tags.slice(0, 3)"
              :key="tag"
              size="small"
              effect="plain"
            >
              {{ tag }}
            </el-tag>
          </div>
        </button>
      </el-scrollbar>
    </aside>

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
          <el-tab-pane name="home" :closable="false">
            <template #label>
              <span class="tab-label">首页</span>
            </template>
            <MemoryWelcome
              :recent-list="fragmentList"
              :search-results="[]"
              :tag-list="tagList"
              :selected-tags="[]"
              :query="''"
              :loading="false"
              @open-fragment="openFragment"
              @create-fragment="createFragment"
              @toggle-tag="toggleTag"
              @clear-filter="clearFilter"
            />
          </el-tab-pane>

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
                    <span v-if="submittedSelectedTags.length > 0">标签：{{ submittedSelectedTags.join('、') }}</span>
                    <span>模式：{{ searchModeText(submittedSearchMode) }}</span>
                    <span>命中：{{ searchResults.length }}</span>
                  </div>
                </div>
                <div v-if="submittedSelectedTags.length > 0" class="search-result-tags">
                  <el-tag
                    v-for="tag in submittedSelectedTags"
                    :key="tag"
                    size="small"
                    closable
                    @close="toggleSubmittedTag(tag)"
                  >
                    {{ tag }}
                  </el-tag>
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
                  <div v-if="item.tags && item.tags.length > 0" class="search-result-item-tags">
                    <el-tag
                      v-for="tag in item.tags.slice(0, 5)"
                      :key="tag"
                      size="small"
                      effect="plain"
                    >
                      {{ tag }}
                    </el-tag>
                  </div>
                </button>
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
              :fragment="tab.fragment"
              :saved-fragment="tab.savedFragment"
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
    />
  </div>
</template>

<script>
import { Plus, Search } from '@element-plus/icons-vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'
import MemoryWelcome from '@/components/memory/MemoryWelcome.vue'
import MemoryEditor from '@/components/memory/MemoryEditor.vue'
import MemoryHistoryDialog from '@/components/memory/MemoryHistoryDialog.vue'

export default {
  name: 'MemoryFragment',
  components: {
    Plus,
    Search,
    MemoryWelcome,
    MemoryEditor,
    MemoryHistoryDialog,
  },
  data() {
    return {
      fragmentList: [],
      tagList: [],
      searchResults: [],
      searchQuery: '',
      searchMode: 'hybrid',
      selectedTags: [],
      searchTabVisible: false,
      submittedSearchQuery: '',
      submittedSearchMode: 'hybrid',
      submittedSelectedTags: [],
      activeTab: 'home',
      fragmentTabs: [],
      searchLoading: false,
      historyDialogVisible: false,
      historyFragmentId: 0,
    }
  },
  computed: {
    // activeFragmentId 返回当前激活的片段 id。
    activeFragmentId() {
      const tab = this.fragmentTabs.find(item => item.name === this.activeTab)
      return tab ? tab.fragment.id : 0
    },
    // searchTabLabel 返回搜索结果标签名称。
    searchTabLabel() {
      if (this.submittedSearchQuery.trim() !== '') {
        return `搜索结果: ${this.submittedSearchQuery}`
      }
      if (this.submittedSelectedTags.length > 0) {
        return `搜索结果: ${this.submittedSelectedTags.join('、')}`
      }
      return '搜索结果'
    }
  },
  mounted() {
    this.loadFragmentList()
    this.loadTagList()
  },
  methods: {
    // loadFragmentList 加载左侧片段列表。
    loadFragmentList() {
      MemoryFragmentApi.MemoryFragmentList(0, (response) => {
        this.fragmentList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    // loadTagList 加载标签筛选列表。
    loadTagList() {
      MemoryFragmentApi.MemoryFragmentTagList((response) => {
        this.tagList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    // runSearch 根据指定条件执行搜索。
    runSearch(query, mode, selectedTags) {
      this.searchLoading = true
      MemoryFragmentApi.MemoryFragmentSearch(
        query,
        mode,
        selectedTags,
        50,
        (response) => {
          this.searchLoading = false
          this.searchResults = Array.isArray(response.Data) ? response.Data : []
        }
      )
    },
    // submitSearch 提交当前搜索条件并打开搜索结果 tab。
    submitSearch() {
      if (this.searchQuery.trim() === '' && this.selectedTags.length === 0) {
        this.clearFilter()
        return
      }
      this.submittedSearchQuery = this.searchQuery.trim()
      this.submittedSearchMode = this.searchMode
      this.submittedSelectedTags = [...this.selectedTags]
      this.searchTabVisible = true
      this.activeTab = 'search'
      this.runSearch(this.submittedSearchQuery, this.submittedSearchMode, this.submittedSelectedTags)
    },
    // searchModeText 返回搜索模式中文描述。
    searchModeText(mode) {
      switch (mode) {
        case 'keyword':
          return '关键词'
        case 'natural':
          return '自然语言'
        default:
          return '混合'
      }
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
      this.runSearch(this.submittedSearchQuery, this.submittedSearchMode, this.submittedSelectedTags)
    },
    // handleSearchClear 处理搜索输入框清空。
    handleSearchClear() {
      this.searchQuery = ''
    },
    // clearFilter 清空左侧搜索条件，并关闭结果 tab。
    clearFilter() {
      this.searchQuery = ''
      this.searchMode = 'hybrid'
      this.selectedTags = []
      this.submittedSearchQuery = ''
      this.submittedSearchMode = 'hybrid'
      this.submittedSelectedTags = []
      this.searchTabVisible = false
      this.searchResults = []
      if (this.activeTab === 'search') {
        this.activeTab = 'home'
      }
    },
    // toggleTag 切换左侧待提交的标签筛选条件。
    toggleTag(tagName) {
      if (this.selectedTags.includes(tagName)) {
        this.selectedTags = this.selectedTags.filter(item => item !== tagName)
      } else {
        this.selectedTags = [...this.selectedTags, tagName]
      }
    },
    // toggleSubmittedTag 在搜索结果页中切换标签并重新搜索。
    toggleSubmittedTag(tagName) {
      if (this.submittedSelectedTags.includes(tagName)) {
        this.submittedSelectedTags = this.submittedSelectedTags.filter(item => item !== tagName)
      } else {
        this.submittedSelectedTags = [...this.submittedSelectedTags, tagName]
      }
      this.selectedTags = [...this.submittedSelectedTags]
      this.searchTabVisible = true
      this.activeTab = 'search'
      this.rerunSubmittedSearch()
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
        if (keyword === '' || keywordMap[keyword]) {
          return
        }
        keywordMap[keyword] = true
        keywords.push(keyword)
      })
      this.submittedSelectedTags.forEach((item) => {
        const keyword = item.trim()
        if (keyword === '' || keywordMap[keyword]) {
          return
        }
        keywordMap[keyword] = true
        keywords.push(keyword)
      })
      return keywords
    },
    // getSearchSnippetList 生成最多 3 条搜索命中片段。
    getSearchSnippetList(item) {
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
      MemoryFragmentApi.MemoryFragmentSave(0, '新记忆片段', '# 新记忆片段\n\n在这里开始记录。', [], (response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.loadFragmentList()
        this.loadTagList()
        this.upsertFragmentTab(response.Data, true)
      })
    },
    // openFragment 打开指定片段 tab。
    openFragment(fragmentId) {
      const existingTab = this.fragmentTabs.find(item => item.fragment.id === fragmentId)
      if (existingTab) {
        this.activeTab = existingTab.name
        return
      }
      MemoryFragmentApi.MemoryFragmentInfo(fragmentId, (response) => {
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.upsertFragmentTab(response.Data, true)
      })
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
        id: fragment.id,
        title: fragment.title || '',
        content: fragment.content || '',
        tags: Array.isArray(fragment.tags) ? [...fragment.tags] : [],
        index_status: fragment.index_status || 'pending',
        index_status_desc: fragment.index_status_desc || '待索引',
        update_time_desc: fragment.update_time_desc || '',
        create_time_desc: fragment.create_time_desc || '',
      }
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
      this.loadFragmentList()
      this.loadTagList()
      this.rerunSubmittedSearch()
    },
    // handleFragmentDeleted 删除片段后清理 tab 和列表。
    handleFragmentDeleted(fragmentId) {
      this.fragmentTabs = this.fragmentTabs.filter(item => item.fragment.id !== fragmentId)
      this.loadFragmentList()
      this.loadTagList()
      this.rerunSubmittedSearch()
      if (this.activeTab === `fragment-${fragmentId}`) {
        this.activeTab = 'home'
      }
    },
    // showHistory 打开历史记录弹窗。
    showHistory(fragmentId) {
      this.historyFragmentId = fragmentId
      this.historyDialogVisible = true
    },
    // closeTab 关闭一个编辑 tab 或搜索结果 tab。
    closeTab(tabName) {
      if (tabName === 'search') {
        this.searchTabVisible = false
        this.searchResults = []
        if (this.activeTab === 'search') {
          this.activeTab = 'home'
        }
        return
      }
      const targetIndex = this.fragmentTabs.findIndex(item => item.name === tabName)
      if (targetIndex < 0) {
        return
      }
      this.fragmentTabs.splice(targetIndex, 1)
      if (this.activeTab === tabName) {
        this.activeTab = this.fragmentTabs.length > 0 ? this.fragmentTabs[Math.max(targetIndex - 1, 0)].name : 'home'
      }
    },
    // handleTabChange 切换 tab 时保持页面状态一致。
    handleTabChange(tabPane) {
      this.activeTab = tabPane.paneName
    }
  }
}
</script>

<style scoped>
.memory-page {
  display: flex;
  gap: 14px;
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
  gap: 12px;
  padding: 16px;
  border-bottom: 1px solid #ecece4;
  background: #f7f7f2;
}

.sidebar-title {
  font-size: 15px;
  font-weight: 600;
  color: #4a5a45;
}

.sidebar-scroll {
  flex: 1;
}

.sidebar-item {
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

.sidebar-item:hover,
.sidebar-item.active {
  border-color: #cfe0c8;
  background: #f2f8ec;
}

.sidebar-item-main {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.sidebar-item-title {
  flex: 1;
  font-size: 14px;
  font-weight: 600;
  color: #32402f;
  line-height: 1.5;
}

.sidebar-item-time {
  color: #7b8576;
  font-size: 12px;
  white-space: nowrap;
}

.sidebar-item-tags {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  margin-top: 10px;
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
  gap: 12px;
  align-items: flex-start;
  margin-top: 14px;
}

.tag-filter-label {
  min-width: 64px;
  color: #60705a;
  font-size: 13px;
  line-height: 34px;
}

.tag-filter-list {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
}

.filter-chip {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  padding: 8px 12px;
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
  margin-top: 14px;
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
