import base from "@/utils/base";

function MysqlList(params, callBack){
    if(typeof params === 'function'){
        callBack = params
        params = {}
    }
    base.BasePost('/api/Set/MysqlList', params, callBack)
}
function MysqlAdd(data , callBack){
    base.BasePost(
        '/api/Set/MysqlAdd',
        data,
        function (response) {
            callBack(response)
        }
    )
}
function MysqlDelete(data , callBack){
    base.BasePost(
        '/api/Set/MysqlDelete',
        data,
        function (response) {
            callBack(response)
        }
    )
}
export default {
    MysqlList,
    MysqlAdd,
    MysqlDelete,
}