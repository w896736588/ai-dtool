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
          <div v-if="gitPendingStatus.memoryPending" class="menu-countdown-bar">
            <div class="menu-countdown-bar__fill" :style="{ width: memoryCountdownPercent + '%' }"></div>
          </div>
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
          <div v-if="gitPendingStatus.mainDBPending" class="menu-countdown-bar">
            <div class="menu-countdown-bar__fill" :style="{ width: mainDBCountdownPercent + '%' }"></div>
          </div>
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
                                <a v-if="task.tapd_url" :href="task.tapd_url" target="_blank" class="home-task-card__tapd-link">TAPD需求</a>
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
    title="任务清单设置"
    width="80%"
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
import asyncTaskApi from '@/utils/base/async_task'
import sseDistribute from '@/utils/base/sse_distribute'
const { mergeHomeTaskFragmentOptions } = require('@/utils/home_task_fragment_options.cjs')
const {
  HOME_DASHBOARD_PAGE_SWITCH_HOT_ZONE_WIDTH,
  resolveHomeDashboardPageSwitchBlocker,
  isHomeDashboardPageSwitchHotZone,
} = require('@/utils/home_dashboard_wheel.cjs')
import Tools from "@/components/Tools.vue";
import Markdown from '@/components/Markdown.vue'
import GitActionButton from "@/components/base/GitActionButton.vue";
import SettingsDialog from '@/components/base/SettingsDialog.vue'
import HomeTaskReportSetting from '@/components/set/home_task_report.vue'
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
const HOME_TASK_DAILY_REPORT_SUCCESS_MESSAGE = '工作日报任务已加入异步任务列表'
const HOME_TASK_DAILY_REPORT_FAILED_MESSAGE = '工作日报生成失败'
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
    tapd_url: '',
  }
}

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
      ASYNC_TASK_ACTION_SAVE_DAILY_REPORT,
      ASYNC_TASK_ACTION_OVERWRITE_MEMORY_FRAGMENT,
      ASYNC_TASK_ACTION_OVERWRITE_FRAGMENT_WITH_SCRAPE,
      ASYNC_TASK_ACTION_DISCARD,
      ASYNC_TASK_STATUS_AWAIT_CONFIRM,
      ASYNC_TASK_STATUS_PENDING,
      ASYNC_TASK_TYPE_DAILY_REPORT,
      ASYNC_TASK_TYPE_MEMORY_ARRANGE,
      ASYNC_TASK_TYPE_MAIN_DB_SYNC,
      ASYNC_TASK_TYPE_MEMORY_DB_SYNC,
      ASYNC_TASK_TYPE_TAPD_SCRAPE,
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
      asyncTaskDialogVisible: false,
      asyncTaskLoading: false,
      asyncTaskActing: false,
      asyncTaskDeleting: false,
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
      gitPendingStatus: {
        mainDBPending: false,
        memoryPending: false,
        mainDBNextPush: 0,
        memoryNextPush: 0,
        mainDBInterval: 600,
        memoryInterval: 60,
      },
      countdownNow: Math.floor(Date.now() / 1000),
      countdownTimer: null,
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
    // memoryCountdownPercent 计算记忆库倒计时进度百分比（0~100）。
    memoryCountdownPercent() {
      if (!this.gitPendingStatus.memoryPending || !this.gitPendingStatus.memoryNextPush || !this.gitPendingStatus.memoryInterval) return 0
      const remain = this.gitPendingStatus.memoryNextPush - this.countdownNow
      const total = this.gitPendingStatus.memoryInterval
      if (remain <= 0) return 100
      const pct = Math.round((1 - remain / total) * 100)
      return Math.max(0, Math.min(100, pct))
    },
    // mainDBCountdownPercent 计算主库倒计时进度百分比（0~100）。
    mainDBCountdownPercent() {
      if (!this.gitPendingStatus.mainDBPending || !this.gitPendingStatus.mainDBNextPush || !this.gitPendingStatus.mainDBInterval) return 0
      const remain = this.gitPendingStatus.mainDBNextPush - this.countdownNow
      const total = this.gitPendingStatus.mainDBInterval
      if (remain <= 0) return 100
      const pct = Math.round((1 - remain / total) * 100)
      return Math.max(0, Math.min(100, pct))
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
    // 注册 Git 待提交状态 SSE 监听（替代轮询）
    sseDistribute.RegisterReceive('git_pending_status', function(data) {
      _that.handleGitPendingStatusUpdate(data)
    })
    this.loadHomeTaskFragmentOptions()
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
    this.ensureAsyncTaskNotificationPermission()
    this.menuName = this.$route.path || '/Dashboard'
    window.addEventListener('resize', function () {});
    // 监听全局登录失效事件
    if (this.$eventBus) {
      this.$eventBus.on('safe_auth_required', this.showSafeLogin)
    }
    // 启动倒计时每秒更新
    this.countdownTimer = setInterval(function() {
      _that.countdownNow = Math.floor(Date.now() / 1000)
    }, 1000)
  },
  provide() {
    return {
      showTerminal: this.showTerminal,
      resizeTerminal: this.resizeTerminal,
    };
  },
  methods: {
    // 处理 SSE 推送的 Git 待提交状态数据
    handleGitPendingStatusUpdate: function (data) {
      if (!data) return
      this.gitPendingStatus.mainDBPending = !!data.main_db_pending
      this.gitPendingStatus.memoryPending = !!data.memory_pending
      this.gitPendingStatus.mainDBNextPush = data.main_db_next_push || 0
      this.gitPendingStatus.memoryNextPush = data.memory_next_push || 0
      this.gitPendingStatus.mainDBInterval = data.main_db_interval || 600
      this.gitPendingStatus.memoryInterval = data.memory_interval || 60
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
    handleHomeTaskTabChange(tabName) {
      if (tabName === HOME_TASK_TAB_ACTIVE) {
        this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
        return
      }
      this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
    },
    // isDashboardCommandRunning 判断首页命令区是否存在执行中的命令，执行中时滚轮不触发整屏切换。
    isDashboardCommandRunning() {
      const dashboardRef = this.$refs.currentRef
      const dashboardIsExecuting = dashboardRef?.isExecuting
      if (dashboardIsExecuting === true) {
        return true
      }
      if (
        dashboardIsExecuting &&
        typeof dashboardIsExecuting === 'object' &&
        dashboardIsExecuting.value === true
      ) {
        return true
      }
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
      const blockingScrollableAncestor = resolveHomeDashboardPageSwitchBlocker(event.target, deltaY, currentTarget)
      if (!isRightHotZone && blockingScrollableAncestor) {
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
        this.loadAsyncTaskSummary(false)
      })
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
        if (action === ASYNC_TASK_ACTION_SAVE_DAILY_REPORT || action === ASYNC_TASK_ACTION_OVERWRITE_MEMORY_FRAGMENT || action === ASYNC_TASK_ACTION_OVERWRITE_FRAGMENT_WITH_SCRAPE) {
          // 中文注释：首页任务卡片会内嵌知识片段摘要，这里立即刷新避免覆盖原文后还显示旧内容。
          // English comment: Home task cards embed fragment previews, so refresh right away after async writes.
          this.refreshAllHomeTaskList()
        }
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
        tapd_url: task.tapd_url || '',
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
        tapd_url: String(this.homeTaskForm.tapd_url || '').trim(),
        api_host: base.GetApiHost() || window.location.origin,
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
          // 更新当前详情中的状态字段，保持其他数据不变
          this.asyncTaskDetail = {
            ...this.asyncTaskDetail,
            ...activeTask,
            result_payload_map: this.asyncTaskDetail.result_payload_map || {}
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
    if (this.countdownTimer) {
      clearInterval(this.countdownTimer)
      this.countdownTimer = null
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
    DiffMarkdown,
    MarkdownRenderer,
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

/* Git 待提交变更倒计时进度条 */
.menu-countdown-bar {
  position: absolute;
  bottom: 2px;
  left: 12px;
  right: 12px;
  height: 2px;
  border-radius: 1px;
  background: rgba(111, 142, 103, 0.12);
  overflow: hidden;
}
.menu-countdown-bar__fill {
  height: 100%;
  border-radius: 1px;
  background: linear-gradient(90deg, #6f8e67 0%, #89a27c 100%);
  transition: width 1s linear;
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

.async-task-entry {
  width: 100%;
  margin-top: 10px;
  position: relative;
  overflow: hidden;
}

.footer-action--sand {
  color: #5f6158;
}

.async-task-entry--idle {
  background: linear-gradient(180deg, #f3f4f5 0%, #e8ebee 100%);
  color: #61656d;
}

.async-task-entry--active {
  background: linear-gradient(180deg, #edf7ee 0%, #dff0e2 100%);
  color: #476651;
}

.async-task-entry--await-confirm {
  background: linear-gradient(180deg, #fbf2e4 0%, #f6e5cb 100%);
  color: #7a5c33;
}

.async-task-entry--pending {
  background: linear-gradient(180deg, #eaf2ff 0%, #dbe8ff 100%);
  color: #3d5f92;
}

.async-task-entry--failed {
  background: linear-gradient(180deg, #fbeaea 0%, #f4d9d9 100%);
  color: #835252;
}

.async-task-entry--running {
  box-shadow: 0 0 0 1px rgba(115, 164, 122, 0.22), 0 0 0 8px rgba(115, 164, 122, 0.08);
  animation: async-task-entry-pulse 1.6s ease-in-out infinite;
}

.async-task-entry--running::after {
  content: '';
  position: absolute;
  inset: -30%;
  background: radial-gradient(circle, rgba(138, 188, 143, 0.18) 0%, rgba(138, 188, 143, 0) 62%);
  animation: async-task-entry-sheen 2.1s linear infinite;
  pointer-events: none;
}

.footer-action__meta {
  display: block;
  margin-top: 2px;
  font-size: 11px;
  opacity: 0.78;
}

.async-task-entry__title {
  width: 100%;
}

.async-task-entry__label {
  display: block;
}

.async-task-entry__summary {
  display: inline-flex;
  align-items: baseline;
  margin-top: 6px;
}

.async-task-entry__digit {
  display: inline-block;
  font-size: 13px;
  font-weight: 700;
  line-height: 1.4;
  white-space: nowrap;
  cursor: help;
}

.async-task-entry__digit--running {
  color: #2d6a3f;
}

.async-task-entry__digit--pending {
  color: #2f5ca8;
}

.async-task-entry__digit--await-confirm {
  color: #9a5b08;
}

.async-task-entry__digit--failed {
  color: #b13c3c;
}

.async-task-entry__slash {
  display: inline-block;
  margin: 0 1px;
  font-size: 12px;
  font-weight: 400;
  color: #8a8d83;
  line-height: 1.4;
}

.async-task-entry__spinner {
  display: inline-block;
  width: 9px;
  height: 9px;
  margin-left: 6px;
  border-radius: 999px;
  border: 2px solid rgba(71, 102, 81, 0.24);
  border-top-color: #64916e;
  animation: async-task-spinner-rotate 0.9s linear infinite;
  vertical-align: middle;
}

.async-task-toolbar {
  display: flex;
  gap: 10px;
  margin-bottom: 14px;
}

.async-task-layout {
  display: grid;
  grid-template-columns: 280px 1fr;
  gap: 16px;
  min-height: 480px;
}

.async-task-list {
  border-right: 1px solid rgba(210, 214, 203, 0.9);
  padding-right: 12px;
  overflow: auto;
}

.async-task-item {
  width: 100%;
  border: 1px solid rgba(218, 223, 212, 0.92);
  background: #fbfbf7;
  border-radius: 12px;
  padding: 12px;
  margin-bottom: 10px;
  text-align: left;
  cursor: pointer;
}

.async-task-item--active {
  border-color: #94ad78;
  box-shadow: 0 0 0 1px rgba(148, 173, 120, 0.32);
}

.async-task-item__header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.async-task-item__title {
  font-size: 14px;
  font-weight: 700;
  color: #374332;
}

.async-task-item__meta,
.async-task-item__time,
.async-task-detail__meta {
  font-size: 12px;
  color: #697262;
}

.async-task-item__meta,
.async-task-item__time {
  margin-top: 6px;
}

.async-task-detail {
  overflow: auto;
}

.async-task-detail__header {
  display: flex;
  align-items: flex-start;
  justify-content: space-between;
  gap: 16px;
  margin-bottom: 14px;
}

.async-task-detail__title {
  font-size: 18px;
  font-weight: 700;
  color: #31402e;
}

.async-task-detail__meta {
  display: flex;
  gap: 12px;
  flex-wrap: wrap;
  margin-top: 6px;
}

.async-task-detail__section-title {
  margin-bottom: 8px;
  font-size: 13px;
  font-weight: 700;
  color: #4a5441;
}

.async-task-detail__content {
  margin-bottom: 14px;
}

.async-task-detail__pre {
  margin: 0;
  padding: 14px;
  border-radius: 12px;
  background: #f5f6ef;
  white-space: pre-wrap;
  word-break: break-word;
  max-height: 420px;
  overflow: auto;
  min-height: 220px;
}

/* TAPD 抓取预览：查看/源码切换按钮区 */
.async-task-detail__view-toggle {
  display: inline-flex;
  gap: 4px;
  margin-left: 12px;
  vertical-align: middle;
}

/* 查看模式下 MarkdownRenderer 容器样式 */
.async-task-detail__markdown {
  max-height: 420px;
  min-height: 220px;
}

/* section-title 内嵌切换按钮时用 flex 水平排列 */
.async-task-detail__section-title:has(.async-task-detail__view-toggle) {
  display: flex;
  align-items: center;
}

.async-task-detail__note {
  padding: 12px 14px;
  border: 1px solid #e4ebf3;
  border-radius: 12px;
  background: linear-gradient(180deg, #f6f9ff 0%, #eef4ff 100%);
  color: #41597e;
  line-height: 1.7;
}

.async-task-detail__sub-meta {
  margin-top: 10px;
  font-size: 12px;
  color: #667891;
}

.async-task-detail__actions {
  display: flex;
  gap: 10px;
  margin-top: 16px;
}

.async-task-detail__error,
.async-task-empty {
  padding: 18px;
  border-radius: 12px;
  background: #f7f4ea;
  color: #705e42;
}

@keyframes async-task-entry-pulse {
  0%, 100% {
    transform: translateY(0);
  }
  50% {
    transform: translateY(-1px);
  }
}

@keyframes async-task-entry-sheen {
  0% {
    transform: translateX(-18%) translateY(0);
    opacity: 0.45;
  }
  50% {
    opacity: 0.9;
  }
  100% {
    transform: translateX(18%) translateY(0);
    opacity: 0.35;
  }
}

@keyframes async-task-spinner-rotate {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
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


.home-task-card__tapd-link {
  color: #3a7a3a;
  text-decoration: none;
  margin-left: 8px;
}
.home-task-card__tapd-link:hover {
  text-decoration: underline;
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

/* Safe 登录弹窗样式 */
.safe-login-dialog :deep(.el-dialog) {
  border-radius: 16px;
  overflow: hidden;
  box-shadow: 0 24px 48px rgba(0, 0, 0, 0.2);
}

/* 登录弹窗背景模糊效果 */
.safe-login-modal {
  backdrop-filter: blur(8px) !important;
  -webkit-backdrop-filter: blur(8px) !important;
  background-color: rgba(0, 0, 0, 0.35) !important;
}

.safe-login-content {
  padding: 20px 10px;
}

.safe-login-desc {
  margin: 0 0 16px 0;
  font-size: 14px;
  color: #5a6d5a;
  text-align: center;
}

.safe-login-error {
  margin: 12px 0 0 0;
  font-size: 13px;
  color: #d9534f;
  text-align: center;
  min-height: 20px;
}

.safe-login-dialog :deep(.el-input__wrapper) {
  border-radius: 10px;
}

.safe-login-dialog :deep(.el-input__inner) {
  height: 44px;
  font-size: 15px;
}

.safe-login-dialog :deep(.el-dialog__header) {
  padding: 20px 20px 10px;
  margin-right: 0;
}

.safe-login-dialog :deep(.el-dialog__title) {
  font-weight: 600;
  color: #3a5a3a;
}

.safe-login-dialog :deep(.el-dialog__footer) {
  padding: 10px 20px 20px;
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

  .async-task-layout {
    grid-template-columns: 1fr;
  }

  .async-task-list {
    border-right: none;
    border-bottom: 1px solid rgba(210, 214, 203, 0.9);
    padding-right: 0;
    padding-bottom: 12px;
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

