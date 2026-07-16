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
          <el-option key="__all__" label="全部" :value="0"></el-option>
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
      <el-table :data="composeList" :row-class-name="getColumnColor" class="compose-table" height="calc(100vh - 220px)">
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
        <el-table-column v-if="!chooseSshId" label="环境" sortable width="120">
          <template #default="scope">
            <span v-html="scope.row.ssh_name"></span>
          </template>
        </el-table-column>
        <el-table-column label="位置" sortable width="260">
          <template #default="scope">
            <code class="path-text" v-html="scope.row.compose_yml_path"></code>
          </template>
        </el-table-column>
        <el-table-column fixed="right" label="操作" min-width="360">
          <template #default="scope">
            <div v-if="scope.row.default_service_list && scope.row.default_service_list.length" class="operation-block">
              <span class="operation-title">快速重启：</span>
              <div class="quick-actions">
                <template v-for="item in scope.row.default_service_list" :key="`restart_${scope.row.id}_${item}`">
                  <pl-button size="small" type="success" plain @click="restart(scope.row , item)">{{ item }}</pl-button>
                </template>
              </div>
            </div>
            <div v-if="scope.row.default_service_list && scope.row.default_service_list.length" class="operation-block">
              <span class="operation-title">快速停止：</span>
              <div class="quick-actions">
                <template v-for="item in scope.row.default_service_list" :key="`stop_${scope.row.id}_${item}`">
                  <pl-button size="small" type="warning" plain @click="stop(scope.row , item)">{{ item }}</pl-button>
                </template>
              </div>
            </div>
            <div v-if="scope.row.default_service_list && scope.row.default_service_list.length" class="operation-block">
              <span class="operation-title">快速 down：</span>
              <div class="quick-actions">
                <template v-for="item in scope.row.default_service_list" :key="`down_${scope.row.id}_${item}`">
                  <pl-button size="small" type="danger" plain @click="down(scope.row , item)">{{ item }}</pl-button>
                </template>
              </div>
            </div>
            <div v-if="scope.row.default_service_list && scope.row.default_service_list.length" class="operation-block">
              <span class="operation-title">快速 pull：</span>
              <div class="quick-actions">
                <template v-for="item in scope.row.default_service_list" :key="`pull_${scope.row.id}_${item}`">
                  <pl-button size="small" type="primary" plain @click="pull(scope.row , item)">{{ item }}</pl-button>
                </template>
              </div>
            </div>
          </template>
        </el-table-column>
        <el-table-column fixed="right" label="更多操作" width="120">
          <template #default="scope">
            <el-dropdown @command="(command) => handleComposeRowActionCommand(scope.row, command)">
              <pl-button size="small" type="primary" plain class="operation-dropdown-trigger">
                更多操作
                <el-icon class="el-icon--right"><ArrowDown /></el-icon>
              </pl-button>
              <template #dropdown>
                <el-dropdown-menu>
                  <el-dropdown-item :command="COMPOSE_ROW_ACTION_SERVICE_LIST">服务列表</el-dropdown-item>
                  <el-dropdown-item :command="COMPOSE_ROW_ACTION_STATUS">运行状态</el-dropdown-item>
                  <el-dropdown-item :command="COMPOSE_ROW_ACTION_SHOW_CONFIG" divided>查看 compose.yml</el-dropdown-item>
                  <el-dropdown-item :command="COMPOSE_ROW_ACTION_SHOW_ENV">查看 env</el-dropdown-item>
                  <el-dropdown-item :command="COMPOSE_ROW_ACTION_RESTART" divided>重启（restart）</el-dropdown-item>
                  <el-dropdown-item :command="COMPOSE_ROW_ACTION_STOP">停止（stop）</el-dropdown-item>
                  <el-dropdown-item :command="COMPOSE_ROW_ACTION_START">启动（up -d）</el-dropdown-item>
                  <el-dropdown-item :command="COMPOSE_ROW_ACTION_DOWN">拆除（down）</el-dropdown-item>
                  <el-dropdown-item :command="COMPOSE_ROW_ACTION_PULL">拉取（pull）</el-dropdown-item>
                </el-dropdown-menu>
              </template>
            </el-dropdown>
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
            <pl-button link type="danger" @click="down(dialogServiceConfig , scope.row.name)">down</pl-button>
            <pl-button link type="primary" @click="pull(dialogServiceConfig , scope.row.name)">pull</pl-button>
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
              <pl-button size="small" type="primary" plain @click="showImageContainers(scope.row)">查看容器</pl-button>
              <pl-button size="small" type="danger" plain @click="confirmRemoveImage(scope.row)">移除镜像</pl-button>
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
              <pl-button size="small" type="warning" plain @click="confirmStopContainer(scope.row)">停止</pl-button>
              <pl-button size="small" type="danger" plain @click="confirmRemoveContainer(scope.row)">移除</pl-button>
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
// COMPOSE_ROW_ACTION_* 统一定义表格行下拉动作命令，避免模板里散落魔法字符串。
// COMPOSE_ROW_ACTION_* centralizes row dropdown commands to avoid magic strings in the template.
const COMPOSE_ROW_ACTION_SERVICE_LIST = 'service_list'
const COMPOSE_ROW_ACTION_STATUS = 'status'
const COMPOSE_ROW_ACTION_SHOW_CONFIG = 'show_config'
const COMPOSE_ROW_ACTION_SHOW_ENV = 'show_env'
const COMPOSE_ROW_ACTION_RESTART = 'restart'
const COMPOSE_ROW_ACTION_STOP = 'stop'
const COMPOSE_ROW_ACTION_START = 'start'
const COMPOSE_ROW_ACTION_DOWN = 'down'
const COMPOSE_ROW_ACTION_PULL = 'pull'

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
      chooseSshId: 0,
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
      COMPOSE_ROW_ACTION_SERVICE_LIST,
      COMPOSE_ROW_ACTION_STATUS,
      COMPOSE_ROW_ACTION_SHOW_CONFIG,
      COMPOSE_ROW_ACTION_SHOW_ENV,
      COMPOSE_ROW_ACTION_RESTART,
      COMPOSE_ROW_ACTION_STOP,
      COMPOSE_ROW_ACTION_START,
      COMPOSE_ROW_ACTION_DOWN,
      COMPOSE_ROW_ACTION_PULL,
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
          _that.chooseSshId = _that.getLastSshId()
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
    // handleComposeRowActionCommand 统一分发项目级下拉操作，保持模板简洁且便于后续扩展。
    // handleComposeRowActionCommand routes compose row dropdown actions to keep the template compact and extensible.
    handleComposeRowActionCommand: function (row, command) {
      // switch 分支只做动作分发，不改变原有接口调用语义。
      // The switch only dispatches actions and preserves existing API behavior.
      switch (command) {
        case COMPOSE_ROW_ACTION_SERVICE_LIST:
          this.dialogServices(row)
          return
        case COMPOSE_ROW_ACTION_STATUS:
          this.status(row)
          return
        case COMPOSE_ROW_ACTION_SHOW_CONFIG:
          this.configShow(row)
          return
        case COMPOSE_ROW_ACTION_SHOW_ENV:
          this.envShow(row)
          return
        case COMPOSE_ROW_ACTION_RESTART:
          this.restart(row)
          return
        case COMPOSE_ROW_ACTION_STOP:
          this.stop(row)
          return
        case COMPOSE_ROW_ACTION_START:
          this.start(row)
          return
        case COMPOSE_ROW_ACTION_DOWN:
          this.down(row)
          return
        case COMPOSE_ROW_ACTION_PULL:
          this.pull(row)
          return
        default:
          return
      }
    },
    prepareActionSse: function (action) {
      let _that = this
      if (_that.sse_distribute_id) {
        sseDistribute.UnRegisterReceive(_that.sse_distribute_id)
      }
      _that.sse_distribute_id = sseDistribute.GetSseDistributeId(`docker`)
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
    // stripHtmlTags 清除搜索高亮产生的 HTML 标签，防止保存时将富文本写入数据库。
    stripHtmlTags: function (value) {
      if (typeof value !== 'string') return value
      return value.replace(/<[^>]*>/g, '')
    },
    buildComposeSavePayload: function (row, defaultService) {
      return {
        id: row.id,
        name: this.stripHtmlTags(row.name),
        compose_yml_path: this.stripHtmlTags(row.compose_yml_path),
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
      let sshId = _that.getSshId(_that.dialogServiceConfig)
      //优先从缓存拿
      let servicesKey = 'docker_services_' + sshId + '_' + _that.dialogServiceConfig.id
      let data = {
        ssh_id: sshId,
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
      let sshId = _that.getSshId(row)
      let data = {
        ssh_id: sshId,
        id: row.id,
        sse_distribute_id: _that.sse_distribute_id,
      }
      //优先从缓存拿
      let servicesKey = 'docker_services_' + sshId + '_' + row.id
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
      if (chooseSshId === null || chooseSshId === undefined || chooseSshId === '' || isNaN(chooseSshId)) {
        return 0
      }
      chooseSshId = parseInt(chooseSshId)
      if (chooseSshId === 0) {
        return 0
      }
      for (let i in _that.sshList) {
        if (parseInt(_that.sshList[i].id) === chooseSshId) {
          return chooseSshId
        }
      }
      return 0
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
        ssh_id: _that.getSshId(value),
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
        ssh_id: _that.getSshId(value),
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
        ssh_id: _that.getSshId(value),
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
    down: function (value , service) {
      let _that = this
      _that.shellController.isRunning = true
      _that.prepareActionSse('down')
      let data = {
        ssh_id: _that.getSshId(value),
        id: value.id,
        sse_distribute_id: _that.sse_distribute_id,
        service : service,
      }
      compose.DockerComposeDown(data, function (response) {
            _that.$helperNotify.success('成功')
            _that.getComposeList()
            _that.shellController.isRunning = false
          }
      )
    },
    pull: function (value , service) {
      let _that = this
      _that.shellController.isRunning = true
      _that.prepareActionSse('pull')
      let data = {
        ssh_id: _that.getSshId(value),
        id: value.id,
        sse_distribute_id: _that.sse_distribute_id,
        service : service,
      }
      compose.DockerComposePull(data, function (response) {
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
        ssh_id: _that.getSshId(value),
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
        ssh_id: _that.getSshId(value),
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
        ssh_id: _that.getSshId(value),
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
    getSshId: function (row) {
      return parseInt(this.chooseSshId) || row.ssh_id
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

<style scoped src="@/css/components/Docker.css"></style>

