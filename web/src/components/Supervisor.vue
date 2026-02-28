<template>
  <div class="supervisor-page-container">
    <!-- 顶部操作区域 -->
    <div class="supervisor-header-card">
      <div class="header-title">
        <svg class="header-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect x="3" y="3" width="18" height="18" rx="2" stroke="currentColor" stroke-width="2"/>
          <path d="M9 9H15" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          <path d="M9 13H15" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          <path d="M9 17H12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        </svg>
        <span>Supervisor 进程管理</span>
      </div>
      <div class="control-row">
        <el-select v-model="chooseSupervisorId" placeholder="选择环境" @change="changeSupervisor" class="env-select">
          <el-option v-for="(value) in supervisorConfigList" :key="value.name" :label="value.name" :value="value.id">
          </el-option>
        </el-select>
        <div class="action-buttons">
          <el-button :loading="loadingStatus['supervisor_restart_all']" type="warning" plain @click="restartSupervisorAll">
            <el-icon><RefreshRight /></el-icon>重启所有
          </el-button>
          <el-button :loading="loadingStatus['supervisor_status_list']" type="primary" plain @click="supervisorStatusList">
            <el-icon><View /></el-icon>查看状态
          </el-button>
          <el-tooltip content="停止选中的进程，可降低内存占用" placement="top">
            <el-button :loading="loadingStatus['stopListConsumer']" type="danger" plain @click="stopListSupervisor">
              <el-icon><VideoPause /></el-icon>停止选中 ({{ searchNum }})
            </el-button>
          </el-tooltip>
        </div>
        <el-input
          v-model="searchKey"
          autocomplete="off"
          placeholder="搜索名称/进程名，空格多条件"
          class="search-input"
          @input="searchList"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
    </div>

    <!-- 进程列表 -->
    <div class="process-table-card">
      <el-table :data="configMap" :row-class-name="getColumnColor" class="process-table" stripe>
        <el-table-column label="自定义名称" min-width="200">
          <template #default="scope">
            <div class="name-cell">
              <span class="custom-name" v-html="scope.row.showName"></span>
              <el-icon class="edit-icon" @click="editName(scope.row)">
                <Edit />
              </el-icon>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="进程名称" min-width="200">
          <template #default="scope">
            <code class="process-name" v-html="scope.row.name"></code>
          </template>
        </el-table-column>
        <el-table-column label="运行状态" width="180" sortable>
          <template #default="scope">
            <div class="status-cell">
              <span v-html="scope.row.running_status" class="status-text"></span>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="进程数" prop="processNum" width="100" align="center">
          <template #default="scope">
            <el-tag size="small" type="info">{{ scope.row.processNum }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column fixed="right" label="操作" width="280">
          <template #default="scope">
            <div class="action-cell">
              <el-button size="small" type="success" plain @click="restart(scope.row)">
                <el-icon><RefreshRight /></el-icon>重启
              </el-button>
              <el-button size="small" type="warning" plain @click="stop(scope.row)">
                <el-icon><VideoPause /></el-icon>停止
              </el-button>
              <el-button size="small" type="primary" plain @click="configShow(scope.row)">
                <el-icon><Document /></el-icon>配置
              </el-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 编辑名称弹窗 -->
    <el-dialog v-model="dialogShowEditName" title="编辑自定义名称" width="400px" class="edit-name-dialog">
      <el-form label-width="80px">
        <el-form-item label="名称">
          <el-input v-model="inputNameValue" autocomplete="off" placeholder="输入自定义名称" />
        </el-form-item>
      </el-form>
      <template #footer>
        <el-button @click="dialogShowEditName = false">取消</el-button>
        <el-button type="primary" @click="dialogShowEditName = false; editNameValueFunc()">确定</el-button>
      </template>
    </el-dialog>

    <shellResult ref="shellRef" :shellShowResult="shellController.sshResult" :isRunning="shellController.isRunning" :show-model="shellController.showModel"></shellResult>
  </div>
</template>
<script>
import { RefreshRight, View, VideoPause, Search, Edit, Document } from '@element-plus/icons-vue';
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
    RefreshRight,
    View,
    VideoPause,
    Search,
    Edit,
    Document,
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
      sse_distribute_id: '',
      sseThrottleStringFunc: null,
    }
  },
  inject: ["showTerminal", "resizeTerminal"],
  mounted: function () {
    let _that = this
    _that.prepareActionSse('init')
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
  beforeUnmount() {
    if (this.sse_distribute_id) {
      sseDistribute.UnRegisterReceive(this.sse_distribute_id)
    }
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
    prepareActionSse: function (action) {
      let _that = this
      if (_that.sse_distribute_id) {
        sseDistribute.UnRegisterReceive(_that.sse_distribute_id)
      }
      _that.sse_distribute_id = sseDistribute.GetSseDistributeId(`supervisor_${action}_${Date.now()}`)
      if (!_that.sseThrottleStringFunc) {
        _that.sseThrottleStringFunc = new Throttle_string(50, text => {
          _that.shellController.sshResult += text
          const maxLen = 10000
          if (_that.shellController.sshResult.length > maxLen) {
            _that.shellController.sshResult = _that.shellController.sshResult.slice(-maxLen)
          }
          let result = format.formatResult(_that.shellController.sshResult, ['copy', 'color', 'replace'])
          result = format.formatResult(result, ['length'])
          _that.shellController.sshResult = result
        })
      }
      sseDistribute.RegisterReceive(_that.sse_distribute_id , function (msg){
        _that.sseThrottleStringFunc.update(msg)
      })
      return _that.sse_distribute_id
    },
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
      _that.prepareActionSse('restart')
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
      _that.prepareActionSse('stop')
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
      _that.prepareActionSse('config_show')
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
      _that.prepareActionSse('config_list')
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
      _that.prepareActionSse('restart_all')
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
      _that.prepareActionSse('status_list')
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
/* 页面容器 */
.supervisor-page-container {
  padding: 0;
  width: 100%;
}

/* 顶部卡片样式 */
.supervisor-header-card {
  background: linear-gradient(135deg, #11998e 0%, #38ef7d 100%);
  border-radius: 16px;
  padding: 20px 24px;
  margin-bottom: 20px;
  box-shadow: 0 4px 20px rgba(17, 153, 142, 0.25);
}

.header-title {
  display: flex;
  align-items: center;
  gap: 12px;
  color: #fff;
  font-size: 20px;
  font-weight: 600;
  margin-bottom: 16px;
}

.header-icon {
  width: 28px;
  height: 28px;
  color: #fff;
}

.control-row {
  display: flex;
  gap: 12px;
  align-items: center;
  flex-wrap: wrap;
}

.env-select {
  width: 200px;
}

.env-select :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.action-buttons .el-button {
  background: rgba(255, 255, 255, 0.9);
  border: none;
}

.action-buttons .el-button:hover {
  background: #fff;
}

.search-input {
  flex: 1;
  max-width: 400px;
  min-width: 200px;
}

.search-input :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

/* 进程表格卡片 */
.process-table-card {
  background: #fff;
  border-radius: 16px;
  padding: 20px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
}

.process-table {
  border-radius: 12px;
  overflow: hidden;
}

.process-table :deep(.el-table__header th) {
  background: #f5f7fa;
  color: #606266;
  font-weight: 600;
}

.process-table :deep(.el-table__row:hover > td) {
  background-color: #f0fdf4 !important;
}

.name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.custom-name {
  font-weight: 500;
  color: #303133;
}

.edit-icon {
  color: #11998e;
  cursor: pointer;
  transition: all 0.3s;
}

.edit-icon:hover {
  color: #38ef7d;
  transform: scale(1.2);
}

.process-name {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  color: #11998e;
  background: #f0fdf4;
  padding: 2px 8px;
  border-radius: 4px;
}

.status-cell {
  display: flex;
  align-items: center;
}

.status-text {
  font-size: 13px;
}

.action-cell {
  display: flex;
  gap: 6px;
}

/* 表格行状态 */
.el-table .warning-row {
  --el-table-tr-bg-color: #fef3cd;
}

.el-table .success-row {
  --el-table-tr-bg-color: #d4edda;
}

.el-table .error-row {
  --el-table-tr-bg-color: #f8d7da;
}

.row-hide {
  display: none;
}

/* 弹窗样式 */
.edit-name-dialog :deep(.el-dialog) {
  border-radius: 16px;
}

.edit-name-dialog :deep(.el-dialog__header) {
  border-bottom: 1px solid #ebeef5;
  padding: 16px 20px;
  margin: 0;
}

.edit-name-dialog :deep(.el-dialog__body) {
  padding: 20px;
}

.edit-name-dialog :deep(.el-dialog__footer) {
  border-top: 1px solid #ebeef5;
  padding: 12px 20px;
}

/* 响应式 */
@media (max-width: 1200px) {
  .control-row {
    flex-direction: column;
    align-items: stretch;
  }
  
  .action-buttons {
    flex-wrap: wrap;
  }
  
  .search-input {
    max-width: 100%;
  }
}
</style>
