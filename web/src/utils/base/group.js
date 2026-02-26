import base from "@/utils/base";


function GroupList(data , callBack){
    base.BasePost('/api/Set/GroupList', data , callBack)
}
function GroupAdd(data , callBack){
    base.BasePost('/api/Set/GroupAdd', data, callBack)
}
function GroupDelete(data , callBack){
    base.BasePost('/api/Set/GroupDelete', data, callBack)
}

export default {
    GroupList,
    GroupAdd,
    GroupDelete,
}