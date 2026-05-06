<template>
  <!-- 本地客户端模式状态卡片 -->
  <el-alert
      v-if="runtimeConfig.run_mode === 'local_client'"
      :closable="false"
      :show-icon="true"
      :type="clientStatusType"
      style="margin-bottom: 8px;"
  >
    <template #title>{{ clientStatusTitle }}</template>
    <div>{{ clientStatusMessage }}</div>
    <div v-if="runtimeConfig.run_mode === 'local_client' && !clientStatus.client_connected" style="margin-top: 8px;">
      <el-button type="primary" size="small" :loading="isClientDownloadBusy('windows')" @click="downloadClient('windows')">{{ getClientDownloadButtonText('windows') }}</el-button>
      <el-button type="primary" size="small" :loading="isClientDownloadBusy('macos')" @click="downloadClient('macos')">{{ getClientDownloadButtonText('macos') }}</el-button>
      <el-button size="small" @click="refreshClientStatus">刷新状态</el-button>
    </div>
  </el-alert>

  <el-alert v-if="is_install === 1 && runtimeConfig.run_mode === 'server'" :closable="false" show-icon title="正在安装中，看网速大约5-20分钟" type="warning"/>
  <el-alert
      v-if="node_install_tip.show && runtimeConfig.run_mode === 'server'"
      :closable="false"
      show-icon
      type="error"
      style="margin-bottom: 8px;"
  >
    <template #title>未检测到 Node.js，当前无法使用自定义网页</template>
    <div>{{ node_install_tip.install_tip }}</div>
    <el-link :href="node_install_tip.install_url" target="_blank" type="primary" style="margin-top: 4px;">
      前往下载 Node.js
    </el-link>
  </el-alert>
  <div class="link-run-page">
    <div class="link-run-header-card">
      <div class="link-run-header-title">
        <div class="link-run-header-title__main">自定义网页</div>
        <div class="link-run-header-title__desc">集中管理页面入口、运行方式和流程跳转，顶部操作区独立展示更利于快速切换。</div>
      </div>
      <div class="link-run-toolbar">
        <el-tag size="small" type="info" effect="light">已打开 Page {{ openPageNum }}</el-tag>
        <GitActionButton variant="warning" @click="openAccountSettings">
          <el-icon><User /></el-icon>账号设置
        </GitActionButton>
        <GitActionButton @click="showCreateDialog">
          <el-icon><Plus /></el-icon>创建
        </GitActionButton>
        <template v-if="runtimeConfig.run_mode === 'server'">
          <GitActionButton @click="install">
            <el-icon><Tools /></el-icon>安装核心
          </GitActionButton>
          <GitActionButton variant="warning" @click="recycle">
            <el-icon><Refresh /></el-icon>释放内存
          </GitActionButton>
          <GitActionButton variant="info" @click="downloadPath">
            <el-icon><Download /></el-icon>下载目录
          </GitActionButton>
          <GitActionButton variant="info" @click="openDataDir">
            <el-icon><FolderOpened /></el-icon>数据存储
          </GitActionButton>
        </template>
        <GitActionButton variant="info" @click="drawerVisibleMarkdown = true">
          <el-icon><QuestionFilled /></el-icon>帮助文档
        </GitActionButton>
      </div>
    </div>
    <div class="link-run-content">
      <!--      <pl-button type="primary" @click="showDialogRunLog">运行日志({{shellController.sshResult.length}})</pl-button>&nbsp;-->
      <!--      <el-link type="primary" @click="showMarkdown">使用说明</el-link>-->
      <div v-for="(smartValue, smartLinkIndex) in smartList" :key="smartLinkIndex" class="link-run-card">
        <a style="display: inline-block;text-decoration: underline;cursor:pointer;font-size:17px;font-weight: bold;" @click="showEditDialog(smartValue)">
          {{ smartValue.id + " " + smartValue.name }}
        </a>
        <el-tooltip content="编辑" placement="top">
          <el-icon size="small" style="margin-left:20px;" @click="showEditDialog(smartValue)">
            <Setting/>
          </el-icon>
        </el-tooltip>
        <el-tooltip content="展示账号密码" placement="top">
          <el-icon size="small" style="margin: 10px;" @click="showUserPasswordList(smartValue)">
            <Notebook/>
          </el-icon>
        </el-tooltip>
        <el-tooltip content="删除" placement="top">
          <el-popconfirm
              cancel-button-text="取消"
              confirm-button-text="删除"
              icon-color="#626AEF"
              title="确定删除吗?"
              @confirm="deleteSmartLink(smartValue)"
          >
            <template #reference>
              <el-icon size="small">
                <Delete/>
              </el-icon>
            </template>
          </el-popconfirm>
        </el-tooltip>
        <el-row :gutter="20" class="link-run-links-row">
          <el-col v-for="(linkValue, linkIndex) in smartValue.linkList" :key="linkIndex" :span="4">
            <div class="grid-content bg-purple">
              <!--            选择后内置核心打开-->
              <template v-if="(linkValue.userList && linkValue.userList.length > 0) || parseInt(smartValue.open_num) > 0">
                <!--              供选择的环境列表-->
                <el-radio v-model="smartValue.chooseSmartLinkIndex" :label="linkValue.label"
                          @change="changeChooseLink(smartLinkIndex , linkIndex)">
                  {{ linkValue.label }}

                  <span v-if="linkValue.runNum" style="font-size: 12px;color:green;">({{ linkValue.runNum }})</span>
                </el-radio>
              </template>

              <!--            直接打开-->
              <template v-if="!linkValue.userList && parseInt(smartValue.open_type) === 1 && parseInt(smartValue.open_num) === 0">
                <el-link style="padding: 10px;" type="primary" @click="redirectLink(linkValue)">
                  {{ linkValue.label }}
                </el-link>
              </template>

              <!--            内置核心打开-->
              <template v-if="(!linkValue.userList || linkValue.userList.length === 0) && (parseInt(smartValue.open_type) === 2 || parseInt(smartValue.open_type) === 3) && parseInt(smartValue.open_num) === 0">
                <el-link style="padding: 10px;" type="primary" @click="smartLinkRun(smartLinkIndex,linkIndex)">
                  {{ linkValue.label }}
                  <span v-if="linkValue.runNum" style="font-size: 12px;color:green;">
                    ({{ linkValue.runNum }})
                  </span>
                </el-link>
              </template>

            </div>
          </el-col>
        </el-row>
        <!--      账号列表-->
        <el-form v-if="smartValue.linkList[smartValue.chooseLinkIndex] &&
          (smartValue.linkList[smartValue.chooseLinkIndex].userList || smartValue.open_num > 0 )" :inline="true" class="demo-form-inline"
                 label-width="auto" style="margin: 0 auto;">
          <el-form-item v-if="smartValue.linkList[smartValue.chooseLinkIndex].userList && smartValue.linkList[smartValue.chooseLinkIndex].userList.length > 0" label="账号列表">
            <el-select v-model="smartValue.linkList[smartValue.chooseLinkIndex].chooseUserName" placeholder="选择账号">
              <template v-for="(user,userkey) in smartValue.linkList[smartValue.chooseLinkIndex].userList" :key="userkey">
                <el-option :label="user.user_name" :value="user.user_name"/>
              </template>
            </el-select>
          </el-form-item>
          <el-form-item v-if="smartValue.open_type === 2" label="打开方式">
            <el-select v-model="smartValue.open_type_new" placeholder="选择类型">
              <template v-for="(value,key) in openTypeList" :key="key">
                <el-option :label="value.label" :value="value.value"/>
              </template>
            </el-select>
          </el-form-item>
          <el-form-item v-if="smartValue.open_num > 0" label="打开数">
            <el-input v-model="smartValue.open_num_new" placeholder="Please input" style="width: 240px"/>
          </el-form-item>
          <el-form-item>
            <GitActionButton v-if="smartValue.linkList[smartValue.chooseLinkIndex].chooseUserName || smartValue.open_num > 0" @click="smartLinkRun(smartLinkIndex,null)">
              执行
            </GitActionButton>
          </el-form-item>
        </el-form>
      </div>
    </div>
  </div>
  <!--新增弹窗-->
  <el-dialog v-model="dialogSmartLink" title="创建/编辑链接" width="90%" class="smart-link-dialog">
    <el-form label-width="auto" class="smart-link-dialog__form">
      <el-form-item label="名称">
        <el-input v-model="smartLinkConfig.name"/>
      </el-form-item>
      <el-form-item label="类型">
        <el-select v-model="smartLinkConfig.open_type" placeholder="选择类型">
          <template v-for="(value,key) in openTypeList" :key="key">
            <el-option :label="value.label" :value="value.value"/>
          </template>
        </el-select>
      </el-form-item>
      <el-form-item v-if="parseInt(smartLinkConfig.open_type) !== 1" label="浏览器">
        <el-alert :closable="false" show-icon title="如果选择chrome，那么支持播放视频（其他区别还没发现）但是耗费更多内存，会卡一些" type="info"/>
        <el-select v-model="smartLinkConfig.channel" placeholder="选择类型">
          <template v-for="(value,key) in channelList" :key="key">
            <el-option :label="value.label" :value="value.value"/>
          </template>
        </el-select>
      </el-form-item>
      <el-form-item label="自动关闭">
        <el-alert :closable="false" show-icon title="多少秒之内没有操作，页面自动关闭，0表示无限" type="info"/>
        <el-input v-model="smartLinkConfig.auto_close_second" type="text"></el-input>
      </el-form-item>
      <el-form-item label="打开次数">
        <el-alert :closable="false" show-icon title="如果只需要打开一个，那么设置为0，否则设置为1，运行时会出现输入框自定义" type="info"/>
        <el-input v-model="smartLinkConfig.open_num" type="text"></el-input>
      </el-form-item>
      <!--      <el-form-item label="下载匹配项">-->
      <!--        <el-alert title="哪些请求路由会被定义为下载，英文逗号分割" type="info" show-icon :closable="false"/>-->
      <!--        <el-input v-model="smartLinkConfig.download_finds" type="textarea" :rows="5"/>-->
      <!--      </el-form-item>-->
      <el-form-item label="执行逻辑">
        <el-alert :closable="false" show-icon title="打开链接后执行的流程，切换到编辑执行逻辑页面，可查看执行逻辑" type="info"/>
        <el-select v-model="smartLinkConfig.process_id" placeholder="选择执行逻辑">
          <template v-for="(value,key) in processList" :key="key">
            <el-option :label="value.name" :value="value.id"/>
          </template>
        </el-select>
      </el-form-item>
      <el-form-item v-if="dialogSmartLink" label="链接配置" class="smart-link-dialog__link-config">
        <LinkConfigEditor v-model="smartLinkConfig" />
      </el-form-item>
      <el-form-item label="排序值">
        <el-input v-model="smartLinkConfig.weight" type="text"/>
      </el-form-item>
    </el-form>
    <template class="dialog-footer">
      <GitActionButton @click="dialogSmartLink = false">取 消</GitActionButton>
      <GitActionButton @click="saveSmartLink">确 定</GitActionButton>
    </template>
  </el-dialog>

  <shellResult ref="shellRef" :btnName="'运行日志'" :isRunning="shellController.isRunning" :shellShowResult="shellController.sshResult" :show-model="shellController.showModel"></shellResult>
  <el-dialog v-model="dialogShowUserPass" title="账号密码列表" width="90%">
    <el-input v-model="userPassSearchKeyword" clearable placeholder="搜索环境、用户名或密码" style="margin-bottom: 12px" />
    <el-table
        :data="filteredUserPassList"
        border
        highlight-current-row
        stripe
        style="width: 100%"
    >
      <el-table-column
          label="环境"
          prop="label"
      ></el-table-column>
      <el-table-column
          label="用户名"
          prop="username"
      ></el-table-column>
      <el-table-column
          label="密码"
          prop="password"
      ></el-table-column>
    </el-table>
  </el-dialog>

  <el-drawer
      v-model="drawerVisibleMarkdown"
      direction="rtl"
      size="90%"
      title="文档"
  >
    <Markdown v-if="drawerVisibleMarkdown" :markdownType="markdownType"></Markdown>
  </el-drawer>

  <SettingsDialog
      v-model="accountSettingsVisible"
      title="账号设置"
      width="82%"
      @closed="refreshLinkAfterAccountSettingsClose"
  >
    <AccountSettingPage @changed="handleAccountSettingsChanged" />
  </SettingsDialog>
</template>
<style scoped src="@/css/components/smart_link/link_run.css"></style>
<script>
import smart_link_set from "@/utils/base/smart_link_set"
import base from "@/utils/base";
import ticker_step from "@/utils/base/ticker_step"
import t from "@/utils/base/type"
import Markdown from "@/components/Markdown.vue";
import Init from "@/utils/base/set_init";
import Process from '@/utils/base/smart_link_proces'
import shellResult from "@/components/shell/result_button.vue";
import sse from "@/utils/base/sse";
import sseDistribute from "@/utils/base/sse_distribute";
import LinkConfigEditor from "@/components/smart_link/LinkConfigEditor.vue";
import GitActionButton from "@/components/base/GitActionButton.vue";
import SettingsDialog from '@/components/base/SettingsDialog.vue'
import AccountSettingPage from '@/components/set/account.vue'
import { Plus, Tools, Refresh, Download, QuestionFilled, Setting, Notebook, Delete, User, FolderOpened } from '@element-plus/icons-vue'

const { mergeSavedSmartLinkIntoList } = require('@/utils/smart_link_config_sync.cjs')
const { DEFAULT_RUNTIME_CONFIG, buildRuntimeApiUrl, buildRuntimeRequestOptions, resolveRuntimeRefreshActions } = require('@/utils/smart_link_runtime.cjs')
const { buildDownloadUrlWithToken } = require('@/utils/download_url.cjs')

export default {
  props: {
    shellShowResult: {
      type: String
    },
  },
  components: {
    shellResult,
    Markdown,
    Plus,
    Tools,
    Refresh,
    Download,
    QuestionFilled,
    Setting,
    Notebook,
    Delete,
    User,
    FolderOpened,
    LinkConfigEditor,
    GitActionButton,
    SettingsDialog,
    AccountSettingPage,
  },
  data() {
    return {
      shellController: {
        sshResult: '',
        sourceSshResult: '',
        isRunning: false,
        showModel: 'button',
        divHeight: 330,
      },
      drawerVisibleMarkdown: false,
      markdownType: 'Link',
      dialogShowMarkdown: false,
      dialogShowUserPass: false,
      dialogSsePushLog: false,
      showUserPassList: [],
      userPassSearchKeyword: '',
      dialogSmartLink: false,
      openTypeList: [
        {label: '通过js直接打开', value: 1},
        {label: '静默打开(内置核心打开)', value: 2},
        {label: '浏览器打开(内置核心打开)', value: 3}
      ],
      channelList: [
        {label: '请选择', value: ''},
        {label: 'chrome(完整浏览器功能)', value: 'chrome'},
        {label: 'chromium(内存占用低)', value: 'chromium'},
      ],
      sse_distribute_id : '',
      processList: [],
      smartLinkConfig: {
        id: 0,
        name: '',
        links: '',
        open_num: 0,
        open_type: '',
        status: '',
        combine_type: 4,
        weight: 0,
        download_finds: '',
        channel: '',
        show_cookies: '',
        process_id: 0,
      },
      defaultSmartLinkConfig: {
        id: 0,
        name: '',
        links: '',
        open_num: 0,
        open_type: '',
        status: '',
        combine_type: 4,
        weight: 0,
        download_finds: '',
        channel: '',
        show_cookies: '',
        process_id: 0,
      },
      name: 'Link',
      smartList: [],
      smartValue: {},
      smartLinkProcessList: [],
      smartLinkRunList: {},
      tickerKey: 'link',
      versionInfo: {},
      //当前选中的配置
      chooseSmartLinkIndex: 0, //选中的第几个配置
      //已打开数量
      openPageNum: 0,
      //是否在安装中
      is_install: 0,
      // Node.js 安装提示
      node_install_tip: {
        show: false,
        install_url: 'https://nodejs.org/zh-cn/download',
        install_tip: '请先安装 Node.js（建议 LTS 版本），安装完成后刷新当前页面。',
      },
      accountSettingsVisible: false,
      // 本地客户端模式相关
      runtimeConfig: {...DEFAULT_RUNTIME_CONFIG},
      clientStatus: {
        client_connected: false,
        client_status: 'offline',
        client_name: '',
        client_version: '',
        client_version_match: false,
        client_last_seen_at: 0,
        client_os: '',
        client_arch: ''
      },
      clientDownloadStates: {
        windows: {
          status: 'idle',
          text: '下载 Windows 客户端',
          progress: 0,
          jobId: '',
          timerId: null,
        },
        macos: {
          status: 'idle',
          text: '下载 macOS 客户端',
          progress: 0,
          jobId: '',
          timerId: null,
        },
      },
    }
  },
  computed: {
    clientStatusType() {
      if (this.runtimeConfig.run_mode !== 'local_client') return 'info'
      if (this.clientStatus.client_connected && this.clientStatus.client_version_match) return 'success'
      if (this.clientStatus.client_status === 'preparing_runtime') return 'warning'
      return 'error'
    },
    clientStatusTitle() {
      if (this.runtimeConfig.run_mode !== 'local_client') return '服务端执行模式'
      if (this.clientStatus.client_connected && this.clientStatus.client_version_match) return '本地客户端在线'
      if (this.clientStatus.client_status === 'preparing_runtime') return '运行环境准备中'
      if (this.clientStatus.client_status === 'version_mismatch') return '客户端版本不匹配'
      return '本地客户端未连接'
    },
    clientStatusMessage() {
      if (this.runtimeConfig.run_mode !== 'local_client') {
        return '当前使用服务端 Playwright 执行自定义网页'
      }
      if (this.clientStatus.client_connected && this.clientStatus.client_version_match) {
        return `本地客户端 ${this.clientStatus.client_name} 在线，版本 ${this.clientStatus.client_version}，可执行自定义网页`
      }
      if (this.clientStatus.client_status === 'preparing_runtime') {
        return '本地客户端正在准备运行环境，请稍后重试'
      }
      if (this.clientStatus.client_status === 'version_mismatch') {
        return `当前客户端版本为 ${this.clientStatus.client_version}，要求版本为 ${this.runtimeConfig.required_client_version}，请重新下载并启动`
      }
      return '当前已启用本地客户端执行，但未检测到客户端连接，请下载安装并启动本地客户端'
    },
    canExecute() {
      // 本地客户端模式下检查客户端状态
      if (this.runtimeConfig.run_mode === 'local_client') {
        return this.clientStatus.client_connected && this.clientStatus.client_version_match
      }
      // 服务端模式下检查 Node.js
      return !this.node_install_tip.show
    },
    filteredUserPassList() {
      const keyword = this.userPassSearchKeyword.trim().toLowerCase()
      if (!keyword) return this.showUserPassList
      return this.showUserPassList.filter(item =>
        (item.label || '').toLowerCase().includes(keyword) ||
        (item.username || '').toLowerCase().includes(keyword) ||
        (item.password || '').toLowerCase().includes(keyword)
      )
    },
  },
  mounted: function () {
    this.sse_distribute_id = sseDistribute.GetSseDistributeId('link')
    this.sseCreate()
    this.init()
    this.refreshRuntimeConfigState()
    // 注册 SSE 客户端状态推送
    sseDistribute.RegisterReceive('smart_link_client_status', this.handleClientStatusSSE)
  },
  beforeUnmount() {
    sseDistribute.UnRegisterReceive('smart_link_client_status')
    this.clearClientDownloadPoll('windows')
    this.clearClientDownloadPoll('macos')
  },
  activated() {
    if (Init.GetIsInit('smart_link') === true) {
      let _that = this
      _that.init()
      Init.DelInit('smart_link')
    }
    // 页面从设置页切回时需要重新读取运行模式，否则会继续显示旧的 server 模式。
    // Reload runtime mode when the page is re-activated so the UI does not stay on stale server mode after settings changes.
    this.refreshRuntimeConfigState()
  },
  methods: {
    sseCreate: function () {
      let _that = this
      sseDistribute.RegisterReceive(_that.sse_distribute_id , function (msg,msgType,sseDistributeId){
        if (msg === sse.SseEventClean) {
          _that.shellController.sshResult = ''
          _that.shellController.sourceSshResult = '';
        } else if (msg === sse.SseEventLogin) {
          _that.dialogLoginUserName = true
        } else if (msg.startsWith(sse.SseEventProcess)) { //准备替换
          _that.replaceRegex = msg.replace(sse.SseEventProcess, '')
        } else {
          _that.shellController.sourceSshResult += msg
          _that.shellController.sshResult = _that.shellController.sourceSshResult
        }
      })
    },
    init: function () {
      let _that = this
      _that.GetProcessList()
      _that.GetConfigList()
      let _height = base.GetDivHeight()
      _that.windowChange()
      _that.tickerRunList()
      _that.SmartLinkChromeVersion()
      setTimeout(function () {
        let _height = base.GetDivHeight2()
        _that.shellController.divHeight = parseInt(_height) - 60
        _that.windowChange()
      }, 1000)
    },
    // openAccountSettings 打开账号设置弹窗，在自定义网页页内维护账号与分组。
    // Open the account settings modal so account and group maintenance stays inside the custom web page.
    openAccountSettings: function () {
      this.accountSettingsVisible = true
    },
    // handleAccountSettingsChanged 账号配置变化后刷新自定义网页配置列表，让账号选择立即生效。
    // Refresh smart link configs after account settings change so account selections take effect immediately.
    handleAccountSettingsChanged: function () {
      this.GetConfigList()
    },
    // refreshLinkAfterAccountSettingsClose 在弹窗关闭时再刷新一次，兜底覆盖更多修改路径。
    // Refresh once more when the modal closes as a fallback for additional account edit flows.
    refreshLinkAfterAccountSettingsClose: function () {
      this.GetConfigList()
    },
    // applyNodeInstallTip 解析并展示 Node.js 安装提示
    applyNodeInstallTip: function (response) {
      let _that = this
      let data = response && response.Data ? response.Data : {}
      let needInstall = data.need_install_node === 1
      _that.node_install_tip.show = needInstall
      if (needInstall) {
        _that.node_install_tip.install_url = data.install_url || 'https://nodejs.org/zh-cn/download'
        _that.node_install_tip.install_tip = data.install_tip || '请先安装 Node.js（建议 LTS 版本），安装完成后刷新当前页面。'
      }
      return needInstall
    },
    SmartLinkChromeVersion: function () {
      let _that = this
      smart_link_set.SmartLinkChromeVersion(_that.sse_distribute_id , function (response) {
        if (response.ErrCode === 0) {
          _that.versionInfo = response.Data.version
          _that.is_install = response.Data.is_install
          _that.applyNodeInstallTip(response)
        } else {
          if (!_that.applyNodeInstallTip(response)) {
            _that.$helperNotify.error('失败')
          }
        }
      })
    },
    changeChooseLink: function (smartLinkIndex, linkIndex) {
      let _that = this
      _that.chooseSmartLinkIndex = smartLinkIndex
      _that.smartList[smartLinkIndex].chooseLinkIndex = linkIndex
      ticker_step.Active(_that.tickerKey)
    },
    // loadRuntimeConfig 拉取最新运行模式。
    // loadRuntimeConfig fetches the latest run mode.
    loadRuntimeConfig: function () {
      let _that = this
      return fetch(
        buildRuntimeApiUrl(base.GetApiHost(), '/api/smart-link/runtime-config'),
        buildRuntimeRequestOptions(base.GetSafeToken())
      )
        .then(res => res.json())
        .then(data => {
          if (data.ErrCode === 0 && data.Data) {
            const nextState = resolveRuntimeRefreshActions(_that.runtimeConfig, data.Data)
            _that.runtimeConfig = nextState.runtimeConfig
            // 本地客户端模式下，runtimeConfig 加载完毕后立即拉取一次客户端状态，避免 SSE 推送时序问题。
            // Immediately fetch client status after runtimeConfig is loaded in local_client mode.
            if (nextState.shouldLoadClientStatus) {
              _that.refreshClientStatus()
            }
          }
        })
        .catch(() => {
          // 使用默认值
        })
    },
    // refreshRuntimeConfigState 在页面初始化和重新激活时同步运行模式。
    // refreshRuntimeConfigState keeps runtime mode in sync on mount and when the kept-alive page is activated again.
    refreshRuntimeConfigState: function () {
      return this.loadRuntimeConfig()
    },
    // handleClientStatusSSE 处理 SSE 推送的客户端状态。
    handleClientStatusSSE: function (data) {
      if (data && this.runtimeConfig.run_mode === 'local_client') {
        this.clientStatus = data
      }
    },
    // 刷新客户端状态（手动触发一次 HTTP 拉取兜底）
    refreshClientStatus: function () {
      let _that = this
      if (_that.runtimeConfig.run_mode !== 'local_client') return
      fetch(
        buildRuntimeApiUrl(base.GetApiHost(), '/api/smart-link/client-status'),
        buildRuntimeRequestOptions(base.GetSafeToken())
      )
        .then(res => res.json())
        .then(data => {
          if (data.ErrCode === 0 && data.Data) {
            _that.clientStatus = data.Data
          }
        })
        .catch(() => {})
      _that.$message.success('状态已刷新')
    },
    // getClientDownloadDefaultText 返回不同平台下载按钮的默认文案。
    // getClientDownloadDefaultText returns the default download button label for each platform.
    getClientDownloadDefaultText: function (platform) {
      return platform === 'windows' ? '下载 Windows 客户端' : '下载 macOS 客户端'
    },
    // getClientDownloadButtonText 根据当前任务状态展示实时进度文案。
    // getClientDownloadButtonText renders live task progress text for the download button.
    getClientDownloadButtonText: function (platform) {
      const state = this.clientDownloadStates[platform]
      if (!state || !state.text) {
        return this.getClientDownloadDefaultText(platform)
      }
      return state.text
    },
    // isClientDownloadBusy 判断当前平台是否处于编译或下载中。
    // isClientDownloadBusy reports whether the platform-specific download task is active.
    isClientDownloadBusy: function (platform) {
      const state = this.clientDownloadStates[platform]
      if (!state) return false
      return ['pending', 'building', 'ready', 'downloading'].includes(state.status)
    },
    // setClientDownloadState 更新按钮展示状态，避免模板里散落复杂判断。
    // setClientDownloadState updates the per-platform button state so template logic stays simple.
    setClientDownloadState: function (platform, status, text, progress, extras) {
      const currentState = this.clientDownloadStates[platform]
      if (!currentState) return
      this.clientDownloadStates[platform] = {
        ...currentState,
        ...(extras || {}),
        status: status || currentState.status,
        text: text || currentState.text,
        progress: typeof progress === 'number' ? progress : currentState.progress,
      }
    },
    clearClientDownloadPoll: function (platform) {
      const state = this.clientDownloadStates[platform]
      if (!state || !state.timerId) return
      clearTimeout(state.timerId)
      this.clientDownloadStates[platform].timerId = null
    },
    // resolveClientBuildHost 解析需要注入到客户端里的默认服务端地址。
    // resolveClientBuildHost resolves the server host that should be embedded into the downloaded client.
    resolveClientBuildHost: function () {
      const apiHost = String(base.GetApiHost() || '').trim()
      if (apiHost) {
        return apiHost
      }
      if (window && window.location && window.location.origin) {
        return window.location.origin
      }
      return ''
    },
    // scheduleClientBuildStatusPoll 统一管理轮询节奏，避免重复 setTimeout 叠加。
    // scheduleClientBuildStatusPoll centralizes polling cadence so repeated setTimeout calls do not stack.
    scheduleClientBuildStatusPoll: function (platform, jobId) {
      this.clearClientDownloadPoll(platform)
      this.clientDownloadStates[platform].timerId = setTimeout(() => {
        this.pollClientBuildStatus(platform, jobId)
      }, 800)
    },
    // pollClientBuildStatus 轮询后端编译任务状态，并驱动按钮文案实时变化。
    // pollClientBuildStatus polls backend build progress and drives the live button label updates.
    pollClientBuildStatus: function (platform, jobId) {
      const _that = this
      fetch(
        buildRuntimeApiUrl(base.GetApiHost(), `/api/smart-link/client-build/status?job_id=${encodeURIComponent(jobId)}`),
        buildRuntimeRequestOptions(base.GetSafeToken())
      )
        .then(res => res.json())
        .then(data => {
          if (data.ErrCode !== 0 || !data.Data) {
            throw new Error(data.ErrMsg || '获取编译状态失败')
          }
          const statusData = data.Data
          const progressText = statusData.message || _that.getClientDownloadDefaultText(platform)
          _that.setClientDownloadState(platform, statusData.status, `${progressText}${statusData.progress >= 0 ? ` (${statusData.progress}%)` : ''}`, statusData.progress, {
            jobId,
          })
          // 失败后立即停止轮询，避免错误提示和状态更新重复触发。
          // Stop polling immediately on failure so error toasts and state transitions do not repeat.
          if (statusData.status === 'failed') {
            _that.clearClientDownloadPoll(platform)
            _that.$message.error(statusData.error || statusData.message || '编译失败')
            return
          }
          // 任务 ready 后切到二进制下载；completed 兜底兼容后端已提前结束的状态。
          // Switch to binary download when ready; completed is kept as a fallback for already-finished backend states.
          if (statusData.status === 'ready' || statusData.status === 'completed') {
            _that.clearClientDownloadPoll(platform)
            _that.downloadBuiltClient(platform, jobId, statusData.file_name)
            return
          }
          _that.scheduleClientBuildStatusPoll(platform, jobId)
        })
        .catch(err => {
          _that.clearClientDownloadPoll(platform)
          _that.setClientDownloadState(platform, 'failed', '编译失败', 100)
          _that.$message.error(err.message || '获取编译状态失败')
        })
    },
    // downloadBuiltClient 下载已编译好的二进制文件，并触发浏览器保存。
    // downloadBuiltClient downloads the compiled binary artifact and triggers the browser save flow.
    downloadBuiltClient: function (platform, jobId, fallbackFileName) {
      const _that = this
      const requestUrl = buildDownloadUrlWithToken(
        buildRuntimeApiUrl(base.GetApiHost(), `/api/smart-link/client-build/download/${encodeURIComponent(jobId)}`),
        base.GetSafeToken()
      )
      const xhr = new XMLHttpRequest()
      xhr.open('GET', requestUrl, true)
      xhr.responseType = 'blob'
      if (base.GetSafeToken()) {
        xhr.setRequestHeader('Token', base.GetSafeToken())
      }
      _that.setClientDownloadState(platform, 'downloading', '正在下载客户端', 100, { jobId })
      xhr.onprogress = function (event) {
        // 浏览器能提供长度时展示实时百分比，否则退化为通用“正在下载”文案。
        // Show live percentage when the browser exposes content length, otherwise fall back to a generic downloading message.
        if (event.lengthComputable && event.total > 0) {
          const percent = Math.min(100, Math.round((event.loaded / event.total) * 100))
          _that.setClientDownloadState(platform, 'downloading', `正在下载客户端 (${percent}%)`, 100, { jobId })
          return
        }
        _that.setClientDownloadState(platform, 'downloading', '正在下载客户端', 100, { jobId })
      }
      xhr.onload = function () {
        if (xhr.status < 200 || xhr.status >= 300) {
          _that.setClientDownloadState(platform, 'failed', '下载失败', 100, { jobId })
          _that.$message.error('客户端下载失败')
          return
        }
        const headerFileName = xhr.getResponseHeader('X-Download-Filename')
        const fileName = headerFileName || fallbackFileName || _that.getClientDownloadDefaultText(platform)
        const blobUrl = window.URL.createObjectURL(xhr.response)
        const a = document.createElement('a')
        a.href = blobUrl
        a.download = fileName
        document.body.appendChild(a)
        a.click()
        a.remove()
        window.URL.revokeObjectURL(blobUrl)
        _that.setClientDownloadState(platform, 'completed', '下载完成', 100, { jobId })
        setTimeout(() => {
          _that.setClientDownloadState(platform, 'idle', _that.getClientDownloadDefaultText(platform), 0, { jobId: '' })
        }, 2000)
      }
      xhr.onerror = function () {
        _that.setClientDownloadState(platform, 'failed', '下载失败', 100, { jobId })
        _that.$message.error('客户端下载失败')
      }
      xhr.send()
    },
    // downloadClient 创建编译任务并开始轮询，是按钮点击后的主入口。
    // downloadClient creates the build job and starts polling; it is the main button click entry point.
    downloadClient: function (platform) {
      const _that = this
      // 未知平台直接拦截，避免前端状态机和后端参数校验出现双重噪音。
      // Reject unknown platforms early so both the frontend state machine and backend validation stay clean.
      if (!_that.clientDownloadStates[platform]) {
        _that.$message.error('不支持的客户端平台')
        return
      }
      // 正在编译或下载时不重复提交，避免产生多个并发任务抢占同一个按钮状态。
      // Do not submit again while a job is active, otherwise concurrent tasks would fight over one button state.
      if (_that.isClientDownloadBusy(platform)) {
        return
      }
      const host = _that.resolveClientBuildHost()
      // host 会编译进客户端，因此拿不到有效地址时必须直接终止。
      // host is compiled into the client, so the flow must stop if no valid server address can be resolved.
      if (!host) {
        _that.$message.error('当前服务端地址不可用')
        return
      }
      _that.setClientDownloadState(platform, 'pending', '准备编译参数 (5%)', 5, { jobId: '' })
      fetch(
        buildRuntimeApiUrl(base.GetApiHost(), '/api/smart-link/client-build/start'),
        buildRuntimeRequestOptions(base.GetSafeToken(), {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify({
            platform,
            host,
          })
        })
      )
        .then(res => res.json())
        .then(data => {
          if (data.ErrCode !== 0 || !data.Data) {
            throw new Error(data.ErrMsg || '创建编译任务失败')
          }
          const jobId = data.Data.job_id
          _that.setClientDownloadState(platform, data.Data.status || 'pending', `${data.Data.message || '准备编译参数'} (${data.Data.progress || 0}%)`, data.Data.progress || 0, {
            jobId,
          })
          _that.scheduleClientBuildStatusPoll(platform, jobId)
        })
        .catch(err => {
          _that.setClientDownloadState(platform, 'failed', '编译失败', 100)
          _that.$message.error(err.message || '创建编译任务失败')
        })
    },
    smartLinkRun: function (smartLinkIndex, linkIndex) {
      let _that = this

      // 本地客户端模式下检查执行权限
      if (_that.runtimeConfig.run_mode === 'local_client') {
        if (!_that.canExecute) {
          _that.$message.error('本地客户端未连接或版本不匹配，无法执行')
          return
        }
        // 本地客户端模式：创建任务
        _that.createLocalClientTask(smartLinkIndex, linkIndex)
        return
      }

      if (smartLinkIndex !== null && smartLinkIndex !== undefined && linkIndex === null) {
        _that.chooseSmartLinkIndex = smartLinkIndex
        smartLinkIndex = _that.chooseSmartLinkIndex
        linkIndex = _that.smartList[_that.chooseSmartLinkIndex].chooseLinkIndex
      }

      let chooseSmartLink = _that.smartList[smartLinkIndex]
      let chooseLink = chooseSmartLink.linkList[linkIndex]

      let chooseUser = {}
      for (let i in chooseLink.userList) {
        if (chooseLink.userList[i].user_name === chooseLink.chooseUserName) {
          chooseUser = chooseLink.userList[i]
        }
      }
      let runParams = {
        id: chooseSmartLink.id,
        label: chooseLink.label,
        user_name: chooseUser.user_name,
        password: chooseUser.password,
        open_num: chooseSmartLink.open_num_new,
        open_type : chooseSmartLink.open_type_new,
        sse_distribute_id : _that.sse_distribute_id,
      }
      smart_link_set.SmartLinkRun(runParams, function (response) {
        if (response.ErrCode !== 0) {
          if (!_that.applyNodeInstallTip(response)) {
            _that.$helperNotify.error(response.ErrMsg || '执行失败')
          }
          return
        }
        ticker_step.Active(_that.tickerKey)
      });
    },
    // 创建本地客户端任务
    createLocalClientTask: function (smartLinkIndex, linkIndex) {
      let _that = this
      if (smartLinkIndex !== null && smartLinkIndex !== undefined && linkIndex === null) {
        _that.chooseSmartLinkIndex = smartLinkIndex
        smartLinkIndex = _that.chooseSmartLinkIndex
        linkIndex = _that.smartList[_that.chooseSmartLinkIndex].chooseLinkIndex
      }

      let chooseSmartLink = _that.smartList[smartLinkIndex]
      let chooseLink = chooseSmartLink.linkList[linkIndex]

      let chooseUser = {}
      for (let i in chooseLink.userList) {
        if (chooseLink.userList[i].user_name === chooseLink.chooseUserName) {
          chooseUser = chooseLink.userList[i]
        }
      }

      let taskData = {
        smart_link_id: chooseSmartLink.id,
        label: chooseLink.label,
        user_name: chooseUser.user_name || '',
        password: chooseUser.password || '',
        open_num: chooseSmartLink.open_num_new,
      }

      fetch(
        buildRuntimeApiUrl(base.GetApiHost(), '/api/smart-link/task/create'),
        buildRuntimeRequestOptions(base.GetSafeToken(), {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json'
          },
          body: JSON.stringify(taskData)
        })
      )
        .then(res => res.json())
        .then(data => {
          if (data.ErrCode === 0) {
            _that.$message.success('任务已创建，等待本地客户端执行')
            ticker_step.Active(_that.tickerKey)
          } else {
            _that.$message.error(data.ErrMsg || '创建任务失败')
          }
        })
        .catch(err => {
          _that.$message.error('创建任务失败: ' + err.message)
        })
    },
    tickerRunList: function () {
      let _that = this
      ticker_step.Register(_that.tickerKey, 5, function () {
        _that.runList()
      })
    },
    processMsg: function (msg) {
      let _that = this
      //处理进度
      const regReceivingObjects = new RegExp(_that.replaceRegex);

      let regList = [regReceivingObjects]
      let boolFind = false
      for (let i in regList) {
        let reg = regList[i]
        let matchList = msg.match(reg); //收到的消息是否匹配到
        if (t.IsArray(matchList) && matchList.length > 0) {
          let strMatchList = _that.shellController.sourceSshResult.match(reg)
          if (t.IsArray(strMatchList) && strMatchList.length > 0) {
            _that.shellController.sourceSshResult = _that.shellController.sourceSshResult.replace(reg, matchList[0])
            boolFind = true
          }
        }
      }
      if (!boolFind) {
        _that.shellController.sourceSshResult += msg
      }
    },
    runList: function () {
      let _that = this
      smart_link_set.SmartLinkRunList(_that.sse_distribute_id , function (response) {
        if (response.ErrCode !== 0) {
          _that.applyNodeInstallTip(response)
          return
        }
        let runList = response.Data
        _that.openPageNum = 0
        _that.smartLinkRunList = {};
        for (let i in runList) {
          if (_that.smartLinkRunList[runList[i].name]) {
            _that.smartLinkRunList[runList[i].name] += runList[i].page_num
          } else {
            _that.smartLinkRunList[runList[i].name] = runList[i].page_num
          }
          _that.openPageNum += runList[i].page_num
        }
        //赋值
        for (let i in _that.smartList) {
          for (let j in _that.smartList[i].linkList) {
            _that.smartList[i].linkList[j].runNum = 0
            for (let runName in _that.smartLinkRunList) {
              if (runName === "link_id_"+_that.smartList[i].id + "_label_" + _that.smartList[i].linkList[j].label) {
                _that.smartList[i].linkList[j].runNum = _that.smartLinkRunList[runName]
              }
            }
          }
        }
      });
    },
    saveSmartLink: function () {
      let _that = this
      _that.smartLinkConfig.linkList = JSON.parse(_that.smartLinkConfig.links || '[]')
      _that.smartLinkConfig.combine_type = 4
      smart_link_set.SmartLinkAdd(_that.smartLinkConfig, function (response) {
        if (response.ErrCode === 0) {
          _that.smartList = mergeSavedSmartLinkIntoList(_that.smartList, response.Data)
          _that.dialogSmartLink = false
        } else {
          _that.$helperNotify.error('失败')
        }
        ticker_step.Active(_that.tickerKey)
      })
    },
    showEditDialog: function (smartLink) {
      let _that = this
      if (smartLink !== undefined) {
        _that.smartLinkConfig = JSON.parse(JSON.stringify(smartLink))
      }
      _that.dialogSmartLink = true
    },
    showUserPasswordList: function (smartLink) {
      let linkList = smartLink.linkList
      let showList = []
      if (linkList && t.IsArray(linkList)) {
        for (let i in linkList) {
          //浏览器自带验证
          if (linkList[i].browser_auth_username) {
            showList.push({
              label: linkList[i].label,
              username: linkList[i].browser_auth_username,
              password: linkList[i].browser_auth_password
            })
          }
          //自己设置的账号密码
          let userList = linkList[i].userList
          if (userList && t.IsArray(userList)) {
            for (let j in userList) {
              showList.push({
                label: linkList[i].label,
                username: userList[j].user_name,
                password: userList[j].password
              })
            }
          }
        }
      }
      this.showUserPassList = showList
      this.dialogShowUserPass = true
    },
    deleteSmartLink: function (smartLink) {
      let _that = this
      smart_link_set.SmartLinkDelete(smartLink, function (response) {
        if (response.ErrCode === 0) {
          _that.GetConfigList()
        } else {
          _that.$helperNotify.error('失败')
        }
      })
    },
    showMarkdown: function () {
      let _that = this
      _that.dialogShowMarkdown = true
    },
    changeToProcess: function () {
      let _that = this
      _that.$emit('changeModelToEditProcess')
    },
    changeToFlow : function () {
      let _that = this
      _that.$emit('changeModelToFlow')
    },
    showDialogRunLog : function (){
      let _that = this
      _that.dialogSsePushLog = true
      setTimeout(function (){
        // shell.ShellDivToBottom(true)
      },500)
    },
    downloadPath: function () {
      let _that = this
      smart_link_set.SmartLinkDownloadPath(_that.sse_distribute_id , function (response) {
        if (response.ErrCode !== 0) {
          if (!_that.applyNodeInstallTip(response)) {
            _that.$helperNotify.error('失败')
          }
        }
      })
    },
    openDataDir: function () {
      let _that = this
      smart_link_set.SmartLinkOpenDataDir(function (response) {
        if (response.ErrCode !== 0) {
          _that.$helperNotify.error(response.ErrMsg || '打开失败')
        }
      })
    },
    install: function () {
      let _that = this
      smart_link_set.SmartLinkChromeUpdate(_that.sse_distribute_id , function (response) {
        if (response.ErrCode === 0) {
          _that.GetConfigList()
          _that.runList()
        } else {
          if (!_that.applyNodeInstallTip(response)) {
            _that.$helperNotify.error('失败')
          }
        }
      })
    },
    recycle: function () {
      let _that = this
      smart_link_set.SmartLinkRecycle(_that.sse_distribute_id , function (response) {
        if (response.ErrCode === 0) {
          _that.GetConfigList()
          _that.runList()
        } else {
          if (!_that.applyNodeInstallTip(response)) {
            _that.$helperNotify.error('失败')
          }
        }
      })
    },
    showCreateDialog: function () {
      let _that = this
      _that.smartLinkConfig = JSON.parse(JSON.stringify(_that.defaultSmartLinkConfig))
      _that.dialogSmartLink = true
    },
    GetConfigList: function () {
      let _that = this
      smart_link_set.SmartLinkList(function (response) {
        if (response.ErrCode === 0) {
          _that.smartList = response.Data.smart_link_list
          _that.formatSmartList()
        } else {
          _that.$helperNotify.error('失败')
        }
      })
    },
    GetProcessList: function () {
      let _that = this
      Process.SmartProcessList(function (response) {
        if (response.ErrCode === 0) {
          _that.processList = response.Data.list
        } else {
          _that.$helperNotify.error('获取执行列表失败')
        }
      })
    },
    windowChange: function () {
      let _that = this
      window.addEventListener('resize', function () {
        let _height = base.GetDivHeight2()
        _that.shellController.divHeight = parseInt(_height) - 60
      });
    },
    formatSmartList: function () {
      let _that = this
      _that.chooseSmartLinkIndex = '0' //默认第一个
      for (let i in _that.smartList) {
        _that.smartList[i].linkList = JSON.parse(_that.smartList[i].links)
        _that.smartList[i].open_num_new = _that.smartList[i].open_num
        _that.smartList[i].open_type_new = _that.smartList[i].open_type
        //排序
        // _that.smartList[i].linkList = arr.SortByKey(this.smartList[i].linkList, 'value', 'asc')
        //默认选中第一个的第一个
        if (parseInt(i) === 0) {
          _that.smartList[i].chooseSmartLinkIndex = _that.smartList[i].linkList[0].value //默认选中第一个的第一个
          _that.smartList[i].chooseLinkIndex = 0 //默认选中第一个的第一个的账号列表
        }
        //如果有userList 那么默认第一个
        for (let j in _that.smartList[i].linkList) {
          if (_that.smartList[i].linkList[j].userList && _that.smartList[i].linkList[j].userList.length > 0 && !_that.smartList[i].linkList[j].chooseUserName) {
            _that.smartList[i].linkList[j].chooseUserName = _that.smartList[i].linkList[j].userList[0].user_name
          }
        }
      }
    },
    redirectLink: function (linkValue) {
      let _that = this
      window.open(linkValue.link, '_blank')
      ticker_step.Active(_that.tickerKey)
    }
  }
}
</script>

<style scoped>
.link-run-page {
  min-height: calc(100vh - 110px);
  color: #4a4a4a;
}

.link-run-header-card {
  background: #fff;
  border: 1px solid #e8e8e0;
  border-radius: 12px;
  padding: 16px 18px;
  margin-bottom: 12px;
}

.link-run-header-title {
  margin-bottom: 12px;
}

.link-run-header-title__main {
  color: #4a4a4a;
  font-size: 18px;
  font-weight: 600;
}

.link-run-header-title__desc {
  margin-top: 6px;
  color: #74806f;
  font-size: 13px;
  line-height: 1.6;
}

.link-run-content {
  padding: 0 2px 2px;
}

.link-run-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
}

.link-run-card {
  min-height: 70px;
  padding: 14px 14px 12px;
  margin-bottom: 12px;
  background: #fff;
  border: 1px solid #e6e8de;
  border-radius: 10px;
  box-sizing: border-box;
}

.link-run-links-row {
  margin-top: 15px;
  margin-bottom: 10px;
}

.smart-link-dialog :deep(.el-dialog__body) {
  padding-top: 18px;
}

.smart-link-dialog__form {
  width: 100%;
}

.smart-link-dialog__link-config {
  width: 100%;
}

.smart-link-dialog__link-config :deep(.el-form-item__content) {
  width: 100%;
  display: block;
}
</style>


