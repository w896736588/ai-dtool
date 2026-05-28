<template>
  <div class="task-workflow-page" v-loading="loading">
    <div class="task-workflow-shell">
      <header class="task-workflow-header">
        <div class="task-workflow-header__main">
          <div class="task-workflow-header__title-row">
            <h1 class="task-workflow-header__title">{{ homeTask.name || `任务 #${taskId}` }}</h1>
            <div class="task-workflow-header__actions">
          <el-tooltip content="返回首页" placement="bottom">
            <el-button class="task-workflow-home-btn" @click="goHome">
              <el-icon :size="18"><HomeFilled /></el-icon>
            </el-button>
          </el-tooltip>
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
            <div v-for="(cfg, idx) in parsedTaskDevConfigs" :key="idx" class="task-workflow-header__dev-row">
              <span class="task-workflow-header__dev-item">Git仓库: {{ getTaskConfigName('git', cfg.git_id) }}</span>
              <span class="task-workflow-header__dev-sep">|</span>
              <span class="task-workflow-header__dev-item task-workflow-header__dev-item--link" @click="openApiDevDialog(cfg)">接口集合: {{ truncateWorkflowLabel(getTaskConfigApiLabel(cfg)) }}</span>
              <span class="task-workflow-header__dev-sep">|</span>
              <span class="task-workflow-header__dev-item">父分支: {{ cfg.parent_branch || '-' }}</span>
              <span class="task-workflow-header__dev-sep">|</span>
              <span class="task-workflow-header__dev-item">分支名: <span class="task-workflow-header__branch" @click="copyText(cfg.branch_name, '分支名已复制')" :title="cfg.branch_name">{{ truncateWorkflowLabel(cfg.branch_name || '-') }}</span><el-tooltip v-if="cfg.local_dir && cfg.branch_name && branchStatusMap[cfg.local_dir + '|' + cfg.branch_name]" :content="branchStatusMap[cfg.local_dir + '|' + cfg.branch_name].matched ? '分支匹配' : '当前分支: ' + (branchStatusMap[cfg.local_dir + '|' + cfg.branch_name].current_branch || '未知')" placement="top"><span style="display: inline-flex; align-items: center; vertical-align: middle; margin-left: 4px;" :style="{ color: branchStatusMap[cfg.local_dir + '|' + cfg.branch_name].matched ? '#4caf50' : '#e53935' }"><svg v-if="branchStatusMap[cfg.local_dir + '|' + cfg.branch_name].matched" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg><svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg></span></el-tooltip></span>
              <span class="task-workflow-header__dev-sep">|</span>
              <span class="task-workflow-header__dev-item">本地目录: {{ cfg.local_dir || '-' }}<el-tooltip v-if="cfg.local_dir && localDirStatusMap[cfg.local_dir] !== undefined" :content="localDirStatusMap[cfg.local_dir] ? '目录存在' : '目录不存在'" placement="top"><span style="display: inline-flex; align-items: center; vertical-align: middle; margin-left: 4px;" :style="{ color: localDirStatusMap[cfg.local_dir] ? '#4caf50' : '#e53935' }"><svg v-if="localDirStatusMap[cfg.local_dir]" viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><path d="M20 6L9 17l-5-5"/></svg><svg v-else viewBox="0 0 24 24" width="14" height="14" fill="none" stroke="currentColor" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="10"/><line x1="12" y1="8" x2="12" y2="12"/><line x1="12" y1="16" x2="12.01" y2="16"/></svg></span></el-tooltip></span>
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

      <section class="task-workflow-nodes">
        <button
          v-for="node in workflowNodes"
          :key="node.key"
          type="button"
          class="task-workflow-node"
          :class="{
            'task-workflow-node--active': activeNode === node.key,
            'task-workflow-node--success': node.key === 'requirement-fetch' && activeNode === node.key && requirementFetchStatus === 'success',
            'task-workflow-node--failed': node.key === 'requirement-fetch' && activeNode === node.key && requirementFetchStatus === 'failed',
            'task-workflow-node--running': node.key === 'requirement-fetch' && activeNode === node.key && requirementFetchStatus === 'running',
            'task-workflow-node--status-pending': getNodeStatus(node.key) === 'pending',
            'task-workflow-node--status-running': getNodeStatus(node.key) === 'running',
            'task-workflow-node--status-completed': getNodeStatus(node.key) === 'completed',
            'task-workflow-node--status-skipped': getNodeStatus(node.key) === 'skipped',
          }"
          @click="selectNode(node.key)"
        >
          <span class="task-workflow-node__status-icon">
            <span v-if="getNodeStatus(node.key) === 'completed'" class="status-icon status-icon--completed">&#10003;</span>
            <span v-else-if="getNodeStatus(node.key) === 'skipped'" class="status-icon status-icon--skipped">&#10003;</span>
            <span v-else-if="getNodeStatus(node.key) === 'pending'" class="status-icon status-icon--pending"></span>
            <span v-else class="status-icon status-icon--running"><span class="spinner-ring"></span></span>
          </span>
          <span class="task-workflow-node__label">{{ node.label }}</span>
          <span class="task-workflow-node__desc">{{ node.desc }}</span>
        </button>
      </section>

      <section class="task-workflow-content">
        <div v-if="activeNode === 'task-config'" class="task-workflow-tab">
          <div class="task-workflow-card task-workflow-config-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">任务配置</div>
            </div>
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
                  <el-descriptions-item label="TAPD地址">
                    <a v-if="homeTask.tapd_url" :href="homeTask.tapd_url" target="_blank" class="task-workflow-config-link">{{ homeTask.tapd_url }}</a>
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
                    <el-descriptions-item label="本地目录">{{ cfg.local_dir || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="父分支">{{ cfg.parent_branch || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="分支名">{{ truncateWorkflowLabel(cfg.branch_name || '-') }}</el-descriptions-item>
                    <el-descriptions-item label="规则入口">{{ cfg.rule_entry_file || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="自定义网页">{{ getTaskConfigName('smart_link', cfg.smart_link_id) }}</el-descriptions-item>
                    <el-descriptions-item label="网页标签">{{ cfg.smart_link_label || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="账号">{{ cfg.smart_link_account || '-' }}</el-descriptions-item>
                  </el-descriptions>
                </div>
              </div>
            </div>
            <div class="task-workflow-config-section">
              <div class="task-workflow-config-section__title">关联知识片段</div>
              <el-table :data="workflowFragments" border size="small" empty-text="暂无关联知识片段">
                <el-table-column label="片段类型" prop="label" width="180" />
                <el-table-column label="片段ID" prop="id" width="120">
                  <template #default="{ row }">
                    <span v-if="row.id">{{ row.id }}</span>
                    <span v-else class="task-workflow-config-hint">未绑定</span>
                  </template>
                </el-table-column>
                <el-table-column label="操作" width="100">
                  <template #default="{ row }">
                    <el-button v-if="row.id" size="small" text type="primary" @click="openFragmentInDialog(row.id, row.label)">
                      <el-icon><Link /></el-icon>
                      打开
                    </el-button>
                  </template>
                </el-table-column>
              </el-table>
            </div>
          </div>
        </div>

        <div v-else-if="activeNode === 'requirement-fetch'" class="task-workflow-tab">
          <div class="task-workflow-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">抓取 TAPD 需求</div>
              <div class="task-workflow-card__switch">
                <div class="task-workflow-inner-tabs">
                  <button
                    :class="['task-workflow-inner-tab', { 'task-workflow-inner-tab--active': requirementFetchActiveTab === 'tapd-fetch' }]"
                    @click="requirementFetchActiveTab = 'tapd-fetch'"
                  >抓取 TAPD 需求内容</button>
                  <button
                    :class="['task-workflow-inner-tab', { 'task-workflow-inner-tab--active': requirementFetchActiveTab === 'plain-text-prompt' }]"
                    @click="requirementFetchActiveTab = 'plain-text-prompt'"
                  >纯文本需求提示词</button>
                </div>
              </div>
            </div>

            <div v-show="requirementFetchActiveTab === 'tapd-fetch'" class="task-workflow-tapd-fetch-section">
              <div class="task-workflow-card__switch" style="margin-bottom: 12px;">
                <GitActionButton compact :loading="requirementFetchRunning" @click="triggerRequirementFetch(false)">
                  重新抓取
                </GitActionButton>
                <GitActionButton compact variant="info" @click="openFragmentInDialog(requirementFragmentId, 'TAPD需求文档')" :disabled="!requirementFragmentId">
                  <template #icon><el-icon><Link /></el-icon></template>
                  打开知识片段
                </GitActionButton>
              </div>
              <div v-if="workflow.requirement_fetch_error" class="task-workflow-card__hint task-workflow-card__hint--error">
                最近错误：{{ workflow.requirement_fetch_error }}
              </div>
              <div v-if="!homeTask.tapd_url" class="task-workflow-card__hint">
                当前任务未配置 TAPD 地址，无法自动抓取。
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
            </div>

            <div v-show="requirementFetchActiveTab === 'plain-text-prompt'" class="task-workflow-prompt-section">
              <div class="task-workflow-card__switch" style="margin-bottom: 12px;">
                <GitActionButton compact :loading="promptSaving === 'plain_text_requirement'" @click="savePrompts('plain_text_requirement')">
                  保存提示词
                </GitActionButton>
                <GitActionButton compact @click="copyText(workflow.prompt_plain_text_requirement || '', '提示词已复制')">
                  <template #icon><el-icon><CopyDocument /></el-icon></template>
                  复制提示词
                </GitActionButton>
                <GitActionButton compact variant="success" @click="openPromptExecDialog('plain_text_requirement', workflow.prompt_plain_text_requirement || '')">
                  <template #icon><el-icon><VideoPlay /></el-icon></template>
                  执行
                </GitActionButton>
                <ChatHistoryButton
                  compact
                  variant="info"
                  :running="getPromptChatCounts('plain_text_requirement').running > 0"
                  :running-count="getPromptChatCounts('plain_text_requirement').running"
                  :interrupted-count="getPromptChatCounts('plain_text_requirement').interrupted"
                  :total-count="getPromptChatCounts('plain_text_requirement').total"
                  @click="openPromptChatHistory('plain_text_requirement')"
                >
                  执行历史
                </ChatHistoryButton>
                <GitActionButton compact variant="warning" :loading="promptRestoring === 'plain_text_requirement'" @click="restorePrompts('plain_text_requirement')">
                  还原为默认提示词
                </GitActionButton>
                <GitActionButton compact variant="info" @click="openFragmentInDialog(plainTextReqFragmentId, '纯文本需求文档')" :disabled="!plainTextReqFragmentId">
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
              <MdEditor
                v-model="workflow.prompt_plain_text_requirement"
                class="task-workflow-prompt-editor"
                preview-theme="github"
                :preview="true"
                :toolbars="promptEditorToolbars"
                height="100%"
              />
            </div>
          </div>
        </div>

        <div v-else-if="activeNode === 'requirement'" class="task-workflow-tab">
          <div class="task-workflow-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">需求分析</div>
              <div class="task-workflow-card__switch">
                <div class="task-workflow-inner-tabs">


                </div>
              </div>
            </div>

            <div v-show="requirementActiveTab === 'requirement-prompt'" class="task-workflow-prompt-section">
              <div class="task-workflow-card__switch" style="margin-bottom: 12px;">
                <GitActionButton compact variant="info" @click="openFragmentInDialog(designPlanReqFragmentId, '需求设计方案文档')" :disabled="!designPlanReqFragmentId">
                  <template #icon><el-icon><Link /></el-icon></template>
                  需求设计方案文档
                </GitActionButton>
                <GitActionButton compact :loading="promptSaving === 'requirement'" @click="savePrompts('requirement')">
                  保存提示词
                </GitActionButton>
                <GitActionButton compact @click="copyText(workflow.prompt_requirement || '', '提示词已复制')">
                  <template #icon><el-icon><CopyDocument /></el-icon></template>
                  复制提示词
                </GitActionButton>
                <GitActionButton compact variant="success" @click="openPromptExecDialog('requirement', workflow.prompt_requirement || '')">
                  <template #icon><el-icon><VideoPlay /></el-icon></template>
                  执行
                </GitActionButton>
                <ChatHistoryButton
                  compact
                  variant="info"
                  :running="getPromptChatCounts('requirement').running > 0"
                  :running-count="getPromptChatCounts('requirement').running"
                  :interrupted-count="getPromptChatCounts('requirement').interrupted"
                  :total-count="getPromptChatCounts('requirement').total"
                  @click="openPromptChatHistory('requirement')"
                >
                  执行历史
                </ChatHistoryButton>
                <GitActionButton compact variant="warning" :loading="promptRestoring === 'requirement'" @click="restorePrompts('requirement')">
                  还原为默认提示词
                </GitActionButton>
                <div class="task-workflow-node-status-inline">
                  <span class="task-workflow-node-status-inline__label">当前步骤状态</span>
                  <button
                    class="task-workflow-node-status-inline__btn"
                    :class="'task-workflow-node-status-inline__btn--' + getNodeStatus('requirement')"
                    :disabled="nodeStatusSaving"
                    @click="cycleNodeStatus('requirement')"
                  >{{ getNodeStatusLabel('requirement') }}</button>
                </div>
              </div>
              <div class="task-workflow-card__hint">
                当前片段：{{ requirementFragmentTitle }}
              </div>
              <MdEditor
                v-model="workflow.prompt_requirement"
                class="task-workflow-prompt-editor"
                preview-theme="github"
                :preview="true"
                :toolbars="promptEditorToolbars"
                height="100%"
              />
            </div>

            <div v-show="requirementActiveTab === 'design-plan-prompt'" class="task-workflow-prompt-section">
              <div class="task-workflow-card__switch" style="margin-bottom: 12px;">
                <GitActionButton compact :loading="promptSaving === 'design_plan_requirement'" @click="savePrompts('design_plan_requirement')">
                  保存提示词
                </GitActionButton>
                <GitActionButton compact @click="copyText(workflow.prompt_design_plan_requirement || '', '提示词已复制')">
                  <template #icon><el-icon><CopyDocument /></el-icon></template>
                  复制提示词
                </GitActionButton>
                <GitActionButton compact variant="success" @click="openPromptExecDialog('design_plan_requirement', workflow.prompt_design_plan_requirement || '')">
                  <template #icon><el-icon><VideoPlay /></el-icon></template>
                  执行
                </GitActionButton>
                <ChatHistoryButton
                  compact
                  variant="info"
                  :running="getPromptChatCounts('design_plan_requirement').running > 0"
                  :running-count="getPromptChatCounts('design_plan_requirement').running"
                  :interrupted-count="getPromptChatCounts('design_plan_requirement').interrupted"
                  :total-count="getPromptChatCounts('design_plan_requirement').total"
                  @click="openPromptChatHistory('design_plan_requirement')"
                >
                  执行历史
                </ChatHistoryButton>
                <GitActionButton compact variant="warning" :loading="promptRestoring === 'design_plan_requirement'" @click="restorePrompts('design_plan_requirement')">
                  还原为默认提示词
                </GitActionButton>
                <GitActionButton compact variant="info" @click="openFragmentInDialog(designPlanReqFragmentId, '需求设计方案文档')" :disabled="!designPlanReqFragmentId">
                  <template #icon><el-icon><Link /></el-icon></template>
                  打开知识片段
                </GitActionButton>
                <div class="task-workflow-node-status-inline">
                  <span class="task-workflow-node-status-inline__label">当前步骤状态</span>
                  <button
                    class="task-workflow-node-status-inline__btn"
                    :class="'task-workflow-node-status-inline__btn--' + getNodeStatus('requirement')"
                    :disabled="nodeStatusSaving"
                    @click="cycleNodeStatus('requirement')"
                  >{{ getNodeStatusLabel('requirement') }}</button>
                </div>
              </div>
              <MdEditor
                v-model="workflow.prompt_design_plan_requirement"
                class="task-workflow-prompt-editor"
                preview-theme="github"
                :preview="true"
                :toolbars="promptEditorToolbars"
                height="100%"
              />
            </div>
          </div>
        </div>

        <div v-else-if="activeNode === 'design'" class="task-workflow-tab">
          <div class="task-workflow-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">开发提示词</div>
              <div class="task-workflow-card__switch">
                <GitActionButton compact :loading="promptSaving === 'design'" @click="savePrompts('design')">
                  保存提示词
                </GitActionButton>
                <GitActionButton compact @click="copyText(workflow.prompt_design || '', '提示词已复制')">
                  <template #icon><el-icon><CopyDocument /></el-icon></template>
                  复制提示词
                </GitActionButton>
                <GitActionButton compact variant="success" @click="openPromptExecDialog('design', workflow.prompt_design || '')">
                  <template #icon><el-icon><VideoPlay /></el-icon></template>
                  执行
                </GitActionButton>
                <ChatHistoryButton
                  compact
                  variant="info"
                  :running="getPromptChatCounts('design').running > 0"
                  :running-count="getPromptChatCounts('design').running"
                  :interrupted-count="getPromptChatCounts('design').interrupted"
                  :total-count="getPromptChatCounts('design').total"
                  @click="openPromptChatHistory('design')"
                >
                  执行历史
                </ChatHistoryButton>
                <GitActionButton compact variant="warning" :loading="promptRestoring === 'design'" @click="restorePrompts('design')">
                  还原为默认提示词
                </GitActionButton>
                <div class="task-workflow-node-status-inline">
                  <span class="task-workflow-node-status-inline__label">当前步骤状态</span>
                  <button
                    class="task-workflow-node-status-inline__btn"
                    :class="'task-workflow-node-status-inline__btn--' + getNodeStatus('design')"
                    :disabled="nodeStatusSaving"
                    @click="cycleNodeStatus('design')"
                  >{{ getNodeStatusLabel('design') }}</button>
                </div>
              </div>
            </div>
            <MdEditor
              v-model="workflow.prompt_design"
              class="task-workflow-prompt-editor"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              height="100%"
            />
          </div>
        </div>

        <div v-else-if="activeNode === 'api-dev'" class="task-workflow-tab">
          <div class="task-workflow-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">接口开发生成提示词</div>
              <div class="task-workflow-card__switch">
                <GitActionButton compact @click="openFragmentInDialog(workflow.api_doc_fragment_id, '接口文档')">
                  <template #icon><el-icon><Link /></el-icon></template>
                  接口文档
                </GitActionButton>
                <GitActionButton compact variant="warning" :loading="apiDocResetting" @click="resetApiDoc">
                  重置接口文档
                </GitActionButton>
                <GitActionButton compact :loading="promptSaving === 'api_dev'" @click="savePrompts('api_dev')">
                  保存提示词
                </GitActionButton>
                <GitActionButton compact @click="copyText(workflow.prompt_api_dev || '', '提示词已复制')">
                  <template #icon><el-icon><CopyDocument /></el-icon></template>
                  复制提示词
                </GitActionButton>
                <GitActionButton compact variant="success" @click="openPromptExecDialog('api_dev', workflow.prompt_api_dev || '')">
                  <template #icon><el-icon><VideoPlay /></el-icon></template>
                  执行
                </GitActionButton>
                <ChatHistoryButton compact variant="info" :running="getPromptChatCounts('api_dev').running > 0" :running-count="getPromptChatCounts('api_dev').running" :interrupted-count="getPromptChatCounts('api_dev').interrupted" :total-count="getPromptChatCounts('api_dev').total" @click="openPromptChatHistory('api_dev')">
                  执行历史
                </ChatHistoryButton>
                <GitActionButton compact variant="warning" :loading="promptRestoring === 'api_dev'" @click="restorePrompts('api_dev')">
                  还原为默认提示词
                </GitActionButton>
                <div class="task-workflow-node-status-inline">
                  <span class="task-workflow-node-status-inline__label">当前步骤状态</span>
                  <button
                    class="task-workflow-node-status-inline__btn"
                    :class="'task-workflow-node-status-inline__btn--' + getNodeStatus('api-dev')"
                    :disabled="nodeStatusSaving"
                    @click="cycleNodeStatus('api-dev')"
                  >{{ getNodeStatusLabel('api-dev') }}</button>
                </div>
              </div>
            </div>
            <MdEditor
              v-model="workflow.prompt_api_dev"
              class="task-workflow-prompt-editor"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              height="100%"
            />
          </div>
        </div>

        <div v-else-if="activeNode === 'code-review'" class="task-workflow-tab">
          <div class="task-workflow-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">代码检查提示词</div>
              <div class="task-workflow-card__switch">
                <GitActionButton compact :loading="promptSaving === 'code_review'" @click="savePrompts('code_review')">
                  保存提示词
                </GitActionButton>
                <GitActionButton compact @click="copyText(workflow.prompt_code_review || '', '提示词已复制')">
                  <template #icon><el-icon><CopyDocument /></el-icon></template>
                  复制提示词
                </GitActionButton>
                <GitActionButton compact variant="success" @click="openPromptExecDialog('code_review', workflow.prompt_code_review || '')">
                  <template #icon><el-icon><VideoPlay /></el-icon></template>
                  执行
                </GitActionButton>
                <ChatHistoryButton compact variant="info" :running="getPromptChatCounts('code_review').running > 0" :running-count="getPromptChatCounts('code_review').running" :interrupted-count="getPromptChatCounts('code_review').interrupted" :total-count="getPromptChatCounts('code_review').total" @click="openPromptChatHistory('code_review')">
                  执行历史
                </ChatHistoryButton>
                <GitActionButton compact variant="warning" :loading="promptRestoring === 'code_review'" @click="restorePrompts('code_review')">
                  还原为默认提示词
                </GitActionButton>
                <div class="task-workflow-node-status-inline">
                  <span class="task-workflow-node-status-inline__label">当前步骤状态</span>
                  <button
                    class="task-workflow-node-status-inline__btn"
                    :class="'task-workflow-node-status-inline__btn--' + getNodeStatus('code-review')"
                    :disabled="nodeStatusSaving"
                    @click="cycleNodeStatus('code-review')"
                  >{{ getNodeStatusLabel('code-review') }}</button>
                </div>
              </div>
            </div>
            <MdEditor
              v-model="workflow.prompt_code_review"
              class="task-workflow-prompt-editor"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              height="100%"
            />
          </div>
        </div>

        <div v-else-if="activeNode === 'browser-test'" class="task-workflow-tab">
          <div class="task-workflow-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">需求核对浏览器测试提示词</div>
              <div class="task-workflow-card__switch">
                <GitActionButton compact :loading="promptSaving === 'browser_test'" @click="savePrompts('browser_test')">
                  保存提示词
                </GitActionButton>
                <GitActionButton compact @click="copyText(workflow.prompt_browser_test || '', '提示词已复制')">
                  <template #icon><el-icon><CopyDocument /></el-icon></template>
                  复制提示词
                </GitActionButton>
                <GitActionButton compact variant="success" @click="openPromptExecDialog('browser_test', workflow.prompt_browser_test || '')">
                  <template #icon><el-icon><VideoPlay /></el-icon></template>
                  执行
                </GitActionButton>
                <ChatHistoryButton compact variant="info" :running="getPromptChatCounts('browser_test').running > 0" :running-count="getPromptChatCounts('browser_test').running" :interrupted-count="getPromptChatCounts('browser_test').interrupted" :total-count="getPromptChatCounts('browser_test').total" @click="openPromptChatHistory('browser_test')">
                  执行历史
                </ChatHistoryButton>
                <GitActionButton compact variant="warning" :loading="promptRestoring === 'browser_test'" @click="restorePrompts('browser_test')">
                  还原为默认提示词
                </GitActionButton>
                <div class="task-workflow-node-status-inline">
                  <span class="task-workflow-node-status-inline__label">当前步骤状态</span>
                  <button
                    class="task-workflow-node-status-inline__btn"
                    :class="'task-workflow-node-status-inline__btn--' + getNodeStatus('browser-test')"
                    :disabled="nodeStatusSaving"
                    @click="cycleNodeStatus('browser-test')"
                  >{{ getNodeStatusLabel('browser-test') }}</button>
                </div>
              </div>
            </div>
            <MdEditor
              v-model="workflow.prompt_browser_test"
              class="task-workflow-prompt-editor"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              height="100%"
            />
          </div>
        </div>

        <div v-else class="task-workflow-tab">
          <div class="task-workflow-card">
            <div class="task-workflow-card__header">
              <div class="task-workflow-card__title">接口自动化测试修复提示词</div>
              <div class="task-workflow-card__switch">
                <GitActionButton compact :loading="promptSaving === 'api_test'" @click="savePrompts('api_test')">
                  保存提示词
                </GitActionButton>
                <GitActionButton compact @click="copyText(workflow.prompt_api_test || '', '提示词已复制')">
                  <template #icon><el-icon><CopyDocument /></el-icon></template>
                  复制提示词
                </GitActionButton>
                <GitActionButton compact variant="success" @click="openPromptExecDialog('api_test', workflow.prompt_api_test || '')">
                  <template #icon><el-icon><VideoPlay /></el-icon></template>
                  执行
                </GitActionButton>
                <ChatHistoryButton compact variant="info" :running="getPromptChatCounts('api_test').running > 0" :running-count="getPromptChatCounts('api_test').running" :interrupted-count="getPromptChatCounts('api_test').interrupted" :total-count="getPromptChatCounts('api_test').total" @click="openPromptChatHistory('api_test')">
                  执行历史
                </ChatHistoryButton>
                <GitActionButton compact variant="warning" :loading="promptRestoring === 'api_test'" @click="restorePrompts('api_test')">
                  还原为默认提示词
                </GitActionButton>
                <div class="task-workflow-node-status-inline">
                  <span class="task-workflow-node-status-inline__label">当前步骤状态</span>
                  <button
                    class="task-workflow-node-status-inline__btn"
                    :class="'task-workflow-node-status-inline__btn--' + getNodeStatus('api-test-fix')"
                    :disabled="nodeStatusSaving"
                    @click="cycleNodeStatus('api-test-fix')"
                  >{{ getNodeStatusLabel('api-test-fix') }}</button>
                </div>
              </div>
            </div>
            <MdEditor
              v-model="workflow.prompt_api_test"
              class="task-workflow-prompt-editor"
              preview-theme="github"
              :preview="true"
              :toolbars="promptEditorToolbars"
              height="100%"
            />
          </div>
        </div>
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
          <MdEditor
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
</template>

<script>
import { HomeFilled } from '@element-plus/icons-vue'
import GitActionButton from '@/components/base/GitActionButton.vue'
import ChatHistoryButton from '@/components/shared/ChatHistoryButton.vue'
import ChatHistoryDialog from '@/components/shared/ChatHistoryDialog.vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'
import taskWorkflowApi from '@/utils/base/task_workflow'
import homeTaskApi from '@/utils/base/home_task'
import baseUtils from '@/utils/base'
import sseDistribute from '@/utils/base/sse_distribute'
import chatParser from '@/utils/chat_parser'
import TaskProgressPanel from '@/components/TaskProgressPanel.vue'
import taskProgressStore from '@/utils/task_progress_store'
import gitApi from '@/utils/base/git'
import mysqlSetApi from '@/utils/base/mysql_set'
import apiManagement from '@/utils/base/api'
import dockerApi from '@/utils/base/compose'
import smartLinkSetApi from '@/utils/base/smart_link_set'
import agentCliApi from '@/utils/base/agent_cli'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'
import MarkdownIt from 'markdown-it'

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

const WORKFLOW_NODES = [
  { key: 'task-config', label: '任务配置', desc: '查看当前任务的所有配置信息' },
  { key: 'requirement-fetch', label: '1.抓取TAPD需求', desc: '自动登录和解析tapd需求到知识片段，转为markdown格式供AI解析' },
  { key: 'requirement', label: '2.需求分析', desc: '编写提示词，AI自动结合数据库和代码分析需求，形成开发文档' },
  { key: 'design', label: '3.开发执行', desc: '编写提示词，AI自动结合数据库，代码和开发文档进行开发' },
  { key: 'api-dev', label: '4.接口生成', desc: '编写提示词，AI自动获取登录态，将所有改动接口写入接口开发中' },
  { key: 'api-test-fix', label: '5.自动化测试+修复', desc: 'AI自动根据接口开发中的接口设计测试流程，自动上传代码+自动重启服务+自动修复BUG' },
  { key: 'code-review', label: '6.代码检查', desc: '让AI进行code review' },
  { key: 'browser-test', label: '7.需求核对浏览器测试', desc: '编写提示词，AI核对浏览器测试结果是否满足需求' },
]

// markdown-it 实例，用于在"执行历史"对话框中渲染 markdown（包括表格）
const md = new MarkdownIt({ html: true, breaks: true, linkify: true })

export default {
  name: 'TaskWorkflow',
  components: {
    HomeFilled,
    GitActionButton,
    ChatHistoryButton,
    ChatHistoryDialog,
    MdEditor,
    TaskProgressPanel,
  },
  data() {
    return {
      workflowNodes: WORKFLOW_NODES,
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
      requirementActiveTab: 'requirement-prompt',
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
      chatDetailSSERegistered: false,
      chatDetailSSELines: [], // SSE 累积的原始行
      chatDetailAutoScroll: true,
      _autoScrollLocked: false, // 程序化滚动锁
      _sseLineBuffer: [], // SSE 行缓冲（批处理），每100ms刷新一次
      _sseBatchTimer: null, // 批处理定时器
      _sseParseState: null, // 增量解析状态 { currentMessage, toolUseMap, pendingPatches }
      _initialSseRetryCount: 0, // 初始 start=1 SSE 失败后的重试计数
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
      chatDetailModelName: '',
      chatDetailAgentName: '',
      chatDetailLocalDir: '',
      chatDetailThinkingIntensity: '',
      chatDetailCliType: 'claude',
      _backgroundChatEventSources: {},
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
    }
  },
  computed: {
    taskId() {
      return Number(this.$route.params.taskId || 0)
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
    plainTextReqFragmentId() {
      return String(this.workflow.plain_text_requirement_fragment_id || '').trim()
    },
    designPlanReqFragmentId() {
      return String(this.workflow.design_plan_requirement_fragment_id || '').trim()
    },
    requirementFragmentTitle() {
      return String(this.requirementFragment.title || '').trim() || (this.requirementFragmentId ? `#${this.requirementFragmentId}` : '-')
    },
    devPlanFragmentId() {
      return String(this.workflow.dev_plan_fragment_id || '').trim()
    },
    workflowFragments() {
      return [
        { label: 'TAPD需求文档', id: this.requirementFragmentId },
        { label: '纯文本需求文档', id: this.plainTextReqFragmentId },
        { label: '需求设计方案文档', id: this.designPlanReqFragmentId },
      ]
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
      return 'task-config'
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
  },
  mounted() {
    this.loadWorkflowPage()
    this.loadTaskConfigLookupData()
    window.addEventListener('keydown', this.handleCtrlS)
  },
  beforeUnmount() {
    window.removeEventListener('keydown', this.handleCtrlS)
    this._stopChatHistoryDurationTimer()
    if (this._sseBatchTimer) { clearTimeout(this._sseBatchTimer); this._sseBatchTimer = null }
    this.unregisterWorkflowSse()
    this.stopAllBackgroundChatStreams()
    if (this._chatEventSource) {
      this._chatEventSource.close()
      this._chatEventSource = null
    }
  },
  watch: {
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
      this.requirementActiveTab = 'requirement-prompt'
      this.unregisterWorkflowSse()
      this.loadWorkflowPage()
    },
  },
  methods: {
    handleCtrlS(e) {
      if (!(e.ctrlKey && e.key === 's')) return
      e.preventDefault()
      const nodeToPrompt = { requirement: 'requirement', design: 'design', 'api-dev': 'api_dev', 'api-test-fix': 'api_test', 'code-review': 'code_review', 'browser-test': 'browser_test' }
      let promptType = nodeToPrompt[this.activeNode]
      if (this.activeNode === 'requirement-fetch' && this.requirementFetchActiveTab === 'plain-text-prompt') {
        promptType = 'plain_text_requirement'
      }
      if (this.activeNode === 'requirement' && this.requirementActiveTab === 'design-plan-prompt') {
        promptType = 'design_plan_requirement'
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
        this.activeNode = this.restoreActiveNodeCache() || this.firstRunningNodeKey
        this.loadRequirementFragment(() => {
          this.loading = false
          this.ensureWorkflowSse()
          this.maybeAutoFetchRequirement()
        })
        this.loadChatCounts()
      })
    },
    applyWorkflowPayload(data) {
      this.workflow = data.workflow || {}
      this.homeTask = data.home_task || this.homeTask || {}
      this.workflowId = Number(this.workflow.id || 0)
      this.requirementFetchConfig = data.requirement_fetch_config || this.requirementFetchConfig || {}
      this.parseNodeStatuses()
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
        const apiHost = String(baseUtils.GetApiHost() || window.location.origin).trim()
        this.requirementShareUrl = new URL(`/share/${encodeURIComponent(token)}`, apiHost).toString()
        this.replaceRequirementShareUrlPlaceholder()
      })
    },
    replaceRequirementShareUrlPlaceholder() {
      if (!this.requirementShareUrl) {
        return
      }
      const placeholder = '{需求文档地址}'
      if (this.workflow.prompt_requirement && this.workflow.prompt_requirement.includes(placeholder)) {
        this.workflow.prompt_requirement = this.workflow.prompt_requirement.replaceAll(placeholder, this.requirementShareUrl)
      }
      if (this.workflow.prompt_plain_text_requirement && this.workflow.prompt_plain_text_requirement.includes(placeholder)) {
        this.workflow.prompt_plain_text_requirement = this.workflow.prompt_plain_text_requirement.replaceAll(placeholder, this.requirementShareUrl)
      }
      if (this.workflow.prompt_design_plan_requirement && this.workflow.prompt_design_plan_requirement.includes(placeholder)) {
        this.workflow.prompt_design_plan_requirement = this.workflow.prompt_design_plan_requirement.replaceAll(placeholder, this.requirementShareUrl)
      }
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
      if (data.type === 'chat_status_change') {
        this.loadChatCounts()
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
      this.requirementFetchLogs.push({
        workflow_id: Number(data.workflow_id || 0),
        step: String(data.step || '').trim(),
        status: String(data.status || '').trim(),
        message: String(data.message || '').trim(),
        time: Number(data.time || 0),
      })
    },
    maybeAutoFetchRequirement() {
      if (this.requirementFetchAutoTriggered) {
        return
      }
      if (!String(this.homeTask.tapd_url || '').trim()) {
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
    triggerRequirementFetch(isAuto) {
      if (this.workflowId <= 0 || this.requirementFetchRunning) {
        return
      }
      if (!String(this.homeTask.tapd_url || '').trim()) {
        this.$helperNotify.error('当前任务未配置 TAPD 地址')
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
          this.errorMessage = response?.ErrMsg || '抓取 TAPD 需求失败'
          this.$helperNotify.error(this.errorMessage)
          this.loadWorkflowPage()
          return
        }
        this.$helperNotify.success('TAPD 需求已抓取并写入知识片段')
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
    openPlainTextReqFragment() {
      if (!this.plainTextReqFragmentId) {
        this.$helperNotify.error('当前工作流未绑定纯文本需求知识片段')
        return
      }
      const routeInfo = this.$router.resolve({
        path: '/MemoryFragment',
        query: {
          fragment_id: this.plainTextReqFragmentId,
          hide_menu: '1',
        },
      })
      window.open(routeInfo.href, '_blank')
    },
    openDesignPlanReqFragment() {
      if (!this.designPlanReqFragmentId) {
        this.$helperNotify.error('当前工作流未绑定需求设计方案知识片段')
        return
      }
      const routeInfo = this.$router.resolve({
        path: '/MemoryFragment',
        query: {
          fragment_id: this.designPlanReqFragmentId,
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
    openApiDocFragment() {
      const fragmentId = this.workflow.api_doc_fragment_id
      if (!fragmentId) {
        this.$message.warning('接口文档片段未初始化')
        return
      }
      this.openFragmentById(fragmentId)
    },
    resetApiDoc() {
      if (this.apiDocResetting || this.workflowId <= 0) return
      this.apiDocResetting = true
      const _this = this
      taskWorkflowApi.TaskWorkflowApiDocReset(this.workflowId, function (res) {
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
      taskWorkflowApi.TaskWorkflowPromptsSave({
        workflow_id: this.workflowId,
        prompt_requirement: this.workflow.prompt_requirement || '',
        prompt_api_dev: this.workflow.prompt_api_dev || '',
        prompt_api_test: this.workflow.prompt_api_test || '',
        prompt_design: this.workflow.prompt_design || '',
        prompt_plain_text_requirement: this.workflow.prompt_plain_text_requirement || '',
        prompt_design_plan_requirement: this.workflow.prompt_design_plan_requirement || '',
        prompt_browser_test: this.workflow.prompt_browser_test || '',
        prompt_code_review: this.workflow.prompt_code_review || '',
      }, (response) => {
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
        if (response.Data?.workflow) {
          this.workflow = response.Data.workflow
          this.$nextTick(() => {
            this.replaceRequirementShareUrlPlaceholder()
          })
        }
      })
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
          const list = res.Data.list || []
          this.updateChatCountsFromList(list)
          this.syncBackgroundChatStreams(list, this.promptChatDetailId || this.chatDetailId)
        }
      })
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
          const c = byType[pt] || { running: 0, interrupted: 0, total: 0 }
          c.total++
          if (item.status === 'running') c.running++
          else if (item.status === 'interrupted') c.interrupted++
          byType[pt] = c
        }
      }
      this.promptChatCounts = byType
    },
    // 加载对话详情
    loadChatDetail() {
      if (!this.chatDetailId) return
      taskWorkflowApi.TaskWorkflowChatDetail(this.chatDetailId, (res) => {
        if (res.ErrCode === 0 && res.Data) {
          const data = res.Data
          this.chatDetailPrompt = data.prompt || ''
          this.chatDetailSessionId = data.session_id || ''
          this.chatDetailStatus = data.status || ''
          this.chatDetailModelName = data.model_name || ''
          this.chatDetailAgentName = data.agent_cli_name || ''
          this.chatDetailLocalDir = data.local_dir || ''
          this.chatDetailThinkingIntensity = data.thinking_intensity || ''
          this.chatDetailLastUsageSummary = data.last_usage_summary || null
          this.chatDetailCliType = data.cli_type || 'claude'
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
          // 正在执行的对话未连接 SSE 时自动重连，保证刷新后仍能实时更新
          if (this.chatDetailStatus === 'running' && this._sseChatId !== this.chatDetailId) {
            this.connectChatStream(this.chatDetailId)
          }
        }
      })
    },
    // connectChatStream 创建专用 EventSource 连接以实时接收对话输出。
    // isNewChat: true 表示新对话首次启动（需后端启动 claude 进程），false 表示重连已有对话。
    connectChatStream(chatId, continuePrompt, isNewChat) {
      if (this._sseChatId === chatId && this._chatEventSource && this._chatEventSource.readyState !== EventSource.CLOSED) return
      // 关闭上一个 chat 的 SSE 连接
      if (this._chatEventSource) {
        this._chatEventSource.close()
        this._chatEventSource = null
      }
      this._sseChatId = chatId
      this.chatDetailSSERegistered = true
      this._thinkingStreamStartTime = 0 // 当前对话思考计时的起始时间戳
      // 初始化增量解析状态
      this._sseParseState = this.chatDetailCliType === 'codex'
        ? { currentItems: new Map(), pendingPatches: [] }
        : { currentMessage: null, toolUseMap: new Map(), pendingPatches: [] }
      this._sseLineBuffer = []
      if (this._sseBatchTimer) { clearTimeout(this._sseBatchTimer); this._sseBatchTimer = null }
      // 启动思考耗时动态更新定时器
      if (this._thinkingTimer) { clearInterval(this._thinkingTimer); this._thinkingTimer = null }
      this.thinkingStreamElapsed = 0
      this._thinkingTimer = setInterval(() => {
        if (this._thinkingStreamStartTime > 0) {
          this.thinkingStreamElapsed = Math.floor((Date.now() - this._thinkingStreamStartTime) / 1000)
        } else {
          this.thinkingStreamElapsed = 0
        }
      }, 200)
      const sseHost = baseUtils.GetSseApiHost()
      let url = sseHost + '/api/task/workflow/chat/stream?chat_id=' + chatId + '&token=' + encodeURIComponent(baseUtils.GetSafeToken())
      if (isNewChat) {
        url += '&start=1'
      }
      if (continuePrompt) {
        url += '&continue=1&prompt=' + encodeURIComponent(continuePrompt)
      }
      const es = new EventSource(url)
      this._chatEventSource = es
      es.onmessage = (event) => {
        const line = event.data
        if (!line) return
        try {
          const obj = JSON.parse(line)
          if (obj.type === 'chat' && obj.subtype === 'completed') {
            this._flushSseBatch()
            this.chatDetailSSELines.push(line)
            this._sseChatId = 0
            this.chatDetailSSERegistered = false
            es.close()
            this._chatEventSource = null
            this._sseParseState = null
            this.loadChatDetail()
            this.loadChatCounts()
            this.$nextTick(() => { this.scrollPromptChatToBottom() })
            return
          }
          // 追踪思考耗时：首次 thinking_delta 时记录起始时间
          if (obj.type === 'stream_event') {
            const evt = obj.event || {}
            if (evt.type === 'content_block_delta') {
              const delta = evt.delta || {}
              if (delta.type === 'thinking_delta' && this._thinkingStreamStartTime === 0) {
                this._thinkingStreamStartTime = Date.now()
              }
            } else if (evt.type === 'message_stop' && this._thinkingStreamStartTime > 0) {
              const durationMs = Date.now() - this._thinkingStreamStartTime
              this._thinkingStreamStartTime = 0
              // 将耗时写入消息——会在 parseChatLines 后应用到对应消息
              this._pendingThinkingDurationMs = durationMs
            }
          }
        } catch (e) { /* ignore parse errors */ }
        // 行缓冲：每 100ms 批量刷新，避免每条 SSE 事件都触发全量解析和 DOM 更新
        this._sseLineBuffer.push(line)
        if (!this._sseBatchTimer) {
          this._sseBatchTimer = setTimeout(() => {
            this._flushSseBatch()
          }, 100)
        }
      }
      es.onerror = () => {
        this._flushSseBatch()
        if (this._thinkingTimer) { clearInterval(this._thinkingTimer); this._thinkingTimer = null }
        this.thinkingStreamElapsed = 0
        this.chatDetailSSERegistered = false
        es.close()
        this._chatEventSource = null
        this._sseParseState = null
        // 如果是初始 start=1 连接失败且尚无任何输出，重试一次 start=1 连接
        // 避免 loadChatDetail 中不带 start/continue 的重连将对话错误标记为"中断"
        if (this._initialSseRetryCount < 1 && this.chatDetailSSELines.length === 0 && this.chatDetailStatus === 'running') {
          this._initialSseRetryCount++
          this.connectChatStream(this.chatDetailId, null, true)
          return
        }
        this.loadChatDetail()
        this.loadChatCounts()
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
        if (this._pendingThinkingDurationMs > 0) {
          for (let i = this.chatDetailMessages.length - 1; i >= 0; i--) {
            const msg = this.chatDetailMessages[i]
            if (msg.type === 'assistant' && msg.thinking) {
              msg._thinkingTiming = msg._thinkingTiming || { startMs: 0, durationMs: 0 }
              msg._thinkingTiming.durationMs = this._pendingThinkingDurationMs
              if (!msg._thinkingManuallyToggled) {
                msg._thinkingCollapsed = true
              }
              break
            }
          }
          this._pendingThinkingDurationMs = 0
        }
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
      if (this._thinkingStreamStartTime === 0) return false
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
      this._initialSseRetryCount = 0
      if (this._chatEventSource) {
        this._chatEventSource.close()
        this._chatEventSource = null
      }
      this._sseChatId = 0
      this.chatDetailSSERegistered = false
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
    // isChatContinueDisabled 统一发送区按钮可用状态，确保“发送/新对话”行为一致。 // Keeps the action buttons under the same enabled-state rule.
    isChatContinueDisabled() {
      return this.chatContinueLoading || !String(this.chatContinueInput || '').trim()
    },
    continueChat() {
      const input = this.chatContinueInput.trim()
      if (!input) return
      this.chatContinueLoading = true
      taskWorkflowApi.TaskWorkflowChatContinue(this.chatDetailId, input, (res) => {
        this.chatContinueLoading = false
        if (res.ErrCode === 0) {
          this.chatContinueInput = ''
          this.chatDetailStatus = 'running'
          this.connectChatStream(this.chatDetailId, input)
          setTimeout(() => { this.loadChatDetail() }, 500)
        } else {
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
            this.chatContinueInput = ''
            this.$helperNotify.success('已新建对话并发送到 ' + cliLabel + ' 执行')
            this.chatDetailId = chatId
            this.chatDetailStatus = 'running'
            this.chatDetailCliType = cliType
            this.chatDetailSSELines = []
            this.chatDetailMessages = []
            this.chatDetailLastUsageSummary = null
            taskProgressStore.reset()
            this._initialSseRetryCount = 0
            this.connectChatStream(chatId, null, true)
            this.loadChatDetail()
            this.loadChatCounts()
            this.openPromptChatHistory(promptType, chatId)
          }
        )
      })
    },
    // 停止对话
    stopChat() {
      // 关闭 SSE 连接
      if (this._chatEventSource) {
        this._chatEventSource.close()
        this._chatEventSource = null
      }
      this._sseChatId = 0
      this.chatDetailSSERegistered = false
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
          // 仅展示“已启用且配置可用”的 Agent，避免把停用实例带入执行弹窗。
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
              this.$helperNotify.success('已发送到 ' + cliLabel + ' 执行')
              this.promptExecDialogVisible = false
              // 初始化对话显示状态并连接 SSE 流以启动执行
              this.chatDetailId = chatId
              this.chatDetailStatus = 'running'
              this.chatDetailCliType = cliType
              this.chatDetailSSELines = []
              this.chatDetailMessages = []
              taskProgressStore.reset()
              this._initialSseRetryCount = 0
              this.connectChatStream(chatId, null, true)
              this.loadChatDetail()
              this.loadChatCounts()
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
      return this.promptChatCounts[promptType] || { running: 0, interrupted: 0, total: 0 }
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
      this.promptChatHistoryTitle = titleMap[promptType] || promptType
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
          this.syncBackgroundChatStreams(this.promptChatHistoryList, focusChatId || this.promptChatDetailId || this.chatDetailId)
          if (!promptType) {
            this.updateChatCountsFromList(this.promptChatHistoryList)
          }
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
      this.promptChatDetailId = row.id
      this.chatDetailId = row.id
      this.chatDetailStatus = row.status
      this.chatDetailAutoScroll = true
      this.promptChatDetailShowScrollBtn = false
      // 切到不同 chat 时才断开旧 SSE
      if (this._chatEventSource && this._sseChatId !== row.id) {
        this._chatEventSource.close()
        this._chatEventSource = null
        this._sseChatId = 0
      }
      // 中文注释：先关闭旧前台 SSE，再决定哪些对话转入后台监听，避免旧选中对话漏挂后台连接。
      // English comment: Close the previous foreground SSE before recalculating background streams so the old active chat can move into background tracking.
      this.syncBackgroundChatStreams(this.promptChatHistoryList, row.id)
      if (this._sseChatId !== row.id) {
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
    // 关闭执行历史弹窗（保留 SSE 连接和聊天状态）
    onPromptChatHistoryClosed() {
      this._stopChatHistoryDurationTimer()
      this.stopAllBackgroundChatStreams()
    },
    // 彻底关闭对话详情（仅在用户主动停止或切换时调用）
    closePromptChatDetail() {
      this.closeChatDetail()
      this.promptChatDetailId = 0
    },
    updateChatListStatus(chatId, status) {
      const updateItem = (list) => {
        const item = list.find(i => i.id === chatId)
        if (item) item.status = status
      }
      updateItem(this.promptChatHistoryList)
      this.syncBackgroundChatStreams(this.promptChatHistoryList, this.promptChatDetailId || this.chatDetailId)
    },
    syncBackgroundChatStreams(list, selectedChatId) {
      const normalizedSelectedChatId = Number(selectedChatId || 0)
      const runningIds = new Set(
        (Array.isArray(list) ? list : [])
          .filter(item => item && item.status === 'running')
          .map(item => Number(item.id || 0))
          .filter(id => id > 0 && id !== normalizedSelectedChatId)
      )
      Object.keys(this._backgroundChatEventSources || {}).forEach((key) => {
        const chatId = Number(key)
        if (!runningIds.has(chatId)) {
          this.stopBackgroundChatStream(chatId)
        }
      })
      runningIds.forEach((chatId) => {
        if (chatId !== Number(this._sseChatId || 0)) {
          this.startBackgroundChatStream(chatId)
        }
      })
    },
    startBackgroundChatStream(chatId) {
      const normalizedChatId = Number(chatId || 0)
      if (normalizedChatId <= 0) return
      if (normalizedChatId === Number(this._sseChatId || 0)) return
      if (this._backgroundChatEventSources[normalizedChatId]) return
      const currentItem = this.promptChatHistoryList.find(item => Number(item.id || 0) === normalizedChatId)
      const sseHost = baseUtils.GetSseApiHost()
      const url = sseHost + '/api/task/workflow/chat/stream?chat_id=' + normalizedChatId + '&token=' + encodeURIComponent(baseUtils.GetSafeToken())
      const es = new EventSource(url)
      const state = {
        es,
        // 中文注释：背景 SSE 维护独立行数基线，确保非选中执行中的历史项也能实时更新 xx 条。
        // English comment: Background SSE keeps its own line-count baseline so non-selected running chats can still update the visible message counter.
        lineCount: Number(currentItem?.line_count || 0),
      }
      this._backgroundChatEventSources = {
        ...this._backgroundChatEventSources,
        [normalizedChatId]: state,
      }
      es.onmessage = (event) => {
        const line = event.data
        if (!line) return
        // 中文注释：先按消息级别递增行数，再解析终态事件更新状态，避免必须等 DB flush 才看到条数变化。
        // English comment: Increment line count before terminal-event parsing so the UI reflects new lines immediately without waiting for DB flushes.
        state.lineCount += 1
        this.updateBackgroundChatListItem(normalizedChatId, {
          line_count: state.lineCount,
        })
        try {
          const obj = JSON.parse(line)
          if (obj.type === 'chat' && obj.subtype === 'completed') {
            this.updateBackgroundChatListItem(normalizedChatId, {
              status: String(obj.status || 'completed').trim() || 'completed',
              line_count: state.lineCount,
            })
            this.stopBackgroundChatStream(normalizedChatId)
            this.loadPromptChatHistoryListSilently()
            this.loadChatCounts()
          }
        } catch (e) {
          // 中文注释：普通流式消息解析失败时只保留计数更新；状态仍由终态事件或静默刷新兜底。
          // English comment: Parse failures on normal stream lines only skip terminal-state parsing; the count update already happened.
        }
      }
      es.onerror = () => {
        this.stopBackgroundChatStream(normalizedChatId)
      }
    },
    updateBackgroundChatListItem(chatId, patch) {
      const normalizedChatId = Number(chatId || 0)
      if (normalizedChatId <= 0 || !patch) return
      const item = this.promptChatHistoryList.find(row => Number(row.id || 0) === normalizedChatId)
      if (!item) return
      Object.keys(patch).forEach((key) => {
        if (patch[key] !== undefined) {
          item[key] = patch[key]
        }
      })
      if (normalizedChatId === Number(this.chatDetailId || 0) && patch.status) {
        this.chatDetailStatus = patch.status
      }
      this.promptChatHistoryList = this.promptChatHistoryList.slice()
    },
    stopBackgroundChatStream(chatId) {
      const normalizedChatId = Number(chatId || 0)
      if (normalizedChatId <= 0) return
      const entry = this._backgroundChatEventSources[normalizedChatId]
      if (entry?.es) {
        entry.es.close()
      }
      const nextMap = { ...this._backgroundChatEventSources }
      delete nextMap[normalizedChatId]
      this._backgroundChatEventSources = nextMap
    },
    stopAllBackgroundChatStreams() {
      Object.keys(this._backgroundChatEventSources || {}).forEach((key) => {
        this.stopBackgroundChatStream(Number(key))
      })
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
        if (!promptType) {
          this.updateChatCountsFromList(this.promptChatHistoryList)
        }
        this.syncBackgroundChatStreams(this.promptChatHistoryList, this.promptChatDetailId || this.chatDetailId)
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
        if (this._sseChatId > 0 && this.chatDetailSSELines.length > 0) {
          const count = this.chatDetailSSELines.length
          const updateLineCount = (list) => {
            const item = list.find(i => i.id === this._sseChatId)
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
      smartLinkSetApi.SmartLinkList((response) => {
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
        return item ? item.name : String(id)
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
      return ''
    },
    selectNode(key) {
      this.activeNode = key
      this.saveActiveNodeCache()
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
      return localStorage.getItem(this.getActiveNodeCacheKey())
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
      if (item.status === 'running' && this._sseChatId > 0 && item.id === this._sseChatId) {
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
    // 批量检查工作流页面中本地目录的当前 Git 分支是否与配置的分支名匹配
    checkWorkflowBranchStatus() {
      const configs = this.parsedTaskDevConfigs
      const items = []
      const seen = new Set()
      for (const cfg of configs) {
        const dir = String(cfg.local_dir || '').trim()
        const branch = String(cfg.branch_name || '').trim()
        if (!dir || !branch) continue
        const key = dir + '|' + branch
        if (seen.has(key)) continue
        seen.add(key)
        items.push({ local_dir: dir, branch_name: branch })
      }
      if (items.length === 0) return
      homeTaskApi.LocalBranchBatchCheck(items, (response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.branchStatusMap = { ...this.branchStatusMap, ...response.Data }
        }
      })
    },
    // 检查所有 dev_config 的分支是否匹配，返回不匹配的列表
    getMismatchedBranches() {
      const mismatched = []
      for (const cfg of this.parsedTaskDevConfigs) {
        const dir = String(cfg.local_dir || '').trim()
        const branch = String(cfg.branch_name || '').trim()
        if (!dir || !branch) continue
        const key = dir + '|' + branch
        const status = this.branchStatusMap[key]
        if (status && !status.matched) {
          mismatched.push({
            local_dir: dir,
            expected_branch: branch,
            current_branch: status.current_branch || '未知',
          })
        }
      }
      return mismatched
    },
    // 执行前检查分支是否匹配，不匹配时弹出确认框，返回 Promise<boolean>
    confirmBranchBeforeExec() {
      const mismatched = this.getMismatchedBranches()
      if (mismatched.length === 0) {
        return Promise.resolve(true)
      }
      const lines = mismatched.map(m => `${m.local_dir}\n  期望分支: ${m.expected_branch}\n  当前分支: ${m.current_branch}`)
      const msg = '以下目录的分支与配置不一致：\n\n' + lines.join('\n\n') + '\n\n是否继续执行？'
      return this.$confirm(msg, '分支不匹配警告', {
        confirmButtonText: '继续执行',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => true).catch(() => false)
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
}

.task-workflow-header__title {
  margin: 0;
  font-size: 22px;
  line-height: 1.3;
  color: #303133;
  flex-shrink: 0;
}

.task-workflow-header__meta {
  display: flex;
  flex-direction: column;
  flex-wrap: wrap;
  gap: 4px;
  margin-top: 6px;
  color: #909399;
  font-size: 13px;
}

.task-workflow-header__dev-row {
  display: flex;
  flex-wrap: wrap;
  gap: 6px;
  align-items: center;
}

.task-workflow-header__dev-item {
  white-space: nowrap;
}

.task-workflow-header__dev-sep {
  color: #dcdfe6;
  font-size: 12px;
}

.task-workflow-header__branch {
  color: #3a7a3a;
  cursor: pointer;
  transition: color 0.2s;
}

.task-workflow-header__branch:hover {
  color: #2d5f2d;
  text-decoration: underline;
}

.task-workflow-header__dev-item--link {
  color: #3a7a3a;
  cursor: pointer;
  transition: color 0.2s;
}

.task-workflow-header__dev-item--link:hover {
  color: #2d5f2d;
  text-decoration: underline;
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
  display: grid;
  grid-template-columns: repeat(8, minmax(0, 1fr));
  gap: 10px;
  flex-shrink: 0;
}

.task-workflow-node {
  border: 1px solid #e8e8e0;
  border-radius: 8px;
  background: #fff;
  min-height: 50px;
  padding: 14px 16px;
  text-align: left;
  cursor: pointer;
  transition: all 0.2s ease;
  display: flex;
  flex-direction: column;
  gap: 6px;
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
  font-size: 15px;
  line-height: 1.4;
  color: #303133;
  font-weight: 600;
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
}

.task-workflow-tab {
  display: flex;
  flex-direction: column;
  gap: 12px;
  flex: 1;
  height: 100%;
  min-height: 0;
}

.task-workflow-grid {
  display: grid;
  grid-template-columns: repeat(4, minmax(0, 1fr));
  gap: 12px;
}

.task-workflow-card {
  border-radius: 12px;
  padding: 16px;
  background: #fafaf7;
  border: 1px solid #e8e8e0;
  flex: 1;
  height: 100%;
  min-height: 0;
  display: flex;
  flex-direction: column;
  overflow: hidden;
}

.task-workflow-prompt-editor {
  flex: 1;
  height: 100%;
  min-height: 0;
}

.task-workflow-card :deep(.md-editor) {
  flex: 1;
  height: 100%;
  min-height: 0;
}

.task-workflow-card :deep(.md-editor-content) {
  min-height: 0;
}

.task-workflow-card :deep(.md-editor-input-wrapper),
.task-workflow-card :deep(.md-editor-preview-wrapper) {
  overflow: auto;
}

/* MdEditor 滚动条绿色 */
.task-workflow-card :deep(.md-editor) {
  --md-scrollbar-bg-color: #edf3e8;
  --md-scrollbar-thumb-color: #9fb39a;
  --md-scrollbar-thumb-hover-color: #869c82;
  --md-scrollbar-thumb-active-color: #7a8f76;
}

.task-workflow-card :deep(.md-editor .md-editor-preview ::-webkit-scrollbar) {
  width: 10px !important;
  height: 10px !important;
}

.task-workflow-card :deep(.md-editor .md-editor-preview ::-webkit-scrollbar-track) {
  background: #edf3e8 !important;
  border-radius: 999px !important;
}

.task-workflow-card :deep(.md-editor .md-editor-preview ::-webkit-scrollbar-thumb) {
  background: #9fb39a !important;
  border: 2px solid #edf3e8 !important;
  border-radius: 999px !important;
}

.task-workflow-card :deep(.md-editor .md-editor-preview ::-webkit-scrollbar-thumb:hover) {
  background: #869c82 !important;
}

.task-workflow-card :deep(.md-editor .md-editor-preview ::-webkit-scrollbar-corner) {
  background: #edf3e8 !important;
}

/* fragment-view 原生滚动条绿色 */
.task-workflow-fragment-view {
  scrollbar-width: thin;
  scrollbar-color: #9fb39a #edf3e8;
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

@media (max-width: 1100px) {
  .task-workflow-nodes {
    grid-template-columns: repeat(2, minmax(0, 1fr));
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

  .task-workflow-card__header {
    flex-direction: column;
    align-items: flex-start;
  }
}

/* 节点状态图标 */
.task-workflow-node {
  position: relative;
}

.task-workflow-node__status-icon {
  position: absolute;
  top: 8px;
  right: 8px;
  display: flex;
  align-items: center;
  justify-content: center;
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
  border-left: 3px solid #909399;
}
.task-workflow-node--status-completed {
  border-left: 3px solid #67c23a;
}

.task-workflow-node--status-skipped {
  border-left: 3px solid #e6a23c;
}

.task-workflow-node--status-running {
  border-left: 3px solid #409eff;
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
</style>
