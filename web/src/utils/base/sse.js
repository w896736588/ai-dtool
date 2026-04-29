import t from "@/utils/base/type";
import base from '@/utils/base'
const SseEventClean = '[CLEAN]'
const SseEventLogin = '[LOGIN_USERNAME_PASSWORD]'
const SseEventProcess = '[PROCESS]'
const SseMap = {} //存放Sse连接
const SseMsg = {} //存放每个Sse的消息
function GetSseHost(clientId) {
    if (SseMap[clientId]) {
        return SseMap[clientId]
    }
    let url = ''
    let params = 'client_id=' + encodeURIComponent(clientId) + '&token=' + encodeURIComponent(base.GetSafeToken())
    //process.env.NODE_ENV === 'production'
    url = base.GetSseApiHost() + '/sse?' + params
    SseMap[clientId] = new EventSource(url)
    SseMsg[clientId] = ''
    return SseMap[clientId]
}

function SseExist(clientId){
    return SseMap && SseMap[clientId]
}

function SseMsgClean(clientId){
    if(SseExist(clientId)){
        SseMsg[clientId] = ''
    }
}

//设置Sse 创建链接成功回调函数
function SetSseOpenFunc(clientId, callFunc) {
    GetSseHost(clientId).onopen = function (event)  {
        callFunc(event)
    };
}

//设置Sse 创建链接失败回调函数
function SetSseErrorFunc(clientId, callFunc) {
    GetSseHost(clientId).onerror = function (event)  {
        callFunc(event)
    };
}

//设置Sse回调函数
function SetSseMessageFunc(clientId, callFunc) {
    GetSseHost(clientId).onmessage = function (event)  {
        callFunc(event.data);
    };
}

//uniqueKey格式必须为xx#xx格式
function SseCreate(clientId , milliseconds , receiveMsgFunc , openFunc){
    if(SseExist(clientId)){
        return '已存在连接'
    }
    SetSseOpenFunc(clientId, function () {
        if(openFunc && typeof openFunc === 'function'){
            openFunc(clientId)
        }
    })
    SetSseErrorFunc(clientId , function (event){
        if (event.readyState  === EventSource.CLOSED) {
            //receiveMsgFunc('连接已断开，不会重连')
        }else if (event.readyState  === EventSource.CONNECTING) {
            //receiveMsgFunc('连接已断开，重连中')
        }else if (event.readyState  === EventSource.OPEN) {
            //receiveMsgFunc('连接已断开，连接已建立')
        }
    })
    SetSseMessageFunc(clientId, function (msg) {
        let returnMsg = ''
        if(msg === '[DONE]'){
            returnMsg = "\n"
            receiveMsgFunc(returnMsg)
        }else if(msg === '[CLEAN]'){
            returnMsg = msg
            receiveMsgFunc(returnMsg)
        }
        try{
            let msgObj = JSON.parse(msg)
            if(t.IsObject(msgObj) && msgObj.choices && t.IsArray(msgObj.choices)){
                for (let i = 0; i < msgObj.choices.length; i++) {
                    let choice = msgObj.choices[i];
                    if(choice.delta && choice.delta.content){
                        returnMsg += choice.delta.content
                    }
                }
            }
        }catch (e){
            console.log('未识别的消息类型',e ,msg)
        }
        receiveMsgFunc(returnMsg)
    })
}

// 关闭并清理指定 clientId 的 SSE 连接
function SseClose(clientId) {
    if (!SseMap[clientId]) return

    SseMap[clientId].close()          // 1. 关闭长连接
    delete SseMap[clientId]           // 2. 移除连接对象
    delete SseMsg[clientId]           // 3. 移除缓存消息
}

export default {
    SetSseMessageFunc,
    SetSseErrorFunc,
    SseCreate,
    SseExist,
    SseMsg,
    SseMsgClean,
    SseEventClean,
    SseEventProcess,
    SseEventLogin,
    SseClose,
}
