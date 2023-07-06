<template>
  <el-card>
    <el-card>
      <div v-for="(parentTypeValue,k) in parentTypeList">
        <h3 style="display: inline-block;">
          {{ parentTypeValue.Title }}
        </h3>
        <el-row :gutter="20">
          <el-col :span="2" v-for="(value,key) in codeEnvList" style="margin:5px;"
                  v-if="value.ParentType === parentTypeValue.Name">
            <div>
              <el-radio @change="codeChange(value)" size="medium " v-model="chooseEvnName" :label="value.Name">
                {{ value.NameTitle }}
              </el-radio>
            </div>
          </el-col>
        </el-row>

      </div>
      <br/>
      <el-button type="primary" :loading="loadingStatus['pull_branch_origin']" @click="ExecType = 'pull_branch_origin';exec()">拉取最新代码
      </el-button>
      <el-button type="primary" :loading="loadingStatus['git_status']" @click="ExecType = 'git_status';exec()">查看分支变更</el-button>
      <el-input v-if="showChangeBranch" style="width:300px;margin-right:20px;" v-model="BranchName" placeholder="请输入分支名"></el-input>
      <el-button type="primary" :loading="loadingStatus['change_branch']" @click="showChangeBranch = true;ExecType = 'change_branch';exec()">切换分支
      </el-button>
      <el-button type="primary" :loading="loadingStatus['query_current_branch']" @click="ExecType = 'query_current_branch';chooseSshName = chooseEvnName.SshName;exec()">查看分支
      </el-button>


    </el-card>
    <el-input style="margin-top: 20px;" id="resultTextarea" type="textarea" v-model="execResult" rows="25"></el-input>
  </el-card>

</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";

let codeList = require("../config/codeList.json")
let dockerList = require("../config/dockerList.json")
let businessTypeList = require("../config/businessTypeList.json")
export default {
  data() {
    return {
      name: "Consumer",
      //接口地址
      apiHost: '',
      //ssh config
      sshConfig: {},
      wkSshConfig: {},
      //输入框
      showChangeBranch: false,
      //选中的环境
      chooseEvnName: "common3-xkf",
      //代码环境
      codeEnvList: [],
      //docker
      dockerList: [],
      //按钮状态
      btnLoading: {
        exec: false,
        pull: false,
        change: false,
        status: false,
      },
      //操作业务类型
      chooseBusinessType: "git",
      businessTypeList: businessTypeList,
      //操作父类型
      chooseSshName: "xkf",
      parentTypeList: [
        {Title: "小客服（php）", Name: "xkf"},
        {Title: "企微（php）", Name: "wk"},
        {Title: "视频号小店（golang）", Name: "weixin_shop_golang"},
        // {Title: "预发布", Name: "prodTest"},
      ],
      //总的操作类型
      ExecType: "pull_branch_origin",
      //操作类型
      dialogSshConfig: false,
      BranchName: "",  //分支名
      execResult: "",//操作结果
      loadingStatus : {},
    }
  },
  mounted: function () {
    this.apiHost = this.$helperConfig.getApiHost()
    this.sshConfig = this.$helperConfig.getXkfDevSshConfig()
    this.wkSshConfig = this.$helperConfig.getWkDevSshConfig()
    this.codeEnvList = this.$helperConfig.getCodeEnvList()
    this.dockerList = this.$helperConfig.getDockerList()
    this.ExecType = 'query_current_branch'
    this.loadingStatus = this.$helperLoad.getExecTypeStatus()
    this.exec()
  },
  methods: {
    //textarea滚动到最后
    textareaScroll: function () {
      var d = document.getElementById("resultTextarea").scrollHeight;
      document.getElementById("resultTextarea").scrollTop = d;
    },
    //查看日志
    showLog: function (logName) {
      this.ExecType = 'docker_exec'
      let _that = this
      //找到环境配置
      let env_config = this.$helperConfig.getCodeEnvConfigByCodeEnvName(this.codeEnvList , this.chooseEvnName)
      this.dockerExecCommand = 'tail -n 1000 /var/www/' + env_config['DockerCodePath'] + '/' + env_config['LogPath'] + '/' + logName
      this.exec()
    },
    //改变代码环境
    codeChange: function (value) {
      this.ExecType = 'query_current_branch'
      this.chooseSshName = value.SshName
      this.exec()
    },
    //改变业务类型
    changeBusinessType: function () {
      this.chooseEvnName = ""
      this.ExecType = ''
    },
    //执行
    exec: function () {
      let _that = this
      //找到环境配置
      let env_config = this.$helperConfig.getCodeEnvConfigByCodeEnvName(this.codeEnvList , this.chooseEvnName)
      //根据类型判断
      let chooseSshConfig = _that.sshConfig
      if (env_config.SshName === 'wk') {
        chooseSshConfig = _that.wkSshConfig
      } else if (env_config.SshName === "xkf") {
        chooseSshConfig = _that.sshConfig
      }
      let params = {
        SshConfig: chooseSshConfig,
        CodePath: env_config.CodePath,
        BranchName: this.BranchName,
        ExecType: this.ExecType,
        DockerList: this.dockerList,
        ParentType: env_config.ParentType,
        DockerId: _that.$helperConfig.getDockerIdByCodeEnvConfig(this.dockerList , env_config),
        DockerCodePath: env_config.DockerCodePath,
        DockerExecCommand: _that.dockerExecCommand,
      }
      if (params.ExecType === 'change_branch' && params.BranchName === '') {
        return
      }
      if (env_config.ParentType !== 'wk' && params.DockerId === ``) {
        _that.$helperNotify.error('代码环境找不到对应的docker')
        return
      }
      //按钮加载状态
      _that.setLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.cancelLoading(params)
        if (params.ExecType === 'change_branch') {
          _that.showChangeBranch = false
          _that.BranchName = ''
        }
        setTimeout(function () {
          _that.textareaScroll()
        }, 500)
      });
    },
    setLoading : function (params){
      this.loadingStatus[params.ExecType] = true
      let that = this
      setTimeout(function (){
        that.loadingStatus[params.ExecType] = false
      } , 25000)
    },
    cancelLoading : function (params){
      let that = this
      setTimeout(function (){
        that.loadingStatus[params.ExecType] = false
      } , 1000)
    },
  },
}
</script>

<style scoped>

</style>
