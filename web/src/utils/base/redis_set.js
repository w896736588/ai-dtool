import base from "@/utils/base";

function RedisList(callBack){
    base.BasePost('/api/Set/RedisList', {} , callBack)
}
function RedisAdd(data , callBack){
    base.BasePost(
        '/api/Set/RedisAdd',
        data,
        function (response) {
            callBack(response)
        }
    )
}
function RedisDelete(data , callBack){
    base.BasePost(
        '/api/Set/RedisDelete',
        data,
        function (response) {
            callBack(response)
        }
    )
}
export default {
    RedisList,
    RedisAdd,
    RedisDelete,
}