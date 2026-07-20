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

function SmartLinkRecycle(sseDistributeId , callBack){
    base.BasePost('/api/SmartLinkRecycle', {sse_distribute_id : sseDistributeId} , callBack)
}

function SmartLinkDownloadPath(sseDistributeId , callBack){
    base.BasePost('/api/SmartLinkDownloadPath', {sse_distribute_id : sseDistributeId} , callBack)
}

function SmartLinkOpenDataDir(callBack){
    base.BasePost('/api/SmartLinkOpenDataDir', {} , callBack)
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

// SmartLinkItemList 新表 smart_link 列表（带分组信息）
function SmartLinkItemList(dataOrCallBack, maybeCallBack) {
    const data = typeof dataOrCallBack === 'function' ? {} : (dataOrCallBack || {})
    const callBack = typeof dataOrCallBack === 'function' ? dataOrCallBack : maybeCallBack
    base.BasePost('/api/SmartLinkItemList', data, callBack)
}

// SmartLinkItemAdd 新表 smart_link 新增/编辑
function SmartLinkItemAdd(data, callBack) {
    base.BasePost('/api/SmartLinkItemAdd', data, callBack)
}

// SmartLinkItemDelete 新表 smart_link 删除
function SmartLinkItemDelete(data, callBack) {
    base.BasePost('/api/SmartLinkItemDelete', data, callBack)
}

// SmartLinkItemInfo 新表 smart_link 详情
function SmartLinkItemInfo(data, callBack) {
    base.BasePost('/api/SmartLinkItemInfo', data, callBack)
}

// SmartLinkMigrateOldData 触发老数据迁移
function SmartLinkMigrateOldData(callBack) {
    base.BasePost('/api/SmartLinkMigrateOldData', {}, callBack)
}

export default {
    SetSmartLinkGroupList,
    SetSmartLinkGroupAdd,
    SetSmartLinkGroupDelete,
    SmartLinkRun,
    SmartLinkRunList,
    SmartLinkChromeVersion,
    SmartLinkChromeUpdate,
    SmartLinkRecycle,
    SmartLinkDownloadPath,
    SmartLinkOpenDataDir,
    // 新表接口
    SmartLinkItemList,
    SmartLinkItemAdd,
    SmartLinkItemDelete,
    SmartLinkItemInfo,
    SmartLinkMigrateOldData,
}
