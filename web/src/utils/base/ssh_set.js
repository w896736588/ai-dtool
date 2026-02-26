import base from "@/utils/base";

function SshList(callBack){
    base.BasePost('/api/Set/SshList', {} , callBack)
}
function SshAdd(data , callBack){
    base.BasePost(
        '/api/Set/SshAdd',
        data,
        function (response) {
            callBack(response)
        }
    )
}
function SshDelete(data , callBack){
    base.BasePost(
        '/api/Set/SshDelete',
        data,
        function (response) {
            callBack(response)
        }
    )
}
function GetConnections(callBack){
    base.BasePost('/api/shellOutConnections', {} , callBack)
}
function ReconnectConnection(shellClientId, callBack){
    base.BasePost('/api/shellOutReconnect', {shell_client_id: shellClientId} , callBack)
}
export default {
    SshList,
    SshAdd,
    SshDelete,
    GetConnections,
    ReconnectConnection,
}