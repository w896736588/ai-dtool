<template>
  <div class="task-workflow-page" v-loading="loading">
    <div class="task-workflow-shell">
      <header class="task-workflow-header">
        <div class="task-workflow-header__main">
          <div class="task-workflow-header__title-row">
            <h1 class="task-workflow-header__title" :title="homeTask.name || `任务 #${taskId}`">{{ homeTask.name || `任务 #${taskId}` }}</h1>
            <div class="task-workflow-header__actions">
          <el-tooltip content="返回首页" placement="bottom">
            <el-button class="task-workflow-home-btn" @click="goHome">
              <el-icon :size="18"><HomeFilled /></el-icon>
            </el-button>
          </el-tooltip>
          <GitActionButton compact variant="primary" @click="taskConfigDialogVisible = true">
            任务配置
          </GitActionButton>
          <el-dropdown trigger="click" @command="handleTaskStatusChange">
            <GitActionButton compact :loading="statusUpdating">
              状态切换（{{ homeTask.task_status }}）
            </GitActionButton>
            <template #dropdown>
              <el-dropdown-menu>
                <el-dropdown-item
                  v-for="status in taskStatusOptions"
                  :key="status"
                  :command="status"
                  :disabled="homeTask.task_status === status"
                >
                  {{ status }}
                </el-dropdown-item>
              </el-dropdown-menu>
            </template>
          </el-dropdown>
          <GitActionButton compact variant="info" @click="goBackToTaskList">
            返回任务清单
          </GitActionButton>
          <GitActionButton compact :loading="loading" @click="reloadWorkflowPage">
            刷新
          </GitActionButton>
          <GitActionButton compact variant="warning" @click="openIssueFixDialog">
            问题修改提示词
          </GitActionButton>
          <ChatHistoryButton
            compact
            variant="info"
            :running="getPromptChatCounts('issue_fix').running > 0"
            :running-count="getPromptChatCounts('issue_fix').running"
            :interrupted-count="getPromptChatCounts('issue_fix').interrupted"
            :total-count="getPromptChatCounts('issue_fix').total"
            :unread="hasUnreadInPromptType('issue_fix')"
            @click="openChatHistoryDialog"
          >
            历史对话
          </ChatHistoryButton>
          <!--
          <GitActionButton compact variant="success" @click="openZcodeConfigDialog">
            zcode配置
          </GitActionButton>
          -->
            </div>
          </div>
          <div v-if="parsedTaskDevConfigs.length > 0" class="task-workflow-header__meta">
            <div v-for="(cfg, idx) in parsedTaskDevConfigs" :key="idx" class="task-workflow-header__dev-card">
              <div class="task-workflow-header__field-grid">
                <!-- Git仓库 -->
                <div class="task-workflow-header__field task-workflow-header__field--compact">
                  <span class="task-workflow-header__field-label">Git仓库</span>
                  <span class="task-workflow-header__field-value" :title="getTaskConfigName('git', cfg.git_id)">{{ getTaskConfigName('git', cfg.git_id) }}</span>
                </div>
                <!-- 接口集合 -->
                <div class="task-workflow-header__field task-workflow-header__field--link task-workflow-header__field--wrap" @click="openApiDevDialog(cfg)">
                  <span class="task-workflow-header__field-label">接口集合</span>
                  <span class="task-workflow-header__field-value task-workflow-header__field-value--wrap">{{ getTaskConfigApiLabel(cfg) || '-' }}</span>
                </div>
                <!-- 分支名 -->
                <div class="task-workflow-header__field task-workflow-header__field--wrap task-workflow-header__field--branch">
                  <span class="task-workflow-header__field-label">分支名</span>
                  <span class="task-workflow-header__field-value task-workflow-header__field-value--wrap">
                    <span class="task-workflow-header__branch" @click="copyText(cfg.branch_name, '分支名已复制')" :title="cfg.branch_name">{{ cfg.branch_name || '-' }}</span>
                    <el-tooltip v-if="cfg.local_dir && cfg.branch_name && branchStatusMap[cfg.local_dir + '|' + cfg.branch_name]" :content="branchStatusMap[cfg.local_dir + '|' + cfg.branch_name].matched ? '分支匹配' : '当前分支: ' + (branchStatusMap[cfg.local_dir + '|' + cfg.branch_name].current_branch || '未知')" placement="top">
                      <span class="task-workflow-header__status-icon" :class="branchStatusMap[cfg.local_dir + '|' + cfg.branch_name].matched ? 'task-workflow-header__status-icon--ok' : 'task-workflow-header__status-icon--err'">
                        <svg v-if="branchStatusMap[cfg.local_dir + '|' + cfg.branch_name].matched" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right:3px"><rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 21h8"/><path d="M12 17v4"/></svg><svg v-if="branchStatusMap[cfg.local_dir + '|' + cfg.branch_name].matched" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg>
                        <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                      </span>
                    </el-tooltip>
                    <!-- 远程分支状态（推送状态 + 远程工作目录分支） -->
                    <el-tooltip v-if="showRemoteBranchWarning(cfg)" :content="getRemoteBranchStatusTooltip(cfg)" placement="top">
                      <span class="task-workflow-header__status-icon task-workflow-header__status-icon--remote-warn" @click.stop="openRemoteBranchDialog(cfg)">
                        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right:3px"><path d="M18 10h-1.26A8 8 0 109 20h9a5 5 0 000-10z"/></svg><svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><path d="M10.29 3.86L1.82 18a2 2 0 001.71 3h16.94a2 2 0 001.71-3L13.71 3.86a2 2 0 00-3.42 0z"/><line x1="12" y1="9" x2="12" y2="13"/><line x1="12" y1="17" x2="12.01" y2="17"/></svg>
                      </span>
                    </el-tooltip>
                    <el-tooltip v-else-if="showRemoteBranchOk(cfg)" content="远程分支已同步且工作目录分支匹配" placement="top">
                      <span class="task-workflow-header__status-icon task-workflow-header__status-icon--ok">
                        <svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" style="margin-right:3px"><path d="M18 10h-1.26A8 8 0 109 20h9a5 5 0 000-10z"/></svg><svg viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg>
                      </span>
                    </el-tooltip>
                  </span>
                </div>
                <!-- 本地目录 -->
                <div class="task-workflow-header__field">
                  <span class="task-workflow-header__field-label">本地目录</span>
                  <span class="task-workflow-header__field-value">
                    {{ cfg.local_dir || '-' }}
                    <el-tooltip v-if="cfg.local_dir && localDirStatusMap[cfg.local_dir] !== undefined" :content="localDirStatusMap[cfg.local_dir] ? '目录存在' : '目录不存在'" placement="top">
                      <span class="task-workflow-header__status-icon" :class="localDirStatusMap[cfg.local_dir] ? 'task-workflow-header__status-icon--ok' : 'task-workflow-header__status-icon--err'">
                        <svg v-if="localDirStatusMap[cfg.local_dir]" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg>
                        <svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg>
                      </span>
                    </el-tooltip>
                    <el-dropdown v-if="cfg.local_dir" trigger="click" @command="(editor) => openInEditor(cfg.local_dir, editor)" style="margin-left:6px">
                      <span class="task-workflow-header__open-editor-btn" title="选择 IDE 打开目录">▶</span>
                      <template #dropdown>
                        <el-dropdown-menu>
                          <el-dropdown-item command="vscode">VS Code</el-dropdown-item>
                          <el-dropdown-item command="cursor">Cursor</el-dropdown-item>
                          <el-dropdown-item command="goland">GoLand</el-dropdown-item>
                          <el-dropdown-item command="phpstorm">PhpStorm</el-dropdown-item>
                        </el-dropdown-menu>
                      </template>
                    </el-dropdown>
                  </span>
                </div>
                <!-- 文件变更 -->
                <div class="task-workflow-header__field task-workflow-header__field--wrap">
                  <span class="task-workflow-header__field-label">文件变更</span>
                  <span class="task-workflow-header__field-value task-workflow-header__field-value--wrap">
                    <template v-if="cfg.local_dir && fileChangesMap[cfg.local_dir] && !fileChangesMap[cfg.local_dir].error">
                      <span class="file-changes-inline file-changes-inline--clickable" title="点击查看文件变更详情" @click="openFileChangesDetail(cfg)">
                        <span class="file-changes-inline__item file-changes-inline__item--committed" :title="'已提交文件数'">{{ fileChangesMap[cfg.local_dir].summary.committed || 0 }}<small>(C)</small></span>
                        <span class="file-changes-inline__sep">/</span>
                        <span class="file-changes-inline__item file-changes-inline__item--staged" :title="'已暂存文件数'">{{ fileChangesMap[cfg.local_dir].summary.staged || 0 }}<small>(S)</small></span>
                        <span class="file-changes-inline__sep">/</span>
                        <span class="file-changes-inline__item file-changes-inline__item--modified" :title="'已修改文件数'">{{ fileChangesMap[cfg.local_dir].summary.modified || 0 }}<small>(M)</small></span>
                        <span class="file-changes-inline__sep">/</span>
                        <span class="file-changes-inline__item file-changes-inline__item--untracked" :title="'未跟踪文件数'">{{ fileChangesMap[cfg.local_dir].summary.untracked || 0 }}<small>(U)</small></span>
                      </span>
                    </template>
                    <template v-else-if="cfg.local_dir && fileChangesMap[cfg.local_dir] && fileChangesMap[cfg.local_dir].error">
                      <span class="task-workflow-header__field-value--dim" :title="fileChangesMap[cfg.local_dir].error">检测失败</span>
                    </template>
                    <template v-else>
                      <span class="task-workflow-header__field-value--dim">-</span>
                    </template>
                  </span>
                </div>
              </div>
            </div>
          </div>
        </div>
      </header>

      <el-alert
        v-if="errorMessage"
        type="error"
        :closable="false"
        :title="errorMessage"
        class="task-workflow-alert"
      />

      <el-dialog
        v-model="branchMismatchDialogVisible"
        title="工作区分支检查"
        width="820px"
        :close-on-click-modal="true"
        @close="handleBranchMismatchDialogClose"
      >
        <div v-loading="branchMismatchLoading" class="branch-mismatch-dialog">
          <div class="branch-mismatch-dialog__summary">
            <span>一致 {{ branchMismatchMatchedCount }}</span>
            <span>不一致 {{ branchMismatchMismatchedCount }}</span>
          </div>
          <div v-if="branchMismatchDetailList.length === 0" class="branch-mismatch-dialog__empty">暂无分支检查结果</div>
          <div v-for="item in branchMismatchDetailList" :key="getBranchMismatchItemKey(item)" class="branch-mismatch-card">
            <div class="branch-mismatch-card__header">
              <div class="branch-mismatch-card__title">{{ item.local_dir || '-' }}</div>
              <el-tag :type="item.matched ? 'success' : 'danger'" size="small">{{ item.matched ? '一致' : '不一致' }}</el-tag>
            </div>
            <div class="branch-mismatch-card__meta">
              <div>父分支：{{ item.parent_branch || '-' }}</div>
              <div>期望分支：{{ item.expected_branch || '-' }}</div>
              <div>当前分支：{{ item.current_branch || '-' }}</div>
            </div>
            <div v-if="item.error" class="branch-mismatch-card__error">检查失败：{{ item.error }}</div>
            <div v-if="!item.matched" class="branch-mismatch-card__changes">
              <div class="branch-mismatch-card__changes-title">未提交或已变更文件</div>
              <div v-if="item.changed_files_error" class="branch-mismatch-card__error">{{ item.changed_files_error }}</div>
              <div v-else-if="Array.isArray(item.changed_files) && item.changed_files.length > 0" class="branch-mismatch-card__file-list">
                <div v-for="(file, fileIdx) in item.changed_files" :key="fileIdx" class="branch-mismatch-card__file-item">{{ file }}</div>
              </div>
              <div v-else class="branch-mismatch-card__empty-tip">当前没有未提交文件，但仍会清理工作区后重建目标分支。</div>
            </div>
            <div v-if="!item.matched" class="branch-mismatch-card__actions">
              <GitActionButton
                compact
                variant="danger"
                :loading="branchSwitchingKey === getBranchMismatchItemKey(item)"
                @click="handleCleanupAndSwitchBranch(item)"
              >
                清理并切换指定分支
              </GitActionButton>
            </div>
          </div>
        </div>
        <template #footer>
          <div class="branch-mismatch-dialog__footer">
            <GitActionButton compact variant="info" @click="closeBranchMismatchDialog(false)">关闭</GitActionButton>
            <GitActionButton
              v-if="branchMismatchDialogMode === 'exec_confirm'"
              compact
              :disabled="branchMismatchLoading"
              @click="closeBranchMismatchDialog(true)"
            >
              继续执行
            </GitActionButton>
          </div>
        </template>
      </el-dialog>

      <el-dialog
        v-model="branchSwitchStreamDialogVisible"
        title="分支切换日志"
        width="860px"
        :close-on-click-modal="true"
        @close="closeBranchSwitchStreamDialog"
      >
        <div class="branch-switch-stream">
          <div class="branch-switch-stream__meta">
            <div>执行位置：本地工作目录</div>
            <div>目录：{{ branchSwitchStreamMeta.local_dir || '-' }}</div>
            <div>基线分支：{{ branchSwitchStreamMeta.base_branch || '-' }}</div>
            <div>目标分支：{{ branchSwitchStreamMeta.branch_name || '-' }}</div>
          </div>
          <div v-if="branchSwitchStreamRunning" class="branch-switch-stream__status">
            <span class="branch-switch-stream__spinner" />
            <span>命令执行中，请稍候...</span>
          </div>
          <div ref="branchSwitchStreamLog" class="branch-switch-stream__log">
            <div v-if="branchSwitchStreamLines.length === 0" class="branch-switch-stream__placeholder">
              正在建立执行连接，日志会在这里实时输出...
            </div>
            <div
              v-for="(line, idx) in branchSwitchStreamLines"
              :key="idx"
              class="branch-switch-stream__line"
              :class="'branch-switch-stream__line--' + (line.level || 'info')"
            >
              {{ line.text }}
            </div>
          </div>
        </div>
        <template #footer>
          <div class="branch-switch-stream__footer">
            <GitActionButton compact variant="info" @click="closeBranchSwitchStreamDialog">关闭</GitActionButton>
          </div>
        </template>
      </el-dialog>

      <!-- 文件变更详情弹窗 -->
      <FileChangesDialog
        v-model:visible="fileChangesDialogVisible"
        :local-dir="fileChangesDetailLocalDir"
        :parent-branch="fileChangesDetailParentBranch"
        :initial-summary="fileChangesDetailInitialSummary"
        :initial-files="fileChangesDetailInitialFiles"
      />

      <!-- 远程分支状态弹窗 -->
      <el-dialog
        v-model="remoteBranchDialogVisible"
        title="远程分支检查"
        width="700px"
        :close-on-click-modal="true"
        @close="closeRemoteBranchDialog"
      >
        <div v-loading="remoteBranchDialogLoading" class="remote-branch-dialog">
          <div class="remote-branch-dialog__info">
            <div class="remote-branch-dialog__row">
              <span class="remote-branch-dialog__label">本地目录</span>
              <span class="remote-branch-dialog__value">{{ remoteBranchDialogItem.local_dir || '-' }}</span>
            </div>
            <div class="remote-branch-dialog__row">
              <span class="remote-branch-dialog__label">当前分支</span>
              <span class="remote-branch-dialog__value">{{ remoteBranchDialogItem.current_branch || '-' }}</span>
            </div>
            <div class="remote-branch-dialog__row">
              <span class="remote-branch-dialog__label">期望分支</span>
              <span class="remote-branch-dialog__value">{{ remoteBranchDialogItem.branch_name || '-' }}</span>
            </div>
            <div class="remote-branch-dialog__row">
              <span class="remote-branch-dialog__label">远程工作空间的本地分支</span>
              <span class="remote-branch-dialog__value">
                <template v-if="remoteBranchDialogItem.remote_dir_current_branch">
                  {{ remoteBranchDialogItem.remote_dir_current_branch }}
                  <el-tag :type="remoteBranchDialogItem.remote_dir_branch_match ? 'success' : 'warning'" size="small" style="margin-left: 8px;">
                    {{ remoteBranchDialogItem.remote_dir_branch_match ? '匹配' : '不匹配' }}
                  </el-tag>
                </template>
                <span v-else style="color: #c0c4cc;">-</span>
              </span>
            </div>
            <div class="remote-branch-dialog__row">
              <span class="remote-branch-dialog__label">远程工作空间的远程分支</span>
              <span class="remote-branch-dialog__value">{{ remoteBranchDialogItem.remote_dir_remote_branch || '-' }}</span>
            </div>
          </div>
          <div class="remote-branch-dialog__status">
            <div class="remote-branch-dialog__status-row">
              <span class="remote-branch-dialog__status-label">推送状态</span>
              <el-tag :type="remoteBranchDialogItem.pushed ? 'success' : 'danger'" size="small">
                {{ remoteBranchDialogItem.pushed ? '已推送' : '未推送' }}
              </el-tag>
            </div>
            <div class="remote-branch-dialog__status-row">
              <span class="remote-branch-dialog__status-label">远程分支</span>
              <el-tag :type="remoteBranchDialogItem.remote_exists ? 'success' : 'info'" size="small">
                {{ remoteBranchDialogItem.remote_exists ? (remoteBranchDialogItem.remote_branch_name || '存在') : '不存在' }}
              </el-tag>
            </div>
            <div class="remote-branch-dialog__status-row">
              <span class="remote-branch-dialog__status-label">同步状态</span>
              <span v-if="remoteBranchDialogItem.error" class="remote-branch-dialog__error">{{ remoteBranchDialogItem.error }}</span>
              <template v-else-if="remoteBranchDialogItem.remote_exists">
                <span style="color: #303133; font-size: 13px;">
                  本地领先 <b>{{ remoteBranchDialogItem.local_ahead }}</b> 个提交，远程领先 <b>{{ remoteBranchDialogItem.remote_ahead }}</b> 个提交
                </span>
                <el-tag v-if="remoteBranchDialogItem.consistent" type="success" size="small">已同步</el-tag>
                <el-tag v-else type="warning" size="small">未同步</el-tag>
              </template>
              <span v-else style="color: #909399; font-size: 13px;">—</span>
            </div>
            <div class="remote-branch-dialog__status-row">
              <span class="remote-branch-dialog__status-label">远程工作目录分支</span>
              <template v-if="remoteBranchDialogItem.remote_dir_error">
                <span class="remote-branch-dialog__error">{{ remoteBranchDialogItem.remote_dir_error }}</span>
                <el-tag type="danger" size="small">检测失败</el-tag>
              </template>
              <template v-else-if="remoteBranchDialogItem.remote_dir_branch_match !== undefined">
                <span style="color: #303133; font-size: 13px;">{{ remoteBranchDialogItem.remote_dir_current_branch || '-' }}</span>
                <el-tag :type="remoteBranchDialogItem.remote_dir_branch_match ? 'success' : 'warning'" size="small">
                  {{ remoteBranchDialogItem.remote_dir_branch_match ? '一致' : '不一致' }}
                </el-tag>
              </template>
              <span v-else style="color: #909399; font-size: 13px;">—</span>
            </div>
          </div>
          <div v-if="remoteBranchDialogPushResult" class="remote-branch-dialog__result" :class="remoteBranchDialogPushResult.success ? 'remote-branch-dialog__result--ok' : 'remote-branch-dialog__result--err'">
            {{ remoteBranchDialogPushResult.message }}
          </div>
        </div>
        <template #footer>
          <div class="remote-branch-dialog__footer">
            <GitActionButton compact variant="info" @click="closeRemoteBranchDialog">关闭</GitActionButton>
            <GitActionButton compact variant="warning" @click="refreshRemoteBranchDialog">刷新检测</GitActionButton>
            <GitActionButton
              compact
              variant="success"
              :loading="remoteBranchPushing"
              @click="handleRemoteBranchPush"
            >
              推送并切换分支
            </GitActionButton>
          </div>
        </template>
      </el-dialog>

      <section class="task-workflow-nodes">
        <button
          v-for="(node, idx) in workflowNodes"
          :key="node.key"
          type="button"
          class="task-workflow-node"
          :class="{
            'task-workflow-node--active': activeNode === node.key,
            'task-workflow-node--status-pending': getNodeStatus(node.key) === 'pending',
            'task-workflow-node--status-running': getNodeStatus(node.key) === 'running',
            'task-workflow-node--status-completed': getNodeStatus(node.key) === 'completed',
            'task-workflow-node--status-skipped': getNodeStatus(node.key) === 'skipped',
          }"
          @click="selectNode(node.key)"
        >
          <span class="task-workflow-node__row">
            <span class="task-workflow-node__badge">{{ idx + 1 }}</span>
            <span class="task-workflow-node__label">{{ node.label }}</span>
            <span class="task-workflow-node__status-icon">
              <span v-if="getNodeStatus(node.key) === 'completed'" class="status-icon status-icon--completed">&#10003;</span>
              <span v-else-if="getNodeStatus(node.key) === 'skipped'" class="status-icon status-icon--skipped">&#10003;</span>
              <span v-else-if="getNodeStatus(node.key) === 'pending'" class="status-icon status-icon--pending"></span>
              <span v-else class="status-icon status-icon--running"><span class="spinner-ring"></span></span>
            </span>

          </span>
        </button>
      </section>

      <section class="task-workflow-content">
        <!-- 抓取需求步骤（固定显示知识片段，无提示词） -->
        <template v-if="activeNode === 'requirement-fetch'">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">
                {{ getActiveNodeLabel() }}
                <span v-if="getActiveNodeDesc()" class="task-workflow-card__title-desc">{{ getActiveNodeDesc() }}</span>
              </div>
              <div class="task-workflow-card__switch">
                <GitActionButton compact :loading="requirementFetchRunning" @click="triggerRequirementFetch(false)">
                  重新抓取
                </GitActionButton>
                <GitActionButton compact variant="info" @click="openFragmentInDialog(requirementFragmentId, requirementSourceName + '需求文档')" :disabled="!requirementFragmentId">
                  <template #icon><el-icon><Link /></el-icon></template>
                  打开知识片段
                </GitActionButton>
                <div class="task-workflow-node-status-inline">
                  <span class="task-workflow-node-status-inline__label">当前步骤状态</span>
                  <button
                    class="task-workflow-node-status-inline__btn"
                    :class="'task-workflow-node-status-inline__btn--' + getNodeStatus('requirement-fetch')"
                    :disabled="nodeStatusSaving"
                    @click="cycleNodeStatus('requirement-fetch')"
                  >{{ getNodeStatusLabel('requirement-fetch') }}</button>
                </div>
              </div>
            </div>
            <div v-if="workflow.requirement_fetch_error" class="task-workflow-card__hint task-workflow-card__hint--error">
              最近错误：{{ workflow.requirement_fetch_error }}
            </div>
            <div v-if="!requirementSourceUrl" class="task-workflow-card__hint">
              当前任务未配置 {{ requirementSourceName }} 地址，无法自动抓取。
            </div>
            <div class="task-workflow-fragment-view">
              <iframe
                v-if="requirementShareUrl"
                :src="requirementShareUrl"
                class="task-workflow-fragment-view__iframe"
                title="需求知识片段预览"
              />
              <div v-else class="task-workflow-fragment-view__empty">
                知识片段分享链接生成中...
              </div>
            </div>
        </template>

        <!-- 通用步骤渲染（左侧Tab: 提示词 + 文档） -->
        <template v-else>
            <!-- 左侧Tab + 右侧内容 -->
            <div class="step-tab-layout">
              <div class="step-tab-sidebar">
                <button
                  type="button"
                  class="step-tab-btn"
                  :class="{ 'step-tab-btn--active': stepActiveTab === '__prompt__', 'step-tab-btn--prompt': true }"
                  @click="stepActiveTab = '__prompt__'"
                >提示词</button>
                <button
                  v-for="doc in getActiveStepDocuments()"
                  :key="doc.id || doc.name"
                  type="button"
                  class="step-tab-btn"
                  :class="{ 'step-tab-btn--active': stepActiveTab === (doc.id || doc.name), 'step-tab-btn--doc': true, 'step-tab-btn--unread': docUnreadFlags[doc.id || doc.name] }"
                  :title="doc.name"
                  @click="switchToDocTab(doc)"
                ><span class="step-tab-btn__dot" v-if="docUnreadFlags[doc.id || doc.name]"></span>{{ doc.name }}</button>
                <div class="step-tab-sidebar__actions">
                  <div class="step-tab-sidebar__actions-title">步骤操作</div>
                  <div class="task-workflow-node-status-inline">
                    <button
                      class="task-workflow-node-status-inline__btn"
                      :class="'task-workflow-node-status-inline__btn--' + getNodeStatus(activeNode)"
                      :disabled="nodeStatusSaving"
                      @click="cycleNodeStatus(activeNode)"
                    >{{ getNodeStatusLabel(activeNode) }}</button>
                  </div>
                  <GitActionButton compact variant="success" @click="openPromptExecDialog(activeNode, getStepPrompt(activeNode))">
                    <template #icon><el-icon><VideoPlay /></el-icon></template>
                    执行
                  </GitActionButton>
                  <ChatHistoryButton
                    compact
                    variant="info"
                    :running="getPromptChatCounts(activeNode).running > 0"
                    :running-count="getPromptChatCounts(activeNode).running"
                    :interrupted-count="getPromptChatCounts(activeNode).interrupted"
                    :total-count="getPromptChatCounts(activeNode).total"
                    :unread="hasUnreadInPromptType(activeNode)"
                    @click="openPromptChatHistory(activeNode)"
                  >
                    执行历史
                  </ChatHistoryButton>
                  <GitActionButton compact variant="warning" :loading="promptRestoring === activeNode" @click="restorePrompts(activeNode)" :disabled="stepActiveTab !== '__prompt__'">
                    还原为默认提示词
                  </GitActionButton>
                  <GitActionButton v-if="activeStepHasApiDoc()" compact variant="warning" :loading="apiDocResetting" @click="resetApiDoc">
                    重置接口文档
                  </GitActionButton>
                </div>
              </div>
              <div class="step-tab-content">
                <!-- 提示词 Tab -->
                <div v-if="stepActiveTab === '__prompt__'" class="step-tab-panel">
                  <div class="editor-body-toolbar editor-body-toolbar--saved">
                    <div class="editor-body-toolbar-main">
                      <div class="editor-body-toolbar-top">
                        <div class="editor-body-toolbar-left">
                          <el-input
                            v-model="stepPromptTitle"
                            class="title-input editor-toolbar-title-input"
                            placeholder="输入步骤名称"
                          />
                        </div>
                        <div class="editor-body-toolbar-right">
                          <div class="editor-body-actions">
                            <el-tooltip content="查看" placement="top">
                              <GitActionButton
                                variant="info"
                                compact
                                class="toolbar-icon-button"
                                :class="{ 'mode-button-active': !promptEditMode }"
                                @click="promptEditMode = false"
                              >
                                <el-icon><View /></el-icon>
                              </GitActionButton>
                            </el-tooltip>
                            <el-tooltip content="编辑" placement="top">
                              <GitActionButton
                                compact
                                class="toolbar-icon-button"
                                :class="{ 'mode-button-active': promptEditMode }"
                                @click="promptEditMode = true"
                              >
                                <el-icon><Edit /></el-icon>
                              </GitActionButton>
                            </el-tooltip>
                            <el-tooltip content="复制内容" placement="top">
                              <GitActionButton
                                variant="info"
                                compact
                                class="toolbar-icon-button"
                                @click="copyText(getStepPrompt(activeNode), '内容已复制')"
                              >
                                <el-icon><CopyDocument /></el-icon>
                              </GitActionButton>
                            </el-tooltip>
                            <el-tooltip content="保存" placement="top">
                              <GitActionButton
                                compact
                                class="toolbar-icon-button"
                                :loading="promptSaving === activeNode"
                                @click="savePrompts(activeNode)"
                              >
                                <el-icon><Check /></el-icon>
                              </GitActionButton>
                            </el-tooltip>
                          </div>
                        </div>
                      </div>
                    </div>
                  </div>
                  <div v-if="promptEditMode" class="task-workflow-prompt-editor" :data-prompt-type="activeNode">
                    <UnifiedMdEditor
                      :model-value="getStepPrompt(activeNode)"
                      @update:model-value="val => setStepPrompt(activeNode, val)"
                      preview-theme="github"
                      :preview="true"
                      :toolbars="promptEditorToolbars"
                      height="100%"
                    />
                  </div>
                  <div v-else class="preview-body" :class="{ 'preview-body--with-outline': promptTocItems.length > 0 }">
                    <div ref="promptPreviewBody" class="preview-renderer" @scroll="handlePromptPreviewScroll">
                      <MdPreview
                        :model-value="getStepPrompt(activeNode)"
                        preview-theme="github"
                        :auto-fold-threshold="999999"
                        :on-get-catalog="onPromptCatalog"
                      />
                    </div>
                    <aside v-if="promptTocItems.length > 0" class="preview-outline">
                      <div class="preview-outline-card">
                        <div class="preview-outline-title">目录</div>
                        <button
                          v-for="item in promptTocItems"
                          :key="item.id"
                          type="button"
                          class="preview-outline-item"
                          :class="[
                            { active: promptActiveHeading === item.id },
                            { 'preview-outline-item--child': item.level > 1 },
                            { 'preview-outline-item--grandchild': item.level > 2 },
                          ]"
                          @click="scrollToPromptHeading(item.id)"
                        >
                          {{ item.text }}
                        </button>
                      </div>
                    </aside>
                  </div>
                </div>
                <!-- 文档 Tab -->
                <div v-else class="step-tab-panel">
                  <div class="editor-body-toolbar editor-body-toolbar--saved">
                    <div class="editor-body-toolbar-main">
                      <div class="editor-body-toolbar-top">
                        <div class="editor-body-toolbar-left">
                          <el-input
                            v-model="stepDocTitle"
                            class="title-input editor-toolbar-title-input"
                            placeholder="输入文档名称"
                          />
                        </div>
                        <div class="editor-body-toolbar-right">
                          <div class="editor-body-actions">
                            <el-tooltip content="查看" placement="top">
                              <GitActionButton
                                variant="info"
                                compact
                                class="toolbar-icon-button"
                                :class="{ 'mode-button-active': !docEditMode }"
                                @click="switchDocToViewMode"
                              >
                                <el-icon><View /></el-icon>
                              </GitActionButton>
                            </el-tooltip>
                            <el-tooltip content="编辑" placement="top">
                              <GitActionButton
                                compact
                                class="toolbar-icon-button"
                                :class="{ 'mode-button-active': docEditMode }"
                                @click="switchDocToEditMode"
                              >
                                <el-icon><Edit /></el-icon>
                              </GitActionButton>
                            </el-tooltip>
                            <el-tooltip content="复制内容" placement="top">
                              <GitActionButton
                                variant="info"
                                compact
                                class="toolbar-icon-button"
                                @click="copyText(getActiveDocContent(), '内容已复制')"
                                :disabled="!isActiveDocLoaded()"
                              >
                                <el-icon><CopyDocument /></el-icon>
                              </GitActionButton>
                            </el-tooltip>
                            <el-tooltip content="下载ZIP" placement="top">
                              <GitActionButton
                                variant="info"
                                compact
                                class="toolbar-icon-button"
                                @click="downloadActiveDocZip"
                                :disabled="!getActiveDocFileId()"
                              >
                                <el-icon><Download /></el-icon>
                              </GitActionButton>
                            </el-tooltip>
                            <el-tooltip content="保存" placement="top">
                              <GitActionButton
                                compact
                                class="toolbar-icon-button"
                                @click="saveActiveDoc"
                                :disabled="!isActiveDocLoaded()"
                              >
                                <el-icon><Check /></el-icon>
                              </GitActionButton>
                            </el-tooltip>
                            <el-tooltip content="分享" placement="top">
                              <GitActionButton
                                variant="info"
                                compact
                                class="toolbar-icon-button"
                                :loading="docSharing"
                                @click="shareActiveDoc"
                                :disabled="!getActiveDocFileId()"
                              >
                                <el-icon><Share /></el-icon>
                              </GitActionButton>
                            </el-tooltip>
                            <el-tooltip content="打开知识片段" placement="top">
                              <GitActionButton
                                variant="info"
                                compact
                                class="toolbar-icon-button"
                                @click="openActiveDocFragment"
                                :disabled="!getActiveDocFileId()"
                              >
                                <el-icon><Link /></el-icon>
                              </GitActionButton>
                            </el-tooltip>
                            <el-tooltip content="全屏编辑" placement="top">
                              <GitActionButton
                                variant="info"
                                compact
                                class="toolbar-icon-button"
                                @click="openActiveDocFullscreen"
                                :disabled="!getActiveDocFileId()"
                              >
                                <el-icon><FullScreen /></el-icon>
                              </GitActionButton>
                            </el-tooltip>
                            <el-tooltip content="搜索" placement="top">
                              <GitActionButton
                                variant="info"
                                compact
                                class="toolbar-icon-button"
                                :class="{ 'mode-button-active': docSearchBarVisible }"
                                @click="toggleDocSearchBar"
                                :disabled="!isActiveDocLoaded()"
                              >
                                <el-icon><Search /></el-icon>
                              </GitActionButton>
                            </el-tooltip>
                            <el-dropdown
                              trigger="click"
                              class="editor-action-dropdown"
                              @command="handleDocToolbarAction"
                            >
                              <button class="editor-action-trigger" type="button" aria-label="更多操作">
                                <el-icon><MoreFilled /></el-icon>
                              </button>
                              <template #dropdown>
                                <el-dropdown-menu>
                                  <el-dropdown-item command="history">
                                    历史记录
                                  </el-dropdown-item>
                                </el-dropdown-menu>
                              </template>
                            </el-dropdown>
                          </div>
                        </div>
                      </div>
                  <!-- 文档搜索栏 -->
                  <div v-show="docSearchBarVisible" class="editor-search-row">
                    <div class="editor-search-main">
                      <el-input
                        ref="docSearchInput"
                        v-model="docSearchQuery"
                        class="editor-search-input"
                        clearable
                        placeholder="在当前文档中搜索..."
                        @input="handleDocSearchInput"
                        @keydown.enter.prevent="jumpDocSearchMatch(1)"
                        @keydown.shift.enter.prevent="jumpDocSearchMatch(-1)"
                        @keydown.esc.prevent="clearDocSearch"
                        @clear="clearDocSearch"
                      />
                      <span class="editor-search-summary">
                        {{ docSearchSummaryText }}
                      </span>
                    </div>
                    <div class="editor-search-actions">
                      <GitActionButton
                        variant="info"
                        :disabled="!docHasSearchMatches"
                        @click="jumpDocSearchMatch(-1)"
                      >
                        上一个
                      </GitActionButton>
                      <GitActionButton
                        variant="info"
                        :disabled="!docHasSearchMatches"
                        @click="jumpDocSearchMatch(1)"
                      >
                        下一个
                      </GitActionButton>
                      <GitActionButton
                        variant="info"
                        :disabled="!docSearchQuery"
                        @click="clearDocSearch"
                      >
                        清空
                      </GitActionButton>
                    </div>
                  </div>
                    </div>
                  </div>
              <div v-if="getActiveDocLoading()" class="step-tab-panel__loading">
                    <el-icon class="is-loading"><Loading /></el-icon>
                    <span>加载中...</span>
                  </div>
                  <div v-else-if="!isActiveDocLoaded()" class="step-tab-panel__empty">
                    <el-empty description="暂无文档内容" :image-size="60" />
                  </div>
                  <template v-else>
                    <!-- 编辑模式：v-show 保持 DOM 存活，避免 md-editor mount/unmount 崩溃 -->
                    <div v-show="docEditMode" class="task-workflow-doc-editor">
                        <UnifiedMdEditor
                          :model-value="getActiveDocContent()"
                          @update:model-value="val => setActiveDocContent(val)"
                          preview-theme="github"
                          :preview="true"
                          :toolbars="promptEditorToolbars"
                          height="100%"
                        />
                    </div>
                    <!-- 查看模式 -->
                    <div v-show="!docEditMode" class="preview-body" :class="{ 'preview-body--with-outline': docTocItems.length > 0 }">
                        <div ref="docPreviewBody" class="preview-renderer" @scroll="handleDocPreviewScroll">
                          <MdPreview
                            :model-value="getActiveDocContent()"
                            preview-theme="github"
                            :auto-fold-threshold="999999"
                            :on-get-catalog="onDocCatalog"
                          />
                        </div>
                        <aside v-if="docTocItems.length > 0" class="preview-outline">
                          <div class="preview-outline-card">
                            <div class="preview-outline-title">目录</div>
                            <button
                              v-for="item in docTocItems"
                              :key="item.id"
                              type="button"
                              class="preview-outline-item"
                              :class="[
                                { active: docActiveHeading === item.id },
                                { 'preview-outline-item--child': item.level > 1 },
                                { 'preview-outline-item--grandchild': item.level > 2 },
                              ]"
                              @click="scrollToDocHeading(item.id)"
                            >
                              {{ item.text }}
                            </button>
                          </div>
                        </aside>
                    </div>
                  </template>
                </div>
              </div>
            </div>
        </template>
      </section>
    </div>

    <el-dialog
      v-model="fragmentDialogVisible"
      :title="fragmentDialogTitle"
      width="80%"
      top="3vh"
      destroy-on-close
      class="task-workflow-fragment-dialog"
    >
      <div class="task-workflow-fragment-dialog__body">
        <iframe
          v-if="fragmentDialogUrl"
          :src="fragmentDialogUrl"
          class="task-workflow-fragment-dialog__iframe"
        />
        <div v-else class="task-workflow-fragment-dialog__empty">
          暂无内容
        </div>
      </div>
    </el-dialog>

    <MemoryHistoryDialog
      v-model="docHistoryVisible"
      :fragment-id="docHistoryFragmentId"
    />

    <el-dialog
      v-model="issueFixDialogVisible"
      title="问题修改提示词"
      width="1200px"
      top="3vh"
      destroy-on-close
      :show-close="false"
      class="task-workflow-issue-fix-dialog"
    >
      <div class="task-workflow-issue-fix__close-bar">
        <el-button @click="issueFixDialogVisible = false" type="danger">关闭</el-button>
      </div>
        <div style="margin-bottom: 12px; display: flex; gap: 8px;">
          <el-button type="primary" @click="sendToClaudeCode">
           执行
          </el-button>
        </div>
        <div v-if="issueFixZcodeMappings.length > 0" style="margin-bottom: 12px;">
          <div style="font-size: 13px; color: #909399; margin-bottom: 4px;">当前任务本地目录对应的 Settings 配置</div>
          <el-table :data="issueFixZcodeMappings" border size="small" max-height="160">
            <el-table-column prop="workspace_path" label="本地工作目录" show-overflow-tooltip />
            <el-table-column prop="settings_path" label="Settings 配置文件" show-overflow-tooltip />
          </el-table>
        </div>
      <div class="task-workflow-issue-fix">
        <div class="task-workflow-issue-fix__input">
          <div class="task-workflow-issue-fix__label">改动要求</div>
          <el-input
            v-model="issueFixInput"
            type="textarea"
            :rows="4"
            placeholder="请描述需要修改的问题"
          />
        </div>
        <div class="task-workflow-issue-fix__output">
          <div style="display: flex; align-items: center; gap: 8px; margin-bottom: 4px;">
            <span class="task-workflow-issue-fix__label" style="margin-bottom: 0;">完整提示词</span>
            <el-switch
              v-model="issueFixUseDefaultPrompt"
              active-text="使用默认提示词"
              inactive-text=""
              style="--el-switch-on-color: #5a8a5a; margin-left: 8px;"
            />
          </div>
          <UnifiedMdEditor
            v-model="issueFixCombinedText"
            preview-theme="github"
            :preview="true"
            :toolbars="['preview', 'fullscreen']"
            class="task-workflow-issue-fix__editor"
          />
        </div>
      </div>
      <template #footer>
        <el-button @click="issueFixDialogVisible = false">关闭</el-button>
        <el-button type="primary" @click="copyIssueFixText">复制到剪贴板</el-button>
      </template>
    </el-dialog>

  </div>


    <!-- zcode 配置弹窗 -->
    <el-dialog
      v-model="zcodeConfigDialogVisible"
      title="zcode 配置"
      width="800px"
      destroy-on-close
    >
      <div style="margin-bottom: 16px;">
        <div style="font-size: 14px; font-weight: 500; margin-bottom: 8px; color: #303133;">zcode 工作目录地址</div>
        <div style="display: flex; gap: 8px;">
          <el-input
            v-model="zcodeDirInput"
            placeholder="例如: C:\Users\test\.zcode\v2\acp-config\claude"
            style="flex: 1;"
          />
          <el-button type="primary" :loading="zcodeSaving" @click="saveZcodeConfig">解析并保存</el-button>
          <el-button type="danger" plain :disabled="!zcodeProjectList.length" @click="deleteZcodeConfig">删除配置</el-button>
        </div>
      </div>
      <el-table v-if="zcodeProjectList.length > 0" :data="zcodeProjectList" border size="small" empty-text="暂无项目映射">
        <el-table-column prop="project_key" label="项目Key" width="200" />
        <el-table-column prop="workspace_path" label="工作目录" show-overflow-tooltip />
        <el-table-column prop="settings_path" label="Settings 配置文件" show-overflow-tooltip />
      </el-table>
      <div v-else style="color: #909399; text-align: center; padding: 20px;">暂无 zcode 项目映射</div>
      <template #footer>
        <el-button @click="zcodeConfigDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 执行任务弹窗 -->
    <el-dialog
      v-model="promptExecDialogVisible"
      title="执行任务"
      width="450px"
      destroy-on-close
    >
      <el-form label-width="80px">
        <el-form-item label="Agent">
          <el-select v-model="promptExecCliId" style="width: 100%;" placeholder="请选择 Agent 实例" @change="onPromptExecCliChange">
            <el-option
              v-for="cli in promptExecCliList"
              :key="cli.id"
              :label="cli.name + ' (' + cli.current_model + ')' + (cli.type === 'codex-cli' ? ' [Codex]' : '')"
              :value="cli.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="模型">
          <el-select v-model="promptExecModelName" style="width: 100%;" placeholder="请选择模型">
            <el-option
              v-for="modelName in promptExecModelOptions"
              :key="modelName"
              :label="modelName"
              :value="modelName"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="getSelectedCliType() !== 'codex'" label="思考强度">
          <el-select v-model="promptExecThinkingIntensity" style="width: 100%;" placeholder="请选择思考强度">
            <el-option label="低" value="低" />
            <el-option label="中等" value="中等" />
            <el-option label="高" value="高" />
            <el-option label="很高" value="很高" />
            <el-option label="最高" value="最高" />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="promptExecDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="promptExecLoading" @click="execPromptToClaude">
          开始执行
        </el-button>
      </template>
    </el-dialog>

    <!-- 执行历史弹窗（按 prompt_type） -->
    <ChatHistoryDialog
      ref="promptChatHistoryDialog"
      v-model="promptChatHistoryVisible"
      :title="promptChatHistoryPromptType ? '执行历史 - ' + promptChatHistoryTitle : '历史对话'"
      :loading="promptChatHistoryLoading"
      :items="promptChatHistoryList"
      :selected-id="promptChatDetailId"
      :detail-title="homeTask.name || '-'"
      :model-name="chatDetailModelName"
      :agent-name="chatDetailAgentName"
      :local-dir="chatDetailLocalDir"
      :thinking-intensity="chatDetailThinkingIntensity"
      :detail-status="chatDetailStatus"
      :detail-cli-type="chatDetailCliType"
      :detail-messages="chatDetailMessages"
      :last-usage-summary-data="chatDetailLastUsageSummary"
      :continue-input="chatContinueInput"
      :continue-loading="chatContinueLoading"
      :continue-disabled="isChatContinueDisabled()"
      :show-new-chat-button="true"
      :scroll-button-visible="promptChatDetailShowScrollBtn"
      :running-text="'等待 claude code 响应...'"
      :thinking-stream-elapsed="thinkingStreamElapsed"
      :item-msg-count-fn="getItemMsgCount"
      :runtime-duration-text-fn="runtimeDurationText"
      :format-duration-display-fn="formatDurationDisplay"
      :format-created-at-fn="formatCreatedAt"
      :render-markdown-fn="renderMarkdown"
      :is-current-thinking-fn="isCurrentThinking"
      :format-cli-type-fn="formatCliType"
      :is-long-text-fn="isLongText"
      :truncate-cmd-prompt-fn="truncateCmdPrompt"
      :stop-reason-label-fn="stopReasonLabel"
      :format-num-fn="formatNum"
      @select="onPromptChatRowClick"
      @update:continueInput="chatContinueInput = $event"
      @continue="continueChat"
      @new-chat="startNewChatFromHistory"
      @stop="stopChat"
      @scroll="onPromptChatDetailScroll"
      @scroll-to-bottom="scrollPromptChatToBottom(true)"
      @closed="onPromptChatHistoryClosed"
    >
      <template #before-input>
        <TaskProgressPanel @scroll-to-msg="onPromptTaskPanelScrollToMsg" />
      </template>
    </ChatHistoryDialog>

    <!-- 接口开发弹窗 -->
    <el-dialog
      v-model="apiDevDialogVisible"
      :title="apiDevDialogTitle"
      width="90%"
      top="2vh"
      destroy-on-close
      class="task-workflow-api-dev-dialog"
    >
      <div class="task-workflow-api-dev-dialog__body">
        <iframe
          v-if="apiDevDialogUrl"
          :src="apiDevDialogUrl"
          class="task-workflow-api-dev-dialog__iframe"
        />
        <div v-else class="task-workflow-api-dev-dialog__empty">
          暂无内容
        </div>
      </div>
    </el-dialog>

    <!-- 任务配置弹窗 -->
    <el-dialog
      v-model="taskConfigDialogVisible"
      title="任务配置"
      width="80%"
      top="3vh"
      destroy-on-close
      class="task-workflow-config-dialog"
    >
      <div class="task-workflow-config-content">
        <div class="task-workflow-config-section">
          <div class="task-workflow-config-section__title">基本信息</div>
          <el-descriptions :column="2" border size="small">
            <el-descriptions-item label="任务名称">{{ homeTask.name || '-' }}</el-descriptions-item>
            <el-descriptions-item label="任务状态">
              <el-tag size="small" effect="light" :type="getTaskStatusTagType(homeTask.task_status)">{{ homeTask.task_status || '-' }}</el-tag>
              <el-dropdown trigger="click" @command="handleTaskStatusChange" style="margin-left: 8px;">
                <el-button size="small" :loading="statusUpdating" text type="primary">
                  切换状态
                </el-button>
                <template #dropdown>
                  <el-dropdown-menu>
                    <el-dropdown-item
                      v-for="status in taskStatusOptions"
                      :key="status"
                      :command="status"
                      :disabled="homeTask.task_status === status"
                    >
                      {{ status }}
                    </el-dropdown-item>
                  </el-dropdown-menu>
                </template>
              </el-dropdown>
            </el-descriptions-item>
            <el-descriptions-item label="开始日期">{{ homeTask.start_time_desc || '-' }}</el-descriptions-item>
            <el-descriptions-item label="抓取类型">{{ requirementSourceName }}</el-descriptions-item>
            <el-descriptions-item :label="requirementSourceName + '地址'">
              <a v-if="requirementSourceUrl" :href="requirementSourceUrl" target="_blank" class="task-workflow-config-link">{{ requirementSourceUrl }}</a>
              <span v-else>-</span>
            </el-descriptions-item>
            <el-descriptions-item label="使用工作流程">{{ Number(homeTask.use_workflow || 0) === 1 ? '是' : '否' }}</el-descriptions-item>
            <el-descriptions-item label="最后操作">{{ homeTask.last_operated_at_desc || '-' }}</el-descriptions-item>
          </el-descriptions>
        </div>
        <div v-if="parsedTaskDevConfigs.length > 0" class="task-workflow-config-section">
          <div class="task-workflow-config-section__title">开发项目配置</div>
          <div v-for="(cfg, idx) in parsedTaskDevConfigs" :key="idx" class="task-workflow-config-dev">
            <div v-if="parsedTaskDevConfigs.length > 1" class="task-workflow-config-dev__index">配置 #{{ idx + 1 }}</div>
            <el-descriptions :column="2" border size="small">
              <el-descriptions-item label="Git仓库">{{ getTaskConfigName('git', cfg.git_id) }}</el-descriptions-item>
              <el-descriptions-item label="Docker">{{ getTaskConfigName('docker', cfg.docker_id) }}</el-descriptions-item>
              <el-descriptions-item label="Db">{{ getTaskConfigName('mysql', cfg.mysql_id) }}</el-descriptions-item>
              <el-descriptions-item label="接口集合"><span class="task-workflow-config-link" @click="openApiDevDialog(cfg)">{{ truncateWorkflowLabel(getTaskConfigApiLabel(cfg)) }}</span></el-descriptions-item>
              <el-descriptions-item label="本地目录">
                {{ cfg.local_dir || '-' }}
                <el-dropdown v-if="cfg.local_dir" trigger="click" @command="(editor) => openInEditor(cfg.local_dir, editor)" style="margin-left:6px">
                  <span class="task-workflow-config__open-editor-btn">▶ 打开方式</span>
                  <template #dropdown>
                    <el-dropdown-menu>
                      <el-dropdown-item command="vscode">VS Code</el-dropdown-item>
                      <el-dropdown-item command="cursor">Cursor</el-dropdown-item>
                      <el-dropdown-item command="goland">GoLand</el-dropdown-item>
                      <el-dropdown-item command="phpstorm">PhpStorm</el-dropdown-item>
                    </el-dropdown-menu>
                  </template>
                </el-dropdown>
              </el-descriptions-item>
              <el-descriptions-item label="父分支">{{ cfg.parent_branch || '-' }}</el-descriptions-item>
              <el-descriptions-item label="分支名">{{ truncateWorkflowLabel(cfg.branch_name || '-') }}</el-descriptions-item>
              <el-descriptions-item label="规则入口">{{ cfg.rule_entry_file || '-' }}</el-descriptions-item>
              <el-descriptions-item label="自定义网页">{{ getTaskConfigName('smart_link', cfg.smart_link_id) }}</el-descriptions-item>
              <el-descriptions-item label="账号">{{ cfg.smart_link_account || '-' }}</el-descriptions-item>
            </el-descriptions>
          </div>
        </div>
      </div>
      <template #footer>
        <el-button @click="taskConfigDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>
</template>

<script>
import { HomeFilled, View, Edit, Check, CopyDocument, MoreFilled, RefreshLeft, Link, Download, Search, Share, FullScreen } from '@element-plus/icons-vue'
import GitActionButton from '@/components/base/GitActionButton.vue'
import ChatHistoryButton from '@/components/shared/ChatHistoryButton.vue'
import ChatHistoryDialog from '@/components/shared/ChatHistoryDialog.vue'
import agentCliApi from '@/utils/base/agent_cli'
import MemoryFragmentApi from '@/utils/base/memory_fragment'
import taskWorkflowApi from '@/utils/base/task_workflow'
import homeTaskApi from '@/utils/base/home_task'
import baseUtils from '@/utils/base'
import sseDistribute from '@/utils/base/sse_distribute'
import sseBusiness from '@/utils/base/sse_business'
import chatParser from '@/utils/chat_parser'
import TaskProgressPanel from '@/components/TaskProgressPanel.vue'
import taskProgressStore from '@/utils/task_progress_store'
import gitApi from '@/utils/base/git'
import mysqlSetApi from '@/utils/base/mysql_set'
import apiManagement from '@/utils/base/api'
import dockerApi from '@/utils/base/compose'
import smartLinkSetApi from '@/utils/base/smart_link_set'
import set from '@/utils/base/git_set'
import MarkdownIt from 'markdown-it'
import { MdPreview } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import UnifiedMdEditor from '@/components/base/UnifiedMdEditor.vue'
import FileChangesDialog from '@/components/FileChangesDialog.vue'
import MemoryHistoryDialog from '@/components/memory/MemoryHistoryDialog.vue'

const PROMPT_EDITOR_TOOLBARS = [
  'bold', 'italic', 'strikeThrough', 'title', 'quote',
  'unorderedList', 'orderedList', 'task', 'link', 'code',
  'codeRow', 'table', 'preview', 'fullscreen',
]

// 节点状态常量
const NODE_STATUS_PENDING = 'pending'
const NODE_STATUS_RUNNING = 'running'
const NODE_STATUS_COMPLETED = 'completed'
const NODE_STATUS_SKIPPED = 'skipped'

// 节点状态选项（切换按钮循环顺序）
const NODE_STATUS_OPTIONS = [NODE_STATUS_PENDING, NODE_STATUS_RUNNING, NODE_STATUS_COMPLETED, NODE_STATUS_SKIPPED]

// 节点状态文案映射
const NODE_STATUS_LABELS = {
  [NODE_STATUS_PENDING]: '待执行',
  [NODE_STATUS_RUNNING]: '执行中',
  [NODE_STATUS_COMPLETED]: '已完成',
  [NODE_STATUS_SKIPPED]: '已跳过',
}

const ACTIVE_NODE_CACHE_PREFIX = 'task_workflow_active_node_'
const PROMPT_EXEC_CACHE_PREFIX = 'task_workflow_prompt_exec_'

const TASK_STATUS_TODO = '待开始'
const TASK_STATUS_DEVELOPING = '开发中'
const TASK_STATUS_DEV_COMPLETED = '开发完'
const TASK_STATUS_SELF_TESTING = '自测中'
const TASK_STATUS_SELF_TESTED = '自测完'
const TASK_STATUS_PENDING_INTEGRATION = '待对接'
const TASK_STATUS_INTEGRATING = '对接中'
const TASK_STATUS_TESTING = '测试中'
const TASK_STATUS_RELEASING = '上线中'
const TASK_STATUS_ONLINE = '已上线'
const TASK_STATUS_PENDING_TEST = '待测试'
const TASK_STATUS_ABANDONED = '已废弃'
const TASK_STATUS_OPTIONS = [
  TASK_STATUS_TODO,
  TASK_STATUS_DEVELOPING,
  TASK_STATUS_DEV_COMPLETED,
  TASK_STATUS_SELF_TESTING,
  TASK_STATUS_SELF_TESTED,
  TASK_STATUS_PENDING_INTEGRATION,
  TASK_STATUS_INTEGRATING,
  TASK_STATUS_TESTING,
  TASK_STATUS_PENDING_TEST,
  TASK_STATUS_RELEASING,
  TASK_STATUS_ONLINE,
  TASK_STATUS_ABANDONED,
]

const TASK_WORKFLOW_CONFIG_MAX_CHARS = 20

// 默认步骤节点（无模板时使用，desc 为空，由模板 remark 驱动）
const WORKFLOW_NODES = [
  { key: 'requirement-fetch', label: '抓取需求', desc: '' },
]

// markdown-it 实例，用于在"执行历史"对话框中渲染 markdown（包括表格）
const md = new MarkdownIt({ html: true, breaks: true, linkify: true })

export default {
  name: 'TaskWorkflow',
  components: {
    HomeFilled,
    View,
    Edit,
    Check,
    CopyDocument,
    MoreFilled,
    RefreshLeft,
    Link,
    Download,
    Search,
    Share,
    GitActionButton,
    ChatHistoryButton,
    ChatHistoryDialog,
    MemoryHistoryDialog,
    UnifiedMdEditor,
    FileChangesDialog,
    TaskProgressPanel,
    MdPreview,
  },
  data() {
    return {
      workflowNodes: [...WORKFLOW_NODES],
      _customStepPrompts: {}, // 自定义步骤的提示词暂存
      activeNode: 'requirement-fetch',
      loading: false,
      errorMessage: '',
      workflowId: 0,
      workflow: {},
      homeTask: {},
      requirementFragment: {},
      requirementShareUrl: '',
      requirementFetchConfig: {},
      requirementFetchLogs: [],
      requirementFetchRunning: false,
      requirementFetchAutoTriggered: false,
      workflowSseDistributeId: '',
      promptSaving: '',
      promptRestoring: '',
      apiDocResetting: false,
      requirementFetchActiveTab: 'tapd-fetch',
      requirementFetchConfigs: [],
      promptEditorToolbars: PROMPT_EDITOR_TOOLBARS,
      taskStatusOptions: TASK_STATUS_OPTIONS,
      statusUpdating: false,
      taskConfigGitRepoList: [],
      taskConfigDockerList: [],
      taskConfigMysqlList: [],
      taskConfigCollectionList: [],
      taskConfigSmartLinkList: [],
      taskConfigApiFolderMap: {},
      nodeStatuses: {},
      nodeStatusSaving: false,
      taskConfigDialogVisible: false,
      fragmentDialogVisible: false,
      fragmentDialogUrl: '',
      fragmentDialogTitle: '',
      fragmentDialogLoading: false,
      issueFixDialogVisible: false,
      issueFixInput: '',
      issueFixResolvedTemplate: '',
      issueFixUseDefaultPrompt: true, // 是否使用默认问题修改提示词
      // claude code 对话
      _chatHistoryDurationTimer: null, // 历史对话列表运行中对话的实时耗时定时器
      promptChatCounts: {},
      chatDetailId: 0,
      chatDetailPrompt: '',
      chatDetailSessionId: '',
      chatDetailStatus: '',
      chatDetailMessages: [],
      chatDetailLastUsageSummary: null,
      chatDetailSSELines: [], // SSE 累积的原始行
      chatDetailAutoScroll: true,
      _autoScrollLocked: false, // 程序化滚动锁
      _sseLineBuffer: [], // SSE 行缓冲（批处理），每100ms刷新一次
      _sseBatchTimer: null, // 批处理定时器
      _sseParseState: null, // 增量解析状态 { currentMessage, toolUseMap, pendingPatches }
      _continueInProgress: false, // continueChat 进行中标志，防止旧 chat_status_change SSE 覆盖本地 running 状态
      thinkingStreamElapsed: 0, // 思考流式阶段的实时已用秒数
      chatContinueInput: '',
      chatContinueLoading: false,
      // 执行任务
      promptExecDialogVisible: false,
      promptExecCliId: 0,
      promptExecCliList: [],
      promptExecModelName: '',
      promptExecLoading: false,
      promptExecPromptType: '',
      promptExecPromptValue: '',
      promptExecThinkingIntensity: '高',
      // 执行历史（按 prompt_type）
      promptChatHistoryVisible: false,
      promptChatHistoryTitle: '',
      promptChatHistoryList: [],
      promptChatHistoryLoading: false,
      promptChatHistoryPromptType: '',
      promptChatDetailId: 0,
      promptChatDetailShowScrollBtn: false,
      promptChatUnreadCounts: {},
      workflowUnreadCount: 0,
      _workflowUnreadSseId: '',
      _promptChatHistoryHideHandled: false,
      chatDetailModelName: '',
      chatDetailAgentName: '',
      chatDetailLocalDir: '',
      chatDetailThinkingIntensity: '',
      chatDetailCliType: 'claude',
      // zcode 配置
      zcodeConfigDialogVisible: false,
      zcodeDirInput: '',
      zcodeProjectList: [],
      zcodeSaving: false,
      issueFixZcodeMappings: [],
      // 接口开发弹窗
      apiDevDialogVisible: false,
      apiDevDialogUrl: '',
      apiDevDialogTitle: '',
      // 本地目录与分支状态检测
      localDirStatusMap: {},
      branchStatusMap: {},
      branchMismatchDialogVisible: false,
      branchMismatchDialogMode: 'notice',
      branchMismatchDialogResolver: null,
      branchMismatchLoading: false,
      branchMismatchDetailList: [],
      branchMismatchPromptedTaskId: 0,
      branchSwitchingKey: '',
      branchSwitchStreamDialogVisible: false,
      branchSwitchStreamRunning: false,
      branchSwitchStreamLines: [],
      branchSwitchStreamMeta: {
        local_dir: '',
        base_branch: '',
        branch_name: '',
      },
      _branchSwitchEventSource: null,
      // 远程分支状态检测
      remoteBranchStatusMap: {},
      remoteBranchDialogVisible: false,
      remoteBranchDialogItem: {},
      remoteBranchDialogLoading: false,
      remoteBranchPushing: false,
      remoteBranchDialogPushResult: null,
      // 文件变更检测
      fileChangesMap: {},
      _fileChangesPollTimer: null,
      fileChangesDialogVisible: false,
      fileChangesDetailLocalDir: '',
      fileChangesDetailParentBranch: '',
      fileChangesDetailInitialSummary: null,
      fileChangesDetailInitialFiles: [],
      // 工作流文档独立表相关
      workflowDocuments: [],
      _cachedTemplateSteps: [], // 缓存原始模板步骤数据（含 step_documents）
      // 步骤左侧Tab相关
      stepActiveTab: '__prompt__', // 当前激活的左侧Tab，'__prompt__' 为提示词，其他为文档id/name
      promptEditMode: true, // 提示词是否为编辑模式（false为查看模式）
      docEditMode: false, // 文档Tab是否为编辑模式（false为查看模式）
      stepDocContents: {}, // 文档内容缓存 { docKey: { content, loading, fileId } }
      docUnreadFlags: {}, // 文档未读红点标记 { docKey: true } — SSE 推送后置位，切换Tab时清除
      stepPromptTitle: '', // 提示词Tab标题（可编辑）
      stepDocTitle: '', // 文档Tab标题（可编辑）
      // 文档编辑/搜索/分享/历史
      docSearchBarVisible: false,
      docSearchQuery: '',
      docSearchMatchIndex: -1,
      docSearchMatchCount: 0,
      _docSearchMatches: null,
      docSharing: false,
      docHistoryVisible: false,
      docHistoryFragmentId: '',
      // 提示词预览目录导航相关
      promptTocItems: [], // 提示词预览目录项列表
      promptActiveHeading: '', // 提示词预览当前高亮的目录项 id
      promptScrollObserver: null, // 提示词预览滚动监听器
      // 文档预览目录导航相关
      docTocItems: [], // 文档预览目录项列表
      docActiveHeading: '', // 文档预览当前高亮的目录项 id
      docScrollObserver: null, // 文档预览滚动监听器
    }
  },
  computed: {
    taskId() {
      return Number(this.$route.params.taskId || 0)
    },
    requirementSourceType() {
      return String(this.homeTask.fetch_type || 'tapd').toLowerCase()
    },
    requirementSourceName() {
      const cfg = this.getFetchConfigByType(this.requirementSourceType)
      if (cfg) return cfg.name
      return this.requirementSourceType === 'zentao' ? '禅道' : 'TAPD'
    },
    requirementSourceUrl() {
      if (this.requirementSourceType === 'zentao') {
        return String(this.homeTask.zentao_url || '').trim()
      }
      return String(this.homeTask.tapd_url || '').trim()
    },
    requirementFetchStatus() {
      return String(this.workflow.requirement_fetch_status || 'idle').trim() || 'idle'
    },
    requirementFetchStatusText() {
      const map = {
        idle: '待执行',
        running: '执行中',
        success: '已完成',
        failed: '执行失败',
      }
      return map[this.requirementFetchStatus] || this.requirementFetchStatus
    },
    requirementFragmentId() {
      return String(this.workflow.requirement_fragment_id || '').trim()
    },
    parsedTaskDevConfigs() {
      const raw = this.homeTask.dev_configs
      if (!raw) return []
      try {
        const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
        return Array.isArray(parsed) ? parsed : []
      } catch {
        return []
      }
    },
    // 获取某个节点的状态，task-config 固定为 completed
    getNodeStatus() {
      return (nodeKey) => {
        if (nodeKey === 'task-config') return NODE_STATUS_COMPLETED
        return this.nodeStatuses[nodeKey] || NODE_STATUS_PENDING
      }
    },
    // 查找第一个"执行中"状态的节点 key
    firstRunningNodeKey() {
      for (const node of this.workflowNodes) {
        if (node.key !== 'task-config' && this.getNodeStatus(node.key) === NODE_STATUS_RUNNING) {
          return node.key
        }
      }
      return 'requirement-fetch'
    },
    issueFixCombinedText() {
      const input = (this.issueFixInput || '').trim()
      // 根据开关决定是否使用默认问题修改提示词模板
      const template = this.issueFixUseDefaultPrompt ? (this.issueFixResolvedTemplate || '').trim() : ''
      if (!input && !template) return ''
      if (!input) return template
      if (!template) return input
      return input + '\n\n' + template
    },
    // promptExecModelOptions 返回执行弹窗当前 Agent 卡片可选的模型列表。
    // promptExecModelOptions returns model options for the currently selected Agent card in the execution dialog.
    promptExecModelOptions() {
      return this.getSelectedCliModelOptions()
    },
    branchMismatchMatchedCount() {
      return this.branchMismatchDetailList.filter(item => item && item.matched).length
    },
    branchMismatchMismatchedCount() {
      return this.branchMismatchDetailList.filter(item => item && !item.matched).length
    },
    // 模板步骤原始数据映射（按 step_key 索引）
    templateStepsMap() {
      const map = {}
      for (const step of (this._cachedTemplateSteps || [])) {
        map[step.step_key] = step
      }
      return map
    },
    // 是否使用了工作流模板（有模板步骤数据）
    hasTemplate() {
      return this._cachedTemplateSteps && this._cachedTemplateSteps.length > 0
    },
    // 文档搜索是否有匹配结果
    docHasSearchMatches() {
      return this.docSearchMatchCount > 0
    },
    // 文档搜索结果摘要文本
    docSearchSummaryText() {
      if (!this.docSearchQuery) return ''
      if (this.docSearchMatchCount === 0) return '无匹配'
      return `${this.docSearchMatchIndex + 1}/${this.docSearchMatchCount}`
    },
  },
  mounted() {
    this.loadWorkflowPage()
    this.loadTaskConfigLookupData()
    this.loadTaskStatusList()
    this.ensureWorkflowUnreadSse()
    this.connectBusinessSse()
    this.registerFragmentUpdateSse()
    window.addEventListener('keydown', this.handleCtrlS, true)
  },
  activated() {
    this.ensureWorkflowUnreadSse()
    this.loadWorkflowPage()
    if (this.promptChatHistoryVisible) {
      this.loadPromptChatHistoryListSilently()
    } else {
      this.loadChatCounts()
    }
    this.startFileChangesPolling()
  },
  beforeUnmount() {
    window.removeEventListener('keydown', this.handleCtrlS, true)
    this._stopChatHistoryDurationTimer()
    if (this._sseBatchTimer) { clearTimeout(this._sseBatchTimer); this._sseBatchTimer = null }
    this.stopFileChangesPolling()
    this.unregisterWorkflowSse()
    this.unregisterChatOutputSse()
    this.unregisterFragmentUpdateSse()
    sseBusiness.CloseBusinessSse('task_workflow')
    this.unregisterWorkflowUnreadSse()
    this.cleanupPromptScrollObserver()
    this.cleanupDocScrollObserver()
  },
  watch: {
    promptChatHistoryVisible(val, oldVal) {
      if (val) {
        this._promptChatHistoryHideHandled = false
        return
      }
      if (oldVal) {
        this.handlePromptChatHistoryHide()
      }
    },
    parsedTaskDevConfigs: {
      handler(configs) {
        const seen = new Set()
        for (const cfg of configs) {
          const colId = Number(cfg.collection_id || 0)
          if (colId > 0 && !seen.has(colId)) {
            seen.add(colId)
            this.loadTaskConfigApiFoldersForCollection(colId)
          }
        }
      },
      immediate: true,
    },
    '$route.params.taskId'() {
      this.requirementFetchAutoTriggered = false
      this.requirementFetchLogs = []
      this.nodeStatuses = {}
      this.activeNode = 'requirement-fetch'
      this.requirementFetchActiveTab = 'tapd-fetch'
      this.branchMismatchDialogVisible = false
      this.branchMismatchDialogResolver = null
      this.branchMismatchDetailList = []
      this.branchMismatchPromptedTaskId = 0
      this.unregisterWorkflowSse()
      this.loadWorkflowPage()
    },
    // chatContinueInput 变更时实时缓存到 localStorage（按会话ID区分）
    chatContinueInput(newVal) {
      if (this.chatDetailId) {
        this.saveChatInputCache(this.chatDetailId, newVal)
      }
    },
    // 步骤切换时重置Tab状态
    activeNode(newVal) {
      this.stepActiveTab = '__prompt__'
      this.promptEditMode = true
      this.stepPromptTitle = this.getActiveNodeLabel()
      this.stepDocTitle = ''
      this.docEditMode = false
      // 清理目录导航状态和搜索状态
      this.clearDocSearch()
      this.cleanupPromptScrollObserver()
      this.cleanupDocScrollObserver()
      this.promptTocItems = []
      this.promptActiveHeading = ''
      this.docTocItems = []
      this.docActiveHeading = ''
    },
  },
  methods: {
    handleCtrlS(e) {
      if (!((e.ctrlKey || e.metaKey) && String(e.key || '').toLowerCase() === 's')) return
      e.preventDefault()
      const activeElement = document.activeElement
      const promptEditor = activeElement && typeof activeElement.closest === 'function'
        ? activeElement.closest('[data-prompt-type]')
        : null
      const promptTypeFromEditor = String(promptEditor?.dataset?.promptType || '').trim()
      if (promptTypeFromEditor) {
        this.savePrompts(promptTypeFromEditor)
        return
      }
      // 如果在文档编辑模式（编辑自定义文档），Ctrl+S 应保存文档而非提示词
      if (this.docEditMode && this.stepActiveTab !== '__prompt__') {
        this.saveActiveDoc()
        return
      }
      const nodeToPrompt = {}
      // 当前步骤即 promptType
      let promptType = this.activeNode
      if (this.activeNode === 'requirement-fetch') {
        // 抓取需求步骤固定显示知识片段，无提示词，不需要 Ctrl+S 保存
        return
      }
      if (promptType) {
        this.savePrompts(promptType)
      }
    },
    goBackToTaskList() {
      this.$router.push('/HomeTask')
    },
    goHome() {
      this.$router.push('/Dashboard')
    },
    reloadWorkflowPage() {
      this.requirementFetchAutoTriggered = false
      this.requirementFetchLogs = []
      this.requirementFetchActiveTab = 'tapd-fetch'
      this.nodeStatuses = {}
      this._customStepPrompts = {}
      this.loadWorkflowPage()
    },
    loadWorkflowPage() {
      if (this.taskId <= 0) {
        this.errorMessage = '任务 id 不合法'
        return
      }
      this.loading = true
      this.errorMessage = ''
      taskWorkflowApi.TaskWorkflowCreateOrGet(this.taskId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.loading = false
          this.errorMessage = response?.ErrMsg || '工作流加载失败'
          return
        }
        this.applyWorkflowPayload(response.Data)
        this.checkWorkflowLocalDirExists()
        this.checkWorkflowBranchStatus()
        this.checkWorkflowRemoteBranchStatus()
        this.startFileChangesPolling()
        this.activeNode = this.restoreActiveNodeCache() || this.firstRunningNodeKey
        this.loadRequirementFragment(() => {
          this.loading = false
          this.ensureWorkflowSse()
          this.maybeAutoFetchRequirement()
        })
        this.loadChatCounts()
      })
    },
    loadTaskStatusList() {
      homeTaskApi.TaskStatusList((response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          const list = Array.isArray(response.Data.list) ? response.Data.list : []
          if (list.length > 0) {
            this.taskStatusOptions = list.map(s => s.name)
          }
        }
      })
    },
    applyWorkflowPayload(data) {
      console.log('[DEBUG applyWorkflowPayload] 被调用, data.documents=', data.documents, 'type=', typeof data.documents, 'isArray=', Array.isArray(data.documents), 'length=', data.documents?.length, '调用栈:', new Error().stack)
      this.workflow = data.workflow || {}
      this.homeTask = data.home_task || this.homeTask || {}
      const previousWorkflowId = Number(this.workflowId || 0)
      this.workflowId = Number(this.workflow.id || 0)
      this.requirementFetchConfig = data.requirement_fetch_config || this.requirementFetchConfig || {}
      // 加载需求抓取配置列表（用于显示自定义名称）
      this.loadRequirementFetchConfigs()
      // 从模板步骤动态生成工作流节点列表
      if (data.template_steps && data.template_steps.length > 0) {
        this._cachedTemplateSteps = data.template_steps
        this.workflowNodes = data.template_steps
          .filter(step => step.step_key !== 'task-config')
          .map(step => ({
          key: step.step_key,
          label: step.name,
          desc: step.remark || '',
          _isFixed: step.is_fixed === 1,
          _promptContent: step.prompt_content || '',
        }))
      }
      this.parseNodeStatuses()
      // 捕获工作流文档独立表数据
      // 仅在响应中包含文档数据时才更新，避免并发请求返回空数组覆盖已加载的正确数据
      if (data.documents && Array.isArray(data.documents) && data.documents.length > 0) {
        console.log('[DEBUG applyWorkflowPayload] 设置 workflowDocuments, documents count=', data.documents.length)
        this.workflowDocuments = data.documents
      } else {
        console.log('[DEBUG applyWorkflowPayload] 跳过设置 workflowDocuments (data.documents 为空)')
      }
      if (this.workflowId !== previousWorkflowId) {
        this.unregisterWorkflowUnreadSse()
        this.ensureWorkflowUnreadSse()
      }
      document.title = this.homeTask.name || '任务工作流程'
    },
    // 解析后端返回的 node_statuses JSON 字符串
    parseNodeStatuses() {
      const raw = String(this.workflow.node_statuses || '').trim()
      if (!raw) {
        this.nodeStatuses = {}
        return
      }
      try {
        const parsed = typeof raw === 'string' ? JSON.parse(raw) : raw
        this.nodeStatuses = (parsed && typeof parsed === 'object') ? parsed : {}
      } catch {
        this.nodeStatuses = {}
      }
    },
    // 切换节点状态（循环：执行中 -> 已完成 -> 已跳过 -> 执行中）
    cycleNodeStatus(nodeKey) {
      if (nodeKey === 'task-config' || this.nodeStatusSaving) return
      const current = this.getNodeStatus(nodeKey)
      const currentIdx = NODE_STATUS_OPTIONS.indexOf(current)
      const nextIdx = (currentIdx + 1) % NODE_STATUS_OPTIONS.length
      const nextStatus = NODE_STATUS_OPTIONS[nextIdx]
      const updated = { ...this.nodeStatuses, [nodeKey]: nextStatus }
      this.saveNodeStatuses(updated)
    },
    // 保存节点状态到后端
    saveNodeStatuses(updatedStatuses) {
      if (this.workflowId <= 0) return
      this.nodeStatusSaving = true
      const jsonStr = JSON.stringify(updatedStatuses)
      taskWorkflowApi.TaskWorkflowNodeStatusUpdate(this.workflowId, jsonStr, (response) => {
        this.nodeStatusSaving = false
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '节点状态保存失败')
          return
        }
        this.nodeStatuses = updatedStatuses
      })
    },
    // 获取节点状态文案
    getNodeStatusLabel(nodeKey) {
      return NODE_STATUS_LABELS[this.getNodeStatus(nodeKey)] || '待执行'
    },
    loadRequirementFragment(done) {
      const fragmentId = this.requirementFragmentId
      if (!fragmentId) {
        this.requirementFragment = {}
        this.requirementShareUrl = ''
        if (typeof done === 'function') done()
        return
      }
      MemoryFragmentApi.MemoryFragmentInfo(fragmentId, (response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.requirementFragment = response.Data || {}
          this.refreshRequirementShareUrl()
        } else {
          this.errorMessage = response?.ErrMsg || '需求文档加载失败'
        }
        if (typeof done === 'function') done()
      })
    },
    refreshRequirementShareUrl() {
      const fragmentId = this.requirementFragmentId
      if (!fragmentId) {
        this.requirementShareUrl = ''
        return
      }
      MemoryFragmentApi.MemoryFragmentShareCreate(fragmentId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) {
          return
        }
        const token = String(response.Data.token || '').trim()
        if (!token) {
          this.requirementShareUrl = ''
          return
        }
        this.requirementShareUrl = baseUtils.BuildMemoryFragmentShareUrl(fragmentId, token, response.Data.url)
        this.replaceRequirementShareUrlPlaceholder()
      })
    },
    ensureWorkflowSse() {
      if (this.workflowId <= 0) {
        return
      }
      const nextDistributeId = `task_workflow_${this.workflowId}`
      if (this.workflowSseDistributeId === nextDistributeId) {
        return
      }
      this.unregisterWorkflowSse()
      sseDistribute.InitFromLoginStatus().then((created) => {
        if (!created && !sseDistribute.GetSseClientId()) {
          return
        }
        this.workflowSseDistributeId = nextDistributeId
        sseDistribute.RegisterReceive(nextDistributeId, this.handleWorkflowSseMessage)
      })
    },
    unregisterWorkflowSse() {
      if (!this.workflowSseDistributeId) {
        return
      }
      sseDistribute.UnRegisterReceive(this.workflowSseDistributeId)
      this.workflowSseDistributeId = ''
    },
    handleWorkflowSseMessage(data) {
      if (!data || Number(data.workflow_id || 0) !== this.workflowId) {
        return
      }
      // chat 状态变更时刷新执行历史按钮的计数和动画
      if (data.type === 'chat_status_change' || data.type === 'chat_read_change') {
        this.applyWorkflowChatListSnapshot(data.chat_list || [])
        return
      }
      // 节点状态变更时直接更新本地 nodeStatuses，无需重新请求接口
      if (data.type === 'node_status_change') {
        try {
          const parsed = data.node_statuses ? JSON.parse(data.node_statuses) : {}
          this.nodeStatuses = parsed
        } catch (e) {
          // 解析失败忽略，保留当前状态
        }
        return
      }
      // 提示词内容变更时更新本地提示词
      if (data.type === 'step_prompt_change') {
        const stepKey = String(data.step_key || '').trim()
        const promptContent = String(data.prompt_content || '').trim()
        if (stepKey) {
          // 清除自定义缓存，让 getStepPrompt 从 workflow.step_prompts 重新读取
          if (this._customStepPrompts) {
            delete this._customStepPrompts[stepKey]
          }
          // 同步更新 workflow.step_prompts
          if (this.workflow.step_prompts) {
            try {
              const prompts = typeof this.workflow.step_prompts === 'string'
                ? JSON.parse(this.workflow.step_prompts) : this.workflow.step_prompts
              prompts[stepKey] = promptContent
              this.workflow.step_prompts = typeof this.workflow.step_prompts === 'string'
                ? JSON.stringify(prompts) : prompts
            } catch (e) { /* ignore */ }
          }
        }
        return
      }
      this.requirementFetchLogs.push({
        workflow_id: Number(data.workflow_id || 0),
        step: String(data.step || '').trim(),
        status: String(data.status || '').trim(),
        message: String(data.message || '').trim(),
        time: Number(data.time || 0),
      })
    },
    // 知识片段变更 SSE 回调：若变更的 fragment 正在步骤文档Tab中展示，则重新加载内容
    handleMemoryFragmentUpdateForWorkflow(payload) {
      const fragmentId = String(payload?.fragment_id || payload?.fragment?.id || payload?.fragment?.file_id || '').trim()
      if (!fragmentId) return
      // 查找该 fileId 对应的文档 key（用于未读红点标记）
      const docKey = this.findDocKeyByFileId(fragmentId)
      this.reloadDocContentByFileId(fragmentId)
      // 如果不是当前正在查看的文档，则标记未读红点
      if (docKey && docKey !== this.stepActiveTab) {
        this.docUnreadFlags = { ...this.docUnreadFlags, [docKey]: true }
      }
    },
    maybeAutoFetchRequirement() {
      if (this.requirementFetchAutoTriggered) {
        return
      }
      if (!this.requirementSourceUrl) {
        return
      }
      if (this.requirementFetchStatus === 'success') {
        return
      }
      if (this.requirementFetchStatus === 'running') {
        this.requirementFetchRunning = true
        return
      }
      if (this.requirementFetchStatus !== 'idle') {
        return
      }
      this.requirementFetchAutoTriggered = true
      this.triggerRequirementFetch(true)
    },
    // 根据 fetch_type 查找对应的需求抓取配置
    getFetchConfigByType(type) {
      if (!type) return null
      return this.requirementFetchConfigs.find(c => c.type === type) || null
    },
    // 加载需求抓取配置列表
    loadRequirementFetchConfigs() {
      if (this.requirementFetchConfigs.length > 0) return
      set.HomeTaskConfigGet((response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          const configs = response.Data.home_task_requirement_fetch_configs
          this.requirementFetchConfigs = Array.isArray(configs) ? [...configs] : []
        }
      })
    },
    triggerRequirementFetch(isAuto) {
      if (this.workflowId <= 0 || this.requirementFetchRunning) {
        return
      }
      if (!this.requirementSourceUrl) {
        this.$helperNotify.error(`当前任务未配置 ${this.requirementSourceName} 地址`)
        return
      }
      if (!isAuto) {
        this.requirementFetchAutoTriggered = true
      }
      this.requirementFetchRunning = true
      this.errorMessage = ''
      if (!isAuto) {
        this.requirementFetchLogs.push({
          step: 'manual',
          status: 'running',
          message: '手动触发重新抓取',
          time: Math.floor(Date.now() / 1000),
        })
      }
      taskWorkflowApi.TaskWorkflowRequirementFetch(this.workflowId, (response) => {
        this.requirementFetchRunning = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.errorMessage = response?.ErrMsg || `抓取${this.requirementSourceName}需求失败`
          this.$helperNotify.error(this.errorMessage)
          this.loadWorkflowPage()
          return
        }
        this.$helperNotify.success(`${this.requirementSourceName} 需求已抓取并写入知识片段`)
        this.applyWorkflowPayload({
          workflow: response.Data.workflow || {},
          home_task: this.homeTask,
          requirement_fetch_config: response.Data.requirement_fetch_config || {},
        })
        this.loadRequirementFragment(() => {})
      })
    },
    openRequirementFragment() {
      if (!this.requirementFragmentId) {
        this.$helperNotify.error('当前工作流未绑定需求知识片段')
        return
      }
      const routeInfo = this.$router.resolve({
        path: '/MemoryFragment',
        query: {
          fragment_id: this.requirementFragmentId,
          hide_menu: '1',
        },
      })
      window.open(routeInfo.href, '_blank')
    },
    openFragmentById(fragmentId) {
      if (!fragmentId) return
      const routeInfo = this.$router.resolve({
        path: '/MemoryFragment',
        query: {
          fragment_id: fragmentId,
          hide_menu: '1',
        },
      })
      window.open(routeInfo.href, '_blank')
    },
    resetApiDoc() {
      if (this.apiDocResetting || this.workflowId <= 0) return
      this.apiDocResetting = true
      const _this = this
      taskWorkflowApi.TaskWorkflowApiDocReset(this.workflowId, this.activeNode, function (res) {
        _this.apiDocResetting = false
        if (res.ErrCode !== 0) {
          _this.$message.error(res.ErrMsg || '重置接口文档失败')
          return
        }
        _this.$message.success('接口文档已重置')
      })
    },
    savePrompts(promptType) {
      if (this.promptSaving || this.workflowId <= 0) {
        return
      }
      this.promptSaving = promptType
      const savePayload = {
        workflow_id: this.workflowId,
        step_key: promptType,
        step_prompt: this.getStepPrompt(promptType),
      }
      taskWorkflowApi.TaskWorkflowPromptsSave(savePayload, (response) => {
        this.promptSaving = ''
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '提示词保存失败')
          return
        }
        this.$helperNotify.success('提示词已保存')
        if (response.Data?.workflow) {
          this.workflow = { ...this.workflow, ...response.Data.workflow }
        }
      })
    },
    restorePrompts(promptType) {
      if (this.promptRestoring || this.workflowId <= 0) {
        return
      }
      this.$confirm('确定要还原为默认提示词吗？当前内容将被覆盖。', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        this.doRestorePrompts(promptType)
      }).catch(() => {})
    },
    doRestorePrompts(promptType) {
      this.promptRestoring = promptType
      taskWorkflowApi.TaskWorkflowPromptsRestore(this.workflowId, (response) => {
        this.promptRestoring = ''
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '还原提示词失败')
          return
        }
        this.$helperNotify.success('提示词已还原为默认值')
        // 清空编辑器缓存，让 getStepPrompt 从 workflow.step_prompts 重新读取
        this._customStepPrompts = {}
        if (response.Data?.workflow) {
          this.workflow = response.Data.workflow
          this.workflowDocuments = response.Data.documents || []
        }
      })
    },
    // getStepPrompt 根据 step_key 获取当前步骤的提示词内容。
    getStepPrompt(stepKey) {
      if (!stepKey) return ''
      // 优先从编辑器缓存读取（setStepPrompt 写入的值）
      if (this._customStepPrompts && this._customStepPrompts[stepKey] !== undefined) {
        return this._customStepPrompts[stepKey] || ''
      }
      // 从 step_prompts 读取
      if (this.workflow.step_prompts) {
        try {
          const prompts = typeof this.workflow.step_prompts === 'string'
            ? JSON.parse(this.workflow.step_prompts)
            : this.workflow.step_prompts
          if (prompts[stepKey] !== undefined) {
            return prompts[stepKey] || ''
          }
        } catch (e) { /* ignore */ }
      }
      return ''
    },
    // setStepPrompt 设置自定义步骤的提示词值（暂存到组件实例）。
    setStepPrompt(stepKey, value) {
      if (!stepKey) return
      if (!this._customStepPrompts) {
        this._customStepPrompts = {}
      }
      this._customStepPrompts[stepKey] = value
    },
    // isCustomStepNode 判断当前激活节点是否为自定义步骤。
    isCustomStepNode(key) {
      const knownKeys = ['requirement-fetch', 'requirement', 'design', 'api-dev', 'api-test-fix', 'code-review', 'browser-test', 'issue_fix', 'plain_text_requirement', 'design_plan_requirement']
      return key && !knownKeys.includes(key)
    },
    // getActiveNodeLabel 获取当前激活节点的显示名称。
    getActiveNodeLabel() {
      const node = this.workflowNodes.find(n => n.key === this.activeNode)
      return node ? node.label : this.activeNode
    },
    // getActiveNodeDesc 获取当前激活节点的描述。
    getActiveNodeDesc() {
      const node = this.workflowNodes.find(n => n.key === this.activeNode)
      return node ? (node.desc || '') : ''
    },
    copyText(text, successMessage) {
      const value = String(text || '').trim()
      if (!value) {
        this.$helperNotify.error('没有可复制的内容')
        return
      }
      if (navigator.clipboard && navigator.clipboard.writeText) {
        navigator.clipboard.writeText(value).then(() => {
          this.$helperNotify.success(successMessage)
        }).catch(() => {
          this.fallbackCopyText(value, successMessage)
        })
        return
      }
      this.fallbackCopyText(value, successMessage)
    },
    fallbackCopyText(text, successMessage) {
      const textArea = document.createElement('textarea')
      textArea.value = text
      textArea.style.position = 'fixed'
      textArea.style.left = '-999999px'
      textArea.style.top = '-999999px'
      document.body.appendChild(textArea)
      textArea.focus()
      textArea.select()
      try {
        document.execCommand('copy')
        this.$helperNotify.success(successMessage)
      } catch (error) {
        this.$helperNotify.error('复制失败')
      }
      document.body.removeChild(textArea)
    },
    openIssueFixDialog() {
      this.issueFixDialogVisible = true
      this.issueFixInput = ''
      this.issueFixResolvedTemplate = ''
      this.issueFixUseDefaultPrompt = true
      this.issueFixZcodeMappings = []
      // 获取当前任务所有本地目录
      const taskDirs = this.parsedTaskDevConfigs
        .map(cfg => (cfg.local_dir || '').trim())
        .filter(Boolean)
      if (this.workflowId <= 0) return
      taskWorkflowApi.TaskWorkflowIssueFixResolve(this.workflowId, (response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.issueFixResolvedTemplate = response.Data.prompt || ''
        }
      })
      // 只显示与当前任务本地目录匹配的 zcode 配置
      taskWorkflowApi.TaskWorkflowZcodeGet((res) => {
        if (res && res.ErrCode === 0 && res.Data) {
          const allProjects = res.Data.projects || []
          const workspaceSet = new Set(taskDirs)
          this.issueFixZcodeMappings = allProjects.filter(p => workspaceSet.has(p.workspace_path))
        }
      })
    },
    copyIssueFixText() {
      this.copyText(this.issueFixCombinedText, '已复制到剪贴板')
    },
    // zcode 配置弹窗
    openZcodeConfigDialog() {
      this.zcodeConfigDialogVisible = true
      this.zcodeDirInput = ''
      this.zcodeProjectList = []
      taskWorkflowApi.TaskWorkflowZcodeGet((res) => {
        if (res && res.ErrCode === 0 && res.Data) {
          this.zcodeDirInput = res.Data.zcode_dir || ''
          this.zcodeProjectList = res.Data.projects || []
        }
      })
    },
    saveZcodeConfig() {
      const dir = (this.zcodeDirInput || '').trim()
      if (!dir) {
        this.$helperNotify.warning('请输入 zcode 工作目录地址')
        return
      }
      this.zcodeSaving = true
      taskWorkflowApi.TaskWorkflowZcodeSave(dir, (res) => {
        this.zcodeSaving = false
        if (res && res.ErrCode === 0 && res.Data) {
          this.$helperNotify.success('zcode 配置已保存')
          this.zcodeDirInput = res.Data.zcode_dir || ''
          this.zcodeProjectList = res.Data.projects || []
        }
      })
    },
    deleteZcodeConfig() {
      this.$confirm('确定要删除 zcode 配置吗？关联的项目映射也会被清空。', '提示', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        taskWorkflowApi.TaskWorkflowZcodeDelete((res) => {
          if (res && res.ErrCode === 0) {
            this.$helperNotify.success('zcode 配置已删除')
            this.zcodeDirInput = ''
            this.zcodeProjectList = []
          }
        })
      }).catch(() => {})
    },
    // 加载对话计数（按钮上显示）
    loadChatCounts() {
      if (this.workflowId <= 0) return
      taskWorkflowApi.TaskWorkflowChatList(this.workflowId, (res) => {
        if (res.ErrCode === 0 && res.Data) {
          this.applyWorkflowChatListSnapshot(res.Data.list || [])
        }
      })
    },
    applyWorkflowChatListSnapshot(list) {
      const normalizedList = Array.isArray(list) ? list : []
      this.updateChatCountsFromList(normalizedList)
      if (this.promptChatHistoryVisible) {
        const promptType = String(this.promptChatHistoryPromptType || '').trim()
        this.promptChatHistoryList = promptType
          ? normalizedList.filter(item => String(item.prompt_type || '').trim() === promptType)
          : normalizedList.slice()
      }
      if (this.chatDetailId > 0) {
        const current = normalizedList.find(item => Number(item.id || 0) === Number(this.chatDetailId || 0))
        // 正在进行 continue 操作时，不覆盖本地已设置的 running 状态，
        // 避免 chat_status_change SSE 中携带的旧状态（interrupted）覆盖掉 running
        if (current && !this._continueInProgress) {
          this.chatDetailStatus = current.status || this.chatDetailStatus
        }
      }
    },
    // 打开历史对话弹窗（复用执行历史弹窗，查全部对话）
    openChatHistoryDialog() {
      this.openPromptChatHistory('issue_fix')
    },
    updateChatCountsFromList(list) {
      const byType = {}
      for (const item of list) {
        const pt = item.prompt_type || ''
        if (pt) {
          const c = byType[pt] || { running: 0, interrupted: 0, total: 0, unread: 0 }
          c.total++
          if (item.status === 'running') c.running++
          else if (item.status === 'interrupted') c.interrupted++
          if (item.is_read === false && item.status !== 'running') {
            c.unread++
          }
          byType[pt] = c
        }
      }
      this.promptChatCounts = byType
    },
    adjustPromptUnreadCount(promptType, delta) {
      const normalizedPromptType = String(promptType || '').trim()
      if (!normalizedPromptType || !delta) return
      const currentCounts = this.getPromptChatCounts(normalizedPromptType)
      const nextUnread = Math.max(0, Number(currentCounts.unread || 0) + delta)
      this.promptChatCounts = {
        ...this.promptChatCounts,
        [normalizedPromptType]: {
          ...currentCounts,
          unread: nextUnread,
        },
      }
      const nextUnreadMap = { ...this.promptChatUnreadCounts }
      if (nextUnread > 0) {
        nextUnreadMap[normalizedPromptType] = nextUnread
      } else {
        delete nextUnreadMap[normalizedPromptType]
      }
      this.promptChatUnreadCounts = nextUnreadMap
      this.workflowUnreadCount = Math.max(0, Number(this.workflowUnreadCount || 0) + delta)
    },
    markPromptChatReadLocally(chatId) {
      const item = this.promptChatHistoryList.find(row => Number(row.id || 0) === Number(chatId || 0))
      if (!item || item.is_read !== false || item.status === 'running') return
      item.is_read = true
      this.promptChatHistoryList = this.promptChatHistoryList.slice()
      this.adjustPromptUnreadCount(item.prompt_type, -1)
    },
    markPromptChatReadOnServer(chatId) {
      const normalizedChatId = Number(chatId || 0)
      if (normalizedChatId <= 0) return
      agentCliApi.AgentChatMarkRead(normalizedChatId, (res) => {
        if (res && res.ErrCode === 0) {
          this.markPromptChatReadLocally(normalizedChatId)
        }
      })
    },
    markPromptChatRunningLocally(promptType, chatId, extra = {}) {
      const normalizedPromptType = String(promptType || '').trim()
      const normalizedChatId = Number(chatId || 0)
      if (!normalizedPromptType || normalizedChatId <= 0) return
      const currentCounts = this.getPromptChatCounts(normalizedPromptType)
      this.promptChatCounts = {
        ...this.promptChatCounts,
        [normalizedPromptType]: {
          running: currentCounts.running + 1,
          interrupted: currentCounts.interrupted,
          total: Math.max(currentCounts.total + 1, currentCounts.running + currentCounts.interrupted + 1),
          unread: currentCounts.unread,
        },
      }
      if (!this.promptChatHistoryVisible || this.promptChatHistoryPromptType !== normalizedPromptType) return
      const existing = this.promptChatHistoryList.find(item => Number(item.id || 0) === normalizedChatId)
      if (existing) {
        existing.status = 'running'
        if (existing.is_read === false) {
          existing.is_read = true
          this.adjustPromptUnreadCount(normalizedPromptType, -1)
        }
        this.promptChatHistoryList = this.promptChatHistoryList.slice()
        return
      }
      this.promptChatHistoryList = [{
        id: normalizedChatId,
        prompt_type: normalizedPromptType,
        status: 'running',
        is_read: true,
        line_count: 0,
        created_at: new Date().toISOString(),
        prompt: extra.prompt || '',
        model_name: extra.modelName || '',
        agent_cli_name: extra.agentName || '',
        local_dir: extra.localDir || '',
        thinking_intensity: extra.thinkingIntensity || '',
        cli_type: extra.cliType || 'claude',
      }, ...this.promptChatHistoryList]
    },
    // 加载对话详情
    loadChatDetail() {
      if (!this.chatDetailId) return
      console.log('[loadChatDetail] 开始加载对话详情 chatId=', this.chatDetailId, '当前本地状态=', this.chatDetailStatus)
      console.trace('[loadChatDetail] 调用栈:')
      taskWorkflowApi.TaskWorkflowChatDetail(this.chatDetailId, (res) => {
        // loadChatDetail 完成后清除 continue 保护标志，让后续 chat_status_change 正常同步状态
        this._continueInProgress = false
        if (res.ErrCode === 0 && res.Data) {
          const data = res.Data
          const oldStatus = this.chatDetailStatus
          this.chatDetailPrompt = data.prompt || ''
          this.chatDetailSessionId = data.session_id || ''
          this.chatDetailStatus = data.status || ''
          this.chatDetailModelName = data.model_name || ''
          this.chatDetailAgentName = data.agent_cli_name || ''
          this.chatDetailLocalDir = data.local_dir || ''
          this.chatDetailThinkingIntensity = data.thinking_intensity || ''
          this.chatDetailLastUsageSummary = data.last_usage_summary || null
          this.chatDetailCliType = data.cli_type || 'claude'
          console.log('[loadChatDetail] 状态变化: ' + oldStatus + ' -> ' + this.chatDetailStatus + ' (来自服务端status=' + data.status + ')')
          // 同步更新左侧列表中的状态
          this.updateChatListStatus(this.chatDetailId, this.chatDetailStatus)
          // 合并历史行 + SSE 加载期间收到的新行（有则去重）
          const historicalLines = data.lines || []
          const sseLines = this.chatDetailSSELines
          const newSseLines = sseLines.filter(l => !historicalLines.includes(l))
          this.chatDetailSSELines = [...historicalLines, ...newSseLines]
          this.chatDetailMessages = chatParser.parseChatLines(this.chatDetailSSELines, this.chatDetailCliType)
          // 历史对话：自动折叠所有思考（对话已完成或已结束）
          this.chatDetailMessages.forEach(msg => {
            if (msg.type === 'assistant' && msg.thinking) {
              msg._thinkingCollapsed = true
            }
            if (msg.type === 'assistant_thinking') {
              msg._thinkingCollapsed = true
            }
          })
          this.$nextTick(() => { this.scrollPromptChatToBottom(true) })
          if (this.chatDetailStatus === 'running') {
            this._flushSseBatch()
            this._sseParseState = null
            this._sseLineBuffer = []
          }
        }
      })
    },
    connectBusinessSse() {
      sseBusiness.fetchAvailableSsePort().then(port => {
        if (!port) return
        const clientId = baseUtils.GenerateSseClientId('work_flow')
        sseBusiness.ConnectBusinessSse('task_workflow', port, clientId)
        this.registerChatOutputSse()
      })
    },
    registerChatOutputSse() {
      this._chatOutputHandler = (data) => {
        if (!data || data.line === undefined || data.chat_id === undefined) return
        const chatId = Number(data.chat_id)
        const line = data.line
        if (chatId === Number(this.chatDetailId || 0)) {
          this._processChatSseLine(line)
        }
      }
      sseBusiness.RegisterBusinessReceive('task_workflow', 'task_workflow_chat_output', this._chatOutputHandler)
    },
    unregisterChatOutputSse() {
      if (this._chatOutputHandler) {
        sseBusiness.UnRegisterBusinessReceive('task_workflow', 'task_workflow_chat_output', this._chatOutputHandler)
        this._chatOutputHandler = null
      }
    },
    // 注册知识片段变更 SSE，用于文档Tab内容实时更新
    registerFragmentUpdateSse() {
      sseDistribute.RegisterReceive('memory_fragment_updates', this.handleMemoryFragmentUpdateForWorkflow)
    },
    // 注销知识片段变更 SSE
    unregisterFragmentUpdateSse() {
      sseDistribute.UnRegisterReceive('memory_fragment_updates', this.handleMemoryFragmentUpdateForWorkflow)
    },
    _processChatSseLine(line) {
      if (!line) return
      if (!this._sseParseState) {
        this._sseParseState = this.chatDetailCliType === 'codex'
          ? { currentItems: new Map(), pendingPatches: [] }
          : { currentMessage: null, toolUseMap: new Map(), pendingPatches: [] }
        this._sseLineBuffer = []
        this._thinkingStreamStartTime = 0
        if (!this._thinkingTimer) {
          this.thinkingStreamElapsed = 0
          this._thinkingTimer = setInterval(() => {
            if (this._thinkingStreamStartTime > 0) {
              this.thinkingStreamElapsed = Math.floor((Date.now() - this._thinkingStreamStartTime) / 1000)
            } else {
              this.thinkingStreamElapsed = 0
            }
          }, 200)
        }
      }
      try {
        const obj = JSON.parse(line)
        if (obj.type === 'chat' && obj.subtype === 'completed') {
          this._flushSseBatch()
          this.chatDetailSSELines.push(line)
          this._sseParseState = null
          // 当前正在查看该对话，自动标记为已读
          if (obj.chat_id === Number(this.chatDetailId || 0)) {
            this.markPromptChatReadOnServer(obj.chat_id)
          }
          this.loadChatDetail()
          this.loadChatCounts()
          this.$nextTick(() => { this.scrollPromptChatToBottom() })
          return
        }
      } catch (e) {}
      this._sseLineBuffer.push(line)
      if (!this._sseBatchTimer) {
        this._sseBatchTimer = setTimeout(() => { this._flushSseBatch() }, 100)
      }
    },
    // _flushSseBatch 将缓冲区中的 SSE 行批量增量解析并追加到消息列表。
    _flushSseBatch() {
      if (this._sseBatchTimer) {
        clearTimeout(this._sseBatchTimer)
        this._sseBatchTimer = null
      }
      const newLines = this._sseLineBuffer.splice(0)
      if (newLines.length === 0) return
      for (const l of newLines) {
        this.chatDetailSSELines.push(l)
      }
      const result = chatParser.parseChatLinesIncremental(newLines, this._sseParseState, this.chatDetailMessages.length, this.chatDetailCliType)
      this._sseParseState = result.parseState
      if (result.newMessages.length > 0) {
        this._autoScrollLocked = true
        for (const msg of result.newMessages) {
          this.chatDetailMessages.push(msg)
        }
      }
      for (const patch of result.parseState.pendingPatches) {
        for (let i = this.chatDetailMessages.length - 1; i >= 0; i--) {
          const msg = this.chatDetailMessages[i]
          if (msg.type === 'assistant') {
            for (const block of (msg.content || [])) {
              if (block.type === 'tool_use' && block.id === patch.blockId) {
                block._result = patch.resultData
              }
            }
          } else if (msg.type === 'tool_use' && msg.id === patch.blockId) {
            msg._result = patch.resultData
          }
        }
      }
      result.parseState.pendingPatches.length = 0
      if (result.newMessages.length > 0) {
        this.$nextTick(() => {
          this.scrollPromptChatToBottom()
          const boxes = document.querySelectorAll('.thinking-blockquote')
          boxes.forEach(box => { box.scrollTop = box.scrollHeight })
          requestAnimationFrame(() => {
            requestAnimationFrame(() => {
              this._autoScrollLocked = false
            })
          })
        })
      }
    },
    // 切换思考过程的折叠/展开
    toggleThinkingCollapse(msg) {
      msg._thinkingCollapsed = !msg._thinkingCollapsed
      msg._thinkingManuallyToggled = true
    },
    needCollapseBtn(text) {
      return (text || '').split('\n').length > 10
    },
    // system_command 气泡预览：优先展示 cmdLine（> 格式），否则展示 text
    
    formatCliType(cliType) {
      if (!cliType) return '提示词'
      return cliType.charAt(0).toUpperCase() + cliType.slice(1)
    },
    displayCmdPreview(msg) {
      const source = msg.cmdLine || msg.text || ''
      const preview = this.truncateUtf8(source, 20)
      return msg.cmdLine ? '> ' + preview : preview
    },
    // 判断文本 UTF-8 字节长度是否超过指定值
    isLongText(text, maxBytes) {
      if (!text) return false
      return new TextEncoder().encode(text).length > maxBytes
    },
    // UTF-8 安全截取：截取前 maxBytes 字节并追加 "..."
    truncateUtf8(text, maxBytes) {
      if (!text) return ''
      const bytes = new TextEncoder().encode(text)
      if (bytes.length <= maxBytes) return text
      let end = maxBytes
      while (end > 0 && (bytes[end] & 0xc0) === 0x80) {
        end--
      }
      return new TextDecoder().decode(bytes.slice(0, end)) + '...'
    },
    // 截取命令行中 -p / exec / --json 后的提示词内容
    truncateCmdPrompt(cmdLine, maxLen) {
      if (!cmdLine) return ''
      return cmdLine.replace(/(-p |exec |--json )"([^"]+)"/, (full, prefix, prompt) => {
        const bytes = new TextEncoder().encode(prompt)
        if (bytes.length <= maxLen) return full
        let end = maxLen
        while (end > 0 && (bytes[end] & 0xc0) === 0x80) end--
        return prefix + '"' + new TextDecoder().decode(bytes.slice(0, end)) + '..."'
      })
    },
    // 判断当前消息是否正在思考中（实时流式阶段）
    isCurrentThinking(msg) {
      const timing = msg && msg._thinkingTiming ? msg._thinkingTiming : null
      if (!timing || !timing.startMs || timing.durationMs > 0) return false
      for (let i = this.chatDetailMessages.length - 1; i >= 0; i--) {
        const m = this.chatDetailMessages[i]
        if (m.type === 'assistant' && m.thinking) {
          return m === msg
        }
      }
      return false
    },
    // 关闭对话详情（彻底断开 SSE 并清空状态）
    closeChatDetail() {
      if (this._sseBatchTimer) { clearTimeout(this._sseBatchTimer); this._sseBatchTimer = null }
      this._sseLineBuffer = []
      this._sseParseState = null
      if (this._thinkingTimer) { clearInterval(this._thinkingTimer); this._thinkingTimer = null }
      this.thinkingStreamElapsed = 0
      this._thinkingStreamStartTime = 0
      this.chatDetailMessages = []
      this.chatDetailSSELines = []
      taskProgressStore.reset()
      this.chatDetailId = 0
      this.chatContinueInput = ''
    },
    sendToClaudeCode() {
      const prompt = this.issueFixCombinedText
      if (!prompt || !prompt.trim()) {
        this.$helperNotify.warning('提示词为空，无法发送')
        return
      }
      this.issueFixDialogVisible = false
      this.openPromptExecDialog('issue_fix', prompt)
    },
    // 继续对话
    // isChatContinueDisabled 统一发送区按钮可用状态，确保"发送/新对话"行为一致。 // Keeps the action buttons under the same enabled-state rule.
    isChatContinueDisabled() {
      return this.chatContinueLoading || !String(this.chatContinueInput || '').trim()
    },
    // ===== 对话输入框 localStorage 缓存 =====
    // 缓存键名格式：dtool_chat_input_{chatId}
    getChatInputCacheKey(chatId) {
      return 'dtool_chat_input_' + (Number(chatId) || 0)
    },
    saveChatInputCache(chatId, value) {
      try {
        const key = this.getChatInputCacheKey(chatId)
        const val = String(value || '').trim()
        if (val) {
          localStorage.setItem(key, val)
        } else {
          // 空值时清除该缓存，避免残留垃圾数据
          localStorage.removeItem(key)
        }
      } catch (_) {
        // localStorage 不可用时静默忽略
      }
    },
    getChatInputCache(chatId) {
      try {
        const key = this.getChatInputCacheKey(chatId)
        return localStorage.getItem(key) || ''
      } catch (_) {
        return ''
      }
    },
    clearChatInputCache(chatId) {
      try {
        const key = this.getChatInputCacheKey(chatId)
        localStorage.removeItem(key)
      } catch (_) {
        // localStorage 不可用时静默忽略
      }
    },
    continueChat() {
      const input = this.chatContinueInput.trim()
      if (!input) return
      console.log('[continueChat] ========== 开始继续对话 ==========')
      console.log('[continueChat] chatId=', this.chatDetailId, 'promptLength=', input.length)
      // 在 API 调用前先清空旧会话数据，避免 SSE 消息在 API 回调前到达时被覆盖丢失
      // （SSE 消息可能在 API 回调触发前就到达，若回调中再清除会丢掉已收到的"继续"气泡等行）
      this.chatDetailSSELines = []
      this.chatDetailMessages = []
      this.chatDetailLastUsageSummary = null
      taskProgressStore.reset()
      this._sseParseState = null
      this._sseLineBuffer = []
      // 标记正在进行 continue 操作，防止 chat_status_change SSE 中的旧状态覆盖本地 running 状态
      this._continueInProgress = true
      this.chatContinueLoading = true
      taskWorkflowApi.TaskWorkflowChatContinue(this.chatDetailId, input, (res) => {
        this.chatContinueLoading = false
        console.log('[continueChat] API返回: ErrCode=', res.ErrCode, 'ErrMsg=', res.ErrMsg)
        if (res.ErrCode === 0) {
          this.chatContinueInput = ''
          this.clearChatInputCache(this.chatDetailId)
          console.log('[continueChat] 设置本地状态为 running')
          this.chatDetailStatus = 'running'
          this.updateChatListStatus(this.chatDetailId, 'running')
          setTimeout(() => { this.loadChatDetail() }, 500)
        } else {
          this._continueInProgress = false
          this.$helperNotify.error(res.ErrMsg || '发送失败')
        }
      })
    },
    // startNewChatFromHistory 在执行历史里直接基于当前输入新建对话，并切换左侧选中项。 // Creates a brand-new chat from history input and focuses the new row immediately.
    startNewChatFromHistory() {
      const prompt = this.chatContinueInput.trim()
      if (!prompt || this.chatContinueLoading) return
      const promptType = String(this.promptChatHistoryPromptType || '').trim()
      const modelName = String(this.chatDetailModelName || this.promptExecModelName || '').trim()
      const thinkingIntensity = String(this.chatDetailThinkingIntensity || this.promptExecThinkingIntensity || '高').trim() || '高'
      this.chatContinueLoading = true
      taskWorkflowApi.TaskWorkflowChatDirs(this.workflowId, (res) => {
        if (!(res && res.ErrCode === 0 && res.Data)) {
          this.chatContinueLoading = false
          this.$helperNotify.error(res?.ErrMsg || '获取工作目录失败')
          return
        }
        const dirs = Array.isArray(res.Data.dirs) ? res.Data.dirs : []
        const localDir = String(this.chatDetailLocalDir || dirs[0] || '').trim()
        if (!localDir) {
          this.chatContinueLoading = false
          this.$helperNotify.error('没有可用的工作目录')
          return
        }
        const cliType = this.chatDetailCliType || this.getSelectedCliType()
        const agentCliId = Number(this.promptExecCliId || 0)
        taskWorkflowApi.TaskWorkflowChatSend(
          this.workflowId,
          prompt,
          promptType,
          localDir,
          cliType,
          agentCliId,
          modelName,
          thinkingIntensity,
          (chatRes) => {
            this.chatContinueLoading = false
            if (!(chatRes && chatRes.ErrCode === 0 && chatRes.Data)) {
              this.$helperNotify.error(chatRes?.ErrMsg || '新建对话失败')
              return
            }
            const chatId = chatRes.Data.chat_id
            const cliLabel = cliType === 'codex' ? 'codex' : 'claude code'
            this.clearChatInputCache(this.chatDetailId)
            this.chatContinueInput = ''
            this.$helperNotify.success('已新建对话并发送到 ' + cliLabel + ' 执行')
            this.chatDetailId = chatId
            this.chatDetailStatus = 'running'
            this.chatDetailCliType = cliType
            this.chatDetailSSELines = []
            this.chatDetailMessages = []
            this.chatDetailLastUsageSummary = null
            taskProgressStore.reset()
            this.markPromptChatRunningLocally(promptType, chatId, {
              prompt,
              modelName,
              thinkingIntensity,
              localDir,
              cliType,
              agentName: this.chatDetailAgentName || '',
            })
            this._sseParseState = null
            this._sseLineBuffer = []
            this.loadChatDetail()
            this.openPromptChatHistory(promptType, chatId)
          }
        )
      })
    },
    // 停止对话
    stopChat() {
      // 关闭 SSE 连接
      this._sseParseState = null
      this._sseLineBuffer = []
      // 通知后端停止
      taskWorkflowApi.TaskWorkflowChatStop(this.chatDetailId, (res) => {
        if (res.ErrCode !== 0) {
          this.$helperNotify.error(res.ErrMsg || '停止失败')
          return
        }
        // 从响应中获取 killed_pid 并更新到对话列表项（刷新页面后即消失）
        const killedPid = res.Data && res.Data.killed_pid
        if (killedPid > 0) {
          const item = this.promptChatHistoryList.find(i => i.id === this.chatDetailId)
          if (item) {
            item._killed_pid = killedPid
          }
        }
      })
      // 前端立即更新状态
      this.chatDetailStatus = 'interrupted'
      this.updateChatListStatus(this.chatDetailId, 'interrupted')
    },
    // 打开执行任务弹窗
    openPromptExecDialog(promptType, promptValue) {
      if (!promptValue || !promptValue.trim()) {
        this.$helperNotify.warning('提示词为空，无法执行')
        return
      }
      this.promptExecPromptType = promptType
      this.promptExecPromptValue = promptValue
      // 从缓存恢复该 promptType 上次的选择
      const cached = this.getPromptExecCache(promptType)
      if (cached) {
        this.promptExecCliId = cached.cliId || 0
        this.promptExecModelName = cached.modelName || ''
        this.promptExecThinkingIntensity = cached.thinkingIntensity || '高'
      } else {
        this.promptExecCliId = 0
        this.promptExecModelName = ''
        this.promptExecThinkingIntensity = '高'
      }
      this.promptExecDialogVisible = true
      // 加载 Agent CLI 列表
      agentCliApi.AgentCliList((res) => {
        if (res.ErrCode === 0 && res.Data) {
          // 仅展示"已启用且配置可用"的 Agent，避免把停用实例带入执行弹窗。
          this.promptExecCliList = (res.Data.list || []).filter(cli => cli.displayed_enabled)
          // 如果无缓存且仅有一个 CLI，自动选中
          if (!cached && this.promptExecCliList.length === 1) {
            this.promptExecCliId = this.promptExecCliList[0].id
          }
          // 如果有缓存 CLI，校验其是否仍在列表中
          if (cached && cached.cliId) {
            const found = this.promptExecCliList.find(c => c.id === cached.cliId)
            if (!found) {
              this.promptExecCliId = 0
              this.promptExecModelName = ''
            }
          }
          this.syncPromptExecModelSelection(cached)
        }
      })
    },
    // CLI 选中变更
    onPromptExecCliChange() {
      this.syncPromptExecModelSelection()
    },
    // 获取当前选中的 CLI 实例对象
    getSelectedCli() {
      if (!this.promptExecCliId) return null
      return this.promptExecCliList.find(c => c.id === this.promptExecCliId) || null
    },
    // promptExecModelOptions 返回当前选中 Agent 卡片可用的模型列表；无配置时回退到 current_model。
    // promptExecModelOptions returns model options for the selected Agent card and falls back to current_model.
    getSelectedCliModelOptions() {
      const cli = this.getSelectedCli()
      if (!cli) return []
      const rawOptions = Array.isArray(cli.model_options) ? cli.model_options : []
      const normalizedOptions = rawOptions
        .map(item => String(item || '').trim())
        .filter(Boolean)
      if (normalizedOptions.length > 0) {
        return normalizedOptions
      }
      const currentModel = String(cli.current_model || '').trim()
      return currentModel && currentModel !== '-' ? [currentModel] : []
    },
    // 获取当前选中 CLI 的 cli_type（'claude' 或 'codex'）
    getSelectedCliType() {
      const cli = this.getSelectedCli()
      if (!cli) return 'claude'
      if (cli.type === 'codex-cli') return 'codex'
      return 'claude'
    },
    // 执行任务
    execPromptToClaude() {
      if (!this.promptExecCliId) {
        this.$helperNotify.warning('请选择 Agent 实例')
        return
      }
      if (this.promptExecModelOptions.length > 0 && !this.promptExecModelName) {
        this.$helperNotify.warning('请选择模型')
        return
      }
      // 执行前检查分支是否匹配
      this.confirmBranchBeforeExec().then((confirmed) => {
        if (!confirmed) return
        this._doExecPromptToClaude()
      })
    },
    _doExecPromptToClaude() {
      // 记录本次选择到缓存
      this.savePromptExecCache(this.promptExecPromptType)
      // 获取第一个可用目录
      taskWorkflowApi.TaskWorkflowChatDirs(this.workflowId, (res) => {
        if (res.ErrCode !== 0) {
          this.$helperNotify.error(res.ErrMsg || '获取工作目录失败')
          return
        }
        const dirs = res.Data.dirs || []
        if (dirs.length === 0) {
          this.$helperNotify.error('没有可用的工作目录')
          return
        }
        const localDir = dirs[0]
        const cliType = this.getSelectedCliType()
        const selectedCli = this.getSelectedCli()
        this.promptExecLoading = true
      taskWorkflowApi.TaskWorkflowChatSend(
        this.workflowId,
        this.promptExecPromptValue,
        this.promptExecPromptType,
        localDir,
        cliType,
        this.promptExecCliId,
        this.promptExecModelName,
        this.promptExecThinkingIntensity,
        (chatRes) => {
            this.promptExecLoading = false
            if (chatRes.ErrCode === 0 && chatRes.Data) {
              const chatId = chatRes.Data.chat_id
              const cliLabel = cliType === 'codex' ? 'codex' : 'claude code'
              this.clearChatInputCache(this.chatDetailId)
              this.$helperNotify.success('已发送到 ' + cliLabel + ' 执行')
              this.promptExecDialogVisible = false
              // 初始化对话显示状态并连接 SSE 流以启动执行
              this.chatDetailId = chatId
              this.chatDetailStatus = 'running'
              this.chatDetailCliType = cliType
              this.chatDetailSSELines = []
              this.chatDetailMessages = []
              taskProgressStore.reset()
              this.markPromptChatRunningLocally(this.promptExecPromptType, chatId, {
                prompt: this.promptExecPromptValue,
                modelName: this.promptExecModelName,
                thinkingIntensity: this.promptExecThinkingIntensity,
                localDir,
                cliType,
                agentName: selectedCli?.name || '',
              })
              this._sseParseState = null
              this._sseLineBuffer = []
              this.loadChatDetail()
              // 打开执行历史，定位到新对话
              this.openPromptChatHistory(this.promptExecPromptType, chatId)
            } else {
              this.$helperNotify.error(chatRes.ErrMsg || '发送失败')
            }
          }
        )
      })
    },
    getPromptChatCounts(promptType) {
      return this.promptChatCounts[promptType] || { running: 0, interrupted: 0, total: 0, unread: 0 }
    },
    hasUnreadInPromptType(promptType) {
      return Number(this.promptChatUnreadCounts[promptType] || 0) > 0
    },
    ensureWorkflowUnreadSse() {
      if (this._workflowUnreadSseId) return
      const nextId = 'workflow_unread_detail'
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
      if (!data || data.type !== 'workflow_unread_snapshot') return
      const list = Array.isArray(data.workflow_detail_badges) ? data.workflow_detail_badges : []
      const detail = list.find(item => Number(item?.workflow_id || 0) === Number(this.workflowId || 0))
      if (!detail) {
        this.workflowUnreadCount = 0
        this.promptChatUnreadCounts = {}
        return
      }
      this.workflowUnreadCount = Number(detail.workflow_unread || 0)
      this.promptChatUnreadCounts = { ...(detail.prompt_type_unread || {}) }
    },
    // 打开按类型的执行历史弹窗
    openPromptChatHistory(promptType, focusChatId) {
      const titleMap = {
        '': '历史对话',
        plain_text_requirement: '纯文本需求',
        requirement: '需求分析',
        design_plan_requirement: '需求设计方案',
        design: '开发提示词',
        api_dev: '接口开发生成',
        code_review: '代码检查',
        browser_test: '浏览器测试',
        api_test: '接口测试修复',
        issue_fix: '问题修改',
      }
      const node = this.workflowNodes.find(n => n.key === promptType)
      this.promptChatHistoryTitle = titleMap[promptType] || (node ? node.label : promptType)
      this.promptChatHistoryPromptType = promptType
      this.promptChatHistoryVisible = true
      this.promptChatHistoryLoading = true
      this.promptChatDetailId = 0
      const loadApi = promptType
        ? (cb) => taskWorkflowApi.TaskWorkflowChatListByPromptType(this.workflowId, promptType, cb)
        : (cb) => taskWorkflowApi.TaskWorkflowChatList(this.workflowId, cb)
      loadApi((res) => {
        this.promptChatHistoryLoading = false
        if (res.ErrCode === 0 && res.Data) {
          this.promptChatHistoryList = res.Data.list || []
          this._startChatHistoryDurationTimer()
          if (focusChatId) {
            const found = this.promptChatHistoryList.find(item => item.id === focusChatId)
            if (found) {
              this.onPromptChatRowClick(found)
              return
            }
            if (this.chatDetailId === focusChatId) {
              this.promptChatDetailId = focusChatId
            }
            return
          }
          if (this.promptChatHistoryList.length > 0) {
            this.onPromptChatRowClick(this.promptChatHistoryList[0])
          }
        } else if (focusChatId && this.chatDetailId === focusChatId) {
          this.promptChatDetailId = focusChatId
        }
      })
    },
    // 点击执行历史列表项
    onPromptChatRowClick(row) {
      if (this.promptChatDetailId === row.id) return
      // 切换前保存当前会话的输入到缓存
      if (this.chatDetailId) {
        this.saveChatInputCache(this.chatDetailId, this.chatContinueInput)
      }
      const oldChatDetailId = this.chatDetailId
      this.promptChatDetailId = row.id
      this.chatDetailId = row.id
      // 切换后从缓存恢复新会话的输入
      this.chatContinueInput = this.getChatInputCache(row.id)
      this.chatDetailStatus = row.status
      this.chatDetailAutoScroll = true
      this.promptChatDetailShowScrollBtn = false
      // 重置 SSE 解析状态，新对话输出由业务 SSE 回调自动接收
      this._sseParseState = null
      this._sseLineBuffer = []
      if (row.is_read === false && row.status !== 'running') {
        this.markPromptChatReadOnServer(row.id)
      }
      if (oldChatDetailId !== row.id) {
        this.chatDetailSSELines = []
        this.chatDetailMessages = []
        this._thinkingStreamStartTime = 0
        this.thinkingStreamElapsed = 0
        taskProgressStore.reset()
        this.loadChatDetail()
      } else {
        this.$nextTick(() => { this.scrollPromptChatToBottom() })
      }
    },
    // 执行历史对话框滚动
    onPromptChatDetailScroll() {
      if (this._autoScrollLocked) return
      const dialog = this.$refs.promptChatHistoryDialog
      if (!dialog || !dialog.isDetailNearBottom) return
      const atBottom = dialog.isDetailNearBottom(30)
      if (atBottom) {
        this.chatDetailAutoScroll = true
        this.promptChatDetailShowScrollBtn = false
      } else {
        this.chatDetailAutoScroll = false
        this.promptChatDetailShowScrollBtn = true
      }
    },
    // 执行历史滚动到底部
    scrollPromptChatToBottom(force) {
      if (!force && !this.chatDetailAutoScroll) return
      if (force) {
        this.chatDetailAutoScroll = true
        this.promptChatDetailShowScrollBtn = false
      }
      this.$nextTick(() => {
        const dialog = this.$refs.promptChatHistoryDialog
        if (dialog && dialog.scrollDetailToBottom) {
          dialog.scrollDetailToBottom('auto')
        }
      })
    },
    handlePromptChatHistoryHide() {
      if (this._promptChatHistoryHideHandled) return
      this._promptChatHistoryHideHandled = true
      this._stopChatHistoryDurationTimer()
      this.closePromptChatDetail()
    },
    // 关闭执行历史弹窗时立即清理前后台 SSE，避免弹窗已关但流仍然存活
    onPromptChatHistoryClosed() {
      this.handlePromptChatHistoryHide()
    },
    // 彻底关闭对话详情（仅在用户主动停止或切换时调用）
    closePromptChatDetail() {
      this.closeChatDetail()
      this.promptChatDetailId = 0
    },
    updateChatListStatus(chatId, status) {
      const normalizedChatId = Number(chatId || 0)
      const selectedPromptHistoryChatId = this.promptChatHistoryVisible
        ? Number(this.promptChatDetailId || 0)
        : 0
      const updateItem = (list) => {
        const item = list.find(i => Number(i.id || 0) === normalizedChatId)
        if (item) {
          item.status = status
        }
      }
      updateItem(this.promptChatHistoryList)
    },
    loadPromptChatHistoryListSilently() {
      if (!this.promptChatHistoryVisible || this.promptChatHistoryLoading) return
      const promptType = this.promptChatHistoryPromptType
      const loadApi = promptType
        ? (cb) => taskWorkflowApi.TaskWorkflowChatListByPromptType(this.workflowId, promptType, cb)
        : (cb) => taskWorkflowApi.TaskWorkflowChatList(this.workflowId, cb)
      loadApi((res) => {
        if (!(res && res.ErrCode === 0 && res.Data)) return
        this.promptChatHistoryList = res.Data.list || []
        if (this.chatDetailId > 0) {
          const current = this.promptChatHistoryList.find(item => item.id === this.chatDetailId)
          if (current) {
            this.chatDetailStatus = current.status || this.chatDetailStatus
          }
        }
      })
    },
    statusText(status) {
      const map = { running: '执行中', completed: '已完成', error: '异常终止', interrupted: '中断' }
      return map[status] || status || '-'
    },
    // 格式化 duration_ms 为可读形式（XmXs 或 Xs）
    formatDurationDisplay(durationMs) {
      const ms = Number(durationMs || 0)
      if (ms <= 0) { return '' }
      const totalSeconds = Math.floor(ms / 1000)
      const minutes = Math.floor(totalSeconds / 60)
      const seconds = totalSeconds % 60
      if (minutes > 0) {
        return minutes + 'm' + seconds + 's'
      }
      return seconds + 's'
    },
    // 计算运行中对话的实时耗时（从 created_at 到现在），返回格式化字符串
    runtimeDurationText(item) {
      if (!item || !item.created_at) { return '' }
      const created = new Date(item.created_at.replace(/-/g, '/'))
      if (isNaN(created.getTime())) { return '' }
      const ms = Date.now() - created.getTime()
      return this.formatDurationDisplay(ms)
    },
    // 启动历史对话列表运行中对话的实时耗时更新定时器
    _startChatHistoryDurationTimer() {
      this._stopChatHistoryDurationTimer()
      this._chatHistoryDurationTimer = setInterval(() => {
        if (this.promptChatHistoryList.some(item => item.status === 'running')) {
          this.promptChatHistoryList = this.promptChatHistoryList.slice()
        }
        // SSE 运行时同步 line_count
        if (this.chatDetailId > 0 && this.chatDetailSSELines.length > 0) {
          const count = this.chatDetailSSELines.length
          const updateLineCount = (list) => {
            const item = list.find(i => i.id === this.chatDetailId)
            if (item && item.line_count !== count) {
              item.line_count = count
            }
          }
          updateLineCount(this.promptChatHistoryList)
        }
      }, 1000)
    },
    // 停止历史对话列表运行中对话的实时耗时更新定时器
    _stopChatHistoryDurationTimer() {
      if (this._chatHistoryDurationTimer) {
        clearInterval(this._chatHistoryDurationTimer)
        this._chatHistoryDurationTimer = null
      }
    },
    // 将 markdown 文本渲染为 HTML，用于"执行历史"对话框中显示表格等格式
    renderMarkdown(text) {
      if (!text) return ''
      return md.render(text)
    },
    formatUnixTime(unixTime) {
      const value = Number(unixTime || 0)
      if (value <= 0) {
        return '-'
      }
      return new Date(value * 1000).toLocaleString()
    },
    // 格式化数字（千位分隔），用于 Token 用量展示
    formatNum(num) {
      if (num == null) return '0'
      return Number(num).toLocaleString()
    },
    // 将 stop_reason 转为中文标签
    stopReasonLabel(reason) {
      const map = {
        end_turn: '正常结束',
        stop_sequence: '停止序列',
        max_tokens: '达到上限',
        tool_use: '工具调用',
      }
      return map[reason] || reason
    },
    loadTaskConfigLookupData() {
      gitApi.GitConfigList({}, (response) => {
        if (response && response.ErrCode === 0) {
          this.taskConfigGitRepoList = Array.isArray(response.Data?.git_list) ? response.Data.git_list : []
        }
      })
      mysqlSetApi.MysqlList((response) => {
        if (response && response.ErrCode === 0) {
          this.taskConfigMysqlList = Array.isArray(response.Data) ? response.Data : []
        }
      })
      apiManagement.CollectionListBasic({}, (response) => {
        if (response && response.ErrCode === 0) {
          this.taskConfigCollectionList = Array.isArray(response.Data?.list) ? response.Data.list : []
        }
      })
      dockerApi.DockerComposeList({}, (response) => {
        if (response && response.ErrCode === 0) {
          this.taskConfigDockerList = Array.isArray(response.Data?.list) ? response.Data.list : []
        }
      })
      smartLinkSetApi.SmartLinkItemList((response) => {
        if (response && response.ErrCode === 0) {
          this.taskConfigSmartLinkList = Array.isArray(response.Data?.smart_link_list) ? response.Data.smart_link_list : []
        }
      })
    },
    loadTaskConfigApiFoldersForCollection(collectionId) {
      if (!collectionId) return
      if (this.taskConfigApiFolderMap[collectionId]) return
      if (this._apiFolderLoading && this._apiFolderLoading[collectionId]) return
      if (!this._apiFolderLoading) this._apiFolderLoading = {}
      this._apiFolderLoading[collectionId] = true
      apiManagement.CollectionFoldersBasic({ collection_id: collectionId }, (response) => {
        this._apiFolderLoading[collectionId] = false
        if (response && response.ErrCode === 0) {
          const list = Array.isArray(response.Data?.list) ? response.Data.list : []
          this.taskConfigApiFolderMap = { ...this.taskConfigApiFolderMap, [collectionId]: list }
        }
      })
    },
    getTaskConfigName(type, id) {
      const numId = Number(id || 0)
      if (numId <= 0) return '-'
      if (type === 'git') {
        const item = this.taskConfigGitRepoList.find(r => Number(r.id) === numId)
        return item ? item.name : String(id)
      }
      if (type === 'docker') {
        const item = this.taskConfigDockerList.find(d => Number(d.id) === numId)
        return item ? item.name : String(id)
      }
      if (type === 'mysql') {
        const item = this.taskConfigMysqlList.find(m => Number(m.id) === numId)
        return item ? item.name : String(id)
      }
      if (type === 'smart_link') {
        const item = this.taskConfigSmartLinkList.find(s => Number(s.id) === numId)
        return item ? item.label : String(id)
      }
      return String(id)
    },
    getTaskConfigApiLabel(cfg) {
      const colId = Number(cfg.collection_id || 0)
      if (colId <= 0) return '-'
      const col = this.taskConfigCollectionList.find(c => Number(c.id) === colId)
      if (!col) return String(cfg.collection_id)
      let label = col.name
      const dirId = Number(cfg.dir_id || 0)
      if (dirId > 0) {
        const folders = this.taskConfigApiFolderMap[colId] || []
        const dir = folders.find(d => Number(d.id) === dirId)
        if (dir) {
          label += '/' + dir.name
        }
      }
      return label
    },
    handleTaskStatusChange(newStatus) {
      if (this.statusUpdating || this.taskId <= 0) return
      if (!newStatus || this.homeTask.task_status === newStatus) return
      this.statusUpdating = true
      homeTaskApi.HomeTaskStatusQuickUpdate(this.taskId, newStatus, (response) => {
        this.statusUpdating = false
        if (!(response && response.ErrCode === 0)) {
          this.$helperNotify.error(response?.ErrMsg || '状态切换失败')
          return
        }
        this.$helperNotify.success('状态已切换')
        this.homeTask = { ...this.homeTask, task_status: newStatus }
      })
    },
    openFragmentInDialog(fragmentId, title) {
      if (!fragmentId) {
        this.$helperNotify.error('片段ID不存在')
        return
      }
      this.fragmentDialogVisible = true
      this.fragmentDialogTitle = title || `知识片段 #${fragmentId}`
      this.fragmentDialogUrl = ''
      this.fragmentDialogLoading = false
      this.fragmentDialogUrl = new URL(`/#/MemoryFragment?fragment_id=${encodeURIComponent(fragmentId)}&hide_menu=1&embed=1`, window.location.origin).toString()
    },
    openApiDevDialog(cfg) {
      const colId = Number(cfg.collection_id || 0)
      if (colId <= 0) {
        this.$helperNotify.warning('未配置接口集合')
        return
      }
      const dirId = Number(cfg.dir_id || 0)
      const label = this.getTaskConfigApiLabel(cfg)
      this.apiDevDialogTitle = '接口开发 - ' + label
      const params = new URLSearchParams()
      params.set('collection_id', String(colId))
      if (dirId > 0) {
        params.set('folder_id', String(dirId))
        params.set('filter_folder_id', String(dirId))
      }
      params.set('hide_menu', '1')
      this.apiDevDialogUrl = new URL(`/#/Api?${params.toString()}`, window.location.origin).toString()
      this.apiDevDialogVisible = true
    },
    getTaskStatusTagType(taskStatus) {
      if (taskStatus === TASK_STATUS_DEVELOPING) {
        return 'success'
      }
      if (taskStatus === TASK_STATUS_SELF_TESTING || taskStatus === TASK_STATUS_TESTING) {
        return 'warning'
      }
      if (taskStatus === TASK_STATUS_TODO) {
        return 'info'
      }
      if (taskStatus === TASK_STATUS_ONLINE) {
        return 'info'
      }
      if (taskStatus === TASK_STATUS_PENDING_TEST) {
        return 'info'
      }
      if (taskStatus === TASK_STATUS_ABANDONED) {
        return 'danger'
      }
      return 'info'
    },
    selectNode(key) {
      this.activeNode = key
      this.saveActiveNodeCache()
    },
    // 判断当前步骤是否包含 is_api_doc 文档（用于展示重置接口文档按钮）
    activeStepHasApiDoc() {
      const stepData = this.templateStepsMap[this.activeNode]
      if (!stepData || !stepData.step_documents) return false
      try {
        const docs = typeof stepData.step_documents === 'string'
          ? JSON.parse(stepData.step_documents) : stepData.step_documents
        return Array.isArray(docs) && docs.some(doc => doc.is_api_doc === true)
      } catch { return false }
    },
    // 获取当前步骤的文档配置列表
    getActiveStepDocuments() {
      const stepData = this.templateStepsMap[this.activeNode]
      if (!stepData || !stepData.step_documents) return []
      try {
        const docs = typeof stepData.step_documents === 'string'
          ? JSON.parse(stepData.step_documents) : stepData.step_documents
        return Array.isArray(docs) ? docs : []
      } catch { return [] }
    },
    // 切换到文档Tab，并每次重新加载文档内容
    switchToDocTab(doc) {
      const docKey = doc.id || doc.name
      this.stepActiveTab = docKey
      this.stepDocTitle = doc.name || ''
      // 清除该文档的未读红点
      if (this.docUnreadFlags[docKey]) {
        this.docUnreadFlags = { ...this.docUnreadFlags, [docKey]: false }
      }
      this.loadDocContent(doc)
    },
    // 强制重新加载文档的知识片段内容（每次切换Tab都会调用API）
    loadDocContent(doc) {
      const docKey = doc.id || doc.name
      const fileId = this.findDocFileId(doc)
      this.stepDocContents[docKey] = { content: '', loading: true, fileId: fileId || '' }
      if (!fileId) {
        this.stepDocContents[docKey].loading = false
        this.stepDocContents[docKey].content = ''
        this.stepDocContents = { ...this.stepDocContents }
        return
      }
      this.stepDocContents = { ...this.stepDocContents }
      MemoryFragmentApi.MemoryFragmentInfo(fileId, (response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.stepDocContents[docKey] = {
            content: response.Data.content || '',
            loading: false,
            fileId: fileId,
          }
        } else {
          this.stepDocContents[docKey] = {
            content: '',
            loading: false,
            fileId: fileId,
          }
        }
        this.stepDocContents = { ...this.stepDocContents }
      })
    },
    // 查找文档对应的 file_id（从 workflowDocuments 中匹配）
    findDocFileId(doc) {
      const docId = doc.id || ''
      const docName = doc.name || ''
      // 从 workflowDocuments 中查找匹配的文档记录
      for (const wd of this.workflowDocuments) {
        const wdDocId = String(wd.document_id || '')
        const wdName = String(wd.document_name || '')
        if (docId && wdDocId === docId) return String(wd.file_id || '')
        if (docName && wdName === docName) return String(wd.file_id || '')
      }
      return ''
    },
    // 获取当前激活文档Tab的名称
    getActiveDocName() {
      const docs = this.getActiveStepDocuments()
      for (const doc of docs) {
        if ((doc.id || doc.name) === this.stepActiveTab) return doc.name || ''
      }
      return ''
    },
    // 获取当前激活文档Tab的内容
    getActiveDocContent() {
      const docKey = this.stepActiveTab
      return this.stepDocContents[docKey]?.content || ''
    },
    // 获取当前激活文档Tab的加载状态
    getActiveDocLoading() {
      const docKey = this.stepActiveTab
      return this.stepDocContents[docKey]?.loading || false
    },
    // 判断当前激活文档Tab是否已加载完成（content 字段已定义，不再处于 loading 状态）
    // 与 getActiveDocContent() 不同：内容为空字符串时仍返回 true，避免空内容文档无法编辑
    isActiveDocLoaded() {
      const docKey = this.stepActiveTab
      const doc = this.stepDocContents[docKey]
      return doc && doc.content !== undefined && !doc.loading && !!doc.fileId
    },
    // 获取当前激活文档Tab的 file_id
    getActiveDocFileId() {
      const docKey = this.stepActiveTab
      return this.stepDocContents[docKey]?.fileId || ''
    },
    // 打开当前文档对应的知识片段
    openActiveDocFragment() {
      const fileId = this.getActiveDocFileId()
      if (!fileId) return
      this.openFragmentInDialog(fileId, this.getActiveDocName())
    },
    // 全屏编辑当前文档（在新窗口中打开独立编辑页面）
    openActiveDocFullscreen() {
      const fileId = this.getActiveDocFileId()
      if (!fileId) {
        this.$helperNotify.warning('请先保存文档后再全屏编辑')
        return
      }
      const url = this.$router.resolve({
        name: 'memory-fragment-fullscreen',
        query: { fragment_id: fileId },
      }).href
      window.open(url, '_blank')
    },
    // 切换到文档查看模式
    switchDocToViewMode() {
      this.clearDocSearch()
      this.docEditMode = false
    },
    // 切换到文档编辑模式
    switchDocToEditMode() {
      this.clearDocSearch()
      this.docEditMode = true
    },
    // 设置当前激活文档内容（编辑模式下双向绑定使用）
    setActiveDocContent(val) {
      const docKey = this.stepActiveTab
      if (this.stepDocContents[docKey]) {
        this.stepDocContents[docKey] = { ...this.stepDocContents[docKey], content: val }
        this.stepDocContents = { ...this.stepDocContents }
      }
    },
    // 下载当前激活文档的知识片段 ZIP 包
    downloadActiveDocZip() {
      const fileId = this.getActiveDocFileId()
      if (!fileId) {
        this.$helperNotify.error('无法下载，文档未关联知识片段')
        return
      }
      MemoryFragmentApi.MemoryFragmentDownloadZip(fileId)
    },
    // 分享当前激活文档（创建24小时有效分享链接并复制到剪贴板）
    shareActiveDoc() {
      if (this.docSharing) return
      const fileId = this.getActiveDocFileId()
      if (!fileId) {
        this.$helperNotify.error('无法分享，文档未关联知识片段')
        return
      }
      this.docSharing = true
      MemoryFragmentApi.MemoryFragmentShareCreate(fileId, async (response) => {
        this.docSharing = false
        if (response.ErrCode !== 0 || !response.Data || !response.Data.token) {
          if (response.ErrMsg) {
            this.$helperNotify.error(response.ErrMsg)
          }
          return
        }
        const shareUrl = baseUtils.BuildMemoryFragmentShareUrl(fileId, response.Data.token, response.Data.url)
        this.copyText(shareUrl, '分享链接已复制到剪贴板（24小时有效）')
      })
    },
    // 处理文档工具栏下拉菜单命令（目前仅支持历史记录）
    handleDocToolbarAction(command) {
      if (command === 'history') {
        const fileId = this.getActiveDocFileId()
        if (!fileId) {
          this.$helperNotify.error('无法查看历史，文档未关联知识片段')
          return
        }
        this.docHistoryFragmentId = fileId
        this.docHistoryVisible = true
      }
    },
    // 切换文档搜索栏显示/隐藏
    toggleDocSearchBar() {
      this.docSearchBarVisible = !this.docSearchBarVisible
      if (this.docSearchBarVisible) {
        this.$nextTick(() => {
          const inputEl = this.$refs.docSearchInput
          if (inputEl && typeof inputEl.focus === 'function') {
            inputEl.focus()
          }
        })
        return
      }
      this.clearDocSearch()
    },
    // 清空文档搜索状态
    clearDocSearch() {
      this.docSearchQuery = ''
      this.docSearchMatchIndex = -1
      this.docSearchMatchCount = 0
      this.removeDocSearchHighlights()
    },
    // 搜索输入变化时触发搜索
    handleDocSearchInput() {
      this.docSearchMatchIndex = 0
      this.$nextTick(() => {
        this.performDocSearch()
      })
    },
    // 执行文档内容搜索（在 MdPreview 渲染的 DOM 中搜索文本并高亮，编辑模式下不生效）
    performDocSearch() {
      this.removeDocSearchHighlights()
      if (!this.docSearchQuery || this.docEditMode) {
        this.docSearchMatchCount = 0
        this.docSearchMatchIndex = -1
        return
      }
      const previewBody = this.$refs.docPreviewBody
      if (!previewBody) return
      const previewEl = previewBody.querySelector('.md-editor-preview')
      if (!previewEl) return
      this._docSearchMatches = []
      const query = this.docSearchQuery
      const lowerQuery = query.toLowerCase()
      const iterateTextNodes = (node) => {
        if (node.nodeType === Node.TEXT_NODE) {
          const text = node.textContent
          const lowerText = text.toLowerCase()
          let startIdx = 0
          while ((startIdx = lowerText.indexOf(lowerQuery, startIdx)) !== -1) {
            this._docSearchMatches.push({ node, start: startIdx, end: startIdx + query.length })
            startIdx += query.length
          }
          return
        }
        if (node.nodeType !== Node.ELEMENT_NODE) return
        const tag = node.tagName?.toLowerCase()
        if (tag === 'script' || tag === 'style') return
        for (const child of node.childNodes) {
          iterateTextNodes(child)
        }
      }
      iterateTextNodes(previewEl)
      this.docSearchMatchCount = this._docSearchMatches.length
      this.docSearchMatchIndex = this.docSearchMatchCount > 0 ? 0 : -1
      this.$nextTick(() => {
        this.highlightDocSearchMatches()
        if (this.docSearchMatchCount > 0) {
          this.scrollToDocSearchMatch(0)
        }
      })
    },
    // 高亮文档搜索结果（在所有匹配文本节点外包一层 span）
    highlightDocSearchMatches() {
      if (!this._docSearchMatches || this._docSearchMatches.length === 0) return
      // 从后往前处理，避免前面的 DOM 修改影响后续索引
      for (let i = this._docSearchMatches.length - 1; i >= 0; i--) {
        const m = this._docSearchMatches[i]
        const textNode = m.node
        if (!textNode || !textNode.parentNode) continue
        try {
          const range = document.createRange()
          range.setStart(textNode, m.start)
          range.setEnd(textNode, m.end)
          const wrapper = document.createElement('span')
          wrapper.className = 'doc-search-highlight'
          range.surroundContents(wrapper)
        } catch (e) {
          // 如果文本节点已被修改则忽略
        }
      }
    },
    // 移除文档搜索结果高亮标记
    removeDocSearchHighlights() {
      this._docSearchMatches = null
      const previewBody = this.$refs.docPreviewBody
      if (!previewBody) return
      // 从 DOM 中移除所有高亮 span（unwrap 恢复原始文本）
      const highlights = previewBody.querySelectorAll('.doc-search-highlight')
      highlights.forEach(el => {
        const parent = el.parentNode
        if (!parent) return
        while (el.firstChild) {
          parent.insertBefore(el.firstChild, el)
        }
        parent.removeChild(el)
      })
    },
    // 跳转到上一个/下一个搜索匹配项（delta: -1 上一个, 1 下一个）
    jumpDocSearchMatch(delta) {
      if (this.docSearchMatchCount === 0) return
      let idx = this.docSearchMatchIndex + delta
      if (idx >= this.docSearchMatchCount) idx = 0
      if (idx < 0) idx = this.docSearchMatchCount - 1
      this.docSearchMatchIndex = idx
      this.scrollToDocSearchMatch(idx)
    },
    // 滚动到指定索引的搜索匹配项
    scrollToDocSearchMatch(idx) {
      const previewBody = this.$refs.docPreviewBody
      if (!previewBody) return
      const highlights = previewBody.querySelectorAll('.doc-search-highlight')
      if (!highlights || highlights.length <= idx) return
      const target = highlights[idx]
      target.scrollIntoView({ behavior: 'smooth', block: 'center' })
    },
    // 保存当前激活文档内容为知识片段
    saveActiveDoc() {
      const content = this.getActiveDocContent()
      // 允许保存空内容（用户可能需要清空文档）
      const fileId = this.getActiveDocFileId()
      if (!fileId) {
        this.$helperNotify.error('未找到对应的知识片段标识')
        return
      }
      const docName = this.getActiveDocName() || '文档'
      // 调用知识片段保存接口 (id, title, content, tags, folderName, callBack)
      MemoryFragmentApi.MemoryFragmentSave(fileId, docName, content, '', '', (response) => {
        if (response && response.ErrCode === 0) {
          this.$helperNotify.success('文档已保存')
          // 更新缓存中的文件ID
          const docKey = this.stepActiveTab
          if (this.stepDocContents[docKey]) {
            const newFileId = response.Data?.file_id || response.Data?.FileId || fileId
            this.stepDocContents[docKey] = { ...this.stepDocContents[docKey], fileId: newFileId }
            this.stepDocContents = { ...this.stepDocContents }
          }
        } else {
          this.$helperNotify.error(response?.ErrMsg || '保存文档失败')
        }
      })
    },
    // ===== 提示词预览目录导航方法 =====
    // MdPreview 渲染完成后的目录回调，结合 DOM 获取标题 ID
    onPromptCatalog(catalogList) {
      if (!catalogList || catalogList.length === 0) {
        this.promptTocItems = []
        return
      }
      this.$nextTick(() => {
        const previewEl = this.$refs.promptPreviewBody
        if (!previewEl) {
          this.promptTocItems = catalogList.map((item) => ({
            level: item.level,
            text: item.text,
            id: '',
          }))
          return
        }
        const mdPreviewEl = previewEl.closest('.preview-body')?.querySelector('.md-editor-preview')
        const domHeadings = (mdPreviewEl || previewEl).querySelectorAll('h1, h2, h3, h4')
        this.promptTocItems = catalogList.map((item, i) => ({
          level: item.level,
          text: item.text,
          id: domHeadings[i] ? domHeadings[i].id : '',
        }))
        if (this.promptTocItems.length > 0) {
          this.$nextTick(() => this.setupPromptScrollSpy())
        }
      })
    },
    // 点击目录项，平滑滚动到对应标题
    scrollToPromptHeading(id) {
      const el = document.getElementById(id)
      if (el) {
        el.scrollIntoView({ behavior: 'smooth', block: 'start' })
        this.promptActiveHeading = id
      }
    },
    // 提示词预览滚动时同步目录高亮（使用 requestAnimationFrame 节流）
    handlePromptPreviewScroll() {
      if (this._promptPreviewScrollRafId) {
        cancelAnimationFrame(this._promptPreviewScrollRafId)
      }
      this._promptPreviewScrollRafId = requestAnimationFrame(() => {
        this.syncPromptOutlineByScroll()
      })
    },
    // 根据预览区滚动位置高亮最接近的目录项
    syncPromptOutlineByScroll() {
      const previewBody = this.$refs.promptPreviewBody
      if (!previewBody || this.promptTocItems.length === 0) {
        this.promptActiveHeading = ''
        return
      }
      const mdPreviewEl = previewBody.closest('.preview-body')?.querySelector('.md-editor-preview')
      const headingList = (mdPreviewEl || previewBody).querySelectorAll('h1, h2, h3, h4')
      if (headingList.length === 0) {
        this.promptActiveHeading = this.promptTocItems[0].id
        return
      }
      const scrollTop = previewBody.scrollTop
      const matchedHeading = Array.from(headingList).reduce((current, heading) => {
        if (heading.offsetTop <= scrollTop + 100) return heading
        return current
      }, headingList[0])
      this.promptActiveHeading = matchedHeading.id || this.promptTocItems[0].id
    },
    // 使用 IntersectionObserver 追踪当前可见标题
    setupPromptScrollSpy() {
      this.cleanupPromptScrollObserver()
      const previewBody = this.$refs.promptPreviewBody
      if (!previewBody) return
      const mdPreviewEl = previewBody.closest('.preview-body')?.querySelector('.md-editor-preview')
      const headingEls = (mdPreviewEl || previewBody).querySelectorAll('h1, h2, h3, h4')
      if (headingEls.length === 0) return
      this.promptScrollObserver = new IntersectionObserver(
        (entries) => {
          for (const entry of entries) {
            if (entry.isIntersecting && entry.target.id) {
              this.promptActiveHeading = entry.target.id
              break
            }
          }
        },
        { rootMargin: '-80px 0px -60% 0px', threshold: 0.1 }
      )
      headingEls.forEach((el) => this.promptScrollObserver.observe(el))
    },
    // 清理提示词预览滚动监听器
    cleanupPromptScrollObserver() {
      if (this.promptScrollObserver) {
        this.promptScrollObserver.disconnect()
        this.promptScrollObserver = null
      }
      if (this._promptPreviewScrollRafId) {
        cancelAnimationFrame(this._promptPreviewScrollRafId)
        this._promptPreviewScrollRafId = 0
      }
    },
    // ===== 文档预览目录导航方法 =====
    // MdPreview 渲染完成后的目录回调，结合 DOM 获取标题 ID
    onDocCatalog(catalogList) {
      if (!catalogList || catalogList.length === 0) {
        this.docTocItems = []
        return
      }
      this.$nextTick(() => {
        const previewEl = this.$refs.docPreviewBody
        if (!previewEl) {
          this.docTocItems = catalogList.map((item) => ({
            level: item.level,
            text: item.text,
            id: '',
          }))
          return
        }
        const mdPreviewEl = previewEl.closest('.preview-body')?.querySelector('.md-editor-preview')
        const domHeadings = (mdPreviewEl || previewEl).querySelectorAll('h1, h2, h3, h4')
        this.docTocItems = catalogList.map((item, i) => ({
          level: item.level,
          text: item.text,
          id: domHeadings[i] ? domHeadings[i].id : '',
        }))
        if (this.docTocItems.length > 0) {
          this.$nextTick(() => this.setupDocScrollSpy())
        }
      })
    },
    // 点击目录项，平滑滚动到对应标题
    scrollToDocHeading(id) {
      const el = document.getElementById(id)
      if (el) {
        el.scrollIntoView({ behavior: 'smooth', block: 'start' })
        this.docActiveHeading = id
      }
    },
    // 文档预览滚动时同步目录高亮（使用 requestAnimationFrame 节流）
    handleDocPreviewScroll() {
      if (this._docPreviewScrollRafId) {
        cancelAnimationFrame(this._docPreviewScrollRafId)
      }
      this._docPreviewScrollRafId = requestAnimationFrame(() => {
        this.syncDocOutlineByScroll()
      })
    },
    // 根据预览区滚动位置高亮最接近的目录项
    syncDocOutlineByScroll() {
      const previewBody = this.$refs.docPreviewBody
      if (!previewBody || this.docTocItems.length === 0) {
        this.docActiveHeading = ''
        return
      }
      const mdPreviewEl = previewBody.closest('.preview-body')?.querySelector('.md-editor-preview')
      const headingList = (mdPreviewEl || previewBody).querySelectorAll('h1, h2, h3, h4')
      if (headingList.length === 0) {
        this.docActiveHeading = this.docTocItems[0].id
        return
      }
      const scrollTop = previewBody.scrollTop
      const matchedHeading = Array.from(headingList).reduce((current, heading) => {
        if (heading.offsetTop <= scrollTop + 100) return heading
        return current
      }, headingList[0])
      this.docActiveHeading = matchedHeading.id || this.docTocItems[0].id
    },
    // 使用 IntersectionObserver 追踪当前可见标题
    setupDocScrollSpy() {
      this.cleanupDocScrollObserver()
      const previewBody = this.$refs.docPreviewBody
      if (!previewBody) return
      const mdPreviewEl = previewBody.closest('.preview-body')?.querySelector('.md-editor-preview')
      const headingEls = (mdPreviewEl || previewBody).querySelectorAll('h1, h2, h3, h4')
      if (headingEls.length === 0) return
      this.docScrollObserver = new IntersectionObserver(
        (entries) => {
          for (const entry of entries) {
            if (entry.isIntersecting && entry.target.id) {
              this.docActiveHeading = entry.target.id
              break
            }
          }
        },
        { rootMargin: '-80px 0px -60% 0px', threshold: 0.1 }
      )
      headingEls.forEach((el) => this.docScrollObserver.observe(el))
    },
    // 清理文档预览滚动监听器
    cleanupDocScrollObserver() {
      if (this.docScrollObserver) {
        this.docScrollObserver.disconnect()
        this.docScrollObserver = null
      }
      if (this._docPreviewScrollRafId) {
        cancelAnimationFrame(this._docPreviewScrollRafId)
        this._docPreviewScrollRafId = 0
      }
    },
    // 根据 fileId 反向查找文档 key（用于未读红点标记）
    findDocKeyByFileId(fileId) {
      if (!fileId) return null
      const fid = String(fileId)
      // 1. 先从已加载的 stepDocContents 中查找（快速路径）
      for (const docKey of Object.keys(this.stepDocContents)) {
        if (String(this.stepDocContents[docKey]?.fileId || '') === fid) return docKey
      }
      // 2. 从 workflowDocuments 构建反向映射（file_id → document_id / document_name）
      const fidToDocId = {}
      const fidToDocName = {}
      for (const wd of this.workflowDocuments) {
        const wFid = String(wd.file_id || '')
        if (!wFid) continue
        if (wd.document_id) fidToDocId[wFid] = String(wd.document_id)
        if (wd.document_name) fidToDocName[wFid] = String(wd.document_name)
      }
      // 3. 遍历所有模板步骤的文档，匹配 fileId
      for (const step of (this._cachedTemplateSteps || [])) {
        const docs = this.parseStepDocumentsRaw(step)
        for (const doc of docs) {
          const docId = String(doc.id || '')
          const docName = String(doc.name || '')
          if ((docId && fidToDocId[fid] === docId) || (docName && fidToDocName[fid] === docName)) {
            return docId || docName
          }
        }
      }
      return null
    },
    // 解析步骤文档 JSON（与 getActiveStepDocuments 逻辑一致，但不依赖 activeNode）
    parseStepDocumentsRaw(step) {
      if (!step || !step.step_documents) return []
      try {
        const docs = typeof step.step_documents === 'string'
          ? JSON.parse(step.step_documents) : step.step_documents
        return Array.isArray(docs) ? docs : []
      } catch { return [] }
    },
    // SSE推送文档内容变更时重新加载对应文档
    reloadDocContentByFileId(fileId) {
      if (!fileId) return
      for (const docKey of Object.keys(this.stepDocContents)) {
        if (this.stepDocContents[docKey]?.fileId === fileId) {
          this.stepDocContents[docKey] = { ...this.stepDocContents[docKey], loading: true }
          this.stepDocContents = { ...this.stepDocContents }
          MemoryFragmentApi.MemoryFragmentInfo(fileId, (response) => {
            const newContent = (response && response.ErrCode === 0 && response.Data) ? (response.Data.content || '') : ''
            this.stepDocContents[docKey] = { content: newContent, loading: false, fileId: fileId }
            this.stepDocContents = { ...this.stepDocContents }
          })
        }
      }
    },
    getActiveNodeCacheKey() {
      return ACTIVE_NODE_CACHE_PREFIX + this.taskId
    },
    saveActiveNodeCache() {
      if (this.taskId > 0 && this.activeNode) {
        localStorage.setItem(this.getActiveNodeCacheKey(), this.activeNode)
      }
    },
    restoreActiveNodeCache() {
      if (this.taskId <= 0) return null
      const cached = localStorage.getItem(this.getActiveNodeCacheKey())
      if (cached === 'task-config') return null
      return cached
    },
    truncateWorkflowLabel(label) {
      const str = String(label || '-')
      if (str.length <= TASK_WORKFLOW_CONFIG_MAX_CHARS) {
        return str
      }
      return str.slice(0, TASK_WORKFLOW_CONFIG_MAX_CHARS) + '...'
    },
    // 执行历史弹窗：点击任务项滚动到对应消息
    onPromptTaskPanelScrollToMsg(msgIndex) {
      const dialog = this.$refs.promptChatHistoryDialog
      const container = dialog && dialog.getDetailContainer ? dialog.getDetailContainer() : null
      if (!container) return
      const children = container.children
      if (msgIndex >= 0 && msgIndex < children.length) {
        children[msgIndex].scrollIntoView({ behavior: 'smooth', block: 'center' })
      }
    },
    // 获取列表项的消息数：运行中的对话使用实时SSE消息计数，否则使用数据库持久化的line_count
    getItemMsgCount(item) {
      if (item.status === 'running' && this.chatDetailId > 0 && item.id === this.chatDetailId) {
        return this.chatDetailSSELines.length
      }
      return item.line_count || 0
    },
    // 格式化创建时间为 2026/01/02 12:12:12 格式
    formatCreatedAt(createdAt) {
      if (!createdAt) { return '' }
      // created_at 格式为 2026-01-02 12:12:12 或 ISO 格式，统一转为目标格式
      const d = new Date(createdAt.replace(/-/g, '/'))
      if (isNaN(d.getTime())) { return '' }
      const pad = (n) => String(n).padStart(2, '0')
      return d.getFullYear() + '/' + pad(d.getMonth() + 1) + '/' + pad(d.getDate()) + ' ' +
        pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds())
    },
    // 获取执行弹窗缓存的配置（按 promptType 区分）
    getPromptExecCacheKey(promptType) {
      return PROMPT_EXEC_CACHE_PREFIX + promptType
    },
    getPromptExecCache(promptType) {
      try {
        const raw = localStorage.getItem(this.getPromptExecCacheKey(promptType))
        return raw ? JSON.parse(raw) : null
      } catch {
        return null
      }
    },
    savePromptExecCache(promptType) {
      const data = {
        cliId: this.promptExecCliId,
        modelName: this.promptExecModelName,
        thinkingIntensity: this.promptExecThinkingIntensity,
      }
      localStorage.setItem(this.getPromptExecCacheKey(promptType), JSON.stringify(data))
    },
    // syncPromptExecModelSelection 根据当前 Agent 和缓存自动恢复模型选择；仅一个模型时自动选中。
    // syncPromptExecModelSelection restores the model choice from cache and auto-selects when only one model exists.
    syncPromptExecModelSelection(cachedValue) {
      const cached = cachedValue || this.getPromptExecCache(this.promptExecPromptType) || null
      const modelOptions = this.getSelectedCliModelOptions()
      if (modelOptions.length === 0) {
        this.promptExecModelName = ''
        return
      }
      if (cached && cached.cliId === this.promptExecCliId && cached.modelName && modelOptions.includes(cached.modelName)) {
        this.promptExecModelName = cached.modelName
        return
      }
      if (modelOptions.includes(this.promptExecModelName)) {
        return
      }
      if (modelOptions.length === 1) {
        this.promptExecModelName = modelOptions[0]
        return
      }
      this.promptExecModelName = ''
    },
    // 批量检查工作流页面中本地目录是否存在
    checkWorkflowLocalDirExists() {
      const configs = this.parsedTaskDevConfigs
      const paths = []
      for (const cfg of configs) {
        const dir = String(cfg.local_dir || '').trim()
        if (dir && !paths.includes(dir)) {
          paths.push(dir)
        }
      }
      if (paths.length === 0) return
      homeTaskApi.LocalDirBatchCheck(paths, (response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.localDirStatusMap = { ...this.localDirStatusMap, ...response.Data }
        }
      })
    },
    // 在指定 IDE 中打开工作目录
    openInEditor(localDir, editorType) {
      if (!localDir || !editorType) return
      taskWorkflowApi.TaskWorkflowOpenInEditor(localDir, editorType, (response) => {
        if (response && response.ErrCode === 0) {
          this.$message.success(response.Msg || `已在 ${editorType} 中打开目录`)
        } else {
          this.$message.error((response && response.Msg) || `打开失败`)
        }
      })
    },
    // 批量检查工作流页面中本地目录的当前 Git 分支是否与配置的分支名匹配
    checkWorkflowBranchStatus() {
      const items = this.getWorkflowBranchCheckItems()
      if (items.length === 0) return
      homeTaskApi.LocalBranchBatchCheck(items, (response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.branchStatusMap = { ...this.branchStatusMap, ...response.Data }
          this.maybeAutoOpenBranchMismatchDialog()
          if (this.branchMismatchDialogVisible) {
            this.loadBranchMismatchDetail({ showDialog: false, mode: this.branchMismatchDialogMode })
          }
        }
      })
    },
    getWorkflowBranchCheckItems() {
      const items = []
      const seen = new Set()
      for (const cfg of this.parsedTaskDevConfigs) {
        const dir = String(cfg.local_dir || '').trim()
        const branch = String(cfg.branch_name || '').trim()
        if (!dir || !branch) continue
        const key = dir + '|' + branch
        if (seen.has(key)) continue
        seen.add(key)
        items.push({
          git_id: Number(cfg.git_id || 0),
          parent_branch: String(cfg.parent_branch || '').trim(),
          local_dir: dir,
          branch_name: branch,
        })
      }
      return items
    },
    // ===== 文件变更检测（10s 轮询）=====
    // 获取所有需要检测文件变更的本地目录列表
    getFileChangesLocalDirs() {
      const items = []
      const seen = new Set()
      for (const cfg of this.parsedTaskDevConfigs) {
        const dir = String(cfg.local_dir || '').trim()
        if (!dir || seen.has(dir)) continue
        seen.add(dir)
        items.push({
          local_dir: dir,
          parent_branch: String(cfg.parent_branch || '').trim(),
        })
      }
      return items
    },
    // 检查文件变更（调用 API）
    loadFileChangesSummary() {
      const items = this.getFileChangesLocalDirs()
      if (items.length === 0) return
      taskWorkflowApi.TaskWorkflowFileChangesSummary(items, (response) => {
        if (response && response.ErrCode === 0 && response.Data && response.Data.dirs) {
          this.fileChangesMap = { ...this.fileChangesMap, ...response.Data.dirs }
        }
      })
    },
    // 启动文件变更轮询（10s 间隔）
    startFileChangesPolling() {
      this.stopFileChangesPolling()
      this.loadFileChangesSummary()
      this._fileChangesPollTimer = setInterval(() => {
        this.loadFileChangesSummary()
      }, 10000)
    },
    // 停止文件变更轮询
    stopFileChangesPolling() {
      if (this._fileChangesPollTimer) {
        clearInterval(this._fileChangesPollTimer)
        this._fileChangesPollTimer = null
      }
    },
    // 获取指定目录的文件变更汇总
    getFileChangesSummary(localDir) {
      if (!localDir || !this.fileChangesMap[localDir]) return null
      const item = this.fileChangesMap[localDir]
      if (item.error) return null
      return item.summary || null
    },
    // 打开文件变更详情弹窗
    openFileChangesDetail(cfg) {
      const localDir = String(cfg.local_dir || '').trim()
      const parentBranch = String(cfg.parent_branch || '').trim()
      this.fileChangesDetailLocalDir = localDir
      this.fileChangesDetailParentBranch = parentBranch
      this.fileChangesDetailInitialSummary = this.fileChangesMap[localDir]?.summary || null
      this.fileChangesDetailInitialFiles = this.fileChangesMap[localDir]?.files || []
      this.fileChangesDialogVisible = true
    },
    getBranchMismatchItemKey(item) {
      return String(item.local_dir || '') + '|' + String(item.expected_branch || item.branch_name || '')
    },
    // 检查所有 dev_config 的分支是否匹配，返回不匹配的列表
    getMismatchedBranches() {
      const mismatched = []
      for (const cfg of this.getWorkflowBranchCheckItems()) {
        const dir = String(cfg.local_dir || '').trim()
        const branch = String(cfg.branch_name || '').trim()
        if (!dir || !branch) continue
        const key = dir + '|' + branch
        const status = this.branchStatusMap[key]
        if (status && !status.matched) {
          mismatched.push({
            git_id: Number(cfg.git_id || 0),
            parent_branch: String(cfg.parent_branch || '').trim(),
            local_dir: dir,
            expected_branch: branch,
            current_branch: status.current_branch || '未知',
          })
        }
      }
      return mismatched
    },
    maybeAutoOpenBranchMismatchDialog() {
      if (this.branchMismatchPromptedTaskId === this.taskId) return
      if (this.getMismatchedBranches().length === 0) return
      this.branchMismatchPromptedTaskId = this.taskId
      this.loadBranchMismatchDetail({ showDialog: true, mode: 'notice' })
    },
    loadBranchMismatchDetail({ showDialog = true, mode = 'notice' } = {}) {
      const items = this.getWorkflowBranchCheckItems()
      if (items.length === 0) {
        this.branchMismatchDetailList = []
        return Promise.resolve([])
      }
      this.branchMismatchLoading = true
      this.branchMismatchDialogMode = mode
      if (showDialog) {
        this.branchMismatchDialogVisible = true
      }
      return new Promise((resolve) => {
        homeTaskApi.LocalBranchMismatchDetail(items, (response) => {
          this.branchMismatchLoading = false
          if (response && response.ErrCode === 0 && Array.isArray(response.Data)) {
            this.branchMismatchDetailList = response.Data
            resolve(response.Data)
            return
          }
          this.branchMismatchDetailList = []
          this.$helperNotify.error(response?.ErrMsg || '加载分支检查详情失败')
          resolve([])
        })
      })
    },
    openBranchMismatchDialog(mode = 'notice') {
      if (this.getMismatchedBranches().length === 0) {
        return Promise.resolve(true)
      }
      return new Promise((resolve) => {
        this.branchMismatchDialogResolver = resolve
        this.loadBranchMismatchDetail({ showDialog: true, mode }).then(() => {})
      })
    },
    closeBranchMismatchDialog(confirmed) {
      this.branchMismatchDialogVisible = false
      const resolver = this.branchMismatchDialogResolver
      this.branchMismatchDialogResolver = null
      if (resolver) {
        resolver(!!confirmed)
      }
    },
    handleBranchMismatchDialogClose() {
      if (this.branchMismatchDialogVisible) return
      const resolver = this.branchMismatchDialogResolver
      this.branchMismatchDialogResolver = null
      if (resolver) {
        resolver(false)
      }
    },
    handleCleanupAndSwitchBranch(item) {
      const expectedBranch = String(item.expected_branch || '').trim()
      const baseBranch = String(item.parent_branch || '').trim()
      const localDir = String(item.local_dir || '').trim()
      const gitId = Number(item.git_id || 0)
      if (!localDir || !expectedBranch || !baseBranch) {
        this.$helperNotify.warning('缺少本地目录、父分支或目标分支，无法自动切换')
        return
      }
      this.$confirm(
        `将清理本地目录 ${item.local_dir || ''} 下所有未提交或已变更文件，并先切换到 ${baseBranch} 拉取最新代码，再基于此分支创建 ${expectedBranch}。\n\n确定继续吗？`,
        '确认清理并切换分支',
        {
          confirmButtonText: '确认清理并切换',
          cancelButtonText: '取消',
          type: 'warning',
        }
      ).then(() => {
        const switchKey = this.getBranchMismatchItemKey(item)
        this.branchSwitchingKey = switchKey
        this.startBranchSwitchStream({
          git_id: gitId,
          local_dir: localDir,
          base_branch: baseBranch,
          branch_name: expectedBranch,
          switchKey,
        })
      }).catch(() => {})
    },
    closeBranchSwitchStreamDialog() {
      this.branchSwitchStreamDialogVisible = false
      this.branchSwitchStreamRunning = false
      this.branchSwitchingKey = ''
      if (this._branchSwitchEventSource) {
        this._branchSwitchEventSource.close()
        this._branchSwitchEventSource = null
      }
    },
    // 远程分支状态检测
    getRemoteBranchStatusKey(cfg) {
      return (cfg.local_dir || '') + '|' + (cfg.branch_name || '')
    },
    getRemoteBranchStatus(cfg) {
      return this.remoteBranchStatusMap[this.getRemoteBranchStatusKey(cfg)] || null
    },
    getRemoteBranchStatusTooltip(cfg) {
      const s = this.getRemoteBranchStatus(cfg)
      if (!s) return ''
      if (s.error) return '远程检测失败: ' + s.error
      if (!s.remote_exists) return '远程分支不存在，点击查看详情'
      if (!s.consistent) return '本地与远程不一致（本地领先' + (s.local_ahead || 0) + '，远程领先' + (s.remote_ahead || 0) + '），点击查看详情'
      if (!s.pushed) return '分支未推送到远程，点击查看详情'
      if (s.remote_dir_error) return '远程工作目录检测失败: ' + s.remote_dir_error
      if (!s.remote_dir_branch_match) return '远程工作目录分支不一致（当前: ' + (s.remote_dir_current_branch || '未知') + '），点击查看详情'
      return ''
    },
    showRemoteBranchWarning(cfg) {
      const s = this.getRemoteBranchStatus(cfg)
      if (!s) return false
      if (s.error) return true
      if (s.remote_dir_error) return true
      return !s.remote_exists || !s.pushed || !s.consistent || !s.remote_dir_branch_match
    },
    showRemoteBranchOk(cfg) {
      const s = this.getRemoteBranchStatus(cfg)
      if (!s) return false
      return !s.error && s.remote_exists && s.pushed && s.consistent && s.remote_dir_branch_match !== false
    },
    checkWorkflowRemoteBranchStatus() {
      const items = this.getWorkflowBranchCheckItems()
      if (items.length === 0) return
      homeTaskApi.RemoteBranchCheck(items, (response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.remoteBranchStatusMap = { ...this.remoteBranchStatusMap, ...response.Data }
        }
      })
    },
    openRemoteBranchDialog(cfg) {
      this.remoteBranchDialogVisible = true
      this.remoteBranchDialogPushResult = null
      this.remoteBranchDialogLoading = true
      this.remoteBranchDialogItem = {
        local_dir: cfg.local_dir || '',
        branch_name: cfg.branch_name || '',
        git_id: Number(cfg.git_id || 0),
      }
      // 重新检测远程分支状态
      const item = {
        local_dir: cfg.local_dir || '',
        branch_name: cfg.branch_name || '',
        git_id: Number(cfg.git_id || 0),
      }
      homeTaskApi.RemoteBranchCheck([item], (response) => {
        this.remoteBranchDialogLoading = false
        if (response && response.ErrCode === 0 && response.Data) {
          const key = this.getRemoteBranchStatusKey(cfg)
          const info = response.Data[key]
          if (info) {
            this.remoteBranchDialogItem = { ...this.remoteBranchDialogItem, ...info }
            this.remoteBranchStatusMap = { ...this.remoteBranchStatusMap, [key]: info }
          }
        } else {
          this.remoteBranchDialogItem = { ...this.remoteBranchDialogItem, error: '检测失败' }
        }
      })
    },
    closeRemoteBranchDialog() {
      this.remoteBranchDialogVisible = false
      this.remoteBranchDialogPushResult = null
    },
    refreshRemoteBranchDialog() {
      this.openRemoteBranchDialog({
        local_dir: this.remoteBranchDialogItem.local_dir,
        branch_name: this.remoteBranchDialogItem.branch_name,
        git_id: this.remoteBranchDialogItem.git_id,
      })
    },
    handleRemoteBranchPush() {
      if (!this.remoteBranchDialogItem.local_dir || !this.remoteBranchDialogItem.branch_name) {
        this.$helperNotify.warning('目录或分支信息不完整')
        return
      }
      const item = this.remoteBranchDialogItem
      // 如果已推送且同步，跳过推送，直接切换远程工作目录分支
      if (item.pushed && item.consistent) {
        this.switchRemoteBranch()
        return
      }
      // 否则先推送再切换
      this.remoteBranchPushing = true
      this.remoteBranchDialogPushResult = null
      homeTaskApi.RemoteBranchPush({
        local_dir: item.local_dir,
        branch_name: item.branch_name,
        git_id: item.git_id,
      }, (response) => {
        if (response && response.ErrCode === 0) {
          this.$helperNotify.success('远程分支推送成功')
          // 推送成功后继续切换远程工作目录分支
          this.switchRemoteBranch()
        } else {
          this.remoteBranchPushing = false
          this.remoteBranchDialogPushResult = { success: false, message: response?.ErrMsg || response?.Msg || '推送失败' }
          this.$helperNotify.error('远程分支推送失败')
        }
      })
    },
    switchRemoteBranch() {
      const item = this.remoteBranchDialogItem
      if (!item.git_id || !item.branch_name) {
        this.remoteBranchPushing = false
        this.$helperNotify.warning('缺少Git配置或分支名，无法切换远程分支')
        return
      }
      this.remoteBranchPushing = true
      homeTaskApi.RemoteBranchSwitch(item.git_id, item.branch_name, (response) => {
        this.remoteBranchPushing = false
        if (response && response.ErrCode === 0) {
          this.remoteBranchDialogPushResult = { success: true, message: '远程工作目录分支切换成功' }
          this.$helperNotify.success('远程工作目录分支切换成功')
          setTimeout(() => {
            this.refreshRemoteBranchDialog()
          }, 1500)
        } else {
          this.remoteBranchDialogPushResult = { success: false, message: response?.ErrMsg || response?.Msg || '远程分支切换失败' }
          this.$helperNotify.error('远程分支切换失败')
        }
      })
    },
    appendBranchSwitchLog(text, level = 'info') {
      this.branchSwitchStreamLines.push({ text, level })
      this.$nextTick(() => {
        this.scrollBranchSwitchLogToBottom()
      })
    },
    scrollBranchSwitchLogToBottom() {
      const logEl = this.$refs.branchSwitchStreamLog
      if (!logEl) {
        return
      }
      logEl.scrollTop = logEl.scrollHeight
    },
    startBranchSwitchStream(payload) {
      if (this._branchSwitchEventSource) {
        this._branchSwitchEventSource.close()
        this._branchSwitchEventSource = null
      }
      this.branchSwitchStreamMeta = {
        local_dir: payload.local_dir || '',
        base_branch: payload.base_branch || '',
        branch_name: payload.branch_name || '',
      }
      this.branchSwitchStreamRunning = true
      this.branchSwitchStreamLines = []
      this.branchSwitchStreamDialogVisible = true
      this.$nextTick(() => {
        this.scrollBranchSwitchLogToBottom()
      })
      const url = gitApi.GitCleanupAndSwitchBranchByIdStreamUrl(payload)
      if (!url) {
        this.branchSwitchingKey = ''
        this.branchSwitchStreamRunning = false
        this.appendBranchSwitchLog('SSE 连接不可用，无法实时展示切换步骤', 'error')
        this.$helperNotify.error('SSE 连接不可用')
        return
      }
      this.appendBranchSwitchLog(`开始在本地目录执行：${payload.local_dir}`)
      const es = new EventSource(url)
      this._branchSwitchEventSource = es
      es.onmessage = (event) => {
        const raw = event.data
        if (!raw || raw === '[DONE]' || raw === '[CONNECT]') return
        let parsed = null
        try {
          parsed = JSON.parse(raw)
        } catch (e) {
          this.appendBranchSwitchLog(raw)
          return
        }
        if (parsed.type === 'line') {
          this.appendBranchSwitchLog(parsed.message || '')
          return
        }
        if (parsed.type === 'step') {
          const level = parsed.status === 'error' ? 'error' : (parsed.status === 'done' ? 'success' : 'info')
          this.appendBranchSwitchLog(parsed.message || '', level)
          return
        }
        if (parsed.type === 'meta') {
          this.appendBranchSwitchLog(parsed.message || '', 'info')
          return
        }
        if (parsed.type === 'done') {
          this.branchSwitchStreamRunning = false
          if (parsed.status === 'success') {
            this.appendBranchSwitchLog(parsed.message || '分支切换完成', 'success')
            this.$helperNotify.success(`已切换到 ${payload.branch_name}`)
            this.checkWorkflowBranchStatus()
            this.checkWorkflowRemoteBranchStatus()
            this.loadBranchMismatchDetail({ showDialog: true, mode: this.branchMismatchDialogMode })
          } else {
            this.appendBranchSwitchLog(parsed.message || '清理并切换分支失败', 'error')
            this.$helperNotify.error(parsed.message || '清理并切换分支失败')
          }
          this.branchSwitchingKey = ''
          es.close()
          this._branchSwitchEventSource = null
        }
      }
      es.onerror = () => {
        if (this._branchSwitchEventSource === es) {
          this.branchSwitchStreamRunning = false
          this.appendBranchSwitchLog('连接已断开', 'error')
          this.branchSwitchingKey = ''
          es.close()
          this._branchSwitchEventSource = null
        }
      }
    },
    // 执行前检查分支是否匹配，不匹配时弹出确认框，返回 Promise<boolean>
    confirmBranchBeforeExec() {
      return this.openBranchMismatchDialog('exec_confirm')
    },
  },
}
</script>

<style scoped>
.task-workflow-page {
  height: 100vh;
  background: linear-gradient(180deg, #fdfdfb 0%, #f8faf5 100%);
  padding: 16px;
  box-sizing: border-box;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.branch-switch-stream__meta {
  display: grid;
  grid-template-columns: repeat(2, minmax(0, 1fr));
  gap: 8px 16px;
  margin-bottom: 12px;
  color: #606266;
  font-size: 13px;
}

.branch-switch-stream__status {
  display: inline-flex;
  align-items: center;
  gap: 10px;
  margin-bottom: 12px;
  padding: 10px 12px;
  border-radius: 10px;
  background: #eef6ea;
  border: 1px solid #d7e7cf;
  color: #49624a;
  font-size: 13px;
}

.branch-switch-stream__spinner {
  width: 16px;
  height: 16px;
  border-radius: 50%;
  border: 2px solid rgba(73, 98, 74, 0.2);
  border-top-color: #49624a;
  animation: branch-switch-stream-spin 0.8s linear infinite;
  flex: 0 0 auto;
}

.branch-switch-stream__log {
  max-height: 420px;
  overflow: auto;
  background: linear-gradient(180deg, #fcfdf8 0%, #f6f8f2 100%);
  color: #4a5560;
  border: 1px solid #dfe7d6;
  border-radius: 10px;
  padding: 12px 14px;
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.75);
  font-family: Consolas, "Courier New", monospace;
  font-size: 12px;
  line-height: 1.6;
}

.branch-switch-stream__placeholder {
  color: #8b958c;
}

.branch-switch-stream__line {
  white-space: pre-wrap;
  word-break: break-word;
}

.branch-switch-stream__line + .branch-switch-stream__line {
  margin-top: 4px;
}

.branch-switch-stream__line--error {
  color: #c25555;
}

.branch-switch-stream__line--success {
  color: #4d8b57;
}

@keyframes branch-switch-stream-spin {
  from {
    transform: rotate(0deg);
  }
  to {
    transform: rotate(360deg);
  }
}

.task-workflow-shell {
  width: 100%;
  display: flex;
  flex-direction: column;
  flex: 1;
  min-height: 0;
  gap: 12px;
}

.task-workflow-header {
  display: flex;
  justify-content: space-between;
  gap: 16px;
  align-items: flex-start;
  padding: 20px 24px;
  border-radius: 12px;
  background: #fff;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border: 1px solid #e8e8e0;
  flex-shrink: 0;
}

.task-workflow-header__main {
  flex: 1;
  min-width: 0;
}

.task-workflow-header__eyebrow {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.task-workflow-header__title-row {
  display: flex;
  justify-content: space-between;
  align-items: center;
  gap: 16px;
  min-width: 0;
}

.task-workflow-header__title {
  margin: 0;
  font-size: 22px;
  line-height: 1.3;
  color: #303133;
  flex: 1;
  min-width: 0;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-workflow-header__meta {
  display: flex;
  flex-direction: column;
  gap: 8px;
  margin-top: 10px;
}

.task-workflow-header__unread {
  color: #e53935;
  font-size: 14px;
  font-weight: 600;
}

.task-workflow-header__dev-card {
  background: #f9fafb;
  border: 1px solid #e8ecf1;
  border-radius: 8px;
  padding: 12px 16px;
  transition: border-color 0.2s;
}

.task-workflow-header__dev-card:hover {
  border-color: #c8cdd5;
}

.task-workflow-header__field-grid {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(120px, auto));
  gap: 10px 20px;
}

.task-workflow-header__field {
  display: flex;
  flex-direction: column;
  gap: 2px;
  min-width: 0;
  padding: 4px 0;
}

.task-workflow-header__field--compact {
  /* 紧凑模式：按内容宽度自适应，不再设置固定 min-width */
}

.task-workflow-header__field--link {
  cursor: pointer;
}

.task-workflow-header__field--branch {
  /* 分支名列：内容自适应，不再设置固定 min-width */
}

.task-workflow-header__field--link .task-workflow-header__field-value {
  color: #3a7a3a;
  transition: color 0.2s;
}

.task-workflow-header__field--link:hover .task-workflow-header__field-value {
  color: #2d5f2d;
  text-decoration: underline;
}

.task-workflow-header__field-label {
  font-size: 11px;
  font-weight: 600;
  color: #909399;
  text-transform: uppercase;
  letter-spacing: 0.5px;
  user-select: none;
}

.task-workflow-header__field-value {
  font-size: 13px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  display: flex;
  align-items: center;
  gap: 4px;
}

.task-workflow-header__branch {
  color: #3a7a3a;
  cursor: pointer;
  transition: color 0.2s;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

.task-workflow-header__branch:hover {
  color: #2d5f2d;
  text-decoration: underline;
}

.task-workflow-header__status-icon {
  display: inline-flex;
  align-items: center;
  flex-shrink: 0;
  margin-left: 2px;
}

.task-workflow-header__status-icon--ok {
  color: #4caf50;
}

.task-workflow-header__status-icon--err {
  color: #e53935;
}

.task-workflow-header__field--wrap {
  grid-column: span 2;
}

.task-workflow-header__field-value--wrap {
  white-space: normal;
  word-break: break-all;
  overflow: visible;
}

.task-workflow-header__status-icon--remote-warn {
  color: #e6a23c;
  cursor: pointer;
  transition: color 0.2s;
}

.task-workflow-header__status-icon--remote-warn:hover {
  color: #d48806;
}

/* 远程分支检查弹窗 */
.remote-branch-dialog__info {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 16px;
  margin-bottom: 16px;
}

.remote-branch-dialog__row {
  display: flex;
  align-items: center;
  gap: 12px;
  padding: 6px 0;
}

.remote-branch-dialog__row + .remote-branch-dialog__row {
  border-top: 1px solid #ebeef5;
}

.remote-branch-dialog__label {
  font-size: 13px;
  color: #909399;
  min-width: 80px;
  flex-shrink: 0;
}

.remote-branch-dialog__value {
  font-size: 13px;
  color: #303133;
  overflow: hidden;
  text-overflow: ellipsis;
}

.remote-branch-dialog__status {
  display: flex;
  flex-direction: column;
  gap: 12px;
  margin-bottom: 16px;
}

.remote-branch-dialog__status-row {
  display: flex;
  align-items: center;
  gap: 12px;
}

.remote-branch-dialog__status-label {
  font-size: 13px;
  color: #909399;
  min-width: 80px;
  flex-shrink: 0;
}

.remote-branch-dialog__error {
  font-size: 13px;
  color: #e53935;
}

.remote-branch-dialog__result {
  padding: 12px 16px;
  border-radius: 6px;
  font-size: 13px;
  margin-bottom: 8px;
}

.remote-branch-dialog__result--ok {
  background: #f0f9f0;
  color: #2d5f2d;
  border: 1px solid #c8e6c9;
}

.remote-branch-dialog__result--err {
  background: #fef0f0;
  color: #c62828;
  border: 1px solid #ffcdd2;
}

.remote-branch-dialog__footer {
  display: flex;
  gap: 8px;
  justify-content: flex-end;
}

.task-workflow-header__link {
  color: #3a7a3a;
  text-decoration: none;
}

.task-workflow-header__link:hover {
  text-decoration: underline;
}

.task-workflow-header__actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
  align-items: center;
  flex-shrink: 0;
  justify-content: flex-end;
}

.task-workflow-home-btn {
  width: 32px;
  height: 32px;
  border-radius: 50%;
  border: 1px solid #e0e0d8;
  background: #fff;
  color: #909399;
  padding: 0;
  display: inline-flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  transition: border-color 0.2s, color 0.2s;
}

.task-workflow-home-btn:hover {
  border-color: #3a7a3a;
  color: #3a7a3a;
}

.task-workflow-alert {
  margin-bottom: 0;
}

.task-workflow-nodes {
  display: flex;
  gap: 10px;
  flex-shrink: 0;
}

.task-workflow-node {
  border: 1px solid #e8e8e0;
  border-radius: 8px;
  background: #fff;
  min-height: 46px;
  padding: 8px 10px;
  text-align: center;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  flex: 1;
  min-width: 0;
}

.task-workflow-node:hover {
  border-color: #b7c9a8;
  transform: translateY(-1px);
}

.task-workflow-node--active {
  border-color: #3a7a3a;
  background: #f3f8ef;
  box-shadow: 0 6px 18px rgba(58, 122, 58, 0.14);
}

.task-workflow-node--success {
  background: #f3f8ef;
}

.task-workflow-node--failed {
  background: #fff5f4;
}

.task-workflow-node--running {
  background: #fff9ec;
}

.task-workflow-node__index {
  font-size: 12px;
  color: #909399;
}

.task-workflow-node__label {
  font-size: 13px;
  line-height: 1.3;
  color: #303133;
  font-weight: 600;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  max-width: 100%;
}

.task-workflow-node__row {
  display: flex;
  align-items: center;
  gap: 6px;
  max-width: 100%;
  overflow: hidden;
}

.task-workflow-node__badge {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  background: #9fb39a;
  color: #fff;
  font-size: 11px;
  font-weight: 700;
  line-height: 1;
  flex-shrink: 0;
}

.task-workflow-node__desc {
  font-size: 12px;
  line-height: 1.5;
  color: #909399;
  font-weight: 400;
}

.task-workflow-content {
  background: #fff;
  border-radius: 12px;
  padding: 16px 20px 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border: 1px solid #e8e8e0;
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
  gap: 12px;
}

.task-workflow-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.task-workflow-prompt-editor {
  flex: 1;
  height: 100%;
  min-height: 0;
}

.task-workflow-content :deep(.md-editor) {
  flex: 1;
  height: 100%;
  min-height: 0;
}

.task-workflow-content :deep(.md-editor-content) {
  min-height: 0;
}

.task-workflow-content :deep(.md-editor-input-wrapper),
.task-workflow-content :deep(.md-editor-preview-wrapper) {
  overflow: auto;
}

.task-workflow-prompt-editor :deep(.md-editor-input),
.task-workflow-prompt-editor :deep(.md-editor-preview),
.task-workflow-prompt-editor :deep(.md-editor-preview-wrapper),
.task-workflow-issue-fix__editor :deep(.md-editor-input),
.task-workflow-issue-fix__editor :deep(.md-editor-preview),
.task-workflow-issue-fix__editor :deep(.md-editor-preview-wrapper) {
  font-size: 13px;
  line-height: 1.55;
}

/* MdEditor 滚动条绿色 */
.task-workflow-content :deep(.md-editor) {
  --md-scrollbar-bg-color: #edf3e8;
  --md-scrollbar-thumb-color: #9fb39a;
  --md-scrollbar-thumb-hover-color: #869c82;
  --md-scrollbar-thumb-active-color: #7a8f76;
}

.task-workflow-content :deep(.md-editor .md-editor-preview ::-webkit-scrollbar) {
  width: 10px !important;
  height: 10px !important;
}

.task-workflow-content :deep(.md-editor .md-editor-preview ::-webkit-scrollbar-track) {
  background: #edf3e8 !important;
  border-radius: 999px !important;
}

.task-workflow-content :deep(.md-editor .md-editor-preview ::-webkit-scrollbar-thumb) {
  background: #9fb39a !important;
  border: 2px solid #edf3e8 !important;
  border-radius: 999px !important;
}

.task-workflow-content :deep(.md-editor .md-editor-preview ::-webkit-scrollbar-thumb:hover) {
  background: #869c82 !important;
}

.task-workflow-content :deep(.md-editor .md-editor-preview ::-webkit-scrollbar-corner) {
  background: #edf3e8 !important;
}

/* fragment-view 原生滚动条绿色 */
.task-workflow-fragment-view {
  scrollbar-width: thin;
  scrollbar-color: #9fb39a #edf3e8;
}

/* ===== 步骤左侧Tab布局 ===== */
.step-tab-layout {
  display: flex;
  flex: 1;
  min-height: 0;
  overflow: hidden;
  gap: 0;
}

.step-tab-sidebar {
  width: 120px;
  min-width: 120px;
  display: flex;
  flex-direction: column;
  gap: 2px;
  padding: 0 8px 8px 0;
  border-right: 1px solid #e8e8e0;
  flex-shrink: 0;
  overflow-y: auto;
}
.step-tab-sidebar__actions {
  margin-top: auto;
  padding: 8px;
  border: 1px solid #d0d0d0;
  border-radius: 6px;
  display: flex;
  flex-direction: column;
  align-items: center;
  gap: 4px;
}
.step-tab-sidebar__actions :deep(.git-action-button),
.step-tab-sidebar__actions :deep(.el-button),
.step-tab-sidebar__actions :deep(.pl-button),
.step-tab-sidebar__actions .task-workflow-node-status-inline,
.step-tab-sidebar__actions .task-workflow-node-status-inline__btn {
  display: flex !important;
  width: 100%;
  justify-content: center !important;
  text-align: center !important;
}
.step-tab-sidebar__actions :deep(.el-button + .el-button),
.step-tab-sidebar__actions :deep(.git-action-button + .git-action-button),
.step-tab-sidebar__actions :deep(.pl-button + .pl-button) {
  margin-left: 0;
}
.step-tab-sidebar__actions :deep(.el-button .el-button__icon) {
  margin-right: 4px;
}
.step-tab-sidebar__actions :deep(.el-button span) {
  width: auto;
}
.step-tab-sidebar__actions-title {
  font-size: 11px;
  color: #999;
  text-align: center;
  margin-bottom: 2px;
}

.step-tab-btn {
  display: flex;
  align-items: center;
  padding: 8px 10px;
  border: none;
  background: transparent;
  border-radius: 6px;
  cursor: pointer;
  font-size: 13px;
  color: #606266;
  text-align: left;
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
  transition: background 0.15s, color 0.15s;
  line-height: 1.4;
  width: 100%;
}

.step-tab-btn:hover {
  background: #f0f2e8;
}

.step-tab-btn--active {
  background: #e8f0e0;
  color: #3a5a2c;
  font-weight: 600;
}

.step-tab-btn--prompt {
  font-weight: 600;
}

.step-tab-btn--prompt.step-tab-btn--active {
  color: #3a5a2c;
}

.step-tab-btn--doc {
  font-weight: 400;
  font-size: 12px;
}

.step-tab-btn--unread {
  position: relative;
}

.step-tab-btn__dot {
  display: inline-block;
  width: 7px;
  height: 7px;
  border-radius: 50%;
  background: #f56c6c;
  margin-right: 5px;
  flex-shrink: 0;
  align-self: center;
}

.step-tab-content {
  flex: 1;
  min-width: 0;
  min-height: 0;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  padding-left: 16px;
}

.step-tab-panel {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  overflow: hidden;
}

/* ===== 复用 MemoryEditor 工具栏样式 ===== */
.step-tab-panel .editor-body-toolbar {
  display: flex;
  flex-direction: column;
  gap: 10px;
  padding: 12px 16px;
  border-bottom: 1px solid rgba(226, 232, 216, 0.9);
  border-radius: 14px 14px 0 0;
  background: #f8faf5;
  flex-shrink: 0;
}

.step-tab-panel .editor-body-toolbar--saved {
  background: linear-gradient(180deg, #f6faf2 0%, #edf5e7 100%);
  border-bottom-color: rgba(196, 217, 186, 0.9);
}

.step-tab-panel .editor-body-toolbar-main {
  display: flex;
  flex-direction: column;
  align-items: stretch;
  gap: 10px;
  width: 100%;
}

.step-tab-panel .editor-body-toolbar-top {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  width: 100%;
}

.step-tab-panel .editor-body-toolbar-left {
  display: flex;
  align-items: center;
  gap: 12px;
  min-width: 0;
  flex: 1;
}

.step-tab-panel .editor-toolbar-title-input {
  width: 100%;
  flex: 1 1 200px;
}

.step-tab-panel .editor-body-toolbar-right {
  display: flex;
  align-items: center;
  justify-content: flex-end;
  gap: 12px;
  min-width: 0;
  flex-shrink: 0;
}

.step-tab-panel .editor-body-actions {
  display: flex;
  gap: 6px;
  flex-wrap: wrap;
  flex-shrink: 0;
}

.step-tab-panel .editor-body-actions > .el-tooltip__trigger {
  margin: 0;
  padding: 0;
}

.step-tab-panel .editor-body-actions :deep(.toolbar-icon-button),
.step-tab-panel .editor-body-actions :deep(.toolbar-icon-button.git-action-button--compact) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 30px;
  min-width: 30px;
  height: 30px;
  padding: 0 !important;
  line-height: 1;
}

.step-tab-panel .editor-body-actions :deep(.toolbar-icon-button span) {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 100%;
  height: 100%;
}

.step-tab-panel .editor-body-actions :deep(.toolbar-icon-button .el-icon) {
  margin: 0;
  font-size: 15px;
  line-height: 1;
}

.step-tab-panel .editor-action-dropdown {
  margin-left: 2px;
}

.step-tab-panel .editor-action-trigger {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 34px;
  height: 34px;
  border: 1px solid rgba(210, 219, 203, 0.95);
  border-radius: 10px;
  background: #ffffff;
  color: #5d6e57;
  cursor: pointer;
  transition: border-color 0.2s ease, color 0.2s ease, background-color 0.2s ease;
}

.step-tab-panel .editor-action-trigger:hover {
  border-color: #b9c9b1;
  color: #3d5237;
  background: #f7fbf2;
}

.step-tab-panel .mode-button-active {
  position: relative;
}

.step-tab-panel .mode-button-active::after {
  content: '';
  position: absolute;
  inset: -1px;
  border-radius: 8px;
  box-shadow: inset 0 0 0 1px rgba(79, 128, 79, 0.35);
  pointer-events: none;
}

/* ===== 预览区域样式（复用 MemoryEditor） ===== */
.step-tab-panel .preview-body {
  flex: 1;
  min-height: 0;
  display: flex;
  gap: 18px;
  padding: 18px 22px;
  background: #fff;
}

.step-tab-panel .preview-renderer {
  flex: 1;
  height: 100%;
  min-width: 0;
  min-height: 0;
  overflow: auto;
  scrollbar-gutter: stable;
  scrollbar-width: thin;
  scrollbar-color: #9fb39a #edf3e8;
  padding-right: 6px;
}

.step-tab-panel .preview-renderer :deep(.md-editor-preview) {
  font-size: 14px;
  color: #33422f;
  padding: 0 !important;
}

.step-tab-panel .preview-renderer :deep(.md-editor-preview-wrapper) {
  padding: 0 !important;
}

.step-tab-panel .preview-renderer::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

.step-tab-panel .preview-renderer::-webkit-scrollbar-track {
  background: #edf3e8;
  border-radius: 999px;
}

.step-tab-panel .preview-renderer::-webkit-scrollbar-thumb {
  background: #9fb39a;
  border: 2px solid #edf3e8;
  border-radius: 999px;
}

.step-tab-panel .preview-renderer::-webkit-scrollbar-thumb:hover {
  background: #869c82;
}

.step-tab-panel__loading {
  display: flex;
  align-items: center;
  gap: 6px;
  padding: 24px;
  color: #909399;
  font-size: 13px;
}

.step-tab-panel__empty {
  display: flex;
  justify-content: center;
  align-items: center;
  flex: 1;
  min-height: 0;
}

/* ===== 步骤Tab预览区目录导航（复用 MemoryEditor 样式） ===== */
.step-tab-panel .preview-body {
  gap: 18px;
}

.step-tab-panel .preview-body--with-outline {
  align-items: stretch;
}

.step-tab-panel .preview-outline {
  width: 220px;
  flex: 0 0 220px;
  min-width: 0;
}

.step-tab-panel .preview-outline-card {
  position: sticky;
  top: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
  max-height: calc(100vh - 320px);
  overflow: auto;
  padding: 14px 12px;
  border: 1px solid rgba(196, 217, 186, 0.9);
  border-radius: 12px;
  background: linear-gradient(180deg, #fbfcf8 0%, #f6f8f2 100%);
}

.step-tab-panel .preview-outline-title {
  margin-bottom: 4px;
  font-size: 13px;
  font-weight: 700;
  color: #50604a;
}

.step-tab-panel .preview-outline-item {
  width: 100%;
  padding: 8px 10px;
  border: none;
  border-radius: 10px;
  background: transparent;
  color: #4a5b44;
  font-size: 13px;
  cursor: pointer;
  line-height: 1.5;
  text-align: left;
  transition: background-color 0.2s ease, color 0.2s ease;
}

.step-tab-panel .preview-outline-item:hover {
  background: #eef4e8;
  color: #30412d;
}

.step-tab-panel .preview-outline-item.active {
  background: #e6f1de;
  color: #264224;
  font-weight: 600;
}

.step-tab-panel .preview-outline-item--child {
  padding-left: 22px;
}

.step-tab-panel .preview-outline-item--grandchild {
  padding-left: 34px;
}

.step-tab-panel .preview-outline-card::-webkit-scrollbar {
  width: 8px;
  height: 8px;
}

.step-tab-panel .preview-outline-card::-webkit-scrollbar-track {
  background: #edf3e8;
  border-radius: 999px;
}

.step-tab-panel .preview-outline-card::-webkit-scrollbar-thumb {
  background: #9fb39a;
  border: 2px solid #edf3e8;
  border-radius: 999px;
}

.step-tab-panel .preview-outline-card::-webkit-scrollbar-thumb:hover {
  background: #869c82;
}

.task-workflow-content .task-workflow-prompt-editor {
  flex: 1;
  min-height: 0;
}


.task-workflow-fragment-view::-webkit-scrollbar {
  width: 10px;
  height: 10px;
}

.task-workflow-fragment-view::-webkit-scrollbar-track {
  background: #edf3e8;
  border-radius: 999px;
}

.task-workflow-fragment-view::-webkit-scrollbar-thumb {
  background: #9fb39a;
  border: 2px solid #edf3e8;
  border-radius: 999px;
}

.task-workflow-fragment-view::-webkit-scrollbar-thumb:hover {
  background: #869c82;
}

.task-workflow-card__header {
  display: flex;
  justify-content: space-between;
  gap: 12px;
  align-items: center;
  margin-bottom: 12px;
}

.task-workflow-card__title {
  font-size: 16px;
  font-weight: 600;
  color: #303133;
  display: flex;
  align-items: center;
  gap: 8px;
  flex-shrink: 0;
}

.task-workflow-card__title-desc {
  font-size: 12px;
  font-weight: 400;
  color: #909399;
}

.task-workflow-card__hint {
  margin-bottom: 10px;
  font-size: 13px;
  color: #909399;
  word-break: break-all;
}

.task-workflow-card__hint--error {
  color: #c45656;
}

.task-workflow-card__switch {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.task-workflow-summary-list {
  display: flex;
  gap: 10px;
  flex-wrap: wrap;
}

.task-workflow-summary-list--compact {
  margin-bottom: 12px;
}

.task-workflow-summary-item {
  min-width: 140px;
  padding: 12px 14px;
  border-radius: 8px;
  background: #fff;
  border: 1px solid #e8e8e0;
  color: #909399;
  display: flex;
  justify-content: space-between;
  gap: 12px;
}

.task-workflow-summary-item strong {
  color: #303133;
  max-width: 360px;
  text-align: right;
  word-break: break-all;
}

.task-workflow-fragment-view {
  border-radius: 10px;
  border: 1px solid #e8e8e0;
  background: #fff;
  overflow: auto;
  min-height: 0;
  flex: 1;
}

.task-workflow-inner-tabs {
  display: flex;
  gap: 4px;
}

.task-workflow-inner-tab {
  padding: 4px 12px;
  font-size: 13px;
  border: 1px solid #e8e8e0;
  border-radius: 6px;
  background: #fff;
  color: #606266;
  cursor: pointer;
  transition: all 0.2s ease;
}

.task-workflow-inner-tab:hover {
  border-color: #b7c9a8;
  color: #3a7a3a;
}

.task-workflow-inner-tab--active {
  background: #3a7a3a;
  color: #fff;
  border-color: #3a7a3a;
}

.task-workflow-tapd-fetch-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
}

.task-workflow-prompt-section {
  flex: 1;
  display: flex;
  flex-direction: column;
  min-height: 0;
  height: 100%;
}

.task-workflow-fragment-view__iframe {
  width: 100%;
  height: 100%;
  min-height: 520px;
  border: 0;
  display: block;
}

.task-workflow-fragment-view__empty {
  min-height: 520px;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
  font-size: 13px;
}

.task-workflow-config-card {
  overflow: auto;
}

.task-workflow-config-content {
  display: flex;
  flex-direction: column;
  gap: 16px;
}

.task-workflow-config-section__title {
  font-size: 14px;
  font-weight: 600;
  color: #303133;
  margin-bottom: 8px;
}

.task-workflow-config-dev {
  margin-bottom: 12px;
}

.task-workflow-config-dev__index {
  font-size: 13px;
  font-weight: 600;
  color: #3a7a3a;
  margin-bottom: 6px;
}

.task-workflow-config-link {
  color: #3a7a3a;
  text-decoration: none;
  word-break: break-all;
}

.task-workflow-config-link:hover {
  text-decoration: underline;
}

/* IDE 打开按钮 - header 区域 */
.task-workflow-header__open-editor-btn {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 3px;
  background: #ecf5ff;
  color: #409eff;
  font-size: 10px;
  cursor: pointer;
  vertical-align: middle;
  transition: background 0.2s;
}
.task-workflow-header__open-editor-btn:hover {
  background: #d9ecff;
}

/* IDE 打开按钮 - config 区域 */
.task-workflow-config__open-editor-btn {
  color: #409eff;
  font-size: 12px;
  cursor: pointer;
  white-space: nowrap;
}
.task-workflow-config__open-editor-btn:hover {
  color: #337ecc;
}

@media (max-width: 1100px) {
  .task-workflow-nodes {
    flex-wrap: wrap;
  }

  .task-workflow-grid {
    grid-template-columns: repeat(2, minmax(0, 1fr));
  }

}

@media (max-width: 900px) {
  .task-workflow-page {
    padding: 12px;
  }

  .task-workflow-header {
    flex-direction: column;
    padding: 16px;
  }

  .task-workflow-header__title-row {
    flex-direction: column;
    align-items: flex-start;
  }

  .task-workflow-header__title {
    width: 100%;
    white-space: normal;
  }

  .task-workflow-card__header {
    flex-direction: column;
    align-items: flex-start;
  }
}

/* 节点状态图标 */
.task-workflow-node__status-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  flex-shrink: 0;
}

.status-icon {
  display: inline-flex;
  align-items: center;
  justify-content: center;
  width: 20px;
  height: 20px;
  border-radius: 50%;
  font-size: 12px;
  font-weight: 700;
}

.status-icon--completed {
  background: #67c23a;
  color: #fff;
}

.status-icon--skipped {
  background: #e6a23c;
  color: #fff;
}

.status-icon--pending {
  background: #909399;
  width: 14px;
  height: 14px;
}

.status-icon--running {
  background: transparent;
  width: 20px;
  height: 20px;
}

.spinner-ring {
  display: inline-block;
  width: 16px;
  height: 16px;
  border: 2px solid #409eff;
  border-top-color: transparent;
  border-radius: 50%;
  animation: status-icon-spin 0.8s linear infinite;
}

@keyframes status-icon-spin {
  to { transform: rotate(360deg); }
}
/* 节点按钮状态边框色 */
.task-workflow-node--status-pending {
  border-top: 3px solid #909399;
}
.task-workflow-node--status-completed {
  border-top: 3px solid #67c23a;
}

.task-workflow-node--status-skipped {
  border-top: 3px solid #e6a23c;
}

.task-workflow-node--status-running {
  border-top: 3px solid #409eff;
}

/* 节点状态切换（内联，位于还原为默认提示词右侧） */
.task-workflow-node-status-inline {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-left: auto;
  flex-shrink: 0;
}

.task-workflow-node-status-inline__label {
  font-size: 13px;
  color: #909399;
  white-space: nowrap;
}

.task-workflow-node-status-inline__btn {
  display: inline-flex;
  align-items: center;
  gap: 4px;
  padding: 4px 14px;
  font-size: 13px;
  font-weight: 600;
  border: 1px solid;
  border-radius: 6px;
  cursor: pointer;
  transition: all 0.2s ease;
}

.task-workflow-node-status-inline__btn:disabled {
  opacity: 0.5;
  cursor: not-allowed;
}

.task-workflow-node-status-inline__btn--completed {
  background: #f0f9eb;
  color: #67c23a;
  border-color: #c2e7b0;
}

.task-workflow-node-status-inline__btn--completed:hover:not(:disabled) {
  background: #e1f3d8;
}

.task-workflow-node-status-inline__btn--skipped {
  background: #fdf6ec;
  color: #e6a23c;
  border-color: #f5dab1;
}

.task-workflow-node-status-inline__btn--skipped:hover:not(:disabled) {
  background: #faecd8;
}

.task-workflow-node-status-inline__btn--running {
  background: #ecf5ff;
  color: #409eff;
  border-color: #b3d8ff;
}

.task-workflow-node-status-inline__btn--running:hover:not(:disabled) {
  background: #d9ecff;
}

.task-workflow-node-status-inline__btn--pending {
  background: #f4f4f5;
  color: #909399;
  border-color: #d4d4d8;
}

.task-workflow-node-status-inline__btn--pending:hover:not(:disabled) {
  background: #e9e9eb;
}

/* 思考过程文本区 */
.thinking-blockquote {
  white-space: pre-wrap;
  font-size: 12px;
  color: #606266;
  border-left: 3px solid #dcdfe6;
  background: #f5f7fa;
  padding: 8px 12px;
  margin: 0;
  border-radius: 0 4px 4px 0;
  max-height: 150px;
  overflow-y: auto;
}

.thinking-blockquote::-webkit-scrollbar {
  width: 6px;
}

.thinking-blockquote::-webkit-scrollbar-track {
  background: #f0f0f0;
  border-radius: 3px;
}

.thinking-blockquote::-webkit-scrollbar-thumb {
  background: #c0c4cc;
  border-radius: 3px;
}

.thinking-blockquote::-webkit-scrollbar-thumb:hover {
  background: #909399;
}

/* ===== 文件变更样式（header 列的统计标签，详情弹窗样式在 FileChangesDialog.vue 中） ===== */
.task-workflow-header__field-value--dim {
  color: #c0c4cc;
  font-style: italic;
}

/* 文件变更内联统计 */
.file-changes-inline {
  display: inline-flex;
  align-items: center;
  gap: 0;
  font-size: 12px;
  white-space: nowrap;
}

.file-changes-inline--clickable {
  cursor: pointer;
  padding: 2px 6px;
  border-radius: 4px;
  transition: background-color 0.15s;
}

.file-changes-inline--clickable:hover {
  background-color: #f0f2f5;
}

.file-changes-inline__item {
  font-weight: 700;
  font-size: 13px;
}

.file-changes-inline__item small {
  font-weight: 400;
  font-size: 10px;
  margin-left: 1px;
  opacity: 0.7;
}

.file-changes-inline__item--committed {
  color: #1a7f37;
}

.file-changes-inline__item--staged {
  color: #0366d6;
}

.file-changes-inline__item--modified {
  color: #9a6700;
}

.file-changes-inline__item--untracked {
  color: #656d76;
}

.file-changes-inline__sep {
  color: #c0c4cc;
  margin: 0 5px;
  font-size: 11px;
}

/* 文档编辑模式下的编辑器容器 */
.task-workflow-doc-editor {
  flex: 1;
  min-height: 0;
}

</style>

<style>
/* 知识片段弹窗 — 非 scoped，因为 el-dialog 被 teleport 到 body */
.task-workflow-fragment-dialog .el-dialog__body {
  padding: 0 !important;
  height: calc(90vh - 54px) !important;
  overflow: hidden !important;
}

.task-workflow-fragment-dialog__body {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.task-workflow-fragment-dialog__iframe {
  width: 100%;
  height: 100%;
  border: 0;
  display: block;
  flex: 1;
}

.task-workflow-fragment-dialog__empty {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
  font-size: 14px;
}

.task-workflow-issue-fix__close-bar {
  display: flex;
  justify-content: flex-end;
  margin-bottom: 12px;
}
.task-workflow-issue-fix__input {
  margin-bottom: 16px;
}
.task-workflow-issue-fix__label {
  font-size: 14px;
  font-weight: 500;
  color: #303133;
  margin-bottom: 8px;
}
.task-workflow-issue-fix__output {
  margin-top: 16px;
}
.task-workflow-issue-fix__editor {
  height: 400px;
}

.task-workflow-issue-fix-dialog .el-dialog__body {
  max-height: calc(90vh - 54px);
  overflow-y: auto;
}

.task-workflow-issue-fix-dialog .el-dialog {
  max-height: 90vh;
  overflow: hidden;
}

/* 问题修改提示词弹窗：开关文字颜色跟随主题色 */
.task-workflow-issue-fix-dialog .el-switch__label.is-active {
  color: #5a8a5a;
}

.chat-result-section {
  margin-top: 8px;
}
.chat-result-section-title {
  color: #909399;
  font-size: 11px;
  margin-bottom: 4px;
}
.chat-result-tokens {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  color: #606266;
}
.chat-result-model-row {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  padding: 4px 8px;
  background: #f2f3f5;
  border-radius: 4px;
  margin-bottom: 2px;
}
.chat-result-model-name {
  font-weight: 600;
  color: #303133;
}
.chat-result-permission-item {
  padding: 2px 8px;
  font-size: 11px;
}
.chat-result-text {
  white-space: pre-wrap;
  font-size: 11px;
  max-height: 120px;
  overflow-y: auto;
  font-family: Consolas, monospace;
  background: #f2f3f5;
  padding: 6px 8px;
  border-radius: 4px;
  margin: 0;
  color: #606266;
}

/* 接口开发弹窗 */
.task-workflow-api-dev-dialog .el-dialog__body {
  padding: 0 !important;
  height: calc(96vh - 54px) !important;
  overflow: hidden !important;
}

.task-workflow-api-dev-dialog__body {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.task-workflow-api-dev-dialog__iframe {
  width: 100%;
  height: 100%;
  border: 0;
  display: block;
  flex: 1;
}

.task-workflow-api-dev-dialog__empty {
  height: 100%;
  display: flex;
  align-items: center;
  justify-content: center;
  color: #909399;
  font-size: 14px;
}

/* 对话输入框样式（非 scoped，对话框内容被 teleport 到 body） */
.chat-detail-textarea .el-textarea__inner {
  border-color: #dcdfe6 !important;
  transition: border-color 0.2s;
  resize: none;
}

.chat-detail-textarea .el-textarea__inner:focus {
  border-color: #409eff !important;
  box-shadow: 0 0 0 1px #409eff inset !important;
}

.branch-mismatch-dialog__summary {
  display: flex;
  gap: 16px;
  margin-bottom: 16px;
  color: #4b5a54;
  font-size: 13px;
}

.branch-mismatch-dialog__empty {
  padding: 20px 0;
  color: #909399;
  text-align: center;
}

.branch-mismatch-card {
  border: 1px solid #dbe7df;
  border-radius: 12px;
  padding: 14px;
  background: #fbfcfa;
}

.branch-mismatch-card + .branch-mismatch-card {
  margin-top: 12px;
}

.branch-mismatch-card__header,
.branch-mismatch-dialog__footer {
  display: flex;
  align-items: center;
  justify-content: space-between;
}

.branch-mismatch-card__title {
  font-size: 14px;
  font-weight: 600;
  color: #20312b;
  word-break: break-all;
}

.branch-mismatch-card__meta {
  display: grid;
  gap: 6px;
  margin-top: 10px;
  color: #4b5a54;
  font-size: 13px;
}

.branch-mismatch-card__changes {
  margin-top: 12px;
}

.branch-mismatch-card__changes-title {
  margin-bottom: 8px;
  color: #20312b;
  font-size: 13px;
  font-weight: 600;
}

.branch-mismatch-card__file-list {
  max-height: 180px;
  overflow: auto;
  padding: 10px;
  border-radius: 8px;
  background: #f3f7f4;
  border: 1px solid #e0e8e2;
}

.branch-mismatch-card__file-item {
  font-family: Consolas, monospace;
  font-size: 12px;
  line-height: 1.5;
  color: #2a3933;
  word-break: break-all;
}

.branch-mismatch-card__empty-tip,
.branch-mismatch-card__error {
  margin-top: 6px;
  color: #8a4d24;
  font-size: 12px;
  line-height: 1.5;
}

.branch-mismatch-card__actions {
  margin-top: 12px;
  display: flex;
  justify-content: flex-end;
}

/* 文档搜索结果高亮（非 scoped，由 JavaScript 动态添加到 DOM） */
.doc-search-highlight {
  background-color: #ffd54f;
  color: #000;
  padding: 1px 0;
  border-radius: 2px;
}
</style>
