<template>
  <div class="shell-console">
    <div id="mainCard" ref="mainCard" v-if="parseInt(urlParams.id) !== 0 && getExecutionInfo(urlParams.id)" class="execution-info" style="display: flex; align-items: center; gap: 8px;">
      <h4>{{urlParams.title}}</h4>
      <el-popover
          placement="top-start"
          trigger="click"
          width="300px"
      >
        <!-- 触发按钮 -->
        <template #reference>
          <el-button plain size="small" type="primary">查看命令</el-button>
        </template>

        <!-- 气泡内容 -->
        <div class="command-popover">
          <h4>完整命令</h4>
          <pre class="full-command">{{ getExecutionInfo(urlParams.id).command }}</pre>
          <div class="command-actions">
            <el-button
                size="small"
                type="primary"
                @click="copyCommand(getExecutionInfo(urlParams.id).command)"
            >
              复制命令
            </el-button>
          </div>
        </div>
      </el-popover>
      <el-button
          :disabled="getErrorCount(urlParams.id) === 0"
          size="small"
          type="danger"
          @click="showErrorDialog(urlParams.id)"
      >
        {{ getErrorCount(urlParams.id) }} 个错误
      </el-button>
      <el-button
          size="small"
          type="info"
          @click="showFilterDialog(urlParams.id)"
      >
        {{ getFilterCount(urlParams.id) }} 个过滤
      </el-button>
      <el-button
          :disabled="getErrorCount(urlParams.id) === 0"
          size="small"
          @click="clearErrors(urlParams.id)"
      >
        清空错误
      </el-button>
<!--      <el-button-->
<!--          size="small"-->
<!--          @click="removeTab(urlParams.id)"-->
<!--      >-->
<!--        删除-->
<!--      </el-button>-->
      <el-button
          size="small"
          type="primary"
          @click="restartTab(urlParams.id)"
      >
        重启
      </el-button>
<!--      <el-button-->
<!--          size="small" @click="startByTabId(urlParams.id)"-->
<!--      >-->
<!--        启动-->
<!--      </el-button>-->
<!--      <el-button-->
<!---->
<!--          size="small" @click="stopByTabId(urlParams.id)"-->
<!--      >-->
<!--        停止-->
<!--      </el-button>-->
<!--      <el-button-->
<!--          size="small" @click="restartByTabId(urlParams.id)"-->
<!--      >-->
<!--        重启-->
<!--      </el-button>-->
      <el-button
          size="small" @click="cleanLog(urlParams.id)"
      >
        清除日志
      </el-button>
      <el-button
          size="small" @click="up()"
      >
        {{isReceive ? '暂停接收' : '开始接收'}}
      </el-button>
      <!--            <el-tag style="margin: 5px;">链接：{{ getSshName(tab.ssh_id) }}</el-tag>-->
      <el-tag style="margin: 5px;">内容长度：{{ getContentLength(activeTabId) }}</el-tag>
      <el-input
          v-model="searchContent"
          placeholder="输入搜索，多个之间用##分隔"
          size="small"
          style="width: auto; min-width: 200px;"
      />
      <el-button size="small" @click="searchByContent">搜索</el-button>
    </div>
    <!-- 输出区 -->
    <shellResult ref="shellRef" :divHeight="shellController.divHeight" :isRunning="shellController.isRunning"
                 :shellShowResult="contentMapList[activeTabId]" :show-model="shellController.showModel"></shellResult>

    <!-- Error list dialog -->
    <el-dialog
        v-model="errorDialogVisible"
        :title="`错误列表 - ${currentErrorTabName}`"
        width="80%"
    >
      <div class="error-list">
        <div
            v-for="(error) in errorMapList[activeTabId]"
            :key="error.line_number"
            class="error-item card-item"
            :data-line="error.line_number"
        >
          <div class="error-header">
            <span class="error-time">{{ error.time }}</span>
            <el-tag type="danger" size="small" effect="plain">Error</el-tag>
          </div>
          <div class="error-content">
            <span style="line-height:1.6" v-html="highlightErrors(error.error_line)"></span>
          </div>
          <div class="error-actions">
            <el-button
                type="primary"
                link
                size="small"
                @click="getErrorContent(error.line_number)"
                class="context-btn"
            >
              <span class="btn-icon">📋</span>
              查看上下文
            </el-button>
          </div>
        </div>
        <div v-if="activeTabId > 0 && errorMapList[activeTabId].length === 0" class="no-errors">
          <div class="no-data-icon">✅</div>
          <div>暂无错误信息</div>
        </div>
      </div>
      <template #footer>
        <el-button @click="errorDialogVisible = false">关闭</el-button>
        <el-button
            :disabled="errorMapList[activeTabId].length === 0"
            type="danger"
            @click="clearErrors(activeTabId)"
        >
          清空错误
        </el-button>
      </template>
    </el-dialog>

    <!-- 过滤信息弹窗 -->
    <el-dialog
        v-model="filterDialogVisible"
        :title="`过滤列表`"
        width="80%"
    >
      <el-table :data="getFilterList()" style="width: 100%">
        <el-table-column label="名称" prop="name" width="200"/>
        <el-table-column label="正则" prop="key"/>
        <el-table-column label="过滤次数" prop="number" width="120"/>
      </el-table>
    </el-dialog>

    <!-- Error context dialog -->
    <el-dialog
        v-model="errorDialogContextVisible"
        title="上下文(20行)"
        width="80%"
    >
      <div class="error-list context-list">
        <div
            v-for="(line) in errorContext"
            :key="line.line_number"
            class="context-item"
            :class="{ 'highlight-line': line.line_number === parseInt(errorLine) }"
        >
          <span class="line-number">##{{line.line_number}}##</span>
          <span class="line-content" v-html="highlightErrors(line.content, [])"></span>
        </div>
      </div>
      <template #footer>
        <el-button @click="errorDialogContextVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <!-- 搜索信息弹窗 -->
    <el-dialog
        v-model="searchDialogVisible"
        :title="`搜索列表 - ${searchContent} - 搜索到${searchNumber}条`"
        width="80%"
    >
      <div class="error-list">
        <div
            v-for="(search) in searchContents"
            :key="search.LineNumber"
            class="search-item"
            :style="{ 'background-color': search.IsRead ? '#fafafa' : '#ffffff' , border: !search.IsRead ? '1px solid #f56c6c' : '0px'}"
            :data-line="search.LineNumber"
        >
          <div class="search-content">
            <span style="line-height:1.6" v-html="highlightErrors(search.Content , searchContent.split('##'))"></span>
          </div>
          <div class="search-actions">
            <el-button 
                type="primary" 
                link 
                size="small"
                @click="getErrorContent(search.LineNumber)"
                class="context-btn"
            >
              <span class="btn-icon">📋</span>
              查看上下文
            </el-button>
          </div>
        </div>
        <div v-if="!searchContents || searchContents.length === 0" class="no-errors">
          <div class="no-data-icon">🔍</div>
          <div>未搜索到信息</div>
        </div>
      </div>
      <template #footer>
        <el-button @click="searchDialogVisible = false">关闭</el-button>
      </template>
    </el-dialog>

    <el-dialog
        v-model="groupDialog"
        title="分组"
        width="80%"
    >
      <Group :extra1Title="'过滤正则'" :extra1Type="'textarea'"
             :extra2Title="'错误捕获正则'" :extra2Type="'textarea'"
             :extra3Title="'排除捕获的错误'" :extra3Type="'textarea'" :groupTitle="'终端输出'" :groupType="groupType" @update="groupUpdate"></Group>
    </el-dialog>
  </div>
</template>

<script>
/* 以下 import 保持你原来的即可 */
import base from '@/utils/base.js'
import sse from '@/utils/base/sse'
import shell from '@/utils/base/shell'
import ssh from '@/utils/base/ssh_set'
import {ref, onMounted, nextTick} from 'vue'
import copy from '@/utils/base/copy'
import Init from "@/utils/base/set_init";
import shellOut from "@/utils/base/shell_out"
import format from "@/utils/base/format";
import shellResult from "@/components/shell/result_div.vue";
import type from "@/utils/base/type"
import Group from "@/components/group/group_list.vue"
import group from "@/utils/base/group"
import store from "@/utils/base/store"
import sseDistribute from "@/utils/base/sse_distribute";
import {Throttle_string} from "@/utils/base/throttle_string"
import {useRoute} from 'vue-router';
import Typ from "@/utils/base/type";


const StoreChooseGroupIdKey = 'shell_out_choose_group_id'
const StoreChooseShellOutKey = 'shell_out_choose_shell_group'
export default {
  components: {shellResult, Group},
  data() {
    return {
      shellController: {
        sshResult: '',
        sourceSshResult: '',
        isRunning: false,
        showModel: 'div',
        divHeight: 250,
      },
      groupDialog: false,
      sshList: [],
      chooseGroupId: '',
      //编辑 能够编辑的项
      editTabConfigData: {
        id: 0,
        ssh_id: '',
        group_id: '',
        command: '',
        name: '',
      },
      activeTabId: '',
      scrollMap: {},
      shellInstances: new Map(),

      // 错误弹窗相关
      errorDialogVisible: false,
      filterDialogVisible: false,
      currentErrorTabId: '',
      currentErrorTabName: '',

      errorMapList: [], //错误捕获
      filterMapList: [], //过滤
      contentMapList: [], //输出内容
      throttleStringFunc: [],//节流回调
      tabConfigList: [],//配置
      groupList: [], //分组列表
      groupType: `6`,
      errorDialogContextVisible: false,
      searchDialogVisible: false,
      errorContext: '',
      errorLine: '',
      searchContent: '',
      searchContents: [],
      searchNumber: 0,
      urlParams: {},
      isReceive : true, //是否接受
    }
  },
  mounted() {
    let _that = this
    _that.loadSshList()
    _that.windowChange()
    window.addEventListener('resize', function () {
      _that.windowChange()
    });
    setTimeout(function () {
      shell.calculateShellDivHeight(_that)
    } , 1000)
    _that.getGroupList()
    _that.getFullPageParams()
    //如果是单独展示的页面 里面返回的就是传参的
    _that.chooseGroupId = _that.getStoreGroupId()
    _that.activeTabId = _that.getStoreActiveTabId()
    let storeContent = store.getStore('search_content_' + _that.activeTabId)
    if(type.IsString(storeContent)){
      _that.searchContent = storeContent
    }
  },
  activated: function () {
    let _that = this
    _that.windowChange()
  },
  deactivated() {

  },
  methods: {
    // 获取当前 URL 参数
    getFullPageParams: function () {
      let route = useRoute();
      this.urlParams = route.query; // {group_id: "1", id: "3" , title : "xxx"}
      if(this.urlParams.title){
        document.title = this.urlParams.title
      }
      if(!this.urlParams || !this.urlParams.id){
        this.urlParams.id = 0
      }
    },
    getFilterList: function () {
      let filters = [];
      let _that = this
      if (!_that.filterMapList[_that.activeTabId]) {
        return []
      }
      for (let i in _that.filterMapList[_that.activeTabId]) {
        let keyParams = i.split('#')
        let number = _that.filterMapList[_that.activeTabId][i]
        filters.push({
          name: keyParams[0],
          number: number,
          key: keyParams[1]
        })
      }
      return filters
    },
    getStoreGroupId: function () {
      let _that = this
      //地址栏传参
      if (_that.urlParams.group_id) {
        return _that.urlParams.group_id + ''
      }
      //从缓存找到活跃的组
      let storeGroupId = store.getStore(StoreChooseGroupIdKey) == null ? 0 : '' + store.getStore(StoreChooseGroupIdKey)
      //组列表空时返回空
      if (!_that.groupList || _that.groupList.length === 0) {
        return ''
      }
      //如果是全部分组
      if (parseInt(storeGroupId) === -1) {
        return '-1'
      }
      for (let i in _that.groupList) {
        if (parseInt(storeGroupId) === parseInt(_that.groupList[i].id)) {
          return storeGroupId + ''
        }
      }
      return _that.groupList[0].id + ''
    },
    getStoreActiveTabId: function () {
      let _that = this
      if (_that.urlParams.id) {
        return _that.urlParams.id + ''
      }
      //从本地缓存找到活跃的tab
      let storeActiveTabId = store.getStore(StoreChooseShellOutKey + _that.chooseGroupId) == null ? '' : '' + (store.getStore(StoreChooseShellOutKey + _that.chooseGroupId))
      //没有任何配置时直接返回空
      if (!_that.tabConfigList || _that.tabConfigList.length === 0) {
        return ''
      }
      //从配置列表找到活跃的tab
      for (let i in _that.tabConfigList) {
        if (parseInt(storeActiveTabId) === parseInt(_that.tabConfigList[i].id)) {
          return storeActiveTabId + ''
        }
      }
      for (let i in _that.tabConfigList) {
        if (parseInt(_that.chooseGroupId) === 0) {
          return _that.tabConfigList[0].id + ''
        }
        if (parseInt(_that.chooseGroupId) === parseInt(_that.tabConfigList[i].group_id)) {
          return _that.tabConfigList[i].id + ''
        }
      }
      return 0
    },
    getContentLength: function (activeTabId) {
      let _that = this
      return _that.contentMapList[activeTabId] ? _that.contentMapList[activeTabId].length : 0
    },
    groupUpdate: function () {
      this.getGroupList()
    },
    getGroupList: function () {
      let _that = this
      group.GroupList({type: _that.groupType}, function (response) {
        if (response.ErrCode === 0) {
          _that.groupList = response.Data
          _that.chooseGroupId = _that.getStoreGroupId()
          _that.activeTabId = _that.getStoreActiveTabId()
        }
      })
    },
    loadShellOuts() {
      let _that = this
      shellOut.ShellOuts({}, function (res) {
        if (res.ErrCode !== 0) {
          _that.$helperNotify.error('失败')
        } else {
          _that.initTabsFromLocal(res.Data)
        }
      })
    },
    // 复制命令到剪贴板
    copyCommand(command) {
      let index = copy.SetCopyContent(command)
      copy.handleCopy(index)
    },
    searchByContent: function () {
      let _that = this
      let tabConfig = _that.getTabConfigById(_that.activeTabId)
      _that.searchDialogVisible = true
      store.setStore('search_content_' + _that.activeTabId , _that.searchContent)
      shellOut.ShellOutSearchContent({
        'shell_client_id': tabConfig['shell_client_id'],
        'search_content': _that.searchContent,
      }, function (res) {
        if (res.ErrCode === 0) {
          _that.searchDialogVisible = true
          _that.searchContents = res.Data.lines
          _that.searchNumber = res.Data.number
        } else {
          _that.$helperNotify.error('失败')
        }
      })
    },
    getErrorContent: function (errorLine) {
      let _that = this
      let tabConfig = _that.getTabConfigById(_that.activeTabId)
      _that.errorLine = errorLine
      shellOut.ShellOutErrorContext({
        'shell_client_id': tabConfig['shell_client_id'],
        'error_line': errorLine,
      }, function (res) {
        if (res.ErrCode === 0) {
          _that.errorDialogContextVisible = true
          _that.errorContext = res.Data.lines
          
          // Scroll to highlighted line after DOM update
          nextTick(() => {
            setTimeout(() => {
              const highlightedElement = document.querySelector('.context-list .highlight-line')
              if (highlightedElement) {
                highlightedElement.scrollIntoView({
                  behavior: 'smooth',
                  block: 'center',
                  inline: 'nearest'
                })
              }
            }, 100)
          })
        } else {
          _that.$helperNotify.error('失败')
        }
      })
    },
    getCurrentGroupConfig: function () {
      let _that = this
      for (let i in _that.groupList) {
        if (parseInt(_that.chooseGroupId) === parseInt(_that.groupList[i].id)) {
          return _that.groupList[i]
        }
      }
    },
    // Highlight error keywords
    highlightErrors(text, keywords) {
      if (text === '' || text === undefined || text === null) {
        return text
      }
      let _that = this
      if (keywords === undefined) {
        keywords = []
        let groupConfig = _that.getCurrentGroupConfig()
        if (!groupConfig || !groupConfig.extra_2) {
          return text
        }
        let extra2 = groupConfig.extra_2
        let regexErrors = extra2.split("\n")
        for (let i in regexErrors) {
          let regex = regexErrors[i]
          let regexParams = regex.split('#')
          if (regexParams.length === 2) {
            regex = regexParams[1]
          }
          keywords.push(regex)
        }
      }
      
      // Filter out non-string and empty keywords
      const validKeywords = keywords.filter(k => 
        typeof k === 'string' && k.trim() !== ''
      )
      
      if (validKeywords.length === 0) {
        return text
      }
      
      const regex = new RegExp(
          validKeywords.map(k => k.replace(/[.*+?^${}()|[\]\\]/g, '\\$&')).join('|'),
          'gi'
      );
      return text.replace(regex, match => {
        return `<span style="color:red">${match}</span>`;
      });
    },

    // 获取执行信息
    getExecutionInfo(tabId) {
      let _that = this
      return _that.getTabConfigById(tabId)
    },

    // 获取错误数量
    getErrorCount(tabId) {
      let _that = this
      if (parseInt(tabId) === 0) {
        return 0
      }
      return _that.errorMapList[tabId].length
    },
    // 获取过滤数量
    getFilterCount(tabId) {
      let _that = this
      if (parseInt(tabId) === 0) {
        return 0
      }
      let total = 0
      for (let i in _that.filterMapList[tabId]) {
        total += _that.filterMapList[tabId][i]
      }
      return total
    },
    // 显示错误弹窗
    showErrorDialog(tabId) {
      let _that = this
      _that.currentErrorTabId = tabId
      const tabConfig = _that.getTabConfigById(tabId)
      _that.currentErrorTabName = tabConfig ? tabConfig.name : '未知标签'
      _that.errorDialogVisible = true
    },
    // 显示过滤弹窗
    showFilterDialog(tabId) {
      let _that = this
      _that.currentErrorTabId = tabId
      _that.filterDialogVisible = true
    },
    // 清空错误
    clearErrors(tabId) {
      let _that = this
      const tabConfig = _that.getTabConfigById(tabId)
      shellOut.ShellOutCleanErrors({
        shell_client_id: tabConfig.shell_client_id,
      }, function () {
      })
      _that.errorMapList[tabId] = []
      _that.initFilterMap(tabId, tabConfig.group_id)
      _that.$forceUpdate() // 强制更新以刷新界面
      _that.updateErrorTotal()
    },
    initFilterMap: function (tabId, groupId) {
      let _that = this
      for (let i in _that.groupList) {
        if (parseInt(groupId) === parseInt(_that.groupList[i].id)) {
          let extra1 = _that.groupList[i].extra_1
          if(!Typ.IsString(extra1)){
            continue
          }
          let regexErrors = extra1.split("\n")
          for (let i in regexErrors) {
            let regex = regexErrors[i]
            _that.filterMapList[tabId] = _that.filterMapList[tabId] || {};
            _that.filterMapList[tabId][regex] = _that.filterMapList[tabId][regex] || 0;
          }
        }
      }
    },
    updateErrorTotal: function () {
      let _that = this
      let totalError = 0
      for (let i in _that.errorMapList) {
        totalError += _that.errorMapList[i].length
      }
    },
    // 窗口变化调整高度
    windowChange: function () {
      let _that = this
      window.addEventListener('resize', function () {
        shell.calculateShellDivHeight(_that)
      });
    },

    // 加载SSH列表
    loadSshList() {
      let _that = this
      ssh.SshList(res => {
        if (res.ErrCode === 0) {
          _that.sshList = res.Data
          _that.loadShellOuts()
        }
      })
    },
    // 从本地存储初始化标签页
    initTabsFromLocal(shellOuts) {
      let _that = this
      shellOuts.forEach(item => {
        let tabId = item.id
        if(_that.urlParams.id && parseInt(_that.urlParams.id) !== parseInt(tabId)){
          return
        }
        _that.createByTabId(tabId, item)
      })
      _that.activeTabId = _that.getStoreActiveTabId()
    },
    //创建sse管理器
    sseCreateHandle: function (sse_distribute_id, tabId, openFunc) {
      let _that = this
      sseDistribute.RegisterReceive(sse_distribute_id, function (msg, msgType, sseDistributeId) {
        if (msgType === 'msg') { //加入输出中
          if(!_that.isReceive){
            return
          }
          _that.throttleStringFunc[tabId].update(msg)
        } else if (msgType === 'error') { //新捕获到一个错误
          if (type.IsObject(msg)) {
            _that.errorMapList[tabId].unshift(msg)
          }
        } else if (msgType === 'filter') { //拦截
          _that.filterMapList[tabId] = _that.filterMapList[tabId] || {};
          _that.filterMapList[tabId][msg] = (_that.filterMapList[tabId][msg] || 0) + 1;
        } else if (msgType === 'error_list') { //重新链接时推送所有的错误列表
          if (type.IsArray(msg)) {
            _that.errorMapList[tabId] = msg
          }
        } else if (msgType === 'filter_list') { //重新链接时推送所有的错误列表
          if (type.IsObject(msg)) {
            _that.filterMapList[tabId] = msg
          }
        }
        // 限制长度：最多保留最后 50000 个字符
        const maxLen = 100000;
        if (_that.contentMapList[tabId].length > maxLen) {
          _that.contentMapList[tabId] = _that.contentMapList[tabId].slice(-maxLen);
        }
        let txt = format.formatResult(_that.contentMapList[tabId], ['copy', 'color', 'replace']);
        txt = format.formatResult(txt, ['length']);
        _that.contentMapList[tabId] = txt;
        _that.updateErrorTotal()
      })
      openFunc()
    },
    getSseDistributeIdByTabId: function (tabId) {
      let _that = this
      let tabConfig = _that.getTabConfigById(tabId)
      return tabConfig.sse_distribute_id
    },
    startByTabId: function (tabId) {
      let _that = this
      let item = _that.getTabConfigById(tabId)
      item.is_run = 1
      for (let i in _that.tabConfigList) {
        if (_that.tabConfigList[i].id === tabId) {
          _that.registerReceiveMsg(tabId)
          _that.tabConfigList[i] = item
        }
      }
      //创建sse
      const sse_distribute_id = sseDistribute.GetSseDistributeId(tabId)
      _that.createSseByTabConfig(sse_distribute_id, item)
    },
    createByTabId: function (tabId, item) {
      let _that = this
      const sse_distribute_id = sseDistribute.GetSseDistributeId(tabId)
      _that.scrollMap[tabId] = true
      _that.errorMapList[tabId] = []
      _that.initFilterMap(tabId, item.group_id)
      _that.contentMapList[tabId] = ''

      item.sse_distribute_id = sse_distribute_id
      _that.registerReceiveMsg(tabId)
      _that.tabConfigList.push(item)
      //如果是运行状态
      if (_that.urlParams.id) {
        if (parseInt(_that.urlParams.id) !== parseInt(item.id)) {
          item.is_run = 0
          return
        }
        item.is_run = 1
      }
      if (parseInt(item.is_run) === 0) {
        return
      }

      _that.createSseByTabConfig(sse_distribute_id, item)
    },
    registerReceiveMsg: function (tabId) {
      let _that = this
      _that.throttleStringFunc[tabId] = new Throttle_string(50, text => {
        _that.contentMapList[tabId] += text
      });
    },
    stopByTabId: function (tabId, back) {
      let _that = this
      shellOut.ShellOutStop(_that.getTabConfigById(tabId), function (res) {
        if (res.ErrCode !== 0) {
          _that.$helperNotify.error('停止失败')
        } else {
          for (let i in _that.tabConfigList) {
            if (_that.tabConfigList[i].id === tabId) {
              sse.SseClose(_that.tabConfigList[i].sse_distribute_id)
              _that.tabConfigList[i].is_run = 0
              _that.tabConfigList[i].shell_client_id = ''
              _that.contentMapList[tabId] = ''
              _that.initFilterMap(tabId, _that.tabConfigList[i].group_id)
              _that.errorMapList[tabId] = []
              _that.initFilterMap(tabId, _that.tabConfigList[i].group_id)
              _that.$forceUpdate()
            }
          }
        }
        if (back !== undefined && back !== null) {
          back()
        }
      })
    },
    restartByTabId: function (tabId, back) {
      let _that = this
      shellOut.ShellOutRestart(_that.getTabConfigById(tabId), function (res) {
        if (res.ErrCode !== 0) {
          _that.$helperNotify.error('停止失败')
        } else {
          for (let i in _that.tabConfigList) {
            if (_that.tabConfigList[i].id === tabId) {
              sse.SseClose(_that.tabConfigList[i].sse_distribute_id)
              _that.tabConfigList[i].is_run = 0
              _that.tabConfigList[i].shell_client_id = ''
              _that.contentMapList[tabId] = ''
              _that.initFilterMap(tabId, _that.tabConfigList[i].group_id)
              _that.errorMapList[tabId] = []
              _that.initFilterMap(tabId, _that.tabConfigList[i].group_id)
              _that.$forceUpdate()
            }
          }
        }
      })
    },
    up : function (){
      this.isReceive = !this.isReceive
    },
    cleanLog: function (tabId, back) {
      let _that = this
      shellOut.ShellOutCleanLog(_that.getTabConfigById(tabId), function (res) {
        if (res.ErrCode !== 0) {
          _that.$helperNotify.error('停止失败')
        } else {
          _that.contentMapList[tabId] = ''
        }
        if (back !== undefined && back !== null) {
          back()
        }
      })
    },
    getTabConfigById(tabId) {
      return this.tabConfigList.find(t => parseInt(t.id) === parseInt(tabId))
    },
    editTabConfig: function () {
      let _that = this
      let oldTabConfig = _that.getTabConfigById(_that.editTabConfigData.id)
      shell.ShellOutEdit(_that.editTabConfigData, function (res) {
        if (res.ErrCode !== 0) {
          _that.$helperNotify.error('编辑失败')
        } else {
          _that.$helperNotify.success('编辑成功')
          //重新启动命令
          if (_that.editTabConfigData.command !== oldTabConfig.command ||
              _that.editTabConfigData.ssh_id !== oldTabConfig.ssh_id) {
            _that.stopByTabId(oldTabConfig.id, function () {
              _that.startByTabId(oldTabConfig.id)
            })
          }
          for (let i in _that.tabConfigList) {
            if (parseInt(_that.tabConfigList[i].id) === parseInt(oldTabConfig.id)) {
              _that.tabConfigList[i].command = _that.editTabConfigData.command
              _that.tabConfigList[i].name = _that.editTabConfigData.name
              _that.tabConfigList[i].group_id = _that.editTabConfigData.group_id
              _that.tabConfigList[i].ssh_id = _that.editTabConfigData.ssh_id
            }
          }
          _that.cleanEditTabConfigData()
        }
      })
    },
    cleanEditTabConfigData: function () {
      let _that = this
      _that.editTabConfigData.command = ''
      _that.editTabConfigData.ssh_id = ''
      _that.editTabConfigData.name = ''
      _that.editTabConfigData.group_id = ''
      _that.editTabConfigData.id = ''
    },
    // 执行命令
    executeCommand() {
      let _that = this
      if (!_that.editTabConfigData.ssh_id || !_that.editTabConfigData.command || !_that.editTabConfigData.name) {
        this.$message.warning('请填写完整信息')
        return
      }
      if (parseInt(_that.editTabConfigData.id) > 0) {
        _that.editTabConfig()
        return
      }
      // 存储执行信息
      let tabConfig = {
        id: '',
        command: _that.editTabConfigData.command,
        sse_distribute_id: '',
        shell_client_id: '',
        ssh_id: _that.editTabConfigData.ssh_id,
        name: _that.editTabConfigData.name,
        is_run: _that.urlParams.id ? 0 : 1,
        group_id: _that.editTabConfigData.group_id,
      }
      // 调接口
      shell.ShellOutStart(tabConfig, (res) => {
        let tabId = res.Data.id
        let sse_distribute_id = sseDistribute.GetSseDistributeId(tabId)
        let shellClientId = res.Data.shell_client_id
        tabConfig.sse_distribute_id = sse_distribute_id
        _that.scrollMap[tabId] = true
        _that.errorMapList[tabId] = []
        _that.initFilterMap(tabId, tabConfig.group_id)
        _that.contentMapList[tabId] = ''
        _that.activeTabId = tabId
        tabConfig.shell_client_id = shellClientId
        tabConfig.id = tabId
        _that.tabConfigList.push(tabConfig)

        //创建sse
        _that.createSseByTabConfig(sse_distribute_id, tabConfig)
      })

      // 清空输入
      _that.cleanEditTabConfigData()
    },
    createSseByTabConfig: function (sse_distribute_id, tabConfig) {
      let _that = this
      _that.sseCreateHandle(sse_distribute_id, tabConfig.id, function () {
        shell.ShellOutSetSeeId({
          sse_distribute_id: sse_distribute_id,
          shell_client_id: tabConfig.shell_client_id,
          ssh_id: tabConfig.ssh_id,
          command: tabConfig.command,
          id: tabConfig.id,
          group_id: tabConfig.group_id,
          is_run: _that.urlParams.id ? 0 : 1,
        }, function (res) {
          if (res.ErrCode !== 0) {
            _that.$helperNotify.error('建立链接失败')
          }
        })
      })
      for (let i in _that.tabConfigList) {
        if (parseInt(_that.tabConfigList[i].id) === parseInt(tabConfig['id'])) {
          _that.tabConfigList[i].sse_distribute_id = sse_distribute_id
        }
      }
      shell.calculateShellDivHeight(_that)
    },
    restartTab(tabId){
      let _that = this
      shellOut.ShellOutStop(_that.getTabConfigById(tabId), function (res) {
        if (res.ErrCode !== 0) {
          _that.$helperNotify.error('停止失败')
        } else {
          window.location.reload()
        }
      })
    },
    // 移除标签页
    removeTab(tabId) {
      let _that = this
      shellOut.ShellOutDelete(_that.getTabConfigById(tabId), function (res) {
        if (res.ErrCode !== 0) {
          _that.$helperNotify.error('删除失败')
        } else {
          for (let i in _that.tabConfigList) {
            if (_that.tabConfigList[i].id !== tabId) {
              _that.activeTabId = _that.tabConfigList[i].id
              break
            }
          }
          delete _that.contentMapList[tabId]
          delete _that.scrollMap[tabId]
          for (let i in _that.tabConfigList) {
            if (_that.tabConfigList[i].id === tabId) {
              _that.tabConfigList.splice(i, 1)
              break
            }
          }
        }
      })

    },
  }
}
</script>

<style lang="scss" scoped>
.shell-console {
  padding: 12px;
  background: #fafafa;
}

.toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 12px;

  .select {
    width: 200px
  }

  .command {
    width: 300px;
    margin: 0 10px
  }

  .name {
    width: 200px;
    margin-right: 10px
  }
}

:deep(.el-tabs--card > .el-tabs__header .el-tabs__item.is-active) {
  background-color: #eeeeee !important; // 选中背景色
  border: 1px solid #409eff !important; // 边框颜色
  border-bottom-color: #409eff !important;
  font-weight: bold;
}

// 标签标题样式
.tab-label {
  display: flex;
  align-items: center;
  gap: 8px;
}

.tab-badge {
  :deep(.el-badge__content) {
    transform: scale(0.8);
  }
}

.error-line {
  white-space: pre-line; /* 把 \n 变成换行，不保留多余空格 */
}

// 执行信息区域样式
.execution-info {
  margin-bottom: 12px;
  padding: 12px;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);

  :deep(.el-descriptions) {
    .el-descriptions__label {
      font-weight: bold;
    }
  }
}

// 命令弹窗样式
.command-popover {
  h4 {
    margin: 0 0 8px 0;
    color: #333;
  }

  .full-command {
    background: #f5f5f5;
    padding: 8px;
    border-radius: 4px;
    font-family: 'Consolas', monospace;
    font-size: 12px;
    margin: 0 0 12px 0;
    max-height: 200px;
    overflow-y: auto;
  }

  .command-actions {
    display: flex;
    justify-content: flex-end;
  }
}

// 错误列表样式
.error-list {
  max-height: 60vh;
  overflow-y: auto;
  padding: 8px;
  background: #f8f9fa;
  border-radius: 8px;
  
  // Custom scrollbar
  &::-webkit-scrollbar {
    width: 8px;
  }
  
  &::-webkit-scrollbar-track {
    background: #f1f1f1;
    border-radius: 4px;
  }
  
  &::-webkit-scrollbar-thumb {
    background: #c1c1c1;
    border-radius: 4px;
    
    &:hover {
      background: #a8a8a8;
    }
  }
}

.search-item {
  margin-bottom: 12px;
  padding: 12px 16px;
  border-radius: 6px;
  background: #fff;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  border-left: 3px solid transparent;
  
  &:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
    transform: translateX(2px);
  }
  
  &:last-child {
    margin-bottom: 0;
  }
  
  // Unread state - highlight with red border
  &:not([style*="border: 0px"]) {
    border-left-color: #f56c6c;
    background: linear-gradient(to right, #fff5f5, #fff);
  }
  
  // Line number styling
  &::before {
    content: "##" attr(data-line) "##";
    display: inline-block;
    background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
    color: #fff;
    padding: 2px 8px;
    border-radius: 4px;
    font-size: 11px;
    font-weight: 600;
    margin-right: 8px;
    font-family: 'Consolas', 'Courier New', monospace;
    box-shadow: 0 2px 4px rgba(102, 126, 234, 0.3);
    vertical-align: middle;
  }
}

.search-content {
  display: inline;
  font-size: 13px;
  color: #2c3e50;
  line-height: 1.6;
  word-break: break-all;
  
  // Highlight text styling
  :deep(span[style*="color:red"]) {
    background: linear-gradient(135deg, #ff6b6b 0%, #ee5a5a 100%);
    color: #fff !important;
    padding: 2px 6px;
    border-radius: 3px;
    font-weight: 600;
    box-shadow: 0 1px 3px rgba(255, 107, 107, 0.4);
  }
}

.search-actions {
  margin-top: 8px;
  padding-top: 8px;
  border-top: 1px dashed #e0e0e0;
  display: flex;
  justify-content: flex-end;
}

.context-btn {
  font-weight: 500;
  transition: all 0.2s ease;
  
  .btn-icon {
    margin-right: 4px;
    font-size: 14px;
  }
  
  &:hover {
    transform: translateY(-1px);
  }
}

.error-item {
  margin-bottom: 6px;
  padding: 6px;
  border: 1px solid #f56c6c;
  border-radius: 1px;
  background: #eeeeee;

  &:last-child {
    margin-bottom: 0;
  }
}

// Card style for error items
.card-item {
  margin-bottom: 12px;
  padding: 12px 16px;
  border-radius: 6px;
  background: #fff;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.08);
  transition: all 0.3s ease;
  border-left: 3px solid #f56c6c;

  &:hover {
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.12);
    transform: translateX(2px);
  }

  &:last-child {
    margin-bottom: 0;
  }

  // Line number styling using pseudo element
  &::before {
    content: "##" attr(data-line) "##";
    display: none;
  }
}

.error-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px dashed #e0e0e0;
}

.error-time {
  font-size: 12px;
  color: #909399;
  background: #f5f7fa;
  padding: 4px 10px;
  border-radius: 4px;
  font-weight: 500;
}

.error-content {
  margin-bottom: 8px;
  font-size: 13px;
  color: #2c3e50;
  line-height: 1.6;
  word-break: break-all;
  white-space: pre-wrap;
  font-family: 'Consolas', 'Courier New', monospace;

  // Highlight error styling
  :deep(span[style*="color:red"]) {
    background: linear-gradient(135deg, #ff6b6b 0%, #ee5a5a 100%);
    color: #fff !important;
    padding: 2px 6px;
    border-radius: 3px;
    font-weight: 600;
    box-shadow: 0 1px 3px rgba(255, 107, 107, 0.4);
  }
}

.error-actions {
  display: flex;
  justify-content: flex-end;
  padding-top: 8px;
  border-top: 1px dashed #e0e0e0;
}

.error-context-info {
  font-size: 12px;
  color: #67c23a;
  background: #f0f9eb;
  padding: 2px 6px;
  border-radius: 3px;
  flex-basis: 100%;
}

// Context list styles
.context-list {
  background: #f5f7fa;
  border-radius: 8px;
  padding: 12px;
}

.context-item {
  display: flex;
  align-items: flex-start;
  padding: 8px 12px;
  background: #fff;
  border-radius: 4px;
  margin-bottom: 8px;
  transition: all 0.2s ease;

  &:last-child {
    margin-bottom: 0;
  }

  &:hover {
    box-shadow: 0 2px 6px rgba(0, 0, 0, 0.08);
  }

  &.highlight-line {
    background: linear-gradient(135deg, #fff5f5 0%, #ffe8e8 100%);
    border-left: 4px solid #f56c6c;
    box-shadow: 0 2px 8px rgba(245, 108, 108, 0.3);
    animation: highlight-pulse 2s ease-in-out;
  }
}

.line-number {
  flex-shrink: 0;
  display: inline-block;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  padding: 2px 8px;
  border-radius: 4px;
  font-size: 11px;
  font-weight: 600;
  margin-right: 12px;
  font-family: 'Consolas', 'Courier New', monospace;
  box-shadow: 0 2px 4px rgba(102, 126, 234, 0.3);
  min-width: 80px;
  text-align: center;
}

.line-content {
  flex: 1;
  font-size: 13px;
  color: #2c3e50;
  line-height: 1.6;
  word-break: break-all;
  white-space: pre-wrap;
  font-family: 'Consolas', 'Courier New', monospace;
}

@keyframes highlight-pulse {
  0%, 100% {
    transform: translateX(0);
  }
  50% {
    transform: translateX(4px);
  }
}


.error-time {
  font-size: 12px;
  color: #909399;
}

.error-line {
  font-size: 12px;
  color: #606266;
  background: #e6e6e6;
  padding: 2px 6px;
  border-radius: 3px;
}

.error-content {
  white-space: pre-line; /* 把 \n 变成换行，不保留多余空格 */
  background: #2d2d2d;
  color: #e0e0e0;
  padding: 12px;
  border-radius: 4px;
  font-family: 'Consolas', 'Courier New', monospace;
  font-size: 12px;
  line-height: 1.4;
  word-break: break-all;
  margin: 0;

  // 高亮错误关键词的样式

  :deep(.error-highlight) {
    color: #ff6b6b;
    font-weight: bold;
    background: rgba(255, 107, 107, 0.1);
    padding: 2px 4px;
    border-radius: 3px;

    // 严重错误

    &.error-critical {
      color: #ff4757;
      background: rgba(255, 71, 87, 0.1);
    }

    // 警告

    &.error-warning {
      color: #ffa502;
      background: rgba(255, 165, 2, 0.1);
    }

    // 数据库错误 - 使用紫色

    &.error-database {
      color: #a29bfe;
      background: rgba(162, 155, 254, 0.1);
      border-left: 3px solid #a29bfe;
    }

    // 语法错误 - 使用橙色

    &.error-syntax {
      color: #fd9644;
      background: rgba(253, 150, 68, 0.1);
      border-left: 3px solid #fd9644;
    }
  }

  :deep(.error-line-marker) {
    background: rgba(255, 107, 107, 0.2) !important;
    border: 2px solid #ff6b6b;
    padding: 4px 8px;
    display: block;
    margin: 8px 0;
    border-radius: 6px;
    font-weight: bold;
    color: #ff6b6b;
  }
}

.no-errors {
  text-align: center;
  color: #909399;
  padding: 60px 20px;
  font-size: 14px;
  
  .no-data-icon {
    font-size: 48px;
    margin-bottom: 12px;
    opacity: 0.6;
  }
}

pre {
  margin: 0;
  padding: 10px 0 20px 0;
  white-space: pre-wrap;
  word-break: break-all;
  line-height: 1.4;
}

@keyframes gentle-blink {
  0%, 100% {
    opacity: 0.7;
  }
  50% {
    opacity: 0.3;
  }
}

.running-tab {
  color: #52c41a !important;
  font-weight: bold !important;
}

// Search dialog specific styles
:deep(.el-dialog__body) {
  padding: 20px !important;
  background: #f5f7fa;
}

// Search statistics badge
.search-stats {
  display: inline-flex;
  align-items: center;
  padding: 4px 12px;
  background: linear-gradient(135deg, #667eea 0%, #764ba2 100%);
  color: #fff;
  border-radius: 12px;
  font-size: 12px;
  font-weight: 600;
  margin-left: 8px;
  box-shadow: 0 2px 6px rgba(102, 126, 234, 0.4);
}
</style>