<template>
  <div class="shell-console">
    <!-- 操作栏 -->
    <div class="toolbar">
      <el-select
          v-model="chooseGroupId"
          class="select"
          placeholder="筛选分组"
          @change="chooseGroupIdChange"
      >
        <el-option
            key="0"
            label="全部"
            value="-1"
        />
        <el-option
            v-for="g in groupList"
            :key="g.id"
            :label="g.name"
            :value="String(g.id)"
        />
      </el-select>&nbsp;
      <el-button
          type="primary"
          @click="createTab"
      >
        创建
      </el-button>
      <el-button
          type="info"
          @click="groupDialog = true"
      >
        分组管理
      </el-button>
      <el-button
          type="success"
          @click=""
      >
        运行总览
      </el-button>
    </div>

    <template v-for="tab in tabConfigList" :key="tab.id">
      <!-- 执行信息显示区域 -->
      <div v-if="getExecutionInfo(tab.id) && (!chooseGroupId || parseInt(chooseGroupId) === -1 || Number(tab.group_id) === Number(chooseGroupId))" class="execution-info-card">
        <div class="execution-header">
          <div class="execution-title">
            <span class="tab-id">#{{ tab.id }}</span>
            <span class="tab-name">{{ tab.name }}</span>
          </div>
          <div class="execution-command">
            {{ getExecutionInfo(tab.id).command }}
          </div>
        </div>

        <div class="execution-actions">
          <el-button
              size="default"
              icon="Edit"
              @click="showEditTabConfig(tab.id)"
          >
            编辑
          </el-button>
          <el-button
              size="default"
              icon="CopyDocument"
              @click="showCopyCreateTabConfig(tab.id)"
          >
            复制
          </el-button>
          <el-button
              type="danger"
              size="default"
              icon="Delete"
              @click="removeTab(tab.id)"
          >
            删除
          </el-button>
          <el-button
              type="primary"
              size="default"
              icon="Position"
              @click="openNewTab(tab)">
            新窗口
          </el-button>
        </div>
      </div>
    </template>

    <el-dialog
        v-model="shellOutDialog"
        title="创建终端输出"
        width="50%"
        destroy-on-close
    >
      <el-form
          ref="createFormRef"
          :model="editTabConfigData"
          label-width="100px"
          style=" "
      >
        <el-form-item
            label="名称"
            prop="name"
            :rules="[{ required: true, message: '请输入名称', trigger: 'blur' }]"
        >
          <el-input
              v-model="editTabConfigData.name"
              placeholder="请输入名称"
              clearable
          />
        </el-form-item>
        <el-form-item
            label="SSH 环境"
            prop="ssh_id"
            :rules="[{ required: true, message: '请选择 SSH 环境', trigger: 'change' }]"
        >
          <el-select
              v-model="editTabConfigData.ssh_id"
              placeholder="请选择 SSH 环境"
              style="width: 100%"
          >
            <el-option
                v-for="s in sshList"
                :key="s.id"
                :label="s.name"
                :value="s.id"
            />
          </el-select>
        </el-form-item>

        <el-form-item label="分组">
          <el-select
              v-model="editTabConfigData.group_id"
              placeholder="请选择分组"
              style="width: 100%"
          >
            <el-option
                v-for="g in groupList"
                :key="g.id"
                :label="g.name"
                :value="g.id"
            />
          </el-select>
        </el-form-item>
        <el-form-item
            label="命令"
            prop="command"
            :rules="[{ required: true, message: '请输入命令', trigger: 'blur' }]"
        >
          <el-input
              v-model="editTabConfigData.command"
              placeholder="请输入命令"
              type="textarea"
              :rows="4"
              clearable
          />
        </el-form-item>
        <el-form-item>
          <el-button
              type="primary"
              @click="executeCommand"
              style="width: 100%"
          >
            {{ editTabConfigData.id ? '保存更改' : '创建' }}
          </el-button>
        </el-form-item>
      </el-form>
    </el-dialog>

    <el-dialog
        v-model="groupDialog"
        title="分组管理"
        width="60%"
    >
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
.shell-console {
  padding: 20px;
  background: linear-gradient(135deg, #f5f7fa 0%, #e4edf5 100%);
  min-height: 100vh;
}

.toolbar {
  display: flex;
  align-items: center;
  margin-bottom: 20px;
  padding: 16px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);

  .select {
    width: 220px;
    margin-right: 12px;
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

.execution-info-card {
  margin-bottom: 16px;
  padding: 16px;
  background: white;
  border-radius: 12px;
  box-shadow: 0 2px 12px rgba(0, 0, 0, 0.05);
  border-left: 4px solid #409eff;
  transition: all 0.3s ease;

  &:hover {
    box-shadow: 0 4px 20px rgba(0, 0, 0, 0.1);
    transform: translateY(-2px);
  }

  .execution-header {
    margin-bottom: 12px;

    .execution-title {
      display: flex;
      align-items: center;
      margin-bottom: 8px;

      .tab-id {
        font-weight: bold;
        color: #409eff;
        margin-right: 8px;
        background: #ecf5ff;
        padding: 2px 6px;
        border-radius: 4px;
        font-size: 16px;
      }

      .tab-name {
        font-size: 16px;
        font-weight: 600;
        color: #303133;
      }
    }

    .execution-command {
      font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
      font-size: 14px;
      color: #606266;
      background: #f8f9fa;
      padding: 8px 12px;
      border-radius: 6px;
      border-left: 3px solid #dcdfe6;
      word-break: break-all;
    }
  }

  .execution-actions {
    display: flex;
    gap: 8px;
    justify-content: flex-end;
  }
}

:deep(.el-tabs--card > .el-tabs__header .el-tabs__item.is-active) {
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
:deep(.el-button) {
  border-radius: 6px;
  transition: all 0.3s ease;

  &:hover {
    transform: translateY(-1px);
    box-shadow: 0 2px 8px rgba(0, 0, 0, 0.1);
  }
}

// 优化表单样式
:deep(.el-form-item__label) {
  font-weight: 500;
  color: #606266;
}

:deep(.el-input__wrapper),
:deep(.el-textarea__inner) {
  border-radius: 8px;
  transition: all 0.3s ease;

  &:focus-within {
    box-shadow: 0 0 0 2px rgba(64, 158, 255, 0.2);
  }
}

// 优化对话框样式
:deep(.el-dialog) {
  border-radius: 12px;
  overflow: hidden;
}

:deep(.el-dialog__header) {
  //background: linear-gradient(135deg, #409eff 0%, #337ecc 100%);
  color: white;
  //padding: 16px 20px;
  margin: 0;
}

:deep(.el-dialog__body) {
  //padding: 20px;
}
</style>



