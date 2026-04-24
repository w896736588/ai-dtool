<template>
  <div class="api-detail" tabindex="0" @keydown="handleKeyDown" @keyup="handleKeyUp">
    <div class="api-header">
      <el-input v-model="apiForm.name" class="api-name-input" placeholder="输入接口名称" type="text" @blur="handleBlurSave"></el-input>
      <div class="api-title-section">
        <el-select v-model="apiForm.method" class="api-method-select">
          <el-option label="GET" value="GET"/>
          <el-option label="POST" value="POST"/>
        </el-select>
        <el-input v-model="apiForm.url" class="api-url-input" placeholder="输入请求URL" @blur="handleBlurSave"></el-input>
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
        <pl-button :loading="executing" type="primary" @click="handleExecute">
          执行接口
        </pl-button>
        <pl-button @click="handleSave">保存</pl-button>
        <pl-button type="info" @click="showResult">结果</pl-button>
      </div>
    </div>

    <el-tabs v-model="configActiveTab" class="detail-tabs" style="min-height: 500px;" @tab-change="responseTabChange">
      <el-tab-pane label="备注" name="desc">
        <MdEditor  v-model="apiForm.desc" @blur="handleBlurSave" :onSave="handleSave" />
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
            <el-radio-button value="raw">Raw</el-radio-button>
          </el-radio-group>

          <div v-if="apiForm.content_type === 'application/json'" class="body-editor body-editor-json">
            <json-editor-vue v-model="apiForm.body_json_data" class="json-box" @blur="handleBlurSave"/>
          </div>
          <div v-else-if="['application/x-www-form-urlencoded', 'multipart/form-data'].includes(apiForm.content_type)" class="body-editor">
            <div style="display: flex; justify-content: flex-end; margin-bottom: 8px;">
              <pl-button size="small" @click="openBodyJsonImportDialog">通过JSON导入</pl-button>
            </div>
            <key-value-editor @update="handleSaveBodyFormData" :list="apiForm.body_form_data"/>
          </div>
          <div v-else-if="apiForm.content_type === 'raw'" class="body-editor body-editor-raw">
            <el-input
                v-model="apiForm.body_raw_data"
                placeholder="输入原始数据"
                type="textarea"
                :autosize="{ minRows: 12, maxRows: 40 }"
                @blur="handleBlurSave"
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
            <el-radio-button
                v-for="codeType in codeTypeOptions"
                :key="codeType"
                :value="codeType"
            >
              {{ codeType }}
            </el-radio-button>
          </el-radio-group>

          <div class="response-body-container" style="margin-top:5px;">
            <pl-button class="copy-btn" link @click="copyTextToClipboard(apiForm.code)">复制</pl-button>
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
            max-height="calc(100vh - 200px)"
            class="result-field-table"
        >
          <el-table-column label="字段" width="400" align="center" fixed="right">
            <template #default="{ row }">
              <el-input
                  v-model="row.key"
                  placeholder=""
                  @blur="handleBlurSave"
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
                  @blur="handleBlurSave"
              />
            </template>
          </el-table-column>
          <el-table-column label="操作" width="200" align="center" fixed="right">
            <template #default="{ row }">
              <pl-button link type="danger" @click="removeTakeResult(row.key)">删除</pl-button>
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
          <pl-button @click="closeEnvDialog">关闭</pl-button>
        </span>
      </template>
    </el-dialog>

    <el-drawer v-model="drawerHistoryShow" direction="rtl" size="60%">
      <div v-if="apiForm.last_result_data">
        <div class="request-url-bar">
          <div class="request-url-main">
            <div class="request-url-text">{{ apiForm.method }} {{ apiForm.last_result_data.url }}</div>
            <pl-button
              class="request-url-copy-btn"
              link
              type="primary"
              @click="copyUrl(apiForm.last_result_data.url)"
            >
              <el-icon><CopyDocument /></el-icon>
            </pl-button>
          </div>
          <pl-button class="request-run-btn" type="success" :loading="executing" @click="handleExecute">
            <el-icon><VideoPlay /></el-icon>执行
          </pl-button>
        </div>
        <div class="response-status">
          <div style="color:green;font-size:14px;">状态: {{ apiForm.last_result_data.status }}</div>
          <div v-if="apiForm.last_result_data.errmsg" style="color:red;font-size:14px;">执行错误:
            {{ apiForm.last_result_data.errmsg }}
          </div>
          <div style="color:green;font-size:14px;">耗时: {{ apiForm.last_result_data.millisecond }}ms</div>
          <div style="color:green;font-size:14px;">时间: {{ apiForm.last_result_data.request_time }}</div>
        </div>

        <el-tabs v-model="responseActiveTab" class="detail-tabs" @tab-change="handleSave">
          <el-tab-pane label="返回结果" name="body">
            <div class="response-body-container" ref="responseContainerRef" @scroll="handleResponseScroll">
              <div class="response-toolbar">
                <pl-button class="response-toolbar-btn" @click="copyTextToClipboard(responseResultText)">
                  复制
                </pl-button>
                <pl-button
                    v-if="isJsonResponse(responseResultText)"
                    class="response-toolbar-btn"
                    type="info"
                    @click="takeToResult(apiForm.id , responseResultText)"
                >
                  提取Json到文档
                </pl-button>
                <el-radio-group v-model="responseViewMode" class="detail-segmented response-view-mode" size="small">
                  <el-radio-button value="auto">自动</el-radio-button>
                  <el-radio-button value="json">JSON</el-radio-button>
                  <el-radio-button value="html">HTML预览</el-radio-button>
                  <el-radio-button value="raw">原始</el-radio-button>
                </el-radio-group>
              </div>
              <pre v-if="resolvedResponseViewMode === 'json'" class="response-body json-body">{{
                  formatJson(responseResultText)
                }}</pre>
              <iframe
                  v-else-if="resolvedResponseViewMode === 'html'"
                  class="response-html-preview"
                  :srcdoc="sanitizedResponseHtml"
                  sandbox=""
              />
              <pre v-else class="response-body">{{ responseResultText }}</pre>
              <div
                v-if="showScrollBtn"
                class="scroll-to-btn"
                @click="handleScrollTo"
                :title="isScrolledToBottom ? '回到顶部' : '滚动到底部'"
              >
                <el-icon :size="20"><ArrowUp v-if="isScrolledToBottom" /><ArrowDown v-else /></el-icon>
              </div>
            </div>
          </el-tab-pane>
          <el-tab-pane label="请求头" name="headers">
            <key-value-view :data="requestHeadersData" />
          </el-tab-pane>
          <el-tab-pane label="返回头" name="responseHeaders">
            <key-value-view :data="responseHeadersData" />
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

    <el-dialog v-model="bodyJsonImportVisible" title="通过JSON导入" width="600" @keydown.enter.prevent="handleBodyJsonImport">
      <el-alert title="请输入JSON对象，键值将自动转换为表单参数。支持嵌套对象（自动展平为 key.subkey 格式）" type="info" :closable="false" style="margin-bottom: 12px;"/>
      <el-input
          v-model="bodyJsonImportText"
          type="textarea"
          :rows="12"
          placeholder='例如：{"username":"admin","password":"123456"}'
      />
      <template #footer>
        <pl-button @click="bodyJsonImportVisible = false">取消</pl-button>
        <pl-button type="primary" @click="handleBodyJsonImport">导入</pl-button>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import {Link, Radio, RadioButton, CopyDocument, VideoPlay, ArrowUp, ArrowDown} from '@element-plus/icons-vue'
import KeyValueEditor from './KeyValueEditor.vue'
import KeyValueView from './KeyValueView.vue'
import typ from '@/utils/base/type'
import HeadersValueEditor from "@/components/api/HeadersValueEditor.vue"
import ResponseTakeEditor from "@/components/api/ResponseTakeEditor.vue"
import Api from '@/utils/base/api'
import Copy from '@/utils/base/copy'
import apiDetailParser from '@/utils/api_detail_parser.cjs'
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
    ResponseTakeEditor,
    CopyDocument,
    VideoPlay,
    ArrowUp,
    ArrowDown
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
      responseViewMode: 'auto',
      headerSuggestions: [
        'Content-Type',
        'Authorization',
        'User-Agent',
        'Accept',
        'Cookie',
        'Token',
      ],
      envs: [],
      codeTypeOptions: [
        'curl bash(chrome)',
        'curl shell(apifox)',
        'JavaScript fetch',
        'JavaScript axios',
        'Python requests',
        'PHP cURL',
        'Golang net/http',
        'Postman collection',
      ],
      envItems: [],
      currentEnvId: '0',
      showEnvDialog: false,
      envVariables: [],
      currentEnvName: '',
        keyup: null,
        isTabNavigating: false,
        drawerHistoryShow: false,
        takeResultActiveTabName : 'take_result_data',
        bodyJsonImportVisible: false,
        bodyJsonImportText: '',
        bodyFormNextId: 1,
        isScrolledToBottom: false,
        showScrollBtn: false,
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
    responseResultText() {
      return this.normalizeResponseBody(this.apiForm?.last_result_data?.result)
    },
    resolvedResponseViewMode() {
      return this.resolveResponseViewMode(
          this.responseViewMode,
          this.responseResultText,
          this.responseHeadersData
      )
    },
    sanitizedResponseHtml() {
      if (this.resolvedResponseViewMode !== 'html') {
        return ''
      }
      return this.sanitizeHtmlForPreview(this.responseResultText)
    },
    requestHeadersData() {
      return this.normalizeHeadersData(
          this.apiForm?.last_result_data?.request_headers || this.apiForm?.last_result_data?.headers
      )
    },
    responseHeadersData() {
      return this.normalizeHeadersData(this.apiForm?.last_result_data?.response_headers)
    },
  },
  expose: ['InitApiDetail', 'handleExecute'],
  methods: {
    removeTakeResult : function (key){
      this.apiForm.take_result_data = this.apiForm.take_result_data.filter((value, index) => value.key !== key);
      this.handleSave()
    },
    // ensureCodeType 中文：确保代码 tab 总有一个可用类型。 English: Ensure the code tab always has a valid snippet type.
    ensureCodeType() {
      if (!this.codeTypeOptions.includes(this.apiForm.code_type)) {
        this.apiForm.code_type = this.codeTypeOptions[0]
      }
    },
    // ensureCodeSnippetLoaded 中文：仅在代码 tab 激活时拉取代码片段。 English: Load snippets only when the code tab is active.
    ensureCodeSnippetLoaded() {
      this.ensureCodeType()
      if (!this.apiForm.id || this.configActiveTab !== 'code') {
        return
      }
      this.handleCodeTypeChange()
    },
    handleCodeTypeChange: function () {
      let _that = this
      _that.ensureCodeType()
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
      this.handleBlurSave()
    },
    openBodyJsonImportDialog() {
      this.bodyJsonImportText = ''
      this.bodyJsonImportVisible = true
    },
    handleBodyJsonImport() {
      if (!this.bodyJsonImportText.trim()) {
        this.$message.warning('请输入JSON数据')
        return
      }
      let parsed
      try {
        parsed = JSON.parse(this.bodyJsonImportText)
      } catch (e) {
        this.$message.error('JSON格式错误，请检查输入')
        return
      }
      if (typeof parsed !== 'object' || parsed === null || Array.isArray(parsed)) {
        this.$message.error('请输入JSON对象，而非数组或其他类型')
        return
      }
      const flattened = this.flattenJsonObject(parsed)
      const maxId = this.apiForm.body_form_data.reduce((max, item) => {
        return item.id && item.id > max ? item.id : max
      }, 0)
      this.bodyFormNextId = maxId + 1
      const newItems = flattened.map(([key, value]) => ({
        id: this.bodyFormNextId++,
        field: key,
        type: this.detectFormValueType(value),
        value: typeof value === 'object' && value !== null ? JSON.stringify(value) : String(value),
        description: '',
      }))
      this.apiForm.body_form_data = newItems
      this.handleBlurSave()
      this.bodyJsonImportVisible = false
      this.$message.success(`已导入 ${newItems.length} 个参数`)
    },
    flattenJsonObject(obj, prefix = '') {
      const result = []
      for (const [key, value] of Object.entries(obj)) {
        const fullKey = prefix ? `${prefix}.${key}` : key
        if (value !== null && typeof value === 'object' && !Array.isArray(value)) {
          result.push(...this.flattenJsonObject(value, fullKey))
        } else if (Array.isArray(value)) {
          result.push([fullKey, JSON.stringify(value)])
        } else {
          result.push([fullKey, value])
        }
      }
      return result
    },
    detectFormValueType(value) {
      if (typeof value === 'number') {
        return Number.isInteger(value) ? 'integer' : 'float'
      }
      if (typeof value === 'boolean') return 'boolean'
      return 'string'
    },

    handleKeyUp: function (event) {
      let _that = this
      _that.initKeyUp()
      if (event.key === 'Tab') {
        _that.isTabNavigating = false
      }
      _that.keyup.keyUp(event.key)
    },
    handleKeyDown: function (event) {
      let _that = this
      _that.initKeyUp()
      if (event.key === 'Tab') {
        _that.isTabNavigating = true
      }
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
      _that.$nextTick(function () {
        _that.ensureCodeSnippetLoaded()
      })
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
      _that.handleBlurSave()
    },
    responseTabChange: function (key) {
      let _that = this
      Store.setStore(_that.apiForm.id + '_last_tab_name' , _that.configActiveTab)
      if (_that.configActiveTab === 'env_items') {
        _that.loadEnvItems(_that.apiForm.env_id)
      } else if (_that.configActiveTab === 'code') {
        _that.ensureCodeSnippetLoaded()
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
      _that.apiForm.header_list = apiDetailParser.parseApiObjectField(_that.apiForm.headers, {})
      if (!typ.IsObject(_that.apiForm.header_list)) {
        _that.apiForm.header_list = {}
      }
      //请求参数处理
      _that.apiForm.query_params_data = apiDetailParser.parseApiArrayField(_that.apiForm.query_params, [])
      if (!typ.IsArray(_that.apiForm.query_params_data)) {
        _that.apiForm.query_params_data = []
      }
      //body_json处理
      _that.apiForm.body_json_data = apiDetailParser.parseApiObjectField(_that.apiForm.body_json, {})
      if (!typ.IsObject(_that.apiForm.body_json_data)) {
        _that.apiForm.body_json_data = {}
      }
      //body_form处理
      _that.apiForm.body_form_data = apiDetailParser.parseApiArrayField(_that.apiForm.body_form, [])
      if (!typ.IsArray(_that.apiForm.body_form_data)) {
        _that.apiForm.body_form_data = []
      }
      //body_raw处理
      _that.apiForm.body_raw_data = _that.apiForm.body_raw || ''
      _that.ensureCodeType()
      //结果提取配置处理
      _that.apiForm.response_take_data = apiDetailParser.parseApiArrayField(_that.apiForm.response_take, [])
      if (!typ.IsArray(_that.apiForm.response_take_data)) {
        _that.apiForm.response_take_data = []
      }
      //最后执行结果的处理
      _that.apiForm.last_result_data = apiDetailParser.parseApiObjectField(_that.apiForm.last_result, {})
      if (!typ.IsObject(_that.apiForm.last_result_data)) {
        _that.apiForm.last_result_data = {}
      }
      //提取结果
      _that.apiForm.take_result_data = apiDetailParser.parseApiArrayField(_that.apiForm.take_result, [])
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
      _that.responseViewMode = 'auto'
      _that.$nextTick(() => {
        _that.showScrollBtn = false
        _that.isScrolledToBottom = false
      })
    },
    handleResponseScroll(e) {
      const el = e.target
      const threshold = 50
      this.isScrolledToBottom = el.scrollHeight - el.scrollTop - el.clientHeight < threshold
      this.showScrollBtn = el.scrollHeight > el.clientHeight
    },
    handleScrollTo() {
      const el = this.$refs.responseContainerRef
      if (!el) return
      if (this.isScrolledToBottom) {
        el.scrollTo({ top: 0, behavior: 'smooth' })
      } else {
        el.scrollTo({ top: el.scrollHeight, behavior: 'smooth' })
      }
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
        _that.responseViewMode = 'auto'
        _that.mainActiveTab = 'response'
        _that.drawerHistoryShow = true
      })
    },
    descChange : function (result){
      this.apiForm.desc = result
    },
    handleSave() {
      // 当请求体类型为 application/json 时，校验 body_json_data 是否为合法 JSON
      if (this.apiForm.method === 'POST' && this.apiForm.content_type === 'application/json') {
        const bodyJsonData = this.apiForm.body_json_data
        // json-editor-vue3 在内容不合法时可能返回 undefined、空字符串或解析失败的对象
        if (bodyJsonData === undefined || bodyJsonData === null || bodyJsonData === '') {
          this.$message.error('请求体JSON格式错误，请检查输入内容是否为合法的JSON')
          return
        }
        // 尝试序列化，验证是否能正确转为 JSON 字符串
        try {
          const jsonStr = typeof bodyJsonData === 'object' ? JSON.stringify(bodyJsonData) : String(bodyJsonData)
          if (!jsonStr || jsonStr.trim() === '') {
            this.$message.error('请求体JSON格式错误，请检查输入内容是否为合法的JSON')
            return
          }
          // 如果是字符串，尝试解析验证
          if (typeof bodyJsonData === 'string') {
            JSON.parse(bodyJsonData)
          }
        } catch (e) {
          this.$message.error('请求体JSON格式错误，请检查输入内容是否为合法的JSON')
          return
        }
      }
      this.$emit('update', {
        ...this.apiForm
      })
      this.$message.success('保存成功')
    },
    // handleBlurSave 只处理失焦引发的自动保存，按 Tab 切换字段时跳过一次，避免焦点被刷新打断。
    handleBlurSave() {
      if (this.isTabNavigating) {
        return
      }
      this.handleSave()
    },
    handleSaveHeaders : function (result){
      this.apiForm.header_list = result
      this.handleBlurSave()
    },
    handleSaveUrls : function (result){
      this.apiForm.query_params_data = result
      this.handleBlurSave()
    },

    // 检查响应体是否为JSON格式
    normalizeResponseBody(body) {
      if (body === null || body === undefined) {
        return ''
      }
      if (typeof body === 'string') {
        return body
      }
      try {
        return JSON.stringify(body, null, 2)
      } catch (error) {
        return String(body)
      }
    },
    normalizeHeadersData(headers) {
      if (!typ.IsObject(headers)) {
        return {}
      }
      return headers
    },
    isJsonResponse(body) {
      const text = this.normalizeResponseBody(body).trim()
      if (text === '') {
        return false
      }
      try {
        JSON.parse(text)
        return true
      } catch (error) {
        return false
      }
    },
    isHtmlResponse(body) {
      const text = this.normalizeResponseBody(body).trim()
      if (text === '') {
        return false
      }
      return /<!doctype\s+html|<html[\s>]|<head[\s>]|<body[\s>]|<[a-z][\s\S]*>/i.test(text)
    },
    getHeaderValue(headers, key) {
      if (!typ.IsObject(headers)) {
        return ''
      }
      const target = String(key || '').toLowerCase()
      const headerKey = Object.keys(headers).find((item) => item.toLowerCase() === target)
      if (!headerKey) {
        return ''
      }
      const value = headers[headerKey]
      if (value === null || value === undefined) {
        return ''
      }
      return String(value)
    },
    resolveResponseViewMode(selectedMode, body, headers) {
      if (selectedMode !== 'auto') {
        return selectedMode
      }
      const contentType = this.getHeaderValue(headers, 'content-type').toLowerCase()
      const bodyIsJson = this.isJsonResponse(body)
      const bodyIsHtml = this.isHtmlResponse(body)

      if (bodyIsHtml && !bodyIsJson) {
        return 'html'
      }
      if (bodyIsJson && !bodyIsHtml) {
        return 'json'
      }

      if ((contentType.includes('text/html') || contentType.includes('application/xhtml+xml')) && this.isHtmlResponse(body)) {
        return 'html'
      }
      if ((contentType.includes('application/json') || contentType.includes('+json')) && this.isJsonResponse(body)) {
        return 'json'
      }

      if (bodyIsHtml) {
        return 'html'
      }
      if (bodyIsJson) {
        return 'json'
      }
      return 'raw'
    },
    escapeHtml(text) {
      return String(text)
          .replace(/&/g, '&amp;')
          .replace(/</g, '&lt;')
          .replace(/>/g, '&gt;')
          .replace(/"/g, '&quot;')
          .replace(/'/g, '&#39;')
    },
    buildPreviewDocument(bodyContent, headContent = '') {
      return `<!doctype html><html><head><meta charset="utf-8">${headContent}<style>html,body{margin:0;padding:0;background:#fff;}body{padding:12px;box-sizing:border-box;word-break:break-word;}img,video,canvas,svg{max-width:100%;height:auto;}</style></head><body>${bodyContent}</body></html>`
    },
    sanitizeHtmlForPreview(html) {
      const text = this.normalizeResponseBody(html)
      if (text.trim() === '') {
        return this.buildPreviewDocument('<div>Empty response.</div>')
      }
      if (typeof DOMParser === 'undefined') {
        return this.buildPreviewDocument(`<pre>${this.escapeHtml(text)}</pre>`)
      }
      try {
        const parser = new DOMParser()
        const doc = parser.parseFromString(text, 'text/html')
        const forbiddenSelectors = ['script', 'iframe', 'object', 'embed', 'base', 'meta[http-equiv="refresh"]']
        forbiddenSelectors.forEach((selector) => {
          doc.querySelectorAll(selector).forEach((node) => node.remove())
        })

        doc.querySelectorAll('*').forEach((element) => {
          Array.from(element.attributes).forEach((attribute) => {
            const attrName = attribute.name.toLowerCase()
            const attrValue = attribute.value || ''
            if (attrName.startsWith('on') || attrName === 'srcdoc') {
              element.removeAttribute(attribute.name)
              return
            }
            if (['href', 'src', 'xlink:href'].includes(attrName) && /^\s*javascript:/i.test(attrValue)) {
              element.removeAttribute(attribute.name)
              return
            }
            if (attrName === 'style' && /expression\s*\(/i.test(attrValue)) {
              element.removeAttribute(attribute.name)
            }
          })
        })
        const headContent = doc.head ? doc.head.innerHTML : ''
        const bodyContent = doc.body ? doc.body.innerHTML : this.escapeHtml(text)
        return this.buildPreviewDocument(bodyContent, headContent)
      } catch (error) {
        return this.buildPreviewDocument(`<pre>${this.escapeHtml(text)}</pre>`)
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
  ,
  beforeUnmount() {
    this.isTabNavigating = false
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
  margin-top: 0;
  border: 0;
  border-radius: 12px;
  overflow: hidden;
  background: transparent;
  box-shadow: none;
}

.api-header {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 10px;
  background: #f7f9f5;
  border: 1px solid #e6ece0;
  border-radius: 10px;
  padding: 10px 12px;
  position: sticky;
  top: 0;
  z-index: 100;
}

.api-name-input {
  flex: 0 0 240px;
  max-width: 300px;
}

.api-title-section {
  display: flex;
  align-items: center;
  gap: 12px;
  flex: 1 1 auto;
  min-width: 0;
}

.api-method-select {
  flex: 0 0 92px;
}

.api-url-input {
  flex: 1 1 auto;
  min-width: 0;
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
  flex: 0 0 auto;
  align-items: center;
  gap: 8px;
  margin-left: auto;
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

.body-editor-json {
  padding: 10px;
  border: 1px solid #e6ece0;
  border-radius: 12px;
  background: linear-gradient(180deg, #f8fbf6 0%, #f4f8f1 100%);
  box-shadow: inset 0 1px 0 rgba(255, 255, 255, 0.9);
  transition: border-color 0.2s ease, box-shadow 0.2s ease, background 0.2s ease;
}

.body-editor-json:focus-within {
  border-color: #8db28a;
  box-shadow: 0 0 0 3px rgba(122, 166, 118, 0.16);
  background: linear-gradient(180deg, #fbfdf9 0%, #f6faf3 100%);
}

.body-editor-raw {
  max-height: calc(100vh - 340px);
  overflow-y: auto;
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

.request-url-bar {
  display: flex;
  align-items: center;
  gap: 12px;
  margin-bottom: 12px;
  padding: 10px 12px;
  border: 1px solid #e6ece0;
  border-radius: 12px;
  background: linear-gradient(180deg, #f8fbf5 0%, #f3f8ef 100%);
  box-shadow: 0 6px 16px rgba(84, 116, 84, 0.08);
}

.request-url-main {
  display: flex;
  align-items: center;
  gap: 4px;
  flex: 1 1 auto;
  min-width: 0;
}

.request-url-text {
  color: #3f6f3f;
  font-size: 14px;
  font-weight: 600;
  line-height: 1.5;
  word-break: break-all;
  min-width: 0;
}

.request-url-copy-btn {
  flex: 0 0 auto;
  font-size: 15px;
  padding: 0;
}

.request-run-btn {
  flex: 0 0 auto;
  min-height: 34px;
  padding: 0 14px;
  border-radius: 10px;
  border-color: #6ea46b;
  background: linear-gradient(180deg, #7cbc76 0%, #659f61 100%);
  box-shadow: 0 8px 18px rgba(101, 159, 97, 0.24);
}

.request-run-btn:hover,
.request-run-btn:focus-visible {
  border-color: #5d9758;
  background: linear-gradient(180deg, #72b56d 0%, #5a9356 100%);
}

.response-time {
  color: #909399;
  font-size: 14px;
}

.response-body-container {
  background: #eef3ea;
  color: #2f3d32;
  padding: 16px;
  border-radius: 12px;
  border: 0;
  overflow: auto;
  flex: 1; /* 占据剩余空间 */
  max-height: min(56vh, calc(100vh - 320px));
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 14px;
  box-shadow: none;
  scrollbar-width: auto;
  scrollbar-color: #8ea88f #dbe7d8;
  position: relative;
}

.response-toolbar {
  display: flex;
  align-items: center;
  flex-wrap: wrap;
  gap: 8px;
  margin-bottom: 12px;
  position: sticky;
  top: -16px;
  z-index: 10;
  background: #eef3ea;
  padding: 8px 0;
  margin-top: -8px;
}

.scroll-to-btn {
  position: sticky;
  bottom: 8px;
  float: right;
  width: 36px;
  height: 36px;
  border-radius: 50%;
  background: rgba(126, 172, 123, 0.85);
  color: #fff;
  display: flex;
  align-items: center;
  justify-content: center;
  cursor: pointer;
  box-shadow: 0 2px 8px rgba(0, 0, 0, 0.15);
  transition: background 0.2s, transform 0.2s;
  z-index: 10;
}

.scroll-to-btn:hover {
  background: rgba(101, 159, 97, 0.95);
  transform: scale(1.1);
}

.response-view-mode {
  margin-left: auto;
}

.response-toolbar-btn {
  border-radius: 8px;
  border-color: #b7c7bb;
  background: rgba(255, 255, 255, 0.82);
  color: #35543b;
}

.response-toolbar-btn:hover,
.response-toolbar-btn:focus-visible {
  border-color: #8fac95;
  background: #ffffff;
  color: #27422d;
}

.response-body {
  margin: 0;
  word-wrap: break-word;
  max-width: 100%;
  overflow-x: auto;
  color: inherit;
}

.response-html-preview {
  width: 100%;
  min-height: 360px;
  border: 1px solid #d8e4d5;
  border-radius: 10px;
  background: #fff;
}

.response-body-container :deep(pre.response-body) {
  border: 0 !important;
  border-radius: 0 !important;
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

/* JSON editor skin / JSON 编辑器浅色卡片皮肤 */
:deep(.body-editor-json .jse-theme-light),
:deep(.body-editor-json .jse-main),
:deep(.body-editor-json .jse-container),
:deep(.body-editor-json .jsoneditor) {
  border-radius: 10px !important;
  border: 1px solid #dbe7d6 !important;
  background: #ffffff !important;
  box-shadow: 0 8px 18px rgba(80, 110, 80, 0.07) !important;
  overflow: hidden !important;
}

:deep(.body-editor-json .jse-menu),
:deep(.body-editor-json .jsoneditor-menu),
:deep(.body-editor-json .jse-navigation-bar),
:deep(.body-editor-json .jsoneditor-navigation-bar),
:deep(.body-editor-json .jse-status-bar),
:deep(.body-editor-json .jsoneditor-statusbar) {
  background: #f7f9f5 !important;
  border-color: #e6ece0 !important;
  color: #5a6755 !important;
  box-shadow: none !important;
}

:deep(.body-editor-json .jse-menu button),
:deep(.body-editor-json .jsoneditor-menu button),
:deep(.body-editor-json .jse-navigation-bar button),
:deep(.body-editor-json .jsoneditor-navigation-bar button) {
  border-radius: 8px !important;
  color: #5b6857 !important;
}

:deep(.body-editor-json .jse-menu button:hover),
:deep(.body-editor-json .jsoneditor-menu button:hover),
:deep(.body-editor-json .jse-navigation-bar button:hover),
:deep(.body-editor-json .jsoneditor-navigation-bar button:hover) {
  background: #edf5e9 !important;
  color: #456e45 !important;
}

:deep(.body-editor-json .jse-contents),
:deep(.body-editor-json .jsoneditor-outer),
:deep(.body-editor-json .jsoneditor-tree),
:deep(.body-editor-json .jsoneditor-text),
:deep(.body-editor-json .cm-editor),
:deep(.body-editor-json .cm-scroller),
:deep(.body-editor-json .cm-content),
:deep(.body-editor-json textarea) {
  background: #ffffff !important;
  color: #3f4c3b !important;
}

:deep(.body-editor-json .cm-editor) {
  min-height: 338px;
}

:deep(.body-editor-json .cm-focused),
:deep(.body-editor-json .jsoneditor:focus-within),
:deep(.body-editor-json .jse-main:focus-within) {
  outline: none !important;
  box-shadow: inset 0 0 0 1px #7aa676 !important;
}

:deep(.body-editor-json .cm-activeLine),
:deep(.body-editor-json .cm-activeLineGutter),
:deep(.body-editor-json .jsoneditor tr:hover) {
  background: #f4faf2 !important;
}

:deep(.body-editor-json .cm-selectionBackground),
:deep(.body-editor-json .cm-focused .cm-selectionBackground),
:deep(.body-editor-json ::selection) {
  background: rgba(122, 166, 118, 0.20) !important;
}

:deep(.body-editor-json .cm-gutters),
:deep(.body-editor-json .jsoneditor-tree .jsoneditor-contextmenu),
:deep(.body-editor-json .jsoneditor-tree .jsoneditor-field),
:deep(.body-editor-json .jsoneditor-tree .jsoneditor-value) {
  background: #fbfdf9 !important;
  color: #61705d !important;
  border-color: #edf2ea !important;
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
  width: 10px;
  height: 10px;
}

.response-body-container::-webkit-scrollbar-track,
.body-editor::-webkit-scrollbar-track {
  background: #dbe7d8;
  border-radius: 999px;
}

.response-body-container::-webkit-scrollbar-thumb,
.body-editor::-webkit-scrollbar-thumb {
  background: #8ea88f;
  border-radius: 999px;
  border: 2px solid #dbe7d8;
}

.response-body-container::-webkit-scrollbar-thumb:hover,
.body-editor::-webkit-scrollbar-thumb:hover {
  background: #6f8f72;
}

.body-editor-json::-webkit-scrollbar,
:deep(.body-editor-json .cm-scroller::-webkit-scrollbar),
:deep(.body-editor-json .jse-contents::-webkit-scrollbar),
:deep(.body-editor-json .jsoneditor::-webkit-scrollbar) {
  width: 8px;
  height: 8px;
}

.body-editor-json::-webkit-scrollbar-track,
:deep(.body-editor-json .cm-scroller::-webkit-scrollbar-track),
:deep(.body-editor-json .jse-contents::-webkit-scrollbar-track),
:deep(.body-editor-json .jsoneditor::-webkit-scrollbar-track) {
  background: #eef4ea;
  border-radius: 999px;
}

.body-editor-json::-webkit-scrollbar-thumb,
:deep(.body-editor-json .cm-scroller::-webkit-scrollbar-thumb),
:deep(.body-editor-json .jse-contents::-webkit-scrollbar-thumb),
:deep(.body-editor-json .jsoneditor::-webkit-scrollbar-thumb) {
  background: #bfd0b9;
  border-radius: 999px;
}

.body-editor-json::-webkit-scrollbar-thumb:hover,
:deep(.body-editor-json .cm-scroller::-webkit-scrollbar-thumb:hover),
:deep(.body-editor-json .jse-contents::-webkit-scrollbar-thumb:hover),
:deep(.body-editor-json .jsoneditor::-webkit-scrollbar-thumb:hover) {
  background: #a9bda3;
}

.result-field-table :deep(.el-table__body-wrapper) {
  scrollbar-width: thin;
  scrollbar-color: #5a8a5a #e6f0e3;
}

.result-field-table :deep(.el-table__body-wrapper::-webkit-scrollbar) {
  width: 10px;
  height: 10px;
}

.result-field-table :deep(.el-table__body-wrapper::-webkit-scrollbar-track) {
  background: #e6f0e3;
  border-radius: 999px;
}

.result-field-table :deep(.el-table__body-wrapper::-webkit-scrollbar-thumb) {
  background: #5a8a5a;
  border-radius: 999px;
  border: 2px solid #e6f0e3;
}

.result-field-table :deep(.el-table__body-wrapper::-webkit-scrollbar-thumb:hover) {
  background: #4a7a4a;
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
  .api-header {
    align-items: stretch;
    flex-wrap: wrap;
  }

  .api-name-input,
  .api-title-section,
  .api-actions {
    flex: 1 1 100%;
    max-width: none;
    min-width: 0;
  }

  .api-title-section,
  .api-actions {
    flex-wrap: wrap;
  }

  .api-method-select {
    flex-basis: 100px;
  }

  .api-actions {
    justify-content: flex-start;
    margin-left: 0;
  }

  .response-view-mode {
    margin-left: 0;
  }

  :deep(.detail-tabs > .el-tabs__header .el-tabs__item) {
    padding: 0 10px;
  }

  :deep(.detail-segmented) {
    width: 100%;
  }
}
</style>




