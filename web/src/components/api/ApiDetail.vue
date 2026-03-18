<template>
  <div class="api-detail" tabindex="0" @keydown="handleKeyDown" @keyup="handleKeyUp">
    <div class="api-header">
      <el-input v-model="apiForm.name" placeholder="输入接口名称" style="width: 300px;margin-right:5px;" type="text" @blur="handleSave"></el-input>
      <div class="api-title-section">
        <el-input v-model="apiForm.url" placeholder="输入请求URL" @blur="handleSave">
          <template #prepend>
            <el-select v-model="apiForm.method" style="width: 80px">
              <el-option label="GET" value="GET"/>
              <el-option label="POST" value="POST"/>
            </el-select>
          </template>
        </el-input>
      </div>

      <div class="api-actions">
        <el-select v-model="envIdString" placeholder="选择环境" style="width: 160px" @change="changeEnv">
          <el-option
              v-for="env in envs"
              :key="env.id"
              :label="env.name"
              :value="env.id"
          />
        </el-select>
        <el-button :loading="executing" type="primary" @click="handleExecute">
          执行接口
        </el-button>
        <el-button @click="handleSave">保存</el-button>
        <el-button type="info" @click="showResult">结果</el-button>
      </div>
    </div>

    <el-tabs v-model="configActiveTab" class="detail-tabs" style="min-height: 500px;" @tab-change="responseTabChange">
      <el-tab-pane label="备注" name="desc">
        <MdEditor  v-model="apiForm.desc" @blur="handleSave" :onSave="handleSave" />
      </el-tab-pane>
      <el-tab-pane :label="'请求头(' + (isObject(apiForm.header_list) ? Object.keys(apiForm.header_list).length : 0) + ')'" name="headers">
        <headers-value-editor
            v-model="apiForm.header_list"
            :keys="headerSuggestions"
            @update="handleSaveHeaders"
        />
      </el-tab-pane>
      <el-tab-pane :label="'Url参数(' + (isArray(apiForm.query_params_data) ? apiForm.query_params_data.length : 0) + ')'" name="params">
        <key-value-editor :list="apiForm.query_params_data" @update="handleSaveUrls"/>
      </el-tab-pane>
      <el-tab-pane v-if="apiForm.method !== 'GET'" :label="'请求体(' + (isArray(apiForm.body_form_data) ? apiForm.body_form_data.length : 0) + ')'" name="body">
        <!-- 请求体内容（同原逻辑） -->
        <div style="width: 100%">
          <el-radio-group v-model="apiForm.content_type" class="detail-segmented" size="small" @change="handleSave">
            <el-radio-button value="application/json">application/json</el-radio-button>
            <el-radio-button value="application/x-www-form-urlencoded">x-www-form-urlencoded</el-radio-button>
            <el-radio-button value="multipart/form-data">multipart/form-data</el-radio-button>
            <el-radio-button value="text/plain">text/plain</el-radio-button>
            <el-radio-button value="raw">Raw</el-radio-button>
          </el-radio-group>

          <div v-if="apiForm.content_type === 'application/json'" class="body-editor">
            <json-editor-vue v-model="apiForm.body_json_data" class="json-box" @blur="handleSave"/>
          </div>
          <div v-else-if="['application/x-www-form-urlencoded', 'multipart/form-data'].includes(apiForm.content_type)" class="body-editor">
            <key-value-editor @update="handleSaveBodyFormData" :list="apiForm.body_form_data"/>
          </div>
          <div v-else-if="['text/plain', 'raw'].includes(apiForm.content_type)" class="body-editor">
            <el-input
                v-model="apiForm.body_raw_data"
                :rows="Number(6)"
                placeholder="输入原始数据"
                type="textarea"
                @blur="handleSave"
            />
          </div>
        </div>
      </el-tab-pane>
      <el-tab-pane :label="'结果提取(' + (isArray(apiForm.response_take_data) ? apiForm.response_take_data.length : 0) + ')'" name="response_take">
        <el-alert title="提取示例：data.0.token表示提取data数组的第一个元素中的token" type="info" :closable="false" style="margin: 5px;"/>
        <response-take-editor
            v-if="configActiveTab === 'response_take'"
            v-model="apiForm.response_take_data"
            :envItems="envItems"
            @update="updateResponseTake"
        />
      </el-tab-pane>
      <el-tab-pane v-if="parseInt(apiForm.env_id) > 0" :label="'环境变量(' + (isArray(envItems) ? envItems.length : 0) + ')'" lazy name="env_items" style="width: 96%;">
        <div class="config-section" v-if="parseInt(apiForm.env_id) > 0">
          <variable-manager
              ref="refVariableManager"
              @update="handleVariablesUpdate"
          />
        </div>
      </el-tab-pane>
      <el-tab-pane label="代码" lazy name="code">
        <div style="width: 100%">
          <el-radio-group v-model="apiForm.code_type" class="detail-segmented" size="small" @change="handleCodeTypeChange">
            <el-radio-button value="curl bash(chrome)">curl bash(chrome)</el-radio-button>
            <el-radio-button value="curl shell(apifox)">curl shell(apifox)</el-radio-button>
          </el-radio-group>

          <div class="response-body-container" style="margin-top:5px;">
            <button class="copy-btn" link @click="copyTextToClipboard(apiForm.code)">复制</button>
            <pre class="response-body json-body">{{
                apiForm.code
              }}
                  </pre>
          </div>
        </div>
      </el-tab-pane>
      <el-tab-pane :label="'结果字段备注(' + (isArray(apiForm.take_result_data) ? apiForm.take_result_data.length : 0) + ')'" lazy name="result_field_desc">
        <el-table
            :data="apiForm.take_result_data"
            style="width: 100%"
        >
          <el-table-column label="字段" width="400" align="center" fixed="right">
            <template #default="{ row }">
              <el-input
                  v-model="row.key"
                  placeholder=""
                  @blur="handleSave"
              />
            </template>
          </el-table-column>
          <el-table-column label="类型" width="200" align="center" fixed="right">
            <template #default="{ row }">
              <el-select v-model="row.type" placeholder="请选择" @change="handleSave">
                <el-option label="string" value="string"/>
                <el-option label="number" value="number"/>
                <el-option label="boolean" value="boolean"/>
                <el-option label="object" value="object"/>
                <el-option label="array" value="array"/>
              </el-select>
            </template>
          </el-table-column>
          <el-table-column label="备注" align="center" fixed="right">
            <template #default="{ row }">
              <el-input
                  v-model="row.desc"
                  placeholder=""
                  @blur="handleSave"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" align="center" fixed="right">
            <template #default="{ row }">
              <el-button link type="danger" @click="removeTakeResult(row.key)">删除</el-button>
            </template>
          </el-table-column>
        </el-table>
      </el-tab-pane>
    </el-tabs>

    <!-- 环境变量详情弹窗 -->
    <el-dialog
        v-model="showEnvDialog"
        :before-close="closeEnvDialog"
        title="环境变量详情"
        width="60%"
    >
      <div v-if="currentEnvName" class="dialog-env-name">
        当前环境: <strong>{{ currentEnvName }}</strong>
      </div>
      <el-table
          :data="envVariables"
          :show-header="true"
          max-height="400"
          style="width: 100%"
      >
        <el-table-column label="变量名" prop="key" show-overflow-tooltip width="150">
          <template #default="{ row }">
            ${{ row.key }}$
          </template>
        </el-table-column>
        <el-table-column label="值" prop="value" show-overflow-tooltip></el-table-column>
        <el-table-column label="描述" prop="desc" show-overflow-tooltip width="200"></el-table-column>
      </el-table>
      <div v-if="envVariables.length === 0" class="no-variables">
        该环境暂无变量
      </div>
      <template #footer>
        <span class="dialog-footer">
          <el-button @click="closeEnvDialog">关闭</el-button>
        </span>
      </template>
    </el-dialog>

    <el-drawer v-model="drawerHistoryShow" direction="rtl" size="60%">
      <div v-if="apiForm.last_result_data">
        <h5 @click="copyUrl(apiForm.last_result_data.url)">{{ apiForm.method }} {{ apiForm.last_result_data.url }}</h5>
        <div class="response-status">
          <el-button type="primary" :loading="executing" @click="handleExecute">执行</el-button>
          <div style="color:green;font-size:14px;">状态: {{ apiForm.last_result_data.status }}</div>
          <div v-if="apiForm.last_result_data.errmsg" style="color:red;font-size:14px;">执行错误:
            {{ apiForm.last_result_data.errmsg }}
          </div>
          <div style="color:green;font-size:14px;">耗时: {{ apiForm.last_result_data.millisecond }}ms</div>
          <div style="color:green;font-size:14px;">时间: {{ apiForm.last_result_data.request_time }}</div>
        </div>

        <el-tabs v-model="responseActiveTab" class="detail-tabs" @tab-change="handleSave">
          <el-tab-pane label="返回结果" name="body">
            <div class="response-body-container">
              <button class="copy-btn" link style="margin-right:120px;" @click="copyTextToClipboard(apiForm.last_result_data.result)">
                复制
              </button>
              <button v-if="isJsonResponse(apiForm.last_result_data.result)" class="copy-btn" link @click="takeToResult(apiForm.id , apiForm.last_result_data.result)">提取Json到文档</button>
              <pre v-if="isJsonResponse(apiForm.last_result_data.result)" class="response-body json-body">{{
                  formatJson(apiForm.last_result_data.result)
                }}
                  </pre>
              <pre v-else class="response-body">{{ apiForm.last_result_data.result }}</pre>
            </div>
          </el-tab-pane>
          <el-tab-pane label="请求头" name="headers">
            <key-value-view :data="apiForm.last_result_data.headers" />
          </el-tab-pane>
          <el-tab-pane
              v-if="apiForm.last_result_data.body_forms && apiForm.last_result_data.body_forms.length > 0"
              label="请求体"
              name="bodyForms"
          >
            <el-table :data="apiForm.last_result_data.body_forms" style="width: 100%">
              <el-table-column label="字段名" prop="field" width="200"></el-table-column>
              <el-table-column label="类型" prop="type" width="120"></el-table-column>
              <el-table-column label="值" prop="value" show-overflow-tooltip></el-table-column>
            </el-table>
          </el-tab-pane>
          <el-tab-pane
              v-if="apiForm.last_result_data.response_take && apiForm.last_result_data.response_take.length > 0"
              label="结果提取"
              name="responseTake"
          >
            <el-table :data="apiForm.last_result_data.response_take" style="width: 100%">
              <el-table-column label="字段名" prop="item_key" width="100"></el-table-column>
              <el-table-column label="提取路径" prop="value" show-overflow-tooltip></el-table-column>
              <el-table-column label="提取值" prop="take_value" show-overflow-tooltip></el-table-column>
            </el-table>
          </el-tab-pane>
        </el-tabs>
      </div>
      <div v-else class="no-response">
        <el-empty description="尚未执行请求"/>
      </div>
    </el-drawer>
  </div>
</template>

<script>
import {Link, Radio, RadioButton} from '@element-plus/icons-vue'
import KeyValueEditor from './KeyValueEditor.vue'
import KeyValueView from './KeyValueView.vue'
import typ from '@/utils/base/type'
import HeadersValueEditor from "@/components/api/HeadersValueEditor.vue"
import ResponseTakeEditor from "@/components/api/ResponseTakeEditor.vue"
import Api from '@/utils/base/api'
import Copy from '@/utils/base/copy'
import JsonEditorVue from 'json-editor-vue3'
import KeyDebounceDetector from '@/utils/base/keyup'
import VariableManager from "@/components/api/VariableManager.vue";
import Store from "@/utils/base/store"
import { MdEditor } from 'md-editor-v3';
import 'md-editor-v3/lib/style.css';

export default {
  name: 'ApiDetail',
  components: {
    MdEditor,
    VariableManager,
    KeyValueEditor,
    KeyValueView,
    HeadersValueEditor,
    JsonEditorVue,
    ResponseTakeEditor
  },
  props: {
    environment: {
      type: String,
      default: 'dev'
    }

  },
  data() {
    return {
      Link,
      mainActiveTab: 'config', // 默认显示请求配置
      executing: false,
      apiForm: {
        method: 'GET',
      },
      configActiveTab: 'body', // 默认显示请求头标签页
      responseActiveTab: 'body',
      headerSuggestions: [
        'Content-Type',
        'Authorization',
        'User-Agent',
        'Accept',
        'Cookie',
        'Token',
      ],
      envs: [],
      envItems: [],
      currentEnvId: '0',
      showEnvDialog: false,
      envVariables: [],
      currentEnvName: '',
      keyup: null,
      drawerHistoryShow: false,
      takeResultActiveTabName : 'take_result_data',
    }
  },
  computed: {
    envIdString: {
      get() {
        return this.apiForm.env_id
      },
      set(value) {
        this.apiForm.env_id = parseInt(value) || 0
      }
    },
  },
  expose: ['InitApiDetail', 'handleExecute'],
  methods: {
    removeTakeResult : function (key){
      this.apiForm.take_result_data = this.apiForm.take_result_data.filter((value, index) => value.key !== key);
      this.handleSave()
    },
    handleCodeTypeChange: function () {
      let _that = this
      Api.ApiCode({
        code_type: this.apiForm.code_type,
        id: _that.apiForm.id,
      }, function (res) {
        _that.apiForm.code = res.Data.code
      })
    },
    isArray: function (an) {
      return typ.IsArray(an)
    },
    isObject: function (an) {
      return typ.IsObject(an)
    },
    handleVariablesUpdate(variables) {
      this.envVariables = variables
    },
    handleSaveBodyFormData(bodyFormData){
      this.apiForm.body_form_data = bodyFormData
      this.handleSave()
    },

    handleKeyUp: function (event) {
      let _that = this
      _that.initKeyUp()
      _that.keyup.keyUp(event.key)
    },
    handleKeyDown: function (event) {
      let _that = this
      _that.initKeyUp()
      _that.keyup.keyDown(event.key)
    },
    InitApiDetail: function (apiInfo) {
      let _that = this
      _that.loadApiData(apiInfo)
      _that.loadEnvs()
      _that.loadEnvItems(apiInfo.env_id)
      _that.initKeyUp()
      _that.configActiveTab = Store.getStore(apiInfo.id + '_last_tab_name')
      if(_that.configActiveTab === '' || _that.configActiveTab === undefined || _that.configActiveTab === null){
        _that.configActiveTab = 'body'
      }
    },
    initKeyUp: function () {
      let _that = this
      if (_that.keyup !== null) {
        return
      }
      _that.keyup = new KeyDebounceDetector()
      _that.keyup.Register('Control', 's', null, function () {
        _that.handleSave()
      })
      _that.keyup.Register('Control', 'Enter', null, function () {
        _that.handleExecute()
      })
    },
    updateResponseTake: function (responseTakeData) {
      let _that = this
      _that.apiForm.response_take_data = responseTakeData
      _that.handleSave()
    },
    responseTabChange: function (key) {
      let _that = this
      Store.setStore(_that.apiForm.id + '_last_tab_name' , _that.configActiveTab)
      if (_that.configActiveTab === 'env_items') {
        _that.loadEnvItems(_that.apiForm.env_id)
      } else if (_that.configActiveTab === 'code') {
        _that.apiForm.code_type = 'curl bash(chrome)';
        _that.handleCodeTypeChange()
      }
    },
    changeEnv: function (env_id) {
      let _that = this
      _that.handleSave()
      _that.apiForm.env_id = env_id
      _that.loadEnvItems(_that.apiForm.env_id)
    },
    loadEnvItems: function (env_id) {
      let _that = this
      Api.CollectionEnvItems({
        collection_id: _that.apiForm.collection_id,
        env_id: env_id,
      }, function (res) {
        if (res.ErrCode !== 0) {
          _that.$message.error(res.ErrMsg)
          return
        }
        _that.envItems = res.Data.list
        _that.$nextTick(function () {
          if (_that.$refs.refVariableManager) {
            _that.$refs.refVariableManager.loadVariables({
              id: _that.apiForm.env_id,
              collection_id: _that.apiForm.collection_id,
              variables: _that.envItems
            })
          }
        })

      })
    },
    loadEnvs: function () {
      let _that = this
      Api.CollectionEnvs({
        collection_id: _that.apiForm.collection_id,
      }, function (res) {
        if (res.ErrCode !== 0) {
          _that.$message.error(res.ErrMsg)
          return
        }
        _that.envs = res.Data.list
        _that.envs.unshift({
          id: 0,
          name: '请选择',
        })
      })
    },
    closeEnvDialog: function () {
      this.showEnvDialog = false
      this.envVariables = []
      this.currentEnvName = ''
    },
    copyUrl: function (url) {
      if (typeof url !== 'string' || url.trim() === '') {
        this.$message.error('无可复制内容')
        return
      }
      let index = Copy.SetCopyContent(url)
      Copy.handleCopy(index)
    },
    loadApiData(api) {
      let _that = this
      _that.apiForm = JSON.parse(JSON.stringify(api))
      //headers处理
      _that.apiForm.header_list = JSON.parse(_that.apiForm.headers)
      if (!typ.IsObject(_that.apiForm.header_list)) {
        _that.apiForm.header_list = {}
      }
      //请求参数处理
      _that.apiForm.query_params_data = JSON.parse(_that.apiForm.query_params)
      if (!typ.IsArray(_that.apiForm.query_params_data)) {
        _that.apiForm.query_params_data = []
      }
      //body_json处理
      _that.apiForm.body_json_data = JSON.parse(_that.apiForm.body_json || '{}')
      if (!typ.IsObject(_that.apiForm.body_json_data)) {
        _that.apiForm.body_json_data = {}
      }
      //body_form处理
      _that.apiForm.body_form_data = JSON.parse(_that.apiForm.body_form)
      if (!typ.IsArray(_that.apiForm.body_form_data)) {
        _that.apiForm.body_form_data = []
      }
      //body_raw处理
      _that.apiForm.body_raw_data = _that.apiForm.body_raw || ''
      //结果提取配置处理
      _that.apiForm.response_take_data = JSON.parse(_that.apiForm.response_take)
      if (!typ.IsArray(_that.apiForm.response_take_data)) {
        _that.apiForm.response_take_data = []
      }
      //最后执行结果的处理
      _that.apiForm.last_result_data = JSON.parse(_that.apiForm.last_result || '{}')
      if (!typ.IsObject(_that.apiForm.last_result_data)) {
        _that.apiForm.last_result_data = {}
      }
      //提取结果
      _that.apiForm.take_result_data = JSON.parse(_that.apiForm.take_result || '[]')
      if (!typ.IsArray(_that.apiForm.take_result_data)) {
        _that.apiForm.take_result_data = []
      }
      console.log(_that.apiForm)
      if (_that.apiForm.method === 'GET') {
        _that.configActiveTab = 'params'
      } else if (_that.apiForm.method === 'POST') {
        _that.configActiveTab = 'body'
      }
    },
    showResult: function () {
      let _that = this
      _that.drawerHistoryShow = true
      _that.mainActiveTab = 'response'
    },
    handleExecute() {
      let _that = this
      _that.executing = true
      Api.ApiRun({id: _that.apiForm.id}, function (res) {
        _that.executing = false
        if (res.ErrCode !== 0) {
          _that.$message.error(res.ErrMsg)
          return
        }
        _that.apiForm.last_result_data = res.Data
        _that.apiForm.last_result = JSON.stringify(res.Data)
        _that.responseActiveTab = 'body'
        _that.mainActiveTab = 'response'
        _that.drawerHistoryShow = true
      })
    },
    descChange : function (result){
      this.apiForm.desc = result
    },
    handleSave() {
      this.$emit('update', {
        ...this.apiForm
      })
      this.$message.success('保存成功')
    },
    handleSaveHeaders : function (result){
      this.apiForm.header_list = result
      this.handleSave()
    },
    handleSaveUrls : function (result){
      this.apiForm.query_params_data = result
      this.handleSave()
    },

    // 检查响应体是否为JSON格式
    isJsonResponse(body) {
      try {
        JSON.parse(body)
        return true
      } catch (e) {
        return false
      }
    },
    takeToResult(id , jsonResult) {
      let _that = this
      Api.ApiTakeJsonResult({json : jsonResult , id : id} , function (res){
        if(res.ErrCode === 0){
          _that.apiForm.take_result_data = res.Data
          console.log(_that.apiForm.take_result_data)
          _that.handleSave()
          _that.configActiveTab = 'result_field_desc'
        }else{
          _that.$message.error(res.ErrMsg)
        }
      })
    },
    copyTextToClipboard(data) {
      let index = Copy.SetCopyContent(data)
      Copy.handleCopy(index)
    },
    formatJson(body) {
      try {
        const obj = JSON.parse(body);
        return JSON.stringify(obj, null, 2); // 缩进2个空格
      } catch (e) {
        return body;
      }
    },
  }
}
</script>

<style scoped>
.api-detail {
  padding: 12px 0;
  height: 100vh;
  overflow: hidden;
  display: flex;
  flex-direction: column;
  min-height: 720px;
}

.json-box {
  width: 100%;
  height: 360px;
  margin-top: 12px;
  border: 1px solid #dbe7d6;
  border-radius: 10px;
  overflow: auto;
  background: #fff;
}

.api-header {
  display: flex;
  justify-content: space-between;
  align-items: flex-start;
  margin-bottom: 10px;
  background: #f7f9f5;
  border: 1px solid #e6ece0;
  border-radius: 10px;
  padding: 10px 12px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.api-title-section {
  display: flex;
  align-items: center;
  gap: 12px;
  width: 95%;
  margin-right: 3px;
}

.config-section,
.response-section {
  height: 100%;
  overflow: auto;
  padding: 16px;
}

.method-tag {
  font-weight: bold;
  min-width: 60px;
  text-align: center;
}

.api-title {
  margin: 0;
  color: #303133;
}

.api-actions {
  display: flex;
  gap: 5px;
}

.api-content {
  display: flex;
  gap: 20px;
  flex: 1;
  width: 100%;
  overflow: hidden;
  height: calc(100vh - 120px); /* 计算可用高度 */
}

.config-section,
.response-section {
  background: #fff;
  border: 1px solid #e8eee5;
  padding: 20px;
  border-radius: 12px;
  box-shadow: 0 6px 18px rgba(80, 110, 80, 0.08);
  margin-top: 5px;
  overflow: auto;
  flex: 1;
  width: 100%; /* 限制最大宽度为50% */
  display: flex;
  flex-direction: column;
}

.el-button + .el-button {
  margin-left: 5px;
}

.config-section h3,
.response-section h3 {
  margin-top: 0;
  margin-bottom: 16px;
  color: #303133;
}

.body-editor {
  margin-top: 12px;
  width: 100%;
  overflow: hidden;
}

.empty-body {
  text-align: center;
  color: #909399;
  padding: 20px;
}

.response-status {
  display: block;
  align-items: center;
  gap: 12px;
  margin-bottom: 16px;
  border: 1px solid #e6ece0;
  background: #f7f9f5;
  border-radius: 10px;
  padding: 10px 12px;
}

.response-time {
  color: #909399;
  font-size: 14px;
}

.response-body-container {
  background: #2d2d2d;
  color: #f8f8f2;
  padding: 16px;
  border-radius: 8px;
  border: 1px solid #2f3a2f;
  overflow: auto;
  flex: 1; /* 占据剩余空间 */
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;

}

.response-body {
  margin: 0;
  word-wrap: break-word;
  max-width: 100%;
  overflow-x: auto;
}

.json-body {
  white-space: pre-wrap;
}

.no-response {
  text-align: center;
  padding: 40px 0;
  flex: 1;
  display: flex;
  align-items: center;
  justify-content: center;
}

.body-forms-container {
  padding: 10px 0;
}

:deep(.el-tabs__content) {
  padding: 12px;
  overflow: auto;
}

:deep(.detail-tabs > .el-tabs__header) {
  margin-bottom: 12px;
}

:deep(.detail-tabs > .el-tabs__header .el-tabs__nav-wrap::after) {
  background: #e6ece0;
}

:deep(.detail-tabs > .el-tabs__header .el-tabs__item) {
  height: 38px;
  padding: 0 14px;
  color: #5a6755;
  font-weight: 500;
  transition: color 0.2s ease;
}

:deep(.detail-tabs > .el-tabs__header .el-tabs__item:hover) {
  color: #4f7d4f;
}

:deep(.detail-tabs > .el-tabs__header .el-tabs__item.is-active) {
  color: #4f7d4f;
  font-weight: 600;
}

:deep(.detail-tabs > .el-tabs__header .el-tabs__active-bar) {
  background: #7aa676;
  height: 3px;
  border-radius: 999px;
}

:deep(.detail-segmented) {
  display: inline-flex;
  flex-wrap: wrap;
  gap: 8px;
  padding: 6px;
  background: #f7f9f5;
  border: 1px solid #e6ece0;
  border-radius: 12px;
}

:deep(.detail-segmented .el-radio-button__inner) {
  min-height: 32px;
  padding: 0 14px;
  border: 1px solid #d7e2d2;
  border-radius: 8px !important;
  background: #fff;
  color: #596655;
  font-weight: 500;
  line-height: 30px;
  box-shadow: none;
  transition: all 0.2s ease;
}

:deep(.detail-segmented .el-radio-button:first-child .el-radio-button__inner),
:deep(.detail-segmented .el-radio-button:last-child .el-radio-button__inner) {
  border-radius: 8px !important;
}

:deep(.detail-segmented .el-radio-button__original-radio:checked + .el-radio-button__inner) {
  background: #5a8a5a;
  border-color: #5a8a5a;
  color: #fff;
  box-shadow: 0 6px 14px rgba(90, 138, 90, 0.22);
}

:deep(.detail-segmented .el-radio-button__inner:hover) {
  color: #456e45;
  border-color: #afc7aa;
  background: #f4faf2;
}

:deep(.el-table) {
  border: 1px solid #e6ece0;
  border-radius: 10px;
  overflow: hidden;
}

:deep(.el-tabs__nav-wrap) {
  overflow: hidden;
}

:deep(.el-tabs__nav-scroll) {
  overflow: hidden;
}

:deep(.el-tabs__nav) {
  min-width: 100%;
}

/* 滚动条样式 */
.config-section::-webkit-scrollbar,
.response-section::-webkit-scrollbar {
  width: 6px;
}

.config-section::-webkit-scrollbar-track,
.response-section::-webkit-scrollbar-track {
  background: #f1f1f1;
  border-radius: 3px;
}

.config-section::-webkit-scrollbar-thumb,
.response-section::-webkit-scrollbar-thumb {
  background: #c1c1c1;
  border-radius: 3px;
}

.config-section::-webkit-scrollbar-thumb:hover,
.response-section::-webkit-scrollbar-thumb:hover {
  background: #a8a8a8;
}

.response-body-container::-webkit-scrollbar,
.body-editor::-webkit-scrollbar {
  width: 6px;
  height: 6px;
}

.response-body-container::-webkit-scrollbar-track,
.body-editor::-webkit-scrollbar-track {
  background: #444;
}

.response-body-container::-webkit-scrollbar-thumb,
.body-editor::-webkit-scrollbar-thumb {
  background: #666;
  border-radius: 3px;
}

.response-body-container::-webkit-scrollbar-thumb:hover,
.body-editor::-webkit-scrollbar-thumb:hover {
  background: #888;
}

.dialog-env-name {
  margin-bottom: 15px;
  padding: 10px;
  background-color: #f4faf2;
  border-left: 4px solid #5a8a5a;
  border-radius: 8px;
}

.no-variables {
  text-align: center;
  color: #909399;
  padding: 20px;
  font-size: 14px;
}

@media (max-width: 768px) {
  :deep(.detail-tabs > .el-tabs__header .el-tabs__item) {
    padding: 0 10px;
  }

  :deep(.detail-segmented) {
    width: 100%;
  }
}
</style>



