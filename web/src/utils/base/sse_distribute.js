//单sse连接，用于所有sse
const SseClientId = 'distribute'
import t from "@/utils/base/type";
import base from '@/utils/base'
import store from "@/utils/base/store"

let SseConn = null
let SseReceiveIdFunc = {}

let sseClientId = ''

//全局获取sse 客户端id
function GetSseClientId(){
    return sseClientId
}
function Create() {
    sseClientId = base.GenerateId(SseClientId)
    let params = 'client_id=' + sseClientId
    let url = base.GetApiHost() + '/sse?' + params
    SseConn = new EventSource(url)
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