import base from "@/utils/base";

function SupervisorList(callBack){
    base.BasePost('/api/Set/SupervisorList', {} , callBack)
}
function SupervisorAdd(data , callBack){
    base.BasePost('/api/Set/SupervisorAdd', data, callBack)
}
function SupervisorDelete(data , callBack){
    base.BasePost('/api/Set/SupervisorDelete', data, callBack)
}

function SupervisorGroupList(callBack){
    base.BasePost('/api/Set/SupervisorGroupList', {} , callBack)
}
function SupervisorGroupAdd(data , callBack){
    base.BasePost('/api/Set/SupervisorGroupAdd', data, callBack)
}
function SupervisorGroupDelete(data , callBack){
    base.BasePost('/api/Set/SupervisorGroupDelete', data, callBack)
}

function SupervisorQuickList(data , callBack){
    base.BasePost('/api/Set/SupervisorQuickList', data, callBack)
}
export default {
    SupervisorList,
    SupervisorAdd,
    SupervisorDelete,
    SupervisorGroupList,
    SupervisorGroupAdd,
    SupervisorGroupDelete,
    SupervisorQuickList,
}