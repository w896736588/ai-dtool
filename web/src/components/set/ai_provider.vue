<template>
  <div class="ai-config">
    <div class="ai-config-header">
      <div class="ai-config-title-row">
        <h2>AI 服务商与模型配置</h2>
        <div style="display:flex;align-items:center;gap:8px;">
          <el-button size="small" @click="OpenLogDialog">请求日志</el-button>
          <el-button type="primary" size="small" @click="ShowAddProvider">新增 Provider</el-button>
        </div>
      </div>
      <p class="ai-config-desc">服务商仅保存基础域名，模型保存具体 URI，并区分 LLM 与嵌入模型</p>
    </div>

    <!-- Provider 卡片列表 -->
    <div class="provider-list">
      <div
        v-for="p in state.providerList"
        :key="p.id"
        class="provider-card"
        :class="{ 'provider-card--expanded': state.expandedProviderIds.has(p.id) }"
      >
        <div class="provider-card__header" @click="toggleExpand(p.id)">
          <div class="provider-card__info">
            <span class="provider-card__name">{{ p.name }}</span>
            <el-tag size="small" effect="plain">{{ FormatProviderType(p.request_format || p.provider_type || 'openai') }}</el-tag>
            <span class="provider-card__url">{{ p.base_url }}</span>
          </div>
          <div class="provider-card__actions" @click.stop>
            <el-button text size="small" @click="ShowEditProvider(p, false)">编辑</el-button>
            <el-button text size="small" @click="ShowEditProvider(p, true)">复制</el-button>
            <el-button text size="small" @click="ShowAddModel(p)">+ 添加模型</el-button>
            <el-button text size="small" type="danger" @click="DeleteProvider(p)">删除</el-button>
            <el-icon class="provider-card__arrow" :class="{ 'provider-card__arrow--open': state.expandedProviderIds.has(p.id) }">
              <ArrowRight />
            </el-icon>
          </div>
        </div>
        <div v-if="state.expandedProviderIds.has(p.id)" class="provider-card__models">
          <el-table :data="getProviderModels(p.id)" class="config-table" empty-text="暂无模型，点击「+ 添加模型」创建">
            <el-table-column prop="name" label="展示名" min-width="130" />
            <el-table-column prop="model" label="模型标识" min-width="180" />
            <el-table-column label="类型" width="80">
              <template #default="{ row }">
                <el-tag size="small" effect="light">{{ row.model_type === 'embedding' ? '嵌入' : 'LLM' }}</el-tag>
              </template>
            </el-table-column>
            <el-table-column label="URI" min-width="200">
              <template #default="{ row }">
                <span class="model-url">{{ BuildRequestUrl(p.base_url, row.uri) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="280" fixed="right">
              <template #default="{ row }">
                <el-button text size="small" @click="ShowEditModel(row, false)">编辑</el-button>
                <el-button text size="small" @click="ShowEditModel(row, true)">复制</el-button>
                <el-button
                  text
                  size="small"
                  type="warning"
                  :loading="Number(state.testingModelId) === Number(row.id)"
                  @click="TestModel(row)"
                >
                  测试
                </el-button>
                <el-button text size="small" type="danger" @click="DeleteModel(row)">删除</el-button>
              </template>
            </el-table-column>
          </el-table>
        </div>
      </div>
      <div v-if="state.providerList.length === 0" class="empty-hint">暂无 Provider，请点击上方按钮添加</div>
    </div>

    <!-- 请求日志弹窗 -->
    <el-dialog v-model="state.dialogLog" title="请求日志" width="900" @opened="LoadRequestLogList">
      <div class="log-filters" style="margin-bottom:12px;">
        <el-select
          v-model="state.logProviderId"
          size="small"
          style="width: 180px;"
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
          size="small"
          style="width: 130px;"
          placeholder="模型类型"
          clearable
          @change="LoadRequestLogList"
        >
          <el-option label="LLM" value="llm"/>
          <el-option label="嵌入模型" value="embedding"/>
        </el-select>
        <el-button size="small" @click="LoadRequestLogList">刷新</el-button>
      </div>

      <el-table :data="state.requestLogList" class="config-table" row-key="id" :max-height="500" empty-text="暂无请求日志">
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
        <el-table-column prop="model_type" label="类型" width="80">
          <template #default="scope">
            <el-tag size="small" effect="light">{{ scope.row.model_type === 'embedding' ? '嵌入' : 'LLM' }}</el-tag>
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
            <el-button text size="small" @click="ShowRequestLogDetail(scope.row)">详情</el-button>
          </template>
        </el-table-column>
      </el-table>
    </el-dialog>

    <!-- Provider 编辑对话框 -->
    <el-dialog v-model="state.dialogProvider" title="编辑 Provider" width="560">
      <el-form label-width="100px">
        <el-form-item label="名称">
          <el-input v-model="state.editProvider.name" autocomplete="off" placeholder="如 OpenAI、DeepSeek"/>
        </el-form-item>
        <el-form-item label="请求格式">
          <el-select v-model="state.editProvider.request_format" style="width: 100%;">
            <el-option label="OpenAI Chat Completions" value="openai"/>
            <el-option label="OpenAI Responses" value="openai-responses"/>
            <el-option label="Anthropic Messages" value="anthropic"/>
            <el-option label="DeepSeek (OpenAI兼容)" value="deepseek"/>
          </el-select>
        </el-form-item>
        <el-form-item label="基础域名">
          <el-input
            v-model="state.editProvider.base_url"
            autocomplete="off"
            placeholder="https://api.openai.com"
          />
        </el-form-item>
        <el-form-item label="API Key">
          <el-input
            v-model="state.editProvider.api_key"
            :type="state.showApiKey ? 'text' : 'password'"
            autocomplete="off"
            :placeholder="Number(state.editProvider.id) > 0 && !state.apiKeyFetched ? '已保存（点击右侧眼睛查看）' : '请输入 API Key'"
          >
            <template #suffix>
              <el-icon
                class="api-key-eye"
                :class="{ 'api-key-eye--loading': state.apiKeyLoading }"
                :title="state.showApiKey ? '隐藏' : '显示'"
                @click="ToggleApiKey"
              >
                <View v-if="state.showApiKey" />
                <Hide v-else />
              </el-icon>
            </template>
          </el-input>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="state.dialogProvider = false">取消</el-button>
          <el-button type="primary" @click="SaveProvider">保存</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 模型编辑对话框 -->
    <el-dialog v-model="state.dialogModel" title="编辑模型" width="560">
      <el-form label-width="100px">
        <el-form-item label="所属 Provider">
          <el-select v-model="state.editModel.provider_id" style="width: 100%;" :disabled="state.editModel.id > 0">
            <template v-for="(provider, idx) in state.providerList" :key="idx">
              <el-option :label="provider.name + ' (' + (provider.request_format || provider.provider_type || 'openai') + ')'" :value="provider.id"/>
            </template>
          </el-select>
        </el-form-item>
        <el-form-item label="展示名称">
          <el-input v-model="state.editModel.name" autocomplete="off" placeholder="如 GPT-4o Mini"/>
        </el-form-item>
        <el-form-item label="模型类型">
          <el-select v-model="state.editModel.model_type" style="width: 100%;" @change="HandleModelTypeChange">
            <el-option label="LLM" value="llm"/>
            <el-option label="嵌入模型" value="embedding"/>
          </el-select>
        </el-form-item>
        <el-form-item label="模型标识">
          <el-input v-model="state.editModel.model" autocomplete="off" placeholder="如 gpt-4o-mini"/>
        </el-form-item>
        <el-form-item label="URI">
          <el-input v-model="state.editModel.uri" autocomplete="off" :placeholder="DefaultUriForProvider(getProviderType(state.editModel.provider_id), state.editModel.model_type || 'llm')"/>
          <div class="field-hint">留空自动使用默认 URI</div>
        </el-form-item>
        <el-form-item label="完整地址预览">
          <div class="url-preview">{{ BuildRequestUrl(CurrentEditProviderBaseURL(), state.editModel.uri) || '-' }}</div>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="state.dialogModel = false">取消</el-button>
          <el-button type="primary" @click="SaveModel">保存</el-button>
        </div>
      </template>
    </el-dialog>

    <!-- 请求日志详情 -->
    <el-dialog v-model="state.dialogLogDetail" title="请求日志详情" width="700">
      <el-descriptions :column="2" border size="small">
        <el-descriptions-item label="服务商">{{ state.logDetail.provider_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="模型">{{ state.logDetail.model_name || '-' }}</el-descriptions-item>
        <el-descriptions-item label="模型标识">{{ state.logDetail.model || '-' }}</el-descriptions-item>
        <el-descriptions-item label="模型类型">{{ state.logDetail.model_type === 'embedding' ? '嵌入模型' : 'LLM' }}</el-descriptions-item>
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
          <el-button @click="state.dialogLogDetail = false">关闭</el-button>
        </div>
      </template>
    </el-dialog>
  </div>
</template>

<script>
import {defineComponent, getCurrentInstance, reactive} from 'vue'
import {ArrowRight, View, Hide, Loading} from '@element-plus/icons-vue'
import common from '@/utils/common'
import aiSet from '@/utils/base/ai_set'

export default defineComponent({
  components: {ArrowRight, View, Hide, Loading},
  setup() {
    const proxy = getCurrentInstance().proxy
    const instance = getCurrentInstance().appContext.config.globalProperties

    const state = reactive({
      providerList: [],
      allModels: [], // 所有 Provider 的全部模型
      expandedProviderIds: new Set(),
      dialogProvider: false,
      dialogModel: false,
      testingModelId: 0,
      // API Key 显隐控制
      showApiKey: false,        // 是否明文显示
      apiKeyFetched: false,     // 是否已拉取真实密钥（区分“未改动”与“已查看”）
      apiKeyLoading: false,     // 拉取真实密钥中
      editProvider: {},
      editModel: {},
      // 请求日志相关
      dialogLog: false,
      requestLogList: [],
      logProviderId: null,
      logModelType: '',
      dialogLogDetail: false,
      logDetail: {},
    })

    // --- 工具函数 ---

    const FormatProviderType = function (t){
      switch (String(t || '').toLowerCase()) {
        case 'openai': return 'OpenAI Chat Completions'
        case 'openai-responses': return 'OpenAI Responses'
        case 'anthropic': return 'Anthropic Messages'
        case 'deepseek': return 'DeepSeek (OpenAI兼容)'
        default: return t || '-'
      }
    }

    const NormalizeUri = function (uri){
      const str = String(uri || '').trim()
      if(str === '') return ''
      return str.startsWith('/') ? str : '/' + str
    }

    const BuildRequestUrl = function (baseUrl, uri){
      const cleanBase = String(baseUrl || '').trim().replace(/\/+$/, '')
      const cleanUri = NormalizeUri(uri)
      if(cleanBase === '') return cleanUri
      if(cleanUri === '') return cleanBase
      return cleanBase + cleanUri
    }

    const DefaultUriForProvider = function (providerType, modelType){
      const mt = String(modelType || 'llm').toLowerCase()
      switch (String(providerType || '').toLowerCase()) {
        case 'openai':
        case 'deepseek':
          return mt === 'embedding' ? '/v1/embeddings' : '/v1/chat/completions'
        case 'openai-responses':
          return mt === 'embedding' ? '/v1/embeddings' : '/v1/responses'
        case 'anthropic':
          return mt === 'embedding' ? '/v1/embeddings' : '/v1/messages'
        default:
          return mt === 'embedding' ? '/v1/embeddings' : '/v1/chat/completions'
      }
    }

    const NormalizeModelRow = function (item){
      return {...item, model_type: item.model_type || 'llm', uri: NormalizeUri(item.uri || '')}
    }

    const getProviderType = function (providerId){
      const p = (state.providerList || []).find(function (item){ return Number(item.id) === Number(providerId) })
      return p ? (p.request_format || p.provider_type || 'openai') : 'openai'
    }

    const CurrentEditProviderBaseURL = function (){
      const provider = (state.providerList || []).find(function (item){ return Number(item.id) === Number(state.editModel.provider_id) })
      return provider ? provider.base_url : ''
    }

    const getProviderModels = function (providerId){
      return (state.allModels || []).filter(function (m){ return Number(m.provider_id) === Number(providerId) })
    }

    // --- 展开/折叠 ---

    const toggleExpand = function (providerId){
      if(state.expandedProviderIds.has(providerId)){
        state.expandedProviderIds.delete(providerId)
      } else {
        state.expandedProviderIds.add(providerId)
      }
    }

    // --- 数据加载 ---

    const LoadProviderList = function (){
      aiSet.AiProviderList(function (response){
        if(response.ErrCode === 0){
          state.providerList = (response.Data || []).map(function (item){
            return {...item, request_format: item.request_format || item.provider_type || 'openai'}
          })
          // 默认展开全部 Provider
          ;(response.Data || []).forEach(function (item){
            state.expandedProviderIds.add(item.id)
          })
          LoadAllModels()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
      })
    }

    const LoadAllModels = function (){
      aiSet.AiModelList({}, function (response){
        if(response.ErrCode === 0){
          state.allModels = (response.Data || []).map(NormalizeModelRow)
        }
      })
    }

    // --- Provider 操作 ---

    const ShowAddProvider = function (){
      state.editProvider = {request_format: 'openai', api_key: ''}
      state.showApiKey = false
      state.apiKeyFetched = false
      state.apiKeyLoading = false
      state.dialogProvider = true
    }

    const ShowEditProvider = function (row, isCopy){
      state.editProvider = {...row, request_format: row.request_format || row.provider_type || 'openai'}
      // 编辑已有服务商时，API Key 保留后端返回的脱敏值用于 * 号展示；
      // 复制时视为新建，不携带原密钥。
      if(isCopy) state.editProvider.id = 0
      if(Number(state.editProvider.id) === 0) state.editProvider.api_key = ''
      state.showApiKey = false
      state.apiKeyFetched = false
      state.apiKeyLoading = false
      state.dialogProvider = true
    }

    // 切换 API Key 显隐：编辑已有服务商且尚未拉取真实密钥时，先请求明文再显示。
    const ToggleApiKey = function (){
      if(state.showApiKey){
        state.showApiKey = false
        return
      }
      if(Number(state.editProvider.id) > 0 && !state.apiKeyFetched){
        state.apiKeyLoading = true
        aiSet.AiProviderKeyGet({id: state.editProvider.id}, function (res){
          state.apiKeyLoading = false
          if(res.ErrCode === 0){
            state.editProvider.api_key = res.Data.api_key || ''
            state.apiKeyFetched = true
            state.showApiKey = true
          }else{
            instance.$helperNotify.error(res.ErrMsg)
          }
        })
      }else{
        state.showApiKey = true
      }
    }

    const SaveProvider = function (){
      const submitData = {...state.editProvider, request_format: state.editProvider.request_format || 'openai'}
      // 编辑已有服务商且未查看/未修改 API Key 时，不提交该字段，
      // 避免用脱敏值（如 sk-****1234）覆盖数据库中保存的真实密钥。
      if(Number(state.editProvider.id) > 0 && !state.apiKeyFetched){
        delete submitData.api_key
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
            state.expandedProviderIds.delete(row.id)
            LoadProviderList()
          }else{
            instance.$helperNotify.error(response.ErrMsg)
          }
        })
      })
    }

    // --- 模型操作 ---

    const ShowAddModel = function (provider){
      const providerType = (provider.request_format || provider.provider_type || 'openai')
      state.editModel = {
        provider_id: provider.id,
        model_type: 'llm',
        uri: DefaultUriForProvider(providerType, 'llm'),
      }
      state.dialogModel = true
    }

    const HandleModelTypeChange = function (newModelType){
      const pt = getProviderType(state.editModel.provider_id)
      state.editModel.uri = DefaultUriForProvider(pt, newModelType)
    }

    const ShowEditModel = function (row, isCopy){
      state.editModel = NormalizeModelRow({...row})
      if(isCopy) state.editModel.id = 0
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
          LoadAllModels()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
      })
    }

    const DeleteModel = function (row){
      common.ConfirmProxyDelete(proxy, function (){
        aiSet.AiModelDelete(row, function (response){
          if(response.ErrCode === 0){
            LoadAllModels()
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

    // --- 请求日志 ---

    const OpenLogDialog = function (){
      state.dialogLog = true
    }

    const LoadRequestLogList = function (){
      const params = {limit: 100}
      if(state.logProviderId) params.provider_id = state.logProviderId
      if(state.logModelType) params.model_type = state.logModelType
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
      if(!str) return ''
      if(typeof str === 'object') return JSON.stringify(str, null, 2)
      try{ return JSON.stringify(JSON.parse(str), null, 2) }catch(e){ return String(str) }
    }

    LoadProviderList()

    return {
      state,
      FormatProviderType,
      BuildRequestUrl,
      DefaultUriForProvider,
      HandleModelTypeChange,
      CurrentEditProviderBaseURL,
      getProviderType,
      getProviderModels,
      toggleExpand,
      LoadProviderList,
      LoadAllModels,
      OpenLogDialog,
      ShowAddProvider,
      ShowEditProvider,
      ToggleApiKey,
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

<style scoped src="@/css/components/set/ai_provider.css"></style>

