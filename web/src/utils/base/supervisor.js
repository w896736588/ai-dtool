import base from '../base'
import mod from '../module'

//拿到配置
function SupervisorConfList(supervisorConfig, callBack) {
  base.BasePost('/api/SupervisorConfList',supervisorConfig, callBack)
}

function SupervisorRestartAll(supervisorConfig , callBack) {
  base.BasePost('/api/SupervisorRestartAll', supervisorConfig , callBack)
}

function SupervisorStatusList(supervisorConfig, callBack) {
  base.BasePost('/api/SupervisorStatusList', supervisorConfig , callBack)
}

function SupervisorConfigShow(supervisorConfig,configDir, callBack) {
    supervisorConfig.config_path = configDir
    base.BasePost('/api/SupervisorConfigShow', supervisorConfig , callBack)
}

function SupervisorRestart(supervisorConfig, SupervisorName, callBack) {
    supervisorConfig.supervisor_name = SupervisorName
    base.BasePost('/api/SupervisorRestart', supervisorConfig , callBack)
}

function SupervisorStop(supervisorConfig , SupervisorName, callBack) {
    supervisorConfig.supervisor_name = SupervisorName
    base.BasePost('/api/SupervisorStop', supervisorConfig , callBack)
}

function SupervisorConfigList(config , callBack) {
    base.BasePost('/api/SupervisorConfigList', config,callBack)
}

export default {
    SupervisorConfList,
    SupervisorRestartAll,
    SupervisorStatusList,
    SupervisorConfigShow,
    SupervisorRestart,
    SupervisorStop,
    SupervisorConfigList,
}
