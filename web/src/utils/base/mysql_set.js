import base from "@/utils/base";

function MysqlList(callBack){
    base.BasePost('/api/Set/MysqlList', {} , callBack)
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