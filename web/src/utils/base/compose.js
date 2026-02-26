import base from "@/utils/base";

function DockerComposeList(data , callBack){
    console.log(data , callBack)
    base.BasePost('/api/DockerComposeList', data , callBack)
}
function DockerComposeRestart(data , callBack){
    base.BasePost('/api/DockerComposeRestart', data, callBack)
}
function DockerComposeStop(data , callBack){
    base.BasePost('/api/DockerComposeStop', data, callBack)
}
function DockerComposeStart(data , callBack){
    base.BasePost('/api/DockerComposeStart', data, callBack)
}
function DockerComposeStatus(data , callBack){
    base.BasePost('/api/DockerComposeStatus', data, callBack)
}
function DockerComposeServices(data , callBack){
    base.BasePost('/api/DockerComposeServices', data, callBack)
}

function DockerComposeConfigShow(data , callBack){
    base.BasePost('/api/DockerComposeConfigShow', data, callBack)
}
export default {
    DockerComposeList,
    DockerComposeRestart,
    DockerComposeStop,
    DockerComposeConfigShow,
    DockerComposeStart,
    DockerComposeStatus,
    DockerComposeServices,
}