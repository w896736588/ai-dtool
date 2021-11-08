import Vue from 'vue'
import Router from 'vue-router'
import HelloWorld from '@/components/HelloWorld'
import CacheIndex from '@/components/CacheIndex'
Vue.use(Router)

export default new Router({
  routes: [
    {
      path: '/',
      name: 'cacheIndex',
      component: CacheIndex
    }
  ]
})
