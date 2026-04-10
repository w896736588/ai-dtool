<template>
  <div class="layout-container" :class="{ 'layout-container--hide-sidebar': hideAppSidebar }">
    <!-- 左侧菜单 -->
    <aside v-if="!hideAppSidebar" class="sidebar">
      <div class="sidebar-header">
        <span class="logo">🛠️</span>
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
        <el-menu-item index="/Dashboard">
          <el-icon><HomeFilled /></el-icon>
          <span>首页</span>
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
          <span>自定义网页</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('variable')" index="/Variable">
          <el-icon><Document /></el-icon>
          <span>自定义脚本</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('memory_fragment')" index="/MemoryFragment">
          <el-icon><Memo /></el-icon>
          <span>知识片段</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('docker')" index="/Docker">
          <el-icon><Box /></el-icon>
          <span>Docker</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('api')" index="/Api">
          <el-icon><Connection /></el-icon>
          <span>接口开发</span>
        </el-menu-item>
        <el-menu-item v-if="checkModuleOpen('shellout')" index="/shellout">
          <el-icon><Monitor /></el-icon>
          <span>终端输出</span>
        </el-menu-item>
        <el-menu-item index="/Set" class="menu-item-settings">
          <el-icon><Setting /></el-icon>
          <span>配置</span>
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
                          <!-- <div v-if="task.memory_fragment_id > 0" class="home-task-card__memory"> -->
                            <!-- <div class="home-task-card__memory-label">关联知识片段</div> -->
                            <!-- <div class="home-task-card__memory-title"> -->
                              <!-- {{ task.memory_fragment?.title || `#${task.memory_fragment_id}` }} -->
                            <!-- </div> -->
                            <!-- <div v-if="hasHomeTaskMemoryFragment(task) && task.memory_fragment?.content" class="home-task-card__memory-content">
                              <pre class="memory-content-text">{{ getFragmentPreview(task.memory_fragment.content, task.id) }}</pre>
                              <button
                                v-if="isFragmentExpandable(task.memory_fragment.content)"
                                type="button"
                                class="memory-content-toggle"
                                @click="toggleFragmentExpand(task.id)"
                              >
                                {{ homeTaskExpandedFragments[task.id] ? '收起' : '展开' }}
                              </button>
                            </div> -->
                            <!-- <div v-if="Array.isArray(task.memory_fragment?.tags) && task.memory_fragment.tags.length > 0" class="home-task-card__memory-tags">
                              <el-tag
                                v-for="tag in task.memory_fragment.tags"
                                :key="`${task.id}-${tag}`"
                                size="small"
                                effect="plain"
                              >
                                {{ tag }}
                              </el-tag>
                            </div> -->
                          <!-- </div> -->
                          
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
import memoryFragmentApi from '@/utils/base/memory_fragment'
import sseDistribute from '@/utils/base/sse_distribute'
const { mergeHomeTaskFragmentOptions } = require('@/utils/home_task_fragment_options.cjs')
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
// HOME_TASK_ARCHIVED_* 对应后端归档状态常量。
const HOME_TASK_ARCHIVED_NO = 0
const HOME_TASK_ARCHIVED_YES = 1
// HOME_TASK_STATUS_* 与后端状态常量保持一致。
const HOME_TASK_STATUS_TODO = '待开始'
const HOME_TASK_STATUS_DEVELOPING = '开发中'
const HOME_TASK_STATUS_SELF_TESTING = '自测中'
const HOME_TASK_STATUS_SELF_TESTED = '自测完'
const HOME_TASK_STATUS_PENDING_INTEGRATION = '待对接'
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
const HOME_TASK_DELETE_CONFIRM_MESSAGE_PREFIX = '确定要删除任务"'
const HOME_TASK_DELETE_CONFIRM_MESSAGE_SUFFIX = '"吗？该操作不可恢复。'
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
// HOME_DASHBOARD_RUNNING_SELECTOR 用于识别首页命令面板是否仍有执行中的任务标记。
const HOME_DASHBOARD_RUNNING_SELECTOR = '.message-list .command-status-running, .message-list .result-line-running'
// HOME_DASHBOARD_WHEEL_SWITCH_THRESHOLD 控制翻屏触发的滚轮阈值，避免轻微触控板抖动误切屏。
const HOME_DASHBOARD_WHEEL_SWITCH_THRESHOLD = 24
// HOME_DASHBOARD_ANIMATION_DURATION_MS 与 CSS 动画时长保持一致。
const HOME_DASHBOARD_ANIMATION_DURATION_MS = 560
// HOME_TASK_EMPTY_START_DATE 表示未设置开始日期。
const HOME_TASK_EMPTY_START_DATE = ''
// HOME_TASK_STATUS_OPTIONS 用于渲染状态选择和快捷状态按钮。
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
  }
}

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
    }
  },
  computed: {
    // hideAppSidebar 控制某些独立页卡场景下隐藏应用左侧主菜单。
    hideAppSidebar() {
      return String(this.$route.query.hide_menu || '') === '1'
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
      this.menuName = newPath || '/Dashboard'
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
    // 注册Shell连接状态SSE监听
    sseDistribute.RegisterReceive('shell_connections', function(data, type, distributeId) {
      _that.handleSshConnectionsUpdate(data)
    })
    this.loadHomeTaskFragmentOptions()
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
    this.menuName = this.$route.path || '/Dashboard'
    window.addEventListener('resize', function () {});
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
    // isDashboardCommandRunning 判断首页命令区是否存在执行中的命令，执行中时滚轮不触发整屏切换。
    isDashboardCommandRunning() {
      if (!(this.$el instanceof HTMLElement)) {
        return false
      }
      return !!this.$el.querySelector(HOME_DASHBOARD_RUNNING_SELECTOR)
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
        if (!isRightHotZone && this.isDashboardCommandRunning()) {
          return
        }
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
      // 在任务清单页面向上滚动时，只有在右侧热区内才允许切换到首页
      if (deltaY < 0 && scrollTop <= 0) {
        if (isRightHotZone) {
          event.preventDefault()
          this.switchHomeDashboardPage(HOME_DASHBOARD_PAGE_COMMAND)
        }
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
    buildHomeTaskFragmentOptionLabel(fragment) {
      const tagText = Array.isArray(fragment?.tags) && fragment.tags.length > 0 ? ` [${fragment.tags.join('、')}]` : ''
      return `${fragment.title || `#${fragment.id}`}${tagText}`
    },
    // openCreateHomeTaskDialog 打开新增任务弹窗，并重置为默认表单。
    openCreateHomeTaskDialog() {
      this.resetHomeTaskForm()
      this.loadHomeTaskFragmentOptions()
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
        // 生成成功后打开日报所属的知识片段
        const fragmentId = this.normalizeHomeTaskMemoryFragmentId(response.Data?.memory_fragment?.file_id || response.Data?.memory_fragment?.id)
        if (fragmentId) {
          this.openMemoryFragmentById(fragmentId)
        }
      })
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
    // closeHomeTaskDialog 关闭任务弹窗，并清空临时表单内容。
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
      }
      this.loadHomeTaskFragmentOptions(task.memory_fragment)
      this.homeTaskDialogVisible = true
    },
    // openHomeTaskMemoryFragment 在新页卡中打开单独的知识片段详情页。
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
    // getHomeTaskMemoryTagText 生成任务状态右侧的知识片段标签文案。
    getHomeTaskMemoryTagText(task) {
      const fragmentId = this.normalizeHomeTaskMemoryFragmentId(task?.memory_fragment?.file_id || task?.memory_fragment_id)
      if (!fragmentId) {
        return ''
      }
      const fragmentTitle = String(task?.memory_fragment?.title || '').trim()
      const displayTitle = fragmentTitle || `#${fragmentId}`
      return `已关联知识片段 "${displayTitle}"`
    },
    // normalizeHomeTaskMemoryFragmentId 统一规范任务关联知识片段ID，避免数字转换导致 UUID 丢失。
    normalizeHomeTaskMemoryFragmentId(rawId) {
      const fragmentId = String(rawId || '').trim()
      if (!fragmentId || fragmentId === '0') {
        return ''
      }
      return fragmentId
    },
    // hasHomeTaskMemoryFragment 判断任务是否已关联知识片段。
    hasHomeTaskMemoryFragment(task) {
      return this.normalizeHomeTaskMemoryFragmentId(task?.memory_fragment?.file_id || task?.memory_fragment_id) !== ''
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
        start_time: this.convertHomeTaskDateToUnix(this.homeTaskForm.start_date),
        memory_fragment_id: String(this.homeTaskForm.memory_fragment_id || '').trim(),
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
        // 编辑任务时触发边框环绕特效
        if (isEdit) {
          const taskId = Number(this.homeTaskForm.id)
          this.triggerHomeTaskEditFeedback(taskId)
        }
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
        // 只更新本地任务状态，不刷新整个列表，避免任务位置更换
        // 后端返回的任务数据直接在 response.Data 中
        const updatedTask = response.Data
        if (updatedTask && updatedTask.id) {
          this.updateHomeTaskInList(updatedTask)
        }
      })
    },
    // updateHomeTaskInList 更新本地列表中的任务数据，保持位置不变
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
    // triggerHomeTaskEditFeedback 触发任务编辑成功后的边框环绕特效
    triggerHomeTaskEditFeedback(taskId) {
      const normalizedId = Number(taskId || 0)
      if (normalizedId <= 0) {
        return
      }
      if (this.homeTaskEditFeedbackTimers[normalizedId]) {
        window.clearTimeout(this.homeTaskEditFeedbackTimers[normalizedId])
      }
      // 使用 Vue 的响应式方式更新对象
      this.homeTaskEditFeedbackMap = { ...this.homeTaskEditFeedbackMap, [normalizedId]: Date.now() }
      this.homeTaskEditFeedbackTimers[normalizedId] = window.setTimeout(() => {
        const { [normalizedId]: _, ...rest } = this.homeTaskEditFeedbackMap
        this.homeTaskEditFeedbackMap = rest
        delete this.homeTaskEditFeedbackTimers[normalizedId]
      }, this.homeTaskEditFeedbackDurationMs)
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
    // getHomeTaskActionButtonVariant 根据当前任务状态返回操作按钮视觉类型。
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
    handleSelect(key, keyPath) {
      if (keyPath[0].indexOf('Doc-') >= 0) {
        return
      }
      if (keyPath[0].indexOf('Ignore-') >= 0) {
        return;
      }
      this.menuName = keyPath[0]
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
.layout-container {
  --layout-sidebar-width: 140px;
  --layout-content-padding: 20px;
  display: flex;
  height: 100vh;
  width: 100%;
  background-color: #f8f8f5;
}

.layout-container--hide-sidebar {
  --layout-content-padding: 0px;
}

.sidebar {
  width: 140px;
  background:
    linear-gradient(180deg, #f9fbf6 0%, #f3f5ee 100%);
  display: flex;
  flex-direction: column;
  flex-shrink: 0;
  border-right: 1px solid rgba(212, 220, 205, 0.9);
  box-shadow: inset -1px 0 0 rgba(255, 255, 255, 0.72);
}

.sidebar-header {
  height: 50px;
  display: flex;
  align-items: center;
  padding: 0 12px;
  border-bottom: 1px solid rgba(214, 223, 208, 0.82);
  background: linear-gradient(180deg, rgba(255, 255, 255, 0.68) 0%, rgba(245, 248, 239, 0.32) 100%);
}

.logo {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 28px;
  height: 28px;
  margin-right: 8px;
  font-size: 16px;
  border-radius: 10px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.95) 0%, rgba(234, 241, 229, 0.98) 100%);
  box-shadow: 0 6px 14px rgba(118, 141, 104, 0.14);
}

.title {
  color: #455446;
  font-size: 15px;
  font-weight: 700;
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
  padding: 8px 10px 10px;
  border-top: 1px solid rgba(214, 223, 208, 0.82);
  display: flex;
  flex-direction: column;
  align-items: stretch;
  background: linear-gradient(180deg, rgba(246, 248, 242, 0.4) 0%, rgba(255, 255, 255, 0.72) 100%);
}

.footer-buttons {
  display: flex;
  flex-direction: column;
  width: 100%;
  gap: 6px;
  margin-bottom: 6px;
}

.footer-action {
  width: 100%;
  min-height: 34px;
  padding: 6px 8px;
  border: 1px solid rgba(126, 145, 117, 0.12);
  border-radius: 10px;
  background: linear-gradient(145deg, rgba(255, 255, 255, 0.98) 0%, rgba(244, 248, 240, 0.98) 100%);
  box-shadow: 0 5px 12px rgba(119, 137, 112, 0.07);
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  text-align: center;
  transition: transform 0.18s ease, box-shadow 0.18s ease, border-color 0.18s ease, background 0.18s ease;
}

.footer-action:hover {
  transform: translateY(-1px);
  box-shadow: 0 8px 18px rgba(119, 137, 112, 0.1);
}

.footer-action:active {
  transform: translateY(0);
}

.footer-action__title {
  display: block;
  font-size: 12px;
  font-weight: 600;
  line-height: 1.35;
  color: #425142;
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
  padding: var(--layout-content-padding);
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
  padding: 8px 0;
}

.sidebar-menu .el-menu-item {
  position: relative;
  height: 42px;
  line-height: 42px;
  margin: 3px 8px;
  border-radius: 12px;
  padding-left: 12px !important;
  border: 1px solid transparent;
  transition: background-color 0.18s ease, transform 0.18s ease, border-color 0.18s ease, box-shadow 0.18s ease;
}

.sidebar-menu .el-menu-item:hover {
  background: linear-gradient(135deg, rgba(237, 246, 232, 0.9) 0%, rgba(250, 252, 246, 0.98) 100%) !important;
  border-color: rgba(168, 194, 149, 0.26);
  transform: translateX(2px);
  box-shadow: 0 6px 14px rgba(141, 163, 126, 0.1);
}

.sidebar-menu .el-menu-item.is-active {
  background: linear-gradient(135deg, rgba(221, 238, 203, 0.96) 0%, rgba(241, 248, 231, 0.98) 100%) !important;
  border-color: rgba(130, 173, 107, 0.3);
  color: #376e38 !important;
  box-shadow: 0 8px 18px rgba(128, 160, 112, 0.12);
}

.sidebar-menu .el-menu-item .el-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 24px;
  height: 24px;
  margin-right: 8px;
  font-size: 16px;
  color: #71836f;
  border-radius: 8px;
  background: rgba(255, 255, 255, 0.55);
  box-shadow: inset 0 0 0 1px rgba(206, 216, 198, 0.45);
  transition: transform 0.18s ease, color 0.18s ease, background-color 0.18s ease, box-shadow 0.18s ease;
}

.sidebar-menu .el-menu-item:hover .el-icon,
.sidebar-menu .el-menu-item.is-active .el-icon {
  transform: translateY(-1px) scale(1.03);
  color: #537953;
  background: rgba(255, 255, 255, 0.88);
  box-shadow: inset 0 0 0 1px rgba(178, 198, 166, 0.42), 0 5px 12px rgba(137, 162, 122, 0.14);
}

.sidebar-menu .el-menu-item.menu-item-common-actions .el-icon {
  color: #4f8a5b;
  filter: drop-shadow(0 2px 4px rgba(92, 143, 101, 0.24));
}

.sidebar-menu .el-menu-item.menu-item-settings .el-icon {
  color: #de8a2a;
  transform-origin: center;
  transition: transform 0.25s ease, color 0.25s ease, filter 0.25s ease;
  filter: drop-shadow(0 2px 6px rgba(222, 138, 42, 0.28));
}

.sidebar-menu .el-menu-item.menu-item-settings:hover .el-icon,
.sidebar-menu .el-menu-item.menu-item-settings.is-active .el-icon {
  color: #f29f38;
  transform: rotate(24deg) scale(1.08);
}

.sidebar-menu .el-menu-item.menu-item-settings.is-active::after {
  content: '';
  position: absolute;
  top: 7px;
  right: 10px;
  width: 7px;
  height: 7px;
  border-radius: 999px;
  background: radial-gradient(circle, #ffd66b 0%, #f29f38 70%, rgba(242, 159, 56, 0) 100%);
  box-shadow: 0 0 10px rgba(255, 214, 107, 0.7);
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

.home-task-form :deep(.el-form-item) {
  margin-bottom: 16px;
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
  position: relative;
  padding: 16px 18px;
  border: 1px solid #e3e9dd;
  border-radius: 16px;
  background: linear-gradient(180deg, #ffffff 0%, #fbfcf8 100%);
  box-shadow: 0 10px 24px rgba(144, 160, 132, 0.08);
}

.home-task-card::before {
  content: '';
  position: absolute;
  inset: -1px;
  border-radius: inherit;
  padding: 2px;
  background: conic-gradient(
    from var(--task-edit-border-angle, 0deg),
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

.home-task-card.edit-success::before {
  opacity: 1;
  animation: task-edit-border-flow 1s linear 1;
}

@property --task-edit-border-angle {
  syntax: '<angle>';
  initial-value: 0deg;
  inherits: false;
}

@keyframes task-edit-border-flow {
  from {
    --task-edit-border-angle: 0deg;
  }
  to {
    --task-edit-border-angle: 360deg;
  }
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

.home-task-card__status-group {
  display: inline-flex;
  align-items: center;
  gap: 8px;
}

.home-task-memory-link-tag {
  cursor: pointer;
  max-width: min(100%, 320px);
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.home-task-card__actions {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  justify-content: flex-end;
  align-items: flex-start;
}

.home-task-card__memory {
  margin-top: 14px;
  padding: 12px 14px;
  border: 1px solid #e7ede1;
  border-radius: 12px;
  background: linear-gradient(180deg, #f8faf4 0%, #f4f7ee 100%);
  color: #536251;
  font-size: 13px;
  line-height: 1.6;
}

.home-task-card__memory-label {
  font-size: 12px;
  color: #7a8974;
  letter-spacing: 0.04em;
}

.home-task-card__memory-title {
  margin-top: 4px;
  font-size: 14px;
  font-weight: 600;
  color: #3c4b3a;
  word-break: break-word;
}

.home-task-card__memory-tags {
  display: flex;
  flex-wrap: wrap;
  gap: 8px;
  margin-top: 10px;
}

.home-task-card__memory-content {
  margin-top: 10px;
  padding: 10px 12px;
  background: #fff;
  border: 1px solid #e0e6da;
  border-radius: 8px;
}

.memory-content-text {
  margin: 0;
  font-size: 12px;
  line-height: 1.6;
  color: #4a5a48;
  white-space: pre-wrap;
  word-break: break-word;
  font-family: inherit;
}

.memory-content-toggle {
  margin-top: 8px;
  padding: 4px 12px;
  font-size: 12px;
  color: #5a8a5a;
  background: #f0f5ee;
  border: 1px solid #d4e0cf;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.memory-content-toggle:hover {
  background: #e8f0e5;
  border-color: #b8d4b0;
}

.home-task-dialog__footer {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

.home-task-dialog :deep(.el-dialog) {
  max-width: min(920px, calc(100vw - 32px));
}

.home-task-dialog :deep(.el-dialog__body) {
  padding-top: 18px;
  padding-bottom: 12px;
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

