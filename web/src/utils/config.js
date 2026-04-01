import store from './base/store'
import notify from './base/notify'
import { normalizeSupervisorProgramNameFromHeader } from './supervisor_status'

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
    const configName = typeof configParam[0] === 'string' ? configParam[0].trim() : ''
    const supervisorName = typeof configParam[1] === 'string' ? configParam[1] : ''
    if (configName === '' || supervisorName === '') {
      continue
    }
    let configFileName = consumerConfig.config_dir + '/' + configName
    configParam[1] = normalizeSupervisorProgramNameFromHeader(supervisorName)
    //建立配置
    let showName = store.getStore(configName)
    if (showName === null || showName === undefined) {
      showName = configName.split('.')[0]
    }
    addConfigList.push({
      name: configName,
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
