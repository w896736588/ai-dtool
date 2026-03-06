import store from "./base/store"
import base from './base'
function GetSystemName() {
  return 'zhima'
}

//启用的模块
function GetOpenModuleList() {
  switch (GetSystemName()) {
    case 'zhima':
      return ['redis', 'supervisor', 'command', 'git' , 'login' , 'variable' , 'model' , 'open_link' , 'shellout' , 'qr_code' , 'time_transfer' , 'tools' , 'docker' , 'code' , 'markdown' , 'api', 'memory_fragment']
    default:
      return ['redis', 'supervisor', 'git' , 'tools' , 'markdown', 'memory_fragment']
  }
}

//拿到格式化列表
function GetFormatList() {
  return require('../config/' + GetSystemName() + '/formatResult.json')
}

export default {
  GetOpenModuleList,
  GetFormatList,
}
