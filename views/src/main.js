// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'

//引入element ui
import 'element-ui/lib/theme-chalk/index.css';
import ElementUI from 'element-ui';
import '../src/css/reset_element.css';
Vue.use(ElementUI);

//引入加载进度动画
import NProgress from 'nprogress' // nprogress插件
import 'nprogress/nprogress.css' // nprogress样式

Vue.use(NProgress);
//自定义通用方法
import helperStore from './utils/store'
Vue.prototype.$helperStore = helperStore
import helperNotify from './utils/notify'
Vue.prototype.$helperNotify = helperNotify
import helperConfig from './utils/config'
Vue.prototype.$helperConfig = helperConfig
import helperCommon from './utils/common'
Vue.prototype.$helperCommon = helperCommon
import helperApi from './utils/api'
Vue.prototype.$helperApi = helperApi
import helperLoad from './utils/load'
Vue.prototype.$helperLoad = helperLoad
import ElButtonCustom from './components/base/button'
Vue.component('el-button-custom', ElButtonCustom)

//引入axios
import Axios from 'axios'
import VueAxios from 'vue-axios'
import { Message } from 'element-ui';
Axios.defaults.timeout = 100000
Axios.defaults.baseURL = '/'
Axios.defaults.headers.post['Content-Type'] = 'text/xml';
// 添加请求拦截器
Axios.interceptors.request.use((config) => {
    NProgress.done();
    NProgress.start();
    return config
  },
  (error) => {
    NProgress.done();
    return Promise.reject(error)
  }
)
Axios.interceptors.response.use(
  response => {
    setTimeout(function (){
      NProgress.set(0.6);
    },400)
    setTimeout(function (){
      NProgress.done();
    },700);
    if(response.data.ErrCode === 1){
      // window.location.reload();
      Message.error(response.data.ErrMsg);
    }else if(response.data.ErrCode !== 0 ){
      Message.error(response.data.ErrMsg);
    }
    return response.data;
  }
)
Vue.use(VueAxios,Axios);


import App from './App'
import router from './router'

Vue.config.productionTip = false


/* eslint-disable no-new */
new Vue({
  el: '#app',
  router,
  components: { App },
  template: '<App/>'
})
