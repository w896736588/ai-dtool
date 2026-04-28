//单sse连接，用于所有sse
const SseClientId = 'sse_client_id'
import t from "@/utils/base/type";
import base from '@/utils/base'
import store from "@/utils/base/store"

let SseConn = null
let SseReceiveIdFunc = {}

let sseClientId = ''
let sseUrl = ''
let initFromLoginStatusPromise = null
//全局获取sse 客户端id
function GetSseClientId(){
    return sseClientId
}
function Create(ssePort) {
    if (ssePort) {
        base.SetSsePort(ssePort)
    }
    const nextClientId = sseClientId || base.GenerateId(SseClientId)
    let params = 'client_id=' + nextClientId + '&token=' + encodeURIComponent(base.GetSafeToken())
    const sseHost = base.GetSseApiHost()
    if (!sseHost) {
        return false
    }
    let url = sseHost + '/sse?' + params
    if (SseConn && sseUrl === url) {
        return true
    }
    if (SseConn) {
        SseConn.close()
    }
    sseClientId = nextClientId
    sseUrl = url
    SseConn = new EventSource(url)
    return true
}

function OpenFunc(callFunc) {
    if (!SseConn) {
        return
    }
    SseConn.onopen = callFunc;
}

function ErrorFunc(callFunc) {
    if (!SseConn) {
        return
    }
    SseConn.onerror = callFunc
}

function CloseFunc(callFunc) {
    if (!SseConn) {
        return
    }
    SseConn.onclosse = callFunc
}

function ReceiveMessage() {
    if (!SseConn) {
        return
    }
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
    if (!SseConn) {
        return
    }
    SseConn.close()
    SseConn = null
    sseUrl = ''
    sseClientId = ''
    SseReceiveIdFunc = {}
}

function InitFromLoginStatus(openFunc, errorFunc, closeFunc) {
    if (initFromLoginStatusPromise) {
        return initFromLoginStatusPromise
    }
    initFromLoginStatusPromise = new Promise(resolve => {
        base.BaseLoginStatus(function (response) {
            if (response.ErrCode !== 0) {
                resolve(false)
                return
            }
            const data = response.Data || {}
            const created = Create(data.sse_port)
            if (created) {
                if (openFunc) {
                    OpenFunc(openFunc)
                }
                if (errorFunc) {
                    ErrorFunc(errorFunc)
                }
                if (closeFunc) {
                    CloseFunc(closeFunc)
                }
                ReceiveMessage()
            }
            resolve(created)
        })
    })
    return initFromLoginStatusPromise
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
    InitFromLoginStatus,
    CloseFunc,
    ReceiveMessage,
    GetSseDistributeId,
    GetSseClientId,
}
