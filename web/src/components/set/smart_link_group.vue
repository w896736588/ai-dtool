<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">自动化链接组管理</h3>
      <div class="set-config-actions">
        <pl-button type="primary" @click="ShowAddSmartLinkGroup">添加分组</pl-button>
      </div>
    </div>
    <div class="set-config-table-card">
      <el-table :data="state.SmartLinkGroupList" class="set-config-table">
        <el-table-column prop="id" label="#id" width="80" />
        <el-table-column prop="name" label="组名" min-width="180" />
        <el-table-column label="操作" width="150">
          <template #default="scope">
            <div class="set-op-group">
              <pl-button type="primary" link @click="ShowEditSmartLinkGroup(scope.row)">编辑</pl-button>
              <pl-button link type="danger" @click="DeleteSmartLinkGroup(scope.row)">删除</pl-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog v-model="state.dialogEditSmartLinkGroup" title="编辑分组" width="500">
      <el-form label-width="80px">
        <el-form-item label="组名">
          <el-input v-model="state.editSmartLinkGroupConfig.name" autocomplete="off" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogEditSmartLinkGroup = false">取消</pl-button>
          <pl-button type="primary" @click="EditSmartLinkGroup">保存</pl-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>
<script>
import {defineExpose , defineComponent , inject , defineEmits , getCurrentInstance , reactive} from 'vue';
import set from '../../utils/base/smart_link_set'
import common from '../../utils/common'
export default defineComponent({
  props: {
  },
  data() {
    return {
    }
  },
  setup() {
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties
    const SmartLinkGroupList = function (){
      set.SetSmartLinkGroupList(function (response){
        if(response.ErrCode === 0){
          state.SmartLinkGroupList = response.Data
        }
      })
    }
    const ShowEditSmartLinkGroup = function (SmartLinkGroupConfig){
      state.dialogEditSmartLinkGroup = true
      state.editSmartLinkGroupConfig = SmartLinkGroupConfig
    }
    const ShowAddSmartLinkGroup = function (){
      state.dialogEditSmartLinkGroup = true
      state.editSmartLinkGroupConfig = {}
    }
    const EditSmartLinkGroup = function (){
      set.SetSmartLinkGroupAdd(state.editSmartLinkGroupConfig , function (response){
        if(response.ErrCode === 0){
          SmartLinkGroupList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
        state.dialogEditSmartLinkGroup = false
      })
    }
    const DeleteSmartLinkGroup = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.SetSmartLinkGroupDelete(rowData , function (response){
          if(response.ErrCode === 0){
            SmartLinkGroupList()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
        })
      })
    }
    //固有属性
    const state = reactive({
      sshList :[],
      SmartLinkGroupList : [],
      dialogEditSmartLinkGroup : false,
      editSmartLinkGroupConfig : {},
      quickFilterKeysResult : [],
      dialogEditSmartLinkQuick : false,
      loading : {
        quick : false
      }
    })
    //初始化
    SmartLinkGroupList()
    return {
      state,
      ShowEditSmartLinkGroup,
      ShowAddSmartLinkGroup,
      EditSmartLinkGroup,
      DeleteSmartLinkGroup,
      SmartLinkGroupList,
    }
  },
  mounted() {

  },
  methods: {
  },
})
</script>

<style scoped src="@/css/components/set/smart_link_group.css"></style>

