import base from '../base'
import module from '../module'
import store from './store'

function SetCmdGroupList(callBack){
    base.BasePost('/api/Set/CmdGroupList', {}, callBack)
}

function SetCmdGroupAdd(editCmdGroupConfig , callBack) {
    base.BasePost('/api/Set/CmdGroupAdd', editCmdGroupConfig, callBack)
}

function SetCmdGroupDelete(editCmdGroupConfig , callBack){
    base.BasePost('/api/Set/CmdGroupDelete', editCmdGroupConfig, callBack)
}

function CmdList(callBack){
    base.BasePost('/api/CmdList', {}, callBack)
}

function CmdAdd(id , type , name , Cmd_group_id , config , remark , callBack){
    base.BasePost('/api/CmdAdd', {
        id : id,
        type : type ,
        name : name ,
        Cmd_group_id : Cmd_group_id ,
        config : JSON.stringify(config) ,
        remark : remark,
    }, callBack)
}

function CmdDelete(id , callBack){
    base.BasePost('/api/CmdDel', {id : id}, callBack)
}

export default {
    SetCmdGroupList,
    SetCmdGroupAdd,
    SetCmdGroupDelete,
    CmdList,
    CmdAdd,
    CmdDelete
}
