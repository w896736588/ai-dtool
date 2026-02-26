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
    CreateDir,
    CreateApi,
    Apis,
    ApiRun,
    CreateCollectionEnv,
    CollectionEnvs,
    CreateCollectionEnvItem,
    CollectionEnvItems,
    DeleteCollection,
    DeleteApi,
    DeleteDir,
    ApiCode,
    ApiWeightDown,
    ApiTakeJsonResult,
    ApiImportJson,
    FolderDetail,
}