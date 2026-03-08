<template>
  <div class="set-config-page">
    <div class="set-config-header">
      <h3 class="set-config-title">AI 服务商与模型配置</h3>
      <p class="set-config-desc">模型属于服务商子项，当前请求格式仅支持 openai</p>
    </div>

    <el-tabs v-model="state.activeTab" class="set-config-inner-tabs" @tab-change="HandleInnerTabChange">
      <el-tab-pane label="服务商配置" name="provider">
        <div class="set-config-actions" style="margin-bottom: 10px;">
          <el-button type="primary" @click="ShowAddProvider">新增服务商</el-button>
        </div>
        <div class="set-config-table-card">
          <el-table :data="state.providerList" class="set-config-table" row-key="id">
            <el-table-column prop="id" label="#id" width="70"/>
            <el-table-column prop="name" label="服务商名称" min-width="160"/>
            <el-table-column prop="request_format" label="请求格式" width="140"/>
            <el-table-column prop="base_url" label="Base URL" min-width="220"/>
            <el-table-column prop="api_key" label="API Key" min-width="160">
              <template #default="scope">
                <span>{{ MaskKey(scope.row.api_key) }}</span>
              </template>
            </el-table-column>
            <el-table-column label="操作" width="220">
              <template #default="scope">
                <div class="set-op-group">
                  <el-button type="primary" link @click="ShowEditProvider(scope.row, false)">编辑</el-button>
                  <el-button type="primary" link @click="ShowEditProvider(scope.row, true)">复制新增</el-button>
                  <el-button type="primary" link @click="SwitchToModelTab(scope.row)">管理模型</el-button>
                  <el-button link type="danger" @click="DeleteProvider(scope.row)">删除</el-button>
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
              <el-option
                  :label="provider.name + ' (' + provider.request_format + ')'
"
                  :value="provider.id"
              />
            </template>
          </el-select>
          <el-button type="primary" :disabled="!state.currentProviderId" @click="ShowAddModel">新增模型</el-button>
        </div>

        <div class="set-config-table-card">
          <el-table :data="state.modelList" class="set-config-table" row-key="id">
            <el-table-column prop="id" label="#id" width="70"/>
            <el-table-column prop="provider_name" label="所属服务商" min-width="150"/>
            <el-table-column prop="name" label="展示名" min-width="170"/>
            <el-table-column prop="model" label="模型标识" min-width="220"/>
            <el-table-column label="操作" width="200">
              <template #default="scope">
                <div class="set-op-group">
                  <el-button type="primary" link @click="ShowEditModel(scope.row, false)">编辑</el-button>
                  <el-button type="primary" link @click="ShowEditModel(scope.row, true)">复制新增</el-button>
                  <el-button link type="danger" @click="DeleteModel(scope.row)">删除</el-button>
                </div>
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
        <el-form-item label="Base URL">
          <el-input v-model="state.editProvider.base_url" autocomplete="off" placeholder="例如: https://api.openai.com/v1/chat/completions"/>
        </el-form-item>
        <el-form-item label="API Key">
          <el-input v-model="state.editProvider.api_key" type="password" show-password autocomplete="off"/>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="state.dialogProvider = false">取消</el-button>
          <el-button type="primary" @click="SaveProvider">保存</el-button>
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
        <el-form-item label="模型标识">
          <el-input v-model="state.editModel.model" autocomplete="off" placeholder="例如: gpt-4o-mini"/>
        </el-form-item>
      </el-form>
      <template #footer>
        <div class="dialog-footer">
          <el-button @click="state.dialogModel = false">取消</el-button>
          <el-button type="primary" @click="SaveModel">保存</el-button>
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

    // state 页面状态容器
    const state = reactive({
      activeTab: 'provider',
      providerList: [],
      currentProviderId: 0,
      modelList: [],
      dialogProvider: false,
      dialogModel: false,
      editProvider: {},
      editModel: {},
    })

    // MaskKey 隐藏敏感 Key 文本
    const MaskKey = function (key){
      const str = String(key || '')
      if(str.length <= 6){
        return str === '' ? '' : '******'
      }
      return str.slice(0, 3) + '******' + str.slice(-3)
    }

    // EnsureCurrentProvider 确保模型页有可用服务商
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

    // LoadProviderList 拉取服务商配置
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

    // LoadModelList 拉取当前服务商模型列表
    const LoadModelList = function (){
      if(!state.currentProviderId){
        state.modelList = []
        return
      }
      aiSet.AiModelList({provider_id: state.currentProviderId}, function (response){
        if(response.ErrCode === 0){
          state.modelList = response.Data || []
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
      })
    }

    // SwitchToModelTab 切换到模型配置页
    const SwitchToModelTab = function (provider){
      state.currentProviderId = provider.id
      state.activeTab = 'model'
      LoadModelList()
    }

    // HandleInnerTabChange 处理内层标签切换，进入模型页时自动加载列表
    const HandleInnerTabChange = function (tabName){
      if(String(tabName) === 'model'){
        LoadModelList()
      }
    }

    // ShowAddProvider 打开新增服务商弹窗
    const ShowAddProvider = function (){
      state.editProvider = {
        request_format: 'openai',
      }
      state.dialogProvider = true
    }

    // ShowEditProvider 打开编辑服务商弹窗
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

    // SaveProvider 保存服务商配置
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

    // DeleteProvider 删除服务商配置
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

    // ShowAddModel 打开新增模型弹窗
    const ShowAddModel = function (){
      if(!state.currentProviderId){
        instance.$helperNotify.error('请先选择服务商')
        return
      }
      state.editModel = {
        provider_id: state.currentProviderId,
      }
      state.dialogModel = true
    }

    // ShowEditModel 打开编辑模型弹窗
    const ShowEditModel = function (row, isCopy){
      state.editModel = {
        ...row,
      }
      if(isCopy){
        state.editModel.id = 0
      }
      state.dialogModel = true
    }

    // SaveModel 保存模型配置
    const SaveModel = function (){
      aiSet.AiModelAdd(state.editModel, function (response){
        if(response.ErrCode === 0){
          state.dialogModel = false
          LoadModelList()
        }else{
          instance.$helperNotify.error(response.ErrMsg)
        }
      })
    }

    // DeleteModel 删除模型配置
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

    // 初始化页面数据
    LoadProviderList()

    return {
      state,
      MaskKey,
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

.set-config-inner-tabs :deep(.el-tabs__header) {
  margin-bottom: 10px;
}

/* 统一 AI 配置页按钮风格，避免显示 Element 默认蓝色 */
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

.set-config-page :deep(.el-button--primary.is-link) {
  color: #4f804f !important;
}

.set-config-page :deep(.el-button--primary.is-link:hover) {
  color: #3f6f3f !important;
}
</style>
