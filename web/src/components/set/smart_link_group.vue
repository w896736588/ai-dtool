<template>
  配置自动化链接组 <el-button type="primary" link @click="ShowAddSmartLinkGroup">添加</el-button>
  <el-table :data="state.SmartLinkGroupList" style="width: 100%">
    <el-table-column prop="id" label="#id" />
    <el-table-column prop="name" label="name" />
    <el-table-column label="操作" >
      <template #default="scope">
        <el-button type="primary" link @click="ShowEditSmartLinkGroup(scope.row)">编辑</el-button>
        <el-button link type="danger" @click="DeleteSmartLinkGroup(scope.row)">删除</el-button>
      </template>
    </el-table-column>
  </el-table>
  <p></p>
  <el-dialog v-model="state.dialogEditSmartLinkGroup" title="编辑" width="500">
    <el-form>
      <el-form-item label="组名" :label-width="80">
        <el-input v-model="state.editSmartLinkGroupConfig.name" autocomplete="off" />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <el-button @click="state.dialogEditSmartLinkGroup = false">取消</el-button>
        <el-button type="primary" @click="EditSmartLinkGroup">
          保存
        </el-button>
      </div>
    </template>
  </el-dialog>

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

<style scoped>

</style>