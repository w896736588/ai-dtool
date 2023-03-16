<template>
  <el-card>
    <!--      代码环境-->
    <div>
      <h3 style="display: inline-block;">
        {{chooseParentName}} - {{chooseEvnName}} &nbsp;

      </h3>
      <el-select v-model="chooseParentType" @change="changeParentType" placeholder="请选择系统">
        <el-option
          v-for="(value,key) in parentTypeList"
          :key="value.Name"
          :label="value.Title"
          :value="value.Name">
        </el-option>
      </el-select>
      <el-row :gutter="20">
        <el-col :span="2" v-for="(value,key) in codeEnvList" style="margin:5px;" v-if="value.ParentType === chooseParentType">
          <div>
            <el-radio @change="codeChange" size="medium " v-model="chooseEvnName" :label="value.Name">{{value.Name}}</el-radio>
          </div>
        </el-col>
      </el-row>
    </div>
    <br/>
    <el-button type="primary" :loading="btnLoading.pull" @click="ExecType = 'pull_branch_origin';exec()">拉取最新代码</el-button>
    <el-button type="primary" :loading="btnLoading.status" @click="ExecType = 'git_status';exec()">查看分支变更</el-button>
    <el-input v-if="showChangeBranch" style="width:300px;margin-right:20px;" v-model="BranchName" placeholder="请输入分支名"></el-input>
    <el-button type="primary" :loading="btnLoading.change" @click="showChangeBranch = true;ExecType = 'change_branch';exec()" >切换分支</el-button>
<!--    日志操作：-->
<!--    <el-button type="primary" :loading="btnLoading.status" v-if="chooseParentType === 'xkf'" @click="showLog('application.log')"> application.log</el-button>-->
<!--    <el-button type="primary" :loading="btnLoading.status" v-if="chooseParentType === 'xkf'" @click="showLog('default.log')"> default.log</el-button>-->

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
      apiHost: 'http://localhost:7070',
      //ssh config
      sshConfig: {},
      prodTestSshConfig : {},
      //输入框
      showChangeBranch : false,
      //选中的环境
      chooseEvnName: "common3",
      //代码环境
      codeEnvList: codeList,
      //docker
      dockerList: dockerList,
      //按钮状态
      btnLoading : {
        exec : false,
        pull : false,
        change : false,
        status : false,
      },
      //操作业务类型
      chooseBusinessType: "git",
      businessTypeList: businessTypeList,
      //操作父类型
      chooseParentType: "xkf",
      chooseParentName : "小客服（php）",
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
    }
  },
  mounted: function () {
    if(process.env.NODE_ENV === 'production'){
      this.apiHost = '';
    }
    let sshConfig = this.getStore('sshConfig')
    if (sshConfig !== null) {
      this.sshConfig = JSON.parse(sshConfig)
    }
    if(!this.sshConfig || !this.sshConfig.username || this.sshConfig.username === ''){
      this.error("请先配置ssh");
      return
    }
    let prodTestSshConfig = this.getStore('prodTestSshConfig')
    if (prodTestSshConfig !== null) {
      this.prodTestSshConfig = JSON.parse(prodTestSshConfig)
    }
    this.ExecType = 'query_current_branch'
    this.exec()
  },
  methods: {
    //textarea滚动到最后
    textareaScroll : function (){
      var d = document.getElementById("resultTextarea").scrollHeight;
      document.getElementById("resultTextarea").scrollTop = d;
    },
    //查看日志
    showLog : function (logName){
      this.ExecType = 'docker_exec'
      let _that = this
      //找到环境配置
      let env_config = this.getEnvConfig();
      if (env_config === {} || env_config === undefined || env_config === 'undefined' || !env_config) {
        _that.error("不存在的配置");
        return
      }
      this.dockerExecCommand = 'tail -n 1000 /var/www/' + env_config['DockerCodePath'] + '/' + env_config['LogPath'] + '/' + logName
      this.exec()
    },
    //改变代码环境
    codeChange : function (){
      this.ExecType = 'query_current_branch'
      this.exec()
    },
    //改变父类类型
    changeParentType: function () {
      this.chooseEvnName = ''
      this.ExecType = ''
      this.chooseBusinessType = ''
      for(let i in this.parentTypeList){
        if(this.parentTypeList[i].Name === this.chooseParentType){
          this.chooseParentName = this.parentTypeList[i].Title
        }
      }
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
      let env_config = this.getEnvConfig();
      if (env_config === {} || env_config === undefined || env_config === 'undefined' || !env_config) {
        _that.error("不存在的配置");
        return
      }
      if(this.CodePath === ''){
        _that.error("请选择代码环境");
        return
      }
      //根据类型判断
      let params = {
        SshConfig: _that.sshConfig,
        CodePath: env_config.CodePath,
        BranchName: this.BranchName,
        ExecType: this.ExecType,
        WechatKefuAppid: this.chooseWechatKefuAppid,
        DockerList: this.dockerList,
        ParentType : env_config.ParentType,
        DockerId: "",
        DockerCodePath: env_config.DockerCodePath,
        DockerExecCommand : _that.dockerExecCommand,
      }
      if (params.ExecType === 'change_branch' && params.BranchName === '') {
        return
      }

      //需要找到code对应的docker
      for (let j in this.dockerList) {
        if (env_config.DockerName === this.dockerList[j].Name) {
          params.DockerId = this.dockerList[j].Id
        }
      }
      if (env_config.ParentType !== 'wk' && params.DockerId === ``) {
        _that.error('代码环境找不到对应的docker')
        return
      }
      //按钮加载状态
      _that.setBtnLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
        _that.cancelBtnLoading(params)
        if(params.ExecType === 'change_branch'){
          _that.showChangeBranch = false
          _that.BranchName = ''
        }
        setTimeout(function (){
          _that.textareaScroll()
        } , 500)
      });
    },
    getEnvConfig : function (){
      //找到环境配置
      let env_config = {};
      for (let i in this.codeEnvList) {
        if (this.codeEnvList[i].Name === this.chooseEvnName && this.codeEnvList[i].ParentType === this.chooseParentType) {
          env_config = this.codeEnvList[i]
          break
        }
      }
      return env_config;
    },
    setBtnLoading : function (params){
      if(params.ExecType === 'pull_branch_origin'){
        this.btnLoading.pull = true
      }else if (params.ExecType === 'change_branch'){
        this.btnLoading.change = true
      }else if(params.ExecType === 'git_status'){
        this.btnLoading.status = true
      }

      let _this = this
      let _set_params = params
      setTimeout(function (){
        _this.cancelBtnLoading(_set_params)
      } , 15000)
    },
    cancelBtnLoading : function (params){
      if(params.ExecType === 'pull_branch_origin'){
        this.btnLoading.pull = false
      }else if (params.ExecType === 'change_branch'){
        this.btnLoading.change = false
      }else if(params.ExecType === 'git_status'){
        this.btnLoading.status = false
      }
    },
    success: function (msg) {
      // Message.success(msg);
      this.$notify({title: '提示', message: msg, type: 'success' , duration : 1000});
    },
    warning: function (msg) {
      // Message.warning(msg);
      this.$notify({title: '提示', message: msg, type: 'warning' , duration : 1000});
    },
    info: function (msg) {
      // Message.info(msg);
      //this.$notify({title: '提示', message: msg});
      this.$notify({title: '提示', message: msg, type: 'info' , duration : 1000});
    },
    error: function (msg) {
      // Message.error(msg);
      this.$notify({title: '提示', message: msg, type: 'error' , duration : 1000});
    },
    setStore: function (key, value) {
      localStorage.setItem(key, value);
    },
    getStore: function (key) {
      return localStorage.getItem(key);
    }
  },
}
</script>

<style scoped>

</style>
