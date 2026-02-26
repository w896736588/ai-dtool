import base from "@/utils/base";

function ComposeList(callBack){
    base.BasePost('/api/Set/DockerComposeList', {} , callBack)
}
function ComposeAdd(data , callBack){
    base.BasePost(
        '/api/Set/DockerComposeAdd',
        data,
        function (response) {
            callBack(response)
        }
    )
}
function ComposeDelete(data , callBack){
    base.BasePost(
        '/api/Set/DockerComposeDelete',
        data,
        function (response) {
            callBack(response)
        }
    )
}
export default {
    ComposeList,
    ComposeAdd,
    ComposeDelete,
}