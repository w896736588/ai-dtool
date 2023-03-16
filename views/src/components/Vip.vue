<template>
  <el-card>
    <div>
      <h3 style="display: inline-block;">
        客服系统版本变更
      </h3>
      <el-select v-model="chooseSystemType" @change="queryVipInfo" placeholder="请选择系统">
        <el-option
          v-for="(value,key) in systemTypeList"
          :key="value.name"
          :label="value.name"
          :value="value.level">
        </el-option>
      </el-select>
      <el-select v-model="chooseVipLevel" @change="queryVipInfo" placeholder="请选择版本">
        <el-option
          v-for="(value,key) in vipList"
          :key="value.name"
          :label="value.name"
          :value="value.level">
        </el-option>
      </el-select>

      <el-select v-model="account"  @change="queryVipInfo" placeholder="请选择账号">
        <el-option
          v-for="(value,key) in userNameList"
          :key="value.Name"
          :label="value.Name"
          :value="value.UserName">
        </el-option>
        <el-option key="-1" label="其他" value="-1">
        </el-option>
      </el-select>
      <el-input v-if="chooseVipUserName === '-1'" style="width:300px;margin-right:20px;" v-model="account" placeholder="请输入账号或管理员ID"></el-input>
      <el-input style="width:300px;margin-right:20px;" v-model="expiredDay" placeholder="天数"></el-input>
<!--      <div class="block">-->
<!--        <span class="demonstration">默认</span>-->
<!--        <el-date-picker-->
<!--          v-model="expiredTime"-->
<!--          type="date"-->
<!--          placeholder="选择日期">-->
<!--        </el-date-picker>-->
<!--      </div>-->
      <el-button type="primary" @click="exec()">变更</el-button>
    </div>
    <br/>

    <h3 style="display: inline-block;">
      客服系统登录
    </h3>
    <el-select v-model="chooseUserName" placeholder="请选择小客服环境">
      <el-option
        v-for="(value,key) in userNameList"
        :key="value.Name"
        :label="value.Name"
        :value="value.Name">
      </el-option>
    </el-select>
    <el-select v-model="chooseLoginType" placeholder="请选择登录类型">
      <el-option
        v-for="(value,key) in loginTypeList"
        :key="value.value"
        :label="value.loginName"
        :value="value.value">
      </el-option>
    </el-select>


    <el-button type="primary" @click="login()">登录</el-button>

    <el-input style="margin-top: 20px;" id="resultTextarea" type="textarea" v-model="execResult" rows="25"></el-input>
  </el-card>

</template>

<script>
import Vue from "vue";
import {Message} from "element-ui";
import redisList from "../config/redisList.json";
let vipList = require("../config/vipList.json")
let systemTypeList = require("../config/systemTypeList.json")
let userNameList = require("../config/userName.json")
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
      //过期时间
      expiredTime : '',
      chooseVipUserName : '',
      //账号信息
      account : '',
      //系统类型
      systemTypeList : systemTypeList,
      //账号
      userNameList : userNameList,
      //vip版本
      vipList : vipList,
      //过期时间
      expiredDay : 10,
      //选择的系统
      chooseUserName : "common3",
      chooseLoginType : "1",
      loginTypeList : [
        {
          "loginName" : "主环境",
          "value" : "1",
        },
        {
          "loginName" : "子环境环境",
          "value" : "2",
        },
        {
          "loginName" : "主环境运营后台",
          "value" : "3",
        },
        {
          "loginName" : "子环境运营后台",
          "value" : "4",
        }
      ],
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
        redisConfigList : _that.redisConfigList,
        expiredDay : this.expiredDay,
      }
      if(this.account === '-1' || this.account === ''){
        this.error('请输入或选择账号');
        return
      }
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
      });
    },
    //登录
    login : function (){
      let _that = this
      let loginUrl = ''
      if(this.chooseLoginType === '1' || this.chooseLoginType === '2'){
        loginUrl = '/index/index';
      }else{
        loginUrl = '/XkfOperate/CustomerList';
      }
      let loginHost = ``
      if(this.chooseLoginType === '1' || this.chooseLoginType === '3'){
        for(let i in this.userNameList){
          if(this.userNameList[i].Name === this.chooseUserName){
            loginHost = this.userNameList[i].Host
          }
        }
      }else{
        for(let i in this.userNameList){
          if(this.userNameList[i].Name === this.chooseUserName){
            loginHost = this.userNameList[i].HostChild
          }
        }
      }
      let account = ''
      for(let i in this.userNameList){
        if(this.userNameList[i].Name === this.chooseUserName){
          account = this.userNameList[i].UserName
        }
      }
      let params = {
        Account : account,
        loginUrl : loginUrl,
        loginHost : loginHost,
        SshConfig: _that.sshConfig,
        ExecType: 'login_xkf',
        xkfDevDbConfig : this.xkfDevDbConfig,
        VipLevel : this.chooseVipLevel,
        SystemType : this.chooseSystemType,
        redisConfigList : _that.redisConfigList,
      }
      console.log(this.chooseLoginType)
      if(this.chooseLoginType === '3' || this.chooseLoginType === '4'){
        params.Account = '2@163.com'
      }
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
        window.open(response.Data,'_blank');
      });
    },
    queryVipInfo : function (){
      let _that = this
      //根据类型判断
      let params = {
        SshConfig: _that.sshConfig,
        ExecType: 'query_vip_info',
        xkfDevDbConfig : this.xkfDevDbConfig,
        Account : this.account,
        VipLevel : this.chooseVipLevel,
        SystemType : this.chooseSystemType,
        redisConfigList : _that.redisConfigList,
        expiredDay : this.expiredDay,
      }
      if(this.account === '-1' || this.account === '' || this.chooseSystemType === '' || this.chooseSystemType === -1){
        return
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
