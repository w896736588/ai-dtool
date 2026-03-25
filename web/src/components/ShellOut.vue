<template>
  <div class="shell-page-container">
    <!-- 顶部操作区域 -->
    <div class="shell-header-card">
      <div class="header-title">
        <svg class="header-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
          <rect x="2" y="3" width="20" height="14" rx="2" stroke="currentColor" stroke-width="2"/>
          <path d="M6 7L10 10L6 13" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          <path d="M12 13H18" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
        </svg>
        <span>终端输出管理</span>
      </div>
      <div class="control-row">
        <el-select v-model="chooseGroupId" class="group-select" placeholder="筛选分组" @change="chooseGroupIdChange" clearable>
          <el-option key="0" label="全部" value="-1" />
          <el-option v-for="g in groupList" :key="g.id" :label="g.name" :value="String(g.id)" />
        </el-select>
        <div class="action-buttons">
          <pl-button type="primary" @click="createTab">
            <el-icon><Plus /></el-icon>创建
          </pl-button>
          <pl-button @click="groupDialog = true">
            <el-icon><FolderOpened /></el-icon>分组管理
          </pl-button>
          <pl-button type="success">
            <el-icon><DataLine /></el-icon>运行总览
          </pl-button>
        </div>
        <!-- 本地搜索框 -->
        <el-input
          v-model="localSearchKey"
          placeholder="搜索名称/命令，空格多条件"
          class="search-input"
          clearable
        >
          <template #prefix>
            <el-icon><Search /></el-icon>
          </template>
        </el-input>
      </div>
    </div>

    <!-- 执行卡片列表 -->
    <template v-for="tab in tabConfigList" :key="tab.id">
      <div v-if="getExecutionInfo(tab.id) && (!chooseGroupId || parseInt(chooseGroupId) === -1 || Number(tab.group_id) === Number(chooseGroupId)) && matchSearch(tab)" class="execution-card">
        <div class="card-header">
          <div class="card-info">
            <el-tag class="tab-id-tag">#{{ tab.id }}</el-tag>
            <span class="tab-name">{{ tab.name }}</span>
          </div>
          <div class="card-actions">
            <pl-button size="small" @click="showEditTabConfig(tab.id)">
              <el-icon><Edit /></el-icon>编辑
            </pl-button>
            <pl-button size="small" @click="showCopyCreateTabConfig(tab.id)">
              <el-icon><CopyDocument /></el-icon>复制
            </pl-button>
            <pl-button size="small" type="danger" @click="removeTab(tab.id)">
              <el-icon><Delete /></el-icon>删除
            </pl-button>
            <pl-button size="small" type="primary" @click="openNewTab(tab)">
              <el-icon><Position /></el-icon>新窗口
            </pl-button>
          </div>
        </div>
        <div class="card-command">
          <el-icon class="command-icon"><Terminal /></el-icon>
          <code class="command-text">{{ getExecutionInfo(tab.id).command }}</code>
        </div>
      </div>
    </template>

    <!-- 创建/编辑弹窗 -->
    <el-dialog v-model="shellOutDialog" title="创建终端输出" width="550px" destroy-on-close class="create-dialog">
      <el-form ref="createFormRef" :model="editTabConfigData" label-width="90px" class="create-form">
        <el-form-item label="名称" prop="name" :rules="[{ required: true, message: '请输入名称', trigger: 'blur' }]">
          <el-input v-model="editTabConfigData.name" placeholder="输入任务名称" clearable>
            <template #prefix>
              <el-icon><CollectionTag /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="SSH环境" prop="ssh_id" :rules="[{ required: true, message: '请选择 SSH 环境', trigger: 'change' }]">
          <el-select v-model="editTabConfigData.ssh_id" placeholder="选择SSH服务器" style="width: 100%">
            <el-option v-for="s in sshList" :key="s.id" :label="s.name" :value="s.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="分组">
          <el-select v-model="editTabConfigData.group_id" placeholder="选择分组" style="width: 100%">
            <el-option v-for="g in groupList" :key="g.id" :label="g.name" :value="g.id" />
          </el-select>
        </el-form-item>
        <el-form-item label="命令" prop="command" :rules="[{ required: true, message: '请输入命令', trigger: 'blur' }]">
          <el-input v-model="editTabConfigData.command" placeholder="输入要执行的命令" type="textarea" :rows="4" clearable />
        </el-form-item>
      </el-form>
      <template #footer>
        <pl-button @click="shellOutDialog = false">取消</pl-button>
        <pl-button type="primary" @click="executeCommand" style="min-width: 120px;">
          <el-icon><Check /></el-icon>{{ editTabConfigData.id ? '保存更改' : '创建' }}
        </pl-button>
      </template>
    </el-dialog>

    <!-- 分组管理弹窗 -->
    <el-dialog v-model="groupDialog" title="分组管理" width="70%" class="group-dialog">
      <Group
        :extra1Title="'过滤正则'"
        :extra1Type="'textarea'"
        :extra2Title="'错误捕获正则'"
        :extra2Type="'textarea'"
        :extra3Title="'排除捕获的错误'"
        :extra3Type="'textarea'"
        :groupTitle="'终端输出'"
        :groupType="groupType"
        @update="groupUpdate">
      </Group>
    </el-dialog>
  </div>
</template>

<script>
/* 以下 import 保持你原来的即可 */
import { Plus, FolderOpened, DataLine, Edit, CopyDocument, Delete, Position, CollectionTag, Check, Terminal, Search } from '@element-plus/icons-vue';
import base from '@/utils/base.js'
import sse from '@/utils/base/sse'
import shell from '@/utils/base/shell'
import ssh from '@/utils/base/ssh_set'
import {ref, onMounted} from 'vue'
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
import Typ from '@/utils/base/type'

const StoreChooseGroupIdKey = 'shell_out_choose_group_id'
const StoreChooseShellOutKey = 'shell_out_choose_shell_group'
export default {
  components: {
    shellResult,
    Group,
    Plus,
    FolderOpened,
    DataLine,
    Edit,
    CopyDocument,
    Delete,
    Position,
    CollectionTag,
    Check,
    Terminal,
    Search,
  },
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
      shellOutDialog: false,
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
      scrollMap: {},
      tabConfigList: [],//配置
      groupList: [], //分组列表
      groupType: `6`,
      urlParams: {},
      // local search
      localSearchKey: '',
    }
  },
  mounted() {
    let _that = this
    _that.loadSshList()
    _that.windowChange()
    window.addEventListener('resize', function () {
      _that.windowChange()
    });
    shell.calculateShellDivHeight(_that)
    _that.getGroupList()
    _that.getFullPageParams()
    //如果是单独展示的页面 里面返回的就是传参的
    _that.chooseGroupId = _that.getStoreGroupId()
  },
  activated: function () {
    let _that = this
    _that.windowChange()
  },
  deactivated() {

  },
  methods: {
    // Local search filter
    matchSearch: function (tab) {
      let _that = this
      if (!_that.localSearchKey || _that.localSearchKey.trim() === '') {
        return true
      }
      let keywords = _that.localSearchKey.trim().split(/\s+/)
      for (let keyword of keywords) {
        if (keyword === '') continue
        let lowerKeyword = keyword.toLowerCase()
        if ((tab.name && tab.name.toLowerCase().indexOf(lowerKeyword) !== -1) ||
            (tab.command && tab.command.toLowerCase().indexOf(lowerKeyword) !== -1)) {
          return true
        }
      }
      return false
    },
    openNewTab : function(tab){
      let url = window.location.origin +'/#/fullpage?group_id=' +this.chooseGroupId+'&id='+tab.id+'&title=' + tab.name
      window.open(url, '_blank');
    },
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
    groupUpdate: function () {
      this.getGroupList()
    },
    getGroupList: function () {
      let _that = this
      group.GroupList({type: _that.groupType}, function (response) {
        if (response.ErrCode === 0) {
          _that.groupList = response.Data
          _that.chooseGroupId = _that.getStoreGroupId()
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
    // 获取执行信息
    getExecutionInfo(tabId) {
      let _that = this
      return _that.getTabConfigById(tabId)
    },

    createTab: function () {
      let _that = this
      _that.shellOutDialog = true
      _that.editTabConfigData.id = ''
      _that.editTabConfigData.name = ''
      _that.editTabConfigData.command = ''
      _that.editTabConfigData.ssh_id = ''
      _that.editTabConfigData.group_id = ''
    },
    chooseGroupIdChange: function () {
      let _that = this
      store.setStore(StoreChooseGroupIdKey, _that.chooseGroupId)
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
    },
    createByTabId: function (tabId, item) {
      let _that = this
      const sseId = sseDistribute.GetSseDistributeId(tabId)
      _that.scrollMap[tabId] = true

      item.sse_id = sseId
      _that.tabConfigList.push(item)
      //如果是运行状态
      if (_that.urlParams.id) {
        if (parseInt(_that.urlParams.id) !== parseInt(item.id)) {
          item.is_run = 0
          return
        }
        item.is_run = 1
      }
    },
    stopByTabId: function (tabId, back) {
      let _that = this
      shellOut.ShellOutStop(_that.getTabConfigById(tabId), function (res) {
        if (res.ErrCode !== 0) {
          _that.$helperNotify.error('停止失败')
        } else {
          for (let i in _that.tabConfigList) {
            if (_that.tabConfigList[i].id === tabId) {
              sse.SseClose(_that.tabConfigList[i].sse_id)
              _that.tabConfigList[i].is_run = 0
              _that.tabConfigList[i].shell_client_id = ''
              _that.$forceUpdate()
            }
          }
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
      _that.shellOutDialog = false
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
        sse_id: '',
        shell_client_id: '',
        ssh_id: _that.editTabConfigData.ssh_id,
        name: _that.editTabConfigData.name,
        is_run: _that.urlParams.id ? 0 : 1,
        group_id: _that.editTabConfigData.group_id,
      }
      // 调接口
      shell.ShellOutStart(tabConfig, (res) => {
        let tabId = res.Data.id
        let sseId = sseDistribute.GetSseDistributeId(tabId)
        let shellClientId = res.Data.shell_client_id
        tabConfig.sse_id = sseId
        _that.scrollMap[tabId] = true
        tabConfig.shell_client_id = shellClientId
        tabConfig.id = tabId
        _that.tabConfigList.push(tabConfig)

      })

      // 清空输入
      _that.cleanEditTabConfigData()
    },
    showCopyCreateTabConfig: function (tabId) {
      let _that = this
      let tabConfig = _that.getTabConfigById(tabId)
      _that.editTabConfigData.id = ''
      _that.editTabConfigData.command = tabConfig.command
      _that.editTabConfigData.ssh_id = tabConfig.ssh_id
      _that.editTabConfigData.name = tabConfig.name
      _that.editTabConfigData.group_id = tabConfig.group_id
      _that.shellOutDialog = true
    },
    showEditTabConfig: function (tabId) {
      let _that = this
      let tabConfig = _that.getTabConfigById(tabId)
      _that.editTabConfigData.id = tabConfig.id
      _that.editTabConfigData.command = tabConfig.command
      _that.editTabConfigData.ssh_id = tabConfig.ssh_id
      _that.editTabConfigData.name = tabConfig.name
      _that.editTabConfigData.group_id = tabConfig.group_id
      _that.shellOutDialog = true
    },
    // 移除标签页
    removeTab(tabId) {
      let _that = this
      let tabConfig = _that.getTabConfigById(tabId)
      _that.$confirm(`确定要删除接口 "${tabConfig.name}" 吗？`, '确认删除', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        type: 'warning'
      }).then(() => {
        shellOut.ShellOutDelete(_that.getTabConfigById(tabId), function (res) {
          if (res.ErrCode !== 0) {
            _that.$helperNotify.error('删除失败')
          } else {
            for (let i in _that.tabConfigList) {
              if (_that.tabConfigList[i].id !== tabId) {
                break
              }
            }
            delete _that.scrollMap[tabId]
            for (let i in _that.tabConfigList) {
              if (_that.tabConfigList[i].id === tabId) {
                _that.tabConfigList.splice(i, 1)
                break
              }
            }
          }
        })
      }).catch(() => {
      })
    },
  }
}
</script>

<style scoped>
.shell-page-container {
  padding: 0;
  width: 100%;
  height: 100%;
  min-height: 100%;
  color: #4a4a4a;
}

.shell-header-card {
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

.group-select {
  width: 220px;
}

.group-select :deep(.el-input__wrapper),
.search-input :deep(.el-input__wrapper) {
  border-radius: 8px;
  background: #fff;
  box-shadow: 0 0 0 1px #dde3d8 inset;
}

.group-select :deep(.el-input__wrapper.is-focus),
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
  max-width: 360px;
  min-width: 200px;
}

.execution-card {
  background: #fff;
  border: 1px solid #e8e8e0;
  border-left: 3px solid #b8ceb6;
  border-radius: 12px;
  padding: 14px;
  margin-bottom: 10px;
}

.execution-card:hover {
  border-left-color: #93b793;
  background: #fcfdfb;
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 10px;
  flex-wrap: wrap;
  gap: 8px;
}

.card-info {
  display: flex;
  align-items: center;
  gap: 10px;
}

.tab-id-tag {
  background: #eaf3e6;
  color: #3f6f3f;
  border: 1px solid #c7dbc5;
  font-size: 12px;
}

.tab-name {
  font-size: 15px;
  font-weight: 600;
  color: #303133;
}

.card-actions {
  display: flex;
  gap: 8px;
  flex-wrap: wrap;
}

.card-command {
  display: flex;
  align-items: flex-start;
  gap: 10px;
  background: #f7f9f5;
  border: 1px solid #e6ece0;
  padding: 10px 12px;
  border-radius: 8px;
}

.command-icon {
  color: #5f8f5f;
  font-size: 16px;
  flex-shrink: 0;
}

.command-text {
  font-family: Consolas, Monaco, monospace;
  font-size: 13px;
  color: #5a5a5a;
  word-break: break-all;
  line-height: 1.45;
}

.create-form :deep(.el-input__wrapper),
.create-form :deep(.el-textarea__inner),
.create-form :deep(.el-select .el-input__wrapper) {
  border-radius: 8px;
}

@media (max-width: 768px) {
  .control-row {
    flex-direction: column;
    align-items: stretch;
  }

  .group-select,
  .search-input {
    width: 100%;
    max-width: 100%;
  }

  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }

  .card-actions {
    width: 100%;
  }
}
</style>

