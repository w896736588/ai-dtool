//单sse连接，用于所有sse
const SseClientId = 'sse_client_id'
import t from "@/utils/base/type";
import base from '@/utils/base'
import store from "@/utils/base/store"

let SseConn = null
let SseReceiveIdFunc = {}

let sseClientId = ''

//Shell连接状态SSE
let ShellConnectionsSseConn = null
const SseShellConnections = 'shell_connections'

//全局获取sse 客户端id
function GetSseClientId(){
    return sseClientId
}
function Create() {
    sseClientId = base.GenerateId(SseClientId)
    let params = 'client_id=' + sseClientId
    let url = base.GetApiHost() + '/sse?' + params
    SseConn = new EventSource(url)
    //创建Shell连接状态SSE连接
    CreateShellConnectionsSse()
}

//创建Shell连接状态SSE连接
function CreateShellConnectionsSse(){
    let url = base.GetApiHost() + '/sse?client_id=' + SseShellConnections
    ShellConnectionsSseConn = new EventSource(url)
    ShellConnectionsSseConn.onopen = function(event){
        console.log('ShellConnections SSE连接已建立')
    }
    ShellConnectionsSseConn.onerror = function(event){
        console.log('ShellConnections SSE连接错误', event)
        //关闭后重新建立连接
        setTimeout(() => {
            CreateShellConnectionsSse()
        }, 3000)
    }
    ShellConnectionsSseConn.onmessage = function(event){
        let objData = null
        try {
            objData = JSON.parse(event.data)
        } catch (e) {
            console.log('解析ShellConnections SSE内容失败', event.data, e)
            return
        }
        if (objData && objData.sse_distribute_id === SseShellConnections) {
            if (SseReceiveIdFunc[SseShellConnections]) {
                try {
                    SseReceiveIdFunc[SseShellConnections](objData.data, objData.type, objData.sse_distribute_id)
                } catch (e) {
                    console.log('回调处理ShellConnections SSE内容失败', objData.data, e)
                }
            }
        }
    }
}

function OpenFunc(callFunc) {
    SseConn.onopen = callFunc;
}

function ErrorFunc(callFunc) {
    SseConn.onerror = callFunc
}

function CloseFunc(callFunc) {
    SseConn.onclosse = callFunc
}

function ReceiveMessage() {
    SseConn.onmessage = function (event) {
        let objData = null
        try {
            objData = JSON.parse(event.data)
        } catch (e) {
            console.log('解析sse内容失败 %s', '----' + event.data + '----', e)
            return
        }
        if (objData && objData.sse_distribute_id) {
            if (SseReceiveIdFunc[objData.sse_distribute_id]) {
                try {
                    SseReceiveIdFunc[objData.sse_distribute_id](objData.data, objData.type,objData.sse_distribute_id)
                } catch (e) {
                    console.log('回调处理sse内容失败 %s', '----' + event.data + '----', e)
                }
            } else {
                console.log('未找到对应的回调函数 %s', objData.sse_distribute_id)
            }
        } else {
            console.log('未找到对应的回调函数 %s', event.data)
        }

    };
}

function RegisterReceive(receiveId, callFunc) {
    SseReceiveIdFunc[receiveId] = callFunc
}

function UnRegisterReceive(receiveId){
    delete SseReceiveIdFunc[receiveId]
}

function Close() {
    SseConn.close()
    if (ShellConnectionsSseConn) {
        ShellConnectionsSseConn.close()
        ShellConnectionsSseConn = null
    }
    SseReceiveIdFunc = {}
}

//获取分发id
function GetSseDistributeId(businessId){
    return businessId
}

export default {
    OpenFunc,
    ErrorFunc,
    RegisterReceive,
    UnRegisterReceive,
    Close,
    Create,
    CloseFunc,
    ReceiveMessage,
    GetSseDistributeId,
    GetSseClientId,
}