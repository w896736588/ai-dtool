import base from '../base'

// McpTypeList 获取 MCP 类型列表
function McpTypeList(callBack) {
  base.BasePost('/api/McpTypeList', {}, callBack)
}

// McpBindingList 获取指定 MCP 类型的目录绑定列表
function McpBindingList(mcpType, agentTargetId, callBack) {
  base.BasePost('/api/McpBindingList', {
    mcp_type: mcpType,
    agent_target_id: agentTargetId,
  }, callBack)
}

// McpBindingAdd 添加绑定
function McpBindingAdd(mcpType, mappingId, agentTargetId, callBack) {
  base.BasePost('/api/McpBindingAdd', {
    mcp_type: mcpType,
    mapping_id: mappingId,
    agent_target_id: agentTargetId,
  }, callBack)
}

// McpBindingRemove 移除绑定
function McpBindingRemove(bindingId, callBack) {
  base.BasePost('/api/McpBindingRemove', {
    binding_id: bindingId,
  }, callBack)
}

// McpBindingInstruction 获取 AI 使用说明
function McpBindingInstruction(mcpType, mappingId, callBack) {
  base.BasePost('/api/McpBindingInstruction', {
    mcp_type: mcpType,
    mapping_id: mappingId,
  }, callBack)
}

// McpAgentTargetList 获取目标智能体列表
function McpAgentTargetList(callBack) {
  base.BasePost('/api/McpAgentTargetList', {}, callBack)
}

// McpAgentTargetSave 新增/编辑目标智能体
function McpAgentTargetSave(data, callBack) {
  base.BasePost('/api/McpAgentTargetSave', data, callBack)
}

// McpAgentTargetDelete 删除目标智能体
function McpAgentTargetDelete(id, callBack) {
  base.BasePost('/api/McpAgentTargetDelete', { id: id }, callBack)
}

// McpConfigPreview 获取配置文件预览（前后对比）
function McpConfigPreview(agentTargetId, callBack) {
  base.BasePost('/api/McpConfigPreview', {
    agent_target_id: agentTargetId,
  }, callBack)
}

export default {
  McpTypeList,
  McpBindingList,
  McpBindingAdd,
  McpBindingRemove,
  McpBindingInstruction,
  McpAgentTargetList,
  McpAgentTargetSave,
  McpAgentTargetDelete,
  McpConfigPreview,
}
