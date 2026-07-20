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
function DockerComposeDown(data , callBack){
    base.BasePost('/api/DockerComposeDown', data, callBack)
}
function DockerComposePull(data , callBack){
    base.BasePost('/api/DockerComposePull', data, callBack)
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
function DockerImageList(data , callBack){
    base.BasePost('/api/DockerImageList', data, callBack)
}
function DockerImageContainers(data , callBack){
    base.BasePost('/api/DockerImageContainers', data, callBack)
}
function DockerImageRemove(data , callBack){
    base.BasePost('/api/DockerImageRemove', data, callBack)
}
function DockerContainerStop(data , callBack){
    base.BasePost('/api/DockerContainerStop', data, callBack)
}
function DockerContainerRemove(data , callBack){
    base.BasePost('/api/DockerContainerRemove', data, callBack)
}
function DockerContainerLogTruncate(data , callBack){
    base.BasePost('/api/DockerContainerLogTruncate', data, callBack)
}
export default {
    DockerComposeList,
    DockerComposeRestart,
    DockerComposeStop,
    DockerComposeDown,
    DockerComposePull,
    DockerComposeConfigShow,
    DockerComposeStart,
    DockerComposeStatus,
    DockerComposeServices,
    DockerImageList,
    DockerImageContainers,
    DockerImageRemove,
    DockerContainerStop,
    DockerContainerRemove,
    DockerContainerLogTruncate,
}

