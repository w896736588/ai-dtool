<template>
  <div class="shell-console">
    <div id="mainCard" ref="mainCard" v-if="parseInt(urlParams.id) !== 0 && getExecutionInfo(urlParams.id)" class="execution-info" style="display: flex; align-items: center; gap: 8px;">
      <h4>{{urlParams.title}}</h4>
      <el-tag size="small" effect="light" type="success">
        规则集：{{ getRuleSetName(urlParams.id) }}
      </el-tag>
      <el-popover
          placement="top-start"
          trigger="click"
          width="300px"
      >
        <!-- 触发按钮 -->
        <template #reference>
          <pl-button plain size="small" type="primary">查看命令</pl-button>
        </template>

        <!-- 气泡内容 -->
        <div class="command-popover">
          <h4>完整命令</h4>
          <pre class="full-command">{{ getExecutionInfo(urlParams.id).command }}</pre>
          <div class="command-actions">
            <pl-button
                size="small"
                type="primary"
                @click="copyCommand(getExecutionInfo(urlParams.id).command)"
            >
              复制命令
            </pl-button>
          </div>
        </div>
      </el-popover>
      <pl-button
          size="small"
          type="danger"
          @click="showAlertRulesDialog(urlParams.id)"
      >
        {{ getErrorCount(urlParams.id) }} 个告警
      </pl-button>
      <pl-button
          size="small"
          type="info"
          @click="showFilterRulesDialog(urlParams.id)"
      >
        {{ getFilterCount(urlParams.id) }} 次过滤
      </pl-button>
      <pl-button
          :disabled="getErrorCount(urlParams.id) === 0"
          size="small"
          @click="clearErrors(urlParams.id)"
      >
        清空错误
      </pl-button>
<!--      <pl-button-->
<!--          size="small"-->
<!--          @click="removeTab(urlParams.id)"-->
<!--      >-->
<!--        删除-->
<!--      </pl-button>-->
      <pl-button
          size="small"
          type="primary"
          @click="restartTab(urlParams.id)"
      >
        重启
      </pl-button>
<!--      <pl-button-->
<!--          size="small" @click="startByTabId(urlParams.id)"-->
<!--      >-->
<!--        启动-->
<!--      </pl-button>-->
<!--      <pl-button-->
<!---->
<!--          size="small" @click="stopByTabId(urlParams.id)"-->
<!--      >-->
<!--        停止-->
<!--      </pl-button>-->
<!--      <pl-button-->
<!--          size="small" @click="restartByTabId(urlParams.id)"-->
<!--      >-->
<!--        重启-->
<!--      </pl-button>-->
      <pl-button
          size="small" @click="cleanLog(urlParams.id)"
      >
        清除日志
      </pl-button>
      <pl-button
          size="small" @click="up()"
      >
        {{isReceive ? '暂停接收' : '开始接收'}}
      </pl-button>
      <!--            <el-tag style="margin: 5px;">链接：{{ getSshName(tab.ssh_id) }}</el-tag>-->
      <el-tag style="margin: 5px;">内容长度：{{ getContentLength(activeTabId) }}</el-tag>
      <el-tag v-if="selectingTabIds[activeTabId] && rawBufferMap[activeTabId]" type="warning" size="small" effect="dark" style="margin: 5px;">
        缓冲中({{ rawBufferMap[activeTabId].length }}字符)
      </el-tag>
      <el-autocomplete
          v-model="searchContent"
          placeholder="输入搜索，多个之间用##分隔"
          size="small"
          style="width: auto; min-width: 200px;"
          :fetch-suggestions="fetchSearchSuggestions"
          @select="onSearchSelect"
          @keyup.enter="searchByContent"
      >
        <template #default="{ item }">
          <span>{{ item.value }}</span>
        </template>
      </el-autocomplete>
      <pl-button size="small" @click="searchByContent">搜索</pl-button>
      <el-input
          v-model="filterContent"
          placeholder="SSE过滤，多个关键词用空格分隔（全部包含才推送）"
          size="small"
          style="width: auto; min-width: 260px;"
          @keyup.enter="applyFilter"
      />
      <pl-button size="small" type="warning" :class="{'filter-active-btn': isFilterActive}" @click="applyFilter">{{ isFilterActive ? '过滤中' : '开始过滤' }}</pl-button>
    </div>
    <!-- 输出区 -->
    <shellResult ref="shellRef" :divHeight="shellController.divHeight" :isRunning="shellController.isRunning"
                 :shellShowResult="contentMapList[activeTabId]" :show-model="shellController.showModel"
                 @selection-change="onShellSelectionChange"></shellResult>

    <!-- Error list dialog -->
    <el-dialog
        v-model="errorDialogVisible"
        :title="`告警列表 - ${currentErrorTabName}`"
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
            <div class="error-tag-group">
              <el-tag :type="getErrorTagType(error.level)" size="small" effect="plain">{{ error.level || 'warning' }}</el-tag>
              <el-tag v-if="error.rule_name" type="info" size="small" effect="plain">{{ error.rule_name }}</el-tag>
              <el-tag v-if="error.category" size="small" effect="plain">{{ error.category }}</el-tag>
            </div>
          </div>
          <div class="error-content">
            <span style="line-height:1.6" v-html="highlightErrors(error.error_line)"></span>
          </div>
          <div class="error-actions">
            <pl-button
                type="primary"
                link
                size="small"
                @click="getErrorContent(error.line_number)"
                class="context-btn"
            >
              <span class="btn-icon">📋</span>
              查看上下文
            </pl-button>
          </div>
        </div>
        <div v-if="activeTabId > 0 && errorMapList[activeTabId].length === 0" class="no-errors">
          <div class="no-data-icon">✅</div>
          <div>暂无告警信息</div>
        </div>
      </div>
      <template #footer>
        <pl-button @click="errorDialogVisible = false">关闭</pl-button>
        <pl-button
            :disabled="errorMapList[activeTabId].length === 0"
            type="danger"
            @click="clearErrors(activeTabId)"
        >
          清空告警
        </pl-button>
      </template>
    </el-dialog>

    <!-- 告警规则列表弹窗 -->
    <el-dialog
        v-model="alertRulesDialogVisible"
        :title="`告警规则触发列表 - ${currentErrorTabName}`"
        width="80%"
    >
      <el-table :data="getAlertRulesList()" style="width: 100%" stripe>
        <el-table-column label="触发次数" prop="triggerCount" width="100" align="center" sortable :default-sort="{prop: 'triggerCount', order: 'descending'}">
          <template #default="scope">
            <el-tag type="danger" size="small">{{ scope.row.triggerCount }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="规则名称" prop="name" min-width="180" show-overflow-tooltip />
        <el-table-column label="匹配方式" width="100" align="center">
          <template #default="scope">
            <span class="match-type-text">{{ scope.row.matchType }}</span>
          </template>
        </el-table-column>
        <el-table-column label="匹配内容" prop="pattern" min-width="200" show-overflow-tooltip />
        <el-table-column label="告警级别" width="100" align="center">
          <template #default="scope">
            <el-tag :type="getAlertLevelType(scope.row.level)" size="small">{{ scope.row.level || 'warning' }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="分类" prop="category" width="120" show-overflow-tooltip />
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="scope">
            <pl-button type="primary" link size="small" @click="viewAlertRuleDetails(scope.row)">查看详情</pl-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- 告警规则详情弹窗 -->
    <el-dialog
        v-model="alertRuleDetailDialogVisible"
        :title="`告警详情 - ${currentAlertRuleName} - 共${currentAlertRuleErrors.length}条`"
        width="80%"
    >
      <div class="error-list">
        <div
            v-for="(error) in currentAlertRuleErrors"
            :key="error.line_number"
            class="error-item card-item"
            :data-line="error.line_number"
        >
          <div class="error-header">
            <span class="error-time">{{ error.time }}</span>
            <div class="error-tag-group">
              <el-tag :type="getErrorTagType(error.level)" size="small" effect="plain">{{ error.level || 'warning' }}</el-tag>
              <el-tag v-if="error.category" size="small" effect="plain">{{ error.category }}</el-tag>
            </div>
          </div>
          <div class="error-content">
            <span style="line-height:1.6" v-html="highlightErrors(error.error_line)"></span>
          </div>
          <div class="error-actions">
            <pl-button
                type="primary"
                link
                size="small"
                @click="getErrorContent(error.line_number)"
                class="context-btn"
            >
              <span class="btn-icon">📋</span>
              查看上下文
            </pl-button>
          </div>
        </div>
        <div v-if="currentAlertRuleErrors.length === 0" class="no-errors">
          <div class="no-data-icon">✅</div>
          <div>该规则暂无触发记录</div>
        </div>
      </div>
      <template #footer>
        <pl-button @click="alertRuleDetailDialogVisible = false">关闭</pl-button>
      </template>
    </el-dialog>

    <!-- 过滤规则列表弹窗 -->
    <el-dialog
        v-model="filterRulesDialogVisible"
        :title="`过滤规则触发列表 - ${currentErrorTabName}`"
        width="80%"
    >
      <el-table :data="getFilterRulesList()" style="width: 100%" stripe>
        <el-table-column label="触发次数" prop="triggerCount" width="100" align="center" sortable :default-sort="{prop: 'triggerCount', order: 'descending'}">
          <template #default="scope">
            <el-tag type="info" size="small">{{ scope.row.triggerCount }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column label="规则名称" prop="name" min-width="180" show-overflow-tooltip />
        <el-table-column label="匹配方式" width="100" align="center">
          <template #default="scope">
            <span class="match-type-text">{{ scope.row.matchType }}</span>
          </template>
        </el-table-column>
        <el-table-column label="匹配内容" prop="pattern" min-width="200" show-overflow-tooltip />
        <el-table-column label="操作" width="120" fixed="right" align="center">
          <template #default="scope">
            <pl-button type="primary" link size="small" @click="viewFilterRuleDetails(scope.row)">查看详情</pl-button>
          </template>
        </el-table-column>
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
        <pl-button @click="errorDialogContextVisible = false">关闭</pl-button>
      </template>
    </el-dialog>

    <!-- 搜索信息弹窗 -->
    <el-dialog
        v-model="searchDialogVisible"
        :title="`搜索列表 - ${searchContent} - 搜索到${searchNumber}条`"
        width="80%"
        class="search-dialog"
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
            <pl-button 
                type="primary" 
                link 
                size="small"
                @click="getErrorContent(search.LineNumber)"
                class="context-btn"
            >
              <span class="btn-icon">📋</span>
              查看上下文
            </pl-button>
          </div>
        </div>
        <div v-if="!searchContents || searchContents.length === 0" class="no-errors">
          <div class="no-data-icon">🔍</div>
          <div>未搜索到信息</div>
        </div>
      </div>
      <template #footer>
        <pl-button @click="searchDialogVisible = false">关闭</pl-button>
      </template>
    </el-dialog>

    <el-dialog
        v-model="groupDialog"
        title="分组"
        width="80%"
    >
      <Group :groupTitle="'终端输出'" :groupType="groupType" @update="groupUpdate"></Group>
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
import shellOutRule from "@/utils/base/shell_out_rule"
import format from "@/utils/base/format";
import shellResult from "@/components/shell/result_div.vue";
import type from "@/utils/base/type"
import Group from "@/components/group/group_list.vue"
import group from "@/utils/base/group"
import store from "@/utils/base/store"
import sseDistribute from "@/utils/base/sse_distribute";
import {Throttle_string} from "@/utils/base/throttle_string"
import {useRoute} from 'vue-router';


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
      ruleSetList: [],
      ruleSetInfoMap: {},
      chooseGroupId: '',
      //编辑 能够编辑的项
      editTabConfigData: {
        id: 0,
        ssh_id: '',
        group_id: '',
        rule_set_id: '',
        command: '',
        name: '',
      },
      activeTabId: '',
      scrollMap: {},
      shellInstances: new Map(),

      // 错误弹窗相关
      errorDialogVisible: false,
      filterDialogVisible: false,
      alertRulesDialogVisible: false,
      filterRulesDialogVisible: false,
      alertRuleDetailDialogVisible: false,
      currentAlertRuleName: '',
      currentAlertRuleErrors: [],
      currentErrorTabId: '',
      currentErrorTabName: '',

      errorMapList: [], //错误捕获
      filterMapList: [], //过滤
      contentMapList: [], //输出内容
      throttleStringFunc: [],//节流回调
      tabConfigList: [],//配置
      // 用户拖选文字时暂存增量日志，避免 v-html 更新打断选区
      rawBufferMap: [],
      selectingTabIds: {},
      groupList: [], //分组列表
      groupType: `6`,
      errorDialogContextVisible: false,
      searchDialogVisible: false,
      errorContext: '',
      errorLine: '',
      searchContent: '',
      searchContents: [],
      searchNumber: 0,
      searchHistory: [], // 搜索历史记录
      filterContent: '', // SSE推送过滤关键词
      isFilterActive: false, // 当前是否处于过滤状态
      urlParams: {},
      isReceive : true, //是否接受
      // Fullpage 独立 SSE 连接（不再使用通用的 /sse 混用）
      fullpageSseConn: null,
      fullpageClientId: '',
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
    _that.loadRuleSetList()
    _that.getFullPageParams()
    _that.loadSearchHistory()
    //如果是单独展示的页面 里面返回的就是传参的
    _that.chooseGroupId = _that.getStoreGroupId()
    _that.activeTabId = _that.getStoreActiveTabId()
  },
  activated: function () {
    let _that = this
    _that.windowChange()
  },
  deactivated() {

  },
  beforeUnmount() {
    let _that = this
    // 关闭 Fullpage 独立 SSE 连接
    _that.closeFullpageSse()
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
      const dropRuleMap = _that.getRuleItemMapByType(_that.activeTabId, 'drop')
      for (let i in _that.filterMapList[_that.activeTabId]) {
        let number = _that.filterMapList[_that.activeTabId][i]
        const ruleItem = dropRuleMap[i] || {}
        filters.push({
          name: ruleItem.name || i,
          number: number,
          pattern: ruleItem.pattern || i,
        })
      }
      filters.sort((a, b) => b.number - a.number)
      return filters
    },
    // loadRuleSetList 预加载规则集列表，终端详情页据此解析当前绑定关系。 // Preload rule-set metadata so the shell detail page can resolve the bound rule set quickly.
    loadRuleSetList() {
      let _that = this
      shellOutRule.ShellOutRuleSetList({}, function (response) {
        if (response.ErrCode !== 0) {
          return
        }
        _that.ruleSetList = Array.isArray(response.Data) ? response.Data : []
      })
    },
    // ensureRuleSetInfo 按需拉取规则集详情，避免页面初始化时把所有规则项全量加载。 // Load rule-set details on demand instead of fetching every nested rule item up front.
    ensureRuleSetInfo(ruleSetId) {
      let _that = this
      const currentRuleSetId = parseInt(ruleSetId)
      if (currentRuleSetId <= 0 || _that.ruleSetInfoMap[currentRuleSetId]) {
        return
      }
      shellOutRule.ShellOutRuleSetInfo({id: currentRuleSetId}, function (response) {
        if (response.ErrCode !== 0) {
          return
        }
        const items = Array.isArray(response.Data?.rule_items) ? response.Data.rule_items : []
        _that.ruleSetInfoMap[currentRuleSetId] = {
          rule_set: response.Data?.rule_set || {},
          rule_items: items,
        }
      })
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
      _that.saveSearchHistory(_that.searchContent)
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
    // 加载搜索历史记录
    loadSearchHistory: function () {
      let raw = store.getStore('shell_out_search_history')
      if (raw) {
        try {
          this.searchHistory = JSON.parse(raw)
        } catch (e) {
          this.searchHistory = []
        }
      }
    },
    // 保存搜索记录到历史（去重，最多保留10条）
    saveSearchHistory: function (content) {
      content = (content || '').trim()
      if (!content) return
      // 去重：先移除已存在的相同记录
      this.searchHistory = this.searchHistory.filter(function (item) {
        return item !== content
      })
      // 添加到最前面
      this.searchHistory.unshift(content)
      // 最多保留10条
      if (this.searchHistory.length > 10) {
        this.searchHistory = this.searchHistory.slice(0, 10)
      }
      store.setStore('shell_out_search_history', JSON.stringify(this.searchHistory))
    },
    // el-autocomplete 获取建议列表
    fetchSearchSuggestions: function (queryString, callback) {
      let history = this.searchHistory
      if (!queryString) {
        callback(history.map(function (item) {
          return { value: item }
        }))
      } else {
        callback(history.filter(function (item) {
          return item.toLowerCase().indexOf(queryString.toLowerCase()) >= 0
        }).map(function (item) {
          return { value: item }
        }))
      }
    },
    // 选中历史记录时触发搜索
    onSearchSelect: function (item) {
      this.searchContent = item.value
      this.searchByContent()
    },
    // 应用 SSE 推送过滤：将关键词传入后端后刷新页面，使历史日志也逐行经过过滤回放
    applyFilter: function () {
      let _that = this
      let tabConfig = _that.getTabConfigById(_that.activeTabId)
      if (!tabConfig || !tabConfig['shell_client_id']) {
        _that.$helperNotify.warning('当前没有活跃的终端会话')
        return
      }
      shellOut.ShellOutSetFilter({
        'shell_client_id': tabConfig['shell_client_id'],
        'filter_content': _that.filterContent,
      }, function (res) {
        if (res.ErrCode === 0) {
          window.location.reload()
        } else {
          _that.$helperNotify.error('设置过滤失败')
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
    getRuleSetName(tabId) {
      const tabConfig = this.getTabConfigById(tabId)
      if (!tabConfig || parseInt(tabConfig.rule_set_id) <= 0) {
        return '未启用规则'
      }
      const ruleSet = this.ruleSetList.find(item => parseInt(item.id) === parseInt(tabConfig.rule_set_id))
      return ruleSet ? ruleSet.name : `规则集#${tabConfig.rule_set_id}`
    },
    getCurrentRuleSet(tabId) {
      const tabConfig = this.getTabConfigById(tabId)
      if (!tabConfig || parseInt(tabConfig.rule_set_id) <= 0) {
        return null
      }
      this.ensureRuleSetInfo(tabConfig.rule_set_id)
      return this.ruleSetInfoMap[parseInt(tabConfig.rule_set_id)] || null
    },
    getRuleItemsByType(tabId, ruleType) {
      const currentRuleSet = this.getCurrentRuleSet(tabId)
      if (!currentRuleSet || !Array.isArray(currentRuleSet.rule_items)) {
        return []
      }
      return currentRuleSet.rule_items.filter(item => item.rule_type === ruleType && Number(item.is_enabled) === 1)
    },
    getRuleItemMapByType(tabId, ruleType) {
      const result = {}
      this.getRuleItemsByType(tabId, ruleType).forEach((item) => {
        if (!item || !item.name) {
          return
        }
        result[item.name] = item
      })
      return result
    },
    getHighlightKeywords(tabId) {
      const alertRules = this.getRuleItemsByType(tabId, 'alert')
      return alertRules
          .map(item => item.pattern)
          .filter(item => typeof item === 'string' && item.trim() !== '')
    },
    getTopFilterRule(tabId) {
      const filterList = this.getFilterList()
      if (parseInt(this.activeTabId) !== parseInt(tabId)) {
        const filterMap = this.filterMapList[tabId] || {}
        const dropRuleMap = this.getRuleItemMapByType(tabId, 'drop')
        const tabFilters = Object.keys(filterMap).map((key) => ({
          name: dropRuleMap[key]?.name || key,
          number: filterMap[key],
        })).sort((a, b) => b.number - a.number)
        if (tabFilters.length === 0) {
          return '暂无过滤命中'
        }
        return `${tabFilters[0].name} ${tabFilters[0].number} 次`
      }
      if (filterList.length === 0) {
        return '暂无过滤命中'
      }
      return `${filterList[0].name} ${filterList[0].number} 次`
    },
    getTopAlertRule(tabId) {
      const alerts = Array.isArray(this.errorMapList[tabId]) ? this.errorMapList[tabId] : []
      if (alerts.length === 0) {
        return '暂无告警命中'
      }
      const counter = {}
      alerts.forEach((item) => {
        const key = item.rule_name || '未命名规则'
        counter[key] = (counter[key] || 0) + 1
      })
      const topRuleName = Object.keys(counter).sort((a, b) => counter[b] - counter[a])[0]
      return `${topRuleName} ${counter[topRuleName]} 条`
    },
    getErrorTagType(level) {
      if (level === 'error') {
        return 'danger'
      }
      if (level === 'warning') {
        return 'warning'
      }
      return 'info'
    },
    // Highlight error keywords
    highlightErrors(text, keywords) {
      if (text === '' || text === undefined || text === null) {
        return text
      }
      let _that = this
      if (keywords === undefined) {
        keywords = _that.getHighlightKeywords(_that.activeTabId)
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
      return Array.isArray(_that.errorMapList[tabId]) ? _that.errorMapList[tabId].length : 0
    },
    // 获取过滤数量
    getFilterCount(tabId) {
      let _that = this
      if (parseInt(tabId) === 0) {
        return 0
      }
      let total = 0
      const filterMap = _that.filterMapList[tabId] || {}
      for (let i in filterMap) {
        total += filterMap[i]
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
    // 显示告警规则列表弹窗
    showAlertRulesDialog(tabId) {
      let _that = this
      _that.currentErrorTabId = tabId
      const tabConfig = _that.getTabConfigById(tabId)
      _that.currentErrorTabName = tabConfig ? tabConfig.name : '未知标签'
      _that.alertRulesDialogVisible = true
    },
    // 显示过滤规则列表弹窗
    showFilterRulesDialog(tabId) {
      let _that = this
      _that.currentErrorTabId = tabId
      const tabConfig = _that.getTabConfigById(tabId)
      _that.currentErrorTabName = tabConfig ? tabConfig.name : '未知标签'
      _that.filterRulesDialogVisible = true
    },
    // 获取告警规则列表（带触发次数）
    getAlertRulesList() {
      let _that = this
      const tabId = _that.currentErrorTabId
      const alerts = Array.isArray(_that.errorMapList[tabId]) ? _that.errorMapList[tabId] : []
      const alertRules = _that.getRuleItemsByType(tabId, 'alert')
      
      // 统计每个规则的触发次数
      const triggerCountMap = {}
      alerts.forEach((item) => {
        const ruleName = item.rule_name || '未命名规则'
        triggerCountMap[ruleName] = (triggerCountMap[ruleName] || 0) + 1
      })
      
      // 构建规则列表，包含触发次数
      const rulesList = alertRules.map((rule) => {
        const ruleName = rule.name || '未命名规则'
        return {
          id: rule.id,
          name: ruleName,
          pattern: rule.pattern || '',
          matchType: rule.match_type === 'regex' ? '正则匹配' : '包含文字',
          level: rule.config_json ? JSON.parse(rule.config_json).level : 'warning',
          category: rule.config_json ? JSON.parse(rule.config_json).category : '',
          triggerCount: triggerCountMap[ruleName] || 0,
          isEnabled: Number(rule.is_enabled) === 1
        }
      })
      
      // 按触发次数降序排列
      return rulesList.sort((a, b) => b.triggerCount - a.triggerCount)
    },
    // 获取过滤规则列表（带触发次数）
    getFilterRulesList() {
      let _that = this
      const tabId = _that.currentErrorTabId
      const filterMap = _that.filterMapList[tabId] || {}
      const dropRules = _that.getRuleItemsByType(tabId, 'drop')
      const dropRuleMap = _that.getRuleItemMapByType(tabId, 'drop')
      
      // 构建规则列表，包含触发次数
      const rulesList = dropRules.map((rule) => {
        const ruleName = rule.name || '未命名规则'
        return {
          id: rule.id,
          name: ruleName,
          pattern: rule.pattern || '',
          matchType: rule.match_type === 'regex' ? '正则匹配' : '包含文字',
          triggerCount: filterMap[ruleName] || 0,
          isEnabled: Number(rule.is_enabled) === 1
        }
      })
      
      // 按触发次数降序排列
      return rulesList.sort((a, b) => b.triggerCount - a.triggerCount)
    },
    // 获取告警级别对应的标签类型
    getAlertLevelType(level) {
      if (level === 'error') return 'danger'
      if (level === 'warning') return 'warning'
      return 'info'
    },
    // 查看告警规则详情
    viewAlertRuleDetails(rule) {
      const tabId = this.currentErrorTabId
      const alerts = Array.isArray(this.errorMapList[tabId]) ? this.errorMapList[tabId] : []
      const ruleErrors = alerts.filter(item => (item.rule_name || '未命名规则') === rule.name)
      this.currentAlertRuleName = rule.name
      this.currentAlertRuleErrors = ruleErrors
      this.alertRuleDetailDialogVisible = true
    },
    // 查看过滤规则详情
    viewFilterRuleDetails(rule) {
      this.$alert(
        `<div style="max-height: 400px; overflow-y: auto;">
          <p><strong>规则名称：</strong>${rule.name}</p>
          <p><strong>触发次数：</strong>${rule.triggerCount}</p>
          <p><strong>匹配方式：</strong>${rule.matchType}</p>
          <p><strong>匹配内容：</strong><code style="background: #f5f7fa; padding: 2px 6px; border-radius: 4px;">${rule.pattern}</code></p>
        </div>`,
        '规则详情',
        {
          dangerouslyUseHTMLString: true,
          confirmButtonText: '关闭'
        }
      )
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
      _that.$forceUpdate() // 强制更新以刷新界面
      _that.updateErrorTotal()
    },
    // resetRuleRuntimeState 每次创建、停止或重启时重置前端运行态，避免残留上一次统计。 // Reset per-tab runtime state on create/stop/restart so rule counters never leak across sessions.
    resetRuleRuntimeState(tabId) {
      this.filterMapList[tabId] = {}
      this.errorMapList[tabId] = []
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
    // 创建独立的 Fullpage SSE 连接（专用于 /sse/fullpage，不与通用 SSE 混用）
    createFullpageSseConnection: function (sse_distribute_id, tabId, openFunc) {
      let _that = this
      _that.closeFullpageSse()

      // 先通过 API 获取可用 SSE 端口（通用 SSE 被排除后，runtimeSsePort 未设置）
      sseDistribute.fetchAvailableSsePort().then(function(availablePort) {
        if (!availablePort) {
          console.error('[Fullpage-SSE] 无可用SSE端口')
          return
        }
        base.SetSsePort(availablePort)
        let sseHost = base.GetSseApiHost(availablePort)

        _that.fullpageClientId = 'fullpage_' + base.GenerateId('sse_client_id')
        let url = sseHost + '/sse/fullpage?client_id=' + encodeURIComponent(_that.fullpageClientId) + '&token=' + encodeURIComponent(base.GetSafeToken())
        console.log('[Fullpage-SSE] 连接URL:', url)

        let es = new EventSource(url)
        _that.fullpageSseConn = es

        es.onopen = function () {
          console.log('[Fullpage-SSE] 连接已打开 client_id=' + _that.fullpageClientId)
          openFunc()
        }

      es.onmessage = function (event) {
        let objData = null
        try {
          objData = JSON.parse(event.data)
        } catch (e) {
          // 非JSON消息（如 [CONNECT]），忽略
          return
        }
        if (!objData || !objData.sse_distribute_id) {
          return
        }
        // 只处理当前 tab 的消息
        let msg = objData.data
        let msgType = objData.type

        if (msgType === 'msg') {
          if (!_that.isReceive) return
          if (_that.throttleStringFunc[tabId]) {
            _that.throttleStringFunc[tabId].update(msg)
          }
        } else if (msgType === 'error') {
          if (type.IsObject(msg)) {
            _that.errorMapList[tabId] = _that.errorMapList[tabId] || []
            _that.errorMapList[tabId].unshift(msg)
          }
        } else if (msgType === 'filter') {
          _that.filterMapList[tabId] = _that.filterMapList[tabId] || {}
          _that.filterMapList[tabId][msg] = (_that.filterMapList[tabId][msg] || 0) + 1
        } else if (msgType === 'error_list') {
          if (type.IsArray(msg)) {
            _that.errorMapList[tabId] = msg || []
          }
        } else if (msgType === 'filter_list') {
          if (type.IsObject(msg)) {
            _that.filterMapList[tabId] = msg || {}
          }
        }

        // 限制长度
        const maxLen = 100000
        if (_that.contentMapList[tabId] && _that.contentMapList[tabId].length > maxLen) {
          _that.contentMapList[tabId] = _that.contentMapList[tabId].slice(-maxLen)
        }
        // 仅在内容长度发生变化时才重新格式化，避免选区期间无意义的 DOM 重建
        if (!_that._lastFormattedLen) _that._lastFormattedLen = {}
        let curLen = (_that.contentMapList[tabId] || '').length
        if (_that._lastFormattedLen[tabId] !== curLen) {
          let txt = _that.contentMapList[tabId] || ''
          txt = format.formatResult(txt, ['copy', 'color', 'replace'])
          txt = format.formatResult(txt, ['length'])
          _that.contentMapList[tabId] = txt
          _that._lastFormattedLen[tabId] = txt.length
        }
        _that.updateErrorTotal()
      }

      es.onerror = function (err) {
        console.error('[Fullpage-SSE] 连接错误', err)
      }
      })
    },
    // 关闭 Fullpage 独立 SSE 连接
    closeFullpageSse: function () {
      let _that = this
      if (_that.fullpageSseConn) {
        console.log('[Fullpage-SSE] 关闭连接 client_id=' + _that.fullpageClientId)
        _that.fullpageSseConn.close()
        _that.fullpageSseConn = null
        _that.fullpageClientId = ''
      }
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
      _that.resetRuleRuntimeState(tabId)
      _that.contentMapList[tabId] = ''
      _that.ensureRuleSetInfo(item.rule_set_id)

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
      _that.rawBufferMap[tabId] = ''
      _that.throttleStringFunc[tabId] = new Throttle_string(50, text => {
        // 检查用户是否在内容区有活动文本选区，若有则缓冲增量避免 DOM 更新打断选区
        if (_that.$refs.shellRef && _that.$refs.shellRef.hasActiveSelection()) {
          _that.rawBufferMap[tabId] += text
          _that.selectingTabIds[tabId] = true
          return
        }
        // 先刷新之前缓冲的内容
        if (_that.rawBufferMap[tabId]) {
          _that.contentMapList[tabId] = (_that.contentMapList[tabId] || '') + _that.rawBufferMap[tabId]
          _that.rawBufferMap[tabId] = ''
          _that.selectingTabIds[tabId] = false
        }
        _that.contentMapList[tabId] = (_that.contentMapList[tabId] || '') + text
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
              _that.resetRuleRuntimeState(tabId)
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
              _that.resetRuleRuntimeState(tabId)
              _that.$forceUpdate()
            }
          }
        }
      })
    },
    up : function (){
      this.isReceive = !this.isReceive
    },
    // onShellSelectionChange - 当 result_div 检测到选区状态变化时回调
    // 用户清除选区后，刷新缓冲区中的增量内容
    onShellSelectionChange: function (isSelecting) {
      let _that = this
      if (isSelecting) return
      // 选区已清除，刷新所有标签页的缓冲区
      for (let tabId in _that.rawBufferMap) {
        if (_that.rawBufferMap[tabId]) {
          _that.contentMapList[tabId] = (_that.contentMapList[tabId] || '') + _that.rawBufferMap[tabId]
          _that.rawBufferMap[tabId] = ''
          _that.selectingTabIds[tabId] = false
        }
      }
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
              _that.editTabConfigData.ssh_id !== oldTabConfig.ssh_id ||
              parseInt(_that.editTabConfigData.rule_set_id || 0) !== parseInt(oldTabConfig.rule_set_id || 0)) {
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
              _that.tabConfigList[i].rule_set_id = _that.editTabConfigData.rule_set_id
            }
          }
          _that.ensureRuleSetInfo(_that.editTabConfigData.rule_set_id)
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
      _that.editTabConfigData.rule_set_id = ''
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
        rule_set_id: _that.editTabConfigData.rule_set_id,
      }
      // 调接口
      shell.ShellOutStart(tabConfig, (res) => {
        let tabId = res.Data.id
        let sse_distribute_id = sseDistribute.GetSseDistributeId(tabId)
        let shellClientId = res.Data.shell_client_id
        tabConfig.sse_distribute_id = sse_distribute_id
        _that.scrollMap[tabId] = true
        _that.resetRuleRuntimeState(tabId)
        _that.contentMapList[tabId] = ''
        _that.activeTabId = tabId
        tabConfig.shell_client_id = shellClientId
        tabConfig.id = tabId
        _that.ensureRuleSetInfo(tabConfig.rule_set_id)
        _that.tabConfigList.push(tabConfig)

        //创建sse
        _that.createSseByTabConfig(sse_distribute_id, tabConfig)
      })

      // 清空输入
      _that.cleanEditTabConfigData()
    },
    createSseByTabConfig: function (sse_distribute_id, tabConfig) {
      let _that = this
      _that.createFullpageSseConnection(sse_distribute_id, tabConfig.id, function () {
        shell.ShellOutSetSeeId({
          sse_distribute_id: sse_distribute_id,
          shell_client_id: tabConfig.shell_client_id,
          ssh_id: tabConfig.ssh_id,
          command: tabConfig.command,
          id: tabConfig.id,
          group_id: tabConfig.group_id,
          rule_set_id: tabConfig.rule_set_id,
          is_run: _that.urlParams.id ? 0 : 1,
          sse_client_id: _that.fullpageClientId,
        }, function (res) {
          if (res.ErrCode !== 0) {
            _that.$helperNotify.error('建立链接失败: ' + (res.ErrMsg || ''))
          } else {
            // 绑定成功后拉取当前过滤关键词，回填输入框
            shellOut.ShellOutGetFilter({
              'shell_client_id': tabConfig.shell_client_id,
            }, function (filterRes) {
              if (filterRes.ErrCode === 0 && filterRes.Data && filterRes.Data.keywords && filterRes.Data.keywords.length > 0) {
                _that.filterContent = filterRes.Data.keywords.join(' ')
                _that.isFilterActive = true
              } else {
                _that.filterContent = ''
                _that.isFilterActive = false
              }
            })
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

<style scoped lang="scss" src="@/css/components/shellout/ShellOut.scss"></style>

<style>
/* 过滤激活按钮动画：全局非 scoped，确保穿透 pl-button → el-button 组件层级 */
.filter-active-btn {
  animation: filter-pulse 1.5s ease-in-out infinite alternate;
}

@keyframes filter-pulse {
  from {
    outline: 2px solid #e6a23c;
    outline-offset: 1px;
  }
  to {
    outline: 2px solid #f56c6c;
    outline-offset: 1px;
  }
}
</style>

