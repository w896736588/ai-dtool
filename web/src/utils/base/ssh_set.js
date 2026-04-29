import base from "@/utils/base";

function SshList(callBack, params = {}){
    base.BasePost('/api/Set/SshList', params , callBack)
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
function ReconnectConnection(shellClientId, callBack){
    base.BasePost('/api/shellOutReconnect', {shell_client_id: shellClientId} , callBack)
}
export default {
    SshList,
    SshAdd,
    SshDelete,
    ReconnectConnection,
}