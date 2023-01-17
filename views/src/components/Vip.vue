<template>
  <el-card>
    <div>
      <h3 style="display: inline-block;">
        客服系统版本变更
      </h3>
      <el-select v-model="chooseSystemType" placeholder="请选择系统">
        <el-option
          v-for="(value,key) in systemTypeList"
          :key="value.name"
          :label="value.name"
          :value="value.level">
        </el-option>
      </el-select>
      <el-select v-model="chooseVipLevel" placeholder="请选择版本">
        <el-option
          v-for="(value,key) in vipList"
          :key="value.name"
          :label="value.name"
          :value="value.level">
        </el-option>
      </el-select>
      <el-input style="width:300px;margin-right:20px;" v-model="account" placeholder="请输入账号或管理员ID"></el-input>
      <el-button type="primary" @click="exec()">变更</el-button>
    </div>
    <br/>

    <el-input style="margin-top: 20px;" id="resultTextarea" type="textarea" v-model="execResult" rows="25"></el-input>
  </el-card>

</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";
import redisList from "../config/redisList.json";
let vipList = require("../config/vipList.json")
let systemTypeList = require("../config/systemTypeList.json")
export default {
  data() {
    return {
      name: "Vip",
      //接口地址
      apiHost: 'http://localhost:7070',
      //ssh config
      xkfDevDbConfig : {},
      sshConfig: {},
      prodTestSshConfig : {},
      //选中的vip版本
      chooseVipLevel : -1,
      chooseSystemType : -1,
      //账号信息
      account : '',
      //系统类型
      systemTypeList : systemTypeList,
      //vip版本
      vipList : vipList,
      //总的操作类型
      ExecType: "",
      execResult: "",//操作结果
      redisConfigList : [],
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
    let xkfDevDbConfig = this.getStore('devTestDbConfig')
    if (xkfDevDbConfig !== null) {
      this.xkfDevDbConfig = JSON.parse(xkfDevDbConfig)
    }
    //增加uniKey
    for( let i in redisList){
      redisList[i].UniKey = redisList[i].Name
    }
    this.redisConfigList = redisList
  },
  methods: {
    //执行
    exec: function () {
      let _that = this
      //根据类型判断
      let params = {
        SshConfig: _that.sshConfig,
        ExecType: 'change_vip_type',
        xkfDevDbConfig : this.xkfDevDbConfig,
        Account : this.account,
        VipLevel : this.chooseVipLevel,
        SystemType : this.chooseSystemType,
        redisConfigList : _that.redisConfigList
      }
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
      });
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
