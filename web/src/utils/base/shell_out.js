import base from "@/utils/base";

function ShellOutCleanErrors(data , callBack){
    base.BasePost('/api/shellOutCleanErrors', data, callBack)
}

function ShellOuts(data , callBack){
    base.BasePost('/api/shellOuts', data, callBack)
}

function ShellOutDelete(data , callBack){
    base.BasePost('/api/shellOutDelete', data, callBack)
}

function ShellOutStop(data , callBack){
    base.BasePost('/api/shellOutStop', data, callBack)
}

function ShellOutRestart(data , callBack){
    base.BasePost('/api/shellOutRestart', data, callBack)
}

function ShellOutStart(data , callBack){
    base.BasePost('/api/shellOutStart', data, callBack)
}

function ShellOutErrorContext(data , callBack){
    base.BasePost('/api/shellOutErrorContext', data, callBack)
}

function ShellOutSearchContent(data , callBack){
    base.BasePost('/api/shellOutSearchContent', data, callBack)
}

function ShellOutCleanLog(data , callBack){
    base.BasePost('/api/shellOutCleanLog', data, callBack)
}

function ShellOutSetFilter(data , callBack){
    base.BasePost('/api/shellOutSetFilter', data, callBack)
}

function ShellOutGetFilter(data , callBack){
    base.BasePost('/api/shellOutGetFilter', data, callBack)
}

export default {
    ShellOutCleanErrors,
    ShellOuts,
    ShellOutDelete,
    ShellOutStop,
    ShellOutErrorContext,
    ShellOutSearchContent,
    ShellOutCleanLog,
    ShellOutRestart,
    ShellOutSetFilter,
    ShellOutGetFilter,
}