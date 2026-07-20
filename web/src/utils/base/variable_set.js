import base from '../base'
import module from '../module'
import store from './store'

function SetVariableGroupList(callBack) {
    base.BasePost('/api/Set/VariableGroupList', {}, callBack)
}

function SetVariableGroupAdd(editVariableGroupConfig, callBack) {
    base.BasePost('/api/Set/VariableGroupAdd', editVariableGroupConfig, callBack)
}

function SetVariableGroupDelete(editVariableGroupConfig, callBack) {
    base.BasePost('/api/Set/VariableGroupDelete', editVariableGroupConfig, callBack)
}

function VariableList(dataOrCallBack, maybeCallBack) {
    const data = typeof dataOrCallBack === 'function' ? {} : (dataOrCallBack || {})
    const callBack = typeof dataOrCallBack === 'function' ? dataOrCallBack : maybeCallBack
    base.BasePost('/api/VariableList', data, callBack)
}

function VariableAdd(variableConfig, callBack) {
    base.BasePost('/api/VariableAdd', variableConfig, callBack)
}

function VariableSetLogin(username, password, callBack) {
    base.BasePost('/api/VariableSetLogin', {username: username, password: password}, callBack)
}

function VariableDelete(variableConfig, callBack) {
    base.BasePost('/api/VariableDel', variableConfig, callBack)
}

function VariableInfo(variableConfig, callBack) {
    base.BasePost('/api/VariableInfo', variableConfig, callBack)
}

function VariableCmdAdd(variable_cmd, callBack) {
    base.BasePost('/api/VariableCmdAdd', variable_cmd, callBack)
}

function VariableCmdDel(variable_cmd, callBack) {
    base.BasePost('/api/VariableCmdDel', variable_cmd, callBack)
}

function VariableRun(sse_distribute_id ,variable_id, run_cmd_id, is_run, replace_list, callBack) {
    base.BasePost('/api/VariableRun', {
        sse_distribute_id : sse_distribute_id,
        variable_id: variable_id,
        run_cmd_id: run_cmd_id,
        is_run: is_run,
        replace_list: replace_list,
    }, callBack)
}

function VariableSet(variable_id, run_cmd_id, replace_list, edit_value, callBack) {
    base.BasePost('/api/VariableSet', {
        variable_id: variable_id,
        run_cmd_id: run_cmd_id,
        replace_list: replace_list,
        edit_value: edit_value,
    }, callBack)
}

export default {
    SetVariableGroupList,
    SetVariableGroupAdd,
    SetVariableGroupDelete,
    VariableList,
    VariableAdd,
    VariableDelete,
    VariableInfo,
    VariableCmdAdd,
    VariableCmdDel,
    VariableRun,
    VariableSet,
    VariableSetLogin,
}
