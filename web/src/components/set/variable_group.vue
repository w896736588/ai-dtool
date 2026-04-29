<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">脚本合集组管理</h3>
      <div class="set-config-actions">
        <pl-button type="primary" @click="ShowAddVariableGroup">添加分组</pl-button>
      </div>
    </div>
    <div class="set-config-table-card">
      <el-table :data="state.VariableGroupList" class="set-config-table">
        <el-table-column prop="id" label="#id" width="80" />
        <el-table-column prop="name" label="组名" min-width="180" />
        <el-table-column label="操作" width="150">
          <template #default="scope">
            <div class="set-op-group">
              <pl-button type="primary" link @click="ShowEditVariableGroup(scope.row)">编辑</pl-button>
              <pl-button link type="danger" @click="DeleteVariableGroup(scope.row)">删除</pl-button>
            </div>
          </template>
        </el-table-column>
      </el-table>
    </div>
    <el-dialog v-model="state.dialogEditVariableGroup" title="编辑分组" width="500">
      <el-form label-width="80px">
        <el-form-item label="组名">
          <el-input v-model="state.editVariableGroupConfig.name" autocomplete="off" />
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogEditVariableGroup = false">取消</pl-button>
          <pl-button type="primary" @click="EditVariableGroup">保存</pl-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>
<script>
import {defineExpose , defineComponent , inject , defineEmits , getCurrentInstance , reactive} from 'vue';
import set from '../../utils/base/variable_set'
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
    const VariableGroupList = function (){
      set.SetVariableGroupList(function (response){
        if(response.ErrCode === 0){
          state.VariableGroupList = response.Data
        }
      })
    }
    const ShowEditVariableGroup = function (VariableGroupConfig){
      state.dialogEditVariableGroup = true
      state.editVariableGroupConfig = VariableGroupConfig
    }
    const ShowAddVariableGroup = function (){
      state.dialogEditVariableGroup = true
      state.editVariableGroupConfig = {}
    }
    const EditVariableGroup = function (){
      set.SetVariableGroupAdd(state.editVariableGroupConfig , function (response){
        if(response.ErrCode === 0){
          VariableGroupList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
        state.dialogEditVariableGroup = false
      })
    }
    const DeleteVariableGroup = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        set.SetVariableGroupDelete(rowData , function (response){
          if(response.ErrCode === 0){
            VariableGroupList()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
        })
      })
    }
    //固有属性
    const state = reactive({
      sshList :[],
      VariableGroupList : [],
      dialogEditVariableGroup : false,
      editVariableGroupConfig : {},
      quickFilterKeysResult : [],
      dialogEditVariableQuick : false,
      loading : {
        quick : false
      }
    })
    //初始化
    VariableGroupList()
    return {
      state,
      ShowEditVariableGroup,
      ShowAddVariableGroup,
      EditVariableGroup,
      DeleteVariableGroup,
      VariableGroupList,
    }
  },
  mounted() {

  },
  methods: {
  },
})
</script>

<style scoped src="@/css/components/set/variable_group.css"></style>

