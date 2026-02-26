<template>
  <!--  子操作选项列表-->
  <div style="text-align: center;">
    <!--    环境-->
    <el-select v-model="chooseSupervisorId" placeholder="请选择环境" @change="changeSupervisor" style="width:300px;">
      <el-option v-for="(value) in supervisorConfigList" :key="value.name" :label="value.name" :value="value.id">
      </el-option>
    </el-select>
    <el-button :loading="loadingStatus['supervisor_restart_all']" style="margin-left:5px;" type="primary" @click="restartSupervisorAll">重启所有</el-button>
    <el-button :loading="loadingStatus['supervisor_status_list']" style="margin-left:5px;" type="primary" @click="supervisorStatusList">查看所有</el-button>

    <el-tooltip class="item" content="停止,可降低docker内存占用" effect="dark" placement="top" style="margin-left:5px;">
      <el-button :loading="loadingStatus['stopListConsumer']" type="primary" @click="stopListSupervisor">停止以下{{ searchNum }}个
      </el-button>
    </el-tooltip>
    <el-input
        v-model="searchKey"
        autocomplete="off"
        placeholder="搜索名称/进程名/程序名等,多条件使用空格分割"
        style="width: 400px;margin-left:5px;"
        @input="searchList"></el-input>

    <br/>
    <br/>

  </div>

  <el-table :data="configMap" :row-class-name="getColumnColor" style="width: 100%;font-size:14px;margin-top: 10px;">
    <el-table-column label="自定义名称" width="300" >
      <template #default="scope">
        <span v-html="scope.row.showName"></span>
        <el-icon color="#409eff" size="small" @click="editName(scope.row)">
          <Edit></Edit>
        </el-icon>
      </template>
    </el-table-column>
    <el-table-column label="名称" width="300">
      <template #default="scope">
        <span v-html="scope.row.name"></span>
      </template>
    </el-table-column>
    <el-table-column label="运行状态" sortable>
      <template #default="scope">
        <span v-html="scope.row.running_status"></span>
      </template>
    </el-table-column>
    <el-table-column label="进程数" prop="processNum" width="100"/>
    <el-table-column fixed="right" label="操作" width="300">
      <template #default="scope">
        <el-button class="button" size="small"  @click="restart(scope.row)">重新启动</el-button>
        <el-button class="button" size="small" @click="stop(scope.row)">停止
        </el-button>
        <el-button class="button" size="small" @click="configShow(scope.row)">查看配置</el-button>
      </template>
    </el-table-column>
    <div style="height:600px;"></div>
  </el-table>
  <div style="height:300px;"></div>

  <el-dialog v-model="dialogShowEditName" title="输入名称" width="30%">
    <el-input
        v-model="inputNameValue"
        autocomplete="off"
        placeholder="输入名称"
        style="width: 400px"
    ></el-input>
    <template #footer>
      <el-button @click="dialogShowEditName = false">取 消</el-button>
      <el-button
          type="primary"
          @click="
            dialogShowEditName = false;
            editNameValueFunc()
          "
      >确 定
      </el-button
      >
    </template>
  </el-dialog>
  <shellResult ref="shellRef" :shellShowResult="shellController.sshResult" :isRunning="shellController.isRunning" :show-model="shellController.showModel"></shellResult>
</template>
<script>
import store from '../utils/base/store'
import supervisor from '../utils/base/supervisor'
import base from '../utils/base.js'
import array from '@/utils/base/array'
import shellResult from '../components/shell/result_button.vue'
import socket from "@/utils/base/socket";
import format from "@/utils/base/format";
import arr from "@/utils/base/array";
import sse from "@/utils/base/sse";
import t from "@/utils/base/type";
import shell from "@/utils/base/shell";
import Init from '@/utils/base/set_init'
import sseDistribute from "@/utils/base/sse_distribute";
import {Throttle_string} from "@/utils/base/throttle_string";
import search from "@/utils/base/search";

export default {
  props : {
  },
  components: {
    shellResult,
  },
  activated: function () {
    this.resizeTerminal()
    if(Init.GetIsInit('supervisor') === true){
      let _that = this
      supervisor.SupervisorConfigList({sse_distribute_id : _that.sse_distribute_id},function (response){
        if(response.ErrCode === 0){
          _that.supervisorConfigList = response.Data.supervisor_list
          arr.SortByKey(_that.supervisorConfigList , 'name' , 'asc')
          Init.DelInit('supervisor')
        }
      })
    }
  },
  data() {
    return {
      name: 'Supervisor',
      //shell
      shellController : {
        sshResult : '',
        isRunning : false,
        showModel : 'button',
      },
      //选中的环境
      chooseSupervisorId: '0',
      chooseSupervisorConfig : {},
      //是否显示所有的消费者
      showAllSupervisor: false,
      showResultDialog: false,
      dialogShowEditName: false,
      inputNameValue: '',
      editNameValue: {},
      searchNum: 0,
      //消费者环境
      supervisorConfigList: [],
      //存储所有的消费者配置文件
      configMap: [],
      execResult: '', //操作结果
      //历史记录
      useSortSupervisorList: [],
      //搜索key
      searchKey: '',
      supervisorOriginConfList: [],
      //终端
      showInteraction: false,
      showInteractionTitle: '',
      showInteractionSshConfig: {},
      loadingStatus: {},
      sseId : '',
    }
  },
  inject: ["showTerminal", "resizeTerminal"],
  mounted: function () {
    let _that = this
    _that.sse_distribute_id = sseDistribute.GetSseDistributeId('supervisor')
    supervisor.SupervisorConfigList({sse_distribute_id : _that.sse_distribute_id},function (response){
      if(response.ErrCode === 0){
        _that.supervisorConfigList = response.Data.supervisor_list
        arr.SortByKey(_that.supervisorConfigList , 'name' , 'asc')
        _that.chooseSupervisorId = _that.getLastSupervisorId()
        _that.changeSupervisor()
      }
    })
    _that.loadingStatus = _that.$helperLoad.getExecTypeStatus()
  },
  onload: function () {
  },
  filters: {
    limitTo(value, length) {
      return value.slice(0, length)
    },
    substr(value, length) {
      return value.substr(0, length)
    },
  },
  methods: {
    getLastSupervisorId : function (){
      let _that = this
      let chooseSupervisorId = _that.$helperStore.getStore('chooseSupervisorId')
      if(chooseSupervisorId === null || chooseSupervisorId === undefined || isNaN(chooseSupervisorId)){
        chooseSupervisorId = 0
      }
      if(chooseSupervisorId === 0 && _that.supervisorConfigList.length > 0){
        return _that.supervisorConfigList[0].id
      }
      for(let i in _that.supervisorConfigList){
        if(parseInt(_that.supervisorConfigList[i].id) === parseInt(chooseSupervisorId)){
          chooseSupervisorId = _that.supervisorConfigList[i].id
        }
      }
      return chooseSupervisorId
    },
    //获取列背景颜色
    getColumnColor: function (value) {
      if (!value.row.show) {
        return 'row-hide';
      }
      if (value.row.running_status) {
        if (value.row.running_status.indexOf('未启动') >= 0) {
          return 'warning-row';
        } else if (value.row.running_status.indexOf('FATAL') >= 0) {
          return 'error-row';
        } else {
          return '';
        }
      } else {
        return '';
      }
    },
    restart: function (value) {
      let _that = this
      _that.shellController.isRunning = true
      _that.chooseSupervisorConfig.sse_distribute_id = _that.sse_distribute_id
      supervisor.SupervisorRestart(_that.chooseSupervisorConfig, value.supervisor_name, function (response) {
            _that.$helperNotify.success('成功')
            _that.execResult = response.Data
            _that.supervisorStatusList()
            _that.searchList()
            _that.shellController.isRunning = false
          }
      )
    },
    stop: function (value) {
      let _that = this
      _that.shellController.isRunning = true
      _that.chooseSupervisorConfig.sse_distribute_id = _that.sse_distribute_id
      supervisor.SupervisorStop(_that.chooseSupervisorConfig, value.supervisor_name, function (response) {
            _that.$helperNotify.success('成功')
            _that.execResult = response.Data
            _that.supervisorStatusList()
            _that.searchList()
            _that.shellController.isRunning = false
          }
      )
    },
    configShow: function (value) {
      let _that = this
      _that.openShellResult()
      _that.shellController.isRunning = true
      _that.chooseSupervisorConfig.sse_distribute_id = _that.sse_distribute_id
      supervisor.SupervisorConfigShow(_that.chooseSupervisorConfig,value.supervisor_config, function (response) {
            _that.execResult = response.Data
            _that.supervisorStopRestartExplain(value)
            _that.searchList()
            _that.shellController.isRunning = false
          }
      )
    },
    stopAll: function () {
    },
    //停止列表下面的消费者
    stopListSupervisor: function () {
      if (this.searchKey === '') {
        this.stopAll()
        return
      }
      for (let i in this.configMap) {
        if (this.configMap[i].show === true) {
          this.stop(this.configMap[i])
        }
      }
    },
    //打开shell
    openShellResult : function (){
      this.$refs.shellRef.openDrawer()
    },
    //拿到config 列表
    getOriginSupervisorConf: function () {
      let _that = this
      if(!_that.chooseSupervisorConfig || !_that.chooseSupervisorConfig.ssh_id){
        return
      }
      _that.shellController.isRunning = true
      _that.chooseSupervisorConfig.sse_distribute_id = _that.sse_distribute_id
      supervisor.SupervisorConfList(_that.chooseSupervisorConfig, function (response) {
            let tempList = response.Data.split(`\n`)
            let confList = []
            for (let i in tempList) {
              confList.push(tempList[i].split('---'))
            }
            _that.configMap = _that.$helperConfig.getSupervisorConfigList(confList, _that.chooseSupervisorConfig)
            _that.supervisorStatusList()
            _that.shellController.isRunning = false
          }
      )
    },
    //选择代码环境
    changeSupervisor: function () {
      let _that = this
      for(let i in _that.supervisorConfigList){
        if(parseInt(_that.supervisorConfigList[i].id) === parseInt(_that.chooseSupervisorId)){
          _that.chooseSupervisorConfig = _that.supervisorConfigList[i]
        }
      }
      _that.$helperStore.setStore('chooseSupervisorId' , _that.chooseSupervisorId)
      let throttleStringFunc = new Throttle_string(50, text => {
        _that.shellController.sshResult += text
        // 限制长度：最多保留最后 10000 个字符
        const maxLen = 10000;
        if (_that.shellController.sshResult.length > maxLen) {
          _that.shellController.sshResult = _that.shellController.sshResult.slice(-maxLen);
        }
        let result = format.formatResult(_that.shellController.sshResult, ['copy', 'color', 'replace'])
        result = format.formatResult(result, ['length'])
        _that.shellController.sshResult = result
      })
      sseDistribute.RegisterReceive(_that.sse_distribute_id , function (msg,msgType,sseDistributeId){
        throttleStringFunc.update(msg)
      })
      _that.getOriginSupervisorConf()
    },
    //搜索消费者列表
    searchList: function () {
      let _that = this
      let ret = search.SearchListObj(_that.configMap, _that.searchKey)
      _that.searchNum = ret[0]
      _that.configMap = ret[1]
    },
    //重启所有的消费者
    restartSupervisorAll: function () {
      let _that = this
      _that.shellController.isRunning = true
      _that.chooseSupervisorConfig.sse_distribute_id = _that.sse_distribute_id
      supervisor.SupervisorRestartAll(_that.chooseSupervisorConfig, function (response) {
            _that.execResult = response.Data
            _that.supervisorStatusList()
            _that.searchList()
            _that.shellController.isRunning = false
          }
      )
    },
    //查看所有的消费者运行状态列表
    supervisorStatusList: function () {
      let _that = this
      _that.shellController.isRunning = true
      _that.chooseSupervisorConfig.sse_distribute_id = _that.sse_distribute_id
      supervisor.SupervisorStatusList(_that.chooseSupervisorConfig, function (response) {
            _that.execResult = response.Data
            _that.supervisorStatusExplain()
            _that.searchList()
            _that.shellController.isRunning = false
          }
      )
    },
    //修改名称
    editName: function (param) {
      this.editNameValue = param
      this.inputNameValue = this.editNameValue.showName
      this.dialogShowEditName = true
    },
    editNameValueFunc: function () {
      this.$helperStore.setStore(this.editNameValue.name, this.inputNameValue)
      this.flushConfigList()
      this.refreshUseSortSupervisor()
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
    //刷新排序
    refreshUseSortSupervisor: function () {
      let cackeKey = 'useSortSupervisor'
      let useSortSupervisor = this.$helperStore.getStore(cackeKey)
      if (useSortSupervisor === null || useSortSupervisor === undefined) {
        this.useSortSupervisorList = []
      } else {
        this.useSortSupervisorList = JSON.parse(useSortSupervisor)
      }
      this.useSortSupervisorList.sort(function (a, b) {
        return b.key - a.key
      })
      this.useSortSupervisorList = this.useSortSupervisorList.slice(0, 10)
      for (let j in this.useSortSupervisorList) {
        let showName = this.$helperStore.getStore(
            this.useSortSupervisorList[j].name
        )
        if (showName === null || showName === undefined) {
          showName = this.useSortSupervisorList[j].name
        }
        this.useSortSupervisorList[j].showName = showName
      }
    },
    //分析重启或者停止后的结果
    supervisorStopRestartExplain: function (param) {
      let supervisorStatusList = this.execResult.split('\n')
      for (let i in supervisorStatusList) {
        if (supervisorStatusList[i] === '') {
          continue
        }
        if (supervisorStatusList[i].indexOf('RUNNING') !== -1) {
          let runningStatus = supervisorStatusList[i].substr(
              supervisorStatusList[i].indexOf('RUNNING')
          )
          this.getRunningStatus(runningStatus, param.name)
        }

        if (supervisorStatusList[i].indexOf('FATAL') !== -1) {
          let runningStatus = supervisorStatusList[i].substr(
              supervisorStatusList[i].indexOf('FATAL')
          )
          this.getRunningStatus(runningStatus, param.name)
        }

        if (supervisorStatusList[i].indexOf('STOPPED') !== -1) {
          let runningStatus = supervisorStatusList[i].substr(
              supervisorStatusList[i].indexOf('STOPPED')
          )
          this.getRunningStatus(runningStatus, param.name)
        }
      }
    },
    getRunningStatus: function (runningStatus, name) {
      for (let n in this.configMap) {
        if (this.configMap[n].name === name) {
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
      let supervisorStatusList = this.execResult.split('\n')
      for (let i in supervisorStatusList) {
        if (supervisorStatusList[i] === '') {
          continue
        }
        //根据；分割
        let name_params = []
        if(supervisorStatusList[i].match(/^[^\s]+/g)){
          name_params.push(supervisorStatusList[i].match(/^[^\s]+/g)[0])
        }else{
          name_params.push('-')
        }
        name_params.push(supervisorStatusList[i].replace(name_params[0], ''))
        //循环判断
        let name_params_two = this.filterArray(name_params)
        //获取supervisor进程名
        if (name_params_two.length === 0) {
          continue
        }
        let name = name_params_two[0]
        let name_params_four = this.filterArray(name.split(':'))
        if (name_params_four.length === 0) {
          continue
        }
        //给与状态
        for (let n in this.configMap) {
          if (this.configMap[n].supervisor_name === name_params_four[0]) {
            this.configMap[n].running_status = name_params_two[1]
            //重启名
            if (name_params_four.length === 2) {
              this.configMap[n].supervisor_restart_name =
                  name_params_four[0] + ':'
            } else {
              this.configMap[n].supervisor_restart_name = name_params_four[0]
            }
            this.configMap[n].show = true
            this.configMap[n].processNum++
            break
          } else {
            this.configMap[n].show = true
          }
        }
      }
      for (let k in this.configMap) {
        if (this.configMap[k].running_status === ``) {
          this.configMap[k].running_status = '未启动'
        }
      }
      this.configMap = array.SortByKey(this.configMap , 'running_status' , 'asc')
    },
    //过滤数组空数据
    filterArray: function (array) {
      let return_array = []
      for (let m in array) {
        if (array[m] !== '') {
          return_array.push(array[m])
        }
      }
      return return_array
    },
  },
}
</script>

<style>
.supervisorCommand {
  padding: 3px;
  font-size: 14px;
}
.card-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
}

.el-table__body-wrapper {
  /*margin-bottom: 300px;*/
}

.text {
  font-size: 14px;
}

.item {
  margin-bottom: 18px;
  text-align: left;
}

.el-table .warning-row {
  --el-table-tr-bg-color: var(--el-color-warning-light-9);
}

.el-table .success-row {
  --el-table-tr-bg-color: var(--el-color-success-light-9);
}

.el-table .error-row {
  --el-table-tr-bg-color: var(--el-color-error-light-9);
}

.row-hide {
  display: none;
}
</style>
