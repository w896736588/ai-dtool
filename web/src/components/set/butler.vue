<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">AI 管家配置</h3>
      <p class="set-config-desc">管理管家机器人、角色与运行参数，管家进程（dtool-butler）只读这些配置表</p>
    </div>

    <el-tabs v-model="state.activeTab" class="set-config-inner-tabs" @tab-change="HandleInnerTabChange">
      <!-- 机器人配置 -->
      <el-tab-pane label="机器人配置" name="bot">
        <div class="set-config-actions" style="margin-bottom: 10px;">
          <pl-button type="primary" @click="ShowAddBotConfig">新增机器人</pl-button>
        </div>
        <el-alert type="info" :closable="false" style="margin-bottom: 10px;">
          <template #title>
            <span style="font-size: 12px; line-height: 1.8;">
              <b>AppKey / AppSecret</b>：流式机器人必需，用于建立 Stream 长连接收发消息<br/>
              <b>RobotCode</b>：机器人编码，用于通过 Open API 主动发送单聊消息
            </span>
          </template>
        </el-alert>
        <div class="set-config-table-card">
          <el-table :data="state.botConfigList" class="set-config-table" row-key="id">
            <el-table-column prop="id" label="#id" width="70"/>
            <el-table-column prop="name" label="名称" min-width="120"/>
            <el-table-column prop="platform" label="平台" width="100"/>
            <el-table-column prop="app_key" label="AppKey" min-width="160"/>
            <el-table-column prop="robot_code" label="RobotCode" min-width="140"/>
            <el-table-column prop="status" label="启用" width="70">
              <template #default="scope">
                <el-tag size="small" :type="scope.row.status === 1 ? 'success' : 'danger'" effect="light">
                  {{ scope.row.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="conn_status" label="连接" width="90">
              <template #default="scope">
                <el-tag size="small" :type="connStatusTagType(scope.row.conn_status)" effect="light">
                  {{ connStatusText(scope.row.conn_status) }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="280">
              <template #default="scope">
                <div class="set-op-group">
                  <pl-button size="small" type="success" plain @click="ShowEditBotConfig(scope.row, false)">编辑</pl-button>
                  <pl-button size="small" type="primary" plain @click="ShowBotMessageLog(scope.row)">消息日志</pl-button>
                  <el-dropdown trigger="click" @command="(cmd) => HandleBotMoreOp(cmd, scope.row)">
                    <pl-button size="small" type="info" plain>
                      更多操作<el-icon class="el-icon--right"><arrow-down /></el-icon>
                    </pl-button>
                    <template #dropdown>
                      <el-dropdown-menu>
                        <el-dropdown-item command="copy">复制新增</el-dropdown-item>
                        <el-dropdown-item command="test">测试连接</el-dropdown-item>
                        <el-dropdown-item command="delete" divided>删除</el-dropdown-item>
                      </el-dropdown-menu>
                    </template>
                  </el-dropdown>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <!-- 角色配置 -->
      <el-tab-pane label="角色配置" name="role">
        <div class="set-config-actions" style="margin-bottom: 10px;">
          <pl-button type="primary" @click="ShowAddRole">新增角色</pl-button>
        </div>
        <div class="set-config-table-card">
          <el-table :data="state.roleList" class="set-config-table" row-key="id">
            <el-table-column prop="id" label="#id" width="70"/>
            <el-table-column prop="name" label="角色名称" min-width="120"/>
            <el-table-column prop="persona" label="定位" min-width="160"/>
            <el-table-column prop="tone" label="语气" min-width="120"/>
            <el-table-column prop="init_greeting" label="打招呼语" min-width="200"/>
            <el-table-column prop="status" label="状态" width="80">
              <template #default="scope">
                <el-tag size="small" :type="scope.row.status === 1 ? 'success' : 'danger'" effect="light">
                  {{ scope.row.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="200">
              <template #default="scope">
                <div class="set-op-group">
                  <pl-button size="small" type="success" plain @click="ShowEditRole(scope.row, false)">编辑</pl-button>
                  <pl-button size="small" type="success" plain @click="ShowEditRole(scope.row, true)">复制新增</pl-button>
                  <pl-button size="small" type="danger" plain @click="DeleteRole(scope.row)">删除</pl-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <!-- 管家运行参数 -->
      <el-tab-pane label="管家配置" name="config">
        <div class="set-config-actions" style="margin-bottom: 10px;">
          <pl-button type="primary" @click="ShowAddButlerConfig">新增管家</pl-button>
        </div>
        <div class="set-config-table-card">
          <el-table :data="state.configList" class="set-config-table" row-key="id">
            <el-table-column prop="id" label="#id" width="70"/>
            <el-table-column prop="name" label="管家名称" min-width="120"/>
            <el-table-column prop="role_name" label="关联角色" min-width="100"/>
            <el-table-column label="主模型" min-width="140">
              <template #default="scope">
                <div>{{ scope.row.model_name || '-' }}</div>
                <div style="font-size: 12px; color: #909399;">{{ scope.row.model_provider_name || '-' }}</div>
              </template>
            </el-table-column>
            <el-table-column prop="fc_model_name" label="FC模型" min-width="120"/>
            <el-table-column prop="agent_cli_name" label="Agent CLI" min-width="120"/>
            <el-table-column prop="bot_config_name" label="机器人配置" min-width="120"/>
            <el-table-column prop="active_timeout_minutes" label="超时(min)" width="90"/>
            <el-table-column prop="max_loop" label="Loop上限" width="80"/>
            <el-table-column prop="max_history_store" label="历史上限" width="90"/>
            <el-table-column prop="auto_init_on_start" label="启动初始化" width="100">
              <template #default="scope">
                <el-tag size="small" :type="scope.row.auto_init_on_start === 1 ? 'success' : 'info'" effect="light">
                  {{ scope.row.auto_init_on_start === 1 ? '是' : '否' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="tool_call_push_enabled" label="工具推送" width="100">
              <template #default="scope">
                <el-tag size="small" :type="scope.row.tool_call_push_enabled === 1 ? 'success' : 'info'" effect="light">
                  {{ scope.row.tool_call_push_enabled === 1 ? '开启' : '关闭' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="status" label="状态" width="80">
              <template #default="scope">
                <el-tag size="small" :type="scope.row.status === 1 ? 'success' : 'danger'" effect="light">
                  {{ scope.row.status === 1 ? '启用' : '禁用' }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="270">
              <template #default="scope">
                <div class="set-op-group">
                  <pl-button size="small" type="success" plain @click="ShowEditButlerConfig(scope.row, false)">编辑</pl-button>
                  <pl-button size="small" type="primary" plain @click="ShowConfigLoopLog(scope.row)">Loop日志</pl-button>
                  <pl-button size="small" type="success" plain @click="ShowEditButlerConfig(scope.row, true)">复制新增</pl-button>
                  <pl-button size="small" type="danger" plain @click="DeleteButlerConfig(scope.row)">删除</pl-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>
    </el-tabs>

    <!-- 机器人配置编辑弹窗 -->
    <el-dialog v-model="state.dialogBotConfig" title="编辑机器人配置" width="560">
      <el-form label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="state.editBotConfig.name" autocomplete="off" placeholder="机器人显示名称"/>
        </el-form-item>
        <el-form-item label="平台">
          <el-select v-model="state.editBotConfig.platform" style="width: 100%;">
            <el-option label="钉钉" value="dingtalk"/>
            <el-option label="飞书" value="feishu"/>
            <el-option label="企微" value="wecom"/>
          </el-select>
        </el-form-item>
        <el-form-item label="AppKey">
          <el-input v-model="state.editBotConfig.app_key" autocomplete="off" placeholder="钉钉应用 AppKey"/>
        </el-form-item>
        <el-form-item label="AppSecret">
          <el-input v-model="state.editBotConfig.app_secret" type="password" show-password autocomplete="off" placeholder="留空则保留原值；需修改请重新输入"/>
        </el-form-item>
        <el-form-item label="RobotCode">
          <el-input v-model="state.editBotConfig.robot_code" autocomplete="off" placeholder="机器人 robotCode"/>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="state.editBotConfig.status" style="width: 100%;">
            <el-option label="启用" :value="1"/>
            <el-option label="禁用" :value="0"/>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogBotConfig = false">取消</pl-button>
          <pl-button type="primary" @click="SaveBotConfig">保存</pl-button>
        </div>
      </template>
    </el-dialog>

    <!-- 角色配置编辑弹窗 -->
    <el-dialog v-model="state.dialogRole" title="编辑角色配置" width="560">
      <el-form label-width="100px">
        <el-form-item label="角色名称">
          <el-input v-model="state.editRole.name" autocomplete="off" placeholder="如：技术管家"/>
        </el-form-item>
        <el-form-item label="定位（Persona）">
          <el-input v-model="state.editRole.persona" autocomplete="off" placeholder="如：严谨的技术管家"/>
        </el-form-item>
        <el-form-item label="语气（Tone）">
          <el-input v-model="state.editRole.tone" autocomplete="off" placeholder="如：简洁专业"/>
        </el-form-item>
        <el-form-item label="System Prompt">
          <el-input v-model="state.editRole.system_prompt" type="textarea" :rows="4" autocomplete="off" placeholder="完整 system prompt（优先使用此字段，为空则用 persona+tone 组合）"/>
        </el-form-item>
        <el-form-item label="打招呼语">
          <el-input v-model="state.editRole.init_greeting" type="textarea" :rows="2" autocomplete="off" placeholder="管家启动时自动发送的消息"/>
        </el-form-item>
        <el-form-item label="状态">
          <el-select v-model="state.editRole.status" style="width: 100%;">
            <el-option label="启用" :value="1"/>
            <el-option label="禁用" :value="0"/>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogRole = false">取消</pl-button>
          <pl-button type="primary" @click="SaveRole">保存</pl-button>
        </div>
      </template>
    </el-dialog>

    <!-- 管家运行参数编辑弹窗 -->
    <el-dialog v-model="state.dialogConfig" title="编辑管家运行参数" width="750">
      <el-form label-width="150px">
        <el-form-item>
          <template #label>
            管家名称
            <el-tooltip content="管家配置的显示名称，用于标识和区分不同的管家实例" placement="top">
              <el-icon class="form-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </template>
          <el-input v-model="state.editConfig.name" autocomplete="off" placeholder="如：默认管家"/>
        </el-form-item>
        <el-form-item>
          <template #label>
            关联角色
            <el-tooltip content="选择管家使用的角色配置，包含定位、语气、打招呼语等设定" placement="top">
              <el-icon class="form-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </template>
          <el-select v-model="state.editConfig.role_id" style="width: 100%;" placeholder="请选择角色">
            <el-option label="不关联" :value="0"/>
            <template v-for="(role, idx) in state.roleList" :key="idx">
              <el-option :label="role.name" :value="role.id"/>
            </template>
          </el-select>
        </el-form-item>
        <el-form-item>
          <template #label>
            主模型
            <el-tooltip content="AI 对话使用的主模型，处理意图分析、任务路由和最终回复" placement="top">
              <el-icon class="form-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </template>
          <el-select v-model="state.editConfig.model_id" style="width: 100%;" placeholder="请选择模型" filterable>
            <el-option label="不指定" :value="0"/>
            <template v-for="(model, idx) in state.aiModelList" :key="idx">
              <el-option :label="model.name + ' (' + model.provider_name + ')'" :value="model.id"/>
            </template>
          </el-select>
        </el-form-item>
        <el-form-item>
          <template #label>
            FC 模型
            <el-tooltip content="Function Calling 专用模型，用于工具调用循环。为 0 时回落使用主模型" placement="top">
              <el-icon class="form-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </template>
          <el-select v-model="state.editConfig.fc_model_id" style="width: 100%;" placeholder="请选择 FC 模型" filterable>
            <el-option label="不指定（回落主模型）" :value="0"/>
            <template v-for="(model, idx) in state.aiModelList" :key="idx">
              <el-option :label="model.name + ' (' + model.provider_name + ')'" :value="model.id"/>
            </template>
          </el-select>
        </el-form-item>
        <el-form-item>
          <template #label>
            Agent CLI
            <el-tooltip content="复杂任务交由 Agent CLI 执行，简单任务走 FC 循环。为空时所有任务均走 FC" placement="top">
              <el-icon class="form-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </template>
          <el-select v-model="state.editConfig.agent_cli_id" style="width: 100%;" placeholder="请选择 Agent CLI">
            <el-option label="不指定（始终走 FC）" :value="0"/>
            <template v-for="(cli, idx) in state.agentCliList" :key="idx">
              <el-option :label="cli.name" :value="cli.id"/>
            </template>
          </el-select>
        </el-form-item>
        <el-form-item>
          <template #label>
            机器人配置
            <el-tooltip content="关联的钉钉机器人，管家通过该机器人接收消息和发送回复" placement="top">
              <el-icon class="form-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </template>
          <el-select v-model="state.editConfig.bot_config_id" style="width: 100%;" placeholder="请选择机器人配置">
            <el-option label="不关联" :value="0"/>
            <template v-for="(bot, idx) in state.botConfigList" :key="idx">
              <el-option :label="bot.name + ' (' + bot.platform + ')'" :value="bot.id"/>
            </template>
          </el-select>
        </el-form-item>
        <el-form-item>
          <template #label>
            激活超时(min)
            <el-tooltip content="会话无操作超过此时间后进入休眠状态，下次发消息时重新激活并发送打招呼语" placement="top">
              <el-icon class="form-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </template>
          <el-input-number v-model="state.editConfig.active_timeout_minutes" :min="5" :max="120" :step="5"/>
        </el-form-item>
        <el-form-item>
          <template #label>
            Loop次数上限
            <el-tooltip content="Function Calling 工具调用循环的最大迭代次数，防止 AI 陷入无限循环" placement="top">
              <el-icon class="form-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </template>
          <el-input-number v-model="state.editConfig.max_loop" :min="3" :max="50" :step="1"/>
        </el-form-item>
        <el-form-item>
          <template #label>
            历史上限
            <el-tooltip content="控制 AI 上下文窗口大小。超过上限自动删除最旧消息，发送 /clean 可手动清空全部" placement="top">
              <el-icon class="form-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </template>
          <el-input-number v-model="state.editConfig.max_history_store" :min="10" :max="500" :step="10"/>
        </el-form-item>
        <el-form-item>
          <template #label>
            索引文档路径
            <el-tooltip content="管家技能索引文档（apis.md / scripts.md / capabilities.md）的存储目录，留空使用默认路径" placement="top">
              <el-icon class="form-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </template>
          <el-input v-model="state.editConfig.index_doc_path" autocomplete="off" placeholder="留空则用默认 {memoryDbPath}/butler/index/"/>
        </el-form-item>
        <el-form-item>
          <template #label>
            启动自动初始化
            <el-tooltip content="管家启动时自动扫描 skills/ 目录生成索引文档，索引已存在则跳过" placement="top">
              <el-icon class="form-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </template>
          <el-select v-model="state.editConfig.auto_init_on_start" style="width: 100%;">
            <el-option label="是" :value="1"/>
            <el-option label="否" :value="0"/>
          </el-select>
        </el-form-item>
        <el-form-item>
          <template #label>
            工具调用推送
            <el-tooltip content='开启后，工具执行时会实时推送进度消息（如「正在查询Git仓库...」）到钉钉机器人' placement="top">
              <el-icon class="form-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </template>
          <el-select v-model="state.editConfig.tool_call_push_enabled" style="width: 100%;">
            <el-option label="开启" :value="1"/>
            <el-option label="关闭" :value="0"/>
          </el-select>
        </el-form-item>
        <el-form-item>
          <template #label>
            状态
            <el-tooltip content="启用后管家开始工作，接收并处理消息；禁用后管家停止响应" placement="top">
              <el-icon class="form-help-icon"><QuestionFilled /></el-icon>
            </el-tooltip>
          </template>
          <el-select v-model="state.editConfig.status" style="width: 100%;">
            <el-option label="启用" :value="1"/>
            <el-option label="禁用" :value="0"/>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogConfig = false">取消</pl-button>
          <pl-button type="primary" @click="SaveButlerConfig">保存</pl-button>
        </div>
      </template>
    </el-dialog>

    <!-- 消息日志弹窗 -->
    <el-dialog v-model="state.dialogMessageLog" :title="'消息日志 - ' + state.messageLogBotName" width="800" @closed="OnMessageLogClosed">
      <el-table :data="state.messageList" class="set-config-table" row-key="id" size="small" max-height="450">
        <el-table-column prop="id" label="#id" width="60"/>
        <el-table-column prop="role" label="角色" width="80">
          <template #default="scope">
            <el-tag size="small" :type="scope.row.role === 'user' ? 'warning' : 'success'" effect="plain">
              {{ scope.row.role === 'user' ? '用户' : '管家' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="content" label="内容" min-width="300" show-overflow-tooltip/>
        <el-table-column prop="session_id" label="会话ID" width="140" show-overflow-tooltip/>
        <el-table-column prop="topic" label="话题" width="120" show-overflow-tooltip/>
        <el-table-column prop="created_at" label="时间" width="150">
          <template #default="scope">
            {{ FormatMsgTime(scope.row.created_at) }}
          </template>
        </el-table-column>
      </el-table>
      <div style="margin-top: 12px; text-align: right;">
        <el-pagination
          v-model:current-page="state.messagePage"
          v-model:page-size="state.messagePageSize"
          :total="state.messageTotal"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="LoadMessageLog"
          @current-change="LoadMessageLog"
        />
      </div>
    </el-dialog>

    <!-- Loop日志弹窗 -->
    <el-dialog v-model="state.dialogLoopLog" :title="'Loop日志 - ' + state.loopLogBotName" width="900" @closed="OnLoopLogClosed">
      <el-table :data="state.loopLogList" class="set-config-table" row-key="id" size="small" max-height="450">
        <el-table-column prop="id" label="#id" width="60"/>
        <el-table-column prop="title" label="用户任务" min-width="160" show-overflow-tooltip/>
        <el-table-column prop="executor" label="执行器" width="90">
          <template #default="scope">
            <el-tag size="small" :type="scope.row.executor === 'agent_cli' ? 'warning' : 'primary'" effect="plain">
              {{ scope.row.executor === 'agent_cli' ? 'Agent CLI' : 'FC' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="plan" label="工具调用链" min-width="200" show-overflow-tooltip/>
        <el-table-column prop="status" label="状态" width="80">
          <template #default="scope">
            <el-tag size="small" :type="scope.row.status === 'done' ? 'success' : 'danger'" effect="light">
              {{ scope.row.status === 'done' ? '成功' : '失败' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="result" label="AI 回复" min-width="220" show-overflow-tooltip/>
        <el-table-column prop="created_at" label="时间" width="150">
          <template #default="scope">
            {{ FormatMsgTime(scope.row.created_at) }}
          </template>
        </el-table-column>
      </el-table>
      <div style="margin-top: 12px; text-align: right;">
        <el-pagination
          v-model:current-page="state.loopLogPage"
          v-model:page-size="state.loopLogPageSize"
          :total="state.loopLogTotal"
          :page-sizes="[20, 50, 100]"
          layout="total, sizes, prev, pager, next"
          @size-change="LoadLoopLog"
          @current-change="LoadLoopLog"
        />
      </div>
    </el-dialog>
  </div>
</template>

<script>
import {defineComponent, getCurrentInstance, reactive} from 'vue'
import {ArrowDown, QuestionFilled} from '@element-plus/icons-vue'
import butlerSet from '@/utils/base/butler_set'
import aiSet from '@/utils/base/ai_set'
import agentCli from '@/utils/base/agent_cli'

export default defineComponent({
  setup() {
    const proxy = getCurrentInstance().proxy

    const state = reactive({
      activeTab: 'bot',
      // 机器人配置
      botConfigList: [],
      dialogBotConfig: false,
      editBotConfig: {
        id: 0,
        platform: 'dingtalk',
        name: '',
        app_key: '',
        app_secret: '',
        robot_code: '',
        status: 1,
      },
      // 角色配置
      roleList: [],
      dialogRole: false,
      editRole: {
        id: 0,
        name: '',
        persona: '',
        tone: '',
        system_prompt: '',
        init_greeting: '',
        status: 1,
      },
      // 管家运行参数
      configList: [],
      dialogConfig: false,
      editConfig: {
        id: 0,
        name: '',
        role_id: 0,
        model_id: 0,
        fc_model_id: 0,
        agent_cli_id: 0,
        bot_config_id: 0,
        active_timeout_minutes: 30,
        max_loop: 10,
        max_history_store: 100,
        index_doc_path: '',
        auto_init_on_start: 1,
        tool_call_push_enabled: 0,
        status: 1,
      },
      // 下拉选择数据
      aiModelList: [],
      agentCliList: [],
      // 消息日志
      dialogMessageLog: false,
      messageLogBotId: 0,
      messageLogBotName: '',
      messageList: [],
      messagePage: 1,
      messagePageSize: 20,
      messageTotal: 0,
      // Loop日志
      dialogLoopLog: false,
      loopLogBotId: 0,
      loopLogBotName: '',
      loopLogList: [],
      loopLogPage: 1,
      loopLogPageSize: 20,
      loopLogTotal: 0,
    })

    // ========== 机器人配置 ==========
    const LoadBotConfigList = function () {
      butlerSet.ButlerBotConfigList(function (response) {
        if (response.ErrCode === 0) {
          state.botConfigList = response.Data || []
        }
      })
    }

    const ShowAddBotConfig = function () {
      state.editBotConfig = {
        id: 0, platform: 'dingtalk', name: '', app_key: '', app_secret: '',
        robot_code: '', status: 1,
      }
      state.dialogBotConfig = true
    }

    const ShowEditBotConfig = function (row, isCopy) {
      state.editBotConfig = {
        id: isCopy ? 0 : row.id,
        platform: row.platform || 'dingtalk',
        name: isCopy ? row.name + '_copy' : row.name,
        app_key: row.app_key || '',
        app_secret: '', // 编辑时 app_secret 已脱敏，不回填，需重新输入
        robot_code: row.robot_code || '',
        status: row.status,
      }
      state.dialogBotConfig = true
    }

    const SaveBotConfig = function () {
      butlerSet.ButlerBotConfigAdd(state.editBotConfig, function (response) {
        if (response.ErrCode === 0) {
          proxy.$message.success('保存成功')
          state.dialogBotConfig = false
          LoadBotConfigList()
        } else {
          proxy.$message.error(response.ErrMsg || '保存失败')
        }
      })
    }

    const DeleteBotConfig = function (row) {
      proxy.$confirm('确认删除机器人配置 "' + row.name + '"？', '提示', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        butlerSet.ButlerBotConfigDelete({id: row.id}, function (response) {
          if (response.ErrCode === 0) {
            proxy.$message.success('删除成功')
            LoadBotConfigList()
          } else {
            proxy.$message.error(response.ErrMsg || '删除失败')
          }
        })
      }).catch(() => {})
    }

    const TestBotConfig = function (row) {
      proxy.$message.info('正在测试连接并发送测试消息...')
      butlerSet.ButlerBotConfigTest({id: row.id}, function (response) {
        if (response.ErrCode === 0) {
          const data = response.Data || {}
          const testResult = data.test_result || ''
          proxy.$message.success(response.ErrMsg + (testResult ? ' | ' + testResult : '') || '测试成功')
          LoadBotConfigList()
        } else {
          proxy.$message.error(response.ErrMsg || '测试失败')
        }
      })
    }

    // 连接状态文本
    const connStatusText = function (status) {
      switch (status) {
        case 1: return '已连接'
        case 2: return '连接失败'
        case 3: return '已断开'
        default: return '未知'
      }
    }

    // 连接状态标签类型
    const connStatusTagType = function (status) {
      switch (status) {
        case 1: return 'success'
        case 2: return 'danger'
        case 3: return 'info'
        default: return 'info'
      }
    }

    // ========== 消息日志 ==========
    const ShowBotMessageLog = function (row) {
      state.messageLogBotId = row.id
      state.messageLogBotName = row.name
      state.messagePage = 1
      state.messagePageSize = 20
      state.messageTotal = 0
      state.messageList = []
      state.dialogMessageLog = true
      LoadMessageLog()
    }

    const LoadMessageLog = function () {
      butlerSet.ButlerMessageList({
        bot_config_id: state.messageLogBotId,
        page: state.messagePage,
        page_size: state.messagePageSize,
      }, function (response) {
        if (response.ErrCode === 0) {
          const data = response.Data || {}
          state.messageList = data.list || []
          state.messageTotal = data.total || 0
        }
      })
    }

    const OnMessageLogClosed = function () {
      state.messageList = []
      state.messageTotal = 0
    }

    // ========== Loop日志 ==========
    const ShowBotLoopLog = function (row) {
      state.loopLogBotId = row.id
      state.loopLogBotName = row.name
      state.loopLogPage = 1
      state.loopLogPageSize = 20
      state.loopLogTotal = 0
      state.loopLogList = []
      state.dialogLoopLog = true
      LoadLoopLog()
    }

    // ShowConfigLoopLog 管家配置列表的 Loop 日志入口，使用关联的 bot_config_id。
    const ShowConfigLoopLog = function (row) {
      const botConfigId = row.bot_config_id || 0
      if (botConfigId === 0) {
        proxy.$message.warning('该管家未关联机器人配置，无法查看 Loop 日志')
        return
      }
      // 从已加载的机器人列表中查找名称
      const botName = state.botConfigList.find(b => b.id === botConfigId)?.name || ''
      state.loopLogBotId = botConfigId
      state.loopLogBotName = row.name + (botName ? ' → ' + botName : '')
      state.loopLogPage = 1
      state.loopLogPageSize = 20
      state.loopLogTotal = 0
      state.loopLogList = []
      state.dialogLoopLog = true
      LoadLoopLog()
    }

    const LoadLoopLog = function () {
      butlerSet.ButlerTaskList({
        bot_config_id: state.loopLogBotId,
        page: state.loopLogPage,
        page_size: state.loopLogPageSize,
      }, function (response) {
        if (response.ErrCode === 0) {
          const data = response.Data || {}
          state.loopLogList = data.list || []
          state.loopLogTotal = data.total || 0
        }
      })
    }

    const OnLoopLogClosed = function () {
      state.loopLogList = []
      state.loopLogTotal = 0
    }

    // ========== 更多操作处理 ==========
    const HandleBotMoreOp = function (cmd, row) {
      switch (cmd) {
        case 'copy':
          ShowEditBotConfig(row, true)
          break
        case 'test':
          TestBotConfig(row)
          break
        case 'delete':
          DeleteBotConfig(row)
          break
      }
    }

    // 格式化消息时间
    const FormatMsgTime = function (timestamp) {
      if (!timestamp) return ''
      const d = new Date(timestamp * 1000)
      const pad = function (n) { return n < 10 ? '0' + n : '' + n }
      return d.getFullYear() + '-' + pad(d.getMonth() + 1) + '-' + pad(d.getDate()) +
        ' ' + pad(d.getHours()) + ':' + pad(d.getMinutes()) + ':' + pad(d.getSeconds())
    }



    // ========== 角色配置 ==========
    const LoadRoleList = function () {
      butlerSet.ButlerRoleList(function (response) {
        if (response.ErrCode === 0) {
          state.roleList = response.Data || []
        }
      })
    }

    const ShowAddRole = function () {
      state.editRole = {
        id: 0, name: '', persona: '', tone: '', system_prompt: '',
        init_greeting: '', status: 1,
      }
      state.dialogRole = true
    }

    const ShowEditRole = function (row, isCopy) {
      state.editRole = {
        id: isCopy ? 0 : row.id,
        name: isCopy ? row.name + '_copy' : row.name,
        persona: row.persona || '',
        tone: row.tone || '',
        system_prompt: row.system_prompt || '',
        init_greeting: row.init_greeting || '',
        status: row.status,
      }
      state.dialogRole = true
    }

    const SaveRole = function () {
      butlerSet.ButlerRoleAdd(state.editRole, function (response) {
        if (response.ErrCode === 0) {
          proxy.$message.success('保存成功')
          state.dialogRole = false
          LoadRoleList()
        } else {
          proxy.$message.error(response.ErrMsg || '保存失败')
        }
      })
    }

    const DeleteRole = function (row) {
      proxy.$confirm('确认删除角色 "' + row.name + '"？', '提示', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        butlerSet.ButlerRoleDelete({id: row.id}, function (response) {
          if (response.ErrCode === 0) {
            proxy.$message.success('删除成功')
            LoadRoleList()
          } else {
            proxy.$message.error(response.ErrMsg || '删除失败')
          }
        })
      }).catch(() => {})
    }

    // ========== 管家运行参数 ==========
    const LoadConfigList = function () {
      // 同时加载角色、模型、Agent CLI 下拉数据
      LoadRoleList()
      LoadAiModelList()
      LoadAgentCliList()
      butlerSet.ButlerConfigList(function (response) {
        if (response.ErrCode === 0) {
          state.configList = response.Data || []
        }
      })
    }

    const LoadAiModelList = function () {
      aiSet.AiModelList({}, function (response) {
        if (response.ErrCode === 0) {
          state.aiModelList = response.Data || []
        }
      })
    }

    const LoadAgentCliList = function () {
      agentCli.AgentCliList(function (response) {
        if (response.ErrCode === 0) {
          state.agentCliList = response.Data || []
        }
      })
    }

    const ShowAddButlerConfig = function () {
      state.editConfig = {
        id: 0, name: '', role_id: 0, model_id: 0, fc_model_id: 0,
        agent_cli_id: 0, bot_config_id: 0, active_timeout_minutes: 30,
        max_loop: 10, max_history_store: 100,
        index_doc_path: '', auto_init_on_start: 1, tool_call_push_enabled: 0, status: 1,
      }
      state.dialogConfig = true
    }

    const ShowEditButlerConfig = function (row, isCopy) {
      state.editConfig = {
        id: isCopy ? 0 : row.id,
        name: isCopy ? row.name + '_copy' : row.name,
        role_id: row.role_id || 0,
        model_id: row.model_id || 0,
        fc_model_id: row.fc_model_id || 0,
        agent_cli_id: row.agent_cli_id || 0,
        bot_config_id: row.bot_config_id || 0,
        active_timeout_minutes: row.active_timeout_minutes || 30,
        max_loop: row.max_loop || 10,
        max_history_store: row.max_history_store || 100,
        index_doc_path: row.index_doc_path || '',
        auto_init_on_start: row.auto_init_on_start || 1,
        tool_call_push_enabled: row.tool_call_push_enabled != null ? row.tool_call_push_enabled : 0,
        status: row.status,
      }
      state.dialogConfig = true
    }

    const SaveButlerConfig = function () {
      butlerSet.ButlerConfigAdd(state.editConfig, function (response) {
        if (response.ErrCode === 0) {
          proxy.$message.success('保存成功')
          state.dialogConfig = false
          LoadConfigList()
        } else {
          proxy.$message.error(response.ErrMsg || '保存失败')
        }
      })
    }

    const DeleteButlerConfig = function (row) {
      proxy.$confirm('确认删除管家配置 "' + row.name + '"？', '提示', {
        confirmButtonText: '删除',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(() => {
        butlerSet.ButlerConfigDelete({id: row.id}, function (response) {
          if (response.ErrCode === 0) {
            proxy.$message.success('删除成功')
            LoadConfigList()
          } else {
            proxy.$message.error(response.ErrMsg || '删除失败')
          }
        })
      }).catch(() => {})
    }

    // ========== Tab 切换 ==========
    const HandleInnerTabChange = function (tabName) {
      switch (tabName) {
        case 'bot':
          LoadBotConfigList()
          break
        case 'role':
          LoadRoleList()
          break
        case 'config':
          LoadConfigList()
          break
      }
    }

    // 暴露给 Set.vue 的按需加载方法
    const LoadData = function () {
      LoadBotConfigList()
    }

    return {
      state,
      LoadData,
      LoadBotConfigList,
      LoadRoleList,
      LoadConfigList,
      ShowAddBotConfig, ShowEditBotConfig, SaveBotConfig, DeleteBotConfig, TestBotConfig,
      connStatusText, connStatusTagType,
      ShowBotMessageLog, LoadMessageLog, OnMessageLogClosed, FormatMsgTime,
      ShowBotLoopLog, ShowConfigLoopLog, LoadLoopLog, OnLoopLogClosed, HandleBotMoreOp,
      ShowAddRole, ShowEditRole, SaveRole, DeleteRole,
      ShowAddButlerConfig, ShowEditButlerConfig, SaveButlerConfig, DeleteButlerConfig,
      HandleInnerTabChange,
    }
  },
  components: {
    ArrowDown,
    QuestionFilled,
  },
})
</script>
<style scoped src="@/css/set_module_unified.css"></style>
<style scoped>
.set-config-inner-tabs :deep(.el-tabs__item) {
  height: 36px;
  color: #5c6856;
  font-weight: 500;
}
.set-config-inner-tabs :deep(.el-tabs__item.is-active) {
  color: #4f804f;
}
.set-config-inner-tabs :deep(.el-tabs__active-bar) {
  background-color: #4f804f;
}
.set-config-inner-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: #e8e8e0;
}

/* 表单问号帮助图标 */
.form-help-icon {
  margin-left: 4px;
  font-size: 14px;
  color: #909399;
  cursor: help;
  vertical-align: -2px;
}
.form-help-icon:hover {
  color: #4f804f;
}
</style>
