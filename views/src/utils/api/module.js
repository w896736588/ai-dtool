
function GetSystemName(){
  return 'zhima';
}

//启用的模块
function GetOpenModuleList(){
  switch (GetSystemName()){
    case 'zhima':
      return ['redis' , 'consumer' , 'docker' , 'git' , 'vip' , 'wechatkefu'];
    default:
      return ['redis' , 'consumer' , 'git'];
  }
}

//拿到redis配置
function GetRedisConfigList(){
  switch (GetSystemName()){
    case 'zhima':
      return require("../../config/zhima/redisList.json")
    default:
      return [];
  }
}


//拿到code
function GetCodeConfigList(){
  let codeList = [];
  switch (GetSystemName()){
    case 'zhima':
      codeList = require("../../config/zhima/codeList.json")
      break;
    default:
      codeList = [];
  }
  return codeList;
}

//拿到mysql配置
function GetMysqlConfigList(){
  switch (GetSystemName()){
    case 'zhima':
      return require("../../config/zhima/mysqlList.json")
    default:
      return [];
  }
}

//拿到shell配置
function GetShellConfigList(){
  switch (GetSystemName()){
    case 'zhima':
      return require("../../config/zhima/shellConfigList.json")
    default:
      return [];
  }
}

//拿到shell配置
function GetConsumerConfigList(){
  switch (GetSystemName()){
    case 'zhima':
      return require("../../config/zhima/consumer.json")
    default:
      return [];
  }
}

//拿到encrypt配置
function GetEncryptConfig(){
  switch (GetSystemName()){
    case 'zhima':
      return require("../../config/zhima/encrypt.json")
    default:
      return [];
  }
}

export default {
  GetSystemName,
  GetOpenModuleList,
  GetRedisConfigList,
  GetMysqlConfigList,
  GetShellConfigList,
  GetEncryptConfig,
  GetCodeConfigList,
  GetConsumerConfigList,
}
