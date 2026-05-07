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
    ElMessage.error(response.data.ErrMsg)
  } else if (response.data.ErrCode !== 0) {
    ElMessage.error(response.data.ErrMsg)
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
const debounce = (fn, delay) => {
  let timer = null;
  return function () {
    let context = this;
    let args = arguments;
    clearTimeout(timer);
    timer = setTimeout(function () {
      try{
        fn.apply(context, args);  //会影响table自动宽度
      }catch (error){
        console.log(error)
      }

    }, delay);
  }
}

const _ResizeObserver = window.ResizeObserver;
window.ResizeObserver = class ResizeObserver extends _ResizeObserver{
  constructor(callback) {
    callback = debounce(callback, 100);
    super(callback);
  }
}

import sse from '@/utils/base/sse_distribute'

// 不需要 SSE 连接的独立页面路由（这些页面不依赖 SSE 推送，无需占用浏览器连接数）
const SSE_EXCLUDED_ROUTES = ['/HomeTaskSetting', '/ApiDocument', '/TaskWorkflow']

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
    console.log('链接错误', e && e.message)
  }, function (){
    console.log('链接关闭')
  })
}

