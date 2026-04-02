<template>
  <div class="docker-page-container">
    <div class="docker-header-card">
      <div class="header-title">
        <svg class="header-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect x="3" y="7" width="18" height="12" rx="2" stroke="currentColor" stroke-width="2"/>
          <path d="M8 11H11" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          <path d="M13 11H16" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          <path d="M8 15H12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        </svg>
        <span>Docker Compose 管理</span>
      </div>
      <div class="control-row">
        <el-select v-model="chooseSshId" placeholder="选择环境" @change="changeSsh" class="env-select">
          <el-option v-for="(value) in sshList" :key="value.name" :label="value.name" :value="value.id">
          </el-option>
        </el-select>
        <div class="action-buttons">
          <pl-button :loading="loadingStatus['supervisor_status_list']" type="primary" plain @click="getComposeList">
            刷新
          </pl-button>
        </div>
        <el-input
            v-model="searchKey"
            autocomplete="off"
            placeholder="搜索名称等，多条件使用空格分割"
            class="search-input"
            @input="searchList"
            clearable
        ></el-input>
        <div class="header-tail-actions">
          <pl-button
            class="header-tail-btn"
            type="warning"
            plain
            @click="openComposeSettings"
          >
            设置
          </pl-button>
          <pl-button
            :disabled="!chooseSshId"
            :loading="containerLogCleaning"
            class="header-tail-btn"
            type="danger"
            plain
            @click="confirmTruncateContainerLogs"
          >
            清理容器日志
          </pl-button>
          <pl-button
            :disabled="!chooseSshId"
            :loading="imageListLoading"
            class="header-tail-btn"
            type="primary"
            plain
            @click="openImageListDialog"
          >
            镜像列表
          </pl-button>
        </div>
      </div>
    </div>

    <div class="compose-table-card">
      <el-table :data="composeList" :row-class-name="getColumnColor" class="compose-table">
        <el-table-column label="名称" sortable width="200">
          <template #default="scope">
            <div class="name-cell">
              <el-icon
                  :size="18"
                  :color="scope.row.starred ? '#e6a23c' : '#c0c4cc'"
                  @click="toggleStar(scope.row)"
                  class="star-icon"
                  :class="{ 'starred': scope.row.starred }"
              >
                <Star />
              </el-icon>
              <div v-html="scope.row.name" class="name-text"></div>
            </div>
          </template>
        </el-table-column>
        <el-table-column label="位置" sortable width="260">
          <template #default="scope">
            <code class="path-text" v-html="scope.row.compose_yml_path"></code>
          </template>
        </el-table-column>
        <el-table-column label="env file" sortable width="260">
          <template #default="scope">
            <code class="path-text" v-html="scope.row.env_file"></code>
          </template>
        </el-table-column>
        <el-table-column fixed="right" label="操作" min-width="620">
          <template #default="scope">
            <div class="operation-block">
              <span class="operation-title">常用操作：</span>
              <div class="operation-buttons">
                <pl-button class="operation-btn operation-btn-primary" size="small" plain @click="dialogServices(scope.row)">服务列表</pl-button>
                <pl-button class="operation-btn operation-btn-primary" size="small" plain @click="status(scope.row)">运行状态</pl-button>
                <pl-button class="operation-btn operation-btn-success" size="small" plain @click="start(scope.row)">启动（up -d）</pl-button>
                <pl-button class="operation-btn operation-btn-success" size="small" plain @click="restart(scope.row)">重启（restart）</pl-button>
                <pl-button class="operation-btn operation-btn-danger" size="small" plain @click="stop(scope.row)">停止(stop)</pl-button>
                <pl-button class="operation-btn operation-btn-primary" size="small" plain @click="configShow(scope.row)">查看compose.yml</pl-button>
                <pl-button class="operation-btn operation-btn-primary" size="small" plain @click="envShow(scope.row)">查看env</pl-button>
              </div>
            </div>
            <div class="operation-block">
              <span class="operation-title">快速重启：</span>
              <div class="quick-actions">
                <template v-for="item in scope.row.default_service_list" :key="`restart_${scope.row.id}_${item}`">
                  <pl-button class="quick-action-btn quick-action-restart" size="small" plain @click="restart(scope.row , item)">{{ item }}</pl-button>
                </template>
              </div>
            </div>
            <div class="operation-block">
              <span class="operation-title">快速停止：</span>
              <div class="quick-actions">
                <template v-for="item in scope.row.default_service_list" :key="`stop_${scope.row.id}_${item}`">
                  <pl-button class="quick-action-btn quick-action-stop" size="small" plain @click="stop(scope.row , item)">{{ item }}</pl-button>
                </template>
              </div>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

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
      <pl-button type="primary" link @click="refreshServices()">刷新服务列表</pl-button>
      <el-table :data="dialogServiceConfig.services" style="width: 100%">
        <el-table-column label="服务名" prop="name" width="250"/>
        <el-table-column label="操作">
          <template #default="scope">
            <pl-button link type="primary" @click="restart(dialogServiceConfig , scope.row.name)">restart</pl-button>
            <pl-button link type="primary" @click="stop(dialogServiceConfig , scope.row.name)">stop</pl-button>
            <pl-button link type="primary" @click="start(dialogServiceConfig , scope.row.name)">up</pl-button>
            <pl-button
              v-if="!isDefaultService(dialogServiceConfig, scope.row.name)"
              link
              type="primary"
              :loading="defaultServiceLoadingMap[getDefaultServiceLoadingKey(dialogServiceConfig, scope.row.name)]"
              @click="toggleDefaultService(dialogServiceConfig, scope.row.name, true)"
            >加入默认服务</pl-button>
            <pl-button
              v-else
              link
              type="danger"
              :loading="defaultServiceLoadingMap[getDefaultServiceLoadingKey(dialogServiceConfig, scope.row.name)]"
              @click="toggleDefaultService(dialogServiceConfig, scope.row.name, false)"
            >移除默认服务</pl-button>
<!--          <pl-button link type="primary" @click="status(dialogServiceConfig , scope.row.name)">上传可执行文件并重启</pl-button>-->
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="dialogImageList" :append-to-body="true" title="镜像列表" width="82%">
      <div class="dialog-toolbar">
        <div class="dialog-toolbar-text">当前环境下的全部 Docker 镜像</div>
        <pl-button type="primary" link :loading="imageListLoading" @click="fetchImageList">刷新镜像列表</pl-button>
      </div>
      <el-table :data="imageList" style="width: 100%">
        <el-table-column label="镜像名" min-width="220">
          <template #default="scope">
            <div class="image-name-cell">{{ getImageDisplayName(scope.row) }}</div>
          </template>
        </el-table-column>
        <el-table-column label="镜像 ID" prop="image_id" min-width="180"/>
        <el-table-column label="创建时间" prop="created" min-width="140"/>
        <el-table-column label="大小" prop="size" width="120"/>
        <el-table-column label="操作" min-width="220">
          <template #default="scope">
            <div class="operation-buttons">
              <pl-button class="operation-btn" size="small" plain @click="showImageContainers(scope.row)">查看容器</pl-button>
              <pl-button class="operation-btn operation-btn-danger" size="small" plain @click="confirmRemoveImage(scope.row)">移除镜像</pl-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <el-dialog v-model="dialogImageContainers" :append-to-body="true" :title="imageContainerDialogTitle" width="82%">
      <div class="dialog-toolbar">
        <div class="dialog-toolbar-text">该镜像下的全部容器</div>
        <pl-button type="primary" link :loading="imageContainerLoading" @click="fetchImageContainers">刷新容器列表</pl-button>
      </div>
      <el-table :data="imageContainerList" style="width: 100%">
        <el-table-column label="容器名" prop="container_name" min-width="180"/>
        <el-table-column label="容器 ID" prop="container_id" min-width="160"/>
        <el-table-column label="镜像" prop="image" min-width="180"/>
        <el-table-column label="状态" prop="status" min-width="180"/>
        <el-table-column label="操作" min-width="220">
          <template #default="scope">
            <div class="operation-buttons">
              <pl-button class="operation-btn operation-btn-danger" size="small" plain @click="confirmStopContainer(scope.row)">停止</pl-button>
              <pl-button class="operation-btn operation-btn-danger" size="small" plain @click="confirmRemoveContainer(scope.row)">移除</pl-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <SettingsDialog
      v-model="composeSettingsVisible"
      title="Docker Compose 设置"
      width="82%"
      @closed="refreshComposeAfterSettingsClose"
    >
      <ComposeSettingPage @changed="handleComposeSettingsChanged" />
    </SettingsDialog>
  </div>
</template>
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
import composeSet from "@/utils/base/compose_set";
import dockerDefaultService from "@/utils/docker_default_service.cjs";
import SettingsDialog from '@/components/base/SettingsDialog.vue'
import ComposeSettingPage from '@/components/set/compose.vue'

const TRUNCATE_CONTAINER_LOG_TITLE = '确认清理容器日志'
// TRUNCATE_CONTAINER_LOG_MESSAGE 提示本次清理会作用于当前 SSH 环境的全部 Docker 容器日志。
const TRUNCATE_CONTAINER_LOG_MESSAGE = '确定清理当前环境下全部容器日志吗？该操作会执行 truncate -s 0 /var/lib/docker/containers/*/*-json.log。'
const TRUNCATE_CONTAINER_LOG_SUCCESS = '容器日志已清理'
const TRUNCATE_CONTAINER_LOG_ERROR = '容器日志清理失败'

export default {
  props: {},
  components: {
    shellResult,
    SettingsDialog,
    ComposeSettingPage,
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
      containerLogCleaning: false,
      dialogImageList: false,
      imageList: [],
      imageListLoading: false,
      dialogImageContainers: false,
      imageContainerList: [],
      imageContainerLoading: false,
      currentImageRow: {},
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
      sseThrottleStringFunc: null,
      defaultServiceLoadingMap: {},
      composeSettingsVisible: false,
    }
  },
  inject: ["showTerminal", "resizeTerminal"],
  activated: function () {
    this.resizeTerminal()
  },
  mounted: function () {
    let _that = this
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
      _that.sse_distribute_id = sseDistribute.GetSseDistributeId(`docker_${action}_${Date.now()}`)
      if (!_that.sseThrottleStringFunc) {
        _that.sseThrottleStringFunc = new Throttle_string(50, text => {
          _that.shellController.sshResult += text
          const maxLen = 50000
          if (_that.shellController.sshResult.length > maxLen) {
            _that.shellController.sshResult = _that.shellController.sshResult.slice(-maxLen)
          }
          let result = format.formatResult(
              _that.shellController.sshResult, ['copy', 'color', 'replace'])
          result = format.formatResult(result, ['length'])
          _that.shellController.sshResult = result
        })
      }
      sseDistribute.RegisterReceive(_that.sse_distribute_id, function (msg) {
        _that.sseThrottleStringFunc.update(msg)
      })
      return _that.sse_distribute_id
    },
    // openComposeSettings 打开 Compose 设置弹窗，在 Docker 页面内完成配置维护。
    // Open the compose settings modal so Docker page configuration stays in place.
    openComposeSettings: function () {
      this.composeSettingsVisible = true
    },
    // handleComposeSettingsChanged 配置改动后立即刷新 Compose 项目列表。
    // Reload compose projects immediately after settings change.
    handleComposeSettingsChanged: function () {
      this.getComposeList()
    },
    // refreshComposeAfterSettingsClose 关闭弹窗时再补一次刷新，覆盖更多编辑路径。
    // Refresh once more on dialog close as a fallback for additional edit flows.
    refreshComposeAfterSettingsClose: function () {
      this.getComposeList()
    },
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
    getDefaultServiceList: function (defaultService) {
      return dockerDefaultService.normalizeDockerDefaultServices(defaultService)
    },
    isDefaultService: function (row, serviceName) {
      return dockerDefaultService.isDockerDefaultServiceEnabled(row?.default_service, serviceName)
    },
    getDefaultServiceLoadingKey: function (row, serviceName) {
      return `${row.id}_${serviceName}`
    },
    buildComposeSavePayload: function (row, defaultService) {
      return {
        id: row.id,
        name: row.name,
        compose_yml_path: row.compose_yml_path,
        env_file: row.env_file,
        ssh_id: row.ssh_id,
        docker_cmd: row.docker_cmd,
        default_service: defaultService,
        upload_exes: row.upload_exes || '',
      }
    },
    syncComposeDefaultService: function (row, defaultService) {
      let _that = this
      let normalizedDefaultService = dockerDefaultService.stringifyDockerDefaultServices(defaultService)
      let defaultServiceList = _that.getDefaultServiceList(normalizedDefaultService)
      let applyRow = function (target) {
        if (!target) {
          return
        }
        target.default_service = normalizedDefaultService
        target.default_service_list = defaultServiceList
      }

      applyRow(row)
      applyRow(_that.dialogServiceConfig)
      applyRow(_that.composeList.find(item => parseInt(item.id) === parseInt(row.id)))
    },
    toggleDefaultService: function (row, serviceName, enabled) {
      let _that = this
      let loadingKey = _that.getDefaultServiceLoadingKey(row, serviceName)
      let nextDefaultService = dockerDefaultService.toggleDockerDefaultService(row.default_service, serviceName, enabled)
      if (nextDefaultService === dockerDefaultService.stringifyDockerDefaultServices(row.default_service)) {
        return
      }

      _that.defaultServiceLoadingMap[loadingKey] = true
      composeSet.ComposeAdd(_that.buildComposeSavePayload(row, nextDefaultService), function (response) {
        _that.defaultServiceLoadingMap[loadingKey] = false
        if (response.ErrCode === 0) {
          _that.syncComposeDefaultService(row, nextDefaultService)
          _that.$helperNotify.success(enabled ? '已加入默认服务' : '已移除默认服务')
          return
        }
        _that.$helperNotify.error(response.ErrMsg || '默认服务更新失败')
      })
    },
    refreshServices : function (row){
      let _that = this
      _that.prepareActionSse('services_refresh')
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
            _that.dialogServiceConfig.default_service_list = _that.getDefaultServiceList(_that.dialogServiceConfig.default_service)
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
        _that.dialogServiceConfig.default_service_list = _that.getDefaultServiceList(_that.dialogServiceConfig.default_service)
        return
      }
      _that.prepareActionSse('services_list')
      compose.DockerComposeServices(data, function (response) {
            _that.$helperNotify.success('成功')
            _that.shellController.isRunning = false
            _that.dialogShowService = true
            _that.dialogServiceConfig.services = response.Data.services
            _that.dialogServiceConfig.default_service_list = _that.getDefaultServiceList(_that.dialogServiceConfig.default_service)
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
      _that.prepareActionSse('restart')
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
      _that.prepareActionSse('stop')
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
      _that.prepareActionSse('start')
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
      _that.prepareActionSse('status')
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
      _that.prepareActionSse('show_compose')
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
      _that.prepareActionSse('show_env')
      let envFile = value.env_file
      if (envFile === '') {
        envFile = value.compose_yml_path.replace(/\/[^/]+\.yml$/, '/.env')
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
      _that.prepareActionSse('compose_list')
      compose.DockerComposeList({ssh_id: _that.chooseSshId, sse_distribute_id: _that.sse_distribute_id}, function (response) {
            if (response.ErrCode === 0) {
              _that.composeList = response.Data.list
              for (let i in _that.composeList) {
                _that.composeList[i].show = true
                _that.composeList[i].default_service_list = _that.getDefaultServiceList(_that.composeList[i].default_service)
              }
              // 初始化星标状态
              _that.initStarStatus()
              // 重新拉取列表后保留当前本地搜索条件
              _that.searchList()
            }
            _that.shellController.isRunning = false
          }
      )
    },
    //选择代码环境
    changeSsh: function () {
      let _that = this
      _that.$helperStore.setStore('dockerChooseSshId', _that.chooseSshId)
      _that.getComposeList()
    },
    //搜索消费者列表
    searchList: function () {
      let _that = this
      let ret = search.SearchListObj(_that.composeList, _that.searchKey)
      _that.searchNum = ret[0]
      _that.composeList = ret[1]
    },
    confirmTruncateContainerLogs: function () {
      let _that = this
      _that.$confirm(TRUNCATE_CONTAINER_LOG_MESSAGE, TRUNCATE_CONTAINER_LOG_TITLE, {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(function () {
        _that.containerLogCleaning = true
        compose.DockerContainerLogTruncate({ssh_id: _that.chooseSshId}, function (response) {
          _that.containerLogCleaning = false
          if (response.ErrCode === 0) {
            _that.$helperNotify.success(TRUNCATE_CONTAINER_LOG_SUCCESS)
            return
          }
          _that.$helperNotify.error(response.ErrMsg || TRUNCATE_CONTAINER_LOG_ERROR)
        })
      }).catch(function () {
        return false
      })
    },
    getImageDisplayName: function (row) {
      if (!row) {
        return ''
      }
      if (row.repository && row.repository !== '<none>' && row.tag && row.tag !== '<none>') {
        return `${row.repository}:${row.tag}`
      }
      return row.image_id || row.image_ref || '未命名镜像'
    },
    openImageListDialog: function () {
      this.dialogImageList = true
      this.fetchImageList()
    },
    fetchImageList: function () {
      let _that = this
      if (!_that.chooseSshId) {
        return
      }
      _that.imageListLoading = true
      compose.DockerImageList({ssh_id: _that.chooseSshId}, function (response) {
        _that.imageListLoading = false
        if (response.ErrCode === 0) {
          _that.imageList = response.Data.list || []
          return
        }
        _that.$helperNotify.error(response.ErrMsg || '镜像列表加载失败')
      })
    },
    showImageContainers: function (row) {
      this.currentImageRow = row || {}
      this.dialogImageContainers = true
      this.fetchImageContainers()
    },
    fetchImageContainers: function () {
      let _that = this
      if (!_that.chooseSshId || !_that.currentImageRow.image_ref) {
        return
      }
      _that.imageContainerLoading = true
      compose.DockerImageContainers({
        ssh_id: _that.chooseSshId,
        image_ref: _that.currentImageRow.image_ref,
      }, function (response) {
        _that.imageContainerLoading = false
        if (response.ErrCode === 0) {
          _that.imageContainerList = response.Data.list || []
          return
        }
        _that.$helperNotify.error(response.ErrMsg || '镜像容器列表加载失败')
      })
    },
    confirmRemoveImage: function (row) {
      let _that = this
      _that.$confirm(`确定移除镜像“${_that.getImageDisplayName(row)}”吗？`, '确认移除镜像', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(function () {
        compose.DockerImageRemove({
          ssh_id: _that.chooseSshId,
          image_ref: row.image_ref,
        }, function (response) {
          if (response.ErrCode === 0) {
            _that.$helperNotify.success('镜像已移除')
            _that.fetchImageList()
            return
          }
          _that.$helperNotify.error(response.ErrMsg || '镜像移除失败')
        })
      }).catch(function () {
        return false
      })
    },
    confirmStopContainer: function (row) {
      let _that = this
      _that.$confirm(`确定停止容器“${row.container_name}”吗？`, '确认停止容器', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(function () {
        compose.DockerContainerStop({
          ssh_id: _that.chooseSshId,
          container_id: row.container_id,
        }, function (response) {
          if (response.ErrCode === 0) {
            _that.$helperNotify.success('容器已停止')
            _that.fetchImageContainers()
            return
          }
          _that.$helperNotify.error(response.ErrMsg || '容器停止失败')
        })
      }).catch(function () {
        return false
      })
    },
    confirmRemoveContainer: function (row) {
      let _that = this
      _that.$confirm(`确定移除容器“${row.container_name}”吗？`, '确认移除容器', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning',
      }).then(function () {
        compose.DockerContainerRemove({
          ssh_id: _that.chooseSshId,
          container_id: row.container_id,
        }, function (response) {
          if (response.ErrCode === 0) {
            _that.$helperNotify.success('容器已移除')
            _that.fetchImageContainers()
            _that.fetchImageList()
            return
          }
          _that.$helperNotify.error(response.ErrMsg || '容器移除失败')
        })
      }).catch(function () {
        return false
      })
    },
  },
  computed: {
    imageContainerDialogTitle: function () {
      let title = this.getImageDisplayName(this.currentImageRow)
      if (!title) {
        return '镜像容器'
      }
      return `镜像容器 - ${title}`
    },
  },
}
</script>

<style scoped>
.docker-page-container {
  padding: 0;
  width: 100%;
  color: #4a4a4a;
}

.docker-header-card {
  background: #fff;
  border: 1px solid #e8e8e0;
  border-radius: 12px;
  padding: 16px 18px;
  margin-bottom: 12px;
}

.header-title {
  display: flex;
  align-items: center;
  gap: 8px;
  color: #4a4a4a;
  font-size: 18px;
  font-weight: 600;
  margin-bottom: 12px;
}

.header-icon {
  width: 20px;
  height: 20px;
  color: #5a8a5a;
}

.control-row {
  display: flex;
  gap: 10px;
  align-items: center;
  flex-wrap: wrap;
}

.env-select {
  width: 260px;
}

.env-select :deep(.el-input__wrapper),
.search-input :deep(.el-input__wrapper) {
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 0 0 1px #dde3d8 inset;
}

.env-select :deep(.el-input__wrapper.is-focus),
.search-input :deep(.el-input__wrapper.is-focus) {
  box-shadow: 0 0 0 1px #93b793 inset;
}

.action-buttons {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.action-buttons .el-button {
  border-radius: 8px;
  border: 1px solid #d8ded2;
  background: #f6f8f3;
  color: #4f804f;
}

.action-buttons .el-button:hover {
  background: #eef4ea;
  border-color: #bfd1bf;
  color: #3f6f3f;
}

.search-input {
  flex: 1;
  max-width: 420px;
  min-width: 220px;
}

.header-tail-actions {
  margin-left: auto;
  display: flex;
  justify-content: flex-end;
}

.header-tail-btn {
  border-radius: 8px;
}

.compose-table-card {
  background: #fff;
  border: 1px solid #e8e8e0;
  border-radius: 12px;
  padding: 12px;
}

.compose-table {
  width: 100%;
  font-size: 14px;
  border-radius: 10px;
  overflow: hidden;
}

.compose-table :deep(.el-table__header th) {
  background: #f7f7f2;
  color: #606050;
  font-weight: 600;
}

.compose-table :deep(.el-table__row:hover > td) {
  background-color: #f3f7ef !important;
}

.name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
  min-height: 24px;
}

.name-text {
  line-height: 1.2;
  display: flex;
  align-items: center;
}

.path-text {
  font-family: Consolas, Monaco, monospace;
  font-size: 13px;
  color: #4f804f;
  background: #f3f8ef;
  padding: 2px 8px;
  border-radius: 4px;
  word-break: break-all;
}

.operation-block {
  margin-top: 8px;
  display: flex;
  align-items: flex-start;
  gap: 10px;
  padding: 9px 10px;
  border: 1px solid #e8eee4;
  border-radius: 10px;
  background: #fbfdf9;
}

.operation-block:first-child {
  margin-top: 0;
}

.operation-title {
  flex: 0 0 auto;
  font-weight: 500;
  color: #4a4a4a;
  line-height: 28px;
}

.operation-buttons {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  flex: 1;
}

.operation-btn {
  min-height: 30px;
  padding: 0 12px;
  border-radius: 8px;
  font-weight: 500;
  transition: all 0.18s ease;
}

.operation-btn.operation-btn-primary {
  border-color: #d8ded2;
  color: #4f804f;
  background: #f6f8f3;
}

.operation-btn.operation-btn-primary:hover {
  background: #eef4ea;
  border-color: #bfd1bf;
  color: #3f6f3f;
}

.operation-btn.operation-btn-success {
  border-color: #c1d9ba;
  color: #356a35;
  background: #edf6e9;
}

.operation-btn.operation-btn-success:hover {
  border-color: #9fc49c;
  color: #2d5a2d;
  background: #e2f0dd;
}

.operation-btn.operation-btn-danger {
  border-color: #e6c9be;
  color: #9b523d;
  background: #fcf4f1;
}

.operation-btn.operation-btn-danger:hover {
  border-color: #d8ad9f;
  color: #864434;
  background: #f8e9e4;
}

.quick-actions {
  display: flex;
  align-items: center;
  gap: 6px;
  flex-wrap: wrap;
  flex: 1;
}

.quick-action-btn {
  min-height: 28px;
  border-radius: 8px;
  font-size: 12px;
  font-weight: 500;
  padding: 0 10px;
}

.quick-action-restart {
  border-color: #b7d7b2;
  color: #2f6e37;
  background: #eff8ec;
}

.quick-action-restart:hover {
  border-color: #93be8d;
  color: #285e30;
  background: #e3f2df;
}

.quick-action-stop {
  border-color: #e8ccb9;
  color: #965139;
  background: #fff5ef;
}

.quick-action-stop:hover {
  border-color: #deaf94;
  color: #844530;
  background: #feede3;
}

.dialog-toolbar {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 12px;
  margin-bottom: 12px;
}

.dialog-toolbar-text {
  color: #606050;
  font-size: 13px;
}

.dialog-toolbar-actions {
  display: inline-flex;
  align-items: center;
  gap: 8px;
  flex-wrap: wrap;
}

.image-name-cell {
  line-height: 1.4;
  word-break: break-all;
}

.el-table .warning-row {
  --el-table-tr-bg-color: #fdf6e6;
}

.el-table .success-row {
  --el-table-tr-bg-color: #eef7ea;
}

.el-table .error-row {
  --el-table-tr-bg-color: #fbeeee;
}

.compose-table :deep(.row-hide) {
  display: none;
}

.star-icon {
  cursor: pointer;
  transition: all 0.3s ease;
  flex-shrink: 0;
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

@media (max-width: 1200px) {
  .control-row {
    align-items: stretch;
  }

  .search-input {
    max-width: 100%;
  }

  .header-tail-actions {
    margin-left: 0;
    justify-content: flex-start;
  }

  .operation-block {
    flex-direction: column;
    gap: 8px;
  }

  .operation-title {
    line-height: 1.4;
  }
}
</style>

