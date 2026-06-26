<template>
  <div class="home-task-page">
    <div class="home-task-header-card">
      <div class="home-task-header-title">
        <svg class="home-task-header-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <path d="M9 11L12 14L22 4" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M21 12V19C21 20.1046 20.1046 21 19 21H5C3.89543 21 3 20.1046 3 19V5C3 3.89543 3.89543 3 5 3H16" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
        </svg>
        <span>任务清单</span>
      </div>
      <div class="home-task-header-actions">
        <GitActionButton compact variant="warning" @click="openHomeTaskSettingsPage">
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

    <div class="home-task-body">
      <!-- 左侧状态管理面板 -->
      <div class="home-task-status-panel">
        <div class="home-task-status-panel__header">
          <span class="home-task-status-panel__title">任务状态管理</span>
          <el-button size="small" type="primary" text @click="openAddTaskStatusDialog">
            + 新增
          </el-button>
        </div>
        <div class="home-task-status-panel__list">
          <!-- 固定项：活跃中 -->
          <div
            class="home-task-status-item"
            :class="{ 'home-task-status-item--active': selectedStatus === HOME_TASK_SELECTED_ACTIVE }"
            @click="handleStatusClick(HOME_TASK_SELECTED_ACTIVE)"
          >
            <span class="home-task-status-item__name">活跃中</span>
            <span class="home-task-status-item__count">{{ homeTaskActiveList.length }}</span>
            <span class="home-task-status-item__more" style="visibility:hidden;pointer-events:none;"></span>
          </div>
          <!-- 动态状态列表 -->
          <div
            v-for="(status, idx) in taskStatusList"
            :key="status.id"
            class="home-task-status-item"
            :class="{ 'home-task-status-item--active': selectedStatus === status.name }"
            @click="handleStatusClick(status.name)"
          >
            <span class="home-task-status-item__name">{{ status.name }}</span>
            <span class="home-task-status-item__count">{{ taskStatusCountMap[status.name] || 0 }}</span>
            <el-popover
              placement="right-start"
              :width="120"
              trigger="click"
            >
              <template #reference>
                <span class="home-task-status-item__more" @click.stop>
                  <svg viewBox="0 0 24 24" width="15" height="15" fill="currentColor"><circle cx="12" cy="5" r="2"/><circle cx="12" cy="12" r="2"/><circle cx="12" cy="19" r="2"/></svg>
                </span>
              </template>
              <div class="home-task-status-popover">
                <div class="home-task-status-popover__item" @click.stop="openEditTaskStatusDialog(status)">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 20h9"/><path d="M16.5 3.5a2.121 2.121 0 0 1 3 3L7 19l-4 1 1-4L16.5 3.5z"/></svg>
                  <span>编辑</span>
                </div>
                <div class="home-task-status-popover__item" :class="{ 'home-task-status-popover__item--disabled': idx === 0 }" @click.stop="idx > 0 && moveTaskStatusUp(idx)">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="18 15 12 9 6 15"/></svg>
                  <span>上移</span>
                </div>
                <div class="home-task-status-popover__item" :class="{ 'home-task-status-popover__item--disabled': idx === taskStatusList.length - 1 }" @click.stop="idx < taskStatusList.length - 1 && moveTaskStatusDown(idx)">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="6 9 12 15 18 9"/></svg>
                  <span>下移</span>
                </div>
                <div class="home-task-status-popover__item home-task-status-popover__item--danger" @click.stop="deleteTaskStatus(status)">
                  <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2"><polyline points="3 6 5 6 21 6"/><path d="M19 6v14a2 2 0 0 1-2 2H7a2 2 0 0 1-2-2V6m3 0V4a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2"/></svg>
                  <span>删除</span>
                </div>
              </div>
            </el-popover>
          </div>
          <!-- 固定项：已归档 -->
          <div
            class="home-task-status-item home-task-status-item--archived"
            :class="{ 'home-task-status-item--active': selectedStatus === HOME_TASK_SELECTED_ARCHIVED }"
            @click="handleStatusClick(HOME_TASK_SELECTED_ARCHIVED)"
          >
            <span class="home-task-status-item__name">已归档</span>
            <span class="home-task-status-item__count">{{ homeTaskArchivedTotal }}</span>
            <span class="home-task-status-item__more" style="visibility:hidden;pointer-events:none;"></span>
          </div>
        </div>
      </div>

      <!-- 右侧任务列表 -->
      <!-- 活跃中 / 按状态筛选 -->
      <div v-if="!isArchivedMode" v-loading="homeTaskLoadingActive" class="home-task-content">
        <div v-if="filteredTaskList.length === 0" class="home-task-empty">
          {{ selectedStatus === HOME_TASK_SELECTED_ACTIVE ? '当前没有未归档任务' : selectedStatus === HOME_TASK_SELECTED_ARCHIVED ? '当前没有已归档任务' : '当前没有"' + selectedStatus + '"状态的任务' }}
        </div>
        <div
          v-for="task in filteredTaskList"
          :key="task.id"
          class="home-task-card"
          :class="{ 'edit-success': !!homeTaskEditFeedbackMap[task.id] }"
        >
              <div class="home-task-card__header">
                <div>
                  <div class="home-task-card__title">
                    <span>{{ task.name }}</span>
                  </div>
                  <div class="home-task-card__meta">
                    <span>开始时间：{{ task.start_time_desc || '-' }}</span>
                    <span>最后操作：{{ task.last_operated_at_desc || '-' }}</span>
                    <a v-if="getHomeTaskRequirementUrl(task)" :href="getHomeTaskRequirementUrl(task)" target="_blank" class="home-task-card__tapd-link">{{ getHomeTaskRequirementLabel(task) }}</a>
                  </div>
                  <table v-if="Number(task.use_workflow) !== HOME_TASK_USE_WORKFLOW_NO && getHomeTaskDevConfigTags(task).length > 0" class="home-task-config-table">
                    <thead>
                      <tr>
                        <th v-for="col in homeTaskConfigTableColumns" :key="col.key" class="home-task-config-table__header">{{ col.label }}</th>
                      </tr>
                    </thead>
                    <tr v-for="(group, gIdx) in getHomeTaskDevConfigTags(task)" :key="gIdx">
                      <td v-for="col in homeTaskConfigTableColumns" :key="col.key" class="home-task-config-table__cell">
                        <template v-for="(tag, tagIdx) in group" :key="tagIdx">
                          <el-tooltip
                            v-if="tag.type === col.key"
                            :content="tag.label"
                            placement="top"
                            :disabled="!shouldTruncateTagLabel(tag)"
                          >
                            <span class="home-task-config-tag-wrapper">
                              <el-tag
                                size="small"
                                effect="plain"
                                :type="tag.tagType"
                                :class="['home-task-config-tag', tag.type === 'branch_name' ? 'home-task-config-tag--copy' : '']"
                                @click.stop="tag.type === 'branch_name' ? copyHomeTaskBranchName(tag.label) : navigateToDevConfig(tag)"
                              >
                                {{ shouldTruncateTagLabel(tag) ? truncateTagLabel(tag.label) : tag.label }}
                              </el-tag>
                              <span v-if="tag.type === 'local_dir' && homeTaskLocalDirStatusMap[tag.label] !== undefined" class="home-task-dir-status" :class="homeTaskLocalDirStatusMap[tag.label] ? 'home-task-dir-status--ok' : 'home-task-dir-status--err'">
                                <svg v-if="homeTaskLocalDirStatusMap[tag.label]" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg>
                                <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                              </span>
                              <el-tooltip v-if="tag.type === 'branch_name' && tag.localDir && homeTaskBranchStatusMap[tag.localDir + '|' + tag.label] !== undefined" :content="homeTaskBranchStatusMap[tag.localDir + '|' + tag.label].matched ? '分支匹配' : '当前分支: ' + (homeTaskBranchStatusMap[tag.localDir + '|' + tag.label].current_branch || '未知')" placement="top">
                                <span class="home-task-dir-status" :class="homeTaskBranchStatusMap[tag.localDir + '|' + tag.label].matched ? 'home-task-dir-status--ok' : 'home-task-dir-status--err'">
                                  <svg v-if="homeTaskBranchStatusMap[tag.localDir + '|' + tag.label].matched" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg>
                                  <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                                </span>
                              </el-tooltip>
                            </span>
                          </el-tooltip>
                        </template>
                      </td>
                    </tr>
                  </table>
                </div>
                <div class="home-task-card__actions">
                  <GitActionButton
                    compact
                    variant="primary"
                    :class="{ 'home-task-workflow-btn--unread': hasHomeTaskWorkflowUnread(task) }"
                    :disabled="isHomeTaskBusy(task.id)"
                    v-if="Number(task.use_workflow) !== HOME_TASK_USE_WORKFLOW_NO"
                    @click="openTaskWorkflow(task)"
                  >
                    <template #icon>
                      <svg viewBox="0 0 24 24" width="14" height="14" stroke="currentColor" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="2" y="4" width="7" height="7" rx="1.5"/>
                        <path d="M9 7.5h4"/>
                        <polyline points="12 5.5 14 7.5 12 9.5"/>
                        <rect x="15" y="4" width="7" height="7" rx="1.5"/>
                      </svg>
                    </template>
                    工作流程 {{ getHomeTaskWorkflowCountText(task) }}
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
                      状态变更（{{ task.task_status }}）
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
      <!-- 已归档任务列表 -->
      <div v-if="isArchivedMode" v-loading="homeTaskLoadingArchived" class="home-task-content" ref="archivedListRef" @scroll="handleArchivedScroll">
        <div v-if="homeTaskArchivedList.length === 0 && !homeTaskLoadingArchived" class="home-task-empty">
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
                  <div class="home-task-card__title">
                    <span>{{ task.name }}</span>
                  </div>
                  <div class="home-task-card__meta">
                    <span>开始时间：{{ task.start_time_desc || '-' }}</span>
                    <span>最后操作：{{ task.last_operated_at_desc || '-' }}</span>
                    <a v-if="getHomeTaskRequirementUrl(task)" :href="getHomeTaskRequirementUrl(task)" target="_blank" class="home-task-card__tapd-link">{{ getHomeTaskRequirementLabel(task) }}</a>
                    <span v-if="hasHomeTaskMemoryFragment(task)" class="home-task-card__status-group">
                      <el-tag
                        size="small"
                        effect="plain"
                        class="home-task-memory-link-tag"
                        @click.stop="openHomeTaskMemoryFragment(task)"
                      >
                        {{ getHomeTaskMemoryTagText(task) }}
                      </el-tag>
                    </span>
                  </div>
                  <table v-if="Number(task.use_workflow) !== HOME_TASK_USE_WORKFLOW_NO && getHomeTaskDevConfigTags(task).length > 0" class="home-task-config-table">
                    <thead>
                      <tr>
                        <th v-for="col in homeTaskConfigTableColumns" :key="col.key" class="home-task-config-table__header">{{ col.label }}</th>
                      </tr>
                    </thead>
                    <tr v-for="(group, gIdx) in getHomeTaskDevConfigTags(task)" :key="gIdx">
                      <td v-for="col in homeTaskConfigTableColumns" :key="col.key" class="home-task-config-table__cell">
                        <template v-for="(tag, tagIdx) in group" :key="tagIdx">
                          <el-tooltip
                            v-if="tag.type === col.key"
                            :content="tag.label"
                            placement="top"
                            :disabled="!shouldTruncateTagLabel(tag)"
                          >
                            <span class="home-task-config-tag-wrapper">
                              <el-tag
                                size="small"
                                effect="plain"
                                :type="tag.tagType"
                                :class="['home-task-config-tag', tag.type === 'branch_name' ? 'home-task-config-tag--copy' : '']"
                                @click.stop="tag.type === 'branch_name' ? copyHomeTaskBranchName(tag.label) : navigateToDevConfig(tag)"
                              >
                                {{ shouldTruncateTagLabel(tag) ? truncateTagLabel(tag.label) : tag.label }}
                              </el-tag>
                              <span v-if="tag.type === 'local_dir' && homeTaskLocalDirStatusMap[tag.label] !== undefined" class="home-task-dir-status" :class="homeTaskLocalDirStatusMap[tag.label] ? 'home-task-dir-status--ok' : 'home-task-dir-status--err'">
                                <svg v-if="homeTaskLocalDirStatusMap[tag.label]" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg>
                                <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                              </span>
                              <el-tooltip v-if="tag.type === 'branch_name' && tag.localDir && homeTaskBranchStatusMap[tag.localDir + '|' + tag.label] !== undefined" :content="homeTaskBranchStatusMap[tag.localDir + '|' + tag.label].matched ? '分支匹配' : '当前分支: ' + (homeTaskBranchStatusMap[tag.localDir + '|' + tag.label].current_branch || '未知')" placement="top">
                                <span class="home-task-dir-status" :class="homeTaskBranchStatusMap[tag.localDir + '|' + tag.label].matched ? 'home-task-dir-status--ok' : 'home-task-dir-status--err'">
                                  <svg v-if="homeTaskBranchStatusMap[tag.localDir + '|' + tag.label].matched" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg>
                                  <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                                </span>
                              </el-tooltip>
                            </span>
                          </el-tooltip>
                        </template>
                      </td>
                    </tr>
                  </table>
                </div>
                <div class="home-task-card__actions">
                  <GitActionButton
                    compact
                    variant="primary"
                    :class="{ 'home-task-workflow-btn--unread': hasHomeTaskWorkflowUnread(task) }"
                    :disabled="isHomeTaskBusy(task.id)"
                    v-if="Number(task.use_workflow) !== HOME_TASK_USE_WORKFLOW_NO"
                    @click="openTaskWorkflow(task)"
                  >
                    <template #icon>
                      <svg viewBox="0 0 24 24" width="14" height="14" stroke="currentColor" fill="none" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
                        <rect x="2" y="4" width="7" height="7" rx="1.5"/>
                        <path d="M9 7.5h4"/>
                        <polyline points="12 5.5 14 7.5 12 9.5"/>
                        <rect x="15" y="4" width="7" height="7" rx="1.5"/>
                      </svg>
                    </template>
                    工作流程 {{ getHomeTaskWorkflowCountText(task) }}
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
                      状态变更（{{ task.task_status }}）
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
      </div>
      <div v-if="homeTaskArchivedTotal > homeTaskArchivedPageSize" class="home-task-pagination">
        <div v-if="homeTaskArchivedLoadingMore" class="home-task-loading-more">
          <i class="el-icon-loading" style="margin-right:6px"></i>加载中...
        </div>
        <div v-else-if="homeTaskArchivedNoMore" class="home-task-no-more">
          已加载全部 {{ homeTaskArchivedTotal }} 条归档任务
        </div>
      </div>

    <el-dialog
      v-model="homeTaskDialogVisible"
      :title="homeTaskDialogTitle"
      width="70%"
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
                maxlength="200"
                show-word-limit
                placeholder="例如：整理缓存淘汰策略"
                @keyup.enter="saveHomeTask"
              />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="抓取类型">
              <el-select v-model="homeTaskForm.fetch_type" style="width: 100%">
                <el-option
                  v-for="cfg in requirementFetchConfigs"
                  :key="cfg.type"
                  :label="cfg.name"
                  :value="cfg.type"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item :label="fetchTypeLabel">
              <el-input
                v-model="homeTaskForm.requirement_url"
                :placeholder="'请输入需求文档网页地址'"
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
          <el-col :xs="24" :sm="12" :md="12">
            <el-form-item label="使用工作流程">
              <el-switch
                v-model="homeTaskForm.use_workflow"
                :active-value="HOME_TASK_USE_WORKFLOW_YES"
                :inactive-value="HOME_TASK_USE_WORKFLOW_NO"
                active-text="是"
                inactive-text="否"
                style="--el-switch-on-color: #3a7a3a; --el-color-primary: #3a7a3a"
              />
            </el-form-item>
          </el-col>
        </el-row>
        <!-- 工作流模板选择 -->
        <el-row v-if="homeTaskForm.use_workflow === HOME_TASK_USE_WORKFLOW_YES" :gutter="12">
          <el-col :xs="24" :sm="12" :md="12">
            <el-form-item label="工作流模板">
              <el-select
                v-model="homeTaskForm.workflow_template_id"
                placeholder="请选择工作流模板"
                style="width: 100%"
                :loading="homeTaskTemplateLoading"
                @focus="loadHomeTaskTemplateList"
              >
                <el-option
                  v-for="tpl in homeTaskTemplateList"
                  :key="tpl.id"
                  :label="tpl.name + (tpl.is_default === 1 ? ' (默认)' : '')"
                  :value="tpl.id"
                />
              </el-select>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row v-if="homeTaskUnusedLocalDirs.length > 0" :gutter="12">
          <el-col :span="24">
            <el-form-item label="可用目录">
              <div class="home-task-unused-dirs">
                <el-tag
                  v-for="(dir, idx) in homeTaskUnusedLocalDirs"
                  :key="idx"
                  class="home-task-unused-dir-tag"
                  type="info"
                  @click="copyUnusedLocalDir(dir)"
                >
                  {{ dir }}
                </el-tag>
              </div>
            </el-form-item>
          </el-col>
        </el-row>
        <el-row :gutter="12" v-if="homeTaskForm.use_workflow === HOME_TASK_USE_WORKFLOW_YES">
          <el-col :xs="24" :sm="12" :md="12">
            <el-form-item label="知识文件夹">
              <el-select
                v-model="homeTaskForm.workflow_fragment_folder_name"
                filterable
                style="width: 100%"
                placeholder="请选择知识片段文件夹"
                :loading="homeTaskMemoryFolderLoading"
              >
                <el-option
                  v-for="item in homeTaskMemoryFolderList"
                  :key="item.folder_name"
                  :label="item.name || item.folder_name"
                  :value="item.folder_name"
                />
              </el-select>
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <div style="margin: -4px 0 12px; color: #7a7a6a; font-size: 12px;">
              工作流里创建、读取和回写的知识片段都会使用这个文件夹。
            </div>
          </el-col>
          <el-col :span="24">
            <el-form-item label="开发项目配置">
              <div v-for="(cfg, cfgIdx) in homeTaskForm.dev_configs" :key="cfgIdx" style="border: 2px solid #c8d5b9; border-radius: 4px; padding: 12px 12px 4px; margin-bottom: 10px; position: relative;">
                <el-button
                  v-if="homeTaskForm.dev_configs.length > 1"
                  type="danger"
                  plain
                  size="small"
                  style="position: absolute; top: 4px; right: 4px; padding: 2px 6px; z-index: 1;"
                  @click="removeDevConfig(cfgIdx)"
                >
                  移除
                </el-button>
                <div class="home-task-config-divider">
                  <span class="home-task-config-divider__text">Git项目节点</span>
                </div>
                <el-row :gutter="12">
                  <el-col :xs="24" :sm="12" :md="12">
                    <el-form-item label="Git仓库" label-width="72px">
                      <el-select
                        v-model="cfg.git_id"
                        clearable
                        filterable
                        style="width: 100%"
                        placeholder="选择Git仓库（可选）"
                        :loading="homeTaskGitRepoLoading"
                        @change="handleDevConfigGitChange(cfgIdx)"
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
                  <el-col :xs="24" :sm="12" :md="12">
                    <el-form-item label="Docker" label-width="72px">
                      <el-select
                        v-model="cfg.docker_id"
                        clearable
                        filterable
                        style="width: 100%"
                        placeholder="选择Docker配置（可选）"
                        :loading="homeTaskDockerLoading"
                      >
                        <el-option
                          v-for="item in homeTaskDockerList"
                          :key="item.id"
                          :label="item.name"
                          :value="Number(item.id)"
                        />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :sm="12" :md="12">
                    <el-form-item label="Db" label-width="72px">
                      <el-select
                        v-model="cfg.mysql_id"
                        clearable
                        filterable
                        style="width: 100%"
                        placeholder="选择Db配置（可选）"
                        :loading="homeTaskMysqlLoading"
                      >
                        <el-option
                          v-for="item in homeTaskMysqlList"
                          :key="item.id"
                          :label="item.name"
                          :value="Number(item.id)"
                        />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :sm="12" :md="12">
                    <el-form-item label="分支名" label-width="72px">
                      <div style="display: flex; gap: 8px; width: 100%;">
                        <el-input
                          v-model="cfg.branch_name"
                          clearable
                          placeholder="输入或AI生成分支名"
                        />
                        <el-button
                          class="home-task-ai-btn"
                          :loading="cfg._branchGenerating"
                          @click="generateBranchName(cfgIdx)"
                        >
                          AI生成
                        </el-button>
                      </div>
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :sm="12" :md="12">
                    <el-form-item label="本地目录" label-width="72px">
                      <el-input
                        v-model="cfg.local_dir"
                        clearable
                        style="width: 100%"
                        placeholder="本地项目目录路径（可选）"
                      />
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :sm="12" :md="12">
                    <el-form-item label="父分支" label-width="72px">
                      <el-input
                        v-model="cfg.parent_branch"
                        clearable
                        style="width: 100%"
                        placeholder="父分支名称（可选）"
                      />
                    </el-form-item>
                  </el-col>
                                  <el-col v-if="false" :xs="24" :sm="12" :md="12">
                    <el-form-item label="规则入口" label-width="72px">
                      <el-input
                        v-model="cfg.rule_entry_file"
                        clearable
                        style="width: 100%"
                        placeholder="规则入口文件路径（可选）"
                      />
                    </el-form-item>
                  </el-col>
</el-row>
                <div class="home-task-config-divider">
                  <span class="home-task-config-divider__text">接口开发节点</span>
                </div>
                <el-row :gutter="12">
                  <el-col :xs="24" :sm="12" :md="12">
                    <el-form-item label="接口集合" label-width="72px">
                      <el-select
                        v-model="cfg.collection_id"
                        filterable
                        clearable
                        style="width: 100%"
                        placeholder="选择接口集合（可选）"
                        :loading="homeTaskApiCollectionLoading"
                        @change="handleDevConfigCollectionChange(cfgIdx)"
                      >
                        <el-option
                          v-for="col in homeTaskApiCollectionList"
                          :key="col.id"
                          :label="col.name"
                          :value="Number(col.id)"
                        />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :sm="12" :md="12">
                    <el-form-item label="文件夹" label-width="72px">
                      <el-select
                        v-model="cfg.dir_id"
                        filterable
                        clearable
                        style="width: 100%"
                        placeholder="留空则自动创建"
                        :loading="homeTaskApiFolderLoadingMap && homeTaskApiFolderLoadingMap[cfg.collection_id]"
                        :disabled="!cfg.collection_id"
                      >
                        <el-option
                          v-for="dir in getDevConfigFolders(cfgIdx)"
                          :key="dir.id"
                          :label="dir.name"
                          :value="Number(dir.id)"
                        />
                      </el-select>
                    </el-form-item>
                  </el-col>
                </el-row>
                <div class="home-task-config-divider">
                  <span class="home-task-config-divider__text">自定义网页</span>
                </div>
                <el-row :gutter="12">
                  <el-col :xs="24" :sm="12" :md="12">
                    <el-form-item label="网页" label-width="72px">
                      <el-select
                        v-model="cfg.smart_link_id"
                        clearable
                        filterable
                        style="width: 100%"
                        placeholder="选择自定义网页（可选）"
                        :loading="homeTaskSmartLinkLoading"
                        @change="handleDevConfigSmartLinkChange(cfgIdx)"
                      >
                        <el-option
                              v-for="item in homeTaskSmartLinkList"
                              :key="item.id"
                              :label="item.name"
                              :value="Number(item.id)"
                        />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :sm="12" :md="12">
                    <el-form-item label="网页标签" label-width="72px">
                      <el-select
                        v-model="cfg.smart_link_label"
                        clearable
                        filterable
                        style="width: 100%"
                        placeholder="选择标签（可选）"
                        :disabled="!cfg.smart_link_id"
                        @change="handleDevConfigSmartLinkLabelChange(cfgIdx)"
                      >
                        <el-option
                              v-for="link in getDevConfigSmartLinkLabels(cfgIdx)"
                              :key="link.label"
                              :label="link.label"
                              :value="link.label"
                        />
                      </el-select>
                    </el-form-item>
                  </el-col>
                  <el-col :xs="24" :sm="12" :md="12">
                    <el-form-item label="账号" label-width="72px">
                      <el-select
                        v-model="cfg.smart_link_account"
                        clearable
                        filterable
                        style="width: 100%"
                        placeholder="选择账号（可选）"
                        :disabled="!cfg.smart_link_label"
                      >
                        <el-option
                              v-for="acct in getDevConfigSmartLinkAccounts(cfgIdx)"
                              :key="acct.user_name"
                              :label="acct.user_name"
                              :value="acct.user_name"
                        />
                      </el-select>
                    </el-form-item>
                  </el-col>
                </el-row>
              </div>
              <el-button type="primary" plain size="small" @click="addDevConfig">
                + 添加开发项目配置
              </el-button>
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

    <!-- 状态编辑弹窗 -->
    <el-dialog
      v-model="taskStatusDialogVisible"
      :title="taskStatusDialogTitle"
      width="420px"
      top="25vh"
      destroy-on-close
    >
      <el-form label-width="72px" @submit.prevent>
        <el-form-item label="状态名称">
          <el-input
            v-model="taskStatusForm.name"
            maxlength="50"
            placeholder="例如：待开始"
            @keyup.enter="saveTaskStatus"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="home-task-dialog__footer">
          <GitActionButton compact variant="info" @click="taskStatusDialogVisible = false">
            取消
          </GitActionButton>
          <GitActionButton compact :loading="taskStatusSaving" @click="saveTaskStatus">
            保存
          </GitActionButton>
        </div>
      </template>
    </el-dialog>

    </div><!-- /home-task-body -->

  </div>
</template>

<script>
import base from '../utils/base'
import homeTaskApi from '@/utils/base/home_task'
import gitApi from '@/utils/base/git'
import mysqlSetApi from '@/utils/base/mysql_set'
import apiManagement from '@/utils/base/api'
import dockerApi from '@/utils/base/compose'
import smartLinkSetApi from '@/utils/base/smart_link_set'
import taskWorkflowApi from '@/utils/base/task_workflow'
import workflowTemplateApi from '@/utils/base/workflow_template'
import sseDistribute from '@/utils/base/sse_distribute'
import memoryFragmentApi from '@/utils/base/memory_fragment'
import set from '@/utils/base/git_set'
import GitActionButton from "@/components/base/GitActionButton.vue"

const HOME_TASK_TAB_ACTIVE = 'active'
const HOME_TASK_TAB_ARCHIVED = 'archived'
const HOME_TASK_SELECTED_ACTIVE = 'active'
const HOME_TASK_SELECTED_ARCHIVED = 'archived'
const HOME_TASK_ARCHIVED_NO = 0
const HOME_TASK_ARCHIVED_YES = 1
const HOME_TASK_ARCHIVED_ALL = -1
const HOME_TASK_STATUS_TODO = '待开始'
const HOME_TASK_STATUS_DEVELOPING = '开发中'
const HOME_TASK_STATUS_DEV_COMPLETED = '开发完'
const HOME_TASK_STATUS_SELF_TESTING = '自测中'
const HOME_TASK_STATUS_SELF_TESTED = '自测完'
const HOME_TASK_STATUS_PENDING_INTEGRATION = '待对接'
const HOME_TASK_STATUS_INTEGRATING = '对接中'
const HOME_TASK_STATUS_TESTING = '测试中'
const HOME_TASK_STATUS_RELEASING = '上线中'
const HOME_TASK_STATUS_ONLINE = '已上线'
const HOME_TASK_STATUS_PENDING_TEST = '待测试'
const HOME_TASK_STATUS_ABANDONED = '已废弃'
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
const HOME_TASK_CONFIG_TAG_MAX_LENGTH = 20
const HOME_TASK_USE_WORKFLOW_YES = 1
const HOME_TASK_USE_WORKFLOW_NO = 0
const HOME_TASK_FETCH_TYPE_TAPD = 'tapd'
const HOME_TASK_FETCH_TYPE_ZENTAO = 'zentao'
const HOME_TASK_DEFAULT_WORKFLOW_FRAGMENT_FOLDER = 'fragments'
const HOME_TASK_WORKFLOW_NODE_KEYS = [
  'requirement-fetch',
  'requirement',
  'design',
  'api-dev',
  'api-test-fix',
  'code-review',
  'browser-test',
]
const HOME_TASK_STATUS_OPTIONS = [
  HOME_TASK_STATUS_TODO,
  HOME_TASK_STATUS_DEVELOPING,
  HOME_TASK_STATUS_DEV_COMPLETED,
  HOME_TASK_STATUS_SELF_TESTING,
  HOME_TASK_STATUS_SELF_TESTED,
  HOME_TASK_STATUS_PENDING_INTEGRATION,
  HOME_TASK_STATUS_INTEGRATING,
  HOME_TASK_STATUS_TESTING,
  HOME_TASK_STATUS_PENDING_TEST,
  HOME_TASK_STATUS_RELEASING,
  HOME_TASK_STATUS_ONLINE,
  HOME_TASK_STATUS_ABANDONED,
]

function getTodayDateText() {
  const now = new Date()
  const year = now.getFullYear()
  const month = String(now.getMonth() + 1).padStart(2, '0')
  const day = String(now.getDate()).padStart(2, '0')
  return `${year}-${month}-${day}`
}

function safeParseJSON(text, fallback) {
  try {
    const parsed = JSON.parse(text)
    return Array.isArray(parsed) ? parsed : fallback
  } catch {
    return fallback
  }
}

function createHomeTaskDefaultForm() {
  return {
    id: 0,
    name: '',
    task_status: HOME_TASK_STATUS_TODO,
    start_date: getTodayDateText(),
    created_date: getTodayDateText(),
    fetch_type: HOME_TASK_FETCH_TYPE_TAPD,
    requirement_url: '',
    use_workflow: HOME_TASK_USE_WORKFLOW_YES,
    workflow_fragment_folder_name: HOME_TASK_DEFAULT_WORKFLOW_FRAGMENT_FOLDER,
    workflow_template_id: 0,
    dev_configs: [{ git_id: '', collection_id: '', dir_id: '', docker_id: '', mysql_id: '', local_dir: '', parent_branch: '', branch_name: '', rule_entry_file: '', _branchGenerating: false, smart_link_id: '', smart_link_label: '', smart_link_account: '' }],
  }
}

export default {
  data() {
    return {
      HOME_TASK_TAB_ACTIVE,
      HOME_TASK_TAB_ARCHIVED,
      HOME_TASK_SELECTED_ACTIVE,
      HOME_TASK_SELECTED_ARCHIVED,
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
      HOME_TASK_CONFIG_TAG_MAX_LENGTH,
      HOME_TASK_USE_WORKFLOW_YES,
      HOME_TASK_USE_WORKFLOW_NO,
      HOME_TASK_FETCH_TYPE_TAPD,
      HOME_TASK_FETCH_TYPE_ZENTAO,
      HOME_TASK_DEFAULT_WORKFLOW_FRAGMENT_FOLDER,
      requirementFetchConfigs: [],
      selectedStatus: HOME_TASK_SELECTED_ACTIVE,
      homeTaskDialogVisible: false,
      homeTaskLoadingActive: false,
      homeTaskLoadingArchived: false,
      homeTaskGeneratingDailyReport: false,
      homeTaskSaving: false,
      homeTaskOperatingId: 0,
      homeTaskOperatingType: '',
      homeTaskActiveList: [],
      homeTaskArchivedList: [],
      homeTaskArchivedTotal: 0,
      homeTaskArchivedPage: 1,
      homeTaskArchivedPageSize: 20,
      homeTaskArchivedLoaded: false,
      homeTaskArchivedLoadingMore: false,
      homeTaskArchivedNoMore: false,
      homeTaskForm: createHomeTaskDefaultForm(),
      homeTaskExpandedFragments: {},
      homeTaskEditFeedbackMap: {},
      homeTaskEditFeedbackTimers: {},
      homeTaskEditFeedbackDurationMs: 1000,
      homeTaskWorkflowCountMap: {},
      homeTaskWorkflowUnreadMap: {},
      homeTaskLocalDirStatusMap: {},
      homeTaskBranchStatusMap: {},
      homeTaskUnusedLocalDirs: [],
      homeTaskGitRepoList: [],
      homeTaskGitRepoLoading: false,
      homeTaskApiCollectionList: [],
      homeTaskApiFolderMap: {},
      homeTaskApiCollectionLoading: false,
      homeTaskApiFolderLoadingMap: {},
      homeTaskMysqlList: [],
      homeTaskMysqlLoading: false,
      homeTaskDockerList: [],
      homeTaskDockerLoading: false,
      homeTaskSmartLinkList: [],
      homeTaskSmartLinkLoading: false,
      homeTaskMemoryFolderList: [],
      homeTaskMemoryFolderLoading: false,
      homeTaskTemplateList: [],
      homeTaskTemplateLoading: false,
      homeTaskConfigTableColumns: [
        { key: 'git', label: 'Git仓库' },
        { key: 'api', label: '接口集合' },
        { key: 'parent_branch', label: '父分支' },
        { key: 'branch_name', label: '分支名' },
        { key: 'local_dir', label: '本地目录' },
      ],
      // 任务状态管理
      taskStatusList: [],
      taskStatusLoading: false,
      taskStatusDialogVisible: false,
      taskStatusSaving: false,
      taskStatusForm: { id: 0, name: '' },
    }
  },
  computed: {
    homeTaskDialogTitle() {
      return this.homeTaskForm.id > 0 ? '编辑任务' : '新增任务'
    },
    homeTaskStatusOptions() {
      if (this.taskStatusList.length > 0) {
        return this.taskStatusList.map(s => s.name)
      }
      // 兜底：使用默认选项
      return HOME_TASK_STATUS_OPTIONS
    },
    taskStatusDialogTitle() {
      return this.taskStatusForm.id > 0 ? '编辑状态' : '新增状态'
    },
    taskStatusCountMap() {
      const map = {}
      for (const task of this.homeTaskActiveList) {
        const s = task.task_status || ''
        map[s] = (map[s] || 0) + 1
      }
      return map
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
        groupMap[groupName].push({ label: repo.name, value: Number(repo.id) })
      }
      return groupOrder.map(name => ({ label: name, options: groupMap[name] }))
    },
    isArchivedMode() {
      return this.selectedStatus === HOME_TASK_SELECTED_ARCHIVED
    },
    fetchTypeLabel() {
      const cfg = this.getFetchConfigByType(this.homeTaskForm.fetch_type)
      return cfg ? cfg.name + '地址' : '需求地址'
    },
    filteredTaskList() {
      if (this.selectedStatus === HOME_TASK_SELECTED_ACTIVE) {
        return this.homeTaskActiveList
      }
      return this.homeTaskActiveList.filter(t => t.task_status === this.selectedStatus)
    },
  },
  mounted() {
    // 仅注册 SSE 监听；数据加载由 activated() 统一处理
    // （keep-alive 下首次挂载 mounted 和 activated 都会触发，避免重复请求）
    this.ensureWorkflowUnreadSse()
    this.ensureHomeTaskPageDataSse()
  },
  activated() {
    this.ensureWorkflowUnreadSse()
    this.ensureHomeTaskPageDataSse()
    // 默认只加载活跃任务；归档数量通过轻量接口获取，归档数据点击 tab 时才懒加载
    this.loadHomeTaskCounts()
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
    this.loadTaskStatusList()
  },
  beforeUnmount() {
    this.unregisterWorkflowUnreadSse()
    this.unregisterHomeTaskPageDataSse()
  },
  methods: {
    ensureWorkflowUnreadSse() {
      if (this._workflowUnreadSseId) return
      const nextId = 'workflow_unread_home_task'
      this._workflowUnreadSseId = nextId
      sseDistribute.InitFromLoginStatus().then((created) => {
        if (!created && !sseDistribute.GetSseClientId()) return
        sseDistribute.RegisterReceive(nextId, this.handleWorkflowUnreadSseMessage)
      })
    },
    unregisterWorkflowUnreadSse() {
      if (!this._workflowUnreadSseId) return
      sseDistribute.UnRegisterReceive(this._workflowUnreadSseId)
      this._workflowUnreadSseId = ''
    },
    handleWorkflowUnreadSseMessage(data) {
      if (!data || data.type !== 'workflow_unread_snapshot') {
        return
      }
      this.homeTaskWorkflowUnreadMap = { ...(data.workflow_task_badges || {}) }
    },
    // ========== HomeTask 页面附加数据 SSE ==========
    ensureHomeTaskPageDataSse() {
      if (this._homeTaskPageDataSsePromise) return this._homeTaskPageDataSsePromise
      this._homeTaskPageDataSsePromise = sseDistribute.InitFromLoginStatus().then((created) => {
        if (!created || !sseDistribute.GetSseClientId()) {
          return false
        }
        // 等待 SSE 连接真正建立（onopen 事件触发），避免 POST 请求先于 SSE 握手到达后端
        return sseDistribute.WaitForOpen().then((opened) => {
          if (!opened) return false
          const pDataId = 'home_task_page_data'
          const pDirId = 'home_task_page_data_dir_status'
          const pBranchId = 'home_task_page_data_branch_status'
          this._ssePageDataId = pDataId
          this._sseDirStatusId = pDirId
          this._sseBranchStatusId = pBranchId
          sseDistribute.RegisterReceive(pDataId, this.handleHomeTaskPageData)
          sseDistribute.RegisterReceive(pDirId, this.handleHomeTaskDirStatus)
          sseDistribute.RegisterReceive(pBranchId, this.handleHomeTaskBranchStatus)
          return true
        })
      })
      return this._homeTaskPageDataSsePromise
    },
    unregisterHomeTaskPageDataSse() {
      this._homeTaskPageDataSsePromise = null
      if (this._ssePageDataId) {
        sseDistribute.UnRegisterReceive(this._ssePageDataId)
        this._ssePageDataId = ''
      }
      if (this._sseDirStatusId) {
        sseDistribute.UnRegisterReceive(this._sseDirStatusId)
        this._sseDirStatusId = ''
      }
      if (this._sseBranchStatusId) {
        sseDistribute.UnRegisterReceive(this._sseBranchStatusId)
        this._sseBranchStatusId = ''
      }
    },
    handleHomeTaskPageData(data) {
      if (!data) return
      // Git 仓库列表
      if (Array.isArray(data.git_list) && data.git_list.length > 0) {
        const groupList = Array.isArray(data.git_group_list) ? data.git_group_list : []
        const groupMap = {}
        for (const g of groupList) {
          groupMap[Number(g.id)] = g.name
        }
        this.homeTaskGitRepoList = data.git_list.map(repo => ({
          ...repo,
          git_group_name: groupMap[Number(repo.git_group_id)] || '未分组',
        }))
        this.homeTaskGitRepoLoading = false
      }
      // API 集合列表
      if (Array.isArray(data.api_collection_list)) {
        this.homeTaskApiCollectionList = data.api_collection_list
        this.homeTaskApiCollectionLoading = false
      }
      // Docker 列表
      if (Array.isArray(data.docker_list)) {
        this.homeTaskDockerList = data.docker_list
        this.homeTaskDockerLoading = false
      }
      // MySQL 列表
      if (Array.isArray(data.mysql_list)) {
        this.homeTaskMysqlList = data.mysql_list
        this.homeTaskMysqlLoading = false
      }
      // SmartLink 列表
      if (Array.isArray(data.smart_link_list)) {
        this.homeTaskSmartLinkList = data.smart_link_list
        this.homeTaskSmartLinkLoading = false
      }
      // 记忆库文件夹列表
      if (Array.isArray(data.memory_folder_list)) {
        this.homeTaskMemoryFolderList = data.memory_folder_list
        this.homeTaskMemoryFolderLoading = false
      }
      // 工作流节点状态
      if (data.workflow_node_statuses_map && data.workflow_unread_count_map !== undefined) {
        this.applyWorkflowCountsFromSSE(data.workflow_node_statuses_map, data.workflow_unread_count_map)
      }
    },
    handleHomeTaskDirStatus(data) {
      if (!data || !data.dir_status_map) return
      this.homeTaskLocalDirStatusMap = { ...this.homeTaskLocalDirStatusMap, ...data.dir_status_map }
    },
    handleHomeTaskBranchStatus(data) {
      if (!data || !data.branch_status_map) return
      this.homeTaskBranchStatusMap = { ...this.homeTaskBranchStatusMap, ...data.branch_status_map }
    },
    applyWorkflowCountsFromSSE(nodeStatusesMap, unreadCountMap) {
      const allTasks = [...this.homeTaskActiveList, ...this.homeTaskArchivedList]
      const newCountMap = { ...this.homeTaskWorkflowCountMap }
      const newUnreadMap = { ...this.homeTaskWorkflowUnreadMap }
      for (const task of allTasks) {
        const taskId = Number(task.id)
        if (Number(task.use_workflow) === HOME_TASK_USE_WORKFLOW_NO) continue
        const raw = String(nodeStatusesMap[String(taskId)] || '').trim()
        let nodeStatuses = {}
        if (raw) {
          try { nodeStatuses = JSON.parse(raw) } catch (e) { /* ignore */ }
        }
        let completed = 0
        let skipped = 0
        for (const key of HOME_TASK_WORKFLOW_NODE_KEYS) {
          const status = nodeStatuses[key] || 'pending'
          if (status === 'skipped') skipped++
          else if (status === 'completed') completed++
        }
        const total = HOME_TASK_WORKFLOW_NODE_KEYS.length
        const nonSkipped = total - skipped
        newCountMap[taskId] = completed + '/' + nonSkipped
        newUnreadMap[taskId] = Math.max(0, Number(unreadCountMap[String(taskId)] || unreadCountMap[taskId] || 0))
      }
      this.homeTaskWorkflowCountMap = newCountMap
      this.homeTaskWorkflowUnreadMap = newUnreadMap
    },
    triggerHomeTaskPageDataLoad() {
      this.ensureHomeTaskPageDataSse().then((ready) => {
        if (!ready) return
        const clientId = sseDistribute.GetSseClientId()
        if (!clientId) return
        const taskIds = this.homeTaskActiveList.map(t => Number(t.id))
        homeTaskApi.HomeTaskPageDataLoad(clientId, taskIds, (response) => {
          if (!(response && response.ErrCode === 0)) {
            console.warn('[HomeTaskSSE] 附加数据加载触发失败:', response?.ErrMsg)
          }
        })
      })
    },
    triggerLocalDirCheck(taskList) {
      this.ensureHomeTaskPageDataSse().then((ready) => {
        if (!ready) return
        const clientId = sseDistribute.GetSseClientId()
        if (!clientId) return
        const paths = []
        for (const t of taskList) {
          if (Array.isArray(t.dev_configs)) {
            for (const cfg of t.dev_configs) {
              const dir = String(cfg.local_dir || '').trim()
              if (dir && !paths.includes(dir)) paths.push(dir)
            }
          }
        }
        if (paths.length === 0) return
        homeTaskApi.HomeTaskPageDataDirCheck(clientId, paths, (response) => {
          if (!(response && response.ErrCode === 0)) {
            console.warn('[HomeTaskSSE] 目录检查触发失败:', response?.ErrMsg)
          }
        })
      })
    },
    triggerBranchStatusCheck(taskList) {
      this.ensureHomeTaskPageDataSse().then((ready) => {
        if (!ready) return
        const clientId = sseDistribute.GetSseClientId()
        if (!clientId) return
        const items = []
        const seen = new Set()
        for (const t of taskList) {
          if (!Array.isArray(t.dev_configs)) continue
          for (const cfg of t.dev_configs) {
            const dir = String(cfg.local_dir || '').trim()
            const branch = String(cfg.branch_name || '').trim()
            if (!dir || !branch) continue
            const key = dir + '|' + branch
            if (seen.has(key)) continue
            seen.add(key)
            items.push({ local_dir: dir, branch_name: branch })
          }
        }
        if (items.length === 0) return
        homeTaskApi.HomeTaskPageDataBranchCheck(clientId, items, (response) => {
          if (!(response && response.ErrCode === 0)) {
            console.warn('[HomeTaskSSE] 分支检查触发失败:', response?.ErrMsg)
          }
        })
      })
    },
    handleStatusClick(statusKey) {
      if (this.selectedStatus === statusKey) return
      this.selectedStatus = statusKey
      // 切换到归档时，首次懒加载归档数据
      if (statusKey === HOME_TASK_SELECTED_ARCHIVED && !this.homeTaskArchivedLoaded) {
        this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES, 1, this.homeTaskArchivedPageSize)
      }
    },
    loadHomeTaskList(isArchived, page = 0, pageSize = 0) {
      const isAllMode = isArchived === HOME_TASK_ARCHIVED_ALL
      const isArchivedTab = isArchived === HOME_TASK_ARCHIVED_YES
      const isPagination = page > 0 && pageSize > 0

      // 归档首页显示全页 loading，翻页时显示底部加载提示
      const isArchivedFirstPage = isArchivedTab && (!isPagination || page <= 1)
      if (isArchivedFirstPage) {
        this.homeTaskLoadingArchived = true
      } else if (isArchived === HOME_TASK_ARCHIVED_NO) {
        this.homeTaskLoadingActive = true
      } else {
        this.homeTaskLoadingActive = true
        this.homeTaskLoadingArchived = true
      }
      homeTaskApi.HomeTaskList(isArchived, (response) => {
        if (isArchivedFirstPage) {
          this.homeTaskLoadingArchived = false
        } else if (isArchived === HOME_TASK_ARCHIVED_NO) {
          this.homeTaskLoadingActive = false
        } else {
          this.homeTaskLoadingActive = false
          this.homeTaskLoadingArchived = false
        }
        if (!(response && response.ErrCode === 0)) {
          this.homeTaskLoadingActive = false
          this.homeTaskLoadingArchived = false
          this.homeTaskArchivedLoadingMore = false
          this.$helperNotify.error(response?.ErrMsg || '任务列表加载失败')
          return
        }
        const taskList = (Array.isArray(response.Data?.task_list) ? response.Data.task_list : []).map(t => ({
          ...t,
          git_ids: safeParseJSON(t.git_ids, []),
          api_dev_entries: safeParseJSON(t.api_dev_entries, []),
          dev_configs: safeParseJSON(t.dev_configs, []),
        }))
        if (isAllMode) {
          this.homeTaskActiveList = taskList.filter(t => t.is_archived === HOME_TASK_ARCHIVED_NO)
          this.homeTaskArchivedList = taskList.filter(t => t.is_archived === HOME_TASK_ARCHIVED_YES)
          this.loadHomeTaskWorkflowCounts(this.homeTaskActiveList)
          this.loadHomeTaskWorkflowCounts(this.homeTaskArchivedList)
        } else if (isArchivedTab) {
          // 归档：分页时首页替换、后续页追加；非分页直接替换
          if (isPagination && page > 1) {
            this.homeTaskArchivedList = [...this.homeTaskArchivedList, ...taskList]
          } else {
            this.homeTaskArchivedList = taskList
          }
          this.homeTaskArchivedTotal = isPagination ? (response.Data?.total || taskList.length) : taskList.length
          this.homeTaskArchivedPage = isPagination ? (response.Data?.page || page) : 1
          this.homeTaskArchivedLoaded = true
          this.homeTaskArchivedLoadingMore = false
          // 判断是否已加载全部
          this.homeTaskArchivedNoMore = this.homeTaskArchivedList.length >= this.homeTaskArchivedTotal
          this.loadHomeTaskWorkflowCounts(taskList)
        } else {
          // 活跃任务：全量加载（任务数少，无需分页）
          this.homeTaskActiveList = taskList
        }

        // 从列表接口中直接获取 Git 仓库列表和 API 集合列表
        this.populateGitRepoListFromResponse(response.Data)
        this.populateApiCollectionListFromResponse(response.Data)

        // 对当前页任务列表触发 SSE 异步检查：本地目录是否存在、分支是否匹配
        this.triggerLocalDirCheck(taskList)
        this.triggerBranchStatusCheck(taskList)
      }, page, pageSize)
    },
    refreshAllHomeTaskList() {
      // 刷新时重置归档懒加载状态，并重新拉取数量和活跃列表
      this.homeTaskArchivedLoaded = false
      this.homeTaskArchivedPage = 1
      this.homeTaskArchivedNoMore = false
      this.homeTaskArchivedLoadingMore = false
      this.loadHomeTaskCounts()
      this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
      // 如果当前正在查看归档，也需要刷新归档列表
      if (this.selectedStatus === HOME_TASK_SELECTED_ARCHIVED) {
        this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES, 1, this.homeTaskArchivedPageSize)
      }
    },
    // loadHomeTaskCounts 获取活跃和归档任务数量，用于 tab 标签显示。
    loadHomeTaskCounts() {
      homeTaskApi.HomeTaskCount((response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) return
        this.homeTaskArchivedTotal = response.Data.archived_count || 0
      })
    },
    // handleArchivedScroll 归档列表滚动监听，接近底部时自动加载下一页。
    handleArchivedScroll() {
      if (this.homeTaskArchivedLoadingMore || this.homeTaskArchivedNoMore) return
      const el = this.$refs.archivedListRef
      if (!el) return
      const distanceToBottom = el.scrollHeight - el.clientHeight - el.scrollTop
      if (distanceToBottom < 80) {
        this.homeTaskArchivedLoadingMore = true
        const nextPage = this.homeTaskArchivedPage + 1
        this.homeTaskArchivedPage = nextPage
        this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES, nextPage, this.homeTaskArchivedPageSize)
      }
    },
    resetHomeTaskForm() {
      this.homeTaskForm = createHomeTaskDefaultForm()
    },
    populateGitRepoListFromResponse(data) {
      if (!data) return
      const gitList = Array.isArray(data.git_list) ? data.git_list : []
      const groupList = Array.isArray(data.git_group_list) ? data.git_group_list : []
      if (gitList.length === 0) return
      const groupMap = {}
      for (const g of groupList) {
        groupMap[Number(g.id)] = g.name
      }
      this.homeTaskGitRepoList = gitList.map(repo => ({
        ...repo,
        git_group_name: groupMap[Number(repo.git_group_id)] || '未分组',
      }))
      this.homeTaskGitRepoLoading = false
    },
    populateApiCollectionListFromResponse(data) {
      if (!data) return
      const list = Array.isArray(data.api_collection_list) ? data.api_collection_list : []
      if (list.length === 0) return
      this.homeTaskApiCollectionList = list
      this.homeTaskApiCollectionLoading = false
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
    loadHomeTaskDockerList() {
      this.homeTaskDockerLoading = true
      dockerApi.DockerComposeList({}, (response) => {
        this.homeTaskDockerLoading = false
        if (!(response && response.ErrCode === 0)) {
          return
        }
        this.homeTaskDockerList = Array.isArray(response.Data?.list) ? response.Data.list : []
      })
    },
    loadHomeTaskApiFoldersForCollection(collectionId) {
      if (!collectionId) return
      if (this.homeTaskApiFolderMap[collectionId]) return
      if (this.homeTaskApiFolderLoadingMap && this.homeTaskApiFolderLoadingMap[collectionId]) return
      if (!this.homeTaskApiFolderLoadingMap) this.homeTaskApiFolderLoadingMap = {}
      this.homeTaskApiFolderLoadingMap[collectionId] = true
      apiManagement.CollectionFoldersBasic({ collection_id: collectionId }, (response) => {
        this.homeTaskApiFolderLoadingMap[collectionId] = false
        if (!(response && response.ErrCode === 0)) return
        const list = Array.isArray(response.Data?.list) ? response.Data.list : []
        this.homeTaskApiFolderMap = { ...this.homeTaskApiFolderMap, [collectionId]: list }
      })
    },
    getDevConfigFolders(cfgIdx) {
      const colId = this.homeTaskForm.dev_configs[cfgIdx]?.collection_id
      if (!colId) return []
      return this.homeTaskApiFolderMap[colId] || []
    },
    handleDevConfigCollectionChange(cfgIdx) {
      const cfg = this.homeTaskForm.dev_configs[cfgIdx]
      cfg.dir_id = 0
      this.loadHomeTaskApiFoldersForCollection(cfg.collection_id)
    },
    loadHomeTaskSmartLinkList() {
      this.homeTaskSmartLinkLoading = true
      smartLinkSetApi.SmartLinkItemList((response) => {
        this.homeTaskSmartLinkLoading = false
        if (!(response && response.ErrCode === 0)) {
          return
        }
        const rawList = Array.isArray(response.Data?.smart_link_list) ? response.Data.smart_link_list : []
        this.homeTaskSmartLinkList = rawList
      })
    },
    loadHomeTaskMemoryFolderList() {
      this.homeTaskMemoryFolderLoading = true
      memoryFragmentApi.MemoryFragmentFolderList((response) => {
        this.homeTaskMemoryFolderLoading = false
        if (!(response && response.ErrCode === 0)) {
          return
        }
        this.homeTaskMemoryFolderList = Array.isArray(response.Data) ? response.Data : []
        const exists = this.homeTaskMemoryFolderList.some(item => item.folder_name === this.homeTaskForm.workflow_fragment_folder_name)
        if (!exists) {
          this.homeTaskForm.workflow_fragment_folder_name = HOME_TASK_DEFAULT_WORKFLOW_FRAGMENT_FOLDER
        }
      })
    },
    loadHomeTaskTemplateList() {
      this.homeTaskTemplateLoading = true
      workflowTemplateApi.WorkflowTemplateListBasic((response) => {
        this.homeTaskTemplateLoading = false
        if (response && response.ErrCode === 0 && response.Data && Array.isArray(response.Data.list)) {
          this.homeTaskTemplateList = response.Data.list
          // 如果未选择模板且有列表，自动选择默认或第一个
          if (!this.homeTaskForm.workflow_template_id && this.homeTaskTemplateList.length > 0) {
            const defaultTpl = this.homeTaskTemplateList.find(t => t.is_default === 1)
            this.homeTaskForm.workflow_template_id = defaultTpl ? defaultTpl.id : this.homeTaskTemplateList[0].id
          }
        }
      })
    },
    getDevConfigSmartLinkLabels(cfgIdx) {
      const cfg = this.homeTaskForm.dev_configs[cfgIdx]
      if (!cfg || !cfg.smart_link_id) return []
      const smartLink = this.homeTaskSmartLinkList.find(s => Number(s.id) === Number(cfg.smart_link_id))
      if (!smartLink) return []
      return Array.isArray(smartLink.linkList) ? smartLink.linkList : []
    },
    getDevConfigSmartLinkAccounts(cfgIdx) {
      const cfg = this.homeTaskForm.dev_configs[cfgIdx]
      if (!cfg || !cfg.smart_link_label) return []
      const labels = this.getDevConfigSmartLinkLabels(cfgIdx)
      const link = labels.find(l => l.label === cfg.smart_link_label)
      if (!link) return []
      return Array.isArray(link.userList) ? link.userList : []
    },
    handleDevConfigSmartLinkChange(cfgIdx) {
      const cfg = this.homeTaskForm.dev_configs[cfgIdx]
      cfg.smart_link_label = ''
      cfg.smart_link_account = ''
    },
    handleDevConfigSmartLinkLabelChange(cfgIdx) {
      const cfg = this.homeTaskForm.dev_configs[cfgIdx]
      cfg.smart_link_account = ''
    },
    handleDevConfigGitChange(cfgIdx) {
      const cfg = this.homeTaskForm.dev_configs[cfgIdx]
      const gitId = Number(cfg.git_id || 0)
      // 清空 Git 时不填充
      if (gitId <= 0) {
        return
      }
      homeTaskApi.HomeTaskLastDevConfigByGitId(gitId, (response) => {
        if (!(response && response.ErrCode === 0) || !response.Data) {
          return
        }
        const lastCfg = response.Data
        // 仅在新建任务时自动填充，编辑时保留用户已设置的值
        if (this.homeTaskForm.id > 0) {
          return
        }
        cfg.docker_id = Number(lastCfg.docker_id || 0) || ''
        cfg.collection_id = Number(lastCfg.collection_id || 0) || ''
        cfg.mysql_id = Number(lastCfg.mysql_id || 0) || ''
        cfg.local_dir = String(lastCfg.local_dir || '')
        cfg.parent_branch = String(lastCfg.parent_branch || '')
        cfg.rule_entry_file = String(lastCfg.rule_entry_file || '')
        cfg.smart_link_id = Number(lastCfg.smart_link_id || 0) || ''
        cfg.smart_link_label = String(lastCfg.smart_link_label || '')
        cfg.smart_link_account = String(lastCfg.smart_link_account || '')
        // 如果有接口集合，加载对应的文件夹列表
        if (cfg.collection_id > 0) {
          this.loadHomeTaskApiFoldersForCollection(cfg.collection_id)
        }
      })
    },
    generateBranchName(cfgIdx) {
      const cfg = this.homeTaskForm.dev_configs[cfgIdx]
      const taskName = String(this.homeTaskForm.name || '').trim()
      if (!taskName) {
        this.$helperNotify.error('请先填写任务名称')
        return
      }
      const branchName = String(cfg.branch_name || '').trim()
      if (branchName) {
        this.$confirm('当前分支名不为空，重新生成将覆盖已有内容，是否继续？', '提示', {
          confirmButtonText: '确定',
          cancelButtonText: '取消',
          type: 'warning',
        }).then(() => {
          this._doGenerateBranchName(cfgIdx)
        }).catch(() => {})
        return
      }
      this._doGenerateBranchName(cfgIdx)
    },
    _doGenerateBranchName(cfgIdx) {
      const cfg = this.homeTaskForm.dev_configs[cfgIdx]
      const taskName = String(this.homeTaskForm.name || '').trim()
      const createdDate = String(this.homeTaskForm.created_date || '').trim()
      cfg._branchGenerating = true
      homeTaskApi.HomeTaskBranchNameGenerate(taskName, String(cfg.parent_branch || '').trim(), createdDate, (response) => {
        cfg._branchGenerating = false
        if (response && response.ErrCode === 0 && response.Data) {
          cfg.branch_name = response.Data.branch_name || ''
          this.$helperNotify.success('分支名已生成')
        } else {
          this.$helperNotify.error((response && response.ErrMsg) || '生成分支名失败')
        }
      })
    },

    addDevConfig() {
      this.homeTaskForm.dev_configs.push({ git_id: '', collection_id: '', dir_id: '', docker_id: '', local_dir: '', parent_branch: '', branch_name: '', rule_entry_file: '', _branchGenerating: false, smart_link_id: '', smart_link_label: '', smart_link_account: '' })
    },
    removeDevConfig(idx) {
      this.homeTaskForm.dev_configs.splice(idx, 1)
      if (this.homeTaskForm.dev_configs.length === 0) {
        this.addDevConfig()
      }
    },
    openCreateHomeTaskDialog() {
      this.resetHomeTaskForm()
      this.loadHomeTaskMysqlList()
      this.loadHomeTaskDockerList()
      this.loadHomeTaskSmartLinkList()
      this.loadHomeTaskMemoryFolderList()
      this.loadHomeTaskUnusedLocalDirs(0)
      this.loadHomeTaskTemplateList()
      this.loadRequirementFetchConfigs()
      this.homeTaskDialogVisible = true
    },
    // 加载需求抓取配置列表（用于 fetch_type 下拉选项）
    loadRequirementFetchConfigs() {
      set.HomeTaskConfigGet((response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          const configs = response.Data.home_task_requirement_fetch_configs
          this.requirementFetchConfigs = Array.isArray(configs) ? [...configs] : []
        }
      })
    },
    // 根据 fetch_type 查找配置
    getFetchConfigByType(type) {
      if (!type) return null
      return this.requirementFetchConfigs.find(c => c.type === type) || null
    },
    openHomeTaskSettingsPage() {
      const routeInfo = this.$router.resolve({ path: '/HomeTaskSetting' })
      window.open(routeInfo.href, '_blank')
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
      homeTaskApi.HomeTaskInfo(task.id, (response) => {
        if (!response || response.ErrCode !== 0) {
          this.$helperNotify.error(response?.ErrMsg || '任务详情查询失败')
          return
        }
        const detail = response.Data || {}
        detail.git_ids = safeParseJSON(detail.git_ids, [])
        detail.api_dev_entries = safeParseJSON(detail.api_dev_entries, [])
        detail.dev_configs = safeParseJSON(detail.dev_configs, [])
        this.fillHomeTaskEditForm(detail)
      })
    },
    fillHomeTaskEditForm(task) {
      let devConfigs = []
      if (Array.isArray(task.dev_configs) && task.dev_configs.length > 0) {
        devConfigs = task.dev_configs.map(cfg => ({
          git_id: Number(cfg.git_id || 0) || '',
          collection_id: Number(cfg.collection_id || 0) || '',
          dir_id: Number(cfg.dir_id || 0) || '',
          docker_id: Number(cfg.docker_id || 0) || '',
          mysql_id: Number(cfg.mysql_id || 0) || '',
          local_dir: String(cfg.local_dir || ''),
          parent_branch: String(cfg.parent_branch || ''),
          branch_name: String(cfg.branch_name || ''),
          rule_entry_file: String(cfg.rule_entry_file || ''),
          smart_link_id: Number(cfg.smart_link_id || 0) || '',
          smart_link_label: String(cfg.smart_link_label || ''),
          smart_link_account: String(cfg.smart_link_account || ''),
        }))
      } else {
        let gitIds = Array.isArray(task.git_ids) && task.git_ids.length > 0
          ? task.git_ids.map(id => Number(id))
          : (Number(task.git_id || 0) > 0 ? [Number(task.git_id)] : [])
        let apiEntries = Array.isArray(task.api_dev_entries) && task.api_dev_entries.length > 0
          ? task.api_dev_entries
          : (Number(task.api_collection_id || 0) > 0
            ? [{ collection_id: Number(task.api_collection_id), dir_id: Number(task.api_dir_id || 0) }]
            : [])
        const maxLen = Math.max(gitIds.length, apiEntries.length, 1)
        for (let i = 0; i < maxLen; i++) {
          devConfigs.push({
            git_id: gitIds[i] || '',
            collection_id: Number(apiEntries[i]?.collection_id || 0) || '',
            dir_id: Number(apiEntries[i]?.dir_id || 0) || '',
            docker_id: '',
            mysql_id: Number(task.mysql_id || 0) || '',
            local_dir: '',
            parent_branch: '',
            branch_name: '',
            rule_entry_file: '',
            smart_link_id: '',
            smart_link_label: '',
            smart_link_account: '',
          })
        }
      }
      if (devConfigs.length === 0) {
        devConfigs = [{ git_id: '', collection_id: '', dir_id: '', docker_id: '', local_dir: '', parent_branch: '', branch_name: '', rule_entry_file: '', _branchGenerating: false, smart_link_id: '', smart_link_label: '', smart_link_account: '' }]
      }
      this.homeTaskForm = {
        id: Number(task.id || 0),
        name: task.name || '',
        task_status: task.task_status || HOME_TASK_STATUS_TODO,
        start_date: task.start_time_desc || getTodayDateText(),
        created_date: (task.create_time_desc || '').split(' ')[0] || '',
        fetch_type: String(task.fetch_type || HOME_TASK_FETCH_TYPE_TAPD),
        requirement_url: String(task.fetch_type || '') === HOME_TASK_FETCH_TYPE_ZENTAO ? (task.zentao_url || '') : (task.tapd_url || ''),
        use_workflow: Number(task.use_workflow ?? HOME_TASK_USE_WORKFLOW_YES) === HOME_TASK_USE_WORKFLOW_YES ? HOME_TASK_USE_WORKFLOW_YES : HOME_TASK_USE_WORKFLOW_NO,
        workflow_fragment_folder_name: String(task.workflow_fragment_folder_name || HOME_TASK_DEFAULT_WORKFLOW_FRAGMENT_FOLDER).trim() || HOME_TASK_DEFAULT_WORKFLOW_FRAGMENT_FOLDER,
        workflow_template_id: Number(task.workflow_template_id || 0),
        dev_configs: devConfigs,
      }
      this.loadRequirementFetchConfigs()
      this.loadHomeTaskMysqlList()
      this.loadHomeTaskDockerList()
      this.loadHomeTaskSmartLinkList()
      this.loadHomeTaskMemoryFolderList()
      const devColIds = new Set()
      for (const cfg of devConfigs) {
        if (cfg.collection_id > 0) devColIds.add(cfg.collection_id)
      }
      for (const colId of devColIds) {
        this.loadHomeTaskApiFoldersForCollection(colId)
      }
      this.loadHomeTaskUnusedLocalDirs(Number(task.id || 0))
      this.loadHomeTaskTemplateList()
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
    getHomeTaskDevConfigTags(task) {
      const DEV_CONFIG_TAG_TYPE_GIT = 'success'
      const DEV_CONFIG_TAG_TYPE_API = 'info'
      const DEV_CONFIG_TAG_TYPE_DOCKER = 'info'
      const DEV_CONFIG_TAG_TYPE_DB = 'warning'
      const DEV_CONFIG_TAG_TYPE_DIR = 'danger'
      let configs = []
      if (Array.isArray(task.dev_configs) && task.dev_configs.length > 0) {
        configs = task.dev_configs
      }
      if (configs.length === 0) return []
      const groups = []
      for (const cfg of configs) {
        const group = []
        if (Number(cfg.git_id || 0) > 0) {
          const repo = this.homeTaskGitRepoList.find(r => Number(r.id) === Number(cfg.git_id))
          if (repo) {
            group.push({ type: 'git', label: repo.name, id: Number(cfg.git_id), tagType: DEV_CONFIG_TAG_TYPE_GIT })
          }
        }
        if (Number(cfg.collection_id || 0) > 0) {
          const col = this.homeTaskApiCollectionList.find(c => Number(c.id) === Number(cfg.collection_id))
          if (col) {
            let label = col.name
            let folderId = 0
            if (Number(cfg.dir_id || 0) > 0) {
              const folders = this.homeTaskApiFolderMap[cfg.collection_id] || []
              const dir = folders.find(d => Number(d.id) === Number(cfg.dir_id))
              if (dir) {
                label += '/' + dir.name
                folderId = Number(cfg.dir_id)
              }
            }
            group.push({
              type: 'api',
              label: label,
              collectionId: Number(cfg.collection_id),
              folderId: folderId,
              tagType: DEV_CONFIG_TAG_TYPE_API,
            })
          }
        }
        if (Number(cfg.docker_id || 0) > 0) {
          const docker = this.homeTaskDockerList.find(d => Number(d.id) === Number(cfg.docker_id))
          if (docker) {
            group.push({ type: 'docker', label: 'Docker: ' + docker.name, id: Number(cfg.docker_id), tagType: DEV_CONFIG_TAG_TYPE_DOCKER })
          }
        }
        if (Number(cfg.mysql_id || 0) > 0) {
          const mysql = this.homeTaskMysqlList.find(m => Number(m.id) === Number(cfg.mysql_id))
          if (mysql) {
            group.push({ type: 'mysql', label: 'Db: ' + mysql.name, id: Number(cfg.mysql_id), tagType: DEV_CONFIG_TAG_TYPE_DB })
          }
        }
        if (String(cfg.local_dir || '').trim() !== '') {
          const dirPath = String(cfg.local_dir).trim()
          group.push({ type: 'local_dir', label: dirPath, fullPath: dirPath, tagType: DEV_CONFIG_TAG_TYPE_DIR })
        }
        if (String(cfg.parent_branch || '').trim() !== '') {
          group.push({ type: 'parent_branch', label: '分支: ' + String(cfg.parent_branch).trim(), tagType: 'info' })
        }
        if (String(cfg.branch_name || '').trim() !== '') {
          group.push({ type: 'branch_name', label: String(cfg.branch_name).trim(), tagType: 'success', localDir: String(cfg.local_dir || '').trim() })
        }
        if (Number(cfg.smart_link_id || 0) > 0) {
          const sl = this.homeTaskSmartLinkList.find(s => Number(s.id) === Number(cfg.smart_link_id))
          if (sl) {
            group.push({ type: 'smart_link', label: sl.name, id: Number(cfg.smart_link_id), tagType: 'info' })
          }
        }
        if (String(cfg.smart_link_label || '').trim() !== '') {
          group.push({ type: 'smart_link_label', label: String(cfg.smart_link_label).trim(), tagType: 'info' })
        }
        if (String(cfg.smart_link_account || '').trim() !== '') {
          group.push({ type: 'smart_link_account', label: String(cfg.smart_link_account).trim(), tagType: 'warning' })
        }
        if (group.length > 0) {
          groups.push(group)
        }
      }
      return groups
    },
    navigateToDevConfig(tag) {
      if (tag.type === 'local_dir') {
        homeTaskApi.OpenLocalDir(tag.fullPath, (response) => {
          if (!(response && response.ErrCode === 0)) {
            this.$helperNotify.error(response?.ErrMsg || '打开目录失败')
          }
        })
        return
      }
      let path = ''
      if (tag.type === 'git') {
        path = '/Git'
      } else if (tag.type === 'api') {
        const query = {}
        if (tag.collectionId > 0) {
          query.collection_id = String(tag.collectionId)
        }
        if (tag.folderId > 0) {
          query.folder_id = String(tag.folderId)
        }
        const routeInfo = this.$router.resolve({ path: '/Api', query })
        window.open(routeInfo.href, '_blank')
        return
      } else if (tag.type === 'docker') {
        path = '/Docker'
      } else if (tag.type === 'mysql') {
        path = '/Set'
      } else if (tag.type === 'smart_link') {
        path = '/Link'
      }
      if (!path) return
      const routeInfo = this.$router.resolve({ path })
      window.open(routeInfo.href, '_blank')
    },
    copyHomeTaskBranchName(branchName) {
      navigator.clipboard.writeText(branchName).then(() => {
        this.$message.success('已复制分支名')
      })
    },
    saveHomeTask() {
      if (this.homeTaskSaving) {
        return
      }
      const taskName = String(this.homeTaskForm.name || '').trim()
      const fetchType = String(this.homeTaskForm.fetch_type || HOME_TASK_FETCH_TYPE_TAPD).trim()
      const requirementUrl = String(this.homeTaskForm.requirement_url || '').trim()
      // 兼容后端：tapd 和 zentao 仍使用独立字段；自定义类型统一走 tapd_url
      const tapdUrl = (fetchType === HOME_TASK_FETCH_TYPE_ZENTAO) ? '' : requirementUrl
      const zentaoUrl = (fetchType === HOME_TASK_FETCH_TYPE_ZENTAO) ? requirementUrl : ''
      if (!taskName) {
        this.$helperNotify.error('任务名称不能为空')
        return
      }
      if (taskName.length > 200) {
        this.$helperNotify.error('任务名称不能超过200字')
        return
      }
      const workflowFragmentFolderName = String(this.homeTaskForm.workflow_fragment_folder_name || '').trim()
      if (Number(this.homeTaskForm.use_workflow) === HOME_TASK_USE_WORKFLOW_YES && !workflowFragmentFolderName) {
        this.$helperNotify.error('知识片段文件夹不能为空')
        return
      }
      const validConfigs = this.homeTaskForm.dev_configs
        .filter(cfg => Number(cfg.git_id || 0) > 0 || Number(cfg.collection_id || 0) > 0 || Number(cfg.docker_id || 0) > 0 || Number(cfg.mysql_id || 0) > 0 || String(cfg.local_dir || '').trim() !== '' || String(cfg.parent_branch || '').trim() !== '' || String(cfg.branch_name || '').trim() !== '' || String(cfg.rule_entry_file || '').trim() !== '' || Number(cfg.smart_link_id || 0) > 0)
        .map(cfg => ({
          git_id: Number(cfg.git_id || 0),
          collection_id: Number(cfg.collection_id || 0),
          dir_id: Number(cfg.dir_id || 0),
          docker_id: Number(cfg.docker_id || 0),
          mysql_id: Number(cfg.mysql_id || 0),
          local_dir: String(cfg.local_dir || '').trim(),
          parent_branch: String(cfg.parent_branch || '').trim(),
          branch_name: String(cfg.branch_name || '').trim(),
          rule_entry_file: String(cfg.rule_entry_file || '').trim(),
          smart_link_id: Number(cfg.smart_link_id || 0),
          smart_link_label: String(cfg.smart_link_label || '').trim(),
          smart_link_account: String(cfg.smart_link_account || '').trim(),
        }))
      const isEdit = this.homeTaskForm.id > 0
      const useWorkflow = this.homeTaskForm.use_workflow
      this.homeTaskSaving = true
      this.homeTaskOperatingType = HOME_TASK_OPERATE_SAVE
      homeTaskApi.HomeTaskSave({
        id: Number(this.homeTaskForm.id || 0),
        name: taskName,
        task_status: this.homeTaskForm.task_status,
        start_time: this.convertHomeTaskDateToUnix(this.homeTaskForm.start_date),
        fetch_type: fetchType,
        tapd_url: tapdUrl,
        zentao_url: zentaoUrl,
        dev_configs: JSON.stringify(validConfigs),
        use_workflow: useWorkflow,
        workflow_fragment_folder_name: workflowFragmentFolderName,
        workflow_template_id: Number(this.homeTaskForm.workflow_template_id || 0),
        api_host: base.GetApiHost() || window.location.origin,
        api_token: base.GetSafeToken(),
      }, (response) => {
        this.homeTaskSaving = false
        this.homeTaskOperatingType = ''
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '任务保存失败')
          return
        }
        const createdTaskId = Number(response?.Data?.id || 0)
        this.$helperNotify.success(isEdit ? '任务已更新' : '任务已创建')
        this.closeHomeTaskDialog()
        if (isEdit) {
          const taskId = Number(this.homeTaskForm.id)
          this.triggerHomeTaskEditFeedback(taskId)
        }
        this.refreshAllHomeTaskList()
        if (!isEdit && createdTaskId > 0 && useWorkflow === HOME_TASK_USE_WORKFLOW_YES && requirementUrl) {
          this.openTaskWorkflow({ id: createdTaskId })
        }
      })
    },
    getHomeTaskRequirementUrl(task) {
      const fetchType = String(task?.fetch_type || HOME_TASK_FETCH_TYPE_TAPD).toLowerCase()
      if (fetchType === HOME_TASK_FETCH_TYPE_ZENTAO) {
        return String(task?.zentao_url || '').trim()
      }
      return String(task?.tapd_url || '').trim()
    },
    getHomeTaskRequirementLabel(task) {
      const fetchType = String(task?.fetch_type || HOME_TASK_FETCH_TYPE_TAPD)
      const cfg = this.getFetchConfigByType(fetchType)
      if (cfg) return cfg.name + '需求'
      return fetchType === HOME_TASK_FETCH_TYPE_ZENTAO ? '禅道需求' : 'TAPD需求'
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
      const parsedTask = {
        ...updatedTask,
        git_ids: safeParseJSON(updatedTask.git_ids, []),
        api_dev_entries: safeParseJSON(updatedTask.api_dev_entries, []),
        dev_configs: safeParseJSON(updatedTask.dev_configs, []),
      }
      const activeIndex = this.homeTaskActiveList.findIndex(t => Number(t.id) === taskId)
      if (activeIndex >= 0) {
        this.homeTaskActiveList[activeIndex] = { ...this.homeTaskActiveList[activeIndex], ...parsedTask }
        return
      }
      const archivedIndex = this.homeTaskArchivedList.findIndex(t => Number(t.id) === taskId)
      if (archivedIndex >= 0) {
        this.homeTaskArchivedList[archivedIndex] = { ...this.homeTaskArchivedList[archivedIndex], ...parsedTask }
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
      if (taskStatus === HOME_TASK_STATUS_DEVELOPING || taskStatus === HOME_TASK_STATUS_DEV_COMPLETED) {
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
      if (taskStatus === HOME_TASK_STATUS_PENDING_TEST) {
        return 'info'
      }
      if (taskStatus === HOME_TASK_STATUS_ABANDONED) {
        return 'danger'
      }
      return ''
    },
    getHomeTaskActionButtonVariant(taskStatus) {
      return 'warning'
    },
    loadHomeTaskUnusedLocalDirs(excludeTaskId) {
      homeTaskApi.HomeTaskUnusedLocalDirs(excludeTaskId || 0, (response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.homeTaskUnusedLocalDirs = Array.isArray(response.Data.dirs) ? response.Data.dirs : []
        } else {
          this.homeTaskUnusedLocalDirs = []
        }
      })
    },
    // 批量检查任务列表中的本地目录是否存在
    checkLocalDirExists(taskList) {
      const paths = []
      for (const t of taskList) {
        if (Array.isArray(t.dev_configs)) {
          for (const cfg of t.dev_configs) {
            const dir = String(cfg.local_dir || '').trim()
            if (dir && !paths.includes(dir)) {
              paths.push(dir)
            }
          }
        }
      }
      if (paths.length === 0) return
      homeTaskApi.LocalDirBatchCheck(paths, (response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.homeTaskLocalDirStatusMap = { ...this.homeTaskLocalDirStatusMap, ...response.Data }
        }
      })
    },
    // 批量检查任务列表中本地目录的当前 Git 分支是否与配置的分支名匹配
    checkBranchStatus(taskList) {
      const items = []
      const seen = new Set()
      for (const t of taskList) {
        if (!Array.isArray(t.dev_configs)) continue
        for (const cfg of t.dev_configs) {
          const dir = String(cfg.local_dir || '').trim()
          const branch = String(cfg.branch_name || '').trim()
          if (!dir || !branch) continue
          const key = dir + '|' + branch
          if (seen.has(key)) continue
          seen.add(key)
          items.push({ local_dir: dir, branch_name: branch })
        }
      }
      if (items.length === 0) return
      homeTaskApi.LocalBranchBatchCheck(items, (response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.homeTaskBranchStatusMap = { ...this.homeTaskBranchStatusMap, ...response.Data }
        }
      })
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
    loadHomeTaskWorkflowCounts(taskList) {
      const workflowTaskIds = []
      const nextCountMap = { ...this.homeTaskWorkflowCountMap }
      const nextUnreadMap = { ...this.homeTaskWorkflowUnreadMap }
      for (const task of taskList) {
        const taskId = Number(task.id || 0)
        delete nextCountMap[taskId]
        delete nextUnreadMap[taskId]
        if (Number(task.use_workflow) !== HOME_TASK_USE_WORKFLOW_NO) {
          workflowTaskIds.push(taskId)
        }
      }
      this.homeTaskWorkflowCountMap = nextCountMap
      this.homeTaskWorkflowUnreadMap = nextUnreadMap
      if (workflowTaskIds.length === 0) return
      taskWorkflowApi.TaskWorkflowBatchNodeStatus(workflowTaskIds, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) return
        const nodeStatusesMap = response.Data.node_statuses_map || {}
        const unreadCountMap = response.Data.unread_count_map || {}
        const newMap = { ...this.homeTaskWorkflowCountMap }
        const newUnreadMap = { ...this.homeTaskWorkflowUnreadMap }
        for (const task of taskList) {
          const taskId = Number(task.id)
          if (Number(task.use_workflow) === HOME_TASK_USE_WORKFLOW_NO) continue
          const raw = String(nodeStatusesMap[String(taskId)] || '').trim()
          let nodeStatuses = {}
          if (raw) {
            try {
              nodeStatuses = JSON.parse(raw)
            } catch (e) { /* ignore */ }
          }
          let completed = 0
          let skipped = 0
          for (const key of HOME_TASK_WORKFLOW_NODE_KEYS) {
            const status = nodeStatuses[key] || 'pending'
            if (status === 'skipped') {
              skipped++
            } else if (status === 'completed') {
              completed++
            }
          }
          const total = HOME_TASK_WORKFLOW_NODE_KEYS.length
          const nonSkipped = total - skipped
          newMap[taskId] = completed + '/' + nonSkipped
          newUnreadMap[taskId] = Math.max(0, Number(unreadCountMap[String(taskId)] || unreadCountMap[taskId] || 0))
        }
        this.homeTaskWorkflowCountMap = newMap
        this.homeTaskWorkflowUnreadMap = newUnreadMap
      })
    },
    hasHomeTaskWorkflowUnread(task) {
      const taskId = Number(task?.id || 0)
      return Number(this.homeTaskWorkflowUnreadMap[taskId] || 0) > 0
    },
    shouldTruncateTagLabel(tag) {
      // 接口集合(api)和分支名(branch_name)需要截断显示
      if (tag.type === 'api' || tag.type === 'branch_name') {
        return tag.label.length > HOME_TASK_CONFIG_TAG_MAX_LENGTH
      }
      // 其他类型保持原有逻辑：除 local_dir, parent_branch 外截断
      if (tag.type === 'local_dir' || tag.type === 'parent_branch') {
        return false
      }
      return tag.label.length > HOME_TASK_CONFIG_TAG_MAX_LENGTH
    },
    truncateTagLabel(label) {
      return label.slice(0, HOME_TASK_CONFIG_TAG_MAX_LENGTH) + '...'
    },
    getHomeTaskWorkflowCountText(task) {
      const taskId = Number(task?.id || 0)
      const display = this.homeTaskWorkflowCountMap[taskId]
      return display || ''
    },
    copyUnusedLocalDir(dir) {
      navigator.clipboard.writeText(dir).then(() => {
        this.$message.success('已复制')
      })
    },
    // ========== 任务状态管理 ==========
    loadTaskStatusList() {
      homeTaskApi.TaskStatusList((response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.taskStatusList = Array.isArray(response.Data.list) ? response.Data.list : []
        }
      })
    },
    openAddTaskStatusDialog() {
      this.taskStatusForm = { id: 0, name: '' }
      this.taskStatusDialogVisible = true
    },
    openEditTaskStatusDialog(status) {
      this.taskStatusForm = { id: status.id, name: status.name }
      this.taskStatusDialogVisible = true
    },
    saveTaskStatus() {
      const name = (this.taskStatusForm.name || '').trim()
      if (!name) {
        this.$message.warning('状态名称不能为空')
        return
      }
      const sortOrder = this.taskStatusForm.id > 0
        ? undefined
        : this.taskStatusList.length
      this.taskStatusSaving = true
      homeTaskApi.TaskStatusSave({
        id: this.taskStatusForm.id,
        name: name,
        sort_order: sortOrder,
      }, (response) => {
        this.taskStatusSaving = false
        if (response && response.ErrCode === 0 && response.Data) {
          this.taskStatusList = Array.isArray(response.Data.list) ? response.Data.list : []
          this.taskStatusDialogVisible = false
          this.$message.success(this.taskStatusForm.id > 0 ? '状态已更新' : '状态已添加')
        } else {
          this.$message.error(response?.ErrMsg || '保存失败')
        }
      })
    },
    deleteTaskStatus(status) {
      this.$confirm(`确定要删除状态"${status.name}"吗？`, '确认删除', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        homeTaskApi.TaskStatusDelete(status.id, (response) => {
          if (response && response.ErrCode === 0 && response.Data) {
            this.taskStatusList = Array.isArray(response.Data.list) ? response.Data.list : []
            this.$message.success('状态已删除')
          } else {
            this.$message.error(response?.ErrMsg || '删除失败')
          }
        })
      }).catch(() => {})
    },
    moveTaskStatusUp(idx) {
      if (idx <= 0) return
      const ids = this.taskStatusList.map(s => s.id)
      const tmp = ids[idx]
      ids[idx] = ids[idx - 1]
      ids[idx - 1] = tmp
      this.saveTaskStatusSort(ids)
    },
    moveTaskStatusDown(idx) {
      if (idx >= this.taskStatusList.length - 1) return
      const ids = this.taskStatusList.map(s => s.id)
      const tmp = ids[idx]
      ids[idx] = ids[idx + 1]
      ids[idx + 1] = tmp
      this.saveTaskStatusSort(ids)
    },
    saveTaskStatusSort(ids) {
      homeTaskApi.TaskStatusSort(ids, (response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.taskStatusList = Array.isArray(response.Data.list) ? response.Data.list : []
        } else {
          this.$message.error(response?.ErrMsg || '排序失败')
        }
      })
    },
  },
  components: {
    GitActionButton,
  },
}
</script>

<style scoped src="@/css/components/HomeTask.css"></style>
