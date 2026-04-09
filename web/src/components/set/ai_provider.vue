<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">AI 服务商与模型配置</h3>
      <p class="set-config-desc">服务商仅保存基础域名，模型保存具体 URI，并区分 LLM 与嵌入模型</p>
    </div>

    <el-tabs v-model="state.activeTab" class="set-config-inner-tabs" @tab-change="HandleInnerTabChange">
      <el-tab-pane label="服务商配置" name="provider">
        <div class="set-config-actions" style="margin-bottom: 10px;">
          <pl-button type="primary" @click="ShowAddProvider">新增服务商</pl-button>
        </div>
        <div class="set-config-table-card">
          <el-table :data="state.providerList" class="set-config-table" row-key="id">
            <el-table-column prop="id" label="#id" width="70"/>
            <el-table-column prop="name" label="服务商名称" min-width="160"/>
            <el-table-column prop="request_format" label="请求格式" width="140"/>
            <el-table-column prop="base_url" label="基础域名" min-width="220"/>
            <el-table-column prop="api_key" label="API Key" min-width="160">
              <template #default="scope">
                <span>{{ MaskKey(scope.row.api_key) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="380">
              <template #default="scope">
                <div class="set-op-group">
                  <pl-button size="small" type="success" plain @click="ShowEditProvider(scope.row, false)">编辑</pl-button>
                  <pl-button size="small" type="success" plain @click="ShowEditProvider(scope.row, true)">复制新增</pl-button>
                  <pl-button size="small" type="success" plain @click="SwitchToModelTab(scope.row)">管理模型</pl-button>
                  <pl-button size="small" type="danger" plain @click="DeleteProvider(scope.row)">删除</pl-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <el-tab-pane label="模型配置" name="model">
        <div class="set-config-actions model-actions">
          <el-select
            v-model="state.currentProviderId"
            style="width: 280px;"
            placeholder="请选择服务商"
            @change="LoadModelList"
          >
            <template v-for="(provider, idx) in state.providerList" :key="idx">
              <el-option :label="provider.name + ' (' + provider.request_format + ')'" :value="provider.id"/>
            </template>
          </el-select>
          <pl-button type="primary" :disabled="!state.currentProviderId" @click="ShowAddModel">新增模型</pl-button>
        </div>

        <div class="set-config-table-card">
          <el-table :data="state.modelList" class="set-config-table" row-key="id">
            <el-table-column prop="id" label="#id" width="70"/>
            <el-table-column prop="provider_name" label="所属服务商" min-width="150"/>
            <el-table-column prop="name" label="展示名" min-width="170"/>
            <el-table-column prop="model_type" label="模型类型" width="120">
              <template #default="scope">
                <el-tag size="small" effect="light">{{ ModelTypeLabel(scope.row.model_type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="model" label="模型标识" min-width="180"/>
            <el-table-column prop="uri" label="URI" min-width="190"/>
            <el-table-column label="完整地址" min-width="260">
              <template #default="scope">
                <span>{{ BuildRequestUrl(scope.row.base_url, scope.row.uri) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="380">
              <template #default="scope">
                <div class="set-op-group">
                  <pl-button size="small" type="success" plain @click="ShowEditModel(scope.row, false)">编辑</pl-button>
                  <pl-button size="small" type="success" plain @click="ShowEditModel(scope.row, true)">复制新增</pl-button>
                  <pl-button
                    size="small"
                    type="success"
                    plain
                    :loading="Number(state.testingModelId) === Number(scope.row.id)"
                    @click="TestModel(scope.row)"
                  >
                    测试
                  </pl-button>
                  <pl-button size="small" type="danger" plain @click="DeleteModel(scope.row)">删除</pl-button>
                </div>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>

      <el-tab-pane label="请求日志" name="log">
        <div class="set-config-actions log-actions">
          <el-select
            v-model="state.logProviderId"
            style="width: 200px;"
            placeholder="筛选服务商"
            clearable
            @change="LoadRequestLogList"
          >
            <template v-for="(provider, idx) in state.providerList" :key="idx">
              <el-option :label="provider.name" :value="provider.id"/>
            </template>
          </el-select>
          <el-select
            v-model="state.logModelType"
            style="width: 140px;"
            placeholder="模型类型"
            clearable
            @change="LoadRequestLogList"
          >
            <el-option label="LLM" value="llm"/>
            <el-option label="嵌入模型" value="embedding"/>
          </el-select>
          <pl-button @click="LoadRequestLogList">刷新</pl-button>
        </div>

        <div class="set-config-table-card">
          <el-table :data="state.requestLogList" class="set-config-table" row-key="id" :max-height="500">
            <el-table-column prop="id" label="#id" width="70"/>
            <el-table-column prop="provider_name" label="服务商" min-width="120"/>
            <el-table-column prop="model_name" label="模型" min-width="140">
              <template #default="scope">
                <div>
                  <div>{{ scope.row.model_name || '-' }}</div>
                  <div class="log-model-id">{{ scope.row.model }}</div>
                </div>
              </template>
            </el-table-column>
            <el-table-column prop="model_type" label="类型" width="90">
              <template #default="scope">
                <el-tag size="small" effect="light">{{ ModelTypeLabel(scope.row.model_type) }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="input_tokens" label="输入Token" width="100" align="right"/>
            <el-table-column prop="output_tokens" label="输出Token" width="100" align="right"/>
            <el-table-column prop="cost_time_desc" label="耗时" width="90" align="right"/>
            <el-table-column prop="response_status_code" label="状态" width="70" align="center">
              <template #default="scope">
                <el-tag
                  size="small"
                  :type="scope.row.success === 1 ? 'success' : 'danger'"
                  effect="light"
                >
                  {{ scope.row.response_status_code || (scope.row.success === 1 ? '200' : 'err') }}
                </el-tag>
              </template>
            </el-table-column>
            <el-table-column prop="create_time_desc" label="时间" width="160"/>
            <el-table-column label="操作" width="80" fixed="right">
              <template #default="scope">
                <pl-button size="small" type="success" plain @click="ShowRequestLogDetail(scope.row)">详情</pl-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </el-tab-pane>
    </el-tabs>

    <el-dialog v-model="state.dialogProvider" title="编辑服务商配置" width="560">
      <el-form label-width="100px">
        <el-form-item label="服务商名称">
          <el-input v-model="state.editProvider.name" autocomplete="off"/>
        </el-form-item>
        <el-form-item label="请求格式">
          <el-select v-model="state.editProvider.request_format" style="width: 100%;">
            <el-option label="openai" value="openai"/>
          </el-select>
        </el-form-item>
        <el-form-item label="基础域名">
          <el-input
            v-model="state.editProvider.base_url"
            autocomplete="off"
            placeholder="例如: https://api.openai.com"
          />
        </el-form-item>
        <el-form-item label="API Key">
          <el-input v-model="state.editProvider.api_key" type="password" show-password autocomplete="off"/>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogProvider = false">取消</pl-button>
          <pl-button type="primary" @click="SaveProvider">保存</pl-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="state.dialogModel" title="编辑模型配置" width="560">
      <el-form label-width="100px">
        <el-form-item label="所属服务商">
          <el-select v-model="state.editModel.provider_id" style="width: 100%;" :disabled="state.editModel.id > 0">
            <template v-for="(provider, idx) in state.providerList" :key="idx">
              <el-option :label="provider.name + ' (' + provider.request_format + ')'" :value="provider.id"/>
            </template>
          </el-select>
        </el-form-item>
        <el-form-item label="展示名称">
          <el-input v-model="state.editModel.name" autocomplete="off"/>
        </el-form-item>
        <el-form-item label="模型类型">
          <el-select v-model="state.editModel.model_type" style="width: 100%;">
            <el-option label="LLM" value="llm"/>
            <el-option label="嵌入模型" value="embedding"/>
          </el-select>
        </el-form-item>
        <el-form-item label="模型标识">
          <el-input v-model="state.editModel.model" autocomplete="off" placeholder="例如: gpt-4o-mini"/>
        </el-form-item>
        <el-form-item label="URI">
          <el-input v-model="state.editModel.uri" autocomplete="off" placeholder="例如: /v1/chat/completions"/>
        </el-form-item>
        <el-form-item label="完整地址预览">
          <div class="request-preview">{{ BuildRequestUrl(CurrentEditProviderBaseURL(), state.editModel.uri) || '-' }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogModel = false">取消</pl-button>
          <pl-button type="primary" @click="SaveModel">保存</pl-button>
        </div>
      </template>
    </el-dialog>

    <el-dialog v-model="state.dialogLogDetail" title="请求日志详情" width="700">
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="服务商">{{ state.logDetail.provider_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="模型">{{ state.logDetail.model_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="模型标识">{{ state.logDetail.model || '-' }}</el-descriptions-item>
        <el-descriptions-item label="模型类型">{{ ModelTypeLabel(state.logDetail.model_type) }}</el-descriptions-item>
        <el-descriptions-item label="输入Token">{{ state.logDetail.input_tokens || 0 }}</el-descriptions-item>
        <el-descriptions-item label="输出Token">{{ state.logDetail.output_tokens || 0 }}</el-descriptions-item>
        <el-descriptions-item label="耗时">{{ state.logDetail.cost_time_desc || '-' }}</el-descriptions-item>
        <el-descriptions-item label="状态码">{{ state.logDetail.response_status_code || '-' }}</el-descriptions-item>
        <el-descriptions-item label="请求地址" :span="2">{{ state.logDetail.request_url || '-' }}</el-descriptions-item>
        <el-descriptions-item label="时间" :span="2">{{ state.logDetail.create_time_desc || '-' }}</el-descriptions-item>
        <el-descriptions-item label="错误信息" :span="2">
          <span v-if="state.logDetail.success === 1">-</span>
          <span v-else class="error-text">{{ state.logDetail.error_message || '-' }}</span>
        </el-descriptions-item>
      </el-descriptions>

      <el-divider content-position="left">请求参数</el-divider>
      <pre class="json-preview">{{ FormatJson(state.logDetail.request_params) }}</pre>

      <el-divider content-position="left">响应内容</el-divider>
      <pre class="json-preview">{{ FormatJson(state.logDetail.response_body) }}</pre>

      <template #footer>
        <div class="dialog-footer">
          <pl-button @click="state.dialogLogDetail = false">关闭</pl-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import {defineComponent, getCurrentInstance, reactive} from 'vue'
import common from '@/utils/common'
import aiSet from '@/utils/base/ai_set'

export default defineComponent({
  setup() {
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties

    const state = reactive({
      activeTab: 'provider',
      providerList: [],
      currentProviderId: 0,
      modelList: [],
      dialogProvider: false,
      dialogModel: false,
      testingModelId: 0,
      editProvider: {},
      editModel: {},
      // 请求日志相关
      requestLogList: [],
      logProviderId: null,
      logModelType: '',
      dialogLogDetail: false,
      logDetail: {},
    })

    const MaskKey = function (key){
      const str = String(key || '')
      if(str.length <= 6){
        return str === '' ? '' : '******'
      }
      return str.slice(0, 3) + '******' + str.slice(-3)
    }

    const NormalizeUri = function (uri){
      const str = String(uri || '').trim()
      if(str === ''){
        return ''
      }
      return str.startsWith('/') ? str : '/' + str
    }

    const BuildRequestUrl = function (baseUrl, uri){
      const cleanBase = String(baseUrl || '').trim().replace(/\/+$/, '')
      const cleanUri = NormalizeUri(uri)
      if(cleanBase === ''){
        return cleanUri
      }
      if(cleanUri === ''){
        return cleanBase
      }
      return cleanBase + cleanUri
    }

    const ModelTypeLabel = function (modelType){
      return String(modelType || 'llm') === 'embedding' ? '嵌入模型' : 'LLM'
    }

    const EnsureCurrentProvider = function (){
      if(state.providerList.length === 0){
        state.currentProviderId = 0
        return
      }
      const exists = state.providerList.some(function (item){
        return Number(item.id) === Number(state.currentProviderId)
      })
      if(!exists){
        state.currentProviderId = state.providerList[0].id
      }
    }

    const CurrentEditProviderBaseURL = function (){
      const provider = (state.providerList || []).find(function (item){
        return Number(item.id) === Number(state.editModel.provider_id)
      })
      return provider ? provider.base_url : ''
    }

    const NormalizeModelRow = function (item){
      return {
        ...item,
        model_type: item.model_type || 'llm',
        uri: NormalizeUri(item.uri || ''),
      }
    }

    const LoadProviderList = function (){
      aiSet.AiProviderList(function (response){
        if(response.ErrCode === 0){
          state.providerList = (response.Data || []).map(function (item){
            return {
              ...item,
              request_format: item.request_format || item.provider_type || 'openai',
            }
          })
          EnsureCurrentProvider()
          if(state.activeTab === 'model'){
            LoadModelList()
          }
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
      })
    }

    const LoadModelList = function (){
      if(!state.currentProviderId){
        state.modelList = []
        return
      }
      aiSet.AiModelList({provider_id: state.currentProviderId}, function (response){
        if(response.ErrCode === 0){
          state.modelList = (response.Data || []).map(NormalizeModelRow)
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
      })
    }

    const SwitchToModelTab = function (provider){
      state.currentProviderId = provider.id
      state.activeTab = 'model'
      LoadModelList()
    }

    const HandleInnerTabChange = function (tabName){
      if(String(tabName) === 'model'){
        LoadModelList()
      } else if(String(tabName) === 'log'){
        LoadRequestLogList()
      }
    }

    const ShowAddProvider = function (){
      state.editProvider = {
        request_format: 'openai',
      }
      state.dialogProvider = true
    }

    const ShowEditProvider = function (row, isCopy){
      state.editProvider = {
        ...row,
        request_format: row.request_format || row.provider_type || 'openai',
      }
      if(isCopy){
        state.editProvider.id = 0
      }
      state.dialogProvider = true
    }

    const SaveProvider = function (){
      const submitData = {
        ...state.editProvider,
        request_format: state.editProvider.request_format || 'openai',
      }
      aiSet.AiProviderAdd(submitData, function (response){
        if(response.ErrCode === 0){
          state.dialogProvider = false
          LoadProviderList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
      })
    }

    const DeleteProvider = function (row){
      common.ConfirmProxyDelete(proxy, function (){
        aiSet.AiProviderDelete(row, function (response){
          if(response.ErrCode === 0){
            LoadProviderList()
            LoadModelList()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
        })
      })
    }

    const ShowAddModel = function (){
      if(!state.currentProviderId){
        instance.$helperNotify.error('请先选择服务商')
        return
      }
      state.editModel = {
        provider_id: state.currentProviderId,
        model_type: 'llm',
        uri: '/v1/chat/completions',
      }
      state.dialogModel = true
    }

    const ShowEditModel = function (row, isCopy){
      state.editModel = NormalizeModelRow({
        ...row,
      })
      if(isCopy){
        state.editModel.id = 0
      }
      state.dialogModel = true
    }

    const SaveModel = function (){
      const submitData = {
        ...state.editModel,
        model_type: state.editModel.model_type || 'llm',
        uri: NormalizeUri(state.editModel.uri),
      }
      aiSet.AiModelAdd(submitData, function (response){
        if(response.ErrCode === 0){
          state.dialogModel = false
          LoadModelList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
      })
    }

    const DeleteModel = function (row){
      common.ConfirmProxyDelete(proxy, function (){
        aiSet.AiModelDelete(row, function (response){
          if(response.ErrCode === 0){
            LoadModelList()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
        })
      })
    }

    const TestModel = function (row){
      state.testingModelId = row.id
      aiSet.AiModelTest({id: row.id}, function (response){
        state.testingModelId = 0
        if(response.ErrCode === 0){
          instance.$helperNotify.success((row.name || row.model || '模型') + ' 连通成功')
        }else{
          instance.$helperNotify.error(response.ErrMsg || '连通失败')
        }
      })
    }

    const LoadRequestLogList = function (){
      const params = {
        limit: 100,
      }
      if(state.logProviderId){
        params.provider_id = state.logProviderId
      }
      if(state.logModelType){
        params.model_type = state.logModelType
      }
      aiSet.AiRequestLogList(params, function (response){
        if(response.ErrCode === 0){
          state.requestLogList = response.Data || []
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
      })
    }

    const ShowRequestLogDetail = function (row){
      state.logDetail = {...row}
      state.dialogLogDetail = true
    }

    const FormatJson = function (str){
      if(!str){
        return ''
      }
      if(typeof str === 'object'){
        return JSON.stringify(str, null, 2)
      }
      try{
        return JSON.stringify(JSON.parse(str), null, 2)
      }catch(e){
        return String(str)
      }
    }

    LoadProviderList()

    return {
      state,
      MaskKey,
      BuildRequestUrl,
      ModelTypeLabel,
      CurrentEditProviderBaseURL,
      LoadProviderList,
      LoadModelList,
      SwitchToModelTab,
      HandleInnerTabChange,
      ShowAddProvider,
      ShowEditProvider,
      SaveProvider,
      DeleteProvider,
      ShowAddModel,
      ShowEditModel,
      SaveModel,
      DeleteModel,
      TestModel,
      LoadRequestLogList,
      ShowRequestLogDetail,
      FormatJson,
    }
  },
})
</script>

<style scoped>
@import "@/css/set_module_unified.css";

.model-actions {
  margin-bottom: 10px;
  gap: 10px;
}

.request-preview {
  width: 100%;
  min-height: 32px;
  padding: 6px 10px;
  border: 1px solid #dcdfe6;
  border-radius: 6px;
  background: #f8faf6;
  color: #53624d;
  line-height: 1.5;
  word-break: break-all;
}

.set-config-inner-tabs :deep(.el-tabs__header) {
  margin-bottom: 10px;
}

.set-config-page :deep(.el-button--primary),
.set-config-page :deep(.el-button--primary.is-plain) {
  border-color: #d8ded2 !important;
  background: #f6f8f3 !important;
  color: #4f804f !important;
}

.set-config-page :deep(.el-button--primary:hover),
.set-config-page :deep(.el-button--primary.is-plain:hover) {
  border-color: #bfd1bf !important;
  background: #eef4ea !important;
  color: #3f6f3f !important;
}

.log-actions {
  margin-bottom: 10px;
  gap: 10px;
}

.log-model-id {
  font-size: 12px;
  color: #999;
}

.json-preview {
  background: #f5f7fa;
  padding: 12px;
  border-radius: 4px;
  max-height: 300px;
  overflow: auto;
  font-size: 12px;
  line-height: 1.5;
  white-space: pre-wrap;
  word-break: break-all;
}

.error-text {
  color: #f56c6c;
}

</style>

