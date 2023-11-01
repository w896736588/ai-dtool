import base from "./base";
import mod from "./module";

//拿到配置
function ConsumerConfigList(DockerName , ShellName , callBack){
  base.BasePost('/api/ConsumerConfigList', {
    DockerName : DockerName,
    ShellName : ShellName,
  }, function (response) {
    callBack(response)
  })
}

function ConsumerRestartAll(DockerName , ShellName , callBack){
  base.BasePost('/api/ConsumerRestartAll', {
    DockerName : DockerName,
    ShellName : ShellName,
  }, function (response) {
    callBack(response)
  })
}

function ConsumerStopAll(){

}

function ConsumerStatusList(DockerName , ShellName , callBack){
  base.BasePost('/api/ConsumerStatusList', {
    DockerName : DockerName,
    ShellName : ShellName,
  }, function (response) {
    callBack(response)
  })
}

function ConsumerConfigShow(DockerName , ShellName , SupervisorConfigPath , callBack){
  base.BasePost('/api/ConsumerConfigShow', {
    DockerName : DockerName,
    ShellName : ShellName,
    SupervisorConfigPath : SupervisorConfigPath,
  }, function (response) {
    callBack(response)
  })
}

function ConsumerRestart(DockerName , ShellName , ConsumerName , callBack){
  base.BasePost('/api/ConsumerRestart', {
    DockerName : DockerName,
    ShellName : ShellName,
    ConsumerName : ConsumerName,
  }, function (response) {
    callBack(response)
  })
}

function ConsumerStop(DockerName , ShellName , ConsumerName , callBack){
  base.BasePost('/api/ConsumerStop', {
    DockerName : DockerName,
    ShellName : ShellName,
    ConsumerName : ConsumerName,
  }, function (response) {
    callBack(response)
  })
}

//根据name拿到消费者配置
function GetConsumerConfigByName(consumerName){
  let consumerConfigList = mod.GetConsumerConfigList()
  console.log(consumerConfigList,consumerName)
  for(let i in consumerConfigList){
    if(consumerConfigList[i].Unikey === consumerName){
      return consumerConfigList[i]
    }
  }
}

export default {
  ConsumerConfigList,
  ConsumerRestartAll,
  ConsumerStopAll,
  ConsumerStatusList,
  ConsumerConfigShow,
  ConsumerRestart,
  ConsumerStop,
  GetConsumerConfigByName,
}
