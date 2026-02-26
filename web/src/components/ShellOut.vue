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
          <el-button type="primary" @click="createTab">
            <el-icon><Plus /></el-icon>创建
          </el-button>
          <el-button @click="groupDialog = true">
            <el-icon><FolderOpened /></el-icon>分组管理
          </el-button>
          <el-button type="success">
            <el-icon><DataLine /></el-icon>运行总览
          </el-button>
        </div>
      </div>
    </div>

    <!-- 执行卡片列表 -->
    <template v-for="tab in tabConfigList" :key="tab.id">
      <div v-if="getExecutionInfo(tab.id) && (!chooseGroupId || parseInt(chooseGroupId) === -1 || Number(tab.group_id) === Number(chooseGroupId))" class="execution-card">
        <div class="card-header">
          <div class="card-info">
            <el-tag class="tab-id-tag">#{{ tab.id }}</el-tag>
            <span class="tab-name">{{ tab.name }}</span>
          </div>
          <div class="card-actions">
            <el-button size="small" @click="showEditTabConfig(tab.id)">
              <el-icon><Edit /></el-icon>编辑
            </el-button>
            <el-button size="small" @click="showCopyCreateTabConfig(tab.id)">
              <el-icon><CopyDocument /></el-icon>复制
            </el-button>
            <el-button size="small" type="danger" @click="removeTab(tab.id)">
              <el-icon><Delete /></el-icon>删除
            </el-button>
            <el-button size="small" type="primary" @click="openNewTab(tab)">
              <el-icon><Position /></el-icon>新窗口
            </el-button>
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
        <el-button @click="shellOutDialog = false">取消</el-button>
        <el-button type="primary" @click="executeCommand" style="min-width: 120px;">
          <el-icon><Check /></el-icon>{{ editTabConfigData.id ? '保存更改' : '创建' }}
        </el-button>
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
import { Plus, FolderOpened, DataLine, Edit, CopyDocument, Delete, Position, CollectionTag, Check, Terminal } from '@element-plus/icons-vue';
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

<style lang="scss" scoped>
/* 页面容器 */
.shell-page-container {
  padding: 0;
  width: 100%;
}

/* 顶部卡片样式 */
.shell-header-card {
  background: linear-gradient(135deg, #3498db 0%, #2980b9 100%);
  border-radius: 16px;
  padding: 20px 24px;
  margin-bottom: 20px;
  box-shadow: 0 4px 20px rgba(52, 152, 219, 0.25);
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

.group-select {
  width: 200px;
}

.group-select :deep(.el-input__wrapper) {
  background: rgba(255, 255, 255, 0.95);
  border-radius: 10px;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
}

.action-buttons {
  display: flex;
  gap: 8px;
}

.action-buttons .el-button:not([class*="el-button--"]) {
  background: rgba(255, 255, 255, 0.9);
  border: none;
  color: #333;
}

.action-buttons .el-button:not([class*="el-button--"]):hover {
  background: #fff;
}

.action-buttons .el-button--primary {
  background: linear-gradient(135deg, #409eff 0%, #337ecc 100%);
  border: none;
  color: #fff;
}

.action-buttons .el-button--primary:hover {
  background: linear-gradient(135deg, #66b1ff 0%, #409eff 100%);
}

.action-buttons .el-button--success {
  background: linear-gradient(135deg, #67c23a 0%, #529b2e 100%);
  border: none;
  color: #fff;
}

.action-buttons .el-button--success:hover {
  background: linear-gradient(135deg, #85ce61 0%, #67c23a 100%);
}

/* 执行卡片 */
.execution-card {
  background: #fff;
  border-radius: 16px;
  padding: 20px;
  margin-bottom: 16px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.08);
  border-left: 4px solid #3498db;
  transition: all 0.3s ease;
}

.execution-card:hover {
  box-shadow: 0 4px 20px rgba(0, 0, 0, 0.12);
  transform: translateY(-2px);
}

.card-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 12px;
  flex-wrap: wrap;
  gap: 12px;
}

.card-info {
  display: flex;
  align-items: center;
  gap: 12px;
}

.tab-id-tag {
  background: linear-gradient(135deg, #3498db 0%, #2980b9 100%);
  color: #fff;
  font-weight: 600;
  border: none;
  font-size: 13px;
}

.tab-name {
  font-size: 16px;
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
  background: #f8f9fc;
  padding: 12px 16px;
  border-radius: 10px;
}

.command-icon {
  color: #3498db;
  font-size: 18px;
  flex-shrink: 0;
  margin-top: 2px;
}

.command-text {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 13px;
  color: #606266;
  word-break: break-all;
  line-height: 1.5;
}

/* 弹窗样式 */
.create-dialog :deep(.el-dialog),
.group-dialog :deep(.el-dialog) {
  border-radius: 16px;
}

.create-dialog :deep(.el-dialog__header),
.group-dialog :deep(.el-dialog__header) {
  border-bottom: 1px solid #ebeef5;
  padding: 16px 20px;
  margin: 0;
}

.create-dialog :deep(.el-dialog__body),
.group-dialog :deep(.el-dialog__body) {
  padding: 20px;
}

.create-dialog :deep(.el-dialog__footer),
.group-dialog :deep(.el-dialog__footer) {
  border-top: 1px solid #ebeef5;
  padding: 12px 20px;
}

.create-form :deep(.el-input__wrapper),
.create-form :deep(.el-textarea__inner) {
  border-radius: 8px;
}

.create-form :deep(.el-select .el-input__wrapper) {
  border-radius: 8px;
}

/* 响应式 */
@media (max-width: 768px) {
  .control-row {
    flex-direction: column;
    align-items: stretch;
  }
  
  .group-select {
    width: 100%;
  }
  
  .action-buttons {
    flex-wrap: wrap;
  }
  
  .card-header {
    flex-direction: column;
    align-items: flex-start;
  }
  
  .card-actions {
    width: 100%;
    justify-content: flex-start;
  }
}

// 标签页样式
::deep(.el-tabs--card > .el-tabs__header .el-tabs__item.is-active) {
  background-color: #ecf5ff !important;
  border: 1px solid #409eff !important;
  border-bottom-color: #409eff !important;
  color: #409eff;
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
  white-space: pre-line;
}

// 执行信息区域样式
.execution-info {
  margin-bottom: 10px;
  padding: 8px;
  background: #fff;
  border-radius: 4px;
  box-shadow: 0 1px 3px rgba(0, 0, 0, 0.1);
  font-size : 14px;
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
    font-size: 14px;
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

.error-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 8px;
  padding-bottom: 8px;
  border-bottom: 1px solid #fbc4c4;
  flex-wrap: wrap;
  gap: 8px;
}

.error-context-info {
  font-size: 14px;
  color: #67c23a;
  background: #f0f9eb;
  padding: 2px 6px;
  border-radius: 3px;
  flex-basis: 100%;
}

.error-time {
  font-size: 14px;
  color: #909399;
}

.error-line {
  font-size: 14px;
  color: #606266;
  background: #e6e6e6;
  padding: 2px 6px;
  border-radius: 3px;
}

.error-content {
  white-space: pre-line;
  background: #2d2d2d;
  color: #e0e0e0;
  padding: 12px;
  border-radius: 4px;
  font-family: 'Consolas', 'Courier New', monospace;
  font-size: 14px;
  line-height: 1.4;
  word-break: break-all;
  margin: 0;

  :deep(.error-highlight) {
    color: #ff6b6b;
    font-weight: bold;
    background: rgba(255, 107, 107, 0.1);
    padding: 2px 4px;
    border-radius: 3px;

    &.error-critical {
      color: #ff4757;
      background: rgba(255, 71, 87, 0.1);
    }

    &.error-warning {
      color: #ffa502;
      background: rgba(255, 165, 2, 0.1);
    }

    &.error-database {
      color: #a29bfe;
      background: rgba(162, 155, 254, 0.1);
      border-left: 3px solid #a29bfe;
    }

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
  padding: 40px;
  font-size: 14px;
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

// 优化按钮样式
::deep(.el-button) {
  border-radius: 6px;
  transition: all 0.3s ease;

  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }
}

// 优化表单样式
::deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

::deep(.el-input__wrapper),
::deep(.el-textarea__inner) {
  border-radius: 8px;
  transition: all 0.3s ease;

  &:focus-within {
    box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
  }
}

// 优化对话框样式
::deep(.el-dialog) {
  border-radius: 12px;
  overflow: hidden;
}

::deep(.el-dialog__header) {
  color: white;
  margin: 0;
}

::deep(.el-dialog__body) {
  //padding: 20px;
}
</style>
