import store from "./store";
import notify from "./notify";
import wechatKefuList from "../config/wechatKefuList.json";
import supervisorConfigList from "../config/supervisorConfig.json";
import redisList from "../config/redisList.json";
import {getDayCountOfMonth} from "element-ui";

//拿到xkf dev ssh config
function getXkfDevSshConfig(){
  let sshConfig = store.getStore('sshConfig')
  if (sshConfig !== null) {
    return JSON.parse(sshConfig)
  }
  notify.error("请先配置小客服ssh");
  return false
}

//拿到xkf dev ssh config
function getWkDevSshConfig(){
  let wkSshConfig = store.getStore('wkSshConfig')
  if (wkSshConfig !== null) {
    return JSON.parse(wkSshConfig)
  }
  notify.error("请先配置企微ssh");
  return false
}

//xkf dev db
function getXkfDevDbConfig(){
  let xkfDevDbConfig = store.getStore('devTestDbConfig')
  if (xkfDevDbConfig !== null) {
    return JSON.parse(xkfDevDbConfig)
  }
  notify.error("请先配置企微ssh");
  return false
}

//拿到接口地址
function getApiHost(){
  if (process.env.NODE_ENV === 'production') {
    return '';
  }
  return 'http://localhost:7070';
}

//拿到代码环境
function getCodeEnvList(){
  let codeList = require("../config/codeList.json")
  for(let i in codeList){
    if(codeList[i].NameTitle){
      continue;
    }
    codeList[i].NameTitle = codeList[i].Name
    codeList[i].Name = codeList[i].Name + "-" + codeList[i].ParentType
  }
  return codeList
}

//docker list
function getDockerList(){
  return require("../config/dockerList.json")
}

//拿到微信客服列表
function getWechatKefuList(){
  return require("../config/wechatKefuList.json")
}

//拿到账号列表
function getUsernameList(){
  return require("../config/userName.json")
}

//链接地址
function getLinkList(){
  return require("../config/urlList.json")
}

//消费者列表
function getSupervisorConfigList(supervisorOriginConfList , chooseParentType){
  let addConfigList = [];
  for(let i in supervisorOriginConfList){
    let configParam = supervisorOriginConfList[i]
    if(configParam.length !== 2){
      continue;
    }
    let configFileName = ''
    if(chooseParentType === 'xkf'){
      configFileName = '/var/www/dockerfiles/dev_test/docker_volumes/supervisor/etc/supervisor/conf.d/' + configParam[0]
    }else{
      configFileName = '/etc/supervisor/conf.d/' + configParam[0]
    }
    configParam[1] = configParam[1].replaceAll('[' , '')
    configParam[1] = configParam[1].replaceAll(']' , '')
    configParam[1] = configParam[1].replaceAll('program:' , '')
    configParam[1] = configParam[1].replaceAll('\r' , '')
    //建立配置
    let showName = store.getStore(configParam[0])
    if(showName === null || showName === undefined){
      showName = configParam[0].split('.')[0]
    }
    addConfigList.push({
      "name" : configParam[0],
      "supervisor_config" : configFileName,
      "supervisor_name" : configParam[1],
      "running_status" : "",
      "showName" : showName,
    })
  }
  return addConfigList
}

//根据环境获取userName
function getUserNameByEnvCode(userNameList , envName){
  let userName = '';
  for (let i in userNameList) {
    if (userNameList[i].Name === envName || userNameList[i].NameChild === envName) {
      userName = userNameList[i].UserName
    }
  }
  return userName
}

//拿到缓存列表
function getRedisList(){
  let redisList = require("../config/redisList.json")
  for (let i in redisList) {
    redisList[i].UniKey = redisList[i].Name
  }
  return redisList
}

//通过代码环境name获取代码环境配置
function getCodeEnvConfigByCodeEnvName(codeEnvList , envName , parentType = ''){
  let env_config = {};
  for (let i in codeEnvList) {
    if (codeEnvList[i].Name === envName + parentType) {
      env_config = codeEnvList[i]
      break
    }
  }
  if(env_config === {}){
    notify.error("不存在的配置");
  }
  return env_config
}

//通过代码环境配置拿到dockerId
function getDockerIdByCodeEnvConfig(dockerList , codeEnvConfig){
  for (let j in dockerList) {
    if (codeEnvConfig.DockerName === dockerList[j].Name) {
      return dockerList[j].Id
    }
  }
  return ""
}

export default {
  getXkfDevSshConfig,
  getApiHost,
  getWkDevSshConfig,
  getXkfDevDbConfig,
  getCodeEnvList,
  getDockerList,
  getWechatKefuList,
  getUsernameList,
  getLinkList,
  getRedisList,
  getSupervisorConfigList,
  getCodeEnvConfigByCodeEnvName,
  getDockerIdByCodeEnvConfig,
  getUserNameByEnvCode,
}
