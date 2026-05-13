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
          @click="activeNode = node.key"
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
      width="900px"
      :close-on-click-modal="false"
      destroy-on-close
    >
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
  </div>
</template>

<script>
import { HomeFilled } from '@element-plus/icons-vue'
import GitActionButton from '@/components/base/GitActionButton.vue'
import MemoryFragmentApi from '@/utils/base/memory_fragment'
import taskWorkflowApi from '@/utils/base/task_workflow'
import homeTaskApi from '@/utils/base/home_task'
import baseUtils from '@/utils/base'
import sseDistribute from '@/utils/base/sse_distribute'
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
        this.activeNode = this.firstRunningNodeKey
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
      if (this.workflowId <= 0) return
      taskWorkflowApi.TaskWorkflowIssueFixResolve(this.workflowId, (response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.issueFixResolvedTemplate = response.Data.prompt || ''
        }
      })
    },
    copyIssueFixText() {
      this.copyText(this.issueFixCombinedText, '已复制到剪贴板')
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
  min-height: 300px;
}
</style>
