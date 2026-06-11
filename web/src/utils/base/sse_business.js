// 业务级独立 SSE 连接管理
// AgentCli 和 TaskWorkflow 各自维护一条独立 SSE 长连接，消息格式与全局 SSE 一致
import base from '@/utils/base'
import sseDistribute from '@/utils/base/sse_distribute'

// 每个业务独立维护自己的 EventSource、URL 和分发回调
const businessSse = {
  agent_cli: { conn: null, url: '', receiveHandlers: {} },
  task_workflow: { conn: null, url: '', receiveHandlers: {} },
}

// ConnectBusinessSse 创建到指定业务 SSE 路由的 EventSource
// businessType: 'agent_cli' | 'task_workflow'
// ssePort: SSE 端口
// clientId: 客户端唯一标识
function ConnectBusinessSse(businessType, ssePort, clientId) {
  if (!businessType || !businessSse[businessType]) {
    console.error('[sse_business] 无效的 businessType:', businessType)
    return false
  }
  const biz = businessSse[businessType]
  let params = 'client_id=' + encodeURIComponent(clientId) + '&token=' + encodeURIComponent(base.GetSafeToken())
  const sseHost = base.GetSseApiHost(ssePort || undefined)
  if (!sseHost) {
    console.error('[sse_business] 无法获取 SSE API Host')
    return false
  }
  const url = sseHost + '/sse/' + businessType + '?' + params
  if (biz.conn && biz.url === url) {
    return true
  }
  // 关闭旧连接
  if (biz.conn) {
    biz.conn.close()
    biz.conn = null
  }
  biz.url = url
  biz.receiveHandlers = {}
  biz.conn = new EventSource(url)
  biz.conn.onmessage = function (event) {
    let objData = null
    try {
      objData = JSON.parse(event.data)
    } catch (e) {
      return
    }
    if (objData && objData.sse_distribute_id) {
      const receiveHandlers = biz.receiveHandlers[objData.sse_distribute_id]
      if (receiveHandlers) {
        try {
          if (receiveHandlers instanceof Set) {
            receiveHandlers.forEach(function (handler) {
              if (typeof handler === 'function') {
                handler(objData.data, objData.type, objData.sse_distribute_id)
              }
            })
          } else if (typeof receiveHandlers === 'function') {
            receiveHandlers(objData.data, objData.type, objData.sse_distribute_id)
          }
        } catch (e) {
          console.log('[sse_business] 回调处理失败', businessType, e)
        }
      }
    }
  }
  biz.conn.onerror = function (event) {
    console.warn('[sse_business] 连接错误，EventSource 将自动重连:', businessType, event)
  }
  return true
}

// RegisterBusinessReceive 在指定业务的 SSE 连接上注册分发回调
// businessType: 'agent_cli' | 'task_workflow'
// receiveId: 分发 ID，如 'agent_cli_chat_output'
// callFunc: 回调函数
function RegisterBusinessReceive(businessType, receiveId, callFunc) {
  if (!businessType || !businessSse[businessType] || !receiveId || typeof callFunc !== 'function') {
    return
  }
  const biz = businessSse[businessType]
  const currentHandlers = biz.receiveHandlers[receiveId]
  if (!currentHandlers) {
    biz.receiveHandlers[receiveId] = new Set([callFunc])
    return
  }
  if (currentHandlers instanceof Set) {
    currentHandlers.add(callFunc)
    return
  }
  if (typeof currentHandlers === 'function') {
    biz.receiveHandlers[receiveId] = new Set([currentHandlers, callFunc])
  }
}

// UnRegisterBusinessReceive 注销指定业务的分发回调
// businessType: 'agent_cli' | 'task_workflow'
// receiveId: 分发 ID
// callFunc: 要注销的回调函数
function UnRegisterBusinessReceive(businessType, receiveId, callFunc) {
  if (!businessType || !businessSse[businessType] || !businessSse[businessType].receiveHandlers[receiveId]) {
    return
  }
  const biz = businessSse[businessType]
  if (typeof callFunc !== 'function') {
    delete biz.receiveHandlers[receiveId]
    return
  }
  const currentHandlers = biz.receiveHandlers[receiveId]
  if (currentHandlers instanceof Set) {
    currentHandlers.delete(callFunc)
    if (currentHandlers.size === 0) {
      delete biz.receiveHandlers[receiveId]
    }
    return
  }
  if (currentHandlers === callFunc) {
    delete biz.receiveHandlers[receiveId]
  }
}

// CloseBusinessSse 关闭指定业务的 SSE 连接，清空所有回调
// businessType: 'agent_cli' | 'task_workflow'
function CloseBusinessSse(businessType) {
  if (!businessType || !businessSse[businessType]) {
    return
  }
  const biz = businessSse[businessType]
  if (biz.conn) {
    biz.conn.close()
    biz.conn = null
  }
  biz.url = ''
  biz.receiveHandlers = {}
}

// 复用全局 SSE 的端口查询
function fetchAvailableSsePort() {
  return sseDistribute.fetchAvailableSsePort ? sseDistribute.fetchAvailableSsePort() : Promise.resolve(null)
}

export default {
  ConnectBusinessSse,
  RegisterBusinessReceive,
  UnRegisterBusinessReceive,
  CloseBusinessSse,
  fetchAvailableSsePort,
}
