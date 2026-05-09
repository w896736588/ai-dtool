//单sse连接，用于所有sse
const SseClientId = 'sse_client_id'
import t from "@/utils/base/type";
import base from '@/utils/base'
import store from "@/utils/base/store"
import { ElMessageBox } from 'element-plus'

let SseConn = null
let SseReceiveIdFunc = {}

let sseClientId = ''
let sseUrl = ''
let initFromLoginStatusPromise = null
let sseLimitDialogShown = false
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
    const sseHost = base.GetSseApiHost(ssePort || undefined)
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

// showSseLimitDialog 无可用端口时弹窗确认是否关闭页面，防重复弹出
function showSseLimitDialog() {
    if (sseLimitDialogShown) return
    sseLimitDialogShown = true
    ElMessageBox.confirm('SSE连接数已超限，无可用端口。是否关闭此页面？', '连接超限', {
        confirmButtonText: '关闭页面',
        cancelButtonText: '继续停留',
        type: 'warning'
    }).then(() => {
        window.close()
    }).catch(() => {})
}

// fetchAvailableSsePort 调用后端接口获取一个可用的 SSE 端口
// 返回 Promise<string|null>，null 表示无可用端口
function fetchAvailableSsePort() {
    return new Promise(resolve => {
        base.BasePost('/api/SseAvailablePort', {}, function (response) {
            if (response.ErrCode !== 0 || !response.Data || !response.Data.sse_ports) {
                resolve(null)
                return
            }
            const ports = response.Data.sse_ports || []
            for (let i = 0; i < ports.length; i++) {
                if (ports[i].available) {
                    resolve(ports[i].port)
                    return
                }
            }
            resolve(null)
        })
    })
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
            // 优先使用后端返回的 sse_ports 数组
            const ssePorts = data.sse_ports || []
            if (ssePorts.length === 0) {
                resolve(false)
                return
            }
            connectWithAvailablePort(openFunc, errorFunc, closeFunc, resolve)
        })
    })
    return initFromLoginStatusPromise
}

// connectWithAvailablePort 查询可用端口并建连
function connectWithAvailablePort(openFunc, errorFunc, closeFunc, resolve) {
    fetchAvailableSsePort().then(availablePort => {
        if (!availablePort) {
            showSseLimitDialog()
            if (errorFunc) {
                errorFunc(new Event('error'))
            }
            resolve(false)
            return
        }
        const created = Create(availablePort)
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
}

// ConnectToAvailablePort 直接查询可用端口并建连（不重复调 BaseLoginStatus）
function ConnectToAvailablePort(openFunc, errorFunc, closeFunc) {
    return new Promise(resolve => {
        connectWithAvailablePort(openFunc, errorFunc, closeFunc, resolve)
    })
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
    ConnectToAvailablePort,
    CloseFunc,
    ReceiveMessage,
    GetSseDistributeId,
    GetSseClientId,
}
