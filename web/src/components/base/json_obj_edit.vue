<template>
    <el-form :model="state.form" label-width="auto" style="max-width: 600px">
      <!--      这里为示例对象 里面的元素一定是对象-->
      <template v-for="[tKey , column] in Object.entries(state.templateObj)" :key="key">
        <el-form-item :label="tKey">
          <template v-if="!IsArray(state.form[tKey]) && !IsObject(state.form[tKey])">
            <el-input v-model="state.form[tKey]"></el-input>
          </template>
          <template v-if="IsArray(state.form[tKey])">
            <el-button size="small" type="primary" @click="AddColumn(tKey)">新增</el-button>
            <el-table :data="state.form[tKey]" style="width: 100%;font-size:13px;">
              <template v-for="[key, column] in Object.entries(state.templateObj[tKey][0])">
                <el-table-column :label="key" class-name="ellipsis-column">
                  <template #default="scope">
                    <template v-if="!IsArray(scope.row[key]) && !IsObject(scope.row[key])">
                      <el-input v-model="scope.row[key]" @input="Update">{{scope.row[key]}}</el-input>
                    </template>
                    <template v-if="IsObject(scope.row[key])">
                      <el-tooltip :content="JSON.stringify(scope.row[key])" placement="top-end">
                        <span @click="Copy(JSON.stringify(scope.row[key]))">Object：{{ JSON.stringify(scope.row[key]) }}</span>
                      </el-tooltip>
                    </template>
                    <template v-if="IsArray(scope.row[key])">
                      <el-tooltip :content="JSON.stringify(scope.row[key])" placement="top-end">
                        <span @click="Copy(JSON.stringify(scope.row[key]))">Array：{{ scope.row[key].length }}</span>
                      </el-tooltip>
                    </template>
                  </template>
                </el-table-column>
              </template>

              <el-table-column fixed="right" label="操作" min-width="70">
                <template #default="scope">
                  <el-popconfirm
                      cancel-button-text="取消"
                      confirm-button-text="删除"
                      icon-color="#626AEF"
                      title="确定删除吗?"
                      @confirm="DeleteColumn(tKey , scope.$index)"
                  >
                    <template #reference>
                      <el-button link size="small" type="danger" >
                        删除
                      </el-button>
                    </template>
                  </el-popconfirm>

                  <el-button link size="small" type="primary" @click="EditColumn(tKey,scope.$index)">编辑
                  </el-button>
                  <el-button link size="small" type="primary" @click="UpMove(tKey,scope.$index)">
                    <el-icon><ArrowUp /></el-icon>
                  </el-button>
                </template>
              </el-table-column>
            </el-table>
          </template>
        </el-form-item>
      </template>

      <el-form-item>
        <el-button type="primary" @click="Save">保存</el-button>
        <el-button @click="Cancel">取消</el-button>
      </el-form-item>
    </el-form>

</template>
<style>
.el-form-item {
  padding: 5px;
}

.ellipsis-column .cell {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}
</style>

<script>
import {reactive} from "vue";
import copy from "@/utils/base/copy"
import obj from "@/utils/base/obj"
import {computed, watch} from 'vue';
import json from '@/utils/base/json'
import t from '@/utils/base/type'

export default {
  name: 'json_edit',
  props: {
    templateObj: {
      type: [Object],
      default: '',
    },
    form: {
      type: [Object],
      default: ''
    },
  },
  data() {
    return {}
  },
  setup(props, {emit}) {
    //表格增加一行
    const AddColumn = function (tKey) {
      state.form[tKey].push(state.templateObj[tKey][0])
      // EditColumn(tKey , 0)
    }
    //编辑其中的表格 通知父组件
    const EditColumn = function (tKey, index) {
      emit('formEdit', tKey , index)
    }
    const UpMove = function (tKey , index){
      if(parseInt(index) === 0){
        return
      }
      let temp = state.form[tKey][index-1]
      state.form[tKey][index-1] = state.form[tKey][index]
      state.form[tKey][index] = temp
    }
    const DeleteColumn = function (tKey , index){
      state.form[tKey].splice(index , 1)
    }
    const Copy = function (content) {
      let index = copy.SetCopyContent(content)
      copy.handleCopy(index)
    }
    //取消时还原
    const Cancel = function (){
      obj.Copy(state.form , JSON.parse(state.formJson))
      emit('formSave')
    }
    //保存 不需要处理 js 对象和数组是引用传值
    const Save = function (){
      emit('formSave')
    }
    //仅更新
    const Update = function (){
      emit('formUpdate')
    }
    const state = reactive({
      templateObj: {}, //模板
      formJson : '',
      form: {}, //数据
    })

    //初始化
    function init() {
      state.templateObj = props.templateObj
      state.formJson = JSON.stringify(props.form) //用于取消
      state.form = props.form
    }

    function IsString(value) {
      return t.IsString(value)
    }

    function IsArray(value) {
      return t.IsArray(value)
    }

    function IsObject(value) {
      return t.IsObject(value)
    }

    init();
    return {
      state,
      AddColumn,
      EditColumn,
      UpMove,
      IsString,
      IsArray,
      IsObject,
      Copy,
      Cancel,
      Save,
      Update,
      DeleteColumn,
    }
  },
}
</script>

<el-button>
<slot/>
</el-button>
