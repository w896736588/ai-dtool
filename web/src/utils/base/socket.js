
const socketMap = {} //存放socket连接
const socketMsg = {} //存放每个socket的消息
const socketReConnection = {} //存放重连以及定时器
function GetSocketHost(uniqueKey) {
    if (socketMap[uniqueKey]) {
        return socketMap[uniqueKey]
    }
    let url = ''
    let params = 'uniqueKey=' + encodeURIComponent(uniqueKey)
    if (process.env.NODE_ENV === 'production') {
        url = 'ws://localhost:17171/socket?' + params
    } else {
        url = 'ws://localhost:17171/socket?' + params
    }
    socketMap[uniqueKey] = new WebSocket(url)
    socketMsg[uniqueKey] = ''
    return socketMap[uniqueKey]
}

function SocketExist(uniqueKey){
    return socketMap && socketMap[uniqueKey]
}

function SocketMsgClean(uniqueKey){
    if(SocketExist(uniqueKey)){
        socketMsg[uniqueKey] = ''
    }
}

//发送消息
function SendSocketSendMsg(uniqueKey , msg) {
    GetSocketHost(uniqueKey).send(msg)
}

//设置socket 创建连接回调函数
function SetSocketOnOpenFunc(uniqueKey , callFunc) {
    GetSocketHost(uniqueKey).onopen = () => {
        callFunc()
    }
}

//设置socket 创建链接失败回调函数
function SetSocketErrorFunc(uniqueKey, callFunc) {
    GetSocketHost(uniqueKey).onerror = (error) => {
        callFunc(error)
    }
}

//设置socket 创建链接失败回调函数
function SetSocketCloseFunc(uniqueKey, callFunc) {
    GetSocketHost(uniqueKey).onclose = () => {
        callFunc()
    }
}

//设置socket回调函数
function SetSocketMessageFunc(uniqueKey, callFunc) {
    GetSocketHost(uniqueKey).onmessage = (message) => {
        callFunc(message.data)
    }
}

//ping
function SocketPing(uniqueKey) {
    //GetSocketHost(uniqueKey).send(``)
}

//设置心跳
function SetSocketHeart(uniqueKey) {
    SocketPing(uniqueKey)
    setInterval(function () {
        SocketPing(uniqueKey)
    }, 20000)
}

function SocketMsg(uniqueKey){
    if(socketMsg[uniqueKey]){
        return socketMsg[uniqueKey]
    }
    return '不存在'
}
//uniqueKey格式必须为xx#xx格式
function SocketCreate(uniqueKey , milliseconds , receiveMsgFunc){
    if(SocketExist(uniqueKey)){
        //socketMsg[uniqueKey] = '' //清空
        return '已存在连接'
    }
    SetSocketCloseFunc(uniqueKey, function () {
        socketMsg[uniqueKey] += 'socket连接已断开，下一次任意动作将触发重连'
        receiveMsgFunc(socketMsg[uniqueKey])
        console.log('链接已断开')
        delete socketMap[uniqueKey]
    })
    SetSocketOnOpenFunc(uniqueKey, function () {
        // socketMsg[uniqueKey] += 'connection success ..'
        receiveMsgFunc(socketMsg[uniqueKey])
    })
    SetSocketErrorFunc(uniqueKey , function (){
        console.log('socket 链接失败' , uniqueKey)
        socketMsg[uniqueKey] += 'socket连接已断开，下一次任意动作将触发重连'
        receiveMsgFunc(socketMsg[uniqueKey])
        console.log('链接已断开')
        delete socketMap[uniqueKey]
    })
    SetSocketMessageFunc(uniqueKey, function (msg) {
        socketMsg[uniqueKey] += msg
        receiveMsgFunc(socketMsg[uniqueKey])
        setTimeout(function () {
            let obj = document.getElementById('showShellResult')
            if (obj) {
                obj.scrollTop = obj.scrollHeight + 200
            }
        }, milliseconds)
    })
}

export default {
    SendSocketSendMsg,
    SetSocketMessageFunc,
    SetSocketErrorFunc,
    SetSocketOnOpenFunc,
    SetSocketHeart,
    SocketCreate,
    SocketExist,
    SocketMsg,
    SocketMsgClean,
}