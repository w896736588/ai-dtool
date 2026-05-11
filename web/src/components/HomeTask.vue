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

    <div class="home-task-content-card">

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
                            :disabled="tag.type === 'branch_name' || tag.type === 'local_dir' || tag.label.length <= HOME_TASK_CONFIG_TAG_MAX_LENGTH"
                          >
                            <span class="home-task-config-tag-wrapper">
                              <el-tag
                                size="small"
                                effect="plain"
                                :type="tag.tagType"
                                :class="['home-task-config-tag', tag.type === 'branch_name' ? 'home-task-config-tag--copy' : '']"
                                @click.stop="tag.type === 'branch_name' ? copyHomeTaskBranchName(tag.label) : navigateToDevConfig(tag)"
                              >
                                {{ (tag.type === 'branch_name' || tag.type === 'local_dir') ? tag.label : (tag.label.length > HOME_TASK_CONFIG_TAG_MAX_LENGTH ? tag.label.slice(0, HOME_TASK_CONFIG_TAG_MAX_LENGTH) + '...' : tag.label) }}
                              </el-tag>
                              <span v-if="tag.type === 'local_dir' && homeTaskLocalDirStatusMap[tag.label] !== undefined" class="home-task-dir-status" :class="homeTaskLocalDirStatusMap[tag.label] ? 'home-task-dir-status--ok' : 'home-task-dir-status--err'">
                                <svg v-if="homeTaskLocalDirStatusMap[tag.label]" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg>
                                <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                              </span>
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
                    :disabled="isHomeTaskBusy(task.id)"
                    v-if="Number(task.use_workflow) !== HOME_TASK_USE_WORKFLOW_NO"
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
                            :disabled="tag.type === 'branch_name' || tag.type === 'local_dir' || tag.label.length <= HOME_TASK_CONFIG_TAG_MAX_LENGTH"
                          >
                            <span class="home-task-config-tag-wrapper">
                              <el-tag
                                size="small"
                                effect="plain"
                                :type="tag.tagType"
                                :class="['home-task-config-tag', tag.type === 'branch_name' ? 'home-task-config-tag--copy' : '']"
                                @click.stop="tag.type === 'branch_name' ? copyHomeTaskBranchName(tag.label) : navigateToDevConfig(tag)"
                              >
                                {{ (tag.type === 'branch_name' || tag.type === 'local_dir') ? tag.label : (tag.label.length > HOME_TASK_CONFIG_TAG_MAX_LENGTH ? tag.label.slice(0, HOME_TASK_CONFIG_TAG_MAX_LENGTH) + '...' : tag.label) }}
                              </el-tag>
                              <span v-if="tag.type === 'local_dir' && homeTaskLocalDirStatusMap[tag.label] !== undefined" class="home-task-dir-status" :class="homeTaskLocalDirStatusMap[tag.label] ? 'home-task-dir-status--ok' : 'home-task-dir-status--err'">
                                <svg v-if="homeTaskLocalDirStatusMap[tag.label]" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg>
                                <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                              </span>
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
                    :disabled="isHomeTaskBusy(task.id)"
                    v-if="Number(task.use_workflow) !== HOME_TASK_USE_WORKFLOW_NO"
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
          </div>
        </el-tab-pane>
      </el-tabs>
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
                maxlength="80"
                show-word-limit
                placeholder="例如：整理缓存淘汰策略"
                @keyup.enter="saveHomeTask"
              />
            </el-form-item>
          </el-col>
          <el-col :span="24">
            <el-form-item label="tapd地址">
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
        <el-row :gutter="12" v-if="homeTaskForm.use_workflow === HOME_TASK_USE_WORKFLOW_YES">
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
                        :loading="homeTaskApiFolderLoading"
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
import GitActionButton from "@/components/base/GitActionButton.vue"

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
const HOME_TASK_CONFIG_TAG_MAX_LENGTH = 15
const HOME_TASK_USE_WORKFLOW_YES = 1
const HOME_TASK_USE_WORKFLOW_NO = 0
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
    tapd_url: '',
    use_workflow: HOME_TASK_USE_WORKFLOW_YES,
    dev_configs: [{ git_id: '', collection_id: '', dir_id: '', docker_id: '', mysql_id: '', local_dir: '', parent_branch: '', branch_name: '', rule_entry_file: '', _branchGenerating: false, smart_link_id: '', smart_link_label: '', smart_link_account: '' }],
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
      HOME_TASK_CONFIG_TAG_MAX_LENGTH,
      HOME_TASK_USE_WORKFLOW_YES,
      HOME_TASK_USE_WORKFLOW_NO,
      homeTaskActiveTab: HOME_TASK_TAB_ACTIVE,
      homeTaskDialogVisible: false,
      homeTaskLoadingActive: false,
      homeTaskLoadingArchived: false,
      homeTaskGeneratingDailyReport: false,
      homeTaskSaving: false,
      homeTaskOperatingId: 0,
      homeTaskOperatingType: '',
      homeTaskActiveList: [],
      homeTaskArchivedList: [],
      homeTaskStatusOptions: HOME_TASK_STATUS_OPTIONS,
      homeTaskForm: createHomeTaskDefaultForm(),
      homeTaskExpandedFragments: {},
      homeTaskEditFeedbackMap: {},
      homeTaskEditFeedbackTimers: {},
      homeTaskEditFeedbackDurationMs: 1000,
      homeTaskLocalDirStatusMap: {},
      homeTaskGitRepoList: [],
      homeTaskGitRepoLoading: false,
      homeTaskApiCollectionList: [],
      homeTaskApiFolderMap: {},
      homeTaskApiCollectionLoading: false,
      homeTaskApiFolderLoading: false,
      homeTaskMysqlList: [],
      homeTaskMysqlLoading: false,
      homeTaskDockerList: [],
      homeTaskDockerLoading: false,
      homeTaskSmartLinkList: [],
      homeTaskSmartLinkLoading: false,
      homeTaskConfigTableColumns: [
        { key: 'git', label: 'Git仓库' },
        { key: 'api', label: '接口集合' },
        { key: 'parent_branch', label: '父分支' },
        { key: 'branch_name', label: '分支名' },
        { key: 'local_dir', label: '本地目录' },
      ],
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
        groupMap[groupName].push({ label: repo.name, value: Number(repo.id) })
      }
      return groupOrder.map(name => ({ label: name, options: groupMap[name] }))
    },
  },
  mounted() {
    this.loadHomeTaskGitRepoList()
    this.loadHomeTaskApiCollections()
    this.loadHomeTaskDockerList()
    this.loadHomeTaskMysqlList()
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
    this.loadHomeTaskSmartLinkList()
  },
  activated() {
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
    this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
    this.loadHomeTaskGitRepoList()
    this.loadHomeTaskApiCollections()
    this.loadHomeTaskDockerList()
    this.loadHomeTaskMysqlList()
    this.loadHomeTaskSmartLinkList()
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
        const taskList = (Array.isArray(response.Data?.task_list) ? response.Data.task_list : []).map(t => ({
          ...t,
          git_ids: safeParseJSON(t.git_ids, []),
          api_dev_entries: safeParseJSON(t.api_dev_entries, []),
          dev_configs: safeParseJSON(t.dev_configs, []),
        }))
        if (isArchived === HOME_TASK_ARCHIVED_YES) {
          this.homeTaskArchivedList = taskList
        } else {
          this.homeTaskActiveList = taskList
        }
        // 预加载 dev_configs 中引用的 API 文件夹
        for (const t of taskList) {
          if (Array.isArray(t.dev_configs)) {
            for (const cfg of t.dev_configs) {
              if (Number(cfg.collection_id || 0) > 0) {
                this.loadHomeTaskApiFoldersForCollection(cfg.collection_id)
              }
            }
          }
        }
        // 批量检查本地目录是否存在
        this.checkLocalDirExists(taskList)
      })
    },
    refreshAllHomeTaskList() {
      this.loadHomeTaskList(HOME_TASK_ARCHIVED_NO)
      this.loadHomeTaskList(HOME_TASK_ARCHIVED_YES)
    },
    resetHomeTaskForm() {
      this.homeTaskForm = createHomeTaskDefaultForm()
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
      this.homeTaskApiFolderLoading = true
      apiManagement.CollectionFoldersBasic({ collection_id: collectionId }, (response) => {
        this.homeTaskApiFolderLoading = false
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
      smartLinkSetApi.SmartLinkList((response) => {
        this.homeTaskSmartLinkLoading = false
        if (!(response && response.ErrCode === 0)) {
          return
        }
        const rawList = Array.isArray(response.Data?.smart_link_list) ? response.Data.smart_link_list : []
        this.homeTaskSmartLinkList = rawList.map(item => {
          let linkList = []
          try {
            linkList = JSON.parse(item.links || '[]')
          } catch (e) { /* ignore */ }
          return { ...item, linkList }
        })
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
      cfg._branchGenerating = true
      homeTaskApi.HomeTaskBranchNameGenerate(taskName, String(cfg.parent_branch || '').trim(), (response) => {
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
      this.loadHomeTaskGitRepoList()
      this.loadHomeTaskApiCollections()
      this.loadHomeTaskMysqlList()
      this.loadHomeTaskDockerList()
      this.loadHomeTaskSmartLinkList()
      this.homeTaskDialogVisible = true
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
        tapd_url: task.tapd_url || '',
        use_workflow: Number(task.use_workflow || HOME_TASK_USE_WORKFLOW_YES) === HOME_TASK_USE_WORKFLOW_YES ? HOME_TASK_USE_WORKFLOW_YES : HOME_TASK_USE_WORKFLOW_NO,
        dev_configs: devConfigs,
      }
      this.loadHomeTaskGitRepoList()
      this.loadHomeTaskApiCollections()
      this.loadHomeTaskMysqlList()
      this.loadHomeTaskDockerList()
      this.loadHomeTaskSmartLinkList()
      for (const cfg of devConfigs) {
        if (cfg.collection_id > 0) {
          this.loadHomeTaskApiFoldersForCollection(cfg.collection_id)
        }
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
    getHomeTaskDevConfigTags(task) {
      const DEV_CONFIG_TAG_TYPE_GIT = 'success'
      const DEV_CONFIG_TAG_TYPE_API = ''
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
          group.push({ type: 'parent_branch', label: '分支: ' + String(cfg.parent_branch).trim(), tagType: '' })
        }
        if (String(cfg.branch_name || '').trim() !== '') {
          group.push({ type: 'branch_name', label: String(cfg.branch_name).trim(), tagType: 'success' })
        }
        if (Number(cfg.smart_link_id || 0) > 0) {
          const sl = this.homeTaskSmartLinkList.find(s => Number(s.id) === Number(cfg.smart_link_id))
          if (sl) {
            group.push({ type: 'smart_link', label: sl.name, id: Number(cfg.smart_link_id), tagType: '' })
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
      if (!taskName) {
        this.$helperNotify.error('任务名称不能为空')
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
      this.homeTaskSaving = true
      this.homeTaskOperatingType = HOME_TASK_OPERATE_SAVE
      homeTaskApi.HomeTaskSave({
        id: Number(this.homeTaskForm.id || 0),
        name: taskName,
        task_status: this.homeTaskForm.task_status,
        start_time: this.convertHomeTaskDateToUnix(this.homeTaskForm.start_date),
        tapd_url: String(this.homeTaskForm.tapd_url || '').trim(),
        dev_configs: JSON.stringify(validConfigs),
        use_workflow: this.homeTaskForm.use_workflow,
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
        const createdTaskId = Number(response?.Data?.id || 0)
        this.$helperNotify.success(isEdit ? '任务已更新' : '任务已创建')
        this.closeHomeTaskDialog()
        if (isEdit) {
          const taskId = Number(this.homeTaskForm.id)
          this.triggerHomeTaskEditFeedback(taskId)
        }
        this.refreshAllHomeTaskList()
        if (!isEdit && createdTaskId > 0 && this.homeTaskForm.use_workflow === HOME_TASK_USE_WORKFLOW_YES) {
          this.openTaskWorkflow({ id: createdTaskId })
        }
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
  },
}
</script>

<style scoped src="@/css/components/HomeTask.css"></style>
