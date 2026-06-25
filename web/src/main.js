import { createApp } from 'vue'
import App from './App.vue'
const app = createApp(App)

//引入element ui
import ElementUI from 'element-plus'
import 'element-plus/dist/index.css'
import '../src/css/reset_element.css'
import '../src/css/el_textarea.css'
import '../src/css/pretty_json.css'
import '../src/css/markdown.css'
import '../src/css/api_module_unified.css'
import { config as configureMdEditor } from 'md-editor-v3'
app.use(ElementUI)
import { ElMessage } from 'element-plus'
const { buildMdEditorCodeMirrorExtensions } = require('./utils/md_editor_config.cjs')

configureMdEditor({
  codeMirrorExtensions(extensions) {
    return buildMdEditorCodeMirrorExtensions(extensions)
  },
  markdownItConfig(mdit) {
    const defaultRender = mdit.renderer.rules.link_open || function (tokens, idx, options, _env, self) {
      return self.renderToken(tokens, idx, options)
    }
    mdit.renderer.rules.link_open = function (tokens, idx, options, env, self) {
      tokens[idx].attrSet('target', '_blank')
      tokens[idx].attrSet('rel', 'noopener noreferrer')
      return defaultRender(tokens, idx, options, env, self)
    }
  },
})


//自定义通用方法
import helperStore from './utils/base/store'
app.config.globalProperties.$helperStore = helperStore
import helperNotify from './utils/base/notify'
app.config.globalProperties.$helperNotify = helperNotify
import helperConfig from './utils/config'
app.config.globalProperties.$helperConfig = helperConfig
import helperCommon from './utils/common'
app.config.globalProperties.$helperCommon = helperCommon
import helperApi from './utils/api'
app.config.globalProperties.$helperApi = helperApi
import helperLoad from './utils/load'
app.config.globalProperties.$helperLoad = helperLoad
import ElButtonCustom from './components/base/button'
app.component('el-button-custom', ElButtonCustom)
app.component('pl-button', ElButtonCustom)

//各个页面初始化的定义 用来解决每次初始化就调用所有接口问题
app.config.globalProperties.$pageInit = {}


//引入axios
import Axios from 'axios'
import VueAxios from 'vue-axios'
import SseDistribute from '@/utils/base/sse_distribute'
Axios.defaults.timeout = 600000
Axios.defaults.baseURL = '/'
Axios.defaults.headers.post['Content-Type'] = 'text/xml'
// 添加请求拦截器
Axios.interceptors.request.use(
  (config) => {
    //默认添加sse_client_id
    config.headers.SseClientId = SseDistribute.GetSseClientId()
    return config
  },
  (error) => {
    return Promise.reject(error)
  }
)
Axios.interceptors.response.use((response) => {
  if (response.data.ErrCode === 1) {
    // 打印接口路径和响应体，方便定位异常接口
    console.error('[API Error] ErrCode=1', response.config.url, JSON.stringify(response.data))
    ElMessage.error(response.data.ErrMsg || '请求失败')
  } else if (response.data.ErrCode !== 0) {
    console.error('[API Error] ErrCode=' + response.data.ErrCode, response.config.url, JSON.stringify(response.data))
    ElMessage.error(response.data.ErrMsg || '请求异常')
  }
  return response.data
} , error => {
  if (error.code === 'ECONNABORTED') {
    console.log('请求超时');
  } else {
    ElMessage.error(error.message)
  }
  return Promise.reject(error); // 一定要返回一个 rejected 状态的 Promise，否则请求将会被重试
})
app.use(VueAxios, Axios)
app.config.globalProperties.$axios = Axios

// 创建全局事件总线（用于登录失效等全局事件）
import mitt from 'mitt'
const eventBus = mitt()
app.config.globalProperties.$eventBus = eventBus

import router from './router/index'

app.config.productionTip = false
import JsonViewer from "vue3-json-viewer"
app.use(JsonViewer)

app.use(router).mount('#app')

import * as ElementPlusIconsVue from '@element-plus/icons-vue'
for (const [key, component] of Object.entries(ElementPlusIconsVue)) {
  app.component(key, component)
}
export const globals = app.config.globalProperties

//解决不停报错 ERROR ResizeObserver loop completed with undelivered notifications.
// 同时防止组件销毁后，防抖回调因延迟执行而访问已卸载的 DOM 元素（如 el-select 的 resetSelectionWidth 调用 getBoundingClientRect 报错）
const debounce = (fn, delay, { onCleanup } = {}) => {
  let timer = null;
  const debouncedFn = function () {
    let context = this;
    let args = arguments;
    clearTimeout(timer);
    timer = setTimeout(function () {
      timer = null;
      try{
        // 检查 ResizeObserverEntry 中的 target 是否仍在 DOM 中，
        // 防止组件销毁后（如 el-dialog destroy-on-close）回调访问已卸载元素
        const entries = args && args[0]
        if (entries && entries.length > 0 && typeof entries[0].target !== 'undefined') {
          const validEntries = Array.from(entries).filter(entry => {
            return entry && entry.target && document.contains(entry.target)
          })
          if (validEntries.length === 0) {
            return  // 所有观察目标都已卸载，跳过本次回调
          }
        }
        fn.apply(context, args);  //会影响table自动宽度
      }catch (error){
        console.log(error)
      }
    }, delay);
  };
  // 保存清理函数引用，供 disconnect 时取消待执行的 timer
  if (onCleanup) {
    debouncedFn._cleanup = () => {
      if (timer) {
        clearTimeout(timer);
        timer = null;
      }
    };
  }
  return debouncedFn;
}

const _ResizeObserver = window.ResizeObserver;
const _roCleanupMap = new WeakMap();

window.ResizeObserver = class ResizeObserver extends _ResizeObserver{
  constructor(callback) {
    callback = debounce(callback, 100, { onCleanup: true });
    super(callback);
    _roCleanupMap.set(this, callback._cleanup || null);
  }
}

// 重写 disconnect 原型方法：断开观察时主动清除防抖 timer，防止组件销毁后的延迟回调
const _nativeDisconnect = _ResizeObserver.prototype.disconnect;
window.ResizeObserver.prototype.disconnect = function () {
  const cleanup = _roCleanupMap.get(this);
  if (typeof cleanup === 'function') {
    cleanup();
    _roCleanupMap.delete(this);
  }
  return _nativeDisconnect.call(this);
};

import sse from '@/utils/base/sse_distribute'

// 不需要 SSE 连接的独立页面路由（这些页面不依赖 SSE 推送，无需占用浏览器连接数）
const SSE_EXCLUDED_ROUTES = ['/HomeTaskSetting', '/ApiDocument', '/TaskWorkflow', '/fullpage']

function shouldInitSse() {
  const hash = window.location.hash
  if (!hash) return true
  const route = hash.startsWith('#') ? hash.slice(1) : hash
  return !SSE_EXCLUDED_ROUTES.some(r => route.startsWith(r))
}

if (shouldInitSse()) {
  sse.InitFromLoginStatus(function (){
    console.log('打开链接')
  }, function (e){
    // 无可用端口时的弹窗已在 sse_distribute.js 内统一处理
  }, function (){
    console.log('链接关闭')
  })
}

// 页面刷新或关闭时主动断开共享 SSE，避免同一 client_id 在浏览器侧短时间残留多条连接记录。
window.addEventListener('pagehide', () => {
  sse.Close()
})

