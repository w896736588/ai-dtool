// The Vue build version to load with the `import` command
// (runtime-only or standalone) has been set in webpack.base.conf with an alias.
import Vue from 'vue'

//引入element ui
import 'element-ui/lib/theme-chalk/index.css';
import ElementUI from 'element-ui';
import '../src/css/reset_element.css';
Vue.use(ElementUI);

//引入jsonp
import { VueJsonp } from 'vue-jsonp'
Vue.use(VueJsonp)

//引入axios
import Axios from 'axios'
import VueAxios from 'vue-axios'
import { Message } from 'element-ui';
Axios.defaults.timeout = 15000
Axios.defaults.baseURL = '/api'
Axios.defaults.headers.post['Content-Type'] = 'text/xml';
Axios.interceptors.response.use(
  response => {
    if(response.data.ErrCode === 1){
      window.location.reload();
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
