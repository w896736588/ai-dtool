<template>
  <div v-if="showInitialSetupState" class="setup-state">
    <div class="setup-state__panel">
      <div class="setup-state__badge">DT</div>
      <div class="setup-state__eyebrow">First Setup</div>
      <h2 class="setup-state__title">还没有检测到任何配置</h2>
      <p class="setup-state__desc">
        当前界面先隐藏业务页面，避免默认兜底值造成误解。点击下面的按钮进入设置页，完成基础配置后再开始使用。
      </p>
      <pl-button type="primary" class="setup-state__button" @click="openSetupPage">
        点击设置
      </pl-button>
    </div>
  </div>
  <div v-else class="layout-container">
    <!-- 左侧菜单 -->
    <aside class="sidebar">
      <div class="sidebar-header">
        <span class="logo">DT</span>
        <span class="title">DevTools</span>
      </div>
      
      <el-menu
        :default-active="menuName"
        active-text-color="#3a7a3a"
        background-color="#f5f5f0"
        text-color="#5a5a5a"
        router
        class="sidebar-menu"
        @select="handleSelect"
      >
        <el-menu-item
          v-for="item in visibleNavigationItems"
          :key="item.index"
          :index="item.index"
          :class="['sidebar-menu__item', `sidebar-menu__item--${item.tone}`]"
        >
          <span class="sidebar-menu__glyph">
            <el-icon><component :is="item.icon" /></el-icon>
          </span>
          <span>{{ item.label }}</span>
        </el-menu-item>
      </el-menu>

      <!-- 底部工具栏 -->
      <div class="sidebar-footer">
        <el-tag v-if="ip" size="small" type="info" @click="copyIp()" style="cursor: pointer; margin-bottom: 8px;">
          {{ ip }}
        </el-tag>
        <div class="footer-buttons">
          <button type="button" class="footer-action footer-action--leaf" @click="OpenNewBlank()">
            <span class="footer-action__title">新页卡</span>
            <span class="footer-action__desc">快速打开一个独立工作页</span>
          </button>
          <button type="button" class="footer-action footer-action--mint" @click="drawerVisibleTools = true">
            <span class="footer-action__title">小工具</span>
            <span class="footer-action__desc">常用功能集中入口</span>
          </button>
          <button type="button" class="footer-action footer-action--sky" @click="openSshConnectionsDialog">
            <span class="footer-action__title">当前 SSH 连接数 {{ sshConnectionCount }}</span>
            <span class="footer-action__desc">查看活跃连接和当前命令</span>
          </button>
        </div>
        <pl-button v-if="loginInfo.dialog" size="small" @click="loginInfo.dialog = true">登录</pl-button>
      </div>
    </aside>

    <!-- 主内容区域 -->
    <main class="main-content">
      <div class="main-content__body">
        <div
          v-if="isDashboardRoute"
          class="home-dashboard-stage"
          @wheel="handleDashboardWheel"
        >
          <div
            class="home-dashboard-track"
            :class="{ 'home-dashboard-track--animating': homeDashboardAnimating }"
            :style="homeDashboardTrackStyle"
          >
            <section class="home-dashboard-screen home-dashboard-screen--command">
              <div class="main-content__view main-content__view--dashboard">
                <router-view v-slot="{ Component, route }" name="home">
                  <keep-alive>
                    <component :is="Component" ref="currentRef"/>
                  </keep-alive>
                </router-view>
              </div>
            </section>

            <section class="home-dashboard-screen home-dashboard-screen--task">
              <div ref="homeTaskPanelScroll" class="home-task-panel-scroll">
                <section class="home-task-panel">
                  <div class="home-task-panel__header">
                    <div class="home-task-panel__heading">
                      <div class="home-task-panel__eyebrow">Home Tasks</div>
                      <div class="home-task-panel__title">任务清单</div>
                      <div class="home-task-panel__desc">保持和命令快捷操作一致的轻量层级，随时记录、切换和清理任务</div>
                    </div>
                  </div>

                  <div class="home-task-toolbar">
                    <div class="home-task-toolbar__text">
                      <div class="home-task-toolbar__title">任务管理</div>
                      <div class="home-task-toolbar__desc">保留状态切换和归档能力，并支持基于活跃任务一键生成工作日报</div>
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
                        >
                          <div class="home-task-card__header">
                            <div>
                              <div class="home-task-card__title">{{ task.name }}</div>
                              <div class="home-task-card__meta">
                                <span>开始时间：{{ task.start_time_desc || '-' }}</span>
                                <span>最后操作：{{ task.last_operated_at_desc || '-' }}</span>
                              </div>
                            </div>
                            <el-tag size="small" effect="light" :type="getHomeTaskStatusTagType(task.task_status)">
                              {{ task.task_status }}
                            </el-tag>
                          </div>
                          <div v-if="task.remark" class="home-task-card__remark">
                            {{ task.remark }}
                          </div>
                          <div class="home-task-card__actions">
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
                        >
                          <div class="home-task-card__header">
                            <div>
                              <div class="home-task-card__title">{{ task.name }}</div>
                              <div class="home-task-card__meta">
                                <span>开始时间：{{ task.start_time_desc || '-' }}</span>
                                <span>最后操作：{{ task.last_operated_at_desc || '-' }}</span>
                              </div>
                            </div>
                            <el-tag size="small" effect="light" :type="getHomeTaskStatusTagType(task.task_status)">
                              {{ task.task_status }}
                            </el-tag>
                          </div>
                          <div v-if="task.remark" class="home-task-card__remark">
                            {{ task.remark }}
                          </div>
                          <div class="home-task-card__actions">
                            <el-dropdown
                              trigger="click"
                              :disabled="isHomeTaskBusy(task.id)"
                              @command="handleHomeTaskActionCommand(task, $event)"
                            >
                              <GitActionButton
                                compact
                                :loading="isHomeTaskBusy(task.id, HOME_TASK_OPERATE_STATUS) || isHomeTaskBusy(task.id, HOME_TASK_OPERATE_ARCHIVE)"
                                variant="info"
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
                </section>
              </div>
            </section>
          </div>

          <div class="home-dashboard-pager">
            <button
              type="button"
              class="home-dashboard-pager__dot"
              :class="{ 'home-dashboard-pager__dot--active': homeDashboardPageIndex === 0 }"
              @click="switchHomeDashboardPage(0)"
            />
            <button
              type="button"
              class="home-dashboard-pager__dot"
              :class="{ 'home-dashboard-pager__dot--active': homeDashboardPageIndex === 1 }"
              @click="switchHomeDashboardPage(1)"
            />
          </div>
        </div>
        <div v-else class="main-content__view">
          <router-view v-slot="{ Component, route }" name="home">
            <keep-alive>
              <component :is="Component" ref="currentRef"/>
            </keep-alive>
          </router-view>
        </div>
      </div>
    </main>
  </div>

  <el-drawer
    v-model="drawerVisibleTools"
    direction="rtl"
    size="90%"
    title="小工具"
  >
    <tools></tools>
  </el-drawer>

  <el-dialog v-model="loginInfo.dialog" title="登录" width="500">
    <el-form>
      <el-form-item :label-width="80" label="username">
        <el-input v-model="loginInfo.username" autocomplete="off"/>
      </el-form-item>
      <el-form-item :label-width="80" label="password">
        <el-input v-model="loginInfo.password" autocomplete="off" show-password/>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="loginInfo.dialog = false">取消</pl-button>
        <pl-button type="primary" @click="login">保存</pl-button>
      </div>
    </template>
  </el-dialog>

  <el-dialog
    v-model="sshConnectionsDialogVisible"
    title="当前SSH连接列表"
    width="82%"
    top="6vh"
    class="ssh-connections-dialog"
  >
    <div class="ssh-dialog-toolbar">
      <el-tag type="success" effect="light">连接数 {{ sshConnectionCount }}</el-tag>
      <pl-button size="small" type="primary" plain @click="refreshSshConnections(true)">刷新列表</pl-button>
    </div>
    <el-table
      v-loading="sshConnectionsLoading"
      :data="sshConnections"
      stripe
      border
      style="width: 100%"
      class="ssh-connections-table"
      max-height="62vh"
      empty-text="暂无活跃SSH连接"
    >
      <el-table-column prop="ssh_name" label="SSH" width="180" />
      <el-table-column prop="shell_client_id" label="客户端ID" width="220" />
      <el-table-column prop="current_command" label="当前命令" min-width="320" show-overflow-tooltip />
      <el-table-column prop="status" label="状态" width="120">
        <template #default="scope">
          <el-tag
            size="small"
            effect="light"
            :type="scope.row.status === 'busy' ? 'warning' : 'success'"
          >
            {{ scope.row.status || '-' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column prop="connect_time" label="连接开始时间" width="180" />
      <el-table-column prop="connect_seconds" label="连接时长(秒)" width="120" />
      <el-table-column prop="type" label="类型" width="120" />
    </el-table>
  </el-dialog>

  <el-dialog
    v-model="homeTaskDialogVisible"
    :title="homeTaskDialogTitle"
    width="720px"
    destroy-on-close
  >
    <el-form label-width="88px" class="home-task-form" @submit.prevent>
      <el-row :gutter="12">
        <el-col :xs="24" :sm="24" :md="10">
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
        <el-col :xs="24" :sm="12" :md="7">
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
        <el-col :xs="24" :sm="12" :md="7">
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
          <el-form-item label="任务备注">
            <el-input
              v-model="homeTaskForm.remark"
              type="textarea"
              :rows="4"
              maxlength="300"
              show-word-limit
              placeholder="补充当前任务背景、阻塞点或上线说明"
            />
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
    title="工作日报 AI 设置"
    width="760px"
    @closed="refreshHomeTaskReportSettings"
  >
    <HomeTaskReportSetting ref="homeTaskReportSetting" />
  </SettingsDialog>
</template>

<script>
import base from '../utils/base'
import mod from '../utils/module'
import socket from '../utils/base/socket'
import store from "@/utils/base/store";
import format from "@/utils/base/format"
import Clipboard from "clipboard";
import copy from "@/utils/base/copy"
import module from "@/utils/module"
import baseApi from '@/utils/base/base_api'
import sshSet from '@/utils/base/ssh_set'
import homeTaskApi from '@/utils/base/home_task'
import setApi from '@/utils/base/git_set'
const {
  HOME_DASHBOARD_PAGE_SWITCH_HOT_ZONE_WIDTH,
  isHomeDashboardPageSwitchHotZone,
  shouldBlockHomeDashboardPageSwitch,
} = require('@/utils/home_dashboard_wheel.cjs')
import Tools from "@/components/Tools.vue";
import Markdown from '@/components/Markdown.vue'
import GitActionButton from "@/components/base/GitActionButton.vue";
import SettingsDialog from '@/components/base/SettingsDialog.vue'
import HomeTaskReportSetting from '@/components/set/home_task_report.vue'
import { 
  HomeFilled,
  Coin,
  Setting,
  Folder,
  Link,
  Document,
  Memo,
  Box,
  Connection,
  Monitor,
  Tools as ToolsIcon
} from "@element-plus/icons-vue";

// SSH_CONNECTION_REFRESH_INTERVAL_MS 统一控制 SSH 连接轮询周期。
const SSH_CONNECTION_REFRESH_INTERVAL_MS = 5000
// HOME_TASK_TAB_* 用于区分任务弹窗内的标签页。
const HOME_TASK_TAB_ACTIVE = 'active'
const HOME_TASK_TAB_ARCHIVED = 'archived'
// HOME_ROUTE_DASHBOARD 标识首页路由路径。
const HOME_ROUTE_DASHBOARD = '/Dashboard'
// HOME_ROUTE_SET 标识设置页路由路径。
const HOME_ROUTE_SET = '/Set'
// HOME_SETUP_EMPTY_KEYS 用于判断配置文件是否仍然保持空白。
const HOME_SETUP_EMPTY_KEYS = [
  'db_path_raw',
  'db_name_raw',
  'memory_db_path_raw',
  'memory_db_name_raw',
  'webkit_driver_path',
  'webkit_data_path',
  'webkit_download_path',
]
// HOME_TASK_ARCHIVED_* 对应后端归档状态常量。
const HOME_TASK_ARCHIVED_NO = 0
const HOME_TASK_ARCHIVED_YES = 1
// HOME_TASK_STATUS_* 与后端状态常量保持一致。
const HOME_TASK_STATUS_TODO = '待开始'
const HOME_TASK_STATUS_DEVELOPING = '开发中'
const HOME_TASK_STATUS_SELF_TESTING = '自测中'
const HOME_TASK_STATUS_INTEGRATING = '对接中'
const HOME_TASK_STATUS_TESTING = '测试中'
const HOME_TASK_STATUS_RELEASING = '上线中'
const HOME_TASK_STATUS_ONLINE = '已上线'
// HOME_TASK_OPERATE_* 标识当前任务操作类型，便于按钮精确展示 loading。
const HOME_TASK_OPERATE_SAVE = 'save'
const HOME_TASK_OPERATE_STATUS = 'status'
const HOME_TASK_OPERATE_ARCHIVE = 'archive'
const HOME_TASK_OPERATE_DELETE = 'delete'
// HOME_TASK_ACTION_COMMAND_* 用于统一处理任务卡片下拉操作。
const HOME_TASK_ACTION_COMMAND_EDIT = 'edit'
const HOME_TASK_ACTION_COMMAND_ARCHIVE = 'archive'
const HOME_TASK_ACTION_COMMAND_UNARCHIVE = 'unarchive'
// HOME_TASK_DELETE_CONFIRM_* 统一任务删除确认文案，避免危险操作提示分散。
const HOME_TASK_DELETE_CONFIRM_TITLE = '确认删除'
const HOME_TASK_DELETE_CONFIRM_MESSAGE_PREFIX = '确定要删除任务“'
const HOME_TASK_DELETE_CONFIRM_MESSAGE_SUFFIX = '”吗？该操作不可恢复。'
const HOME_TASK_DELETE_SUCCESS_MESSAGE = '任务已删除'
// HOME_TASK_EDIT_BUTTON_TEXT 统一定义任务编辑按钮文案。
const HOME_TASK_EDIT_BUTTON_TEXT = '编辑任务'
// HOME_TASK_DAILY_REPORT_BUTTON_TEXT 统一定义日报生成按钮文案。
const HOME_TASK_DAILY_REPORT_BUTTON_TEXT = 'AI 生成工作日报'
// HOME_TASK_DAILY_REPORT_* 统一定义日报生成提示文案。
const HOME_TASK_DAILY_REPORT_SUCCESS_MESSAGE = '工作日报已生成并保存到记忆'
const HOME_TASK_DAILY_REPORT_FAILED_MESSAGE = '工作日报生成失败'
// HOME_TASK_ACTION_COMMAND_STATUS_PREFIX 标识状态切换指令前缀。
const HOME_TASK_ACTION_COMMAND_STATUS_PREFIX = 'status:'
// HOME_DASHBOARD_PAGE_* 标识首页双屏结构中的页索引。
const HOME_DASHBOARD_PAGE_COMMAND = 0
const HOME_DASHBOARD_PAGE_TASK = 1
// HOME_DASHBOARD_PAGE_TOTAL 表示首页双屏总页数。
const HOME_DASHBOARD_PAGE_TOTAL = 2
// HOME_DASHBOARD_WHEEL_SWITCH_THRESHOLD 控制翻屏触发的滚轮阈值，避免轻微触控板抖动误切屏。
const HOME_DASHBOARD_WHEEL_SWITCH_THRESHOLD = 24
// HOME_DASHBOARD_ANIMATION_DURATION_MS 与 CSS 动画时长保持一致。
const HOME_DASHBOARD_ANIMATION_DURATION_MS = 560
// HOME_TASK_EMPTY_START_DATE 表示未设置开始日期。
const HOME_TASK_EMPTY_START_DATE = ''
// HOME_TASK_EMPTY_REMARK 表示未填写备注。
const HOME_TASK_EMPTY_REMARK = ''
// HOME_TASK_DEFAULT_FORM 定义任务表单的初始值。
const HOME_TASK_DEFAULT_FORM = {
  id: 0,
  name: '',
  task_status: HOME_TASK_STATUS_TODO,
  remark: HOME_TASK_EMPTY_REMARK,
  start_date: HOME_TASK_EMPTY_START_DATE,
}
// HOME_TASK_STATUS_OPTIONS 用于渲染状态选择和快捷状态按钮。
const HOME_TASK_STATUS_OPTIONS = [
  HOME_TASK_STATUS_TODO,
  HOME_TASK_STATUS_DEVELOPING,
  HOME_TASK_STATUS_SELF_TESTING,
  HOME_TASK_STATUS_INTEGRATING,
  HOME_TASK_STATUS_TESTING,
  HOME_TASK_STATUS_RELEASING,
  HOME_TASK_STATUS_ONLINE,
]

export default {
  data() {
    return {
      drawerVisibleTools: false,
      drawerVisibleMarkdown: false,
      loginInfo: {
        dialog: false,
        username: 'default',
        password: '111',
      },
      menuKeyStore: 'lastMenuName.v2',
      menuName: '/Dashboard',
      minHeightMap: {},
      showShellMap: ['/Git', '/Consumer', '/WechatKefu'],
      tags: [],
      showTextarea: true,
      shellShowResult: "",
      sshMapList: [],
      xtermMapList: [],
      runCommand: '',
      openModuleList: [],
      term: '',
      lastShellInfo: {
        sshId: "",
        business: "",
        lastShellInfo: "",
      },
      ip: '',
      sshConnectionCount: 0,
      sshConnections: [],
      sshConnectionsDialogVisible: false,
      sshConnectionsLoading: false,
      sshConnectionTimer: null,
      setupGateDismissed: false,
      runtimeConfigSnapshot: {
        db_path_raw: '',
        db_name_raw: '',
        memory_db_path_raw: '',
        memory_db_name_raw: '',
        webkit_driver_path: '',
        webkit_data_path: '',
        webkit_download_path: '',
        db_is_git_repo: false,
        memory_db_is_git_repo: false,
      },
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
      homeDashboardPageIndex: HOME_DASHBOARD_PAGE_COMMAND,
      homeDashboardAnimating: false,
      homeTaskLoadingActive: false,
      homeTaskLoadingArchived: false,
      homeTaskGeneratingDailyReport: false,
      homeTaskSaving: false,
      homeTaskOperatingId: 0,
      homeTaskOperatingType: '',
      homeTaskActiveList: [],
      homeTaskArchivedList: [],
      homeTaskStatusOptions: HOME_TASK_STATUS_OPTIONS,
      homeTaskForm: {
        ...HOME_TASK_DEFAULT_FORM,
      },
    }
  },
  computed: {
    // showInitialSetupState 在完全未配置且尚未点击设置时显示引导卡片。
    showInitialSetupState() {
      return !this.setupGateDismissed && !this.hasAnyUserConfig
    },
    // hasAnyUserConfig 依据 ini 原始值判断是否已有任何显式配置。
    hasAnyUserConfig() {
      const hasTextConfig = HOME_SETUP_EMPTY_KEYS.some((key) => String(this.runtimeConfigSnapshot[key] || '').trim() !== '')
      return hasTextConfig || !!this.runtimeConfigSnapshot.db_is_git_repo || !!this.runtimeConfigSnapshot.memory_db_is_git_repo
    },
    // visibleNavigationItems 统一定义左侧导航条目与图标风格。
    visibleNavigationItems() {
      return [
        { index: '/Dashboard', label: '首页', icon: HomeFilled, tone: 'earth' },
        { index: '/Redis', label: 'Redis', icon: Coin, tone: 'rose', gate: 'redis' },
        { index: '/Supervisor', label: 'Supervisor', icon: Setting, tone: 'amber', gate: 'supervisor' },
        { index: '/Git', label: 'Git', icon: Folder, tone: 'forest', gate: 'git' },
        { index: '/CommonActions', label: '常用操作', icon: ToolsIcon, tone: 'mint', gate: 'tools' },
        { index: '/Link', label: '自定义网页', icon: Link, tone: 'sky', gate: 'login' },
        { index: '/Variable', label: '自定义脚本', icon: Document, tone: 'violet', gate: 'variable' },
        { index: '/MemoryFragment', label: '知识片段', icon: Memo, tone: 'sea', gate: 'memory_fragment' },
        { index: '/Docker', label: 'Docker', icon: Box, tone: 'ocean', gate: 'docker' },
        { index: '/Api', label: '接口开发', icon: Connection, tone: 'teal', gate: 'api' },
        { index: '/shellout', label: '终端输出', icon: Monitor, tone: 'slate', gate: 'shellout' },
        { index: '/Set', label: '配置', icon: Setting, tone: 'sun' },
      ].filter((item) => !item.gate || this.checkModuleOpen(item.gate))
    },
    // isDashboardRoute 控制任务清单只在首页底部展示。
    isDashboardRoute() {
      return this.$route.path === HOME_ROUTE_DASHBOARD
    },
    // homeDashboardTrackStyle 控制首页双屏容器的整体位移。
    homeDashboardTrackStyle() {
      return {
        transform: `translate3d(0, -${this.homeDashboardPageIndex * 50}%, 0)`,
      }
    },
    // homeTaskDialogTitle 统一控制新增和编辑弹窗标题。
    homeTaskDialogTitle() {
      return this.homeTaskForm.id > 0 ? '编辑任务' : '新增任务'
    },
  },
  watch: {
    // 当用户切回首页时主动刷新任务，避免跨页面停留后数据过期。
    '$route.path'(newPath) {
      if (newPath === HOME_ROUTE_SET) {
        this.setupGateDismissed = true
      }
      this.loadRuntimeConfigSnapshot()
      if (newPath !== HOME_ROUTE_DASHBOARD) {
        this.homeDashboardPageIndex = HOME_DASHBOARD_PAGE_COMMAND
        return
      }
      this.homeDashboardPageIndex = HOME_DASHBOARD_PAGE_COMMAND
      this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
      this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
    },
  },
  created() {
    window.handleCopy = copy.handleCopy;
  },
  mounted: function () {
    let _that = this
    _that.openModuleList = module.GetOpenModuleList()
    base.BaseLogin(_that.loginInfo.username, _that.loginInfo.password, function (response) {
      if (response.ErrCode === 0) {
        store.setStore('token', response.Data.token)
      } else {
        _that.$helperNotify.error('登录失败')
      }
    })
    this.forceIp(false)
    this.refreshSshConnections(false)
    this.loadRuntimeConfigSnapshot()
    this.sshConnectionTimer = setInterval(() => {
      this.refreshSshConnections(false)
    }, SSH_CONNECTION_REFRESH_INTERVAL_MS)
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
    this.menuName = this.$helperStore.getStore(this.menuKeyStore)
    if (this.$route.path === HOME_ROUTE_SET) {
      this.setupGateDismissed = true
    }
    if (this.$route.path !== this.menuName && this.menuName != null) {
      this.$router.push(this.menuName)
    }
    window.addEventListener('resize', function () {});
  },
  provide() {
    return {
      showTerminal: this.showTerminal,
      resizeTerminal: this.resizeTerminal,
    };
  },
  methods: {
    // loadRuntimeConfigSnapshot 拉取首页空状态判断所需的原始配置快照。
    loadRuntimeConfigSnapshot() {
      setApi.MemoryConfigGet((response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          return
        }
        this.runtimeConfigSnapshot = {
          db_path_raw: response.Data.db_path_raw || '',
          db_name_raw: response.Data.db_name_raw || '',
          memory_db_path_raw: response.Data.memory_db_path_raw || '',
          memory_db_name_raw: response.Data.memory_db_name_raw || '',
          webkit_driver_path: response.Data.webkit_driver_path || '',
          webkit_data_path: response.Data.webkit_data_path || '',
          webkit_download_path: response.Data.webkit_download_path || '',
          db_is_git_repo: !!response.Data.db_is_git_repo,
          memory_db_is_git_repo: !!response.Data.memory_db_is_git_repo,
        }
      })
    },
    // openSetupPage 进入设置页并解除首页首次配置挡板。
    openSetupPage() {
      this.setupGateDismissed = true
      if (this.$route.path !== HOME_ROUTE_SET) {
        this.$router.push(HOME_ROUTE_SET)
      }
    },
    OpenNewBlank: function () {
      window.open(window.location.href, '_blank');
    },
    copyIp: function () {
      let index = copy.SetCopyContent(this.ip)
      copy.handleCopy(index)
    },
    forceIp: function (forceIp) {
      let _that = this
      baseApi.Ip({}, function (ip) {
        _that.ip = ip
      }, forceIp)
    },
    login: function () {
      let _that = this
      base.BaseLogin(_that.loginInfo.username, _that.loginInfo.password, function (response) {
        if (response.ErrCode === 0) {
          store.setStore('token', response.Data.token)
          window.location.reload()
        } else {
          _that.$helperNotify.error('登录失败')
        }
      })
    },
    checkModuleOpen: function (moduleName) {
      return this.openModuleList.includes(moduleName)
    },
    resetConn: function () {
      store.removeStore('Unikey')
    },
    showNotify: function (notifyList) {
      this.tags = notifyList
    },
    showTerminal(uniqueKey) {
      this.lastShellInfo.uniqueKey = uniqueKey
      this.shellSetShowResult(uniqueKey)
      this.shellDrawerScrollTop(2000)
    },
    resizeTerminal: function () {},
    shellSetShowResult: function (uniqueKey) {
      for (let i in this.sshMapList) {
        if (this.sshMapList[i].uniqueKey === uniqueKey) {
          this.shellShowResult = this.sshMapList[i].shellResult
        }
      }
    },
    shellDrawerScrollTop: function (milliseconds) {
      setTimeout(function () {
        let obj = document.getElementById('showShellResult')
        if (obj) {
          obj.scrollTop = obj.scrollHeight + 200
        }
      }, milliseconds)
    },
    openSshConnectionsDialog() {
      this.sshConnectionsDialogVisible = true
      this.refreshSshConnections(true)
    },
    handleHomeTaskTabChange(tabName) {
      if (tabName === HOME_TASK_TAB_ACTIVE) {
        this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
        return
      }
      this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
    },
    // handleDashboardWheel 只在首页双屏区域接管滚轮，实现整屏切换。
    handleDashboardWheel(event) {
      if (!this.isDashboardRoute || this.homeDashboardAnimating) {
        return
      }
      const deltaY = Number(event.deltaY || 0)
      if (Math.abs(deltaY) < HOME_DASHBOARD_WHEEL_SWITCH_THRESHOLD) {
        return
      }
      const currentTarget = event.currentTarget
      const currentTargetRect = currentTarget instanceof HTMLElement ? currentTarget.getBoundingClientRect() : null
      const isRightHotZone = isHomeDashboardPageSwitchHotZone(event.clientX, currentTargetRect, HOME_DASHBOARD_PAGE_SWITCH_HOT_ZONE_WIDTH)
      // 命中首页最右侧热区时，优先允许整屏翻页，不再被命令区全宽滚动容器拦截。
      if (!isRightHotZone && shouldBlockHomeDashboardPageSwitch(event.target, deltaY, currentTarget)) {
        return
      }
      if (this.homeDashboardPageIndex === HOME_DASHBOARD_PAGE_COMMAND) {
        if (deltaY > 0) {
          event.preventDefault()
          this.switchHomeDashboardPage(HOME_DASHBOARD_PAGE_TASK)
        }
        return
      }
      const panelScroll = this.$refs.homeTaskPanelScroll
      if (!(panelScroll instanceof HTMLElement)) {
        return
      }
      const scrollTop = panelScroll.scrollTop
      const maxScrollTop = Math.max(panelScroll.scrollHeight - panelScroll.clientHeight, 0)
      if (deltaY < 0 && scrollTop <= 0) {
        event.preventDefault()
        this.switchHomeDashboardPage(HOME_DASHBOARD_PAGE_COMMAND)
        return
      }
      if (deltaY > 0 && scrollTop >= maxScrollTop) {
        event.preventDefault()
      }
    },
    // switchHomeDashboardPage 切换首页双屏页码，并锁定动画期间的重复操作。
    switchHomeDashboardPage(pageIndex) {
      const nextPageIndex = Math.min(Math.max(pageIndex, HOME_DASHBOARD_PAGE_COMMAND), HOME_DASHBOARD_PAGE_TOTAL - 1)
      if (nextPageIndex === this.homeDashboardPageIndex) {
        return
      }
      this.homeDashboardAnimating = true
      this.homeDashboardPageIndex = nextPageIndex
      window.setTimeout(() => {
        this.homeDashboardAnimating = false
      }, HOME_DASHBOARD_ANIMATION_DURATION_MS)
    },
    // loadHomeTaskList 按归档状态刷新任务列表，避免前端本地状态和后端脱节。
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
    // refreshAllHomeTaskList 统一刷新活跃和归档列表，避免各操作分散重复调用。
    refreshAllHomeTaskList() {
      this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
      this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
    },
    resetHomeTaskForm() {
      this.homeTaskForm = {
        ...HOME_TASK_DEFAULT_FORM,
      }
    },
    // openCreateHomeTaskDialog 打开新增任务弹窗，并重置为默认表单。
    openCreateHomeTaskDialog() {
      this.resetHomeTaskForm()
      this.homeTaskDialogVisible = true
    },
    // openHomeTaskReportSettingsDialog 打开工作日报 AI 设置弹窗。
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
    // refreshHomeTaskReportSettings 在弹窗关闭时兜底刷新设置组件状态。
    refreshHomeTaskReportSettings() {
      if (this.$refs.homeTaskReportSetting && this.$refs.homeTaskReportSetting.loadConfig) {
        this.$refs.homeTaskReportSetting.loadConfig()
      }
    },
    // generateHomeTaskDailyReport 调用后端基于活跃任务生成日报并写入记忆。
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
    // closeHomeTaskDialog 关闭任务弹窗，并清空临时表单内容。
    closeHomeTaskDialog() {
      this.homeTaskDialogVisible = false
      this.resetHomeTaskForm()
    },
    editHomeTask(task) {
      this.homeTaskForm = {
        id: Number(task.id || 0),
        name: task.name || '',
        task_status: task.task_status || HOME_TASK_STATUS_TODO,
        remark: task.remark || HOME_TASK_EMPTY_REMARK,
        start_date: task.start_time_desc || HOME_TASK_EMPTY_START_DATE,
      }
      this.homeTaskDialogVisible = true
    },
    // saveHomeTask 保存表单任务，新增和编辑共用同一个入口。
    saveHomeTask() {
      if (this.homeTaskSaving) {
        return
      }
      const taskName = String(this.homeTaskForm.name || '').trim()
      if (!taskName) {
        this.$helperNotify.error('任务名称不能为空')
        return
      }
      this.homeTaskSaving = true
      this.homeTaskOperatingType = HOME_TASK_OPERATE_SAVE
      homeTaskApi.HomeTaskSave({
        id: Number(this.homeTaskForm.id || 0),
        name: taskName,
        task_status: this.homeTaskForm.task_status,
        remark: String(this.homeTaskForm.remark || '').trim(),
        start_time: this.convertHomeTaskDateToUnix(this.homeTaskForm.start_date),
      }, (response) => {
        this.homeTaskSaving = false
        this.homeTaskOperatingType = ''
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '任务保存失败')
          return
        }
        this.$helperNotify.success(this.homeTaskForm.id > 0 ? '任务已更新' : '任务已创建')
        this.closeHomeTaskDialog()
        this.refreshAllHomeTaskList()
      })
    },
    // isHomeTaskBusy 判断指定任务是否正处于某个操作中，避免多个危险操作并发触发。
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
        this.refreshAllHomeTaskList()
      })
    },
    // buildHomeTaskStatusCommand 生成状态切换菜单命令，避免状态值和其它动作混淆。
    buildHomeTaskStatusCommand(taskStatus) {
      return `${HOME_TASK_ACTION_COMMAND_STATUS_PREFIX}${taskStatus}`
    },
    // handleHomeTaskActionCommand 统一处理任务卡片上的下拉操作。
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
    // deleteHomeTask 删除首页任务，使用确认弹窗降低误删风险。
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
      if (!dateText) {
        return 0
      }
      return Math.floor(new Date(`${dateText}T00:00:00`).getTime() / 1000)
    },
    getHomeTaskStatusTagType(taskStatus) {
      if (taskStatus === HOME_TASK_STATUS_DEVELOPING) {
        return 'success'
      }
      if (taskStatus === HOME_TASK_STATUS_SELF_TESTING || taskStatus === HOME_TASK_STATUS_TESTING) {
        return 'warning'
      }
      if (taskStatus === HOME_TASK_STATUS_INTEGRATING || taskStatus === HOME_TASK_STATUS_RELEASING) {
        return 'primary'
      }
      if (taskStatus === HOME_TASK_STATUS_ONLINE) {
        return 'info'
      }
      return ''
    },
    // getHomeTaskActionButtonVariant 根据当前任务状态返回操作按钮视觉类型。
    getHomeTaskActionButtonVariant(taskStatus) {
      if (taskStatus === HOME_TASK_STATUS_DEVELOPING) {
        return 'primary'
      }
      if (taskStatus === HOME_TASK_STATUS_SELF_TESTING || taskStatus === HOME_TASK_STATUS_TESTING) {
        return 'warning'
      }
      if (taskStatus === HOME_TASK_STATUS_INTEGRATING || taskStatus === HOME_TASK_STATUS_RELEASING) {
        return 'info'
      }
      if (taskStatus === HOME_TASK_STATUS_ONLINE) {
        return 'info'
      }
      return ''
    },
    refreshSshConnections(showLoading) {
      if (showLoading) {
        this.sshConnectionsLoading = true
      }
      const _that = this
      sshSet.SshList(function (sshResponse) {
        const sshNameMap = {}
        if (sshResponse && sshResponse.ErrCode === 0 && Array.isArray(sshResponse.Data)) {
          sshResponse.Data.forEach(item => {
            sshNameMap[String(item.id)] = item.name || `#${item.id}`
          })
        }
        sshSet.GetConnections(function (connResponse) {
          if (_that.sshConnectionsLoading) {
            _that.sshConnectionsLoading = false
          }
          if (!(connResponse && connResponse.ErrCode === 0)) {
            _that.sshConnectionCount = 0
            _that.sshConnections = []
            return
          }
          const list = Array.isArray(connResponse.Data?.connections) ? connResponse.Data.connections : []
          const normalized = list.map(conn => {
            const shellClientId = String(conn.shell_client_id || '')
            const sshId = shellClientId.split('#')[0]
            return {
              ...conn,
              ssh_name: sshNameMap[sshId] || `#${sshId || '-'}`
            }
          })
          _that.sshConnectionCount = normalized.length
          _that.sshConnections = normalized
        })
      })
    },
    handleSelect(key, keyPath) {
      let _that = this
      if (keyPath[0].indexOf('Doc-') >= 0) {
        return
      }
      if (keyPath[0].indexOf('Ignore-') >= 0) {
        return;
      }
      if (key === HOME_ROUTE_SET) {
        this.setupGateDismissed = true
      }
      this.menuName = keyPath[0]
      this.$helperStore.setStore(_that.menuKeyStore, this.menuName)
    },
  },
  beforeUnmount() {
    if (this.sshConnectionTimer) {
      clearInterval(this.sshConnectionTimer)
      this.sshConnectionTimer = null
    }
  },
  components: {
    HomeFilled,
    Coin,
    Setting,
    Folder,
    Link,
    Document,
    Memo,
    Box,
    Connection,
    Monitor,
    ToolsIcon,
    GitActionButton,
    SettingsDialog,
    HomeTaskReportSetting,
    Markdown,
    Tools,
    Clipboard,
  },
}
</script>

<style scoped>
.setup-state {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 100vh;
  padding: 24px;
  background:
    radial-gradient(circle at top left, rgba(188, 214, 180, 0.32), transparent 28%),
    radial-gradient(circle at bottom right, rgba(197, 223, 236, 0.36), transparent 24%),
    linear-gradient(145deg, #f3f7ef 0%, #fcfdf9 56%, #f6f4eb 100%);
}

.setup-state__panel {
  width: min(520px, 100%);
  padding: 38px 34px;
  border: 1px solid rgba(120, 145, 114, 0.16);
  border-radius: 28px;
  background: rgba(255, 255, 255, 0.92);
  box-shadow: 0 24px 80px rgba(72, 88, 68, 0.12);
  text-align: center;
}

.setup-state__badge {
  width: 58px;
  height: 58px;
  margin: 0 auto 16px;
  border-radius: 18px;
  display: grid;
  place-items: center;
  font-size: 20px;
  font-weight: 700;
  letter-spacing: 0.08em;
  color: #fdfdf8;
  background: linear-gradient(145deg, #455b41 0%, #73906f 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.18);
}

.setup-state__eyebrow {
  font-size: 11px;
  letter-spacing: 0.22em;
  text-transform: uppercase;
  color: #80927b;
}

.setup-state__title {
  margin: 14px 0 10px;
  font-size: 28px;
  line-height: 1.25;
  color: #324130;
}

.setup-state__desc {
  margin: 0 auto;
  max-width: 360px;
  font-size: 14px;
  line-height: 1.85;
  color: #667463;
}

.setup-state__button {
  margin-top: 24px;
  min-width: 150px;
}

.layout-container {
  display: flex;
  height: 100vh;
  width: 100%;
  background-color: #f8f8f5;
}

.sidebar {
  width: 140px;
  background: linear-gradient(180deg, #f6f7f1 0%, #f1f4ed 100%);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  border-right: 1px solid #e4e8db;
  box-shadow: inset -1px 0 0 rgba(255, 255, 255, 0.7);
}

.sidebar-header {
  height: 62px;
  display: flex;
  align-items: center;
  gap: 10px;
  padding: 0 14px;
  border-bottom: 1px solid #e4e8db;
}

.logo {
  width: 34px;
  height: 34px;
  border-radius: 12px;
  display: grid;
  place-items: center;
  font-size: 13px;
  font-weight: 700;
  letter-spacing: 0.1em;
  color: #f9fbf7;
  background: linear-gradient(145deg, #445943 0%, #728b6f 100%);
  box-shadow: 0 8px 18px rgba(71, 92, 70, 0.16);
}

.title {
  color: #40503e;
  font-size: 16px;
  font-weight: 600;
  letter-spacing: 0.02em;
}

.sidebar-menu {
  flex: 1;
  border-right: none;
  overflow-y: auto;
}

.sidebar-menu:not(.el-menu--collapse) {
  width: 140px;
}

.sidebar-footer {
  padding: 12px 10px 14px;
  border-top: 1px solid #e4e8db;
  display: flex;
  flex-direction: column;
  align-items: stretch;
  background: linear-gradient(180deg, rgba(246, 248, 242, 0.86) 0%, rgba(242, 245, 238, 0.96) 100%);
}

.footer-buttons {
  display: flex;
  flex-direction: column;
  width: 100%;
  gap: 10px;
  margin-bottom: 8px;
}

.footer-action {
  width: 100%;
  padding: 10px 12px;
  border: 1px solid rgba(120, 140, 111, 0.14);
  border-radius: 14px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.97) 0%, rgba(247, 249, 243, 0.98) 100%);
  box-shadow: 0 10px 24px rgba(119, 137, 112, 0.08);
  cursor: pointer;
  text-align: left;
  transition: transform 0.18s ease, box-shadow 0.18s ease, border-color 0.18s ease;
}

.footer-action:hover {
  transform: translateY(-1px);
  box-shadow: 0 14px 28px rgba(119, 137, 112, 0.12);
}

.footer-action:active {
  transform: translateY(0);
}

.footer-action__title {
  display: block;
  font-size: 13px;
  font-weight: 600;
  color: #3f5040;
}

.footer-action__desc {
  display: block;
  margin-top: 4px;
  font-size: 11px;
  line-height: 1.5;
  color: #73806f;
}

.footer-action--leaf:hover {
  border-color: rgba(87, 126, 80, 0.28);
}

.footer-action--mint:hover {
  border-color: rgba(56, 128, 109, 0.28);
}

.footer-action--sky:hover {
  border-color: rgba(53, 119, 166, 0.28);
}

.footer-action:focus-visible {
  outline: 2px solid rgba(83, 123, 77, 0.24);
  outline-offset: 2px;
}

.footer-buttons :deep(.git-action-button) {
  width: 100%;
  justify-content: center;
}

.main-content {
  flex: 1;
  overflow: auto;
  display: flex;
  flex-direction: column;
  min-height: 0;
  background-color: #fafaf7;
  height: 100%;
  padding: 20px;
  box-sizing: border-box;
}

.main-content__body {
  /* 让路由页面拿到稳定的可用高度，避免子页面 100% 高度失效。
     Keep a continuous height chain for routed views so child pages can truly fill the viewport. */
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
}

.main-content__view {
  /* 非首页路由也需要继承剩余高度，接口开发页才不会在底部露白。
     Non-dashboard routes must stretch to the remaining height to avoid bottom whitespace. */
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
}

.main-content__view > * {
  flex: 1;
  min-height: 0;
}

.main-content__view--dashboard {
  height: 100%;
  overflow: auto;
}

.home-dashboard-stage {
  position: relative;
  height: calc(100vh - 40px);
  overflow: hidden;
  border-radius: 24px;
  background:
    radial-gradient(circle at top right, rgba(168, 194, 149, 0.18), transparent 32%),
    linear-gradient(180deg, #fdfdf9 0%, #f5f7f1 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.7);
}

.home-dashboard-track {
  height: 200%;
  will-change: transform;
}

.home-dashboard-track--animating {
  transition: transform 0.56s cubic-bezier(0.22, 1, 0.36, 1);
}

.home-dashboard-screen {
  height: 50%;
  padding: 10px;
  box-sizing: border-box;
}

.home-dashboard-screen--command {
  display: flex;
  flex-direction: column;
}

.home-dashboard-screen--task {
  position: relative;
  display: flex;
  min-height: 0;
}

.home-dashboard-pager {
  position: absolute;
  top: 50%;
  right: 16px;
  z-index: 2;
  display: flex;
  flex-direction: column;
  gap: 10px;
  transform: translateY(-50%);
}

.home-dashboard-pager__dot {
  width: 10px;
  height: 34px;
  border: none;
  border-radius: 999px;
  background: rgba(104, 123, 96, 0.2);
  cursor: pointer;
  transition: transform 0.25s ease, background-color 0.25s ease;
}

.home-dashboard-pager__dot--active {
  background: linear-gradient(180deg, #6f8e67 0%, #89a27c 100%);
  transform: scale(1.05);
}

.home-task-panel-scroll {
  height: 100%;
  flex: 1;
  min-height: 0;
  overflow: hidden;
  padding-right: 12px;
}

.home-task-panel {
  /* 任务清单卡片限制在当前一屏内，避免整块内容继续向下滚动。
     Keep the task card within one screen so the whole panel does not scroll vertically. */
  height: 100%;
  max-height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  padding: 20px 22px 22px;
  border: 1px solid #e5ebde;
  border-radius: 20px;
  background: linear-gradient(180deg, #fdfdfb 0%, #f8faf5 100%);
  box-shadow: 0 16px 36px rgba(138, 154, 126, 0.08);
}

.home-task-panel__header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 16px;
}

.home-task-panel__heading {
  display: flex;
  flex-direction: column;
}

.home-task-panel__eyebrow {
  font-size: 11px;
  line-height: 1;
  letter-spacing: 0.16em;
  text-transform: uppercase;
  color: #93a18f;
}

.home-task-panel__title {
  margin-top: 8px;
  font-size: 22px;
  font-weight: 600;
  color: #3d4b3d;
}

.home-task-panel__desc {
  margin-top: 8px;
  font-size: 13px;
  color: #72816f;
  line-height: 1.6;
}

/* 覆盖 Element Plus 菜单样式 */
.sidebar-menu {
  padding: 6px 0;
}

.sidebar-menu .el-menu-item {
  position: relative;
  height: 46px;
  line-height: 46px;
  margin: 3px 8px;
  border-radius: 14px;
  padding-left: 10px !important;
  transition: background-color 0.22s ease, transform 0.22s ease, box-shadow 0.22s ease;
}

.sidebar-menu .el-menu-item:hover {
  background-color: #edf4ea !important;
  transform: translateX(2px);
}

.sidebar-menu .el-menu-item.is-active {
  background: linear-gradient(90deg, rgba(114, 145, 102, 0.16) 0%, rgba(114, 145, 102, 0.08) 100%) !important;
  color: #38573d !important;
  box-shadow: inset 0 0 0 1px rgba(114, 145, 102, 0.14);
}

.sidebar-menu__glyph {
  width: 28px;
  height: 28px;
  margin-right: 10px;
  border-radius: 10px;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: inset 0 0 0 1px rgba(92, 112, 83, 0.08);
}

.sidebar-menu .el-menu-item .el-icon {
  margin-right: 0;
  font-size: 15px;
}

.sidebar-menu__item--earth .sidebar-menu__glyph { color: #506947; }
.sidebar-menu__item--rose .sidebar-menu__glyph { color: #b15457; }
.sidebar-menu__item--amber .sidebar-menu__glyph { color: #a37327; }
.sidebar-menu__item--forest .sidebar-menu__glyph { color: #497451; }
.sidebar-menu__item--mint .sidebar-menu__glyph { color: #33806d; }
.sidebar-menu__item--sky .sidebar-menu__glyph { color: #3577a6; }
.sidebar-menu__item--violet .sidebar-menu__glyph { color: #6955a6; }
.sidebar-menu__item--sea .sidebar-menu__glyph { color: #2f7b74; }
.sidebar-menu__item--ocean .sidebar-menu__glyph { color: #2e6f93; }
.sidebar-menu__item--teal .sidebar-menu__glyph { color: #2c7b7b; }
.sidebar-menu__item--slate .sidebar-menu__glyph { color: #59667a; }
.sidebar-menu__item--sun .sidebar-menu__glyph { color: #c68727; }

.sidebar-menu .el-menu-item.is-active .sidebar-menu__glyph {
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.98) 0%, rgba(241, 247, 237, 0.98) 100%);
  box-shadow: inset 0 0 0 1px rgba(112, 142, 103, 0.12), 0 6px 16px rgba(104, 132, 96, 0.1);
}

.sidebar-menu .el-menu-item span {
  font-size: 13px;
  font-weight: 600;
  letter-spacing: 0.01em;
}

.ssh-dialog-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 10px;
  margin-bottom: 12px;
  padding: 10px 12px;
  border: 1px solid #e8eee4;
  border-radius: 10px;
  background: linear-gradient(135deg, #f6fbf4 0%, #ffffff 100%);
}

.ssh-connections-table :deep(.el-table__header-wrapper th) {
  background: #f5faf4;
  color: #44584a;
  font-weight: 600;
}

.ssh-connections-table :deep(.el-table__row td) {
  padding-top: 9px;
  padding-bottom: 9px;
}

.home-task-toolbar {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: center;
  padding: 14px 16px;
  margin-bottom: 18px;
  border: 1px solid #e4eadc;
  border-radius: 16px;
  background: linear-gradient(135deg, #f7faf4 0%, #fcfdf9 100%);
}

.home-task-toolbar__title {
  font-size: 15px;
  font-weight: 600;
  color: #425241;
}

.home-task-toolbar__desc {
  margin-top: 6px;
  font-size: 12px;
  color: #6c7d68;
  line-height: 1.6;
}

.home-task-toolbar__actions {
  display: flex;
  align-items: center;
  gap: 8px;
}

.home-task-form {
  margin-bottom: 0;
}

.home-task-tabs :deep(.el-tabs__header) {
  margin-bottom: 16px;
}

.home-task-tabs {
  /* Tabs 容器占满剩余空间，把滚动交给列表区域而不是整张卡片。
     Let tabs consume the remaining space and delegate scrolling to the list area only. */
  flex: 1;
  min-height: 0;
  display: flex;
  flex-direction: column;
}

.home-task-tabs :deep(.el-tabs__content) {
  flex: 1;
  min-height: 0;
}

.home-task-tabs :deep(.el-tab-pane) {
  height: 100%;
}

.home-task-tabs :deep(.el-tabs__item) {
  color: #6f7d6d;
}

.home-task-tabs :deep(.el-tabs__item.is-active) {
  color: #456c45;
}

.home-task-list {
  height: 100%;
  min-height: 0;
  overflow-y: auto;
  padding-right: 4px;
}

.home-task-empty {
  display: flex;
  align-items: center;
  justify-content: center;
  min-height: 240px;
  border: 1px dashed #d9e3d2;
  border-radius: 12px;
  color: #73806d;
  background: #fafcf8;
}

.home-task-card {
  padding: 16px 18px;
  border: 1px solid #e3e9dd;
  border-radius: 16px;
  background: linear-gradient(180deg, #ffffff 0%, #fbfcf8 100%);
  box-shadow: 0 10px 24px rgba(144, 160, 132, 0.08);
}

.home-task-card + .home-task-card {
  margin-top: 14px;
}

.home-task-card--archived {
  opacity: 0.94;
  background: linear-gradient(180deg, #fcfdfb 0%, #f7faf5 100%);
}

.home-task-card__header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.home-task-card__title {
  font-size: 16px;
  font-weight: 600;
  color: #39463a;
}

.home-task-card__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 6px;
  font-size: 12px;
  color: #70806b;
}

.home-task-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  justify-content: flex-end;
  margin-top: 16px;
}

.home-task-card__remark {
  margin-top: 14px;
  padding: 12px 14px;
  border: 1px solid #e7ede1;
  border-radius: 12px;
  background: linear-gradient(180deg, #f8faf4 0%, #f4f7ee 100%);
  color: #536251;
  font-size: 13px;
  line-height: 1.6;
  white-space: pre-wrap;
  word-break: break-word;
}

.home-task-dialog__footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

@media (max-width: 900px) {
  .home-dashboard-stage {
    height: calc(100vh - 40px);
    border-radius: 18px;
  }

  .home-dashboard-pager {
    right: 10px;
  }

  .ssh-dialog-toolbar {
    flex-direction: column;
    align-items: flex-start;
  }

  .home-task-toolbar {
    flex-direction: column;
  }

  .home-task-toolbar__actions {
    width: 100%;
    justify-content: flex-start;
    flex-wrap: wrap;
  }

  .home-task-card__header {
    flex-direction: column;
  }

  .home-task-panel {
    padding: 14px;
  }
}
</style>

