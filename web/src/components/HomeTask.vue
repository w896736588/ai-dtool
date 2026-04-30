<template>
  <div class="home-task-page">
    <div class="home-task-page__body">
      <div class="home-task-panel__header">
        <div class="home-task-panel__heading">
          <div class="home-task-panel__title">任务清单</div>
        </div>
        <div class="home-task-toolbar__actions">
          <GitActionButton compact variant="warning" @click="openHomeTaskReportSettingsDialog">
            设置
          </GitActionButton>
          <GitActionButton compact variant="info" :loading="homeTaskGeneratingDailyReport" @click="generateHomeTaskDailyReport">
            {{ HOME_TASK_DAILY_REPORT_BUTTON_TEXT }}
          </GitActionButton>
          <GitActionButton compact @click="openCreateHomeTaskDialog">
            新增任务
          </GitActionButton>
        </div>
      </div>

      <el-tabs v-model="homeTaskActiveTab" class="home-task-tabs" @tab-change="handleHomeTaskTabChange">
        <el-tab-pane label="活跃中" :name="HOME_TASK_TAB_ACTIVE">
          <div v-loading="homeTaskLoadingActive" class="home-task-list">
            <div v-if="homeTaskActiveList.length === 0" class="home-task-empty">
              当前没有未归档任务
            </div>
            <div
              v-for="task in homeTaskActiveList"
              :key="task.id"
              class="home-task-card"
              :class="{ 'edit-success': !!homeTaskEditFeedbackMap[task.id] }"
            >
              <div class="home-task-card__header">
                <div>
                  <div class="home-task-card__title">{{ task.name }}</div>
                  <div class="home-task-card__meta">
                    <span>开始时间：{{ task.start_time_desc || '-' }}</span>
                    <span>最后操作：{{ task.last_operated_at_desc || '-' }}</span>
                    <a v-if="task.tapd_url" :href="task.tapd_url" target="_blank" class="home-task-card__tapd-link">TAPD需求</a>
                    <span v-if="getHomeTaskGitRepoName(task)" class="home-task-card__git-repo">{{ getHomeTaskGitRepoName(task) }}</span>
                    <span v-if="task.api_dev_enabled === 1" class="home-task-card__api-dev">接口: {{ getHomeTaskApiDevLabel(task) }}</span>
                    <span class="home-task-card__status-group">
                      <el-tag size="small" effect="light" :type="getHomeTaskStatusTagType(task.task_status)">
                        {{ task.task_status }}
                      </el-tag>
                      <el-tag
                        v-if="hasHomeTaskMemoryFragment(task)"
                        size="small"
                        effect="plain"
                        class="home-task-memory-link-tag"
                        @click.stop="openHomeTaskMemoryFragment(task)"
                      >
                        {{ getHomeTaskMemoryTagText(task) }}
                      </el-tag>
                    </span>
                  </div>
                </div>
                <div class="home-task-card__actions">
                  <GitActionButton
                    compact
                    variant="primary"
                    :disabled="isHomeTaskBusy(task.id)"
                    @click="openTaskWorkflow(task)"
                  >
                    工作流程
                  </GitActionButton>
                  <el-dropdown
                    trigger="click"
                    :disabled="isHomeTaskBusy(task.id)"
                    @command="handleHomeTaskActionCommand(task, $event)"
                  >
                    <GitActionButton
                      compact
                      :loading="isHomeTaskBusy(task.id, HOME_TASK_OPERATE_STATUS) || isHomeTaskBusy(task.id, HOME_TASK_OPERATE_ARCHIVE)"
                      :variant="getHomeTaskActionButtonVariant(task.task_status)"
                    >
                      状态变更
                    </GitActionButton>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item
                          v-for="status in homeTaskStatusOptions"
                          :key="status"
                          :command="buildHomeTaskStatusCommand(status)"
                          :disabled="task.task_status === status"
                        >
                          {{ status }}
                        </el-dropdown-item>
                        <el-dropdown-item :command="HOME_TASK_ACTION_COMMAND_ARCHIVE">
                          归档任务
                        </el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                  <GitActionButton
                    compact
                    variant="info"
                    :disabled="isHomeTaskBusy(task.id)"
                    @click="editHomeTask(task)"
                  >
                    {{ HOME_TASK_EDIT_BUTTON_TEXT }}
                  </GitActionButton>
                  <GitActionButton
                    compact
                    variant="danger"
                    :loading="isHomeTaskBusy(task.id, HOME_TASK_OPERATE_DELETE)"
                    :disabled="isHomeTaskBusy(task.id) && !isHomeTaskBusy(task.id, HOME_TASK_OPERATE_DELETE)"
                    @click="deleteHomeTask(task)"
                  >
                    删除任务
                  </GitActionButton>
                </div>
              </div>
            </div>
          </div>
        </el-tab-pane>
        <el-tab-pane label="归档" :name="HOME_TASK_TAB_ARCHIVED">
          <div v-loading="homeTaskLoadingArchived" class="home-task-list">
            <div v-if="homeTaskArchivedList.length === 0" class="home-task-empty">
              当前没有归档任务
            </div>
            <div
              v-for="task in homeTaskArchivedList"
              :key="task.id"
              class="home-task-card home-task-card--archived"
              :class="{ 'edit-success': !!homeTaskEditFeedbackMap[task.id] }"
            >
              <div class="home-task-card__header">
                <div>
                  <div class="home-task-card__title">{{ task.name }}</div>
                  <div class="home-task-card__meta">
                    <span>开始时间：{{ task.start_time_desc || '-' }}</span>
                    <span>最后操作：{{ task.last_operated_at_desc || '-' }}</span>
                    <a v-if="task.tapd_url" :href="task.tapd_url" target="_blank" class="home-task-card__tapd-link">TAPD需求</a>
                    <span v-if="getHomeTaskGitRepoName(task)" class="home-task-card__git-repo">{{ getHomeTaskGitRepoName(task) }}</span>
                    <span v-if="task.api_dev_enabled === 1" class="home-task-card__api-dev">接口: {{ getHomeTaskApiDevLabel(task) }}</span>
                  </div>
                </div>
                <div class="home-task-card__status-group">
                  <el-tag size="small" effect="light" :type="getHomeTaskStatusTagType(task.task_status)">
                    {{ task.task_status }}
                  </el-tag>
                  <el-tag
                    v-if="hasHomeTaskMemoryFragment(task)"
                    size="small"
                    effect="plain"
                    class="home-task-memory-link-tag"
                    @click.stop="openHomeTaskMemoryFragment(task)"
                  >
                    {{ getHomeTaskMemoryTagText(task) }}
                  </el-tag>
                </div>
              </div>
              <div v-if="hasHomeTaskMemoryFragment(task)" class="home-task-card__memory">
                <div class="home-task-card__memory-label">关联知识片段</div>
                <div class="home-task-card__memory-title">
                  {{ task.memory_fragment?.title || `#${task.memory_fragment_id}` }}
                </div>
                <div v-if="task.memory_fragment?.content" class="home-task-card__memory-content">
                  <pre class="memory-content-text">{{ getFragmentPreview(task.memory_fragment.content, task.id) }}</pre>
                  <button
                    v-if="isFragmentExpandable(task.memory_fragment.content)"
                    type="button"
                    class="memory-content-toggle"
                    @click="toggleFragmentExpand(task.id)"
                  >
                    {{ homeTaskExpandedFragments[task.id] ? '收起' : '展开' }}
                  </button>
                </div>
                <div v-if="Array.isArray(task.memory_fragment?.tags) && task.memory_fragment.tags.length > 0" class="home-task-card__memory-tags">
                  <el-tag
                    v-for="tag in task.memory_fragment.tags"
                    :key="`${task.id}-${tag}`"
                    size="small"
                    effect="plain"
                  >
                    {{ tag }}
                  </el-tag>
                </div>
              </div>
              <div class="home-task-card__actions">
                <GitActionButton
                  compact
                  variant="primary"
                  :disabled="isHomeTaskBusy(task.id)"
                  @click="openTaskWorkflow(task)"
                >
                  工作流程
                </GitActionButton>
                <el-dropdown
                  trigger="click"
                  :disabled="isHomeTaskBusy(task.id)"
                  @command="handleHomeTaskActionCommand(task, $event)"
                >
                  <GitActionButton
                    compact
                    :loading="isHomeTaskBusy(task.id, HOME_TASK_OPERATE_STATUS) || isHomeTaskBusy(task.id, HOME_TASK_OPERATE_ARCHIVE)"
                    :variant="getHomeTaskActionButtonVariant(task.task_status)"
                  >
                    状态变更
                  </GitActionButton>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item
                        v-for="status in homeTaskStatusOptions"
                        :key="status"
                        :command="buildHomeTaskStatusCommand(status)"
                        :disabled="task.task_status === status"
                      >
                        {{ status }}
                      </el-dropdown-item>
                      <el-dropdown-item :command="HOME_TASK_ACTION_COMMAND_UNARCHIVE">
                        取消归档
                      </el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
                <GitActionButton
                  compact
                  variant="info"
                  :disabled="isHomeTaskBusy(task.id)"
                  @click="editHomeTask(task)"
                >
                  {{ HOME_TASK_EDIT_BUTTON_TEXT }}
                </GitActionButton>
                <GitActionButton
                  compact
                  variant="danger"
                  :loading="isHomeTaskBusy(task.id, HOME_TASK_OPERATE_DELETE)"
                  :disabled="isHomeTaskBusy(task.id) && !isHomeTaskBusy(task.id, HOME_TASK_OPERATE_DELETE)"
                  @click="deleteHomeTask(task)"
                >
                  删除任务
                </GitActionButton>
              </div>
            </div>
          </div>
        </el-tab-pane>
      </el-tabs>
    </div>

    <el-dialog
      v-model="homeTaskDialogVisible"
      :title="homeTaskDialogTitle"
      width="920px"
      top="5vh"
      class="home-task-dialog"
      destroy-on-close
    >
      <el-form label-width="88px" class="home-task-form" @submit.prevent>
        <el-row :gutter="12">
          <el-col :span="24">
            <el-form-item label="任务名称">
              <el-input
                v-model="homeTaskForm.name"
                maxlength="80"
                show-word-limit
                placeholder="例如：整理缓存淘汰策略"
                @keyup.enter="saveHomeTask"
              />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="tapd需求地址">
              <el-input
                v-model="homeTaskForm.tapd_url"
                placeholder="例如：https://www.tapd.cn/123456"
              />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="Git仓库">
              <el-select
                v-model="homeTaskForm.git_id"
                clearable
                filterable
                style="width: 100%"
                placeholder="选择关联的Git仓库（可选）"
                :loading="homeTaskGitRepoLoading"
              >
                <el-option-group
                  v-for="group in homeTaskGitRepoGroupedOptions"
                  :key="group.label"
                  :label="group.label"
                >
                  <el-option
                    v-for="repo in group.options"
                    :key="repo.value"
                    :label="repo.label"
                    :value="repo.value"
                  />
                </el-option-group>
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="MySQL配置">
              <el-select
                v-model="homeTaskForm.mysql_id"
                clearable
                filterable
                style="width: 100%"
                placeholder="选择关联的MySQL配置（可选）"
                :loading="homeTaskMysqlLoading"
              >
                <el-option
                  v-for="item in homeTaskMysqlList"
                  :key="item.id"
                  :label="item.name"
                  :value="item.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="12">
            <el-form-item label="任务状态">
              <el-select v-model="homeTaskForm.task_status" style="width: 100%">
                <el-option
                  v-for="status in homeTaskStatusOptions"
                  :key="status"
                  :label="status"
                  :value="status"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :xs="24" :sm="12" :md="12">
            <el-form-item label="开始日期">
              <el-date-picker
                v-model="homeTaskForm.start_date"
                type="date"
                value-format="YYYY-MM-DD"
                placeholder="选择开始日期"
                style="width: 100%"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="24">
            <el-form-item label="接口开发">
              <el-switch
                v-model="homeTaskForm.api_dev_enabled"
                :active-value="1"
                :inactive-value="0"
                active-text="启用"
                inactive-text="关闭"
              />
            </el-form-item>
          </el-col>
          <template v-if="homeTaskForm.api_dev_enabled === 1">
            <el-col :xs="24" :sm="12" :md="12">
              <el-form-item label="选择集合">
                <el-select
                  v-model="homeTaskForm.api_collection_id"
                  filterable
                  style="width: 100%"
                  placeholder="请选择接口集合（必选）"
                  :loading="homeTaskApiCollectionLoading"
                  @change="handleHomeTaskApiCollectionChange"
                >
                  <el-option
                    v-for="col in homeTaskApiCollectionList"
                    :key="col.id"
                    :label="col.name"
                    :value="col.id"
                  />
                </el-select>
              </el-form-item>
            </el-col>
            <el-col :xs="24" :sm="12" :md="12">
              <el-form-item label="选择文件夹">
                <el-select
                  v-model="homeTaskForm.api_dir_id"
                  filterable
                  clearable
                  style="width: 100%"
                  placeholder="留空则自动创建（可选）"
                  :loading="homeTaskApiFolderLoading"
                  :disabled="!homeTaskForm.api_collection_id"
                >
                  <el-option
                    v-for="dir in homeTaskApiFolderList"
                    :key="dir.id"
                    :label="dir.name"
                    :value="dir.id"
                  />
                </el-select>
              </el-form-item>
            </el-col>
          </template>
        </el-row>
        <el-row :gutter="12">
          <el-col :span="24">
            <el-form-item label="知识片段">
              <el-select
                v-model="homeTaskForm.memory_fragment_id"
                filterable
                clearable
                style="width: 100%"
                placeholder="可选择已有知识片段；留空则保存时自动新建"
                :loading="homeTaskFragmentLoading"
              >
                <el-option
                  v-for="fragment in homeTaskFragmentOptions"
                  :key="fragment.id"
                  :label="buildHomeTaskFragmentOptionLabel(fragment)"
                  :value="fragment.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
      </el-form>
      <template #footer>
        <div class="home-task-dialog__footer">
          <GitActionButton compact variant="info" @click="closeHomeTaskDialog">
            取消
          </GitActionButton>
          <GitActionButton compact :loading="homeTaskSaving" @click="saveHomeTask">
            {{ homeTaskForm.id > 0 ? '保存修改' : '添加任务' }}
          </GitActionButton>
        </div>
      </template>
    </el-dialog>

    <SettingsDialog
      v-model="homeTaskReportSettingsDialogVisible"
      title="任务清单设置"
      width="80%"
      @closed="refreshHomeTaskReportSettings"
    >
      <HomeTaskReportSetting ref="homeTaskReportSetting" />
    </SettingsDialog>
  </div>
</template>

<script>
import base from '../utils/base'
import homeTaskApi from '@/utils/base/home_task'
import memoryFragmentApi from '@/utils/base/memory_fragment'
import gitApi from '@/utils/base/git'
import mysqlSetApi from '@/utils/base/mysql_set'
import apiManagement from '@/utils/base/api'
const { mergeHomeTaskFragmentOptions } = require('@/utils/home_task_fragment_options.cjs')
import GitActionButton from "@/components/base/GitActionButton.vue";
import SettingsDialog from '@/components/base/SettingsDialog.vue'
import HomeTaskReportSetting from '@/components/set/home_task_report.vue'

const HOME_TASK_TAB_ACTIVE = 'active'
const HOME_TASK_TAB_ARCHIVED = 'archived'
const HOME_TASK_ARCHIVED_NO = 0
const HOME_TASK_ARCHIVED_YES = 1
const HOME_TASK_STATUS_TODO = '待开始'
const HOME_TASK_STATUS_DEVELOPING = '开发中'
const HOME_TASK_STATUS_SELF_TESTING = '自测中'
const HOME_TASK_STATUS_SELF_TESTED = '自测完'
const HOME_TASK_STATUS_PENDING_INTEGRATION = '待对接'
const HOME_TASK_STATUS_INTEGRATING = '对接中'
const HOME_TASK_STATUS_TESTING = '测试中'
const HOME_TASK_STATUS_RELEASING = '上线中'
const HOME_TASK_STATUS_ONLINE = '已上线'
const HOME_TASK_OPERATE_SAVE = 'save'
const HOME_TASK_OPERATE_STATUS = 'status'
const HOME_TASK_OPERATE_ARCHIVE = 'archive'
const HOME_TASK_OPERATE_DELETE = 'delete'
const HOME_TASK_ACTION_COMMAND_EDIT = 'edit'
const HOME_TASK_ACTION_COMMAND_ARCHIVE = 'archive'
const HOME_TASK_ACTION_COMMAND_UNARCHIVE = 'unarchive'
const HOME_TASK_DELETE_CONFIRM_TITLE = '确认删除'
const HOME_TASK_DELETE_CONFIRM_MESSAGE_PREFIX = '确定要删除任务"'
const HOME_TASK_DELETE_CONFIRM_MESSAGE_SUFFIX = '"吗？该操作不可恢复。'
const HOME_TASK_DELETE_SUCCESS_MESSAGE = '任务已删除'
const HOME_TASK_EDIT_BUTTON_TEXT = '编辑任务'
const HOME_TASK_DAILY_REPORT_BUTTON_TEXT = 'AI 生成工作日报'
const HOME_TASK_DAILY_REPORT_SUCCESS_MESSAGE = '工作日报任务已加入异步任务列表'
const HOME_TASK_DAILY_REPORT_FAILED_MESSAGE = '工作日报生成失败'
const HOME_TASK_ACTION_COMMAND_STATUS_PREFIX = 'status:'
const HOME_TASK_EMPTY_START_DATE = ''
const HOME_TASK_STATUS_OPTIONS = [
  HOME_TASK_STATUS_TODO,
  HOME_TASK_STATUS_DEVELOPING,
  HOME_TASK_STATUS_SELF_TESTING,
  HOME_TASK_STATUS_SELF_TESTED,
  HOME_TASK_STATUS_PENDING_INTEGRATION,
  HOME_TASK_STATUS_INTEGRATING,
  HOME_TASK_STATUS_TESTING,
  HOME_TASK_STATUS_RELEASING,
  HOME_TASK_STATUS_ONLINE,
]

function getTodayDateText() {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function createHomeTaskDefaultForm() {
  return {
    id: 0,
    name: '',
    task_status: HOME_TASK_STATUS_TODO,
    start_date: getTodayDateText(),
    memory_fragment_id: '',
    tapd_url: '',
    git_id: 0,
    api_dev_enabled: 0,
    api_collection_id: 0,
    api_dir_id: 0,
    mysql_id: 0,
  }
}

export default {
  data() {
    return {
      HOME_TASK_TAB_ACTIVE,
      HOME_TASK_TAB_ARCHIVED,
      HOME_TASK_ARCHIVED_NO,
      HOME_TASK_ARCHIVED_YES,
      HOME_TASK_OPERATE_STATUS,
      HOME_TASK_OPERATE_ARCHIVE,
      HOME_TASK_OPERATE_DELETE,
      HOME_TASK_ACTION_COMMAND_EDIT,
      HOME_TASK_ACTION_COMMAND_ARCHIVE,
      HOME_TASK_ACTION_COMMAND_UNARCHIVE,
      HOME_TASK_EDIT_BUTTON_TEXT,
      HOME_TASK_DAILY_REPORT_BUTTON_TEXT,
      homeTaskActiveTab: HOME_TASK_TAB_ACTIVE,
      homeTaskDialogVisible: false,
      homeTaskReportSettingsDialogVisible: false,
      homeTaskLoadingActive: false,
      homeTaskLoadingArchived: false,
      homeTaskGeneratingDailyReport: false,
      homeTaskSaving: false,
      homeTaskFragmentLoading: false,
      homeTaskOperatingId: 0,
      homeTaskOperatingType: '',
      homeTaskActiveList: [],
      homeTaskArchivedList: [],
      homeTaskFragmentOptions: [],
      homeTaskStatusOptions: HOME_TASK_STATUS_OPTIONS,
      homeTaskForm: createHomeTaskDefaultForm(),
      homeTaskExpandedFragments: {},
      homeTaskEditFeedbackMap: {},
      homeTaskEditFeedbackTimers: {},
      homeTaskEditFeedbackDurationMs: 1000,
      homeTaskGitRepoList: [],
      homeTaskGitRepoLoading: false,
      homeTaskApiCollectionList: [],
      homeTaskApiFolderList: [],
      homeTaskApiCollectionLoading: false,
      homeTaskApiFolderLoading: false,
      homeTaskMysqlList: [],
      homeTaskMysqlLoading: false,
    }
  },
  computed: {
    homeTaskDialogTitle() {
      return this.homeTaskForm.id > 0 ? '编辑任务' : '新增任务'
    },
    homeTaskGitRepoGroupedOptions() {
      const groupMap = {}
      const groupOrder = []
      for (const repo of this.homeTaskGitRepoList) {
        const groupName = repo.git_group_name || '未分组'
        if (!groupMap[groupName]) {
          groupMap[groupName] = []
          groupOrder.push(groupName)
        }
        groupMap[groupName].push({ label: repo.name, value: repo.id })
      }
      return groupOrder.map(name => ({ label: name, options: groupMap[name] }))
    },
  },
  mounted() {
    this.loadHomeTaskFragmentOptions()
    this.loadHomeTaskGitRepoList()
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
  },
  activated() {
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
    this.loadHomeTaskFragmentOptions()
    this.loadHomeTaskGitRepoList()
  },
  methods: {
    handleHomeTaskTabChange(tabName) {
      if (tabName === HOME_TASK_TAB_ACTIVE) {
        this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
        return
      }
      this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
    },
    loadHomeTaskList(isArchived) {
      if (isArchived === HOME_TASK_ARCHIVED_YES) {
        this.homeTaskLoadingArchived = true
      } else {
        this.homeTaskLoadingActive = true
      }
      homeTaskApi.HomeTaskList(isArchived, (response) => {
        if (isArchived === HOME_TASK_ARCHIVED_YES) {
          this.homeTaskLoadingArchived = false
        } else {
          this.homeTaskLoadingActive = false
        }
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '任务列表加载失败')
          return
        }
        const taskList = Array.isArray(response.Data?.task_list) ? response.Data.task_list : []
        if (isArchived === HOME_TASK_ARCHIVED_YES) {
          this.homeTaskArchivedList = taskList
        } else {
          this.homeTaskActiveList = taskList
        }
      })
    },
    refreshAllHomeTaskList() {
      this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
      this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
      this.loadHomeTaskFragmentOptions()
    },
    resetHomeTaskForm() {
      this.homeTaskForm = createHomeTaskDefaultForm()
    },
    loadHomeTaskFragmentOptions(selectedFragment = null) {
      this.homeTaskFragmentLoading = true
      memoryFragmentApi.MemoryFragmentList(200, (response) => {
        this.homeTaskFragmentLoading = false
        if (!(response && response.ErrCode === 0 && Array.isArray(response.Data))) {
          return
        }
        this.homeTaskFragmentOptions = mergeHomeTaskFragmentOptions(response.Data, selectedFragment)
      })
    },
    ensureHomeTaskFragmentOption(fragment) {
      this.homeTaskFragmentOptions = mergeHomeTaskFragmentOptions(this.homeTaskFragmentOptions, fragment)
    },
    loadHomeTaskGitRepoList() {
      this.homeTaskGitRepoLoading = true
      gitApi.GitConfigList({}, (response) => {
        this.homeTaskGitRepoLoading = false
        if (!(response && response.ErrCode === 0)) {
          return
        }
        const gitList = Array.isArray(response.Data?.git_list) ? response.Data.git_list : []
        const groupList = Array.isArray(response.Data?.git_group_list) ? response.Data.git_group_list : []
        const groupMap = {}
        for (const g of groupList) {
          groupMap[g.id] = g.name
        }
        this.homeTaskGitRepoList = gitList.map(repo => ({
          ...repo,
          git_group_name: groupMap[repo.git_group_id] || '未分组',
        }))
      })
    },
    loadHomeTaskMysqlList() {
      this.homeTaskMysqlLoading = true
      mysqlSetApi.MysqlList((response) => {
        this.homeTaskMysqlLoading = false
        if (!(response && response.ErrCode === 0)) {
          return
        }
        this.homeTaskMysqlList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    loadHomeTaskApiCollections() {
      this.homeTaskApiCollectionLoading = true
      apiManagement.CollectionListBasic({}, (response) => {
        this.homeTaskApiCollectionLoading = false
        if (!(response && response.ErrCode === 0)) {
          return
        }
        this.homeTaskApiCollectionList = Array.isArray(response.Data?.list) ? response.Data.list : []
      })
    },
    loadHomeTaskApiFolders(collectionId) {
      if (!collectionId) {
        this.homeTaskApiFolderList = []
        return
      }
      this.homeTaskApiFolderLoading = true
      apiManagement.CollectionFoldersBasic({ collection_id: collectionId }, (response) => {
        this.homeTaskApiFolderLoading = false
        if (!(response && response.ErrCode === 0)) {
          return
        }
        this.homeTaskApiFolderList = Array.isArray(response.Data?.list) ? response.Data.list : []
      })
    },
    handleHomeTaskApiCollectionChange(collectionId) {
      this.homeTaskForm.api_dir_id = 0
      this.loadHomeTaskApiFolders(collectionId)
    },
    buildHomeTaskFragmentOptionLabel(fragment) {
      const tagText = Array.isArray(fragment?.tags) && fragment.tags.length > 0 ? ` [${fragment.tags.join('、')}]` : ''
      return `${fragment.title || `#${fragment.id}`}${tagText}`
    },
    openCreateHomeTaskDialog() {
      this.resetHomeTaskForm()
      this.loadHomeTaskFragmentOptions()
      this.loadHomeTaskGitRepoList()
      this.loadHomeTaskApiCollections()
      this.loadHomeTaskMysqlList()
      this.homeTaskDialogVisible = true
    },
    openHomeTaskReportSettingsDialog() {
      this.homeTaskReportSettingsDialogVisible = true
      this.$nextTick(() => {
        if (this.$refs.homeTaskReportSetting && this.$refs.homeTaskReportSetting.loadConfig) {
          this.$refs.homeTaskReportSetting.loadConfig()
        }
        if (this.$refs.homeTaskReportSetting && this.$refs.homeTaskReportSetting.loadAiModelList) {
          this.$refs.homeTaskReportSetting.loadAiModelList()
        }
      })
    },
    refreshHomeTaskReportSettings() {
      if (this.$refs.homeTaskReportSetting && this.$refs.homeTaskReportSetting.loadConfig) {
        this.$refs.homeTaskReportSetting.loadConfig()
      }
    },
    generateHomeTaskDailyReport() {
      if (this.homeTaskGeneratingDailyReport) {
        return
      }
      this.homeTaskGeneratingDailyReport = true
      homeTaskApi.HomeTaskDailyReportGenerate((response) => {
        this.homeTaskGeneratingDailyReport = false
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || HOME_TASK_DAILY_REPORT_FAILED_MESSAGE)
          return
        }
        this.$helperNotify.success(HOME_TASK_DAILY_REPORT_SUCCESS_MESSAGE)
      })
    },
    closeHomeTaskDialog() {
      this.homeTaskDialogVisible = false
      this.resetHomeTaskForm()
    },
    editHomeTask(task) {
      this.ensureHomeTaskFragmentOption(task.memory_fragment)
      const fragmentID = this.normalizeHomeTaskMemoryFragmentId(task?.memory_fragment?.file_id || task?.memory_fragment_id)
      this.homeTaskForm = {
        id: Number(task.id || 0),
        name: task.name || '',
        task_status: task.task_status || HOME_TASK_STATUS_TODO,
        start_date: task.start_time_desc || getTodayDateText(),
        memory_fragment_id: fragmentID,
        tapd_url: task.tapd_url || '',
        git_id: Number(task.git_id || 0),
        api_dev_enabled: Number(task.api_dev_enabled || 0),
        api_collection_id: Number(task.api_collection_id || 0),
        api_dir_id: Number(task.api_dir_id || 0),
        mysql_id: Number(task.mysql_id || 0),
      }
      this.loadHomeTaskFragmentOptions(task.memory_fragment)
      this.loadHomeTaskGitRepoList()
      this.loadHomeTaskApiCollections()
      this.loadHomeTaskMysqlList()
      if (Number(task.api_collection_id || 0) > 0) {
        this.loadHomeTaskApiFolders(Number(task.api_collection_id))
      }
      this.homeTaskDialogVisible = true
    },
    openHomeTaskMemoryFragment(task) {
      const fragmentId = this.normalizeHomeTaskMemoryFragmentId(task?.memory_fragment?.file_id || task?.memory_fragment_id)
      if (!fragmentId) {
        this.$helperNotify.error('当前任务还没有关联知识片段')
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
    openTaskWorkflow(task) {
      const taskId = Number(task?.id || 0)
      if (taskId <= 0) {
        this.$helperNotify.error('任务 id 不合法')
        return
      }
      const routeInfo = this.$router.resolve({
        path: `/TaskWorkflow/${taskId}`,
      })
      window.open(routeInfo.href, '_blank')
    },
    getHomeTaskMemoryTagText(task) {
      const fragmentId = this.normalizeHomeTaskMemoryFragmentId(task?.memory_fragment?.file_id || task?.memory_fragment_id)
      if (!fragmentId) {
        return ''
      }
      const fragmentTitle = String(task?.memory_fragment?.title || '').trim()
      const displayTitle = fragmentTitle || `#${fragmentId}`
      return `已关联知识片段 "${displayTitle}"`
    },
    normalizeHomeTaskMemoryFragmentId(rawId) {
      const fragmentId = String(rawId || '').trim()
      if (!fragmentId || fragmentId === '0') {
        return ''
      }
      return fragmentId
    },
    hasHomeTaskMemoryFragment(task) {
      return this.normalizeHomeTaskMemoryFragmentId(task?.memory_fragment?.file_id || task?.memory_fragment_id) !== ''
    },
    getHomeTaskGitRepoName(task) {
      const gitId = Number(task?.git_id || 0)
      if (gitId <= 0) return ''
      const repo = this.homeTaskGitRepoList.find(r => Number(r.id) === gitId)
      return repo ? repo.name : ''
    },
    getHomeTaskApiDevLabel(task) {
      const collectionId = Number(task?.api_collection_id || 0)
      if (collectionId <= 0) return ''
      const col = this.homeTaskApiCollectionList.find(c => Number(c.id) === collectionId)
      const colName = col ? col.name : `#${collectionId}`
      const dirId = Number(task?.api_dir_id || 0)
      if (dirId <= 0) return colName
      const dir = this.homeTaskApiFolderList.find(d => Number(d.id) === dirId)
      const dirName = dir ? dir.name : `#${dirId}`
      return `${colName} / ${dirName}`
    },
    saveHomeTask() {
      if (this.homeTaskSaving) {
        return
      }
      const taskName = String(this.homeTaskForm.name || '').trim()
      if (!taskName) {
        this.$helperNotify.error('任务名称不能为空')
        return
      }
      if (this.homeTaskForm.api_dev_enabled === 1 && !this.homeTaskForm.api_collection_id) {
        this.$helperNotify.error('启用接口开发时必须选择集合')
        return
      }
      this.homeTaskSaving = true
      this.homeTaskOperatingType = HOME_TASK_OPERATE_SAVE
      homeTaskApi.HomeTaskSave({
        id: Number(this.homeTaskForm.id || 0),
        name: taskName,
        task_status: this.homeTaskForm.task_status,
        start_time: this.convertHomeTaskDateToUnix(this.homeTaskForm.start_date),
        memory_fragment_id: String(this.homeTaskForm.memory_fragment_id || '').trim(),
        tapd_url: String(this.homeTaskForm.tapd_url || '').trim(),
        git_id: Number(this.homeTaskForm.git_id || 0),
        api_dev_enabled: Number(this.homeTaskForm.api_dev_enabled || 0),
        api_collection_id: Number(this.homeTaskForm.api_collection_id || 0),
        api_dir_id: Number(this.homeTaskForm.api_dir_id || 0),
        api_host: base.GetApiHost() || window.location.origin,
        api_token: base.GetSafeToken(),
        mysql_id: Number(this.homeTaskForm.mysql_id || 0),
        api_token: base.GetSafeToken(),
      }, (response) => {
        this.homeTaskSaving = false
        this.homeTaskOperatingType = ''
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '任务保存失败')
          return
        }
        const isEdit = this.homeTaskForm.id > 0
        this.$helperNotify.success(isEdit ? '任务已更新' : '任务已创建')
        this.closeHomeTaskDialog()
        if (isEdit) {
          const taskId = Number(this.homeTaskForm.id)
          this.triggerHomeTaskEditFeedback(taskId)
        }
        this.refreshAllHomeTaskList()
      })
    },
    isHomeTaskBusy(taskId, operateType = '') {
      const normalizedTaskId = Number(taskId || 0)
      if (normalizedTaskId <= 0 || this.homeTaskOperatingId !== normalizedTaskId) {
        return false
      }
      if (!operateType) {
        return true
      }
      return this.homeTaskOperatingType === operateType
    },
    quickUpdateHomeTaskStatus(task, taskStatus) {
      if (this.homeTaskOperatingId > 0) {
        return
      }
      this.homeTaskOperatingId = Number(task.id || 0)
      this.homeTaskOperatingType = HOME_TASK_OPERATE_STATUS
      homeTaskApi.HomeTaskStatusQuickUpdate(this.homeTaskOperatingId, taskStatus, (response) => {
        this.homeTaskOperatingId = 0
        this.homeTaskOperatingType = ''
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '状态切换失败')
          return
        }
        const updatedTask = response.Data
        if (updatedTask && updatedTask.id) {
          this.updateHomeTaskInList(updatedTask)
        }
      })
    },
    updateHomeTaskInList(updatedTask) {
      const taskId = Number(updatedTask.id || 0)
      const activeIndex = this.homeTaskActiveList.findIndex(t => Number(t.id) === taskId)
      if (activeIndex >= 0) {
        this.homeTaskActiveList[activeIndex] = { ...this.homeTaskActiveList[activeIndex], ...updatedTask }
        return
      }
      const archivedIndex = this.homeTaskArchivedList.findIndex(t => Number(t.id) === taskId)
      if (archivedIndex >= 0) {
        this.homeTaskArchivedList[archivedIndex] = { ...this.homeTaskArchivedList[archivedIndex], ...updatedTask }
      }
    },
    triggerHomeTaskEditFeedback(taskId) {
      const normalizedId = Number(taskId || 0)
      if (normalizedId <= 0) {
        return
      }
      if (this.homeTaskEditFeedbackTimers[normalizedId]) {
        window.clearTimeout(this.homeTaskEditFeedbackTimers[normalizedId])
      }
      this.homeTaskEditFeedbackMap = { ...this.homeTaskEditFeedbackMap, [normalizedId]: Date.now() }
      this.homeTaskEditFeedbackTimers[normalizedId] = window.setTimeout(() => {
        const { [normalizedId]: _, ...rest } = this.homeTaskEditFeedbackMap
        this.homeTaskEditFeedbackMap = rest
        delete this.homeTaskEditFeedbackTimers[normalizedId]
      }, this.homeTaskEditFeedbackDurationMs)
    },
    buildHomeTaskStatusCommand(taskStatus) {
      return `${HOME_TASK_ACTION_COMMAND_STATUS_PREFIX}${taskStatus}`
    },
    handleHomeTaskActionCommand(task, command) {
      if (typeof command !== 'string') {
        return
      }
      if (command === HOME_TASK_ACTION_COMMAND_EDIT) {
        this.editHomeTask(task)
        return
      }
      if (command === HOME_TASK_ACTION_COMMAND_ARCHIVE) {
        this.toggleHomeTaskArchive(task, HOME_TASK_ARCHIVED_YES)
        return
      }
      if (command === HOME_TASK_ACTION_COMMAND_UNARCHIVE) {
        this.toggleHomeTaskArchive(task, HOME_TASK_ARCHIVED_NO)
        return
      }
      if (!command.startsWith(HOME_TASK_ACTION_COMMAND_STATUS_PREFIX)) {
        return
      }
      this.quickUpdateHomeTaskStatus(task, command.slice(HOME_TASK_ACTION_COMMAND_STATUS_PREFIX.length))
    },
    toggleHomeTaskArchive(task, isArchived) {
      if (this.homeTaskOperatingId > 0) {
        return
      }
      this.homeTaskOperatingId = Number(task.id || 0)
      this.homeTaskOperatingType = HOME_TASK_OPERATE_ARCHIVE
      homeTaskApi.HomeTaskArchiveToggle(this.homeTaskOperatingId, isArchived, (response) => {
        this.homeTaskOperatingId = 0
        this.homeTaskOperatingType = ''
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '归档状态更新失败')
          return
        }
        this.refreshAllHomeTaskList()
      })
    },
    deleteHomeTask(task) {
      if (this.homeTaskOperatingId > 0) {
        return
      }
      const taskId = Number(task?.id || 0)
      const taskName = String(task?.name || '').trim() || `#${taskId}`
      this.$confirm(
        `${HOME_TASK_DELETE_CONFIRM_MESSAGE_PREFIX}${taskName}${HOME_TASK_DELETE_CONFIRM_MESSAGE_SUFFIX}`,
        HOME_TASK_DELETE_CONFIRM_TITLE,
        {
          type: 'warning',
          confirmButtonText: '确认删除',
          cancelButtonText: '取消',
        }
      ).then(() => {
        this.homeTaskOperatingId = taskId
        this.homeTaskOperatingType = HOME_TASK_OPERATE_DELETE
        homeTaskApi.HomeTaskDelete(taskId, (response) => {
          this.homeTaskOperatingId = 0
          this.homeTaskOperatingType = ''
          if (!(response && response.ErrCode === 0)) {
            this.$helperNotify.error(response?.ErrMsg || '任务删除失败')
            return
          }
          this.$helperNotify.success(HOME_TASK_DELETE_SUCCESS_MESSAGE)
          this.refreshAllHomeTaskList()
        })
      }).catch(() => {})
    },
    convertHomeTaskDateToUnix(dateText) {
      const normalizedDateText = String(dateText || '').trim() || getTodayDateText()
      return Math.floor(new Date(`${normalizedDateText}T00:00:00`).getTime() / 1000)
    },
    getHomeTaskStatusTagType(taskStatus) {
      if (taskStatus === HOME_TASK_STATUS_DEVELOPING) {
        return 'success'
      }
      if (taskStatus === HOME_TASK_STATUS_SELF_TESTING || taskStatus === HOME_TASK_STATUS_TESTING) {
        return 'warning'
      }
      if (taskStatus === HOME_TASK_STATUS_SELF_TESTED) {
        return 'success'
      }
      if (taskStatus === HOME_TASK_STATUS_PENDING_INTEGRATION) {
        return 'info'
      }
      if (taskStatus === HOME_TASK_STATUS_INTEGRATING || taskStatus === HOME_TASK_STATUS_RELEASING) {
        return 'primary'
      }
      if (taskStatus === HOME_TASK_STATUS_ONLINE) {
        return 'info'
      }
      return ''
    },
    getHomeTaskActionButtonVariant(taskStatus) {
      return 'primary'
    },
    toggleFragmentExpand(taskId) {
      this.homeTaskExpandedFragments[taskId] = !this.homeTaskExpandedFragments[taskId]
    },
    getFragmentPreview(content, taskId) {
      const maxLength = 100
      if (!content) return ''
      const isExpanded = this.homeTaskExpandedFragments[taskId]
      if (isExpanded || content.length <= maxLength) {
        return content
      }
      return content.slice(0, maxLength) + '...'
    },
    isFragmentExpandable(content) {
      const maxLength = 100
      return content && content.length > maxLength
    },
  },
  components: {
    GitActionButton,
    SettingsDialog,
    HomeTaskReportSetting,
  },
}
</script>

<style scoped src="@/css/components/HomeTask.css"></style>
