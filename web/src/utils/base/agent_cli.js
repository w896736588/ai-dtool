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

// AgentCliToggleClaudeMem 切换 claude-mem 启停
function AgentCliToggleClaudeMem(data, callBack) {
  base.BasePost('/api/AgentCliToggleClaudeMem', data, callBack)
}

// AgentCliToggleEnabled 切换 Agent CLI 启停
function AgentCliToggleEnabled(data, callBack) {
  base.BasePost('/api/AgentCliToggleEnabled', data, callBack)
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

export default {
  AgentCliList,
  AgentCliSave,
  AgentCliDelete,
  AgentCliReadSettings,
  AgentCliWriteMcpServers,
  AgentCliWriteDeepSeek,
  AgentCliToggleClaudeMem,
  AgentCliToggleEnabled,
  WebhookConfigList,
  WebhookConfigSave,
  WebhookConfigDelete,
}
