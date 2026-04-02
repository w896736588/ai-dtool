<template>
  <div class="supervisor-config-page">
    <div class="page-header">
      <div class="header-content">
        <div class="header-left">
          <svg class="header-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <rect x="3" y="3" width="18" height="18" rx="2" stroke="currentColor" stroke-width="2"/>
            <path d="M9 9H15" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            <path d="M9 13H15" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
            <path d="M9 17H12" stroke="currentColor" stroke-width="2" stroke-linecap="round"/>
          </svg>
          <h3>Supervisor 配置管理</h3>
        </div>
        <pl-button type="primary" @click="ShowAddSupervisor">
          <el-icon><Plus /></el-icon>
          添加配置
        </pl-button>
      </div>
      <p class="header-desc">管理Supervisor进程监控配置，支持SSH远程和Docker环境</p>
    </div>

    <div class="config-table-card">
      <el-table :data="state.supervisorList" style="width: 100%" class="config-table" stripe>
        <el-table-column prop="id" label="#" width="60" align="center" />
        <el-table-column prop="name" label="名称" min-width="120">
          <template #default="scope">
            <div class="name-cell">
              <el-icon class="config-icon"><Setting /></el-icon>
              <span class="name-text">{{ scope.row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="ssh_name" label="SSH" width="140">
          <template #default="scope">
            <el-tag v-if="scope.row.ssh_id && scope.row.ssh_id !== '0'" size="small" type="success">
              {{ scope.row.ssh_name || '已配置' }}
            </el-tag>
            <el-tag v-else size="small" type="info">本地</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="docker_name" label="Docker" width="140">
          <template #default="scope">
            <el-tag v-if="scope.row.docker_name" size="small" type="warning">
              {{ scope.row.docker_name }}
            </el-tag>
            <span v-else class="empty-text">-</span>
          </template>
        </el-table-column>
        <el-table-column prop="config_dir" label="配置目录" min-width="250">
          <template #default="scope">
            <code class="dir-code">{{ scope.row.config_dir || '-' }}</code>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="scope">
            <div class="action-buttons">
              <pl-button type="primary" link size="small" @click="ShowEditSupervisor(scope.row , true)">
                <el-icon><CopyDocument /></el-icon>复制新增
              </pl-button>
              <pl-button type="primary" link size="small" @click="ShowEditSupervisor(scope.row , false)">
                <el-icon><Edit /></el-icon>编辑
              </pl-button>
              <pl-button type="danger" link size="small" @click="DeleteSupervisor(scope.row)">
                <el-icon><Delete /></el-icon>删除
              </pl-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <!-- 编辑弹窗 -->
    <el-dialog v-model="state.dialogEditSupervisor" title="编辑Supervisor配置" width="500px" :close-on-click-modal="false" class="edit-dialog">
      <el-form label-width="90px" class="edit-form">
        <el-form-item label="名称">
          <el-input v-model="state.editSupervisorConfig.name" autocomplete="off" placeholder="配置名称">
            <template #prefix>
              <el-icon><CollectionTag /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="Docker名称">
          <el-input v-model="state.editSupervisorConfig.docker_name" autocomplete="off" placeholder="可选，Docker容器名">
            <template #prefix>
              <el-icon><Box /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="配置目录">
          <el-input v-model="state.editSupervisorConfig.config_dir" autocomplete="off" placeholder="supervisord.conf所在目录">
            <template #prefix>
              <el-icon><Folder /></el-icon>
            </template>
          </el-input>
        </el-form-item>
        <el-form-item label="SSH隧道">
          <el-select v-model="state.editSupervisorConfig.ssh_id" placeholder="选择SSH连接" style="width: 100%">
            <el-option v-for="item in state.sshList" :key="item.id" :label="item.name" :value="item.id"/>
          </el-select>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogEditSupervisor = false">取消</pl-button>
          <pl-button type="primary" @click="EditSupervisor">
            <el-icon><Check /></el-icon>保存
          </pl-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>
<script>
import {defineComponent, getCurrentInstance, reactive, onActivated} from 'vue';
import { Plus, Setting, CopyDocument, Edit, Delete, CollectionTag, Box, Folder, Check } from '@element-plus/icons-vue';
import ssh_set from '../../utils/base/ssh_set'
import set from '../../utils/base/supervisor_set'
import common from '../../utils/common'
import list from "@/utils/base/list";
import Init from "@/utils/base/set_init";
export default defineComponent({
  components: { Plus, Setting, CopyDocument, Edit, Delete, CollectionTag, Box, Folder, Check },
  props: {
  },
  emits: ['changed'],
  data() {
    return {
    }
  },
  setup(props, { emit }) {
    onActivated(() => {
      if(Init.GetIsInit('supervisor') === true){
        SupervisorList()
        SshList()
        Init.DelInit('supervisor')
      }
    });
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties
    const SupervisorList = function (){
      set.SupervisorList(function (response){
        if(response.ErrCode === 0){
          state.supervisorList = response.Data
        }
      })
    }
    const ShowEditSupervisor = function (supervisorConfig , isCopy){
      state.dialogEditSupervisor = true
      state.editSupervisorConfig = supervisorConfig
      if(isCopy){
        state.editSupervisorConfig.id = 0
      }
    }
    const SetInit = function(){
      Init.SetIsInit('supervisor')
    }
    const ShowAddSupervisor = function (){
      state.dialogEditSupervisor = true
      state.editSupervisorConfig = {}
    }
    // emitChanged 告知宿主页面 Supervisor 配置已变化，便于刷新环境列表。
    // Notify host pages when Supervisor settings changed so environment lists can refresh immediately.
    const emitChanged = function (){
      emit('changed')
    }
    const EditSupervisor = function (){
      set.SupervisorAdd(state.editSupervisorConfig , function (response){
        if(response.ErrCode === 0){
          SupervisorList()
          emitChanged()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
        state.dialogEditSupervisor = false
        SetInit()
      })
    }
    const DeleteSupervisor = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.SupervisorDelete(rowData , function (response){
          if(response.ErrCode === 0){
            SupervisorList()
            emitChanged()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
          SetInit()
        })
      })
    }

    const SshList = function (){
      ssh_set.SshList(function (response){
        if(response.ErrCode === 0){
          state.sshList = response.Data
        }
      })
    }
    //固有属性
    const state = reactive({
      sshList :[],
      supervisorList : [],
      dialogEditSupervisor : false,
      editSupervisorConfig : {},
      filterValue : '',
    })
    //初始化
    SupervisorList()
    SshList()
    return {
      state,
      ShowEditSupervisor,
      ShowAddSupervisor,
      EditSupervisor,
      DeleteSupervisor,
      SupervisorList,
      SshList,
    }
  },
  mounted() {

  },
  methods: {
  },
})
</script>

<style scoped>
@import "@/css/set_module_unified.css";

.supervisor-config-page {
  padding: 0;
}

.page-header {
  background: #fff;
  border: 1px solid #e8e8e0;
  border-radius: 12px;
  padding: 16px;
  margin-bottom: 12px;
}

.header-content {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 8px;
}

.header-left {
  display: flex;
  align-items: center;
  gap: 10px;
}

.header-icon {
  width: 22px;
  height: 22px;
  color: #5a8a5a;
}

.header-left h3 {
  color: #4a4a4a;
  margin: 0;
  font-size: 18px;
  font-weight: 600;
}

.header-desc {
  color: #7a7a6a;
  margin: 0;
  font-size: 13px;
}

.config-table-card {
  background: #fff;
  border: 1px solid #e8e8e0;
  border-radius: 12px;
  padding: 12px;
}

.config-table {
  border-radius: 10px;
  overflow: hidden;
}

.config-table :deep(.el-table__header th) {
  background: #f7f7f2;
  color: #606050;
  font-weight: 600;
}

.config-table :deep(.el-table__row:hover > td) {
  background-color: #f3f7ef !important;
}

.name-cell {
  display: flex;
  align-items: center;
  gap: 8px;
}

.config-icon {
  color: #5a8a5a;
  font-size: 16px;
}

.name-text {
  font-weight: 500;
  color: #303133;
}

.empty-text {
  color: #c0c4cc;
}

.dir-code {
  font-family: Consolas, Monaco, monospace;
  font-size: 13px;
  color: #4f804f;
  background: #f3f8ef;
  padding: 2px 8px;
  border-radius: 4px;
}

.action-buttons {
  display: flex;
  gap: 4px;
}

.edit-form :deep(.el-input__wrapper),
.edit-form :deep(.el-select .el-input__wrapper) {
  border-radius: 8px;
}
</style>

