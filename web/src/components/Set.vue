<template>
  <div class="set-page-tabs">
    <el-tabs v-model="activeLabel" tab-position="top" class="set-tabs" @tab-click="handleTabClick">
      <el-tab-pane label="Ssh" name="Ssh" class="set-tab-pane">
        <ssh ref="ssh"></ssh>
      </el-tab-pane>
      <el-tab-pane label="Mysql" name="Mysql" class="set-tab-pane">
        <mysql ref="mysql"></mysql>
      </el-tab-pane>
      <el-tab-pane label="Global" name="Global" class="set-tab-pane">
        <global ref="global"></global>
      </el-tab-pane>
      <el-tab-pane label="AI" name="AI" class="set-tab-pane">
        <ai_provider ref="ai_provider"></ai_provider>
      </el-tab-pane>
      <el-tab-pane label="Config" name="Config" class="set-tab-pane">
        <memory ref="memory" :show-runtime-config="true"></memory>
      </el-tab-pane>
      <el-tab-pane label="Schedule" name="Schedule" class="set-tab-pane">
        <cron_setting ref="cron_setting"></cron_setting>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import set from '@/utils/base/ssh_set'
import ssh from './set/ssh.vue'
import mysql from './set/mysql.vue'
import store from '@/utils/base/store'
import global from '@/components/set/global.vue'
import ai_provider from '@/components/set/ai_provider.vue'
import memory from '@/components/set/memory.vue'
import cron_setting from '@/components/set/cron_setting.vue'

// SET_ACTIVE_TABS 定义当前仍保留在配置页中的标签页，避免旧缓存命中已迁出的业务设置。
// Keep the tabs that still belong to the settings page to avoid stale cache pointing to moved pages.
const SET_ACTIVE_TABS = ['Ssh', 'Mysql', 'Global', 'AI', 'Config', 'Schedule']

export default {
  props: {
    shellShowResult: {
      type: String,
    },
  },
  components: {
    ssh,
    mysql,
    global,
    ai_provider,
    memory,
    cron_setting,
  },
  data() {
    return {
      name: 'Ssh',
      activeLabel: 'Ssh',
      sshList: [],
    }
  },
  mounted() {
    if (process.env.NODE_ENV === 'production') {
      this.apiHost = ''
    }
    this.syncActiveLabel()
    this.SshList()
  },
  activated() {
    this.syncActiveLabel()
  },
  methods: {
    syncActiveLabel() {
      this.activeLabel = String(store.getStore('set_active_label'))
      if (this.activeLabel === '' || !SET_ACTIVE_TABS.includes(this.activeLabel)) {
        this.activeLabel = 'Ssh'
      }
      this.loadActiveTabData()
    },
    handleTabClick(tab) {
      this.activeLabel = tab.props.name
      store.setStore('set_active_label', tab.props.name)
      this.loadActiveTabData()
    },
    // loadActiveTabData 在切换配置标签时按需刷新当前页数据，避免全部标签同时请求。
    // Refresh only the active settings tab on demand instead of loading every tab at once.
    loadActiveTabData() {
      switch (this.activeLabel) {
        case 'Ssh':
          this.$refs.ssh && this.$refs.ssh.SshList()
          break
        case 'AI':
          this.$refs.ai_provider && this.$refs.ai_provider.LoadProviderList()
          this.$refs.ai_provider && this.$refs.ai_provider.LoadModelList()
          break
        case 'Config':
          this.$refs.memory && this.$refs.memory.loadConfig && this.$refs.memory.loadConfig()
          this.$refs.memory && this.$refs.memory.loadAiModelList && this.$refs.memory.loadAiModelList()
          break
        case 'Schedule':
          this.$refs.cron_setting && this.$refs.cron_setting.loadConfig && this.$refs.cron_setting.loadConfig()
          break
        default:
          break
      }
    },
    SshList() {
      let _that = this
      set.SshList(function (response) {
        if (response.ErrCode === 0) {
          _that.sshList = response.Data
        }
      })
    },
    getStore(key) {
      return localStorage.getItem(key)
    },
  },
}
</script>

<style scoped src="@/css/components/Set.css"></style>
