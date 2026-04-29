import {globals} from '@/main'
function success(msg) {
  globals.$notify({
    title: '提示',
    message: msg,
    type: 'success',
    duration: 1000,
  })
}

function warning(msg) {
  globals.$notify({
    title: '提示',
    message: msg,
    type: 'warning',
    duration: 1000,
  })
}
function info(msg) {
  globals.$notify({
    title: '提示',
    message: msg,
    type: 'info',
    duration: 1000,
  })
}

function error(msg) {
  // 如果消息为空，不显示通知
  if (!msg || msg === '') {
    return
  }
  globals.$notify({
    title: '提示',
    message: msg,
    type: 'error',
    duration: 1000,
  })
}

export default {
  success,
  warning,
  info,
  error,
}
