import base from "@/utils/base";

function AccountList(callBack){
    base.BasePost('/api/Set/AccountList', {} , callBack)
}
function AccountAdd(data , callBack){
    base.BasePost('/api/Set/AccountAdd', data, callBack)
}
function AccountDelete(data , callBack){
    base.BasePost('/api/Set/AccountDelete', data, callBack)
}

function AccountGroupList(callBack){
    base.BasePost('/api/Set/AccountGroupList', {} , callBack)
}
function AccountGroupAdd(data , callBack){
    base.BasePost('/api/Set/AccountGroupAdd', data, callBack)
}
function AccountGroupDelete(data , callBack){
    base.BasePost('/api/Set/AccountGroupDelete', data, callBack)
}

export default {
    AccountList,
    AccountAdd,
    AccountDelete,
    AccountGroupList,
    AccountGroupAdd,
    AccountGroupDelete,
}