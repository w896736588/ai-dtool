import base from '../base'

// AgentCliList 获取 Agent Cli 列表（含状态摘要）
function AgentCliList(callBack) {
  base.BasePost('/api/AgentCliList', {}, callBack)
}

// AgentCliSave 新建/编辑
function AgentCliSave(data, callBack) {
  base.BasePost('/api/AgentCliSave', data, callBack)
}

// AgentCliDelete 删除
function AgentCliDelete(id, callBack) {
  base.BasePost('/api/AgentCliDelete', { id: id }, callBack)
}

// AgentCliReadSettings 读取 settings.json
function AgentCliReadSettings(id, callBack) {
  base.BasePost('/api/AgentCliReadSettings', { id: id }, callBack)
}

// AgentCliWriteMcpServers 写入 mcpServers
function AgentCliWriteMcpServers(id, callBack) {
  base.BasePost('/api/AgentCliWriteMcpServers', { id: id }, callBack)
}

// AgentCliWriteDeepSeek 写入 DeepSeek 配置
function AgentCliWriteDeepSeek(data, callBack) {
  base.BasePost('/api/AgentCliWriteDeepSeek', data, callBack)
}

// AgentCliToggleEnabled 切换 Agent CLI 启停
function AgentCliToggleEnabled(data, callBack) {
  base.BasePost('/api/AgentCliToggleEnabled', data, callBack)
}

// AgentChatSend 发送独立 Agent CLI 对话
function AgentChatSend(data, callBack) {
  base.BasePost('/api/agent/chat/send', data, callBack)
}

// AgentChatListByAgentCli 按 Agent CLI 查询独立执行历史
function AgentChatListByAgentCli(agentCliId, callBack) {
  base.BasePost('/api/agent/chat/list-by-agent-cli', {
    agent_cli_id: agentCliId,
  }, callBack)
}

// AgentChatMarkRead 将执行历史对话标记为已读
function AgentChatMarkRead(chatId, callBack) {
  base.BasePost('/api/agent/chat/mark-read', {
    chat_id: chatId,
  }, callBack)
}

// WebhookConfigList 获取 Webhook 配置列表
function WebhookConfigList(callBack) {
  base.BasePost('/api/WebhookConfigList', {}, callBack)
}

// WebhookConfigSave 新建/编辑 Webhook 配置
function WebhookConfigSave(data, callBack) {
  base.BasePost('/api/WebhookConfigSave', data, callBack)
}

// WebhookConfigDelete 删除 Webhook 配置
function WebhookConfigDelete(id, callBack) {
  base.BasePost('/api/WebhookConfigDelete', { id: id }, callBack)
}

// WebhookConfigTest 测试发送 Webhook 配置
function WebhookConfigTest(data, callBack) {
  base.BasePost('/api/WebhookConfigTest', data, callBack)
}

// AgentCliGroupList 获取 AgentCli 专用分组列表
function AgentCliGroupList(callBack) {
  base.BasePost('/api/AgentCliGroupList', {}, callBack)
}

// AgentCliGroupSave 新增/编辑 AgentCli 分组
function AgentCliGroupSave(data, callBack) {
  base.BasePost('/api/AgentCliGroupSave', data, callBack)
}

// AgentCliGroupDelete 删除 AgentCli 分组
function AgentCliGroupDelete(id, callBack) {
  base.BasePost('/api/AgentCliGroupDelete', { id: id }, callBack)
}

// AgentCliGroupRelSave 保存 AgentCli 的分组关联
function AgentCliGroupRelSave(data, callBack) {
  base.BasePost('/api/AgentCliGroupRelSave', data, callBack)
}

export default {
  AgentCliList,
  AgentCliSave,
  AgentCliDelete,
  AgentCliReadSettings,
  AgentCliWriteMcpServers,
  AgentCliWriteDeepSeek,
  AgentCliToggleEnabled,
  AgentChatSend,
  AgentChatListByAgentCli,
  AgentChatMarkRead,
  WebhookConfigList,
  WebhookConfigSave,
  WebhookConfigDelete,
  WebhookConfigTest,
  AgentCliGroupList,
  AgentCliGroupSave,
  AgentCliGroupDelete,
  AgentCliGroupRelSave,
}
