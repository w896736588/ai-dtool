<template>
  <div class="provider-model-panel">
    <!-- 头部 -->
    <div v-if="mode === 'set'" class="pmp-header">
      <div class="pmp-title-row">
        <h2>AI 服务商与模型配置</h2>
        <div style="display:flex;align-items:center;gap:8px;">
          <el-button v-if="showRequestLog" size="small" @click="OpenLogDialog">请求日志</el-button>
          <el-button type="primary" size="small" @click="ShowAddProvider">新增 Provider</el-button>
        </div>
      </div>
      <p class="pmp-desc">服务商仅保存基础域名，模型保存具体 URI，并区分 LLM 与嵌入模型</p>
    </div>
    <div v-else class="pmp-header">
      <el-button type="primary" size="small" @click="ShowAddProvider">新增 Provider</el-button>
      <span class="pmp-hint">管理所有 LLM 服务商及其模型</span>
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
            <el-tag size="small" effect="plain">{{ formatProviderType(p.request_format || p.provider_type || 'openai') }}</el-tag>
            <span class="provider-card__url">{{ p.base_url }}</span>
          </div>
          <div class="provider-card__actions" @click.stop>
            <el-button text size="small" @click="ShowEditProvider(p, false)">编辑</el-button>
            <el-button v-if="showCopy" text size="small" @click="ShowEditProvider(p, true)">复制</el-button>
            <el-button text size="small" @click="ShowAddModel(p)">+ 添加模型</el-button>
            <el-button text size="small" type="danger" @click="DeleteProvider(p)">删除</el-button>
            <el-icon class="provider-card__arrow" :class="{ 'provider-card__arrow--open': state.expandedProviderIds.has(p.id) }">
              <ArrowRight />
            </el-icon>
          </div>
        </div>
        <div v-if="state.expandedProviderIds.has(p.id)" class="provider-card__models">
          <el-table :data="getProviderModels(p.id)" class="config-table" empty-text="暂无模型">
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
            <el-table-column v-if="mode === 'agent'" label="上下文窗口" width="110">
              <template #default="{ row }">{{ fmtCtx(row.context_size) }}</template>
            </el-table-column>
            <el-table-column :width="showCopy ? 280 : 200" label="操作" fixed="right">
              <template #default="{ row }">
                <el-button text size="small" @click="ShowEditModel(row, false)">编辑</el-button>
                <el-button v-if="showCopy" text size="small" @click="ShowEditModel(row, true)">复制</el-button>
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

    <!-- 请求日志弹窗（仅 set 模式） -->
    <el-dialog v-if="showRequestLog" v-model="state.dialogLog" title="请求日志" width="900" @opened="LoadRequestLogList">
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
            <el-tag size="small" :type="scope.row.success === 1 ? 'success' : 'danger'" effect="light">
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
    <el-dialog v-model="state.dialogProvider" :title="state.editProvider.id > 0 ? '编辑 Provider' : '新增 Provider'" width="560">
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
            <el-option label="Google Generative AI" value="google"/>
          </el-select>
          <div class="field-hint">选择 API 的请求格式，决定 Pi 调用时的 endpoint 路径</div>
        </el-form-item>
        <el-form-item label="基础域名">
          <el-input v-model="state.editProvider.base_url" autocomplete="off" placeholder="https://api.openai.com"/>
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
                <Loading v-if="state.apiKeyLoading"/>
                <View v-else-if="state.showApiKey"/>
                <Hide v-else/>
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
    <el-dialog v-model="state.dialogModel" :title="state.editModel.id > 0 ? '编辑模型' : '新增模型'" width="560">
      <el-form label-width="100px">
        <el-form-item label="所属 Provider">
          <el-select v-model="state.editModel.provider_id" style="width: 100%;" :disabled="state.editModel.id > 0">
            <template v-for="(provider, idx) in state.providerList" :key="idx">
              <el-option
                :label="provider.name + ' (' + formatProviderType(provider.request_format || provider.provider_type || 'openai') + ')'"
                :value="provider.id"
              />
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
        <el-form-item v-if="mode === 'agent'" label="上下文窗口">
          <el-input-number v-model="state.editModel.context_size" :min="1000" :max="4194304" :step="1000" style="width:100%"/>
          <div class="field-hint">以 token 为单位的最大上下文窗口大小</div>
        </el-form-item>
        <el-form-item label="URI">
          <el-input
            v-model="state.editModel.uri"
            autocomplete="off"
            :placeholder="DefaultUriForProvider(getProviderType(state.editModel.provider_id), state.editModel.model_type || 'llm')"
          />
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

    <!-- 请求日志详情（仅 set 模式） -->
    <el-dialog v-if="showRequestLog" v-model="state.dialogLogDetail" title="请求日志详情" width="700">
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
import { defineComponent, reactive, getCurrentInstance } from 'vue'
import { ArrowRight, View, Hide, Loading } from '@element-plus/icons-vue'
import { ElMessage, ElNotification } from 'element-plus'
import common from '@/utils/common'
import aiSet from '@/utils/base/ai_set'
import Base from '@/utils/base.js'

export default defineComponent({
  name: 'ProviderModelPanel',
  components: { ArrowRight, View, Hide, Loading },
  props: {
    // 'agent'：走 AgentV2 读写/测试接口；'set'：走 Set 全局配置接口
    mode: { type: String, default: 'set' },
    // 是否显示「复制」按钮（Set 页特有）
    showCopy: { type: Boolean, default: false },
    // 是否显示「请求日志」入口与弹窗（Set 页特有）
    showRequestLog: { type: Boolean, default: false },
  },
  setup(props, { expose }) {
    const { proxy } = getCurrentInstance()

    const state = reactive({
      providerList: [],
      allModels: [], // 所有 Provider 的全部模型
      expandedProviderIds: new Set(),
      dialogProvider: false,
      dialogModel: false,
      testingModelId: 0,
      // API Key 显隐控制
      showApiKey: false,
      apiKeyFetched: false,
      apiKeyLoading: false,
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
    const formatProviderType = common.formatProviderType

    const NormalizeUri = function (uri) {
      const str = String(uri || '').trim()
      if (str === '') return ''
      return str.startsWith('/') ? str : '/' + str
    }

    const BuildRequestUrl = function (baseUrl, uri) {
      const cleanBase = String(baseUrl || '').trim().replace(/\/+$/, '')
      const cleanUri = NormalizeUri(uri)
      if (cleanBase === '') return cleanUri
      if (cleanUri === '') return cleanBase
      return cleanBase + cleanUri
    }

    const DefaultUriForProvider = function (providerType, modelType) {
      const mt = String(modelType || 'llm').toLowerCase()
      switch (String(providerType || '').toLowerCase()) {
        case 'openai':
        case 'deepseek':
          return mt === 'embedding' ? '/v1/embeddings' : '/v1/chat/completions'
        case 'openai-responses':
          return mt === 'embedding' ? '/v1/embeddings' : '/v1/responses'
        case 'anthropic':
          return mt === 'embedding' ? '/v1/embeddings' : '/v1/messages'
        case 'google':
          return mt === 'embedding' ? '/v1/embeddings' : '/v1beta/models'
        default:
          return mt === 'embedding' ? '/v1/embeddings' : '/v1/chat/completions'
      }
    }

    const NormalizeModelRow = function (item) {
      return { ...item, model_type: item.model_type || 'llm', uri: NormalizeUri(item.uri || '') }
    }

    const getProviderType = function (providerId) {
      const p = state.providerList.find(function (item) { return Number(item.id) === Number(providerId) })
      return p ? (p.request_format || p.provider_type || 'openai') : 'openai'
    }

    const CurrentEditProviderBaseURL = function () {
      const provider = state.providerList.find(function (item) { return Number(item.id) === Number(state.editModel.provider_id) })
      return provider ? provider.base_url : ''
    }

    const getProviderModels = function (providerId) {
      return state.allModels.filter(function (m) { return Number(m.provider_id) === Number(providerId) })
    }

    const fmtCtx = function (size) {
      if (!size) return '—'
      if (size >= 1000000) return (size / 1000000).toFixed(1) + 'M'
      if (size >= 1000) return (size / 1000).toFixed(0) + 'K'
      return String(size)
    }

    // --- 展开/折叠 ---
    const toggleExpand = function (providerId) {
      if (state.expandedProviderIds.has(providerId)) {
        state.expandedProviderIds.delete(providerId)
      } else {
        state.expandedProviderIds.add(providerId)
      }
    }

    // --- 数据加载（按 mode 选择读取接口） ---
    const LoadData = function () {
      state.expandedProviderIds = new Set()
      if (props.mode === 'agent') {
        Base.BasePost('/api/AgentV2ProviderModels', {}, (res) => {
          if (res.ErrCode === 0 && res.Data && res.Data.providers) {
            const providers = res.Data.providers
            state.providerList = providers.map(function (p) {
              return { ...p, request_format: p.provider_type || p.request_format || 'openai' }
            })
            providers.forEach(function (p) { state.expandedProviderIds.add(p.id) })
            state.allModels = []
            providers.forEach(function (p) {
              ;(p.models || []).forEach(function (m) {
                state.allModels.push(NormalizeModelRow({ ...m, provider_id: p.id }))
              })
            })
          } else if (res.ErrMsg) {
            ElMessage.error(res.ErrMsg)
          }
        })
      } else {
        aiSet.AiProviderList(function (response) {
          if (response.ErrCode === 0) {
            state.providerList = (response.Data || []).map(function (item) {
              return { ...item, request_format: item.request_format || item.provider_type || 'openai' }
            })
            ;(response.Data || []).forEach(function (item) { state.expandedProviderIds.add(item.id) })
            LoadAllModels()
          } else {
            ElMessage.error(response.ErrMsg)
          }
        })
      }
    }

    const LoadAllModels = function () {
      aiSet.AiModelList({}, function (response) {
        if (response.ErrCode === 0) {
          state.allModels = (response.Data || []).map(NormalizeModelRow)
        } else {
          ElMessage.error(response.ErrMsg)
        }
      })
    }

    // --- Provider 操作 ---
    const ShowAddProvider = function () {
      state.editProvider = { request_format: 'openai', api_key: '' }
      state.showApiKey = false
      state.apiKeyFetched = false
      state.apiKeyLoading = false
      state.dialogProvider = true
    }

    const ShowEditProvider = function (row, isCopy) {
      state.editProvider = { ...row, request_format: row.request_format || row.provider_type || 'openai' }
      // 复制视为新建，不携带原密钥
      if (isCopy) state.editProvider.id = 0
      if (Number(state.editProvider.id) === 0) state.editProvider.api_key = ''
      state.showApiKey = false
      state.apiKeyFetched = false
      state.apiKeyLoading = false
      state.dialogProvider = true
    }

    const ToggleApiKey = function () {
      if (state.showApiKey) {
        state.showApiKey = false
        return
      }
      if (Number(state.editProvider.id) > 0 && !state.apiKeyFetched) {
        state.apiKeyLoading = true
        aiSet.AiProviderKeyGet({ id: state.editProvider.id }, function (res) {
          state.apiKeyLoading = false
          if (res.ErrCode === 0) {
            state.editProvider.api_key = res.Data.api_key || ''
            state.apiKeyFetched = true
            state.showApiKey = true
          } else {
            ElMessage.error(res.ErrMsg)
          }
        })
      } else {
        state.showApiKey = true
      }
    }

    const SaveProvider = function () {
      const ep = state.editProvider
      const submitData = {
        id: ep.id || undefined,
        name: ep.name,
        request_format: ep.request_format || 'openai',
        base_url: ep.base_url,
        api_key: ep.api_key,
      }
      // 编辑已有服务商且未查看/未修改 API Key 时，不提交该字段，避免用空值覆盖真实密钥
      if (Number(ep.id) > 0 && !state.apiKeyFetched) delete submitData.api_key
      aiSet.AiProviderAdd(submitData, function (response) {
        if (response.ErrCode === 0) {
          state.dialogProvider = false
          LoadData()
        } else {
          ElMessage.error(response.ErrMsg)
        }
      })
    }

    const DeleteProvider = function (row) {
      common.ConfirmProxyDelete(proxy, function () {
        aiSet.AiProviderDelete(row, function (response) {
          if (response.ErrCode === 0) {
            state.expandedProviderIds.delete(row.id)
            LoadData()
          } else {
            ElMessage.error(response.ErrMsg)
          }
        })
      })
    }

    // --- 模型操作 ---
    const ShowAddModel = function (provider) {
      const providerType = provider.request_format || provider.provider_type || 'openai'
      state.editModel = {
        provider_id: provider.id,
        model_type: 'llm',
        uri: DefaultUriForProvider(providerType, 'llm'),
      }
      if (props.mode === 'agent') state.editModel.context_size = 128000
      state.dialogModel = true
    }

    const HandleModelTypeChange = function (newModelType) {
      const pt = getProviderType(state.editModel.provider_id)
      state.editModel.uri = DefaultUriForProvider(pt, newModelType)
    }

    const ShowEditModel = function (row, isCopy) {
      state.editModel = NormalizeModelRow({ ...row })
      if (isCopy) state.editModel.id = 0
      state.dialogModel = true
    }

    const SaveModel = function () {
      const em = state.editModel
      const submitData = {
        id: em.id || undefined,
        provider_id: em.provider_id,
        name: em.name || em.model,
        model_type: em.model_type || 'llm',
        model: em.model,
        uri: NormalizeUri(em.uri),
      }
      if (props.mode === 'agent') submitData.context_size = em.context_size || 128000
      aiSet.AiModelAdd(submitData, function (response) {
        if (response.ErrCode === 0) {
          state.dialogModel = false
          LoadData()
        } else {
          ElMessage.error(response.ErrMsg)
        }
      })
    }

    const DeleteModel = function (row) {
      common.ConfirmProxyDelete(proxy, function () {
        aiSet.AiModelDelete(row, function (response) {
          if (response.ErrCode === 0) {
            LoadData()
          } else {
            ElMessage.error(response.ErrMsg)
          }
        })
      })
    }

    // 模型测试：按 mode 选择不同测试接口
    const TestModel = function (row) {
      state.testingModelId = row.id
      if (props.mode === 'agent') {
        Base.BasePost('/api/AgentV2ModelTest', {
          provider_id: row.provider_id,
          model_id: row.id,
        }, function (res) {
          state.testingModelId = 0
          const provider = state.providerList.find(function (p) { return p.id === row.provider_id })
          const pname = provider ? provider.name : 'Unknown'
          if (res.ErrCode === 0) {
            ElNotification({
              title: `测试通过 — ${pname} / ${row.name || row.model}`,
              message: (res.Data && res.Data.response) || '',
              type: 'success',
              duration: 5000,
            })
          } else {
            ElNotification({
              title: `测试失败 — ${pname} / ${row.name || row.model}`,
              message: res.ErrMsg || '未知错误',
              type: 'error',
              duration: 8000,
            })
          }
        })
      } else {
        aiSet.AiModelTest({ id: row.id }, function (response) {
          state.testingModelId = 0
          if (response.ErrCode === 0) {
            ElMessage.success((row.name || row.model || '模型') + ' 连通成功')
          } else {
            ElMessage.error(response.ErrMsg || '连通失败')
          }
        })
      }
    }

    // --- 请求日志（仅 set 模式） ---
    const OpenLogDialog = function () {
      state.dialogLog = true
    }

    const LoadRequestLogList = function () {
      const params = { limit: 100 }
      if (state.logProviderId) params.provider_id = state.logProviderId
      if (state.logModelType) params.model_type = state.logModelType
      aiSet.AiRequestLogList(params, function (response) {
        if (response.ErrCode === 0) {
          state.requestLogList = response.Data || []
        } else {
          ElMessage.error(response.ErrMsg)
        }
      })
    }

    const ShowRequestLogDetail = function (row) {
      state.logDetail = { ...row }
      state.dialogLogDetail = true
    }

    const FormatJson = function (str) {
      if (!str) return ''
      if (typeof str === 'object') return JSON.stringify(str, null, 2)
      try { return JSON.stringify(JSON.parse(str), null, 2) } catch (e) { return String(str) }
    }

    LoadData()

    // 供父组件（如 Set.vue 切换 tab 时）主动刷新数据
    expose({ reload: LoadData })

    return {
      state,
      mode: props.mode,
      showCopy: props.showCopy,
      showRequestLog: props.showRequestLog,
      formatProviderType,
      BuildRequestUrl,
      DefaultUriForProvider,
      fmtCtx,
      getProviderType,
      CurrentEditProviderBaseURL,
      getProviderModels,
      toggleExpand,
      OpenLogDialog,
      ShowAddProvider,
      ShowEditProvider,
      ToggleApiKey,
      SaveProvider,
      DeleteProvider,
      ShowAddModel,
      HandleModelTypeChange,
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
