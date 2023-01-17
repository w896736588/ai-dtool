<template>
  <el-card>
    <el-card>
<!--      代码环境-->
      <div style="margin-top: 10px;">
        <h3>
          代码环境 - {{chooseEvnName}}
          <el-input style="width:500px;display: inline-block;" v-model="chooseWechatKefuAppid" placeholder="请输入微信客服appid或应用id"></el-input>
        </h3>  &nbsp;
        <el-row :gutter="20">
          <el-col v-if="value.ParentType === 'xkf'" :span="2" v-for="(value,key) in codeEnvList" style="margin:5px;">
            <div>
              <el-radio size="medium " v-model="chooseEvnName" :label="value.Name">{{value.Name}}</el-radio>
            </div>
          </el-col>
        </el-row>
      </div>
      <br/>
      <el-button type="primary" :loading="btnLoading.wechatKefuChange" @click="ExecType = 'wechat_kefu_change';exec()">切换到当前选择的环境</el-button>
      <el-button type="primary" :loading="btnLoading.wechatKefuStatus" @click="ExecType = 'wechat_kefu_status';exec()">查看所在环境</el-button>
    </el-card>

    <el-input style="margin-top: 20px;" type="textarea" v-model="execResult" rows="25"></el-input>
  </el-card>

</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";
let codeList = require("../config/codeList.json")
let dockerList = require("../config/dockerList.json")
let wechatKefuList = require("../config/wechatKefuList.json")
export default {
  data() {
    return {
      name: "WechatKefu",
      //接口地址
      apiHost: 'http://localhost:7070',
      //ssh config
      sshConfig: {},
      xkfDevDbConfig : {},
      //选中的环境
      chooseEvnName: "common3",
      //代码环境
      codeEnvList: codeList,
      //docker
      dockerList: dockerList,
      //按钮状态
      btnLoading : {
        wechatKefuStatus : false,
        wechatKefuChange : false,
      },
      //微信客服合集
      chooseWechatKefuName: "",
      chooseWechatKefuAppid: "",
      wechatKefuList: wechatKefuList,
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

    let xkfDevDbConfig = this.getStore('devTestDbConfig')
    if (xkfDevDbConfig !== null) {
      this.xkfDevDbConfig = JSON.parse(xkfDevDbConfig)
    }
  },
  methods: {
    //执行
    exec: function () {
      let _that = this
      //找到环境配置
      let env_config = {};
      for (let i in this.codeEnvList) {
        if (this.codeEnvList[i].Name === this.chooseEvnName) {
          env_config = this.codeEnvList[i]
          break
        }
      }
      if (env_config === {}) {
        _that.error("不存在的配置");
        return
      }
      env_config.SshConfig = _that.sshConfig
      //根据类型判断
      let params = {
        SshConfig: env_config.SshConfig,
        CodePath: env_config.CodePath,
        BranchName: this.BranchName,
        ExecType: this.ExecType,
        WechatKefuAppid: this.chooseWechatKefuAppid,
        DockerList: this.dockerList,
        DockerId: "",
        DockerCodePath: env_config.DockerCodePath,
        xkfDevDbConfig : this.xkfDevDbConfig,
      }
      if (params.ExecType === 'wechat_kefu_status' && params.WechatKefuAppid === '') {
        _that.error('请输入应用id或appid')
        return
      } else if (params.ExecType === 'wechat_kefu_change' && (params.WechatKefuAppid === '' || params.CodePath === '')) {
        _that.error('选择微信客服以及代码环境')
        return
      }

      //如果是切换微信客服 需要找到code对应的docker
      for (let j in this.dockerList) {
        if (env_config.DockerName === this.dockerList[j].Name) {
          params.DockerId = this.dockerList[j].Id
        }
      }
      if (params.ExecType === 'wechat_kefu_change' && params.DockerId === ``) {
        _that.error('代码环境找不到对应的docker')
        return
      }
      //按钮加载状态
      _that.setBtnLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
        _that.cancelBtnLoading(params)
      });
    },
    setBtnLoading : function (params){
      if(params.ExecType === 'wechat_kefu_status'){
        this.btnLoading.wechatKefuStatus = true
      }else if(params.ExecType === 'wechat_kefu_change'){
        this.btnLoading.wechatKefuChange = true
      }
      let _this = this
      let _set_params = params
      setTimeout(function (){
        _this.cancelBtnLoading(_set_params)
      } , 15000)
    },
    cancelBtnLoading : function (params){
      if(params.ExecType === 'wechat_kefu_status'){
        this.btnLoading.wechatKefuStatus = false
      }else if(params.ExecType === 'wechat_kefu_change'){
        this.btnLoading.wechatKefuChange = false
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
