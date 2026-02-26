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
  Vue.prototype.$notify({
    title: '提示',
    message: msg,
    type: 'info',
    duration: 1000,
  })
}

function error(msg) {
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
