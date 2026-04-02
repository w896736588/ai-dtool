<template>
  <div class="link-module-shell">
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
.link-module-shell {
  width: 100%;
}

.link-module-shell :deep(.link-run-page),
.link-module-shell :deep(.smart-process-container),
.link-module-shell :deep(.smart-process-canvas-container) {
  color: #3f4a3f;
  --el-color-primary: #6fa56f;
  --el-color-primary-light-3: #8db88d;
  --el-color-primary-light-5: #a7c8a7;
  --el-color-primary-light-7: #c2dac2;
  --el-color-primary-light-8: #d5e6d5;
  --el-color-primary-light-9: #e7f1e7;
  --el-color-primary-dark-2: #5f8f5f;
}

.link-module-shell :deep(.git-action-button--primary),
.link-module-shell :deep(.pl-button--primary) {
  --git-button-text-color: #ffffff;
  --git-button-border-color: #4f804f;
  --git-button-background-color: linear-gradient(180deg, #5a8a5a 0%, #4f804f 100%);
  --git-button-hover-text-color: #ffffff;
  --git-button-hover-border-color: #3f6f3f;
  --git-button-hover-background-color: linear-gradient(180deg, #4f804f 0%, #3f6f3f 100%);
  --pl-button-text-color: #ffffff;
  --pl-button-border-color: #4f804f;
  --pl-button-background-color: linear-gradient(180deg, #5a8a5a 0%, #4f804f 100%);
  --pl-button-hover-text-color: #ffffff;
  --pl-button-hover-border-color: #3f6f3f;
  --pl-button-hover-background-color: linear-gradient(180deg, #4f804f 0%, #3f6f3f 100%);
  color: #ffffff !important;
  border-color: #4f804f !important;
  background: linear-gradient(180deg, #5a8a5a 0%, #4f804f 100%) !important;
}

.link-module-shell :deep(.git-action-button--primary:hover),
.link-module-shell :deep(.git-action-button--primary:focus-visible),
.link-module-shell :deep(.pl-button--primary:hover),
.link-module-shell :deep(.pl-button--primary:focus-visible) {
  color: #ffffff !important;
  border-color: #3f6f3f !important;
  background: linear-gradient(180deg, #4f804f 0%, #3f6f3f 100%) !important;
}

.link-module-shell :deep(.git-action-button--warning),
.link-module-shell :deep(.git-action-button--info),
.link-module-shell :deep(.pl-button--warning),
.link-module-shell :deep(.pl-button--info) {
  --git-button-text-color: #ffffff;
  --git-button-border-color: #4f804f;
  --git-button-background-color: linear-gradient(180deg, #5a8a5a 0%, #4f804f 100%);
  --git-button-hover-text-color: #ffffff;
  --git-button-hover-border-color: #3f6f3f;
  --git-button-hover-background-color: linear-gradient(180deg, #4f804f 0%, #3f6f3f 100%);
  --pl-button-text-color: #ffffff;
  --pl-button-border-color: #4f804f;
  --pl-button-background-color: linear-gradient(180deg, #5a8a5a 0%, #4f804f 100%);
  --pl-button-hover-text-color: #ffffff;
  --pl-button-hover-border-color: #3f6f3f;
  --pl-button-hover-background-color: linear-gradient(180deg, #4f804f 0%, #3f6f3f 100%);
  color: #ffffff !important;
  border-color: #4f804f !important;
  background: linear-gradient(180deg, #5a8a5a 0%, #4f804f 100%) !important;
}

.link-module-shell :deep(.git-action-button--warning:hover),
.link-module-shell :deep(.git-action-button--warning:focus-visible),
.link-module-shell :deep(.git-action-button--info:hover),
.link-module-shell :deep(.git-action-button--info:focus-visible),
.link-module-shell :deep(.pl-button--warning:hover),
.link-module-shell :deep(.pl-button--warning:focus-visible),
.link-module-shell :deep(.pl-button--info:hover),
.link-module-shell :deep(.pl-button--info:focus-visible) {
  color: #ffffff !important;
  border-color: #3f6f3f !important;
  background: linear-gradient(180deg, #4f804f 0%, #3f6f3f 100%) !important;
}

.link-module-shell :deep(.git-action-button--danger),
.link-module-shell :deep(.pl-button--danger) {
  --git-button-text-color: #ffffff;
  --git-button-border-color: #d65c5c;
  --git-button-background-color: linear-gradient(180deg, #de6f6f 0%, #d65c5c 100%);
  --git-button-hover-text-color: #ffffff;
  --git-button-hover-border-color: #bb4747;
  --git-button-hover-background-color: linear-gradient(180deg, #c95757 0%, #bb4747 100%);
  --pl-button-text-color: #ffffff;
  --pl-button-border-color: #d65c5c;
  --pl-button-background-color: linear-gradient(180deg, #de6f6f 0%, #d65c5c 100%);
  --pl-button-hover-text-color: #ffffff;
  --pl-button-hover-border-color: #bb4747;
  --pl-button-hover-background-color: linear-gradient(180deg, #c95757 0%, #bb4747 100%);
  color: #ffffff !important;
  border-color: #d65c5c !important;
  background: linear-gradient(180deg, #de6f6f 0%, #d65c5c 100%) !important;
}

.link-module-shell :deep(.git-action-button--danger:hover),
.link-module-shell :deep(.git-action-button--danger:focus-visible),
.link-module-shell :deep(.pl-button--danger:hover),
.link-module-shell :deep(.pl-button--danger:focus-visible) {
  color: #ffffff !important;
  border-color: #bb4747 !important;
  background: linear-gradient(180deg, #c95757 0%, #bb4747 100%) !important;
}

.link-module-shell :deep(.git-action-button.pl-button--plain),
.link-module-shell :deep(.pl-button--plain.pl-button--primary),
.link-module-shell :deep(.pl-button--plain.pl-button--warning),
.link-module-shell :deep(.pl-button--plain.pl-button--info),
.link-module-shell :deep(.pl-button--plain.pl-button--danger) {
  color: #ffffff !important;
}

.link-module-shell :deep(.smart-process-container),
.link-module-shell :deep(.smart-process-canvas-container) {
  height: calc(100vh - 140px);
  background: #fafaf7;
  border: 1px solid #e6e8de;
  border-radius: 10px;
  overflow: hidden;
}

.link-module-shell :deep(.left-sidebar) {
  background: #f5f5f0;
  border-right: 1px solid #e6e8de;
}

.link-module-shell :deep(.right-content) {
  background: #fafaf7;
}

.link-module-shell :deep(.search-box),
.link-module-shell :deep(.add-btn),
.link-module-shell :deep(.canvas-header) {
  background: #f7f8f2;
}

.link-module-shell :deep(.canvas-header) {
  border-bottom: 1px solid #e6e8de;
}

.link-module-shell :deep(.process-item) {
  border-radius: 6px;
  margin: 2px 8px;
  color: #465246;
}

.link-module-shell :deep(.process-item:hover) {
  background: #e8f2e5 !important;
}

.link-module-shell :deep(.process-item.active) {
  background: #dcedc8 !important;
  color: #3a7a3a;
}

.link-module-shell :deep(.process-item-card),
.link-module-shell :deep(.box-card),
.link-module-shell :deep(.custom-node) {
  background: #ffffff;
  border: 1px solid #e6e8de;
  border-radius: 10px;
  box-shadow: 0 1px 2px rgba(80, 96, 80, 0.08);
}

.link-module-shell :deep(.custom-node.selected) {
  border-color: #7cb87c;
  box-shadow: 0 0 0 2px rgba(124, 184, 124, 0.2);
}

.link-module-shell :deep(.node-type) {
  background: #f1f5ec;
  color: #4f5b4f;
}

.link-module-shell :deep(.el-link.el-link--primary),
.link-module-shell :deep(a) {
  color: #4a8b4a;
}
</style>
