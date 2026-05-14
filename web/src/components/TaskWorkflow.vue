<template>
  <div class="task-workflow-page" v-loading="loading">
    <div class="task-workflow-shell">
      <header class="task-workflow-header">
        <div class="task-workflow-header__main">
          <div class="task-workflow-header__eyebrow">任务工作流程</div>
          <h1 class="task-workflow-header__title">{{ homeTask.name || `任务 #${taskId}` }}</h1>
        </div>
        <div class="task-workflow-header__actions">
          <el-tooltip content="返回首页" placement="bottom">
            <el-button class="task-workflow-home-btn" @click="goHome">
              <el-icon :size="18"><HomeFilled /></el-icon>
            </el-button>
          </el-tooltip>
          <GitActionButton compact variant="info" @click="goBackToTaskList">
            返回任务清单
          </GitActionButton>
          <GitActionButton compact :loading="loading" @click="reloadWorkflowPage">
            刷新
          </GitActionButton>
          <GitActionButton compact variant="warning" @click="openIssueFixDialog">
            问题修改提示词
          </GitActionButton>
          <GitActionButton compact @click="openChatHistoryDialog">
            历史对话
          </GitActionButton>
          <GitActionButton compact variant="success" @click="openZcodeConfigDialog">
            zcode配置
          </GitActionButton>
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
            <span v-else class="status-icon status-icon--running"></span>
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
                    <el-descriptions-item label="接口集合">{{ getTaskConfigApiLabel(cfg) }}</el-descriptions-item>
                    <el-descriptions-item label="本地目录">{{ cfg.local_dir || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="父分支">{{ cfg.parent_branch || '-' }}</el-descriptions-item>
                    <el-descriptions-item label="分支名">{{ cfg.branch_name || '-' }}</el-descriptions-item>
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
      class="task-workflow-issue-fix-dialog"
    >
      <div class="task-workflow-issue-fix__close-bar">
        <el-button @click="issueFixDialogVisible = false" type="danger">关闭</el-button>
      </div>
        <div style="margin-bottom: 12px; display: flex; gap: 8px;">
          <el-button type="primary" :loading="sendingToClaude" @click="sendToClaudeCode">
            发送到 claude code 执行
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
          <div class="task-workflow-issue-fix__label">完整提示词</div>
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

    <!-- 对话配置弹窗 -->
    <el-dialog
      v-model="chatConfigDialogVisible"
      title="Claude Code 对话配置"
      width="500px"
      destroy-on-close
    >
      <el-form label-width="100px">
        <el-form-item label="工作目录">
          <el-select v-model="chatConfigLocalDir" style="width: 100%;" placeholder="请选择工作目录">
            <el-option
              v-for="(dir, idx) in chatConfigDirs"
              :key="idx"
              :label="dir"
              :value="dir"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="模型">
          <el-select v-model="chatConfigModelId" style="width: 100%;" placeholder="请选择模型">
            <el-option
              v-for="m in chatConfigModels"
              :key="m.id"
              :label="m.name + ' (' + m.model + ')'"
              :value="m.id"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="chatConfigDialogVisible = false">取消</el-button>
        <el-button type="primary" :disabled="!chatConfigModelId || !chatConfigLocalDir" @click="startClaudeChat">
          开始对话
        </el-button>
      </template>
    </el-dialog>
  </div>

    <!-- 历史对话列表弹窗 -->
    <el-dialog
      v-model="chatHistoryDialogVisible"
      title="历史对话"
      width="700px"
      :close-on-click-modal="false"
      destroy-on-close
    >
      <el-table :data="chatHistoryList" style="width: 100%" v-loading="chatHistoryLoading" @row-click="openChatDetail" row-style="cursor: pointer;">
        <el-table-column prop="id" label="ID" width="60" />
        <el-table-column prop="created_at" label="时间" width="170" />
        <el-table-column prop="prompt" label="提示词" show-overflow-tooltip>
          <template #default="{ row }">
            {{ (row.prompt || '').substring(0, 80) }}{{ (row.prompt || '').length > 80 ? '...' : '' }}
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="90">
          <template #default="{ row }">
            <el-tag :type="row.status === 'running' ? 'warning' : 'success'" size="small">
              {{ row.status === 'running' ? '执行中' : '已完成' }}
            </el-tag>
          </template>
        </el-table-column>
      </el-table>
      <template #footer>
        <el-button @click="chatHistoryDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 对话详情弹窗 -->
    <el-dialog
      v-model="chatDetailDialogVisible"
      :title="'对话 #' + chatDetailId"
      width="1200px"
      :close-on-click-modal="false"
      destroy-on-close
      @closed="closeChatDetail"
    >
      <div v-if="chatDetailModelName || chatDetailLocalDir" style="margin-bottom: 12px; color: #888; font-size: 12px;">
        <span v-if="chatDetailModelName">模型: {{ chatDetailModelName }}</span>
        <span v-if="chatDetailModelName && chatDetailLocalDir"> | </span>
        <span v-if="chatDetailLocalDir">目录: {{ chatDetailLocalDir }}</span>
      </div>
      <div class="chat-detail-container" style="max-height: 70vh; overflow-y: auto; background: #1e1e1e; border-radius: 8px; padding: 16px; color: #d4d4d4; font-family: 'Consolas', monospace; font-size: 13px;">
        <div v-if="chatDetailMessages.length === 0 && chatDetailStatus === 'running'" style="text-align: center; padding: 40px; color: #888;">
          <div>等待 claude code 响应...</div>
        </div>
        <div v-for="(msg, idx) in chatDetailMessages" :key="idx" style="margin-bottom: 8px;">
          <!-- system_init -->
          <div v-if="msg.type === 'system_init'" style="color: #6a9955; font-size: 12px; padding: 4px 0;">
            ✔ {{ msg.text }} | model: {{ msg.model }}
          </div>
          <!-- system_command -->
          <div v-else-if="msg.type === 'system_command'" style="background: #1e1e3f; border-radius: 4px; padding: 8px 12px; margin: 4px 0; color: #ce9178; font-size: 12px; word-break: break-all;">
            <span style="color: #569cd6;">$</span> {{ msg.text }}
          </div>
          <!-- system_hook -->
          <div v-else-if="msg.type === 'system_hook'" style="color: #888; font-size: 12px;">
            <span @click="msg.collapsed = !msg.collapsed" style="cursor: pointer;">{{ msg.collapsed ? '▶' : '▼' }} {{ msg.text }}</span>
          </div>
          <!-- system (generic) -->
          <div v-else-if="msg.type === 'system'" style="color: #888; font-size: 11px;">{{ msg.text }}</div>
          <!-- assistant message -->
          <div v-else-if="msg.type === 'assistant'" style="background: #2d2d2d; border-radius: 8px; padding: 12px; margin: 8px 0;">
            <!-- thinking -->
            <div v-if="msg.thinking" style="color: #888; font-size: 12px; margin-bottom: 8px;">
              <span @click="msg._thinkingCollapsed = !msg._thinkingCollapsed" style="cursor: pointer; font-weight: bold;">
                {{ msg._thinkingCollapsed ? '▶ 思考过程' : '▼ 思考过程' }}
              </span>
              <pre v-if="!msg._thinkingCollapsed" style="white-space: pre-wrap; margin-top: 4px; color: #999;">{{ msg.thinking }}</pre>
            </div>
            <!-- content blocks -->
            <div v-for="(block, bi) in msg.content" :key="bi">
              <div v-if="block.type === 'text'" style="white-space: pre-wrap; line-height: 1.5;">{{ block.text }}</div>
              <div v-else-if="block.type === 'tool_use'" style="background: #1a3a1a; border-radius: 4px; padding: 8px; margin: 4px 0;">
                <span style="color: #4ec9b0;">🔧 {{ block.name }}</span>
                <pre style="white-space: pre-wrap; font-size: 12px; color: #ce9178; margin-top: 4px;">{{ block.input }}</pre>
              </div>
            </div>
            <!-- usage -->
            <div v-if="msg.usage" style="color: #888; font-size: 11px; margin-top: 8px; border-top: 1px solid #444; padding-top: 4px;">
              input: {{ msg.usage.input_tokens }} | output: {{ msg.usage.output_tokens }}
            </div>
          </div>
          <!-- tool_result -->
          <div v-else-if="msg.type === 'tool_result'" style="color: #888; font-size: 12px;">
            <span @click="msg.collapsed = !msg.collapsed" style="cursor: pointer;">{{ msg.collapsed ? '▶' : '▼' }} 工具执行结果</span>
            <pre v-if="!msg.collapsed" style="white-space: pre-wrap; font-size: 11px; margin-top: 4px; max-height: 200px; overflow-y: auto;">{{ msg.text }}</pre>
          </div>
          <!-- assistant_text -->
          <div v-else-if="msg.type === 'assistant_text'" style="white-space: pre-wrap; line-height: 1.5;">{{ msg.text }}</div>
          <!-- assistant_thinking -->
          <div v-else-if="msg.type === 'assistant_thinking'" style="color: #888; font-size: 12px;">
            <span @click="msg.collapsed = !msg.collapsed" style="cursor: pointer;">{{ msg.collapsed ? '▶ 思考' : '▼ 思考' }}</span>
            <pre v-if="!msg.collapsed" style="white-space: pre-wrap; margin-top: 4px;">{{ msg.text }}</pre>
          </div>
          <!-- result -->
          <div v-else-if="msg.type === 'result'" style="color: #6a9955; font-size: 12px; border-top: 1px solid #444; padding-top: 8px; margin-top: 8px;">
            {{ msg.isError ? '✘ 错误' : '✔ 完成' }} | 耗时: {{ (msg.durationMs / 1000).toFixed(1) }}s | {{ msg.numTurns }} 轮
            <span v-if="msg.usage"> | input: {{ msg.usage.input_tokens }} output: {{ msg.usage.output_tokens }}</span>
          </div>
          <!-- chat_completed -->
          <div v-else-if="msg.type === 'chat_completed'" style="color: #6a9955; text-align: center; padding: 16px;">
            ✔ {{ msg.text }}
          </div>
        </div>
      </div>
      <template #footer>
        <div style="display: flex; gap: 8px; align-items: center;">
          <el-input
            v-if="chatDetailStatus !== 'running'"
            v-model="chatContinueInput"
            placeholder="输入新消息继续对话..."
            @keyup.enter="continueChat"
            style="flex: 1;"
          />
          <el-button v-if="chatDetailStatus !== 'running'" type="primary" :loading="chatContinueLoading" @click="continueChat">发送</el-button>
          <el-button @click="chatDetailDialogVisible = false">关闭</el-button>
        </div>
      </template>
    </el-dialog>

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
</template>

<script>
import { HomeFilled } from '@element-plus/icons-vue'
import GitActionButton from '@/components/base/GitActionButton.vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'
import taskWorkflowApi from '@/utils/base/task_workflow'
import homeTaskApi from '@/utils/base/home_task'
import baseUtils from '@/utils/base'
import sseDistribute from '@/utils/base/sse_distribute'
import chatParser from '@/utils/chat_parser'
import aiSet from '@/utils/base/ai_set'
import gitApi from '@/utils/base/git'
import mysqlSetApi from '@/utils/base/mysql_set'
import apiManagement from '@/utils/base/api'
import dockerApi from '@/utils/base/compose'
import smartLinkSetApi from '@/utils/base/smart_link_set'
import { MdEditor } from 'md-editor-v3'
import 'md-editor-v3/lib/style.css'

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

const TASK_STATUS_TODO = '待开始'
const TASK_STATUS_DEVELOPING = '开发中'
const TASK_STATUS_SELF_TESTING = '自测中'
const TASK_STATUS_SELF_TESTED = '自测完'
const TASK_STATUS_PENDING_INTEGRATION = '待对接'
const TASK_STATUS_INTEGRATING = '对接中'
const TASK_STATUS_TESTING = '测试中'
const TASK_STATUS_RELEASING = '上线中'
const TASK_STATUS_ONLINE = '已上线'
const TASK_STATUS_OPTIONS = [
  TASK_STATUS_TODO,
  TASK_STATUS_DEVELOPING,
  TASK_STATUS_SELF_TESTING,
  TASK_STATUS_SELF_TESTED,
  TASK_STATUS_PENDING_INTEGRATION,
  TASK_STATUS_INTEGRATING,
  TASK_STATUS_TESTING,
  TASK_STATUS_RELEASING,
  TASK_STATUS_ONLINE,
]

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

export default {
  name: 'TaskWorkflow',
  components: {
    HomeFilled,
    GitActionButton,
    MdEditor,
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
      // claude code 对话
      chatHistoryDialogVisible: false,
      chatHistoryList: [],
      chatHistoryLoading: false,
      chatDetailDialogVisible: false,
      chatDetailId: 0,
      chatDetailPrompt: '',
      chatDetailSessionId: '',
      chatDetailStatus: '',
      chatDetailMessages: [],
      chatDetailSSERegistered: false,
      sendingToClaude: false,
      chatContinueInput: '',
      chatContinueLoading: false,
      chatConfigDialogVisible: false,
      chatConfigModelId: 0,
      chatConfigLocalDir: '',
      chatConfigDirs: [],
      chatConfigModels: [],
      _pendingChatPrompt: '',
      chatDetailModelName: '',
      chatDetailLocalDir: '',
      // zcode 配置
      zcodeConfigDialogVisible: false,
      zcodeDirInput: '',
      zcodeProjectList: [],
      zcodeSaving: false,
      issueFixZcodeMappings: [],
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
      const template = (this.issueFixResolvedTemplate || '').trim()
      if (!input && !template) return ''
      if (!input) return template
      if (!template) return input
      return input + '\n\n' + template
    },
  },
  mounted() {
    this.loadWorkflowPage()
    this.loadTaskConfigLookupData()
    window.addEventListener('keydown', this.handleCtrlS)
  },
  beforeUnmount() {
    window.removeEventListener('keydown', this.handleCtrlS)
    this.unregisterWorkflowSse()
  },
  watch: {
    parsedTaskDevConfigs: {
      handler(configs) {
        for (const cfg of configs) {
          const colId = Number(cfg.collection_id || 0)
          if (colId > 0) {
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
        this.activeNode = this.restoreActiveNodeCache() || this.firstRunningNodeKey
        this.loadRequirementFragment(() => {
          this.loading = false
          this.ensureWorkflowSse()
          this.maybeAutoFetchRequirement()
        })
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
    // 打开历史对话列表
    openChatHistoryDialog() {
      this.chatHistoryDialogVisible = true
      this.chatHistoryLoading = true
      taskWorkflowApi.TaskWorkflowChatList(this.workflowId, (res) => {
        this.chatHistoryLoading = false
        if (res.ErrCode === 0 && res.Data) {
          this.chatHistoryList = res.Data.list || []
        }
      })
    },
    // 打开对话详情
    openChatDetail(row) {
      this.chatDetailId = row.id
      this.chatDetailDialogVisible = true
      this.chatDetailStatus = row.status
      this.loadChatDetail()
      if (row.status === 'running') {
        this.registerChatSSE(row.id)
      }
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
          this.chatDetailModelName = data.model_id ? '#' + data.model_id : ''
          this.chatDetailLocalDir = data.local_dir || ''
          this.chatDetailMessages = chatParser.parseChatLines(data.lines || [])
        }
      })
    },
    // 注册 SSE 接收
    registerChatSSE(chatId) {
      if (this.chatDetailSSERegistered) return
      this.chatDetailSSERegistered = true
      const sseId = 'task_workflow_chat_' + chatId
      sseDistribute.RegisterReceive(sseId, (data) => {
        if (data && data.line) {
          const line = data.line
          try {
            const obj = JSON.parse(line)
            if (obj.type === 'chat' && obj.subtype === 'completed') {
              this.chatDetailStatus = 'completed'
              this.chatDetailSSERegistered = false
              sseDistribute.UnRegisterReceive(sseId)
              return
            }
          } catch (e) { /* ignore parse errors */ }
          const newMsgs = chatParser.parseChatLines([line])
          if (newMsgs.length > 0) {
            this.chatDetailMessages = [...this.chatDetailMessages, ...newMsgs]
          }
        }
      })
    },
    // 关闭对话详情
    closeChatDetail() {
      if (this.chatDetailSSERegistered) {
        const sseId = 'task_workflow_chat_' + this.chatDetailId
        sseDistribute.UnRegisterReceive(sseId)
        this.chatDetailSSERegistered = false
      }
      this.chatDetailMessages = []
      this.chatDetailId = 0
      this.chatContinueInput = ''
    },
    // 打开对话配置弹窗
    sendToClaudeCode() {
      const prompt = this.issueFixCombinedText
      if (!prompt || !prompt.trim()) {
        this.$helperNotify.warning('提示词为空，无法发送')
        return
      }
      this._pendingChatPrompt = prompt
      // 加载模型列表
      aiSet.AiModelList({}, (res) => {
        if (res.ErrCode === 0) {
          this.chatConfigModels = (res.Data || []).filter(m =>
            (m.request_format || m.provider_type || '') === 'anthropic' && (m.model_type || 'llm') === 'llm'
          )
        }
      })
      // 加载可用目录
      taskWorkflowApi.TaskWorkflowChatDirs(this.workflowId, (res) => {
        if (res.ErrCode === 0) {
          this.chatConfigDirs = res.Data.dirs || []
          if (this.chatConfigDirs.length > 0) {
            this.chatConfigLocalDir = this.chatConfigDirs[0]
          }
        }
      })
      this.chatConfigDialogVisible = true
    },
    // 执行对话配置，开始对话
    startClaudeChat() {
      const prompt = this._pendingChatPrompt
      if (!prompt || !this.chatConfigModelId || !this.chatConfigLocalDir) return
      this.sendingToClaude = true
      taskWorkflowApi.TaskWorkflowChatSend(this.workflowId, prompt, this.chatConfigModelId, this.chatConfigLocalDir, (res) => {
        this.sendingToClaude = false
        if (res.ErrCode === 0 && res.Data) {
          const chatId = res.Data.chat_id
          this.$helperNotify.success('已发送到 claude code 执行')
          this.chatConfigDialogVisible = false
          this.issueFixDialogVisible = false
          this.chatDetailId = chatId
          this.chatDetailStatus = 'running'
          this.chatDetailDialogVisible = true
          this.registerChatSSE(chatId)
          setTimeout(() => { this.loadChatDetail() }, 500)
        } else {
          this.$helperNotify.error(res.ErrMsg || '发送失败')
        }
      })
    },
    // 继续对话
    continueChat() {
      const input = this.chatContinueInput.trim()
      if (!input) return
      this.chatContinueLoading = true
      taskWorkflowApi.TaskWorkflowChatContinue(this.chatDetailId, input, (res) => {
        this.chatContinueLoading = false
        if (res.ErrCode === 0) {
          this.chatContinueInput = ''
          this.chatDetailStatus = 'running'
          this.registerChatSSE(this.chatDetailId)
          setTimeout(() => { this.loadChatDetail() }, 500)
        } else {
          this.$helperNotify.error(res.ErrMsg || '发送失败')
        }
      })
    },
    formatUnixTime(unixTime) {
      const value = Number(unixTime || 0)
      if (value <= 0) {
        return '-'
      }
      return new Date(value * 1000).toLocaleString()
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
      apiManagement.CollectionFoldersBasic({ collection_id: collectionId }, (response) => {
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

.task-workflow-header__eyebrow {
  font-size: 12px;
  color: #909399;
  margin-bottom: 4px;
}

.task-workflow-header__title {
  margin: 0;
  font-size: 22px;
  line-height: 1.3;
  color: #303133;
}

.task-workflow-header__meta {
  display: flex;
  flex-wrap: wrap;
  gap: 12px;
  margin-top: 6px;
  color: #909399;
  font-size: 13px;
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
  background: #409eff;
  width: 14px;
  height: 14px;
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
</style>
