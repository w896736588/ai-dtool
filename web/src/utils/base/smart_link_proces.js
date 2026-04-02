import base from "@/utils/base";

function SmartProcessList(callBack){
    base.BasePost('/api/SmartProcessList', {} , callBack)
}

// data为json 包括id（新增时为0,编辑时不为0）name
function SmartProcessAdd(data , callBack){
    base.BasePost('/api/SmartProcessAdd', data, callBack)
}
// data 为json，只有一个id
function SmartProcessDelete(data , callBack){
    base.BasePost('/api/SmartProcessDelete', data, callBack)
}

//data为json，只有一个参数，smart_link_process_id 是执行逻辑的id
function SmartProcessItemList(data,callBack){
    base.BasePost('/api/SmartProcessItemList', data , callBack)
}

// data为json，新增或者编辑执行逻辑子项，字段见tbl_smart_link_process_item表
function SmartProcessItemAdd(data , callBack){
    base.BasePost('/api/SmartProcessItemAdd', data, callBack)
}

function SmartProcessItemDelete(data , callBack){
    base.BasePost('/api/SmartProcessItemDelete', data, callBack)
}
//排序所有的执行逻辑子项，data为json 包括 smart_link_process_id 是执行逻辑的id smart_link_process_item_ids 执行逻辑子项合集，用英文逗号分割，按照拖动后的排序位置取id
function SmartProcessItemSort(data , callBack){
    base.BasePost('/api/SmartProcessItemSort', data, callBack)
}

//获取执行逻辑的明细，data为json 包括 smart_link_process_id 是执行逻辑的id 返回这个执行逻辑下面所有的执行逻辑子项
function SmartProcessItemDetail(data , callBack){
    base.BasePost('/api/SmartProcessItemDetail', data, callBack)
}

// data为json 关联节点
function SmartProcessSetRelation(data , callBack){
    base.BasePost('/api/SmartProcessSetRelation', data, callBack)
}

// data为json 取消关联节点
function SmartProcessCancelRelation(data , callBack){
    base.BasePost('/api/SmartProcessCancelRelation', data, callBack)
}

// data为json 设置节点位置
function SmartProcessSetPosition(data , callBack){
    base.BasePost('/api/SmartProcessSetPosition', data, callBack)
}

// data 为自动提取基础定位配置的请求参数。
// data stores params for AI auto extraction of base locator config.
function SmartLinkLocatorAutoExtract(data , callBack){
    base.BasePost('/api/SmartLinkLocatorAutoExtract', data, callBack)
}


export default {
    SmartProcessList,
    SmartProcessAdd,
    SmartProcessDelete,
    SmartProcessItemList,
    SmartProcessItemAdd,
    SmartProcessItemDelete,
    SmartProcessItemSort,
    SmartProcessItemDetail,
    SmartProcessSetRelation,
    SmartProcessCancelRelation,
    SmartProcessSetPosition,
    SmartLinkLocatorAutoExtract,
}
