<template>
  <div class="set-page-tabs">
    <el-tabs v-model="activeLabel" tab-position="top" class="set-tabs" @tab-click="handleTabClick">
      <el-tab-pane label="Ssh" name="Ssh" class="set-tab-pane">
        <ssh ref="ssh"></ssh>
      </el-tab-pane>
      <el-tab-pane label="Git" name="Git" class="set-tab-pane">
        <git ref="git"></git>
      </el-tab-pane>
      <el-tab-pane label="Supervisor" name="Supervisor" class="set-tab-pane">
        <supervisor ref="supervisor"></supervisor>
      </el-tab-pane>
      <el-tab-pane label="Redis" name="Redis" class="set-tab-pane">
        <redis ref="redis"></redis>
      </el-tab-pane>
      <el-tab-pane label="Mysql" name="Mysql" class="set-tab-pane">
        <mysql ref="mysql"></mysql>
      </el-tab-pane>
<!--      <el-tab-pane label="脚本合集组">-->
<!--        <variable_group ref="variable_group"></variable_group>-->
<!--      </el-tab-pane>-->
      <el-tab-pane label="Compose" name="Compose" class="set-tab-pane">
        <compose ref="compose"></compose>
      </el-tab-pane>
      <el-tab-pane label="账号" name="Account" class="set-tab-pane">
        <account ref="account"></account>
      </el-tab-pane>
<!--      <el-tab-pane label="命令组">-->
<!--        <cmd_group ref="cmd_group"></cmd_group>-->
<!--      </el-tab-pane>-->
<!--      <el-tab-pane label="GitlabToken" name="GitlabToken" style="padding:5px;">-->
<!--        <gitlab_token ref="gitlabToken"></gitlab_token>-->
<!--      </el-tab-pane>-->
      <el-tab-pane label="Global" name="Global" class="set-tab-pane">
        <global ref="global"></global>
      </el-tab-pane>
    </el-tabs>
  </div>
</template>

<script>
import set from '@/utils/base/ssh_set'
import ssh from "./set/ssh.vue"
import git from "./set/git.vue"
import git_group from "./set/git_group.vue"
import supervisor from "./set/supervisor.vue"
import redis from "./set/redis.vue"
import mysql from "./set/mysql.vue"
import variable_group from "./set/variable_group.vue"
import Cmd_group from "@/components/set/cmd_group.vue";
import smart_link_group from "./set/smart_link_group.vue"
import compose from "./set/compose.vue"
import gitlab_token from "@/components/set/gitlab_token.vue"
import store from "@/utils/base/store"
import global from "@/components/set/global.vue"
import account from "@/components/set/account.vue";
export default {
  props : {
    shellShowResult : {
      type : String
    },
  },
  components: {
    account,
    ssh,
    git,
    git_group,
    supervisor,
    redis,
    mysql,
    compose,
    gitlab_token ,
    global,
  },
  data() {
    return {
      name: 'Ssh',
      activeLabel : 'Ssh',
      sshList : [],
    }
  },
  mounted: function () {
    if (process.env.NODE_ENV === 'production') {
      this.apiHost = ''
    }
    this.activeLabel = String(store.getStore("set_active_label"))
    if  (this.activeLabel === '') {
      this.activeLabel = 'Ssh'
    }
    this.SshList()
  },
  methods: {
    handleTabClick : function (tab){
      let index = tab.index
      this.activeLabel = tab.props.name
      console.log(tab , this.activeLabel)
      store.setStore("set_active_label", tab.props.name)
      switch (this.activeLabel){
        case 'Ssh':
          this.$refs.ssh.SshList();
          break
        case 'Git':
          this.$refs.git.GitList()
          this.$refs.git.GitGroupList()
          break
        case 'Account':
          this.$refs.account.AccountList()
          this.$refs.account.AccountGroupList()
          break
      }
    },
    SshList : function (){
      let _that = this
      set.SshList(function (response){
        console.log(response)
        if(response.ErrCode === 0){
          _that.sshList = response.Data
        }
      })
    },
    getStore: function (key) {
      return localStorage.getItem(key)
    },
  },
}
</script>

<style scoped>
.set-page-tabs {
  background: #fff;
  border: 1px solid #e8e8e0;
  border-radius: 12px;
  padding: 10px 12px;
}

.set-tabs :deep(.el-tabs__header) {
  margin-bottom: 10px;
}

.set-tabs :deep(.el-tabs__nav-wrap::after) {
  background-color: #e8e8e0;
}

.set-tabs :deep(.el-tabs__item) {
  height: 36px;
  color: #5c6856;
  font-weight: 500;
}

.set-tabs :deep(.el-tabs__item.is-active) {
  color: #4f804f;
}

.set-tabs :deep(.el-tabs__active-bar) {
  background-color: #4f804f;
}

.set-tab-pane {
  padding: 4px;
}

@media (max-width: 768px) {
  .set-page-tabs {
    padding: 8px;
  }
}
</style>
