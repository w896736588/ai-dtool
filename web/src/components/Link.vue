<template>
  <Links @changeModelToFlow="changeToFlow" @changeModelToEditProcess="changeToEditProcess" v-if="model === 'links'"/>
  <Process @changeModelToLinks="changeToLinks" v-if="model === 'process'"/>
  <Flow @changeModelToLinks="changeToLinks" @changeModelToFlow="changeToFlow" v-if="model === 'flow'"/>
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
    }
  }
}
</script>

<style scoped>
:deep(.link-run-page),
:deep(.smart-process-container),
:deep(.smart-process-canvas-container) {
  color: #3f4a3f;
  --el-color-primary: #6fa56f;
  --el-color-primary-light-3: #8db88d;
  --el-color-primary-light-5: #a7c8a7;
  --el-color-primary-light-7: #c2dac2;
  --el-color-primary-light-8: #d5e6d5;
  --el-color-primary-light-9: #e7f1e7;
  --el-color-primary-dark-2: #5f8f5f;
}

:deep(.smart-process-container),
:deep(.smart-process-canvas-container) {
  height: calc(100vh - 140px);
  background: #fafaf7;
  border: 1px solid #e6e8de;
  border-radius: 10px;
  overflow: hidden;
}

:deep(.left-sidebar) {
  background: #f5f5f0;
  border-right: 1px solid #e6e8de;
}

:deep(.right-content) {
  background: #fafaf7;
}

:deep(.search-box),
:deep(.add-btn),
:deep(.canvas-header) {
  background: #f7f8f2;
}

:deep(.canvas-header) {
  border-bottom: 1px solid #e6e8de;
}

:deep(.process-item) {
  border-radius: 6px;
  margin: 2px 8px;
  color: #465246;
}

:deep(.process-item:hover) {
  background: #e8f2e5 !important;
}

:deep(.process-item.active) {
  background: #dcedc8 !important;
  color: #3a7a3a;
}

:deep(.process-item-card),
:deep(.box-card),
:deep(.custom-node) {
  background: #ffffff;
  border: 1px solid #e6e8de;
  border-radius: 10px;
  box-shadow: 0 1px 2px rgba(80, 96, 80, 0.08);
}

:deep(.custom-node.selected) {
  border-color: #7cb87c;
  box-shadow: 0 0 0 2px rgba(124, 184, 124, 0.2);
}

:deep(.node-type) {
  background: #f1f5ec;
  color: #4f5b4f;
}

:deep(.el-link.el-link--primary),
:deep(a) {
  color: #4a8b4a;
}
</style>
