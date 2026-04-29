<template>
  配置{{groupTitle}}组 <pl-button type="primary" link @click="ShowAddGroup">添加</pl-button>
  <el-table :data="state.groupList" style="width: 100%">
    <el-table-column prop="id" label="#id" />
    <el-table-column prop="name" label="name" />
    <el-table-column prop="extra_1" v-if="extra1Type !== ''" :type="extra1Type" :label="extra1Title" show-overflow-tooltip min-width="100">
      <template #default="scope">
        <div class="ellipsis-cell">
          {{ scope.row.extra_1 }}
        </div>
      </template>
    </el-table-column>
    <el-table-column prop="extra_2" v-if="extra2Type !== ''" :type="extra2Type" :label="extra2Title" show-overflow-tooltip>
      <template #default="scope">
        <div class="ellipsis-cell">
          {{ scope.row.extra_2 }}
        </div>
      </template>
    </el-table-column>
    <el-table-column prop="extra_3" v-if="extra3Type !== ''" :type="extra3Type" :label="extra3Title" show-overflow-tooltip>
      <template #default="scope">
        <div class="ellipsis-cell">
          {{ scope.row.extra_3 }}
        </div>
      </template>
    </el-table-column>
    <el-table-column prop="extra_4" v-if="extra4Type !== ''" :type="extra4Type" :label="extra4Title" show-overflow-tooltip>
      <template #default="scope">
        <div class="ellipsis-cell">
          {{ scope.row.extra_4 }}
        </div>
      </template>
    </el-table-column>
    <el-table-column prop="extra_5" v-if="extra5Type !== ''" :type="extra5Type" :label="extra5Title" show-overflow-tooltip>
      <template #default="scope">
        <div class="ellipsis-cell">
          {{ scope.row.extra_5 }}
        </div>
      </template>
    </el-table-column>
    <el-table-column prop="extra_6" v-if="extra6Type !== ''" :type="extra6Type" :label="extra6Title" show-overflow-tooltip>
      <template #default="scope">
        <div class="ellipsis-cell">
          {{ scope.row.extra_6 }}
        </div>
      </template>
    </el-table-column>
    <el-table-column label="操作" >
      <template #default="scope">
        <pl-button type="primary" link @click="ShowEditGroup(scope.row)">编辑</pl-button>
        <pl-button link type="danger" @click="DeleteGroup(scope.row)">删除</pl-button>
      </template>
    </el-table-column>
  </el-table>
  <p></p>
  <el-dialog v-model="state.dialogEditGroup" title="编辑" width="70%">
    <el-form>
      <el-form-item label="组名" :label-width="80">
        <el-input v-model="state.editGroupConfig.name" autocomplete="off" />
      </el-form-item>
      <el-form-item v-if="extra1Type !== ''" :label="extra1Title" :label-width="80">
        <el-input v-model="state.editGroupConfig.extra_1" :type="extra1Type" rows="5" autocomplete="off" />
      </el-form-item>
      <el-form-item v-if="extra2Type !== ''" :label="extra2Title" :label-width="80">
        <el-input v-if="extra2Type !== ''"  v-model="state.editGroupConfig.extra_2" :type="extra2Type" rows="5" autocomplete="off" />
      </el-form-item>
      <el-form-item v-if="extra3Type !== ''" :label="extra3Title" :label-width="80">
        <el-input v-if="extra3Type !== ''"  v-model="state.editGroupConfig.extra_3" :type="extra3Type" rows="5" autocomplete="off" />
      </el-form-item>
      <el-form-item v-if="extra4Type !== ''" :label="extra4Title" :label-width="80">
        <el-input v-if="extra4Type !== ''"  v-model="state.editGroupConfig.extra_4" :type="extra4Type" rows="5" autocomplete="off" />
      </el-form-item>
      <el-form-item v-if="extra5Type !== ''" :label="extra5Title" :label-width="80">
        <el-input v-if="extra5Type !== ''"  v-model="state.editGroupConfig.extra_5" :type="extra5Type" rows="5" autocomplete="off" />
      </el-form-item>
      <el-form-item v-if="extra6Type !== ''" :label="extra6Title" :label-width="80">
        <el-input v-if="extra6Type !== ''" v-model="state.editGroupConfig.extra_6" :type="extra6Type" rows="5" autocomplete="off" />
      </el-form-item>
    </el-form>
    <template #footer>
      <div class="dialog-footer">
        <pl-button @click="state.dialogEditGroup = false">取消</pl-button>
        <pl-button type="primary" @click="EditGroup">
          保存
        </pl-button>
      </div>
    </template>
  </el-dialog>

</template>
<script>
import {defineExpose , defineComponent , inject , defineEmits , getCurrentInstance , reactive} from 'vue';
import group from '@/utils/base/group'
import common from '../../utils/common'
export default defineComponent({
  props: {
    groupType: {
      type: String,
      default: 0,
    },
    groupTitle: {
      type: String,
      default: '',
    },
    extra1Type : {
      type: String,
      default: '',
    },
    extra1Title : {
      type: String,
      default: '',
    },
    extra2Type : {
      type: String,
      default: '',
    },
    extra2Title : {
      type: String,
      default: '',
    },
    extra3Type : {
      type: String,
      default: '',
    },
    extra3Title : {
      type: String,
      default: '',
    },
    extra4Type : {
      type: String,
      default: '',
    },
    extra4Title : {
      type: String,
      default: '',
    },
    extra5Type : {
      type: String,
      default: '',
    },
    extra5Title : {
      type: String,
      default: '',
    },
    extra6Type : {
      type: String,
      default: '',
    },
    extra6Title : {
      type: String,
      default: '',
    },
  },
  emits: ['update'],
  data() {
    return {
    }
  },
  setup(props , {emit}) {
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties

    const GroupList = function (){
      group.GroupList({type : props.groupType} , function (response){
        if(response.ErrCode === 0){
          state.groupList = response.Data
        }
      })
    }
    const ShowEditGroup = function (groupConfig){
      state.dialogEditGroup = true
      state.editGroupConfig = groupConfig
    }
    const ShowAddGroup = function (){
      state.dialogEditGroup = true
      state.editGroupConfig = {}
    }
    const EditGroup = function (){
      state.editGroupConfig.type = props.groupType
      group.GroupAdd(state.editGroupConfig , function (response){
        if(response.ErrCode === 0){
          GroupList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
        state.dialogEditGroup = false
        emit('update')
      })
    }
    const DeleteGroup = function (rowData){
      common.ConfirmProxyDelete(proxy , function () {
        group.GroupDelete(rowData , function (response){
          if(response.ErrCode === 0){
            GroupList()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
          emit('update')
        })
      })
    }
    //固有属性
    const state = reactive({
      groupList : [],
      dialogEditGroup : false,
      editGroupConfig : {},
    })
    //初始化
    GroupList()
    return {
      state,
      ShowEditGroup,
      ShowAddGroup,
      EditGroup,
      DeleteGroup,
      GroupList,
    }
  },
  mounted() {

  },
  methods: {
  },
})
</script>

<style scoped src="@/css/components/group/group_list.css"></style>
