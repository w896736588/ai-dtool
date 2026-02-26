<template>
  <div style="text-align: center;">
    <el-select v-model="chooseSshId" placeholder="请选择环境" style="width:300px;" @change="changeSsh">
      <el-option v-for="(value) in sshList" :key="value.name" :label="value.name" :value="value.id">
      </el-option>
    </el-select>
    <!--    <el-button :loading="loadingStatus['supervisor_restart_all']" style="margin-left:5px;" type="primary" @click="restartSupervisorAll">重启所有</el-button>-->
    <el-button :loading="loadingStatus['supervisor_status_list']" style="margin-left:5px;" type="primary" @click="getComposeList">
      刷新
    </el-button>
    <el-input
        v-model="searchKey"
        autocomplete="off"
        placeholder="搜索名称等,多条件使用空格分割"
        style="width: 400px;margin-left:5px;"
        @input="searchList"></el-input>
    <br/>
    <br/>
  </div>

  <el-table :data="composeList" :row-class-name="getColumnColor" style="width: 100%;font-size:14px;margin-top: 10px;">
    <el-table-column label="名称" sortable width="160">
      <template #default="scope">
        <div style="display: flex; align-items: center; gap: 8px; height: 100%;">
          <!-- 星标按钮 -->
          <el-icon
              :size="18"
              :color="scope.row.starred ? '#e6a23c' : '#c0c4cc'"
              style="cursor: pointer; transition: all 0.3s ease; flex-shrink: 0;"
              @click="toggleStar(scope.row)"
              class="star-icon"
              :class="{ 'starred': scope.row.starred }"
          >
            <Star />
          </el-icon>
          <div v-html="scope.row.name" style="line-height: 1.2; display: flex; align-items: center;"></div>
        </div>
      </template>
    </el-table-column>
    <el-table-column label="位置" sortable width="200">
      <template #default="scope">
        <div v-html="scope.row.compose_yml_path"></div>
      </template>
    </el-table-column>
    <el-table-column label="env file" sortable width="200">
      <template #default="scope">
        <div v-html="scope.row.env_file"></div>
      </template>
    </el-table-column>
    <el-table-column fixed="right" label="操作">
      <template #default="scope">
        <div style="margin-top: 10px;">
          <span style="font-weight:400;">常用操作：</span>
          <el-button class="button" size="small" @click="dialogServices(scope.row)">服务列表</el-button>
          <el-button class="button" size="small" @click="status(scope.row)">运行状态</el-button>
          <el-button class="button" size="small" @click="start(scope.row)">启动（up -d）</el-button>
          <el-button class="button" size="small" @click="restart(scope.row)">重启（restart）</el-button>
          <el-button class="button" size="small" type="danger" @click="stop(scope.row)">停止(stop)
          </el-button>
          <el-button class="button" size="small" @click="configShow(scope.row)">查看compose.yml</el-button>
          <el-button class="button" size="small" @click="envShow(scope.row)">查看env</el-button>
        </div>
        <div style="margin-top: 10px;">
          <span style="font-weight:400;">快速重启：</span>
          <template v-for="(item, index) in scope.row.default_service_list">
            <el-button link type="primary" @click="restart(scope.row , item)">{{ item }}</el-button>
          </template>
        </div>
        <div style="margin-top: 10px;">
          <span style="font-weight:400;">快速停止：</span>
          <template v-for="(item, index) in scope.row.default_service_list">
            <el-button link type="warning" @click="stop(scope.row , item)">{{ item }}</el-button>
          </template>
        </div>
      </template>
    </el-table-column>
    <div style="height:600px;"></div>
  </el-table>
  <div style="height:300px;"></div>

  <shellResult ref="shellRef" :isRunning="shellController.isRunning" :shellShowResult="shellController.sshResult" :show-model="shellController.showModel"></shellResult>

  <el-dialog v-model="dialogStatus" :append-to-body="true" title="状态" width="80%">
    <el-table :data="dialogStatusData" style="width: 100%">
      <el-table-column label="服务名" prop="NAME" width="250"/>
      <el-table-column label="CPU使用率" prop="CPU %" width="120"/>
      <el-table-column label="内存使用量" prop="MEM USAGE / LIMIT" width="240"/>
      <el-table-column label="内存使用率" prop="MEM %" width="120"/>
      <el-table-column label="网络收发流量" prop="NET I/O"/>
      <el-table-column label="磁盘块设备读写量" prop="BLOCK I/O"/>
    </el-table>
  </el-dialog>

  <el-dialog v-model="dialogShowService" :append-to-body="true" title="服务" width="80%">
    <el-button type="primary" link @click="refreshServices()">刷新服务列表</el-button>
    <el-table :data="dialogServiceConfig.services" style="width: 100%">
      <el-table-column label="服务名" prop="name" width="250"/>
      <el-table-column label="操作">
        <template #default="scope">
          <el-button link type="primary" @click="restart(dialogServiceConfig , scope.row.name)">restart</el-button>
          <el-button link type="primary" @click="stop(dialogServiceConfig , scope.row.name)">stop</el-button>
          <el-button link type="primary" @click="start(dialogServiceConfig , scope.row.name)">up</el-button>
<!--          <el-button link type="primary" @click="status(dialogServiceConfig , scope.row.name)">上传可执行文件并重启</el-button>-->
        </template>
      </el-table-column>
    </el-table>
  </el-dialog>
</template>
<style>
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
<script>
import store from '../utils/base/store'
import compose from '../utils/base/compose'
import base from '../utils/base.js'
import array from '@/utils/base/array'
import shellResult from '../components/shell/result_button.vue'
import socket from "@/utils/base/socket";
import format from "@/utils/base/format";
import arr from "@/utils/base/array";
import ssh from "@/utils/base/ssh_set"
import search from "@/utils/base/search"
import sse from "@/utils/base/sse";
import t from "@/utils/base/type";
import shell from "@/utils/base/shell";
import sseDistribute from "@/utils/base/sse_distribute";
import {Throttle_string} from "@/utils/base/throttle_string";
import type from "@/utils/base/type";

export default {
  props: {},
  components: {
    shellResult,
  },
  data() {
    return {
      name: 'Compose',
      //shell
      shellController: {
        sshResult: '',
        isRunning: false,
        showModel: 'button',
      },
      dialogStatus: false,
      dialogStatusData: [],
      dialogShowService: false,
      dialogServiceConfig: {},
      //选中的环境
      chooseSshId: '',
      chooseComposeeConfig: {},
      //是否显示所有的消费者
      showAllSupervisor: false,
      showResultDialog: false,
      dialogShowEditName: false,
      inputNameValue: '',
      editNameValue: {},
      searchNum: 0,
      composeList: [],
      sshList: [],
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
      sse_distribute_id: '',
    }
  },
  inject: ["showTerminal", "resizeTerminal"],
  activated: function () {
    this.resizeTerminal()
  },
  mounted: function () {
    let _that = this
    _that.sse_distribute_id = sseDistribute.GetSseDistributeId('docker')
    ssh.SshList(function (response) {
      if (response.ErrCode === 0) {
        _that.sshList = response.Data
        if (_that.sshList.length > 0) {
          _that.chooseSshId = parseInt(_that.getLastSshId())
          let exist = false
          for (let i in _that.sshList) {
            if (parseInt(_that.sshList[i]['id']) === parseInt(_that.chooseSshId)) {
              exist = true
            }
          }
          if (!exist) {
            _that.chooseSshId = '' + _that.sshList[0]['id']
          }
          _that.changeSsh()
        }
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
    // 切换星标状态
    toggleStar: function(row) {
      let _that = this
      // 获取当前星标列表（确保是数组）
      let starredList = _that.getStarredList()

      if (row.starred) {
        // 取消星标
        const index = starredList.indexOf(row.id)
        if (index > -1) {
          starredList.splice(index, 1)
        }
        row.starred = false
      } else {
        // 添加星标
        if (!starredList.includes(row.id)) {
          starredList.push(row.id)
        }
        row.starred = true
      }

      // 保存到本地存储（确保保存为字符串化的数组）
      _that.saveStarredList(starredList)

      // 重新排序列表，星标项目优先显示
      _that.sortComposeList()
    },

    // 获取星标列表
    getStarredList: function() {
      let _that = this
      let starredList = _that.$helperStore.getStore('dockerComposeStarredList')

      // 如果不存在，返回空数组
      if (!starredList) {
        return []
      }

      // 如果已经是数组，直接返回
      if (Array.isArray(starredList)) {
        return starredList
      }

      // 如果是字符串，尝试解析
      if (typeof starredList === 'string') {
        try {
          const parsed = JSON.parse(starredList)
          return Array.isArray(parsed) ? parsed : []
        } catch (e) {
          console.error('解析星标列表失败:', e)
          return []
        }
      }

      // 其他情况返回空数组
      return []
    },

    // 保存星标列表
    saveStarredList: function(starredList) {
      let _that = this
      // 确保是数组
      if (!Array.isArray(starredList)) {
        starredList = []
      }
      // 保存为 JSON 字符串
      _that.$helperStore.setStore('dockerComposeStarredList', JSON.stringify(starredList))
    },

    // 排序列表，星标项目优先
    sortComposeList: function () {
      let _that = this
      let starredList = _that.getStarredList()

      _that.composeList.sort((a, b) => {
        const aStarred = starredList.includes(a.id)
        const bStarred = starredList.includes(b.id)

        if (aStarred && !bStarred) {
          return -1 // a排在b前面
        } else if (!aStarred && bStarred) {
          return 1 // b排在a前面
        } else {
          return 0 // 保持原顺序
        }
      })
    },

    // 初始化星标状态
    initStarStatus: function () {
      let _that = this
      let starredList = _that.getStarredList()

      // 为每个项目设置星标状态
      _that.composeList.forEach(item => {
        item.starred = starredList.includes(item.id)
      })

      // 排序列表
      _that.sortComposeList()
    },
    refreshServices : function (row){
      let _that = this
      //优先从缓存拿
      let servicesKey = 'docker_services_' + _that.chooseSshId + '_' + _that.dialogServiceConfig.id
      let data = {
        ssh_id: _that.chooseSshId,
        id: _that.dialogServiceConfig.id,
        sse_distribute_id: _that.sse_distribute_id,
      }
      compose.DockerComposeServices(data, function (response) {
            _that.$helperNotify.success('成功')
            _that.shellController.isRunning = false
            _that.dialogServiceConfig.services = response.Data.services
            store.setStore(servicesKey,JSON.stringify(response.Data.services || []))
          }
      )
    },
    dialogServices: function (row) {
      let _that = this
      _that.dialogServiceConfig = row
      _that.shellController.isRunning = true
      let data = {
        ssh_id: _that.chooseSshId,
        id: row.id,
        sse_distribute_id: _that.sse_distribute_id,
      }
      //优先从缓存拿
      let servicesKey = 'docker_services_' + _that.chooseSshId + '_' + row.id
      let services =store.getStore(servicesKey)
      if(type.IsString(services)){
        _that.shellController.isRunning = false
        _that.dialogShowService = true
        _that.dialogServiceConfig.services = JSON.parse(services)
        return
      }
      compose.DockerComposeServices(data, function (response) {
            _that.$helperNotify.success('成功')
            _that.shellController.isRunning = false
            _that.dialogShowService = true
            _that.dialogServiceConfig.services = response.Data.services
            store.setStore(servicesKey,JSON.stringify(response.Data.services || []))
          }
      )
    },
    getLastSshId: function () {
      let _that = this
      let chooseSshId = _that.$helperStore.getStore('dockerChooseSshId')
      if (chooseSshId === null || chooseSshId === undefined || isNaN(chooseSshId)) {
        chooseSshId = 0
      }
      if (chooseSshId === 0 && _that.composeList.length > 0) {
        return _that.composeList[0].id
      }
      for (let i in _that.composeList) {
        if (parseInt(_that.composeList[i].id) === parseInt(chooseSshId)) {
          chooseSshId = _that.composeList[i].id
        }
      }
      return chooseSshId
    },
    //获取列背景颜色
    getColumnColor: function (value) {
      if (!value.row.show) {
        return 'row-hide';
      }
      if (value.row.State) {
        if (value.row.State.indexOf('Up') >= 0) {
          return 'success-row';
        } else if (value.row.running_status.indexOf('FATAL') >= 0) {
          return 'error-row';
        } else {
          return '';
        }
      } else {
        return '';
      }
    },
    restart: function (value, service) {
      let _that = this
      _that.shellController.isRunning = true
      let data = {
        ssh_id: _that.chooseSshId,
        id: value.id,
        sse_distribute_id: _that.sse_distribute_id,
        service: service,
      }
      compose.DockerComposeRestart(data, function (response) {
            _that.$helperNotify.success('成功')
            _that.getComposeList()
            _that.shellController.isRunning = false
          }
      )
    },
    stop: function (value , service) {
      let _that = this
      _that.shellController.isRunning = true
      let data = {
        ssh_id: _that.chooseSshId,
        id: value.id,
        sse_distribute_id: _that.sse_distribute_id,
        service : service,
      }
      compose.DockerComposeStop(data, function (response) {
            _that.$helperNotify.success('成功')
            _that.getComposeList()
            _that.shellController.isRunning = false
          }
      )
    },
    start: function (value , service) {
      let _that = this
      _that.shellController.isRunning = true
      let data = {
        ssh_id: _that.chooseSshId,
        id: value.id,
        sse_distribute_id: _that.sse_distribute_id,
        service : service,
      }
      compose.DockerComposeStart(data, function (response) {
            _that.$helperNotify.success('成功')
            _that.getComposeList()
            _that.shellController.isRunning = false
          }
      )
    },
    status: function (value) {
      let _that = this
      _that.shellController.isRunning = true
      let data = {
        ssh_id: _that.chooseSshId,
        id: value.id,
        sse_distribute_id: _that.sse_distribute_id,
      }
      compose.DockerComposeStatus(data, function (response) {
            _that.$helperNotify.success('成功')
            _that.shellController.isRunning = false
            _that.dialogStatus = true
            _that.dialogStatusData = response.Data.status
          }
      )
    },
    configShow: function (value) {
      let _that = this
      _that.openShellResult()
      _that.shellController.isRunning = true
      let data = {
        config_path: value.compose_yml_path,
        ssh_id: _that.chooseSshId,
        sse_distribute_id: _that.sse_distribute_id,
      }
      compose.DockerComposeConfigShow(data, function (response) {
            _that.execResult = response.Data
            _that.shellController.isRunning = false
          }
      )
    },
    envShow: function (value) {
      let _that = this
      _that.openShellResult()
      _that.shellController.isRunning = true
      let envFile = value.env_file
      if (envFile === '') {
        envFile = value.compose_yml_path.replace(/\/[^\/]+\.yml$/, '/.env')
      }
      if (envFile === '') {
        _that.$helperNotify.error('未找到.env路径')
        return;
      }
      let data = {
        config_path: envFile,
        ssh_id: _that.chooseSshId,
        sse_distribute_id: _that.sse_distribute_id,
      }
      compose.DockerComposeConfigShow(data, function (response) {
            _that.execResult = response.Data
            _that.shellController.isRunning = false
          }
      )
    },
    //打开shell
    openShellResult: function () {
      this.$refs.shellRef.openDrawer()
    },
    getComposeList: function () {
      let _that = this
      if (!_that.chooseSshId) {
        return
      }
      _that.shellController.isRunning = true
      compose.DockerComposeList({ssh_id: _that.chooseSshId, sse_distribute_id: _that.sse_distribute_id}, function (response) {
            if (response.ErrCode === 0) {
              _that.composeList = response.Data.list
              for (let i in _that.composeList) {
                _that.composeList[i].show = true
                _that.composeList[i].default_service_list = _that.composeList[i].default_service.split(',')
              }
              // 初始化星标状态
              _that.initStarStatus()
            }
            _that.shellController.isRunning = false
          }
      )
    },
    //选择代码环境
    changeSsh: function () {
      let _that = this
      _that.$helperStore.setStore('dockerChooseSshId', _that.chooseSshId)
      let throttleStringFunc = new Throttle_string(50, text => {
        _that.shellController.sshResult += text
        // 限制长度：最多保留最后 50000 个字符
        const maxLen = 50000;
        if (_that.shellController.sshResult.length > maxLen) {
          _that.shellController.sshResult = _that.shellController.sshResult.slice(-maxLen);
        }
        let result = format.formatResult(
            _that.shellController.sshResult, ['copy', 'color', 'replace']);
        result = format.formatResult(result, ['length']);
        _that.shellController.sshResult = result;   // 一次性赋值，减少 watcher 抖动
      })
      sseDistribute.RegisterReceive(_that.sse_distribute_id, function (msg, msgType, sseDistributeId) {
        throttleStringFunc.update(msg)
      })
      _that.getComposeList()
    },
    //搜索消费者列表
    searchList: function () {
      let _that = this
      let ret = search.SearchListObj(_that.composeList, _that.searchKey)
      _that.searchNum = ret[0]
      _that.composeList = ret[1]
    },
  },
}
</script>

<style>
.supervisorCommand {
  padding: 3px;
  font-size: 14px;
}

.star-icon:hover {
  transform: scale(1.2);
}

.star-icon.starred {
  animation: starPulse 0.3s ease;
}

@keyframes starPulse {
  0% { transform: scale(1); }
  50% { transform: scale(1.3); }
  100% { transform: scale(1); }
}
</style>
