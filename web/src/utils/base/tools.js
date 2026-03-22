import base from '@/utils/base'

function ToolPortProcessList(data, callBack) {
  base.BasePost('/api/ToolPortProcessList', data, callBack)
}

function ToolPortProcessKill(data, callBack) {
  base.BasePost('/api/ToolPortProcessKill', data, callBack)
}

function ToolManagedProcessStatus(data, callBack) {
  base.BasePost('/api/ToolManagedProcessStatus', data, callBack)
}

function ToolManagedProcessEnsureRunning(data, callBack) {
  base.BasePost('/api/ToolManagedProcessEnsureRunning', data, callBack)
}

function ToolManagedProcessStart(data, callBack) {
  base.BasePost('/api/ToolManagedProcessStart', data, callBack)
}

function ToolManagedProcessStop(data, callBack) {
  base.BasePost('/api/ToolManagedProcessStop', data, callBack)
}

function ToolManagedProcessRestart(data, callBack) {
  base.BasePost('/api/ToolManagedProcessRestart', data, callBack)
}

function ToolManagedProcessLogTail(data, callBack) {
  base.BasePost('/api/ToolManagedProcessLogTail', data, callBack)
}

export default {
  ToolPortProcessList,
  ToolPortProcessKill,
  ToolManagedProcessStatus,
  ToolManagedProcessEnsureRunning,
  ToolManagedProcessStart,
  ToolManagedProcessStop,
  ToolManagedProcessRestart,
  ToolManagedProcessLogTail,
}
