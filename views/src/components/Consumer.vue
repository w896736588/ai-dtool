<template>
  <el-card>
    <el-select v-model="chooseParentType" @change="changeParentType" placeholder="请选择系统">
      <el-option
        v-for="(value,key) in parentTypeList"
        :key="value.Name"
        :label="value.Title"
        :value="value.Name">
      </el-option>
    </el-select>

    <!--    环境-->
    <el-select v-model="chooseEvnName" @change="changeCode" placeholder="请选择代码环境">
      <el-option
        v-for="(value,key) in codeEnvList" v-if="value.ParentType === chooseParentType"
        :key="value.Name"
        :label="value.Name"
        :value="value.Name">
      </el-option>
    </el-select>


    <!--  子操作选项列表-->
    <el-card style="margin-top: 20px;">
      消费者操作列表<br/><br/>
      <el-button type="primary" @click="restartSupervisorAll">重启{{chooseEvnName}}所有消费者</el-button>
      <el-button type="primary" @click="showSupervisorList" :loading="btnLoading.supervisorStatusListStatus">查看所有消费者</el-button>
      <el-input style="width: 400px" autocomplete="off" placeholder="搜索名称" v-model="searchKey" @input="searchList"></el-input>
    </el-card>

    <el-row :gutter="24" style="margin-top: 10px" >
      <el-col :span="13">
        <el-col :span="12" v-for="(value,key) in supervisorConfigList" style="margin-top:5px;" v-if="value.show">
          <div class="grid-content bg-purple">
            <el-card class="box-card">
              <div slot="header" class="clearfix">
                <span>{{value.name}}</span>
                <el-button style="float: right; padding: 3px 0" type="text" @click="toTop(value)">置顶</el-button>
              </div>
              <div class="supervisorCommand" style="overflow:hidden;">
                命令:{{ value.commandS }} <br/>
              </div>
              <div class="supervisorCommand" style="overflow:hidden;">
                状态：{{value.running_status}}
              </div>

              <div class="bottom clearfix">
                <el-button type="text" class="button" @click="ExecType = 'supervisor_restart';exec(value)">重新启动</el-button>
                <el-button type="text" class="button" @click="ExecType = 'supervisor_stop';exec(value)">停止</el-button>
                <el-button type="text" class="button" @click="ExecType = 'supervisor_config_show';exec(value)">查看配置</el-button>
                <el-button type="text" class="button" @click="execDockerFunc('searchProcess' , value)">搜索溢出进程</el-button>
              </div>
            </el-card>
          </div>
        </el-col>
      </el-col>
      <el-col :span="9">
        <el-input style="margin-top: 20px;width:600px;" type="textarea" v-model="execResult" rows="25" ></el-input>
      </el-col>


    </el-row>

    <el-dialog :title="supervisorConfigShow.name" :visible.sync="supervisorConfigShow.dialog">
      <el-form :model="supervisorConfigShow">
        <el-form-item label="配置地址" :label-width="30">
          <el-input v-model="supervisorConfigShow.path" autocomplete="off"></el-input>
        </el-form-item>
        <el-form-item label="配置内容" :label-width="30">
          <el-input style="margin-top: 20px;" type="textarea" v-model="supervisorConfigShow.content" rows="15"></el-input>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="supervisorConfigShow.dialog = false">取 消</el-button>
      </div>
    </el-dialog>
  </el-card>


</template>

<style>
.supervisorCommand{
  padding : 3px;
  font-size:14px;
}
</style>

<script>
import Vue from "vue";
import {Message} from "element-ui";
let codeList = require("../config/codeList.json")
let dockerList = require("../config/dockerList.json")
let supervisorConfigList = require("../config/supervisorConfig.json")
export default {
  data() {
    return {
      name: "Consumer",
      //接口地址
      apiHost: 'http://localhost:7070',
      //ssh config
      sshConfig: {
        username: "",
        password: "",
        host: "121.40.109.241",
        port: "22",
      },
      //选中的环境
      chooseEvnName: "common3",
      //是否显示所有的消费者
      showAllSupervisor : false,
      //代码环境
      codeEnvList: codeList,
      //docker
      dockerList: dockerList,
      //消费者列表
      supervisorConfigList : [],
      //按钮状态
      btnLoading : {
        supervisorStatusListStatus : false,
        restart : false,
        stop : false,
      },
      //操作父类型
      chooseParentType: "xkf",
      parentTypeList: [
        {Title: "小客服", Name: "xkf"},
        {Title: "企微", Name: "wk"},
      ],
      //总的操作类型
      ExecType: "query_current_branch",
      //操作类型
      dialogSshConfig: false,
      BranchName: "",  //分支名
      execResult: "",//操作结果
      //docker内执行的命令
      dockerExecCommand : "",
      //搜索key
      searchKey : "",
      //消费者配置查看
      supervisorConfigShow : {
        name : "",
        path : "",
        content : "",
        dialog : false,
      }
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
    //初始化命令
    let sliceLength = 30
    for(let i in supervisorConfigList){
      let command = supervisorConfigList[i].command
      supervisorConfigList[i].commandS = '...' + command.substr(command.length - sliceLength , sliceLength)
    }
    this.supervisorConfigList = supervisorConfigList
    //拿到所有消费者
    this.showSupervisorList()
  },
  onload: function(){
    this.initSort()
  },
  methods: {
    execDockerFunc : function (type , value){
      if(type === 'searchProcess'){
        this.ExecType = 'docker_exec';
        let command = value.command
        command = command.replaceAll('  ' , ' ')
        let command_params = command.split(' ')
        if(command_params.length > 1){
          this.dockerExecCommand = 'ps -aux|grep -i ' + command_params[command_params.length - 1]
        }else{
          this.error('进程名找不到')
          return
        }
      }
      this.exec(value)
    },
    //选择代码环境
    changeCode : function (){
      this.showSupervisorList()
    },
    //搜索消费者列表
    searchList : function (){
      for(let i in supervisorConfigList){
        if(supervisorConfigList[i].command.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1){
          supervisorConfigList[i].show = true
          continue;
        }
        if(supervisorConfigList[i].commandS.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1){
          supervisorConfigList[i].show = true
          continue;
        }
        if(supervisorConfigList[i].name.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1){
          supervisorConfigList[i].show = true
          continue;
        }
        if(supervisorConfigList[i].running_status.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1){
          supervisorConfigList[i].show = true
          continue;
        }
        if(supervisorConfigList[i].supervisor_config.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1){
          supervisorConfigList[i].show = true
          continue;
        }
        if(supervisorConfigList[i].supervisor_name.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1){
          supervisorConfigList[i].show = true
          continue;
        }
        supervisorConfigList[i].show = false
      }
    },
    //重启所有的消费者
    restartSupervisorAll : function (){
      this.ExecType = 'supervisor_restart_all';
      this.showAllSupervisor = false
      this.exec()
    },
    //查看所有的消费者列表
    showSupervisorList : function (){
      this.ExecType = 'supervisor_status_list';
      this.exec()
    },
    //改变父类类型
    changeParentType: function () {
      this.chooseEvnName = '配置'
      this.ExecType = ''
      this.chooseBusinessType = ''
    },
    //执行
    exec: function (param) {
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
        DockerExecCommand : this.dockerExecCommand,
      }
      if (params.ExecType === 'supervisor_restart_all' && params.CodePath === '') {
        _that.error('请选择代码环境')
        return
      } else if(params.ExecType === 'supervisor_status_list' && params.CodePath === ''){
        _that.error('请选择代码环境')
      }
      //查看消费者的配置内容
      if(params.ExecType === 'supervisor_config_show'){
        params.SupervisorConfigPath = param.supervisor_config
      }else if(params.ExecType === 'supervisor_restart' || params.ExecType === 'supervisor_stop'){
        params.SupervisorRestartName = param.supervisor_restart_name
      }

      //如果是切换微信客服 需要找到code对应的docker
      for (let j in this.dockerList) {
        if (env_config.DockerName === this.dockerList[j].Name) {
          params.DockerId = this.dockerList[j].Id
        }
      }
      if (params.ExecType === 'supervisor_restart_all' && params.DockerId === ``) {
        _that.error('代码环境找不到对应的docker')
        return
      }
      if (params.ExecType === 'supervisor_status_list' && params.DockerId === ``) {
        _that.error('代码环境找不到对应的docker')
        return
      }
      //按钮加载状态
      _that.setBtnLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.success('成功');
        _that.execResult = response.Data
        _that.cancelBtnLoading(params)
        if(params.ExecType === 'supervisor_status_list' || params.ExecType === 'supervisor_restart' || params.ExecType === 'supervisor_stop'){     //查看消费者列表
          _that.supervisorStatusExplain();
        }else if(params.ExecType === 'supervisor_config_show'){ //查看supervisor配置
          _that.supervisorConfigShowMethod(param);
        }
      });
    },
    //置顶
    toTop : function (param){
      this.sortConsumerList(param.supervisor_name)
    },
    //消费者排序 按照最后使用时间排序
    sortConsumerList : function (supervisor_name){
      console.log('点击了' + supervisor_name)
      let currentTime = parseInt(Date.parse(new Date())/1000)
      console.log('当前时间戳 ' + currentTime)
      this.setStore('supervisor_' + supervisor_name , currentTime)
      this.initSort()
    },
    initSort : function (){
      let currentTime = parseInt(Date.parse(new Date())/1000)
      for(let i in supervisorConfigList){
        let supervisorName = supervisorConfigList[i].supervisor_name
        let sortTime = this.getStore('supervisor_' + supervisorName)
        if(sortTime === null || sortTime === undefined){
          this.setStore('supervisor_' + supervisorName , currentTime)
          supervisorConfigList[i].sortTime = currentTime - 99999999
        }else{
          supervisorConfigList[i].sortTime = sortTime
        }
      }
      supervisorConfigList.sort(function (a,b){
        return b.sortTime - a.sortTime
      })
      console.log(supervisorConfigList)
    },
    //查看消费者配置
    supervisorConfigShowMethod : function (param){
      // this.supervisorConfigShow.dialog = true
      // this.supervisorConfigShow.path = param.supervisor_config
      // this.supervisorConfigShow.name = param.name
      // this.supervisorConfigShow.content = this.execResult
    },
    //分析消费者结果
    supervisorStatusExplain : function (){
      //分析结果
      let consumerStatusList = this.execResult.split('\n')
      for(let i in consumerStatusList){
        //根据；分割
        let name_params = consumerStatusList[i].split('    ')
        //循环判断
        let name_params_two = this.filterArray(name_params);
        //获取supervisor进程名
        if(name_params_two.length === 0){
          continue;
        }
        let name = name_params_two[0]
        let name_params_four = this.filterArray(name.split(':'));
        if(name_params_four.length === 0){
          continue;
        }
        //给与状态
        for(let n in this.supervisorConfigList){
          if(this.supervisorConfigList[n].supervisor_name === name_params_four[0]){
            this.supervisorConfigList[n].running_status = name_params_two[1]
            //重启名
            if(name_params_four.length === 2){
              this.supervisorConfigList[n].supervisor_restart_name = name_params_four[0] + ':'
            }else{
              this.supervisorConfigList[n].supervisor_restart_name = name_params_four[0]
            }
            this.supervisorConfigList[n].show = true
          }
        }
      }
      this.initSort();
    },
    setBtnLoading : function (params){
      if(params.ExecType === 'supervisor_status_list'){
        this.btnLoading.supervisorStatusListStatus = true
      }else if(params.ExecType === 'supervisor_restart'){
        this.btnLoading.restart = true
      }else if(params.ExecType === 'supervisor_stop'){
        this.btnLoading.stop = true
      }
      let _this = this
      let _set_params = params
      setTimeout(function (){
        _this.cancelBtnLoading(_set_params)
      } , 15000)
    },
    cancelBtnLoading : function (params){
      if(params.ExecType === 'supervisor_status_list'){
        this.btnLoading.supervisorStatusListStatus = false
      }else if(params.ExecType === 'supervisor_restart'){
        this.btnLoading.restart = false
      }if(params.ExecType === 'supervisor_stop'){
        this.btnLoading.stop = false
      }
    },
    //过滤数组空数据
    filterArray : function (array){
      let return_array = [];
      for(let m in array){
        if(array[m] !== ''){
          return_array.push(array[m])
        }
      }
      return return_array;
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


