<template>
  <div class="layout-container" :class="{ 'layout-container--hide-sidebar': hideAppSidebar }">
    <!-- 左侧菜单 -->
    <aside v-if="!hideAppSidebar" class="sidebar">
      <div class="sidebar-header">
        <span class="logo">
          <img src="../../assets/dtool-logo-32.png" alt="DTool logo" class="logo-image">
        </span>
        <span class="title">DTool</span>
      </div>
      
      <el-menu
        :default-active="menuName"
        active-text-color="#3a7a3a"
        background-color="#f5f5f0"
        text-color="#5a5a5a"
        class="sidebar-menu"
        @click="onMenuNativeClick"
        @select="handleSelect"
      >
        <el-menu-item index="/Dashboard">
          <el-icon><HomeFilled /></el-icon>
          <span>Command</span>
        </el-menu-item>
        <el-menu-item index="/HomeTask" class="menu-item-workflow" :class="{ 'menu-item-workflow--alert': workflowUnreadVisible }">
          <el-icon><List /></el-icon>
          <span>Workflow</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('redis')" index="/Redis">
          <el-icon><Coin /></el-icon>
          <span>Redis</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('supervisor')" index="/Supervisor">
          <el-icon><Setting /></el-icon>
          <span>Supervisor</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('git')" index="/Git">
          <el-icon><Folder /></el-icon>
          <span>Git</span>
        </el-menu-item>
        <!-- <el-menu-item v-if="checkModuleOpen('tools')" index="/CommonActions" class="menu-item-common-actions"> -->
          <!-- <el-icon><ToolsIcon /></el-icon> -->
          <!-- <span>常用操作</span> -->
        <!-- </el-menu-item> -->
        <el-menu-item v-if="checkModuleOpen('login')" index="/Link">
          <el-icon><Link /></el-icon>
          <span>Playwright</span>
        </el-menu-item>
<!--        <el-menu-item v-if="checkModuleOpen('variable')" index="/Variable">-->
<!--          <el-icon><Document /></el-icon>-->
<!--          <span>Script</span>-->
<!--        </el-menu-item>-->
        <el-menu-item v-if="checkModuleOpen('memory_fragment')" index="/MemoryFragment">
          <el-icon><Memo /></el-icon>
          <span>Knowledge</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('docker')" index="/Docker">
          <el-icon><Box /></el-icon>
          <span>Docker</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('api')" index="/Api">
          <el-icon><Connection /></el-icon>
          <span>Api Manage</span>
        </el-menu-item>
        <el-menu-item index="/AgentCli" class="menu-item-agent-cli" :class="{ 'menu-item-agent-cli--alert': agentCliUnreadVisible }">
          <el-icon><Monitor /></el-icon>
          <span>Agent Cli</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('shellout')" index="/shellout">
          <el-icon><Monitor /></el-icon>
          <span>Log Witch</span>
        </el-menu-item>
        <el-menu-item index="/Set" class="menu-item-settings" :class="{ 'menu-item-settings--alert': mainDbStorageAlertVisible }">
          <el-icon><Setting /></el-icon>
          <span>Setting</span>
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
          </button>
          <button type="button" class="footer-action footer-action--mint" @click="drawerVisibleTools = true">
            <span class="footer-action__title">小工具</span>
          </button>
          <button type="button" class="footer-action footer-action--sky" @click="openSshConnectionsDialog">
            <span class="footer-action__title">当前 SSH 连接数 {{ sshConnectionCount }}</span>
          </button>
        </div>
        <button
          type="button"
          class="footer-action footer-action--sand async-task-entry"
          :class="[getAsyncTaskEntryClassName(), { 'async-task-entry--running': hasRunningAsyncTask() }]"
          @click="openAsyncTaskDialog"
        >
          <span class="footer-action__title async-task-entry__title">
            <span class="async-task-entry__label">任务</span>
            <span class="async-task-entry__summary">
              <span
                class="async-task-entry__digit async-task-entry__digit--running"
                :title="getAsyncTaskCounterDescription('running')"
              >{{ asyncTaskSummary.running_count || 0 }}</span>
              <span class="async-task-entry__slash">/</span>
              <span
                class="async-task-entry__digit async-task-entry__digit--pending"
                :title="getAsyncTaskCounterDescription('pending')"
              >{{ asyncTaskSummary.pending_count || 0 }}</span>
              <span class="async-task-entry__slash">/</span>
              <span
                class="async-task-entry__digit async-task-entry__digit--await-confirm"
                :title="getAsyncTaskCounterDescription('await_confirm')"
              >{{ asyncTaskSummary.await_confirm_count || 0 }}</span>
              <span class="async-task-entry__slash">/</span>
              <span
                class="async-task-entry__digit async-task-entry__digit--failed"
                :title="getAsyncTaskCounterDescription('failed')"
              >{{ asyncTaskSummary.failed_count || 0 }}</span>
            </span>
            <span v-if="hasRunningAsyncTask()" class="async-task-entry__spinner" aria-hidden="true"></span>
          </span>
        </button>

      </div>
    </aside>

    <!-- 主内容区域 -->
    <main class="main-content">
      <div class="main-content__body">
        <div class="main-content__view">
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

  <!-- Safe 登录弹窗 -->
  <el-dialog
    v-model="safeLoginVisible"
    title="后台登录"
    width="420"
    :close-on-click-modal="false"
    :close-on-press-escape="false"
    :show-close="false"
    class="safe-login-dialog"
    :modal-class="'safe-login-modal'"
  >
    <div class="safe-login-content">
      <p class="safe-login-desc">{{ safeLoginMessage || '请输入后台访问密码' }}</p>
      <el-input
        v-model="safeLoginPassword"
        type="password"
        placeholder="请输入密码"
        show-password
        @keyup.enter="handleSafeLogin"
      />
      <p v-if="safeLoginError" class="safe-login-error">{{ safeLoginError }}</p>
    </div>
    <template #footer>
      <div class="dialog-footer">
        <pl-button type="primary" :loading="safeLoginLoading" @click="handleSafeLogin">登录</pl-button>
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
    v-model="asyncTaskDialogVisible"
    title="异步任务"
    width="78%"
    top="6vh"
    class="async-task-dialog"
  >
    <div class="async-task-toolbar">
      <el-tag type="success" effect="light" :title="getAsyncTaskCounterDescription('running')">运行中 {{ asyncTaskSummary.running_count || 0 }}</el-tag>
      <el-tag type="info" effect="light" :title="getAsyncTaskCounterDescription('pending')">准备中 {{ asyncTaskSummary.pending_count || 0 }}</el-tag>
      <el-tag type="warning" effect="light" :title="getAsyncTaskCounterDescription('await_confirm')">待处理 {{ asyncTaskSummary.await_confirm_count || 0 }}</el-tag>
      <el-tag type="danger" effect="light" :title="getAsyncTaskCounterDescription('failed')">失败 {{ asyncTaskSummary.failed_count || 0 }}</el-tag>
      <el-tag effect="plain">列表 {{ asyncTaskList.length }}</el-tag>
    </div>
    <div class="async-task-layout">
      <div v-loading="asyncTaskLoading" class="async-task-list">
        <button
          v-for="task in asyncTaskList"
          :key="task.id"
          type="button"
          class="async-task-item"
          :class="{ 'async-task-item--active': Number(asyncTaskSelectedId) === Number(task.id) }"
          @click="selectAsyncTask(task)"
        >
          <div class="async-task-item__header">
            <div class="async-task-item__title">{{ task.title || getAsyncTaskTypeText(task) }}</div>
            <el-tag size="small" effect="light" :type="getAsyncTaskStatusTagType(task.task_status)">
              {{ getAsyncTaskStatusText(task.task_status) }}
            </el-tag>
          </div>
          <div class="async-task-item__meta">{{ getAsyncTaskTypeText(task) }}</div>
          <div class="async-task-item__time">创建时间 {{ formatAsyncTaskTime(task.create_time) }}</div>
        </button>
        <div v-if="!asyncTaskLoading && asyncTaskList.length === 0" class="async-task-empty">
          当前没有异步任务
        </div>
      </div>
      <div class="async-task-detail">
        <div v-if="!asyncTaskDetail.id" class="async-task-empty">
          请选择左侧任务查看详情
        </div>
        <template v-else>
          <div class="async-task-detail__header">
            <div>
              <div class="async-task-detail__title">{{ asyncTaskDetail.title || getAsyncTaskTypeText(asyncTaskDetail) }}</div>
              <div class="async-task-detail__meta">
                <span>{{ getAsyncTaskTypeText(asyncTaskDetail) }}</span>
                <span>状态 {{ getAsyncTaskStatusText(asyncTaskDetail.task_status) }}</span>
                <span>完成时间 {{ formatAsyncTaskTime(asyncTaskDetail.finish_time) }}</span>
              </div>
            </div>
            <GitActionButton compact variant="danger" :loading="asyncTaskDeleting" @click="deleteAsyncTask(asyncTaskDetail)">
              删除
            </GitActionButton>
          </div>
          <div v-if="asyncTaskDetail.error_message" class="async-task-detail__error">
            {{ asyncTaskDetail.error_message }}
          </div>
          <div v-if="asyncTaskDetail.run_logs" class="async-task-detail__logs">
            <div class="async-task-detail__logs-title">运行日志</div>
            <pre class="async-task-detail__logs-pre">{{ asyncTaskDetail.run_logs }}</pre>
          </div>
          <div v-if="asyncTaskDetail.task_type === ASYNC_TASK_TYPE_DAILY_REPORT" class="async-task-detail__content">
            <div class="async-task-detail__section-title">日报预览</div>
            <pre class="async-task-detail__pre">{{ asyncTaskDetail.result_payload_map?.markdown || '' }}</pre>
          </div>
          <div v-else-if="asyncTaskDetail.task_type === ASYNC_TASK_TYPE_MEMORY_ARRANGE" class="async-task-detail__content">
            <div class="async-task-detail__section-title">正文差异</div>
            <diff-markdown
              :old-text="asyncTaskDetail.result_payload_map?.original_content || ''"
              :new-text="asyncTaskDetail.result_payload_map?.arranged_content || ''"
              title="正文差异"
            />
          </div>
          <div v-else-if="asyncTaskDetail.task_type === ASYNC_TASK_TYPE_TAPD_SCRAPE" class="async-task-detail__content">
            <div class="async-task-detail__section-title">
              抓取内容预览
              <div class="async-task-detail__view-toggle">
                <GitActionButton
                  variant="info"
                  compact
                  :class="{ 'mode-button-active': tapdScrapeViewMode === 'preview' }"
                  @click="tapdScrapeViewMode = 'preview'"
                >
                  查看
                </GitActionButton>
                <GitActionButton
                  compact
                  :class="{ 'mode-button-active': tapdScrapeViewMode === 'source' }"
                  @click="tapdScrapeViewMode = 'source'"
                >
                  源码
                </GitActionButton>
                 <GitActionButton

                              compact
                              :loading="asyncTaskRetrying"
                              @click="retryAsyncTask"
                            >
                              重试
                            </GitActionButton>
              </div>
            </div>
            <MarkdownRenderer
              v-if="tapdScrapeViewMode === 'preview'"
              :source="asyncTaskDetail.result_payload_map?.markdown || ''"
              class="async-task-detail__markdown"
            />
            <pre v-else class="async-task-detail__pre">{{ asyncTaskDetail.result_payload_map?.markdown || '' }}</pre>
            <div v-if="asyncTaskDetail.result_payload_map?.image_count > 0" class="async-task-detail__sub-meta">
              包含 {{ asyncTaskDetail.result_payload_map?.image_count }} 张图片
            </div>
          </div>
          <div
            v-else-if="asyncTaskDetail.task_type === ASYNC_TASK_TYPE_MAIN_DB_SYNC || asyncTaskDetail.task_type === ASYNC_TASK_TYPE_MEMORY_DB_SYNC"
            class="async-task-detail__content"
          >
            <div class="async-task-detail__section-title">任务说明</div>
            <div class="async-task-detail__note">{{ getAsyncTaskDescription(asyncTaskDetail) }}</div>
            <div v-if="getAsyncTaskScheduledTime(asyncTaskDetail)" class="async-task-detail__sub-meta">
              预计同步时间 {{ getAsyncTaskScheduledTime(asyncTaskDetail) }}
            </div>
          </div>
          <div class="async-task-detail__actions">
            <GitActionButton

              compact
              :loading="asyncTaskRetrying"
              @click="retryAsyncTask"
            >
              重试
            </GitActionButton>
            <GitActionButton
              v-if="asyncTaskDetail.task_status === ASYNC_TASK_STATUS_AWAIT_CONFIRM && asyncTaskDetail.task_type === ASYNC_TASK_TYPE_DAILY_REPORT"
              compact
              :loading="asyncTaskActing"
              @click="runAsyncTaskAction(ASYNC_TASK_ACTION_SAVE_DAILY_REPORT)"
            >
              保存为知识片段
            </GitActionButton>
            <GitActionButton
              v-if="asyncTaskDetail.task_status === ASYNC_TASK_STATUS_AWAIT_CONFIRM && asyncTaskDetail.task_type === ASYNC_TASK_TYPE_MEMORY_ARRANGE"
              compact
              :loading="asyncTaskActing"
              @click="runAsyncTaskAction(ASYNC_TASK_ACTION_OVERWRITE_MEMORY_FRAGMENT)"
            >
              覆盖原文
            </GitActionButton>
            <GitActionButton
              v-if="asyncTaskDetail.task_status === ASYNC_TASK_STATUS_AWAIT_CONFIRM && asyncTaskDetail.task_type === ASYNC_TASK_TYPE_TAPD_SCRAPE"
              compact
              :loading="asyncTaskActing"
              @click="runAsyncTaskAction(ASYNC_TASK_ACTION_OVERWRITE_FRAGMENT_WITH_SCRAPE)"
            >
              更新到知识片段
            </GitActionButton>
            <GitActionButton
              v-if="asyncTaskDetail.task_status === ASYNC_TASK_STATUS_AWAIT_CONFIRM"
              compact
              variant="info"
              :loading="asyncTaskActing"
              @click="runAsyncTaskAction(ASYNC_TASK_ACTION_DISCARD)"
            >
              丢弃
            </GitActionButton>
          </div>
        </template>
      </div>
    </div>
  </el-dialog>

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
import asyncTaskApi from '@/utils/base/async_task'
import agentCliApi from '@/utils/base/agent_cli'
import homeTaskApi from '@/utils/base/home_task'
import taskWorkflowApi from '@/utils/base/task_workflow'
import sseDistribute from '@/utils/base/sse_distribute'
import Tools from "@/components/Tools.vue";
import Markdown from '@/components/Markdown.vue'
import GitActionButton from "@/components/base/GitActionButton.vue";
import DiffMarkdown from '@/components/base/diff_markwodn.vue'
import MarkdownRenderer from '@/components/base/markdown.vue'
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
  List,
  Tools as ToolsIcon
} from "@element-plus/icons-vue";

// SSH_CONNECTION_REFRESH_INTERVAL_MS 统一控制 SSH 连接轮询周期。
const SSH_CONNECTION_REFRESH_INTERVAL_MS = 5000
// ASYNC_TASK_ACTION_* 统一定义异步任务动作常量。
const ASYNC_TASK_ACTION_SAVE_DAILY_REPORT = 'save_daily_report'
const ASYNC_TASK_ACTION_OVERWRITE_MEMORY_FRAGMENT = 'overwrite_memory_fragment'
const ASYNC_TASK_ACTION_OVERWRITE_FRAGMENT_WITH_SCRAPE = 'overwrite_fragment_with_scrape'
const ASYNC_TASK_ACTION_DISCARD = 'discard'
// ASYNC_TASK_STATUS_* 统一定义异步任务状态常量。
const ASYNC_TASK_STATUS_AWAIT_CONFIRM = 'await_confirm'
const ASYNC_TASK_STATUS_PENDING = 'pending'
const ASYNC_TASK_STATUS_RUNNING = 'running'
const ASYNC_TASK_STATUS_FAILED = 'failed'
const ASYNC_TASK_STATUS_CONFIRMED = 'confirmed'
const ASYNC_TASK_STATUS_REJECTED = 'rejected'
// ASYNC_TASK_TYPE_* 统一定义异步任务类型常量。
const ASYNC_TASK_TYPE_DAILY_REPORT = 'home_task_daily_report'
const ASYNC_TASK_TYPE_MEMORY_ARRANGE = 'memory_fragment_arrange'
const ASYNC_TASK_TYPE_MAIN_DB_SYNC = 'main_db_sync'
const ASYNC_TASK_TYPE_MEMORY_DB_SYNC = 'memory_db_sync'
const ASYNC_TASK_TYPE_TAPD_SCRAPE = 'home_task_tapd_scrape'

export default {
  data() {
    return {
      drawerVisibleTools: false,
      drawerVisibleMarkdown: false,
      // Safe 登录相关
      safeLoginVisible: false,
      safeLoginPassword: '',
      safeLoginLoading: false,
      safeLoginError: '',
      safeLoginChecked: false,
      safeLoginMessage: '',
      menuName: '/Dashboard',
      menuCtrlKey: false,
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
      workflowUnreadVisible: false,
      workflowUnreadTotal: 0,
      workflowUnreadMap: {},
      mainDbStorageAlertVisible: false,
      agentCliUnreadVisible: false,
      agentCliUnreadTotal: 0,
      sshConnections: [],
      sshConnectionsDialogVisible: false,
      sshConnectionsLoading: false,
      sshConnectionTimer: null,
      ASYNC_TASK_ACTION_SAVE_DAILY_REPORT,
      ASYNC_TASK_ACTION_OVERWRITE_MEMORY_FRAGMENT,
      ASYNC_TASK_ACTION_OVERWRITE_FRAGMENT_WITH_SCRAPE,
      ASYNC_TASK_ACTION_DISCARD,
      ASYNC_TASK_STATUS_AWAIT_CONFIRM,
      ASYNC_TASK_STATUS_PENDING,
      ASYNC_TASK_STATUS_FAILED,
      ASYNC_TASK_TYPE_DAILY_REPORT,
      ASYNC_TASK_TYPE_MEMORY_ARRANGE,
      ASYNC_TASK_TYPE_MAIN_DB_SYNC,
      ASYNC_TASK_TYPE_MEMORY_DB_SYNC,
      ASYNC_TASK_TYPE_TAPD_SCRAPE,
      asyncTaskDialogVisible: false,
      asyncTaskLoading: false,
      asyncTaskActing: false,
      asyncTaskDeleting: false,
      asyncTaskRetrying: false,
      asyncTaskSelectedId: 0,
      asyncTaskList: [],
      asyncTaskDetail: {},
      // tapdScrapeViewMode 控制 TAPD 抓取预览的显示模式：preview 为渲染查看，source 为源码。
      tapdScrapeViewMode: 'preview',
      asyncTaskNotifiedStateMap: {},
      asyncTaskNotificationPermissionRequested: false,
      asyncTaskSummary: {
        pending_count: 0,
        await_confirm_count: 0,
        running_count: 0,
        failed_count: 0,
        total: 0,
      },
    }
  },
  computed: {
    // hideAppSidebar 控制某些独立页卡场景下隐藏应用左侧主菜单。
    hideAppSidebar() {
      return String(this.$route.query.hide_menu || '') === '1'
    },
  },
  watch: {
    '$route.path'(newPath) {
      this.menuName = newPath || '/Dashboard'
    },
  },
  created() {
    window.handleCopy = copy.handleCopy;
  },
  mounted: function () {
    let _that = this
    _that.openModuleList = module.GetOpenModuleList()
    // Safe 登录检查
    _that.checkSafeLoginStatus()
    this.forceIp(false)
    // 注册Shell连接状态SSE监听
    sseDistribute.RegisterReceive('shell_connections', function(data, type, distributeId) {
      _that.handleSshConnectionsUpdate(data)
    })
    // 注册异步任务状态SSE监听
    sseDistribute.RegisterReceive('async_tasks', function(data) {
      _that.handleAsyncTasksUpdate(data)
    })
    // 注册安全认证失效SSE监听
    sseDistribute.RegisterReceive('safe_auth_required', function(data) {
      _that.handleSafeAuthRequired(data)
    })
    sseDistribute.RegisterReceive('git_pending_status', function(data) {
      _that.handleGitPendingStatusUpdate(data)
    })
    sseDistribute.RegisterReceive('agent_cli_unread_home', function(data) {
      _that.handleAgentCliUnreadUpdate(data)
    })
    sseDistribute.RegisterReceive('workflow_unread_home_menu', function(data) {
      _that.handleWorkflowUnreadUpdate(data)
    })
    this.ensureAsyncTaskNotificationPermission()
    this.menuName = this.$route.path || '/Dashboard'
    window.addEventListener('resize', function () {});
    this.loadWorkflowUnreadSummary()
    this.loadAgentCliUnreadSummary()
    // 监听全局登录失效事件
    if (this.$eventBus) {
      this.$eventBus.on('safe_auth_required', this.showSafeLogin)
      this.$eventBus.on('main_db_storage_alert_changed', this.handleMainDbStorageAlertEvent)
    }
  },
  provide() {
    return {
      showTerminal: this.showTerminal,
      resizeTerminal: this.resizeTerminal,
    };
  },
  methods: {
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
    // Safe 登录检查
    checkSafeLoginStatus: function () {
      let _that = this
      base.BaseLoginStatus(function (response) {
        _that.safeLoginChecked = true
        if (response.ErrCode !== 0) {
          // 接口调用失败，但不阻止进入
          return
        }
        const data = response.Data || {}
        _that.initSseAfterLoginStatus(data.sse_ports)
        // enabled=false：未启用密码保护，直接进入
        if (!data.enabled) {
          return
        }
        // enabled=true && logged_in=true：已登录，直接进入
        if (data.logged_in) {
          return
        }
        // enabled=true && logged_in=false：未登录，显示登录框
        _that.showSafeLogin()
      })
    },
    initSseAfterLoginStatus: function (ssePorts) {
      // sse_distribute 内部会查询可用端口，无可用时自动弹窗
      sseDistribute.ConnectToAvailablePort()
    },
    // 显示 Safe 登录弹窗
    showSafeLogin: function (options) {
      // 防止重复弹窗：如果弹窗已经显示，不再重复弹出
      if (this.safeLoginVisible) {
        // 如果传入新的消息，更新提示文字
        if (options && options.message) {
          this.safeLoginMessage = options.message
        }
        return
      }
      this.safeLoginVisible = true
      this.safeLoginPassword = ''
      this.safeLoginError = ''
      // 支持传入自定义提示消息
      if (options && options.message) {
        this.safeLoginMessage = options.message
      } else {
        this.safeLoginMessage = ''
      }
    },
    // 处理 SSE 推送的安全认证失效事件
    handleSafeAuthRequired: function (data) {
      // 清除本地 token
      base.ClearSafeToken()
      // 显示登录弹窗
      const message = data && data.message ? data.message : '登录态已失效，请重新登录'
      this.showSafeLogin({ message: message })
    },
    // 处理 Safe 登录
    handleSafeLogin: function () {
      let _that = this
      const password = _that.safeLoginPassword.trim()
      if (!password) {
        _that.safeLoginError = '请输入密码'
        return
      }
      _that.safeLoginLoading = true
      _that.safeLoginError = ''
      base.BaseLogin(password, function (response) {
        _that.safeLoginLoading = false
        if (response.ErrCode === 0) {
          _that.safeLoginVisible = false
          _that.safeLoginPassword = ''
          _that.$helperNotify.success('登录成功')
          // 登录成功后刷新页面，确保所有状态正确
          window.location.reload()
        } else {
          const errMsg = response.ErrMsg || '登录失败'
          if (response.Data && response.Data.enabled === false) {
            // 未启用密码保护
            _that.safeLoginVisible = false
            return
          }
          _that.safeLoginError = errMsg
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
    // openAsyncTaskDialog 打开异步任务弹窗，并刷新摘要与详情。
    openAsyncTaskDialog() {
      this.asyncTaskSelectedId = 0
      this.asyncTaskDetail = {}
      this.asyncTaskDialogVisible = true
      this.loadAsyncTaskSummary(true)
    },

    // loadAsyncTaskSummary 刷新异步任务摘要与最近任务列表。
    loadAsyncTaskSummary(showLoading) {
      if (showLoading) {
        this.asyncTaskLoading = true
      }
      asyncTaskApi.AsyncTaskList(20, (response) => {
        this.asyncTaskLoading = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          return
        }
        const list = Array.isArray(response.Data.list) ? response.Data.list : []
        this.processAsyncTaskNotifications(list)
        this.asyncTaskList = list
        this.asyncTaskSummary = {
          pending_count: Number(response.Data.pending_count || 0),
          await_confirm_count: Number(response.Data.await_confirm_count || 0),
          running_count: Number(response.Data.running_count || 0),
          failed_count: Number(response.Data.failed_count || 0),
          total: Number(response.Data.total || list.length),
        }
        if (!this.asyncTaskSelectedId && list.length > 0) {
          this.selectAsyncTask(list[0])
          return
        }
        if (this.asyncTaskSelectedId) {
          const activeTask = list.find(item => Number(item.id) === Number(this.asyncTaskSelectedId))
          if (activeTask) {
            this.selectAsyncTask(activeTask)
          } else {
            this.asyncTaskSelectedId = 0
            this.asyncTaskDetail = {}
          }
        }
      })
    },
    // ensureAsyncTaskNotificationPermission 尝试申请浏览器通知权限，只在首次进入时请求一次。 // ensureAsyncTaskNotificationPermission requests browser notification permission once for async task completion alerts.
    ensureAsyncTaskNotificationPermission() {
      if (this.asyncTaskNotificationPermissionRequested) {
        return
      }
      this.asyncTaskNotificationPermissionRequested = true
      if (typeof window === 'undefined' || typeof Notification === 'undefined') {
        return
      }
      if (Notification.permission !== 'default') {
        return
      }
      Notification.requestPermission().catch(() => {})
    },
    // processAsyncTaskNotifications 检测新完成任务并触发浏览器通知。 // processAsyncTaskNotifications detects newly finished tasks and raises browser notifications.
    processAsyncTaskNotifications(taskList) {
      const nextStateMap = {}
      taskList.forEach((task) => {
        const taskId = Number(task?.id || 0)
        if (taskId <= 0) {
          return
        }
        const status = String(task.task_status || '')
        const previousStatus = String(this.asyncTaskNotifiedStateMap[taskId] || '')
        nextStateMap[taskId] = status
        const shouldNotifySuccess = status === ASYNC_TASK_STATUS_AWAIT_CONFIRM && previousStatus && previousStatus !== status
        const shouldNotifyFailed = status === ASYNC_TASK_STATUS_FAILED && previousStatus && previousStatus !== status
        if (shouldNotifySuccess || shouldNotifyFailed) {
          this.notifyAsyncTaskCompletion(task)
        }
      })
      this.asyncTaskNotifiedStateMap = nextStateMap
    },
    // notifyAsyncTaskCompletion 使用浏览器原生通知提醒用户异步任务已完成或失败。 // notifyAsyncTaskCompletion uses the browser Notification API to alert the user when an async task finishes or fails.
    notifyAsyncTaskCompletion(task) {
      if (typeof window === 'undefined' || typeof Notification === 'undefined') {
        return
      }
      if (Notification.permission !== 'granted') {
        return
      }
      const taskType = String(task?.task_type || '')
      const taskStatus = String(task?.task_status || '')
      let title = '异步任务有新结果'
      let body = '请打开异步任务查看详情'
      if (taskStatus === ASYNC_TASK_STATUS_FAILED) {
        title = '异步任务执行失败'
        body = '请打开异步任务查看失败原因'
      } else if (taskType === ASYNC_TASK_TYPE_DAILY_REPORT) {
        title = '工作日报已生成'
        body = '等待你确认是否保存为知识片段'
      } else if (taskType === ASYNC_TASK_TYPE_MEMORY_ARRANGE) {
        title = '知识片段整理完成'
        body = '等待你确认是否覆盖原文'
      }
      const notification = new Notification(title, {
        body: body,
        tag: `async-task-${task.id}`,
      })
      notification.onclick = () => {
        window.focus()
        this.openAsyncTaskDialog()
        this.selectAsyncTask(task)
        notification.close()
      }
    },
    // selectAsyncTask 选中指定异步任务并加载详情。
    selectAsyncTask(task) {
      const taskId = Number(task?.id || 0)
      if (taskId <= 0) {
        return
      }
      // 切换任务时重置预览模式为渲染查看。
      this.tapdScrapeViewMode = 'preview'
      this.asyncTaskSelectedId = taskId
      asyncTaskApi.AsyncTaskInfo(taskId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          return
        }
        this.asyncTaskDetail = this.normalizeAsyncTaskDetail(response.Data)
      })
    },
    // normalizeAsyncTaskDetail 解析详情里的 JSON 字段，方便模板直接读取结果内容。
    normalizeAsyncTaskDetail(task) {
      const normalizedTask = { ...(task || {}), result_payload_map: {}, request_payload_map: {} }
      try {
        normalizedTask.result_payload_map = JSON.parse(String(normalizedTask.result_payload || '{}'))
      } catch (error) {
        normalizedTask.result_payload_map = {}
      }
      try {
        normalizedTask.request_payload_map = JSON.parse(String(normalizedTask.request_payload || '{}'))
      } catch (error) {
        normalizedTask.request_payload_map = {}
      }
      return normalizedTask
    },
    // runAsyncTaskAction 处理异步任务确认或丢弃动作，并刷新当前选中详情。
    runAsyncTaskAction(action) {
      if (!this.asyncTaskDetail.id) {
        return
      }
      this.asyncTaskActing = true
      asyncTaskApi.AsyncTaskAction(this.asyncTaskDetail.id, action, (response) => {
        this.asyncTaskActing = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.$helperNotify.error(response?.ErrMsg || '异步任务处理失败')
          return
        }
        this.asyncTaskDetail = this.normalizeAsyncTaskDetail(response.Data)
        this.loadAsyncTaskSummary(false)
        this.$helperNotify.success(action === ASYNC_TASK_ACTION_DISCARD ? '异步任务结果已丢弃' : '异步任务结果已处理')
      })
    },
    // deleteAsyncTask 删除异步任务记录，并同步刷新列表与详情。
    deleteAsyncTask(task) {
      const taskId = Number(task?.id || 0)
      if (taskId <= 0) {
        return
      }
      this.asyncTaskDeleting = true
      asyncTaskApi.AsyncTaskDelete(taskId, (response) => {
        this.asyncTaskDeleting = false
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '异步任务删除失败')
          return
        }
        if (Number(this.asyncTaskSelectedId) === taskId) {
          this.asyncTaskSelectedId = 0
          this.asyncTaskDetail = {}
        }
        this.loadAsyncTaskSummary(false)
        this.$helperNotify.success('异步任务记录已删除')
      })
    },
    // retryAsyncTask 重试失败的异步任务。
    retryAsyncTask() {
      const taskId = Number(this.asyncTaskDetail?.id || 0)
      if (taskId <= 0) {
        return
      }
      this.asyncTaskRetrying = true
      asyncTaskApi.AsyncTaskRetry(taskId, (response) => {
        this.asyncTaskRetrying = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.$helperNotify.error(response?.ErrMsg || '异步任务重试失败')
          return
        }
        this.asyncTaskDetail = this.normalizeAsyncTaskDetail(response.Data)
        this.loadAsyncTaskSummary(false)
        this.$helperNotify.success('异步任务已重新执行')
      })
    },
    // getAsyncTaskTypeText 统一格式化异步任务类型文案。
    getAsyncTaskTypeText(task) {
      const taskType = String(task?.task_type || '')
      if (taskType === ASYNC_TASK_TYPE_DAILY_REPORT) {
        return '工作日报'
      }
      if (taskType === ASYNC_TASK_TYPE_MEMORY_ARRANGE) {
        return '知识片段整理'
      }
      if (taskType === ASYNC_TASK_TYPE_MAIN_DB_SYNC) {
        return '主库同步'
      }
      if (taskType === ASYNC_TASK_TYPE_MEMORY_DB_SYNC) {
        return '记忆库同步'
      }
      if (taskType === ASYNC_TASK_TYPE_TAPD_SCRAPE) {
        return 'TAPD网页抓取'
      }
      return '异步任务'
    },
    // getAsyncTaskStatusText 统一格式化异步任务状态文案。
    getAsyncTaskStatusText(taskStatus) {
      const normalizedStatus = String(taskStatus || '')
      if (normalizedStatus === ASYNC_TASK_STATUS_AWAIT_CONFIRM) {
        return '待处理'
      }
      if (normalizedStatus === ASYNC_TASK_STATUS_PENDING) {
        return '准备中'
      }
      if (normalizedStatus === ASYNC_TASK_STATUS_RUNNING) {
        return '运行中'
      }
      if (normalizedStatus === ASYNC_TASK_STATUS_FAILED) {
        return '失败'
      }
      if (normalizedStatus === ASYNC_TASK_STATUS_CONFIRMED) {
        return '已确认'
      }
      if (normalizedStatus === ASYNC_TASK_STATUS_REJECTED) {
        return '已丢弃'
      }
      return normalizedStatus || '-'
    },
    // getAsyncTaskStatusTagType 根据状态返回标签样式。
    getAsyncTaskStatusTagType(taskStatus) {
      const normalizedStatus = String(taskStatus || '')
      if (normalizedStatus === ASYNC_TASK_STATUS_AWAIT_CONFIRM) {
        return 'warning'
      }
      if (normalizedStatus === ASYNC_TASK_STATUS_PENDING) {
        return 'info'
      }
      if (normalizedStatus === ASYNC_TASK_STATUS_RUNNING) {
        return 'success'
      }
      if (normalizedStatus === ASYNC_TASK_STATUS_FAILED) {
        return 'danger'
      }
      if (normalizedStatus === ASYNC_TASK_STATUS_CONFIRMED) {
        return 'success'
      }
      return ''
    },
    // hasRunningAsyncTask 判断当前是否存在执行中的异步任务，用于左下角入口动画提示。 // hasRunningAsyncTask checks whether any async task is currently running so the footer entry can animate.
    hasRunningAsyncTask() {
      return Number(this.asyncTaskSummary.running_count || 0) > 0
    },
    // getAsyncTaskEntryState 根据汇总数量决定左下角入口背景优先级。 // getAsyncTaskEntryState decides the footer entry state by summary priority.
    getAsyncTaskEntryState() {
      const failedCount = Number(this.asyncTaskSummary.failed_count || 0)
      if (failedCount > 0) {
        return 'failed'
      }
      const awaitConfirmCount = Number(this.asyncTaskSummary.await_confirm_count || 0)
      if (awaitConfirmCount > 0) {
        return 'await-confirm'
      }
      const pendingCount = Number(this.asyncTaskSummary.pending_count || 0)
      if (pendingCount > 0) {
        return 'pending'
      }
      const runningCount = Number(this.asyncTaskSummary.running_count || 0)
      // 失败和待处理都没有时，执行中优先展示柔和绿色。 // Show the soft green running state only when failed and await-confirm are both absent.
      if (runningCount > 0) {
        return 'running'
      }
      return 'idle'
    },
    // getAsyncTaskEntryClassName 返回左下角入口对应的背景样式类。 // getAsyncTaskEntryClassName returns the footer entry background class name.
    getAsyncTaskEntryClassName() {
      const state = this.getAsyncTaskEntryState()
      if (state === 'failed') {
        return 'async-task-entry--failed'
      }
      if (state === 'await-confirm') {
        return 'async-task-entry--await-confirm'
      }
      if (state === 'pending') {
        return 'async-task-entry--pending'
      }
      if (state === 'running') {
        return 'async-task-entry--active'
      }
      return 'async-task-entry--idle'
    },
    // getAsyncTaskCounterDescription 返回任务汇总指标的悬停说明，附带当前数量。 // Return hover descriptions with current count for async task summary counters.
    getAsyncTaskCounterDescription(type) {
      if (type === 'running') {
        return '运行中: ' + (this.asyncTaskSummary.running_count || 0)
      }
      if (type === 'pending') {
        return '准备中: ' + (this.asyncTaskSummary.pending_count || 0)
      }
      if (type === 'await_confirm') {
        return '待处理: ' + (this.asyncTaskSummary.await_confirm_count || 0)
      }
      if (type === 'failed') {
        return '失败: ' + (this.asyncTaskSummary.failed_count || 0)
      }
      return '异步任务汇总'
    },
    getAsyncTaskDescription(task) {
      let desc = String(task?.request_payload_map?.task_description || '').trim()
      // 中文注释：主库/记忆库同步任务在完成态都补充“已于 xx 完成”，统一详情文案。
      // English comment: Append the finished-at suffix for both main-db and memory-db sync tasks once confirmed.
      if (
        (String(task?.task_type || '') === ASYNC_TASK_TYPE_MAIN_DB_SYNC ||
          String(task?.task_type || '') === ASYNC_TASK_TYPE_MEMORY_DB_SYNC) &&
        String(task?.task_status || '') === ASYNC_TASK_STATUS_CONFIRMED
      ) {
        const finishTime = this.formatAsyncTaskTime(task?.finish_time)
        if (finishTime && finishTime !== '-') {
          desc += '，已于 ' + finishTime + ' 同步完成'
        }
      }
      return desc
    },
    getAsyncTaskScheduledTime(task) {
      return String(task?.request_payload_map?.schedule?.scheduled_at_desc || '').trim()
    },
    // formatAsyncTaskTime 统一格式化异步任务时间戳。
    formatAsyncTaskTime(unixTime) {
      const timeValue = Number(unixTime || 0)
      if (timeValue <= 0) {
        return '-'
      }
      const date = new Date(timeValue * 1000)
      const year = date.getFullYear()
      const month = String(date.getMonth() + 1).padStart(2, '0')
      const day = String(date.getDate()).padStart(2, '0')
      const hour = String(date.getHours()).padStart(2, '0')
      const minute = String(date.getMinutes()).padStart(2, '0')
      const second = String(date.getSeconds()).padStart(2, '0')
      return `${year}-${month}-${day} ${hour}:${minute}:${second}`
    },
    // openMemoryFragmentById 根据知识片段ID在新页卡中打开详情页。
    openMemoryFragmentById(fragmentId) {
      const routeInfo = this.$router.resolve({
        path: '/MemoryFragment',
        query: {
          fragment_id: String(fragmentId),
          hide_menu: '1',
        },
      })
      window.open(routeInfo.href, '_blank')
    },
    // 处理SSE推送的Shell连接状态更新
    handleSshConnectionsUpdate(data) {
      if (!data || typeof data !== 'object') {
        this.sshConnectionCount = 0
        this.sshConnections = []
        return
      }
      const list = Array.isArray(data.connections) ? data.connections : []
      this.sshConnectionCount = data.total || list.length
      this.sshConnections = list
    },
    // 处理SSE推送的异步任务状态更新
    handleAsyncTasksUpdate(data) {
      if (!data || typeof data !== 'object') {
        return
      }
      const list = Array.isArray(data.list) ? data.list : []
      this.processAsyncTaskNotifications(list)
      this.asyncTaskList = list
      this.asyncTaskSummary = {
        pending_count: Number(data.pending_count || 0),
        await_confirm_count: Number(data.await_confirm_count || 0),
        running_count: Number(data.running_count || 0),
        failed_count: Number(data.failed_count || 0),
        total: Number(data.total || list.length),
      }
      // 如果当前有选中的任务，更新其详情
      if (this.asyncTaskSelectedId) {
        const activeTask = list.find(item => Number(item.id) === Number(this.asyncTaskSelectedId))
        if (activeTask) {
          // 当 result_payload 发生变化时重新解析，避免任务完成后详情内容仍为空
          let resultPayloadMap = this.asyncTaskDetail.result_payload_map || {}
          if (activeTask.result_payload !== this.asyncTaskDetail.result_payload) {
            try {
              resultPayloadMap = JSON.parse(String(activeTask.result_payload || '{}'))
            } catch (e) {
              resultPayloadMap = {}
            }
          }
          this.asyncTaskDetail = {
            ...this.asyncTaskDetail,
            ...activeTask,
            result_payload_map: resultPayloadMap
          }
        }
      }
    },
    // 刷新SSH连接状态（仅用于显示loading状态）
    refreshSshConnections(showLoading) {
      if (showLoading) {
        this.sshConnectionsLoading = true
        // 关闭loading
        setTimeout(() => {
          this.sshConnectionsLoading = false
        }, 500)
      }
    },
    handleGitPendingStatusUpdate(data) {
      const exceedsLimit = !!data?.main_db_storage_alert?.exceeds_limit
      this.mainDbStorageAlertVisible = exceedsLimit
      if (this.$eventBus) {
        this.$eventBus.emit('main_db_storage_alert_changed', data?.main_db_storage_alert || { exceeds_limit: exceedsLimit })
      }
    },
    handleMainDbStorageAlertEvent(payload) {
      this.mainDbStorageAlertVisible = !!payload?.exceeds_limit
    },
    handleAgentCliUnreadUpdate(data) {
      if (!data || data.type !== 'agent_cli_unread_home') return
      this.setAgentCliUnreadTotal(data.agent_cli_unread)
    },
    handleWorkflowUnreadUpdate(data) {
      if (!data || data.type !== 'workflow_unread_snapshot') return
      this.setWorkflowUnreadSummary(data.workflow_task_badges || {}, data.workflow_menu_badge || {})
    },
    setAgentCliUnreadTotal(total) {
      const nextTotal = Math.max(0, Number(total || 0))
      this.agentCliUnreadTotal = nextTotal
      this.agentCliUnreadVisible = nextTotal > 0
    },
    setWorkflowUnreadSummary(unreadMap, menuBadge = {}) {
      const nextMap = {}
      let total = 0
      Object.keys(unreadMap || {}).forEach((key) => {
        const taskId = Number(key || 0)
        const unread = Math.max(0, Number(unreadMap[key] || 0))
        if (taskId <= 0 || unread <= 0) return
        nextMap[taskId] = unread
        total += unread
      })
      const hasMenuUnread = Object.prototype.hasOwnProperty.call(menuBadge || {}, 'has_unread')
      const hasMenuTotal = Object.prototype.hasOwnProperty.call(menuBadge || {}, 'unread_total')
      const nextTotal = hasMenuTotal ? Math.max(0, Number(menuBadge.unread_total || 0)) : total
      const nextVisible = hasMenuUnread ? !!menuBadge.has_unread : nextTotal > 0
      this.workflowUnreadMap = nextMap
      this.workflowUnreadTotal = nextTotal
      this.workflowUnreadVisible = nextVisible
    },
    setWorkflowUnreadByTask(taskId, total) {
      const normalizedTaskId = Number(taskId || 0)
      if (normalizedTaskId <= 0) return
      const nextMap = { ...this.workflowUnreadMap }
      const nextTotal = Math.max(0, Number(total || 0))
      if (nextTotal > 0) {
        nextMap[normalizedTaskId] = nextTotal
      } else {
        delete nextMap[normalizedTaskId]
      }
      this.setWorkflowUnreadSummary(nextMap)
    },
    loadWorkflowUnreadSummary() {
      const workflowTaskIdSet = new Set()
      let pending = 2
      const finalize = () => {
        pending -= 1
        if (pending > 0) return
        const workflowTaskIds = Array.from(workflowTaskIdSet)
        if (workflowTaskIds.length === 0) {
          this.setWorkflowUnreadSummary({})
          return
        }
        taskWorkflowApi.TaskWorkflowBatchNodeStatus(workflowTaskIds, (response) => {
          if (!(response && response.ErrCode === 0 && response.Data)) return
          this.setWorkflowUnreadSummary(response.Data.unread_count_map || {})
        })
      }
      const collectWorkflowTasks = (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) return
        const taskList = Array.isArray(response.Data.task_list) ? response.Data.task_list : []
        taskList.forEach((task) => {
          if (Number(task?.use_workflow) !== 0) {
            const taskId = Number(task?.id || 0)
            if (taskId > 0) {
              workflowTaskIdSet.add(taskId)
            }
          }
        })
      }
      homeTaskApi.HomeTaskList(0, (response) => {
        collectWorkflowTasks(response)
        finalize()
      })
      homeTaskApi.HomeTaskList(1, (response) => {
        collectWorkflowTasks(response)
        finalize()
      })
    },
    loadAgentCliUnreadSummary() {
      agentCliApi.AgentCliList((response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) return
        const items = Array.isArray(response.Data.list) ? response.Data.list : []
        const rows = items.filter(item => Number(item?.id || 0) > 0)
        if (rows.length === 0) {
          this.setAgentCliUnreadTotal(0)
          return
        }
        let pending = rows.length
        let unreadTotal = 0
        rows.forEach((row) => {
          agentCliApi.AgentChatListByAgentCli(row.id, (chatResponse) => {
            if (chatResponse && chatResponse.ErrCode === 0 && chatResponse.Data) {
              const list = Array.isArray(chatResponse.Data.list) ? chatResponse.Data.list : []
              unreadTotal += list.filter(item => item.is_read === false && item.status !== 'running').length
            }
            pending -= 1
            if (pending === 0) {
              this.setAgentCliUnreadTotal(unreadTotal)
            }
          })
        })
      })
    },
    onMenuNativeClick(event) {
      this.menuCtrlKey = event.ctrlKey || event.metaKey
    },
    handleSelect(key, keyPath) {
      if (keyPath[0].indexOf('Doc-') >= 0) {
        return
      }
      if (keyPath[0].indexOf('Ignore-') >= 0) {
        return;
      }
      // Ctrl/Cmd+Click: 在新标签页打开
      if (this.menuCtrlKey) {
        const resolved = this.$router.resolve(keyPath[0])
        window.open(resolved.href, '_blank')
        this.menuCtrlKey = false
        return
      }
      this.menuName = keyPath[0]
      this.$router.push(keyPath[0])
    },
  },
  beforeUnmount() {
    if (this.sshConnectionTimer) {
      clearInterval(this.sshConnectionTimer)
      this.sshConnectionTimer = null
    }
    if (this.$eventBus) {
      this.$eventBus.off('safe_auth_required', this.showSafeLogin)
      this.$eventBus.off('main_db_storage_alert_changed', this.handleMainDbStorageAlertEvent)
    }
    sseDistribute.UnRegisterReceive('agent_cli_unread_home')
    sseDistribute.UnRegisterReceive('workflow_unread_home_menu')
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
    List,
    ToolsIcon,
    GitActionButton,
    Markdown,
    DiffMarkdown,
    MarkdownRenderer,
    Tools,
    Clipboard,
  },
}
</script>

<style scoped src="@/css/components/Home.css"></style>

