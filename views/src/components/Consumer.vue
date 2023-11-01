<template>
  <div>
    <!--  子操作选项列表-->
    <el-card>
      <!--    环境-->
      <el-select v-model="chooseConsumerName" @change="changeCode" placeholder="请选择环境" >
        <el-option
          v-for="(value,key) in consumerConfigList"
          :key="value.Name"
          :label="value.NameTitle"
          :value="value.Name">
        </el-option>
      </el-select>
      <el-button type="primary" :loading="loadingStatus['supervisor_restart_all']" @click="restartSupervisorAll">重启所有</el-button>
      <el-button type="primary" :loading="loadingStatus['supervisor_status_list']" @click="supervisorStatusList">查看所有</el-button>
<!--      <el-tooltip class="item" effect="dark" content="关闭不常用消费者,可降低20%内存占用" placement="top">-->
<!--        <el-button type="primary" :loading="loadingStatus['reduce_memory']" @click="reduce_memory">降低内存</el-button>-->
<!--      </el-tooltip>-->
      <el-tooltip class="item" effect="dark" content="停止,可降低docker内存占用" placement="top">
        <el-button type="primary" :loading="loadingStatus['stopListConsumer']" @click="stopListConsumer">停止以下{{searchNum}}个</el-button>-->
      </el-tooltip>

      <!--        <el-button type="primary" @click="supervisorStatusList" >更新所有消费者-->
      <!--        </el-button>-->
      <el-input style="width: 400px" autocomplete="off" placeholder="搜索名称/进程名/程序名等" v-model="searchKey"
                @input="searchList"></el-input>
      <!--      <el-row style="margin-top: 10px;">-->
      <!--        <el-tag>ffff</el-tag>-->
      <!--      </el-row>-->
            <br/> <br/>
            <el-tag v-for="sortConsumer in useSortConsumerList" closable style="cursor:default;margin:5px;" @click="searchKey = sortConsumer.name;searchList()" @close="delSortConsumer(sortConsumer)">
              {{sortConsumer.showName}}
            </el-tag>
    </el-card>

    <el-row :gutter="24" style="margin-top: 10px">
      <el-col :span="6" v-for="(value,key) in configMap" style="margin-top:5px;" v-if="value.show">
        <div class="grid-content bg-purple">
          <el-card class="box-card">
            <div slot="header" class="clearfix">
              <span>{{ value.showName }}</span>
              <el-button type="text" class="button" style="float: right;margin-top: -8px;" @click="editName(value);">修改名称
              </el-button>
            </div>
            <!--            <div class="supervisorCommand" style="overflow:hidden;">-->
            <!--              命令:{{ value.commandS }} <br/>-->
            <!--            </div>-->
            <div class="supervisorCommand" style="overflow:hidden;" v-if="value.running_status">
              <span>{{ value.name }}</span><br/>
              <span>{{ value.running_status.replace(/^\s*|\s*$/g,"").substr( 0 ,40) }}</span><br/>
              process num {{value.processNum}}
            </div>

            <div class="supervisorCommand" style="overflow:hidden;" v-if="!value.running_status">
              未启动
            </div>

            <div class="bottom clearfix">
              <el-button type="text" class="button" size="small"
                         @click="restart(value);">重新启动
              </el-button>
              <el-button type="text" class="button" size="small"
                         @click="stop(value)">停止
              </el-button>
              <el-button type="text" class="button" size="small"
                         @click="configShow(value);">查看配置
              </el-button>
              <el-button type="text" class="button" size="small" disabled
                         @click="showInteractionFunc(value)">后台运行
              </el-button>

              <el-button type="text" class="button" size="small" disabled
                         @click="execDockerFunc('searchProcess' , value)">孤儿进程</el-button>
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

    <Interaction ref="Interaction" :visible="showInteraction" :title="showInteractionTitle" :sshConfig="showInteractionSshConfig" @before-close="showInteractionBeforeClose"></Interaction>
</div>

</template>

<style>
.supervisorCommand {
  padding: 3px;
  font-size: 14px;
}
</style>

<script>
import Vue from "vue";
import store from "../utils/store";
import Interaction from "./Interaction"
import consumer from "../utils/api/consumer.js"
import base from "../utils/api/base.js"
import mod from "../utils/api/module.js"

export default {
  components : {
    Interaction,
  },
  data() {
    return {
      name: "Consumer",
      //接口地址
      apiHost: '',
      //ssh config
      sshConfig: {},
      //选中的环境
      chooseConsumerName: "xkf_common3",
      //是否显示所有的消费者
      showAllSupervisor: false,
      showResultDialog: false,
      dialogShowEditName: false,
      inputNameValue: '',
      editNameValue: {},
      searchNum : 0,
      //消费者环境
      consumerConfigList: [],
      //docker
      dockerList: [],
      //存储所有的消费者配置文件
      configMap: {},
      //总的操作类型
      ExecType: "query_current_branch",
      //操作类型
      dialogSshConfig: false,
      BranchName: "",  //分支名
      execResult: "",//操作结果
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
      //终端
      showInteraction : false,
      showInteractionTitle : "",
      showInteractionSshConfig : {},
      loadingStatus : {},
    }
  },
  mounted: function () {
    let _that = this
    setTimeout(function () {
      _that.sshConfig = _that.$helperConfig.getXkfDevSshConfig()
      _that.wkSshConfig = _that.$helperConfig.getWkDevSshConfig()
      _that.apiHost = _that.$helperConfig.getApiHost()
      _that.consumerConfigList = mod.GetConsumerConfigList()
      _that.getOriginSupervisorConfig()
      _that.dockerList = _that.$helperConfig.getDockerList()
      _that.loadingStatus = _that.$helperLoad.getExecTypeStatus()
      _that.refreshUseSortConsumer()
    } , 2000)

  },
  onload: function () {

  },
  filters: {
    limitTo(value, length) {
      return value.slice(0, length);
    },
    substr(value, length){
      return value.substr( 0, length);
    }
  },
  methods: {
    //降低内存
    reduce_memory : function (){
      this.searchKey = 'STOPPED';
      this.searchList()
      this.ExecType = 'supervisor_stop'
      let consumerNameList = this.$helperConfig.getReduceMemoryConsumerName()
      for(let key in this.configMap){
        let boolFind = false
        for(let j in consumerNameList){
          if(this.configMap[key].name === consumerNameList[j]){
            boolFind = true
            break
          }
        }
        if(boolFind){
          this.exec(this.configMap[key]);
        }
      }
    },
    restartAll : function (){
      let _that = this
      let consumerConfig = consumer.GetConsumerConfigByName(this.chooseConsumerName)
      consumer.ConsumerRestartAll(consumerConfig.DockerName , consumerConfig.SshName , function (response) {
        let tempList = response.Data.split(`\n`)
        let supervisorOriginConfList = []
        for (let i in tempList) {
          supervisorOriginConfList.push(tempList[i].split('---'))
        }
        _that.configMap = _that.$helperConfig.getSupervisorConfigList(supervisorOriginConfList, consumerConfig.SshName)
        console.log('configMap' , _that.configMap)
        _that.supervisorStatusList()
      })
    },
    restart : function (value){
      let _that = this
      let consumerConfig = consumer.GetConsumerConfigByName(this.chooseConsumerName)
      _that.addUse(value)
      consumer.ConsumerRestart(consumerConfig.DockerName , consumerConfig.SshName , value.supervisor_name , function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.supervisorStopRestartExplain(value);
        _that.searchList()
      })
    },
    stop : function (value){
      let _that = this
      let consumerConfig = consumer.GetConsumerConfigByName(this.chooseConsumerName)
      _that.addUse(value)
      consumer.ConsumerStop(consumerConfig.DockerName , consumerConfig.SshName , value.supervisor_name , function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.supervisorStopRestartExplain(value);
        _that.searchList()
      })
    },
    configShow : function (value){
      let _that = this
      let consumerConfig = consumer.GetConsumerConfigByName(this.chooseConsumerName)
      _that.addUse(value)
      console.log(value)
      consumer.ConsumerConfigShow(consumerConfig.DockerName , consumerConfig.SshName , value.supervisor_config , function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.supervisorStopRestartExplain(value);
        _that.searchList()
      })
    },
    stopAll : function (){

    },
    //停止列表下面的消费者
    stopListConsumer : function (){
      if(this.searchKey === ''){
        this.stopAll()
        return
      }
      for(let i in this.configMap){
        if(this.configMap[i].show === true){
          this.stop(this.configMap[i])
        }
      }
    },
    showInteractionFunc : function (value){
      // this.showInteractionTitle = value.name;
      // this.showInteractionSshConfig = this.sshConfig
      // this.showInteraction = true
      // this.$refs.Interaction.createShell4()
    },
    //拿到config 列表
    getOriginSupervisorConfig: function () {
      let _that = this
      let consumerConfig = consumer.GetConsumerConfigByName(this.chooseConsumerName)
      consumer.ConsumerConfigList(consumerConfig.DockerName , consumerConfig.SshName , function (response) {
        let tempList = response.Data.split(`\n`)
        let supervisorOriginConfList = []
        for (let i in tempList) {
          supervisorOriginConfList.push(tempList[i].split('---'))
        }
        _that.configMap = _that.$helperConfig.getSupervisorConfigList(supervisorOriginConfList, consumerConfig.SshName)
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
      let searchNum = 0
      for (let i in this.configMap) {
        if (this.configMap[i].name && this.configMap[i].name.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1) {
          this.configMap[i].show = true
          searchNum++;
          continue;
        }

        if (this.configMap[i].running_status && this.configMap[i].running_status.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1) {
          this.configMap[i].show = true
          searchNum++;
          continue;
        }
        if (this.configMap[i].supervisor_config && this.configMap[i].supervisor_config.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1) {
          this.configMap[i].show = true
          searchNum++;
          continue;
        }
        if (this.configMap[i].supervisor_name && this.configMap[i].supervisor_name.toLowerCase().indexOf(this.searchKey.toLowerCase()) !== -1) {
          this.configMap[i].show = true
          searchNum++;
          continue;
        }
        if (this.configMap[i].showName && this.configMap[i].showName.indexOf(this.searchKey.toLowerCase()) !== -1) {
          this.configMap[i].show = true
          searchNum++;
          continue;
        }
        this.configMap[i].show = false
      }
      this.searchNum = searchNum
    },
    //重启所有的消费者
    restartSupervisorAll: function () {
      let _that = this
      let consumerConfig = consumer.GetConsumerConfigByName(this.chooseConsumerName)
      consumer.ConsumerRestartAll(consumerConfig.DockerName , consumerConfig.SshName , function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.supervisorStatusList();
        _that.searchList()
      })
    },
    //查看所有的消费者列表
    supervisorStatusList: function () {
      let _that = this
      let consumerConfig = consumer.GetConsumerConfigByName(this.chooseConsumerName)
      consumer.ConsumerStatusList(consumerConfig.DockerName , consumerConfig.SshName , function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.supervisorStatusExplain();
        _that.searchList()
        _that.$parent.$parent.showTerminal(base.GetUnikey() , consumerConfig.SshName)
      })
    },
    //执行
    exec: function (param) {
      let _that = this
      let env_config = {}
      //找到代码配置
      env_config = this.$helperConfig.getCodeEnvConfigByCodeEnvName(this.consumerConfigList, this.chooseConsumerName)
      if (env_config === {}) {
        _that.$helperNotify.error("不存在的配置");
        _that.cancelLoading(params)
        return
      }
      let dockerId = this.$helperConfig.getDockerIdByCodeEnvConfig(this.dockerList, env_config)
      //根据dockerId获取wk
      for(let dockerKey in this.dockerList){
        if(this.dockerList[dockerKey].Id === dockerId){
          if(this.dockerList[dockerKey].SshName === 'wk'){
            env_config.SshConfig = _that.wkSshConfig
          }else{
            env_config.SshConfig = _that.sshConfig
          }
        }
      }

      //根据类型判断
      let params = {
        SshConfig: env_config.SshConfig,
        CodePath: env_config.CodePath,
        ExecType: this.ExecType,
        DockerList: this.dockerList,
        DockerId: dockerId,
        DockerCodePath: env_config.DockerCodePath,
        DockerExecCommand: this.dockerExecCommand,
      }
      if (params.ExecType === 'supervisor_restart_all' && params.CodePath === '') {
        _that.$helperNotify.error('请选择代码环境')
        _that.cancelLoading(params)
        return
      }

      //查看消费者的配置内容
      if (params.ExecType === 'supervisor_config_show') {
        params.SupervisorConfigPath = param.supervisor_config
      } else if (params.ExecType === 'supervisor_restart' || params.ExecType === 'supervisor_stop') {
        params.SupervisorRestartName = param.supervisor_restart_name
      }
      console.log(param)

      if (params.ExecType === 'supervisor_restart_all' && params.DockerId === ``) {
        _that.$helperNotify.error('代码环境找不到对应的docker')
        _that.cancelLoading(params)
        return
      }

      if (params.ExecType === 'supervisor_config_show') {
        _that.editNameValue = param
        _that.addUse(param)
      } else if (params.ExecType === 'supervisor_restart' || params.ExecType === 'supervisor_stop') {
        _that.editNameValue = param
        _that.addUse(param)
      }

      //按钮加载状态
      _that.setLoading(params)
      Vue.axios.post(this.apiHost + '/api/shell/exec', params).then(function (response) {
        _that.$helperNotify.success('成功');
        _that.execResult = response.Data
        _that.cancelLoading(params)
        if (params.ExecType === 'supervisor_restart' || params.ExecType === 'supervisor_stop') {     //查看消费者列表
          _that.supervisorStopRestartExplain(param);
        } else if (params.ExecType === 'supervisor_config_show') { //查看supervisor配置
        }else if (params.ExecType === 'supervisor_restart_all'){
          _that.supervisorStatusExplain();
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
      for (let i in this.configMap) {
        let showName = store.getStore(this.configMap[i].name)
        if (showName === null || showName === undefined) {
          showName = this.configMap[i].name.split('.')[0]
        }
        this.configMap[i].showName = showName
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
      let cackeKey = 'useSortConsumer'
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
      let cackeKey = 'useSortConsumer'
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
      let cackeKey = 'useSortConsumer'
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
        if(consumerStatusList[i] === ''){
          continue
        }
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
      for (let n in this.configMap) {
        if(this.configMap[n].name === name){
          this.configMap[n].running_status = runningStatus
          return
        }
      }
    },
    //分析消费者结果
    supervisorStatusExplain: function () {
      //重置某些参数
      for (let n in this.configMap) {
        this.configMap[n].processNum = 0
      }
      //分析结果
      let consumerStatusList = this.execResult.split('\n')
      console.log(consumerStatusList)
      for (let i in consumerStatusList) {
        if(consumerStatusList[i] === ''){
          continue
        }
        //根据；分割
        let name_params = []
        name_params.push(consumerStatusList[i].match(/^[^\s]+/g)[0])
        console.log(consumerStatusList[i] , consumerStatusList[i].match(/^[^\s]+/g)[0])
        name_params.push(consumerStatusList[i].replace(name_params[0] , ''))
        //循环判断
        let name_params_two = this.filterArray(name_params);
        console.log('name_params_two ' , name_params_two)
        //获取supervisor进程名
        if (name_params_two.length === 0) {
          continue;
        }
        let name = name_params_two[0]
        let name_params_four = this.filterArray(name.split(':'));
        if (name_params_four.length === 0) {
          continue;
        }
        console.log('name_params_four ' , name_params_four)
        //给与状态
        for (let n in this.configMap) {
          if (this.configMap[n].supervisor_name === name_params_four[0]) {

            this.configMap[n].running_status = name_params_two[1]
            //重启名
            if (name_params_four.length === 2) {
              this.configMap[n].supervisor_restart_name = name_params_four[0] + ':'
            } else {
              this.configMap[n].supervisor_restart_name = name_params_four[0]
            }
            this.configMap[n].show = true
            this.configMap[n].processNum++;
            break;
          } else {
            this.configMap[n].show = true;
          }
        }
      }
      for (let k in this.configMap) {
        if (this.configMap[k].running_status === ``) {
          this.configMap[k].running_status = '未启动';
        }
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
    showInteractionBeforeClose : function (done){
      this.showInteraction = false
    },
    setLoading : function (params){
      this.loadingStatus[params.ExecType] = true
      let that = this
      setTimeout(function (){
        that.loadingStatus[params.ExecType] = false
      } , 25000)
    },
    cancelLoading : function (params){
      this.loadingStatus[params.ExecType] = false
    },
  },

}
</script>


