<template>
  配置脚本合集组 <el-button type="primary" link @click="ShowAddVariableGroup">添加</el-button>
  <el-table :data="state.VariableGroupList" style="width: 100%">
    <el-table-column prop="id" label="#id" />
    <el-table-column prop="name" label="name" />
    <el-table-column label="操作" >
      <template #default="scope">
        <el-button type="primary" link @click="ShowEditVariableGroup(scope.row)">编辑</el-button>
        <el-button link type="danger" @click="DeleteVariableGroup(scope.row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>
  <p></p>
  <el-dialog v-model="state.dialogEditVariableGroup" title="编辑" width="500">
    <el-form>
      <el-form-item label="组名" :label-width="80">
        <el-input v-model="state.editVariableGroupConfig.name" autocomplete="off" />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogEditVariableGroup = false">取消</el-button>
        <el-button type="primary" @click="EditVariableGroup">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>

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

<style scoped>

</style>