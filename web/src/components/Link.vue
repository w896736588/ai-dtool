<template>
  <div class="link-module-shell">
    <div class="link-module-header">
      <div class="link-module-header__title">自定义网页工作台</div>
      <div class="link-module-header__switch">
        <span class="link-module-header__switch-label" :class="{ 'is-active': !isLinks }">运行逻辑</span>
        <el-switch
          v-model="isLinks"
          class="link-module-header__switch-control"
        />
        <span class="link-module-header__switch-label" :class="{ 'is-active': isLinks }">执行</span>
      </div>
    </div>
    <Links @changeModelToFlow="changeToFlow" @changeModelToEditProcess="changeToEditProcess" v-if="model === 'links'"/>
    <Process @changeModelToLinks="changeToLinks" v-if="model === 'process'"/>
    <Flow @changeModelToLinks="changeToLinks" @changeModelToFlow="changeToFlow" v-if="model === 'flow'"/>
  </div>
</template>
<script>
import Links from '@/components/smart_link/link_run.vue'
import Process from '@/components/smart_link/link_process.vue'
import Flow from '@/components/smart_link/link_flow.vue'
import store from '@/utils/base/store'
export default {
  props: {
    shellShowResult: {
      type: String
    },
  },
  components: {
    Links,
    Process,
    Flow,
  },
  data() {
    return {
      model : 'links',
    }
  },
  computed: {
    // isLinks 开关状态：开=执行(links)，关=运行逻辑(process/flow 均视为运行逻辑一侧
    isLinks: {
      get() {
        return this.model === 'links'
      },
      set(val) {
        if (val) {
          this.changeToLinks()
        } else {
          this.changeToEditProcess()
        }
      },
    },
  },
  mounted: function () {
    let _that = this
    let linkModel = store.getStore('link_model')
    if(!linkModel){
      _that.model = 'links'
    }else{
      _that.model = linkModel
    }
  },
  methods: {
    changeToEditProcess : function (){
      let _that = this
      _that.model = 'process'
      store.setStore('link_model' , _that.model)
    },
    changeToLinks : function (){
      let _that = this
      _that.model = 'links'
      store.setStore('link_model' , _that.model)
    },
    changeToFlow : function (){
      let _that = this
      _that.model = 'flow'
      store.setStore('link_model' , _that.model)
    }
  }
}
</script>

<style scoped src="@/css/components/Link.css"></style>
