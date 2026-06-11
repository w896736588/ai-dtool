﻿<template>
  <div class="memory-page">
    <aside v-if="memoryConfigured && !sidebarCollapsed" class="memory-sidebar">
      <div class="sidebar-header">
        <div class="sidebar-header-actions">
          <div class="sidebar-header-actions-row">
            <el-tooltip content="上传ZIP" placement="bottom">
              <pl-button plain size="small" class="icon-only-btn" @click="triggerUploadZip" :loading="zipUploading">
                <el-icon><Upload /></el-icon>
              </pl-button>
            </el-tooltip>
            <input ref="zipFileInput" type="file" accept=".zip" style="display:none" @change="handleZipUpload" />
            <el-tooltip content="搜索" placement="bottom">
              <pl-button plain size="small" class="icon-only-btn" @click="searchDialogVisible = true">
                <el-icon><Search /></el-icon>
              </pl-button>
            </el-tooltip>
            <el-tooltip content="新建" placement="bottom">
              <pl-button type="primary" plain size="small" class="icon-only-btn" @click="createFragment">
                <el-icon><Plus /></el-icon>
              </pl-button>
            </el-tooltip>
            <el-tooltip content="回收站" placement="bottom">
              <pl-button plain size="small" class="icon-only-btn" @click="openTrashTab">
                <el-icon><Delete /></el-icon>
              </pl-button>
            </el-tooltip>
            <el-tooltip content="设置" placement="bottom">
              <pl-button plain size="small" class="icon-only-btn" @click="openSettingsDialog">
                <el-icon><Setting /></el-icon>
              </pl-button>
            </el-tooltip>
          </div>
          <div class="sidebar-header-actions-row sidebar-header-actions-row--folders">
            <pl-button plain size="small" class="sidebar-header-folder-btn" @click="openFolderCreateDialog">
              <el-icon><FolderAdd /></el-icon>
              <span>新建文件夹</span>
            </pl-button>
            <pl-button plain size="small" class="sidebar-header-folder-btn" @click="openFolderManageDialog">
              文件夹管理
            </pl-button>
          </div>
        </div>
      </div>

      <el-scrollbar ref="sidebarScrollRef" v-show="!sidebarCollapsed" class="sidebar-scroll" @scroll="handleSidebarScroll">
        <div class="sidebar-filter-row">
          <el-input
            v-model="sidebarFilterQuery"
            clearable
            placeholder="过滤列表"
            size="small"
          >
            <template #prefix>
              <el-icon><Search /></el-icon>
            </template>
          </el-input>
          <el-select v-model="selectedFolderName" size="small" class="sidebar-folder-select">
            <el-option label="全部文件夹" value="" />
            <el-option
              v-for="folder in selectableFolders"
              :key="folder.folder_name"
              :label="folder.name"
              :value="folder.folder_name"
            />
          </el-select>
        </div>
        <div v-if="sidebarFilterLoading" class="sidebar-filter-loading">
          <el-icon class="is-loading"><Loading /></el-icon>
          <span>搜索中...</span>
        </div>
        <button
          v-for="item in filteredFragmentList"
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
          <div v-if="fragmentRefCount(item.id) > 0 || item.folder_label || item.folder_name" class="sidebar-item-refs">
            <div
              v-if="fragmentRefCount(item.id) > 0"
              class="sidebar-item-refs-summary"
              @click.stop="toggleFragmentRefs(item.id)"
            >
              <span class="sidebar-item-refs-badge">被 {{ fragmentRefCount(item.id) }} 个位置使用</span>
            </div>
            <el-dropdown
              class="sidebar-item-folder-dropdown"
              trigger="click"
              @command="(folderName) => changeFragmentFolder(item, folderName)"
            >
              <span class="sidebar-item-folder-tag">{{ item.folder_label || item.folder_name || '默认文件夹' }}</span>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item
                    v-for="folder in movableFolders"
                    :key="folder.folder_name"
                    :command="folder.folder_name"
                    :disabled="folder.folder_name === item.folder_name"
                  >
                    {{ folder.name }}
                  </el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
            <div
              v-if="fragmentRefCount(item.id) > 0 && expandedFragmentRefs[normalizeFragmentId(item.id)]"
              class="sidebar-item-refs-list"
            >
              <div
                v-for="ref in getFragmentRefs(item.id)"
                :key="ref.type + '-' + ref.id"
                class="sidebar-item-refs-item"
                @click.stop="openFragmentRef(ref)"
              >
                <span class="sidebar-item-refs-type">{{ ref.type === 'workflow' ? '工作流程' : '知识片段' }}</span>
                <span class="sidebar-item-refs-name">{{ ref.name || ref.title }}</span>
              </div>
            </div>
          </div>
          <div v-if="saveFeedbackMap[item.id]" class="sidebar-item-check" aria-hidden="true">
            <el-icon><Check /></el-icon>
          </div>
        </button>
        <div v-if="fragmentLoadingMore || !fragmentListHasMore" class="sidebar-load-status">
          <span v-if="fragmentLoadingMore" class="sidebar-load-loading">
            <el-icon class="is-loading"><Loading /></el-icon>
            <span>加载中...</span>
          </span>
          <span v-else-if="!fragmentListHasMore && fragmentList.length > 0" class="sidebar-load-nomore">没有更多了</span>
        </div>
      </el-scrollbar>

      <div v-if="!sidebarCollapsed" class="sidebar-footer">
        <el-tooltip content="返回首页" placement="right">
          <button class="memory-home-btn" @click="goHome">
            <el-icon :size="13"><HomeFilled /></el-icon>
          </button>
        </el-tooltip>
        <span v-if="fragmentTotalCount > 0" class="sidebar-count-badge">{{ fragmentList.length }}/{{ fragmentTotalCount }}</span>
      </div>
    </aside>

    <button v-if="memoryConfigured && !isEmbedded" class="sidebar-collapse-btn" :title="sidebarCollapsed ? '展开列表' : '收起列表'" @click="toggleSidebar">
      <el-icon :size="12"><component :is="sidebarCollapsed ? 'DArrowRight' : 'DArrowLeft'" /></el-icon>
    </button>

    <section class="memory-main">
      <div class="workspace-card">
        <el-tabs
          v-model="activeTab"
          type="card"
          closable
          :class="['memory-tabs', { 'memory-tabs--embedded': isEmbedded }]"
          @tab-remove="closeTab"
          @tab-click="handleTabChange"
        >
          <el-tab-pane
            name="home"
            :closable="false"
          >
            <template #label>
              <span class="tab-label">首页</span>
            </template>
            <div class="memory-home-panel">
              <div class="memory-home-hero">
                <div class="memory-home-hero__main">
                  <div class="memory-home-hero__eyebrow">Knowledge Home</div>
                  <div class="memory-home-hero__title">知识片段文件夹总览</div>
                  <div class="memory-home-hero__desc">
                    当前共 {{ folderOverviewList.length }} 个文件夹，收录 {{ totalFolderFragmentCount }} 条知识片段。
                  </div>
                </div>
                <div class="memory-home-hero__stats">
                  <div class="memory-home-stat">
                    <div class="memory-home-stat__label">文件夹</div>
                    <div class="memory-home-stat__value">{{ folderOverviewList.length }}</div>
                  </div>
                  <div class="memory-home-stat">
                    <div class="memory-home-stat__label">知识片段</div>
                    <div class="memory-home-stat__value">{{ totalFolderFragmentCount }}</div>
                  </div>
                </div>
              </div>

              <div v-if="folderOverviewList.length > 0" class="memory-home-grid">
                <button
                  v-for="folder in folderOverviewList"
                  :key="folder.folder_name"
                  class="memory-home-folder-card"
                  @click="enterFolder(folder.folder_name)"
                >
                  <div class="memory-home-folder-card__head">
                    <div class="memory-home-folder-card__title">{{ folder.name || folder.folder_name }}</div>
                    <div class="memory-home-folder-card__count">{{ folder.fragment_count || 0 }}</div>
                  </div>
                  <div class="memory-home-folder-card__path">{{ folder.folder_name }}</div>
                  <div class="memory-home-folder-card__meta">
                    <span>{{ folder.fragment_count || 0 }} 个知识片段</span>
                    <span v-if="Number(folder.system) === 1">系统</span>
                    <span v-else>自定义</span>
                  </div>
                </button>
              </div>

              <el-empty
                v-else
                description="暂无文件夹"
              />
            </div>
          </el-tab-pane>

          <el-tab-pane
            v-if="aiSearchTabVisible"
            name="ai_search"
          >
            <template #label>
              <span class="tab-label">AI 搜索: {{ aiSearchQuery }}</span>
            </template>
            <div class="ai-search-panel">
              <div class="ai-search-timeline">
                <div v-for="(event, index) in aiSearchEvents" :key="index"
                  :class="['ai-search-step', `ai-search-step--` + event.status]"
                >
                  <div class="ai-search-step-row">
                    <div class="ai-search-step-icon">
                      <el-icon v-if="event.status === 'running'" class="is-loading"><Loading /></el-icon>
                      <el-icon v-else-if="event.status === 'done'" class="ai-search-step-done-icon"><Check /></el-icon>
                      <el-icon v-else-if="event.status === 'error'" class="ai-search-step-error-icon"><Close /></el-icon>
                    </div>
                    <div class="ai-search-step-content">
                      <span v-if="event.status === 'running'" class="ai-search-step-message">
                        {{ event.message || getStepLabel(event.step) }}（已用时 {{ aiSearchStepElapsed[event.step] || 0 }} 秒）...
                      </span>
                      <span v-else-if="event.status === 'done'" class="ai-search-step-message ai-search-step-done-text">
                        {{ getStepDoneText(event) }}
                      </span>
                      <span v-else-if="event.status === 'error'" class="ai-search-step-message ai-search-step-error-text">
                        {{ event.message }}
                      </span>
                    </div>
                    <button
                      v-if="event.status === 'done' && canExpandStep(event)"
                      class="ai-search-step-expand-btn"
                      @click="toggleStepExpand(event.step)"
                    >
                      <el-icon :size="12">
                        <component :is="aiSearchExpandedSteps[event.step] ? 'ArrowDown' : 'ArrowRight'" />
                      </el-icon>
                    </button>
                  </div>
                  <div v-if="event.step === 'read' && event.status === 'running' && event.data" class="ai-search-step-progress">
                    读取中 {{ event.data.current }}/{{ event.data.total }}：{{ event.data.title }}
                  </div>
                  <div v-if="aiSearchExpandedSteps[event.step] && event.status === 'done'" class="ai-search-step-detail-panel">
                    <div v-if="event.step === 'keywords'" class="ai-search-detail-section">
                      <div class="ai-search-detail-label">提示词</div>
                      <pre class="ai-search-detail-code">{{ event.prompt }}</pre>
                      <div class="ai-search-detail-label">AI 回复</div>
                      <pre class="ai-search-detail-code">{{ event.response }}</pre>
                      <div v-if="event.data && event.data.keywords" class="ai-search-detail-label">解析出的关键词</div>
                      <div v-if="event.data && event.data.keywords" class="ai-search-detail-keywords">
                        <span v-for="kw in event.data.keywords" :key="kw" class="ai-search-detail-keyword-chip">{{ kw }}</span>
                      </div>
                    </div>
                    <div v-if="event.step === 'search'" class="ai-search-detail-section">
                      <div class="ai-search-detail-label">搜索关键词</div>
                      <pre class="ai-search-detail-code">{{ event.prompt }}</pre>
                      <div v-if="event.data && event.data.fragments && event.data.fragments.length > 0" class="ai-search-detail-label">
                        找到的片段（{{ event.data.fragments.length }} 个）
                      </div>
                      <div v-if="event.data && event.data.fragments" class="ai-search-detail-fragments">
                        <a v-for="frag in event.data.fragments" :key="frag.id" class="ai-search-detail-fragment-link"
                          href="javascript:void(0)" @click="openFragment(frag.id)"
                        >{{ frag.title || '未命名片段' }}</a>
                      </div>
                    </div>
                    <div v-if="event.step === 'judge'" class="ai-search-detail-section">
                      <div class="ai-search-detail-label">提示词</div>
                      <pre class="ai-search-detail-code">{{ event.prompt }}</pre>
                      <div class="ai-search-detail-label">AI 回复</div>
                      <pre class="ai-search-detail-code">{{ event.response }}</pre>
                      <div v-if="event.data && event.data.selected_fragments" class="ai-search-detail-label">
                        选中的片段（{{ event.data.selected_fragments.length }} 个）
                      </div>
                      <div v-if="event.data && event.data.selected_fragments" class="ai-search-detail-fragments">
                        <a v-for="frag in event.data.selected_fragments" :key="frag.id" class="ai-search-detail-fragment-link"
                          href="javascript:void(0)" @click="openFragment(frag.id)"
                        >{{ frag.title || '未命名片段' }}</a>
                      </div>
                    </div>
                    <div v-if="event.step === 'read'" class="ai-search-detail-section">
                      <div v-if="event.data && event.data.read_fragments" class="ai-search-detail-label">
                        已读取片段（{{ event.data.read_fragments.length }} 个，共 {{ event.data.total_chars || 0 }} 字）
                      </div>
                      <div v-if="event.data && event.data.read_fragments" class="ai-search-detail-fragments">
                        <a v-for="frag in event.data.read_fragments" :key="frag.id" class="ai-search-detail-fragment-link"
                          href="javascript:void(0)" @click="openFragment(frag.id)"
                        >{{ frag.title || '未命名片段' }}</a>
                      </div>
                    </div>
                    <div v-if="event.step === 'answer'" class="ai-search-detail-section">
                      <div class="ai-search-detail-label">用户问题</div>
                      <pre class="ai-search-detail-code">{{ event.prompt }}</pre>
                    </div>
                  </div>
                </div>
              </div>
              <div v-if="aiSearchAnswer" class="ai-search-answer">
                <div class="ai-search-answer-header">搜索结果</div>
                <div class="ai-search-answer-content markdown-body" v-html="renderMarkdown(aiSearchAnswer)"></div>
              </div>
              <div v-if="aiSearchReferencedFragments.length > 0 && !aiSearchLoading" class="ai-search-references">
                <div class="ai-search-references-title">相关片段</div>
                <div v-for="ref in aiSearchReferencedFragments" :key="ref.id"
                  class="ai-search-reference-item"
                >
                  <a href="javascript:void(0)" @click="openFragment(ref.id)">{{ ref.title || '未命名片段' }}</a>
                </div>
              </div>
              <div v-if="aiSearchLoading && !aiSearchAnswer" v-loading="true" class="ai-search-loading"></div>
            </div>
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
              <span class="tab-label" :title="tab.fragment.title || '未命名片段'">
                {{ truncateTabLabel(tab.fragment.title) }}<span v-if="tab.dirty"> *</span>
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
      :git-repo-enabled="memoryIsGitRepo"
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

    <el-dialog
      v-model="searchDialogVisible"
      title="搜索知识片段"
      width="580px"
      :close-on-click-modal="true"
      destroy-on-close
      class="search-dialog"
    >
      <div class="search-dialog-body">
        <el-input
          v-model="searchQuery"
          type="textarea"
          :autosize="{ minRows: 4, maxRows: 10 }"
          :placeholder="searchPlaceholder"
          @keydown.enter.prevent="submitSearchFromDialog"
        />
        <div class="search-dialog-mode-row">
          <span class="search-dialog-mode-label">搜索模式</span>
          <el-switch
            v-model="searchModeSemantic"
            active-text="智能搜索"
            inactive-text="全文搜索"
            class="search-mode-switch"
            @change="handleSearchModeSwitch"
          />
        </div>
        <div class="search-dialog-actions">
          <pl-button type="primary" @click="submitSearchFromDialog">
            <el-icon><Search /></el-icon>
            搜索
          </pl-button>
          <pl-button plain @click="clearFilterAndCloseDialog">清空条件</pl-button>
        </div>
      </div>
    </el-dialog>

    <el-dialog
      v-model="folderDialogVisible"
      title="创建文件夹"
      width="420px"
      destroy-on-close
    >
      <el-form label-position="top">
        <el-form-item label="名称">
          <el-input v-model="folderForm.name" maxlength="50" />
        </el-form-item>
        <el-form-item label="文件夹名称">
          <el-input v-model="folderForm.folder_name" maxlength="50" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="folderDialogVisible = false">取消</el-button>
        <el-button type="primary" @click="submitFolderCreate">创建</el-button>
      </template>
    </el-dialog>

    <el-dialog
      v-model="folderManageDialogVisible"
      title="文件夹管理"
      width="520px"
      destroy-on-close
    >
      <div v-if="folderManageDrafts.length > 0" class="folder-manage-list">
        <div
          v-for="folder in folderManageDrafts"
          :key="folder.folder_name"
          class="folder-manage-item"
        >
          <div class="folder-manage-item__meta">
            <div class="folder-manage-item__title">{{ folder.folder_name }}</div>
            <div class="folder-manage-item__label">显示名称</div>
          </div>
          <div class="folder-manage-item__actions">
            <el-input v-model="folder.name" maxlength="50" />
            <el-button
              type="primary"
              :disabled="!folder.folder_name || !folder.name || folder.name === folder.original_name"
              @click="submitFolderRename(folder)"
            >
              保存
            </el-button>
          </div>
        </div>
      </div>
      <el-empty v-else description="暂无可管理的文件夹" :image-size="88" />
      <template #footer>
        <el-button @click="folderManageDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import { ArrowDown, ArrowRight, Check, Close, DArrowLeft, DArrowRight, Delete, Download, FolderAdd, HomeFilled, Loading, Plus, Search, Setting, Upload } from '@element-plus/icons-vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'
import set from '@/utils/base/git_set'
import MemoryEditor from '@/components/memory/MemoryEditor.vue'
import MemoryHistoryDialog from '@/components/memory/MemoryHistoryDialog.vue'
import MemorySettingPage from '@/components/set/memory.vue'
import GitActionButton from '@/components/base/GitActionButton.vue'
import SettingsDialog from '@/components/base/SettingsDialog.vue'
import sseDistribute from '@/utils/base/sse_distribute'
import base from '@/utils/base'
import { marked } from 'marked'
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
// HOME_TAB_NAME 统一定义知识片段首页标签页名称。
const HOME_TAB_NAME = 'home'
// TRASH_TAB_NAME 统一定义回收站标签页名称，避免散落硬编码。
const TRASH_TAB_NAME = 'trash'
// AI_SEARCH_TAB_NAME 统一定义 AI 智能搜索标签页名称。
const AI_SEARCH_TAB_NAME = 'ai_search'
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
    ArrowDown,
    ArrowRight,
    Check,
    Close,
    DArrowLeft,
    DArrowRight,
    Delete,
    Download,
    FolderAdd,
    HomeFilled,
    Plus,
    Search,
    Setting,
    GitActionButton,
    MemoryEditor,
    MemoryHistoryDialog,
    MemorySettingPage,
    SettingsDialog,
  },
  data() {
    return {
      fragmentList: [],
      fragmentPageSize: 10,
      fragmentOffset: 0,
      fragmentHasMore: true,
      fragmentLoadingMore: false,
      fragmentTotalCount: 0,
      trashList: [],
      searchResults: [],
      searchQuery: '',
      searchMode: KEYWORD_SEARCH_MODE,
      searchModeSemantic: false,
      searchDialogVisible: false,
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
      sidebarCollapsed: localStorage.getItem('memorySidebarCollapsed') === 'true',
      sidebarFilterQuery: '',
      sidebarFilterTimer: null,
      sidebarFilterLoading: false,
      sidebarFilterResults: [],
      aiSearchTabVisible: false,
      aiSearchQuery: '',
      aiSearchEvents: [],
      aiSearchAnswer: '',
      aiSearchLoading: false,
      aiSearchSseClient: null,
      aiSearchReferencedFragments: [],
      aiSearchStepElapsed: {},
      aiSearchExpandedSteps: {},
      aiSearchStepTimerId: null,
      zipUploading: false,
      fragmentReferences: {},
      expandedFragmentRefs: {},
      folderList: [],
      selectedFolderName: 'fragments',
      folderDialogVisible: false,
      folderManageDialogVisible: false,
      folderForm: {
        name: '',
        folder_name: '',
      },
      folderManageDrafts: [],
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
      routeFolderName() {
        return String(this.$route.query.folder_name || '').trim()
      },
      isEmbedded() {
        return String(this.$route.query.embed || '') === '1'
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
    // searchPlaceholder 根据搜索模式返回对应的输入框提示文本。
    // searchPlaceholder 根据搜索模式返回对应的输入框提示文本。
    searchPlaceholder() {
      if (this.searchMode === 'semantic') {
        return '输入想要查询的内容，自然语言描述，回车打开结果页'
      }
      return '输入关键词，多个关键词使用空格分隔，回车打开结果页'
    },
    // filteredFragmentList 侧边栏过滤结果，由 watch sidebarFilterQuery 驱动。
    filteredFragmentList() {
      if (!this.sidebarFilterQuery.trim()) return this.fragmentList
      return this.sidebarFilterResults
    },
    // fragmentListHasMore 侧边栏列表是否还有更多数据可加载。
      fragmentListHasMore() {
        if (this.sidebarFilterQuery.trim()) return false
        return this.fragmentHasMore
      },
      folderOverviewList() {
        return this.selectableFolders
      },
      totalFolderFragmentCount() {
        return this.folderOverviewList.reduce((total, item) => total + Number(item.fragment_count || 0), 0)
      },
      selectableFolders() {
        return this.folderList.filter(item => item.folder_name !== 'trash')
      },
    movableFolders() {
      return this.folderList.filter(item => item.folder_name !== 'trash')
    },
    editableFolders() {
      return this.folderList.filter(item => Number(item.editable) > 0)
    },
    manageableFolders() {
      return this.editableFolders.filter(item => item.folder_name !== 'fragments' && item.folder_name !== 'trash')
    }
  },
  created() {
    this.applySidebarHiddenQuery()
    this.syncFolderFromRoute()
  },
  mounted() {
    this.aiSearchStepStartTimes = {}
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
    this.stopAiSearchSse()
    this.clearAllStepTimers()
    this.clearSaveFeedbackTimers()
  },
    watch: {
      '$route.fullPath'() {
        this.routeFragmentHandled = false
        this.syncFolderFromRoute()
        this.tryOpenRouteFragmentOnEntry()
        this.applySidebarHiddenQuery()
      },
    selectedFolderName() {
      this.loadFragmentList()
      this.rerunSidebarFilter()
      this.rerunSubmittedSearch()
    },
    sidebarFilterQuery() {
      clearTimeout(this.sidebarFilterTimer)
      const query = this.sidebarFilterQuery.trim()
      if (!query) {
        this.sidebarFilterResults = []
        this.sidebarFilterLoading = false
        return
      }
      this.sidebarFilterLoading = true
      this.sidebarFilterTimer = setTimeout(() => {
        MemoryFragmentApi.MemoryFragmentSearch(
          this.sidebarFilterQuery.trim(),
          KEYWORD_SEARCH_MODE,
          [],
          this.selectedFolderName,
          0,
          (response) => {
            this.sidebarFilterLoading = false
            this.sidebarFilterResults = Array.isArray(response.Data) ? response.Data : []
          }
        )
      }, 300)
    },
  },
  methods: {
    // truncateTabLabel 截断tab标签，最多显示maxWidth个字符宽度（中文算2，英文算1）。
    truncateTabLabel(text, maxWidth = 15) {
      if (!text) return '未命名片段'
      let width = 0
      for (let i = 0; i < text.length; i++) {
        width += text.charCodeAt(i) > 127 ? 2 : 1
        if (width > maxWidth) return text.slice(0, i) + '…'
      }
      return text
    },
    // toggleSidebar 切换左侧列表的折叠/展开状态。
    toggleSidebar() {
      this.sidebarCollapsed = !this.sidebarCollapsed
      localStorage.setItem('memorySidebarCollapsed', this.sidebarCollapsed)
    },
    syncFolderFromRoute() {
      const routeFolderName = this.routeFolderName
      this.selectedFolderName = routeFolderName || 'fragments'
    },
    loadFolderList() {
      MemoryFragmentApi.MemoryFragmentFolderList((response) => {
        if (!(response && response.ErrCode === 0)) {
          return
        }
        this.folderList = Array.isArray(response.Data) ? response.Data : []
        const exists = this.folderList.some(item => item.folder_name === this.selectedFolderName)
        if (!exists && this.selectedFolderName) {
          this.selectedFolderName = 'fragments'
        }
        if (this.routeFolderName && this.folderList.some(item => item.folder_name === this.routeFolderName)) {
          this.selectedFolderName = this.routeFolderName
        }
      })
    },
    openFolderCreateDialog() {
      this.folderForm = {
        name: '',
        folder_name: '',
      }
      this.folderDialogVisible = true
    },
    openFolderManageDialog() {
      this.folderManageDrafts = this.manageableFolders.map(item => ({
        folder_name: item.folder_name,
        name: item.name || '',
        original_name: item.name || '',
      }))
      this.folderManageDialogVisible = true
    },
    submitFolderCreate() {
      MemoryFragmentApi.MemoryFragmentFolderCreate(this.folderForm.name, this.folderForm.folder_name, (response) => {
        if (!(response && response.ErrCode === 0)) {
          return
        }
        this.folderDialogVisible = false
        this.loadFolderList()
        if (response.Data && response.Data.folder_name) {
          this.selectedFolderName = response.Data.folder_name
        }
      })
    },
    submitFolderRename(folder) {
      if (!folder || !folder.folder_name || !folder.name) {
        return
      }
      MemoryFragmentApi.MemoryFragmentFolderUpdate(folder.folder_name, folder.name, (response) => {
        if (!(response && response.ErrCode === 0)) {
          return
        }
        folder.original_name = folder.name
        this.loadFolderList()
      })
    },
    changeFragmentFolder(item, folderName) {
      if (!item || !folderName) {
        return
      }
      MemoryFragmentApi.MemoryFragmentFolderChange(item.id || item.file_id, folderName, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          return
        }
        this.upsertFragmentTab(response.Data, false)
        this.loadFragmentList()
        this.rerunSidebarFilter()
        this.rerunSubmittedSearch()
      })
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
      this.memoryIsGitRepo = !!(data && data.is_git_repo)
      this.nextPushTime = data && data.next_push_time ? Number(data.next_push_time) : 0
      this.lastPushTime = data && data.last_push_time ? Number(data.last_push_time) : 0
      this.lastPushTimeDesc = data && data.last_push_time_desc ? data.last_push_time_desc : '-'
      this.lastPushError = data && data.last_push_error ? data.last_push_error : ''
      this.fragmentTotalCount = data && data.fragment_count ? Number(data.fragment_count) : 0
      if (!this.memoryConfigured) {
        this.fragmentList = []
        this.folderList = []
        this.trashList = []
        this.searchResults = []
        this.fragmentTabs = []
        this.activeTab = ''
        this.memoryIsGitRepo = false
        this.nextPushTime = 0
        this.lastPushTime = 0
        return
      }
      // 首次加载时需要加载列表
      if (this.fragmentList.length === 0 && this.trashList.length === 0) {
        this.loadFolderList()
        this.loadFragmentList()
        this.loadTrashList()
      }
      this.tryOpenRouteFragmentOnEntry()
    },
    // handleMemoryFragmentSseUpdate 处理来自其他页面或异步任务的知识片段变更。
    handleMemoryFragmentSseUpdate(payload) {
      const fragmentId = this.normalizeFragmentId(payload?.fragment_id || payload?.fragment?.id || payload?.fragment?.file_id)
      const action = String(payload?.action || '').trim()
      if (!fragmentId) {
        // 无法定点更新时才全量刷新作为兜底。
        this.loadFragmentList()
        this.loadTrashList()
        this.rerunSubmittedSearch()
        this.rerunSidebarFilter()
        return
      }
      if (action === MEMORY_FRAGMENT_SSE_ACTION_DELETE) {
        this.handleRemoteDeletedFragment(fragmentId)
        this.loadTrashList()
        this.rerunSidebarFilter()
        return
      }
      if (action !== MEMORY_FRAGMENT_SSE_ACTION_UPSERT) {
        this.loadFragmentList()
        this.loadTrashList()
        this.rerunSubmittedSearch()
        this.rerunSidebarFilter()
        return
      }
      // upsert：定点更新列表项，若文件夹变更则从当前列表移除，无需全量刷新。
      this.handleRemoteUpsertFragment(fragmentId, payload?.fragment || null)
      this.applySseUpsertToList(fragmentId, payload?.fragment || null)
      this.loadTrashList()
    },
    // handleRemoteDeletedFragment 同步处理远端删除的片段。
    handleRemoteDeletedFragment(fragmentId) {
      this.fragmentList = this.fragmentList.filter(item => this.normalizeFragmentId(item.id || item.file_id) !== fragmentId)
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
        // 中文注释：当 SSE 携带的片段内容与本地草稿一致时，说明是本地保存操作的 SSE 回传，
        // 此时静默同步即可，避免竞态条件导致误弹警告。
        // English comment: When SSE content matches the local draft, it's our own save being echoed back.
        // Silently sync instead of warning to avoid a race-condition false alarm.
        if (fragment && typeof fragment === 'object' && fragment.content !== undefined) {
          const localContent = (targetTab.fragment.content || '').trim()
          const remoteContent = (fragment.content || '').trim()
          if (localContent === remoteContent) {
            this.upsertFragmentTab(fragment, false)
            return
          }
        }
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
        this.memoryIsGitRepo = !!(response.Data && response.Data.is_git_repo)
        this.nextPushTime = response.Data && response.Data.next_push_time ? Number(response.Data.next_push_time) : 0
        this.lastPushTime = response.Data && response.Data.last_push_time ? Number(response.Data.last_push_time) : 0
        this.lastPushTimeDesc = response.Data && response.Data.last_push_time_desc ? response.Data.last_push_time_desc : '-'
        this.lastPushError = response.Data && response.Data.last_push_error ? response.Data.last_push_error : ''
        this.fragmentTotalCount = response.Data && response.Data.fragment_count ? Number(response.Data.fragment_count) : 0
        if (!this.memoryConfigured) {
          this.fragmentList = []
          this.folderList = []
          this.trashList = []
          this.searchResults = []
          this.fragmentTabs = []
          this.activeTab = ''
          this.memoryIsGitRepo = false
          this.nextPushTime = 0
          this.lastPushTime = 0
          return
        }
        if (needReloadLists) {
          this.loadFolderList()
          this.loadFragmentList()
          this.loadTrashList()
        }
        this.tryOpenRouteFragmentOnEntry()
      })
    },
    // loadFragmentList 加载左侧片段列表，append 为 true 时追加到现有列表。
    loadFragmentList(append = false) {
      if (!this.memoryConfigured) {
        return
      }
      if (this.fragmentLoadingMore) {
        return
      }
      if (append && !this.fragmentHasMore) {
        return
      }
      const offset = append ? this.fragmentOffset : 0
      if (!append) {
        this.fragmentOffset = 0
        this.fragmentHasMore = true
      }
      this.fragmentLoadingMore = true
      MemoryFragmentApi.MemoryFragmentList(this.fragmentPageSize, offset, this.selectedFolderName, (response) => {
        this.fragmentLoadingMore = false
        const list = Array.isArray(response.Data && response.Data.list) ? response.Data.list : (Array.isArray(response.Data) ? response.Data : [])
        const hasMore = typeof response.Data === 'object' && response.Data !== null && 'has_more' in response.Data ? response.Data.has_more : false
        if (append) {
          this.fragmentList = this.fragmentList.concat(list)
        } else {
          this.fragmentList = list
        }
        this.fragmentOffset = offset + list.length
        this.fragmentHasMore = hasMore
        this.ensureDefaultFragmentTab()
        // 批量查询当前列表片段的引用信息。
        const ids = list.map(item => this.normalizeFragmentId(item.id || item.file_id)).filter(Boolean)
        this.loadFragmentReferences(ids)
      })
    },
    // loadFragmentListPreservingOrder 重置分页并加载最新数据，保持侧边栏列表的原有顺序。
    loadFragmentListPreservingOrder() {
      if (!this.memoryConfigured) {
        return
      }
      const currentOrderIds = this.fragmentList.map(item => this.normalizeFragmentId(item.id || item.file_id))
      MemoryFragmentApi.MemoryFragmentList(this.fragmentPageSize, 0, this.selectedFolderName, (response) => {
        const rawList = Array.isArray(response.Data && response.Data.list) ? response.Data.list : (Array.isArray(response.Data) ? response.Data : [])
        const hasMore = typeof response.Data === 'object' && response.Data !== null && 'has_more' in response.Data ? response.Data.has_more : false
        const newMap = new Map(rawList.map(item => [this.normalizeFragmentId(item.id || item.file_id), item]))
        const ordered = []
        currentOrderIds.forEach(id => {
          const item = newMap.get(id)
          if (item) {
            ordered.push(item)
            newMap.delete(id)
          }
        })
        newMap.forEach(item => ordered.push(item))
        this.fragmentList = ordered
        this.fragmentOffset = ordered.length
        this.fragmentHasMore = hasMore
        this.ensureDefaultFragmentTab()
      })
    },
    // loadMoreFragments 上拉加载更多片段。
    loadMoreFragments() {
      this.loadFragmentList(true)
    },
    // handleSidebarScroll 监听侧边栏滚动事件，到达底部时触发加载更多。
    handleSidebarScroll({ scrollTop }) {
      const scrollbarEl = this.$refs.sidebarScrollRef
      if (!scrollbarEl) {
        return
      }
      const wrap = scrollbarEl.wrapRef
      if (!wrap) {
        return
      }
      const distanceToBottom = wrap.scrollHeight - wrap.clientHeight - scrollTop
      if (distanceToBottom < 60) {
        this.loadMoreFragments()
      }
    },
    // ensureDefaultFragmentTab 在没有激活片段时自动打开列表中的第一个知识片段。
    ensureDefaultFragmentTab() {
      if (!this.memoryConfigured) {
        return
      }
      if (!this.routeFragmentId) {
        // 如果当前已经在查看某个知识片段 tab，不要强行切回首页，避免打断用户操作。
        if (isMemoryFragmentTabName(this.activeTab)) {
          return
        }
        this.activeTab = HOME_TAB_NAME
        return
      }
      if (this.activeTab === SEARCH_TAB_NAME || this.activeTab === TRASH_TAB_NAME || this.activeTab === AI_SEARCH_TAB_NAME) {
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
        this.selectedFolderName,
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
      if (this.searchMode === SEMANTIC_SEARCH_MODE) {
        this.openAiSearchTab(this.searchQuery.trim())
        return
      }
      this.submittedSearchQuery = this.searchQuery.trim()
      this.submittedSearchMode = this.searchMode
      this.searchTabVisible = true
      this.activeTab = SEARCH_TAB_NAME
      this.runSearch(this.submittedSearchQuery, this.submittedSearchMode)
    },
    // submitSearchFromDialog 从弹窗提交搜索，搜索后关闭弹窗。
    submitSearchFromDialog() {
      this.searchDialogVisible = false
      this.submitSearch()
    },
    // clearFilterAndCloseDialog 清空搜索条件并关闭弹窗。
    clearFilterAndCloseDialog() {
      this.clearFilter()
      this.searchDialogVisible = false
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
    // rerunSidebarFilter 重新执行侧边栏过滤搜索。
    rerunSidebarFilter() {
      const query = this.sidebarFilterQuery.trim()
      if (!query) return
      this.sidebarFilterLoading = true
      MemoryFragmentApi.MemoryFragmentSearch(
        query,
        KEYWORD_SEARCH_MODE,
        [],
        this.selectedFolderName,
        0,
        (response) => {
          this.sidebarFilterLoading = false
          this.sidebarFilterResults = Array.isArray(response.Data) ? response.Data : []
        }
      )
    },
    // handleSearchClear 处理搜索输入框清空。
    handleSearchClear() {
      this.searchQuery = ''
    },
    // handleSearchModeSwitch 处理搜索模式 switch 切换。
    handleSearchModeSwitch(semantic) {
      this.searchMode = semantic ? SEMANTIC_SEARCH_MODE : KEYWORD_SEARCH_MODE
    },
    // clearFilter 清空左侧搜索条件，并关闭结果 tab。
    clearFilter() {
      this.searchQuery = ''
      this.searchMode = KEYWORD_SEARCH_MODE
      this.searchModeSemantic = false
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
    // triggerUploadZip 触发隐藏的文件选择框。
    triggerUploadZip() {
      if (!this.memoryConfigured || this.zipUploading) {
        return
      }
      this.$refs.zipFileInput.click()
    },
    // handleZipUpload 处理 ZIP 文件上传，成功后创建片段并打开。
    handleZipUpload(event) {
      const file = event.target.files[0]
      if (!file) {
        return
      }
      this.zipUploading = true
      const apiBaseURL = base.GetAbsoluteApiHost()
      MemoryFragmentApi.MemoryFragmentUploadZip(file, apiBaseURL, (response) => {
        this.zipUploading = false
        // 重置 input，允许重复选择同一文件
        event.target.value = ''
        if (response.ErrCode !== 0 || !response.Data) {
          return
        }
        this.loadFragmentList()
        this.upsertFragmentTab(response.Data, true)
      })
    },
    createFragment() {
      if (!this.memoryConfigured) {
        return
      }
      MemoryFragmentApi.MemoryFragmentSave(0, '新知识片段', '# 标签\n\n在这里开始记录。', [], this.selectedFolderName || 'fragments', (response) => {
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
        this.activeTab = HOME_TAB_NAME
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
      const existingTab = this.fragmentTabs.find(item => item.name === tabName)
      if (existingTab) {
        // 原地更新现有 tab，避免 splice 导致 el-tabs 检测到当前 tab 暂时消失而重置 activeTab。
        existingTab.fragment = normalized
        existingTab.savedFragment = this.cloneFragment(normalized)
        existingTab.dirty = false
      } else {
        this.fragmentTabs.push({
          name: tabName,
          fragment: normalized,
          savedFragment: this.cloneFragment(normalized),
          dirty: false,
        })
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
        folder_name: fragment.folder_name || 'fragments',
        folder_label: fragment.folder_label || '',
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
    // applySseUpsertToList SSE upsert 事件时定点更新侧边栏列表，处理文件夹变更的移除逻辑。
    applySseUpsertToList(fragmentId, fragment) {
      if (!fragmentId) return
      const currentFolder = this.selectedFolderName
      // 若携带了 fragment 信息，判断文件夹是否变更。
      if (fragment && typeof fragment === 'object' && Object.keys(fragment).length > 0) {
        const newFolderName = fragment.folder_name || 'fragments'
        // 如果当前筛选了特定文件夹，且片段不再属于该文件夹，从列表中移除。
        if (currentFolder && newFolderName !== currentFolder) {
          const idx = this.fragmentList.findIndex(item => this.normalizeFragmentId(item.id || item.file_id) === fragmentId)
          if (idx !== -1) {
            this.fragmentList.splice(idx, 1)
          }
          const filterIdx = this.sidebarFilterResults.findIndex(item => this.normalizeFragmentId(item.id || item.file_id) === fragmentId)
          if (filterIdx !== -1) {
            this.sidebarFilterResults.splice(filterIdx, 1)
          }
          // 更新首页文件夹统计。
          this.loadFolderList()
          return
        }
        // 文件夹未变，定点更新。
        this.updateFragmentListItem(fragment)
        return
      }
      // SSE 未携带 fragment 详情，用 API 获取后定点更新。
      MemoryFragmentApi.MemoryFragmentInfo(fragmentId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) return
        const frag = response.Data
        const newFolderName = frag.folder_name || 'fragments'
        if (currentFolder && newFolderName !== currentFolder) {
          const idx = this.fragmentList.findIndex(item => this.normalizeFragmentId(item.id || item.file_id) === fragmentId)
          if (idx !== -1) {
            this.fragmentList.splice(idx, 1)
          }
          const filterIdx = this.sidebarFilterResults.findIndex(item => this.normalizeFragmentId(item.id || item.file_id) === fragmentId)
          if (filterIdx !== -1) {
            this.sidebarFilterResults.splice(filterIdx, 1)
          }
          this.loadFolderList()
          return
        }
        this.updateFragmentListItem(frag)
      })
    },
    // updateFragmentListItem 定点更新侧边栏列表中指定片段的信息，避免全量刷新导致 tab 切换。
    updateFragmentListItem(fragment) {
      const normalizedId = this.normalizeFragmentId(fragment.id || fragment.file_id)
      if (!normalizedId) return
      const fields = {
        title: fragment.title,
        file_path: fragment.file_path,
        folder_name: fragment.folder_name,
        folder_label: fragment.folder_label,
        update_time_desc: fragment.update_time_desc,
        create_time_desc: fragment.create_time_desc,
      }
      const patch = (list) => {
        const idx = list.findIndex(item => this.normalizeFragmentId(item.id || item.file_id) === normalizedId)
        if (idx !== -1) {
          const existing = list[idx]
          const merged = { ...existing }
          for (const key of Object.keys(fields)) {
            if (fields[key] !== undefined && fields[key] !== '') {
              merged[key] = fields[key]
            }
          }
          list.splice(idx, 1, merged)
        }
      }
      patch(this.fragmentList)
      if (this.sidebarFilterQuery.trim()) {
        patch(this.sidebarFilterResults)
      }
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
      // 定点更新侧边栏列表中该片段的信息，避免全量刷新列表导致 tab 被切换到首页。
      this.updateFragmentListItem(fragment)
      // 更新首页文件夹统计信息。
      this.loadFolderList()
    },
    // handleFragmentDeleted 删除片段后清理 tab 和列表。
    handleFragmentDeleted(fragmentId) {
      this.fragmentTabs = this.fragmentTabs.filter(item => item.fragment.id !== fragmentId)
      this.fragmentList = this.fragmentList.filter(item => this.normalizeFragmentId(item.id || item.file_id) !== fragmentId)
      this.loadFragmentList()
      this.loadTrashList()
      this.rerunSubmittedSearch()
      this.rerunSidebarFilter()
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
        this.rerunSidebarFilter()
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
        this.rerunSidebarFilter()
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
      if (tabName === AI_SEARCH_TAB_NAME) {
        this.aiSearchTabVisible = false
        this.stopAiSearchSse()
        this.aiSearchEvents = []
        this.aiSearchAnswer = ''
        this.aiSearchLoading = false
        this.aiSearchReferencedFragments = []
        if (this.activeTab === AI_SEARCH_TAB_NAME) {
          this.activeTab = ''
          this.ensureDefaultFragmentTab()
        }
        return
      }
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
    goHome() {
      this.activeTab = HOME_TAB_NAME
    },
    enterFolder(folderName) {
      const nextFolderName = folderName || 'fragments'
      this.selectedFolderName = nextFolderName
      this.activeTab = HOME_TAB_NAME
      this.$router.replace({
        path: '/MemoryFragment',
        query: {
          ...this.$route.query,
          folder_name: nextFolderName,
        },
      })
    },
    // openAiSearchTab 打开 AI 智能搜索 tab 并启动 SSE 连接。
    openAiSearchTab(query) {
      this.stopAiSearchSse()
      this.clearAllStepTimers()
      this.aiSearchQuery = query
      this.aiSearchTabVisible = true
      this.aiSearchEvents = []
      this.aiSearchAnswer = ''
      this.aiSearchLoading = true
      this.aiSearchReferencedFragments = []
      this.aiSearchStepStartTimes = {}
      this.aiSearchStepElapsed = {}
      this.aiSearchExpandedSteps = {}
      this.activeTab = AI_SEARCH_TAB_NAME
      this.$nextTick(() => {
        this.startAiSearchSse(query)
      })
    },
    // startAiSearchSse 创建 EventSource 连接到 AI 搜索 SSE 端点。
    startAiSearchSse(query) {
      const sseHost = base.GetSseApiHost()
      if (!sseHost) {
        this.aiSearchEvents.push({ step: 'error', status: 'error', message: 'SSE 连接不可用', data: null })
        this.aiSearchLoading = false
        return
      }
      const params = 'query=' + encodeURIComponent(query) + '&token=' + encodeURIComponent(base.GetSafeToken()) + '&t=' + Date.now()
      const url = sseHost + '/api/MemoryFragmentAiSearch?' + params
      const eventSource = new EventSource(url)
      this.aiSearchSseClient = eventSource

      eventSource.onmessage = (event) => {
        const rawData = event.data
        if (rawData === '[DONE]' || rawData === '[CONNECT]') {
          return
        }
        try {
          const parsed = JSON.parse(rawData)
          if (parsed.step) {
            this.handleAiSearchEvent(parsed)
            return
          }
        } catch (e) {
          // 非 JSON，当作 answer 流式文本
        }
        this.aiSearchAnswer += rawData
      }

      eventSource.onerror = () => {
        this.aiSearchLoading = false
        this.stopAiSearchSse()
        this.clearAllStepTimers()
      }
    },
    // stopAiSearchSse 关闭 AI 搜索 SSE 连接。
    stopAiSearchSse() {
      if (this.aiSearchSseClient) {
        this.aiSearchSseClient.close()
        this.aiSearchSseClient = null
      }
    },
    // handleAiSearchEvent 处理 AI 搜索 SSE 事件。
    handleAiSearchEvent(event) {
      if (event.step === 'answer' && event.status === 'streaming') {
        this.aiSearchAnswer += event.data || ''
        return
      }
      if (event.step === 'answer' && event.status === 'running') {
        if (event.message) {
          this.aiSearchEvents.push(event)
        }
        this.startStepTimer(event.step)
        return
      }
      // 同一步骤可能有多条 running（如 read 进度更新），只保留最后一条
      if (event.status === 'running' && event.step !== 'done') {
        // 去掉同 step 之前的 running 事件，只保留最新进度
        const prevRunningIdx = this.aiSearchEvents.findLastIndex(e => e.step === event.step && e.status === 'running')
        if (prevRunningIdx >= 0) {
          this.aiSearchEvents.splice(prevRunningIdx, 1, event)
        } else {
          this.aiSearchEvents.push(event)
        }
        this.startStepTimer(event.step)
      } else if (event.status === 'done' && event.step !== 'done') {
        const runningIdx = this.aiSearchEvents.findLastIndex(e => e.step === event.step && e.status === 'running')
        if (runningIdx >= 0) {
          this.aiSearchEvents.splice(runningIdx, 1, event)
        } else {
          this.aiSearchEvents.push(event)
        }
        this.stopStepTimer(event.step)
      } else {
        this.aiSearchEvents.push(event)
      }
      if (event.step === 'done') {
        this.aiSearchLoading = false
        this.stopAiSearchSse()
        this.clearAllStepTimers()
        if (event.data && event.data.referenced_fragments) {
          this.aiSearchReferencedFragments = event.data.referenced_fragments
        }
      }
      if (event.step === 'error') {
        this.aiSearchLoading = false
        this.stopAiSearchSse()
        this.clearAllStepTimers()
      }
    },
    // startStepTimer 为指定步骤启动已用时间计时器。
    startStepTimer(step) {
      if (this.aiSearchStepStartTimes[step]) {
        return
      }
      this.aiSearchStepStartTimes[step] = Date.now()
      this.ensureStepTimer()
    },
    // ensureStepTimer 确保全局步骤计时器在运行。
    ensureStepTimer() {
      if (this.aiSearchStepTimerId) {
        return
      }
      this.aiSearchStepTimerId = setInterval(() => {
        const startTimes = this.aiSearchStepStartTimes
        const keys = Object.keys(startTimes)
        if (keys.length === 0) {
          return
        }
        const now = Date.now()
        const updated = {}
        keys.forEach(s => {
          updated[s] = Math.floor((now - startTimes[s]) / 1000)
        })
        this.aiSearchStepElapsed = updated
      }, 1000)
    },
    // stopStepTimer 停止指定步骤的计时器。
    stopStepTimer(step) {
      if (!this.aiSearchStepStartTimes[step]) {
        return
      }
      const elapsed = Math.floor((Date.now() - this.aiSearchStepStartTimes[step]) / 1000)
      const newElapsed = Object.assign({}, this.aiSearchStepElapsed)
      newElapsed[step] = elapsed
      this.aiSearchStepElapsed = newElapsed
      delete this.aiSearchStepStartTimes[step]
    },
    // clearAllStepTimers 清除所有步骤计时器。
    clearAllStepTimers() {
      if (this.aiSearchStepTimerId) {
        clearInterval(this.aiSearchStepTimerId)
        this.aiSearchStepTimerId = null
      }
      this.aiSearchStepStartTimes = {}
    },
    // toggleStepExpand 切换步骤详情的展开/收起状态。
    toggleStepExpand(step) {
      this.aiSearchExpandedSteps[step] = !this.aiSearchExpandedSteps[step]
    },
    // getStepLabel 返回步骤的中文名称。
    getStepLabel(step) {
      const labels = {
        keywords: '扩展搜索关键词',
        search: '搜索知识片段',
        judge: '评估片段相关性',
        read: '读取片段内容',
        answer: '生成综合回答',
      }
      return labels[step] || step
    },
    // getStepDoneText 返回步骤完成时的摘要文字。
    getStepDoneText(event) {
      const label = this.getStepLabel(event.step)
      const parts = [label]
      if (event.duration_ms) {
        parts.push(`用时 ${this.formatDuration(event.duration_ms)}`)
      }
      if (event.input_tokens || event.output_tokens) {
        parts.push(`输入 ${event.input_tokens || 0} token，输出 ${event.output_tokens || 0} token`)
      }
      if (event.step === 'search' && event.data && event.data.total !== undefined) {
        parts.splice(1, 0, `找到 ${event.data.total} 个片段`)
      }
      if (event.step === 'judge' && event.data && event.data.selected_count !== undefined) {
        parts.splice(1, 0, `选中 ${event.data.selected_count} 个片段`)
      }
      if (event.step === 'read' && event.data && event.data.read_fragments) {
        parts.splice(1, 0, `读取 ${event.data.read_fragments.length} 个片段`)
      }
      if (event.step === 'keywords' && event.data && event.data.keywords) {
        parts.splice(1, 0, `生成 ${event.data.keywords.length} 个关键词`)
      }
      return parts.join('，')
    },
    // canExpandStep 判断步骤是否可以展开查看详情。
    canExpandStep(event) {
      if (event.prompt) return true
      if (event.response) return true
      if (event.data && event.data.fragments && event.data.fragments.length > 0) return true
      if (event.data && event.data.selected_fragments && event.data.selected_fragments.length > 0) return true
      if (event.data && event.data.read_fragments && event.data.read_fragments.length > 0) return true
      if (event.data && event.data.keywords && event.data.keywords.length > 0) return true
      return false
    },
    // renderMarkdown 将 Markdown 文本渲染为 HTML。
    renderMarkdown(content) {
      if (!content) return ''
      try {
        return marked(content || '')
      } catch (e) {
        return this.escapeHtml(content)
      }
    },
    // formatDuration 将毫秒格式化为可读的时间文本。
    formatDuration(ms) {
      if (!ms || ms <= 0) return '-'
      if (ms < 1000) return ms + 'ms'
      const seconds = (ms / 1000).toFixed(1)
      return seconds + 's'
    },
    applySidebarHiddenQuery() {
      if (this.isEmbedded || String(this.$route.query.hide_sidebar || '') === '1') {
        this.sidebarCollapsed = true
      }
    },
    // fragmentRefCount 返回指定片段被引用的次数。
    fragmentRefCount(fragmentId) {
      const refs = this.getFragmentRefs(fragmentId)
      return refs.length
    },
    // getFragmentRefs 返回指定片段的引用列表。
    getFragmentRefs(fragmentId) {
      const normalized = this.normalizeFragmentId(fragmentId)
      if (!normalized) return []
      return this.fragmentReferences[normalized] || []
    },
    // toggleFragmentRefs 展开/收起片段的引用列表。
    toggleFragmentRefs(fragmentId) {
      const normalized = this.normalizeFragmentId(fragmentId)
      if (!normalized) return
      this.expandedFragmentRefs[normalized] = !this.expandedFragmentRefs[normalized]
    },
    // openFragmentRef 根据引用类型跳转到工作流程或打开知识片段。
    openFragmentRef(ref) {
      if (!ref) return
      if (ref.type === 'workflow') {
        const routeData = this.$router.resolve({ path: '/TaskWorkflow/' + ref.id })
        window.open(routeData.href, '_blank')
        return
      }
      if (ref.type === 'fragment') {
        this.openFragment(ref.id)
      }
    },
    // loadFragmentReferences 批量查询片段引用信息。
    loadFragmentReferences(fragmentIds) {
      if (!fragmentIds || fragmentIds.length === 0) return
      MemoryFragmentApi.MemoryFragmentReferences(fragmentIds, (response) => {
        if (response.ErrCode !== 0) return
        const data = response.Data || {}
        Object.keys(data).forEach((fid) => {
          this.fragmentReferences[fid] = data[fid] || []
        })
      })
    },
  }
}
</script>

<style scoped src="@/css/components/MemoryFragment.css"></style>

<style>
.memory-tabs--embedded .el-tabs__header {
  display: none !important;
}
</style>

