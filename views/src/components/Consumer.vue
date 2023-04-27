<template>
  <el-card>
    <!--  子操作选项列表-->
    <el-card>
      <h3>
        消费者操作
      </h3>
      <el-select v-model="chooseParentType" @change="changeParentType" placeholder="请选择系统">
        <el-option
          v-for="(value,key) in parentTypeList"
          :key="value.Name"
          :label="value.Title"
          :value="value.Name">
        </el-option>
      </el-select>

      <!--    环境-->
      <el-select v-model="chooseEvnName" @change="changeCode" placeholder="请选择代码环境" v-if="chooseParentType === 'xkf'">
        <el-option
          v-for="(value,key) in codeEnvList" v-if="value.ParentType === chooseParentType"
          :key="value.Name"
          :label="value.NameTitle"
          :value="value.Name">
        </el-option>
      </el-select>
      <el-button type="primary" @click="restartSupervisorAll">重启{{ chooseEvnName }}所有消费者</el-button>
      <el-button type="primary" @click="supervisorStatusList" :loading="btnLoading.supervisorStatusListStatus">查看所有消费者
      </el-button>
      <!--        <el-button type="primary" @click="supervisorStatusList" >更新所有消费者-->
      <!--        </el-button>-->
      <el-input style="width: 400px" autocomplete="off" placeholder="搜索名称/进程名/程序名等" v-model="searchKey"
                @input="searchList"></el-input>
      <!--      <el-row style="margin-top: 10px;">-->
      <!--        <el-tag>ffff</el-tag>-->
      <!--      </el-row>-->
            <br/>
            <el-tag v-for="sortConsumer in useSortConsumerList" closable style="cursor:default;margin:5px;" @click="searchKey = sortConsumer.name;searchList()" @close="delSortConsumer(sortConsumer)">
              {{sortConsumer.showName}}
            </el-tag>
    </el-card>

    <el-row :gutter="24" style="margin-top: 10px">
      <el-col :span="6" v-for="(value,key) in configMap[chooseParentType]" style="margin-top:5px;" v-if="value.show">
        <div class="grid-content bg-purple">
          <el-card class="box-card">
            <div slot="header" class="clearfix">
              <span>{{ value.showName }}</span>
              <el-button type="text" class="button" style="float: right;" @click="editName(value);">修改名称
              </el-button>
            </div>
            <!--            <div class="supervisorCommand" style="overflow:hidden;">-->
            <!--              命令:{{ value.commandS }} <br/>-->
            <!--            </div>-->
            <div class="supervisorCommand" style="overflow:hidden;" v-if="value.running_status">
              <span>{{ value.name }}</span><br/>
              {{ value.running_status }}
            </div>

            <div class="supervisorCommand" style="overflow:hidden;" v-if="!value.running_status">
              未启动
            </div>

            <div class="bottom clearfix">
              <el-button type="text" class="button"
                         @click="ExecType = 'supervisor_restart';exec(value);showResultDialog=true;">重新启动
              </el-button>
              <el-button type="text" class="button"
                         @click="ExecType = 'supervisor_stop';exec(value);showResultDialog=true;">停止
              </el-button>
              <el-button type="text" class="button"
                         @click="ExecType = 'supervisor_config_show';exec(value);showResultDialog=true;">查看配置
              </el-button>

              <!--              <el-button type="text" class="button" @click="execDockerFunc('searchProcess' , value)">搜索溢出进程</el-button>-->
            </div>
          </el-card>
        </div>
      </el-col>
    </el-row>

    <el-dialog :title="supervisorConfigShow.name" :visible.sync="supervisorConfigShow.dialog">
      <el-form :model="supervisorConfigShow">
        <el-form-item label="配置地址" :label-width="30">
          <el-input v-model="supervisorConfigShow.path" autocomplete="off"></el-input>
        </el-form-item>
        <el-form-item label="配置内容" :label-width="30">
          <el-input style="margin-top: 20px;" type="textarea" v-model="supervisorConfigShow.content"
                    rows="15"></el-input>
        </el-form-item>
      </el-form>
      <div slot="footer" class="dialog-footer">
        <el-button @click="supervisorConfigShow.dialog = false">取 消</el-button>
      </div>
    </el-dialog>
    <!--配置或执行结果弹窗-->
    <el-dialog title="内容" :visible.sync="showResultDialog" width="70%;">
      <el-alert
        :title="editNameValue.supervisor_config"
        type="info">
      </el-alert>

      <el-input style="margin-top: 20px;width:100%;" type="textarea" v-model="settingResult" rows="15"></el-input>
      <div slot="footer" class="dialog-footer">
        <el-button @click="showResultDialog = false">取 消</el-button>
        <!--      <el-button type="primary" @click="createCache">确 定</el-button>-->
      </div>
    </el-dialog>

    <el-dialog
      title="输入名称"
      :visible.sync="dialogShowEditName"
      width="30%">
      <el-input style="width: 400px" autocomplete="off" placeholder="输入名称" v-model="inputNameValue"></el-input>
      <span slot="footer" class="dialog-footer">
        <el-button @click="dialogShowEditName = false;">取 消</el-button>
        <el-button type="primary" @click="dialogShowEditName = false;editNameValueFunc()">确 定</el-button>
      </span>
    </el-dialog>
  </el-card>


</template>

<style>
.supervisorCommand {
  padding: 3px;
  font-size: 14px;
}
</style>

<script>
import Vue from "vue";
import {Message} from "element-ui";
import store from "../utils/store";

export default {
  data() {
    return {
      name: "Consumer",
      //接口地址
      apiHost: '',
      //ssh config
      sshConfig: {},
      //选中的环境
      chooseEvnName: "common3-xkf",
      //是否显示所有的消费者
      showAllSupervisor: false,
      showResultDialog: false,
      dialogShowEditName: false,
      inputNameValue: '',
      editNameValue: {},
      //代码环境
      codeEnvList: [],
      //docker
      dockerList: [],
      //存储所有的消费者配置文件
      configMap: {},
      //按钮状态
      btnLoading: {
        supervisorStatusListStatus: false,
        restart: false,
        stop: false,
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
      settingResult: "",
      //docker内执行的命令
      dockerExecCommand: "",
      //历史记录
      useSortConsumerList : [],
      //搜索key
      searchKey: "",
      //消费者配置查看
      supervisorConfigShow: {
        name: "",
        path: "",
        content: "",
        dialog: false,
      },
      supervisorOriginConfList: [],
    }
  },
  mounted: function () {
    this.sshConfig = this.$helperConfig.getXkfDevSshConfig()
    this.wkSshConfig = this.$helperConfig.getWkDevSshConfig()
    this.apiHost = this.$helperConfig.getApiHost()
    this.codeEnvList = this.$helperConfig.getCodeEnvList()
    this.getOriginSupervisorConfig()
    this.dockerList = this.$helperConfig.getDockerList()
    this.refreshUseSortConsumer()
  },
  onload: function () {

  },
  filters: {
    limitTo(value, length) {
      return value.slice(0, length);
    }
  },
  methods: {
    //拿到config 列表
    getOriginSupervisorConfig: function () {
      let _that = this
      let supervisorOriginConfList = []
      let env_config = this.$helperConfig.getCodeEnvConfigByCodeEnvName(this.codeEnvList, this.chooseEvnName)
      if (this.chooseParentType === 'xkf') {
        env_config.SshConfig = _that.sshConfig
      } else {
        env_config.SshConfig = _that.wkSshConfig
      }

      let params = {
        SshConfig: env_config.SshConfig,
        ExecType: 'SupervisorConfList',
        ParentType: _that.chooseParentType,
      }
      //按钮加载状态
      this.$helperApi.ajaxDefault(params, function (response) {
        let tempList = response.Data.split(`\n`)
        for (let i in tempList) {
          supervisorOriginConfList.push(tempList[i].split('---'))
        }
        _that.configMap[_that.chooseParentType] = _that.$helperConfig.getSupervisorConfigList(supervisorOriginConfList, _that.chooseParentType)
        _that.supervisorStatusList()
      })
    },
    execDockerFunc: function (type, value) {
      if (type === 'searchProcess') {
        this.ExecType = 'docker_exec';
        let command = value.command
        command = command.replaceAll('  ', ' ')
        let command_params = command.split(' ')
        if (command_params.length > 1) {
          this.dockerExecCommand = 'ps -aux|grep -i ' + command_params[command_params.length - 1]
        } else {
          this.$helperNotify.error('进程名找不到')
          return
        }
      }
      this.exec(value)
    },
    //选择代码环境
    changeCode: function () {
      this.supervisorStatusList()
    },
    //搜索消费者列表
    searchList: function () {
      for (let i in this.configMap[this.chooseParentType]) {
        if (this.configMap[this.chooseParentType][i].name && this.configMap[this.chooseParentType][i].name.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1) {
          this.configMap[this.chooseParentType][i].show = true
          continue;
        }

        if (this.configMap[this.chooseParentType][i].running_status && this.configMap[this.chooseParentType][i].running_status.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1) {
          this.configMap[this.chooseParentType][i].show = true
          continue;
        }
        if (this.configMap[this.chooseParentType][i].supervisor_config && this.configMap[this.chooseParentType][i].supervisor_config.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1) {
          this.configMap[this.chooseParentType][i].show = true
          continue;
        }
        if (this.configMap[this.chooseParentType][i].supervisor_name && this.configMap[this.chooseParentType][i].supervisor_name.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1) {
          this.configMap[this.chooseParentType][i].show = true
          continue;
        }
        if (this.configMap[this.chooseParentType][i].showName && this.configMap[this.chooseParentType][i].showName.indexOf(this.searchKey.toLowerCase()) !== -1) {
          this.configMap[this.chooseParentType][i].show = true
          continue;
        }
        this.configMap[this.chooseParentType][i].show = false
      }
    },
    //重启所有的消费者
    restartSupervisorAll: function () {
      this.ExecType = 'supervisor_restart_all';
      this.showAllSupervisor = false
      this.exec()
    },
    //查看所有的消费者列表
    supervisorStatusList: function () {
      this.ExecType = 'supervisor_status_list';
      let _that = this
      let env_config = {}
      if (this.chooseParentType === 'xkf') {
        //找到代码配置
        env_config = this.$helperConfig.getCodeEnvConfigByCodeEnvName(this.codeEnvList, this.chooseEvnName)
        env_config.SshConfig = _that.sshConfig
      } else {
        env_config = {}
        env_config.SshConfig = _that.wkSshConfig
      }

      //根据类型判断
      let params = {
        SshConfig: env_config.SshConfig,
        CodePath: env_config.CodePath,
        ParentType: this.chooseParentType,
        ExecType: this.ExecType,
        DockerList: this.dockerList,
        DockerId: this.$helperConfig.getDockerIdByCodeEnvConfig(this.dockerList, env_config),
        DockerCodePath: env_config.DockerCodePath,
        DockerExecCommand: this.dockerExecCommand,
      }
      if (this.chooseParentType === 'xkf') {
        if (params.ExecType === 'supervisor_status_list' && params.CodePath === '') {
          _that.$helperNotify.error('请选择代码环境')
        }
        if (params.ExecType === 'supervisor_status_list' && params.DockerId === ``) {
          _that.$helperNotify.error('代码环境找不到对应的docker')
          return
        }
      }
      //按钮加载状态
      _that.setBtnLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.cancelBtnLoading(params)
        _that.supervisorStatusExplain();
        _that.searchList()
      });
    },
    //改变父类类型
    changeParentType: function () {
      if (this.chooseParentType === 'xkf') {
        this.chooseEvnName = 'common3-xkf'
      } else {
        this.chooseEvnName = '企微'
      }

      this.ExecType = ''
      if (this.configMap[this.chooseParentType]) {
        this.supervisorStatusList()
      } else {
        this.getOriginSupervisorConfig()
      }
      this.refreshUseSortConsumer()
    },
    //执行
    exec: function (param) {
      let _that = this
      let env_config = {}
      if (this.chooseParentType === 'xkf') {
        //找到代码配置
        env_config = this.$helperConfig.getCodeEnvConfigByCodeEnvName(this.codeEnvList, this.chooseEvnName)
        if (env_config === {}) {
          _that.$helperNotify.error("不存在的配置");
          return
        }
        env_config.SshConfig = _that.sshConfig
      } else {
        env_config = {}
        env_config.SshConfig = _that.wkSshConfig
      }

      //根据类型判断
      let params = {
        SshConfig: env_config.SshConfig,
        CodePath: env_config.CodePath,
        ExecType: this.ExecType,
        DockerList: this.dockerList,
        ParentType: this.chooseParentType,
        DockerId: this.$helperConfig.getDockerIdByCodeEnvConfig(this.dockerList, env_config),
        DockerCodePath: env_config.DockerCodePath,
        DockerExecCommand: this.dockerExecCommand,
      }
      if (this.chooseParentType === 'xkf') {
        if (params.ExecType === 'supervisor_restart_all' && params.CodePath === '') {
          _that.$helperNotify.error('请选择代码环境')
          return
        }

      }
      //查看消费者的配置内容
      if (params.ExecType === 'supervisor_config_show') {
        params.SupervisorConfigPath = param.supervisor_config
      } else if (params.ExecType === 'supervisor_restart' || params.ExecType === 'supervisor_stop') {
        params.SupervisorRestartName = param.supervisor_restart_name
      }

      if (params.ExecType === 'supervisor_restart_all' && params.DockerId === ``) {
        _that.$helperNotify.error('代码环境找不到对应的docker')
        return
      }

      if (params.ExecType === 'supervisor_config_show') {
        _that.settingResult = '查询中...';
        _that.editNameValue = param
        _that.addUse(param)
      } else if (params.ExecType === 'supervisor_restart' || params.ExecType === 'supervisor_stop') {
        _that.settingResult = '获取结果中...';
        _that.editNameValue = param
        _that.addUse(param)
      }

      //按钮加载状态
      _that.setBtnLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.cancelBtnLoading(params)
        if (params.ExecType === 'supervisor_restart' || params.ExecType === 'supervisor_stop') {     //查看消费者列表
          _that.supervisorStopRestartExplain(param);
          _that.settingResult = response.Data
        } else if (params.ExecType === 'supervisor_config_show') { //查看supervisor配置
          _that.settingResult = response.Data
        }
      _that.searchList()
      });
    },
    //修改名称
    editName: function (param) {
      this.editNameValue = param
      this.addUse(param)
      this.dialogShowEditName = true
    },
    editNameValueFunc: function () {
      let _that = this
      this.$helperStore.setStore(this.editNameValue.name, this.inputNameValue)
      this.flushConfigList()
      this.refreshUseSortConsumer()
    },
    flushConfigList: function () {
      for (let i in this.configMap[this.chooseParentType]) {
        let showName = store.getStore(this.configMap[this.chooseParentType][i].name)
        if (showName === null || showName === undefined) {
          showName = this.configMap[this.chooseParentType][i].name.split('.')[0]
        }
        this.configMap[this.chooseParentType][i].showName = showName
      }
    },
    //查看消费者配置
    supervisorConfigShowMethod: function (param) {
      // this.supervisorConfigShow.dialog = true
      // this.supervisorConfigShow.path = param.supervisor_config
      // this.supervisorConfigShow.name = param.name
      // this.supervisorConfigShow.content = this.execResult
    },
    //增加了累计使用
    addUse : function (value){
      let cackeKey = this.chooseParentType + 'useSortConsumer'
      let useSortConsumer = this.$helperStore.getStore(cackeKey)
      if(useSortConsumer === null || useSortConsumer === undefined){
        this.$helperStore.setStore(cackeKey , JSON.stringify([{
          'name' : value.name,
          'useNum' : 1,
        }]))
      }else{
        let useSortConsumerList = JSON.parse(useSortConsumer)
        let boolFind = false
        for(let i in useSortConsumerList){
          if(useSortConsumerList[i].name === value.name){
            useSortConsumerList[i].useNum ++
            boolFind = true
            break
          }
        }
        if(!boolFind){
          useSortConsumerList.push({
            'name' : value.name,
            'useNum' : 1,
          })
        }
        this.$helperStore.setStore(cackeKey, JSON.stringify(useSortConsumerList))
      }
      this.refreshUseSortConsumer()
    },
    //刷新排序
    refreshUseSortConsumer : function (){
      let cackeKey = this.chooseParentType + 'useSortConsumer'
      let useSortConsumer = this.$helperStore.getStore(cackeKey)
      if(useSortConsumer === null || useSortConsumer === undefined){
        this.useSortConsumerList = []
      }else{
        this.useSortConsumerList = JSON.parse(useSortConsumer)
      }
      this.useSortConsumerList.sort(function(a, b) {
        return b.key - a.key;
      });
      this.useSortConsumerList = this.useSortConsumerList.slice(0 , 10)
      for (let j in this.useSortConsumerList){
        let showName = this.$helperStore.getStore(this.useSortConsumerList[j].name)
        if(showName === null || showName === undefined){
          showName = this.useSortConsumerList[j].name
        }
        this.useSortConsumerList[j].showName = showName
      }
    },
    delSortConsumer : function (value){
      let cackeKey = this.chooseParentType + 'useSortConsumer'
      let useSortConsumer = this.$helperStore.getStore(cackeKey)
      if(useSortConsumer === null || useSortConsumer === undefined){
        this.useSortConsumerList = []
      }else{
        this.useSortConsumerList = JSON.parse(useSortConsumer)
      }
      let returnList = []
      for (let j in this.useSortConsumerList){
        if(this.useSortConsumerList[j].name !== value.name){
          returnList.push(this.useSortConsumerList[j])
        }
      }
      this.$helperStore.setStore(cackeKey, JSON.stringify(returnList))
      this.refreshUseSortConsumer()
    },
    //分析重启或者停止后的结果
    supervisorStopRestartExplain : function (param){
      let consumerStatusList = this.execResult.split('\n')
      for (let i in consumerStatusList) {
        if(consumerStatusList[i].indexOf('RUNNING') !== -1){
          let runningStatus = consumerStatusList[i].substr(consumerStatusList[i].indexOf('RUNNING'))
          this.getRunningStatus(runningStatus , param.name)
        }

        if(consumerStatusList[i].indexOf('FATAL') !== -1){
          let runningStatus = consumerStatusList[i].substr(consumerStatusList[i].indexOf('FATAL'))
          this.getRunningStatus(runningStatus , param.name)
        }

        if(consumerStatusList[i].indexOf('STOPPED') !== -1){
          let runningStatus = consumerStatusList[i].substr(consumerStatusList[i].indexOf('STOPPED'))
          this.getRunningStatus(runningStatus , param.name)
        }
      }
    },
    getRunningStatus : function (runningStatus , name){
      for (let n in this.configMap[this.chooseParentType]) {
        if(this.configMap[this.chooseParentType][n].name === name){
          this.configMap[this.chooseParentType][n].running_status = runningStatus
          return
        }
      }
    },
    //分析消费者结果
    supervisorStatusExplain: function () {
      //分析结果
      let consumerStatusList = this.execResult.split('\n')
      console.log(consumerStatusList)
      for (let i in consumerStatusList) {
        //根据；分割
        let name_params = consumerStatusList[i].split('    ')
        //循环判断
        let name_params_two = this.filterArray(name_params);
        //获取supervisor进程名
        if (name_params_two.length === 0) {
          continue;
        }
        let name = name_params_two[0]
        let name_params_four = this.filterArray(name.split(':'));
        if (name_params_four.length === 0) {
          continue;
        }
        //给与状态
        for (let n in this.configMap[this.chooseParentType]) {
          if (this.configMap[this.chooseParentType][n].supervisor_name === name_params_four[0]) {
            this.configMap[this.chooseParentType][n].running_status = name_params_two[1]
            //重启名
            if (name_params_four.length === 2) {
              this.configMap[this.chooseParentType][n].supervisor_restart_name = name_params_four[0] + ':'
            } else {
              this.configMap[this.chooseParentType][n].supervisor_restart_name = name_params_four[0]
            }
            this.configMap[this.chooseParentType][n].show = true
            break;
          } else {
            this.configMap[this.chooseParentType][n].show = true;
          }
        }
      }
      for (let k in this.configMap[this.chooseParentType]) {
        if (this.configMap[this.chooseParentType][k].running_status === ``) {
          this.configMap[this.chooseParentType][k].running_status = '未启动';
        }
      }
    },
    setBtnLoading: function (params) {
      if (params.ExecType === 'supervisor_status_list') {
        this.btnLoading.supervisorStatusListStatus = true
      } else if (params.ExecType === 'supervisor_restart') {
        this.btnLoading.restart = true
      } else if (params.ExecType === 'supervisor_stop') {
        this.btnLoading.stop = true
      }
      let _this = this
      let _set_params = params
      setTimeout(function () {
        _this.cancelBtnLoading(_set_params)
      }, 15000)
    },
    cancelBtnLoading: function (params) {
      if (params.ExecType === 'supervisor_status_list') {
        this.btnLoading.supervisorStatusListStatus = false
      } else if (params.ExecType === 'supervisor_restart') {
        this.btnLoading.restart = false
      }
      if (params.ExecType === 'supervisor_stop') {
        this.btnLoading.stop = false
      }
    },
    //过滤数组空数据
    filterArray: function (array) {
      let return_array = [];
      for (let m in array) {
        if (array[m] !== '') {
          return_array.push(array[m])
        }
      }
      return return_array;
    },
  },
}
</script>


