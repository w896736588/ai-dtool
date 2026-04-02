<template>
  <div class="redis-config-page">
    <div class="page-header">
      <div class="header-content">
        <div class="header-left">
          <svg class="header-icon" viewBox="0 0 24 24" fill="none" xmlns="http://www.w3.org/2000/svg">
            <path d="M12 2L2 7L12 12L22 7L12 2Z" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 17L12 22L22 17" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
            <path d="M2 12L12 17L22 12" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"/>
          </svg>
          <h3>Redis 配置管理</h3>
        </div>
        <pl-button type="primary" @click="ShowAddRedis">
          <el-icon><Plus /></el-icon>
          添加Redis实例
        </pl-button>
      </div>
      <p class="header-desc">管理Redis服务器连接配置，支持直连和SSH隧道</p>
    </div>

    <div class="config-table-card">
      <el-table :data="state.redisList" style="width: 100%" class="config-table" stripe>
        <el-table-column prop="id" label="#" width="60" align="center" />
        <el-table-column prop="name" label="名称" min-width="150">
          <template #default="scope">
            <div class="name-cell">
              <span class="name-text">{{ scope.row.name }}</span>
            </div>
          </template>
        </el-table-column>
        <el-table-column prop="ssh_name" label="SSH隧道" width="140">
          <template #default="scope">
            <el-tag v-if="scope.row.ssh_id && scope.row.ssh_id !== '0'" size="small" type="success">
              {{ scope.row.ssh_name || '已配置' }}
            </el-tag>
            <el-tag v-else size="small" type="info">直连</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="host" label="Host" min-width="180">
          <template #default="scope">
            <code class="host-code">{{ scope.row.host }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="port" label="Port" width="100" align="center">
          <template #default="scope">
            <el-tag size="small" type="info">{{ scope.row.port }}</el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="status" label="状态" width="100" align="center">
          <template #default="scope">
            <el-tag :type="scope.row.status === 'connected' ? 'success' : 'danger'" size="small">
              {{ scope.row.status === 'connected' ? '已连接' : '未连接' }}
            </el-tag>
          </template>
        </el-table-column>
        <el-table-column prop="username" label="用户名" min-width="150">
          <template #default="scope">
            <span class="username-text">{{ scope.row.username || '-' }}</span>
          </template>
        </el-table-column>
        <el-table-column label="操作" width="220" fixed="right">
          <template #default="scope">
            <div class="action-buttons">
              <pl-button type="primary" link size="small" @click="ShowEditRedis(scope.row , true)">
                <el-icon><CopyDocument /></el-icon>复制新增
              </pl-button>
              <pl-button type="primary" link size="small" @click="ShowEditRedis(scope.row , false)">
                <el-icon><Edit /></el-icon>编辑
              </pl-button>
              <pl-button type="danger" link size="small" @click="DeleteRedis(scope.row)">
                <el-icon><Delete /></el-icon>删除
              </pl-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

  <!-- 编辑弹窗 -->
  <el-dialog v-model="state.dialogEditRedis" title="编辑Redis配置" width="500px" class="edit-dialog">
    <el-form :model="state.starForm" label-width="80px" class="edit-form">
      <el-form-item label="名称">
        <el-input v-model="state.editRedisConfig.name" autocomplete="off" placeholder="为Redis实例命名">
          <template #prefix>
            <el-icon><CollectionTag /></el-icon>
          </template>
        </el-input>
      </el-form-item>
      <el-form-item label="Host">
        <el-input v-model="state.editRedisConfig.host" autocomplete="off" placeholder="Redis服务器地址">
          <template #prefix>
            <el-icon><Link /></el-icon>
          </template>
        </el-input>
      </el-form-item>
      <el-form-item label="Port">
        <el-input v-model="state.editRedisConfig.port" autocomplete="off" placeholder="默认6379">
          <template #prefix>
            <el-icon><Connection /></el-icon>
          </template>
        </el-input>
      </el-form-item>
      <el-form-item label="用户名">
        <el-input v-model="state.editRedisConfig.username" autocomplete="off" placeholder="可选">
          <template #prefix>
            <el-icon><User /></el-icon>
          </template>
        </el-input>
      </el-form-item>
      <el-form-item label="密码">
        <el-input v-model="state.editRedisConfig.password" type="password" autocomplete="off" placeholder="Redis密码" show-password>
          <template #prefix>
            <el-icon><Lock /></el-icon>
          </template>
        </el-input>
      </el-form-item>
      <el-form-item label="SSH隧道">
        <el-select v-model="state.editRedisConfig.ssh_id" placeholder="选择SSH隧道" style="width: 100%">
          <el-option v-for="item in state.sshList" :key="item.id" :label="item.name" :value="item.id"/>
        </el-select>
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="state.dialogEditRedis = false">取消</pl-button>
        <pl-button type="primary" @click="EditRedis">
          <el-icon><Check /></el-icon>保存
        </pl-button>
      </div>
    </template>
  </el-dialog>
  </div>
</template>
<script>
import {defineComponent, getCurrentInstance, reactive, onActivated} from 'vue';
import { Plus, CopyDocument, Edit, Delete, CollectionTag, Link, Connection, User, Lock, Check } from '@element-plus/icons-vue';
import set from '../../utils/base/redis_set'
import common from '../../utils/common'
import ssh_set from "@/utils/base/ssh_set";
import Init from "@/utils/base/set_init";
export default defineComponent({
  components: { Plus, CopyDocument, Edit, Delete, CollectionTag, Link, Connection, User, Lock, Check },
  props: {
  },
  emits: ['changed'],
  data() {
    return {
    }
  },
  setup(props, { emit }) {
    onActivated(() => {
      if(Init.GetIsInit('redis') === true){
        RedisList()
        SshList()
        Init.DelInit('redis')
      }
    });
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties
    const RedisList = function (){
      set.RedisList(function (response){
        if(response.ErrCode === 0){
          state.redisList = response.Data
        }
      })
    }
    const ShowEditRedis = function (redisConfig , isCopy){
      state.dialogEditRedis = true
      state.editRedisConfig = redisConfig
      if(isCopy){
        state.editRedisConfig.id = 0
      }
    }
    const ShowAddRedis = function (){
      state.dialogEditRedis = true
      state.editRedisConfig = {}
    }
    // emitChanged 告知宿主页面 Redis 配置已变化，便于立即重载实例列表。
    // Notify host pages when Redis settings changed so instance lists can reload right away.
    const emitChanged = function (){
      emit('changed')
    }
    const EditRedis = function (){
      set.RedisAdd(state.editRedisConfig , function (response){
        if(response.ErrCode === 0){
          RedisList()
          emitChanged()
        }else{
          instance.$helperNotify.success(response.ErrMsg)
        }
        state.dialogEditRedis = false
        SetInit()
      })
    }
    const DeleteRedis = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.RedisDelete(rowData , function (response){
          if(response.ErrCode === 0){
            RedisList()
            emitChanged()
          }else{
            instance.$helperNotify.success(response.ErrMsg)
          }
          SetInit()
        })
      })
    }
    const SshList = function (){
      ssh_set.SshList(function (response){
        if(response.ErrCode === 0){
          state.sshList = response.Data
          state.sshList.unshift({id : "0" , name : '请选择'})
        }
      })
    }
    const SetInit = function(){
      Init.SetIsInit('redis')
    }
    SshList()
    //固有属性
    const state = reactive({
      sshList : [],
      redisList : [],
      dialogEditRedis : false,
      editRedisConfig : {},
    })
    //初始化
    RedisList()
    return {
      state,
      ShowEditRedis,
      ShowAddRedis,
      EditRedis,
      DeleteRedis,
      RedisList,
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

.redis-config-page {
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

.config-table :deep(.el-table__header-wrapper),
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

.name-text {
  font-weight: 500;
  color: #303133;
}

.host-code {
  font-family: Consolas, Monaco, monospace;
  font-size: 13px;
  color: #4f804f;
  background: #f3f8ef;
  padding: 2px 8px;
  border-radius: 4px;
}

.username-text {
  font-family: Consolas, Monaco, monospace;
  font-size: 13px;
  color: #606266;
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

