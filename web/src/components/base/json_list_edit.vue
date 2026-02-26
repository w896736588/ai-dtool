<template>
  <!--  数组-->
  <template v-if="showSource" style="height:50px;">
    模板：<span v-if="state.templateErr !== ''" style="color: red;">
          {{ state.templateErr }}
          </span>
    <span v-else style="color:green">正确</span>&nbsp;
    <el-button size="small" type="primary" @click="WitchTemplate">解析JSON</el-button>
    <el-input v-model="state.template" rows="2" style="margin-bottom:0px;" type="textarea"></el-input>
  </template>
  <template v-if="IsArray(state.configData) && IsObject(state.templateObj)">
    <el-row style="width: 100%;margin-top:3px;">
      <el-col :span="21">
        <el-input v-model="state.tableSearch" placeholder="搜索，多条件空格分割" type="text" @input="tableFilter"></el-input>
      </el-col>
      <el-col :span="3" style="text-align: center;">
        <el-button size="small" type="primary" @click="AddColumn">新增</el-button>
      </el-col>
    </el-row>
    <table class="json_list_edit" style="width: 100%;margin-top: 3px;">
      <thead>
      <tr>
        <template v-for="[tKey, column] in Object.entries(state.templateObj)" :key="tKey">
          <th>
          <span v-if="tKey.length > 15">
            <el-tooltip
                :content="tKey"
                effect="dark"
                placement="top-start">
              {{ tKey.substring(0, 15) + '..' }}
            </el-tooltip>
          </span>
            <span v-else>{{ tKey }}</span>
          </th>

        </template>
        <td>操作</td>
      </tr>

      <template v-for="(row, index) in state.configData" :key="index">
        <tr v-if="state.tableIndexList.indexOf(index) !== -1">
          <template v-for="[tKey, column] in Object.entries(state.templateObj)" :key="tKey">
            <td>
              <template v-if="!IsArray(row[tKey]) && !IsObject(row[tKey])">
                <el-input v-model="row[tKey]" @input="SourceDataUpdate">{{ row[tKey] }}</el-input>
              </template>
              <template v-if="IsObject(row[tKey])">
                {{ JSON.stringify(row[tKey]) }}
              </template>
              <template v-if="IsArray(row[tKey])">
                元素：{{ row[tKey].length }}
              </template>
            </td>
          </template>
          <td style="min-width:120px;">
            <el-popconfirm
                cancel-button-text="取消"
                confirm-button-text="删除"
                icon-color="#626AEF"
                title="确定删除吗?"
                @confirm="DeleteColumn(index)"
            >
              <template #reference>
                <el-button link size="small" type="danger">
                  删除
                </el-button>
              </template>
            </el-popconfirm>

            <el-button link size="small" type="primary" @click="EditForm(index,row)">编辑</el-button>
            <el-button link size="small" type="primary" @click="UpMove(index)">
              <el-icon>
                <ArrowUp/>
              </el-icon>
            </el-button>
          </td>
        </tr>
      </template>
      </thead>
    </table>
<!--    <el-pagination :current-page="state.tablePage" :page-size="state.tableSize" :total="state.configData.length" background layout="prev, pager, next" size="small" @current-change="tableChangePage"/>-->
  </template>

  <!--  对象-->
  <el-dialog v-model="state.dialogFormVisible1" title="编辑" width="800">
    <json_obj_edit v-if="state.dialogFormVisible1" :form="state.form1" :templateObj="state.templateObjForm1" @formEdit="form1Edit" @formSave="form1Save" @formUpdate="form1Update"></json_obj_edit>
  </el-dialog>

  <el-dialog v-model="state.dialogFormVisible2" title="编辑" width="800">
    <json_obj_edit v-if="state.dialogFormVisible2" :form="state.form2" :templateObj="state.templateObjForm2" @formEdit="form2Edit" @formSave="form2Save" @formUpdate="form2Update"></json_obj_edit>
  </el-dialog>

  <el-dialog v-model="state.dialogFormVisible3" title="编辑" width="800">
    <json_obj_edit v-if="state.dialogFormVisible3" :form="state.form3" :templateObj="state.templateObjForm3" @formEdit="form3Edit" @formSave="form3Save" @formUpdate="form3Update"></json_obj_edit>
  </el-dialog>

</template>
<style>
td {
  padding: 5px;
}

th {
  text-align: left;
}

.el-form-item {
  padding: 5px;
}

.ellipsis-column .cell {
  white-space: nowrap;
  overflow: hidden;
  text-overflow: ellipsis;
}

.json_list_edit {
  overflow-y: auto !important; /* 垂直滚动条 */
  border: 1px solid #dcdfe6;
  border-radius: 4px;
  padding: 10px;
}
</style>

<script>
import {reactive, ref} from "vue";
import array from "@/utils/base/array"
import copy from "@/utils/base/copy"
import {computed, watch} from 'vue';
import json from '@/utils/base/json'
import obj from '@/utils/base/obj'
import t from '@/utils/base/type'
import explain_json from '@/utils/base/explain_json'
import json_obj_edit from '@/components/base/json_obj_edit.vue'

export default {
  name: 'json_list_edit',
  props: {
    data: {
      type: [String, Object, Array],
      default: ''
    },
    showSource: {
      type: [Boolean],
      default: true
    },
  },
  data() {
    return {}
  },
  components: {
    json_obj_edit,
  },
  setup(props, {emit}) {
    //解析sourceData
    const explainSourceData = function () {
      let explainResult = explain_json.Explain({template: state.template, config: state.config})
      state.configData = explainResult.returnConfigData
      if (t.IsArray(state.configData)) {
        obj.ToEmpty(explainResult.returnTemplateObj)
        state.template = JSON.stringify(explainResult.returnTemplateObj)
        state.templateObj = explainResult.returnTemplateObj[0]
        state.templateObjForm1 = state.templateObj
        tableChangePage(1)
      } else if (t.IsObject(state.configData)) {
        obj.ToEmpty(explainResult.returnTemplateObj)
        state.template = JSON.stringify(explainResult.returnTemplateObj)
        state.templateObj = explainResult.returnTemplateObj
        state.templateObjForm1 = state.templateObj
      }
    }
    const tableFilter = function (search) {
      let searchParams = search.split(' ')
      searchParams = array.FilterEmpty(searchParams)
      state.tableIndexList = []
      for (let i = 0; i < state.configData.length; i++) {
        let keys = Object.keys(state.configData[i])
        for (let key of keys) {
          if (t.IsString(state.configData[i][key])) {
            let findNum = 0
            for (let searchParam of searchParams) {
              if (state.configData[i][key].indexOf(searchParam) !== -1) {
                findNum++
              }
            }
            // 全部找到
            if (findNum === searchParams.length) {
              state.tableIndexList.push(i)
            }
          }
        }
      }
    }
    const tableChangePage = function (page) {
      state.tableIndexList = []
      for (let i = 0; i < state.configData.length; i++) {
        if (i >= (page - 1) * state.tableSize && i < page * state.tableSize) {
          state.tableIndexList.push(i)
        }
      }
      state.tablePage = page
    }
    //第0次编辑
    const EditForm = function (index, data) {
      state.form1 = data
      state.table1Index = index
      state.dialogFormVisible1 = true
    }
    const UpMove = function (index) {
      if (parseInt(index) === 0) {
        return
      }
      let temp = state.configData[index - 1]
      state.configData[index - 1] = state.configData[index]
      state.configData[index] = temp
      SourceDataUpdate()
    }
    //第1次编辑
    const form1Edit = function (tKey, index) {
      state.templateObjForm2 = state.templateObjForm1[tKey][0]
      state.table2Key = tKey
      state.table2Index = index
      state.form2 = state.form1[tKey][index]
      state.dialogFormVisible2 = true
    }
    const form2Edit = function (tKey, index) {
      state.table3Key = tKey
      state.table3Index = index
      state.templateObjForm3 = state.templateObjForm2[tKey][0]
      state.form3 = state.form2[tKey][index]
      state.dialogFormVisible3 = true
    }
    const form3Edit = function (tKey, index) {
    }
    const form1Save = function () {
      state.dialogFormVisible1 = false
      SourceDataUpdate()
    }
    const form1Update = function () {
      SourceDataUpdate()
    }
    const form2Update = function () {
      SourceDataUpdate()
    }
    const form3Update = function () {
      SourceDataUpdate()
    }
    const form2Save = function () {
      state.dialogFormVisible2 = false
      SourceDataUpdate()
    }
    const form3Save = function () {
      state.dialogFormVisible3 = false
      SourceDataUpdate()
    }
    const AddColumn = function () {
      state.configData.push(JSON.parse(JSON.stringify(state.templateObj)))
      // EditForm(0,state.configData[0])
      tableChangePage(1)
      SourceDataUpdate()
    }
    const DeleteColumn = function (index) {
      state.configData.splice(index, 1)
      SourceDataUpdate()
    }
    const SourceDataUpdate = function () {
      emit('update:sourceData', {template: JSON.parse(state.template), config: state.configData})
    }
    const WitchTemplate = function () {
      try {
        JSON.parse(state.template)
        explainSourceData()
        SourceDataUpdate()
        state.templateErr = ''
      } catch (e) {
        state.templateErr = '错误'
      }
    }
    const Copy = function (content) {
      let index = copy.SetCopyContent(content)
      copy.handleCopy(index)
    }
    const state = reactive({
      template: '',//原始的配置模板
      templateErr: '',
      config: '', //原始的配置内容
      templateObj: {},
      templateObjForm1: {},
      templateObjForm2: {},
      templateObjForm3: {},
      configData: null,
      dialogFormVisible1: false,
      dialogFormVisible2: false,
      dialogFormVisible3: false,
      table1Index: '',
      table2Index: '',
      table3Index: '',
      table1Key: '',
      table2Key: '',
      table3Key: '',
      form1: {},
      form2: {},
      form3: {},
      tableIndexList: [],
      tablePage: 1,
      tableSize: 1000,
      tableSearch: '',
    })

    //初始化
    function init() {
      let data = props.data
      if (t.IsString(data)) {
        if (data === '') {
          state.template = []
          state.config = []
          state.configData = []
          state.templateObj = null
          return
        }
        data = JSON.parse(data)
      }
      if (data.template) {
        state.template = data.template || '[]'
        state.config = data.config || []
      } else {
        state.template = JSON.stringify([data[0]])
        state.config = data
      }
      explainSourceData()
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
      EditForm,
      UpMove,
      IsString,
      IsArray,
      IsObject,
      SourceDataUpdate,
      WitchTemplate,
      Copy,
      form1Edit,
      form2Edit,
      form3Edit,
      form1Save,
      form2Save,
      form3Save,
      form1Update,
      form2Update,
      form3Update,
      DeleteColumn,
      tableChangePage,
      tableFilter,
    }
  },
}
</script>

<el-button>
<slot/>
</el-button>
