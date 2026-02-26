import base from '../base'
import module from '../module'
import store from './store'

function SetSmartLinkGroupList(callBack){
    base.BasePost('/api/Set/SmartLinkGroupList', {}, callBack)
}

function SetSmartLinkGroupAdd(editSmartLinkGroupConfig , callBack) {
    base.BasePost('/api/Set/SmartLinkGroupAdd', editSmartLinkGroupConfig, callBack)
}

function SetSmartLinkGroupDelete(editSmartLinkGroupConfig , callBack){
    base.BasePost('/api/Set/SmartLinkGroupDelete', editSmartLinkGroupConfig, callBack)
}

function SmartLinkList(callBack){
    base.BasePost('/api/SmartLinkList', {}, callBack)
}

function SmartLinkAdd(smart_link_config , callBack){
    base.BasePost('/api/SmartLinkAdd', smart_link_config , callBack)
}

function SmartLinkDelete(smart_link_config , callBack){
    base.BasePost('/api/SmartLinkDel', smart_link_config , callBack)
}

function SmartLinkRecycle(sseDistributeId , callBack){
    base.BasePost('/api/SmartLinkRecycle', {sse_distribute_id : sseDistributeId} , callBack)
}

function SmartLinkDownloadPath(sseDistributeId , callBack){
    base.BasePost('/api/SmartLinkDownloadPath', {sse_distribute_id : sseDistributeId} , callBack)
}

function SmartLinkRun(runParams , callBack){
    base.BasePost('/api/SmartLinkRun', runParams, callBack)
}

function SmartLinkRunList(sseDistributeId , callBack){
    base.BasePost('/api/SmartLinkRunList', {sse_distribute_id : sseDistributeId}, callBack)
}

function SmartLinkChromeVersion(sseDistributeId , callBack){
    base.BasePost('/api/SmartLinkChromeVersion', {sse_distribute_id : sseDistributeId}, callBack)
}

function SmartLinkChromeUpdate(sseDistributeId , callBack){
    base.BasePost('/api/SmartLinkChromeDownload', {sse_distribute_id : sseDistributeId}, callBack)
}

export default {
    SetSmartLinkGroupList,
    SetSmartLinkGroupAdd,
    SetSmartLinkGroupDelete,
    SmartLinkList,
    SmartLinkAdd,
    SmartLinkDelete,
    SmartLinkRun,
    SmartLinkRunList,
    SmartLinkChromeVersion,
    SmartLinkChromeUpdate,
    SmartLinkRecycle,
    SmartLinkDownloadPath,
}
