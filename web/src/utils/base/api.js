import base from "@/utils/base";

function CreateCollection(data , callBack){
    base.BasePost('/api/CreateCollection', data , callBack)
}

function DeleteCollection(data , callBack){
    base.BasePost('/api/DeleteCollection', data , callBack)
}

function DeleteDir(data , callBack){
    base.BasePost('/api/DeleteDir', data , callBack)
}

function Collections(data , callBack){
    base.BasePost('/api/Collections', data , callBack)
}

// 中文注释：查询所有集合基础信息。
function CollectionListBasic(data , callBack){
    base.BasePost('/api/CollectionListBasic', data , callBack)
}

// 中文注释：按集合查询文件夹基础信息。
function CollectionFoldersBasic(data , callBack){
    base.BasePost('/api/CollectionFoldersBasic', data , callBack)
}

function CreateCollectionEnv(data , callBack){
    base.BasePost('/api/CreateCollectionEnv', data , callBack)
}

function CollectionEnvs(data , callBack){
    base.BasePost('/api/CollectionEnvs', data , callBack)
}

function CreateDir(data , callBack){
    base.BasePost('/api/CreateDir', data , callBack)
}

function CreateApi(data , callBack){
    base.BasePost('/api/CreateApi', data , callBack)
}

function Apis(data , callBack){
    base.BasePost('/api/Apis', data , callBack)
}

// 中文注释：按文件夹查询接口基础信息。
function FolderApisBasic(data , callBack){
    base.BasePost('/api/FolderApisBasic', data , callBack)
}

// 中文注释：按若干接口ID查询接口明细。
function ApisDetailByIds(data , callBack){
    base.BasePost('/api/ApisDetailByIds', data , callBack)
}

function ApiRun(data , callBack){
    base.BasePost('/api/ApiRun', data , callBack)
}

function CreateCollectionEnvItem(data , callBack){
    base.BasePost('/api/CreateCollectionEnvItem', data , callBack)
}

function CollectionEnvItems(data , callBack){
    base.BasePost('/api/CollectionEnvItems', data , callBack)
}

function DeleteApi(data , callBack){
    base.BasePost('/api/DeleteApi', data , callBack)
}

function ApiWeightDown(data , callBack){
    base.BasePost('/api/ApiWeightDown', data , callBack)
}

function ApiMove(data , callBack){
    base.BasePost('/api/ApiMove', data , callBack)
}

function ApiCode(data , callBack){
    base.BasePost('/api/ApiCode', data , callBack)
}

function ApiTakeJsonResult(data , callBack){
    base.BasePost('/api/ApiTakeJsonResult', data , callBack)
}

function ApiImportJson(data , callBack){
    const formData = new FormData()
    formData.append('collection_id', data.collection_id)
    formData.append('json', data.json)
    base.BasePostForm('/api/ApiBatchImport', formData , callBack)
}

function FolderDetail(data , callBack){
    base.BasePost('/api/FolderDetail', data , callBack)
}

export default {
    CreateCollection,
    Collections,
    CollectionListBasic,
    CollectionFoldersBasic,
    CreateDir,
    CreateApi,
    Apis,
    FolderApisBasic,
    ApisDetailByIds,
    ApiRun,
    CreateCollectionEnv,
    CollectionEnvs,
    CreateCollectionEnvItem,
    CollectionEnvItems,
    DeleteCollection,
    DeleteApi,
    DeleteDir,
    ApiCode,
    ApiMove,
    ApiWeightDown,
    ApiTakeJsonResult,
    ApiImportJson,
    FolderDetail,
}
