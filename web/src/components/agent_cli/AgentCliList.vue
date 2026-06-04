<template>
  <div class="agent-cli-page">
    <div class="agent-cli-header-card">
      <div class="agent-cli-header-title">
        <svg class="agent-cli-header-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect x="3" y="4" width="18" height="16" rx="2.5" stroke="currentColor" stroke-width="2" />
          <path d="M7 9H17" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
          <path d="M7 13H13" stroke="currentColor" stroke-width="2" stroke-linecap="round" />
          <path d="M15.5 16.5L17 18L19.5 15.5" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" />
        </svg>
        <span>Agent Cli 管理</span>
      </div>
      <div class="agent-cli-top-notice">
        <div class="agent-cli-top-notice__icon">TIP</div>
        <div class="agent-cli-top-notice__content">
          <div class="agent-cli-top-notice__title">如需聚合统计多个 API，可接入 ai-proxy-stats</div>
          <div class="agent-cli-top-notice__desc">
            该方案支持按 OpenAI 与 Anthropic 兼容格式聚合转发多个 API，并统一查看调用统计。
            <a
              href="https://github.com/w896736588/ai-proxy-stats/blob/main/README_CN.md"
              target="_blank"
              rel="noopener noreferrer"
            >查看 README_CN</a>
          </div>
        </div>
      </div>
      <div class="agent-cli-header-actions">
        <GitActionButton compact @click="openCreateDialog">新建</GitActionButton>
        <GitActionButton compact variant="info" @click="openWebhookDialog">Webhook 配置</GitActionButton>
        <GitActionButton compact variant="warning" @click="chromeDevtoolsDialogVisible = true">ChromeDevTools</GitActionButton>
        <GitActionButton compact variant="primary" @click="openGroupDialog">分组管理</GitActionButton>
      </div>
      <!-- 分组筛选栏 -->
      <div v-if="groupList.length > 0" class="agent-cli-group-filter">
        <span class="agent-cli-group-filter__label">分组：</span>
        <div class="agent-cli-group-filter__tags">
          <span
            class="agent-cli-group-filter__tag"
            :class="{ 'agent-cli-group-filter__tag--active': selectedGroupId === 0 }"
            @click="selectGroup(0)"
          >全部</span>
          <span
            v-for="g in groupList"
            :key="g.id"
            class="agent-cli-group-filter__tag"
            :class="{ 'agent-cli-group-filter__tag--active': selectedGroupId === g.id }"
            @click="selectGroup(g.id)"
          >{{ g.name }}</span>
        </div>
      </div>
    </div>

    <div v-loading="loading" class="agent-cli-list">
      <div v-if="filteredList.length === 0" class="agent-cli-empty">
        暂无 Agent Cli 实例，点击"新建"创建
      </div>
      <div
        v-for="row in filteredList"
        :key="row.id"
        class="agent-cli-card"
        :class="{ 'agent-cli-card--inactive': !row.displayed_enabled }"
      >
        <div class="agent-cli-card__header">
          <div class="agent-cli-card__main">
            <div class="agent-cli-card__title-row">
              <div class="agent-cli-card__title">{{ row.name || '-' }}</div>
              <el-tag size="small" type="info">{{ formatTypeLabel(row.type) }}</el-tag>
              <span class="agent-cli-card__status-dot" :class="row.displayed_enabled ? 'agent-cli-card__status-dot--active' : 'agent-cli-card__status-dot--inactive'"></span>
              <span class="agent-cli-card__status-text">{{ row.displayed_enabled ? '已启用' : '已停止' }}</span>
            </div>
            <div class="agent-cli-card__meta">
              <span>ID：{{ row.id }}</span>
              <span v-if="row.type !== 'codex-cli'">配置文件：{{ row.settings_exists ? '存在' : '不存在' }}</span>
              <span>可选模型：{{ formatModelOptions(row.model_options) }}</span>
              <span>McpServers：{{ row.mcp_server_count || 0 }} 个</span>
            </div>

            <div class="agent-cli-config-table-wrap">
              <table class="agent-cli-config-table">
                <tbody>
                  <tr>
                    <th>启停状态</th>
                    <td>
                      <div class="agent-cli-table-control">
                        <el-switch
                          :model-value="row.displayed_enabled"
                          size="small"
                          :loading="row._togglingEnabled"
                          @change="toggleEnabled(row, $event)"
                        />
                        <span class="agent-cli-switch-line__text">{{ row.displayed_enabled ? '运行中' : '已停止' }}</span>
                      </div>
                    </td>
                    <th>通知配置</th>
                    <td>
                      <el-select
                        v-model="row.webhook_config_id"
                        size="small"
                        placeholder="未配置"
                        clearable
                        class="agent-cli-inline-select"
                        :loading="row._savingWebhook"
                        @change="updateWebhookConfig(row)"
                      >
                        <el-option
                          v-for="wh in webhookOptions"
                          :key="wh.id"
                          :label="wh.name"
                          :value="String(wh.id)"
                        />
                      </el-select>
                    </td>
                  </tr>
                  <tr>
                    <th>分组</th>
                    <td>
                      <el-select
                        v-model="row.group_ids"
                        multiple
                        clearable
                        size="small"
                        placeholder="未分组"
                        class="agent-cli-inline-select"
                        :loading="row._savingGroups"
                        @change="updateGroupIds(row)"
                      >
                        <el-option
                          v-for="g in groupList"
                          :key="g.id"
                          :label="g.name"
                          :value="g.id"
                        />
                      </el-select>
                    </td>
                    <th>请求地址</th>
                    <td class="agent-cli-config-table__value agent-cli-config-table__value--break">
                      <div class="agent-cli-table-cell-text">
                        <span>{{ row.request_url || '-' }}</span>
                        <el-button
                          v-if="row.request_url"
                          size="small"
                          link
                          type="primary"
                          @click="copyText(row.request_url)"
                        >
                          复制
                        </el-button>
                      </div>
                    </td>
                  </tr>
                  <tr v-if="row.type !== 'codex-cli'">
                    <th>路径</th>
                    <td class="agent-cli-config-table__value agent-cli-config-table__value--break">{{ row.settings_path || '-' }}</td>
                    <th>配置文件</th>
                    <td>
                      <el-tag :type="row.settings_exists ? 'success' : 'danger'" size="small">
                        {{ row.settings_exists ? '存在' : '不存在' }}
                      </el-tag>
                    </td>
                  </tr>
                  <tr v-if="row.type !== 'codex-cli'">
                    <th>McpServers</th>
                    <td>{{ row.mcp_server_count || 0 }} 个</td>
                    <th>claude-mem</th>
                    <td>
                      <div class="agent-cli-table-control">
                        <el-switch
                          v-model="row.claude_mem_enabled"
                          size="small"
                          :loading="row._togglingMem"
                          @change="toggleClaudeMem(row)"
                        />
                        <span class="agent-cli-switch-line__text">{{ row.claude_mem_enabled ? '已启用' : '已禁用' }}</span>
                      </div>
                    </td>
                  </tr>
                  <tr v-else>
                    <th>McpServers</th>
                    <td>{{ row.mcp_server_count || 0 }} 个</td>
                    <th>Wire API</th>
                    <td>{{ getCodexWireApi(row) }}</td>
                  </tr>
                  <tr v-if="row.type === 'codex-cli'">
                    <th>WebSocket</th>
                    <td>{{ getCodexSupportsWebsocketsText(row) }}</td>
                    <td colspan="2"></td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>

          <div class="agent-cli-card__actions">
            <GitActionButton compact variant="success" :disabled="!row.displayed_enabled" @click="openAgentExecDialog(row)">执行</GitActionButton>
            <ChatHistoryButton
              variant="info"
              :running="getAgentChatCounts(row.id).running > 0"
              :running-count="getAgentChatCounts(row.id).running"
              :interrupted-count="0"
              :total-count="getAgentChatCounts(row.id).total"
              :unread="getAgentChatCounts(row.id).unread > 0"
              @click="openAgentChatHistory(row)"
            >
              执行历史
            </ChatHistoryButton>
            <GitActionButton
              compact
              variant="primary"
              @click="configureMcp(row)"
            >
              配置DevtoolsMcp
            </GitActionButton>
            <div class="agent-cli-card__actions-tail">
              <GitActionButton compact variant="info" @click="editItem(row)">编辑</GitActionButton>
              <GitActionButton compact variant="danger" @click="deleteItem(row)">删除</GitActionButton>
            </div>
          </div>
        </div>
      </div>
    </div>

    <!-- ChromeDevTools 弹窗 -->
    <el-dialog
      v-model="chromeDevtoolsDialogVisible"
      title="Chrome DevTools 管理"
      width="1000px"
      top="5vh"
      :destroy-on-close="true"
    >
      <iframe
        src="/#/Mcp/chrome-devtools?hide_menu=1&embed=1"
        style="width: 100%; height: 78vh; border: none;"
      />
    </el-dialog>

    <!-- 新建/编辑对话框 -->
    <!-- 新建/编辑弹窗：加宽并允许点击蒙层关闭 / Wider dialog and close on backdrop click. -->
    <el-dialog v-model="dialogVisible" :title="editingId > 0 ? '编辑' : '新建 Agent Cli'" width="720px">
      <el-form :model="form" label-width="140px">
        <el-form-item label="名称">
          <el-input v-model="form.name" :placeholder="form.type === 'codex-cli' ? '默认 Codex CLI' : '默认 Claude Code CLI'" />
        </el-form-item>
        <el-form-item label="类型">
          <el-select v-model="form.type" style="width: 100%" @change="onTypeChange">
            <el-option label="Claude Code CLI" value="claude-code-cli" />
            <el-option label="Codex CLI" value="codex-cli" />
          </el-select>
        </el-form-item>
        <!-- Claude Code CLI 配置 -->
        <template v-if="form.type !== 'codex-cli'">
          <el-form-item label="settings.json 路径" required>
            <el-input v-model="form.settings_path" placeholder="请输入 settings.json 目标绝对路径，文件可先不存在" />
            <div class="agent-cli-form-tip">例如: C:\Users\xxx\.claude\settings.json。文件不存在时，后续写入模型/MCP 配置会自动创建。</div>
          </el-form-item>
          <el-form-item label="模型列表">
            <el-input
              v-model="form.model_list_text"
              type="textarea"
              :rows="4"
              placeholder="每行一个模型；首个模型作为默认模型；执行任务时可再选择具体模型。"
            />
            <div class="agent-cli-form-tip">执行任务时可选择不同于 settings.json 配置的模型。</div>
          </el-form-item>
          <el-form-item label="API Key">
            <el-input v-model="form.api_key" type="password" show-password placeholder="请输入 API Key" />
          </el-form-item>
          <el-form-item label="Base URL">
            <el-input v-model="form.base_url" placeholder="https://api.deepseek.com/anthropic" />
          </el-form-item>
        </template>
        <!-- Codex CLI 配置 -->
        <template v-else>
          <el-form-item label="API Key" required>
            <el-input v-model="form.codex_api_key" type="password" show-password placeholder="请输入 OpenAI API Key" />
          </el-form-item>
          <el-form-item label="模型列表">
            <el-input
              v-model="form.codex_model_list_text"
              type="textarea"
              :rows="4"
              placeholder="每行一个模型；留空则仅使用上方默认模型"
            />
            <div class="agent-cli-form-tip">首个模型会作为默认模型；执行任务时可再选择具体模型。</div>
          </el-form-item>
          <el-form-item label="Base URL">
            <el-input v-model="form.codex_base_url" placeholder="自定义 API 端点（可选）" />
          </el-form-item>
          <el-form-item label="Wire API">
            <el-select v-model="form.codex_wire_api" style="width: 100%">
              <el-option label="responses" value="responses" />
              <el-option label="chat" value="chat" />
            </el-select>
            <div class="agent-cli-form-tip">对应 config.toml 中的 wire_api，请求格式可选 responses 或 chat。</div>
          </el-form-item>
          <el-form-item label="WebSocket">
            <el-switch v-model="form.codex_supports_websockets" />
            <div class="agent-cli-form-tip">关闭后写入 supports_websockets = false，强制走 HTTP。 Turn off to force HTTP instead of WebSocket.</div>
          </el-form-item>
          <el-form-item label="Sandbox Mode">
            <el-input v-model="form.codex_sandbox_mode" placeholder="danger-full-access" />
            <div class="agent-cli-form-tip">默认: danger-full-access</div>
          </el-form-item>
        </template>
        <el-form-item label="分组">
          <el-select v-model="form.group_ids" multiple placeholder="选择分组（可多选）" style="width: 100%">
            <el-option
              v-for="g in groupList"
              :key="g.id"
              :label="g.name"
              :value="g.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item label="Webhook 通知">
          <el-select v-model="form.webhook_config_id" placeholder="不通知" clearable style="width: 100%">
            <el-option
              v-for="wh in webhookOptions"
              :key="wh.id"
              :label="wh.name"
              :value="String(wh.id)"
            />
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="saving" @click="saveItem">保存</el-button>
      </template>
    </el-dialog>

    <!-- 分组管理弹窗 -->
    <el-dialog v-model="groupDialogVisible" title="AgentCli 分组管理" width="560px">
      <div style="margin-bottom: 12px; text-align: right;">
        <el-button type="primary" size="small" @click="openGroupForm(null)">新增分组</el-button>
      </div>
      <el-table :data="groupDialogList" v-loading="groupDialogLoading" size="small" border>
        <el-table-column prop="name" label="名称" min-width="120" />
        <el-table-column prop="sort_order" label="排序" width="70" />
        <el-table-column prop="cli_count" label="关联数" width="70" />
        <el-table-column label="操作" width="130" fixed="right">
          <template #default="{ row }">
            <el-button size="small" link type="primary" @click="openGroupForm(row)">编辑</el-button>
            <el-button size="small" link type="danger" @click="deleteGroup(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 内嵌新增/编辑分组表单 -->
      <div v-if="groupFormVisible" class="webhook-form-section">
        <div class="webhook-form-section__title">{{ groupForm.id > 0 ? '编辑分组' : '新增分组' }}</div>
        <el-form :model="groupForm" label-width="80px" size="small">
          <el-form-item label="名称">
            <el-input v-model="groupForm.name" placeholder="分组名称" />
          </el-form-item>
          <el-form-item label="排序">
            <el-input-number v-model="groupForm.sort_order" :min="0" :max="9999" />
            <div class="agent-cli-form-tip">数值越小越靠前</div>
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="groupSaving" @click="saveGroup">保存</el-button>
            <el-button @click="groupFormVisible = false">取消</el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-dialog>

    <!-- Webhook 配置管理弹窗 -->
    <el-dialog v-model="webhookDialogVisible" title="Webhook 通知配置" width="640px">
      <div style="margin-bottom: 12px; text-align: right;">
        <el-button type="primary" size="small" @click="openWebhookForm(null)">新增</el-button>
      </div>
      <el-table :data="webhookList" v-loading="webhookLoading" size="small" border>
        <el-table-column prop="name" label="名称" min-width="100" />
        <el-table-column prop="type" label="类型" width="90">
          <template #default="{ row }">
            <el-tag size="small">{{ webhookTypeLabel(row.type) }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="webhook_url" label="Webhook 地址" min-width="180" show-overflow-tooltip />
        <el-table-column label="操作" width="180" fixed="right">
          <template #default="{ row }">
            <el-button size="small" link type="primary" @click="openWebhookForm(row)">编辑</el-button>
            <el-button size="small" link type="success" :loading="row._testingWebhook === true" @click="testWebhook(row)">测试</el-button>
            <el-button size="small" link type="danger" @click="deleteWebhook(row)">删除</el-button>
          </template>
        </el-table-column>
      </el-table>

      <!-- 内嵌新增/编辑表单 -->
      <div v-if="webhookFormVisible" class="webhook-form-section">
        <div class="webhook-form-section__title">{{ webhookForm.id > 0 ? '编辑配置' : '新增配置' }}</div>
        <el-form :model="webhookForm" label-width="100px" size="small">
          <el-form-item label="名称">
            <el-input v-model="webhookForm.name" placeholder="如: 前端组钉钉群" />
          </el-form-item>
          <el-form-item label="类型">
            <el-select v-model="webhookForm.type" style="width: 100%">
              <el-option label="钉钉" value="dingtalk" />
              <el-option label="飞书" value="feishu" />
              <el-option label="企业微信" value="wecom" />
            </el-select>
          </el-form-item>
          <el-form-item label="Webhook 地址">
            <el-input v-model="webhookForm.webhook_url" :placeholder="webhookUrlPlaceholder" />
          </el-form-item>
          <el-form-item label="签名密钥">
            <el-input v-model="webhookForm.secret" :placeholder="webhookSecretPlaceholder" />
          </el-form-item>
          <el-form-item label="测试链接">
            <el-input v-model="webhookForm.test_single_url" placeholder="https://example.com/detail (可选，用于测试卡片按钮)" />
          </el-form-item>
          <el-form-item>
            <el-button type="primary" :loading="webhookSaving" @click="saveWebhook">保存</el-button>
            <el-button type="success" plain :loading="webhookTesting" @click="testWebhook(webhookForm)">测试发送</el-button>
            <el-button type="warning" plain :loading="webhookLinkTesting" @click="testWebhookWithLink(webhookForm)">测试卡片按钮</el-button>
            <el-button @click="webhookFormVisible = false">取消</el-button>
          </el-form-item>
        </el-form>
      </div>
    </el-dialog>

    <el-dialog v-model="agentExecDialogVisible" title="执行任务" width="560px" destroy-on-close>
      <el-form label-width="92px">
        <el-form-item label="Agent">
          <el-input :model-value="agentExecCliName" disabled />
        </el-form-item>
        <el-form-item label="历史目录">
          <div class="agent-exec-history-dir-panel">
            <div v-if="agentExecHistoryDirLoading" class="agent-exec-history-dir-panel__state">正在加载历史工作目录...</div>
            <div v-else-if="agentExecHistoryDirs.length === 0" class="agent-exec-history-dir-panel__state">暂无历史工作目录</div>
            <div v-else class="agent-exec-history-dir-list">
              <el-tag
                v-for="historyDir in agentExecHistoryDirs"
                :key="historyDir"
                class="agent-exec-history-dir-tag"
                effect="plain"
                @click="applyAgentExecHistoryDir(historyDir)"
              >
                {{ historyDir }}
              </el-tag>
            </div>
          </div>
        </el-form-item>
        <el-form-item label="工作目录">
          <el-input v-model="agentExecLocalDir" placeholder="请输入本地工作目录绝对路径" />
        </el-form-item>
        <el-form-item label="模型">
          <el-select v-model="agentExecModelName" style="width: 100%;" placeholder="请选择模型">
            <el-option
              v-for="modelName in agentExecModelOptions"
              :key="modelName"
              :label="modelName"
              :value="modelName"
            />
          </el-select>
        </el-form-item>
        <el-form-item v-if="getAgentExecCliType() !== 'codex'" label="思考强度">
          <el-select v-model="agentExecThinkingIntensity" style="width: 100%;" placeholder="请选择思考强度">
            <el-option label="低" value="低" />
            <el-option label="中等" value="中等" />
            <el-option label="高" value="高" />
            <el-option label="很高" value="很高" />
            <el-option label="最高" value="最高" />
          </el-select>
        </el-form-item>
        <el-form-item label="提示词">
          <el-input
            v-model="agentExecPrompt"
            type="textarea"
            :rows="10"
            placeholder="请输入要发送给当前 Agent CLI 的提示词"
          />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="agentExecDialogVisible = false">取消</el-button>
        <el-button type="primary" :loading="agentExecLoading" @click="execAgentPrompt">开始执行</el-button>
      </template>
    </el-dialog>

    <ChatHistoryDialog
      ref="agentChatHistoryDialog"
      v-model="agentChatHistoryVisible"
      :title="'执行历史 - ' + (agentChatHistoryTitle || 'Agent CLI')"
      :loading="agentChatHistoryLoading"
      :items="agentChatHistoryList"
      :selected-id="agentChatDetailId"
      :detail-title="agentChatHistoryTitle || '-'"
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
      :scroll-button-visible="agentChatDetailShowScrollBtn"
      :running-text="'等待 Agent CLI 响应...'"
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
      @select="onAgentChatRowClick"
      @update:continueInput="chatContinueInput = $event"
      @continue="continueChat"
      @new-chat="startNewChatFromHistory"
      @stop="stopChat"
      @scroll="onAgentChatDetailScroll"
      @scroll-to-bottom="scrollAgentChatToBottom(true)"
      @closed="onAgentChatHistoryClosed"
    />

  </div>
</template>

<script>
import agentCliApi from '../../utils/base/agent_cli'
import GitActionButton from '@/components/base/GitActionButton.vue'
import ChatHistoryButton from '@/components/shared/ChatHistoryButton.vue'
import ChatHistoryDialog from '@/components/shared/ChatHistoryDialog.vue'
import taskWorkflowApi from '@/utils/base/task_workflow'
import baseUtils from '@/utils/base'
import sseDistribute from '@/utils/base/sse_distribute'
import chatParser from '@/utils/chat_parser'
import taskProgressStore from '@/utils/task_progress_store'
import MarkdownIt from 'markdown-it'
import copy from '@/utils/base/copy'

// AGENT_CLI_ENABLED_SORT_TRUE 启用状态排序值，启用项排在前面。 // Sort weight for enabled rows so active items stay at the top.
const AGENT_CLI_ENABLED_SORT_TRUE = 1
// AGENT_CLI_ENABLED_SORT_FALSE 禁用状态排序值，禁用项排在后面。 // Sort weight for disabled rows so inactive items move below active ones.
const AGENT_CLI_ENABLED_SORT_FALSE = 0
// AGENT_EXEC_CACHE_PREFIX 按 Agent CLI 记录最近一次执行配置。 // Cache key prefix for per-Agent execution settings.
const AGENT_EXEC_CACHE_PREFIX = 'agent_cli_exec_'
// AGENT_CLI_GROUP_CACHE_KEY 记住上次选中的分组 ID。 // LocalStorage key for remembering the last selected group filter.
const AGENT_CLI_GROUP_CACHE_KEY = 'agent_cli_selected_group'
// AGENT_EXEC_SHARED_HISTORY_LIMIT 控制执行弹窗共享历史目录展示数量。 // Max shared working directories shown in the execution dialog.
const AGENT_EXEC_SHARED_HISTORY_LIMIT = 20
// AGENT_EXEC_SHARED_PROMPT_KEY 所有 Agent CLI 共用最近一次执行提示词。 // Shared prompt cache key for all Agent CLI cards.
const AGENT_EXEC_SHARED_PROMPT_KEY = 'agent_cli_exec_shared_prompt'
// markdown-it 实例，用于在"执行历史"对话框中渲染 markdown（包括表格）。 // Markdown renderer for execution history detail.
const md = new MarkdownIt({ html: true, breaks: true, linkify: true })

export default {
  components: {
    GitActionButton,
    ChatHistoryButton,
    ChatHistoryDialog,
  },
  data() {
    return {
      loading: false,
      list: [],
      // 新建/编辑
      dialogVisible: false,
      chromeDevtoolsDialogVisible: false,
      editingId: 0,
      saving: false,
      form: {
        name: '',
        type: 'claude-code-cli',
        settings_path: '',
        webhook_config_id: '',
        enabled: 1,
        model_name: '',
        model_list_text: '',
        api_key: '',
        base_url: '',
        // Codex CLI 专属字段
        codex_api_key: '',
        codex_model_list_text: '',
        codex_base_url: '',
        codex_wire_api: 'responses',
        codex_sandbox_mode: '',
        codex_supports_websockets: false,
        // 分组多选
        group_ids: [],
      },
      // 分组筛选
      groupList: [],
      selectedGroupId: 0,
      // 分组管理弹窗
      groupDialogVisible: false,
      groupDialogLoading: false,
      groupDialogList: [],
      groupFormVisible: false,
      groupSaving: false,
      groupForm: {
        id: 0,
        name: '',
        sort_order: 0,
      },
      // webhook 配置
      webhookDialogVisible: false,
      webhookLoading: false,
      webhookList: [],
      webhookOptions: [],
      webhookFormVisible: false,
      webhookSaving: false,
      webhookTesting: false,
      webhookLinkTesting: false,
      webhookForm: {
        id: 0,
        name: '',
        type: 'dingtalk',
        webhook_url: '',
        secret: '',
        test_single_url: '',
      },
      agentExecDialogVisible: false,
      agentExecLoading: false,
      agentExecCliId: 0,
      agentExecCliName: '',
      agentExecPrompt: '',
      agentExecLocalDir: '',
      agentExecHistoryDirLoading: false,
      agentExecHistoryDirs: [],
      agentExecModelName: '',
      agentExecThinkingIntensity: '高',
      agentChatHistoryVisible: false,
      agentChatHistoryLoading: false,
      agentChatHistoryTitle: '',
      agentChatHistoryCliId: 0,
      agentChatHistoryList: [],
      agentChatCounts: {},
      agentUnreadTotal: 0,
      _agentUnreadSseId: '',
      agentChatDetailId: 0,
      chatDetailId: 0,
      chatDetailStatus: '',
      chatDetailPrompt: '',
      chatDetailSessionId: '',
      chatDetailModelName: '',
      chatDetailAgentName: '',
      chatDetailLocalDir: '',
      chatDetailThinkingIntensity: '',
      chatDetailCliType: 'claude',
      chatDetailLastUsageSummary: null,
      chatDetailMessages: [],
      chatDetailSSELines: [],
      chatDetailAutoScroll: true,
      chatDetailSSERegistered: false,
      _backgroundChatEventSources: {},
      agentChatDetailShowScrollBtn: false,
      chatContinueInput: '',
      chatContinueLoading: false,
      thinkingStreamElapsed: 0,
    }
  },
  computed: {
    // filteredList 根据选中的分组筛选列表；选中"全部"(0)时不过滤。 // Returns the filtered Agent CLI list based on the selected group.
    filteredList() {
      if (this.selectedGroupId === 0) return this.list
      return this.list.filter(item => {
        const gids = Array.isArray(item.group_ids) ? item.group_ids : []
        return gids.includes(this.selectedGroupId)
      })
    },
    // agentExecModelOptions 返回当前卡片可选模型列表；无配置时回退到 current_model。 // Returns model options for the selected Agent CLI card.
    agentExecModelOptions() {
      const cli = this.getAgentExecCli()
      if (!cli) return []
      const rawOptions = Array.isArray(cli.model_options) ? cli.model_options : []
      const normalizedOptions = rawOptions.map(item => String(item || '').trim()).filter(Boolean)
      if (normalizedOptions.length > 0) {
        return normalizedOptions
      }
      const currentModel = String(cli.current_model || '').trim()
      return currentModel && currentModel !== '-' ? [currentModel] : []
    },
    webhookUrlPlaceholder() {
      return this.webhookUrlPlaceholderByType(this.webhookForm.type)
    },
    webhookSecretPlaceholder() {
      return this.webhookSecretPlaceholderByType(this.webhookForm.type)
    },
  },
  mounted() {
    this.loadList()
    this.loadWebhookOptions()
    this.loadGroupList()
    this.ensureAgentUnreadSse()
    // 恢复上次选中的分组
    try {
      const cached = parseInt(localStorage.getItem(AGENT_CLI_GROUP_CACHE_KEY))
      if (cached > 0) this.selectedGroupId = cached
    } catch {}
  },
  activated() {
    this.ensureAgentUnreadSse()
    this.loadList()
    if (this.agentChatHistoryVisible && Number(this.agentChatHistoryCliId || 0) > 0) {
      this.openAgentChatHistory({
        id: this.agentChatHistoryCliId,
        name: this.agentChatHistoryTitle,
      }, this.agentChatDetailId || this.chatDetailId)
    }
  },
  beforeUnmount() {
    this.closeChatDetail()
    this.stopAllBackgroundChatStreams()
    this._stopAgentChatHistoryDurationTimer()
    this.unregisterAgentUnreadSse()
  },
  watch: {
    agentExecPrompt(value) {
      try {
        localStorage.setItem(AGENT_EXEC_SHARED_PROMPT_KEY, String(value || ''))
      } catch {
        return
      }
    },
  },
  methods: {
    // openWebhookDialog 打开 webhook 配置弹窗并同步刷新列表。 // openWebhookDialog opens the webhook dialog and refreshes its list before display.
    openWebhookDialog() {
      this.webhookDialogVisible = true
      this.loadWebhookList()
    },
    // formatTypeLabel 统一格式化实例类型文案，避免页面直接暴露内部值。 // formatTypeLabel normalizes instance type labels so the UI does not expose raw internal values.
    formatTypeLabel(type) {
      if (type === 'codex-cli') {
        return 'Codex CLI'
      }
      if (type === 'claude-code-cli') {
        return 'Claude Code CLI'
      }
      return type || '-'
    },
    loadList() {
      this.loading = true
      agentCliApi.AgentCliList((response) => {
        this.loading = false
        if (response && response.ErrCode === 0 && response.Data) {
          const items = response.Data.list || []
          items.forEach(item => {
            item.webhook_config_id = item.webhook_config_id ? String(item.webhook_config_id) : ''
            item.displayed_enabled = !!item.displayed_enabled
            item.group_ids = Array.isArray(item.group_ids) ? item.group_ids : []
            item._lastGroupIds = [...item.group_ids]
            item._lastWebhookConfigId = item.webhook_config_id
            item._savingWebhook = false
            item._savingGroups = false
          })
          this.list = this.sortAgentCliList(items)
          this.loadAgentChatCounts()
        }
      })
    },
    ensureAgentUnreadSse() {
      if (this._agentUnreadSseId) return
      const nextId = 'agent_cli_unread_global'
      this._agentUnreadSseId = nextId
      sseDistribute.InitFromLoginStatus().then((created) => {
        if (!created && !sseDistribute.GetSseClientId()) return
        sseDistribute.RegisterReceive(nextId, this.handleAgentUnreadSseMessage)
      })
    },
    unregisterAgentUnreadSse() {
      if (!this._agentUnreadSseId) return
      sseDistribute.UnRegisterReceive(this._agentUnreadSseId)
      this._agentUnreadSseId = ''
    },
    handleAgentUnreadSseMessage(data) {
      if (!data || data.type !== 'agent_cli_unread_global') return
      this.agentUnreadTotal = Number(data.agent_cli_unread || 0)
      this.loadAgentChatCounts()
    },
    // loadAgentChatCounts 刷新每个卡片的执行历史计数，用于按钮数字和执行中动画。 // Refreshes per-card execution counters and running animation state.
    loadAgentChatCounts() {
      const rows = Array.isArray(this.list) ? this.list.filter(item => item.id > 0) : []
      rows.forEach((row) => {
        agentCliApi.AgentChatListByAgentCli(row.id, (response) => {
          if (!(response && response.ErrCode === 0 && response.Data)) return
          const list = Array.isArray(response.Data.list) ? response.Data.list : []
          this.agentChatCounts = {
            ...this.agentChatCounts,
            [row.id]: {
              running: list.filter(item => item.status === 'running').length,
              interrupted: list.filter(item => item.status === 'interrupted').length,
              total: list.length,
              unread: list.filter(item => item.is_read === false && item.status !== 'running').length,
            },
          }
        })
      })
    },
    markAgentChatRunningLocally(agentCliId, chatId, extra = {}) {
      const normalizedAgentCliId = Number(agentCliId || 0)
      const normalizedChatId = Number(chatId || 0)
      if (normalizedAgentCliId <= 0 || normalizedChatId <= 0) return
      const currentCounts = this.getAgentChatCounts(normalizedAgentCliId)
      this.agentChatCounts = {
        ...this.agentChatCounts,
        [normalizedAgentCliId]: {
          running: currentCounts.running + 1,
          interrupted: currentCounts.interrupted,
          total: Math.max(currentCounts.total + 1, currentCounts.running + currentCounts.interrupted + 1),
          unread: currentCounts.unread,
        },
      }
      if (!this.agentChatHistoryVisible || Number(this.agentChatHistoryCliId || 0) !== normalizedAgentCliId) return
      const existing = this.agentChatHistoryList.find(item => Number(item.id || 0) === normalizedChatId)
      if (existing) {
        existing.status = 'running'
        existing.is_read = true
        this.agentChatHistoryList = this.agentChatHistoryList.slice()
        this.syncBackgroundChatStreams(this.agentChatHistoryList, this.agentChatDetailId || this.chatDetailId)
        return
      }
      this.agentChatHistoryList = [{
        id: normalizedChatId,
        agent_cli_id: normalizedAgentCliId,
        prompt_type: 'agent_cli_manual',
        status: 'running',
        is_read: true,
        line_count: 0,
        created_at: new Date().toISOString(),
        prompt: extra.prompt || '',
        local_dir: extra.localDir || '',
        model_name: extra.modelName || '',
        thinking_intensity: extra.thinkingIntensity || '',
        cli_type: extra.cliType || 'claude',
      }, ...this.agentChatHistoryList]
      this.syncBackgroundChatStreams(this.agentChatHistoryList, this.agentChatDetailId || this.chatDetailId)
    },
    getAgentChatCounts(agentCliId) {
      return this.agentChatCounts[agentCliId] || { running: 0, interrupted: 0, total: 0, unread: 0 }
    },
    getAgentExecCli() {
      return this.list.find(item => Number(item.id) === Number(this.agentExecCliId)) || null
    },
    getAgentExecCliType() {
      const cli = this.getAgentExecCli()
      if (!cli) return 'claude'
      return cli.type === 'codex-cli' ? 'codex' : 'claude'
    },
    getAgentExecCacheKey(agentCliId) {
      return AGENT_EXEC_CACHE_PREFIX + String(agentCliId || 0)
    },
    getAgentExecCache(agentCliId) {
      try {
        const raw = localStorage.getItem(this.getAgentExecCacheKey(agentCliId))
        return raw ? JSON.parse(raw) : null
      } catch {
        return null
      }
    },
    saveAgentExecCache() {
      const data = {
        localDir: this.agentExecLocalDir,
        modelName: this.agentExecModelName,
        thinkingIntensity: this.agentExecThinkingIntensity,
        prompt: this.agentExecPrompt,
      }
      localStorage.setItem(this.getAgentExecCacheKey(this.agentExecCliId), JSON.stringify(data))
    },
    openAgentExecDialog(row) {
      if (!row?.displayed_enabled) {
        this.$message.warning('请先启用当前 Agent CLI')
        return
      }
      this.agentExecCliId = row.id
      this.agentExecCliName = row.name || '-'
      const cached = this.getAgentExecCache(row.id)
      let sharedPrompt = ''
      try {
        sharedPrompt = localStorage.getItem(AGENT_EXEC_SHARED_PROMPT_KEY) || ''
      } catch {
        sharedPrompt = ''
      }
      this.agentExecLocalDir = cached?.localDir || ''
      this.agentExecHistoryDirs = []
      this.agentExecModelName = cached?.modelName || ''
      this.agentExecThinkingIntensity = cached?.thinkingIntensity || '高'
      this.agentExecPrompt = sharedPrompt || cached?.prompt || ''
      if (this.agentExecModelOptions.length === 1 && !this.agentExecModelName) {
        this.agentExecModelName = this.agentExecModelOptions[0]
      }
      this.agentExecDialogVisible = true
      this.loadAgentExecHistoryDirs()
    },
    // loadAgentExecHistoryDirs 加载所有 Agent CLI 共享的历史工作目录。 // Loads globally shared history directories across all Agent CLI cards.
    loadAgentExecHistoryDirs() {
      this.agentExecHistoryDirLoading = true
      const rows = Array.isArray(this.list) ? this.list.filter(item => Number(item.id) > 0) : []
      if (rows.length === 0) {
        this.agentExecHistoryDirLoading = false
        this.agentExecHistoryDirs = []
        return
      }
      let pending = rows.length
      const mergedList = []
      rows.forEach((row) => {
        agentCliApi.AgentChatListByAgentCli(row.id, (response) => {
          if (response && response.ErrCode === 0 && response.Data) {
            const list = Array.isArray(response.Data.list) ? response.Data.list : []
            mergedList.push(...list)
          }
          pending -= 1
          if (pending <= 0) {
            this.agentExecHistoryDirLoading = false
            this.agentExecHistoryDirs = this.extractAgentExecHistoryDirs(mergedList)
          }
        })
      })
    },
    // extractAgentExecHistoryDirs 从执行历史中提取去重后的目录列表，最近使用优先。 // Extracts a unique directory list from execution history while preserving recency order.
    extractAgentExecHistoryDirs(list) {
      const sortedList = Array.isArray(list) ? [...list] : []
      sortedList.sort((firstItem, secondItem) => {
        const secondTime = new Date(String(secondItem?.created_at || '').replace(/-/g, '/')).getTime() || 0
        const firstTime = new Date(String(firstItem?.created_at || '').replace(/-/g, '/')).getTime() || 0
        return secondTime - firstTime
      })
      const dirSet = new Set()
      const dirList = []
      sortedList.forEach((item) => {
        const localDir = String(item?.local_dir || '').trim()
        // 空目录不展示，避免把无效历史写回输入框。 // Skip empty directories so invalid history values are never re-applied.
        if (!localDir || dirSet.has(localDir) || dirList.length >= AGENT_EXEC_SHARED_HISTORY_LIMIT) {
          return
        }
        dirSet.add(localDir)
        dirList.push(localDir)
      })
      return dirList
    },
    // applyAgentExecHistoryDir 点击历史目录后直接回填到工作目录输入框。 // Applies a clicked history directory into the working-directory input immediately.
    applyAgentExecHistoryDir(historyDir) {
      this.agentExecLocalDir = String(historyDir || '').trim()
    },
    execAgentPrompt() {
      if (!this.agentExecCliId) {
        this.$message.warning('Agent 实例不存在')
        return
      }
      const currentCli = this.getAgentExecCli()
      if (!currentCli || !currentCli.displayed_enabled) {
        this.$message.warning('请选择已启用的 Agent 实例')
        return
      }
      if (!String(this.agentExecLocalDir || '').trim()) {
        this.$message.warning('请输入工作目录')
        return
      }
      if (!String(this.agentExecPrompt || '').trim()) {
        this.$message.warning('请输入提示词')
        return
      }
      if (this.agentExecModelOptions.length > 0 && !this.agentExecModelName) {
        this.$message.warning('请选择模型')
        return
      }
      this.agentExecLoading = true
      this.saveAgentExecCache()
      agentCliApi.AgentChatSend({
        agent_cli_id: this.agentExecCliId,
        prompt: this.agentExecPrompt,
        prompt_type: 'agent_cli_manual',
        local_dir: this.agentExecLocalDir.trim(),
        cli_type: this.getAgentExecCliType(),
        model_name: this.agentExecModelName,
        thinking_intensity: this.agentExecThinkingIntensity,
      }, (response) => {
          this.agentExecLoading = false
          if (!(response && response.ErrCode === 0 && response.Data)) {
            this.$message.error(response?.ErrMsg || '发送失败')
            return
          }
          const chatId = response.Data.chat_id
          this.$message.success('已发送到 Agent CLI 执行')
          this.agentExecDialogVisible = false
          this.chatDetailId = chatId
          this.agentChatDetailId = chatId
          this.chatDetailStatus = 'running'
          this.chatDetailCliType = this.getAgentExecCliType()
          this.chatDetailSSELines = []
          this.chatDetailMessages = []
          taskProgressStore.reset()
          this._initialSseRetryCount = 0
          this.markAgentChatRunningLocally(this.agentExecCliId, chatId, {
            prompt: this.agentExecPrompt,
            localDir: this.agentExecLocalDir.trim(),
            modelName: this.agentExecModelName,
            thinkingIntensity: this.agentExecThinkingIntensity,
            cliType: this.getAgentExecCliType(),
          })
          this.connectChatStream(chatId, null, true)
          this.loadChatDetail()
          this.loadAgentChatCounts()
          if (currentCli) {
            this.openAgentChatHistory(currentCli, chatId)
          }
        })
    },
    openAgentChatHistory(row, focusChatId) {
      this.agentChatHistoryCliId = row.id
      this.agentChatHistoryTitle = row.name || '-'
      this.agentChatHistoryVisible = true
      this.agentChatHistoryLoading = true
      this.agentChatDetailId = 0
      agentCliApi.AgentChatListByAgentCli(row.id, (response) => {
        this.agentChatHistoryLoading = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.$message.error(response?.ErrMsg || '加载执行历史失败')
          return
        }
        this.agentChatHistoryList = Array.isArray(response.Data.list) ? response.Data.list : []
        this._startAgentChatHistoryDurationTimer()
        this.syncBackgroundChatStreams(this.agentChatHistoryList, focusChatId || this.agentChatDetailId || this.chatDetailId)
        this.agentChatCounts = {
          ...this.agentChatCounts,
          [row.id]: {
            running: this.agentChatHistoryList.filter(item => item.status === 'running').length,
            interrupted: this.agentChatHistoryList.filter(item => item.status === 'interrupted').length,
            total: this.agentChatHistoryList.length,
            unread: this.agentChatHistoryList.filter(item => item.is_read === false && item.status !== 'running').length,
          },
        }
        if (focusChatId) {
          const found = this.agentChatHistoryList.find(item => item.id === focusChatId)
          if (found) {
            this.onAgentChatRowClick(found)
            return
          }
        }
        if (this.agentChatHistoryList.length > 0) {
          this.onAgentChatRowClick(this.agentChatHistoryList[0])
        }
      })
    },
    onAgentChatRowClick(row) {
      if (this.agentChatDetailId === row.id) return
      this.agentChatDetailId = row.id
      this.chatDetailId = row.id
      this.chatDetailStatus = row.status
      this.chatDetailAutoScroll = true
      this.agentChatDetailShowScrollBtn = false
      if (this._chatEventSource && this._sseChatId !== row.id) {
        this._chatEventSource.close()
        this._chatEventSource = null
        this._sseChatId = 0
      }
      // 中文注释：先关闭旧前台 SSE，再把旧选中且仍在运行的对话切到后台监听。
      // English comment: Close the previous foreground SSE first, then reattach the previously selected running chat as a background stream if needed.
      this.syncBackgroundChatStreams(this.agentChatHistoryList, row.id)
      if (row.is_read === false && row.status !== 'running') {
        agentCliApi.AgentChatMarkRead(row.id, (res) => {
          if (res && res.ErrCode === 0) {
            const item = this.agentChatHistoryList.find(i => i.id === row.id)
            if (item) item.is_read = true
            this.agentChatCounts = {
              ...this.agentChatCounts,
              [this.agentChatHistoryCliId]: {
                running: this.agentChatHistoryList.filter(item => item.status === 'running').length,
                interrupted: this.agentChatHistoryList.filter(item => item.status === 'interrupted').length,
                total: this.agentChatHistoryList.length,
                unread: this.agentChatHistoryList.filter(item => item.is_read === false && item.status !== 'running').length,
              },
            }
          }
        })
      }
      if (this._sseChatId !== row.id) {
        this.chatDetailSSELines = []
        this.chatDetailMessages = []
        this._thinkingStreamStartTime = 0
        this.thinkingStreamElapsed = 0
        taskProgressStore.reset()
        this.loadChatDetail()
      } else {
        this.$nextTick(() => { this.scrollAgentChatToBottom() })
      }
    },
    onAgentChatHistoryClosed() {
      this._stopAgentChatHistoryDurationTimer()
      this.stopAllBackgroundChatStreams()
    },
    onAgentChatDetailScroll() {
      if (this._autoScrollLocked) return
      const dialog = this.$refs.agentChatHistoryDialog
      if (!dialog || !dialog.isDetailNearBottom) return
      const atBottom = dialog.isDetailNearBottom(30)
      if (atBottom) {
        this.chatDetailAutoScroll = true
        this.agentChatDetailShowScrollBtn = false
      } else {
        this.chatDetailAutoScroll = false
        this.agentChatDetailShowScrollBtn = true
      }
    },
    scrollAgentChatToBottom(force) {
      if (!force && !this.chatDetailAutoScroll) return
      if (force) {
        this.chatDetailAutoScroll = true
        this.agentChatDetailShowScrollBtn = false
      }
      this.$nextTick(() => {
        const dialog = this.$refs.agentChatHistoryDialog
        if (dialog && dialog.scrollDetailToBottom) {
          dialog.scrollDetailToBottom('auto')
        }
      })
    },
    _startAgentChatHistoryDurationTimer() {
      this._stopAgentChatHistoryDurationTimer()
      this._chatHistoryDurationTimer = setInterval(() => {
        if (this.agentChatHistoryList.some(item => item.status === 'running')) {
          this.agentChatHistoryList = this.agentChatHistoryList.slice()
        }
        if (this._sseChatId > 0 && this.chatDetailSSELines.length > 0) {
          const count = this.chatDetailSSELines.length
          const item = this.agentChatHistoryList.find(row => row.id === this._sseChatId)
          if (item && item.line_count !== count) {
            item.line_count = count
          }
        }
      }, 1000)
    },
    _stopAgentChatHistoryDurationTimer() {
      if (this._chatHistoryDurationTimer) {
        clearInterval(this._chatHistoryDurationTimer)
        this._chatHistoryDurationTimer = null
      }
    },
    openCreateDialog() {
      this.editingId = 0
      this.form = {
        name: '',
        type: 'claude-code-cli',
        settings_path: '',
        webhook_config_id: '',
        enabled: 1,
        model_name: '',
        model_list_text: '',
        api_key: '',
        base_url: '',
        codex_api_key: '',
        codex_model_list_text: '',
        codex_base_url: '',
        codex_wire_api: 'responses',
        codex_sandbox_mode: '',
        codex_supports_websockets: false,
        group_ids: [],
      }
      this.dialogVisible = true
    },
    onTypeChange() {
      // 类型切换时同步默认启停状态 / Sync default enabled status when type changes.
      this.form.enabled = this.form.type === 'codex-cli' ? 0 : 1
    },
    saveItem() {
      const isCodex = this.form.type === 'codex-cli'
      const claudeModels = this.parseModelList(this.form.model_list_text, '')
      const codexModels = this.parseModelList(this.form.codex_model_list_text, '')
      if (isCodex) {
        // Codex 现在只保留模型列表字段，首项作为默认模型 / Codex now only uses the model list and the first item becomes default.
        if (!this.form.codex_api_key.trim()) {
          this.$message.warning('请输入 API Key')
          return
        }
        if (codexModels.length === 0) {
          this.$message.warning('请输入模型列表')
          return
        }
      } else {
        if (!this.form.settings_path.trim()) {
          this.$message.warning('请输入 settings.json 目标路径')
          return
        }
      }
      this.saving = true
      const data = {
        id: this.editingId,
        name: this.form.name,
        type: this.form.type,
        settings_path: isCodex ? '' : this.form.settings_path.trim(),
        enabled: this.form.enabled,
        webhook_config_id: parseInt(this.form.webhook_config_id) || 0,
      }
      // Codex 类型：将配置序列化为 config JSON
      if (isCodex) {
        data.config = JSON.stringify({
          api_key: this.form.codex_api_key.trim(),
          model: codexModels[0] || '',
          models: codexModels,
          base_url: this.form.codex_base_url.trim() || undefined,
          wire_api: this.form.codex_wire_api || 'responses',
          sandbox_mode: this.form.codex_sandbox_mode.trim() || undefined,
          supports_websockets: this.form.codex_base_url.trim() ? !!this.form.codex_supports_websockets : undefined,
        })
      }
      agentCliApi.AgentCliSave(data, (response) => {
        if (response && response.ErrCode === 0) {
          // 新建时从返回值取 ID，后续 DeepSeek 写入依赖此 ID
          if (!this.editingId && response.Data && response.Data.id) {
            this.editingId = response.Data.id
          }
          // 保存分组关联，并同步本地列表，避免页面仍显示旧分组数据。
          this._saveGroupRel(this.editingId)
          // Claude 类型：只要配置项有输入就同步写入 settings.json，避免仅改模型时运行仍读取旧配置。
          if (!isCodex && (claudeModels.length > 0 || this.form.api_key.trim() || this.form.base_url.trim())) {
            const dsData = {
              id: this.editingId,
              model_name: claudeModels.length > 0 ? claudeModels[0] : '',
              model_list: claudeModels,
              api_key: this.form.api_key.trim(),
              base_url: this.form.base_url.trim(),
            }
            agentCliApi.AgentCliWriteDeepSeek(dsData, (dsResponse) => {
              this.saving = false
              if (dsResponse && dsResponse.ErrCode === 0) {
                this.dialogVisible = false
                this.$message.success('保存成功')
                this.loadList()
              } else {
                this.$message.error(dsResponse?.ErrMsg || '密钥保存失败')
              }
            })
            return
          }
          this.saving = false
          this.dialogVisible = false
          this.$message.success('保存成功')
          this.loadList()
        } else {
          this.saving = false
          this.$message.error(response?.ErrMsg || '保存失败')
        }
      })
    },
    deleteItem(item) {
      this.$confirm(`确定要删除 "${item.name}" 吗？` + (item.type !== 'codex-cli' ? '此操作不删除 settings.json 文件。' : ''), '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        agentCliApi.AgentCliDelete(item.id, (response) => {
          if (response && response.ErrCode === 0) {
            this.$message.success('删除成功')
            this.loadList()
          } else {
            this.$message.error(response?.ErrMsg || '删除失败')
          }
        })
      }).catch(() => {})
    },
    configureMcp(item) {
      const loading = this.$loading({ text: '正在写入 mcpServers 配置...' })
      agentCliApi.AgentCliWriteMcpServers(item.id, (response) => {
        loading.close()
        if (response && response.ErrCode === 0) {
          this.$message.success('ChromeDevtoolsMcp 配置已写入')
          this.loadList()
        } else {
          this.$message.error(response?.ErrMsg || '配置失败')
        }
      })
    },
    toggleClaudeMem(item) {
      item._togglingMem = true
      agentCliApi.AgentCliToggleClaudeMem({ id: item.id, enable: item.claude_mem_enabled }, (response) => {
        item._togglingMem = false
        if (response && response.ErrCode === 0) {
          this.$message.success(`claude-mem 已${item.claude_mem_enabled ? '启用' : '禁用'}`)
        } else {
          this.$message.error(response?.ErrMsg || '操作失败')
          item.claude_mem_enabled = !item.claude_mem_enabled
        }
      })
    },
    toggleEnabled(item, enabled) {
      const previousDisplayedEnabled = !!item.displayed_enabled
      const previousEnabled = Number(item.enabled || 0)
      item._togglingEnabled = true
      item.displayed_enabled = !!enabled
      item.enabled = enabled ? 1 : 0
      // 启停切换后先本地重排，保证启用实例即时显示在顶部。 // Re-sort immediately after toggling so enabled rows move to the top without waiting for reload.
      this.list = this.sortAgentCliList(this.list)
      agentCliApi.AgentCliToggleEnabled({ id: item.id, enable: !!enabled }, (response) => {
        item._togglingEnabled = false
        if (response && response.ErrCode === 0) {
          this.$message.success(`Agent CLI 已${enabled ? '启用' : '停止'}`)
          this.loadList()
        } else {
          this.$message.error(response?.ErrMsg || '操作失败')
          item.displayed_enabled = previousDisplayedEnabled
          item.enabled = previousEnabled
          this.list = this.sortAgentCliList(this.list)
        }
      })
    },
    // 打开编辑对话框，预填当前条目数据并读取配置
    editItem(item) {
      this.editingId = item.id
      const isCodex = item.type === 'codex-cli'
      this.form = {
        name: item.name || '',
        type: item.type || 'claude-code-cli',
        settings_path: item.settings_path || '',
        webhook_config_id: item.webhook_config_id || '',
        enabled: item.enabled || 0,
        model_name: '',
        model_list_text: '',
        api_key: '',
        base_url: '',
        codex_api_key: '',
        codex_model_list_text: '',
        codex_base_url: '',
        codex_wire_api: 'responses',
        codex_sandbox_mode: '',
        codex_supports_websockets: false,
        group_ids: Array.isArray(item.group_ids) ? [...item.group_ids] : [],
      }
      // Codex: 从 config JSON 预填
      if (isCodex && item.config) {
        try {
          const cfg = JSON.parse(item.config)
          this.form.codex_api_key = cfg.api_key || ''
          this.form.codex_model_list_text = Array.isArray(cfg.models) ? cfg.models.join('\n') : (cfg.model || '')
          this.form.codex_base_url = cfg.base_url || ''
          this.form.codex_wire_api = cfg.wire_api || 'responses'
          this.form.codex_sandbox_mode = cfg.sandbox_mode || ''
          this.form.codex_supports_websockets = cfg.supports_websockets !== false
        } catch (e) {
          // 解析失败时忽略，保持编辑弹窗可用。 // Ignore parse errors and keep the editor usable.
        }
      }
      this.dialogVisible = true
      // Claude: 读取 settings.json 以预填密钥字段
      if (!isCodex) {
        agentCliApi.AgentCliReadSettings(item.id, (response) => {
          if (response && response.ErrCode === 0 && response.Data && response.Data.content) {
            try {
              const config = JSON.parse(response.Data.content)
              // 优先读取 dtool_models 列表，兼容旧版单模型配置
              const modelList = Array.isArray(config.dtool_models) ? config.dtool_models : []
              if (modelList.length > 0) {
                this.form.model_list_text = modelList.join('\n')
              } else if (config.model) {
                this.form.model_list_text = config.model
              }
              if (config.env) {
                this.form.api_key = config.env.ANTHROPIC_AUTH_TOKEN || ''
                this.form.base_url = config.env.ANTHROPIC_BASE_URL || ''
              }
            } catch(e) {
              // 读取历史配置失败时忽略。 // Ignore malformed history config.
            }
          }
        })
      }
    },
    // parseModelList 解析文本模型列表，并确保默认模型排在首位。
    // parseModelList parses textarea models and keeps the default model at the front.
    parseModelList(modelListText, defaultModel) {
      const list = String(modelListText || '')
        .split(/\r?\n/)
        .map(item => item.trim())
        .filter(Boolean)
      const merged = []
      const seen = new Set()
      const normalizedDefaultModel = String(defaultModel || '').trim()
      if (normalizedDefaultModel) {
        merged.push(normalizedDefaultModel)
        seen.add(normalizedDefaultModel)
      }
      list.forEach(modelName => {
        if (seen.has(modelName)) {
          return
        }
        merged.push(modelName)
        seen.add(modelName)
      })
      return merged
    },
    // formatModelOptions 统一格式化模型列表文案，避免空值时出现脏展示。
    // formatModelOptions normalizes the model list text and keeps the card display compact.
    formatModelOptions(modelOptions) {
      if (!Array.isArray(modelOptions) || modelOptions.length === 0) {
        return '-'
      }
      return modelOptions.join(' / ')
    },
    getCodexConfig(row) {
      if (!row || row.type !== 'codex-cli' || !row.config) {
        return {}
      }
      try {
        return JSON.parse(row.config) || {}
      } catch (e) {
        return {}
      }
    },
    getCodexWireApi(row) {
      const cfg = this.getCodexConfig(row)
      return cfg.wire_api || 'responses'
    },
    getCodexSupportsWebsocketsText(row) {
      const cfg = this.getCodexConfig(row)
      if (!cfg.base_url) {
        return '-'
      }
      return cfg.supports_websockets === false ? '关闭' : '开启'
    },
    // loadChatDetail 加载执行详情并在运行态自动接回 SSE。 // Loads execution detail and reconnects SSE for running records.
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
          this.updateChatListStatus(this.chatDetailId, this.chatDetailStatus)
          const historicalLines = data.lines || []
          const sseLines = this.chatDetailSSELines
          const newSseLines = sseLines.filter(line => !historicalLines.includes(line))
          this.chatDetailSSELines = [...historicalLines, ...newSseLines]
          this.chatDetailMessages = chatParser.parseChatLines(this.chatDetailSSELines, this.chatDetailCliType)
          this.chatDetailMessages.forEach(msg => {
            if (msg.type === 'assistant' && msg.thinking) {
              msg._thinkingCollapsed = true
            }
            if (msg.type === 'assistant_thinking') {
              msg._thinkingCollapsed = true
            }
          })
          this.$nextTick(() => { this.scrollAgentChatToBottom(true) })
          if (this.chatDetailStatus === 'running' && this._sseChatId !== this.chatDetailId) {
            this.connectChatStream(this.chatDetailId)
          }
        }
      })
    },
    // connectChatStream 创建专用 EventSource 连接以实时接收对话输出。 // Creates a dedicated EventSource for execution output streaming.
    connectChatStream(chatId, continuePrompt, isNewChat) {
      if (this._sseChatId === chatId && this._chatEventSource && this._chatEventSource.readyState !== EventSource.CLOSED) return
      if (this._chatEventSource) {
        this._chatEventSource.close()
        this._chatEventSource = null
      }
      this._sseChatId = chatId
      this.chatDetailSSERegistered = true
      this._thinkingStreamStartTime = 0
      this._sseParseState = this.chatDetailCliType === 'codex'
        ? { currentItems: new Map(), pendingPatches: [] }
        : { currentMessage: null, toolUseMap: new Map(), pendingPatches: [] }
      this._sseLineBuffer = []
      if (this._sseBatchTimer) { clearTimeout(this._sseBatchTimer); this._sseBatchTimer = null }
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
            this.loadAgentChatCounts()
            this.$nextTick(() => { this.scrollAgentChatToBottom() })
            return
          }
        } catch (e) {
          // SSE 解析失败时跳过该行。 // Skip malformed SSE lines.
        }
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
        if (this._initialSseRetryCount < 1 && this.chatDetailSSELines.length === 0 && this.chatDetailStatus === 'running') {
          this._initialSseRetryCount++
          this.connectChatStream(this.chatDetailId, null, true)
          return
        }
        this.loadChatDetail()
        this.loadAgentChatCounts()
      }
    },
    _flushSseBatch() {
      if (this._sseBatchTimer) {
        clearTimeout(this._sseBatchTimer)
        this._sseBatchTimer = null
      }
      const newLines = this._sseLineBuffer.splice(0)
      if (newLines.length === 0) return
      for (const line of newLines) {
        this.chatDetailSSELines.push(line)
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
          this.scrollAgentChatToBottom()
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
    toggleThinkingCollapse(msg) {
      msg._thinkingCollapsed = !msg._thinkingCollapsed
      msg._thinkingManuallyToggled = true
    },
    isCurrentThinking(msg) {
      const timing = msg && msg._thinkingTiming ? msg._thinkingTiming : null
      if (!timing || !timing.startMs || timing.durationMs > 0) return false
      for (let i = this.chatDetailMessages.length - 1; i >= 0; i--) {
        const item = this.chatDetailMessages[i]
        if (item.type === 'assistant' && item.thinking) {
          return item === msg
        }
      }
      return false
    },
    needCollapseBtn(text) {
      return (text || '').split('\n').length > 10
    },
    formatCliType(cliType) {
      if (!cliType) return '提示词'
      return cliType.charAt(0).toUpperCase() + cliType.slice(1)
    },
    isLongText(text, maxBytes) {
      if (!text) return false
      return new TextEncoder().encode(text).length > maxBytes
    },
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
      this.chatContinueInput = ''
      taskProgressStore.reset()
      this.chatDetailId = 0
      this.agentChatDetailId = 0
    },
    updateChatListStatus(chatId, status) {
      const item = this.agentChatHistoryList.find(row => row.id === chatId)
      if (item) item.status = status
      this.syncBackgroundChatStreams(this.agentChatHistoryList, this.agentChatDetailId || this.chatDetailId)
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
      const currentItem = this.agentChatHistoryList.find(item => Number(item.id || 0) === normalizedChatId)
      const sseHost = baseUtils.GetSseApiHost()
      const url = sseHost + '/api/task/workflow/chat/stream?chat_id=' + normalizedChatId + '&token=' + encodeURIComponent(baseUtils.GetSafeToken())
      const es = new EventSource(url)
      const state = {
        es,
        // 中文注释：背景 SSE 自己维护 line_count，避免未选中执行历史只能等列表刷新。
        // English comment: Background SSE keeps line_count locally so non-selected execution rows do not have to wait for a list refresh.
        lineCount: Number(currentItem?.line_count || 0),
      }
      this._backgroundChatEventSources = {
        ...this._backgroundChatEventSources,
        [normalizedChatId]: state,
      }
      es.onmessage = (event) => {
        const line = event.data
        if (!line) return
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
            this.loadAgentChatHistoryListSilently()
            this.loadAgentChatCounts()
          }
        } catch (e) {
          // 中文注释：普通增量消息只更新计数；状态由 completed 终态事件或后续静默刷新修正。
          // English comment: Non-terminal background messages only advance the counter; terminal state still comes from completed events or silent refresh fallback.
        }
      }
      es.onerror = () => {
        this.stopBackgroundChatStream(normalizedChatId)
      }
    },
    updateBackgroundChatListItem(chatId, patch) {
      const normalizedChatId = Number(chatId || 0)
      if (normalizedChatId <= 0 || !patch) return
      const item = this.agentChatHistoryList.find(row => Number(row.id || 0) === normalizedChatId)
      if (!item) return
      Object.keys(patch).forEach((key) => {
        if (patch[key] !== undefined) {
          item[key] = patch[key]
        }
      })
      if (normalizedChatId === Number(this.chatDetailId || 0) && patch.status) {
        this.chatDetailStatus = patch.status
      }
      this.agentChatHistoryList = this.agentChatHistoryList.slice()
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
    loadAgentChatHistoryListSilently() {
      if (!this.agentChatHistoryVisible || this.agentChatHistoryLoading || this.agentChatHistoryCliId <= 0) return
      agentCliApi.AgentChatListByAgentCli(this.agentChatHistoryCliId, (response) => {
        if (!(response && response.ErrCode === 0 && response.Data)) return
        this.agentChatHistoryList = Array.isArray(response.Data.list) ? response.Data.list : []
        this.syncBackgroundChatStreams(this.agentChatHistoryList, this.agentChatDetailId || this.chatDetailId)
        this.agentChatCounts = {
          ...this.agentChatCounts,
          [this.agentChatHistoryCliId]: {
            running: this.agentChatHistoryList.filter(item => item.status === 'running').length,
            interrupted: this.agentChatHistoryList.filter(item => item.status === 'interrupted').length,
            total: this.agentChatHistoryList.length,
            unread: this.agentChatHistoryList.filter(item => item.is_read === false && item.status !== 'running').length,
          },
        }
        if (this.chatDetailId > 0) {
          const current = this.agentChatHistoryList.find(item => item.id === this.chatDetailId)
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
    formatDurationDisplay(durationMs) {
      const ms = Number(durationMs || 0)
      if (ms <= 0) return ''
      const totalSeconds = Math.floor(ms / 1000)
      const minutes = Math.floor(totalSeconds / 60)
      const seconds = totalSeconds % 60
      if (minutes > 0) return minutes + 'm' + seconds + 's'
      return seconds + 's'
    },
    runtimeDurationText(item) {
      if (!item || !item.created_at) return ''
      const created = new Date(item.created_at.replace(/-/g, '/'))
      if (isNaN(created.getTime())) return ''
      const ms = Date.now() - created.getTime()
      return this.formatDurationDisplay(ms)
    },
    getItemMsgCount(item) {
      if (item.status === 'running' && this._sseChatId > 0 && item.id === this._sseChatId) {
        return this.chatDetailSSELines.length
      }
      return item.line_count || 0
    },
    formatCreatedAt(createdAt) {
      if (!createdAt) return ''
      const d = new Date(createdAt.replace(/-/g, '/'))
      if (isNaN(d.getTime())) return ''
      const pad = (n) => String(n).padStart(2, '0')
      return d.getFullYear() + '/' + pad(d.getMonth() + 1) + '/' + pad(d.getDate()) + ' ' +
        pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds())
    },
    renderMarkdown(text) {
      if (!text) return ''
      return md.render(text)
    },
    // formatNum 格式化数字，统一结果卡片中的 Token 展示。 // formatNum renders grouped numbers for result cards.
    formatNum(num) {
      if (num == null) return '0'
      return Number(num).toLocaleString()
    },
    // isChatContinueDisabled 统一发送区按钮可用状态，确保“发送/新对话”禁用规则一致。 // Keeps send and new-chat buttons under the same disabled rule.
    isChatContinueDisabled() {
      return this.chatContinueLoading || !String(this.chatContinueInput || '').trim()
    },
    // stopReasonLabel 将停止原因映射为中文文案。 // stopReasonLabel maps structured stop reasons into readable labels.
    stopReasonLabel(reason) {
      const map = {
        end_turn: '正常结束',
        stop_sequence: '停止序列',
        max_tokens: '达到上限',
        tool_use: '工具调用',
      }
      return map[reason] || reason
    },
    // continueChat 继续当前 Agent CLI 历史对话。 // continueChat resumes the selected standalone AgentCli chat with a new user message.
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
          this.$message.error(res.ErrMsg || '发送失败')
        }
      })
    },
    // startNewChatFromHistory 在执行历史中直接创建一个全新独立对话，并切换左侧焦点。 // Creates a brand-new standalone AgentCli chat from the history input and focuses it.
    startNewChatFromHistory() {
      const prompt = this.chatContinueInput.trim()
      if (!prompt || this.chatContinueLoading) return
      const agentCliId = Number(this.agentChatHistoryCliId || 0)
      if (agentCliId <= 0) {
        this.$message.warning('Agent 实例不存在')
        return
      }
      this.chatContinueLoading = true
      agentCliApi.AgentChatSend({
        agent_cli_id: agentCliId,
        prompt: prompt,
        prompt_type: 'agent_cli_manual',
        local_dir: String(this.chatDetailLocalDir || '').trim(),
        cli_type: this.chatDetailCliType || 'claude',
        model_name: String(this.chatDetailModelName || '').trim(),
        thinking_intensity: String(this.chatDetailThinkingIntensity || '高').trim() || '高',
      }, (response) => {
        this.chatContinueLoading = false
        if (!(response && response.ErrCode === 0 && response.Data)) {
          this.$message.error(response?.ErrMsg || '新建对话失败')
          return
        }
        const chatId = response.Data.chat_id
        this.chatContinueInput = ''
        this.$message.success('已创建新对话')
        this.chatDetailId = chatId
        this.agentChatDetailId = chatId
        this.chatDetailStatus = 'running'
        this.chatDetailSSELines = []
        this.chatDetailMessages = []
        this.chatDetailLastUsageSummary = null
        taskProgressStore.reset()
        this._initialSseRetryCount = 0
        this.markAgentChatRunningLocally(agentCliId, chatId, {
          prompt,
          localDir: String(this.chatDetailLocalDir || '').trim(),
          modelName: String(this.chatDetailModelName || '').trim(),
          thinkingIntensity: String(this.chatDetailThinkingIntensity || '高').trim() || '高',
          cliType: this.chatDetailCliType || 'claude',
        })
        this.connectChatStream(chatId, null, true)
        this.loadChatDetail()
        this.loadAgentChatCounts()
        this.openAgentChatHistory({ id: agentCliId, name: this.agentChatHistoryTitle }, chatId)
      })
    },
    // stopChat 停止当前 Agent CLI 历史对话。 // stopChat interrupts the selected standalone AgentCli chat immediately on both UI and backend.
    stopChat() {
      if (this._chatEventSource) {
        this._chatEventSource.close()
        this._chatEventSource = null
      }
      this._sseChatId = 0
      this.chatDetailSSERegistered = false
      taskWorkflowApi.TaskWorkflowChatStop(this.chatDetailId, (res) => {
        if (res.ErrCode !== 0) {
          this.$message.error(res.ErrMsg || '停止失败')
        }
      })
      this.chatDetailStatus = 'interrupted'
      this.updateChatListStatus(this.chatDetailId, 'interrupted')
    },
    // ---- Webhook 配置相关 ----
    loadWebhookOptions() {
      agentCliApi.WebhookConfigList((response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.webhookOptions = response.Data.list || []
        }
      })
    },
    loadWebhookList() {
      this.webhookLoading = true
      agentCliApi.WebhookConfigList((response) => {
        this.webhookLoading = false
        if (response && response.ErrCode === 0 && response.Data) {
          this.webhookList = response.Data.list || []
          this.webhookOptions = this.webhookList
        }
      })
    },
    webhookTypeLabel(type) {
      const map = { dingtalk: '钉钉', feishu: '飞书', wecom: '企微' }
      return map[type] || type
    },
    webhookUrlPlaceholderByType(type) {
      const map = {
        dingtalk: 'https://oapi.dingtalk.com/robot/send?access_token=xxx',
        feishu: 'https://open.feishu.cn/open-apis/bot/v2/hook/xxx',
        wecom: 'https://qyapi.weixin.qq.com/cgi-bin/webhook/send?key=xxx',
      }
      return map[type] || '请输入 Webhook 地址'
    },
    webhookSecretPlaceholderByType(type) {
      const map = {
        dingtalk: 'SEC... (可选)',
        feishu: '签名校验密钥 (可选)',
        wecom: '暂无扩展字段',
      }
      return map[type] || '请输入签名密钥'
    },
    openWebhookForm(row) {
      if (row) {
        this.webhookForm = {
          id: row.id,
          name: row.name,
          type: row.type,
          webhook_url: row.webhook_url,
          secret: row.secret,
          test_single_url: row.test_single_url || '',
        }
      } else {
        this.webhookForm = { id: 0, name: '', type: 'dingtalk', webhook_url: '', secret: '', test_single_url: '' }
      }
      this.webhookFormVisible = true
    },
    saveWebhook() {
      if (!this.webhookForm.name.trim()) {
        this.$message.warning('请输入配置名称')
        return
      }
      if (!this.webhookForm.webhook_url.trim()) {
        this.$message.warning('请输入 Webhook 地址')
        return
      }
      this.webhookSaving = true
      agentCliApi.WebhookConfigSave(this.webhookForm, (response) => {
        this.webhookSaving = false
        if (response && response.ErrCode === 0) {
          this.$message.success('保存成功')
          this.webhookFormVisible = false
          this.loadWebhookList()
        } else {
          this.$message.error(response?.ErrMsg || '保存失败')
        }
      })
    },
    testWebhook(row) {
      this.sendWebhookTest(row, '')
    },
    testWebhookWithLink(row) {
      const singleURL = String(row?.test_single_url || '').trim()
      if (!singleURL) {
        this.$message.warning('请输入测试链接')
        return
      }
      this.sendWebhookTest(row, singleURL)
    },
    sendWebhookTest(row, singleURL = '') {
      const payload = {
        name: String(row?.name || '').trim(),
        type: String(row?.type || 'dingtalk').trim(),
        webhook_url: String(row?.webhook_url || '').trim(),
        secret: String(row?.secret || '').trim(),
        single_url: String(singleURL || '').trim(),
      }
      if (!payload.name) {
        this.$message.warning('请输入配置名称')
        return
      }
      if (!payload.webhook_url) {
        this.$message.warning('请输入 Webhook 地址')
        return
      }
      const isFormTarget = row === this.webhookForm
      const isLinkTest = !!payload.single_url
      if (isFormTarget) {
        if (isLinkTest) {
          this.webhookLinkTesting = true
        } else {
          this.webhookTesting = true
        }
        } else {
          this.$set(row, '_testingWebhook', true)
        }
      agentCliApi.WebhookConfigTest(payload, (response) => {
        if (isFormTarget) {
          if (isLinkTest) {
            this.webhookLinkTesting = false
          } else {
            this.webhookTesting = false
          }
          } else {
            this.$set(row, '_testingWebhook', false)
          }
        if (response && response.ErrCode === 0) {
          this.$message.success(response?.ErrMsg || '测试发送成功')
        } else {
          this.$message.error(response?.ErrMsg || '测试发送失败')
        }
      })
    },
    deleteWebhook(row) {
      this.$confirm(`确定要删除 "${row.name}" 吗？`, '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        agentCliApi.WebhookConfigDelete(row.id, (response) => {
          if (response && response.ErrCode === 0) {
            this.$message.success('删除成功')
            this.loadWebhookList()
          } else {
            this.$message.error(response?.ErrMsg || '删除失败')
          }
        })
      }).catch(() => {})
    },
    updateWebhookConfig(item) {
      const previousWebhookConfigId = item._lastWebhookConfigId ?? item.webhook_config_id
      item._savingWebhook = true
      const data = {
        id: item.id,
        name: item.name,
        type: item.type,
        settings_path: item.settings_path || '',
        enabled: item.displayed_enabled ? 1 : 0,
        webhook_config_id: parseInt(item.webhook_config_id) || 0,
      }
      if (item.config) data.config = item.config
      agentCliApi.AgentCliSave(data, (response) => {
        item._savingWebhook = false
        if (response && response.ErrCode === 0) {
          item._lastWebhookConfigId = item.webhook_config_id
          item.webhook_config_name = this.getWebhookNameById(item.webhook_config_id)
          this.$message.success('通知配置已更新')
        } else {
          item.webhook_config_id = previousWebhookConfigId || ''
          this.$message.error(response?.ErrMsg || '更新失败')
        }
      })
    },
    updateGroupIds(item) {
      const previousGroupIds = Array.isArray(item._lastGroupIds) ? [...item._lastGroupIds] : []
      item._savingGroups = true
      agentCliApi.AgentCliGroupRelSave({
        agent_cli_id: item.id,
        group_ids: Array.isArray(item.group_ids) ? item.group_ids : [],
      }, (response) => {
        item._savingGroups = false
        if (response && response.ErrCode === 0) {
          item._lastGroupIds = Array.isArray(item.group_ids) ? [...item.group_ids] : []
          this._syncLocalGroupIds(item.id, item.group_ids)
          this.$message.success('分组已更新')
        } else {
          item.group_ids = previousGroupIds
          this.$message.error(response?.ErrMsg || '分组更新失败')
        }
      })
    },
    getWebhookNameById(webhookConfigId) {
      const targetId = parseInt(webhookConfigId) || 0
      if (!targetId) {
        return ''
      }
      const found = this.webhookOptions.find(item => Number(item.id) === targetId)
      return found ? found.name : ''
    },
    copyText(text) {
      const content = String(text || '').trim()
      if (!content) {
        return
      }
      const copyIndex = copy.SetCopyContent(content)
      copy.handleCopy(copyIndex)
    },
    // sortAgentCliList 将启用实例排在前面，同状态下按 ID 倒序，减少列表跳动并保留最近项优先。 // sortAgentCliList keeps enabled items first and orders same-state rows by descending ID.
    sortAgentCliList(items) {
      const sortedList = Array.isArray(items) ? [...items] : []
      sortedList.sort((firstItem, secondItem) => {
        const firstEnabledWeight = firstItem?.displayed_enabled ? AGENT_CLI_ENABLED_SORT_TRUE : AGENT_CLI_ENABLED_SORT_FALSE
        const secondEnabledWeight = secondItem?.displayed_enabled ? AGENT_CLI_ENABLED_SORT_TRUE : AGENT_CLI_ENABLED_SORT_FALSE
        if (firstEnabledWeight !== secondEnabledWeight) {
          return secondEnabledWeight - firstEnabledWeight
        }
        return (secondItem?.id || 0) - (firstItem?.id || 0)
      })
      return sortedList
    },
    // ==================== 分组相关方法 ====================
    // loadGroupList 加载分组列表，用于筛选栏和编辑弹窗的多选下拉。 // Loads group list for both filter bar and edit dialog multi-select.
    loadGroupList() {
      agentCliApi.AgentCliGroupList((response) => {
        if (response && response.ErrCode === 0 && response.Data) {
          this.groupList = response.Data.list || []
        }
      })
    },
    // selectGroup 选中分组筛选，记住到 localStorage。 // Selects a group filter and persists it to localStorage.
    selectGroup(groupId) {
      this.selectedGroupId = groupId
      localStorage.setItem(AGENT_CLI_GROUP_CACHE_KEY, String(groupId))
    },
    // getGroupName 根据分组 ID 获取名称。 // Returns group name by ID.
    getGroupName(groupId) {
      const g = this.groupList.find(item => item.id === groupId)
      return g ? g.name : ''
    },
    // openGroupDialog 打开分组管理弹窗。 // Opens the group management dialog.
    openGroupDialog() {
      this.groupDialogVisible = true
      this.groupFormVisible = false
      this.loadGroupDialogList()
    },
    // loadGroupDialogList 加载分组管理弹窗中的列表。 // Loads the group list inside the management dialog.
    loadGroupDialogList() {
      this.groupDialogLoading = true
      agentCliApi.AgentCliGroupList((response) => {
        this.groupDialogLoading = false
        if (response && response.ErrCode === 0 && response.Data) {
          this.groupDialogList = response.Data.list || []
        }
      })
    },
    // openGroupForm 打开分组新增/编辑内嵌表单。 // Opens the inline add/edit form for a group.
    openGroupForm(row) {
      if (row) {
        this.groupForm = { id: row.id, name: row.name, sort_order: row.sort_order || 0 }
      } else {
        this.groupForm = { id: 0, name: '', sort_order: 0 }
      }
      this.groupFormVisible = true
    },
    // saveGroup 保存分组（新增或编辑）。 // Saves a group (create or update).
    saveGroup() {
      if (!this.groupForm.name.trim()) {
        this.$message.warning('分组名称不能为空')
        return
      }
      this.groupSaving = true
      agentCliApi.AgentCliGroupSave({
        id: this.groupForm.id || undefined,
        name: this.groupForm.name.trim(),
        sort_order: this.groupForm.sort_order || 0,
      }, (response) => {
        this.groupSaving = false
        if (response && response.ErrCode === 0) {
          this.$message.success('保存成功')
          this.groupFormVisible = false
          this.loadGroupDialogList()
          this.loadGroupList()
          this.loadList()
        } else {
          this.$message.error(response?.ErrMsg || '保存失败')
        }
      })
    },
    // deleteGroup 删除分组。 // Deletes a group.
    deleteGroup(row) {
      this.$confirm(`确定要删除分组 "${row.name}" 吗？`, '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        agentCliApi.AgentCliGroupDelete(row.id, (response) => {
          if (response && response.ErrCode === 0) {
            this.$message.success('删除成功')
            this.loadGroupDialogList()
            this.loadGroupList()
            this.loadList()
            // 如果删除的是当前选中的分组，重置为全部
            if (this.selectedGroupId === row.id) {
              this.selectGroup(0)
            }
          } else {
            this.$message.error(response?.ErrMsg || '删除失败')
          }
        })
      }).catch(() => {})
    },
    // _syncLocalGroupIds 将最新分组关联同步到本地列表，确保筛选和标签立即更新。 // Syncs latest group relations into local list so filters and tags update immediately.
    _syncLocalGroupIds(agentCliId, groupIds) {
      const normalizedGroupIds = Array.isArray(groupIds) ? [...groupIds] : []
      const target = this.list.find(item => Number(item.id) === Number(agentCliId))
      if (target) {
        target.group_ids = normalizedGroupIds
        target._lastGroupIds = [...normalizedGroupIds]
      }
    },
    // _saveGroupRel 保存某个 AgentCli 的分组关联。 // Saves group relations for an Agent CLI and syncs the local row on success.
    _saveGroupRel(agentCliId) {
      if (!agentCliId || agentCliId <= 0) return
      agentCliApi.AgentCliGroupRelSave({
        agent_cli_id: agentCliId,
        group_ids: this.form.group_ids || [],
      }, (response) => {
        if (response && response.ErrCode === 0) {
          this._syncLocalGroupIds(agentCliId, this.form.group_ids)
          return
        }
        this.$message.error(response?.ErrMsg || '分组保存失败')
      })
    },
  },
}
</script>

<style scoped src="@/css/components/agent_cli/AgentCliList.css"></style>
