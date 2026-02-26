import store from './base/store'
import notify from './base/notify'

//拿到接口地址
function getApiHost() {
  if (process.env.NODE_ENV === 'production') {
    return ''
  }
  return 'http://localhost:7173'
}

//消费者列表
function getSupervisorConfigList(confList, consumerConfig) {
  let addConfigList = []
  for (let i in confList) {
    let configParam = confList[i]
    if (configParam.length !== 2) {
      continue
    }
    let configFileName = consumerConfig.config_dir + '/' + configParam[0]
    configParam[1] = configParam[1].replaceAll('[', '')
    configParam[1] = configParam[1].replaceAll(']', '')
    configParam[1] = configParam[1].replaceAll('program:', '')
    configParam[1] = configParam[1].replaceAll('\r', '')
    //建立配置
    let showName = store.getStore(configParam[0])
    if (showName === null || showName === undefined) {
      showName = configParam[0].split('.')[0]
    }
    addConfigList.push({
      name: configParam[0],
      supervisor_config: configFileName,
      supervisor_name: configParam[1],
      running_status: '',
      showName: showName,
      processNum: 0,
    })
  }
  return addConfigList
}

export default {
  getApiHost,
  getSupervisorConfigList,
}
