<template>
  <div class="api-detail" tabindex="0" @keydown="handleKeyDown" @keyup="handleKeyUp">
    <div class="api-header">
      <el-tag type="info" size="small" class="api-id-tag">ID: {{ apiForm.id }}</el-tag>
      <el-input v-model="apiForm.name" class="api-name-input" placeholder="输入接口名称" type="text" @blur="handleBlurSave"></el-input>
      <div class="api-title-section">
        <el-select v-model="apiForm.method" class="api-method-select">
          <el-option label="GET" value="GET"/>
          <el-option label="POST" value="POST"/>
        </el-select>
        <el-input v-model="apiForm.url" class="api-url-input" placeholder="输入请求URL" @blur="handleBlurSave"></el-input>
      </div>

      <div class="api-actions">
        <el-select v-model="envIdString" placeholder="选择环境" class="api-env-select" @change="changeEnv">
          <el-option
              v-for="env in envs"
              :key="env.id"
              :label="env.id === 0 && folderEnvId ? '继承文件夹环境' : env.name"
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

    <el-tabs v-model="configActiveTab" class="detail-tabs api-config-tabs" @tab-change="responseTabChange">
      <el-tab-pane label="备注" name="desc" class="desc-tab-pane">
        <MdEditor class="desc-editor" v-model="apiForm.desc" @blur="handleBlurSave" :onSave="handleSave" />
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
      <el-tab-pane v-if="apiForm.method !== 'GET'" :label="'请求体(' + bodyParamsCount + ')'" name="body">
        <!-- 请求体内容（同原逻辑） -->
        <div class="api-pane-full">
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
            <div class="body-import-toolbar">
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
        <el-alert title="提取示例：data.0.token表示提取data数组的第一个元素中的token" type="info" :closable="false" class="result-tip-alert"/>
        <response-take-editor
            v-if="configActiveTab === 'response_take'"
            v-model="apiForm.response_take_data"
            :envItems="envItems"
            @update="updateResponseTake"
        />
      </el-tab-pane>
      <el-tab-pane v-if="parseInt(apiForm.env_id) > 0 || folderEnvId > 0" :label="'环境变量(' + (isArray(envItems) ? envItems.length : 0) + ')'" lazy name="env_items" class="env-items-tab">
        <div class="config-section" v-if="parseInt(apiForm.env_id) > 0 || folderEnvId > 0">
          <el-alert v-if="!apiForm.env_id && folderEnvId" title="当前环境变量继承自所属文件夹" type="info" :closable="false" class="env-inherit-alert"/>
          <variable-manager
              ref="refVariableManager"
              @update="handleVariablesUpdate"
          />
        </div>
      </el-tab-pane>
      <el-tab-pane label="代码" lazy name="code">
        <div class="api-pane-full">
          <el-radio-group v-model="apiForm.code_type" class="detail-segmented" size="small" @change="handleCodeTypeChange">
            <el-radio-button
                v-for="codeType in codeTypeOptions"
                :key="codeType"
                :value="codeType"
            >
              {{ codeType }}
            </el-radio-button>
          </el-radio-group>

          <div class="response-body-container code-preview">
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
          <div class="response-status-item response-status-item--ok">状态: {{ apiForm.last_result_data.status }}</div>
          <div v-if="apiForm.last_result_data.errmsg" class="response-status-item response-status-item--error">执行错误:
            {{ apiForm.last_result_data.errmsg }}
          </div>
          <div class="response-status-item response-status-item--ok">耗时: {{ apiForm.last_result_data.millisecond }}ms</div>
          <div class="response-status-item response-status-item--ok">时间: {{ apiForm.last_result_data.request_time }}</div>
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
      <el-alert title="请输入JSON对象，键值将自动转换为表单参数。支持嵌套对象（自动展平为 key.subkey 格式）" type="info" :closable="false" class="json-import-alert"/>
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
      folderEnvId: 0,
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
    // 统计请求体参数数量：JSON类型解析key数量，其他类型取form_data长度
    bodyParamsCount() {
      if (this.apiForm.content_type === 'application/json') {
        const jsonData = this.apiForm.body_json_data
        if (typeof jsonData === 'string') {
          try {
            const parsed = JSON.parse(jsonData)
            if (parsed && typeof parsed === 'object' && !Array.isArray(parsed)) {
              return Object.keys(parsed).length
            }
            if (Array.isArray(parsed)) {
              return parsed.length
            }
          } catch (e) {
            return 0
          }
        }
        if (jsonData && typeof jsonData === 'object' && !Array.isArray(jsonData)) {
          return Object.keys(jsonData).length
        }
        return 0
      }
      return this.isArray(this.apiForm.body_form_data) ? this.apiForm.body_form_data.length : 0
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
      _that.folderEnvId = apiInfo.folder_env_id || 0
      // 接口无环境时使用文件夹环境
      let effectiveEnvId = apiInfo.env_id
      if (!effectiveEnvId && _that.folderEnvId) {
        effectiveEnvId = _that.folderEnvId
      }
      _that.loadEnvItems(effectiveEnvId)
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
      // 接口无环境时加载文件夹环境变量
      let effectiveEnvId = env_id || _that.folderEnvId
      _that.loadEnvItems(effectiveEnvId)
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
          name: _that.folderEnvId ? '继承文件夹环境' : '请选择',
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

<style scoped src="@/css/components/api/ApiDetail.css"></style>




