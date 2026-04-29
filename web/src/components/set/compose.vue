<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">Docker Compose 配置</h3>
      <p class="set-config-desc">管理 compose 项目目录、默认服务与执行命令</p>
      <div class="set-config-actions">
        <pl-button type="primary" @click="ShowAddCompose">添加 Compose</pl-button>
        <el-input
          v-model="state.searchKey"
          autocomplete="off"
          placeholder="搜索名称、目录、SSH等"
          class="set-config-search"
          clearable
        ></el-input>
      </div>
    </div>
    <div class="set-config-table-card">
      <el-table :data="filteredComposeList" class="set-config-table" height="calc(100vh - 300px)">
        <el-table-column prop="id" label="#id" width="60"/>
        <el-table-column prop="name" label="名称" min-width="150"/>
        <el-table-column prop="compose_yml_path" label="compose.yml目录" min-width="220">
          <template #default="scope">
            <code class="set-mono">{{ scope.row.compose_yml_path }}</code>
          </template>
        </el-table-column>
        <el-table-column prop="env_file" label="env file" min-width="180" />
        <el-table-column prop="ssh_name" label="SSH" width="140"/>
        <el-table-column prop="docker_cmd" label="命令" min-width="120"/>
        <el-table-column prop="default_service" label="默认服务" min-width="180" />
        <el-table-column label="操作" width="200">
          <template #default="scope">
            <div class="set-op-group">
              <pl-button type="primary" link @click="ShowEditCompose(scope.row , true)">复制新增</pl-button>
              <pl-button type="primary" link @click="ShowEditCompose(scope.row , false)">编辑</pl-button>
              <pl-button link type="danger" @click="DeleteCompose(scope.row)">删除</pl-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>

    <el-dialog v-model="state.dialogEditCompose" title="编辑Compose配置" width="70%">
      <el-form :model="state.starForm" label-width="180px">
        <el-form-item label="名称">
          <el-input v-model="state.editComposeConfig.name" autocomplete="off" />
        </el-form-item>
        <el-form-item label="compose.yml">
          <el-input v-model="state.editComposeConfig.compose_yml_path" autocomplete="off" />
        </el-form-item>
        <el-form-item label="env file(为空默认为.env)">
          <el-input v-model="state.editComposeConfig.env_file" autocomplete="off" />
        </el-form-item>
        <el-form-item label="SSH">
          <el-select v-model="state.editComposeConfig.ssh_id" placeholder="选择SSH" style="width: 140px">
            <el-option v-for="item in state.sshList" :key="item.id" :label="item.name" :value="item.id"/>
          </el-select>
        </el-form-item>
        <el-form-item label="docker命令">
          <el-input v-model="state.editComposeConfig.docker_cmd" autocomplete="off" />
        </el-form-item>
        <el-form-item label="默认服务(多个英文逗号分割)">
          <el-input v-model="state.editComposeConfig.default_service" autocomplete="off" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogEditCompose = false">取消</pl-button>
          <pl-button type="primary" @click="EditCompose">保存</pl-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>
<script>
import {defineComponent , getCurrentInstance , reactive, computed} from 'vue';
import set from '../../utils/base/compose_set'
import common from '../../utils/common'
import ssh_set from "@/utils/base/ssh_set";
export default defineComponent({
  props: {
  },
  emits: ['changed'],
  data() {
    return {
    }
  },
  setup(props, { emit }) {
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties
    const ComposeList = function (){
      set.ComposeList(function (response){
        if(response.ErrCode === 0){
          state.composeList = response.Data
        }
      })
    }
    const ShowEditCompose = function (composeConfig , isCopy){
      state.dialogEditCompose = true
      state.editComposeConfig = composeConfig
      if(isCopy){
        state.editComposeConfig.id = 0
      }
    }
    const ShowAddCompose = function (){
      state.dialogEditCompose = true
      state.editComposeConfig = {}
    }
    // emitChanged 告知宿主页面 Compose 配置已变化，便于刷新项目列表。
    // Notify host pages when compose settings changed so project lists can reload right away.
    const emitChanged = function (){
      emit('changed')
    }
    const EditCompose = function (){
      set.ComposeAdd(state.editComposeConfig , function (response){
        if(response.ErrCode === 0){
          ComposeList()
          emitChanged()
        }else{
          instance.$helperNotify.success(response.ErrMsg)
        }
        state.dialogEditCompose = false
      })
    }
    const DeleteCompose = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.ComposeDelete(rowData , function (response){
          if(response.ErrCode === 0){
            ComposeList()
            emitChanged()
          }else{
            instance.$helperNotify.success(response.ErrMsg)
          }
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
      sshList : [],
      composeList : [],
      dialogEditCompose : false,
      editComposeConfig : {},
      searchKey : '',
    })

    const filteredComposeList = computed(() => {
      const key = (state.searchKey || '').trim().toLowerCase()
      if (!key) return state.composeList
      const keywords = key.split(/\s+/)
      return state.composeList.filter(row => {
        const text = [row.name, row.compose_yml_path, row.env_file, row.ssh_name, row.docker_cmd, row.default_service].filter(Boolean).join(' ').toLowerCase()
        return keywords.every(k => text.includes(k))
      })
    })
    //初始化
    ComposeList()
    SshList()

    return {
      state,
      filteredComposeList,
      ShowEditCompose,
      ShowAddCompose,
      EditCompose,
      DeleteCompose,
      ComposeList,
      SshList,
    }
  },
  mounted() {

  },
  methods: {
  },
})
</script>

<style scoped src="@/css/components/set/compose.css"></style>

