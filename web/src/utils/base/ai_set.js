import base from "@/utils/base";

// AiProviderList 查询 AI 服务商列表
function AiProviderList(callBack){
    base.BasePost('/api/Set/AiProviderList', {} , callBack)
}

// AiProviderAdd 新增或更新 AI 服务商
function AiProviderAdd(data , callBack){
    base.BasePost('/api/Set/AiProviderAdd', data, callBack)
}

// AiProviderDelete 删除 AI 服务商
function AiProviderDelete(data , callBack){
    base.BasePost('/api/Set/AiProviderDelete', data, callBack)
}

// AiModelList 查询 AI 模型列表
function AiModelList(data , callBack){
    base.BasePost('/api/Set/AiModelList', data || {}, callBack)
}

// AiModelAdd 新增或更新 AI 模型
function AiModelAdd(data , callBack){
    base.BasePost('/api/Set/AiModelAdd', data, callBack)
}

// AiModelDelete 删除 AI 模型
function AiModelDelete(data , callBack){
    base.BasePost('/api/Set/AiModelDelete', data, callBack)
}

// AiModelTest 测试 AI 模型连通性
function AiModelTest(data , callBack){
    base.BasePost('/api/Set/AiModelTest', data, callBack)
}

// AiRequestLogList 查询 AI 请求日志列表
function AiRequestLogList(data , callBack){
    base.BasePost('/api/Set/AiRequestLogList', data || {}, callBack)
}

export default {
    AiProviderList,
    AiProviderAdd,
    AiProviderDelete,
    AiModelList,
    AiModelAdd,
    AiModelDelete,
    AiModelTest,
    AiRequestLogList,
}

