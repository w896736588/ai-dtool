//登录
import Vue from "vue";
import store from "../store";
import module from "../api/module"

//登录拿到 unikey
function BaseLogin(userName, password) {
  BasePost('/api/BaseLogin', {
    UserName: userName,
    Password: password,
  }, function (response) {
    console.log(response)
    store.setStore('Unikey', response.Data.unikey)
    BaseRegisterService()
  })
}

//注册服务
function BaseRegisterService() {
  let redisConfigList = module.GetRedisConfigList()
  let mysqlConfigList = module.GetMysqlConfigList()
  let shellConfigList = module.GetShellConfigList()
  let encryptConfig = module.GetEncryptConfig()
  BasePost('/api/BaseRegisterService', {
    Unikey: store.getStore('Unikey'),
    MysqlConfigList : mysqlConfigList,
    RedisConfigList : redisConfigList,
    ShellConfigList : shellConfigList,
    EncryptKey : encryptConfig['EncryptKey'],
    EncryptIv : encryptConfig['EncryptIv']
  }, function (response) {


  })
}

//检查unikey是否已经登录注册
function BaseCheckService() {
  let unikey = store.getStore('Unikey')
  BasePost('/api/BaseCheckUnikeyExist', {
    Unikey: unikey,
  }, function (response) {
    if (response.Data.NeedLogin === '1') {
      let userName = store.getStore('UserName')
      let password = store.getStore('Password')
      BaseLogin(userName, password)
    }
  })
}

//POST请求
function BasePost(uri, params, callBack) {
  params.Unikey = store.getStore('Unikey')
  Vue.axios.post(GetApiHost() + uri, params).then(function (response) {
    callBack(response)
  });
}

function GetUnikey(){
  return store.getStore('Unikey')
}

//拿到接口地址
function GetApiHost() {
  if (process.env.NODE_ENV === 'production') {
    return '';
  }
  return 'http://localhost:7070';
}

//拿到socket链接
var socketMap = {}
function GetSocketHost(unikey , shellName){
  let unikeyConn = unikey +'#' + shellName
  if(socketMap[unikeyConn]){
    return socketMap[unikeyConn]
  }
  let url = '';
  let params = 'Unikey=' + unikey + '&ShellName=' + shellName;
  if (process.env.NODE_ENV === 'production') {
    url = 'ws://localhost:7071/socket?' + params;
  }else{
    url = 'ws://localhost:7071/socket?' + params;
  }
  socketMap[unikeyConn] = new WebSocket(url)
  return socketMap[unikeyConn]
}

//发送消息
function SendSocketSendMsg(unikey , shellName , msg){
  GetSocketHost(unikey , shellName).send(msg)
}

//设置socket 创建连接回调函数
function SetSocketOnOpenFunc(unikey , shellName , callFunc){
  GetSocketHost(unikey , shellName).onopen = () => {
    callFunc();
  };
}

//设置socket 创建链接失败回调函数
function SetSocketErrorFunc(unikey , shellName , callFunc){
  GetSocketHost(unikey , shellName).onerror = (error) => {
    callFunc(error);
  };
}

//设置socket回调函数
function SetSocketMessageFunc(unikey , shellName , callFunc){
  GetSocketHost(unikey , shellName).onmessage = (message) => {
    callFunc(message.data);
  };
}

//ping
function SocketPing(unikey , shellName){
  GetSocketHost(unikey , shellName).send(`ping`)
}

//设置心跳
function SetSocketHeart(unikey , shellName){
  SocketPing(unikey , shellName)
  setInterval(function (){
    SocketPing(unikey , shellName)
  } , 20000)
}

export default {
  BaseRegisterService,
  BasePost,
  BaseCheckService,
  GetApiHost,
  GetUnikey,
  SendSocketSendMsg,
  SetSocketMessageFunc,
  SetSocketErrorFunc,
  SetSocketOnOpenFunc,
  SetSocketHeart,
}
